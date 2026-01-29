package typechecker

import (
	"fmt"

	"f1c/ast"
	"f1c/types"
)

// Checker performs type checking on an F1-Lang AST.
type Checker struct {
	errors []string
	env    *Environment
}

// New creates a new type checker.
func New() *Checker {
	return &Checker{
		errors: []string{},
		env:    NewEnvironment(),
	}
}

// Errors returns the list of type errors found during checking.
func (c *Checker) Errors() []string {
	return c.errors
}

// TypeOf returns the type of a variable in the current scope.
func (c *Checker) TypeOf(name string) types.Type {
	if typ, ok := c.env.Get(name); ok {
		return typ
	}
	return nil
}

func (c *Checker) addError(format string, args ...interface{}) {
	c.errors = append(c.errors, fmt.Sprintf(format, args...))
}

// Check type-checks a program.
func (c *Checker) Check(program *ast.Program) {
	for _, stmt := range program.Statements {
		c.checkStatement(stmt)
	}
}

func (c *Checker) checkStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.VarDecl:
		c.checkVarDecl(s)
	case *ast.Assignment:
		c.checkAssignment(s)
	case *ast.FunctionDecl:
		c.checkFunctionDecl(s)
	case *ast.ReturnStmt:
		c.checkReturnStmt(s)
	case *ast.IfStmt:
		c.checkIfStmt(s)
	case *ast.ForStmt:
		c.checkForStmt(s)
	case *ast.BlockStmt:
		c.checkBlockStmt(s)
	case *ast.ExpressionStmt:
		c.checkExpression(s.Expression)
	}
}

func (c *Checker) checkVarDecl(decl *ast.VarDecl) {
	// Check for redeclaration in same scope
	if c.env.ExistsInCurrentScope(decl.Name.Value) {
		c.addError("'%s' already declared in this scope", decl.Name.Value)
		return
	}

	// Infer type from initializer
	typ := c.checkExpression(decl.Value)
	if typ == nil {
		return
	}

	c.env.Set(decl.Name.Value, typ)
}

func (c *Checker) checkAssignment(assign *ast.Assignment) {
	// Check that variable exists
	varType, ok := c.env.Get(assign.Name.Value)
	if !ok {
		c.addError("undefined variable '%s'", assign.Name.Value)
		return
	}

	// Check that value type matches variable type
	valueType := c.checkExpression(assign.Value)
	if valueType == nil {
		return
	}

	if !types.Equal(varType, valueType) {
		c.addError("cannot assign %s to variable of type %s", valueType, varType)
	}
}

func (c *Checker) checkFunctionDecl(fn *ast.FunctionDecl) {
	// Check for redeclaration
	if c.env.ExistsInCurrentScope(fn.Name.Value) {
		c.addError("'%s' already declared in this scope", fn.Name.Value)
		return
	}

	// Create function type (params are all int for now - we infer from usage)
	paramTypes := make([]types.Type, len(fn.Params))
	for i := range fn.Params {
		paramTypes[i] = types.Int // Parameters default to int
	}

	// Enter function scope
	c.env = NewEnclosedEnvironment(c.env)

	// Add parameters to scope
	for i, param := range fn.Params {
		c.env.Set(param.Name.Value, paramTypes[i])
	}

	// Check function body and collect return types
	returnType := c.checkFunctionBody(fn.Body)

	// Exit function scope
	c.env = c.env.outer

	// Create and store function type
	fnType := &types.FunctionType{
		Params: paramTypes,
		Return: returnType,
	}
	c.env.Set(fn.Name.Value, fnType)
}

func (c *Checker) checkFunctionBody(body *ast.BlockStmt) types.Type {
	var returnTypes []types.Type

	// Collect return types from all statements (including nested)
	c.collectReturnTypes(body, &returnTypes)

	// Determine return type
	if len(returnTypes) == 0 {
		return types.Void
	}

	// Check all return types are consistent
	firstType := returnTypes[0]
	for i := 1; i < len(returnTypes); i++ {
		if !types.Equal(returnTypes[i], firstType) {
			c.addError("inconsistent return types: %s and %s", firstType, returnTypes[i])
		}
	}

	return firstType
}

