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
