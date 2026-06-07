function Stop-LocalnetProcesses {
  param(
    [string]$OutputDir,
    [string]$PidDir
  )

  if (Test-Path -LiteralPath $PidDir) {
    Get-ChildItem -LiteralPath $PidDir -Filter *.pid | ForEach-Object {
      $pidValue = [int](Get-Content -Raw -LiteralPath $_.FullName)
      $proc = Get-Process -Id $pidValue -ErrorAction SilentlyContinue
      if ($proc) {
        Stop-Process -Id $pidValue -Force -ErrorAction SilentlyContinue
        Wait-Process -Id $pidValue -Timeout 10 -ErrorAction SilentlyContinue
        Write-Host "Stopped pid=$pidValue"
      }
      Remove-Item -LiteralPath $_.FullName -Force
    }
  }

  $resolved = Resolve-LocalnetPath -Path $OutputDir -DefaultRelativePath ".localnet"
  $escaped = [regex]::Escape($resolved)
  Get-CimInstance Win32_Process -ErrorAction SilentlyContinue |
    Where-Object {
      $_.Name -like "aetrad*" -and
      $_.CommandLine -match $escaped
    } |
    ForEach-Object {
      $pidValue = [int]$_.ProcessId
      $proc = Get-Process -Id $pidValue -ErrorAction SilentlyContinue
      if ($proc) {
        Stop-Process -Id $pidValue -Force -ErrorAction SilentlyContinue
        Wait-Process -Id $pidValue -Timeout 10 -ErrorAction SilentlyContinue
        Write-Host "Stopped orphan localnet pid=$pidValue"
      }
    }
}
