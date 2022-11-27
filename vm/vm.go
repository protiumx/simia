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

		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			err := vm.execBinaryOp(op)
			if err != nil {
				return err
			}
		case code.OpPop:
			vm.pop()
		}
	}

	return nil
}

func (vm *VM) execBinaryOp(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	if left.Type() == value.INTEGER_VALUE && right.Type() == value.INTEGER_VALUE {
		return vm.execBinaryIntegerOp(op, left, right)
	}

	return fmt.Errorf("unsupported types for binary operation: %s %d %s", left.Type(), op, right.Type())
}

func (vm *VM) execBinaryIntegerOp(op code.Opcode, left, right value.Value) error {
	lValue := left.(*value.Integer).Value
	rValue := right.(*value.Integer).Value
	var result int64

	switch op {
	case code.OpAdd:
		result = lValue + rValue
	case code.OpSub:
		result = lValue - rValue
	case code.OpMul:
		result = lValue * rValue
	case code.OpDiv:
		result = lValue / rValue
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}

	return vm.push(&value.Integer{Value: result})
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

// LastPoppedStackElement uses the stack pointer to retrieve the last element that was popped
func (vm *VM) LastPoppedStackElement() value.Value {
	return vm.stack[vm.sp]
}
