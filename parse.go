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
func (Assign) isStmt()   {}
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
	// Function body.
	stmts := make([]Stmt, 0)
	for len(tokens) > 0 {
		stmts = append(stmts, stmt())
	}

	funcs := make([]Function, 0)
	funcs = append(funcs, Function{stmts, locals, 0})
	return funcs
}

func stmt() Stmt {
	// Return statement.
	if consume("return") {
		stmtN := Return{expr()}
		expect(";")
		return stmtN
	}

	// Block.
	if consume("{") {
		stmts := make([]Stmt, 0)
		for !consume("}") {
			stmts = append(stmts, stmt())
		}
		return Block{stmts}
	}

	// Assignment statement.
	exprN := equality()
	if consume("=") {
		stmtN := Assign{"=", exprN, expr()}
		expect(";")
		return stmtN
	}

	// Expression statement.
	expect(";")
	return ExprStmt{exprN}
}

func expr() Expr {
	return equality()
}

func equality() Expr {
	exprN := relational()

	for len(tokens) > 0 {
		if consume("==") {
			exprN = Binary{"==", exprN, relational()}
		} else if consume("!=") {
			exprN = Binary{"!=", exprN, relational()}
		} else {
			return exprN
		}
	}
	return exprN
}

func relational() Expr {
	exprN := add()

	for len(tokens) > 0 {
		if consume("<") {
			exprN = Binary{"<", exprN, add()}
		} else if consume("<=") {
			exprN = Binary{"<=", exprN, add()}
		} else if consume(">") {
			exprN = Binary{"<", add(), exprN}
		} else if consume(">=") {
			exprN = Binary{"<=", add(), exprN}
		} else {
			return exprN
		}
	}
	return exprN
}

func add() Expr {
	exprN := mul()

	for len(tokens) > 0 {
		if consume("+") {
			exprN = Binary{"+", exprN, mul()}
		} else if consume("-") {
			exprN = Binary{"-", exprN, mul()}
		} else {
			return exprN
		}
	}
	return exprN
}

func mul() Expr {
	exprN := unary()

	for len(tokens) > 0 {
		if consume("*") {
			exprN = Binary{"*", exprN, unary()}
		} else if consume("/") {
			exprN = Binary{"/", exprN, unary()}
		} else {
			return exprN
		}
	}
	return exprN
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
		exprN := expr()
		expect(")")
		return exprN
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
