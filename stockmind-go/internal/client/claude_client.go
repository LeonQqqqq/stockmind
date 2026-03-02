package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"stockmind-go/internal/config"
	"stockmind-go/internal/model"
)

type ClaudeClient struct {
	apiKey     string
	baseURL    string
	modelName  string
	maxTokens  int
	httpClient *http.Client
}

func NewClaudeClient(cfg config.ClaudeConfig) *ClaudeClient {
	return &ClaudeClient{
		apiKey:    cfg.APIKey,
		baseURL:   cfg.BaseURL,
		modelName: cfg.Model,
		maxTokens: cfg.MaxTokens,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func (c *ClaudeClient) SendMessage(req model.ClaudeRequest) (*model.ClaudeResponse, error) {
	req.Model = c.modelName
	if req.MaxTokens == 0 {
		req.MaxTokens = c.maxTokens
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/v1/messages", c.baseURL)
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("claude API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var claudeResp model.ClaudeResponse
	if err := json.Unmarshal(respBody, &claudeResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &claudeResp, nil
}
