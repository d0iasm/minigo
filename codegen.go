package main

import (
	"fmt"
	"os"
)

var labelseq int = 1

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

func gen(node interface{}) {
	switch n := node.(type) {
	case int:
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
	//case If:
	//seq := labelseq
	//labelseq++
	case Return:
		gen(n.child)
		fmt.Printf("  pop rax\n")
		fmt.Printf("  jmp .Lreturn\n")
		return
	case Add:
		gen(n.lhs)
		gen(n.rhs)
		fmt.Printf("  pop rdi\n")
		fmt.Printf("  pop rax\n")

		fmt.Printf("  add rax, rdi\n")
	case Sub:
		gen(n.lhs)
		gen(n.rhs)
		fmt.Printf("  pop rdi\n")
		fmt.Printf("  pop rax\n")

		fmt.Printf("  sub rax, rdi\n")
	case Mul:
		gen(n.lhs)
		gen(n.rhs)
		fmt.Printf("  pop rdi\n")
		fmt.Printf("  pop rax\n")

		fmt.Printf("  imul rax, rdi\n")
	case Div:
		gen(n.lhs)
		gen(n.rhs)
		fmt.Printf("  pop rdi\n")
		fmt.Printf("  pop rax\n")

		fmt.Printf("  cqo\n")
		fmt.Printf("  idiv rdi\n")
	case Eq:
		gen(n.lhs)
		gen(n.rhs)
		fmt.Printf("  pop rdi\n")
		fmt.Printf("  pop rax\n")

		fmt.Printf("  cmp rax, rdi\n")
		fmt.Printf("  sete al\n")
		fmt.Printf("  movzx rax, al\n")
	case Ne:
		gen(n.lhs)
		gen(n.rhs)
		fmt.Printf("  pop rdi\n")
		fmt.Printf("  pop rax\n")

		fmt.Printf("  cmp rax, rdi\n")
		fmt.Printf("  setne al\n")
		fmt.Printf("  movzx rax, al\n")
	case Lt:
		gen(n.lhs)
		gen(n.rhs)
		fmt.Printf("  pop rdi\n")
		fmt.Printf("  pop rax\n")

		fmt.Printf("  cmp rax, rdi\n")
		fmt.Printf("  setl al\n")
		fmt.Printf("  movzx rax, al\n")
	case Le:
		gen(n.lhs)
		gen(n.rhs)
		fmt.Printf("  pop rdi\n")
		fmt.Printf("  pop rax\n")

		fmt.Printf("  cmp rax, rdi\n")
		fmt.Printf("  setle al\n")
		fmt.Printf("  movzx rax, al\n")
	default:
		fmt.Println("[Error] Unexpected node:", n)
		os.Exit(1)
	}
	fmt.Printf("  push rax\n")
}

func codegen(funcs []Function) {
	fmt.Printf(".intel_syntax noprefix\n")
	fmt.Printf(".global main\n")
	fmt.Printf("main:\n")

	// Prologue
	fmt.Printf("  push rbp\n")
	fmt.Printf("  mov rbp, rsp\n")
	fmt.Printf("  sub rsp, 208\n")

	for _, f := range funcs {
		for _, n := range f.nodes {
			gen(n)
		}
	}

	fmt.Printf(".Lreturn:\n")
	fmt.Printf("  mov rsp, rbp\n")
	fmt.Printf("  pop rbp\n")
	fmt.Printf("  ret\n")
}
