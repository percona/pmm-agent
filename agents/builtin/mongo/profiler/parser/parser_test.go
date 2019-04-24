package parser

import (
	"reflect"
	"testing"
	"time"

	pm "github.com/percona/percona-toolkit/src/go/mongolib/proto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/percona/pmm-agent/agents/builtin/mongo/profiler/aggregator"
	pc "github.com/percona/pmm-agent/agents/builtin/mongo/proto/config"
	"github.com/percona/pmm-agent/agents/builtin/mongo/proto/qan"
)

func TestNew(t *testing.T) {
	docsChan := make(chan pm.SystemProfile)
	pcQan := pc.QAN{
		Interval: 60,
	}
	a := aggregator.New(time.Now(), pcQan)

	type args struct {
		docsChan   <-chan pm.SystemProfile
		aggregator *aggregator.Aggregator
	}
	tests := []struct {
		name string
		args args
		want *Parser
	}{
		{
			name: "TestNew",
			args: args{
				docsChan:   docsChan,
				aggregator: a,
			},
			want: New(docsChan, a),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.docsChan, tt.args.aggregator); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New(%v, %v) = %v, want %v", tt.args.docsChan, tt.args.aggregator, got, tt.want)
			}
		})
	}
}

func TestParser_StartStop(t *testing.T) {
	var err error
	docsChan := make(chan pm.SystemProfile)
	pcQan := pc.QAN{
		Interval: 60,
	}
	a := aggregator.New(time.Now(), pcQan)

	parser1 := New(docsChan, a)
	err = parser1.Start()
	require.NoError(t, err)

	// running multiple Start() should be idempotent
	err = parser1.Start()
	require.NoError(t, err)

	// running multiple Stop() should be idempotent
	parser1.Stop()
	parser1.Stop()
}

func TestParser_running(t *testing.T) {
	docsChan := make(chan pm.SystemProfile)
	pcQan := pc.QAN{
		Interval: 1,
	}
	a := aggregator.New(time.Now(), pcQan)
	reportChan := a.Start()
	defer a.Stop()
	d := time.Duration(pcQan.Interval) * time.Second

	parser1 := New(docsChan, a)
	err := parser1.Start()
	require.NoError(t, err)

	now := time.Now().UTC()
	timeStart := now.Truncate(d).Add(d)
	timeEnd := timeStart.Add(d)

	select {
	case docsChan <- pm.SystemProfile{
		Ts: timeStart,
		Query: pm.BsonD{
			{"find", "test"},
		},
		ResponseLength: 100,
		DocsExamined:   200,
		Nreturned:      300,
		Millis:         4000,
	}:
	case <-time.After(5 * time.Second):
		t.Error("test timeout")
	}

	sp := pm.SystemProfile{
		Ts: timeEnd.Add(1 * time.Second),
	}
	select {
	case docsChan <- sp:
	case <-time.After(5 * time.Second):
		t.Error("test timeout")
	}

	select {
	case actual := <-reportChan:
		expected := qan.Report{
			StartTs: timeStart,
			EndTs:   timeEnd,
		}
		assert.Equal(t, expected.StartTs, actual.StartTs)
		assert.Equal(t, expected.EndTs, actual.EndTs)
		assert.EqualValues(t, actual.Global.TotalQueries, 1)
		assert.EqualValues(t, actual.Global.UniqueQueries, 1)

		// verify time metrics
		assert.Len(t, actual.Global.Metrics.TimeMetrics, 1)
		assert.NotEmpty(t, actual.Global.Metrics.TimeMetrics["Query_time"])

		// verify number metrics
		assert.Len(t, actual.Global.Metrics.NumberMetrics, 3)
		assert.NotEmpty(t, actual.Global.Metrics.NumberMetrics["Rows_sent"])
		assert.NotEmpty(t, actual.Global.Metrics.NumberMetrics["Rows_examined"])
		assert.NotEmpty(t, actual.Global.Metrics.NumberMetrics["Bytes_sent"])

		// verify bool metrics
		assert.Len(t, actual.Global.Metrics.BoolMetrics, 0)
	case <-time.After(d + 5*time.Second):
		t.Error("test timeout")
	}

	parser1.Stop()
}
