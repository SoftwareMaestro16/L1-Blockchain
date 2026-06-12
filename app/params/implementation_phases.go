package params

import (
	"fmt"
	"sort"
)

const (
	ImplementationPhaseBaselineAudit	= "phase_0_baseline_audit"
	ImplementationPhaseStakingPolicyCap	= "phase_1_staking_policy_validator_cap"
	ImplementationPhaseEconomicsFeeSplit	= "phase_2_economics_fee_split"
	ImplementationPhaseValidatorScore	= "phase_3_validator_score_accountability"
	ImplementationPhaseSlashingHardening	= "phase_4_slashing_hardening"
	ImplementationPhaseAVMIntegration	= "phase_5_avm_integration"
	ImplementationPhaseFinalityPerformance	= "phase_6_finality_performance_profile"
	ImplementationPhasePublicTestnetReady	= "phase_7_public_testnet_readiness"

	PhaseTaskInspectVersions			= "inspect_current_cosmos_sdk_and_cometbft_versions"
	PhaseTaskDocumentModuleGraph			= "document_current_app_module_graph"
	PhaseTaskIdentifyOverlappingModules		= "identify_modules_overlapping_custom_aetra_modules"
	PhaseTaskDecideRenameReuseWrap			= "decide_modules_renamed_reused_or_wrapped"
	PhaseTaskVerifyNaetStakingDenom			= "verify_naet_staking_denom"
	PhaseTaskVerifyEconomyWiring			= "verify_fee_collector_burn_treasury_emissions_mint_authority_wiring"
	PhaseTaskVerifyLocalnetAndCoverage		= "verify_localnet_scripts_and_test_coverage"
	PhaseTaskImplementEffectivePowerCap		= "implement_effective_voting_power_cap"
	PhaseTaskImplementOverflowAccounting		= "implement_overflow_stake_accounting"
	PhaseTaskImplementCommissionPolicy		= "implement_commission_floor_max_change_policy"
	PhaseTaskAddConcentrationMetrics		= "add_concentration_metrics"
	PhaseTaskAddStakeQueries			= "add_validator_raw_effective_overflow_queries"
	PhaseTaskAddGovernanceParams			= "add_governance_params_with_validation"
	PhaseTaskWireModuleLifecycle			= "wire_module_into_app_lifecycle"
	PhaseTaskImplementInflationBounds		= "implement_dynamic_inflation_bounds"
	PhaseTaskImplementTargetBondedRatio		= "implement_target_bonded_ratio_logic"
	PhaseTaskImplementFeeSplit			= "implement_fee_split_to_burn_rewards_treasury"
	PhaseTaskImplementRewardSmoothing		= "implement_reward_smoothing"
	PhaseTaskExposeAPREstimateQuery			= "expose_apr_estimate_query"
	PhaseTaskExposeSupplyTreasuryQueries		= "expose_burned_supply_and_treasury_accounting_queries"
	PhaseTaskAddEconomicsGovernanceParams		= "add_economics_governance_param_controls"
	PhaseTaskImplementUptimeScore			= "implement_uptime_score"
	PhaseTaskImplementSlashHistory			= "implement_slash_history"
	PhaseTaskImplementGovernanceScore		= "implement_governance_participation_score"
	PhaseTaskImplementDecentralizationScore		= "implement_decentralization_score"
	PhaseTaskImplementValidatorMetricQueries	= "implement_public_validator_metrics_queries"
	PhaseTaskIntegrateObjectiveRewardModifier	= "integrate_score_with_reward_modifier_only_for_objective_inputs"
	PhaseTaskConfigureDoubleSignSlashTombstone	= "configure_double_sign_slash_fraction_and_tombstone_behavior"
	PhaseTaskConfigureDowntimeJail			= "configure_downtime_windows_and_jail_duration"
	PhaseTaskImplementProgressiveDowntime		= "implement_progressive_downtime_if_not_covered_by_standard_module"
	PhaseTaskAddObjectiveTimestampProposalPolicy	= "add_timestamp_proposal_violation_policy_where_objective"
	PhaseTaskDocumentEvidenceLifecycle		= "document_evidence_lifecycle_and_unbonding_interaction"
	PhaseTaskFinalizeAVMWiring			= "finalize_avm_module_wiring"
	PhaseTaskDefineCodeUploadPolicy			= "define_code_upload_policy"
	PhaseTaskDefineContractGasLimits		= "define_contract_gas_limits"
	PhaseTaskDefineContractSizeLimits		= "define_contract_size_limits"
	PhaseTaskIntegrateStorageRentPricing		= "integrate_storage_rent_or_storage_pricing"
	PhaseTaskExposeContractIndexerEvents		= "expose_contract_events_for_indexers"
	PhaseTaskDocumentContractDeveloperFlow		= "document_contract_developer_flow"
	PhaseTaskConfigureBlockTimeTargets		= "configure_block_time_targets"
	PhaseTaskConfigureBlockSizeGasLimits		= "configure_block_size_gas_limits"
	PhaseTaskProfile100ValidatorLocalnet		= "profile_100_validator_localnet"
	PhaseTaskProfile150To200Validators		= "profile_150_200_validator_scenario_if_feasible"
	PhaseTaskEstimate250To300Requirements		= "estimate_250_300_validator_operational_requirements"
	PhaseTaskMeasureFinalityUnderLoad		= "measure_finality_under_load"
	PhaseTaskMeasureFinalityUnderValidatorFailure	= "measure_finality_under_partial_validator_failure"
	PhaseTaskWriteValidatorSetupDocs		= "write_validator_setup_docs"
	PhaseTaskWriteSentryArchitectureDocs		= "write_sentry_architecture_docs"
	PhaseTaskWriteMonitoringDocs			= "write_monitoring_docs"
	PhaseTaskPublishGenesisParamExplanation		= "publish_genesis_parameter_explanation"
	PhaseTaskPublishEconomicModelExplanation	= "publish_economic_model_explanation"
	PhaseTaskPublishSlashingRiskExplanation		= "publish_slashing_risk_explanation"
	PhaseTaskPublishDelegationPoolGuide		= "publish_delegation_and_nomination_pool_guide"
	PhaseTaskPublishAVMDeveloperGuide		= "publish_avm_developer_guide"
	PhaseTaskPreparePublicDashboards		= "prepare_public_dashboards"
	PhaseTaskPrepareIncidentResponseProcess		= "prepare_incident_response_process"

	PhaseDeliverableModuleInventory		= "module_inventory"
	PhaseDeliverableGapAnalysis		= "gap_analysis"
	PhaseDeliverableRiskList		= "risk_list"
	PhaseDeliverableImplementationChecklist	= "updated_implementation_checklist"

	PhaseTestFullUnitRun			= "current_full_unit_test_run"
	PhaseTestIntegrationRun			= "current_integration_test_run"
	PhaseTestLocalnetSmoke			= "current_localnet_smoke_test"
	PhaseTestExportImport			= "current_export_import_test"
	PhaseTestCapMathUnit			= "cap_math_unit_tests"
	PhaseTestValidatorSetTransition		= "validator_set_transition_tests"
	PhaseTestConcentrationQuery		= "concentration_query_tests"
	PhaseTestCommissionBounds		= "commission_bounds_tests"
	PhaseTestStakingIntegration		= "integration_tests_with_staking"
	PhaseTestStakingExportImport		= "staking_policy_export_import_tests"
	PhaseTestInvariant			= "invariant_tests"
	PhaseTestInflationCurve			= "inflation_curve_tests"
	PhaseTestBondedRatio			= "bonded_ratio_tests"
	PhaseTestFeeSplit			= "fee_split_tests"
	PhaseTestBurnAccounting			= "burn_accounting_tests"
	PhaseTestTreasuryAccounting		= "treasury_accounting_tests"
	PhaseTestAPRQuery			= "apr_query_tests"
	PhaseTestSupplyInvariant		= "supply_invariant_tests"
	PhaseTestEconomicsExportImport		= "economics_export_import_tests"
	PhaseTestUptimeWindow			= "uptime_window_tests"
	PhaseTestMissedBlock			= "missed_block_tests"
	PhaseTestSlashHistory			= "slash_history_tests"
	PhaseTestGovernanceParticipation	= "governance_participation_tests"
	PhaseTestScoreDeterminism		= "score_determinism_tests"
	PhaseTestRewardModifier			= "reward_modifier_tests"
	PhaseTestValidatorScoreExportImport	= "validator_score_export_import_tests"
	PhaseTestDoubleSignEvidence		= "double_sign_evidence_tests_where_feasible"
	PhaseTestDowntime			= "downtime_tests"
	PhaseTestJailUnjail			= "jail_unjail_tests"
	PhaseTestProgressiveDowntime		= "progressive_downtime_tests"
	PhaseTestSlashingAccounting		= "slashing_accounting_tests"
	PhaseTestDelegatorLoss			= "delegator_loss_tests"
	PhaseTestTombstone			= "tombstone_tests"
	PhaseTestEvidenceExpiry			= "evidence_expiry_tests"
	PhaseTestContractTxFlow			= "instantiate_execute_query_tests"
	PhaseTestContractMigration		= "migration_tests"
	PhaseTestContractGasLimit		= "gas_limit_tests"
	PhaseTestContractStorageLimitRent	= "storage_limit_rent_tests"
	PhaseTestMaliciousContract		= "malicious_contract_tests"
	PhaseTestContractExportImport		= "export_import_tests_with_contracts"
	PhaseTestLocalnetAVMSmoke		= "localnet_avm_smoke_test"
	PhaseTestLocalnetLoadProfile		= "localnet_load_profile"
	PhaseTestMempoolPressure		= "mempool_pressure_test"
	PhaseTestBlockTimeMeasurement		= "block_time_measurement"
	PhaseTestFinalityMeasurement		= "finality_measurement"
	PhaseTestValidatorFailureScenario	= "validator_failure_scenario"
	PhaseTestRestartScenario		= "restart_scenario"
	PhaseTestStateSyncSnapshotScenario	= "state_sync_snapshot_scenario"
	PhaseTestCleanNodeBootstrapDocs		= "clean_node_bootstrap_from_docs"
	PhaseTestValidatorJoinDocs		= "validator_join_from_docs"
	PhaseTestSnapshotRestoreDocs		= "snapshot_restore_from_docs"
	PhaseTestStateSyncDocs			= "state_sync_from_docs"
	PhaseTestTxFlowSmoke			= "tx_flow_smoke_tests"
	PhaseTestGovernanceProposalSmoke	= "governance_proposal_smoke_tests"
	PhaseTestPublicEndpointSmoke		= "public_rpc_grpc_rest_smoke_tests"

	PhaseAcceptanceNoValidatorExceedsCap		= "no_validator_can_exceed_effective_power_cap"
	PhaseAcceptanceExcessNoVotingPower		= "excess_stake_does_not_increase_voting_power"
	PhaseAcceptanceParamsSafeBounds			= "params_cannot_be_set_outside_safe_bounds"
	PhaseAcceptanceDeterministicExportImport	= "state_remains_deterministic_after_export_import"
	PhaseAcceptanceInflationWithinBounds		= "inflation_remains_within_configured_bounds"
	PhaseAcceptanceFeeSplitSumsToFullAmount		= "fee_split_sums_to_100_percent"
	PhaseAcceptanceBurnReducesSupply		= "burned_fees_reduce_supply_according_to_chain_accounting"
	PhaseAcceptanceTreasuryReceivesAmount		= "treasury_receives_correct_amount"
	PhaseAcceptanceRewardsDeterministic		= "rewards_are_deterministic"
	PhaseAcceptanceScoreDeterministic		= "score_is_deterministic"
	PhaseAcceptanceScoreObjectiveOnly		= "score_cannot_be_manipulated_through_subjective_inputs"
	PhaseAcceptanceScoreQueryable			= "score_is_queryable_for_explorers_and_wallets"
	PhaseAcceptanceScoreConsensusSafe		= "score_does_not_break_consensus_safety"
	PhaseAcceptanceDoubleSignTombstone		= "double_sign_leads_to_severe_slash_and_permanent_tombstone"
	PhaseAcceptanceDowntimeBoundedProgressive	= "downtime_penalties_are_bounded_and_progressive"
	PhaseAcceptanceNoSubjectiveSlashing		= "no_subjective_slashing_path_exists"
	PhaseAcceptanceSlashingStakeShareSafe		= "slashing_cannot_underflow_stake_or_corrupt_shares"
	PhaseAcceptanceContractsDeterministic		= "contracts_are_deterministic"
	PhaseAcceptanceContractGasBounded		= "contract_gas_is_bounded"
	PhaseAcceptanceMaliciousContractsSafe		= "malicious_contracts_cannot_halt_chain"
	PhaseAcceptanceContractStateExportImport	= "contract_state_survives_export_import"
	PhaseAcceptanceNormalFinalityWithinTarget	= "normal_finality_remains_within_target"
	PhaseAcceptanceStressedFinalityUnderLimit	= "stressed_finality_remains_below_120_seconds_in_healthy_majority_scenario"
	PhaseAcceptanceMediumNodeRequirements		= "node_requirements_remain_medium_level"
	PhaseAcceptanceNoExcessiveConsensusPayloads	= "no_excessive_consensus_payloads"
	PhaseAcceptanceValidatorCanJoinFromDocs		= "new_validator_can_join_using_docs"
	PhaseAcceptancePublicEndpointsObservable	= "public_endpoints_are_observable"
	PhaseAcceptanceNetworkRecoversRestarts		= "network_can_recover_from_node_restarts"
	PhaseAcceptanceCoreFlowsEndToEnd		= "core_economic_and_staking_flows_work_end_to_end"
)

