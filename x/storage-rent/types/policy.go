package types

import (
	"errors"
	"fmt"
	"math"
)

const (
	StateClassWallet		= "wallet"
	StateClassContract		= "contract"
	StateClassPoolContract		= "pool_contract"
	StateClassPoolShare		= "pool_share"
	StateClassPoolAllocation	= "pool_allocation"
	StateClassPoolRewardIndex	= "pool_reward_index"
	StateClassPoolUnbonding		= "pool_unbonding"
	StateClassDomainRecord		= "domain_record"
	StateClassStakingReputation	= "staking_reputation"
	StateClassSystemModule		= "system_module"
	StateClassValidatorRecord	= "validator_record"

	ContractStatusFrozenLimited	= "frozen_limited"

	PoolActionDeposit		= "deposit"
	PoolActionClaim			= "claim"
	PoolActionUnbond		= "unbond"
	PoolActionMaturedWithdrawal	= "matured_withdrawal"
	PoolActionGovernanceRecovery	= "governance_recovery"

	SystemRentAlertWarning		= "warning"
	SystemRentAlertCritical		= "critical"
	SystemRentAlertInvariant	= "invariant"

	SystemRentPayerFeeCollector	= "fee_collector"
	SystemRentPayerTreasury		= "treasury"
	SystemRentPayerGovernance	= "governance_configured_payer"

	RentProcessingStepSystemTopUp	= "system_top_up"
	RentProcessingStepUserFreeze	= "user_freeze"
)

type PersistentStateRecord struct {
	SubjectID		string
	Class			string
	CodeBytes		uint64
	DataBytes		uint64
	IndexBytes		uint64
	Persistent		bool
	Status			string
	ProtocolCritical	bool
	OfficialPool		bool
	RentDebt		uint64
	LastChargedHeight	uint64
}

type SubjectRentResult struct {
	Subject		PersistentStateRecord
	StorageBytes	uint64
	RentDelta	uint64
	ProtocolPaid	bool
	UserFacingFee	bool
}

type PoolRentPayer struct {
	ProtocolFeeReserve	uint64
	GovernanceReserve	uint64
	UserFacingCharge	uint64
}

type SystemRentAccounting struct {
	AvailableFunds				uint64
	ProjectedRentPerBlock			uint64
	WarningRunwayBlocks			uint64
	CriticalRunwayBlocks			uint64
	FeeCollectorBalance			uint64
	TreasuryBalance				uint64
	GovernanceConfiguredPayerBalance	uint64
	RequiredTopUp				uint64
	ProtocolCriticalExecutable		bool
}

type SystemRentResult struct {
	RunwayBlocks	uint64
	Alert		string
	TopUpAmount	uint64
	TopUpSources	[]SystemRentTopUpSource
	RemainingDebt	uint64
	FreezeForbidden	bool
	Executable	bool
}

type SystemRentTopUpSource struct {
	Source	string
	Amount	uint64
}

type RentProcessingInput struct {
	Params			StorageRentParams
	System			SystemRentAccounting
	Subjects		[]PersistentStateRecord
	CurrentUnixSeconds	uint64
	FreezeDebtThreshold	uint64
}

type RentProcessingResult struct {
	System		SystemRentResult
	Subjects	[]PersistentStateRecord
	Order		[]string
}

func PersistentStorageSize(record PersistentStateRecord) (uint64, error) {
	total := record.CodeBytes
	if total > math.MaxUint64-record.DataBytes {
		return 0, errors.New("storage rent persistent state size overflow")
	}
	return total + record.DataBytes, nil
}

