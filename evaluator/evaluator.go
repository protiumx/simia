package evaluator

import (
	"fmt"

	"protiumx.dev/simia/ast"
	"protiumx.dev/simia/value"
)

var (
	NIL   = &value.Nil{}
	TRUE  = &value.Boolean{Value: true}
	FALSE = &value.Boolean{Value: false}
)

func Eval(node ast.Node) value.Value {
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node)
	case *ast.ExpressionStatement:
		return Eval(node.Expression)
	case *ast.IntegerLiteral:
		return &value.Integer{Value: node.Value}
	case *ast.Boolean:
		return booleanValue(node.Value)
	case *ast.PrefixExpression:
		right := Eval(node.Right)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left)
		if isError(left) {
			return left
		}

		right := Eval(node.Right)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)
	case *ast.BlockStatment:
		return evalBlockStatement(node)
	case *ast.IfExpression:
		return evalIfExpression(node)
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue)
		if isError(val) {
			return val
		}
		return &value.Return{Value: val}
	}

	return nil
}

func booleanValue(input bool) *value.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func evalProgram(program *ast.Program) value.Value {
	var ret value.Value
	for _, stmt := range program.Statements {
		ret = Eval(stmt)

		switch retType := ret.(type) {
		case *value.Return:
			return retType.Value
		case *value.Error:
			return retType
		}
	}
	return ret
}

func evalBlockStatement(block *ast.BlockStatment) value.Value {
	var ret value.Value
	for _, stmt := range block.Statements {
		ret = Eval(stmt)
		if ret != nil {
			retType := ret.Type()
			if retType == value.RETURN_VALUE || retType == value.ERROR_VALUE {
				return ret
			}
		}
	}
	return ret
}

func evalPrefixExpression(operator string, right value.Value) value.Value {
	switch operator {
	case "!":
		return evalBangOperator(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalInfixExpression(op string, left value.Value, right value.Value) value.Value {
	if left.Type() != right.Type() {
		return newError("type mismatch: %s %s %s", left.Type(), op, right.Type())
	}

	if left.Type() == value.INTEGER_VALUE && right.Type() == value.INTEGER_VALUE {
		return evalIntegerInfixExpression(op, left, right)
	}

	// Use pointer comparison for boolean values
	switch op {
	case "==":
		return booleanValue(left == right)
	case "!=":
		return booleanValue(left != right)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

func evalBangOperator(right value.Value) value.Value {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NIL:
		// TODO: remove this in favour of option type
		return TRUE
	default:
		return FALSE
	}
}

func evalMinusPrefixOperatorExpression(right value.Value) value.Value {
	if right.Type() != value.INTEGER_VALUE {
		return newError("unknown operator: -%s", right.Type())
	}

	val := right.(*value.Integer).Value
	return &value.Integer{Value: -val}
}

func evalIntegerInfixExpression(op string, left, right value.Value) value.Value {
	leftVal := left.(*value.Integer).Value
	rightVal := right.(*value.Integer).Value

	switch op {
	case "+":
		return &value.Integer{Value: leftVal + rightVal}
	case "-":
		return &value.Integer{Value: leftVal - rightVal}
	case "*":
		return &value.Integer{Value: leftVal * rightVal}
	case "/":
		return &value.Integer{Value: leftVal / rightVal}
	case "<":
		return booleanValue(leftVal < rightVal)
	case ">":
		return booleanValue(leftVal > rightVal)
	case "==":
		return booleanValue(leftVal == rightVal)
	case "!=":
		return booleanValue(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

func evalIfExpression(ie *ast.IfExpression) value.Value {
	condition := Eval(ie.Condition)
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(ie.Consequence)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative)
	} else {
		return NIL
	}
}

func isTruthy(val value.Value) bool {
	switch val {
	case NIL, FALSE:
		return false
	case TRUE:
		return true
	default:
		return true
	}
}

func newError(format string, args ...any) *value.Error {
	return &value.Error{Message: fmt.Sprintf(format, args...)}
}

func isError(val value.Value) bool {
	return val != nil && val.Type() == value.ERROR_VALUE
}
