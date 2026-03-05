package client

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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
			Timeout: 300 * time.Second,
		},
	}
}

type streamRequest struct {
	model.ClaudeRequest
	Stream bool `json:"stream"`
}

func (c *ClaudeClient) SendMessage(req model.ClaudeRequest) (*model.ClaudeResponse, error) {
	req.Model = c.modelName
	if req.MaxTokens == 0 {
		req.MaxTokens = c.maxTokens
	}

	sReq := streamRequest{ClaudeRequest: req, Stream: true}
	body, err := json.Marshal(sReq)
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

	if resp.StatusCode != http.StatusOK {
		var buf bytes.Buffer
		buf.ReadFrom(resp.Body)
		return nil, fmt.Errorf("claude API error (status %d): %s", resp.StatusCode, buf.String())
	}

	// Parse SSE stream and reconstruct a ClaudeResponse
	result := &model.ClaudeResponse{}
	var contentBlocks []model.ContentBlock
	// Track JSON input accumulators per content block index
	inputAccum := make(map[int]string)

	scanner := bufio.NewScanner(resp.Body)
	// Increase buffer for large events
	scanner.Buffer(make([]byte, 0, 256*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := line[6:]

		var event struct {
			Type         string              `json:"type"`
			Index        int                 `json:"index"`
			Message      *model.ClaudeResponse `json:"message,omitempty"`
			ContentBlock *model.ContentBlock  `json:"content_block,omitempty"`
			Delta        *struct {
				Type        string      `json:"type"`
				Text        string      `json:"text,omitempty"`
				PartialJSON string      `json:"partial_json,omitempty"`
			} `json:"delta,omitempty"`
		}
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}

		switch event.Type {
		case "message_start":
			if event.Message != nil {
				result.ID = event.Message.ID
				result.Type = event.Message.Type
				result.Role = event.Message.Role
				result.Usage = event.Message.Usage
			}

		case "content_block_start":
			if event.ContentBlock != nil {
				// Grow slice to fit index
				for len(contentBlocks) <= event.Index {
					contentBlocks = append(contentBlocks, model.ContentBlock{})
				}
				contentBlocks[event.Index] = *event.ContentBlock
			}

		case "content_block_delta":
			if event.Delta != nil && event.Index < len(contentBlocks) {
				switch event.Delta.Type {
				case "text_delta":
					contentBlocks[event.Index].Text += event.Delta.Text
				case "input_json_delta":
					inputAccum[event.Index] += event.Delta.PartialJSON
				}
			}

		case "content_block_stop":
			// Parse accumulated JSON input for tool_use blocks
			if event.Index < len(contentBlocks) && contentBlocks[event.Index].Type == "tool_use" {
				if raw, ok := inputAccum[event.Index]; ok && raw != "" {
					var parsed interface{}
					if err := json.Unmarshal([]byte(raw), &parsed); err == nil {
						contentBlocks[event.Index].Input = parsed
					}
				}
			}

		case "message_delta":
			// Contains stop_reason and final usage
			var msgDelta struct {
				Type  string `json:"type"`
				Delta struct {
					StopReason string `json:"stop_reason"`
				} `json:"delta"`
				Usage model.Usage `json:"usage"`
			}
			if err := json.Unmarshal([]byte(data), &msgDelta); err == nil {
				result.StopReason = msgDelta.Delta.StopReason
			}
		}
	}

	result.Content = contentBlocks
	return result, nil
}
