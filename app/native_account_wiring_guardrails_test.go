package app

import (
	"testing"

	"github.com/stretchr/testify/require"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	burntypes "github.com/sovereign-l1/l1/x/burn/types"
	feestypes "github.com/sovereign-l1/l1/x/fees/types"
	nativeaccounttypes "github.com/sovereign-l1/l1/x/native-account/types"
	nominatorpooltypes "github.com/sovereign-l1/l1/x/nominator-pool/types"
	reputationtypes "github.com/sovereign-l1/l1/x/reputation/types"
	singlenominatorpooltypes "github.com/sovereign-l1/l1/x/single-nominator-pool/types"
	stakeconcentrationtypes "github.com/sovereign-l1/l1/x/stake-concentration/types"
	storagerenttypes "github.com/sovereign-l1/l1/x/storage-rent/types"
	treasurytypes "github.com/sovereign-l1/l1/x/treasury/types"
	validatorelectiontypes "github.com/sovereign-l1/l1/x/validator-election/types"
	validatorinsurancetypes "github.com/sovereign-l1/l1/x/validator-insurance/types"
	validatorregistrytypes "github.com/sovereign-l1/l1/x/validator-registry/types"
)

func TestNativeAccountGuardrailsKeepSDKAndCustomModuleWiring(t *testing.T) {
	app, _ := setup(true, 5)

	requiredSDKModules := []string{
		authtypes.ModuleName,
		banktypes.ModuleName,
		stakingtypes.ModuleName,
		slashingtypes.ModuleName,
		distrtypes.ModuleName,
	}
	requiredCustomModules := []string{
		storagerenttypes.ModuleName,
		nominatorpooltypes.ModuleName,
		singlenominatorpooltypes.ModuleName,
		validatorelectiontypes.ModuleName,
		validatorinsurancetypes.ModuleName,
		validatorregistrytypes.ModuleName,
		stakeconcentrationtypes.ModuleName,
		feestypes.ModuleName,
		burntypes.ModuleName,
		treasurytypes.ModuleName,
		reputationtypes.ModuleName,
	}

	initOrder := aetherCoreInitGenesisOrder()
	exportOrder := aetherCoreExportGenesisOrder()
	for _, moduleName := range append(requiredSDKModules, requiredCustomModules...) {
		require.Contains(t, app.ModuleManager.Modules, moduleName)
		require.Contains(t, initOrder, moduleName)
		require.Contains(t, exportOrder, moduleName)
	}
	require.NoError(t, app.ValidateAetraCoreWiringGate())
}

func TestNativeAccountGuardrailsDoNotRegisterNativeTokenNFTOrDEXModules(t *testing.T) {
	app, _ := setup(true, 5)

	moduleNames := make([]string, 0, len(app.ModuleManager.Modules))
	for moduleName := range app.ModuleManager.Modules {
		moduleNames = append(moduleNames, moduleName)
	}

	require.NoError(t, nativeaccounttypes.ValidateAssetRoutes(nativeaccounttypes.DefaultAssetRoutes()))
	require.NoError(t, nativeaccounttypes.ValidateNoNativeAssetModules(moduleNames))
	for _, denied := range nativeaccounttypes.NativeAssetModuleDenylist() {
		require.NotContains(t, app.ModuleManager.Modules, denied)
	}
}
