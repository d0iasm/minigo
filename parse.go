package main

import (
	"fmt"
	"os"
)

var locals []*Var

type Function struct {
	nodes     []interface{}
	locals    []*Var
	stackSize int
}

type Binary struct {
	lhs interface{}
	rhs interface{}
}

type Unary struct {
	child interface{}
}

type Block struct {
	children []interface{}
}

type Var struct {
	name   string // Variable name
	offset int    // Offset from RBP
}

type Add Binary
type Sub Binary
type Mul Binary
type Div Binary
type Eq Binary
type Ne Binary
type Lt Binary
type Le Binary
type Assign Binary

type Return Unary
type ExprStmt Unary

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
	nodes := make([]interface{}, 0)
	for len(tokens) > 0 {
		nodes = append(nodes, stmt())
	}

	funcs := make([]Function, 0)
	funcs = append(funcs, Function{nodes, locals, 0})
	return funcs
}

func stmt() interface{} {
	if consume("return") {
		node := Return{expr()}
		expect(";")
		return node
	}

	if consume("{") {
		stmts := make([]interface{}, 0)
		for !consume("}") {
			stmts = append(stmts, stmt())
		}
		return Block{stmts}
	}
	node := ExprStmt{expr()}
	expect(";")
	return node
}

func expr() interface{} {
	return assign()
}

func assign() interface{} {
	node := equality()
	if consume("=") {
		node = Assign{node, assign()}
	}
	return node
}

func equality() interface{} {
	node := relational()

	for len(tokens) > 0 {
		if consume("==") {
			node = Eq{node, relational()}
		} else if consume("!=") {
			node = Ne{node, relational()}
		} else {
			return node
		}
	}
	return node
}

func relational() interface{} {
	node := add()

	for len(tokens) > 0 {
		if consume("<") {
			node = Lt{node, add()}
		} else if consume("<=") {
			node = Le{node, add()}
		} else if consume(">") {
			node = Lt{add(), node}
		} else if consume(">=") {
			node = Le{add(), node}
		} else {
			return node
		}
	}
	return node
}

func add() interface{} {
	node := mul()

	for len(tokens) > 0 {
		if consume("+") {
			node = Add{node, mul()}
		} else if consume("-") {
			node = Sub{node, mul()}
		} else {
			return node
		}
	}
	return node
}

func mul() interface{} {
	node := unary()

	for len(tokens) > 0 {
		if consume("*") {
			node = Mul{node, unary()}
		} else if consume("/") {
			node = Div{node, unary()}
		} else {
			return node
		}
	}
	return node
}

func unary() interface{} {
	if consume("+") {
		return unary()
	} else if consume("-") {
		return Sub{0, unary()} // -val = 0 - val
	}
	return primary()
}

func primary() interface{} {
	if consume("(") {
		node := expr()
		expect(")")
		return node
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
	n := tokens[0].val
	tokens = tokens[1:]
	return n
}

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
