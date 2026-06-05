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
    "GOV-01" = [ordered]@{
      flow                = "proposal creation, voting, tallying, parameter update, delayed execution"
      state               = "accepted governance proposal updates proposal status, votes, tally result, params, and delayed execution queue deterministically"
      attack              = "valid governance lifecycle baseline plus unauthorized proposer or voter control sample"
      invariant           = "governance lifecycle executes only accepted proposals after deterministic voting and delay rules"
      expected_behavior   = "valid proposal, vote, tally, and delayed execution update authorized governance state exactly once"
      expected_events     = "governance events match proposal status, vote, tally, and executed parameter deltas"
      expected_error_path = "unauthorized proposer or voter control sample is rejected before proposal or vote state mutation"
      mutation_inputs     = "valid proposal, valid vote, valid tally, valid delayed execution, unauthorized proposer, unauthorized voter"
      expected_rejection  = "unauthorized governance lifecycle variants must fail without proposal, vote, tally, or params mutation"
    }
    "GOV-02" = [ordered]@{
      flow                = "proposal spam, malformed params, expired voting period, zero deposit, invalid authority"
      state               = "invalid governance edge cases leave proposals, deposits, votes, params, and execution queues unchanged"
      attack              = "proposal spam, malformed params, expired voting period, zero deposit, invalid authority"
      invariant           = "governance accepts only valid authority, deposits, voting windows, and bounded params"
      expected_behavior   = "valid governance boundaries execute deterministically; invalid boundaries reject before state mutation"
      expected_events     = "accepted boundary proposals emit accurate governance events; rejected edge cases emit no success events"
      expected_error_path = "governance validation rejects malformed params, expired periods, zero deposits, or invalid authority"
      mutation_inputs     = "many proposals in one block, malformed param value, expired voting period vote, zero deposit, invalid authority address"
      expected_rejection  = "invalid governance edge cases must not alter proposal status, deposits, votes, params, or execution queues"
    }
    "GOV-03" = [ordered]@{
      flow                = "governance replay, proposal front-running, emergency parameter abuse, upgrade hijack"
      state               = "adversarial governance attempts cannot replay execution, bypass delays, abuse emergency params, or hijack upgrades"
      attack              = "governance replay, proposal front-running, emergency parameter abuse, upgrade hijack"
      invariant           = "governance execution is single-use, authority-bound, delay-bound, and limited to authorized handlers"
      expected_behavior   = "adversarial governance mutations fail or resolve deterministically without unauthorized params or upgrades"
      expected_events     = "failed governance attacks emit no misleading execution, authority, or upgrade success events"
      expected_error_path = "governance handler rejects replay, front-running, emergency abuse, or upgrade hijack before commit"
      mutation_inputs     = "replayed executed proposal, duplicate proposal id, front-run param proposal, emergency param outside bounds, forged upgrade authority"
      expected_rejection  = "governance attacks must not execute twice, bypass delay, change unauthorized params, or schedule unauthorized upgrades"
    }
    "GOV-04" = [ordered]@{
      flow                = "authorized params update and rejected proposal state integrity"
      state               = "accepted proposals update only authorized params and rejected proposals leave state unchanged"
      attack              = "state drift attempt through mixed accepted, rejected, expired, and unauthorized governance proposals"
      invariant           = "accepted proposals update only authorized params; rejected proposals preserve pre-proposal state"
      expected_behavior   = "governance state integrity holds across proposal execution, rejection, expiry, and export/import"
      expected_events     = "governance events reconcile to authorized param changes and no-op rejected proposals"
      expected_error_path = "failed proposals preserve pre-failure params, module state, deposits, votes, and delayed execution queue"
      mutation_inputs     = "accepted fee param update followed by rejected staking param update, expired proposal, unauthorized param route, export/import after proposals"
      expected_rejection  = "rejected governance operations must preserve params, deposits, votes, and delayed execution state"
    }
    "GOV-05" = [ordered]@{
      flow                = "governance economic abuse around fee, inflation, staking, and burn params"
      state               = "governance cannot set fee, inflation, staking, or burn params outside hard protocol bounds"
      attack              = "unsafe fee param, unsafe inflation param, unsafe staking param, unsafe burn param"
      invariant           = "governance-controlled economic params remain inside hard protocol bounds"
      expected_behavior   = "governance economic rules preserve hard bounds and delayed execution requirements"
      expected_events     = "no economic param success event appears for rejected out-of-bounds governance proposals"
      expected_error_path = "economic param abuse rejects before params, mint, burn, fee, or staking state mutation"
      mutation_inputs     = "negative fee param, unbounded inflation, zero unbonding time, slash-free staking params, burn rate above max"
      expected_rejection  = "governance economic abuse must not set fee, inflation, staking, or burn params outside hard protocol bounds"
    }
    "DIST-01" = [ordered]@{
      flow                = "validator commission, delegator rewards, community pool accounting, reward withdrawal"
      state               = "validator commission, delegator rewards, community pool, and withdrawal balances update deterministically"
      attack              = "valid distribution lifecycle baseline plus unauthorized withdrawal control sample"
      invariant           = "distribution rewards are paid only through authorized reward and commission paths"
      expected_behavior   = "valid commission, reward accrual, community pool accounting, and withdrawals update balances exactly once"
      expected_events     = "distribution events match commission, reward, community pool, and withdrawal deltas"
      expected_error_path = "unauthorized reward withdrawal control sample is rejected before distribution or bank state mutation"
      mutation_inputs     = "valid validator commission, valid delegator reward accrual, valid community pool accounting, valid withdraw, unauthorized withdraw"
      expected_rejection  = "unauthorized distribution lifecycle variants must fail without reward, commission, or pool mutation"
    }
    "DIST-02" = [ordered]@{
      flow                = "tiny rewards, rounding remainders, jailed validators, zero delegations, repeated withdrawals"
      state               = "distribution edge cases preserve deterministic rewards, remainders, commission, and module balances"
      attack              = "tiny reward rounding, jailed validator reward claim, zero delegation claim, repeated withdrawal"
      invariant           = "distribution rounding and edge rewards cannot create extra funds or negative outstanding rewards"
      expected_behavior   = "valid distribution boundaries execute deterministically; invalid boundaries reject before balance mutation"
      expected_events     = "accepted boundary withdrawals emit accurate reward events; rejected edge cases emit no success events"
      expected_error_path = "distribution validation rejects invalid jailed, zero-delegation, or repeated withdrawal paths"
      mutation_inputs     = "one-atto reward, rounding remainder, jailed validator, zero delegation, repeated withdrawal in same sequence"
      expected_rejection  = "invalid distribution edge cases must not alter rewards, commission, community pool, or module balances"
    }
    "DIST-03" = [ordered]@{
      flow                = "reward double claim, commission bypass, reward inflation, module balance desync"
      state               = "adversarial distribution attempts cannot double claim, bypass commission, inflate rewards, or desync module balances"
      attack              = "reward double claim, commission bypass, reward inflation, module balance desync"
      invariant           = "distribution cannot pay more than accrued rewards or desynchronize bank/module accounting"
      expected_behavior   = "adversarial distribution mutations fail or resolve deterministically without extra rewards"
      expected_events     = "failed distribution attacks emit no misleading withdraw, commission, or reward success events"
      expected_error_path = "distribution keeper or bank send path rejects invalid reward manipulation before commit"
      mutation_inputs     = "duplicate withdraw, commission address spoof, inflated outstanding rewards, module account balance mismatch"
      expected_rejection  = "distribution attacks must not double pay rewards, bypass commission, inflate rewards, or desync module balances"
    }
    "DIST-04" = [ordered]@{
      flow                = "outstanding rewards, module balances, supply accounting determinism"
      state               = "outstanding rewards, module balances, and supply accounting reconcile after accepted and rejected distribution operations"
      attack              = "state drift attempt through mixed accepted and rejected reward withdrawals and commission claims"
      invariant           = "outstanding rewards, module balances, and supply accounting stay deterministic"
      expected_behavior   = "distribution state integrity holds across withdrawal, commission, community pool, and export/import sequences"
      expected_events     = "distribution events reconcile to final reward, commission, pool, and balance deltas"
      expected_error_path = "failed distribution operations preserve pre-failure outstanding rewards and module balance snapshot"
      mutation_inputs     = "accepted withdraw followed by failed duplicate withdraw, accepted commission followed by failed spoof, export/import after rewards"
      expected_rejection  = "rejected distribution operations must preserve outstanding reward, module balance, and supply consistency"
    }
    "DIST-05" = [ordered]@{
      flow                = "distribution economic abuse around unauthorized mint and treasury or community-pool leakage"
      state               = "distribution cannot mint outside authorized reward path or leak treasury/community-pool funds"
      attack              = "unauthorized reward mint, treasury leak, community pool leak, reward path spoof"
      invariant           = "distribution cannot mint outside authorized reward path or leak treasury/community-pool funds"
      expected_behavior   = "distribution economic rules preserve module account authorization, reward limits, and community-pool accounting"
      expected_events     = "no mint, treasury leak, or community-pool leak event appears for rejected economic abuse paths"
      expected_error_path = "economic abuse rejects before mint, bank send, treasury, community pool, or reward state mutation"
      mutation_inputs     = "mint-shaped reward withdrawal, unauthorized treasury transfer, community pool drain, forged reward module account"
      expected_rejection  = "distribution economic abuse must not mint unauthorized rewards or leak treasury/community-pool funds"
    }
    "FEES-01" = [ordered]@{
      flow                = "valid naet fee collection, minimum fee enforcement, split accounting, params query"
      state               = "accepted fee path collects naet, applies minimum fee, records split accounting, and exposes stable params"
      attack              = "valid fee baseline plus unauthorized fee collector control sample"
      invariant           = "valid fee collection accepts only naet and preserves configured split accounting"
      expected_behavior   = "valid naet fee tx passes ante validation, collects fees exactly once, and returns deterministic params query output"
      expected_events     = "fee collection events match collected amount, denom, and split accounting deltas"
      expected_error_path = "unauthorized fee collector control sample is rejected before fee accounting mutation"
      mutation_inputs     = "valid naet fee, valid minimum fee, valid split accounting, valid params query, unauthorized collector control"
      expected_rejection  = "unauthorized fee collection variants must fail without fee accounting or message execution"
    }
    "FEES-02" = [ordered]@{
      flow                = "missing fee, zero fee, multi-denom fee, malformed fee, max fee, simulation mode"
      state               = "fee edge cases either pass explicit simulation rules or leave fee accounting and message state unchanged"
      attack              = "missing fee, zero fee, multi-denom fee, malformed fee, max fee, simulation mode abuse"
      invariant           = "fee validation is naet-only, bounded, deterministic, and simulation-mode explicit"
      expected_behavior   = "valid fee boundaries execute deterministically; invalid fee boundaries reject before message execution"
      expected_events     = "accepted boundary fee events match accounting deltas; rejected fee edge cases emit no success events"
      expected_error_path = "ante fee decorator rejects missing, zero, multi-denom, malformed, or unsafe max-fee input unless explicit simulation mode applies"
      mutation_inputs     = "missing fee, zero naet fee, multi-denom fee with naet plus factory denom, malformed fee amount, max int fee, simulation mode tx"
      expected_rejection  = "invalid fee edge cases must not execute messages or mutate fee accounting"
    }
    "FEES-03" = [ordered]@{
      flow                = "fee underpayment, fee denom spoofing, fee-griefing spam, non-FeeTx bypass"
      state               = "adversarial fee attempts cannot underpay, spoof denom, consume priority, or bypass FeeTx validation"
      attack              = "fee underpayment, fee denom spoofing, fee-griefing spam, non-FeeTx bypass"
      invariant           = "fee policy cannot be bypassed by tx shape, denom spoofing, spam, or missing FeeTx interface"
      expected_behavior   = "adversarial fee mutations fail deterministically before message execution or priority admission"
      expected_events     = "failed fee attacks emit no fee acceptance, priority success, or message execution events"
      expected_error_path = "ante fee validation rejects underpaid, spoofed, spam-shaped, or non-FeeTx transactions before commit"
      mutation_inputs     = "below-min fee, factory denom named like naet, many low-fee spam txs, tx object without FeeTx semantics"
      expected_rejection  = "fee attacks must not underpay, spoof naet, bypass FeeTx checks, or gain accepted priority"
    }
    "FEES-04" = [ordered]@{
      flow                = "failed ante check state integrity for message execution and fee accounting"
      state               = "failed ante checks do not execute messages, increment business state, or corrupt fee accounting"
      attack              = "state corruption attempt through failed fee ante path"
      invariant           = "failed fee ante checks cannot execute messages or mutate fee accounting"
      expected_behavior   = "fee state integrity holds across rejected ante paths and export/import"
      expected_events     = "no module message, fee split, burn, treasury, or validator reward success event after failed ante"
      expected_error_path = "failed fee ante path returns before message server execution and before persistent fee accounting"
      mutation_inputs     = "bad fee followed by state-changing msg, non-naet fee with tokenfactory mint msg, malformed fee with bank send"
      expected_rejection  = "failed ante fee checks must preserve message state, account state, and fee accounting snapshots"
    }
    "FEES-05" = [ordered]@{
      flow                = "fee split, burn, treasury, validator reward accounting under tx shape and load state"
      state               = "fee split, burn, treasury, and validator reward accounting cannot be manipulated by tx shape or load state"
      attack              = "fee split manipulation, burn bypass, treasury reroute, validator reward skew through tx shape or load state"
      invariant           = "fee split, burn, treasury, and validator reward accounting are deterministic and bounded"
      expected_behavior   = "fee economic rules preserve configured accounting regardless of valid tx shape or deterministic load state"
      expected_events     = "fee distribution events reconcile exactly to collected naet and configured split"
      expected_error_path = "economic fee abuse rejects before burn, treasury, reward, or split accounting mutation"
      mutation_inputs     = "multi-msg tx fee shape, high-load fee path, low-load fee path, malformed split params, forged validator reward target"
      expected_rejection  = "fee economic abuse must not manipulate split, burn, treasury, or validator reward accounting"
    }
    "TF-01" = [ordered]@{
      flow                = "create denom, mint, burn, change admin, metadata query"
      state               = "denom records, admin metadata, bank supply, and metadata query output update deterministically"
      attack              = "valid tokenfactory lifecycle baseline plus unauthorized admin control sample"
      invariant           = "tokenfactory lifecycle mutates denom state only under current admin authority"
      expected_behavior   = "valid create, mint, burn, admin change, and metadata query update denom state and supply exactly once"
      expected_events     = "tokenfactory events match denom creation, mint, burn, admin change, and metadata deltas"
      expected_error_path = "unauthorized admin control sample is rejected before denom, supply, or metadata mutation"
      mutation_inputs     = "valid create denom, valid mint, valid burn, valid change admin, valid metadata query, unauthorized admin control"
      expected_rejection  = "unauthorized tokenfactory lifecycle variants must fail without denom, supply, or admin mutation"
    }
    "TF-02" = [ordered]@{
      flow                = "invalid subdenom, duplicate denom, zero admin, native denom spoof, max metadata size"
      state               = "invalid tokenfactory edge cases leave denom records, admin metadata, supply, and bank metadata unchanged"
      attack              = "invalid subdenom, duplicate denom, zero admin, native denom spoof, oversized metadata"
      invariant           = "tokenfactory accepts only canonical denoms, non-zero admins, unique denoms, and bounded metadata"
      expected_behavior   = "valid tokenfactory boundaries execute deterministically; invalid boundaries reject before state mutation"
      expected_events     = "accepted boundary tokenfactory events match state deltas; rejected edge cases emit no success events"
      expected_error_path = "tokenfactory validation rejects invalid subdenom, duplicate denom, zero admin, native spoof, or oversized metadata"
      mutation_inputs     = "empty subdenom, invalid subdenom chars, duplicate denom, zero admin, naet-like denom, max metadata size plus one"
      expected_rejection  = "invalid tokenfactory edge cases must not alter denom records, supply, admin metadata, or native metadata"
    }
    "TF-03" = [ordered]@{
      flow                = "unauthorized mint, unauthorized burn, admin takeover, metadata spoofing, burn-from mismatch"
      state               = "adversarial tokenfactory attempts cannot mint, burn, take admin, spoof metadata, or burn from another account"
      attack              = "unauthorized mint, unauthorized burn, admin takeover, metadata spoofing, burn-from mismatch"
      invariant           = "tokenfactory authority, metadata, and burn source checks cannot be bypassed"
      expected_behavior   = "adversarial tokenfactory mutations fail deterministically before supply or authority mutation"
      expected_events     = "failed tokenfactory attacks emit no mint, burn, metadata, or admin success events"
      expected_error_path = "tokenfactory msg server rejects unauthorized mint/burn/admin/metadata/burn-from paths before bank movement"
      mutation_inputs     = "non-admin mint, non-admin burn, forged change admin, spoofed metadata update, burn from different holder"
      expected_rejection  = "tokenfactory attacks must not mint, burn, change admin, spoof metadata, or burn from mismatched accounts"
    }
    "TF-04" = [ordered]@{
      flow                = "supply delta exactness and authority metadata consistency"
      state               = "supply changes exactly by minted or burned amount and authority metadata remains consistent"
      attack              = "state drift attempt through mixed accepted and rejected tokenfactory mint, burn, and admin operations"
      invariant           = "tokenfactory supply delta is exact and authority metadata remains consistent"
      expected_behavior   = "tokenfactory state integrity holds across lifecycle sequences and export/import"
      expected_events     = "tokenfactory events reconcile to final supply, admin, metadata, and bank balance deltas"
      expected_error_path = "failed tokenfactory operations preserve pre-failure supply, admin, metadata, and balances"
      mutation_inputs     = "accepted mint followed by failed mint, accepted burn followed by failed burn-from mismatch, change admin then old-admin mint, export/import after supply changes"
      expected_rejection  = "rejected tokenfactory operations must preserve supply delta exactness and authority metadata consistency"
    }
    "TF-05" = [ordered]@{
      flow                = "tokenfactory economic abuse around protocol fees, AET spoofing, and native supply inflation"
      state               = "tokenfactory assets cannot pay protocol fees, spoof AET, or inflate native supply"
      attack              = "factory asset fee payment, AET spoof, native supply inflation, native metadata collision"
      invariant           = "tokenfactory assets cannot pay protocol fees, spoof AET, or inflate native supply"
      expected_behavior   = "tokenfactory economic rules keep factory denoms separate from native fee and native supply authority"
      expected_events     = "no protocol fee acceptance, native mint, or native metadata spoof event appears for rejected tokenfactory abuse paths"
      expected_error_path = "economic abuse rejects before fee acceptance, native supply mutation, or native metadata mutation"
      mutation_inputs     = "factory denom as fee, factory denom named AET, factory denom named naet, mint shaped as native supply, native metadata spoof"
      expected_rejection  = "tokenfactory economic abuse must not pay protocol fees, spoof AET, or inflate native supply"
    }
    "DEX-01" = [ordered]@{
      flow                = "pool creation, add liquidity, remove liquidity, swap, LP mint, LP burn"
      state               = "pool records, reserves, LP supply, shares, and module balances update deterministically"
      attack              = "valid DEX lifecycle baseline plus unauthorized pool module account control sample"
      invariant           = "DEX lifecycle preserves reserves, LP supply, shares, and module bank balance consistency"
      expected_behavior   = "valid pool creation, liquidity changes, swaps, LP mint, and LP burn update AMM state exactly once"
      expected_events     = "DEX events match pool, reserve, LP supply, share, and swap deltas"
      expected_error_path = "unauthorized pool module account control sample is rejected before pool, LP, or bank state mutation"
      mutation_inputs     = "valid create pool, valid add liquidity, valid remove liquidity, valid swap, valid LP mint, valid LP burn, unauthorized pool module movement"
      expected_rejection  = "unauthorized DEX lifecycle variants must fail without pool, reserve, LP, or bank mutation"
    }
    "DEX-02" = [ordered]@{
      flow                = "duplicate pair, tiny reserves, zero liquidity, same denom pair, invalid pool id, max amount"
      state               = "invalid DEX edge cases leave pool records, reserves, LP supply, shares, and balances unchanged"
      attack              = "duplicate pair, tiny reserve, zero liquidity, same denom pair, invalid pool id, max amount boundary"
      invariant           = "DEX accepts only canonical pairs, positive liquidity, valid pool ids, and bounded amounts"
      expected_behavior   = "valid DEX boundaries execute deterministically; invalid boundaries reject before pool or bank mutation"
      expected_events     = "accepted boundary DEX events match accounting deltas; rejected edge cases emit no success events"
      expected_error_path = "DEX validation rejects duplicate pairs, tiny invalid reserves, zero liquidity, same denom pairs, invalid pool ids, or unsafe max amounts"
      mutation_inputs     = "duplicate pair, one-unit reserve, zero liquidity, same denom pair, invalid pool id, max amount plus one"
      expected_rejection  = "invalid DEX edge cases must not alter pool records, reserves, LP supply, shares, or balances"
    }
    "DEX-03" = [ordered]@{
      flow                = "pool drain, LP inflation, reserve desync, failed bank movement partial update, slippage bypass"
      state               = "adversarial DEX attempts cannot drain pools, inflate LP, desync reserves, partially move funds, or bypass slippage"
      attack              = "pool drain, LP inflation, reserve desync, failed bank movement partial update, slippage bypass"
      invariant           = "DEX bank movements are atomic and AMM reserves, LP supply, and slippage constraints cannot be bypassed"
      expected_behavior   = "adversarial DEX mutations fail deterministically before pool or bank state corruption"
      expected_events     = "failed DEX attacks emit no misleading pool, LP, reserve, or swap success events"
      expected_error_path = "DEX keeper or bank send path rejects pool drain, LP inflation, reserve desync, partial bank movement, or slippage bypass before commit"
      mutation_inputs     = "oversized swap drain, forged LP mint, manual reserve mismatch, bank send failure after reserve update attempt, min-out slippage bypass"
      expected_rejection  = "DEX attacks must not drain pools, inflate LP supply, desync reserves, partially commit bank movement, or bypass slippage"
    }
    "DEX-04" = [ordered]@{
      flow                = "reserves match module balances, LP supply matches shares, failed operations leave pool state unchanged"
      state               = "reserves, module balances, LP supply, shares, and failed operation snapshots reconcile deterministically"
      attack              = "state drift attempt through mixed accepted and rejected DEX pool, liquidity, and swap operations"
      invariant           = "reserves match module balances and LP supply matches shares"
      expected_behavior   = "DEX state integrity holds across pool, liquidity, swap, failure, and export/import sequences"
      expected_events     = "DEX events reconcile to final reserve, balance, LP supply, and share deltas"
      expected_error_path = "failed DEX operations preserve pre-failure pool, reserve, LP supply, share, and module balance snapshots"
      mutation_inputs     = "accepted add liquidity followed by failed swap, accepted swap followed by failed remove liquidity, accepted create pool followed by duplicate create, export/import after pool updates"
      expected_rejection  = "rejected DEX operations must preserve reserve/module balance and LP/share consistency"
    }
    "DEX-05" = [ordered]@{
      flow                = "constant-product and fee-adjusted swap math under rounding, ordering, and malformed denom input"
      state               = "constant-product, fee-adjusted reserves, LP supply, and pool balances cannot be exploited by rounding, ordering, or malformed denoms"
      attack              = "rounding leak, ordering manipulation, malformed denom pair, fee-adjusted swap math abuse"
      invariant           = "constant-product and fee-adjusted swap math cannot be exploited through rounding, ordering, or malformed denoms"
      expected_behavior   = "DEX economic math preserves protocol-favorable rounding, canonical ordering, and denom validation"
      expected_events     = "no profitable rounding, ordering, or malformed-denom success event appears for rejected DEX economic abuse paths"
      expected_error_path = "DEX economic abuse rejects before swap, LP, reserve, or bank state mutation"
      mutation_inputs     = "tiny swap rounding loop, reversed denom order, malformed denom, fee bypass amount, repeated swap ordering sequence"
      expected_rejection  = "DEX economic abuse must not violate constant-product math, bypass fees, leak rounding value, or accept malformed denoms"
    }
    "ID-01" = [ordered]@{
      flow                = "domain auction, assignment, renewal, expiry, resolver update, reverse lookup, subdomain flow"
      state               = "domain records, ownership, expiry, resolver records, reverse lookup, subdomains, and NFT representation update deterministically"
      attack              = "valid identity lifecycle baseline plus unauthorized owner control sample"
      invariant           = "identity lifecycle preserves domain uniqueness, owner authority, resolver validity, and NFT ownership consistency"
      expected_behavior   = "valid auction, assignment, renewal, expiry, resolver update, reverse lookup, and subdomain operations update identity state exactly once"
      expected_events     = "identity events match domain, owner, resolver, reverse, subdomain, and NFT deltas"
      expected_error_path = "unauthorized owner control sample is rejected before domain, resolver, reverse, subdomain, or NFT state mutation"
      mutation_inputs     = "valid domain auction, valid assignment, valid renewal, valid expiry, valid resolver update, valid reverse lookup, valid subdomain, unauthorized owner control"
      expected_rejection  = "unauthorized identity lifecycle variants must fail without domain, resolver, reverse, subdomain, or NFT mutation"
    }
    "ID-02" = [ordered]@{
      flow                = "invalid name, duplicate name, expired domain, missing resolver, zero resolver, max metadata"
      state               = "invalid identity edge cases leave domain records, resolver records, reverse records, expiry, metadata, and NFT state unchanged"
      attack              = "invalid name, duplicate name, expired domain resolution, missing resolver, zero resolver, oversized metadata"
      invariant           = "identity accepts only canonical unique names, active domains, valid resolvers, and bounded metadata"
      expected_behavior   = "valid identity boundaries execute deterministically; invalid boundaries reject before registry mutation"
      expected_events     = "accepted boundary identity events match state deltas; rejected edge cases emit no success events"
      expected_error_path = "identity validation rejects invalid names, duplicate names, expired domains, missing resolvers, zero resolvers, or oversized metadata"
      mutation_inputs     = "empty name, mixed-case duplicate name, expired domain action, missing resolver query, zero resolver address, max metadata plus one"
      expected_rejection  = "invalid identity edge cases must not alter domain records, resolver records, reverse records, expiry, metadata, or NFT state"
    }
    "ID-03" = [ordered]@{
      flow                = "domain hijack, auction manipulation, resolver overwrite, reverse lookup poisoning, subdomain collision"
      state               = "adversarial identity attempts cannot hijack domains, manipulate auctions, overwrite resolvers, poison reverse lookup, or collide subdomains"
      attack              = "domain hijack, auction manipulation, resolver overwrite, reverse lookup poisoning, subdomain collision"
      invariant           = "identity ownership, auction ordering, resolver authority, reverse lookup authorization, and subdomain uniqueness cannot be bypassed"
      expected_behavior   = "adversarial identity mutations fail deterministically before ownership or resolver state corruption"
      expected_events     = "failed identity attacks emit no misleading domain transfer, resolver update, reverse, auction, or subdomain success events"
      expected_error_path = "identity state transition rejects hijack, auction manipulation, resolver overwrite, reverse poisoning, or subdomain collision before commit"
      mutation_inputs     = "non-owner transfer, bid reveal manipulation, non-owner resolver overwrite, reverse lookup for another address, duplicate subdomain"
      expected_rejection  = "identity attacks must not hijack domains, manipulate auctions, overwrite resolvers, poison reverse lookup, or create subdomain collisions"
    }
    "ID-04" = [ordered]@{
      flow                = "registry owner, resolver record, expiry, NFT representation consistency"
      state               = "registry owner, resolver record, expiry, reverse records, subdomains, and NFT representation do not diverge"
      attack              = "state drift attempt through mixed accepted and rejected identity ownership, resolver, expiry, and NFT operations"
      invariant           = "registry owner, resolver record, expiry, and NFT representation do not diverge"
      expected_behavior   = "identity state integrity holds across lifecycle, resolver, subdomain, transfer, and export/import sequences"
      expected_events     = "identity events reconcile to final owner, resolver, expiry, reverse, subdomain, and NFT deltas"
      expected_error_path = "failed identity operations preserve pre-failure owner, resolver, expiry, reverse, subdomain, metadata, and NFT snapshots"
      mutation_inputs     = "accepted assignment followed by failed resolver update, accepted renewal followed by failed transfer, accepted NFT transfer followed by old-owner resolver update, export/import after domain changes"
      expected_rejection  = "rejected identity operations must preserve owner, resolver, expiry, and NFT representation consistency"
    }
    "ID-05" = [ordered]@{
      flow                = "auction bids, renewal fees, refunds, and domain payments to valid targets"
      state               = "auction bids, renewal fees, refunds, and domain payments cannot be stolen or routed to invalid targets"
      attack              = "auction bid theft, renewal fee theft, refund theft, invalid payment target routing"
      invariant           = "identity economic flows preserve bid escrow, renewal fee accounting, refunds, and valid payment targets"
      expected_behavior   = "identity economic rules keep auction escrow, renewal fees, refunds, and payments deterministic and authorized"
      expected_events     = "no bid theft, renewal fee theft, refund theft, or invalid-target payment event appears for rejected identity economic abuse paths"
      expected_error_path = "identity economic abuse rejects before escrow, fee, refund, payment, owner, or resolver state mutation"
      mutation_inputs     = "losing bid refund redirect, renewal fee redirect, auction escrow drain, invalid payment target, duplicate refund claim"
      expected_rejection  = "identity economic abuse must not steal bids, renewal fees, refunds, or route domain payments to invalid targets"
    }
    "REP-01" = [ordered]@{
      flow                = "score updates, decay, level assignment, rate limit, priority signal"
      state               = "reputation score, decay state, level, rate limit, and priority signal update deterministically"
      attack              = "valid reputation lifecycle baseline plus unauthorized score writer control sample"
      invariant           = "reputation lifecycle preserves bounded score, deterministic decay, rate limits, and priority class rules"
      expected_behavior   = "valid score update, decay, level assignment, rate limit, and priority signal update reputation state exactly once"
      expected_events     = "reputation events match score, decay, level, rate-limit, and priority deltas"
      expected_error_path = "unauthorized score writer control sample is rejected before reputation state mutation"
      mutation_inputs     = "valid score update, valid decay tick, valid level assignment, valid rate-limit update, valid priority signal, unauthorized score writer"
      expected_rejection  = "unauthorized reputation lifecycle variants must fail without score, level, rate-limit, or priority mutation"
    }
    "REP-02" = [ordered]@{
      flow                = "score floor, score ceiling, inactive accounts, new accounts, zero activity, max activity"
      state               = "reputation edge cases clamp score, level, rate limit, and priority deterministically"
      attack              = "score floor bypass, score ceiling overflow, inactive account mutation, new account bootstrap abuse, zero or max activity boundary"
      invariant           = "reputation scores remain within floor and ceiling and account activity boundaries are deterministic"
      expected_behavior   = "valid reputation boundaries execute deterministically; invalid boundaries reject before score mutation"
      expected_events     = "accepted boundary reputation events match state deltas; rejected edge cases emit no success events"
      expected_error_path = "reputation validation rejects out-of-bound score, inactive-account abuse, invalid new-account state, or unsafe max activity"
      mutation_inputs     = "score floor minus one, score ceiling plus one, inactive account update, new account with forged history, zero activity, max activity plus one"
      expected_rejection  = "invalid reputation edge cases must not alter score, level, rate limit, or priority state"
    }
    "REP-03" = [ordered]@{
      flow                = "reputation farming, sybil bypass, spam with low score, priority manipulation"
      state               = "adversarial reputation attempts cannot farm score, bypass sybil controls, spam low-score lanes, or manipulate priority"
      attack              = "reputation farming, sybil bypass, spam with low score, priority manipulation"
      invariant           = "reputation cannot be farmed or used to bypass bounded rate limits and priority ordering"
      expected_behavior   = "adversarial reputation mutations fail deterministically before score, rate-limit, or priority corruption"
      expected_events     = "failed reputation attacks emit no misleading score, level, rate-limit, or priority success events"
      expected_error_path = "reputation transition rejects farming, sybil bypass, low-score spam, or priority manipulation before commit"
      mutation_inputs     = "rapid self-activity loop, sybil account fanout, low-score spam burst, forged high-priority reputation signal, repeated decay timing sequence"
      expected_rejection  = "reputation attacks must not farm score, bypass sybil controls, spam accepted lanes, or manipulate priority"
    }
    "REP-04" = [ordered]@{
      flow                = "deterministic score replay, export, and import integrity"
      state               = "score updates, decay, levels, limits, and priority signals do not diverge across replay/export/import"
      attack              = "state drift attempt through mixed accepted and rejected reputation updates plus replay/export/import"
      invariant           = "score updates preserve deterministic replay and do not diverge across replay/export/import"
      expected_behavior   = "reputation state integrity holds across score update, decay, rejected update, replay, export, and import sequences"
      expected_events     = "reputation events reconcile to final score, decay, level, rate-limit, and priority deltas"
      expected_error_path = "failed reputation operations preserve pre-failure score, level, rate-limit, priority, and replay snapshots"
      mutation_inputs     = "accepted score update followed by failed update, accepted decay followed by replay, export/import after priority updates"
      expected_rejection  = "rejected reputation operations must preserve deterministic replay and export/import consistency"
    }
    "REP-05" = [ordered]@{
      flow                = "reputation economic abuse around direct purchase, fee bypass, deposit bypass, and signer bypass"
      state               = "reputation cannot be bought directly and cannot bypass required fees, deposits, or signer checks"
      attack              = "direct reputation purchase, required fee bypass, required deposit bypass, signer check bypass"
      invariant           = "reputation cannot substitute for protocol fees, deposits, or signer authorization"
      expected_behavior   = "reputation economic rules keep score influence bounded to allowed priority/rate-limit effects"
      expected_events     = "no fee bypass, deposit bypass, signer bypass, or direct score purchase event appears for rejected reputation abuse paths"
      expected_error_path = "economic reputation abuse rejects before fee, deposit, signer, score, or priority state mutation"
      mutation_inputs     = "payment shaped as score purchase, high reputation with zero fee, high reputation without deposit, forged signer reputation proof, priority lane without signer"
      expected_rejection  = "reputation economic abuse must not buy score directly or bypass required fees, deposits, or signer checks"
    }
    "EXEC-01" = [ordered]@{
      flow                = "transaction pipeline order, dispatch, route output, events, deterministic trace"
      state               = "execution result, route output, events, receipts, and deterministic trace update in canonical pipeline order"
      attack              = "valid execution pipeline baseline plus unauthorized dispatch control sample"
      invariant           = "execution pipeline order, dispatch target, route output, events, and trace are deterministic"
      expected_behavior   = "valid pipeline validation, dispatch, route output, event emission, and trace recording execute exactly once"
      expected_events     = "execution events match dispatch, route, receipt, and trace deltas"
      expected_error_path = "unauthorized dispatch control sample is rejected before execution state mutation"
      mutation_inputs     = "valid pipeline tx, valid dispatch, valid route output, valid event trace, unauthorized dispatch control"
      expected_rejection  = "unauthorized execution lifecycle variants must fail without dispatch, receipt, route, event, or trace mutation"
    }
    "EXEC-02" = [ordered]@{
      flow                = "malformed payload, missing route, invalid module, failed dispatch, max tx size"
      state               = "execution edge cases leave dispatch state, route output, events, receipts, and traces unchanged unless explicitly accepted"
      attack              = "malformed payload, missing route, invalid module, failed dispatch, max tx size boundary"
      invariant           = "execution accepts only valid payloads, routes, modules, dispatch results, and bounded tx sizes"
      expected_behavior   = "valid execution boundaries execute deterministically; invalid boundaries reject before state mutation"
      expected_events     = "accepted boundary execution events match trace deltas; rejected edge cases emit no success events"
      expected_error_path = "execution validation rejects malformed payloads, missing routes, invalid modules, failed dispatches, or unsafe max tx sizes"
      mutation_inputs     = "malformed payload bytes, missing route, invalid module id, forced dispatch failure, max tx size plus one"
      expected_rejection  = "invalid execution edge cases must not alter dispatch state, route output, events, receipts, or traces"
    }
    "EXEC-03" = [ordered]@{
      flow                = "partial rollback, wrong module dispatch, invalid state transition after ante failure, execution desync"
      state               = "adversarial execution attempts cannot partially commit, dispatch to wrong modules, mutate after ante failure, or desync traces"
      attack              = "partial rollback, wrong module dispatch, invalid state transition after ante failure, execution desync"
      invariant           = "execution cannot commit partial writes, bypass ante failure, dispatch to wrong module, or produce nondeterministic traces"
      expected_behavior   = "adversarial execution mutations fail deterministically before state, receipt, or trace corruption"
      expected_events     = "failed execution attacks emit no misleading dispatch, receipt, event, or trace success events"
      expected_error_path = "execution path rejects rollback, dispatch, ante-bypass, or desync attack before commit"
      mutation_inputs     = "panic-shaped partial write, wrong module route, ante-failed tx with state write, nondeterministic trace input, duplicate dispatch"
      expected_rejection  = "execution attacks must not partially commit, wrong-dispatch, mutate after ante failure, or desync execution traces"
    }
    "EXEC-04" = [ordered]@{
      flow                = "failed execution no partial writes and accepted execution stable receipts"
      state               = "failed execution does not commit partial writes and accepted execution emits stable receipts"
      attack              = "state drift attempt through mixed accepted and rejected execution, dispatch, route, and receipt operations"
      invariant           = "failed execution does not commit partial writes and accepted execution emits stable receipts"
      expected_behavior   = "execution state integrity holds across accepted dispatch, failed dispatch, replay, export, and import sequences"
      expected_events     = "execution events reconcile to final receipts, route output, trace, and committed state deltas"
      expected_error_path = "failed execution preserves pre-failure state writes, route output, events, receipts, and trace snapshots"
      mutation_inputs     = "accepted dispatch followed by failed dispatch, accepted route followed by invalid module, failed execution then replay, export/import after receipts"
      expected_rejection  = "rejected execution operations must preserve no-partial-write and stable-receipt consistency"
    }
    "EXEC-05" = [ordered]@{
      flow                = "execution economic abuse around fee, gas, memo, reputation, and routing constraints"
      state               = "execution cannot bypass fee, gas, memo, reputation, or routing constraints"
      attack              = "fee bypass, gas bypass, memo bypass, reputation bypass, routing constraint bypass"
      invariant           = "execution cannot bypass fee, gas, memo, reputation, or routing constraints"
      expected_behavior   = "execution economic rules preserve ante, gas, memo, reputation, and routing gates before dispatch"
      expected_events     = "no dispatch, receipt, or trace success event appears for rejected execution constraint bypass paths"
      expected_error_path = "execution economic abuse rejects before dispatch, state write, receipt, event, or trace mutation"
      mutation_inputs     = "zero fee execution, gas underpayment execution, oversized memo execution, forged reputation execution, missing route execution"
      expected_rejection  = "execution economic abuse must not bypass fee, gas, memo, reputation, or routing constraints"
    }
    "VM-01" = [ordered]@{
      flow                = "AVM deploy, external call, internal call, bounced call, query, migrate entrypoint validation"
      state               = "contract code, contract state, call results, query output, migration state, and emitted messages update deterministically"
      attack              = "valid AVM lifecycle baseline plus unauthorized code owner or contract admin control sample"
      invariant           = "AVM lifecycle preserves deterministic entrypoint validation, gas accounting, state writes, and emitted messages"
      expected_behavior   = "valid deploy, external call, internal call, bounced call, query, and migrate paths validate entrypoints and execute exactly once"
      expected_events     = "AVM events match deploy, call, bounce, query, migrate, state, gas, and emitted-message deltas"
      expected_error_path = "unauthorized code owner or contract admin control sample is rejected before code, state, or queue mutation"
      mutation_inputs     = "valid AVM deploy, valid external call, valid internal call, valid bounced call, valid query, valid migrate, unauthorized admin control"
      expected_rejection  = "unauthorized AVM lifecycle variants must fail without contract state, code, queue, gas, or message mutation"
    }
    "VM-02" = [ordered]@{
      flow                = "max code size, missing entrypoint, bad code hash, zero gas, max gas, malformed bytecode"
      state               = "AVM edge cases leave code store, contract state, gas accounting, and emitted messages unchanged unless explicitly accepted"
      attack              = "oversized code, missing entrypoint, bad code hash, zero gas, max gas boundary, malformed bytecode"
      invariant           = "AVM accepts only bounded code, valid entrypoints, valid code hashes, valid gas limits, and well-formed bytecode"
      expected_behavior   = "valid AVM boundaries execute deterministically; invalid boundaries reject before contract state mutation"
      expected_events     = "accepted boundary AVM events match gas/state deltas; rejected edge cases emit no success events"
      expected_error_path = "AVM validation rejects oversized code, missing entrypoints, bad code hashes, zero gas, unsafe max gas, or malformed bytecode"
      mutation_inputs     = "max code size plus one, missing entrypoint, bad code hash, zero gas, max gas plus one, malformed bytecode"
      expected_rejection  = "invalid AVM edge cases must not alter code store, contract state, gas accounting, queues, or emitted messages"
    }
    "VM-03" = [ordered]@{
      flow                = "VM crash input, infinite loop, stack overflow, sandbox escape, nondeterministic host behavior"
      state               = "adversarial AVM attempts cannot crash consensus, escape sandbox, exhaust unbounded gas, or use nondeterministic host behavior"
      attack              = "VM crash input, infinite loop, stack overflow, sandbox escape, nondeterministic host behavior"
      invariant           = "AVM malformed input does not panic and host functions remain gas-bounded, sandboxed, and deterministic"
      expected_behavior   = "adversarial AVM mutations fail deterministically with bounded gas and without state corruption"
      expected_events     = "failed AVM attacks emit no misleading deploy, execute, migrate, state, gas, or message success events"
      expected_error_path = "AVM runtime rejects crash input, infinite loop, stack overflow, sandbox escape, or nondeterministic host call before commit"
      mutation_inputs     = "panic-shaped bytecode, infinite loop bytecode, deep stack bytecode, forbidden host call, local-time host call, randomness host call"
      expected_rejection  = "AVM attacks must not panic consensus, escape sandbox, exceed gas bounds, commit state, or use nondeterministic host behavior"
    }
    "VM-04" = [ordered]@{
      flow                = "deterministic contract state changes and rejected execution no state commit"
      state               = "contract state changes are deterministic and rejected execution cannot commit state"
      attack              = "state drift attempt through mixed accepted and rejected AVM deploy, execute, query, migrate, and bounced call operations"
      invariant           = "contract state changes are deterministic and rejected AVM execution cannot commit state"
      expected_behavior   = "AVM state integrity holds across deploy, execute, query, migrate, bounce, replay, export, and import sequences"
      expected_events     = "AVM events reconcile to final contract state, gas, queue, and emitted-message deltas"
      expected_error_path = "failed AVM execution preserves pre-failure code store, contract state, gas, queue, and emitted-message snapshots"
      mutation_inputs     = "accepted execute followed by failed execute, accepted migrate followed by failed migrate, failed query side-effect attempt, export/import after contract state changes"
      expected_rejection  = "rejected AVM operations must preserve deterministic state and no-state-commit consistency"
    }
    "VM-05" = [ordered]@{
      flow                = "AVM economic abuse around gas underpayment, non-naet protocol fees, double-refund, and storage limits"
      state               = "AVM cannot underpay gas, pay protocol fees in non-naet, double-refund, or bypass storage limits"
      attack              = "gas underpayment, non-naet fee payment, double-refund, storage limit bypass"
      invariant           = "AVM cannot underpay gas, pay protocol fees in non-naet, double-refund, or bypass storage limits"
      expected_behavior   = "AVM economic rules enforce ante fees, gas metering, refund idempotence, and bounded storage before commit"
      expected_events     = "no gas underpayment, non-naet fee acceptance, double-refund, or storage-limit success event appears for rejected AVM economic abuse paths"
      expected_error_path = "AVM economic abuse rejects before contract state, refund, storage, gas, or fee accounting mutation"
      mutation_inputs     = "execute with underpaid gas, factory denom fee, duplicate refund receipt, oversized storage write, query response storage bypass"
      expected_rejection  = "AVM economic abuse must not underpay gas, pay protocol fees in non-naet, double-refund, or bypass storage limits"
    }
    "MSG-01" = [ordered]@{
      flow                = "async send, internal message delivery, proof/receipt fields, cross-zone message classification"
      state               = "message state, receipt fields, proof references, classification, and delivery markers update deterministically"
      attack              = "valid messaging lifecycle baseline plus unauthorized sender or destination control sample"
      invariant           = "messaging lifecycle preserves deterministic classification, proof fields, receipts, delivery, and replay markers"
      expected_behavior   = "valid async send, internal delivery, proof/receipt validation, and cross-zone classification execute exactly once"
      expected_events     = "messaging events match send, delivery, proof, receipt, classification, and replay-marker deltas"
      expected_error_path = "unauthorized sender or destination control sample is rejected before message, receipt, or queue mutation"
      mutation_inputs     = "valid async send, valid internal delivery, valid proof fields, valid receipt fields, valid cross-zone classification, unauthorized sender control"
      expected_rejection  = "unauthorized messaging lifecycle variants must fail without message, receipt, proof, queue, or replay-marker mutation"
    }
    "MSG-02" = [ordered]@{
      flow                = "missing destination, expired message, max body, zero value, malformed opcode, invalid query id"
      state               = "messaging edge cases leave message state, receipts, queues, and value accounting unchanged unless explicitly accepted"
      attack              = "missing destination, expired message, oversized body, zero value, malformed opcode, invalid query id"
      invariant           = "messaging accepts only valid destinations, live messages, bounded bodies, valid value semantics, opcodes, and query ids"
      expected_behavior   = "valid messaging boundaries execute deterministically; invalid boundaries reject before message or value mutation"
      expected_events     = "accepted boundary messaging events match receipt/value deltas; rejected edge cases emit no success events"
      expected_error_path = "messaging validation rejects missing destination, expired message, oversized body, invalid zero-value path, malformed opcode, or invalid query id"
      mutation_inputs     = "missing destination, expired message, max body plus one, zero value where value required, malformed opcode, invalid query id"
      expected_rejection  = "invalid messaging edge cases must not alter message state, receipts, queues, replay markers, or value accounting"
    }
    "MSG-03" = [ordered]@{
      flow                = "message replay, message ordering attack, forged proof, stale receipt replay, message starvation"
      state               = "adversarial messaging attempts cannot replay messages, reorder canonically, forge proofs, replay stale receipts, or starve queues"
      attack              = "message replay, message ordering attack, forged proof, stale receipt replay, message starvation"
      invariant           = "messaging proof validation, replay markers, receipt uniqueness, canonical ordering, and queue progress cannot be bypassed"
      expected_behavior   = "adversarial messaging mutations fail deterministically before message, receipt, queue, or value state corruption"
      expected_events     = "failed messaging attacks emit no misleading send, delivery, proof, receipt, replay, or queue success events"
      expected_error_path = "messaging transition rejects replay, ordering, forged proof, stale receipt, or starvation attack before commit"
      mutation_inputs     = "duplicate message id, out-of-order delivery, forged proof root, stale receipt replay, high-priority starvation sequence"
      expected_rejection  = "messaging attacks must not replay messages, reorder delivery, forge proofs, replay stale receipts, or starve messages"
    }
    "MSG-04" = [ordered]@{
      flow                = "message state, receipts, and queue entries deterministic across replay/export/import"
      state               = "message state, receipt state, queue entries, replay markers, and ordering do not diverge across replay/export/import"
      attack              = "state drift attempt through mixed accepted and rejected messaging send, delivery, receipt, proof, and queue operations"
      invariant           = "message state, receipts, and queue entries are deterministic across replay/export/import"
      expected_behavior   = "messaging state integrity holds across send, delivery, receipt, proof, replay, export, and import sequences"
      expected_events     = "messaging events reconcile to final message, receipt, queue, replay-marker, and ordering deltas"
      expected_error_path = "failed messaging operations preserve pre-failure message, receipt, proof, queue, replay-marker, and value snapshots"
      mutation_inputs     = "accepted send followed by failed replay, accepted receipt followed by stale receipt replay, accepted proof followed by forged proof, export/import after message delivery"
      expected_rejection  = "rejected messaging operations must preserve deterministic replay, receipt, queue, and export/import consistency"
    }
    "MSG-05" = [ordered]@{
      flow                = "message forwarding fees, value transfer, bounce, and refund double-spend prevention"
      state               = "message forwarding fees, value transfer, bounce, and refund cannot double-spend"
      attack              = "forwarding fee bypass, value transfer double-spend, bounce double-spend, refund double-spend"
      invariant           = "message forwarding fees, value transfer, bounce, and refund cannot double-spend"
      expected_behavior   = "messaging economic rules enforce fee payment, value conservation, single-use bounces, and single-use refunds"
      expected_events     = "no fee bypass, value double-spend, bounce double-spend, or refund double-spend event appears for rejected messaging economic abuse paths"
      expected_error_path = "messaging economic abuse rejects before fee, value, bounce, refund, receipt, or replay-marker mutation"
      mutation_inputs     = "forward without fee, duplicate value transfer, duplicate bounce receipt, duplicate refund receipt, forged refund destination"
      expected_rejection  = "messaging economic abuse must not bypass forwarding fees, double-spend value, double-bounce, or double-refund"
    }
    "QUEUE-01" = [ordered]@{
      flow                = "enqueue, delayed execution, dequeue, bounce, refund, per-block processing limit"
      state               = "queue items, delayed execution state, dequeue markers, bounce state, refund state, and per-block counters update deterministically"
      attack              = "valid queue lifecycle baseline plus unauthorized queue actor control sample"
      invariant           = "queue lifecycle preserves deterministic ordering, sequence counters, refund uniqueness, and per-block processing limits"
      expected_behavior   = "valid enqueue, delayed execution, dequeue, bounce, refund, and per-block processing paths update queue state exactly once"
      expected_events     = "queue events match enqueue, dequeue, delay, bounce, refund, and per-block counter deltas"
      expected_error_path = "unauthorized queue actor control sample is rejected before queue, bounce, refund, or value state mutation"
      mutation_inputs     = "valid enqueue, valid delayed execution, valid dequeue, valid bounce, valid refund, valid per-block processing limit, unauthorized actor control"
      expected_rejection  = "unauthorized queue lifecycle variants must fail without queue item, sequence, bounce, refund, or value mutation"
    }
    "QUEUE-02" = [ordered]@{
      flow                = "empty queue, max queue, max depth, expired item, duplicate sequence, missing actor"
      state               = "queue edge cases leave queue items, depth counters, sequence counters, actors, and refund state unchanged unless explicitly accepted"
      attack              = "empty queue dequeue, max queue overflow, max depth overflow, expired item replay, duplicate sequence, missing actor"
      invariant           = "queue accepts only bounded queue depth, valid sequence counters, live items, and valid actors"
      expected_behavior   = "valid queue boundaries execute deterministically; invalid boundaries reject before queue or value mutation"
      expected_events     = "accepted boundary queue events match queue deltas; rejected edge cases emit no success events"
      expected_error_path = "queue validation rejects empty dequeue, max queue overflow, max depth overflow, expired items, duplicate sequence, or missing actor"
      mutation_inputs     = "empty queue dequeue, max queue plus one, max depth plus one, expired item, duplicate sequence, missing actor"
      expected_rejection  = "invalid queue edge cases must not alter queue items, depth counters, sequence counters, actors, bounce, refund, or value state"
    }
    "QUEUE-03" = [ordered]@{
      flow                = "queue flooding, message loop, starvation, priority manipulation, duplicate sequence injection"
      state               = "adversarial queue attempts cannot flood unbounded state, create message loops, starve items, manipulate priority, or inject duplicate sequences"
      attack              = "queue flooding, message loop, starvation, priority manipulation, duplicate sequence injection"
      invariant           = "queue depth, ordering, progress, priority, and sequence uniqueness cannot be bypassed"
      expected_behavior   = "adversarial queue mutations fail deterministically before queue, value, or priority state corruption"
      expected_events     = "failed queue attacks emit no misleading enqueue, dequeue, bounce, refund, priority, or sequence success events"
      expected_error_path = "queue transition rejects flooding, loops, starvation, priority manipulation, or duplicate sequence injection before commit"
      mutation_inputs     = "many enqueue burst, self-reenqueuing loop, high-priority starvation sequence, forged priority, duplicate sequence item"
      expected_rejection  = "queue attacks must not flood unbounded state, loop indefinitely, starve valid items, manipulate priority, or inject duplicate sequences"
    }
    "QUEUE-04" = [ordered]@{
      flow                = "queue ordering, sequence counters, deterministic export and import stability"
      state               = "queue ordering, sequence counters, depth counters, bounce/refund state, and delayed execution state do not diverge across replay/export/import"
      attack              = "state drift attempt through mixed accepted and rejected enqueue, dequeue, bounce, refund, delay, and export/import operations"
      invariant           = "queue ordering and sequence counters are deterministic and export/import stable"
      expected_behavior   = "queue state integrity holds across enqueue, dequeue, delay, bounce, refund, replay, export, and import sequences"
      expected_events     = "queue events reconcile to final queue ordering, sequence, depth, bounce, refund, and delayed execution deltas"
      expected_error_path = "failed queue operations preserve pre-failure queue ordering, sequence counters, depth, bounce, refund, and value snapshots"
      mutation_inputs     = "accepted enqueue followed by duplicate sequence, accepted dequeue followed by replay, accepted bounce followed by failed refund, export/import after queue processing"
      expected_rejection  = "rejected queue operations must preserve deterministic ordering, sequence counters, and export/import consistency"
    }
    "QUEUE-05" = [ordered]@{
      flow                = "queued value refund uniqueness, forwarding fees, and malformed bounce path trapping"
      state               = "queued value cannot be refunded twice, forwarded without fee, or trapped by malformed bounce path"
      attack              = "double refund, fee-free forward, malformed bounce trap, forged refund target"
      invariant           = "queued value cannot be refunded twice, forwarded without fee, or trapped by malformed bounce path"
      expected_behavior   = "queue economic rules enforce forwarding fees, value conservation, single-use refunds, and safe malformed bounce handling"
      expected_events     = "no double refund, fee-free forward, trapped value, or malformed bounce success event appears for rejected queue economic abuse paths"
      expected_error_path = "queue economic abuse rejects before fee, value, bounce, refund, queue, or replay-marker mutation"
      mutation_inputs     = "duplicate refund item, forward without fee, malformed bounce destination, forged refund target, bounce path that traps queued value"
      expected_rejection  = "queue economic abuse must not let queued value be refunded twice, forwarded without fee, or trapped by malformed bounce path"
    }
    "EVENTS-01" = [ordered]@{
      flow                = "deterministic event emission for bank, fees, DEX, identity, execution, queue, and memo paths"
      state               = "event stream, receipt linkage, attributes, indexes, and committed state references are emitted deterministically"
      attack              = "valid event emission baseline plus unauthorized event source control sample"
      invariant           = "event emission is deterministic and derived from committed module state transitions only"
      expected_behavior   = "valid bank, fees, DEX, identity, execution, queue, and memo paths emit deterministic events exactly once"
      expected_events     = "event stream order and attributes match committed state and receipts"
      expected_error_path = "unauthorized event source control sample is rejected before event, receipt, or index mutation"
      mutation_inputs     = "valid bank event, valid fees event, valid DEX event, valid identity event, valid execution event, valid queue event, valid memo event, unauthorized event source"
      expected_rejection  = "unauthorized event emission variants must fail without event stream, receipt, index, or state mutation"
    }
    "EVENTS-02" = [ordered]@{
      flow                = "empty attributes, max attribute size, duplicate event keys, failed tx event behavior"
      state               = "event edge cases leave event stream, receipt linkage, indexes, and committed state references deterministic"
      attack              = "empty attributes, oversized attributes, duplicate event keys, misleading failed tx event"
      invariant           = "events use bounded attributes, deterministic key ordering, and explicit failed tx behavior"
      expected_behavior   = "valid event boundaries execute deterministically; invalid boundaries reject or sanitize before event/index mutation"
      expected_events     = "accepted boundary events match receipt and state deltas; rejected edge cases emit no misleading success events"
      expected_error_path = "event validation rejects oversized attributes, duplicate event keys where disallowed, and misleading failed tx success event behavior"
      mutation_inputs     = "empty attributes, max attribute size plus one, duplicate event keys, failed tx with success-shaped event"
      expected_rejection  = "invalid event edge cases must not create misleading event stream, receipt linkage, indexes, or state references"
    }
    "EVENTS-03" = [ordered]@{
      flow                = "event spoofing, inconsistent event order, misleading success event after failure"
      state               = "adversarial event attempts cannot spoof authority, reorder inconsistently, or emit success after failed state transitions"
      attack              = "event spoofing, inconsistent event order, misleading success event after failure"
      invariant           = "events cannot spoof state authority, violate deterministic ordering, or claim success after failure"
      expected_behavior   = "adversarial event mutations fail deterministically before event stream or receipt corruption"
      expected_events     = "failed event attacks emit no misleading authority, balance, resolver, execution, or success events"
      expected_error_path = "event emission path rejects spoofed events, inconsistent ordering, or success-after-failure before commit"
      mutation_inputs     = "spoofed bank transfer event, reversed event order, success event after failed tx, forged resolver event, forged execution event"
      expected_rejection  = "event attacks must not spoof authority, reorder events nondeterministically, or emit misleading success after failure"
    }
    "EVENTS-04" = [ordered]@{
      flow                = "events match committed state and receipts"
      state               = "event stream, receipt linkage, indexes, and committed state references reconcile after accepted and rejected operations"
      attack              = "state drift attempt through mixed accepted and rejected event-emitting module operations"
      invariant           = "events match committed state and receipts"
      expected_behavior   = "event state integrity holds across event emission, receipt linkage, replay, export, and import sequences"
      expected_events     = "events reconcile to committed state, receipts, indexes, and final app traces"
      expected_error_path = "failed event-emitting operations preserve pre-failure event stream, receipt linkage, indexes, and committed state references"
      mutation_inputs     = "accepted bank event followed by failed spoof, accepted execution receipt followed by failed event, export/import after event-emitting txs"
      expected_rejection  = "rejected event operations must preserve committed-state and receipt consistency"
    }
    "EVENTS-05" = [ordered]@{
      flow                = "event economic abuse around authority for balances, fees, resolver targets, and execution success"
      state               = "events cannot be used as authority for balances, fees, resolver targets, or execution success"
      attack              = "event-as-authority balance spoof, fee spoof, resolver target spoof, execution success spoof"
      invariant           = "events are observational and cannot authorize balances, fees, resolver targets, or execution success"
      expected_behavior   = "event economic rules keep authority in committed state, not event text or indexes"
      expected_events     = "no balance, fee, resolver, or execution authority event is accepted without matching committed state"
      expected_error_path = "event authority abuse rejects before balance, fee, resolver, execution, receipt, or index state mutation"
      mutation_inputs     = "balance update from event only, fee paid event without bank state, resolver target from event only, execution success event without receipt"
      expected_rejection  = "events must not be used as authority for balances, fees, resolver targets, or execution success"
    }
    "ACTOR-01" = [ordered]@{
      flow                = "actor lifecycle, mailbox processing, logical time, isolated state transition"
      state               = "actor state, mailbox state, logical time, lifecycle markers, and isolated state transition update deterministically"
      attack              = "valid actor lifecycle baseline plus unauthorized actor controller sample"
      invariant           = "actor lifecycle preserves isolated state transitions, mailbox ordering, and monotonic logical time"
      expected_behavior   = "valid actor creation, activation, mailbox processing, logical-time update, and isolated state transition execute exactly once"
      expected_events     = "actor events match lifecycle, mailbox, logical-time, and isolated state deltas"
      expected_error_path = "unauthorized actor controller sample is rejected before actor, mailbox, logical-time, or state mutation"
      mutation_inputs     = "valid actor create, valid mailbox process, valid logical-time tick, valid isolated state transition, unauthorized actor controller"
      expected_rejection  = "unauthorized actor lifecycle variants must fail without actor state, mailbox, logical-time, or isolated state mutation"
    }
    "ACTOR-02" = [ordered]@{
      flow                = "missing actor, inactive actor, max mailbox, max state size, actor deletion and migration boundaries"
      state               = "actor edge cases leave actor state, mailbox state, logical time, deletion markers, and migration state unchanged unless explicitly accepted"
      attack              = "missing actor call, inactive actor call, max mailbox overflow, max state size overflow, invalid deletion or migration boundary"
      invariant           = "actors accept only existing active actors, bounded mailboxes, bounded state, and valid deletion or migration boundaries"
      expected_behavior   = "valid actor boundaries execute deterministically; invalid boundaries reject before actor or mailbox mutation"
      expected_events     = "accepted boundary actor events match state deltas; rejected edge cases emit no success events"
      expected_error_path = "actor validation rejects missing actor, inactive actor, mailbox overflow, state size overflow, or invalid deletion/migration boundary"
      mutation_inputs     = "missing actor id, inactive actor message, max mailbox plus one, max state size plus one, delete active actor with pending mailbox, invalid migration target"
      expected_rejection  = "invalid actor edge cases must not alter actor state, mailbox state, logical time, deletion markers, or migration state"
    }
    "ACTOR-03" = [ordered]@{
      flow                = "cross-actor direct state mutation, mailbox flood, logical-time spoof, actor takeover"
      state               = "adversarial actor attempts cannot directly mutate another actor, flood mailboxes unboundedly, spoof logical time, or take over actors"
      attack              = "cross-actor direct state mutation, mailbox flood, logical-time spoof, actor takeover"
      invariant           = "actor isolation, mailbox bounds, monotonic logical time, and ownership checks cannot be bypassed"
      expected_behavior   = "adversarial actor mutations fail deterministically before actor, mailbox, logical-time, or ownership state corruption"
      expected_events     = "failed actor attacks emit no misleading state, mailbox, logical-time, ownership, or migration success events"
      expected_error_path = "actor transition rejects direct cross-actor mutation, mailbox flood, logical-time spoof, or takeover before commit"
      mutation_inputs     = "direct write to another actor, mailbox flood burst, lower logical time, future logical time jump, forged actor owner"
      expected_rejection  = "actor attacks must not mutate another actor directly, flood mailboxes, spoof logical time, or take over actor ownership"
    }
    "ACTOR-04" = [ordered]@{
      flow                = "actor isolation through committed messages only"
      state               = "one actor cannot mutate another actor except through committed messages"
      attack              = "state drift attempt through mixed accepted and rejected cross-actor mailbox and direct state operations"
      invariant           = "one actor cannot mutate another actor except through committed messages"
      expected_behavior   = "actor state integrity holds across mailbox delivery, rejected direct mutation, replay, export, and import sequences"
      expected_events     = "actor events reconcile to committed message, mailbox, logical-time, and isolated state deltas"
      expected_error_path = "failed actor operations preserve pre-failure actor state, mailbox state, logical time, and ownership snapshots"
      mutation_inputs     = "accepted committed message followed by direct mutation attempt, accepted mailbox delivery followed by replay, export/import after actor transitions"
      expected_rejection  = "rejected actor operations must preserve actor isolation and committed-message-only mutation consistency"
    }
    "ACTOR-05" = [ordered]@{
      flow                = "actor economic abuse around storage, execution, message costs, actor splitting, and mailbox abuse"
      state               = "actor storage, execution, and message costs cannot be avoided through actor splitting or mailbox abuse"
      attack              = "storage cost bypass, execution cost bypass, message cost bypass, actor splitting, mailbox abuse"
      invariant           = "actor storage, execution, and message costs cannot be avoided through actor splitting or mailbox abuse"
      expected_behavior   = "actor economic rules enforce storage, execution, and message cost accounting across actor and mailbox layouts"
      expected_events     = "no storage, execution, message, actor splitting, or mailbox cost bypass event appears for rejected actor economic abuse paths"
      expected_error_path = "actor economic abuse rejects before actor state, mailbox, storage, execution, message, or fee accounting mutation"
      mutation_inputs     = "many tiny actors to bypass storage cost, mailbox fanout to bypass message cost, repeated cheap execution, split actor state, unpaid mailbox message"
      expected_rejection  = "actor economic abuse must not avoid storage, execution, or message costs through actor splitting or mailbox abuse"
    }
    "SCHED-01" = [ordered]@{
      flow                = "deterministic ordering, task selection, read/write set handling, priority class handling"
      state               = "plan output, selected tasks, read/write conflict results, priority class ordering, and task status update deterministically"
      attack              = "valid scheduler lifecycle baseline plus unauthorized scheduler input control sample"
      invariant           = "scheduler planning preserves deterministic ordering, read/write conflict handling, and bounded priority class rules"
      expected_behavior   = "valid task ordering, selection, read/write handling, and priority handling produce one deterministic execution plan"
      expected_events     = "scheduler events match plan output, task selection, conflict result, and priority deltas"
      expected_error_path = "unauthorized scheduler input control sample is rejected before plan, task, or priority state mutation"
      mutation_inputs     = "valid task set, valid priority class, valid read/write set, valid dependency graph, unauthorized scheduler input"
      expected_rejection  = "unauthorized scheduler lifecycle variants must fail without plan output, task status, conflict result, or priority mutation"
    }
    "SCHED-02" = [ordered]@{
      flow                = "empty plan, duplicate task id, max tasks, conflicting read/write sets, dependency boundaries"
      state               = "scheduler edge cases leave plan output, task status, dependency state, conflict results, and priority state deterministic"
      attack              = "empty plan abuse, duplicate task id, max tasks overflow, conflicting read/write set, dependency boundary violation"
      invariant           = "scheduler accepts only canonical task ids, bounded task counts, deterministic conflicts, and valid dependency boundaries"
      expected_behavior   = "valid scheduler boundaries execute deterministically; invalid boundaries reject before plan mutation"
      expected_events     = "accepted boundary scheduler events match plan deltas; rejected edge cases emit no success events"
      expected_error_path = "scheduler validation rejects duplicate task ids, max task overflow, invalid conflicts, and dependency boundary violations"
      mutation_inputs     = "empty plan, duplicate task id, max tasks plus one, conflicting read/write sets, missing dependency, cyclic dependency"
      expected_rejection  = "invalid scheduler edge cases must not alter plan output, task status, dependency state, conflict results, or priority state"
    }
    "SCHED-03" = [ordered]@{
      flow                = "scheduling manipulation, starvation, priority gaming, nondeterministic tie-break"
      state               = "adversarial scheduler attempts cannot manipulate plans, starve tasks, game priority, or use nondeterministic tie-breaks"
      attack              = "scheduling manipulation, starvation, priority gaming, nondeterministic tie-break"
      invariant           = "scheduler ordering, starvation prevention, priority bounds, and tie-breaks remain deterministic"
      expected_behavior   = "adversarial scheduler mutations fail deterministically before plan, priority, or task-status corruption"
      expected_events     = "failed scheduler attacks emit no misleading plan, priority, selection, starvation, or tie-break success events"
      expected_error_path = "scheduler transition rejects manipulation, starvation, priority gaming, or nondeterministic tie-break before commit"
      mutation_inputs     = "task order permutation, high-priority starvation sequence, forged priority class, equal priority tie without hash key, validator-local ordering hint"
      expected_rejection  = "scheduler attacks must not manipulate planning, starve tasks, game priority, or introduce nondeterministic tie-breaks"
    }
    "SCHED-04" = [ordered]@{
      flow                = "same tasks and state produce same execution plan across nodes"
      state               = "plan output, task status, conflict results, dependency ordering, and priority ordering do not diverge across replay/export/import"
      attack              = "state drift attempt through same task set with different insertion order, replay, export, and import"
      invariant           = "same tasks and state produce the same execution plan across nodes"
      expected_behavior   = "scheduler state integrity holds across task planning, conflict resolution, replay, export, and import sequences"
      expected_events     = "scheduler events reconcile to final plan output, task status, conflict result, dependency, and priority deltas"
      expected_error_path = "failed scheduler operations preserve pre-failure plan, task status, conflict result, dependency, and priority snapshots"
      mutation_inputs     = "same tasks in different insertion order, accepted plan followed by replay, export/import after planning, equal priority tie-break sequence"
      expected_rejection  = "rejected scheduler operations must preserve deterministic same-input same-plan consistency"
    }
    "SCHED-05" = [ordered]@{
      flow                = "scheduler economic abuse around priority or market signals, starvation, fee caps, and reputation caps"
      state               = "priority or market signals cannot starve normal users or bypass fee/reputation caps"
      attack              = "market signal starvation, priority signal starvation, fee cap bypass, reputation cap bypass"
      invariant           = "priority or market signals cannot starve normal users or bypass fee/reputation caps"
      expected_behavior   = "scheduler economic rules enforce bounded priority influence, anti-starvation, fee caps, and reputation caps"
      expected_events     = "no starvation, fee cap bypass, reputation cap bypass, or priority abuse success event appears for rejected scheduler economic paths"
      expected_error_path = "scheduler economic abuse rejects before plan, priority, market signal, fee, reputation, or task status mutation"
      mutation_inputs     = "high fee spam starving normal tasks, forged market signal, priority above cap, reputation above cap, repeated priority-only plan"
      expected_rejection  = "scheduler economic abuse must not starve normal users or bypass fee/reputation caps through priority or market signals"
    }
    "STORE-01" = [ordered]@{
      flow                = "KV writes, reads, versioning, snapshots, export/import, state sync"
      state               = "key/value state, versions, snapshot roots, exported state, imported state, and state-sync metadata update deterministically"
      attack              = "valid storage lifecycle baseline plus unauthorized store writer control sample"
      invariant           = "storage lifecycle preserves deterministic reads/writes, versioning, snapshot roots, export/import equality, and state-sync integrity"
      expected_behavior   = "valid KV write, read, versioning, snapshot, export/import, and state-sync paths update storage state exactly once"
      expected_events     = "storage events match write, read, version, snapshot, export/import, and state-sync deltas"
      expected_error_path = "unauthorized store writer control sample is rejected before key/value, version, snapshot, or export state mutation"
      mutation_inputs     = "valid KV write, valid KV read, valid version increment, valid snapshot, valid export/import, valid state sync, unauthorized store writer"
      expected_rejection  = "unauthorized storage lifecycle variants must fail without key/value, version, snapshot, export, import, or state-sync mutation"
    }
    "STORE-02" = [ordered]@{
      flow                = "max key, max value, empty value, duplicate key, deleted key, pagination boundaries"
      state               = "storage edge cases leave key/value state, versions, deleted markers, pagination cursors, and snapshot roots deterministic"
      attack              = "max key overflow, max value overflow, empty value ambiguity, duplicate key write, deleted key resurrection, pagination boundary abuse"
      invariant           = "storage accepts only bounded keys/values, canonical deletion semantics, and bounded deterministic pagination"
      expected_behavior   = "valid storage boundaries execute deterministically; invalid boundaries reject before state mutation"
      expected_events     = "accepted boundary storage events match state deltas; rejected edge cases emit no success events"
      expected_error_path = "storage validation rejects oversized keys, oversized values, invalid duplicate writes, deleted-key misuse, or unsafe pagination boundaries"
      mutation_inputs     = "max key plus one, max value plus one, empty value, duplicate key write, deleted key read/write, pagination page limit plus one, invalid pagination next key"
      expected_rejection  = "invalid storage edge cases must not alter key/value state, versions, deleted markers, pagination cursors, or snapshot roots"
    }
    "STORE-03" = [ordered]@{
      flow                = "state root collision, snapshot poisoning, malformed import, unbounded iteration"
      state               = "adversarial storage attempts cannot collide roots, poison snapshots, import malformed state, or trigger unbounded iteration"
      attack              = "state root collision, snapshot poisoning, malformed import, unbounded iteration"
      invariant           = "storage roots, snapshots, imports, and iterators are deterministic, validated, and bounded"
      expected_behavior   = "adversarial storage mutations fail deterministically before root, snapshot, import, or iterator state corruption"
      expected_events     = "failed storage attacks emit no misleading root, snapshot, import, or state-sync success events"
      expected_error_path = "storage path rejects root collision attempt, poisoned snapshot, malformed import, or unbounded iteration before commit"
      mutation_inputs     = "two states with forged same root, poisoned snapshot chunk, malformed import payload, unsorted import keys, unbounded prefix iterator request"
      expected_rejection  = "storage attacks must not collide state roots, poison snapshots, import malformed state, or execute unbounded iteration"
    }
    "STORE-04" = [ordered]@{
      flow                = "committed state root, snapshot root, exported state determinism"
      state               = "committed state root, snapshot root, exported state, versions, and imported state do not diverge across replay/export/import"
      attack              = "state drift attempt through mixed accepted and rejected storage writes, snapshots, export, and import operations"
      invariant           = "committed state root, snapshot root, and exported state are deterministic"
      expected_behavior   = "storage state integrity holds across write, delete, snapshot, replay, export, import, and state-sync sequences"
      expected_events     = "storage events reconcile to final key/value, root, snapshot, export, import, and version deltas"
      expected_error_path = "failed storage operations preserve pre-failure key/value state, roots, snapshots, versions, and export/import snapshots"
      mutation_inputs     = "accepted write followed by failed import, accepted delete followed by replay, snapshot then export/import, same state with different key insertion order"
      expected_rejection  = "rejected storage operations must preserve deterministic committed root, snapshot root, and exported state consistency"
    }
    "STORE-05" = [ordered]@{
      flow                = "storage economic abuse around growth, rent/deposit, and contract state size limits"
      state               = "storage growth, storage rent/deposit, and contract state size limits cannot be bypassed"
      attack              = "storage growth bypass, storage rent bypass, storage deposit bypass, contract state size limit bypass"
      invariant           = "storage growth, rent/deposit, and contract state size limits cannot be bypassed"
      expected_behavior   = "storage economic rules enforce bounded growth, rent/deposit accounting, and contract state size limits before commit"
      expected_events     = "no storage growth, rent, deposit, or contract size limit bypass event appears for rejected storage economic abuse paths"
      expected_error_path = "storage economic abuse rejects before key/value, size, rent, deposit, contract state, or root mutation"
      mutation_inputs     = "many small keys to bypass rent, oversized contract state write, zero deposit storage write, split storage writes, expired rent update"
      expected_rejection  = "storage economic abuse must not bypass storage growth, storage rent/deposit, or contract state size limits"
    }
    "MEMO-01" = [ordered]@{
      flow                = "optional UTF-8 memo on bank, identity, token, DEX, and contract calls"
      state               = "memo metadata, tx metadata, index entries, event attributes, and execution inputs record optional UTF-8 memo deterministically"
      attack              = "valid memo lifecycle baseline plus unauthorized memo mutation control sample"
      invariant           = "memo lifecycle preserves optional UTF-8 validation and cannot alter business execution semantics"
      expected_behavior   = "valid optional UTF-8 memo on bank, identity, token, DEX, and contract calls is recorded exactly once"
      expected_events     = "memo events and indexes match committed tx metadata without changing execution result"
      expected_error_path = "unauthorized memo mutation control sample is rejected before memo, index, event, or execution state mutation"
      mutation_inputs     = "valid bank memo, valid identity memo, valid token memo, valid DEX memo, valid contract memo, unauthorized memo mutation"
      expected_rejection  = "unauthorized memo lifecycle variants must fail without memo metadata, index, event, or execution mutation"
    }
    "MEMO-02" = [ordered]@{
      flow                = "empty memo, max memo, invalid UTF-8, control chars, oversized byte length"
      state               = "memo edge cases leave tx metadata, memo records, index entries, events, and execution results deterministic"
      attack              = "empty memo ambiguity, max memo overflow, invalid UTF-8, control chars, oversized byte length"
      invariant           = "memo accepts only valid UTF-8, bounded byte length, and deterministic control-character policy"
      expected_behavior   = "valid memo boundaries execute deterministically; invalid boundaries reject before memo/index mutation"
      expected_events     = "accepted boundary memo events match metadata deltas; rejected edge cases emit no misleading success events"
      expected_error_path = "memo validation rejects invalid UTF-8, disallowed control chars, and oversized byte length before tx metadata mutation"
      mutation_inputs     = "empty memo, max memo, invalid UTF-8 bytes, control chars, oversized byte length, boundary multibyte UTF-8"
      expected_rejection  = "invalid memo edge cases must not alter tx metadata, memo records, indexes, events, or execution results"
    }
    "MEMO-03" = [ordered]@{
      flow                = "memo spam, binary payload injection, indexing abuse, misleading memo on failed tx"
      state               = "adversarial memo attempts cannot spam indexes, inject binary payloads, abuse indexing, or mislead failed transaction state"
      attack              = "memo spam, binary payload injection, indexing abuse, misleading memo on failed tx"
      invariant           = "memo metadata remains bounded, UTF-8, index-safe, and non-authoritative on failed txs"
      expected_behavior   = "adversarial memo mutations fail deterministically before memo, index, event, or execution corruption"
      expected_events     = "failed memo attacks emit no misleading memo, index, success, or execution events"
      expected_error_path = "memo path rejects spam, binary injection, indexing abuse, or misleading failed-tx memo before commit"
      mutation_inputs     = "many oversized memos, binary payload bytes, index key injection, failed tx with success-shaped memo, repeated memo spam burst"
      expected_rejection  = "memo attacks must not spam indexes, inject binary payloads, abuse indexing, or mislead failed transaction state"
    }
    "MEMO-04" = [ordered]@{
      flow                = "memo immutability after block inclusion and no execution result mutation"
      state               = "memo metadata, tx metadata, indexes, events, and execution results do not diverge across replay/export/import"
      attack              = "state drift attempt through memo mutation after inclusion, replay, export, import, and failed execution"
      invariant           = "memo metadata is immutable after block inclusion and cannot alter execution result"
      expected_behavior   = "memo state integrity holds across block inclusion, replay, export, import, indexing, and execution sequences"
      expected_events     = "memo events reconcile to final tx metadata, indexes, receipts, and unchanged execution results"
      expected_error_path = "failed memo operations preserve pre-failure memo metadata, indexes, events, receipts, and execution result snapshots"
      mutation_inputs     = "accepted memo then post-inclusion mutation, replay with changed memo, export/import after memo indexing, failed tx memo mutation attempt"
      expected_rejection  = "rejected memo operations must preserve immutability and no-execution-result-mutation consistency"
    }
    "MEMO-05" = [ordered]@{
      flow                = "memo economic abuse around memo cost, byte fee, and reputation multiplier"
      state               = "memo cost, byte fee, and reputation multiplier cannot be bypassed"
      attack              = "memo cost bypass, byte fee bypass, reputation multiplier bypass"
      invariant           = "memo cost, byte fee, and reputation multiplier cannot be bypassed"
      expected_behavior   = "memo economic rules enforce byte-based cost, fee accounting, and bounded reputation multiplier before inclusion"
      expected_events     = "no memo cost, byte fee, or reputation multiplier bypass event appears for rejected memo economic abuse paths"
      expected_error_path = "memo economic abuse rejects before memo metadata, fee, reputation, index, event, or execution state mutation"
      mutation_inputs     = "long memo with zero byte fee, compressed-looking memo underpay, high reputation memo fee bypass, control chars to avoid byte count, split memo fields"
      expected_rejection  = "memo economic abuse must not bypass memo cost, byte fee, or reputation multiplier"
    }
    "INDEX-01" = [ordered]@{
      flow                = "query indexing for tx hash, sender, receiver, domain, contract, memo, event, token, and NFT surfaces"
      state               = "index records, query cursors, surface mappings, receipt links, and rebuild metadata update deterministically from committed state"
      attack              = "valid index lifecycle baseline plus unauthorized index writer control sample"
      invariant           = "index output is deterministic, rebuildable, and non-authoritative over consensus state"
      expected_behavior   = "valid tx hash, sender, receiver, domain, contract, memo, event, token, and NFT indexing records committed state exactly once"
      expected_events     = "index events or rebuild logs match committed event/state source records without changing consensus state"
      expected_error_path = "unauthorized index writer control sample is rejected before index record, cursor, cache, or query state mutation"
      mutation_inputs     = "valid tx hash index, valid sender index, valid receiver index, valid domain index, valid contract index, valid memo index, valid event index, valid token index, valid NFT index, unauthorized index writer"
      expected_rejection  = "unauthorized index lifecycle variants must fail without index record, cursor, cache, or consensus state mutation"
    }
    "INDEX-02" = [ordered]@{
      flow                = "empty result, pagination, duplicate records, deleted state, max query size"
      state               = "index edge cases leave records, cursors, deleted markers, query bounds, and rebuild metadata deterministic"
      attack              = "empty result ambiguity, pagination abuse, duplicate record injection, deleted state resurrection, max query size overflow"
      invariant           = "index queries are bounded, pagination-safe, duplicate-safe, and consistent with deleted consensus state"
      expected_behavior   = "valid index boundaries execute deterministically; invalid boundaries reject before index or query mutation"
      expected_events     = "accepted boundary index events match query/index deltas; rejected edge cases emit no success events"
      expected_error_path = "index validation rejects unsafe pagination, duplicate records, deleted-state resurrection, and max query size overflow"
      mutation_inputs     = "empty result query, pagination page limit plus one, invalid pagination next key, duplicate index record, deleted state lookup, max query size plus one"
      expected_rejection  = "invalid index edge cases must not alter index records, cursors, deleted markers, query bounds, rebuild metadata, or consensus state"
    }
    "INDEX-03" = [ordered]@{
      flow                = "index poisoning, stale resolver lookup, fake event indexing, inconsistent domain cache"
      state               = "adversarial index attempts cannot poison records, serve stale resolver state, index fake events, or diverge domain cache"
      attack              = "index poisoning, stale resolver lookup, fake event indexing, inconsistent domain cache"
      invariant           = "index data must be rebuildable from committed events/state and cannot override canonical state"
      expected_behavior   = "adversarial index mutations fail deterministically before index, cache, query, or rebuild corruption"
      expected_events     = "failed index attacks emit no misleading index, resolver, domain, event, or query success records"
      expected_error_path = "index path rejects poisoning, stale resolver, fake event, or inconsistent domain cache before serving authoritative-looking output"
      mutation_inputs     = "forged index record, stale resolver cache, fake event record, inconsistent domain cache entry, mismatched committed height"
      expected_rejection  = "index attacks must not poison indexes, serve stale resolver lookup, index fake events, or create inconsistent domain cache"
    }
    "INDEX-04" = [ordered]@{
      flow                = "index output never overrides consensus state and can be rebuilt from committed events/state"
      state               = "index records, query output, cursors, caches, and rebuild state reconcile with committed events/state"
      attack              = "state drift attempt through index rebuild, stale cache, deleted state, fake event, and replay/export/import operations"
      invariant           = "index output never overrides consensus state and can be rebuilt from committed events/state"
      expected_behavior   = "index state integrity holds across query, rebuild, cache invalidation, replay, export, and import sequences"
      expected_events     = "index events and rebuild logs reconcile to committed state, receipts, events, and query output"
      expected_error_path = "failed index operations preserve pre-failure index records, cursors, caches, rebuild state, and consensus state references"
      mutation_inputs     = "accepted index then rebuild, fake event then rebuild, stale cache invalidation, export/import after indexed txs, deleted state rebuild"
      expected_rejection  = "rejected index operations must preserve non-authoritative output and committed-state rebuildability"
    }
    "INDEX-05" = [ordered]@{
      flow                = "index economic abuse around priority/search, fund routing, balance changes, and protocol fee bypass"
      state               = "index priority/search cannot route funds, change balances, or bypass protocol fees"
      attack              = "index priority fund routing, search result balance change, protocol fee bypass through index priority"
      invariant           = "index priority/search cannot route funds, change balances, or bypass protocol fees"
      expected_behavior   = "index economic rules keep query priority and search ranking observational and non-consensus-authoritative"
      expected_events     = "no fund route, balance change, fee bypass, or consensus mutation event appears for rejected index economic abuse paths"
      expected_error_path = "index economic abuse rejects before fund routing, balance, fee, query priority, search, or consensus state mutation"
      mutation_inputs     = "paid search priority routing funds, index result as balance proof, fake fee-paid index record, priority query changing route, search rank changing transfer target"
      expected_rejection  = "index economic abuse must not route funds, change balances, or bypass protocol fees"
    }
    "SHARD-01" = [ordered]@{
      flow                = "LOAD_SCORE, zone selection, shard activation, shard assignment, commitment output"
      state               = "load state, route decision, active shard set, shard assignment, and commitment output update deterministically"
      attack              = "valid sharding/load lifecycle baseline plus unauthorized routing input control sample"
      invariant           = "load score, zone route, shard activation, shard assignment, and commitments are deterministic"
      expected_behavior   = "valid LOAD_SCORE update, zone selection, shard activation, shard assignment, and commitment output execute exactly once"
      expected_events     = "sharding simulator events match load, route, shard, assignment, and commitment deltas"
      expected_error_path = "unauthorized routing input control sample is rejected before load, route, shard, or commitment mutation"
      mutation_inputs     = "valid LOAD_SCORE update, valid zone selection, valid shard activation, valid shard assignment, valid commitment output, unauthorized route input"
      expected_rejection  = "unauthorized sharding lifecycle variants must fail without load, route, shard, assignment, or commitment mutation"
    }
    "SHARD-02" = [ordered]@{
      flow                = "zero load, max load, oscillating load, empty shard, max shard count, routing epoch changes"
      state               = "sharding edge cases leave load windows, active shards, route decisions, routing epochs, and commitments deterministic"
      attack              = "zero load boundary, max load boundary, oscillating load, empty shard, max shard count overflow, routing epoch boundary"
      invariant           = "load/routing accepts bounded score inputs, bounded shard counts, deterministic epochs, and stable oscillation handling"
      expected_behavior   = "valid sharding boundaries execute deterministically; invalid boundaries reject before route or shard mutation"
      expected_events     = "accepted boundary sharding events match load/route/shard deltas; rejected edge cases emit no success events"
      expected_error_path = "sharding validation rejects invalid load scores, shard counts, empty shard misuse, and unsafe routing epoch changes"
      mutation_inputs     = "zero load, max load, oscillating load sequence, empty shard, max shard count plus one, routing epoch change, invalid negative load"
      expected_rejection  = "invalid sharding edge cases must not alter load windows, active shards, route decisions, epochs, or commitments"
    }
    "SHARD-03" = [ordered]@{
      flow                = "load poisoning, shard overload targeting, routing loop, route desync, shard starvation"
      state               = "adversarial sharding attempts cannot poison load, target overload, loop routes, desync routes, or starve shards"
      attack              = "load poisoning, shard overload targeting, routing loop, route desync, shard starvation"
      invariant           = "load score bounds, deterministic routing, shard assignment, and starvation prevention cannot be bypassed"
      expected_behavior   = "adversarial sharding mutations fail deterministically before load, route, shard, or commitment corruption"
      expected_events     = "failed sharding attacks emit no misleading load, route, shard, starvation, or commitment success events"
      expected_error_path = "sharding simulator rejects load poisoning, overload targeting, routing loops, route desync, or starvation before commit"
      mutation_inputs     = "poisoned load metric, hot-shard targeting key, routing loop input, divergent route hint, shard starvation sequence"
      expected_rejection  = "sharding attacks must not poison load, target overload, create routing loops, desync routes, or starve shards"
    }
    "SHARD-04" = [ordered]@{
      flow                = "same tx and state produce same route, shard, commitment, and replay output"
      state               = "load state, route decision, shard assignment, commitment output, and replay output do not diverge across replay/export/import"
      attack              = "state drift attempt through same tx/state with different insertion order, route hints, replay, export, and import"
      invariant           = "same tx and state produce same route, same shard, same commitment, and same replay output"
      expected_behavior   = "sharding state integrity holds across load update, route, shard assignment, commitment, replay, export, and import sequences"
      expected_events     = "sharding events reconcile to final load, route, shard, commitment, and replay output deltas"
      expected_error_path = "failed sharding operations preserve pre-failure load, route, shard, commitment, and replay snapshots"
      mutation_inputs     = "same tx with different map insertion order, accepted route followed by replay, same state export/import, same route with ignored hint, commitment recomputation"
      expected_rejection  = "rejected sharding operations must preserve deterministic same-tx same-route same-shard same-commitment consistency"
    }
    "SHARD-05" = [ordered]@{
      flow                = "sharding economic abuse around fee level, reputation, priority, and deterministic protocol routing"
      state               = "fee level, reputation, or priority cannot manipulate routing outside deterministic protocol rules"
      attack              = "fee-level route manipulation, reputation route manipulation, priority route manipulation, deterministic routing bypass"
      invariant           = "fee level, reputation, or priority cannot manipulate routing outside deterministic protocol rules"
      expected_behavior   = "sharding economic rules keep fee, reputation, and priority inputs bounded to deterministic routing policy"
      expected_events     = "no route manipulation, shard manipulation, commitment manipulation, or priority bypass event appears for rejected sharding economic abuse paths"
      expected_error_path = "sharding economic abuse rejects before route, shard, commitment, fee, reputation, or priority state mutation"
      mutation_inputs     = "overpaid fee route hint, forged reputation class, priority above cap, route hint outside epoch, validator-local routing preference"
      expected_rejection  = "sharding economic abuse must not let fee level, reputation, or priority manipulate routing outside deterministic protocol rules"
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

function Get-AexsInvariantChecklistPrefix {
  param([string]$Category)
  switch -Regex ($Category) {
    '^Economic Invariants$' { return "ECON" }
    '^Consensus And State Invariants$' { return "STATE" }
    '^DEX Invariants$' { return "DEXINV" }
    '^Load, Routing, And Sharding Invariants$' { return "LOAD" }
    '^Identity And Resolver Invariants$' { return "IDINV" }
    '^Execution, AVM, And Queue Invariants$' { return "EXECINV" }
    default {
      $words = @([regex]::Matches($Category.ToUpperInvariant(), '[A-Z0-9]+') | ForEach-Object { $_.Value })
      if ($words.Count -eq 0) {
        return "INV"
      }
      return (($words | Select-Object -First 2) -join "")
    }
  }
}

function Get-AexsInvariantChecklistScope {
  param([string]$Category)
  switch -Regex ($Category) {
    '^Economic Invariants$' { return "global/economics" }
    '^Consensus And State Invariants$' { return "app/consensus-state" }
    '^DEX Invariants$' { return "x/dex" }
    '^Load, Routing, And Sharding Invariants$' { return "x/sharding/sim" }
    '^Identity And Resolver Invariants$' { return "x/identity" }
    '^Execution, AVM, And Queue Invariants$' { return "x/vm+x/queue" }
    default { return "global/invariants" }
  }
}

function Get-AexsInvariantChecklistOverride {
  param([string]$InvariantId)
  $overrides = @{
    "ECON-01" = [ordered]@{
      flow                = "global supply equation across account balances, module balances, burned accounting, and staked accounting"
      state               = "accepted value-moving transitions preserve the equation sum(account balances) + module balances + burned + staked = total_supply"
      attack              = "supply drift through mixed bank, module, burn, staking, mint, export, and import transitions"
      expected_behavior   = "every executed block recomputes the global supply equation from canonical stores and records any mismatch as a critical invariant break"
      expected_events     = "no supply drift, unauthorized mint, unauthorized burn, or accounting mismatch event appears after accepted value transitions"
      expected_error_path = "malformed or unauthorized value transition rejects before balance, module account, burned, staked, or supply mutation"
      mutation_inputs     = "unauthorized mint, unauthorized burn, partial bank send failure, staking loop, distribution withdrawal loop, export/import replay"
      expected_rejection  = "any transition that would make the supply equation false is rejected or reported as a critical invariant failure"
    }
    "ECON-02" = [ordered]@{
      flow                = "balance non-negativity across account and module balances"
      state               = "accepted transfers, burns, fees, staking, and DEX operations cannot create negative balances"
      attack              = "underflow, oversized send, partial multi-send, fee drain, burn overflow, and insufficient-funds mutation"
      expected_behavior   = "bank and module accounting reject any transition that would underflow an account or module balance"
      expected_events     = "no negative-balance event or committed negative balance appears in exported state"
      expected_error_path = "underflow path rejects before balance mutation"
      mutation_inputs     = "max amount send, insufficient funds send, duplicated output multi-send, oversized burn, fee larger than balance"
      expected_rejection  = "negative balance attempts must fail without mutating state"
    }
    "ECON-03" = [ordered]@{
      flow                = "mint authority boundaries for native, module, and tokenfactory supply"
      state               = "only explicitly authorized mint paths can increase supply"
      attack              = "unauthorized mint, module account spoofing, tokenfactory admin bypass, governance authority spoofing"
      expected_behavior   = "minting requires the module-specific authority and denom-specific permission before any supply increase"
      expected_events     = "no mint event is emitted from an unauthorized signer or unauthorized module account"
      expected_error_path = "unauthorized mint rejects before bank supply or denom metadata mutation"
      mutation_inputs     = "wrong admin, zero admin, wrong module account, forged authority, native denom mint through tokenfactory"
      expected_rejection  = "unauthorized mint attempts cannot increase supply"
    }
    "ECON-04" = [ordered]@{
      flow                = "burn authority boundaries for user, module, and tokenfactory supply"
      state               = "only explicitly authorized burn paths can decrease supply or burn from a source account"
      attack              = "unauthorized burn, burn-from mismatch, module account burn spoofing, native-denom burn abuse"
      expected_behavior   = "burning requires owner/module/admin authorization and exact source-account control"
      expected_events     = "no burn event is emitted for an unauthorized signer, wrong source, or wrong denom authority"
      expected_error_path = "unauthorized burn rejects before balance, supply, or metadata mutation"
      mutation_inputs     = "wrong admin burn, burn from another account, zero admin, oversized burn, native denom burn through factory path"
      expected_rejection  = "unauthorized burn attempts cannot decrease supply or debit another account"
    }
    "ECON-05" = [ordered]@{
      flow                = "protocol fee denom admission policy"
      state               = "ante fee checks accept no fee denom other than naet"
      attack              = "fee denom spoofing, multi-denom fee bypass, tokenfactory asset fee payment, missing fee bypass"
      expected_behavior   = "the fee policy rejects every transaction whose protocol fee set contains a non-naet denom"
      expected_events     = "no fee-collected event exists for non-naet denoms"
      expected_error_path = "non-naet fee rejects in ante before message execution"
      mutation_inputs     = "non-naet fee, mixed naet/factory fee, malformed denom, missing fee, zero fee where min fee applies"
      expected_rejection  = "non-naet protocol fee paths cannot enter execution"
    }
    "ECON-06" = [ordered]@{
      flow                = "fee collection and distribution accounting"
      state               = "fee distribution totals match collected fees for treasury, burn, validators, and community pool"
      attack              = "rounding loss, split overflow, malformed tx shape, duplicate fee deduction, distribution module desync"
      expected_behavior   = "collected fees are allocated exactly once according to bounded deterministic split params"
      expected_events     = "fee collected and fee distributed events reconcile to the same total amount"
      expected_error_path = "invalid fee split or accounting mismatch rejects before distribution state mutation"
      mutation_inputs     = "tiny fee, max fee, split edge percentages, multi-message tx, failed message after ante, export/import"
      expected_rejection  = "fee distribution cannot create, lose, or double-count collected fees"
    }
    "ECON-07" = [ordered]@{
      flow                = "treasury, burn, validator reward, and community pool deterministic accounting"
      state               = "economic sinks and reward destinations are deterministic for the same block input"
      attack              = "map-order distribution drift, rounding nondeterminism, validator order manipulation, treasury routing spoof"
      expected_behavior   = "destination ordering and rounding rules are canonical and produce identical accounting on replay"
      expected_events     = "reward, burn, treasury, and community pool events are stable for same input"
      expected_error_path = "invalid destination or unbounded split rejects before accounting mutation"
      mutation_inputs     = "different validator insertion order, tiny rewards, max rewards, malformed treasury address, replay/export/import"
      expected_rejection  = "economic destination accounting cannot diverge across nodes"
    }
    "ECON-08" = [ordered]@{
      flow                = "staking reward loop resistance"
      state               = "staking rewards cannot be farmed through repeated delegation, unbonding, redelegation, withdrawal, or export/import loops"
      attack              = "reward farming loop, timing manipulation, repeated withdrawal, delegation-share rounding exploit"
      expected_behavior   = "reward accrual depends only on canonical stake, time/height rules, and distribution indexes"
      expected_events     = "no repeated withdrawal or stake-loop event pays more than accrued rewards"
      expected_error_path = "invalid reward-loop sequence rejects before reward or staking state mutation"
      mutation_inputs     = "delegate/withdraw loop, redelegate/withdraw loop, unbond/rebond loop, tiny share rounding, export/import replay"
      expected_rejection  = "staking reward loops cannot mint or redirect extra rewards"
    }
    "ECON-09" = [ordered]@{
      flow                = "supply export/import round-trip stability"
      state               = "supply cannot drift after deterministic export/import"
      attack              = "export ordering drift, missing module account, duplicate balance, corrupted genesis, denom metadata mismatch"
      expected_behavior   = "exported state imports to the same supply, balances, staking, burned accounting, and module balances"
      expected_events     = "no supply-changing event appears during pure export/import validation"
      expected_error_path = "corrupted export/import state rejects during genesis validation before chain start"
      mutation_inputs     = "duplicate balance, duplicate module account, missing staking pool, malformed supply, reordered denoms"
      expected_rejection  = "export/import cannot change total supply or accepted accounting"
    }
    "STATE-01" = [ordered]@{
      flow                = "same block input app-hash determinism"
      state               = "same block input produces same app hash"
      attack              = "map iteration drift, random source, wall-clock dependency, goroutine race, platform-dependent integer behavior"
      expected_behavior   = "FinalizeBlock/Commit over identical genesis and tx sequence yields identical AppHash"
      expected_events     = "block events and app hash are identical for same ordered input"
      expected_error_path = "nondeterministic state transition is recorded as a critical consensus failure"
      mutation_inputs     = "same tx list with different local map insertion order, replayed genesis, repeated block execution, restart replay"
      expected_rejection  = "nondeterministic app hash is never accepted as production-safe"
    }
    "STATE-02" = [ordered]@{
      flow                = "signed transaction replay protection"
      state               = "same signed transaction cannot execute twice"
      attack              = "exact signed byte replay, sequence reuse, mempool rebroadcast, export/import replay"
      expected_behavior   = "accepted transaction increments sequence and any byte-for-byte replay fails before message execution"
      expected_events     = "second execution emits no state-mutating success event"
      expected_error_path = "replayed tx rejects in ante before account, balance, fee, or module state mutation"
      mutation_inputs     = "same signed tx bytes twice, same sequence with altered memo, same tx after restart, wrong chain id"
      expected_rejection  = "replayed signed transaction cannot mutate state twice"
    }
    "STATE-03" = [ordered]@{
      flow                = "invalid signer state mutation prevention"
      state               = "invalid signer cannot mutate state"
      attack              = "wrong signer, missing signer, malformed public key, zero address signer, unauthorized authority"
      expected_behavior   = "signature and signer checks complete before any message state transition"
      expected_events     = "no module success event is emitted for invalid signer paths"
      expected_error_path = "invalid signer rejects in ante or msg ValidateBasic before state mutation"
      mutation_inputs     = "invalid signature, zero signer, wrong admin, wrong authority, malformed public key"
      expected_rejection  = "invalid signer paths leave all module state unchanged"
    }
    "STATE-04" = [ordered]@{
      flow                = "malformed transaction pre-mutation failure"
      state               = "malformed transaction fails before state mutation"
      attack              = "malformed protobuf, invalid memo bytes, malformed auth info, malformed message fields, oversized tx"
      expected_behavior   = "decode, basic validation, ante, and message validation reject malformed tx before writes"
      expected_events     = "malformed tx does not emit successful module events"
      expected_error_path = "malformed transaction rejects before cache context write commit"
      mutation_inputs     = "corrupted protobuf bytes, invalid UTF-8 memo, missing auth info, oversized payload, malformed address"
      expected_rejection  = "malformed transaction cannot partially mutate state"
    }
    "STATE-05" = [ordered]@{
      flow                = "validator set and staking keeper consistency"
      state               = "validator set matches staking keeper state"
      attack              = "validator power spoofing, stale validator update, missing slash update, invalid gentx, export/import mismatch"
      expected_behavior   = "EndBlock validator updates derive only from canonical staking keeper state"
      expected_events     = "validator-set update events match staking keeper validator power and status"
      expected_error_path = "invalid validator-set update source rejects or fails deterministic invariant check"
      mutation_inputs     = "validator create/remove, delegation power change, slash, jail, unbonding, exported validator state reorder"
      expected_rejection  = "CometBFT validator updates cannot diverge from staking keeper state"
    }
    "STATE-06" = [ordered]@{
      flow                = "objective deterministic slashing evidence"
      state               = "slashing evidence is objective and deterministic"
      attack              = "stale evidence, duplicate evidence, malformed proof, redelegation timing evasion, unbonding timing evasion"
      expected_behavior   = "the same valid evidence produces the same slash/jail/tombstone transition on every node"
      expected_events     = "slash, jail, and tombstone events are identical for the same objective evidence"
      expected_error_path = "invalid or duplicate evidence rejects before slashing state mutation"
      mutation_inputs     = "duplicate evidence, stale height, malformed validator address, invalid signature proof, evidence after unbonding"
      expected_rejection  = "subjective or malformed evidence cannot mutate slashing state"
    }
    "STATE-07" = [ordered]@{
      flow                = "genesis validation for malformed state"
      state               = "genesis validation rejects malformed accounts, balances, params, and module state"
      attack              = "duplicate accounts, duplicate balances, invalid params, malformed custom module genesis, invalid denom metadata"
      expected_behavior   = "ValidateGenesis rejects malformed state before InitChain can commit it"
      expected_events     = "no startup success or state migration event exists for rejected genesis"
      expected_error_path = "malformed genesis fails startup-only validation with explicit error"
      mutation_inputs     = "duplicate account, negative/overflow balance, invalid naet metadata, invalid params, duplicate custom module ids"
      expected_rejection  = "malformed genesis cannot start a chain"
    }
    "STATE-08" = [ordered]@{
      flow                = "upgrade and migration state preservation"
      state               = "upgrade and migration paths preserve state roots and module invariants"
      attack              = "migration ordering drift, missing module version, panic path, duplicate migration, corrupted version map"
      expected_behavior   = "upgrade handlers and migrations are deterministic, version-gated, and preserve declared module invariants"
      expected_events     = "migration events and post-migration roots are stable for the same pre-upgrade state"
      expected_error_path = "invalid migration state rejects before committing partial upgraded state"
      mutation_inputs     = "future version no-op migration, corrupted version map, reordered module state, malformed export, restart after migration"
      expected_rejection  = "migration cannot alter roots or invariants outside explicit migration contract"
    }
    "DEXINV-01" = [ordered]@{
      flow                = "DEX reserve and module account balance reconciliation"
      state               = "Pool reserves match module account balances"
      attack              = "reserve desync through failed bank send, direct module account drift, corrupted pool reserves, swap/add/remove partial update"
      expected_behavior   = "every accepted DEX transition reconciles pool reserve fields against module account balances for both pool denoms"
      expected_events     = "pool reserve, add liquidity, remove liquidity, and swap events reconcile to module account balance deltas"
      expected_error_path = "reserve mismatch rejects or records a critical invariant failure before pool state is accepted"
      mutation_inputs     = "corrupted reserve field, missing bank movement, partial add liquidity, partial swap, export/import with mismatched module balances"
      expected_rejection  = "pool state cannot commit when reserves do not match module account balances"
    }
    "DEXINV-02" = [ordered]@{
      flow                = "DEX LP supply and pool total share reconciliation"
      state               = "LP supply matches pool total shares"
      attack              = "LP inflation, forged LP mint, missing LP burn, corrupted pool total shares, export/import LP supply drift"
      expected_behavior   = "LP token total supply and pool total shares move by the same deterministic delta on add/remove liquidity"
      expected_events     = "LP mint and burn events reconcile exactly to pool total share changes"
      expected_error_path = "LP supply/share mismatch rejects before liquidity state mutation is committed"
      mutation_inputs     = "oversized LP mint, wrong LP denom, remove liquidity without burn, corrupted total shares, duplicate pool id export"
      expected_rejection  = "LP supply cannot diverge from pool total shares"
    }
    "DEXINV-03" = [ordered]@{
      flow                = "DEX liquidity non-negativity across reserves and shares"
      state               = "No negative liquidity"
      attack              = "underflow during remove liquidity, zero-liquidity pool manipulation, oversized withdrawal, signed integer conversion"
      expected_behavior   = "reserves, total shares, and user LP burns are checked before arithmetic that could underflow"
      expected_events     = "no negative liquidity, negative reserve, or negative share event is emitted"
      expected_error_path = "negative-liquidity path rejects before bank movement or pool mutation"
      mutation_inputs     = "remove more shares than supply, zero reserve swap, max amount subtraction, tiny liquidity rounding edge"
      expected_rejection  = "DEX arithmetic cannot produce negative reserves, shares, or output liquidity"
    }
    "DEXINV-04" = [ordered]@{
      flow                = "DEX LP denom authenticity"
      state               = "No fake LP token"
      attack              = "wrong LP denom burn, forged LP token, pool id collision, tokenfactory spoofed lp/{pool_id} denom"
      expected_behavior   = "remove liquidity accepts only the canonical LP denom derived from the exact pool id"
      expected_events     = "LP burn events always reference the canonical pool LP denom"
      expected_error_path = "fake LP token rejects before reserves, shares, or module balances mutate"
      mutation_inputs     = "wrong lp denom, duplicate pool id, forged lp/{pool_id}, non-LP denom, tokenfactory-created LP-like denom"
      expected_rejection  = "fake LP tokens cannot redeem reserves or mutate pool state"
    }
    "DEXINV-05" = [ordered]@{
      flow                = "DEX swap output non-negativity"
      state               = "Swap output is non-negative"
      attack              = "rounding underflow, fee-adjusted input underflow, tiny reserve edge, max amount overflow, zero-input swap"
      expected_behavior   = "swap math rejects zero/invalid inputs and computes output with non-negative bounded integer arithmetic"
      expected_events     = "swap events never contain negative output or impossible reserve delta"
      expected_error_path = "invalid swap output rejects before bank send or reserve mutation"
      mutation_inputs     = "zero amount in, max amount in, tiny reserves, fee rounding edge, corrupted reserve signs"
      expected_rejection  = "swap cannot produce negative output or commit impossible reserve deltas"
    }
    "DEXINV-06" = [ordered]@{
      flow                = "DEX slippage bound enforcement"
      state               = "Slippage bounds are enforced"
      attack              = "slippage bypass, stale quote replay, min-output off-by-one, rounded output below bound, route manipulation"
      expected_behavior   = "swap execution compares deterministic output against user min-output before any state mutation"
      expected_events     = "no successful swap event appears when output is below the requested slippage bound"
      expected_error_path = "slippage failure rejects before reserve, LP, or bank state mutation"
      mutation_inputs     = "min output one above computed amount, stale quote, tiny reserves, rounded fee output, replayed route"
      expected_rejection  = "swap below slippage bound cannot execute or move funds"
    }
    "DEXINV-07" = [ordered]@{
      flow                = "DEX fee-adjusted constant-product preservation"
      state               = "Constant-product constraints hold after fee-adjusted swaps"
      attack              = "constant-product break, rounding leak, fee bypass, reserve ordering bug, repeated tiny swap drain"
      expected_behavior   = "post-swap reserves satisfy the configured fee-adjusted constant-product inequality"
      expected_events     = "swap events preserve or improve the expected fee-adjusted invariant within deterministic rounding rules"
      expected_error_path = "constant-product violation rejects before reserve or bank state mutation"
      mutation_inputs     = "tiny repeated swaps, extreme reserve ratio, max input, zero fee edge, denom order reversal"
      expected_rejection  = "fee-adjusted swaps cannot reduce reserves below the constant-product constraint"
    }
    "DEXINV-08" = [ordered]@{
      flow                = "DEX atomic bank movement and pool state rollback"
      state               = "Failed bank movement cannot mutate pool state"
      attack              = "failed send partial update, insufficient funds after reserve write, module account error, LP mint failure after reserve mutation"
      expected_behavior   = "DEX message handlers commit pool state only after all required bank movements succeed in the cached context"
      expected_events     = "failed bank movement emits no successful pool, reserve, swap, or LP event"
      expected_error_path = "bank movement failure rolls back reserve, share, LP supply, and module balance changes"
      mutation_inputs     = "insufficient funds, blocked module account send, bad denom, LP mint failure, bank send panic surrogate in simulator"
      expected_rejection  = "failed bank movement cannot leave partially mutated pool state"
    }
    "LOAD-01" = [ordered]@{
      flow                = "LOAD_SCORE range enforcement"
      state               = '`LOAD_SCORE` is always in `[0,1]`'
      attack              = "out-of-range metric injection, overflowed metric normalization, negative metric value, corrupted genesis params"
      expected_behavior   = "load calculation clamps or rejects metric inputs so the final score remains in the protocol range"
      expected_events     = "load update events record only values between 0 and 1"
      expected_error_path = "invalid metric or param rejects before load state mutation"
      mutation_inputs     = "negative mempool score, block utilization above one, max integer execution time, invalid weight sum, corrupted params"
      expected_rejection  = "LOAD_SCORE outside [0,1] cannot be committed"
    }
    "LOAD-02" = [ordered]@{
      flow                = "deterministic EMA smoothing"
      state               = "EMA smoothing is deterministic"
      attack              = "node-local latency input, floating-point drift, map-order metric aggregation, wall-clock metric sampling"
      expected_behavior   = "EMA uses protocol constants and deterministic fixed-point or canonical arithmetic for identical inputs"
      expected_events     = "same input window produces identical EMA and load update events on replay"
      expected_error_path = "non-deterministic metric source is rejected from consensus load state"
      mutation_inputs     = "same metrics in different map insertion order, repeated replay, platform variation, node-local latency sample"
      expected_rejection  = "EMA smoothing cannot depend on local time, local latency, random values, or unordered iteration"
    }
    "LOAD-03" = [ordered]@{
      flow                = "LOAD_SCORE spike cap enforcement"
      state               = '`LOAD_SCORE` cannot jump beyond `MAX_DELTA`'
      attack              = "spam burst, gas spike, execution-time spike, oscillating load poisoning, EMA slow poison"
      expected_behavior   = "per-block score update applies the MAX_DELTA cap after deterministic EMA calculation"
      expected_events     = "load update event delta never exceeds MAX_DELTA"
      expected_error_path = "uncapped load jump is rejected or reduced before state mutation"
      mutation_inputs     = "zero-to-one load burst, alternating max/min metrics, max mempool spam, saturated gas block"
      expected_rejection  = "LOAD_SCORE cannot move by more than MAX_DELTA in one block"
    }
    "LOAD-04" = [ordered]@{
      flow                = "deterministic zone routing decision"
      state               = "Same transaction and state produce the same zone decision"
      attack              = "routing hint manipulation, message classification ambiguity, validator preference injection, map-order tx metadata"
      expected_behavior   = "zone selection is a pure function of stable tx class, protocol state, and deterministic load state"
      expected_events     = "same transaction and state emit the same zone route decision on replay"
      expected_error_path = "ambiguous or unknown tx class rejects before route state is accepted"
      mutation_inputs     = "unknown msg type, conflicting routing hint, same tx with different local metadata order, validator-local preference"
      expected_rejection  = "same tx/state cannot route to different zones"
    }
    "LOAD-05" = [ordered]@{
      flow                = "deterministic shard routing decision"
      state               = "Same transaction and state produce the same shard decision"
      attack              = "routing epoch manipulation, active shard count mismatch, primary actor ambiguity, hash tie-break manipulation"
      expected_behavior   = "shard selection uses canonical zone id, primary actor, routing epoch, active shard count, and deterministic hash"
      expected_events     = "same transaction and state emit the same shard id on replay"
      expected_error_path = "missing primary actor or zero active shards rejects before shard decision is committed"
      mutation_inputs     = "empty primary actor, zero active shards, changed local shard order, same tx across routing epoch boundary"
      expected_rejection  = "same tx/state cannot route to different shards"
    }
    "LOAD-06" = [ordered]@{
      flow                = "routing loop prevention"
      state               = "No routing loop"
      attack              = "cross-zone route cycle, async message bounce loop, self-routing route hint, recursive message forwarding"
      expected_behavior   = "routing records destination once per routing step and bounded message forwarding prevents cyclic route execution"
      expected_events     = "route and message events contain bounded acyclic routing path markers"
      expected_error_path = "looping route or exceeded route depth rejects before queue or route mutation"
      mutation_inputs     = "destination equals source with forwarding, A->B->A route sequence, malformed async route, bounce recursion"
      expected_rejection  = "routing cannot create an infinite loop or unbounded message cycle"
    }
    "LOAD-07" = [ordered]@{
      flow                = "shard starvation prevention"
      state               = "No shard starvation"
      attack              = "priority queue starvation, high-fee monopolization, skewed shard assignment, delayed low-priority queue never drains"
      expected_behavior   = "deterministic priority and scheduling rules preserve bounded progress for valid queued shard work"
      expected_events     = "scheduler and shard events show bounded deferral instead of permanent starvation"
      expected_error_path = "priority or load policy that would permanently starve a shard fails invariant checks"
      mutation_inputs     = "continuous high-priority spam, skewed primary actor set, repeated routing epoch changes, overloaded single shard"
      expected_rejection  = "valid shard work cannot be starved indefinitely by fee, priority, or routing manipulation"
    }
    "LOAD-08" = [ordered]@{
      flow                = "hot-zone monopolization resistance"
      state               = "No hot-zone monopolization"
      attack              = "single-zone spam monopoly, load poisoning to force full sharding, fee gaming to crowd out other zones, hot actor targeting"
      expected_behavior   = "load-driven routing and shard activation isolate hot zones without consuming unrelated zone execution capacity"
      expected_events     = "hot-zone load events do not suppress unrelated zone route and shard progress events"
      expected_error_path = "cross-zone resource monopolization fails load/routing invariant checks"
      mutation_inputs     = "Financial Zone spam burst, contract hot account spam, high-fee single-zone flood, oscillating load poison"
      expected_rejection  = "one hot zone cannot monopolize all zones or unrelated shard execution"
    }
    "LOAD-09" = [ordered]@{
      flow                = "deterministic priority ordering across nodes"
      state               = "No priority ordering divergence across nodes"
      attack              = "priority tie-break drift, fee-class overflow, reputation over-cap, tx hash tie-break mismatch, local mempool order dependency"
      expected_behavior   = "priority ordering uses canonical priority class, bounded fee class, bounded reputation, admission height, and tx hash"
      expected_events     = "same admitted tx set produces identical priority order and route events on every node"
      expected_error_path = "unbounded fee/reputation class or missing tie-break rejects before priority queue admission"
      mutation_inputs     = "same tx set in different local order, fee overpayment, reputation above cap, identical priority keys except tx hash"
      expected_rejection  = "priority ordering cannot diverge across nodes due to local ordering or unbounded classes"
    }
  }
  if ($overrides.ContainsKey($InvariantId)) {
    return $overrides[$InvariantId]
  }
  return $null
}

function Get-AexsInvariantChecklistRecords {
  param([string]$Text, [string]$CampaignId)
  $section = Get-AexsMarkdownSection -Text $Text -Heading "Invariant Checklist"
  if ([string]::IsNullOrWhiteSpace($section)) {
    return @()
  }

  $items = @()
  $category = ""
  $current = $null
  foreach ($line in ($section -split "`r?`n")) {
    if ($line -match '^###\s+(.+?)\s*$') {
      if ($null -ne $current) {
        $items += $current
        $current = $null
      }
      $category = $Matches[1].Trim()
      continue
    }

    if ($line -match '^- \[ \]\s+(.+?)\s*$') {
      if ($null -ne $current) {
        $items += $current
      }
      $current = [ordered]@{
        category    = $category
        description = $Matches[1].Trim()
      }
      continue
    }

    if ($null -ne $current -and $line -match '^\s+(.+?)\s*$') {
      $current["description"] = ($current["description"] + " " + $Matches[1].Trim()).Trim()
    }
  }
  if ($null -ne $current) {
    $items += $current
  }

  $categoryCounts = @{}
  $records = @()
  foreach ($item in $items) {
    $category = [string]$item["category"]
    if ([string]::IsNullOrWhiteSpace($category)) {
      $category = "Uncategorized Invariants"
    }
    $prefix = Get-AexsInvariantChecklistPrefix -Category $category
    if (-not $categoryCounts.ContainsKey($prefix)) {
      $categoryCounts[$prefix] = 0
    }
    $categoryCounts[$prefix] = [int]$categoryCounts[$prefix] + 1
    $invariantId = "{0}-{1:00}" -f $prefix, [int]$categoryCounts[$prefix]
    $invariantText = ([string]$item["description"]).Trim().TrimEnd(".")
    $scope = Get-AexsInvariantChecklistScope -Category $category
    $override = Get-AexsInvariantChecklistOverride -InvariantId $invariantId
    $seedHash = (Get-AexsSha256Hex -Text "$CampaignId|invariant-checklist|$invariantId|$invariantText").Substring(0, 16)
    $seed = "aexs-$($invariantId.ToLowerInvariant())-$seedHash"
    $flow = Get-AexsOverrideValue -Override $override -Field "flow" -Fallback "$category mandatory invariant check"
    $state = Get-AexsOverrideValue -Override $override -Field "state" -Fallback $invariantText
    $attack = Get-AexsOverrideValue -Override $override -Field "attack" -Fallback "malformed input, replay, unauthorized signer, boundary values, state corruption, or module-specific exploit against $category"
    $expectedBehavior = Get-AexsOverrideValue -Override $override -Field "expected_behavior" -Fallback "campaign executes this invariant after relevant simulated transitions"
    $expectedEvents = Get-AexsOverrideValue -Override $override -Field "expected_events" -Fallback "events remain consistent with committed state or no events are emitted on rejection"
    $expectedErrorPath = Get-AexsOverrideValue -Override $override -Field "expected_error_path" -Fallback "invalid transition fails before committing state that violates the invariant"
    $mutationInputs = Get-AexsOverrideValue -Override $override -Field "mutation_inputs" -Fallback "generated scenario seed, mutated tx bytes, corrupted genesis fragment, replayed sequence, and boundary values"
    $expectedRejection = Get-AexsOverrideValue -Override $override -Field "expected_rejection" -Fallback "any invariant violation is rejected or recorded as a critical audit failure"

    $records += [ordered]@{
      module                         = $scope
      category                       = $category
      invariant_id                   = $invariantId
      task_id                        = $invariantId
      function_or_flow_covered       = $flow
      state_transition_covered       = $state
      attack_surface_covered         = $attack
      invariant_tested               = $invariantText
      defensive_analysis_result      = [ordered]@{
        status                    = "planned_not_executed"
        expected_behavior         = $expectedBehavior
        expected_state_transition = $state
        expected_events           = $expectedEvents
        expected_error_path       = $expectedErrorPath
        expected_invariant        = $invariantText
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
          "Run AEXS mandatory invariant check $invariantId",
          "Use seed $seed",
          "Execute defensive state-transition replay",
          "Execute adversarial mutation simulation",
          "Record pass_fail_result only after executed evidence exists"
        )
      }
      valid                          = $true
      invalid_reasons                = @()
    }
  }
  return $records
}

