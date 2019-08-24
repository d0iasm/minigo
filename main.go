package main

import (
	"fmt"
	"os"
)

var in string

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

	fmt.Printf(".intel_syntax noprefix\n")
	fmt.Printf(".global _main\n")
	fmt.Printf("_main:\n")
	fmt.Printf("  mov rax, %d\n", toInt())
	for len(in) > 0 {
		switch in[0] {
		case '+':
			in = in[1:]
			fmt.Printf("  add rax, %d\n", toInt())
		case '-':
			in = in[1:]
			fmt.Printf("  sub rax, %d\n", toInt())
		default:
			panic("Unexcected character: " + string(in[0]))
		}
	}
	fmt.Printf("  ret\n")
	return
}
