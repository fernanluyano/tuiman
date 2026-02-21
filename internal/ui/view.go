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
	innerW := m.width - 2

	reqOuterH := mainH / 2
	respOuterH := mainH - reqOuterH
	reqInnerH := reqOuterH - 2
	respInnerH := respOuterH - 2

	requestBox := paneStyle(m.focused == 0).
		Width(innerW).
		Height(reqInnerH).
		Render(m.renderRequest(innerW, reqInnerH))

	responseBox := paneStyle(m.focused == 1).
		Width(innerW).
		Height(respInnerH).
		Render(paneTitle(" Response ", m.focused == 1))

	mainArea := lipgloss.JoinVertical(lipgloss.Left, requestBox, responseBox)
	return lipgloss.JoinVertical(lipgloss.Left, mainArea, m.renderFooter())
}

func (m Model) renderRequest(w, h int) string {
	div := lipgloss.NewStyle().Foreground(colorDimmed).Render(strings.Repeat("─", w))
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
	dim := lipgloss.NewStyle().Foreground(colorDimmed)

	mStyle, ok := methodStyles[m.methodInput]
	if !ok {
		mStyle = lipgloss.NewStyle().Foreground(colorDimmed)
	}
	mLabel := dim.Render("(m)")
	badge := mStyle.Bold(true).Render(m.methodInput) + dim.Render(" ▾")

	urlHint := dim.Render("(e)")
	if m.editingURL {
		urlHint = dim.Render("(esc)")
	}

	sendLabel := dim.Render("(s)")
	sendBtn := lipgloss.NewStyle().Foreground(colorOrange).Bold(true).Render("Send ▶")

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
		cursor := lipgloss.NewStyle().Foreground(colorOrange).Render("█")
		text := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).
			MaxWidth(urlAvail - 1).Render(m.urlInput)
		urlRendered = text + cursor
	} else if m.urlInput != "" {
		urlRendered = lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")).
			MaxWidth(urlAvail).Render(m.urlInput)
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

	keyStyle := lipgloss.NewStyle().Foreground(colorDimmed)
	var parts []string
	for i, t := range tabs {
		keyHint := keyStyle.Render("(" + t.key + ")")
		if m.requestTab == t.idx {
			tab := lipgloss.NewStyle().Foreground(colorGreen).Bold(true).Underline(true).
				Render(" " + labels[i] + " ")
			parts = append(parts, keyHint+tab)
		} else {
			tab := lipgloss.NewStyle().Foreground(colorDimmed).Render(" " + labels[i] + " ")
			parts = append(parts, keyHint+tab)
		}
	}
	return " " + strings.Join(parts, "  ")
}

func (m Model) renderRequestTabContent(w, h int) string {
	if m.activeRequest == nil {
		msg := "No request selected — press f to open folders"
		return lipgloss.NewStyle().Foreground(colorDimmed).Render("  " + msg)
	}

	switch m.requestTab {
	case 0:
		return renderKVTable(paramsToKV(m.activeRequest.params), "Key", "Value", w)
	case 1:
		return renderAuthContent(m.activeRequest.auth)
	case 2:
		return renderKVTable(headersToKV(m.activeRequest.headers), "Key", "Value", w)
	case 3:
		return renderBodyContent(m.activeRequest.body)
	}
	return ""
}

func renderKVTable(pairs [][2]string, keyHeader, valHeader string, w int) string {
	keyW := w * 4 / 10

	dim := lipgloss.NewStyle().Foreground(colorDimmed)
	val := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))

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

