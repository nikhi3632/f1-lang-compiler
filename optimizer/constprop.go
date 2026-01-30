package optimizer

import (
	"f1c/ir"
)

// ConstantPropagation replaces uses of registers holding constants
// with the constant values themselves.
type ConstantPropagation struct{}

// NewConstantPropagation creates a new constant propagation pass.
func NewConstantPropagation() *ConstantPropagation {
	return &ConstantPropagation{}
}

// Run executes constant propagation on the module.
func (cp *ConstantPropagation) Run(mod *ir.Module) {
	for _, fn := range mod.Functions {
		cp.propagateFunction(fn)
	}
}

func (cp *ConstantPropagation) propagateFunction(fn *ir.Function) {
	// Map from register ID to constant value
	constants := make(map[int]ir.Value)

	for _, block := range fn.Blocks {
		// First pass: collect constants
		for _, instr := range block.Instrs {
			if c, ok := instr.(*ir.Copy); ok {
				// If we're copying a constant to a register, track it
				if isConstant(c.Val) {
					constants[c.Dest.ID] = c.Val
				}
			}
			if store, ok := instr.(*ir.Store); ok {
				// If storing a constant, we might want to track it
				// But for simplicity, we only track Copy for now
				_ = store
			}
		}

		// Second pass: substitute constants
		for _, instr := range block.Instrs {
			cp.substituteInInstruction(instr, constants)
		}

		// Also substitute in terminator
		if term := block.Term; term != nil {
			cp.substituteInTerminator(term, constants)
		}
	}
}

func (cp *ConstantPropagation) substituteInInstruction(instr ir.Instruction, constants map[int]ir.Value) {
	switch i := instr.(type) {
	case *ir.BinOp:
		if c := substituteValue(i.Left, constants); c != nil {
			i.Left = c
		}
		if c := substituteValue(i.Right, constants); c != nil {
			i.Right = c
		}
	case *ir.UnaryOp:
		if c := substituteValue(i.Operand, constants); c != nil {
			i.Operand = c
		}
	case *ir.Store:
		if c := substituteValue(i.Val, constants); c != nil {
			i.Val = c
		}
	case *ir.Call:
		for j, arg := range i.Args {
			if c := substituteValue(arg, constants); c != nil {
				i.Args[j] = c
			}
		}
	case *ir.CallIndirect:
		for j, arg := range i.Args {
			if c := substituteValue(arg, constants); c != nil {
				i.Args[j] = c
			}
		}
	case *ir.Copy:
		if c := substituteValue(i.Val, constants); c != nil {
			i.Val = c
		}
	case *ir.Phi:
		for j, entry := range i.Entries {
			if c := substituteValue(entry.Val, constants); c != nil {
				i.Entries[j].Val = c
			}
		}
	}
}

func (cp *ConstantPropagation) substituteInTerminator(term ir.Terminator, constants map[int]ir.Value) {
	switch t := term.(type) {
	case *ir.Ret:
		if c := substituteValue(t.Val, constants); c != nil {
			t.Val = c
		}
	case *ir.CondBr:
		if c := substituteValue(t.Cond, constants); c != nil {
			t.Cond = c
		}
	}
}

func substituteValue(v ir.Value, constants map[int]ir.Value) ir.Value {
	if reg, ok := v.(*ir.Reg); ok {
		if constVal, found := constants[reg.ID]; found {
			return constVal
		}
	}
	return nil
}

func isConstant(v ir.Value) bool {
	switch v.(type) {
	case *ir.IntConst, *ir.BoolConst:
		return true
	default:
		return false
	}
}
