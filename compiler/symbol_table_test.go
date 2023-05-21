package compiler

import "testing"

func TestDefine(t *testing.T) {
	expected := map[string]Symbol{
		"a": {Name: "a", Scope: GlobalScope, Index: 0},
		"b": {Name: "b", Scope: GlobalScope, Index: 1},
		"c": {Name: "c", Scope: LocalScope, Index: 0},
		"d": {Name: "d", Scope: LocalScope, Index: 1},
		"e": {Name: "e", Scope: LocalScope, Index: 0},
		"f": {Name: "f", Scope: LocalScope, Index: 1},
	}

	global := NewSymbolTable()

	for _, k := range [...]string{"a", "b"} {
		s := global.Define(k)
		if s != expected[k] {
			t.Errorf("expected %s=%+v, got=%+v", k, expected[k], s)
		}
	}

	localA := NewEnclosedSymbolTable(global)
	for _, k := range [...]string{"c", "d"} {
		s := localA.Define(k)
		if s != expected[k] {
			t.Errorf("expected %s=%+v, got=%+v", k, expected[k], s)
		}
	}

	localB := NewEnclosedSymbolTable(localA)
	for _, k := range [...]string{"e", "f"} {
		s := localB.Define(k)
		if s != expected[k] {
			t.Errorf("expected %s=%+v, got=%+v", k, expected[k], s)
		}
	}
}

func TestResolveGlobal(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	expected := map[string]Symbol{
		"a": {Name: "a", Scope: GlobalScope, Index: 0},
		"b": {Name: "b", Scope: GlobalScope, Index: 1},
	}

	for _, s := range expected {
		resolved, ok := global.Resolve(s.Name)
		if !ok {
			t.Errorf("name %s not defiend", s.Name)
			continue
		}

		if resolved != s {
			t.Errorf("expected %s to resolve to %+v, got=%+v", s.Name, s, resolved)
		}
	}
}

func TestResolveLocal(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	local := NewEnclosedSymbolTable(global)
	local.Define("c")
	local.Define("d")

	expected := []Symbol{
		{Name: "a", Scope: GlobalScope, Index: 0},
		{Name: "b", Scope: GlobalScope, Index: 1},
		{Name: "c", Scope: LocalScope, Index: 0},
		{Name: "d", Scope: LocalScope, Index: 1},
	}

	for _, sym := range expected {
		checkSymbol(t, local, sym)
	}
}

func checkSymbol(t *testing.T, s *SymbolTable, sym Symbol) {
	result, ok := s.Resolve(sym.Name)
	if !ok {
		t.Errorf("name %s not resolvable", sym.Name)
		return
	}
	if result != sym {
		t.Errorf("expected %s to resolve to %+v, got=%+v", sym.Name, sym, result)
	}
}

func TestResolveNestedLocal(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")

	first := NewEnclosedSymbolTable(global)
	first.Define("b")

	second := NewEnclosedSymbolTable(first)
	second.Define("c")

	tests := []struct {
		table           *SymbolTable
		expectedSymbols []Symbol
	}{
		{
			first,
			[]Symbol{
				{Name: "a", Scope: GlobalScope, Index: 0},
				{Name: "b", Scope: LocalScope, Index: 0},
			},
		},
		{
			second,
			[]Symbol{
				{Name: "a", Scope: GlobalScope, Index: 0},
				{Name: "b", Scope: LocalScope, Index: 0},
				{Name: "c", Scope: LocalScope, Index: 0},
			},
		},
	}

	for _, tt := range tests {
		for _, sym := range tt.expectedSymbols {
			checkSymbol(t, tt.table, sym)
		}
	}
}

func TestDefineResolveBuiltins(t *testing.T) {
	global := NewSymbolTable()
	localA := NewEnclosedSymbolTable(global)
	localB := NewEnclosedSymbolTable(localA)

	expected := []Symbol{
		Symbol{Name: "a", Scope: BuiltinScope, Index: 0},
		Symbol{Name: "c", Scope: BuiltinScope, Index: 1},
		Symbol{Name: "e", Scope: BuiltinScope, Index: 2},
		Symbol{Name: "f", Scope: BuiltinScope, Index: 3},
	}

	for i, v := range expected {
		global.DefineBuiltin(i, v.Name)
	}

	for _, table := range []*SymbolTable{global, localA, localB} {
		for _, sym := range expected {
			result, ok := table.Resolve(sym.Name)
			if !ok {
				t.Errorf("name %s not resolvable", sym.Name)
				continue
			}

			if result != sym {
				t.Errorf("expected %s to resolve %+v, got=%+v", sym.Name, sym, result)
			}
		}
	}
}
