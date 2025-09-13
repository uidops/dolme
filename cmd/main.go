package main

import (
	"dolme/internal/compiler"
	"dolme/internal/logger"
	"dolme/pkg/color"
	"flag"
	"fmt"
	"os"

	"github.com/charmbracelet/log"
)

// Main entry point for the Dolme compiler.
func main() {
	options := compiler.Compiler{}

	flag.BoolVar(&options.Help, "h", false, "Show help")
	flag.BoolVar(&options.Verbose, "v", false, "Verbose mode")
	flag.BoolVar(&options.ShouldInterpret, "r", false, "Run with interpreter")
	flag.BoolVar(&options.ShouldCompile, "c", false, "Compile to binary")
	flag.BoolVar(&options.NoColor, "n", false, "No color")
	flag.StringVar(&options.TargetArch, "a", "arm64-macos", "Target architecture (e.g., arm64-macos, x86_64-linux)")
	flag.StringVar(&options.OutputFile, "o", "a.out", "Output binary name")

	flag.Parse()
	args := flag.Args()

	logger.Init(options.Verbose, options.NoColor)
	if options.Help {
		fmt.Printf("Usage: %s [options] <file>\n", os.Args[0])
		fmt.Println("Options:")
		flag.PrintDefaults()
		return
	}

	if options.NoColor {
		color.EnableColor(false)
	}

	if len(args) == 0 {
		log.Fatal("No input file provided", "help", fmt.Sprintf("%s -h", os.Args[0]))
	}

	options.SourceFile = args[0]

	err := options.Compile()
	if err != nil {
		log.Fatal("Compilation failed", "error", err)
	}
}
