package params

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMainnetReadinessRequiresAllCriteria(t *testing.T) {
	evidence := completeMainnetReadinessEvidence()

	report := BuildMainnetReadinessReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 14, report.Required)
	require.NoError(t, ValidateMainnetReadiness(evidence))
}

func TestMainnetReadinessRejectsMissingConsensusAndEconomicsCriteria(t *testing.T) {
	evidence := completeMainnetReadinessEvidence()
	evidence.ValidatorSetPolicyImplementedAndTested = false
	evidence.AntiConcentrationRewardsImplementedAndTested = false
	evidence.FeeBurnTreasuryRewardSplitImplementedAndTested = false

	report := BuildMainnetReadinessReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, MainnetReadinessValidatorSetPolicy)
	require.Contains(t, report.Failed, MainnetReadinessAntiConcentrationRewards)
	require.Contains(t, report.Failed, MainnetReadinessFeeBurnTreasuryRewards)
	require.Error(t, ValidateMainnetReadiness(evidence))
}

func TestMainnetReadinessRejectsMissingOperationalAndSecurityCriteria(t *testing.T) {
	evidence := completeMainnetReadinessEvidence()
	evidence.PublicTestnetObservedValidatorBehavior = false
	evidence.LoadTestsDemonstrateFinalityTarget = false
	evidence.SecurityAuditCompleted = false
	evidence.CriticalFindingsFixed = false
	evidence.DocsComplete = false

	report := BuildMainnetReadinessReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, MainnetReadinessPublicTestnetDuration)
	require.Contains(t, report.Failed, MainnetReadinessFinalityLoadTests)
	require.Contains(t, report.Failed, MainnetReadinessSecurityAudit)
	require.Contains(t, report.Failed, MainnetReadinessCriticalFindingsFixed)
	require.Contains(t, report.Failed, MainnetReadinessDocsComplete)
}

func completeMainnetReadinessEvidence() MainnetReadinessEvidence {
	return MainnetReadinessEvidence{
		ValidatorSetPolicyImplementedAndTested:		true,
		EffectivePowerCapImplementedAndTested:		true,
		AntiConcentrationRewardsImplementedAndTested:	true,
		DynamicInflationImplementedAndTested:		true,
		FeeBurnTreasuryRewardSplitImplementedAndTested:	true,
		SlashingConfiguredAndTested:			true,
		AVMIntegratedAndTested:				true,
		ExportImportStable:				true,
		StateSyncSnapshotsStable:			true,
		PublicTestnetObservedValidatorBehavior:		true,
		LoadTestsDemonstrateFinalityTarget:		true,
		SecurityAuditCompleted:				true,
		CriticalFindingsFixed:				true,
		DocsComplete:					true,
	}
}
