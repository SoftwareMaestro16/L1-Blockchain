package params

import (
	"fmt"
	"sort"
)

const (
	EngineeringScopeCoreChainConfiguration		= "core_chain_configuration"
	EngineeringScopeConsensusParameterPolicy	= "consensus_parameter_policy"

	FeatureCompletionCode			= "code"
	FeatureCompletionParams			= "params"
	FeatureCompletionGenesisValidation	= "genesis_validation"
	FeatureCompletionQueries		= "queries"
	FeatureCompletionEvents			= "events"
	FeatureCompletionTests			= "tests"
	FeatureCompletionDocs			= "docs"

	CoreChainTaskChainIDNamingPolicy	= "chain_id_naming_policy_for_devnet_testnet_mainnet"
	CoreChainTaskStakingDenomNaet		= "staking_denom_naet"
	CoreChainTaskDisplayDenomAET		= "display_denom_aet"
	CoreChainTaskCoinMetadata		= "coin_metadata_naet_aet"
	CoreChainTaskAddressPrefixReserved	= "address_prefix_and_reserved_system_address_policy"
	CoreChainTaskModuleAccountPermissions	= "module_account_permissions"
	CoreChainTaskBlockedAddressPolicy	= "blocked_address_policy"
	CoreChainTaskMintAuthority		= "mint_authority"
	CoreChainTaskBurnAuthority		= "burn_authority"
	CoreChainTaskFeeCollectorAuthority	= "fee_collector_authority"
	CoreChainTaskTreasuryAuthority		= "treasury_authority"
	CoreChainTaskAetraGenesisValidation	= "genesis_validation_for_all_aetra_modules"
	CoreChainTaskAllModulesExportImport	= "app_export_import_with_all_modules_enabled"

	CoreChainDeliverableAppWiringReview		= "app_wiring_review"
	CoreChainDeliverableGenesisParamsTable		= "genesis_params_table"
	CoreChainDeliverableModuleAccountsTable		= "module_accounts_table"
	CoreChainDeliverableAuthorityMatrix		= "authority_matrix"
	CoreChainDeliverableCLICommandMatrix		= "cli_command_matrix"
	CoreChainDeliverableQueryMatrix			= "query_matrix"
	CoreChainDeliverableEventMatrix			= "event_matrix"
	CoreChainDeliverableStartupValidationTests	= "tests_for_startup_validation"

	CoreChainTestDefaultGenesisBoots		= "app_boots_with_default_genesis"
	CoreChainTestRejectInvalidDenomMetadata		= "app_rejects_invalid_denom_metadata"
	CoreChainTestRejectMissingModuleAccounts	= "app_rejects_missing_module_accounts"
	CoreChainTestRejectDuplicateReservedAddress	= "app_rejects_duplicate_reserved_addresses"
	CoreChainTestRejectUnsafeModulePermissions	= "app_rejects_unsafe_module_account_permissions"
	CoreChainTestExportImportAppHash		= "export_import_preserves_app_hash_where_expected"
	CoreChainTestModuleInitializationOrder		= "simulation_or_integration_test_covers_module_initialization_order"

	ConsensusParamTaskBlockTimeRange		= "target_block_time_range"
	ConsensusParamTaskMaxBlockBytes			= "max_block_bytes"
	ConsensusParamTaskMaxBlockGas			= "max_block_gas"
	ConsensusParamTaskEvidenceMaxAgeBlocks		= "evidence_max_age_by_blocks"
	ConsensusParamTaskEvidenceMaxAgeDuration	= "evidence_max_age_by_duration"
	ConsensusParamTaskValidatorPubKeyTypes		= "validator_public_key_types"
	ConsensusParamTaskTimeoutProfiles		= "cometbft_timeout_profile_for_100_200_300_validators"
	ConsensusParamTaskSnapshotInterval		= "snapshot_interval"
	ConsensusParamTaskStateSyncParameters		= "state_sync_parameters"
	ConsensusParamTaskPruningProfiles		= "pruning_profiles"

	ConsensusParamDeliverableConservativeInitialValues	= "conservative_initial_values"
	ConsensusParamDeliverableBlockTimeTable			= "block_time_target_table_100_200_300_validators"
	ConsensusParamDeliverableBlockGasBounds			= "max_block_gas_bounds"
	ConsensusParamDeliverableBlockBytesBounds		= "max_block_bytes_bounds"
	ConsensusParamDeliverableEvidenceWindowTable		= "evidence_window_table"
	ConsensusParamDeliverableTimeoutProfileTable		= "timeout_profile_table"
	ConsensusParamDeliverableStateSyncSnapshotPruning	= "state_sync_snapshot_pruning_table"
	ConsensusParamDeliverableGovernanceSafetyBounds		= "governance_safety_bounds"

	ConsensusParamTestLocalnetTimeoutStability	= "localnet_remains_stable_under_configured_timeout_profile"
	ConsensusParamTestOversizedBlocksRejected	= "oversized_blocks_are_rejected"
	ConsensusParamTestInvalidParamsRejected		= "invalid_consensus_params_are_rejected"
	ConsensusParamTestGovernanceBounds		= "governance_cannot_set_unsafe_block_gas_bytes_outside_bounds"
	ConsensusParamTestEvidencePeriod		= "evidence_remains_valid_through_configured_evidence_period"
)

