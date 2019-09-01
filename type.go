package main

import (
	"fmt"
)

type TypeKind int

const (
	TY_NONE TypeKind = iota
	TY_BOOL
	TY_INT32
	TY_INT64
	TY_STRING

	TY_PTR

	TY_ARRAY
)

type Type struct {
	kind   TypeKind
	length int
	base   *Type
}

func typeKind(s string) TypeKind {
	switch s {
	case "bool":
		return TY_BOOL
	case "int32":
		return TY_INT32
	case "int64":
		return TY_INT64
	case "string":
		return TY_STRING
	case "pointer":
		return TY_PTR
	case "array":
		return TY_ARRAY
	default:
		return TY_NONE
	}
}

func newLiteral(s string, l int) *Type {
	return &Type{typeKind(s), l, nil}
}

func pointerTo(base *Type) Type {
	return Type{TY_PTR, 1, base}
}

func supportType(s string) bool {
	if typeKind(s) == TY_NONE {
		return false
	}
	return true
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
		if n.ty.kind != TY_NONE {
			return
		}
		n.ty = newLiteral("int64", 1)
	case StringLit:
		if n.ty.kind == TY_STRING {
			return
		}
		n.ty = newLiteral("string", 1)
	case Addr:
		addType(n.child)
		if n.ty.kind == TY_PTR {
			return
		}
		n.ty = newLiteral("pointer", 1)
	case Deref:
		addType(n.child)
		if n.child.getType().kind == TY_PTR {
			// TODO: how to get the type of child of child?
			n.ty = newLiteral("int64", 1)
			return
		}
		*n.ty = Type{n.child.getType().kind, n.child.getType().length, nil}
	case Binary:
		addType(n.lhs)
		addType(n.rhs)
		typeCheck(n.lhs.getType(), n.rhs.getType(), n.op)
		switch n.op {
		case "+", "-", "*", "/":
			*n.ty = Type{n.lhs.getType().kind, n.lhs.getType().length, nil}
		case "==", "!=", "<", "<=":
			n.ty = newLiteral("bool", 1)
		}
	case Var:
		// The type of variables are defined at Assgin node.
	case ArrayRef:
		addType(n.lhs)
		addType(n.rhs)
		if n.ty.kind == TY_PTR {
			return
		}
		n.ty = newLiteral("pointer", 1)
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
			panic(fmt.Sprintf("not same length %d != %d", len(n.lvals), len(n.rvals)))
		}
		for i := range n.lvals {
			// Add type from right-side node.
			addType(n.rvals[i])
			if n.lvals[i].getType().kind == TY_NONE {
				n.lvals[i].setType(*n.rvals[i].getType())
			}
			addType(n.lvals[i])
		}
	default:
		panic(fmt.Sprintf("unexpected node type %#v", n))
	}
}
