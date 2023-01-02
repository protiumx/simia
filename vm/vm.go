package vm

import (
	"fmt"

	"protiumx.dev/simia/code"
	"protiumx.dev/simia/compiler"
	"protiumx.dev/simia/value"
)

const (
	StackSize   = (1 << 10) * 2
	GlobalsSize = (2 << 15)
)

var (
	True  = &value.Boolean{Value: true}
	False = &value.Boolean{Value: false}
	Nil   = &value.Nil{}
)

type VM struct {
	constants    []value.Value
	globals      []value.Value
	stack        []value.Value
	instructions code.Instructions
	sp           int // Stack pointer points to next free slot in stack
}

func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,
		globals:      make([]value.Value, GlobalsSize),
		stack:        make([]value.Value, StackSize),
		sp:           0,
	}
}

func NewWithGlobalStore(bytecode *compiler.Bytecode, s []value.Value) *VM {
	vm := New(bytecode)
	vm.globals = s
	return vm
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

		case code.OpTrue:
			err := vm.push(True)
			if err != nil {
				return err
			}

		case code.OpFalse:
			err := vm.push(False)
			if err != nil {
				return err
			}

		case code.OpNil:
			err := vm.push(Nil)
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

		case code.OpJump:
			pos := int(code.ReadUint16(vm.instructions[ip+1:]))
			// account for addition in for loop
			ip = pos - 1

		case code.OpJumpIfBranch:
			pos := int(code.ReadUint16(vm.instructions[ip+1:]))
			// assume condition is truthy
			ip += 2
			condition := vm.pop()
			if !isTruthy(condition) {
				// jump to else branch
				ip = pos - 1
			}

		case code.OpEqual, code.OpNotEqual, code.OpGreaterThan:
			err := vm.execComparison(op)
			if err != nil {
				return err
			}

		case code.OpBang:
			err := vm.execBangOperator()
			if err != nil {
				return err
			}

		case code.OpMinus:
			err := vm.execMinusOperator()
			if err != nil {
				return err
			}

		case code.OpSetGlobal:
			gIndex := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2

			vm.globals[gIndex] = vm.pop()

		case code.OpGetGlobal:
			gIndex := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2

			err := vm.push(vm.globals[gIndex])
			if err != nil {
				return err
			}
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

func (vm *VM) execComparison(op code.Opcode) error {
	right, left := vm.pop(), vm.pop()
	if left.Type() == value.INTEGER_VALUE && right.Type() == value.INTEGER_VALUE {
		return vm.execIntegerComparison(op, left, right)
	}

	switch op {
	case code.OpEqual:
		return vm.push(getBoolean(right == left))
	case code.OpNotEqual:
		return vm.push(getBoolean(right != left))
	default:
		return fmt.Errorf("unknown operator: %d (%s %s)", op, left.Type(), right.Type())
	}
}

func (vm *VM) execIntegerComparison(op code.Opcode, left, right value.Value) error {
	leftVal, rightVal := left.(*value.Integer).Value, right.(*value.Integer).Value

	switch op {
	case code.OpEqual:
		return vm.push(getBoolean(rightVal == leftVal))
	case code.OpNotEqual:
		return vm.push(getBoolean(rightVal != leftVal))
	case code.OpGreaterThan:
		return vm.push(getBoolean(rightVal < leftVal))
	default:
		return fmt.Errorf("unknown operator: %d", op)
	}
}

func (vm *VM) execBangOperator() error {
	operand := vm.pop()
	switch operand {
	case True:
		return vm.push(False)
	case False:
		return vm.push(True)
	default:
		return fmt.Errorf("unkown operand %s for Bang operator", operand.Type())
	}
}

func (vm *VM) execMinusOperator() error {
	operand := vm.pop()
	if operand.Type() != value.INTEGER_VALUE {
		return fmt.Errorf("unsupported type for negation: %s", operand.Type())
	}
	v := operand.(*value.Integer).Value
	return vm.push(&value.Integer{Value: -v})
}

func getBoolean(v bool) *value.Boolean {
	if v {
		return True
	}
	return False
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

func isTruthy(val value.Value) bool {
	switch val := val.(type) {
	case *value.Boolean:
		return val.Value
	case *value.Integer:
		return val.Value != 0
	case *value.Nil:
		return false
	default:
		return false
	}
}

// LastPoppedStackElement uses the stack pointer to retrieve the last element that was popped
func (vm *VM) LastPoppedStackElement() value.Value {
	return vm.stack[vm.sp]
}
