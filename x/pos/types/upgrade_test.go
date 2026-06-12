package types

import (
	"slices"
	"strings"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestLayeredPoSArchitectureMatchesTargetStack(t *testing.T) {
	architecture := DefaultLayeredPosArchitecture()

	require.NoError(t, architecture.Validate())
	require.Equal(t, []PosLayer{
		PosLayerEconomicConsensus,
		PosLayerTaskAssignment,
		PosLayerValidatorExecution,
		PosLayerStakingCapital,
		PosLayerBaseCometBFT,
	}, DefaultPosLayerOrder())
	require.Equal(t, ComputeLayeredPosArchitectureRoot(architecture.Layers), architecture.Root)

	specs := posLayerSpecsByLayer(architecture)
	require.Equal(t, []string{
		"finality",
		"proposal and vote protocol",
		"validator public key set",
		"consensus safety and liveness",
	}, specs[PosLayerBaseCometBFT].Responsibilities)
	require.Equal(t, []string{
		"validators",
		"delegators",
		"bonded stake",
		"unbonding",
		"redelegation",
		"capital risk preferences",
		"commission and delegation market metadata",
	}, specs[PosLayerStakingCapital].Responsibilities)
	require.Equal(t, []string{
		"block production",
		"state transition verification",
		"cross-domain proof verification",
		"signature production",
		"fault rejection",
	}, specs[PosLayerValidatorExecution].Responsibilities)
	require.Equal(t, []string{
		"workload grouping",
		"shard validator groups",
		"zone validator groups",
		"evidence verification subsets",
		"collator and verifier assignments",
	}, specs[PosLayerTaskAssignment].Responsibilities)
	require.Equal(t, []string{
		"validator scoring",
		"performance incentives",
		"stake saturation",
		"role-specific reward weights",
		"slashing severity",
		"reporter incentives",
		"treasury, burn, and stabilization routing",
	}, specs[PosLayerEconomicConsensus].Responsibilities)
}

func TestLayeredPoSArchitectureRejectsReorderedOrUpwardDependencies(t *testing.T) {
	architecture := DefaultLayeredPosArchitecture()
	slices.Reverse(architecture.Layers)
	architecture.Root = ComputeLayeredPosArchitectureRoot(architecture.Layers)
	require.ErrorContains(t, architecture.Validate(), "pos layer 0")

	architecture = DefaultLayeredPosArchitecture()
	architecture.Layers[3].DependsOn = append(architecture.Layers[3].DependsOn, PosLayerEconomicConsensus)
	architecture.Root = ComputeLayeredPosArchitectureRoot(architecture.Layers)
	require.ErrorContains(t, architecture.Validate(), "lower layers")
}

func TestCosmosSDKCompatibilityManifestExtendsBaselineModules(t *testing.T) {
	manifest := DefaultCosmosSDKCompatibilityManifest()
	require.NoError(t, manifest.Validate())
	require.Equal(t, ComputeCosmosSDKCompatibilityRoot(manifest), manifest.Root)
	require.Equal(t, []string{"epoch", "validator_economy", "taskgroups", "evidence", "performance"}, RequiredPoSModuleNames(manifest))
	require.Equal(t, []string{"delegation_market", "collators", "fishermen", "security_metrics"}, OptionalPoSModuleNames(manifest))

	extensions := compatibilityExtensionsByName(manifest)
	for _, moduleName := range []string{"staking", "slashing", "distribution", "mint"} {
		extension, found := extensions[moduleName]
		require.True(t, found, "missing sdk extension %s", moduleName)
		require.Equal(t, CosmosSDKExtensionModeExtend, extension.ExtensionMode)
		require.NotEmpty(t, extension.PreservedInterfaces)
	}
	names := compatibilityMiddlewareNames(manifest)
	require.Contains(t, names, "validator_scoring")
	require.Contains(t, names, "epoch_management")
	require.Contains(t, names, "task_assignment")
	require.Contains(t, names, "performance_accounting")
	require.Contains(t, names, "evidence_slashing")
}

func TestCosmosSDKCompatibilityManifestRejectsReplacementOrMissingModules(t *testing.T) {
	manifest := DefaultCosmosSDKCompatibilityManifest()
	replacement := manifest
	replacement.Extensions = append([]CosmosSDKModuleExtension{}, manifest.Extensions...)
	replacement.Extensions[0].ExtensionMode = CosmosSDKExtensionModeReplace
	replacement.Root = ComputeCosmosSDKCompatibilityRoot(replacement)
	require.ErrorContains(t, replacement.Validate(), "extended, not replaced")

	missing := manifest
	missing.Modules = append([]PosModuleRequirement{}, manifest.Modules...)
	for i, module := range missing.Modules {
		if module.ModuleName == "performance" {
			missing.Modules = append(missing.Modules[:i], missing.Modules[i+1:]...)
			break
		}
	}
	missing.Root = ComputeCosmosSDKCompatibilityRoot(missing)
	require.ErrorContains(t, missing.Validate(), "required pos module performance")

	unknown := manifest
	unknown.Middleware = append([]PosCompatibilityMiddleware{}, manifest.Middleware...)
	unknown.Middleware[0].WritesModules = append([]string{}, unknown.Middleware[0].WritesModules...)
	unknown.Middleware[0].WritesModules = append(unknown.Middleware[0].WritesModules, "unknown_module")
	unknown.Root = ComputeCosmosSDKCompatibilityRoot(unknown)
	require.ErrorContains(t, unknown.Validate(), "unknown module")
}

func TestPoSModuleBoundariesMatchRequiredCompatibilityModules(t *testing.T) {
	compatibility := DefaultCosmosSDKCompatibilityManifest()
	boundaries := DefaultPoSModuleBoundaryManifest()
	require.NoError(t, boundaries.Validate(compatibility))
	require.Equal(t, ComputePoSModuleBoundaryRoot(boundaries), boundaries.Root)

	require.Equal(t, RequiredPoSModuleNames(compatibility), posBoundaryModuleNames(boundaries))
	epoch, found := PoSModuleBoundaryByName(boundaries, "epoch")
	require.True(t, found)
	require.Equal(t, []string{"epoch lifecycle", "phase transitions", "epoch seed", "epoch queries"}, epoch.Owns)
	validatorEconomy, found := PoSModuleBoundaryByName(boundaries, "validator_economy")
	require.True(t, found)
	require.Equal(t, []string{"validator score", "effective stake", "stake saturation", "election ranking", "role eligibility"}, validatorEconomy.Owns)
	taskgroups, found := PoSModuleBoundaryByName(boundaries, "taskgroups")
	require.True(t, found)
	require.Equal(t, []string{"workload registry", "task group assignment", "proposer rotation", "verification groups"}, taskgroups.Owns)
	evidence, found := PoSModuleBoundaryByName(boundaries, "evidence")
	require.True(t, found)
	require.Equal(t, []string{"structured evidence records", "evidence deposits", "verification group decisions", "reporter rewards"}, evidence.Owns)
	require.Contains(t, evidence.WritesModules, "slashing")
	require.Contains(t, evidence.WritesModules, "distribution")
	performance, found := PoSModuleBoundaryByName(boundaries, "performance")
	require.True(t, found)
	require.Equal(t, []string{"uptime", "latency", "correctness", "task completion", "reward multipliers"}, performance.Owns)
	require.Contains(t, performance.QueryEndpoints, "QueryRewardMultiplier")
}

func TestPoSModuleBoundariesRejectOverlapMissingAndUnknownReferences(t *testing.T) {
	compatibility := DefaultCosmosSDKCompatibilityManifest()
	boundaries := DefaultPoSModuleBoundaryManifest()

	overlap := boundaries
	overlap.Boundaries = append([]PosModuleBoundary{}, boundaries.Boundaries...)
	overlap.Boundaries[1].Owns = append([]string{}, overlap.Boundaries[1].Owns...)
	overlap.Boundaries[1].Owns = append(overlap.Boundaries[1].Owns, "epoch seed")
	overlap.Root = ComputePoSModuleBoundaryRoot(overlap)
	require.ErrorContains(t, overlap.Validate(compatibility), "overlaps")

	missing := boundaries
	missing.Boundaries = append([]PosModuleBoundary{}, boundaries.Boundaries...)
	missing.Boundaries = missing.Boundaries[:len(missing.Boundaries)-1]
	missing.Root = ComputePoSModuleBoundaryRoot(missing)
	require.ErrorContains(t, missing.Validate(compatibility), "performance")

	unknown := boundaries
	unknown.Boundaries = append([]PosModuleBoundary{}, boundaries.Boundaries...)
	unknown.Boundaries[0].ReadsModules = append([]string{}, unknown.Boundaries[0].ReadsModules...)
	unknown.Boundaries[0].ReadsModules = append(unknown.Boundaries[0].ReadsModules, "unknown_module")
	unknown.Root = ComputePoSModuleBoundaryRoot(unknown)
	require.ErrorContains(t, unknown.Validate(compatibility), "unknown module")
}

func TestKeeperIntegrationManifestCoversSDKKeepersHooksAndExportImport(t *testing.T) {
	compatibility := DefaultCosmosSDKCompatibilityManifest()
	boundaries := DefaultPoSModuleBoundaryManifest()
	manifest := DefaultKeeperIntegrationManifest()
	require.NoError(t, manifest.Validate(compatibility, boundaries))
	require.Equal(t, ComputeKeeperIntegrationRoot(manifest), manifest.Root)

	keepers := keeperInterfacesByName(manifest)
	for _, keeperName := range []string{"staking", "slashing", "distribution", "mint", "bank", "gov"} {
		require.Contains(t, keepers, keeperName)
	}
	require.Equal(t, "validator and delegation state", keepers["staking"].IntegrationPoint)
	require.Equal(t, "jail tombstone and slash execution", keepers["slashing"].IntegrationPoint)
	require.Equal(t, "reward allocation", keepers["distribution"].IntegrationPoint)
	require.Equal(t, "epoch reward budget", keepers["mint"].IntegrationPoint)
	require.Equal(t, "deposits reporter rewards and penalty routing", keepers["bank"].IntegrationPoint)
	require.Equal(t, "parameter updates", keepers["gov"].IntegrationPoint)

	require.Contains(t, keeperHookNames(manifest.StakingLifecycleHooks), "AfterDelegationModified")
	require.Contains(t, keeperHookNames(manifest.StakingLifecycleHooks), "BeforeDelegationRemoved")
	slashingHooks := keeperHookNames(manifest.SlashingHooks)
	require.Contains(t, slashingHooks, "AfterValidatorSlashed")
	require.Contains(t, slashingHooks, "AfterValidatorJailed")
	require.Contains(t, slashingHooks, "AfterValidatorTombstoned")
	require.Equal(t, "reward_multiplier_bps", manifest.RewardIntegrations[0].MultiplierField)
	require.Equal(t, []string{"uptime", "latency", "correctness", "task completion"}, manifest.RewardIntegrations[0].RewardInputs)
	require.Equal(t, RequiredPoSModuleNames(compatibility), migrationModuleNames(manifest.MigrationHandlers))
	require.Equal(t, RequiredPoSModuleNames(compatibility), exportImportModuleNames(manifest.ExportImport))
}

func TestKeeperIntegrationManifestRejectsUnsafeHooksMigrationsAndRewards(t *testing.T) {
	compatibility := DefaultCosmosSDKCompatibilityManifest()
	boundaries := DefaultPoSModuleBoundaryManifest()
	manifest := DefaultKeeperIntegrationManifest()

	unsafeHook := manifest
	unsafeHook.StakingLifecycleHooks = append([]KeeperHookSpec{}, manifest.StakingLifecycleHooks...)
	unsafeHook.StakingLifecycleHooks[0].PreservesBaseState = false
	unsafeHook.Root = ComputeKeeperIntegrationRoot(unsafeHook)
	require.ErrorContains(t, unsafeHook.Validate(compatibility, boundaries), "preserve base sdk state")

	badReward := manifest
	badReward.RewardIntegrations = append([]RewardMultiplierIntegration{}, manifest.RewardIntegrations...)
	badReward.RewardIntegrations[0].DistributionKeeper = "bank"
	badReward.Root = ComputeKeeperIntegrationRoot(badReward)
	require.ErrorContains(t, badReward.Validate(compatibility, boundaries), "performance to distribution and mint")

	badMigration := manifest
	badMigration.MigrationHandlers = append([]MigrationHandlerSpec{}, manifest.MigrationHandlers...)
	badMigration.MigrationHandlers[1].PreservesExistingStakingState = false
	badMigration.Root = ComputeKeeperIntegrationRoot(badMigration)
	require.ErrorContains(t, badMigration.Validate(compatibility, boundaries), "preserve existing staking state")

	badExport := manifest
	badExport.ExportImport = append([]ModuleExportImportSpec{}, manifest.ExportImport...)
	badExport.ExportImport[2].DeterministicEncoding = false
	badExport.Root = ComputeKeeperIntegrationRoot(badExport)
	require.ErrorContains(t, badExport.Validate(compatibility, boundaries), "deterministic")
}

func TestStateModelManifestDefinesAllRequiredKeyPrefixes(t *testing.T) {
	manifest := DefaultStateModelManifest()
	require.NoError(t, manifest.Validate())
	require.Equal(t, ComputeStateModelRoot(manifest), manifest.Root)
	require.Equal(t, []string{
		"epoch/current",
		"epoch/records/{epoch_id}",
		"epoch/phase/{epoch_id}",
		"epoch/seed/{epoch_id}",
		"valecon/scores/{epoch_id}/{validator}",
		"valecon/effective_stake/{epoch_id}/{validator}",
		"valecon/saturation/{epoch_id}/{validator}",
		"valecon/roles/{epoch_id}/{validator}/{role}",
		"taskgroups/groups/{epoch_id}/{task_group_id}",
		"taskgroups/workloads/{workload_id}",
		"taskgroups/assignments/{epoch_id}/{validator}/{task_group_id}",
		"taskgroups/proposer/{epoch_id}/{slot}/{task_group_id}",
		"evidence/records/{evidence_id}",
		"evidence/by_accused/{validator}/{evidence_id}",
		"evidence/by_reporter/{reporter}/{evidence_id}",
		"evidence/verification_groups/{evidence_id}",
		"evidence/deposits/{evidence_id}",
		"performance/records/{epoch_id}/{operator}/{role}",
		"performance/uptime/{epoch_id}/{validator}",
		"performance/correctness/{epoch_id}/{validator}",
		"performance/tasks/{epoch_id}/{validator}",
		"risk/unbonding/{delegator}/{validator}/{creation_height}",
		"risk/redelegation/{delegator}/{src_validator}/{dst_validator}/{epoch_id}",
		"risk/exposure/{epoch_id}/{validator}/{delegator}",
	}, stateKeyTemplates(manifest))
}

func TestStateKeyBuildersProduceDeterministicPaths(t *testing.T) {
	require.Equal(t, "epoch/current", EpochCurrentKey())
	require.Equal(t, "epoch/records/42", EpochRecordKey(42))
	require.Equal(t, "epoch/phase/42", EpochPhaseKey(42))
	require.Equal(t, "epoch/seed/42", EpochSeedKey(42))
	require.Equal(t, "valecon/scores/42/val-a", mustStateKey(t, func() (string, error) { return ValidatorScoreKey(42, "val-a") }))
	require.Equal(t, "valecon/effective_stake/42/val-a", mustStateKey(t, func() (string, error) { return ValidatorEffectiveStakeKey(42, "val-a") }))
	require.Equal(t, "valecon/saturation/42/val-a", mustStateKey(t, func() (string, error) { return ValidatorSaturationKey(42, "val-a") }))
	require.Equal(t, "valecon/roles/42/val-a/verifier", mustStateKey(t, func() (string, error) { return ValidatorRoleKey(42, "val-a", ValidatorRoleVerifier) }))
	require.Equal(t, "taskgroups/groups/42/tg-a", mustStateKey(t, func() (string, error) { return TaskGroupKey(42, "tg-a") }))
	require.Equal(t, "taskgroups/workloads/workload-a", mustStateKey(t, func() (string, error) { return WorkloadKey("workload-a") }))
	require.Equal(t, "taskgroups/assignments/42/val-a/tg-a", mustStateKey(t, func() (string, error) { return TaskAssignmentKey(42, "val-a", "tg-a") }))
	require.Equal(t, "taskgroups/proposer/42/7/tg-a", mustStateKey(t, func() (string, error) { return ProposerKey(42, 7, "tg-a") }))
	require.Equal(t, "evidence/records/evidence-a", mustStateKey(t, func() (string, error) { return EvidenceRecordKey("evidence-a") }))
	require.Equal(t, "evidence/by_accused/val-a/evidence-a", mustStateKey(t, func() (string, error) { return EvidenceByAccusedKey("val-a", "evidence-a") }))
	require.Equal(t, "evidence/by_reporter/reporter-a/evidence-a", mustStateKey(t, func() (string, error) { return EvidenceByReporterKey("reporter-a", "evidence-a") }))
	require.Equal(t, "evidence/verification_groups/evidence-a", mustStateKey(t, func() (string, error) { return EvidenceVerificationGroupKey("evidence-a") }))
	require.Equal(t, "evidence/deposits/evidence-a", mustStateKey(t, func() (string, error) { return EvidenceDepositKey("evidence-a") }))
	require.Equal(t, "performance/records/42/operator-a/verifier", mustStateKey(t, func() (string, error) { return PerformanceRecordKey(42, "operator-a", ValidatorRoleVerifier) }))
	require.Equal(t, "performance/uptime/42/val-a", mustStateKey(t, func() (string, error) { return PerformanceUptimeKey(42, "val-a") }))
	require.Equal(t, "performance/correctness/42/val-a", mustStateKey(t, func() (string, error) { return PerformanceCorrectnessKey(42, "val-a") }))
	require.Equal(t, "performance/tasks/42/val-a", mustStateKey(t, func() (string, error) { return PerformanceTasksKey(42, "val-a") }))
	require.Equal(t, "risk/unbonding/delegator-a/val-a/100", mustStateKey(t, func() (string, error) { return RiskUnbondingKey("delegator-a", "val-a", 100) }))
	require.Equal(t, "risk/redelegation/delegator-a/val-a/val-b/42", mustStateKey(t, func() (string, error) { return RiskRedelegationKey("delegator-a", "val-a", "val-b", 42) }))
	require.Equal(t, "risk/exposure/42/val-a/delegator-a", mustStateKey(t, func() (string, error) { return RiskExposureKey(42, "val-a", "delegator-a") }))
}

func TestStateModelRejectsAmbiguousPrefixesAndPathInjection(t *testing.T) {
	manifest := DefaultStateModelManifest()
	duplicate := manifest
	duplicate.Keys = append([]StateKeySpec{}, manifest.Keys...)
	duplicate.Keys[1].Template = duplicate.Keys[0].Template
	duplicate.Keys[1].Components = nil
	duplicate.Root = ComputeStateModelRoot(duplicate)
	require.ErrorContains(t, duplicate.Validate(), "duplicate state key template")

	tampered := manifest
	tampered.Keys = append([]StateKeySpec{}, manifest.Keys...)
	tampered.Keys[0].Name = "current2"
	require.ErrorContains(t, tampered.Validate(), "root mismatch")

	_, err := ValidatorScoreKey(42, "val/a")
	require.ErrorContains(t, err, "path separator")
	_, err = EvidenceRecordKey(" evidence-a")
	require.ErrorContains(t, err, "surrounding whitespace")
}

func TestEpochDefinitionDefaultsMatchLifecycleParameters(t *testing.T) {
	params := DefaultParams()

	require.NoError(t, params.Validate())
	require.Equal(t, uint64(43_200), params.EpochDurationSeconds)
	require.GreaterOrEqual(t, params.EpochDurationSeconds, MinEpochDurationSeconds)
	require.LessOrEqual(t, params.EpochDurationSeconds, MaxEpochDurationSeconds)
	require.Equal(t, EpochSeedSourcePreviousSeedValidatorSet, params.EffectiveEpochSeedSource())
	require.Equal(t, DefaultMaxValidatorSetChangeRateBps, params.MaxValidatorSetChangeRateBps)

	durations := params.EffectivePhaseDurations()
	require.Equal(t, params.EpochDurationSeconds, durations.TotalSeconds())
	require.Positive(t, durations.DelegationSeconds)
	require.Positive(t, durations.ElectionSeconds)
	require.Positive(t, durations.AssignmentSeconds)
	require.Positive(t, durations.ActiveValidationSeconds)
	require.Positive(t, durations.SettlementSeconds)

	params.EpochSeedSource = "unknown"
	require.ErrorContains(t, params.Validate(), "seed source")

	params = DefaultParams()
	params.MaxValidatorSetChangeRateBps = 0
	require.ErrorContains(t, params.Validate(), "validator set change rate")
}

func TestEpochLifecycleTransitionsFollowTargetOrder(t *testing.T) {
	lifecycle := DefaultEpochLifecycle()
	require.NoError(t, ValidateEpochLifecycle(lifecycle))
	require.Equal(t, []EpochLifecycleStep{
		{Phase: EpochPhaseDelegation, Name: "delegation phase", DurationKey: "delegation_phase_duration"},
		{Phase: EpochPhaseElection, Name: "validator election", DurationKey: "election_phase_duration"},
		{Phase: EpochPhaseAssignment, Name: "task group assignment", DurationKey: "assignment_phase_duration"},
		{Phase: EpochPhaseActive, Name: "active validation", DurationKey: "active_validation_duration"},
		{Phase: EpochPhaseSettlement, Name: "settlement + reward + slash finality", DurationKey: "settlement_phase_duration"},
	}, lifecycle)

	next, closes, err := NextEpochPhase(EpochPhaseDelegation)
	require.NoError(t, err)
	require.False(t, closes)
	require.Equal(t, EpochPhaseElection, next)
	require.NoError(t, ValidateEpochPhaseTransition(EpochPhaseElection, EpochPhaseAssignment))
	require.NoError(t, ValidateEpochPhaseTransition(EpochPhaseAssignment, EpochPhaseActive))
	require.NoError(t, ValidateEpochPhaseTransition(EpochPhaseActive, EpochPhaseSettlement))

	next, closes, err = NextEpochPhase(EpochPhaseSettlement)
	require.NoError(t, err)
	require.True(t, closes)
	require.Equal(t, EpochPhaseClosed, next)
	require.ErrorContains(t, ValidateEpochPhaseTransition(EpochPhaseDelegation, EpochPhaseActive), "invalid epoch phase transition")
}

func TestEpochLifecyclePhasesAndSeedAreDeterministic(t *testing.T) {
	params := DefaultParams()
	params.MinStakeNaet = sdkmath.NewInt(100)
	validators := scoredCandidates(t, params, makeCandidates(3, 1_000))
	durations := params.EffectivePhaseDurations()

	phase, err := EpochPhaseAt(params, 1_000, 1_000)
	require.NoError(t, err)
	require.Equal(t, EpochPhaseDelegation, phase)

	phase, err = EpochPhaseAt(params, 1_000, 1_000+durations.DelegationSeconds)
	require.NoError(t, err)
	require.Equal(t, EpochPhaseElection, phase)

	phase, err = EpochPhaseAt(params, 1_000, 1_000+params.EpochDurationSeconds)
	require.NoError(t, err)
	require.Equal(t, EpochPhaseClosed, phase)

	left, err := NewEpochRecord(params, 7, 100, 199, EpochPhaseAssignment, "", validators)
	require.NoError(t, err)
	right, err := NewEpochRecord(params, 7, 100, 199, EpochPhaseAssignment, "", validators)
	require.NoError(t, err)
	require.Equal(t, left.Seed, right.Seed)
	require.Equal(t, left.ValidatorSetHash, right.ValidatorSetHash)

	next, err := NewEpochRecord(params, 8, 200, 299, EpochPhaseDelegation, left.Seed, validators)
	require.NoError(t, err)
	require.NotEqual(t, left.Seed, next.Seed)

	sourceParams := params
	sourceParams.EpochSeedSource = EpochSeedSourceCometBFTBlockID
	sourceRecord, err := NewEpochRecord(sourceParams, 7, 100, 199, EpochPhaseAssignment, "", validators)
	require.NoError(t, err)
	require.NotEqual(t, left.Seed, sourceRecord.Seed)

	_, err = CloseEpochRecord(left, PosEmptyRootHash, PosEmptyRootHash, PosEmptyRootHash)
	require.ErrorContains(t, err, "settlement phase")

	settlementRecord, err := NewEpochRecord(params, 7, 100, 199, EpochPhaseSettlement, "", validators)
	require.NoError(t, err)
	closed, err := CloseEpochRecord(settlementRecord, PosEmptyRootHash, PosEmptyRootHash, PosEmptyRootHash)
	require.NoError(t, err)
	require.Equal(t, EpochPhaseClosed, closed.Phase)
	require.Equal(t, SettlementStatusFinalized, closed.SettlementStatus)
}

func TestEpochRecordStateContractMatchesDesignFieldsAndPhaseValues(t *testing.T) {
	require.Equal(t, []string{
		"epoch_id",
		"start_height",
		"end_height",
		"phase",
		"seed",
		"validator_set_hash",
		"task_group_root",
		"performance_root",
		"reward_root",
		"slash_root",
		"settlement_status",
	}, EpochRecordFieldNames())
	require.Equal(t, []EpochPhase{
		EpochPhaseDelegation,
		EpochPhaseElection,
		EpochPhaseAssignment,
		EpochPhaseActive,
		EpochPhaseSettlement,
		EpochPhaseClosed,
	}, EpochPhaseValues())

	params := DefaultParams()
	params.MinStakeNaet = sdkmath.NewInt(100)
	record, err := NewEpochRecord(params, 2, 10, 20, EpochPhaseDelegation, "", scoredCandidates(t, params, makeCandidates(3, 1_000)))
	require.NoError(t, err)
	require.NoError(t, record.Validate())

	record.Phase = "invalid"
	require.ErrorContains(t, record.Validate(), "unsupported epoch phase")
}

func TestEpochRulesEnforceValidatorSetChangeBoundaries(t *testing.T) {
	params := DefaultParams()
	params.MinStakeNaet = sdkmath.NewInt(100)
	previousSettlement, err := NewEpochRecord(params, 5, 100, 199, EpochPhaseSettlement, "", scoredCandidates(t, params, makeCandidates(3, 1_000)))
	require.NoError(t, err)
	previous, err := CloseEpochRecord(previousSettlement, PosEmptyRootHash, PosEmptyRootHash, PosEmptyRootHash)
	require.NoError(t, err)
	next, err := NewEpochRecord(params, 6, 200, 299, EpochPhaseDelegation, previous.Seed, scoredCandidates(t, params, makeCandidates(3, 1_000)))
	require.NoError(t, err)

	require.NoError(t, ValidateConsecutiveEpochs(previous, next))
	require.NoError(t, ValidateValidatorSetChangeActivation(next, next.StartHeight))
	require.ErrorContains(t, ValidateValidatorSetChangeActivation(next, next.StartHeight+1), "epoch boundary")

	next.StartHeight = 201
	require.ErrorContains(t, ValidateConsecutiveEpochs(previous, next), "previous end height plus one")
}

func TestEpochRulesEnforceDelegationDelayAndEvidenceWindow(t *testing.T) {
	params := DefaultParams()
	params.DelegationActivationEpochs = 2
	params.EvidenceWindowEpochs = 4

	effectiveEpoch, err := DelegationEffectiveElectionEpoch(params, 10)
	require.NoError(t, err)
	require.Equal(t, uint64(12), effectiveEpoch)

	active, err := DelegationAffectsElection(params, 10, 11)
	require.NoError(t, err)
	require.False(t, active)
	active, err = DelegationAffectsElection(params, 10, 12)
	require.NoError(t, err)
	require.True(t, active)

	within, err := EvidenceWithinSlashableWindow(params, 10, 14)
	require.NoError(t, err)
	require.True(t, within)
	within, err = EvidenceWithinSlashableWindow(params, 10, 15)
	require.NoError(t, err)
	require.False(t, within)
	_, err = EvidenceWithinSlashableWindow(params, 10, 9)
	require.ErrorContains(t, err, "before evidence epoch")
}

func TestUnbondingRiskWindowExtendsBeyondUnbonding(t *testing.T) {
	params := DefaultParams()
	params.EpochDurationSeconds = MinEpochDurationSeconds
	params.PhaseDurations = DefaultEpochPhaseDurations(params.EpochDurationSeconds)
	params.UnbondingSeconds = MinUnbondingSeconds
	params.EvidenceWindowEpochs = 3
	window, err := UnbondingRiskWindowForParams(params)
	require.NoError(t, err)
	require.Equal(t, uint64(14), window.UnbondingEpochs)
	require.Equal(t, uint64(3), window.SlashableWindowEpochs)
	require.Equal(t, uint64(17), window.TotalRiskEpochs)

	record, err := BeginUnbondingRisk(params, "delegator-a", "val-a", sdkmath.NewInt(1_000), 10)
	require.NoError(t, err)
	require.Equal(t, uint64(24), record.ExitEpoch)
	require.Equal(t, uint64(27), record.SlashableUntilEpoch)
	require.Equal(t, ComputeUnbondingRiskHistoryKey(record), record.RiskHistoryKey)

	exposure, err := PendingUnbondingSlashExposure(PendingUnbondingSlashExposureInput{
		Record:		record,
		FaultEpoch:	23,
		EvidenceEpoch:	27,
	})
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(1_000), exposure)

	exposure, err = PendingUnbondingSlashExposure(PendingUnbondingSlashExposureInput{
		Record:		record,
		FaultEpoch:	24,
		EvidenceEpoch:	25,
	})
	require.NoError(t, err)
	require.True(t, exposure.IsZero())

	exposure, err = PendingUnbondingSlashExposure(PendingUnbondingSlashExposureInput{
		Record:		record,
		FaultEpoch:	23,
		EvidenceEpoch:	28,
	})
	require.NoError(t, err)
	require.True(t, exposure.IsZero())
}

