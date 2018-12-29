package supervisor

import (
	"math"
	"math/rand"
	"sync/atomic"
	"time"
)

type restartCounter struct {
	count int32
}

func (r *restartCounter) Inc() {
	atomic.AddInt32(&r.count, 1)
}

func (r *restartCounter) Reset() {
	atomic.CompareAndSwapInt32(&r.count, r.count, 1)
}

func (r *restartCounter) Delay() time.Duration {
	max := math.Pow(2, float64(r.count)) - 1
	delay := rand.Int63n(int64(max))
	return (1 + time.Duration(delay)) * time.Millisecond
}
