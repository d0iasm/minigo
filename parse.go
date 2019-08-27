package main

import (
	"fmt"
)

var tmpLocals []Var
var varOffset int = 8

type Expr interface {
	isExpr()
}

type Stmt interface {
	isStmt()
}

type Program struct {
	funcs []Function
}

type Function struct {
	name      string
	params    []Var
	locals    []Var
	stmts     []Stmt
	stackSize int
}

type FuncCall struct {
	name string
	args []Expr
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
	init Stmt
	cond Expr
	post Stmt
	then Stmt
}

type Var struct {
	name   string // Variable name
	offset int    // Offset from RBP
}

type Empty struct{}

type IntLit int

// Expressions.
func (FuncCall) isExpr() {}
func (Binary) isExpr()   {}
func (Var) isExpr()      {}
func (IntLit) isExpr()   {}
func (Empty) isExpr()    {}

// Statements.
func (Assign) isStmt()   {}
func (Return) isStmt()   {}
func (ExprStmt) isStmt() {}
func (Block) isStmt()    {}
func (If) isStmt()       {}
func (For) isStmt()      {}
func (Empty) isStmt()    {}

func findVar(tok Token) *Var {
	for _, v := range tmpLocals {
		if v.name == tok.str {
			return &v
		}
	}
	return nil
}

func consume(op string) bool {
	if len(tokens) > 0 && tokens[0].str == op {
		tokens = tokens[1:]
		return true
	}
	return false
}

func consumeIdent() *Token {
	if len(tokens) > 0 && tokens[0].kind == TK_IDENT {
		tok := tokens[0]
		tokens = tokens[1:]
		return &tok
	}
	return nil
}

func assert(op string) {
	if len(tokens) > 0 && tokens[0].str == op {
		tokens = tokens[1:]
		return
	}
	panic(fmt.Sprintf("tokens: %s\n[Error] expected %s but got %s\n", tokens, op, tokens[0].str))
}

func next(op string) bool {
	return len(tokens) != 0 && tokens[0].str == op
}

func program() Program {
	funcs := make([]Function, 0)
	for len(tokens) > 0 {
		funcs = append(funcs, function())
	}
	return Program{funcs}
}

func function() Function {
	// Initialize for a function.
	varOffset = 8
	tmpLocals = make([]Var, 0)

	assert("func")
	tok := consumeIdent()
	if tok == nil {
		panic("Expect an identifier after 'func' keyword.")
	}
	name := tok.str
	assert("(")
	params := funcParams()
	assert(")")

	stmts := make([]Stmt, 0)
	for len(tokens) > 0 && !next("func") {
		stmts = append(stmts, stmt())
	}
	return Function{name, params, tmpLocals, stmts, len(tmpLocals) * 8}
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
		init, cond, post := forHeaders()
		return For{init, cond, post, stmt()}
	}

	return simpleStmt(expr())
}

func ifHeaders() (Stmt, Expr) {
	s1 := Stmt(nil)
	e1 := expr()
	if !next("{") {
		s1 = simpleStmt(e1)
		consume(";")
		e1 = expr()
	}
	return s1, e1
}

func forHeaders() (Stmt, Expr, Stmt) {
	s1 := Stmt(nil)
	e1 := Expr(nil)
	s2 := Stmt(nil)
	// No options.
	if next("{") {
		return s1, e1, s2
	}
	// Condition.
	e1 = expr()
	// For clause ([init] ; [cond] ; [post]).
	if !next("{") {
		s1 = simpleStmt(e1)
		consume(";") // No semi colon when e1 is Empty.
		e1 = expr()
	}
	if !next("{") {
		s2 = simpleStmt(expr())
		consume(";") // No semi colon when expr() is Empty.
	}
	return s1, e1, s2
}

func simpleStmt(exprN Expr) Stmt {
	switch exprN.(type) {
	case Empty:
		return Empty{}
	}

	// Assignment statement.
	if consume("=") {
		return Assign{"=", exprN, expr()}
	}

	// Expression statement.
	return ExprStmt{exprN}
}

func expr() Expr {
	if consume(";") {
		return Empty{}
	}
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

	// Identifiers.
	tok := consumeIdent()
	if tok != nil {
		// Function call.
		if consume("(") {
			return FuncCall{tok.str, funcArgs()}
		}

		// Variables.
		varp := findVar(*tok)
		if varp == nil {
			varp = &Var{tok.str, varOffset}
			varOffset += 8
			tmpLocals = append(tmpLocals, *varp)
		}
		return *varp
	}

	// Integer literals.
	n := IntLit(tokens[0].val)
	tokens = tokens[1:]
	return n
}

func funcArgs() []Expr {
	args := make([]Expr, 0)
	if consume(")") {
		return args
	}

	args = append(args, expr())
	for consume(",") {
		args = append(args, expr())
	}
	assert(")")
	return args
}

func funcParams() []Var {
	params := make([]Var, 0)
	if next(")") {
		return params
	}

	for {
		tok := consumeIdent()
		if tok == nil {
			panic("Expect an identifier inside function parameters.")
		}

		v := Var{tok.str, varOffset}
		varOffset += 8
		tmpLocals = append(tmpLocals, v)
		params = append(params, v)

		if !consume(",") {
			break
		}
	}
	return params
}

func printNodes(funcs []Function) {
	for i, f := range funcs {
		fmt.Println("[Function]:", i, f.name)
		for i, s := range f.stmts {
			fmt.Println("[Print Node] node:", i)
			printNode(s, 0)
		}
		fmt.Println("========================")
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
		fmt.Printf("FOR dep: %d %#v\n", dep, n)
		printNode(n.init, dep)
		printNode(n.cond, dep)
		printNode(n.post, dep)
		printNode(n.then, dep)
	case Return:
		fmt.Printf("RETURN dep: %d %#v\n", dep, n)
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
