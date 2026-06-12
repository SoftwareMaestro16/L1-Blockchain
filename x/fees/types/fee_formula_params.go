package types

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
)

// FeeFormulaParams holds the extended fee formula governance parameters that
// supplement the base Params. All values are in naet unless noted.
//
// Formula (Requirement 1.1):
//
//	transfer_fee_naet = max(min_tx_fee_naet, base_transfer_fee_naet)
//	  + gas_used * current_base_fee_per_gas_naet
//	  + tx_size_bytes * byte_fee_naet
//	  + message_count * message_fee_naet
//	  + bounded_congestion_surcharge_naet
//	  + low_reputation_premium_naet
//	  + storage_rent_side_effects_naet
//	  - bounded_reputation_discount_naet
type FeeFormulaParams struct {
	// TargetTransferFeeNaet is the anchor fee for a normal transfer (Requirement 1.2).
	// Default: 10_000_000 naet == 0.01 AET.
	TargetTransferFeeNaet	string	`json:"target_transfer_fee_naet"`

	// BaseFeePerGasNaet is the cost per gas unit in naet.
	BaseFeePerGasNaet	string	`json:"base_fee_per_gas_naet"`

	// ByteFeeNaet is the cost per transaction byte in naet.
	ByteFeeNaet	string	`json:"byte_fee_naet"`

	// MessageFeeNaet is the cost per message in naet.
	MessageFeeNaet	string	`json:"message_fee_naet"`

	// MaxCongestionSurchargeNaet is the upper bound of the congestion surcharge.
	// The actual surcharge is proportional to block utilization above the threshold.
	MaxCongestionSurchargeNaet	string	`json:"max_congestion_surcharge_naet"`

	// LowReputationPremiumCapNaet is the maximum bounded premium added for low-reputation
	// senders (Requirement 1.4). Never blocks a transaction.
	LowReputationPremiumCapNaet	string	`json:"low_reputation_premium_cap_naet"`

	// HighReputationDiscountCapNaet is the maximum bounded discount applied for
	// high-reputation senders (Requirement 1.5). Never zeroes the protocol fee.
	HighReputationDiscountCapNaet	string	`json:"high_reputation_discount_cap_naet"`

	// StorageRentSideEffectsNaet is the default fee budget for transactions that create
	// or increase persistent state (Requirement 6.6). May be overridden per-tx.
	StorageRentSideEffectsNaet	string	`json:"storage_rent_side_effects_naet"`
}

// DefaultFeeFormulaParams returns safe governance defaults for the extended fee formula.
func DefaultFeeFormulaParams() FeeFormulaParams {
	return FeeFormulaParams{
		TargetTransferFeeNaet:		DefaultTargetTransferFeeAmount,
		BaseFeePerGasNaet:		DefaultBaseGasFeePerGas,
		ByteFeeNaet:			DefaultByteFeeNaet,
		MessageFeeNaet:			DefaultMessageFeeNaet,
		MaxCongestionSurchargeNaet:	"2000000",
		LowReputationPremiumCapNaet:	DefaultLowReputationPremiumCap,
		HighReputationDiscountCapNaet:	DefaultHighReputationDiscountCap,
		StorageRentSideEffectsNaet:	DefaultStorageRentSideEffectsNaet,
	}
}

// Validate checks that all FeeFormulaParams are within acceptable bounds.
func (p FeeFormulaParams) Validate() error {
	if _, err := p.TargetTransferFeeInt(); err != nil {
		return err
	}
	if _, err := validateNonNegativeAmount("base_fee_per_gas_naet", p.BaseFeePerGasNaet); err != nil {
		return err
	}
	if _, err := validateNonNegativeAmount("byte_fee_naet", p.ByteFeeNaet); err != nil {
		return err
	}
	if _, err := validateNonNegativeAmount("message_fee_naet", p.MessageFeeNaet); err != nil {
		return err
	}
	if _, err := validateNonNegativeAmount("max_congestion_surcharge_naet", p.MaxCongestionSurchargeNaet); err != nil {
		return err
	}
	if _, err := validateNonNegativeAmount("low_reputation_premium_cap_naet", p.LowReputationPremiumCapNaet); err != nil {
		return err
	}
	if _, err := validateNonNegativeAmount("high_reputation_discount_cap_naet", p.HighReputationDiscountCapNaet); err != nil {
		return err
	}
	if _, err := validateNonNegativeAmount("storage_rent_side_effects_naet", p.StorageRentSideEffectsNaet); err != nil {
		return err
	}
	return nil
}

