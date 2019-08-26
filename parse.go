package main

import (
	"fmt"
)

var locals []Var
var varOffset int = 8

type Expr interface {
	isExpr()
}

type Stmt interface {
	isStmt()
}

type Function struct {
	stmts     []Stmt
	locals    []Var
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

type If struct {
	init Stmt
	cond Expr
	then Stmt
	els  Stmt
}

type For struct {
	//init Stmt // TODO: implement
	cond Expr
	//post Stmt // TODO: implement
	then Stmt
}

type Var struct {
	name   string // Variable name
	offset int    // Offset from RBP
}

type Empty struct{}

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
func (If) isStmt()       {}
func (For) isStmt()      {}
func (Empty) isStmt()    {}

func findVar(tok Token) *Var {
	for _, v := range locals {
		if v.name == tok.str {
			return &v
		}
	}
	return nil
}

func consume(op string) bool {
	if len(tokens) != 0 && tokens[0].str == op {
		tokens = tokens[1:]
		return true
	}
	return false
}

func assert(op string) {
	if len(tokens) != 0 && tokens[0].str == op {
		tokens = tokens[1:]
		return
	}
	panic(fmt.Sprintf("%s \n [Error] expected %s but got %s\n", tokens, op, tokens[0].str))
}

func next(op string) bool {
	return len(tokens) != 0 && tokens[0].str == op
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
		assert(";")
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

	// If statement.
	if consume("if") {
		init, cond := ifHeaders()
		ifstmt := If{init, cond, nil, nil}
		if next("{") {
			ifstmt.then = stmt()
		} else {
			assert("{")
		}
		if consume("else") {
			if next("if") || next("{") {
				ifstmt.els = stmt()
			} else {
				assert("if / {")
			}
		}
		return ifstmt
	}

	// For statement.
	if consume("for") {
		forstmt := For{nil, nil}
		if tokens[0].str != "{" {
			forstmt.cond = expr()
		}
		forstmt.then = stmt()
		return forstmt
	}

	return simpleStmt(expr())
}

func ifHeaders() (Stmt, Expr) {
	stmtN := Stmt(nil)
	exprN := expr()
	if !next("{") {
		stmtN = simpleStmt(exprN)
		exprN = expr()
	}
	return stmtN, exprN
}

func simpleStmt(exprN Expr) Stmt {
	// Assignment statement.
	if consume("=") {
		stmtN := Assign{"=", exprN, expr()}
		assert(";")
		return stmtN
	}

	// Expression statement.
	if consume(";") {
		return ExprStmt{exprN}
	}
	return Empty{}
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
		assert(")")
		return exprN
	}

	// Identifiers
	if tokens[0].kind == TK_IDENT {
		varp := findVar(tokens[0])
		if varp == nil {
			varp = &Var{tokens[0].str, varOffset}
			varOffset += 8
			locals = append(locals, *varp)
		}
		tokens = tokens[1:]
		return *varp
	}

	// Integer literals
	n := IntLit(tokens[0].val)
	tokens = tokens[1:]
	return n
}

func printNodes(stmts []Stmt) {
	for i, s := range stmts {
		fmt.Println("[Print Node] node:", i)
		printNode(s, 0)
	}
}

func printNode(node interface{}, dep int) {
	if node == nil {
		fmt.Printf("nil, dep: %d\n", dep)
		return
	}

	switch n := node.(type) {
	case IntLit:
		fmt.Printf("INT dep: %d, val: %d\n", dep, n)
	case Var:
		fmt.Printf("VAR dep: %d, name: %s, offset: %d, addr: %p\n", dep, n.name, n.offset, &n)
	case Block:
		fmt.Printf("BLOCK dep: %d\n", dep)
		for _, c := range n.children {
			printNode(c, dep)
		}
	case If:
		fmt.Printf("If dep: %d\n", dep, n)
		printNode(n.init, dep)
		printNode(n.cond, dep)
		printNode(n.then, dep)
		printNode(n.els, dep)
	case For:
		fmt.Printf("FOR dep: %d %v\n", dep, n)
		printNode(n.cond, dep)
		printNode(n.then, dep)
	case Return:
		fmt.Printf("RETURN dep: %d %v\n", dep, n)
		printNode(n.child, dep+1)
	case ExprStmt:
		fmt.Printf("ExprStmt dep: %d\n", dep)
		printNode(n.child, dep+1)
	case Binary:
		fmt.Printf("BINARY dep: %d\n", dep)
		printNode(n.lhs, dep+1)
		printNode(n.rhs, dep+1)
	case Assign:
		fmt.Printf("ASSIGN dep: %d\n", dep)
		printNode(n.lhs, dep+1)
		printNode(n.rhs, dep+1)
	}
}
