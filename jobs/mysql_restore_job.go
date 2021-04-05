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
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/percona/pmm/api/agentpb"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	xtrabackupBin       = "xtrabackup"
	xbcloudBin          = "xbcloud"
	xbstreamBin         = "xbstream"
	qpressBin           = "qpress"
	mySQLServiceName    = "mysql"
	mySQLUserName       = "mysql"
	mySQLGroupName      = "mysql"
	mySQLDirectory      = "/var/lib/mysql"
	stopTimeout         = 5 * time.Second
	activeCheckInterval = time.Second
)

type MySQLRestoreJob struct {
	id       string
	timeout  time.Duration
	l        *logrus.Entry
	name     string
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

func NewMySQLRestoreJob(id string, timeout time.Duration, name string, locationConfig BackupLocationConfig) *MySQLRestoreJob {
	return &MySQLRestoreJob{
		id:       id,
		timeout:  timeout,
		l:        logrus.WithFields(logrus.Fields{"id": id, "type": "mysql_backup"}),
		name:     name,
		location: locationConfig,
	}

}

// ID returns job id.
func (j *MySQLRestoreJob) ID() string {
	return j.id
}

// Type returns job type.
func (j *MySQLRestoreJob) Type() string {
	return "mysql_restore"
}

// Timeouts returns job timeout.
func (j *MySQLRestoreJob) Timeout() time.Duration {
	return j.timeout
}

func binariesInstalled() error {
	if _, err := exec.LookPath(xtrabackupBin); err != nil {
		return errors.Wrapf(err, "lookpath: %s", xtrabackupBin)
	}

	//if j.location.S3Config != nil {
	if _, err := exec.LookPath(xbcloudBin); err != nil {
		return errors.Wrapf(err, "lookpath: %s", xbcloudBin)
	}

	if _, err := exec.LookPath(xbstreamBin); err != nil {
		return errors.Wrapf(err, "lookpath: %s", xbstreamBin)
	}
	//}

	if _, err := exec.LookPath(qpressBin); err != nil {
		return errors.Wrapf(err, "lookpath: %s", qpressBin)
	}

	return nil
}

// stdout and stderr could be returned even if rerr is not nil
func restoreMySQLFromS3(
	ctx context.Context,
	backupName string,
	config *BackupLocationConfig,
	targetDirectory string,
) (stdout, stderr *bytes.Buffer, rerr error) {
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
	}()

	// TODO: redirect log (2>download.log)
	xbcloudCmd := exec.CommandContext(ctx, xbcloudBin,
		"get",
		fmt.Sprintf("s3://%s/%s", config.S3Config.BucketName, backupName),
		"--storage=s3",
		"--s3-endpoint="+config.S3Config.Endpoint,
		"--s3-access-key="+config.S3Config.AccessKey,
		"--s3-secret-key="+config.S3Config.SecretKey,
		"--s3-bucket="+config.S3Config.BucketName,
		"--s3-region="+config.S3Config.BucketRegion,
		"--parallel=10",
	)

	var xbcloudStderr bytes.Buffer
	xbcloudCmd.Stderr = &xbcloudStderr

	xbcloudStdout, err := xbcloudCmd.StdoutPipe()
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to get xbcloud stdout pipe")
	}

	xbstreamCmd := exec.CommandContext(ctx, xbstreamBin,
		"restore",
		"-x",
		"--directory="+targetDirectory,
		"--parallel=10",
		"--decompress",
	)

	var xbstreamStdout bytes.Buffer
	var xbstreamStderr bytes.Buffer
	xbstreamCmd.Stdin = xbcloudStdout
	xbstreamCmd.Stdout = &xbstreamStdout
	xbstreamCmd.Stderr = &xbstreamStderr

	if err := xbcloudCmd.Start(); err != nil {
		return nil, nil, errors.WithStack(err)
	}
	defer func() {
		err := xbcloudCmd.Wait()
		if err == nil {
			return
		}

		if rerr != nil {
			rerr = errors.Wrapf(rerr, "xbcloud waiting error: %s", err)
		} else {
			rerr = errors.WithStack(err)
		}
	}()

	if err := xbstreamCmd.Start(); err != nil {
		return nil, nil, errors.WithStack(err)
	}

	if err := xbstreamCmd.Wait(); err != nil {
		return &xbstreamStdout, &xbstreamStderr, errors.WithStack(err)
	}

	return &xbstreamStdout, &xbstreamStderr, nil
}

