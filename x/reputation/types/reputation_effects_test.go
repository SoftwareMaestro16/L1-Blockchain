package types

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
)

func TestLowReputationHigherFeeUnsignedInt(t *testing.T) {
	effectParams := DefaultReputationEffectParams()

	_, discountHigh := ComputeBoundedFeeAdjustment(9000, true, effectParams)
	premiumLow, _ := ComputeBoundedFeeAdjustment(1000, true, effectParams)
	_, discountNeutral := ComputeBoundedFeeAdjustment(5000, true, effectParams)
	premiumNeutral, _ := ComputeBoundedFeeAdjustment(5000, true, effectParams)

	require.Greater(t, premiumLow, int64(0))
	require.Equal(t, int64(0), premiumNeutral)
	require.Greater(t, discountHigh, int64(0))
	require.Equal(t, int64(0), discountNeutral)
}

func TestSpamTxReducesScoreDeterministically(t *testing.T) {
	rep := NewIdentityReputation(aetraaddress.FormatAccAddress(bytes20(0x77)))
	rep.RecordSuccessfulTx(10)
	rep.RecordSuccessfulTx(11)
	rep.RecordStakeTime(86400, 12)
	rep.Score = ComputeIdentityScore(rep)
	beforeScore := rep.Score

	effectParams := DefaultReputationEffectParams()
	signal := ReputationSignal{
		ProviderType:	SignalProviderUser,
		ProviderID:	rep.Account,
		SignalType:	SignalTypeTxSpam,
		Height:		100,
	}
	updated, err := ApplyIdentitySignal(rep, signal, effectParams)
	require.NoError(t, err)
	require.Less(t, updated.Score, beforeScore)
}

func TestStakeTimeClaimIncreasesScoreWithCap(t *testing.T) {
	rep := NewIdentityReputation(aetraaddress.FormatAccAddress(bytes20(0x88)))
	rep.Score = ComputeIdentityScore(rep)
	beforeScore := rep.Score

	rep.RecordStakeTime(3600*50, 10)
	rep.Score = ComputeIdentityScore(rep)
	require.GreaterOrEqual(t, rep.Score, beforeScore)
}

func TestStakeTimeDoubleClaimRejected(t *testing.T) {
	rep := NewIdentityReputation(aetraaddress.FormatAccAddress(bytes20(0x99)))
	rep.RecordStakeTime(1000, 10)
	require.Equal(t, uint64(1000), rep.ClaimedStakeTimeSeconds)
	require.Equal(t, uint64(1000), rep.SignalCounters.StakeTimeSeconds)

	rep.RecordStakeTime(500, 15)
	require.Equal(t, uint64(1000), rep.ClaimedStakeTimeSeconds, "smaller claim rejected")
	require.Equal(t, uint64(1000), rep.SignalCounters.StakeTimeSeconds)
}

func TestJailedSlashedValidatorZeroScore(t *testing.T) {
	vs := NewValidatorScore("val-jailed")
	vs.UptimeScore = 800
	vs.CommissionBehavior = 500
	vs.IsJailed = true
	vs.TotalScore = ComputeValidatorTotalScore(vs)
	require.Equal(t, uint32(0), vs.TotalScore)

	effectParams := DefaultReputationEffectParams()
	weight := ComputePoolAllocationWeight(vs, effectParams)
	require.Equal(t, uint32(0), weight)
}

func TestValidatorScoreAffectsPoolAllocationDeterministically(t *testing.T) {
	effectParams := DefaultReputationEffectParams()

	vsHigh := NewValidatorScore("val-high")
	vsHigh.UptimeScore = 900
	vsHigh.CommissionBehavior = 800
	vsHigh.GovernanceParticipation = 500
	vsHigh.PoolAllocationScore = 600
	vsHigh.TotalScore = ComputeValidatorTotalScore(vsHigh)

	vsLow := NewValidatorScore("val-low")
	vsLow.UptimeScore = 200
	vsLow.CommissionBehavior = 100
	vsLow.GovernanceParticipation = 50
	vsLow.PoolAllocationScore = 100
	vsLow.TotalScore = ComputeValidatorTotalScore(vsLow)

	wHigh := ComputePoolAllocationWeight(vsHigh, effectParams)
	wLow := ComputePoolAllocationWeight(vsLow, effectParams)
	require.Greater(t, wHigh, wLow)

	wHigh2 := ComputePoolAllocationWeight(vsHigh, effectParams)
	require.Equal(t, wHigh, wHigh2)
}

