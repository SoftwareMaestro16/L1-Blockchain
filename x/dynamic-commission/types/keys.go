package types

const (
	ModuleName	= "dynamiccommission"
	StoreKey	= ModuleName
	RouterKey	= ModuleName
)

var (
	ParamsKey		= []byte{0x01}
	CommissionPrefix	= []byte{0x02}
	HistoryPrefix		= []byte{0x03}
)

const (
	BasisPoints uint32 = 10_000
)
