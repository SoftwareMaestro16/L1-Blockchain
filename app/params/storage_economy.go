package params

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
)

const (
	StorageOperationWrite	= "write"
	StorageOperationUpdate	= "update"
	StorageOperationDelete	= "delete"

	StorageClassAccount		= "account"
	StorageClassContract		= "contract"
	StorageClassProtocolCritical	= "protocol_critical"
	StorageClassExempt		= "exempt"

	StorageRentStatusActive			= "active"
	StorageRentStatusWarning		= "warning"
	StorageRentStatusFrozen			= "frozen"
	StorageRentStatusLimitedExec		= "limited_execution"
	StorageRentStatusCleanupEligible	= "cleanup_eligible"
	StorageRentStatusExempt			= "exempt"

	StorageFeeEventWrite	= "storage_write_fee"
	StorageFeeEventUpdate	= "storage_update_fee"
	StorageFeeEventDelete	= "storage_delete_fee"
	StorageFeeEventRefund	= "storage_delete_refund"

	DefaultStateWriteFeePerByteNaet		= int64(4)
	DefaultStateUpdateFeePerByteNaet	= int64(1)
	DefaultDeleteRefundRatioBps		= int64(5_000)
	DefaultDeleteRefundCapBps		= int64(5_000)
	DefaultDeleteRefundDecayBpsPerPeriod	= int64(500)
	DefaultStorageRentRatePerBytePeriod	= int64(1)
	DefaultStorageRentPeriodBlocks		= uint64(43_200)
	DefaultStorageRentWarningPeriods	= uint64(2)
	DefaultStorageRentFreezeGracePeriods	= uint64(1)
	DefaultStorageRentCleanupGracePeriods	= uint64(2)
)

type StorageEconomyParams struct {
	StateWriteFeePerByteNaet	sdkmath.Int
	StateUpdateFeePerByteNaet	sdkmath.Int
	DeleteRefundRatioBps		int64
	DeleteRefundCapBps		int64
	DeleteRefundDecayBpsPerPeriod	int64
	RentRatePerBytePeriodNaet	sdkmath.Int
	RentPeriodBlocks		uint64
	WarningPeriodsBeforeExhaustion	uint64
	FreezeGracePeriods		uint64
	CleanupGracePeriods		uint64
	AccountClassMultiplierBps	int64
	ContractClassMultiplierBps	int64
	ProtocolCriticalExempt		bool
}

type StorageFootprintRecord struct {
	OwnerID			string
	ContractID		string
	Class			string
	Bytes			int64
	OriginalCostNaet	sdkmath.Int
	PrepaidBalanceNaet	sdkmath.Int
	LastRentHeight		uint64
	ConsensusCritical	bool
}

type StorageFeeInput struct {
	OwnerID			string
	ContractID		string
	Class			string
	Operation		string
	CurrentBytes		int64
	DeltaBytes		int64
	DeletedBytes		int64
	OriginalCostNaet	sdkmath.Int
	StorageAgePeriods	uint64
	Params			StorageEconomyParams
}

type StorageFeeEvent struct {
	Type		string
	OwnerID		string
	ContractID	string
	Class		string
	Operation	string
	Bytes		int64
	FeeNaet		sdkmath.Int
	RefundNaet	sdkmath.Int
	FootprintBytes	int64
}

type StorageFeeOutput struct {
	OwnerID			string
	ContractID		string
	Class			string
	Operation		string
	FeeNaet			sdkmath.Int
	RefundNaet		sdkmath.Int
	NetChargeNaet		sdkmath.Int
	RefundCapNaet		sdkmath.Int
	RefundDecayBps		int64
	NewFootprintBytes	int64
	Events			[]StorageFeeEvent
}

type StorageFootprintQueryInput struct {
	Records		[]StorageFootprintRecord
	OwnerID		string
	ContractID	string
}

type StorageFootprintQueryOutput struct {
	Records			[]StorageFootprintRecord
	TotalBytes		int64
	TotalPrepaidNaet	sdkmath.Int
	AccountBytes		int64
	ContractBytes		int64
	ConsensusCriticalBytes	int64
}

type StorageRentInput struct {
	Record		StorageFootprintRecord
	CurrentHeight	uint64
	Params		StorageEconomyParams
}

