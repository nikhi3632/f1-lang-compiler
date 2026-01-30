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

	// Lambda counter for unique names
	lambdaCounter int

	// Track function variables (for indirect calls)
	funcVars map[string]string // variable name -> lifted function name

	// Track closure variables (function name + environment register)
	closureVars map[string]*closureInfo

	// Track function return types for call generation
	funcReturnTypes map[string]types.Type

	// Track variable types for type inference
	varTypes map[string]types.Type
}

type closureInfo struct {
	funcName string
	envReg   *ir.Reg
}

// New creates a new IR generator.
func New() *Generator {
	return &Generator{
		module:          &ir.Module{},
		locals:          make(map[string]*ir.Reg),
		funcVars:        make(map[string]string),
		closureVars:     make(map[string]*closureInfo),
		funcReturnTypes: make(map[string]types.Type),
		varTypes:        make(map[string]types.Type),
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
	g.varTypes = make(map[string]types.Type)
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

	// Store return type for later call generation
	g.funcReturnTypes[fn.Name.Value] = returnType

	g.curFunc = &ir.Function{
		Name:       fn.Name.Value,
		Params:     params,
		ReturnType: returnType,
	}
	g.module.Functions = append(g.module.Functions, g.curFunc)

	g.curBlock = g.newBlock("entry")
	g.curFunc.Blocks = append(g.curFunc.Blocks, g.curBlock)

	// Reset locals, types, and counters for this function
	g.locals = make(map[string]*ir.Reg)
	g.varTypes = make(map[string]types.Type)
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
	case *ast.Identifier:
		// Look up variable type
		if t, ok := g.varTypes[e.Value]; ok {
			return t
		}
		return types.Int // default for unknown variables (e.g., parameters)
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
	case *ast.LambdaExpr:
		// Lambda expressions have function type
		paramTypes := make([]types.Type, len(e.Params))
		for i := range e.Params {
			paramTypes[i] = types.Int
		}
		return &types.FunctionType{
			Params: paramTypes,
			Return: g.inferReturnType(e.Body),
		}
	case *ast.CallExpr:
		// Look up return type of the function being called
		if ident, ok := e.Function.(*ast.Identifier); ok {
			if rt, ok := g.funcReturnTypes[ident.Value]; ok {
				return rt
			}
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
	// Check if this is a lambda assignment
	if lambda, ok := decl.Value.(*ast.LambdaExpr); ok {
		// Generate the lambda (lifts to a function)
		result := g.generateLambda(lambda)

		// Check if it's a closure (returns *ir.Reg) or plain function (returns *ir.FuncRef)
		switch v := result.(type) {
		case *ir.FuncRef:
			// No captures - track as simple function variable
			g.funcVars[decl.Name.Value] = v.Name
		case *ir.Reg:
			// Has captures - track as closure with environment
			// Find the function name from the most recent lambda
			funcName := fmt.Sprintf("__lambda_%d", g.lambdaCounter-1)
			g.closureVars[decl.Name.Value] = &closureInfo{
				funcName: funcName,
				envReg:   v,
			}
		}
		// Don't allocate storage - we use the function/closure directly
		return
	}

	// Allocate space
	typ := g.inferExprType(decl.Value)
	ptr := g.newReg()
	g.emit(&ir.Alloca{Dest: ptr, Type: typ})

	// Generate value
	val := g.generateExpression(decl.Value)

	// Store value
	g.emit(&ir.Store{Type: typ, Val: val, Ptr: ptr})

	// Remember location and type
	g.locals[decl.Name.Value] = ptr
	g.varTypes[decl.Name.Value] = typ
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

	case *ast.LambdaExpr:
		return g.generateLambda(e)

	default:
		return &ir.IntConst{Value: 0}
	}
}

func (g *Generator) generateIdentifier(ident *ast.Identifier) ir.Value {
	// Check if it's a function variable (lambda assignment)
	if funcName, ok := g.funcVars[ident.Value]; ok {
		// Return function reference for higher-order function passing
		return &ir.FuncRef{Name: funcName}
	}

	// Check if it's a closure variable
	if closure, ok := g.closureVars[ident.Value]; ok {
		// Return the environment register for closure passing
		return closure.envReg
	}

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

	// Use the variable's actual type
	var varType types.Type = types.Int
	if t, ok := g.varTypes[ident.Value]; ok {
		varType = t
	}

	dest := g.newReg()
	g.emit(&ir.Load{Dest: dest, Type: varType, Ptr: ptr})
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
	// Handle short-circuit evaluation for && and ||
	if expr.Operator == "&&" {
		return g.generateShortCircuitAnd(expr)
	}
	if expr.Operator == "||" {
		return g.generateShortCircuitOr(expr)
	}

	// Handle string comparison via strcmp
	leftType := g.inferExprType(expr.Left)
	if types.Equal(leftType, types.String) && (expr.Operator == "==" || expr.Operator == "!=") {
		return g.generateStringComparison(expr)
	}

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
	}

	g.emit(&ir.BinOp{Dest: dest, Op: op, Type: typ, Left: left, Right: right})
	return dest
}

