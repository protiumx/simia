package evaluator

import (
	"protiumx.dev/simia/ast"
	"protiumx.dev/simia/value"
)

func Eval(node ast.Node) value.Value {
	switch node := node.(type) {
	case *ast.Program:
		return evalStatements(node.Statements)
	case *ast.ExpressionStatement:
		return Eval(node.Expression)
	case *ast.IntegerLiteral:
		return &value.Integer{Value: node.Value}
	}

	return nil
}

func evalStatements(stmts []ast.Statement) value.Value {
	var ret value.Value
	for _, stmt := range stmts {
		ret = Eval(stmt)
	}
	return ret
}
