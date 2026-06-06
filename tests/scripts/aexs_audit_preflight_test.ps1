param(
  [string]$OutputDir = ".work\aexs-test"
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version 2.0

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))

function Resolve-RepoPath {
  param([string]$Path)
  if ([System.IO.Path]::IsPathRooted($Path)) {
    return [System.IO.Path]::GetFullPath($Path)
  }
  return [System.IO.Path]::GetFullPath((Join-Path $RepoRoot $Path))
}

function Assert-True {
  param([bool]$Condition, [string]$Message)
  if (-not $Condition) {
    throw $Message
  }
}

function Assert-Contains {
  param([object[]]$Values, [string]$Expected, [string]$Message)
  if ($Expected -notin $Values) {
    throw $Message
  }
}

$resolvedOutput = Resolve-RepoPath $OutputDir
$repoPrefix = $RepoRoot.TrimEnd('\', '/') + [System.IO.Path]::DirectorySeparatorChar
Assert-True ($resolvedOutput.StartsWith($repoPrefix, [System.StringComparison]::OrdinalIgnoreCase)) "AEXS test output must stay under repository"

if (Test-Path -LiteralPath $resolvedOutput) {
  Remove-Item -LiteralPath $resolvedOutput -Recurse -Force
}

Push-Location $RepoRoot
try {
  $jsonText = & .\scripts\security\aexs-audit.ps1 -OutputDir $OutputDir -Json
  $result = $jsonText | ConvertFrom-Json

  Assert-True ($result.campaign_id -match '^aexs-[0-9a-f]{12}-[0-9a-f]{16}$') "campaign id must be deterministic and commit-based"
  Assert-True ($result.output_dir.StartsWith($resolvedOutput, [System.StringComparison]::OrdinalIgnoreCase)) "runtime report must be under requested .work output"
  Assert-True ($result.source_task_file -eq "TO_AUDIT.md") "TO_AUDIT must be the task source"
  Assert-True ($result.source_pipeline_doc -eq "docs\security\aetheris-fuzzing-invariant-pipeline.md") "pipeline doc must be the primary source"
  Assert-True ($result.planned_coverage_percent -ge 95) "planned coverage must meet 95 percent threshold"
  Assert-True ($result.audit_passed -eq $false) "pre-campaign audit must not be marked passed"
  Assert-True ($result.production_safe -eq $false) "pre-campaign audit must not be production safe"
  Assert-True ($result.mandatory_invariant_pass_rate -eq 0) "pre-campaign invariant pass rate must be zero until execution evidence exists"
  Assert-True (@($result.modules_below_planned_threshold).Count -eq 0) "no module can be below planned coverage threshold"
  Assert-True ($result.atomic_task_count -ge 120) "AEXS must generate at least five atomic task records for every target module"
  Assert-True ($result.invalid_atomic_task_count -eq 0) "AEXS must not generate invalid atomic task records"
  Assert-True (@($result.modules_with_invalid_atomic_tasks).Count -eq 0) "no module can have invalid atomic task records"
  Assert-True ($result.invalid_stop_condition_count -eq 0) "AEXS must not generate invalid stop conditions"
  Assert-True ($result.scenario_generator_count -ge 11) "AEXS must record all required scenario generator families"
  Assert-True ($result.invalid_scenario_generator_count -eq 0) "AEXS must not generate invalid scenario generator records"
  Assert-True ($result.transaction_mutator_count -ge 17) "AEXS must record all required transaction mutator families"
  Assert-True ($result.invalid_transaction_mutator_count -eq 0) "AEXS must not generate invalid transaction mutator records"
  Assert-True ($result.invariant_checklist_count -ge 17) "AEXS must record the required economic and consensus invariant checklist"
  Assert-True ($result.invalid_invariant_checklist_count -eq 0) "AEXS must not generate invalid invariant checklist records"
  Assert-True ($result.core_exploit_count -ge 13) "AEXS must record consensus and Aether Core exploit catalog entries"
  Assert-True ($result.invalid_core_exploit_count -eq 0) "AEXS must not generate invalid consensus exploit records"
  Assert-True ($result.slashing_exploit_count -ge 7) "AEXS must record slashing bypass exploit catalog entries"
  Assert-True ($result.invalid_slashing_exploit_count -eq 0) "AEXS must not generate invalid slashing exploit records"
  Assert-True ($result.tx_auth_bank_exploit_count -ge 11) "AEXS must record transaction/auth/bank exploit catalog entries"
  Assert-True ($result.invalid_tx_auth_bank_exploit_count -eq 0) "AEXS must not generate invalid transaction/auth/bank exploit records"
  Assert-True ($result.token_economy_exploit_count -ge 10) "AEXS must record token/economy exploit catalog entries"
  Assert-True ($result.invalid_token_economy_exploit_count -eq 0) "AEXS must not generate invalid token/economy exploit records"
  Assert-True ($result.dex_exploit_count -ge 10) "AEXS must record DEX exploit catalog entries"
  Assert-True ($result.invalid_dex_exploit_count -eq 0) "AEXS must not generate invalid DEX exploit records"
  Assert-True ($result.load_system_exploit_count -ge 9) "AEXS must record load system exploit catalog entries"
  Assert-True ($result.invalid_load_system_exploit_count -eq 0) "AEXS must not generate invalid load system exploit records"
  Assert-True ($result.routing_engine_exploit_count -ge 9) "AEXS must record routing engine exploit catalog entries"
  Assert-True ($result.invalid_routing_engine_exploit_count -eq 0) "AEXS must not generate invalid routing engine exploit records"
  Assert-True ($result.execution_zone_avm_exploit_count -ge 15) "AEXS must record execution zone and AVM exploit catalog entries"
  Assert-True ($result.invalid_execution_zone_avm_exploit_count -eq 0) "AEXS must not generate invalid execution zone and AVM exploit records"
  Assert-True ($result.compute_shard_exploit_count -ge 10) "AEXS must record compute shard exploit catalog entries"
  Assert-True ($result.invalid_compute_shard_exploit_count -eq 0) "AEXS must not generate invalid compute shard exploit records"
  Assert-True ($result.mesh_cross_zone_exploit_count -ge 10) "AEXS must record Aether Mesh and cross-zone exploit catalog entries"
  Assert-True ($result.invalid_mesh_cross_zone_exploit_count -eq 0) "AEXS must not generate invalid Aether Mesh and cross-zone exploit records"
  Assert-True ($result.identity_domain_exploit_count -ge 10) "AEXS must record identity and .aet domain exploit catalog entries"
  Assert-True ($result.invalid_identity_domain_exploit_count -eq 0) "AEXS must not generate invalid identity and .aet exploit records"
  Assert-True ($result.governance_exploit_count -ge 9) "AEXS must record governance exploit catalog entries"
  Assert-True ($result.invalid_governance_exploit_count -eq 0) "AEXS must not generate invalid governance exploit records"
  Assert-True ($result.genesis_upgrade_state_exploit_count -ge 10) "AEXS must record genesis, upgrade, and state exploit catalog entries"
  Assert-True ($result.invalid_genesis_upgrade_state_exploit_count -eq 0) "AEXS must not generate invalid genesis, upgrade, and state exploit records"
  Assert-True ($result.mempool_network_exploit_count -ge 10) "AEXS must record mempool and network exploit catalog entries"
  Assert-True ($result.invalid_mempool_network_exploit_count -eq 0) "AEXS must not generate invalid mempool and network exploit records"
  Assert-True ($result.combined_full_stack_exploit_count -ge 10) "AEXS must record combined full-stack exploit catalog entries"
  Assert-True ($result.invalid_combined_full_stack_exploit_count -eq 0) "AEXS must not generate invalid combined full-stack exploit records"
  Assert-True ($result.exploit_count -ge 153) "AEXS must record all current exploit catalog entries"
  Assert-True ($result.invalid_exploit_count -eq 0) "AEXS must not generate invalid exploit records"

  foreach ($module in @(
      "app",
      "x/fees",
      "x/aetherisvm/standards/aft",
      "avm-dex-contract",
      "x/aetherisvm",
      "x/execution",
      "x/vm",
      "x/messaging",
      "x/queue",
      "x/events",
      "x/actors",
      "x/scheduler",
      "x/storage",
      "x/identity",
      "x/reputation",
      "x/sharding/sim"
    )) {
    Assert-Contains -Values $result.target_modules -Expected $module -Message "AEXS target module missing: $module"
  }

  foreach ($name in @(
      "summary.json",
      "campaign-setup.json",
      "coverage-matrix.json",
      "atomic-tasks.json",
      "atomic-tasks.md",
      "invariant-checklist.json",
      "invariant-checklist.md",
      "exploit-catalog.json",
      "exploit-catalog.md",
      "scenario-generator.json",
      "scenario-generator.md",
      "transaction-mutator.json",
      "transaction-mutator.md",
      "AUDIT_RESULT.md",
      "TO_AUDIT.md"
    )) {
    Assert-True (Test-Path -LiteralPath (Join-Path $result.output_dir $name)) "AEXS output missing $name"
  }

  $invariantChecklist = Get-Content -Raw -LiteralPath (Join-Path $result.output_dir "invariant-checklist.json") | ConvertFrom-Json
  Assert-True (@($invariantChecklist).Count -eq $result.invariant_checklist_count) "summary invariant checklist count must match invariant-checklist.json"
  $invariantById = @{}
  foreach ($record in $invariantChecklist) {
    $invariantById[$record.invariant_id] = $record
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
      Assert-True (-not [string]::IsNullOrWhiteSpace([string]$record.$field)) "invariant checklist $($record.invariant_id) missing $field"
    }
    foreach ($field in @(
        "status",
        "expected_behavior",
        "expected_state_transition",
        "expected_events",
        "expected_error_path",
        "expected_invariant"
      )) {
      Assert-True (-not [string]::IsNullOrWhiteSpace([string]$record.defensive_analysis_result.$field)) "invariant checklist $($record.invariant_id) missing defensive_analysis_result.$field"
    }
    foreach ($field in @(
        "status",
        "attack_attempt",
        "mutation_inputs",
        "expected_rejection",
        "replay_mode"
      )) {
      Assert-True (-not [string]::IsNullOrWhiteSpace([string]$record.adversarial_simulation_result.$field)) "invariant checklist $($record.invariant_id) missing adversarial_simulation_result.$field"
    }
    Assert-True (-not [string]::IsNullOrWhiteSpace([string]$record.reproduction_seed_or_steps.seed)) "invariant checklist $($record.invariant_id) missing reproduction seed"
    Assert-True (@($record.reproduction_seed_or_steps.steps).Count -gt 0) "invariant checklist $($record.invariant_id) missing reproduction steps"
    Assert-True ($record.pass_fail_result -eq "not_executed") "preflight invariant $($record.invariant_id) must stay not_executed"
    Assert-True ($record.valid -eq $true) "invariant checklist $($record.invariant_id) must be valid"
  }
  foreach ($invariantId in @(
      "ECON-01",
      "ECON-02",
      "ECON-03",
      "ECON-04",
      "ECON-05",
      "ECON-06",
      "ECON-07",
      "ECON-08",
      "ECON-09",
      "STATE-01",
      "STATE-02",
      "STATE-03",
      "STATE-04",
      "STATE-05",
      "STATE-06",
      "STATE-07",
      "STATE-08",
      "DEXINV-01",
      "DEXINV-02",
      "DEXINV-03",
      "DEXINV-04",
      "DEXINV-05",
      "DEXINV-06",
      "DEXINV-07",
      "DEXINV-08",
      "LOAD-01",
      "LOAD-02",
      "LOAD-03",
      "LOAD-04",
      "LOAD-05",
      "LOAD-06",
      "LOAD-07",
      "LOAD-08",
      "LOAD-09",
      "IDINV-01",
      "IDINV-02",
      "IDINV-03",
      "IDINV-04",
      "IDINV-05",
      "IDINV-06",
      "IDINV-07",
      "IDINV-08",
      "EXECINV-01",
      "EXECINV-02",
      "EXECINV-03",
      "EXECINV-04",
      "EXECINV-05",
      "EXECINV-06",
      "EXECINV-07",
      "EXECINV-08",
      "EXECINV-09"
    )) {
    Assert-True ($invariantById.ContainsKey($invariantId)) "mandatory invariant checklist record missing: $invariantId"
  }
  Assert-True ($invariantById["ECON-01"].invariant_tested -match "total_supply") "ECON-01 must record the global supply equation"
  Assert-True ($invariantById["ECON-05"].attack_surface_covered -match "fee denom spoofing") "ECON-05 must record non-naet fee attack surface"
  Assert-True ($invariantById["ECON-06"].state_transition_covered -match "Fee distribution totals match collected fees") "ECON-06 must record collected fee distribution totals"
  Assert-True ($invariantById["ECON-08"].attack_surface_covered -match "reward farming loop") "ECON-08 must record staking reward farming"
  Assert-True ($invariantById["ECON-09"].state_transition_covered -match "Supply cannot drift after deterministic export/import") "ECON-09 must record export/import supply drift"
  Assert-True ($invariantById["STATE-01"].state_transition_covered -match "Same block input produces same app hash") "STATE-01 must record app hash determinism"
  Assert-True ($invariantById["STATE-02"].attack_surface_covered -match "signed byte replay") "STATE-02 must record tx replay"
  Assert-True ($invariantById["STATE-05"].state_transition_covered -match "validator set matches staking keeper state") "STATE-05 must record validator set/staking consistency"
  Assert-True ($invariantById["STATE-06"].attack_surface_covered -match "malformed proof") "STATE-06 must record slashing evidence adversarial path"
  Assert-True ($invariantById["STATE-07"].defensive_analysis_result.expected_error_path -match "malformed genesis") "STATE-07 must record malformed genesis rejection"
  Assert-True ($invariantById["STATE-08"].state_transition_covered -match "preserve state roots") "STATE-08 must record migration root preservation"
  Assert-True ($invariantById["DEXINV-01"].attack_surface_covered -match "reserve desync") "DEXINV-01 must record reserve desync"
  Assert-True ($invariantById["DEXINV-02"].attack_surface_covered -match "LP inflation") "DEXINV-02 must record LP inflation"
  Assert-True ($invariantById["DEXINV-03"].attack_surface_covered -match "underflow") "DEXINV-03 must record liquidity underflow"
  Assert-True ($invariantById["DEXINV-04"].attack_surface_covered -match "fake LP token|forged LP token") "DEXINV-04 must record fake LP token"
  Assert-True ($invariantById["DEXINV-05"].attack_surface_covered -match "rounding underflow") "DEXINV-05 must record non-negative swap output adversarial path"
  Assert-True ($invariantById["DEXINV-06"].attack_surface_covered -match "slippage bypass") "DEXINV-06 must record slippage bypass"
  Assert-True ($invariantById["DEXINV-07"].attack_surface_covered -match "constant-product break") "DEXINV-07 must record constant-product break"
  Assert-True ($invariantById["DEXINV-08"].attack_surface_covered -match "failed send partial update") "DEXINV-08 must record failed bank movement partial update"
  Assert-True ($invariantById["LOAD-01"].attack_surface_covered -match "out-of-range metric") "LOAD-01 must record load score bounds"
  Assert-True ($invariantById["LOAD-02"].attack_surface_covered -match "node-local latency") "LOAD-02 must record deterministic EMA risks"
  Assert-True ($invariantById["LOAD-03"].attack_surface_covered -match "spam burst") "LOAD-03 must record MAX_DELTA spike abuse"
  Assert-True ($invariantById["LOAD-04"].attack_surface_covered -match "routing hint manipulation") "LOAD-04 must record zone route manipulation"
  Assert-True ($invariantById["LOAD-05"].attack_surface_covered -match "routing epoch manipulation") "LOAD-05 must record shard route manipulation"
  Assert-True ($invariantById["LOAD-06"].attack_surface_covered -match "route cycle") "LOAD-06 must record routing loop"
  Assert-True ($invariantById["LOAD-07"].attack_surface_covered -match "starvation") "LOAD-07 must record shard starvation"
  Assert-True ($invariantById["LOAD-08"].attack_surface_covered -match "single-zone spam") "LOAD-08 must record hot-zone monopolization"
  Assert-True ($invariantById["LOAD-09"].attack_surface_covered -match "local mempool order") "LOAD-09 must record priority divergence"
  Assert-True ($invariantById["IDINV-01"].attack_surface_covered -match "duplicate normalized name") "IDINV-01 must record duplicate normalized name"
  Assert-True ($invariantById["IDINV-02"].attack_surface_covered -match "active domain re-auction") "IDINV-02 must record active domain re-auction"
  Assert-True ($invariantById["IDINV-03"].attack_surface_covered -match "expired domain direct takeover") "IDINV-03 must record expired domain takeover"
  Assert-True ($invariantById["IDINV-04"].attack_surface_covered -match "zero address resolver") "IDINV-04 must record zero resolver target"
  Assert-True ($invariantById["IDINV-05"].attack_surface_covered -match "unresolved name payment") "IDINV-05 must record unresolved resolver payment"
  Assert-True ($invariantById["IDINV-06"].attack_surface_covered -match "reverse lookup poisoning") "IDINV-06 must record reverse lookup poisoning"
  Assert-True ($invariantById["IDINV-07"].attack_surface_covered -match "NFT transfer without registry update") "IDINV-07 must record registry/NFT divergence"
  Assert-True ($invariantById["IDINV-08"].attack_surface_covered -match "parent policy bypass") "IDINV-08 must record subdomain parent policy bypass"
  Assert-True ($invariantById["EXECINV-01"].attack_surface_covered -match "malformed bytecode") "EXECINV-01 must record malformed AVM input"
  Assert-True ($invariantById["EXECINV-02"].attack_surface_covered -match "gas underpayment") "EXECINV-02 must record gas abuse"
  Assert-True ($invariantById["EXECINV-03"].attack_surface_covered -match "infinite loop bytecode") "EXECINV-03 must record infinite loop bytecode"
  Assert-True ($invariantById["EXECINV-04"].attack_surface_covered -match "wall-clock dependency") "EXECINV-04 must record nondeterministic host access"
  Assert-True ($invariantById["EXECINV-05"].attack_surface_covered -match "queue insertion order drift") "EXECINV-05 must record queue ordering drift"
  Assert-True ($invariantById["EXECINV-06"].attack_surface_covered -match "duplicate mesh message") "EXECINV-06 must record cross-zone replay"
  Assert-True ($invariantById["EXECINV-07"].attack_surface_covered -match "duplicate refund receipt") "EXECINV-07 must record refund double spend"
  Assert-True ($invariantById["EXECINV-08"].attack_surface_covered -match "recursive message loop") "EXECINV-08 must record message loop bounds"
  Assert-True ($invariantById["EXECINV-09"].attack_surface_covered -match "queue export ordering drift") "EXECINV-09 must record queue export/import drift"

  $exploitCatalog = Get-Content -Raw -LiteralPath (Join-Path $result.output_dir "exploit-catalog.json") | ConvertFrom-Json
  Assert-True (@($exploitCatalog).Count -eq $result.exploit_count) "summary exploit count must match exploit-catalog.json"
  $exploitById = @{}
  foreach ($exploit in $exploitCatalog) {
    $exploitById[$exploit.exploit_id] = $exploit
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
      Assert-True (-not [string]::IsNullOrWhiteSpace([string]$exploit.$field)) "exploit record $($exploit.exploit_id) missing $field"
    }
    Assert-True (@($exploit.step_list).Count -gt 0) "exploit record $($exploit.exploit_id) must include step list"
    Assert-True (@($exploit.affected_modules).Count -gt 0) "exploit record $($exploit.exploit_id) must include affected modules"
    Assert-True ($exploit.actual_state -eq "not_executed_preflight") "preflight exploit $($exploit.exploit_id) must record actual state as not executed"
    Assert-True ($exploit.status -eq "planned_not_executed") "preflight exploit $($exploit.exploit_id) must stay planned"
    Assert-True ($exploit.valid -eq $true) "exploit record $($exploit.exploit_id) must be valid"
  }
  foreach ($exploitId in @(
      "COREEXP-01",
      "COREEXP-02",
      "COREEXP-03",
      "COREEXP-04",
      "COREEXP-05",
      "COREEXP-06",
      "COREEXP-07",
      "COREEXP-08",
      "COREEXP-09",
      "COREEXP-10",
      "COREEXP-11",
      "COREEXP-12",
      "COREEXP-13",
      "SLASHEXP-01",
      "SLASHEXP-02",
      "SLASHEXP-03",
      "SLASHEXP-04",
      "SLASHEXP-05",
      "SLASHEXP-06",
      "SLASHEXP-07",
      "TXEXP-01",
      "TXEXP-02",
      "TXEXP-03",
      "TXEXP-04",
      "TXEXP-05",
      "TXEXP-06",
      "TXEXP-07",
      "TXEXP-08",
      "TXEXP-09",
      "TXEXP-10",
      "TXEXP-11",
      "TOKENEXP-01",
      "TOKENEXP-02",
      "TOKENEXP-03",
      "TOKENEXP-04",
      "TOKENEXP-05",
      "TOKENEXP-06",
      "TOKENEXP-07",
      "TOKENEXP-08",
      "TOKENEXP-09",
      "TOKENEXP-10",
      "DEXEXP-01",
      "DEXEXP-02",
      "DEXEXP-03",
      "DEXEXP-04",
      "DEXEXP-05",
      "DEXEXP-06",
      "DEXEXP-07",
      "DEXEXP-08",
      "DEXEXP-09",
      "DEXEXP-10",
      "LOADEXP-01",
      "LOADEXP-02",
      "LOADEXP-03",
      "LOADEXP-04",
      "LOADEXP-05",
      "LOADEXP-06",
      "LOADEXP-07",
      "LOADEXP-08",
      "LOADEXP-09",
      "ROUTEEXP-01",
      "ROUTEEXP-02",
      "ROUTEEXP-03",
      "ROUTEEXP-04",
      "ROUTEEXP-05",
      "ROUTEEXP-06",
      "ROUTEEXP-07",
      "ROUTEEXP-08",
      "ROUTEEXP-09",
      "EXECZONEEXP-01",
      "EXECZONEEXP-02",
      "EXECZONEEXP-03",
      "EXECZONEEXP-04",
      "EXECZONEEXP-05",
      "EXECZONEEXP-06",
      "EXECZONEEXP-07",
      "EXECZONEEXP-08",
      "EXECZONEEXP-09",
      "EXECZONEEXP-10",
      "EXECZONEEXP-11",
      "EXECZONEEXP-12",
      "EXECZONEEXP-13",
      "EXECZONEEXP-14",
      "EXECZONEEXP-15",
      "SHARDEXP-01",
      "SHARDEXP-02",
      "SHARDEXP-03",
      "SHARDEXP-04",
      "SHARDEXP-05",
      "SHARDEXP-06",
      "SHARDEXP-07",
      "SHARDEXP-08",
      "SHARDEXP-09",
      "SHARDEXP-10",
      "MESHEXP-01",
      "MESHEXP-02",
      "MESHEXP-03",
      "MESHEXP-04",
      "MESHEXP-05",
      "MESHEXP-06",
      "MESHEXP-07",
      "MESHEXP-08",
      "MESHEXP-09",
      "MESHEXP-10",
      "IDENTEXP-01",
      "IDENTEXP-02",
      "IDENTEXP-03",
      "IDENTEXP-04",
      "IDENTEXP-05",
      "IDENTEXP-06",
      "IDENTEXP-07",
      "IDENTEXP-08",
      "IDENTEXP-09",
      "IDENTEXP-10",
      "GOVEXP-01",
      "GOVEXP-02",
      "GOVEXP-03",
      "GOVEXP-04",
      "GOVEXP-05",
      "GOVEXP-06",
      "GOVEXP-07",
      "GOVEXP-08",
      "GOVEXP-09",
      "STATEEXP-01",
      "STATEEXP-02",
      "STATEEXP-03",
      "STATEEXP-04",
      "STATEEXP-05",
      "STATEEXP-06",
      "STATEEXP-07",
      "STATEEXP-08",
      "STATEEXP-09",
      "STATEEXP-10",
      "NETEXP-01",
      "NETEXP-02",
      "NETEXP-03",
      "NETEXP-04",
      "NETEXP-05",
      "NETEXP-06",
      "NETEXP-07",
      "NETEXP-08",
      "NETEXP-09",
      "NETEXP-10",
      "FULLSTACKEXP-01",
      "FULLSTACKEXP-02",
      "FULLSTACKEXP-03",
      "FULLSTACKEXP-04",
      "FULLSTACKEXP-05",
      "FULLSTACKEXP-06",
      "FULLSTACKEXP-07",
      "FULLSTACKEXP-08",
      "FULLSTACKEXP-09",
      "FULLSTACKEXP-10"
    )) {
    Assert-True ($exploitById.ContainsKey($exploitId)) "exploit catalog record missing: $exploitId"
  }
  Assert-True ($exploitById["COREEXP-01"].exploit_path -match "conflicting blocks") "COREEXP-01 must record double-sign fork path"
  Assert-True ($exploitById["COREEXP-01"].severity -eq "Critical") "COREEXP-01 must be Critical"
  Assert-True ($exploitById["COREEXP-03"].exploit_path -match "long-range history rewrite") "COREEXP-03 must record long-range rewrite"
  Assert-True (@($exploitById["COREEXP-03"].affected_modules) -contains "x/staking") "COREEXP-03 must affect staking"
  Assert-True ($exploitById["COREEXP-07"].exploit_path -match "self-delegation inflation") "COREEXP-07 must record self-delegation inflation"
  Assert-True ($exploitById["COREEXP-11"].exploit_path -match "fork choice manipulation") "COREEXP-11 must record fork choice manipulation"
  Assert-True ($exploitById["COREEXP-13"].exploit_path -match "Byzantine majority") "COREEXP-13 must record Byzantine majority simulator"
  Assert-True ($exploitById["SLASHEXP-01"].exploit_path -match "evidence after delay") "SLASHEXP-01 must record delayed evidence bypass"
  Assert-True ($exploitById["SLASHEXP-02"].exploit_path -match "malformed equivocation proof") "SLASHEXP-02 must record malformed equivocation proof"
  Assert-True ($exploitById["SLASHEXP-03"].exploit_path -match "race slashing evidence") "SLASHEXP-03 must record slashing race"
  Assert-True ($exploitById["SLASHEXP-05"].exploit_path -match "unbond stake") "SLASHEXP-05 must record unbonding slash evasion"
  Assert-True ($exploitById["SLASHEXP-06"].exploit_path -match "jailed validator") "SLASHEXP-06 must record jail escape through upgrade timing"
  Assert-True ($exploitById["SLASHEXP-07"].exploit_path -match "invalid/stale/duplicate evidence") "SLASHEXP-07 must record invalid evidence replay"
  Assert-True ($exploitById["TXEXP-01"].exploit_path -match "signed transaction bytes") "TXEXP-01 must record signature replay"
  Assert-True ($exploitById["TXEXP-02"].exploit_path -match "wrong chain id") "TXEXP-02 must record cross-context replay"
  Assert-True ($exploitById["TXEXP-03"].exploit_path -match "account sequence") "TXEXP-03 must record invalid nonce"
  Assert-True ($exploitById["TXEXP-05"].exploit_path -match "underpay fee") "TXEXP-05 must record fee underpayment"
  Assert-True ($exploitById["TXEXP-08"].exploit_path -match "multi-send") "TXEXP-08 must record multi-send partial failure"
  Assert-True ($exploitById["TXEXP-09"].exploit_path -match "double spend") "TXEXP-09 must record race-condition double spend"
  Assert-True ($exploitById["TXEXP-11"].exploit_path -match "zero address") "TXEXP-11 must record zero-address path"
  Assert-True ($exploitById["TOKENEXP-01"].exploit_path -match "admin takeover") "TOKENEXP-01 must record contract-assets admin takeover"
  Assert-True ($exploitById["TOKENEXP-02"].exploit_path -match "unauthorized burn") "TOKENEXP-02 must record unauthorized burn"
  Assert-True ($exploitById["TOKENEXP-03"].exploit_path -match "governance parameter changes") "TOKENEXP-03 must record governance inflation timing"
  Assert-True ($exploitById["TOKENEXP-04"].exploit_path -match "fee routing") "TOKENEXP-04 must record fee routing manipulation"
  Assert-True ($exploitById["TOKENEXP-05"].exploit_path -match "treasury drain") "TOKENEXP-05 must record treasury drain"
  Assert-True ($exploitById["TOKENEXP-06"].exploit_path -match "staking rewards") "TOKENEXP-06 must record staking reward inflation"
  Assert-True ($exploitById["TOKENEXP-07"].exploit_path -match "delegate, redelegate, unbond") "TOKENEXP-07 must record reward farming loop"
  Assert-True ($exploitById["TOKENEXP-08"].exploit_path -match "edge-case mint") "TOKENEXP-08 must record edge-case mint path"
  Assert-True ($exploitById["TOKENEXP-09"].exploit_path -match "Spoof native denom|spoof native denom") "TOKENEXP-09 must record native denom spoofing"
  Assert-True ($exploitById["TOKENEXP-10"].exploit_path -match "display/base decimal mismatch") "TOKENEXP-10 must record decimal mismatch"
  Assert-True ($exploitById["DEXEXP-01"].exploit_path -match "constant product") "DEXEXP-01 must record constant-product break"
  Assert-True ($exploitById["DEXEXP-02"].exploit_path -match "drain pool liquidity") "DEXEXP-02 must record liquidity drain"
  Assert-True ($exploitById["DEXEXP-03"].exploit_path -match "initialize pool") "DEXEXP-03 must record pool initialization manipulation"
  Assert-True ($exploitById["DEXEXP-04"].exploit_path -match "Inflate LP tokens|inflate LP tokens") "DEXEXP-04 must record LP token inflation"
  Assert-True ($exploitById["DEXEXP-05"].exploit_path -match "race liquidity removal") "DEXEXP-05 must record liquidity removal race"
  Assert-True ($exploitById["DEXEXP-06"].exploit_path -match "zero-liquidity") "DEXEXP-06 must record zero-liquidity edge"
  Assert-True ($exploitById["DEXEXP-07"].exploit_path -match "Desynchronize pool reserves|desynchronize pool reserves") "DEXEXP-07 must record reserve/module desync"
  Assert-True ($exploitById["DEXEXP-08"].exploit_path -match "bank movement failure") "DEXEXP-08 must record failed bank movement partial update"
  Assert-True ($exploitById["DEXEXP-09"].exploit_path -match "Bypass slippage|bypass slippage") "DEXEXP-09 must record slippage bypass"
  Assert-True ($exploitById["DEXEXP-10"].exploit_path -match "rounding") "DEXEXP-10 must record rounding exploit"
  Assert-True ($exploitById["LOADEXP-01"].exploit_path -match "LOAD_SCORE") "LOADEXP-01 must record LOAD_SCORE spam manipulation"
  Assert-True ($exploitById["LOADEXP-02"].exploit_path -match "mempool size") "LOADEXP-02 must record artificial mempool inflation"
  Assert-True ($exploitById["LOADEXP-03"].exploit_path -match "Saturate blocks|saturate blocks") "LOADEXP-03 must record block saturation"
  Assert-True ($exploitById["LOADEXP-04"].exploit_path -match "execution delay") "LOADEXP-04 must record execution delay amplification"
  Assert-True ($exploitById["LOADEXP-05"].exploit_path -match "poison EMA") "LOADEXP-05 must record EMA slow-poison attack"
  Assert-True ($exploitById["LOADEXP-06"].exploit_path -match "oscillate load") "LOADEXP-06 must record load spike oscillation"
  Assert-True ($exploitById["LOADEXP-07"].exploit_path -match "overload a shard") "LOADEXP-07 must record shard overload targeting"
  Assert-True ($exploitById["LOADEXP-08"].exploit_path -match "priority fees") "LOADEXP-08 must record priority fee gaming"
  Assert-True ($exploitById["LOADEXP-09"].exploit_path -match "adaptive fees") "LOADEXP-09 must record adaptive fee destabilization"
  Assert-True ($exploitById["ROUTEEXP-01"].exploit_path -match "bias routing decisions") "ROUTEEXP-01 must record routing bias exploitation"
  Assert-True ($exploitById["ROUTEEXP-02"].exploit_path -match "execution zone") "ROUTEEXP-02 must record zone congestion targeting"
  Assert-True ($exploitById["ROUTEEXP-03"].exploit_path -match "starve compute shards") "ROUTEEXP-03 must record compute shard starvation"
  Assert-True ($exploitById["ROUTEEXP-04"].exploit_path -match "hot zone") "ROUTEEXP-04 must record hot-zone monopolization"
  Assert-True ($exploitById["ROUTEEXP-05"].exploit_path -match "Predict deterministic routes|predict deterministic routes") "ROUTEEXP-05 must record route prediction abuse"
  Assert-True ($exploitById["ROUTEEXP-06"].exploit_path -match "cross-zone routing loops") "ROUTEEXP-06 must record cross-zone routing loop"
  Assert-True ($exploitById["ROUTEEXP-07"].exploit_path -match "route desync") "ROUTEEXP-07 must record routing desync"
  Assert-True ($exploitById["ROUTEEXP-08"].exploit_path -match "Misclassify transactions|misclassify transactions") "ROUTEEXP-08 must record transaction misclassification"
  Assert-True ($exploitById["ROUTEEXP-09"].exploit_path -match "fee-class overflow") "ROUTEEXP-09 must record fee-based routing gaming"
  Assert-True ($exploitById["EXECZONEEXP-01"].exploit_path -match "state roots") "EXECZONEEXP-01 must record execution zone state divergence"
  Assert-True ($exploitById["EXECZONEEXP-03"].exploit_path -match "local time") "EXECZONEEXP-03 must record AVM determinism violation"
  Assert-True ($exploitById["EXECZONEEXP-07"].exploit_path -match "partial contract writes") "EXECZONEEXP-07 must record partial rollback"
  Assert-True ($exploitById["EXECZONEEXP-12"].exploit_path -match "hijack contract upgrade") "EXECZONEEXP-12 must record upgrade hijack"
  Assert-True ($exploitById["EXECZONEEXP-15"].exploit_path -match "sandbox escape") "EXECZONEEXP-15 must record sandbox escape"
  Assert-True ($exploitById["SHARDEXP-01"].exploit_path -match "partition imbalance") "SHARDEXP-01 must record shard partition imbalance"
  Assert-True ($exploitById["SHARDEXP-04"].exploit_path -match "cross-shard inconsistency") "SHARDEXP-04 must record cross-shard inconsistency"
  Assert-True ($exploitById["SHARDEXP-05"].exploit_path -match "shard activation") "SHARDEXP-05 must record load spoofing for shard activation"
  Assert-True ($exploitById["SHARDEXP-10"].exploit_path -match "flood shard queues") "SHARDEXP-10 must record queue flooding"
  Assert-True ($exploitById["MESHEXP-01"].exploit_path -match "cross-zone messages") "MESHEXP-01 must record cross-zone message replay"
  Assert-True ($exploitById["MESHEXP-04"].exploit_path -match "duplicate asset commitments") "MESHEXP-04 must record asset duplication"
  Assert-True ($exploitById["MESHEXP-06"].exploit_path -match "forge source proof") "MESHEXP-06 must record proof forgery"
  Assert-True ($exploitById["MESHEXP-10"].exploit_path -match "stale receipts") "MESHEXP-10 must record stale receipt replay"
  Assert-True ($exploitById["IDENTEXP-01"].exploit_path -match "overwriting resolver") "IDENTEXP-01 must record resolver overwrite hijack"
  Assert-True ($exploitById["IDENTEXP-02"].exploit_path -match "expired domain") "IDENTEXP-02 must record expired domain takeover"
  Assert-True ($exploitById["IDENTEXP-06"].exploit_path -match "reverse lookup") "IDENTEXP-06 must record reverse lookup poisoning"
  Assert-True ($exploitById["IDENTEXP-10"].exploit_path -match "multi-resolver inconsistency") "IDENTEXP-10 must record multi-resolver inconsistency"
  Assert-True ($exploitById["GOVEXP-01"].exploit_path -match "voting power") "GOVEXP-01 must record governance capture"
  Assert-True ($exploitById["GOVEXP-03"].exploit_path -match "emergency parameters") "GOVEXP-03 must record emergency parameter abuse"
  Assert-True ($exploitById["GOVEXP-04"].exploit_path -match "upgrade plan") "GOVEXP-04 must record upgrade hijack"
  Assert-True ($exploitById["GOVEXP-09"].exploit_path -match "grief parameters") "GOVEXP-09 must record parameter griefing"
  Assert-True ($exploitById["STATEEXP-01"].exploit_path -match "malformed genesis") "STATEEXP-01 must record malformed genesis injection"
  Assert-True ($exploitById["STATEEXP-02"].exploit_path -match "exported state") "STATEEXP-02 must record state export tampering"
  Assert-True ($exploitById["STATEEXP-06"].exploit_path -match "hidden privileged account") "STATEEXP-06 must record privileged account injection"
  Assert-True ($exploitById["STATEEXP-10"].exploit_path -match "snapshot") "STATEEXP-10 must record snapshot poisoning"
  Assert-True ($exploitById["NETEXP-01"].exploit_path -match "Flood mempool|flood mempool") "NETEXP-01 must record mempool flooding"
  Assert-True ($exploitById["NETEXP-03"].exploit_path -match "gossip") "NETEXP-03 must record gossip poisoning"
  Assert-True ($exploitById["NETEXP-07"].exploit_path -match "reorder transactions") "NETEXP-07 must record transaction reordering"
  Assert-True ($exploitById["NETEXP-08"].exploit_path -match "network latency") "NETEXP-08 must record network latency exploitation"
  Assert-True ($exploitById["FULLSTACKEXP-01"].exploit_path -match "spam bursts") "FULLSTACKEXP-01 must record coordinated spam plus routing"
  Assert-True ($exploitById["FULLSTACKEXP-03"].exploit_path -match "DEX swaps") "FULLSTACKEXP-03 must record DEX plus mempool plus routing"
  Assert-True ($exploitById["FULLSTACKEXP-06"].exploit_path -match "identity resolution") "FULLSTACKEXP-06 must record identity plus routing hijack"
  Assert-True ($exploitById["FULLSTACKEXP-10"].exploit_path -match "full stack|full-stack") "FULLSTACKEXP-10 must record full-stack destabilization"

  $campaignSetup = Get-Content -Raw -LiteralPath (Join-Path $result.output_dir "campaign-setup.json") | ConvertFrom-Json
  Assert-True ($campaignSetup.campaign_id -eq $result.campaign_id) "campaign setup campaign id must match summary"
  Assert-True ($campaignSetup.git_commit -eq $result.git_commit) "campaign setup git commit must match summary"
  Assert-True ($campaignSetup.setup_complete -eq $true) "campaign setup must be complete"
  Assert-True (@($campaignSetup.fuzz_seeds).Count -eq @($result.fuzz_seeds).Count) "campaign setup must record fuzz seed list"
  Assert-True (@($campaignSetup.target_modules).Count -eq @($result.target_modules).Count) "campaign setup must record target modules"
  foreach ($mode in @(
      "stateless fuzzing",
      "stateful multi-block fuzzing",
      "adversarial red-team fuzzing",
      "deterministic replay",
      "stress mode",
      "chaos mode"
    )) {
    Assert-Contains -Values @($campaignSetup.runtime_modes | ForEach-Object { $_.name }) -Expected $mode -Message "runtime mode missing: $mode"
  }
  foreach ($mode in @(
      "in-memory app runner",
      "single-validator localnet",
      "multi-validator localnet",
      "sharding simulator"
    )) {
    Assert-Contains -Values @($campaignSetup.simulator_modes | ForEach-Object { $_.name }) -Expected $mode -Message "simulator mode missing: $mode"
  }
  foreach ($condition in @(
      "first_critical_exploit",
      "max_run_count",
      "max_wall_clock_duration",
      "coverage_threshold_reached",
      "deterministic_divergence"
    )) {
    Assert-Contains -Values @($campaignSetup.stop_conditions | ForEach-Object { $_.id }) -Expected $condition -Message "stop condition missing: $condition"
  }
  Assert-True (@($campaignSetup.stop_conditions | Where-Object { $_.valid -ne $true }).Count -eq 0) "all stop conditions must be valid"

  $scenarioCatalog = Get-Content -Raw -LiteralPath (Join-Path $result.output_dir "scenario-generator.json") | ConvertFrom-Json
  Assert-True ($scenarioCatalog.campaign_id -eq $result.campaign_id) "scenario catalog campaign id must match summary"
  Assert-True ($scenarioCatalog.generator_count -eq $result.scenario_generator_count) "scenario generator count must match summary"
  Assert-True ($scenarioCatalog.invalid_generator_count -eq 0) "scenario catalog must not contain invalid generators"
  Assert-True ($scenarioCatalog.seed_policy.deterministic_seed_required -eq $true) "scenario catalog must require deterministic seeds"
  Assert-True ($scenarioCatalog.seed_policy.step_list_required -eq $true) "scenario catalog must require step lists"
  foreach ($scenario in $scenarioCatalog.generators) {
    foreach ($field in @(
        "id",
        "name",
        "flow_covered",
        "state_transitions",
        "attack_surfaces",
        "invariant_targets",
        "status"
      )) {
      Assert-True (-not [string]::IsNullOrWhiteSpace([string]$scenario.$field)) "scenario generator $($scenario.id) missing $field"
    }
    Assert-True ($scenario.seed_required -eq $true) "scenario generator $($scenario.id) must require seed preservation"
    Assert-True ($scenario.step_list_required -eq $true) "scenario generator $($scenario.id) must require step list preservation"
    Assert-True ($scenario.valid -eq $true) "scenario generator $($scenario.id) must be valid"
  }

  $mutatorCatalog = Get-Content -Raw -LiteralPath (Join-Path $result.output_dir "transaction-mutator.json") | ConvertFrom-Json
  Assert-True ($mutatorCatalog.campaign_id -eq $result.campaign_id) "transaction mutator catalog campaign id must match summary"
  Assert-True ($mutatorCatalog.mutator_count -eq $result.transaction_mutator_count) "transaction mutator count must match summary"
  Assert-True ($mutatorCatalog.invalid_mutator_count -eq 0) "transaction mutator catalog must not contain invalid mutators"
  Assert-True ($mutatorCatalog.metadata_policy.mutation_metadata_required -eq $true) "transaction mutator catalog must require mutation metadata"
  Assert-True ($mutatorCatalog.metadata_policy.deterministic_seed_required -eq $true) "transaction mutator catalog must require deterministic seeds"
  Assert-True ($mutatorCatalog.metadata_policy.expected_rejection_required -eq $true) "transaction mutator catalog must require expected rejection paths"
  foreach ($mutatorId in @(
      "invalid_signatures",
      "replay_accepted_tx_bytes",
      "nonce_sequence_manipulation",
      "fee_field_corruption",
      "missing_or_non_naet_fee",
      "extreme_gas_values",
      "malformed_addresses",
      "zero_address_fields",
      "malformed_memo_fields",
      "malformed_routing_hints",
      "invalid_domain_resolution",
      "fake_cross_zone_messages",
      "queue_depth_abuse",
      "oversized_avm_payloads",
      "invalid_avm_entrypoints",
      "malformed_genesis_fragments",
      "mutation_metadata_recording"
    )) {
    Assert-Contains -Values @($mutatorCatalog.mutators | ForEach-Object { $_.id }) -Expected $mutatorId -Message "transaction mutator missing: $mutatorId"
  }
  foreach ($mutator in $mutatorCatalog.mutators) {
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
      Assert-True (-not [string]::IsNullOrWhiteSpace([string]$mutator.$field)) "transaction mutator $($mutator.id) missing $field"
    }
    Assert-True (@($mutator.target_modules).Count -gt 0) "transaction mutator $($mutator.id) must target at least one module"
    Assert-True ($mutator.seed_required -eq $true) "transaction mutator $($mutator.id) must require seed preservation"
    Assert-True ($mutator.metadata_required -eq $true) "transaction mutator $($mutator.id) must require mutation metadata"
    Assert-True ($mutator.valid -eq $true) "transaction mutator $($mutator.id) must be valid"
  }

  $coverage = Get-Content -Raw -LiteralPath (Join-Path $result.output_dir "coverage-matrix.json") | ConvertFrom-Json
  Assert-True (@($coverage).Count -ge 24) "coverage matrix must include all required module surfaces"
  Assert-True (@($coverage | Where-Object { $_.task_count -lt 5 }).Count -eq 0) "every module must have at least five tasks"
  Assert-True (@($coverage | Where-Object { $_.atomic_task_records -lt 5 }).Count -eq 0) "every module must have at least five atomic task records"
  Assert-True (@($coverage | Where-Object { @($_.invalid_atomic_tasks).Count -gt 0 }).Count -eq 0) "no module may contain invalid atomic task records"
  Assert-True (@($coverage | Where-Object { $_.planned_coverage_percent -lt 95 }).Count -eq 0) "every module must meet planned coverage threshold"
  Assert-True (@($coverage | Where-Object { $_.safe -eq $true }).Count -eq 0) "no module may be marked safe by preflight alone"

  $atomicTasks = Get-Content -Raw -LiteralPath (Join-Path $result.output_dir "atomic-tasks.json") | ConvertFrom-Json
  Assert-True (@($atomicTasks).Count -eq $result.atomic_task_count) "summary atomic task count must match atomic-tasks.json"
  $atomicTaskById = @{}
  foreach ($task in $atomicTasks) {
    $atomicTaskById[$task.task_id] = $task
    foreach ($field in @(
        "module",
        "task_id",
        "function_or_flow_covered",
        "state_transition_covered",
        "attack_surface_covered",
        "invariant_tested",
        "pass_fail_result"
      )) {
      Assert-True (-not [string]::IsNullOrWhiteSpace([string]$task.$field)) "atomic task $($task.task_id) missing $field"
    }
    foreach ($field in @(
        "status",
        "expected_behavior",
        "expected_state_transition",
        "expected_events",
        "expected_error_path",
        "expected_invariant"
      )) {
      Assert-True (-not [string]::IsNullOrWhiteSpace([string]$task.defensive_analysis_result.$field)) "atomic task $($task.task_id) missing defensive_analysis_result.$field"
    }
    foreach ($field in @(
        "status",
        "attack_attempt",
        "mutation_inputs",
        "expected_rejection",
        "replay_mode"
      )) {
      Assert-True (-not [string]::IsNullOrWhiteSpace([string]$task.adversarial_simulation_result.$field)) "atomic task $($task.task_id) missing adversarial_simulation_result.$field"
    }
    Assert-True (-not [string]::IsNullOrWhiteSpace([string]$task.reproduction_seed_or_steps.seed)) "atomic task $($task.task_id) missing reproduction seed"
    Assert-True (@($task.reproduction_seed_or_steps.steps).Count -gt 0) "atomic task $($task.task_id) missing reproduction steps"
    Assert-True ($task.pass_fail_result -eq "not_executed") "preflight atomic task $($task.task_id) must stay not_executed"
    Assert-True ($task.valid -eq $true) "atomic task $($task.task_id) must be valid"
  }
  foreach ($taskId in @(
      "AUTH-01",
      "AUTH-02",
      "AUTH-03",
      "AUTH-04",
      "AUTH-05",
      "BANK-01",
      "BANK-02",
      "BANK-03",
      "BANK-04",
      "BANK-05",
      "STAKE-01",
      "STAKE-02",
      "STAKE-03",
      "STAKE-04",
      "STAKE-05",
      "SLASH-01",
      "SLASH-02",
      "SLASH-03",
      "SLASH-04",
      "SLASH-05",
      "GOV-01",
      "GOV-02",
      "GOV-03",
      "GOV-04",
      "GOV-05",
      "DIST-01",
      "DIST-02",
      "DIST-03",
      "DIST-04",
      "DIST-05",
      "FEES-01",
      "FEES-02",
      "FEES-03",
      "FEES-04",
      "FEES-05",
      "TF-01",
      "TF-02",
      "TF-03",
      "TF-04",
      "TF-05",
      "DEX-01",
      "DEX-02",
      "DEX-03",
      "DEX-04",
      "DEX-05",
      "ID-01",
      "ID-02",
      "ID-03",
      "ID-04",
      "ID-05",
      "REP-01",
      "REP-02",
      "REP-03",
      "REP-04",
      "REP-05",
      "EXEC-01",
      "EXEC-02",
      "EXEC-03",
      "EXEC-04",
      "EXEC-05",
      "VM-01",
      "VM-02",
      "VM-03",
      "VM-04",
      "VM-05",
      "MSG-01",
      "MSG-02",
      "MSG-03",
      "MSG-04",
      "MSG-05",
      "QUEUE-01",
      "QUEUE-02",
      "QUEUE-03",
      "QUEUE-04",
      "QUEUE-05",
      "EVENTS-01",
      "EVENTS-02",
      "EVENTS-03",
      "EVENTS-04",
      "EVENTS-05",
      "ACTOR-01",
      "ACTOR-02",
      "ACTOR-03",
      "ACTOR-04",
      "ACTOR-05",
      "SCHED-01",
      "SCHED-02",
      "SCHED-03",
      "SCHED-04",
      "SCHED-05",
      "STORE-01",
      "STORE-02",
      "STORE-03",
      "STORE-04",
      "STORE-05",
      "MEMO-01",
      "MEMO-02",
      "MEMO-03",
      "MEMO-04",
      "MEMO-05",
      "INDEX-01",
      "INDEX-02",
      "INDEX-03",
      "INDEX-04",
      "INDEX-05",
      "SHARD-01",
      "SHARD-02",
      "SHARD-03",
      "SHARD-04",
      "SHARD-05"
    )) {
    Assert-True ($atomicTaskById.ContainsKey($taskId)) "required base-chain atomic task missing: $taskId"
  }
  Assert-True ($atomicTaskById["AUTH-01"].function_or_flow_covered -match "signature verification") "AUTH-01 must use task-specific signature flow"
  Assert-True ($atomicTaskById["AUTH-03"].adversarial_simulation_result.mutation_inputs -match "bit-flipped signature") "AUTH-03 must record concrete invalid signature mutation"
  Assert-True ($atomicTaskById["AUTH-04"].defensive_analysis_result.expected_state_transition -match "does not increment sequence") "AUTH-04 must record rejected auth state invariant"
  Assert-True ($atomicTaskById["AUTH-05"].adversarial_simulation_result.attack_attempt -match "fee bypass") "AUTH-05 must record fee/priority abuse surface"
  Assert-True ($atomicTaskById["BANK-01"].function_or_flow_covered -match "module account transfers") "BANK-01 must use task-specific transfer flow"
  Assert-True ($atomicTaskById["BANK-02"].adversarial_simulation_result.mutation_inputs -match "zero coin") "BANK-02 must record zero amount mutation"
  Assert-True ($atomicTaskById["BANK-03"].adversarial_simulation_result.expected_rejection -match "partial recipient credits") "BANK-03 must record atomic multi-send rejection"
  Assert-True ($atomicTaskById["BANK-04"].invariant_tested -match "total supply") "BANK-04 must record supply consistency invariant"
  Assert-True ($atomicTaskById["BANK-05"].adversarial_simulation_result.attack_attempt -match "native denom spoof") "BANK-05 must record native denom spoofing attack"
  Assert-True ($atomicTaskById["STAKE-01"].function_or_flow_covered -match "validator creation") "STAKE-01 must record staking lifecycle flow"
  Assert-True ($atomicTaskById["STAKE-02"].adversarial_simulation_result.mutation_inputs -match "non-naet bond denom") "STAKE-02 must record non-naet bond denom mutation"
  Assert-True ($atomicTaskById["STAKE-03"].adversarial_simulation_result.attack_attempt -match "stake grinding") "STAKE-03 must record stake grinding attack"
  Assert-True ($atomicTaskById["STAKE-04"].invariant_tested -match "validator tokens") "STAKE-04 must record validator token/share invariant"
  Assert-True ($atomicTaskById["STAKE-05"].adversarial_simulation_result.expected_rejection -match "extra rewards") "STAKE-05 must record reward inflation rejection"
  Assert-True ($atomicTaskById["SLASH-01"].function_or_flow_covered -match "downtime evidence") "SLASH-01 must record slashing evidence flow"
  Assert-True ($atomicTaskById["SLASH-02"].adversarial_simulation_result.mutation_inputs -match "duplicate evidence") "SLASH-02 must record duplicate evidence mutation"
  Assert-True ($atomicTaskById["SLASH-03"].adversarial_simulation_result.attack_attempt -match "redelegation slash evasion") "SLASH-03 must record redelegation slashing bypass"
  Assert-True ($atomicTaskById["SLASH-04"].invariant_tested -match "validator-set removal") "SLASH-04 must record validator-set removal invariant"
  Assert-True ($atomicTaskById["SLASH-05"].adversarial_simulation_result.expected_rejection -match "restore stake") "SLASH-05 must record slashed stake recovery rejection"
  Assert-True ($atomicTaskById["GOV-01"].function_or_flow_covered -match "proposal creation") "GOV-01 must record governance lifecycle flow"
  Assert-True ($atomicTaskById["GOV-02"].adversarial_simulation_result.mutation_inputs -match "zero deposit") "GOV-02 must record zero deposit mutation"
  Assert-True ($atomicTaskById["GOV-03"].adversarial_simulation_result.attack_attempt -match "upgrade hijack") "GOV-03 must record upgrade hijack attack"
  Assert-True ($atomicTaskById["GOV-04"].invariant_tested -match "authorized params") "GOV-04 must record authorized params invariant"
  Assert-True ($atomicTaskById["GOV-05"].adversarial_simulation_result.expected_rejection -match "hard protocol bounds") "GOV-05 must record economic hard-bounds rejection"
  Assert-True ($atomicTaskById["DIST-01"].function_or_flow_covered -match "validator commission") "DIST-01 must record distribution lifecycle flow"
  Assert-True ($atomicTaskById["DIST-02"].adversarial_simulation_result.mutation_inputs -match "rounding remainder") "DIST-02 must record rounding remainder mutation"
  Assert-True ($atomicTaskById["DIST-03"].adversarial_simulation_result.attack_attempt -match "reward double claim") "DIST-03 must record reward double claim attack"
  Assert-True ($atomicTaskById["DIST-04"].invariant_tested -match "outstanding rewards") "DIST-04 must record outstanding rewards invariant"
  Assert-True ($atomicTaskById["DIST-05"].adversarial_simulation_result.expected_rejection -match "treasury/community-pool funds") "DIST-05 must record treasury/community-pool leakage rejection"
  Assert-True ($atomicTaskById["FEES-01"].function_or_flow_covered -match "valid naet fee collection") "FEES-01 must record naet fee collection flow"
  Assert-True ($atomicTaskById["FEES-02"].adversarial_simulation_result.mutation_inputs -match "multi-denom") "FEES-02 must record multi-denom fee mutation"
  Assert-True ($atomicTaskById["FEES-03"].adversarial_simulation_result.attack_attempt -match "non-FeeTx bypass") "FEES-03 must record non-FeeTx bypass attack"
  Assert-True ($atomicTaskById["FEES-04"].invariant_tested -match "failed fee ante checks") "FEES-04 must record failed ante integrity invariant"
  Assert-True ($atomicTaskById["FEES-05"].adversarial_simulation_result.expected_rejection -match "validator reward accounting") "FEES-05 must record fee accounting manipulation rejection"
  Assert-True ($atomicTaskById["TF-01"].function_or_flow_covered -match "create denom") "TF-01 must record contract-assets lifecycle flow"
  Assert-True ($atomicTaskById["TF-02"].adversarial_simulation_result.mutation_inputs -match "zero admin") "TF-02 must record zero admin mutation"
  Assert-True ($atomicTaskById["TF-03"].adversarial_simulation_result.attack_attempt -match "burn-from mismatch") "TF-03 must record burn-from mismatch attack"
  Assert-True ($atomicTaskById["TF-04"].invariant_tested -match "supply delta") "TF-04 must record exact supply delta invariant"
  Assert-True ($atomicTaskById["TF-05"].adversarial_simulation_result.expected_rejection -match "spoof AET") "TF-05 must record native spoofing rejection"
  Assert-True ($atomicTaskById["DEX-01"].function_or_flow_covered -match "pool creation") "DEX-01 must record DEX lifecycle flow"
  Assert-True ($atomicTaskById["DEX-02"].adversarial_simulation_result.mutation_inputs -match "duplicate pair") "DEX-02 must record duplicate pair mutation"
  Assert-True ($atomicTaskById["DEX-03"].adversarial_simulation_result.attack_attempt -match "pool drain") "DEX-03 must record pool drain attack"
  Assert-True ($atomicTaskById["DEX-04"].invariant_tested -match "reserves match module balances") "DEX-04 must record reserves/balances invariant"
  Assert-True ($atomicTaskById["DEX-05"].adversarial_simulation_result.expected_rejection -match "constant-product") "DEX-05 must record constant-product rejection"
  Assert-True ($atomicTaskById["ID-01"].function_or_flow_covered -match "domain auction") "ID-01 must record identity lifecycle flow"
  Assert-True ($atomicTaskById["ID-02"].adversarial_simulation_result.mutation_inputs -match "zero resolver") "ID-02 must record zero resolver mutation"
  Assert-True ($atomicTaskById["ID-03"].adversarial_simulation_result.attack_attempt -match "domain hijack") "ID-03 must record domain hijack attack"
  Assert-True ($atomicTaskById["ID-04"].invariant_tested -match "NFT representation") "ID-04 must record NFT representation invariant"
  Assert-True ($atomicTaskById["ID-05"].adversarial_simulation_result.expected_rejection -match "invalid targets") "ID-05 must record invalid payment target rejection"
  Assert-True ($atomicTaskById["REP-01"].function_or_flow_covered -match "score updates") "REP-01 must record reputation lifecycle flow"
  Assert-True ($atomicTaskById["REP-02"].adversarial_simulation_result.mutation_inputs -match "score floor") "REP-02 must record score floor mutation"
  Assert-True ($atomicTaskById["REP-03"].adversarial_simulation_result.attack_attempt -match "reputation farming") "REP-03 must record reputation farming attack"
  Assert-True ($atomicTaskById["REP-04"].invariant_tested -match "deterministic replay") "REP-04 must record replay determinism invariant"
  Assert-True ($atomicTaskById["REP-05"].adversarial_simulation_result.expected_rejection -match "required fees") "REP-05 must record fee/deposit/signer bypass rejection"
  Assert-True ($atomicTaskById["EXEC-01"].function_or_flow_covered -match "transaction pipeline order") "EXEC-01 must record execution pipeline flow"
  Assert-True ($atomicTaskById["EXEC-02"].adversarial_simulation_result.mutation_inputs -match "missing route") "EXEC-02 must record missing route mutation"
  Assert-True ($atomicTaskById["EXEC-03"].adversarial_simulation_result.attack_attempt -match "partial rollback") "EXEC-03 must record partial rollback attack"
  Assert-True ($atomicTaskById["EXEC-04"].invariant_tested -match "failed execution") "EXEC-04 must record no-partial-write invariant"
  Assert-True ($atomicTaskById["EXEC-05"].adversarial_simulation_result.expected_rejection -match "routing constraints") "EXEC-05 must record routing constraint rejection"
  Assert-True ($atomicTaskById["VM-01"].function_or_flow_covered -match "AVM deploy") "VM-01 must record AVM lifecycle flow"
  Assert-True ($atomicTaskById["VM-02"].adversarial_simulation_result.mutation_inputs -match "zero gas") "VM-02 must record zero gas mutation"
  Assert-True ($atomicTaskById["VM-03"].adversarial_simulation_result.attack_attempt -match "sandbox escape") "VM-03 must record sandbox escape attack"
  Assert-True ($atomicTaskById["VM-04"].invariant_tested -match "rejected AVM execution") "VM-04 must record rejected execution no-commit invariant"
  Assert-True ($atomicTaskById["VM-05"].adversarial_simulation_result.expected_rejection -match "double-refund") "VM-05 must record double-refund rejection"
  Assert-True ($atomicTaskById["MSG-01"].function_or_flow_covered -match "async send") "MSG-01 must record messaging lifecycle flow"
  Assert-True ($atomicTaskById["MSG-02"].adversarial_simulation_result.mutation_inputs -match "expired message") "MSG-02 must record expired message mutation"
  Assert-True ($atomicTaskById["MSG-03"].adversarial_simulation_result.attack_attempt -match "forged proof") "MSG-03 must record forged proof attack"
  Assert-True ($atomicTaskById["MSG-04"].invariant_tested -match "replay/export/import") "MSG-04 must record replay/export/import invariant"
  Assert-True ($atomicTaskById["MSG-05"].adversarial_simulation_result.expected_rejection -match "double-refund") "MSG-05 must record refund double-spend rejection"
  Assert-True ($atomicTaskById["QUEUE-01"].function_or_flow_covered -match "enqueue") "QUEUE-01 must record queue lifecycle flow"
  Assert-True ($atomicTaskById["QUEUE-02"].adversarial_simulation_result.mutation_inputs -match "duplicate sequence") "QUEUE-02 must record duplicate sequence mutation"
  Assert-True ($atomicTaskById["QUEUE-03"].adversarial_simulation_result.attack_attempt -match "queue flooding") "QUEUE-03 must record queue flooding attack"
  Assert-True ($atomicTaskById["QUEUE-04"].invariant_tested -match "sequence counters") "QUEUE-04 must record sequence counter invariant"
  Assert-True ($atomicTaskById["QUEUE-05"].adversarial_simulation_result.expected_rejection -match "refunded twice") "QUEUE-05 must record double refund rejection"
  Assert-True ($atomicTaskById["EVENTS-01"].function_or_flow_covered -match "deterministic event emission") "EVENTS-01 must record deterministic event flow"
  Assert-True ($atomicTaskById["EVENTS-02"].adversarial_simulation_result.mutation_inputs -match "duplicate event keys") "EVENTS-02 must record duplicate event key mutation"
  Assert-True ($atomicTaskById["EVENTS-03"].adversarial_simulation_result.attack_attempt -match "event spoofing") "EVENTS-03 must record event spoofing attack"
  Assert-True ($atomicTaskById["EVENTS-04"].invariant_tested -match "committed state and receipts") "EVENTS-04 must record committed-state receipt invariant"
  Assert-True ($atomicTaskById["EVENTS-05"].adversarial_simulation_result.expected_rejection -match "authority for balances") "EVENTS-05 must record event authority rejection"
  Assert-True ($atomicTaskById["ACTOR-01"].function_or_flow_covered -match "actor lifecycle") "ACTOR-01 must record actor lifecycle flow"
  Assert-True ($atomicTaskById["ACTOR-02"].adversarial_simulation_result.mutation_inputs -match "max mailbox") "ACTOR-02 must record max mailbox mutation"
  Assert-True ($atomicTaskById["ACTOR-03"].adversarial_simulation_result.attack_attempt -match "mailbox flood") "ACTOR-03 must record mailbox flood attack"
  Assert-True ($atomicTaskById["ACTOR-04"].invariant_tested -match "committed messages") "ACTOR-04 must record committed-message isolation invariant"
  Assert-True ($atomicTaskById["ACTOR-05"].adversarial_simulation_result.expected_rejection -match "actor splitting") "ACTOR-05 must record actor splitting cost rejection"
  Assert-True ($atomicTaskById["SCHED-01"].function_or_flow_covered -match "deterministic ordering") "SCHED-01 must record scheduler lifecycle flow"
  Assert-True ($atomicTaskById["SCHED-02"].adversarial_simulation_result.mutation_inputs -match "duplicate task id") "SCHED-02 must record duplicate task id mutation"
  Assert-True ($atomicTaskById["SCHED-03"].adversarial_simulation_result.attack_attempt -match "nondeterministic tie-break") "SCHED-03 must record nondeterministic tie-break attack"
  Assert-True ($atomicTaskById["SCHED-04"].invariant_tested -match "same tasks and state") "SCHED-04 must record same-input same-plan invariant"
  Assert-True ($atomicTaskById["SCHED-05"].adversarial_simulation_result.expected_rejection -match "fee/reputation caps") "SCHED-05 must record fee/reputation cap rejection"
  Assert-True ($atomicTaskById["STORE-01"].function_or_flow_covered -match "KV writes") "STORE-01 must record storage lifecycle flow"
  Assert-True ($atomicTaskById["STORE-02"].adversarial_simulation_result.mutation_inputs -match "pagination") "STORE-02 must record pagination boundary mutation"
  Assert-True ($atomicTaskById["STORE-03"].adversarial_simulation_result.attack_attempt -match "state root collision") "STORE-03 must record state root collision attack"
  Assert-True ($atomicTaskById["STORE-04"].invariant_tested -match "snapshot root") "STORE-04 must record root determinism invariant"
  Assert-True ($atomicTaskById["STORE-05"].adversarial_simulation_result.expected_rejection -match "storage rent/deposit") "STORE-05 must record rent/deposit bypass rejection"
  Assert-True ($atomicTaskById["MEMO-01"].function_or_flow_covered -match "UTF-8 memo") "MEMO-01 must record memo lifecycle flow"
  Assert-True ($atomicTaskById["MEMO-02"].adversarial_simulation_result.mutation_inputs -match "invalid UTF-8") "MEMO-02 must record invalid UTF-8 mutation"
  Assert-True ($atomicTaskById["MEMO-03"].adversarial_simulation_result.attack_attempt -match "memo spam") "MEMO-03 must record memo spam attack"
  Assert-True ($atomicTaskById["MEMO-04"].invariant_tested -match "immutable after block inclusion") "MEMO-04 must record memo immutability invariant"
  Assert-True ($atomicTaskById["MEMO-05"].adversarial_simulation_result.expected_rejection -match "byte fee") "MEMO-05 must record memo fee bypass rejection"
  Assert-True ($atomicTaskById["INDEX-01"].function_or_flow_covered -match "tx hash") "INDEX-01 must record index lifecycle flow"
  Assert-True ($atomicTaskById["INDEX-02"].adversarial_simulation_result.mutation_inputs -match "pagination") "INDEX-02 must record pagination mutation"
  Assert-True ($atomicTaskById["INDEX-03"].adversarial_simulation_result.attack_attempt -match "index poisoning") "INDEX-03 must record index poisoning attack"
  Assert-True ($atomicTaskById["INDEX-04"].invariant_tested -match "never overrides consensus state") "INDEX-04 must record non-authoritative consensus invariant"
  Assert-True ($atomicTaskById["INDEX-05"].adversarial_simulation_result.expected_rejection -match "route funds") "INDEX-05 must record no fund routing rejection"
  Assert-True ($atomicTaskById["SHARD-01"].function_or_flow_covered -match "LOAD_SCORE") "SHARD-01 must record load/routing lifecycle flow"
  Assert-True ($atomicTaskById["SHARD-02"].adversarial_simulation_result.mutation_inputs -match "oscillating load") "SHARD-02 must record oscillating load mutation"
  Assert-True ($atomicTaskById["SHARD-03"].adversarial_simulation_result.attack_attempt -match "load poisoning") "SHARD-03 must record load poisoning attack"
  Assert-True ($atomicTaskById["SHARD-04"].invariant_tested -match "same tx and state") "SHARD-04 must record same-input routing invariant"
  Assert-True ($atomicTaskById["SHARD-05"].adversarial_simulation_result.expected_rejection -match "deterministic protocol rules") "SHARD-05 must record deterministic routing abuse rejection"

  $enforceFailed = $false
  try {
    & .\scripts\security\aexs-audit.ps1 -OutputDir $OutputDir -EnforceSafe | Out-Null
  } catch {
    $enforceFailed = $true
  }
  Assert-True $enforceFailed "EnforceSafe must fail until executed fuzz/invariant evidence passes"
} finally {
  Pop-Location
}

Write-Host "AEXS audit preflight test passed"
