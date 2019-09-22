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
	TK_RESERVED TokenKind = iota // Keywords or punctuators
	TK_IDENT                     // Identifiers
	TK_TYPE                      // Types
	TK_NUM                       // Integer literals
	TK_STRING                    // String literals
	TK_LIBS                      // Call standard libraies
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

func startReserved() string {
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

	if strings.Contains("+-*/()<>;={},&[]'", in[0:1]) {
		return in[0:1]
	}
	return ""
}

func startLib() string {
	stdlibs := []string{"println"}
	for _, lib := range stdlibs {
		if strings.HasPrefix(in, lib) {
			if len(lib) == len(in) || !isAlnum(in[len(lib)]) {
				return lib
			}
		}
	}
	return ""
}

func startType() string {
	typeStrs := []string{"bool", "int", "int8", "int32", "int64", "string"}
	for _, t := range typeStrs {
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
	if strings.Contains("{}", tokens[len(tokens)-1].str) {
		return
	}
	if tokens[len(tokens)-1].str == ";" {
		return
	}
	tokens = append(tokens, Token{TK_RESERVED, -1, ";"})
}

func readLine() string {
	str := in[0:1]
	in = in[1:]
	for len(in) > 0 && in[0] != '\n' && in[0] != ';' {
		str += in[0:1]
		in = in[1:]
	}
	if len(in) > 0 {
		// Remove `\n` or `;`.
		in = in[1:]
	}
	return str
}

func readUntil(s string) string {
	str := ""
	for in[0:1] != s {
		str += in[0:1]
		in = in[1:]
		if in[0:1] == "\n" {
			panic("expected end of string '\"' but got \\n")
		}
	}
	return str
}

func readChunk() string {
	str := in[0:1]
	in = in[1:]
	for len(in) > 0 && isAlnum(in[0]) && in[0] != '\n' {
		str += in[0:1]
		in = in[1:]
	}
	return str
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
		if len(in) > 2 && in[0:2] == "//" {
			// Ignore comments.
			in = in[2:]
			_ = readLine()
			continue
		}

		lib := startLib()
		if len(lib) != 0 {
			tokens = append(tokens, Token{TK_LIBS, -1, lib})
			in = in[len(lib):]
			continue
		}

		ty := startType()
		if len(ty) != 0 {
			tokens = append(tokens, Token{TK_TYPE, -1, ty})
			in = in[len(ty):]
			continue
		}

		kw := startReserved()
		if len(kw) != 0 {
			tokens = append(tokens, Token{TK_RESERVED, -1, kw})
			in = in[len(kw):]
			continue
		}

		// String.
		if in[0:1] == "\"" {
			tokens = append(tokens, Token{TK_RESERVED, -1, "\""})
			in = in[1:]
			str := readUntil("\"")
			tokens = append(tokens, Token{TK_STRING, -1, str})
			tokens = append(tokens, Token{TK_RESERVED, -1, "\""})
			in = in[1:]
			continue
		}

		if isAlpha(in[0]) {
			// Character.
			if tokens[len(tokens)-1].str == "'" {
				tokens = append(tokens, Token{TK_NUM, int(in[0]), in[0:1]})
				in = in[1:]
				if in[0] != '\'' {
					panic("invalid character literal (more than one character)")
				}
				continue
			}

			// Variable.
			str := readChunk()
			tokens = append(tokens, Token{TK_IDENT, -1, str})
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
