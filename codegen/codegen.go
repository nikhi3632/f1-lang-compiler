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
	// Declare external functions
	g.sb.WriteString("; External declarations\n")
	// I/O functions
	g.sb.WriteString("declare void @f1_radio_int(i64)\n")
	g.sb.WriteString("declare void @f1_radio_str(ptr)\n")
	g.sb.WriteString("declare void @f1_radio_bool(i1)\n")
	g.sb.WriteString("declare void @f1_bono(ptr)\n")
	// Graphics functions
	g.sb.WriteString("declare void @f1_canvas(i64, i64)\n")
	g.sb.WriteString("declare void @f1_pixel(i64, i64, i64, i64, i64)\n")
	g.sb.WriteString("declare void @f1_render(ptr)\n")
	g.sb.WriteString("declare void @f1_snapshot(i64)\n")
	g.sb.WriteString("declare void @f1_framedir(ptr)\n")
	// String comparison
	g.sb.WriteString("declare i32 @strcmp(ptr, ptr)\n")
	// Memory allocation for closures
	g.sb.WriteString("declare ptr @malloc(i64)\n")
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
		// For closures, the first parameter is the environment pointer
		if fn.IsClosure && i == 0 && p.Name == "__env" {
			params[i] = "ptr %__env"
		} else {
			params[i] = fmt.Sprintf("%s %%%s", g.llvmType(p.Type), p.Name)
		}
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
	case *ir.MakeClosure:
		g.emitMakeClosure(i)
	case *ir.GetEnvField:
		g.emitGetEnvField(i)
	case *ir.CallIndirect:
		g.emitCallIndirect(i)
	case *ir.ClosureCall:
		g.emitClosureCall(i)
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

	// Build argument list with types, converting FuncRefs to i64
	args := make([]string, len(c.Args))
	for i, arg := range c.Args {
		if funcRef, ok := arg.(*ir.FuncRef); ok {
			// Function references need to be converted to i64 for passing
			// Use ptrtoint to convert function pointer to integer
			args[i] = fmt.Sprintf("i64 ptrtoint (ptr @%s to i64)", funcRef.Name)
		} else {
			// Use provided ArgTypes if available, otherwise infer
			var argType string
			if i < len(c.ArgTypes) && c.ArgTypes[i] != nil {
				argType = g.llvmType(c.ArgTypes[i])
			} else {
				argType = g.inferValueType(arg)
			}
			args[i] = fmt.Sprintf("%s %s", argType, g.llvmValue(arg))
		}
	}

	// Use musttail for tail calls (TCO)
	callPrefix := "call"
	if c.IsTail {
		callPrefix = "musttail call"
	}

	if c.Dest == nil || types.Equal(c.RetType, types.Void) {
		g.sb.WriteString(fmt.Sprintf("%s void @%s(%s)", callPrefix, c.Func, strings.Join(args, ", ")))
	} else {
		g.sb.WriteString(fmt.Sprintf("%s = %s %s @%s(%s)",
			g.llvmValue(c.Dest), callPrefix, retType, c.Func, strings.Join(args, ", ")))
	}
}

// emitCallIndirect generates LLVM IR for indirect function calls (higher-order functions).
func (g *Generator) emitCallIndirect(c *ir.CallIndirect) {
	retType := g.llvmType(c.RetType)

	// Build argument list with types
	args := make([]string, len(c.Args))
	for i, arg := range c.Args {
		argType := g.inferValueType(arg)
		args[i] = fmt.Sprintf("%s %s", argType, g.llvmValue(arg))
	}

	// Build function type for indirect call
	argTypes := make([]string, len(c.Args))
	for i := range c.Args {
		argTypes[i] = "i64"
	}
	funcType := fmt.Sprintf("%s (%s)", retType, strings.Join(argTypes, ", "))

	// Convert function pointer (i64) to ptr
	funcPtr := g.llvmValue(c.FuncPtr)
	ptrReg := fmt.Sprintf("%%__fptr_%d", c.Dest.ID)
	g.sb.WriteString(fmt.Sprintf("%s = inttoptr i64 %s to ptr\n    ", ptrReg, funcPtr))

	if c.Dest == nil || types.Equal(c.RetType, types.Void) {
		g.sb.WriteString(fmt.Sprintf("call void %s(%s)", ptrReg, strings.Join(args, ", ")))
	} else {
		g.sb.WriteString(fmt.Sprintf("%s = call %s %s(%s)",
			g.llvmValue(c.Dest), funcType, ptrReg, strings.Join(args, ", ")))
	}
}