func TestRedelegationRetainsSourceRiskAndSelfBondChangesAreDelayed(t *testing.T) {
	params := DefaultParams()
	params.DelegationActivationEpochs = 2
	params.EvidenceWindowEpochs = 4
	redelegation, err := CreateRedelegationRiskRecord(params, "delegator-a", "val-source", "val-dest", sdkmath.NewInt(500), 20)
	require.NoError(t, err)
	require.Equal(t, uint64(22), redelegation.ActivationEpoch)
	require.Greater(t, redelegation.SourceSlashableUntilEpoch, redelegation.ActivationEpoch)
	require.Equal(t, ComputeRedelegationRiskHistoryKey(redelegation), redelegation.RiskHistoryKey)

	tampered := redelegation
	tampered.DestinationValidatorID = "val-other"
	require.ErrorContains(t, tampered.Validate(), "risk history key")

	_, err = CreateRedelegationRiskRecord(params, "delegator-a", "val-source", "val-source", sdkmath.NewInt(500), 20)
	require.ErrorContains(t, err, "destination")

	selfBond, err := PlanSelfBondChange(params, "val-source", sdkmath.NewInt(1_000), sdkmath.NewInt(2_000), 20)
	require.NoError(t, err)
	require.Equal(t, uint64(22), selfBond.ActivationEpoch)
	require.NoError(t, selfBond.Validate())

	_, err = PlanSelfBondChange(params, "val-source", sdkmath.NewInt(1_000), sdkmath.NewInt(-1), 20)
	require.ErrorContains(t, err, "self bond")
}

