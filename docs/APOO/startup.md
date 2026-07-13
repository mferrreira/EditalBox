# Startup e Operacao

## Visao geral

Use os scripts na raiz do projeto:

- `start-mac.sh`: inicia o serviço auxiliar no MacBook.
- `start-linux.sh`: inicia o serviço auxiliar em Linux.
- `start-tvbox.sh`: inicia o servico principal na TV Box.
- `install-tvbox.sh`: instala a TV Box em modo de producao com `systemd`.
- `start-windows.ps1`: inicia o serviço auxiliar no Windows.

Os scripts preparam o arquivo `.env`, validam dependências e iniciam os processos necessários.

## MacBook

### O que o script faz

`start-mac.sh`

- cria `agent/.env` se ele ainda nao existir;
- carrega as variaveis do serviço auxiliar;
- verifica se o `ollama` esta instalado;
- se o `ollama` nao existir:
  - tenta `brew install --cask ollama`;
  - se `brew` nao existir, tenta o instalador oficial `curl -fsSL https://ollama.com/install.sh | sh`;
- verifica se a API do Ollama esta respondendo;
- se necessario, sobe `ollama serve`;
- verifica se o modelo definido em `EDITALBOX_AGENT_OLLAMA_MODEL` existe;
- se faltar, executa `ollama pull <modelo>`;
- inicia o serviço auxiliar em Python.

### Como usar

```bash
git clone git@github.com:mferrreira/EditalBox.git
cd EditalBox
chmod +x start-mac.sh
./start-mac.sh
```

### Links uteis

- Download do Ollama: [ollama.com/download/mac](https://ollama.com/download/mac)
- Biblioteca de modelos: [ollama.com/library](https://ollama.com/library)

## Linux

### Como usar

```bash
cd EditalBox
chmod +x start-linux.sh
./start-linux.sh
```

### O que muda no Linux

- usa o instalador oficial `curl -fsSL https://ollama.com/install.sh | sh`;
- sobe `ollama serve` em background se a API nao estiver respondendo;
- usa o mesmo fluxo de `.env` e `ollama pull` do script do Mac.

Links uteis:

- Ollama Linux: [ollama.com/download/linux](https://ollama.com/download/linux)
- Biblioteca de modelos: [ollama.com/library](https://ollama.com/library)

## Windows

### Como usar

No PowerShell:

```powershell
cd C:\caminho\para\EditalBox
powershell -ExecutionPolicy Bypass -File .\start-windows.ps1
```

### Observacao sobre o Ollama no Windows

No Windows, o script valida a presenca do `ollama`, mas nao tenta instalar sozinho. Se ele nao estiver presente, instale primeiro por:

- [ollama.com/download/windows](https://ollama.com/download/windows)

Depois rode o script novamente. O restante do fluxo continua automático: iniciar `ollama serve`, validar o modelo e executar `ollama pull` quando necessário.

## Telegram

### Criar o bot

1. Abra o [@BotFather](https://t.me/BotFather) no Telegram.
2. Rode `/newbot`.
3. Escolha nome e username.
4. Guarde o token gerado. Ele vai para `EDITALBOX_TELEGRAM_TOKEN`.

Referencias oficiais:

- Guia geral de bots: [core.telegram.org/bots](https://core.telegram.org/bots)
- Tutorial do BotFather: [core.telegram.org/bots/tutorial](https://core.telegram.org/bots/tutorial)

### Descobrir o `chat_id`

1. Inicie conversa com o bot e envie qualquer mensagem.
2. Depois chame:

```bash
curl "https://api.telegram.org/bot<SEU_TOKEN>/getUpdates"
```

3. Procure o campo `message.chat.id`.
4. Use esse valor em `EDITALBOX_TELEGRAM_ALLOWED_CHAT_IDS`.

## TV Box

### O que o script faz

`start-tvbox.sh`

- cria `tvbox/.env` se ele ainda nao existir;
- pede o token do Telegram se ele estiver vazio;
- pede o `chat_id` permitido se ele estiver vazio;
- pede a URL do serviço auxiliar no MacBook se a configuracao ainda estiver em `localhost`;
- valida a presenca do Go;
- testa o endpoint `/health` do serviço auxiliar, sem bloquear a subida se ele estiver temporariamente fora do ar;
- inicia o servico principal em Go.

### Como usar

```bash
cd EditalBox
chmod +x start-tvbox.sh
./start-tvbox.sh
```

### Instalacao com systemd

Para instalar em modo persistente na TV Box:

```bash
cd EditalBox
chmod +x install-tvbox.sh
./install-tvbox.sh
```

Esse instalador:

- compila o binario Go;
- instala o projeto em `/opt/editalbox/tvbox`;
- copia o `.env`;
- instala a unit `deploy/systemd/editalbox-tvbox.service`;
- roda `systemctl daemon-reload`;
- executa `systemctl enable --now editalbox-tvbox`.

### URL correta do serviço auxiliar

A TV Box deve apontar para o IP do MacBook na rede local, por exemplo:

```text
http://192.168.0.10:8090
```

No MacBook, voce pode descobrir o IP local com:

```bash
ipconfig getifaddr en0
```

ou, em algumas interfaces:

```bash
ipconfig getifaddr en1
```

## Ordem recomendada

1. No MacBook, execute `./start-mac.sh`.
2. Confirme que `http://IP_DO_MAC:8090/health` responde na rede local.
3. Na TV Box, execute `./start-tvbox.sh`.
4. Envie `/status` para o bot no Telegram.

## Arquivos de configuracao

- `agent/.env.example`
- `tvbox/.env.example`

## Observacao importante

Os scripts atendem à execução local e aos procedimentos de instalação. Para operação contínua na TV Box, a recomendação é utilizar as unidades `systemd` em `deploy/systemd`.
