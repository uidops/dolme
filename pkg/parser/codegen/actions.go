package codegen

import (
	"dolme/pkg/lexer"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
)

// labelAction pushes the current instruction index onto the stack for backpatching later
func (c *Codegen) labelAction() {
	c.push(c.i)
}

// labelWhileAction pushes a unique label for the start of a while loop onto the stack
func (c *Codegen) labelWhileAction() {
	c.pushString(fmt.Sprintf("$while_%d", c.i))
}

// saveAction saves the current instruction index by appending a NOP instruction and pushing the index onto the stack
func (c *Codegen) saveAction() {
	c.pb = append(c.pb, Instruction{OpNop, nil, nil, nil, lexer.EOF})
	c.push(c.i)
	c.i++
}

// saveBreakAction saves a break action by appending a NOP instruction and pushing the index and a unique break label onto the stack
func (c *Codegen) saveBreakAction() {
	c.pb = append(c.pb, Instruction{OpNop, nil, nil, nil, lexer.EOF})
	c.pushString(fmt.Sprintf("$break_%d", c.i))
	c.i++
}

// jmpNonBackpatchAction appends an unconditional jump instruction to the program and pops the top of the stack
func (c *Codegen) jmpNonBackpatchAction() {
	if c.ss.Size() >= 1 {
		topStr := c.topString()
		var location int

		if strings.HasPrefix(topStr, "$while_") || strings.HasPrefix(topStr, "$break_") {
			trimmed := topStr
			trimmed, _ = strings.CutPrefix(trimmed, "$while_")
			trimmed, _ = strings.CutPrefix(trimmed, "$break_")

			if n, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
				location = int(n)
			} else {
				location = c.top()
			}
		} else {
			location = c.top()
		}

		c.pb = append(c.pb, Instruction{Op: OpJmp, Arg1: nil, Arg2: nil, Arg3: location, Type: lexer.EOF})
		c.pop(1)
		c.i++
	}
}

// jmpfBreakAction handles break statements within loops by patching them to jump to the instruction after the loop
func (c *Codegen) jmpfBreakAction() {
	for c.ss.Size() >= 1 && strings.HasPrefix(c.topString(), "$break_") {
		location := c.topString()
		loc, err := strconv.ParseInt(strings.TrimPrefix(location, "$break_"), 10, 64)
		if err != nil {
			log.Error("Invalid break label", "label", location)
			continue
		}

		c.pb[int(loc)] = Instruction{Op: OpJmp, Arg1: nil, Arg2: nil, Arg3: c.i + 1, Type: lexer.EOF}
		c.popString(1)
	}

	c.jmpfAction()
}

// jmpAction appends an unconditional jump instruction to the program and patches the target address
func (c *Codegen) jmpAction() {
	if c.ss.Size() >= 1 {
		location := c.top()
		if location < len(c.pb) {
			c.pb[location] = Instruction{Op: OpJmp, Arg1: nil, Arg2: nil, Arg3: c.i, Type: lexer.EOF}
		}
		c.pop(1)
	}
}

// jmpfNormalAction appends a conditional jump instruction to the program and patches the target address
func (c *Codegen) jmpfNormalAction() {
	j := 0
	for c.ss.Size() >= 2 {
		if strings.HasPrefix(c.topStringMinus(j), "$break_") || strings.HasPrefix(c.topStringMinus(j), "$while_") {
			j++
			continue
		}

		location := c.topMinus(j)
		condition := c.topMinus(j + 1)
		if location < len(c.pb) {
			c.pb[location] = Instruction{Op: OpJmpf, Arg1: condition, Arg2: nil, Arg3: c.i, Type: lexer.EOF}
		}

		c.popAtOffsetEnd(j)
		c.popAtOffsetEnd(j)

		break
	}
}

// jmpfAction appends a conditional jump instruction to the program and patches the target address to the next instruction
func (c *Codegen) jmpfAction() {
	if c.ss.Size() >= 2 {
		location := c.top()
		condition := c.topMinus(1)
		if location < len(c.pb) {
			c.pb[location] = Instruction{Op: OpJmpf, Arg1: condition, Arg2: nil, Arg3: c.i + 1, Type: lexer.EOF}
		}
		c.pop(2)
	}
}

// binaryOpAction generates code for binary operations like addition, subtraction, etc.
func (c *Codegen) binaryOpAction(op Operation) {
	if c.ss.Size() >= 2 {
		op2 := c.top()
		op1 := c.topMinus(1)
		t := c.getTemp()

		newType := lexer.FLOAT
		if c.GetVariableType(op1) == lexer.INT && c.GetVariableType(op2) == lexer.INT {
			newType = lexer.INT
		}

		c.setVariableType(t, newType)

		c.pb = append(c.pb, Instruction{Op: op, Arg1: op1, Arg2: op2, Arg3: t, Type: newType})
		c.pop(2)
		c.push(t)
		c.i++
	}
}

