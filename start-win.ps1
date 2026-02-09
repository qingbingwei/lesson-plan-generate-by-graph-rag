[CmdletBinding()]
param(
  [switch]$SkipDocker
)

$ErrorActionPreference = 'Stop'

$ProjectDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$LogDir = Join-Path $env:TEMP 'lesson-plan'

$BackendLog = Join-Path $LogDir 'backend.log'
$AgentLog = Join-Path $LogDir 'agent.log'
$FrontendLog = Join-Path $LogDir 'frontend.log'

$BackendPidFile = Join-Path $LogDir 'backend.pid'
$AgentPidFile = Join-Path $LogDir 'agent.pid'
$FrontendPidFile = Join-Path $LogDir 'frontend.pid'
$EnvFile = Join-Path $ProjectDir '.env'

function Write-Title([string]$Text) {
  Write-Host '========================================' -ForegroundColor Green
  Write-Host "  $Text" -ForegroundColor Green
  Write-Host '========================================' -ForegroundColor Green
}

function Write-Step([string]$Text) {
  Write-Host "`n$Text" -ForegroundColor Yellow
}

function Write-Ok([string]$Text) {
  Write-Host "[OK] $Text" -ForegroundColor Green
}

function Write-WarnText([string]$Text) {
  Write-Host "[WARN] $Text" -ForegroundColor Yellow
}

function Test-CommandAvailable([string]$Name) {
  return $null -ne (Get-Command $Name -ErrorAction SilentlyContinue)
}

function Import-EnvFile([string]$FilePath) {
  if (-not (Test-Path $FilePath)) {
    Write-WarnText ".env file not found at $FilePath"
    return
  }

  Get-Content $FilePath | ForEach-Object {
    $line = $_.Trim()
    if (-not $line -or $line.StartsWith('#')) {
      return
    }

    $idx = $line.IndexOf('=')
    if ($idx -le 0) {
      return
    }

    $key = $line.Substring(0, $idx).Trim()
    $value = $line.Substring($idx + 1).Trim()

    if (($value.StartsWith('"') -and $value.EndsWith('"')) -or ($value.StartsWith("'") -and $value.EndsWith("'"))) {
      $value = $value.Substring(1, $value.Length - 2)
    }

    [System.Environment]::SetEnvironmentVariable($key, $value, 'Process')
  }
}

function Ensure-NodeDependencies([string]$Workdir, [string]$ServiceName) {
  $packageJson = Join-Path $Workdir 'package.json'
  if (-not (Test-Path $packageJson)) {
    return
  }

  $nodeModules = Join-Path $Workdir 'node_modules'
  if (Test-Path $nodeModules) {
    return
  }

  Write-Step -Text "Installing $ServiceName dependencies..."
  Push-Location $Workdir
  try {
    & npm install
    if ($LASTEXITCODE -ne 0) {
      throw "npm install failed for $ServiceName"
    }
  }
  finally {
    Pop-Location
  }
}

function Get-PortPids([int]$Port) {
  $connections = Get-NetTCPConnection -LocalPort $Port -State Listen -ErrorAction SilentlyContinue
  if (-not $connections) {
    return @()
  }
  return $connections | Select-Object -ExpandProperty OwningProcess -Unique
}

function Stop-PortProcess([int]$Port, [string]$ServiceName) {
  $processIds = Get-PortPids -Port $Port
  if (-not $processIds -or $processIds.Count -eq 0) {
    return
  }

  Write-WarnText "Port $Port is in use. Stopping $ServiceName..."
  foreach ($processId in $processIds) {
    try {
      Stop-Process -Id $processId -Force -ErrorAction Stop
    }
    catch {
      Write-WarnText "Cannot stop PID=$($processId): $($_.Exception.Message)"
    }
  }
  Start-Sleep -Seconds 1
}

function Stop-AgentLegacyProcesses([string]$RootPath) {
  $agentPath = Join-Path $RootPath 'agent'
  $escapedPath = [Regex]::Escape($agentPath)

  $targets = @(Get-CimInstance Win32_Process -Filter "Name='node.exe' OR Name='cmd.exe'" -ErrorAction SilentlyContinue |
    Where-Object {
      $_.CommandLine -and
      $_.CommandLine -match $escapedPath -and
      ($_.CommandLine -match 'ts-node-dev' -or $_.CommandLine -match 'src\\index\.ts' -or $_.CommandLine -match 'lesson-plan-agent')
    })

  foreach ($proc in $targets) {
    try {
      Stop-Process -Id $proc.ProcessId -Force -ErrorAction Stop
      Write-WarnText "Stopped stale agent process PID=$($proc.ProcessId)"
    }
    catch {
      Write-WarnText "Cannot stop stale agent process PID=$($proc.ProcessId)"
    }
  }
}

function Get-DockerComposeCommand {
  if (Test-CommandAvailable -Name 'docker') {
    try {
      docker compose version *> $null
      if ($LASTEXITCODE -eq 0) {
        return @('docker', 'compose')
      }
    }
    catch {
    }
  }

  if (Test-CommandAvailable -Name 'docker-compose') {
    return @('docker-compose')
  }

  return $null
}

