package logger

import (
	"io"
	"os"
	"time"

	"github.com/charmbracelet/log"
	"github.com/muesli/termenv"
)

// Init initializes the logger
func Init(debug, noColor bool) {
	log.SetDefault(log.NewWithOptions(io.MultiWriter(os.Stderr),
		log.Options{
			ReportCaller:    true,
			ReportTimestamp: false, // we don't need timestamps, as we have them in the prompt
			TimeFormat:      time.RFC3339,
			Prefix:          "DOLME",
		}))

	if !debug {
		log.SetLevel(log.ErrorLevel | log.WarnLevel)
	}

	log.SetColorProfile(termenv.ANSI256)
	if noColor {
		log.SetColorProfile(termenv.Ascii)
	}
}
