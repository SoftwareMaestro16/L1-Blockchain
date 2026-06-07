package params

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultAetraValidatorScoreSpecCoversModulePurpose(t *testing.T) {
	evidence := DefaultAetraValidatorScoreSpecEvidence()

	report := BuildAetraValidatorScoreSpecReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.Equal(t, AetraValidatorScoreModuleName, report.ModuleName)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 1, report.Required)
	require.NoError(t, ValidateAetraValidatorScoreSpec(evidence))
}

func TestAetraValidatorScoreSpecRejectsMissingPurpose(t *testing.T) {
	evidence := DefaultAetraValidatorScoreSpecEvidence()
	evidence.ModuleName = "x/validator-score"
	evidence.PublicAccountabilityWithoutSubjectiveConsensusControl = false

	report := BuildAetraValidatorScoreSpecReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_must_be_"+AetraValidatorScoreModuleName)
	require.Contains(t, report.Failed, AetraValidatorScorePurposePublicAccountability)
	require.Error(t, ValidateAetraValidatorScoreSpec(evidence))
}

func TestDefaultAetraValidatorScoreResponsibilitiesCoverSection241(t *testing.T) {
	evidence := DefaultAetraValidatorScoreResponsibilitiesEvidence()

	report := BuildAetraValidatorScoreResponsibilitiesReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.Equal(t, AetraValidatorScoreModuleName, report.ModuleName)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 10, report.Required)
	require.NoError(t, ValidateAetraValidatorScoreResponsibilities(evidence))
}

func TestAetraValidatorScoreResponsibilitiesRejectMissingRequiredItems(t *testing.T) {
	evidence := DefaultAetraValidatorScoreResponsibilitiesEvidence()
	evidence.ModuleName = ""
	evidence.TracksValidatorUptime = false
	evidence.TracksMissedBlockWindows = false
	evidence.TracksJailHistory = false
	evidence.TracksSlashingHistory = false
	evidence.TracksCommissionBehavior = false
	evidence.TracksSelfBondRatio = false
	evidence.TracksGovernanceParticipation = false
	evidence.TracksConcentrationStatus = false
	evidence.ProducesPublicScore = false
	evidence.ExposesExplorerFriendlyQueries = false

	report := BuildAetraValidatorScoreResponsibilitiesReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_required")
	require.Contains(t, report.Failed, AetraValidatorScoreResponsibilityTrackUptime)
	require.Contains(t, report.Failed, AetraValidatorScoreResponsibilityTrackMissedBlockWindows)
	require.Contains(t, report.Failed, AetraValidatorScoreResponsibilityTrackJailHistory)
	require.Contains(t, report.Failed, AetraValidatorScoreResponsibilityTrackSlashingHistory)
	require.Contains(t, report.Failed, AetraValidatorScoreResponsibilityTrackCommissionBehavior)
	require.Contains(t, report.Failed, AetraValidatorScoreResponsibilityTrackSelfBondRatio)
	require.Contains(t, report.Failed, AetraValidatorScoreResponsibilityTrackGovernanceParticipation)
	require.Contains(t, report.Failed, AetraValidatorScoreResponsibilityTrackConcentrationStatus)
	require.Contains(t, report.Failed, AetraValidatorScoreResponsibilityProducePublicScore)
	require.Contains(t, report.Failed, AetraValidatorScoreResponsibilityExplorerFriendlyQueries)
	require.Error(t, ValidateAetraValidatorScoreResponsibilities(evidence))
}

func TestDefaultAetraValidatorScoreSubjectiveControlGuards(t *testing.T) {
	evidence := DefaultAetraValidatorScoreSubjectiveControlEvidence()

	report := BuildAetraValidatorScoreSubjectiveControlReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.Equal(t, AetraValidatorScoreModuleName, report.ModuleName)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 5, report.Required)
	require.NoError(t, ValidateAetraValidatorScoreSubjectiveControl(evidence))
}

