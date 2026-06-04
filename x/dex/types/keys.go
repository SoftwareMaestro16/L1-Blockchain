package types

const (
	ModuleName = "dex"
	StoreKey   = ModuleName
	RouterKey  = ModuleName
)

var (
	PoolPrefix        = []byte{0x01}
	NextPoolIDKey     = []byte{0x02}
	ParamsKey         = []byte{0x03}
	PoolPairPrefix    = []byte{0x04}
	DefaultNextPoolID = uint64(1)
)

const (
	LPDenomPrefix        = "lp"
	DefaultSwapFeeBps    = uint32(30)
	DefaultMaxSwapFeeBps = uint32(1_000)
	BpsDenominator       = int64(10_000)
)