// NormalizeFeeFormulaParams fills zero/empty fields with defaults.
func NormalizeFeeFormulaParams(p FeeFormulaParams) FeeFormulaParams {
	def := DefaultFeeFormulaParams()
	if p.TargetTransferFeeNaet == "" {
		p.TargetTransferFeeNaet = def.TargetTransferFeeNaet
	}
	if p.BaseFeePerGasNaet == "" {
		p.BaseFeePerGasNaet = def.BaseFeePerGasNaet
	}
	if p.ByteFeeNaet == "" {
		p.ByteFeeNaet = def.ByteFeeNaet
	}
	if p.MessageFeeNaet == "" {
		p.MessageFeeNaet = def.MessageFeeNaet
	}
	if p.MaxCongestionSurchargeNaet == "" {
		p.MaxCongestionSurchargeNaet = def.MaxCongestionSurchargeNaet
	}
	if p.LowReputationPremiumCapNaet == "" {
		p.LowReputationPremiumCapNaet = def.LowReputationPremiumCapNaet
	}
	if p.HighReputationDiscountCapNaet == "" {
		p.HighReputationDiscountCapNaet = def.HighReputationDiscountCapNaet
	}
	if p.StorageRentSideEffectsNaet == "" {
		p.StorageRentSideEffectsNaet = def.StorageRentSideEffectsNaet
	}
	return p
}

// TargetTransferFeeInt parses TargetTransferFeeNaet to sdkmath.Int.
func (p FeeFormulaParams) TargetTransferFeeInt() (sdkmath.Int, error) {
	amount, ok := sdkmath.NewIntFromString(p.TargetTransferFeeNaet)
	if !ok || !amount.IsPositive() {
		return sdkmath.Int{}, fmt.Errorf("target_transfer_fee_naet must be a positive integer, got %q", p.TargetTransferFeeNaet)
	}
	return amount, nil
}

// BaseFeePerGasInt parses BaseFeePerGasNaet.
func (p FeeFormulaParams) BaseFeePerGasInt() (sdkmath.Int, error) {
	return parseNonNegativeInt("base_fee_per_gas_naet", p.BaseFeePerGasNaet)
}

// ByteFeeInt parses ByteFeeNaet.
func (p FeeFormulaParams) ByteFeeInt() (sdkmath.Int, error) {
	return parseNonNegativeInt("byte_fee_naet", p.ByteFeeNaet)
}

// MessageFeeInt parses MessageFeeNaet.
func (p FeeFormulaParams) MessageFeeInt() (sdkmath.Int, error) {
	return parseNonNegativeInt("message_fee_naet", p.MessageFeeNaet)
}

// MaxCongestionSurchargeInt parses MaxCongestionSurchargeNaet.
func (p FeeFormulaParams) MaxCongestionSurchargeInt() (sdkmath.Int, error) {
	return parseNonNegativeInt("max_congestion_surcharge_naet", p.MaxCongestionSurchargeNaet)
}

// LowReputationPremiumCapInt parses LowReputationPremiumCapNaet.
func (p FeeFormulaParams) LowReputationPremiumCapInt() (sdkmath.Int, error) {
	return parseNonNegativeInt("low_reputation_premium_cap_naet", p.LowReputationPremiumCapNaet)
}

// HighReputationDiscountCapInt parses HighReputationDiscountCapNaet.
func (p FeeFormulaParams) HighReputationDiscountCapInt() (sdkmath.Int, error) {
	return parseNonNegativeInt("high_reputation_discount_cap_naet", p.HighReputationDiscountCapNaet)
}

// StorageRentSideEffectsInt parses StorageRentSideEffectsNaet.
func (p FeeFormulaParams) StorageRentSideEffectsInt() (sdkmath.Int, error) {
	return parseNonNegativeInt("storage_rent_side_effects_naet", p.StorageRentSideEffectsNaet)
}

