package value

func NewEnvironment(outer *Environment) *Environment {
	s := make(map[string]Value)
	return &Environment{store: s, outer: outer}
}

type Environment struct {
	store map[string]Value
	outer *Environment
}

func (e *Environment) Get(name string) (Value, bool) {
	val, ok := e.store[name]
	if !ok && e.outer != nil {
		val, ok = e.outer.Get(name)
	}
	return val, ok
}

func (e *Environment) Set(name string, val Value) Value {
	e.store[name] = val
	return val
}
