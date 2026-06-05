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

  $coverage = Get-Content -Raw -LiteralPath (Join-Path $result.output_dir "coverage-matrix.json") | ConvertFrom-Json
  Assert-True (@($coverage).Count -ge 24) "coverage matrix must include all required module surfaces"
  Assert-True (@($coverage | Where-Object { $_.task_count -lt 5 }).Count -eq 0) "every module must have at least five tasks"
  Assert-True (@($coverage | Where-Object { $_.atomic_task_records -lt 5 }).Count -eq 0) "every module must have at least five atomic task records"
  Assert-True (@($coverage | Where-Object { @($_.invalid_atomic_tasks).Count -gt 0 }).Count -eq 0) "no module may contain invalid atomic task records"
  Assert-True (@($coverage | Where-Object { $_.planned_coverage_percent -lt 95 }).Count -eq 0) "every module must meet planned coverage threshold"
  Assert-True (@($coverage | Where-Object { $_.safe -eq $true }).Count -eq 0) "no module may be marked safe by preflight alone"

  $atomicTasks = Get-Content -Raw -LiteralPath (Join-Path $result.output_dir "atomic-tasks.json") | ConvertFrom-Json
  Assert-True (@($atomicTasks).Count -eq $result.atomic_task_count) "summary atomic task count must match atomic-tasks.json"
  foreach ($task in $atomicTasks) {
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