func TestServiceTrustAffectsRoutingOnly(t *testing.T) {
	effectParams := DefaultReputationEffectParams()
	sts := NewServiceTrustScore("rpc-node-1")
	sts.Trust = 8000
	sts.Reliability = 7500

	boost := ComputeServiceTrustRoutingBoost(sts, effectParams)
	require.Greater(t, boost, uint32(0))
	require.LessOrEqual(t, boost, effectParams.MaxServiceTrustRoutingBps)
	require.True(t, ServiceTrustCannotMoveFunds(sts))
	require.True(t, ServiceTrustCannotBypassFees(sts))
}

func TestContractExecutionUpdatesCallerSignalCounters(t *testing.T) {
	caller := NewIdentityReputation(aetraaddress.FormatAccAddress(bytes20(0x55)))
	before := caller.SignalCounters.ContractInteractions

	ApplyIdentityContractOutcome(caller, true, 100)
	require.Equal(t, before+1, caller.SignalCounters.ContractInteractions)

	ApplyIdentityContractOutcome(caller, false, 101)
	require.Equal(t, uint64(1), caller.SignalCounters.ContractFailures)
}

func TestExportImportPreservesAllIdentityState(t *testing.T) {
	id := NewIdentityReputation(aetraaddress.FormatAccAddress(bytes20(0x66)))
	id.RecordSuccessfulTx(10)
	id.RecordFailedTx(11)
	id.RecordContractInteraction(12)
	id.RecordStakeTime(86400, 13)
	id.RecordSpam(14)
	id.RecordSlashEvent(15)
	id.RecordRecoveryEvent(16)
	id.Score = ComputeIdentityScore(id)
	id.Confidence = ComputeConfidence(id)
	id.DecayEpoch = 5

	claim := id.ExportClaim(100)
	imported, err := ImportReputationFromClaim(claim)
	require.NoError(t, err)
	require.Equal(t, id.Account, imported.Account)
	require.Equal(t, id.Score, imported.Score)
	require.Equal(t, id.Confidence, imported.Confidence)
	require.Equal(t, id.StakeTimeAccumulator, imported.StakeTimeAccumulator)
	require.Equal(t, id.ClaimedStakeTimeSeconds, imported.ClaimedStakeTimeSeconds)
	require.Equal(t, id.DecayEpoch, imported.DecayEpoch)
	require.Equal(t, id.SignalCounters.SuccessfulTxs, imported.SignalCounters.SuccessfulTxs)
}

func TestGoldenDeterministicFeePriorityAllocation(t *testing.T) {
	effParams := DefaultReputationEffectParams()

	type state struct {
		score	uint32
		found	bool
	}

	states := []state{
		{100, true},
		{1000, true},
		{5000, true},
		{9000, true},
	}

	type result struct {
		FeeAdjusted	uint64
		QueueBoost	uint32
		QueuePenalty	uint32
	}

	baseFee := uint64(10_000_000)
	results1 := make([]result, len(states))
	results2 := make([]result, len(states))

	for i, s := range states {
		fee := ComputeBoundedFeeAdjustmentNaet(baseFee, s.score, s.found, effParams)
		boost := ComputeQueuePriorityBoost(s.score, effParams)
		penalty := ComputeQueuePriorityPenalty(s.score, effParams)
		r := result{FeeAdjusted: fee, QueueBoost: boost, QueuePenalty: penalty}
		results1[i] = r
	}

	for i, s := range states {
		fee := ComputeBoundedFeeAdjustmentNaet(baseFee, s.score, s.found, effParams)
		boost := ComputeQueuePriorityBoost(s.score, effParams)
		penalty := ComputeQueuePriorityPenalty(s.score, effParams)
		r := result{FeeAdjusted: fee, QueueBoost: boost, QueuePenalty: penalty}
		results2[i] = r
	}

	for i := range results1 {
		require.Equal(t, results1[i], results2[i], "run %d must match across runs", i)
	}
	require.Equal(t, results1[3].FeeAdjusted, results1[3].FeeAdjusted)
	require.Greater(t, results1[0].FeeAdjusted, baseFee)
	require.Less(t, results1[3].FeeAdjusted, baseFee)
	require.Greater(t, results1[3].FeeAdjusted, baseFee/10)

	bz, _ := json.Marshal(results1)
	h := sha256.Sum256(bz)
	fmt.Printf("golden_deterministic_hash: %x\n", h)
}

func TestPerEpochUpdateCaps(t *testing.T) {
	params := DefaultReputationEffectParams()

	capped := ApplyPerEpochScoreCap(5000, 6000, params.PerEpochScoreCap)
	require.LessOrEqual(t, capped-5000, params.PerEpochScoreCap)

	cappedDown := ApplyPerEpochScoreCap(5000, 4000, params.PerEpochScoreCap)
	require.LessOrEqual(t, uint32(5000)-cappedDown, params.PerEpochScoreCap)

	cappedConf := ApplyPerEpochConfidenceCap(5000, 7000, params.PerEpochConfidenceCap)
	require.LessOrEqual(t, cappedConf-5000, params.PerEpochConfidenceCap)
}

