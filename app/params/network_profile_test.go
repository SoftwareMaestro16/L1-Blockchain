package params

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultNetworkProfileMatchesAetraBaseModel(t *testing.T) {
	profile := DefaultNetworkProfile()

	require.NoError(t, profile.Validate())
	require.Equal(t, AetraConsensusEngine, profile.ConsensusEngine)
	require.Equal(t, AetraStakingModel, profile.StakingModel)
	require.Equal(t, "AVM", profile.PrimaryVM)
	require.Equal(t, AetraPrimaryVM, profile.PrimaryVM)
	require.Equal(t, "CosmWasm gated; EVM later", profile.OptionalVM)
	require.Equal(t, AetraOptionalVM, profile.OptionalVM)
	require.Equal(t, AetraHardwareProfile, profile.HardwareProfile)

	require.Equal(t, 100, profile.ValidatorSetMin)
	require.Equal(t, 100, profile.ValidatorSetGenesisMin)
	require.Equal(t, 128, profile.ValidatorSetGenesisMax)
	require.Equal(t, 150, profile.ValidatorSetGrowthMin)
	require.Equal(t, 200, profile.ValidatorSetGrowthMax)
	require.Equal(t, 250, profile.ValidatorSetMatureMin)
	require.Equal(t, 300, profile.ValidatorSetMatureMax)
	require.Equal(t, 300, profile.ValidatorSetMax)

	require.Equal(t, 5, profile.BlockTimeMinSeconds)
	require.Equal(t, 8, profile.BlockTimeMaxSeconds)
	require.Equal(t, 5, profile.NormalFinalityMinSeconds)
	require.Equal(t, 15, profile.NormalFinalityMaxSeconds)
	require.Equal(t, 20, profile.StressFinalityMinSeconds)
	require.Equal(t, 90, profile.StressFinalityMaxSeconds)
	require.Equal(t, 120, profile.WorstFinalityTargetSeconds)

	require.Equal(t, int64(5_500), profile.TargetBondedRatioMinBps)
	require.Equal(t, int64(6_500), profile.TargetBondedRatioMaxBps)
	require.Equal(t, int64(6_000), profile.TargetBondedRatioDefaultBps)
	require.Equal(t, int64(300), profile.NormalInflationMinBps)
	require.Equal(t, int64(400), profile.NormalInflationMaxBps)
	require.Equal(t, int64(400), profile.DelegatorAPRTargetMinBps)
	require.Equal(t, int64(700), profile.DelegatorAPRTargetMaxBps)
	require.Equal(t, int64(600), profile.ValidatorNetAPRTargetMinBps)
	require.Equal(t, int64(900), profile.ValidatorNetAPRTargetMaxBps)
}

func TestNetworkProfileKeepsMintTargetsInsideEconomicsRanges(t *testing.T) {
	profile := DefaultNetworkProfile()

	require.GreaterOrEqual(t, DefaultTargetInflationBps, profile.NormalInflationMinBps)
	require.LessOrEqual(t, DefaultTargetInflationBps, profile.NormalInflationMaxBps)
	require.LessOrEqual(t, MinInflationBps, profile.NormalInflationMinBps)
	require.GreaterOrEqual(t, MaxInflationBps, profile.NormalInflationMaxBps)
	require.Equal(t, profile.TargetBondedRatioDefaultBps, DefaultTargetStakeBps)
	require.GreaterOrEqual(t, DefaultTargetStakeBps, profile.TargetBondedRatioMinBps)
	require.LessOrEqual(t, DefaultTargetStakeBps, profile.TargetBondedRatioMaxBps)
}

func TestNetworkProfileFeeRangesMatchBurnRewardTreasuryModel(t *testing.T) {
	profile := DefaultNetworkProfile()

	require.Equal(t, int64(3_000), profile.FeeBurnShareMinBps)
	require.Equal(t, int64(6_000), profile.FeeBurnShareMaxBps)
	require.Equal(t, int64(2_000), profile.FeeRewardShareMinBps)
	require.Equal(t, int64(4_000), profile.FeeRewardShareMaxBps)
	require.Equal(t, int64(1_000), profile.FeeTreasuryShareMinBps)
	require.Equal(t, int64(2_000), profile.FeeTreasuryShareMaxBps)

	exampleBurnShare := int64(5_000)
	exampleRewardShare := int64(3_500)
	exampleTreasuryShare := int64(1_500)
	require.GreaterOrEqual(t, exampleBurnShare, profile.FeeBurnShareMinBps)
	require.LessOrEqual(t, exampleBurnShare, profile.FeeBurnShareMaxBps)
	require.GreaterOrEqual(t, exampleRewardShare, profile.FeeRewardShareMinBps)
	require.LessOrEqual(t, exampleRewardShare, profile.FeeRewardShareMaxBps)
	require.GreaterOrEqual(t, exampleTreasuryShare, profile.FeeTreasuryShareMinBps)
	require.LessOrEqual(t, exampleTreasuryShare, profile.FeeTreasuryShareMaxBps)
	require.Equal(t, BasisPoints, exampleBurnShare+exampleRewardShare+exampleTreasuryShare)
}

