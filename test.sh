cat <<EOF | gcc -xc -c -o tmp2.o -
int ret3() { return 3; }
int ret5() { return 5; }
int add(int x, int y) { return x+y; }
int sub(int x, int y) { return x-y; }
int add6(int a, int b, int c, int d, int e, int f) {
  return a+b+c+d+e+f;
}
EOF

assert() {
  expected="$1"
  input="$2"

  go build main.go tokenize.go parse.go codegen.go type.go
  ./main -in "$input" > tmp.s
  gcc -static -o tmp tmp.s tmp2.o
  ./tmp
  actual="$?"

  if [ "$actual" = "$expected" ]; then
    echo "$input => $actual"
  else
    echo "$input => $expected expected, but got $actual"
    exit 1
  fi
}

echo
echo 'simple arithmetic'
echo
assert 0 'func main() { return 0; }'
assert 42 'func main() { return 42; }'
assert 21 'func main() { return 5+20-4; }'
assert 41 'func main() { return  12 + 34 - 5 ; }'
assert 47 'func main() { return 5+6*7; }'
assert 15 'func main() { return 5*(9-6); }'
assert 4 'func main() { return (3+5)/2; }'
assert 10 'func main() { return -10+20; }'
assert 10 'func main() { return - -10; }'
assert 10 'func main() { return - - +10; }'

echo
echo 'equality operators'
echo
assert 0 'func main() { return 0==1; }'
assert 1 'func main() { return 42==42; }'
assert 1 'func main() { return 0!=1; }'
assert 0 'func main() { return 42!=42; }'

echo
echo 'relational operators'
echo
assert 1 'func main() { return 0<1; }'
assert 0 'func main() { return 1<1; }'
assert 0 'func main() { return 2<1; }'
assert 1 'func main() { return 0<=1; }'
assert 1 'func main() { return 1<=1; }'
assert 0 'func main() { return 2<=1; }'

assert 1 'func main() { return 1>0; }'
assert 0 'func main() { return 1>1; }'
assert 0 'func main() { return 1>2; }'
assert 1 'func main() { return 1>=0; }'
assert 1 'func main() { return 1>=1; }'
assert 0 'func main() { return 1>=2; }'

echo
echo 'assignments'
echo
assert 3 'func main() { a:=3; return a; }'
assert 8 'func main() { a:=3; z:=5; return a+z; }'

assert 1 'func main() { return 1; 2; 3; }'
assert 2 'func main() { 1; return 2; 3; }'
assert 3 'func main() { 1; 2; return 3; }'

assert 3 'func main() { foo:=3; return foo; }'
assert 8 'func main() { foo123:=3; bar:=5; return foo123+bar; }'

echo
echo 'blocks'
echo
assert 3 'func main() { {1; {2;} return 3;} }'

echo
echo 'if statements'
echo
assert 3 'func main() { if 0 {return 2;} return 3; }'
assert 3 'func main() { if 1-1 {return 2;} return 3; }'
assert 2 'func main() { if 1 {return 2;} return 3; }'
assert 2 'func main() { if 2-1 {return 2;} return 3; }'

assert 3 'func main() { if 0 {return 2;} else if 1 {return 3;} }'
assert 1 'func main() { if 1 { if 2 { return 1; } } }'
assert 2  'func main() { if 1 { if 0 { return 1; } return 2; } }'
assert 1  'func main() { if i:=1; i>0 { return i; } }'
assert 3  'func main() { if i:=1; i>0 { i=i+2; return i; } }'

echo
echo 'for statements'
echo
assert 3 'func main() { for { return 3; } }'
assert 10 'func main() { i:=0; for i<10 { i=i+2; i=i-1; } return i; }'
assert 10 'func main() { for i:=0; i<10; i=i+1 { 1; } return i; }'
assert 10 'func main() { i:=0; for ; i<10; i=i+1; { i=i; } return i; }'
assert 10 'func main() { for i:=0; i<10; ; { i=i+1; } return i; }'
assert 11 'func main() { for i:=0; ; i=i+1; { if i>10 { return i; } } }'

echo
echo 'function'
echo
assert 3 'func main() { return ret3(); }'
assert 5 'func main() { return ret5(); }'
assert 8 'func main() { return add(3, 5); }'
assert 2 'func main() { return sub(5, 3); }'
assert 21 'func main() { return add6(1,2,3,4,5,6); }'

assert 32 'func main() { return ret32(); } func ret32() { return 32; }'
assert 7 'func main() { return add2(3,4); } func add2(x int64, y int64) { return x+y; }'
assert 1 'func main() { return sub2(4,3); } func sub2(x int64, y int64) { return x-y; }'
assert 55 'func main() { return fib(9); } func fib(x int64) { if x<=1 { return 1; } return fib(x-1) + fib(x-2); }'

echo
echo 'pointers'
echo
assert 3 'func main() { x:=3; return *&x; }'
assert 3 'func main() { x:=3; y:=&x; return *y; }'
assert 3 'func main() { x:=3; y:=&x; z:=&y; return **z; }'
assert 5 'func main() { x:=3; y:=&x; *y=5; return x; }'

echo
echo 'declarations'
echo
assert 2 'func main() { x:=2; return x; }'
assert 2 'func main() { x:=5; y:=3; return x-y; }'

assert 42 'func main() { var x int64; x=42; return x; }'
assert 3 'func main() { var x int64=3; return x; }'
assert 3 'func main() { var x int64=5; var y int64=2; return x-y; }'
assert 4 'func main() { var x int64; x=3; return x+1; }'
assert 4 'func main() { var x int64; x=3; var y=1; return x+y; }'
assert 4 'func main() { var x int64; x=3; y:=1; return x+y; }'

echo
echo 'arrays'
echo
assert 1 'func main() { var x [2]int64; x[0]=1; x[1]=2; return x[0]; }'
assert 2 'func main() { var x [2]int64; x[0]=1; x[1]=2; return x[1]; }'
assert 3 'func main() { var x [2]int64; x[0]=2; x[1]=5; return x[1]-x[0]; }'
assert 2 'func main() { var x [2]int64 = [2]int64{1, 2}; return x[1]; }'
assert 3 'func main() { x:=[2]int64{2, 5}; return x[1]-x[0]; }'

# Should fail but work well for some reason.
#assert 42 'func main() { var x[2]int64; *x=42; return *x; }'
#assert 42 'func main() { var x[2]int64; *(x+1)=42; return *(x+1); }'
#assert 3 'func main() { var x[2]int64; *x=1; *(x+1)=2; return *x + *(x+1); }'

echo
echo 'global variables'
echo
assert 3 'var a int64; func main() { a=3; return a; }'
assert 5 'var a int64=5; func main() { return a; }'
assert 3 'var a [3]int64=[3]int64{1,2,3}; func main() { return a[2]; }'

echo OK
