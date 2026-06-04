param(
  [string]$Checklist = "docs\security\cosmos-security-checklist.md"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
$ChecklistPath = if ([System.IO.Path]::IsPathRooted($Checklist)) {
  [System.IO.Path]::GetFullPath($Checklist)
} else {
  [System.IO.Path]::GetFullPath((Join-Path $RepoRoot $Checklist))
}

function Assert-Contains {
  param([string]$Text, [string]$Pattern, [string]$Message)
  if ($Text -notmatch $Pattern) { throw $Message }
}

$text = Get-Content -Raw -LiteralPath $ChecklistPath

foreach ($heading in @(
    "Release Rule",
    "Review Record",
    "App",
    "Tokenfactory",
    "DEX",
    "Fees",
    "Localnet Scripts",
    "Proto And Query",
    "Release Artifacts",
    "Risk To Test Map"
  )) {
  Assert-Contains -Text $text -Pattern "## $([regex]::Escape($heading))" -Message "checklist missing heading: $heading"
}

foreach ($term in @(
    "cosmos-vulnerability-scanner",
    "Missing denom validation",
    "insufficient authorization",
    "missing balance check",
    "ABCI panic",
    "nondeterminism",
    "rounding",
    "unbounded loops",
    "Critical",
    "High",
    "owner",
    "issue",
    "decision"
  )) {
  Assert-Contains -Text $text -Pattern ([regex]::Escape($term)) -Message "checklist missing required term: $term"
}

foreach ($module in @("tokenfactory", "dex", "fees", "norb", "MsgUpdateParams", "buf lint")) {
  Assert-Contains -Text $text -Pattern ([regex]::Escape($module)) -Message "checklist missing module/API coverage: $module"
}

foreach ($testRef in @(
    "app/determinism_test.go",
    "scripts/security/determinism-gate.ps1",
    "x/tokenfactory/keeper/msg_server_test.go",
    "x/dex/keeper/msg_server_test.go",
    "x/fees/keeper/ante_test.go",
    "tests/e2e/query_surface_smoke.ps1",
    "tests/scripts/prototype_release_package_test.ps1"
  )) {
  Assert-Contains -Text $text -Pattern ([regex]::Escape($testRef)) -Message "checklist missing test reference: $testRef"
}

Assert-Contains -Text $text -Pattern "Future modules must add rows" -Message "checklist must be reusable for future modules"
Assert-Contains -Text $text -Pattern "Prototype release is blocked when any ``?Critical``? or ``?High``? finding is untriaged" -Message "checklist must block untriaged Critical/High"

Write-Host "cosmos security checklist doc test passed"
