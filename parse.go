package main

import (
	"fmt"
	"os"
)

type NodeKind int

const (
	ND_ADD = iota // +
	ND_SUB        // -
	ND_MUL        // *
	ND_DIV        // /
	ND_EQ         // ==
	ND_NE         // !=
	ND_LT         // <
	ND_LE         // <=
	ND_NUM        // Integer
)

type Node struct {
	kind NodeKind
	lhs  *Node
	rhs  *Node
	val  int
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

func expr() *Node {
	return equality()
}

func equality() *Node {
	node := relational()

	for len(tokens) > 0 {
		if consume("==") {
			node = &Node{ND_EQ, node, relational(), -1}
		} else if consume("!=") {
			node = &Node{ND_NE, node, relational(), -1}
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
			node = &Node{ND_LT, node, add(), -1}
		} else if consume("<=") {
			node = &Node{ND_LE, node, add(), -1}
		} else if consume(">") {
			node = &Node{ND_LT, add(), node, -1}
		} else if consume(">=") {
			node = &Node{ND_LE, add(), node, -1}
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
			node = &Node{ND_ADD, node, mul(), -1}
		} else if consume("-") {
			node = &Node{ND_SUB, node, mul(), -1}
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
			node = &Node{ND_MUL, node, unary(), -1}
		} else if consume("/") {
			node = &Node{ND_DIV, node, unary(), -1}
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
		num := &Node{ND_NUM, nil, nil, 0}
		return &Node{ND_SUB, num, unary(), -1} // -val = 0 - val
	}
	return primary()
}

func primary() *Node {
	if consume("(") {
		node := expr()
		expect(")")
		return node
	}

	n := Node{ND_NUM, nil, nil, tokens[0].val}
	tokens = tokens[1:]
	return &n
}

func printNode(node *Node, dep int) {
	if node == nil {
		return
	}

	printNode(node.lhs, dep+1)
	printNode(node.rhs, dep+1)
	fmt.Printf("dep: %d, kind: %d, val: %d\n", dep, node.kind, node.val)
}

