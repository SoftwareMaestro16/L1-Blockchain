package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestComputeScoreClampsAndCapsDomainInfluence(t *testing.T) {
	record := ReputationRecord{
		Account:		addr(1),
		AgeScore:		15,
		StakingScore:		20,
		TxSuccessScore:		25,
		VolumeScore:		20,
		DomainScore:		100,
		ContractScore:		100,
		SpamPenalty:		5,
		FailedTxPenalty:	3,
		SlashPenalty:		2,
	}
	require.Equal(t, uint8(95), ComputeScore(record))

	record.SpamPenalty = 200
	require.Equal(t, uint8(0), ComputeScore(record))

	record.SpamPenalty = 0
	record.AgeScore = 100
	record.StakingScore = 100
	require.Equal(t, uint8(100), ComputeScore(record))
}

func TestLevelForScore(t *testing.T) {
	require.Equal(t, LevelRestricted, LevelForScore(0))
	require.Equal(t, LevelRestricted, LevelForScore(19))
	require.Equal(t, LevelNew, LevelForScore(20))
	require.Equal(t, LevelNew, LevelForScore(49))
	require.Equal(t, LevelNormal, LevelForScore(50))
	require.Equal(t, LevelNormal, LevelForScore(79))
	require.Equal(t, LevelTrusted, LevelForScore(80))
	require.Equal(t, LevelTrusted, LevelForScore(94))
	require.Equal(t, LevelElite, LevelForScore(95))
	require.Equal(t, LevelElite, LevelForScore(100))
}

func TestApplyInactivityDecay(t *testing.T) {
	params := DecayParams{InactiveAfterEpochs: 2, DecayRatePerEpoch: 3}
	require.Equal(t, uint8(80), ApplyInactivityDecay(80, 2, params))
	require.Equal(t, uint8(77), ApplyInactivityDecay(80, 3, params))
	require.Equal(t, uint8(0), ApplyInactivityDecay(5, 10, params))
}

func TestValidateReputationRecord(t *testing.T) {
	record := ApplyComputedScore(ReputationRecord{
		Account:	addr(1),
		AgeScore:	10,
		StakingScore:	10,
		TxSuccessScore:	10,
	})
	require.NoError(t, ValidateReputationRecord(record))

	record.Score++
	require.ErrorContains(t, ValidateReputationRecord(record), "score mismatch")

	record = ApplyComputedScore(ReputationRecord{Account: make([]byte, 20)})
	require.ErrorContains(t, ValidateReputationRecord(record), "reputation account")

	record = ApplyComputedScore(ReputationRecord{Account: addr(1), DomainScore: MaxDomainScore + 1})
	require.ErrorContains(t, ValidateReputationRecord(record), "domain score")
}

func TestProgressiveLimitsByLevel(t *testing.T) {
	require.Equal(t, uint32(1), LimitsForScore(0).MaxTxsPerBlock)
	require.Equal(t, uint32(5), LimitsForScore(20).MaxTxsPerBlock)
	require.Equal(t, uint32(25), LimitsForScore(50).MaxTxsPerBlock)
	require.Equal(t, uint32(100), LimitsForScore(80).MaxTxsPerBlock)
	require.Equal(t, uint32(250), LimitsForScore(95).MaxTxsPerBlock)
	require.False(t, IsDirectReputationPurchaseAllowed())
}

func addr(seed byte) sdk.AccAddress {
	out := make([]byte, 20)
	out[19] = seed
	return sdk.AccAddress(out)
}
