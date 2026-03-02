# StockMind — 个人投资研究助手 (Claude Code 完整开发提示词)

## 项目概述

构建一个个人投资研究助手工具 StockMind。核心理念：**AI 是研究助手和复盘伙伴，不是决策者**。

支持三个市场：A股、美股、加密货币。用户可以在对话中提及股票代码/币种，系统自动识别并注入实时行情数据到 Claude 的上下文中，实现"带数据的投资分析对话"。同时支持从对话中提取投资经验，形成可积累的投资笔记。

## 技术架构

三层微服务：

```
React 前端 (port 3000)
    ↓
Go 后端 Gin (port 8080) — API网关 + 业务逻辑 + AI编排
    ↓                ↓
Python FastAPI       Claude API (Sonnet 4.5)
(port 8001)              ↓
    ↓               SQLite (对话 + 经验)
AKShare / yfinance / ccxt
```

## 重要：网络代理问题

开发环境使用 VPN，导致：
- **海外接口（Yahoo Finance, OKX/Binance）需要走系统代理** 才能访问
- **国内接口（东方财富 via AKShare）需要直连**，海外 IP 会被拒绝

解决方案：在 Python 数据服务中实现 `no_proxy()` 上下文管理器：
- A股的 AKShare 调用包裹在 `with no_proxy():` 中，临时清除代理环境变量
- 美股和加密货币保持走系统代理

```python
# app/utils/proxy.py
import os
from contextlib import contextmanager

_PROXY_KEYS = ["HTTP_PROXY", "HTTPS_PROXY", "http_proxy", "https_proxy", "ALL_PROXY", "all_proxy"]

@contextmanager
def no_proxy():
    """临时清除代理环境变量，退出后恢复。用于国内接口直连。"""
    saved = {}
    for key in _PROXY_KEYS:
        if key in os.environ:
            saved[key] = os.environ.pop(key)
    old_no = os.environ.get("NO_PROXY")
    old_no_l = os.environ.get("no_proxy")
    os.environ["NO_PROXY"] = "*"
    os.environ["no_proxy"] = "*"
    try:
        yield
    finally:
        for key, val in saved.items():
            os.environ[key] = val
        if old_no is not None: os.environ["NO_PROXY"] = old_no
        else: os.environ.pop("NO_PROXY", None)
        if old_no_l is not None: os.environ["no_proxy"] = old_no_l
        else: os.environ.pop("no_proxy", None)
```

A股 service 中的异步执行需要在 no_proxy 内调用：
```python
def _run_sync(func, *args, **kwargs):
    def wrapper():
        with no_proxy():
            return func(*args, **kwargs)
    loop = asyncio.get_event_loop()
    return loop.run_in_executor(None, wrapper)
```

美股和加密货币的 `_run_sync` 不需要包 `no_proxy`，保持正常代理。

---

## Part 1: Python 数据服务 (stockmind-data/)

### 目录结构
```
stockmind-data/
├── main.py                  # FastAPI 入口
├── requirements.txt
└── app/
    ├── __init__.py
    ├── models/
    │   ├── __init__.py      # re-export StockQuote, KlineBar, etc.
    │   └── schemas.py       # Pydantic models
    ├── routers/
    │   ├── __init__.py      # re-export cn_router, us_router, crypto_router
    │   ├── cn_router.py     # A股路由 prefix=/cn
    │   ├── us_router.py     # 美股路由 prefix=/us
    │   └── crypto_router.py # 加密货币路由 prefix=/crypto
    ├── services/
    │   ├── __init__.py
    │   ├── cn_stock.py      # AKShare，所有调用包裹 no_proxy()
    │   ├── us_stock.py      # yfinance
    │   └── crypto.py        # ccxt (默认 binance，可配置 okx)
    └── utils/
        ├── __init__.py
        ├── cache.py         # TTL 缓存 (cachetools)
        └── proxy.py         # no_proxy() 上下文管理器
```

### 依赖
```
fastapi==0.115.0
uvicorn==0.30.6
akshare>=1.14.0
yfinance>=0.2.40
ccxt>=4.3.0
pandas>=2.0.0
pydantic>=2.0.0
cachetools>=5.3.0
```

