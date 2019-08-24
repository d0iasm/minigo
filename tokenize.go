package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

var tokens []Token

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

func tokenError(f string, vars ...string) {
	n := len(userIn) - len(in)
	fmt.Println(userIn)
	fmt.Println(strings.Repeat(" ", n) + "^")
	fmt.Println("[Error]", f, vars)
	os.Exit(1)
}

func isInt() bool {
	_, err := strconv.Atoi(in[0:1])
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

func tokenize() []Token {
	tokens := make([]Token, 0)
	for len(in) > 0 {
		if in[0] == ' ' {
			in = in[1:]
			continue
		}
		if len(in) > 6 {
			if in[0:6] == "return" {
				tokens = append(tokens, Token{TK_RESERVED, -1, in[0:6]})
				in = in[6:]
				continue
			}
		}
		if len(in) > 2 {
			if in[0:2] == "==" || in[0:2] == "!=" ||
				in[0:2] == "<=" || in[0:2] == ">=" {
				tokens = append(tokens, Token{TK_RESERVED, -1, in[0:2]})
				in = in[2:]
				continue
			}
		}
		if strings.Contains("+-*/()<>;", in[0:1]) {
			tokens = append(tokens, Token{TK_RESERVED, -1, in[0:1]})
			in = in[1:]
			continue
		}
		if isInt() {
			tokens = append(tokens, Token{TK_NUM, toInt(), ""})
			continue
		}
		tokenError("Unexcected character:", in[0:1])
	}
	return tokens
}