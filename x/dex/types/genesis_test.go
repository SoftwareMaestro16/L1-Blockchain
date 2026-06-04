package types

import (
	"testing"

	appparams "github.com/sovereign-l1/l1/app/params"
)

func validGenesisState() GenesisState {
	return GenesisState{
		NextPoolId: 2,
		Pools: []Pool{
			{
				Id:          1,
				Denom0:      appparams.BaseDenom,
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
			gs.Pools[0].Denom1 = appparams.BaseDenom
		},
		"duplicate pair": func(gs *GenesisState) {
			gs.NextPoolId = 3
			gs.Pools = append(gs.Pools, Pool{
				Id:          2,
				Denom0:      appparams.BaseDenom,
				Denom1:      "uatom",
				Reserve0:    "50",
				Reserve1:    "50",
				TotalShares: "50",
				LpDenom:     "lp/2",
			})
		},
		"duplicate pool id": func(gs *GenesisState) {
			gs.NextPoolId = 3
			gs.Pools = append(gs.Pools, Pool{
				Id:          1,
				Denom0:      "uosmo",
				Denom1:      "uatom",
				Reserve0:    "50",
				Reserve1:    "50",
				TotalShares: "50",
				LpDenom:     "lp/1",
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
