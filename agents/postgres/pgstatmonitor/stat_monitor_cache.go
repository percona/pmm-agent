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

package pgstatmonitor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/AlekSi/pointer"
	pgquery "github.com/lfittl/pg_query_go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/reform.v1"

	"github.com/percona/pmm-agent/utils/truncate"
)

// statMonitorCacheStats contains statMonitorCache statistics.
type statMonitorCacheStats struct {
	current uint
	oldest  time.Time
	newest  time.Time
}

func (s statMonitorCacheStats) String() string {
	d := s.newest.Sub(s.oldest)
	return fmt.Sprintf("current=%d; %s - %s (%s)",
		s.current,
		s.oldest.UTC().Format("2006-01-02T15:04:05Z"), s.newest.UTC().Format("2006-01-02T15:04:05Z"), d)
}

// statMonitorCache provides cached access to pg_stat_monitor.
// It retains data longer than this table.
type statMonitorCache struct {
	l *logrus.Entry

	rw    sync.RWMutex
	items map[time.Time]map[string]*pgStatMonitorExtended
}

// newStatMonitorCache creates new statMonitorCache.
func newStatMonitorCache(l *logrus.Entry) *statMonitorCache {
	return &statMonitorCache{
		l:     l,
		items: make(map[time.Time]map[string]*pgStatMonitorExtended),
	}
}

// getStatMonitorExtended returns the current state of pg_stat_monitor table with extended information (database, username)
// and the previous cashed state grouped by bucket start time.
func (ssc *statMonitorCache) getStatMonitorExtended(ctx context.Context, q *reform.Querier, normalizedQuery bool) (current, cache map[time.Time]map[string]*pgStatMonitorExtended, err error) {
	var totalN, newN, newSharedN, oldN int
	start := time.Now()
	defer func() {
		dur := time.Since(start)
		ssc.l.Debugf("Selected %d rows from pg_stat_monitor in %s: %d new (%d shared tables), %d old.", totalN, dur, newN, newSharedN, oldN)
	}()

	ssc.rw.RLock()
	current = make(map[time.Time]map[string]*pgStatMonitorExtended)
	cache = make(map[time.Time]map[string]*pgStatMonitorExtended)
	for k, v := range ssc.items {
		cache[k] = v
	}
	ssc.rw.RUnlock()

	// load all databases and usernames first as we can't use querier while iterating over rows below
	databases := queryDatabases(q)
	usernames := queryUsernames(q)

	pgMonitorVersion, e := getPGMonitorVersion(q)
	if e != nil {
		err = errors.Wrap(e, "failed to get pg_stat_monitor version")
		return
	}

	var view reform.View
	var row reform.Struct
	switch {
	case pgMonitorVersion >= 0.9:
		view = pgStatMonitor09View
		row = &pgStatMonitor09{}
	case pgMonitorVersion >= 0.8:
		view = pgStatMonitor08View
		row = &pgStatMonitor08{}
	default:
		view = pgStatMonitorDefaultView
		row = &pgStatMonitorDefault{}
	}

	rows, e := q.SelectRows(view, "WHERE queryid IS NOT NULL AND query IS NOT NULL")
	if e != nil {
		err = errors.Wrap(e, "failed to query pg_stat_monitor")
		return
	}
	defer rows.Close() //nolint:errcheck

	for ctx.Err() == nil {
		if err = q.NextRow(row, rows); err != nil {
			if errors.Is(err, reform.ErrNoRows) {
				err = nil
			}
			break
		}
		totalN++

		var c pgStatMonitorExtended
		switch r := row.(type) {
		case *pgStatMonitorDefault:
			c.pgStatMonitor = r.ToPgStatMonitor()
			c.Database = databases[r.DBID]
			c.Username = usernames[r.UserID]
		case *pgStatMonitor08:
			m, e := r.ToPgStatMonitor()
			if e != nil {
				err = e
				break
			}

			c.pgStatMonitor = m
			c.Database = r.DatName
			c.Username = r.User
		case *pgStatMonitor09:
			m, e := r.ToPgStatMonitor()
			if e != nil {
				err = e
				break
			}

			c.pgStatMonitor = m
			c.Database = r.DatName
			c.Username = r.User
		}

		for _, m := range cache {
			if p, ok := m[c.QueryID]; ok {
				oldN++
				c.Fingerprint = p.Fingerprint
				c.Example = p.Example
				c.IsQueryTruncated = p.IsQueryTruncated
				break
			}
		}

		if c.Fingerprint == "" {
			newN++
			fingerprint := c.Query
			example := ""
			if !normalizedQuery {
				example = c.Query
				fingerprint = ssc.generateFingerprint(c.Query)
			}
			var isTruncated bool
			c.Fingerprint, isTruncated = truncate.Query(fingerprint)
			if isTruncated {
				c.IsQueryTruncated = isTruncated
			}
			c.Example, isTruncated = truncate.Query(example)
			if isTruncated {
				c.IsQueryTruncated = isTruncated
			}
		}

		if current[c.BucketStartTime] == nil {
			current[c.BucketStartTime] = make(map[string]*pgStatMonitorExtended)
		}
		current[c.BucketStartTime][c.QueryID] = &c
	}

	if ctx.Err() != nil {
		err = ctx.Err()
	}

	if err != nil {
		err = errors.Wrap(err, "failed to fetch pg_stat_monitor")
	}

	return current, cache, err
}

func (ssc *statMonitorCache) generateFingerprint(example string) string {
	fingerprint, e := pgquery.Normalize(example)
	if e != nil {
		ssc.l.Errorln(e)
	}
	return fingerprint
}

// stats returns statMonitorCache statistics.
func (ssc *statMonitorCache) stats() statMonitorCacheStats {
	ssc.rw.RLock()
	defer ssc.rw.RUnlock()

	oldest := time.Now()
	var newest time.Time
	for t := range ssc.items {
		if oldest.After(t) {
			oldest = t
		}
		if newest.Before(t) {
			newest = t
		}
	}

	return statMonitorCacheStats{
		current: uint(len(ssc.items[newest])),
		oldest:  oldest,
		newest:  newest,
	}
}

// refresh replaces old cache with a new one.
func (ssc *statMonitorCache) refresh(current map[time.Time]map[string]*pgStatMonitorExtended) {
	ssc.rw.Lock()
	defer ssc.rw.Unlock()

	ssc.items = current
}

func queryDatabases(q *reform.Querier) map[int64]string {
	structs, err := q.SelectAllFrom(pgStatDatabaseView, "")
	if err != nil {
		return nil
	}

	res := make(map[int64]string, len(structs))
	for _, str := range structs {
		d := str.(*pgStatDatabase)
		res[d.DatID] = pointer.GetString(d.DatName)
	}
	return res
}

func queryUsernames(q *reform.Querier) map[int64]string {
	structs, err := q.SelectAllFrom(pgUserView, "")
	if err != nil {
		return nil
	}

	res := make(map[int64]string, len(structs))
	for _, str := range structs {
		u := str.(*pgUser)
		res[u.UserID] = pointer.GetString(u.UserName)
	}
	return res
}
