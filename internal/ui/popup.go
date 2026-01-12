package ui

import "github.com/charmbracelet/lipgloss"

type Popup struct {
	title   string
	content string
	width   int
}

func NewPopup(title, content string, width int) *Popup {
	return &Popup{
		title:   title,
		content: content,
		width:   width,
	}
}

func (p *Popup) View() string {
	style := lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("6")).
		Width(p.width)

	header := lipgloss.NewStyle().
		Bold(true).
		Render(p.title)

	body := lipgloss.NewStyle().
		Foreground(lipgloss.Color("7")).
		Render(p.content)

	return style.Render(header + "\n\n" + body)
}

func popupCentered(w, h int, content string) string {
	return lipgloss.Place(
		w,
		h,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}
