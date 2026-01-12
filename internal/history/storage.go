package history

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	appDirName     = "llmv"
	historyDirName = "history"
)

// FileStorage implements the Storage interface using JSON files.
type FileStorage struct {
	dataDir string
}

// NewFileStorage creates in:
// ~/.llmv/history
func NewFileStorage() (*FileStorage, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home dir: %w", err)
	}

	dataDir := filepath.Join(homeDir, "."+appDirName, historyDirName)

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create history directory: %w", err)
	}

	return &FileStorage{dataDir: dataDir}, nil
}

// SaveHistory saves a single chat history to disk.
// It will not save histories with no messages.
func (s *FileStorage) SaveHistory(h History) error {
	// Do not save sessions that have no messages.
	if len(h.Messages) == 0 {
		return nil
	}

	if h.ID == "" {
		h.ID = uuid.New().String()
	}
	if h.CreatedAt.IsZero() {
		h.CreatedAt = time.Now()
	}
	h.UpdatedAt = time.Now()

	data, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	filePath := s.getHistoryFilePath(h.ID)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write history file: %w", err)
	}

	return nil
}

// GetHistory loads a single chat history by ID.
func (s *FileStorage) GetHistory(id string) (History, error) {
	filePath := s.getHistoryFilePath(id)

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return History{}, fmt.Errorf("history with ID %s not found", id)
		}
		return History{}, fmt.Errorf("failed to read history file: %w", err)
	}

	var h History
	if err := json.Unmarshal(data, &h); err != nil {
		return History{}, fmt.Errorf("failed to unmarshal history: %w", err)
	}

	return h, nil
}

// GetHistories loads all chat histories sorted by most recently updated.
func (s *FileStorage) GetHistories() ([]History, error) {
	var histories []History

	err := filepath.WalkDir(s.dataDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !strings.HasPrefix(d.Name(), "history_") || !strings.HasSuffix(d.Name(), ".json") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read history file %s: %w", d.Name(), err)
		}

		var h History
		if err := json.Unmarshal(data, &h); err != nil {
			return fmt.Errorf("failed to unmarshal history file %s: %w", d.Name(), err)
		}

		// Skip empty histories defensively
		if len(h.Messages) == 0 {
			return nil
		}

		histories = append(histories, h)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk history directory: %w", err)
	}

	sort.Slice(histories, func(i, j int) bool {
		return histories[i].UpdatedAt.After(histories[j].UpdatedAt)
	})

	return histories, nil
}

// DeleteHistory removes a history file by ID.
func (s *FileStorage) DeleteHistory(id string) error {
	filePath := s.getHistoryFilePath(id)

	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("history with ID %s not found", id)
		}
		return fmt.Errorf("failed to delete history file: %w", err)
	}

	return nil
}

// getHistoryFilePath returns the full path to a history file.
func (s *FileStorage) getHistoryFilePath(id string) string {
	return filepath.Join(s.dataDir, fmt.Sprintf("history_%s.json", id))
}
