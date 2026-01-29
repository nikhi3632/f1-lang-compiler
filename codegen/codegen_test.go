package codegen

import (
	"strings"
	"testing"

	"f1c/ir"
	"f1c/types"
)

func TestGenerate_EmptyModule(t *testing.T) {
	mod := &ir.Module{}
	gen := New()
	output := gen.Generate(mod)

	// Empty module should produce minimal valid LLVM IR
	if !strings.Contains(output, "target triple") {
		t.Errorf("output should contain target triple declaration")
	}
}

func TestGenerate_SimpleFunction(t *testing.T) {
	// Create: define i64 @main() { entry: ret i64 0 }
	mod := &ir.Module{
		Functions: []*ir.Function{
			{
				Name:       "main",
				Params:     []*ir.Param{},
				ReturnType: types.Int,
				Blocks: []*ir.BasicBlock{
					{
						Label:  "entry",
						Instrs: []ir.Instruction{},
						Term:   &ir.Ret{Type: types.Int, Val: &ir.IntConst{Value: 0}},
					},
				},
			},
		},
	}

	gen := New()
	output := gen.Generate(mod)

	expectedParts := []string{
		"define i64 @main()",
		"entry:",
		"ret i64 0",
	}

	for _, part := range expectedParts {
		if !strings.Contains(output, part) {
			t.Errorf("output should contain %q, got:\n%s", part, output)
		}
	}
}

func TestGenerate_FunctionWithParams(t *testing.T) {
	// Create: define i64 @add(i64 %a, i64 %b) { ... }
	mod := &ir.Module{
		Functions: []*ir.Function{
			{
				Name: "add",
				Params: []*ir.Param{
					{Name: "a", Type: types.Int},
					{Name: "b", Type: types.Int},
				},
				ReturnType: types.Int,
				Blocks: []*ir.BasicBlock{
					{
						Label:  "entry",
						Instrs: []ir.Instruction{},
						Term:   &ir.Ret{Type: types.Int, Val: &ir.ParamRef{Name: "a"}},
					},
				},
			},
		},
	}

	gen := New()
	output := gen.Generate(mod)

	if !strings.Contains(output, "define i64 @add(i64 %a, i64 %b)") {
		t.Errorf("output should contain function signature with params, got:\n%s", output)
	}
}

func TestGenerate_Alloca(t *testing.T) {
	mod := &ir.Module{
		Functions: []*ir.Function{
			{
				Name:       "main",
				ReturnType: types.Void,
				Blocks: []*ir.BasicBlock{
					{
						Label: "entry",
						Instrs: []ir.Instruction{
							&ir.Alloca{Dest: ir.NewReg(0), Type: types.Int},
						},
						Term: &ir.RetVoid{},
					},
				},
			},
		},
	}

	gen := New()
	output := gen.Generate(mod)

	if !strings.Contains(output, "%0 = alloca i64") {
		t.Errorf("output should contain alloca instruction, got:\n%s", output)
	}
}

func TestGenerate_LoadStore(t *testing.T) {
	mod := &ir.Module{
		Functions: []*ir.Function{
			{
				Name:       "main",
				ReturnType: types.Int,
				Blocks: []*ir.BasicBlock{
					{
						Label: "entry",
						Instrs: []ir.Instruction{
							&ir.Alloca{Dest: ir.NewReg(0), Type: types.Int},
							&ir.Store{Type: types.Int, Val: &ir.IntConst{Value: 42}, Ptr: ir.NewReg(0)},
							&ir.Load{Dest: ir.NewReg(1), Type: types.Int, Ptr: ir.NewReg(0)},
						},
						Term: &ir.Ret{Type: types.Int, Val: ir.NewReg(1)},
					},
				},
			},
		},
	}

	gen := New()
	output := gen.Generate(mod)

	expectedParts := []string{
		"%0 = alloca i64",
		"store i64 42, ptr %0",
		"%1 = load i64, ptr %0",
		"ret i64 %1",
	}

	for _, part := range expectedParts {
		if !strings.Contains(output, part) {
			t.Errorf("output should contain %q, got:\n%s", part, output)
		}
	}
}

