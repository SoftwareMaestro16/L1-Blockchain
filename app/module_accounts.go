package app

import (
	"fmt"
	"maps"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	protocolpooltypes "github.com/cosmos/cosmos-sdk/x/protocolpool/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	aetherisaddress "github.com/sovereign-l1/l1/app/addressing"
	burntypes "github.com/sovereign-l1/l1/x/burn/types"
	configtypes "github.com/sovereign-l1/l1/x/config/types"
	delegatorprotectiontypes "github.com/sovereign-l1/l1/x/delegator-protection/types"
	dextypes "github.com/sovereign-l1/l1/x/dex/types"
	feecollectortypes "github.com/sovereign-l1/l1/x/fee-collector/types"
	feestypes "github.com/sovereign-l1/l1/x/fees/types"
	mintauthoritytypes "github.com/sovereign-l1/l1/x/mint-authority/types"
	storagerenttypes "github.com/sovereign-l1/l1/x/storage-rent/types"
	systemregistrytypes "github.com/sovereign-l1/l1/x/system-registry/types"
	tokenfactorytypes "github.com/sovereign-l1/l1/x/tokenfactory/types"
	validatorelectiontypes "github.com/sovereign-l1/l1/x/validator-election/types"
	validatorinsurancetypes "github.com/sovereign-l1/l1/x/validator-insurance/types"
)

var maccPerms = map[string][]string{
	authtypes.FeeCollectorName:                     nil,
	distrtypes.ModuleName:                          nil,
	minttypes.ModuleName:                           {authtypes.Minter},
	stakingtypes.BondedPoolName:                    {authtypes.Burner, authtypes.Staking},
	stakingtypes.NotBondedPoolName:                 {authtypes.Burner, authtypes.Staking},
	govtypes.ModuleName:                            {authtypes.Burner},
	protocolpooltypes.ModuleName:                   nil,
	protocolpooltypes.ProtocolPoolEscrowAccount:    nil,
	tokenfactorytypes.ModuleName:                   {authtypes.Minter, authtypes.Burner},
	dextypes.ModuleName:                            {authtypes.Minter, authtypes.Burner},
	burntypes.ModuleName:                           {authtypes.Burner},
	feecollectortypes.CollectorModuleName:          {authtypes.Burner},
	feecollectortypes.TreasuryModuleName:           nil,
	feecollectortypes.ProtectionModuleName:         nil,
	feecollectortypes.ValidatorInsuranceModuleName: nil,
	feecollectortypes.EcosystemGrantsModuleName:    nil,
	feecollectortypes.StorageRentReserveModuleName: nil,
	feecollectortypes.BurnModuleName:               nil,
	feecollectortypes.ReporterRewardsModuleName:    nil,
	mintauthoritytypes.ModuleName:                  {authtypes.Minter},
	storagerenttypes.ModuleName:                    nil,
	delegatorprotectiontypes.ModuleName:            nil,
	validatorinsurancetypes.ModuleName:             nil,
	configtypes.ModuleName:                         nil,
	systemregistrytypes.ModuleName:                 nil,
	validatorelectiontypes.ModuleName:              nil,
	feestypes.ModuleName:                           nil,
}

func GetMaccPerms() map[string][]string {
	return maps.Clone(maccPerms)
}

func BlockedAddresses() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range GetMaccPerms() {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}
	for _, address := range aetherisaddress.AllSystemAddresses() {
		bz, err := aetherisaddress.Parse(address.Raw)
		if err != nil {
			panic(fmt.Errorf("invalid reserved system address %s: %w", address.Name, err))
		}
		key := sdk.AccAddress(bz).String()
		if address.CanReceiveUserFunds {
			delete(modAccAddrs, key)
			continue
		}
		modAccAddrs[key] = true
	}

	delete(modAccAddrs, authtypes.NewModuleAddress(govtypes.ModuleName).String())

	return modAccAddrs
}
