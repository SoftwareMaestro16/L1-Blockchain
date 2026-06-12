package types

import (
	"errors"
	"fmt"
	"strings"
)

const (
	OperationBeginBlock		= "begin_block"
	OperationEndBlock		= "end_block"
	OperationDeposit		= "deposit"
	OperationRewardClaim		= "reward_claim"
	OperationReputationClaim	= "reputation_claim"
	OperationStorageRentCharge	= "storage_rent_charge"
	OperationQuery			= "query"

	DefaultMaxValidatorsPerBlock		= uint64(300)
	DefaultMaxMaintenanceItemsPerBlock	= uint64(200)
	DefaultMaxRewardClaimKeys		= uint64(16)
	DefaultMaxReputationClaimKeys		= uint64(16)
	DefaultMaxStorageRentChargeKeys		= uint64(12)
	DefaultMaxDepositKeys			= uint64(12)
	DefaultMaxPageSize			= uint64(200)
	DefaultPageSize				= uint64(50)
)

type ScalabilityLimits struct {
	MaxValidatorsPerBlock		uint64
	MaxMaintenanceItemsPerBlock	uint64
	MaxRewardClaimKeys		uint64
	MaxReputationClaimKeys		uint64
	MaxStorageRentChargeKeys	uint64
	MaxDepositKeys			uint64
	DefaultPageSize			uint64
	MaxPageSize			uint64
}

type OperationMetrics struct {
	Operation			string
	FullAccountScan			bool
	FullPoolUserScan		bool
	FullPoolShareScan		bool
	AccountsIterated		uint64
	PoolUsersIterated		uint64
	PoolSharesIterated		uint64
	AllocationItemsIterated		uint64
	ValidatorSetItemsIterated	uint64
	MaintenanceItemsProcessed	uint64
	KeysRead			uint64
	KeysWritten			uint64
	PageLimit			uint64
	PrefixKey			string
}

type StoreKeyDescriptor struct {
	Name		string
	Prefix		string
	Deterministic	bool
	Paginated	bool
}

func DefaultScalabilityLimits() ScalabilityLimits {
	return ScalabilityLimits{
		MaxValidatorsPerBlock:		DefaultMaxValidatorsPerBlock,
		MaxMaintenanceItemsPerBlock:	DefaultMaxMaintenanceItemsPerBlock,
		MaxRewardClaimKeys:		DefaultMaxRewardClaimKeys,
		MaxReputationClaimKeys:		DefaultMaxReputationClaimKeys,
		MaxStorageRentChargeKeys:	DefaultMaxStorageRentChargeKeys,
		MaxDepositKeys:			DefaultMaxDepositKeys,
		DefaultPageSize:		DefaultPageSize,
		MaxPageSize:			DefaultMaxPageSize,
	}
}

func (l ScalabilityLimits) Validate() error {
	if l.MaxValidatorsPerBlock == 0 || l.MaxValidatorsPerBlock > 300 {
		return errors.New("scalability validator block limit must be in 1..300")
	}
	if l.MaxMaintenanceItemsPerBlock == 0 {
		return errors.New("scalability maintenance queue limit must be positive")
	}
	if l.MaxRewardClaimKeys == 0 || l.MaxReputationClaimKeys == 0 || l.MaxStorageRentChargeKeys == 0 || l.MaxDepositKeys == 0 {
		return errors.New("scalability hot-path key limits must be positive")
	}
	if l.DefaultPageSize == 0 || l.MaxPageSize == 0 || l.DefaultPageSize > l.MaxPageSize {
		return errors.New("scalability pagination bounds are invalid")
	}
	return nil
}

func ValidateOperationMetrics(metrics OperationMetrics, limits ScalabilityLimits) error {
	if err := limits.Validate(); err != nil {
		return err
	}
	switch metrics.Operation {
	case OperationBeginBlock, OperationEndBlock:
		return ValidateBlockLifecycleMetrics(metrics, limits)
	case OperationDeposit:
		return ValidateDepositMetrics(metrics, limits)
	case OperationRewardClaim:
		return ValidateRewardClaimMetrics(metrics, limits)
	case OperationReputationClaim:
		return ValidateReputationClaimMetrics(metrics, limits)
	case OperationStorageRentCharge:
		return ValidateStorageRentChargeMetrics(metrics, limits)
	case OperationQuery:
		_, err := NormalizeQueryPageLimit(metrics.PageLimit, limits)
		return err
	default:
		return fmt.Errorf("unsupported scalability operation %q", metrics.Operation)
	}
}

func ValidateBlockLifecycleMetrics(metrics OperationMetrics, limits ScalabilityLimits) error {
	if err := rejectFullUserScans(metrics); err != nil {
		return err
	}
	if metrics.AccountsIterated != 0 || metrics.PoolUsersIterated != 0 || metrics.PoolSharesIterated != 0 {
		return errors.New("block lifecycle must not iterate accounts, pool users, or pool shares")
	}
	if metrics.ValidatorSetItemsIterated > limits.MaxValidatorsPerBlock {
		return fmt.Errorf("block lifecycle validator iteration exceeds %d", limits.MaxValidatorsPerBlock)
	}
	if metrics.AllocationItemsIterated > limits.MaxValidatorsPerBlock {
		return fmt.Errorf("block lifecycle allocation iteration exceeds %d", limits.MaxValidatorsPerBlock)
	}
	if metrics.MaintenanceItemsProcessed > limits.MaxMaintenanceItemsPerBlock {
		return fmt.Errorf("block lifecycle maintenance queue exceeds %d", limits.MaxMaintenanceItemsPerBlock)
	}
	return nil
}

