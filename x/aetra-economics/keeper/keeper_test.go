package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	economicskeeper "github.com/sovereign-l1/l1/x/aetra-economics/keeper"
	"github.com/sovereign-l1/l1/x/aetra-economics/types"
)

const authority = "ae1economicsgov"

func TestKeeperQueriesExposeEconomicsState(t *testing.T) {
	k := economicskeeper.NewKeeper(authority)
	params := fastEpochParams()
	require.NoError(t, k.SetParams(params))

	summary, err := k.ApplyEpoch(epochInput(1, 1_000_000_000, 600_000_000, 100_000))
	require.NoError(t, err)

	inflation, err := k.QueryCurrentInflation(types.QueryCurrentInflationRequest{})
	require.NoError(t, err)
	require.Equal(t, summary.InflationBps, inflation.InflationBps)

	bonded, err := k.QueryCurrentBondedRatio(types.QueryCurrentBondedRatioRequest{})
	require.NoError(t, err)
	require.Equal(t, uint32(6_000), bonded.BondedRatioBps)

	apr, err := k.QueryEstimatedAPR(types.QueryEstimatedAPRRequest{ValidatorCommissionBps: 1_000, ValidatorOperatingCostBps: 50})
	require.NoError(t, err)
	require.True(t, apr.IsEstimate)
	require.Equal(t, "estimate_not_guaranteed_return", apr.EstimateLabel)
	require.Equal(t, summary.EstimatedAPRBps, apr.InflationOnlyAPRBps)
	require.GreaterOrEqual(t, apr.FeeAdjustedAPRBps, apr.InflationOnlyAPRBps)
	require.Greater(t, apr.ValidatorCommissionImpactBps, uint32(0))
	require.Less(t, apr.EstimatedDelegatorAPRBps, apr.FeeAdjustedAPRBps)
	require.Greater(t, apr.EstimatedValidatorGrossAPRBps, apr.FeeAdjustedAPRBps)
	require.Less(t, apr.EstimatedValidatorNetAPRBps, apr.EstimatedValidatorGrossAPRBps)

	feeSplit, err := k.QueryFeeSplitParams(types.QueryFeeSplitParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, params.BurnCurrentBps, feeSplit.BurnCurrentBps)
	require.Equal(t, params.ValidatorRewardMinBps, feeSplit.ValidatorRewardMinBps)
	require.Equal(t, params.ValidatorRewardMaxBps, feeSplit.ValidatorRewardMaxBps)
	require.Equal(t, params.TreasuryMinBps, feeSplit.TreasuryMinBps)
	require.Equal(t, params.TreasuryMaxBps, feeSplit.TreasuryMaxBps)

	burned, err := k.QueryBurnedSupply(types.QueryBurnedSupplyRequest{})
	require.NoError(t, err)
	require.Equal(t, summary.BurnedSupply, burned.BurnedSupply)

	treasury, err := k.QueryTreasuryBalance(types.QueryTreasuryBalanceRequest{})
	require.NoError(t, err)
	require.Equal(t, summary.TreasuryBalance, treasury.TreasuryBalance)

	epoch, err := k.QueryEpochRewardSummary(types.QueryEpochRewardSummaryRequest{Epoch: 1})
	require.NoError(t, err)
	require.Equal(t, summary, epoch.Summary)
}

func TestKeeperExportImportPreservesEconomicsState(t *testing.T) {
	source := economicskeeper.NewKeeper(authority)
	require.NoError(t, source.SetParams(fastEpochParams()))
	_, err := source.ApplyEpoch(epochInput(1, 1_000_000_000, 600_000_000, 100_000))
	require.NoError(t, err)

	exported, err := source.ExportGenesis()
	require.NoError(t, err)
	require.NoError(t, exported.Validate())

	target := economicskeeper.NewKeeper(authority)
	require.NoError(t, target.InitGenesis(exported))
	imported, err := target.ExportGenesis()
	require.NoError(t, err)
	require.Equal(t, exported, imported)
}

func TestMarshalUnmarshalGenesisRoundTrip(t *testing.T) {
	source := economicskeeper.NewKeeper(authority)
	require.NoError(t, source.SetParams(fastEpochParams()))
	_, err := source.ApplyEpoch(epochInput(1, 1_000_000_000, 600_000_000, 100_000))
	require.NoError(t, err)

	bz, err := source.MarshalGenesis()
	require.NoError(t, err)

	target := economicskeeper.NewKeeper(authority)
	require.NoError(t, target.UnmarshalGenesis(bz))
	imported, err := target.ExportGenesis()
	require.NoError(t, err)
	exported, err := source.ExportGenesis()
	require.NoError(t, err)
	require.Equal(t, exported, imported)
}

func TestGovernanceAuthorityRequiredForMessages(t *testing.T) {
	k := economicskeeper.NewKeeper(authority)
	msgServer := economicskeeper.NewMsgServerImpl(&k)
	params := fastEpochParams()
	params.BurnCurrentBps = 5_000
	params.ValidatorRewardBps = 3_500
	params.TreasuryBps = 1_500

	err := msgServer.UpdateEconomicsParams(types.MsgUpdateEconomicsParams{
		Authority:	"ae1notgov",
		Params:		params,
	})
	require.ErrorIs(t, err, types.ErrUnauthorized)

	require.NoError(t, msgServer.UpdateEconomicsParams(types.MsgUpdateEconomicsParams{
		Authority:	authority,
		Params:		params,
	}))
	feeSplit, err := k.QueryFeeSplitParams(types.QueryFeeSplitParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, uint32(5_000), feeSplit.BurnCurrentBps)

	err = msgServer.ApplyEpochEconomics(types.MsgApplyEpochEconomics{
		Authority:	"ae1notgov",
		Input:		epochInput(1, 1_000_000_000, 600_000_000, 100_000),
	})
	require.ErrorIs(t, err, types.ErrUnauthorized)

	require.NoError(t, msgServer.ApplyEpochEconomics(types.MsgApplyEpochEconomics{
		Authority:	authority,
		Input:		epochInput(1, 1_000_000_000, 600_000_000, 100_000),
	}))
}

func TestGovernanceInvalidParamsRejected(t *testing.T) {
	k := economicskeeper.NewKeeper(authority)
	msgServer := economicskeeper.NewMsgServerImpl(&k)
	params := fastEpochParams()
	params.BurnCurrentBps = 6_001
	params.ValidatorRewardBps = 2_499

	err := msgServer.UpdateEconomicsParams(types.MsgUpdateEconomicsParams{
		Authority:	authority,
		Params:		params,
	})
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrInvalidParams)
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
