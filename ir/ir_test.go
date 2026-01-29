package ir

import (
	"strings"
	"testing"

	"f1c/types"
)

func TestModule_String(t *testing.T) {
	mod := &Module{
		Functions: []*Function{
			{
				Name:       "main",
				Params:     nil,
				ReturnType: types.Void,
				Blocks: []*BasicBlock{
					{
						Label: "entry",
						Instrs: []Instruction{
							&Alloca{Dest: NewReg(0), Type: types.Int},
							&Store{Type: types.Int, Val: &IntConst{Value: 42}, Ptr: NewReg(0)},
						},
						Term: &RetVoid{},
					},
				},
			},
		},
	}

	str := mod.String()
	if !strings.Contains(str, "define void @main()") {
		t.Errorf("expected 'define void @main()', got:\n%s", str)
	}
	if !strings.Contains(str, "entry:") {
		t.Errorf("expected 'entry:', got:\n%s", str)
	}
	if !strings.Contains(str, "alloca") {
		t.Errorf("expected 'alloca', got:\n%s", str)
	}
	if !strings.Contains(str, "ret void") {
		t.Errorf("expected 'ret void', got:\n%s", str)
	}
}

func TestFunction_WithParams(t *testing.T) {
	fn := &Function{
		Name: "add",
		Params: []*Param{
			{Name: "a", Type: types.Int},
			{Name: "b", Type: types.Int},
		},
		ReturnType: types.Int,
		Blocks: []*BasicBlock{
			{
				Label: "entry",
				Instrs: []Instruction{
					&BinOp{
						Dest: NewReg(0),
						Op:   OpAdd,
						Type: types.Int,
						Left: &ParamRef{Name: "a"},
						Right: &ParamRef{Name: "b"},
					},
				},
				Term: &Ret{Type: types.Int, Val: NewReg(0)},
			},
		},
	}

	str := fn.String()
	if !strings.Contains(str, "define int @add(int %a, int %b)") {
		t.Errorf("expected function signature, got:\n%s", str)
	}
	if !strings.Contains(str, "add int") {
		t.Errorf("expected 'add int', got:\n%s", str)
	}
	if !strings.Contains(str, "ret int") {
		t.Errorf("expected 'ret int', got:\n%s", str)
	}
}

func TestBasicBlock_ConditionalBranch(t *testing.T) {
	block := &BasicBlock{
		Label: "entry",
		Instrs: []Instruction{
			&BinOp{Dest: NewReg(0), Op: OpGt, Type: types.Int, Left: &ParamRef{Name: "x"}, Right: &IntConst{Value: 0}},
		},
		Term: &CondBr{Cond: NewReg(0), Then: "positive", Else: "negative"},
	}

	str := block.String()
	if !strings.Contains(str, "entry:") {
		t.Errorf("expected 'entry:', got:\n%s", str)
	}
	if !strings.Contains(str, "br %0, positive, negative") {
		t.Errorf("expected conditional branch, got:\n%s", str)
	}
}

