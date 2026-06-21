from __future__ import annotations

import json
from urllib import error, request

from .config import Config


class OllamaClient:
    def __init__(self, config: Config) -> None:
        self.base_url = config.ollama_url.rstrip("/")
        self.model = config.ollama_model

    def health(self) -> bool:
        try:
            req = request.Request(f"{self.base_url}/api/tags", method="GET")
            with request.urlopen(req, timeout=3) as response:
                return response.status == 200
        except Exception:
            return False

    def generate(self, prompt: str) -> str | None:
        payload = json.dumps(
            {
                "model": self.model,
                "prompt": prompt,
                "stream": False,
            }
        ).encode()
        req = request.Request(
            f"{self.base_url}/api/generate",
            data=payload,
            headers={"Content-Type": "application/json"},
            method="POST",
        )
        try:
            with request.urlopen(req, timeout=20) as response:
                data = json.loads(response.read().decode())
                return str(data.get("response", "")).strip()
        except error.URLError:
            return None