func TestAetraValidatorScoreSubjectiveControlRejectsMissingGuards(t *testing.T) {
	evidence := DefaultAetraValidatorScoreSubjectiveControlEvidence()
	evidence.ModuleName = "x/reputation"
	evidence.NoSubjectiveCensorshipMechanism = false
	evidence.InformationalFirst = false
	evidence.RewardAffectingOnlyObjectiveData = false
	evidence.ConsensusOverrideDisabledDefault = false
	evidence.ObjectiveInputsDeterministic = false

	report := BuildAetraValidatorScoreSubjectiveControlReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_must_be_"+AetraValidatorScoreModuleName)
	require.Contains(t, report.Failed, AetraValidatorScoreGuardNoSubjectiveCensorship)
	require.Contains(t, report.Failed, AetraValidatorScoreGuardInformationalFirst)
	require.Contains(t, report.Failed, AetraValidatorScoreGuardObjectiveRewardOnly)
	require.Contains(t, report.Failed, AetraValidatorScoreGuardConsensusOverrideDisabled)
	require.Contains(t, report.Failed, AetraValidatorScoreGuardObjectiveInputsDeterministic)
	require.Error(t, ValidateAetraValidatorScoreSubjectiveControl(evidence))
}

func TestDefaultAetraValidatorScoreStateSpecCoversSection242(t *testing.T) {
	evidence := DefaultAetraValidatorScoreStateSpecEvidence()

	report := BuildAetraValidatorScoreStateSpecReport(evidence)
	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.Equal(t, AetraValidatorScoreModuleName, report.ModuleName)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 22, report.Required)
	require.NoError(t, ValidateAetraValidatorScoreStateSpec(evidence))
}

func TestAetraValidatorScoreStateSpecRejectsMissingFields(t *testing.T) {
	evidence := DefaultAetraValidatorScoreStateSpecEvidence()
	evidence.ParamsFields = removeValidatorScoreString(evidence.ParamsFields, AetraValidatorScoreStateParamRewardModifierEnabled)
	evidence.ValidatorScoreFields = removeValidatorScoreString(evidence.ValidatorScoreFields, AetraValidatorScoreStateScoreLastUpdatedHeight)

	report := BuildAetraValidatorScoreStateSpecReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, AetraValidatorScoreStateParams+"."+AetraValidatorScoreStateParamRewardModifierEnabled+":missing")
	require.Contains(t, report.Failed, AetraValidatorScoreStateValidatorScore+"."+AetraValidatorScoreStateScoreLastUpdatedHeight+":missing")
	require.Error(t, ValidateAetraValidatorScoreStateSpec(evidence))
}

func TestAetraValidatorScoreStateSpecRejectsDuplicateUnexpectedAndWrongModule(t *testing.T) {
	evidence := DefaultAetraValidatorScoreStateSpecEvidence()
	evidence.ModuleName = "x/reputation"
	evidence.ParamsFields = append(evidence.ParamsFields, AetraValidatorScoreStateParamUptimeWindow, "SubjectivePenaltyWeight")
	evidence.ValidatorScoreFields = append(evidence.ValidatorScoreFields, AetraValidatorScoreStateScoreScore, "LocalLatencyScore")

	report := BuildAetraValidatorScoreStateSpecReport(evidence)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, "module_name_must_be_"+AetraValidatorScoreModuleName)
	require.Contains(t, report.Failed, AetraValidatorScoreStateParams+"."+AetraValidatorScoreStateParamUptimeWindow+":duplicate")
	require.Contains(t, report.Failed, AetraValidatorScoreStateParams+".SubjectivePenaltyWeight:unexpected")
	require.Contains(t, report.Failed, AetraValidatorScoreStateValidatorScore+"."+AetraValidatorScoreStateScoreScore+":duplicate")
	require.Contains(t, report.Failed, AetraValidatorScoreStateValidatorScore+".LocalLatencyScore:unexpected")
	require.Error(t, ValidateAetraValidatorScoreStateSpec(evidence))
}

func removeValidatorScoreString(values []string, targets ...string) []string {
	targetSet := map[string]bool{}
	for _, target := range targets {
		targetSet[target] = true
	}

	filtered := make([]string, 0, len(values))
	for _, value := range values {
		if !targetSet[value] {
			filtered = append(filtered, value)
		}
	}
	return filtered
}
