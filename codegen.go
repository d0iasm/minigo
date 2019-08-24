package main

import (
	"fmt"
	"os"
)

func gen(node *Node) {
	switch node.kind {
	case ND_NUM:
		fmt.Printf("  push %d\n", node.val)
		return
	case ND_EXPR_STMT:
		gen(node.lhs)                // Use only left-side child node for expression statement.
		fmt.Printf("  add rsp, 8\n") // Throw away the result of an expression.
		return
	case ND_RETURN:
		gen(node.lhs) // Use only left-side child node for return statement.
		fmt.Printf("  pop rax\n")
		fmt.Printf("  ret\n")
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
	fmt.Printf(".global _main\n")
	fmt.Printf("_main:\n")

	for _, n := range nodes {
		gen(n)
	}

	fmt.Printf("  ret\n")
}
