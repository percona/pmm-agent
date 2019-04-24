package collector

import (
	"reflect"
	"testing"
	"time"

	"github.com/percona/percona-toolkit/src/go/mongolib/proto"
	"github.com/percona/pmgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/percona/pmm-agent/agents/builtin/mongo/test/profiling"
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

func TestCollector_StartStop(t *testing.T) {
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

func TestCollector_Stop(t *testing.T) {
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