type StorageRentStatus struct {
	OwnerID				string
	ContractID			string
	Class				string
	Status				string
	Bytes				int64
	RentDueNaet			sdkmath.Int
	PrepaidBalanceNaet		sdkmath.Int
	PeriodsElapsed			uint64
	PeriodsCovered			uint64
	PeriodsUntilExhaustion		uint64
	RecoveryRequiredNaet		sdkmath.Int
	CanExecute			bool
	LimitedExecution		bool
	Frozen				bool
	CleanupEligible			bool
	ConsensusCriticalProtected	bool
	Events				[]StorageFeeEvent
}

func DefaultStorageEconomyParams() StorageEconomyParams {
	return StorageEconomyParams{
		StateWriteFeePerByteNaet:	sdkmath.NewInt(DefaultStateWriteFeePerByteNaet),
		StateUpdateFeePerByteNaet:	sdkmath.NewInt(DefaultStateUpdateFeePerByteNaet),
		DeleteRefundRatioBps:		DefaultDeleteRefundRatioBps,
		DeleteRefundCapBps:		DefaultDeleteRefundCapBps,
		DeleteRefundDecayBpsPerPeriod:	DefaultDeleteRefundDecayBpsPerPeriod,
		RentRatePerBytePeriodNaet:	sdkmath.NewInt(DefaultStorageRentRatePerBytePeriod),
		RentPeriodBlocks:		DefaultStorageRentPeriodBlocks,
		WarningPeriodsBeforeExhaustion:	DefaultStorageRentWarningPeriods,
		FreezeGracePeriods:		DefaultStorageRentFreezeGracePeriods,
		CleanupGracePeriods:		DefaultStorageRentCleanupGracePeriods,
		AccountClassMultiplierBps:	BasisPoints,
		ContractClassMultiplierBps:	BasisPoints,
		ProtocolCriticalExempt:		true,
	}
}

func ComputeStorageFee(input StorageFeeInput) (StorageFeeOutput, error) {
	params := input.Params.withDefaults()
	if err := params.Validate(); err != nil {
		return StorageFeeOutput{}, err
	}
	if err := validateStorageFeeInput(input); err != nil {
		return StorageFeeOutput{}, err
	}
	if isStorageClassExempt(input.Class, params) {
		return StorageFeeOutput{
			OwnerID:		input.OwnerID,
			ContractID:		input.ContractID,
			Class:			input.Class,
			Operation:		input.Operation,
			FeeNaet:		sdkmath.ZeroInt(),
			RefundNaet:		sdkmath.ZeroInt(),
			NetChargeNaet:		sdkmath.ZeroInt(),
			RefundCapNaet:		sdkmath.ZeroInt(),
			NewFootprintBytes:	input.CurrentBytes,
		}, nil
	}

	fee := sdkmath.ZeroInt()
	refund := sdkmath.ZeroInt()
	refundCap := sdkmath.ZeroInt()
	refundDecayBps := int64(0)
	newFootprint := input.CurrentBytes
	events := make([]StorageFeeEvent, 0, 2)

	switch input.Operation {
	case StorageOperationWrite:
		if input.DeltaBytes <= 0 {
			return StorageFeeOutput{}, fmt.Errorf("write delta_bytes must be positive")
		}
		fee = classAdjustedAmount(params.StateWriteFeePerByteNaet.MulRaw(input.DeltaBytes), input.Class, params)
		newFootprint += input.DeltaBytes
		events = append(events, storageEvent(StorageFeeEventWrite, input, input.DeltaBytes, fee, sdkmath.ZeroInt(), newFootprint))
	case StorageOperationUpdate:
		if input.DeltaBytes == 0 {
			return StorageFeeOutput{}, fmt.Errorf("update delta_bytes must not be zero")
		}
		chargedBytes := absInt64(input.DeltaBytes)
		fee = classAdjustedAmount(params.StateUpdateFeePerByteNaet.MulRaw(chargedBytes), input.Class, params)
		newFootprint += input.DeltaBytes
		if newFootprint < 0 {
			return StorageFeeOutput{}, fmt.Errorf("storage update cannot make footprint negative")
		}
		events = append(events, storageEvent(StorageFeeEventUpdate, input, chargedBytes, fee, sdkmath.ZeroInt(), newFootprint))
	case StorageOperationDelete:
		if input.DeletedBytes <= 0 {
			return StorageFeeOutput{}, fmt.Errorf("delete deleted_bytes must be positive")
		}
		if input.DeletedBytes > input.CurrentBytes {
			return StorageFeeOutput{}, fmt.Errorf("delete cannot exceed current footprint")
		}
		refund, refundCap, refundDecayBps = computeDeleteRefund(input, params)
		newFootprint -= input.DeletedBytes
		events = append(events,
			storageEvent(StorageFeeEventDelete, input, input.DeletedBytes, sdkmath.ZeroInt(), sdkmath.ZeroInt(), newFootprint),
			storageEvent(StorageFeeEventRefund, input, input.DeletedBytes, sdkmath.ZeroInt(), refund, newFootprint),
		)
	default:
		return StorageFeeOutput{}, fmt.Errorf("unsupported storage operation %q", input.Operation)
	}

	return StorageFeeOutput{
		OwnerID:		input.OwnerID,
		ContractID:		input.ContractID,
		Class:			input.Class,
		Operation:		input.Operation,
		FeeNaet:		fee,
		RefundNaet:		refund,
		NetChargeNaet:		fee.Sub(refund),
		RefundCapNaet:		refundCap,
		RefundDecayBps:		refundDecayBps,
		NewFootprintBytes:	newFootprint,
		Events:			events,
	}, nil
}

