package main

import (
	"fmt"
)

var labelseq int = 1
var argreg = []string{"rdi", "rsi", "rdx", "rcx", "r8", "r9"}
var funcname string

func genAddr(node interface{}) {
	if n, ok := node.(Var); ok {
		fmt.Printf("  lea rax, [rbp-%d]\n", n.offset)
		fmt.Printf("  push rax\n")
		return
	}
	panic("Not a lvalue")
}

func load() {
	fmt.Printf("  pop rax\n")
	fmt.Printf("  mov rax, [rax]\n")
	fmt.Printf("  push rax\n")
}

func store() {
	fmt.Printf("  pop rdi\n")
	fmt.Printf("  pop rax\n")
	fmt.Printf("  mov [rax], rdi\n")
	fmt.Printf("  push rdi\n")
}

func isEmpty(node interface{}) bool {
	if node == nil {
		return true
	}
	switch node.(type) {
	case Empty:
		return true
	}
	return false
}

func gen(node interface{}) {
	switch n := node.(type) {
	case Empty:
		return
	case IntLit:
		fmt.Printf("  push %d\n", n)
		return
	case ExprStmt:
		gen(n.child)
		fmt.Printf("  add rsp, 8\n") // Throw away the result of an expression.
		return
	case Var:
		genAddr(n)
		load()
		return
	case Assign:
		genAddr(n.lhs)
		gen(n.rhs)
		store()
		return
	case Block:
		for _, c := range n.children {
			gen(c)
		}
		return
	case If:
		seq := labelseq
		labelseq++
		if n.init != nil {
			gen(n.init)
		}
		if n.els != nil {
			gen(n.cond)
			fmt.Printf("  pop rax\n")
			fmt.Printf("  cmp rax, 0\n")
			fmt.Printf("  je  .Lelse%d\n", seq)
			gen(n.then)
			fmt.Printf("  jmp .Lend%d\n", seq)
			fmt.Printf(".Lelse%d:\n", seq)
			gen(n.els)
			fmt.Printf(".Lend%d:\n", seq)
		} else {
			gen(n.cond)
			fmt.Printf("  pop rax\n")
			fmt.Printf("  cmp rax, 0\n")
			fmt.Printf("  je  .Lend%d\n", seq)
			gen(n.then)
			fmt.Printf(".Lend%d:\n", seq)
		}
		return
	case For:
		seq := labelseq
		labelseq++
		if !isEmpty(n.init) {
			gen(n.init)
		}
		fmt.Printf(".Lbegin%d:\n", seq)
		if !isEmpty(n.cond) {
			gen(n.cond)
			fmt.Printf("  pop rax\n")
			fmt.Printf("  cmp rax, 0\n")
			fmt.Printf("  je  .Lend%d\n", seq)
		}
		gen(n.then)
		if !isEmpty(n.post) {
			gen(n.post)
		}
		fmt.Printf("  jmp .Lbegin%d\n", seq)
		fmt.Printf(".Lend%d:\n", seq)
		return
	case Return:
		gen(n.child)
		fmt.Printf("  pop rax\n")
		fmt.Printf("  jmp .Lreturn.%s\n", funcname)
		return
	case FuncCall:
		for _, arg := range n.args {
			gen(arg)
		}
		for i := len(n.args) - 1; i >= 0; i-- {
			fmt.Printf("  pop %s\n", argreg[i])
		}
		fmt.Printf("  call %s\n", n.name)
		fmt.Printf("  push rax\n")
		return
	}

	n := node.(Binary)
	gen(n.lhs)
	gen(n.rhs)
	fmt.Printf("  pop rdi\n")
	fmt.Printf("  pop rax\n")

	switch n.op {
	case "+":
		fmt.Printf("  add rax, rdi\n")
	case "-":
		fmt.Printf("  sub rax, rdi\n")
	case "*":
		fmt.Printf("  imul rax, rdi\n")
	case "/":
		fmt.Printf("  cqo\n")
		fmt.Printf("  idiv rdi\n")
	case "==":
		fmt.Printf("  cmp rax, rdi\n")
		fmt.Printf("  sete al\n")
		fmt.Printf("  movzx rax, al\n")
	case "!=":
		fmt.Printf("  cmp rax, rdi\n")
		fmt.Printf("  setne al\n")
		fmt.Printf("  movzx rax, al\n")
	case "<":
		fmt.Printf("  cmp rax, rdi\n")
		fmt.Printf("  setl al\n")
		fmt.Printf("  movzx rax, al\n")
	case "<=":
		fmt.Printf("  cmp rax, rdi\n")
		fmt.Printf("  setle al\n")
		fmt.Printf("  movzx rax, al\n")
	default:
		panic("[Error] Unexpected node")
	}
	fmt.Printf("  push rax\n")
}

func codegen(prog Program) {
	fmt.Printf(".intel_syntax noprefix\n")

	for _, f := range prog.funcs {
		funcname = f.name
		fmt.Printf(".global %s\n", funcname)
		fmt.Printf("%s:\n", funcname)

		// Prologue.
		fmt.Printf("  push rbp\n")
		fmt.Printf("  mov rbp, rsp\n")
		fmt.Printf("  sub rsp, %d\n", f.stackSize)

		// Push parameters to the stack.
		for i, p := range f.params {
			fmt.Printf("  mov [rbp-%d], %s\n", p.offset, argreg[i])
		}

		// Emit code.
		for _, s := range f.stmts {
			gen(s)
		}

		// Epilogue.
		fmt.Printf(".Lreturn.%s:\n", funcname)
		fmt.Printf("  mov rsp, rbp\n")
		fmt.Printf("  pop rbp\n")
		fmt.Printf("  ret\n")
	}
}
