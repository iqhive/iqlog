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

func BenchmarkByteIndexAppender(b *testing.B) {
	bi := biapool.Get().(*byteIndexAppender)
	for i := 0; i < b.N; i++ {
		bi.Write([]byte("hello"))
	}
}
