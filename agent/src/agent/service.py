from __future__ import annotations

from dataclasses import asdict, dataclass

from .ollama import OllamaClient
from .store import IndexedNotice, Store


@dataclass(slots=True)
class AnswerItem:
    title: str
    url: str
    status: str
    justification: str


class AgentService:
    def __init__(self, store: Store, ollama: OllamaClient) -> None:
        self.store = store
        self.ollama = ollama

    def ingest(self, payload: list[dict]) -> tuple[int, int]:
        notices = [
            IndexedNotice(
                notice_id=int(item["id"]),
                title=str(item.get("title", "")),
                url=str(item.get("url", "")),
                status=str(item.get("status", "unknown")),
                excerpt=str(item.get("excerpt", "")),
                body_text=str(item.get("body_text", "")),
                updated_at=str(item.get("updated_at", "")),
            )
            for item in payload
        ]
        return self.store.upsert_notices(notices)

    def answer(self, question: str, session_summary: str, candidates: list[dict], limit: int) -> dict:
        ranked = self._rank(question, candidates, limit)
        structured = [
            AnswerItem(
                title=item["title"],
                url=item["url"],
                status=item["status"],
                justification=item["justification"],
            )
            for item in ranked
        ]
        text = self._compose_text(question, session_summary, structured)
        used_ollama = False
        if self.ollama.health():
            generated = self.ollama.generate(self._prompt(question, session_summary, structured))
            if generated and looks_safe(generated):
                text = generated
                used_ollama = True
        return {
            "text": text,
            "structured": [asdict(item) for item in structured],
            "used_ollama": used_ollama,
        }

    def _rank(self, question: str, candidates: list[dict], limit: int) -> list[dict]:
        q_tokens = expanded_query_tokens(question, session_summary="")
        scored = []
        for item in candidates:
            text = " ".join(
                [
                    str(item.get("title", "")),
                    str(item.get("excerpt", "")),
                    str(item.get("body_text", "")),
                ]
            ).lower()
            tokens = set(tokenize(text))
            overlap = len(q_tokens & tokens)
            score = overlap * 4
            score += intent_bonus(question, item)
            if score <= 0:
                continue
            if item.get("status") == "open":
                score += 3
            if item.get("status") == "in_progress":
                score += 1
            scored.append(
                (
                    score,
                    {
                        "title": item.get("title", ""),
                        "url": item.get("url", ""),
                        "status": item.get("status", "unknown"),
                        "justification": self._justify(question, item),
                    },
                )
            )
        if not scored:
            db_hits = self.store.search(question, limit)
            for row in db_hits:
                scored.append(
                    (
                        1,
                        {
                            "title": row["title"],
                            "url": row["url"],
                            "status": row["status"],
                            "justification": "Correspondencia por termos expandidos e texto indexado.",
                        },
                    )
                )
        if not scored and candidates:
            for item in candidates[:limit]:
                scored.append(
                    (
                        fallback_candidate_score(item),
                        {
                            "title": item.get("title", ""),
                            "url": item.get("url", ""),
                            "status": item.get("status", "unknown"),
                            "justification": "Selecionado como melhor candidato disponivel na base local.",
                        },
                    )
                )
        scored.sort(key=lambda pair: pair[0], reverse=True)
        return [item for _, item in scored[:limit]]

    def _compose_text(self, question: str, session_summary: str, structured: list[AnswerItem]) -> str:
        if not structured:
            return "Nao encontrei editais compativeis com a pergunta na base indexada."
        if session_summary.strip():
            return "Considerei a pergunta atual e o contexto recente da conversa para selecionar os editais mais aderentes."
        return "Selecionei os editais com maior aderencia textual ao pedido atual."

    def _justify(self, question: str, item: dict) -> str:
        title = str(item.get("title", ""))
        excerpt = str(item.get("excerpt", ""))
        query_tokens = expanded_query_tokens(question, session_summary="")
        text_tokens = set(tokenize(f"{title} {excerpt}"))
        common = sorted(query_tokens & text_tokens)
        if common:
            return f"Possui termos relevantes em comum com a pergunta: {', '.join(common[:5])}."
        bonus_reason = detect_intent_reason(question, item)
        if bonus_reason:
            return bonus_reason
        return "Foi mantido como candidato por proximidade textual com o contexto."

    def _prompt(self, question: str, session_summary: str, structured: list[AnswerItem]) -> str:
        lines = [
            "Voce esta respondendo sobre editais do IFNMG.",
            "Use apenas os itens fornecidos como fonte.",
            f"Pergunta: {question}",
            f"Contexto da sessao: {session_summary or 'sem contexto adicional'}",
            "Itens selecionados:",
        ]
        for item in structured:
            lines.append(f"- {item.title} | {item.status} | {item.url} | {item.justification}")
        lines.append("Responda em portugues com um paragrafo curto e objetivo.")
        return "\n".join(lines)


