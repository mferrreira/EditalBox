# EditalBox

Monorepo do EditalBox, dividido em dois componentes:

- `tvbox/`: serviço principal em Go para coleta, armazenamento local, bot do Telegram e orquestração.
- `agent/`: serviço local em Python para busca contextual, ranqueamento e geração de resposta com Ollama.

## Estrutura

- `docs/architecture.md`: arquitetura, fontes de coleta e roadmap.
- `docs/startup.md`: tutorial operacional do MacBook e da TV Box.
- `tvbox/`: aplicação para a TV Box.
- `agent/`: serviço local para o MacBook.
- `deploy/systemd/`: unidades `systemd` de referência.
- `start-mac.sh`: bootstrap e inicialização do agent no MacBook.
- `start-linux.sh`: bootstrap e inicialização do agent em Linux.
- `start-tvbox.sh`: bootstrap e inicialização do serviço principal na TV Box.
- `start-windows.ps1`: bootstrap e inicialização do agent em Windows.

## Objetivo da fase atual

Esta base entrega a infraestrutura inicial do sistema, com contratos entre os serviços e um pipeline funcional para:

- coletar páginas e links candidatos do portal do IFNMG;
- armazenar editais, documentos e eventos em SQLite na TV Box;
- sincronizar com um agent local;
- responder comandos do Telegram;
- executar busca contextual no MacBook com fallback lexical e integração opcional com Ollama.

## Execução rápida

### 1. MacBook

```bash
cd /Users/marcio/Projetos/EditalBox
chmod +x start-mac.sh
./start-mac.sh
```

### 2. TV Box

```bash
cd /Users/marcio/Projetos/EditalBox
chmod +x start-tvbox.sh
./start-tvbox.sh
```

## Observações

- O `tvbox` usa SQLite local e foi modelado para ser o nó principal.
- O `agent` mantém um índice local separado e aceita reindexação incremental.
- Nesta fase inicial, o ranqueamento local do agent é lexical/determinístico; se o Ollama estiver disponível, ele é usado para gerar a resposta final.
- O portal do IFNMG apresentou bloqueio `403` para clientes simples; por isso o coletor usa cabeçalhos de navegador e uma estratégia dirigida por seeds conhecidas.
- O guia operacional completo está em [docs/startup.md](/Users/marcio/Projetos/EditalBox/docs/startup.md).