### 统一数据模型 (schemas.py)
```python
class StockQuote(BaseModel):
    symbol: str
    name: str
    market: str          # cn / us / crypto
    price: float
    change: float        # 涨跌额
    change_pct: float    # 涨跌幅 %
    open: Optional[float] = None
    high: Optional[float] = None
    low: Optional[float] = None
    prev_close: Optional[float] = None
    volume: Optional[float] = None
    turnover: Optional[float] = None
    market_cap: Optional[float] = None
    pe_ratio: Optional[float] = None
    timestamp: Optional[str] = None
    extra: Optional[dict] = None    # 市场特有字段

class KlineBar(BaseModel):
    date: str
    open: float; high: float; low: float; close: float
    volume: float
    turnover: Optional[float] = None
    change_pct: Optional[float] = None

class KlineResponse(BaseModel):
    symbol: str; market: str; period: str; bars: list[KlineBar]

class ApiResponse(BaseModel):
    code: int = 0; message: str = "ok"; data: Optional[dict | list] = None
```

### API 端点
所有路由注册在 `app.include_router(xxx_router, prefix="/api/v1/stock")`

| 端点 | 说明 |
|------|------|
| GET /api/v1/stock/cn/realtime?symbol=600519 | A股实时行情 |
| GET /api/v1/stock/cn/history?symbol=600519&period=daily | A股K线 |
| GET /api/v1/stock/cn/money_flow?symbol=600519 | A股资金流向 |
| GET /api/v1/stock/cn/search?keyword=茅台 | A股搜索 |
| GET /api/v1/stock/us/realtime?symbol=AAPL | 美股行情 |
| GET /api/v1/stock/us/history?symbol=AAPL&period=1y&interval=1d | 美股K线 |
| GET /api/v1/stock/us/search?keyword=apple | 美股搜索 |
| GET /api/v1/stock/crypto/realtime?symbol=BTC | 加密货币行情 |
| GET /api/v1/stock/crypto/history?symbol=BTC&timeframe=1d&limit=365 | 加密K线 |
| GET /api/v1/stock/crypto/market | 主流加密概览 |
| GET /api/v1/stock/crypto/search?keyword=BTC | 加密搜索 |
| GET /health | 健康检查 |

### 缓存策略
- `realtime`: TTL 30秒, maxsize 500
- `kline_daily`: TTL 3600秒(1小时), maxsize 500
- `general`: TTL 3600秒, maxsize 200

### 数据源详情

**A股 (cn_stock.py) — AKShare**
- 实时行情: `ak.stock_zh_a_spot_em()` 获取全量数据后筛选（东方财富源）
- 历史K线: `ak.stock_zh_a_hist(symbol, period, start_date, end_date, adjust)`
- 资金流向: `ak.stock_individual_fund_flow(stock, market)`
- 搜索: 在全量数据中按代码/名称模糊匹配
- **注意**: 所有 AKShare 调用必须在 `no_proxy()` 上下文中执行

**美股 (us_stock.py) — yfinance**
- 实时行情: `yf.Ticker(symbol).info`
- 历史K线: `yf.Ticker(symbol).history(period, interval)`
- 搜索: `yf.Search(keyword).quotes`
- yfinance 有时会 rate limit，需要做好错误处理

**加密货币 (crypto.py) — ccxt**
- 默认交易所: binance (中国网络需配合 VPN)
- 标准化: 输入 BTC → 自动转为 BTC/USDT
- 实时行情: `exchange.fetch_ticker(pair)`
- 历史K线: `exchange.fetch_ohlcv(pair, timeframe, since, limit)`
- 市场概览: `exchange.fetch_tickers([top_symbols])`
- 搜索: `exchange.load_markets()` 后按 base/symbol 匹配

---

## Part 2: Go 后端 (stockmind-go/)

### 目录结构
```
stockmind-go/
├── cmd/server/main.go       # 入口
├── configs/config.yaml      # 配置文件
├── go.mod
└── internal/
    ├── config/config.go     # 配置加载 (yaml + 环境变量替换)
    ├── model/model.go       # 数据模型 + ApiResponse
    ├── client/
    │   ├── data_client.go   # Python数据服务 HTTP 客户端
    │   └── claude_client.go # Claude API 客户端 (普通 + SSE streaming)
    ├── handler/handler.go   # Gin HTTP handlers
    ├── service/
    │   └── chat_service.go  # 核心：股票识别 → 数据获取 → Prompt构建 → Claude调用
    ├── store/sqlite.go      # SQLite 存储 (sessions, messages, memories)
    └── pkg/prompt/templates.go # System prompt 模板
```

