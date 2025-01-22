package iqlog

import (
	"bytes"
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
		resp.output = make([]byte, maxLineLen)
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