// generateStringComparison implements string equality via strcmp
func (g *Generator) generateStringComparison(expr *ast.InfixExpr) ir.Value {
	left := g.generateExpression(expr.Left)
	right := g.generateExpression(expr.Right)

	// Call strcmp(left, right)
	cmpResult := g.newReg()
	g.emit(&ir.Call{
		Dest:     cmpResult,
		Func:     "strcmp",
		Args:     []ir.Value{left, right},
		ArgTypes: []types.Type{types.String, types.String},
		RetType:  types.Int,
	})

	// Compare result with 0
	dest := g.newReg()
	if expr.Operator == "==" {
		// strcmp returns 0 when equal, so: result == 0
		g.emit(&ir.BinOp{Dest: dest, Op: ir.OpEq, Type: types.Int, Left: cmpResult, Right: &ir.IntConst{Value: 0}})
	} else {
		// strcmp returns non-zero when not equal, so: result != 0
		g.emit(&ir.BinOp{Dest: dest, Op: ir.OpNeq, Type: types.Int, Left: cmpResult, Right: &ir.IntConst{Value: 0}})
	}

	return dest
}

// generateShortCircuitAnd implements short-circuit &&// If left is false, right is NOT evaluated
func (g *Generator) generateShortCircuitAnd(expr *ast.InfixExpr) ir.Value {
	// Evaluate left
	left := g.generateExpression(expr.Left)

	// Create blocks
	evalRightBlock := g.newBlock("and.right")
	mergeBlock := g.newBlock("and.merge")

	// If left is false, skip right evaluation
	g.curBlock.Term = &ir.CondBr{Cond: left, Then: evalRightBlock.Label, Else: mergeBlock.Label}
	leftBlock := g.curBlock

	// Evaluate right (only if left was true)
	g.curFunc.Blocks = append(g.curFunc.Blocks, evalRightBlock)
	g.curBlock = evalRightBlock
	right := g.generateExpression(expr.Right)
	rightBlock := g.curBlock
	g.curBlock.Term = &ir.Br{Target: mergeBlock.Label}

	// Merge block with phi
	g.curFunc.Blocks = append(g.curFunc.Blocks, mergeBlock)
	g.curBlock = mergeBlock

	dest := g.newReg()
	g.emit(&ir.Phi{
		Dest: dest,
		Type: types.Bool,
		Entries: []ir.PhiEntry{
			{Val: &ir.BoolConst{Value: false}, Block: leftBlock.Label}, // short-circuit: false
			{Val: right, Block: rightBlock.Label},                       // evaluated right
		},
	})

	return dest
}

// generateShortCircuitOr implements short-circuit ||// If left is true, right is NOT evaluated
func (g *Generator) generateShortCircuitOr(expr *ast.InfixExpr) ir.Value {
	// Evaluate left
	left := g.generateExpression(expr.Left)

	// Create blocks
	evalRightBlock := g.newBlock("or.right")
	mergeBlock := g.newBlock("or.merge")

	// If left is true, skip right evaluation
	g.curBlock.Term = &ir.CondBr{Cond: left, Then: mergeBlock.Label, Else: evalRightBlock.Label}
	leftBlock := g.curBlock

	// Evaluate right (only if left was false)
	g.curFunc.Blocks = append(g.curFunc.Blocks, evalRightBlock)
	g.curBlock = evalRightBlock
	right := g.generateExpression(expr.Right)
	rightBlock := g.curBlock
	g.curBlock.Term = &ir.Br{Target: mergeBlock.Label}

	// Merge block with phi
	g.curFunc.Blocks = append(g.curFunc.Blocks, mergeBlock)
	g.curBlock = mergeBlock

	dest := g.newReg()
	g.emit(&ir.Phi{
		Dest: dest,
		Type: types.Bool,
		Entries: []ir.PhiEntry{
			{Val: &ir.BoolConst{Value: true}, Block: leftBlock.Label}, // short-circuit: true
			{Val: right, Block: rightBlock.Label},                      // evaluated right
		},
	})

	return dest
}

