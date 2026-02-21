package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const helpText = `
 [yellow]Navigation[-]

   [green]j / ↓[-]      Move down
   [green]k / ↑[-]      Move up

 [yellow]Collections[-]

   [green]space[-]       Expand / collapse
   [green]enter[-]       Open request

 [yellow]Pane Focus[-]

   [green]tab[-]         Next pane
   [green]shift+tab[-]   Previous pane
   [green]S[-]           Jump to Sidebar
   [green]R[-]           Jump to Request
   [green]P[-]           Jump to resPonse

 [yellow]Global[-]

   [green]?[-]           Toggle this help
   [green]q[-]           Quit


 Press [green]?[-] or [green]q[-] to close help
`

func newFooter() *tview.TextView {
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetText("  [yellow]q[-] quit   [yellow]?[-] help   [yellow]tab[-] next pane   [yellow]S/R/P[-] jump to pane   [yellow]↑↓/jk[-] navigate   [yellow]space[-] expand")
	tv.SetBackgroundColor(tcell.ColorDefault)
	return tv
}

func newHelpModal(onClose func()) *tview.TextView {
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetText(helpText).
		SetDoneFunc(func(key tcell.Key) { onClose() })
	tv.SetBorder(true).
		SetTitle(" Keybindings ").
		SetTitleColor(tcell.ColorYellow).
		SetBorderColor(tcell.ColorYellow)
	return tv
}

func newPlaceholder(title string) *tview.TextView {
	tv := tview.NewTextView().
		SetText("").
		SetTextAlign(tview.AlignCenter)
	tv.SetBorder(true).SetTitle(title)
	return tv
}

// centeredOverlay wraps a primitive in a centered overlay of the given dimensions.
func centeredOverlay(p tview.Primitive, width, height int) tview.Primitive {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(
			tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(p, height, 0, true).
				AddItem(nil, 0, 1, false),
			width, 0, true,
		).
		AddItem(nil, 0, 1, false)
}

// postmanOrange is the accent color used to highlight the active pane border.
var postmanOrange = tcell.NewRGBColor(255, 108, 55)

// borderSetter is satisfied by any tview widget that exposes SetBorderColor.
type borderSetter interface {
	SetBorderColor(tcell.Color) *tview.Box
}

func NewApp() *tview.Application {
	app := tview.NewApplication()

	sidebar := newSidebar()
	requestPanel := newPlaceholder(" Request [green](R)[-] ")
	responsePanel := newPlaceholder(" Response [green](P)[-] ")

	rightPanel := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(requestPanel, 0, 1, false).
		AddItem(responsePanel, 0, 1, false)

	mainFlex := tview.NewFlex().
		AddItem(sidebar, 30, 0, true).
		AddItem(rightPanel, 0, 1, false)

	footer := newFooter()

	mainLayout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(mainFlex, 0, 1, true).
		AddItem(footer, 1, 0, false)

	// Pane focus cycling: 0=sidebar, 1=request, 2=response
	focusedPane := 0
	panes := []tview.Primitive{sidebar, requestPanel, responsePanel}
	borders := []borderSetter{sidebar, requestPanel, responsePanel}

	highlightBorders := func(active int) {
		for i, b := range borders {
			if i == active {
				b.SetBorderColor(postmanOrange)
			} else {
				b.SetBorderColor(tcell.ColorDefault)
			}
		}
	}

	// Apply initial highlight to the sidebar.
	highlightBorders(0)

	var pages *tview.Pages
	closeHelp := func() {
		pages.HidePage("help")
		app.SetFocus(panes[focusedPane])
	}
	helpModal := newHelpModal(closeHelp)

	setFocus := func(i int) {
		focusedPane = i
		highlightBorders(i)
		app.SetFocus(panes[i])
	}

	pages = tview.NewPages().
		AddPage("main", mainLayout, true, true).
		AddPage("help", centeredOverlay(helpModal, 50, 22), true, false)

	app.SetRoot(pages, true).SetFocus(sidebar)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		frontPage, _ := pages.GetFrontPage()

		if frontPage == "help" {
			switch {
			case event.Key() == tcell.KeyRune && event.Rune() == 'q',
				event.Key() == tcell.KeyRune && event.Rune() == '?':
				closeHelp()
				return nil
			}
			return event
		}

		switch {
		case event.Key() == tcell.KeyRune && event.Rune() == 'q':
			app.Stop()
			return nil
		case event.Key() == tcell.KeyRune && event.Rune() == '?':
			pages.ShowPage("help")
			app.SetFocus(helpModal)
			return nil
		case event.Key() == tcell.KeyTab:
			setFocus((focusedPane + 1) % len(panes))
			return nil
		case event.Key() == tcell.KeyBacktab:
			setFocus((focusedPane + len(panes) - 1) % len(panes))
			return nil
		case event.Key() == tcell.KeyRune && event.Rune() == 'S':
			setFocus(0)
			return nil
		case event.Key() == tcell.KeyRune && event.Rune() == 'R':
			setFocus(1)
			return nil
		case event.Key() == tcell.KeyRune && event.Rune() == 'P':
			setFocus(2)
			return nil
		}

		return event
	})

	return app
}
