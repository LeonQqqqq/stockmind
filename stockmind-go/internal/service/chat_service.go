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

	customTools := prompt.GetTools()
	webTools := prompt.GetWebTools()

	for loop := 0; loop < maxToolLoops; loop++ {
		// Send with web_search only first; relay API crashes when mixing server tools + custom tools
		req := model.ClaudeRequest{
			System:   prompt.SystemPrompt,
			Messages: messages,
			Tools:    append(webTools, customTools...),
		}

		resp, err := s.claude.SendMessage(req)
		if err != nil {
			// Relay API may fail when web_search fires alongside custom tools.
			// Retry without web_search.
			req.Tools = customTools
			resp, err = s.claude.SendMessage(req)
			if err != nil {
				return fmt.Errorf("claude API call: %w", err)
			}
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

		// Has tool calls - append assistant message (filter out server_tool_use/web_search blocks
		// to avoid relay API "Invalid encrypted_content" error on next round)
		var cleanContent []model.ContentBlock
		for _, block := range resp.Content {
			if block.Type == "server_tool_use" || block.Type == "web_search_tool_result" {
				continue
			}
			cleanContent = append(cleanContent, block)
		}
		messages = append(messages, model.ClaudeMessage{
			Role:    "assistant",
			Content: cleanContent,
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

				data, err := s.callTool(toolUse.Name, inputMap)
				if err != nil {
					errMsg, _ := json.Marshal(map[string]string{"error": err.Error()})
					data = string(errMsg)
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

// callTool routes tool calls to the appropriate handler.
func (s *ChatService) callTool(name string, input map[string]interface{}) (string, error) {
	switch name {
	case "save_experience":
		expType, _ := input["type"].(string)
		title, _ := input["title"].(string)
		content, _ := input["content"].(string)
		tags, _ := input["tags"].(string)
		id, err := s.store.CreateExperience(expType, title, content, tags)
		if err != nil {
			return "", err
		}
		result, _ := json.Marshal(map[string]interface{}{
			"success": true,
			"id":      id,
			"message": "经验已保存",
		})
		return string(result), nil

	case "search_experience":
		keyword, _ := input["keyword"].(string)
		exps, err := s.store.SearchExperiences(keyword)
		if err != nil {
			return "", err
		}
		result, _ := json.Marshal(map[string]interface{}{
			"count":       len(exps),
			"experiences": exps,
		})
		return string(result), nil

	case "save_opinion":
		author, _ := input["author"].(string)
		content, _ := input["content"].(string)
		tags, _ := input["tags"].(string)
		id, err := s.store.CreateOpinion(author, content, tags)
		if err != nil {
			return "", err
		}
		result, _ := json.Marshal(map[string]interface{}{
			"success": true,
			"id":      id,
			"message": "观点已记录",
		})
		return string(result), nil

	case "search_opinions":
		keyword, _ := input["keyword"].(string)
		ops, err := s.store.SearchOpinions(keyword)
		if err != nil {
			return "", err
		}
		result, _ := json.Marshal(map[string]interface{}{
			"count":    len(ops),
			"opinions": ops,
		})
		return string(result), nil

	default:
		return s.data.CallTool(name, input)
	}
}
