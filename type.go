package main

import (
	"fmt"
)

type TypeKind int

const (
	None TypeKind = iota
	Bool
	Int
	Uint
)

type Type struct {
	ty TypeKind
}

func check(lty TypeKind, rty TypeKind) {
	if lty != rty {
		panic(fmt.Sprintf("Expected type of a left-side node %d, but got %d\n", rty, lty))
	}
}

func addType(node interface{}) {
	switch n := node.(type) {
	case Empty:
		return
	case IntLit:
		*n.ty = Type{Int}
		return
	case Var:
		return
	case VarDecl:
		addType(n.ident)
		addType(n.rval)
		return
	case Assign:
		addType(n.lval)
		addType(n.rval)
		return
	case Addr:
		addType(n.child)
		return
	case Deref:
		addType(n.child)
		return
	case Block:
		for _, c := range n.children {
			addType(c)
		}
		return
	case ExprStmt:
		addType(n.child)
		return
	case If:
		if n.init != nil {
			addType(n.init)
		}
		addType(n.cond)
		addType(n.then)
		if n.els != nil {
			addType(n.els)
		}
		return
	case For:
		if !isEmpty(n.init) {
			addType(n.init)
		}
		if !isEmpty(n.cond) {
			addType(n.cond)
		}
		addType(n.then)
		if !isEmpty(n.post) {
			addType(n.post)
		}
		return
	case Return:
		addType(n.child)
		return
	case FuncCall:
		for _, arg := range n.args {
			addType(arg)
		}
		return
	}

	n := node.(Binary)

	// Tree traversal from a right-side node.
	addType(n.rhs)
	addType(n.lhs)

	switch n.op {
	case "+":
		/**
		switch rhs := n.rhs.(type) {
		case IntLit:
			switch lhs := n.lhs.(type) {
			case IntLit:
				rhs.ty = Int
				lhs.ty = Int
			case Var:
				rh

			}
		case Var:
			return
		}

		if n.lhs.ty == None && n.rhs.ty != None {
			n.lhs.ty = n.rhs.ty
		}
		check(n.lhs.ty, n.rhs.ty)
		*/
	case "-":
	case "*":
	case "/":
	case "==":
	case "!=":
	case "<":
	case "<=":
	default:
		panic("[Error] Unexpected node")
	}
}
