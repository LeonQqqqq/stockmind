package prompt

import "stockmind-go/internal/model"

const SystemPrompt = `你是 StockMind，用户的私人投资研究搭子。说话自然随意，像一个懂行的朋友在聊天，不要用"您"，用"你"。

风格要求：
- 说人话，别打官腔。少用"首先/其次/综上所述"这类八股句式
- 直接给观点，别铺垫太多。用户问"怎么看"，先亮态度再说理由
- 可以用口语化表达，比如"这波拉升挺猛的"、"资金面有点虚"
- 数据要精准，但解读要接地气
- 该提风险就提，但别每句话都"投资有风险"
- 适当用 Markdown 让数据部分更清晰，但别过度格式化，正文不要动不动就加粗

你可以：
- 拉实时行情、K线、资金流向（A股/美股/加密）
- 联网搜新闻、政策、宏观数据
- 搜索股票/币种

市场代码：A股用数字如 "600519"(market="cn")，美股如 "AAPL"(market="us")，加密如 "BTC"(market="crypto")。`

func GetTools() []model.ClaudeTool {
	return []model.ClaudeTool{
		{
			Type:    "web_search_20250305",
			Name:    "web_search",
			MaxUses: 5,
		},
		{
			Name:        "get_realtime_quote",
			Description: "获取股票/加密货币实时行情。返回最新价格、涨跌幅、成交量等数据。",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"market": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"cn", "us", "crypto"},
						"description": "市场类型: cn=A股, us=美股, crypto=加密货币",
					},
					"symbol": map[string]interface{}{
						"type":        "string",
						"description": "股票/币种代码，如 600519、AAPL、BTC",
					},
				},
				"required": []string{"market", "symbol"},
			},
		},
		{
			Name:        "get_kline",
			Description: "获取K线数据(日线/周线/月线)。用于技术分析和趋势判断。",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"market": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"cn", "us", "crypto"},
						"description": "市场类型",
					},
					"symbol": map[string]interface{}{
						"type":        "string",
						"description": "股票/币种代码",
					},
					"period": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"daily", "weekly", "monthly"},
						"description": "K线周期",
					},
					"count": map[string]interface{}{
						"type":        "integer",
						"description": "返回条数，默认60",
					},
				},
				"required": []string{"market", "symbol"},
			},
		},
		{
			Name:        "get_money_flow",
			Description: "获取A股个股资金流向数据。包含主力、超大单、大单、中单、小单的净流入金额和占比。仅支持A股。",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"symbol": map[string]interface{}{
						"type":        "string",
						"description": "A股股票代码，如 600519",
					},
				},
				"required": []string{"symbol"},
			},
		},
		{
			Name:        "search_stock",
			Description: "搜索股票/加密货币。根据关键词搜索匹配的股票或币种。",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"market": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"cn", "us", "crypto"},
						"description": "市场类型",
					},
					"keyword": map[string]interface{}{
						"type":        "string",
						"description": "搜索关键词，如公司名称或代码",
					},
				},
				"required": []string{"market", "keyword"},
			},
		},
	}
}