// pushAction pushes a literal value onto the stack and generates an assignment instruction
func (c *Codegen) pushAction() {
	var value string

	switch c.currentToken.Type {
	case lexer.NUM:
		value = "#" + c.currentToken.Lexeme
	case lexer.TRUE:
		value = "#true"
	case lexer.FALSE:
		value = "#false"
	case lexer.STRING:
		value = "#" + c.currentToken.Lexeme
	default:
		value = "#" + c.currentToken.Lexeme
	}

	value_ := value[1:]
	type_ := lexer.EOF
	if value_ == "true" || value_ == "false" {
		type_ = lexer.BOOL
	} else if strings.Contains(value_, ".") {
		type_ = lexer.FLOAT
	} else if _, err := strconv.ParseInt(value_, 10, 64); err == nil {
		type_ = lexer.INT
	} else if _, err := strconv.ParseFloat(value_, 64); err == nil {
		type_ = lexer.FLOAT
	}

	t := c.getTemp()

	c.setVariableType(t, type_)

	c.pb = append(c.pb, Instruction{Op: OpAssign, Arg1: value, Arg2: nil, Arg3: t, Type: type_})
	c.push(t)
	c.i++
}

// loadAction looks up a variable's address and pushes it onto the stack
func (c *Codegen) loadAction() {
	varName := c.currentToken.Lexeme
	if addr, exists := c.getVariableAddress(varName); exists {
		c.push(addr)
	} else {
		c.addUndefinedVariableError(varName, c.currentToken.Pos)
	}
}

// assignAction generates code for assignment operations
func (c *Codegen) assignAction() {
	if c.ss.Size() >= 2 {
		value := c.top()
		targetAddr := c.topMinus(1)

		if c.GetVariableType(value) != c.GetVariableType(targetAddr) {
			c.addTypeMismatchError(c.GetVariableType(targetAddr), c.GetVariableType(value), c.currentToken.Pos)
			return
		}

		c.pb = append(c.pb, Instruction{Op: OpAssign, Arg1: value, Arg2: nil, Arg3: targetAddr, Type: c.GetVariableType(value)})
		c.pop(2)
		c.i++
	}
}

// defineAction generates code for variable declaration and initialization
func (c *Codegen) defineAction() {
	if c.ss.Size() >= 3 {
		value := c.top()
		varName := c.topStringMinus(2)

		if c.isVariableDeclared(varName) {
			c.addRedeclarationError(varName, c.currentToken.Pos)
		}

		varAddr := c.getVariable()
		c.declareVariable(varName, varAddr)

		c.setVariableType(varAddr, c.GetVariableType(value))

		c.pb = append(c.pb, Instruction{Op: OpAssign, Arg1: value, Arg2: nil, Arg3: varAddr, Type: c.GetVariableType(value)})
		c.pop(3)
		c.i++
	}
}

// printAction generates code for print statements
func (c *Codegen) printAction() {
	if c.ss.Size() >= 1 {
		a := c.top()
		c.pb = append(c.pb, Instruction{Op: OpPrint, Arg1: a, Arg2: nil, Arg3: nil, Type: c.GetVariableType(a)})
		c.pop(1)
		c.i++
	}
}

// notAction generates code for logical NOT operations
func (c *Codegen) notAction() {
	if c.ss.Size() >= 1 {
		op1 := c.top()
		temp := c.getTemp()

		c.pb = append(c.pb, Instruction{Op: OpNot, Arg1: op1, Arg2: nil, Arg3: temp, Type: lexer.EOF})
		c.pop(1)
		c.push(temp)
		c.i++
	}
}

// relAction generates code for relational operations
func (c *Codegen) relAction() {
	// Parsing order: BoolPrimary RelOp BoolPrimary @rel
	// Stack order: [op1, operator_string, op2] (top to bottom)
	if c.ss.Size() >= 3 {
		op2 := c.top()               // Second operand (right side)
		opStr := c.topStringMinus(1) // Operator
		op1 := c.topMinus(2)         // First operand (left side)

		var relOp Operation
		switch opStr {
		case "<":
			relOp = OpLt
		case ">":
			relOp = OpGt
		case "<=":
			relOp = OpLe
		case ">=":
			relOp = OpGe
		case "==":
			relOp = OpEq
		case "!=":
			relOp = OpNeq
		default:
			relOp = OpEq
		}

		temp := c.getTemp()

		newType := lexer.FLOAT
		if c.GetVariableType(op1) == lexer.INT && c.GetVariableType(op2) == lexer.INT {
			newType = lexer.INT
		}
		c.setVariableType(temp, lexer.BOOL)

		c.pb = append(c.pb, Instruction{Op: relOp, Arg1: op1, Arg2: op2, Arg3: temp, Type: newType})
		c.pop(3)
		c.push(temp)
		c.i++
	}
}

