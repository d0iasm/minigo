package main

import (
	"fmt"
	"os"
)

var locals []*Var

type Expr interface {
	isExpr()
}

type Stmt interface {
	isStmt()
}

type Function struct {
	stmts     []Stmt
	locals    []*Var
	stackSize int
}

type Binary struct {
	op  string
	lhs Expr
	rhs Expr
}

type Assign Binary

type Unary struct {
	child Expr
}

type Return Unary
type ExprStmt Unary

type Block struct {
	children []Stmt
}

type Var struct {
	name   string // Variable name
	offset int    // Offset from RBP
}

type IntLit int

// Expressions.
func (Binary) isExpr() {}
func (Var) isExpr()    {}
func (IntLit) isExpr() {}

// Statements.
func (Assign) isExpr()   {} // TODO: expr -> stmt
func (Return) isStmt()   {}
func (ExprStmt) isStmt() {}
func (Block) isStmt()    {}

func findVar(tok Token) *Var {
	for _, v := range locals {
		if v.name == tok.str {
			return v
		}
	}
	return nil
}

func consume(op string) bool {
	if tokens[0].str == op {
		tokens = tokens[1:]
		return true
	}
	return false
}

func expect(op string) {
	if tokens[0].str == op {
		tokens = tokens[1:]
		return
	}
	fmt.Println(tokens)
	fmt.Printf("[Error] expected %s but got %s\n", op, tokens[0].str)
	os.Exit(1)
}

func program() []Function {
	stmts := make([]Stmt, 0)
	for len(tokens) > 0 {
		stmts = append(stmts, stmt())
	}

	funcs := make([]Function, 0)
	funcs = append(funcs, Function{stmts, locals, 0})
	return funcs
}

func stmt() Stmt {
	if consume("return") {
		stmt := Return{expr()}
		expect(";")
		return stmt
	}

	if consume("{") {
		stmts := make([]Stmt, 0)
		for !consume("}") {
			stmts = append(stmts, stmt())
		}
		return Block{stmts}
	}
	stmt := ExprStmt{expr()}
	expect(";")
	return stmt
}

func expr() Expr {
	return assign()
}

// TODO: Expr -> Stmt
func assign() Expr {
	expr := equality()
	if consume("=") {
		expr = Assign{"=", expr, assign()}
	}
	return expr
}

func equality() Expr {
	expr := relational()

	for len(tokens) > 0 {
		if consume("==") {
			expr = Binary{"==", expr, relational()}
		} else if consume("!=") {
			expr = Binary{"!=", expr, relational()}
		} else {
			return expr
		}
	}
	return expr
}

func relational() Expr {
	expr := add()

	for len(tokens) > 0 {
		if consume("<") {
			expr = Binary{"<", expr, add()}
		} else if consume("<=") {
			expr = Binary{"<=", expr, add()}
		} else if consume(">") {
			expr = Binary{"<", add(), expr}
		} else if consume(">=") {
			expr = Binary{"<=", add(), expr}
		} else {
			return expr
		}
	}
	return expr
}

func add() Expr {
	expr := mul()

	for len(tokens) > 0 {
		if consume("+") {
			expr = Binary{"+", expr, mul()}
		} else if consume("-") {
			expr = Binary{"-", expr, mul()}
		} else {
			return expr
		}
	}
	return expr
}

func mul() Expr {
	expr := unary()

	for len(tokens) > 0 {
		if consume("*") {
			expr = Binary{"*", expr, unary()}
		} else if consume("/") {
			expr = Binary{"/", expr, unary()}
		} else {
			return expr
		}
	}
	return expr
}

func unary() Expr {
	if consume("+") {
		return unary()
	} else if consume("-") {
		return Binary{"-", IntLit(0), unary()} // -val = 0 - val
	}
	return primary()
}

func primary() Expr {
	if consume("(") {
		expr := expr()
		expect(")")
		return expr
	}

	// Identifiers
	if tokens[0].kind == TK_IDENT {
		varp := findVar(tokens[0])
		if varp == nil {
			varp = &Var{tokens[0].str, 0}
			locals = append(locals, varp)
		}
		tokens = tokens[1:]
		return *varp
	}

	// Integer literals
	n := IntLit(tokens[0].val)
	tokens = tokens[1:]
	return n
}

/**
func printNodes(nodes []interface{}) {
	for i, n := range nodes {
		fmt.Println("[Print Node] node:", i)
		printNode(n, 0)
	}
}

func printNode(node interface{}, dep int) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case int:
		fmt.Printf("INT dep: %d, val: %d\n", dep, n)
	case Var:
		fmt.Printf("VAR dep: %d, name: %s, offset: %d\n", dep, n.name, n.offset)
	case Block:
		fmt.Printf("BLOCK dep: %d\n", dep)
		for _, c := range n.children {
			printNode(c, dep)
		}
	case Unary:
		fmt.Printf("UNARY dep: %d\n", dep)
		printNode(n.child, dep+1)
	case Binary:
		fmt.Printf("BINARY dep: %d\n", dep)
		printNode(n.lhs, dep+1)
		printNode(n.rhs, dep+1)
	}
}
*/
