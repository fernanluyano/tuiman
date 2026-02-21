package ui

import tea "github.com/charmbracelet/bubbletea"

// Model holds all application state. Implements tea.Model.
type Model struct {
	width, height int
	focused       int  // 0=sidebar, 1=request, 2=response
	showHelp      bool

	// sidebar
	collections []collection
	expanded    map[string]bool
	cursor      int
	items       []sidebarItem
}

// New creates the initial application model with mocked data.
func New() Model {
	m := Model{
		collections: mockCollections,
		expanded:    make(map[string]bool),
	}
	for _, c := range mockCollections {
		m.expanded[c.name] = true
	}
	m.items = buildItems(m.collections, m.expanded)
	return m
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if m.showHelp {
			switch msg.String() {
			case "q", "?", "esc":
				m.showHelp = false
			}
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "?":
			m.showHelp = true
		case "tab":
			m.focused = (m.focused + 1) % 3
		case "shift+tab":
			m.focused = (m.focused + 2) % 3
		case "S":
			m.focused = 0
		case "R":
			m.focused = 1
		case "P":
			m.focused = 2
		case "j", "down":
			if m.focused == 0 && m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.focused == 0 && m.cursor > 0 {
				m.cursor--
			}
		case " ":
			if m.focused == 0 && len(m.items) > 0 {
				item := m.items[m.cursor]
				if item.kind == itemCollection {
					m.expanded[item.colName] = !m.expanded[item.colName]
					m.items = buildItems(m.collections, m.expanded)
					if m.cursor >= len(m.items) {
						m.cursor = len(m.items) - 1
					}
				}
			}
		}
	}

	return m, nil
}
