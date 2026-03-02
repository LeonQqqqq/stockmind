package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
		url = fmt.Sprintf("%s/api/v1/stock/%s/realtime?symbol=%s", c.baseURL, market, symbol)

	case "get_kline":
		market, _ := input["market"].(string)
		symbol, _ := input["symbol"].(string)
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
		url = fmt.Sprintf("%s/api/v1/stock/cn/money_flow?symbol=%s", c.baseURL, symbol)

	case "search_stock":
		market, _ := input["market"].(string)
		keyword, _ := input["keyword"].(string)
		url = fmt.Sprintf("%s/api/v1/stock/%s/search?keyword=%s", c.baseURL, market, keyword)

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

	var dataResp model.DataResponse
	if err := json.Unmarshal(body, &dataResp); err != nil {
		return string(body), nil
	}

	if dataResp.Code != 0 {
		return "", fmt.Errorf("data service error: %s", dataResp.Message)
	}

	result, _ := json.Marshal(dataResp.Data)
	return string(result), nil
}
