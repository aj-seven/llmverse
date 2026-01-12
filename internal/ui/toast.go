package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Messages

type ToastShowMsg struct {
	Text     string
	Duration time.Duration
}

type toastHideMsg struct{}

// Model

type Toast struct {
	visible  bool
	text     string
	width    int
	duration time.Duration
	style    lipgloss.Style
}

// Constructor

func NewToast() *Toast {
	return &Toast{
		width: 24,
		style: lipgloss.NewStyle().
			Padding(0, 1).
			Background(lipgloss.Color("4")),
	}
}

// Commands

func ShowToast(text string, d time.Duration) tea.Cmd {
	return func() tea.Msg {
		return ToastShowMsg{
			Text:     text,
			Duration: d,
		}
	}
}

func hideToastAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg {
		return toastHideMsg{}
	})
}

// Update

func (t *Toast) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {

	case ToastShowMsg:
		t.visible = true
		t.text = msg.Text
		t.duration = msg.Duration
		return hideToastAfter(msg.Duration)

	case toastHideMsg:
		t.visible = false
		t.text = ""
		return nil
	}

	return nil
}

// View

func (t *Toast) View() string {
	if !t.visible {
		return ""
	}

	content := t.style.Render("✔ " + t.text)

	return lipgloss.NewStyle().
		MarginRight(1).
		Render(content)
}

// Helpers

func (t *Toast) SetWidth(w int) {
	t.width = w
}

func (t *Toast) IsVisible() bool {
	return t.visible
}
