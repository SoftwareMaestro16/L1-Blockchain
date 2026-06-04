package app

import (
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	protocolpooltypes "github.com/cosmos/cosmos-sdk/x/protocolpool/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	dextypes "github.com/sovereign-l1/l1/x/dex/types"
	feestypes "github.com/sovereign-l1/l1/x/fees/types"
	tokenfactorytypes "github.com/sovereign-l1/l1/x/tokenfactory/types"
)

var moduleAccountPerms = map[string][]string{
	authtypes.FeeCollectorName:                  nil,
	distrtypes.ModuleName:                       nil,
	minttypes.ModuleName:                        {authtypes.Minter},
	stakingtypes.BondedPoolName:                 {authtypes.Burner, authtypes.Staking},
	stakingtypes.NotBondedPoolName:              {authtypes.Burner, authtypes.Staking},
	govtypes.ModuleName:                         {authtypes.Burner},
	protocolpooltypes.ModuleName:                nil,
	protocolpooltypes.ProtocolPoolEscrowAccount: nil,
	tokenfactorytypes.ModuleName:                {authtypes.Minter, authtypes.Burner},
	dextypes.ModuleName:                         {authtypes.Minter, authtypes.Burner},
	feestypes.ModuleName:                        nil,
}

func GetMaccPerms() map[string][]string {
	perms := make(map[string][]string, len(moduleAccountPerms))
	for moduleName, modulePerms := range moduleAccountPerms {
		if modulePerms == nil {
			perms[moduleName] = nil
			continue
		}
		perms[moduleName] = append([]string(nil), modulePerms...)
	}
	return perms
}

func BlockedAddresses() map[string]bool {
	modAccAddrs := make(map[string]bool, len(moduleAccountPerms))
	for acc := range moduleAccountPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	delete(modAccAddrs, authtypes.NewModuleAddress(govtypes.ModuleName).String())

	return modAccAddrs
}
