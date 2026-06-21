from __future__ import annotations

import os
from dataclasses import dataclass


@dataclass(slots=True)
class Config:
    host: str = "127.0.0.1"
    port: int = 8090
    db_path: str = "./data/agent.db"
    ollama_url: str = "http://127.0.0.1:11434"
    ollama_model: str = "qwen2.5:7b-instruct"

    @classmethod
    def load(cls) -> "Config":
        return cls(
            host=os.getenv("EDITALBOX_AGENT_HOST", "127.0.0.1"),
            port=int(os.getenv("EDITALBOX_AGENT_PORT", "8090")),
            db_path=os.getenv("EDITALBOX_AGENT_DB_PATH", "./data/agent.db"),
            ollama_url=os.getenv("EDITALBOX_AGENT_OLLAMA_URL", "http://127.0.0.1:11434"),
            ollama_model=os.getenv("EDITALBOX_AGENT_OLLAMA_MODEL", "qwen2.5:7b-instruct"),
        )
