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

package collector

import (
	"reflect"
	"testing"
	"time"

	"github.com/percona/percona-toolkit/src/go/mongolib/proto"
	"github.com/percona/pmgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/percona/qan-agent/test/profiling"
)

func TestNew(t *testing.T) {
	t.Parallel()

	dialer := pmgo.NewDialer()
	dialInfo, _ := pmgo.ParseURL("127.0.0.1:27017")
	session, err := dialer.DialWithInfo(dialInfo)
	require.NoError(t, err)

	type args struct {
		session pmgo.SessionManager
		dbName  string
	}
	tests := []struct {
		name string
		args args
		want *Collector
	}{
		{
			name: "127.0.0.1:27017",
			args: args{
				session: session,
				dbName:  "",
			},
			want: &Collector{
				session: session,
				dbName:  "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.session, tt.args.dbName); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New(%v, %v) = %v, want %v", tt.args.session, tt.args.dbName, got, tt.want)
			}
		})
	}
}

func TestCollectorStartStop(t *testing.T) {
	t.Parallel()

	dialer := pmgo.NewDialer()
	dialInfo, _ := pmgo.ParseURL("127.0.0.1:27017")
	session, err := dialer.DialWithInfo(dialInfo)
	require.NoError(t, err)

	collector1 := New(session, "")
	docsChan, err := collector1.Start()
	require.NoError(t, err)
	assert.NotNil(t, docsChan)

	defer collector1.Stop()
}

func TestCollectorStop(t *testing.T) {
	t.Parallel()

	dialer := pmgo.NewDialer()
	dialInfo, _ := pmgo.ParseURL("127.0.0.1:27017")
	session, err := dialer.DialWithInfo(dialInfo)
	require.NoError(t, err)

	// #1
	notStarted := New(session, "")

	// #2
	started := New(session, "")
	_, err = started.Start()
	require.NoError(t, err)

	tests := []struct {
		name string
		self *Collector
	}{
		{
			name: "not started",
			self: notStarted,
		},
		{
			name: "started",
			self: started,
		},
		// repeat to be sure Stop() is idempotent
		{
			name: "not started",
			self: notStarted,
		},
		{
			name: "started",
			self: started,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.self.Stop()
		})
	}
}

func TestCollector(t *testing.T) {
	// Disable profiling.
	err := profiling.New("").DisableAll()
	require.NoError(t, err)
	// Enable profiling for default db.
	err = profiling.New("").Enable("")
	require.NoError(t, err)

	// create separate connection to db for collector
	dialer := pmgo.NewDialer()
	dialInfo, _ := pmgo.ParseURL("")
	session, err := dialer.DialWithInfo(dialInfo)
	require.NoError(t, err)

	// create collector
	collector := New(session, "")
	docsChan, err := collector.Start()
	require.NoError(t, err)
	defer collector.Stop()

	// add some data to mongo e.g. people
	people := []map[string]string{
		{"name": "Kamil"},
		{"name": "Carlos"},
	}

	// add data through separate connection
	session, err = dialer.DialWithInfo(dialInfo)
	require.NoError(t, err)
	for _, person := range people {
		err = session.DB("test").C("people").Insert(&person)
		require.NoError(t, err)
	}

	actual := []proto.SystemProfile{}
F:
	for {
		select {
		case doc, ok := <-docsChan:
			if !ok {
				break F
			}
			if doc.Ns == "test.people" && doc.Op == "insert" {
				actual = append(actual, doc)
			}
			if len(actual) == len(people) {
				// stopping collector should also close docsChan
				collector.Stop()
			}
		case <-time.After(10 * time.Second):
			t.Fatal("didn't recieve enough samples before timeout")
		}
	}
	assert.Len(t, actual, len(people))
}