### 依赖
```
github.com/gin-gonic/gin
github.com/gin-contrib/cors
github.com/mattn/go-sqlite3
gopkg.in/yaml.v3
```

### 配置 (config.yaml)
```yaml
server:
  port: 8080             # int 类型

data_service:
  url: "http://localhost:8001"
  timeout: 30            # int 秒

claude:
  api_key: "${CLAUDE_API_KEY}"   # 通过环境变量注入
  base_url: "https://api.anthropic.com"  # 中转站改成你的地址，如 https://xxx.com
  model: "claude-sonnet-4-5-20250929"
  max_tokens: 4096

database:
  path: "./data/stockmind.db"
```

config.go 实现环境变量 `${}` 语法替换，并设置合理默认值。ClaudeConfig 需要包含 `BaseURL string` 字段，默认值 `https://api.anthropic.com`。

### 关键类型对应关系（务必保持一致）

| 包 | 类型/方法名 |
|----|------------|
| model | `OK(data)`, `Fail(code, msg)`, `ApiResponse`, `ChatRequest`, `ChatResponse`, `Memory` |
| store | `Store` (不是 SQLiteStore), `New(dbPath)`, `CreateSession() → (int64, error)`, `SaveMessage()`, `GetMessages()`, `CreateMemory()`, `SearchMemories()`, `ListMemories()`, `DeleteMemory()` |
| client | `DataClient`, `NewDataClient(url, timeout)`, `GetRealtimeQuote(market, symbol) → (json.RawMessage, error)`, `GetHistory(market, symbol, params map[string]string)`, `Search(market, keyword)`, `Health()` |
| client | `ClaudeClient`, `ClaudeMessage{Role, Content}`, `NewClaudeClient(apiKey, baseURL, model, maxTokens)`, `CreateMessage(system, messages) → (string, error)`, `CreateMessageStream(system, messages) → (<-chan StreamEvent, error)`, `StreamEvent{Type, Text, Done}` |
| service | `ChatService`, `NewChatService(dc, cc, store)`, `Chat(ctx, sessionID, message)`, `ChatStream(ctx, sessionID, message, callback)`, `SummarizeExperience(ctx, sessionID)` |
| handler | `Handler`, `New(dc, cs, store)`, `RegisterRoutes(r *gin.Engine)` |
| prompt | `BuildSystemPrompt(stockDataStr, memoriesStr string) → string`, `FormatStockData(json, symbol, market)`, `FormatMemories([]string)`, `ExtractMemoryPrompt` |

### DataClient 设计
- Python 数据服务返回 `{code: 0, message: "ok", data: ...}` 格式
- DataClient 解析后返回 `json.RawMessage`（data 字段原样透传给前端）
- Handler 层直接代理 Python 服务的响应

### ChatService 核心流程
1. 用正则从用户消息中提取股票代码/币种
   - A股: `\b([036]\d{5})\b` → market=cn
   - Crypto: `(?i)\b(BTC|ETH|SOL|...)\b` → market=crypto
   - US: `\b([A-Z]{1,5})\b` + 已知 ticker 白名单 → market=us
2. 用 goroutine **并发**请求每只股票的实时行情（通过 DataClient）
3. 搜索相关投资经验（SQLite keyword search）
4. 用 prompt 模板拼装 system prompt（注入数据 + 经验）
5. 构建对话 history（从 SQLite 取）
6. 调 Claude API（普通/流式）
7. 保存 assistant 回复到 SQLite

### Claude API Client
- **支持中转站**: base_url 可配置，不硬编码 `api.anthropic.com`
- 请求地址: `{base_url}/v1/messages`
- 普通请求: POST, stream=false
- 流式请求: stream=true, 解析 SSE (content_block_delta → text, message_stop → done)
- Header: Content-Type, x-api-key, anthropic-version: 2023-06-01

