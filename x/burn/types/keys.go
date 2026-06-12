package types

import appparams "github.com/sovereign-l1/l1/app/params"

const (
	ModuleName	= "burn"
	StoreKey	= ModuleName
	RouterKey	= ModuleName
	BaseDenom	= appparams.BaseDenom
)

var (
	ParamsKey		= []byte{0x01}
	BurnedDenomPrefix	= []byte{0x02}
	BurnedEpochPrefix	= []byte{0x03}
	BurnReasonPrefix	= []byte{0x04}
	NextBurnReasonIDKey	= []byte{0x05}
)

const DefaultMaxReasonBytes = uint32(256)
