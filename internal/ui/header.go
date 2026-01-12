package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type Header struct {
	width int
	title string

	// Main header content
	showMain bool

	// Toast
	showToast bool
}

func (h *Header) Height() int {
	if !h.showMain {
		return 1
	}
	return 3 // 1 row + bottom border
}

func NewHeader() *Header {
	return &Header{
		showMain: true, // header main is usually always visible
		title:    "llmverse",
	}
}

/* ───── Setters ───── */

func (h *Header) SetWidth(w int) {
	h.width = w
}

func (h *Header) SetTitle(title string) {
	h.title = title
}

func (h *Header) ShowMain(show bool) {
	h.showMain = show
}

func (h *Header) ShowToast(show bool) {
	h.showToast = show
}

// Header View

func (h *Header) View() string {
	if h.width <= 0 {
		return ""
	}

	borderColor := lipgloss.Color("238")

	outerBorder := lipgloss.NewStyle().
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(borderColor)

	container := lipgloss.NewStyle().
		Width(h.width).
		Padding(0, 1)

	innerWidth := h.width - 2

	// Main Header Row

	version := "(v0.1.0)" // TODO: Need to set version easy way
	name := "llmverse"
	author := "by AJ"

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("81")).
		Bold(true)

	versionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	authorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("7"))

	nameStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("81")).
		Bold(true)

	leftBlock := lipgloss.JoinHorizontal(
		lipgloss.Left,
		titleStyle.Render(h.title),
	)

	rightBlock := lipgloss.JoinHorizontal(
		lipgloss.Left,
		nameStyle.Render(name),
		" - ",
		authorStyle.Render(author),
		" ",
		versionStyle.Render(version),
	)

	space := innerWidth -
		lipgloss.Width(leftBlock) -
		lipgloss.Width(rightBlock)

	if space < 1 {
		space = 1
	}

	mainRow := lipgloss.NewStyle().
		Width(innerWidth).
		Render(
			lipgloss.JoinHorizontal(
				lipgloss.Top,
				leftBlock,
				strings.Repeat(" ", space),
				rightBlock,
			),
		)

	// Divider

	divider := lipgloss.NewStyle().
		Foreground(borderColor).
		Render(strings.Repeat("─", innerWidth))

	// Compose Rows Conditionally

	var rows []string

	if h.showMain {
		rows = append(rows, mainRow)
	}

	if h.showMain && h.showToast {
		rows = append(rows, divider)
	}

	if len(rows) == 0 {
		return ""
	}

	body := lipgloss.JoinVertical(lipgloss.Top, rows...)
	return outerBorder.Render(container.Render(body))
}
