package main

import (
	//"fmt"
	"os"
)

var in string
var userIn string

func main() {
	if len(os.Args) != 2 {
		panic("Invalid number of command line arguments.")
	}
	userIn = os.Args[1]
	in = os.Args[1]

	tokens = tokenize()
	//fmt.Println(tokens)
	nodes := program()
	//printNode(node, 0)

	codegen(nodes)
	return
}
