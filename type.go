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
	base   *Type
	size   int // default is 0.
	aryLen int // default is 1.
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

func typeSize(k TypeKind) int {
	switch k {
	case TY_BOOL:
		return 1
	case TY_INT32:
		return 4
	case TY_INT64:
		return 8
	case TY_STRING:
		return 0
	case TY_PTR:
		return 8
	case TY_ARRAY:
		return 0
	default:
		return 0
	}
}

func newNoneType() Type {
	return Type{TY_NONE, nil, 0, 1}
}

func newLiteralType(s string) Type {
	return Type{typeKind(s), nil, typeSize(typeKind(s)), 1}
}

func pointerTo(base *Type) Type {
	return Type{TY_PTR, base, 8, 1}
}

func arrayOf(base *Type, length int) Type {
	return Type{TY_ARRAY, base, length * typeSize(base.kind), length}
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
		n.setType(newLiteralType("int64"))
	case StringLit:
		if n.ty.kind == TY_STRING {
			return
		}
		n.setType(newLiteralType("string"))
	case Addr:
		addType(n.child)
		n.setType(pointerTo(n.child.getType()))
	case Deref:
		addType(n.child)
		if n.child.getType().kind == TY_PTR {
			n.setType(*n.child.getType().base)
			return
		}
		n.setType(newLiteralType("int64"))
	case Binary:
		addType(n.lhs)
		addType(n.rhs)
		typeCheck(n.lhs.getType(), n.rhs.getType(), n.op)
		//if n.lhs.getType().kind == TY_PTR || n.rhs.getType().kind == TY_PTR {
		//panic(fmt.Sprintf("invalid operation %#v %s %#v", n.lhs.getType(), n.op, n.rhs.getType()))
		//}
		switch n.op {
		case "+", "-", "*", "/":
			n.setType(*n.lhs.getType())
		case "==", "!=", "<", "<=":
			n.setType(newLiteralType("bool"))
		}
	case Var:
		// The type of variables are defined at Assgin node.
	case ArrayRef:
		addType(n.lhs)
		addType(n.rhs)
		if n.ty.kind == TY_PTR {
			return
		}
		n.setType(newLiteralType("pointer"))
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