function Test-AexsInvariantChecklistRecord {
  param([object]$Record)
  $reasons = @()
  foreach ($field in @(
      "module",
      "category",
      "invariant_id",
      "function_or_flow_covered",
      "state_transition_covered",
      "attack_surface_covered",
      "invariant_tested",
      "pass_fail_result"
    )) {
    if ([string]::IsNullOrWhiteSpace([string]$Record[$field])) {
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
    if ([string]::IsNullOrWhiteSpace([string]$Record["defensive_analysis_result"][$field])) {
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
    if ([string]::IsNullOrWhiteSpace([string]$Record["adversarial_simulation_result"][$field])) {
      $reasons += "missing adversarial_simulation_result.$field"
    }
  }
  if ($null -eq $Record["reproduction_seed_or_steps"] -or [string]::IsNullOrWhiteSpace([string]$Record["reproduction_seed_or_steps"]["seed"])) {
    $reasons += "missing reproduction seed"
  }
  if ($null -eq $Record["reproduction_seed_or_steps"] -or @($Record["reproduction_seed_or_steps"]["steps"]).Count -eq 0) {
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
$invariantChecklistSection = Get-AexsMarkdownSection -Text $taskText -Heading "Invariant Checklist"

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

$requiredInvariantChecklistTerms = @(
  "Economic Invariants",
  "sum(account balances) + module balances + burned + staked = total_supply",
  'No fee denom other than `naet` is accepted.',
  "Fee distribution totals match collected fees.",
  "Staking rewards cannot be farmed through state loops.",
  "Supply cannot drift after export/import.",
  "Consensus And State Invariants",
  "Same block input produces same app hash.",
  "Same signed transaction cannot execute twice.",
  "Invalid signer cannot mutate state.",
  "Malformed transaction fails before state mutation.",
  "Validator set matches staking keeper state.",
  "Slashing evidence is objective and deterministic.",
  "Genesis validation rejects malformed accounts, balances, params, and",
  "Upgrade and migration paths preserve state roots and module invariants.",
  "DEX Invariants",
  "Pool reserves match module account balances.",
  "LP supply matches pool total shares.",
  "No negative liquidity.",
  "No fake LP token.",
  "Swap output is non-negative.",
  "Slippage bounds are enforced.",
  "Constant-product constraints hold after fee-adjusted swaps.",
  "Failed bank movement cannot mutate pool state.",
  "Load, Routing, And Sharding Invariants",
  '`LOAD_SCORE` is always in `[0,1]`.',
  'EMA smoothing is deterministic.',
  '`LOAD_SCORE` cannot jump beyond `MAX_DELTA`.',
  "Same transaction and state produce the same zone decision.",
  "Same transaction and state produce the same shard decision.",
  "No routing loop.",
  "No shard starvation.",
  "No hot-zone monopolization.",
  "No priority ordering divergence across nodes."
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
if ([string]::IsNullOrWhiteSpace($invariantChecklistSection)) {
  $sourceFailures += "missing Invariant Checklist section"
} else {
  foreach ($term in @(Get-AexsMissingTerms -Text $invariantChecklistSection -Terms $requiredInvariantChecklistTerms)) {
    $sourceFailures += "missing invariant checklist term: $term"
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
$invariantChecklistRecords = @(Get-AexsInvariantChecklistRecords -Text $taskText -CampaignId $campaignId)
foreach ($record in $invariantChecklistRecords) {
  $invalidReasons = @(Test-AexsInvariantChecklistRecord -Record $record)
  $record["valid"] = $invalidReasons.Count -eq 0
  $record["invalid_reasons"] = $invalidReasons
}
$invalidInvariantChecklistRecords = @($invariantChecklistRecords | Where-Object { -not $_["valid"] })
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
  invariant_checklist_count           = $invariantChecklistRecords.Count
  invalid_invariant_checklist_count   = $invalidInvariantChecklistRecords.Count
  invariant_checklist_ids             = @($invariantChecklistRecords | ForEach-Object { $_["invariant_id"] })
  invariant_checklist_categories      = @($invariantChecklistRecords | ForEach-Object { $_["category"] } | Sort-Object -Unique)
  invalid_invariant_checklist_records = @($invalidInvariantChecklistRecords | ForEach-Object { $_["invariant_id"] })
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
$invariantChecklistPath = Join-Path $campaignDir "invariant-checklist.json"
$invariantChecklistMarkdownPath = Join-Path $campaignDir "invariant-checklist.md"
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
$invariantChecklistRecords | ConvertTo-Json -Depth 10 | Set-Content -LiteralPath $invariantChecklistPath
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

$invariantReport = @()
$invariantReport += "# AEXS Mandatory Invariant Checklist"
$invariantReport += ""
$invariantReport += "- campaign id: $campaignId"
$invariantReport += "- invariant count: $($invariantChecklistRecords.Count)"
$invariantReport += "- invalid invariant count: $($invalidInvariantChecklistRecords.Count)"
$invariantReport += "- status: planned_not_executed"
$invariantReport += ""
$invariantReport += "| Invariant | Category | Scope | State transition | Attack surface | Defensive status | Adversarial status | Result | Seed |"
$invariantReport += "| --- | --- | --- | --- | --- | --- | --- | --- | --- |"
foreach ($record in $invariantChecklistRecords) {
  $state = ([string]$record["state_transition_covered"]).Replace("|", "/")
  $attack = ([string]$record["attack_surface_covered"]).Replace("|", "/")
  $seed = [string]$record["reproduction_seed_or_steps"]["seed"]
  $invariantReport += "| $($record["invariant_id"]) | $($record["category"]) | $($record["module"]) | $state | $attack | $($record["defensive_analysis_result"]["status"]) | $($record["adversarial_simulation_result"]["status"]) | $($record["pass_fail_result"]) | $seed |"
}
$invariantReport | Set-Content -LiteralPath $invariantChecklistMarkdownPath

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
$report += "- mandatory invariant checklist records: $($invariantChecklistRecords.Count)"
$report += "- invalid mandatory invariant checklist records: $($invalidInvariantChecklistRecords.Count)"
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
$report += "- invalid mandatory invariant checklist records: $(@($invalidInvariantChecklistRecords | ForEach-Object { $_["invariant_id"] }) -join ', ')"
$report += ""
$report += "## Module Matrix"
$report += ""
$report += "| Module | Tasks | Atomic records | Planned coverage | Invariant evidence | Fuzz evidence | Adversarial evidence | Safe |"
$report += "| --- | ---: | ---: | ---: | --- | --- | --- | --- |"
foreach ($row in $moduleRows) {
$report += "| $($row["module"]) | $($row["task_count"]) | $($row["atomic_task_records"]) | $($row["planned_coverage_percent"])% | $($row["has_invariant_evidence"]) | $($row["has_fuzz_evidence"]) | $($row["has_adversarial_evidence"]) | $($row["safe"]) |"
}
$report += ""
$report += "## Mandatory Invariant Checklist"
$report += ""
$report += "| Invariant | Category | Scope | Result |"
$report += "| --- | --- | --- | --- |"
foreach ($record in $invariantChecklistRecords) {
  $report += "| $($record["invariant_id"]) | $($record["category"]) | $($record["module"]) | $($record["pass_fail_result"]) |"
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
if ($invariantChecklistRecords.Count -lt 17) {
  throw "AEXS invariant checklist validation failed: fewer than required economic and consensus invariant records"
}
if ($invalidInvariantChecklistRecords.Count -gt 0) {
  throw "AEXS invariant checklist validation failed for record(s): $(@($invalidInvariantChecklistRecords | ForEach-Object { $_["invariant_id"] }) -join ', ')"
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
