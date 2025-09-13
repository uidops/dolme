package parser

import (
	"dolme/pkg/color"
	"dolme/pkg/lexer"
	"fmt"
)

// handleTerminalError is called when a terminal on the stack doesn't match current token.
// It only reports an error. It does NOT advance tokens or modify the stack.
func (p *Parser) handleTerminalError(expected string) bool {
	// Heuristic: if we expected ';' but current token clearly starts a new statement,
	// closes a block, or ends input, report "Missing semicolon".
	if expected == ";" && p.isStatementBoundary(p.currentToken.Type) {
		p.addError("Missing semicolon")
		return false
	}

	// Specific: assignment without identifier like `let = 42;`
	if expected == "id" && p.currentToken.Type == lexer.ASSIGN {
		p.addError("Missing identifier")
		return false
	}

	// Default contextual error
	p.addContextualError(expected)
	return false
}

// handleNonTerminalError is called when there is no production for top non-terminal and current token.
// It only reports an error. It does NOT advance tokens or modify the stack.
func (p *Parser) handleNonTerminalError(expected string) bool {
	// Special: ArgList followed by a boundary like ';' (e.g., id '(' ; )
	// This likely means a missing ')'. Emit that once.
	if expected == "ArgList" {
		if p.currentToken.Type == lexer.SEMICOLON ||
			p.currentToken.Type == lexer.RBRACE ||
			p.isStatementBoundary(p.currentToken.Type) {
			p.addError("Missing closing parenthesis")
			return false
		}
	}

	// Empty condition: if () or while()
	if expected == "Cond" && p.currentToken.Type == lexer.RPAREN {
		p.addError("Empty condition")
		return false
	}

	// If parsing an expression tail/postfix and the next token begins a new statement or closes a block,
	// this is often a missing semicolon.
	if p.isExpressionTail(expected) && p.isStatementBoundary(p.currentToken.Type) {
		p.addError("Missing semicolon")
		return false
	}

	// Default contextual error
	p.addContextualError(expected)
	return false
}

// handleUnexpectedEndOfInput is called when input ends but stack is not empty
func (p *Parser) handleUnexpectedEndOfInput() {
	p.addError(fmt.Sprintf("Unexpected token '%s' at end of input", p.currentToken.Type))
}

// addError records a parsing error with location
func (p *Parser) addError(msg string) {
	pos := p.currentToken.Pos
	formatted := color.RedText(msg) + " at " + color.YellowText(fmt.Sprintf("Line: %d, Column %d", pos.Line, pos.Column))
	p.errors = append(p.errors, formatted)
}

// Errors returns the list of parsing errors
func (p *Parser) Errors() []string {
	return p.errors
}

// isExpressionTail checks if the non-terminal is part of an expression
func (p *Parser) isExpressionTail(sym string) bool {
	switch sym {
	case "Expr", "Expr'", "Term", "Term'", "Factor", "FactorSuffix":
		return true
	default:
		return false
	}
}

// isStatementBoundary checks if a token type indicates the start of a new statement or block boundary
func (p *Parser) isStatementBoundary(t lexer.TokenType) bool {
	switch t {
	case lexer.LET, lexer.ID, lexer.IF, lexer.WHILE, lexer.PRINT, lexer.RETURN, lexer.CONTINUE, lexer.BREAK, lexer.ELSE, lexer.RBRACE, lexer.EOF:
		return true
	default:
		return false
	}
}

// addContextualError generates a contextual error message based on expected and current token
func (p *Parser) addContextualError(expected string) {
	current := p.currentToken
	p.addError(p.categorizeError(expected, current))
}

// categorizeError provides a specific error message based on expected symbol and current token
func (p *Parser) categorizeError(expected string, current lexer.Token) string {
	// Delimiters
	switch expected {
	case ")":
		return "Missing closing parenthesis"
	case "}":
		return "Missing closing brace"
	case "{":
		return "Missing opening brace"
	case ";":
		return "Missing semicolon"
	case "=":
		return "Missing assignment operator"
	case "(":
		if current.Type == lexer.LBRACE {
			return "Wrong bracket type - expected parenthesis"
		}
		return "Missing opening parenthesis"
	}

	// Identifiers and literals
	switch expected {
	case "id":
		if current.Type == lexer.ASSIGN || current.Type == lexer.SEMICOLON {
			return "Missing identifier"
		}
		if p.isReservedKeywordAsId(current) {
			return "Cannot use reserved keyword as identifier"
		}
		return "Expected identifier"
	case "num":
		return "Expected number"
	case "string":
		if current.Type == lexer.ID {
			return "Missing quotes around string"
		}
		return "Expected string"
	}

	// Expressions missing before ) or ; or relation
	if (expected == "Expr" || expected == "Term" || expected == "Factor") &&
		(current.Type == lexer.SEMICOLON || current.Type == lexer.RPAREN) {
		return "Missing expression"
	}

	// Non-terminal Cond with ')'
	if expected == "Cond" && current.Type == lexer.RPAREN {
		return "Empty condition"
	}

	return "Syntax error"
}

// isReservedKeywordAsId checks if an identifier token is actually a reserved keyword
func (p *Parser) isReservedKeywordAsId(current lexer.Token) bool {
	if current.Type != lexer.ID {
		return false
	}
	reserved := map[lexer.TokenType]bool{
		lexer.LET: true, lexer.FUNC: true, lexer.RETURN: true, lexer.IF: true, lexer.ELSE: true,
		lexer.WHILE: true, lexer.BREAK: true, lexer.CONTINUE: true, lexer.PRINT: true,
		lexer.TRUE: true, lexer.FALSE: true, lexer.AND: true, lexer.OR: true, lexer.NOT: true,
		lexer.INT: true, lexer.FLOAT: true, lexer.BOOL: true,
	}
	return reserved[current.Type]
}
