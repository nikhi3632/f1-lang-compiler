package ir

import (
	"fmt"
	"strings"

	"f1c/types"
)

// Module is the top-level IR container.
type Module struct {
	Globals   []*Global
	Functions []*Function
}

func (m *Module) String() string {
	var sb strings.Builder

	for _, g := range m.Globals {
		sb.WriteString(g.String())
		sb.WriteString("\n")
	}

	if len(m.Globals) > 0 {
		sb.WriteString("\n")
	}

	for i, f := range m.Functions {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(f.String())
	}

	return sb.String()
}

// Global represents a global constant (e.g., string literals).
type Global struct {
	Name  string
	Value string
}

func (g *Global) String() string {
	return fmt.Sprintf("@%s = %q", g.Name, g.Value)
}

// Function represents an IR function.
type Function struct {
	Name       string
	Params     []*Param
	ReturnType types.Type
	Blocks     []*BasicBlock
}

func (f *Function) String() string {
	var sb strings.Builder

	// Function signature
	params := make([]string, len(f.Params))
	for i, p := range f.Params {
		params[i] = fmt.Sprintf("%s %%%s", p.Type.String(), p.Name)
	}

	sb.WriteString(fmt.Sprintf("define %s @%s(%s) {\n",
		f.ReturnType.String(), f.Name, strings.Join(params, ", ")))

	// Blocks
	for _, block := range f.Blocks {
		sb.WriteString(block.String())
	}

	sb.WriteString("}\n")
	return sb.String()
}

// Param represents a function parameter.
type Param struct {
	Name string
	Type types.Type
}

// BasicBlock is a sequence of instructions ending with a terminator.
type BasicBlock struct {
	Label  string
	Instrs []Instruction
	Term   Terminator
}

func (b *BasicBlock) String() string {
	var sb strings.Builder

	sb.WriteString(b.Label)
	sb.WriteString(":\n")

	for _, instr := range b.Instrs {
		sb.WriteString("    ")
		sb.WriteString(instr.String())
		sb.WriteString("\n")
	}

	sb.WriteString("    ")
	sb.WriteString(b.Term.String())
	sb.WriteString("\n")

	return sb.String()
}

// Value represents an IR value (constant, register, or parameter).
type Value interface {
	String() string
	valueNode()
}

// Reg is a virtual register.
type Reg struct {
	ID int
}

func NewReg(id int) *Reg {
	return &Reg{ID: id}
}

func (r *Reg) String() string  { return fmt.Sprintf("%%%d", r.ID) }
func (r *Reg) valueNode()      {}

// IntConst is an integer constant.
type IntConst struct {
	Value int64
}

func (c *IntConst) String() string { return fmt.Sprintf("%d", c.Value) }
func (c *IntConst) valueNode()     {}

// BoolConst is a boolean constant.
type BoolConst struct {
	Value bool
}

func (c *BoolConst) String() string {
	if c.Value {
		return "true"
	}
	return "false"
}
func (c *BoolConst) valueNode() {}

// StringConst references a global string.
type StringConst struct {
	Index int
	Name  string
}

func (c *StringConst) String() string { return fmt.Sprintf("@%s", c.Name) }
func (c *StringConst) valueNode()     {}

// ParamRef references a function parameter.
type ParamRef struct {
	Name string
}

func (p *ParamRef) String() string { return fmt.Sprintf("%%%s", p.Name) }
func (p *ParamRef) valueNode()     {}

// Instruction is an IR instruction.
type Instruction interface {
	String() string
	instrNode()
}

// Terminator is a block terminator.
type Terminator interface {
	String() string
	termNode()
}

// --- Instructions ---

// Alloca allocates stack space.
type Alloca struct {
	Dest *Reg
	Type types.Type
}

func (a *Alloca) String() string {
	return fmt.Sprintf("%s = alloca %s", a.Dest.String(), a.Type.String())
}
func (a *Alloca) instrNode() {}

// Load loads a value from memory.
type Load struct {
	Dest *Reg
	Type types.Type
	Ptr  Value
}

func (l *Load) String() string {
	return fmt.Sprintf("%s = load %s, %s", l.Dest.String(), l.Type.String(), l.Ptr.String())
}
func (l *Load) instrNode() {}

// Store stores a value to memory.
type Store struct {
	Type types.Type
	Val  Value
	Ptr  Value
}

