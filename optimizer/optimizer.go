package optimizer

import "f1c/ir"

// Pass is an optimization pass that transforms IR.
type Pass interface {
	Run(mod *ir.Module)
}

// Optimizer runs a sequence of optimization passes.
type Optimizer struct {
	passes []Pass
}

// New creates a new optimizer with default passes.
func New() *Optimizer {
	return &Optimizer{
		passes: []Pass{
			NewConstantFolding(),
			NewDCE(),
		},
	}
}

// Run executes all optimization passes on the module.
func (o *Optimizer) Run(mod *ir.Module) {
	for _, pass := range o.passes {
		pass.Run(mod)
	}
}

// AddPass adds an optimization pass.
func (o *Optimizer) AddPass(p Pass) {
	o.passes = append(o.passes, p)
}
