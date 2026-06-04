package types

import (
	"strings"

	appparams "github.com/sovereign-l1/l1/app/params"
)

const (
	ModuleName = "tokenfactory"
	StoreKey   = ModuleName
	RouterKey  = ModuleName
)

var (
	DenomPrefix = []byte{0x01}
)

const FactoryDenomPrefix = "factory"

func IsReservedNativeSubdenom(subdenom string) bool {
	normalized := strings.ToLower(strings.TrimSpace(subdenom))
	for _, reserved := range []string{
		appparams.BaseDenom,
		appparams.DisplayDenom,
		appparams.TokenSymbol,
		appparams.TokenName,
	} {
		if normalized == strings.ToLower(reserved) {
			return true
		}
	}
	return false
}