func TestRiskWindowRecordQueriesSlashExposureForDelayedEvidence(t *testing.T) {
	params := DefaultParams()
	params.EpochDurationSeconds = MinEpochDurationSeconds
	params.PhaseDurations = DefaultEpochPhaseDurations(params.EpochDurationSeconds)
	params.UnbondingSeconds = MinUnbondingSeconds
	params.EvidenceWindowEpochs = 3
	unbonding, err := BeginUnbondingRisk(params, "delegator-a", "val-a", sdkmath.NewInt(1_000), 10)
	require.NoError(t, err)
	active, err := RiskWindowFromUnbonding(unbonding, 10)
	require.NoError(t, err)
	require.Equal(t, RiskWindowStatusActive, active.Status)
	require.Equal(t, "delegator-a", active.StakeOwner)
	require.Equal(t, "val-a", active.ValidatorAddress)
	require.Equal(t, uint64(10), active.StartEpoch)
	require.Equal(t, uint64(24), active.EndEpoch)
	require.Equal(t, uint64(27), active.SlashableUntilEpoch)
	require.Equal(t, ComputeRiskWindowRoot(active), active.RiskHistoryRoot)

	exited, err := RiskWindowFromUnbonding(unbonding, 25)
	require.NoError(t, err)
	require.Equal(t, RiskWindowStatusExited, exited.Status)
	expired, err := RiskWindowFromUnbonding(unbonding, 28)
	require.NoError(t, err)
	require.Equal(t, RiskWindowStatusExpired, expired.Status)

	result, err := QuerySlashExposure([]RiskWindowRecord{exited}, SlashExposureQuery{
		StakeOwner:		"delegator-a",
		ValidatorAddress:	"val-a",
		FaultEpoch:		23,
		EvidenceEpoch:		27,
	})
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(1_000), result.ExposureNaet)
	require.Len(t, result.MatchingWindows, 1)

	result, err = QuerySlashExposure([]RiskWindowRecord{exited}, SlashExposureQuery{
		StakeOwner:		"delegator-a",
		ValidatorAddress:	"val-a",
		FaultEpoch:		23,
		EvidenceEpoch:		28,
	})
	require.NoError(t, err)
	require.True(t, result.ExposureNaet.IsZero())

	tampered := exited
	tampered.AmountNaet = sdkmath.NewInt(2_000)
	require.ErrorContains(t, tampered.Validate(), "history root")
}

func TestRiskWindowRecordKeepsRedelegationSourceExposure(t *testing.T) {
	params := DefaultParams()
	params.DelegationActivationEpochs = 2
	params.EvidenceWindowEpochs = 4
	redelegation, err := CreateRedelegationRiskRecord(params, "delegator-a", "val-source", "val-dest", sdkmath.NewInt(500), 20)
	require.NoError(t, err)
	window, err := RiskWindowFromRedelegation(redelegation, 22)
	require.NoError(t, err)
	require.Equal(t, "val-source", window.ValidatorAddress)
	require.Equal(t, RiskWindowStatusExited, window.Status)

	result, err := QuerySlashExposure([]RiskWindowRecord{window}, SlashExposureQuery{
		StakeOwner:		"delegator-a",
		ValidatorAddress:	"val-source",
		FaultEpoch:		21,
		EvidenceEpoch:		24,
	})
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(500), result.ExposureNaet)

	result, err = QuerySlashExposure([]RiskWindowRecord{window}, SlashExposureQuery{
		StakeOwner:		"delegator-a",
		ValidatorAddress:	"val-dest",
		FaultEpoch:		21,
		EvidenceEpoch:		24,
	})
	require.NoError(t, err)
	require.True(t, result.ExposureNaet.IsZero())
}

func TestEconomicSecurityMetricsComputeFormulaAndRequiredMetrics(t *testing.T) {
	params := DefaultParams()
	params.MinStakeNaet = sdkmath.NewInt(100)
	params.StakeSaturationThresholdNaet = sdkmath.NewInt(1_000)
	params.StakeSaturationCapFactorBps = BasisPoints
	validators := scoredCandidates(t, params, []Candidate{
		candidate("val-a", 2_000, 0),
		candidate("val-b", 1_000, 0),
		candidate("val-c", 500, 0),
	})
	windows := []RiskWindowRecord{
		riskWindow("delegator-a", "val-a", 600, RiskWindowStatusActive),
		riskWindow("delegator-b", "val-b", 400, RiskWindowStatusExited),
		riskWindow("delegator-c", "val-c", 300, RiskWindowStatusExpired),
	}

	metrics, err := ComputeEconomicSecurityMetrics(EconomicSecurityInput{
		Validators:			validators,
		RiskWindows:			windows,
		TopN:				1,
		ParticipatingValidators:	2,
		EligibleValidators:		3,
		AcceptedSlashEvents:		3,
		DetectedFaultEvents:		4,
		AcceptedEvidence:		5,
		SubmittedEvidence:		8,
		CompletedTasks:			9,
		ExpectedTasks:			12,
	})
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(3_500), metrics.TotalBondedStakeNaet)
	require.Equal(t, sdkmath.NewInt(2_500), metrics.EffectiveStakeNaet)
	require.Equal(t, sdkmath.NewInt(1_000), metrics.TotalStakeAtRiskNaet)
	require.Equal(t, uint32(2_857), metrics.StakeSaturationRatioBps)
	require.Equal(t, uint32(4_000), metrics.TopNVotingPowerConcentrationBps)
	require.Equal(t, uint32(6_666), metrics.ParticipationRateBps)
	require.Equal(t, uint32(7_500), metrics.SlashingEfficiencyBps)
	require.Equal(t, uint32(6_250), metrics.EvidenceAcceptanceRateBps)
	require.Equal(t, sdkmath.NewInt(833), metrics.AverageValidatorScore)
	require.Equal(t, uint32(7_500), metrics.TaskCompletionRateBps)
	require.Equal(t, sdkmath.NewInt(499), metrics.SecurityNaet)
	require.Equal(t, []DelegationRiskBucket{
		{ValidatorAddress: "val-a", ExposureNaet: sdkmath.NewInt(600), RiskWindowCount: 1},
		{ValidatorAddress: "val-b", ExposureNaet: sdkmath.NewInt(400), RiskWindowCount: 1},
	}, metrics.DelegationRiskDistribution)

	override, err := ComputeEconomicSecurityMetrics(EconomicSecurityInput{
		Validators:			validators,
		RiskWindows:			windows,
		StakeAtRiskNaet:		sdkmath.NewInt(2_000),
		TopN:				2,
		ParticipatingValidators:	3,
		EligibleValidators:		3,
		AcceptedSlashEvents:		1,
		DetectedFaultEvents:		2,
		CompletedTasks:			1,
		ExpectedTasks:			1,
	})
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(2_000), override.TotalStakeAtRiskNaet)
	require.Equal(t, uint32(8_000), override.TopNVotingPowerConcentrationBps)
	require.Equal(t, sdkmath.NewInt(1_000), override.SecurityNaet)
}

func TestEconomicSecurityMetricsRejectsManipulatedInputs(t *testing.T) {
	params := DefaultParams()
	params.MinStakeNaet = sdkmath.NewInt(100)
	validators := scoredCandidates(t, params, []Candidate{candidate("val-a", 1_000, 0)})
	windows := []RiskWindowRecord{riskWindow("delegator-a", "val-a", 1_000, RiskWindowStatusActive)}
	input := EconomicSecurityInput{
		Validators:			validators,
		RiskWindows:			windows,
		TopN:				1,
		ParticipatingValidators:	1,
		EligibleValidators:		1,
		AcceptedSlashEvents:		1,
		DetectedFaultEvents:		1,
		AcceptedEvidence:		1,
		SubmittedEvidence:		1,
		CompletedTasks:			1,
		ExpectedTasks:			1,
	}
	_, err := ComputeEconomicSecurityMetrics(input)
	require.NoError(t, err)

	tampered := input
	tampered.ParticipatingValidators = 2
	require.ErrorContains(t, computeSecurity(tampered), "participating validators")

	tampered = input
	tampered.AcceptedSlashEvents = 2
	require.ErrorContains(t, computeSecurity(tampered), "accepted slash events")

	tampered = input
	tampered.AcceptedEvidence = 2
	require.ErrorContains(t, computeSecurity(tampered), "accepted evidence")

	tampered = input
	tampered.CompletedTasks = 2
	require.ErrorContains(t, computeSecurity(tampered), "completed assigned tasks")

	tampered = input
	tampered.RiskWindows[0].AmountNaet = sdkmath.NewInt(2_000)
	require.ErrorContains(t, computeSecurity(tampered), "history root")
}

