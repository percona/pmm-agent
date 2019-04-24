package sender

import (
	"expvar"
)

type stats struct {
	In      *expvar.Int `name:"in"`
	Out     *expvar.Int `name:"out"`
	ErrIter *expvar.Int `name:"err-iter"`
}
