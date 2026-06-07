package genesisvalidation

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
	appparams "github.com/sovereign-l1/l1/app/params"
	feestypes "github.com/sovereign-l1/l1/x/fees/types"
)

type State map[string]json.RawMessage

func ValidateAetraGenesis(cdc codec.JSONCodec, genesis State, bondDenom string) error {
	if err := ValidateAuthGenesis(cdc, genesis); err != nil {
		return err
	}
	if err := ValidateBankGenesis(cdc, genesis); err != nil {
		return err
	}
	if err := ValidateStakingGenesis(cdc, genesis, bondDenom); err != nil {
		return err
	}
	if err := ValidateMintGenesis(cdc, genesis); err != nil {
		return err
	}
	return ValidateFeeGenesis(cdc, genesis)
}

func ValidateAuthGenesis(cdc codec.JSONCodec, genesis State) error {
	var authGenesis authtypes.GenesisState
	if genesis[authtypes.ModuleName] == nil {
		return fmt.Errorf("missing %s genesis state", authtypes.ModuleName)
	}
	if err := cdc.UnmarshalJSON(genesis[authtypes.ModuleName], &authGenesis); err != nil {
		return fmt.Errorf("invalid %s genesis state: %w", authtypes.ModuleName, err)
	}
	accounts, err := authtypes.UnpackAccounts(authGenesis.Accounts)
	if err != nil {
		return err
	}
	seenAccounts := make(map[string]struct{}, len(accounts))
	for _, account := range accounts {
		addr := account.GetAddress()
		addrText := addr.String()
		if _, found := seenAccounts[addrText]; found {
			return fmt.Errorf("duplicate auth genesis account: %s", aetraaddress.FormatAccAddress(addr))
		}
		seenAccounts[addrText] = struct{}{}
		if aetraaddress.IsZeroAccAddress(addr) {
			return fmt.Errorf("auth genesis account %s must not be zero address", aetraaddress.ZeroRawAddress)
		}
	}
	return nil
}

func ValidateBankGenesis(cdc codec.JSONCodec, genesis State) error {
	var bankGenesis banktypes.GenesisState
	if genesis[banktypes.ModuleName] == nil {
		return fmt.Errorf("missing %s genesis state", banktypes.ModuleName)
	}
	if err := cdc.UnmarshalJSON(genesis[banktypes.ModuleName], &bankGenesis); err != nil {
		return fmt.Errorf("invalid %s genesis state: %w", banktypes.ModuleName, err)
	}
	if err := bankGenesis.Validate(); err != nil {
		return err
	}
	for _, balance := range bankGenesis.Balances {
		addr, err := aetraaddress.ParseAccAddress(balance.Address)
		if err != nil {
			return fmt.Errorf("invalid bank balance address %s: %w", balance.Address, err)
		}
		if aetraaddress.IsZeroAccAddress(addr) {
			return fmt.Errorf("bank balance address %s must not be zero address", aetraaddress.ZeroRawAddress)
		}
	}
	return nil
}

func ValidateStakingGenesis(cdc codec.JSONCodec, genesis State, bondDenom string) error {
	var stakingGenesis stakingtypes.GenesisState
	if genesis[stakingtypes.ModuleName] == nil {
		return fmt.Errorf("missing %s genesis state", stakingtypes.ModuleName)
	}
	if err := cdc.UnmarshalJSON(genesis[stakingtypes.ModuleName], &stakingGenesis); err != nil {
		return fmt.Errorf("invalid %s genesis state: %w", stakingtypes.ModuleName, err)
	}
	if stakingGenesis.Params.BondDenom != bondDenom {
		return fmt.Errorf("invalid staking denom: expected %s, got %s", bondDenom, stakingGenesis.Params.BondDenom)
	}
	return nil
}

func ValidateMintGenesis(cdc codec.JSONCodec, genesis State) error {
	var mintGenesis minttypes.GenesisState
	if genesis[minttypes.ModuleName] == nil {
		return fmt.Errorf("missing %s genesis state", minttypes.ModuleName)
	}
	if err := cdc.UnmarshalJSON(genesis[minttypes.ModuleName], &mintGenesis); err != nil {
		return fmt.Errorf("invalid %s genesis state: %w", minttypes.ModuleName, err)
	}
	if err := minttypes.ValidateGenesis(mintGenesis); err != nil {
		return fmt.Errorf("invalid %s genesis state: %w", minttypes.ModuleName, err)
	}
	expected := appparams.AetraMintParams()
	if mintGenesis.Params.MintDenom != expected.MintDenom {
		return fmt.Errorf("invalid mint denom: expected %s, got %s", expected.MintDenom, mintGenesis.Params.MintDenom)
	}
	if !mintGenesis.Params.InflationRateChange.Equal(expected.InflationRateChange) {
		return fmt.Errorf("invalid mint inflation rate change: expected %s, got %s", expected.InflationRateChange, mintGenesis.Params.InflationRateChange)
	}
	if !mintGenesis.Params.InflationMin.Equal(expected.InflationMin) {
		return fmt.Errorf("invalid mint min inflation: expected %s, got %s", expected.InflationMin, mintGenesis.Params.InflationMin)
	}
	if !mintGenesis.Params.InflationMax.Equal(expected.InflationMax) {
		return fmt.Errorf("invalid mint max inflation: expected %s, got %s", expected.InflationMax, mintGenesis.Params.InflationMax)
	}
	if !mintGenesis.Params.GoalBonded.Equal(expected.GoalBonded) {
		return fmt.Errorf("invalid mint goal bonded: expected %s, got %s", expected.GoalBonded, mintGenesis.Params.GoalBonded)
	}
	if !mintGenesis.Params.MaxSupply.Equal(expected.MaxSupply) {
		return fmt.Errorf("invalid mint max supply: expected %s, got %s", expected.MaxSupply, mintGenesis.Params.MaxSupply)
	}
	if mintGenesis.Minter.Inflation.LT(expected.InflationMin) || mintGenesis.Minter.Inflation.GT(expected.InflationMax) {
		return fmt.Errorf("invalid mint current inflation: expected within %s..%s, got %s", expected.InflationMin, expected.InflationMax, mintGenesis.Minter.Inflation)
	}
	return nil
}

func ValidateFeeGenesis(cdc codec.JSONCodec, genesis State) error {
	var feesGenesis feestypes.GenesisState
	if genesis[feestypes.ModuleName] == nil {
		return fmt.Errorf("missing %s genesis state", feestypes.ModuleName)
	}
	if err := cdc.UnmarshalJSON(genesis[feestypes.ModuleName], &feesGenesis); err != nil {
		return fmt.Errorf("invalid %s genesis state: %w", feestypes.ModuleName, err)
	}
	if err := feesGenesis.Validate(); err != nil {
		return err
	}
	if err := appparams.ValidateNativeFeeDenomsV1(feesGenesis.Params.AllowedFeeDenoms, feestypes.MaxAllowedFeeDenomsV1); err != nil {
		return err
	}
	return nil
}
