// pmm-agent
// Copyright (C) 2018 Percona LLC
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package sender

// TODO: Refactor or remove
//import (
//	"reflect"
//	"testing"
//
//	"github.com/sirupsen/logrus"
//	"github.com/stretchr/testify/require"
//
//	"github.com/percona/pmm-agent/agents/builtin/mongo/proto/qan"
//	"github.com/percona/pmm-agent/agents/builtin/mongo/test/mock"
//)
//
//func TestNew(t *testing.T) {
//	reportChan := make(chan *qan.Report)
//	dataChan := make(chan *qan.Report)
//	spool := mock.NewSpooler(dataChan)
//	logger := logrus.WithField("component", "sender-test")
//	sender1 := New(reportChan, spool, logger)
//
//	type args struct {
//		reportChan <-chan *qan.Report
//		spool      Spooler
//		logger     *logrus.Entry
//	}
//	tests := []struct {
//		name string
//		args args
//		want *Sender
//	}{
//		{
//			name: "TestNew",
//			args: args{
//				reportChan: reportChan,
//				spool:      spool,
//				logger:     logger,
//			},
//			want: sender1,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if got := New(tt.args.reportChan, tt.args.spool, tt.args.logger); !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("New(%v, %v, %v) = %v, want %v", tt.args.reportChan, tt.args.spool, tt.args.logger, got, tt.want)
//			}
//		})
//	}
//}
//
//func TestSender_Start(t *testing.T) {
//	reportChan := make(chan *qan.Report)
//	dataChan := make(chan *qan.Report)
//	spool := mock.NewSpooler(dataChan)
//	logger := logrus.WithField("component", "sender-test")
//	sender1 := New(reportChan, spool, logger)
//
//	// start sender
//	err := sender1.Start()
//	require.NoError(t, err)
//
//	// running multiple Start() should be idempotent
//	err = sender1.Start()
//	require.NoError(t, err)
//
//	// running multiple Stop() should be idempotent
//	sender1.Stop()
//	sender1.Stop()
//}
