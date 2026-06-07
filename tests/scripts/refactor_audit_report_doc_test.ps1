$ErrorActionPreference = "Stop"

$repo = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)
$doc = Join-Path $repo "docs/security/refactor-audit-report.md"

if (!(Test-Path $doc)) {
  throw "missing refactor audit report: $doc"
}

$content = Get-Content -LiteralPath $doc -Raw

$required = @(
  "Refactor Audit Report",
  "x/aetravm/async",
  "types.go",
  "params.go",
  "address.go",
  "validation.go",
  "executor.go",
  "queue.go",
  "process.go",
  "export.go",
  "clone.go",
  "test_helpers_test.go",
  "range over maps",
  "time.Now",
  "rand",
  "goroutines",
  "select",
  "floating point",
  "Params.Validate",
  "ValidateExportedState",
  "sorted owner keys",
  "not consensus-critical",
  "go test ./...",
  "go vet ./...",
  "buf lint",
  "git diff --check",
  "No PR",
  "Direct push"
)

foreach ($term in $required) {
  if ($content -notlike "*$term*") {
    throw "refactor audit report missing required term: $term"
  }
}

Write-Host "refactor audit report documentation guard passed"
