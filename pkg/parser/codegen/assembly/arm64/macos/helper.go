package arm64_macos

import (
	"dolme/pkg/parser/codegen"
	"strings"
)

// addText adds an instruction to the text section
func (a *arm64Macos) addText(instruction string) {
	a.text.WriteString(instruction + "\n")
}

// addCString adds an instruction to the cstring section
func (a *arm64Macos) addCString(instruction string) {
	a.cstring.WriteString(instruction + "\n")
}

// findFunctionEnd finds the end of a function given the index of its label
func (a *arm64Macos) findFunctionEnd(labelIdx int) int {
	for i := labelIdx + 1; i < len(a.pb); i++ {
		if a.pb[i].Op == codegen.OpEnd {
			return i - 1
		}
	}
	return -1
}

// normalizeImmediate normalizes boolean literals to 1 and 0
func normalizeImmediate(val string) string {
	if val == "true" {
		return "1"
	}
	if val == "false" {
		return "0"
	}
	return val
}

// escapeString escapes special characters in a string
func escapeString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}
