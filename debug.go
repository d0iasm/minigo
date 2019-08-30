package main

import (
	"fmt"
)

func printNodes(funcs []Function) {
	for i, f := range funcs {
		fmt.Println("")
		fmt.Println("[Function]", i, f.name)
		for i, s := range f.stmts {
			fmt.Println("")
			fmt.Println("[Statements]", i)
			printNode(s, 0)
		}
		fmt.Println("========================")
	}
}

func printNode(node interface{}, dep int) {
	if node == nil {
		fmt.Printf("dep: %d, nil\n", dep)
		return
	}

	switch n := node.(type) {
	// Expressions.
	case IntLit:
		fmt.Printf("dep: %d, node: %#v, type: %#v \n", dep, n, n.ty)
	case Addr:
		fmt.Printf("dep: %d, node: %#v, type: %#v \n", dep, n, n.ty)
		printNode(n.child, dep+1)
	case Deref:
		fmt.Printf("dep: %d, node: %#v, type: %#v \n", dep, n, n.ty)
		printNode(n.child, dep+1)
	case Binary:
		fmt.Printf("dep: %d, node: %#v, type: %#v \n", dep, n, n.ty)
		printNode(n.lhs, dep+1)
		printNode(n.rhs, dep+1)
	case Var:
		fmt.Printf("dep: %d, node: %#v, type: %#v \n", dep, n, n.ty)
	case ArrayRef:
		fmt.Printf("dep: %d, node: %#v, type: %#v \n", dep, n, n.ty)
		printNode(n.v, dep+1)
	case FuncCall:
		fmt.Printf("dep: %d, node: %#v, type: %#v \n", dep, n, n.ty)
		for _, arg := range n.args {
			printNode(arg, dep+1)
		}
	// Statements.
	case Empty:
		fmt.Printf("dep: %d, node: %#v\n", dep, n)
	case ExprStmt:
		fmt.Printf("dep: %d, node: %#v\n", dep, n)
		printNode(n.child, dep+1)
	case Return:
		fmt.Printf("dep: %d, node: %#v\n", dep, n)
		printNode(n.child, dep+1)
	case Block:
		fmt.Printf("dep: %d, node: %#v\n", dep, n)
		for _, c := range n.children {
			printNode(c, dep+1)
		}
	case Assign:
		fmt.Printf("dep: %d, node: %#v\n", dep, n)
		for i := range n.lvals {
			printNode(n.lvals[i], dep+1)
			printNode(n.rvals[i], dep+1)
		}
	case If:
		fmt.Printf("dep: %d, node: %#v\n", dep, n)
		printNode(n.init, dep+1)
		printNode(n.cond, dep+1)
		printNode(n.then, dep+1)
		printNode(n.els, dep+1)
	case For:
		fmt.Printf("dep: %d, node: %#v\n", dep, n)
		printNode(n.init, dep+1)
		printNode(n.cond, dep+1)
		printNode(n.post, dep+1)
		printNode(n.then, dep+1)
	}
}
