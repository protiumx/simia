package vm

import (
	"protiumx.dev/simia/code"
	"protiumx.dev/simia/value"
)

type Frame struct {
  fn *value.CompiledFunction
  ip int
}

func NewFrame(fn *value.CompiledFunction) *Frame {
  return &Frame{fn, -1}
}

func (f *Frame) Instructions() code.Instructions {
  return f.fn.Instructions
}
