package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"time"

	"stockmind-go/internal/config"
	"stockmind-go/internal/model"
)

type DataClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewDataClient(cfg config.DataServiceConfig) *DataClient {
	return &DataClient{
		baseURL: cfg.BaseURL,
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		},
	}
}

func (c *DataClient) CallTool(toolName string, input map[string]interface{}) (string, error) {
	var url string

	switch toolName {
	case "get_realtime_quote":
		market, _ := input["market"].(string)
		symbol, _ := input["symbol"].(string)
		if market == "" || symbol == "" {
			return "", fmt.Errorf("get_realtime_quote requires market and symbol")
		}
		url = fmt.Sprintf("%s/api/v1/stock/%s/realtime?symbol=%s", c.baseURL, market, symbol)

	case "get_kline":
		market, _ := input["market"].(string)
		symbol, _ := input["symbol"].(string)
		if market == "" || symbol == "" {
			return "", fmt.Errorf("get_kline requires market and symbol")
		}
		period, _ := input["period"].(string)
		if period == "" {
			period = "daily"
		}
		count := 60
		if c, ok := input["count"].(float64); ok {
			count = int(c)
		}
		url = fmt.Sprintf("%s/api/v1/stock/%s/kline?symbol=%s&period=%s&count=%d",
			c.baseURL, market, symbol, period, count)

	case "get_money_flow":
		symbol, _ := input["symbol"].(string)
		if symbol == "" {
			return "", fmt.Errorf("get_money_flow requires symbol")
		}
		url = fmt.Sprintf("%s/api/v1/stock/cn/money_flow?symbol=%s", c.baseURL, symbol)

	case "search_stock":
		market, _ := input["market"].(string)
		keyword, _ := input["keyword"].(string)
		if market == "" || keyword == "" {
			return "", fmt.Errorf("search_stock requires market and keyword")
		}
		url = fmt.Sprintf("%s/api/v1/stock/%s/search?keyword=%s", c.baseURL, market, neturl.QueryEscape(keyword))

	default:
		return "", fmt.Errorf("unknown tool: %s", toolName)
	}

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("data service call failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("data service HTTP %d: %s", resp.StatusCode, string(body))
	}

	var dataResp model.DataResponse
	if err := json.Unmarshal(body, &dataResp); err != nil {
		return "", fmt.Errorf("data service parse error: %w", err)
	}

	if dataResp.Code != 0 {
		return "", fmt.Errorf("data service error: %s", dataResp.Message)
	}

	result, _ := json.Marshal(dataResp.Data)
	return string(result), nil
}
