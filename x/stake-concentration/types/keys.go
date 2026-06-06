package types

const (
	ModuleName = "stakeconcentration"
	StoreKey   = ModuleName
	RouterKey  = ModuleName

	BasisPoints uint32 = 10_000
)

var (
	ParamsKey  = []byte{0x01}
	NetworkKey = []byte{0x02}
)
