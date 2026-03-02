from fastapi import APIRouter, HTTPException
from app.models.schemas import ApiResponse
from app.services import crypto

router = APIRouter(prefix="/crypto", tags=["加密货币"])


@router.get("/realtime")
async def realtime(symbol: str):
    try:
        quote = await crypto.get_realtime(symbol)
        return ApiResponse(data=quote.model_dump())
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/kline")
async def kline(symbol: str, period: str = "daily", count: int = 60):
    try:
        bars = await crypto.get_kline(symbol, period, count)
        return ApiResponse(data=[b.model_dump() for b in bars])
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/search")
async def search(keyword: str):
    try:
        results = await crypto.search_stock(keyword)
        return ApiResponse(data=[r.model_dump() for r in results])
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
