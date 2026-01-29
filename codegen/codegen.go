package codegen

import (
	"fmt"
	"strings"

	"f1c/ir"
	"f1c/types"
)

// Generator produces LLVM IR from F1-IR.
type Generator struct {
	sb strings.Builder
}

// New creates a new LLVM IR generator.
func New() *Generator {
	return &Generator{}
}

// Generate produces LLVM IR text from an F1-IR module.
func (g *Generator) Generate(mod *ir.Module) string {
	g.sb.Reset()

	// Header with target triple
	g.emitHeader()

	// External declarations (runtime functions)
	g.emitExternals()

	// Global strings
	for _, global := range mod.Globals {
		g.emitGlobal(global)
	}

	if len(mod.Globals) > 0 {
		g.sb.WriteString("\n")
	}

	// Functions
	for i, fn := range mod.Functions {
		if i > 0 {
			g.sb.WriteString("\n")
		}
		g.emitFunction(fn)
	}

	return g.sb.String()
}

func (g *Generator) emitHeader() {
	// Use a generic target triple that works on most systems
	g.sb.WriteString("; F1-Lang LLVM IR\n")
	g.sb.WriteString("target triple = \"x86_64-unknown-linux-gnu\"\n")
	g.sb.WriteString("\n")
}

func (g *Generator) emitExternals() {
	// Declare external functions that might be used
	g.sb.WriteString("; External declarations\n")
	g.sb.WriteString("declare i64 @print(i64)\n")
	g.sb.WriteString("declare i64 @println(i64)\n")
	g.sb.WriteString("declare i64 @printstr(ptr)\n")
	g.sb.WriteString("\n")
}

func (g *Generator) emitGlobal(global *ir.Global) {
	// Emit global string constant
	// Format: @str.0 = private constant [6 x i8] c"hello\00"
	escaped := escapeString(global.Value)
	length := len(global.Value) + 1 // +1 for null terminator
	g.sb.WriteString(fmt.Sprintf("@%s = private constant [%d x i8] c\"%s\\00\"\n",
		global.Name, length, escaped))
}

func (g *Generator) emitFunction(fn *ir.Function) {
	// Function signature
	retType := g.llvmType(fn.ReturnType)
	params := make([]string, len(fn.Params))
	for i, p := range fn.Params {
		params[i] = fmt.Sprintf("%s %%%s", g.llvmType(p.Type), p.Name)
	}

	g.sb.WriteString(fmt.Sprintf("define %s @%s(%s) {\n",
		retType, fn.Name, strings.Join(params, ", ")))

	// Basic blocks
	for _, block := range fn.Blocks {
		g.emitBlock(block)
	}

	g.sb.WriteString("}\n")
}

func (g *Generator) emitBlock(block *ir.BasicBlock) {
	g.sb.WriteString(block.Label)
	g.sb.WriteString(":\n")

	// Instructions
	for _, instr := range block.Instrs {
		g.sb.WriteString("    ")
		g.emitInstruction(instr)
		g.sb.WriteString("\n")
	}

	// Terminator
	g.sb.WriteString("    ")
	g.emitTerminator(block.Term)
	g.sb.WriteString("\n")
}

func (g *Generator) emitInstruction(instr ir.Instruction) {
	switch i := instr.(type) {
	case *ir.Alloca:
		g.emitAlloca(i)
	case *ir.Load:
		g.emitLoad(i)
	case *ir.Store:
		g.emitStore(i)
	case *ir.BinOp:
		g.emitBinOp(i)
	case *ir.UnaryOp:
		g.emitUnaryOp(i)
	case *ir.Call:
		g.emitCall(i)
	case *ir.Copy:
		g.emitCopy(i)
	case *ir.Phi:
		g.emitPhi(i)
	default:
		g.sb.WriteString(fmt.Sprintf("; unknown instruction: %T", instr))
	}
}

func (g *Generator) emitAlloca(a *ir.Alloca) {
	g.sb.WriteString(fmt.Sprintf("%s = alloca %s",
		g.llvmValue(a.Dest), g.llvmType(a.Type)))
}

