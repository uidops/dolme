package parser

import "dolme/pkg/lexer"

type Production struct {
	LHS string
	RHS []string
}

type ParsingTable map[string]map[lexer.TokenType]Production

var grammer = []Production{
	{}, // 0 - empty
	{LHS: "Program", RHS: []string{"DeclList"}}, // 1

	{LHS: "DeclList", RHS: []string{"Decl", "DeclList"}}, // 2
	{LHS: "DeclList", RHS: []string{"ε"}},                // 3

	{LHS: "Decl", RHS: []string{"FuncDecl"}}, // 4
	{LHS: "Decl", RHS: []string{"Stmt"}},     // 5

	{LHS: "FuncDecl", RHS: []string{"func", "id", "@func_start", "(", "ParamList", ")", ":", "Type", "@func_return_type", "{", "StmtList", "}", "@func_end"}}, // 6

	{LHS: "ParamList", RHS: []string{"Param", "Param'"}}, // 7
	{LHS: "ParamList", RHS: []string{"ε"}},               // 8

	{LHS: "Param'", RHS: []string{",", "Param", "Param'"}}, // 9
	{LHS: "Param'", RHS: []string{"ε"}},                    // 10

	{LHS: "Param", RHS: []string{"id", "@capture_param_name", ":", "Type", "@capture_type", "@param"}}, // 11

	{LHS: "Type", RHS: []string{"int"}},   // 12
	{LHS: "Type", RHS: []string{"float"}}, // 13
	{LHS: "Type", RHS: []string{"bool"}},  // 14

	{LHS: "StmtList", RHS: []string{"Stmt", "StmtList"}}, // 15
	{LHS: "StmtList", RHS: []string{"ε"}},                // 16

	{LHS: "Stmt", RHS: []string{"VarDecl"}},      // 17
	{LHS: "Stmt", RHS: []string{"Assign"}},       // 18
	{LHS: "Stmt", RHS: []string{"IfStmt"}},       // 19
	{LHS: "Stmt", RHS: []string{"WhileStmt"}},    // 20
	{LHS: "Stmt", RHS: []string{"PrintStmt"}},    // 21
	{LHS: "Stmt", RHS: []string{"ReturnStmt"}},   // 22
	{LHS: "Stmt", RHS: []string{"ContinueStmt"}}, // 23
	{LHS: "Stmt", RHS: []string{"BreakStmt"}},    // 24

	{LHS: "VarDecl", RHS: []string{"let", "id", "@capture_decl_var", ":", "Type", "@capture_type", "=", "Expr", ";", "@define"}}, // 25

	{LHS: "Assign", RHS: []string{"id", "@capture_assign_target", "AssignSuffix", ";"}}, // 26

	{LHS: "AssignSuffix", RHS: []string{"=", "Expr", "@assign"}},       // 27
	{LHS: "AssignSuffix", RHS: []string{"(", "ArgList", ")", "@call"}}, // 28

	{LHS: "IfStmt", RHS: []string{"if", "(", "Cond", ")", "@save", "{", "StmtList", "}", "ElsePart"}}, // 29

	{LHS: "ElsePart", RHS: []string{"@jmpf", "@save", "else", "{", "StmtList", "}", "@jmp"}}, // 30
	{LHS: "ElsePart", RHS: []string{"@jmpf_normal", "ε"}},                                    // 31

	{LHS: "WhileStmt", RHS: []string{"while", "@label_while", "(", "Cond", ")", "@save", "{", "StmtList", "}", "@jmpf_break", "@jmp_nonbackpatch"}}, // 32

	{LHS: "ContinueStmt", RHS: []string{"continue", ";", "@continue"}}, // 33

	{LHS: "BreakStmt", RHS: []string{"break", ";", "@save_break"}}, // 34

	{LHS: "PrintStmt", RHS: []string{"print", "(", "id", "@load", ")", ";", "@print"}}, // 35

	{LHS: "ReturnStmt", RHS: []string{"return", "ReturnValue", ";", "@return"}}, // 36

	{LHS: "ReturnValue", RHS: []string{"Expr"}}, // 37
	{LHS: "ReturnValue", RHS: []string{"ε"}},    // 38

	{LHS: "Expr", RHS: []string{"Term", "Expr'"}}, // 39

	{LHS: "Expr'", RHS: []string{"+", "Term", "Expr'", "@add"}}, // 40
	{LHS: "Expr'", RHS: []string{"-", "Term", "Expr'", "@sub"}}, // 41
	{LHS: "Expr'", RHS: []string{"ε"}},                          // 42

	{LHS: "Term", RHS: []string{"Factor", "Term'"}}, // 43

	{LHS: "Term'", RHS: []string{"*", "Factor", "Term'", "@mul"}}, // 44
	{LHS: "Term'", RHS: []string{"/", "Factor", "Term'", "@div"}}, // 45
	{LHS: "Term'", RHS: []string{"%", "Factor", "Term'", "@mod"}}, // 46
	{LHS: "Term'", RHS: []string{"ε"}},                            // 47

	// Left-factored Factor and FactorSuffix productions
	{LHS: "Factor", RHS: []string{"id", "FactorSuffix"}}, // 48
	{LHS: "Factor", RHS: []string{"num", "@push"}},       // 49
	{LHS: "Factor", RHS: []string{"true", "@push"}},      // 50
	{LHS: "Factor", RHS: []string{"false", "@push"}},     // 51
	{LHS: "Factor", RHS: []string{"(", "Expr", ")"}},     // 52

	{LHS: "FactorSuffix", RHS: []string{"@load"}},                                         // 53
	{LHS: "FactorSuffix", RHS: []string{"@call_start", "(", "ArgList", ")", "@call_end"}}, // 54

	{LHS: "ArgList", RHS: []string{"Expr", "@arg", "ArgList'"}}, // 55
	{LHS: "ArgList", RHS: []string{"ε"}},                        // 56

	{LHS: "ArgList'", RHS: []string{",", "Expr", "@arg", "ArgList'"}}, // 57
	{LHS: "ArgList'", RHS: []string{"ε"}},                             // 58

	{LHS: "Cond", RHS: []string{"OrExpr"}}, // 59

	{LHS: "OrExpr", RHS: []string{"AndExpr", "OrExpr'"}}, // 60

	{LHS: "OrExpr'", RHS: []string{"or", "AndExpr", "OrExpr'", "@or"}}, // 61
	{LHS: "OrExpr'", RHS: []string{"ε"}},                               // 62

	{LHS: "AndExpr", RHS: []string{"NotExpr", "AndExpr'"}}, // 63

	{LHS: "AndExpr'", RHS: []string{"and", "NotExpr", "AndExpr'", "@and"}}, // 64
	{LHS: "AndExpr'", RHS: []string{"ε"}},                                  // 65

	{LHS: "NotExpr", RHS: []string{"not", "NotExpr", "@not"}}, // 66
	{LHS: "NotExpr", RHS: []string{"RelExpr"}},                // 67

	{LHS: "RelExpr", RHS: []string{"BoolPrimary", "RelExpr'"}}, // 68

	{LHS: "RelExpr'", RHS: []string{"RelOp", "BoolPrimary", "@rel"}}, // 69
	{LHS: "RelExpr'", RHS: []string{"ε"}},                            // 70

	{LHS: "BoolPrimary", RHS: []string{"true", "@push"}},  // 71
	{LHS: "BoolPrimary", RHS: []string{"false", "@push"}}, // 72
	{LHS: "BoolPrimary", RHS: []string{"Expr"}},           // 73

	{LHS: "RelOp", RHS: []string{"<", "@push_relop"}},  // 74
	{LHS: "RelOp", RHS: []string{">", "@push_relop"}},  // 75
	{LHS: "RelOp", RHS: []string{"<=", "@push_relop"}}, // 76
	{LHS: "RelOp", RHS: []string{">=", "@push_relop"}}, // 77
	{LHS: "RelOp", RHS: []string{"==", "@push_relop"}}, // 78
	{LHS: "RelOp", RHS: []string{"!=", "@push_relop"}}, // 79
}

