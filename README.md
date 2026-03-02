# StockMind

AI 驱动的个人投资研究助手。基于 Claude Tool Use 实现"带数据的投资分析对话"，支持 A股、美股、加密货币三大市场实时数据 + 联网搜索新闻和宏观信息。

## 架构

```
React 前端 (3000) → Go 后端 (8080) → Python 数据服务 (8001)
                         ↓
                    Claude API (Tool Use + Web Search)
```

- **stockmind-data** — Python FastAPI，负责拉取市场数据（新浪行情、AKShare、yfinance、ccxt OKX）
- **stockmind-go** — Go Gin，负责 Claude Tool Use 编排循环、SSE 流式输出、SQLite 持久化
- **stockmind-web** — React + TypeScript + Tailwind，两栏布局（AI 对话 + 投资经验库）

## 功能

- 实时行情查询（A股/美股/加密货币）
- K线数据与技术分析
- A股资金流向分析
- 联网搜索财经新闻、政策、宏观环境
- 投资经验库（CRUD + 搜索）
- 多轮对话，历史记录持久化
- SSE 流式输出

## 快速开始

### 前置要求

- Python 3.12+
- Go 1.20+
- Node.js 18+
- Claude API Key

### 1. 安装依赖

```bash
# Python
cd stockmind-data
pip install -r requirements.txt

# Go
cd stockmind-go
go mod tidy

# React
cd stockmind-web
npm install
```

### 2. 配置环境变量

```bash
export CLAUDE_API_KEY="your-api-key"
export CLAUDE_BASE_URL="https://api.anthropic.com"  # 或你的中转地址
```

### 3. 启动服务

三个终端分别启动：

```bash
# 终端 1：Python 数据服务
cd stockmind-data
python -m uvicorn main:app --host 0.0.0.0 --port 8001

# 终端 2：Go 后端
cd stockmind-go
go run cmd/server/main.go

# 终端 3：React 前端
cd stockmind-web
npx vite --host
```

打开浏览器访问 `http://localhost:3000`

## 项目结构

```
stockmind/
├── stockmind-data/          # Python 数据服务
│   ├── main.py              # FastAPI 入口
│   ├── app/
│   │   ├── routers/         # cn / us / crypto 路由
│   │   ├── services/        # 数据源实现
│   │   └── utils/           # 代理、缓存工具
│   └── requirements.txt
├── stockmind-go/            # Go 后端
│   ├── cmd/server/main.go   # 入口
│   ├── configs/config.yaml  # 配置
│   └── internal/
│       ├── client/          # Claude API / 数据服务客户端
│       ├── service/         # Tool Use 编排循环
│       ├── handler/         # HTTP handlers
│       ├── store/           # SQLite
│       └── prompt/          # System prompt + 工具定义
└── stockmind-web/           # React 前端
    └── src/
        ├── components/      # ChatPanel / MemoryPanel / Sidebar
        ├── hooks/           # useChat (SSE 流式)
        └── stores/          # Zustand 状态管理
```

## 数据源

| 数据 | 来源 | 说明 |
|------|------|------|
| A股实时行情 | 新浪 hq.sinajs.cn | 免费、轻量 |
| A股K线 | AKShare (新浪源) | `stock_zh_a_daily` |
| A股资金流向 | AKShare | `stock_individual_fund_flow` |
| 美股 | yfinance | 需要代理 |
| 加密货币 | ccxt OKX | 需要代理 |
| 新闻/宏观 | Claude Web Search | API 内置 |

## 代理配置

如果需要代理访问美股/加密数据，在 `stockmind-data/main.py` 启动前设置：

```bash
export ALL_PROXY="socks5h://your-proxy:port"
```

ccxt 需要在构造参数中显式传入 `proxies`，已在代码中处理。

## License

MIT
