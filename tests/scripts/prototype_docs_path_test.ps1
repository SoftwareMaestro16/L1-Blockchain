param(
  [string]$Readme = "README.md",
  [string]$Governance = "docs\engineering-governance.md",
  [string]$SecurityTesting = "docs\security-testing.md"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
$ReadmePath = if ([System.IO.Path]::IsPathRooted($Readme)) { $Readme } else { Join-Path $RepoRoot $Readme }
$GovernancePath = if ([System.IO.Path]::IsPathRooted($Governance)) { $Governance } else { Join-Path $RepoRoot $Governance }
$SecurityTestingPath = if ([System.IO.Path]::IsPathRooted($SecurityTesting)) { $SecurityTesting } else { Join-Path $RepoRoot $SecurityTesting }

function Assert-Contains {
  param(
    [string]$Text,
    [string]$Pattern,
    [string]$Message
  )
  if ($Text -notmatch $Pattern) {
    throw $Message
  }
}

function Assert-NotContains {
  param(
    [string]$Text,
    [string]$Pattern,
    [string]$Message
  )
  if ($Text -match $Pattern) {
    throw $Message
  }
}

$readmeText = Get-Content -Raw -LiteralPath $ReadmePath
$governanceText = Get-Content -Raw -LiteralPath $GovernancePath
$securityText = Get-Content -Raw -LiteralPath $SecurityTestingPath

foreach ($link in @(
    'docs/prototype-contract\.md',
    'docs/operator-commands\.md',
    'docs/operator-troubleshooting\.md',
    'docs/transaction-lifecycle-matrix\.md',
    'docs/event-contract\.md',
    'docs/prototype-acceptance-suite\.md',
    'docs/security/prototype-audit-gate\.md',
    'docs/release/prototype-package\.md',
    'docs/release/prototype-limitations\.md',
    'docs/query-surface\.md',
    'docs/observability\.md',
    'docs/engineering-governance\.md',
    'docs/security-testing\.md',
    'docs/security/cosmos-security-checklist\.md',
    'docs/test-pyramid\.md'
  )) {
  Assert-Contains -Text $readmeText -Pattern $link -Message "README missing prototype path link: $link"
}

foreach ($command in @(
    '\.\\scripts\\build-aetrad\.ps1',
    '\.\\scripts\\localnet\\init\.ps1',
    '\.\\scripts\\localnet\\start\.ps1',
    '\.\\tests\\e2e\\prototype_smoke\.ps1',
    '\.\\scripts\\security\\prototype-audit\.ps1 -Profile Fast'
  )) {
  Assert-Contains -Text $readmeText -Pattern $command -Message "README missing runnable prototype command: $command"
}

Assert-Contains -Text $readmeText -Pattern 'not mainnet-ready' -Message "README must clearly warn that prototype is not mainnet-ready"
Assert-Contains -Text $readmeText -Pattern 'Redis or PostgreSQL.*do not require|do not require Redis or PostgreSQL' -Message "README must state external databases are not consensus requirements"
Assert-NotContains -Text $readmeText -Pattern '(?i)redis://|postgresql://|mnemonic:\s*[a-z]|private[_-]?key\s*[:=]' -Message "README must not contain secrets"

Assert-Contains -Text $governanceText -Pattern 'current working branch or owner-approved target branch' -Message "governance must allow owner-approved working-branch push"
Assert-Contains -Text $governanceText -Pattern 'Push directly to `main` only' -Message "governance must not default direct main pushes"
Assert-Contains -Text $governanceText -Pattern 'STEP\.md.*STEP_V2\.md.*STEP_V3\.md' -Message "governance must protect local STEP roadmap files"
Assert-NotContains -Text $governanceText -Pattern 'Push completed commits to `origin/main` unless explicitly told not to' -Message "governance must not force origin/main push"

Assert-Contains -Text $securityText -Pattern 'Prototype Security And Determinism Audit Gate' -Message "security testing must link prototype audit gate"
Assert-Contains -Text $securityText -Pattern 'Cosmos Security Audit Checklist' -Message "security testing must link Cosmos checklist"
Assert-Contains -Text $securityText -Pattern 'Prototype Transaction Lifecycle Matrix' -Message "security testing must link transaction lifecycle matrix"
Assert-Contains -Text $securityText -Pattern 'Prototype Tx Event Contract' -Message "security testing must link event contract"
Assert-Contains -Text $securityText -Pattern 'Prototype Test Pyramid' -Message "security testing must link test pyramid"

Write-Host "prototype docs path test passed"