// functionStartAction handles the start of a function definition
func (c *Codegen) functionStartAction() {
	funcName := c.currentToken.Lexeme
	c.pb = append(c.pb, Instruction{Op: OpLabel, Arg1: funcName, Arg2: nil, Arg3: nil, Type: lexer.EOF})
	c.setInFunction(true)
	c.pushString(funcName)
	c.i++
}

// funcEndAction handles the end of a function definition
func (c *Codegen) funcEndAction() {
	// only add return if last instruction isn't already a return
	if len(c.pb) == 0 || c.pb[len(c.pb)-1].Op != OpRet {
		c.pb = append(c.pb, Instruction{Op: OpRet, Arg1: nil, Arg2: nil, Arg3: nil, Type: lexer.EOF})
		c.paramCounter = 0
		c.i++
	}

	c.setInFunction(false)
	c.pb = append(c.pb, Instruction{Op: OpEnd, Arg1: nil, Arg2: nil, Arg3: nil, Type: lexer.EOF})
	c.i++
}

// paramAction handles function parameter declarations
func (c *Codegen) paramAction() {
	if c.ss.Size() >= 2 {
		typeStr := c.topString()
		paramName := c.topStringMinus(1)
		paramAddr := c.getLocalVariable()

		c.declareVariable(paramName, paramAddr)
		c.setVariableType(paramAddr, lexer.Keywords[typeStr])

		c.pb = append(c.pb, Instruction{Op: OpParam, Arg1: paramAddr, Arg2: c.paramCounter, Arg3: nil, Type: lexer.Keywords[typeStr]})
		c.paramCounter += 1
		c.pop(2)
		c.i++
	}
}

// callStartAction prepares for a function call by pushing the function name onto the stack
func (c *Codegen) callStartAction() {
	funcName := c.currentToken.Lexeme
	c.pushString(funcName)
}

// callEndAction finalizes a function call by generating the call instruction and handling the return value
func (c *Codegen) callEndAction() {
	if c.ss.Size() >= 1 {
		funcName := c.topString()
		returnTemp := c.getTemp()

		ret, ok := c.functionReturns[funcName]
		if !ok {
			c.addUndefinedFunctionError(funcName, c.currentToken.Pos)
		}

		c.setVariableType(returnTemp, ret)

		c.pb = append(c.pb, Instruction{Op: OpCall, Arg1: funcName, Arg2: c.argsCounter, Arg3: returnTemp, Type: c.functionReturns[funcName]})
		c.argsCounter = 0

		c.pop(1)
		c.push(returnTemp)
		c.i++
	}
}

// argAction handles function call arguments by generating argument instructions
func (c *Codegen) argAction() {
	if c.ss.Size() >= 1 {
		arg := c.top()
		c.pb = append(c.pb, Instruction{Op: OpArg, Arg1: arg, Arg2: c.argsCounter, Arg3: nil, Type: c.GetVariableType(arg)})
		c.argsCounter += 1
		c.pop(1)
		c.i++
	}
}

// returnAction generates the return instruction for a function
func (c *Codegen) returnAction() {
	if c.ss.Size() >= 1 {
		returnVal := c.top()
		c.pb = append(c.pb, Instruction{Op: OpRet, Arg1: returnVal, Arg2: nil, Arg3: nil, Type: c.GetVariableType(returnVal)})
		c.pop(1)
	} else {
		c.pb = append(c.pb, Instruction{Op: OpRet, Arg1: nil, Arg2: nil, Arg3: nil, Type: lexer.EOF})
	}

	c.paramCounter = 0
	c.i++
}

// continueAction generates a jump instruction to the start of the loop for continue statements
func (c *Codegen) continueAction() {
	j := 0
	for c.ss.Size() >= 1 && !strings.HasPrefix(c.topStringMinus(j), "$while_") {
		j++
	}

	loc, err := strconv.ParseInt(strings.TrimPrefix(c.topStringMinus(j), "$while_"), 10, 64)
	if err != nil {
		log.Error("Invalid while label", "label", c.topStringMinus(j))
		return
	}

	c.pb = append(c.pb, Instruction{Op: OpJmp, Arg1: nil, Arg2: nil, Arg3: int(loc), Type: lexer.EOF})
	c.i++
}

