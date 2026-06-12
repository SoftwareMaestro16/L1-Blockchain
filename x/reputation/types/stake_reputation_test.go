package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
)

func TestStakeTimeReputationIncreasesWithPoolShareExposure(t *testing.T) {
	state := reputationStateForStakeTest(t)
	account := aeAddr(0x33)

	state, claim, err := ApplyClaimStakeReputation(state, MsgClaimStakeReputation{
		Authority:		state.Params.Authority,
		Account:		account,
		PoolID:			"pool-a",
		PoolShares:		100,
		PoolTotalShares:	1_000,
		PoolActiveStake:	10_000,
		TimestampUnix:		10,
	})
	require.NoError(t, err)
	require.Zero(t, claim.ReputationDelta)

	state, claim, err = ApplyClaimStakeReputation(state, MsgClaimStakeReputation{
		Authority:		state.Params.Authority,
		Account:		account,
		PoolID:			"pool-a",
		PoolShares:		100,
		PoolTotalShares:	1_000,
		PoolActiveStake:	10_000,
		TimestampUnix:		13,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(3_000), claim.StakeWeightedSeconds)
	require.Equal(t, uint16(3), claim.ReputationDelta)
	require.Equal(t, uint16(3), claim.ClaimedStakeReputation)

	record, stake, found := QueryAccountReputation(state, mustParseAE(t, account))
	require.True(t, found)
	require.Equal(t, uint16(3), record.StakingScore)
	require.Equal(t, account, stake.AccountUser)
}

func TestStakeReputationNoStakeTimeNoReputation(t *testing.T) {
	state := reputationStateForStakeTest(t)
	account := aeAddr(0x34)

	state, claim, err := ApplyClaimStakeReputation(state, MsgClaimStakeReputation{
		Authority:		state.Params.Authority,
		Account:		account,
		PoolID:			"pool-a",
		PoolShares:		0,
		PoolTotalShares:	1_000,
		PoolActiveStake:	10_000,
		TimestampUnix:		10,
	})
	require.NoError(t, err)
	require.Zero(t, claim.ReputationDelta)

	state, claim, err = ApplyClaimStakeReputation(state, MsgClaimStakeReputation{
		Authority:		state.Params.Authority,
		Account:		account,
		PoolID:			"pool-a",
		PoolShares:		0,
		PoolTotalShares:	1_000,
		PoolActiveStake:	10_000,
		TimestampUnix:		20,
	})
	require.NoError(t, err)
	require.Zero(t, claim.ReputationDelta)
	require.Zero(t, claim.StakeWeightedSeconds)
	require.NoError(t, state.Validate())
}

func TestStakeReputationClaimDeterministicGolden(t *testing.T) {
	state := reputationStateForStakeTest(t)
	account := aeAddr(0x35)
	state, _, err := ApplyClaimStakeReputation(state, MsgClaimStakeReputation{
		Authority:		state.Params.Authority,
		Account:		account,
		PoolID:			"pool-golden",
		PoolShares:		25,
		PoolTotalShares:	100,
		PoolActiveStake:	4_000,
		TimestampUnix:		100,
	})
	require.NoError(t, err)
	state, claim, err := ApplyClaimStakeReputation(state, MsgClaimStakeReputation{
		Authority:		state.Params.Authority,
		Account:		account,
		PoolID:			"pool-golden",
		PoolShares:		25,
		PoolTotalShares:	100,
		PoolActiveStake:	4_000,
		TimestampUnix:		107,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(7_000), claim.StakeWeightedSeconds)
	require.Equal(t, uint16(7), claim.ReputationDelta)
	require.Equal(t, "ef7adbc4d797595f556eabc21197fc85c97556a42fbc0c6c5dcf04da65a8be5b", claim.ClaimHash)
}

func TestSlashedOrJailedValidatorCannotReceivePositiveValidatorBonus(t *testing.T) {
	params := DefaultReputationParams()
	operatorExposure := StakePoolExposure{
		PoolID:			"validator-self",
		Shares:			100,
		TotalPoolShares:	100,
		PoolActiveStake:	1_000,
		LastUpdatedUnix:	1,
		ValidatorOperator:	true,
		ValidatorBonusBps:	params.ValidatorStakeBonusBps,
	}
	require.Equal(t, uint64(1_200), EffectivePoolStakeExposure(params, operatorExposure))

	operatorExposure.ValidatorJailed = true
	record := NewStakeReputationRecord(mustParseAE(t, aeAddr(0x36)))
	updated, err := AccumulateStakeExposure(params, record, operatorExposure)
	require.NoError(t, err)
	require.True(t, updated.PoolExposures[0].ValidatorBonusBlocked)
	require.Zero(t, updated.PoolExposures[0].ValidatorBonusBps)
	require.Zero(t, updated.PoolExposures[0].EffectiveStake)
}

func TestDelegatorOnJailedValidatorUsesConfiguredAllowedExposure(t *testing.T) {
	state := reputationStateForStakeTest(t)
	state.Params.JailedPoolExposureBps = 2_500
	account := aeAddr(0x37)
	state, _, err := ApplyClaimStakeReputation(state, MsgClaimStakeReputation{
		Authority:		state.Params.Authority,
		Account:		account,
		PoolID:			"pool-jailed",
		PoolShares:		50,
		PoolTotalShares:	100,
		PoolActiveStake:	8_000,
		TimestampUnix:		1,
		ValidatorJailed:	true,
	})
	require.NoError(t, err)
	state, claim, err := ApplyClaimStakeReputation(state, MsgClaimStakeReputation{
		Authority:		state.Params.Authority,
		Account:		account,
		PoolID:			"pool-jailed",
		PoolShares:		50,
		PoolTotalShares:	100,
		PoolActiveStake:	8_000,
		TimestampUnix:		2,
		ValidatorJailed:	true,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1_000), claim.StakeWeightedSeconds)
	require.Equal(t, uint16(1), claim.ReputationDelta)
}

func TestPoolUserReputationIsNotTiedToConcreteValidatorChoice(t *testing.T) {
	state := reputationStateForStakeTest(t)
	account := aeAddr(0x38)
	msg := MsgClaimStakeReputation{
		Authority:		state.Params.Authority,
		Account:		account,
		PoolID:			"official-liquid-pool",
		PoolShares:		10,
		PoolTotalShares:	100,
		PoolActiveStake:	10_000,
		TimestampUnix:		5,
	}
	state, _, err := ApplyClaimStakeReputation(state, msg)
	require.NoError(t, err)
	msg.TimestampUnix = 6
	state, claim, err := ApplyClaimStakeReputation(state, msg)
	require.NoError(t, err)
	require.Equal(t, uint16(1), claim.ReputationDelta)

	stake, found := QueryStakeReputation(state, mustParseAE(t, account))
	require.True(t, found)
	require.Equal(t, "official-liquid-pool", stake.PoolExposures[0].PoolID)
}

func TestStakeReputationExportImportPreservesAccumulator(t *testing.T) {
	state := reputationStateForStakeTest(t)
	account := aeAddr(0x39)
	state, _, err := ApplyClaimStakeReputation(state, MsgClaimStakeReputation{
		Authority:		state.Params.Authority,
		Account:		account,
		PoolID:			"pool-export",
		PoolShares:		1,
		PoolTotalShares:	1,
		PoolActiveStake:	2_000,
		TimestampUnix:		10,
	})
	require.NoError(t, err)
	state, _, err = ApplyClaimStakeReputation(state, MsgClaimStakeReputation{
		Authority:		state.Params.Authority,
		Account:		account,
		PoolID:			"pool-export",
		PoolShares:		1,
		PoolTotalShares:	1,
		PoolActiveStake:	2_000,
		TimestampUnix:		12,
	})
	require.NoError(t, err)

	exported, err := ExportReputationState(state)
	require.NoError(t, err)
	imported, err := ImportReputationState(exported)
	require.NoError(t, err)
	require.Equal(t, exported, imported)
}

func TestStakeReputationCannotBeMintedByMetadataOrTransferredAsset(t *testing.T) {
	record := NewStakeReputationRecord(mustParseAE(t, aeAddr(0x40)))
	require.ErrorContains(t, ValidateStakeReputationTransfer(record, "domain"), "cannot be transferred")

	state := reputationStateForStakeTest(t)
	metadataOnly := ApplyComputedScore(ReputationRecord{
		Account:	mustParseAE(t, aeAddr(0x41)),
		DomainScore:	MaxDomainScore,
	})
	state.Accounts = append(state.Accounts, metadataOnly)
	state = NormalizeReputationState(state)
	require.NoError(t, state.Validate())
	stake, found := QueryStakeReputation(state, metadataOnly.Account)
	require.False(t, found)
	require.Empty(t, stake.AccountUser)
}

func reputationStateForStakeTest(t *testing.T) ReputationState {
	t.Helper()
	params := DefaultReputationParams()
	params.Authority = aeAddr(0x01)
	params.StakeSecondsPerPoint = 1_000
	state, err := NewReputationState(params)
	require.NoError(t, err)
	return state
}

func aeAddr(fill byte) string {
	bz := make([]byte, 20)
	for i := range bz {
		bz[i] = fill
	}
	return aetraaddress.FormatAccAddress(sdk.AccAddress(bz))
}

func mustParseAE(t *testing.T, text string) sdk.AccAddress {
	t.Helper()
	addr, err := aetraaddress.ParseUserAddress("test account", text)
	require.NoError(t, err)
	return addr
}
