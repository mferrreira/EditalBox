# 06. Sequência de Sistema

## Coletar edital (CU-001)
1. Serviço consulta seeds.
2. Sistema extrai páginas/links.
3. Sistema normaliza conteúdo.
4. Sistema deduplica por URL/conteúdo.
5. Sistema persiste notices/docs/events.

## Consulta NLP (CU-005/CU-006)
1. Usuário envia pergunta.
2. Bot transfere a agent.
3. Agent consulta índice de chunks.
4. Agent ranqueia contexto.
5. Agent gera resposta com fontes.
6. Bot responde usuário.

## Sync incremental (CU-002)
1. Job executa delta.
2. Sistema compara com base atual.
3. Sistema insere/atualiza notices.
4. Sistema atualiza índice quando necessário.

## Health/base status (CU-008)
1. Usuário/administrador consulta status.
2. Sistema retorna contadores e lista recente.
