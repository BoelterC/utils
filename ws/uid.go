package ws

import "sync/atomic"

type UniqueID struct {
	counter int64
}

func (c *UniqueID) get() int64 {
	for {
		val := atomic.LoadInt64(&c.counter)
		if atomic.CompareAndSwapInt64(&c.counter, val, val+1) {
			return val
		}
	}
}
