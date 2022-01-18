package color

import (
	"os"
	"strings"
)

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorGrey   = "\033[90m"
)

func Is256ColorSupported() bool {
	if strings.Contains(os.Getenv("TERM"), "256") || strings.Contains(os.Getenv("COLORTERM"), "256") {
		return true
	}
	return false
}