func QueryStorageFootprint(input StorageFootprintQueryInput) (StorageFootprintQueryOutput, error) {
	records := make([]StorageFootprintRecord, 0, len(input.Records))
	totalBytes := int64(0)
	accountBytes := int64(0)
	contractBytes := int64(0)
	criticalBytes := int64(0)
	totalPrepaid := sdkmath.ZeroInt()
	for _, record := range input.Records {
		if err := validateStorageFootprint(record); err != nil {
			return StorageFootprintQueryOutput{}, err
		}
		if input.OwnerID != "" && record.OwnerID != input.OwnerID {
			continue
		}
		if input.ContractID != "" && record.ContractID != input.ContractID {
			continue
		}
		records = append(records, record)
		totalBytes += record.Bytes
		totalPrepaid = totalPrepaid.Add(normalizeInt(record.PrepaidBalanceNaet))
		switch record.Class {
		case StorageClassAccount:
			accountBytes += record.Bytes
		case StorageClassContract:
			contractBytes += record.Bytes
		case StorageClassProtocolCritical:
			criticalBytes += record.Bytes
		}
	}
	return StorageFootprintQueryOutput{
		Records:		records,
		TotalBytes:		totalBytes,
		TotalPrepaidNaet:	totalPrepaid,
		AccountBytes:		accountBytes,
		ContractBytes:		contractBytes,
		ConsensusCriticalBytes:	criticalBytes,
	}, nil
}

