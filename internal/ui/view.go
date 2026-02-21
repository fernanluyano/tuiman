package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.width == 0 {
		return ""
	}
	if m.showHelp {
		return m.renderHelp()
	}
	return m.renderMain()
}

func (m Model) renderMain() string {
	const sidebarOuterW = 30
	const footerH = 1

	mainH := m.height - footerH

	// Inner dimensions (border takes 1 char on each side).
	sidebarInnerW := sidebarOuterW - 2
	sidebarInnerH := mainH - 2

	rightOuterW := m.width - sidebarOuterW
	rightInnerW := rightOuterW - 2

	reqOuterH := mainH / 2
	respOuterH := mainH - reqOuterH
	reqInnerH := reqOuterH - 2
	respInnerH := respOuterH - 2

	sidebarBox := paneStyle(m.focused == 0).
		Width(sidebarInnerW).
		Height(sidebarInnerH).
		Render(m.renderSidebar(sidebarInnerW))

	requestBox := paneStyle(m.focused == 1).
		Width(rightInnerW).
		Height(reqInnerH).
		Render(paneTitle(" Request (R) ", m.focused == 1))

	responseBox := paneStyle(m.focused == 2).
		Width(rightInnerW).
		Height(respInnerH).
		Render(paneTitle(" Response (P) ", m.focused == 2))

	rightPanel := lipgloss.JoinVertical(lipgloss.Left, requestBox, responseBox)
	mainArea := lipgloss.JoinHorizontal(lipgloss.Top, sidebarBox, rightPanel)

	return lipgloss.JoinVertical(lipgloss.Left, mainArea, m.renderFooter())
}

func paneTitle(title string, active bool) string {
	if active {
		return lipgloss.NewStyle().Foreground(colorOrange).Render(title)
	}
	return lipgloss.NewStyle().Foreground(colorDimmed).Render(title)
}

func (m Model) renderSidebar(innerW int) string {
	var lines []string

	header := " Collections (S)"
	if m.focused == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(colorOrange).Bold(true).Render(header))
	} else {
		lines = append(lines, lipgloss.NewStyle().Foreground(colorDimmed).Render(header))
	}
	lines = append(lines, "")

	for i, item := range m.items {
		selected := i == m.cursor && m.focused == 0

		var line string
		switch item.kind {
		case itemCollection:
			chevron := "▸"
			if m.expanded[item.colName] {
				chevron = "▾"
			}
			label := chevron + " " + item.colName
			if selected {
				line = cursorStyle.Render("> ") + collectionStyle.Render(item.colName)
			} else {
				line = collectionStyle.Render(label)
			}

		case itemRequest:
			st, ok := methodStyles[item.req.method]
			if !ok {
				st = lipgloss.NewStyle()
			}
			method := st.Render(fmt.Sprintf("%-6s", item.req.method))
			if selected {
				line = cursorStyle.Render(">") + "   " + method + " " + item.req.name
			} else {
				line = "    " + method + " " + item.req.name
			}
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderFooter() string {
	hints := []struct{ key, desc string }{
		{"q", "quit"},
		{"?", "help"},
		{"tab", "next pane"},
		{"S/R/P", "jump"},
		{"↑↓ jk", "navigate"},
		{"space", "expand"},
	}

	var parts []string
	for _, h := range hints {
		k := footerKeyStyle.Render(h.key)
		d := footerDescStyle.Render(" " + h.desc)
		parts = append(parts, "  "+k+d)
	}

	return footerDescStyle.Render(strings.Join(parts, ""))
}

func (m Model) renderHelp() string {
	type row struct{ key, desc string }
	sections := []struct {
		title string
		rows  []row
	}{
		{"Navigation", []row{
			{"j / ↓", "move down"},
			{"k / ↑", "move up"},
		}},
		{"Collections", []row{
			{"space", "expand / collapse"},
			{"enter", "open request"},
		}},
		{"Pane Focus", []row{
			{"tab", "next pane"},
			{"shift+tab", "previous pane"},
			{"S", "jump to sidebar"},
			{"R", "jump to request"},
			{"P", "jump to response"},
		}},
		{"Global", []row{
			{"?", "toggle help"},
			{"q", "quit"},
		}},
	}

	var lines []string
	lines = append(lines, helpTitleStyle.Render("  Keybindings  "))
	lines = append(lines, "")

	for _, sec := range sections {
		lines = append(lines, "  "+helpTitleStyle.Render(sec.title))
		for _, r := range sec.rows {
			k := helpKeyStyle.Render(fmt.Sprintf("   %-16s", r.key))
			lines = append(lines, k+r.desc)
		}
		lines = append(lines, "")
	}

	lines = append(lines, footerDescStyle.Render("  press ? or q to close"))

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorYellow).
		Padding(0, 2).
		Render(strings.Join(lines, "\n"))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}