func mySQLActive() (bool, error) {
	out, err := exec.Command("systemctl", "is-active", mySQLServiceName).Output()
	if err != nil {
		return false, err
	}

	switch strings.TrimSpace(string(out)) {
	case "active":
		return true, nil
	case "inactive":
		return false, nil
	default:
		return false, fmt.Errorf("unknown status '%s'", string(out))
	}
}

func stopMySQL() error {
	if _, err := exec.Command("systemctl", "stop", mySQLServiceName).Output(); err != nil {
		return err
	}

	timeout := time.After(stopTimeout)
	ticker := time.NewTicker(activeCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			active, err := mySQLActive()
			if err != nil {
				return errors.Wrap(err, "couldn't get MySQL status")
			}
			if !active {
				return nil
			}
		case <-timeout:
			return errors.New("couldn't stop MySQL: timeout")
		}
	}
}

func startMySQL() error {
	if _, err := exec.Command("systemctl", "start", mySQLServiceName).Output(); err != nil {
		return err
	}

	timeout := time.After(stopTimeout)
	ticker := time.NewTicker(activeCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			active, err := mySQLActive()
			if err != nil {
				return errors.Wrap(err, "couldn't get MySQL status")
			}
			if active {
				return nil
			}
		case <-timeout:
			return errors.New("couldn't start MySQL: timeout")
		}
	}
}

func chownRecursive(path string, uid, gid int) error {
	return filepath.Walk(path, func(name string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		return errors.WithStack(os.Chown(name, uid, gid))
	})
}

func mySQLUserAndGroupIDs() (uid, gid int, rerr error) {
	u, err := user.LookupId(mySQLUserName)
	if err != nil {
		return 0, 0, errors.WithStack(err)
	}

	uid, err = strconv.Atoi(u.Uid)
	if err != nil {
		return 0, 0, errors.WithStack(err)
	}

	g, err := user.LookupGroup(mySQLGroupName)
	if err != nil {
		return 0, 0, errors.WithStack(err)
	}

	gid, err = strconv.Atoi(g.Gid)
	if err != nil {
		return 0, 0, errors.WithStack(err)
	}

	return uid, gid, nil
}

func restoreBackup(backupDirectory, mySQLDirectory string) error {
	if _, err := exec.Command(xtrabackupBin, "--prepare", "--target-dir="+backupDirectory).Output(); err != nil {
		return errors.WithStack(err)
	}

	if err := os.Rename(mySQLDirectory, mySQLDirectory+".old"); err != nil {
		return errors.WithStack(err)
	}

	// TODO: restore target dir!!!
	if _, err := exec.Command(xtrabackupBin, "--copy-back", "--target-dir="+backupDirectory).Output(); err != nil {
		return errors.WithStack(err)
	}

	uid, gid, err := mySQLUserAndGroupIDs()
	if err != nil {
		return errors.WithStack(err)
	}
	if err := chownRecursive(mySQLDirectory, uid, gid); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (j *MySQLRestoreJob) Run(ctx context.Context, send Send) (rerr error) {
	if err := binariesInstalled(); err != nil {
		return errors.WithStack(err)
	}

	if _, _, err := mySQLUserAndGroupIDs(); err != nil {
		return errors.WithStack(err)
	}

	tmpDir, err := ioutil.TempDir("", "backup-restore")
	defer func() {
		err := os.RemoveAll(tmpDir)
		if err == nil {
			return
		}

		if rerr != nil {
			rerr = errors.Wrapf(rerr, "removing temporary directory error: %s", err)
		} else {
			rerr = errors.WithStack(err)
		}
	}()

	stdout, stderr, err := restoreMySQLFromS3(ctx, j.name, &j.location, tmpDir)
	if err != nil {
		return errors.WithStack(err)
	}

	// TODO: stream stdout, stderr
	_, _ = stdout, stderr

	active, err := mySQLActive()
	if err != nil {
		return errors.WithStack(err)
	}
	if active {
		if err := stopMySQL(); err != nil {
			return errors.WithStack(err)
		}
	}

	if err := restoreBackup(tmpDir, mySQLDirectory); err != nil {
		return errors.WithStack(err)
	}

	if err := startMySQL(); err != nil {
		return errors.WithStack(err)
	}

	send(&agentpb.JobResult{
		JobId:     j.id,
		Timestamp: ptypes.TimestampNow(),
		Result: &agentpb.JobResult_MysqlBackupRestore{
			MysqlBackupRestore: &agentpb.JobResult_MySQLBackupRestore{},
		},
	})

	return nil
}
