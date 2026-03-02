import os
from contextlib import contextmanager

PROXY = "socks5h://192.168.208.1:10808"

PROXY_KEYS = [
    "http_proxy", "https_proxy", "HTTP_PROXY", "HTTPS_PROXY",
    "all_proxy", "ALL_PROXY",
]


def set_proxy_env():
    """Set proxy environment variables for tools like yfinance."""
    for key in PROXY_KEYS:
        os.environ[key] = PROXY


@contextmanager
def no_proxy():
    """Temporarily remove proxy env vars (for domestic APIs like sina/akshare)."""
    saved = {}
    for key in PROXY_KEYS:
        if key in os.environ:
            saved[key] = os.environ.pop(key)
    try:
        yield
    finally:
        os.environ.update(saved)


def ccxt_proxies() -> dict:
    """Return proxies dict for ccxt constructor."""
    return {
        "http": PROXY,
        "https": PROXY,
    }
