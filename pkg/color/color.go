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

func StatusCodeColor(code int) string {
	switch {
	case code >= 200 && code < 300:
		return ColorGreen
	case code >= 300 && code < 400:
		return ColorYellow
	case code >= 400:
		return ColorRed
	default:
		return ColorWhite
	}
}

func MethodColor(method string) string {
	switch method {
	case "GET":
		return ColorCyan
	case "POST":
		return ColorGreen
	case "PUT":
		return ColorYellow
	case "DELETE":
		return ColorRed
	default:
		return ColorWhite
	}
}

func Is256ColorSupported() bool {
	if strings.Contains(os.Getenv("TERM"), "256") || strings.Contains(os.Getenv("COLORTERM"), "256") {
		return true
	}
	return false
}
