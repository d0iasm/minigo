package main

import (
	"fmt"
	"os"
	"strconv"
)

var in string

type TokenKind int

const (
	TK_RESERVED = iota
	TK_NUM
)

type Token struct {
	kind TokenKind
	val  string
        num int
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
		panic("Unexcected character: " + string(in[0]))
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

func main() {
	if len(os.Args) != 2 {
		panic("Invalid number of command line arguments.")
	}
	in = os.Args[1]

	tokens := tokenize()

	fmt.Printf(".intel_syntax noprefix\n")
	fmt.Printf(".global _main\n")
	fmt.Printf("_main:\n")
	fmt.Printf("  mov rax, %d\n", tokens[0].num)
	tokens = tokens[1:]
	for len(tokens) > 0 {
		switch tokens[0].val {
		case "+":
			fmt.Printf("  add rax, %d\n", tokens[1].num)
			tokens = tokens[2:]
		case "-":
			fmt.Printf("  sub rax, %d\n", tokens[1].num)
			tokens = tokens[2:]
		default:
			panic("Unexcected character: " + tokens[0].val)
		}
	}
	fmt.Printf("  ret\n")
	return
}
