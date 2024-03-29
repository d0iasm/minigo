# minigo
Minimum Go compiler that aims to do self-hosting. Its grammar is based on the official specification (https://golang.org/ref/spec), but it only supports parts of them.

## Milestones
1. Calculate a fibonacci funcion ([88baba9](https://github.com/d0iasm/minigo/commit/88baba94a917221ffd6b4f15304ac560e7f2a6a8), 08/27/2019)
2.

## Grammars
```
TopLevelDecl = FunctionDecl

// Declarations.
FunctionDecl = "func" FunctionName Block
FunctionName = identifier
identifier = letter { letter | unicode_digit } .

// Statements.
Statement = SimpleStmt | ReturnStmt | Block | IfStmt | ForStmt

SimpleStmt = EmptyStmt | ExpressionStmt | Assignment

ReturnStmt = "return" Expression

Block = "{" StatementList "}"
StatementList = { Statement ";" }

IfStmt = "if" [ SimpleStmt ";" ] Expression Block [ "else" ( IfStmt | Block ) ]

ForStmt = "for" [ Condition | ForClause ] Block
Condition = Expression
ForClause = [ InitStmt ] ";" [ Condition ] ";" [ PostStmt ] .
InitStmt = SimpleStmt .
PostStmt = SimpleStmt .

// Expressions
Expression = UnaryExpr | Expression binary_op Expression
UnaryExpr  = unary_op UnaryExpr
unary_op   = "+" | "-" | "*" | "&"
binary_op  = rel_op | add_op | mul_op
rel_op     = "==" | "!=" | "<" | "<=" | ">" | ">="
add_op     = "+" | "-"
mul_op     = "*" | "/"
```

## References
- https://github.com/rui314/chibicc
- https://www.sigbus.info/compilerbook
