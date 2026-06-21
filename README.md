# EditalBox

Sistema embarcado para coleta, manutenção e consulta inteligente de editais publicados no site do IFNMG Campus Januária.

## Visão geral

O EditalBox foi desenhado para executar em uma TV Box com Armbian como nó principal de coleta e armazenamento, mantendo o processamento pesado de linguagem natural fora desse equipamento. A TV Box coleta os editais, atualiza uma base local e responde consultas simples; um serviço auxiliar na rede local faz a busca contextual e a geração de respostas em linguagem natural.

Essa divisão atende aos requisitos centrais do projeto:

- manter os dados localmente na TV Box;
- operar com baixo consumo de memória e CPU;
- evitar dependência de APIs pagas;
- tolerar mudanças na estrutura HTML do portal;
- funcionar em rede local com um nó auxiliar de processamento.

## Objetivos funcionais

O sistema foi estruturado para cobrir estes objetivos do documento de análise:

- buscar editais publicados no site do IFNMG Campus Januária;
- armazenar localmente os editais encontrados;
- atualizar a base periodicamente;
- evitar duplicação;
- permitir consulta por comandos, palavras-chave e linguagem natural;
- retornar título, link e justificativa quando possível;
- marcar editais indisponíveis, encerrados ou finalizados;
- informar o estado atual da base local.

## Arquitetura

O monorepo está dividido em dois componentes.

### `tvbox/`

Serviço principal em Go, responsável por:

- coleta dirigida das páginas do IFNMG;
- extração genérica de links e normalização;
- persistência local em SQLite;
- deduplicação e atualização incremental;
- bot do Telegram com polling;
- busca textual local;
- orquestração da consulta inteligente.

### `agent/`

Serviço local em Python, responsável por:

- ingestão incremental dos editais novos ou alterados;
- indexação local por chunks;
- recuperação e ranqueamento contextual;
- integração com modelo local via Ollama;
- geração de resposta em linguagem natural com base em fontes locais.

## Projeto orientado a objetos e domínios

O desenho do projeto segue os domínios extraídos do documento de requisitos.

### Edital

Entidade central do sistema. No código, aparece principalmente em:

- `tvbox/internal/domain/models.go`
- tabela `notices`

Representa o edital com:

- título;
- URL de origem;
- resumo;
- texto bruto essencial;
- status;
- datas derivadas relevantes.

### Fonte de coleta

Representa os pontos de partida usados na sincronização. No código, fica concentrada em:

- `tvbox/internal/config/config.go`
- `tvbox/internal/collector/collector.go`

### Consulta

Representa a recuperação dos editais para listagem, busca textual e pergunta em linguagem natural. A orquestração está em:

- `tvbox/internal/app/app.go`
- `agent/src/agent/service.py`

### Usuário

Representa a interação conversacional via Telegram. O sistema mantém apenas sessão temporária, sem perfil persistente:

- `tvbox/internal/telegram/bot.go`
- tabelas `telegram_sessions` e `telegram_messages`

### Serviço de interpretação

Representa o nó auxiliar na rede local responsável pela interpretação das consultas:

- `agent/src/agent/main.py`
- `agent/src/agent/ollama.py`

## Estrutura do repositório

- `docs/architecture.md`: arquitetura técnica resumida.
- `docs/startup.md`: guia operacional do ambiente.
- `tvbox/`: aplicação principal em Go.
- `agent/`: serviço local em Python.
- `deploy/systemd/`: unidades `systemd`.
- `start-mac.sh`: inicialização do serviço auxiliar no macOS.
- `start-linux.sh`: inicialização do serviço auxiliar em Linux.
- `start-windows.ps1`: inicialização do serviço auxiliar em Windows.
- `start-tvbox.sh`: inicialização do serviço principal em ambiente local.
- `install-tvbox.sh`: instalação de produção da TV Box com `systemd`.

## Modos de execução

### Desenvolvimento local

Suba o agent:

```bash
git clone git@github.com:mferrreira/EditalBox.git
cd EditalBox
chmod +x start-mac.sh
./start-mac.sh
```

Depois suba a TV Box localmente:

```bash
cd EditalBox
chmod +x start-tvbox.sh
./start-tvbox.sh
```

### Produção na TV Box

Para instalar o serviço na TV Box com `systemd`:

```bash
cd EditalBox
chmod +x install-tvbox.sh
./install-tvbox.sh
```

Esse instalador:

- garante a presença do `.env`;
- pede token do Telegram e URL do agent, se necessário;
- compila o binário Go;
- instala arquivos em `/opt/editalbox/tvbox`;
- instala a unit `systemd`;
- executa `systemctl daemon-reload`;
- faz `systemctl enable --now editalbox-tvbox`.

## Fontes de coleta priorizadas

As seeds principais do coletor são:

- `https://www.ifnmg.edu.br/januaria`
- `https://www.ifnmg.edu.br/editais-ifnmg`
- `https://www.ifnmg.edu.br/mais-noticias-januaria`
- `https://www.ifnmg.edu.br/assistenciaestudantil-januaria/editais-assistenciaestudantil-januaria`
- `https://www.ifnmg.edu.br/extensao-januaria/editais`
- `https://www.ifnmg.edu.br/pesquisa-januaria/pesquisa/editais-pesquisa-januaria`
- `https://www.ifnmg.edu.br/processoseletivo`
- `https://www.ifnmg.edu.br/professor-substituto`

O projeto evita depender de um único seletor HTML e trabalha com extração mais genérica de links e conteúdo.

## Dados e persistência

Na TV Box:

- SQLite local em `tvbox/data/editalbox.db` no modo dev;
- SQLite em `/opt/editalbox/tvbox/data/editalbox.db` no modo instalado;
- tabelas principais:
  - `notices`
  - `notice_documents`
  - `notice_events`
  - `sync_runs`
  - `telegram_sessions`
  - `telegram_messages`

No agent:

- índice local em `agent/data/agent.db`;
- tabelas principais:
  - `indexed_notices`
  - `indexed_chunks`

## Requisitos não funcionais atendidos pela arquitetura

- baixo consumo na TV Box: Go + SQLite + processamento local leve;
- open source: Go, Python, SQLite e Ollama;
- sem API paga obrigatória;
- armazenamento local dos dados coletados;
- operação em LAN com serviço auxiliar separado;
- tolerância a falhas no HTML por coleta dirigida e parsing genérico.

## Capacidades implementadas

O projeto já contempla:

- coleta funcional;
- persistência local;
- sincronização incremental com o agent;
- bot do Telegram;
- consulta textual;
- consulta em linguagem natural com apoio do agent;
- integração com modelo local via Ollama;
- scripts de inicialização para macOS, Linux e Windows;
- instalador de produção da TV Box com `systemd`.

## Documentação complementar

- `docs/architecture.md`
- `docs/startup.md`
