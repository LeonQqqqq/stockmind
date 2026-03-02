import time
import functools
from typing import Callable


class TTLCache:
    """Simple TTL cache decorator."""

    def __init__(self, ttl_seconds: int = 60):
        self.ttl = ttl_seconds
        self._cache: dict = {}

    def __call__(self, func: Callable) -> Callable:
        @functools.wraps(func)
        async def wrapper(*args, **kwargs):
            key = (func.__name__, args, tuple(sorted(kwargs.items())))
            now = time.time()
            if key in self._cache:
                result, ts = self._cache[key]
                if now - ts < self.ttl:
                    return result
            result = await func(*args, **kwargs)
            self._cache[key] = (result, now)
            return result
        return wrapper
