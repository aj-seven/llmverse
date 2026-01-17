package ui

import (
	"strings"
	"time"

	"github.com/aj-seven/llmverse/internal/config"
	"github.com/aj-seven/llmverse/internal/history"
	aihub "github.com/aj-seven/llmverse/internal/providers/ollama"
	"github.com/aj-seven/llmverse/pkg/chat"
	"github.com/aj-seven/llmverse/pkg/messages"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

const (
	minInputLines = 1
	maxInputLines = 5
)

// Messages

type streamChunkMsg struct{ chunk string }
type streamDoneMsg struct{}
type startStreamMsg struct{}
type animationTickMsg struct{}

// Chat Model

type ChatModel struct {
	viewport viewport.Model
	textarea textarea.Model
	system   *SystemModel
	spinner  spinner.Model

	userStyle        lipgloss.Style
	botStyle         lipgloss.Style
	bubble           lipgloss.Style
	sendBtn          lipgloss.Style
	sendBtnDisabled  lipgloss.Style
	stopBtn          lipgloss.Style
	thinkingStyle    lipgloss.Style
	placeholderStyle lipgloss.Style
	promptStyle      lipgloss.Style
	animationStyle   lipgloss.Style
	bubbleFocused    lipgloss.Style
	bubbleUnfocused  lipgloss.Style

	modelName string
	back      bool

	historyManager *history.Manager
	cfg            *config.Config

	stream       <-chan string
	cancelStream chan struct{}
	streaming    bool

	width       int
	height      int
	maxMsgWidth int
	ready       bool
	lockScroll  bool

	animationStep int

	// click-to-focus geometry
	inputX int
	inputY int
	inputW int
	inputH int
}

// Constructor

func NewChatModel(
	modelName string,
	hm *history.Manager,
	cfg *config.Config,
) *ChatModel {

	ta := textarea.New()
	ta.Placeholder = "Start asking anything..."
	ta.Prompt = "❯ "
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(true)
	ta.Focus()
	ta.Cursor.Blink = true

	sp := spinner.New()
	sp.Spinner = spinner.MiniDot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))

	system := NewSystemModel(cfg)
	if system == nil {
		system = &SystemModel{}
	}

	return &ChatModel{
		modelName:      modelName,
		textarea:       ta,
		system:         system,
		historyManager: hm,
		cfg:            cfg,
		spinner:        sp,

		userStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("5")).Bold(true),

		botStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("6")).Bold(true),

		bubble: lipgloss.NewStyle().
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")),

		bubbleFocused: lipgloss.NewStyle().
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("7")),

		bubbleUnfocused: lipgloss.NewStyle().
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")),

		sendBtn: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("2")),

		sendBtnDisabled: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")),

		stopBtn: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("1")),

		thinkingStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Italic(true),

		placeholderStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Italic(true),

		promptStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("7")).
			Bold(true),

		animationStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("7")).
			Bold(true),
	}
}

// Init

func (m *ChatModel) Init() tea.Cmd {
	m.updateViewport(true)
	return m.FocusInput()
}

// Focus Helpers

func (m *ChatModel) FocusInput() tea.Cmd {
	m.textarea.Focus()
	return textarea.Blink
}

func (m *ChatModel) blurInput() {
	m.textarea.Blur()
}

func isTypingKey(msg tea.KeyMsg) bool {
	switch msg.Type {
	case tea.KeyRunes, tea.KeySpace:
		return true
	default:
		return false
	}
}

// Update