func (g *Generator) generateCall(call *ast.CallExpr) ir.Value {
	// Get function name
	ident, ok := call.Function.(*ast.Identifier)
	if !ok {
		return &ir.IntConst{Value: 0}
	}

	// Handle built-in functions specially
	if result := g.generateBuiltinCall(ident.Value, call.Arguments); result != nil {
		return result
	}

	// Generate arguments
	args := make([]ir.Value, len(call.Arguments))
	for i, arg := range call.Arguments {
		args[i] = g.generateExpression(arg)
	}

	// Check if this is a call to a closure variable
	if closure, ok := g.closureVars[ident.Value]; ok {
		// Use ClosureCall to properly extract env from closure struct
		dest := g.newReg()
		g.emit(&ir.ClosureCall{Dest: dest, ClosurePtr: closure.envReg, Args: args, RetType: types.Int})
		return dest
	}

	// Check if this is a call to a simple lambda variable (no captures)
	if liftedName, ok := g.funcVars[ident.Value]; ok {
		// Call the lifted function directly
		dest := g.newReg()
		g.emit(&ir.Call{Dest: dest, Func: liftedName, Args: args, RetType: types.Int})
		return dest
	}

	// Check if this is a call through a parameter (higher-order function)
	if g.inFunction {
		for _, p := range g.curFunc.Params {
			if p.Name == ident.Value {
				// This is a call through a function pointer parameter
				// Load the function pointer and call indirectly
				ptr := g.locals[ident.Value]
				if ptr != nil {
					funcPtr := g.newReg()
					g.emit(&ir.Load{Dest: funcPtr, Type: types.Int, Ptr: ptr})
					dest := g.newReg()
					g.emit(&ir.CallIndirect{Dest: dest, FuncPtr: funcPtr, Args: args, RetType: types.Int})
					return dest
				}
			}
		}
	}

	// Check if this is a local variable with function type (closure returned from function)
	if ptr, ok := g.locals[ident.Value]; ok {
		if varType, hasType := g.varTypes[ident.Value]; hasType {
			if funcType, isFunc := varType.(*types.FunctionType); isFunc {
				// Load closure pointer and use ClosureCall
				closurePtr := g.newReg()
				g.emit(&ir.Load{Dest: closurePtr, Type: funcType, Ptr: ptr})
				dest := g.newReg()
				g.emit(&ir.ClosureCall{Dest: dest, ClosurePtr: closurePtr, Args: args, RetType: types.Int})
				return dest
			}
		}
	}

	// Look up actual return type from function declarations
	var retType types.Type = types.Int // default
	if rt, ok := g.funcReturnTypes[ident.Value]; ok {
		retType = rt
	}

	dest := g.newReg()
	g.emit(&ir.Call{Dest: dest, Func: ident.Value, Args: args, RetType: retType})

	return dest
}

