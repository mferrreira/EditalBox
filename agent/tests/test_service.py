from __future__ import annotations

import tempfile
import unittest

from agent.ollama import OllamaClient
from agent.service import AgentService, looks_safe
from agent.store import Store


class FakeOllama:
    def __init__(self, healthy: bool = False, response: str | None = None) -> None:
        self._healthy = healthy
        self._response = response

    def health(self) -> bool:
        return self._healthy

    def generate(self, prompt: str) -> str | None:
        return self._response


class AgentServiceTest(unittest.TestCase):
    def setUp(self) -> None:
        self.tempdir = tempfile.TemporaryDirectory()
        self.store = Store(f"{self.tempdir.name}/agent.db")

    def tearDown(self) -> None:
        self.store.close()
        self.tempdir.cleanup()

    def test_prefers_bolsa_candidate_for_bolsa_question(self) -> None:
        service = AgentService(self.store, FakeOllama(False))
        result = service.answer(
            question="quais editais de bolsa estao abertos?",
            session_summary="",
            limit=5,
            candidates=[
                {
                    "id": 1,
                    "title": "Edital do Programa de Bolsa Treinamento 2026",
                    "url": "https://example.com/bolsa",
                    "status": "open",
                    "excerpt": "bolsa treinamento para estudantes",
                    "body_text": "inscricoes abertas para bolsa treinamento",
                },
                {
                    "id": 2,
                    "title": "Concurso Professor Substituto",
                    "url": "https://example.com/professor",
                    "status": "in_progress",
                    "excerpt": "processo seletivo para professor",
                    "body_text": "resultado preliminar",
                },
            ],
        )
        self.assertTrue(result["structured"])
        self.assertEqual(result["structured"][0]["url"], "https://example.com/bolsa")

    def test_rejects_suspicious_model_output(self) -> None:
        service = AgentService(self.store, FakeOllama(True, "中文输出 mixed output"))
        result = service.answer(
            question="quais editais de bolsa estao abertos?",
            session_summary="",
            limit=5,
            candidates=[
                {
                    "id": 1,
                    "title": "Edital do Programa de Bolsa Treinamento 2026",
                    "url": "https://example.com/bolsa",
                    "status": "open",
                    "excerpt": "bolsa treinamento para estudantes",
                    "body_text": "inscricoes abertas para bolsa treinamento",
                }
            ],
        )
        self.assertFalse(result["used_ollama"])
        self.assertIn("Selecionei os editais", result["text"])


class LooksSafeTest(unittest.TestCase):
    def test_detects_suspicious_text(self) -> None:
        self.assertFalse(looks_safe("中文输出 mixed output"))
        self.assertTrue(looks_safe("O edital de bolsa treinamento esta com inscricoes abertas."))


if __name__ == "__main__":
    unittest.main()
