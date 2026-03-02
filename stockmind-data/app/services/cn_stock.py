import re
import httpx
import akshare as ak
from typing import Optional
from app.models.schemas import StockQuote, KlineBar, MoneyFlow, SearchResult
from app.utils.proxy import no_proxy
from app.utils.cache import TTLCache


def _market_prefix(symbol: str) -> str:
    """Add sh/sz prefix based on stock code."""
    code = symbol.lstrip("shsz")
    if code.startswith(("6", "9", "5")):
        return f"sh{code}"
    return f"sz{code}"


def _market_tag(symbol: str) -> str:
    """Return 'sh' or 'sz' string for fund flow API."""
    code = symbol.lstrip("shsz")
    if code.startswith(("6", "9", "5")):
        return "sh"
    return "sz"


@TTLCache(ttl_seconds=10)
async def get_realtime(symbol: str) -> StockQuote:
    """Get realtime quote from Sina."""
    code = symbol.lstrip("shsz")
    prefixed = _market_prefix(code)
    url = f"https://hq.sinajs.cn/list={prefixed}"
    headers = {"Referer": "https://finance.sina.com.cn"}
    async with httpx.AsyncClient(verify=False, trust_env=False) as client:
        resp = await client.get(url, headers=headers)
    text = resp.content.decode("gbk")
    m = re.search(r'"(.+)"', text)
    if not m:
        raise ValueError(f"No data for {symbol}")
    parts = m.group(1).split(",")
    return StockQuote(
        symbol=code,
        name=parts[0],
        price=float(parts[3]),
        change=round(float(parts[3]) - float(parts[2]), 2),
        change_pct=round((float(parts[3]) - float(parts[2])) / float(parts[2]) * 100, 2) if float(parts[2]) else 0,
        open=float(parts[1]),
        high=float(parts[4]),
        low=float(parts[5]),
        volume=float(parts[8]),
        amount=float(parts[9]),
        timestamp=f"{parts[30]} {parts[31]}",
    )


@TTLCache(ttl_seconds=300)
async def get_kline(symbol: str, period: str = "daily", count: int = 60) -> list[KlineBar]:
    """Get kline data using AKShare (sina source)."""
    code = symbol.lstrip("shsz")
    prefixed = _market_prefix(code)
    with no_proxy():
        df = ak.stock_zh_a_daily(symbol=prefixed, adjust="qfq")
    df = df.tail(count)
    bars = []
    for _, row in df.iterrows():
        bars.append(KlineBar(
            date=str(row["date"]),
            open=float(row["open"]),
            high=float(row["high"]),
            low=float(row["low"]),
            close=float(row["close"]),
            volume=float(row["volume"]),
        ))
    return bars


@TTLCache(ttl_seconds=300)
async def get_money_flow(symbol: str) -> list[MoneyFlow]:
    """Get money flow data using AKShare."""
    code = symbol.lstrip("shsz")
    market = _market_tag(code)
    with no_proxy():
        df = ak.stock_individual_fund_flow(stock=code, market=market)
    df = df.tail(30)
    flows = []
    cols = df.columns.tolist()
    for _, row in df.iterrows():
        flows.append(MoneyFlow(
            date=str(row[cols[0]]),
            main_net=float(row[cols[1]]) if len(cols) > 1 else 0,
            main_pct=float(row[cols[2]]) if len(cols) > 2 else 0,
            super_large_net=float(row[cols[3]]) if len(cols) > 3 else 0,
            large_net=float(row[cols[5]]) if len(cols) > 5 else 0,
            medium_net=float(row[cols[7]]) if len(cols) > 7 else 0,
            small_net=float(row[cols[9]]) if len(cols) > 9 else 0,
        ))
    return flows


@TTLCache(ttl_seconds=60)
async def search_stock(keyword: str) -> list[SearchResult]:
    """Search A-share stocks using Sina suggest API."""
    url = f"https://suggest3.sinajs.cn/suggest/type=11,12&key={keyword}"
    headers = {"Referer": "https://finance.sina.com.cn"}
    async with httpx.AsyncClient(verify=False, trust_env=False) as client:
        resp = await client.get(url, headers=headers)
    text = resp.content.decode("gbk")
    m = re.search(r'"(.+)"', text)
    if not m or not m.group(1):
        return []
    results = []
    for item in m.group(1).split(";"):
        parts = item.split(",")
        if len(parts) >= 4:
            results.append(SearchResult(
                symbol=parts[2],
                name=parts[4] if len(parts) > 4 else parts[2],
                market="cn",
            ))
    return results[:10]