func ComputeStorageRentStatus(input StorageRentInput) (StorageRentStatus, error) {
	params := input.Params.withDefaults()
	if err := params.Validate(); err != nil {
		return StorageRentStatus{}, err
	}
	if err := validateStorageFootprint(input.Record); err != nil {
		return StorageRentStatus{}, err
	}
	if input.CurrentHeight < input.Record.LastRentHeight {
		return StorageRentStatus{}, fmt.Errorf("current_height must be >= last_rent_height")
	}
	record := input.Record
	if isStorageClassExempt(record.Class, params) || record.ConsensusCritical {
		return StorageRentStatus{
			OwnerID:			record.OwnerID,
			ContractID:			record.ContractID,
			Class:				record.Class,
			Status:				StorageRentStatusExempt,
			Bytes:				record.Bytes,
			PrepaidBalanceNaet:		normalizeInt(record.PrepaidBalanceNaet),
			CanExecute:			true,
			ConsensusCriticalProtected:	record.ConsensusCritical || record.Class == StorageClassProtocolCritical,
		}, nil
	}

	periodsElapsed := uint64(0)
	if params.RentPeriodBlocks > 0 {
		periodsElapsed = (input.CurrentHeight - record.LastRentHeight) / params.RentPeriodBlocks
	}
	periodRent := classAdjustedAmount(params.RentRatePerBytePeriodNaet.MulRaw(record.Bytes), record.Class, params)
	rentDue := periodRent.MulRaw(int64(periodsElapsed))
	prepaid := normalizeInt(record.PrepaidBalanceNaet)
	coveredPeriods := uint64(0)
	if periodRent.IsPositive() {
		coveredPeriods = prepaid.Quo(periodRent).Uint64()
	}
	status := StorageRentStatusActive
	canExecute := true
	limited := false
	frozen := false
	cleanup := false
	recovery := sdkmath.ZeroInt()
	periodsUntilExhaustion := uint64(0)
	if coveredPeriods > periodsElapsed {
		periodsUntilExhaustion = coveredPeriods - periodsElapsed
		if periodsUntilExhaustion <= params.WarningPeriodsBeforeExhaustion {
			status = StorageRentStatusWarning
		}
	} else {
		overduePeriods := periodsElapsed - coveredPeriods
		recovery = periodRent.MulRaw(int64(overduePeriods + 1))
		switch {
		case overduePeriods <= params.FreezeGracePeriods:
			status = StorageRentStatusFrozen
			canExecute = false
			frozen = true
		case overduePeriods <= params.FreezeGracePeriods+params.CleanupGracePeriods:
			status = StorageRentStatusLimitedExec
			limited = true
			canExecute = true
		default:
			status = StorageRentStatusCleanupEligible
			cleanup = true
			canExecute = false
		}
	}
	return StorageRentStatus{
		OwnerID:		record.OwnerID,
		ContractID:		record.ContractID,
		Class:			record.Class,
		Status:			status,
		Bytes:			record.Bytes,
		RentDueNaet:		rentDue,
		PrepaidBalanceNaet:	prepaid,
		PeriodsElapsed:		periodsElapsed,
		PeriodsCovered:		coveredPeriods,
		PeriodsUntilExhaustion:	periodsUntilExhaustion,
		RecoveryRequiredNaet:	recovery,
		CanExecute:		canExecute,
		LimitedExecution:	limited,
		Frozen:			frozen,
		CleanupEligible:	cleanup,
	}, nil
}

