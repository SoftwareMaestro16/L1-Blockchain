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

func TestDefaultParamsValidate(t *testing.T) {
	params := DefaultParams()
	if params.MinSubdenomLength != DefaultMinSubdenomLength || params.MaxSubdenomLength != DefaultMaxSubdenomLength {
		t.Fatalf("unexpected default subdenom bounds: %+v", params)
	}
	if !params.DenomCreationEnabled || !params.MintingEnabled || !params.BurningEnabled {
		t.Fatal("default tokenfactory params should enable protocol operations")
	}
	if err := params.Validate(); err != nil {
		t.Fatalf("default params should validate: %v", err)
	}
}

func TestParamsRejectUnsafeSubdenomBounds(t *testing.T) {
	tests := map[string]func(*Params){
		"zero min": func(params *Params) {
			params.MinSubdenomLength = 0
		},
		"max below min": func(params *Params) {
			params.MinSubdenomLength = 10
			params.MaxSubdenomLength = 9
		},
		"max above cap": func(params *Params) {
			params.MaxSubdenomLength = DefaultMaxSubdenomLength + 1
		},
	}

	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			params := DefaultParams()
			mutate(&params)
			if err := params.Validate(); err == nil {
				t.Fatal("expected invalid tokenfactory params")
			}
		})
	}
}

func TestValidateSubdenomRejectsMalformedInput(t *testing.T) {
	tests := []string{"1gold", "go//ld", "go ld", ""}
	for _, subdenom := range tests {
		if err := ValidateSubdenom(subdenom, DefaultParams()); err == nil {
			t.Fatalf("expected malformed subdenom %q to fail", subdenom)
		}
	}
}
