package main

import (
	"fmt"
)

var globals []Var
var tmpLocals []Var
var varOffset int = 8

var contents []String
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
	contents []String
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

type String struct {
	val   string
	label string
	idx   int
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
	v   Var
	idx int
	ty  *Type
}

type FuncCall struct {
	name string
	args []Expr
	ty   *Type // TODO: support function type.
}

type Addr Unary
type Deref Unary

func (Binary) isExpr()   {}
func (FuncCall) isExpr() {}
func (Var) isExpr()      {}
func (Addr) isExpr()     {}
func (Deref) isExpr()    {}
func (ArrayRef) isExpr() {}
func (IntLit) isExpr()   {}
func (String) isExpr()   {}
func (Empty) isExpr()    {}

func (b Binary) getType() *Type   { return b.ty }
func (f FuncCall) getType() *Type { return f.ty }
func (v Var) getType() *Type      { return v.ty }
func (a Addr) getType() *Type     { return a.ty }
func (d Deref) getType() *Type    { return d.ty }
func (a ArrayRef) getType() *Type { return a.ty }
func (i IntLit) getType() *Type   { return i.ty }
func (s String) getType() *Type   { return s.ty }
func (e Empty) getType() *Type    { return nil }

func (b Binary) setType(ty Type)   { *b.ty = ty }
func (f FuncCall) setType(ty Type) { *f.ty = ty }
func (v Var) setType(ty Type)      { *v.ty = ty }
func (a Addr) setType(ty Type)     { *a.ty = ty }
func (d Deref) setType(ty Type)    { *d.ty = ty }
func (a ArrayRef) setType(ty Type) { *a.ty = ty }
func (i IntLit) setType(ty Type)   { *i.ty = ty }
func (s String) setType(ty Type)   { *s.ty = ty }
func (e Empty) setType(ty Type)    {}

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

func varSpec() Var {
	tokId := consumeToken(TK_IDENT)
	if tokId == nil {
		panic(fmt.Sprintf("Expected an identifier but got %#v\n", tokId))
	}

	length := arrayLength()
	if length == -1 {
		length = 1
	}

	tokTy := consumeToken(TK_TYPE)
	if tokTy == nil {
		return Var{tokId.str, varOffset, true, &Type{"none", length}}
	}

	if !supportType(tokTy.str) {
		panic(fmt.Sprintf("Unsupported type %s\n", tokTy.str))
	}

	return Var{tokId.str, varOffset, true, &Type{tokTy.str, length}}
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
	panic(fmt.Sprintf("tokens: %s\nExpected %s but got %s\n", tokens, op, tokens[0].str))
}

func assertType() string {
	tok := consumeToken(TK_TYPE)
	if tok == nil {
		panic(fmt.Sprintf("Expected type but got %#v\n", tokens[0]))
	}

	if !supportType(tok.str) {
		panic(fmt.Sprintf("Unsupported type %s\n", tok.str))
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

		varOffset += (v.ty.length * 8)
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

	preStmts := make([]Stmt, 0)
	funcs = append(funcs, Function{"preMain", []Var{}, []Var{}, nil, 8})
	for len(tokens) > 0 {
		if consume("func") {
			funcs = append(funcs, function())
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
				length := arrayLength()
				if length == -1 {
					preStmts = append(preStmts, Assign{[]Expr{v}, []Expr{expr()}})
					assert(";")
					continue
				}

				// Multiple elements in an array.
				assertType()
				assert("{")

				v.ty.length = length

				lvals := make([]Expr, length)
				rvals := exprList()
				// Expand left-side expressions.
				for i := 0; i < length; i++ {
					lvals[i] = ArrayRef{v, i, &Type{"pointer", 1}}
				}
				assert("}")
				assert(";")
				preStmts = append(preStmts, Assign{lvals, rvals})
			}
		}
	}
	preStmts = append(preStmts, Return{IntLit{0, &Type{"int64", 1}}})
	funcs[0].stmts = preStmts
	return Program{globals, contents, funcs}
}

