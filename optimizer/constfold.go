package optimizer

import (
	"f1c/ir"
	"f1c/types"
)

// ConstantFolding evaluates constant expressions at compile time.
type ConstantFolding struct{}

// NewConstantFolding creates a new constant folding pass.
func NewConstantFolding() *ConstantFolding {
	return &ConstantFolding{}
}

// Run executes constant folding on the module.
func (cf *ConstantFolding) Run(mod *ir.Module) {
	for _, fn := range mod.Functions {
		for _, block := range fn.Blocks {
			cf.foldBlock(block)
		}
	}
}

func (cf *ConstantFolding) foldBlock(block *ir.BasicBlock) {
	for i, instr := range block.Instrs {
		if binop, ok := instr.(*ir.BinOp); ok {
			if folded := cf.foldBinOp(binop); folded != nil {
				block.Instrs[i] = folded
			}
		}
	}
}

func (cf *ConstantFolding) foldBinOp(binop *ir.BinOp) ir.Instruction {
	leftConst, leftIsConst := binop.Left.(*ir.IntConst)
	rightConst, rightIsConst := binop.Right.(*ir.IntConst)

	// Both operands are constants - fold completely
	if leftIsConst && rightIsConst {
		return cf.foldConstantOp(binop.Dest, binop.Op, binop.Type, leftConst.Value, rightConst.Value)
	}

	// Algebraic simplifications
	if rightIsConst {
		// x + 0 → x
		if binop.Op == ir.OpAdd && rightConst.Value == 0 {
			return &ir.Copy{Dest: binop.Dest, Type: binop.Type, Val: binop.Left}
		}
		// x - 0 → x
		if binop.Op == ir.OpSub && rightConst.Value == 0 {
			return &ir.Copy{Dest: binop.Dest, Type: binop.Type, Val: binop.Left}
		}
		// x * 1 → x
		if binop.Op == ir.OpMul && rightConst.Value == 1 {
			return &ir.Copy{Dest: binop.Dest, Type: binop.Type, Val: binop.Left}
		}
		// x * 0 → 0
		if binop.Op == ir.OpMul && rightConst.Value == 0 {
			return &ir.Copy{Dest: binop.Dest, Type: binop.Type, Val: &ir.IntConst{Value: 0}}
		}
		// x / 1 → x
		if binop.Op == ir.OpDiv && rightConst.Value == 1 {
			return &ir.Copy{Dest: binop.Dest, Type: binop.Type, Val: binop.Left}
		}
	}

	if leftIsConst {
		// 0 + x → x
		if binop.Op == ir.OpAdd && leftConst.Value == 0 {
			return &ir.Copy{Dest: binop.Dest, Type: binop.Type, Val: binop.Right}
		}
		// 0 * x → 0
		if binop.Op == ir.OpMul && leftConst.Value == 0 {
			return &ir.Copy{Dest: binop.Dest, Type: binop.Type, Val: &ir.IntConst{Value: 0}}
		}
		// 1 * x → x
		if binop.Op == ir.OpMul && leftConst.Value == 1 {
			return &ir.Copy{Dest: binop.Dest, Type: binop.Type, Val: binop.Right}
		}
	}

	// x - x → 0 (same register)
	if leftReg, lok := binop.Left.(*ir.Reg); lok {
		if rightReg, rok := binop.Right.(*ir.Reg); rok {
			if binop.Op == ir.OpSub && leftReg.ID == rightReg.ID {
				return &ir.Copy{Dest: binop.Dest, Type: binop.Type, Val: &ir.IntConst{Value: 0}}
			}
		}
	}

	return nil // No folding possible
}

func (cf *ConstantFolding) foldConstantOp(dest *ir.Reg, op ir.Op, typ types.Type, left, right int64) ir.Instruction {
	var result ir.Value

	switch op {
	case ir.OpAdd:
		result = &ir.IntConst{Value: left + right}
	case ir.OpSub:
		result = &ir.IntConst{Value: left - right}
	case ir.OpMul:
		result = &ir.IntConst{Value: left * right}
	case ir.OpDiv:
		if right != 0 {
			result = &ir.IntConst{Value: left / right}
		} else {
			return nil // Don't fold division by zero
		}
	case ir.OpMod:
		if right != 0 {
			result = &ir.IntConst{Value: left % right}
		} else {
			return nil
		}
	case ir.OpLt:
		result = &ir.BoolConst{Value: left < right}
	case ir.OpGt:
		result = &ir.BoolConst{Value: left > right}
	case ir.OpLte:
		result = &ir.BoolConst{Value: left <= right}
	case ir.OpGte:
		result = &ir.BoolConst{Value: left >= right}
	case ir.OpEq:
		result = &ir.BoolConst{Value: left == right}
	case ir.OpNeq:
		result = &ir.BoolConst{Value: left != right}
	default:
		return nil
	}

	// Determine the correct type for the copy
	copyType := typ
	if _, isBool := result.(*ir.BoolConst); isBool {
		copyType = types.Bool
	}

	return &ir.Copy{Dest: dest, Type: copyType, Val: result}
}
