function ConvertTo-LocalnetRedactedText {
  param([AllowNull()][string]$Text)

  if ($null -eq $Text) { return "" }
  $redacted = $Text
  $redacted = [regex]::Replace($redacted, '(?im)(mnemonic|private[_-]?key|secret|seed|password|token)\s*[:=]\s*"?[^";,\r\n]+', '$1=[REDACTED]')
  $redacted = [regex]::Replace($redacted, '(?im)"(mnemonic|private[_-]?key|secret|seed|password|token)"\s*:\s*"[^"]*"', '"$1":"[REDACTED]"')
  $redacted = [regex]::Replace($redacted, '(?im)(AWS|GITHUB|OPENAI|DATABASE|DB|API)[A-Z0-9_]*(KEY|TOKEN|SECRET|PASSWORD)\s*=\s*[^\s]+', '$1_[REDACTED]=[REDACTED]')
  return $redacted
}

function Get-LocalnetNodeTelemetry {
  param([string]$NodeHome)

  $appToml = Join-Path $NodeHome "config\app.toml"
  $telemetry = [ordered]@{
    enabled = $null
    sink    = $null
  }
  if (-not (Test-Path -LiteralPath $appToml)) {
    return $telemetry
  }

  $content = Get-Content -Raw -LiteralPath $appToml
  if ($content -match '(?ms)\[telemetry\].*?^enabled\s*=\s*(true|false)') {
    $telemetry.enabled = [System.Convert]::ToBoolean($Matches[1])
  }
  if ($content -match '(?ms)\[telemetry\].*?^sink\s*=\s*"([^"]*)"') {
    $telemetry.sink = $Matches[1]
  }
  return $telemetry
}

function Get-LocalnetProcessSnapshot {
  param([string]$OutputDir)

  $resolved = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet"
  $pidDir = Join-Path $resolved "pids"
  $processes = @()

  if (Test-Path -LiteralPath $pidDir) {
    foreach ($pidFile in @(Get-ChildItem -LiteralPath $pidDir -Filter *.pid -ErrorAction SilentlyContinue | Sort-Object Name)) {
      $pidValue = [int](Get-Content -Raw -LiteralPath $pidFile.FullName)
      $proc = Get-Process -Id $pidValue -ErrorAction SilentlyContinue
      $processes += [ordered]@{
        node    = [System.IO.Path]::GetFileNameWithoutExtension($pidFile.Name)
        pid     = $pidValue
        running = [bool]$proc
        name    = if ($proc) { $proc.ProcessName } else { $null }
      }
    }
  }

  if ($processes.Count -eq 0) {
    $escaped = [regex]::Escape($resolved)
    foreach ($proc in @(Get-CimInstance Win32_Process -ErrorAction SilentlyContinue |
        Where-Object { $_.Name -like "aetrad*" -and $_.CommandLine -match $escaped } |
        Sort-Object ProcessId)) {
      $processes += [ordered]@{
        node    = "unknown"
        pid     = [int]$proc.ProcessId
        running = $true
        name    = $proc.Name
      }
    }
  }

  return @($processes)
}

function Get-LocalnetRecentLogs {
  param(
    [string]$OutputDir,
    [int]$TailLines = 40
  )

  if ($TailLines -lt 1) { $TailLines = 1 }
  if ($TailLines -gt 200) { $TailLines = 200 }

  $resolved = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet"
  $logDir = Join-Path $resolved "logs"
  $logs = @()
  if (-not (Test-Path -LiteralPath $logDir)) {
    return @($logs)
  }

  foreach ($file in @(Get-ChildItem -LiteralPath $logDir -Filter "*.log" -File -ErrorAction SilentlyContinue | Sort-Object Name)) {
    $tail = @(Get-Content -LiteralPath $file.FullName -Tail $TailLines -ErrorAction SilentlyContinue)
    $text = ConvertTo-LocalnetRedactedText -Text ($tail -join "`n")
    $logs += [ordered]@{
      file       = $file.Name
      line_count = $tail.Count
      recent     = $text
    }
  }
  return @($logs)
}

function Copy-LocalnetRedactedFile {
  param(
    [string]$Source,
    [string]$Destination
  )

  $text = Get-Content -Raw -LiteralPath $Source
  ConvertTo-LocalnetRedactedText -Text $text | Set-Content -LiteralPath $Destination
}
