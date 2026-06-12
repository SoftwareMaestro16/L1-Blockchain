package types

import appparams "github.com/sovereign-l1/l1/app/params"

const (
	ModuleName	= "emissions"
	StoreKey	= ModuleName
	RouterKey	= ModuleName

	BaseDenom		= appparams.BaseDenom
	BasisPoints	uint32	= 10_000
)

var (
	ParamsKey			= []byte{0x01}
	EpochPrefix			= []byte{0x02}
	TotalMintedAccountingKey	= []byte{0x03}
)