func TestNetworkProfileRejectsUnsafeValidatorCounts(t *testing.T) {
	profile := DefaultNetworkProfile()
	profile.ValidatorSetMin = 80
	require.ErrorContains(t, profile.Validate(), "100-300")

	profile = DefaultNetworkProfile()
	profile.ValidatorSetMax = 500
	require.ErrorContains(t, profile.Validate(), "100-300")

	profile = DefaultNetworkProfile()
	profile.ValidatorSetGenesisMax = 150
	require.ErrorContains(t, profile.Validate(), "100-128")
}

func TestNetworkProfileRejectsUnsafeLatencyTargets(t *testing.T) {
	profile := DefaultNetworkProfile()
	profile.BlockTimeMinSeconds = 1
	require.ErrorContains(t, profile.Validate(), "5-8")

	profile = DefaultNetworkProfile()
	profile.NormalFinalityMaxSeconds = 30
	require.ErrorContains(t, profile.Validate(), "5-15")

	profile = DefaultNetworkProfile()
	profile.WorstFinalityTargetSeconds = 180
	require.ErrorContains(t, profile.Validate(), "120")
}

func TestNetworkProfileRejectsUnsafeEconomicTargets(t *testing.T) {
	profile := DefaultNetworkProfile()
	profile.TargetBondedRatioDefaultBps = 6_700
	require.ErrorContains(t, profile.Validate(), "target bonded ratio")

	profile = DefaultNetworkProfile()
	profile.NormalInflationMinBps = 200
	require.ErrorContains(t, profile.Validate(), "normal_inflation")

	profile = DefaultNetworkProfile()
	profile.DelegatorAPRTargetMaxBps = 800
	require.ErrorContains(t, profile.Validate(), "delegator_apr_target")

	profile = DefaultNetworkProfile()
	profile.ValidatorNetAPRTargetMinBps = 500
	require.ErrorContains(t, profile.Validate(), "validator_net_apr_target")

	profile = DefaultNetworkProfile()
	profile.FeeBurnShareMinBps = 2_000
	require.ErrorContains(t, profile.Validate(), "fee_burn_share")
}

func TestValidatorSetPhasePoliciesMatchAetraGrowthPlan(t *testing.T) {
	phases := DefaultValidatorSetPhasePolicies()
	require.Len(t, phases, 3)

	require.Equal(t, ValidatorSetPhasePolicy{
		Name:				AetraValidatorPhaseGenesis,
		MinActiveValidators:		100,
		MaxActiveValidators:		128,
		BlockTimeMinSeconds:		5,
		BlockTimeMaxSeconds:		6,
		TargetBlockTimeSeconds:		6,
		NormalFinalityMinSeconds:	5,
		NormalFinalityMaxSeconds:	10,
	}, phases[0])

	require.Equal(t, ValidatorSetPhasePolicy{
		Name:				AetraValidatorPhaseGrowth,
		MinActiveValidators:		150,
		MaxActiveValidators:		200,
		BlockTimeMinSeconds:		6,
		BlockTimeMaxSeconds:		6,
		TargetBlockTimeSeconds:		6,
		NormalFinalityMinSeconds:	6,
		NormalFinalityMaxSeconds:	12,
	}, phases[1])

	require.Equal(t, ValidatorSetPhasePolicy{
		Name:				AetraValidatorPhaseMature,
		MinActiveValidators:		250,
		MaxActiveValidators:		300,
		BlockTimeMinSeconds:		7,
		BlockTimeMaxSeconds:		8,
		TargetBlockTimeSeconds:		8,
		NormalFinalityMinSeconds:	8,
		NormalFinalityMaxSeconds:	15,
		RequiresOperatorReadiness:	true,
	}, phases[2])
}

func TestBlockTimeTargetRangesMatchConsensusPolicy(t *testing.T) {
	profile := DefaultNetworkProfile()

	min, max, err := profile.BlockTimeTargetRange(100)
	require.NoError(t, err)
	require.Equal(t, 5, min)
	require.Equal(t, 6, max)

	min, max, err = profile.BlockTimeTargetRange(200)
	require.NoError(t, err)
	require.Equal(t, 6, min)
	require.Equal(t, 6, max)

	min, max, err = profile.BlockTimeTargetRange(300)
	require.NoError(t, err)
	require.Equal(t, 7, min)
	require.Equal(t, 8, max)
}

func TestNetworkProfileSelectsValidatorSetPhase(t *testing.T) {
	profile := DefaultNetworkProfile()

	phase, err := profile.ValidatorSetPhase(100)
	require.NoError(t, err)
	require.Equal(t, AetraValidatorPhaseGenesis, phase.Name)

	phase, err = profile.ValidatorSetPhase(128)
	require.NoError(t, err)
	require.Equal(t, AetraValidatorPhaseGenesis, phase.Name)

	phase, err = profile.ValidatorSetPhase(150)
	require.NoError(t, err)
	require.Equal(t, AetraValidatorPhaseGrowth, phase.Name)

	phase, err = profile.ValidatorSetPhase(200)
	require.NoError(t, err)
	require.Equal(t, AetraValidatorPhaseGrowth, phase.Name)

	phase, err = profile.ValidatorSetPhase(250)
	require.NoError(t, err)
	require.Equal(t, AetraValidatorPhaseMature, phase.Name)

	phase, err = profile.ValidatorSetPhase(300)
	require.NoError(t, err)
	require.Equal(t, AetraValidatorPhaseMature, phase.Name)
}

