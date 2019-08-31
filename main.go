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

	flag.Parse()

	isDev = *devPtr

	if inPtr != nil {
		if len(os.Args) < 1 {
			panic("invalid number of arguments")
		}
		b, err := ioutil.ReadFile(os.Args[1]) // just pass the file name

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

	prog := program()
	if isDev {
		printNodes(prog.funcs)
	}

	codegen(prog)
	return
}
