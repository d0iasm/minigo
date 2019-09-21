package main

func assert(expected int, actual int, code string) {
	if expected == actual {
		println("ok")
	} else {
		println(code)
	}
}

// TODO: add return type
func ret3() {
	return 3
}

func ret5() {
	return 5
}

func add(x int, y int) {
	return x + y
}

func sub(x int, y int) {
	return x - y
}

func add6(a int, b int, c int, d int, e int, f int) {
	return a + b + c + d + e + f
}

func fib(x int) {
	if x <= 1 {
		return 1
	}
	return fib(x-1) + fib(x-2)
}

func main() {
	println("simple arithmetic")
	assert(0, 0, "0")
	assert(42, 42, "42")
	assert(5, 5, "0")
	assert(21, 5+20-4, "5+20-4")
	assert(41, 12+34-5, "12+34-5")
	assert(15, 5*(9-6), "5*(9-6)")
	assert(4, (3+5)/2, "(3+5)/2")
	assert(10, -10+20, "-10+20")
	assert(10, - -10, "- -10")
	assert(10, - -+10, "- - +10")

	println("equality operators")
	assert(0, 0 == 1, "0==1")
	assert(1, 42 == 42, "42==42")
	assert(1, 0 != 1, "0!=1")
	assert(0, 42 != 42, "42!=42")

	println("relational operators")
	assert(1, 0 < 1, "0<1")
	assert(0, 1 < 1, "1<1")
	assert(0, 2 < 1, "2<1")
	assert(1, 0 <= 1, "0<=1")
	assert(1, 1 <= 1, "1<=1")
	assert(0, 2 <= 1, "2<=1")

	assert(1, 1 > 0, "1>0")
	assert(0, 1 > 1, "1>1")
	assert(0, 1 > 2, "1>2")
	assert(1, 1 >= 0, "1>=0")
	assert(1, 1 >= 1, "1>=1")
	assert(0, 1 >= 2, "1>=2")
}
