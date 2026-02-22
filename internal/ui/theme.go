package ui

import "github.com/charmbracelet/lipgloss"

// Theme defines the semantic color palette used across the UI.
type Theme struct {
	Name         string
	Accent       lipgloss.Color // active borders, cursor, focused elements
	Highlight    lipgloss.Color // titles, key labels
	Dimmed       lipgloss.Color // muted text, inactive elements
	Error        lipgloss.Color // errors, delete actions
	Text         lipgloss.Color // primary editable text
	TextMuted    lipgloss.Color // secondary values, read-only content
	MethodGET    lipgloss.Color
	MethodPOST   lipgloss.Color
	MethodPUT    lipgloss.Color
	MethodPATCH  lipgloss.Color
	MethodDELETE lipgloss.Color
}

func (t Theme) methodStyle(method string) (lipgloss.Style, bool) {
	var c lipgloss.Color
	switch method {
	case "GET":
		c = t.MethodGET
	case "POST":
		c = t.MethodPOST
	case "PUT":
		c = t.MethodPUT
	case "PATCH":
		c = t.MethodPATCH
	case "DELETE":
		c = t.MethodDELETE
	default:
		return lipgloss.NewStyle(), false
	}
	return lipgloss.NewStyle().Foreground(c).Bold(true), true
}

var themes = map[string]Theme{
	"rosepine":   themeRosePine,
	"xcode":      themeXcode,
	"catppuccin": themeCatppuccin,
	"tokyonight": themeTokyoNight,
	"sonokai":    themeSonokai,
}

// Rose Pine — warm purples and dusty rose on deep navy.
// Palette: https://github.com/rose-pine/neovim
var themeRosePine = Theme{
	Name:         "Rose Pine",
	Accent:       "#EB6F92", // love
	Highlight:    "#F6C177", // gold
	Dimmed:       "#6E6A86", // muted
	Error:        "#EB6F92", // love
	Text:         "#E0DEF4", // text
	TextMuted:    "#908CAA", // subtle
	MethodGET:    "#9CCFD8", // foam
	MethodPOST:   "#31748F", // pine
	MethodPUT:    "#F6C177", // gold
	MethodPATCH:  "#EBBCBA", // rose
	MethodDELETE: "#EB6F92", // love
}

// Xcode Dark — Apple's editor palette.
// Palette: https://github.com/fraeso/xcodedark.nvim
var themeXcode = Theme{
	Name:         "Xcode Dark",
	Accent:       "#6BDFFF", // type/cyan — signature Apple blue
	Highlight:    "#D9C97C", // number/yellow
	Dimmed:       "#6C7986", // comment
	Error:        "#FF5257", // cursor red
	Text:         "#FFFFFF",
	TextMuted:    "#D4D4D4", // fg_alt
	MethodGET:    "#67B7A4", // function_name/teal
	MethodPOST:   "#4EB0CC", // constant/blue
	MethodPUT:    "#D9C97C", // number/yellow
	MethodPATCH:  "#FD8F3F", // attribute/orange
	MethodDELETE: "#FF7AB2", // keyword/pink
}

// Catppuccin Mocha — soft pastel dark theme.
// Palette: https://github.com/catppuccin/nvim
var themeCatppuccin = Theme{
	Name:         "Catppuccin Mocha",
	Accent:       "#CBA6F7", // mauve — the catppuccin signature purple
	Highlight:    "#F9E2AF", // yellow
	Dimmed:       "#7F849C", // overlay1
	Error:        "#F38BA8", // red
	Text:         "#CDD6F4", // text
	TextMuted:    "#A6ADC8", // subtext0
	MethodGET:    "#A6E3A1", // green
	MethodPOST:   "#89B4FA", // blue
	MethodPUT:    "#F9E2AF", // yellow
	MethodPATCH:  "#FAB387", // peach
	MethodDELETE: "#F38BA8", // red
}

// Tokyo Night Storm — deep blue-black with electric accents.
// Palette: https://github.com/folke/tokyonight.nvim
var themeTokyoNight = Theme{
	Name:         "Tokyo Night",
	Accent:       "#7AA2F7", // blue — the iconic tokyo night blue
	Highlight:    "#E0AF68", // yellow
	Dimmed:       "#565F89", // comment
	Error:        "#F7768E", // red
	Text:         "#C0CAF5", // fg
	TextMuted:    "#A9B1D6", // fg_dark
	MethodGET:    "#9ECE6A", // green
	MethodPOST:   "#7AA2F7", // blue
	MethodPUT:    "#E0AF68", // yellow
	MethodPATCH:  "#FF9E64", // orange
	MethodDELETE: "#F7768E", // red
}

// Sonokai — vibrant colors on a dark grey base.
// Palette: https://github.com/sainnhe/sonokai
var themeSonokai = Theme{
	Name:         "Sonokai",
	Accent:       "#9ED072", // green — sonokai's classic green
	Highlight:    "#E7C664", // yellow
	Dimmed:       "#7F8490", // grey
	Error:        "#FC5D7C", // red
	Text:         "#E2E2E3", // fg
	TextMuted:    "#B0B0B2", // between fg and grey
	MethodGET:    "#9ED072", // green
	MethodPOST:   "#76CCE0", // blue
	MethodPUT:    "#E7C664", // yellow
	MethodPATCH:  "#F39660", // orange
	MethodDELETE: "#FC5D7C", // red
}
