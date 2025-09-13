package interpreter

import (
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"

	"dolme/pkg/lexer"
	"dolme/pkg/parser/codegen"
)

const LocalAddrBase = 800

// Interpreter executes three-address IR (PB) produced by codegen
type Interpreter struct {
	pb []codegen.Instruction // program block (list of instructions)
	ip int                   // instruction pointer for top-level execution

	globals map[int]Value // global variables (addr -> value)

	stack []*Frame // call stack (frames)

	argBuf map[int]Value // argument staging buffer (pos -> value)

	labelIndex map[int]string             // PB index -> label string (e.g., function name)
	funcIndex  map[string]int             // function name -> PB index (label location)
	funcEnd    map[int]int                // label PB index -> end PB index (index of last instr in function)
	retTypes   map[string]lexer.TokenType // function name -> return type

	out io.Writer // output writer for print

	// Exec hook (implemented in another file, via SetExecStep)
	execStep func(*Interpreter) (halted bool, err error)

	maxSteps int // maximum steps (0 = unlimited)
	steps    int // steps executed
}

type Option func(*Interpreter)

// WithWriter sets the output writer for print statements
func WithWriter(w io.Writer) Option {
	return func(i *Interpreter) { i.out = w }
}

// WithMaxSteps sets a maximum number of interpreter steps before returning ErrMaxStepsExceeded
func WithMaxSteps(n int) Option {
	return func(i *Interpreter) { i.maxSteps = n }
}

// NewInterpreter creates a new Interpreter instance
func NewInterpreter(pb []codegen.Instruction, opts ...Option) *Interpreter {
	it := &Interpreter{
		pb:         append([]codegen.Instruction(nil), pb...),
		ip:         0,
		globals:    make(map[int]Value),
		stack:      make([]*Frame, 0, 8),
		argBuf:     make(map[int]Value),
		labelIndex: make(map[int]string),
		funcIndex:  make(map[string]int),
		funcEnd:    make(map[int]int),
		retTypes:   make(map[string]lexer.TokenType),
		out:        nil, // caller should set, or use WithWriter
		maxSteps:   0,   // 0 => unlimited
	}

	it.indexProgram()
	for _, o := range opts {
		o(it)
	}

	if it.out == nil {
		it.out = os.Stdout
	}

	if it.execStep == nil {
		it.execStep = coreStep
	}

	return it
}

// Load replaces the current program block with a new one, resetting state
func (i *Interpreter) Load(pb []codegen.Instruction) {
	i.pb = append([]codegen.Instruction(nil), pb...)
	i.Reset()
	i.indexProgram()
}

// Reset clears runtime state (globals, call stack, IP, counters)
func (i *Interpreter) Reset() {
	i.ip = 0
	i.globals = make(map[int]Value)
	i.stack = i.stack[:0]
	i.argBuf = make(map[int]Value)
	i.steps = 0
}

// Program returns the active PB
func (i *Interpreter) Program() []codegen.Instruction {
	return i.pb
}

// Output returns the output writer used for print
func (i *Interpreter) Output() io.Writer {
	return i.out
}

// SetExecStep installs the core step function (implemented in step.go or similar)
func (i *Interpreter) SetExecStep(fn func(*Interpreter) (bool, error)) {
	i.execStep = fn
}

// Step executes a single instruction, returning (halted, error)
func (i *Interpreter) Step() (bool, error) {
	if i.execStep == nil {
		return false, ErrNotImplemented
	}

	if i.maxSteps > 0 && i.steps >= i.maxSteps {
		return false, ErrMaxStepsExceeded
	}

	halted, err := i.execStep(i)
	i.steps++

	return halted, err
}

// Run executes until halt or error
func (i *Interpreter) Run() error {
	for {
		halted, err := i.Step()
		if err != nil {
			return err
		}

		if halted {
			return nil
		}
	}
}

// PC returns the current instruction pointer, considering call frames
func (i *Interpreter) PC() int {
	if f := i.currentFrame(); f != nil {
		return f.IP
	}

	return i.ip
}

// SetPC sets the current instruction pointer, considering call frames
func (i *Interpreter) SetPC(pc int) {
	if f := i.currentFrame(); f != nil {
		f.IP = pc
		return
	}

	i.ip = pc
}

// currentFrame returns the current call frame, or nil if none
func (i *Interpreter) currentFrame() *Frame {
	if len(i.stack) == 0 {
		return nil
	}

	return i.stack[len(i.stack)-1]
}

// PushFrame pushes a new call frame for the given function name, return IP, and return temp address
func (i *Interpreter) PushFrame(funcName string, retToIP, retTemp int) *Frame {
	frame := &Frame{
		FuncName:   funcName,
		IP:         i.funcIndex[funcName], // start at function label; executor should advance past label on first step
		Locals:     make(map[int]Value),
		ReturnToIP: retToIP,
		RetTemp:    retTemp,
	}

	i.stack = append(i.stack, frame)
	return frame
}

// PopFrame pops the current call frame
func (i *Interpreter) PopFrame() *Frame {
	if len(i.stack) == 0 {
		return nil
	}

	f := i.stack[len(i.stack)-1]
	i.stack = i.stack[:len(i.stack)-1]
	return f
}