func (m *ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	systemModel, cmd := m.system.Update(msg)
	m.system = systemModel.(*SystemModel)
	cmds = append(cmds, cmd)

	if m.system.codePopup && !isWindowSizeMsg(msg) {
		return m, tea.Batch(cmds...)
	}

	m.spinner, cmd = m.spinner.Update(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {

	case tea.KeyMsg:
		if m.system.codePopup {
			break
		}

		if isTypingKey(msg) && !m.streaming && !m.textarea.Focused() {
			cmds = append(cmds, m.FocusInput())
		}

		switch msg.String() {

		case "ctrl+q":
			if m.historyManager != nil {
				m.historyManager.Close()
			}
			return m, tea.Quit

		case "ctrl+a":
			m.blurInput()
			if !m.streaming {
				m.system.Focus()
			}
			return m, nil

		case "esc":
			if m.streaming {
				cmd := m.stopStreaming()
				return m, cmd
			}
			return m, nil

		case "enter":
			if strings.TrimSpace(m.textarea.Value()) != "" && !m.streaming {
				return m.handleUserInput()
			}

		case "up", "down", "ctrl+up", "ctrl+down":
			m.blurInput()
			m.lockScroll = true
		}

	case tea.MouseMsg:
		switch msg.Type {

		case tea.MouseLeft:
		if msg.X >= m.inputX && msg.X < m.inputX+m.inputW &&
			msg.Y >= m.inputY && msg.Y < m.inputY+m.inputH {

			m.textarea.Focus()
			return m, textarea.Blink
		}

		m.textarea.Blur()
		return m, nil

		case tea.MouseWheelUp:
			m.blurInput()
			m.viewport.ScrollUp(3)

		case tea.MouseWheelDown:
			m.blurInput()
			m.viewport.ScrollDown(3)
		}

	case startStreamMsg:
		cmds = append(cmds, m.startStream())

	case streamChunkMsg:
		m.historyManager.UpdateAssistantMessage(msg.chunk)
		m.updateViewport(true)
		cmds = append(cmds, readStreamCmd(m.stream, m.cancelStream))

	case animationTickMsg:
		if m.streaming {
			m.animationStep++
			m.updateViewport(true)
			cmds = append(cmds, animationTick())
		}

	case streamDoneMsg:
		cmd := m.finishStream()
		cmds = append(cmds, cmd)
		cmds = append(cmds, func() tea.Msg {
			return messages.ChatCompletionMsg{}
		})
	}

	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)

	m.resizeTextarea()
	return m, tea.Batch(cmds...)
}

// View
func (m *ChatModel) View() string {
	if !m.ready {
		return "Loading..."
	}

	if systemView := m.system.View(); systemView != "" {
		return systemView
	}

	viewportView := m.viewport.View()
	inputView := m.renderInputRow()
	divider := dividerStyle.Render(strings.Repeat("─", m.width))

	dividerHeight := 1
	m.inputY = m.viewport.Height + dividerHeight + 2
	m.inputX = (m.width - m.maxMsgWidth)
	m.inputW = m.maxMsgWidth
	m.inputH = lipgloss.Height(inputView)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		viewportView,
		divider,
		inputView,
	)
}

// Input Row
func (m *ChatModel) renderInputRow() string {
	box := m.bubbleUnfocused
	if m.textarea.Focused() && !m.streaming {
		box = m.bubbleFocused
	}
	box = box.Width(m.maxMsgWidth)

	// Button (fixed width)
	btn := m.sendBtnDisabled.Render(" ➤ ")
	if m.streaming {
		btn = m.stopBtn.Render(" ● ")
	} else if strings.TrimSpace(m.textarea.Value()) != "" {
		btn = m.sendBtn.Render(" ➤ ")
	}
	btnWidth := lipgloss.Width(btn)

	// Content width = bubble width - button
	contentWidth := m.maxMsgWidth - btnWidth - 2 // padding safety

	var content string
	switch {
	case m.streaming:
		content = lipgloss.PlaceHorizontal(
			contentWidth,
			lipgloss.Left,
			m.spinner.View()+" "+m.thinkingStyle.Render("Thinking... (esc to stop)"),
		)
	case m.textarea.Value() == "":
		content = lipgloss.PlaceHorizontal(
			contentWidth,
			lipgloss.Left,
			m.promptStyle.Render(m.textarea.Prompt)+
				m.placeholderStyle.Render(m.textarea.Placeholder),
		)
	default:
		content = lipgloss.PlaceHorizontal(
			contentWidth,
			lipgloss.Left,
			m.textarea.View(),
		)
	}

	row := box.Render(
		lipgloss.JoinHorizontal(lipgloss.Top, content, btn),
	)

	// Center input bar on screen
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Center,
		row,
	)
}

// Helpers

func (m *ChatModel) resizeTextarea() {
	lines := strings.Count(m.textarea.Value(), "\n") + 1
	if lines < minInputLines {
		lines = minInputLines
	} else if lines > maxInputLines {
		lines = maxInputLines
	}
	m.textarea.SetHeight(lines)
}

func (m *ChatModel) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.maxMsgWidth = int(float64(w) * 0.98)

	// Input sizing
	m.textarea.SetWidth(w - 8)
	m.resizeTextarea()

	// Calculate the exact height of the input row to ensure correct viewport sizing
	inputRowHeight := lipgloss.Height(m.renderInputRow())

	// Viewport gets remaining space
	m.viewport.Width = w
	m.viewport.Height = h - inputRowHeight

	// System popup still overlays everything
	m.system.SetSize(w, h)

	m.ready = true
	m.updateViewport(true)
}

