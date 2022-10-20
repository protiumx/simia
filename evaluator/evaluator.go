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

func Eval(node ast.Node, env *value.Environment) value.Value {
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node, env)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.IntegerLiteral:
		return &value.Integer{Value: node.Value}
	case *ast.StringLiteral:
		return &value.String{Value: node.Value}
	case *ast.Boolean:
		return booleanValue(node.Value)
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}

		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)
	case *ast.BlockStatment:
		return evalBlockStatement(node, env)
	case *ast.IfExpression:
		return evalIfExpression(node, env)
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &value.Return{Value: val}
	case *ast.LetStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)
	case *ast.Identifier:
		return evalIdentifier(node, env)
	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &value.Function{Parameters: params, Env: env, Body: body}
	case *ast.CallExpression:
		fn := Eval(node.Function, env)
		if isError(fn) {
			return fn
		}
		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}
		return applyFunction(fn, args)
	}

	return nil
}

func booleanValue(input bool) *value.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func evalProgram(program *ast.Program, env *value.Environment) value.Value {
	var ret value.Value
	for _, stmt := range program.Statements {
		ret = Eval(stmt, env)

		switch retType := ret.(type) {
		case *value.Return:
			return retType.Value
		case *value.Error:
			return retType
		}
	}
	return ret
}

func evalBlockStatement(block *ast.BlockStatment, env *value.Environment) value.Value {
	var ret value.Value
	for _, stmt := range block.Statements {
		ret = Eval(stmt, env)
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
	// Do not support expressions between different value types
	if left.Type() != right.Type() {
		return newError("type mismatch: %s %s %s", left.Type(), op, right.Type())
	}

	// Use pointer comparison for boolean values
	switch {
	case left.Type() == value.INTEGER_VALUE:
		return evalIntegerInfixExpression(op, left, right)
	case left.Type() == value.STRING_VALUE:
		return evalStringInfixExpression(op, left, right)
	case op == "==":
		return booleanValue(left == right)
	case op == "!=":
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

func evalIfExpression(ie *ast.IfExpression, env *value.Environment) value.Value {
	condition := Eval(ie.Condition, env)
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env)
	} else {
		return NIL
	}
}

func evalIdentifier(node *ast.Identifier, env *value.Environment) value.Value {
	if val, ok := env.Get(node.Value); ok {
		return val
	}

	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}

	return newError("identifier not found: " + node.Value)
}

func evalExpressions(exps []ast.Expression, env *value.Environment) []value.Value {
	ret := make([]value.Value, len(exps))

	for i, exp := range exps {
		evaluated := Eval(exp, env)
		if isError(evaluated) {
			return []value.Value{evaluated}
		}

		ret[i] = evaluated
	}

	return ret
}

func evalStringInfixExpression(op string, left, right value.Value) value.Value {
	if op != "+" {
		return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
	}

	leftVal := left.(*value.String).Value
	rightVal := right.(*value.String).Value
	return &value.String{Value: leftVal + rightVal}
}

func applyFunction(fnValue value.Value, args []value.Value) value.Value {
	switch fn := fnValue.(type) {
	case *value.Function:
		fnEnv := extendFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, fnEnv)
		return unwrapReturnValue(evaluated)
	case *value.Builtin:
		return fn.Fn(args...)
	default:
		return newError("not a function: %s", fnValue.Type())
	}
}

func extendFunctionEnv(fn *value.Function, args []value.Value) *value.Environment {
	env := value.NewEnvironment(fn.Env)

	for i, param := range fn.Parameters {
		env.Set(param.Value, args[i])
	}

	return env
}

func unwrapReturnValue(val value.Value) value.Value {
	if ret, ok := val.(*value.Return); ok {
		return ret.Value
	}

	return val
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
