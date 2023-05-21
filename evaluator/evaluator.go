package evaluator

import (
	"fmt"

	"protiumx.dev/simia/ast"
	"protiumx.dev/simia/token"
	"protiumx.dev/simia/value"
)

const loopLimit = 10000

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
		// Eval pipiline expression
		if node.Operator == token.PIPELINE {
			fnCall, ok := node.Right.(*ast.CallExpression)
			if !ok {
				return newError("expected function call in pipiline expression. got=%T", node.Right)
			}
			// prepend the left node to the fn arguments
			fnCall.Arguments = append([]ast.Expression{node.Left}, fnCall.Arguments...)
			return Eval(node.Right, env)
		}

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

	case *ast.ForExpression:
		return evalForExpression(node, env)

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

	case *ast.ArrayLiteral:
		elements := evalExpressions(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}

		return &value.Array{Elements: elements}

	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}

		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}

		return evalIndexExpression(left, index)

	case *ast.HashLiteral:
		return evalHashLiteral(node, env)

	case *ast.AssignExpression:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}

		_, ok := env.Get(node.Identifier.Value)
		if !ok {
			return newError("error assigning undeclared variable \"%s\"", node.Identifier.Value)
		}
		env.Set(node.Identifier.Value, val)
		return NIL
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
	case "..":
		return rangeValue(leftVal, rightVal)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

func rangeValue(left, right int64) value.Value {
	if left == right {
		return newError("range start and end must be different: %d..%d", left, right)
	}
	return &value.Range{Start: left, End: right}
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

func evalForExpression(exp *ast.ForExpression, env *value.Environment) value.Value {
	switch condition := exp.Condition.(type) {
	case *ast.InExpression:
		loopEnv := value.NewEnvironment(env)
		elementIdentifier, ok := condition.Element.(*ast.Identifier)
		if !ok {
			return newError("expected identifier on left side of in-expression. got=%T", condition.Element)
		}

		iterable := Eval(condition.Iterable, env)
		if isError(iterable) {
			return iterable
		}

		switch iterable := iterable.(type) {
		case *value.Range:
			evalForLoopRange(elementIdentifier, iterable, exp.Body, loopEnv)
		case *value.Array:
			evalForLoopArray(elementIdentifier, iterable, exp.Body, loopEnv)
		default:
			return newError("for-loop not supported for type %s", iterable.Type())
		}

	case *ast.InfixExpression, *ast.Identifier, *ast.Boolean, *ast.IntegerLiteral:
		loopEnv := value.NewEnvironment(env)
		return evalForLoopCondition(condition, exp.Body, loopEnv)
	default:
		return newError("invalid %T expression in for-loop", exp.Condition)
	}

	return NIL
}

func evalForLoopCondition(condition ast.Expression, body *ast.BlockStatment, env *value.Environment) value.Value {
	loopCounter := 0
	for {
		if loopCounter > loopLimit {
			return newError("max loop call exceed")
		}

		e := Eval(condition, env)
		if isError(e) {
			return e
		}

		if !isTruthy(e) {
			return NIL
		}

		e = evalBlockStatement(body, env)
		if isError(e) {
			return e
		}
		loopCounter++
	}
}

func evalForLoopArray(
	elementIdentifier *ast.Identifier,
	array *value.Array,
	body *ast.BlockStatment,
	env *value.Environment,
) value.Value {
	for _, element := range array.Elements {
		env.Set(elementIdentifier.Value, element)
		r := evalBlockStatement(body, env)
		if isError(r) {
			return r
		}
	}
	return NIL
}

func evalForLoopRange(
	elementIdentifier *ast.Identifier,
	rangeVal *value.Range,
	body *ast.BlockStatment,
	env *value.Environment,
) value.Value {
	ascDirection := true
	if rangeVal.End < rangeVal.Start {
		ascDirection = false
	}

	currentValue := value.Integer{Value: rangeVal.Start}
	env.Set(elementIdentifier.Value, &currentValue)
	loopCounter := 0

	for currentValue.Value != rangeVal.End {
		if loopCounter > loopLimit {
			return newError("max loop call exceed")
		}

		e := evalBlockStatement(body, env)
		if isError(e) {
			return e
		}

		if ascDirection {
			currentValue.Value++
		} else {
			currentValue.Value--
		}
		loopCounter++
	}
	return NIL
}

func evalIdentifier(node *ast.Identifier, env *value.Environment) value.Value {
	if val, ok := env.Get(node.Value); ok {
		return val
	}

	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}

	return newError("%s not defined", node.Value)
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
		if result := fn.Fn(args...); result != nil {
			return result
		}

		return NIL

	default:
		return newError("not a function: %s", fnValue.Type())
	}
}

func evalIndexExpression(left, index value.Value) value.Value {
	switch {
	case left.Type() == value.ARRAY_VALUE && index.Type() == value.INTEGER_VALUE:
		return evalArrayIndexExpression(left, index)
	case left.Type() == value.HASH_VALUE:
		return evalHashIndexExpression(left, index)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

func evalArrayIndexExpression(arrayVal, index value.Value) value.Value {
	array := arrayVal.(*value.Array)
	idx := index.(*value.Integer).Value
	max := int64(len(array.Elements) - 1)
	if idx < 0 || idx > max {
		// TODO: return error
		return NIL
	}

	return array.Elements[idx]
}

func evalHashLiteral(node *ast.HashLiteral, env *value.Environment) value.Value {
	pairs := make(map[string]value.Value)

	for k, v := range node.Pairs {
		key := Eval(k, env)
		if isError(key) {
			return key
		}

		var keyVal string
		switch keyType := key.(type) {
		case *value.String:
			keyVal = keyType.Value
		default:
			return newError("key is not string: %s", key.Type())
		}

		value := Eval(v, env)
		if isError(value) {
			return value
		}

		pairs[keyVal] = value
	}

	return &value.Hash{Pairs: pairs}
}

func evalHashIndexExpression(hash, index value.Value) value.Value {
	hashVal := hash.(*value.Hash)

	key, ok := index.(*value.String)
	if !ok {
		return newError("key is not string: %s", index.Type())
	}

	val, ok := hashVal.Pairs[key.Value]
	if !ok {
		return NIL
	}

	return val
}

func extendFunctionEnv(fn *value.Function, args []value.Value) *value.Environment {
	return extendEnv(fn.Env, fn.Parameters, args)
}

func extendEnv(currentEnv *value.Environment, identifiers []*ast.Identifier, values []value.Value) *value.Environment {
	env := value.NewEnvironment(currentEnv)

	for i, identifier := range identifiers {
		env.Set(identifier.Value, values[i])
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
	switch val := val.(type) {
	case *value.Integer:
		return val.Value != 0
	case *value.Boolean:
		return val.Value
	default:
		return false
	}
}

func newError(format string, args ...any) *value.Error {
	return &value.Error{Message: fmt.Sprintf(format, args...)}
}

func isError(val value.Value) bool {
	return val != nil && val.Type() == value.ERROR_VALUE
}
