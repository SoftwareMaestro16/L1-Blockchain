param(
  [string]$OutputDir = ".work\aexs",
  [string]$TaskFile = "TO_AUDIT.md",
  [string]$PipelineDoc = "docs\security\aetheris-fuzzing-invariant-pipeline.md",
  [switch]$Json,
  [switch]$EnforceSafe
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version 2.0

function Get-AexsRepoRoot {
  return [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))
}

function Resolve-AexsPath {
  param([string]$Path)
  $repoRoot = Get-AexsRepoRoot
  if ([System.IO.Path]::IsPathRooted($Path)) {
    return [System.IO.Path]::GetFullPath($Path)
  }
  return [System.IO.Path]::GetFullPath((Join-Path $repoRoot $Path))
}

function Assert-AexsWorkspacePath {
  param([string]$Path, [string]$Purpose)
  $repoRoot = (Get-AexsRepoRoot).TrimEnd('\', '/')
  $fullPath = [System.IO.Path]::GetFullPath($Path).TrimEnd('\', '/')
  $prefix = $repoRoot + [System.IO.Path]::DirectorySeparatorChar
  if ($fullPath.Equals($repoRoot, [System.StringComparison]::OrdinalIgnoreCase)) {
    throw "Refusing to use repository root as $Purpose`: $fullPath"
  }
  if (-not $fullPath.StartsWith($prefix, [System.StringComparison]::OrdinalIgnoreCase)) {
    throw "Refusing to use path outside repository as $Purpose`: $fullPath"
  }
}

function Get-AexsRelativePath {
  param([string]$BasePath, [string]$TargetPath)
  $base = [System.IO.Path]::GetFullPath($BasePath).TrimEnd('\', '/')
  $target = [System.IO.Path]::GetFullPath($TargetPath)
  $prefix = $base + [System.IO.Path]::DirectorySeparatorChar
  if ($target.StartsWith($prefix, [System.StringComparison]::OrdinalIgnoreCase)) {
    return $target.Substring($prefix.Length)
  }
  return $target
}

function Invoke-AexsTextCommand {
  param([string]$FilePath, [string[]]$Arguments)
  try {
    $output = & $FilePath @Arguments 2>$null
    if ($LASTEXITCODE -ne 0) {
      return ""
    }
    return ($output -join "`n").Trim()
  } catch {
    return ""
  }
}

function Get-AexsGoVersion {
  $repoRoot = Get-AexsRepoRoot
  $bundled = Join-Path $repoRoot ".work\tools\go1.25.11\go\bin\go.exe"
  if (Test-Path -LiteralPath $bundled) {
    return Invoke-AexsTextCommand -FilePath $bundled -Arguments @("version")
  }
  $go = Get-Command go -ErrorAction SilentlyContinue
  if ($null -ne $go) {
    return Invoke-AexsTextCommand -FilePath $go.Source -Arguments @("version")
  }
  return "go version unavailable"
}

function Get-AexsSha256Hex {
  param([string]$Text)
  $sha = [System.Security.Cryptography.SHA256]::Create()
  try {
    $bytes = [System.Text.Encoding]::UTF8.GetBytes($Text)
    $hash = $sha.ComputeHash($bytes)
    return ([System.BitConverter]::ToString($hash)).Replace("-", "").ToLowerInvariant()
  } finally {
    $sha.Dispose()
  }
}

function Get-AexsTaskCount {
  param([string]$Text, [string]$Prefix)
  return ([regex]::Matches($Text, "(?m)^- \[ \]\s+$([regex]::Escape($Prefix))-\d{2}\b")).Count
}

function Get-AexsMatrixRowText {
  param([string]$Text, [string]$Label)
  if ($Text -match "(?m)^\|\s*$([regex]::Escape($Label))\s*\|.*$") {
    return $Matches[0]
  }
  return ""
}

function Test-AexsMatrixRow {
  param([string]$Text, [string]$Label)
  return (Get-AexsMatrixRowText -Text $Text -Label $Label) -ne ""
}

function Get-AexsMatrixCells {
  param([string]$Row)
  if ([string]::IsNullOrWhiteSpace($Row)) {
    return @()
  }
  return @($Row.Trim().Trim('|').Split('|') | ForEach-Object { $_.Trim() })
}

function Test-AexsTextAny {
  param([string]$Text, [string[]]$Terms)
  foreach ($term in $Terms) {
    if ($Text -match [regex]::Escape($term)) {
      return $true
    }
  }
  return $false
}

function Get-AexsMarkdownSection {
  param([string]$Text, [string]$Heading)
  $lines = $Text -split "`r?`n"
  $section = @()
  $capture = $false
  foreach ($line in $lines) {
    if ($line -eq "## $Heading") {
      $capture = $true
      continue
    }
    if ($capture -and $line -match '^##\s+') {
      break
    }
    if ($capture) {
      $section += $line
    }
  }
  return ($section -join "`n")
}

function Get-AexsMissingTerms {
  param([string]$Text, [string[]]$Terms)
  $missing = @()
  foreach ($term in $Terms) {
    if (-not (Test-AexsTextAny -Text $Text -Terms @($term))) {
      $missing += $term
    }
  }
  return $missing
}

function Get-AexsOverrideValue {
  param([object]$Override, [string]$Field, [object]$Fallback)
  if ($null -ne $Override -and $Override.Contains($Field)) {
    return $Override[$Field]
  }
  return $Fallback
}

function Get-AexsAtomicTaskOverride {
  param([string]$TaskId)
  $overrides = @{
    "AUTH-01" = [ordered]@{
      flow                = "valid signature verification, signer extraction, account sequence increment"
      state               = "accepted tx increments account sequence exactly once"
      attack              = "valid tx baseline plus mismatched signer control sample"
      invariant           = "valid signer can mutate only authorized account state"
      expected_behavior   = "valid signatures and signer extraction pass ante validation and execute exactly once"
      expected_events     = "auth success path emits stable tx/auth events and downstream message events"
      expected_error_path = "mismatched signer control sample is rejected before message execution"
      mutation_inputs     = "valid signature baseline, wrong signer control, duplicated signer control"
      expected_rejection  = "wrong or duplicate signer variants must fail without sequence mutation"
    }
    "AUTH-02" = [ordered]@{
      flow                = "empty signer set, duplicate signers, malformed public keys, wrong chain id, max-size auth info"
      state               = "rejected auth edge cases leave account sequence and balances unchanged"
      attack              = "malformed public key, wrong chain id, duplicate signer, max-size auth info"
      invariant           = "malformed auth info cannot mutate account state"
      expected_behavior   = "edge-case auth inputs are rejected deterministically with stable errors"
      expected_events     = "rejected edge cases emit no success events"
      expected_error_path = "validation rejects before fee/message state transition when auth info is invalid"
      mutation_inputs     = "empty signer array, duplicate signer array, malformed pubkey bytes, wrong chain id, oversized auth info"
      expected_rejection  = "all invalid auth edge cases fail before account sequence increment"
    }
    "AUTH-03" = [ordered]@{
      flow                = "invalid signature injection, replayed signed bytes, tx malleability, nonce manipulation"
      state               = "replay and malleability attempts cannot execute or alter sequence twice"
      attack              = "invalid signature, replayed tx bytes, mutated sign bytes, stale/future nonce"
      invariant           = "same signed transaction cannot execute twice"
      expected_behavior   = "auth adversarial mutations fail deterministically before state mutation"
      expected_events     = "failed replay or signature mutation emits no success events"
      expected_error_path = "ante signature or sequence decorator rejects before message server execution"
      mutation_inputs     = "bit-flipped signature, replayed accepted tx bytes, altered sign doc, stale sequence, future sequence"
      expected_rejection  = "invalid, replayed, or malleated tx must fail without sequence or balance mutation"
    }
    "AUTH-04" = [ordered]@{
      flow                = "rejected auth path state integrity"
      state               = "failed auth path does not increment sequence or mutate account state"
      attack              = "state corruption attempt through rejected auth path"
      invariant           = "invalid signer cannot mutate state"
      expected_behavior   = "auth rejection preserves account number, sequence, balances, and module state"
      expected_events     = "no state-changing events after rejected auth path"
      expected_error_path = "auth failure returns before fee deduction or message execution"
      mutation_inputs     = "invalid signer, invalid sequence, malformed signer, rejected fee payer"
      expected_rejection  = "state snapshot before and after rejected auth path must match"
    }
    "AUTH-05" = [ordered]@{
      flow                = "auth failure interaction with fee deduction, priority rules, and rate limits"
      state               = "auth failure cannot charge or bypass fees outside expected ante semantics"
      attack              = "fee bypass, priority bypass, rate-limit bypass through auth failure"
      invariant           = "auth failure cannot bypass fee deduction, priority, or rate limits"
      expected_behavior   = "invalid auth cannot be used to gain free execution or priority"
      expected_events     = "no fee distribution, priority success, or rate-limit success event after invalid auth"
      expected_error_path = "invalid auth fails before any message execution and cannot mark tx as accepted"
      mutation_inputs     = "invalid signer with high priority, invalid signer with zero fee, repeated invalid signer spam"
      expected_rejection  = "invalid auth must not execute messages, bypass fee policy, or consume accepted priority lane"
    }
    "BANK-01" = [ordered]@{
      flow                = "naet sends, module account transfers, multi-send success paths"
      state               = "sender, recipient, module account balances, and supply remain consistent after accepted sends"
      attack              = "valid transfer baseline plus unauthorized module transfer control sample"
      invariant           = "accepted bank sends preserve total supply and authorization"
      expected_behavior   = "valid naet sends and valid multi-sends update balances exactly by transferred amount"
      expected_events     = "bank transfer events match committed balance deltas"
      expected_error_path = "unauthorized module transfer control sample is rejected before balance mutation"
      mutation_inputs     = "valid send, valid multi-send, valid module transfer, unauthorized module transfer control"
      expected_rejection  = "unauthorized module movement must fail without partial balance update"
    }
    "BANK-02" = [ordered]@{
      flow                = "zero amount, max amount, insufficient funds, malformed denom, zero address, self-transfer"
      state               = "bank edge cases either update balances exactly or leave state unchanged"
      attack              = "zero amount, max amount, insufficient funds, malformed denom, zero address, self-transfer"
      invariant           = "bank edge cases cannot create negative balances or invalid denom state"
      expected_behavior   = "valid boundaries execute deterministically; invalid boundaries reject before mutation"
      expected_events     = "accepted boundary sends emit accurate transfer events; rejected paths emit no success events"
      expected_error_path = "invalid amount, denom, address, or insufficient funds rejects before balance mutation"
      mutation_inputs     = "zero coin, max int coin, insufficient funds coin, malformed denom, zero address recipient, self-transfer"
      expected_rejection  = "invalid boundary inputs must not mutate balances or supply"
    }
    "BANK-03" = [ordered]@{
      flow                = "double spend, partial multi-send failure, negative balance creation, overflow"
      state               = "failed adversarial bank operations leave all account and module balances unchanged"
      attack              = "double spend, partial multi-send failure, negative balance attempt, arithmetic overflow attempt"
      invariant           = "no negative balances and no partial multi-send commits"
      expected_behavior   = "adversarial bank mutations fail atomically"
      expected_events     = "failed adversarial sends emit no misleading success transfer events"
      expected_error_path = "bank validation or keeper send path rejects before partial commit"
      mutation_inputs     = "two spends of same funds, multi-send with one invalid output, negative-like amount encoding, overflow-size amount"
      expected_rejection  = "failed bank attack must not create funds, negative balances, or partial recipient credits"
    }
    "BANK-04" = [ordered]@{
      flow                = "balance, module balance, and total supply state integrity"
      state               = "balances, module balances, and total supply remain consistent after accepted and rejected sends"
      attack              = "state drift attempt through mixed accepted and rejected bank sends"
      invariant           = "sum(account balances) plus module balances equals total supply"
      expected_behavior   = "bank state integrity holds across multi-step send sequences"
      expected_events     = "events reconcile to final committed balance deltas"
      expected_error_path = "failed sends preserve pre-failure balance and supply snapshot"
      mutation_inputs     = "accepted send followed by failed send, module transfer followed by failed multi-send, export/import after sends"
      expected_rejection  = "any rejected send must preserve total supply and module/account balance consistency"
    }
    "BANK-05" = [ordered]@{
      flow                = "bank economic abuse around mint, burn, native denom metadata, and protocol fee denom"
      state               = "bank paths cannot change supply except through authorized module mint/burn paths"
      attack              = "unauthorized mint, unauthorized burn, native denom spoof, non-naet protocol fee payment"
      invariant           = "bank cannot mint, burn, spoof native denom metadata, or pay protocol fees with non-naet assets"
      expected_behavior   = "bank transfers move existing balances only and cannot alter native token authority"
      expected_events     = "no mint/burn/native metadata event from plain bank send"
      expected_error_path = "unauthorized supply or fee-denom abuse rejects before state mutation"
      mutation_inputs     = "bank send shaped as mint, bank send shaped as burn, denom metadata spoof, non-naet fee asset"
      expected_rejection  = "bank economic abuse must not alter supply, native metadata, or protocol fee acceptance"
    }
    "STAKE-01" = [ordered]@{
      flow                = "validator creation, delegation, redelegation, unbonding, reward eligibility"
      state               = "validator records, delegations, redelegations, unbonding entries, and reward eligibility update deterministically"
      attack              = "valid staking lifecycle baseline plus unauthorized operator control sample"
      invariant           = "validator set and staking pool remain consistent after accepted staking lifecycle operations"
      expected_behavior   = "valid staking lifecycle messages update validator power, delegator shares, and unbonding records exactly once"
      expected_events     = "staking lifecycle events match committed validator and delegation deltas"
      expected_error_path = "unauthorized operator or delegator control sample is rejected before staking state mutation"
      mutation_inputs     = "valid create-validator, valid delegate, valid redelegate, valid unbond, unauthorized operator control"
      expected_rejection  = "unauthorized staking lifecycle variants must fail without validator power or delegation mutation"
    }
    "STAKE-02" = [ordered]@{
      flow                = "invalid validator address, non-naet bond denom, zero self-delegation, max commission, unbonding window boundaries"
      state               = "invalid staking edge cases leave validator, delegation, and pool state unchanged"
      attack              = "bad validator address, wrong bond denom, zero self-delegation, extreme commission, boundary unbonding window"
      invariant           = "staking accepts only valid validator addresses, naet bond denom, bounded commission, and valid unbonding windows"
      expected_behavior   = "valid staking boundaries execute deterministically; invalid boundaries reject before staking mutation"
      expected_events     = "accepted boundary staking events match state deltas; rejected edge cases emit no success events"
      expected_error_path = "staking message validation rejects invalid address, denom, self-delegation, commission, or unbonding boundary"
      mutation_inputs     = "malformed validator address, non-naet bond denom, zero self-delegation, max commission, expired or premature unbonding window"
      expected_rejection  = "invalid staking edge cases must not alter validator power, delegator shares, or pools"
    }
    "STAKE-03" = [ordered]@{
      flow                = "stake grinding, delegation manipulation, reward farming loop, validator power spoofing"
      state               = "adversarial staking attempts cannot create unauthorized power, shares, or reward eligibility"
      attack              = "stake grinding, delegation manipulation, reward farming loop, validator power spoofing"
      invariant           = "validator power and delegator shares derive only from valid bonded stake"
      expected_behavior   = "adversarial staking mutations fail or resolve deterministically without extra rewards"
      expected_events     = "failed staking attacks emit no misleading validator power or reward events"
      expected_error_path = "staking keeper or ante validation rejects invalid stake manipulation before commit"
      mutation_inputs     = "rapid redelegation loop, repeated delegate/unbond loop, spoofed validator operator, manipulated shares, reward timing loop"
      expected_rejection  = "staking attacks must not inflate validator power, delegator shares, or rewards"
    }
    "STAKE-04" = [ordered]@{
      flow                = "validator tokens, delegator shares, staking pools, validator-set update integrity"
      state               = "validator tokens, shares, pools, and validator updates reconcile after accepted and rejected staking operations"
      attack              = "state drift attempt through mixed accepted and rejected staking lifecycle operations"
      invariant           = "validator tokens, delegator shares, staking pools, and validator set updates remain consistent"
      expected_behavior   = "staking state integrity holds across lifecycle sequences and export/import"
      expected_events     = "staking events reconcile to validator power and pool deltas"
      expected_error_path = "failed staking operations preserve pre-failure validator, delegation, pool, and validator-update state"
      mutation_inputs     = "accepted delegate followed by failed redelegate, accepted unbond followed by failed delegate, export/import after validator updates"
      expected_rejection  = "rejected staking operations must preserve staking pool and validator-set consistency"
    }
    "STAKE-05" = [ordered]@{
      flow                = "staking economic abuse around reward inflation, unbonding risk, and slash-immune stake"
      state               = "staking paths cannot inflate rewards, bypass unbonding, or create slash-immune stake"
      attack              = "reward inflation, unbonding risk bypass, slash-immune stake creation"
      invariant           = "staking cannot inflate rewards, bypass unbonding risk, or create slash-immune stake"
      expected_behavior   = "staking economic rules preserve bonded risk and reward eligibility"
      expected_events     = "no reward or unbonding success event appears for rejected economic abuse paths"
      expected_error_path = "economic abuse rejects before reward, unbonding, or slash-protection state mutation"
      mutation_inputs     = "delegate/redelegate reward loop, immediate unbond bypass, hidden self-delegation, slash-exempt stake marker"
      expected_rejection  = "staking economic abuse must not create extra rewards, bypass unbonding, or avoid slashing exposure"
    }
    "SLASH-01" = [ordered]@{
      flow                = "downtime evidence, equivocation evidence, validator status update, stake penalty"
      state               = "valid slashing evidence updates signing info, validator status, slash amount, and validator set deterministically"
      attack              = "valid evidence baseline plus mismatched validator control sample"
      invariant           = "valid objective evidence causes deterministic stake penalty and validator status change"
      expected_behavior   = "valid downtime or equivocation evidence slashes and jails/tombstones according to params"
      expected_events     = "slashing events match validator status and stake penalty deltas"
      expected_error_path = "mismatched validator evidence control sample is rejected before slashing state mutation"
      mutation_inputs     = "valid downtime evidence, valid equivocation evidence, mismatched validator evidence control"
      expected_rejection  = "invalid evidence must not slash or alter validator status"
    }
    "SLASH-02" = [ordered]@{
      flow                = "stale evidence, duplicate evidence, unknown validator, jailed validator, boundary heights"
      state               = "invalid slashing edge cases leave validator status, signing info, and stake unchanged"
      attack              = "stale evidence, duplicate evidence, unknown validator, jailed validator replay, boundary height evidence"
      invariant           = "slashing evidence is objective, fresh, unique, and bound to known validators"
      expected_behavior   = "valid boundary evidence applies once; invalid edge evidence rejects deterministically"
      expected_events     = "duplicate or stale evidence emits no second slash success event"
      expected_error_path = "slashing evidence validation rejects stale, duplicate, unknown, or invalid-height evidence"
      mutation_inputs     = "stale evidence height, duplicate evidence bytes, unknown validator address, already jailed validator, min/max boundary height"
      expected_rejection  = "invalid slashing edge cases must not double-slash or alter stake"
    }
    "SLASH-03" = [ordered]@{
      flow                = "slashing bypass through redelegation, unbonding, delayed evidence, malformed proof"
      state               = "slashing exposure persists across redelegation and unbonding windows"
      attack              = "redelegation slash evasion, unbonding slash evasion, delayed evidence bypass, malformed proof acceptance"
      invariant           = "slashing cannot be bypassed by stake movement or malformed evidence"
      expected_behavior   = "slashable stake remains slashable across valid evidence windows"
      expected_events     = "slashing or rejection events reconcile with evidence validity and stake exposure"
      expected_error_path = "malformed proof rejects while valid delayed evidence still applies within protocol window"
      mutation_inputs     = "redelegate before evidence, unbond before evidence, delayed evidence within window, malformed evidence proof"
      expected_rejection  = "bypass attempts must not protect slashable stake or accept malformed proof"
    }
    "SLASH-04" = [ordered]@{
      flow                = "slash accounting, jailed/tombstoned state, validator-set removal determinism"
      state               = "slash accounting, jail/tombstone flags, and validator updates remain deterministic after slashing"
      attack              = "state drift attempt through repeated slashing, jail, tombstone, and validator update paths"
      invariant           = "slash accounting, jailed/tombstoned state, and validator-set removal are deterministic"
      expected_behavior   = "slashing state integrity holds across evidence sequences and export/import"
      expected_events     = "slashing events reconcile to stake penalty and validator update output"
      expected_error_path = "failed or duplicate slashing paths preserve pre-failure signing info and validator state"
      mutation_inputs     = "valid slash then duplicate slash, tombstone then duplicate evidence, export/import after validator removal"
      expected_rejection  = "duplicate or invalid slashing paths must not alter slash accounting twice"
    }
    "SLASH-05" = [ordered]@{
      flow                = "slashed stake recovery through timing, migration, export/import, governance parameter race"
      state               = "slashed stake cannot be recovered by timing, migration, export/import, or parameter races"
      attack              = "timing recovery, migration recovery, export/import recovery, governance parameter race"
      invariant           = "slashed stake and tombstone state remain final unless protocol explicitly permits recovery"
      expected_behavior   = "slashing economic finality survives operational and governance boundary paths"
      expected_events     = "no stake recovery event appears for rejected slashing recovery paths"
      expected_error_path = "recovery attempts reject or preserve slashed/tombstoned state during migration/export/import/param changes"
      mutation_inputs     = "slash then parameter change, slash then export/import, slash then migration, slash then unjail timing race"
      expected_rejection  = "slashed stake recovery attempts must not restore stake or remove tombstone incorrectly"
    }
  }
  if ($overrides.ContainsKey($TaskId)) {
    return $overrides[$TaskId]
  }
  return $null
}

function Get-AexsAtomicTaskRecords {
  param(
    [string]$Text,
    [object]$Module,
    [string[]]$MatrixCells,
    [string]$CampaignId
  )

  $records = @()
  $current = $null
  $lines = $Text -split "`r?`n"
  $taskPattern = "^- \[ \]\s+($([regex]::Escape($Module["Prefix"]))-\d{2})\s+([^:]+):\s*(.*)$"
  foreach ($line in $lines) {
    if ($line -match $taskPattern) {
      if ($null -ne $current) {
        $records += $current
      }
      $current = [ordered]@{
        task_id     = $Matches[1]
        task_type   = $Matches[2].Trim()
        description = $Matches[3].Trim()
      }
      continue
    }

    if ($null -eq $current) {
      continue
    }

    if ($line -match '^- \[ \]\s+[A-Z]+-\d{2}\b' -or $line -match '^###\s+' -or $line -match '^##\s+') {
      $records += $current
      $current = $null
      continue
    }

    if ($line -match '^\s+\S') {
      $current["description"] = ($current["description"] + " " + $line.Trim()).Trim()
    }
  }
  if ($null -ne $current) {
    $records += $current
  }

  $functionCell = if ($MatrixCells.Count -ge 2) { $MatrixCells[1] } else { "" }
  $stateCell = if ($MatrixCells.Count -ge 3) { $MatrixCells[2] } else { "" }
  $attackCell = if ($MatrixCells.Count -ge 4) { $MatrixCells[3] } else { "" }
  $invariantCell = if ($MatrixCells.Count -ge 5) { $MatrixCells[4] } else { "" }

  $out = @()
  foreach ($record in $records) {
    $override = Get-AexsAtomicTaskOverride -TaskId $record["task_id"]
    $seedHash = (Get-AexsSha256Hex -Text "$CampaignId|$($Module["Module"])|$($record["task_id"])").Substring(0, 16)
    $seed = "aexs-$($record["task_id"].ToLowerInvariant())-$seedHash"
    $flow = if ([string]::IsNullOrWhiteSpace($functionCell)) { $record["description"] } else { $functionCell }
    $attack = if ([string]::IsNullOrWhiteSpace($attackCell)) { $record["description"] } else { $attackCell }
    $invariant = if ([string]::IsNullOrWhiteSpace($invariantCell)) { "module-specific invariant from TO_AUDIT task $($record["task_id"])" } else { $invariantCell }
    $state = if ([string]::IsNullOrWhiteSpace($stateCell)) { "state transition from TO_AUDIT task $($record["task_id"])" } else { $stateCell }
    $flow = Get-AexsOverrideValue -Override $override -Field "flow" -Fallback $flow
    $attack = Get-AexsOverrideValue -Override $override -Field "attack" -Fallback $attack
    $invariant = Get-AexsOverrideValue -Override $override -Field "invariant" -Fallback $invariant
    $state = Get-AexsOverrideValue -Override $override -Field "state" -Fallback $state
    $expectedBehavior = Get-AexsOverrideValue -Override $override -Field "expected_behavior" -Fallback $record["description"]
    $expectedEvents = Get-AexsOverrideValue -Override $override -Field "expected_events" -Fallback "stable module events or no events for rejected path"
    $expectedErrorPath = Get-AexsOverrideValue -Override $override -Field "expected_error_path" -Fallback "malformed, unauthorized, replayed, or boundary input must fail before unintended state mutation"
    $mutationInputs = Get-AexsOverrideValue -Override $override -Field "mutation_inputs" -Fallback "malformed input, replay, unauthorized signer, bad fee, boundary values, state corruption attempt, or module-specific exploit"
    $expectedRejection = Get-AexsOverrideValue -Override $override -Field "expected_rejection" -Fallback "attack must not violate invariant or mutate state outside the expected transition"

    $out += [ordered]@{
      module                         = $Module["Module"]
      task_id                        = $record["task_id"]
      task_type                      = $record["task_type"]
      function_or_flow_covered       = $flow
      state_transition_covered       = $state
      attack_surface_covered         = $attack
      invariant_tested               = $invariant
      defensive_analysis_result      = [ordered]@{
        status                    = "planned_not_executed"
        expected_behavior         = $expectedBehavior
        expected_state_transition = $state
        expected_events           = $expectedEvents
        expected_error_path       = $expectedErrorPath
        expected_invariant        = $invariant
      }
      adversarial_simulation_result  = [ordered]@{
        status             = "planned_not_executed"
        attack_attempt     = $attack
        mutation_inputs    = $mutationInputs
        expected_rejection = $expectedRejection
        replay_mode        = "deterministic replay by seed and step list"
      }
      pass_fail_result               = "not_executed"
      reproduction_seed_or_steps     = [ordered]@{
        seed  = $seed
        steps = @(
          "Run AEXS campaign for task $($record["task_id"])",
          "Use seed $seed",
          "Record defensive analysis result",
          "Record adversarial simulation result",
          "Update pass_fail_result with pass or fail after execution"
        )
      }
      valid                          = $true
      invalid_reasons                = @()
    }
  }
  return $out
}

function Test-AexsAtomicTaskRecord {
  param([object]$Task)
  $reasons = @()
  foreach ($field in @(
      "module",
      "task_id",
      "function_or_flow_covered",
      "state_transition_covered",
      "attack_surface_covered",
      "invariant_tested",
      "pass_fail_result"
    )) {
    if ([string]::IsNullOrWhiteSpace([string]$Task[$field])) {
      $reasons += "missing $field"
    }
  }
  foreach ($field in @(
      "status",
      "expected_behavior",
      "expected_state_transition",
      "expected_events",
      "expected_error_path",
      "expected_invariant"
    )) {
    if ([string]::IsNullOrWhiteSpace([string]$Task["defensive_analysis_result"][$field])) {
      $reasons += "missing defensive_analysis_result.$field"
    }
  }
  foreach ($field in @(
      "status",
      "attack_attempt",
      "mutation_inputs",
      "expected_rejection",
      "replay_mode"
    )) {
    if ([string]::IsNullOrWhiteSpace([string]$Task["adversarial_simulation_result"][$field])) {
      $reasons += "missing adversarial_simulation_result.$field"
    }
  }
  if ($null -eq $Task["reproduction_seed_or_steps"] -or [string]::IsNullOrWhiteSpace([string]$Task["reproduction_seed_or_steps"]["seed"])) {
    $reasons += "missing reproduction seed"
  }
  if ($null -eq $Task["reproduction_seed_or_steps"] -or @($Task["reproduction_seed_or_steps"]["steps"]).Count -eq 0) {
    $reasons += "missing reproduction steps"
  }
  return $reasons
}

function Get-AexsEvidence {
  param([object]$Module)
  $repoRoot = Get-AexsRepoRoot
  $files = @()
  foreach ($root in $Module.EvidenceRoots) {
    $fullRoot = Resolve-AexsPath -Path $root
    if (-not (Test-Path -LiteralPath $fullRoot)) {
      continue
    }
    if (Test-Path -LiteralPath $fullRoot -PathType Leaf) {
      $files += Get-Item -LiteralPath $fullRoot
      continue
    }
    $files += Get-ChildItem -LiteralPath $fullRoot -Recurse -File -Include *.go,*.md,*.ps1
  }

  $matchedFiles = @()
  $invariantFiles = @()
  $fuzzFiles = @()
  $adversarialFiles = @()
  $determinismFiles = @()
  foreach ($file in $files) {
    $text = Get-Content -Raw -LiteralPath $file.FullName
    $matched = $false
    foreach ($term in $Module.EvidenceTerms) {
      if ($text -match [regex]::Escape($term)) {
        $matched = $true
        break
      }
    }
    if (-not $matched) {
      continue
    }
    $relative = Get-AexsRelativePath -BasePath $repoRoot -TargetPath $file.FullName
    $matchedFiles += $relative
    if ($text -match '(?i)invariant|invariants|state integrity|supply consistency|app hash|deterministic replay') {
      $invariantFiles += $relative
    }
    if ($text -match '(?i)\bFuzz[A-Za-z0-9_]*|fuzz') {
      $fuzzFiles += $relative
    }
    if ($text -match '(?i)adversarial|attack|exploit|malformed|unauthorized|replay') {
      $adversarialFiles += $relative
    }
    if ($text -match '(?i)deterministic|determinism|same input|same tx|same genesis') {
      $determinismFiles += $relative
    }
  }

  return [ordered]@{
    files              = @($matchedFiles | Sort-Object -Unique)
    invariant_files    = @($invariantFiles | Sort-Object -Unique)
    fuzz_files         = @($fuzzFiles | Sort-Object -Unique)
    adversarial_files  = @($adversarialFiles | Sort-Object -Unique)
    determinism_files  = @($determinismFiles | Sort-Object -Unique)
  }
}

$repoRoot = Get-AexsRepoRoot
$taskPath = Resolve-AexsPath -Path $TaskFile
$pipelinePath = Resolve-AexsPath -Path $PipelineDoc
$outputRoot = Resolve-AexsPath -Path $OutputDir
Assert-AexsWorkspacePath -Path $outputRoot -Purpose "AEXS output directory"

if (-not (Test-Path -LiteralPath $taskPath)) {
  throw "AEXS task file not found: $taskPath"
}
if (-not (Test-Path -LiteralPath $pipelinePath)) {
  throw "AEXS pipeline source not found: $pipelinePath"
}

$taskText = Get-Content -Raw -LiteralPath $taskPath
$pipelineText = Get-Content -Raw -LiteralPath $pipelinePath
$commit = Invoke-AexsTextCommand -FilePath "git" -Arguments @("rev-parse", "--short=12", "HEAD")
if ([string]::IsNullOrWhiteSpace($commit)) {
  $commit = "no-git-commit"
}
$branch = Invoke-AexsTextCommand -FilePath "git" -Arguments @("branch", "--show-current")
if ([string]::IsNullOrWhiteSpace($branch)) {
  $branch = "detached-or-unknown"
}
$dirtyStatus = Invoke-AexsTextCommand -FilePath "git" -Arguments @("status", "--short")
$sourceHash = (Get-AexsSha256Hex -Text ($taskText + "`n---PIPELINE---`n" + $pipelineText)).Substring(0, 16)
$campaignId = "aexs-$commit-$sourceHash"
$campaignDir = Join-Path $outputRoot $campaignId
New-Item -ItemType Directory -Force -Path $campaignDir | Out-Null

$moduleCatalog = @(
  [ordered]@{ Module = "x/auth"; Label = '`x/auth`'; Prefix = "AUTH"; Value = $true; EvidenceRoots = @("app", "tests\integration", "tests\adversarial", "docs\transaction-lifecycle-matrix.md"); EvidenceTerms = @("auth", "signature", "sequence", "signer") },
  [ordered]@{ Module = "x/bank"; Label = '`x/bank`'; Prefix = "BANK"; Value = $true; EvidenceRoots = @("app", "tests\integration", "tests\adversarial", "docs\transaction-lifecycle-matrix.md"); EvidenceTerms = @("bank", "balance", "send", "supply") },
  [ordered]@{ Module = "x/staking"; Label = '`x/staking`'; Prefix = "STAKE"; Value = $true; EvidenceRoots = @("app", "tests\integration", "tests\e2e", "docs\security\pos-staking-correctness.md"); EvidenceTerms = @("staking", "delegate", "validator", "unbond") },
  [ordered]@{ Module = "x/slashing"; Label = '`x/slashing`'; Prefix = "SLASH"; Value = $true; EvidenceRoots = @("app", "docs\security\slashing-system.md", "docs\security\pos-staking-correctness.md"); EvidenceTerms = @("slashing", "slash", "evidence", "tombstone") },
  [ordered]@{ Module = "x/gov"; Label = '`x/gov`'; Prefix = "GOV"; Value = $true; EvidenceRoots = @("app", "docs", "tests\integration"); EvidenceTerms = @("governance", "proposal", "vote", "authority") },
  [ordered]@{ Module = "x/distribution"; Label = '`x/distribution`'; Prefix = "DIST"; Value = $true; EvidenceRoots = @("app", "docs\security\pos-staking-correctness.md", "tests\integration"); EvidenceTerms = @("distribution", "reward", "commission", "community pool") },
  [ordered]@{ Module = "app"; Label = '`app` / BaseApp'; Prefix = "APP"; Value = $true; EvidenceRoots = @("app", "tests\integration", "docs\genesis-migrations.md", "docs\state-export-import.md"); EvidenceTerms = @("BaseApp", "app hash", "genesis", "export", "determinism") },
  [ordered]@{ Module = "x/fees"; Label = '`x/fees`'; Prefix = "FEES"; Value = $true; EvidenceRoots = @("x\fees", "tests\adversarial", "tests\integration", "docs\fees-ante-policy.md"); EvidenceTerms = @("fees", "fee", "naet", "ante") },
  [ordered]@{ Module = "x/tokenfactory"; Label = '`x/tokenfactory`'; Prefix = "TF"; Value = $true; EvidenceRoots = @("x\tokenfactory", "tests\adversarial", "tests\e2e", "docs\security\module-bank-movement-audit.md"); EvidenceTerms = @("tokenfactory", "mint", "burn", "admin") },
  [ordered]@{ Module = "x/dex"; Label = '`x/dex`'; Prefix = "DEX"; Value = $true; EvidenceRoots = @("x\dex", "tests\adversarial", "tests\e2e", "docs\architecture\dex-direction.md"); EvidenceTerms = @("dex", "pool", "swap", "liquidity", "reserve") },
  [ordered]@{ Module = "x/identity"; Label = '`x/identity`'; Prefix = "ID"; Value = $true; EvidenceRoots = @("x\identity", "tests\adversarial", "docs\architecture\aetheris-modular-execution-os.md"); EvidenceTerms = @("identity", ".aet", "domain", "resolver") },
  [ordered]@{ Module = "x/reputation"; Label = '`x/reputation`'; Prefix = "REP"; Value = $true; EvidenceRoots = @("x\reputation", "docs\module-boundaries.md", "docs\test-production-gates.md"); EvidenceTerms = @("reputation", "score", "rate limit", "priority") },
  [ordered]@{ Module = "x/execution"; Label = '`x/execution`'; Prefix = "EXEC"; Value = $true; EvidenceRoots = @("x\execution", "docs\architecture\execution-os.md", "docs\module-boundaries.md"); EvidenceTerms = @("execution", "dispatch", "route", "receipt") },
  [ordered]@{ Module = "x/vm"; Label = '`x/vm` / AVM'; Prefix = "VM"; Value = $true; EvidenceRoots = @("x\vm", "x\aetherisvm", "docs\architecture\avm.md", "docs\architecture\vm-direction.md"); EvidenceTerms = @("vm", "AVM", "bytecode", "gas") },
  [ordered]@{ Module = "x/aetherisvm"; Label = '`x/vm` / AVM'; Prefix = "VM"; Value = $true; EvidenceRoots = @("x\aetherisvm", "docs\architecture\avm.md", "docs\architecture\vm-direction.md"); EvidenceTerms = @("AVM", "async", "contract", "gas") },
  [ordered]@{ Module = "x/messaging"; Label = '`x/messaging`'; Prefix = "MSG"; Value = $true; EvidenceRoots = @("x\messaging", "x\mesh", "tests\adversarial", "docs\architecture\execution-os.md"); EvidenceTerms = @("messaging", "message", "receipt", "proof") },
  [ordered]@{ Module = "x/queue"; Label = '`x/queue`'; Prefix = "QUEUE"; Value = $true; EvidenceRoots = @("x\queue", "x\aetherisvm\async", "docs\architecture\async-smart-contract-execution.md"); EvidenceTerms = @("queue", "bounce", "refund", "delayed") },
  [ordered]@{ Module = "x/events"; Label = '`x/events`'; Prefix = "EVENTS"; Value = $false; EvidenceRoots = @("x\events", "docs\event-contract.md", "tests\scripts\event_contract_doc_test.ps1"); EvidenceTerms = @("events", "event", "receipt", "attributes") },
  [ordered]@{ Module = "x/actors"; Label = '`x/actors`'; Prefix = "ACTOR"; Value = $true; EvidenceRoots = @("x\actors", "docs\module-boundaries.md"); EvidenceTerms = @("actor", "mailbox", "logical time") },
  [ordered]@{ Module = "x/scheduler"; Label = '`x/scheduler`'; Prefix = "SCHED"; Value = $true; EvidenceRoots = @("x\scheduler", "x\schedulerv2", "docs\module-boundaries.md"); EvidenceTerms = @("scheduler", "schedule", "task", "priority") },
  [ordered]@{ Module = "x/storage"; Label = '`x/storage`'; Prefix = "STORE"; Value = $true; EvidenceRoots = @("x\storage", "docs\module-boundaries.md", "docs\architecture\avm.md"); EvidenceTerms = @("storage", "snapshot", "export", "state root") },
  [ordered]@{ Module = "x/memo"; Label = '`x/memo`'; Prefix = "MEMO"; Value = $true; EvidenceRoots = @("x\memo", "docs\mempool-checktx-negative-flow.md", "docs\transaction-lifecycle-matrix.md"); EvidenceTerms = @("memo", "UTF-8", "metadata") },
  [ordered]@{ Module = "x/indexer"; Label = '`x/indexer`'; Prefix = "INDEX"; Value = $false; EvidenceRoots = @("x\indexer", "app\indexer", "docs\event-contract.md", "docs\query-surface.md"); EvidenceTerms = @("index", "indexer", "query", "event") },
  [ordered]@{ Module = "x/sharding/sim"; Label = '`x/sharding/sim` and load/routing'; Prefix = "SHARD"; Value = $true; EvidenceRoots = @("x\sharding", "x\load", "x\routing", "tests\adversarial", "docs\architecture\sharding-rd.md"); EvidenceTerms = @("sharding", "LOAD_SCORE", "route", "shard") }
)

$requiredSourceTerms = @(
  "docs/security/aetheris-fuzzing-invariant-pipeline.md",
  ".work/aexs/",
  "100%",
  "95%",
  "defensive analysis",
  "adversarial simulation",
  "mandatory invariants",
  "Coverage Matrix"
)

$campaignSetupSection = Get-AexsMarkdownSection -Text $taskText -Heading "Campaign Setup Tasks"
$scenarioGeneratorSection = Get-AexsMarkdownSection -Text $taskText -Heading "Scenario Generator Tasks"
$transactionMutatorSection = Get-AexsMarkdownSection -Text $taskText -Heading "Transaction Mutator Tasks"

$requiredCampaignTerms = @(
  "Create a deterministic campaign id.",
  "Record git commit, branch, dirty status, Go version, OS, and test command",
  "Record the fuzz seed list.",
  "Record the target modules.",
  "stateless fuzzing",
  "stateful multi-block fuzzing",
  "adversarial red-team fuzzing",
  "deterministic replay",
  "stress mode",
  "chaos mode",
  "in-memory app runner",
  "single-validator localnet",
  "multi-validator localnet",
  "sharding simulator",
  "first critical exploit",
  "max run count",
  "max wall-clock duration",
  "coverage threshold reached",
  "deterministic divergence"
)

$requiredScenarioTerms = @(
  "Generate random bank transfer sequences.",
  "Generate random staking, delegation, unbonding, redelegation, and reward",
  "Generate random validator lifecycle and slashing evidence sequences.",
  "Generate random fee and spam bursts.",
  "Generate random tokenfactory create, mint, burn, and admin sequences.",
  "Generate random DEX create-pool, add-liquidity, remove-liquidity, and swap",
  "Generate random governance proposal, vote, and parameter update sequences.",
  "Generate random identity domain registration, auction, renewal, resolver",
  "Generate random AVM deploy, external call, internal call, bounced call",
  "Generate random async message, queue, delayed execution, bounce, and refund",
  "routing, and shard activation scenarios.",
  "Preserve seed and step list for every generated scenario."
)

$requiredMutatorTerms = @(
  "Inject invalid signatures.",
  "Replay already accepted transaction bytes.",
  "Manipulate nonce and sequence values.",
  "Corrupt fee denom and fee amount fields.",
  "fee paths.",
  "Inject extreme gas values.",
  "Inject malformed addresses.",
  "DEX actor fields.",
  "Corrupt memo fields, including invalid UTF-8 and oversized memo payloads.",
  "Inject malformed routing hints.",
  "Inject invalid domain resolution and expired domain actions.",
  "Inject fake cross-zone messages.",
  "Inject queue depth abuse.",
  "Inject oversized AVM payloads.",
  "Inject invalid AVM entrypoint inputs.",
  "Inject malformed genesis fragments for simulator startup tests.",
  "Record mutation metadata for every scenario."
)

$sourceFailures = @()
foreach ($term in $requiredSourceTerms) {
  if (-not (Test-AexsTextAny -Text $taskText -Terms @($term)) -and -not (Test-AexsTextAny -Text $pipelineText -Terms @($term))) {
    $sourceFailures += "missing source term: $term"
  }
}
if ([string]::IsNullOrWhiteSpace($campaignSetupSection)) {
  $sourceFailures += "missing Campaign Setup Tasks section"
} else {
  foreach ($term in @(Get-AexsMissingTerms -Text $campaignSetupSection -Terms $requiredCampaignTerms)) {
    $sourceFailures += "missing campaign setup term: $term"
  }
}
if ([string]::IsNullOrWhiteSpace($scenarioGeneratorSection)) {
  $sourceFailures += "missing Scenario Generator Tasks section"
} else {
  foreach ($term in @(Get-AexsMissingTerms -Text $scenarioGeneratorSection -Terms $requiredScenarioTerms)) {
    $sourceFailures += "missing scenario generator term: $term"
  }
}
if ([string]::IsNullOrWhiteSpace($transactionMutatorSection)) {
  $sourceFailures += "missing Transaction Mutator Tasks section"
} else {
  foreach ($term in @(Get-AexsMissingTerms -Text $transactionMutatorSection -Terms $requiredMutatorTerms)) {
    $sourceFailures += "missing transaction mutator term: $term"
  }
}

$moduleRows = @()
$atomicTasks = @()
foreach ($module in $moduleCatalog) {
  $taskCount = Get-AexsTaskCount -Text $taskText -Prefix $module.Prefix
  $matrixRow = Get-AexsMatrixRowText -Text $taskText -Label $module.Label
  $matrixCells = Get-AexsMatrixCells -Row $matrixRow
  $moduleAtomicTasks = @(Get-AexsAtomicTaskRecords -Text $taskText -Module $module -MatrixCells $matrixCells -CampaignId $campaignId)
  foreach ($task in $moduleAtomicTasks) {
    $invalidReasons = @(Test-AexsAtomicTaskRecord -Task $task)
    $task["valid"] = $invalidReasons.Count -eq 0
    $task["invalid_reasons"] = $invalidReasons
    $atomicTasks += $task
  }
  $hasMatrixRow = $matrixRow -ne ""
  $hasAttackSurface = $hasMatrixRow -and $matrixCells.Count -ge 5 -and -not [string]::IsNullOrWhiteSpace($matrixCells[3])
  $hasInvariantPlan = $hasMatrixRow -and $matrixCells.Count -ge 5 -and -not [string]::IsNullOrWhiteSpace($matrixCells[4])
  $hasValueTask = (-not $module.Value) -or ($taskText -match "$([regex]::Escape($module.Prefix))-05")
  $plannedChecks = @(($taskCount -ge 5), $hasMatrixRow, $hasAttackSurface, $hasInvariantPlan, $hasValueTask)
  $passedPlanned = @($plannedChecks | Where-Object { $_ -eq $true }).Count
  $plannedCoverage = [math]::Round(($passedPlanned / $plannedChecks.Count) * 100, 2)
  $evidence = Get-AexsEvidence -Module $module
  $hasInvariantEvidence = @($evidence.invariant_files).Count -gt 0
  $hasAdversarialEvidence = @($evidence.adversarial_files).Count -gt 0
  $hasFuzzEvidence = @($evidence.fuzz_files).Count -gt 0
  $safe = $false
  $reasons = @()
  if ($taskCount -lt 5) { $reasons += "fewer than five atomic audit tasks" }
  if (-not $hasMatrixRow) { $reasons += "missing mandatory coverage matrix row" }
  if (-not $hasInvariantPlan) { $reasons += "missing planned invariant mapping" }
  if (-not $hasInvariantEvidence) { $reasons += "no invariant evidence found" }
  if (-not $hasAdversarialEvidence) { $reasons += "no adversarial evidence found" }
  if (-not $hasFuzzEvidence) { $reasons += "no fuzz evidence found" }
  if ($plannedCoverage -lt 95) { $reasons += "planned coverage below 95 percent" }

  $moduleRows += [ordered]@{
    module                    = $module.Module
    task_prefix               = $module.Prefix
    task_count                = $taskCount
    atomic_task_records       = $moduleAtomicTasks.Count
    invalid_atomic_tasks      = @($moduleAtomicTasks | Where-Object { -not $_["valid"] } | ForEach-Object { $_["task_id"] })
    planned_coverage_percent  = $plannedCoverage
    has_matrix_row            = $hasMatrixRow
    has_attack_surface        = $hasAttackSurface
    has_invariant_plan        = $hasInvariantPlan
    has_value_task            = $hasValueTask
    has_invariant_evidence    = $hasInvariantEvidence
    has_adversarial_evidence  = $hasAdversarialEvidence
    has_fuzz_evidence         = $hasFuzzEvidence
    evidence_files            = $evidence.files
    invariant_files           = $evidence.invariant_files
    fuzz_files                = $evidence.fuzz_files
    adversarial_files         = $evidence.adversarial_files
    determinism_files         = $evidence.determinism_files
    safe                      = $safe
    safe_blockers             = $reasons
  }
}

$plannedCoverageTotal = 0.0
foreach ($row in $moduleRows) {
  $plannedCoverageTotal += [double]$row["planned_coverage_percent"]
}
$plannedCoverageAverage = [math]::Round(($plannedCoverageTotal / [double]$moduleRows.Count), 2)
$invalidAtomicTasks = @($atomicTasks | Where-Object { -not $_["valid"] })
$modulesWithInvalidAtomicTasks = @($moduleRows | Where-Object { @($_["invalid_atomic_tasks"]).Count -gt 0 })
$modulesBelowPlan = @($moduleRows | Where-Object { $_["planned_coverage_percent"] -lt 95 -or $_["task_count"] -lt 5 -or $_["atomic_task_records"] -lt 5 -or -not $_["has_matrix_row"] -or -not $_["has_invariant_plan"] -or @($_["invalid_atomic_tasks"]).Count -gt 0 })
$modulesWithoutInvariantEvidence = @($moduleRows | Where-Object { -not $_["has_invariant_evidence"] })
$modulesWithoutFuzzEvidence = @($moduleRows | Where-Object { -not $_["has_fuzz_evidence"] })
$modulesWithoutAdversarialEvidence = @($moduleRows | Where-Object { -not $_["has_adversarial_evidence"] })
$mandatoryInvariantPassRate = 0
$auditPassed = $false
$productionSafe = $false

$runtimeModes = @(
  "stateless fuzzing",
  "stateful multi-block fuzzing",
  "adversarial red-team fuzzing",
  "deterministic replay",
  "stress mode",
  "chaos mode"
)
$simulatorModes = @(
  "in-memory app runner",
  "single-validator localnet",
  "multi-validator localnet",
  "sharding simulator"
)
$fuzzSeeds = @(
  "aexs-auth-replay-0001",
  "aexs-fee-denom-spoof-0002",
  "aexs-tokenfactory-admin-0003",
  "aexs-dex-reserve-desync-0004",
  "aexs-identity-resolver-hijack-0005",
  "aexs-avm-malformed-bytecode-0006",
  "aexs-routing-load-poison-0007",
  "aexs-mesh-replay-0008"
)
$testCommands = @(
  "go test ./...",
  "go vet ./...",
  "buf lint",
  "powershell -NoProfile -ExecutionPolicy Bypass -File tests\scripts\determinism_gate_test.ps1",
  "go test -run '^$' -fuzz <target> -fuzztime <duration>"
)

$stopConditions = @(
  [ordered]@{ id = "first_critical_exploit"; label = "first critical exploit"; enabled = $true; action = "stop immediately and write exploit report" },
  [ordered]@{ id = "max_run_count"; label = "max run count"; enabled = $true; value = 10000; action = "stop after deterministic run budget" },
  [ordered]@{ id = "max_wall_clock_duration"; label = "max wall-clock duration"; enabled = $true; value = "4h"; action = "stop after bounded campaign duration" },
  [ordered]@{ id = "coverage_threshold_reached"; label = "coverage threshold reached"; enabled = $true; value = "95% planned coverage and 100% mandatory invariant execution for selected scope"; action = "stop successful campaign scope" },
  [ordered]@{ id = "deterministic_divergence"; label = "deterministic divergence"; enabled = $true; action = "stop immediately and preserve replay inputs" }
)

$scenarioGenerators = @(
  [ordered]@{
    id                 = "bank_transfer_sequences"
    name               = "random bank transfer sequences"
    flow_covered       = "bank send, multi-send, module transfer, balance query"
    state_transitions  = "account balances, module balances, total supply checks"
    attack_surfaces    = "double spend, malformed address, zero address, overflow, partial multi-send"
    invariant_targets  = "no negative balances; total supply consistency"
    seed_required      = $true
    step_list_required = $true
    status             = "planned_not_executed"
  },
  [ordered]@{
    id                 = "staking_lifecycle_sequences"
    name               = "random staking, delegation, unbonding, redelegation, and reward sequences"
    flow_covered       = "create validator, delegate, redelegate, unbond, reward eligibility"
    state_transitions  = "validator power, delegator shares, staking pool, unbonding records"
    attack_surfaces    = "stake grinding, reward farming, non-naet bond, slash evasion"
    invariant_targets  = "validator set consistency; staking pool consistency"
    seed_required      = $true
    step_list_required = $true
    status             = "planned_not_executed"
  },
  [ordered]@{
    id                 = "validator_slashing_sequences"
    name               = "random validator lifecycle and slashing evidence sequences"
    flow_covered       = "validator create/edit, liveness, downtime, equivocation evidence"
    state_transitions  = "validator status, jail/tombstone state, slash amount, active set removal"
    attack_surfaces    = "stale evidence, malformed proof, duplicate evidence, redelegation evasion"
    invariant_targets  = "deterministic evidence; slash accounting consistency"
    seed_required      = $true
    step_list_required = $true
    status             = "planned_not_executed"
  },
  [ordered]@{
    id                 = "fee_spam_bursts"
    name               = "random fee and spam bursts"
    flow_covered       = "ante fee validation, fee denom policy, fee split, spam admission"
    state_transitions  = "fee collection, burn/treasury/reward accounting, rejected tx state"
    attack_surfaces    = "fee underpayment, non-naet fee, missing fee, fee-griefing spam"
    invariant_targets  = "naet-only fees; exact fee distribution"
    seed_required      = $true
    step_list_required = $true
    status             = "planned_not_executed"
  },
  [ordered]@{
    id                 = "tokenfactory_admin_sequences"
    name               = "random tokenfactory create, mint, burn, and admin sequences"
    flow_covered       = "create denom, mint, burn, change admin, metadata query"
    state_transitions  = "denom state, supply, admin authority, bank metadata"
    attack_surfaces    = "unauthorized mint/burn, admin takeover, native denom spoof, burn-from mismatch"
    invariant_targets  = "supply delta exact; authority consistency"
    seed_required      = $true
    step_list_required = $true
    status             = "planned_not_executed"
  },
  [ordered]@{
    id                 = "dex_pool_swap_sequences"
    name               = "random DEX create-pool, add-liquidity, remove-liquidity, and swap sequences"
    flow_covered       = "create pool, add liquidity, remove liquidity, swap, LP mint/burn"
    state_transitions  = "reserves, LP supply, pair index, module balances"
    attack_surfaces    = "pool drain, reserve desync, LP inflation, slippage bypass, rounding exploit"
    invariant_targets  = "reserves match balances; LP supply matches shares"
    seed_required      = $true
    step_list_required = $true
    status             = "planned_not_executed"
  },
  [ordered]@{
    id                 = "governance_param_sequences"
    name               = "random governance proposal, vote, and parameter update sequences"
    flow_covered       = "proposal, deposit, vote, tally, delayed param execution"
    state_transitions  = "proposal status, vote state, module params, execution queue"
    attack_surfaces    = "governance replay, proposal spam, parameter abuse, upgrade hijack"
    invariant_targets  = "authorized params only; hard bounds preserved"
    seed_required      = $true
    step_list_required = $true
    status             = "planned_not_executed"
  },
  [ordered]@{
    id                 = "identity_domain_sequences"
    name               = "random identity domain registration, auction, renewal, resolver, reverse lookup, and subdomain sequences"
    flow_covered       = "domain auction, assign, renew, expire, resolver update, reverse lookup, subdomain"
    state_transitions  = "domain record, resolver record, reverse mapping, NFT representation"
    attack_surfaces    = "resolver hijack, expired reuse, auction manipulation, subdomain collision"
    invariant_targets  = "domain uniqueness; resolver validity; owner consistency"
    seed_required      = $true
    step_list_required = $true
    status             = "planned_not_executed"
  },
  [ordered]@{
    id                 = "avm_contract_sequences"
    name               = "random AVM deploy, external call, internal call, bounced call, query, and migrate sequences"
    flow_covered       = "deploy, external/internal/bounced call, query, migrate"
    state_transitions  = "contract state, gas use, output messages, migration state"
    attack_surfaces    = "crash input, infinite loop, bad bytecode, sandbox escape, nondeterministic host behavior"
    invariant_targets  = "bounded gas; deterministic state; no panic"
    seed_required      = $true
    step_list_required = $true
    status             = "planned_not_executed"
  },
  [ordered]@{
    id                 = "async_queue_sequences"
    name               = "random async message, queue, delayed execution, bounce, and refund sequences"
    flow_covered       = "async send, enqueue, delayed execution, bounce, refund, receipt"
    state_transitions  = "message state, queue state, receipt state, refund marker"
    attack_surfaces    = "message replay, queue flood, double refund, stale receipt replay"
    invariant_targets  = "deterministic ordering; refund uniqueness; no double spend"
    seed_required      = $true
    step_list_required = $true
    status             = "planned_not_executed"
  },
  [ordered]@{
    id                 = "load_routing_sharding_sequences"
    name               = "random LOAD_SCORE, routing, and shard activation scenarios"
    flow_covered       = "LOAD_SCORE update, route decision, shard activation, routing epoch"
    state_transitions  = "load window, route output, shard assignment, shard activation state"
    attack_surfaces    = "load poisoning, route desync, shard overload targeting, priority fee gaming"
    invariant_targets  = "score bounds; MAX_DELTA; deterministic route; no starvation"
    seed_required      = $true
    step_list_required = $true
    status             = "planned_not_executed"
  }
)

$transactionMutators = @(
  [ordered]@{
    id                = "invalid_signatures"
    name              = "inject invalid signatures"
    mutation_type     = "signature_corruption"
    target_modules    = @("x/auth", "app")
    flow_covered      = "signature verification and signer extraction"
    state_transitions = "auth rejection before account sequence or state mutation"
    attack_surfaces   = "invalid signature, wrong public key, tx malleability"
    invariant_targets = "invalid signer cannot mutate state"
    expected_rejection = "ante handler rejects before message execution"
    seed_required     = $true
    metadata_required = $true
    status            = "planned_not_executed"
  },
  [ordered]@{
    id                = "replay_accepted_tx_bytes"
    name              = "replay already accepted transaction bytes"
    mutation_type     = "replay"
    target_modules    = @("x/auth", "app", "x/bank")
    flow_covered      = "account sequence validation and replay prevention"
    state_transitions = "accepted tx increments sequence once; replay is rejected"
    attack_surfaces   = "same signed bytes, cross-context replay, duplicate delivery"
    invariant_targets = "same signed transaction cannot execute twice"
    expected_rejection = "replay fails sequence or chain-id validation"
    seed_required     = $true
    metadata_required = $true
    status            = "planned_not_executed"
  },
  [ordered]@{
    id                = "nonce_sequence_manipulation"
    name              = "manipulate nonce and sequence values"
    mutation_type     = "sequence_corruption"
    target_modules    = @("x/auth", "app")
    flow_covered      = "nonce and account sequence checks"
    state_transitions = "sequence changes only for accepted txs"
    attack_surfaces   = "future nonce, stale nonce, duplicate sequence, account mismatch"
    invariant_targets = "rejected auth paths do not increment sequence"
    expected_rejection = "invalid sequence rejected before state mutation"
    seed_required     = $true
    metadata_required = $true
    status            = "planned_not_executed"
  },
  [ordered]@{
    id                = "fee_field_corruption"
    name              = "corrupt fee denom and fee amount fields"
    mutation_type     = "fee_corruption"
    target_modules    = @("x/fees", "x/auth", "app")
    flow_covered      = "fee denom, amount, gas limit, and ante fee checks"
    state_transitions = "fee accounting changes only for accepted native-fee txs"
    attack_surfaces   = "malformed fee amount, multi-denom fee, denom spoofing"
    invariant_targets = "naet-only fees; exact fee distribution"
    expected_rejection = "malformed or disallowed fee rejected before message execution"
    seed_required     = $true
    metadata_required = $true
    status            = "planned_not_executed"
  },
  [ordered]@{
    id                = "missing_or_non_naet_fee"
    name              = "inject missing fee and non-naet fee paths"
    mutation_type     = "fee_policy_bypass"
    target_modules    = @("x/fees", "app")
    flow_covered      = "native fee policy and fee bypass rejection"
    state_transitions = "no module state changes after rejected fee path"
    attack_surfaces   = "missing fee, zero fee, non-naet fee, non-FeeTx bypass"
    invariant_targets = "no fee denom other than naet is accepted"
    expected_rejection = "non-native or missing fee rejected by ante policy"
    seed_required     = $true
    metadata_required = $true
    status            = "planned_not_executed"
  },
  [ordered]@{
    id                = "extreme_gas_values"
    name              = "inject extreme gas values"
    mutation_type     = "gas_boundary"
    target_modules    = @("x/fees", "x/vm", "app")
    flow_covered      = "gas limit, gas accounting, simulation mode, VM gas bounds"
    state_transitions = "gas exhaustion rejects or aborts without partial state commit"
    attack_surfaces   = "zero gas, max gas, overflow-like gas, gas griefing"
    invariant_targets = "gas is bounded and deterministic"
    expected_rejection = "invalid gas rejected or exhausted execution rolled back"
    seed_required     = $true
    metadata_required = $true
    status            = "planned_not_executed"
  },
  [ordered]@{
    id                = "malformed_addresses"
    name              = "inject malformed addresses"
    mutation_type     = "address_corruption"
    target_modules    = @("x/auth", "x/bank", "x/tokenfactory", "x/dex", "x/identity")
    flow_covered      = "address parsing, signer checks, recipient/admin/resolver validation"
    state_transitions = "malformed address rejection before state mutation"
    attack_surfaces   = "bad bech32, wrong prefix, truncated bytes, overlong bytes"
    invariant_targets = "invalid signer cannot mutate state; resolver validity"
    expected_rejection = "malformed address rejected by validation path"
    seed_required     = $true
    metadata_required = $true
    status            = "planned_not_executed"
  },
  [ordered]@{
    id                = "zero_address_fields"
    name              = "inject zero address in signer, recipient, admin, authority, resolver, and DEX actor fields"
    mutation_type     = "zero_address"
    target_modules    = @("x/auth", "x/bank", "x/tokenfactory", "x/dex", "x/identity", "x/vm")
    flow_covered      = "zero address validation across value-bearing actor fields"
    state_transitions = "zero address input fails before ownership or balance mutation"
    attack_surfaces   = "zero signer, recipient, admin, authority, resolver, DEX actor"
    invariant_targets = "no zero address ownership or value routing"
    expected_rejection = "zero address rejected by message validation or ante path"
    seed_required     = $true
    metadata_required = $true
    status            = "planned_not_executed"
  },
  [ordered]@{
    id                = "malformed_memo_fields"
    name              = "corrupt memo fields, including invalid UTF-8 and oversized memo payloads"
    mutation_type     = "memo_corruption"
    target_modules    = @("x/memo", "app", "x/indexer")
    flow_covered      = "memo validation, byte length checks, memo indexing"
    state_transitions = "memo metadata written only for accepted txs"
    attack_surfaces   = "invalid UTF-8, oversized memo, binary payload, indexing abuse"
    invariant_targets = "memo cannot alter execution result"
    expected_rejection = "invalid or oversized memo rejected before message execution"
    seed_required     = $true
    metadata_required = $true
    status            = "planned_not_executed"
  },
  [ordered]@{
    id                = "malformed_routing_hints"
    name              = "inject malformed routing hints"
    mutation_type     = "routing_corruption"
    target_modules    = @("x/routing", "x/sharding/sim", "x/execution")
    flow_covered      = "route classification, zone decision, shard assignment"
    state_transitions = "routing hints cannot override deterministic route"
    attack_surfaces   = "invalid zone, invalid shard, forged priority, hot-zone steering"
    invariant_targets = "same tx and state produce same route"
    expected_rejection = "malformed hint ignored or rejected without route divergence"
    seed_required     = $true
    metadata_required = $true
    status            = "planned_not_executed"
  },
  [ordered]@{
    id                = "invalid_domain_resolution"
    name              = "inject invalid domain resolution and expired domain actions"
    mutation_type     = "identity_resolution_corruption"
    target_modules    = @("x/identity", "x/indexer")
    flow_covered      = "domain normalization, expiry, resolver lookup, reverse lookup"
    state_transitions = "domain and resolver state unchanged for invalid actions"
    attack_surfaces   = "expired domain action, resolver spoof, duplicate normalized name"
    invariant_targets = "domain uniqueness; resolver validity; owner consistency"
    expected_rejection = "expired or invalid resolver action rejected before value movement"
    seed_required     = $true
    metadata_required = $true
    status            = "planned_not_executed"
  },
  [ordered]@{
    id                = "fake_cross_zone_messages"
    name              = "inject fake cross-zone messages"
    mutation_type     = "mesh_message_forgery"
    target_modules    = @("x/messaging", "x/mesh", "x/sharding/sim")
    flow_covered      = "cross-zone message proof, receipt, replay marker, finality reference"
    state_transitions = "message/receipt state changes only after valid proof"
    attack_surfaces   = "forged proof, stale finality, wrong destination, duplicate receipt"
    invariant_targets = "cross-zone replay is rejected; no double spend"
    expected_rejection = "fake message rejected by proof or replay validation"
    seed_required     = $true
    metadata_required = $true
    status            = "planned_not_executed"
  },
  [ordered]@{
    id                = "queue_depth_abuse"
    name              = "inject queue depth abuse"
    mutation_type     = "queue_dos"
    target_modules    = @("x/queue", "x/messaging", "x/aetherisvm")
    flow_covered      = "enqueue, delayed execution, depth limit, per-block processing limit"
    state_transitions = "queue depth and sequence counters remain bounded"
    attack_surfaces   = "queue flood, message loop, starvation, duplicate sequence"
    invariant_targets = "queue order; depth bounds; refund uniqueness"
    expected_rejection = "queue overflow or invalid depth rejected without sequence corruption"
    seed_required     = $true
    metadata_required = $true
    status            = "planned_not_executed"
  },
  [ordered]@{
    id                = "oversized_avm_payloads"
    name              = "inject oversized AVM payloads"
    mutation_type     = "vm_size_boundary"
    target_modules    = @("x/vm", "x/aetherisvm")
    flow_covered      = "AVM code size, payload size, query response size, storage size"
    state_transitions = "oversized payload rejected without contract state commit"
    attack_surfaces   = "oversized code, oversized message, oversized query response, storage bloat"
    invariant_targets = "AVM gas and storage are bounded and deterministic"
    expected_rejection = "oversized VM payload rejected by limits"
    seed_required     = $true
    metadata_required = $true
    status            = "planned_not_executed"
  },
  [ordered]@{
    id                = "invalid_avm_entrypoints"
    name              = "inject invalid AVM entrypoint inputs"
    mutation_type     = "vm_entrypoint_corruption"
    target_modules    = @("x/vm", "x/aetherisvm")
    flow_covered      = "deploy, execute, query, migrate, bounced-call entrypoint validation"
    state_transitions = "invalid entrypoint rejects before contract state transition"
    attack_surfaces   = "missing entrypoint, malformed args, invalid migrate, bounced-call spoof"
    invariant_targets = "AVM malformed input does not panic; rejected execution cannot commit state"
    expected_rejection = "invalid entrypoint rejected without panic or partial commit"
    seed_required     = $true
    metadata_required = $true
    status            = "planned_not_executed"
  },
  [ordered]@{
    id                = "malformed_genesis_fragments"
    name              = "inject malformed genesis fragments for simulator startup tests"
    mutation_type     = "genesis_corruption"
    target_modules    = @("app", "x/auth", "x/bank", "x/staking", "x/fees", "x/tokenfactory", "x/dex")
    flow_covered      = "genesis validation, InitGenesis, export/import startup path"
    state_transitions = "invalid genesis rejected before app startup state commit"
    attack_surfaces   = "duplicate accounts, duplicate denoms, invalid params, hidden privileged account"
    invariant_targets = "genesis validation rejects malformed module state"
    expected_rejection = "malformed genesis fragment rejected by validation"
    seed_required     = $true
    metadata_required = $true
    status            = "planned_not_executed"
  },
  [ordered]@{
    id                = "mutation_metadata_recording"
    name              = "record mutation metadata for every scenario"
    mutation_type     = "reporting_integrity"
    target_modules    = @("aexs")
    flow_covered      = "mutation metadata, scenario seed, step list, replay manifest"
    state_transitions = "runtime report records mutation without changing consensus state"
    attack_surfaces   = "missing seed, missing step list, ambiguous reproduction, non-replayable exploit"
    invariant_targets = "every mutation is reproducible by seed and exact steps"
    expected_rejection = "audit record invalid if metadata is incomplete"
    seed_required     = $true
    metadata_required = $true
    status            = "planned_not_executed"
  }
)

$invalidStopConditions = @()
foreach ($condition in $stopConditions) {
  $conditionReasons = @()
  foreach ($field in @("id", "label", "action")) {
    if ([string]::IsNullOrWhiteSpace([string]$condition[$field])) {
      $conditionReasons += "missing $field"
    }
  }
  if ($condition["enabled"] -ne $true) {
    $conditionReasons += "condition disabled"
  }
  $condition["valid"] = $conditionReasons.Count -eq 0
  $condition["invalid_reasons"] = $conditionReasons
  if ($conditionReasons.Count -gt 0) {
    $invalidStopConditions += $condition
  }
}

$invalidScenarioGenerators = @()
foreach ($scenario in $scenarioGenerators) {
  $scenarioReasons = @()
  foreach ($field in @(
      "id",
      "name",
      "flow_covered",
      "state_transitions",
      "attack_surfaces",
      "invariant_targets",
      "status"
    )) {
    if ([string]::IsNullOrWhiteSpace([string]$scenario[$field])) {
      $scenarioReasons += "missing $field"
    }
  }
  if ($scenario["seed_required"] -ne $true) {
    $scenarioReasons += "seed not required"
  }
  if ($scenario["step_list_required"] -ne $true) {
    $scenarioReasons += "step list not required"
  }
  $scenario["valid"] = $scenarioReasons.Count -eq 0
  $scenario["invalid_reasons"] = $scenarioReasons
  if ($scenarioReasons.Count -gt 0) {
    $invalidScenarioGenerators += $scenario
  }
}

$invalidTransactionMutators = @()
foreach ($mutator in $transactionMutators) {
  $mutatorReasons = @()
  foreach ($field in @(
      "id",
      "name",
      "mutation_type",
      "flow_covered",
      "state_transitions",
      "attack_surfaces",
      "invariant_targets",
      "expected_rejection",
      "status"
    )) {
    if ([string]::IsNullOrWhiteSpace([string]$mutator[$field])) {
      $mutatorReasons += "missing $field"
    }
  }
  if (@($mutator["target_modules"]).Count -eq 0) {
    $mutatorReasons += "missing target_modules"
  }
  if ($mutator["seed_required"] -ne $true) {
    $mutatorReasons += "seed not required"
  }
  if ($mutator["metadata_required"] -ne $true) {
    $mutatorReasons += "metadata not required"
  }
  $mutator["valid"] = $mutatorReasons.Count -eq 0
  $mutator["invalid_reasons"] = $mutatorReasons
  if ($mutatorReasons.Count -gt 0) {
    $invalidTransactionMutators += $mutator
  }
}

$campaignSetup = [ordered]@{
  campaign_id       = $campaignId
  git_commit        = $commit
  git_branch        = $branch
  git_dirty_status  = $dirtyStatus
  go_version        = Get-AexsGoVersion
  os                = [System.Runtime.InteropServices.RuntimeInformation]::OSDescription
  test_commands     = $testCommands
  fuzz_seeds        = $fuzzSeeds
  target_modules    = @($moduleRows | ForEach-Object { $_["module"] })
  runtime_modes     = @($runtimeModes | ForEach-Object { [ordered]@{ name = $_; enabled = $true; source = "TO_AUDIT.md" } })
  simulator_modes   = @($simulatorModes | ForEach-Object { [ordered]@{ name = $_; enabled = $true; source = "TO_AUDIT.md" } })
  stop_conditions   = $stopConditions
  setup_complete    = ($invalidStopConditions.Count -eq 0)
}

$scenarioCatalog = [ordered]@{
  campaign_id             = $campaignId
  generator_count         = $scenarioGenerators.Count
  invalid_generator_count = $invalidScenarioGenerators.Count
  invalid_generators      = @($invalidScenarioGenerators | ForEach-Object { $_["id"] })
  seed_policy             = [ordered]@{
    deterministic_seed_required = $true
    step_list_required          = $true
    output_path                 = ".work/aexs/scenarios/"
    replay_requirement          = "Every generated scenario must preserve seed and exact step list."
  }
  generators              = $scenarioGenerators
}

$transactionMutatorCatalog = [ordered]@{
  campaign_id           = $campaignId
  mutator_count         = $transactionMutators.Count
  invalid_mutator_count = $invalidTransactionMutators.Count
  invalid_mutators      = @($invalidTransactionMutators | ForEach-Object { $_["id"] })
  metadata_policy       = [ordered]@{
    mutation_metadata_required = $true
    deterministic_seed_required = $true
    target_module_required     = $true
    expected_rejection_required = $true
    output_path                = ".work/aexs/mutations/"
    replay_requirement         = "Every mutated scenario must preserve mutation id, seed, input delta, expected rejection, and exact step list."
  }
  mutators              = $transactionMutators
}

$summary = [ordered]@{
  campaign_id                         = $campaignId
  output_dir                          = $campaignDir
  source_task_file                    = Get-AexsRelativePath -BasePath $repoRoot -TargetPath $taskPath
  source_pipeline_doc                 = Get-AexsRelativePath -BasePath $repoRoot -TargetPath $pipelinePath
  source_hash                         = $sourceHash
  git_commit                          = $commit
  git_branch                          = $branch
  git_dirty_status                    = $dirtyStatus
  go_version                          = Get-AexsGoVersion
  os                                  = [System.Runtime.InteropServices.RuntimeInformation]::OSDescription
  test_commands                       = $testCommands
  fuzz_seeds                          = $fuzzSeeds
  runtime_modes                       = $runtimeModes
  simulator_modes                     = $simulatorModes
  target_modules                      = @($moduleRows | ForEach-Object { $_["module"] })
  module_count                        = $moduleRows.Count
  stop_conditions                     = $stopConditions
  invalid_stop_condition_count        = $invalidStopConditions.Count
  scenario_generator_count            = $scenarioGenerators.Count
  invalid_scenario_generator_count    = $invalidScenarioGenerators.Count
  scenario_generators                 = @($scenarioGenerators | ForEach-Object { $_["id"] })
  transaction_mutator_count           = $transactionMutators.Count
  invalid_transaction_mutator_count   = $invalidTransactionMutators.Count
  transaction_mutators                = @($transactionMutators | ForEach-Object { $_["id"] })
  atomic_task_count                   = $atomicTasks.Count
  invalid_atomic_task_count           = $invalidAtomicTasks.Count
  invalid_atomic_tasks                = @($invalidAtomicTasks | ForEach-Object { $_["task_id"] })
  modules_with_invalid_atomic_tasks   = @($modulesWithInvalidAtomicTasks | ForEach-Object { $_["module"] })
  planned_coverage_percent            = $plannedCoverageAverage
  modules_below_planned_threshold     = @($modulesBelowPlan | ForEach-Object { $_["module"] })
  modules_without_invariant_evidence  = @($modulesWithoutInvariantEvidence | ForEach-Object { $_["module"] })
  modules_without_fuzz_evidence       = @($modulesWithoutFuzzEvidence | ForEach-Object { $_["module"] })
  modules_without_adversarial_evidence = @($modulesWithoutAdversarialEvidence | ForEach-Object { $_["module"] })
  mandatory_invariant_pass_rate       = $mandatoryInvariantPassRate
  audit_passed                        = $auditPassed
  production_safe                     = $productionSafe
  decision                            = "NOT_SAFE_PRE_CAMPAIGN"
  decision_reason                     = "AEXS structural plan can be audited, but full fuzz campaign execution and 100 percent invariant pass evidence are not recorded yet."
  source_failures                     = $sourceFailures
}

$coveragePath = Join-Path $campaignDir "coverage-matrix.json"
$atomicTasksPath = Join-Path $campaignDir "atomic-tasks.json"
$atomicTasksMarkdownPath = Join-Path $campaignDir "atomic-tasks.md"
$campaignSetupPath = Join-Path $campaignDir "campaign-setup.json"
$scenarioCatalogPath = Join-Path $campaignDir "scenario-generator.json"
$scenarioCatalogMarkdownPath = Join-Path $campaignDir "scenario-generator.md"
$transactionMutatorPath = Join-Path $campaignDir "transaction-mutator.json"
$transactionMutatorMarkdownPath = Join-Path $campaignDir "transaction-mutator.md"
$summaryPath = Join-Path $campaignDir "summary.json"
$resultPath = Join-Path $campaignDir "AUDIT_RESULT.md"
$taskCopyPath = Join-Path $campaignDir "TO_AUDIT.md"
$moduleRows | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath $coveragePath
$atomicTasks | ConvertTo-Json -Depth 10 | Set-Content -LiteralPath $atomicTasksPath
$campaignSetup | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath $campaignSetupPath
$scenarioCatalog | ConvertTo-Json -Depth 10 | Set-Content -LiteralPath $scenarioCatalogPath
$transactionMutatorCatalog | ConvertTo-Json -Depth 10 | Set-Content -LiteralPath $transactionMutatorPath
$summary | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath $summaryPath
Copy-Item -LiteralPath $taskPath -Destination $taskCopyPath -Force

$taskReport = @()
$taskReport += "# AEXS Atomic Audit Tasks"
$taskReport += ""
$taskReport += "- campaign id: $campaignId"
$taskReport += "- task count: $($atomicTasks.Count)"
$taskReport += "- invalid task count: $($invalidAtomicTasks.Count)"
$taskReport += ""
$taskReport += "| Task | Module | Type | Flow | Invariant | Defensive status | Adversarial status | Result | Seed |"
$taskReport += "| --- | --- | --- | --- | --- | --- | --- | --- | --- |"
foreach ($task in $atomicTasks) {
  $flow = ([string]$task["function_or_flow_covered"]).Replace("|", "/")
  $invariant = ([string]$task["invariant_tested"]).Replace("|", "/")
  $seed = [string]$task["reproduction_seed_or_steps"]["seed"]
  $taskReport += "| $($task["task_id"]) | $($task["module"]) | $($task["task_type"]) | $flow | $invariant | $($task["defensive_analysis_result"]["status"]) | $($task["adversarial_simulation_result"]["status"]) | $($task["pass_fail_result"]) | $seed |"
}
$taskReport | Set-Content -LiteralPath $atomicTasksMarkdownPath

$scenarioReport = @()
$scenarioReport += "# AEXS Scenario Generator Catalog"
$scenarioReport += ""
$scenarioReport += "- campaign id: $campaignId"
$scenarioReport += "- generator count: $($scenarioGenerators.Count)"
$scenarioReport += "- invalid generator count: $($invalidScenarioGenerators.Count)"
$scenarioReport += "- seed policy: deterministic seed and exact step list required for every generated scenario"
$scenarioReport += ""
$scenarioReport += "| Scenario | Flow | State transitions | Attack surfaces | Invariants | Status |"
$scenarioReport += "| --- | --- | --- | --- | --- | --- |"
foreach ($scenario in $scenarioGenerators) {
  $flow = ([string]$scenario["flow_covered"]).Replace("|", "/")
  $state = ([string]$scenario["state_transitions"]).Replace("|", "/")
  $attack = ([string]$scenario["attack_surfaces"]).Replace("|", "/")
  $invariant = ([string]$scenario["invariant_targets"]).Replace("|", "/")
  $scenarioReport += "| $($scenario["id"]) | $flow | $state | $attack | $invariant | $($scenario["status"]) |"
}
$scenarioReport | Set-Content -LiteralPath $scenarioCatalogMarkdownPath

$mutatorReport = @()
$mutatorReport += "# AEXS Transaction Mutator Catalog"
$mutatorReport += ""
$mutatorReport += "- campaign id: $campaignId"
$mutatorReport += "- mutator count: $($transactionMutators.Count)"
$mutatorReport += "- invalid mutator count: $($invalidTransactionMutators.Count)"
$mutatorReport += "- metadata policy: mutation id, deterministic seed, target modules, expected rejection, and replay steps required"
$mutatorReport += ""
$mutatorReport += "| Mutator | Type | Targets | Flow | Attack surfaces | Invariants | Expected rejection | Status |"
$mutatorReport += "| --- | --- | --- | --- | --- | --- | --- | --- |"
foreach ($mutator in $transactionMutators) {
  $targets = (@($mutator["target_modules"]) -join ", ").Replace("|", "/")
  $flow = ([string]$mutator["flow_covered"]).Replace("|", "/")
  $attack = ([string]$mutator["attack_surfaces"]).Replace("|", "/")
  $invariant = ([string]$mutator["invariant_targets"]).Replace("|", "/")
  $expected = ([string]$mutator["expected_rejection"]).Replace("|", "/")
  $mutatorReport += "| $($mutator["id"]) | $($mutator["mutation_type"]) | $targets | $flow | $attack | $invariant | $expected | $($mutator["status"]) |"
}
$mutatorReport | Set-Content -LiteralPath $transactionMutatorMarkdownPath

$report = @()
$report += "# AEXS Audit Result"
$report += ""
$report += "- campaign id: $campaignId"
$report += "- git commit: $commit"
$report += "- branch: $branch"
$report += "- output dir: $campaignDir"
$report += "- decision: NOT_SAFE_PRE_CAMPAIGN"
$report += "- planned coverage: $plannedCoverageAverage%"
$report += "- mandatory invariant pass rate: $mandatoryInvariantPassRate%"
$report += "- atomic task records: $($atomicTasks.Count)"
$report += "- invalid atomic task records: $($invalidAtomicTasks.Count)"
$report += "- stop conditions: $($stopConditions.Count)"
$report += "- scenario generators: $($scenarioGenerators.Count)"
$report += "- invalid scenario generators: $($invalidScenarioGenerators.Count)"
$report += "- transaction mutators: $($transactionMutators.Count)"
$report += "- invalid transaction mutators: $($invalidTransactionMutators.Count)"
$report += ""
$report += "## Gate Decision"
$report += ""
$report += "The audit is not passed yet. This preflight validates that the campaign plan and matrix are machine-checkable, but production-safe status requires an executed fuzz/invariant campaign with 100% mandatory invariant pass rate and no untriaged Critical or High exploit."
$report += ""
$report += "## Coverage Gaps"
$report += ""
$report += "- modules below 95% planned coverage: $(@($modulesBelowPlan | ForEach-Object { $_["module"] }) -join ', ')"
$report += "- modules without invariant evidence: $(@($modulesWithoutInvariantEvidence | ForEach-Object { $_["module"] }) -join ', ')"
$report += "- modules without fuzz evidence: $(@($modulesWithoutFuzzEvidence | ForEach-Object { $_["module"] }) -join ', ')"
$report += "- modules without adversarial evidence: $(@($modulesWithoutAdversarialEvidence | ForEach-Object { $_["module"] }) -join ', ')"
$report += "- modules with invalid atomic tasks: $(@($modulesWithInvalidAtomicTasks | ForEach-Object { $_["module"] }) -join ', ')"
$report += ""
$report += "## Module Matrix"
$report += ""
$report += "| Module | Tasks | Atomic records | Planned coverage | Invariant evidence | Fuzz evidence | Adversarial evidence | Safe |"
$report += "| --- | ---: | ---: | ---: | --- | --- | --- | --- |"
foreach ($row in $moduleRows) {
  $report += "| $($row["module"]) | $($row["task_count"]) | $($row["atomic_task_records"]) | $($row["planned_coverage_percent"])% | $($row["has_invariant_evidence"]) | $($row["has_fuzz_evidence"]) | $($row["has_adversarial_evidence"]) | $($row["safe"]) |"
}
$report += ""
$report += "## Required Next Step"
$report += ""
$report += "Run the AEXS fuzzing and invariant campaign, write generated scenarios, minimized exploits, state diffs, and final results under `.work/aexs/`, then update this result with executed invariant pass rates and triage status."
$report | Set-Content -LiteralPath $resultPath

if ($sourceFailures.Count -gt 0) {
  throw "AEXS audit source validation failed: $($sourceFailures -join '; ')"
}
if ($invalidStopConditions.Count -gt 0) {
  throw "AEXS stop condition validation failed for condition(s): $(@($invalidStopConditions | ForEach-Object { $_["id"] }) -join ', ')"
}
if ($invalidScenarioGenerators.Count -gt 0) {
  throw "AEXS scenario generator validation failed for generator(s): $(@($invalidScenarioGenerators | ForEach-Object { $_["id"] }) -join ', ')"
}
if ($invalidTransactionMutators.Count -gt 0) {
  throw "AEXS transaction mutator validation failed for mutator(s): $(@($invalidTransactionMutators | ForEach-Object { $_["id"] }) -join ', ')"
}
if ($invalidAtomicTasks.Count -gt 0) {
  throw "AEXS atomic task validation failed for task(s): $(@($invalidAtomicTasks | ForEach-Object { $_["task_id"] }) -join ', ')"
}
if ($modulesBelowPlan.Count -gt 0) {
  throw "AEXS planned coverage gate failed for module(s): $(@($modulesBelowPlan | ForEach-Object { $_["module"] }) -join ', ')"
}
if ($EnforceSafe -and -not $auditPassed) {
  throw "AEXS audit is not production safe; see $resultPath"
}

if ($Json) {
  $summary | ConvertTo-Json -Depth 8
} else {
  Write-Host "AEXS audit preflight complete"
  Write-Host "Campaign: $campaignId"
  Write-Host "Planned coverage: $plannedCoverageAverage%"
  Write-Host "Decision: NOT_SAFE_PRE_CAMPAIGN"
  Write-Host "Report: $resultPath"
}
