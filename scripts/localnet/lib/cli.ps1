function Invoke-ExternalChecked {
  param(
    [Parameter(Mandatory = $true)][string]$FilePath,
    [string[]]$Arguments = @(),
    [string]$FailureMessage = "external command failed"
  )

  $previousErrorActionPreference = $ErrorActionPreference
  $ErrorActionPreference = "Continue"
  try {
    $output = & $FilePath @Arguments 2>&1
    $exitCode = $LASTEXITCODE
  } finally {
    $ErrorActionPreference = $previousErrorActionPreference
  }
  if ($exitCode -ne 0) {
    throw "$FailureMessage`: $FilePath $($Arguments -join ' ')`n$($output -join "`n")"
  }
  return $output
}

function Invoke-LocalnetCliJson {
  param(
    [string]$Binary,
    [string[]]$Arguments
  )

  $previousErrorActionPreference = $ErrorActionPreference
  $ErrorActionPreference = "Continue"
  try {
    $output = & $Binary @Arguments 2>&1
    $exitCode = $LASTEXITCODE
  } finally {
    $ErrorActionPreference = $previousErrorActionPreference
  }

  if ($exitCode -ne 0) {
    throw "aetrad command failed: $Binary $($Arguments -join ' ')`n$($output -join "`n")"
  }

  $text = $output -join "`n"
  $jsonStart = $text.IndexOf("{")
  if ($jsonStart -lt 0) {
    $jsonStart = $text.IndexOf("[")
  }
  if ($jsonStart -lt 0) {
    throw "aetrad command did not return JSON: $Binary $($Arguments -join ' ')`n$text"
  }
  return $text.Substring($jsonStart) | ConvertFrom-Json
}

function Wait-LocalnetHeightIncreasing {
  param(
    [int]$RPCPort = 26657,
    [int]$TimeoutSeconds = 60
  )

  $startHeight = Get-LocalnetHeight -RPCPort $RPCPort
  return Wait-LocalnetCondition -TimeoutSeconds $TimeoutSeconds -Description "height greater than $startHeight on RPC port $RPCPort" -Condition {
    $height = Get-LocalnetHeight -RPCPort $RPCPort
    if ($height -gt $startHeight) {
      return @{
        StartHeight   = $startHeight
        CurrentHeight = $height
      }
    }
    return $null
  }
}

function Invoke-LocalnetCliJsonAllowFailure {
  param(
    [string]$Binary,
    [string[]]$Arguments
  )

  $previousErrorActionPreference = $ErrorActionPreference
  $ErrorActionPreference = "Continue"
  try {
    $output = & $Binary @Arguments 2>&1
  } finally {
    $ErrorActionPreference = $previousErrorActionPreference
  }
  $text = $output -join "`n"
  $jsonStart = $text.IndexOf("{")
  if ($jsonStart -lt 0) {
    $jsonStart = $text.IndexOf("[")
  }
  if ($jsonStart -lt 0) {
    throw "aetrad command did not return JSON: $Binary $($Arguments -join ' ')`n$text"
  }
  return $text.Substring($jsonStart) | ConvertFrom-Json
}

function Assert-LocalnetCliFailure {
  param(
    [string]$Binary,
    [string[]]$Arguments,
    [string]$ExpectedLog = ""
  )

  $previousErrorActionPreference = $ErrorActionPreference
  $ErrorActionPreference = "Continue"
  try {
    $output = & $Binary @Arguments 2>&1
    $exitCode = $LASTEXITCODE
  } finally {
    $ErrorActionPreference = $previousErrorActionPreference
  }

  if ($exitCode -eq 0) {
    throw "aetrad command succeeded but failure was expected: $Binary $($Arguments -join ' ')"
  }

  $text = $output -join "`n"
  if ($ExpectedLog -and ($text -notmatch [regex]::Escape($ExpectedLog))) {
    throw "aetrad command failed, but output did not contain '$ExpectedLog': $text"
  }
  return $text
}

function Get-LocalnetObjectProperty {
  param(
    [object]$InputObject,
    [string]$Name
  )

  if ($null -eq $InputObject) {
    return $null
  }
  $property = $InputObject.PSObject.Properties[$Name]
  if ($null -eq $property) {
    return $null
  }
  return $property.Value
}