type ImplementationPhaseItem struct {
	ID		string
	Kind		string
	Required	bool
	Done		bool
	Evidence	string
}

type ImplementationPhasePlan struct {
	PhaseID	string
	Items	[]ImplementationPhaseItem
}

type ImplementationPhaseReport struct {
	PhaseID		string
	Required	int
	Done		int
	Failed		[]string
	Ready		bool
}

func DefaultImplementationPhasePlans() []ImplementationPhasePlan {
	return []ImplementationPhasePlan{
		{
			PhaseID:	ImplementationPhaseBaselineAudit,
			Items: []ImplementationPhaseItem{
				phaseItem("task", PhaseTaskInspectVersions),
				phaseItem("task", PhaseTaskDocumentModuleGraph),
				phaseItem("task", PhaseTaskIdentifyOverlappingModules),
				phaseItem("task", PhaseTaskDecideRenameReuseWrap),
				phaseItem("task", PhaseTaskVerifyNaetStakingDenom),
				phaseItem("task", PhaseTaskVerifyEconomyWiring),
				phaseItem("task", PhaseTaskVerifyLocalnetAndCoverage),
				phaseItem("deliverable", PhaseDeliverableModuleInventory),
				phaseItem("deliverable", PhaseDeliverableGapAnalysis),
				phaseItem("deliverable", PhaseDeliverableRiskList),
				phaseItem("deliverable", PhaseDeliverableImplementationChecklist),
				phaseItem("test", PhaseTestFullUnitRun),
				phaseItem("test", PhaseTestIntegrationRun),
				phaseItem("test", PhaseTestLocalnetSmoke),
				phaseItem("test", PhaseTestExportImport),
			},
		},
		{
			PhaseID:	ImplementationPhaseStakingPolicyCap,
			Items: []ImplementationPhaseItem{
				phaseItem("task", PhaseTaskImplementEffectivePowerCap),
				phaseItem("task", PhaseTaskImplementOverflowAccounting),
				phaseItem("task", PhaseTaskImplementCommissionPolicy),
				phaseItem("task", PhaseTaskAddConcentrationMetrics),
				phaseItem("task", PhaseTaskAddStakeQueries),
				phaseItem("task", PhaseTaskAddGovernanceParams),
				phaseItem("task", PhaseTaskWireModuleLifecycle),
				phaseItem("test", PhaseTestCapMathUnit),
				phaseItem("test", PhaseTestValidatorSetTransition),
				phaseItem("test", PhaseTestConcentrationQuery),
				phaseItem("test", PhaseTestCommissionBounds),
				phaseItem("test", PhaseTestStakingIntegration),
				phaseItem("test", PhaseTestStakingExportImport),
				phaseItem("test", PhaseTestInvariant),
				phaseItem("acceptance", PhaseAcceptanceNoValidatorExceedsCap),
				phaseItem("acceptance", PhaseAcceptanceExcessNoVotingPower),
				phaseItem("acceptance", PhaseAcceptanceParamsSafeBounds),
				phaseItem("acceptance", PhaseAcceptanceDeterministicExportImport),
			},
		},
		{
			PhaseID:	ImplementationPhaseEconomicsFeeSplit,
			Items: []ImplementationPhaseItem{
				phaseItem("task", PhaseTaskImplementInflationBounds),
				phaseItem("task", PhaseTaskImplementTargetBondedRatio),
				phaseItem("task", PhaseTaskImplementFeeSplit),
				phaseItem("task", PhaseTaskImplementRewardSmoothing),
				phaseItem("task", PhaseTaskExposeAPREstimateQuery),
				phaseItem("task", PhaseTaskExposeSupplyTreasuryQueries),
				phaseItem("task", PhaseTaskAddEconomicsGovernanceParams),
				phaseItem("test", PhaseTestInflationCurve),
				phaseItem("test", PhaseTestBondedRatio),
				phaseItem("test", PhaseTestFeeSplit),
				phaseItem("test", PhaseTestBurnAccounting),
				phaseItem("test", PhaseTestTreasuryAccounting),
				phaseItem("test", PhaseTestAPRQuery),
				phaseItem("test", PhaseTestSupplyInvariant),
				phaseItem("test", PhaseTestEconomicsExportImport),
				phaseItem("acceptance", PhaseAcceptanceInflationWithinBounds),
				phaseItem("acceptance", PhaseAcceptanceFeeSplitSumsToFullAmount),
				phaseItem("acceptance", PhaseAcceptanceBurnReducesSupply),
				phaseItem("acceptance", PhaseAcceptanceTreasuryReceivesAmount),
				phaseItem("acceptance", PhaseAcceptanceRewardsDeterministic),
			},
		},
		{
			PhaseID:	ImplementationPhaseValidatorScore,
			Items: []ImplementationPhaseItem{
				phaseItem("task", PhaseTaskImplementUptimeScore),
				phaseItem("task", PhaseTaskImplementSlashHistory),
				phaseItem("task", PhaseTaskImplementGovernanceScore),
				phaseItem("task", PhaseTaskImplementDecentralizationScore),
				phaseItem("task", PhaseTaskImplementValidatorMetricQueries),
				phaseItem("task", PhaseTaskIntegrateObjectiveRewardModifier),
				phaseItem("test", PhaseTestUptimeWindow),
				phaseItem("test", PhaseTestMissedBlock),
				phaseItem("test", PhaseTestSlashHistory),
				phaseItem("test", PhaseTestGovernanceParticipation),
				phaseItem("test", PhaseTestScoreDeterminism),
				phaseItem("test", PhaseTestRewardModifier),
				phaseItem("test", PhaseTestValidatorScoreExportImport),
				phaseItem("acceptance", PhaseAcceptanceScoreDeterministic),
				phaseItem("acceptance", PhaseAcceptanceScoreObjectiveOnly),
				phaseItem("acceptance", PhaseAcceptanceScoreQueryable),
				phaseItem("acceptance", PhaseAcceptanceScoreConsensusSafe),
			},
		},
		{
			PhaseID:	ImplementationPhaseSlashingHardening,
			Items: []ImplementationPhaseItem{
				phaseItem("task", PhaseTaskConfigureDoubleSignSlashTombstone),
				phaseItem("task", PhaseTaskConfigureDowntimeJail),
				phaseItem("task", PhaseTaskImplementProgressiveDowntime),
				phaseItem("task", PhaseTaskAddObjectiveTimestampProposalPolicy),
				phaseItem("task", PhaseTaskDocumentEvidenceLifecycle),
				phaseItem("test", PhaseTestDoubleSignEvidence),
				phaseItem("test", PhaseTestDowntime),
				phaseItem("test", PhaseTestJailUnjail),
				phaseItem("test", PhaseTestProgressiveDowntime),
				phaseItem("test", PhaseTestSlashingAccounting),
				phaseItem("test", PhaseTestDelegatorLoss),
				phaseItem("test", PhaseTestTombstone),
				phaseItem("test", PhaseTestEvidenceExpiry),
				phaseItem("acceptance", PhaseAcceptanceDoubleSignTombstone),
				phaseItem("acceptance", PhaseAcceptanceDowntimeBoundedProgressive),
				phaseItem("acceptance", PhaseAcceptanceNoSubjectiveSlashing),
				phaseItem("acceptance", PhaseAcceptanceSlashingStakeShareSafe),
			},
		},
		{
			PhaseID:	ImplementationPhaseAVMIntegration,
			Items: []ImplementationPhaseItem{
				phaseItem("task", PhaseTaskFinalizeAVMWiring),
				phaseItem("task", PhaseTaskDefineCodeUploadPolicy),
				phaseItem("task", PhaseTaskDefineContractGasLimits),
				phaseItem("task", PhaseTaskDefineContractSizeLimits),
				phaseItem("task", PhaseTaskIntegrateStorageRentPricing),
				phaseItem("task", PhaseTaskExposeContractIndexerEvents),
				phaseItem("task", PhaseTaskDocumentContractDeveloperFlow),
				phaseItem("test", PhaseTestContractTxFlow),
				phaseItem("test", PhaseTestContractMigration),
				phaseItem("test", PhaseTestContractGasLimit),
				phaseItem("test", PhaseTestContractStorageLimitRent),
				phaseItem("test", PhaseTestMaliciousContract),
				phaseItem("test", PhaseTestContractExportImport),
				phaseItem("test", PhaseTestLocalnetAVMSmoke),
				phaseItem("acceptance", PhaseAcceptanceContractsDeterministic),
				phaseItem("acceptance", PhaseAcceptanceContractGasBounded),
				phaseItem("acceptance", PhaseAcceptanceMaliciousContractsSafe),
				phaseItem("acceptance", PhaseAcceptanceContractStateExportImport),
			},
		},
		{
			PhaseID:	ImplementationPhaseFinalityPerformance,
			Items: []ImplementationPhaseItem{
				phaseItem("task", PhaseTaskConfigureBlockTimeTargets),
				phaseItem("task", PhaseTaskConfigureBlockSizeGasLimits),
				phaseItem("task", PhaseTaskProfile100ValidatorLocalnet),
				phaseItem("task", PhaseTaskProfile150To200Validators),
				phaseItem("task", PhaseTaskEstimate250To300Requirements),
				phaseItem("task", PhaseTaskMeasureFinalityUnderLoad),
				phaseItem("task", PhaseTaskMeasureFinalityUnderValidatorFailure),
				phaseItem("test", PhaseTestLocalnetLoadProfile),
				phaseItem("test", PhaseTestMempoolPressure),
				phaseItem("test", PhaseTestBlockTimeMeasurement),
				phaseItem("test", PhaseTestFinalityMeasurement),
				phaseItem("test", PhaseTestValidatorFailureScenario),
				phaseItem("test", PhaseTestRestartScenario),
				phaseItem("test", PhaseTestStateSyncSnapshotScenario),
				phaseItem("acceptance", PhaseAcceptanceNormalFinalityWithinTarget),
				phaseItem("acceptance", PhaseAcceptanceStressedFinalityUnderLimit),
				phaseItem("acceptance", PhaseAcceptanceMediumNodeRequirements),
				phaseItem("acceptance", PhaseAcceptanceNoExcessiveConsensusPayloads),
			},
		},
		{
			PhaseID:	ImplementationPhasePublicTestnetReady,
			Items: []ImplementationPhaseItem{
				phaseItem("task", PhaseTaskWriteValidatorSetupDocs),
				phaseItem("task", PhaseTaskWriteSentryArchitectureDocs),
				phaseItem("task", PhaseTaskWriteMonitoringDocs),
				phaseItem("task", PhaseTaskPublishGenesisParamExplanation),
				phaseItem("task", PhaseTaskPublishEconomicModelExplanation),
				phaseItem("task", PhaseTaskPublishSlashingRiskExplanation),
				phaseItem("task", PhaseTaskPublishDelegationPoolGuide),
				phaseItem("task", PhaseTaskPublishAVMDeveloperGuide),
				phaseItem("task", PhaseTaskPreparePublicDashboards),
				phaseItem("task", PhaseTaskPrepareIncidentResponseProcess),
				phaseItem("test", PhaseTestCleanNodeBootstrapDocs),
				phaseItem("test", PhaseTestValidatorJoinDocs),
				phaseItem("test", PhaseTestSnapshotRestoreDocs),
				phaseItem("test", PhaseTestStateSyncDocs),
				phaseItem("test", PhaseTestTxFlowSmoke),
				phaseItem("test", PhaseTestGovernanceProposalSmoke),
				phaseItem("test", PhaseTestPublicEndpointSmoke),
				phaseItem("acceptance", PhaseAcceptanceValidatorCanJoinFromDocs),
				phaseItem("acceptance", PhaseAcceptancePublicEndpointsObservable),
				phaseItem("acceptance", PhaseAcceptanceNetworkRecoversRestarts),
				phaseItem("acceptance", PhaseAcceptanceCoreFlowsEndToEnd),
			},
		},
	}
}

