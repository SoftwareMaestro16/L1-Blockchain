package aetracore

import (
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	epochstypes "github.com/cosmos/cosmos-sdk/x/epochs/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	protocolpooltypes "github.com/cosmos/cosmos-sdk/x/protocolpool/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	actorregistrytypes "github.com/sovereign-l1/l1/x/actor-registry/types"
	aetraeconomicstypes "github.com/sovereign-l1/l1/x/aetra-economics/types"
	aetrastakingpolicytypes "github.com/sovereign-l1/l1/x/aetra-staking-policy/types"
	aetravalidatorscoretypes "github.com/sovereign-l1/l1/x/aetra-validator-score/types"
	aetracoretypes "github.com/sovereign-l1/l1/x/aetracore/types"
	avmschedulertypes "github.com/sovereign-l1/l1/x/avm-scheduler/types"
	bridgehubtypes "github.com/sovereign-l1/l1/x/bridge-hub/types"
	burntypes "github.com/sovereign-l1/l1/x/burn/types"
	configvotingtypes "github.com/sovereign-l1/l1/x/config-voting/types"
	configtypes "github.com/sovereign-l1/l1/x/config/types"
	constitutiontypes "github.com/sovereign-l1/l1/x/constitution/types"
	contractstypes "github.com/sovereign-l1/l1/x/contracts/types"
	crosschainregistrytypes "github.com/sovereign-l1/l1/x/cross-chain-registry/types"
	delegatorprotectiontypes "github.com/sovereign-l1/l1/x/delegator-protection/types"
	dynamiccommissiontypes "github.com/sovereign-l1/l1/x/dynamic-commission/types"
	emissionstypes "github.com/sovereign-l1/l1/x/emissions/types"
	nativeevidencetypes "github.com/sovereign-l1/l1/x/evidence/types"
	feecollectortypes "github.com/sovereign-l1/l1/x/fee-collector/types"
	feestypes "github.com/sovereign-l1/l1/x/fees/types"
	identityroottypes "github.com/sovereign-l1/l1/x/identity-root/types"
	loadtypes "github.com/sovereign-l1/l1/x/load/types"
	meshtypes "github.com/sovereign-l1/l1/x/mesh/types"
	mintauthoritytypes "github.com/sovereign-l1/l1/x/mint-authority/types"
	nativeaccounttypes "github.com/sovereign-l1/l1/x/native-account/types"
	networkingtypes "github.com/sovereign-l1/l1/x/networking/types"
	nominatorpooltypes "github.com/sovereign-l1/l1/x/nominator-pool/types"
	paymentstypes "github.com/sovereign-l1/l1/x/payments/types"
	performancetypes "github.com/sovereign-l1/l1/x/performance/types"
	reportertypes "github.com/sovereign-l1/l1/x/reporter/types"
	reputationtypes "github.com/sovereign-l1/l1/x/reputation/types"
	routingtypes "github.com/sovereign-l1/l1/x/routing/types"
	schedulertypes "github.com/sovereign-l1/l1/x/scheduler/types"
	shardingcoordinatortypes "github.com/sovereign-l1/l1/x/sharding-coordinator/types"
	singlenominatorpooltypes "github.com/sovereign-l1/l1/x/single-nominator-pool/types"
	stakeconcentrationtypes "github.com/sovereign-l1/l1/x/stake-concentration/types"
	storagerenttypes "github.com/sovereign-l1/l1/x/storage-rent/types"
	systemregistrytypes "github.com/sovereign-l1/l1/x/system-registry/types"
	treasurytypes "github.com/sovereign-l1/l1/x/treasury/types"
	validatorelectiontypes "github.com/sovereign-l1/l1/x/validator-election/types"
	validatorinsurancetypes "github.com/sovereign-l1/l1/x/validator-insurance/types"
	validatorregistrytypes "github.com/sovereign-l1/l1/x/validator-registry/types"
	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func PreBlockerOrder() []string {
	return []string{upgradetypes.ModuleName, authtypes.ModuleName}
}

func BeginBlockerOrder() []string {
	return []string{
		minttypes.ModuleName,
		distrtypes.ModuleName,
		protocolpooltypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		stakingtypes.ModuleName,
		genutiltypes.ModuleName,
		authz.ModuleName,
		epochstypes.ModuleName,
		constitutiontypes.ModuleName,
		systemregistrytypes.ModuleName,
		nativeevidencetypes.ModuleName,
		reportertypes.ModuleName,
		nominatorpooltypes.ModuleName,
		singlenominatorpooltypes.ModuleName,
		validatorelectiontypes.ModuleName,
		validatorinsurancetypes.ModuleName,
		validatorregistrytypes.ModuleName,
		aetrastakingpolicytypes.ModuleName,
		aetraeconomicstypes.ModuleName,
		aetravalidatorscoretypes.ModuleName,
		configtypes.ModuleName,
		configvotingtypes.ModuleName,
		aetracoretypes.ModuleName,
		loadtypes.ModuleName,
		routingtypes.ModuleName,
		zonestypes.ModuleName,
		meshtypes.ModuleName,
		networkingtypes.ModuleName,
		paymentstypes.ModuleName,
		schedulertypes.ModuleName,
		avmschedulertypes.ModuleName,
		actorregistrytypes.ModuleName,
		contractstypes.ModuleName,
		storagerenttypes.ModuleName,
		identityroottypes.ModuleName,
		bridgehubtypes.ModuleName,
		crosschainregistrytypes.ModuleName,
		shardingcoordinatortypes.ModuleName,
	}
}

func EndBlockerOrder() []string {
	return []string{
		banktypes.ModuleName,
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		genutiltypes.ModuleName,
		feegrant.ModuleName,
		protocolpooltypes.ModuleName,
		constitutiontypes.ModuleName,
		systemregistrytypes.ModuleName,
		nativeevidencetypes.ModuleName,
		reportertypes.ModuleName,
		nominatorpooltypes.ModuleName,
		singlenominatorpooltypes.ModuleName,
		validatorelectiontypes.ModuleName,
		validatorinsurancetypes.ModuleName,
		validatorregistrytypes.ModuleName,
		aetrastakingpolicytypes.ModuleName,
		aetraeconomicstypes.ModuleName,
		aetravalidatorscoretypes.ModuleName,
		configtypes.ModuleName,
		configvotingtypes.ModuleName,
		aetracoretypes.ModuleName,
		loadtypes.ModuleName,
		routingtypes.ModuleName,
		zonestypes.ModuleName,
		meshtypes.ModuleName,
		networkingtypes.ModuleName,
		paymentstypes.ModuleName,
		schedulertypes.ModuleName,
		avmschedulertypes.ModuleName,
		actorregistrytypes.ModuleName,
		contractstypes.ModuleName,
		storagerenttypes.ModuleName,
		identityroottypes.ModuleName,
		bridgehubtypes.ModuleName,
		crosschainregistrytypes.ModuleName,
		shardingcoordinatortypes.ModuleName,
	}
}

func InitGenesisOrder() []string {
	return []string{
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		slashingtypes.ModuleName,
		govtypes.ModuleName,
		minttypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		upgradetypes.ModuleName,
		vestingtypes.ModuleName,
		consensusparamtypes.ModuleName,
		epochstypes.ModuleName,
		protocolpooltypes.ModuleName,
		constitutiontypes.ModuleName,
		systemregistrytypes.ModuleName,
		nativeevidencetypes.ModuleName,
		reportertypes.ModuleName,
		nominatorpooltypes.ModuleName,
		singlenominatorpooltypes.ModuleName,
		validatorelectiontypes.ModuleName,
		validatorinsurancetypes.ModuleName,
		validatorregistrytypes.ModuleName,
		aetrastakingpolicytypes.ModuleName,
		aetraeconomicstypes.ModuleName,
		aetravalidatorscoretypes.ModuleName,
		configtypes.ModuleName,
		nativeaccounttypes.ModuleName,
		configvotingtypes.ModuleName,
		aetracoretypes.ModuleName,
		loadtypes.ModuleName,
		routingtypes.ModuleName,
		zonestypes.ModuleName,
		meshtypes.ModuleName,
		networkingtypes.ModuleName,
		paymentstypes.ModuleName,
		schedulertypes.ModuleName,
		avmschedulertypes.ModuleName,
		actorregistrytypes.ModuleName,
		contractstypes.ModuleName,
		storagerenttypes.ModuleName,
		identityroottypes.ModuleName,
		bridgehubtypes.ModuleName,
		crosschainregistrytypes.ModuleName,
		shardingcoordinatortypes.ModuleName,
		burntypes.ModuleName,
		treasurytypes.ModuleName,
		emissionstypes.ModuleName,
		mintauthoritytypes.ModuleName,
		delegatorprotectiontypes.ModuleName,
		reputationtypes.ModuleName,
		performancetypes.ModuleName,
		dynamiccommissiontypes.ModuleName,
		stakeconcentrationtypes.ModuleName,
		feecollectortypes.ModuleName,
		feestypes.ModuleName,
	}
}

func ExportGenesisOrder() []string {
	return []string{
		consensusparamtypes.ModuleName,
		authtypes.ModuleName,
		protocolpooltypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		slashingtypes.ModuleName,
		govtypes.ModuleName,
		minttypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		upgradetypes.ModuleName,
		vestingtypes.ModuleName,
		epochstypes.ModuleName,
		constitutiontypes.ModuleName,
		systemregistrytypes.ModuleName,
		nativeevidencetypes.ModuleName,
		reportertypes.ModuleName,
		nominatorpooltypes.ModuleName,
		singlenominatorpooltypes.ModuleName,
		validatorelectiontypes.ModuleName,
		validatorinsurancetypes.ModuleName,
		validatorregistrytypes.ModuleName,
		aetrastakingpolicytypes.ModuleName,
		aetraeconomicstypes.ModuleName,
		aetravalidatorscoretypes.ModuleName,
		configtypes.ModuleName,
		nativeaccounttypes.ModuleName,
		configvotingtypes.ModuleName,
		aetracoretypes.ModuleName,
		loadtypes.ModuleName,
		routingtypes.ModuleName,
		zonestypes.ModuleName,
		meshtypes.ModuleName,
		networkingtypes.ModuleName,
		paymentstypes.ModuleName,
		schedulertypes.ModuleName,
		avmschedulertypes.ModuleName,
		actorregistrytypes.ModuleName,
		contractstypes.ModuleName,
		storagerenttypes.ModuleName,
		identityroottypes.ModuleName,
		bridgehubtypes.ModuleName,
		crosschainregistrytypes.ModuleName,
		shardingcoordinatortypes.ModuleName,
		burntypes.ModuleName,
		treasurytypes.ModuleName,
		emissionstypes.ModuleName,
		mintauthoritytypes.ModuleName,
		delegatorprotectiontypes.ModuleName,
		reputationtypes.ModuleName,
		performancetypes.ModuleName,
		dynamiccommissiontypes.ModuleName,
		stakeconcentrationtypes.ModuleName,
		feecollectortypes.ModuleName,
		feestypes.ModuleName,
	}
}
