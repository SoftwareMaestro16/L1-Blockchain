$ErrorActionPreference = "Stop"

$localnetLib = Join-Path $PSScriptRoot "lib"
foreach ($helper in @(
    "paths.ps1",
    "ports.ps1",
    "wait.ps1",
    "cli.ps1",
    "queries.ps1",
    "process.ps1",
    "observability.ps1",
    "profiles.ps1"
  )) {
  . (Join-Path $localnetLib $helper)
}