func TestInstructions_String(t *testing.T) {
	tests := []struct {
		name     string
		instr    Instruction
		expected string
	}{
		{
			name:     "alloca",
			instr:    &Alloca{Dest: NewReg(0), Type: types.Int},
			expected: "%0 = alloca int",
		},
		{
			name:     "load",
			instr:    &Load{Dest: NewReg(1), Type: types.Int, Ptr: NewReg(0)},
			expected: "%1 = load int, %0",
		},
		{
			name:     "store",
			instr:    &Store{Type: types.Int, Val: &IntConst{Value: 5}, Ptr: NewReg(0)},
			expected: "store int 5, %0",
		},
		{
			name:     "add",
			instr:    &BinOp{Dest: NewReg(0), Op: OpAdd, Type: types.Int, Left: &IntConst{Value: 1}, Right: &IntConst{Value: 2}},
			expected: "%0 = add int 1, 2",
		},
		{
			name:     "sub",
			instr:    &BinOp{Dest: NewReg(0), Op: OpSub, Type: types.Int, Left: NewReg(1), Right: &IntConst{Value: 1}},
			expected: "%0 = sub int %1, 1",
		},
		{
			name:     "mul",
			instr:    &BinOp{Dest: NewReg(0), Op: OpMul, Type: types.Int, Left: NewReg(1), Right: NewReg(2)},
			expected: "%0 = mul int %1, %2",
		},
		{
			name:     "div",
			instr:    &BinOp{Dest: NewReg(0), Op: OpDiv, Type: types.Int, Left: NewReg(1), Right: NewReg(2)},
			expected: "%0 = div int %1, %2",
		},
		{
			name:     "mod",
			instr:    &BinOp{Dest: NewReg(0), Op: OpMod, Type: types.Int, Left: NewReg(1), Right: NewReg(2)},
			expected: "%0 = mod int %1, %2",
		},
		{
			name:     "eq",
			instr:    &BinOp{Dest: NewReg(0), Op: OpEq, Type: types.Int, Left: NewReg(1), Right: NewReg(2)},
			expected: "%0 = eq int %1, %2",
		},
		{
			name:     "neq",
			instr:    &BinOp{Dest: NewReg(0), Op: OpNeq, Type: types.Int, Left: NewReg(1), Right: NewReg(2)},
			expected: "%0 = neq int %1, %2",
		},
		{
			name:     "lt",
			instr:    &BinOp{Dest: NewReg(0), Op: OpLt, Type: types.Int, Left: NewReg(1), Right: NewReg(2)},
			expected: "%0 = lt int %1, %2",
		},
		{
			name:     "gt",
			instr:    &BinOp{Dest: NewReg(0), Op: OpGt, Type: types.Int, Left: NewReg(1), Right: NewReg(2)},
			expected: "%0 = gt int %1, %2",
		},
		{
			name:     "and",
			instr:    &BinOp{Dest: NewReg(0), Op: OpAnd, Type: types.Bool, Left: &BoolConst{Value: true}, Right: &BoolConst{Value: false}},
			expected: "%0 = and bool true, false",
		},
		{
			name:     "or",
			instr:    &BinOp{Dest: NewReg(0), Op: OpOr, Type: types.Bool, Left: &BoolConst{Value: true}, Right: &BoolConst{Value: false}},
			expected: "%0 = or bool true, false",
		},
		{
			name:     "neg",
			instr:    &UnaryOp{Dest: NewReg(0), Op: OpNeg, Type: types.Int, Operand: NewReg(1)},
			expected: "%0 = neg int %1",
		},
		{
			name:     "not",
			instr:    &UnaryOp{Dest: NewReg(0), Op: OpNot, Type: types.Bool, Operand: NewReg(1)},
			expected: "%0 = not bool %1",
		},
		{
			name:     "call",
			instr:    &Call{Dest: NewReg(0), Func: "add", Args: []Value{&IntConst{Value: 1}, &IntConst{Value: 2}}, RetType: types.Int},
			expected: "%0 = call int @add(1, 2)",
		},
		{
			name:     "call void",
			instr:    &Call{Func: "print", Args: []Value{&IntConst{Value: 42}}, RetType: types.Void},
			expected: "call void @print(42)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.instr.String()
			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestTerminators_String(t *testing.T) {
	tests := []struct {
		name     string
		term     Terminator
		expected string
	}{
		{
			name:     "ret void",
			term:     &RetVoid{},
			expected: "ret void",
		},
		{
			name:     "ret int",
			term:     &Ret{Type: types.Int, Val: NewReg(0)},
			expected: "ret int %0",
		},
		{
			name:     "ret const",
			term:     &Ret{Type: types.Int, Val: &IntConst{Value: 42}},
			expected: "ret int 42",
		},
		{
			name:     "br unconditional",
			term:     &Br{Target: "loop"},
			expected: "br loop",
		},
		{
			name:     "br conditional",
			term:     &CondBr{Cond: NewReg(0), Then: "then", Else: "else"},
			expected: "br %0, then, else",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.term.String()
			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestValues_String(t *testing.T) {
	tests := []struct {
		name     string
		val      Value
		expected string
	}{
		{"int const", &IntConst{Value: 42}, "42"},
		{"negative const", &IntConst{Value: -5}, "-5"},
		{"bool true", &BoolConst{Value: true}, "true"},
		{"bool false", &BoolConst{Value: false}, "false"},
		{"register 0", NewReg(0), "%0"},
		{"register 10", NewReg(10), "%10"},
		{"param", &ParamRef{Name: "x"}, "%x"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.val.String()
			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestNewReg(t *testing.T) {
	r0 := NewReg(0)
	r1 := NewReg(1)

	if r0.String() != "%0" {
		t.Errorf("NewReg(0).String() = %q, want %%0", r0.String())
	}
	if r1.String() != "%1" {
		t.Errorf("NewReg(1).String() = %q, want %%1", r1.String())
	}
}