def tokenize(text: str) -> list[str]:
    stopwords = {
        "a", "as", "o", "os", "de", "da", "das", "do", "dos", "e", "ou", "um", "uma",
        "para", "por", "com", "sem", "que", "qual", "quais", "me", "mostrar", "mostra",
        "tem", "sobre", "quero", "preciso", "esta", "estao", "aberto", "abertos",
        "aberta", "abertas", "mais", "menos", "como",
    }
    current = []
    out = []
    for char in text.lower():
        if char.isalnum():
            current.append(char)
        else:
            if current:
                out.append("".join(current))
                current = []
    if current:
        out.append("".join(current))
    normalized = []
    for token in out:
        token = normalize_token(token)
        if token and token not in stopwords and len(token) >= 3:
            normalized.append(token)
    return normalized


def normalize_token(token: str) -> str:
    translation = str.maketrans(
        "áàâãäéèêëíìîïóòôõöúùûüç",
        "aaaaaeeeeiiiiooooouuuuc",
    )
    token = token.translate(translation)
    if len(token) > 4 and token.endswith("s"):
        token = token[:-1]
    return token


def expanded_query_tokens(question: str, session_summary: str) -> set[str]:
    tokens = set(tokenize(question))
    tokens.update(tokenize(session_summary))
    expanded = set(tokens)
    synonym_groups = {
        "bolsa": {"bolsa", "auxilio", "apoio", "permanencia", "moradia", "treinamento", "monitoria", "beneficio"},
        "estagio": {"estagio", "treinamento", "vaga", "vagas"},
        "professor": {"professor", "docente", "substituto", "temporario"},
        "inscricao": {"inscricao", "matricula", "prazo", "aberto", "aberta", "abertos", "abertas"},
        "resultado": {"resultado", "preliminar", "final", "recurso", "homologacao"},
        "pesquisa": {"pesquisa", "cientifica", "pibic", "tecnico"},
        "extensao": {"extensao", "projeto", "comunidade", "cultura"},
        "curso": {"curso", "tecnico", "graduacao", "fic", "seletivo"},
    }
    for group in synonym_groups.values():
        if tokens & group:
            expanded.update(group)
    return expanded


def intent_bonus(question: str, item: dict) -> int:
    text = " ".join(
        [str(item.get("title", "")), str(item.get("excerpt", "")), str(item.get("body_text", ""))]
    ).lower()
    q = " ".join(tokenize(question))
    bonus = 0
    if any(word in q for word in ["bolsa", "auxilio", "monitoria", "treinamento"]) and any(
        word in text for word in ["bolsa", "auxilio", "monitoria", "treinamento"]
    ):
        bonus += 6
    if any(word in q for word in ["professor", "docente", "substituto"]) and any(
        word in text for word in ["professor", "docente", "substituto"]
    ):
        bonus += 6
    if any(word in q for word in ["curso", "tecnico", "graduacao", "vestibular", "matricula"]) and any(
        word in text for word in ["curso", "tecnico", "graduacao", "vestibular", "matricula"]
    ):
        bonus += 4
    if any(word in q for word in ["resultado", "preliminar", "final", "recurso"]) and any(
        word in text for word in ["resultado", "preliminar", "final", "recurso"]
    ):
        bonus += 4
    if "aberto" in q or "inscricao" in q or "prazo" in q:
        if item.get("status") == "open":
            bonus += 5
        elif item.get("status") == "in_progress":
            bonus += 2
    return bonus


def detect_intent_reason(question: str, item: dict) -> str | None:
    text = " ".join(
        [str(item.get("title", "")), str(item.get("excerpt", "")), str(item.get("body_text", ""))]
    ).lower()
    q = " ".join(tokenize(question))
    if any(word in q for word in ["bolsa", "auxilio", "monitoria", "treinamento"]) and any(
        word in text for word in ["bolsa", "auxilio", "monitoria", "treinamento"]
    ):
        return "Foi priorizado por tratar de bolsa, auxilio, monitoria ou treinamento."
    if any(word in q for word in ["professor", "docente", "substituto"]) and any(
        word in text for word in ["professor", "docente", "substituto"]
    ):
        return "Foi priorizado por tratar de selecao para docente ou professor substituto."
    if any(word in q for word in ["resultado", "recurso", "homologacao"]) and any(
        word in text for word in ["resultado", "recurso", "homologacao"]
    ):
        return "Foi priorizado por corresponder a uma consulta de resultados ou etapas do cronograma."
    return None


def fallback_candidate_score(item: dict) -> int:
    status = str(item.get("status", "unknown"))
    if status == "open":
        return 5
    if status == "in_progress":
        return 3
    return 1


def looks_safe(text: str) -> bool:
    if len(text.strip()) < 24:
        return False
    suspicious = 0
    for char in text:
        code = ord(char)
        if 0x4E00 <= code <= 0x9FFF:
            suspicious += 1
    return suspicious == 0
