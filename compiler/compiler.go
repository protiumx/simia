package compiler

import (
	"fmt"
	"sort"

	"protiumx.dev/simia/ast"
	"protiumx.dev/simia/code"
	"protiumx.dev/simia/value"
)

type Compiler struct {
	constants   []value.Value
	symbolTable *SymbolTable
	scopes      []Scope
	scopeIndex  int
}

type Bytecode struct {
	Instructions code.Instructions
	Constants    []value.Value
}

type EmittedInstruction struct {
	Opcode   code.Opcode
	Position int
}

type Scope struct {
	instructions    code.Instructions
	lastInstruction EmittedInstruction
	prevInstruction EmittedInstruction
}

func New() *Compiler {
	scope := Scope{
		instructions:    code.Instructions{},
		prevInstruction: EmittedInstruction{},
		lastInstruction: EmittedInstruction{},
	}
	return &Compiler{
		constants:   []value.Value{},
		symbolTable: NewSymbolTable(),
		scopes:      []Scope{scope},
		scopeIndex:  0,
	}
}

func NewWithState(s *SymbolTable, constants []value.Value) *Compiler {
	scope := Scope{
		instructions:    code.Instructions{},
		prevInstruction: EmittedInstruction{},
		lastInstruction: EmittedInstruction{},
	}
	return &Compiler{
		constants:   constants,
		symbolTable: s,
		scopes:      []Scope{scope},
		scopeIndex:  0,
	}
}

func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}

	case *ast.ExpressionStatement:
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}

		c.emit(code.OpPop)

	case *ast.PrefixExpression:
		err := c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "!":
			c.emit(code.OpBang)
		case "-":
			c.emit(code.OpMinus)
		default:
			return fmt.Errorf("unkown operator %s", node.Operator)
		}

	case *ast.InfixExpression:
		if node.Operator == "<" {
			// invert operands
			node.Left, node.Right = node.Right, node.Left
		}

		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "+":
			c.emit(code.OpAdd)
		case "-":
			c.emit(code.OpSub)
		case "*":
			c.emit(code.OpMul)
		case "/":
			c.emit(code.OpDiv)
		case ">", "<":
			c.emit(code.OpGreaterThan)
		case "==":
			c.emit(code.OpEqual)
		case "!=":
			c.emit(code.OpNotEqual)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}

	case *ast.IntegerLiteral:
		integer := &value.Integer{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(integer))

	case *ast.StringLiteral:
		str := &value.String{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(str))

	case *ast.Boolean:
		if node.Value {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}

	case *ast.ArrayLiteral:
		for _, el := range node.Elements {
			err := c.Compile(el)
			if err != nil {
				return err
			}
		}

		c.emit(code.OpArray, len(node.Elements))

	case *ast.HashLiteral:
		keys := []ast.Expression{}
		for k := range node.Pairs {
			keys = append(keys, k)
		}

		// Sorting guarantees the order of constants in the tests
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].String() < keys[j].String()
		})

		for _, k := range keys {
			err := c.Compile(k)
			if err != nil {
				return err
			}

			err = c.Compile(node.Pairs[k])
			if err != nil {
				return err
			}
		}

		c.emit(code.OpHash, len(node.Pairs)*2)

	case *ast.IfExpression:
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}

		jumpIfBrachPos := c.emit(code.OpJumpIfBranch, -1)
		err = c.Compile(node.Consequence)
		if err != nil {
			return err
		}

		if c.lastInstructionIs(code.OpPop) {
			c.removeLastPop()
		}

		jumpPos := c.emit(code.OpJump, -1)

		c.changeOperandAt(jumpIfBrachPos, len(c.currentInstructions()))

		if node.Alternative == nil {
			// as if are expressions, the alternative must be a OpNil
			c.emit(code.OpNil)
		} else {
			err := c.Compile(node.Alternative)
			if err != nil {
				return err
			}

			if c.lastInstructionIs(code.OpPop) {
				c.removeLastPop()
			}

		}
		c.changeOperandAt(jumpPos, len(c.currentInstructions()))

	case *ast.BlockStatment:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}

	case *ast.LetStatement:
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}

		symbol := c.symbolTable.Define(node.Name.Value)
		c.emit(code.OpSetGlobal, symbol.Index)

	case *ast.Identifier:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("undefined variable %s", node.Value)
		}

		c.emit(code.OpGetGlobal, symbol.Index)
	case *ast.IndexExpression:
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Index)
		if err != nil {
			return err
		}

		c.emit(code.OpIndex)

	case *ast.FunctionLiteral:
		c.enterScope()
		err := c.Compile(node.Body)
		if err != nil {
			return err
		}

		if c.lastInstructionIs(code.OpPop) {
			c.replaceLastPopWithReturn()
		}

		if !c.lastInstructionIs(code.OpReturnValue) {
			c.emit(code.OpReturn)
		}

		instructions := c.leaveScope()
		compiledFn := &value.CompiledFunction{Instructions: instructions}
		c.emit(code.OpConstant, c.addConstant(compiledFn))

	case *ast.ReturnStatement:
		err := c.Compile(node.ReturnValue)
		if err != nil {
			return err
		}

		c.emit(code.OpReturnValue)

  case *ast.CallExpression:
    err := c.Compile(node.Function)
    if err != nil {
      return err 
    }

    c.emit(code.OpCall)
	}

	return nil
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.currentInstructions(),
		Constants:    c.constants,
	}
}