func ValidateDepositMetrics(metrics OperationMetrics, limits ScalabilityLimits) error {
	if err := rejectFullUserScans(metrics); err != nil {
		return err
	}
	if metrics.PoolSharesIterated > 1 || metrics.PoolUsersIterated > 1 || metrics.AccountsIterated > 1 {
		return errors.New("deposit path must touch only caller account and caller pool share")
	}
	if metrics.KeysRead+metrics.KeysWritten > limits.MaxDepositKeys {
		return fmt.Errorf("deposit key touches exceed %d", limits.MaxDepositKeys)
	}
	return nil
}

func ValidateRewardClaimMetrics(metrics OperationMetrics, limits ScalabilityLimits) error {
	if err := rejectFullUserScans(metrics); err != nil {
		return err
	}
	if metrics.PoolSharesIterated > 1 || metrics.PoolUsersIterated > 1 || metrics.AccountsIterated > 1 {
		return errors.New("reward claim must be lazy and touch only caller state")
	}
	if metrics.KeysRead+metrics.KeysWritten > limits.MaxRewardClaimKeys {
		return fmt.Errorf("reward claim key touches exceed %d", limits.MaxRewardClaimKeys)
	}
	return nil
}

func ValidateReputationClaimMetrics(metrics OperationMetrics, limits ScalabilityLimits) error {
	if err := rejectFullUserScans(metrics); err != nil {
		return err
	}
	if metrics.PoolSharesIterated > 1 || metrics.PoolUsersIterated > 1 || metrics.AccountsIterated > 1 {
		return errors.New("reputation claim must be lazy and touch only caller state")
	}
	if metrics.KeysRead+metrics.KeysWritten > limits.MaxReputationClaimKeys {
		return fmt.Errorf("reputation claim key touches exceed %d", limits.MaxReputationClaimKeys)
	}
	return nil
}

func ValidateStorageRentChargeMetrics(metrics OperationMetrics, limits ScalabilityLimits) error {
	if err := rejectFullUserScans(metrics); err != nil {
		return err
	}
	if metrics.AccountsIterated > 1 || metrics.PoolSharesIterated != 0 || metrics.PoolUsersIterated != 0 {
		return errors.New("storage rent charge must be lazy or queue-bounded")
	}
	if metrics.MaintenanceItemsProcessed > 1 && metrics.MaintenanceItemsProcessed > limits.MaxMaintenanceItemsPerBlock {
		return fmt.Errorf("storage rent maintenance queue exceeds %d", limits.MaxMaintenanceItemsPerBlock)
	}
	if metrics.KeysRead+metrics.KeysWritten > limits.MaxStorageRentChargeKeys {
		return fmt.Errorf("storage rent key touches exceed %d", limits.MaxStorageRentChargeKeys)
	}
	return nil
}

func NormalizeQueryPageLimit(requested uint64, limits ScalabilityLimits) (uint64, error) {
	if err := limits.Validate(); err != nil {
		return 0, err
	}
	if requested == 0 {
		return limits.DefaultPageSize, nil
	}
	if requested > limits.MaxPageSize {
		return 0, fmt.Errorf("query page limit %d exceeds max %d", requested, limits.MaxPageSize)
	}
	return requested, nil
}

func ValidateStoreKeyDescriptors(descriptors []StoreKeyDescriptor) error {
	if len(descriptors) == 0 {
		return errors.New("store key descriptors are required")
	}
	seen := map[string]struct{}{}
	for _, descriptor := range descriptors {
		if strings.TrimSpace(descriptor.Name) == "" || strings.TrimSpace(descriptor.Prefix) == "" {
			return errors.New("store key descriptor name and prefix are required")
		}
		if _, found := seen[descriptor.Name]; found {
			return fmt.Errorf("duplicate store key descriptor %s", descriptor.Name)
		}
		seen[descriptor.Name] = struct{}{}
		if !descriptor.Deterministic {
			return fmt.Errorf("store key %s must be deterministic", descriptor.Name)
		}
		if !descriptor.Paginated {
			return fmt.Errorf("store key %s queries must be paginated", descriptor.Name)
		}
	}
	return nil
}

func DefaultScalabilityStoreKeys() []StoreKeyDescriptor {
	return []StoreKeyDescriptor{
		{Name: "account/by_user", Prefix: "account/by_user/", Deterministic: true, Paginated: true},
		{Name: "account/by_raw", Prefix: "account/by_raw/", Deterministic: true, Paginated: true},
		{Name: "pool/share", Prefix: "pool/share/", Deterministic: true, Paginated: true},
		{Name: "pool/allocation", Prefix: "pool/allocation/", Deterministic: true, Paginated: true},
		{Name: "pool/reward", Prefix: "pool/reward/", Deterministic: true, Paginated: true},
		{Name: "reputation/account", Prefix: "reputation/account/", Deterministic: true, Paginated: true},
		{Name: "storage-rent/debt", Prefix: "storage-rent/debt/", Deterministic: true, Paginated: true},
	}
}

func rejectFullUserScans(metrics OperationMetrics) error {
	if metrics.FullAccountScan || metrics.FullPoolUserScan || metrics.FullPoolShareScan {
		return errors.New("full account, pool user, or pool share scans are forbidden in hot paths")
	}
	return nil
}
