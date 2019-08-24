package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

var in string
var userIn string

// tokenizer

type TokenKind int

const (
	TK_RESERVED = iota
	TK_NUM
)

type Token struct {
	kind TokenKind
	val  int
	str  string
}

func tokenize() []Token {
	tokens := make([]Token, 0)
	for len(in) > 0 {
		if in[0] == ' ' {
			in = in[1:]
			continue
		}
		if strings.Contains("+-*/", string(in[0])) {
			tokens = append(tokens, Token{TK_RESERVED, -1, string(in[0])})
			in = in[1:]
			continue
		}
		if isInt() {
			tokens = append(tokens, Token{TK_NUM, toInt(), ""})
			continue
		}
		tokenError("Unexcected character:", string(in[0]))
	}
	return tokens
}

// parser

type NodeKind int

const (
	ND_ADD = iota // +
	ND_SUB        // -
	ND_MUL        // *
	ND_DIV        // /
	ND_NUM        // Integer
)

type Node struct {
	kind NodeKind
	lhs  *Node
	rhs  *Node
	val  int
}

var tokens []Token

func expr() *Node {
	node := mul()

	for len(tokens) > 0 {
		if tokens[0].str == "+" {
			tokens = tokens[1:]
			node = &Node{ND_ADD, node, mul(), -1}
		} else if tokens[0].str == "-" {
			tokens = tokens[1:]
			node = &Node{ND_SUB, node, mul(), -1}
		} else {
			return node
		}
	}
	return node
}

func mul() *Node {
	node := primary()

	for len(tokens) > 0{
		if tokens[0].str == "*" {
			tokens = tokens[1:]
			node = &Node{ND_MUL, node, primary(), -1}
		} else if tokens[0].str == "/" {
			tokens = tokens[1:]
			node = &Node{ND_DIV, node, primary(), -1}
		} else {
			return node
		}
	}
	return node
}

func primary() *Node {
	if tokens[0].str == "(" {
		node := expr()
		if tokens[0].str != ")" {
			fmt.Println(tokens)
			fmt.Printf("[Error] expected ) but got %s\n", tokens[0].str)
			os.Exit(1)
		}
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

// generator

func gen(node *Node) {
	if node.kind == ND_NUM {
		fmt.Printf("  push %d\n", node.val)
		return
	}

	gen(node.lhs)
	gen(node.rhs)

	fmt.Printf("  pop rdi\n")
	fmt.Printf("  pop rax\n")

	switch node.kind {
	case ND_ADD:
		fmt.Printf("  add rax, rdi\n")
	case ND_SUB:
		fmt.Printf("  sub rax, rdi\n")
	case ND_MUL:
		fmt.Printf("  imul rax, rdi\n")
	case ND_DIV:
		fmt.Printf("  cqo\n")
		fmt.Printf("  idiv rdi\n")
	default:
		fmt.Println("[Error] Unexpected node:", node)
		os.Exit(1)
	}
	fmt.Printf("  push rax\n")
}

func isInt() bool {
	_, err := strconv.Atoi(string(in[0]))
	return err == nil
}

func toInt() int {
	n := 0
	for len(in) > 0 && '0' <= in[0] && in[0] <= '9' {
		n = n*10 + int(in[0]-'0')
		in = in[1:]
	}
	return n
}

func tokenError(f string, vars ...string) {
	n := len(userIn) - len(in)
	fmt.Println(userIn)
	fmt.Println(strings.Repeat(" ", n) + "^")
	fmt.Println("[Error]", f, vars)
	os.Exit(1)
}

func main() {
	if len(os.Args) != 2 {
		panic("Invalid number of command line arguments.")
	}
	userIn = os.Args[1]
	in = os.Args[1]

	tokens = tokenize()
        //fmt.Println(tokens)
	node := expr()
	//printNode(node, 0)

	fmt.Printf(".intel_syntax noprefix\n")
	fmt.Printf(".global _main\n")
	fmt.Printf("_main:\n")

	gen(node)

	fmt.Printf("  pop rax\n")
	fmt.Printf("  ret\n")
	return
}
