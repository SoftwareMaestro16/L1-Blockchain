param(
  [string]$Buf = "buf"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
$GeneratedRoot = Join-Path $RepoRoot ".work\bufgen\github.com\sovereign-l1\l1\x"
$CheckedRoot = Join-Path $RepoRoot "x"

function Get-RelativeChildPath {
  param(
    [string]$Root,
    [string]$Path
  )

  $resolvedRoot = [System.IO.Path]::GetFullPath($Root).TrimEnd("\", "/")
  $resolvedPath = [System.IO.Path]::GetFullPath($Path)
  $prefix = $resolvedRoot + [System.IO.Path]::DirectorySeparatorChar
  if (-not $resolvedPath.StartsWith($prefix, [System.StringComparison]::OrdinalIgnoreCase)) {
    throw "$Path is not under $Root"
  }
  return $resolvedPath.Substring($prefix.Length)
}

Push-Location $RepoRoot
try {
  & $Buf generate
  if ($LASTEXITCODE -ne 0) {
    throw "buf generate failed with exit code $LASTEXITCODE"
  }

  if (-not (Test-Path $GeneratedRoot)) {
    throw "generated output missing: $GeneratedRoot"
  }

  $generatedFiles = Get-ChildItem $GeneratedRoot -Recurse -File |
    Where-Object { $_.Name -like "*.pb.go" -or $_.Name -like "*.pb.gw.go" }

  if (@($generatedFiles).Count -eq 0) {
    throw "no generated protobuf files found under $GeneratedRoot"
  }

  $errors = New-Object System.Collections.Generic.List[string]

  foreach ($file in $generatedFiles) {
    $relative = Get-RelativeChildPath -Root $GeneratedRoot -Path $file.FullName
    $checked = Join-Path $CheckedRoot $relative
    if (-not (Test-Path $checked)) {
      $errors.Add("missing checked-in generated file: x\$relative")
      continue
    }

    $generatedText = (Get-Content -Raw -LiteralPath $file.FullName) -replace "`r`n", "`n"
    $checkedText = (Get-Content -Raw -LiteralPath $checked) -replace "`r`n", "`n"
    if ($generatedText -ne $checkedText) {
      $errors.Add("generated drift: x\$relative")
    }
  }

  $checkedFiles = Get-ChildItem $CheckedRoot -Recurse -File |
    Where-Object { $_.FullName -match "\\types\\" -and ($_.Name -like "*.pb.go" -or $_.Name -like "*.pb.gw.go") }

  foreach ($file in $checkedFiles) {
    $relative = Get-RelativeChildPath -Root $CheckedRoot -Path $file.FullName
    $generated = Join-Path $GeneratedRoot $relative
    if (-not (Test-Path $generated)) {
      $errors.Add("checked-in generated file has no buf output: x\$relative")
    }
  }

  if ($errors.Count -gt 0) {
    $errors | ForEach-Object { Write-Host $_ }
    throw "generated protobuf verification failed"
  }

  Write-Host "generated protobuf verification passed for $(@($generatedFiles).Count) files"
} finally {
  Pop-Location
}
