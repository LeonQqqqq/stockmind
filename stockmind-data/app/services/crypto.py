import ccxt
from app.models.schemas import StockQuote, KlineBar, SearchResult
from app.utils.proxy import ccxt_proxies
from app.utils.cache import TTLCache

_exchange = None


def _get_exchange() -> ccxt.okx:
    global _exchange
    if _exchange is None:
        _exchange = ccxt.okx({
            "proxies": ccxt_proxies(),
            "timeout": 15000,
        })
    return _exchange


def _normalize_symbol(symbol: str) -> str:
    """Normalize symbol to ccxt format, e.g. BTC -> BTC/USDT."""
    s = symbol.upper().replace("-", "/")
    if "/" not in s:
        s = f"{s}/USDT"
    return s


@TTLCache(ttl_seconds=15)
async def get_realtime(symbol: str) -> StockQuote:
    """Get crypto realtime quote via OKX."""
    exchange = _get_exchange()
    sym = _normalize_symbol(symbol)
    ticker = exchange.fetch_ticker(sym)
    return StockQuote(
        symbol=sym,
        name=sym,
        price=float(ticker["last"]),
        change=float(ticker.get("change") or 0),
        change_pct=float(ticker.get("percentage") or 0),
        open=float(ticker.get("open") or 0),
        high=float(ticker.get("high") or 0),
        low=float(ticker.get("low") or 0),
        volume=float(ticker.get("baseVolume") or 0),
        timestamp=str(ticker.get("datetime", "")),
    )


@TTLCache(ttl_seconds=300)
async def get_kline(symbol: str, period: str = "daily", count: int = 60) -> list[KlineBar]:
    """Get crypto kline via OKX."""
    exchange = _get_exchange()
    sym = _normalize_symbol(symbol)
    tf_map = {"daily": "1d", "weekly": "1w", "monthly": "1M"}
    timeframe = tf_map.get(period, "1d")
    ohlcv = exchange.fetch_ohlcv(sym, timeframe=timeframe, limit=count)
    bars = []
    for candle in ohlcv:
        bars.append(KlineBar(
            date=exchange.iso8601(candle[0]),
            open=float(candle[1]),
            high=float(candle[2]),
            low=float(candle[3]),
            close=float(candle[4]),
            volume=float(candle[5]),
        ))
    return bars


@TTLCache(ttl_seconds=60)
async def search_stock(keyword: str) -> list[SearchResult]:
    """Search crypto markets on OKX."""
    exchange = _get_exchange()
    exchange.load_markets()
    keyword_upper = keyword.upper()
    results = []
    for sym in exchange.symbols:
        if keyword_upper in sym:
            results.append(SearchResult(symbol=sym, name=sym, market="crypto"))
        if len(results) >= 10:
            break
    return results
