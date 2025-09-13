package interpreter

import (
	"fmt"
	"math"
	"os"

	"dolme/pkg/lexer"
	"dolme/pkg/parser/codegen"
)

// Exec runs a PB program with the default step function and stdout as writer
func Exec(pb []codegen.Instruction) error {
	it := NewInterpreter(pb, WithWriter(os.Stdout))
	it.SetExecStep(coreStep)
	return it.Run()
}

// coreStep is the main single-step execution function
// it returns (halted, error).
func coreStep(i *Interpreter) (bool, error) {
	// fetch current PC and instruction
	pc := i.PC()
	if pc < 0 || pc >= len(i.pb) {
		// halt if PC goes out of bounds
		return true, nil
	}

	in := i.pb[pc]

	switch in.Op {
	case codegen.OpNop:
		i.SetPC(pc + 1)
		return false, nil

	case codegen.OpEnd:
		// always advance past OpEnd; OpRet handles function returns, and top-level continues
		i.SetPC(pc + 1)
		return false, nil

	case codegen.OpLabel:
		// at top-level, skip function bodies entirely; inside a function just advance
		if i.currentFrame() == nil {
			if end, ok := i.funcEnd[pc]; ok {
				i.SetPC(end + 1)
			} else {
				i.SetPC(pc + 1)
			}
		} else {
			i.SetPC(pc + 1)
		}
		return false, nil

	case codegen.OpAssign:
		// Arg1 can be immediate "#..." or address (int). Arg3 is destination addr (int)
		dst, _ := in.Arg3.(int)
		val, err := i.loadOperand(in.Arg1, in.Type)
		if err != nil {
			return false, err
		}
		i.SetVar(dst, val)
		i.SetPC(pc + 1)
		return false, nil

	case codegen.OpAdd, codegen.OpSub, codegen.OpMul, codegen.OpDiv, codegen.OpMod,
		codegen.OpAnd, codegen.OpOr,
		codegen.OpEq, codegen.OpNeq, codegen.OpLt, codegen.OpLe, codegen.OpGt, codegen.OpGe:
		// Arg1, Arg2 operands; Arg3 destination
		dst, _ := in.Arg3.(int)
		v1, err := i.loadOperand(in.Arg1, in.Type)
		if err != nil {
			return false, err
		}
		v2, err := i.loadOperand(in.Arg2, in.Type)
		if err != nil {
			return false, err
		}
		res, err := i.evalBinary(in.Op, v1, v2, in.Type)
		if err != nil {
			return false, err
		}
		i.SetVar(dst, res)
		i.SetPC(pc + 1)
		return false, nil

	case codegen.OpPrint:
		// ensure writer
		if i.out == nil {
			i.out = os.Stdout
		}
		// Arg1 is operand to print
		v, err := i.loadOperand(in.Arg1, in.Type)
		if err != nil {
			return false, err
		}
		switch v.Kind {
		case KindFloat:
			fmt.Fprintf(i.out, "%.20f\n", v.F64)
		case KindInt:
			fmt.Fprintf(i.out, "%d\n", v.I64)
		case KindBool:
			if v.Bool {
				fmt.Fprintln(i.out, "true")
			} else {
				fmt.Fprintln(i.out, "false")
			}
		case KindString:
			fmt.Fprintln(i.out, v.Str)
		default:
			fmt.Fprintln(i.out, "<nil>")
		}
		i.SetPC(pc + 1)
		return false, nil

	case codegen.OpParam:
		// parameters are handled by call setup; no action here
		i.SetPC(pc + 1)
		return false, nil

	case codegen.OpArg:
		// Stage argument in interpreter argument buffer
		pos, _ := in.Arg2.(int)
		val, err := i.loadOperand(in.Arg1, in.Type)
		if err != nil {
			return false, err
		}
		i.StageArg(pos, val)
		i.SetPC(pc + 1)
		return false, nil

	case codegen.OpCall:
		// Arg1 funcName, Arg2 argCount, Arg3 return temp address
		funcName, _ := in.Arg1.(string)
		argCount, _ := in.Arg2.(int)
		retTemp, _ := in.Arg3.(int)
		// Determine return-to IP (next instruction)
		returnTo := pc + 1
		// Push callee frame
		callee := i.PushFrame(funcName, returnTo, retTemp)
		// Move staged args into callee parameter slots 800,801,...
		for p := 0; p < argCount; p++ {
			if v, ok := i.ConsumeArg(p); ok {
				callee.Locals[LocalAddrBase+p] = v
			} else {
				// default missing args to 0 (int)
				callee.Locals[LocalAddrBase+p] = newInt(0)
			}
		}
		// clear any remaining staged args
		i.ClearArgs()
		// set callee PC to first instruction after label
		start := i.funcIndex[funcName]
		i.SetPC(start + 1)
		return false, nil

	case codegen.OpRet:
		// return from function (or end of top-level "main" block without frame)
		var retVal Value
		if in.Arg1 != nil {
			v, err := i.loadOperand(in.Arg1, in.Type)
			if err != nil {
				return false, err
			}
			retVal = v
		}
		if fr := i.currentFrame(); fr != nil {
			// pop frame and write return value to caller
			done := i.PopFrame()
			// if the caller expects a return temp, store there in caller scope (globals or caller locals)
			if done.RetTemp != 0 && done.RetTemp != -1 {
				if cf := i.currentFrame(); cf != nil {
					cf.Locals[done.RetTemp] = retVal
				} else {
					i.globals[done.RetTemp] = retVal
				}
			}
			// continue at return location
			i.SetPC(done.ReturnToIP)
			return false, nil
		}
		// top-level ret => halt
		return true, nil

	case codegen.OpJmp:
		// unconditional jump
		target, _ := in.Arg3.(int)
		i.SetPC(target)
		return false, nil

	case codegen.OpJmpf:
		// jump if false (zero)
		condAddr, _ := in.Arg1.(int)
		cond, _ := i.GetVar(condAddr)
		b, err := cond.AsBool()
		if err != nil {
			return false, err
		}
		if !b {
			target, _ := in.Arg3.(int)
			i.SetPC(target)
		} else {
			i.SetPC(pc + 1)
		}
		return false, nil

	case codegen.OpJmpt:
		// jump if true (non-zero)
		condAddr, _ := in.Arg1.(int)
		cond, _ := i.GetVar(condAddr)
		b, err := cond.AsBool()
		if err != nil {
			return false, err
		}
		if b {
			target, _ := in.Arg3.(int)
			i.SetPC(target)
		} else {
			i.SetPC(pc + 1)
		}
		return false, nil

	default:
		// dor anything unhandled, print to stderr and stop
		fmt.Fprintf(os.Stderr, "Unhandled op at %d: %s (%v,%v,%v) type=%v\n", pc, in.Op, in.Arg1, in.Arg2, in.Arg3, in.Type)
		return true, nil
	}
}

