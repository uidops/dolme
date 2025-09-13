package arm64_macos

import (
	"bytes"

	"dolme/pkg/lexer"
	"dolme/pkg/parser/codegen"
	"dolme/pkg/parser/codegen/assembly"
)

type arm64Macos struct {
	pb []codegen.Instruction // three-address code instructions
	cg *codegen.Codegen      // reference to codegen for type lookups

	output string // output file name (for comments)

	text    bytes.Buffer // .text section
	cstring bytes.Buffer // .cstring section
	data    bytes.Buffer // .data section (for float/double constants)
	globl   bytes.Buffer // .globl section (just comments listing globals)

	globalOffsets map[int]int            // addr -> offset within global frame
	funcLocals    map[string]map[int]int // func -> (addr -> offset)
	globalSize    int                    // bytes reserved for globals
	localSizes    map[string]int         // func -> size bytes reserved for locals
	inFunction    bool                   // are we currently emitting inside a function
	currentFunc   string                 // name of current function being emitted

	funcTypes map[string]map[int]lexer.TokenType // Mapping of function names to their local variable types (addr -> type)

	strCounter int // string literal counter

	pbLabels map[int]string                // PB index -> label name
	callArgs map[int][]codegen.Instruction // PB index of OpCall -> list of OpArg instructions
}

// NewArm64Macos creates a new arm64Macos assembly generator instance
func NewArm64Macos(PB []codegen.Instruction, cg *codegen.Codegen, output string) assembly.Assembly {
	return &arm64Macos{
		pb:            PB,
		cg:            cg,
		output:        output,
		globalOffsets: make(map[int]int),
		funcLocals:    make(map[string]map[int]int),
		localSizes:    make(map[string]int),
		pbLabels:      make(map[int]string),
		callArgs:      make(map[int][]codegen.Instruction),
		funcTypes:     make(map[string]map[int]lexer.TokenType),
	}
}

// Generate generates the assembly code from the PB instructions
func (a *arm64Macos) Generate() error {
	// Phase 0: collect addresses used by globals and per-function locals
	a.collectStackLayout()

	// Phase 1: collect labels and call arguments
	a.collectLabelsAndCallArgs()

	// Emit header
	a.addText("\t.text")
	a.addText("\t.globl _main")

	// Emit functions and main
	a.emitMainAndFunctions()

	return nil
}

// GetCode returns the generated assembly code as a string
func (a *arm64Macos) GetCode() string {
	var b bytes.Buffer

	// .text first
	b.Write(a.text.Bytes())

	// emit globals declaration (just list of globals if any as comment)
	if a.globl.Len() > 0 {
		b.WriteString("\n")
		b.Write(a.globl.Bytes())
	}

	// .cstring section for string literals
	if a.cstring.Len() > 0 {
		b.WriteString("\n\t.section\t__TEXT,__cstring\n")
		b.Write(a.cstring.Bytes())
	}

	// .const (read-only data) for float/double constants
	if a.data.Len() > 0 {
		b.WriteString("\n\t.section\t__DATA,__const\n")
		b.Write(a.data.Bytes())
	}

	return b.String()
}
