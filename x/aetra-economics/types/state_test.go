package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/x/aetra-economics/types"
)

const authority = "ae1economicsgov"

func TestInflationCurveRespondsToBondedRatio(t *testing.T) {
	params := types.DefaultParams(authority)
	atTarget := types.ComputeInflationBps(params, params.TargetBondedRatioBps)
	lowBonded := types.ComputeInflationBps(params, 3_000)
	highBonded := types.ComputeInflationBps(params, 8_000)

	require.Greater(t, lowBonded, atTarget)
	require.Less(t, highBonded, atTarget)
	require.Equal(t, uint32(350), atTarget)
}

func TestBoundedInflation(t *testing.T) {
	params := types.DefaultParams(authority)
	require.Equal(t, params.InflationMaxBps, types.ComputeInflationBps(params, 0))
	require.Equal(t, params.InflationMinBps, types.ComputeInflationBps(params, types.BasisPoints))
}

func TestInflationChangePerEpochIsBounded(t *testing.T) {
	params := types.DefaultParams(authority)
	params.InflationChangeRateBps = 25

	increase := types.ComputeNextInflationBps(params, 350, 0)
	decrease := types.ComputeNextInflationBps(params, 350, types.BasisPoints)

	require.Equal(t, uint32(375), increase)
	require.Equal(t, uint32(325), decrease)
	require.LessOrEqual(t, increase-350, params.InflationChangeRateBps)
	require.LessOrEqual(t, uint32(350)-decrease, params.InflationChangeRateBps)
}

func TestApplyEpochUsesBoundedInflationStep(t *testing.T) {
	params := fastEpochParams()
	params.InflationChangeRateBps = 10
	state := types.DefaultGenesisState(authority).State
	state.CurrentInflationBps = 350

	next, summary, err := types.ApplyEpoch(params, state, epochInput(1, 1_000_000_000, 0, 0))
	require.NoError(t, err)
	require.Equal(t, uint32(360), summary.InflationBps)
	require.Equal(t, summary.InflationBps, next.CurrentInflationBps)
	require.LessOrEqual(t, summary.InflationBps-state.CurrentInflationBps, params.InflationChangeRateBps)
}

func TestFeeSplitAccounting(t *testing.T) {
	split, err := types.ComputeFeeSplit(types.DefaultParams(authority), 1_000_000)
	require.NoError(t, err)
	require.Equal(t, uint32(5_000), split.BurnBps)
	require.Equal(t, uint32(3_500), split.ValidatorRewardBps)
	require.Equal(t, uint32(1_500), split.TreasuryBps)
	require.Equal(t, uint64(500_000), split.BurnAmount)
	require.Equal(t, uint64(150_000), split.TreasuryAmount)
	require.Equal(t, uint64(350_000), split.ValidatorDelegatorRewards)
	require.Equal(t, split.FeesCollected, split.BurnAmount+split.TreasuryAmount+split.ValidatorDelegatorRewards)
}

func TestEconomicsAuthorityPathIsSinglePolicySource(t *testing.T) {
	path := types.DefaultEconomicsAuthorityPath()
	require.NoError(t, types.ValidateEconomicsAuthorityPath(path))

	authoritative := make([]string, 0)
	execution := map[string]bool{}
	for _, role := range path {
		if role.AuthoritativePolicy {
			authoritative = append(authoritative, role.Module)
		}
		if role.ExecutionAccounting {
			execution[role.Module] = true
		}
	}
	require.Equal(t, []string{types.EconomicsAuthoritativePolicyModule}, authoritative)
	require.True(t, execution[types.EconomicsFeesModule])
	require.True(t, execution[types.EconomicsFeeCollectorModule])
	require.True(t, execution[types.EconomicsBurnModule])
	require.True(t, execution[types.EconomicsTreasuryModule])
	require.True(t, execution[types.EconomicsEmissionsModule])
	require.True(t, execution[types.EconomicsMintAuthorityModule])
}

