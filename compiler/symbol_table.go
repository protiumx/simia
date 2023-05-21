package compiler

type SymbolScope string

const (
	GlobalScope  SymbolScope = "GLOBAL"
	LocalScope   SymbolScope = "LOCAL"
	BuiltinScope SymbolScope = "BUILTIN"
)

type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

type SymbolTable struct {
	Outer *SymbolTable

	store       map[string]Symbol
	definitions int
}

func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	return &SymbolTable{store: s}
}

func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	s := NewSymbolTable()
	s.Outer = outer
	return s
}

func (t *SymbolTable) Define(name string) Symbol {
	s := Symbol{name, GlobalScope, t.definitions}
	if t.Outer == nil {
		s.Scope = GlobalScope
	} else {
		s.Scope = LocalScope
	}

	t.store[name] = s
	t.definitions++
	return s
}

func (t *SymbolTable) Resolve(name string) (Symbol, bool) {
	s, ok := t.store[name]
	if !ok && t.Outer != nil {
		s, ok = t.Outer.Resolve(name)
		return s, ok
	}
	return s, ok
}

func (t *SymbolTable) DefineBuiltin(index int, name string) Symbol {
	s := Symbol{Name: name, Index: index, Scope: BuiltinScope}
	t.store[name] = s
	return s
}
