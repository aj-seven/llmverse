package history

import (
	"time"

	"github.com/aj-seven/llmverse/pkg/chat"
)

// History represents a single chat history.
type History struct {
	ID        string         `json:"id"`
	Title     string         `json:"title"`
	Model     string         `json:"model"`
	Messages  []chat.Message `json:"messages"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// Storage defines the interface for history storage operations.
type Storage interface {
	SaveHistory(history History) error
	GetHistory(id string) (History, error)
	GetHistories() ([]History, error)
	DeleteHistory(id string) error
}
