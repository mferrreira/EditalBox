from __future__ import annotations

import json
import os
import sqlite3
from dataclasses import dataclass
from pathlib import Path
from typing import Iterable


@dataclass(slots=True)
class IndexedNotice:
    notice_id: int
    title: str
    url: str
    status: str
    excerpt: str
    body_text: str
    updated_at: str


class Store:
    def __init__(self, db_path: str) -> None:
        Path(os.path.dirname(db_path) or ".").mkdir(parents=True, exist_ok=True)
        self.conn = sqlite3.connect(db_path, check_same_thread=False)
        self.conn.row_factory = sqlite3.Row
        self._migrate()

    def close(self) -> None:
        self.conn.close()

    def _migrate(self) -> None:
        self.conn.executescript(
            """
            CREATE TABLE IF NOT EXISTS indexed_notices (
              notice_id INTEGER PRIMARY KEY,
              title TEXT NOT NULL,
              url TEXT NOT NULL,
              status TEXT NOT NULL,
              excerpt TEXT NOT NULL,
              body_text TEXT NOT NULL,
              updated_at TEXT NOT NULL
            );
            CREATE TABLE IF NOT EXISTS indexed_chunks (
              id INTEGER PRIMARY KEY AUTOINCREMENT,
              notice_id INTEGER NOT NULL,
              chunk_text TEXT NOT NULL,
              keywords_json TEXT NOT NULL,
              FOREIGN KEY(notice_id) REFERENCES indexed_notices(notice_id) ON DELETE CASCADE
            );
            """
        )
        self.conn.commit()

    def upsert_notices(self, notices: Iterable[IndexedNotice]) -> tuple[int, int]:
        indexed = 0
        chunks = 0
        for notice in notices:
            self.conn.execute(
                """
                INSERT INTO indexed_notices (notice_id, title, url, status, excerpt, body_text, updated_at)
                VALUES (?, ?, ?, ?, ?, ?, ?)
                ON CONFLICT(notice_id) DO UPDATE SET
                  title = excluded.title,
                  url = excluded.url,
                  status = excluded.status,
                  excerpt = excluded.excerpt,
                  body_text = excluded.body_text,
                  updated_at = excluded.updated_at
                """,
                (
                    notice.notice_id,
                    notice.title,
                    notice.url,
                    notice.status,
                    notice.excerpt,
                    notice.body_text,
                    notice.updated_at,
                ),
            )
            self.conn.execute("DELETE FROM indexed_chunks WHERE notice_id = ?", (notice.notice_id,))
            notice_chunks = list(self._chunk_notice(notice))
            for chunk_text, keywords in notice_chunks:
                self.conn.execute(
                    """
                    INSERT INTO indexed_chunks (notice_id, chunk_text, keywords_json)
                    VALUES (?, ?, ?)
                    """,
                    (notice.notice_id, chunk_text, json.dumps(sorted(keywords))),
                )
                chunks += 1
            indexed += 1
        self.conn.commit()
        return indexed, chunks

    def _chunk_notice(self, notice: IndexedNotice) -> Iterable[tuple[str, set[str]]]:
        text = f"{notice.title}\n{notice.excerpt}\n{notice.body_text}".strip()
        if not text:
            return []
        window = 900
        overlap = 150
        chunks = []
        start = 0
        while start < len(text):
            end = min(len(text), start + window)
            chunk = text[start:end]
            keywords = {token for token in tokenize(chunk) if len(token) >= 4}
            chunks.append((chunk, keywords))
            if end == len(text):
                break
            start = end - overlap
        return chunks

    def search(self, query: str, limit: int) -> list[sqlite3.Row]:
        rows = self.conn.execute(
            """
            SELECT n.notice_id, n.title, n.url, n.status, n.excerpt, n.body_text, c.chunk_text, c.keywords_json
            FROM indexed_notices n
            JOIN indexed_chunks c ON c.notice_id = n.notice_id
            """
        ).fetchall()
        scored: list[tuple[float, sqlite3.Row]] = []
        q_tokens = set(tokenize(query))
        for row in rows:
            keywords = set(json.loads(row["keywords_json"]))
            overlap = len(q_tokens & keywords)
            if overlap == 0:
                continue
            score = overlap * 10
            low_chunk = row["chunk_text"].lower()
            for token in q_tokens:
                if token and token in low_chunk:
                    score += 1
            scored.append((score, row))
        scored.sort(key=lambda item: item[0], reverse=True)
        dedup: dict[int, sqlite3.Row] = {}
        for _, row in scored:
            dedup.setdefault(int(row["notice_id"]), row)
            if len(dedup) >= limit:
                break
        return list(dedup.values())

    def indexed_count(self) -> int:
        row = self.conn.execute("SELECT COUNT(*) AS total FROM indexed_notices").fetchone()
        return int(row["total"])


def tokenize(text: str) -> list[str]:
    cleaned = []
    current = []
    for char in text.lower():
        if char.isalnum():
            current.append(char)
        else:
            if current:
                cleaned.append("".join(current))
                current = []
    if current:
        cleaned.append("".join(current))
    return [normalize_token(token) for token in cleaned if normalize_token(token)]


def normalize_token(token: str) -> str:
    translation = str.maketrans(
        "áàâãäéèêëíìîïóòôõöúùûüç",
        "aaaaaeeeeiiiiooooouuuuc",
    )
    token = token.translate(translation)
    if len(token) > 4 and token.endswith("s"):
        token = token[:-1]
    return token
