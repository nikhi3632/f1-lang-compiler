package optimizer

import (
	"testing"

	"f1c/ir"
	"f1c/types"
)

// --- Constant Folding Tests ---

func TestConstantFolding_AddConstants(t *testing.T) {
	// %0 = add int 3, 5  →  %0 = copy int 8
	block := &ir.BasicBlock{
		Label: "entry",
		Instrs: []ir.Instruction{
			&ir.BinOp{
				Dest:  ir.NewReg(0),
				Op:    ir.OpAdd,
				Type:  types.Int,
				Left:  &ir.IntConst{Value: 3},
				Right: &ir.IntConst{Value: 5},
			},
		},
		Term: &ir.Ret{Type: types.Int, Val: ir.NewReg(0)},
	}

	fn := &ir.Function{
		Name:       "test",
		ReturnType: types.Int,
		Blocks:     []*ir.BasicBlock{block},
	}

	mod := &ir.Module{Functions: []*ir.Function{fn}}
	opt := NewConstantFolding()
	opt.Run(mod)

	// Check that the add was replaced with a copy of 8
	instr := fn.Blocks[0].Instrs[0]
	if copy, ok := instr.(*ir.Copy); ok {
		if c, ok := copy.Val.(*ir.IntConst); ok {
			if c.Value != 8 {
				t.Errorf("expected constant 8, got %d", c.Value)
			}
		} else {
			t.Errorf("expected IntConst, got %T", copy.Val)
		}
	} else {
		t.Errorf("expected Copy instruction, got %T", instr)
	}
}

func TestConstantFolding_SubConstants(t *testing.T) {
	block := &ir.BasicBlock{
		Label: "entry",
		Instrs: []ir.Instruction{
			&ir.BinOp{
				Dest:  ir.NewReg(0),
				Op:    ir.OpSub,
				Type:  types.Int,
				Left:  &ir.IntConst{Value: 10},
				Right: &ir.IntConst{Value: 3},
			},
		},
		Term: &ir.Ret{Type: types.Int, Val: ir.NewReg(0)},
	}

	fn := &ir.Function{
		Name:       "test",
		ReturnType: types.Int,
		Blocks:     []*ir.BasicBlock{block},
	}

	mod := &ir.Module{Functions: []*ir.Function{fn}}
	opt := NewConstantFolding()
	opt.Run(mod)

	instr := fn.Blocks[0].Instrs[0]
	if copy, ok := instr.(*ir.Copy); ok {
		if c, ok := copy.Val.(*ir.IntConst); ok {
			if c.Value != 7 {
				t.Errorf("expected constant 7, got %d", c.Value)
			}
		}
	} else {
		t.Errorf("expected Copy instruction, got %T", instr)
	}
}

func TestConstantFolding_MulConstants(t *testing.T) {
	block := &ir.BasicBlock{
		Label: "entry",
		Instrs: []ir.Instruction{
			&ir.BinOp{
				Dest:  ir.NewReg(0),
				Op:    ir.OpMul,
				Type:  types.Int,
				Left:  &ir.IntConst{Value: 4},
				Right: &ir.IntConst{Value: 5},
			},
		},
		Term: &ir.Ret{Type: types.Int, Val: ir.NewReg(0)},
	}

	fn := &ir.Function{
		Name:       "test",
		ReturnType: types.Int,
		Blocks:     []*ir.BasicBlock{block},
	}

	mod := &ir.Module{Functions: []*ir.Function{fn}}
	opt := NewConstantFolding()
	opt.Run(mod)

	instr := fn.Blocks[0].Instrs[0]
	if copy, ok := instr.(*ir.Copy); ok {
		if c, ok := copy.Val.(*ir.IntConst); ok {
			if c.Value != 20 {
				t.Errorf("expected constant 20, got %d", c.Value)
			}
		}
	} else {
		t.Errorf("expected Copy instruction, got %T", instr)
	}
}

