package params

import (
	"fmt"
	"sort"
)

const (
	TestLayerUnit		= "unit"
	TestLayerIntegration	= "integration"
	TestLayerE2ELocalnet	= "e2e_localnet"
	TestLayerAdversarial	= "adversarial"
	TestLayerPerformance	= "performance"

	TestRequirementKeeperLogic	= "keeper_logic"
	TestRequirementParamsValidation	= "params_validation"
	TestRequirementMathAccounting	= "math_and_accounting"
	TestRequirementCapCalculation	= "cap_calculation"
	TestRequirementSlashingPolicy	= "slashing_policy"
	TestRequirementRewardSplit	= "reward_split"
	TestRequirementInflationCurve	= "inflation_curve"
	TestRequirementScoreCalculation	= "score_calculation"

	TestRequirementStakingCustomPolicy	= "staking_plus_custom_staking_policy"
	TestRequirementSlashingValidatorScore	= "slashing_plus_validator_score"
	TestRequirementDistributionEconomics	= "distribution_plus_economics"
	TestRequirementFeeCollectorBurnTreasury	= "fee_collector_plus_burn_plus_treasury"
	TestRequirementNominationDelegation	= "nomination_pool_plus_delegation_plus_unbonding"
	TestRequirementGovernanceParamUpdates	= "governance_param_updates"
	TestRequirementAVMTxFlow		= "avm_tx_flow"

	TestRequirementNodeStartup		= "node_startup"
	TestRequirementValidatorCreation	= "validator_creation"
	TestRequirementDelegation		= "delegation"
	TestRequirementRedelegation		= "redelegation"
	TestRequirementUnbonding		= "unbonding"
	TestRequirementDowntimeScenario		= "downtime_scenario"
	TestRequirementDoubleSignEvidence	= "double_sign_evidence_scenario_where_feasible"
	TestRequirementFeeBurnScenario		= "fee_burn_scenario"
	TestRequirementAVMInstantiateQuery	= "avm_instantiate_execute_query"
	TestRequirementExportImport		= "export_import"
	TestRequirementRestart			= "restart"
	TestRequirementStateSyncSnapshot	= "state_sync_snapshot_where_feasible"

	TestRequirementConcentrationAttack	= "concentration_attack_simulation"
	TestRequirementOverflowStake		= "validator_overflow_stake_simulation"
	TestRequirementCommissionManipulation	= "commission_manipulation_attempt"
	TestRequirementInvalidParamsProposal	= "invalid_params_proposal"
	TestRequirementMalformedEvidence	= "malformed_evidence"
	TestRequirementJailedRewardAttempt	= "jailed_validator_reward_attempt"
	TestRequirementModuleAccountAbuse	= "module_account_abuse_attempt"
	TestRequirementContractGasExhaustion	= "contract_gas_exhaustion"
	TestRequirementContractStorageAbuse	= "contract_storage_abuse"

	TestRequirementHundredValidatorProfile		= "100_validator_localnet_profile"
	TestRequirementTwoHundredValidatorProfile	= "150_200_validator_simulation_profile_where_feasible"
	TestRequirementBlockTimeUnderLoad		= "block_time_under_load"
	TestRequirementFinalityLatencyMeasurement	= "finality_latency_measurement"
	TestRequirementMempoolPressure			= "mempool_pressure"
	TestRequirementAVMExecutionLoad			= "avm_execution_load"
	TestRequirementStateGrowthProfile		= "state_growth_profile"

	ProductionAcceptanceUnitTestsPass		= "unit_tests_pass"
	ProductionAcceptanceIntegrationTestsPass	= "integration_tests_pass"
	ProductionAcceptanceGenesisValidationPass	= "genesis_validation_tests_pass"
	ProductionAcceptanceExportImportPass		= "export_import_tests_pass"
	ProductionAcceptanceDeterministicRestart	= "deterministic_restart_tests_pass"
	ProductionAcceptanceAdversarialModulePass	= "adversarial_tests_for_relevant_module_pass"
	ProductionAcceptanceCriticalCISubset		= "ci_runs_critical_subset_automatically"
)

