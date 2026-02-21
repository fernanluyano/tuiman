package ui

import "github.com/charmbracelet/lipgloss"

var (
	colorOrange = lipgloss.Color("#FF6C37")
	colorYellow = lipgloss.Color("#FFD700")
	colorDimmed = lipgloss.Color("#555555")
	colorGreen  = lipgloss.Color("#23D18B")
	colorBlue   = lipgloss.Color("#3B9EFF")
	colorRed    = lipgloss.Color("#FF453A")
	colorAmber  = lipgloss.Color("#FFB547")

	methodStyles = map[string]lipgloss.Style{
		"GET":    lipgloss.NewStyle().Foreground(colorGreen).Bold(true),
		"POST":   lipgloss.NewStyle().Foreground(colorBlue).Bold(true),
		"PUT":    lipgloss.NewStyle().Foreground(colorAmber).Bold(true),
		"PATCH":  lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8C37")).Bold(true),
		"DELETE": lipgloss.NewStyle().Foreground(colorRed).Bold(true),
	}

	collectionStyle = lipgloss.NewStyle().Foreground(colorYellow).Bold(true)
	cursorStyle     = lipgloss.NewStyle().Foreground(colorOrange).Bold(true)
	footerKeyStyle  = lipgloss.NewStyle().Foreground(colorYellow)
	footerDescStyle = lipgloss.NewStyle().Foreground(colorDimmed)
	helpTitleStyle  = lipgloss.NewStyle().Foreground(colorYellow).Bold(true)
	helpKeyStyle    = lipgloss.NewStyle().Foreground(colorGreen)
)

func paneStyle(active bool) lipgloss.Style {
	s := lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
	if active {
		return s.BorderForeground(colorOrange)
	}
	return s.BorderForeground(colorDimmed)
}