func TestGenerate_BinaryOps(t *testing.T) {
	tests := []struct {
		name     string
		op       ir.Op
		expected string
	}{
		{"add", ir.OpAdd, "%2 = add i64 %0, %1"},
		{"sub", ir.OpSub, "%2 = sub i64 %0, %1"},
		{"mul", ir.OpMul, "%2 = mul i64 %0, %1"},
		{"div", ir.OpDiv, "%2 = sdiv i64 %0, %1"},
		{"mod", ir.OpMod, "%2 = srem i64 %0, %1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mod := &ir.Module{
				Functions: []*ir.Function{
					{
						Name:       "test",
						ReturnType: types.Int,
						Blocks: []*ir.BasicBlock{
							{
								Label: "entry",
								Instrs: []ir.Instruction{
									&ir.BinOp{
										Dest:  ir.NewReg(2),
										Op:    tt.op,
										Type:  types.Int,
										Left:  ir.NewReg(0),
										Right: ir.NewReg(1),
									},
								},
								Term: &ir.Ret{Type: types.Int, Val: ir.NewReg(2)},
							},
						},
					},
				},
			}

			gen := New()
			output := gen.Generate(mod)

			if !strings.Contains(output, tt.expected) {
				t.Errorf("output should contain %q, got:\n%s", tt.expected, output)
			}
		})
	}
}

func TestGenerate_ComparisonOps(t *testing.T) {
	tests := []struct {
		name     string
		op       ir.Op
		expected string
	}{
		{"eq", ir.OpEq, "%2 = icmp eq i64 %0, %1"},
		{"neq", ir.OpNeq, "%2 = icmp ne i64 %0, %1"},
		{"lt", ir.OpLt, "%2 = icmp slt i64 %0, %1"},
		{"gt", ir.OpGt, "%2 = icmp sgt i64 %0, %1"},
		{"lte", ir.OpLte, "%2 = icmp sle i64 %0, %1"},
		{"gte", ir.OpGte, "%2 = icmp sge i64 %0, %1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mod := &ir.Module{
				Functions: []*ir.Function{
					{
						Name:       "test",
						ReturnType: types.Bool,
						Blocks: []*ir.BasicBlock{
							{
								Label: "entry",
								Instrs: []ir.Instruction{
									&ir.BinOp{
										Dest:  ir.NewReg(2),
										Op:    tt.op,
										Type:  types.Int,
										Left:  ir.NewReg(0),
										Right: ir.NewReg(1),
									},
								},
								Term: &ir.Ret{Type: types.Bool, Val: ir.NewReg(2)},
							},
						},
					},
				},
			}

			gen := New()
			output := gen.Generate(mod)

			if !strings.Contains(output, tt.expected) {
				t.Errorf("output should contain %q, got:\n%s", tt.expected, output)
			}
		})
	}
}

func TestGenerate_LogicalOps(t *testing.T) {
	tests := []struct {
		name     string
		op       ir.Op
		expected string
	}{
		{"and", ir.OpAnd, "%2 = and i1 %0, %1"},
		{"or", ir.OpOr, "%2 = or i1 %0, %1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mod := &ir.Module{
				Functions: []*ir.Function{
					{
						Name:       "test",
						ReturnType: types.Bool,
						Blocks: []*ir.BasicBlock{
							{
								Label: "entry",
								Instrs: []ir.Instruction{
									&ir.BinOp{
										Dest:  ir.NewReg(2),
										Op:    tt.op,
										Type:  types.Bool,
										Left:  ir.NewReg(0),
										Right: ir.NewReg(1),
									},
								},
								Term: &ir.Ret{Type: types.Bool, Val: ir.NewReg(2)},
							},
						},
					},
				},
			}

			gen := New()
			output := gen.Generate(mod)

			if !strings.Contains(output, tt.expected) {
				t.Errorf("output should contain %q, got:\n%s", tt.expected, output)
			}
		})
	}
}