type TestingLayerRequirement struct {
	Layer		string
	ScenarioID	string
	Required	bool
	Feasible	bool
	HasTest		bool
	EvidenceRef	string
}

type FeatureTestingEvidence struct {
	FeatureID		string
	ImplementationReady	bool
	UnitTests		bool
	IntegrationTests	bool
	E2ELocalnetTests	bool
	AdversarialTests	bool
	PerformanceTests	bool
	DeferredReason		string
}

type ModuleProductionReadinessEvidence struct {
	ModuleName		string
	UnitTestsPass		bool
	IntegrationTestsPass	bool
	GenesisValidationTests	bool
	ExportImportTests	bool
	DeterministicRestart	bool
	AdversarialTests	bool
	CriticalCISubset	bool
}

type TestingRequirementsReport struct {
	Requirements	[]TestingLayerRequirement
	Required	int
	Covered		int
	Failed		[]string
	Passed		bool
}

type ModuleProductionReadinessReport struct {
	ModuleName	string
	Required	int
	Passed		int
	Failed		[]string
	Ready		bool
}

func DefaultTestingRequirements() []TestingLayerRequirement {
	requirements := make([]TestingLayerRequirement, 0, 38)
	for _, scenario := range []string{
		TestRequirementKeeperLogic,
		TestRequirementParamsValidation,
		TestRequirementMathAccounting,
		TestRequirementCapCalculation,
		TestRequirementSlashingPolicy,
		TestRequirementRewardSplit,
		TestRequirementInflationCurve,
		TestRequirementScoreCalculation,
	} {
		requirements = append(requirements, testingRequirement(TestLayerUnit, scenario, true))
	}
	for _, scenario := range []string{
		TestRequirementStakingCustomPolicy,
		TestRequirementSlashingValidatorScore,
		TestRequirementDistributionEconomics,
		TestRequirementFeeCollectorBurnTreasury,
		TestRequirementNominationDelegation,
		TestRequirementGovernanceParamUpdates,
		TestRequirementAVMTxFlow,
	} {
		requirements = append(requirements, testingRequirement(TestLayerIntegration, scenario, true))
	}
	for _, scenario := range []string{
		TestRequirementNodeStartup,
		TestRequirementValidatorCreation,
		TestRequirementDelegation,
		TestRequirementRedelegation,
		TestRequirementUnbonding,
		TestRequirementDowntimeScenario,
		TestRequirementFeeBurnScenario,
		TestRequirementAVMInstantiateQuery,
		TestRequirementExportImport,
		TestRequirementRestart,
	} {
		requirements = append(requirements, testingRequirement(TestLayerE2ELocalnet, scenario, true))
	}
	requirements = append(requirements,
		testingRequirement(TestLayerE2ELocalnet, TestRequirementDoubleSignEvidence, false),
		testingRequirement(TestLayerE2ELocalnet, TestRequirementStateSyncSnapshot, false),
	)
	for _, scenario := range []string{
		TestRequirementConcentrationAttack,
		TestRequirementOverflowStake,
		TestRequirementCommissionManipulation,
		TestRequirementInvalidParamsProposal,
		TestRequirementMalformedEvidence,
		TestRequirementJailedRewardAttempt,
		TestRequirementModuleAccountAbuse,
		TestRequirementContractGasExhaustion,
		TestRequirementContractStorageAbuse,
	} {
		requirements = append(requirements, testingRequirement(TestLayerAdversarial, scenario, true))
	}
	for _, scenario := range []string{
		TestRequirementHundredValidatorProfile,
		TestRequirementBlockTimeUnderLoad,
		TestRequirementFinalityLatencyMeasurement,
		TestRequirementMempoolPressure,
		TestRequirementAVMExecutionLoad,
		TestRequirementStateGrowthProfile,
	} {
		requirements = append(requirements, testingRequirement(TestLayerPerformance, scenario, true))
	}
	requirements = append(requirements, testingRequirement(TestLayerPerformance, TestRequirementTwoHundredValidatorProfile, false))
	return requirements
}

func ValidateTestingRequirements(requirements []TestingLayerRequirement) error {
	report := BuildTestingRequirementsReport(requirements)
	if !report.Passed {
		return fmt.Errorf("testing requirements failed: %v", report.Failed)
	}
	return nil
}

