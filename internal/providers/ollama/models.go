package aihub

import (
	"encoding/json"
	"net/http"
	"github.com/aj-seven/llmverse/internal/config"
	"time"
)

type OllamaModels struct {
	Models []OllamaModel `json:"models"`
}

type OllamaModel struct {
	Name       string    `json:"name"`
	Model      string    `json:"model"`
	ModifiedAt time.Time `json:"modified_at"`
	Size       int64     `json:"size"`
	Digest     string    `json:"digest"`
	Details    struct {
		Family            string `json:"family"`
		ParameterSize     string `json:"parameter_size"`
		QuantizationLevel string `json:"quantization_level"`
	} `json:"details"`
}

func GetModelsDetailed(cfg *config.Config) ([]OllamaModel, error) {
	resp, err := http.Get(cfg.Host + "/api/tags")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var models OllamaModels
	if err := json.NewDecoder(resp.Body).Decode(&models); err != nil {
		return nil, err
	}

	return models.Models, nil
}
