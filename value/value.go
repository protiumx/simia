package value

import (
	"fmt"
	"strings"

	"protiumx.dev/simia/ast"
)

type ValueType string

const (
	INTEGER_VALUE  ValueType = "INTEGER"
	STRING_VALUE             = "STRING"
	BOOLEAN_VALUE            = "BOOLEAN"
	NIL_VALUE                = "NIL"
	RETURN_VALUE             = "RETURN"
	ERROR_VALUE              = "ERROR"
	FUNCTION_VALUE           = "FN"
	BUILTIN_VALUE            = "BUILTIN"
	ARRAY_VALUE              = "ARRAY"
	HASH_VALUE               = "HASH"
	RANGE_VALUE              = "RANGE"
	// For expressions that do not return a value
	EMPTY_VALUE = "EMPTY"
)

type Value interface {
	Type() ValueType
	Inspect() string
}

type Integer struct {
	Value int64
}

func (i *Integer) Inspect() string {
	return fmt.Sprintf("%d", i.Value)
}

func (i *Integer) Type() ValueType {
	return INTEGER_VALUE
}

type String struct {
	Value string
}

func (s *String) Type() ValueType {
	return STRING_VALUE
}

func (s *String) Inspect() string {
	return s.Value
}

type Boolean struct {
	Value bool
}

func (b *Boolean) Inspect() string {
	return fmt.Sprintf("%t", b.Value)
}

func (b *Boolean) Type() ValueType {
	return BOOLEAN_VALUE
}

type Nil struct{}

func (n *Nil) Inspect() string {
	return "nil"
}

func (n *Nil) Type() ValueType {
	return NIL_VALUE
}

type Return struct {
	Value Value
}

func (r *Return) Type() ValueType { return RETURN_VALUE }

func (r *Return) Inspect() string {
	return r.Value.Inspect()
}

// TODO: add line and column
type Error struct {
	Message string
}

func (e *Error) Type() ValueType {
	return ERROR_VALUE
}

func (e *Error) Inspect() string {
	return "ERROR: " + e.Message
}

type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatment
	Env        *Environment
}

func (fn *Function) Type() ValueType {
	return FUNCTION_VALUE
}

func (fn *Function) Inspect() string {
	var out strings.Builder
	params := make([]string, len(fn.Parameters))
	for i, p := range fn.Parameters {
		params[i] = p.String()
	}

	out.WriteString("fn")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ","))
	out.WriteString(") {\n")
	out.WriteString(fn.Body.String())
	out.WriteString("\n}")

	return out.String()
}

type BuiltinFunction func(args ...Value) Value

type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() ValueType {
	return BUILTIN_VALUE
}

func (b *Builtin) Inspect() string {
	return "builtin function"
}

type Array struct {
	Elements []Value
}

func (a *Array) Type() ValueType {
	return ARRAY_VALUE
}

func (a *Array) Inspect() string {
	var out strings.Builder

	elements := make([]string, len(a.Elements))
	for i, e := range a.Elements {
		elements[i] = e.Inspect()
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}

type Hash struct {
	Pairs map[string]Value
}

func (h *Hash) Type() ValueType {
	return HASH_VALUE
}

func (h *Hash) Inspect() string {
	var out strings.Builder

	pairs := make([]string, len(h.Pairs), len(h.Pairs))
	i := 0
	for k, v := range h.Pairs {
		pairs[i] = fmt.Sprintf("%s: %s", k, v.Inspect())
		i++
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")
	return out.String()
}

type Range struct {
	Start int64
	End   int64
}

func (r *Range) Type() ValueType {
	return RANGE_VALUE
}

func (r *Range) Inspect() string {
	var out strings.Builder
	out.WriteString("[")
	out.WriteString(fmt.Sprintf("%d", r.Start))
	out.WriteString("..")
	out.WriteString(fmt.Sprintf("%d", r.End))
	out.WriteString("]")
	return out.String()
}
