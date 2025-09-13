package lexer

type Lexer struct {
	input        string // input string to be tokenized
	length       int    // length of the input string
	position     int    // current position in the input string
	line         int    // current line number for error reporting
	column       int    // current column number for error reporting
	currentToken Token  // current token for context (e.g., unary minus handling)
}

// Create a new lexer instance
func NewLexer(s string) *Lexer {
	return &Lexer{
		input:        s,
		length:       len(s),
		position:     0,
		line:         1,
		column:       1,
		currentToken: Token{},
	}
}

// Get the next token from the input
func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	// End of input
	if l.position >= l.length {
		tok := NewToken(EOF, "", "", l.currentPosition())
		l.currentToken = tok
		return tok
	}

	// Handle signed numbers in contexts that allow unary minus.
	// We only treat '-' + number as a single NUM when:
	// - previous token allows unary (start of input or after operators/delimiters)
	// - '-' is immediately followed by a digit and the remainder matches NUM
	if l.input[l.position] == '-' && l.prevAllowsUnary() {
		if l.position+1 < l.length && isDigit(l.input[l.position+1]) {
			remaining := l.input[l.position+1:]
			t, lex, matched := MatchToken(remaining)
			if matched && t == NUM && lex != "" {
				lexeme := "-" + lex

				tok := NewToken(NUM, lexeme, lexeme, l.currentPosition())

				l.advance(len(lexeme))
				l.currentToken = tok

				return tok
			}
		}
	}

	// Regex match the first token it sees from the remaining input from current position to the end
	remaining := l.input[l.position:]
	token_type, lexeme, matched := MatchToken(remaining)

	if !matched || token_type == EOF {
		if token_type == EOF && lexeme != "" {
			l.advance(len(lexeme))
			return l.NextToken()
		}

		char := string(l.input[l.position])
		l.advance(1)

		tok := NewToken(ILLEGAL, char, "", l.currentPosition())
		l.currentToken = tok
		return tok
	}

	var literal string
	switch token_type {
	case NUM:
		literal = lexeme
	case TRUE:
		literal = "true"
	case FALSE:
		literal = "false"
	case STRING:
		// Remove the surrounding quotes from the lexeme
		literal = lexeme[1 : len(lexeme)-1]
	default:
		literal = lexeme
	}

	tok := NewToken(token_type, lexeme, literal, l.currentPosition())
	l.advance(len(lexeme))
	l.currentToken = tok

	return tok
}

// View next token without advancing the position
func (l *Lexer) Peek() Token {
	// save state
	cpos := l.position
	cline := l.line
	ccol := l.column
	ctok := l.currentToken

	token := l.NextToken()

	// restore state
	l.position = cpos
	l.line = cline
	l.column = ccol
	l.currentToken = ctok

	return token
}

// Check if there are more characters to read
func (l *Lexer) HasMore() bool {
	return l.position < l.length
}

// Skip whitespace and comments
func (l *Lexer) skipWhitespace() {
	for l.position < l.length {
		ch := l.input[l.position]

		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			// handle whitespace and new lines
			if ch == '\n' {
				l.line++
				l.column = 1
			} else {
				l.column++
			}
			l.position++

		} else if l.position+1 < l.length && ch == '/' && l.input[l.position+1] == '/' {
			// handle comments
			for l.position < l.length {
				ch := l.input[l.position]
				l.position++
				if ch == '\n' {
					l.line++
					l.column = 1
					break
				} else {
					l.column++
				}
			}
		} else {
			break
		}
	}
}

// Advance the lexer position by n characters
func (l *Lexer) advance(n int) {
	for range n {
		if l.position >= l.length {
			break
		}

		if l.input[l.position] == '\n' {
			l.line++
			l.column = 1
		} else {
			l.column++
		}

		l.position++
	}
}

// Get the current position of the lexer
func (l *Lexer) currentPosition() Position {
	return Position{
		Line:   l.line,
		Column: l.column,
		Offset: l.position,
	}
}

// Check if the previous token allows a unary operator (like - or +)
func (l *Lexer) prevAllowsUnary() bool {
	switch l.currentToken.Type {
	case EOF,
		ASSIGN,    // =
		LPAREN,    // (
		COMMA,     // ,
		COLON,     // :
		SEMICOLON, // ;
		LBRACE,    // {
		// arithmetic operators
		PLUS, MINUS, MULT, DIV, MOD,
		// relational operators
		LT, GT, LE, GE, EQ, NE,
		// logical operators / keywords that can precede expr
		AND, OR, NOT, RETURN, IF, WHILE, PRINT:
		return true
	default:
		return false
	}
}