// ComputeFullTransferFee calculates the complete deterministic fee for a transaction
// using all formula components from Requirement 1.1.
//
//	transfer_fee_naet = max(min_tx_fee_naet, base_transfer_fee_naet)
//	  + gas_used * current_base_fee_per_gas_naet
//	  + tx_size_bytes * byte_fee_naet
//	  + message_count * message_fee_naet
//	  + bounded_congestion_surcharge_naet
//	  + low_reputation_premium_naet
//	  + storage_rent_side_effects_naet
//	  - bounded_reputation_discount_naet
func ComputeFullTransferFee(
	baseParams Params,
	formulaParams FeeFormulaParams,
	gasUsed uint64,
	txSizeBytes uint64,
	messageCount uint64,
	blockUtilizationBps uint32,
	reputationScore uint32,
	reputationFound bool,
	storageRentSideEffectsNaet sdkmath.Int,
) (sdkmath.Int, error) {
	baseParams = NormalizeParams(baseParams)
	formulaParams = NormalizeFeeFormulaParams(formulaParams)

	if err := baseParams.Validate(); err != nil {
		return sdkmath.Int{}, err
	}
	if err := formulaParams.Validate(); err != nil {
		return sdkmath.Int{}, err
	}

	minFee, err := baseParams.MinFeeInt()
	if err != nil {
		return sdkmath.Int{}, err
	}
	baseFee, err := baseParams.BaseFeeInt()
	if err != nil {
		return sdkmath.Int{}, err
	}
	base := minFee
	if baseFee.GT(minFee) {
		base = baseFee
	}

	gasFeePerGas, err := formulaParams.BaseFeePerGasInt()
	if err != nil {
		return sdkmath.Int{}, err
	}
	gasComponent := gasFeePerGas.MulRaw(int64(gasUsed))

	byteFee, err := formulaParams.ByteFeeInt()
	if err != nil {
		return sdkmath.Int{}, err
	}
	byteComponent := byteFee.MulRaw(int64(txSizeBytes))

	msgFee, err := formulaParams.MessageFeeInt()
	if err != nil {
		return sdkmath.Int{}, err
	}
	msgComponent := msgFee.MulRaw(int64(messageCount))

	maxSurcharge, err := formulaParams.MaxCongestionSurchargeInt()
	if err != nil {
		return sdkmath.Int{}, err
	}
	congestionSurcharge := computeBoundedCongestionSurcharge(maxSurcharge, blockUtilizationBps, baseParams.CongestionThresholdBps)

	lowPremiumCap, err := formulaParams.LowReputationPremiumCapInt()
	if err != nil {
		return sdkmath.Int{}, err
	}
	discountCap, err := formulaParams.HighReputationDiscountCapInt()
	if err != nil {
		return sdkmath.Int{}, err
	}
	premium, discount := computeReputationAdjustments(reputationScore, reputationFound, lowPremiumCap, discountCap)

	storageRent := storageRentSideEffectsNaet
	if storageRent.IsNil() || storageRent.IsNegative() {
		storageRent = sdkmath.ZeroInt()
	}

	total := base.
		Add(gasComponent).
		Add(byteComponent).
		Add(msgComponent).
		Add(congestionSurcharge).
		Add(premium).
		Add(storageRent).
		Sub(discount)

	if total.LT(minFee) {
		total = minFee
	}

	return total, nil
}

// computeBoundedCongestionSurcharge computes a surcharge proportional to how far
// block utilization exceeds the congestion threshold. This uses only KV-state bps
// (deterministic), never wall-clock or mempool data (Requirement 1.3).
func computeBoundedCongestionSurcharge(maxSurcharge sdkmath.Int, utilizationBps, thresholdBps uint32) sdkmath.Int {
	if utilizationBps <= thresholdBps || maxSurcharge.IsZero() {
		return sdkmath.ZeroInt()
	}
	remainingBps := uint64(BasisPoints) - uint64(thresholdBps)
	if remainingBps == 0 {
		return maxSurcharge
	}
	overBps := uint64(utilizationBps - thresholdBps)

	surcharge := maxSurcharge.MulRaw(int64(overBps)).QuoRaw(int64(remainingBps))
	if surcharge.GT(maxSurcharge) {
		return maxSurcharge
	}
	return surcharge
}

// computeReputationAdjustments returns (premium, discount) in naet for the given
// reputation score. Neutral score (5000) → both zero. Low score → bounded premium.
// High score → bounded discount (Requirement 1.4, 1.5).
//
// Score is in [0..10000] bps where ReputationNeutralScore (5000) == neutral.
func computeReputationAdjustments(score uint32, found bool, premiumCap, discountCap sdkmath.Int) (premium, discount sdkmath.Int) {
	if !found {

		return sdkmath.ZeroInt(), sdkmath.ZeroInt()
	}

	neutral := ReputationNeutralScore

	if score < neutral {

		deficit := uint64(neutral - score)
		p := premiumCap.MulRaw(int64(deficit)).QuoRaw(int64(neutral))
		if p.GT(premiumCap) {
			p = premiumCap
		}
		return p, sdkmath.ZeroInt()
	}

	if score > neutral {

		excess := uint64(score - neutral)
		remaining := uint64(BasisPoints) - uint64(neutral)
		if remaining == 0 {
			return sdkmath.ZeroInt(), discountCap
		}
		d := discountCap.MulRaw(int64(excess)).QuoRaw(int64(remaining))
		if d.GT(discountCap) {
			d = discountCap
		}
		return sdkmath.ZeroInt(), d
	}

	return sdkmath.ZeroInt(), sdkmath.ZeroInt()
}

// validateNonNegativeAmount validates that a string integer is >= 0.
func validateNonNegativeAmount(name, value string) (sdkmath.Int, error) {
	return parseNonNegativeInt(name, value)
}

func parseNonNegativeInt(name, value string) (sdkmath.Int, error) {
	if value == "" {
		return sdkmath.ZeroInt(), nil
	}
	amount, ok := sdkmath.NewIntFromString(value)
	if !ok {
		return sdkmath.Int{}, fmt.Errorf("%s must be a non-negative integer, got %q", name, value)
	}
	if amount.IsNegative() {
		return sdkmath.Int{}, fmt.Errorf("%s must be non-negative, got %s", name, amount.String())
	}
	return amount, nil
}
