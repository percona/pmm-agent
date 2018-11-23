package logger

import (
	"container/ring"
	"sync"
)

type CircularWriter struct {
	r  *ring.Ring
	rw sync.Mutex
}

func New(len int) *CircularWriter {
	return &CircularWriter{
		r: ring.New(len),
	}
}

func (c *CircularWriter) Write(p []byte) (n int, err error) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.r.Value = p
	c.r = c.r.Next()
	return len(p), nil
}

func (c *CircularWriter) String() string {
	result := ""
	c.r.Do(func(i interface{}) {
		if i != nil {
			result += string(i.([]byte))
		}
	})
	return result
}