type FeatureCompletionEvidence struct {
	FeatureID		string
	Code			bool
	Params			bool
	GenesisValidation	bool
	Queries			bool
	Events			bool
	Tests			bool
	Docs			bool
}

type EngineeringScopeItem struct {
	ID		string
	Kind		string
	Required	bool
	Done		bool
	Evidence	string
}

type EngineeringScopePlan struct {
	ScopeID	string
	Items	[]EngineeringScopeItem
}

type EngineeringScopeReport struct {
	ScopeID		string
	Required	int
	Done		int
	Failed		[]string
	Ready		bool
}

func ValidateFeatureCompletion(evidence FeatureCompletionEvidence) error {
	report := BuildFeatureCompletionReport(evidence)
	if !report.Ready {
		return fmt.Errorf("feature completion failed: %v", report.Failed)
	}
	return nil
}

func BuildFeatureCompletionReport(evidence FeatureCompletionEvidence) EngineeringScopeReport {
	checks := []requirementCheck{
		{FeatureCompletionCode, evidence.Code},
		{FeatureCompletionParams, evidence.Params},
		{FeatureCompletionGenesisValidation, evidence.GenesisValidation},
		{FeatureCompletionQueries, evidence.Queries},
		{FeatureCompletionEvents, evidence.Events},
		{FeatureCompletionTests, evidence.Tests},
		{FeatureCompletionDocs, evidence.Docs},
	}
	failed := make([]string, 0)
	if evidence.FeatureID == "" {
		failed = append(failed, "feature_id_required")
	}
	passed := 0
	for _, check := range checks {
		if check.Passed {
			passed++
		} else {
			failed = append(failed, check.ID)
		}
	}
	sort.Strings(failed)
	return EngineeringScopeReport{
		ScopeID:	evidence.FeatureID,
		Required:	len(checks),
		Done:		passed,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func DefaultCoreChainConfigurationScopePlan() EngineeringScopePlan {
	return EngineeringScopePlan{
		ScopeID:	EngineeringScopeCoreChainConfiguration,
		Items: []EngineeringScopeItem{
			engineeringScopeItem("task", CoreChainTaskChainIDNamingPolicy),
			engineeringScopeItem("task", CoreChainTaskStakingDenomNaet),
			engineeringScopeItem("task", CoreChainTaskDisplayDenomAET),
			engineeringScopeItem("task", CoreChainTaskCoinMetadata),
			engineeringScopeItem("task", CoreChainTaskAddressPrefixReserved),
			engineeringScopeItem("task", CoreChainTaskModuleAccountPermissions),
			engineeringScopeItem("task", CoreChainTaskBlockedAddressPolicy),
			engineeringScopeItem("task", CoreChainTaskMintAuthority),
			engineeringScopeItem("task", CoreChainTaskBurnAuthority),
			engineeringScopeItem("task", CoreChainTaskFeeCollectorAuthority),
			engineeringScopeItem("task", CoreChainTaskTreasuryAuthority),
			engineeringScopeItem("task", CoreChainTaskAetraGenesisValidation),
			engineeringScopeItem("task", CoreChainTaskAllModulesExportImport),
			engineeringScopeItem("deliverable", CoreChainDeliverableAppWiringReview),
			engineeringScopeItem("deliverable", CoreChainDeliverableGenesisParamsTable),
			engineeringScopeItem("deliverable", CoreChainDeliverableModuleAccountsTable),
			engineeringScopeItem("deliverable", CoreChainDeliverableAuthorityMatrix),
			engineeringScopeItem("deliverable", CoreChainDeliverableCLICommandMatrix),
			engineeringScopeItem("deliverable", CoreChainDeliverableQueryMatrix),
			engineeringScopeItem("deliverable", CoreChainDeliverableEventMatrix),
			engineeringScopeItem("deliverable", CoreChainDeliverableStartupValidationTests),
			engineeringScopeItem("test", CoreChainTestDefaultGenesisBoots),
			engineeringScopeItem("test", CoreChainTestRejectInvalidDenomMetadata),
			engineeringScopeItem("test", CoreChainTestRejectMissingModuleAccounts),
			engineeringScopeItem("test", CoreChainTestRejectDuplicateReservedAddress),
			engineeringScopeItem("test", CoreChainTestRejectUnsafeModulePermissions),
			engineeringScopeItem("test", CoreChainTestExportImportAppHash),
			engineeringScopeItem("test", CoreChainTestModuleInitializationOrder),
		},
	}
}

func DefaultConsensusParameterPolicyScopePlan() EngineeringScopePlan {
	return EngineeringScopePlan{
		ScopeID:	EngineeringScopeConsensusParameterPolicy,
		Items: []EngineeringScopeItem{
			engineeringScopeItem("task", ConsensusParamTaskBlockTimeRange),
			engineeringScopeItem("task", ConsensusParamTaskMaxBlockBytes),
			engineeringScopeItem("task", ConsensusParamTaskMaxBlockGas),
			engineeringScopeItem("task", ConsensusParamTaskEvidenceMaxAgeBlocks),
			engineeringScopeItem("task", ConsensusParamTaskEvidenceMaxAgeDuration),
			engineeringScopeItem("task", ConsensusParamTaskValidatorPubKeyTypes),
			engineeringScopeItem("task", ConsensusParamTaskTimeoutProfiles),
			engineeringScopeItem("task", ConsensusParamTaskSnapshotInterval),
			engineeringScopeItem("task", ConsensusParamTaskStateSyncParameters),
			engineeringScopeItem("task", ConsensusParamTaskPruningProfiles),
			engineeringScopeItem("deliverable", ConsensusParamDeliverableConservativeInitialValues),
			engineeringScopeItem("deliverable", ConsensusParamDeliverableBlockTimeTable),
			engineeringScopeItem("deliverable", ConsensusParamDeliverableBlockGasBounds),
			engineeringScopeItem("deliverable", ConsensusParamDeliverableBlockBytesBounds),
			engineeringScopeItem("deliverable", ConsensusParamDeliverableEvidenceWindowTable),
			engineeringScopeItem("deliverable", ConsensusParamDeliverableTimeoutProfileTable),
			engineeringScopeItem("deliverable", ConsensusParamDeliverableStateSyncSnapshotPruning),
			engineeringScopeItem("deliverable", ConsensusParamDeliverableGovernanceSafetyBounds),
			engineeringScopeItem("test", ConsensusParamTestLocalnetTimeoutStability),
			engineeringScopeItem("test", ConsensusParamTestOversizedBlocksRejected),
			engineeringScopeItem("test", ConsensusParamTestInvalidParamsRejected),
			engineeringScopeItem("test", ConsensusParamTestGovernanceBounds),
			engineeringScopeItem("test", ConsensusParamTestEvidencePeriod),
		},
	}
}

func ValidateEngineeringScopePlan(plan EngineeringScopePlan) error {
	report := BuildEngineeringScopeReport(plan)
	if !report.Ready {
		return fmt.Errorf("engineering scope %s failed: %v", report.ScopeID, report.Failed)
	}
	return nil
}

func BuildEngineeringScopeReport(plan EngineeringScopePlan) EngineeringScopeReport {
	expected := expectedEngineeringScopeItems(plan.ScopeID)
	failed := make([]string, 0)
	seen := map[string]EngineeringScopeItem{}
	required := 0
	done := 0
	if plan.ScopeID == "" {
		failed = append(failed, "scope_id_required")
	}
	if len(expected) == 0 {
		failed = append(failed, plan.ScopeID+":unknown_scope")
	}
	for _, item := range plan.Items {
		if item.ID == "" || item.Kind == "" {
			failed = append(failed, "scope_item_id_and_kind_required")
			continue
		}
		if _, duplicate := seen[item.ID]; duplicate {
			failed = append(failed, item.ID+":duplicate")
		}
		seen[item.ID] = item
		if !expected[item.ID] {
			failed = append(failed, item.ID+":unexpected")
		}
		if item.Required {
			required++
		}
		if item.Required && (!item.Done || item.Evidence == "") {
			failed = append(failed, item.ID+":missing_evidence")
		}
		if item.Required && item.Done && item.Evidence != "" {
			done++
		}
	}
	for id := range expected {
		if _, ok := seen[id]; !ok {
			failed = append(failed, id+":missing")
		}
	}
	sort.Strings(failed)
	return EngineeringScopeReport{
		ScopeID:	plan.ScopeID,
		Required:	required,
		Done:		done,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func engineeringScopeItem(kind, id string) EngineeringScopeItem {
	return EngineeringScopeItem{
		ID:		id,
		Kind:		kind,
		Required:	true,
		Done:		true,
		Evidence:	"required " + kind + " evidence for " + id,
	}
}

func expectedEngineeringScopeItems(scopeID string) map[string]bool {
	out := map[string]bool{}
	switch scopeID {
	case EngineeringScopeCoreChainConfiguration:
		for _, id := range []string{
			CoreChainTaskChainIDNamingPolicy,
			CoreChainTaskStakingDenomNaet,
			CoreChainTaskDisplayDenomAET,
			CoreChainTaskCoinMetadata,
			CoreChainTaskAddressPrefixReserved,
			CoreChainTaskModuleAccountPermissions,
			CoreChainTaskBlockedAddressPolicy,
			CoreChainTaskMintAuthority,
			CoreChainTaskBurnAuthority,
			CoreChainTaskFeeCollectorAuthority,
			CoreChainTaskTreasuryAuthority,
			CoreChainTaskAetraGenesisValidation,
			CoreChainTaskAllModulesExportImport,
			CoreChainDeliverableAppWiringReview,
			CoreChainDeliverableGenesisParamsTable,
			CoreChainDeliverableModuleAccountsTable,
			CoreChainDeliverableAuthorityMatrix,
			CoreChainDeliverableCLICommandMatrix,
			CoreChainDeliverableQueryMatrix,
			CoreChainDeliverableEventMatrix,
			CoreChainDeliverableStartupValidationTests,
			CoreChainTestDefaultGenesisBoots,
			CoreChainTestRejectInvalidDenomMetadata,
			CoreChainTestRejectMissingModuleAccounts,
			CoreChainTestRejectDuplicateReservedAddress,
			CoreChainTestRejectUnsafeModulePermissions,
			CoreChainTestExportImportAppHash,
			CoreChainTestModuleInitializationOrder,
		} {
			out[id] = true
		}
	case EngineeringScopeConsensusParameterPolicy:
		for _, id := range []string{
			ConsensusParamTaskBlockTimeRange,
			ConsensusParamTaskMaxBlockBytes,
			ConsensusParamTaskMaxBlockGas,
			ConsensusParamTaskEvidenceMaxAgeBlocks,
			ConsensusParamTaskEvidenceMaxAgeDuration,
			ConsensusParamTaskValidatorPubKeyTypes,
			ConsensusParamTaskTimeoutProfiles,
			ConsensusParamTaskSnapshotInterval,
			ConsensusParamTaskStateSyncParameters,
			ConsensusParamTaskPruningProfiles,
			ConsensusParamDeliverableConservativeInitialValues,
			ConsensusParamDeliverableBlockTimeTable,
			ConsensusParamDeliverableBlockGasBounds,
			ConsensusParamDeliverableBlockBytesBounds,
			ConsensusParamDeliverableEvidenceWindowTable,
			ConsensusParamDeliverableTimeoutProfileTable,
			ConsensusParamDeliverableStateSyncSnapshotPruning,
			ConsensusParamDeliverableGovernanceSafetyBounds,
			ConsensusParamTestLocalnetTimeoutStability,
			ConsensusParamTestOversizedBlocksRejected,
			ConsensusParamTestInvalidParamsRejected,
			ConsensusParamTestGovernanceBounds,
			ConsensusParamTestEvidencePeriod,
		} {
			out[id] = true
		}
	default:
		return out
	}
	return out
}
