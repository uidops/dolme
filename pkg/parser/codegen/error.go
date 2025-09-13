package codegen

import (
	"dolme/pkg/color"
	"dolme/pkg/lexer"
	"fmt"
)

func (c *Codegen) addError(e string) {
	c.errors = append(c.errors, e)
}

func (c *Codegen) addUndefinedVariableError(varName string, pos lexer.Position) {
	msg := color.RedText("Undefined variable") + " `" + color.BlueText(varName) + "`"
	msg += " at " + color.YellowText(fmt.Sprintf("Line: %d, Column %d", pos.Line, pos.Column))
	c.addError(msg)
}

func (c *Codegen) addTypeMismatchError(expected, found lexer.TokenType, pos lexer.Position) {
	msg := color.RedText("Type mismatch") + " expected " + color.BlueText(fmt.Sprintf("%v", expected)) + ", found " + color.BlueText(fmt.Sprintf("%v", found))
	msg += " at " + color.YellowText(fmt.Sprintf("Line: %d, Column %d", pos.Line, pos.Column))
	c.addError(msg)
}

func (c *Codegen) addUndefinedFunctionError(funcName string, pos lexer.Position) {
	msg := color.RedText("Undefined function") + " `" + color.BlueText(funcName) + "`"
	msg += " at " + color.YellowText(fmt.Sprintf("Line: %d, Column %d", pos.Line, pos.Column))
	c.addError(msg)
}

func (c *Codegen) addRedeclarationError(varName string, pos lexer.Position) {
	msg := color.RedText("Redeclaration of variable") + " `" + color.BlueText(varName) + "`"
	msg += " at " + color.YellowText(fmt.Sprintf("Line: %d, Column %d", pos.Line, pos.Column))
	c.addError(msg)
}

func (c *Codegen) GetErrors() []string {
	return c.errors
}
