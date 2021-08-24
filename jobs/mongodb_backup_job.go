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
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/percona/pmm/api/agentpb"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type pbmLogEntry struct {
	TS         int64 `json:"ts"`
	pbmLogKeys `json:",inline"`
	Msg        string `json:"msg"`
}

type pbmLogKeys struct {
	Severity int    `json:"s"`
	RS       string `json:"rs"`
	Node     string `json:"node"`
	Event    string `json:"e"`
	ObjName  string `json:"eobj"`
	OPID     string `json:"opid,omitempty"`
}

type pbmBackup struct {
	Name    string `json:"name"`
	Storage string `json:"storage"`
}

const (
	pbmBin = "pbm"

	cmdTimeout          = time.Minute
	resyncTimeout       = 5 * time.Minute
	statusCheckInterval = 5 * time.Second
	logsCheckInterval   = 3 * time.Second
	waitForLogs         = 30 * time.Second
)

// This regexp checks that there is no running pbm operations.
var noRunningPBMOperationsRE = regexp.MustCompile(`Currently running:\n=*\n\(none\)`)

// MongoDBBackupJob implements Job from MongoDB backup.
type MongoDBBackupJob struct {
	id       string
	timeout  time.Duration
	l        logrus.FieldLogger
	name     string
	dbURL    *url.URL
	location BackupLocationConfig
}

// NewMongoDBBackupJob creates new Job for MongoDB backup.
func NewMongoDBBackupJob(id string, timeout time.Duration, name string, dbConfig DBConnConfig, locationConfig BackupLocationConfig) *MongoDBBackupJob {
	return &MongoDBBackupJob{
		id:       id,
		timeout:  timeout,
		l:        logrus.WithFields(logrus.Fields{"id": id, "type": "mongodb_backup", "name": name}),
		name:     name,
		dbURL:    createDBURL(dbConfig),
		location: locationConfig,
	}
}

// ID returns Job id.
func (j *MongoDBBackupJob) ID() string {
	return j.id
}

// Type returns Job type.
func (j *MongoDBBackupJob) Type() string {
	return "mongodb_backup"
}

// Timeout returns Job timeout.
func (j *MongoDBBackupJob) Timeout() time.Duration {
	return j.timeout
}

// Run starts Job execution.
func (j *MongoDBBackupJob) Run(ctx context.Context, send Send) error {
	if _, err := exec.LookPath(pbmBin); err != nil {
		return errors.Wrapf(err, "lookpath: %s", pbmBin)
	}

	switch {
	case j.location.S3Config != nil:
		if err := pbmSetupS3(ctx, j.l, j.dbURL, j.name, j.location.S3Config, false); err != nil {
			return errors.Wrap(err, "failed to setup S3 location")
		}
	default:
		return errors.New("unknown location config")
	}

	rCtx, cancel := context.WithTimeout(ctx, resyncTimeout)
	if err := waitForNoRunningPBMOperations(rCtx, j.l, j.dbURL); err != nil {
		cancel()
		return errors.Wrap(err, "failed to wait pbm resync completion")
	}
	cancel()

	pbmBackupOut, err := j.startBackup(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to start backup")
	}
	streamCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		err := j.streamLogs(streamCtx, send, pbmBackupOut.Name)
		if err != nil && err != io.EOF && err != context.Canceled {
			j.l.Errorf("stream logs: %v", err)
		}
	}()

	if err := waitForNoRunningPBMOperations(ctx, j.l, j.dbURL); err != nil {
		return errors.Wrap(err, "failed to wait backup completion")
	}

	send(&agentpb.JobResult{
		JobId:     j.id,
		Timestamp: ptypes.TimestampNow(),
		Result: &agentpb.JobResult_MongodbBackup{
			MongodbBackup: &agentpb.JobResult_MongoDBBackup{},
		},
	})

	select {
	case <-ctx.Done():
	case <-time.After(waitForLogs):
	}
	return nil
}

func createDBURL(dbConfig DBConnConfig) *url.URL {
	var host string
	switch {
	case dbConfig.Address != "":
		if dbConfig.Port > 0 {
			host = net.JoinHostPort(dbConfig.Address, strconv.Itoa(dbConfig.Port))
		} else {
			host = dbConfig.Address
		}
	case dbConfig.Socket != "":
		host = dbConfig.Socket
	}

	var user *url.Userinfo
	if dbConfig.User != "" {
		user = url.UserPassword(dbConfig.User, dbConfig.Password)
	}

	return &url.URL{
		Scheme: "mongodb",
		User:   user,
		Host:   host,
	}
}

func (j *MongoDBBackupJob) startBackup(ctx context.Context) (pbmBackup, error) {
	j.l.Info("Starting backup.")

	nCtx, cancel := context.WithTimeout(ctx, cmdTimeout)
	defer cancel()

	var stderr bytes.Buffer
	var stdout bytes.Buffer
	cmd := exec.CommandContext(nCtx, pbmBin, "backup", "--out=json", "--mongodb-uri="+j.dbURL.String()) // #nosec G204
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return pbmBackup{}, errors.Wrapf(err, "pbm backup error: %s", stderr.String())
	}

	var result pbmBackup
	if err := json.NewDecoder(&stdout).Decode(&result); err != nil {
		return pbmBackup{}, err
	}

	return result, nil
}

