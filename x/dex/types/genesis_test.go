package types

import "testing"

func validGenesisState() GenesisState {
	return GenesisState{
		NextPoolId: 2,
		Pools: []Pool{
			{
				Id:          1,
				Denom0:      "norb",
				Denom1:      "uatom",
				Reserve0:    "100",
				Reserve1:    "200",
				TotalShares: "100",
				LpDenom:     "lp/1",
			},
		},
	}
}

func TestGenesisValidateRejectsInvalidPoolState(t *testing.T) {
	tests := map[string]func(*GenesisState){
		"invalid reserve": func(gs *GenesisState) {
			gs.Pools[0].Reserve0 = "not-int"
		},
		"zero reserve": func(gs *GenesisState) {
			gs.Pools[0].Reserve0 = "0"
		},
		"zero shares": func(gs *GenesisState) {
			gs.Pools[0].TotalShares = "0"
		},
		"lp denom mismatch": func(gs *GenesisState) {
			gs.Pools[0].LpDenom = "lp/2"
		},
		"next id not greater than pool id": func(gs *GenesisState) {
			gs.NextPoolId = 1
		},
		"non canonical denoms": func(gs *GenesisState) {
			gs.Pools[0].Denom0 = "uatom"
			gs.Pools[0].Denom1 = "norb"
		},
		"duplicate pair": func(gs *GenesisState) {
			gs.NextPoolId = 3
			gs.Pools = append(gs.Pools, Pool{
				Id:          2,
				Denom0:      "norb",
				Denom1:      "uatom",
				Reserve0:    "50",
				Reserve1:    "50",
				TotalShares: "50",
				LpDenom:     "lp/2",
			})
		},
	}

	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			gs := validGenesisState()
			mutate(&gs)
			if err := gs.Validate(); err == nil {
				t.Fatalf("expected invalid genesis state")
			}
		})
	}
}

func TestGenesisValidateAcceptsValidPoolState(t *testing.T) {
	if err := validGenesisState().Validate(); err != nil {
		t.Fatalf("expected valid genesis state: %v", err)
	}
}

func TestDefaultParamsValidate(t *testing.T) {
	params := DefaultParams()
	if params.SwapFeeBps != DefaultSwapFeeBps {
		t.Fatalf("unexpected default swap fee: %d", params.SwapFeeBps)
	}
	if !params.PoolCreationEnabled || !params.SwapsEnabled || !params.LiquidityEnabled {
		t.Fatal("default DEX params should enable protocol operations")
	}
	if err := params.Validate(); err != nil {
		t.Fatalf("default params should validate: %v", err)
	}
}

func TestParamsRejectUnsafeValues(t *testing.T) {
	tests := map[string]func(*Params){
		"fee above cap": func(params *Params) {
			params.SwapFeeBps = params.MaxSwapFeeBps + 1
		},
		"cap above immutable bound": func(params *Params) {
			params.MaxSwapFeeBps = DefaultMaxSwapFeeBps + 1
		},
		"zero min initial liquidity": func(params *Params) {
			params.MinInitialLiquidity = "0"
		},
		"malformed min initial liquidity": func(params *Params) {
			params.MinInitialLiquidity = "not-int"
		},
	}

	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			params := DefaultParams()
			mutate(&params)
			if err := params.Validate(); err == nil {
				t.Fatal("expected invalid DEX params")
			}
		})
	}
}
