param(
  [string]$OutputDir = ".work\aexs",
  [string]$TaskFile = "TO_AUDIT.md",
  [string]$PipelineDoc = "docs\security\aetra-fuzzing-invariant-pipeline.md",
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
      mutation_inputs     = "bad fee followed by state-changing msg, non-naet fee with contract-assets mint msg, malformed fee with bank send"
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
      attack              = "valid contract-assets lifecycle baseline plus unauthorized admin control sample"
      invariant           = "contract-assets lifecycle mutates denom state only under current admin authority"
      expected_behavior   = "valid create, mint, burn, admin change, and metadata query update denom state and supply exactly once"
      expected_events     = "contract-assets events match denom creation, mint, burn, admin change, and metadata deltas"
      expected_error_path = "unauthorized admin control sample is rejected before denom, supply, or metadata mutation"
      mutation_inputs     = "valid create denom, valid mint, valid burn, valid change admin, valid metadata query, unauthorized admin control"
      expected_rejection  = "unauthorized contract-assets lifecycle variants must fail without denom, supply, or admin mutation"
    }
    "TF-02" = [ordered]@{
      flow                = "invalid subdenom, duplicate denom, zero admin, native denom spoof, max metadata size"
      state               = "invalid contract-assets edge cases leave denom records, admin metadata, supply, and bank metadata unchanged"
      attack              = "invalid subdenom, duplicate denom, zero admin, native denom spoof, oversized metadata"
      invariant           = "contract-assets accepts only canonical denoms, non-zero admins, unique denoms, and bounded metadata"
      expected_behavior   = "valid contract-assets boundaries execute deterministically; invalid boundaries reject before state mutation"
      expected_events     = "accepted boundary contract-assets events match state deltas; rejected edge cases emit no success events"
      expected_error_path = "contract-assets validation rejects invalid subdenom, duplicate denom, zero admin, native spoof, or oversized metadata"
      mutation_inputs     = "empty subdenom, invalid subdenom chars, duplicate denom, zero admin, naet-like denom, max metadata size plus one"
      expected_rejection  = "invalid contract-assets edge cases must not alter denom records, supply, admin metadata, or native metadata"
    }
    "TF-03" = [ordered]@{
      flow                = "unauthorized mint, unauthorized burn, admin takeover, metadata spoofing, burn-from mismatch"
      state               = "adversarial contract-assets attempts cannot mint, burn, take admin, spoof metadata, or burn from another account"
      attack              = "unauthorized mint, unauthorized burn, admin takeover, metadata spoofing, burn-from mismatch"
      invariant           = "contract-assets authority, metadata, and burn source checks cannot be bypassed"
      expected_behavior   = "adversarial contract-assets mutations fail deterministically before supply or authority mutation"
      expected_events     = "failed contract-assets attacks emit no mint, burn, metadata, or admin success events"
      expected_error_path = "contract-assets msg server rejects unauthorized mint/burn/admin/metadata/burn-from paths before bank movement"
      mutation_inputs     = "non-admin mint, non-admin burn, forged change admin, spoofed metadata update, burn from different holder"
      expected_rejection  = "contract-assets attacks must not mint, burn, change admin, spoof metadata, or burn from mismatched accounts"
    }
    "TF-04" = [ordered]@{
      flow                = "supply delta exactness and authority metadata consistency"
      state               = "supply changes exactly by minted or burned amount and authority metadata remains consistent"
      attack              = "state drift attempt through mixed accepted and rejected contract-assets mint, burn, and admin operations"
      invariant           = "contract-assets supply delta is exact and authority metadata remains consistent"
      expected_behavior   = "contract-assets state integrity holds across lifecycle sequences and export/import"
      expected_events     = "contract-assets events reconcile to final supply, admin, metadata, and bank balance deltas"
      expected_error_path = "failed contract-assets operations preserve pre-failure supply, admin, metadata, and balances"
      mutation_inputs     = "accepted mint followed by failed mint, accepted burn followed by failed burn-from mismatch, change admin then old-admin mint, export/import after supply changes"
      expected_rejection  = "rejected contract-assets operations must preserve supply delta exactness and authority metadata consistency"
    }
    "TF-05" = [ordered]@{
      flow                = "contract-assets economic abuse around protocol fees, AET spoofing, and native supply inflation"
      state               = "contract-assets assets cannot pay protocol fees, spoof AET, or inflate native supply"
      attack              = "factory asset fee payment, AET spoof, native supply inflation, native metadata collision"
      invariant           = "contract-assets assets cannot pay protocol fees, spoof AET, or inflate native supply"
      expected_behavior   = "contract-assets economic rules keep factory denoms separate from native fee and native supply authority"
      expected_events     = "no protocol fee acceptance, native mint, or native metadata spoof event appears for rejected contract-assets abuse paths"
      expected_error_path = "economic abuse rejects before fee acceptance, native supply mutation, or native metadata mutation"
      mutation_inputs     = "factory denom as fee, factory denom named AET, factory denom named naet, mint shaped as native supply, native metadata spoof"
      expected_rejection  = "contract-assets economic abuse must not pay protocol fees, spoof AET, or inflate native supply"
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
    '^DEX Invariants$' { return "avm-dex-contract" }
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
      flow                = "mint authority boundaries for native, module, and contract-assets supply"
      state               = "only explicitly authorized mint paths can increase supply"
      attack              = "unauthorized mint, module account spoofing, contract-assets admin bypass, governance authority spoofing"
      expected_behavior   = "minting requires the module-specific authority and denom-specific permission before any supply increase"
      expected_events     = "no mint event is emitted from an unauthorized signer or unauthorized module account"
      expected_error_path = "unauthorized mint rejects before bank supply or denom metadata mutation"
      mutation_inputs     = "wrong admin, zero admin, wrong module account, forged authority, native denom mint through contract-assets"
      expected_rejection  = "unauthorized mint attempts cannot increase supply"
    }
    "ECON-04" = [ordered]@{
      flow                = "burn authority boundaries for user, module, and contract-assets supply"
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
      attack              = "fee denom spoofing, multi-denom fee bypass, contract-assets asset fee payment, missing fee bypass"
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
      attack              = "wrong LP denom burn, forged LP token, pool id collision, contract-assets spoofed lp/{pool_id} denom"
      expected_behavior   = "remove liquidity accepts only the canonical LP denom derived from the exact pool id"
      expected_events     = "LP burn events always reference the canonical pool LP denom"
      expected_error_path = "fake LP token rejects before reserves, shares, or module balances mutate"
      mutation_inputs     = "wrong lp denom, duplicate pool id, forged lp/{pool_id}, non-LP denom, contract-assets-created LP-like denom"
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
    "IDINV-01" = [ordered]@{
      flow                = "identity domain normalization and uniqueness"
      state               = "Domain names are unique"
      attack              = "duplicate normalized name, mixed-case spoof, whitespace spoof, unicode confusable label, duplicate import entry"
      expected_behavior   = "domain registration normalizes labels canonically and rejects duplicates before ownership state changes"
      expected_events     = "successful registration events contain one canonical domain id and no duplicate normalized name"
      expected_error_path = "duplicate or malformed normalized domain rejects before registry, resolver, auction, or NFT state mutation"
      mutation_inputs     = "AET case variants, leading/trailing whitespace, empty label, duplicate genesis domain, confusable label fixture"
      expected_rejection  = "two active records cannot exist for the same normalized domain"
    }
    "IDINV-02" = [ordered]@{
      flow                = "active domain auction exclusion"
      state               = "Active domains cannot be re-auctioned"
      attack              = "active domain re-auction, commit/reveal replay, auction id collision, owner griefing through duplicate auction"
      expected_behavior   = "auction start checks canonical domain lifecycle and rejects names that are active and unexpired"
      expected_events     = "no auction-created event is emitted for an active domain"
      expected_error_path = "active domain auction attempt rejects before escrow, bid, ownership, or resolver state mutation"
      mutation_inputs     = "auction start for active domain, replayed reveal, duplicate auction id, active domain with pending renewal"
      expected_rejection  = "active domain ownership cannot be displaced by a new auction"
    }
    "IDINV-03" = [ordered]@{
      flow                = "expired domain renewal and re-auction lifecycle"
      state               = "Expired domains require auction or explicit renewal path"
      attack              = "expired domain direct takeover, post-expiry resolver hijack, renewal window bypass, stale owner replay"
      expected_behavior   = "expired domains resolve only through explicit renewal rules or the configured auction/re-registration path"
      expected_events     = "expired domain state transitions emit renewal or auction lifecycle events before owner changes"
      expected_error_path = "direct expired-domain mutation rejects before owner, resolver, reverse lookup, or NFT state mutation"
      mutation_inputs     = "expired domain resolver update, direct owner set, renewal outside window, stale owner signature replay"
      expected_rejection  = "expired domain cannot be reassigned or resolved outside the lifecycle rules"
    }
    "IDINV-04" = [ordered]@{
      flow                = "resolver target validation"
      state               = "Resolver cannot point to malformed or zero address"
      attack              = "zero address resolver, malformed bech32 resolver, wrong prefix resolver, malformed contract target, zone endpoint spoof"
      expected_behavior   = "resolver updates validate target type, address format, zero-address rejection, and domain ownership before commit"
      expected_events     = "resolver update events never contain malformed or zero address targets"
      expected_error_path = "invalid resolver target rejects before resolver, reverse lookup, payment route, or index state mutation"
      mutation_inputs     = "zero address, malformed bech32, wrong prefix, empty contract address, oversized zone endpoint"
      expected_rejection  = "resolver records cannot commit malformed or zero-address targets"
    }
    "IDINV-05" = [ordered]@{
      flow                = "resolver-based payment preflight and rollback"
      state               = "Resolver-based payment fails before funds move if unresolved"
      attack              = "unresolved name payment drain, stale resolver cache, resolver deletion race, failed lookup after fee deduction"
      expected_behavior   = "name-based payment resolves canonical destination before any fund movement or value escrow"
      expected_events     = "unresolved payment emits no successful send, escrow, or resolver-payment event"
      expected_error_path = "unresolved or expired resolver path rejects before bank send, fee split, or queue state mutation"
      mutation_inputs     = "unknown domain, expired domain, deleted resolver, stale cache hit, resolver target zeroed between check and send"
      expected_rejection  = "resolver-based payment cannot move funds without a live valid resolver target"
    }
    "IDINV-06" = [ordered]@{
      flow                = "reverse lookup owner-approved consistency"
      state               = "Reverse lookup is consistent with owner-approved mapping"
      attack              = "reverse lookup poisoning, unauthorized primary name set, stale reverse after transfer, address-owner spoof"
      expected_behavior   = "reverse lookup updates require address-owner authorization and consistency with the forward domain owner mapping"
      expected_events     = "reverse lookup events reference the authorized address owner and canonical domain"
      expected_error_path = "unauthorized or inconsistent reverse lookup rejects before reverse index or resolver state mutation"
      mutation_inputs     = "wrong signer reverse update, transferred domain stale reverse, domain not owned by address, malformed address"
      expected_rejection  = "reverse lookup cannot claim a primary domain without owner-approved forward consistency"
    }
    "IDINV-07" = [ordered]@{
      flow                = "domain registry owner and NFT owner reconciliation"
      state               = "Domain registry owner and NFT representation owner do not diverge"
      attack              = "NFT transfer without registry update, registry owner overwrite, export/import owner drift, duplicate NFT id"
      expected_behavior   = "domain ownership changes update the registry owner and deterministic NFT owner atomically"
      expected_events     = "domain transfer and NFT transfer events reconcile to the same owner and domain id"
      expected_error_path = "owner divergence rejects before registry, NFT, resolver, or reverse lookup state commits"
      mutation_inputs     = "registry-only transfer, NFT-only transfer, duplicate NFT id import, stale owner export/import, failed resolver invalidation"
      expected_rejection  = "domain registry owner and NFT representation owner cannot diverge"
    }
    "IDINV-08" = [ordered]@{
      flow                = "subdomain ownership and resolver delegation boundaries"
      state               = "Subdomain ownership and resolver delegation do not bypass parent rules"
      attack              = "subdomain takeover, parent policy bypass, child resolver overwrite, sibling collision, delegated resolver privilege escalation"
      expected_behavior   = "subdomain issuance and resolver delegation obey parent owner policy and child owner authorization"
      expected_events     = "subdomain and delegation events include parent policy, child owner, and canonical subdomain id"
      expected_error_path = "unauthorized subdomain or delegation update rejects before subdomain, resolver, reverse lookup, or NFT state mutation"
      mutation_inputs     = "wrong parent signer, child resolver overwrite by parent when policy forbids, duplicate child label, delegated admin escalation"
      expected_rejection  = "subdomain ownership or resolver delegation cannot bypass parent policy or child owner authorization"
    }
    "EXECINV-01" = [ordered]@{
      flow                = "AVM malformed input panic safety"
      state               = "AVM malformed input does not panic"
      attack              = "malformed bytecode, invalid entrypoint payload, truncated message, oversized payload, invalid opcode stream"
      expected_behavior   = "AVM parser and dispatcher return deterministic errors for malformed input without panics or partial writes"
      expected_events     = "malformed AVM input emits no successful deploy, execute, query, migrate, or state-write event"
      expected_error_path = "malformed AVM input rejects before contract state, queue, gas accounting, or emitted message mutation"
      mutation_inputs     = "truncated bytecode, invalid opcode, oversized payload, invalid entrypoint args, random byte stream"
      expected_rejection  = "malformed AVM input cannot panic or commit state"
    }
    "EXECINV-02" = [ordered]@{
      flow                = "AVM deterministic gas accounting"
      state               = "AVM gas is bounded and deterministic"
      attack              = "gas underpayment, host function gas bypass, platform-dependent gas, storage write gas evasion, query gas overflow"
      expected_behavior   = "gas metering uses deterministic protocol costs and rejects execution before exceeding configured limits"
      expected_events     = "gas consumed and failure events are stable for the same input and state"
      expected_error_path = "out-of-gas rejects before committing contract, queue, or forwarded message state"
      mutation_inputs     = "zero gas, max gas, deep query, storage spam, host function loop, message forwarding burst"
      expected_rejection  = "AVM execution cannot exceed deterministic gas or underpay gas"
    }
    "EXECINV-03" = [ordered]@{
      flow                = "AVM infinite loop termination"
      state               = "Infinite loops are stopped by gas or instruction limits"
      attack              = "infinite loop bytecode, recursive internal call, unbounded host iteration, instruction counter overflow"
      expected_behavior   = "gas or instruction limits stop execution deterministically before node resource exhaustion"
      expected_events     = "loop termination emits deterministic out-of-gas or instruction-limit failure"
      expected_error_path = "looping execution rejects before state writes, queue messages, or refunds are committed"
      mutation_inputs     = "jump-to-self bytecode, recursive call payload, unbounded iterator contract, max-depth call sequence"
      expected_rejection  = "infinite loops cannot halt the chain or mutate state after limit failure"
    }
    "EXECINV-04" = [ordered]@{
      flow                = "contract state update determinism"
      state               = "Contract state updates are deterministic"
      attack              = "wall-clock dependency, randomness dependency, map iteration drift, external API dependency, platform-dependent int behavior"
      expected_behavior   = "contract writes derive only from deterministic inputs, canonical storage ordering, and protocol host functions"
      expected_events     = "same contract input and state produce identical write set, receipts, and emitted messages"
      expected_error_path = "nondeterministic host access rejects before contract state or queue mutation"
      mutation_inputs     = "same state with different local map order, random host call attempt, local time request, external API request"
      expected_rejection  = "contract state cannot diverge across deterministic replay"
    }
    "EXECINV-05" = [ordered]@{
      flow                = "queue canonical ordering"
      state               = "Queue ordering is deterministic"
      attack              = "queue insertion order drift, duplicate sequence, priority tie-break mismatch, map-order dequeue, delayed execution reorder"
      expected_behavior   = "queue ordering uses canonical height, priority, sequence, message id, and deterministic tie-breakers"
      expected_events     = "same queued messages produce identical dequeue and execution event order"
      expected_error_path = "duplicate sequence or ambiguous ordering rejects before queue state mutation"
      mutation_inputs     = "same queue items in different local insertion order, duplicate sequence, identical priority except message id"
      expected_rejection  = "queue ordering cannot depend on local map or mempool order"
    }
    "EXECINV-06" = [ordered]@{
      flow                = "cross-zone message replay prevention"
      state               = "Cross-zone message replay is rejected"
      attack              = "duplicate mesh message, stale receipt replay, forged proof replay, same message id on different shard"
      expected_behavior   = "replay markers and receipt uniqueness prevent the same cross-zone message from executing twice"
      expected_events     = "duplicate message emits no successful delivery, receipt, value transfer, or contract execution event"
      expected_error_path = "replayed cross-zone message rejects before destination state or value movement"
      mutation_inputs     = "same message id twice, stale finality reference, duplicate receipt, wrong source shard replay"
      expected_rejection  = "cross-zone replay cannot mutate state or spend assets twice"
    }
    "EXECINV-07" = [ordered]@{
      flow                = "bounce and refund double-spend prevention"
      state               = "Bounce/refund cannot double-spend"
      attack              = "duplicate refund receipt, bounce/refund loop, failed execution double refund, invalid destination refund replay"
      expected_behavior   = "refund and bounce receipts are single-use and cannot produce multiple value releases for one message"
      expected_events     = "one source message produces at most one finalized refund or bounce value movement"
      expected_error_path = "duplicate bounce or refund rejects before value transfer, receipt, marker, or queue state mutation"
      mutation_inputs     = "same failed message refund twice, bounce of bounce, duplicate refund receipt, timeout plus execution-failure refund"
      expected_rejection  = "bounce/refund paths cannot double-spend or loop indefinitely"
    }
    "EXECINV-08" = [ordered]@{
      flow                = "message loop depth and per-block processing bounds"
      state               = "Message loops are bounded by depth and per-block limits"
      attack              = "recursive message loop, queue flood, bounce storm, unbounded internal call fanout, per-block limit bypass"
      expected_behavior   = "message processing enforces deterministic depth, emitted-message, and per-block queue limits"
      expected_events     = "bounded deferral or rejection events appear when loop depth or per-block limits are reached"
      expected_error_path = "limit breach rejects or defers before unbounded queue growth or repeated value movement"
      mutation_inputs     = "A->B->A recursive messages, many emitted messages, bounce storm, delayed execution loop"
      expected_rejection  = "message loops cannot exhaust block execution or bypass per-block limits"
    }
    "EXECINV-09" = [ordered]@{
      flow                = "queue export/import exactness"
      state               = "Export/import preserves queue state exactly"
      attack              = "queue export ordering drift, missing replay marker, duplicate queued item, corrupted delayed execution height, receipt marker loss"
      expected_behavior   = "exported queue state imports to the same ordered items, sequence counters, replay markers, receipts, and delayed heights"
      expected_events     = "pure export/import emits no queue execution, refund, bounce, or receipt mutation event"
      expected_error_path = "corrupted queue import rejects during genesis validation before chain start"
      mutation_inputs     = "reordered queue items, duplicate sequence, missing marker, corrupted receipt, malformed delayed height"
      expected_rejection  = "export/import cannot reorder or lose queue state"
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

function Get-AexsCoreExploitOverride {
  param([string]$ExploitId)
  $overrides = @{
    "COREEXP-01" = [ordered]@{
      path            = "attempt to make one validator sign conflicting blocks for the same height and round, then feed both commits into replay and evidence validation"
      expected_state  = "only one canonical block is committed; duplicate signing is recorded as objective evidence; slash/tombstone path is deterministic"
      affected        = @("app", "x/slashing", "x/staking", "CometBFT consensus")
      severity        = "Critical"
      fix             = "enforce duplicate-sign evidence validation, tombstone persistence, deterministic slash accounting, and replay regression for conflicting vote evidence"
    }
    "COREEXP-02" = [ordered]@{
      path            = "attempt equivocation with conflicting prevote/precommit evidence across heights and rounds, including stale and duplicate evidence variants"
      expected_state  = "valid equivocation evidence slashes once; stale, malformed, or duplicate evidence is rejected before validator state mutation"
      affected        = @("x/slashing", "x/staking", "app evidence handling", "CometBFT consensus")
      severity        = "Critical"
      fix             = "harden evidence freshness, uniqueness, validator identity binding, and deterministic slash/jail/tombstone tests"
    }
    "COREEXP-03" = [ordered]@{
      path            = "attempt long-range history rewrite with an old validator set after unbonding and export/import replay"
      expected_state  = "current trust period, unbonding windows, and checkpointed app state reject rewritten history and preserve canonical app hash"
      affected        = @("x/staking", "x/slashing", "app", "CometBFT light-client trust model")
      severity        = "Critical"
      fix             = "enforce unbonding/evidence windows, state-sync trust assumptions, checkpoint documentation, and long-range replay tests"
    }
    "COREEXP-04" = [ordered]@{
      path            = "attempt stake grinding by cycling delegation, validator creation, and reward timing to bias proposer or validator-set selection"
      expected_state  = "validator power and proposer eligibility derive only from canonical bonded stake and deterministic staking state"
      affected        = @("x/staking", "x/distribution", "app validator updates")
      severity        = "High"
      fix             = "add stake-grinding simulations, delegation timing invariants, and validator power update regression tests"
    }
    "COREEXP-05" = [ordered]@{
      path            = "simulate validator cartel concentration to test governance, slashing, and validator-set safety under concentrated voting power"
      expected_state  = "validator concentration cannot bypass deterministic slashing, parameter bounds, or documented governance quorum/threshold rules"
      affected        = @("x/staking", "x/gov", "x/slashing", "distribution economics")
      severity        = "High"
      fix             = "document concentration risk, enforce hard parameter bounds, and add cartel governance/slashing scenario tests"
    }
    "COREEXP-06" = [ordered]@{
      path            = "attempt delegation manipulation through rapid delegate/redelegate/unbond sequences and share rounding boundaries"
      expected_state  = "delegator shares, validator tokens, unbonding records, and validator power remain consistent after every accepted or rejected transition"
      affected        = @("x/staking", "x/distribution", "x/slashing")
      severity        = "High"
      fix             = "add staking share-rounding, redelegation, unbonding, and slash-period invariant regression tests"
    }
    "COREEXP-07" = [ordered]@{
      path            = "attempt self-delegation inflation by manipulating validator self-bond, shares, commission, and minimum self-delegation boundaries"
      expected_state  = "self-delegation cannot inflate validator tokens, delegator shares, voting power, commission, or rewards"
      affected        = @("x/staking", "x/distribution", "app validator updates")
      severity        = "Critical"
      fix             = "enforce self-delegation bounds, share math invariants, and validator power/supply accounting tests"
    }
    "COREEXP-08" = [ordered]@{
      path            = "attempt fake validator liveness by spoofing signing info, missed blocks, or evidence state across restart/export/import"
      expected_state  = "validator signing info, missed block counters, jail state, and liveness updates remain objective and deterministic"
      affected        = @("x/slashing", "x/staking", "app restart/export")
      severity        = "High"
      fix             = "add liveness replay tests, signing-info export/import checks, and downtime edge-case coverage"
    }
    "COREEXP-09" = [ordered]@{
      path            = "simulate validator eclipse by isolating validator peers and attempting delayed evidence, stale blocks, or divergent mempool assumptions"
      expected_state  = "network partition cannot create accepted conflicting state; evidence and finality remain objective after reconnection"
      affected        = @("CometBFT P2P", "app", "x/slashing", "localnet tooling")
      severity        = "High"
      fix             = "add localnet partition/evidence smoke tests and document validator peer diversity requirements"
    }
    "COREEXP-10" = [ordered]@{
      path            = "attempt block withholding by proposer or validator subset to delay finality or bias transaction inclusion"
      expected_state  = "withheld blocks do not commit; liveness/finality delay is observable; safety and app hash determinism remain intact"
      affected        = @("CometBFT consensus", "mempool", "x/slashing", "observability")
      severity        = "High"
      fix             = "add block-withholding localnet tests, liveness metrics, and downtime/slashing evidence verification"
    }
    "COREEXP-11" = [ordered]@{
      path            = "attempt fork choice manipulation with conflicting proposals, delayed commits, and reordered evidence delivery"
      expected_state  = "CometBFT finality chooses only a 2/3+ committed block and ABCI replay produces one canonical app hash"
      affected        = @("CometBFT consensus", "app FinalizeBlock", "x/slashing")
      severity        = "Critical"
      fix             = "add fork-choice replay fixtures, evidence ordering tests, and app-hash determinism checks under conflicting proposal scenarios"
    }
    "COREEXP-12" = [ordered]@{
      path            = "attempt finality delay manipulation through validator delay, vote withholding, mempool spam, and block propagation lag"
      expected_state  = "finality delay is bounded by consensus/liveness assumptions and does not mutate app state inconsistently"
      affected        = @("CometBFT consensus", "mempool", "x/slashing", "observability")
      severity        = "High"
      fix             = "add deterministic chaos tests for delayed votes, delayed propagation, spam pressure, and finality health metrics"
    }
    "COREEXP-13" = [ordered]@{
      path            = "simulate Byzantine majority to verify the test harness detects the safety boundary and does not mark the scope production safe"
      expected_state  = "campaign records safety failure or Byzantine-boundary condition; production-safe decision is blocked until triaged"
      affected        = @("CometBFT consensus", "app", "x/staking", "x/slashing", "test harness")
      severity        = "Critical"
      fix             = "keep Byzantine-majority scenario as a blocking stress test and document that protocol safety assumes less than one-third Byzantine voting power"
    }
  }
  if ($overrides.ContainsKey($ExploitId)) {
    return $overrides[$ExploitId]
  }
  return $null
}

function Get-AexsCoreExploitRecords {
  param([string]$Text, [string]$CampaignId)
  $section = Get-AexsMarkdownSection -Text $Text -Heading "Exploit Task Catalog"
  if ([string]::IsNullOrWhiteSpace($section)) {
    return @()
  }

  $capture = $false
  $items = @()
  foreach ($line in ($section -split "`r?`n")) {
    if ($line -match '^###\s+1\.\s+Consensus And Aether Core Exploits\s*$') {
      $capture = $true
      continue
    }
    if ($capture -and $line -match '^###\s+') {
      break
    }
    if ($capture -and $line -match '^- \[ \]\s+(.+?)\s*$') {
      $items += $Matches[1].Trim().TrimEnd(".")
    }
  }

  $records = @()
  for ($i = 0; $i -lt $items.Count; $i++) {
    $id = "COREEXP-{0:00}" -f ($i + 1)
    $description = $items[$i]
    $override = Get-AexsCoreExploitOverride -ExploitId $id
    $seedHash = (Get-AexsSha256Hex -Text "$CampaignId|core-exploit|$id|$description").Substring(0, 16)
    $seed = "aexs-$($id.ToLowerInvariant())-$seedHash"
    $affected = if ($null -ne $override -and $override.Contains("affected")) { @($override["affected"]) } else { @("app", "x/staking", "x/slashing", "CometBFT consensus") }
    $path = Get-AexsOverrideValue -Override $override -Field "path" -Fallback $description
    $expected = Get-AexsOverrideValue -Override $override -Field "expected_state" -Fallback "exploit attempt must not violate Aether Core safety, validator-set, staking, slashing, or app-hash invariants"
    $severity = Get-AexsOverrideValue -Override $override -Field "severity" -Fallback "High"
    $fix = Get-AexsOverrideValue -Override $override -Field "fix" -Fallback "add deterministic replay, adversarial localnet, invariant, and regression coverage for this exploit path"

    $records += [ordered]@{
      exploit_id         = $id
      category           = "Consensus And Aether Core Exploits"
      description        = $description
      exploit_path       = $path
      seed               = $seed
      step_list          = @(
        "Run AEXS exploit scenario $id",
        "Use seed $seed",
        "Construct the adversarial consensus or staking sequence",
        "Record expected state before execution",
        "Record actual state after execution",
        "If exploit succeeds, minimize the sequence and write AUDIT_RESULT.md"
      )
      expected_state     = $expected
      actual_state       = "not_executed_preflight"
      affected_modules   = $affected
      severity           = $severity
      fix_recommendation = $fix
      status             = "planned_not_executed"
      valid              = $true
      invalid_reasons    = @()
    }
  }
  return $records
}

function Get-AexsSlashingExploitOverrides {
  $overrides = @{
    "SLASHEXP-01" = [ordered]@{
      path            = "submit valid evidence after delay boundaries and through redelegation/unbonding timing to attempt avoiding the slash"
      expected_state  = "fresh valid delayed evidence still applies deterministic slash/jail/tombstone effects; stale evidence is rejected without mutation"
      affected        = @("x/slashing", "x/staking", "app evidence handling")
      severity        = "Critical"
      fix             = "add delayed-evidence freshness tests, slash-period accounting checks, and redelegation/unbonding evidence regression coverage"
    }
    "SLASHEXP-02" = [ordered]@{
      path            = "feed malformed equivocation proof variants to evidence handling and attempt accidental acceptance"
      expected_state  = "malformed equivocation proof is rejected before slash accounting, validator status, or tombstone state changes"
      affected        = @("x/slashing", "app evidence handling", "CometBFT evidence")
      severity        = "Critical"
      fix             = "harden proof decoding, validator identity binding, signature checks, and malformed evidence tests"
    }
    "SLASHEXP-03" = [ordered]@{
      path            = "race slashing evidence against delegation, redelegation, unbonding, jail, and export/import transitions"
      expected_state  = "slash accounting is atomic and deterministic; stake movement cannot avoid objective slash effects"
      affected        = @("x/slashing", "x/staking", "x/distribution")
      severity        = "Critical"
      fix             = "add stateful race-sequence tests for slashing with staking transitions and cache-context rollback assertions"
    }
    "SLASHEXP-04" = [ordered]@{
      path            = "redelegate stake after slashable behavior and attempt partial slash evasion across source and destination validators"
      expected_state  = "redelegated stake remains slashable according to slash period and historical staking state"
      affected        = @("x/slashing", "x/staking")
      severity        = "High"
      fix             = "add redelegation slash-period invariants and historical stake lookup regression tests"
    }
    "SLASHEXP-05" = [ordered]@{
      path            = "unbond stake during evidence delay and attempt to exit before objective evidence applies"
      expected_state  = "unbonding stake remains slashable throughout the protocol evidence window and cannot escape via timing"
      affected        = @("x/slashing", "x/staking")
      severity        = "Critical"
      fix             = "test unbonding evidence windows, completion height boundaries, and slash accounting over unbonding delegations"
    }
    "SLASHEXP-06" = [ordered]@{
      path            = "trigger upgrade/export/import timing around jailed validator state and attempt jail escape or tombstone loss"
      expected_state  = "jailed and tombstoned state persists across upgrades, migrations, export/import, and restart"
      affected        = @("x/slashing", "x/staking", "app upgrades", "genesis export")
      severity        = "High"
      fix             = "add migration/export/import tests for signing info, jailed state, tombstone state, and validator status"
    }
    "SLASHEXP-07" = [ordered]@{
      path            = "replay previously accepted invalid/stale/duplicate evidence and attempt repeated or unintended slash mutation"
      expected_state  = "invalid evidence replay is rejected; valid evidence is single-use and cannot slash twice"
      affected        = @("x/slashing", "app evidence handling")
      severity        = "High"
      fix             = "add evidence replay markers, duplicate evidence tests, and exact-once slash accounting assertions"
    }
  }
  return $overrides
}

function Get-AexsTxAuthBankExploitOverrides {
  $overrides = @{
    "TXEXP-01" = [ordered]@{
      path            = "replay accepted signed transaction bytes through mempool, block replay, restart, and export/import paths"
      expected_state  = "accepted transaction mutates state once; byte-for-byte replay fails sequence or replay checks before message execution"
      affected        = @("x/auth", "app ante", "mempool", "x/bank")
      severity        = "Critical"
      fix             = "add signed-byte replay tests across CheckTx/FinalizeBlock/restart and sequence no-mutation assertions"
    }
    "TXEXP-02" = [ordered]@{
      path            = "reuse signed bytes or sign doc across wrong chain id, account number, or context to attempt cross-context execution"
      expected_state  = "wrong chain id or sign context rejects in ante before fee, sequence, or message state mutation"
      affected        = @("x/auth", "app ante")
      severity        = "Critical"
      fix             = "add wrong-chain-id replay fixtures and sign-doc domain separation regression tests"
    }
    "TXEXP-03" = [ordered]@{
      path            = "manipulate account sequence with stale, future, duplicate, or account-mismatched nonce values"
      expected_state  = "invalid nonce rejects before sequence increment, fee mutation, or message execution"
      affected        = @("x/auth", "app ante")
      severity        = "High"
      fix             = "add nonce boundary tests for stale/future/duplicate sequences and rejected-state snapshot comparisons"
    }
    "TXEXP-04" = [ordered]@{
      path            = "alter transaction encoding, memo, auth info, or protobuf field ordering to attempt malleable equivalent execution"
      expected_state  = "malleated transaction is either a distinct valid transaction with its own signature/sequence or rejected before mutation"
      affected        = @("x/auth", "app tx decoding", "x/memo")
      severity        = "High"
      fix             = "add tx canonicalization, sign bytes, malformed protobuf, and memo mutation regression tests"
    }
    "TXEXP-05" = [ordered]@{
      path            = "underpay fee with zero, missing, below-min, wrong-denom, or multi-denom fee fields"
      expected_state  = "fee underpayment rejects in ante before messages execute or module state mutates"
      affected        = @("x/fees", "x/auth", "app ante")
      severity        = "Critical"
      fix             = "add fee ante tests for missing/zero/below-min/non-naet/multi-denom underpayment and no-message-execution assertions"
    }
    "TXEXP-06" = [ordered]@{
      path            = "manipulate fee amount, gas, split rounding, or refund accounting to inflate fee credit or rewards"
      expected_state  = "fee accounting is exact and deterministic; malformed fee inflation attempts cannot mint, refund, or over-credit rewards"
      affected        = @("x/fees", "x/distribution", "x/bank", "app ante")
      severity        = "High"
      fix             = "add fee split, rounding, max fee, refund, and distribution accounting invariants"
    }
    "TXEXP-07" = [ordered]@{
      path            = "flood low-fee or malformed transactions to grief mempool, route priority, and block inclusion without paying required cost"
      expected_state  = "low-fee spam is rejected or bounded by deterministic admission policy and cannot mutate state"
      affected        = @("x/fees", "mempool", "x/routing", "app ante")
      severity        = "High"
      fix             = "add spam-burst simulations, fee admission tests, priority bounds, and mempool no-state-mutation assertions"
    }
    "TXEXP-08" = [ordered]@{
      path            = "force a multi-send branch failure after earlier outputs and attempt partial committed bank movement"
      expected_state  = "multi-send is atomic; any failed output rejects the whole transaction before balances or supply mutate"
      affected        = @("x/bank", "app cache context")
      severity        = "Critical"
      fix             = "add partial multi-send failure tests with balance/supply snapshots before and after rejection"
    }
    "TXEXP-09" = [ordered]@{
      path            = "race double spend attempts through repeated tx delivery, mempool rebroadcast, same sequence variants, and parallel local submission"
      expected_state  = "only one spend can commit for a sequence and balance state; competing attempts fail before double debit"
      affected        = @("x/auth", "x/bank", "mempool", "app")
      severity        = "Critical"
      fix             = "add repeated delivery, same-sequence, and insufficient-funds replay tests with account sequence/balance invariants"
    }
    "TXEXP-10" = [ordered]@{
      path            = "replay a state transition that fails after partial writes and attempt rollback bypass or cache context leakage"
      expected_state  = "failed state transition rolls back all writes and replay observes the original pre-failure state"
      affected        = @("app cache context", "x/bank", "x/fees", "avm-dex-contract")
      severity        = "High"
      fix             = "add replayed failure rollback tests across bank, fees, contract-assets, and DEX handlers"
    }
    "TXEXP-11" = [ordered]@{
      path            = "inject zero address as signer, recipient, admin, authority, or bank transfer endpoint"
      expected_state  = "zero-address signer or recipient path rejects before account, balance, authority, resolver, or module state mutation"
      affected        = @("x/auth", "x/bank", "x/fees", "app address validation")
      severity        = "Critical"
      fix             = "add zero-address adversarial tests for signer, recipient, admin, authority, and fee/bank paths"
    }
  }
  return $overrides
}

function Get-AexsTokenEconomyExploitOverrides {
  $overrides = @{
    "TOKENEXP-01" = [ordered]@{
      path            = "attempt contract-assets admin takeover by spoofing denom admin, changing admin out of order, or replaying stale authority"
      expected_state  = "factory denom admin cannot be changed or used for mint authority unless the current canonical admin authorizes it"
      affected        = @("x/bank", "app auth")
      severity        = "Critical"
      fix             = "add admin takeover regression tests for create denom, change admin, mint, stale admin replay, and zero-admin boundaries"
    }
    "TOKENEXP-02" = [ordered]@{
      path            = "attempt unauthorized burn from another account, module account, or native denom through contract-assets burn paths"
      expected_state  = "burn requires exact authority and source ownership; unauthorized burn cannot debit balances or reduce supply"
      affected        = @("x/bank")
      severity        = "Critical"
      fix             = "add burn-from mismatch, wrong admin, module account, native denom, and supply delta tests"
    }
    "TOKENEXP-03" = [ordered]@{
      path            = "time governance parameter changes around mint, inflation, staking reward, and distribution state to manipulate inflation"
      expected_state  = "governance timing cannot set inflation or mint parameters outside hard bounds or execute before the configured delay"
      affected        = @("x/gov", "x/mint", "x/distribution", "x/staking")
      severity        = "Critical"
      fix             = "add governance delay, hard-bound params, mint epoch, and distribution replay tests"
    }
    "TOKENEXP-04" = [ordered]@{
      path            = "manipulate fee routing through malformed fee splits, wrong treasury target, non-naet fee denom, or routing epoch changes"
      expected_state  = "fees route only through configured deterministic targets and cannot be redirected by tx shape, denom spoofing, or local routing state"
      affected        = @("x/fees", "x/bank", "x/distribution", "x/routing")
      severity        = "High"
      fix             = "add fee-routing split, treasury target, non-naet rejection, and deterministic routing interaction tests"
    }
    "TOKENEXP-05" = [ordered]@{
      path            = "attempt treasury drain with a governance proposal that sets unsafe recipients, invalid spend params, or unbounded distribution targets"
      expected_state  = "governance-controlled treasury actions remain within hard protocol bounds and reject invalid recipients or unsafe params"
      affected        = @("x/gov", "x/fees", "x/bank", "treasury module account")
      severity        = "Critical"
      fix             = "add governance proposal tests for treasury spend bounds, invalid authority, zero address, and unsafe recipient rejection"
    }
    "TOKENEXP-06" = [ordered]@{
      path            = "inflate staking rewards by manipulating commission, reward withdrawal timing, validator state, or distribution indexes"
      expected_state  = "reward accounting pays no more than accrued deterministic rewards and cannot mint outside authorized distribution paths"
      affected        = @("x/staking", "x/distribution", "x/mint")
      severity        = "Critical"
      fix             = "add staking reward inflation tests for commission, repeated withdrawal, tiny rewards, and distribution index replay"
    }
    "TOKENEXP-07" = [ordered]@{
      path            = "farm staking rewards through delegate, redelegate, unbond, rebond, and withdraw loops around rounding boundaries"
      expected_state  = "staking reward loops cannot produce extra rewards, bypass slash risk, or create accounting drift"
      affected        = @("x/staking", "x/distribution", "x/slashing")
      severity        = "High"
      fix             = "add stateful staking reward farming loop tests and share/reward rounding invariants"
    }
    "TOKENEXP-08" = [ordered]@{
      path            = "manipulate supply through edge-case mint amount, metadata, duplicate denom, zero admin, max amount, or export/import paths"
      expected_state  = "mint amount and denom validation reject edge cases and accepted mints change supply by exactly the requested amount"
      affected        = @("x/bank", "genesis export/import")
      severity        = "Critical"
      fix             = "add edge-case mint tests for zero/max amounts, duplicate denom, invalid metadata, export/import, and exact supply delta"
    }
    "TOKENEXP-09" = [ordered]@{
      path            = "spoof native denom through contract-assets subdenom, bank metadata, display denom, LP denom, or fee denom aliases"
      expected_state  = "factory assets cannot spoof native denom, native metadata, protocol fee denom, staking denom, or LP denom namespace"
      affected        = @("x/bank", "x/fees", "x/staking")
      severity        = "Critical"
      fix             = "add native denom spoof tests for metadata, base/display denom, contract-assets subdenom, fee denom, staking denom, and LP namespace"
    }
    "TOKENEXP-10" = [ordered]@{
      path            = "exploit display/base decimal mismatch between ORB display, naet base units, bank metadata, fees, and localnet genesis balances"
      expected_state  = "display/base metadata remains canonical and cannot alter supply, fee amount, staking amount, or wallet-visible balances"
      affected        = @("app params", "x/bank", "x/fees", "localnet genesis")
      severity        = "High"
      fix             = "add bank metadata, exponent, fee amount, staking denom, and genesis display/base decimal regression tests"
    }
  }
  return $overrides
}

function Get-AexsDexExploitOverrides {
  $overrides = @{
    "DEXEXP-01" = [ordered]@{
      path            = "execute swaps against crafted reserves to reduce the fee-adjusted constant product or violate canonical reserve ordering"
      expected_state  = "fee-adjusted constant-product invariant holds after every accepted swap and violation attempts reject before reserve mutation"
      affected        = @("avm-dex-contract", "x/bank")
      severity        = "Critical"
      fix             = "add constant-product invariant tests for reserve ordering, fee-adjusted math, extreme ratios, and repeated tiny swaps"
    }
    "DEXEXP-02" = [ordered]@{
      path            = "drain pool liquidity through repeated swaps, rounded outputs, stale quotes, and fee edge cases"
      expected_state  = "swap math, slippage checks, and module balances prevent liquidity drain beyond deterministic AMM rules"
      affected        = @("avm-dex-contract", "x/bank")
      severity        = "Critical"
      fix             = "add multi-step liquidity drain simulations with reserve/module balance and slippage assertions"
    }
    "DEXEXP-03" = [ordered]@{
      path            = "initialize pool with same denoms, duplicate pair, tiny reserves, wrong canonical order, or invalid initial LP shares"
      expected_state  = "pool creation rejects malformed pairs and creates canonical pair index, reserves, and LP supply atomically"
      affected        = @("avm-dex-contract", "x/bank")
      severity        = "High"
      fix             = "add pool initialization tests for duplicate pairs, same denom, zero/tiny liquidity, canonical order, and LP genesis"
    }
    "DEXEXP-04" = [ordered]@{
      path            = "inflate LP tokens by forged LP denom, duplicate pool id, add-liquidity rounding, or LP mint without matching reserves"
      expected_state  = "LP supply equals pool total shares and cannot be minted without matching reserve deposits"
      affected        = @("avm-dex-contract", "x/bank")
      severity        = "Critical"
      fix             = "add LP inflation tests for forged LP denom, duplicate pool id, share rounding, and reserve/share reconciliation"
    }
    "DEXEXP-05" = [ordered]@{
      path            = "race liquidity removal against swaps, LP burns, pool updates, and failed bank movement to extract extra reserves"
      expected_state  = "remove liquidity is atomic and deterministic; races cannot burn shares twice or withdraw more than owned"
      affected        = @("avm-dex-contract", "x/bank", "app cache context")
      severity        = "High"
      fix             = "add stateful remove-liquidity race tests with LP supply, reserve, and rollback snapshots"
    }
    "DEXEXP-06" = [ordered]@{
      path            = "swap or add/remove liquidity against zero-liquidity or insolvent pools to trigger divide-by-zero or impossible output"
      expected_state  = "zero-liquidity and insolvent pools reject before arithmetic panic, bank movement, or pool mutation"
      affected        = @("avm-dex-contract", "x/bank")
      severity        = "Critical"
      fix             = "add zero-liquidity, insolvent pool, divide-by-zero, and panic-safety tests"
    }
    "DEXEXP-07" = [ordered]@{
      path            = "desynchronize pool reserves from module account balances through corrupted state, partial bank moves, or export/import drift"
      expected_state  = "pool reserves reconcile with module balances and any mismatch rejects or records a critical invariant failure"
      affected        = @("avm-dex-contract", "x/bank", "genesis export/import")
      severity        = "Critical"
      fix             = "add reserve/module balance reconciliation tests for every DEX operation and export/import round trip"
    }
    "DEXEXP-08" = [ordered]@{
      path            = "force bank movement failure after DEX state update attempt and check for partial reserve, share, or LP mutation"
      expected_state  = "failed bank movement rolls back all DEX state changes in the cached context"
      affected        = @("avm-dex-contract", "x/bank", "app cache context")
      severity        = "Critical"
      fix             = "add failed bank movement simulations for create pool, add liquidity, remove liquidity, swap, and LP mint/burn"
    }
    "DEXEXP-09" = [ordered]@{
      path            = "bypass slippage with stale quote, min-output off-by-one, rounded output, route manipulation, or denom order confusion"
      expected_state  = "swap rejects before state mutation whenever deterministic output is below user slippage constraints"
      affected        = @("avm-dex-contract", "x/routing", "x/bank")
      severity        = "High"
      fix             = "add slippage bypass tests for stale quotes, min-output boundaries, rounding, route hints, and denom ordering"
    }
    "DEXEXP-10" = [ordered]@{
      path            = "exploit AMM rounding through tiny repeated swaps, asymmetric reserves, fee rounding, or LP share truncation"
      expected_state  = "rounding rules are deterministic, bounded, and cannot drain value or violate reserve/share invariants"
      affected        = @("avm-dex-contract", "x/bank")
      severity        = "High"
      fix             = "add property and fuzz tests for swap rounding, LP rounding, tiny amounts, and asymmetric reserve sequences"
    }
  }
  return $overrides
}

function Get-AexsLoadSystemExploitOverrides {
  $overrides = @{
    "LOADEXP-01" = [ordered]@{
      path            = "burst spam transactions to manipulate LOAD_SCORE beyond EMA smoothing and MAX_DELTA caps"
      expected_state  = "LOAD_SCORE remains bounded in [0,1], applies EMA and MAX_DELTA, and cannot jump based on local spam bursts"
      affected        = @("x/load", "x/routing", "mempool", "x/sharding/sim")
      severity        = "High"
      fix             = "add spam-burst load simulations with EMA, MAX_DELTA, deterministic replay, and no local-mempool dependency checks"
    }
    "LOADEXP-02" = [ordered]@{
      path            = "inflate local mempool size artificially and attempt to alter consensus load state or shard activation"
      expected_state  = "node-local mempool inflation cannot directly mutate consensus LOAD_SCORE or deterministic shard activation state"
      affected        = @("x/load", "mempool", "x/sharding/sim")
      severity        = "High"
      fix             = "add local-vs-consensus mempool metric tests and reject node-local-only metrics from consensus state"
    }
    "LOADEXP-03" = [ordered]@{
      path            = "saturate blocks with high gas usage to force unsafe load escalation or unrelated-zone degradation"
      expected_state  = "block utilization contributes through normalized deterministic metrics and cannot bypass spike caps or unrelated zone isolation"
      affected        = @("x/load", "x/fees", "x/sharding/sim", "app")
      severity        = "High"
      fix             = "add block saturation scenarios for normalized gas utilization, zone isolation, and fee/degradation bounds"
    }
    "LOADEXP-04" = [ordered]@{
      path            = "amplify execution delay with slow transactions or VM-like workloads to poison execution_time_score"
      expected_state  = "execution delay metrics are bounded, deterministic, and cannot include node-local wall-clock or hardware-specific timings"
      affected        = @("x/load", "x/vm", "x/execution", "app")
      severity        = "Critical"
      fix             = "add deterministic execution-time metric tests that reject wall-clock and hardware-dependent inputs"
    }
    "LOADEXP-05" = [ordered]@{
      path            = "slowly poison EMA windows with sustained near-threshold load to activate shards or fees at attacker-chosen times"
      expected_state  = "EMA, window size, thresholds, and cooldowns remain deterministic and resistant to slow-poison manipulation"
      affected        = @("x/load", "x/sharding/sim", "x/fees")
      severity        = "High"
      fix             = "add EMA slow-poison simulations across low/medium/high thresholds, cooldowns, and export/import replay"
    }
    "LOADEXP-06" = [ordered]@{
      path            = "oscillate load around thresholds to repeatedly activate/deactivate shards and destabilize fees"
      expected_state  = "MAX_DELTA, EMA, threshold hysteresis, and cooldowns prevent unsafe oscillation and deterministic state churn"
      affected        = @("x/load", "x/sharding/sim", "x/fees")
      severity        = "High"
      fix             = "add oscillating load tests for spike cap, cooldown, shard activation/deactivation, and fee stability"
    }
    "LOADEXP-07" = [ordered]@{
      path            = "target specific routing keys to overload a shard while manipulating load metrics and active shard counts"
      expected_state  = "shard assignment and activation remain deterministic, data-available, and resistant to overload targeting"
      affected        = @("x/load", "x/routing", "x/sharding/sim")
      severity        = "Critical"
      fix             = "add shard overload targeting simulations with routing keys, active shard counts, data availability, and validator reassignment"
    }
    "LOADEXP-08" = [ordered]@{
      path            = "overpay priority fees to game load priority, degrade lower-priority users, or force route changes"
      expected_state  = "fee class is bounded and cannot manipulate routing, load score, or starvation beyond deterministic priority rules"
      affected        = @("x/load", "x/fees", "x/routing", "x/reputation")
      severity        = "High"
      fix             = "add priority fee gaming tests for bounded fee class, starvation prevention, and route invariance"
    }
    "LOADEXP-09" = [ordered]@{
      path            = "destabilize adaptive fees through alternating congestion, fee overpayment, and load threshold manipulation"
      expected_state  = "dynamic fees remain governance-bounded, deterministic, and resistant to oscillation or attacker-driven instability"
      affected        = @("x/load", "x/fees", "x/gov")
      severity        = "High"
      fix             = "add adaptive fee destabilization tests for governance bounds, load thresholds, hysteresis, and deterministic replay"
    }
  }
  return $overrides
}

function Get-AexsRoutingEngineExploitOverrides {
  $overrides = @{
    "ROUTEEXP-01" = [ordered]@{
      path            = "bias routing decisions with crafted message types, actor keys, routing hints, or fee/reputation classes"
      expected_state  = "routing remains a pure deterministic classifier over stable message strings and bounded protocol inputs"
      affected        = @("x/routing", "x/load", "x/fees", "x/reputation")
      severity        = "Critical"
      fix             = "add routing bias tests for msg classification, primary actor extraction, bounded fee/reputation classes, and ignored local hints"
    }
    "ROUTEEXP-02" = [ordered]@{
      path            = "target one execution zone with transactions that manipulate class selection or locality to create zone congestion"
      expected_state  = "zone selection follows deterministic class/locality rules and congestion is isolated without affecting unrelated zones"
      affected        = @("x/routing", "x/zones", "x/load", "x/sharding/sim")
      severity        = "High"
      fix             = "add zone congestion tests for class mapping, locality extraction, load isolation, and unrelated-zone progress"
    }
    "ROUTEEXP-03" = [ordered]@{
      path            = "starve compute shards through skewed routing keys, priority ordering, active shard count changes, or epoch boundaries"
      expected_state  = "valid shard work has bounded progress and shard assignment remains deterministic across routing epochs"
      affected        = @("x/routing", "x/sharding/sim", "x/scheduler")
      severity        = "High"
      fix             = "add compute shard starvation simulations with skewed keys, epoch changes, and priority queue fairness checks"
    }
    "ROUTEEXP-04" = [ordered]@{
      path            = "monopolize a hot zone by concentrating traffic, fees, and actor locality to suppress other zones"
      expected_state  = "hot-zone load cannot monopolize global execution capacity or suppress unrelated zone routing progress"
      affected        = @("x/routing", "x/load", "x/zones", "x/sharding/sim")
      severity        = "High"
      fix             = "add hot-zone monopolization tests for zone-level quotas, shard activation, and unrelated-zone progress"
    }
    "ROUTEEXP-05" = [ordered]@{
      path            = "predict deterministic routes and choose actor keys to target desired shards or validators for MEV or overload"
      expected_state  = "deterministic route predictability cannot violate fairness, safety, shard bounds, or validator assignment rules"
      affected        = @("x/routing", "x/sharding/sim", "x/market")
      severity        = "Medium"
      fix             = "add route prediction abuse simulations with actor key grinding, shard distribution checks, and validator reassignment bounds"
    }
    "ROUTEEXP-06" = [ordered]@{
      path            = "create cross-zone routing loops through async messages, bounce/refund paths, or malformed destination hints"
      expected_state  = "routing loop prevention and message depth limits reject cyclic routes before queue or state mutation"
      affected        = @("x/routing", "x/mesh", "x/queue", "x/messaging")
      severity        = "Critical"
      fix             = "add cross-zone routing loop tests for async messages, bounce/refund, route depth, and replay markers"
    }
    "ROUTEEXP-07" = [ordered]@{
      path            = "cause route desync between nodes with map insertion order, local latency, local mempool order, or wall-clock inputs"
      expected_state  = "same transaction and state produce the same route on every node independent of local ordering, latency, or wall-clock"
      affected        = @("x/routing", "x/load", "app", "mempool")
      severity        = "Critical"
      fix             = "add route determinism tests across map order, local mempool order, replay, restart, and platform variation"
    }
    "ROUTEEXP-08" = [ordered]@{
      path            = "misclassify transactions with ambiguous message type strings, malformed payloads, spoofed domains, or contract/locality hints"
      expected_state  = "unknown or ambiguous transaction class rejects safely and valid class selection is stable across nodes"
      affected        = @("x/routing", "x/identity", "x/vm", "x/execution")
      severity        = "High"
      fix             = "add transaction classification tests for stable type strings, malformed payloads, domain keys, contract keys, and unknown classes"
    }
    "ROUTEEXP-09" = [ordered]@{
      path            = "use overpaid fees, non-native fees, or fee-class overflow to manipulate route destination or priority beyond caps"
      expected_state  = "fee-based routing signals are bounded, naet-only for protocol fees, and cannot override deterministic zone/shard assignment"
      affected        = @("x/routing", "x/fees", "x/reputation")
      severity        = "High"
      fix             = "add fee-based routing gaming tests for overpayment caps, non-naet rejection, fee-class overflow, and route invariance"
    }
  }
  return $overrides
}

function Get-AexsExecutionZoneAvmExploitOverrides {
  $overrides = @{
    "EXECZONEEXP-01" = [ordered]@{
      path            = "drive identical cross-zone inputs through multiple zones and attempt divergent state roots or receipts"
      expected_state  = "same finalized source roots and messages produce deterministic destination state roots, receipts, and commitment outputs"
      affected        = @("x/execution", "x/zones", "x/mesh", "x/vm")
      severity        = "Critical"
      fix             = "add cross-zone deterministic replay tests comparing state roots, receipt roots, and export/import commitments"
    }
    "EXECZONEEXP-02" = [ordered]@{
      path            = "replay finalized cross-zone messages, receipts, or proofs against the same or alternate destination zone"
      expected_state  = "single-use replay markers reject duplicate cross-zone messages before execution or value movement"
      affected        = @("x/mesh", "x/messaging", "x/queue", "x/execution")
      severity        = "Critical"
      fix             = "add cross-zone replay tests for duplicate message ids, stale receipts, wrong destination, and replay marker export/import"
    }
    "EXECZONEEXP-03" = [ordered]@{
      path            = "invoke AVM host behavior that depends on local time, randomness, map order, platform integers, or external APIs"
      expected_state  = "AVM execution rejects nondeterministic host behavior and identical replay produces identical writes and receipts"
      affected        = @("x/aetravm/avm", "x/vm", "x/execution")
      severity        = "Critical"
      fix             = "add AVM determinism tests for forbidden host calls, local time, randomness, map order, platform variation, and replay"
    }
    "EXECZONEEXP-04" = [ordered]@{
      path            = "execute the same contract call sequence on separate nodes and attempt different emitted messages, events, or writes"
      expected_state  = "contract execution trace, gas, writes, events, and emitted messages are deterministic for the same state and input"
      affected        = @("x/aetravm/avm", "x/execution", "x/events")
      severity        = "Critical"
      fix             = "add contract execution trace comparison tests with same genesis, tx sequence, and export/import replay"
    }
    "EXECZONEEXP-05" = [ordered]@{
      path            = "schedule parallel contract executions with conflicting read/write sets and attempt race-dependent final state"
      expected_state  = "scheduler detects conflicts and orders or rejects parallel work deterministically before committing writes"
      affected        = @("x/scheduler", "x/execution", "x/storage", "x/actors")
      severity        = "Critical"
      fix             = "add parallel read/write conflict tests, deterministic scheduling checks, and race-order replay fixtures"
    }
    "EXECZONEEXP-06" = [ordered]@{
      path            = "inject corrupted contract storage, zone state, queue state, or commitment roots and attempt accepted state corruption"
      expected_state  = "state corruption is rejected during validation, export/import, commitment checking, or invariant checks before commit"
      affected        = @("x/storage", "x/zones", "x/queue", "x/execution")
      severity        = "Critical"
      fix             = "add corrupted state fixtures for storage roots, zone commitments, queue entries, and contract namespace validation"
    }
    "EXECZONEEXP-07" = [ordered]@{
      path            = "force failure after partial contract writes, queued messages, emitted events, or value transfers and attempt rollback leakage"
      expected_state  = "failed execution rolls back all writes, messages, events, and value movement in the cached context"
      affected        = @("x/execution", "x/vm", "x/queue", "x/bank")
      severity        = "Critical"
      fix             = "add partial execution rollback tests across contract writes, queues, emitted messages, and bank movements"
    }
    "EXECZONEEXP-08" = [ordered]@{
      path            = "execute opcodes or host calls with nondeterministic ordering, local data, random values, or wall-clock dependencies"
      expected_state  = "nondeterministic opcodes and host calls are unavailable or reject before state mutation"
      affected        = @("x/aetravm/avm", "x/vm")
      severity        = "Critical"
      fix             = "add opcode/host allowlist tests and fuzz malformed host call selectors"
    }
    "EXECZONEEXP-09" = [ordered]@{
      path            = "submit gas-exhausting contracts, deep queries, storage writes, or message fanout to deny service"
      expected_state  = "gas limits and per-block limits stop execution deterministically without unbounded resource use or partial commits"
      affected        = @("x/vm", "x/aetravm/avm", "x/queue")
      severity        = "High"
      fix             = "add gas exhaustion benchmarks and adversarial tests for deploy, execute, query, storage writes, and emitted messages"
    }
    "EXECZONEEXP-10" = [ordered]@{
      path            = "run infinite loops, recursive calls, or bounce loops to grief validators or keep queues busy forever"
      expected_state  = "gas, instruction, depth, and per-block queue limits stop loops and preserve deterministic failure state"
      affected        = @("x/aetravm/avm", "x/queue", "x/messaging")
      severity        = "High"
      fix             = "add infinite-loop, recursive-call, and bounce-loop tests with gas/depth/per-block bounds"
    }
    "EXECZONEEXP-11" = [ordered]@{
      path            = "create storage keys that collide across contracts, zones, namespaces, or export/import boundaries"
      expected_state  = "contract storage namespaces and keys remain isolated and canonical ordering prevents collisions"
      affected        = @("x/storage", "x/vm", "x/zones")
      severity        = "Critical"
      fix             = "add storage namespace collision tests for contract id, zone id, key prefix, export/import, and bounded iteration"
    }
    "EXECZONEEXP-12" = [ordered]@{
      path            = "hijack contract upgrade by spoofing admin, governance gate, code owner, migration authority, or zero-admin path"
      expected_state  = "contract upgrade and migration require authorized admin or governance gate and reject zero or spoofed authority"
      affected        = @("x/vm", "x/aetravm/avm", "x/gov", "app/wasmconfig")
      severity        = "Critical"
      fix             = "add upgrade authority tests for admin, governance gate, code owner, zero admin, and migration-disabled params"
    }
    "EXECZONEEXP-13" = [ordered]@{
      path            = "read or write uninitialized storage slots and exploit default values, missing owner records, or absent queue markers"
      expected_state  = "uninitialized storage has explicit defaults and cannot grant ownership, funds, replay bypass, or resolver rights"
      affected        = @("x/storage", "x/vm", "x/queue", "x/identity")
      severity        = "High"
      fix             = "add uninitialized storage tests for ownership, balances, replay markers, resolver records, and contract state"
    }
    "EXECZONEEXP-14" = [ordered]@{
      path            = "construct deeply nested calls, stack frames, or parser inputs to overflow AVM stack or host recursion"
      expected_state  = "stack and recursion limits reject before panic, node crash, or partial state mutation"
      affected        = @("x/aetravm/avm", "x/vm")
      severity        = "High"
      fix             = "add stack-depth fuzz tests, parser recursion limits, and panic-safety assertions"
    }
    "EXECZONEEXP-15" = [ordered]@{
      path            = "attempt sandbox escape through forbidden host function, direct state mutation, filesystem/network access, or cross-contract state write"
      expected_state  = "AVM sandbox denies external APIs, direct foreign state mutation, filesystem/network access, and unauthorized host functions"
      affected        = @("x/aetravm/avm", "x/vm", "x/storage")
      severity        = "Critical"
      fix             = "add sandbox escape tests for host function allowlist, foreign state mutation, external API denial, and filesystem/network denial"
    }
  }
  return $overrides
}

function Get-AexsComputeShardExploitOverrides {
  $overrides = @{
    "SHARDEXP-01" = [ordered]@{
      path            = "skew routing keys and load inputs to create shard partition imbalance across active shards"
      expected_state  = "deterministic shard assignment preserves bounded distribution rules and records imbalance for audit without route desync"
      affected        = @("x/sharding/sim", "x/routing", "x/load")
      severity        = "High"
      fix             = "add shard distribution simulations for skewed keys, routing epochs, active shard counts, and load windows"
    }
    "SHARDEXP-02" = [ordered]@{
      path            = "starve a shard with priority ordering, queue delays, epoch changes, or validator reassignment"
      expected_state  = "valid shard work makes bounded deterministic progress and starvation attempts fail invariant checks"
      affected        = @("x/sharding/sim", "x/scheduler", "x/queue")
      severity        = "High"
      fix             = "add shard starvation tests for queues, priority, routing epochs, validator reassignment, and progress bounds"
    }
    "SHARDEXP-03" = [ordered]@{
      path            = "overflow a shard with excessive messages, state, or queue depth to collapse split/merge processing"
      expected_state  = "per-shard queue, state, and processing limits reject or defer overflow without corrupting state"
      affected        = @("x/sharding/sim", "x/queue", "x/storage")
      severity        = "High"
      fix             = "add shard overflow tests for queue depth, state size, split/merge limits, and bounded export/import"
    }
    "SHARDEXP-04" = [ordered]@{
      path            = "commit conflicting cross-shard messages, receipts, or state roots and attempt cross-shard inconsistency"
      expected_state  = "cross-shard commitments, receipts, and message ordering validate deterministically before finalization"
      affected        = @("x/sharding/sim", "x/mesh", "x/messaging")
      severity        = "Critical"
      fix             = "add cross-shard consistency tests for message order, receipt roots, source finality, and replay markers"
    }
    "SHARDEXP-05" = [ordered]@{
      path            = "spoof load inputs to trigger shard activation, deactivation, split, or merge at attacker-chosen heights"
      expected_state  = "shard activation uses deterministic normalized load, EMA, MAX_DELTA, and cooldown rules"
      affected        = @("x/sharding/sim", "x/load")
      severity        = "Critical"
      fix             = "add load-spoof shard activation tests for EMA, thresholds, MAX_DELTA, cooldown, and export/import replay"
    }
    "SHARDEXP-06" = [ordered]@{
      path            = "duplicate shard ids, child shard outputs, message queues, or receipts during split/merge transitions"
      expected_state  = "shard split/merge assigns each message, receipt, and child shard exactly once with canonical ids"
      affected        = @("x/sharding/sim", "x/mesh", "x/queue")
      severity        = "Critical"
      fix             = "add shard duplication tests for child ids, queues, receipts, split/merge, and exact-once partitioning"
    }
    "SHARDEXP-07" = [ordered]@{
      path            = "split state into child shards and attempt missing, duplicated, or inconsistent state partitions"
      expected_state  = "state split produces deterministic child roots and recombines without missing or duplicated state"
      affected        = @("x/sharding/sim", "x/storage", "x/zones")
      severity        = "Critical"
      fix             = "add state split/merge consistency tests for roots, key ranges, queue partitions, and export/import"
    }
    "SHARDEXP-08" = [ordered]@{
      path            = "execute conflicting transactions in parallel shards with overlapping actors, contracts, or storage keys"
      expected_state  = "scheduler detects parallel execution collisions and orders or rejects work deterministically"
      affected        = @("x/sharding/sim", "x/scheduler", "x/execution")
      severity        = "Critical"
      fix             = "add parallel collision tests for read/write sets, actor locality, contract storage, and route assignment"
    }
    "SHARDEXP-09" = [ordered]@{
      path            = "manipulate scheduling with priority classes, task ids, dependency edges, or tie-breakers across shards"
      expected_state  = "same shard tasks and state produce the same schedule with bounded priority and canonical tie-breakers"
      affected        = @("x/scheduler", "x/sharding/sim", "x/routing")
      severity        = "High"
      fix             = "add scheduling manipulation tests for duplicate task ids, dependency conflicts, priority gaming, and tie-break determinism"
    }
    "SHARDEXP-10" = [ordered]@{
      path            = "flood shard queues with messages, delayed tasks, bounces, or refunds to exceed per-block processing limits"
      expected_state  = "queue flooding is bounded by deterministic per-shard and per-block limits without message loss or double processing"
      affected        = @("x/sharding/sim", "x/queue", "x/messaging")
      severity        = "High"
      fix             = "add queue flooding tests for per-shard limits, delayed execution, bounce/refund, message ordering, and export/import"
    }
  }
  return $overrides
}

function Get-AexsMeshCrossZoneExploitOverrides {
  $overrides = @{
    "MESHEXP-01" = [ordered]@{
      path            = "replay finalized cross-zone messages against the same destination or a different zone after receipt creation"
      expected_state  = "single-use replay markers reject duplicate cross-zone messages before execution, receipt creation, or asset movement"
      affected        = @("x/mesh", "x/messaging", "x/queue", "x/execution")
      severity        = "Critical"
      fix             = "add mesh replay tests for message ids, finality references, receipt markers, destination binding, and export/import"
    }
    "MESHEXP-02" = [ordered]@{
      path            = "delay relay delivery to reorder valid cross-zone messages around timeout, finality, or queue boundaries"
      expected_state  = "message delay cannot change deterministic ordering, timeout handling, receipt state, or refund eligibility"
      affected        = @("x/mesh", "x/queue", "x/messaging")
      severity        = "High"
      fix             = "add delayed relay simulations for finality height, timeout windows, canonical ordering, and deterministic refunds"
    }
    "MESHEXP-03" = [ordered]@{
      path            = "submit cross-zone messages with conflicting source heights, shard ids, message ids, or sequences to attack ordering"
      expected_state  = "mesh ordering remains canonical by source finality height, source zone, source shard, message id, destination, and sequence"
      affected        = @("x/mesh", "x/messaging", "x/sharding/sim", "x/queue")
      severity        = "Critical"
      fix             = "add ordering attack tests for source finality, source zone/shard, message id, destination, sequence, and tie-breaks"
    }
    "MESHEXP-04" = [ordered]@{
      path            = "duplicate asset commitments across source and destination zones during transfer, bounce, or refund processing"
      expected_state  = "asset commitments are consumed exactly once and destination mint/release cannot exceed finalized source lock/burn"
      affected        = @("x/mesh", "x/bank", "x/queue")
      severity        = "Critical"
      fix             = "add cross-zone asset conservation tests for lock/burn, release/mint, bounce, refund, and replay markers"
    }
    "MESHEXP-05" = [ordered]@{
      path            = "double spend the same cross-zone value through duplicate message, duplicate receipt, stale proof, or bounce/refund race"
      expected_state  = "the same value cannot be spent twice across zones, receipts, bounces, refunds, or replayed proofs"
      affected        = @("x/mesh", "x/messaging", "x/queue", "x/bank")
      severity        = "Critical"
      fix             = "add double-spend tests for duplicate messages, duplicate receipts, stale proofs, bounce/refund, and asset commitments"
    }
    "MESHEXP-06" = [ordered]@{
      path            = "forge source proof, receipt proof, commitment root, or destination binding for a cross-zone action"
      expected_state  = "proof verification rejects forged roots, malformed proofs, wrong destination bindings, and unfinalized source commitments"
      affected        = @("x/mesh", "x/zones", "x/storage", "x/messaging")
      severity        = "Critical"
      fix             = "add proof forgery tests for root formats, commitment chains, destination binding, malformed proofs, and source finality"
    }
    "MESHEXP-07" = [ordered]@{
      path            = "simulate relay censorship by withholding valid messages or receipts to starve a zone or defer refunds"
      expected_state  = "relay censorship cannot violate safety and liveness gaps are bounded by deterministic timeout, retry, and refund rules"
      affected        = @("x/mesh", "x/queue", "x/messaging")
      severity        = "High"
      fix             = "add relay censorship simulations with timeout, retry, refund, queue progress, and operator diagnostic evidence"
    }
    "MESHEXP-08" = [ordered]@{
      path            = "starve mesh messages through priority abuse, queue depth, destination congestion, or shard assignment skew"
      expected_state  = "valid mesh messages make bounded deterministic progress and starvation attempts preserve replay and refund invariants"
      affected        = @("x/mesh", "x/queue", "x/routing", "x/sharding/sim")
      severity        = "High"
      fix             = "add message starvation tests for priority, queue depth, destination congestion, routing keys, and per-block limits"
    }
    "MESHEXP-09" = [ordered]@{
      path            = "mix source and destination commitments from different finality heights to create finality mismatch"
      expected_state  = "mesh rejects finality mismatch before destination execution, receipt creation, or value movement"
      affected        = @("x/mesh", "x/zones", "x/sharding/sim")
      severity        = "Critical"
      fix             = "add finality mismatch tests for source roots, destination heights, stale commitments, and deterministic rejection"
    }
    "MESHEXP-10" = [ordered]@{
      path            = "replay stale receipts after timeout, refund, bounce, export/import, or destination retry"
      expected_state  = "receipt state is single-use and stale receipt replay cannot unlock assets, rerun execution, or clear replay markers"
      affected        = @("x/mesh", "x/queue", "x/messaging", "x/storage")
      severity        = "Critical"
      fix             = "add stale receipt replay tests across timeout, refund, bounce, retry, export/import, and replay marker persistence"
    }
  }
  return $overrides
}

function Get-AexsIdentityDomainExploitOverrides {
  $overrides = @{
    "IDENTEXP-01" = [ordered]@{
      path            = "hijack a .aet domain by overwriting resolver records without current domain owner authorization"
      expected_state  = "resolver updates require current owner authority and rejected hijacks leave resolver, owner, reverse, and NFT state unchanged"
      affected        = @("x/identity", "x/storage", "x/indexer")
      severity        = "Critical"
      fix             = "add resolver overwrite tests for owner checks, NFT owner consistency, reverse lookup, events, and rejected no-commit paths"
    }
    "IDENTEXP-02" = [ordered]@{
      path            = "take over an expired domain directly without auction, renewal rules, or canonical lifecycle transition"
      expected_state  = "expired domains follow deterministic auction or renewal paths and cannot be reassigned by direct state mutation"
      affected        = @("x/identity", "x/gov", "x/storage")
      severity        = "High"
      fix             = "add expired-domain takeover tests for lifecycle state, renewal window, auction entry, owner persistence, and export/import"
    }
    "IDENTEXP-03" = [ordered]@{
      path            = "manipulate sealed-bid auction order, reveal height, refund receipts, or tie-breaker commitment hash"
      expected_state  = "auction winner and refunds are deterministic by bid, reveal height, and commitment hash with no stolen losing bids"
      affected        = @("x/identity", "x/bank", "x/queue")
      severity        = "High"
      fix             = "add auction manipulation tests for commit/reveal, tie-breakers, bid escrow, losing refunds, and replay"
    }
    "IDENTEXP-04" = [ordered]@{
      path            = "spoof resolver target with malformed address, zero address, fake contract, or wrong zone endpoint"
      expected_state  = "resolver records reject malformed or zero targets and cannot route funds or calls to spoofed destinations"
      affected        = @("x/identity", "x/routing", "x/vm")
      severity        = "Critical"
      fix             = "add resolver spoofing tests for address, contract, zone endpoint, zero target, and payment-before-resolution failure"
    }
    "IDENTEXP-05" = [ordered]@{
      path            = "create subdomain collision through normalization, parent policy bypass, or duplicate child ownership"
      expected_state  = "subdomain records are unique after normalization and issuance requires parent owner authorization"
      affected        = @("x/identity", "x/storage")
      severity        = "High"
      fix             = "add subdomain collision tests for normalization, parent policy, child owner resolver control, and export/import"
    }
    "IDENTEXP-06" = [ordered]@{
      path            = "poison reverse lookup by binding an address to a domain without address owner authorization"
      expected_state  = "reverse lookup updates require address owner authorization and remain consistent with forward resolver state"
      affected        = @("x/identity", "x/auth", "x/indexer")
      severity        = "High"
      fix             = "add reverse lookup poisoning tests for signer authority, forward/reverse consistency, zero address, and cache rebuild"
    }
    "IDENTEXP-07" = [ordered]@{
      path            = "race domain binding updates across owner transfer, resolver update, NFT transfer, and index refresh"
      expected_state  = "domain binding changes are atomic and transfer invalidates unauthorized pending resolver or reverse updates"
      affected        = @("x/identity", "x/events", "x/indexer")
      severity        = "High"
      fix             = "add binding race tests for transfer plus resolver updates, pending operations, event order, and index rebuildability"
    }
    "IDENTEXP-08" = [ordered]@{
      path            = "poison index-layer cache with stale resolver, fake event, duplicate domain, or deleted domain state"
      expected_state  = "index output remains non-authoritative and can be rebuilt from committed identity state and events"
      affected        = @("x/indexer", "x/identity", "x/events")
      severity        = "Medium"
      fix             = "add index cache poisoning tests for stale resolver, fake events, duplicate records, deletion, and rebuild from state"
    }
    "IDENTEXP-09" = [ordered]@{
      path            = "inject fake domain resolution through malformed name, unicode/confusable normalization, or forged resolver response"
      expected_state  = "domain normalization and resolver verification reject fake resolution before routing, payment, or contract execution"
      affected        = @("x/identity", "x/routing", "x/execution")
      severity        = "Critical"
      fix             = "add fake resolution tests for normalization, confusables policy, forged resolver response, and funds-before-resolution safety"
    }
    "IDENTEXP-10" = [ordered]@{
      path            = "create multi-resolver inconsistency between address, contract, zone endpoint, reverse lookup, and index records"
      expected_state  = "multi-resolver state updates remain canonical and export/import preserves consistent forward, reverse, and index rebuild state"
      affected        = @("x/identity", "x/indexer", "x/storage")
      severity        = "High"
      fix             = "add multi-resolver consistency tests for forward records, reverse records, zone endpoints, index rebuild, and export/import"
    }
  }
  return $overrides
}

function Get-AexsGovernanceExploitOverrides {
  $overrides = @{
    "GOVEXP-01" = [ordered]@{
      path            = "capture governance with manipulated voting power through delegation timing, validator concentration, or stale snapshots"
      expected_state  = "governance voting power is derived from deterministic staking snapshots and cannot be inflated by timing loops"
      affected        = @("x/gov", "x/staking", "x/distribution")
      severity        = "Critical"
      fix             = "add governance capture simulations for voting power snapshots, delegation timing, validator concentration, and replay"
    }
    "GOVEXP-02" = [ordered]@{
      path            = "flood proposals with low deposits, malformed metadata, duplicate proposals, or expensive tally/query state"
      expected_state  = "proposal spam is bounded by deposit, metadata, pagination, and deterministic processing limits"
      affected        = @("x/gov", "x/fees", "x/bank")
      severity        = "High"
      fix             = "add proposal spam tests for min deposit, malformed metadata, duplicate proposals, pagination, and state growth"
    }
    "GOVEXP-03" = [ordered]@{
      path            = "abuse emergency parameters to bypass hard bounds for fees, staking, slashing, minting, routing, or VM gates"
      expected_state  = "governance parameter updates are bounded by protocol hard limits and delayed execution rules"
      affected        = @("x/gov", "x/fees", "x/staking", "x/slashing", "x/vm")
      severity        = "Critical"
      fix             = "add emergency param abuse tests for bounds, delayed execution, authority, and feature-gate safety"
    }
    "GOVEXP-04" = [ordered]@{
      path            = "hijack upgrade plan, handler name, module version, or contract gate through unauthorized proposal execution"
      expected_state  = "upgrade execution requires authorized governance flow, matching handler, monotonic version, and deterministic migration"
      affected        = @("x/gov", "x/upgrade", "app", "x/vm")
      severity        = "Critical"
      fix             = "add upgrade hijack tests for plan authority, handler names, version checks, migration state, and rollback denial"
    }
    "GOVEXP-05" = [ordered]@{
      path            = "exploit delayed execution by changing dependencies, voting power, params, or module state between pass and execution"
      expected_state  = "delayed governance execution validates all dependencies and executes exactly once at the scheduled height"
      affected        = @("x/gov", "x/params", "x/staking", "app")
      severity        = "High"
      fix             = "add delayed execution tests for dependency drift, execution height, replay, params validation, and no-op rejection"
    }
    "GOVEXP-06" = [ordered]@{
      path            = "replay governance proposals, votes, deposits, or execution messages across chain id, export/import, or upgrade boundaries"
      expected_state  = "governance replay is rejected by proposal ids, vote records, sequence checks, chain id, and migration-safe state"
      affected        = @("x/gov", "x/auth", "x/upgrade")
      severity        = "High"
      fix             = "add governance replay tests for proposal ids, votes, deposits, execution, chain id, and export/import"
    }
    "GOVEXP-07" = [ordered]@{
      path            = "front-run proposals with conflicting params, deposits, metadata, or upgrade plans to alter execution order"
      expected_state  = "proposal ordering, deposits, voting periods, and execution are deterministic and cannot be reordered by front-running"
      affected        = @("x/gov", "mempool", "x/fees")
      severity        = "Medium"
      fix             = "add proposal front-running tests for ordering, conflicting params, deposits, mempool priority, and tie-breaks"
    }
    "GOVEXP-08" = [ordered]@{
      path            = "loop staking delegation, reward withdrawal, redelegation, or unbonding to manipulate governance voting power"
      expected_state  = "staking-loop voting power cannot exceed deterministic staking keeper state or bypass unbonding/slashing risk"
      affected        = @("x/gov", "x/staking", "x/distribution", "x/slashing")
      severity        = "High"
      fix             = "add staking-loop governance tests for delegate/redelegate/unbond/reward sequences and voting power snapshots"
    }
    "GOVEXP-09" = [ordered]@{
      path            = "grief parameters by toggling fee, load, routing, VM, slashing, or identity bounds at unsafe frequencies"
      expected_state  = "parameter updates obey hard bounds, delayed activation, cooldowns, and rollback-safe migration rules"
      affected        = @("x/gov", "x/fees", "x/load", "x/routing", "x/identity")
      severity        = "High"
      fix             = "add parameter griefing tests for hard bounds, cooldowns, delayed activation, rollback safety, and operator visibility"
    }
  }
  return $overrides
}

function Get-AexsGenesisUpgradeStateExploitOverrides {
  $overrides = @{
    "STATEEXP-01" = [ordered]@{
      path            = "inject malformed genesis accounts, balances, params, module state, zone roots, or custom module records"
      expected_state  = "genesis validation rejects malformed state before node start and never panics on corrupt but parseable input"
      affected        = @("app", "x/bank", "x/fees", "avm-dex-contract")
      severity        = "Critical"
      fix             = "add malformed genesis fixtures for accounts, balances, params, custom modules, roots, and panic-free validation"
    }
    "STATEEXP-02" = [ordered]@{
      path            = "tamper exported state by changing balances, denoms, params, commitments, module versions, or ordering"
      expected_state  = "export/import validation detects tampering and deterministic export ordering is stable across runs"
      affected        = @("app", "x/storage", "x/upgrade", "x/zones")
      severity        = "Critical"
      fix             = "add export tampering tests for balances, denoms, params, roots, module versions, ordering, and replay"
    }
    "STATEEXP-03" = [ordered]@{
      path            = "rollback an upgrade plan, module version, store migration, or feature gate to re-enable unsafe state"
      expected_state  = "upgrade versions are monotonic and rollback attempts cannot mutate committed state or bypass handlers"
      affected        = @("x/upgrade", "app", "x/gov")
      severity        = "Critical"
      fix             = "add upgrade rollback tests for monotonic versions, handler names, feature gates, store migrations, and app hash"
    }
    "STATEEXP-04" = [ordered]@{
      path            = "corrupt partial migration by failing after some module writes, version updates, or store key changes"
      expected_state  = "migration executes atomically or fails before commit with deterministic recovery and no partial state roots"
      affected        = @("x/upgrade", "app", "x/storage")
      severity        = "Critical"
      fix             = "add partial migration corruption tests with injected failures, cached context semantics, version maps, and export/import"
    }
    "STATEEXP-05" = [ordered]@{
      path            = "bypass module initialization by omitting module genesis, duplicating module keys, or changing init order"
      expected_state  = "module initialization order and default genesis validation reject missing, duplicate, or unordered module state"
      affected        = @("app", "module manager", "x/params")
      severity        = "High"
      fix             = "add module init bypass tests for missing genesis, duplicate keys, init order, defaults, and module manager wiring"
    }
    "STATEEXP-06" = [ordered]@{
      path            = "inject hidden privileged account, module account, authority, admin, or mint permission into genesis/export state"
      expected_state  = "privileged accounts and module permissions are explicitly validated and unauthorized hidden authorities are rejected"
      affected        = @("app", "x/auth", "x/bank", "x/fees")
      severity        = "Critical"
      fix             = "add privileged account injection tests for module accounts, authority fields, admin state, mint permissions, and export/import"
    }
    "STATEEXP-07" = [ordered]@{
      path            = "bypass InitGenesis validation with duplicate ids, duplicate denoms, invalid reserves, invalid params, or nil-like records"
      expected_state  = "InitGenesis validates duplicates, params, reserves, denoms, and nil-like records before state writes"
      affected        = @("app", "avm-dex-contract", "x/fees", "x/identity")
      severity        = "Critical"
      fix             = "add InitGenesis bypass tests for duplicates, invalid params, reserves, denoms, nil records, and panic-free errors"
    }
    "STATEEXP-08" = [ordered]@{
      path            = "exploit version mismatch between app version, module consensus version, store version, or proto schema"
      expected_state  = "version mismatches fail safely or run explicit migrations and cannot produce divergent app hashes"
      affected        = @("app", "x/upgrade", "proto", "x/storage")
      severity        = "High"
      fix             = "add version mismatch tests for app/module/store/proto versions, no-op migrations, and deterministic app hash"
    }
    "STATEEXP-09" = [ordered]@{
      path            = "create state root collision through malformed keys, canonical encoding ambiguity, or commitment root spoofing"
      expected_state  = "state roots use canonical encoding and reject malformed keys or spoofed roots before commitment acceptance"
      affected        = @("x/storage", "app", "x/zones", "x/mesh")
      severity        = "Critical"
      fix             = "add state root collision tests for key encoding, root format, commitment hashing, and malformed import fixtures"
    }
    "STATEEXP-10" = [ordered]@{
      path            = "poison snapshot or state-sync artifacts with stale chunks, wrong app hash, missing modules, or corrupted metadata"
      expected_state  = "snapshot restore verifies app hash, chunk integrity, module state, and metadata before serving or accepting state"
      affected        = @("app", "x/storage", "state sync", "snapshots")
      severity        = "Critical"
      fix             = "add snapshot poisoning tests for chunk integrity, app hash, module completeness, metadata, and restart persistence"
    }
  }
  return $overrides
}

function Get-AexsMempoolNetworkExploitOverrides {
  $overrides = @{
    "NETEXP-01" = [ordered]@{
      path            = "flood mempool with low-fee, malformed, replayed, oversized, or expensive CheckTx transactions"
      expected_state  = "mempool flooding is bounded by fee, gas, size, replay, and CheckTx validation without consensus state mutation"
      affected        = @("mempool", "x/fees", "x/auth", "app ante")
      severity        = "High"
      fix             = "add mempool flooding tests for malformed txs, replay, oversized payloads, fee bounds, and CheckTx cost"
    }
    "NETEXP-02" = [ordered]@{
      path            = "game transaction prioritization with fee overpayment, reputation spoofing, local ordering, or routing hints"
      expected_state  = "transaction priority is deterministic, bounded, and independent of node-local ordering or spoofed hints"
      affected        = @("mempool", "x/fees", "x/reputation", "x/routing")
      severity        = "High"
      fix             = "add prioritization gaming tests for fee caps, reputation bounds, local order, routing hints, and tx hash tie-breaks"
    }
    "NETEXP-03" = [ordered]@{
      path            = "poison gossip with malformed tx bytes, fake peers, duplicate blocks, stale evidence, or invalid proposal data"
      expected_state  = "gossip poisoning is rejected at decode, evidence, proposal, and block validation boundaries without panics"
      affected        = @("CometBFT P2P", "mempool", "app", "x/slashing")
      severity        = "High"
      fix             = "add gossip poisoning simulations for malformed bytes, duplicate data, stale evidence, proposal validation, and panic safety"
    }
    "NETEXP-04" = [ordered]@{
      path            = "eclipse a node with controlled peers to delay txs, blocks, evidence, or state-sync data"
      expected_state  = "node eclipse affects liveness diagnostics only and cannot change committed state or deterministic validation"
      affected        = @("CometBFT P2P", "mempool", "state sync")
      severity        = "Medium"
      fix             = "add eclipse simulations for peer diversity, delayed tx/block/evidence propagation, diagnostics, and safe recovery"
    }
    "NETEXP-05" = [ordered]@{
      path            = "partition P2P network to delay finality, isolate validators, or create inconsistent mempool views"
      expected_state  = "P2P partition cannot violate deterministic finality and recovery preserves app hash consistency"
      affected        = @("CometBFT consensus", "CometBFT P2P", "mempool")
      severity        = "High"
      fix             = "add P2P partition simulations for finality delay, validator isolation, mempool divergence, and recovery app hash"
    }
    "NETEXP-06" = [ordered]@{
      path            = "delay block propagation to trigger missed votes, stale proposals, evidence timing gaps, or fork-choice pressure"
      expected_state  = "block propagation delay affects liveness only and slashing/evidence handling remains objective and deterministic"
      affected        = @("CometBFT consensus", "x/slashing", "mempool")
      severity        = "High"
      fix             = "add block delay simulations for vote timing, stale proposals, evidence windows, slashing, and finality recovery"
    }
    "NETEXP-07" = [ordered]@{
      path            = "reorder transactions across local mempools, proposal construction, or priority queues to cause state or event divergence"
      expected_state  = "proposal ordering and app execution are deterministic for accepted block order and no local mempool order leaks into AppHash"
      affected        = @("mempool", "app", "x/routing", "x/events")
      severity        = "Critical"
      fix             = "add transaction reordering tests for local mempool order, proposal order, deterministic execution, and event stability"
    }
    "NETEXP-08" = [ordered]@{
      path            = "exploit network latency as an input to load, routing, priority, or consensus-critical decisions"
      expected_state  = "network latency is never a consensus-critical input for load, routing, fees, or state transitions"
      affected        = @("x/load", "x/routing", "mempool", "app")
      severity        = "Critical"
      fix             = "add latency exploitation tests proving local latency cannot affect LOAD_SCORE, route decisions, fees, or AppHash"
    }
    "NETEXP-09" = [ordered]@{
      path            = "exhaust bandwidth with oversized messages, tx gossip, evidence spam, state-sync chunks, or peer churn"
      expected_state  = "bandwidth exhaustion is bounded by message size, rate limits, peer limits, and decode rejection without state mutation"
      affected        = @("CometBFT P2P", "mempool", "state sync")
      severity        = "High"
      fix             = "add bandwidth exhaustion tests for tx size, evidence size, state-sync chunks, peer churn, and decode safety"
    }
    "NETEXP-10" = [ordered]@{
      path            = "target peers or validators based on routing keys, shard assignment, validator set, or relay role"
      expected_state  = "peer targeting cannot bias routing, validator assignment, finality, or committed state"
      affected        = @("CometBFT P2P", "x/routing", "x/sharding/sim", "x/staking")
      severity        = "High"
      fix             = "add peer targeting simulations for validator assignment, routing keys, shard roles, relay roles, and finality safety"
    }
  }
  return $overrides
}

function Get-AexsCombinedFullStackExploitOverrides {
  $overrides = @{
    "FULLSTACKEXP-01" = [ordered]@{
      path            = "coordinate spam bursts with routing-key grinding to overload zones and shards while preserving low apparent fees"
      expected_state  = "spam plus routing attacks remain bounded by fee policy, LOAD_SCORE smoothing, shard activation, and priority caps"
      affected        = @("x/load", "x/routing", "x/fees", "x/sharding/sim", "mempool")
      severity        = "Critical"
      fix             = "add coordinated spam-routing simulations with fee caps, EMA, MAX_DELTA, shard activation, and route determinism"
    }
    "FULLSTACKEXP-02" = [ordered]@{
      path            = "combine load threshold manipulation with governance param changes to destabilize fees, routing, or shard activation"
      expected_state  = "load plus governance changes obey delayed activation, hard bounds, cooldowns, and deterministic replay"
      affected        = @("x/load", "x/gov", "x/fees", "x/routing", "x/sharding/sim")
      severity        = "Critical"
      fix             = "add load-governance combined tests for thresholds, delayed params, hard bounds, cooldowns, and export/import"
    }
    "FULLSTACKEXP-03" = [ordered]@{
      path            = "combine DEX swaps, mempool reordering, fee priority, and routing locality to drain liquidity or desync reserves"
      expected_state  = "DEX accounting and routing remain deterministic and failed or reordered txs cannot violate pool invariants"
      affected        = @("avm-dex-contract", "mempool", "x/routing", "x/fees", "x/bank")
      severity        = "Critical"
      fix             = "add DEX-mempool-routing scenarios for swap ordering, fee priority, reserve checks, slippage, and failed bank movement"
    }
    "FULLSTACKEXP-04" = [ordered]@{
      path            = "coordinate validator collusion with delayed slashing evidence, redelegation, unbonding, or governance timing"
      expected_state  = "slashing remains objective and collusion cannot evade penalties through evidence delay or staking lifecycle timing"
      affected        = @("x/slashing", "x/staking", "x/gov", "CometBFT consensus")
      severity        = "Critical"
      fix             = "add validator-collusion slashing tests for delayed evidence, redelegation, unbonding, governance timing, and tombstone state"
    }
    "FULLSTACKEXP-05" = [ordered]@{
      path            = "coordinate cross-zone value extraction through mesh replay, stale receipt, proof forgery, and queue refund timing"
      expected_state  = "cross-zone value remains conserved and replay, receipt, proof, bounce, and refund paths cannot extract extra funds"
      affected        = @("x/mesh", "x/queue", "x/messaging", "x/bank")
      severity        = "Critical"
      fix             = "add cross-zone value extraction tests for replay markers, receipts, proofs, bounce/refund, and asset conservation"
    }
    "FULLSTACKEXP-06" = [ordered]@{
      path            = "hijack identity resolution and routing to redirect payments, contract calls, or zone endpoints"
      expected_state  = "identity and routing authority checks prevent resolver hijack from changing route, payment, or contract destination"
      affected        = @("x/identity", "x/routing", "x/execution", "x/bank", "x/indexer")
      severity        = "Critical"
      fix             = "add identity-routing hijack tests for resolver authority, index rebuild, route locality, payment resolution, and contract calls"
    }
    "FULLSTACKEXP-07" = [ordered]@{
      path            = "cascade shard overload with fee manipulation, priority gaming, queue flooding, and load slow-poison"
      expected_state  = "shard overload plus fee manipulation is bounded by EMA, MAX_DELTA, fee caps, queue limits, and starvation checks"
      affected        = @("x/sharding/sim", "x/load", "x/fees", "x/queue", "x/routing")
      severity        = "Critical"
      fix             = "add shard-fee cascade simulations for overload, fee caps, queue flooding, slow-poison, and deterministic recovery"
    }
    "FULLSTACKEXP-08" = [ordered]@{
      path            = "combine consensus pressure with mempool flooding, block delay, evidence spam, and malformed proposal data"
      expected_state  = "consensus plus mempool denial-of-service cannot violate finality safety, evidence validity, or app hash determinism"
      affected        = @("CometBFT consensus", "mempool", "x/slashing", "app")
      severity        = "Critical"
      fix             = "add consensus-mempool hybrid tests for flooding, block delay, evidence spam, proposal validation, and recovery app hash"
    }
    "FULLSTACKEXP-09" = [ordered]@{
      path            = "combine economic manipulation with staking starvation, fee griefing, reward timing, and delegation loops"
      expected_state  = "economic plus staking attacks cannot starve honest stake, inflate rewards, bypass fees, or evade unbonding/slashing risk"
      affected        = @("x/staking", "x/distribution", "x/fees", "x/gov", "x/bank")
      severity        = "Critical"
      fix             = "add economic-staking starvation tests for fees, rewards, delegation loops, unbonding, slashing, and validator-set updates"
    }
    "FULLSTACKEXP-10" = [ordered]@{
      path            = "destabilize the full stack with combined governance, load, routing, mesh, identity, DEX, VM, and network faults"
      expected_state  = "full-stack destabilization fails closed with deterministic state, no unauthorized value movement, and triaged exploit evidence"
      affected        = @("app", "x/gov", "x/load", "x/routing", "x/mesh", "x/identity", "avm-dex-contract", "x/vm")
      severity        = "Critical"
      fix             = "add full-stack chaos scenarios with deterministic seeds, invariant registry, state diffs, minimized exploits, and rollback checks"
    }
  }
  return $overrides
}

function Get-AexsExploitRecordsForSection {
  param(
    [string]$Text,
    [string]$CampaignId,
    [int]$SectionNumber,
    [string]$SectionTitle,
    [string]$IdPrefix,
    [string]$SeedNamespace,
    [hashtable]$Overrides,
    [string[]]$DefaultAffectedModules,
    [string]$DefaultSeverity
  )
  $section = Get-AexsMarkdownSection -Text $Text -Heading "Exploit Task Catalog"
  if ([string]::IsNullOrWhiteSpace($section)) {
    return @()
  }

  $capture = $false
  $items = @()
  $headingPattern = "^###\s+$SectionNumber\.\s+$([regex]::Escape($SectionTitle))\s*$"
  foreach ($line in ($section -split "`r?`n")) {
    if ($line -match $headingPattern) {
      $capture = $true
      continue
    }
    if ($capture -and $line -match '^###\s+') {
      break
    }
    if ($capture -and $line -match '^- \[ \]\s+(.+?)\s*$') {
      $items += $Matches[1].Trim().TrimEnd(".")
    }
  }

  $records = @()
  for ($i = 0; $i -lt $items.Count; $i++) {
    $id = "{0}-{1:00}" -f $IdPrefix, ($i + 1)
    $description = $items[$i]
    $override = if ($null -ne $Overrides -and $Overrides.ContainsKey($id)) { $Overrides[$id] } else { $null }
    $seedHash = (Get-AexsSha256Hex -Text "$CampaignId|$SeedNamespace|$id|$description").Substring(0, 16)
    $seed = "aexs-$($id.ToLowerInvariant())-$seedHash"
    $affected = if ($null -ne $override -and $override.Contains("affected")) { @($override["affected"]) } else { @($DefaultAffectedModules) }
    $path = Get-AexsOverrideValue -Override $override -Field "path" -Fallback $description
    $expected = Get-AexsOverrideValue -Override $override -Field "expected_state" -Fallback "exploit attempt must not violate protocol invariants or mutate state outside authorized transitions"
    $severity = Get-AexsOverrideValue -Override $override -Field "severity" -Fallback $DefaultSeverity
    $fix = Get-AexsOverrideValue -Override $override -Field "fix" -Fallback "add deterministic replay, adversarial simulation, invariant coverage, and regression tests for this exploit path"

    $records += [ordered]@{
      exploit_id         = $id
      category           = $SectionTitle
      description        = $description
      exploit_path       = $path
      seed               = $seed
      step_list          = @(
        "Run AEXS exploit scenario $id",
        "Use seed $seed",
        "Construct the adversarial sequence for $SectionTitle",
        "Record expected state before execution",
        "Record actual state after execution",
        "If exploit succeeds, minimize the sequence and write AUDIT_RESULT.md"
      )
      expected_state     = $expected
      actual_state       = "not_executed_preflight"
      affected_modules   = $affected
      severity           = $severity
      fix_recommendation = $fix
      status             = "planned_not_executed"
      valid              = $true
      invalid_reasons    = @()
    }
  }
  return $records
}

function Test-AexsExploitRecord {
  param([object]$Record)
  $reasons = @()
  foreach ($field in @(
      "exploit_id",
      "category",
      "description",
      "exploit_path",
      "seed",
      "expected_state",
      "actual_state",
      "severity",
      "fix_recommendation",
      "status"
    )) {
    if ([string]::IsNullOrWhiteSpace([string]$Record[$field])) {
      $reasons += "missing $field"
    }
  }
  if (@($Record["step_list"]).Count -eq 0) {
    $reasons += "missing step_list"
  }
  if (@($Record["affected_modules"]).Count -eq 0) {
    $reasons += "missing affected_modules"
  }
  if ([string]$Record["severity"] -notin @("Critical", "High", "Medium", "Low")) {
    $reasons += "invalid severity"
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
  [ordered]@{ Module = "avm-dex-contract"; Label = '`avm-dex-contract`'; Prefix = "DEX"; Value = $true; EvidenceRoots = @("avm-dex-contract", "tests\adversarial", "tests\e2e", "docs\architecture\dex-direction.md"); EvidenceTerms = @("dex", "pool", "swap", "liquidity", "reserve") },
  [ordered]@{ Module = "x/identity"; Label = '`x/identity`'; Prefix = "ID"; Value = $true; EvidenceRoots = @("x\identity", "tests\adversarial", "docs\architecture\aetra-modular-execution-os.md"); EvidenceTerms = @("identity", ".aet", "domain", "resolver") },
  [ordered]@{ Module = "x/reputation"; Label = '`x/reputation`'; Prefix = "REP"; Value = $true; EvidenceRoots = @("x\reputation", "docs\module-boundaries.md", "docs\test-production-gates.md"); EvidenceTerms = @("reputation", "score", "rate limit", "priority") },
  [ordered]@{ Module = "x/execution"; Label = '`x/execution`'; Prefix = "EXEC"; Value = $true; EvidenceRoots = @("x\execution", "docs\architecture\execution-os.md", "docs\module-boundaries.md"); EvidenceTerms = @("execution", "dispatch", "route", "receipt") },
  [ordered]@{ Module = "x/vm"; Label = '`x/vm` / AVM'; Prefix = "VM"; Value = $true; EvidenceRoots = @("x\vm", "x\aetravm", "docs\architecture\avm.md", "docs\architecture\vm-direction.md"); EvidenceTerms = @("vm", "AVM", "bytecode", "gas") },
  [ordered]@{ Module = "x/aetravm"; Label = '`x/vm` / AVM'; Prefix = "VM"; Value = $true; EvidenceRoots = @("x\aetravm", "docs\architecture\avm.md", "docs\architecture\vm-direction.md"); EvidenceTerms = @("AVM", "async", "contract", "gas") },
  [ordered]@{ Module = "x/messaging"; Label = '`x/messaging`'; Prefix = "MSG"; Value = $true; EvidenceRoots = @("x\messaging", "x\mesh", "tests\adversarial", "docs\architecture\execution-os.md"); EvidenceTerms = @("messaging", "message", "receipt", "proof") },
  [ordered]@{ Module = "x/queue"; Label = '`x/queue`'; Prefix = "QUEUE"; Value = $true; EvidenceRoots = @("x\queue", "x\aetravm\async", "docs\architecture\async-smart-contract-execution.md"); EvidenceTerms = @("queue", "bounce", "refund", "delayed") },
  [ordered]@{ Module = "x/events"; Label = '`x/events`'; Prefix = "EVENTS"; Value = $false; EvidenceRoots = @("x\events", "docs\event-contract.md", "tests\scripts\event_contract_doc_test.ps1"); EvidenceTerms = @("events", "event", "receipt", "attributes") },
  [ordered]@{ Module = "x/actors"; Label = '`x/actors`'; Prefix = "ACTOR"; Value = $true; EvidenceRoots = @("x\actors", "docs\module-boundaries.md"); EvidenceTerms = @("actor", "mailbox", "logical time") },
  [ordered]@{ Module = "x/scheduler"; Label = '`x/scheduler`'; Prefix = "SCHED"; Value = $true; EvidenceRoots = @("x\scheduler", "docs\module-boundaries.md"); EvidenceTerms = @("scheduler", "schedule", "task", "priority") },
  [ordered]@{ Module = "x/storage"; Label = '`x/storage`'; Prefix = "STORE"; Value = $true; EvidenceRoots = @("x\storage", "docs\module-boundaries.md", "docs\architecture\avm.md"); EvidenceTerms = @("storage", "snapshot", "export", "state root") },
  [ordered]@{ Module = "x/memo"; Label = '`x/memo`'; Prefix = "MEMO"; Value = $true; EvidenceRoots = @("x\memo", "docs\mempool-checktx-negative-flow.md", "docs\transaction-lifecycle-matrix.md"); EvidenceTerms = @("memo", "UTF-8", "metadata") },
  [ordered]@{ Module = "x/indexer"; Label = '`x/indexer`'; Prefix = "INDEX"; Value = $false; EvidenceRoots = @("x\indexer", "app\indexer", "docs\event-contract.md", "docs\query-surface.md"); EvidenceTerms = @("index", "indexer", "query", "event") },
  [ordered]@{ Module = "x/sharding/sim"; Label = '`x/sharding/sim` and load/routing'; Prefix = "SHARD"; Value = $true; EvidenceRoots = @("x\sharding", "x\load", "x\routing", "tests\adversarial", "docs\architecture\sharding-rd.md"); EvidenceTerms = @("sharding", "LOAD_SCORE", "route", "shard") }
)

$requiredSourceTerms = @(
  "docs/security/aetra-fuzzing-invariant-pipeline.md",
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
$exploitCatalogSection = Get-AexsMarkdownSection -Text $taskText -Heading "Exploit Task Catalog"

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
  "Generate random contract-assets create, mint, burn, and admin sequences.",
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
  "No priority ordering divergence across nodes.",
  "Identity And Resolver Invariants",
  "Domain names are unique.",
  "Active domains cannot be re-auctioned.",
  "Expired domains require auction or explicit renewal path.",
  "Resolver cannot point to malformed or zero address.",
  "Resolver-based payment fails before funds move if unresolved.",
  "Reverse lookup is consistent with owner-approved mapping.",
  "Domain registry owner and NFT representation owner do not diverge.",
  "Subdomain ownership and resolver delegation do not bypass parent rules."
)

$requiredExecutionInvariantTerms = @(
  "Execution, AVM, And Queue Invariants",
  "AVM malformed input does not panic.",
  "AVM gas is bounded and deterministic.",
  "Infinite loops are stopped by gas or instruction limits.",
  "Contract state updates are deterministic.",
  "Queue ordering is deterministic.",
  "Cross-zone message replay is rejected.",
  "Bounce/refund cannot double-spend.",
  "Message loops are bounded by depth and per-block limits.",
  "Export/import preserves queue state exactly."
)

$requiredCoreExploitTerms = @(
  "Every exploit task must produce a seed, step list, expected state, actual state,",
  "affected module list, severity, and fix recommendation if it succeeds.",
  "Consensus And Aether Core Exploits",
  "Attempt double-sign fork creation.",
  "Attempt equivocation across heights and rounds.",
  "Attempt long-range history rewrite.",
  "Attempt stake grinding.",
  "Attempt validator cartel concentration scenario.",
  "Attempt stake delegation manipulation.",
  "Attempt self-delegation inflation.",
  "Attempt fake validator liveness.",
  "Attempt validator eclipse simulation.",
  "Attempt block withholding.",
  "Attempt fork choice manipulation.",
  "Attempt finality delay manipulation.",
  "Attempt Byzantine majority simulator scenario."
)

$requiredSlashingExploitTerms = @(
  "Slashing Bypass Exploits",
  "Attempt delayed evidence submission bypass.",
  "Attempt malformed equivocation proof acceptance.",
  "Attempt slashing race condition.",
  "Attempt redelegation-based partial slash evasion.",
  "Attempt unbonding window slash evasion.",
  "Attempt jail escape through upgrade timing.",
  "Attempt invalid evidence replay."
)

$requiredTransactionExploitTerms = @(
  "Transaction, Auth, And Bank Exploits",
  "Attempt signature replay.",
  "Attempt cross-context replay with wrong chain id.",
  "Attempt invalid nonce bypass.",
  "Attempt transaction malleability.",
  "Attempt fee underpayment bypass.",
  "Attempt fee inflation manipulation.",
  "Attempt low-fee spam griefing.",
  "Attempt multi-send partial failure exploit.",
  "Attempt race-condition double spend.",
  "Attempt rollback exploit during replayed state transition.",
  "Attempt zero-address transfer or signer path."
)

$requiredTokenEconomyExploitTerms = @(
  "Token And Economy Exploits",
  "Attempt contract-assets mint authority takeover.",
  "Attempt unauthorized burn bypass.",
  "Attempt inflation manipulation through governance timing.",
  "Attempt fee routing manipulation.",
  "Attempt treasury drain via governance proposal.",
  "Attempt staking reward inflation.",
  "Attempt staking reward farming loop.",
  "Attempt supply manipulation through edge-case mint path.",
  "Attempt native denom spoofing.",
  "Attempt display/base decimal mismatch exploit."
)

$requiredDexExploitTerms = @(
  "DEX Exploits",
  "Attempt constant-product invariant break.",
  "Attempt liquidity drain through swap sequence.",
  "Attempt pool initialization manipulation.",
  "Attempt LP token inflation.",
  "Attempt liquidity removal race.",
  "Attempt zero-liquidity swap edge case.",
  "Attempt reserve/module balance desync.",
  "Attempt failed bank movement partial update.",
  "Attempt slippage bypass.",
  "Attempt rounding exploit."
)

$requiredLoadSystemExploitTerms = @(
  "Load System Exploits",
  'Attempt `LOAD_SCORE` manipulation through spam bursts.',
  "Attempt artificial mempool inflation.",
  "Attempt block saturation.",
  "Attempt execution delay amplification.",
  "Attempt EMA slow-poison attack.",
  "Attempt load spike oscillation.",
  "Attempt shard overload targeting through load manipulation.",
  "Attempt priority fee gaming.",
  "Attempt adaptive fee destabilization."
)

$requiredRoutingEngineExploitTerms = @(
  "Routing Engine Exploits",
  "Attempt routing bias exploitation.",
  "Attempt zone congestion targeting.",
  "Attempt compute shard starvation.",
  "Attempt hot-zone monopolization.",
  "Attempt deterministic route prediction abuse.",
  "Attempt cross-zone routing loop.",
  "Attempt routing desync between nodes.",
  "Attempt transaction misclassification.",
  "Attempt fee-based routing gaming."
)

$requiredExecutionZoneAvmExploitTerms = @(
  "Execution Zone And AVM Exploits",
  "Attempt state divergence between zones.",
  "Attempt cross-zone replay.",
  "Attempt AVM determinism violation.",
  "Attempt contract execution desync.",
  "Attempt parallel execution race condition.",
  "Attempt state corruption.",
  "Attempt partial execution rollback.",
  "Attempt nondeterministic opcode or host behavior.",
  "Attempt gas exhaustion denial-of-service.",
  "Attempt infinite loop griefing.",
  "Attempt storage collision.",
  "Attempt contract upgrade hijack.",
  "Attempt uninitialized storage exploit.",
  "Attempt stack overflow.",
  "Attempt sandbox escape."
)

$requiredComputeShardExploitTerms = @(
  "Compute Shard Exploits",
  "Attempt shard partition imbalance.",
  "Attempt shard starvation.",
  "Attempt shard overflow collapse.",
  "Attempt cross-shard inconsistency.",
  "Attempt load spoofing for shard activation.",
  "Attempt shard duplication.",
  "Attempt state split inconsistency.",
  "Attempt parallel execution collision.",
  "Attempt scheduling manipulation.",
  "Attempt queue flooding."
)

$requiredMeshCrossZoneExploitTerms = @(
  "Aether Mesh And Cross-Zone Exploits",
  "Attempt cross-zone message replay.",
  "Attempt message delay manipulation.",
  "Attempt message ordering attack.",
  "Attempt asset duplication across zones.",
  "Attempt double spend across zones.",
  "Attempt proof forgery.",
  "Attempt relay censorship simulation.",
  "Attempt message starvation.",
  "Attempt finality mismatch.",
  "Attempt stale receipt replay."
)

$requiredIdentityDomainExploitTerms = @(
  'Identity And `.aet` Domain Exploits',
  "Attempt domain hijack through resolver overwrite.",
  "Attempt expired domain takeover without auction.",
  "Attempt auction manipulation.",
  "Attempt resolver spoofing.",
  "Attempt subdomain collision.",
  "Attempt reverse lookup poisoning.",
  "Attempt domain binding race condition.",
  "Attempt index-layer cache poisoning.",
  "Attempt fake domain resolution injection.",
  "Attempt multi-resolver inconsistency."
)

$requiredGovernanceExploitTerms = @(
  "Governance Exploits",
  "Attempt governance capture through voting power manipulation.",
  "Attempt proposal spam.",
  "Attempt emergency parameter abuse.",
  "Attempt upgrade hijack.",
  "Attempt delayed execution exploitation.",
  "Attempt governance replay.",
  "Attempt proposal front-running.",
  "Attempt staking-loop voting power manipulation.",
  "Attempt parameter griefing."
)

$requiredGenesisUpgradeStateExploitTerms = @(
  "Genesis, Upgrade, And State Exploits",
  "Attempt malformed genesis injection.",
  "Attempt state export tampering.",
  "Attempt upgrade rollback.",
  "Attempt partial migration corruption.",
  "Attempt module initialization bypass.",
  "Attempt hidden privileged account injection.",
  'Attempt `InitGenesis` validation bypass.',
  "Attempt version mismatch exploit.",
  "Attempt state root collision.",
  "Attempt snapshot poisoning."
)

$requiredMempoolNetworkExploitTerms = @(
  "Mempool And Network Exploits",
  "Attempt mempool flooding.",
  "Attempt transaction prioritization gaming.",
  "Attempt gossip poisoning.",
  "Attempt node eclipse simulation.",
  "Attempt P2P partition simulation.",
  "Attempt block propagation delay.",
  "Attempt transaction reordering.",
  "Attempt network latency exploitation.",
  "Attempt bandwidth exhaustion.",
  "Attempt peer targeting."
)

$requiredCombinedFullStackExploitTerms = @(
  "Combined Full-Stack Exploits",
  "Attempt coordinated spam plus routing attack.",
  "Attempt load plus governance combined attack.",
  "Attempt DEX plus mempool plus routing exploit.",
  "Attempt validator collusion plus slashing delay exploit.",
  "Attempt cross-zone value extraction coordination.",
  "Attempt identity plus routing hijack.",
  "Attempt shard overload plus fee manipulation cascade.",
  "Attempt consensus plus mempool denial-of-service hybrid.",
  "Attempt economic plus staking starvation.",
  "Attempt full-stack destabilization."
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
  foreach ($term in @(Get-AexsMissingTerms -Text $invariantChecklistSection -Terms $requiredExecutionInvariantTerms)) {
    $sourceFailures += "missing execution invariant checklist term: $term"
  }
}
if ([string]::IsNullOrWhiteSpace($exploitCatalogSection)) {
  $sourceFailures += "missing Exploit Task Catalog section"
} else {
  foreach ($term in @(Get-AexsMissingTerms -Text $exploitCatalogSection -Terms $requiredCoreExploitTerms)) {
    $sourceFailures += "missing core exploit catalog term: $term"
  }
  foreach ($term in @(Get-AexsMissingTerms -Text $exploitCatalogSection -Terms $requiredSlashingExploitTerms)) {
    $sourceFailures += "missing slashing exploit catalog term: $term"
  }
  foreach ($term in @(Get-AexsMissingTerms -Text $exploitCatalogSection -Terms $requiredTransactionExploitTerms)) {
    $sourceFailures += "missing transaction/auth/bank exploit catalog term: $term"
  }
  foreach ($term in @(Get-AexsMissingTerms -Text $exploitCatalogSection -Terms $requiredTokenEconomyExploitTerms)) {
    $sourceFailures += "missing token/economy exploit catalog term: $term"
  }
  foreach ($term in @(Get-AexsMissingTerms -Text $exploitCatalogSection -Terms $requiredDexExploitTerms)) {
    $sourceFailures += "missing DEX exploit catalog term: $term"
  }
  foreach ($term in @(Get-AexsMissingTerms -Text $exploitCatalogSection -Terms $requiredLoadSystemExploitTerms)) {
    $sourceFailures += "missing load system exploit catalog term: $term"
  }
  foreach ($term in @(Get-AexsMissingTerms -Text $exploitCatalogSection -Terms $requiredRoutingEngineExploitTerms)) {
    $sourceFailures += "missing routing engine exploit catalog term: $term"
  }
  foreach ($term in @(Get-AexsMissingTerms -Text $exploitCatalogSection -Terms $requiredExecutionZoneAvmExploitTerms)) {
    $sourceFailures += "missing execution zone/AVM exploit catalog term: $term"
  }
  foreach ($term in @(Get-AexsMissingTerms -Text $exploitCatalogSection -Terms $requiredComputeShardExploitTerms)) {
    $sourceFailures += "missing compute shard exploit catalog term: $term"
  }
  foreach ($term in @(Get-AexsMissingTerms -Text $exploitCatalogSection -Terms $requiredMeshCrossZoneExploitTerms)) {
    $sourceFailures += "missing mesh/cross-zone exploit catalog term: $term"
  }
  foreach ($term in @(Get-AexsMissingTerms -Text $exploitCatalogSection -Terms $requiredIdentityDomainExploitTerms)) {
    $sourceFailures += "missing identity/.aet exploit catalog term: $term"
  }
  foreach ($term in @(Get-AexsMissingTerms -Text $exploitCatalogSection -Terms $requiredGovernanceExploitTerms)) {
    $sourceFailures += "missing governance exploit catalog term: $term"
  }
  foreach ($term in @(Get-AexsMissingTerms -Text $exploitCatalogSection -Terms $requiredGenesisUpgradeStateExploitTerms)) {
    $sourceFailures += "missing genesis/upgrade/state exploit catalog term: $term"
  }
  foreach ($term in @(Get-AexsMissingTerms -Text $exploitCatalogSection -Terms $requiredMempoolNetworkExploitTerms)) {
    $sourceFailures += "missing mempool/network exploit catalog term: $term"
  }
  foreach ($term in @(Get-AexsMissingTerms -Text $exploitCatalogSection -Terms $requiredCombinedFullStackExploitTerms)) {
    $sourceFailures += "missing combined full-stack exploit catalog term: $term"
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
$coreExploitRecords = @(Get-AexsCoreExploitRecords -Text $taskText -CampaignId $campaignId)
foreach ($record in $coreExploitRecords) {
  $invalidReasons = @(Test-AexsExploitRecord -Record $record)
  $record["valid"] = $invalidReasons.Count -eq 0
  $record["invalid_reasons"] = $invalidReasons
}
$invalidCoreExploitRecords = @($coreExploitRecords | Where-Object { -not $_["valid"] })
$slashingExploitRecords = @(Get-AexsExploitRecordsForSection -Text $taskText -CampaignId $campaignId -SectionNumber 2 -SectionTitle "Slashing Bypass Exploits" -IdPrefix "SLASHEXP" -SeedNamespace "slashing-exploit" -Overrides (Get-AexsSlashingExploitOverrides) -DefaultAffectedModules @("x/slashing", "x/staking", "app evidence handling") -DefaultSeverity "High")
foreach ($record in $slashingExploitRecords) {
  $invalidReasons = @(Test-AexsExploitRecord -Record $record)
  $record["valid"] = $invalidReasons.Count -eq 0
  $record["invalid_reasons"] = $invalidReasons
}
$invalidSlashingExploitRecords = @($slashingExploitRecords | Where-Object { -not $_["valid"] })
$txAuthBankExploitRecords = @(Get-AexsExploitRecordsForSection -Text $taskText -CampaignId $campaignId -SectionNumber 3 -SectionTitle "Transaction, Auth, And Bank Exploits" -IdPrefix "TXEXP" -SeedNamespace "tx-auth-bank-exploit" -Overrides (Get-AexsTxAuthBankExploitOverrides) -DefaultAffectedModules @("x/auth", "x/bank", "x/fees", "app ante") -DefaultSeverity "High")
foreach ($record in $txAuthBankExploitRecords) {
  $invalidReasons = @(Test-AexsExploitRecord -Record $record)
  $record["valid"] = $invalidReasons.Count -eq 0
  $record["invalid_reasons"] = $invalidReasons
}
$invalidTxAuthBankExploitRecords = @($txAuthBankExploitRecords | Where-Object { -not $_["valid"] })
$tokenEconomyExploitRecords = @(Get-AexsExploitRecordsForSection -Text $taskText -CampaignId $campaignId -SectionNumber 4 -SectionTitle "Token And Economy Exploits" -IdPrefix "TOKENEXP" -SeedNamespace "token-economy-exploit" -Overrides (Get-AexsTokenEconomyExploitOverrides) -DefaultAffectedModules @("x/fees", "x/bank", "x/gov", "x/staking", "x/distribution") -DefaultSeverity "High")
foreach ($record in $tokenEconomyExploitRecords) {
  $invalidReasons = @(Test-AexsExploitRecord -Record $record)
  $record["valid"] = $invalidReasons.Count -eq 0
  $record["invalid_reasons"] = $invalidReasons
}
$invalidTokenEconomyExploitRecords = @($tokenEconomyExploitRecords | Where-Object { -not $_["valid"] })
$dexExploitRecords = @(Get-AexsExploitRecordsForSection -Text $taskText -CampaignId $campaignId -SectionNumber 5 -SectionTitle "DEX Exploits" -IdPrefix "DEXEXP" -SeedNamespace "dex-exploit" -Overrides (Get-AexsDexExploitOverrides) -DefaultAffectedModules @("avm-dex-contract", "x/bank", "app cache context") -DefaultSeverity "High")
foreach ($record in $dexExploitRecords) {
  $invalidReasons = @(Test-AexsExploitRecord -Record $record)
  $record["valid"] = $invalidReasons.Count -eq 0
  $record["invalid_reasons"] = $invalidReasons
}
$invalidDexExploitRecords = @($dexExploitRecords | Where-Object { -not $_["valid"] })
$loadSystemExploitRecords = @(Get-AexsExploitRecordsForSection -Text $taskText -CampaignId $campaignId -SectionNumber 6 -SectionTitle "Load System Exploits" -IdPrefix "LOADEXP" -SeedNamespace "load-system-exploit" -Overrides (Get-AexsLoadSystemExploitOverrides) -DefaultAffectedModules @("x/load", "x/routing", "x/sharding/sim") -DefaultSeverity "High")
foreach ($record in $loadSystemExploitRecords) {
  $invalidReasons = @(Test-AexsExploitRecord -Record $record)
  $record["valid"] = $invalidReasons.Count -eq 0
  $record["invalid_reasons"] = $invalidReasons
}
$invalidLoadSystemExploitRecords = @($loadSystemExploitRecords | Where-Object { -not $_["valid"] })
$routingEngineExploitRecords = @(Get-AexsExploitRecordsForSection -Text $taskText -CampaignId $campaignId -SectionNumber 7 -SectionTitle "Routing Engine Exploits" -IdPrefix "ROUTEEXP" -SeedNamespace "routing-engine-exploit" -Overrides (Get-AexsRoutingEngineExploitOverrides) -DefaultAffectedModules @("x/routing", "x/load", "x/sharding/sim") -DefaultSeverity "High")
foreach ($record in $routingEngineExploitRecords) {
  $invalidReasons = @(Test-AexsExploitRecord -Record $record)
  $record["valid"] = $invalidReasons.Count -eq 0
  $record["invalid_reasons"] = $invalidReasons
}
$invalidRoutingEngineExploitRecords = @($routingEngineExploitRecords | Where-Object { -not $_["valid"] })
$executionZoneAvmExploitRecords = @(Get-AexsExploitRecordsForSection -Text $taskText -CampaignId $campaignId -SectionNumber 8 -SectionTitle "Execution Zone And AVM Exploits" -IdPrefix "EXECZONEEXP" -SeedNamespace "execution-zone-avm-exploit" -Overrides (Get-AexsExecutionZoneAvmExploitOverrides) -DefaultAffectedModules @("x/execution", "x/aetravm", "x/vm", "x/queue", "x/storage") -DefaultSeverity "High")
foreach ($record in $executionZoneAvmExploitRecords) {
  $invalidReasons = @(Test-AexsExploitRecord -Record $record)
  $record["valid"] = $invalidReasons.Count -eq 0
  $record["invalid_reasons"] = $invalidReasons
}
$invalidExecutionZoneAvmExploitRecords = @($executionZoneAvmExploitRecords | Where-Object { -not $_["valid"] })
$computeShardExploitRecords = @(Get-AexsExploitRecordsForSection -Text $taskText -CampaignId $campaignId -SectionNumber 9 -SectionTitle "Compute Shard Exploits" -IdPrefix "SHARDEXP" -SeedNamespace "compute-shard-exploit" -Overrides (Get-AexsComputeShardExploitOverrides) -DefaultAffectedModules @("x/sharding/sim", "x/routing", "x/load", "x/scheduler", "x/queue") -DefaultSeverity "High")
foreach ($record in $computeShardExploitRecords) {
  $invalidReasons = @(Test-AexsExploitRecord -Record $record)
  $record["valid"] = $invalidReasons.Count -eq 0
  $record["invalid_reasons"] = $invalidReasons
}
$invalidComputeShardExploitRecords = @($computeShardExploitRecords | Where-Object { -not $_["valid"] })
$meshCrossZoneExploitRecords = @(Get-AexsExploitRecordsForSection -Text $taskText -CampaignId $campaignId -SectionNumber 10 -SectionTitle "Aether Mesh And Cross-Zone Exploits" -IdPrefix "MESHEXP" -SeedNamespace "mesh-cross-zone-exploit" -Overrides (Get-AexsMeshCrossZoneExploitOverrides) -DefaultAffectedModules @("x/mesh", "x/messaging", "x/queue", "x/sharding/sim") -DefaultSeverity "High")
foreach ($record in $meshCrossZoneExploitRecords) {
  $invalidReasons = @(Test-AexsExploitRecord -Record $record)
  $record["valid"] = $invalidReasons.Count -eq 0
  $record["invalid_reasons"] = $invalidReasons
}
$invalidMeshCrossZoneExploitRecords = @($meshCrossZoneExploitRecords | Where-Object { -not $_["valid"] })
$identityDomainExploitRecords = @(Get-AexsExploitRecordsForSection -Text $taskText -CampaignId $campaignId -SectionNumber 11 -SectionTitle 'Identity And `.aet` Domain Exploits' -IdPrefix "IDENTEXP" -SeedNamespace "identity-domain-exploit" -Overrides (Get-AexsIdentityDomainExploitOverrides) -DefaultAffectedModules @("x/identity", "x/indexer", "x/routing") -DefaultSeverity "High")
foreach ($record in $identityDomainExploitRecords) {
  $invalidReasons = @(Test-AexsExploitRecord -Record $record)
  $record["valid"] = $invalidReasons.Count -eq 0
  $record["invalid_reasons"] = $invalidReasons
}
$invalidIdentityDomainExploitRecords = @($identityDomainExploitRecords | Where-Object { -not $_["valid"] })
$governanceExploitRecords = @(Get-AexsExploitRecordsForSection -Text $taskText -CampaignId $campaignId -SectionNumber 12 -SectionTitle "Governance Exploits" -IdPrefix "GOVEXP" -SeedNamespace "governance-exploit" -Overrides (Get-AexsGovernanceExploitOverrides) -DefaultAffectedModules @("x/gov", "x/staking", "x/fees", "app") -DefaultSeverity "High")
foreach ($record in $governanceExploitRecords) {
  $invalidReasons = @(Test-AexsExploitRecord -Record $record)
  $record["valid"] = $invalidReasons.Count -eq 0
  $record["invalid_reasons"] = $invalidReasons
}
$invalidGovernanceExploitRecords = @($governanceExploitRecords | Where-Object { -not $_["valid"] })
$genesisUpgradeStateExploitRecords = @(Get-AexsExploitRecordsForSection -Text $taskText -CampaignId $campaignId -SectionNumber 13 -SectionTitle "Genesis, Upgrade, And State Exploits" -IdPrefix "STATEEXP" -SeedNamespace "genesis-upgrade-state-exploit" -Overrides (Get-AexsGenesisUpgradeStateExploitOverrides) -DefaultAffectedModules @("app", "x/upgrade", "x/storage") -DefaultSeverity "High")
foreach ($record in $genesisUpgradeStateExploitRecords) {
  $invalidReasons = @(Test-AexsExploitRecord -Record $record)
  $record["valid"] = $invalidReasons.Count -eq 0
  $record["invalid_reasons"] = $invalidReasons
}
$invalidGenesisUpgradeStateExploitRecords = @($genesisUpgradeStateExploitRecords | Where-Object { -not $_["valid"] })
$mempoolNetworkExploitRecords = @(Get-AexsExploitRecordsForSection -Text $taskText -CampaignId $campaignId -SectionNumber 14 -SectionTitle "Mempool And Network Exploits" -IdPrefix "NETEXP" -SeedNamespace "mempool-network-exploit" -Overrides (Get-AexsMempoolNetworkExploitOverrides) -DefaultAffectedModules @("mempool", "CometBFT P2P", "app") -DefaultSeverity "High")
foreach ($record in $mempoolNetworkExploitRecords) {
  $invalidReasons = @(Test-AexsExploitRecord -Record $record)
  $record["valid"] = $invalidReasons.Count -eq 0
  $record["invalid_reasons"] = $invalidReasons
}
$invalidMempoolNetworkExploitRecords = @($mempoolNetworkExploitRecords | Where-Object { -not $_["valid"] })
$combinedFullStackExploitRecords = @(Get-AexsExploitRecordsForSection -Text $taskText -CampaignId $campaignId -SectionNumber 15 -SectionTitle "Combined Full-Stack Exploits" -IdPrefix "FULLSTACKEXP" -SeedNamespace "combined-full-stack-exploit" -Overrides (Get-AexsCombinedFullStackExploitOverrides) -DefaultAffectedModules @("app", "x/gov", "x/load", "x/routing", "x/mesh") -DefaultSeverity "Critical")
foreach ($record in $combinedFullStackExploitRecords) {
  $invalidReasons = @(Test-AexsExploitRecord -Record $record)
  $record["valid"] = $invalidReasons.Count -eq 0
  $record["invalid_reasons"] = $invalidReasons
}
$invalidCombinedFullStackExploitRecords = @($combinedFullStackExploitRecords | Where-Object { -not $_["valid"] })
$exploitRecords = @($coreExploitRecords + $slashingExploitRecords + $txAuthBankExploitRecords + $tokenEconomyExploitRecords + $dexExploitRecords + $loadSystemExploitRecords + $routingEngineExploitRecords + $executionZoneAvmExploitRecords + $computeShardExploitRecords + $meshCrossZoneExploitRecords + $identityDomainExploitRecords + $governanceExploitRecords + $genesisUpgradeStateExploitRecords + $mempoolNetworkExploitRecords + $combinedFullStackExploitRecords)
$invalidExploitRecords = @($exploitRecords | Where-Object { -not $_["valid"] })
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
  "aexs-contract-assets-admin-0003",
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
    id                 = "contract-assets_admin_sequences"
    name               = "random contract-assets create, mint, burn, and admin sequences"
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
    target_modules    = @("x/auth", "x/bank", "avm-dex-contract", "x/identity")
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
    target_modules    = @("x/auth", "x/bank", "avm-dex-contract", "x/identity", "x/vm")
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
    target_modules    = @("x/queue", "x/messaging", "x/aetravm")
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
    target_modules    = @("x/vm", "x/aetravm")
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
    target_modules    = @("x/vm", "x/aetravm")
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
    target_modules    = @("app", "x/auth", "x/bank", "x/staking", "x/fees", "avm-dex-contract")
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
  exploit_count                       = $exploitRecords.Count
  invalid_exploit_count               = $invalidExploitRecords.Count
  exploit_ids                         = @($exploitRecords | ForEach-Object { $_["exploit_id"] })
  invalid_exploit_records             = @($invalidExploitRecords | ForEach-Object { $_["exploit_id"] })
  core_exploit_count                  = $coreExploitRecords.Count
  invalid_core_exploit_count          = $invalidCoreExploitRecords.Count
  core_exploit_ids                    = @($coreExploitRecords | ForEach-Object { $_["exploit_id"] })
  invalid_core_exploit_records        = @($invalidCoreExploitRecords | ForEach-Object { $_["exploit_id"] })
  slashing_exploit_count              = $slashingExploitRecords.Count
  invalid_slashing_exploit_count      = $invalidSlashingExploitRecords.Count
  slashing_exploit_ids                = @($slashingExploitRecords | ForEach-Object { $_["exploit_id"] })
  invalid_slashing_exploit_records    = @($invalidSlashingExploitRecords | ForEach-Object { $_["exploit_id"] })
  tx_auth_bank_exploit_count          = $txAuthBankExploitRecords.Count
  invalid_tx_auth_bank_exploit_count  = $invalidTxAuthBankExploitRecords.Count
  tx_auth_bank_exploit_ids            = @($txAuthBankExploitRecords | ForEach-Object { $_["exploit_id"] })
  invalid_tx_auth_bank_exploit_records = @($invalidTxAuthBankExploitRecords | ForEach-Object { $_["exploit_id"] })
  token_economy_exploit_count         = $tokenEconomyExploitRecords.Count
  invalid_token_economy_exploit_count = $invalidTokenEconomyExploitRecords.Count
  token_economy_exploit_ids           = @($tokenEconomyExploitRecords | ForEach-Object { $_["exploit_id"] })
  invalid_token_economy_exploit_records = @($invalidTokenEconomyExploitRecords | ForEach-Object { $_["exploit_id"] })
  dex_exploit_count                   = $dexExploitRecords.Count
  invalid_dex_exploit_count           = $invalidDexExploitRecords.Count
  dex_exploit_ids                     = @($dexExploitRecords | ForEach-Object { $_["exploit_id"] })
  invalid_dex_exploit_records         = @($invalidDexExploitRecords | ForEach-Object { $_["exploit_id"] })
  load_system_exploit_count           = $loadSystemExploitRecords.Count
  invalid_load_system_exploit_count   = $invalidLoadSystemExploitRecords.Count
  load_system_exploit_ids             = @($loadSystemExploitRecords | ForEach-Object { $_["exploit_id"] })
  invalid_load_system_exploit_records = @($invalidLoadSystemExploitRecords | ForEach-Object { $_["exploit_id"] })
  routing_engine_exploit_count        = $routingEngineExploitRecords.Count
  invalid_routing_engine_exploit_count = $invalidRoutingEngineExploitRecords.Count
  routing_engine_exploit_ids          = @($routingEngineExploitRecords | ForEach-Object { $_["exploit_id"] })
  invalid_routing_engine_exploit_records = @($invalidRoutingEngineExploitRecords | ForEach-Object { $_["exploit_id"] })
  execution_zone_avm_exploit_count    = $executionZoneAvmExploitRecords.Count
  invalid_execution_zone_avm_exploit_count = $invalidExecutionZoneAvmExploitRecords.Count
  execution_zone_avm_exploit_ids      = @($executionZoneAvmExploitRecords | ForEach-Object { $_["exploit_id"] })
  invalid_execution_zone_avm_exploit_records = @($invalidExecutionZoneAvmExploitRecords | ForEach-Object { $_["exploit_id"] })
  compute_shard_exploit_count         = $computeShardExploitRecords.Count
  invalid_compute_shard_exploit_count = $invalidComputeShardExploitRecords.Count
  compute_shard_exploit_ids           = @($computeShardExploitRecords | ForEach-Object { $_["exploit_id"] })
  invalid_compute_shard_exploit_records = @($invalidComputeShardExploitRecords | ForEach-Object { $_["exploit_id"] })
  mesh_cross_zone_exploit_count       = $meshCrossZoneExploitRecords.Count
  invalid_mesh_cross_zone_exploit_count = $invalidMeshCrossZoneExploitRecords.Count
  mesh_cross_zone_exploit_ids         = @($meshCrossZoneExploitRecords | ForEach-Object { $_["exploit_id"] })
  invalid_mesh_cross_zone_exploit_records = @($invalidMeshCrossZoneExploitRecords | ForEach-Object { $_["exploit_id"] })
  identity_domain_exploit_count       = $identityDomainExploitRecords.Count
  invalid_identity_domain_exploit_count = $invalidIdentityDomainExploitRecords.Count
  identity_domain_exploit_ids         = @($identityDomainExploitRecords | ForEach-Object { $_["exploit_id"] })
  invalid_identity_domain_exploit_records = @($invalidIdentityDomainExploitRecords | ForEach-Object { $_["exploit_id"] })
  governance_exploit_count            = $governanceExploitRecords.Count
  invalid_governance_exploit_count    = $invalidGovernanceExploitRecords.Count
  governance_exploit_ids              = @($governanceExploitRecords | ForEach-Object { $_["exploit_id"] })
  invalid_governance_exploit_records  = @($invalidGovernanceExploitRecords | ForEach-Object { $_["exploit_id"] })
  genesis_upgrade_state_exploit_count = $genesisUpgradeStateExploitRecords.Count
  invalid_genesis_upgrade_state_exploit_count = $invalidGenesisUpgradeStateExploitRecords.Count
  genesis_upgrade_state_exploit_ids   = @($genesisUpgradeStateExploitRecords | ForEach-Object { $_["exploit_id"] })
  invalid_genesis_upgrade_state_exploit_records = @($invalidGenesisUpgradeStateExploitRecords | ForEach-Object { $_["exploit_id"] })
  mempool_network_exploit_count       = $mempoolNetworkExploitRecords.Count
  invalid_mempool_network_exploit_count = $invalidMempoolNetworkExploitRecords.Count
  mempool_network_exploit_ids         = @($mempoolNetworkExploitRecords | ForEach-Object { $_["exploit_id"] })
  invalid_mempool_network_exploit_records = @($invalidMempoolNetworkExploitRecords | ForEach-Object { $_["exploit_id"] })
  combined_full_stack_exploit_count   = $combinedFullStackExploitRecords.Count
  invalid_combined_full_stack_exploit_count = $invalidCombinedFullStackExploitRecords.Count
  combined_full_stack_exploit_ids     = @($combinedFullStackExploitRecords | ForEach-Object { $_["exploit_id"] })
  invalid_combined_full_stack_exploit_records = @($invalidCombinedFullStackExploitRecords | ForEach-Object { $_["exploit_id"] })
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
$coreExploitPath = Join-Path $campaignDir "exploit-catalog.json"
$coreExploitMarkdownPath = Join-Path $campaignDir "exploit-catalog.md"
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
$exploitRecords | ConvertTo-Json -Depth 10 | Set-Content -LiteralPath $coreExploitPath
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

$coreExploitReport = @()
$coreExploitReport += "# AEXS Exploit Catalog"
$coreExploitReport += ""
$coreExploitReport += "- campaign id: $campaignId"
$coreExploitReport += "- exploit count: $($exploitRecords.Count)"
$coreExploitReport += "- invalid exploit count: $($invalidExploitRecords.Count)"
$coreExploitReport += "- status: planned_not_executed"
$coreExploitReport += ""
$coreExploitReport += "| Exploit | Severity | Path | Expected state | Actual state | Affected modules | Seed |"
$coreExploitReport += "| --- | --- | --- | --- | --- | --- | --- |"
foreach ($record in $exploitRecords) {
  $path = ([string]$record["exploit_path"]).Replace("|", "/")
  $expected = ([string]$record["expected_state"]).Replace("|", "/")
  $affected = (@($record["affected_modules"]) -join ", ").Replace("|", "/")
  $coreExploitReport += "| $($record["exploit_id"]) | $($record["severity"]) | $path | $expected | $($record["actual_state"]) | $affected | $($record["seed"]) |"
}
$coreExploitReport | Set-Content -LiteralPath $coreExploitMarkdownPath

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
$report += "- exploit records: $($exploitRecords.Count)"
$report += "- invalid exploit records: $($invalidExploitRecords.Count)"
$report += "- consensus/aether core exploit records: $($coreExploitRecords.Count)"
$report += "- invalid consensus/aether core exploit records: $($invalidCoreExploitRecords.Count)"
$report += "- slashing bypass exploit records: $($slashingExploitRecords.Count)"
$report += "- invalid slashing bypass exploit records: $($invalidSlashingExploitRecords.Count)"
$report += "- transaction/auth/bank exploit records: $($txAuthBankExploitRecords.Count)"
$report += "- invalid transaction/auth/bank exploit records: $($invalidTxAuthBankExploitRecords.Count)"
$report += "- token/economy exploit records: $($tokenEconomyExploitRecords.Count)"
$report += "- invalid token/economy exploit records: $($invalidTokenEconomyExploitRecords.Count)"
$report += "- DEX exploit records: $($dexExploitRecords.Count)"
$report += "- invalid DEX exploit records: $($invalidDexExploitRecords.Count)"
$report += "- load system exploit records: $($loadSystemExploitRecords.Count)"
$report += "- invalid load system exploit records: $($invalidLoadSystemExploitRecords.Count)"
$report += "- routing engine exploit records: $($routingEngineExploitRecords.Count)"
$report += "- invalid routing engine exploit records: $($invalidRoutingEngineExploitRecords.Count)"
$report += "- execution zone/AVM exploit records: $($executionZoneAvmExploitRecords.Count)"
$report += "- invalid execution zone/AVM exploit records: $($invalidExecutionZoneAvmExploitRecords.Count)"
$report += "- compute shard exploit records: $($computeShardExploitRecords.Count)"
$report += "- invalid compute shard exploit records: $($invalidComputeShardExploitRecords.Count)"
$report += "- mesh/cross-zone exploit records: $($meshCrossZoneExploitRecords.Count)"
$report += "- invalid mesh/cross-zone exploit records: $($invalidMeshCrossZoneExploitRecords.Count)"
$report += "- identity/.aet exploit records: $($identityDomainExploitRecords.Count)"
$report += "- invalid identity/.aet exploit records: $($invalidIdentityDomainExploitRecords.Count)"
$report += "- governance exploit records: $($governanceExploitRecords.Count)"
$report += "- invalid governance exploit records: $($invalidGovernanceExploitRecords.Count)"
$report += "- genesis/upgrade/state exploit records: $($genesisUpgradeStateExploitRecords.Count)"
$report += "- invalid genesis/upgrade/state exploit records: $($invalidGenesisUpgradeStateExploitRecords.Count)"
$report += "- mempool/network exploit records: $($mempoolNetworkExploitRecords.Count)"
$report += "- invalid mempool/network exploit records: $($invalidMempoolNetworkExploitRecords.Count)"
$report += "- combined full-stack exploit records: $($combinedFullStackExploitRecords.Count)"
$report += "- invalid combined full-stack exploit records: $($invalidCombinedFullStackExploitRecords.Count)"
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
$report += "- invalid exploit records: $(@($invalidExploitRecords | ForEach-Object { $_["exploit_id"] }) -join ', ')"
$report += "- invalid consensus/aether core exploit records: $(@($invalidCoreExploitRecords | ForEach-Object { $_["exploit_id"] }) -join ', ')"
$report += "- invalid slashing bypass exploit records: $(@($invalidSlashingExploitRecords | ForEach-Object { $_["exploit_id"] }) -join ', ')"
$report += "- invalid transaction/auth/bank exploit records: $(@($invalidTxAuthBankExploitRecords | ForEach-Object { $_["exploit_id"] }) -join ', ')"
$report += "- invalid token/economy exploit records: $(@($invalidTokenEconomyExploitRecords | ForEach-Object { $_["exploit_id"] }) -join ', ')"
$report += "- invalid DEX exploit records: $(@($invalidDexExploitRecords | ForEach-Object { $_["exploit_id"] }) -join ', ')"
$report += "- invalid load system exploit records: $(@($invalidLoadSystemExploitRecords | ForEach-Object { $_["exploit_id"] }) -join ', ')"
$report += "- invalid routing engine exploit records: $(@($invalidRoutingEngineExploitRecords | ForEach-Object { $_["exploit_id"] }) -join ', ')"
$report += "- invalid execution zone/AVM exploit records: $(@($invalidExecutionZoneAvmExploitRecords | ForEach-Object { $_["exploit_id"] }) -join ', ')"
$report += "- invalid compute shard exploit records: $(@($invalidComputeShardExploitRecords | ForEach-Object { $_["exploit_id"] }) -join ', ')"
$report += "- invalid mesh/cross-zone exploit records: $(@($invalidMeshCrossZoneExploitRecords | ForEach-Object { $_["exploit_id"] }) -join ', ')"
$report += "- invalid identity/.aet exploit records: $(@($invalidIdentityDomainExploitRecords | ForEach-Object { $_["exploit_id"] }) -join ', ')"
$report += "- invalid governance exploit records: $(@($invalidGovernanceExploitRecords | ForEach-Object { $_["exploit_id"] }) -join ', ')"
$report += "- invalid genesis/upgrade/state exploit records: $(@($invalidGenesisUpgradeStateExploitRecords | ForEach-Object { $_["exploit_id"] }) -join ', ')"
$report += "- invalid mempool/network exploit records: $(@($invalidMempoolNetworkExploitRecords | ForEach-Object { $_["exploit_id"] }) -join ', ')"
$report += "- invalid combined full-stack exploit records: $(@($invalidCombinedFullStackExploitRecords | ForEach-Object { $_["exploit_id"] }) -join ', ')"
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
$report += "## Consensus And Aether Core Exploit Catalog"
$report += ""
$report += "| Exploit | Severity | Actual state | Status |"
$report += "| --- | --- | --- | --- |"
foreach ($record in $exploitRecords) {
  $report += "| $($record["exploit_id"]) | $($record["severity"]) | $($record["actual_state"]) | $($record["status"]) |"
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
if ($coreExploitRecords.Count -lt 13) {
  throw "AEXS core exploit catalog validation failed: fewer than required consensus and Aether Core exploit records"
}
if ($invalidCoreExploitRecords.Count -gt 0) {
  throw "AEXS core exploit catalog validation failed for record(s): $(@($invalidCoreExploitRecords | ForEach-Object { $_["exploit_id"] }) -join ', ')"
}
if ($slashingExploitRecords.Count -lt 7) {
  throw "AEXS slashing exploit catalog validation failed: fewer than required slashing bypass exploit records"
}
if ($invalidSlashingExploitRecords.Count -gt 0) {
  throw "AEXS slashing exploit catalog validation failed for record(s): $(@($invalidSlashingExploitRecords | ForEach-Object { $_["exploit_id"] }) -join ', ')"
}
if ($txAuthBankExploitRecords.Count -lt 11) {
  throw "AEXS transaction/auth/bank exploit catalog validation failed: fewer than required transaction/auth/bank exploit records"
}
if ($invalidTxAuthBankExploitRecords.Count -gt 0) {
  throw "AEXS transaction/auth/bank exploit catalog validation failed for record(s): $(@($invalidTxAuthBankExploitRecords | ForEach-Object { $_["exploit_id"] }) -join ', ')"
}
if ($tokenEconomyExploitRecords.Count -lt 10) {
  throw "AEXS token/economy exploit catalog validation failed: fewer than required token/economy exploit records"
}
if ($invalidTokenEconomyExploitRecords.Count -gt 0) {
  throw "AEXS token/economy exploit catalog validation failed for record(s): $(@($invalidTokenEconomyExploitRecords | ForEach-Object { $_["exploit_id"] }) -join ', ')"
}
if ($dexExploitRecords.Count -lt 10) {
  throw "AEXS DEX exploit catalog validation failed: fewer than required DEX exploit records"
}
if ($invalidDexExploitRecords.Count -gt 0) {
  throw "AEXS DEX exploit catalog validation failed for record(s): $(@($invalidDexExploitRecords | ForEach-Object { $_["exploit_id"] }) -join ', ')"
}
if ($loadSystemExploitRecords.Count -lt 9) {
  throw "AEXS load system exploit catalog validation failed: fewer than required load system exploit records"
}
if ($invalidLoadSystemExploitRecords.Count -gt 0) {
  throw "AEXS load system exploit catalog validation failed for record(s): $(@($invalidLoadSystemExploitRecords | ForEach-Object { $_["exploit_id"] }) -join ', ')"
}
if ($routingEngineExploitRecords.Count -lt 9) {
  throw "AEXS routing engine exploit catalog validation failed: fewer than required routing engine exploit records"
}
if ($invalidRoutingEngineExploitRecords.Count -gt 0) {
  throw "AEXS routing engine exploit catalog validation failed for record(s): $(@($invalidRoutingEngineExploitRecords | ForEach-Object { $_["exploit_id"] }) -join ', ')"
}
if ($executionZoneAvmExploitRecords.Count -lt 15) {
  throw "AEXS execution zone/AVM exploit catalog validation failed: fewer than required execution zone/AVM exploit records"
}
if ($invalidExecutionZoneAvmExploitRecords.Count -gt 0) {
  throw "AEXS execution zone/AVM exploit catalog validation failed for record(s): $(@($invalidExecutionZoneAvmExploitRecords | ForEach-Object { $_["exploit_id"] }) -join ', ')"
}
if ($computeShardExploitRecords.Count -lt 10) {
  throw "AEXS compute shard exploit catalog validation failed: fewer than required compute shard exploit records"
}
if ($invalidComputeShardExploitRecords.Count -gt 0) {
  throw "AEXS compute shard exploit catalog validation failed for record(s): $(@($invalidComputeShardExploitRecords | ForEach-Object { $_["exploit_id"] }) -join ', ')"
}
if ($meshCrossZoneExploitRecords.Count -lt 10) {
  throw "AEXS mesh/cross-zone exploit catalog validation failed: fewer than required mesh/cross-zone exploit records"
}
if ($invalidMeshCrossZoneExploitRecords.Count -gt 0) {
  throw "AEXS mesh/cross-zone exploit catalog validation failed for record(s): $(@($invalidMeshCrossZoneExploitRecords | ForEach-Object { $_["exploit_id"] }) -join ', ')"
}
if ($identityDomainExploitRecords.Count -lt 10) {
  throw "AEXS identity/.aet exploit catalog validation failed: fewer than required identity/.aet exploit records"
}
if ($invalidIdentityDomainExploitRecords.Count -gt 0) {
  throw "AEXS identity/.aet exploit catalog validation failed for record(s): $(@($invalidIdentityDomainExploitRecords | ForEach-Object { $_["exploit_id"] }) -join ', ')"
}
if ($governanceExploitRecords.Count -lt 9) {
  throw "AEXS governance exploit catalog validation failed: fewer than required governance exploit records"
}
if ($invalidGovernanceExploitRecords.Count -gt 0) {
  throw "AEXS governance exploit catalog validation failed for record(s): $(@($invalidGovernanceExploitRecords | ForEach-Object { $_["exploit_id"] }) -join ', ')"
}
if ($genesisUpgradeStateExploitRecords.Count -lt 10) {
  throw "AEXS genesis/upgrade/state exploit catalog validation failed: fewer than required genesis/upgrade/state exploit records"
}
if ($invalidGenesisUpgradeStateExploitRecords.Count -gt 0) {
  throw "AEXS genesis/upgrade/state exploit catalog validation failed for record(s): $(@($invalidGenesisUpgradeStateExploitRecords | ForEach-Object { $_["exploit_id"] }) -join ', ')"
}
if ($mempoolNetworkExploitRecords.Count -lt 10) {
  throw "AEXS mempool/network exploit catalog validation failed: fewer than required mempool/network exploit records"
}
if ($invalidMempoolNetworkExploitRecords.Count -gt 0) {
  throw "AEXS mempool/network exploit catalog validation failed for record(s): $(@($invalidMempoolNetworkExploitRecords | ForEach-Object { $_["exploit_id"] }) -join ', ')"
}
if ($combinedFullStackExploitRecords.Count -lt 10) {
  throw "AEXS combined full-stack exploit catalog validation failed: fewer than required combined full-stack exploit records"
}
if ($invalidCombinedFullStackExploitRecords.Count -gt 0) {
  throw "AEXS combined full-stack exploit catalog validation failed for record(s): $(@($invalidCombinedFullStackExploitRecords | ForEach-Object { $_["exploit_id"] }) -join ', ')"
}
if ($invalidExploitRecords.Count -gt 0) {
  throw "AEXS exploit catalog validation failed for record(s): $(@($invalidExploitRecords | ForEach-Object { $_["exploit_id"] }) -join ', ')"
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