// generateBuiltinCall handles built-in function calls
func (g *Generator) generateBuiltinCall(name string, args []ast.Expression) ir.Value {
	switch name {
	case "radio":
		// radio is overloaded - dispatch based on argument type
		if len(args) != 1 {
			return &ir.IntConst{Value: 0}
		}
		argType := g.inferExprType(args[0])
		argVal := g.generateExpression(args[0])

		var funcName string
		switch {
		case types.Equal(argType, types.String):
			funcName = "f1_radio_str"
		case types.Equal(argType, types.Bool):
			funcName = "f1_radio_bool"
		default:
			funcName = "f1_radio_int"
		}
		g.emit(&ir.Call{Dest: nil, Func: funcName, Args: []ir.Value{argVal}, ArgTypes: []types.Type{argType}, RetType: types.Void})
		return &ir.IntConst{Value: 0} // void return

	case "bono":
		if len(args) != 1 {
			return &ir.IntConst{Value: 0}
		}
		argVal := g.generateExpression(args[0])
		g.emit(&ir.Call{Dest: nil, Func: "f1_bono", Args: []ir.Value{argVal}, RetType: types.Void})
		return &ir.IntConst{Value: 0}

	case "canvas":
		if len(args) != 2 {
			return &ir.IntConst{Value: 0}
		}
		width := g.generateExpression(args[0])
		height := g.generateExpression(args[1])
		g.emit(&ir.Call{Dest: nil, Func: "f1_canvas", Args: []ir.Value{width, height}, RetType: types.Void})
		return &ir.IntConst{Value: 0}

	case "pixel":
		if len(args) != 5 {
			return &ir.IntConst{Value: 0}
		}
		pixelArgs := make([]ir.Value, 5)
		for i, arg := range args {
			pixelArgs[i] = g.generateExpression(arg)
		}
		g.emit(&ir.Call{Dest: nil, Func: "f1_pixel", Args: pixelArgs, RetType: types.Void})
		return &ir.IntConst{Value: 0}

	case "render":
		if len(args) != 1 {
			return &ir.IntConst{Value: 0}
		}
		argVal := g.generateExpression(args[0])
		g.emit(&ir.Call{Dest: nil, Func: "f1_render", Args: []ir.Value{argVal}, RetType: types.Void})
		return &ir.IntConst{Value: 0}

	case "snapshot":
		if len(args) != 1 {
			return &ir.IntConst{Value: 0}
		}
		argVal := g.generateExpression(args[0])
		g.emit(&ir.Call{Dest: nil, Func: "f1_snapshot", Args: []ir.Value{argVal}, RetType: types.Void})
		return &ir.IntConst{Value: 0}

	case "framedir":
		if len(args) != 1 {
			return &ir.IntConst{Value: 0}
		}
		argVal := g.generateExpression(args[0])
		g.emit(&ir.Call{Dest: nil, Func: "f1_framedir", Args: []ir.Value{argVal}, RetType: types.Void})
		return &ir.IntConst{Value: 0}

	default:
		return nil // Not a built-in
	}
}

// findFreeVariables performs free variable analysis on a lambda.
// Returns a list of variable names that are used but not defined in the lambda.
func (g *Generator) findFreeVariables(lambda *ast.LambdaExpr) []string {
	// Collect parameter names
	paramNames := make(map[string]bool)
	for _, p := range lambda.Params {
		paramNames[p.Name.Value] = true
	}

	// Find all variable references in the body
	freeVars := make(map[string]bool)
	localVars := make(map[string]bool)

	var walkExpr func(expr ast.Expression)
	var walkStmt func(stmt ast.Statement)

	walkExpr = func(expr ast.Expression) {
		if expr == nil {
			return
		}
		switch e := expr.(type) {
		case *ast.Identifier:
			name := e.Value
			// Skip if it's a parameter, local, or built-in
			if !paramNames[name] && !localVars[name] && !isBuiltin(name) {
				// Check if it exists in outer scope
				if _, ok := g.locals[name]; ok {
					freeVars[name] = true
				}
			}
		case *ast.InfixExpr:
			walkExpr(e.Left)
			walkExpr(e.Right)
		case *ast.PrefixExpr:
			walkExpr(e.Right)
		case *ast.CallExpr:
			walkExpr(e.Function)
			for _, arg := range e.Arguments {
				walkExpr(arg)
			}
		case *ast.GroupedExpr:
			walkExpr(e.Expression)
		case *ast.LambdaExpr:
			// Nested lambda - its free vars might be our free vars too
			// For now, skip nested lambdas (they'll be handled separately)
		}
	}

	walkStmt = func(stmt ast.Statement) {
		if stmt == nil {
			return
		}
		switch s := stmt.(type) {
		case *ast.VarDecl:
			walkExpr(s.Value)
			localVars[s.Name.Value] = true
		case *ast.Assignment:
			walkExpr(s.Value)
		case *ast.ReturnStmt:
			walkExpr(s.Value)
		case *ast.IfStmt:
			walkExpr(s.Condition)
			for _, inner := range s.Consequence.Statements {
				walkStmt(inner)
			}
			if s.Alternative != nil {
				if block, ok := s.Alternative.(*ast.BlockStmt); ok {
					for _, inner := range block.Statements {
						walkStmt(inner)
					}
				} else {
					walkStmt(s.Alternative)
				}
			}
		case *ast.ForStmt:
			walkStmt(s.Init)
			walkExpr(s.Condition)
			walkStmt(s.Update)
			for _, inner := range s.Body.Statements {
				walkStmt(inner)
			}
		case *ast.ExpressionStmt:
			walkExpr(s.Expression)
		case *ast.BlockStmt:
			for _, inner := range s.Statements {
				walkStmt(inner)
			}
		}
	}

	for _, stmt := range lambda.Body.Statements {
		walkStmt(stmt)
	}

	// Convert to sorted slice for deterministic order
	result := make([]string, 0, len(freeVars))
	for name := range freeVars {
		result = append(result, name)
	}
	return result
}

