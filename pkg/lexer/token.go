package lexer

import (
	"fmt"
)

type TokenType int
type TokenCategory int

type Token struct {
	Type    TokenType // Type of the token
	Lexeme  string    // Actual string from source code
	Literal string    // Literal value (if applicable), empty string if not
	Pos     Position  // Position in source code
}

// NewToken creates a new Token instance
func NewToken(tokenType TokenType, lexeme string, literal string, Pos Position) Token {
	return Token{
		Type:    tokenType,
		Lexeme:  lexeme,
		Literal: literal,
		Pos:     Pos,
	}
}

const (
	NONE TokenCategory = iota
	KEYWORD
	IDENTIFIER
	LITERAL
	OPERATOR
	DELIMITER
)

const (
	EOF TokenType = iota // End of file

	LET      // let
	FUNC     // func
	RETURN   // return
	IF       // if
	ELSE     // else
	WHILE    // while
	BREAK    // break
	CONTINUE // continue
	PRINT    // print
	AND      // and
	OR       // or
	NOT      // not
	TRUE     // true
	FALSE    // false
	INT      // int
	FLOAT    // float
	BOOL     // bool

	ID     // id (identifier)
	NUM    // num (number)
	STRING // string literal

	ASSIGN // =
	PLUS   // +
	MINUS  // -
	MULT   // *
	DIV    // /
	MOD    // %
	LT     // <
	GT     // >
	LE     // <=
	GE     // >=
	EQ     // ==
	NE     // !=

	SEMICOLON // ;
	COMMA     // ,
	COLON     // :
	LPAREN    // (
	RPAREN    // )
	LBRACE    // {
	RBRACE    // }
	LSBRACE   // [
	RSBRACE   // ]

	ILLEGAL // illegal token
)

var Keywords = map[string]TokenType{
	"let":      LET,
	"func":     FUNC,
	"return":   RETURN,
	"if":       IF,
	"else":     ELSE,
	"while":    WHILE,
	"break":    BREAK,
	"continue": CONTINUE,
	"print":    PRINT,
	"and":      AND,
	"or":       OR,
	"not":      NOT,
	"true":     TRUE,
	"false":    FALSE,
	"int":      INT,
	"float":    FLOAT,
	"bool":     BOOL,
}

// TokenToString converts a TokenType to its string representation
func (t Token) TokenToString() (string, bool) {
	mapping := map[TokenType]string{
		LET:       "let",
		FUNC:      "func",
		RETURN:    "return",
		IF:        "if",
		WHILE:     "while",
		BREAK:     "break",
		CONTINUE:  "continue",
		PRINT:     "print",
		ELSE:      "else",
		TRUE:      "true",
		FALSE:     "false",
		INT:       "int",
		FLOAT:     "float",
		BOOL:      "bool",
		NOT:       "not",
		AND:       "and",
		OR:        "or",
		LPAREN:    "(",
		RPAREN:    ")",
		LBRACE:    "{",
		RBRACE:    "}",
		SEMICOLON: ";",
		COMMA:     ",",
		COLON:     ":",
		ASSIGN:    "=",
		PLUS:      "+",
		MINUS:     "-",
		MULT:      "*",
		DIV:       "/",
		MOD:       "%",
		LT:        "<",
		GT:        ">",
		LE:        "<=",
		GE:        ">=",
		EQ:        "==",
		NE:        "!=",
		ID:        "id",
		NUM:       "num",
		STRING:    "string",
		EOF:       "$",
	}

	str, ok := mapping[t.Type]
	return str, ok
}

// String returns a string representation of the Token
func (t Token) String() string {
	if t.Literal == "" {

		return fmt.Sprintf("T_{%s, %v, nil, %s}",
			t.Type, t.Lexeme, t.Pos.String())
	}

	return fmt.Sprintf("T_{%s, %v, %q, %s}",
		t.Type, t.Lexeme, t.Literal, t.Pos.String())
}

// String returns a string representation of the TokenType
func (t TokenType) String() string {
	if str, ok := (Token{Type: t}).TokenToString(); ok {
		return str
	}

	return fmt.Sprintf("UNKNOWN(%d)", int(t))
}

// GetCategory returns the category of the token
func (t TokenType) GetCategory() TokenCategory {
	switch t {
	case LET, FUNC, RETURN, IF, ELSE, WHILE, BREAK, CONTINUE, PRINT, AND, OR, NOT, TRUE, FALSE, INT, FLOAT, BOOL:
		return KEYWORD
	case ID:
		return IDENTIFIER
	case NUM, STRING:
		return LITERAL
	case ASSIGN, PLUS, MINUS, MULT, DIV, MOD, LT, GT, LE, GE, EQ, NE:
		return OPERATOR
	case SEMICOLON, COMMA, COLON, LPAREN, RPAREN, LBRACE, RBRACE:
		return DELIMITER
	default:
		return NONE
	}
}

// IsKeyword checks if the given identifier is a keyword and returns its TokenType if it is
func IsKeyword(identifier string) (TokenType, bool) {
	tokenType, ok := Keywords[identifier]
	return tokenType, ok
}
