package ui

import (
	"fmt"
	"strings"

	"github.com/aj-seven/llmverse/internal/history"
	messages "github.com/aj-seven/llmverse/pkg/messages"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles

var (
	rowStyle = lipgloss.NewStyle().
		PaddingLeft(1)

	selectedRowStyle = lipgloss.NewStyle().
		PaddingLeft(1).
		Background(lipgloss.Color("57")).
		Foreground(lipgloss.Color("230"))

	dimStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("8"))
)

// History Model types

type HistoryModel struct {
	histories []history.History
	cursor    int

	selectedHistoryID string
	deletedHistoryID  string

	confirm *ConfirmDialog

	width  int
	height int
}

// Constructor
func NewHistoryModel(h []history.History) *HistoryModel {
	return &HistoryModel{
		histories: append([]history.History{}, h...),
	}
}

func (m *HistoryModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *HistoryModel) Init() tea.Cmd { return nil }

// Update

func (m *HistoryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = ws.Width
		m.height = ws.Height
		if m.confirm != nil {
			m.confirm.Update(ws)
		}
		return m, nil
	}

	if m.confirm != nil {
		m.confirm.Update(msg)

		if m.confirm.Choice != nil {
			if *m.confirm.Choice && len(m.histories) > 0 {
				m.deletedHistoryID = m.histories[m.cursor].ID
				m.deleteSelected()
			}
			m.confirm = nil
		}
		return m, nil
	}

	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {

		case "esc":
			return m, func() tea.Msg {
				return messages.GoBackMsg{}
			}

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.histories)-1 {
				m.cursor++
			}

		case "enter":
			if len(m.histories) > 0 {
				m.selectedHistoryID = m.histories[m.cursor].ID
				return m, func() tea.Msg {
					return messages.GoBackMsg{}
				}
			}

		case "ctrl+d":
			if len(m.histories) > 0 {
				m.confirm = NewConfirmDialog(
					"Delete chat?",
					"This action cannot be undone.",
				)
				m.confirm.width = m.width
				m.confirm.height = m.height
			}
		}
	}

	return m, nil
}

// View

func (m *HistoryModel) View() string {
	var b strings.Builder

	if len(m.histories) == 0 {
		b.WriteString(dimStyle.Render("No saved chats."))
	} else {
		// Header
		b.WriteString(m.renderHeader())
		b.WriteString("\n")
		availableHeight := m.height
		maxRows := max(1, availableHeight-4)

		start := max(0, m.cursor-maxRows+1)
		end := min(len(m.histories), start+maxRows)

		for i := start; i < end; i++ {
			row := m.renderRow(i, m.histories[i])
			if i == m.cursor {
				b.WriteString(selectedRowStyle.Render(row))
			} else {
				b.WriteString(rowStyle.Render(row))
			}
			b.WriteString("\n")
		}
	}

	view := b.String()

	if m.confirm != nil {
		confirmView := m.confirm.View()
		return lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			confirmView,
		)
	}

	return view
}

// Helpers
func (m *HistoryModel) renderHeader() string {
	indexW := 4
	modelW := 12
	dateW := 18
	sepW := 3 * 3

	used := indexW + modelW + dateW + sepW
	titleW := max(10, m.width-used)

	return lipgloss.NewStyle().
		Bold(true).
		Render(fmt.Sprintf(
			" %*s │ %-*s │ %-*s │ %-*s",
			indexW-1, "SNO",
			modelW, "MODEL",
			dateW, "DATE",
			titleW, "TITLE",
		))
}

func (m *HistoryModel) renderRow(i int, h history.History) string {
	indexW := 4
	modelW := 12
	dateW := 18
	sepW := 3 * 3

	used := indexW + modelW + dateW + sepW
	titleW := max(10, m.width-used)

	title := deriveTitle(h)
	title = truncate(title, titleW)

	date := h.UpdatedAt.Format("2006-01-02 15:04")

	return fmt.Sprintf(
		"%*d │ %-*s │ %-*s │ %-*s",
		indexW-1, i+1,
		modelW, h.Model,
		dateW, date,
		titleW, title,
	)
}

func deriveTitle(h history.History) string {
	if strings.TrimSpace(h.Title) != "" {
		return truncate(h.Title, 50)
	}

	if len(h.Messages) > 0 {
		return truncate(h.Messages[0].Content, 50)
	}

	return "Empty Chat"
}

func (m *HistoryModel) deleteSelected() {
	if len(m.histories) == 0 {
		return
	}

	m.histories = append(
		m.histories[:m.cursor],
		m.histories[m.cursor+1:]...,
	)

	if m.cursor >= len(m.histories) && m.cursor > 0 {
		m.cursor--
	}
}

func truncate(s string, max int) string {
	s = strings.TrimSpace(strings.ReplaceAll(s, "\n", " "))
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

// Public API

func (m *HistoryModel) SelectedHistoryID() string {
	id := m.selectedHistoryID
	m.selectedHistoryID = ""
	return id
}

func (m *HistoryModel) DeletedID() string {
	id := m.deletedHistoryID
	m.deletedHistoryID = ""
	return id
}
