package value

import "fmt"

type ValueType string

const (
	INTEGER_VALUE ValueType = "INTEGER"
	BOOLEAN_VALUE           = "BOOLEAN"
	NIL_VALUE               = "NIL"
	RETURN_VALUE            = "RETURN"
	ERROR_VALUE             = "ERROR"
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
