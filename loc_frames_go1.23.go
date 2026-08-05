//go:build go1.23

package iqlog

import "runtime"

type (
	runtimeFrame = runtime.Frame

	runtimeFrames struct {
		ptr *PC
		len int
		buf PC // cap

		nextPC PC

		frames     []runtimeFrame
		frameStore [2]runtimeFrame
	}
)