function Get-LocalnetTxHash {
  param([object]$Tx)

  $txHash = Get-LocalnetObjectProperty -InputObject $Tx -Name "txhash"
  $txResponse = Get-LocalnetObjectProperty -InputObject $Tx -Name "tx_response"
  if (-not $txHash -and $txResponse) {
    $txHash = Get-LocalnetObjectProperty -InputObject $txResponse -Name "txhash"
  }
  return $txHash
}

function Get-LocalnetTxCode {
  param([object]$Tx)

  $txResponse = Get-LocalnetObjectProperty -InputObject $Tx -Name "tx_response"
  $responseCode = Get-LocalnetObjectProperty -InputObject $txResponse -Name "code"
  if ($null -ne $responseCode) {
    return [int]$responseCode
  }
  $code = Get-LocalnetObjectProperty -InputObject $Tx -Name "code"
  if ($null -ne $code) {
    return [int]$code
  }
  return 0
}

function Get-LocalnetTxLog {
  param([object]$Tx)

  $txResponse = Get-LocalnetObjectProperty -InputObject $Tx -Name "tx_response"
  $responseRawLog = Get-LocalnetObjectProperty -InputObject $txResponse -Name "raw_log"
  if ($responseRawLog) {
    return [string]$responseRawLog
  }
  $rawLog = Get-LocalnetObjectProperty -InputObject $Tx -Name "raw_log"
  if ($rawLog) {
    return [string]$rawLog
  }
  $log = Get-LocalnetObjectProperty -InputObject $Tx -Name "log"
  if ($log) {
    return [string]$log
  }
  return ""
}

function Assert-LocalnetTxFailure {
  param(
    [object]$Tx,
    [string]$ExpectedLog = ""
  )

  $code = Get-LocalnetTxCode -Tx $Tx
  if ($code -eq 0) {
    throw "transaction succeeded but failure was expected"
  }
  $log = Get-LocalnetTxLog -Tx $Tx
  if ($ExpectedLog -and ($log -notmatch [regex]::Escape($ExpectedLog))) {
    throw "transaction failed with code $code, but log did not contain '$ExpectedLog': $log"
  }
  return $Tx
}

function Assert-LocalnetTxEvent {
  param(
    [object]$Tx,
    [string]$Type,
    [hashtable]$Attributes = @{}
  )

  $events = @()
  $txResponse = Get-LocalnetObjectProperty -InputObject $Tx -Name "tx_response"
  $responseEvents = Get-LocalnetObjectProperty -InputObject $txResponse -Name "events"
  $topLevelEvents = Get-LocalnetObjectProperty -InputObject $Tx -Name "events"
  if ($responseEvents) {
    $events = @($responseEvents)
  } elseif ($topLevelEvents) {
    $events = @($topLevelEvents)
  }

  foreach ($event in $events) {
    if ([string]$event.type -ne $Type) {
      continue
    }

    foreach ($key in $Attributes.Keys) {
      $matches = @($event.attributes | Where-Object { [string]$_.key -eq [string]$key })
      if ($matches.Count -lt 1) {
        throw "tx event $Type missing attribute $key"
      }
      $actual = [string]$matches[0].value
      $expected = [string]$Attributes[$key]
      if ($actual -ne $expected) {
        throw "tx event $Type attribute $key mismatch: expected $expected, got $actual"
      }
    }
    return $event
  }

  throw "tx missing event $Type"
}

function Wait-LocalnetTx {
  param(
    [string]$Binary,
    [string]$TxHash,
    [int]$RPCPort = 26657,
    [int]$TimeoutSeconds = 60,
    [switch]$ExpectFailure,
    [string]$ExpectedLog = ""
  )

  if (-not $TxHash) {
    throw "transaction command did not return txhash"
  }

  $node = "tcp://127.0.0.1:$RPCPort"
  $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
  $lastError = $null
  while ((Get-Date) -lt $deadline) {
    try {
      $result = Invoke-LocalnetCliJson -Binary $Binary -Arguments @("query", "tx", $TxHash, "--node", $node, "--output", "json")
    } catch {
      $lastError = $_.Exception.Message
      Start-Sleep -Milliseconds 500
      continue
    }

    if ($ExpectFailure) {
      return Assert-LocalnetTxFailure -Tx $result -ExpectedLog $ExpectedLog
    }

    $code = Get-LocalnetTxCode -Tx $result
    if ($code -ne 0) {
      throw "tx $TxHash failed with code $code`: $(Get-LocalnetTxLog -Tx $result)"
    }
    return $result
  }

  if ($lastError) {
    throw "Timed out waiting for tx $TxHash; last error: $lastError"
  }
  throw "Timed out waiting for tx $TxHash"
}

