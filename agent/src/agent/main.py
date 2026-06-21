from __future__ import annotations

import json
from http import HTTPStatus
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

from .config import Config
from .ollama import OllamaClient
from .service import AgentService
from .store import Store


def main() -> None:
    config = Config.load()
    store = Store(config.db_path)
    service = AgentService(store, OllamaClient(config))

    class Handler(BaseHTTPRequestHandler):
        def do_GET(self) -> None:  # noqa: N802
            if self.path == "/health":
                self._json(
                    HTTPStatus.OK,
                    {
                        "status": "ok",
                        "ollama_ready": service.ollama.health(),
                        "indexed_count": store.indexed_count(),
                    },
                )
                return
            self._json(HTTPStatus.NOT_FOUND, {"error": "not found"})

        def do_POST(self) -> None:  # noqa: N802
            length = int(self.headers.get("Content-Length", "0"))
            payload = json.loads(self.rfile.read(length) or b"{}")
            if self.path == "/v1/ingest":
                indexed, chunks = service.ingest(list(payload.get("notices", [])))
                self._json(HTTPStatus.OK, {"indexed": indexed, "chunks": chunks})
                return
            if self.path == "/v1/answer":
                answer = service.answer(
                    question=str(payload.get("question", "")),
                    session_summary=str(payload.get("session_summary", "")),
                    candidates=list(payload.get("candidates", [])),
                    limit=int(payload.get("limit", 5)),
                )
                self._json(HTTPStatus.OK, answer)
                return
            self._json(HTTPStatus.NOT_FOUND, {"error": "not found"})

        def log_message(self, format: str, *args) -> None:  # noqa: A003
            return

        def _json(self, status: HTTPStatus, payload: dict) -> None:
            body = json.dumps(payload).encode()
            self.send_response(status)
            self.send_header("Content-Type", "application/json")
            self.send_header("Content-Length", str(len(body)))
            self.end_headers()
            self.wfile.write(body)

    server = ThreadingHTTPServer((config.host, config.port), Handler)
    print(f"agent listening on http://{config.host}:{config.port}")
    try:
        server.serve_forever()
    finally:
        store.close()


if __name__ == "__main__":
    main()
