package main

import (
	"fmt"
	"os"
)

func genAddr(node *Node) {
	if node.kind == ND_LVAR {
		// TODO: node.name now only takes one character.
		offset := (int(node.name[0]) - int('a') + 1) * 8
		fmt.Printf("  lea rax, [rbp-%d]\n", offset)
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

func gen(node *Node) {
	switch node.kind {
	case ND_NUM:
		fmt.Printf("  push %d\n", node.val)
		return
	case ND_EXPR_STMT:
		gen(node.lhs)                // Use only left-side child node for expression statement.
		fmt.Printf("  add rsp, 8\n") // Throw away the result of an expression.
		return
	case ND_LVAR:
		genAddr(node)
		load()
		return
	case ND_ASSIGN:
		genAddr(node.lhs)
		gen(node.rhs)
		store()
		return
	case ND_RETURN:
		gen(node.lhs) // Use only left-side child node for return statement.
		fmt.Printf("  pop rax\n")
		fmt.Printf("  jmp .Lreturn\n")
		return
	}

	gen(node.lhs)
	gen(node.rhs)

	fmt.Printf("  pop rdi\n")
	fmt.Printf("  pop rax\n")

	switch node.kind {
	case ND_ADD:
		fmt.Printf("  add rax, rdi\n")
	case ND_SUB:
		fmt.Printf("  sub rax, rdi\n")
	case ND_MUL:
		fmt.Printf("  imul rax, rdi\n")
	case ND_DIV:
		fmt.Printf("  cqo\n")
		fmt.Printf("  idiv rdi\n")
	case ND_EQ:
		fmt.Printf("  cmp rax, rdi\n")
		fmt.Printf("  sete al\n")
		fmt.Printf("  movzx rax, al\n")
	case ND_NE:
		fmt.Printf("  cmp rax, rdi\n")
		fmt.Printf("  setne al\n")
		fmt.Printf("  movzx rax, al\n")
	case ND_LT:
		fmt.Printf("  cmp rax, rdi\n")
		fmt.Printf("  setl al\n")
		fmt.Printf("  movzx rax, al\n")
	case ND_LE:
		fmt.Printf("  cmp rax, rdi\n")
		fmt.Printf("  setle al\n")
		fmt.Printf("  movzx rax, al\n")
	default:
		fmt.Println("[Error] Unexpected node:", node)
		os.Exit(1)
	}
	fmt.Printf("  push rax\n")
}

func codegen(nodes []*Node) {
	fmt.Printf(".intel_syntax noprefix\n")
	fmt.Printf(".global main\n")
	fmt.Printf("main:\n")

	// Prologue
	fmt.Printf("  push rbp\n")
	fmt.Printf("  mov rbp, rsp\n")
	fmt.Printf("  sub rsp, 208\n")

	for _, n := range nodes {
		gen(n)
	}

	fmt.Printf(".Lreturn:\n")
	fmt.Printf("  mov rsp, rbp\n")
	fmt.Printf("  pop rbp\n")
	fmt.Printf("  ret\n")
}
