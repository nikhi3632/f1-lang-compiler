package irgen

import (
	"fmt"

	"f1c/ast"
	"f1c/ir"
	"f1c/types"
)

// Generator generates F1-IR from an AST.
type Generator struct {
	module     *ir.Module
	curFunc    *ir.Function
	curBlock   *ir.BasicBlock
	regCounter int
	blockCounter int

	// Variable locations (alloca results)
	locals map[string]*ir.Reg

	// Track if we're in main or a function
	inFunction bool
}

// New creates a new IR generator.
func New() *Generator {
	return &Generator{
		module: &ir.Module{},
		locals: make(map[string]*ir.Reg),
	}
}

// Generate generates IR for a program.
func (g *Generator) Generate(program *ast.Program) *ir.Module {
	// Separate function declarations from top-level statements
	var funcDecls []*ast.FunctionDecl
	var topLevelStmts []ast.Statement

	for _, stmt := range program.Statements {
		if fn, ok := stmt.(*ast.FunctionDecl); ok {
			funcDecls = append(funcDecls, fn)
		} else {
			topLevelStmts = append(topLevelStmts, stmt)
		}
	}

	// Generate function declarations first
	for _, fn := range funcDecls {
		g.generateFunction(fn)
	}

	// Generate main function for top-level statements
	if len(topLevelStmts) > 0 {
		g.generateMain(topLevelStmts)
	}

	return g.module
}

func (g *Generator) generateMain(stmts []ast.Statement) {
	g.curFunc = &ir.Function{
		Name:       "main",
		Params:     nil,
		ReturnType: types.Void,
	}
	g.module.Functions = append(g.module.Functions, g.curFunc)

	g.curBlock = g.newBlock("entry")
	g.curFunc.Blocks = append(g.curFunc.Blocks, g.curBlock)

	g.locals = make(map[string]*ir.Reg)
	g.regCounter = 0

	for _, stmt := range stmts {
		g.generateStatement(stmt)
	}

	// Add ret void if not already terminated
	if g.curBlock.Term == nil {
		g.curBlock.Term = &ir.RetVoid{}
	}
}

func (g *Generator) generateFunction(fn *ast.FunctionDecl) {
	g.inFunction = true
	defer func() { g.inFunction = false }()

	// Create parameters
	params := make([]*ir.Param, len(fn.Params))
	for i, p := range fn.Params {
		params[i] = &ir.Param{Name: p.Name.Value, Type: types.Int}
	}

	// Determine return type by checking body for return statements
	returnType := g.inferReturnType(fn.Body)

	g.curFunc = &ir.Function{
		Name:       fn.Name.Value,
		Params:     params,
		ReturnType: returnType,
	}
	g.module.Functions = append(g.module.Functions, g.curFunc)

	g.curBlock = g.newBlock("entry")
	g.curFunc.Blocks = append(g.curFunc.Blocks, g.curBlock)

	// Reset locals and counters for this function
	g.locals = make(map[string]*ir.Reg)
	g.regCounter = 0

	// Create allocas for parameters and store them
	for _, p := range fn.Params {
		ptr := g.newReg()
		g.emit(&ir.Alloca{Dest: ptr, Type: types.Int})
		g.emit(&ir.Store{Type: types.Int, Val: &ir.ParamRef{Name: p.Name.Value}, Ptr: ptr})
		g.locals[p.Name.Value] = ptr
	}

	// Generate body
	for _, stmt := range fn.Body.Statements {
		g.generateStatement(stmt)
	}

	// Add implicit return if needed
	if g.curBlock.Term == nil {
		if types.Equal(returnType, types.Void) {
			g.curBlock.Term = &ir.RetVoid{}
		}
	}
}

func (g *Generator) inferReturnType(body *ast.BlockStmt) types.Type {
	for _, stmt := range body.Statements {
		if ret, ok := stmt.(*ast.ReturnStmt); ok {
			if ret.Value == nil {
				return types.Void
			}
			return g.inferExprType(ret.Value)
		}
		// Check nested if statements
		if ifStmt, ok := stmt.(*ast.IfStmt); ok {
			if t := g.inferReturnTypeFromIf(ifStmt); t != nil {
				return t
			}
		}
	}
	return types.Void
}

func (g *Generator) inferReturnTypeFromIf(ifStmt *ast.IfStmt) types.Type {
	if t := g.inferReturnType(ifStmt.Consequence); !types.Equal(t, types.Void) {
		return t
	}
	if ifStmt.Alternative != nil {
		if block, ok := ifStmt.Alternative.(*ast.BlockStmt); ok {
			if t := g.inferReturnType(block); !types.Equal(t, types.Void) {
				return t
			}
		}
	}
	return nil
}

func (g *Generator) inferExprType(expr ast.Expression) types.Type {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return types.Int
	case *ast.BooleanLiteral:
		return types.Bool
	case *ast.StringLiteral:
		return types.String
	case *ast.InfixExpr:
		switch e.Operator {
		case "<", ">", "<=", ">=", "==", "!=", "&&", "||":
			return types.Bool
		default:
			return types.Int
		}
	case *ast.PrefixExpr:
		if e.Operator == "!" {
			return types.Bool
		}
		return types.Int
	default:
		return types.Int
	}
}

