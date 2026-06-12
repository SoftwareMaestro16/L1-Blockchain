package params

import banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

const (
	ChainName		= "Aetra"
	BaseDenom		= "naet"
	DisplayDenom		= "AET"
	TokenName		= "Aetra"
	TokenSymbol		= "AET"
	DisplayDenomExponent	= uint32(9)
	BaseUnitsPerDisplay	= int64(1_000_000_000)
)

func NativeTokenMetadata() banktypes.Metadata {
	return banktypes.Metadata{
		Description:	"The native staking and fee token of Aetra.",
		Base:		BaseDenom,
		Display:	DisplayDenom,
		Name:		TokenName,
		Symbol:		TokenSymbol,
		DenomUnits: []*banktypes.DenomUnit{
			{Denom: BaseDenom, Exponent: 0},
			{Denom: DisplayDenom, Exponent: DisplayDenomExponent},
		},
	}
}

func EnsureNativeTokenMetadata(metadata []banktypes.Metadata) []banktypes.Metadata {
	native := NativeTokenMetadata()
	for i := range metadata {
		if metadata[i].Base == native.Base {
			metadata[i] = native
			return metadata
		}
	}
	return append(metadata, native)
}