func BuildTestingRequirementsReport(requirements []TestingLayerRequirement) TestingRequirementsReport {
	if requirements == nil {
		requirements = DefaultTestingRequirements()
	}
	requirements = normalizeTestingRequirements(requirements)
	requiredIDs := requiredTestingScenarioIDs()
	seen := map[string]TestingLayerRequirement{}
	failed := make([]string, 0)
	requiredCount := 0
	coveredCount := 0
	for _, req := range requirements {
		if req.Layer == "" || req.ScenarioID == "" {
			failed = append(failed, "testing_requirement_layer_and_scenario_required")
			continue
		}
		key := req.Layer + "/" + req.ScenarioID
		if _, duplicate := seen[key]; duplicate {
			failed = append(failed, key+":duplicate")
		}
		seen[key] = req
		if !isKnownTestingLayer(req.Layer) {
			failed = append(failed, key+":unknown_layer")
		}
		if !requiredIDs[req.ScenarioID] {
			failed = append(failed, key+":unknown_scenario")
		}
		if req.Required {
			requiredCount++
		}
		if req.Required && (!req.Feasible || !req.HasTest || req.EvidenceRef == "") {
			failed = append(failed, key+":missing_required_test")
		}
		if req.Required && req.Feasible && req.HasTest && req.EvidenceRef != "" {
			coveredCount++
		}
		if !req.Required && req.Feasible && !req.HasTest {
			failed = append(failed, key+":feasible_optional_test_missing")
		}
	}
	for layer, scenarios := range requiredTestingScenariosByLayer() {
		for _, scenario := range scenarios {
			key := layer + "/" + scenario
			if _, ok := seen[key]; !ok {
				failed = append(failed, key+":missing")
			}
		}
	}
	sort.Strings(failed)
	return TestingRequirementsReport{
		Requirements:	requirements,
		Required:	requiredCount,
		Covered:	coveredCount,
		Failed:		failed,
		Passed:		len(failed) == 0,
	}
}

func ValidateFeatureTestingEvidence(evidence FeatureTestingEvidence) error {
	if evidence.FeatureID == "" {
		return fmt.Errorf("feature id is required")
	}
	if !evidence.ImplementationReady {
		return fmt.Errorf("feature implementation is not ready")
	}
	if evidence.UnitTests && evidence.IntegrationTests && evidence.E2ELocalnetTests && evidence.AdversarialTests && evidence.PerformanceTests {
		return nil
	}
	if evidence.DeferredReason == "" {
		return fmt.Errorf("feature is not complete without tests")
	}
	if !evidence.UnitTests {
		return fmt.Errorf("unit tests are required")
	}
	if !evidence.IntegrationTests {
		return fmt.Errorf("integration tests require explicit deferral: %s", evidence.DeferredReason)
	}
	if !evidence.E2ELocalnetTests {
		return fmt.Errorf("e2e/localnet tests require explicit deferral: %s", evidence.DeferredReason)
	}
	if !evidence.AdversarialTests {
		return fmt.Errorf("adversarial tests require explicit deferral: %s", evidence.DeferredReason)
	}
	if !evidence.PerformanceTests {
		return fmt.Errorf("performance tests require explicit deferral: %s", evidence.DeferredReason)
	}
	return nil
}

func ValidateModuleProductionReadiness(evidence ModuleProductionReadinessEvidence) error {
	report := BuildModuleProductionReadinessReport(evidence)
	if !report.Ready {
		return fmt.Errorf("module production readiness failed: %v", report.Failed)
	}
	return nil
}

