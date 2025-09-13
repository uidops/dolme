package lexer_test

import (
	"dolme/pkg/lexer"
	"testing"
)

func TestNumbers(t *testing.T) {
	tests := []struct {
		input       string
		expected    lexer.TokenType
		description string
	}{
		{"42", lexer.NUM, "integer"},
		{"0", lexer.NUM, "zero"},

		{"3.14", lexer.NUM, "simple float"},
		{"0.5", lexer.NUM, "float starting with zero"},
		{"123.456", lexer.NUM, "multi-digit float"},

		{"1e5", lexer.NUM, "scientific notation with e"},
		{"1e+5", lexer.NUM, "scientific notation with e+"},
		{"1e-5", lexer.NUM, "scientific notation with e-"},
		{"2.5e10", lexer.NUM, "float with scientific notation"},
		{"3.14e-2", lexer.NUM, "float with negative exponent"},
		{"1.23e+10", lexer.NUM, "float with positive exponent"},

		{"1E5", lexer.NUM, "scientific notation with E"},
		{"1E+5", lexer.NUM, "scientific notation with E+"},
		{"1E-5", lexer.NUM, "scientific notation with E-"},
		{"2.5E10", lexer.NUM, "float with scientific notation E"},
		{"3.14E-2", lexer.NUM, "float with negative exponent E"},
		{"1.23E+10", lexer.NUM, "float with positive exponent E"},

		{"0.0", lexer.NUM, "zero as float"},
		{"0e0", lexer.NUM, "zero in scientific notation"},
		{"1000000", lexer.NUM, "large integer"},
		{"1e6", lexer.NUM, "large number in scientific"},
	}

	for _, test := range tests {
		tokenType, lexeme, matched := lexer.MatchToken(test.input)
		if !matched {
			t.Errorf("Failed to match %s (%s)", test.input, test.description)
		}
		if tokenType != test.expected {
			t.Errorf("Input %s (%s): expected %s, got %s", test.input, test.description, test.expected, tokenType)
		}
		if lexeme != test.input {
			t.Errorf("Input %s (%s): expected lexeme %s, got %s", test.input, test.description, test.input, lexeme)
		}
	}
}
