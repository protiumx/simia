package vm

import (
	"protiumx.dev/simia/code"
	"protiumx.dev/simia/value"
)

type Frame struct {
	fn          *value.CompiledFunction
	ip          int
	basePointer int // also called frame pointer
}

func NewFrame(fn *value.CompiledFunction, basePointer int) *Frame {
	return &Frame{fn, -1, basePointer}
}

func (f *Frame) Instructions() code.Instructions {
	return f.fn.Instructions
}
