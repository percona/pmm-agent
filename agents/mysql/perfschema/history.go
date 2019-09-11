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

package perfschema

import (
	"sync"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/reform.v1"
)

func getHistory(q *reform.Querier) (map[string]*eventsStatementsHistory, error) {
	structs, err := q.SelectAllFrom(eventsStatementsHistoryView, "WHERE DIGEST IS NOT NULL AND SQL_TEXT IS NOT NULL")
	if err != nil {
		return nil, errors.Wrap(err, "failed to query events_statements_history")
	}

	res := make(map[string]*eventsStatementsHistory, len(structs))
	for _, str := range structs {
		esh := str.(*eventsStatementsHistory)
		res[*esh.Digest] = esh
	}
	return res, nil
}

// historyCache provides cached access to performance_schema.events_statements_history.
// It retains data longer than this table.
type historyCache struct {
	retain time.Duration

	rw    sync.RWMutex
	items map[string]*eventsStatementsHistory
	added map[string]time.Time
}

// newHistoryCache creates new historyCache.
func newHistoryCache(retain time.Duration) *historyCache {
	return &historyCache{
		retain: retain,
		items:  make(map[string]*eventsStatementsHistory),
		added:  make(map[string]time.Time),
	}
}

// get returns all current items.
func (c *historyCache) get() map[string]*eventsStatementsHistory {
	c.rw.RLock()
	defer c.rw.RUnlock()

	res := make(map[string]*eventsStatementsHistory, len(c.items))
	for k, v := range c.items {
		res[k] = v
	}
	return res
}

// refresh removes expired items in cache, then adds current items.
func (c *historyCache) refresh(current map[string]*eventsStatementsHistory) {
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
