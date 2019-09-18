package main

import (
	"fmt"
)

var labelseq int = 1

// Comply with System V ABI.
// Lower 8-bit register (1 byte).
var argreg1 = []string{"dil", "sil", "dl", "cl", "r8b", "r9b"}

// 64-bit register (8 bytes).
var argreg8 = []string{"rdi", "rsi", "rdx", "rcx", "r8", "r9"}
var funcname string

func genAddr(node interface{}) {
	switch n := node.(type) {
	case *Var:
		if n.isLocal {
			fmt.Printf("  lea rax, [rbp-%d]\n", n.offset)
			fmt.Printf("  push rax\n")
		} else {
			fmt.Printf("  push offset %s\n", n.name)
		}
		return
	case *Deref:
		gen(n.child)
		return
	case *ArrayRef:
		if n.lhs.getType().kind == TY_STRING {
			genAddr(n.lhs)
			gen(n.rhs)
			// index stored in right-size node.
			fmt.Printf("  pop rdi\n")
			// pointer to string object stored in left-size node.
			fmt.Printf("  pop rax\n")
			fmt.Printf("  mov rax, [rax]\n")
			fmt.Printf("  add rax, rdi\n")
			fmt.Printf("  push rax\n")
			return
		}

		genAddr(n.lhs)
		gen(n.rhs)
		fmt.Printf("  pop rdi\n") // index stored in right-side node.
		fmt.Printf("  pop rax\n") // address stored in left-size node.
		fmt.Printf("  imul rdi, %d\n", n.ty.size)
		fmt.Printf("  add rax, rdi\n")
		fmt.Printf("  push rax\n")
		return
	case *StringLit:
		fmt.Printf("  push offset %s.obj\n", n.label)
		return
	}
	panic(fmt.Sprintf("not a lvalue %#v", node))
}

func load(ty *Type) {
	if ty.kind == TY_INT8 {
		fmt.Printf("  pop rax\n")
		fmt.Printf("  movzx rax, byte ptr [rax]\n")
		fmt.Printf("  push rax\n")
	} else if ty.kind == TY_STRING {
		fmt.Printf("  pop rax\n")
		fmt.Printf("  mov rdi, [rax]\n")
		fmt.Printf("  push rdi\n") // pointer to string
		fmt.Printf("  mov rdi, [rax+8]\n")
		fmt.Printf("  push rdi\n") // length of string
	} else {
		fmt.Printf("  pop rax\n")
		fmt.Printf("  mov rax, [rax]\n")
		fmt.Printf("  push rax\n")
	}
}

func store(ty *Type) {
	if ty.kind == TY_STRING {
		fmt.Printf("  pop rsi\n")
		fmt.Printf("  pop rdi\n")
		fmt.Printf("  pop rax\n")
		fmt.Printf("  mov [rax], rdi\n")
		fmt.Printf("  mov [rax+8], rsi\n")
		return
	}
	fmt.Printf("  pop rdi\n")
	fmt.Printf("  pop rax\n")
	fmt.Printf("  mov [rax], rdi\n")
}

func isEmpty(node interface{}) bool {
	if node == nil {
		return true
	}
	switch node.(type) {
	case *Empty:
		return true
	}
	return false
}

