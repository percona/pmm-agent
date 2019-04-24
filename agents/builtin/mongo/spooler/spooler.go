package spooler

import (
	"github.com/percona/pmm-agent/agents/builtin/mongo/profiler/sender"
	"github.com/percona/pmm-agent/agents/builtin/mongo/proto/qan"
)

type SimpleSpooler struct{}

func New() sender.Spooler {
	return &SimpleSpooler{}
}

// Maps QAN Report to MetricsBuckets
// TODO: Move Spooler implementation to other place
func (s *SimpleSpooler) Write(r *qan.Report) error {
	panic("not implemented")
}