func TestNonGatingEnforcementOnAllRightOps(t *testing.T) {
	require.NoError(t, ValidateNonGatingEnforcement())

	for _, op := range []string{"contract_deployment", "contract_execution", "basic_transaction", "pool_staking"} {
		require.True(t, CanLowReputationPerformOperation(op), "must allow: %s", op)
	}

	lowRep := NewIdentityReputation("AEtest")
	lowRep.Score = 50
	require.True(t, LowReputationCanPerformBasicTransfer(lowRep))
	require.True(t, LowReputationCanPerformContractDeployment(lowRep))
	require.True(t, LowReputationCanPerformContractExecution(lowRep))
	require.True(t, LowReputationCanPerformPoolStaking(lowRep, MinStakeAmountForReputation))
}

func TestValidatorScoreQuerySeparateFromIdentity(t *testing.T) {
	vs := NewValidatorScore("val-1")
	vs.UptimeScore = 800
	vs.CommissionBehavior = 500
	vs.TotalScore = ComputeValidatorTotalScore(vs)

	id := NewIdentityReputation(aetraaddress.FormatAccAddress(bytes20(0x44)))
	id.Score = 3000
	id.Confidence = 2000

	require.NotEqual(t, id.Score, vs.TotalScore)
}

func TestMigrationWithReceiptDeterministic(t *testing.T) {
	old := newTestReputationState(t)
	old.Accounts = []ReputationRecord{
		ApplyComputedScore(ReputationRecord{Account: addr(1), AgeScore: 50, TxSuccessScore: 10, SpamPenalty: 2}),
		ApplyComputedScore(ReputationRecord{Account: addr(2), AgeScore: 30, StakingScore: 5}),
	}
	old.Validators = []ReputationRecord{
		ApplyComputedScore(ReputationRecord{Account: addr(3), AgeScore: 60, TxSuccessScore: 20}),
	}
	old.Reporters = []ReputationRecord{
		ApplyComputedScore(ReputationRecord{Account: addr(4), TxSuccessScore: 15}),
	}
	old.StakeRecords = []StakeReputationRecord{
		{
			Account:			addr(5),
			AccountUser:			aeAddr(0x05),
			StakeWeightedSeconds:		86400,
			ClaimedStakeWeightedSeconds:	86400,
			ClaimedStakeReputation:		24,
			LastUpdatedUnix:		100,
		},
	}

	c1, r1 := MigrateFromReputationStateWithReceipt(old, 1000)
	c2, r2 := MigrateFromReputationStateWithReceipt(old, 1000)

	require.Equal(t, len(c1.Identities), len(c2.Identities))
	require.Equal(t, len(c1.ValidatorScores), len(c2.ValidatorScores))
	require.Equal(t, r1.DeterministicHash, r2.DeterministicHash)
	for i := range c1.Identities {
		require.Equal(t, c1.Identities[i].Score, c2.Identities[i].Score)
	}
	require.NoError(t, r1.Validate())
	require.NoError(t, r2.Validate())
}

func TestMigrationDuplicateStakeTimeRejected(t *testing.T) {
	old := newTestReputationState(t)
	old.StakeRecords = []StakeReputationRecord{
		{Account: addr(1), AccountUser: aeAddr(0x01), StakeWeightedSeconds: 3600, ClaimedStakeWeightedSeconds: 3600, ClaimedStakeReputation: 1, LastUpdatedUnix: 100},
		{Account: addr(1), AccountUser: aeAddr(0x01), StakeWeightedSeconds: 7200, ClaimedStakeWeightedSeconds: 7200, ClaimedStakeReputation: 2, LastUpdatedUnix: 200},
	}

	_, r := MigrateFromReputationStateWithReceipt(old, 300)
	dupEntries := 0
	for _, e := range r.Entries {
		if e.MergeReason == "duplicate_stake_time_rejected" {
			dupEntries++
		}
	}
	require.Greater(t, dupEntries, 0)
}

func TestBoundedScoreConversion(t *testing.T) {
	old := newTestReputationState(t)
	old.Accounts = []ReputationRecord{
		ApplyComputedScore(ReputationRecord{Account: addr(1), AgeScore: 100, TxSuccessScore: 99}),
	}

	state, receipt := MigrateFromReputationStateWithReceipt(old, 0)
	require.NoError(t, state.Validate())
	require.NoError(t, receipt.Validate())

	for _, id := range state.Identities {
		require.LessOrEqual(t, id.Score, IdentityScoreMax)
		require.LessOrEqual(t, id.Confidence, ConfidenceMax)
	}
}

func TestNoContractReputationStateKey(t *testing.T) {
	err := ValidateNoContractReputationState("contract-0xABC")
	require.Error(t, err)
	require.Contains(t, err.Error(), "do not have reputation state")
}

