package types

import (
	"errors"
	"fmt"

	sdkmath "cosmossdk.io/math"
)

const (
	OperationContractDeployment	= "contract_deployment"
	OperationDomainAuctionBid	= "domain_auction_bid"

	DefaultMemoByteCost	= uint64(1)
	DefaultStorageByteCost	= uint64(10)

	RestrictedCostMultiplierBps	= uint32(40_000)
	NewCostMultiplierBps		= uint32(20_000)
	NormalCostMultiplierBps		= uint32(10_000)
	TrustedCostMultiplierBps	= uint32(8_000)
	EliteCostMultiplierBps		= uint32(7_000)

	RestrictedPriorityWeight	= uint32(10)
	NewPriorityWeight		= uint32(25)
	NormalPriorityWeight		= uint32(50)
	TrustedPriorityWeight		= uint32(80)
	ElitePriorityWeight		= uint32(100)
)

type UsagePolicy struct {
	ContractDeployMinScore		uint8
	ContractDeployDeposit		sdkmath.Int
	DomainAuctionMaxBidsByUser	uint32
	BaseMemoByteCost		uint64
	BaseStorageByteCost		uint64
}

type UsageLimits struct {
	TxRateLimit			uint32
	AsyncQueueQuota			uint32
	MaxTxGas			uint64
	MemoByteCost			uint64
	StorageByteCost			uint64
	ContractDeploysPerEpoch		uint32
	DomainAuctionBidsPerBlock	uint32
}

type PriorityKey struct {
	Weight	uint32
	TxIndex	uint32
}

func DefaultUsagePolicy() UsagePolicy {
	return UsagePolicy{
		ContractDeployMinScore:		50,
		ContractDeployDeposit:		sdkmath.NewInt(10_000_000_000),
		DomainAuctionMaxBidsByUser:	5,
		BaseMemoByteCost:		DefaultMemoByteCost,
		BaseStorageByteCost:		DefaultStorageByteCost,
	}
}

func LimitsForReputation(score uint8, policy UsagePolicy) UsageLimits {
	limits := LimitsForScore(score)
	multiplier := CostMultiplierBps(score)
	return UsageLimits{
		TxRateLimit:			limits.MaxTxsPerBlock,
		AsyncQueueQuota:		limits.MaxQueueMsgs,
		MaxTxGas:			limits.MaxTxGas,
		MemoByteCost:			applyBps(policy.BaseMemoByteCost, multiplier),
		StorageByteCost:		applyBps(policy.BaseStorageByteCost, multiplier),
		ContractDeploysPerEpoch:	contractDeployLimit(score),
		DomainAuctionBidsPerBlock:	minU32(policy.DomainAuctionMaxBidsByUser, limits.MaxTxsPerBlock),
	}
}

func ValidateTxUsage(record ReputationRecord, txsUsedThisBlock uint32, requiredFeePaid bool, baseValidationPassed bool, policy UsagePolicy) error {
	if err := ValidateReputationRecord(record); err != nil {
		return err
	}
	if !baseValidationPassed {
		return errors.New("base transaction validation must pass before reputation usage")
	}
	if !requiredFeePaid {
		return errors.New("required protocol fee must be paid")
	}
	limit := LimitsForReputation(record.Score, policy).TxRateLimit
	if txsUsedThisBlock >= limit {
		return fmt.Errorf("reputation tx rate limit %d reached", limit)
	}
	return nil
}

func ValidateAsyncQueueUsage(record ReputationRecord, queuedMessages uint32, requiredFeePaid bool, policy UsagePolicy) error {
	if err := ValidateReputationRecord(record); err != nil {
		return err
	}
	if !requiredFeePaid {
		return errors.New("required protocol fee must be paid")
	}
	limit := LimitsForReputation(record.Score, policy).AsyncQueueQuota
	if queuedMessages >= limit {
		return fmt.Errorf("reputation async queue quota %d reached", limit)
	}
	return nil
}

func ValidateAccessOperation(operation string, record ReputationRecord, deposit sdkmath.Int, policy UsagePolicy) error {
	if err := ValidateReputationRecord(record); err != nil {
		return err
	}
	threshold, requiredDeposit, err := accessRequirement(operation, policy)
	if err != nil {
		return err
	}
	if record.Score >= threshold {
		return nil
	}
	if deposit.GTE(requiredDeposit) {
		return nil
	}
	return fmt.Errorf("%s requires reputation score %d or deposit %s", operation, threshold, requiredDeposit.String())
}

func PriorityForReputation(score uint8, txIndex uint32) PriorityKey {
	return PriorityKey{Weight: PriorityWeight(score), TxIndex: txIndex}
}