func TestEconomicsAuthorityPathRejectsSplitPolicySources(t *testing.T) {
	path := types.DefaultEconomicsAuthorityPath()
	path[1].AuthoritativePolicy = true
	require.ErrorContains(t, types.ValidateEconomicsAuthorityPath(path), "exactly one authoritative")

	path = types.DefaultEconomicsAuthorityPath()
	path = path[1:]
	require.ErrorContains(t, types.ValidateEconomicsAuthorityPath(path), "authoritative policy module")

	path = types.DefaultEconomicsAuthorityPath()
	path[len(path)-1].ExecutionAccounting = false
	require.ErrorContains(t, types.ValidateEconomicsAuthorityPath(path), types.EconomicsMintAuthorityModule)
}

func TestFeeSplitParamsRejectUnsafeRules(t *testing.T) {
	params := types.DefaultParams(authority)
	params.TreasuryBps = 1_400
	require.ErrorContains(t, params.Validate(), "sum")

	params = types.DefaultParams(authority)
	params.BurnCurrentBps = params.BurnMaxBps + 1
	params.ValidatorRewardBps = 3_499
	require.ErrorContains(t, params.Validate(), "burn")

	params = types.DefaultParams(authority)
	params.TreasuryBps = params.TreasuryMaxBps + 1
	params.ValidatorRewardBps = 2_999
	require.ErrorContains(t, params.Validate(), "treasury")

	params = types.DefaultParams(authority)
	params.BurnCurrentBps = 8_500
	params.BurnMaxBps = 8_500
	params.ValidatorRewardBps = 0
	params.TreasuryBps = 1_500
	require.ErrorContains(t, params.Validate(), "zero without emergency")

	params.EmergencyAllowZeroRewardShare = true
	require.NoError(t, params.Validate())
}

func TestBurnTreasuryAndSupplyAccounting(t *testing.T) {
	params := fastEpochParams()
	state := types.DefaultGenesisState(authority).State
	next, summary, err := types.ApplyEpoch(params, state, epochInput(1, 1_000_000_000, 600_000_000, 100_000))
	require.NoError(t, err)

	require.Equal(t, uint64(50_000), summary.BurnedAmount)
	require.Equal(t, uint64(15_000), summary.TreasuryAmount)
	require.Equal(t, summary.BurnedAmount, next.BurnedSupply)
	require.Equal(t, summary.TreasuryAmount, next.TreasuryBalance)
	require.Equal(t, summary.StartingSupply+summary.MintedRewards-summary.BurnedAmount, summary.EndingSupply)
}

func TestAPREstimate(t *testing.T) {
	require.Equal(t, uint32(667), types.EstimateAPRBps(400, 6_000))
	require.Equal(t, uint32(0), types.EstimateAPRBps(400, 0))
}

func TestAPRBreakdownLabelsEstimateAndCommissionImpact(t *testing.T) {
	params := fastEpochParams()
	state, summary, err := types.ApplyEpoch(params, types.DefaultGenesisState(authority).State, epochInput(1, 1_000_000_000, 600_000_000, 100_000))
	require.NoError(t, err)

	apr, err := types.EstimateAPRBreakdown(params, state, types.QueryEstimatedAPRRequest{
		ValidatorCommissionBps:		1_000,
		ValidatorOperatingCostBps:	50,
	})
	require.NoError(t, err)
	require.True(t, apr.IsEstimate)
	require.Equal(t, "estimate_not_guaranteed_return", apr.EstimateLabel)
	require.Equal(t, summary.EstimatedAPRBps, apr.InflationOnlyAPRBps)
	require.Greater(t, apr.FeeAdjustedAPRBps, apr.InflationOnlyAPRBps)
	require.Equal(t, uint32(64), apr.ValidatorCommissionImpactBps)
	require.Equal(t, apr.FeeAdjustedAPRBps-apr.ValidatorCommissionImpactBps, apr.EstimatedDelegatorAPRBps)
	require.Equal(t, apr.FeeAdjustedAPRBps+apr.ValidatorCommissionImpactBps, apr.EstimatedValidatorGrossAPRBps)
	require.Equal(t, apr.EstimatedValidatorGrossAPRBps-50, apr.EstimatedValidatorNetAPRBps)
}

