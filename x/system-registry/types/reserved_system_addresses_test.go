package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/app/addressing"
)

func TestReservedSystemRegistryGenesisValidates(t *testing.T) {
	params := DefaultParams()
	state := DefaultState().Normalize(params)

	require.NoError(t, state.Validate(params))
	require.Len(t, state.Entities, len(addressing.AllSystemAddresses()))
}

func TestReservedSystemRegistryGenesisRejectsMissingAETMint(t *testing.T) {
	params := DefaultParams()
	state := DefaultState().Normalize(params)
	for i := range state.Entities {
		if state.Entities[i].Name == "AETMint" {
			state.Entities[i].Name = ""
			break
		}
	}

	require.ErrorContains(t, state.Validate(params), `required system entity "AETMint" is missing`)
}

func TestReservedSystemRegistryGenesisRejectsWrongAETElectorRaw(t *testing.T) {
	params := DefaultParams()
	state := DefaultState().Normalize(params)
	for i := range state.Entities {
		if state.Entities[i].Name == "AETElector" {
			state.Entities[i].RawAddress = "4:01041041041041041041041041041041041041041041041041041042c4093391"
			break
		}
	}

	require.ErrorContains(t, state.Validate(params), `entity "AETElector" raw mismatch`)
}

func TestReservedSystemRegistryGenesisRejectsDuplicateRaw(t *testing.T) {
	params := DefaultParams()
	state := DefaultState().Normalize(params)
	mint, found := addressing.SystemAddressByName("AETMint")
	require.True(t, found)
	state.Entities = append(state.Entities, SystemEntity{
		Name:			"AETDuplicate",
		ModuleName:		"duplicate-system-address",
		ModuleAccountAddress:	"4:0000000000000000000000001111111111111111111111111111111111111111",
		RawAddress:		mint.Raw,
		AuthorityAddress:	params.Authority,
		Status:			StatusActive,
		Version:		1,
	})

	require.ErrorContains(t, state.Validate(params), "duplicate reserved address bytes")
}

func TestReservedSystemRegistryGenesisRejectsZeroAddress(t *testing.T) {
	params := DefaultParams()
	state := DefaultState().Normalize(params)
	for i := range state.Entities {
		if state.Entities[i].Name == "AETMint" {
			state.Entities[i].RawAddress = addressing.ZeroRawAddress
			break
		}
	}

	require.ErrorContains(t, state.Validate(params), "must not be zero address")
}

func TestReservedSystemRegistryGenesisRejectsUserControlledReservedAddress(t *testing.T) {
	params := DefaultParams()
	state := DefaultState().Normalize(params)
	mint, found := addressing.SystemAddressByName("AETMint")
	require.True(t, found)
	state.UserControlledAccounts = append(state.UserControlledAccounts, mint.Raw)

	require.ErrorContains(t, state.Validate(params), "uses reserved system address")
}
