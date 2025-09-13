package arm64_macos

import (
	"dolme/pkg/lexer"
	"dolme/pkg/parser/codegen"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
)

// collectStackLayout inspects the instruction list to determine which addresses
// should be allocated in the "global frame" vs function local frames. It then
// assigns byte offsets for each address
func (a *arm64Macos) collectStackLayout() {
	globalAddrs := make(map[int]struct{})          // set of global addresses
	funcAddrs := make(map[string]map[int]struct{}) // func name -> set of local addresses

	currFunc := ""
	inFunc := false

	funcParams := make(map[string]map[int]struct{})

	for _, instr := range a.pb {
		switch instr.Op {
		case codegen.OpLabel:
			// function starts
			if name, ok := instr.Arg1.(string); ok {
				currFunc = name
				inFunc = true
				if _, exists := funcAddrs[currFunc]; !exists {
					funcAddrs[currFunc] = make(map[int]struct{})
				}
			}
		case codegen.OpParam:
			// Arg1 is the address for the parameter, Arg2 is the parameter position
			if addr, ok := instr.Arg1.(int); ok {
				if inFunc && currFunc != "" {
					if _, exists := funcParams[currFunc]; !exists {
						funcParams[currFunc] = make(map[int]struct{})
					}
					funcParams[currFunc][addr] = struct{}{}
					// record param type per function to avoid type leaks between functions
					if _, ok := a.funcTypes[currFunc]; !ok {
						a.funcTypes[currFunc] = make(map[int]lexer.TokenType)
					}
					a.funcTypes[currFunc][addr] = instr.Type
				} else {
					// fallback: treat param outside a function as global
					globalAddrs[addr] = struct{}{}
				}
			}
		case codegen.OpRet:
			// keep inFunc true; function may have multiple returns
		case codegen.OpEnd:
			// End of a function - reset function context
			inFunc = false
			currFunc = ""
		default:
			// collect integer addresses used in this instruction
			addrs := a.instructionAddresses(instr)
			for _, addr := range addrs {
				if addr >= 800 {
					// local variable
					if inFunc && currFunc != "" {
						funcAddrs[currFunc][addr] = struct{}{}
					} else {
						// if we somehow see a local outside a function, treat as global
						globalAddrs[addr] = struct{}{}
					}
				} else {
					// temps (<800) used inside a function are treated as locals; at top-level they are globals
					if inFunc && currFunc != "" {
						funcAddrs[currFunc][addr] = struct{}{}
					} else {
						globalAddrs[addr] = struct{}{}
					}
				}
			}
			// if this instruction writes to a local destination (Arg3) and has a known type,
			// record that type in the per-function map to disambiguate float/int locals.
			if inFunc && currFunc != "" && instr.Arg3 != nil {
				if dst, ok := instr.Arg3.(int); ok && instr.Type != 0 {
					if _, ok := a.funcTypes[currFunc]; !ok {
						a.funcTypes[currFunc] = make(map[int]lexer.TokenType)
					}
					a.funcTypes[currFunc][dst] = instr.Type
				}
			}
		}
	}

	// assign offsets for globals (8 bytes each)
	offset := 0
	if len(globalAddrs) > 0 {
		addrs := make([]int, 0, len(globalAddrs))
		for addr := range globalAddrs {
			addrs = append(addrs, addr)
		}
		sort.Ints(addrs)
		for _, addr := range addrs {
			a.globalOffsets[addr] = offset
			offset += 16
		}
	}
	// round to 16
	a.globalSize = ((offset + 15) / 16) * 16

	for fname, set := range funcAddrs {
		// count params for this function (0 if none)
		paramSet := make([]int, 0)
		if pset, ok := funcParams[fname]; ok {
			for addr := range pset {
				paramSet = append(paramSet, addr)
			}
			sort.Ints(paramSet)
		}

		// base offset for locals should start after all params
		base := 0
		// each parameter gets 16 bytes (preserve your previous slot size)
		base += len(paramSet) * 16

		// record param offsets (so addrOffset can return them if needed)
		if _, exists := a.funcLocals[fname]; !exists {
			a.funcLocals[fname] = make(map[int]int)
		}
		for i, addr := range paramSet {
			// param slots are at offsets 0, 16, 32, ... relative to function frame area
			a.funcLocals[fname][addr] = i * 16
		}

		// now assign locals after the params (so locals won't collide with param slots)
		lo := base
		// make sure locals map exists
		if _, exists := a.funcLocals[fname]; !exists {
			a.funcLocals[fname] = make(map[int]int)
		}
		if len(set) > 0 {
			addrs := make([]int, 0, len(set))
			for addr := range set {
				// skip addresses that are parameters
				if _, isParam := funcParams[fname][addr]; isParam {
					continue
				}
				addrs = append(addrs, addr)
			}
			sort.Ints(addrs)
			for _, addr := range addrs {
				a.funcLocals[fname][addr] = lo
				lo += 16
			}
		}
		a.localSizes[fname] = ((lo + 15) / 16) * 16
	}
}

// instructionAddresses returns a slice of integer addresses referenced by instr (Arg1, Arg2, Arg3)
// only returns values that are of type int
func (a *arm64Macos) instructionAddresses(instr codegen.Instruction) []int {
	out := make([]int, 0, 3)
	if v, ok := instr.Arg1.(int); ok {
		out = append(out, v)
	}
	if v, ok := instr.Arg2.(int); ok {
		out = append(out, v)
	}
	if v, ok := instr.Arg3.(int); ok {
		out = append(out, v)
	}
	return out
}

