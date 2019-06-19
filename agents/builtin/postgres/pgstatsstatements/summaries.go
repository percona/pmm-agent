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

package pgstatsstatements

import (
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/reform.v1"
)

func getSummaries(q *reform.Querier) (map[string]*pgStatStatements, error) {
	structs, err := q.SelectAllFrom(pgStatStatementsView, "WHERE queryid IS NOT NULL AND query IS NOT NULL")
	if err != nil {
		return nil, errors.Wrap(err, "failed to query events_statements_summary_by_digest")
	}

	res := make(map[string]*pgStatStatements, len(structs))
	for _, str := range structs {
		ess := str.(*pgStatStatements)

		res[strconv.FormatInt(*ess.Queryid, 10)] = ess
	}
	return res, nil
}

// summaryCache provides cached access to performance_schema.events_statements_summary_by_digest.
// It retains data longer than this table.
type summaryCache struct {
	retain time.Duration

	rw    sync.RWMutex
	items map[string]*pgStatStatements
	added map[string]time.Time
}

// newSummaryCache creates new summaryCache.
func newSummaryCache(retain time.Duration) *summaryCache {
	return &summaryCache{
		retain: retain,
		items:  make(map[string]*pgStatStatements),
		added:  make(map[string]time.Time),
	}
}

// get returns all current items.
func (c *summaryCache) get() map[string]*pgStatStatements {
	c.rw.RLock()
	defer c.rw.RUnlock()

	res := make(map[string]*pgStatStatements, len(c.items))
	for k, v := range c.items {
		res[k] = v
	}
	return res
}

// refresh removes expired items in cache, then adds current items.
func (c *summaryCache) refresh(current map[string]*pgStatStatements) {
	c.rw.Lock()
	defer c.rw.Unlock()

	now := time.Now()

	for k, t := range c.added {
		if now.Sub(t) > c.retain {
			delete(c.items, k)
			delete(c.added, k)
		}
	}

	for k, v := range current {
		c.items[k] = v
		c.added[k] = now
	}
}
