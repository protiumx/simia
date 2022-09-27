package value

import "fmt"

type ValueType string

const (
	INTEGER_VALUE ValueType = "INTEGER"
	BOOLEAN_VALUE ValueType = "BOOLEAN"
	NONE_VALUE    ValueType = "NONE"
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

type None struct{}

func (n *None) Inspect() string {
	return "null"
}

func (n *None) Type() ValueType {
	return NONE_VALUE
}