func TestCentralizationDashboardExposesControlsQueriesAndAlerts(t *testing.T) {
	params := DefaultParams()
	params.MinStakeNaet = sdkmath.NewInt(100)
	params.StakeSaturationThresholdNaet = sdkmath.NewInt(1_000)
	params.StakeSaturationCapFactorBps = BasisPoints
	validators := scoredCandidates(t, params, []Candidate{
		candidate("val-a", 3_000, 0),
		candidate("val-b", 1_000, 0),
		candidate("val-c", 100, 0),
	})
	securityInput := EconomicSecurityInput{
		Validators:			validators,
		RiskWindows:			[]RiskWindowRecord{riskWindow("delegator-a", "val-a", 800, RiskWindowStatusActive), riskWindow("delegator-b", "val-b", 200, RiskWindowStatusActive)},
		TopN:				1,
		ParticipatingValidators:	3,
		EligibleValidators:		3,
		AcceptedSlashEvents:		1,
		DetectedFaultEvents:		1,
		AcceptedEvidence:		1,
		SubmittedEvidence:		1,
		CompletedTasks:			5,
		ExpectedTasks:			5,
	}
	query, err := QuerySecurityMetrics(SecurityMetricQuery{Input: securityInput})
	require.NoError(t, err)
	require.Equal(t, uint32(4_761), query.Metrics.TopNVotingPowerConcentrationBps)

	controlParams := DefaultCentralizationControlParams(params)
	controlParams.MaxValidatorShareBps = 3_000
	controlParams.MaxTopNConcentrationBps = 4_500
	controlParams.MaxStakeSaturationRatioBps = 3_000
	controlParams.MaxDelegationRiskBucketBps = 5_000
	controlParams.MaxTaskAssignmentShareBps = 6_000
	controlParams.BootstrapMaxVotingPowerShareBps = 1_500
	dashboard, err := BuildCentralizationDashboard(CentralizationDashboardInput{
		SecurityInput:	securityInput,
		ControlParams:	controlParams,
		TaskAssignments: []CentralizationTaskAssignment{
			{TaskGroupID: "tg-a", ValidatorAddress: "val-a", AssignmentCount: 4},
			{TaskGroupID: "tg-b", ValidatorAddress: "val-b", AssignmentCount: 1},
		},
	})
	require.NoError(t, err)
	require.Equal(t, uint32(4_878), dashboard.Metrics.StakeSaturationRatioBps)
	require.Equal(t, uint32(8_000), dashboard.TaskAssignmentDiversity.MaxAssignmentShareBps)
	require.Equal(t, uint32(2_000), dashboard.TaskAssignmentDiversity.DiversityScoreBps)
	require.Contains(t, dashboard.TaskAssignmentDiversity.Warnings, CentralizationWarningTaskAssignmentShare)
	require.Len(t, dashboard.DelegationRiskWarnings, 1)
	require.Equal(t, "val-a", dashboard.DelegationRiskWarnings[0].ValidatorAddress)
	require.Equal(t, uint32(8_000), dashboard.DelegationRiskWarnings[0].ExposureShareBps)

	valA := centralizationControlByID(t, dashboard.ValidatorControls, "val-a")
	require.Equal(t, uint32(4_761), valA.VotingPowerShareBps)
	require.Equal(t, uint32(2_515), valA.RewardDampeningBps)
	require.Contains(t, valA.Warnings, CentralizationWarningValidatorShare)
	require.Contains(t, valA.Warnings, CentralizationWarningStakeSaturation)
	require.Contains(t, valA.Warnings, CentralizationWarningRewardDampeningActive)

	valC := centralizationControlByID(t, dashboard.ValidatorControls, "val-c")
	require.True(t, valC.BootstrapEligible)
	require.Contains(t, valC.Warnings, CentralizationWarningBootstrapEligible)
	requireAlert(t, dashboard.Alerts, CentralizationWarningTopNShare)
	requireAlert(t, dashboard.Alerts, CentralizationWarningStakeSaturation)
	requireAlert(t, dashboard.Alerts, CentralizationWarningDelegationRisk)
	requireAlert(t, dashboard.Alerts, CentralizationWarningTaskAssignmentShare)
}

func TestStakeConcentrationAndSplittingSimulations(t *testing.T) {
	params := DefaultParams()
	params.MinStakeNaet = sdkmath.NewInt(100)
	params.StakeSaturationThresholdNaet = sdkmath.NewInt(10_000)
	params.StakeSaturationCapFactorBps = BasisPoints
	concentration, err := SimulateStakeConcentration(StakeConcentrationSimulationInput{
		Params:				params,
		Candidates:			[]Candidate{candidate("val-a", 1_000, 0), candidate("val-b", 1_000, 0), candidate("val-c", 1_000, 0)},
		TargetValidatorID:		"val-a",
		AddedDelegatedStakeNaet:	sdkmath.NewInt(4_000),
		TopN:				1,
	})
	require.NoError(t, err)
	require.Equal(t, uint32(3_333), concentration.Before.TopNVotingPowerConcentrationBps)
	require.Equal(t, uint32(7_142), concentration.After.TopNVotingPowerConcentrationBps)
	require.Equal(t, int32(3_809), concentration.TopNConcentrationDeltaBps)
	require.Equal(t, sdkmath.NewInt(4_000), concentration.TargetEffectiveStakeDeltaNaet)
	requireAlert(t, concentration.Alerts, CentralizationWarningTopNShare)

	params.StakeSaturationThresholdNaet = sdkmath.NewInt(1_000)
	split, err := SimulateStakeSplitting(StakeSplittingSimulationInput{
		Params:		params,
		Candidate:	candidate("val-heavy", 3_000, 0),
		SplitCount:	3,
		TopN:		1,
	})
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(1_000), split.SingleEffectiveStakeNaet)
	require.Equal(t, sdkmath.NewInt(3_000), split.SplitEffectiveStakeNaet)
	require.Equal(t, sdkmath.NewInt(2_000), split.EffectiveStakeGainNaet)
	require.Equal(t, uint32(10_000), split.SingleConcentrationBps)
	require.Equal(t, uint32(3_333), split.SplitConcentrationBps)
}

func TestMaxValidatorSetChangesUsesConfiguredRate(t *testing.T) {
	params := DefaultParams()
	params.MaxValidatorSetChangeRateBps = 1_000

	changes, err := MaxValidatorSetChanges(params, 75)
	require.NoError(t, err)
	require.Equal(t, uint32(8), changes)

	params.MaxValidatorSetChangeRateBps = BasisPoints
	changes, err = MaxValidatorSetChanges(params, 75)
	require.NoError(t, err)
	require.Equal(t, uint32(75), changes)

	_, err = MaxValidatorSetChanges(params, 0)
	require.ErrorContains(t, err, "active validator count")
}

func TestBuildElectionCandidatesAppliesDelayedDelegationsToStakeWeight(t *testing.T) {
	params := DefaultParams()
	params.MinStakeNaet = sdkmath.NewInt(100)
	candidates := []Candidate{candidate("val-market", 1_000, 0)}
	intents := []DelegationIntent{{
		NominatorID:		"nom-a",
		ValidatorID:		"val-market",
		StakeNaet:		sdkmath.NewInt(2_000),
		RequestedEpoch:		10,
		MaxCommissionBps:	500,
		MinPerformanceScoreBps:	9_000,
	}}

	early, rejected, err := BuildElectionCandidates(params, 10, candidates, intents)
	require.NoError(t, err)
	require.Len(t, rejected, 1)
	require.True(t, early[0].DelegatedStakeNaet.IsZero())

	ready, rejected, err := BuildElectionCandidates(params, 11, candidates, intents)
	require.NoError(t, err)
	require.Empty(t, rejected)
	require.Equal(t, sdkmath.NewInt(2_000), ready[0].DelegatedStakeNaet)
	require.Equal(t, []Nomination{{NominatorID: "nom-a", StakeNaet: sdkmath.NewInt(2_000)}}, ready[0].Nominations)

	scored, err := ScoreCandidate(params, ready[0])
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(3_000), scored.TotalStakeNaet)
	require.Equal(t, sdkmath.NewInt(3_000), scored.ScoreComponents.StakeWeightNaet)
}

func TestValidatorElectionFactorCalculatorsUseDeterministicBps(t *testing.T) {
	performance, err := ComputePerformanceFactor(PerformanceFactorInput{
		CompletedTasks:		8,
		MissedTasks:		2,
		CorrectVerifications:	9,
		IncorrectVerifications:	1,
		AvailableWindows:	4,
		CommittedWindows:	5,
	})
	require.NoError(t, err)
	require.Equal(t, uint32(8_400), performance)

	uptime, err := ComputeUptimeFactor(UptimeFactorInput{
		SignedBlocks:			90,
		TotalBlocks:			100,
		TaskParticipations:		8,
		MissedTaskParticipations:	2,
	})
	require.NoError(t, err)
	require.Equal(t, uint32(8_700), uptime)

	latency, err := ComputeLatencyFactor(LatencyFactorInput{
		CommittedWindow:	true,
		TargetMillis:		1_000,
		P95Millis:		2_000,
	})
	require.NoError(t, err)
	require.Equal(t, uint32(5_000), latency)

	latency, err = ComputeLatencyFactor(LatencyFactorInput{CommittedWindow: true, AdvisoryOnly: true})
	require.NoError(t, err)
	require.Equal(t, uint32(BasisPoints), latency)

	reliability, err := ComputeReliabilityIndex(ReliabilityIndexInput{
		PriorIndexBps:		9_000,
		SlashEvents:		1,
		DowntimeEpochs:		2,
		MissedTasks:		3,
		RejectedEvidence:	2,
		RecoveryEpochs:		5,
	})
	require.NoError(t, err)
	require.Equal(t, uint32(5_700), reliability)
}

func TestValidatorElectionFactorCalculatorsRejectUnsafeInputs(t *testing.T) {
	_, err := ComputePerformanceFactor(PerformanceFactorInput{AvailableWindows: 2, CommittedWindows: 1})
	require.ErrorContains(t, err, "available windows")

	_, err = ComputeUptimeFactor(UptimeFactorInput{SignedBlocks: 2, TotalBlocks: 1})
	require.ErrorContains(t, err, "signed blocks")

	_, err = ComputeLatencyFactor(LatencyFactorInput{CommittedWindow: false, TargetMillis: 1, P95Millis: 1})
	require.ErrorContains(t, err, "committed measurement")

	_, err = ComputeReliabilityIndex(ReliabilityIndexInput{PriorIndexBps: BasisPoints + 1})
	require.ErrorContains(t, err, "prior reliability")
}

func TestPerformanceBasedRewardFormulaUsesDeterministicFixedPointMath(t *testing.T) {
	correctness, err := ComputeCorrectnessScore(CorrectnessScoreInput{
		ValidSignatures:	8,
		InvalidSignatures:	1,
		ValidTaskOutputs:	2,
		AcceptedEvidence:	1,
	})
	require.NoError(t, err)
	require.Equal(t, uint32(7_692), correctness)

	completion, err := ComputeTaskCompletionRate(TaskCompletionRateInput{
		CompletedAssignedTasks:	7,
		ExpectedAssignedTasks:	10,
	})
	require.NoError(t, err)
	require.Equal(t, uint32(7_000), completion)

	record, err := ComputePerformanceBasedReward(PerformanceRewardInput{
		EpochID:		14,
		ValidatorID:		"val-performance",
		BaseEmissionNaet:	sdkmath.NewInt(1_000_000),
		UptimeScoreBps:		9_000,
		LatencyScoreBps:	8_000,
		CorrectnessScoreBps:	7_500,
		TaskCompletionRateBps:	5_000,
	})
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(270_000), record.RewardNaet)
	require.Equal(t, ComputePerformanceRewardHash(record), record.RewardHash)
	require.NoError(t, record.Validate())
}

func TestPerformanceBasedRewardRejectsUnsafeInputs(t *testing.T) {
	completion, err := ComputeTaskCompletionRate(TaskCompletionRateInput{})
	require.NoError(t, err)
	require.Equal(t, uint32(BasisPoints), completion)

	_, err = ComputeTaskCompletionRate(TaskCompletionRateInput{CompletedAssignedTasks: 2, ExpectedAssignedTasks: 1})
	require.ErrorContains(t, err, "completed assigned tasks")

	_, err = ComputeCorrectnessScore(CorrectnessScoreInput{ValidSignatures: ^uint64(0), ValidTaskOutputs: 1})
	require.ErrorContains(t, err, "valid unit overflow")

	_, err = ComputeCorrectnessScore(CorrectnessScoreInput{AcceptedEvidence: ^uint64(0), EvidencePenaltyWeight: 2})
	require.ErrorContains(t, err, "evidence penalty overflow")

	_, err = ComputePerformanceBasedReward(PerformanceRewardInput{
		EpochID:		14,
		ValidatorID:		"val-performance",
		BaseEmissionNaet:	sdkmath.NewInt(-1),
		UptimeScoreBps:		BasisPoints,
		LatencyScoreBps:	BasisPoints,
		CorrectnessScoreBps:	BasisPoints,
		TaskCompletionRateBps:	BasisPoints,
	})
	require.ErrorContains(t, err, "base emission")

	_, err = ComputePerformanceBasedReward(PerformanceRewardInput{
		EpochID:		14,
		ValidatorID:		"val-performance",
		BaseEmissionNaet:	sdkmath.NewInt(1),
		UptimeScoreBps:		BasisPoints + 1,
		LatencyScoreBps:	BasisPoints,
		CorrectnessScoreBps:	BasisPoints,
		TaskCompletionRateBps:	BasisPoints,
	})
	require.ErrorContains(t, err, "uptime score")
}

func TestPerformanceRecordComputesRewardMultiplierAndDampening(t *testing.T) {
	record, err := BuildPerformanceRecord(PerformanceRecordInput{
		EpochID:		14,
		OperatorAddress:	"val-performance",
		Role:			ValidatorRoleCollator,
		AssignedTasks:		10,
		CompletedTasks:		7,
		MissedTasks:		2,
		InvalidTasks:		1,
		UptimeScoreBps:		9_000,
		LatencyScoreBps:	8_000,
		CorrectnessScoreBps:	7_500,
	})
	require.NoError(t, err)
	require.Equal(t, []string{
		"epoch_id",
		"operator_address",
		"role",
		"assigned_tasks",
		"completed_tasks",
		"missed_tasks",
		"invalid_tasks",
		"uptime_score",
		"latency_score",
		"correctness_score",
		"task_completion_rate",
		"reward_multiplier",
	}, PerformanceRecordFieldNames())
	require.Equal(t, uint32(7_000), record.TaskCompletionRateBps)
	require.Equal(t, uint32(3_780), record.RewardMultiplierBps)
	require.NoError(t, record.Validate())

	dampened, err := ApplyPerformanceDampening(PerformanceDampeningInput{
		Record:					record,
		CurrentRewardNaet:			sdkmath.NewInt(1_000_000),
		FutureElectionScoreBps:			9_000,
		DelegationAttractivenessBps:		8_000,
		RoleEligibilityBps:			8_500,
		CollatorAssignmentProbabilityBps:	7_000,
	})
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(378_000), dampened.CurrentRewardNaet)
	require.Equal(t, uint32(3_402), dampened.FutureElectionScoreBps)
	require.Equal(t, uint32(3_024), dampened.DelegationAttractivenessBps)
	require.Equal(t, uint32(3_213), dampened.RoleEligibilityBps)
	require.Equal(t, uint32(2_646), dampened.CollatorAssignmentProbabilityBps)
}

