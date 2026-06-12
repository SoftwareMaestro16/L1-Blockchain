package types

import (
	"errors"
	"fmt"
	"strings"
)

const (
	InvariantNoPrivateKeyOnChain				= "no_private_key_on_chain"
	InvariantNoSeedPhraseOnChain				= "no_seed_phrase_on_chain"
	InvariantAEAddressRoundtripStable			= "ae_address_roundtrip_stable"
	InvariantRawAddressRoundtripStable			= "raw_4_address_roundtrip_stable"
	InvariantAccountActivationIdempotency			= "account_activation_idempotency_enforced"
	InvariantAccountCannotActivateTwice			= "account_cannot_activate_twice"
	InvariantPoolStakeDoesNotCreateCoins			= "pool_stake_does_not_create_coins"
	InvariantRewardsCannotExceedAllocation			= "rewards_cannot_exceed_emissions_or_fees"
	InvariantMaxValidatorCountEnforced			= "max_validator_count_enforced"
	InvariantMinValidatorStakeEnforced			= "min_validator_stake_enforced"
	InvariantValidatorSelfStakeRatioEnforced		= "validator_self_stake_ratio_enforced"
	InvariantMinPoolDepositEnforced				= "min_pool_deposit_enforced"
	InvariantDirectUserValidatorDelegationRejected		= "direct_user_validator_delegation_rejected"
	InvariantUnbondingCannotReleaseEarly			= "unbonding_cannot_release_early"
	InvariantReputationRequiresStakeTime			= "reputation_cannot_increase_without_stake_time"
	InvariantJailedValidatorNoPositiveBonus			= "jailed_validator_no_positive_bonus"
	InvariantExportImportPreservesState			= "export_import_preserves_state"
	InvariantModuleBankAccountingConsistent			= "module_bank_accounting_consistent"
	InvariantProtocolCriticalStateNotFrozenByRent		= "protocol_critical_state_not_frozen_by_rent"
	InvariantSystemStorageReserveRunway			= "system_storage_reserve_runway"
	InvariantSystemRentTopUpBeforeUserFreeze		= "system_rent_topup_before_user_freeze"
	InvariantProtocolCriticalExecutableUnderUnderfunding	= "protocol_critical_executable_under_system_rent_underfunding"
)

type NativeAccountInvariant struct {
	ID		string
	Description	string
}

type NativeAccountInvariantInput struct {
	ExportedPayloads	[]string
	Accounts		[]Account

	AEAddressRoundtripStable	bool
	RawAddressRoundtripStable	bool
	ActivationAttempts		map[string]uint64

	PoolActiveStake		uint64
	PoolUnbonding		uint64
	ValidatorSelfStake	uint64
	LiquidBalances		uint64
	TotalSupply		uint64

	RewardsAccrued	uint64
	RewardBudget	uint64

	ActiveValidatorCount	uint64
	MaxValidatorCount	uint64
	ValidatorStakes		[]uint64
	MinValidatorStake	uint64
	ValidatorTotalStake	uint64
	MinSelfStakeBps		uint64

	PoolDeposits			[]uint64
	MinPoolDeposit			uint64
	DirectUserDelegationSeen	bool

	UnbondingEntries	[]InvariantUnbondingEntry

	StakeTimeDelta		uint64
	ReputationDelta		uint64
	JailedBonusAmount	uint64

	ExportImportStable	bool

	BankAccountingScenarios	[]InvariantAccountingScenario

	ProtocolCriticalFrozenByRent	bool
	SystemReserveRunwayBlocks	uint64
	MinSystemReserveRunwayBlocks	uint64
	SystemReserveAlertRaised	bool
	SystemTopUpOrder		[]string
	ProtocolCriticalExecutable	bool
	SystemRentUnderfunded		bool
}

type InvariantUnbondingEntry struct {
	ReleaseHeight	uint64
	CurrentHeight	uint64
	Released	bool
}

type InvariantAccountingScenario struct {
	Name			string
	BankBalance		uint64
	ModuleAccounting	uint64
	ExpectedSupplyChange	int64
	ActualSupplyChange	int64
}

type NativeAccountInvariantFailure struct {
	ID	string
	Error	string
}

