param(
  [int]$MinHeight = 3,
  [int]$TimeoutSeconds = 60
)

$ErrorActionPreference = "Stop"

$RepoRoot = Resolve-Path (Join-Path $PSScriptRoot "..\..")
$Binary = Join-Path $RepoRoot "build\orbitalisd.exe"
$Node = "tcp://127.0.0.1:26657"

Push-Location $RepoRoot
try {
  & .\scripts\localnet\init.ps1
  & .\scripts\localnet\start.ps1

  $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
  $height = 0
  $success = $false

  while ((Get-Date) -lt $deadline) {
    Start-Sleep -Seconds 3
    $json = & $Binary query block --node $Node --output json 2>$null
    if ($LASTEXITCODE -ne 0 -or -not $json) {
      continue
    }

    $jsonText = $json -join "`n"
    $jsonStart = $jsonText.IndexOf("{")
    if ($jsonStart -lt 0) {
      continue
    }

    $block = $jsonText.Substring($jsonStart) | ConvertFrom-Json
    $heightValue = $block.header.height
    if (-not $heightValue -and $block.block) {
      $heightValue = $block.block.header.height
    }

    if (-not $heightValue) {
      continue
    }

    $height = [int]$heightValue
    if ($height -ge $MinHeight) {
      Write-Host "localnet reached height $height"
      $success = $true
      break
    }
  }

  if (-not $success) {
    throw "localnet did not reach height $MinHeight within $TimeoutSeconds seconds; last height $height"
  }
}
finally {
  & .\scripts\localnet\stop.ps1
  Pop-Location
}
