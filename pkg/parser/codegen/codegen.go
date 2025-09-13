package codegen

import (
	"dolme/pkg/lexer"
	"dolme/pkg/parser/stack"
	"strconv"
)

type Codegen struct {
	ss              *stack.Stack               // Semantic stack
	i               int                        // Instruction counter
	pb              []Instruction              // Program Block (list of IR instructions)
	tempCounter     int                        // Temporary variable counter
	globalCounter   int                        // Global variable counter
	localCounter    int                        // Local variable counter
	paramCounter    int                        // Function parameter counter
	argsCounter     int                        // Function arguments counter
	currentToken    lexer.Token                // Current token being processed
	inFunction      bool                       // Flag indicating if inside a function
	symbolTable     map[string]int             // Symbol table mapping variable names to addresses
	functionScope   map[string]int             // Function scope symbol table
	typeTable       map[int]lexer.TokenType    // Type table mapping addresses to types
	functionReturns map[string]lexer.TokenType // Function return types
	errors          []string                   // List of semantic errors
}

// NewCodegen creates a new Codegen instance
func NewCodegen() *Codegen {
	return &Codegen{
		ss:              stack.NewStack(),
		i:               0,
		pb:              make([]Instruction, 0),
		tempCounter:     0,
		globalCounter:   0,
		localCounter:    0,
		paramCounter:    0,
		argsCounter:     0,
		currentToken:    lexer.Token{},
		inFunction:      false,
		symbolTable:     make(map[string]int),
		functionScope:   make(map[string]int),
		typeTable:       make(map[int]lexer.TokenType),
		functionReturns: make(map[string]lexer.TokenType),
	}
}

// GetProgram returns the generated program block
func (c *Codegen) GetProgram() []Instruction {
	return c.pb
}

// SetCurrentToken sets the current token being processed
func (c *Codegen) SetCurrentToken(token lexer.Token) {
	c.currentToken = token
}

// pushString pushes a string onto the semantic stack
func (c *Codegen) pushString(val string) {
	c.ss.Push(val)
}

// push pushes an integer onto the semantic stack as a string
func (c *Codegen) push(val int) {
	c.ss.Push(strconv.Itoa(val))
}

// pop pops an integer from the semantic stack
func (c *Codegen) popString(count int) {
	c.pop(count)
}

// pop pops an integer from the semantic stack and converts it to int
func (c *Codegen) pop(count int) {
	for i := 0; i < count && c.ss.Size() > 0; i++ {
		c.ss.Pop()
	}
}

// popAtOffsetEnd pops an integer from the semantic stack at a specific offset from the end
func (c *Codegen) popAtOffsetEnd(offset int) {
	if c.ss.Size() > offset {
		arr := c.ss.Array()
		newArr := append(arr[:len(arr)-1-offset], arr[len(arr)-offset:]...)
		c.ss = stack.NewStack(newArr...)
	}
}

// top returns the top integer from the semantic stack without popping it
func (c *Codegen) top() int {
	if c.ss.Size() > 0 {
		val, _ := strconv.Atoi(c.ss.Peek())
		return val
	}

	return -1
}

// topString returns the top string from the semantic stack without popping it
func (c *Codegen) topString() string {
	if c.ss.Size() > 0 {
		return c.ss.Peek()
	}

	return ""
}

// topMinus returns the integer at the given offset from the top of the semantic stack
func (c *Codegen) topMinus(offset int) int {
	if c.ss.Size() > offset {
		arr := c.ss.Array()
		val, _ := strconv.Atoi(arr[len(arr)-1-offset])
		return val
	}

	return -1
}

// topStringMinus returns the string at the given offset from the top of the semantic stack
func (c *Codegen) topStringMinus(offset int) string {
	if c.ss.Size() > offset {
		arr := c.ss.Array()
		return arr[len(arr)-1-offset]
	}

	return ""
}

// getTemp returns a new temporary variable address
func (c *Codegen) getTemp() int {
	addr := 600 + c.tempCounter
	c.tempCounter++
	return addr
}

// getVariable returns a new variable address based on the current scope
func (c *Codegen) getVariable() int {
	if c.inFunction {
		return c.getLocalVariable()
	}

	addr := 400 + c.globalCounter
	c.globalCounter++

	return addr
}

// getLocalVariable returns a new local variable address
func (c *Codegen) getLocalVariable() int {
	addr := 800 + c.localCounter
	c.localCounter++

	return addr
}

// setInFunction sets the inFunction flag and resets local scope if exiting a function
func (c *Codegen) setInFunction(state bool) {
	c.inFunction = state
	if !state {
		// reset local counter and clear function scope when exiting function
		c.localCounter = 0
		c.functionScope = make(map[string]int)
	}
}

// declareVariable declares a variable in the appropriate scope
func (c *Codegen) declareVariable(name string, addr int) {
	if c.inFunction {
		c.functionScope[name] = addr
	} else {
		c.symbolTable[name] = addr
	}
}

// isVariableDeclared checks if a variable is declared in the appropriate scope
func (c *Codegen) isVariableDeclared(name string) bool {
	// check function scope first
	if c.inFunction {
		if _, exists := c.functionScope[name]; exists {
			return true
		}
	}

	// check global scope
	if _, exists := c.symbolTable[name]; exists {
		return true
	}

	return false
}

// getVariableAddress retrieves the address of a variable from the appropriate scope
func (c *Codegen) getVariableAddress(name string) (int, bool) {
	// check function scope first
	if c.inFunction {
		if addr, exists := c.functionScope[name]; exists {
			return addr, true
		}
	}

	// check global scope
	if addr, exists := c.symbolTable[name]; exists {
		return addr, true
	}

	return 0, false
}

// GetVariableType retrieves the type of a variable by its address
func (c *Codegen) GetVariableType(addr int) lexer.TokenType {
	if t, exists := c.typeTable[addr]; exists {
		return t
	}
	return lexer.EOF
}

// setVariableType sets the type of a variable by its address
func (c *Codegen) setVariableType(addr int, t lexer.TokenType) {
	c.typeTable[addr] = t
}
