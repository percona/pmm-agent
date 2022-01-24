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

package cache

import (
	"fmt"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type statementsMap map[interface{}]interface{}
type statementsAddedMap map[interface{}]time.Time

// Cache provides cached access to performance statistics tables.
// It retains data longer than those tables.
// Intended to store various subtypes of map.
type Cache struct {
	typ       reflect.Type
	items     statementsMap
	added     statementsAddedMap
	retain    time.Duration
	sizeLimit uint
	l         *logrus.Entry
	rw        sync.RWMutex
	updatedN  uint
	addedN    uint
	removedN  uint
}

// New creates new Cache.
// Argument typ is an instance of type to be stored in Cache, must be a map with chosen key and value types.
func New(typ interface{}, retain time.Duration, sizeLimit uint, l *logrus.Entry) *Cache {
	return &Cache{
		typ:       reflect.TypeOf(typ),
		retain:    retain,
		sizeLimit: sizeLimit,
		l:         l,
		items:     make(statementsMap),
		added:     make(statementsAddedMap),
	}
}

// Get returns all current items if the cache.
func (c *Cache) Get(dest interface{}) {
	if reflect.TypeOf(dest) != c.typ {
		panic(fmt.Sprintf("Wrong argument type. Must be %v, got %v", c.typ, reflect.TypeOf(dest)))
	}
	c.rw.RLock()
	defer c.rw.RUnlock()

	m := reflect.ValueOf(dest)
	for k, v := range c.items {
		m.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v))
	}
}

// Refresh removes expired items in cache, then adds current items, then trims the cache if it's length is more than specified.
func (c *Cache) Refresh(current interface{}) {
	if reflect.TypeOf(current) != c.typ {
		panic(fmt.Sprintf("Wrong argument type. Must be %v, got %v", c.typ, reflect.TypeOf(current)))
	}

	c.rw.Lock()
	defer c.rw.Unlock()
	now := time.Now()

	for k, t := range c.added {
		if now.Sub(t) > c.retain {
			c.removedN++
			delete(c.items, k)
			delete(c.added, k)
		}
	}

	m := reflect.ValueOf(current)
	for _, k := range m.MapKeys() {
		key := k.Interface()
		value := m.MapIndex(k).Interface()
		if _, ok := c.items[key]; ok {
			c.updatedN++
		} else {
			c.addedN++
		}
		c.items[key] = value
		c.added[key] = now
	}

	cacheItemsN := uint(len(c.added))
	if cacheItemsN > c.sizeLimit {
		overSize := cacheItemsN - c.sizeLimit
		itemList := make(sortByTimeSlice, cacheItemsN)
		i := 0
		for k, v := range c.added {
			itemList[i] = sortByTime{k, v}
			i++
		}
		sort.Sort(itemList)

		for _, item := range itemList[0:overSize] {
			c.removedN++
			delete(c.items, item.key)
			delete(c.added, item.key)
		}
		c.l.Debugf("Cache size exceeded the limit of %d items and the oldest values were trimmed. "+
			"Now the oldest query in the cache is of time %s",
			c.sizeLimit, itemList[overSize].t.UTC().Format("2006-01-02T15:04:05Z"))
	}
}

// Stats returns CacheStats statistics.
func (c *Cache) Stats() CacheStats {
	c.rw.RLock()
	defer c.rw.RUnlock()

	oldest := time.Now().Add(c.retain)
	var newest time.Time
	for _, t := range c.added {
		if oldest.After(t) {
			oldest = t
		}
		if newest.Before(t) {
			newest = t
		}
	}

	return CacheStats{
		current:  uint(len(c.added)),
		updatedN: c.updatedN,
		addedN:   c.addedN,
		removedN: c.removedN,
		oldest:   oldest,
		newest:   newest,
	}
}

func (c *Cache) Len() int {
	return len(c.items)
}

// CacheStats contains Cache statistics.
type CacheStats struct {
	current  uint
	updatedN uint
	addedN   uint
	removedN uint
	oldest   time.Time
	newest   time.Time
}

func (s CacheStats) String() string {
	d := s.newest.Sub(s.oldest)
	return fmt.Sprintf("current=%d: updated=%d added=%d removed=%d; %s - %s (%s)",
		s.current, s.updatedN, s.addedN, s.removedN,
		s.oldest.UTC().Format("2006-01-02T15:04:05Z"), s.newest.UTC().Format("2006-01-02T15:04:05Z"), d)
}

// sortByTime used for sorting cache elements by add time.
type sortByTime struct {
	key interface{}
	t   time.Time
}

type sortByTimeSlice []sortByTime

func (s sortByTimeSlice) Len() int           { return len(s) }
func (s sortByTimeSlice) Less(i, j int) bool { return s[i].t.Before(s[j].t) }
func (s sortByTimeSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
