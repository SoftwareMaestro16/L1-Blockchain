package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	policykeeper "github.com/sovereign-l1/l1/x/aetra-staking-policy/keeper"
	"github.com/sovereign-l1/l1/x/aetra-staking-policy/types"
)

const authority = "ae1policygov"

func TestKeeperExportImportPreservesPolicyState(t *testing.T) {
	source := policykeeper.NewKeeper(authority)
	_, err := source.RecomputePolicy(1, []types.ValidatorStake{
		{OperatorAddress: "val-a", RawStake: 70, CommissionBps: 500},
		{OperatorAddress: "val-b", RawStake: 30, CommissionBps: 500},
	})
	require.NoError(t, err)
	msgServer := policykeeper.NewMsgServerImpl(&source)
	require.NoError(t, msgServer.RegisterValidatorIdentity(types.MsgRegisterValidatorIdentity{
		Authority:	authority,
		Identity:	types.ValidatorIdentityMetadata{OperatorAddress: "val-a", Moniker: "Aetra One", Website: "https://validator.example"},
	}))
	require.NoError(t, msgServer.AcknowledgeConcentrationWarning(types.MsgAcknowledgeConcentrationWarning{
		Authority:		authority,
		OperatorAddress:	"val-a",
		Warning:		types.DelegationWarningOverloaded,
		Height:			10,
	}))

	exported, err := source.ExportGenesis()
	require.NoError(t, err)
	require.NoError(t, exported.Validate())

	target := policykeeper.NewKeeper(authority)
	require.NoError(t, target.InitGenesis(exported))
	imported, err := target.ExportGenesis()
	require.NoError(t, err)
	require.Equal(t, exported, imported)
}

func TestGovernanceAuthorityRequiredForMessages(t *testing.T) {
	k := policykeeper.NewKeeper(authority)
	msgServer := policykeeper.NewMsgServerImpl(&k)
	params := types.DefaultParams(authority)
	params.CommissionFloorBps = 400

	err := msgServer.UpdateStakingPolicyParams(types.MsgUpdateStakingPolicyParams{
		Authority:	"ae1notgov",
		Params:		params,
	})
	require.ErrorIs(t, err, types.ErrUnauthorized)

	require.NoError(t, msgServer.UpdateStakingPolicyParams(types.MsgUpdateStakingPolicyParams{
		Authority:	authority,
		Params:		params,
	}))
	res, err := k.QueryParams(types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, uint32(400), res.Params.CommissionFloorBps)
}

func TestDeterministicQueriesExposeRequiredPolicyViews(t *testing.T) {
	first := policykeeper.NewKeeper(authority)
	second := policykeeper.NewKeeper(authority)
	input := []types.ValidatorStake{
		{OperatorAddress: "val-c", RawStake: 10, CommissionBps: 500},
		{OperatorAddress: "val-a", RawStake: 60, CommissionBps: 500},
		{OperatorAddress: "val-b", RawStake: 30, CommissionBps: 500},
	}
	reversed := []types.ValidatorStake{input[2], input[1], input[0]}

	_, err := first.RecomputePolicy(2, input)
	require.NoError(t, err)
	_, err = second.RecomputePolicy(2, reversed)
	require.NoError(t, err)

	stake, err := first.QueryValidatorStake(types.QueryValidatorStakeRequest{OperatorAddress: "val-a"})
	require.NoError(t, err)
	require.Equal(t, uint64(60), stake.RawStake)
	require.Positive(t, stake.OverflowStake)

	effective, err := first.QueryValidatorEffectivePower(types.QueryValidatorEffectivePowerRequest{OperatorAddress: "val-a"})
	require.NoError(t, err)
	require.Equal(t, stake.EffectiveStake, effective.EffectiveStake)

	top, err := first.QueryTopNConcentration(types.QueryTopNConcentrationRequest{N: 10})
	require.NoError(t, err)
	require.True(t, top.Exceeded)

	reward, err := first.QueryValidatorRewardMultiplier(types.QueryValidatorRewardMultiplierRequest{OperatorAddress: "val-a"})
	require.NoError(t, err)
	require.Less(t, reward.RewardMultiplierBps, types.BasisPoints)

	warning, err := first.QueryDelegationWarningStatus(types.QueryDelegationWarningStatusRequest{OperatorAddress: "val-a"})
	require.NoError(t, err)
	require.Equal(t, types.DelegationWarningOverloaded, warning.Warning)

	firstExport, err := first.ExportGenesis()
	require.NoError(t, err)
	secondExport, err := second.ExportGenesis()
	require.NoError(t, err)
	require.Equal(t, firstExport.Network, secondExport.Network)
}

func TestMarshalUnmarshalGenesisRoundTrip(t *testing.T) {
	source := policykeeper.NewKeeper(authority)
	_, err := source.RecomputePolicy(3, []types.ValidatorStake{
		{OperatorAddress: "val-a", RawStake: 100, CommissionBps: 500},
	})
	require.NoError(t, err)

	bz, err := source.MarshalGenesis()
	require.NoError(t, err)

	target := policykeeper.NewKeeper(authority)
	require.NoError(t, target.UnmarshalGenesis(bz))
	imported, err := target.ExportGenesis()
	require.NoError(t, err)
	exported, err := source.ExportGenesis()
	require.NoError(t, err)
	require.Equal(t, exported, imported)
}
