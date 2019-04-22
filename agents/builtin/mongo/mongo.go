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

// Package mongo runs built-in QAN Agent for Mongo profiler.
package mongo

import (
	"context"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql" // register SQL driver
	"github.com/percona/pmm/api/inventorypb"
	"github.com/percona/pmm/api/qanpb"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	retainHistory  = 5 * time.Minute
	refreshHistory = 5 * time.Second

	retainSummaries = 25 * time.Hour // make it work for daily queries
	querySummaries  = time.Minute
)

// Mongo extracts performance data from Mongo op log.
type Mongo struct {
	db      *mongo.Client
	l       *logrus.Entry
	changes chan Change
}

// Params represent Agent parameters.
type Params struct {
	DSN     string
	AgentID string
}

// Change represents Agent status change _or_ QAN collect request.
type Change struct {
	Status  inventorypb.AgentStatus
	Request *qanpb.CollectRequest
}

// New creates new MySQL QAN service.
func New(params *Params, l *logrus.Entry) (*Mongo, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(params.DSN))
	if err != nil {
		return nil, err
	}

	return newMongo(client, l), nil
}

func newMongo(db *mongo.Client, l *logrus.Entry) *Mongo {
	return &Mongo{
		db:      db,
		l:       l,
		changes: make(chan Change, 10),
	}
}

// Run extracts performance data and sends it to the channel until ctx is canceled.
func (m *Mongo) Run(ctx context.Context) {
	defer func() {
		m.db.Disconnect(ctx) //nolint:errcheck
		m.changes <- Change{Status: inventorypb.AgentStatus_DONE}
		close(m.changes)
	}()

	m.changes <- Change{Status: inventorypb.AgentStatus_STARTING}

	err := m.db.Connect(ctx)
	if err != nil {
		m.l.Error(err)
		m.changes <- Change{Status: inventorypb.AgentStatus_STOPPING}
		return
	}

	m.changes <- Change{Status: inventorypb.AgentStatus_RUNNING}
	for {
		select {
		case <-ctx.Done():
			m.changes <- Change{Status: inventorypb.AgentStatus_STOPPING}
			return
		}
		// do some stuff
	}
}

func (m *Mongo) getNewBuckets(periodStart time.Time, periodLength time.Duration) ([]*qanpb.MetricsBucket, error) {
	return makeBuckets()
}

// makeBuckets XXX.
//
// makeBuckets is a pure function for easier testing.
func makeBuckets() ([]*qanpb.MetricsBucket, error) {
	return nil, fmt.Errorf("not implemented yet")
}

// Changes returns channel that should be read until it is closed.
func (m *Mongo) Changes() <-chan Change {
	return m.changes
}
