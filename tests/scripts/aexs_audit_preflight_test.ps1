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

  foreach ($module in @(
      "app",
      "x/fees",
      "x/tokenfactory",
      "x/dex",
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
      "scenario-generator.json",
      "scenario-generator.md",
      "transaction-mutator.json",
      "transaction-mutator.md",
      "AUDIT_RESULT.md",
      "TO_AUDIT.md"
    )) {
    Assert-True (Test-Path -LiteralPath (Join-Path $result.output_dir $name)) "AEXS output missing $name"
  }

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
      "ID-05"
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
  Assert-True ($atomicTaskById["TF-01"].function_or_flow_covered -match "create denom") "TF-01 must record tokenfactory lifecycle flow"
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
