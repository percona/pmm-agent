// pmm-agent
// Copyright 2019 Percona LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package client contains business logic of working with pmm-managed.
package client

import (
	"context"
	"database/sql"
	"net"
	"os/exec"
	"regexp"
	"strconv"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/percona/pmm/api/agentpb"
)

const (
	versionCheckTimeout = 5 * time.Second
	mysqlBin            = "mysql"
	xtrabackupBin       = "xtrabackup"
)

var (
	mysqlVersionRegexp      = regexp.MustCompile("^.*Ver ([!-~]*).*$")
	xtrabackupVersionRegexp = regexp.MustCompile("^xtrabackup version ([!-~]*).*$")
)

func (c *Client) remoteMySQLVersion(s *agentpb.GetVersionRequest_RemoteMysql) (string, error) {
	remoteMySQL := s.RemoteMysql

	if ((remoteMySQL.GetAddress() != "" || remoteMySQL.GetPort() != 0) && remoteMySQL.GetSocket() != "") ||
		((remoteMySQL.GetAddress() == "" || remoteMySQL.GetPort() == 0) && remoteMySQL.GetSocket() == "") {
		return "", errors.Errorf(
			"either address with port or socket should be set: address: %q, port: %d, socket: %q",
			remoteMySQL.GetAddress(),
			remoteMySQL.GetPort(),
			remoteMySQL.GetSocket(),
		)
	}

	cfg := mysql.NewConfig()
	cfg.User = remoteMySQL.GetUser()
	cfg.Passwd = remoteMySQL.GetPassword()

	if remoteMySQL.GetAddress() != "" {
		cfg.Net = "tcp"
		cfg.Addr = net.JoinHostPort(remoteMySQL.GetAddress(), strconv.Itoa(int(remoteMySQL.GetPort())))
	} else {
		cfg.Net = "unix"
		cfg.Addr = remoteMySQL.GetSocket()
	}

	connector, err := mysql.NewConnector(cfg)
	if err != nil {
		return "", errors.WithStack(err)
	}

	db := sql.OpenDB(connector)
	defer func() {
		if err := db.Close(); err != nil {
			c.l.WithError(err).Error("failed to close mysql connection")
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), versionCheckTimeout)
	defer cancel()

	var version string
	if err := db.QueryRowContext(ctx, "SELECT VERSION();").Scan(&version); err != nil {
		return "", errors.WithStack(err)
	}

	return version, nil
}

func (c *Client) localMySQLVersion() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), versionCheckTimeout)
	defer cancel()

	if _, err := exec.LookPath(mysqlBin); err != nil {
		return "", errors.Wrapf(err, "lookpath: %s", mysqlBin)
	}

	versionBytes, err := exec.CommandContext(ctx, mysqlBin, "--version").CombinedOutput()
	if err != nil {
		return "", errors.WithStack(err)
	}

	matches := mysqlVersionRegexp.FindStringSubmatch(string(versionBytes))
	if len(matches) != 2 {
		return "", errors.Errorf("cannot match version from output %q", string(versionBytes))
	}

	return matches[1], nil
}

func (c *Client) xtrabackupVersion() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), versionCheckTimeout)
	defer cancel()

	if _, err := exec.LookPath(xtrabackupBin); err != nil {
		return "", errors.Wrapf(err, "lookpath: %s", xtrabackupBin)
	}

	versionBytes, err := exec.CommandContext(ctx, xtrabackupBin, "--version").CombinedOutput()
	if err != nil {
		return "", errors.WithStack(err)
	}

	matches := xtrabackupVersionRegexp.FindStringSubmatch(string(versionBytes))
	if len(matches) != 2 {
		return "", errors.Errorf("cannot match version from output %q", string(versionBytes))
	}

	return matches[1], nil
}

func (c *Client) handleVersionRequest(r *agentpb.GetVersionRequest) (string, *status.Status) {
	var version string
	var err error
	switch s := r.Software.(type) {
	case *agentpb.GetVersionRequest_RemoteMysql:
		version, err = c.remoteMySQLVersion(s)
	case *agentpb.GetVersionRequest_LocalMysql:
		version, err = c.localMySQLVersion()
	case *agentpb.GetVersionRequest_Xtrabackup:
		version, err = c.xtrabackupVersion()
	default:
		return "", status.Newf(codes.Unknown, "unknown software type %v.", r)
	}

	if err != nil {
		return "", status.New(codes.Internal, err.Error())
	}

	return version, nil
}
