package parser

import (
	"dolme/pkg/lexer"
	"dolme/pkg/parser/codegen"
	"dolme/pkg/parser/stack"
	"strings"
)

type Parser struct {
	stack        *stack.Stack     // LL(1) parsing stack
	lexer        *lexer.Lexer     // lexer instance
	cg           *codegen.Codegen // code generator instance
	currentToken lexer.Token      // current token
	table        ParsingTable     // LL(1) parsing table
	errors       []string         // list of errors
}

// NewParser creates a new parser instance
func NewParser(l *lexer.Lexer) *Parser {
	p := &Parser{
		lexer:  l,
		cg:     codegen.NewCodegen(),
		table:  NewParsingTable(),
		stack:  stack.NewStack("$", "Program"), // Program is start state and $ is bottom of the stack
		errors: []string{},
	}

	// Initialize current token
	p.nextToken()

	return p
}

// Parse starts parsing the input program
func (p *Parser) Parse() {
	for p.stack.Size() > 1 { // While stack is not empty (only $ remains)
		top := p.stack.Pop()

		if p.isTerminal(top) {
			// Check if this is a semantic action
			if p.isSemanticAction(top) {
				p.cg.ExecuteAction(top)
			} else if p.matchTerminal(top) {
				p.cg.SetCurrentToken(p.currentToken)
				p.nextToken()
			} else {
				if p.handleTerminalError(top) {
					break
				}
			}
		} else {
			// Non-terminal: pick production from table
			if production, ok := p.table[top][p.currentToken.Type]; ok {
				rhs_length := len(production.RHS)
				// If production is ε, do not push anything
				if rhs_length == 0 || (rhs_length == 1 && production.RHS[0] == "ε") {
					continue
				}

				// Push RHS of production onto stack in reverse order (so first symbol is on top)
				for i := rhs_length - 1; i >= 0; i-- {
					if production.RHS[i] != "ε" {
						p.stack.Push(production.RHS[i])
					}
				}

			} else {
				if p.handleNonTerminalError(top) {
					break
				}
			}

		}
	}

	if p.currentToken.Type != lexer.EOF {
		p.handleUnexpectedEndOfInput()
		return
	}
}

// nextToken advances to the next token from the lexer
func (p *Parser) nextToken() {
	p.currentToken = p.lexer.NextToken()
}

// isSemanticAction checks if a symbol is a semantic action
func (p *Parser) isSemanticAction(symbol string) bool {
	return strings.HasPrefix(symbol, "@")
}

// GetIRCode returns the generated three-address code
func (p *Parser) GetIRCode() []codegen.Instruction {
	return append(p.cg.GetProgram(), codegen.Instruction{Op: codegen.OpNop, Arg1: nil, Arg2: nil, Arg3: nil, Type: lexer.EOF}) // Append NOP at the end
}

// GetSemanticErrors returns the list of semantic errors
func (p *Parser) GetSemanticErrors() []string {
	return p.cg.GetErrors()
}

// GetCG returns the code generator instance
func (p *Parser) GetCG() *codegen.Codegen {
	return p.cg
}
