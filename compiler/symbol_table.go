package compiler

type SymbolScope string

const (
	GlobalScope   SymbolScope = "GLOBAL"
	LocalScope    SymbolScope = "LOCAL"
	BuiltinScope  SymbolScope = "BUILTIN"
	FreeScope     SymbolScope = "FREE"
	FunctionScope SymbolScope = "FUNCTION"
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

	FreeSymbols []Symbol
}

func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	return &SymbolTable{store: s, FreeSymbols: []Symbol{}}
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
		if !ok {
			return s, ok
		}

		if s.Scope == GlobalScope || s.Scope == BuiltinScope {
			return s, ok
		}

		return t.defineFree(s), true
	}

	return s, ok
}

func (t *SymbolTable) DefineBuiltin(index int, name string) Symbol {
	s := Symbol{Name: name, Index: index, Scope: BuiltinScope}
	t.store[name] = s
	return s
}

func (t *SymbolTable) DefineFunctionName(name string) Symbol {
	// FunctionScopes should only contain the function itself
	s := Symbol{Name: name, Index: 0, Scope: FunctionScope}
	t.store[name] = s
	return s
}

func (t *SymbolTable) defineFree(original Symbol) Symbol {
	t.FreeSymbols = append(t.FreeSymbols, original)
	symbol := Symbol{Name: original.Name, Index: len(t.FreeSymbols) - 1}
	symbol.Scope = FreeScope

	t.store[original.Name] = symbol
	return symbol
}
