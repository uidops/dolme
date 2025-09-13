package color

import (
	"fmt"
	"os"
	"strings"
)

const (
	Reset = "\033[0m"
	Bold  = "\033[1m"

	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"
	Gray    = "\033[90m"

	BrightRed     = "\033[91m"
	BrightGreen   = "\033[92m"
	BrightYellow  = "\033[93m"
	BrightBlue    = "\033[94m"
	BrightMagenta = "\033[95m"
	BrightCyan    = "\033[96m"
	BrightWhite   = "\033[97m"
)

var colorEnabled = true

func init() {
	if os.Getenv("NO_COLOR") != "" || !isTerminal() {
		colorEnabled = false
	}
}

func isTerminal() bool {
	term := os.Getenv("TERM")
	return term != "" && term != "dumb"
}

func EnableColor(enable bool) {
	colorEnabled = enable
}

func IsColorEnabled() bool {
	return colorEnabled
}

func Colorize(color, text string) string {
	if !colorEnabled {
		return text
	}
	return color + text + Reset
}

func RedText(text string) string {
	return Colorize(Red, text)
}

func BrightRedText(text string) string {
	return Colorize(BrightRed, text)
}

func GreenText(text string) string {
	return Colorize(Green, text)
}

func YellowText(text string) string {
	return Colorize(Yellow, text)
}

func BlueText(text string) string {
	return Colorize(Blue, text)
}

func CyanText(text string) string {
	return Colorize(Cyan, text)
}

func GrayText(text string) string {
	return Colorize(Gray, text)
}

func BoldText(text string) string {
	return Colorize(Bold, text)
}

func Error(message string) string {
	if !colorEnabled {
		return message
	}
	return BrightRedText("Error: ") + message
}

func Warning(message string) string {
	if !colorEnabled {
		return message
	}
	return YellowText("Warning: ") + message
}

func Info(message string) string {
	if !colorEnabled {
		return message
	}
	return BlueText("Info: ") + message
}

func Success(message string) string {
	if !colorEnabled {
		return message
	}
	return GreenText("Success: ") + message
}

func Highlight(text, highlight string) string {
	if !colorEnabled {
		return text
	}
	return strings.ReplaceAll(text, highlight, YellowText(highlight))
}

func Position(line, col int) string {
	pos := fmt.Sprintf("%d:%d", line, col)
	if !colorEnabled {
		return pos
	}
	return CyanText(pos)
}

func Code(code string) string {
	if !colorEnabled {
		return code
	}
	return GrayText(code)
}

func ErrorWithPosition(line, col int, message, context string) string {
	if !colorEnabled {
		return fmt.Sprintf("Error at %d:%d: %s\n%s", line, col, message, context)
	}

	return fmt.Sprintf("%s at %s: %s\n%s",
		BrightRedText(BoldText("Error")),
		Position(line, col),
		message,
		GrayText(context))
}
