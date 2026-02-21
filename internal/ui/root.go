package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Model holds all application state. Implements tea.Model.
type Model struct {
	width, height int
	focused       int  // 0=request, 1=response
	showHelp      bool

	activeRequest *request
	folders       []folder
	requestTab    int    // 0=Params, 1=Auth, 2=Headers, 3=Body
	urlInput      string // editable URL (in-memory only)
	editingURL    bool
	methodInput   string // selected HTTP method (in-memory only)

	// method picker
	showMethodPicker bool
	methodCursor     int

	// folder picker
	showFolderPicker bool
	fpLevel          int    // 0=folder list, 1=request list in selected folder
	fpFolderIdx      int    // index into m.folders (active when level=1)
	fpQuery          string
	fpCursor         int
	fpFolderShown    []int // indices into m.folders matching fpQuery
	fpReqShown       []int // indices into m.folders[fpFolderIdx].requests matching fpQuery
	fpInsert         bool   // insert mode (typing filters the list)
	fpAdding         bool
	fpAddKind        string // "folder" or "request"
	fpAddInput       string
	fpConfirmDelete  bool
}

// New creates the initial application model with mocked data.
func New() Model {
	folders := mockFolders
	return Model{
		folders:       folders,
		fpFolderShown: filterFolders(folders, ""),
		methodInput:   "GET",
	}
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

		if m.editingURL {
			return m.updateURLInput(msg)
		}

		if m.showMethodPicker {
			return m.updateMethodPicker(msg)
		}

		if m.showFolderPicker {
			return m.updateFolderPicker(msg), nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "?":
			m.showHelp = true

		// Pane navigation — tab or vim j/k
		case "tab", "shift+tab", "j", "k":
			m.focused = (m.focused + 1) % 2

		// Folder picker
		case "f":
			m.showFolderPicker = true
			m.fpLevel = 0
			m.fpQuery = ""
			m.fpCursor = 0
			m.fpFolderShown = filterFolders(m.folders, "")

		// Send request (placeholder)
		case "s":
			// TODO: trigger HTTP request

		// Method picker
		case "m":
			if m.focused == 0 {
				m.showMethodPicker = true
				for i, meth := range httpMethods {
					if meth == m.methodInput {
						m.methodCursor = i
						break
					}
				}
			}

		// URL editing
		case "e":
			if m.focused == 0 {
				m.editingURL = true
			}

		// Tab cycling in request pane — [/] or arrow keys
		case "[", "left":
			if m.focused == 0 && m.requestTab > 0 {
				m.requestTab--
			}
		case "]", "right":
			if m.focused == 0 && m.requestTab < 3 {
				m.requestTab++
			}
		// Direct tab jump — p/a/h/b
		case "p":
			if m.focused == 0 {
				m.requestTab = 0
			}
		case "a":
			if m.focused == 0 {
				m.requestTab = 1
			}
		case "h":
			if m.focused == 0 {
				m.requestTab = 2
			}
		case "b":
			if m.focused == 0 {
				m.requestTab = 3
			}
		}
	}

	return m, nil
}

