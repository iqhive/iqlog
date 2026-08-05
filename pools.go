package iqlog

import (
	"sync"
)

const (
	initialEventCapacity = 512
	maxPooledCapacity    = 64 << 10
)

var eventPool = sync.Pool{
	New: func() any { return new(Event) },
}

var eventBufferPool = sync.Pool{
	New: func() any {
		buf := make([]byte, 0, initialEventCapacity)
		return &buf
	},
}

func acquireEventBuffer() []byte {
	p := eventBufferPool.Get().(*[]byte)
	buf := (*p)[:0]
	*p = nil
	return buf
}

func releaseEventBuffer(buf []byte) {
	if cap(buf) > maxPooledCapacity {
		return
	}
	buf = buf[:0]
	eventBufferPool.Put(&buf)
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
