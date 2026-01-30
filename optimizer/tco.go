package optimizer

import (
	"f1c/ir"
)

// TailCallOptimization marks tail calls for optimization.
// A tail call is a call that is immediately followed by a return
// of the call's result.
type TailCallOptimization struct{}

// NewTailCallOptimization creates a new TCO pass.
func NewTailCallOptimization() *TailCallOptimization {
	return &TailCallOptimization{}
}

// Run executes tail call optimization on the module.
func (tco *TailCallOptimization) Run(mod *ir.Module) {
	for _, fn := range mod.Functions {
		tco.optimizeFunction(fn)
	}
}

func (tco *TailCallOptimization) optimizeFunction(fn *ir.Function) {
	for _, block := range fn.Blocks {
		tco.optimizeBlock(fn, block)
	}
}

func (tco *TailCallOptimization) optimizeBlock(fn *ir.Function, block *ir.BasicBlock) {
	// Look for the pattern:
	// %n = call T @func(...)
	// ret T %n
	//
	// Transform to:
	// %n = tail call T @func(...)
	// ret T %n
	//
	// For self-recursive calls (calls to the same function), we mark them as tail calls.

	if len(block.Instrs) == 0 {
		return
	}

	// Get the last instruction
	lastInstr := block.Instrs[len(block.Instrs)-1]
	call, isCall := lastInstr.(*ir.Call)
	if !isCall {
		return
	}

	// Check if the terminator is a return of the call's result
	ret, isRet := block.Term.(*ir.Ret)
	if !isRet {
		return
	}

	// Check if the return value matches the call destination
	if call.Dest == nil {
		return
	}
	retReg, isReg := ret.Val.(*ir.Reg)
	if !isReg {
		return
	}
	if retReg.ID != call.Dest.ID {
		return
	}

	// This is a tail call! Mark it.
	// For self-recursive calls, this is especially important.
	call.IsTail = true
}

// TailCall is an extension to the Call instruction to support tail calls.
// We add an IsTail field to the ir.Call struct to track this.
// Note: This requires updating the ir.Call struct.
