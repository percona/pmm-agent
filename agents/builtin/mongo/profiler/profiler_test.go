package profiler

import (
	"testing"
	"time"

	"github.com/percona/pmgo"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/percona/pmm-agent/agents/builtin/mongo/proto/config"
	"github.com/percona/pmm-agent/agents/builtin/mongo/proto/qan"
	"github.com/percona/pmm-agent/agents/builtin/mongo/test/mock"
	"github.com/percona/pmm-agent/agents/builtin/mongo/test/profiling"
)

func TestCollectingAndSendingData(t *testing.T) {
	// Disable profiling.
	err := profiling.New("").DisableAll()
	require.NoError(t, err)
	// Enable profiling for default db.
	err = profiling.New("").Enable("")
	require.NoError(t, err)

	// Create dependencies.
	dialer := pmgo.NewDialer()
	dialInfo, _ := pmgo.ParseURL("")
	logger := logrus.WithField("component", "profiler-test")
	dataChan := make(chan *qan.Report)
	spool := mock.NewSpooler(dataChan)
	// Create the QAN config.
	exampleQueries := true
	qanConfig := config.QAN{
		UUID:           "12345678",
		Interval:       5, // seconds
		ExampleQueries: &exampleQueries,
	}
	plugin := New(dialInfo, dialer, logger, spool, qanConfig)

	assert.Empty(t, plugin.Status())
	err = plugin.Start()
	require.NoError(t, err)
	assert.Equal(t, "Profiling enabled for all queries (ratelimit: 1)", plugin.Status()["collector-profile-test"])

	// Add some data to mongo e.g. people.
	people := []map[string]string{
		{"name": "Kamil"},
		{"name": "Carlos"},
	}
	// Add data through separate connection.
	session, err := dialer.DialWithInfo(dialInfo)
	require.NoError(t, err)
	for _, person := range people {
		err = session.DB("").C("people").Insert(&person)
		require.NoError(t, err)
	}

	// Wait until we receive data
	select {
	case data := <-dataChan:
		qanReport := data
		assert.EqualValues(t, 2, qanReport.Global.TotalQueries)
		assert.EqualValues(t, 1, qanReport.Global.UniqueQueries)
	case <-time.After(2 * time.Duration(qanConfig.Interval) * time.Second):
		t.Fatal("timeout waiting for data")
	}

	status := plugin.Status()
	assert.Equal(t, "Profiling enabled for all queries (ratelimit: 1)", status["collector-profile-test"])
	assert.Equal(t, "2", status["collector-in-test"])
	assert.Equal(t, "2", status["collector-out-test"])
	assert.Equal(t, "2", status["parser-docs-in-test"])
	assert.Equal(t, "2", status["aggregator-docs-in"])
	assert.Equal(t, "1", status["aggregator-reports-out"])
	assert.Equal(t, "1", status["sender-in"])
	assert.Equal(t, "1", status["sender-out"])

	err = plugin.Stop()
	require.NoError(t, err)
	assert.Empty(t, plugin.Status())
}
