# 08. Contratos OCL

Contexto: validar invariantes de coleta e consulta local.

## Edital
- def: Edital.url
  - pre: url not empty and origem valida
  - post: url persistida
  - exception: rejeitar notice sem origem referenciável.

- def: Edital.duplicidade
  - pre: true
  - post: url/título/conteúdo únicos por base
  - exception: notice duplicada não deve ser inserida.

## Indexação
- def: Indice.chunk
  - pre: texto não vazio
  - post: chunk referenciado por notice
  - exception: chunk órfão não deve ser indexado.

## Telegram
- def: Sessao.telegram
  - pre: session_id válido
  - post: contexto temporário mantido
  - exception: sessão sem identificador não deve ser usada.

## Logs
- def: Log.sync_run
  - pre: true
  - post: registrado status/delta quando relevante
  - exception: logging não pode mascarar falhas.
