package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func (m Model) View() string {
	if m.width == 0 {
		return ""
	}
	if m.showHelp {
		return m.renderHelp()
	}
	if m.showCmdHelp {
		return m.renderCmdHelp()
	}
	bg := m.renderMain()
	if m.showMethodPicker {
		return placeOverlayAt(bg, m.renderMethodPicker(), 1, 2)
	}
	if m.showFolderPicker {
		return placeOverlay(bg, m.renderFolderPicker(), m.width)
	}
	return bg
}

func (m Model) renderMain() string {
	const footerH = 1
	mainH := m.height - footerH

	var mainArea string
	if m.splitVertical {
		// Side-by-side: request 60% left, response 40% right
		reqOuterW := m.width * 6 / 10
		respOuterW := m.width - reqOuterW
		reqInnerW := reqOuterW - 2
		respInnerW := respOuterW - 2
		innerH := mainH - 2

		requestBox := m.theme.paneStyle(m.focused == 0).
			Width(reqInnerW).
			Height(innerH).
			Render(m.renderRequest(reqInnerW, innerH))

		responseBox := m.theme.paneStyle(m.focused == 1).
			Width(respInnerW).
			Height(innerH).
			Render(m.theme.paneTitle(" Response ", m.focused == 1))

		mainArea = lipgloss.JoinHorizontal(lipgloss.Top, requestBox, responseBox)
	} else {
		// Stacked: request top, response bottom
		innerW := m.width - 2
		reqOuterH := mainH / 2
		respOuterH := mainH - reqOuterH
		reqInnerH := reqOuterH - 2
		respInnerH := respOuterH - 2

		requestBox := m.theme.paneStyle(m.focused == 0).
			Width(innerW).
			Height(reqInnerH).
			Render(m.renderRequest(innerW, reqInnerH))

		responseBox := m.theme.paneStyle(m.focused == 1).
			Width(innerW).
			Height(respInnerH).
			Render(m.theme.paneTitle(" Response ", m.focused == 1))

		mainArea = lipgloss.JoinVertical(lipgloss.Left, requestBox, responseBox)
	}

	bottomBar := m.renderFooter()
	switch {
	case m.showCmdPalette:
		bottomBar = m.renderCmdPalette()
	}
	return lipgloss.JoinVertical(lipgloss.Left, mainArea, bottomBar)
}

func (m Model) renderRequest(w, h int) string {
	div := m.theme.dim().Render(strings.Repeat("─", w))
	// 4 rows overhead: url bar + divider + tab bar + divider
	contentH := h - 4
	return strings.Join([]string{
		m.renderURLBar(w),
		div,
		m.renderRequestTabs(w),
		div,
		m.renderRequestTabContent(w, contentH),
	}, "\n")
}

func (m Model) renderURLBar(w int) string {
	dim := m.theme.dim()
	accent := m.theme.accent()

	mStyle, ok := m.theme.methodStyle(m.methodInput)
	if !ok {
		mStyle = dim
	}
	mLabel := m.theme.keyHint("m")
	badge := mStyle.Bold(true).Render(m.methodInput) + dim.Render(" ▾")

	urlHint := m.theme.keyHint("e")
	if m.editingURL {
		urlHint = m.theme.keyHint("esc")
	}

	sendLabel := m.theme.keyHint("s")
	sendBtn := accent.Bold(true).Render("Send ▶")

	// Fixed-width elements
	mLabelW := lipgloss.Width(mLabel)
	badgeW := lipgloss.Width(badge)
	urlHintW := lipgloss.Width(urlHint)
	sendLabelW := lipgloss.Width(sendLabel)
	sendW := lipgloss.Width(sendBtn)

	// URL gets the remaining space: total - all fixed elements - spacing chars
	urlAvail := w - mLabelW - badgeW - urlHintW - sendLabelW - sendW - 8
	if urlAvail < 1 {
		urlAvail = 1
	}

	var urlRendered string
	if m.editingURL {
		cursor := accent.Render("█")
		text := m.theme.text().MaxWidth(urlAvail - 1).Render(m.urlInput)
		urlRendered = text + cursor
	} else if m.urlInput != "" {
		urlRendered = m.theme.textMuted().MaxWidth(urlAvail).Render(m.urlInput)
	} else {
		urlRendered = dim.MaxWidth(urlAvail).Render("Enter a URL...")
	}

	left := " " + mLabel + " " + badge + "  " + urlHint + " " + urlRendered
	leftW := lipgloss.Width(left)
	right := sendLabel + " " + sendBtn + " "
	rightW := lipgloss.Width(right)
	gap := w - leftW - rightW
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + right
}

