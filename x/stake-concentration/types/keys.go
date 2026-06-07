package types

const (
	ModuleName = "stakeconcentration"
	StoreKey   = ModuleName
	RouterKey  = ModuleName

	BasisPoints uint32 = 10_000

	AetraValidatorSetPhaseOneMax = 150
	AetraValidatorSetPhaseTwoMax = 250

	AetraPhaseOnePowerCapBps  uint32 = 3_000
	AetraPhaseTwoPowerCapBps  uint32 = 2_500
	AetraMatureSetPowerCapBps uint32 = 2_000
)

var (
	ParamsKey  = []byte{0x01}
	NetworkKey = []byte{0x02}
)
