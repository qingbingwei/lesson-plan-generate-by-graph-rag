[CmdletBinding()]
param(
  [switch]$StopDocker
)

$ErrorActionPreference = 'Continue'

$ProjectDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$LogDir = Join-Path $env:TEMP 'lesson-plan'

$BackendPidFile = Join-Path $LogDir 'backend.pid'
$AgentPidFile = Join-Path $LogDir 'agent.pid'
$FrontendPidFile = Join-Path $LogDir 'frontend.pid'
$EnvFile = Join-Path $ProjectDir '.env'

function Write-Title([string]$Text) {
  Write-Host '========================================' -ForegroundColor Yellow
  Write-Host "  $Text" -ForegroundColor Yellow
  Write-Host '========================================' -ForegroundColor Yellow
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

function Get-PortPids([int]$Port) {
  $connections = Get-NetTCPConnection -LocalPort $Port -State Listen -ErrorAction SilentlyContinue
  if (-not $connections) {
    return @()
  }
  return $connections | Select-Object -ExpandProperty OwningProcess -Unique
}

function Stop-ProcessByPidFile([string]$PidFilePath, [string]$Name) {
  if (-not (Test-Path $PidFilePath)) {
    Write-WarnText "$Name pid file not found."
    return
  }

  $pidValue = Get-Content -Path $PidFilePath -ErrorAction SilentlyContinue | Select-Object -First 1
  if (-not $pidValue) {
    Remove-Item $PidFilePath -Force -ErrorAction SilentlyContinue
    Write-WarnText "$Name pid file is empty."
    return
  }

  try {
    $processId = [int]$pidValue
    $proc = Get-Process -Id $processId -ErrorAction SilentlyContinue
    if (-not $proc) {
      Write-Ok "$Name already exited (stale PID: $processId cleaned)."
      return
    }

    Stop-Process -Id $processId -Force -ErrorAction Stop
    Write-Ok "$Name stopped (PID: $processId)"
  }
  catch {
    Write-WarnText "$Name stop by pid file failed: $($_.Exception.Message)"
  }
  finally {
    Remove-Item $PidFilePath -Force -ErrorAction SilentlyContinue
  }
}

function Stop-PortProcesses([int]$Port, [string]$Name) {
  $processIds = Get-PortPids -Port $Port
  if (-not $processIds -or $processIds.Count -eq 0) {
    Write-WarnText "$Name not running on port $Port."
    return
  }

  foreach ($processId in $processIds) {
    try {
      Stop-Process -Id $processId -Force -ErrorAction Stop
      Write-Ok "$Name stopped (PID: $processId)"
    }
    catch {
      Write-WarnText "$Name stop failed (PID: $($processId)): $($_.Exception.Message)"
    }
  }
}

function Stop-FrontendViteProcesses([string]$RootPath) {
  $frontendPath = Join-Path $RootPath 'frontend'
  $escapedPath = [Regex]::Escape($frontendPath)

  $targets = @(Get-CimInstance Win32_Process -Filter "Name='node.exe' OR Name='cmd.exe'" -ErrorAction SilentlyContinue |
    Where-Object {
      $_.CommandLine -and
      $_.CommandLine -match 'vite' -and
      $_.CommandLine -match $escapedPath
    })

  if (-not $targets) {
    Write-WarnText 'Frontend vite process not running.'
    return
  }

  foreach ($proc in $targets) {
    try {
      Stop-Process -Id $proc.ProcessId -Force -ErrorAction Stop
      Write-Ok "Frontend process stopped (PID: $($proc.ProcessId))"
    }
    catch {
      Write-WarnText "Failed to stop frontend process PID=$($proc.ProcessId)"
    }
  }
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
      Write-Ok "Agent legacy process stopped (PID: $($proc.ProcessId))"
    }
    catch {
      Write-WarnText "Failed to stop agent legacy process PID=$($proc.ProcessId)"
    }
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
    throw 'docker compose is not available.'
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

Write-Title -Text 'Lesson Plan System - Stop Services'
Import-EnvFile -FilePath $EnvFile
$AgentPort = Resolve-AgentPort

Write-Step -Text '[1/3] Stop frontend service...'
Stop-ProcessByPidFile -PidFilePath $FrontendPidFile -Name 'frontend service'
Stop-FrontendViteProcesses -RootPath $ProjectDir

Write-Step -Text '[2/3] Stop agent service...'
Stop-ProcessByPidFile -PidFilePath $AgentPidFile -Name 'agent service'
Stop-PortProcesses -Port $AgentPort -Name 'agent service'
Stop-PortProcesses -Port 3001 -Name 'legacy agent service (port 3001)'
Stop-AgentLegacyProcesses -RootPath $ProjectDir

Write-Step -Text '[3/3] Stop backend service...'
Stop-ProcessByPidFile -PidFilePath $BackendPidFile -Name 'backend service'
Stop-PortProcesses -Port 8080 -Name 'backend service'

Write-Step -Text 'Verify service status...'
$stillRunning = $false

if ((Get-PortPids -Port 8080).Count -gt 0) {
  Write-WarnText 'Backend still running on port 8080.'
  $stillRunning = $true
}

if ((Get-PortPids -Port $AgentPort).Count -gt 0) {
  Write-WarnText "Agent still running on port $AgentPort."
  $stillRunning = $true
}

if (-not $stillRunning) {
  Write-Ok 'All app services are stopped.'
}

if (Test-Path $LogDir) {
  Get-ChildItem -Path $LogDir -Filter '*.pid' -ErrorAction SilentlyContinue | Remove-Item -Force -ErrorAction SilentlyContinue
}

if ($StopDocker) {
  Write-Step -Text 'Stopping Docker containers...'
  if (-not (Test-CommandAvailable -Name 'docker')) {
    Write-WarnText 'Docker not found. Skip container stop.'
  }
  else {
    Push-Location $ProjectDir
    try {
      Invoke-DockerCompose -Arguments @('down')
      Write-Ok 'Docker containers stopped.'
    }
    catch {
      Write-WarnText "Docker stop failed: $($_.Exception.Message)"
    }
    finally {
      Pop-Location
    }
  }
}

Write-Host ''
Write-Host '========================================' -ForegroundColor Green
Write-Host '  Done' -ForegroundColor Green
Write-Host '========================================' -ForegroundColor Green
