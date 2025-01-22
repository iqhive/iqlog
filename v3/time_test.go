package iqlog

import (
	"io"
	"testing"
)

func BenchmarkBufferLineTime(b *testing.B) {
	logger := NewIQLogger(true)
	logger.SetDebugMode(false)
	logger.SetWriter(io.Discard)

	bl := logger.WithBufferLineInfo()

	for i := 0; i < b.N; i++ {
		bl.buffer.Reset()
		bl.AddTime()
	}
}

func BenchmarkBufferLineNLTime(b *testing.B) {
	logger := NewIQLogger(true)
	logger.SetDebugMode(false)
	logger.SetWriter(io.Discard)

	bl := logger.WithBufferLineNLInfo()

	for i := 0; i < b.N; i++ {
		bl.buffer.Reset()
		bl.AddTime()
	}
}

// func BenchmarkBufferLineTime2(b *testing.B) {
// 	logger := NewIQLogger(true)
// 	logger.SetDebugMode(false)
// 	logger.SetWriter(io.Discard)

// 	bl := logger.WithBufferLineInfo()

// 	for i := 0; i < b.N; i++ {
// 		bl.buffer.Reset()
// 		bl.AddTime2()
// 	}
// }

func BenchmarkByteSliceTime(b *testing.B) {
	logger := NewIQLogger(true)
	logger.SetDebugMode(false)
	logger.SetWriter(io.Discard)

	bl := logger.WithByteSliceLineInfo()
	emptySlice := make([]byte, 0)

	for i := 0; i < b.N; i++ {
		bl.output = emptySlice
		bl.AddTime()
	}
}

func BenchmarkPreallocTime(b *testing.B) {
	logger := NewIQLogger(true)
	logger.SetDebugMode(false)
	logger.SetWriter(io.Discard)

	bl := logger.WithPreallocLineInfo()

	for i := 0; i < b.N; i++ {
		bl.bytesUsed = 0
		bl.AddTime()
	}
}
