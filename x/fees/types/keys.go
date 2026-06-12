package types

import appparams "github.com/sovereign-l1/l1/app/params"

const (
	ModuleName	= "fees"
	StoreKey	= ModuleName
	RouterKey	= ModuleName
)

var (
	ParamsKey		= []byte{0x01}
	ProtocolFeeStateKey	= []byte{0x02}
	BlockTxCountKey		= []byte{0x03}
	SenderTxCountPrefix	= []byte{0x04}
	// CongestionStateKey stores the last-finalized block_utilization_bps (KV-backed, deterministic).
	CongestionStateKey	= []byte{0x05}
)

const (
	BondDenom		= appparams.BaseDenom
	MaxAllowedFeeDenomsV1	= 1
	MaxMinFeeAmountV1	= "1000000000000000000"
	MaxFeeAmountV1		= "1000000000000000000"
	PrototypeBaseFeeAmount	= "1000000"
	PrototypeBaseFeeCoin	= PrototypeBaseFeeAmount + BondDenom
	PrototypeMinGasPriceV1	= "0" + BondDenom

	// TargetTransferFeeNaet is the governance anchor for a normal transfer fee (Requirement 1.2).
	// 10_000_000 naet == 0.01 AET.
	TargetTransferFeeNaet	= int64(10_000_000)

	// DefaultTargetTransferFeeAmount is the string form used in genesis params.
	DefaultTargetTransferFeeAmount	= "10000000"

	// Reputation score boundaries.
	// Neutral reputation is 5000 bps (out of 10000).
	ReputationNeutralScore	= uint32(5_000)
	// Default caps as governance params (in naet).
	DefaultLowReputationPremiumCap		= "500000"	// 0.5 * min_fee headroom
	DefaultHighReputationDiscountCap	= "500000"	// symmetric discount cap

	// Default storage rent side-effect budget per state-creating tx (naet).
	DefaultStorageRentSideEffectsNaet	= "100"

	// Default gas/byte/message fee components (naet).
	DefaultBaseGasFeePerGas	= "1"		// naet per gas unit
	DefaultByteFeeNaet	= "1"		// naet per tx byte
	DefaultMessageFeeNaet	= "1000"	// naet per message
)