func RequiredNativeAccountInvariantIDs() []string {
	return []string{
		InvariantNoPrivateKeyOnChain,
		InvariantNoSeedPhraseOnChain,
		InvariantAEAddressRoundtripStable,
		InvariantRawAddressRoundtripStable,
		InvariantAccountActivationIdempotency,
		InvariantAccountCannotActivateTwice,
		InvariantPoolStakeDoesNotCreateCoins,
		InvariantRewardsCannotExceedAllocation,
		InvariantMaxValidatorCountEnforced,
		InvariantMinValidatorStakeEnforced,
		InvariantValidatorSelfStakeRatioEnforced,
		InvariantMinPoolDepositEnforced,
		InvariantDirectUserValidatorDelegationRejected,
		InvariantUnbondingCannotReleaseEarly,
		InvariantReputationRequiresStakeTime,
		InvariantJailedValidatorNoPositiveBonus,
		InvariantExportImportPreservesState,
		InvariantModuleBankAccountingConsistent,
		InvariantProtocolCriticalStateNotFrozenByRent,
		InvariantSystemStorageReserveRunway,
		InvariantSystemRentTopUpBeforeUserFreeze,
		InvariantProtocolCriticalExecutableUnderUnderfunding,
	}
}

func DefaultNativeAccountInvariantRegistry() []NativeAccountInvariant {
	return []NativeAccountInvariant{
		{InvariantNoPrivateKeyOnChain, "private keys must never appear in account, auth, genesis, event, or export payloads"},
		{InvariantNoSeedPhraseOnChain, "seed phrases and mnemonics must never appear on-chain"},
		{InvariantAEAddressRoundtripStable, "AE user-facing address roundtrip remains stable"},
		{InvariantRawAddressRoundtripStable, "4: raw address roundtrip remains stable"},
		{InvariantAccountActivationIdempotency, "account activation is idempotency-safe"},
		{InvariantAccountCannotActivateTwice, "activated accounts cannot be activated a second time"},
		{InvariantPoolStakeDoesNotCreateCoins, "pool active stake, unbonding, validator self stake, and liquid balances conserve supply"},
		{InvariantRewardsCannotExceedAllocation, "rewards cannot exceed emissions and fee allocation"},
		{InvariantMaxValidatorCountEnforced, "MaxValidatorCount is enforced"},
		{InvariantMinValidatorStakeEnforced, "MinValidatorStake is enforced"},
		{InvariantValidatorSelfStakeRatioEnforced, "validator self-stake ratio is enforced"},
		{InvariantMinPoolDepositEnforced, "MinPoolDeposit is enforced"},
		{InvariantDirectUserValidatorDelegationRejected, "direct user validator delegation is rejected"},
		{InvariantUnbondingCannotReleaseEarly, "unbonding cannot release before maturity"},
		{InvariantReputationRequiresStakeTime, "reputation cannot increase without stake-time"},
		{InvariantJailedValidatorNoPositiveBonus, "jailed validators cannot receive positive validator bonus"},
		{InvariantExportImportPreservesState, "export/import preserves all consensus state classes"},
		{InvariantModuleBankAccountingConsistent, "module account and bank accounting remain consistent"},
		{InvariantProtocolCriticalStateNotFrozenByRent, "protocol-critical state is not frozen by storage rent"},
		{InvariantSystemStorageReserveRunway, "system storage reserve has runway or raises an alert"},
		{InvariantSystemRentTopUpBeforeUserFreeze, "system rent top-up runs before user freeze processing"},
		{InvariantProtocolCriticalExecutableUnderUnderfunding, "protocol-critical modules remain executable during system rent underfunding"},
	}
}

func ValidateNativeAccountInvariantRegistry(registry []NativeAccountInvariant) error {
	required := RequiredNativeAccountInvariantIDs()
	seen := make(map[string]struct{}, len(registry))
	for _, invariant := range registry {
		if strings.TrimSpace(invariant.ID) == "" || strings.TrimSpace(invariant.Description) == "" {
			return errors.New("native account invariant id and description are required")
		}
		if _, found := seen[invariant.ID]; found {
			return fmt.Errorf("duplicate native account invariant %s", invariant.ID)
		}
		seen[invariant.ID] = struct{}{}
	}
	for _, id := range required {
		if _, found := seen[id]; !found {
			return fmt.Errorf("native account invariant registry missing %s", id)
		}
	}
	return nil
}
