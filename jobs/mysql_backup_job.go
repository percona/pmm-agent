package jobs

import (
	"bytes"
	"context"
	"github.com/go-sql-driver/mysql"
	"github.com/golang/protobuf/ptypes"
	"github.com/percona/pmm/api/agentpb"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"time"
)

type MySQLBackupJob struct {
	id       string
	timeout  time.Duration
	l        *logrus.Entry
	dsn      string
	location BackupLocationConfig
}

// S3LocationConfig contains required properties for accessing S3 Bucket.
type S3LocationConfig struct {
	Endpoint   string
	AccessKey  string
	SecretKey  string
	BucketName string
}

// BackupLocationConfig groups all backup locations configs.
type BackupLocationConfig struct {
	S3Config *S3LocationConfig
}

func NewMySQLBackupJob(id string, timeout time.Duration, dsn string, locationConfig BackupLocationConfig) *MySQLBackupJob {
	return &MySQLBackupJob{
		id:       id,
		timeout:  timeout,
		l:        logrus.WithFields(logrus.Fields{"id": id, "type": "mysql_backup"}),
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

// Timeouts returns job timeout.
func (j *MySQLBackupJob) Timeout() time.Duration {
	return j.timeout
}

func (j *MySQLBackupJob) Run(ctx context.Context, send Send) error {
	mysqlConfig, err := mysql.ParseDSN(j.dsn)
	if err != nil {
		return err
	}
	// @TODO from params
	backupName := "backup-" + time.Now().Format(time.RFC3339)
	tmpDir := os.TempDir()
	xtrabackupCmd := exec.CommandContext(ctx, "xtrabackup",
		"--user="+mysqlConfig.User,
		"--password="+mysqlConfig.Passwd,
		"--port="+mysqlConfig.Passwd,
		"--host="+mysqlConfig.Addr,
		"--backup",
		"--stream=xbstream",
		"--extra-lsndir="+tmpDir,
		"--target-dir="+tmpDir)

	var xbcloudCmd *exec.Cmd
	switch {
	case j.location.S3Config != nil:
		xbcloudCmd = exec.CommandContext(ctx, "xbcloud",
			"put",
			"--storage=s3",
			"--s3-endpoint="+j.location.S3Config.Endpoint,
			"--s3-access-key="+j.location.S3Config.AccessKey,
			"--s3-secret-key="+j.location.S3Config.SecretKey,
			"--s3-bucket="+j.location.S3Config.BucketName,
			// @TODO region from parameters
			"--s3-region="+"us-east-2",
			"--parallel=10",
			backupName)
	default:
		return errors.Errorf("unknown location config")
	}

	var outBuffer bytes.Buffer
	var errBackupBuffer bytes.Buffer
	var errCloudBuffer bytes.Buffer
	xtrabackupCmd.Stderr = &errBackupBuffer

	xtrabackupStdout, err := xtrabackupCmd.StdoutPipe()
	if err != nil {
		return err
	}

	if xbcloudCmd != nil {
		xbcloudCmd.Stdin = xtrabackupStdout
		xbcloudCmd.Stdout = &outBuffer
		xbcloudCmd.Stderr = &errCloudBuffer
		if err := xbcloudCmd.Start(); err != nil {
			return err
		}
	}

	if err := xtrabackupCmd.Start(); err != nil {
		return err
	}

	if xbcloudCmd != nil {
		if err := xbcloudCmd.Wait(); err != nil {
			j.l.Infoln(errCloudBuffer.String())
			return err
		}
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