func AccruePersistentStateRent(record PersistentStateRecord, params StorageRentParams, height uint64) (SubjectRentResult, error) {
	if err := params.Validate(); err != nil {
		return SubjectRentResult{}, err
	}
	if !record.Persistent || record.Status == ContractStatusDeleted || record.Status == ContractStatusArchived {
		record.LastChargedHeight = height
		return SubjectRentResult{Subject: record}, nil
	}
	if height == 0 || height < record.LastChargedHeight {
		return SubjectRentResult{}, errors.New("storage rent subject height must be monotonic")
	}
	size, err := PersistentStorageSize(record)
	if err != nil {
		return SubjectRentResult{}, err
	}
	delta, err := RentForSeconds(size, params, height-record.LastChargedHeight)
	if err != nil {
		return SubjectRentResult{}, err
	}
	record.LastChargedHeight = height
	if record.RentDebt > math.MaxUint64-delta {
		return SubjectRentResult{}, errors.New("storage rent subject debt overflow")
	}
	record.RentDebt += delta
	return SubjectRentResult{
		Subject:	record,
		StorageBytes:	size,
		RentDelta:	delta,
		ProtocolPaid:	record.ProtocolCritical || record.Class == StateClassSystemModule || record.Class == StateClassValidatorRecord,
		UserFacingFee:	record.Class == StateClassWallet,
	}, nil
}

func EffectiveWalletFee(gasFee, rentDelta, unpaidDebt uint64) (uint64, error) {
	if gasFee > math.MaxUint64-rentDelta {
		return 0, errors.New("wallet effective fee overflow")
	}
	total := gasFee + rentDelta
	if total > math.MaxUint64-unpaidDebt {
		return 0, errors.New("wallet effective fee overflow")
	}
	return total + unpaidDebt, nil
}

func PayPoolRent(payer PoolRentPayer, debt uint64) (PoolRentPayer, uint64) {
	remaining := debt
	payFrom := func(balance *uint64) {
		if remaining == 0 {
			return
		}
		if *balance >= remaining {
			*balance -= remaining
			remaining = 0
			return
		}
		remaining -= *balance
		*balance = 0
	}
	payFrom(&payer.ProtocolFeeReserve)
	payFrom(&payer.GovernanceReserve)
	return payer, remaining
}

func OfficialPoolStatusForDebt(pool PersistentStateRecord, debt uint64) string {
	if pool.ProtocolCritical || pool.Class == StateClassSystemModule || pool.Class == StateClassValidatorRecord {
		if pool.Status == "" {
			return ContractStatusActive
		}
		return pool.Status
	}
	if pool.OfficialPool && debt > 0 {
		return ContractStatusFrozenLimited
	}
	if debt > 0 {
		return ContractStatusFrozen
	}
	if pool.Status == "" {
		return ContractStatusActive
	}
	return pool.Status
}

func StatusAfterRentDebt(record PersistentStateRecord, debt uint64) string {
	if record.ProtocolCritical || record.Class == StateClassSystemModule || record.Class == StateClassValidatorRecord {
		return normalizedActiveStatus(record.Status)
	}
	if record.OfficialPool || record.Class == StateClassPoolContract {
		return OfficialPoolStatusForDebt(record, debt)
	}
	if debt > 0 {
		return ContractStatusFrozen
	}
	return normalizedActiveStatus(record.Status)
}

func FrozenLimitedPoolAllows(action string) bool {
	switch action {
	case PoolActionClaim, PoolActionUnbond, PoolActionMaturedWithdrawal, PoolActionGovernanceRecovery:
		return true
	case PoolActionDeposit:
		return false
	default:
		return false
	}
}

func CanReadPersistentState(record PersistentStateRecord) bool {
	return record.Status != ContractStatusDeleted
}

func CanQueryPersistentStateProof(record PersistentStateRecord) bool {
	return CanReadPersistentState(record)
}

func CanExecutePersistentStateAction(record PersistentStateRecord, action string) bool {
	switch record.Status {
	case "", ContractStatusActive:
		return true
	case ContractStatusFrozen:
		return false
	case ContractStatusFrozenLimited:
		return FrozenLimitedPoolAllows(action)
	default:
		return false
	}
}

