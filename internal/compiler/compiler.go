package compiler

import (
	"dolme/pkg/color"
	"dolme/pkg/interpreter"
	"dolme/pkg/lexer"
	"dolme/pkg/parser"
	"dolme/pkg/parser/codegen/assembly"
	arm64_macos "dolme/pkg/parser/codegen/assembly/arm64/macos"
	"fmt"
	"os"

	"github.com/charmbracelet/log"
)

type Compiler struct {
	Help            bool   // Show help message
	Verbose         bool   // Enable verbose output
	ShouldInterpret bool   // Whether to interpret the code
	ShouldCompile   bool   // Whether to compile the code
	NoColor         bool   // Disable colored output
	TargetArch      string // Target architecture for compilation (e.g., "arm64-macos")
	SourceFile      string // Path to the source file
	OutputFile      string // Path to the output file
}

// Compile processes the source file, generates IR code, and either interprets or compiles it based on the options set.
func (opts *Compiler) Compile() error {
	log.Info("Processing file", "file", opts.SourceFile)

	input, err := os.ReadFile(opts.SourceFile)
	if err != nil {
		log.Fatal("Failed to read file", "file", opts.SourceFile, "error", err)
	}

	l := lexer.NewLexer(string(input))
	p := parser.NewParser(l)
	p.Parse()

	syntaxErrors := p.Errors()
	if len(syntaxErrors) > 0 {
		fmt.Println(color.BrightRedText("=== Syntax Errors ==="))
		fmt.Println(syntaxErrors[0])
		return fmt.Errorf("parsing failed with %d errors", len(syntaxErrors))
	}

	semanticErrors := p.GetSemanticErrors()
	if len(semanticErrors) > 0 {
		fmt.Println(color.BrightRedText("=== Semantic Errors ==="))
		fmt.Println(semanticErrors[0])
		return fmt.Errorf("semantic analysis failed with %d errors", len(semanticErrors))
	}

	instructions := p.GetIRCode()

	if opts.Verbose {
		fmt.Println(color.GreenText("\n=== Generated Three-Address Code ==="))
		if len(instructions) == 0 {
			fmt.Println(color.GrayText("No code generated."))
		} else {
			for i, instr := range instructions {
				arg1 := ""
				arg2 := ""
				arg3 := ""

				if instr.Arg1 != nil {
					arg1 = fmt.Sprintf("%v", instr.Arg1)
				}
				if instr.Arg2 != nil {
					arg2 = fmt.Sprintf("%v", instr.Arg2)
				}
				if instr.Arg3 != nil {
					arg3 = fmt.Sprintf("%v", instr.Arg3)
				}

				fmt.Printf("%s: (%s, %s, %s, %s)\n",
					color.CyanText(fmt.Sprintf("%d", i)),
					color.YellowText(string(instr.Op)),
					color.BlueText(arg1),
					color.BlueText(arg2),
					color.BlueText(arg3))
			}
		}
	}

	var arch assembly.Assembly
	if opts.ShouldCompile {
		switch opts.TargetArch {
		case "arm64-macos":
			arch = arm64_macos.NewArm64Macos(instructions, p.GetCG(), opts.OutputFile)
		}

		if err := arch.Generate(); err != nil {
			return fmt.Errorf("assembly generation failed: %w", err)
		}

		if opts.Verbose {
			fmt.Println(color.GreenText("\nGenerated Assembly code"))
			fmt.Println(arch.GetCode())
		}

		if err := arch.Build(); err != nil {
			return fmt.Errorf("Assembly build failed: %w", err)
		}
	}

	if opts.ShouldInterpret {
		intr := interpreter.NewInterpreter(instructions)
		fmt.Println(color.GreenText("\n=== Program Output ==="))
		if err := intr.Run(); err != nil {
			return fmt.Errorf("interpretation failed: %w", err)
		}
	}

	return nil
}
