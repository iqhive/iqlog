package iqlog

import (
	"testing"
)

func BenchmarkBytesAppender(b *testing.B) {
	ba := baPool.Get().(*bytesAppender)
	for i := 0; i < b.N; i++ {
		ba.Write([]byte("hello"))
	}
}
