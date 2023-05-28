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
				{Name: "b", Scope: FreeScope, Index: 0},
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
		{Name: "a", Scope: BuiltinScope, Index: 0},
		{Name: "c", Scope: BuiltinScope, Index: 1},
		{Name: "e", Scope: BuiltinScope, Index: 2},
		{Name: "f", Scope: BuiltinScope, Index: 3},
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

func TestResolveFree(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	local1 := NewEnclosedSymbolTable(global)
	local1.Define("c")
	local1.Define("d")

	local2 := NewEnclosedSymbolTable(local1)
	local2.Define("e")
	local2.Define("f")

	tests := []struct {
		table               *SymbolTable
		expectedSymbols     []Symbol
		expectedFreeSymbols []Symbol
	}{
		{
			local1,
			[]Symbol{
				{Name: "a", Scope: GlobalScope, Index: 0},
				{Name: "b", Scope: GlobalScope, Index: 1},
				{Name: "c", Scope: LocalScope, Index: 0},
				{Name: "d", Scope: LocalScope, Index: 1},
			},
			[]Symbol{},
		},
		{
			local2,
			[]Symbol{
				{Name: "a", Scope: GlobalScope, Index: 0},
				{Name: "b", Scope: GlobalScope, Index: 1},
				{Name: "c", Scope: FreeScope, Index: 0},
				{Name: "d", Scope: FreeScope, Index: 1},
				{Name: "e", Scope: LocalScope, Index: 0},
				{Name: "f", Scope: LocalScope, Index: 1},
			},
			[]Symbol{
				{Name: "c", Scope: LocalScope, Index: 0},
				{Name: "d", Scope: LocalScope, Index: 1},
			},
		},
	}

	for _, tt := range tests {
		for _, sym := range tt.expectedSymbols {
			result, ok := tt.table.Resolve(sym.Name)
			if !ok {
				t.Errorf("name %s not resolvable", sym.Name)
			}

			if result != sym {
				t.Errorf("expected %s to resolve to %v, got=%+v", sym.Name, sym, result)
			}
		}

		if len(tt.table.FreeSymbols) != len(tt.expectedFreeSymbols) {
			t.Errorf("wrong number of free symbols. got=%d, want=%d", len(tt.table.FreeSymbols), len(tt.expectedFreeSymbols))
			continue
		}

		for i, sym := range tt.expectedFreeSymbols {
			result := tt.table.FreeSymbols[i]
			if result != sym {
				t.Errorf("wrong free symbol. got=%+v, want=%+v", result, sym)
			}

		}
	}
}

func TestResolveUnresolvableFree(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")

	local1 := NewEnclosedSymbolTable(global)
	local1.Define("c")

	local2 := NewEnclosedSymbolTable(local1)
	local2.Define("d")

	expected := []Symbol{
		{Name: "a", Scope: GlobalScope, Index: 0},
		{Name: "c", Scope: FreeScope, Index: 0},
		{Name: "d", Scope: LocalScope, Index: 0},
	}

	for _, sym := range expected {
		result, ok := local2.Resolve(sym.Name)
		if !ok {
			t.Errorf("name %s not resolvable", sym.Name)
			continue
		}

		if result != sym {
			t.Errorf("expected %s to resolve to %+v, got=%+v", sym.Name, sym, result)
		}
	}

	expectedUnresolvable := []string{"e", "b"}
	for _, name := range expectedUnresolvable {
		_, ok := local2.Resolve(name)
		if ok {
			t.Errorf("name %s resolved but was expected not to", name)
		}
	}
}

func TestDefineAndResolveFunctionName(t *testing.T) {
	global := NewSymbolTable()
	global.DefineFunctionName("a")

	expected := Symbol{Name: "a", Scope: FunctionScope, Index: 0}
	result, ok := global.Resolve(expected.Name)
	if !ok {
		t.Fatalf("function name %s is not resolvable", expected.Name)
	}

	if result != expected {
		t.Errorf("expected %s to resolve to %v, got=%+v", expected.Name, expected, result)
	}
}

func TestShadowingFunctionName(t *testing.T) {
	global := NewSymbolTable()
	global.DefineFunctionName("a")
	global.Define("a")

	expected := Symbol{Name: "a", Scope: GlobalScope, Index: 0}
	result, ok := global.Resolve(expected.Name)
	if !ok {
		t.Fatalf("function name %s is not resolvable", expected.Name)
	}

	if result != expected {
		t.Errorf("expected %s to resolve to %v, got=%+v", expected.Name, expected, result)
	}

}
