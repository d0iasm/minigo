package main

import (
	"fmt"
)

var globals []Var
var tmpLocals []Var
var varOffset int = 8

var contents []StringLit
var contentCnt = 0

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
	getType() *Type
	setType(ty Type)
}

// -------------------- Top level program --------------------
type Program struct {
	globals  []Var
	contents []StringLit
	funcs    []Function
}

// -------------------- Declarations --------------------
type Function struct {
	name      string
	params    []Var
	locals    []Var
	stmts     []Stmt
	stackSize int
}

func (Function) isDecl() {}

// -------------------- Statements --------------------
type ExprStmt struct {
	child Expr
}

type Return struct {
	//children []Expr
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
	lvals []Expr
	rvals []Expr
}

type Empty struct{} // It's also an expression

func (Assign) isStmt()   {}
func (Return) isStmt()   {}
func (ExprStmt) isStmt() {}
func (Block) isStmt()    {}
func (If) isStmt()       {}
func (For) isStmt()      {}
func (Empty) isStmt()    {}

// -------------------- Expressions --------------------
type IntLit struct {
	val int
	ty  *Type
}

type StringLit struct {
	val   string
	label string
	ty    *Type
}

type Unary struct {
	child Expr
	ty    *Type
}

type Binary struct {
	op  string
	lhs Expr
	rhs Expr
	ty  *Type
}

type Var struct {
	name    string
	offset  int
	isLocal bool
	ty      *Type
}

type ArrayRef struct {
	lhs Expr
	rhs Expr
	ty  *Type
}

type FuncCall struct {
	name string
	args []Expr
	ty   *Type // TODO: support function type.
}

type Addr Unary
type Deref Unary

func (Binary) isExpr()    {}
func (FuncCall) isExpr()  {}
func (Var) isExpr()       {}
func (Addr) isExpr()      {}
func (Deref) isExpr()     {}
func (ArrayRef) isExpr()  {}
func (IntLit) isExpr()    {}
func (StringLit) isExpr() {}
func (Empty) isExpr()     {}

func (b Binary) getType() *Type    { return b.ty }
func (f FuncCall) getType() *Type  { return f.ty }
func (v Var) getType() *Type       { return v.ty }
func (a Addr) getType() *Type      { return a.ty }
func (d Deref) getType() *Type     { return d.ty }
func (a ArrayRef) getType() *Type  { return a.ty }
func (i IntLit) getType() *Type    { return i.ty }
func (s StringLit) getType() *Type { return s.ty }
func (e Empty) getType() *Type     { return nil }

func (b Binary) setType(ty Type)    { *b.ty = ty }
func (f FuncCall) setType(ty Type)  { *f.ty = ty }
func (v Var) setType(ty Type)       { *v.ty = ty }
func (a Addr) setType(ty Type)      { *a.ty = ty }
func (d Deref) setType(ty Type)     { *d.ty = ty }
func (a ArrayRef) setType(ty Type)  { *a.ty = ty }
func (i IntLit) setType(ty Type)    { *i.ty = ty }
func (s StringLit) setType(ty Type) { *s.ty = ty }
func (e Empty) setType(ty Type)     {}

func findVar(name string) *Var {
	for _, v := range globals {
		if v.name == name {
			return &v
		}
	}
	for _, v := range tmpLocals {
		if v.name == name {
			return &v
		}
	}
	return nil
}

func newLabel() string {
	l := fmt.Sprintf(".L.data.%d", contentCnt)
	contentCnt++
	return l
}

func getContent(s string) *StringLit {
	for _, c := range contents {
		if c.val == s {
			return &c
		}
	}
	return nil
}

func arrayLength() int {
	idx := -1
	// Array.
	if consume("[") {
		// only supports a fixed array.
		idx = tokens[0].val
		tokens = tokens[1:]
		assert("]")
	}
	return idx
}

// VarSpec = Identifier ( Type [ "=" Expression ] )
func varSpec() Var {
	tokId := consumeToken(TK_IDENT)
	if tokId == nil {
		panic(fmt.Sprintf("expected an identifier but got %#v\n", tokId))
	}

	length := arrayLength()
	if length == -1 {
		length = 1
	}

	tokTy := consumeToken(TK_TYPE)
	if tokTy == nil {
		ty := newNoneType()
		return Var{tokId.str, varOffset, true, &ty}
	}

	if !supportType(tokTy.str) {
		panic(fmt.Sprintf("unsupported type %s\n", tokTy.str))
	}

	ty := newLiteralType(tokTy.str)
	ty.aryLen = length
	return Var{tokId.str, varOffset, true, &ty}
}

