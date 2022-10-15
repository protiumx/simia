package value

func NewEnvironment() *Environment {
	s := make(map[string]Value)
	return &Environment{store: s}
}

type Environment struct {
	store map[string]Value
}

func (e *Environment) Get(name string) (Value, bool) {
	val, ok := e.store[name]
	return val, ok
}

func (e *Environment) Set(name string, val Value) Value {
	e.store[name] = val
	return val
}
