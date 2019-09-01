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

	if buildPtr != nil {
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

	tokens = tokenize()
	if isDev {
		fmt.Println(tokens)
	}

	prog, pkg := program()
	if isDev {
		fmt.Println("package name:", pkg)
		printNodes(prog.funcs)
	}

	codegen(prog)
	return
}