func (s *Store) String() string {
	return fmt.Sprintf("store %s %s, %s", s.Type.String(), s.Val.String(), s.Ptr.String())
}
func (s *Store) instrNode() {}

// BinOp is a binary operation.
type BinOp struct {
	Dest  *Reg
	Op    Op
	Type  types.Type
	Left  Value
	Right Value
}

func (b *BinOp) String() string {
	return fmt.Sprintf("%s = %s %s %s, %s",
		b.Dest.String(), b.Op.String(), b.Type.String(),
		b.Left.String(), b.Right.String())
}
func (b *BinOp) instrNode() {}

// UnaryOp is a unary operation.
type UnaryOp struct {
	Dest    *Reg
	Op      Op
	Type    types.Type
	Operand Value
}

func (u *UnaryOp) String() string {
	return fmt.Sprintf("%s = %s %s %s",
		u.Dest.String(), u.Op.String(), u.Type.String(), u.Operand.String())
}
func (u *UnaryOp) instrNode() {}

// Call calls a function.
type Call struct {
	Dest    *Reg // nil for void calls
	Func    string
	Args    []Value
	RetType types.Type
}

func (c *Call) String() string {
	args := make([]string, len(c.Args))
	for i, arg := range c.Args {
		args[i] = arg.String()
	}

	if c.Dest == nil || types.Equal(c.RetType, types.Void) {
		return fmt.Sprintf("call void @%s(%s)", c.Func, strings.Join(args, ", "))
	}
	return fmt.Sprintf("%s = call %s @%s(%s)",
		c.Dest.String(), c.RetType.String(), c.Func, strings.Join(args, ", "))
}
func (c *Call) instrNode() {}

// Copy copies a value (for SSA).
type Copy struct {
	Dest *Reg
	Type types.Type
	Val  Value
}

func (c *Copy) String() string {
	return fmt.Sprintf("%s = copy %s %s", c.Dest.String(), c.Type.String(), c.Val.String())
}
func (c *Copy) instrNode() {}

// Phi is an SSA phi node.
type Phi struct {
	Dest    *Reg
	Type    types.Type
	Entries []PhiEntry
}

type PhiEntry struct {
	Val   Value
	Block string
}

func (p *Phi) String() string {
	entries := make([]string, len(p.Entries))
	for i, e := range p.Entries {
		entries[i] = fmt.Sprintf("[%s, %s]", e.Val.String(), e.Block)
	}
	return fmt.Sprintf("%s = phi %s %s",
		p.Dest.String(), p.Type.String(), strings.Join(entries, ", "))
}
func (p *Phi) instrNode() {}

// --- Terminators ---

// RetVoid returns void.
type RetVoid struct{}

func (r *RetVoid) String() string { return "ret void" }
func (r *RetVoid) termNode()      {}

// Ret returns a value.
type Ret struct {
	Type types.Type
	Val  Value
}

func (r *Ret) String() string {
	return fmt.Sprintf("ret %s %s", r.Type.String(), r.Val.String())
}
func (r *Ret) termNode() {}

// Br is an unconditional branch.
type Br struct {
	Target string
}

func (b *Br) String() string { return fmt.Sprintf("br %s", b.Target) }
func (b *Br) termNode()      {}

// CondBr is a conditional branch.
type CondBr struct {
	Cond Value
	Then string
	Else string
}

func (c *CondBr) String() string {
	return fmt.Sprintf("br %s, %s, %s", c.Cond.String(), c.Then, c.Else)
}
func (c *CondBr) termNode() {}

// --- Operators ---

type Op int

const (
	OpAdd Op = iota
	OpSub
	OpMul
	OpDiv
	OpMod
	OpEq
	OpNeq
	OpLt
	OpGt
	OpLte
	OpGte
	OpAnd
	OpOr
	OpNeg
	OpNot
)

func (o Op) String() string {
	switch o {
	case OpAdd:
		return "add"
	case OpSub:
		return "sub"
	case OpMul:
		return "mul"
	case OpDiv:
		return "div"
	case OpMod:
		return "mod"
	case OpEq:
		return "eq"
	case OpNeq:
		return "neq"
	case OpLt:
		return "lt"
	case OpGt:
		return "gt"
	case OpLte:
		return "lte"
	case OpGte:
		return "gte"
	case OpAnd:
		return "and"
	case OpOr:
		return "or"
	case OpNeg:
		return "neg"
	case OpNot:
		return "not"
	default:
		return "unknown"
	}
}
