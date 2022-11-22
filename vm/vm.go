package vm

import (
	"fmt"

	"protiumx.dev/simia/code"
	"protiumx.dev/simia/compiler"
	"protiumx.dev/simia/value"
)

const StackSize = (1 << 10) * 2

type VM struct {
	constants    []value.Value
	instructions code.Instructions
	stack        []value.Value
	sp           int // Stack pointer points to next free slot in stack
}

func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,
		stack:        make([]value.Value, StackSize),
		sp:           0,
	}
}

func (vm *VM) StackTop() value.Value {
	if vm.sp == 0 {
		return nil
	}

	return vm.stack[vm.sp-1]
}

func (vm *VM) Run() error {
	for ip := 0; ip < len(vm.instructions); ip++ {
		op := code.Opcode(vm.instructions[ip])

		switch op {
		case code.OpConstant:
			idx := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2
			err := vm.push(vm.constants[idx])
			if err != nil {
				return err
			}

		case code.OpAdd:
			right := vm.pop()
			left := vm.pop()
			rValue := right.(*value.Integer).Value
			lValue := left.(*value.Integer).Value
			vm.push(&value.Integer{Value: lValue + rValue})
		}
	}

	return nil
}

func (vm *VM) push(v value.Value) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.sp] = v
	vm.sp++

	return nil
}

func (vm *VM) pop() value.Value {
	v := vm.stack[vm.sp-1]
	vm.sp--
	return v
}