function Invoke-DockerCompose([string[]]$Arguments) {
  $composeCommand = Get-DockerComposeCommand
  if (-not $composeCommand) {
    throw 'docker compose is not available. Please install Docker Desktop first.'
  }

  if ($composeCommand.Count -eq 2) {
    & $composeCommand[0] $composeCommand[1] @Arguments
  }
  else {
    & $composeCommand[0] @Arguments
  }

  if ($LASTEXITCODE -ne 0) {
    throw "docker compose failed: $($Arguments -join ' ')"
  }
}

function Start-BackgroundCommand(
  [string]$ServiceName,
  [string]$WorkingDirectory,
  [string]$Command,
  [string]$LogPath,
  [string]$PidPath,
  [hashtable]$EnvVars = @{}
) {
  if (Test-Path $LogPath) {
    Remove-Item $LogPath -Force -ErrorAction SilentlyContinue
  }

  if (Test-Path $LogPath) {
    Clear-Content -Path $LogPath -Force -ErrorAction SilentlyContinue
  }

  $setEnvCommands = @()
  foreach ($key in $EnvVars.Keys) {
    $value = [string]$EnvVars[$key]
    $setEnvCommands += "set `"$key=$value`""
  }

  $envPrefix = ''
  if ($setEnvCommands.Count -gt 0) {
    $envPrefix = ($setEnvCommands -join ' && ') + ' && '
  }

  $argument = "$envPrefix$Command >> `"$LogPath`" 2>&1"
  $process = Start-Process -FilePath 'cmd.exe' -ArgumentList '/c', $argument -WorkingDirectory $WorkingDirectory -WindowStyle Hidden -PassThru
  Set-Content -Path $PidPath -Value $process.Id -Encoding ASCII
  Write-Host "$ServiceName PID: $($process.Id)" -ForegroundColor DarkGray
}

function Stop-FrontendViteProcesses([string]$RootPath) {
  $frontendPath = Join-Path $RootPath 'frontend'
  $escapedPath = [Regex]::Escape($frontendPath)

  $targets = Get-CimInstance Win32_Process -Filter "Name='node.exe' OR Name='cmd.exe'" |
    Where-Object {
      $_.CommandLine -and
      $_.CommandLine -match 'vite' -and
      $_.CommandLine -match $escapedPath
    }

  foreach ($proc in $targets) {
    try {
      Stop-Process -Id $proc.ProcessId -Force -ErrorAction Stop
    }
    catch {
      Write-WarnText "Cannot stop old frontend process PID=$($proc.ProcessId)"
    }
  }
}

function Get-FrontendPortFromLog([string]$LogPath) {
  if (-not (Test-Path $LogPath)) {
    return $null
  }

  $line = Select-String -Path $LogPath -Pattern 'localhost:(\d+)' -AllMatches -ErrorAction SilentlyContinue |
    Select-Object -Last 1

  if (-not $line) {
    return $null
  }

  $match = [regex]::Match($line.Line, 'localhost:(\d+)')
  if ($match.Success) {
    return $match.Groups[1].Value
  }

  return $null
}

function Wait-ForPort([int]$Port, [int]$TimeoutSeconds, [string]$ServiceName) {
  $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
  while ((Get-Date) -lt $deadline) {
    if ((Get-PortPids -Port $Port).Count -gt 0) {
      return $true
    }
    Start-Sleep -Milliseconds 800
  }

  Write-WarnText "$ServiceName did not open port $Port within ${TimeoutSeconds}s."
  return $false
}

function Show-RecentLog([string]$LogPath, [string]$ServiceName, [int]$TailLines = 30) {
  if (-not (Test-Path $LogPath)) {
    return
  }

  Write-WarnText "$ServiceName log tail:"
  Get-Content -Path $LogPath -Tail $TailLines | ForEach-Object {
    Write-Host "  $_" -ForegroundColor DarkGray
  }
}

function Resolve-AgentPort {
  if ($env:AGENT_PORT -match '^\d+$') {
    return [int]$env:AGENT_PORT
  }

  if ($env:PORT -match '^\d+$') {
    return [int]$env:PORT
  }

  return 13001
}

New-Item -ItemType Directory -Path $LogDir -Force | Out-Null
Import-EnvFile -FilePath $EnvFile
$AgentPort = Resolve-AgentPort

Write-Title -Text 'Lesson Plan System - Start Services'

$AgentEnvPath = Join-Path (Join-Path $ProjectDir 'agent') '.env'
if (Test-Path $AgentEnvPath) {
  Write-WarnText "Detected deprecated file: $AgentEnvPath"
  Write-WarnText 'Environment variables are unified in root .env. Please remove agent/.env to avoid confusion.'
}

Write-Step -Text '[1/4] Check Docker containers...'
if ($SkipDocker) {
  Write-WarnText 'Docker startup skipped by parameter.'
}
else {
  if (-not (Test-CommandAvailable -Name 'docker')) {
    Write-WarnText 'Docker not found. Skip dependency containers.'
  }
  else {
    docker info *> $null
    if ($LASTEXITCODE -ne 0) {
      throw 'Docker Desktop is unavailable or permission denied. Please restart Docker Desktop and run PowerShell as Administrator.'
    }

    Write-WarnText 'Ensuring postgres/neo4j/redis containers are running...'
    Push-Location $ProjectDir
    try {
      Invoke-DockerCompose -Arguments @('up', '-d', 'postgres', 'neo4j', 'redis')
    }
    finally {
      Pop-Location
    }

    $postgresReady = Wait-ForPort -Port 5432 -TimeoutSeconds 45 -ServiceName 'PostgreSQL'
    $redisReady = Wait-ForPort -Port 6379 -TimeoutSeconds 45 -ServiceName 'Redis'
    $neo4jReady = Wait-ForPort -Port 17687 -TimeoutSeconds 90 -ServiceName 'Neo4j'
    if (-not ($postgresReady -and $redisReady -and $neo4jReady)) {
      throw 'Required dependency containers are not ready. Check `docker compose logs`.'
    }

    Write-Ok -Text 'Dependency containers are ready.'
  }
}

Write-Step -Text '[2/4] Start backend service (Go)...'
if (-not (Test-CommandAvailable -Name 'go')) {
  throw 'go command not found. Please install Go.'
}

if (-not (Wait-ForPort -Port 17687 -TimeoutSeconds 20 -ServiceName 'Neo4j')) {
  throw 'Neo4j is not available on localhost:17687. Please check Docker container status.'
}

Stop-PortProcess -Port 8080 -ServiceName 'backend service'
Start-BackgroundCommand -ServiceName 'backend' -WorkingDirectory (Join-Path $ProjectDir 'backend') -Command 'go run ./cmd/server/main.go' -LogPath $BackendLog -PidPath $BackendPidFile

if (Wait-ForPort -Port 8080 -TimeoutSeconds 15 -ServiceName 'Backend') {
  Write-Ok -Text 'Backend is running at http://localhost:8080'
}
else {
  Write-WarnText "Backend may have failed. Check log: $BackendLog"
  Show-RecentLog -LogPath $BackendLog -ServiceName 'Backend' -TailLines 40
}

Write-Step -Text '[3/4] Start agent service (Node.js)...'
if (-not (Test-CommandAvailable -Name 'npm')) {
  throw 'npm command not found. Please install Node.js.'
}

Ensure-NodeDependencies -Workdir (Join-Path $ProjectDir 'agent') -ServiceName 'agent'

Stop-PortProcess -Port $AgentPort -ServiceName 'agent service'
Stop-PortProcess -Port 3001 -ServiceName 'legacy agent service (port 3001)'
Stop-AgentLegacyProcesses -RootPath $ProjectDir
Start-BackgroundCommand -ServiceName 'agent' -WorkingDirectory (Join-Path $ProjectDir 'agent') -Command 'npm run dev' -LogPath $AgentLog -PidPath $AgentPidFile -EnvVars @{ PORT = "$AgentPort"; AGENT_PORT = "$AgentPort" }

if (Wait-ForPort -Port $AgentPort -TimeoutSeconds 20 -ServiceName 'Agent') {
  Write-Ok -Text "Agent is running at http://localhost:$AgentPort"
}
else {
  Write-WarnText "Agent may have failed. Check log: $AgentLog"
  Show-RecentLog -LogPath $AgentLog -ServiceName 'Agent' -TailLines 40
}

Write-Step -Text '[4/4] Start frontend service (Vite)...'
Ensure-NodeDependencies -Workdir (Join-Path $ProjectDir 'frontend') -ServiceName 'frontend'
Stop-FrontendViteProcesses -RootPath $ProjectDir
Start-Sleep -Milliseconds 800
Start-BackgroundCommand -ServiceName 'frontend' -WorkingDirectory (Join-Path $ProjectDir 'frontend') -Command 'npm run dev' -LogPath $FrontendLog -PidPath $FrontendPidFile

Start-Sleep -Seconds 4
$frontendPort = Get-FrontendPortFromLog -LogPath $FrontendLog
if ($frontendPort) {
  Write-Ok -Text "Frontend is running at http://localhost:$frontendPort"
}
else {
  Write-WarnText 'Frontend is still starting. Check log in a few seconds.'
  $frontendPort = '5173'
}

Write-Host ''
Write-Host '========================================' -ForegroundColor Green
Write-Host '  All services started' -ForegroundColor Green
Write-Host '========================================' -ForegroundColor Green
Write-Host ''
Write-Host "Frontend: http://localhost:$frontendPort" -ForegroundColor Green
Write-Host 'Backend:  http://localhost:8080' -ForegroundColor Green
Write-Host "Agent:    http://localhost:$AgentPort" -ForegroundColor Green
Write-Host ''
Write-Host "Logs: $LogDir"
Write-Host '  - backend.log'
Write-Host '  - agent.log'
Write-Host '  - frontend.log'
Write-Host ''
Write-Host 'Use .\stop-win.ps1 to stop all services.' -ForegroundColor Yellow
