package storekeys

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	epochstypes "github.com/cosmos/cosmos-sdk/x/epochs/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
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

func NewKVStoreKeys() map[string]*storetypes.KVStoreKey {
	return storetypes.NewKVStoreKeys(
		authtypes.StoreKey,
		banktypes.StoreKey,
		stakingtypes.StoreKey,
		minttypes.StoreKey,
		distrtypes.StoreKey,
		slashingtypes.StoreKey,
		govtypes.StoreKey,
		consensusparamtypes.StoreKey,
		upgradetypes.StoreKey,
		feegrant.StoreKey,
		evidencetypes.StoreKey,
		authzkeeper.StoreKey,
		epochstypes.StoreKey,
		protocolpooltypes.StoreKey,
		constitutiontypes.StoreKey,
		configtypes.StoreKey,
		configvotingtypes.StoreKey,
		systemregistrytypes.StoreKey,
		nativeevidencetypes.StoreKey,
		reportertypes.StoreKey,
		nominatorpooltypes.StoreKey,
		singlenominatorpooltypes.StoreKey,
		validatorelectiontypes.StoreKey,
		validatorinsurancetypes.StoreKey,
		validatorregistrytypes.StoreKey,
		aetrastakingpolicytypes.StoreKey,
		aetraeconomicstypes.StoreKey,
		aetravalidatorscoretypes.StoreKey,
		burntypes.StoreKey,
		treasurytypes.StoreKey,
		emissionstypes.StoreKey,
		mintauthoritytypes.StoreKey,
		delegatorprotectiontypes.StoreKey,
		reputationtypes.StoreKey,
		performancetypes.StoreKey,
		dynamiccommissiontypes.StoreKey,
		stakeconcentrationtypes.StoreKey,
		feecollectortypes.StoreKey,
		feestypes.StoreKey,
		aetracoretypes.StoreKey,
		loadtypes.StoreKey,
		routingtypes.StoreKey,
		zonestypes.StoreKey,
		meshtypes.StoreKey,
		networkingtypes.StoreKey,
		nativeaccounttypes.StoreKey,
		paymentstypes.StoreKey,
		schedulertypes.StoreKey,
		avmschedulertypes.StoreKey,
		actorregistrytypes.StoreKey,
		contractstypes.StoreKey,
		storagerenttypes.StoreKey,
		identityroottypes.StoreKey,
		bridgehubtypes.StoreKey,
		crosschainregistrytypes.StoreKey,
		shardingcoordinatortypes.StoreKey,
	)
}