func (g *Generator) emitLoad(l *ir.Load) {
	g.sb.WriteString(fmt.Sprintf("%s = load %s, ptr %s",
		g.llvmValue(l.Dest), g.llvmType(l.Type), g.llvmValue(l.Ptr)))
}

func (g *Generator) emitStore(s *ir.Store) {
	g.sb.WriteString(fmt.Sprintf("store %s %s, ptr %s",
		g.llvmType(s.Type), g.llvmValue(s.Val), g.llvmValue(s.Ptr)))
}

func (g *Generator) emitBinOp(b *ir.BinOp) {
	dest := g.llvmValue(b.Dest)
	left := g.llvmValue(b.Left)
	right := g.llvmValue(b.Right)
	typ := g.llvmType(b.Type)

	switch b.Op {
	case ir.OpAdd:
		g.sb.WriteString(fmt.Sprintf("%s = add %s %s, %s", dest, typ, left, right))
	case ir.OpSub:
		g.sb.WriteString(fmt.Sprintf("%s = sub %s %s, %s", dest, typ, left, right))
	case ir.OpMul:
		g.sb.WriteString(fmt.Sprintf("%s = mul %s %s, %s", dest, typ, left, right))
	case ir.OpDiv:
		g.sb.WriteString(fmt.Sprintf("%s = sdiv %s %s, %s", dest, typ, left, right))
	case ir.OpMod:
		g.sb.WriteString(fmt.Sprintf("%s = srem %s %s, %s", dest, typ, left, right))
	case ir.OpEq:
		g.sb.WriteString(fmt.Sprintf("%s = icmp eq %s %s, %s", dest, typ, left, right))
	case ir.OpNeq:
		g.sb.WriteString(fmt.Sprintf("%s = icmp ne %s %s, %s", dest, typ, left, right))
	case ir.OpLt:
		g.sb.WriteString(fmt.Sprintf("%s = icmp slt %s %s, %s", dest, typ, left, right))
	case ir.OpGt:
		g.sb.WriteString(fmt.Sprintf("%s = icmp sgt %s %s, %s", dest, typ, left, right))
	case ir.OpLte:
		g.sb.WriteString(fmt.Sprintf("%s = icmp sle %s %s, %s", dest, typ, left, right))
	case ir.OpGte:
		g.sb.WriteString(fmt.Sprintf("%s = icmp sge %s %s, %s", dest, typ, left, right))
	case ir.OpAnd:
		g.sb.WriteString(fmt.Sprintf("%s = and %s %s, %s", dest, typ, left, right))
	case ir.OpOr:
		g.sb.WriteString(fmt.Sprintf("%s = or %s %s, %s", dest, typ, left, right))
	default:
		g.sb.WriteString(fmt.Sprintf("; unknown binop: %v", b.Op))
	}
}

func (g *Generator) emitUnaryOp(u *ir.UnaryOp) {
	dest := g.llvmValue(u.Dest)
	operand := g.llvmValue(u.Operand)
	typ := g.llvmType(u.Type)

	switch u.Op {
	case ir.OpNeg:
		// neg x = 0 - x
		g.sb.WriteString(fmt.Sprintf("%s = sub %s 0, %s", dest, typ, operand))
	case ir.OpNot:
		// not x = xor x, true
		g.sb.WriteString(fmt.Sprintf("%s = xor %s %s, true", dest, typ, operand))
	default:
		g.sb.WriteString(fmt.Sprintf("; unknown unaryop: %v", u.Op))
	}
}

func (g *Generator) emitCall(c *ir.Call) {
	retType := g.llvmType(c.RetType)

	// Build argument list with types
	args := make([]string, len(c.Args))
	for i, arg := range c.Args {
		argType := g.inferValueType(arg)
		args[i] = fmt.Sprintf("%s %s", argType, g.llvmValue(arg))
	}

	if c.Dest == nil || types.Equal(c.RetType, types.Void) {
		g.sb.WriteString(fmt.Sprintf("call void @%s(%s)", c.Func, strings.Join(args, ", ")))
	} else {
		g.sb.WriteString(fmt.Sprintf("%s = call %s @%s(%s)",
			g.llvmValue(c.Dest), retType, c.Func, strings.Join(args, ", ")))
	}
}

