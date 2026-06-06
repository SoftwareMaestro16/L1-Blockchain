param(
  [string]$Matrix = "docs\test-pyramid.md"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
$MatrixPath = if ([System.IO.Path]::IsPathRooted($Matrix)) {
  [System.IO.Path]::GetFullPath($Matrix)
} else {
  [System.IO.Path]::GetFullPath((Join-Path $RepoRoot $Matrix))
}

function Assert-Contains {
  param([string]$Text, [string]$Pattern, [string]$Message)
  if ($Text -notmatch $Pattern) { throw $Message }
}

$text = Get-Content -Raw -LiteralPath $MatrixPath

foreach ($heading in @("Fast PR Suite", "Nightly Or Manual", "Matrix", "Priority Gaps")) {
  Assert-Contains -Text $text -Pattern "## $([regex]::Escape($heading))" -Message "test pyramid missing heading: $heading"
}

foreach ($column in @("Unit", "Integration", "Adversarial", "E2E Smoke", "Determinism", "Benchmark Or Perf")) {
  Assert-Contains -Text $text -Pattern ([regex]::Escape($column)) -Message "test pyramid missing layer column: $column"
}

foreach ($flow in @(
    "App genesis/export/module accounts",
    "Fees ante and params",
    "Contract assets denom lifecycle",
    "DEX pool and swap lifecycle",
    "Query surface CLI/gRPC/REST",
    "PoS and bank native flow",
    "Localnet scripts and release artifacts",
    "Security gates and docs"
  )) {
  Assert-Contains -Text $text -Pattern ([regex]::Escape($flow)) -Message "test pyramid missing flow row: $flow"
}

foreach ($risk in @(
    "invalid signer",
    "unauthorized admin",
    "malformed denom",
    "corrupted pool",
    "wrong denom",
    "tiny rounding output",
    "invalid authority",
    "untriaged Critical/High"
  )) {
  Assert-Contains -Text $text -Pattern ([regex]::Escape($risk)) -Message "test pyramid missing adversarial risk: $risk"
}

foreach ($testRef in @(
    "x/fees/keeper/ante_test.go",
    "x/aetherisvm/standards/aft/keeper/msg_server_test.go",
    "avm-dex-contract/keeper/msg_server_test.go",
    "x/*/keeper/query_server_test.go",
    "app/determinism_test.go",
    "tests/e2e/dex_smoke.ps1",
    "tests/e2e/query_surface_smoke.ps1",
    "scripts/security/prototype-audit.ps1 -Profile Fast",
    "BenchmarkEmptyBlockFinalizeCommit",
    "BenchmarkDexCreatePoolsAndSwap"
  )) {
  Assert-Contains -Text $text -Pattern ([regex]::Escape($testRef)) -Message "test pyramid missing test reference: $testRef"
}

Assert-Contains -Text $text -Pattern "MUST FIX" -Message "test pyramid must mark known gaps"
Assert-Contains -Text $text -Pattern ([regex]::Escape("go test ./...")) -Message "test pyramid must include fast go test command"
Assert-Contains -Text $text -Pattern ([regex]::Escape("go test -run '^$' -bench")) -Message "test pyramid must include benchmark command"

Write-Host "test pyramid doc test passed"
