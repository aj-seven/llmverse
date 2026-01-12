package history

import (
	"sync"
	"time"

	"github.com/aj-seven/llmverse/pkg/chat"

	"github.com/google/uuid"
)

// Manager is responsible for managing chat histories in memory
// and coordinating with a Storage backend for persistence.
// All history modifications should go through the Manager.
type Manager struct {
	storage Storage
	mu      sync.RWMutex

	currentHistory *History
}

// NewManager creates a new history manager.
func NewManager(storage Storage) *Manager {
	return &Manager{
		storage: storage,
	}
}

// Close persists the current chat session to disk.
func (m *Manager) Close() {
	if m == nil {
		return
	}
	m.SaveCurrent()
}

// SaveCurrent explicitly saves the current history to storage if it exists.
func (m *Manager) SaveCurrent() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.currentHistory == nil || len(m.currentHistory.Messages) == 0 {
		return
	}
	_ = m.storage.SaveHistory(*m.currentHistory)
}

// NewHistory creates a new, empty chat history.
// It saves the previously active history before creating a new one.
func (m *Manager) NewHistory(model string) *History {
	m.SaveCurrent() // Save the old history first.

	m.mu.Lock()
	defer m.mu.Unlock()

	h := &History{
		ID:        uuid.New().String(),
		Model:     model,
		Messages:  []chat.Message{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	m.currentHistory = h
	// Immediately save the new empty history to ensure it's listed.
	_ = m.storage.SaveHistory(*m.currentHistory)
	return h
}

// LoadHistory loads a history from storage and sets it as the current one.
// It saves the previously active history before loading a new one.
func (m *Manager) LoadHistory(id string) (*History, error) {
	m.SaveCurrent() // Save the old history first.

	m.mu.Lock()
	defer m.mu.Unlock()

	h, err := m.storage.GetHistory(id)
	if err != nil {
		return nil, err
	}

	m.currentHistory = &h
	return m.currentHistory, nil
}

// GetCurrentHistory returns a pointer to the current in-memory history.
func (m *Manager) GetCurrentHistory() *History {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentHistory
}

// GetAllHistories returns all histories from the storage.
func (m *Manager) GetAllHistories() ([]History, error) {
	// Ensure the current session is saved so it appears in the list.
	m.SaveCurrent()
	return m.storage.GetHistories()
}

// DeleteHistory removes a history from storage. If it's the current
// history, a new empty one is created.
func (m *Manager) DeleteHistory(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	err := m.storage.DeleteHistory(id)
	if err != nil {
		return err
	}

	if m.currentHistory != nil && m.currentHistory.ID == id {
		// If we deleted the current chat, start a new one.
		h := &History{
			ID:        uuid.New().String(),
			Model:     m.currentHistory.Model, // Keep the same model
			Messages:  []chat.Message{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		m.currentHistory = h
	}
	return nil
}

// AddUserMessage adds a user message to the current history and prepares
// an empty response from the assistant.
func (m *Manager) AddUserMessage(content string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.currentHistory == nil {
		return
	}

	// If this is the first user message, set it as the title.
	if len(m.currentHistory.Messages) == 0 {
		m.currentHistory.Title = content
	}

	m.currentHistory.Messages = append(m.currentHistory.Messages, chat.Message{
		Role:    "user",
		Content: content,
	})
	// Add a placeholder for the assistant's response.
	m.currentHistory.Messages = append(m.currentHistory.Messages, chat.Message{
		Role:    "assistant",
		Content: "",
	})
	m.currentHistory.UpdatedAt = time.Now()
}

// UpdateAssistantMessage updates the content of the last assistant message.
// This is used for streaming responses.
func (m *Manager) UpdateAssistantMessage(chunk string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.currentHistory == nil || len(m.currentHistory.Messages) == 0 {
		return
	}

	lastIndex := len(m.currentHistory.Messages) - 1
	if m.currentHistory.Messages[lastIndex].Role == "assistant" {
		m.currentHistory.Messages[lastIndex].Content += chunk
		m.currentHistory.UpdatedAt = time.Now()
	}
}