func TestGenerate_UnaryOps(t *testing.T) {
	t.Run("neg", func(t *testing.T) {
		mod := &ir.Module{
			Functions: []*ir.Function{
				{
					Name:       "test",
					ReturnType: types.Int,
					Blocks: []*ir.BasicBlock{
						{
							Label: "entry",
							Instrs: []ir.Instruction{
								&ir.UnaryOp{
									Dest:    ir.NewReg(1),
									Op:      ir.OpNeg,
									Type:    types.Int,
									Operand: ir.NewReg(0),
								},
							},
							Term: &ir.Ret{Type: types.Int, Val: ir.NewReg(1)},
						},
					},
				},
			},
		}

		gen := New()
		output := gen.Generate(mod)

		// neg is implemented as: 0 - x
		if !strings.Contains(output, "%1 = sub i64 0, %0") {
			t.Errorf("neg should be implemented as sub 0, x, got:\n%s", output)
		}
	})

	t.Run("not", func(t *testing.T) {
		mod := &ir.Module{
			Functions: []*ir.Function{
				{
					Name:       "test",
					ReturnType: types.Bool,
					Blocks: []*ir.BasicBlock{
						{
							Label: "entry",
							Instrs: []ir.Instruction{
								&ir.UnaryOp{
									Dest:    ir.NewReg(1),
									Op:      ir.OpNot,
									Type:    types.Bool,
									Operand: ir.NewReg(0),
								},
							},
							Term: &ir.Ret{Type: types.Bool, Val: ir.NewReg(1)},
						},
					},
				},
			},
		}

		gen := New()
		output := gen.Generate(mod)

		// not is implemented as: xor x, true
		if !strings.Contains(output, "%1 = xor i1 %0, true") {
			t.Errorf("not should be implemented as xor x, true, got:\n%s", output)
		}
	})
}

func TestGenerate_Call(t *testing.T) {
	t.Run("call with return value", func(t *testing.T) {
		mod := &ir.Module{
			Functions: []*ir.Function{
				{
					Name:       "main",
					ReturnType: types.Int,
					Blocks: []*ir.BasicBlock{
						{
							Label: "entry",
							Instrs: []ir.Instruction{
								&ir.Call{
									Dest:    ir.NewReg(0),
									Func:    "add",
									Args:    []ir.Value{&ir.IntConst{Value: 1}, &ir.IntConst{Value: 2}},
									RetType: types.Int,
								},
							},
							Term: &ir.Ret{Type: types.Int, Val: ir.NewReg(0)},
						},
					},
				},
			},
		}

		gen := New()
		output := gen.Generate(mod)

		if !strings.Contains(output, "%0 = call i64 @add(i64 1, i64 2)") {
			t.Errorf("output should contain call instruction, got:\n%s", output)
		}
	})

	t.Run("void call", func(t *testing.T) {
		mod := &ir.Module{
			Functions: []*ir.Function{
				{
					Name:       "main",
					ReturnType: types.Void,
					Blocks: []*ir.BasicBlock{
						{
							Label: "entry",
							Instrs: []ir.Instruction{
								&ir.Call{
									Dest:    nil,
									Func:    "print",
									Args:    []ir.Value{&ir.IntConst{Value: 42}},
									RetType: types.Void,
								},
							},
							Term: &ir.RetVoid{},
						},
					},
				},
			},
		}

		gen := New()
		output := gen.Generate(mod)

		if !strings.Contains(output, "call void @print(i64 42)") {
			t.Errorf("output should contain void call, got:\n%s", output)
		}
	})
}

func TestGenerate_Copy(t *testing.T) {
	mod := &ir.Module{
		Functions: []*ir.Function{
			{
				Name:       "main",
				ReturnType: types.Int,
				Blocks: []*ir.BasicBlock{
					{
						Label: "entry",
						Instrs: []ir.Instruction{
							&ir.Copy{
								Dest: ir.NewReg(0),
								Type: types.Int,
								Val:  &ir.IntConst{Value: 42},
							},
						},
						Term: &ir.Ret{Type: types.Int, Val: ir.NewReg(0)},
					},
				},
			},
		},
	}

	gen := New()
	output := gen.Generate(mod)

	// Copy is implemented as: add x, 0 (for int) or just the value
	// Using add 42, 0 to assign to register
	if !strings.Contains(output, "%0 = add i64 42, 0") {
		t.Errorf("copy should be implemented as add x, 0, got:\n%s", output)
	}
}

func TestGenerate_Phi(t *testing.T) {
	mod := &ir.Module{
		Functions: []*ir.Function{
			{
				Name:       "test",
				ReturnType: types.Int,
				Blocks: []*ir.BasicBlock{
					{
						Label:  "entry",
						Instrs: []ir.Instruction{},
						Term:   &ir.CondBr{Cond: &ir.BoolConst{Value: true}, Then: "then", Else: "else"},
					},
					{
						Label:  "then",
						Instrs: []ir.Instruction{},
						Term:   &ir.Br{Target: "merge"},
					},
					{
						Label:  "else",
						Instrs: []ir.Instruction{},
						Term:   &ir.Br{Target: "merge"},
					},
					{
						Label: "merge",
						Instrs: []ir.Instruction{
							&ir.Phi{
								Dest: ir.NewReg(0),
								Type: types.Int,
								Entries: []ir.PhiEntry{
									{Val: &ir.IntConst{Value: 1}, Block: "then"},
									{Val: &ir.IntConst{Value: 2}, Block: "else"},
								},
							},
						},
						Term: &ir.Ret{Type: types.Int, Val: ir.NewReg(0)},
					},
				},
			},
		},
	}

	gen := New()
	output := gen.Generate(mod)

	if !strings.Contains(output, "%0 = phi i64 [ 1, %then ], [ 2, %else ]") {
		t.Errorf("output should contain phi instruction, got:\n%s", output)
	}
}

