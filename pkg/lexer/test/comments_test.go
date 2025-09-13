package lexer_test

import (
	"dolme/pkg/lexer"
	"testing"
)

func TestComments(t *testing.T) {
	input := `// test comment
let x : int = 10; // another test comment
// another another test comment
let y : float = 20.0;`

	mylexer := lexer.NewLexer(input)
	expectedTokens := []lexer.TokenType{
		lexer.LET, lexer.ID, lexer.COLON, lexer.INT, lexer.ASSIGN, lexer.NUM, lexer.SEMICOLON,
		lexer.LET, lexer.ID, lexer.COLON, lexer.FLOAT, lexer.ASSIGN, lexer.NUM, lexer.SEMICOLON,
		lexer.EOF,
	}

	for i, expected := range expectedTokens {
		token := mylexer.NextToken()
		if token.Type != expected {
			t.Errorf("Token %d: expected %s, got %s", i, expected, token.Type)
		}
	}
}