func TestConstantFolding_AddZero(t *testing.T) {
	// x + 0 → x
	block := &ir.BasicBlock{
		Label: "entry",
		Instrs: []ir.Instruction{
			&ir.BinOp{
				Dest:  ir.NewReg(1),
				Op:    ir.OpAdd,
				Type:  types.Int,
				Left:  ir.NewReg(0),
				Right: &ir.IntConst{Value: 0},
			},
		},
		Term: &ir.Ret{Type: types.Int, Val: ir.NewReg(1)},
	}

	fn := &ir.Function{
		Name:       "test",
		ReturnType: types.Int,
		Blocks:     []*ir.BasicBlock{block},
	}

	mod := &ir.Module{Functions: []*ir.Function{fn}}
	opt := NewConstantFolding()
	opt.Run(mod)

	instr := fn.Blocks[0].Instrs[0]
	if copy, ok := instr.(*ir.Copy); ok {
		if reg, ok := copy.Val.(*ir.Reg); ok {
			if reg.ID != 0 {
				t.Errorf("expected copy of %%0, got %%%d", reg.ID)
			}
		} else {
			t.Errorf("expected Reg, got %T", copy.Val)
		}
	} else {
		t.Errorf("expected Copy instruction, got %T", instr)
	}
}

func TestConstantFolding_MulOne(t *testing.T) {
	// x * 1 → x
	block := &ir.BasicBlock{
		Label: "entry",
		Instrs: []ir.Instruction{
			&ir.BinOp{
				Dest:  ir.NewReg(1),
				Op:    ir.OpMul,
				Type:  types.Int,
				Left:  ir.NewReg(0),
				Right: &ir.IntConst{Value: 1},
			},
		},
		Term: &ir.Ret{Type: types.Int, Val: ir.NewReg(1)},
	}

	fn := &ir.Function{
		Name:       "test",
		ReturnType: types.Int,
		Blocks:     []*ir.BasicBlock{block},
	}

	mod := &ir.Module{Functions: []*ir.Function{fn}}
	opt := NewConstantFolding()
	opt.Run(mod)

	instr := fn.Blocks[0].Instrs[0]
	if copy, ok := instr.(*ir.Copy); ok {
		if reg, ok := copy.Val.(*ir.Reg); ok {
			if reg.ID != 0 {
				t.Errorf("expected copy of %%0, got %%%d", reg.ID)
			}
		}
	} else {
		t.Errorf("expected Copy instruction, got %T", instr)
	}
}

func TestConstantFolding_MulZero(t *testing.T) {
	// x * 0 → 0
	block := &ir.BasicBlock{
		Label: "entry",
		Instrs: []ir.Instruction{
			&ir.BinOp{
				Dest:  ir.NewReg(1),
				Op:    ir.OpMul,
				Type:  types.Int,
				Left:  ir.NewReg(0),
				Right: &ir.IntConst{Value: 0},
			},
		},
		Term: &ir.Ret{Type: types.Int, Val: ir.NewReg(1)},
	}

	fn := &ir.Function{
		Name:       "test",
		ReturnType: types.Int,
		Blocks:     []*ir.BasicBlock{block},
	}

	mod := &ir.Module{Functions: []*ir.Function{fn}}
	opt := NewConstantFolding()
	opt.Run(mod)

	instr := fn.Blocks[0].Instrs[0]
	if copy, ok := instr.(*ir.Copy); ok {
		if c, ok := copy.Val.(*ir.IntConst); ok {
			if c.Value != 0 {
				t.Errorf("expected constant 0, got %d", c.Value)
			}
		}
	} else {
		t.Errorf("expected Copy instruction, got %T", instr)
	}
}

func TestConstantFolding_Comparison(t *testing.T) {
	// 5 < 10 → true
	block := &ir.BasicBlock{
		Label: "entry",
		Instrs: []ir.Instruction{
			&ir.BinOp{
				Dest:  ir.NewReg(0),
				Op:    ir.OpLt,
				Type:  types.Int,
				Left:  &ir.IntConst{Value: 5},
				Right: &ir.IntConst{Value: 10},
			},
		},
		Term: &ir.Ret{Type: types.Bool, Val: ir.NewReg(0)},
	}

	fn := &ir.Function{
		Name:       "test",
		ReturnType: types.Bool,
		Blocks:     []*ir.BasicBlock{block},
	}

	mod := &ir.Module{Functions: []*ir.Function{fn}}
	opt := NewConstantFolding()
	opt.Run(mod)

	instr := fn.Blocks[0].Instrs[0]
	if copy, ok := instr.(*ir.Copy); ok {
		if b, ok := copy.Val.(*ir.BoolConst); ok {
			if !b.Value {
				t.Errorf("expected true, got false")
			}
		}
	} else {
		t.Errorf("expected Copy instruction, got %T", instr)
	}
}

// --- Constant Propagation Tests ---

