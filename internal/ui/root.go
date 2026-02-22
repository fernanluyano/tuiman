package ui

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Model holds all application state. Implements tea.Model.
type Model struct {
	width, height  int
	focused        int   // 0=request, 1=response
	splitVertical  bool  // true=side-by-side (left/right), false=stacked (top/bottom)
	theme          Theme
	showHelp       bool

	activeFolderIdx int // -1 if no request loaded
	activeReqIdx    int // -1 if no request loaded
	folders         []folder
	requestTab    int    // 0=Params, 1=Auth, 2=Headers, 3=Body
	urlInput      string // editable URL (in-memory only)
	urlInputPrev  string // saved before edit, restored on esc
	editingURL    bool
	methodInput   string // selected HTTP method (in-memory only)

	// method picker
	showMethodPicker bool
	methodCursor     int

	// command palette
	showCmdPalette bool
	cmdInput       string
	cmdError       string
	showCmdHelp    bool

	// folder picker
	showFolderPicker bool
	fpExpanded       map[int]bool // set of expanded folder indices
	fpQuery          string
	fpCursor         int
	fpSearchResults  []fpItem // bleve results when query is non-empty
	fpInsert         bool     // insert mode (typing to search)
	fpAdding         bool
	fpAddKind        string // "folder" or "request"
	fpAddInput       string
	fpConfirmDelete  bool
}

