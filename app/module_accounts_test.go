package app

import (
	"testing"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	protocolpooltypes "github.com/cosmos/cosmos-sdk/x/protocolpool/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	burntypes "github.com/sovereign-l1/l1/x/burn/types"
	configtypes "github.com/sovereign-l1/l1/x/config/types"
	delegatorprotectiontypes "github.com/sovereign-l1/l1/x/delegator-protection/types"
	feecollectortypes "github.com/sovereign-l1/l1/x/fee-collector/types"
	feestypes "github.com/sovereign-l1/l1/x/fees/types"
	mintauthoritytypes "github.com/sovereign-l1/l1/x/mint-authority/types"
	storagerenttypes "github.com/sovereign-l1/l1/x/storage-rent/types"
	systemregistrytypes "github.com/sovereign-l1/l1/x/system-registry/types"
	validatorelectiontypes "github.com/sovereign-l1/l1/x/validator-election/types"
	validatorinsurancetypes "github.com/sovereign-l1/l1/x/validator-insurance/types"
)

func TestPrototypeModuleAccountPermissionsAreNarrow(t *testing.T) {
	expected := map[string][]string{
		authtypes.FeeCollectorName:			nil,
		distrtypes.ModuleName:				nil,
		minttypes.ModuleName:				{authtypes.Minter},
		stakingtypes.BondedPoolName:			{authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName:			{authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:				{authtypes.Burner},
		protocolpooltypes.ModuleName:			nil,
		protocolpooltypes.ProtocolPoolEscrowAccount:	nil,
		burntypes.ModuleName:				{authtypes.Burner},
		feecollectortypes.CollectorModuleName:		{authtypes.Burner},
		feecollectortypes.TreasuryModuleName:		nil,
		feecollectortypes.ProtectionModuleName:		nil,
		feecollectortypes.ValidatorInsuranceModuleName:	nil,
		feecollectortypes.EcosystemGrantsModuleName:	nil,
		feecollectortypes.StorageRentReserveModuleName:	nil,
		feecollectortypes.BurnModuleName:		nil,
		feecollectortypes.ReporterRewardsModuleName:	nil,
		mintauthoritytypes.ModuleName:			{authtypes.Minter},
		storagerenttypes.ModuleName:			nil,
		delegatorprotectiontypes.ModuleName:		nil,
		validatorinsurancetypes.ModuleName:		nil,
		configtypes.ModuleName:				nil,
		systemregistrytypes.ModuleName:			nil,
		validatorelectiontypes.ModuleName:		nil,
		feestypes.ModuleName:				nil,
	}
	require.Equal(t, expected, GetMaccPerms())

	blocked := BlockedAddresses()
	for moduleName := range expected {
		addr := authtypes.NewModuleAddress(moduleName).String()
		if moduleName == govtypes.ModuleName {
			require.False(t, blocked[addr])
			continue
		}
		require.True(t, blocked[addr], moduleName)
	}
}
