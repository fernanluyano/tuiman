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
	case m.confirmQuit:
		bottomBar = m.renderConfirmQuit()
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
	mLabel := dim.Render("(m)")
	badge := mStyle.Bold(true).Render(m.methodInput) + dim.Render(" ▾")

	urlHint := dim.Render("(e)")
	if m.editingURL {
		urlHint = dim.Render("(esc)")
	}

	sendLabel := dim.Render("(s)")
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
		{"h", "Headers", 2},
		{"b", "Body", 3},
	}

	labels := make([]string, len(tabs))
	for i, t := range tabs {
		labels[i] = t.label
	}

	keyStyle := m.theme.dim()
	var parts []string
	for i, t := range tabs {
		keyHint := keyStyle.Render("(" + t.key + ")")
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
	if m.activeRequest == nil {
		msg := "No request selected — press f to open folders"
		return m.theme.dim().Render("  " + msg)
	}

	switch m.requestTab {
	case 0:
		return m.renderKVTable(paramsToKV(m.activeRequest.params), "Key", "Value", w)
	case 1:
		return m.renderAuthContent(m.activeRequest.auth)
	case 2:
		return m.renderKVTable(headersToKV(m.activeRequest.headers), "Key", "Value", w)
	case 3:
		return m.renderBodyContent(m.activeRequest.body)
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

func (m Model) renderConfirmQuit() string {
	dim := m.theme.dim()
	msg := m.theme.accent().Bold(true).Render("Quit tuiman?")
	yes := m.theme.errStyle().Bold(true).Render("(y)")
	no := m.theme.successStyle().Render("(n/esc)")
	return "  " + msg + "  " + yes + dim.Render(" yes  ") + no + dim.Render(" no")
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
		{"][", "tab"},
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
	var headerText string
	if m.fpLevel == 0 {
		headerText = yellow.Render(" Folders") +
			dim.Render("  (i)filter  (n)new  (d)del  (enter)open  (esc)close")
	} else {
		folderName := m.folders[m.fpFolderIdx].name
		headerText = yellow.Render(" ← "+folderName) +
			dim.Render("  (i)filter  (n)new  (d)del  (enter)select  (esc)back")
	}

	// --- Query / add-input / confirm-delete line ---
	var queryLine string
	switch {
	case m.fpConfirmDelete:
		var itemName string
		if m.fpLevel == 0 && len(m.fpFolderShown) > 0 {
			itemName = m.folders[m.fpFolderShown[m.fpCursor]].name
		} else if m.fpLevel == 1 && len(m.fpReqShown) > 0 {
			itemName = m.folders[m.fpFolderIdx].requests[m.fpReqShown[m.fpCursor]].name
		}
		queryLine = dim.Render(" Delete ") +
			m.theme.errStyle().Bold(true).Render(`"`+itemName+`"`) +
			dim.Render("?  ") +
			m.theme.errStyle().Render("(y)") + dim.Render("yes  ") +
			dim.Render("(n)") + dim.Render("no")
	case m.fpAdding:
		prompt := "New folder name: "
		if m.fpAddKind == "request" {
			prompt = "New request name: "
		}
		queryLine = dim.Render(" "+prompt) +
			m.theme.text().Render(m.fpAddInput) +
			orange.Render("█") +
			dim.Render("  (enter)save  (esc)cancel")
	case m.fpInsert:
		queryLine = dim.Render(" -- INSERT --  > ") +
			m.theme.text().Render(m.fpQuery) +
			orange.Render("█") +
			dim.Render("  (esc)normal")
	default: // normal mode
		queryLine = dim.Render(" -- NORMAL --  (i)filter > ") +
			m.theme.textMuted().Render(m.fpQuery)
	}

	hdiv := dim.Render(strings.Repeat("─", pickerInnerW))

	// --- List pane ---
	const maxVisible = 15
	var itemLines []string

	if m.fpLevel == 0 {
		for i, fi := range m.fpFolderShown {
			if i >= maxVisible {
				break
			}
			f := m.folders[fi]
			count := dim.Render(fmt.Sprintf("(%d)", len(f.requests)))
			if i == m.fpCursor {
				prefix := orange.Bold(true).Render("> ")
				text := lipgloss.NewStyle().Bold(true).Render(f.name) + " " + count
				itemLines = append(itemLines, lipgloss.NewStyle().MaxWidth(listW).Render(prefix+text))
			} else {
				text := f.name + " " + count
				itemLines = append(itemLines, lipgloss.NewStyle().MaxWidth(listW).Render(dim.Render("  ")+text))
			}
		}
		if len(m.fpFolderShown) == 0 {
			itemLines = append(itemLines, dim.Render("  no results"))
		}
	} else {
		reqs := m.folders[m.fpFolderIdx].requests
		for i, ri := range m.fpReqShown {
			if i >= maxVisible {
				break
			}
			itemLines = append(itemLines, m.renderFolderReqItem(reqs[ri], i == m.fpCursor, listW))
		}
		if len(m.fpReqShown) == 0 {
			itemLines = append(itemLines, dim.Render("  no results"))
		}
	}

	listPane := lipgloss.NewStyle().
		Width(listW).
		Height(contentH).
		Render(strings.Join(itemLines, "\n"))

	// --- Preview pane ---
	var previewContent string
	if m.fpLevel == 0 {
		if len(m.fpFolderShown) > 0 {
			previewContent = m.renderFolderPreview(m.folders[m.fpFolderShown[m.fpCursor]], previewW)
		} else {
			previewContent = dim.Render("  nothing selected")
		}
	} else {
		if len(m.fpReqShown) > 0 {
			r := m.folders[m.fpFolderIdx].requests[m.fpReqShown[m.fpCursor]]
			previewContent = m.renderRequestPreview(r, previewW)
		} else {
			previewContent = dim.Render("  nothing selected")
		}
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
			{"[ / ]", "prev / next tab"},
			{"p / a / h / b", "jump to Params / Auth / Headers / Body"},
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
