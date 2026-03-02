import yfinance as yf
from app.models.schemas import StockQuote, KlineBar, SearchResult
from app.utils.cache import TTLCache


@TTLCache(ttl_seconds=30)
async def get_realtime(symbol: str) -> StockQuote:
    """Get US stock realtime quote via yfinance."""
    ticker = yf.Ticker(symbol)
    info = ticker.fast_info
    prev_close = info.previous_close or 0
    price = info.last_price or 0
    change = round(price - prev_close, 2)
    change_pct = round(change / prev_close * 100, 2) if prev_close else 0
    return StockQuote(
        symbol=symbol.upper(),
        name=symbol.upper(),
        price=price,
        change=change,
        change_pct=change_pct,
        open=info.open,
        high=info.day_high,
        low=info.day_low,
        volume=info.last_volume,
    )


@TTLCache(ttl_seconds=300)
async def get_kline(symbol: str, period: str = "daily", count: int = 60) -> list[KlineBar]:
    """Get US stock kline via yfinance."""
    interval_map = {"daily": "1d", "weekly": "1wk", "monthly": "1mo"}
    interval = interval_map.get(period, "1d")
    ticker = yf.Ticker(symbol)
    df = ticker.history(period="6mo", interval=interval)
    df = df.tail(count)
    bars = []
    for idx, row in df.iterrows():
        bars.append(KlineBar(
            date=str(idx.date()),
            open=round(float(row["Open"]), 2),
            high=round(float(row["High"]), 2),
            low=round(float(row["Low"]), 2),
            close=round(float(row["Close"]), 2),
            volume=float(row["Volume"]),
        ))
    return bars


@TTLCache(ttl_seconds=60)
async def search_stock(keyword: str) -> list[SearchResult]:
    """Search US stocks - simple pass-through since yfinance uses exact symbols."""
    return [SearchResult(symbol=keyword.upper(), name=keyword.upper(), market="us")]