function Send-LocalnetTx {
  param(
    [string]$Binary,
    [string[]]$Arguments,
    [int]$RPCPort = 26657,
    [int]$TimeoutSeconds = 60,
    [switch]$ExpectFailure,
    [string]$ExpectedLog = ""
  )

  $tx = if ($ExpectFailure) {
    Invoke-LocalnetCliJsonAllowFailure -Binary $Binary -Arguments $Arguments
  } else {
    Invoke-LocalnetCliJson -Binary $Binary -Arguments $Arguments
  }

  if ($ExpectFailure -and (Get-LocalnetTxCode -Tx $tx) -ne 0) {
    return Assert-LocalnetTxFailure -Tx $tx -ExpectedLog $ExpectedLog
  }
  $broadcastCode = Get-LocalnetTxCode -Tx $tx
  if ($broadcastCode -ne 0) {
    throw "broadcast failed with code $broadcastCode`: $(Get-LocalnetTxLog -Tx $tx)"
  }

  return Wait-LocalnetTx `
    -Binary $Binary `
    -TxHash (Get-LocalnetTxHash -Tx $tx) `
    -RPCPort $RPCPort `
    -TimeoutSeconds $TimeoutSeconds `
    -ExpectFailure:$ExpectFailure `
    -ExpectedLog $ExpectedLog
}

function Assert-LocalnetSignedTxReplayFailure {
  param(
    [string]$Binary,
    [string[]]$GenerateArguments,
    [string]$FromKey,
    [string]$FromHome,
    [string]$ChainId,
    [int]$RPCPort = 26657,
    [string]$WorkDir,
    [int]$TimeoutSeconds = 60
  )

  New-Item -ItemType Directory -Force -Path $WorkDir | Out-Null
  $unsignedPath = Join-Path $WorkDir "replay-unsigned.json"
  $signedPath = Join-Path $WorkDir "replay-signed.json"
  $node = "tcp://127.0.0.1:$RPCPort"

  $unsigned = Invoke-LocalnetCliJson -Binary $Binary -Arguments ($GenerateArguments + @(
      "--generate-only",
      "--output", "json"
    ))
  $utf8NoBom = New-Object System.Text.UTF8Encoding $false
  [System.IO.File]::WriteAllText($unsignedPath, ($unsigned | ConvertTo-Json -Depth 100), $utf8NoBom)

  $signArgs = @(
    "tx", "sign", $unsignedPath,
    "--from", $FromKey,
    "--home", $FromHome,
    "--chain-id", $ChainId,
    "--keyring-backend", "test",
    "--node", $node,
    "--output", "json",
    "--output-document", $signedPath
  )
  $previousErrorActionPreference = $ErrorActionPreference
  $ErrorActionPreference = "Continue"
  try {
    $signOutput = & $Binary @signArgs 2>&1
    $signExitCode = $LASTEXITCODE
  } finally {
    $ErrorActionPreference = $previousErrorActionPreference
  }
  if ($signExitCode -ne 0) {
    throw "aetrad tx sign failed: $Binary $($signArgs -join ' ')`n$($signOutput -join "`n")"
  }
  if (-not (Test-Path -LiteralPath $signedPath)) {
    throw "aetrad tx sign did not create signed tx file: $signedPath"
  }

  $broadcastArgs = @(
    "tx", "broadcast", $signedPath,
    "--node", $node,
    "--broadcast-mode", "sync",
    "--output", "json"
  )
  Send-LocalnetTx -Binary $Binary -Arguments $broadcastArgs -RPCPort $RPCPort -TimeoutSeconds $TimeoutSeconds | Out-Null
  Send-LocalnetTx -Binary $Binary -Arguments $broadcastArgs -RPCPort $RPCPort -TimeoutSeconds $TimeoutSeconds -ExpectFailure | Out-Null
}
