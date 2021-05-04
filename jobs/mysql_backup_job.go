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
	"strconv"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/percona/pmm/api/agentpb"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	xtrabackupBin = "xtrabackup"
	xbcloudBin    = "xbcloud"
	qpressBin     = "qpress"
)

// MySQLBackupJob implements Job for MySQL backup.
type MySQLBackupJob struct {
	id       string
	timeout  time.Duration
	l        logrus.FieldLogger
	name     string
	connConf DBConnConfig
	location BackupLocationConfig
}

// DBConnConfig contains required properties for connection to DB.
type DBConnConfig struct {
	User     string
	Password string
	Address  string
	Port     int
	Socket   string
}

// NewMySQLBackupJob constructs new Job for MySQL backup.
func NewMySQLBackupJob(id string, timeout time.Duration, name string, connConf DBConnConfig, locationConfig BackupLocationConfig) *MySQLBackupJob {
	return &MySQLBackupJob{
		id:       id,
		timeout:  timeout,
		l:        logrus.WithFields(logrus.Fields{"id": id, "type": "mysql_backup", "name": name}),
		name:     name,
		connConf: connConf,
		location: locationConfig,
	}
}

// ID returns Job id.
func (j *MySQLBackupJob) ID() string {
	return j.id
}

// Type returns Job type.
func (j *MySQLBackupJob) Type() string {
	return "mysql_backup"
}

// Timeout returns Job timeout.
func (j *MySQLBackupJob) Timeout() time.Duration {
	return j.timeout
}

// Run starts Job execution.
func (j *MySQLBackupJob) Run(ctx context.Context, send Send) (rerr error) {
	if _, err := exec.LookPath(xtrabackupBin); err != nil {
		return errors.Wrapf(err, "lookpath: %s", xtrabackupBin)
	}

	if _, err := exec.LookPath(qpressBin); err != nil {
		return errors.Wrapf(err, "lookpath: %s", qpressBin)
	}

	if j.location.S3Config != nil {
		if _, err := exec.LookPath(xbcloudBin); err != nil {
			return errors.Wrapf(err, "lookpath: %s", xbcloudBin)
		}
	}

	xtrabackupCmd := exec.CommandContext(ctx, xtrabackupBin,
		"--user="+j.connConf.User,
		"--password="+j.connConf.Password,
		"--compress",
		"--backup") // #nosec G204

	switch {
	case j.connConf.Address != "":
		xtrabackupCmd.Args = append(xtrabackupCmd.Args, "--host="+j.connConf.Address)
		if j.connConf.Port > 0 {
			xtrabackupCmd.Args = append(xtrabackupCmd.Args, "--port="+strconv.Itoa(j.connConf.Port))
		}
	case j.connConf.Socket != "":
		xtrabackupCmd.Args = append(xtrabackupCmd.Args, "--socket="+j.connConf.Socket)
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

		defer func() {
			err := xbcloudCmd.Wait()
			if err == nil {
				return
			}

			if rerr != nil {
				rerr = errors.Wrapf(rerr, "xbcloud wait error: %s", err)
			} else {
				rerr = wrapError(err)
			}
		}()
	}

	if err := xtrabackupCmd.Start(); err != nil {
		return wrapError(err)
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

	return nil
}