func PriorityWeight(score uint8) uint32 {
	switch LevelForScore(score) {
	case LevelRestricted:
		return RestrictedPriorityWeight
	case LevelNew:
		return NewPriorityWeight
	case LevelNormal:
		return NormalPriorityWeight
	case LevelTrusted:
		return TrustedPriorityWeight
	default:
		return ElitePriorityWeight
	}
}

func CostMultiplierBps(score uint8) uint32 {
	switch LevelForScore(score) {
	case LevelRestricted:
		return RestrictedCostMultiplierBps
	case LevelNew:
		return NewCostMultiplierBps
	case LevelNormal:
		return NormalCostMultiplierBps
	case LevelTrusted:
		return TrustedCostMultiplierBps
	default:
		return EliteCostMultiplierBps
	}
}

func ApplyContractExecutionOutcome(record ReputationRecord, success bool, epoch uint64) ReputationRecord {
	if success {
		if record.ContractScore < MaxContractScore {
			record.ContractScore++
		}
	} else {
		record.FailedTxPenalty += 5
	}
	record.LastUpdatedEpoch = epoch
	return ApplyComputedScore(record)
}

func accessRequirement(operation string, policy UsagePolicy) (uint8, sdkmath.Int, error) {
	switch operation {
	case OperationContractDeployment:
		return policy.ContractDeployMinScore, policy.ContractDeployDeposit, nil
	case OperationDomainAuctionBid:
		return 0, sdkmath.ZeroInt(), nil
	default:
		return 0, sdkmath.Int{}, fmt.Errorf("unknown reputation operation %q", operation)
	}
}

func contractDeployLimit(score uint8) uint32 {
	switch LevelForScore(score) {
	case LevelRestricted, LevelNew:
		return 0
	case LevelNormal:
		return 1
	case LevelTrusted:
		return 5
	default:
		return 10
	}
}

func applyBps(value uint64, bps uint32) uint64 {
	numerator := value * uint64(bps)
	return (numerator + 9_999) / 10_000
}

func minU32(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}

func LimitsForIdentityScore(score uint32) IdentityProgressiveLimits {
	return GetIdentityProgressiveLimits(ClassifyReputationLevel(score))
}

func ValidateIdentityTxUsage(identity *IdentityReputation, txsUsedThisBlock uint32, requiredFeePaid bool, baseValidationPassed bool, policy UsagePolicy) error {
	if identity == nil {
		return errors.New("identity reputation is required")
	}
	if err := ValidateIdentityReputation(identity); err != nil {
		return err
	}
	if !baseValidationPassed {
		return errors.New("base transaction validation must pass before reputation usage")
	}
	if !requiredFeePaid {
		return errors.New("required protocol fee must be paid")
	}
	limit := LimitsForIdentityScore(identity.Score).MaxTxsPerBlock
	if txsUsedThisBlock >= limit {
		return fmt.Errorf("identity tx rate limit %d reached", limit)
	}
	return nil
}

func ValidateIdentityAsyncQueueUsage(identity *IdentityReputation, queuedMessages uint32, requiredFeePaid bool, policy UsagePolicy) error {
	if identity == nil {
		return errors.New("identity reputation is required")
	}
	if err := ValidateIdentityReputation(identity); err != nil {
		return err
	}
	if !requiredFeePaid {
		return errors.New("required protocol fee must be paid")
	}
	limit := LimitsForIdentityScore(identity.Score).MaxQueueMsgs
	if queuedMessages >= limit {
		return fmt.Errorf("identity async queue quota %d reached", limit)
	}
	return nil
}

func IdentityPriorityWeight(score uint32) uint32 {
	switch ClassifyReputationLevel(score) {
	case ReputationLevelRestricted:
		return RestrictedPriorityWeight
	case ReputationLevelNew:
		return NewPriorityWeight
	case ReputationLevelNormal:
		return NormalPriorityWeight
	case ReputationLevelTrusted:
		return TrustedPriorityWeight
	default:
		return ElitePriorityWeight
	}
}

func IdentityCostMultiplierBps(score uint32) uint32 {
	switch ClassifyReputationLevel(score) {
	case ReputationLevelRestricted:
		return RestrictedCostMultiplierBps
	case ReputationLevelNew:
		return NewCostMultiplierBps
	case ReputationLevelNormal:
		return NormalCostMultiplierBps
	case ReputationLevelTrusted:
		return TrustedCostMultiplierBps
	default:
		return EliteCostMultiplierBps
	}
}

func ApplyIdentityContractOutcome(identity *IdentityReputation, success bool, height uint64) {
	if identity == nil {
		return
	}
	if success {
		identity.RecordContractInteraction(height)
	} else {
		identity.RecordContractFailure(height)
	}
}

// IdentityScore extracts the uint32 score from an identity.
// Returns 0 if nil.
func IdentityScore(identity *IdentityReputation) uint32 {
	if identity == nil {
		return 0
	}
	return identity.Score
}