// collectLabelsAndCallArgs builds mapping of PB indices to labels for jump targets
// and collects arguments for OpCall by finding preceding OpArg instructions
func (a *arm64Macos) collectLabelsAndCallArgs() {
	// find jump targets and labels
	for idx, instr := range a.pb {
		switch instr.Op {
		case codegen.OpJmp, codegen.OpJmpf, codegen.OpJmpt:
			if t, ok := instr.Arg3.(int); ok && t >= 0 && t <= len(a.pb) {
				// create label if not exists
				if _, exists := a.pbLabels[t]; !exists {
					a.pbLabels[t] = fmt.Sprintf("L%d", t)
				}
			}
		case codegen.OpLabel:
			if name, ok := instr.Arg1.(string); ok {
				// map label index to function label
				a.pbLabels[idx] = "_" + name
			}
		}
	}

	// collect call args: look backward from OpCall to collect preceding OpArg entries
	for idx, instr := range a.pb {
		if instr.Op == codegen.OpCall {
			argCount := 0
			if n, ok := instr.Arg2.(int); ok {
				argCount = n
			}
			args := make([]codegen.Instruction, argCount)
			// scan backwards collecting OpArg with matching Arg2 (position)
			for j := idx - 1; j >= 0; j-- {
				if argCount == 0 || a.pb[j].Op == codegen.OpCall {
					break
				}
				pj := a.pb[j]
				if pj.Op != codegen.OpArg {
					continue
				}
				if pos, ok := pj.Arg2.(int); ok {
					if pos >= 0 && pos < argCount {
						args[pos] = a.pb[j]
					}
				}
			}
			a.callArgs[idx] = args
		}
	}
}

// emitMainAndFunctions emits the assembly for top-level code (main) and any functions found in PB
// Top-level instructions are emitted as part of `_main`. Functions are emitted inline when their OpLabel is encountered
func (a *arm64Macos) emitMainAndFunctions() {
	a.emitFunctions()

	a.emitMain()
}

// emitFunctions emits all functions found in PB as separate labels with prologue/epilogue
func (a *arm64Macos) emitFunctions() {
	for idx := 0; idx < len(a.pb); idx++ {
		instr := a.pb[idx]
		if instr.Op != codegen.OpLabel {
			continue
		}
		// Found a function label
		name, _ := instr.Arg1.(string)
		if name == "" {
			continue
		}
		label := "_" + name

		// Determine function range explicitly
		endIdx := a.findFunctionEnd(idx)

		a.addText("") // blank line before function
		a.addText(fmt.Sprintf("%s:", label))
		// function prologue (standard AArch64)
		a.addText("\tstp\tX29, X30, [SP, #-16]!")
		a.addText("\tmov\tX29, SP")
		// allocate locals for this function if any
		size := a.localSizes[name]
		if size > 0 {
			a.addText(fmt.Sprintf("\tsub\tSP, SP, #%d", size))
		}

		// Emit instructions strictly in [idx+1 .. endIdx]
		currentFunc := name
		a.currentFunc = name
		currentFuncLocalsSize := size
		for j := idx + 1; j <= endIdx; j++ {
			// skip OpEnd if it accidentally falls in range (defensive)
			if a.pb[j].Op == codegen.OpEnd {
				continue
			}
			// emit any non-function label for this PB index
			if lbl, ok := a.pbLabels[j]; ok {
				a.addText(fmt.Sprintf("%s:", lbl))
			}

			in := a.pb[j]
			switch in.Op {
			case codegen.OpParam:
				paramAddr, _ := in.Arg1.(int)
				pos, _ := in.Arg2.(int)
				off := a.addrOffset(paramAddr, currentFunc)

				if pos >= 0 && pos <= 7 {
					if in.Type == lexer.FLOAT {
						a.addText(fmt.Sprintf("\tldr\td0, [X29, #%d]", 16+pos*16))
						a.addText(fmt.Sprintf("\tstr\td0, [SP, #%d]", off))
					} else {
						a.addText(fmt.Sprintf("\tldr\tx0, [X29, #%d]", 16+pos*16))
						a.addText(fmt.Sprintf("\tstr\tx0, [SP, #%d]", off))
					}
				} else {
					log.Error("param not supported", "pos", pos, "func", currentFunc)
				}
			case codegen.OpArg:
				// no-op in linear emission, handled at call sites
			case codegen.OpCall:
				a.emitCallAtIndex(in, j, currentFunc)
			case codegen.OpAssign:
				a.emitAssign(in, currentFunc)
			case codegen.OpAdd, codegen.OpSub, codegen.OpMul, codegen.OpDiv, codegen.OpMod, codegen.OpAnd, codegen.OpOr, codegen.OpEq, codegen.OpNeq, codegen.OpLt, codegen.OpLe, codegen.OpGt, codegen.OpGe:
				a.emitBinary(in, currentFunc)
			case codegen.OpPrint:
				a.emitPrint(in, currentFunc)
			case codegen.OpJmp:
				a.emitJmp(in)
			case codegen.OpJmpf, codegen.OpJmpt:
				a.emitJmpCond(in)
			case codegen.OpRet:
				// function return: handle Arg1 if present (we leave epilogue emission to the common epilogue)
				if in.Arg1 != nil {
					switch v := in.Arg1.(type) {
					case string:
						if strings.HasPrefix(v, "#") {
							val := v[1:]
							val = normalizeImmediate(val)
							a.addText(fmt.Sprintf("\tmov\tX0, #%s", val))
						}
					case int:
						off := a.addrOffset(v, currentFunc)
						switch in.Type {
						case lexer.FLOAT:
							a.addText(fmt.Sprintf("\tldr\td0, [SP, #%d]", off))
						case lexer.INT:
							a.addText(fmt.Sprintf("\tldr\tX0, [SP, #%d]", off))
						default:
							a.addText(fmt.Sprintf("\tldr\tX0, [SP, #%d]", off))
						}
						// a.addText(fmt.Sprintf("\tldr\tX0, [SP, #%d]", off))
					}
				}
			case codegen.OpNop:
				// ignore no-op
				continue
			default:
				a.addText(fmt.Sprintf("\t// unhandled op in func %s: %s %v %v %v", currentFunc, in.Op, in.Arg1, in.Arg2, in.Arg3))
			}
		}

		// Emit epilogue (safe to emit unconditionally)
		if currentFuncLocalsSize > 0 && false {
			a.addText(fmt.Sprintf("\tadd\tSP, SP, #%d", currentFuncLocalsSize))
		}

		if size > 0 {
			a.addText(fmt.Sprintf("\tadd\tSP, SP, #%d", size))
		}
		a.addText("\tldp\tX29, X30, [SP], #16")
		a.addText("\tret")

		// reset current function context
		a.currentFunc = ""

		// Advance outer loop to the end of the function range
		idx = endIdx
	}
}

