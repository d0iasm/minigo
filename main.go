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
		if len(in) > 2 {
                  if string(in[0:2]) == "==" || string(in[0:2]) == "!=" ||
			string(in[0:2]) == "<=" || string(in[0:2]) == ">=" {
			tokens = append(tokens, Token{TK_RESERVED, -1, string(in[0:2])})
			in = in[2:]
			continue
                      }
		}
		if strings.Contains("+-*/()<>", string(in[0])) {
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

var tokens []Token

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
	case ND_EQ:
		fmt.Printf("  cmp rax, rdi\n")
		fmt.Printf("  sete al\n")
		fmt.Printf("  movzx rax, al\n")
		break
	case ND_NE:
		fmt.Printf("  cmp rax, rdi\n")
		fmt.Printf("  setne al\n")
		fmt.Printf("  movzx rax, al\n")
		break
	case ND_LT:
		fmt.Printf("  cmp rax, rdi\n")
		fmt.Printf("  setl al\n")
		fmt.Printf("  movzx rax, al\n")
		break
	case ND_LE:
		fmt.Printf("  cmp rax, rdi\n")
		fmt.Printf("  setle al\n")
		fmt.Printf("  movzx rax, al\n")
		break
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