func TestPerformanceRecordRejectsInvalidTaskAccountingAndPreservesNonCollatorAssignmentProbability(t *testing.T) {
	_, err := BuildPerformanceRecord(PerformanceRecordInput{
		EpochID:		14,
		OperatorAddress:	"val-performance",
		Role:			ValidatorRoleVerifier,
		AssignedTasks:		2,
		CompletedTasks:		2,
		MissedTasks:		1,
		UptimeScoreBps:		BasisPoints,
		LatencyScoreBps:	BasisPoints,
		CorrectnessScoreBps:	BasisPoints,
	})
	require.ErrorContains(t, err, "task counts")

	record, err := BuildPerformanceRecord(PerformanceRecordInput{
		EpochID:		14,
		OperatorAddress:	"val-verifier",
		Role:			ValidatorRoleVerifier,
		AssignedTasks:		4,
		CompletedTasks:		4,
		UptimeScoreBps:		5_000,
		LatencyScoreBps:	5_000,
		CorrectnessScoreBps:	5_000,
	})
	require.NoError(t, err)
	dampened, err := ApplyPerformanceDampening(PerformanceDampeningInput{
		Record:					record,
		CurrentRewardNaet:			sdkmath.NewInt(1_000),
		FutureElectionScoreBps:			BasisPoints,
		DelegationAttractivenessBps:		BasisPoints,
		RoleEligibilityBps:			BasisPoints,
		CollatorAssignmentProbabilityBps:	7_000,
	})
	require.NoError(t, err)
	require.Equal(t, uint32(7_000), dampened.CollatorAssignmentProbabilityBps)

	record.RewardMultiplierBps++
	require.ErrorContains(t, record.Validate(), "reward multiplier")
}

func TestEpochLifecycleValidationRejectsReorderedOrDuplicatePhases(t *testing.T) {
	lifecycle := DefaultEpochLifecycle()
	slices.Reverse(lifecycle)
	require.ErrorContains(t, ValidateEpochLifecycle(lifecycle), "lifecycle step 0")

	lifecycle = DefaultEpochLifecycle()
	lifecycle[1] = lifecycle[0]
	require.ErrorContains(t, ValidateEpochLifecycle(lifecycle), "lifecycle step 1")
}

func TestScoreCandidateAppliesLatencyAndReliabilityFactors(t *testing.T) {
	params := DefaultParams()
	params.MinStakeNaet = sdkmath.NewInt(100)
	candidate := candidate("val-latency", 1_000, 0)
	candidate.LatencyFactorBps = 5_000
	candidate.ReliabilityIndexBps = 8_000

	scored, err := ScoreCandidate(params, candidate)
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(400), scored.Score)
}

func TestTaskAssignmentsAreSeededRoleAwareAndStable(t *testing.T) {
	params := DefaultParams()
	params.MinStakeNaet = sdkmath.NewInt(100)
	params.MinTaskGroupValidators = 2
	params.MaxTaskGroupValidators = 3

	candidates := makeCandidates(4, 1_000)
	candidates[0].Roles = []ValidatorRole{ValidatorRoleBlockProducer, ValidatorRoleVerifier}
	candidates[1].Roles = []ValidatorRole{ValidatorRoleBlockProducer}
	candidates[2].Roles = []ValidatorRole{ValidatorRoleVerifier}
	candidates[3].Roles = []ValidatorRole{ValidatorRoleBlockProducer, ValidatorRoleVerifier}
	validators := scoredCandidates(t, params, candidates)
	epoch, err := NewEpochRecord(params, 3, 30, 39, EpochPhaseAssignment, "", validators)
	require.NoError(t, err)

	tasks := []WorkloadTask{{
		TaskID:			"task-a",
		ZoneID:			"CORE",
		ShardID:		"shard-0",
		WorkloadClass:		"settlement",
		RequiredValidators:	2,
		Roles:			[]ValidatorRole{ValidatorRoleVerifier, ValidatorRoleBlockProducer},
	}}
	left, err := BuildTaskAssignments(params, epoch, validators, tasks)
	require.NoError(t, err)
	right, err := BuildTaskAssignments(params, epoch, validators, tasks)
	require.NoError(t, err)
	require.Equal(t, left.Root, right.Root)
	require.Len(t, left.Assignments, 2)
	require.NoError(t, left.Validate())

	candidateByID := map[string]Candidate{}
	for _, candidate := range candidates {
		candidateByID[candidate.ValidatorID] = candidate
	}
	for _, assignment := range left.Assignments {
		require.Len(t, assignment.Validators, 2)
		for _, validatorID := range assignment.Validators {
			require.True(t, ValidatorSupportsRole(candidateByID[validatorID], assignment.Role))
		}
	}

	_, err = BuildTaskAssignments(params, epoch, validators[1:], tasks)
	require.ErrorContains(t, err, "committed validator set hash")
}

func TestTaskGroupsCaptureWorkloadDomainsAndStakeRoots(t *testing.T) {
	params := DefaultParams()
	params.MinStakeNaet = sdkmath.NewInt(100)
	params.MinTaskGroupValidators = 2
	params.MaxTaskGroupValidators = 4

	candidates := makeCandidates(4, 1_000)
	candidates[0].Roles = []ValidatorRole{ValidatorRoleBlockProducer, ValidatorRoleVerifier, ValidatorRoleCollator}
	candidates[1].Roles = []ValidatorRole{ValidatorRoleVerifier, ValidatorRoleEvidenceReviewer}
	candidates[2].Roles = []ValidatorRole{ValidatorRoleCollator, ValidatorRoleVerifier}
	candidates[3].Roles = []ValidatorRole{ValidatorRoleEvidenceReviewer, ValidatorRoleVerifier}
	validators := scoredCandidates(t, params, candidates)
	epoch, err := NewEpochRecord(params, 4, 40, 59, EpochPhaseAssignment, "", validators)
	require.NoError(t, err)

	tasks := []WorkloadTask{
		{
			TaskID:			"zone-exec-a",
			WorkloadID:		"zone-alpha",
			WorkloadType:		WorkloadTypeZoneExecution,
			ZoneID:			"ZONE-A",
			ShardID:		"shard-0",
			WorkloadClass:		"execution",
			RequiredValidators:	2,
			Roles:			[]ValidatorRole{ValidatorRoleVerifier, ValidatorRoleCollator},
		},
		{
			TaskID:			"evidence-a",
			WorkloadID:		"evidence-market",
			WorkloadType:		WorkloadTypeEvidenceVerification,
			ZoneID:			"GLOBAL",
			ShardID:		"evidence",
			WorkloadClass:		"evidence",
			RequiredValidators:	2,
			Roles:			[]ValidatorRole{ValidatorRoleEvidenceReviewer, ValidatorRoleVerifier},
		},
	}
	left, err := BuildTaskGroups(params, epoch, validators, tasks, 41, 60)
	require.NoError(t, err)
	right, err := BuildTaskGroups(params, epoch, validators, tasks, 41, 60)
	require.NoError(t, err)
	require.Equal(t, left.Root, right.Root)
	require.Len(t, left.Groups, 2)
	require.NoError(t, left.Validate())

	group := left.Groups[0]
	require.Equal(t, uint64(4), group.EpochID)
	require.NotEmpty(t, group.TaskGroupID)
	require.Contains(t, []WorkloadType{WorkloadTypeEvidenceVerification, WorkloadTypeZoneExecution}, group.WorkloadType)
	require.GreaterOrEqual(t, len(group.ValidatorMembers), int(group.MinimumGroupSize))
	require.Len(t, group.ProposerOrder, len(group.ValidatorMembers))
	require.NotEmpty(t, group.VerifierSet)
	require.Len(t, group.StakeWeightRoot, PosHashHexLength)
	require.Equal(t, epoch.Seed, group.AssignmentSeed)
	require.Equal(t, uint64(41), group.ActivationHeight)
	require.Equal(t, uint64(60), group.ExpiryHeight)
}

func TestTaskGroupsValidateWorkloadTypesAndHeightWindow(t *testing.T) {
	params := DefaultParams()
	params.MinStakeNaet = sdkmath.NewInt(100)
	params.MinTaskGroupValidators = 2
	params.MaxTaskGroupValidators = 3
	candidates := makeCandidates(3, 1_000)
	for i := range candidates {
		candidates[i].Roles = []ValidatorRole{ValidatorRoleVerifier}
	}
	validators := scoredCandidates(t, params, candidates)
	epoch, err := NewEpochRecord(params, 5, 50, 70, EpochPhaseAssignment, "", validators)
	require.NoError(t, err)

	_, err = BuildTaskGroups(params, epoch, validators, []WorkloadTask{{
		TaskID:			"bad-workload",
		WorkloadID:		"bad-workload",
		WorkloadType:		WorkloadType("invalid"),
		ZoneID:			"GLOBAL",
		ShardID:		"shard",
		WorkloadClass:		"class",
		RequiredValidators:	2,
		Roles:			[]ValidatorRole{ValidatorRoleVerifier},
	}}, 51, 70)
	require.ErrorContains(t, err, "unsupported workload type")

	_, err = BuildTaskGroups(params, epoch, validators, []WorkloadTask{{
		TaskID:			"proof-a",
		WorkloadID:		"proof-a",
		WorkloadType:		WorkloadTypeProofVerification,
		ZoneID:			"GLOBAL",
		ShardID:		"proof",
		WorkloadClass:		"proof",
		RequiredValidators:	2,
		Roles:			[]ValidatorRole{ValidatorRoleVerifier},
	}}, 70, 70)
	require.ErrorContains(t, err, "expiry height")
}

func TestTaskAssignmentsRespectValidatorCapacityAndExclusions(t *testing.T) {
	params := DefaultParams()
	params.MinStakeNaet = sdkmath.NewInt(100)
	params.MinTaskGroupValidators = 1
	params.MaxTaskGroupValidators = 2
	candidates := makeCandidates(3, 1_000)
	for i := range candidates {
		candidates[i].Roles = []ValidatorRole{ValidatorRoleVerifier}
		candidates[i].Capacity = ValidatorCapacity{
			MaxTaskGroups:		1,
			SupportedWorkloads:	[]WorkloadType{WorkloadTypeProofVerification},
			ZoneSupport:		[]string{"ZONE-A"},
			HardwareClassOptional:	"hsm-small",
			NetworkClassOptional:	"wan-low-latency",
			AvailabilityCommitment:	9_800,
		}
	}
	validators := scoredCandidates(t, params, candidates)
	epoch, err := NewEpochRecord(params, 6, 60, 80, EpochPhaseAssignment, "", validators)
	require.NoError(t, err)

	assignments, err := BuildTaskAssignments(params, epoch, validators, []WorkloadTask{
		{
			TaskID:			"proof-a",
			WorkloadID:		"proof-a",
			WorkloadType:		WorkloadTypeProofVerification,
			ZoneID:			"ZONE-A",
			ShardID:		"proof",
			WorkloadClass:		"proof",
			RequiredValidators:	1,
			Roles:			[]ValidatorRole{ValidatorRoleVerifier},
			ExcludedValidators:	[]string{"val-000"},
		},
		{
			TaskID:			"proof-b",
			WorkloadID:		"proof-b",
			WorkloadType:		WorkloadTypeProofVerification,
			ZoneID:			"ZONE-A",
			ShardID:		"proof",
			WorkloadClass:		"proof",
			RequiredValidators:	1,
			Roles:			[]ValidatorRole{ValidatorRoleVerifier},
			ExcludedValidators:	[]string{"val-000"},
		},
	})
	require.NoError(t, err)
	require.Len(t, assignments.Assignments, 2)
	for _, assignment := range assignments.Assignments {
		require.NotContains(t, assignment.Validators, "val-000")
	}
	require.NotEqual(t, assignments.Assignments[0].Validators[0], assignments.Assignments[1].Validators[0])

	_, err = BuildTaskAssignments(params, epoch, validators, []WorkloadTask{{
		TaskID:			"unsupported",
		WorkloadID:		"unsupported",
		WorkloadType:		WorkloadTypeDataAvailability,
		ZoneID:			"ZONE-A",
		ShardID:		"da",
		WorkloadClass:		"da",
		RequiredValidators:	1,
		Roles:			[]ValidatorRole{ValidatorRoleVerifier},
	}})
	require.ErrorContains(t, err, "insufficient validators")

	_, err = BuildTaskAssignments(params, epoch, validators, []WorkloadTask{{
		TaskID:			"unsupported-zone",
		WorkloadID:		"unsupported-zone",
		WorkloadType:		WorkloadTypeProofVerification,
		ZoneID:			"ZONE-B",
		ShardID:		"proof",
		WorkloadClass:		"proof",
		RequiredValidators:	1,
		Roles:			[]ValidatorRole{ValidatorRoleVerifier},
	}})
	require.ErrorContains(t, err, "insufficient validators")
}

func TestValidatorCapacityDeclarationValidationAndSlashableFault(t *testing.T) {
	params := DefaultParams()
	params.MinStakeNaet = sdkmath.NewInt(100)
	candidate := candidate("val-capacity", 1_000, 0)
	candidate.Capacity = ValidatorCapacity{SupportedWorkloads: []WorkloadType{WorkloadTypeProofVerification}}
	require.ErrorContains(t, candidate.Validate(params), "max task groups")

	candidate.Capacity = ValidatorCapacity{MaxTaskGroups: 1, AvailabilityCommitment: BasisPoints + 1}
	require.ErrorContains(t, candidate.Validate(params), "availability commitment")

	slashable, err := IsSlashableCapacityFault(CapacityFaultEvidence{
		ValidatorID:		"val-capacity",
		WorkloadID:		"proof-a",
		WorkloadType:		WorkloadTypeProofVerification,
		AssignmentEpoch:	7,
		EvidenceHeight:		70,
		UsedForAssignment:	true,
		Finalized:		true,
	})
	require.NoError(t, err)
	require.True(t, slashable)

	slashable, err = IsSlashableCapacityFault(CapacityFaultEvidence{
		ValidatorID:		"val-capacity",
		WorkloadID:		"proof-a",
		WorkloadType:		WorkloadTypeProofVerification,
		AssignmentEpoch:	7,
		EvidenceHeight:		70,
		UsedForAssignment:	false,
		Finalized:		true,
	})
	require.NoError(t, err)
	require.False(t, slashable)
}