// emitMain emits the top-level code as `_main`. It skips function bodies (ranges starting at OpLabel)
// because functions have already been emitted by emitFunctions
func (a *arm64Macos) emitMain() {
	a.addText("_main:")
	// standard frame prologue
	a.addText("\tstp\tX29, X30, [SP, #-16]!")
	a.addText("\tmov\tX29, SP")

	// allocate global frame space on stack
	if a.globalSize > 0 {
		a.addText(fmt.Sprintf("\tsub\tSP, SP, #%d", a.globalSize))
	}

	// iterate PB and emit only top-level instructions (skip function bodies and OpEnd)
	for idx := 0; idx < len(a.pb); idx++ {
		// if this is a function label, skip its function range
		if a.pb[idx].Op == codegen.OpLabel {
			endIdx := a.findFunctionEnd(idx)
			idx = endIdx
			continue
		}

		// skip OpEnd entries in top-level
		if a.pb[idx].Op == codegen.OpEnd {
			continue
		}

		// emit label if present for this PB index (and not a function label)
		if lbl, ok := a.pbLabels[idx]; ok {
			a.addText(fmt.Sprintf("%s:", lbl))
		}

		instr := a.pb[idx]
		switch instr.Op {
		case codegen.OpParam:
			// shouldn't happen at top-level, but remain tolerant
			paramAddr, _ := instr.Arg1.(int)
			pos, _ := instr.Arg2.(int)
			off := a.addrOffset(paramAddr, "")
			if pos >= 0 && pos <= 7 {
				a.addText(fmt.Sprintf("\tstr\tX%d, [SP, #%d]", pos, off))
			} else {
				a.addText(fmt.Sprintf("\t// param pos %d not supported", pos))
			}
		case codegen.OpArg:
			// ignore in linear emission; handled at call emit time
		case codegen.OpCall:
			a.emitCallAtIndex(instr, idx, "")
		case codegen.OpAssign:
			a.emitAssign(instr, "")
		case codegen.OpAdd, codegen.OpSub, codegen.OpMul, codegen.OpDiv, codegen.OpMod, codegen.OpAnd, codegen.OpOr, codegen.OpEq, codegen.OpNeq, codegen.OpLt, codegen.OpLe, codegen.OpGt, codegen.OpGe:
			a.emitBinary(instr, "")
		case codegen.OpPrint:
			a.emitPrint(instr, "")
		case codegen.OpJmp:
			a.emitJmp(instr)
		case codegen.OpJmpf, codegen.OpJmpt:
			a.emitJmpCond(instr)
		case codegen.OpRet:
			// return from main - cleanup global area and return
			if instr.Arg1 != nil {
				switch v := instr.Arg1.(type) {
				case string:
					if strings.HasPrefix(v, "#") {
						val := v[1:]
						val = normalizeImmediate(val)
						a.addText(fmt.Sprintf("\tmov\tX0, #%s", val))
					}
				case int:
					off := a.addrOffset(v, "")
					a.addText(fmt.Sprintf("\tldr\tX0, [SP, #%d]", off))
				}
			}
			// cleanup and return from main
			if a.globalSize > 0 {
				a.addText(fmt.Sprintf("\tadd\tSP, SP, #%d", a.globalSize))
			}
			a.addText("\tldp\tX29, X30, [SP], #16")
			a.addText("\tret")
		case codegen.OpNop:
			// ignore no-op
			continue
		default:
			a.addText(fmt.Sprintf("\t// unhandled op: %s %v %v %v", instr.Op, instr.Arg1, instr.Arg2, instr.Arg3))
		}
	}

	// if someone branched to the index just past the last PB instruction (end of main),
	// emit that label here so branches to it land at the epilogue.
	if lbl, ok := a.pbLabels[len(a.pb)]; ok {
		a.addText(fmt.Sprintf("%s:", lbl))
	}

	if a.globalSize > 0 {
		a.addText(fmt.Sprintf("\tadd\tSP, SP, #%d", a.globalSize))
	}
	a.addText("\tldp\tX29, X30, [SP], #16")
	a.addText("\tret")
	// end of main
	a.addText("\t// end of _main")
}

