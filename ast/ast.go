package ast

import (
	"bytes"
	"fmt"
	"strings"

	"f1c/token"
)

// Node is the interface for all AST nodes.
type Node interface {
	TokenLiteral() string
	String() string
}

// Statement is the interface for statement nodes.
type Statement interface {
	Node
	statementNode()
}

// Expression is the interface for expression nodes.
type Expression interface {
	Node
	expressionNode()
}

// Program is the root node of every AST.
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

// =============================================================================
// Statements
// =============================================================================

// VarDecl represents: driver <name> = <value>;
type VarDecl struct {
	Token token.Token // the DRIVER token
	Name  *Identifier
	Value Expression
}

func (v *VarDecl) statementNode()       {}
func (v *VarDecl) TokenLiteral() string { return v.Token.Literal }

// Assignment represents: <name> = <value>;
type Assignment struct {
	Token token.Token // the IDENT token
	Name  *Identifier
	Value Expression
}

func (a *Assignment) statementNode()       {}
func (a *Assignment) TokenLiteral() string { return a.Token.Literal }

// ExpressionStmt represents an expression used as a statement.
type ExpressionStmt struct {
	Token      token.Token // the first token of the expression
	Expression Expression
}

func (e *ExpressionStmt) statementNode()       {}
func (e *ExpressionStmt) TokenLiteral() string { return e.Token.Literal }

// BlockStmt represents: { <statements> }
type BlockStmt struct {
	Token      token.Token // the LBRACE token
	Statements []Statement
}

func (b *BlockStmt) statementNode()       {}
func (b *BlockStmt) TokenLiteral() string { return b.Token.Literal }

// IfStmt represents: drs (<cond>) <then> [defend <else>]
type IfStmt struct {
	Token       token.Token // the DRS token
	Condition   Expression
	Consequence *BlockStmt
	Alternative Statement // can be *BlockStmt or *IfStmt (for else-if chains)
}

func (i *IfStmt) statementNode()       {}
func (i *IfStmt) TokenLiteral() string { return i.Token.Literal }

// ForStmt represents: lap (driver <init> = <val>; <cond>; <update>) <body>
type ForStmt struct {
	Token     token.Token // the LAP token
	Init      *VarDecl    // loop variable declaration
	Condition Expression
	Update    *Assignment
	Body      *BlockStmt
}

func (f *ForStmt) statementNode()       {}
func (f *ForStmt) TokenLiteral() string { return f.Token.Literal }

// ReturnStmt represents: finish [<value>];
type ReturnStmt struct {
	Token token.Token // the FINISH token
	Value Expression  // nil for void return
}

func (r *ReturnStmt) statementNode()       {}
func (r *ReturnStmt) TokenLiteral() string { return r.Token.Literal }

// FunctionDecl represents: pitstop <name>(<params>) <body>
type FunctionDecl struct {
	Token  token.Token   // the PITSTOP token
	Name   *Identifier
	Params []*Parameter
	Body   *BlockStmt
}

func (f *FunctionDecl) statementNode()       {}
func (f *FunctionDecl) TokenLiteral() string { return f.Token.Literal }

// Parameter represents a function parameter: driver <name>
type Parameter struct {
	Token token.Token // the DRIVER token
	Name  *Identifier
}

func (p *Parameter) TokenLiteral() string { return p.Token.Literal }
func (p *Parameter) String() string        { return "driver " + p.Name.Value }

// =============================================================================
// Expressions
// =============================================================================

