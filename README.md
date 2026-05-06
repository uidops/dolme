# Dolme Programming Language

<p align="center">
<img src="https://github.com/uidops/uidops/blob/main/Dolmas+colour.jpg?raw=true">
</p>

Dolme is a small programming language and compiler written in Go for a compiler design course at Azarbaijan Shahid Madani University. The project is built around an LL(1) parser, syntax-directed translation, a three-address-code IR, an interpreter, and an ARM64 macOS assembly backend.

The goal is educational: implement the core pieces of a compiler in a readable codebase, then validate interpreter and compiled behavior against the same example programs.

## Features

- `int`, `float`, and `bool` values
- Variable declarations with `let`
- Arithmetic, comparison, and boolean operators
- `if` / `else`
- `while`, `break`, and `continue`
- Functions with typed parameters and return values
- Recursive and nested function calls
- Interpreter mode for running IR directly
- Native compilation to ARM64 macOS assembly

## Example

```js
func fib(n: int): int {
    if (n <= 1) {
        return n;
    }

    return fib(n - 1) + fib(n - 2);
}

let result: int = fib(8);
print(result);
```

## Build

```bash
make build
```

or directly:

```bash
go build -o bin/dolme cmd/main.go
```

## Usage

Show help:

```bash
bin/dolme -h
```

Run with the interpreter:

```bash
bin/dolme -r example/15_recursive_fibonacci.dolme
```

Compile to an ARM64 macOS binary:

```bash
bin/dolme -c -a arm64-macos -o out example/15_recursive_fibonacci.dolme
./out
```

Print generated IR and assembly while compiling:

```bash
bin/dolme -c -a arm64-macos -v example/15_recursive_fibonacci.dolme
```

## Examples

The `example/` directory contains small programs used as compiler regression cases, including:

- Sine approximation with Taylor series
- Mixed integer/float arithmetic and comparisons
- Boolean logic
- Nested function calls
- Fibonacci, factorial, GCD, primality checking
- Binary exponentiation
- Newton square root approximation
- Collatz steps
- Russian peasant multiplication

## Testing

Run the full Go test suite:

```bash
go test -mod=readonly ./...
```

The integration test in `internal/compiler/examples_test.go` parses every `example/*.dolme` file, runs it through the interpreter, compiles it through the ARM64 macOS backend, runs the native binary, and compares both outputs with the expected output. The compiled-binary check runs only on `darwin/arm64`.

## Project Layout

- `cmd/`: command-line entry point
- `internal/compiler/`: compile/run orchestration and example integration tests
- `pkg/lexer/`: lexer and token definitions
- `pkg/parser/`: LL(1) parser and parsing table
- `pkg/parser/codegen/`: semantic actions and three-address-code IR generation
- `pkg/parser/codegen/assembly/arm64/macos/`: ARM64 macOS assembly backend
- `pkg/interpreter/`: IR interpreter
- `example/`: Dolme programs used for manual testing and regression tests

## Grammar

The grammar and LL(1) parsing table are documented in `dolme.md`.

The original course problem statement is kept separately as `dolme.pdf` when available.

## Target Support

Native compilation currently targets ARM64 macOS only (`-a arm64-macos`). Interpreter mode is portable anywhere Go can run.