func TestEpochRewardSmoothing(t *testing.T) {
	params := fastEpochParams()
	params.RewardSmoothingWindow = 2
	state := types.DefaultGenesisState(authority).State

	next, first, err := types.ApplyEpoch(params, state, epochInput(1, 1_000_000_000, 600_000_000, 0))
	require.NoError(t, err)
	next, second, err := types.ApplyEpoch(params, next, epochInput(2, first.EndingSupply, 600_000_000, 1_000_000))
	require.NoError(t, err)

	require.Equal(t, (first.GrossRewards+second.GrossRewards)/2, second.SmoothedRewards)
	require.Len(t, next.RewardHistory, 2)
}

func TestZeroFeeBlockHandling(t *testing.T) {
	params := fastEpochParams()
	_, summary, err := types.ApplyEpoch(params, types.DefaultGenesisState(authority).State, epochInput(1, 1_000_000_000, 600_000_000, 0))
	require.NoError(t, err)
	require.Equal(t, uint64(0), summary.BurnedAmount)
	require.Equal(t, uint64(0), summary.TreasuryAmount)
	require.Equal(t, uint64(0), summary.ValidatorDelegatorRewards)
	require.Equal(t, summary.MintedRewards, summary.GrossRewards)
	require.Equal(t, summary.StartingSupply+summary.MintedRewards, summary.EndingSupply)
}

func TestHighFeeBlockHandling(t *testing.T) {
	params := fastEpochParams()
	_, summary, err := types.ApplyEpoch(params, types.DefaultGenesisState(authority).State, epochInput(1, 1_000_000_000, 600_000_000, 10_000_000))
	require.NoError(t, err)
	require.Equal(t, uint64(5_000_000), summary.BurnedAmount)
	require.Equal(t, uint64(1_500_000), summary.TreasuryAmount)
	require.Equal(t, uint64(3_500_000), summary.ValidatorDelegatorRewards)
	require.Equal(t, summary.MintedRewards+summary.ValidatorDelegatorRewards, summary.GrossRewards)
	require.Equal(t, int64(summary.MintedRewards)-int64(summary.BurnedAmount), summary.NetSupplyChange)
}

func TestSupplyInvariantForEpochSummary(t *testing.T) {
	params := fastEpochParams()
	_, summary, err := types.ApplyEpoch(params, types.DefaultGenesisState(authority).State, epochInput(1, 1_000_000_000, 600_000_000, 500_000))
	require.NoError(t, err)
	require.NoError(t, summary.Validate(params))
	require.Equal(t, summary.FeesCollected, summary.BurnedAmount+summary.TreasuryAmount+summary.ValidatorDelegatorRewards)
	require.Equal(t, summary.EndingSupply, summary.StartingSupply+summary.MintedRewards-summary.BurnedAmount)
}

func TestSupplyInvariantAfterManyEpochs(t *testing.T) {
	params := fastEpochParams()
	state := types.DefaultGenesisState(authority).State
	supply := uint64(1_000_000_000)

	for epoch := uint64(1); epoch <= 50; epoch++ {
		bonded := supply * 60 / 100
		fees := epoch * 10_000
		next, summary, err := types.ApplyEpoch(params, state, epochInput(epoch, supply, bonded, fees))
		require.NoError(t, err)
		require.NoError(t, summary.Validate(params))
		require.Equal(t, summary.StartingSupply+summary.MintedRewards-summary.BurnedAmount, summary.EndingSupply)
		require.Equal(t, summary.EndingSupply, next.TotalSupply)
		require.Equal(t, summary.BurnedSupply, next.BurnedSupply)
		require.Equal(t, summary.TreasuryBalance, next.TreasuryBalance)
		state = next
		supply = next.TotalSupply
	}

	require.NoError(t, state.Validate(params))
}

func TestGenesisExportStateValidation(t *testing.T) {
	params := fastEpochParams()
	state, _, err := types.ApplyEpoch(params, types.DefaultGenesisState(authority).State, epochInput(1, 1_000_000_000, 600_000_000, 100_000))
	require.NoError(t, err)
	genesis := types.GenesisState{Params: params, State: state}
	require.NoError(t, genesis.Validate())
}

func fastEpochParams() types.Params {
	params := types.DefaultParams(authority)
	params.EpochsPerYear = 100
	return params
}

func epochInput(epoch, supply, bonded, fees uint64) types.EpochEconomicsInput {
	return types.EpochEconomicsInput{
		Epoch:		epoch,
		TotalSupply:	supply,
		BondedTokens:	bonded,
		FeesCollected:	fees,
	}
}
