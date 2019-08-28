package main

import (
	"fmt"
)

var tmpLocals []Var
var varOffset int = 8

// -------------------- Interfaces --------------------
// All declaration nodes must implement the Decl interface.
type Decl interface {
	isDecl()
}

// All statement nodes must implement the Stmt interface.
type Stmt interface {
	isStmt()
}

// All expression nodes must implement the Expr interface.
type Expr interface {
	isExpr()
}

// -------------------- Top level program --------------------
type Program struct {
	funcs []Function
}

// -------------------- Declarations --------------------
type Function struct {
	name      string
	params    []Var
	locals    []Var
	stmts     []Stmt
	stackSize int
}

type VarDecl struct { // It's also a statement.
	ident Var
	rval  Expr
}

func (Function) isDecl() {}
func (VarDecl) isDecl()  {}

// -------------------- Statements --------------------
type Unary struct { // It's also an expression.
	child Expr
}

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

type Assign struct {
	lval Expr
	rval Expr
}

type Empty struct{} // It's also an expression
type Return Unary
type ExprStmt Unary

func (VarDecl) isStmt()  {}
func (Assign) isStmt()   {}
func (Return) isStmt()   {}
func (ExprStmt) isStmt() {}
func (Block) isStmt()    {}
func (If) isStmt()       {}
func (For) isStmt()      {}
func (Empty) isStmt()    {}

// -------------------- Expressions --------------------
type Binary struct {
	op  string
	lhs Expr
	rhs Expr
}

type FuncCall struct {
	name string
	args []Expr
}

type Var struct {
	name   string
	offset int
	ty     *Type
}

type IntLit struct {
	val int
	ty  *Type
}

type Addr Unary
type Deref Unary

func (Binary) isExpr()   {}
func (FuncCall) isExpr() {}
func (Var) isExpr()      {}
func (Addr) isExpr()     {}
func (Deref) isExpr()    {}
func (IntLit) isExpr()   {}
func (Empty) isExpr()    {}

func findVar(name string) *Var {
	for _, v := range tmpLocals {
		if v.name == name {
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

func consumeType() *Token {
	if len(tokens) > 0 && tokens[0].kind == TK_TYPE {
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
	panic(fmt.Sprintf("tokens: %s\nExpected %s but got %s\n", tokens, op, tokens[0].str))
}

func next(op string) bool {
	return len(tokens) != 0 && tokens[0].str == op
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
			panic(fmt.Sprintf("tokens: %s\nExpected an identifier inside function parameters but got %#v\n", tokens, tok))
		}

		tokTy := consumeType()
		if tokTy != nil && !supportType(tokTy.str) {
			panic(fmt.Sprintf("Unsupported type %s\n", tokTy.str))
		}

		v := Var{tok.str, varOffset, &Type{tokTy.str}}
		varOffset += 8
		tmpLocals = append(tmpLocals, v)
		params = append(params, v)

		if !consume(",") {
			break
		}
	}
	return params
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
		panic(fmt.Sprintf("tokens: %s\nExpected an identifier after 'func' keyword but got %#v\n", tokens, tok))
	}
	name := tok.str
	assert("(")
	params := funcParams()
	assert(")")

	stmts := make([]Stmt, 0)
	for len(tokens) > 0 && !next("func") {
		s := stmt()
		addType(s)
		stmts = append(stmts, s)
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

func simpleStmt(exprN Expr) Stmt {
	switch exprN.(type) {
	case Empty:
		return Empty{}
	}

	// Identifier declaration.
	if consume(":=") {
		v := exprN.(Var)

		varp := findVar(v.name)
		if varp != nil {
			panic(fmt.Sprintf("%s is already declared. No new variables on left side of := \n", v.name))
		}
		varOffset += 8
		tmpLocals = append(tmpLocals, v)
		return VarDecl{v, expr()}
	}

	// Assignment statement.
	if consume("=") {
		switch v := exprN.(type) {
		case Var:
			varp := findVar(v.name)
			if varp == nil {
				panic(fmt.Sprintf("Undefined: %s\n", v.name))
			}
		}
		return Assign{exprN, expr()}
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
		return Binary{"-", IntLit{0, &Type{"int"}}, unary()} // -val = 0 - val
	} else if consume("&") {
		return Addr{unary()}
	} else if consume("*") {
		return Deref{unary()}
	}
	return primary()
}

func primary() Expr {
	// Operand "()".
	if consume("(") {
		exprN := expr()
		assert(")")
		return exprN
	}

	// Identifier.
	tok := consumeIdent()
	if tok != nil {
		// Function call.
		if consume("(") {
			return FuncCall{tok.str, funcArgs()}
		}

		// Variable.
		// Not register to `tmpLocals` yet.
		varp := findVar(tok.str)
		if varp == nil {
			return Var{tok.str, varOffset, &Type{"none"}}
		}
		return *varp
	}

	// Integer literal.
	n := IntLit{tokens[0].val, &Type{"int"}}
	tokens = tokens[1:]
	return n
}

// -------------------- Debug functions. --------------------
func printNodes(funcs []Function) {
	for i, f := range funcs {
		fmt.Println("")
		fmt.Println("[Function]", i, f.name)
		for i, s := range f.stmts {
			fmt.Println("")
			fmt.Println("[Statements]", i)
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
	case Empty:
		fmt.Printf("Empty dep: %d\n", dep)
	case IntLit:
		fmt.Printf("IntLit dep: %d, val: %d, type: %#v\n", dep, n.val, n.ty)
	case Var:
		fmt.Printf("Var dep: %d, name: %s, offset: %d, addr: %p, type: %d\n", dep, n.name, n.offset, &n, n.ty)
	case VarDecl:
		fmt.Printf("VarDecl dep: %d\n", dep)
		printNode(n.ident, dep+1)
		printNode(n.rval, dep+1)
	case Assign:
		fmt.Printf("Assign dep: %d\n", dep)
		printNode(n.lval, dep+1)
		printNode(n.rval, dep+1)
	case Addr:
		fmt.Printf("Addr dep: %d\n", dep)
		printNode(n.child, dep+1)
	case Deref:
		fmt.Printf("Deref dep: %d\n", dep)
		printNode(n.child, dep+1)
	case Block:
		fmt.Printf("Block dep: %d\n", dep)
		for _, c := range n.children {
			printNode(c, dep)
		}
	case ExprStmt:
		fmt.Printf("ExprStmt dep: %d\n", dep)
		printNode(n.child, dep+1)
	case If:
		fmt.Printf("If dep: %d\n", dep, n)
		printNode(n.init, dep)
		printNode(n.cond, dep)
		printNode(n.then, dep)
		printNode(n.els, dep)
	case For:
		fmt.Printf("For dep: %d %#v\n", dep, n)
		printNode(n.init, dep)
		printNode(n.cond, dep)
		printNode(n.post, dep)
		printNode(n.then, dep)
	case Return:
		fmt.Printf("Return dep: %d %#v\n", dep, n)
		printNode(n.child, dep+1)
	case FuncCall:
		fmt.Printf("FuncCall dep: %d\n", dep)
		for _, arg := range n.args {
			printNode(arg, dep+1)
		}
	case Binary:
		fmt.Printf("Binary dep: %d\n", dep)
		printNode(n.lhs, dep+1)
		printNode(n.rhs, dep+1)
	}
}
