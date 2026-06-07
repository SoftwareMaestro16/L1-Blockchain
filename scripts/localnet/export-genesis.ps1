param(
  [string]$OutputDir = "",
  [string]$Binary = "",
  [int]$NodeIndex = 0,
  [string]$ExportDir = "",
  [string]$ChainId = "aetra-local-1"
)

$ErrorActionPreference = "Stop"
. (Join-Path $PSScriptRoot "common.ps1")

$OutputDir = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet"
$Binary = Resolve-LocalnetPath -Path $Binary -DefaultRelativePath "build\aetrad.exe"
$ExportDir = Resolve-LocalnetPath -Path $ExportDir -DefaultRelativePath ".work\genesis"
Assert-LocalnetWorkspacePath -Path $OutputDir -Purpose "localnet output directory"
Assert-LocalnetWorkspacePath -Path $ExportDir -Purpose "genesis export directory"

if (!(Test-Path -LiteralPath $Binary)) {
  throw "Binary not found at $Binary"
}

$pidDir = Join-Path $OutputDir "pids"
if (Test-Path -LiteralPath $pidDir) {
  Get-ChildItem -LiteralPath $pidDir -Filter *.pid -ErrorAction SilentlyContinue | ForEach-Object {
    $pidValue = [int](Get-Content -Raw -LiteralPath $_.FullName)
    if (Get-Process -Id $pidValue -ErrorAction SilentlyContinue) {
      throw "Stop localnet before export; running pid from $($_.FullName): $pidValue"
    }
  }
}

$nodes = Get-LocalnetNodes -OutputDir $OutputDir
if ($nodes.Count -lt 1) {
  throw "No localnet node directories found under $OutputDir"
}
if ($NodeIndex -lt 0 -or $NodeIndex -ge $nodes.Count) {
  throw "NodeIndex $NodeIndex is outside available node range 0..$($nodes.Count - 1)"
}

New-Item -ItemType Directory -Force -Path $ExportDir | Out-Null
$node = $nodes[$NodeIndex]
$nodeHome = Join-Path $node.FullName "aetrad"
$exportPath = Join-Path $ExportDir "$($node.Name)-export.json"
$stderrPath = Join-Path $ExportDir "$($node.Name)-export.err.log"

$exportOutput = & $Binary export --home $nodeHome 2> $stderrPath
if ($LASTEXITCODE -ne 0) {
  $err = if (Test-Path -LiteralPath $stderrPath) { Get-Content -Raw -LiteralPath $stderrPath } else { "" }
  throw "genesis export failed for $($node.Name): $err"
}
[System.IO.File]::WriteAllText($exportPath, ($exportOutput -join "`n"), (New-Object System.Text.UTF8Encoding $false))

& $Binary genesis validate-genesis $exportPath --home $nodeHome
if ($LASTEXITCODE -ne 0) {
  throw "exported genesis validation failed for $exportPath"
}

$raw = Get-Content -Raw -LiteralPath $exportPath
if ($raw -match '(?i)\b(mnemonic|private[_-]?key|priv_validator|secret|seed|wallet)\b') {
  throw "exported genesis contains secret-like material"
}

$doc = $raw | ConvertFrom-Json
if ($doc.chain_id -ne $ChainId) {
  throw "exported genesis chain-id mismatch: expected $ChainId, got $($doc.chain_id)"
}

$hash = (Get-FileHash -LiteralPath $exportPath -Algorithm SHA256).Hash
Write-Host "Exported and validated $($node.Name) genesis: $exportPath"
Write-Host "export sha256: $hash"
