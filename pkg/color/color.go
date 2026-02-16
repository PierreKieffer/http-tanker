package color

import "github.com/charmbracelet/lipgloss"

var (
	Red    = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	Green  = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	Yellow = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	Blue   = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	Purple = lipgloss.NewStyle().Foreground(lipgloss.Color("13"))
	Cyan   = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	White  = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	Grey   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	Bold   = lipgloss.NewStyle().Bold(true)
)

func StatusCodeStyle(code int) lipgloss.Style {
	switch {
	case code >= 200 && code < 300:
		return Green
	case code >= 300 && code < 400:
		return Yellow
	case code >= 400:
		return Red
	default:
		return White
	}
}

func MethodStyle(method string) lipgloss.Style {
	switch method {
	case "GET":
		return Cyan
	case "POST":
		return Green
	case "PUT":
		return Yellow
	case "DELETE":
		return Red
	default:
		return White
	}
}