func TestDelegationIntentsRespectActivationDelayAndRiskProfile(t *testing.T) {
	params := DefaultParams()
	params.DelegationActivationEpochs = 2
	params.MinStakeNaet = sdkmath.NewInt(100)
	candidates := []Candidate{candidate("val-market", 1_000, 0)}
	candidates[0].PerformanceScoreBps = 9_000

	intents := []DelegationIntent{
		{
			NominatorID:		"nom-a",
			ValidatorID:		"val-market",
			StakeNaet:		sdkmath.NewInt(500),
			RequestedEpoch:		10,
			MaxCommissionBps:	500,
			MinPerformanceScoreBps:	8_000,
		},
	}
	activations, rejected, err := ActivateDelegationIntents(params, 11, candidates, intents)
	require.NoError(t, err)
	require.Empty(t, activations)
	require.Len(t, rejected, 1)
	require.Contains(t, rejected[0].Reason, "activation delay")

	intents = append(intents, DelegationIntent{
		NominatorID:		"nom-b",
		ValidatorID:		"val-market",
		StakeNaet:		sdkmath.NewInt(500),
		RequestedEpoch:		10,
		MaxCommissionBps:	500,
		MinPerformanceScoreBps:	9_500,
	})
	activations, rejected, err = ActivateDelegationIntents(params, 12, candidates, intents)
	require.NoError(t, err)
	require.Len(t, activations, 1)
	require.Equal(t, "val-market", activations[0].ValidatorID)
	require.Equal(t, sdkmath.NewInt(500), activations[0].TotalStake)
	require.Len(t, rejected, 1)
	require.Contains(t, rejected[0].Reason, "performance below")
}

func TestEvidenceSettlementPaysReporterFromObjectiveSlash(t *testing.T) {
	params := DefaultParams()
	params.ReporterRewardBps = 500
	settlement, err := SettleEvidenceCase(params, 12, EvidenceCase{
		EvidenceID:		"evidence-1",
		ReporterID:		"reporter-a",
		ValidatorID:		"val-a",
		Misbehavior:		MisbehaviorDoubleSign,
		SlashFractionBps:	1_000,
		EvidenceHeight:		99,
		EvidenceEpoch:		10,
		Finalized:		true,
	}, sdkmath.NewInt(1_000), []Nomination{{NominatorID: "nom-a", StakeNaet: sdkmath.NewInt(1_000)}})
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(200), settlement.Slash.TotalSlashedNaet)
	require.Equal(t, sdkmath.NewInt(10), settlement.ReporterRewardNaet)
	require.Equal(t, sdkmath.NewInt(190), settlement.BurnNaet)
	require.Len(t, settlement.SettlementHash, PosHashHexLength)

	_, err = SettleEvidenceCase(params, 12, EvidenceCase{
		EvidenceID:		"evidence-stale",
		ReporterID:		"reporter-a",
		ValidatorID:		"val-a",
		Misbehavior:		MisbehaviorDowntime,
		SlashFractionBps:	100,
		EvidenceHeight:		99,
		EvidenceEpoch:		6,
		Finalized:		true,
	}, sdkmath.NewInt(1_000), nil)
	require.ErrorContains(t, err, "outside slashable window")
}

func TestStructuredEvidenceTypesMapToSlashPolicies(t *testing.T) {
	types := StructuredEvidenceTypes()
	require.Equal(t, []string{
		EvidenceTypeDoubleSignProof,
		EvidenceTypeInvalidStateTransitionProof,
		EvidenceTypeEquivocationProof,
		EvidenceTypeDowntimeProof,
		EvidenceTypeInvalidTaskExecutionProof,
		EvidenceTypeInvalidCollatorOutputProof,
		EvidenceTypeInvalidProofAcceptance,
		EvidenceTypeFalseCapacityDeclaration,
		EvidenceTypeInvalidEvidenceSubmission,
	}, types)

	for _, evidenceType := range types {
		policy, err := DefaultEvidenceSlashPolicy(evidenceType)
		require.NoError(t, err)
		require.Equal(t, evidenceType, policy.EvidenceType)
		require.True(t, IsSlashableMisbehavior(policy.Misbehavior))
		require.NotZero(t, policy.SlashFractionBps)
		require.LessOrEqual(t, policy.SlashFractionBps, BasisPoints)
	}

	_, err := DefaultEvidenceSlashPolicy("unknown")
	require.ErrorContains(t, err, "unsupported structured evidence type")
}

func TestEvidenceRecordMatchesDesignFieldsAndStatusValues(t *testing.T) {
	require.Equal(t, []string{
		"evidence_id",
		"evidence_type",
		"accused_validator",
		"reporter",
		"epoch_id",
		"task_group_id_optional",
		"object_hash",
		"proof_payload_hash",
		"submitted_height",
		"status",
		"verification_group_id",
		"decision_height",
		"penalty_id_optional",
	}, EvidenceRecordFieldNames())
	require.Equal(t, []string{
		EvidenceStatusSubmitted,
		EvidenceStatusInVerification,
		EvidenceStatusAccepted,
		EvidenceStatusRejected,
		EvidenceStatusExpired,
		EvidenceStatusSlashed,
	}, EvidenceRecordStatusValues())

	record, err := NewEvidenceRecord(EvidenceRecord{
		EvidenceID:		"evidence-record-1",
		EvidenceType:		EvidenceTypeInvalidProofAcceptance,
		AccusedValidator:	"val-000",
		Reporter:		"val-001",
		EpochID:		7,
		TaskGroupIDOptional:	"task-group-1",
		ObjectHash:		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		ProofPayloadHash:	"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		SubmittedHeight:	70,
	})
	require.NoError(t, err)
	require.Equal(t, EvidenceStatusSubmitted, record.Status)
	require.Len(t, computeEvidenceRecordHash(record), PosHashHexLength)

	_, err = AdvanceEvidenceRecordStatus(record, EvidenceStatusAccepted, 71, "")
	require.ErrorContains(t, err, "invalid evidence record status transition")
}

func TestEvidenceVerificationGroupSelectionIsDeterministicAndExcludesParties(t *testing.T) {
	params := DefaultParams()
	validators := scoredCandidates(t, params, makeCandidates(6, 1_000_000_000))
	epoch, err := NewEpochRecord(params, 7, 70, 90, EpochPhaseAssignment, "", validators)
	require.NoError(t, err)
	record, err := NewEvidenceRecord(EvidenceRecord{
		EvidenceID:		"evidence-record-2",
		EvidenceType:		EvidenceTypeFalseCapacityDeclaration,
		AccusedValidator:	"val-000",
		Reporter:		"val-001",
		EpochID:		epoch.EpochID,
		ObjectHash:		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		ProofPayloadHash:	"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		SubmittedHeight:	72,
	})
	require.NoError(t, err)

	left, err := SelectEvidenceVerificationGroup(EvidenceVerificationGroupInput{
		Params:			params,
		Epoch:			epoch,
		ActiveValidators:	validators,
		Evidence:		record,
		MinimumGroupSize:	3,
		DecisionThresholdBps:	7_000,
	})
	require.NoError(t, err)
	require.Len(t, left.Members, 3)
	require.NotContains(t, left.Members, "val-000")
	require.NotContains(t, left.Members, "val-001")
	require.Equal(t, []string{"val-000", "val-001"}, left.ExcludedValidators)
	require.Equal(t, uint32(7_000), left.DecisionThresholdBps)
	require.Len(t, left.AssignmentSeed, PosHashHexLength)
	require.Len(t, left.VerificationGroupID, PosHashHexLength)
	require.Len(t, left.GroupHash, PosHashHexLength)

	reversed := append([]ScoredValidator(nil), validators...)
	for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = reversed[j], reversed[i]
	}
	right, err := SelectEvidenceVerificationGroup(EvidenceVerificationGroupInput{
		Params:			params,
		Epoch:			epoch,
		ActiveValidators:	reversed,
		Evidence:		record,
		MinimumGroupSize:	3,
		DecisionThresholdBps:	7_000,
	})
	require.NoError(t, err)
	require.Equal(t, left, right)

	inVerification, err := AssignEvidenceVerificationGroup(record, left)
	require.NoError(t, err)
	require.Equal(t, EvidenceStatusInVerification, inVerification.Status)
	require.Equal(t, left.VerificationGroupID, inVerification.VerificationGroupID)

	accepted, err := AdvanceEvidenceRecordStatus(inVerification, EvidenceStatusAccepted, 80, "")
	require.NoError(t, err)
	require.Equal(t, int64(80), accepted.DecisionHeight)
	slashed, err := AdvanceEvidenceRecordStatus(accepted, EvidenceStatusSlashed, 81, "penalty-1")
	require.NoError(t, err)
	require.Equal(t, EvidenceStatusSlashed, slashed.Status)
	require.Equal(t, "penalty-1", slashed.PenaltyIDOptional)
}

func TestEvidenceVerificationGroupRejectsInsufficientEligibleValidators(t *testing.T) {
	params := DefaultParams()
	validators := scoredCandidates(t, params, makeCandidates(4, 1_000_000_000))
	epoch, err := NewEpochRecord(params, 8, 80, 100, EpochPhaseAssignment, "", validators)
	require.NoError(t, err)
	record, err := NewEvidenceRecord(EvidenceRecord{
		EvidenceID:		"evidence-record-3",
		EvidenceType:		EvidenceTypeDowntimeProof,
		AccusedValidator:	"val-000",
		Reporter:		"val-001",
		EpochID:		epoch.EpochID,
		ObjectHash:		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		ProofPayloadHash:	"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		SubmittedHeight:	82,
	})
	require.NoError(t, err)

	_, err = SelectEvidenceVerificationGroup(EvidenceVerificationGroupInput{
		Params:			params,
		Epoch:			epoch,
		ActiveValidators:	validators,
		Evidence:		record,
		MinimumGroupSize:	3,
	})
	require.ErrorContains(t, err, "insufficient eligible validators")

	_, err = SelectEvidenceVerificationGroup(EvidenceVerificationGroupInput{
		Params:			params,
		Epoch:			epoch,
		ActiveValidators:	validators,
		Evidence:		record,
		MinimumGroupSize:	2,
		DecisionThresholdBps:	BasisPoints + 1,
	})
	require.ErrorContains(t, err, "decision threshold")
}

