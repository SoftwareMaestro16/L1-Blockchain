package types

const (
	ModuleName = "dex"
	StoreKey   = ModuleName
	RouterKey  = ModuleName
)

var (
	PoolPrefix        = []byte{0x01}
	NextPoolIDKey     = []byte{0x02}
	DefaultNextPoolID = uint64(1)
)

const (
	LPDenomPrefix  = "lp"
	PoolFeeBps     = int64(30)
	BpsDenominator = int64(10_000)
)