// SetVar writes a value to an address, considering frame scoping for locals (>= LocalAddrBase)
func (i *Interpreter) SetVar(addr int, v Value) {
	if f := i.currentFrame(); f != nil && addr >= LocalAddrBase {
		f.Locals[addr] = v
		return
	}

	// Temps (<800) used inside functions can be treated as frame-locals too, but
	// we'll keep them in frame by preference if present.
	if f := i.currentFrame(); f != nil {
		if addr < LocalAddrBase {
			if _, ok := f.Locals[addr]; ok {
				f.Locals[addr] = v
				return
			}
		}
	}

	i.globals[addr] = v
}

// GetVar reads a value from an address, considering frame scoping for locals (>= LocalAddrBase)
func (i *Interpreter) GetVar(addr int) (Value, bool) {
	if f := i.currentFrame(); f != nil && addr >= LocalAddrBase {
		v, ok := f.Locals[addr]
		return v, ok
	}

	if f := i.currentFrame(); f != nil {
		if addr < LocalAddrBase {
			if v, ok := f.Locals[addr]; ok {
				return v, ok
			}
		}
	}

	v, ok := i.globals[addr]
	return v, ok
}

// SetVarTyped provides a typed write using lexer.TokenType to choose the underlying ValueKind
func (i *Interpreter) SetVarTyped(addr int, typ lexer.TokenType, raw any) error {
	switch typ {
	case lexer.INT:
		switch x := raw.(type) {
		case int:
			i.SetVar(addr, newInt(int64(x)))
		case int64:
			i.SetVar(addr, newInt(x))
		case float64:
			i.SetVar(addr, newInt(int64(x)))
		case string:
			n, err := strconv.ParseInt(x, 10, 64)
			if err != nil {
				return err
			}
			i.SetVar(addr, newInt(n))
		default:
			return fmt.Errorf("unsupported raw for INT: %T", raw)
		}

	case lexer.FLOAT:
		switch x := raw.(type) {
		case float32:
			i.SetVar(addr, newFloat(float64(x)))
		case float64:
			i.SetVar(addr, newFloat(x))
		case int:
			i.SetVar(addr, newFloat(float64(x)))
		case int64:
			i.SetVar(addr, newFloat(float64(x)))
		case string:
			f, err := strconv.ParseFloat(x, 64)
			if err != nil {
				return err
			}
			i.SetVar(addr, newFloat(f))
		default:
			return fmt.Errorf("unsupported raw for FLOAT: %T", raw)
		}

	case lexer.BOOL:
		switch x := raw.(type) {
		case bool:
			i.SetVar(addr, newBool(x))
		case int:
			i.SetVar(addr, newBool(x != 0))
		case int64:
			i.SetVar(addr, newBool(x != 0))
		case float64:
			i.SetVar(addr, newBool(math.Abs(x) > 0))
		case string:
			switch strings.ToLower(x) {
			case "true", "1":
				i.SetVar(addr, newBool(true))
			case "false", "0":
				i.SetVar(addr, newBool(false))
			default:
				return fmt.Errorf("unsupported bool string: %q", x)
			}
		default:
			return fmt.Errorf("unsupported raw for BOOL: %T", raw)
		}

	case lexer.STRING:
		switch x := raw.(type) {
		case string:
			i.SetVar(addr, newString(x))
		default:
			return fmt.Errorf("unsupported raw for STRING: %T", raw)
		}

	default:
		return fmt.Errorf("unknown type: %v", typ)
	}

	return nil
}

// indexProgram builds label and function indices for jumps and calls
func (i *Interpreter) indexProgram() {
	i.labelIndex = make(map[int]string)
	i.funcIndex = make(map[string]int)
	i.funcEnd = make(map[int]int)
	i.retTypes = make(map[string]lexer.TokenType)

	// record labels and function starts
	for idx, ins := range i.pb {
		if ins.Op == codegen.OpLabel {
			if name, ok := ins.Arg1.(string); ok && name != "" {
				i.labelIndex[idx] = name
				i.funcIndex[name] = idx
			}
		}
	}

	// find function ends (OpEnd after each OpLabel range)
	var currentLabelIdx = -1
	for idx, ins := range i.pb {
		if ins.Op == codegen.OpLabel {
			currentLabelIdx = idx
		} else if ins.Op == codegen.OpEnd && currentLabelIdx != -1 {
			i.funcEnd[currentLabelIdx] = idx - 1
			currentLabelIdx = -1
		}
	}
}

// StageArg stages an argument value for a given position (pos starts at 0)
func (i *Interpreter) StageArg(pos int, v Value) {
	if i.argBuf == nil {
		i.argBuf = make(map[int]Value)
	}
	i.argBuf[pos] = v
}

// ConsumeArg retrieves and removes a staged argument at the given position
func (i *Interpreter) ConsumeArg(pos int) (Value, bool) {
	if i.argBuf == nil {
		return Value{}, false
	}
	v, ok := i.argBuf[pos]
	if ok {
		delete(i.argBuf, pos)
	}
	return v, ok
}

// ClearArgs clears all staged arguments
func (i *Interpreter) ClearArgs() {
	i.argBuf = make(map[int]Value)
}

var (
	ErrNotImplemented   = errors.New("interpreter step function not linked")
	ErrMaxStepsExceeded = errors.New("maximum steps exceeded")
)