func TestIdentityScoreMinMaxConstants(t *testing.T) {
	require.Equal(t, uint32(0), IdentityScoreMin)
	require.Equal(t, uint32(10000), IdentityScoreMax)
	require.Equal(t, uint32(0), ConfidenceMin)
	require.Equal(t, uint32(10000), ConfidenceMax)
}

func TestConsolidatedStateWithEffectParamsRoundTrip(t *testing.T) {
	state := NewConsolidatedReputationState(DefaultReputationParams())
	require.NoError(t, state.EffectParams.Validate())

	bz, err := json.Marshal(state)
	require.NoError(t, err)

	var restored ConsolidatedReputationState
	require.NoError(t, json.Unmarshal(bz, &restored))
	require.Equal(t, state.EffectParams.MaxFeePremiumBps, restored.EffectParams.MaxFeePremiumBps)
	require.Equal(t, state.EffectParams.MaxFeeDiscountBps, restored.EffectParams.MaxFeeDiscountBps)
	require.Equal(t, state.EffectParams.MaxQueuePriorityBoostBps, restored.EffectParams.MaxQueuePriorityBoostBps)
	require.Equal(t, state.EffectParams.MaxValidatorAllocBoostBps, restored.EffectParams.MaxValidatorAllocBoostBps)
	require.Equal(t, state.EffectParams.MaxServiceTrustRoutingBps, restored.EffectParams.MaxServiceTrustRoutingBps)
}

func TestSignalProviderValidation(t *testing.T) {
	signal := ReputationSignal{ProviderType: SignalProviderUser, ProviderID: "test", SignalType: SignalTypeTxSuccess, Height: 1, Amount: 1}
	require.NoError(t, signal.Validate())

	badSignal := ReputationSignal{ProviderType: SignalProviderValidator, ProviderID: "test", SignalType: SignalTypeTxSuccess, Height: 1, Amount: 1}
	require.Error(t, badSignal.Validate())
}

func TestApplyValidatorSignalJailedNoBonus(t *testing.T) {
	vs := NewValidatorScore("val-test")
	vs.UptimeScore = 800
	vs.CommissionBehavior = 500
	vs.TotalScore = ComputeValidatorTotalScore(vs)
	originalScore := vs.TotalScore

	effParams := DefaultReputationEffectParams()
	jailSignal := ReputationSignal{ProviderType: SignalProviderValidator, ProviderID: "val-test", SignalType: SignalTypeValidatorMissed, Height: 100, Amount: 500}

	updated, err := ApplyValidatorSignal(vs, jailSignal, effParams)
	require.NoError(t, err)
	require.Less(t, updated.TotalScore, originalScore)

	updated.IsJailed = true
	updated.TotalScore = ComputeValidatorTotalScore(updated)
	require.Equal(t, uint32(0), updated.TotalScore)

	uptimeSignal := ReputationSignal{ProviderType: SignalProviderValidator, ProviderID: "val-test", SignalType: SignalTypeValidatorUptime, Height: 101, Amount: 100}
	updatedAfterJail, err := ApplyValidatorSignal(updated, uptimeSignal, effParams)
	require.NoError(t, err)
	require.Equal(t, uint32(0), updatedAfterJail.TotalScore, "jailed validator still gets 0")
}

func TestApplyServiceSignalRoutingOnly(t *testing.T) {
	sts := NewServiceTrustScore("svc-test")

	avail := ReputationSignal{ProviderType: SignalProviderService, ProviderID: "svc-test", SignalType: SignalTypeServiceAvailable, Height: 100, Amount: 50}
	updated, err := ApplyServiceSignal(sts, avail)
	require.NoError(t, err)
	require.Greater(t, updated.Trust, uint32(50))
	require.Greater(t, updated.Reliability, uint32(50))

	require.True(t, ServiceTrustCannotMoveFunds(updated))
	require.True(t, ServiceTrustCannotBypassFees(updated))
}

func TestEffectParamsValidation(t *testing.T) {
	p := DefaultReputationEffectParams()
	require.NoError(t, p.Validate())

	bad := p
	bad.MaxFeePremiumBps = 0
	require.Error(t, bad.Validate())

	bad2 := p
	bad2.MaxFeeDiscountBps = 20000
	require.Error(t, bad2.Validate())
}

func TestReputationEffectParamsUnsignedIntValidation(t *testing.T) {
	p := DefaultReputationEffectParams()

	require.NotZero(t, p.PerEpochScoreCap)
	require.NotZero(t, p.PerEpochConfidenceCap)
	require.NotZero(t, p.DecayRatePerEpoch)
	require.NotZero(t, p.ConfidenceGainRate)
	require.NotZero(t, p.ConfidenceLossRate)
	require.NotZero(t, p.MinStakeAmountForPool)
}