func (m Model) renderRequestTabs(w int) string {
	type tabDef struct {
		key   string
		label string
		idx   int
	}
	tabs := []tabDef{
		{"p", "Params", 0},
		{"a", "Auth", 1},
		{"r", "Headers", 2},
		{"b", "Body", 3},
	}

	labels := make([]string, len(tabs))
	for i, t := range tabs {
		labels[i] = t.label
	}

	var parts []string
	for i, t := range tabs {
		keyHint := m.theme.keyHint(t.key)
		if m.requestTab == t.idx {
			tab := m.theme.activeTabStyle().Render(" " + labels[i] + " ")
			parts = append(parts, keyHint+tab)
		} else {
			tab := m.theme.dim().Render(" " + labels[i] + " ")
			parts = append(parts, keyHint+tab)
		}
	}
	return " " + strings.Join(parts, "  ")
}

func (m Model) renderRequestTabContent(w, h int) string {
	if m.activeFolderIdx < 0 {
		msg := "No request selected — press f to open folders"
		return m.theme.dim().Render("  " + msg)
	}

	req := m.folders[m.activeFolderIdx].requests[m.activeReqIdx]
	switch m.requestTab {
	case 0:
		return m.renderKVTable(paramsToKV(req.params), "Key", "Value", w)
	case 1:
		return m.renderAuthContent(req.auth)
	case 2:
		return m.renderKVTable(headersToKV(req.headers), "Key", "Value", w)
	case 3:
		return m.renderBodyContent(req.body)
	}
	return ""
}

func (m Model) renderKVTable(pairs [][2]string, keyHeader, valHeader string, w int) string {
	keyW := w * 4 / 10

	dim := m.theme.dim()
	val := m.theme.textMuted()

	hk := dim.Bold(true).Render(fmt.Sprintf("  %-*s", keyW, keyHeader))
	hv := dim.Bold(true).Render(valHeader)
	sep := dim.Render(strings.Repeat("─", w))

	lines := []string{hk + hv, sep}
	if len(pairs) == 0 {
		lines = append(lines, dim.Render("  (empty)"))
	} else {
		for _, p := range pairs {
			kCell := fmt.Sprintf("  %-*s", keyW, p[0])
			lines = append(lines, dim.Render(kCell)+val.Render(p[1]))
		}
	}
	return strings.Join(lines, "\n")
}

func paramsToKV(params []param) [][2]string {
	out := make([][2]string, len(params))
	for i, p := range params {
		out[i] = [2]string{p.key, p.value}
	}
	return out
}

func headersToKV(headers []header) [][2]string {
	out := make([][2]string, len(headers))
	for i, h := range headers {
		out[i] = [2]string{h.key, h.value}
	}
	return out
}