// New creates the initial application model with mocked data.
func New() Model {
	folders := mockFolders
	for fi := range folders {
		for ri := range folders[fi].requests {
			folders[fi].requests[ri].searchable = folders[fi].requests[ri].searchText()
		}
	}
	return Model{
		folders:         folders,
		fpExpanded:      map[int]bool{},
		methodInput:     "GET",
		splitVertical:   true,
		theme:           themeXcode,
		activeFolderIdx: -1,
		activeReqIdx:    -1,
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

		if m.showCmdHelp {
			switch msg.String() {
			case "q", "esc":
				m.showCmdHelp = false
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

		if m.showCmdPalette {
			return m.updateCmdPalette(msg), nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "?":
			m.showHelp = true

		// Pane navigation
		case "tab", "shift+tab":
			m.focused = (m.focused + 1) % 2

		// Folder picker
		case "f":
			m.showFolderPicker = true
			m.fpQuery = ""
			m.fpCursor = 0

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
				m.urlInputPrev = m.urlInput
			}

		// Command palette
		case ":":
			m.showCmdPalette = true
			m.cmdInput = ""
			m.cmdError = ""

		// Tab cycling in request pane — h/l or arrow keys
		case "h", "left":
			if m.focused == 0 {
				m.requestTab = (m.requestTab + 3) % 4
			}
		case "l", "right":
			if m.focused == 0 {
				m.requestTab = (m.requestTab + 1) % 4
			}
		// Direct tab jump — p/a/r/b
		case "p":
			if m.focused == 0 {
				m.requestTab = 0
			}
		case "a":
			if m.focused == 0 {
				m.requestTab = 1
			}
		case "r":
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

// fpItem represents one row in the flat folder-picker list.
// If reqIdx < 0 it is a folder row; otherwise it is a request row.
type fpItem struct {
	folderIdx int
	reqIdx    int
}

// fpFlatItems returns the list to display.
// When a query is active it returns bleve search results; otherwise the tree view.
func (m Model) fpFlatItems() []fpItem {
	if m.fpQuery != "" {
		return m.fpSearchResults
	}
	var items []fpItem
	for fi := range m.folders {
		items = append(items, fpItem{folderIdx: fi, reqIdx: -1})
		if m.fpExpanded[fi] {
			for ri := range m.folders[fi].requests {
				items = append(items, fpItem{folderIdx: fi, reqIdx: ri})
			}
		}
	}
	return items
}

// search pipes all folder/request data to rg with --smart-case -F and returns matches.
// Format: one line per item as "fi:ri:text"; rg output is parsed back to fpItems.
func (m Model) search(query string) []fpItem {
	if query == "" {
		return nil
	}

	var sb strings.Builder
	for fi, f := range m.folders {
		fmt.Fprintf(&sb, "%d:-1:%s\n", fi, f.name)
		for ri, r := range f.requests {
			fmt.Fprintf(&sb, "%d:%d:%s\n", fi, ri, r.searchable)
		}
	}

	cmd := exec.Command("rg", "--smart-case", "-F", "--color=never", query)
	cmd.Stdin = strings.NewReader(sb.String())
	out, _ := cmd.Output()
	if len(out) == 0 {
		return nil
	}

	var items []fpItem
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		parts := strings.SplitN(line, ":", 3)
		if len(parts) < 2 {
			continue
		}
		fi, e1 := strconv.Atoi(parts[0])
		ri, e2 := strconv.Atoi(parts[1])
		if e1 != nil || e2 != nil || fi < 0 || fi >= len(m.folders) {
			continue
		}
		if ri != -1 && (ri < 0 || ri >= len(m.folders[fi].requests)) {
			continue
		}
		items = append(items, fpItem{folderIdx: fi, reqIdx: ri})
	}
	return items
}

func (m Model) updateFolderPicker(msg tea.KeyMsg) Model {
	// Confirm-delete mode
	if m.fpConfirmDelete {
		switch msg.String() {
		case "y":
			m = m.performDelete()
		case "n", "esc":
			m.fpConfirmDelete = false
		}
		return m
	}

	// Adding mode
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

	// Insert mode: typing runs a global search via rg
	if m.fpInsert {
		switch msg.String() {
		case "esc":
			m.fpInsert = false
		case "enter":
			m.fpInsert = false
		case "backspace":
			if len(m.fpQuery) > 0 {
				runes := []rune(m.fpQuery)
				m.fpQuery = string(runes[:len(runes)-1])
				if m.fpQuery != "" {
					m.fpSearchResults = m.search(m.fpQuery)
				} else {
					m.fpSearchResults = nil
				}
				items := m.fpFlatItems()
				if m.fpCursor >= len(items) {
					m.fpCursor = max(0, len(items)-1)
				}
			}
		default:
			if len([]rune(msg.String())) == 1 {
				m.fpQuery += msg.String()
				m.fpSearchResults = m.search(m.fpQuery)
				m.fpCursor = 0
			}
		}
		return m
	}

	// Normal mode — flat list navigation
	items := m.fpFlatItems()
	switch msg.String() {
	case "esc":
		if m.fpQuery != "" {
			m.fpQuery = ""
			m.fpSearchResults = nil
			m.fpCursor = 0
		} else {
			m.showFolderPicker = false
			m.fpInsert = false
		}
	case "/":
		if len(m.fpExpanded) > 0 {
			m.fpExpanded = map[int]bool{}
		} else {
			for fi := range m.folders {
				m.fpExpanded[fi] = true
			}
		}
	case "enter":
		m = m.performFpEnter()
	case "j", "down", "ctrl+j":
		if m.fpCursor < len(items)-1 {
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
		if len(m.fpExpanded) > 0 {
			m.fpAddKind = "request"
		} else {
			m.fpAddKind = "folder"
		}
		m.fpAddInput = ""
	case "d":
		if len(items) > 0 {
			m.fpConfirmDelete = true
		}
	}
	return m
}

func (m Model) performFpEnter() Model {
	items := m.fpFlatItems()
	if len(items) == 0 || m.fpCursor >= len(items) {
		return m
	}
	item := items[m.fpCursor]
	if item.reqIdx < 0 {
		// folder row: toggle expand
		if m.fpExpanded[item.folderIdx] {
			delete(m.fpExpanded, item.folderIdx)
		} else {
			m.fpExpanded[item.folderIdx] = true
		}
	} else {
		// request row: select and close picker
		req := m.folders[item.folderIdx].requests[item.reqIdx]
		m.activeFolderIdx = item.folderIdx
		m.activeReqIdx = item.reqIdx
		m.urlInput = req.url
		m.methodInput = req.method
		m.showFolderPicker = false
		m.fpQuery = ""
		m.fpSearchResults = nil
		m.fpInsert = false
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
		m.fpCursor = len(m.fpFlatItems()) - 1
	case "request":
		// add to the folder under the cursor
		items := m.fpFlatItems()
		if m.fpCursor < len(items) {
			fi := items[m.fpCursor].folderIdx
			newReq := request{method: "GET", name: m.fpAddInput, auth: requestAuth{kind: authNone}}
			newReq.searchable = newReq.searchText()
			f := m.folders[fi]
			f.requests = append(f.requests, newReq)
			m.folders[fi] = f
			m.fpExpanded[fi] = true
			m.fpCursor = len(m.fpFlatItems()) - 1
		}
	}
	m.fpAdding = false
	m.fpAddInput = ""
	return m
}

func (m Model) performDelete() Model {
	m.fpConfirmDelete = false
	items := m.fpFlatItems()
	if len(items) == 0 || m.fpCursor >= len(items) {
		return m
	}
	item := items[m.fpCursor]
	if item.reqIdx < 0 {
		// delete folder
		idx := item.folderIdx
		m.folders = append(m.folders[:idx], m.folders[idx+1:]...)
		// rebuild expanded map: drop deleted index, shift higher indices down
		next := map[int]bool{}
		for fi := range m.fpExpanded {
			if fi < idx {
				next[fi] = true
			} else if fi > idx {
				next[fi-1] = true
			}
		}
		m.fpExpanded = next
	} else {
		// delete request
		fi, ri := item.folderIdx, item.reqIdx
		f := m.folders[fi]
		f.requests = append(f.requests[:ri], f.requests[ri+1:]...)
		m.folders[fi] = f
	}
	newItems := m.fpFlatItems()
	if m.fpCursor >= len(newItems) {
		m.fpCursor = max(0, len(newItems)-1)
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
		if m.activeFolderIdx >= 0 {
			r := &m.folders[m.activeFolderIdx].requests[m.activeReqIdx]
			r.method = m.methodInput
			r.searchable = r.searchText()
		}
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
	case "enter":
		m.editingURL = false
		if m.activeFolderIdx >= 0 {
			r := &m.folders[m.activeFolderIdx].requests[m.activeReqIdx]
			r.url = m.urlInput
			r.searchable = r.searchText()
		}
	case "esc":
		m.editingURL = false
		m.urlInput = m.urlInputPrev
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


func (m Model) updateCmdPalette(msg tea.KeyMsg) Model {
	switch msg.String() {
	case "esc":
		m.showCmdPalette = false
		m.cmdInput = ""
		m.cmdError = ""
	case "enter":
		m = m.execCmd(strings.TrimSpace(m.cmdInput))
	case "backspace":
		runes := []rune(m.cmdInput)
		if len(runes) > 0 {
			m.cmdInput = string(runes[:len(runes)-1])
		}
	default:
		if len([]rune(msg.String())) == 1 {
			m.cmdInput += msg.String()
			m.cmdError = ""
		}
	}
	return m
}

func (m Model) execCmd(cmd string) Model {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return m
	}
	switch parts[0] {
	case "orient":
		m.splitVertical = !m.splitVertical
		m.showCmdPalette = false
		m.cmdInput = ""
		m.cmdError = ""
	case "theme":
		if len(parts) < 2 {
			m.cmdError = "usage: theme <name>"
		} else if t, ok := themes[parts[1]]; ok {
			m.theme = t
			m.showCmdPalette = false
			m.cmdInput = ""
			m.cmdError = ""
		} else {
			m.cmdError = "unknown theme: " + parts[1]
		}
	case "help":
		m.showCmdHelp = true
		m.showCmdPalette = false
		m.cmdInput = ""
		m.cmdError = ""
	default:
		m.cmdError = "unknown command: " + parts[0]
	}
	return m
}