func (m *ChatModel) handleUserInput() (tea.Model, tea.Cmd) {
	input := strings.TrimSpace(m.textarea.Value())
	if input == "" {
		return m, nil
	}

	m.historyManager.AddUserMessage(input)

	m.streaming = true
	m.animationStep = 0
	m.lockScroll = false

	m.blurInput()
	m.textarea.Reset()
	m.updateViewport(true)

	return m, tea.Batch(
		m.spinner.Tick,
		func() tea.Msg { return startStreamMsg{} },
		animationTick(),
	)
}

func (m *ChatModel) updateViewport(forceBottom bool) {
	if !m.ready {
		return
	}
	m.viewport.SetContent(m.renderMessages())
	if forceBottom && !m.lockScroll {
		m.viewport.GotoBottom()
	}
}

func (m *ChatModel) stopStreaming() tea.Cmd {
	if m.cancelStream != nil {
		close(m.cancelStream)
		m.cancelStream = nil
	}
	m.streaming = false
	m.updateViewport(true)
	// Save the partial response
	m.historyManager.SaveCurrent()
	return m.FocusInput()
}

func (m *ChatModel) startStream() tea.Cmd {
	currentHistory := m.historyManager.GetCurrentHistory()
	if currentHistory == nil {
		return nil
	}
	// Exclude the last (empty) assistant message for the API call
	msgs := currentHistory.Messages[:len(currentHistory.Messages)-1]

	stream, err := aihub.StreamChat(m.modelName, msgs, m.cfg)
	if err != nil {
		return m.finishStream()
	}

	m.stream = stream
	m.cancelStream = make(chan struct{})
	return readStreamCmd(m.stream, m.cancelStream)
}

func (m *ChatModel) finishStream() tea.Cmd {
	m.streaming = false
	m.cancelStream = nil
	m.updateViewport(true)
	// Save the final response
	m.historyManager.SaveCurrent()
	return m.FocusInput()
}

// Rendering
var animationFrames = []string{`.`, `..`, `...`, `..`, `.`}

func (m *ChatModel) renderMessages() string {
	// Render gradient banner
	banner := RenderBanner(m.width, 8)

	// Centered plain text
	startText := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Render("Start chatting with " + m.modelName)

	currentHistory := m.historyManager.GetCurrentHistory()
	if currentHistory == nil || len(currentHistory.Messages) == 0 {

		// Combine banner + text
		content := lipgloss.JoinVertical(
			lipgloss.Center,
			banner,
			"",
			startText,
		)

		// Place in the MIDDLE of the screen
		return lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			content,
		)
	}

	// ---- Normal message rendering below ----

	var out []string
	messages := currentHistory.Messages

	for i, msg := range messages {
		style := m.bubble.Width(m.maxMsgWidth)

		if msg.Role == "user" {
			out = append(out,
				m.userStyle.Render("You")+"\n"+
					style.Render(msg.Content),
			)
			continue
		}

		content := msg.Content
		// Render Markdown
		renderedContent, err := glamour.Render(content, m.cfg.Theme.Markdown)
		if err != nil {
			renderedContent = content
		}

		if m.streaming && i == len(messages)-1 {
			renderedContent += " " + m.animationStyle.Render(
				animationFrames[m.animationStep%len(animationFrames)],
			)
		}

		out = append(out,
			m.botStyle.Render(m.modelName)+"\n"+
				style.Render(renderedContent),
		)
	}

	return strings.Join(out, "\n")
}

// Stream Cmd

func animationTick() tea.Cmd {
	return tea.Tick(150*time.Millisecond, func(t time.Time) tea.Msg {
		return animationTickMsg{}
	})
}

func readStreamCmd(stream <-chan string, cancel <-chan struct{}) tea.Cmd {
	return func() tea.Msg {
		var batch []string
		ticker := time.NewTicker(1 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-cancel:
				return streamDoneMsg{}
			case chunk, ok := <-stream:
				if !ok {
					if len(batch) > 0 {
						return streamChunkMsg{chunk: strings.Join(batch, "")}
					}
					return streamDoneMsg{}
				}
				batch = append(batch, chunk)
			case <-ticker.C:
				if len(batch) > 0 {
					msg := streamChunkMsg{chunk: strings.Join(batch, "")}
					batch = nil
					return msg
				}
			}
		}
	}
}

// Public API
func (m *ChatModel) Back() bool { return m.back }

func (m *ChatModel) ModelName() string { return m.modelName }

func (m *ChatModel) Messages() []chat.Message {
	if h := m.historyManager.GetCurrentHistory(); h != nil {
		return h.Messages
	}
	return nil
}

func (m *ChatModel) InputContent() string {
	return m.textarea.Value()
}

func isWindowSizeMsg(msg tea.Msg) bool {
	_, ok := msg.(tea.WindowSizeMsg)
	return ok
}
