package types

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
)

const (
	IdentityScoreMin	= uint32(0)
	IdentityScoreMax	= uint32(10000)
	IdentityScoreDefault	= uint32(100)

	ConfidenceMin		= uint32(0)
	ConfidenceMax		= uint32(10000)
	ConfidenceDefault	= uint32(100)

	DecayEpochDefaultInterval	= uint64(10)
	DecayRatePerEpochDefault	= uint8(1)
	MinStakeAmountForReputation	= uint64(10_000_000)

	SignalWeightStakeTime	= uint32(40)
	SignalWeightTxSuccess	= uint32(20)
	SignalWeightTxVolume	= uint32(15)
	SignalWeightContractOk	= uint32(10)
	SignalWeightDomain	= uint32(5)
	SignalWeightUptime	= uint32(10)

	NonGatingSoftWeightMax	= uint8(20)
)

type IdentityReputation struct {
	Account			string			`json:"account"`
	Score			uint32			`json:"score"`
	Confidence		uint32			`json:"confidence"`
	LastUpdateHeight	uint64			`json:"last_update_height"`
	LastUpdateTime		int64			`json:"last_update_time"`
	SignalCounters		IdentitySignalCounters	`json:"signal_counters"`
	StakeTimeAccumulator	uint64			`json:"stake_time_accumulator"`
	ClaimedStakeTimeSeconds	uint64			`json:"claimed_stake_time_seconds"`
	DecayEpoch		uint64			`json:"decay_epoch"`
}

type IdentitySignalCounters struct {
	SuccessfulTxs		uint64	`json:"successful_txs"`
	FailedTxs		uint64	`json:"failed_txs"`
	ContractInteractions	uint64	`json:"contract_interactions"`
	ContractFailures	uint64	`json:"contract_failures"`
	SpamCount		uint64	`json:"spam_count"`
	StakeTimeSeconds	uint64	`json:"stake_time_seconds"`
	UptimeBlocks		uint64	`json:"uptime_blocks"`
	DomainRegistrations	uint64	`json:"domain_registrations"`
	SlashEvents		uint64	`json:"slash_events"`
	RecoveryEvents		uint64	`json:"recovery_events"`
}

func NewIdentityReputation(account string) *IdentityReputation {
	return &IdentityReputation{
		Account:			account,
		Score:				IdentityScoreDefault,
		Confidence:			ConfidenceDefault,
		LastUpdateHeight:		0,
		LastUpdateTime:			0,
		SignalCounters:			IdentitySignalCounters{},
		StakeTimeAccumulator:		0,
		ClaimedStakeTimeSeconds:	0,
		DecayEpoch:			0,
	}
}

func ComputeIdentityScore(rep *IdentityReputation) uint32 {
	stakeContribution := computeStakeContribution(rep)
	txSuccessContribution := computeTxSuccessContribution(rep)
	contractContribution := computeContractContribution(rep)
	domainContribution := computeDomainContribution(rep)
	uptimeContribution := computeUptimeContribution(rep)

	positive := stakeContribution + txSuccessContribution + contractContribution + domainContribution + uptimeContribution

	spamPenalty := uint32(rep.SignalCounters.SpamCount) * 50
	failedTxPenalty := uint32(rep.SignalCounters.FailedTxs) * 10
	slashPenalty := uint32(rep.SignalCounters.SlashEvents) * 500
	recoveryPenalty := uint32(rep.SignalCounters.RecoveryEvents) * 100

	negative := spamPenalty + failedTxPenalty + slashPenalty + recoveryPenalty

	if negative >= positive {
		return IdentityScoreMin
	}

	score := positive - negative
	if score > IdentityScoreMax {
		return IdentityScoreMax
	}
	if score < IdentityScoreMin {
		return IdentityScoreMin
	}
	return score
}

func computeStakeContribution(rep *IdentityReputation) uint32 {
	timeWeight := uint32(1)
	if rep.StakeTimeAccumulator > 365*24*3600 {
		timeWeight = 3
	} else if rep.StakeTimeAccumulator > 30*24*3600 {
		timeWeight = 2
	}

	stakeSeconds := rep.SignalCounters.StakeTimeSeconds
	if stakeSeconds == 0 {
		return uint32(SignalWeightStakeTime)
	}

	stakeContribution := uint32(stakeSeconds/3600) * timeWeight
	maxContribution := uint32(SignalWeightStakeTime) * 10

	if stakeContribution > maxContribution {
		stakeContribution = maxContribution
	}
	return stakeContribution
}