// addrOffset returns offset of a variable address within the appropriate frame
// if funcName is non-empty we look up locals; otherwise globalOffsets.
func (a *arm64Macos) addrOffset(addr int, funcName string) int {
	if funcName != "" {
		if addrMap, ok := a.funcLocals[funcName]; ok {
			if off, exists := addrMap[addr]; exists {
				return off
			}
		}
	}
	if off, ok := a.globalOffsets[addr]; ok {
		return off
	}
	// unknown addresses get 0 offset
	return 0
}

// getVarType returns the variable type for an address in the context of a function.
// For local/param addresses (>=800) it prefers the function-local map to avoid
// cross-function type leaks caused by reused address ranges. Falls back to the
// global codegen type table otherwise.
func (a *arm64Macos) getVarType(addr int, funcName string) lexer.TokenType {
	if funcName != "" && addr >= 800 {
		if m, ok := a.funcTypes[funcName]; ok {
			if t, ok2 := m[addr]; ok2 {
				return t
			}
		}
	}
	return a.cg.GetVariableType(addr)
}

// emitAssign handles OpAssign
func (a *arm64Macos) emitAssign(instr codegen.Instruction, funcName string) {
	// instr.Arg1 -> source (could be "#val" or addr int)
	// instr.Arg3 -> destination addr (int)
	destAddr, _ := instr.Arg3.(int)
	destOff := a.addrOffset(destAddr, funcName)

	// Determine if destination should hold float
	destIsFloat := false
	// Prefer explicit type on the instruction; fall back to variable table
	if instr.Type == lexer.FLOAT {
		destIsFloat = true
	} else if t := a.getVarType(destAddr, funcName); t == lexer.FLOAT {
		destIsFloat = true
	}

	// floating-point assignment
	if destIsFloat {
		switch v := instr.Arg1.(type) {
		case string:
			if strings.HasPrefix(v, "#") {
				val := normalizeImmediate(v[1:])
				// float immediate -> emit constant and load into d0
				// If immediate looks like an integer (no dot), scvtf path could be used,
				// but easiest is to emit as double constant to preserve precision.
				label := a.storeFloatConstant(val)
				// use x9 as temporary for address calculation
				a.addText(fmt.Sprintf("\tadrp\tx9, %s@PAGE", label))
				a.addText(fmt.Sprintf("\tadd\tx9, x9, %s@PAGEOFF", label))
				a.addText("\tldr\td0, [x9]")                            // load double into d0
				a.addText(fmt.Sprintf("\tstr\td0, [SP, #%d]", destOff)) // store double to dest
				return
			} else {
				// unexpected: treat like address string? Emit comment
				a.addText(fmt.Sprintf("\t// emitAssign: unexpected string source %s for float dest", v))
				return
			}
		case int:
			// source is an address holding a double or integer; check its type
			if a.getVarType(v, funcName) == lexer.FLOAT {
				srcOff := a.addrOffset(v, funcName)
				a.addText(fmt.Sprintf("\tldr\td0, [SP, #%d]", srcOff))
				a.addText(fmt.Sprintf("\tstr\td0, [SP, #%d]", destOff))
				return
			}
			// source is int -> load integer then convert to double
			srcOff := a.addrOffset(v, funcName)
			a.addText(fmt.Sprintf("\tldr\tx9, [SP, #%d]", srcOff))
			a.addText("\tscvtf\td0, x9")
			a.addText(fmt.Sprintf("\tstr\td0, [SP, #%d]", destOff))
			return
		default:
			a.addText("\t// emitAssign: unsupported Arg1 type for float assign")
			return
		}
	}

	// integer assignment path (existing behavior)
	switch v := instr.Arg1.(type) {
	case string:
		if strings.HasPrefix(v, "#") {
			val := normalizeImmediate(v[1:])
			// string literal?
			if strings.HasPrefix(val, "\"") || instr.Type == lexer.STRING {
				// emit cstring and point to it
				label := a.storeCString(val)
				a.addText(fmt.Sprintf("\tadrp\tX0, %s@PAGE", label))
				a.addText(fmt.Sprintf("\tadd\tX0, X0, %s@PAGEOFF", label))
				// store pointer to stack (64-bit)
				a.addText(fmt.Sprintf("\tstr\tX0, [SP, #%d]", destOff))
				return
			}
			// numeric immediates - use mov
			a.addText(fmt.Sprintf("\tmov\tX0, #%s", val))
		} else {
			// treated as address string? fallback: no-op
			a.addText(fmt.Sprintf("\t// assign source string (unknown): %s", v))
			return
		}
	case int:
		// source is an address - load from its frame
		srcOff := a.addrOffset(v, funcName)
		a.addText(fmt.Sprintf("\tldr\tX0, [SP, #%d]", srcOff))
	default:
		a.addText("\t// assign: unsupported Arg1 type")
		return
	}

	// store X0 into destination slot (integer)
	a.addText(fmt.Sprintf("\tstr\tX0, [SP, #%d]", destOff))
}

