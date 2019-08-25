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

	funcs := program()
	if isDev {
		printNodes(funcs[0].stmts)
	}

	for _, f := range funcs {
		f.stackSize = len(f.locals) * 8
	}

	codegen(funcs)
	return
}
