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
	"io"
	"net"
	"net/url"
	"os/exec"
	"strconv"
	"time"

	"github.com/percona/pmm/api/agentpb"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	pbmBin = "pbm"

	logsCheckInterval = 3 * time.Second
	waitForLogs       = 3 * time.Second
)

// MongoDBBackupJob implements Job from MongoDB backup.
type MongoDBBackupJob struct {
	id         string
	timeout    time.Duration
	l          logrus.FieldLogger
	name       string
	dbURL      *url.URL
	location   BackupLocationConfig
	logChunkID uint32
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
	if err := waitForPBMState(rCtx, j.l, j.dbURL, pbmNoRunningOperations); err != nil {
		cancel()
		return errors.Wrap(err, "failed to wait pbm resync completion")
	}
	cancel()

	pbmBackupOut, err := j.startBackup(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to start backup")
	}
	backupFinished := make(chan struct{}, 1)
	streamCtx, streamCancel := context.WithCancel(ctx)
	defer streamCancel()
	go func() {
		err := j.streamLogs(streamCtx, send, pbmBackupOut.Name, backupFinished)
		if err != nil && err != io.EOF && err != context.Canceled {
			j.l.Errorf("stream logs: %v", err)
		}
		send(&agentpb.JobProgress{
			JobId:     j.id,
			Timestamp: timestamppb.Now(),
			Result: &agentpb.JobProgress_Logs_{
				Logs: &agentpb.JobProgress_Logs{
					Done:    true,
					ChunkId: j.logChunkID,
					Time:    timestamppb.Now(),
				},
			},
		})
	}()

	if err := waitForPBMState(ctx, j.l, j.dbURL, pbmBackupFinished(pbmBackupOut.Name)); err != nil {
		return errors.Wrap(err, "failed to wait backup completion")
	}
	backupFinished <- struct{}{}
	send(&agentpb.JobResult{
		JobId:     j.id,
		Timestamp: timestamppb.Now(),
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

func (j *MongoDBBackupJob) startBackup(ctx context.Context) (*pbmBackup, error) {
	j.l.Info("Starting backup.")
	var result pbmBackup

	if err := getPBMOutput(ctx, j.dbURL, &result, "backup"); err != nil {
		return nil, err
	}

	return &result, nil
}

func (j *MongoDBBackupJob) streamLogs(ctx context.Context, send Send, name string, backupFinished <-chan struct{}) error {
	var (
		err        error
		backupDone bool
		logs       []pbmLogEntry
		buffer     bytes.Buffer
		skip       int
		lastLog    pbmLogEntry
	)
	j.logChunkID = 0
	finished := func() bool {
		if lastLog.Msg == "backup finished" {
			return true
		}
		if backupDone && len(logs) == 0 {
			return true
		}
		return false
	}

	ticker := time.NewTicker(logsCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-backupFinished:
			backupDone = true
		case <-ticker.C:
			logs, err = retrieveLogs(ctx, j.dbURL, "backup/"+name)
			if err != nil {
				return err
			}
			t := timestamppb.Now()
			logs = logs[skip:]
			skip += len(logs)
			if len(logs) == 0 {
				continue
			}
			from, to := 0, maxLogsChunkSize
			for from < len(logs) {
				if to > len(logs) {
					to = len(logs)
				}
				buffer.Reset()
				for i, log := range logs[from:to] {
					_, err := buffer.WriteString(log.String())
					if err != nil {
						return err
					}
					if i != to-from-1 {
						buffer.WriteRune('\n')
					}
				}
				send(&agentpb.JobProgress{
					JobId:     j.id,
					Timestamp: timestamppb.Now(),
					Result: &agentpb.JobProgress_Logs_{
						Logs: &agentpb.JobProgress_Logs{
							ChunkId: j.logChunkID,
							Message: buffer.String(),
							Time:    t,
						},
					},
				})
				j.logChunkID++
				from += maxLogsChunkSize
				to += maxLogsChunkSize
			}
			if finished() {
				return nil
			}
			lastLog = logs[len(logs)-1]
		case <-ctx.Done():
			return ctx.Err()
		}
	}

}