func gen(node interface{}) {
	switch n := node.(type) {
	case *Empty:
		return
	case *IntLit:
		fmt.Printf("  push %d\n", n.val)
		return
	case *StringLit:
		fmt.Printf("  push offset %s\n", n.label)
		fmt.Printf("  push %d\n", len(n.val))
		return
	case *Var:
		genAddr(n)
		load(n.ty)
		return
	case *Assign:
		if len(n.lvals) != len(n.rvals) {
			panic(fmt.Sprintf("Not same length %d != %d", len(n.lvals), len(n.rvals)))
		}
		for i := range n.lvals {
			genAddr(n.lvals[i])
			gen(n.rvals[i])
			store(n.lvals[i].getType())
		}
		return
	case *Addr:
		genAddr(n.child)
		return
	case *Deref:
		gen(n.child)
		load(n.ty)
		return
	case *ArrayRef:
		genAddr(n)
		load(n.ty)
		return
	case *Block:
		for _, c := range n.children {
			gen(c)
		}
		return
	case *ExprStmt:
		gen(n.child)
		// Throw away the result of an expression.
		fmt.Printf("  add rsp, 8\n")
		return
	case *If:
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
	case *For:
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
	case *Return:
		gen(n.child)
		fmt.Printf("  pop rax\n")
		fmt.Printf("  jmp .Lreturn.%s\n", funcname)
		return
	case *FuncCall:
		for _, arg := range n.args {
			gen(arg)
		}
		for i := len(n.args) - 1; i >= 0; i-- {
			fmt.Printf("  pop %s\n", argreg8[i])
		}

		// We need to align RSP to a 16 byte boundary before
		// calling a function because it is an ABI requirement.
		// RAX is set to 0 for variadic function.
		seq := labelseq
		labelseq++
		fmt.Printf("  mov rax, rsp\n")
		fmt.Printf("  and rax, 15\n")
		fmt.Printf("  jnz .Lcall%d\n", seq)
		fmt.Printf("  mov rax, 0\n")
		fmt.Printf("  call %s\n", n.name)
		fmt.Printf("  jmp .Lend%d\n", seq)
		fmt.Printf(".Lcall%d:\n", seq)
		fmt.Printf("  sub rsp, 8\n")
		fmt.Printf("  mov rax, 0\n")
		fmt.Printf("  call %s\n", n.name)
		fmt.Printf("  add rsp, 8\n")
		fmt.Printf(".Lend%d:\n", seq)
		fmt.Printf("  push rax\n")
		return
	case *Stdlib:
		for _, arg := range n.args {
			gen(arg)
		}
		fmt.Printf("  call %s\n", n.name)
		return
	}

	n := node.(*Binary)
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
		panic(fmt.Sprintf("unexpected node %#v", n))
	}
	fmt.Printf("  push rax\n")
}

func emitData(prog Program) {
	fmt.Printf(".data\n")

	for _, g := range prog.globals {
		fmt.Printf("%s:\n", g.name)
		fmt.Printf("  .zero %d\n", g.ty.aryLen*g.ty.size)
	}

	for _, c := range prog.contents {
		fmt.Printf("%s:\n", c.label)
		fmt.Printf("  .string \"%s\"\n", c.val)
		fmt.Printf("%s.obj:\n", c.label)
		fmt.Printf("  .quad %s\n", c.label)
		fmt.Printf("  .quad %d\n", len(c.val))
	}
}

func emitStdlibs() {
	// ssize_t write(int fd, const void *buf, size_t count);
	fmt.Printf("println:\n")
	// Prologue.
	fmt.Printf("  push rbp\n")
	fmt.Printf("  mov rbp, rsp\n")

	//fmt.Printf("  mov rsi, [rbp+16]\n")
	fmt.Printf("  mov rdx, [rbp+16]\n") // third arg: size_t count
	fmt.Printf("  mov rsi, [rbp+24]\n") // second arg: const void *buf
	fmt.Printf("  mov rdi, 1\n")        // first arg: int fd

	fmt.Printf("  call write\n")

	// Newline.
	fmt.Printf("  mov rdi, '\\n'\n")
	fmt.Printf("  call putchar\n")

	// Epilogue.
	fmt.Printf("  mov rsp, rbp\n")
	fmt.Printf("  pop rbp\n")
	fmt.Printf("  ret\n")
}

func emitText(prog Program) {
	fmt.Printf(".text\n")

	for _, f := range prog.funcs {
		fmt.Printf(".global %s\n", f.name)
	}

	emitStdlibs()

	for _, f := range prog.funcs {
		funcname = f.name
		fmt.Printf("%s:\n", funcname)

		if funcname == "main" {
			fmt.Printf("  call %s\n", prog.funcs[0].name)
		}

		// Prologue.
		fmt.Printf("  push rbp\n")
		fmt.Printf("  mov rbp, rsp\n")
		fmt.Printf("  sub rsp, %d\n", f.stackSize)

		// Push parameters to the stack.
		for i, p := range f.params {
			fmt.Printf("  mov [rbp-%d], %s\n", p.offset, argreg8[i])
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

func codegen(prog Program) {
	fmt.Printf(".intel_syntax noprefix\n")
	emitData(prog)
	emitText(prog)
}