// NewParsingTable creates and returns a new LL(1) parsing table
func NewParsingTable() ParsingTable {
	return ParsingTable{
		"Program": {
			lexer.FUNC:   grammer[1],
			lexer.LET:    grammer[1],
			lexer.ID:     grammer[1],
			lexer.IF:     grammer[1],
			lexer.WHILE:  grammer[1],
			lexer.PRINT:  grammer[1],
			lexer.RETURN: grammer[1],
		},

		"DeclList": {
			lexer.FUNC:   grammer[2],
			lexer.LET:    grammer[2],
			lexer.ID:     grammer[2],
			lexer.IF:     grammer[2],
			lexer.WHILE:  grammer[2],
			lexer.PRINT:  grammer[2],
			lexer.RETURN: grammer[2],
			lexer.EOF:    grammer[3],
		},

		"Decl": {
			lexer.FUNC:   grammer[4],
			lexer.LET:    grammer[5],
			lexer.ID:     grammer[5],
			lexer.IF:     grammer[5],
			lexer.WHILE:  grammer[5],
			lexer.PRINT:  grammer[5],
			lexer.RETURN: grammer[5],
		},

		"FuncDecl": {
			lexer.FUNC: grammer[6],
		},

		"ParamList": {
			lexer.ID:     grammer[7],
			lexer.RPAREN: grammer[8],
		},

		"Param'": {
			lexer.COMMA:  grammer[9],
			lexer.RPAREN: grammer[10],
		},

		"Param": {
			lexer.ID: grammer[11],
		},

		"Type": {
			lexer.INT:   grammer[12],
			lexer.FLOAT: grammer[13],
			lexer.BOOL:  grammer[14],
		},

		"StmtList": {
			lexer.LET:      grammer[15],
			lexer.ID:       grammer[15],
			lexer.IF:       grammer[15],
			lexer.WHILE:    grammer[15],
			lexer.PRINT:    grammer[15],
			lexer.RETURN:   grammer[15],
			lexer.CONTINUE: grammer[15],
			lexer.BREAK:    grammer[15],
			lexer.RBRACE:   grammer[16],
		},

		"Stmt": {
			lexer.LET:      grammer[17],
			lexer.ID:       grammer[18],
			lexer.IF:       grammer[19],
			lexer.WHILE:    grammer[20],
			lexer.PRINT:    grammer[21],
			lexer.RETURN:   grammer[22],
			lexer.CONTINUE: grammer[23],
			lexer.BREAK:    grammer[24],
		},

		"VarDecl": {
			lexer.LET: grammer[25],
		},

		"Assign": {
			lexer.ID: grammer[26],
		},

		"AssignSuffix": {
			lexer.ASSIGN: grammer[27],
			lexer.LPAREN: grammer[28],
		},

		"IfStmt": {
			lexer.IF: grammer[29],
		},

		"ElsePart": {
			lexer.ELSE:     grammer[30],
			lexer.LET:      grammer[31],
			lexer.ID:       grammer[31],
			lexer.IF:       grammer[31],
			lexer.WHILE:    grammer[31],
			lexer.PRINT:    grammer[31],
			lexer.RETURN:   grammer[31],
			lexer.CONTINUE: grammer[31],
			lexer.BREAK:    grammer[31],
			lexer.RBRACE:   grammer[31],
			lexer.EOF:      grammer[31],
		},

		"WhileStmt": {
			lexer.WHILE: grammer[32],
		},

		// While bodies use the generic StmtList (see WhileStmt RHS).
		// No separate WhileStmtList/WhileStmtItem productions are defined.

		"ContinueStmt": {
			lexer.CONTINUE: grammer[33],
		},

		"BreakStmt": {
			lexer.BREAK: grammer[34],
		},

		"PrintStmt": {
			lexer.PRINT: grammer[35],
		},

		"ReturnStmt": {
			lexer.RETURN: grammer[36],
		},

		"ReturnValue": {
			lexer.ID:        grammer[37],
			lexer.NUM:       grammer[37],
			lexer.LPAREN:    grammer[37],
			lexer.TRUE:      grammer[37],
			lexer.FALSE:     grammer[37],
			lexer.NOT:       grammer[37],
			lexer.SEMICOLON: grammer[38],
		},

		"Expr": {
			lexer.ID:     grammer[39],
			lexer.NUM:    grammer[39],
			lexer.LPAREN: grammer[39],
			lexer.TRUE:   grammer[39],
			lexer.FALSE:  grammer[39],
		},

		"Expr'": {
			lexer.PLUS:      grammer[40],
			lexer.MINUS:     grammer[41],
			lexer.RPAREN:    grammer[42],
			lexer.SEMICOLON: grammer[42],
			lexer.LT:        grammer[42],
			lexer.GT:        grammer[42],
			lexer.LE:        grammer[42],
			lexer.GE:        grammer[42],
			lexer.EQ:        grammer[42],
			lexer.NE:        grammer[42],
			lexer.AND:       grammer[42],
			lexer.OR:        grammer[42],
			lexer.COMMA:     grammer[42],
		},

		"Term": {
			lexer.ID:     grammer[43],
			lexer.NUM:    grammer[43],
			lexer.LPAREN: grammer[43],
			lexer.TRUE:   grammer[43],
			lexer.FALSE:  grammer[43],
		},

		"Term'": {
			lexer.MULT:      grammer[44],
			lexer.DIV:       grammer[45],
			lexer.MOD:       grammer[46],
			lexer.PLUS:      grammer[47],
			lexer.MINUS:     grammer[47],
			lexer.RPAREN:    grammer[47],
			lexer.SEMICOLON: grammer[47],
			lexer.LT:        grammer[47],
			lexer.GT:        grammer[47],
			lexer.LE:        grammer[47],
			lexer.GE:        grammer[47],
			lexer.EQ:        grammer[47],
			lexer.NE:        grammer[47],
			lexer.AND:       grammer[47],
			lexer.OR:        grammer[47],
			lexer.COMMA:     grammer[47],
		},

		"FactorSuffix": {
			lexer.LPAREN:    grammer[54],
			lexer.MULT:      grammer[53],
			lexer.DIV:       grammer[53],
			lexer.MOD:       grammer[53],
			lexer.PLUS:      grammer[53],
			lexer.MINUS:     grammer[53],
			lexer.RPAREN:    grammer[53],
			lexer.SEMICOLON: grammer[53],
			lexer.LT:        grammer[53],
			lexer.GT:        grammer[53],
			lexer.LE:        grammer[53],
			lexer.GE:        grammer[53],
			lexer.EQ:        grammer[53],
			lexer.NE:        grammer[53],
			lexer.AND:       grammer[53],
			lexer.OR:        grammer[53],
			lexer.COMMA:     grammer[53],
		},

		"Factor": {
			lexer.ID:     grammer[48],
			lexer.NUM:    grammer[49],
			lexer.TRUE:   grammer[50],
			lexer.FALSE:  grammer[51],
			lexer.LPAREN: grammer[52],
		},

		"ArgList": {
			lexer.ID:     grammer[55],
			lexer.NUM:    grammer[55],
			lexer.LPAREN: grammer[55],
			lexer.TRUE:   grammer[55],
			lexer.FALSE:  grammer[55],
			lexer.NOT:    grammer[55],
			lexer.RPAREN: grammer[56],
		},

		"ArgList'": {
			lexer.COMMA:  grammer[57],
			lexer.RPAREN: grammer[58],
		},

		"Cond": {
			lexer.ID:     grammer[59],
			lexer.NUM:    grammer[59],
			lexer.LPAREN: grammer[59],
			lexer.NOT:    grammer[59],
			lexer.TRUE:   grammer[59],
			lexer.FALSE:  grammer[59],
		},

		"OrExpr": {
			lexer.ID:     grammer[60],
			lexer.NUM:    grammer[60],
			lexer.LPAREN: grammer[60],
			lexer.NOT:    grammer[60],
			lexer.TRUE:   grammer[60],
			lexer.FALSE:  grammer[60],
		},

		"OrExpr'": {
			lexer.OR:     grammer[61],
			lexer.RPAREN: grammer[62],
		},

		"AndExpr": {
			lexer.ID:     grammer[63],
			lexer.NUM:    grammer[63],
			lexer.LPAREN: grammer[63],
			lexer.NOT:    grammer[63],
			lexer.TRUE:   grammer[63],
			lexer.FALSE:  grammer[63],
		},

		"AndExpr'": {
			lexer.AND:    grammer[64],
			lexer.OR:     grammer[65],
			lexer.RPAREN: grammer[65],
		},

		"NotExpr": {
			lexer.NOT:    grammer[66],
			lexer.ID:     grammer[67],
			lexer.NUM:    grammer[67],
			lexer.LPAREN: grammer[67],
			lexer.TRUE:   grammer[67],
			lexer.FALSE:  grammer[67],
		},

		"RelExpr": {
			lexer.ID:     grammer[68],
			lexer.NUM:    grammer[68],
			lexer.LPAREN: grammer[68],
			lexer.TRUE:   grammer[68],
			lexer.FALSE:  grammer[68],
		},

		"RelExpr'": {
			lexer.LT:     grammer[69],
			lexer.GT:     grammer[69],
			lexer.LE:     grammer[69],
			lexer.GE:     grammer[69],
			lexer.EQ:     grammer[69],
			lexer.NE:     grammer[69],
			lexer.AND:    grammer[70],
			lexer.OR:     grammer[70],
			lexer.RPAREN: grammer[70],
		},

		"BoolPrimary": {
			lexer.TRUE:   grammer[71],
			lexer.FALSE:  grammer[72],
			lexer.LPAREN: grammer[73],
			lexer.ID:     grammer[73],
			lexer.NUM:    grammer[73],
		},

		"RelOp": {
			lexer.LT: grammer[74],
			lexer.GT: grammer[75],
			lexer.LE: grammer[76],
			lexer.GE: grammer[77],
			lexer.EQ: grammer[78],
			lexer.NE: grammer[79],
		},
	}
}
