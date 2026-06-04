package types

import appparams "github.com/sovereign-l1/l1/app/params"

const (
	ModuleName = "fees"
	StoreKey   = ModuleName
	RouterKey  = ModuleName
)

var (
	ParamsKey           = []byte{0x01}
	ProtocolFeeStateKey = []byte{0x02}
)

const BondDenom = appparams.BaseDenom
