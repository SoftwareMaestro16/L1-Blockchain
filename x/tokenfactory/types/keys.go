package types

const (
	ModuleName = "tokenfactory"
	StoreKey   = ModuleName
	RouterKey  = ModuleName
)

var (
	DenomPrefix = []byte{0x01}
	ParamsKey   = []byte{0x02}
)

const FactoryDenomPrefix = "factory"