// loadOperand resolves an operand that may be:
// - immediate string "#..."
// - address int
// and converts to the type hint when appropriate.
func (i *Interpreter) loadOperand(op any, hint lexer.TokenType) (Value, error) {
	switch v := op.(type) {
	case string:
		if len(v) > 0 && v[0] == '#' {
			val, err := parseImmediate(v)
			if err != nil {
				return Value{}, err
			}
			return val, nil
		}
		return Value{}, fmt.Errorf("unexpected string operand: %q", v)

	case int:
		val, ok := i.GetVar(v)
		if !ok {
			// default-init based on hint
			switch hint {
			case lexer.FLOAT:
				return newFloat(0), nil
			case lexer.INT, lexer.BOOL:
				return newInt(0), nil
			case lexer.STRING:
				return newString(""), nil
			default:
				return Value{}, nil
			}
		}
		return val, nil

	default:
		return Value{}, fmt.Errorf("unsupported operand type: %T", op)
	}
}

// evalBinary evaluates a binary operation on two Values, with an optional type hint
func (i *Interpreter) evalBinary(op codegen.Operation, a, b Value, hint lexer.TokenType) (Value, error) {
	switch op {
	case codegen.OpAdd, codegen.OpSub, codegen.OpMul, codegen.OpDiv, codegen.OpMod:
		// numeric
		// choose float if hint says float or any operand is float
		useFloat := hint == lexer.FLOAT || a.Kind == KindFloat || b.Kind == KindFloat
		if useFloat {
			af, err := a.AsFloat64()
			if err != nil {
				return Value{}, err
			}
			bf, err := b.AsFloat64()
			if err != nil {
				return Value{}, err
			}
			switch op {
			case codegen.OpAdd:
				return newFloat(af + bf), nil
			case codegen.OpSub:
				return newFloat(af - bf), nil
			case codegen.OpMul:
				return newFloat(af * bf), nil
			case codegen.OpDiv:
				return newFloat(af / bf), nil
			case codegen.OpMod:
				// emulate fmod
				return newFloat(math.Mod(af, bf)), nil
			}
		} else {
			ai, err := a.AsInt64()
			if err != nil {
				return Value{}, err
			}
			bi, err := b.AsInt64()
			if err != nil {
				return Value{}, err
			}
			switch op {
			case codegen.OpAdd:
				return newInt(ai + bi), nil
			case codegen.OpSub:
				return newInt(ai - bi), nil
			case codegen.OpMul:
				return newInt(ai * bi), nil
			case codegen.OpDiv:
				// integer division (signed)
				if bi == 0 {
					return Value{}, fmt.Errorf("division by zero")
				}
				return newInt(ai / bi), nil
			case codegen.OpMod:
				if bi == 0 {
					return Value{}, fmt.Errorf("modulo by zero")
				}
				return newInt(ai % bi), nil
			}
		}

	case codegen.OpAnd, codegen.OpOr:
		ab, err := a.AsBool()
		if err != nil {
			return Value{}, err
		}
		bb, err := b.AsBool()
		if err != nil {
			return Value{}, err
		}
		switch op {
		case codegen.OpAnd:
			return newBool(ab && bb), nil
		case codegen.OpOr:
			return newBool(ab || bb), nil
		}

	case codegen.OpEq, codegen.OpNeq, codegen.OpLt, codegen.OpLe, codegen.OpGt, codegen.OpGe:
		// comparison: use float if any float or hint float, else int
		useFloat := hint == lexer.FLOAT || a.Kind == KindFloat || b.Kind == KindFloat
		if useFloat {
			af, err := a.AsFloat64()
			if err != nil {
				return Value{}, err
			}
			bf, err := b.AsFloat64()
			if err != nil {
				return Value{}, err
			}
			switch op {
			case codegen.OpEq:
				return newBool(af == bf), nil
			case codegen.OpNeq:
				return newBool(af != bf), nil
			case codegen.OpLt:
				return newBool(af < bf), nil
			case codegen.OpLe:
				return newBool(af <= bf), nil
			case codegen.OpGt:
				return newBool(af > bf), nil
			case codegen.OpGe:
				return newBool(af >= bf), nil
			}
		} else {
			ai, err := a.AsInt64()
			if err != nil {
				return Value{}, err
			}
			bi, err := b.AsInt64()
			if err != nil {
				return Value{}, err
			}
			switch op {
			case codegen.OpEq:
				return newBool(ai == bi), nil
			case codegen.OpNeq:
				return newBool(ai != bi), nil
			case codegen.OpLt:
				return newBool(ai < bi), nil
			case codegen.OpLe:
				return newBool(ai <= bi), nil
			case codegen.OpGt:
				return newBool(ai > bi), nil
			case codegen.OpGe:
				return newBool(ai >= bi), nil
			}
		}

	default:
		return Value{}, fmt.Errorf("unsupported binary op: %s", op)
	}

	return Value{}, fmt.Errorf("unreachable for op: %s", op)
}
