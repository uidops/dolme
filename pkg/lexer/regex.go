package lexer

import (
	"regexp"
)

type tokenRegex struct {
	Pattern *regexp.Regexp
	Raw     string
}

// Token regex patterns
var tokenRegexes = map[TokenType]tokenRegex{
	LE: {regexp.MustCompile(`^<=`), `^<=`},
	GE: {regexp.MustCompile(`^>=`), `^>=`},
	EQ: {regexp.MustCompile(`^==`), `^==`},
	NE: {regexp.MustCompile(`^!=`), `^!=`},

	LET:      {regexp.MustCompile(`^let\b`), `^let\b`},
	FUNC:     {regexp.MustCompile(`^func\b`), `^func\b`},
	RETURN:   {regexp.MustCompile(`^return\b`), `^return\b`},
	IF:       {regexp.MustCompile(`^if\b`), `^if\b`},
	ELSE:     {regexp.MustCompile(`^else\b`), `^else\b`},
	WHILE:    {regexp.MustCompile(`^while\b`), `^while\b`},
	BREAK:    {regexp.MustCompile(`^break\b`), `^break\b`},
	CONTINUE: {regexp.MustCompile(`^continue\b`), `^continue\b`},
	PRINT:    {regexp.MustCompile(`^print\b`), `^print\b`},
	AND:      {regexp.MustCompile(`^and\b`), `^and\b`},
	OR:       {regexp.MustCompile(`^or\b`), `^or\b`},
	NOT:      {regexp.MustCompile(`^not\b`), `^not\b`},
	TRUE:     {regexp.MustCompile(`^true\b`), `^true\b`},
	FALSE:    {regexp.MustCompile(`^false\b`), `^false\b`},
	INT:      {regexp.MustCompile(`^int\b`), `^int\b`},
	FLOAT:    {regexp.MustCompile(`^float\b`), `^float\b`},
	BOOL:     {regexp.MustCompile(`^bool\b`), `^bool\b`},

	ASSIGN: {regexp.MustCompile(`^=`), `^=`},
	PLUS:   {regexp.MustCompile(`^\+`), `^\+`},
	MINUS:  {regexp.MustCompile(`^-`), `^-`},
	MULT:   {regexp.MustCompile(`^\*`), `^\*`},
	DIV:    {regexp.MustCompile(`^/`), `^/`},
	MOD:    {regexp.MustCompile(`^%`), `^%`},
	LT:     {regexp.MustCompile(`^<`), `^<`},
	GT:     {regexp.MustCompile(`^>`), `^>`},

	SEMICOLON: {regexp.MustCompile(`^;`), `^;`},
	COMMA:     {regexp.MustCompile(`^,`), `^,`},
	COLON:     {regexp.MustCompile(`^:`), `^:`},
	LPAREN:    {regexp.MustCompile(`^\(`), `^\(`},
	RPAREN:    {regexp.MustCompile(`^\)`), `^\)`},
	LBRACE:    {regexp.MustCompile(`^\{`), `^\{`},
	RBRACE:    {regexp.MustCompile(`^\}`), `^\}`},
	LSBRACE:   {regexp.MustCompile(`^\[`), `^\[`},
	RSBRACE:   {regexp.MustCompile(`^\]`), `^\]`},

	NUM:    {regexp.MustCompile(`^\d+(\.\d+)?([eE][+-]?\d+)?`), `^\d+(\.\d+)?([eE][+-]?\d+)?`},
	STRING: {regexp.MustCompile(`^"([^"\\]|\\.)*"`), `^"([^"\\]|\\.)*"`},
	ID:     {regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*`), `^[a-zA-Z][a-zA-Z0-9_]*`},
}

var (
	whitespaceRegex = regexp.MustCompile(`^\s+`)
	commentRegex    = regexp.MustCompile(`^//.*$`)
)

// Token precedence order for matching (longer patterns first)
var tokenPrecedenceOrder = []TokenType{
	CONTINUE, RETURN, BREAK, FALSE, FLOAT, PRINT, WHILE, ELSE, FUNC,
	TRUE, BOOL, AND, INT, LET, NOT, IF, OR, LE, GE, EQ, NE, ASSIGN, PLUS,
	MINUS, MULT, DIV, MOD, LT, GT, SEMICOLON, COMMA, COLON,
	LPAREN, RPAREN, LBRACE, RBRACE, LSBRACE,
	RSBRACE, NUM, STRING, ID,
}

// Get the regex pattern for a token type
func (t TokenType) Regex() *regexp.Regexp {
	if regex, ok := tokenRegexes[t]; ok {
		return regex.Pattern
	}

	return nil
}

// Get the raw regex string for a token type
func (t TokenType) RawRegex() string {
	if regex, ok := tokenRegexes[t]; ok {
		return regex.Raw
	}

	return ""
}

// Match the longest token at the start of the string
func MatchToken(s string) (TokenType, string, bool) {
	if s == "" {
		return EOF, "", false
	} else if match := whitespaceRegex.FindString(s); match != "" {
		return EOF, match, true
	} else if match := commentRegex.FindString(s); match != "" {
		return EOF, match, true
	}

	for _, tokenType := range tokenPrecedenceOrder {
		if regex, ok := tokenRegexes[tokenType]; ok {
			if match := regex.Pattern.FindString(s); match != "" {
				return tokenType, match, true
			}
		}
	}

	return ILLEGAL, string(s[0]), false
}

// Check if a byte is a digit
func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}
