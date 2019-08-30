package main

import (
	"fmt"
)

var typeKinds = []string{"none", "bool", "int64", "string", "pointer"}

type Type struct {
	kind   string
	length int
}

func supportType(kind string) bool {
	for _, t := range typeKinds {
		if kind == t {
			return true
		}
	}
	return false
}

func typeCheck(lty *Type, rty *Type, op string) {
	if lty.kind != rty.kind {
		panic(fmt.Sprintf("invalid operation %#v %s %#v (mismatched types %s and %s)", lty, op, rty, lty.kind, rty.kind))
	}
}

func addType(node interface{}) {
	switch n := node.(type) {
	// Expressions. It should have Type field.
	case IntLit:
		if n.ty.kind == "int64" {
			return
		}
		*n.ty = Type{"int64", 1}
	case Addr:
		addType(n.child)
		if n.ty.kind == "pointer" {
			return
		}
		*n.ty = Type{"pointer", 1}
	case Deref:
		addType(n.child)
		if n.child.getType().kind == "pointer" {
			// TODO: how to get the type of child of child?
			*n.ty = Type{"int64", 1}
		}
		*n.ty = Type{n.child.getType().kind, n.child.getType().length}
	case Binary:
		addType(n.lhs)
		addType(n.rhs)
		typeCheck(n.lhs.getType(), n.rhs.getType(), n.op)
		switch n.op {
		case "+", "-", "*", "/":
			*n.ty = Type{n.lhs.getType().kind, n.lhs.getType().length}
		case "==", "!=", "<", "<=":
			*n.ty = Type{"bool", 1}
		}
	case Var:
		// The type of variables are defined at Assgin node.
	case ArrayRef:
		addType(n.v)
		if n.ty.kind == "pointer" {
			return
		}
		*n.ty = Type{"pointer", 1}
	case FuncCall:
		for _, arg := range n.args {
			addType(arg)
		}
	// Statements.
	case Empty:
	case ExprStmt:
		addType(n.child)
	case Return:
		addType(n.child)
	case Block:
		for _, c := range n.children {
			addType(c)
		}
	case If:
		if n.init != nil {
			addType(n.init)
		}
		addType(n.cond)
		addType(n.then)
		if n.els != nil {
			addType(n.els)
		}
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
	case Assign:
		if len(n.lvals) != len(n.rvals) {
			panic(fmt.Sprintf("Not same length %d != %d", len(n.lvals), len(n.rvals)))
		}
		for i := range n.lvals {
			addType(n.lvals[i])
			addType(n.rvals[i])
			n.lvals[i].setType(*n.rvals[i].getType())
		}
	default:
		panic(fmt.Sprintf("Unexpected node type %#v", n))
	}
}
