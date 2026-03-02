from pydantic import BaseModel
from typing import Optional


class StockQuote(BaseModel):
    symbol: str
    name: str
    price: float
    change: float
    change_pct: float
    open: Optional[float] = None
    high: Optional[float] = None
    low: Optional[float] = None
    volume: Optional[float] = None
    amount: Optional[float] = None
    timestamp: Optional[str] = None


class KlineBar(BaseModel):
    date: str
    open: float
    high: float
    low: float
    close: float
    volume: float
    amount: Optional[float] = None


class MoneyFlow(BaseModel):
    date: str
    main_net: float
    main_pct: float
    super_large_net: Optional[float] = None
    large_net: Optional[float] = None
    medium_net: Optional[float] = None
    small_net: Optional[float] = None


class SearchResult(BaseModel):
    symbol: str
    name: str
    market: str


class ApiResponse(BaseModel):
    code: int = 0
    message: str = "ok"
    data: object = None
