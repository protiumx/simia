package compiler

type SymbolScope string

const GlobalScope SymbolScope = "GLOBAL"

type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

type SymbolTable struct {
	store          map[string]Symbol
	numDefinitions int
}

func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	return &SymbolTable{store: s}
}

func (t *SymbolTable) Define(name string) Symbol {
	s := Symbol{name, GlobalScope, t.numDefinitions}
	t.store[name] = s
	t.numDefinitions++
	return s
}

func (t *SymbolTable) Resolve(name string) (Symbol, bool) {
	s, ok := t.store[name]
	return s, ok
}
