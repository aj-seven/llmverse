package ui

import (
	"github.com/aj-seven/llmverse/pkg/messages"
	aihub "github.com/aj-seven/llmverse/internal/providers/ollama"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles

var (
	modelRowStyle = lipgloss.NewStyle().
		PaddingLeft(1)

	modelSelectedRowStyle = lipgloss.NewStyle().
		PaddingLeft(0).
		Background(lipgloss.Color("57")).
		Foreground(lipgloss.Color("230"))

	modelDimStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("8"))

	dividerStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("238"))

	infoBoxStyle = lipgloss.NewStyle().
		Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("7"))

	infoTitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("6")).
		Align(lipgloss.Center)

	infoLabelStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("8"))

	infoValueStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("230"))
)

// Messages

type ModelSelectedMsg aihub.OllamaModel
type ModelSelectionBackMsg struct{}

// Model

type ModelSelection struct {
	models []aihub.OllamaModel
	cursor int

	width  int
	height int
}

func NewModelSelection(models []aihub.OllamaModel) *ModelSelection {
	return &ModelSelection{models: models}
}

func (m *ModelSelection) Init() tea.Cmd { return nil }

func (m *ModelSelection) SetSize(w, h int) {
	m.width = w
	m.height = h
}

// Update

func (m *ModelSelection) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.models)-1 {
				m.cursor++
			}

		case "esc":
			return m, func() tea.Msg {
				return messages.GoBackMsg{}
			}

		case "enter":
			if len(m.models) > 0 {
				selected := m.models[m.cursor]
				return m, func() tea.Msg {
					return ModelSelectedMsg(selected)
				}
			}
		}
	}

	return m, nil
}

// View

func (m *ModelSelection) View() string {
	if m.width == 0 {
		return "Loading models…"
	}

	infoBoxHeight := 9
	dividerHeight := 1
	listHeight := m.height - infoBoxHeight - dividerHeight

	list := m.renderList(listHeight)
	divider := dividerStyle.Render(strings.Repeat("─", m.width))
	info := m.renderInfoBox(infoBoxHeight)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		list,
		divider,
		info,
	)
}

const visibleRows = 5

// List Rendering

func (m *ModelSelection) renderList(_ int) string {
	var b strings.Builder

	total := len(m.models)
	if total == 0 {
		b.WriteString(modelDimStyle.Render(" No models found"))
		return b.String()
	}
	
	// Header
	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	nameW := max(12, m.width-2)

	// Fixed 5-row scroll window
	start := m.cursor - visibleRows/2
	if start < 0 {
		start = 0
	}
	if start+visibleRows > total {
		start = max(0, total-visibleRows)
	}
	end := min(total, start+visibleRows)

	for i := start; i < end; i++ {
		model := m.models[i]

		row := fmt.Sprintf(
			" %-*s",
			nameW,
			trim(model.Name, nameW),
		)

		if i == m.cursor {
			b.WriteString(
				modelSelectedRowStyle.
					Width(m.width).
					Render(row),
			)
		} else {
			b.WriteString(
				modelRowStyle.
					Width(m.width).
					Render(row),
			)
		}
		b.WriteString("\n")
	}

	return b.String()
}

func (m *ModelSelection) renderHeader() string {
	nameW := max(12, m.width-2)

	return lipgloss.NewStyle().
		Bold(true).
		Render(fmt.Sprintf(
			" %-*s",
			nameW,
			"NAME",
		))
}

// Info Box Rendering

func (m *ModelSelection) renderInfoBox(height int) string {
	if m.cursor < 0 || m.cursor >= len(m.models) {
		return ""
	}

	model := m.models[m.cursor]

	body := fmt.Sprintf(
		"%s %s\n%s %s\n%s %s\n%s %s\n%s %s",
		infoLabelStyle.Render("Family:"),
		infoValueStyle.Render(model.Details.Family),

		infoLabelStyle.Render("Parameters:"),
		infoValueStyle.Render(model.Details.ParameterSize),

		infoLabelStyle.Render("Quantization:"),
		infoValueStyle.Render(model.Details.QuantizationLevel),

		infoLabelStyle.Render("Size:"),
		infoValueStyle.Render(formatSize(model.Size)),

		infoLabelStyle.Render("Modified:"),
		infoValueStyle.Render(model.ModifiedAt.Format("2006-01-02 15:04")),
	)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		infoTitleStyle.Render("Model Info"),
		"",
		body,
	)

	return infoBoxStyle.
		Width(m.width - 2).
		Height(height).
		Render(content)
}

// Helpers

func trim(s string, w int) string {
	if lipgloss.Width(s) <= w {
		return s
	}
	if w <= 1 {
		return s[:w]
	}
	return s[:w-1] + "…"
}

func formatSize(bytes int64) string {
	const gb = 1024 * 1024 * 1024
	return fmt.Sprintf("%.1fG", float64(bytes)/gb)
}
