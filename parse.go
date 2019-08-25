package main

import (
	"fmt"
	"os"
)

var locals []*Var

type NodeKind int

const (
	ND_ADD       = iota // +
	ND_SUB              // -
	ND_MUL              // *
	ND_DIV              // /
	ND_EQ               // ==
	ND_NE               // !=
	ND_LT               // <
	ND_LE               // <=
	ND_ASSIGN           // =
	ND_RETURN           // "return"
	ND_BLOCK            // { ... }
	ND_EXPR_STMT        // Expression statements
	ND_VAR              // Local variables
	ND_NUM              // Integer literals
)

type Node struct {
	kind NodeKind
	lhs  *Node
	rhs  *Node

	body []*Node // Used if kind == ND_BLOCK

	val  int  // Used if kind == ND_NUM, otherwise -1
	varp *Var // Used if kind == ND_VAR, otherwise nil
}

type Var struct {
	name   string // Variable name
	offset int    // Offset from RBP
}

type Function struct {
	nodes     []*Node
	locals    []*Var
	stackSize int
}

func newNode(k NodeKind, l *Node, r *Node) *Node {
	return &Node{k, l, r, nil, -1, &Var{"", 0}}
}

func newNumNode(k NodeKind, v int) *Node {
	return &Node{k, nil, nil, nil, v, &Var{"", 0}}
}

func newVarNode(v *Var) *Node {
	return &Node{ND_VAR, nil, nil, nil, -1, v}
}

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
	nodes := make([]*Node, 0)
	for len(tokens) > 0 {
		nodes = append(nodes, stmt())
	}

	funcs := make([]Function, 0)
	funcs = append(funcs, Function{nodes, locals, 0})
	return funcs
}

func stmt() *Node {
	if consume("return") {
		node := newNode(ND_RETURN, expr(), nil)
		expect(";")
		return node
	}

	if consume("{") {
		stmts := make([]*Node, 0)
		for !consume("}") {
			stmts = append(stmts, stmt())
		}
		node := newNode(ND_BLOCK, nil, nil)
		node.body = stmts
		return node
	}
	node := newNode(ND_EXPR_STMT, expr(), nil)
	expect(";")
	return node
}

func expr() *Node {
	return assign()
}

func assign() *Node {
	node := equality()
	if consume("=") {
		node = newNode(ND_ASSIGN, node, assign())
	}
	return node
}

func equality() *Node {
	node := relational()

	for len(tokens) > 0 {
		if consume("==") {
			node = newNode(ND_EQ, node, relational())
		} else if consume("!=") {
			node = newNode(ND_NE, node, relational())
		} else {
			return node
		}
	}
	return node
}

func relational() *Node {
	node := add()

	for len(tokens) > 0 {
		if consume("<") {
			node = newNode(ND_LT, node, add())
		} else if consume("<=") {
			node = newNode(ND_LE, node, add())
		} else if consume(">") {
			node = newNode(ND_LT, add(), node)
		} else if consume(">=") {
			node = newNode(ND_LE, add(), node)
		} else {
			return node
		}
	}
	return node
}

func add() *Node {
	node := mul()

	for len(tokens) > 0 {
		if consume("+") {
			node = newNode(ND_ADD, node, mul())
		} else if consume("-") {
			node = newNode(ND_SUB, node, mul())
		} else {
			return node
		}
	}
	return node
}

func mul() *Node {
	node := unary()

	for len(tokens) > 0 {
		if consume("*") {
			node = newNode(ND_MUL, node, unary())
		} else if consume("/") {
			node = newNode(ND_DIV, node, unary())
		} else {
			return node
		}
	}
	return node
}

func unary() *Node {
	if consume("+") {
		return unary()
	} else if consume("-") {
		return newNode(ND_SUB, newNumNode(ND_NUM, 0), unary()) // -val = 0 - val
	}
	return primary()
}

func primary() *Node {
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
		n := newVarNode(varp)
		tokens = tokens[1:]
		return n
	}

	// Integer literals
	n := newNumNode(ND_NUM, tokens[0].val)
	tokens = tokens[1:]
	return n
}

func printNodes(nodes []*Node) {
	for i, n := range nodes {
		fmt.Println("[Print Node] node:", i)
		printNode(n, 0)
	}
}

func printNode(node *Node, dep int) {
	if node == nil {
		return
	}

	for _, n := range node.body {
		printNode(n, dep)
	}
	printNode(node.lhs, dep+1)
	printNode(node.rhs, dep+1)
	fmt.Printf("dep: %d, kind: %d, val: %d, name: %s\n", dep, node.kind, node.val, node.varp.name)
}