func TestStructuredEvidenceLifecycleFinalizesAndExecutesSlash(t *testing.T) {
	params := DefaultParams()
	params.ReporterRewardBps = 500
	evidence, err := SubmitStructuredEvidence(StructuredEvidenceRecord{
		EvidenceID:		"evidence-structured-1",
		EvidenceType:		EvidenceTypeInvalidTaskExecutionProof,
		ReporterID:		"reporter-a",
		AccusedValidatorID:	"val-a",
		SubjectID:		"task/proof/1",
		EvidenceHash:		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		EvidenceHeight:		101,
		EvidenceEpoch:		10,
		SubmittedHeight:	102,
		VerificationGroupID:	"evidence-group-a",
	})
	require.NoError(t, err)
	require.Equal(t, EvidenceStatusSubmitted, evidence.Status)
	require.Len(t, evidence.StructuredRecordHash, PosHashHexLength)

	verification, err := VerifyStructuredEvidenceBySubset(evidence, []string{"reviewer-a", "reviewer-b", "reviewer-c"}, []EvidenceVerificationVote{
		{EvidenceID: evidence.EvidenceID, ReviewerID: "reviewer-b", Accepted: true, SignatureHash: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", VoteHeight: 103},
		{EvidenceID: evidence.EvidenceID, ReviewerID: "reviewer-a", Accepted: true, SignatureHash: "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc", VoteHeight: 103},
		{EvidenceID: evidence.EvidenceID, ReviewerID: "reviewer-c", Accepted: true, SignatureHash: "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd", VoteHeight: 103},
	}, 6_700)
	require.NoError(t, err)
	require.True(t, verification.Accepted)
	require.Equal(t, EvidenceStatusVerified, verification.Status)
	require.Equal(t, uint32(10_000), verification.ParticipationBps)
	require.Len(t, verification.VerificationRoot, PosHashHexLength)

	finality, err := FinalizeStructuredEvidence(evidence, verification, []EvidenceFinalityVote{
		{EvidenceID: evidence.EvidenceID, ValidatorID: "val-voter-a", Approve: true, VotingPowerBps: 4_000, SignatureHash: "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee", FinalityHeight: 104},
		{EvidenceID: evidence.EvidenceID, ValidatorID: "val-voter-b", Approve: true, VotingPowerBps: 3_000, SignatureHash: "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", FinalityHeight: 104},
		{EvidenceID: evidence.EvidenceID, ValidatorID: "val-voter-c", Approve: false, VotingPowerBps: 1_000, SignatureHash: "1111111111111111111111111111111111111111111111111111111111111111", FinalityHeight: 104},
	}, 6_700)
	require.NoError(t, err)
	require.True(t, finality.Finalized)
	require.True(t, finality.Accepted)
	require.Equal(t, EvidenceStatusFinalized, finality.Status)
	require.Equal(t, uint32(7_000), finality.AcceptedPowerBps)
	require.Len(t, finality.FinalityVoteRoot, PosHashHexLength)

	settlement, err := ExecuteStructuredEvidenceSlashing(params, 12, evidence, finality, sdkmath.NewInt(1_000), []Nomination{
		{NominatorID: "nom-a", StakeNaet: sdkmath.NewInt(1_000)},
	})
	require.NoError(t, err)
	require.Equal(t, evidence.EvidenceID, settlement.EvidenceID)
	require.Equal(t, MisbehaviorInvalidBlock, settlement.Slash.Misbehavior)
	require.Equal(t, sdkmath.NewInt(150), settlement.Slash.TotalSlashedNaet)
	require.Equal(t, sdkmath.NewInt(7), settlement.ReporterRewardNaet)
	require.Equal(t, sdkmath.NewInt(143), settlement.BurnNaet)
	require.Len(t, settlement.SettlementHash, PosHashHexLength)
}

func TestStructuredEvidenceLifecycleRejectsUnverifiedOrRejectedEvidence(t *testing.T) {
	evidence, err := SubmitStructuredEvidence(StructuredEvidenceRecord{
		EvidenceID:		"evidence-structured-2",
		EvidenceType:		EvidenceTypeDowntimeProof,
		ReporterID:		"reporter-a",
		AccusedValidatorID:	"val-a",
		SubjectID:		"height/100",
		EvidenceHash:		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		EvidenceHeight:		100,
		EvidenceEpoch:		10,
		SubmittedHeight:	101,
		VerificationGroupID:	"evidence-group-a",
	})
	require.NoError(t, err)

	verification, err := VerifyStructuredEvidenceBySubset(evidence, []string{"reviewer-a", "reviewer-b", "reviewer-c"}, []EvidenceVerificationVote{
		{EvidenceID: evidence.EvidenceID, ReviewerID: "reviewer-a", Accepted: false, SignatureHash: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", VoteHeight: 102},
		{EvidenceID: evidence.EvidenceID, ReviewerID: "reviewer-b", Accepted: false, SignatureHash: "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc", VoteHeight: 102},
		{EvidenceID: evidence.EvidenceID, ReviewerID: "reviewer-c", Accepted: false, SignatureHash: "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd", VoteHeight: 102},
	}, 6_700)
	require.NoError(t, err)
	require.True(t, verification.Rejected)
	require.Equal(t, EvidenceStatusRejected, verification.Status)

	_, err = FinalizeStructuredEvidence(evidence, verification, []EvidenceFinalityVote{
		{EvidenceID: evidence.EvidenceID, ValidatorID: "val-voter-a", Approve: true, VotingPowerBps: 7_000, SignatureHash: "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee", FinalityHeight: 103},
	}, 6_700)
	require.ErrorContains(t, err, "verified by consensus subset")

	_, err = VerifyStructuredEvidenceBySubset(evidence, []string{"reviewer-a"}, []EvidenceVerificationVote{
		{EvidenceID: evidence.EvidenceID, ReviewerID: "reviewer-x", Accepted: true, SignatureHash: "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", VoteHeight: 102},
	}, 6_700)
	require.ErrorContains(t, err, "not assigned")
}

func TestWorkloadRewardsSplitByRoleAndCompletedUnits(t *testing.T) {
	settlement, err := SettleWorkloadRewards(WorkloadRewardInput{
		EpochID:		9,
		TotalRewardsNaet:	sdkmath.NewInt(1_000),
		RoleWeights: []RoleRewardWeight{
			{Role: ValidatorRoleBlockProducer, WeightBps: 5_000},
			{Role: ValidatorRoleVerifier, WeightBps: 5_000},
		},
		Outcomes: []AssignmentOutcome{
			{TaskID: "task-a", Role: ValidatorRoleBlockProducer, ValidatorID: "val-a", Completed: true, WorkUnits: 3},
			{TaskID: "task-a", Role: ValidatorRoleBlockProducer, ValidatorID: "val-b", Completed: true, WorkUnits: 1},
			{TaskID: "task-a", Role: ValidatorRoleVerifier, ValidatorID: "val-c", Completed: true, WorkUnits: 2},
			{TaskID: "task-a", Role: ValidatorRoleVerifier, ValidatorID: "val-d", Faulted: true, WorkUnits: 10},
		},
	})
	require.NoError(t, err)
	require.Equal(t, uint64(6), settlement.CompletedUnits)
	require.True(t, settlement.RemainderNaet.IsZero())
	require.Equal(t, []ValidatorWorkloadReward{
		{ValidatorID: "val-a", RewardNaet: sdkmath.NewInt(375), WorkUnits: 3},
		{ValidatorID: "val-b", RewardNaet: sdkmath.NewInt(125), WorkUnits: 1},
		{ValidatorID: "val-c", RewardNaet: sdkmath.NewInt(500), WorkUnits: 2},
	}, settlement.Rewards)
	require.Len(t, settlement.RewardRoot, PosHashHexLength)
}

func TestValidatorRoleExpansionAndRoleRecordFields(t *testing.T) {
	require.Equal(t, []ValidatorRole{
		ValidatorRoleValidator,
		ValidatorRoleProposer,
		ValidatorRoleVerifier,
		ValidatorRoleEvidenceReporter,
		ValidatorRoleDelegationOperator,
		ValidatorRoleCollator,
		ValidatorRoleFisherman,
	}, ValidatorRoleValues())
	require.Equal(t, []string{
		"validator_address",
		"role",
		"epoch_id",
		"status",
		"eligibility_score",
		"capacity",
		"assigned_task_count",
		"performance_score",
	}, RoleRecordFieldNames())
	require.Equal(t, []string{
		RoleStatusEligible,
		RoleStatusAssigned,
		RoleStatusSuspended,
		RoleStatusInactive,
	}, RoleStatusValues())
	for _, role := range ValidatorRoleValues() {
		require.NoError(t, validateValidatorRole(role))
	}
}

func TestRoleRecordsAllowOverlapButRejectDuplicateRoleForEpoch(t *testing.T) {
	capacity := ValidatorCapacity{MaxTaskGroups: 3, SupportedWorkloads: []WorkloadType{WorkloadTypeProofVerification}, AvailabilityCommitment: 9_000}
	verifier, err := NewRoleRecord(RoleRecord{
		ValidatorAddress:	"val-a",
		Role:			ValidatorRoleVerifier,
		EpochID:		12,
		Status:			RoleStatusAssigned,
		EligibilityScore:	9_500,
		Capacity:		capacity,
		AssignedTaskCount:	2,
		PerformanceScore:	9_700,
	})
	require.NoError(t, err)
	reporter, err := NewRoleRecord(RoleRecord{
		ValidatorAddress:	"val-a",
		Role:			ValidatorRoleEvidenceReporter,
		EpochID:		12,
		Status:			RoleStatusEligible,
		EligibilityScore:	9_300,
		Capacity:		capacity,
		AssignedTaskCount:	0,
		PerformanceScore:	9_400,
	})
	require.NoError(t, err)
	require.NoError(t, ValidateRoleRecords([]RoleRecord{verifier, reporter}))

	duplicate := reporter
	duplicate.PerformanceScore = 9_200
	require.ErrorContains(t, ValidateRoleRecords([]RoleRecord{reporter, duplicate}), "duplicate role record")
}

func TestRoleRecordRejectsCapacityOverflowAndInvalidScores(t *testing.T) {
	_, err := NewRoleRecord(RoleRecord{
		ValidatorAddress:	"val-a",
		Role:			ValidatorRoleCollator,
		EpochID:		12,
		Status:			RoleStatusAssigned,
		EligibilityScore:	9_500,
		Capacity:		ValidatorCapacity{MaxTaskGroups: 1, SupportedWorkloads: []WorkloadType{WorkloadTypeShardExecution}},
		AssignedTaskCount:	2,
		PerformanceScore:	9_000,
	})
	require.ErrorContains(t, err, "exceeds capacity")

	_, err = NewRoleRecord(RoleRecord{
		ValidatorAddress:	"val-a",
		Role:			ValidatorRoleFisherman,
		EpochID:		12,
		Status:			RoleStatusEligible,
		EligibilityScore:	BasisPoints + 1,
	})
	require.ErrorContains(t, err, "eligibility score")

	_, err = NewRoleRecord(RoleRecord{
		ValidatorAddress:	"val-a",
		Role:			ValidatorRoleProposer,
		EpochID:		12,
		Status:			RoleStatusAssigned,
	})
	require.ErrorContains(t, err, "assigned task count")
}

func TestRoleRegistryEligibilityChecksRoleSpecificRules(t *testing.T) {
	params := DefaultParams()
	registry := DefaultRoleRegistry()
	require.NoError(t, registry.Validate())
	proposerRule, found, err := registry.Rule(ValidatorRoleProposer)
	require.NoError(t, err)
	require.True(t, found)
	require.True(t, proposerRule.RequiresValidator)
	require.True(t, proposerRule.RequiresMinimumStake)

	proposer := candidate("val-proposer", 1_000_000_000, 0)
	proposer.Roles = []ValidatorRole{ValidatorRoleProposer}
	proposer.PerformanceScoreBps = BasisPoints
	proposer.UptimeFactorBps = BasisPoints
	record, err := CheckRoleEligibility(registry, RoleEligibilityInput{
		Params:		params,
		Role:		ValidatorRoleProposer,
		Candidate:	proposer,
	})
	require.NoError(t, err)
	require.Equal(t, "val-proposer", record.ValidatorAddress)
	require.Equal(t, ValidatorRoleProposer, record.Role)
	require.Equal(t, uint32(BasisPoints), record.EligibilityScore)

	lowStake := proposer
	lowStake.SelfStakeNaet = sdkmath.NewInt(1)
	_, err = CheckRoleEligibility(registry, RoleEligibilityInput{Params: params, Role: ValidatorRoleProposer, Candidate: lowStake})
	require.ErrorContains(t, err, "minimum validator stake")

	fisherman, err := CheckRoleEligibility(registry, RoleEligibilityInput{
		Params:		params,
		Role:		ValidatorRoleFisherman,
		ActorAddress:	"fish-1",
		DepositNaet:	sdkmath.NewInt(1),
	})
	require.NoError(t, err)
	require.Equal(t, "fish-1", fisherman.ValidatorAddress)
	require.Equal(t, ValidatorRoleFisherman, fisherman.Role)

	_, err = CheckRoleEligibility(registry, RoleEligibilityInput{Params: params, Role: ValidatorRoleFisherman, ActorAddress: "fish-1"})
	require.ErrorContains(t, err, "deposit")

	operator := candidate("val-operator", 1_000_000_000, 0)
	operator.Roles = []ValidatorRole{ValidatorRoleDelegationOperator}
	operator.PerformanceScoreBps = BasisPoints
	operator.UptimeFactorBps = BasisPoints
	_, err = CheckRoleEligibility(registry, RoleEligibilityInput{
		Params:		params,
		Role:		ValidatorRoleDelegationOperator,
		Candidate:	operator,
	})
	require.ErrorContains(t, err, "authorization")
	_, err = CheckRoleEligibility(registry, RoleEligibilityInput{
		Params:				params,
		Role:				ValidatorRoleDelegationOperator,
		Candidate:			operator,
		DelegationOperatorAuthorized:	true,
		FeesDisclosed:			true,
		RiskPolicyDisclosed:		true,
	})
	require.NoError(t, err)
}

func TestRolePerformanceMetricsRewardsAndSuspensionRespectOverlap(t *testing.T) {
	capacity := ValidatorCapacity{MaxTaskGroups: 3, SupportedWorkloads: []WorkloadType{WorkloadTypeProofVerification}, AvailabilityCommitment: 9_000}
	verifier, err := NewRoleRecord(RoleRecord{
		ValidatorAddress:	"val-a",
		Role:			ValidatorRoleVerifier,
		EpochID:		12,
		Status:			RoleStatusAssigned,
		EligibilityScore:	9_500,
		Capacity:		capacity,
		AssignedTaskCount:	3,
		PerformanceScore:	9_500,
	})
	require.NoError(t, err)
	collator, err := NewRoleRecord(RoleRecord{
		ValidatorAddress:	"val-a",
		Role:			ValidatorRoleCollator,
		EpochID:		12,
		Status:			RoleStatusAssigned,
		EligibilityScore:	9_000,
		Capacity:		capacity,
		AssignedTaskCount:	1,
		PerformanceScore:	8_000,
	})
	require.NoError(t, err)
	metrics, err := ComputeRolePerformanceMetrics(verifier, 2, 1, 0)
	require.NoError(t, err)
	require.Equal(t, uint32(4_166), metrics.PerformanceScore)

	rewards, err := SettleRoleRewards(RoleRewardInput{
		EpochID:		12,
		TotalRewardsNaet:	sdkmath.NewInt(1_000),
		Records:		[]RoleRecord{verifier, collator},
		Weights: []RoleRewardWeight{
			{Role: ValidatorRoleVerifier, WeightBps: 5_000},
			{Role: ValidatorRoleCollator, WeightBps: 5_000},
		},
	})
	require.NoError(t, err)
	require.Len(t, rewards.Rewards, 1)
	require.Equal(t, "val-a", rewards.Rewards[0].ValidatorID)
	require.Equal(t, sdkmath.NewInt(1_000), rewards.Rewards[0].RewardNaet)

	suspended, err := SuspendRoleOnFault([]RoleRecord{verifier, collator}, "val-a", ValidatorRoleVerifier, 12)
	require.NoError(t, err)
	require.Equal(t, RoleStatusSuspended, suspended[0].Status)
	require.Equal(t, RoleStatusAssigned, suspended[1].Status)

	penalty, err := ComputeSlashingPenalty(SlashingPenaltyInput{
		PenaltyID:		"penalty-role-specific",
		ValidatorID:		"val-a",
		SeverityLevel:		SlashSeverityInvalidTaskExecution,
		StakeExposureNaet:	sdkmath.NewInt(1_000),
		SelfStakeNaet:		sdkmath.NewInt(1_000),
		RoleSuspensions:	[]ValidatorRole{ValidatorRoleVerifier},
	})
	require.NoError(t, err)
	candidate := candidate("val-a", 1_000_000_000, 0)
	candidate.Roles = []ValidatorRole{ValidatorRoleVerifier, ValidatorRoleCollator}
	applied, err := ApplySlashingPenaltyToCandidate(candidate, penalty)
	require.NoError(t, err)
	require.Equal(t, []ValidatorRole{ValidatorRoleCollator}, applied.Roles)
}

func TestCollatorRecordMatchesSpecAndValidatesRegistration(t *testing.T) {
	record, err := NewCollatorRecord(CollatorRecord{
		CollatorID:		"collator-1",
		OperatorAddress:	"operator-1",
		SupportedWorkloads:	[]WorkloadType{WorkloadTypeShardExecution, WorkloadTypeZoneExecution},
		BondOptional:		sdkmath.NewInt(100),
		Reputation:		9_100,
		RegisteredEpoch:	13,
	})
	require.NoError(t, err)
	require.Equal(t, CollatorStatusRegistered, record.Status)
	require.Equal(t, []string{
		"collator_id",
		"operator_address",
		"supported_workloads",
		"bond_optional",
		"reputation",
		"status",
		"registered_epoch",
	}, CollatorRecordFieldNames())
	require.Equal(t, []string{
		CollatorStatusRegistered,
		CollatorStatusActive,
		CollatorStatusSuspended,
		CollatorStatusRetired,
	}, CollatorStatusValues())
	require.True(t, record.SupportsWorkload(WorkloadTypeShardExecution))
	require.False(t, record.SupportsWorkload(WorkloadTypeEvidenceVerification))

	active := record
	active.CollatorID = "collator-2"
	active.Status = CollatorStatusActive
	registry, err := NewCollatorRegistry(13, []CollatorRecord{active, record})
	require.NoError(t, err)
	require.Equal(t, ComputeCollatorRegistryRoot(registry), registry.RegistryRoot)
	found, ok, err := registry.CollatorByID("collator-1")
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "operator-1", found.OperatorAddress)
	activeForShard, err := registry.ActiveCollatorsForWorkload(WorkloadTypeShardExecution)
	require.NoError(t, err)
	require.Len(t, activeForShard, 1)
	require.Equal(t, "collator-2", activeForShard[0].CollatorID)

	tamperedRegistry := registry
	tamperedRegistry.RegistryRoot = strings.Repeat("1", PosHashHexLength)
	require.ErrorContains(t, tamperedRegistry.Validate(), "registry root")

	_, err = NewCollatorRecord(CollatorRecord{
		CollatorID:		"collator-dup",
		OperatorAddress:	"operator-1",
		SupportedWorkloads:	[]WorkloadType{WorkloadTypeShardExecution, WorkloadTypeShardExecution},
		RegisteredEpoch:	13,
	})
	require.ErrorContains(t, err, "duplicate supported workload")

	_, err = NewCollatorRecord(CollatorRecord{
		CollatorID:		"collator-bad-bond",
		OperatorAddress:	"operator-1",
		SupportedWorkloads:	[]WorkloadType{WorkloadTypeShardExecution},
		BondOptional:		sdkmath.NewInt(-1),
		RegisteredEpoch:	13,
	})
	require.ErrorContains(t, err, "bond optional")

	_, err = NewCollatorRecord(CollatorRecord{
		CollatorID:		"collator-bad-reputation",
		OperatorAddress:	"operator-1",
		SupportedWorkloads:	[]WorkloadType{WorkloadTypeShardExecution},
		Reputation:		BasisPoints + 1,
		RegisteredEpoch:	13,
	})
	require.ErrorContains(t, err, "reputation")
}

func TestCollatorBuildsCandidateOutputButRequiresValidatorVerification(t *testing.T) {
	params := DefaultParams()
	collator, err := NewCollatorRecord(CollatorRecord{
		CollatorID:		"collator-1",
		OperatorAddress:	"operator-1",
		SupportedWorkloads:	[]WorkloadType{WorkloadTypeShardExecution},
		Status:			CollatorStatusActive,
		RegisteredEpoch:	13,
	})
	require.NoError(t, err)
	task := WorkloadTask{
		TaskID:			"task-1",
		WorkloadID:		"workload-1",
		WorkloadType:		WorkloadTypeShardExecution,
		ZoneID:			"zone-a",
		ShardID:		"shard-1",
		WorkloadClass:		DefaultWorkloadClass,
		RequiredValidators:	params.MinTaskGroupValidators,
		Roles:			[]ValidatorRole{ValidatorRoleCollator, ValidatorRoleVerifier},
	}
	output, err := BuildCollatorCandidateOutput(params, CollatorCandidateOutputInput{
		EpochID:		13,
		Collator:		collator,
		Task:			task,
		TaskGroupIDOptional:	"task-group-1",
		TransactionRoot:	PosEmptyRootHash,
		StateTransitionRoot:	PosEmptyRootHash,
		ProofBundleRoot:	PosEmptyRootHash,
	})
	require.NoError(t, err)
	require.Equal(t, "collator-1", output.CollatorID)
	require.True(t, output.RequiresValidatorVerification)
	require.False(t, output.Finalized)
	require.Empty(t, output.ValidatorSignatures)
	require.Equal(t, ComputeCollatorCandidateOutputHash(output), output.CandidateOutputHash)
	require.NoError(t, output.Validate())

	tampered := output
	tampered.ProofBundleRoot = strings.Repeat("0", PosHashHexLength)
	require.ErrorContains(t, tampered.Validate(), "hash mismatch")

	finalizedWithoutSignatures := output
	finalizedWithoutSignatures.Finalized = true
	finalizedWithoutSignatures.CandidateOutputHash = ComputeCollatorCandidateOutputHash(finalizedWithoutSignatures)
	require.ErrorContains(t, finalizedWithoutSignatures.Validate(), "validator signatures")

	unsupported := collator
	unsupported.SupportedWorkloads = []WorkloadType{WorkloadTypeEvidenceVerification}
	_, err = BuildCollatorCandidateOutput(params, CollatorCandidateOutputInput{
		EpochID:		13,
		Collator:		unsupported,
		Task:			task,
		TransactionRoot:	PosEmptyRootHash,
		StateTransitionRoot:	PosEmptyRootHash,
		ProofBundleRoot:	PosEmptyRootHash,
	})
	require.ErrorContains(t, err, "does not support workload")

	suspended := collator
	suspended.Status = CollatorStatusSuspended
	_, err = BuildCollatorCandidateOutput(params, CollatorCandidateOutputInput{
		EpochID:		13,
		Collator:		suspended,
		Task:			task,
		TransactionRoot:	PosEmptyRootHash,
		StateTransitionRoot:	PosEmptyRootHash,
		ProofBundleRoot:	PosEmptyRootHash,
	})
	require.ErrorContains(t, err, "not eligible")
}

func TestCollatorOutputVerificationFinalizationAndInvalidEvidence(t *testing.T) {
	params := DefaultParams()
	collator, err := NewCollatorRecord(CollatorRecord{
		CollatorID:		"collator-bonded",
		OperatorAddress:	"operator-1",
		SupportedWorkloads:	[]WorkloadType{WorkloadTypeShardExecution},
		BondOptional:		sdkmath.NewInt(10_000),
		Status:			CollatorStatusActive,
		RegisteredEpoch:	13,
	})
	require.NoError(t, err)
	task := WorkloadTask{
		TaskID:			"task-1",
		WorkloadID:		"workload-1",
		WorkloadType:		WorkloadTypeShardExecution,
		ZoneID:			"zone-a",
		ShardID:		"shard-1",
		WorkloadClass:		DefaultWorkloadClass,
		RequiredValidators:	params.MinTaskGroupValidators,
		Roles:			[]ValidatorRole{ValidatorRoleCollator, ValidatorRoleVerifier},
	}
	output, err := BuildCollatorCandidateOutput(params, CollatorCandidateOutputInput{
		EpochID:		13,
		Collator:		collator,
		Task:			task,
		TransactionRoot:	PosEmptyRootHash,
		StateTransitionRoot:	PosEmptyRootHash,
		ProofBundleRoot:	PosEmptyRootHash,
	})
	require.NoError(t, err)

	validA, err := NewCollatorOutputVerification(output, "val-a", CollatorVerificationResultValid, strings.Repeat("a", PosHashHexLength), 44)
	require.NoError(t, err)
	validB, err := NewCollatorOutputVerification(output, "val-b", CollatorVerificationResultValid, strings.Repeat("b", PosHashHexLength), 44)
	require.NoError(t, err)
	abstainC, err := NewCollatorOutputVerification(output, "val-c", CollatorVerificationResultAbstain, strings.Repeat("c", PosHashHexLength), 44)
	require.NoError(t, err)
	accepted, err := VerifyCollatorOutputByValidators(output, []string{"val-a", "val-b", "val-c"}, []CollatorOutputVerification{validA, validB, abstainC}, 6_000)
	require.NoError(t, err)
	require.True(t, accepted.Accepted)
	require.False(t, accepted.Rejected)
	finalized, err := FinalizeCollatorOutputAfterVerification(output, accepted)
	require.NoError(t, err)
	require.True(t, finalized.Finalized)
	require.Equal(t, output.CandidateOutputHash, finalized.CandidateOutputHash)
	require.Len(t, finalized.ValidatorSignatures, 2)

	invalidA, err := NewCollatorOutputVerification(output, "val-a", CollatorVerificationResultInvalid, strings.Repeat("d", PosHashHexLength), 45)
	require.NoError(t, err)
	invalidB, err := NewCollatorOutputVerification(output, "val-b", CollatorVerificationResultInvalid, strings.Repeat("e", PosHashHexLength), 45)
	require.NoError(t, err)
	rejected, err := VerifyCollatorOutputByValidators(output, []string{"val-a", "val-b", "val-c"}, []CollatorOutputVerification{invalidA, invalidB, abstainC}, 6_000)
	require.NoError(t, err)
	require.False(t, rejected.Accepted)
	require.True(t, rejected.Rejected)
	_, err = FinalizeCollatorOutputAfterVerification(output, rejected)
	require.ErrorContains(t, err, "not accepted")

	evidence, err := BuildInvalidCollatorOutputEvidence("collator-evidence-1", "reporter-1", collator, output, rejected, 45)
	require.NoError(t, err)
	require.Equal(t, EvidenceTypeInvalidCollatorOutputProof, evidence.EvidenceType)
	require.Equal(t, collator.CollatorID, evidence.AccusedValidatorID)
	require.Equal(t, output.CandidateOutputHash, evidence.SubjectID)
	policy, err := DefaultEvidenceSlashPolicy(EvidenceTypeInvalidCollatorOutputProof)
	require.NoError(t, err)
	require.Equal(t, DefaultInvalidCollatorOutputSlashBps, policy.SlashFractionBps)

	evidence.Status = EvidenceStatusAccepted
	evidence.StructuredRecordHash = computeStructuredEvidenceHash(evidence)
	penalty, err := ComputeInvalidCollatorOutputPenalty(collator, evidence)
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(500), penalty)

	unbonded := collator
	unbonded.BondOptional = sdkmath.ZeroInt()
	penalty, err = ComputeInvalidCollatorOutputPenalty(unbonded, evidence)
	require.NoError(t, err)
	require.Equal(t, sdkmath.ZeroInt(), penalty)
}

