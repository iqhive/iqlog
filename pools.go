package iqlog

import (
	"sync"
)

const (
	initialEventCapacity = 512
	maxPooledCapacity    = 64 << 10
)

var eventPool = sync.Pool{
	New: func() any { return &Event{timeSecond: invalidTimestampSecond} },
}

var eventBufferPool = sync.Pool{
	New: func() any {
		buf := make([]byte, 0, initialEventCapacity)
		return &buf
	},
}

// bufferHeaderPool recycles the *[]byte handles eventBufferPool stores.
// sync.Pool.Put takes an interface, so handing it the address of a local
// slice allocates a fresh slice header every time. That cost is invisible to
// synchronous logging, which keeps its buffer attached to the pooled Event,
// but the async writer releases every record's buffer from the background
// goroutine, so it paid a heap allocation per record -- exactly what the
// ownership transfer in WriteOwned exists to avoid.
var bufferHeaderPool = sync.Pool{New: func() any { return new([]byte) }}

func acquireEventBuffer() []byte {
	p := eventBufferPool.Get().(*[]byte)
	buf := (*p)[:0]
	// clear before recycling so the handle does not keep the array alive
	*p = nil
	bufferHeaderPool.Put(p)
	return buf
}

func releaseEventBuffer(buf []byte) {
	if cap(buf) > maxPooledCapacity {
		return
	}
	p := bufferHeaderPool.Get().(*[]byte)
	*p = buf[:0]
	eventBufferPool.Put(p)
}

type bytesAppender struct {
	Bytes []byte
}

func (b *bytesAppender) Write(appendMe []byte) (int, error) {
	b.Bytes = append(b.Bytes, appendMe...)
	return len(appendMe), nil
}

var baPool = sync.Pool{
	New: func() any {
		ba := bytesAppender{
			Bytes: make([]byte, 0, maxLineLen),
		}
		return &ba
	},
}

func acquireBytesAppender() *bytesAppender {
	ba := baPool.Get().(*bytesAppender)
	ba.Bytes = ba.Bytes[:0]
	return ba
}

// releaseBytesAppender returns ba to the pool unless a single oversized
// format grew it past the retention ceiling, which would otherwise pin that
// capacity for the lifetime of the process.
func releaseBytesAppender(ba *bytesAppender) {
	if cap(ba.Bytes) > maxPooledCapacity {
		return
	}
	ba.Bytes = ba.Bytes[:0]
	baPool.Put(ba)
}
