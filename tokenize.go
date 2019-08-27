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
	TK_RESERVED = iota // Keywords or punctuators
	TK_IDENT           // Identifiers
	TK_NUM             // Integer literals
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

func isNum(b byte) bool {
	_, err := strconv.Atoi(string(b))
	return err == nil
}

func getNum() int {
	n := 0
	for len(in) > 0 && '0' <= in[0] && in[0] <= '9' {
		n = n*10 + int(in[0]-'0')
		in = in[1:]
	}
	return n
}

func isAlpha(b byte) bool {
	return ('a' <= b && b <= 'z') || ('A' <= b && b <= 'Z')
}

func isAlnum(b byte) bool {
	return isNum(b) || isAlpha(b)
}

func startsReserved() string {
	keywords := []string{"return", "if", "else", "for"}
	for _, kw := range keywords {
		if strings.HasPrefix(in, kw) {
			if len(kw) == len(in) || !isAlnum(in[len(kw)]) {
				return kw
			}
		}
	}

	ops := []string{"==", "!=", "<=", ">="}
	for _, op := range ops {
		if strings.HasPrefix(in, op) {
			return op
		}
	}

	if strings.Contains("+-*/()<>;={},", in[0:1]) {
		return in[0:1]
	}
	return ""
}

func tokenize() []Token {
	tokens := make([]Token, 0)
	for len(in) > 0 {
		if in[0] == ' ' {
			in = in[1:]
			continue
		}
		kw := startsReserved()
		if len(kw) != 0 {
			tokens = append(tokens, Token{TK_RESERVED, -1, kw})
			in = in[len(kw):]
			continue
		}
		if isAlpha(in[0]) {
			name := in[0:1]
			in = in[1:]
			for len(in) > 0 && isAlnum(in[0]) {
				name += in[0:1]
				in = in[1:]
			}
			tokens = append(tokens, Token{TK_IDENT, -1, name})
			continue
		}
		if isNum(in[0]) {
			tokens = append(tokens, Token{TK_NUM, getNum(), ""})
			continue
		}
		tokenError("Unexcected character:", in[0:1])
	}
	return tokens
}
