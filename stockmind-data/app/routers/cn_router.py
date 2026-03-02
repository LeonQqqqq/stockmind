from fastapi import APIRouter, HTTPException
from app.models.schemas import ApiResponse
from app.services import cn_stock

router = APIRouter(prefix="/cn", tags=["A股"])


@router.get("/realtime")
async def realtime(symbol: str):
    try:
        quote = await cn_stock.get_realtime(symbol)
        return ApiResponse(data=quote.model_dump())
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/kline")
async def kline(symbol: str, period: str = "daily", count: int = 60):
    try:
        bars = await cn_stock.get_kline(symbol, period, count)
        return ApiResponse(data=[b.model_dump() for b in bars])
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/money_flow")
async def money_flow(symbol: str):
    try:
        flows = await cn_stock.get_money_flow(symbol)
        return ApiResponse(data=[f.model_dump() for f in flows])
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/search")
async def search(keyword: str):
    try:
        results = await cn_stock.search_stock(keyword)
        return ApiResponse(data=[r.model_dump() for r in results])
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