func function() Function {
	// Initialize for a function.
	varOffset = 8
	tmpLocals = make([]Var, 0)

	tok := consumeToken(TK_IDENT)
	if tok == nil {
		panic(fmt.Sprintf("Expected an identifier after 'func' keyword but got %#v\n", tok))
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
	// Var declaration.
	if consume("var") {
		v := varSpec()

		varp := findVar(v.name)
		if varp != nil {
			panic(fmt.Sprintf("%s is already declared. No new variables\n", v.name))
		}

		varOffset += (v.ty.length * 8)
		tmpLocals = append(tmpLocals, v)

		if consume("=") {
			length := arrayLength()
			if length == -1 {
				return Assign{[]Expr{v}, []Expr{expr()}}
			}

			// Multiple elements in an array.
			assertType()
			assert("{")

			v.ty.length = length
			varOffset += ((v.ty.length - 1) * 8)

			lvals := make([]Expr, length)
			rvals := exprList()
			// Expand left-side expressions.
			for i := 0; i < length; i++ {
				lvals[i] = ArrayRef{v, i, &Type{"pointer", 1}}
			}
			assert("}")
			return Assign{lvals, rvals}
		}
		// Return Empty struct because of no assignment.
		return Empty{}
	}

	// Return statement.
	if consume("return") {
		stmtN := Return{expr()}
		// TODO: assert(";") is not necessary?
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

		varOffset += (v.ty.length * 8)
		tmpLocals = append(tmpLocals, v)

		length := arrayLength()
		if length == -1 {
			return Assign{[]Expr{v}, []Expr{expr()}}
		}

		// Multiple elements in an array.
		assertType()
		assert("{")

		v.ty.length = length
		varOffset += ((v.ty.length - 1) * 8)

		lvals := make([]Expr, length)
		rvals := exprList()
		// Expand left-side expressions.
		for i := 0; i < length; i++ {
			lvals[i] = ArrayRef{v, i, &Type{"pointer", 1}}
		}
		assert("}")
		return Assign{lvals, rvals}
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

	for len(tokens) > 0 {
		if consume("==") {
			exprN = Binary{"==", exprN, relational(), &Type{"none", 1}}
		} else if consume("!=") {
			exprN = Binary{"!=", exprN, relational(), &Type{"none", 1}}
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
			exprN = Binary{"<", exprN, add(), &Type{"none", 1}}
		} else if consume("<=") {
			exprN = Binary{"<=", exprN, add(), &Type{"none", 1}}
		} else if consume(">") {
			exprN = Binary{"<", add(), exprN, &Type{"none", 1}}
		} else if consume(">=") {
			exprN = Binary{"<=", add(), exprN, &Type{"none", 1}}
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
			exprN = Binary{"+", exprN, mul(), &Type{"none", 1}}
		} else if consume("-") {
			exprN = Binary{"-", exprN, mul(), &Type{"none", 1}}
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
			exprN = Binary{"*", exprN, unary(), &Type{"none", 1}}
		} else if consume("/") {
			exprN = Binary{"/", exprN, unary(), &Type{"none", 1}}
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
		return Binary{"-", IntLit{0, &Type{"int64", 1}}, unary(), &Type{"none", 1}} // -val = 0 - val
	} else if consume("&") {
		return Addr{unary(), &Type{"pointer", 1}}
	} else if consume("*") {
		return Deref{unary(), &Type{"none", 1}}
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
		// Function call.
		if consume("(") {
			return FuncCall{tok.str, funcArgs(), &Type{"none", 1}}
		}

		// Ex. Get 2 in a[2], and get 1 in case of a normal varialbe.
		idx := arrayLength()

		// Variable.
		// Not register to `tmpLocals` yet.
		varp := findVar(tok.str)
		if idx == -1 {
			// Normal variable.
			if varp == nil {
				return Var{tok.str, varOffset, true, &Type{"none", 1}}
			}
			return *varp
		}

		// Array reference should be declared beforehand.
		if varp == nil {
			panic(fmt.Sprintf("Undefined %s", varp.name))
		}

		// Index overflow.
		if varp.ty.length <= idx {
			panic(fmt.Sprintf("Invalid array index %d", idx))
		}
		return ArrayRef{*varp, idx, &Type{"pointer", 1}}
	}
	return literal()
}

func literal() Expr {
	// String literal.
	if consume("\"") {
		n := String{tokens[0].str, newLabel(), -1, &Type{"string", 1}}
		contents = append(contents, n)
		tokens = tokens[1:]
		assert("\"")
		// TODO: how to access "hoge"[2]?
		idx := arrayLength()
		n.idx = idx
		return n
	}

	// Character (int32).
	if consume("'") {
		n := IntLit{tokens[0].val, &Type{"int32", 1}}
		tokens = tokens[1:]
		assert("'")
		return n
	}

	// Integer literal.
	n := IntLit{tokens[0].val, &Type{"int64", 1}}
	tokens = tokens[1:]
	return n
}