func isBuiltin(name string) bool {
	builtins := map[string]bool{
		"radio": true, "bono": true, "canvas": true,
		"pixel": true, "render": true, "snapshot": true,
	}
	return builtins[name]
}

// generateLambda generates IR for a lambda expression with closure support.
// It lifts the lambda to a top-level function and returns a closure if there are captures.
func (g *Generator) generateLambda(lambda *ast.LambdaExpr) ir.Value {
	// Generate unique name for lifted function
	funcName := fmt.Sprintf("__lambda_%d", g.lambdaCounter)
	g.lambdaCounter++

	// Find free variables
	freeVars := g.findFreeVariables(lambda)
	hasClosure := len(freeVars) > 0

	// Capture values from current scope BEFORE switching context
	// This emits loads in the outer function
	var capturedValues []ir.Value
	if hasClosure {
		for _, name := range freeVars {
			if ptr, ok := g.locals[name]; ok {
				// Load the value from the variable
				val := g.newReg()
				g.emit(&ir.Load{Dest: val, Type: types.Int, Ptr: ptr})
				capturedValues = append(capturedValues, val)
			}
		}
	}

	// Save current state AFTER capturing values (so regCounter includes the loads)
	savedFunc := g.curFunc
	savedBlock := g.curBlock
	savedLocals := g.locals
	savedRegCounter := g.regCounter
	savedBlockCounter := g.blockCounter
	savedInFunction := g.inFunction

	// Create the lifted function
	params := make([]*ir.Param, 0, len(lambda.Params)+1)

	// Add environment parameter if this is a closure
	if hasClosure {
		params = append(params, &ir.Param{Name: "__env", Type: types.Int}) // ptr type
	}

	// Add regular parameters
	for _, p := range lambda.Params {
		params = append(params, &ir.Param{Name: p.Name.Value, Type: types.Int})
	}

	// Determine return type
	returnType := g.inferReturnType(lambda.Body)

	g.curFunc = &ir.Function{
		Name:         funcName,
		Params:       params,
		ReturnType:   returnType,
		IsClosure:    hasClosure,
		CaptureNames: freeVars,
	}
	g.module.Functions = append(g.module.Functions, g.curFunc)

	g.curBlock = g.newBlock("entry")
	g.curFunc.Blocks = append(g.curFunc.Blocks, g.curBlock)

	g.locals = make(map[string]*ir.Reg)
	g.varTypes = make(map[string]types.Type)
	g.regCounter = 0
	g.blockCounter = 1 // entry is 0
	g.inFunction = true

	// Load captured variables from environment
	if hasClosure {
		for i, name := range freeVars {
			ptr := g.newReg()
			g.emit(&ir.Alloca{Dest: ptr, Type: types.Int})
			// Get value from environment (simulated as loading from env struct)
			envVal := g.newReg()
			g.emit(&ir.GetEnvField{Dest: envVal, Env: &ir.ParamRef{Name: "__env"}, Index: i, Type: types.Int})
			g.emit(&ir.Store{Type: types.Int, Val: envVal, Ptr: ptr})
			g.locals[name] = ptr
		}
	}

	// Create allocas for parameters
	for _, p := range lambda.Params {
		ptr := g.newReg()
		g.emit(&ir.Alloca{Dest: ptr, Type: types.Int})
		g.emit(&ir.Store{Type: types.Int, Val: &ir.ParamRef{Name: p.Name.Value}, Ptr: ptr})
		g.locals[p.Name.Value] = ptr
	}

	// Generate body
	for _, stmt := range lambda.Body.Statements {
		g.generateStatement(stmt)
	}

	// Add implicit return if needed
	if g.curBlock.Term == nil {
		if types.Equal(returnType, types.Void) {
			g.curBlock.Term = &ir.RetVoid{}
		}
	}

	// Restore state
	g.curFunc = savedFunc
	g.curBlock = savedBlock
	g.locals = savedLocals
	g.regCounter = savedRegCounter
	g.blockCounter = savedBlockCounter
	g.inFunction = savedInFunction

	// Return closure or function reference
	if hasClosure {
		dest := g.newReg()
		g.emit(&ir.MakeClosure{Dest: dest, FuncName: funcName, Captures: capturedValues})
		return dest
	}
	return &ir.FuncRef{Name: funcName}
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