// emitBinary emits arithmetic and logical binary ops (supports int->float conversions)
func (a *arm64Macos) emitBinary(instr codegen.Instruction, funcName string) {
	// Arg1 and Arg2 are operands (int addresses or immediate strings), Arg3 is destination addr (int)
	var1 := instr.Arg1
	var2 := instr.Arg2
	destAddr, _ := instr.Arg3.(int)
	destOff := a.addrOffset(destAddr, funcName)

	// Determine whether to use FP path:
	useFloat := false
	// If instruction explicitly typed as float, prefer FP
	if instr.Type == lexer.FLOAT {
		useFloat = true
	} else {
		// Check operand types: if either operand is float, use float path
		if a.isOpFloat(var1, funcName) || a.isOpFloat(var2, funcName) {
			useFloat = true
		}
		// Also check variable type table for address operands
		if v1, ok := var1.(int); ok {
			if a.getVarType(v1, funcName) == lexer.FLOAT {
				useFloat = true
			}
		}
		if v2, ok := var2.(int); ok {
			if a.getVarType(v2, funcName) == lexer.FLOAT {
				useFloat = true
			}
		}
	}

	if useFloat {
		// Floating-point path: load into d0/d1 and use FP ops, store result as double
		a.loadOperandToFPReg("d0", codegen.Instruction{Op: codegen.OpNop, Arg1: var1, Arg2: nil, Arg3: nil, Type: instr.Type}, funcName)
		a.loadOperandToFPReg("d1", codegen.Instruction{Op: codegen.OpNop, Arg1: var2, Arg2: nil, Arg3: nil, Type: instr.Type}, funcName)

		switch instr.Op {
		case codegen.OpAdd:
			a.addText("\tfadd\td0, d0, d1")
		case codegen.OpSub:
			a.addText("\tfsub\td0, d0, d1")
		case codegen.OpMul:
			a.addText("\tfmul\td0, d0, d1")
		case codegen.OpDiv:
			a.addText("\tfdiv\td0, d0, d1")
		case codegen.OpMod:
			// float remainder not implemented here -> placeholder: call fmod would be needed
			a.addText("\t// float mod not implemented; result set to 0.0")
			label := a.storeFloatConstant("0.0")
			a.addText(fmt.Sprintf("\tadrp\tx9, %s@PAGE", label))
			a.addText(fmt.Sprintf("\tadd\tx9, x9, %s@PAGEOFF", label))
			a.addText("\tldr\td0, [x9]")
		case codegen.OpEq, codegen.OpNeq, codegen.OpLt, codegen.OpLe, codegen.OpGt, codegen.OpGe:
			// FP compare: use fcmp d0, d1 then cset x0, <cond>
			a.addText("\tfcmp\td0, d1")
			switch instr.Op {
			case codegen.OpEq:
				a.addText("\tcset\tX0, eq")
			case codegen.OpNeq:
				a.addText("\tcset\tX0, ne")
			case codegen.OpLt:
				a.addText("\tcset\tX0, lt")
			case codegen.OpLe:
				a.addText("\tcset\tX0, le")
			case codegen.OpGt:
				a.addText("\tcset\tX0, gt")
			case codegen.OpGe:
				a.addText("\tcset\tX0, ge")
			}
			// store integer boolean result
			a.addText(fmt.Sprintf("\tstr\tX0, [SP, #%d]", destOff))
			return
		default:
			a.addText(fmt.Sprintf("\t// unhandled float op: %s", instr.Op))
		}
		// store double result
		a.addText(fmt.Sprintf("\tstr\td0, [SP, #%d]", destOff))
		return
	}

	// integer path
	a.loadOperandToReg("X0", var1, funcName)
	a.loadOperandToReg("X1", var2, funcName)

	switch instr.Op {
	case codegen.OpAdd:
		a.addText("\tadd\tX0, X0, X1")
	case codegen.OpSub:
		a.addText("\tsub\tX0, X0, X1")
	case codegen.OpMul:
		a.addText("\tmul\tX0, X0, X1")
	case codegen.OpDiv:
		a.addText("\tsdiv\tX0, X0, X1")
	case codegen.OpMod:
		a.addText("\tsdiv\tX2, X0, X1")
		a.addText("\tmul\tX2, X2, X1")
		a.addText("\tsub\tX0, X0, X2")
	case codegen.OpAnd:
		a.addText("\tand\tX0, X0, X1")
	case codegen.OpOr:
		a.addText("\torr\tX0, X0, X1")
	case codegen.OpEq:
		a.addText("\tcmp\tX0, X1")
		a.addText("\tcset\tX0, eq")
	case codegen.OpNeq:
		a.addText("\tcmp\tX0, X1")
		a.addText("\tcset\tX0, ne")
	case codegen.OpLt:
		a.addText("\tcmp\tX0, X1")
		a.addText("\tcset\tX0, lt")
	case codegen.OpLe:
		a.addText("\tcmp\tX0, X1")
		a.addText("\tcset\tX0, le")
	case codegen.OpGt:
		a.addText("\tcmp\tX0, X1")
		a.addText("\tcset\tX0, gt")
	case codegen.OpGe:
		a.addText("\tcmp\tX0, X1")
		a.addText("\tcset\tX0, ge")
	default:
		a.addText(fmt.Sprintf("\t// unhandled binary op: %s", instr.Op))
	}

	// store integer result to destination
	a.addText(fmt.Sprintf("\tstr\tX0, [SP, #%d]", destOff))
}