func computeTxSuccessContribution(rep *IdentityReputation) uint32 {
	successful := rep.SignalCounters.SuccessfulTxs
	if successful == 0 {
		return 0
	}
	contribution := uint32(successful) * 2
	maxContribution := uint32(SignalWeightTxSuccess) * 10
	if contribution > maxContribution {
		contribution = maxContribution
	}
	return contribution
}

func computeContractContribution(rep *IdentityReputation) uint32 {
	positive := rep.SignalCounters.ContractInteractions
	negative := rep.SignalCounters.ContractFailures
	if positive == 0 {
		return 0
	}
	var net uint32
	if negative > positive {
		net = 0
	} else {
		net = uint32(positive-negative) * 3
	}
	maxContribution := uint32(SignalWeightContractOk) * 10
	if net > maxContribution {
		net = maxContribution
	}
	return net
}

func computeDomainContribution(rep *IdentityReputation) uint32 {
	count := rep.SignalCounters.DomainRegistrations
	contribution := uint32(count) * 5
	maxContribution := uint32(SignalWeightDomain) * 10
	if contribution > maxContribution {
		contribution = maxContribution
	}
	return contribution
}

func computeUptimeContribution(rep *IdentityReputation) uint32 {
	blocks := rep.SignalCounters.UptimeBlocks
	if blocks == 0 {
		return 0
	}
	contribution := uint32(blocks/100) * 2
	maxContribution := uint32(SignalWeightUptime) * 10
	if contribution > maxContribution {
		contribution = maxContribution
	}
	return contribution
}

func ComputeConfidence(rep *IdentityReputation) uint32 {
	base := ConfidenceDefault

	ageBonus := uint32(0)
	if rep.LastUpdateHeight > 0 {
		ageBlocks := rep.LastUpdateHeight
		if ageBlocks > 1000000 {
			ageBonus = 3000
		} else if ageBlocks > 100000 {
			ageBonus = 1500
		} else if ageBlocks > 10000 {
			ageBonus = 500
		} else if ageBlocks > 1000 {
			ageBonus = 100
		}
	}

	activityBonus := uint32(0)
	totalTx := rep.SignalCounters.SuccessfulTxs + rep.SignalCounters.FailedTxs
	if totalTx > 10000 {
		activityBonus = 2000
	} else if totalTx > 1000 {
		activityBonus = 500
	} else if totalTx > 100 {
		activityBonus = 100
	}

	stakeBonus := uint32(0)
	if rep.StakeTimeAccumulator > 365*24*3600 {
		stakeBonus = 3000
	} else if rep.StakeTimeAccumulator > 30*24*3600 {
		stakeBonus = 1000
	} else if rep.StakeTimeAccumulator > 0 {
		stakeBonus = 200
	}

	diversityBonus := uint32(0)
	signalTypes := 0
	if rep.SignalCounters.SuccessfulTxs > 0 {
		signalTypes++
	}
	if rep.SignalCounters.ContractInteractions > 0 {
		signalTypes++
	}
	if rep.SignalCounters.DomainRegistrations > 0 {
		signalTypes++
	}
	if rep.SignalCounters.StakeTimeSeconds > 0 {
		signalTypes++
	}
	if rep.SignalCounters.UptimeBlocks > 0 {
		signalTypes++
	}
	diversityBonus = uint32(signalTypes) * 200

	confidence := base + ageBonus + activityBonus + stakeBonus + diversityBonus
	if confidence > ConfidenceMax {
		confidence = ConfidenceMax
	}
	return confidence
}