func (m Model) renderAuthContent(auth requestAuth) string {
	dim := m.theme.dim()
	val := m.theme.textMuted()
	label := m.theme.highlight().Bold(true)

	kindLabel := map[authKind]string{
		authNone:   "No Auth",
		authBearer: "Bearer Token",
		authBasic:  "Basic Auth",
		authAPIKey: "API Key",
	}

	var lines []string
	lines = append(lines, dim.Render("  Type    ")+val.Render(kindLabel[auth.kind]))
	lines = append(lines, dim.Render(strings.Repeat("─", 40)))

	switch auth.kind {
	case authNone:
		lines = append(lines, dim.Render("  No authentication configured."))
	case authBearer:
		lines = append(lines, label.Render("  Token"))
		lines = append(lines, "  "+val.Render(auth.token))
	case authBasic:
		lines = append(lines, label.Render("  Username"))
		lines = append(lines, "  "+val.Render(auth.username))
		lines = append(lines, "")
		lines = append(lines, label.Render("  Password"))
		lines = append(lines, "  "+val.Render(strings.Repeat("●", len(auth.password))))
	case authAPIKey:
		lines = append(lines, label.Render("  Key"))
		lines = append(lines, "  "+val.Render(auth.apiKey))
		lines = append(lines, "")
		lines = append(lines, label.Render("  Value"))
		lines = append(lines, "  "+val.Render(auth.apiValue))
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderBodyContent(body string) string {
	if body == "" {
		return m.theme.dim().Render("  (empty)")
	}
	val := m.theme.textMuted()
	var lines []string
	for _, l := range strings.Split(body, "\n") {
		lines = append(lines, val.Render("  "+l))
	}
	return strings.Join(lines, "\n")
}


func (m Model) renderCmdPalette() string {
	dim := m.theme.dim()
	prompt := m.theme.highlight().Bold(true).Render(":")
	cursor := m.theme.accent().Render("█")
	input := m.theme.text().Render(m.cmdInput)

	left := prompt + input + cursor

	var right string
	if m.cmdError != "" {
		right = m.theme.errStyle().Render("  " + m.cmdError)
	} else {
		right = dim.Render("  esc to close")
	}

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + right
}

func (m Model) renderFooter() string {
	hints := []struct{ key, desc string }{
		{"q", "quit"},
		{"?", "help"},
		{"tab", "pane"},
		{"h/l", "tab"},
		{"f", "folders"},
		{"m", "method"},
		{"e", "edit url"},
		{"s", "send"},
		{":", "commands"},
	}

	var parts []string
	for _, h := range hints {
		k := m.theme.footerKeyStyle().Render(h.key)
		d := m.theme.footerDescStyle().Render(" " + h.desc)
		parts = append(parts, "  "+k+d)
	}

	return m.theme.footerDescStyle().Render(strings.Join(parts, ""))
}

// renderFolderPicker renders the two-level floating folder picker.
func (m Model) renderFolderPicker() string {
	pickerOuterW := m.width - 6
	if pickerOuterW < 60 {
		pickerOuterW = 60
	}
	pickerInnerW := pickerOuterW - 4 // border(1 each) + padding(1 each)

	pickerH := m.height * 6 / 10
	if pickerH < 12 {
		pickerH = 12
	}
	// header + query + hdiv = 3 lines; border top+bottom = 2; content rows = pickerH - 5
	contentH := pickerH - 5

	listW := pickerInnerW * 2 / 5
	previewW := pickerInnerW - listW - 1 // 1 for vertical divider

	dim := m.theme.dim()
	yellow := m.theme.highlight().Bold(true)
	orange := m.theme.accent()

	// --- Header line ---
	kh := func(key, label string) string {
		return "  " + m.theme.keyHint(key) + dim.Render(label)
	}
	items := m.fpFlatItems()
	enterHint := "expand"
	if len(items) > 0 && m.fpCursor < len(items) && items[m.fpCursor].reqIdx >= 0 {
		enterHint = "select"
	}
	headerText := yellow.Render(" Folders") +
		kh("i", "filter") + kh("n", "new") + kh("d", "del") + kh("enter", enterHint) + kh("/", "expand all") + kh("esc", "close")

	// --- Query / add-input / confirm-delete line ---
	var queryLine string
	switch {
	case m.fpConfirmDelete:
		var itemName string
		if m.fpCursor < len(items) {
			it := items[m.fpCursor]
			if it.reqIdx < 0 {
				itemName = m.folders[it.folderIdx].name
			} else {
				itemName = m.folders[it.folderIdx].requests[it.reqIdx].name
			}
		}
		errHint := func(key string) string {
			return dim.Render("(") + m.theme.errStyle().Render(key) + dim.Render(")")
		}
		queryLine = dim.Render(" Delete ") +
			m.theme.errStyle().Bold(true).Render(`"`+itemName+`"`) +
			dim.Render("?  ") +
			errHint("y") + dim.Render("yes  ") +
			errHint("n") + dim.Render("no")
	case m.fpAdding:
		prompt := "New folder name: "
		if m.fpAddKind == "request" {
			prompt = "New request name: "
		}
		queryLine = dim.Render(" "+prompt) +
			m.theme.text().Render(m.fpAddInput) +
			orange.Render("█") +
			"  " + m.theme.keyHint("enter") + dim.Render("save") +
			"  " + m.theme.keyHint("esc") + dim.Render("cancel")
	case m.fpInsert:
		queryLine = dim.Render(" -- SEARCH --  > ") +
			m.theme.text().Render(m.fpQuery) +
			orange.Render("█") +
			"  " + m.theme.keyHint("esc") + dim.Render("normal")
	default:
		if m.fpQuery != "" {
			queryLine = dim.Render(" -- SEARCH --  > ") +
				m.theme.textMuted().Render(m.fpQuery) +
				"  " + m.theme.keyHint("esc") + dim.Render("clear")
		} else {
			queryLine = dim.Render(" -- NORMAL --  > ") +
				m.theme.textMuted().Render(m.fpQuery)
		}
	}

	hdiv := dim.Render(strings.Repeat("─", pickerInnerW))

	// --- List pane ---
	const maxVisible = 15
	var itemLines []string

	for i, it := range items {
		if i >= maxVisible {
			break
		}
		if it.reqIdx < 0 {
			// folder row
			f := m.folders[it.folderIdx]
			count := dim.Render(fmt.Sprintf("(%d)", len(f.requests)))
			chevron := dim.Render("▸ ")
			if m.fpExpanded[it.folderIdx] {
				chevron = orange.Render("▾ ")
			}
			if i == m.fpCursor {
				prefix := orange.Bold(true).Render("> ")
				text := chevron + lipgloss.NewStyle().Bold(true).Render(f.name) + " " + count
				itemLines = append(itemLines, lipgloss.NewStyle().MaxWidth(listW).Render(prefix+text))
			} else {
				text := chevron + f.name + " " + count
				itemLines = append(itemLines, lipgloss.NewStyle().MaxWidth(listW).Render(dim.Render("  ")+text))
			}
		} else {
			// request row
			r := m.folders[it.folderIdx].requests[it.reqIdx]
			selected := i == m.fpCursor
			if m.fpQuery != "" {
				// search mode: show flat with folder name as context
				itemLines = append(itemLines, m.renderSearchReqItem(it, r, selected, listW))
			} else {
				// tree mode: indented under the expanded folder
				line := "  " + m.renderFolderReqItem(r, selected, listW-2)
				itemLines = append(itemLines, line)
			}
		}
	}
	if len(items) == 0 {
		itemLines = append(itemLines, dim.Render("  no results"))
	}

	listPane := lipgloss.NewStyle().
		Width(listW).
		Height(contentH).
		Render(strings.Join(itemLines, "\n"))

	// --- Preview pane ---
	var previewContent string
	if len(items) > 0 && m.fpCursor < len(items) {
		it := items[m.fpCursor]
		if it.reqIdx < 0 {
			previewContent = m.renderFolderPreview(m.folders[it.folderIdx], previewW)
		} else {
			previewContent = m.renderRequestPreview(m.folders[it.folderIdx].requests[it.reqIdx], previewW)
		}
	} else {
		previewContent = dim.Render("  nothing selected")
	}

	previewPane := lipgloss.NewStyle().
		Width(previewW).
		Height(contentH).
		Render(previewContent)

	// --- Vertical divider ---
	vdiv := dim.Render(strings.Repeat("│\n", contentH-1) + "│")

	contentArea := lipgloss.JoinHorizontal(lipgloss.Top, listPane, vdiv, previewPane)
	content := strings.Join([]string{headerText, queryLine, hdiv, contentArea}, "\n")

	return m.theme.overlayStyle().
		Padding(0, 1).
		Width(pickerInnerW).
		Render(content)
}

// renderSearchReqItem renders a request row in global search mode,
// appending the folder name as dim context on the right.
func (m Model) renderSearchReqItem(it fpItem, r request, selected bool, maxW int) string {
	st, ok := m.theme.methodStyle(r.method)
	if !ok {
		st = lipgloss.NewStyle()
	}
	method := st.Render(fmt.Sprintf("%-6s", r.method))
	folder := m.theme.dim().Render(" " + m.folders[it.folderIdx].name)

	var line string
	if selected {
		prefix := m.theme.accent().Bold(true).Render("> ")
		line = prefix + lipgloss.NewStyle().Bold(true).Render(method+" "+r.name) + folder
	} else {
		line = m.theme.dim().Render("  ") + method + " " + r.name + folder
	}
	return lipgloss.NewStyle().MaxWidth(maxW).Render(line)
}

func (m Model) renderFolderReqItem(r request, selected bool, maxW int) string {
	st, ok := m.theme.methodStyle(r.method)
	if !ok {
		st = lipgloss.NewStyle()
	}
	method := st.Render(fmt.Sprintf("%-6s", r.method))
	text := method + " " + r.name

	var line string
	if selected {
		prefix := m.theme.accent().Bold(true).Render("> ")
		line = prefix + lipgloss.NewStyle().Bold(true).Render(text)
	} else {
		line = m.theme.dim().Render("  ") + text
	}
	return lipgloss.NewStyle().MaxWidth(maxW).Render(line)
}

func (m Model) renderFolderPreview(f folder, width int) string {
	dim := m.theme.dim()
	val := m.theme.textMuted()
	label := m.theme.highlight().Bold(true)

	n := len(f.requests)
	countStr := "1 request"
	if n != 1 {
		countStr = fmt.Sprintf("%d requests", n)
	}

	var lines []string
	lines = append(lines, label.Render(f.name))
	lines = append(lines, dim.Render(countStr))
	lines = append(lines, dim.Render(strings.Repeat("─", width-2)))

	if n == 0 {
		lines = append(lines, dim.Render("  (empty)"))
	} else {
		for _, r := range f.requests {
			st, ok := m.theme.methodStyle(r.method)
			if !ok {
				st = lipgloss.NewStyle()
			}
			lines = append(lines, "  "+st.Render(fmt.Sprintf("%-6s", r.method))+" "+val.Render(r.name))
		}
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderRequestPreview(r request, width int) string {
	labelStyle := m.theme.highlight().Bold(true)
	dimStyle := m.theme.dim()
	valStyle := m.theme.textMuted()

	st, ok := m.theme.methodStyle(r.method)
	if !ok {
		st = lipgloss.NewStyle()
	}

	var lines []string

	lines = append(lines, st.Bold(true).Render(r.method)+"  "+valStyle.Render(r.url))
	lines = append(lines, dimStyle.Render(strings.Repeat("─", width-2)))

	lines = append(lines, labelStyle.Render("Headers"))
	if len(r.headers) == 0 {
		lines = append(lines, dimStyle.Render("  (none)"))
	} else {
		for _, h := range r.headers {
			lines = append(lines, "  "+dimStyle.Render(h.key+": ")+valStyle.Render(h.value))
		}
	}
	lines = append(lines, "")

	lines = append(lines, labelStyle.Render("Query Params"))
	if len(r.params) == 0 {
		lines = append(lines, dimStyle.Render("  (none)"))
	} else {
		for _, p := range r.params {
			lines = append(lines, "  "+dimStyle.Render(p.key+": ")+valStyle.Render(p.value))
		}
	}
	lines = append(lines, "")

	lines = append(lines, labelStyle.Render("Body"))
	if r.body == "" {
		lines = append(lines, dimStyle.Render("  (none)"))
	} else {
		for _, l := range strings.Split(r.body, "\n") {
			lines = append(lines, "  "+valStyle.Render(l))
		}
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderMethodPicker() string {
	var lines []string
	for i, method := range httpMethods {
		st, ok := m.theme.methodStyle(method)
		if !ok {
			st = lipgloss.NewStyle()
		}
		if i == m.methodCursor {
			prefix := m.theme.accent().Bold(true).Render("> ")
			lines = append(lines, prefix+st.Bold(true).Render(method))
		} else {
			lines = append(lines, m.theme.dim().Render("  ")+st.Render(method))
		}
	}
	return m.theme.overlayStyle().
		Padding(0, 1).
		Render(strings.Join(lines, "\n"))
}

// placeOverlayAt composites fg over bg at a specific (x, y) position.
func placeOverlayAt(bg, fg string, x, y int) string {
	bgLines := strings.Split(bg, "\n")
	fgLines := strings.Split(fg, "\n")

	fgW := 0
	for _, l := range fgLines {
		if w := lipgloss.Width(l); w > fgW {
			fgW = w
		}
	}

	for i, fgLine := range fgLines {
		bgRow := y + i
		if bgRow < 0 || bgRow >= len(bgLines) {
			continue
		}
		plain := []rune(ansi.Strip(bgLines[bgRow]))
		bgLineLen := len(plain)

		left := ""
		if x > 0 {
			end := x
			if end > bgLineLen {
				end = bgLineLen
			}
			left = string(plain[:end])
			if len([]rune(left)) < x {
				left += strings.Repeat(" ", x-len([]rune(left)))
			}
		}
		rightStart := x + fgW
		right := ""
		if rightStart < bgLineLen {
			right = string(plain[rightStart:])
		}
		bgLines[bgRow] = left + fgLine + right
	}
	return strings.Join(bgLines, "\n")
}

// placeOverlay composites the fg string centered over the bg string.
// ANSI codes in fg are preserved; bg is stripped to plain text in the overlay region.
func placeOverlay(bg, fg string, bgW int) string {
	bgLines := strings.Split(bg, "\n")
	fgLines := strings.Split(fg, "\n")
	bgH := len(bgLines)

	fgH := len(fgLines)
	fgW := 0
	for _, l := range fgLines {
		if w := lipgloss.Width(l); w > fgW {
			fgW = w
		}
	}

	x := (bgW - fgW) / 2
	y := (bgH - fgH) / 2

	for i, fgLine := range fgLines {
		bgRow := y + i
		if bgRow < 0 || bgRow >= len(bgLines) {
			continue
		}

		plain := []rune(ansi.Strip(bgLines[bgRow]))
		bgLineLen := len(plain)

		// Left portion of background
		left := ""
		if x > 0 {
			end := x
			if end > bgLineLen {
				end = bgLineLen
			}
			left = string(plain[:end])
			if len([]rune(left)) < x {
				left += strings.Repeat(" ", x-len([]rune(left)))
			}
		}

		// Right portion of background
		rightStart := x + fgW
		right := ""
		if rightStart < bgLineLen {
			right = string(plain[rightStart:])
		}

		bgLines[bgRow] = left + fgLine + right
	}

	return strings.Join(bgLines, "\n")
}

func (m Model) renderCmdHelp() string {
	type row struct {
		cmd  string
		desc string
	}
	commands := []row{
		{":orient", "toggle split direction (left/right ↔ top/bottom)"},
		{":theme <name>", "switch color theme"},
		{"", "rosepine · xcode · catppuccin · tokyonight · sonokai"},
		{":help", "show this commands list"},
	}

	var lines []string
	lines = append(lines, m.theme.helpTitleStyle().Render("  Commands  "))
	lines = append(lines, "")
	for _, r := range commands {
		if r.cmd == "" {
			lines = append(lines, "       "+m.theme.dim().Render(r.desc))
		} else {
			k := m.theme.helpKeyStyle().Render(fmt.Sprintf("   %-16s", r.cmd))
			lines = append(lines, k+r.desc)
		}
	}
	lines = append(lines, "")
	lines = append(lines, m.theme.footerDescStyle().Render("  press esc or q to close"))

	box := m.theme.helpOverlayStyle().
		Padding(0, 2).
		Render(strings.Join(lines, "\n"))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m Model) renderHelp() string {
	type row struct{ key, desc string }
	sections := []struct {
		title string
		rows  []row
	}{
		{"Folders", []row{
			{"f", "open folder picker"},
			{"j / k", "navigate list"},
			{"enter", "open folder / select request"},
			{"i", "enter insert mode (filter list)"},
			{"esc (insert)", "return to normal mode"},
			{"n", "new folder or request (normal mode)"},
			{"d", "delete selected (normal mode)"},
			{"esc (normal)", "back / close picker"},
		}},
		{"Pane Navigation", []row{
			{"tab / shift+tab", "cycle pane"},
		}},
		{"Request Pane", []row{
			{"h / l", "prev / next tab"},
			{"p / a / r / b", "jump to Params / Auth / Headers / Body"},
			{"m", "change method"},
			{"e", "edit URL"},
			{"s", "send request"},
			{"esc / enter", "stop editing"},
		}},
		{"Global", []row{
			{":", "open command palette  (:help for commands)"},
			{"?", "toggle help"},
			{"q", "quit"},
		}},
	}

	var lines []string
	lines = append(lines, m.theme.helpTitleStyle().Render("  Keybindings  "))
	lines = append(lines, "")

	for _, sec := range sections {
		lines = append(lines, "  "+m.theme.helpTitleStyle().Render(sec.title))
		for _, r := range sec.rows {
			k := m.theme.helpKeyStyle().Render(fmt.Sprintf("   %-20s", r.key))
			lines = append(lines, k+r.desc)
		}
		lines = append(lines, "")
	}

	lines = append(lines, m.theme.footerDescStyle().Render("  press ? or q to close"))

	box := m.theme.helpOverlayStyle().
		Padding(0, 2).
		Render(strings.Join(lines, "\n"))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}
