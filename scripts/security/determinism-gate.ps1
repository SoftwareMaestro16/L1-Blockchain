param(
  [string]$OutputDir = "",
  [switch]$Strict,
  [switch]$Json
)

$ErrorActionPreference = "Stop"

function Get-DeterminismRepoRoot {
  return [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
}

function Resolve-DeterminismPath {
  param([string]$Path, [string]$DefaultRelativePath)
  $repoRoot = Get-DeterminismRepoRoot
  if ([string]::IsNullOrWhiteSpace($Path)) {
    $Path = Join-Path $repoRoot $DefaultRelativePath
  } elseif (-not [System.IO.Path]::IsPathRooted($Path)) {
    $Path = Join-Path $repoRoot $Path
  }
  return [System.IO.Path]::GetFullPath($Path)
}

function Assert-DeterminismWorkspacePath {
  param([string]$Path, [string]$Purpose)
  $repoRoot = (Get-DeterminismRepoRoot).TrimEnd('\', '/')
  $fullPath = [System.IO.Path]::GetFullPath($Path).TrimEnd('\', '/')
  $prefix = $repoRoot + [System.IO.Path]::DirectorySeparatorChar
  if ($fullPath.Equals($repoRoot, [System.StringComparison]::OrdinalIgnoreCase)) {
    throw "Refusing to use repository root as $Purpose`: $fullPath"
  }
  if (-not $fullPath.StartsWith($prefix, [System.StringComparison]::OrdinalIgnoreCase)) {
    throw "Refusing to use path outside repository as $Purpose`: $fullPath"
  }
}

function Get-DeterminismSeverityRank {
  param([string]$Severity)
  switch ($Severity) {
    "Critical" { return 4 }
    "High" { return 3 }
    "Medium" { return 2 }
    "Low" { return 1 }
    default { return 0 }
  }
}

function Get-DeterminismClassification {
  param(
    [string]$PatternName,
    [string]$File,
    [string]$Text
  )

  $normalized = $File.Replace("/", "\")
  $isConsensusScope = $normalized.StartsWith("app\") -or $normalized.StartsWith("x\")

  if ($normalized -match '(^|\\).*_test\.go$' -or $normalized -eq "app\test_helpers.go") {
    return @{
      Severity = "Low"
      Status   = "triaged"
      Impact   = "test/dev helper only; excluded from release determinism risk"
    }
  }

  if ($normalized -eq "cmd\l1d\cmd\speedtest.go") {
    return @{
      Severity = "Low"
      Status   = "triaged"
      Impact   = "CLI benchmark path only; no consensus state or AppHash writes"
    }
  }

  if ($normalized -eq "cmd\l1d\cmd\testnet_genesis.go" -and $PatternName -eq "wall-clock") {
    return @{
      Severity = "Low"
      Status   = "triaged"
      Impact   = "local genesis timestamp; one init run writes identical genesis to all nodes"
    }
  }

  if ($normalized -eq "app\abci.go" -and $PatternName -eq "random") {
    return @{
      Severity = "Medium"
      Status   = "triaged"
      Impact   = "dummy vote extension bytes do not write app state; replace or disable before public validators"
    }
  }

  if ($PatternName -eq "panic") {
    if ($normalized -match 'x\\[^\\]+\\keeper\\(msg_server|ante)\.go$') {
      return @{
        Severity = "High"
        Status   = "untriaged"
        Impact   = "panic in tx/ante path can halt validators"
      }
    }
    if ($isConsensusScope) {
      return @{
        Severity = "Medium"
        Status   = "triaged"
        Impact   = "startup/genesis/export/app wiring panic review; malformed tx/query paths must return errors"
      }
    }
    return @{
      Severity = "Low"
      Status   = "triaged"
      Impact   = "CLI or dev path panic; not a consensus state transition"
    }
  }

  if ($PatternName -eq "platform-int") {
    return @{
      Severity = "Low"
      Status   = "triaged"
      Impact   = "platform int review item; current matches are indices, counters, flags, or SDK callback signatures"
    }
  }

  if ($PatternName -in @("map-iteration", "wall-clock", "random", "float", "goroutine", "select", "external-api") -and $isConsensusScope) {
    return @{
      Severity = "High"
      Status   = "untriaged"
      Impact   = "possible nondeterministic consensus state path; requires fix or explicit downgrade"
    }
  }

  return @{
    Severity = "Low"
    Status   = "triaged"
    Impact   = "outside custom consensus path"
  }
}

$RepoRoot = Get-DeterminismRepoRoot
$OutputDir = Resolve-DeterminismPath -Path $OutputDir -DefaultRelativePath ".work\security\determinism-gate-$(Get-Date -Format 'yyyyMMdd-HHmmss')"
Assert-DeterminismWorkspacePath -Path $OutputDir -Purpose "determinism gate output directory"
New-Item -ItemType Directory -Force -Path $OutputDir | Out-Null

$patterns = @(
  @{ Name = "map-iteration"; Regex = "range\s+.*map\[" },
  @{ Name = "wall-clock"; Regex = "(time|cmttime)\.Now\(" },
  @{ Name = "random"; Regex = "(^|[^\w])(rand|crand)\." },
  @{ Name = "float"; Regex = "float32|float64" },
  @{ Name = "goroutine"; Regex = "\bgo\s+func" },
  @{ Name = "select"; Regex = "select\s*\{" },
  @{ Name = "panic"; Regex = "panic\s*\(" },
  @{ Name = "platform-int"; Regex = "\b(int|uint)\b" },
  @{ Name = "external-api"; Regex = "net/http|http\.|grpc\.Dial|os\.Getenv" }
)

$scanRoots = @("app", "x", "cmd\l1d\cmd")
$findings = @()

Push-Location $RepoRoot
try {
  foreach ($pattern in $patterns) {
    $previousErrorActionPreference = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    try {
      $matches = & rg -n --no-heading $pattern.Regex @scanRoots `
        -g "*.go" `
        -g "!**/*.pb.go" `
        -g "!**/*.pb.gw.go" 2>$null
      $rgExitCode = $LASTEXITCODE
    } finally {
      $ErrorActionPreference = $previousErrorActionPreference
    }
    if ($rgExitCode -ne 0) {
      continue
    }

    foreach ($match in $matches) {
      if ($match -notmatch '^(.+?):(\d+):(.*)$') {
        continue
      }
      $file = $Matches[1]
      $line = [int]$Matches[2]
      $text = $Matches[3].Trim()
      $classification = Get-DeterminismClassification -PatternName $pattern.Name -File $file -Text $text
      $findings += [ordered]@{
        pattern  = $pattern.Name
        severity = $classification.Severity
        status   = $classification.Status
        file     = $file
        line     = $line
        impact   = $classification.Impact
        text     = $text
      }
    }
  }
} finally {
  Pop-Location
}

$findings = @($findings | Sort-Object @{ Expression = { Get-DeterminismSeverityRank $_.severity }; Descending = $true }, file, line, pattern)
$jsonPath = Join-Path $OutputDir "determinism-findings.json"
$summaryPath = Join-Path $OutputDir "summary.md"

$findings | ConvertTo-Json -Depth 6 | Set-Content -LiteralPath $jsonPath

$summary = @()
$summary += "# Deterministic Execution Gate"
$summary += ""
$summary += "- output: ``$OutputDir``"
$summary += "- strict: ``$Strict``"
$summary += "- scanned roots: ``$($scanRoots -join ', ')``"
$summary += ""
$summary += "| Severity | Status | Pattern | Location | Consensus impact |"
$summary += "| --- | --- | --- | --- | --- |"
foreach ($finding in $findings) {
  $location = "$($finding.file):$($finding.line)"
  $summary += "| $($finding.severity) | $($finding.status) | $($finding.pattern) | $location | $($finding.impact) |"
}
if ($findings.Count -eq 0) {
  $summary += "| Low | triaged | none | n/a | no matches |"
}
$summary | Set-Content -LiteralPath $summaryPath

$blocking = @($findings | Where-Object {
    $_.status -eq "untriaged" -and (Get-DeterminismSeverityRank $_.severity) -ge (Get-DeterminismSeverityRank "High")
  })
$strictBlocking = @($findings | Where-Object {
    $Strict -and $_.status -eq "untriaged"
  })

if ($Json) {
  [ordered]@{
    output_dir       = $OutputDir
    findings         = $findings.Count
    blocking         = $blocking.Count
    strict_blocking  = $strictBlocking.Count
    findings_json    = $jsonPath
    summary          = $summaryPath
  } | ConvertTo-Json -Depth 6
} else {
  Write-Host "Determinism findings: $($findings.Count)"
  Write-Host "Blocking findings: $($blocking.Count)"
  Write-Host "Summary: $summaryPath"
}

if ($blocking.Count -gt 0) {
  throw "Deterministic execution gate failed with $($blocking.Count) untriaged High/Critical finding(s); see $summaryPath"
}
if ($strictBlocking.Count -gt 0) {
  throw "Deterministic execution strict gate failed with $($strictBlocking.Count) untriaged finding(s); see $summaryPath"
}
