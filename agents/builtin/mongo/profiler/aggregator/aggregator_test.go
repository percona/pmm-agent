package aggregator

import (
	"testing"
	"time"

	"github.com/percona/go-mysql/event"
	"github.com/percona/percona-toolkit/src/go/mongolib/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pc "github.com/percona/pmm-agent/agents/builtin/mongo/proto/config"
	"github.com/percona/pmm-agent/agents/builtin/mongo/proto/qan"
)

func TestAggregator_Add(t *testing.T) {
	t.Parallel()

	timeStart, err := time.Parse("2006-01-02 15:04:05", "2017-07-02 07:55:00")
	require.NoError(t, err)
	timeEnd, err := time.Parse("2006-01-02 15:04:05", "2017-07-02 07:56:00")
	require.NoError(t, err)

	config := pc.QAN{
		UUID:     "abc",
		Interval: 60, // 60s
	}

	aggregator := New(timeStart, config)
	reportChan := aggregator.Start()
	defer aggregator.Stop()

	{
		doc := proto.SystemProfile{
			Ts:           timeStart,
			Millis:       1000,
			DocsExamined: 13,
			Nreturned:    42,
		}
		err := aggregator.Add(doc)
		require.NoError(t, err)
		select {
		case report := <-reportChan:
			t.Error("didn't expect report but got:", report)
		default:
		}
	}

	{
		doc := proto.SystemProfile{
			Ts: timeEnd,
		}
		expected := qan.Report{
			UUID:    config.UUID,
			StartTs: timeStart,
			EndTs:   timeEnd,
			Global: &event.Class{
				TotalQueries:  1,
				UniqueQueries: 1,
				Metrics: &event.Metrics{
					TimeMetrics: map[string]*event.TimeStats{
						"Query_time": {
							Sum: 1,
							Med: event.Float64(1),
							P95: event.Float64(1),
							Min: event.Float64(1),
							Avg: event.Float64(1),
							Max: event.Float64(1),
						},
					},
					NumberMetrics: map[string]*event.NumberStats{
						"Bytes_sent": {
							Sum: 0,
							Min: event.Uint64(0),
							Avg: event.Uint64(0),
							Med: event.Uint64(0),
							P95: event.Uint64(0),
							Max: event.Uint64(0),
						},
						"Rows_examined": {
							Sum: 13,
							Med: event.Uint64(13),
							P95: event.Uint64(13),
							Min: event.Uint64(13),
							Avg: event.Uint64(13),
							Max: event.Uint64(13),
						},
						"Rows_sent": {
							Sum: 42,
							Med: event.Uint64(42),
							P95: event.Uint64(42),
							Min: event.Uint64(42),
							Avg: event.Uint64(42),
							Max: event.Uint64(42),
						},
					},
					BoolMetrics: map[string]*event.BoolStats{},
				},
			},
			Class: []*event.Class{
				{
					Id:            "d41d8cd98f00b204e9800998ecf8427e",
					TotalQueries:  1,
					UniqueQueries: 1,
					Metrics: &event.Metrics{
						TimeMetrics: map[string]*event.TimeStats{
							"Query_time": {
								Sum: 1,
								Min: event.Float64(1),
								Avg: event.Float64(1),
								Med: event.Float64(1),
								P95: event.Float64(1),
								Max: event.Float64(1),
							},
						},
						NumberMetrics: map[string]*event.NumberStats{
							"Bytes_sent": {
								Sum: 0,
								Min: event.Uint64(0),
								Avg: event.Uint64(0),
								Med: event.Uint64(0),
								P95: event.Uint64(0),
								Max: event.Uint64(0),
							},
							"Rows_examined": {
								Sum: 13,
								Min: event.Uint64(13),
								Avg: event.Uint64(13),
								Med: event.Uint64(13),
								P95: event.Uint64(13),
								Max: event.Uint64(13),
							},
							"Rows_sent": {
								Sum: 42,
								Min: event.Uint64(42),
								Avg: event.Uint64(42),
								Med: event.Uint64(42),
								P95: event.Uint64(42),
								Max: event.Uint64(42),
							},
						},
						BoolMetrics: map[string]*event.BoolStats{},
					},
					Example: &event.Example{},
				},
			},
		}
		err := aggregator.Add(doc)
		require.NoError(t, err)
		report, ok := <-reportChan
		assert.True(t, ok)
		assert.Equal(t, expected, *report)
	}
}

// TestAggregator_Add_EmptyInterval verifies that no report is returned if there were no samples in interval #PMM-927
func TestAggregator_Add_EmptyInterval(t *testing.T) {
	t.Parallel()

	timeStart, err := time.Parse("2006-01-02 15:04:05", "2017-07-02 07:55:00")
	require.NoError(t, err)
	timeEnd, err := time.Parse("2006-01-02 15:04:05", "2017-07-02 07:56:00")
	require.NoError(t, err)

	config := pc.QAN{
		UUID:     "abc",
		Interval: 60, // 60s
	}

	aggregator := New(timeStart, config)
	reportChan := aggregator.Start()

	// finish interval immediately
	{
		doc := proto.SystemProfile{
			Ts: timeEnd,
		}
		err := aggregator.Add(doc)
		require.NoError(t, err)
		aggregator.Stop()
		report, ok := <-reportChan
		assert.False(t, ok)

		// no report should be returned
		assert.Nil(t, report)
	}
}

func TestAggregator_StartStop(t *testing.T) {
	var err error
	config := pc.QAN{
		UUID:     "abc",
		Interval: 60, // 60s
	}

	timeStart, err := time.Parse("2006-01-02 15:04:05", "2017-07-02 07:55:00")
	aggregator := New(timeStart, config)
	reportChan1 := aggregator.Start()
	require.NoError(t, err)

	// running multiple Start() should be idempotent
	reportChan2 := aggregator.Start()
	require.NoError(t, err)

	assert.Exactly(t, reportChan1, reportChan2)

	// running multiple Stop() should be idempotent
	aggregator.Stop()
	aggregator.Stop()
}
