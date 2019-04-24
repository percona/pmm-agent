package aggregator

import (
	"expvar"
)

type stats struct {
	DocsIn         *expvar.Int    `name:"docs-in"`
	DocsSkippedOld *expvar.Int    `name:"docs-skipped-old"`
	ReportsOut     *expvar.Int    `name:"reports-out"`
	IntervalStart  *expvar.String `name:"interval-start"`
	IntervalEnd    *expvar.String `name:"interval-end"`
}
