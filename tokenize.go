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
	TK_TYPE            // Types
	TK_NUM             // Integer literals
	TK_STRING          // String literals
)

type Token struct {
	kind TokenKind
	val  int
	str  string
}

func tokenError(f string, vars ...string) {
	// TODO: improve an error message and place
	n := len(userIn) - len(in)
	fmt.Println(userIn)
	fmt.Println(strings.Repeat(" ", n) + "^")
	fmt.Println(f, vars)
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
	keywords := []string{"return", "if", "else", "for", "func", "var", "package"}
	for _, kw := range keywords {
		if strings.HasPrefix(in, kw) {
			if len(kw) == len(in) || !isAlnum(in[len(kw)]) {
				return kw
			}
		}
	}

	ops := []string{"==", "!=", "<=", ">=", ":="}
	for _, op := range ops {
		if strings.HasPrefix(in, op) {
			return op
		}
	}

	if strings.Contains("+-*/()<>;={},&[]'\"", in[0:1]) {
		return in[0:1]
	}
	return ""
}

func startsType() string {
	for _, t := range typeKinds {
		if strings.HasPrefix(in, t) {
			if len(t) == len(in) || !isAlnum(in[len(t)]) {
				return t
			}
		}
	}
	return ""
}

func insertEnd() {
	if len(tokens) < 1 {
		return
	}
	if strings.Contains("{}", tokens[len(tokens) - 1].str) {
		return
	}
	if tokens[len(tokens) - 1].str == ";" {
		return
	}
	tokens = append(tokens, Token{TK_RESERVED, -1, ";"})
}

func tokenize() []Token {
	tokens := make([]Token, 0)
	for len(in) > 0 {
		if in[0] == ' ' || in[0] == '\t' {
			in = in[1:]
			continue
		}
		if in[0] == '\n' {
			insertEnd()
			in = in[1:]
			continue
		}
		ty := startsType()
		if len(ty) != 0 {
			tokens = append(tokens, Token{TK_TYPE, -1, ty})
			in = in[len(ty):]
			continue
		}
		kw := startsReserved()
		if len(kw) != 0 {
			tokens = append(tokens, Token{TK_RESERVED, -1, kw})
			in = in[len(kw):]
			continue
		}
		if isAlpha(in[0]) {
			if tokens[len(tokens)-1].str == "'" {
				// Character.
				tokens = append(tokens, Token{TK_NUM, int(in[0]), in[0:1]})
				in = in[1:]
				if in[0] != '\'' {
					panic("invalid character literal (more than one character)")
				}
				continue
			}

			str := in[0:1]
			in = in[1:]
			for len(in) > 0 && isAlnum(in[0]) {
				str += in[0:1]
				in = in[1:]
			}

			if tokens[len(tokens)-1].str == "\"" {
				// String.
				tokens = append(tokens, Token{TK_STRING, -1, str})
			} else {
				// Variable.
				tokens = append(tokens, Token{TK_IDENT, -1, str})
			}
			continue
		}
		if isNum(in[0]) {
			tokens = append(tokens, Token{TK_NUM, getNum(), ""})
			continue
		}
		tokenError("unexcected character:", in[0:1])
	}
	return tokens
}