func (c *Compiler) currentInstructions() code.Instructions {
	return c.scopes[c.scopeIndex].instructions
}

func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.addInstruction(ins)

	c.setLastInstruction(op, pos)
	return pos
}

func (c *Compiler) setLastInstruction(op code.Opcode, pos int) {
	scope := c.scopes[c.scopeIndex]
	c.scopes[c.scopeIndex].prevInstruction = scope.lastInstruction
	c.scopes[c.scopeIndex].lastInstruction = EmittedInstruction{Opcode: op, Position: pos}
}

func (c *Compiler) lastInstructionIs(op code.Opcode) bool {
	if len(c.currentInstructions()) == 0 {
		return false
	}

	return c.scopes[c.scopeIndex].lastInstruction.Opcode == op
}

func (c *Compiler) replaceInstructionAt(pos int, newInstruction []byte) {
	currentInstructions := c.currentInstructions()
	for i := 0; i < len(newInstruction); i++ {
		currentInstructions[pos+i] = newInstruction[i]
	}
}

func (c *Compiler) removeLastPop() {
	scope := c.scopes[c.scopeIndex]
	c.scopes[c.scopeIndex].instructions = scope.instructions[:scope.lastInstruction.Position]
	c.scopes[c.scopeIndex].lastInstruction = scope.prevInstruction
}

func (c *Compiler) replaceLastPopWithReturn() {
	lastPos := c.scopes[c.scopeIndex].lastInstruction.Position
	c.replaceInstructionAt(lastPos, code.Make(code.OpReturnValue))

	c.scopes[c.scopeIndex].lastInstruction.Opcode = code.OpReturnValue
}

func (c *Compiler) changeOperandAt(opPos int, operand int) {
	op := code.Opcode(c.currentInstructions()[opPos])
	newInstruction := code.Make(op, operand)
	c.replaceInstructionAt(opPos, newInstruction)
}

func (c *Compiler) addInstruction(ins []byte) int {
	newPos := len(c.currentInstructions())
	c.scopes[c.scopeIndex].instructions = append(c.currentInstructions(), ins...)
	return newPos
}

func (c *Compiler) addConstant(v value.Value) int {
	c.constants = append(c.constants, v)
	return len(c.constants) - 1
}

func (c *Compiler) enterScope() {
	scope := Scope{
		instructions:    code.Instructions{},
		lastInstruction: EmittedInstruction{},
		prevInstruction: EmittedInstruction{},
	}
	c.scopes = append(c.scopes, scope)
	c.scopeIndex++
}

func (c *Compiler) leaveScope() code.Instructions {
	instructions := c.currentInstructions()
	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeIndex--
	return instructions
}
