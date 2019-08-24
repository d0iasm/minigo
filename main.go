package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

var in string
var userIn string

type TokenKind int

const (
	TK_RESERVED = iota
	TK_NUM
)

type Token struct {
	kind TokenKind
	val  string
	num  int
}

func tokenize() []Token {
	tokens := make([]Token, 0)
	for len(in) > 0 {
		if in[0] == ' ' {
			in = in[1:]
			continue
		}
		if in[0] == '+' || in[0] == '-' {
			tokens = append(tokens, Token{TK_RESERVED, string(in[0]), -1})
			in = in[1:]
			continue
		}
		if isInt() {
			tokens = append(tokens, Token{TK_NUM, "", toInt()})
			continue
		}
		tokenError("Unexcected character:", string(in[0]))
	}
	return tokens
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
	fmt.Println("[Parse Error]", f, vars)
	os.Exit(1)
}

func main() {
	if len(os.Args) != 2 {
		panic("Invalid number of command line arguments.")
	}
	userIn = os.Args[1]
	in = os.Args[1]

	tokens := tokenize()
	fmt.Println(tokens)

	fmt.Printf(".intel_syntax noprefix\n")
	fmt.Printf(".global _main\n")
	fmt.Printf("_main:\n")
	fmt.Printf("  mov rax, %d\n", tokens[0].num)
	for len(tokens) > 0 {
		switch tokens[0].val {
		case "+":
			fmt.Printf("  add rax, %d\n", tokens[1].num)
			tokens = tokens[2:]
		case "-":
			fmt.Printf("  sub rax, %d\n", tokens[1].num)
			tokens = tokens[2:]
		default:
			fmt.Println(tokens)
			fmt.Println("[Error] Unexpected token:", tokens[0])
			os.Exit(1)
		}
	}
	fmt.Printf("  ret\n")
	return
}
