package main

import (
	"flag"
	"fmt"
)

var in string
var userIn string

var isDev bool

func parseArgs() {
	devPtr := flag.Bool("dev", false, "Output logs for development.")
	inPtr := flag.String("in", "", "Input string directly.")

	flag.Parse()

	isDev = *devPtr
	in = *inPtr
	userIn = *inPtr
}

func main() {
	parseArgs()

	tokens = tokenize()
	if isDev {
		fmt.Println(tokens)
	}

	prog := program()
	if isDev {
		printNodes(prog.funcs)
	}

	codegen(prog)
	return
}
