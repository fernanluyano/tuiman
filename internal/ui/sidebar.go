package ui

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

type sidebarItemKind int

const (
	itemCollection sidebarItemKind = iota
	itemRequest
)

type sidebarItem struct {
	kind    sidebarItemKind
	colName string  // itemCollection
	req     request // itemRequest
}

func buildItems(cols []collection, expanded map[string]bool) []sidebarItem {
	var items []sidebarItem
	for _, c := range cols {
		items = append(items, sidebarItem{kind: itemCollection, colName: c.name})
		if expanded[c.name] {
			for _, r := range c.requests {
				items = append(items, sidebarItem{kind: itemRequest, req: r})
			}
		}
	}
	return items
}