func (j *MongoDBBackupJob) streamLogs(ctx context.Context, send Send, name string) error {
	ticker := time.NewTicker(logsCheckInterval)
	defer ticker.Stop()
	var buffer bytes.Buffer
	skip := 0
	for {
		select {
		case <-ticker.C:
			buffer.Reset()
			logs, err := j.retrieveLogs(ctx, name)
			if err != nil {
				return err
			}
			logs = logs[skip:]
			skip += len(logs)
			if len(logs) == 0 {
				continue
			}
			for _, log := range logs {
				_, err = fmt.Fprintf(&buffer, "%s [%s:%s] %s\n", time.Unix(log.TS, 0), log.RS, log.Node, log.Msg)
				if err != nil {
					return err
				}
			}
			send(&agentpb.JobProgress{
				JobId:     j.id,
				Timestamp: ptypes.TimestampNow(),
				Result: &agentpb.JobProgress_Logs_{
					Logs: &agentpb.JobProgress_Logs{
						Out: buffer.Bytes(),
					},
				},
			})
			if logs[len(logs)-1].Msg == "backup finished" {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}

}
func (j *MongoDBBackupJob) retrieveLogs(ctx context.Context, name string) ([]pbmLogEntry, error) {
	nCtx, cancel := context.WithTimeout(ctx, cmdTimeout)
	defer cancel()

	cmd := exec.CommandContext(nCtx, pbmBin, "logs", "--out=json", "--event=backup/"+name, "--mongodb-uri="+j.dbURL.String())

	var stderr bytes.Buffer
	var stdout bytes.Buffer

	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return nil, errors.Wrapf(err, "logs: %s", stderr.String())
	}

	var logs []pbmLogEntry
	if err := json.NewDecoder(&stdout).Decode(&logs); err != nil {
		return nil, err
	}

	return logs, nil
}

func checkRunningPBMOperations(ctx context.Context, l logrus.FieldLogger, dbURL *url.URL) (bool, error) {
	l.Debug("Checking running pbm operations.")

	nCtx, cancel := context.WithTimeout(ctx, cmdTimeout)
	defer cancel()

	output, err := exec.CommandContext(nCtx, pbmBin, "status", "--mongodb-uri="+dbURL.String()).CombinedOutput() // #nosec G204
	if err != nil {
		return false, errors.Wrapf(err, "pbm status error: %s", string(output))
	}

	return noRunningPBMOperationsRE.Match(output), nil
}

func waitForNoRunningPBMOperations(ctx context.Context, l logrus.FieldLogger, dbURL *url.URL) error {
	l.Info("Waiting for pbm operations completion.")

	ticker := time.NewTicker(statusCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			done, err := checkRunningPBMOperations(ctx, l, dbURL)
			if err != nil {
				return errors.Wrapf(err, "failed to check running operations")
			}

			if done {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func pbmSetupS3(ctx context.Context, l logrus.FieldLogger, dbURL *url.URL, prefix string, s3Config *S3LocationConfig, resync bool) error {
	l.Info("Configuring S3 location.")
	nCtx, cancel := context.WithTimeout(ctx, cmdTimeout)
	defer cancel()

	confFile, err := writePBMConfigFile(prefix, s3Config)
	if err != nil {
		return errors.WithStack(err)
	}
	defer os.Remove(confFile) //nolint:errcheck

	output, err := exec.CommandContext( //nolint:gosec
		nCtx,
		pbmBin,
		"config",
		"--mongodb-uri="+dbURL.String(),
		"--file="+confFile,
	).CombinedOutput()

	if err != nil {
		return errors.Wrapf(err, "pbm config error: %s", string(output))
	}

	if resync {
		nCtx, cancel := context.WithTimeout(ctx, cmdTimeout)
		defer cancel()

		output, err = exec.CommandContext( //nolint:gosec
			nCtx,
			pbmBin,
			"config",
			"--mongodb-uri="+dbURL.String(),
			"--force-resync",
		).CombinedOutput()

		if err != nil {
			return errors.Wrapf(err, "pbm config error: %s", string(output))
		}
	}

	return nil
}

func writePBMConfigFile(prefix string, s3Config *S3LocationConfig) (string, error) {
	tmp, err := ioutil.TempFile("", "pbm-config-*.yml")
	if err != nil {
		return "", errors.Wrap(err, "failed to create pbm configuration file")
	}

	var conf struct {
		Storage struct {
			Type string `yaml:"type"`
			S3   struct {
				Region      string `yaml:"region"`
				Bucket      string `yaml:"bucket"`
				Prefix      string `yaml:"prefix"`
				EndpointURL string `yaml:"endpointUrl"`
				Credentials struct {
					AccessKeyID     string `yaml:"access-key-id"`
					SecretAccessKey string `yaml:"secret-access-key"`
				}
			} `yaml:"s3"`
		} `yaml:"storage"`
	}

	conf.Storage.Type = "s3"
	conf.Storage.S3.EndpointURL = s3Config.Endpoint
	conf.Storage.S3.Region = s3Config.BucketRegion
	conf.Storage.S3.Bucket = s3Config.BucketName
	conf.Storage.S3.Prefix = prefix
	conf.Storage.S3.Credentials.AccessKeyID = s3Config.AccessKey
	conf.Storage.S3.Credentials.SecretAccessKey = s3Config.SecretKey

	bytes, err := yaml.Marshal(&conf)
	if err != nil {
		tmp.Close() //nolint:errcheck
		return "", errors.Wrap(err, "failed to marshall pbm configuration")
	}

	if _, err := tmp.Write(bytes); err != nil {
		tmp.Close() //nolint:errcheck
		return "", errors.Wrap(err, "failed to write pbm configuration file")
	}

	return tmp.Name(), tmp.Close()
}
