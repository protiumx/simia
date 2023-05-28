package vm

import (
	"protiumx.dev/simia/code"
	"protiumx.dev/simia/value"
)

type Frame struct {
	cl          *value.Closure
	ip          int
	basePointer int // also called frame pointer
}

func NewFrame(cl *value.Closure, basePointer int) *Frame {
	return &Frame{cl, -1, basePointer}
}

func (f *Frame) Instructions() code.Instructions {
	return f.cl.Fn.Instructions
}