func TestConstantPropagation_PropagatesConstants(t *testing.T) {
	// %0 = copy int 5
	// %1 = add int %0, 3
	// After propagation: %1 = add int 5, 3
	block := &ir.BasicBlock{
		Label: "entry",
		Instrs: []ir.Instruction{
			&ir.Copy{
				Dest: ir.NewReg(0),
				Type: types.Int,
				Val:  &ir.IntConst{Value: 5},
			},
			&ir.BinOp{
				Dest:  ir.NewReg(1),
				Op:    ir.OpAdd,
				Type:  types.Int,
				Left:  ir.NewReg(0),
				Right: &ir.IntConst{Value: 3},
			},
		},
		Term: &ir.Ret{Type: types.Int, Val: ir.NewReg(1)},
	}

	fn := &ir.Function{
		Name:       "test",
		ReturnType: types.Int,
		Blocks:     []*ir.BasicBlock{block},
	}

	mod := &ir.Module{Functions: []*ir.Function{fn}}
	opt := NewConstantPropagation()
	opt.Run(mod)

	// The add should now have 5 as left operand
	binop := fn.Blocks[0].Instrs[1].(*ir.BinOp)
	if c, ok := binop.Left.(*ir.IntConst); ok {
		if c.Value != 5 {
			t.Errorf("expected 5, got %d", c.Value)
		}
	} else {
		t.Errorf("expected IntConst, got %T", binop.Left)
	}
}

func TestConstantPropagation_PropagatesInReturn(t *testing.T) {
	// %0 = copy int 42
	// ret int %0
	// After propagation: ret int 42
	block := &ir.BasicBlock{
		Label: "entry",
		Instrs: []ir.Instruction{
			&ir.Copy{
				Dest: ir.NewReg(0),
				Type: types.Int,
				Val:  &ir.IntConst{Value: 42},
			},
		},
		Term: &ir.Ret{Type: types.Int, Val: ir.NewReg(0)},
	}

	fn := &ir.Function{
		Name:       "test",
		ReturnType: types.Int,
		Blocks:     []*ir.BasicBlock{block},
	}

	mod := &ir.Module{Functions: []*ir.Function{fn}}
	opt := NewConstantPropagation()
	opt.Run(mod)

	// The return should now have 42 as value
	ret := fn.Blocks[0].Term.(*ir.Ret)
	if c, ok := ret.Val.(*ir.IntConst); ok {
		if c.Value != 42 {
			t.Errorf("expected 42, got %d", c.Value)
		}
	} else {
		t.Errorf("expected IntConst, got %T", ret.Val)
	}
}

// --- Dead Code Elimination Tests ---

func TestDCE_UnusedInstruction(t *testing.T) {
	block := &ir.BasicBlock{
		Label: "entry",
		Instrs: []ir.Instruction{
			&ir.BinOp{
				Dest:  ir.NewReg(0),
				Op:    ir.OpAdd,
				Type:  types.Int,
				Left:  &ir.IntConst{Value: 1},
				Right: &ir.IntConst{Value: 2},
			}, // unused
			&ir.BinOp{
				Dest:  ir.NewReg(1),
				Op:    ir.OpMul,
				Type:  types.Int,
				Left:  &ir.IntConst{Value: 3},
				Right: &ir.IntConst{Value: 4},
			}, // used in return
		},
		Term: &ir.Ret{Type: types.Int, Val: ir.NewReg(1)},
	}

	fn := &ir.Function{
		Name:       "test",
		ReturnType: types.Int,
		Blocks:     []*ir.BasicBlock{block},
	}

	mod := &ir.Module{Functions: []*ir.Function{fn}}
	opt := NewDCE()
	opt.Run(mod)

	// Should only have 1 instruction left (the mul)
	if len(fn.Blocks[0].Instrs) != 1 {
		t.Errorf("expected 1 instruction, got %d", len(fn.Blocks[0].Instrs))
	}

	// The remaining instruction should be the mul
	if binop, ok := fn.Blocks[0].Instrs[0].(*ir.BinOp); ok {
		if binop.Op != ir.OpMul {
			t.Errorf("expected mul, got %v", binop.Op)
		}
	}
}