// emitClosureCall generates LLVM IR for calling through a closure struct.
// Closure struct layout: { ptr func_ptr, ptr env_ptr }
func (g *Generator) emitClosureCall(c *ir.ClosureCall) {
	retType := g.llvmType(c.RetType)
	closurePtr := g.llvmValue(c.ClosurePtr)
	destID := 0
	if c.Dest != nil {
		destID = c.Dest.ID
	}

	// Extract function pointer from closure[0]
	funcSlot := fmt.Sprintf("%%__cc_func_slot_%d", destID)
	funcPtr := fmt.Sprintf("%%__cc_func_%d", destID)
	g.sb.WriteString(fmt.Sprintf("%s = getelementptr ptr, ptr %s, i64 0\n    ", funcSlot, closurePtr))
	g.sb.WriteString(fmt.Sprintf("%s = load ptr, ptr %s\n    ", funcPtr, funcSlot))

	// Extract env pointer from closure[1]
	envSlot := fmt.Sprintf("%%__cc_env_slot_%d", destID)
	envPtr := fmt.Sprintf("%%__cc_env_%d", destID)
	g.sb.WriteString(fmt.Sprintf("%s = getelementptr ptr, ptr %s, i64 1\n    ", envSlot, closurePtr))
	g.sb.WriteString(fmt.Sprintf("%s = load ptr, ptr %s\n    ", envPtr, envSlot))

	// Build argument list with env as first argument
	args := make([]string, len(c.Args)+1)
	args[0] = fmt.Sprintf("ptr %s", envPtr)
	for i, arg := range c.Args {
		argType := g.inferValueType(arg)
		args[i+1] = fmt.Sprintf("%s %s", argType, g.llvmValue(arg))
	}

	// Build function type for indirect call
	// Function signature: retType (ptr env, arg types...)
	argTypes := make([]string, len(c.Args)+1)
	argTypes[0] = "ptr"
	for i := range c.Args {
		argTypes[i+1] = "i64"
	}
	funcType := fmt.Sprintf("%s (%s)", retType, strings.Join(argTypes, ", "))

	if c.Dest == nil || types.Equal(c.RetType, types.Void) {
		g.sb.WriteString(fmt.Sprintf("call void %s(%s)", funcPtr, strings.Join(args, ", ")))
	} else {
		g.sb.WriteString(fmt.Sprintf("%s = call %s %s(%s)",
			g.llvmValue(c.Dest), funcType, funcPtr, strings.Join(args, ", ")))
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

// emitMakeClosure generates LLVM IR for closure creation.
// Creates a closure struct: { ptr func_ptr, ptr env_ptr }
func (g *Generator) emitMakeClosure(m *ir.MakeClosure) {
	numCaptures := len(m.Captures)

	// Allocate closure struct: { func_ptr, env_ptr } = 16 bytes
	g.sb.WriteString(fmt.Sprintf("%s = call ptr @malloc(i64 16)\n",
		g.llvmValue(m.Dest)))

	// Store function pointer at offset 0
	g.sb.WriteString(fmt.Sprintf("    %%__closure_func_%d = getelementptr ptr, ptr %s, i64 0\n",
		m.Dest.ID, g.llvmValue(m.Dest)))
	g.sb.WriteString(fmt.Sprintf("    store ptr @%s, ptr %%__closure_func_%d\n",
		m.FuncName, m.Dest.ID))

	if numCaptures == 0 {
		// No captures - store null for env
		g.sb.WriteString(fmt.Sprintf("    %%__closure_env_slot_%d = getelementptr ptr, ptr %s, i64 1\n",
			m.Dest.ID, g.llvmValue(m.Dest)))
		g.sb.WriteString(fmt.Sprintf("    store ptr null, ptr %%__closure_env_slot_%d",
			m.Dest.ID))
		return
	}

	// Allocate environment struct on heap
	// Size = numCaptures * 8 bytes (i64)
	envSize := numCaptures * 8
	g.sb.WriteString(fmt.Sprintf("    %%__env_%d = call ptr @malloc(i64 %d)\n",
		m.Dest.ID, envSize))

	// Store captured values into environment
	for i, cap := range m.Captures {
		g.sb.WriteString(fmt.Sprintf("    %%__env_ptr_%d_%d = getelementptr i64, ptr %%__env_%d, i64 %d\n",
			m.Dest.ID, i, m.Dest.ID, i))
		g.sb.WriteString(fmt.Sprintf("    store i64 %s, ptr %%__env_ptr_%d_%d\n",
			g.llvmValue(cap), m.Dest.ID, i))
	}

	// Store env pointer at offset 1 in closure struct
	g.sb.WriteString(fmt.Sprintf("    %%__closure_env_slot_%d = getelementptr ptr, ptr %s, i64 1\n",
		m.Dest.ID, g.llvmValue(m.Dest)))
	g.sb.WriteString(fmt.Sprintf("    store ptr %%__env_%d, ptr %%__closure_env_slot_%d",
		m.Dest.ID, m.Dest.ID))
}

// emitGetEnvField generates LLVM IR to load a value from the closure environment.
func (g *Generator) emitGetEnvField(gef *ir.GetEnvField) {
	// Get pointer to field in environment
	g.sb.WriteString(fmt.Sprintf("%%__env_field_%d = getelementptr i64, ptr %s, i64 %d\n",
		gef.Dest.ID, g.llvmValue(gef.Env), gef.Index))
	g.sb.WriteString(fmt.Sprintf("    %s = load i64, ptr %%__env_field_%d",
		g.llvmValue(gef.Dest), gef.Dest.ID))
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
	case *ir.FuncRef:
		return fmt.Sprintf("@%s", val.Name)
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
	case *ir.FuncRef:
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
