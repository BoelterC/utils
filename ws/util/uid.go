package util

import "sync/atomic"

type UID64 struct {
	counter int64
}

func (c *UID64) Get64() int64 {
	for {
		val := atomic.LoadInt64(&c.counter)
		if atomic.CompareAndSwapInt64(&c.counter, val, val+1) {
			return val
		}
	}
}

type UID32 struct {
	counter uint32
}

func (c *UID32) Get32() uint32 {
	for {
		val := atomic.LoadUint32(&c.counter)
		if atomic.CompareAndSwapUint32(&c.counter, val, val+1) {
			return val
		}
	}
}
