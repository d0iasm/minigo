package main

import (
	"fmt"
	"os"
)

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
	ND_EXPR_STMT        // Expression statements
	ND_LVAR             // Local variables
	ND_NUM              // Integer literals
)

type Node struct {
	kind NodeKind
	lhs  *Node
	rhs  *Node
	val  int    // Used if kind == ND_NUM, otherwise -1
	name string // Used if kind == ND_LVAR, otherwise empty
}

func newNode(k NodeKind, l *Node, r *Node) *Node {
	return &Node{k, l, r, -1, ""}
}

func newNumNode(k NodeKind, v int) *Node {
	return &Node{k, nil, nil, v, ""}
}

func newVarNode(n string) *Node {
	return &Node{ND_LVAR, nil, nil, -1, n}
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

func program() []*Node {
	nodes := make([]*Node, 0)
	for len(tokens) > 0 {
		nodes = append(nodes, stmt())
	}
	return nodes
}

func stmt() *Node {
	if consume("return") {
		node := newNode(ND_RETURN, expr(), nil)
		expect(";")
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

	// identifiers.
	if tokens[0].kind == TK_IDENT {
		n := newVarNode(tokens[0].str)
		tokens = tokens[1:]
		return n
	}

	// integer literals.
	n := newNumNode(ND_NUM, tokens[0].val)
	tokens = tokens[1:]
	return n
}

func printNode(node *Node, dep int) {
	if node == nil {
		return
	}

	printNode(node.lhs, dep+1)
	printNode(node.rhs, dep+1)
	fmt.Printf("dep: %d, kind: %d, val: %d, name %s\n", dep, node.kind, node.val, node.name)
}