func (g *Generator) generateStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.VarDecl:
		g.generateVarDecl(s)
	case *ast.Assignment:
		g.generateAssignment(s)
	case *ast.ReturnStmt:
		g.generateReturn(s)
	case *ast.IfStmt:
		g.generateIf(s)
	case *ast.ForStmt:
		g.generateFor(s)
	case *ast.ExpressionStmt:
		g.generateExpression(s.Expression)
	case *ast.BlockStmt:
		for _, inner := range s.Statements {
			g.generateStatement(inner)
		}
	}
}

func (g *Generator) generateVarDecl(decl *ast.VarDecl) {
	// Allocate space
	typ := g.inferExprType(decl.Value)
	ptr := g.newReg()
	g.emit(&ir.Alloca{Dest: ptr, Type: typ})

	// Generate value
	val := g.generateExpression(decl.Value)

	// Store value
	g.emit(&ir.Store{Type: typ, Val: val, Ptr: ptr})

	// Remember location
	g.locals[decl.Name.Value] = ptr
}

func (g *Generator) generateAssignment(assign *ast.Assignment) {
	// Get location
	ptr, ok := g.locals[assign.Name.Value]
	if !ok {
		return // Error: undefined variable (should be caught by type checker)
	}

	// Generate value
	val := g.generateExpression(assign.Value)
	typ := g.inferExprType(assign.Value)

	// Store value
	g.emit(&ir.Store{Type: typ, Val: val, Ptr: ptr})
}

func (g *Generator) generateReturn(ret *ast.ReturnStmt) {
	if ret.Value == nil {
		g.curBlock.Term = &ir.RetVoid{}
		return
	}

	val := g.generateExpression(ret.Value)
	typ := g.inferExprType(ret.Value)
	g.curBlock.Term = &ir.Ret{Type: typ, Val: val}
}

func (g *Generator) generateIf(ifStmt *ast.IfStmt) {
	// Generate condition
	cond := g.generateExpression(ifStmt.Condition)

	// Create blocks
	thenBlock := g.newBlock("then")
	mergeBlock := g.newBlock("merge")

	var elseBlock *ir.BasicBlock
	if ifStmt.Alternative != nil {
		elseBlock = g.newBlock("else")
		g.curBlock.Term = &ir.CondBr{Cond: cond, Then: thenBlock.Label, Else: elseBlock.Label}
	} else {
		g.curBlock.Term = &ir.CondBr{Cond: cond, Then: thenBlock.Label, Else: mergeBlock.Label}
	}

	// Generate then block
	g.curFunc.Blocks = append(g.curFunc.Blocks, thenBlock)
	g.curBlock = thenBlock
	for _, stmt := range ifStmt.Consequence.Statements {
		g.generateStatement(stmt)
	}
	if g.curBlock.Term == nil {
		g.curBlock.Term = &ir.Br{Target: mergeBlock.Label}
	}

	// Generate else block if present
	if elseBlock != nil {
		g.curFunc.Blocks = append(g.curFunc.Blocks, elseBlock)
		g.curBlock = elseBlock
		if block, ok := ifStmt.Alternative.(*ast.BlockStmt); ok {
			for _, stmt := range block.Statements {
				g.generateStatement(stmt)
			}
		} else {
			g.generateStatement(ifStmt.Alternative)
		}
		if g.curBlock.Term == nil {
			g.curBlock.Term = &ir.Br{Target: mergeBlock.Label}
		}
	}

	// Continue with merge block
	g.curFunc.Blocks = append(g.curFunc.Blocks, mergeBlock)
	g.curBlock = mergeBlock
}

func (g *Generator) generateFor(forStmt *ast.ForStmt) {
	// Generate init
	if forStmt.Init != nil {
		g.generateStatement(forStmt.Init)
	}

	// Create blocks
	condBlock := g.newBlock("loop.cond")
	bodyBlock := g.newBlock("loop.body")
	updateBlock := g.newBlock("loop.update")
	exitBlock := g.newBlock("loop.exit")

	// Branch to condition
	g.curBlock.Term = &ir.Br{Target: condBlock.Label}

	// Generate condition
	g.curFunc.Blocks = append(g.curFunc.Blocks, condBlock)
	g.curBlock = condBlock
	if forStmt.Condition != nil {
		cond := g.generateExpression(forStmt.Condition)
		g.curBlock.Term = &ir.CondBr{Cond: cond, Then: bodyBlock.Label, Else: exitBlock.Label}
	} else {
		g.curBlock.Term = &ir.Br{Target: bodyBlock.Label}
	}

	// Generate body
	g.curFunc.Blocks = append(g.curFunc.Blocks, bodyBlock)
	g.curBlock = bodyBlock
	for _, stmt := range forStmt.Body.Statements {
		g.generateStatement(stmt)
	}
	if g.curBlock.Term == nil {
		g.curBlock.Term = &ir.Br{Target: updateBlock.Label}
	}

	// Generate update
	g.curFunc.Blocks = append(g.curFunc.Blocks, updateBlock)
	g.curBlock = updateBlock
	if forStmt.Update != nil {
		g.generateStatement(forStmt.Update)
	}
	g.curBlock.Term = &ir.Br{Target: condBlock.Label}

	// Continue with exit block
	g.curFunc.Blocks = append(g.curFunc.Blocks, exitBlock)
	g.curBlock = exitBlock
}