func ApplyDecay(rep *IdentityReputation, currentEpoch uint64, params DecayParams) *IdentityReputation {
	if rep.DecayEpoch == 0 {
		rep.DecayEpoch = currentEpoch
		return rep
	}

	epochsSinceUpdate := currentEpoch - rep.DecayEpoch
	if epochsSinceUpdate < params.InactiveAfterEpochs {
		return rep
	}

	decayEpochs := epochsSinceUpdate - params.InactiveAfterEpochs

	stakeSlowdownFactor := uint32(1)
	if rep.StakeTimeAccumulator > 365*24*3600 {
		stakeSlowdownFactor = 4
	} else if rep.StakeTimeAccumulator > 30*24*3600 {
		stakeSlowdownFactor = 2
	}

	effectiveDecay := uint32(decayEpochs) * uint32(params.DecayRatePerEpoch)
	if stakeSlowdownFactor > 1 {
		effectiveDecay = effectiveDecay / stakeSlowdownFactor
	}

	if effectiveDecay > rep.Score {
		rep.Score = IdentityScoreDefault
	} else {
		rep.Score -= effectiveDecay
	}

	confidenceDecay := effectiveDecay / 2
	if confidenceDecay > rep.Confidence {
		rep.Confidence = ConfidenceDefault
	} else {
		rep.Confidence -= confidenceDecay
	}

	rep.DecayEpoch = currentEpoch
	return rep
}

func ComputeStakeReputationDelta(stakeAmount uint64, stakeTimeSeconds uint64, consistencyFactor uint32, confidenceFactor uint32) uint32 {
	if stakeAmount == 0 || stakeTimeSeconds == 0 {
		return 0
	}

	timeWeight := uint32(1)
	if stakeTimeSeconds > 365*24*3600 {
		timeWeight = 3
	} else if stakeTimeSeconds > 30*24*3600 {
		timeWeight = 2
	}

	amountScaled := uint32(stakeAmount / 1_000_000)
	if amountScaled > 10000 {
		amountScaled = 10000
	}

	delta := amountScaled * timeWeight
	if consistencyFactor > 0 {
		delta = delta * consistencyFactor / 100
	}
	if confidenceFactor > 0 {
		delta = delta * confidenceFactor / 100
	}

	maxDelta := uint32(1000)
	if delta > maxDelta {
		delta = maxDelta
	}
	return delta
}

func (r *IdentityReputation) RecordSuccessfulTx(height uint64) {
	r.SignalCounters.SuccessfulTxs++
	r.LastUpdateHeight = height
}

func (r *IdentityReputation) RecordFailedTx(height uint64) {
	r.SignalCounters.FailedTxs++
	r.LastUpdateHeight = height
}

func (r *IdentityReputation) RecordContractInteraction(height uint64) {
	r.SignalCounters.ContractInteractions++
	r.LastUpdateHeight = height
}

func (r *IdentityReputation) RecordContractFailure(height uint64) {
	r.SignalCounters.ContractFailures++
	r.LastUpdateHeight = height
}

func (r *IdentityReputation) RecordSpam(height uint64) {
	r.SignalCounters.SpamCount++
	r.LastUpdateHeight = height
}

func (r *IdentityReputation) RecordStakeTime(seconds uint64, height uint64) {
	if seconds == 0 {
		return
	}
	newSeconds := seconds
	if r.ClaimedStakeTimeSeconds > 0 && seconds <= r.ClaimedStakeTimeSeconds {
		return
	}
	if r.ClaimedStakeTimeSeconds > 0 && seconds > r.ClaimedStakeTimeSeconds {
		newSeconds = seconds - r.ClaimedStakeTimeSeconds
	}
	r.SignalCounters.StakeTimeSeconds += newSeconds
	r.StakeTimeAccumulator += newSeconds
	r.ClaimedStakeTimeSeconds = seconds
	r.LastUpdateHeight = height
}

func (r *IdentityReputation) RecordJailed(height uint64) {
	r.SignalCounters.SlashEvents++
	r.LastUpdateHeight = height
}

func (r *IdentityReputation) RecordUnfrozen(height uint64) {
	r.SignalCounters.RecoveryEvents++
	r.LastUpdateHeight = height
}

func (r *IdentityReputation) RecordUptime(blocks uint64, height uint64) {
	r.SignalCounters.UptimeBlocks += blocks
	r.LastUpdateHeight = height
}

func (r *IdentityReputation) RecordDomainRegistration(height uint64) {
	r.SignalCounters.DomainRegistrations++
	r.LastUpdateHeight = height
}

func (r *IdentityReputation) RecordSlashEvent(height uint64) {
	r.SignalCounters.SlashEvents++
	r.LastUpdateHeight = height
}

