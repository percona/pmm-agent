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
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

// random struct to test the cache
type someType struct {
	field1 int
	field2 string
	field3 innerStruct
}

type innerStruct struct{ field float64 }

func TestCache(t *testing.T) {
	set1 := map[int64]*someType{
		1: new(someType),
		2: new(someType),
		3: new(someType),
		4: new(someType),
		5: new(someType)}

	set2 := map[int64]*someType{
		1: new(someType),
		2: new(someType),
		3: new(someType),
		4: new(someType),
		6: new(someType)}

	t.Run("DoesntReachLimits", func(t *testing.T) {
		c := New(map[int64]*someType{}, time.Second*60, 100, logrus.WithField("test", t.Name()))

		now1 := time.Now()
		c.Refresh(set1)
		stats := c.Stats()
		actual := make(map[int64]*someType)
		c.Get(actual)

		assert.True(t, reflect.DeepEqual(actual, set1))

		assert.Equal(t, uint(5), stats.current)
		assert.Equal(t, uint(0), stats.updatedN)
		assert.Equal(t, uint(5), stats.addedN)
		assert.Equal(t, uint(0), stats.removedN)
		assert.InDelta(t, 0, int(stats.oldest.Sub(now1).Seconds()), 1)
		assert.InDelta(t, 0, int(stats.newest.Sub(now1).Seconds()), 1)

		time.Sleep(time.Second * 1)

		now2 := time.Now()
		c.Refresh(set2)
		stats = c.Stats()
		actual = make(map[int64]*someType)
		c.Get(actual)

		expected := make(map[int64]*someType)
		for k, v := range set1 {
			expected[k] = v
		}
		expected[6] = new(someType)

		assert.True(t, reflect.DeepEqual(actual, expected))
		assert.Equal(t, uint(6), stats.current)
		assert.Equal(t, uint(4), stats.updatedN)
		assert.Equal(t, uint(6), stats.addedN)
		assert.Equal(t, uint(0), stats.removedN)
		assert.InDelta(t, 0, int(stats.oldest.Sub(now1).Seconds()), 0.01)
		assert.InDelta(t, 0, int(stats.newest.Sub(now2).Seconds()), 0.01)
	})

	t.Run("ReachesTimeLimit", func(t *testing.T) {
		c := New(map[int64]*someType{}, time.Second*1, 100, logrus.WithField("test", t.Name()))

		c.Refresh(set1)
		time.Sleep(time.Second * 1)
		now := time.Now()
		c.Refresh(set2)
		stats := c.Stats()
		actual := make(map[int64]*someType)
		c.Get(actual)

		expected := make(map[int64]*someType)
		for k, v := range set2 {
			expected[k] = v
		}

		assert.True(t, reflect.DeepEqual(actual, expected))
		assert.Equal(t, uint(5), stats.current)
		assert.Equal(t, uint(0), stats.updatedN)
		assert.Equal(t, uint(10), stats.addedN)
		assert.Equal(t, uint(5), stats.removedN)
		assert.InDelta(t, 0, int(stats.oldest.Sub(now).Seconds()), 0.01)
		assert.InDelta(t, 0, int(stats.newest.Sub(now).Seconds()), 0.01)
	})

	t.Run("ReachesSizeLimit", func(t *testing.T) {
		c := New(map[int64]*someType{}, time.Second*60, 5, logrus.WithField("test", t.Name()))

		c.Refresh(set1)
		time.Sleep(time.Second * 1)
		now := time.Now()
		c.Refresh(set2)
		stats := c.Stats()
		actual := make(map[int64]*someType)
		c.Get(actual)

		assert.Equal(t, uint(5), stats.current)
		assert.Equal(t, uint(4), stats.updatedN)
		assert.Equal(t, uint(6), stats.addedN)
		assert.Equal(t, uint(1), stats.removedN)
		assert.InDelta(t, 0, int(stats.oldest.Sub(now).Seconds()), 0.01)
		assert.InDelta(t, 0, int(stats.newest.Sub(now).Seconds()), 0.01)
	})
}
