package lexer_test

import (
	"dolme/pkg/lexer"
	"testing"
)

func TestTokens(t *testing.T) {
	input := "let x : int = 10 / 2;\n" + "while (true - 20) {\n" + "	x = 5;\n" + "if (x == 5) {\n" + "		print(x);\n" + "	}\n" + "}\nx = 10"
	mylexer := lexer.NewLexer(input)

	expectedTokens := []lexer.TokenType{
		lexer.LET, lexer.ID, lexer.COLON, lexer.INT, lexer.ASSIGN, lexer.NUM, lexer.DIV, lexer.NUM, lexer.SEMICOLON,
		lexer.WHILE, lexer.LPAREN, lexer.TRUE, lexer.MINUS, lexer.NUM, lexer.RPAREN, lexer.LBRACE,
		lexer.ID, lexer.ASSIGN, lexer.NUM, lexer.SEMICOLON,
		lexer.IF, lexer.LPAREN, lexer.ID, lexer.EQ, lexer.NUM, lexer.RPAREN, lexer.LBRACE,
		lexer.PRINT, lexer.LPAREN, lexer.ID, lexer.RPAREN, lexer.SEMICOLON,
		lexer.RBRACE, lexer.RBRACE,
		lexer.ID, lexer.ASSIGN, lexer.NUM,
		lexer.EOF,
	}

	for i, expected := range expectedTokens {
		token := mylexer.NextToken()
		if token.Type != expected {
			t.Errorf("Token %d: expected %s, got %s", i, expected, token.Type)
		}
	}
}