### SQLite Schema
```sql
CREATE TABLE chat_sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE chat_messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id INTEGER NOT NULL REFERENCES chat_sessions(id),
    role TEXT NOT NULL,        -- user / assistant
    content TEXT NOT NULL,
    metadata TEXT DEFAULT '',  -- JSON: 关联的股票数据快照
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE memories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    content TEXT NOT NULL,
    category TEXT DEFAULT '',     -- 技术分析/基本面/情绪/教训/策略
    symbols TEXT DEFAULT '',     -- 逗号分隔的相关代码
    source_chat INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### Go API 端点
| 端点 | 说明 |
|------|------|
| GET /api/v1/stock/:market/realtime?symbol=xxx | 代理 Python 数据服务 |
| GET /api/v1/stock/:market/history | 代理 |
| GET /api/v1/stock/:market/search?keyword=xxx | 代理 |
| POST /api/v1/chat | AI对话 (自动识别股票+注入数据) |
| POST /api/v1/chat/stream | AI流式对话 (SSE) |
| GET /api/v1/chat/sessions | 会话列表 |
| GET /api/v1/chat/sessions/:id | 会话详情 |
| GET /api/v1/chat/sessions/:id/messages | 消息列表 |
| POST /api/v1/chat/sessions/:id/summarize | AI提取投资经验 |
| GET /api/v1/memories | 经验列表 |
| POST /api/v1/memories | 手动添加经验 |
| GET /api/v1/memories/search?q=xxx | 搜索经验 |
| DELETE /api/v1/memories/:id | 删除经验 |

### System Prompt 模板
```
你是 StockMind，一个专业的投资研究助手。

## 你的职责
1. 基于真实市场数据进行客观分析（技术面、基本面、资金面）
2. 帮助用户评估他人的股评观点，做逻辑验证和数据交叉对比
3. 提供多角度思考，指出不同观点的合理性和风险
4. 在分析中明确区分"事实"和"观点"

## 分析原则
- 数据驱动，引用具体数字支撑观点
- 任何分析都要指出潜在风险和不确定性
- 如果数据不足以支撑结论，明确说明
- 不给出明确的买卖建议，而是帮助用户理清逻辑
- 回答要简洁有力，不要废话

## 当前可用的市场数据
{注入的实时行情 JSON}

## 用户积累的投资经验
{注入的历史经验}
```

---

## Part 3: React 前端 (stockmind-web/) — 后续开发

React + TypeScript + Tailwind + shadcn/ui 前端。三栏布局：
- 左栏：市场面板（K线图、自选股、行情概览）
- 中栏：AI对话窗口（SSE 流式接收）
- 右栏：投资经验库

K线图库: lightweight-charts (TradingView 开源)
状态管理: Zustand
流式: SSE EventSource 接收

---

## 开发顺序

1. **先确保 Python 数据服务完全可用**
   - 三个市场的实时行情、K线、搜索都能正常返回
   - A股要测试 no_proxy 是否生效
   - 加密货币在 VPN 环境下是否可达
   - 美股 yfinance 的 rate limit 处理

2. **Go 后端能编译运行**
   - 类型名必须完全对应（见上面的类型对应表）
   - `go mod tidy` + `go run cmd/server/main.go` 无报错
   - curl 测试代理接口和 AI 对话接口

3. **React 前端搭建**

---

## 运行环境

- OS: Windows (CMD/PowerShell)
- Python 3.11+, Go 1.21+, Node 18+
- VPN 常开，网络需做代理分流处理
- Claude API Key 通过环境变量 `CLAUDE_API_KEY` 注入（支持中转站 API Key）
- 中转站用户需在 config.yaml 中修改 `claude.base_url` 为中转站地址
- Windows 设置环境变量: `set CLAUDE_API_KEY=sk-ant-xxx` (CMD) 或 `$env:CLAUDE_API_KEY="sk-ant-xxx"` (PowerShell)

---

## 面试价值

此项目可展示的技术点：
- 微服务架构（Go + Python HTTP 通信）
- RESTful API 设计 + 统一响应格式
- Go 并发（goroutine 并行获取多只股票数据）
- 多级缓存（内存 TTL 缓存）
- SQLite schema 设计 + 索引
- 第三方 API 集成（Claude, AKShare, yfinance, ccxt）
- SSE 流式传输
- 网络代理分流处理
- Clean Architecture 分层
