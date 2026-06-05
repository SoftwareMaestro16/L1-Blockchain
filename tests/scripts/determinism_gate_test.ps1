param(
  [string]$OutputDir = ".work\determinism-gate-test"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
$OutputDir = if ([System.IO.Path]::IsPathRooted($OutputDir)) {
  [System.IO.Path]::GetFullPath($OutputDir)
} else {
  [System.IO.Path]::GetFullPath((Join-Path $RepoRoot $OutputDir))
}

function Assert-TestPathInsideRepo {
  param([string]$Path)
  $repo = $RepoRoot.TrimEnd('\', '/')
  $full = [System.IO.Path]::GetFullPath($Path).TrimEnd('\', '/')
  if ($full -eq $repo -or -not $full.StartsWith($repo + [System.IO.Path]::DirectorySeparatorChar, [System.StringComparison]::OrdinalIgnoreCase)) {
    throw "Refusing test path outside repo: $full"
  }
}

function Assert-True {
  param([bool]$Condition, [string]$Message)
  if (-not $Condition) { throw $Message }
}

Assert-TestPathInsideRepo -Path $OutputDir
if (Test-Path -LiteralPath $OutputDir) {
  Remove-Item -LiteralPath $OutputDir -Recurse -Force
}
New-Item -ItemType Directory -Force -Path $OutputDir | Out-Null

Push-Location $RepoRoot
try {
  $result = & .\scripts\security\determinism-gate.ps1 -OutputDir $OutputDir -Json | ConvertFrom-Json
} finally {
  Pop-Location
}

Assert-True (Test-Path -LiteralPath $result.summary) "determinism summary was not written"
Assert-True (Test-Path -LiteralPath $result.findings_json) "determinism findings JSON was not written"
Assert-True ([int]$result.blocking -eq 0) "determinism gate has blocking findings"

$findings = @(Get-Content -Raw -LiteralPath $result.findings_json | ConvertFrom-Json)
$blocking = @($findings | Where-Object { $_.status -eq "untriaged" -and $_.severity -in @("High", "Critical") })
Assert-True ($blocking.Count -eq 0) "untriaged High/Critical determinism findings found"

$summary = Get-Content -Raw -LiteralPath $result.summary
Assert-True ($summary -match "# Deterministic Execution Gate") "determinism summary has unexpected format"

Write-Host "determinism gate test passed"
