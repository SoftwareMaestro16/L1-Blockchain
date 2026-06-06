param(
  [string]$Limitations = "docs\release\prototype-limitations.md",
  [string]$Readme = "README.md",
  [string]$ReleasePackage = "docs\release\prototype-package.md"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
$LimitationsPath = if ([System.IO.Path]::IsPathRooted($Limitations)) { $Limitations } else { Join-Path $RepoRoot $Limitations }
$ReadmePath = if ([System.IO.Path]::IsPathRooted($Readme)) { $Readme } else { Join-Path $RepoRoot $Readme }
$ReleasePackagePath = if ([System.IO.Path]::IsPathRooted($ReleasePackage)) { $ReleasePackage } else { Join-Path $RepoRoot $ReleasePackage }

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

function Assert-True {
  param([bool]$Condition, [string]$Message)
  if (-not $Condition) { throw $Message }
}

Assert-True (Test-Path -LiteralPath $LimitationsPath) "missing prototype limitations doc"

$limitationsText = Get-Content -Raw -LiteralPath $LimitationsPath
$readmeText = Get-Content -Raw -LiteralPath $ReadmePath
$releasePackageText = Get-Content -Raw -LiteralPath $ReleasePackagePath

foreach ($section in @(
    'Prototype Non-Goals And Limitations',
    'Version scope',
    'Non-Goals',
    'Accepted Prototype Limitations',
    'Blockers',
    'prototype-acceptance-report\.md'
  )) {
  Assert-Contains -Text $limitationsText -Pattern $section -Message "limitations doc missing section or reference: $section"
}

foreach ($nonGoal in @(
    'Mainnet launch',
    'IBC',
    'external bridge',
    'Production governance economics',
    'Exchange-grade DEX',
    'Public faucet',
    'Full external audit',
    'Explorer/API SLA'
  )) {
  Assert-Contains -Text $limitationsText -Pattern $nonGoal -Message "limitations doc missing non-goal: $nonGoal"
}

foreach ($blocker in @(
    'Critical/High',
    'not an accepted limitation',
    'wrong fee denom',
    'unauthorized contract-assets',
    'DEX invariant',
    'nondeterminism',
    'unbounded tx/list',
    'secrets'
  )) {
  Assert-Contains -Text $limitationsText -Pattern $blocker -Message "limitations doc missing blocker rule: $blocker"
}

Assert-Contains -Text $readmeText -Pattern 'docs/release/prototype-limitations\.md' -Message "README must link prototype limitations"
Assert-Contains -Text $releasePackageText -Pattern 'prototype-limitations\.md' -Message "release package doc must link prototype limitations"
Assert-Contains -Text $releasePackageText -Pattern 'Critical/High.*blockers|blockers.*Critical/High' -Message "release package must separate Critical/High blockers from limitations"

Write-Host "prototype limitations doc test passed"