// isOpFloat determines whether operand should be treated as float
func (a *arm64Macos) isOpFloat(op any, funcName string) bool {
	switch v := op.(type) {
	case string:
		// immediate literal like "#3.14" or "#3"
		if strings.HasPrefix(v, "#") {
			val := v[1:]
			// consider float if contains a dot or has exponent part
			if strings.Contains(val, ".") || strings.ContainsAny(val, "eE") {
				return true
			}
			// if it cannot be parsed as int but can be parsed as float, consider float
			if _, err := strconv.ParseInt(val, 10, 64); err != nil {
				if _, err2 := strconv.ParseFloat(val, 64); err2 == nil {
					return true
				}
			}
			return false
		}
		return false
	case int:
		// query per-function variable type table first
		return a.getVarType(v, funcName) == lexer.FLOAT
	default:
		return false
	}
}

// loadOperandToReg loads operand into specified register (e.g., "X0" or "X1")
func (a *arm64Macos) loadOperandToReg(reg string, op any, funcName string) {
	switch v := op.(type) {
	case string:
		if strings.HasPrefix(v, "#") {
			val := normalizeImmediate(v[1:])
			a.addText(fmt.Sprintf("\tmov\t%s, #%s", reg, val))
			return
		}
		a.addText(fmt.Sprintf("\t// loadOperandToReg: unsupported string operand %s", v))
	case int:
		off := a.addrOffset(v, funcName)
		a.addText(fmt.Sprintf("\tldr\t%s, [SP, #%d]", reg, off))
	default:
		a.addText("\t// loadOperandToReg: unsupported type")
	}
}

// loadOperandToFPReg loads an operand into a floating-point register (e.g., "d0" or "d1")
func (a *arm64Macos) loadOperandToFPReg(reg string, op codegen.Instruction, funcName string) {
	// reg is expected to be an FP register name like "d0" or "d1".
	// op.Arg1 is the operand: either a string immediate "#..." or an address (int).
	// prefer the variable type table when deciding whether an address is a float.

	switch v := op.Arg1.(type) {
	case string:
		if !strings.HasPrefix(v, "#") {
			log.Error("loadOperandToFPReg: unsupported string operand (missing #)", "value", v)
			return
		}
		val := normalizeImmediate(v[1:])

		// if instruction explicitly typed as FLOAT, or the literal looks like a float,
		// emit a double constant and load it into the FP register.
		if op.Type == lexer.FLOAT || strings.Contains(val, ".") || strings.ContainsAny(val, "eE") {
			label := a.storeFloatConstant(val)
			a.addText(fmt.Sprintf("\tadrp\tx9, %s@PAGE", label))
			a.addText(fmt.Sprintf("\tadd\tx9, x9, %s@PAGEOFF", label))
			a.addText(fmt.Sprintf("\tldr\t%s, [x9]", reg))
			return
		}

		// otherwise treat as integer immediate -> mov into x9 then convert.
		a.addText(fmt.Sprintf("\tmov\tx9, #%s", val))
		a.addText(fmt.Sprintf("\tscvtf\t%s, x9", reg))
		return

	case int:
		off := a.addrOffset(v, funcName)

		// check if this is a parameter by address range (800-899)
		isParam := v >= 800 && v < 900

		// for parameters, we should respect their original type rather than
		// assuming they need conversion to float
		if isParam {
			varType := a.getVarType(v, funcName)
			if varType == lexer.INT {
				// Load integer parameter, then convert to float
				a.addText(fmt.Sprintf("\tldr\tx9, [SP, #%d]", off))
				a.addText(fmt.Sprintf("\tscvtf\t%s, x9", reg))
				return
			}
		}

		// prefer the variable type table; fall back to op.Type if variable type unknown.
		varIsFloat := a.getVarType(v, funcName) == lexer.FLOAT
		if varIsFloat || op.Type == lexer.FLOAT {
			a.addText(fmt.Sprintf("\tldr\t%s, [SP, #%d]", reg, off))
			return
		}

		// otherwise load integer into x9 and convert to double.
		a.addText(fmt.Sprintf("\tldr\tx9, [SP, #%d]", off))
		a.addText(fmt.Sprintf("\tscvtf\t%s, x9", reg))
		return

	default:
		log.Error("loadOperandToFPReg: unsupported operand type", "type", fmt.Sprintf("%T", op.Arg1))
		return
	}
}