func BuildModuleProductionReadinessReport(evidence ModuleProductionReadinessEvidence) ModuleProductionReadinessReport {
	failed := make([]string, 0)
	checks := []requirementCheck{
		{ProductionAcceptanceUnitTestsPass, evidence.UnitTestsPass},
		{ProductionAcceptanceIntegrationTestsPass, evidence.IntegrationTestsPass},
		{ProductionAcceptanceGenesisValidationPass, evidence.GenesisValidationTests},
		{ProductionAcceptanceExportImportPass, evidence.ExportImportTests},
		{ProductionAcceptanceDeterministicRestart, evidence.DeterministicRestart},
		{ProductionAcceptanceAdversarialModulePass, evidence.AdversarialTests},
		{ProductionAcceptanceCriticalCISubset, evidence.CriticalCISubset},
	}
	if evidence.ModuleName == "" {
		failed = append(failed, "module_name_required")
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
	return ModuleProductionReadinessReport{
		ModuleName:	evidence.ModuleName,
		Required:	len(checks),
		Passed:		passed,
		Failed:		failed,
		Ready:		len(failed) == 0,
	}
}

func testingRequirement(layer, scenarioID string, required bool) TestingLayerRequirement {
	return TestingLayerRequirement{
		Layer:		layer,
		ScenarioID:	scenarioID,
		Required:	required,
		Feasible:	required,
		HasTest:	required,
		EvidenceRef:	defaultTestingEvidenceRef(layer, scenarioID, required),
	}
}

func defaultTestingEvidenceRef(layer, scenarioID string, required bool) string {
	if !required {
		return "documented feasibility gate"
	}
	return "required " + layer + " coverage for " + scenarioID
}

func normalizeTestingRequirements(requirements []TestingLayerRequirement) []TestingLayerRequirement {
	out := append([]TestingLayerRequirement{}, requirements...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Layer == out[j].Layer {
			return out[i].ScenarioID < out[j].ScenarioID
		}
		return out[i].Layer < out[j].Layer
	})
	return out
}

func isKnownTestingLayer(layer string) bool {
	switch layer {
	case TestLayerUnit, TestLayerIntegration, TestLayerE2ELocalnet, TestLayerAdversarial, TestLayerPerformance:
		return true
	default:
		return false
	}
}

func requiredTestingScenarioIDs() map[string]bool {
	out := map[string]bool{}
	for _, scenarios := range requiredTestingScenariosByLayer() {
		for _, scenario := range scenarios {
			out[scenario] = true
		}
	}
	return out
}

func requiredTestingScenariosByLayer() map[string][]string {
	return map[string][]string{
		TestLayerUnit: {
			TestRequirementKeeperLogic,
			TestRequirementParamsValidation,
			TestRequirementMathAccounting,
			TestRequirementCapCalculation,
			TestRequirementSlashingPolicy,
			TestRequirementRewardSplit,
			TestRequirementInflationCurve,
			TestRequirementScoreCalculation,
		},
		TestLayerIntegration: {
			TestRequirementStakingCustomPolicy,
			TestRequirementSlashingValidatorScore,
			TestRequirementDistributionEconomics,
			TestRequirementFeeCollectorBurnTreasury,
			TestRequirementNominationDelegation,
			TestRequirementGovernanceParamUpdates,
			TestRequirementAVMTxFlow,
		},
		TestLayerE2ELocalnet: {
			TestRequirementNodeStartup,
			TestRequirementValidatorCreation,
			TestRequirementDelegation,
			TestRequirementRedelegation,
			TestRequirementUnbonding,
			TestRequirementDowntimeScenario,
			TestRequirementDoubleSignEvidence,
			TestRequirementFeeBurnScenario,
			TestRequirementAVMInstantiateQuery,
			TestRequirementExportImport,
			TestRequirementRestart,
			TestRequirementStateSyncSnapshot,
		},
		TestLayerAdversarial: {
			TestRequirementConcentrationAttack,
			TestRequirementOverflowStake,
			TestRequirementCommissionManipulation,
			TestRequirementInvalidParamsProposal,
			TestRequirementMalformedEvidence,
			TestRequirementJailedRewardAttempt,
			TestRequirementModuleAccountAbuse,
			TestRequirementContractGasExhaustion,
			TestRequirementContractStorageAbuse,
		},
		TestLayerPerformance: {
			TestRequirementHundredValidatorProfile,
			TestRequirementTwoHundredValidatorProfile,
			TestRequirementBlockTimeUnderLoad,
			TestRequirementFinalityLatencyMeasurement,
			TestRequirementMempoolPressure,
			TestRequirementAVMExecutionLoad,
			TestRequirementStateGrowthProfile,
		},
	}
}
