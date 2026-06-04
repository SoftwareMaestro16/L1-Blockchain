param(
  [ValidateSet("Fast", "Full", "Nightly")]
  [string]$Profile = "Fast",
  [string]$OutputDir = "",
  [switch]$Strict,
  [switch]$SkipTests,
  [switch]$SkipGovulncheck,
  [switch]$SkipGosec,
  [switch]$SkipGitleaksHistory,
  [switch]$SkipAcceptance
)

$ErrorActionPreference = "Stop"

function Get-AuditRepoRoot {
  return [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
}

function Resolve-AuditPath {
  param(
    [string]$Path,
    [string]$DefaultRelativePath
  )

  $repoRoot = Get-AuditRepoRoot
  if ([string]::IsNullOrWhiteSpace($Path)) {
    $Path = Join-Path $repoRoot $DefaultRelativePath
  } elseif (-not [System.IO.Path]::IsPathRooted($Path)) {
    $Path = Join-Path $repoRoot $Path
  }

  return [System.IO.Path]::GetFullPath($Path)
}

function Assert-AuditWorkspacePath {
  param(
    [string]$Path,
    [string]$Purpose
  )

  $repoRoot = (Get-AuditRepoRoot).TrimEnd('\', '/')
  $fullPath = [System.IO.Path]::GetFullPath($Path).TrimEnd('\', '/')
  $prefix = $repoRoot + [System.IO.Path]::DirectorySeparatorChar
  if ($fullPath.Equals($repoRoot, [System.StringComparison]::OrdinalIgnoreCase)) {
    throw "Refusing to use repository root as $Purpose`: $fullPath"
  }
  if (-not $fullPath.StartsWith($prefix, [System.StringComparison]::OrdinalIgnoreCase)) {
    throw "Refusing to use path outside repository as $Purpose`: $fullPath"
  }
}

function Resolve-AuditTool {
  param(
    [string]$Name,
    [string]$LocalPath
  )

  if (Test-Path -LiteralPath $LocalPath) { return $LocalPath }
  $cmd = Get-Command $Name -ErrorAction SilentlyContinue
  if ($cmd) { return $cmd.Source }
  return $LocalPath
}

function Add-AuditResult {
  param(
    [string]$Name,
    [string]$Status,
    [int]$ExitCode,
    [string]$Log,
    [string]$Notes = ""
  )

  $script:Results += [ordered]@{
    name      = $Name
    status    = $Status
    exit_code = $ExitCode
    log       = $Log
    notes     = $Notes
  }
}

function Invoke-AuditNative {
  param(
    [string]$Name,
    [string]$Executable,
    [string[]]$Arguments,
    [switch]$AllowFailure,
    [string]$Notes = ""
  )

  $logPath = Join-Path $OutputDir "$Name.log"
  Write-Host "==> $Name"
  $previousErrorActionPreference = $ErrorActionPreference
  $ErrorActionPreference = "Continue"
  try {
    $output = & $Executable @Arguments 2>&1
    $exitCode = $LASTEXITCODE
  } finally {
    $ErrorActionPreference = $previousErrorActionPreference
  }
  $output | Set-Content -LiteralPath $logPath

  if ($exitCode -eq 0) {
    Add-AuditResult -Name $Name -Status "pass" -ExitCode $exitCode -Log $logPath -Notes $Notes
    return
  }

  if ($AllowFailure) {
    Add-AuditResult -Name $Name -Status "triage_required" -ExitCode $exitCode -Log $logPath -Notes $Notes
    if ($Strict) {
      throw "$Name returned exit code $exitCode in strict mode; see $logPath"
    }
    return
  }

  Add-AuditResult -Name $Name -Status "fail" -ExitCode $exitCode -Log $logPath -Notes $Notes
  throw "$Name returned exit code $exitCode; see $logPath"
}

function Invoke-AuditPowerShell {
  param(
    [string]$Name,
    [scriptblock]$Script,
    [switch]$AllowFailure,
    [string]$Notes = ""
  )

  $logPath = Join-Path $OutputDir "$Name.log"
  Write-Host "==> $Name"
  try {
    $output = & $Script 2>&1
    $output | Set-Content -LiteralPath $logPath
    Add-AuditResult -Name $Name -Status "pass" -ExitCode 0 -Log $logPath -Notes $Notes
  } catch {
    $_ | Out-String | Set-Content -LiteralPath $logPath
    if ($AllowFailure) {
      Add-AuditResult -Name $Name -Status "triage_required" -ExitCode 1 -Log $logPath -Notes $Notes
      if ($Strict) {
        throw "$Name failed in strict mode; see $logPath"
      }
      return
    }
    Add-AuditResult -Name $Name -Status "fail" -ExitCode 1 -Log $logPath -Notes $Notes
    throw "$Name failed; see $logPath"
  }
}

$RepoRoot = Get-AuditRepoRoot
$OutputDir = Resolve-AuditPath -Path $OutputDir -DefaultRelativePath ".work\security\prototype-audit-$(Get-Date -Format 'yyyyMMdd-HHmmss')"
Assert-AuditWorkspacePath -Path $OutputDir -Purpose "security audit output directory"
New-Item -ItemType Directory -Force -Path $OutputDir | Out-Null

$Go = Join-Path $RepoRoot ".work\tools\go1.25.11\go\bin\go.exe"
if (!(Test-Path -LiteralPath $Go)) {
  $Go = "go"
}

$ToolBin = Join-Path $RepoRoot ".work\tools\bin"
$Buf = Resolve-AuditTool -Name "buf" -LocalPath (Join-Path $ToolBin "buf.exe")
$Gitleaks = Resolve-AuditTool -Name "gitleaks" -LocalPath (Join-Path $ToolBin "gitleaks.exe")
$Gosec = Resolve-AuditTool -Name "gosec" -LocalPath (Join-Path $ToolBin "gosec.exe")
$Govulncheck = Resolve-AuditTool -Name "govulncheck" -LocalPath (Join-Path $ToolBin "govulncheck.exe")

$goCache = Join-Path $RepoRoot ".work\gocache"
$goTmp = Join-Path $RepoRoot ".work\gotmp"
New-Item -ItemType Directory -Force -Path $goCache, $goTmp | Out-Null
$env:GOCACHE = $goCache
$env:GOTMPDIR = $goTmp
$env:PATH = (Split-Path $Go) + ";" + $ToolBin + ";" + $env:PATH

$script:Results = @()

Push-Location $RepoRoot
try {
  Invoke-AuditNative -Name "go-vet" -Executable $Go -Arguments @("vet", "-p=1", "./...")

  if (-not $SkipTests) {
    Invoke-AuditNative -Name "go-test" -Executable $Go -Arguments @("test", "-p=1", "./...")
  }

  if (Test-Path -LiteralPath $Buf) {
    Invoke-AuditNative -Name "buf-lint" -Executable $Buf -Arguments @("lint")
    Invoke-AuditPowerShell `
      -Name "proto-generated-verify" `
      -Notes "buf generate output must match checked-in x/*/types generated files" `
      -Script {
        & (Join-Path $RepoRoot "scripts\proto\verify-generated.ps1") -Buf $Buf
        if ($LASTEXITCODE -ne 0) {
          throw "proto generated verification failed with exit code $LASTEXITCODE"
        }
      }
  } else {
    Add-AuditResult -Name "buf-lint" -Status "skipped" -ExitCode 0 -Log "" -Notes "buf.exe not found"
    Add-AuditResult -Name "proto-generated-verify" -Status "skipped" -ExitCode 0 -Log "" -Notes "buf.exe not found"
  }

  if (Test-Path -LiteralPath $Gitleaks) {
    Invoke-AuditNative -Name "gitleaks-staged" -Executable $Gitleaks -Arguments @("protect", "--staged", "--redact", "--no-banner")
    if ($Profile -ne "Fast" -and -not $SkipGitleaksHistory) {
      Invoke-AuditNative -Name "gitleaks-history" -Executable $Gitleaks -Arguments @("detect", "--source", ".", "--redact", "--no-banner", "--log-opts", "--all")
    }
  } else {
    Add-AuditResult -Name "gitleaks" -Status "skipped" -ExitCode 0 -Log "" -Notes "gitleaks.exe not found"
  }

  if (-not $SkipGosec -and (Test-Path -LiteralPath $Gosec)) {
    $packageDirs = & $Go list -f "{{.Dir}}" ./...
    $packageDirs | Set-Content -LiteralPath (Join-Path $OutputDir "go-package-dirs.txt")
    $gosecJson = Join-Path $OutputDir "gosec.json"
    Invoke-AuditPowerShell `
      -Name "gosec" `
      -Notes "generated protobuf excluded; see gosec.json for source findings" `
      -Script {
        $previousErrorActionPreference = $ErrorActionPreference
        $ErrorActionPreference = "Continue"
        try {
          $gosecOutput = & $Gosec -exclude-generated -fmt=json -out $gosecJson -no-fail @packageDirs 2>&1
          $gosecExitCode = $LASTEXITCODE
        } finally {
          $ErrorActionPreference = $previousErrorActionPreference
        }
        $gosecOutput
        if ($gosecExitCode -ne 0) {
          throw "gosec command failed with exit code $gosecExitCode"
        }
        if (!(Test-Path -LiteralPath $gosecJson)) {
          throw "gosec did not write report $gosecJson"
        }
        if (Test-Path -LiteralPath $gosecJson) {
          $doc = Get-Content -Raw -LiteralPath $gosecJson | ConvertFrom-Json
          $count = @($doc.Issues).Count
          "gosec source issues: $count"
          foreach ($issue in @($doc.Issues)) {
            "$($issue.severity) $($issue.rule_id) $($issue.file):$($issue.line) $($issue.details)"
          }
        }
      }
  }

  if ($Profile -ne "Fast" -and -not $SkipGovulncheck -and (Test-Path -LiteralPath $Govulncheck)) {
    Invoke-AuditNative `
      -Name "govulncheck" `
      -Executable $Govulncheck `
      -Arguments @("-scan=package", "./...") `
      -AllowFailure `
      -Notes "dependency advisories require triage in docs/security/prototype-audit-gate.md"
  }

  if ($Profile -ne "Fast") {
    Invoke-AuditNative `
      -Name "go-mod-verify" `
      -Executable $Go `
      -Arguments @("mod", "verify") `
      -AllowFailure `
      -Notes "module cache integrity; local cache mutations require cleanup or documented triage"
  }

  Invoke-AuditPowerShell -Name "deterministic-execution-gate" -Script {
    & .\scripts\security\determinism-gate.ps1 -OutputDir (Join-Path $OutputDir "determinism-gate")
  }

  if ($Profile -eq "Nightly" -and -not $SkipAcceptance) {
    Invoke-AuditPowerShell `
      -Name "prototype-acceptance-full-5" `
      -Notes "nightly/manual scale profile" `
      -Script {
        & .\tests\e2e\prototype_acceptance.ps1 -Profile Full -OutputDir .localnet-5 -ValidatorCount 5 -SkipBuild
        if ($LASTEXITCODE -ne 0) {
          throw "prototype acceptance failed"
        }
      }
  }
} finally {
  Pop-Location
}

$summaryPath = Join-Path $OutputDir "summary.md"
$jsonPath = Join-Path $OutputDir "results.json"

$Results | ConvertTo-Json -Depth 5 | Set-Content -LiteralPath $jsonPath

$summary = @()
$summary += "# Prototype Audit Gate Summary"
$summary += ""
$summary += "- profile: ``$Profile``"
$summary += "- strict: ``$Strict``"
$summary += "- output: ``$OutputDir``"
$summary += ""
$summary += "| Check | Status | Exit | Notes |"
$summary += "| --- | --- | ---: | --- |"
foreach ($result in $Results) {
  $summary += "| $($result.name) | $($result.status) | $($result.exit_code) | $($result.notes) |"
}
$summary | Set-Content -LiteralPath $summaryPath

Write-Host "Audit summary written to $summaryPath"

$failed = @($Results | Where-Object { $_.status -eq "fail" })
$triage = @($Results | Where-Object { $_.status -eq "triage_required" })
if ($failed.Count -gt 0) {
  throw "Prototype audit gate failed: $($failed.name -join ', ')"
}
if ($Strict -and $triage.Count -gt 0) {
  throw "Prototype audit gate has unaccepted triage-required findings: $($triage.name -join ', ')"
}
