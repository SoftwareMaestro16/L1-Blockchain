package params

import banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

const (
	ChainName            = "Orbitalis"
	BaseDenom            = "norb"
	DisplayDenom         = "ORB"
	TokenName            = "Orbitalis"
	TokenSymbol          = "ORB"
	DisplayDenomExponent = uint32(9)
)

func NativeTokenMetadata() banktypes.Metadata {
	return banktypes.Metadata{
		Description: "The native staking and fee token of Orbitalis.",
		Base:        BaseDenom,
		Display:     DisplayDenom,
		Name:        TokenName,
		Symbol:      TokenSymbol,
		DenomUnits: []*banktypes.DenomUnit{
			{Denom: BaseDenom, Exponent: 0, Aliases: []string{}},
			{Denom: DisplayDenom, Exponent: DisplayDenomExponent, Aliases: []string{}},
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
