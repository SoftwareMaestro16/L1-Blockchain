package types

import (
	"bytes"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestGenesisRejectsDuplicateDenom(t *testing.T) {
	admin := sdk.AccAddress(bytes.Repeat([]byte{1}, 20)).String()
	denom := "factory/" + admin + "/foo"
	gs := GenesisState{Denoms: []DenomAuthorityMetadata{
		{Denom: denom, Admin: admin},
		{Denom: denom, Admin: admin},
	}}
	if err := gs.Validate(); err == nil {
		t.Fatal("expected duplicate denom to fail")
	}
}

func TestGenesisRejectsInvalidDenomAuthorityMetadata(t *testing.T) {
	admin := sdk.AccAddress(bytes.Repeat([]byte{1}, 20)).String()
	tests := map[string]DenomAuthorityMetadata{
		"invalid admin": {Denom: "factory/" + admin + "/foo", Admin: "not-an-address"},
		"invalid denom": {Denom: "factory/" + admin + "/!", Admin: admin},
		"wrong prefix":  {Denom: "other/" + admin + "/foo", Admin: admin},
	}

	for name, meta := range tests {
		t.Run(name, func(t *testing.T) {
			gs := GenesisState{Denoms: []DenomAuthorityMetadata{meta}}
			if err := gs.Validate(); err == nil {
				t.Fatal("expected invalid denom authority metadata")
			}
		})
	}
}