// Identifier represents a variable or function name.
type Identifier struct {
	Token token.Token // the IDENT token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }

// IntegerLiteral represents an integer value.
type IntegerLiteral struct {
	Token token.Token // the INT token
	Value int64
}

func (i *IntegerLiteral) expressionNode()      {}
func (i *IntegerLiteral) TokenLiteral() string { return i.Token.Literal }

// StringLiteral represents a string value.
type StringLiteral struct {
	Token token.Token // the STRING token
	Value string
}

func (s *StringLiteral) expressionNode()      {}
func (s *StringLiteral) TokenLiteral() string { return s.Token.Literal }

// BooleanLiteral represents greenlight or redlight.
type BooleanLiteral struct {
	Token token.Token // GREENLIGHT or REDLIGHT
	Value bool
}

func (b *BooleanLiteral) expressionNode()      {}
func (b *BooleanLiteral) TokenLiteral() string { return b.Token.Literal }

// PrefixExpr represents: <op><expr> (e.g., -5, !true)
type PrefixExpr struct {
	Token    token.Token // the prefix token (e.g., !, -)
	Operator string
	Right    Expression
}

func (p *PrefixExpr) expressionNode()      {}
func (p *PrefixExpr) TokenLiteral() string { return p.Token.Literal }

// InfixExpr represents: <left> <op> <right>
type InfixExpr struct {
	Token    token.Token // the operator token
	Left     Expression
	Operator string
	Right    Expression
}

func (i *InfixExpr) expressionNode()      {}
func (i *InfixExpr) TokenLiteral() string { return i.Token.Literal }

// CallExpr represents: <function>(<args>)
type CallExpr struct {
	Token     token.Token // the LPAREN token
	Function  Expression  // Identifier or Lambda
	Arguments []Expression
}

func (c *CallExpr) expressionNode()      {}
func (c *CallExpr) TokenLiteral() string { return c.Token.Literal }

// GroupedExpr represents: (<expr>)
type GroupedExpr struct {
	Token      token.Token // the LPAREN token
	Expression Expression
}

func (g *GroupedExpr) expressionNode()      {}
func (g *GroupedExpr) TokenLiteral() string { return g.Token.Literal }

// LambdaExpr represents: pitstop(<params>) <body>
type LambdaExpr struct {
	Token  token.Token // the PITSTOP token
	Params []*Parameter
	Body   *BlockStmt
}

func (l *LambdaExpr) expressionNode()      {}
func (l *LambdaExpr) TokenLiteral() string { return l.Token.Literal }

// =============================================================================
// String() methods for debugging and testing
// =============================================================================

func (p *Program) String() string {
	var out bytes.Buffer
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

func (v *VarDecl) String() string {
	return fmt.Sprintf("driver %s = %s;", v.Name.String(), v.Value.String())
}

func (a *Assignment) String() string {
	return fmt.Sprintf("%s = %s;", a.Name.String(), a.Value.String())
}

func (e *ExpressionStmt) String() string {
	if e.Expression != nil {
		return e.Expression.String()
	}
	return ""
}

func (b *BlockStmt) String() string {
	var out bytes.Buffer
	out.WriteString("{ ")
	for _, s := range b.Statements {
		out.WriteString(s.String())
	}
	out.WriteString(" }")
	return out.String()
}

func (i *IfStmt) String() string {
	var out bytes.Buffer
	out.WriteString("drs (")
	out.WriteString(i.Condition.String())
	out.WriteString(") ")
	out.WriteString(i.Consequence.String())
	if i.Alternative != nil {
		out.WriteString(" defend ")
		out.WriteString(i.Alternative.(interface{ String() string }).String())
	}
	return out.String()
}

func (f *ForStmt) String() string {
	var out bytes.Buffer
	out.WriteString("lap (")
	out.WriteString(f.Init.String())
	out.WriteString(" ")
	out.WriteString(f.Condition.String())
	out.WriteString("; ")
	out.WriteString(f.Update.Name.String())
	out.WriteString(" = ")
	out.WriteString(f.Update.Value.String())
	out.WriteString(") ")
	out.WriteString(f.Body.String())
	return out.String()
}

func (r *ReturnStmt) String() string {
	if r.Value != nil {
		return fmt.Sprintf("finish %s;", r.Value.String())
	}
	return "finish;"
}

func (f *FunctionDecl) String() string {
	var out bytes.Buffer
	out.WriteString("pitstop ")
	out.WriteString(f.Name.String())
	out.WriteString("(")
	params := []string{}
	for _, p := range f.Params {
		params = append(params, "driver "+p.Name.String())
	}
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(f.Body.String())
	return out.String()
}

func (i *Identifier) String() string {
	return i.Value
}

func (i *IntegerLiteral) String() string {
	return i.Token.Literal
}

func (s *StringLiteral) String() string {
	return fmt.Sprintf("%q", s.Value)
}

func (b *BooleanLiteral) String() string {
	return b.Token.Literal
}

func (p *PrefixExpr) String() string {
	return fmt.Sprintf("(%s%s)", p.Operator, p.Right.String())
}

func (i *InfixExpr) String() string {
	return fmt.Sprintf("(%s %s %s)", i.Left.String(), i.Operator, i.Right.String())
}

func (c *CallExpr) String() string {
	var out bytes.Buffer
	out.WriteString(c.Function.String())
	out.WriteString("(")
	args := []string{}
	for _, a := range c.Arguments {
		args = append(args, a.String())
	}
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")
	return out.String()
}

func (g *GroupedExpr) String() string {
	return fmt.Sprintf("(%s)", g.Expression.String())
}

func (l *LambdaExpr) String() string {
	var out bytes.Buffer
	out.WriteString("pitstop(")
	params := []string{}
	for _, p := range l.Params {
		params = append(params, "driver "+p.Name.String())
	}
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(l.Body.String())
	return out.String()
}