func scoredCandidates(t *testing.T, params Params, candidates []Candidate) []ScoredValidator {
	t.Helper()
	validators := make([]ScoredValidator, len(candidates))
	for i, candidate := range candidates {
		scored, err := ScoreCandidate(params, candidate)
		require.NoError(t, err)
		validators[i] = scored
	}
	return validators
}

func riskWindow(stakeOwner string, validatorAddress string, amount int64, status string) RiskWindowRecord {
	window := RiskWindowRecord{
		StakeOwner:		stakeOwner,
		ValidatorAddress:	validatorAddress,
		AmountNaet:		sdkmath.NewInt(amount),
		StartEpoch:		10,
		EndEpoch:		20,
		SlashableUntilEpoch:	24,
		Status:			status,
	}
	window.RiskHistoryRoot = ComputeRiskWindowRoot(window)
	return window
}

func computeSecurity(input EconomicSecurityInput) error {
	_, err := ComputeEconomicSecurityMetrics(input)
	return err
}

func centralizationControlByID(t *testing.T, controls []CentralizationValidatorControl, validatorID string) CentralizationValidatorControl {
	t.Helper()
	for _, control := range controls {
		if control.ValidatorAddress == validatorID {
			return control
		}
	}
	t.Fatalf("centralization control %s not found", validatorID)
	return CentralizationValidatorControl{}
}

func requireAlert(t *testing.T, alerts []ConcentrationInvariantAlert, alertType string) {
	t.Helper()
	require.True(t, slices.ContainsFunc(alerts, func(alert ConcentrationInvariantAlert) bool {
		return alert.AlertType == alertType
	}), "missing concentration alert %s in %+v", alertType, alerts)
}

func posLayerSpecsByLayer(architecture LayeredPosArchitecture) map[PosLayer]PosLayerSpec {
	specs := make(map[PosLayer]PosLayerSpec, len(architecture.Layers))
	for _, layer := range architecture.Layers {
		specs[layer.Layer] = layer
	}
	return specs
}

func compatibilityExtensionsByName(manifest CosmosSDKCompatibilityManifest) map[string]CosmosSDKModuleExtension {
	out := make(map[string]CosmosSDKModuleExtension, len(manifest.Extensions))
	for _, extension := range manifest.Extensions {
		out[extension.ModuleName] = extension
	}
	return out
}

func compatibilityMiddlewareNames(manifest CosmosSDKCompatibilityManifest) []string {
	out := make([]string, len(manifest.Middleware))
	for i, middleware := range manifest.Middleware {
		out[i] = middleware.Name
	}
	return out
}

func posBoundaryModuleNames(manifest PosModuleBoundaryManifest) []string {
	out := make([]string, len(manifest.Boundaries))
	for i, boundary := range manifest.Boundaries {
		out[i] = boundary.ModuleName
	}
	return out
}

func keeperInterfacesByName(manifest KeeperIntegrationManifest) map[string]KeeperInterfaceSpec {
	out := make(map[string]KeeperInterfaceSpec, len(manifest.KeeperInterfaces))
	for _, keeper := range manifest.KeeperInterfaces {
		out[keeper.KeeperName] = keeper
	}
	return out
}

func keeperHookNames(hooks []KeeperHookSpec) []string {
	out := make([]string, len(hooks))
	for i, hook := range hooks {
		out[i] = hook.HookName
	}
	return out
}

func migrationModuleNames(handlers []MigrationHandlerSpec) []string {
	out := make([]string, len(handlers))
	for i, handler := range handlers {
		out[i] = handler.ModuleName
	}
	return out
}

func exportImportModuleNames(specs []ModuleExportImportSpec) []string {
	out := make([]string, len(specs))
	for i, spec := range specs {
		out[i] = spec.ModuleName
	}
	return out
}

func stateKeyTemplates(manifest StateModelManifest) []string {
	out := make([]string, len(manifest.Keys))
	for i, key := range manifest.Keys {
		out[i] = key.Template
	}
	return out
}

func mustStateKey(t *testing.T, build func() (string, error)) string {
	t.Helper()
	key, err := build()
	require.NoError(t, err)
	return key
}
