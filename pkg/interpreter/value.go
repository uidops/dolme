package interpreter

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type ValueKind int

const (
	KindUnknown ValueKind = iota
	KindInt
	KindFloat
	KindBool
	KindString
)

// Value represents a dynamically-typed value in the interpreter.
type Value struct {
	Kind  ValueKind
	I64   int64
	F64   float64
	Bool  bool
	Str   string
	Valid bool
}

// String renders the value as a string.
func (v Value) String() string {
	switch v.Kind {
	case KindInt:
		return fmt.Sprintf("%d", v.I64)
	case KindFloat:
		return fmt.Sprintf("%g", v.F64)
	case KindBool:
		if v.Bool {
			return "true"
		}
		return "false"
	case KindString:
		return v.Str
	default:
		return "<nil>"
	}
}

// AsFloat64 converts the value to float64 if possible.
func (v Value) AsFloat64() (float64, error) {
	switch v.Kind {
	case KindFloat:
		return v.F64, nil
	case KindInt:
		return float64(v.I64), nil
	case KindBool:
		if v.Bool {
			return 1.0, nil
		}
		return 0.0, nil
	default:
		return 0, fmt.Errorf("cannot convert %v to float", v.Kind)
	}
}

// AsInt64 converts the value to int64 if possible.
func (v Value) AsInt64() (int64, error) {
	switch v.Kind {
	case KindInt:
		return v.I64, nil
	case KindFloat:
		return int64(v.F64), nil
	case KindBool:
		if v.Bool {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("cannot convert %v to int", v.Kind)
	}
}

// AsBool converts the value to bool if possible.
func (v Value) AsBool() (bool, error) {
	switch v.Kind {
	case KindBool:
		return v.Bool, nil
	case KindInt:
		return v.I64 != 0, nil
	case KindFloat:
		return math.Abs(v.F64) > 0, nil
	default:
		return false, fmt.Errorf("cannot convert %v to bool", v.Kind)
	}
}

// NewInt creates a new integer Value.
func newInt(i int64) Value {
	return Value{Kind: KindInt, I64: i, Valid: true}
}

// NewFloat creates a new float Value.
func newFloat(f float64) Value {
	return Value{Kind: KindFloat, F64: f, Valid: true}
}

// NewBool creates a new boolean Value.
func newBool(b bool) Value {
	return Value{Kind: KindBool, Bool: b, Valid: true}
}

// NewString creates a new string Value.
func newString(s string) Value {
	return Value{Kind: KindString, Str: s, Valid: true}
}

// ParseImmediate parses a codegen immediate like "#1", "#3.14", "#true", or a quoted string (with leading '#').
func parseImmediate(imm string) (Value, error) {
	if !strings.HasPrefix(imm, "#") {
		return Value{}, fmt.Errorf("immediate must start with '#': %q", imm)
	}
	body := imm[1:]

	// booleans
	if body == "true" {
		return newBool(true), nil
	}
	if body == "false" {
		return newBool(false), nil
	}

	// quoted strings (already include quotes as produced by lexer)
	if len(body) >= 2 && body[0] == '"' && body[len(body)-1] == '"' {
		unquoted := body[1 : len(body)-1]
		return newString(unquoted), nil
	}

	// try integer
	if i, err := strconv.ParseInt(body, 10, 64); err == nil {
		return newInt(i), nil
	}

	// try float
	if f, err := strconv.ParseFloat(body, 64); err == nil {
		return newFloat(f), nil
	}

	return Value{}, fmt.Errorf("unsupported immediate: %q", imm)
}
