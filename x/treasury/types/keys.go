package types

import appparams "github.com/sovereign-l1/l1/app/params"

const (
	ModuleName	= "treasury"
	StoreKey	= ModuleName
	RouterKey	= ModuleName

	TreasuryModuleName	= "feecollector_treasury"
	BaseDenom		= appparams.BaseDenom
	BasisPoints		= uint32(10_000)

	BucketReserve			= "reserve"
	BucketEcosystem			= "ecosystem"
	BucketValidatorIncentives	= "validator_incentives"
	BucketBurn			= "burn"

	StatusPending	= "pending"
	StatusApproved	= "approved"
	StatusRejected	= "rejected"
	StatusExecuted	= "executed"
	StatusCanceled	= "canceled"
)

var (
	ParamsKey		= []byte{0x01}
	AllocationsKey		= []byte{0x02}
	SpendPrefix		= []byte{0x03}
	EpochSpendPrefix	= []byte{0x04}
	NextSpendIDKey		= []byte{0x05}
)

const DefaultMaxMetadataBytes = uint32(512)
