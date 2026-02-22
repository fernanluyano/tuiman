package ui

import "github.com/charmbracelet/lipgloss"

func (t Theme) paneStyle(active bool) lipgloss.Style {
	s := lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
	if active {
		return s.BorderForeground(t.Accent)
	}
	return s.BorderForeground(t.Dimmed)
}

func (t Theme) paneTitle(title string, active bool) string {
	if active {
		return lipgloss.NewStyle().Foreground(t.Accent).Render(title)
	}
	return lipgloss.NewStyle().Foreground(t.Dimmed).Render(title)
}

func (t Theme) dim() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.Dimmed)
}

func (t Theme) accent() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.Accent)
}

func (t Theme) highlight() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.Highlight)
}

func (t Theme) text() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.Text)
}

func (t Theme) textMuted() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.TextMuted)
}

func (t Theme) errStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.Error)
}

func (t Theme) footerKeyStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.Highlight)
}

func (t Theme) footerDescStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.Dimmed)
}

func (t Theme) helpTitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.Highlight).Bold(true)
}

func (t Theme) helpKeyStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.MethodGET)
}

// activeTabStyle is used for the currently selected request tab.
func (t Theme) activeTabStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.Accent).Bold(true).Underline(true)
}

// successStyle is used for safe/positive actions (e.g. "no, don't quit").
func (t Theme) successStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.MethodGET).Bold(true)
}

// overlayStyle is the base border style for floating overlays (pickers).
func (t Theme) overlayStyle() lipgloss.Style {
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(t.Accent)
}

// helpOverlayStyle is the base border style for help/command screens.
func (t Theme) helpOverlayStyle() lipgloss.Style {
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(t.Highlight)
}

// keyHint renders "(key)" with the parentheses dimmed and the key in the highlight color.
func (t Theme) keyHint(key string) string {
	dim := lipgloss.NewStyle().Foreground(t.Dimmed)
	hi := lipgloss.NewStyle().Foreground(t.Highlight)
	return dim.Render("(") + hi.Render(key) + dim.Render(")")
}
