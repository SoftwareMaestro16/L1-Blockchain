package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestDefaultParamsValidate(t *testing.T) {
	params := DefaultParams()
	if len(params.AllowedFeeDenoms) != 1 || params.AllowedFeeDenoms[0] != "norb" {
		t.Fatalf("expected only norb as default fee denom: %v", params.AllowedFeeDenoms)
	}
	if params.MinFeeAmount != "1" {
		t.Fatalf("expected min fee amount 1, got %q", params.MinFeeAmount)
	}
	if params.FeeCollectorModule != FeeCollectorModuleName {
		t.Fatalf("expected fee collector %s, got %q", FeeCollectorModuleName, params.FeeCollectorModule)
	}
	if err := params.Validate(); err != nil {
		t.Fatalf("default params should validate: %v", err)
	}
}

func TestParamsRejectNonNativeFeeDenom(t *testing.T) {
	params := DefaultParams()
	params.AllowedFeeDenoms = []string{"uatom"}
	if err := params.Validate(); err == nil {
		t.Fatal("expected non-native fee denom to fail")
	}
}

func TestParamsRejectInvalidFeeSplitRatios(t *testing.T) {
	tests := map[string]func(*Params){
		"malformed validator ratio": func(params *Params) {
			params.ValidatorRewardsRatio = "not-a-decimal"
		},
		"malformed community ratio": func(params *Params) {
			params.CommunityPoolRatio = "not-a-decimal"
		},
		"negative ratio": func(params *Params) {
			params.ValidatorRewardsRatio = "-0.1"
			params.CommunityPoolRatio = "1.1"
		},
		"sum not one": func(params *Params) {
			params.ValidatorRewardsRatio = "0.80"
			params.CommunityPoolRatio = "0.10"
		},
		"ratio greater than one": func(params *Params) {
			params.ValidatorRewardsRatio = "1.01"
			params.CommunityPoolRatio = "-0.01"
		},
	}

	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			params := DefaultParams()
			mutate(&params)
			if err := params.Validate(); err == nil {
				t.Fatal("expected invalid fee split params to fail")
			}
		})
	}
}

func TestParamsRejectUnsafeProtocolFeeConfig(t *testing.T) {
	tests := map[string]func(*Params){
		"zero min fee": func(params *Params) {
			params.MinFeeAmount = "0"
		},
		"malformed min fee": func(params *Params) {
			params.MinFeeAmount = "not-an-int"
		},
		"unsafe fee collector": func(params *Params) {
			params.FeeCollectorModule = ModuleName
		},
		"unsafe validator target": func(params *Params) {
			params.ValidatorRewardsTarget = CommunityPoolTarget
		},
		"unsafe community target": func(params *Params) {
			params.CommunityPoolTarget = ValidatorRewardsTarget
		},
		"duplicate denom": func(params *Params) {
			params.AllowedFeeDenoms = []string{BondDenom, BondDenom}
		},
	}

	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			params := DefaultParams()
			mutate(&params)
			if err := params.Validate(); err == nil {
				t.Fatal("expected invalid protocol fee params to fail")
			}
		})
	}
}

func TestSplitFeesRoundingPreservesTotal(t *testing.T) {
	params := DefaultParams()
	validator, community, err := SplitFees(params, sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 1)))
	if err != nil {
		t.Fatalf("split fees should not fail: %v", err)
	}
	if !validator.Equal(sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 1))) {
		t.Fatalf("expected validator side to receive integer dust, got %s", validator)
	}
	if !community.Empty() {
		t.Fatalf("expected community side to truncate to zero, got %s", community)
	}
	if !validator.Add(community...).Equal(sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 1))) {
		t.Fatal("split must preserve total")
	}
}

func TestProtocolFeeStateValidateRejectsAccountingMismatch(t *testing.T) {
	state := DefaultProtocolFeeState()
	state.TotalCollected = sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 100))
	state.ValidatorRewards = sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 80))
	state.CommunityPool = sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 10))
	if err := state.Validate(); err == nil {
		t.Fatal("expected accounting mismatch to fail")
	}
}

func TestProtocolFeeStateValidateRejectsWrongDenom(t *testing.T) {
	state := DefaultProtocolFeeState()
	state.TotalCollected = sdk.NewCoins(sdk.NewInt64Coin("uatom", 100))
	state.ValidatorRewards = sdk.NewCoins(sdk.NewInt64Coin("uatom", 100))
	if err := state.Validate(); err == nil {
		t.Fatal("expected wrong accounting denom to fail")
	}
}

func TestGenesisValidateIncludesProtocolFeeState(t *testing.T) {
	gs := DefaultGenesisState()
	if err := gs.Validate(); err != nil {
		t.Fatalf("default genesis should validate: %v", err)
	}

	gs.ProtocolFeeState.TotalCollected = sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 10))
	if err := gs.Validate(); err == nil {
		t.Fatal("expected malformed protocol fee state to fail")
	}
}
