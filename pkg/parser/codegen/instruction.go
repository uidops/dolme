package codegen

import (
	"dolme/pkg/lexer"
	"fmt"
)

type Operation string

// List of IR operations
const (
	OpAssign Operation = "="
	OpAdd    Operation = "+"
	OpSub    Operation = "-"
	OpMul    Operation = "*"
	OpDiv    Operation = "/"
	OpMod    Operation = "%"
	OpAnd    Operation = "&&"
	OpOr     Operation = "||"
	OpNot    Operation = "!"
	OpEq     Operation = "=="
	OpNeq    Operation = "!="
	OpLt     Operation = "<"
	OpLe     Operation = "<="
	OpGt     Operation = ">"
	OpGe     Operation = ">="
	OpJmp    Operation = "jmp"
	OpJmpf   Operation = "jmpf"
	OpJmpt   Operation = "jmpt"
	OpCall   Operation = "call"
	OpRet    Operation = "ret"
	OpArg    Operation = "arg"
	OpParam  Operation = "param"
	OpLabel  Operation = "label"
	OpPrint  Operation = "print"
	OpNop    Operation = "nop"
	OpEnd    Operation = "end"
)

type Instruction struct {
	Op Operation

	Arg1 any
	Arg2 any
	Arg3 any

	Type lexer.TokenType // Result type (for assembly code generation)
}

// String returns a string representation of the instruction
func (i Instruction) String() string {
	arg1 := ""
	if i.Arg1 != nil {
		arg1 = fmt.Sprintf("%v", i.Arg1)
	}

	arg2 := ""
	if i.Arg2 != nil {
		arg2 = fmt.Sprintf("%v", i.Arg2)
	}

	arg3 := ""
	if i.Arg3 != nil {
		arg3 = fmt.Sprintf("%v", i.Arg3)
	}

	return fmt.Sprintf("(%s, %v, %v, %v, %v)", i.Op, arg1, arg2, arg3, i.Type)
}

// GetLexOperation maps a lexer token type to an IR operation
func GetLexOperation(t lexer.TokenType) Operation {
	switch t {
	case lexer.ASSIGN:
		return OpAssign
	case lexer.PLUS:
		return OpAdd
	case lexer.MINUS:
		return OpSub
	case lexer.MULT:
		return OpMul
	case lexer.DIV:
		return OpDiv
	case lexer.MOD:
		return OpMod
	case lexer.AND:
		return OpAnd
	case lexer.OR:
		return OpOr
	case lexer.NOT:
		return OpNot
	case lexer.EQ:
		return OpEq
	case lexer.NE:
		return OpNeq
	case lexer.LT:
		return OpLt
	case lexer.LE:
		return OpLe
	case lexer.GT:
		return OpGt
	case lexer.GE:
		return OpGe
	default:
		return OpNop
	}
}