func consume(op string) bool {
	if len(tokens) > 0 && tokens[0].str == op {
		tokens = tokens[1:]
		return true
	}
	return false
}

func consumeToken(tk TokenKind) *Token {
	if len(tokens) > 0 && tokens[0].kind == tk {
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
	panic(fmt.Sprintf("expected %s but got %s\n", op, tokens[0].str))
}

func assertType() string {
	tok := consumeToken(TK_TYPE)
	if tok == nil {
		panic(fmt.Sprintf("expected type but got %#v\n", tokens[0]))
	}

	if !supportType(tok.str) {
		panic(fmt.Sprintf("unsupported type %s\n", tok.str))
	}
	return tok.str
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
		v := varSpec()

		varOffset += (v.ty.aryLen * 8)
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

func program() (Program, string) {
	assert("package")
	tok := consumeToken(TK_IDENT)
	pkgName := tok.str
	consume(";")

	funcs := make([]Function, 0)

	preStmts := make([]Stmt, 0)
	funcs = append(funcs, Function{"preMain", []Var{}, []Var{}, nil, 8})
	for len(tokens) > 0 {
		if consume("func") {
			funcs = append(funcs, function())
			continue
		}

		// Global variable.
		if consume("var") {
			v := varSpec()

			v.isLocal = false
			globals = append(globals, v)

			if consume(";") {
				continue
			}

			if consume("=") {
				preStmts = append(preStmts, assign(v))
			}
			continue
		}

		if consume(";") {
			continue
		}
	}
	ty := newLiteralType("int64")
	preStmts = append(preStmts, Return{IntLit{0, &ty}})
	funcs[0].stmts = preStmts
	return Program{globals, contents, funcs}, pkgName
}

func function() Function {
	// Initialize for a function.
	varOffset = 8
	tmpLocals = make([]Var, 0)

	tok := consumeToken(TK_IDENT)
	if tok == nil {
		panic(fmt.Sprintf("expected an identifier after 'func' keyword but got %#v\n", tok))
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

func assign(v Var) Stmt {
	length := arrayLength()
	if length == -1 {
		return Assign{[]Expr{v}, []Expr{expr()}}
	}

	// Multiple elements in an array.
	assertType()
	assert("{")

	v.ty.aryLen = length
	varOffset += ((v.ty.aryLen - 1) * 8)

	lvals := make([]Expr, length)
	rvals := exprList()
	// Expand left-side expressions.
	for i := 0; i < length; i++ {
		ity := newLiteralType("int64")
		pty := newLiteralType("pointer")
		lvals[i] = ArrayRef{v, IntLit{i, &ity}, &pty}
	}
	assert("}")
	consume(";")
	return Assign{lvals, rvals}
}

func stmt() Stmt {
	// Var declaration.
	if consume("var") {
		v := varSpec()

		varp := findVar(v.name)
		if varp != nil {
			panic(fmt.Sprintf("%s is already declared. No new variables\n", v.name))
		}

		//varOffset += (v.ty.aryLen * v.ty.size)
		varOffset += (v.ty.aryLen * 8)
		tmpLocals = append(tmpLocals, v)

		if consume("=") {
			return assign(v)
		}
		// Return Empty struct because of no assignment.
		return Empty{}
	}

	// Return statement.
	if consume("return") {
		return Return{expr()}
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

		varOffset += (v.ty.aryLen * 8)
		tmpLocals = append(tmpLocals, v)
		return assign(v)
	}

	// Assignment statement.
	if consume("=") {
		switch v := exprN.(type) {
		case Var:
			varp := findVar(v.name)
			if varp == nil {
				panic(fmt.Sprintf("undefined: %s\n", v.name))
			}
		}
		return Assign{[]Expr{exprN}, []Expr{expr()}}
	}

	// Expression statement.
	return ExprStmt{exprN}
}

func exprList() []Expr {
	exprs := make([]Expr, 0)
	exprs = append(exprs, expr())
	for consume(",") {
		exprs = append(exprs, expr())
	}
	return exprs
}

func expr() Expr {
	if consume(";") {
		return Empty{}
	}
	return equality()
}

func equality() Expr {
	exprN := relational()

	ty := newLiteralType("bool")
	for len(tokens) > 0 {
		if consume("==") {
			exprN = Binary{"==", exprN, relational(), &ty}
		} else if consume("!=") {
			exprN = Binary{"!=", exprN, relational(), &ty}
		} else {
			return exprN
		}
	}
	return exprN
}

func relational() Expr {
	exprN := arrayref()

	ty := newLiteralType("bool")
	for len(tokens) > 0 {
		if consume("<") {
			exprN = Binary{"<", exprN, arrayref(), &ty}
		} else if consume("<=") {
			exprN = Binary{"<=", exprN, arrayref(), &ty}
		} else if consume(">") {
			exprN = Binary{"<", arrayref(), exprN, &ty}
		} else if consume(">=") {
			exprN = Binary{"<=", arrayref(), exprN, &ty}
		} else {
			return exprN
		}
	}
	return exprN
}

func arrayref() Expr {
	exprN := add()

	if consume("[") {
		ty := newNoneType()
		exprN = ArrayRef{exprN, add(), &ty}
		assert("]")
	}
	return exprN
}

func add() Expr {
	exprN := mul()

	ty := newNoneType()
	for len(tokens) > 0 {
		if consume("+") {
			exprN = Binary{"+", exprN, mul(), &ty}
		} else if consume("-") {
			exprN = Binary{"-", exprN, mul(), &ty}
		} else {
			return exprN
		}
	}
	return exprN
}

func mul() Expr {
	exprN := unary()

	ty := newNoneType()
	for len(tokens) > 0 {
		if consume("*") {
			exprN = Binary{"*", exprN, unary(), &ty}
		} else if consume("/") {
			exprN = Binary{"/", exprN, unary(), &ty}
		} else {
			return exprN
		}
	}
	return exprN
}

func unary() Expr {
	nty := newNoneType()
	if consume("+") {
		return unary()
	} else if consume("-") {
		// -val = 0 - val
		ity := newLiteralType("int64")
		return Binary{"-", IntLit{0, &ity}, unary(), &nty}
	} else if consume("&") {
		return Addr{unary(), &nty}
	} else if consume("*") {
		return Deref{unary(), &nty}
	}
	return operand()
}

// Operand = Literal | OperandName | "(" Expression ")" .
func operand() Expr {
	// "(" Expression ")".
	if consume("(") {
		exprN := expr()
		assert(")")
		return exprN
	}

	// OperandName = identifier.
	tok := consumeToken(TK_IDENT)
	if tok != nil {
		nty := newNoneType()
		// Function call.
		if consume("(") {
			return FuncCall{tok.str, funcArgs(), &nty}
		}

		// Ex. Get 2 in a[2], and get 1 in case of a normal varialbe.
		idx := arrayLength()

		// Variable.
		// Not register to `tmpLocals` yet.
		varp := findVar(tok.str)
		if idx == -1 {
			// Normal variable.
			if varp == nil {
				return Var{tok.str, varOffset, true, &nty}
			}
			return *varp
		}

		// Array reference should be declared beforehand.
		if varp == nil {
			panic(fmt.Sprintf("undefined %s", varp.name))
		}

		// Index overflow.
		// TODO: Now, accessing index over size is possible.
		// e.g. hoge:="abc"; hoge[5] is possible because `hoge` doesn't have its type yet
		// so there is no way to check string length or array length.
		// if varp.ty.aryLen <= idx && varp.ty.kind != "string" {
		//	panic(fmt.Sprintf("Invalid array index %d", idx))
		//}
		ity := newLiteralType("int64")
		pty := newLiteralType("pointer")
		return ArrayRef{*varp, IntLit{idx, &ity}, &pty}
	}
	return literal()
}

func literal() Expr {
	// String literal.
	if consume("\"") {
		ty := newLiteralType("string")
		n := StringLit{tokens[0].str, newLabel(), &ty}
		contents = append(contents, n)
		tokens = tokens[1:]
		assert("\"")
		return n
	}

	// Character (int32).
	if consume("'") {
		ty := newLiteralType("int32")
		n := IntLit{tokens[0].val, &ty}
		tokens = tokens[1:]
		assert("'")
		return n
	}

	// Integer literal.
	ty := newLiteralType("int64")
	n := IntLit{tokens[0].val, &ty}
	tokens = tokens[1:]
	return n
}
