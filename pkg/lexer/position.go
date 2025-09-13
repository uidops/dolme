package lexer

import "fmt"

type Position struct {
	Line   int
	Column int
	Offset int
}

// Returns a string representation of the Position
func (p Position) String() string {
	return fmt.Sprintf("%d, %d, %d", p.Line, p.Column, p.Offset)
}

// Creates a new Position instance
func NewPosition(line, column, offset int) Position {
	return Position{
		Line:   line,
		Column: column,
		Offset: offset,
	}
}