func (m Model) updateFolderPicker(msg tea.KeyMsg) Model {
	// Confirm-delete mode: wait for y / esc
	if m.fpConfirmDelete {
		switch msg.String() {
		case "y":
			m = m.performDelete()
		case "n", "esc":
			m.fpConfirmDelete = false
		}
		return m
	}

	// Adding mode: input goes to fpAddInput
	if m.fpAdding {
		switch msg.String() {
		case "esc":
			m.fpAdding = false
			m.fpAddInput = ""
		case "enter":
			m = m.commitFolderAdd()
		case "backspace":
			runes := []rune(m.fpAddInput)
			if len(runes) > 0 {
				m.fpAddInput = string(runes[:len(runes)-1])
			}
		default:
			if len([]rune(msg.String())) == 1 {
				m.fpAddInput += msg.String()
			}
		}
		return m
	}

	// Insert mode: typing filters the list
	if m.fpInsert {
		switch msg.String() {
		case "esc":
			m.fpInsert = false
		case "enter":
			m.fpInsert = false
			m = m.performEnter()
		case "backspace":
			if len(m.fpQuery) > 0 {
				runes := []rune(m.fpQuery)
				m.fpQuery = string(runes[:len(runes)-1])
				if m.fpLevel == 0 {
					m.fpFolderShown = filterFolders(m.folders, m.fpQuery)
					if m.fpCursor >= len(m.fpFolderShown) {
						m.fpCursor = len(m.fpFolderShown) - 1
					}
					if m.fpCursor < 0 {
						m.fpCursor = 0
					}
				} else {
					m.fpReqShown = filterRequests(m.folders[m.fpFolderIdx].requests, m.fpQuery)
					if m.fpCursor >= len(m.fpReqShown) {
						m.fpCursor = len(m.fpReqShown) - 1
					}
					if m.fpCursor < 0 {
						m.fpCursor = 0
					}
				}
			}
		default:
			if len([]rune(msg.String())) == 1 {
				m.fpQuery += msg.String()
				if m.fpLevel == 0 {
					m.fpFolderShown = filterFolders(m.folders, m.fpQuery)
				} else {
					m.fpReqShown = filterRequests(m.folders[m.fpFolderIdx].requests, m.fpQuery)
				}
				m.fpCursor = 0
			}
		}
		return m
	}

	// Normal mode
	if m.fpLevel == 0 {
		switch msg.String() {
		case "esc":
			m.showFolderPicker = false
			m.fpQuery = ""
			m.fpInsert = false
		case "enter":
			m = m.performEnter()
		case "j", "down", "ctrl+j":
			if m.fpCursor < len(m.fpFolderShown)-1 {
				m.fpCursor++
			}
		case "k", "up", "ctrl+k":
			if m.fpCursor > 0 {
				m.fpCursor--
			}
		case "i":
			m.fpInsert = true
		case "n":
			m.fpAdding = true
			m.fpAddKind = "folder"
			m.fpAddInput = ""
		case "d":
			if len(m.fpFolderShown) > 0 {
				m.fpConfirmDelete = true
			}
		}
	} else {
		// Level 1: request list within the selected folder
		switch msg.String() {
		case "esc":
			m.fpLevel = 0
			m.fpQuery = ""
			m.fpCursor = 0
			m.fpInsert = false
			m.fpFolderShown = filterFolders(m.folders, "")
		case "enter":
			m = m.performEnter()
		case "j", "down", "ctrl+j":
			if m.fpCursor < len(m.fpReqShown)-1 {
				m.fpCursor++
			}
		case "k", "up", "ctrl+k":
			if m.fpCursor > 0 {
				m.fpCursor--
			}
		case "i":
			m.fpInsert = true
		case "n":
			m.fpAdding = true
			m.fpAddKind = "request"
			m.fpAddInput = ""
		case "d":
			if len(m.fpReqShown) > 0 {
				m.fpConfirmDelete = true
			}
		}
	}
	return m
}

func (m Model) commitFolderAdd() Model {
	if m.fpAddInput == "" {
		m.fpAdding = false
		return m
	}
	switch m.fpAddKind {
	case "folder":
		m.folders = append(m.folders, folder{name: m.fpAddInput})
		m.fpQuery = ""
		m.fpFolderShown = filterFolders(m.folders, "")
		m.fpCursor = len(m.fpFolderShown) - 1
	case "request":
		newReq := request{method: "GET", name: m.fpAddInput, auth: requestAuth{kind: authNone}}
		f := m.folders[m.fpFolderIdx]
		f.requests = append(f.requests, newReq)
		m.folders[m.fpFolderIdx] = f
		m.fpQuery = ""
		m.fpReqShown = filterRequests(m.folders[m.fpFolderIdx].requests, "")
		m.fpCursor = len(m.fpReqShown) - 1
	}
	m.fpAdding = false
	m.fpAddInput = ""
	return m
}

