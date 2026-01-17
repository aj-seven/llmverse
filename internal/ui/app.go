package ui

import (
	"fmt"
	"time"

	"github.com/aj-seven/llmverse/internal/config"
	"github.com/aj-seven/llmverse/internal/history"
	"github.com/aj-seven/llmverse/pkg/keymap"
	aihub "github.com/aj-seven/llmverse/internal/providers/ollama"
	"github.com/aj-seven/llmverse/pkg/messages"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Views

type View int

const (
	ChatView View = iota
	HistoryView
	ModelSelectionView
)

// Root Model

type Model struct {
	view      View
	viewStack []View

	width  int
	height int

	header         *Header
	footer         *Footer
	toast          *Toast
	chat           *ChatModel
	history        *HistoryModel
	modelSelection *ModelSelection

	models         []aihub.OllamaModel
	historyManager *history.Manager
	currentModel   string

	lastWS tea.WindowSizeMsg
	hasWS  bool

	cfg *config.Config
}

// Constructor

func New(
	models []aihub.OllamaModel,
	historyManager *history.Manager,
	cfg *config.Config,
) *Model {

	toast := NewToast()

	m := &Model{
		view:           ChatView,
		viewStack:      []View{ChatView},
		header:         NewHeader(),
		footer:         newFooter(toast),
		toast:          toast,
		models:         models,
		historyManager: historyManager,
		currentModel:   models[0].Name,
		cfg:            cfg,
	}

	m.newChat(m.currentModel, "")
	m.updateFooterContent()

	return m
}

// Init 

func (m *Model) Init() tea.Cmd {
	return m.chat.Init()
}

// Update
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Toast always first
	if cmd := m.toast.Update(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	switch msg := msg.(type) {

	// WINDOW RESIZE

	case tea.WindowSizeMsg:
		if msg.Width == m.width && msg.Height == m.height {
			break
		}

		m.width = msg.Width
		m.height = msg.Height
		m.lastWS = msg
		m.hasWS = true

		m.header.SetWidth(msg.Width)
		m.footer.SetWidth(msg.Width)
		m.toast.SetWidth(msg.Width)

		m.updateFooterContent()
		m.applyLayout()

	// MODEL SELECTED

	case ModelSelectedMsg:
		m.currentModel = msg.Name
		m.newChat(m.currentModel, "")
		m.view = ChatView
		m.updateFooterContent()
		m.applyLayout()

		return m, ShowToast(
			fmt.Sprintf("Model changed to %s", msg.Name),
			2*time.Second,
		)

	// SYSTEM POPUPS

	case messages.SystemPopupStatusMsg:
		m.updateFooterContent()
		m.applyLayout()
		return m, nil

	// VIEW STACK

	case messages.PushViewMsg:
		m.viewStack = append(m.viewStack, m.view)
		m.view = View(msg.View)
		m.updateFooterContent()
		m.applyLayout()
		return m, nil

	case messages.GoBackMsg:
		if len(m.viewStack) > 1 {
			m.view = m.viewStack[len(m.viewStack)-1]
			m.viewStack = m.viewStack[:len(m.viewStack)-1]
			m.updateFooterContent()
			m.applyLayout()

			if m.view == ChatView {
				return m, m.chat.FocusInput()
			}
			return m, nil
		}
		return m, tea.Quit
	}

	//GLOBAL KEYS 

	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {

		case "ctrl+q":
			m.historyManager.Close()
			return m, tea.Quit

		case "ctrl+h":
			histories, _ := m.historyManager.GetAllHistories()
			m.history = NewHistoryModel(histories)
			m.applyLayout()
			return m, func() tea.Msg {
				return messages.PushViewMsg{View: int(HistoryView)}
			}

		case "ctrl+o":
			m.modelSelection = NewModelSelection(m.models)
			m.applyLayout()
			return m, func() tea.Msg {
				return messages.PushViewMsg{View: int(ModelSelectionView)}
			}
		}
	}

	// ACTIVE VIEW UPDATE

	var cmd tea.Cmd

	switch m.view {

	case ChatView:
		var model tea.Model
		model, cmd = m.chat.Update(msg)
		m.chat = model.(*ChatModel)
		cmds = append(cmds, cmd)

	case HistoryView:
		var model tea.Model
		model, cmd = m.history.Update(msg)
		m.history = model.(*HistoryModel)
		cmds = append(cmds, cmd)

		if id := m.history.DeletedID(); id != "" {
			m.historyManager.DeleteHistory(id)
			histories, _ := m.historyManager.GetAllHistories()
			m.history = NewHistoryModel(histories)
			m.applyLayout()
			return m, ShowToast("Chat deleted", 2*time.Second)
		}

		if id := m.history.SelectedHistoryID(); id != "" {
			m.newChat("", id)
			m.view = ChatView
			m.updateFooterContent()
			m.applyLayout()

			h := m.historyManager.GetCurrentHistory()
			return m, ShowToast(
				fmt.Sprintf(
					"Loaded chat from %s",
					h.UpdatedAt.Format("2006-01-02 15:04"),
				),
				2*time.Second,
			)
		}

	case ModelSelectionView:
		var model tea.Model
		model, cmd = m.modelSelection.Update(msg)
		m.modelSelection = model.(*ModelSelection)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View Model

func (m *Model) View() string {
	var content string

	switch m.view {
	case ChatView:
		content = m.chat.View()
	case HistoryView:
		content = m.history.View()
	case ModelSelectionView:
		content = m.modelSelection.View()
	}

	header := m.header.View()
	footer := m.footer.View()

	availableHeight :=
		m.height -
			lipgloss.Height(header) -
			lipgloss.Height(footer)

	content = lipgloss.
		NewStyle().
		Height(availableHeight).
		Render(content)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		content,
		footer,
	)
}

// Helpers

func (m *Model) contentHeight() int {
	if !m.hasWS {
		return 0
	}
	return m.lastWS.Height -
		m.header.Height() -
		lipgloss.Height(m.footer.View())
}

func (m *Model) applyLayout() {
	if !m.hasWS {
		return
	}

	w := m.lastWS.Width
	h := m.contentHeight()

	if m.chat != nil {
		m.chat.SetSize(w, h)
	}
	if m.history != nil {
		m.history.SetSize(w, h)
	}
	if m.modelSelection != nil {
		m.modelSelection.SetSize(w, h)
	}
}

func (m *Model) newChat(modelName, historyID string) {
	if historyID != "" {
		h, err := m.historyManager.LoadHistory(historyID)
		if err != nil {
			m.historyManager.NewHistory(m.currentModel)
		} else {
			m.currentModel = h.Model
		}
	} else {
		m.historyManager.NewHistory(modelName)
	}

	m.chat = NewChatModel(
		m.currentModel,
		m.historyManager,
		m.cfg,
	)
}

func (m *Model) isSystemMessageSet() bool {
	return m.cfg != nil && m.cfg.Assistant.Message != ""
}

// Footer
func (m *Model) updateFooterContent() {
	m.footer.ShowContent(false)
	m.footer.ShowShortcuts(false)

	switch m.view {

	case ChatView:
		m.header.SetTitle("Chat")

		sysMsgIndicator := " ✖"
		if m.isSystemMessageSet() {
			sysMsgIndicator = " ✔"
		}

		m.footer.SetContent(
			fmt.Sprintf("Model: %s", m.currentModel),
			fmt.Sprintf("System Message: %s", sysMsgIndicator),
		)
		m.footer.SetShortcuts(
			keymap.Shortcut{Key: "ctrl+o", Action: "Models"},
			keymap.Shortcut{Key: "ctrl+h", Action: "History"},
			keymap.Shortcut{Key: "ctrl+a", Action: "System Message"},
			keymap.Shortcut{Key: "ctrl+q", Action: "Quit"},
		)
		m.footer.ShowContent(true)
		m.footer.ShowShortcuts(true)

	case HistoryView:
		m.header.SetTitle("Chat History")
		m.footer.SetShortcuts(
			keymap.Shortcut{Key: "↑/↓", Action: "Navigate"},
			keymap.Shortcut{Key: "enter", Action: "Open"},
			keymap.Shortcut{Key: "ctrl+d", Action: "Delete"},
			keymap.Shortcut{Key: "esc", Action: "Back"},
			keymap.Shortcut{Key: "ctrl+q", Action: "Quit"},
		)
		m.footer.ShowShortcuts(true)

	case ModelSelectionView:
		m.header.SetTitle("Model Selection")
		m.footer.SetShortcuts(
			keymap.Shortcut{Key: "↑/↓", Action: "Navigate"},
			keymap.Shortcut{Key: "enter", Action: "Select"},
			keymap.Shortcut{Key: "esc", Action: "Back"},
			keymap.Shortcut{Key: "ctrl+q", Action: "Quit"},
		)
		m.footer.ShowShortcuts(true)
	}
}