package params

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultAetraFinalTargetFormulaMatchesTarget(t *testing.T) {
	formula := DefaultAetraFinalTargetFormula()

	report := BuildFinalTargetFormulaReport(formula)
	require.True(t, report.Ready, report.Failed)
	require.Empty(t, report.Failed)
	require.Equal(t, report.Required, report.Passed)
	require.Equal(t, 14, report.Required)
	require.NoError(t, ValidateAetraFinalTargetFormula(formula))

	require.Equal(t, 100, formula.MinActiveValidators)
	require.Equal(t, 300, formula.MaxActiveValidators)
	require.Equal(t, 5, formula.MinBlockTimeSeconds)
	require.Equal(t, 8, formula.MaxBlockTimeSeconds)
	require.Equal(t, 120, formula.WorstAcceptableFinalitySeconds)
}

func TestAetraFinalTargetFormulaRejectsMissingCoreStack(t *testing.T) {
	formula := DefaultAetraFinalTargetFormula()
	formula.CometBFTBFTPoS = false
	formula.CosmosSDK = false
	formula.AVMOnlyGenesis = false

	report := BuildFinalTargetFormulaReport(formula)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, FinalTargetConsensusCometBFTBFTPoS)
	require.Contains(t, report.Failed, FinalTargetCosmosSDK)
	require.Contains(t, report.Failed, FinalTargetAVMOnlyGenesis)
	require.Error(t, ValidateAetraFinalTargetFormula(formula))
}

func TestAetraFinalTargetFormulaRejectsUnsafePerformanceTargets(t *testing.T) {
	formula := DefaultAetraFinalTargetFormula()
	formula.MinActiveValidators = 80
	formula.MaxActiveValidators = 500
	formula.MinBlockTimeSeconds = 1
	formula.MaxBlockTimeSeconds = 2
	formula.WorstAcceptableFinalitySeconds = 180

	report := BuildFinalTargetFormulaReport(formula)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, FinalTargetValidatorSetRange)
	require.Contains(t, report.Failed, FinalTargetBlockTimeRange)
	require.Contains(t, report.Failed, FinalTargetWorstFinality)
}

func TestAetraFinalTargetFormulaRejectsMissingTrustEconomicsAndTestPolicy(t *testing.T) {
	formula := DefaultAetraFinalTargetFormula()
	formula.StrictObjectiveSlashing = false
	formula.ValidatorEffectivePowerCap = false
	formula.AntiConcentrationRewards = false
	formula.DynamicLowModerateInflation = false
	formula.FeeBurn = false
	formula.ProtocolTreasury = false
	formula.MandatoryTestsForEveryFeature = false
	formula.TrustOverSpeedOrShortTermAPR = false

	report := BuildFinalTargetFormulaReport(formula)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, FinalTargetObjectiveSlashing)
	require.Contains(t, report.Failed, FinalTargetValidatorPowerCap)
	require.Contains(t, report.Failed, FinalTargetAntiConcentrationRewards)
	require.Contains(t, report.Failed, FinalTargetDynamicLowModerateInflation)
	require.Contains(t, report.Failed, FinalTargetFeeBurn)
	require.Contains(t, report.Failed, FinalTargetProtocolTreasury)
	require.Contains(t, report.Failed, FinalTargetMandatoryFeatureTests)
	require.Contains(t, report.Failed, FinalTargetTrustProductDecision)
}
