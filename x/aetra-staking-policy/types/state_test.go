package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

const testAuthority = "ae1policygov"

func TestCapCalculationTracksRawEffectiveAndOverflowStake(t *testing.T) {
	params := DefaultParams(testAuthority)
	network, err := ComputeNetworkPolicy(params, 1, []ValidatorStake{
		{OperatorAddress: "val-a", RawStake: 500, CommissionBps: 500},
		{OperatorAddress: "val-b", RawStake: 500, CommissionBps: 500},
	}, nil)
	require.NoError(t, err)

	validator := network.Validators[0]
	require.Equal(t, "val-a", validator.OperatorAddress)
	require.Equal(t, uint32(5_000), validator.RawPowerBps)
	require.Equal(t, uint32(300), validator.EffectivePowerBps)
	require.Equal(t, uint64(30), validator.EffectiveStake)
	require.Equal(t, uint64(470), validator.OverflowStake)
	require.Equal(t, DelegationWarningOverloaded, validator.DelegationWarning)
	require.Less(t, validator.RewardMultiplierBps, BasisPoints)
}

func TestCapTransitionAtOneHundredFiftyAndTwoHundredFiftyValidators(t *testing.T) {
	params := DefaultParams(testAuthority)

	require.Equal(t, PhaseOnePowerCapBps, EffectivePowerCapBps(params, 100))
	require.Equal(t, PhaseOnePowerCapBps, EffectivePowerCapBps(params, 150))
	require.Equal(t, PhaseTwoPowerCapBps, EffectivePowerCapBps(params, 151))
	require.Equal(t, PhaseTwoPowerCapBps, EffectivePowerCapBps(params, 250))
	require.Equal(t, MatureSetPowerCapBps, EffectivePowerCapBps(params, 251))
	require.Equal(t, MatureSetPowerCapBps, EffectivePowerCapBps(params, 300))
}

func TestCommissionBoundsAndDailyChangeAreEnforced(t *testing.T) {
	params := DefaultParams(testAuthority)

	_, err := ComputeNetworkPolicy(params, 1, []ValidatorStake{
		{OperatorAddress: "val-a", RawStake: 1_000, CommissionBps: params.CommissionFloorBps - 1},
	}, nil)
	require.ErrorIs(t, err, ErrInvalidPolicy)

	_, err = ComputeNetworkPolicy(params, 1, []ValidatorStake{
		{OperatorAddress: "val-a", RawStake: 1_000, CommissionBps: 500, PreviousCommissionBps: 300},
	}, nil)
	require.ErrorIs(t, err, ErrInvalidPolicy)
}

func TestTopNConcentrationTargets(t *testing.T) {
	params := DefaultParams(testAuthority)
	validators := make([]ValidatorStake, 40)
	for i := range validators {
		stake := uint64(1)
		if i < 10 {
			stake = 3
		}
		validators[i] = ValidatorStake{OperatorAddress: fmt.Sprintf("val-%03d", i), RawStake: stake, CommissionBps: 500}
	}

	network, err := ComputeNetworkPolicy(params, 1, validators, nil)
	require.NoError(t, err)

	require.Equal(t, uint32(5_000), network.Top10PowerBps)
	require.True(t, network.ConcentrationWarn)
}

func TestGenesisValidationRejectsUnsortedValidators(t *testing.T) {
	params := DefaultParams(testAuthority)
	gs := GenesisState{
		Params:	params,
		Network: NetworkPolicy{
			ActiveValidators:	2,
			TotalRawStake:		2,
			PowerCapBps:		PhaseOnePowerCapBps,
			Validators: []ValidatorPolicy{
				{OperatorAddress: "val-b", RawStake: 1, EffectiveStake: 1, RawPowerBps: 5_000, EffectivePowerBps: 300, PowerCapBps: 300, RewardMultiplierBps: BasisPoints, DelegationWarning: DelegationWarningOverloaded, CommissionAllowed: true},
				{OperatorAddress: "val-a", RawStake: 1, EffectiveStake: 1, RawPowerBps: 5_000, EffectivePowerBps: 300, PowerCapBps: 300, RewardMultiplierBps: BasisPoints, DelegationWarning: DelegationWarningOverloaded, CommissionAllowed: true},
			},
		},
	}

	require.ErrorContains(t, gs.Validate(), "sorted")
}

func FuzzEffectivePowerNeverExceedsCap(f *testing.F) {
	f.Add(uint64(1), uint64(1), uint32(100))
	f.Add(uint64(10_000), uint64(1), uint32(151))
	f.Add(uint64(1_000_000), uint64(999_999), uint32(300))

	f.Fuzz(func(t *testing.T, whaleStake uint64, peerStake uint64, validatorCount uint32) {
		if validatorCount == 0 {
			validatorCount = 1
		}
		if validatorCount > 400 {
			validatorCount = validatorCount%400 + 1
		}
		if whaleStake == 0 {
			whaleStake = 1
		}
		if peerStake == 0 {
			peerStake = 1
		}
		validators := []ValidatorStake{{OperatorAddress: "val-000", RawStake: whaleStake, CommissionBps: 500}}
		for i := uint32(1); i < validatorCount; i++ {
			validators = append(validators, ValidatorStake{OperatorAddress: fmt.Sprintf("val-%03d", i), RawStake: peerStake, CommissionBps: 500})
		}

		network, err := ComputeNetworkPolicy(DefaultParams(testAuthority), 1, validators, nil)
		if err != nil {
			t.Skip(err)
		}
		for _, validator := range network.Validators {
			require.LessOrEqual(t, validator.EffectivePowerBps, validator.PowerCapBps)
			require.LessOrEqual(t, validator.EffectiveStake, validator.RawStake)
			require.Equal(t, validator.RawStake, validator.EffectiveStake+validator.OverflowStake)
		}
	})
}