func (m Model) performEnter() Model {
	if m.fpLevel == 0 {
		if len(m.fpFolderShown) > 0 {
			m.fpFolderIdx = m.fpFolderShown[m.fpCursor]
			m.fpLevel = 1
			m.fpQuery = ""
			m.fpCursor = 0
			m.fpInsert = false
			m.fpReqShown = filterRequests(m.folders[m.fpFolderIdx].requests, "")
		}
	} else {
		if len(m.fpReqShown) > 0 {
			reqIdx := m.fpReqShown[m.fpCursor]
			req := m.folders[m.fpFolderIdx].requests[reqIdx]
			m.activeRequest = &req
			m.urlInput = req.url
			m.methodInput = req.method
			m.showFolderPicker = false
			m.fpQuery = ""
			m.fpInsert = false
		}
	}
	return m
}

func (m Model) performDelete() Model {
	m.fpConfirmDelete = false
	if m.fpLevel == 0 {
		if len(m.fpFolderShown) == 0 {
			return m
		}
		idx := m.fpFolderShown[m.fpCursor]
		m.folders = append(m.folders[:idx], m.folders[idx+1:]...)
		m.fpFolderShown = filterFolders(m.folders, m.fpQuery)
		if m.fpCursor >= len(m.fpFolderShown) {
			m.fpCursor = len(m.fpFolderShown) - 1
		}
		if m.fpCursor < 0 {
			m.fpCursor = 0
		}
	} else {
		if len(m.fpReqShown) == 0 {
			return m
		}
		reqIdx := m.fpReqShown[m.fpCursor]
		f := m.folders[m.fpFolderIdx]
		f.requests = append(f.requests[:reqIdx], f.requests[reqIdx+1:]...)
		m.folders[m.fpFolderIdx] = f
		m.fpReqShown = filterRequests(m.folders[m.fpFolderIdx].requests, m.fpQuery)
		if m.fpCursor >= len(m.fpReqShown) {
			m.fpCursor = len(m.fpReqShown) - 1
		}
		if m.fpCursor < 0 {
			m.fpCursor = 0
		}
	}
	return m
}

func (m Model) updateMethodPicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.showMethodPicker = false
	case "enter":
		m.methodInput = httpMethods[m.methodCursor]
		m.showMethodPicker = false
	case "j", "down":
		if m.methodCursor < len(httpMethods)-1 {
			m.methodCursor++
		}
	case "k", "up":
		if m.methodCursor > 0 {
			m.methodCursor--
		}
	}
	return m, nil
}

func (m Model) updateURLInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "enter":
		m.editingURL = false
	case "backspace":
		runes := []rune(m.urlInput)
		if len(runes) > 0 {
			m.urlInput = string(runes[:len(runes)-1])
		}
	default:
		if len([]rune(msg.String())) == 1 {
			m.urlInput += msg.String()
		}
	}
	return m, nil
}

func filterFolders(folders []folder, query string) []int {
	q := strings.ToLower(query)
	var out []int
	for i, f := range folders {
		if q == "" || fuzzyMatch(strings.ToLower(f.name), q) {
			out = append(out, i)
		}
	}
	return out
}

func filterRequests(reqs []request, query string) []int {
	q := strings.ToLower(query)
	var out []int
	for i, r := range reqs {
		display := strings.ToLower(r.method + " " + r.name)
		if q == "" || fuzzyMatch(display, q) {
			out = append(out, i)
		}
	}
	return out
}

func fuzzyMatch(s, pattern string) bool {
	pi := 0
	for si := 0; si < len(s) && pi < len(pattern); si++ {
		if s[si] == pattern[pi] {
			pi++
		}
	}
	return pi == len(pattern)
}