func ComputeSystemRentAccounting(input SystemRentAccounting) SystemRentResult {
	if input.ProjectedRentPerBlock == 0 {
		return SystemRentResult{RunwayBlocks: math.MaxUint64, Executable: true, FreezeForbidden: true}
	}
	runway := input.AvailableFunds / input.ProjectedRentPerBlock
	result := SystemRentResult{
		RunwayBlocks:		runway,
		FreezeForbidden:	true,
		Executable:		input.ProtocolCriticalExecutable,
	}
	if runway < input.WarningRunwayBlocks {
		result.Alert = SystemRentAlertWarning
	}
	if runway < input.CriticalRunwayBlocks {
		result.Alert = SystemRentAlertCritical
		remaining := input.RequiredTopUp
		topUp := uint64(0)
		sources := []string{SystemRentPayerFeeCollector, SystemRentPayerTreasury, SystemRentPayerGovernance}
		for i, balance := range []uint64{input.FeeCollectorBalance, input.TreasuryBalance, input.GovernanceConfiguredPayerBalance} {
			source := sources[i]
			if remaining == 0 {
				break
			}
			if balance >= remaining {
				topUp += remaining
				if remaining > 0 {
					result.TopUpSources = append(result.TopUpSources, SystemRentTopUpSource{Source: source, Amount: remaining})
				}
				remaining = 0
				break
			}
			topUp += balance
			if balance > 0 {
				result.TopUpSources = append(result.TopUpSources, SystemRentTopUpSource{Source: source, Amount: balance})
			}
			remaining -= balance
		}
		result.TopUpAmount = topUp
		result.RemainingDebt = remaining
		if remaining > 0 {
			result.Alert = SystemRentAlertInvariant
		}
	}
	return result
}

func ProcessStorageRent(input RentProcessingInput) (RentProcessingResult, error) {
	if err := input.Params.Validate(); err != nil {
		return RentProcessingResult{}, err
	}
	if input.CurrentUnixSeconds == 0 {
		return RentProcessingResult{}, errors.New("storage rent current unix seconds must be positive")
	}
	system := ComputeSystemRentAccounting(input.System)
	result := RentProcessingResult{
		System:	system,
		Order:	[]string{RentProcessingStepSystemTopUp},
	}
	for _, subject := range input.Subjects {
		accrued, err := AccruePersistentStateRent(subject, input.Params, input.CurrentUnixSeconds)
		if err != nil {
			return RentProcessingResult{}, err
		}
		next := accrued.Subject
		if accrued.ProtocolPaid {
			next.Status = normalizedActiveStatus(next.Status)
		} else if next.RentDebt > input.FreezeDebtThreshold {
			next.Status = StatusAfterRentDebt(next, next.RentDebt)
		}
		if isForbiddenProtocolCriticalStatus(next) {
			return RentProcessingResult{}, errors.New("protocol-critical state cannot be frozen archived or deleted by storage rent")
		}
		result.Subjects = append(result.Subjects, next)
	}
	result.Order = append(result.Order, RentProcessingStepUserFreeze)
	return result, nil
}

func ValidatePersistentStateClass(class string) error {
	switch class {
	case StateClassWallet, StateClassContract, StateClassPoolContract, StateClassPoolShare,
		StateClassPoolAllocation, StateClassPoolRewardIndex, StateClassPoolUnbonding,
		StateClassDomainRecord, StateClassStakingReputation, StateClassSystemModule,
		StateClassValidatorRecord:
		return nil
	default:
		return fmt.Errorf("unsupported storage rent state class %q", class)
	}
}

func normalizedActiveStatus(status string) string {
	switch status {
	case "":
		return ContractStatusActive
	case ContractStatusFrozen, ContractStatusFrozenLimited, ContractStatusArchived, ContractStatusDeleted:
		return ContractStatusActive
	default:
		return status
	}
}

func isForbiddenProtocolCriticalStatus(record PersistentStateRecord) bool {
	if !record.ProtocolCritical && record.Class != StateClassSystemModule && record.Class != StateClassValidatorRecord {
		return false
	}
	switch record.Status {
	case ContractStatusFrozen, ContractStatusFrozenLimited, ContractStatusArchived, ContractStatusDeleted:
		return true
	default:
		return false
	}
}
