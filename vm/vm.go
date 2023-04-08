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
	MaxFrames   = (1 << 10)
)

var (
	True  = &value.Boolean{Value: true}
	False = &value.Boolean{Value: false}
	Nil   = &value.Nil{}
)

type VM struct {
	constants   []value.Value
	globals     []value.Value
	stack       []value.Value
	frames      []*Frame
	framesIndex int
	sp          int // Stack pointer points to next free slot in stack
}

func New(bytecode *compiler.Bytecode) *VM {
	mainFn := &value.CompiledFunction{Instructions: bytecode.Instructions}
	mainFrame := NewFrame(mainFn)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame
	return &VM{
		constants:   bytecode.Constants,
		globals:     make([]value.Value, GlobalsSize),
		stack:       make([]value.Value, StackSize),
		frames:      frames,
		framesIndex: 1,
		sp:          0,
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
	var ip int
	var ins code.Instructions
	var op code.Opcode

	for vm.currentFrame().ip < len(vm.currentFrame().Instructions())-1 {
		currentFrame := vm.currentFrame()
		currentFrame.ip++
		ip = currentFrame.ip
		ins = currentFrame.Instructions()
		op = code.Opcode(ins[ip])

		switch op {
		case code.OpConstant:
			idx := code.ReadUint16(ins[ip+1:])
			currentFrame.ip += 2
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
			pos := int(code.ReadUint16(ins[ip+1:]))
			// account for addition in for loop
			currentFrame.ip = pos - 1

		case code.OpJumpIfBranch:
			pos := int(code.ReadUint16(ins[ip+1:]))
			// assume condition is truthy
			currentFrame.ip += 2
			condition := vm.pop()
			if !isTruthy(condition) {
				// jump to else branch
				currentFrame.ip = pos - 1
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
			gIndex := code.ReadUint16(ins[ip+1:])
			currentFrame.ip += 2

			vm.globals[gIndex] = vm.pop()

		case code.OpGetGlobal:
			gIndex := code.ReadUint16(ins[ip+1:])
			currentFrame.ip += 2

			err := vm.push(vm.globals[gIndex])
			if err != nil {
				return err
			}

		case code.OpArray:
			numElements := int(code.ReadUint16(ins[ip+1:]))
			currentFrame.ip += 2

			arr := vm.buildArray(vm.sp-numElements, vm.sp)
			vm.sp = vm.sp - numElements

			err := vm.push(arr)
			if err != nil {
				return err
			}

		case code.OpHash:
			numElements := int(code.ReadUint16(ins[ip+1:]))
			currentFrame.ip += 2

			hash, err := vm.buildHash(vm.sp-numElements, vm.sp)
			if err != nil {
				return err
			}

			vm.sp = vm.sp - numElements
			err = vm.push(hash)
			if err != nil {
				return err
			}
		case code.OpIndex:
			index := vm.pop()
			left := vm.pop()
			err := vm.execIndexExpression(left, index)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.framesIndex-1]
}

func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.framesIndex] = f
	vm.framesIndex++
}

func (vm *VM) popFrame(f *Frame) *Frame {
	vm.framesIndex--
	return vm.frames[vm.framesIndex]
}

func (vm *VM) buildHash(startIndex, endIndex int) (value.Value, error) {
	hash := &value.Hash{}
	hash.Pairs = make(map[string]value.Value)

	for i := startIndex; i < endIndex; i += 2 {
		k := vm.stack[i]
		v := vm.stack[i+1]

		key, ok := k.(*value.String)
		if !ok {
			return nil, fmt.Errorf("unusable hash key: %s", k.Type())
		}

		hash.Pairs[key.Value] = v
	}

	return hash, nil
}

func (vm *VM) buildArray(startIndex, endIndex int) value.Value {
	elements := make([]value.Value, endIndex-startIndex)

	for i := startIndex; i < endIndex; i++ {
		elements[i-startIndex] = vm.stack[i]

	}

	return &value.Array{Elements: elements}
}

func (vm *VM) execBinaryOp(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	switch left.Type() {
	case value.INTEGER_VALUE:
		if right.Type() == value.INTEGER_VALUE {
			return vm.execBinaryIntegerOp(op, left, right)
		}

	case value.STRING_VALUE:
		if right.Type() == value.STRING_VALUE {
			return vm.execBinaryStringOp(op, left, right)
		}
	}

	return fmt.Errorf("unsupported types for binary operation: %s %d %s", left.Type(), op, right.Type())
}

func (vm *VM) execBinaryStringOp(op code.Opcode, left, right value.Value) error {
	if op != code.OpAdd {
		return fmt.Errorf("unknown string operation: %d", op)
	}

	l, r := left.(*value.String).Value, right.(*value.String).Value
	return vm.push(&value.String{Value: l + r})
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

func (vm *VM) execIndexExpression(left, index value.Value) error {
	switch {
	case left.Type() == value.ARRAY_VALUE && index.Type() == value.INTEGER_VALUE:
		return vm.execArrayIndex(left, index)
	case left.Type() == value.HASH_VALUE && index.Type() == value.STRING_VALUE:
		return vm.execHashIndex(left, index)
	default:
		return fmt.Errorf("index operator not supported: %s", left.Type())
	}
}

func (vm *VM) execArrayIndex(array, index value.Value) error {
	arr := array.(*value.Array)
	i := index.(*value.Integer).Value
	max := int64(len(arr.Elements) - 1)
	if i < 0 || i > max {
		return vm.push(Nil)
	}

	return vm.push(arr.Elements[i])
}

func (vm *VM) execHashIndex(hashValue, index value.Value) error {
	hash := hashValue.(*value.Hash)
	key := index.(*value.String)
	pair, ok := hash.Pairs[key.Value]
	if !ok {
		return vm.push(Nil)
	}

	return vm.push(pair)
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