// emitPrint emits a simple printf-based print for ints and strings.
func (a *arm64Macos) emitPrint(instr codegen.Instruction, funcName string) {
	// Arg1 is the value to print
	switch v := instr.Arg1.(type) {
	case string:
		if strings.HasPrefix(v, "#") {
			val := normalizeImmediate(v[1:])
			// string literal?
			if strings.HasPrefix(val, "\"") {
				label := a.storeCString(val)
				// load pointer into X0
				a.addText(fmt.Sprintf("\tadrp\tX0, %s@PAGE", label))
				a.addText(fmt.Sprintf("\tadd\tX0, X0, %s@PAGEOFF", label))
				// call puts
				a.addText("\tbl\t_puts")
			} else {
				// numeric immediate print - use printf with format "%lld\n"
				fmtLabel := a.ensurePrintfIntFormat()
				// Prepare vararg area, set format and argument into stack save area
				a.addText("\t// prepare register-save / vararg area for printf")
				// Put format pointer in X0
				a.addText(fmt.Sprintf("\tadrp\tX0, %s@PAGE", fmtLabel))
				a.addText(fmt.Sprintf("\tadd\tX0, X0, %s@PAGEOFF", fmtLabel))
				// Move immediate into X1 and save into caller save area [SP,#0]
				a.addText(fmt.Sprintf("\tmov\tX1, #%s", val))
				a.addText("\tsub\tSP, SP, #64")
				a.addText("\tstr\tX1, [SP, #0]")
				// Call printf
				a.addText("\tbl\t_printf")
				// Restore SP
				a.addText("\tadd\tSP, SP, #64")
			}
		} else {
			a.addText(fmt.Sprintf("\t// print: unknown string %s", v))
		}
	case int:
		off := a.addrOffset(v, funcName)
		// choose format depending on variable type
		if a.getVarType(v, funcName) == lexer.FLOAT {
			fmtLabel := a.ensurePrintfFloatFormat()
			// Prepare vararg area for GP+FP: we only need FP saved for printf
			a.addText("\t// prepare register-save / vararg area for printf (float)")
			// Put format pointer in X0
			a.addText(fmt.Sprintf("\tadrp\tX0, %s@PAGE", fmtLabel))
			a.addText(fmt.Sprintf("\tadd\tX0, X0, %s@PAGEOFF", fmtLabel))
			// load double into d0 from variable slot and store into FP save area at SP+64
			a.addText(fmt.Sprintf("\tldr\td0, [SP, #%d]", off))
			a.addText("\tsub\tSP, SP, #192")
			a.addText("\tstr\td0, [SP]")
			// Call printf
			a.addText("\tbl\t_printf")
			// Restore SP
			a.addText("\tadd\tSP, SP, #192")
		} else {
			fmtLabel := a.ensurePrintfIntFormat()
			a.addText("\t// prepare register-save / vararg area for printf (int)")
			a.addText(fmt.Sprintf("\tadrp\tX0, %s@PAGE", fmtLabel))
			a.addText(fmt.Sprintf("\tadd\tX0, X0, %s@PAGEOFF", fmtLabel))
			a.addText(fmt.Sprintf("\tldr\tX1, [SP, #%d]", off))
			a.addText("\tsub\tSP, SP, #64")
			a.addText("\tstr\tX1, [SP, #0]")
			a.addText("\tbl\t_printf")
			a.addText("\tadd\tSP, SP, #64")
		}
	default:
		a.addText("\t// print: unsupported arg type")
	}
}

// emitCallAtIndex handles OpCall at PB index idx, setting up arguments and calling the function
func (a *arm64Macos) emitCallAtIndex(instr codegen.Instruction, idx int, funcName string) {
	funcNameStr, _ := instr.Arg1.(string)
	args := a.callArgs[idx]

	// Reserve register-save area on the stack (caller-side)
	// GP save area: 8 * 8 = 64 bytes
	// FP save area: 8 * 16 = 128 bytes
	// Total = 192 bytes
	// a.addText("\t// prepare register-save / vararg area (GP + FP)")

	// save caller SP as stable base to load args after reserving
	a.addText("\tmov\tx10, SP")
	// reserve 192 bytes (GP + FP save/vararg area)
	a.addText("\tsub\tSP, SP, #192")
	index := 0

	for i := range args {
		op := args[i]

		// decide arg type robustly: op.Type or variable type or literal shape
		argIsFloat := op.Type == lexer.FLOAT
		if !argIsFloat {
			switch v := op.Arg1.(type) {
			case int:
				if a.getVarType(v, funcName) == lexer.FLOAT {
					argIsFloat = true
				}
			case string:
				if strings.HasPrefix(v, "#") {
					val := normalizeImmediate(v[1:])
					if strings.Contains(val, ".") || strings.ContainsAny(val, "eE") {
						argIsFloat = true
					}
				}
			}
		}

		if argIsFloat {
			switch v := op.Arg1.(type) {
			case string:
				if strings.HasPrefix(v, "#") {
					val := normalizeImmediate(v[1:])
					if strings.Contains(val, ".") || strings.ContainsAny(val, "eE") {
						label := a.storeFloatConstant(val)
						a.addText(fmt.Sprintf("\tadrp\tx9, %s@PAGE", label))
						a.addText(fmt.Sprintf("\tadd\tx9, x9, %s@PAGEOFF", label))
						a.addText("\tldr\td0, [x9]")
					} else {
						a.addText(fmt.Sprintf("\tmov\tx9, #%s", val))
						a.addText("\tscvtf\td0, x9")
					}
				} else {
					log.Error("unexpected arg string", "arg", i, "func", funcNameStr)
					a.addText("\tmov\tx9, #0")
					a.addText("\tscvtf\td0, x9")
				}
			case int:
				off := a.addrOffset(v, funcName)
				if a.getVarType(v, funcName) == lexer.FLOAT {
					a.addText(fmt.Sprintf("\tldr\td0, [x10, #%d]", off))
				} else {
					a.addText(fmt.Sprintf("\tldr\tx9, [x10, #%d]", off))
					a.addText("\tscvtf\td0, x9")
				}
			default:
				log.Error("Unsupported arg type", "arg", i, "func", funcNameStr)
				a.addText("\tmov\tx9, #0")
				a.addText("\tscvtf\td0, x9")
			}
			a.addText(fmt.Sprintf("\tstr\td0, [SP, #%d]", index*16))
			index++
		} else {
			switch v := op.Arg1.(type) {
			case string:
				if strings.HasPrefix(v, "#") {
					val := normalizeImmediate(v[1:])
					a.addText(fmt.Sprintf("\tmov\tx0, #%s", val))
				} else {
					log.Error("unexpected arg string", "arg", i, "func", funcNameStr)
					a.addText("\tmov\tx0, #0")
				}
			case int:
				off := a.addrOffset(v, funcName)
				a.addText(fmt.Sprintf("\tldr\tx0, [x10, #%d]", off))
			default:
				log.Error("Unsupported arg type", "arg", i, "func", funcNameStr)
				a.addText("\tmov\tx0, #0")
			}
			a.addText(fmt.Sprintf("\tstr\tx0, [SP, #%d]", index*16))
			index++
		}
	}

	if len(args) > 8 {
		log.Warn("emitCallAtIndex: more than 8 args not supported, extras ignored", "func", funcNameStr)
	}

	// now call the function
	a.addText(fmt.Sprintf("\t// call %s", funcNameStr))
	a.addText(fmt.Sprintf("\tbl\t_%s", funcNameStr))

	// deallocate the 192-byte register-save area
	a.addText("\tadd\tSP, SP, #192")

	// store return value (X0 or d0) into return-temp slot if provided
	if retAddr, ok := instr.Arg3.(int); ok {
		// determine return type
		retIsFloat := false
		if instr.Type == lexer.FLOAT {
			retIsFloat = true
		} else if a.getVarType(retAddr, funcName) == lexer.FLOAT {
			retIsFloat = true
		}

		if retIsFloat {
			off := a.addrOffset(retAddr, funcName)
			a.addText(fmt.Sprintf("\tstr\td0, [SP, #%d]", off))
		} else {
			off := a.addrOffset(retAddr, funcName)
			a.addText(fmt.Sprintf("\tstr\tX0, [SP, #%d]", off))
		}
	}

	// restore stack (undo register-save area reservation)
}

