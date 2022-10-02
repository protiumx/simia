package value

import "fmt"

type ValueType string

const (
	INTEGER_VALUE ValueType = "INTEGER"
	BOOLEAN_VALUE ValueType = "BOOLEAN"
	NIL_VALUE     ValueType = "NIL"
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