func TestNetworkProfileRejectsValidatorCountsOutsideGrowthPhases(t *testing.T) {
	profile := DefaultNetworkProfile()

	_, err := profile.ValidatorSetPhase(80)
	require.ErrorContains(t, err, "100-300")

	_, err = profile.ValidatorSetPhase(129)
	require.ErrorContains(t, err, "outside configured growth phases")

	_, err = profile.ValidatorSetPhase(201)
	require.ErrorContains(t, err, "outside configured growth phases")

	_, err = profile.ValidatorSetPhase(301)
	require.ErrorContains(t, err, "100-300")
}

func TestMatureValidatorSetRequiresOperatorReadiness(t *testing.T) {
	profile := DefaultNetworkProfile()

	require.NoError(t, profile.ValidateMatureLaunch(128, false))
	require.NoError(t, profile.ValidateMatureLaunch(200, false))
	require.ErrorContains(t, profile.ValidateMatureLaunch(300, false), "operator readiness")
	require.NoError(t, profile.ValidateMatureLaunch(300, true))
}

func TestNetworkProfileRejectsUnsafePhasePolicy(t *testing.T) {
	profile := DefaultNetworkProfile()
	profile.ValidatorSetPhases[0].MaxActiveValidators = 80
	require.ErrorContains(t, profile.Validate(), "invalid validator range")

	profile = DefaultNetworkProfile()
	profile.ValidatorSetPhases[1].BlockTimeMinSeconds = 2
	profile.ValidatorSetPhases[1].TargetBlockTimeSeconds = 2
	require.ErrorContains(t, profile.Validate(), "invalid block time range")

	profile = DefaultNetworkProfile()
	profile.ValidatorSetPhases[1].TargetBlockTimeSeconds = 5
	require.ErrorContains(t, profile.Validate(), "invalid target block time")

	profile = DefaultNetworkProfile()
	profile.ValidatorSetPhases[2].NormalFinalityMaxSeconds = 30
	require.ErrorContains(t, profile.Validate(), "invalid finality range")
}

func TestValidatorSetLaunchAssessmentRejectsFiveHundredPlusAtStartup(t *testing.T) {
	profile := DefaultNetworkProfile()

	assessment := profile.AssessValidatorSetLaunch(500, true)

	require.False(t, assessment.Allowed)
	require.Equal(t, 500, assessment.ActiveValidators)
	require.Empty(t, assessment.Phase)
	require.Equal(t, "validator set 500+ is not an acceptable Aetra startup target", assessment.Reason)
	require.ElementsMatch(t, []string{
		AetraValidatorSetRiskConsensusOverhead,
		AetraValidatorSetRiskSyncComplexity,
		AetraValidatorSetRiskLatency,
		AetraValidatorSetRiskWeakOperators,
		AetraValidatorSetRiskInfrastructureQA,
	}, assessment.Risks)
}

func TestValidatorSetLaunchAssessmentAllowsPhasedGrowthOnly(t *testing.T) {
	profile := DefaultNetworkProfile()

	genesis := profile.AssessValidatorSetLaunch(128, false)
	require.True(t, genesis.Allowed)
	require.Equal(t, AetraValidatorPhaseGenesis, genesis.Phase)
	require.Empty(t, genesis.Risks)

	growth := profile.AssessValidatorSetLaunch(200, false)
	require.True(t, growth.Allowed)
	require.Equal(t, AetraValidatorPhaseGrowth, growth.Phase)
	require.Empty(t, growth.Risks)

	withoutReadiness := profile.AssessValidatorSetLaunch(300, false)
	require.False(t, withoutReadiness.Allowed)
	require.Equal(t, AetraValidatorPhaseMature, withoutReadiness.Phase)
	require.Contains(t, withoutReadiness.Risks, AetraValidatorSetRiskWeakOperators)
	require.Contains(t, withoutReadiness.Risks, AetraValidatorSetRiskInfrastructureQA)

	withReadiness := profile.AssessValidatorSetLaunch(300, true)
	require.True(t, withReadiness.Allowed)
	require.Equal(t, AetraValidatorPhaseMature, withReadiness.Phase)
	require.Empty(t, withReadiness.Risks)
}

func TestValidatorSetLaunchAssessmentRejectsPhaseGaps(t *testing.T) {
	profile := DefaultNetworkProfile()

	assessment := profile.AssessValidatorSetLaunch(201, true)

	require.False(t, assessment.Allowed)
	require.Empty(t, assessment.Phase)
	require.Contains(t, assessment.Reason, "outside configured growth phases")
}