func TestGenerate_Branches(t *testing.T) {
	t.Run("unconditional branch", func(t *testing.T) {
		mod := &ir.Module{
			Functions: []*ir.Function{
				{
					Name:       "test",
					ReturnType: types.Void,
					Blocks: []*ir.BasicBlock{
						{
							Label:  "entry",
							Instrs: []ir.Instruction{},
							Term:   &ir.Br{Target: "next"},
						},
						{
							Label:  "next",
							Instrs: []ir.Instruction{},
							Term:   &ir.RetVoid{},
						},
					},
				},
			},
		}

		gen := New()
		output := gen.Generate(mod)

		if !strings.Contains(output, "br label %next") {
			t.Errorf("output should contain unconditional branch, got:\n%s", output)
		}
	})

	t.Run("conditional branch", func(t *testing.T) {
		mod := &ir.Module{
			Functions: []*ir.Function{
				{
					Name:       "test",
					ReturnType: types.Void,
					Blocks: []*ir.BasicBlock{
						{
							Label:  "entry",
							Instrs: []ir.Instruction{},
							Term:   &ir.CondBr{Cond: ir.NewReg(0), Then: "then", Else: "else"},
						},
						{
							Label:  "then",
							Instrs: []ir.Instruction{},
							Term:   &ir.RetVoid{},
						},
						{
							Label:  "else",
							Instrs: []ir.Instruction{},
							Term:   &ir.RetVoid{},
						},
					},
				},
			},
		}

		gen := New()
		output := gen.Generate(mod)

		if !strings.Contains(output, "br i1 %0, label %then, label %else") {
			t.Errorf("output should contain conditional branch, got:\n%s", output)
		}
	})
}

func TestGenerate_GlobalStrings(t *testing.T) {
	mod := &ir.Module{
		Globals: []*ir.Global{
			{Name: "str.0", Value: "hello"},
		},
		Functions: []*ir.Function{
			{
				Name:       "main",
				ReturnType: types.Void,
				Blocks: []*ir.BasicBlock{
					{
						Label:  "entry",
						Instrs: []ir.Instruction{},
						Term:   &ir.RetVoid{},
					},
				},
			},
		},
	}

	gen := New()
	output := gen.Generate(mod)

	// Global strings in LLVM IR format
	if !strings.Contains(output, "@str.0 = private constant") && !strings.Contains(output, "hello") {
		t.Errorf("output should contain global string declaration, got:\n%s", output)
	}
}

func TestGenerate_ReturnVoid(t *testing.T) {
	mod := &ir.Module{
		Functions: []*ir.Function{
			{
				Name:       "test",
				ReturnType: types.Void,
				Blocks: []*ir.BasicBlock{
					{
						Label:  "entry",
						Instrs: []ir.Instruction{},
						Term:   &ir.RetVoid{},
					},
				},
			},
		},
	}

	gen := New()
	output := gen.Generate(mod)

	if !strings.Contains(output, "define void @test()") {
		t.Errorf("output should contain void function, got:\n%s", output)
	}
	if !strings.Contains(output, "ret void") {
		t.Errorf("output should contain ret void, got:\n%s", output)
	}
}

func TestGenerate_BoolConstants(t *testing.T) {
	mod := &ir.Module{
		Functions: []*ir.Function{
			{
				Name:       "test",
				ReturnType: types.Bool,
				Blocks: []*ir.BasicBlock{
					{
						Label:  "entry",
						Instrs: []ir.Instruction{},
						Term:   &ir.Ret{Type: types.Bool, Val: &ir.BoolConst{Value: true}},
					},
				},
			},
		},
	}

	gen := New()
	output := gen.Generate(mod)

	// true in LLVM is 1 (i1)
	if !strings.Contains(output, "ret i1 true") && !strings.Contains(output, "ret i1 1") {
		t.Errorf("output should contain bool return, got:\n%s", output)
	}
}
