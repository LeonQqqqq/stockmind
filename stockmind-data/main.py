import os
import certifi
import warnings
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from app.utils.proxy import set_proxy_env

# SSL cert for httpx/requests
os.environ["SSL_CERT_FILE"] = certifi.where()
os.environ["REQUESTS_CA_BUNDLE"] = certifi.where()

# Suppress InsecureRequestWarning for sina APIs
warnings.filterwarnings("ignore", message="Unverified HTTPS request")

# Set proxy env vars (for yfinance etc.)
set_proxy_env()

app = FastAPI(title="StockMind Data Service", version="1.0.0")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["*"],
    allow_headers=["*"],
)

from app.routers import cn_router, us_router, crypto_router

app.include_router(cn_router.router, prefix="/api/v1/stock")
app.include_router(us_router.router, prefix="/api/v1/stock")
app.include_router(crypto_router.router, prefix="/api/v1/stock")


@app.get("/health")
async def health():
    return {"status": "ok"}