func (r *IdentityReputation) RecordRecoveryEvent(height uint64) {
	r.SignalCounters.RecoveryEvents++
	r.LastUpdateHeight = height
}

type NonGatingRule struct {
	Operation	string
	SoftWeightMax	uint8
	Gated		bool
}

var NonGatingRules = []NonGatingRule{
	{Operation: "contract_deployment", SoftWeightMax: 10, Gated: false},
	{Operation: "contract_execution", SoftWeightMax: 5, Gated: false},
	{Operation: "basic_transaction", SoftWeightMax: 5, Gated: false},
	{Operation: "pool_staking", SoftWeightMax: 15, Gated: false},
	{Operation: "tx_priority", SoftWeightMax: 20, Gated: true},
	{Operation: "async_queue_ordering", SoftWeightMax: 20, Gated: true},
	{Operation: "resource_scheduling", SoftWeightMax: 15, Gated: true},
	{Operation: "fee_discount", SoftWeightMax: 10, Gated: true},
	{Operation: "liquid_staking_yield", SoftWeightMax: 20, Gated: true},
}

func IsOperationGated(operation string) bool {
	for _, rule := range NonGatingRules {
		if rule.Operation == operation {
			return rule.Gated
		}
	}
	return false
}

func GetSoftWeight(operation string) uint8 {
	for _, rule := range NonGatingRules {
		if rule.Operation == operation {
			return rule.SoftWeightMax
		}
	}
	return 0
}

func CanLowReputationPerformOperation(operation string) bool {
	for _, rule := range NonGatingRules {
		if rule.Operation == operation {
			return !rule.Gated
		}
	}
	return true
}

func ValidateNonGatingEnforcement() error {
	nonGated := []string{"contract_deployment", "contract_execution", "basic_transaction", "pool_staking"}
	for _, op := range nonGated {
		if IsOperationGated(op) {
			return fmt.Errorf("IdentityReputation: operation %q must be non-gating", op)
		}
	}
	return nil
}

type ValidatorScore struct {
	ValidatorAddress	string	`json:"validator_address"`
	UptimeScore		uint32	`json:"uptime_score"`
	MissedBlocksPenalty	uint32	`json:"missed_blocks_penalty"`
	SlashingPenalty		uint32	`json:"slashing_penalty"`
	CommissionBehavior	uint32	`json:"commission_behavior"`
	GovernanceParticipation	uint32	`json:"governance_participation"`
	PoolAllocationScore	uint32	`json:"pool_allocation_score"`
	TotalScore		uint32	`json:"total_score"`
	IsJailed		bool	`json:"is_jailed"`
	IsSlashed		bool	`json:"is_slashed"`
	LastUpdateHeight	uint64	`json:"last_update_height"`
}

func NewValidatorScore(validatorAddr string) *ValidatorScore {
	return &ValidatorScore{
		ValidatorAddress:		validatorAddr,
		UptimeScore:			100,
		MissedBlocksPenalty:		0,
		SlashingPenalty:		0,
		CommissionBehavior:		100,
		GovernanceParticipation:	0,
		PoolAllocationScore:		0,
		TotalScore:			100,
		LastUpdateHeight:		0,
	}
}

func ComputeValidatorTotalScore(vs *ValidatorScore) uint32 {
	if vs.IsJailed || vs.IsSlashed {
		return 0
	}
	positive := vs.UptimeScore + vs.CommissionBehavior + vs.GovernanceParticipation + vs.PoolAllocationScore
	negative := vs.MissedBlocksPenalty + vs.SlashingPenalty
	if negative >= positive {
		return 0
	}
	return positive - negative
}

type ServiceTrustScore struct {
	ServiceAddress		string	`json:"service_address"`
	Trust			uint32	`json:"trust"`
	Reliability		uint32	`json:"reliability"`
	LastUpdateHeight	uint64	`json:"last_update_height"`
}

func NewServiceTrustScore(address string) *ServiceTrustScore {
	return &ServiceTrustScore{
		ServiceAddress:		address,
		Trust:			50,
		Reliability:		50,
		LastUpdateHeight:	0,
	}
}

type ReputationLevel string

const (
	ReputationLevelRestricted	ReputationLevel	= "restricted"
	ReputationLevelNew		ReputationLevel	= "new"
	ReputationLevelNormal		ReputationLevel	= "normal"
	ReputationLevelTrusted		ReputationLevel	= "trusted"
	ReputationLevelElite		ReputationLevel	= "elite"
)