func (g *Generator) emitCopy(c *ir.Copy) {
	dest := g.llvmValue(c.Dest)
	val := g.llvmValue(c.Val)
	typ := g.llvmType(c.Type)

	// Copy is implemented as: add x, 0 for integers, or i1 for bools
	if types.Equal(c.Type, types.Bool) {
		// For bools: or x, false (identity)
		g.sb.WriteString(fmt.Sprintf("%s = or %s %s, false", dest, typ, val))
	} else {
		// For integers: add x, 0 (identity)
		g.sb.WriteString(fmt.Sprintf("%s = add %s %s, 0", dest, typ, val))
	}
}

func (g *Generator) emitPhi(p *ir.Phi) {
	dest := g.llvmValue(p.Dest)
	typ := g.llvmType(p.Type)

	entries := make([]string, len(p.Entries))
	for i, e := range p.Entries {
		entries[i] = fmt.Sprintf("[ %s, %%%s ]", g.llvmValue(e.Val), e.Block)
	}

	g.sb.WriteString(fmt.Sprintf("%s = phi %s %s", dest, typ, strings.Join(entries, ", ")))
}

func (g *Generator) emitTerminator(term ir.Terminator) {
	switch t := term.(type) {
	case *ir.RetVoid:
		g.sb.WriteString("ret void")
	case *ir.Ret:
		g.sb.WriteString(fmt.Sprintf("ret %s %s", g.llvmType(t.Type), g.llvmValue(t.Val)))
	case *ir.Br:
		g.sb.WriteString(fmt.Sprintf("br label %%%s", t.Target))
	case *ir.CondBr:
		g.sb.WriteString(fmt.Sprintf("br i1 %s, label %%%s, label %%%s",
			g.llvmValue(t.Cond), t.Then, t.Else))
	default:
		g.sb.WriteString(fmt.Sprintf("; unknown terminator: %T", term))
	}
}

// llvmType converts F1-Lang types to LLVM types.
func (g *Generator) llvmType(t types.Type) string {
	switch tp := t.(type) {
	case types.PrimitiveType:
		switch tp {
		case types.Int:
			return "i64"
		case types.Bool:
			return "i1"
		case types.String:
			return "ptr"
		case types.Void:
			return "void"
		}
	case *types.FunctionType:
		return "ptr" // Function pointers
	}
	return "i64" // Default to i64
}

// llvmValue converts F1-IR values to LLVM values.
func (g *Generator) llvmValue(v ir.Value) string {
	switch val := v.(type) {
	case *ir.Reg:
		return fmt.Sprintf("%%%d", val.ID)
	case *ir.IntConst:
		return fmt.Sprintf("%d", val.Value)
	case *ir.BoolConst:
		if val.Value {
			return "true"
		}
		return "false"
	case *ir.StringConst:
		return fmt.Sprintf("@%s", val.Name)
	case *ir.ParamRef:
		return fmt.Sprintf("%%%s", val.Name)
	default:
		return "0"
	}
}

// inferValueType attempts to infer the LLVM type of a value.
func (g *Generator) inferValueType(v ir.Value) string {
	switch v.(type) {
	case *ir.BoolConst:
		return "i1"
	case *ir.StringConst:
		return "ptr"
	default:
		return "i64"
	}
}

// escapeString escapes special characters for LLVM string constants.
func escapeString(s string) string {
	var result strings.Builder
	for _, r := range s {
		switch r {
		case '\\':
			result.WriteString("\\5C")
		case '"':
			result.WriteString("\\22")
		case '\n':
			result.WriteString("\\0A")
		case '\r':
			result.WriteString("\\0D")
		case '\t':
			result.WriteString("\\09")
		default:
			if r < 32 || r > 126 {
				result.WriteString(fmt.Sprintf("\\%02X", r))
			} else {
				result.WriteRune(r)
			}
		}
	}
	return result.String()
}
