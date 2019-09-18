package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

var in string
var userIn string

var isDev bool

func parseArgs() {
	devPtr := flag.Bool("dev", false, "Output logs for development.")
	inPtr := flag.String("in", "", "Input string directly.")
	buildPtr := flag.String("build", "", "Input file name.")

	flag.Parse()

	isDev = *devPtr

	if len(*buildPtr) > 0 {
		if len(os.Args) < 1 {
			panic("invalid number of arguments")
		}

		b, err := ioutil.ReadFile(*buildPtr)
		if err != nil {
			panic(err)
		}

		in = string(b)
		userIn = string(b)
	} else {
		in = *inPtr
		userIn = *inPtr
	}
}

func main() {
	parseArgs()

	// tokenize
	tokens = tokenize()
	if isDev {
		fmt.Println(tokens)
	}

	// parse
	prog, pkg := program()

	// type
	for _, gv := range prog.globals {
		addType(gv)
	}
	for _, fn := range prog.funcs {
		resetOffset()
		for _, s := range fn.stmts {
			addType(s)
		}
		fn.stackSize = stackSize(fn.locals)
	}

	// debug
	if isDev {
		fmt.Println("package name:", pkg)
		printNodes(prog.funcs)
	}

	// codegen
	codegen(prog)
	return
}
