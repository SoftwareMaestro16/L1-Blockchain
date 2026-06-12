package params

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultImplementationPhasePlansCoverPhase0ThroughPhase7(t *testing.T) {
	plans := DefaultImplementationPhasePlans()
	require.Len(t, plans, 8)

	for _, plan := range plans {
		report := BuildImplementationPhaseReport(plan)
		require.True(t, report.Ready, report.Failed)
		require.Empty(t, report.Failed)
		require.Equal(t, report.Required, report.Done)
		require.NoError(t, ValidateImplementationPhasePlan(plan))
	}
}

func TestImplementationPhaseRejectsMissingEvidence(t *testing.T) {
	plan := DefaultImplementationPhasePlans()[0]
	plan.Items[0].Done = false

	report := BuildImplementationPhaseReport(plan)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, PhaseTaskInspectVersions+":missing_evidence")
	require.Error(t, ValidateImplementationPhasePlan(plan))
}

func TestImplementationPhaseRejectsMissingRequiredItem(t *testing.T) {
	plan := DefaultImplementationPhasePlans()[2]
	plan.Items = plan.Items[:len(plan.Items)-1]

	report := BuildImplementationPhaseReport(plan)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, PhaseAcceptanceRewardsDeterministic+":missing")
}

func TestImplementationPhaseEconomicsFeeSplitRequiresAllAcceptanceGates(t *testing.T) {
	plan := DefaultImplementationPhasePlans()[2]
	report := BuildImplementationPhaseReport(plan)
	require.True(t, report.Ready, report.Failed)

	ids := map[string]bool{}
	for _, item := range plan.Items {
		ids[item.ID] = true
	}
	for _, requiredID := range []string{
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
	} {
		require.True(t, ids[requiredID], requiredID)
	}
}

func TestImplementationPhaseValidatorScoreRequiresAllAcceptanceGates(t *testing.T) {
	plan := DefaultImplementationPhasePlans()[3]
	report := BuildImplementationPhaseReport(plan)
	require.True(t, report.Ready, report.Failed)

	ids := map[string]bool{}
	for _, item := range plan.Items {
		ids[item.ID] = true
	}
	for _, requiredID := range []string{
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
	} {
		require.True(t, ids[requiredID], requiredID)
	}
}

func TestImplementationPhaseSlashingHardeningRequiresAllAcceptanceGates(t *testing.T) {
	plan := DefaultImplementationPhasePlans()[4]
	report := BuildImplementationPhaseReport(plan)
	require.True(t, report.Ready, report.Failed)

	ids := map[string]bool{}
	for _, item := range plan.Items {
		ids[item.ID] = true
	}
	for _, requiredID := range []string{
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
	} {
		require.True(t, ids[requiredID], requiredID)
	}
}

func TestImplementationPhaseAVMIntegrationRequiresAllAcceptanceGates(t *testing.T) {
	plan := DefaultImplementationPhasePlans()[5]
	report := BuildImplementationPhaseReport(plan)
	require.True(t, report.Ready, report.Failed)

	ids := map[string]bool{}
	for _, item := range plan.Items {
		ids[item.ID] = true
	}
	for _, requiredID := range []string{
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
	} {
		require.True(t, ids[requiredID], requiredID)
	}
}

func TestImplementationPhaseFinalityPerformanceRequiresAllAcceptanceGates(t *testing.T) {
	plan := DefaultImplementationPhasePlans()[6]
	report := BuildImplementationPhaseReport(plan)
	require.True(t, report.Ready, report.Failed)

	ids := map[string]bool{}
	for _, item := range plan.Items {
		ids[item.ID] = true
	}
	for _, requiredID := range []string{
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
	} {
		require.True(t, ids[requiredID], requiredID)
	}
}

func TestImplementationPhasePublicTestnetReadinessRequiresAllAcceptanceGates(t *testing.T) {
	plan := DefaultImplementationPhasePlans()[7]
	report := BuildImplementationPhaseReport(plan)
	require.True(t, report.Ready, report.Failed)

	ids := map[string]bool{}
	for _, item := range plan.Items {
		ids[item.ID] = true
	}
	for _, requiredID := range []string{
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
	} {
		require.True(t, ids[requiredID], requiredID)
	}
}

func TestImplementationPhaseRejectsUnknownPhaseAndUnexpectedItem(t *testing.T) {
	plan := ImplementationPhasePlan{
		PhaseID:	"phase_99",
		Items: []ImplementationPhaseItem{
			phaseItem("task", PhaseTaskInspectVersions),
		},
	}

	report := BuildImplementationPhaseReport(plan)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "phase_99:unknown_phase")
	require.Contains(t, report.Failed, PhaseTaskInspectVersions+":unexpected")
}
