package params

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFeatureCompletionRequiresCodeParamsGenesisQueriesEventsTestsDocs(t *testing.T) {
	evidence := FeatureCompletionEvidence{
		FeatureID:		"x/aetra-core",
		Code:			true,
		Params:			true,
		GenesisValidation:	true,
		Queries:		true,
		Events:			true,
		Tests:			true,
		Docs:			true,
	}

	report := BuildFeatureCompletionReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Equal(t, report.Required, report.Done)
	require.NoError(t, ValidateFeatureCompletion(evidence))

	evidence.Queries = false
	evidence.Tests = false
	report = BuildFeatureCompletionReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, FeatureCompletionQueries)
	require.Contains(t, report.Failed, FeatureCompletionTests)
	require.Error(t, ValidateFeatureCompletion(evidence))
}

func TestFeatureCompletionRejectsMissingFeatureID(t *testing.T) {
	report := BuildFeatureCompletionReport(FeatureCompletionEvidence{
		Code:			true,
		Params:			true,
		GenesisValidation:	true,
		Queries:		true,
		Events:			true,
		Tests:			true,
		Docs:			true,
	})
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "feature_id_required")
}

func TestDefaultCoreChainConfigurationScopePlanCoversTasksDeliverablesAndTests(t *testing.T) {
	plan := DefaultCoreChainConfigurationScopePlan()
	report := BuildEngineeringScopeReport(plan)
	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.Equal(t, report.Required, report.Done)
	require.NoError(t, ValidateEngineeringScopePlan(plan))

	ids := map[string]bool{}
	for _, item := range plan.Items {
		ids[item.ID] = true
	}
	for _, requiredID := range []string{
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
		require.True(t, ids[requiredID], requiredID)
	}
}

func TestDefaultConsensusParameterPolicyScopePlanCoversTasksDeliverablesAndTests(t *testing.T) {
	plan := DefaultConsensusParameterPolicyScopePlan()
	report := BuildEngineeringScopeReport(plan)
	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.Equal(t, report.Required, report.Done)
	require.NoError(t, ValidateEngineeringScopePlan(plan))

	ids := map[string]bool{}
	for _, item := range plan.Items {
		ids[item.ID] = true
	}
	for _, requiredID := range []string{
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
		require.True(t, ids[requiredID], requiredID)
	}
}

func TestEngineeringScopeRejectsMissingEvidenceAndRequiredItems(t *testing.T) {
	plan := DefaultConsensusParameterPolicyScopePlan()
	plan.Items[0].Done = false
	plan.Items = plan.Items[:len(plan.Items)-1]

	report := BuildEngineeringScopeReport(plan)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, ConsensusParamTaskBlockTimeRange+":missing_evidence")
	require.Contains(t, report.Failed, ConsensusParamTestEvidencePeriod+":missing")
	require.Error(t, ValidateEngineeringScopePlan(plan))
}

func TestEngineeringScopeRejectsUnknownScopeAndUnexpectedItems(t *testing.T) {
	plan := EngineeringScopePlan{
		ScopeID:	"unknown",
		Items: []EngineeringScopeItem{
			engineeringScopeItem("task", CoreChainTaskStakingDenomNaet),
		},
	}

	report := BuildEngineeringScopeReport(plan)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "unknown:unknown_scope")
	require.Contains(t, report.Failed, CoreChainTaskStakingDenomNaet+":unexpected")
}