func (c *Checker) collectReturnTypes(block *ast.BlockStmt, returnTypes *[]types.Type) {
	for _, stmt := range block.Statements {
		c.collectReturnTypesFromStmt(stmt, returnTypes)
	}
}

func (c *Checker) collectReturnTypesFromStmt(stmt ast.Statement, returnTypes *[]types.Type) {
	switch s := stmt.(type) {
	case *ast.ReturnStmt:
		if s.Value == nil {
			*returnTypes = append(*returnTypes, types.Void)
		} else {
			typ := c.checkExpression(s.Value)
			if typ != nil {
				*returnTypes = append(*returnTypes, typ)
			}
		}
	case *ast.IfStmt:
		c.checkExpression(s.Condition)
		c.collectReturnTypes(s.Consequence, returnTypes)
		if s.Alternative != nil {
			if block, ok := s.Alternative.(*ast.BlockStmt); ok {
				c.collectReturnTypes(block, returnTypes)
			} else if nested, ok := s.Alternative.(*ast.IfStmt); ok {
				c.collectReturnTypesFromStmt(nested, returnTypes)
			}
		}
	case *ast.ForStmt:
		c.collectReturnTypes(s.Body, returnTypes)
	case *ast.BlockStmt:
		c.collectReturnTypes(s, returnTypes)
	case *ast.VarDecl:
		c.checkVarDecl(s)
	case *ast.Assignment:
		c.checkAssignment(s)
	case *ast.ExpressionStmt:
		c.checkExpression(s.Expression)
	}
}

func (c *Checker) checkReturnStmt(ret *ast.ReturnStmt) {
	if ret.Value != nil {
		c.checkExpression(ret.Value)
	}
}

func (c *Checker) checkIfStmt(ifStmt *ast.IfStmt) {
	// Check condition is boolean
	condType := c.checkExpression(ifStmt.Condition)
	if condType != nil && !types.Equal(condType, types.Bool) {
		c.addError("condition must be bool, got %s", condType)
	}

	// Check consequence branch in new scope
	c.env = NewEnclosedEnvironment(c.env)
	c.checkBlockStmt(ifStmt.Consequence)
	c.env = c.env.outer

	// Check alternative branch if present
	if ifStmt.Alternative != nil {
		c.env = NewEnclosedEnvironment(c.env)
		c.checkStatement(ifStmt.Alternative)
		c.env = c.env.outer
	}
}

func (c *Checker) checkForStmt(forStmt *ast.ForStmt) {
	// Create new scope for loop
	c.env = NewEnclosedEnvironment(c.env)

	// Check init
	if forStmt.Init != nil {
		c.checkStatement(forStmt.Init)
	}

	// Check condition is boolean
	if forStmt.Condition != nil {
		condType := c.checkExpression(forStmt.Condition)
		if condType != nil && !types.Equal(condType, types.Bool) {
			c.addError("loop condition must be bool, got %s", condType)
		}
	}

	// Check update
	if forStmt.Update != nil {
		c.checkStatement(forStmt.Update)
	}

	// Check body
	c.checkBlockStmt(forStmt.Body)

	// Exit loop scope
	c.env = c.env.outer
}

func (c *Checker) checkBlockStmt(block *ast.BlockStmt) {
	for _, stmt := range block.Statements {
		c.checkStatement(stmt)
	}
}

func (c *Checker) checkExpression(expr ast.Expression) types.Type {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return types.Int
	case *ast.StringLiteral:
		return types.String
	case *ast.BooleanLiteral:
		return types.Bool
	case *ast.Identifier:
		return c.checkIdentifier(e)
	case *ast.PrefixExpr:
		return c.checkPrefixExpr(e)
	case *ast.InfixExpr:
		return c.checkInfixExpr(e)
	case *ast.CallExpr:
		return c.checkCallExpr(e)
	case *ast.GroupedExpr:
		return c.checkExpression(e.Expression)
	case *ast.LambdaExpr:
		return c.checkLambdaExpr(e)
	default:
		return nil
	}
}