func ClassifyReputationLevel(score uint32) ReputationLevel {
	switch {
	case score < 200:
		return ReputationLevelRestricted
	case score < 500:
		return ReputationLevelNew
	case score < 2000:
		return ReputationLevelNormal
	case score < 5000:
		return ReputationLevelTrusted
	default:
		return ReputationLevelElite
	}
}

func (l ReputationLevel) String() string {
	return string(l)
}

type IdentityProgressiveLimits struct {
	MaxTxsPerBlock	uint32
	MaxTxGas	uint64
	MaxQueueMsgs	uint32
}

func GetIdentityProgressiveLimits(level ReputationLevel) IdentityProgressiveLimits {
	switch level {
	case ReputationLevelRestricted:
		return IdentityProgressiveLimits{MaxTxsPerBlock: 1, MaxTxGas: 100_000, MaxQueueMsgs: 1}
	case ReputationLevelNew:
		return IdentityProgressiveLimits{MaxTxsPerBlock: 5, MaxTxGas: 250_000, MaxQueueMsgs: 4}
	case ReputationLevelNormal:
		return IdentityProgressiveLimits{MaxTxsPerBlock: 25, MaxTxGas: 1_000_000, MaxQueueMsgs: 16}
	case ReputationLevelTrusted:
		return IdentityProgressiveLimits{MaxTxsPerBlock: 100, MaxTxGas: 2_000_000, MaxQueueMsgs: 64}
	case ReputationLevelElite:
		return IdentityProgressiveLimits{MaxTxsPerBlock: 250, MaxTxGas: 5_000_000, MaxQueueMsgs: 128}
	default:
		return IdentityProgressiveLimits{MaxTxsPerBlock: 5, MaxTxGas: 250_000, MaxQueueMsgs: 4}
	}
}

type ReputationClaim struct {
	Account			string			`json:"account"`
	Score			uint32			`json:"score"`
	Confidence		uint32			`json:"confidence"`
	StakeTimeAccumulator	uint64			`json:"stake_time_accumulator"`
	ClaimedStakeTimeSeconds	uint64			`json:"claimed_stake_time_seconds"`
	DecayEpoch		uint64			`json:"decay_epoch"`
	SignalCounters		IdentitySignalCounters	`json:"signal_counters"`
	ClaimHeight		uint64			`json:"claim_height"`
	ClaimHash		string			`json:"claim_hash"`
}

func (r *IdentityReputation) ExportClaim(height uint64) ReputationClaim {
	claim := ReputationClaim{
		Account:			r.Account,
		Score:				r.Score,
		Confidence:			r.Confidence,
		StakeTimeAccumulator:		r.StakeTimeAccumulator,
		ClaimedStakeTimeSeconds:	r.ClaimedStakeTimeSeconds,
		DecayEpoch:			r.DecayEpoch,
		SignalCounters:			r.SignalCounters,
		ClaimHeight:			height,
	}
	claim.ClaimHash = computeReputationClaimHash(claim)
	return claim
}

func ImportReputationFromClaim(claim ReputationClaim) (*IdentityReputation, error) {
	expectedHash := computeReputationClaimHash(claim)
	if claim.ClaimHash != expectedHash {
		return nil, fmt.Errorf("identity reputation: claim hash mismatch")
	}
	return &IdentityReputation{
		Account:			claim.Account,
		Score:				claim.Score,
		Confidence:			claim.Confidence,
		StakeTimeAccumulator:		claim.StakeTimeAccumulator,
		ClaimedStakeTimeSeconds:	claim.ClaimedStakeTimeSeconds,
		DecayEpoch:			claim.DecayEpoch,
		SignalCounters:			claim.SignalCounters,
		LastUpdateHeight:		claim.ClaimHeight,
	}, nil
}