func ValidateImplementationPhasePlan(plan ImplementationPhasePlan) error {
	report := BuildImplementationPhaseReport(plan)
	if !report.Ready {
		return fmt.Errorf("implementation phase %s failed: %v", report.PhaseID, report.Failed)
	}
	return nil
}

func BuildImplementationPhaseReport(plan ImplementationPhasePlan) ImplementationPhaseReport {
	expected := expectedImplementationPhaseItems(plan.PhaseID)
	failed := make([]string, 0)
	seen := map[string]ImplementationPhaseItem{}
	required := 0
	done := 0
	if plan.PhaseID == "" {
		failed = append(failed, "phase_id_required")
	}
	if len(expected) == 0 {
		failed = append(failed, plan.PhaseID+":unknown_phase")
	}
	for _, item := range plan.Items {
		if item.ID == "" || item.Kind == "" {
			failed = append(failed, "phase_item_id_and_kind_required")
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
	return ImplementationPhaseReport{
		PhaseID:	plan.PhaseID,
		Required:	required,
		Done:		done,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func phaseItem(kind, id string) ImplementationPhaseItem {
	return ImplementationPhaseItem{
		ID:		id,
		Kind:		kind,
		Required:	true,
		Done:		true,
		Evidence:	"required " + kind + " evidence for " + id,
	}
}

func expectedImplementationPhaseItems(phaseID string) map[string]bool {
	out := map[string]bool{}
	for _, plan := range defaultImplementationPhaseItemIDs() {
		if plan.phaseID != phaseID {
			continue
		}
		for _, id := range plan.ids {
			out[id] = true
		}
	}
	return out
}

type phaseItemIDs struct {
	phaseID	string
	ids	[]string
}

func defaultImplementationPhaseItemIDs() []phaseItemIDs {
	return []phaseItemIDs{
		{
			phaseID:	ImplementationPhaseBaselineAudit,
			ids: []string{
				PhaseTaskInspectVersions,
				PhaseTaskDocumentModuleGraph,
				PhaseTaskIdentifyOverlappingModules,
				PhaseTaskDecideRenameReuseWrap,
				PhaseTaskVerifyNaetStakingDenom,
				PhaseTaskVerifyEconomyWiring,
				PhaseTaskVerifyLocalnetAndCoverage,
				PhaseDeliverableModuleInventory,
				PhaseDeliverableGapAnalysis,
				PhaseDeliverableRiskList,
				PhaseDeliverableImplementationChecklist,
				PhaseTestFullUnitRun,
				PhaseTestIntegrationRun,
				PhaseTestLocalnetSmoke,
				PhaseTestExportImport,
			},
		},
		{
			phaseID:	ImplementationPhaseStakingPolicyCap,
			ids: []string{
				PhaseTaskImplementEffectivePowerCap,
				PhaseTaskImplementOverflowAccounting,
				PhaseTaskImplementCommissionPolicy,
				PhaseTaskAddConcentrationMetrics,
				PhaseTaskAddStakeQueries,
				PhaseTaskAddGovernanceParams,
				PhaseTaskWireModuleLifecycle,
				PhaseTestCapMathUnit,
				PhaseTestValidatorSetTransition,
				PhaseTestConcentrationQuery,
				PhaseTestCommissionBounds,
				PhaseTestStakingIntegration,
				PhaseTestStakingExportImport,
				PhaseTestInvariant,
				PhaseAcceptanceNoValidatorExceedsCap,
				PhaseAcceptanceExcessNoVotingPower,
				PhaseAcceptanceParamsSafeBounds,
				PhaseAcceptanceDeterministicExportImport,
			},
		},
		{
			phaseID:	ImplementationPhaseEconomicsFeeSplit,
			ids: []string{
				PhaseTaskImplementInflationBounds,
				PhaseTaskImplementTargetBondedRatio,
				PhaseTaskImplementFeeSplit,
				PhaseTaskImplementRewardSmoothing,
				PhaseTaskExposeAPREstimateQuery,
				PhaseTaskExposeSupplyTreasuryQueries,
				PhaseTaskAddEconomicsGovernanceParams,
				PhaseTestInflationCurve,
				PhaseTestBondedRatio,
				PhaseTestFeeSplit,
				PhaseTestBurnAccounting,
				PhaseTestTreasuryAccounting,
				PhaseTestAPRQuery,
				PhaseTestSupplyInvariant,
				PhaseTestEconomicsExportImport,
				PhaseAcceptanceInflationWithinBounds,
				PhaseAcceptanceFeeSplitSumsToFullAmount,
				PhaseAcceptanceBurnReducesSupply,
				PhaseAcceptanceTreasuryReceivesAmount,
				PhaseAcceptanceRewardsDeterministic,
			},
		},
		{
			phaseID:	ImplementationPhaseValidatorScore,
			ids: []string{
				PhaseTaskImplementUptimeScore,
				PhaseTaskImplementSlashHistory,
				PhaseTaskImplementGovernanceScore,
				PhaseTaskImplementDecentralizationScore,
				PhaseTaskImplementValidatorMetricQueries,
				PhaseTaskIntegrateObjectiveRewardModifier,
				PhaseTestUptimeWindow,
				PhaseTestMissedBlock,
				PhaseTestSlashHistory,
				PhaseTestGovernanceParticipation,
				PhaseTestScoreDeterminism,
				PhaseTestRewardModifier,
				PhaseTestValidatorScoreExportImport,
				PhaseAcceptanceScoreDeterministic,
				PhaseAcceptanceScoreObjectiveOnly,
				PhaseAcceptanceScoreQueryable,
				PhaseAcceptanceScoreConsensusSafe,
			},
		},
		{
			phaseID:	ImplementationPhaseSlashingHardening,
			ids: []string{
				PhaseTaskConfigureDoubleSignSlashTombstone,
				PhaseTaskConfigureDowntimeJail,
				PhaseTaskImplementProgressiveDowntime,
				PhaseTaskAddObjectiveTimestampProposalPolicy,
				PhaseTaskDocumentEvidenceLifecycle,
				PhaseTestDoubleSignEvidence,
				PhaseTestDowntime,
				PhaseTestJailUnjail,
				PhaseTestProgressiveDowntime,
				PhaseTestSlashingAccounting,
				PhaseTestDelegatorLoss,
				PhaseTestTombstone,
				PhaseTestEvidenceExpiry,
				PhaseAcceptanceDoubleSignTombstone,
				PhaseAcceptanceDowntimeBoundedProgressive,
				PhaseAcceptanceNoSubjectiveSlashing,
				PhaseAcceptanceSlashingStakeShareSafe,
			},
		},
		{
			phaseID:	ImplementationPhaseAVMIntegration,
			ids: []string{
				PhaseTaskFinalizeAVMWiring,
				PhaseTaskDefineCodeUploadPolicy,
				PhaseTaskDefineContractGasLimits,
				PhaseTaskDefineContractSizeLimits,
				PhaseTaskIntegrateStorageRentPricing,
				PhaseTaskExposeContractIndexerEvents,
				PhaseTaskDocumentContractDeveloperFlow,
				PhaseTestContractTxFlow,
				PhaseTestContractMigration,
				PhaseTestContractGasLimit,
				PhaseTestContractStorageLimitRent,
				PhaseTestMaliciousContract,
				PhaseTestContractExportImport,
				PhaseTestLocalnetAVMSmoke,
				PhaseAcceptanceContractsDeterministic,
				PhaseAcceptanceContractGasBounded,
				PhaseAcceptanceMaliciousContractsSafe,
				PhaseAcceptanceContractStateExportImport,
			},
		},
		{
			phaseID:	ImplementationPhaseFinalityPerformance,
			ids: []string{
				PhaseTaskConfigureBlockTimeTargets,
				PhaseTaskConfigureBlockSizeGasLimits,
				PhaseTaskProfile100ValidatorLocalnet,
				PhaseTaskProfile150To200Validators,
				PhaseTaskEstimate250To300Requirements,
				PhaseTaskMeasureFinalityUnderLoad,
				PhaseTaskMeasureFinalityUnderValidatorFailure,
				PhaseTestLocalnetLoadProfile,
				PhaseTestMempoolPressure,
				PhaseTestBlockTimeMeasurement,
				PhaseTestFinalityMeasurement,
				PhaseTestValidatorFailureScenario,
				PhaseTestRestartScenario,
				PhaseTestStateSyncSnapshotScenario,
				PhaseAcceptanceNormalFinalityWithinTarget,
				PhaseAcceptanceStressedFinalityUnderLimit,
				PhaseAcceptanceMediumNodeRequirements,
				PhaseAcceptanceNoExcessiveConsensusPayloads,
			},
		},
		{
			phaseID:	ImplementationPhasePublicTestnetReady,
			ids: []string{
				PhaseTaskWriteValidatorSetupDocs,
				PhaseTaskWriteSentryArchitectureDocs,
				PhaseTaskWriteMonitoringDocs,
				PhaseTaskPublishGenesisParamExplanation,
				PhaseTaskPublishEconomicModelExplanation,
				PhaseTaskPublishSlashingRiskExplanation,
				PhaseTaskPublishDelegationPoolGuide,
				PhaseTaskPublishAVMDeveloperGuide,
				PhaseTaskPreparePublicDashboards,
				PhaseTaskPrepareIncidentResponseProcess,
				PhaseTestCleanNodeBootstrapDocs,
				PhaseTestValidatorJoinDocs,
				PhaseTestSnapshotRestoreDocs,
				PhaseTestStateSyncDocs,
				PhaseTestTxFlowSmoke,
				PhaseTestGovernanceProposalSmoke,
				PhaseTestPublicEndpointSmoke,
				PhaseAcceptanceValidatorCanJoinFromDocs,
				PhaseAcceptancePublicEndpointsObservable,
				PhaseAcceptanceNetworkRecoversRestarts,
				PhaseAcceptanceCoreFlowsEndToEnd,
			},
		},
	}
}
