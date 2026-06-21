Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$RootDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$AgentDir = Join-Path $RootDir "agent"
$EnvFile = Join-Path $AgentDir ".env"
$EnvExample = Join-Path $AgentDir ".env.example"
$LogDir = Join-Path $AgentDir "logs"

New-Item -ItemType Directory -Path $LogDir -Force | Out-Null

function Write-Info($Message) {
    Write-Host "[windows] $Message" -ForegroundColor Cyan
}

function Write-WarnMsg($Message) {
    Write-Host "[windows] $Message" -ForegroundColor Yellow
}

function Fail($Message) {
    Write-Host "[windows] $Message" -ForegroundColor Red
    exit 1
}

function Require-Command($Name) {
    if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
        Fail "Comando obrigatorio ausente: $Name"
    }
}

function Copy-EnvIfMissing {
    if (-not (Test-Path $EnvFile)) {
        Copy-Item $EnvExample $EnvFile
        Write-Info "Arquivo $EnvFile criado a partir de $EnvExample"
    }
}

function Get-EnvValue($Key) {
    $line = Get-Content $EnvFile | Where-Object { $_ -match "^$Key=" } | Select-Object -First 1
    if ($null -eq $line) { return "" }
    return ($line -replace "^[^=]+=", "")
}

function Set-EnvValue($Key, $Value) {
    $lines = @()
    $done = $false
    foreach ($line in Get-Content $EnvFile) {
        if ($line -match "^$Key=") {
            $lines += "$Key=$Value"
            $done = $true
        } else {
            $lines += $line
        }
    }
    if (-not $done) {
        $lines += "$Key=$Value"
    }
    Set-Content -Path $EnvFile -Value $lines
}

function Prompt-IfEmpty($Key, $Label) {
    $current = Get-EnvValue $Key
    if ([string]::IsNullOrWhiteSpace($current)) {
        $current = Read-Host $Label
        if ([string]::IsNullOrWhiteSpace($current)) {
            Fail "Valor obrigatorio nao informado para $Key"
        }
        Set-EnvValue $Key $current
    }
}

function Load-Env {
    foreach ($line in Get-Content $EnvFile) {
        if ($line -match "^\s*$") { continue }
        if ($line.StartsWith("#")) { continue }
        $parts = $line.Split("=", 2)
        if ($parts.Count -eq 2) {
            [Environment]::SetEnvironmentVariable($parts[0], $parts[1], "Process")
        }
    }
}

function Ensure-OllamaInstalled {
    if (Get-Command ollama -ErrorAction SilentlyContinue) {
        Write-Info "Ollama encontrado."
        return
    }
    Write-WarnMsg "Ollama nao encontrado. Instale manualmente em https://ollama.com/download/windows e rode este script de novo."
    Fail "Ollama ausente."
}

function Ensure-OllamaRunning($ApiUrl) {
    $baseUrl = $ApiUrl -replace "/api.*$", ""
    try {
        Invoke-WebRequest -UseBasicParsing -Uri "$baseUrl/api/tags" | Out-Null
        Write-Info "Ollama API ja esta respondendo em $baseUrl"
        return
    } catch {
        Write-WarnMsg "Ollama nao esta respondendo em $baseUrl. Tentando iniciar 'ollama serve'."
    }

    $logFile = Join-Path $LogDir "ollama.log"
    Start-Process -FilePath "ollama" -ArgumentList "serve" -RedirectStandardOutput $logFile -RedirectStandardError $logFile -WindowStyle Hidden

    for ($i = 0; $i -lt 20; $i++) {
        Start-Sleep -Seconds 1
        try {
            Invoke-WebRequest -UseBasicParsing -Uri "$baseUrl/api/tags" | Out-Null
            Write-Info "Ollama inicializado com sucesso."
            return
        } catch {
        }
    }

    Fail "Ollama nao respondeu apos a inicializacao. Veja $logFile"
}

function Ensure-ModelInstalled($Model) {
    & ollama show $Model *> $null
    if ($LASTEXITCODE -eq 0) {
        Write-Info "Modelo $Model ja instalado."
        return
    }
    Write-WarnMsg "Modelo $Model nao encontrado. Executando 'ollama pull $Model'."
    & ollama pull $Model
}

function Print-Tutorial {
@"

EditalBox Windows

O que este script faz:
- cria agent/.env se faltar
- carrega as variaveis de ambiente
- verifica se o Ollama esta instalado
- inicia o Ollama se ele nao estiver rodando
- baixa o modelo configurado se faltar
- inicia o agent Python local

Links uteis:
- Tutorial completo: $RootDir\docs\startup.md
- Download do Ollama: https://ollama.com/download/windows
- Biblioteca de modelos: https://ollama.com/library

Logs:
- Ollama: $LogDir\ollama.log

"@ | Write-Host
}

Copy-EnvIfMissing
Prompt-IfEmpty "EDITALBOX_AGENT_OLLAMA_MODEL" "Modelo Ollama desejado"
Load-Env

Require-Command python
Ensure-OllamaInstalled
Ensure-OllamaRunning $env:EDITALBOX_AGENT_OLLAMA_URL
Ensure-ModelInstalled $env:EDITALBOX_AGENT_OLLAMA_MODEL
Print-Tutorial

Write-Info "Iniciando agent em http://$($env:EDITALBOX_AGENT_HOST):$($env:EDITALBOX_AGENT_PORT)"
Push-Location $AgentDir
$env:PYTHONPATH = ".\src"
python -m agent.main
Pop-Location
