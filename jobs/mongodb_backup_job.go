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
	"net"
	"net/url"
	"os/exec"
	"regexp"
	"strconv"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/percona/pmm/api/agentpb"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const pbmBin = "pbm"

var backupStatusOutputR = regexp.MustCompile("Currently running:\\n=*\\n\\(none\\)")

type MongoDBBackupJob struct {
	id       string
	timeout  time.Duration
	l        logrus.FieldLogger
	name     string
	dbURL    url.URL
	location BackupLocationConfig
}

// NewMongoDBBackupJob creates new Job for MongoDB backup.
func NewMongoDBBackupJob(id string, timeout time.Duration, name string, dbConfig DatabaseConfig, locationConfig BackupLocationConfig) *MongoDBBackupJob {
	return &MongoDBBackupJob{
		id:      id,
		timeout: timeout,
		l:       logrus.WithFields(logrus.Fields{"id": id, "type": "mongodb_backup", "name": name}),
		name:    name,
		dbURL: url.URL{
			Scheme: "mongodb",
			User:   url.UserPassword(dbConfig.User, dbConfig.Password),
			Host:   net.JoinHostPort(dbConfig.Address, strconv.Itoa(dbConfig.Port)),
		},
		location: locationConfig,
	}
}

// DatabaseConfig contains required properties for connection to DB.
type DatabaseConfig struct {
	User     string
	Password string
	Address  string
	Port     int
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

func (j *MongoDBBackupJob) ID() string {
	return j.id
}

func (j *MongoDBBackupJob) Type() string {
	return "mongodb_backup"
}

func (j *MongoDBBackupJob) Timeout() time.Duration {
	return j.timeout
}

func (j *MongoDBBackupJob) Run(ctx context.Context, send Send) error {
	t := time.Now()
	j.l.Info("MySQL backup started")

	if _, err := exec.LookPath(pbmBin); err != nil {
		return errors.Wrapf(err, "lookpath: %s", pbmBin)
	}

	switch {
	case j.location.S3Config != nil:
		if err := j.setupS3(ctx); err != nil {
			return errors.Wrap(err, "failed to setup S3 location")
		}
	default:
		return errors.New("unknown location config")
	}

	if err := j.startBackup(ctx); err != nil {
		return errors.Wrap(err, "failed to start backup")
	}

	if err := j.waitUntilBackupCompletion(ctx); err != nil {
		return errors.Wrap(err, "failed to wait backup completion")
	}

	send(&agentpb.JobResult{
		JobId:     j.id,
		Timestamp: ptypes.TimestampNow(),
		Result: &agentpb.JobResult_MongodbBackup{
			MongodbBackup: &agentpb.JobResult_MongoDBBackup{},
		},
	})
}

func (j *MongoDBBackupJob) startBackup(ctx context.Context) error {
	output, err := exec.CommandContext(
		ctx,
		pbmBin,
		"backup",
		"--mongodb-uri="+j.dbURL.String(),
	).CombinedOutput()

	if err != nil {
		return errors.Wrapf(err, "pbm backup error: %s", string(output))
	}

	return nil
}

func (j *MongoDBBackupJob) waitUntilBackupCompletion(ctx context.Context) error {
	ticker := time.NewTicker(5 * time.Second)

	cmd := exec.CommandContext(
		ctx,
		pbmBin,
		"status",
		"--mongodb-uri="+j.dbURL.String(),
	)

	for {
		select {
		case <-ticker.C:
			output, err := cmd.CombinedOutput()
			if err != nil {
				return errors.Wrapf(err, "pbm status error: %s", string(output))
			}

			if backupStatusOutputR.Match(output) {
				break
			}

		case <-ctx.Done():
			return ctx.Err()
		}
	}

}

func (j *MongoDBBackupJob) setupS3(ctx context.Context) error {
	return exec.CommandContext( //nolint:gosec
		ctx,
		pbmBin,
		"config",
		"--set.storage.type=s3",
		"--set=storage.s3.prefix="+j.name,
		"--set=storage.s3.region="+j.location.S3Config.BucketRegion,
		"--set=storage.s3.bucket="+j.location.S3Config.BucketName,
		"--set=storage.s3.credentials.access-key-id="+j.location.S3Config.AccessKey,
		"--set=storage.s3.credentials.secret-access-key="+j.location.S3Config.SecretKey,
		"--set=storage.s3.endpointUrl="+j.location.S3Config.Endpoint,
	).Run()
}