func renderAuthContent(auth requestAuth) string {
	dim := lipgloss.NewStyle().Foreground(colorDimmed)
	val := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	label := lipgloss.NewStyle().Foreground(colorYellow).Bold(true)

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

func renderBodyContent(body string) string {
	if body == "" {
		return lipgloss.NewStyle().Foreground(colorDimmed).Render("  (empty)")
	}
	val := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	var lines []string
	for _, l := range strings.Split(body, "\n") {
		lines = append(lines, val.Render("  "+l))
	}
	return strings.Join(lines, "\n")
}

func paneTitle(title string, active bool) string {
	if active {
		return lipgloss.NewStyle().Foreground(colorOrange).Render(title)
	}
	return lipgloss.NewStyle().Foreground(colorDimmed).Render(title)
}

func (m Model) renderFooter() string {
	hints := []struct{ key, desc string }{
		{"q", "quit"},
		{"?", "help"},
		{"tab/j/k", "pane"},
		{"[/]", "tab"},
		{"f", "folders"},
		{"m", "method"},
		{"e", "edit url"},
		{"s", "send"},
	}

	var parts []string
	for _, h := range hints {
		k := footerKeyStyle.Render(h.key)
		d := footerDescStyle.Render(" " + h.desc)
		parts = append(parts, "  "+k+d)
	}

	return footerDescStyle.Render(strings.Join(parts, ""))
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

	dim := lipgloss.NewStyle().Foreground(colorDimmed)
	yellow := lipgloss.NewStyle().Foreground(colorYellow).Bold(true)
	orange := lipgloss.NewStyle().Foreground(colorOrange)

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
			lipgloss.NewStyle().Foreground(colorRed).Bold(true).Render(`"`+itemName+`"`) +
			dim.Render("?  ") +
			lipgloss.NewStyle().Foreground(colorRed).Render("(y)") + dim.Render("yes  ") +
			lipgloss.NewStyle().Foreground(colorDimmed).Render("(n)") + dim.Render("no")
	case m.fpAdding:
		prompt := "New folder name: "
		if m.fpAddKind == "request" {
			prompt = "New request name: "
		}
		queryLine = dim.Render(" "+prompt) +
			lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Render(m.fpAddInput) +
			orange.Render("█") +
			dim.Render("  (enter)save  (esc)cancel")
	case m.fpInsert:
		queryLine = dim.Render(" -- INSERT --  > ") +
			lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Render(m.fpQuery) +
			orange.Render("█") +
			dim.Render("  (esc)normal")
	default: // normal mode
		queryLine = dim.Render(" -- NORMAL --  (i)filter > ") +
			lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")).Render(m.fpQuery)
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
			itemLines = append(itemLines, renderFolderReqItem(reqs[ri], i == m.fpCursor, listW))
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
			previewContent = renderFolderPreview(m.folders[m.fpFolderShown[m.fpCursor]], previewW)
		} else {
			previewContent = dim.Render("  nothing selected")
		}
	} else {
		if len(m.fpReqShown) > 0 {
			r := m.folders[m.fpFolderIdx].requests[m.fpReqShown[m.fpCursor]]
			previewContent = renderRequestPreview(r, previewW)
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

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorOrange).
		Padding(0, 1).
		Width(pickerInnerW).
		Render(content)
}

func renderFolderReqItem(r request, selected bool, maxW int) string {
	st, ok := methodStyles[r.method]
	if !ok {
		st = lipgloss.NewStyle()
	}
	method := st.Render(fmt.Sprintf("%-6s", r.method))
	text := method + " " + r.name

	var line string
	if selected {
		prefix := lipgloss.NewStyle().Foreground(colorOrange).Bold(true).Render("> ")
		line = prefix + lipgloss.NewStyle().Bold(true).Render(text)
	} else {
		line = lipgloss.NewStyle().Foreground(colorDimmed).Render("  ") + text
	}
	return lipgloss.NewStyle().MaxWidth(maxW).Render(line)
}

func renderFolderPreview(f folder, width int) string {
	dim := lipgloss.NewStyle().Foreground(colorDimmed)
	val := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	label := lipgloss.NewStyle().Foreground(colorYellow).Bold(true)

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
			st, ok := methodStyles[r.method]
			if !ok {
				st = lipgloss.NewStyle()
			}
			lines = append(lines, "  "+st.Render(fmt.Sprintf("%-6s", r.method))+" "+val.Render(r.name))
		}
	}
	return strings.Join(lines, "\n")
}

func renderRequestPreview(r request, width int) string {
	labelStyle := lipgloss.NewStyle().Foreground(colorYellow).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(colorDimmed)
	valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))

	st, ok := methodStyles[r.method]
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
		st, ok := methodStyles[method]
		if !ok {
			st = lipgloss.NewStyle()
		}
		if i == m.methodCursor {
			prefix := lipgloss.NewStyle().Foreground(colorOrange).Bold(true).Render("> ")
			lines = append(lines, prefix+st.Bold(true).Render(method))
		} else {
			lines = append(lines, lipgloss.NewStyle().Foreground(colorDimmed).Render("  ")+st.Render(method))
		}
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorOrange).
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
			{"j / k", "cycle pane (vim)"},
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
			k := helpKeyStyle.Render(fmt.Sprintf("   %-20s", r.key))
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