// emitJmp emits unconditional jump by mapping PB index to label
func (a *arm64Macos) emitJmp(instr codegen.Instruction) {
	if t, ok := instr.Arg3.(int); ok {
		if lbl, exists := a.pbLabels[t]; exists {
			a.addText(fmt.Sprintf("\tb\t%s", lbl))
			return
		}
	}
	a.addText("\t// jmp: invalid target")
}

// emitJmpCond emits conditional jumps based on jmpt/jmpf (Arg1 is condition addr)
func (a *arm64Macos) emitJmpCond(instr codegen.Instruction) {
	condAddr, _ := instr.Arg1.(int)
	target, _ := instr.Arg3.(int)
	// load condition into X0
	off := a.addrOffset(condAddr, a.currentFunc)
	a.addText(fmt.Sprintf("\tldr\tX0, [SP, #%d]", off))
	// compare to zero
	a.addText("\tcmp\tX0, #0")
	switch instr.Op {
	case codegen.OpJmpt:
		// jump if true (non-zero)
		if lbl, exists := a.pbLabels[target]; exists {
			a.addText(fmt.Sprintf("\tb.ne\t%s", lbl))
			return
		}
	case codegen.OpJmpf:
		// jump if false (zero)
		if lbl, exists := a.pbLabels[target]; exists {
			a.addText(fmt.Sprintf("\tb.eq\t%s", lbl))
			return
		}
	default:
	}

	log.Error("emitJmpCond: invalid target or op", "instr")
}

// storeCString stores a Go-like string literal into the cstring section and returns its label.
func (a *arm64Macos) storeCString(lit string) string {
	// unquote if quoted
	val := lit
	if strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"") {
		val = val[1 : len(val)-1]
	}
	label := fmt.Sprintf("__dolme_str_%d", a.strCounter)
	a.strCounter++

	// Emit into cstring section via helper method
	a.addCString(fmt.Sprintf("%s:", label))
	a.addCString(fmt.Sprintf("\t.asciz\t\"%s\"", escapeString(val)))
	return label
}

// storeFloatConstant stores a floating-point constant into the data section and returns its label.
func (a *arm64Macos) storeFloatConstant(lit string) string {
	val := lit
	if strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"") {
		val = val[1 : len(val)-1]
	}

	label := fmt.Sprintf("__dolme_float_%d", a.strCounter)
	a.strCounter++

	a.data.WriteString(fmt.Sprintf("%s:\n\t.double\t%s\n", label, val))
	return label
}

// ensurePrintfIntFormat ensures we have a "%lld\n" C-format string available and returns its label.
func (a *arm64Macos) ensurePrintfIntFormat() string {
	label := "__dolme_printf_int"
	if !strings.Contains(a.cstring.String(), label+":") {
		a.addCString(fmt.Sprintf("%s:", label))
		a.addCString("\t.asciz\t\"%lld\\n\"")
	}
	return label
}

// ensurePrintfFloatFormat ensures we have a "%f\n" C-format string available and returns its label.
func (a *arm64Macos) ensurePrintfFloatFormat() string {
	label := "__dolme_printf_float"
	if !strings.Contains(a.cstring.String(), label+":") {
		a.addCString(fmt.Sprintf("%s:", label))
		a.addCString("\t.asciz\t\"%.20lf\\n\"")
	}
	return label
}