func TestDCE_KeepsUsedInstructions(t *testing.T) {
	block := &ir.BasicBlock{
		Label: "entry",
		Instrs: []ir.Instruction{
			&ir.BinOp{
				Dest:  ir.NewReg(0),
				Op:    ir.OpAdd,
				Type:  types.Int,
				Left:  &ir.IntConst{Value: 1},
				Right: &ir.IntConst{Value: 2},
			},
			&ir.BinOp{
				Dest:  ir.NewReg(1),
				Op:    ir.OpMul,
				Type:  types.Int,
				Left:  ir.NewReg(0), // uses %0
				Right: &ir.IntConst{Value: 3},
			},
		},
		Term: &ir.Ret{Type: types.Int, Val: ir.NewReg(1)},
	}

	fn := &ir.Function{
		Name:       "test",
		ReturnType: types.Int,
		Blocks:     []*ir.BasicBlock{block},
	}

	mod := &ir.Module{Functions: []*ir.Function{fn}}
	opt := NewDCE()
	opt.Run(mod)

	// Both instructions should remain (chained dependency)
	if len(fn.Blocks[0].Instrs) != 2 {
		t.Errorf("expected 2 instructions, got %d", len(fn.Blocks[0].Instrs))
	}
}

// --- Tail Call Optimization Tests ---

func TestTCO_MarksTailCall(t *testing.T) {
	// %0 = call int @fib(...)
	// ret int %0
	// Should mark the call as a tail call
	block := &ir.BasicBlock{
		Label: "entry",
		Instrs: []ir.Instruction{
			&ir.Call{
				Dest:    ir.NewReg(0),
				Func:    "fib",
				Args:    []ir.Value{&ir.IntConst{Value: 5}},
				RetType: types.Int,
			},
		},
		Term: &ir.Ret{Type: types.Int, Val: ir.NewReg(0)},
	}

	fn := &ir.Function{
		Name:       "fib",
		ReturnType: types.Int,
		Params:     []*ir.Param{{Name: "n", Type: types.Int}},
		Blocks:     []*ir.BasicBlock{block},
	}

	mod := &ir.Module{Functions: []*ir.Function{fn}}
	opt := NewTailCallOptimization()
	opt.Run(mod)

	// The call should now be marked as a tail call
	call := fn.Blocks[0].Instrs[0].(*ir.Call)
	if !call.IsTail {
		t.Error("expected call to be marked as tail call")
	}
}

func TestTCO_DoesNotMarkNonTailCall(t *testing.T) {
	// %0 = call int @foo(...)
	// %1 = add int %0, 1  <-- Something happens after the call
	// ret int %1
	// Should NOT mark the call as a tail call
	block := &ir.BasicBlock{
		Label: "entry",
		Instrs: []ir.Instruction{
			&ir.Call{
				Dest:    ir.NewReg(0),
				Func:    "foo",
				Args:    []ir.Value{&ir.IntConst{Value: 5}},
				RetType: types.Int,
			},
			&ir.BinOp{
				Dest:  ir.NewReg(1),
				Op:    ir.OpAdd,
				Type:  types.Int,
				Left:  ir.NewReg(0),
				Right: &ir.IntConst{Value: 1},
			},
		},
		Term: &ir.Ret{Type: types.Int, Val: ir.NewReg(1)},
	}

	fn := &ir.Function{
		Name:       "test",
		ReturnType: types.Int,
		Blocks:     []*ir.BasicBlock{block},
	}

	mod := &ir.Module{Functions: []*ir.Function{fn}}
	opt := NewTailCallOptimization()
	opt.Run(mod)

	// The call should NOT be marked as a tail call
	call := fn.Blocks[0].Instrs[0].(*ir.Call)
	if call.IsTail {
		t.Error("expected call NOT to be marked as tail call")
	}
}

// --- Optimizer Pipeline Tests ---

func TestOptimizer_RunAll(t *testing.T) {
	// Test that multiple passes can be chained
	block := &ir.BasicBlock{
		Label: "entry",
		Instrs: []ir.Instruction{
			&ir.BinOp{
				Dest:  ir.NewReg(0),
				Op:    ir.OpAdd,
				Type:  types.Int,
				Left:  &ir.IntConst{Value: 3},
				Right: &ir.IntConst{Value: 5},
			}, // folds to 8, then DCE keeps it because it's used
		},
		Term: &ir.Ret{Type: types.Int, Val: ir.NewReg(0)},
	}

	fn := &ir.Function{
		Name:       "test",
		ReturnType: types.Int,
		Blocks:     []*ir.BasicBlock{block},
	}

	mod := &ir.Module{Functions: []*ir.Function{fn}}

	opt := New()
	opt.Run(mod)

	// After constant folding, should be copy of 8
	instr := fn.Blocks[0].Instrs[0]
	if copy, ok := instr.(*ir.Copy); ok {
		if c, ok := copy.Val.(*ir.IntConst); ok {
			if c.Value != 8 {
				t.Errorf("expected 8, got %d", c.Value)
			}
		}
	}
}
