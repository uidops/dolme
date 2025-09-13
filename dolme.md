# DOLME Grammar and LL(1) Parsing Table

This document describes the grammar, FIRST/FOLLOW sets, and the LL(1) parsing table derived from `pkg/parser/table.go` (the current `NewParsingTable` implementation). Semantic actions (those beginning with `@`) are part of the parser's runtime behavior and are not real tokens of the DOLME language; they appear in productions to indicate actions executed when a production is applied.

Where I show parse-table mappings I use human-readable token names (e.g. `func`, `let`, `id`, `num`, `+`, `-`, `(`, `)`, `{`, `}`, etc.) corresponding to the `lexer.TokenType` values used in `NewParsingTable`.

---

## 1. Context-Free Grammar (productions are numbered to match `pkg/parser/table.go`)

```dolme/dolme.md#L1-200
Program → DeclList                                       (1)

DeclList → Decl DeclList | ε                             (2,3)

Decl → FuncDecl | Stmt                                    (4,5)

FuncDecl → func id @func_start ( ParamList ) : Type @func_return_type { StmtList } @func_end
                                                          (6)

ParamList → Param Param' | ε                              (7,8)
Param' → , Param Param' | ε                               (9,10)
Param → id @capture_param_name : Type @capture_type @param (11)

Type → int | float | bool                                 (12,13,14)

StmtList → Stmt StmtList | ε                              (15,16)

Stmt → VarDecl | Assign | IfStmt | WhileStmt | PrintStmt | ReturnStmt | ContinueStmt | BreakStmt
                                                         (17..24)

VarDecl → let id @capture_decl_var : Type @capture_type = Expr ; @define
                                                         (25)

Assign → id @capture_assign_target AssignSuffix ;         (26)
AssignSuffix → = Expr @assign | ( ArgList ) @call        (27,28)

IfStmt → if ( Cond ) @save { StmtList } ElsePart         (29)
ElsePart → @jmpf @save else { StmtList } @jmp | @jmpf_normal ε
                                                         (30,31)

WhileStmt → while @label_while ( Cond ) @save { StmtList } @jmpf_break @jmp_nonbackpatch
                                                         (32)

ContinueStmt → continue ; @continue                       (33)
BreakStmt → break ; @save_break                           (34)

PrintStmt → print ( id @load ) ; @print                   (35)

ReturnStmt → return ReturnValue ; @return                 (36)
ReturnValue → Expr | ε                                    (37,38)

Expr → Term Expr'                                         (39)
Expr' → + Term Expr' @add | - Term Expr' @sub | ε        (40,41,42)

Term → Factor Term'                                       (43)
Term' → * Factor Term' @mul | / Factor Term' @div | % Factor Term' @mod | ε
                                                         (44,45,46,47)

Factor → id FactorSuffix | num @push | true @push | false @push | ( Expr )
                                                         (48..52)

FactorSuffix → @load | @call_start ( ArgList ) @call_end (53,54)

ArgList → Expr @arg ArgList' | ε                         (55,56)
ArgList' → , Expr @arg ArgList' | ε                      (57,58)

Cond → OrExpr                                            (59)

OrExpr → AndExpr OrExpr'                                 (60)
OrExpr' → or AndExpr OrExpr' @or | ε                     (61,62)

AndExpr → NotExpr AndExpr'                               (63)
AndExpr' → and NotExpr AndExpr' @and | ε                 (64,65)

NotExpr → not NotExpr @not | RelExpr                     (66,67)

RelExpr → BoolPrimary RelExpr'                           (68)
RelExpr' → RelOp BoolPrimary @rel | ε                    (69,70)

BoolPrimary → true @push | false @push | Expr            (71,72,73)

RelOp → < @push_relop | > @push_relop | <= @push_relop | >= @push_relop | == @push_relop | != @push_relop
                                                         (74..79)
```

Notes:
- Semantic action symbols beginning with `@` are not terminals in the lexer; they are executed by the parser when the associated production is reduced/applied.
- Production numbers in parentheses correspond to indices in the `grammer` slice in `pkg/parser/table.go` (the first non-empty production is index 1).

---

## 2. FIRST sets (real tokens only; `ε` included where applicable)

