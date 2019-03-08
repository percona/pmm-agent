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

// Package mysql runs built-in QAN Agent for MySQL.
package mysql

import (
	"context"
	"time"

	"github.com/AlekSi/pointer"
	"github.com/percona/pmm/api/inventory"
	qanpb "github.com/percona/pmm/api/qan"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"gopkg.in/reform.v1"
)

/*
TODO A ton of open questions:
* check that performance schema is enabled?
* check that statement_digest consumer is enabled?
* TRUNCATE events_statements_summary_by_digest before reading?
* check/report the value of performance_schema_digests_size?
* report rows with NULL digest?
* get query by digest from events_statements_history_long?
* check/report the value of performance_schema_events_statements_history_long_size?
* how often to select data?
* condition for FIRST_SEEN / LAST_SEEN?
* should github.com/percona/go-mysql/event be used?
*/

const (
	prometheusNamespace = "pmm_agent"
	prometheusSubsystem = "qan_mysql"
)

// MySQL QAN services connects to MySQL and extracts performance data.
type MySQL struct {
	db      *reform.DB
	ch      chan<- qanpb.AgentMessage
	l       *logrus.Entry
	changes chan inventory.AgentStatus

	mSend prometheus.Counter
}

// New creates new MySQL QAN service.
func New(db *reform.DB, ch chan<- qanpb.AgentMessage) *MySQL {
	return &MySQL{
		db:      db,
		ch:      ch,
		l:       logrus.WithField("component", "mysql"),
		changes: make(chan inventory.AgentStatus, 1),

		mSend: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: prometheusNamespace,
			Subsystem: prometheusSubsystem,
			Name:      "TODOs_sent_total",
			Help:      "A total number of TODOs sent.",
		}),
	}
}

// Run extracts performance data and sends it to the channel until ctx is canceled.
func (m *MySQL) Run(ctx context.Context) {
	t := time.NewTicker(time.Second)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			todos := m.get(ctx)
			for _, t := range todos {
				select {
				case <-ctx.Done():
					return
				case m.ch <- t:
					// nothing
				}
			}
		}
	}
}

func (m *MySQL) get(ctx context.Context) []qanpb.AgentMessage {
	structs, err := m.db.SelectAllFrom(eventsStatementsSummaryByDigestView, "")
	if err != nil {
		m.l.Error(err)
		return nil
	}

	res := make([]qanpb.AgentMessage, 0, len(structs))
	for _, str := range structs {
		ess := str.(*eventsStatementsSummaryByDigest)
		if ess.Digest == nil || ess.DigestText == nil {
			m.l.Warnf("Skipping %s.", ess)
			continue
		}

		m := qanpb.AgentMessage{
			MetricsBucket: []*qanpb.MetricsBucket{{
				Queryid:     *ess.Digest,
				Fingerprint: *ess.DigestText,
				DSchema:     pointer.GetString(ess.SchemaName),
			}},
		}
		res = append(res, m)
	}
	return res
}

// Changes returns channel that should be read until it is closed.
func (m *MySQL) Changes() <-chan inventory.AgentStatus {
	return m.changes
}

// Describe implements prometheus.Collector.
func (m *MySQL) Describe(ch chan<- *prometheus.Desc) {
	m.mSend.Describe(ch)
}

// Collect implement prometheus.Collector.
func (m *MySQL) Collect(ch chan<- prometheus.Metric) {
	m.mSend.Collect(ch)
}

// check interfaces
var (
	_ prometheus.Collector = (*MySQL)(nil)
)