func computeReputationClaimHash(claim ReputationClaim) string {
	type claimForHash struct {
		Account			string	`json:"account"`
		Score			uint32	`json:"score"`
		Confidence		uint32	`json:"confidence"`
		StakeTimeAccumulator	uint64	`json:"stake_time_accumulator"`
		ClaimedStakeTimeSeconds	uint64	`json:"claimed_stake_time_seconds"`
		DecayEpoch		uint64	`json:"decay_epoch"`
		SuccessfulTxs		uint64	`json:"successful_txs"`
		FailedTxs		uint64	`json:"failed_txs"`
		ContractInteractions	uint64	`json:"contract_interactions"`
		ContractFailures	uint64	`json:"contract_failures"`
		SpamCount		uint64	`json:"spam_count"`
		StakeTimeSeconds	uint64	`json:"stake_time_seconds"`
		UptimeBlocks		uint64	`json:"uptime_blocks"`
		DomainRegistrations	uint64	`json:"domain_registrations"`
		SlashEvents		uint64	`json:"slash_events"`
		RecoveryEvents		uint64	`json:"recovery_events"`
		ClaimHeight		uint64	`json:"claim_height"`
	}
	data, _ := json.Marshal(claimForHash{
		Account:			claim.Account,
		Score:				claim.Score,
		Confidence:			claim.Confidence,
		StakeTimeAccumulator:		claim.StakeTimeAccumulator,
		ClaimedStakeTimeSeconds:	claim.ClaimedStakeTimeSeconds,
		DecayEpoch:			claim.DecayEpoch,
		SuccessfulTxs:			claim.SignalCounters.SuccessfulTxs,
		FailedTxs:			claim.SignalCounters.FailedTxs,
		ContractInteractions:		claim.SignalCounters.ContractInteractions,
		ContractFailures:		claim.SignalCounters.ContractFailures,
		SpamCount:			claim.SignalCounters.SpamCount,
		StakeTimeSeconds:		claim.SignalCounters.StakeTimeSeconds,
		UptimeBlocks:			claim.SignalCounters.UptimeBlocks,
		DomainRegistrations:		claim.SignalCounters.DomainRegistrations,
		SlashEvents:			claim.SignalCounters.SlashEvents,
		RecoveryEvents:			claim.SignalCounters.RecoveryEvents,
		ClaimHeight:			claim.ClaimHeight,
	})
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h)
}

type QueryIdentityReputationRequest struct {
	Account string `json:"account"`
}

type QueryIdentityReputationResponse struct {
	Reputation		IdentityReputation		`json:"reputation"`
	Level			ReputationLevel			`json:"level"`
	ProgressiveLimits	IdentityProgressiveLimits	`json:"progressive_limits"`
	ValidatorScore		*ValidatorScore			`json:"validator_score,omitempty"`
	ServiceTrust		*ServiceTrustScore		`json:"service_trust,omitempty"`
}

func QueryIdentityReputation(rep *IdentityReputation, vs *ValidatorScore, sts *ServiceTrustScore) QueryIdentityReputationResponse {
	level := ClassifyReputationLevel(rep.Score)
	return QueryIdentityReputationResponse{
		Reputation:		*rep,
		Level:			level,
		ProgressiveLimits:	GetIdentityProgressiveLimits(level),
		ValidatorScore:		vs,
		ServiceTrust:		sts,
	}
}

func CanStakeMinimum(rep *IdentityReputation, stakeAmount uint64) bool {
	return stakeAmount >= MinStakeAmountForReputation
}

func ValidateIdentityReputation(rep *IdentityReputation) error {
	if rep.Account == "" {
		return fmt.Errorf("identity reputation: account must not be empty")
	}
	if rep.Score > IdentityScoreMax {
		return fmt.Errorf("identity reputation: score %d exceeds maximum %d", rep.Score, IdentityScoreMax)
	}
	if rep.Confidence > ConfidenceMax {
		return fmt.Errorf("identity reputation: confidence %d exceeds maximum %d", rep.Confidence, ConfidenceMax)
	}
	return nil
}

func ValidateNoContractReputation(contractAddr string) error {
	return fmt.Errorf("identity reputation: contracts do not have persistent reputation records; %s is not an identity", contractAddr)
}

func ValidateReputationAccountAddress(account string) error {
	if account == "" {
		return fmt.Errorf("identity reputation: account address must not be empty")
	}
	if !addressing.IsSystemRawAddress(account) {
		if !strings.HasPrefix(account, addressing.UserFriendlyPrefix) && !strings.HasPrefix(account, addressing.RawPrefix) {
			return fmt.Errorf("identity reputation: account address must be AE or 4: format")
		}
	}
	return nil
}
