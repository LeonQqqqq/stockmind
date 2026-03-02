package service

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"stockmind-go/internal/client"
	"stockmind-go/internal/model"
	"stockmind-go/internal/prompt"
	"stockmind-go/internal/store"
)

const maxToolLoops = 10

type ChatService struct {
	claude    *client.ClaudeClient
	data      *client.DataClient
	store     *store.SQLiteStore
}

func NewChatService(claude *client.ClaudeClient, data *client.DataClient, store *store.SQLiteStore) *ChatService {
	return &ChatService{
		claude: claude,
		data:   data,
		store:  store,
	}
}

// Chat processes a user message through the Tool Use loop and returns the final text via a channel.
func (s *ChatService) Chat(sessionID, userMessage string, textCh chan<- string) error {
	defer close(textCh)

	// Load history
	historyMsgs, err := s.store.GetMessages(sessionID)
	if err != nil {
		return fmt.Errorf("load history: %w", err)
	}

	// Build Claude messages from history
	var messages []model.ClaudeMessage
	for _, m := range historyMsgs {
		messages = append(messages, model.ClaudeMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	// Append user message
	messages = append(messages, model.ClaudeMessage{
		Role:    "user",
		Content: userMessage,
	})

	// Save user message
	s.store.SaveMessage(sessionID, "user", userMessage)

	tools := prompt.GetTools()

	for loop := 0; loop < maxToolLoops; loop++ {
		req := model.ClaudeRequest{
			System:   prompt.SystemPrompt,
			Messages: messages,
			Tools:    tools,
		}

		resp, err := s.claude.SendMessage(req)
		if err != nil {
			return fmt.Errorf("claude API call: %w", err)
		}

		// Check for tool_use blocks (only our custom tools, not server_tool_use)
		var toolUses []model.ContentBlock
		var textParts []string

		for _, block := range resp.Content {
			switch block.Type {
			case "tool_use":
				toolUses = append(toolUses, block)
			case "text":
				textParts = append(textParts, block.Text)
			// server_tool_use, web_search_tool_result are handled server-side, just pass through
			}
		}

		if len(toolUses) == 0 {
			// No tool calls - final response
			finalText := ""
			for _, t := range textParts {
				finalText += t
			}
			textCh <- finalText
			s.store.SaveMessage(sessionID, "assistant", finalText)
			return nil
		}

		// Has tool calls - append assistant message, call tools, build results
		messages = append(messages, model.ClaudeMessage{
			Role:    "assistant",
			Content: resp.Content,
		})

		// Call tools concurrently
		type toolResult struct {
			idx    int
			result model.ToolResultContent
		}
		results := make([]toolResult, len(toolUses))
		var wg sync.WaitGroup

		for i, tu := range toolUses {
			wg.Add(1)
			go func(idx int, toolUse model.ContentBlock) {
				defer wg.Done()

				inputMap := make(map[string]interface{})
				inputBytes, _ := json.Marshal(toolUse.Input)
				json.Unmarshal(inputBytes, &inputMap)

				log.Printf("[Tool] %s(%v)", toolUse.Name, inputMap)

				data, err := s.data.CallTool(toolUse.Name, inputMap)
				if err != nil {
					data = fmt.Sprintf(`{"error": "%s"}`, err.Error())
					log.Printf("[Tool] %s error: %v", toolUse.Name, err)
				}

				results[idx] = toolResult{
					idx: idx,
					result: model.ToolResultContent{
						Type:      "tool_result",
						ToolUseID: toolUse.ID,
						Content:   data,
					},
				}
			}(i, tu)
		}
		wg.Wait()

		// Build tool results in order
		var toolResults []interface{}
		for _, r := range results {
			toolResults = append(toolResults, r.result)
		}

		messages = append(messages, model.ClaudeMessage{
			Role:    "user",
			Content: toolResults,
		})

		// Send partial text if any (tool calls may include text)
		for _, t := range textParts {
			if t != "" {
				textCh <- t
			}
		}
	}

	return fmt.Errorf("exceeded max tool loops (%d)", maxToolLoops)
}
