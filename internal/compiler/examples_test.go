package compiler_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"testing"

	"dolme/pkg/interpreter"
	"dolme/pkg/lexer"
	"dolme/pkg/parser"
	"dolme/pkg/parser/codegen"
	arm64_macos "dolme/pkg/parser/codegen/assembly/arm64/macos"
)

func TestExamplesInterpreterAndCompiledOutput(t *testing.T) {
	root := repoRoot(t)
	examplesDir := filepath.Join(root, "example")

	expected := map[string]string{
		"01_sine_series.dolme":                    "0.00159265291648676512\n11\n",
		"02_loop_float_mix.dolme":                 "9\n8\n7\n6\n5\n16.00000000000000000000\n",
		"03_arithmetic_control.dolme":             "25\n4\n12\n11\n10\n9\n100\n",
		"04_functions_globals_returns.dolme":      "14\n5\n14\n",
		"05_boolean_logic.dolme":                  "true\nfalse\n2\n3\n5\n",
		"06_float_math.dolme":                     "1.50000000000000000000\n2.75000000000000000000\n",
		"07_nested_calls.dolme":                   "5\n21\n",
		"08_loop_break_continue.dolme":            "6\n12\n",
		"09_mixed_numeric_arithmetic.dolme":       "10.50000000000000000000\n10.50000000000000000000\n3.50000000000000000000\n-3.50000000000000000000\n24.50000000000000000000\n3.50000000000000000000\n1.50000000000000000000\n3\n1\n",
		"10_numeric_functions_comparisons.dolme":  "8.50000000000000000000\n45\nfalse\ntrue\ntrue\nfalse\n",
		"11_fibonacci_iterative.dolme":            "0\n1\n13\n55\n",
		"12_gcd_euclid.dolme":                     "6\n6\n1\n",
		"13_prime_check.dolme":                    "0\n1\n1\n0\n",
		"14_factorial_iterative.dolme":            "1\n1\n120\n5040\n",
		"15_recursive_fibonacci.dolme":            "8\n21\n",
		"16_binary_exponentiation.dolme":          "1024\n243\n1\n",
		"17_newton_sqrt.dolme":                    "1.41421356237309492343\n2.00000000000000000000\n",
		"18_collatz_steps.dolme":                  "8\n111\n",
		"19_russian_peasant_multiplication.dolme": "1974\n0\n169\n",
	}

	examplePaths, err := filepath.Glob(filepath.Join(examplesDir, "*.dolme"))
	if err != nil {
		t.Fatalf("glob examples: %v", err)
	}

	sort.Strings(examplePaths)

	seen := make(map[string]bool, len(examplePaths))
	for _, examplePath := range examplePaths {
		name := filepath.Base(examplePath)
		want, ok := expected[name]
		if !ok {
			t.Fatalf("missing expected output for example %s", name)
		}

		seen[name] = true

		t.Run(name, func(t *testing.T) {
			pb, cg := parseExample(t, examplePath)

			var interpreted bytes.Buffer
			intr := interpreter.NewInterpreter(pb, interpreter.WithWriter(&interpreted))
			if err := intr.Run(); err != nil {
				t.Fatalf("interpreter run failed: %v", err)
			}

			if got := interpreted.String(); got != want {
				t.Fatalf("interpreter output mismatch\nwant:\n%q\ngot:\n%q", want, got)
			}

			if runtime.GOOS != "darwin" || runtime.GOARCH != "arm64" {
				t.Skip("compiled arm64-macos output check requires darwin/arm64")
			}

			exePath := filepath.Join(t.TempDir(), "example-bin")
			asm := arm64_macos.NewArm64Macos(pb, cg, exePath)
			if err := asm.Generate(); err != nil {
				t.Fatalf("assembly generation failed: %v", err)
			}

			if err := asm.Build(); err != nil {
				t.Fatalf("assembly build failed: %v", err)
			}

			out, err := exec.Command(exePath).CombinedOutput()
			if err != nil {
				t.Fatalf("compiled example failed: %v\noutput:\n%s", err, out)
			}

			if got := string(out); got != want {
				t.Fatalf("compiled output mismatch\nwant:\n%q\ngot:\n%q", want, got)
			}
		})
	}

	for name := range expected {
		if !seen[name] {
			t.Fatalf("expected output listed for missing example %s", name)
		}
	}
}

func parseExample(t *testing.T, path string) ([]codegen.Instruction, *codegen.Codegen) {
	t.Helper()

	input, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read example: %v", err)
	}

	l := lexer.NewLexer(string(input))
	p := parser.NewParser(l)
	p.Parse()

	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("syntax errors: %v", errs)
	}

	if errs := p.GetSemanticErrors(); len(errs) > 0 {
		t.Fatalf("semantic errors: %v", errs)
	}

	return p.GetIRCode(), p.GetCG()
}

func repoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot determine test file path")
	}

	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatal(fmt.Errorf("cannot locate repo root: %w", err))
	}

	return root
}
