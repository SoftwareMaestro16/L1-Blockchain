param(
  [string]$OutputDir = "",
  [string]$Binary = "",
  [int]$ValidatorCount = 0
)

$ErrorActionPreference = "Stop"
$RepoRoot = Resolve-Path (Join-Path $PSScriptRoot "..\..")
if ($OutputDir -eq "") { $OutputDir = Join-Path $RepoRoot ".localnet" }
if ($Binary -eq "") { $Binary = Join-Path $RepoRoot "build\orbitalisd.exe" }

if (!(Test-Path $Binary) -or !(Test-Path $OutputDir)) {
  $initArgs = @{
    OutputDir = $OutputDir
    Binary    = $Binary
  }
  if ($ValidatorCount -gt 0) { $initArgs.ValidatorCount = $ValidatorCount }
  & (Join-Path $PSScriptRoot "init.ps1") @initArgs
}

$nodes = Get-ChildItem -LiteralPath $OutputDir -Directory -Filter "node*" |
  Where-Object { Test-Path (Join-Path $_.FullName "orbitalisd\config\genesis.json") } |
  Sort-Object {
    if ($_.Name -match '^node(\d+)$') { [int]$Matches[1] } else { [int]::MaxValue }
  }

if ($ValidatorCount -gt 0 -and $nodes.Count -ne $ValidatorCount) {
  throw "Expected $ValidatorCount validator nodes under $OutputDir, found $($nodes.Count)"
}
if ($nodes.Count -lt 1) {
  throw "No validator node directories found under $OutputDir"
}

$pidDir = Join-Path $OutputDir "pids"
$logDir = Join-Path $OutputDir "logs"
New-Item -ItemType Directory -Force -Path $pidDir, $logDir | Out-Null

foreach ($node in $nodes) {
  $nodeName = $node.Name
  $nodeHome = Join-Path $node.FullName "orbitalisd"
  $stdout = Join-Path $logDir "$nodeName.out.log"
  $stderr = Join-Path $logDir "$nodeName.err.log"
  $proc = Start-Process -FilePath $Binary `
    -ArgumentList @("start", "--home", $nodeHome) `
    -RedirectStandardOutput $stdout `
    -RedirectStandardError $stderr `
    -WindowStyle Hidden `
    -PassThru
  Set-Content -LiteralPath (Join-Path $pidDir "$nodeName.pid") -Value $proc.Id
  Write-Host "Started $nodeName pid=$($proc.Id)"
}

Write-Host "Logs: $logDir"