func (c *Checker) checkIdentifier(ident *ast.Identifier) types.Type {
	typ, ok := c.env.Get(ident.Value)
	if !ok {
		c.addError("undefined variable '%s'", ident.Value)
		return nil
	}
	return typ
}

func (c *Checker) checkPrefixExpr(expr *ast.PrefixExpr) types.Type {
	rightType := c.checkExpression(expr.Right)
	if rightType == nil {
		return nil
	}

	switch expr.Operator {
	case "-":
		if !types.Equal(rightType, types.Int) {
			c.addError("operator - requires int operand, got %s", rightType)
			return nil
		}
		return types.Int
	case "!":
		if !types.Equal(rightType, types.Bool) {
			c.addError("operator ! requires bool operand, got %s", rightType)
			return nil
		}
		return types.Bool
	default:
		c.addError("unknown prefix operator: %s", expr.Operator)
		return nil
	}
}

func (c *Checker) checkInfixExpr(expr *ast.InfixExpr) types.Type {
	leftType := c.checkExpression(expr.Left)
	rightType := c.checkExpression(expr.Right)

	if leftType == nil || rightType == nil {
		return nil
	}

	switch expr.Operator {
	case "+", "-", "*", "/", "%":
		// Arithmetic operators require int operands
		if !types.Equal(leftType, types.Int) || !types.Equal(rightType, types.Int) {
			c.addError("operator %s requires int operands, got %s and %s",
				expr.Operator, leftType, rightType)
			return nil
		}
		return types.Int

	case "<", ">", "<=", ">=":
		// Comparison operators require int operands, return bool
		if !types.Equal(leftType, types.Int) || !types.Equal(rightType, types.Int) {
			c.addError("operator %s requires int operands, got %s and %s",
				expr.Operator, leftType, rightType)
			return nil
		}
		return types.Bool

	case "==", "!=":
		// Equality operators require same type operands
		if !types.Equal(leftType, rightType) {
			c.addError("operator %s requires same type operands, got %s and %s",
				expr.Operator, leftType, rightType)
			return nil
		}
		// Check if comparing functions (not allowed)
		if _, ok := leftType.(*types.FunctionType); ok {
			c.addError("cannot compare function types")
			return nil
		}
		return types.Bool

	case "&&", "||":
		// Logical operators require bool operands
		if !types.Equal(leftType, types.Bool) || !types.Equal(rightType, types.Bool) {
			c.addError("operator %s requires bool operands, got %s and %s",
				expr.Operator, leftType, rightType)
			return nil
		}
		return types.Bool

	default:
		c.addError("unknown operator: %s", expr.Operator)
		return nil
	}
}

func (c *Checker) checkCallExpr(call *ast.CallExpr) types.Type {
	// Check the function expression
	fnType := c.checkExpression(call.Function)
	if fnType == nil {
		return nil
	}

	fn, ok := fnType.(*types.FunctionType)
	if !ok {
		c.addError("cannot call non-function type %s", fnType)
		return nil
	}

	// Check argument count
	if len(call.Arguments) != len(fn.Params) {
		c.addError("wrong number of arguments: got %d, want %d",
			len(call.Arguments), len(fn.Params))
		return nil
	}

	// Check argument types
	for i, arg := range call.Arguments {
		argType := c.checkExpression(arg)
		if argType == nil {
			continue
		}
		if !types.Equal(argType, fn.Params[i]) {
			c.addError("argument %d: expected %s, got %s",
				i+1, fn.Params[i], argType)
		}
	}

	return fn.Return
}

func (c *Checker) checkLambdaExpr(lambda *ast.LambdaExpr) types.Type {
	// Create parameter types (default to int)
	paramTypes := make([]types.Type, len(lambda.Params))
	for i := range lambda.Params {
		paramTypes[i] = types.Int
	}

	// Enter lambda scope
	c.env = NewEnclosedEnvironment(c.env)

	// Add parameters to scope
	for i, param := range lambda.Params {
		c.env.Set(param.Name.Value, paramTypes[i])
	}

	// Check body and get return type
	returnType := c.checkFunctionBody(lambda.Body)

	// Exit lambda scope
	c.env = c.env.outer

	return &types.FunctionType{
		Params: paramTypes,
		Return: returnType,
	}
}
