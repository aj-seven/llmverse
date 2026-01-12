package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles

var (
	dimOverlayStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("0")).
			Foreground(lipgloss.Color("8"))

	dialogStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("6")).
			Padding(1, 3).
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("230"))

	btnStyle = lipgloss.NewStyle().
			Padding(0, 2)

	btnActiveStyle = btnStyle.
			Background(lipgloss.Color("57")).
			Foreground(lipgloss.Color("230"))
)

// ConfirmDialog types

type ConfirmDialog struct {
	Title   string
	Message string

	width  int
	height int

	cursor int // 0 = Yes, 1 = No
	Choice *bool
}

func (c *ConfirmDialog) SetSize(width int, height int) {
	panic("unimplemented")
}

// ConfirmDialog constructor
func NewConfirmDialog(title, msg string) *ConfirmDialog {
	return &ConfirmDialog{
		Title:   title,
		Message: msg,
		cursor:  0,
	}
}

// ConfirmDialog methods
// Update
func (c *ConfirmDialog) Update(msg tea.Msg) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		c.width = msg.Width
		c.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {

		case "left", "h", "up", "k":
			c.cursor = 0

		case "right", "l", "down", "j":
			c.cursor = 1

		case "enter":
			choice := c.cursor == 0
			c.Choice = &choice

		case "esc":
			f := false
			c.Choice = &f
		}
	}
}

// View

func (c *ConfirmDialog) View() string {
	// Dim background
	dim := dimOverlayStyle.
		Width(c.width).
		Height(c.height).
		Render("")

	// Buttons
	yes := btnStyle.Render(" Yes ")
	no := btnStyle.Render(" No ")

	if c.cursor == 0 {
		yes = btnActiveStyle.Render(" Yes ")
	} else {
		no = btnActiveStyle.Render(" No ")
	}

	buttons := lipgloss.JoinHorizontal(lipgloss.Center, yes, "  ", no)

	box := dialogStyle.Render(
		lipgloss.NewStyle().Bold(true).Render(c.Title) +
			"\n\n" +
			c.Message +
			"\n\n" +
			buttons,
	)

	dialog := lipgloss.Place(
		c.width,
		c.height,
		lipgloss.Center,
		lipgloss.Center,
		box,
	)

	return lipgloss.Place(
		c.width,
		c.height,
		lipgloss.Left,
		lipgloss.Top,
		dim+"\n"+dialog,
	)
}