// captureDeclVarAction stores the variable name for a variable declaration
func (c *Codegen) captureDeclVarAction() {
	c.pushString(c.currentToken.Lexeme)
}

// captureParamNameAction stores the parameter name for a function parameter
func (c *Codegen) captureParamNameAction() {
	c.pushString(c.currentToken.Lexeme)
}

// captureTypeAction stores the type for variable/parameter declaration
func (c *Codegen) captureTypeAction() {
	c.pushString(c.currentToken.Lexeme)
}

// funcReturnTypeAction stores the return type for a function
func (c *Codegen) funcReturnTypeAction() {
	if c.ss.Size() >= 1 {
		funcName := c.topString()
		c.functionReturns[funcName] = lexer.Keywords[c.currentToken.Lexeme]
		c.pop(1)
	}
}

// captureAssignTargetAction looks up or creates the variable address for an assignment target
func (c *Codegen) captureAssignTargetAction() {
	// look up assignment target variable address
	varName := c.currentToken.Lexeme
	if addr, exists := c.getVariableAddress(varName); exists {
		c.push(addr)
	} else {
		c.addUndefinedVariableError(varName, c.currentToken.Pos)
	}
}

// pushRelOpAction pushes the relational operator string onto the stack for relational operations
func (c *Codegen) pushRelOpAction() {
	c.pushString(c.currentToken.Lexeme)
}

// callAction handles function calls in assignment contexts
func (c *Codegen) callAction() {
	funcName := c.currentToken.Lexeme
	returnTemp := c.getTemp()

	c.setVariableType(returnTemp, c.functionReturns[funcName])

	c.pb = append(c.pb, Instruction{Op: OpCall, Arg1: funcName, Arg2: 0, Arg3: returnTemp, Type: c.functionReturns[funcName]})
	c.push(returnTemp)
	c.i++
}

// ExecuteAction executes the semantic action corresponding to the given action name
func (c *Codegen) ExecuteAction(actionName string) {
	// for i, instr := range c.pb {
	// 	fmt.Printf("%d: (%s, %v, %v, %v)\n", i, instr.Op, instr.Arg1, instr.Arg2, instr.Arg3)
	// }
	// fmt.Println()

	SemanticActions := map[string]func(){
		"@label":                 c.labelAction,
		"@save":                  c.saveAction,
		"@jmp_nonbackpatch":      c.jmpNonBackpatchAction,
		"@jmp":                   c.jmpAction,
		"@jmpf_normal":           c.jmpfNormalAction,
		"@jmpf":                  c.jmpfAction,
		"@jmpf_break":            c.jmpfBreakAction,
		"@add":                   func() { c.binaryOpAction(OpAdd) },
		"@sub":                   func() { c.binaryOpAction(OpSub) },
		"@mul":                   func() { c.binaryOpAction(OpMul) },
		"@div":                   func() { c.binaryOpAction(OpDiv) },
		"@mod":                   func() { c.binaryOpAction(OpMod) },
		"@push":                  c.pushAction,
		"@load":                  c.loadAction,
		"@assign":                c.assignAction,
		"@define":                c.defineAction,
		"@print":                 c.printAction,
		"@or":                    func() { c.binaryOpAction(OpOr) },
		"@and":                   func() { c.binaryOpAction(OpAnd) },
		"@not":                   c.notAction,
		"@rel":                   c.relAction,
		"@func_start":            c.functionStartAction,
		"@func_end":              c.funcEndAction,
		"@func_return_type":      c.funcReturnTypeAction,
		"@param":                 c.paramAction,
		"@call_start":            c.callStartAction,
		"@call_end":              c.callEndAction,
		"@call":                  c.callAction,
		"@arg":                   c.argAction,
		"@return":                c.returnAction,
		"@continue":              c.continueAction,
		"@capture_decl_var":      c.captureDeclVarAction,
		"@capture_type":          c.captureTypeAction,
		"@capture_assign_target": c.captureAssignTargetAction,
		"@capture_param_name":    c.captureParamNameAction,
		"@push_relop":            c.pushRelOpAction,
		"@save_break":            c.saveBreakAction,
		"@label_while":           c.labelWhileAction,
	}

	if action, exists := SemanticActions[actionName]; exists {
		action()
	} else {
		log.Error("Unknown semantic action", "action", actionName)
	}
}
