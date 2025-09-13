package parser

import (
	"dolme/pkg/lexer"
	"strings"
)

// isTerminal checks if a symbol is a terminal
func (p *Parser) isTerminal(symbol string) bool {
	// Semantic actions are considered terminals
	if strings.HasPrefix(symbol, "@") {
		return true
	}

	terminals := map[string]bool{
		"let": true, "func": true, "return": true, "if": true, "while": true, "break": true, "continue": true, "print": true, "else": true,
		"true": true, "false": true, "not": true, "and": true, "or": true,
		"int": true, "float": true, "bool": true,
		"(": true, ")": true, "{": true, "}": true, ";": true, ",": true, "=": true, ":": true,
		"+": true, "-": true, "*": true, "/": true, "%": true,
		"<": true, ">": true, "<=": true, ">=": true, "==": true, "!=": true,
		"id": true, "num": true, "string": true, "$": true, "EOF": true,
	}

	return terminals[symbol]
}

// matchTerminal checks if the current token matches the expected terminal
func (p *Parser) matchTerminal(expected string) bool {
	switch expected {
	case "id":
		return p.currentToken.Type == lexer.ID
	case "num":
		return p.currentToken.Type == lexer.NUM
	case "string":
		return p.currentToken.Type == lexer.STRING
	case "$":
		return p.currentToken.Type == lexer.EOF
	case "let":
		return p.currentToken.Type == lexer.LET
	case "func":
		return p.currentToken.Type == lexer.FUNC
	case "return":
		return p.currentToken.Type == lexer.RETURN
	case "if":
		return p.currentToken.Type == lexer.IF
	case "while":
		return p.currentToken.Type == lexer.WHILE
	case "break":
		return p.currentToken.Type == lexer.BREAK
	case "continue":
		return p.currentToken.Type == lexer.CONTINUE
	case "print":
		return p.currentToken.Type == lexer.PRINT
	case "else":
		return p.currentToken.Type == lexer.ELSE
	case "true":
		return p.currentToken.Type == lexer.TRUE
	case "false":
		return p.currentToken.Type == lexer.FALSE
	case "not":
		return p.currentToken.Type == lexer.NOT
	case "and":
		return p.currentToken.Type == lexer.AND
	case "or":
		return p.currentToken.Type == lexer.OR
	case "(":
		return p.currentToken.Type == lexer.LPAREN
	case ")":
		return p.currentToken.Type == lexer.RPAREN
	case "{":
		return p.currentToken.Type == lexer.LBRACE
	case "}":
		return p.currentToken.Type == lexer.RBRACE
	case ";":
		return p.currentToken.Type == lexer.SEMICOLON
	case ",":
		return p.currentToken.Type == lexer.COMMA
	case ":":
		return p.currentToken.Type == lexer.COLON
	case "=":
		return p.currentToken.Type == lexer.ASSIGN
	case "int":
		return p.currentToken.Type == lexer.INT
	case "float":
		return p.currentToken.Type == lexer.FLOAT
	case "bool":
		return p.currentToken.Type == lexer.BOOL
	case "+":
		return p.currentToken.Type == lexer.PLUS
	case "-":
		return p.currentToken.Type == lexer.MINUS
	case "*":
		return p.currentToken.Type == lexer.MULT
	case "/":
		return p.currentToken.Type == lexer.DIV
	case "%":
		return p.currentToken.Type == lexer.MOD
	case "<":
		return p.currentToken.Type == lexer.LT
	case ">":
		return p.currentToken.Type == lexer.GT
	case "<=":
		return p.currentToken.Type == lexer.LE
	case ">=":
		return p.currentToken.Type == lexer.GE
	case "==":
		return p.currentToken.Type == lexer.EQ
	case "!=":
		return p.currentToken.Type == lexer.NE
	default:
		return false
	}
}
