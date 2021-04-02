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

package jobs

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/golang/protobuf/ptypes"
	"github.com/percona/pmm/api/agentpb"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	xtrabackupBin = "xtrabackup"
	xbcloudBin    = "xbcloud"
)

// MySQLBackupJob implements Job for MySQL backup.
type MySQLBackupJob struct {
	id       string
	timeout  time.Duration
	l        *logrus.Entry
	name     string
	dsn      string
	location BackupLocationConfig
}

// S3LocationConfig contains required properties for accessing S3 Bucket.
type S3LocationConfig struct {
	Endpoint     string
	AccessKey    string
	SecretKey    string
	BucketName   string
	BucketRegion string
}

// BackupLocationConfig groups all backup locations configs.
type BackupLocationConfig struct {
	S3Config *S3LocationConfig
}

// NewMySQLBackupJob constructs new Job for MySQL backup.
func NewMySQLBackupJob(id string, timeout time.Duration, name, dsn string, locationConfig BackupLocationConfig) *MySQLBackupJob {
	return &MySQLBackupJob{
		id:       id,
		timeout:  timeout,
		l:        logrus.WithFields(logrus.Fields{"id": id, "type": "mysql_backup", "name": name}),
		name:     name,
		dsn:      dsn,
		location: locationConfig,
	}
}

// ID returns job id.
func (j *MySQLBackupJob) ID() string {
	return j.id
}

// Type returns job type.
func (j *MySQLBackupJob) Type() string {
	return "mysql_backup"
}

// Timeout returns job timeout.
func (j *MySQLBackupJob) Timeout() time.Duration {
	return j.timeout
}

func (j *MySQLBackupJob) Run(ctx context.Context, send Send) error {
	t := time.Now()
	j.l.Info("MySQL backup started")

	mysqlConfig, err := mysql.ParseDSN(j.dsn)
	if err != nil {
		return errors.Wrapf(err, "mysql parse dsn")
	}

	if _, err := exec.LookPath(xtrabackupBin); err != nil {
		return errors.Wrapf(err, "lookpath: %s", xtrabackupBin)
	}

	if j.location.S3Config != nil {
		if _, err := exec.LookPath(xbcloudBin); err != nil {
			return errors.Wrapf(err, "lookpath: %s", xbcloudBin)
		}
	}

	xtrabackupCmd := exec.CommandContext(ctx, xtrabackupBin,
		"--user="+mysqlConfig.User,
		"--password="+mysqlConfig.Passwd,
		"--compress",
		"--backup") // #nosec G204

	switch mysqlConfig.Net {
	case "tcp":
		splitAddr := strings.Split(mysqlConfig.Addr, ":")
		xtrabackupCmd.Args = append(xtrabackupCmd.Args, "--host="+splitAddr[0])
		if len(splitAddr) > 1 {
			xtrabackupCmd.Args = append(xtrabackupCmd.Args, "--port="+splitAddr[1])
		}
	case "unix":
		xtrabackupCmd.Args = append(xtrabackupCmd.Args, "--socket="+mysqlConfig.Addr)
	}

	var xbcloudCmd *exec.Cmd
	switch {
	case j.location.S3Config != nil:
		xtrabackupCmd.Args = append(xtrabackupCmd.Args, "--stream=xbstream")
		xbcloudCmd = exec.CommandContext(ctx, xbcloudBin,
			"put",
			"--storage=s3",
			"--s3-endpoint="+j.location.S3Config.Endpoint,
			"--s3-access-key="+j.location.S3Config.AccessKey,
			"--s3-secret-key="+j.location.S3Config.SecretKey,
			"--s3-bucket="+j.location.S3Config.BucketName,
			"--s3-region="+j.location.S3Config.BucketRegion,
			"--parallel=10",
			j.name) // #nosec G204
	default:
		return errors.Errorf("unknown location config")
	}

	var outBuffer bytes.Buffer
	var errBackupBuffer bytes.Buffer
	var errCloudBuffer bytes.Buffer
	xtrabackupCmd.Stderr = &errBackupBuffer

	xtrabackupStdout, err := xtrabackupCmd.StdoutPipe()
	if err != nil {
		return errors.Wrapf(err, "failed to get xtrabackup stdout pipe")
	}

	wrapError := func(err error) error {
		return errors.Wrapf(err, "xtrabackup err: %s\n xbcloud out: %s\n xbcloud err: %s",
			errBackupBuffer.String(), outBuffer.String(), errCloudBuffer.String())
	}

	if xbcloudCmd != nil {
		xbcloudCmd.Stdin = xtrabackupStdout
		xbcloudCmd.Stdout = &outBuffer
		xbcloudCmd.Stderr = &errCloudBuffer
		if err := xbcloudCmd.Start(); err != nil {
			return wrapError(err)
		}
	}

	if err := xtrabackupCmd.Start(); err != nil {
		return wrapError(err)
	}

	if xbcloudCmd != nil {
		if err := xbcloudCmd.Wait(); err != nil {
			return wrapError(err)
		}
	}

	if err := xtrabackupCmd.Wait(); err != nil {
		return wrapError(err)
	}

	send(&agentpb.JobResult{
		JobId:     j.id,
		Timestamp: ptypes.TimestampNow(),
		Result: &agentpb.JobResult_MysqlBackup{
			MysqlBackup: &agentpb.JobResult_MySQLBackup{},
		},
	})

	j.l.WithField("duration", time.Since(t).String()).Info("MySQL backup finished")
	return nil
}
