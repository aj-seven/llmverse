package aihub

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/aj-seven/llmverse/pkg/chat"
	"github.com/aj-seven/llmverse/internal/config"
)

type OllamaChatRequest struct {
	Model    string         `json:"model"`
	Messages []chat.Message `json:"messages"`
	Stream   bool           `json:"stream"`
}

type OllamaChatResponse struct {
	Message chat.Message `json:"message"`
	Done    bool         `json:"done"`
}

func StreamChat(
	modelName string,
	messages []chat.Message,
	cfg *config.Config,
) (chan string, error) {

	if cfg != nil && cfg.Assistant.Message != "" {
		messages = append([]chat.Message{
			{
				Role:    "system",
				Content: cfg.Assistant.Message,
			},
		}, messages...)
	}

	reqBody := OllamaChatRequest{
		Model:    modelName,
		Messages: messages,
		Stream:   true,
	}

	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(
		cfg.Host+"/api/chat",
		"application/json",
		bytes.NewBuffer(reqBytes),
	)
	if err != nil {
		return nil, err
	}

	stream := make(chan string)

	go func() {
		defer close(stream)
		defer resp.Body.Close()

		decoder := json.NewDecoder(resp.Body)
		for {
			var chatResp OllamaChatResponse
			if err := decoder.Decode(&chatResp); err != nil {
				if err == io.EOF {
					break
				}
				stream <- "Error: " + err.Error()
				return
			}

			stream <- chatResp.Message.Content

			if chatResp.Done {
				break
			}
		}
	}()

	return stream, nil
}
