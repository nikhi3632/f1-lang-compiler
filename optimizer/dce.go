package optimizer

import "f1c/ir"

// DCE performs Dead Code Elimination.
type DCE struct{}

// NewDCE creates a new dead code elimination pass.
func NewDCE() *DCE {
	return &DCE{}
}

// Run executes dead code elimination on the module.
func (d *DCE) Run(mod *ir.Module) {
	for _, fn := range mod.Functions {
		d.eliminateDeadCode(fn)
	}
}

func (d *DCE) eliminateDeadCode(fn *ir.Function) {
	// Build a set of used values
	used := make(map[int]bool)

	// Mark values used in terminators
	for _, block := range fn.Blocks {
		d.markUsedInTerminator(block.Term, used)
	}

	// Mark side-effecting instructions' operands as used FIRST
	for _, block := range fn.Blocks {
		for _, instr := range block.Instrs {
			if d.hasSideEffects(instr) {
				d.markUsedInInstr(instr, used)
			}
		}
	}

	// Propagate used values backwards through the instruction chain
	changed := true
	for changed {
		changed = false
		for _, block := range fn.Blocks {
			for _, instr := range block.Instrs {
				dest := d.getInstrDest(instr)
				if dest != nil && used[dest.ID] {
					// This instruction's result is used, mark its operands as used
					if d.markUsedInInstr(instr, used) {
						changed = true
					}
				}
			}
		}
	}

	// Remove dead instructions
	for _, block := range fn.Blocks {
		newInstrs := make([]ir.Instruction, 0, len(block.Instrs))
		for _, instr := range block.Instrs {
			dest := d.getInstrDest(instr)
			// Keep instruction if it has no dest (side effects), or dest is used
			if dest == nil || used[dest.ID] || d.hasSideEffects(instr) {
				newInstrs = append(newInstrs, instr)
			}
		}
		block.Instrs = newInstrs
	}
}

func (d *DCE) getInstrDest(instr ir.Instruction) *ir.Reg {
	switch i := instr.(type) {
	case *ir.BinOp:
		return i.Dest
	case *ir.UnaryOp:
		return i.Dest
	case *ir.Load:
		return i.Dest
	case *ir.Alloca:
		return i.Dest
	case *ir.Call:
		return i.Dest
	case *ir.CallIndirect:
		return i.Dest
	case *ir.ClosureCall:
		return i.Dest
	case *ir.Copy:
		return i.Dest
	case *ir.Phi:
		return i.Dest
	case *ir.MakeClosure:
		return i.Dest
	case *ir.GetEnvField:
		return i.Dest
	default:
		return nil
	}
}

func (d *DCE) markUsedInTerminator(term ir.Terminator, used map[int]bool) {
	switch t := term.(type) {
	case *ir.Ret:
		d.markUsedValue(t.Val, used)
	case *ir.CondBr:
		d.markUsedValue(t.Cond, used)
	}
}

func (d *DCE) markUsedInInstr(instr ir.Instruction, used map[int]bool) bool {
	changed := false

	switch i := instr.(type) {
	case *ir.BinOp:
		changed = d.markUsedValue(i.Left, used) || changed
		changed = d.markUsedValue(i.Right, used) || changed
	case *ir.UnaryOp:
		changed = d.markUsedValue(i.Operand, used) || changed
	case *ir.Load:
		changed = d.markUsedValue(i.Ptr, used) || changed
	case *ir.Store:
		changed = d.markUsedValue(i.Val, used) || changed
		changed = d.markUsedValue(i.Ptr, used) || changed
	case *ir.Call:
		for _, arg := range i.Args {
			changed = d.markUsedValue(arg, used) || changed
		}
	case *ir.CallIndirect:
		// Mark the function pointer and all arguments as used
		changed = d.markUsedValue(i.FuncPtr, used) || changed
		for _, arg := range i.Args {
			changed = d.markUsedValue(arg, used) || changed
		}
	case *ir.ClosureCall:
		// Mark the closure pointer and all arguments as used
		changed = d.markUsedValue(i.ClosurePtr, used) || changed
		for _, arg := range i.Args {
			changed = d.markUsedValue(arg, used) || changed
		}
	case *ir.Copy:
		changed = d.markUsedValue(i.Val, used) || changed
	case *ir.Phi:
		for _, entry := range i.Entries {
			changed = d.markUsedValue(entry.Val, used) || changed
		}
	case *ir.MakeClosure:
		// Mark captured values as used
		for _, cap := range i.Captures {
			changed = d.markUsedValue(cap, used) || changed
		}
	case *ir.GetEnvField:
		// Mark environment pointer as used
		changed = d.markUsedValue(i.Env, used) || changed
	}

	return changed
}

func (d *DCE) markUsedValue(val ir.Value, used map[int]bool) bool {
	if reg, ok := val.(*ir.Reg); ok {
		if !used[reg.ID] {
			used[reg.ID] = true
			return true
		}
	}
	return false
}

func (d *DCE) hasSideEffects(instr ir.Instruction) bool {
	switch i := instr.(type) {
	case *ir.Store:
		return true
	case *ir.Call:
		// All calls have potential side effects
		return true
	case *ir.CallIndirect:
		// Indirect calls also have potential side effects
		return true
	case *ir.ClosureCall:
		// Closure calls also have potential side effects
		return true
	case *ir.Alloca:
		// Allocas are needed for stores
		// Check if this alloca is used as a store target
		_ = i
		return false // We'll keep allocas if they're used
	default:
		return false
	}
}