func (p StorageEconomyParams) Validate() error {
	if normalizeInt(p.StateWriteFeePerByteNaet).IsNegative() {
		return fmt.Errorf("state_write_fee_per_byte_naet must not be negative")
	}
	if normalizeInt(p.StateUpdateFeePerByteNaet).IsNegative() {
		return fmt.Errorf("state_update_fee_per_byte_naet must not be negative")
	}
	if normalizeInt(p.RentRatePerBytePeriodNaet).IsNegative() {
		return fmt.Errorf("rent_rate_per_byte_period_naet must not be negative")
	}
	if err := validateBps("delete_refund_ratio_bps", p.DeleteRefundRatioBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("delete_refund_cap_bps", p.DeleteRefundCapBps, 0, BasisPoints); err != nil {
		return err
	}
	if err := validateBps("delete_refund_decay_bps_per_period", p.DeleteRefundDecayBpsPerPeriod, 0, BasisPoints); err != nil {
		return err
	}
	if p.RentPeriodBlocks == 0 {
		return fmt.Errorf("rent_period_blocks must be positive")
	}
	if err := validateBps("account_class_multiplier_bps", p.AccountClassMultiplierBps, 0, DefaultMaxLoadMultiplierBps); err != nil {
		return err
	}
	return validateBps("contract_class_multiplier_bps", p.ContractClassMultiplierBps, 0, DefaultMaxLoadMultiplierBps)
}

func (p StorageEconomyParams) withDefaults() StorageEconomyParams {
	defaults := DefaultStorageEconomyParams()
	if p.StateWriteFeePerByteNaet.IsNil() {
		p.StateWriteFeePerByteNaet = defaults.StateWriteFeePerByteNaet
	}
	if p.StateUpdateFeePerByteNaet.IsNil() {
		p.StateUpdateFeePerByteNaet = defaults.StateUpdateFeePerByteNaet
	}
	if p.RentRatePerBytePeriodNaet.IsNil() {
		p.RentRatePerBytePeriodNaet = defaults.RentRatePerBytePeriodNaet
	}
	if p.DeleteRefundRatioBps == 0 {
		p.DeleteRefundRatioBps = defaults.DeleteRefundRatioBps
	}
	if p.DeleteRefundCapBps == 0 {
		p.DeleteRefundCapBps = defaults.DeleteRefundCapBps
	}
	if p.DeleteRefundDecayBpsPerPeriod == 0 {
		p.DeleteRefundDecayBpsPerPeriod = defaults.DeleteRefundDecayBpsPerPeriod
	}
	if p.RentPeriodBlocks == 0 {
		p.RentPeriodBlocks = defaults.RentPeriodBlocks
	}
	if p.WarningPeriodsBeforeExhaustion == 0 {
		p.WarningPeriodsBeforeExhaustion = defaults.WarningPeriodsBeforeExhaustion
	}
	if p.FreezeGracePeriods == 0 {
		p.FreezeGracePeriods = defaults.FreezeGracePeriods
	}
	if p.CleanupGracePeriods == 0 {
		p.CleanupGracePeriods = defaults.CleanupGracePeriods
	}
	if p.AccountClassMultiplierBps == 0 {
		p.AccountClassMultiplierBps = defaults.AccountClassMultiplierBps
	}
	if p.ContractClassMultiplierBps == 0 {
		p.ContractClassMultiplierBps = defaults.ContractClassMultiplierBps
	}
	if !p.ProtocolCriticalExempt {
		p.ProtocolCriticalExempt = defaults.ProtocolCriticalExempt
	}
	return p
}

func validateStorageFeeInput(input StorageFeeInput) error {
	if input.OwnerID == "" && input.ContractID == "" {
		return fmt.Errorf("owner_id or contract_id is required")
	}
	if !isKnownStorageClass(input.Class) {
		return fmt.Errorf("unknown storage class %q", input.Class)
	}
	if input.CurrentBytes < 0 {
		return fmt.Errorf("current_bytes must not be negative")
	}
	if normalizeInt(input.OriginalCostNaet).IsNegative() {
		return fmt.Errorf("original_cost_naet must not be negative")
	}
	return nil
}

func validateStorageFootprint(record StorageFootprintRecord) error {
	if record.OwnerID == "" && record.ContractID == "" {
		return fmt.Errorf("owner_id or contract_id is required")
	}
	if !isKnownStorageClass(record.Class) {
		return fmt.Errorf("unknown storage class %q", record.Class)
	}
	if record.Bytes < 0 {
		return fmt.Errorf("storage bytes must not be negative")
	}
	if normalizeInt(record.OriginalCostNaet).IsNegative() {
		return fmt.Errorf("original_cost_naet must not be negative")
	}
	if normalizeInt(record.PrepaidBalanceNaet).IsNegative() {
		return fmt.Errorf("prepaid_balance_naet must not be negative")
	}
	return nil
}

func computeDeleteRefund(input StorageFeeInput, params StorageEconomyParams) (sdkmath.Int, sdkmath.Int, int64) {
	originalCost := normalizeInt(input.OriginalCostNaet)
	if originalCost.IsZero() {
		originalCost = params.StateWriteFeePerByteNaet.MulRaw(input.DeletedBytes)
	}
	baseRefund := ApplyBps(originalCost, params.DeleteRefundRatioBps)
	decayBps := clampInt64(int64(input.StorageAgePeriods)*params.DeleteRefundDecayBpsPerPeriod, 0, BasisPoints)
	afterDecay := ApplyBps(baseRefund, BasisPoints-decayBps)
	cap := ApplyBps(originalCost, params.DeleteRefundCapBps)
	refund := minInt(afterDecay, cap)
	return refund, cap, decayBps
}

func classAdjustedAmount(amount sdkmath.Int, class string, params StorageEconomyParams) sdkmath.Int {
	switch class {
	case StorageClassContract:
		return ApplyBps(amount, params.ContractClassMultiplierBps)
	case StorageClassAccount:
		return ApplyBps(amount, params.AccountClassMultiplierBps)
	default:
		return amount
	}
}

func storageEvent(eventType string, input StorageFeeInput, bytes int64, fee, refund sdkmath.Int, footprint int64) StorageFeeEvent {
	return StorageFeeEvent{
		Type:		eventType,
		OwnerID:	input.OwnerID,
		ContractID:	input.ContractID,
		Class:		input.Class,
		Operation:	input.Operation,
		Bytes:		bytes,
		FeeNaet:	fee,
		RefundNaet:	refund,
		FootprintBytes:	footprint,
	}
}

func isKnownStorageClass(class string) bool {
	switch class {
	case StorageClassAccount, StorageClassContract, StorageClassProtocolCritical, StorageClassExempt:
		return true
	default:
		return false
	}
}

func isStorageClassExempt(class string, params StorageEconomyParams) bool {
	return class == StorageClassExempt || (params.ProtocolCriticalExempt && class == StorageClassProtocolCritical)
}