func (g *Generator) generateExpression(expr ast.Expression) ir.Value {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return &ir.IntConst{Value: e.Value}

	case *ast.BooleanLiteral:
		return &ir.BoolConst{Value: e.Value}

	case *ast.StringLiteral:
		// Add to globals and return reference
		name := fmt.Sprintf("str.%d", len(g.module.Globals))
		g.module.Globals = append(g.module.Globals, &ir.Global{Name: name, Value: e.Value})
		return &ir.StringConst{Index: len(g.module.Globals) - 1, Name: name}

	case *ast.Identifier:
		return g.generateIdentifier(e)

	case *ast.PrefixExpr:
		return g.generatePrefix(e)

	case *ast.InfixExpr:
		return g.generateInfix(e)

	case *ast.CallExpr:
		return g.generateCall(e)

	case *ast.GroupedExpr:
		return g.generateExpression(e.Expression)

	default:
		return &ir.IntConst{Value: 0}
	}
}

func (g *Generator) generateIdentifier(ident *ast.Identifier) ir.Value {
	// Check if it's a parameter
	if g.inFunction {
		for _, p := range g.curFunc.Params {
			if p.Name == ident.Value {
				// Load from the alloca'd parameter
				ptr := g.locals[ident.Value]
				if ptr != nil {
					dest := g.newReg()
					g.emit(&ir.Load{Dest: dest, Type: p.Type, Ptr: ptr})
					return dest
				}
			}
		}
	}

	// Check locals
	ptr, ok := g.locals[ident.Value]
	if !ok {
		return &ir.IntConst{Value: 0} // Error: undefined (should be caught by type checker)
	}

	dest := g.newReg()
	g.emit(&ir.Load{Dest: dest, Type: types.Int, Ptr: ptr})
	return dest
}

func (g *Generator) generatePrefix(expr *ast.PrefixExpr) ir.Value {
	operand := g.generateExpression(expr.Right)
	dest := g.newReg()

	switch expr.Operator {
	case "-":
		g.emit(&ir.UnaryOp{Dest: dest, Op: ir.OpNeg, Type: types.Int, Operand: operand})
	case "!":
		g.emit(&ir.UnaryOp{Dest: dest, Op: ir.OpNot, Type: types.Bool, Operand: operand})
	}

	return dest
}

func (g *Generator) generateInfix(expr *ast.InfixExpr) ir.Value {
	left := g.generateExpression(expr.Left)
	right := g.generateExpression(expr.Right)
	dest := g.newReg()

	var op ir.Op
	var typ types.Type = types.Int

	switch expr.Operator {
	case "+":
		op = ir.OpAdd
	case "-":
		op = ir.OpSub
	case "*":
		op = ir.OpMul
	case "/":
		op = ir.OpDiv
	case "%":
		op = ir.OpMod
	case "<":
		op = ir.OpLt
	case ">":
		op = ir.OpGt
	case "<=":
		op = ir.OpLte
	case ">=":
		op = ir.OpGte
	case "==":
		op = ir.OpEq
	case "!=":
		op = ir.OpNeq
	case "&&":
		op = ir.OpAnd
		typ = types.Bool
	case "||":
		op = ir.OpOr
		typ = types.Bool
	}

	g.emit(&ir.BinOp{Dest: dest, Op: op, Type: typ, Left: left, Right: right})
	return dest
}

func (g *Generator) generateCall(call *ast.CallExpr) ir.Value {
	// Get function name
	ident, ok := call.Function.(*ast.Identifier)
	if !ok {
		return &ir.IntConst{Value: 0}
	}

	// Generate arguments
	args := make([]ir.Value, len(call.Arguments))
	for i, arg := range call.Arguments {
		args[i] = g.generateExpression(arg)
	}

	// For now, assume all functions return int
	// TODO: Look up actual return type from symbol table
	dest := g.newReg()
	g.emit(&ir.Call{Dest: dest, Func: ident.Value, Args: args, RetType: types.Int})

	return dest
}

// Helper functions

func (g *Generator) newReg() *ir.Reg {
	r := ir.NewReg(g.regCounter)
	g.regCounter++
	return r
}

func (g *Generator) newBlock(prefix string) *ir.BasicBlock {
	label := fmt.Sprintf("%s.%d", prefix, g.blockCounter)
	g.blockCounter++
	return &ir.BasicBlock{Label: label, Instrs: []ir.Instruction{}}
}

func (g *Generator) emit(instr ir.Instruction) {
	g.curBlock.Instrs = append(g.curBlock.Instrs, instr)
}
