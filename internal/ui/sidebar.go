package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type collection struct {
	name     string
	requests []request
}

type request struct {
	method string
	name   string
}

var mockCollections = []collection{
	{
		name: "Examples",
		requests: []request{
			{method: "GET", name: "Httpbin GET"},
			{method: "POST", name: "Httpbin POST"},
		},
	},
}

var methodColors = map[string]tcell.Color{
	"GET":    tcell.ColorGreen,
	"POST":   tcell.ColorBlue,
	"PUT":    tcell.Color214,
	"PATCH":  tcell.Color208,
	"DELETE": tcell.ColorRed,
}

func collectionLabel(name string, expanded bool) string {
	if expanded {
		return "▾ " + name
	}
	return "▸ " + name
}

func newSidebar() *tview.TreeView {
	root := tview.NewTreeNode("").SetSelectable(false)

	for _, col := range mockCollections {
		colNode := tview.NewTreeNode(collectionLabel(col.name, true)).
			SetColor(tcell.ColorYellow).
			SetExpanded(true).
			SetReference(col.name)

		for _, req := range col.requests {
			color, ok := methodColors[req.method]
			if !ok {
				color = tcell.ColorWhite
			}
			label := req.method + "  " + req.name
			reqNode := tview.NewTreeNode(label).SetColor(color)
			colNode.AddChild(reqNode)
		}

		root.AddChild(colNode)
	}

	tree := tview.NewTreeView().
		SetRoot(root).
		SetCurrentNode(root)

	tree.SetBorder(true).SetTitle(" Collections [green](S)[-] ")

	tree.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune && event.Rune() == ' ' {
			node := tree.GetCurrentNode()
			if node != nil && len(node.GetChildren()) > 0 {
				expanded := !node.IsExpanded()
				node.SetExpanded(expanded)
				if name, ok := node.GetReference().(string); ok {
					node.SetText(collectionLabel(name, expanded))
				}
				return nil
			}
		}
		return event
	})

	return tree
}
