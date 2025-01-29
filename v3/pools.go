package iqlog

import (
	"bytes"
	"errors"
	"sync"
)

// Global fixed slice pool of pointers
var fixedSlicePool = sync.Pool{
	New: func() any {
		slice := [maxLineLen]byte{}
		return slice
	},
}

var preallocLinePool = sync.Pool{
	New: func() any {
		resp := preallocLine{}
		slice := [maxLineLen]byte{}
		resp.output = &slice
		return &resp
	},
}

var preallocLine2Pool = sync.Pool{
	New: func() any {
		resp := preallocLine2{}
		resp.output = make([]byte, maxLineLen, maxLineLen)
		resp.output[maxLineLen-1] = '\255'
		return &resp
	},
}

var bytesliceLinePool = sync.Pool{
	New: func() any {
		resp := bytesliceLine{}
		return &resp
	},
}

var bufferLinePool = sync.Pool{
	New: func() any {
		resp := bufferLine{
			buffer: bytes.NewBuffer(make([]byte, 0, maxLineLen)),
		}
		return &resp
	},
}

var bufferLineNLPool = sync.Pool{
	New: func() any {
		resp := bufferLineNL{}
		resp.buffer = bytes.NewBuffer(make([]byte, 0, maxLineLen))
		return &resp
	},
}

// lets test some different bytes appenders

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

type byteIndexAppender struct {
	Bytes [maxLineLen]byte
	Index int
}

func (b *byteIndexAppender) Write(appendMe []byte) (int, error) {
	len := len(appendMe)
	if b.Index+len > maxLineLen {
		return 0, errors.New("byteIndexAppender: we are full")
	}
	copy(b.Bytes[b.Index:], appendMe)
	b.Index += len
	return len, nil
}

var biapool = sync.Pool{
	New: func() any {
		bi := byteIndexAppender{
			Bytes: [maxLineLen]byte{},
			Index: 0,
		}
		return &bi
	},
}
