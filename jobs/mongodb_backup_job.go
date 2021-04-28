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
		id:       id,
		timeout:  timeout,
		l:        logrus.WithFields(logrus.Fields{"id": id, "type": "mongodb_backup", "name": name}),
		name:     name,
		dbURL:    createDBURL(dbConfig),
		location: locationConfig,
	}
}

func createDBURL(dbConfig DatabaseConfig) url.URL {
	var host string
	switch {
	case dbConfig.Address != "":
		if dbConfig.Port > 0 {
			host = net.JoinHostPort(dbConfig.Address, strconv.Itoa(dbConfig.Port))
		} else {
			host = dbConfig.Address
		}
	case dbConfig.Socket != "":
		host = url.QueryEscape(dbConfig.Socket)
	}

	var user *url.Userinfo
	if dbConfig.User != "" {
		user = url.UserPassword(dbConfig.User, dbConfig.Password)
	}

	return url.URL{
		Scheme: "mongodb",
		User:   user,
		Host:   host,
	}
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

	return nil
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
	defer ticker.Stop()

loop:
	for {
		select {
		case <-ticker.C:
			output, err := exec.CommandContext(
				ctx,
				pbmBin,
				"status",
				"--mongodb-uri="+j.dbURL.String(),
			).CombinedOutput()
			if err != nil {
				return errors.Wrapf(err, "pbm status error: %s", string(output))
			}

			if backupStatusOutputR.Match(output) {
				break loop
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

func (j *MongoDBBackupJob) setupS3(ctx context.Context) error {
	output, err := exec.CommandContext( //nolint:gosec
		ctx,
		pbmBin,
		"config",
		"--mongodb-uri="+j.dbURL.String(),
		"--set=storage.type=s3",
		"--set=storage.s3.prefix="+j.name,
		"--set=storage.s3.region="+j.location.S3Config.BucketRegion,
		"--set=storage.s3.bucket="+j.location.S3Config.BucketName,
		"--set=storage.s3.credentials.access-key-id="+j.location.S3Config.AccessKey,
		"--set=storage.s3.credentials.secret-access-key="+j.location.S3Config.SecretKey,
		"--set=storage.s3.endpointUrl="+j.location.S3Config.Endpoint,
	).CombinedOutput()

	if err != nil {
		return errors.Wrapf(err, "pbm config error: %s", string(output))
	}

	return nil
}