- FIRST(Program)      = { func, let, id, if, while, print, return }
- FIRST(DeclList)     = { func, let, id, if, while, print, return, ε }
- FIRST(Decl)         = { func, let, id, if, while, print, return }
- FIRST(FuncDecl)     = { func }
- FIRST(ParamList)    = { id, ε }
- FIRST(Param')       = { ,, ε }
- FIRST(Param)        = { id }
- FIRST(Type)         = { int, float, bool }
- FIRST(StmtList)     = { let, id, if, while, print, return, continue, break, ε }
- FIRST(Stmt)         = { let, id, if, while, print, return, continue, break }
- FIRST(VarDecl)      = { let }
- FIRST(Assign)       = { id }
- FIRST(AssignSuffix) = { =, ( }
- FIRST(IfStmt)       = { if }
- FIRST(ElsePart)     = { else, ε }     # semantic actions precede `else` internally
- FIRST(WhileStmt)    = { while }
- FIRST(ContinueStmt) = { continue }
- FIRST(BreakStmt)    = { break }
- FIRST(PrintStmt)    = { print }
- FIRST(ReturnStmt)   = { return }
- FIRST(ReturnValue)  = { id, num, (, true, false, not, ε }
- FIRST(Expr)         = { id, num, (, true, false }
- FIRST(Expr')        = { +, -, ε }
- FIRST(Term)         = { id, num, (, true, false }
- FIRST(Term')        = { *, /, %, ε }
- FIRST(Factor)       = { id, num, true, false, ( }
- FIRST(FactorSuffix) = { (, ε }        # '(' → call; ε → @load
- FIRST(ArgList)      = { id, num, (, true, false, not, ε }
- FIRST(ArgList')     = { ,, ε }
- FIRST(Cond)         = { not, id, num, (, true, false }
- FIRST(OrExpr)       = { not, id, num, (, true, false }
- FIRST(OrExpr')      = { or, ε }
- FIRST(AndExpr)      = { not, id, num, (, true, false }
- FIRST(AndExpr')     = { and, ε }
- FIRST(NotExpr)      = { not, id, num, (, true, false }
- FIRST(RelExpr)      = { id, num, (, true, false }
- FIRST(RelExpr')     = { <, >, <=, >=, ==, !=, ε }
- FIRST(BoolPrimary)  = { true, false, id, num, ( }
- FIRST(RelOp)        = { <, >, <=, >=, ==, != }

Implementation note:
- `ReturnValue` and `ArgList` include `not` in their FIRST sets and table entries (they dispatch to production 37 / 55 when `not` is seen). However, `Expr` itself does not include `not` in its FIRST set or table entry. This creates an inconsistency: a return value or argument starting with `not` reduces to a production that expects `Expr`, but `Expr` has no `not` entry in the parsing table. See section 5 (Consistency Notes) for details.

---

## 3. FOLLOW sets (symbols are tokens; `$` denotes EOF)

- FOLLOW(Program)     = { $ }
- FOLLOW(DeclList)    = { $ }
- FOLLOW(Decl)        = { func, let, id, if, while, print, return, $ }
- FOLLOW(FuncDecl)    = FOLLOW(Decl)
- FOLLOW(ParamList)   = { ) }
- FOLLOW(Param')      = { ) }
- FOLLOW(Param)       = { ,, ) }
- FOLLOW(Type)        = { =, {, ,, ) }
- FOLLOW(StmtList)    = { } }  (i.e. right brace and what follows the surrounding construct)
- FOLLOW(Stmt)        = { let, id, if, while, print, return, continue, break, }, $ }
- FOLLOW(VarDecl)     = FOLLOW(Stmt)
- FOLLOW(Assign)      = FOLLOW(Stmt)
- FOLLOW(IfStmt)      = FOLLOW(Stmt)
- FOLLOW(WhileStmt)   = FOLLOW(Stmt)
- FOLLOW(PrintStmt)   = FOLLOW(Stmt)
- FOLLOW(ReturnStmt)  = FOLLOW(Stmt)
- FOLLOW(ContinueStmt)= FOLLOW(Stmt)
- FOLLOW(BreakStmt)   = FOLLOW(Stmt)
- FOLLOW(ElsePart)    = FOLLOW(IfStmt)
- FOLLOW(AssignSuffix)= { ; }
- FOLLOW(ReturnValue) = { ; }
- FOLLOW(Expr)        = { ;, ), <, >, <=, >=, ==, !=, and, or, ,, } }
- FOLLOW(Expr')       = FOLLOW(Expr)
- FOLLOW(Term)        = { +, -, ;, ), <, >, <=, >=, ==, !=, and, or, ,, } }
- FOLLOW(Term')       = FOLLOW(Term)
- FOLLOW(Factor)      = { *, /, %, +, -, ;, ), <, >, <=, >=, ==, !=, and, or, ,, } }
- FOLLOW(FactorSuffix)= FOLLOW(Factor)
- FOLLOW(ArgList)     = { ) }
- FOLLOW(ArgList')    = FOLLOW(ArgList)
- FOLLOW(Cond)        = { ) }
- FOLLOW(OrExpr)      = FOLLOW(Cond) = { ) }
- FOLLOW(OrExpr')     = FOLLOW(OrExpr) = { ) }
- FOLLOW(AndExpr)     = { or, ) }
- FOLLOW(AndExpr')    = FOLLOW(AndExpr) = { or, ) }
- FOLLOW(NotExpr)     = { and, or, ) }
- FOLLOW(RelExpr)     = FOLLOW(NotExpr) = { and, or, ) }
- FOLLOW(RelExpr')    = FOLLOW(RelExpr) = { and, or, ) }
- FOLLOW(BoolPrimary) = { <, >, <=, >=, ==, !=, and, or, ) }
- FOLLOW(RelOp)       = { id, num, (, true, false }  # FIRST(BoolPrimary)

---

## 4. LL(1) Parsing Table (compact mapping)

Below are the entries in the `ParsingTable` returned by `NewParsingTable()`. Each mapping is: Non-Terminal → { lookahead token: production-number }.

Program
- func, let, id, if, while, print, return → 1

DeclList
- func, let, id, if, while, print, return → 2
- EOF ($) → 3

Decl
- func → 4
- let, id, if, while, print, return → 5

FuncDecl
- func → 6

ParamList
- id → 7
- ) → 8

Param'
- , → 9
- ) → 10

Param
- id → 11

Type
- int → 12
- float → 13
- bool → 14

StmtList
- let, id, if, while, print, return, continue, break → 15
- } (RBRACE) → 16

Stmt
- let → 17
- id → 18
- if → 19
- while → 20
- print → 21
- return → 22
- continue → 23
- break → 24

VarDecl
- let → 25

Assign
- id → 26

AssignSuffix
- = (ASSIGN) → 27
- ( (LPAREN) → 28

IfStmt
- if → 29

ElsePart
- else → 30
- let, id, if, while, print, return, continue, break, }, EOF → 31

WhileStmt
- while → 32

ContinueStmt
- continue → 33

BreakStmt
- break → 34

PrintStmt
- print → 35

ReturnStmt
- return → 36

ReturnValue
- id, num, (, true, false, not → 37
- ; (SEMICOLON) → 38

Expr
- id, num, (, true, false → 39

Expr'
- + → 40
- - → 41
- ), ;, <, >, <=, >=, ==, !=, and, or, , → 42

Term
- id, num, (, true, false → 43

Term'
- * → 44
- / → 45
- % → 46
- +, -, ), ;, <, >, <=, >=, ==, !=, and, or, , → 47

Factor
- id → 48
- num → 49
- true → 50
- false → 51
- ( → 52

FactorSuffix
- ( → 54
- *, /, %, +, -, ), ;, <, >, <=, >=, ==, !=, and, or, , → 53

ArgList
- id, num, (, true, false, not → 55
- ) → 56

ArgList'
- , → 57
- ) → 58

Cond
- id, num, (, not, true, false → 59

OrExpr
- id, num, (, not, true, false → 60

OrExpr'
- or → 61
- ) → 62

AndExpr
- id, num, (, not, true, false → 63

AndExpr'
- and → 64
- or, ) → 65

NotExpr
- not → 66
- id, num, (, true, false → 67

RelExpr
- id, num, (, true, false → 68

RelExpr'
- <, >, <=, >=, ==, != → 69
- and, or, ) → 70

BoolPrimary
- true → 71
- false → 72
- (, id, num → 73

RelOp
- < → 74
- > → 75
- <= → 76
- >= → 77
- == → 78
- != → 79

Legend:
- Numeric entries are production indices in the `grammer` slice (as declared in `pkg/parser/table.go`).
- Token names are the human-readable equivalents of `lexer.TokenType` constants used in the parser.
- `EOF` / `$` denotes end-of-input.

This compact mapping is a direct transcription of `NewParsingTable()` in `pkg/parser/table.go`. It avoids positional matrix formatting and instead lists explicit lookahead → production mappings.

---

## 5. Consistency / Quality Notes

1. "not" inconsistency:
   - `ReturnValue` and `ArgList` accept `not` as a valid lookahead (they map `not` → production 37 / 55). Those productions route to `Expr` (`ReturnValue → Expr`) or to `Expr` via `ArgList → Expr ...`.
   - However, `Expr` does not include `not` in its FIRST set or parsing table entries — `NotExpr` and boolean operators live under `Cond` / `OrExpr` / `AndExpr` / `NotExpr`. In practice this means a return value or argument that begins with `not` will select the `ReturnValue` / `ArgList` entry but then attempt to parse an `Expr`, for which there is no `not`-entry in the table. That will cause a parse error for inputs like `return not ...` or `foo(not ...)` under the current table.
   - Fix options:
     - Extend `Expr` to include boolean operators (unify arithmetic and boolean expressions), or
     - Change `ReturnValue` and `ArgList` to use `Cond` (or a unified `Expression`) instead of `Expr` to allow `not` and boolean operators.

2. Semantic actions:
   - Productions that begin with semantic actions (e.g. `ElsePart` begins with `@jmpf`) mean the first real token for lookahead is still `else` or ε. For documentation and FIRST/FOLLOW reasoning, semantic actions are ignored.

3. Suggested refactor:
   - Introduce a unified `Expression` non-terminal that covers both arithmetic and boolean expressions with precedence. This will remove duplication and the `not` inconsistency, making parse-table entries clearer and safer.

---

## 6. Next steps / utilities

- If you'd like, I can:
  - generate a CSV/HTML representation of the compact table (easier to review),
  - add a small utility in `cmd/` or `pkg/parser` that prints the `ParsingTable` produced by `NewParsingTable()` (useful to keep documentation in sync with code), or
  - submit a small grammar refactor patch that unifies `Expr` and `Cond` (addresses the `not` inconsistency).

If you prefer a full matrix layout (with columns in a specific order) instead of the compact mapping above, tell me which token order you want and I can produce it (or generate a machine-readable CSV from the `ParsingTable` in code).