package evaluator

import (
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
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left)
		right := Eval(node.Right)
		return evalInfixExpression(node.Operator, left, right)
	case *ast.BlockStatment:
		return evalBlockStatement(node)
	case *ast.IfExpression:
		return evalIfExpression(node)
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue)
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

		if returnVal, ok := ret.(*value.Return); ok {
			return returnVal.Value
		}
	}
	return ret
}

func evalBlockStatement(block *ast.BlockStatment) value.Value {
	var ret value.Value
	for _, stmt := range block.Statements {
		ret = Eval(stmt)
		if ret != nil && ret.Type() == value.RETURN_VALUE {
			return ret
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
		return NIL
	}
}

func evalInfixExpression(op string, left value.Value, right value.Value) value.Value {
	if left.Type() == value.INTEGER_VALUE && right.Type() == value.INTEGER_VALUE {
		return evalIntegerInfixExpression(op, left, right)
	}
	// Use pointer comparison for boolean values
	switch op {
	case "==":
		return booleanValue(left == right)
	case "!=":
		return booleanValue(left != right)
	}

	return NIL
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
		return NIL
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
		return NIL
	}
}

func evalIfExpression(ie *ast.IfExpression) value.Value {
	condition := Eval(ie.Condition)
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
