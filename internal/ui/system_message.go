package ui

import (
	"time"

	"github.com/aj-seven/llmverse/internal/config"
	messages "github.com/aj-seven/llmverse/pkg/messages"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

// System Model
type SystemModel struct {
	codePopup    bool
	codeTextarea textarea.Model
	width        int
	height       int
	config       *config.Config

	toastMessage string
	toastUntil   time.Time
}

// Constructor
func NewSystemModel(cfg *config.Config) *SystemModel {
	ta := textarea.New()
	ta.Placeholder = "Input System message here..."
	ta.Prompt = ""
	ta.ShowLineNumbers = true
	ta.SetHeight(10)
	ta.SetWidth(60)
	ta.Blur()

	return &SystemModel{
		codeTextarea: ta,
		config:       cfg,
	}
}

// Init
func (m *SystemModel) Init() tea.Cmd {
	return nil
}

// Update
func (m *SystemModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !m.codePopup {
		return m, nil
	}

	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "esc":
			m.codePopup = false
			m.codeTextarea.Blur()
			return m, func() tea.Msg {
				return messages.SystemPopupStatusMsg{IsOpen: false}
			}
		case "ctrl+s":
			return m.saveSystemMessage()
		}
	}

	var cmd tea.Cmd
	m.codeTextarea, cmd = m.codeTextarea.Update(msg)
	return m, cmd
}

// View
func (m *SystemModel) View() string {
	if !m.codePopup {
		return ""
	}

	body := m.codeTextarea.View() + "\n\nctrl+s = save • esc = cancel"

	// Render toast if active
	if m.toastMessage != "" && time.Now().Before(m.toastUntil) {
		body += "\n\n✓ " + m.toastMessage
	} else {
		m.toastMessage = ""
	}

	popup := NewPopup(
		"System Message",
		body,
		64,
	)

	return popupCentered(m.width, m.height, popup.View())
}

// Layout

func (m *SystemModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// Focus
func (m *SystemModel) Focus() tea.Cmd {
	m.codePopup = true

	if m.config != nil {
		m.codeTextarea.SetValue(m.config.Assistant.Message)
	} else {
		m.codeTextarea.SetValue("")
	}

	m.codeTextarea.CursorEnd()
	m.codeTextarea.Focus()

	return func() tea.Msg {
		return messages.SystemPopupStatusMsg{IsOpen: true}
	}
}

// Helpers
func (m *SystemModel) saveSystemMessage() (tea.Model, tea.Cmd) {
	if m.config == nil {
		return m, nil
	}

	code := m.codeTextarea.Value()
	_ = m.config.SetSystemMessage(code)

	return m, tea.Batch(
		ShowToast("System Message saved.", 2*time.Second),
		func() tea.Msg {
			return messages.SystemPopupStatusMsg{IsOpen: m.codePopup}
		},
	)
}
