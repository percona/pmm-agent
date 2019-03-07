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

package mysql

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/percona/go-mysql/event" // TODO
	"github.com/percona/pmm/api/qan"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	prometheusNamespace = "pmm_agent"
	prometheusSubsystem = "qan_mysql"
)

// MySQL QAN services connects to MySQL and extracts performance data.
type MySQL struct {
	db *sql.DB
	ch chan<- TODO

	mSend prometheus.Counter
}

// TODO is TODO, d'oh!
type TODO struct {
	mb qan.MetricsBucket
}

// New creates new MySQL QAN service.
func New(db *sql.DB, ch chan<- TODO) *MySQL {
	return &MySQL{
		db: db,
		ch: ch,

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
	t := time.NewTicker(10 * time.Millisecond)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			select {
			case m.ch <- TODO{}:
				m.mSend.Inc()
			case <-ctx.Done():
				// nothing, exit on next iteration
			}
		}
	}
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
