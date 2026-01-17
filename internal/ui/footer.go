package ui

import (
	"strings"

	"github.com/aj-seven/llmverse/pkg/keymap"
	"github.com/charmbracelet/lipgloss"
)

type Footer struct {
	width int
	toast *Toast

	// Content
	primaryContent   string
	secondaryContent string
	showContent      bool

	// Shortcuts
	shortcuts     []keymap.Shortcut
	showShortcuts bool
}

func newFooter(toast *Toast, shortcuts ...keymap.Shortcut) *Footer {
	return &Footer{
		toast:     toast,
		shortcuts: shortcuts,
	}
}

// Layout 

func (f *Footer) Height() int {
	if f.width <= 0 {
		return 0
	}

	height := 1 // top border

	if f.showContent {
		height++
	}

	if f.showContent && f.showShortcuts {
		height++ // divider
	}

	if f.showShortcuts {
		height++
	}

	return height
}

// Setters

func (f *Footer) SetWidth(w int) {
	f.width = w
}

func (f *Footer) SetContent(primary, secondary string) {
	f.primaryContent = primary
	f.secondaryContent = secondary
}

func (f *Footer) SetShortcuts(shortcuts ...keymap.Shortcut) {
	f.shortcuts = shortcuts
}

func (f *Footer) ShowContent(show bool) {
	f.showContent = show
}

func (f *Footer) ShowShortcuts(show bool) {
	f.showShortcuts = show
}

// View

func (f *Footer) View() string {
	if f.width <= 0 {
		return ""
	}

	borderColor := lipgloss.Color("238")

	outerBorder := lipgloss.NewStyle().
		BorderTop(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(borderColor)

	container := lipgloss.NewStyle().
		Width(f.width).
		Padding(0, 1)

	innerWidth := f.width - 2

	// Content Row 

	primaryStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Bold(true)

	secondaryStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))

	// primary content (Model)
	modelContent := f.primaryContent
	modelLabel := ""
	modelName := ""
	if strings.HasPrefix(modelContent, "Model: ") {
		modelLabel = "Model: "
		modelName = strings.TrimPrefix(modelContent, "Model: ")
	} else {
		modelName = modelContent
	}

	// secondary content (System Message)
	sysMsgContent := f.secondaryContent
	sysMsgLabel := ""
	sysMsgIndicator := ""
	if strings.HasPrefix(sysMsgContent, "System Message: ") {
		sysMsgLabel = "System Message: "
		sysMsgIndicator = strings.TrimPrefix(sysMsgContent, "System Message: ")
	}

	var sysMsgStatus string
	var indicatorStyle lipgloss.Style
	if strings.TrimSpace(sysMsgIndicator) == "✔" {
		sysMsgStatus = "Available"
		indicatorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
	} else if strings.TrimSpace(sysMsgIndicator) == "✖" {
		sysMsgStatus = "NA"
		indicatorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
	}

	leftContent := lipgloss.JoinHorizontal(
		lipgloss.Left,
		secondaryStyle.Render(modelLabel),
		primaryStyle.Render(modelName),
		"  |  ", // separator
		secondaryStyle.Render(sysMsgLabel),
		" ",
		primaryStyle.Render(sysMsgStatus),
		" ",
		indicatorStyle.Render(strings.TrimSpace(sysMsgIndicator)),
	)

	rightWidth := innerWidth - lipgloss.Width(leftContent)
	if rightWidth < 0 {
		rightWidth = 0
	}

	rightContent := lipgloss.NewStyle().
		Width(rightWidth).
		Align(lipgloss.Right).
		Render(f.toast.View())

	contentRow := lipgloss.JoinHorizontal(lipgloss.Bottom, leftContent, rightContent)

	// Divider 

	divider := lipgloss.NewStyle().
		Foreground(borderColor).
		Render(strings.Repeat("─", innerWidth))

	// Shortcuts Rows

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("62")).
		Padding(0, 1).
		Bold(true)

	actionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))

	itemStyle := lipgloss.NewStyle().
		Padding(0, 1)

	items := make([]string, 0, len(f.shortcuts))
	for _, s := range f.shortcuts {
		item := lipgloss.JoinHorizontal(
			lipgloss.Left,
			keyStyle.Render(strings.ToUpper(s.Key)),
			" ",
			actionStyle.Render(s.Action),
		)
		items = append(items, itemStyle.Render(item))
	}

	shortcutsRow := lipgloss.NewStyle().
		Width(innerWidth).
		Align(lipgloss.Center).
		Render(lipgloss.JoinHorizontal(lipgloss.Center, items...))

	// Compose 

	var rows []string

	if f.showContent {
		rows = append(rows, contentRow)
	}

	if f.showContent && f.showShortcuts {
		rows = append(rows, divider)
	}

	if f.showShortcuts {
		rows = append(rows, shortcutsRow)
	}

	if len(rows) == 0 {
		return ""
	}

	body := lipgloss.JoinVertical(lipgloss.Top, rows...)
	return outerBorder.Render(container.Render(body))
}
