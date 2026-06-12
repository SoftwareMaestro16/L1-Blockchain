package keeperwiring

import (
	"context"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"

	aetraeconomicskeeper "github.com/sovereign-l1/l1/x/aetra-economics/keeper"
	aetraeconomicstypes "github.com/sovereign-l1/l1/x/aetra-economics/types"
	aetrastakingpolicykeeper "github.com/sovereign-l1/l1/x/aetra-staking-policy/keeper"
	aetrastakingpolicytypes "github.com/sovereign-l1/l1/x/aetra-staking-policy/types"
	aetravalidatorscorekeeper "github.com/sovereign-l1/l1/x/aetra-validator-score/keeper"
	aetravalidatorscoretypes "github.com/sovereign-l1/l1/x/aetra-validator-score/types"
	burnkeeper "github.com/sovereign-l1/l1/x/burn/keeper"
	burntypes "github.com/sovereign-l1/l1/x/burn/types"
	delegatorprotectionkeeper "github.com/sovereign-l1/l1/x/delegator-protection/keeper"
	delegatorprotectiontypes "github.com/sovereign-l1/l1/x/delegator-protection/types"
	dynamiccommissionkeeper "github.com/sovereign-l1/l1/x/dynamic-commission/keeper"
	dynamiccommissiontypes "github.com/sovereign-l1/l1/x/dynamic-commission/types"
	emissionskeeper "github.com/sovereign-l1/l1/x/emissions/keeper"
	emissionstypes "github.com/sovereign-l1/l1/x/emissions/types"
	feecollectorkeeper "github.com/sovereign-l1/l1/x/fee-collector/keeper"
	feecollectortypes "github.com/sovereign-l1/l1/x/fee-collector/types"
	feeskeeper "github.com/sovereign-l1/l1/x/fees/keeper"
	feestypes "github.com/sovereign-l1/l1/x/fees/types"
	mintauthoritykeeper "github.com/sovereign-l1/l1/x/mint-authority/keeper"
	mintauthoritytypes "github.com/sovereign-l1/l1/x/mint-authority/types"
	performancekeeper "github.com/sovereign-l1/l1/x/performance/keeper"
	performancetypes "github.com/sovereign-l1/l1/x/performance/types"
	reputationkeeper "github.com/sovereign-l1/l1/x/reputation/keeper"
	reputationtypes "github.com/sovereign-l1/l1/x/reputation/types"
	stakeconcentrationkeeper "github.com/sovereign-l1/l1/x/stake-concentration/keeper"
	stakeconcentrationtypes "github.com/sovereign-l1/l1/x/stake-concentration/types"
	treasurykeeper "github.com/sovereign-l1/l1/x/treasury/keeper"
	treasurytypes "github.com/sovereign-l1/l1/x/treasury/types"
)

// reputationReaderAdapter wraps the reputation module keeper as a feestypes.ReputationReader.
type reputationReaderAdapter struct {
	Keeper reputationkeeper.Keeper
}

func (a reputationReaderAdapter) GetIdentityReputationScore(ctx context.Context, addr sdk.AccAddress) (uint32, bool, error) {
	return a.Keeper.GetIdentityReputationScore(ctx, addr)
}

// validatorReputationAdapter wraps the reputation keeper as a dynamiccommissiontypes.ReputationKeeper.
type validatorReputationAdapter struct {
	Keeper reputationkeeper.Keeper
}

func (a validatorReputationAdapter) GetValidatorTotalScore(ctx context.Context, addr string) (uint32, bool, error) {
	vs, err := a.Keeper.GetValidatorReputation(ctx, addr)
	if err != nil {
		return 0, false, err
	}
	if vs == nil {
		return 0, false, nil
	}
	return vs.TotalScore, vs.IsJailed || vs.IsSlashed, nil
}

type NativeKeeperDeps struct {
	AppCodec	codec.Codec
	Keys		map[string]*storetypes.KVStoreKey
	AccountKeeper	authkeeper.AccountKeeper
	BankKeeper	bankkeeper.BaseKeeper
	DistrKeeper	distrkeeper.Keeper
	GovAuthority	string
}

type NativeKeepers struct {
	BurnKeeper			burnkeeper.Keeper
	TreasuryKeeper			treasurykeeper.Keeper
	EmissionsKeeper			emissionskeeper.Keeper
	MintAuthorityKeeper		mintauthoritykeeper.Keeper
	DelegatorProtectionKeeper	delegatorprotectionkeeper.Keeper
	ReputationKeeper		reputationkeeper.Keeper
	PerformanceKeeper		performancekeeper.Keeper
	DynamicCommissionKeeper		dynamiccommissionkeeper.Keeper
	StakeConcentrationKeeper	stakeconcentrationkeeper.Keeper
	FeeCollectorKeeper		feecollectorkeeper.Keeper
	FeesKeeper			feeskeeper.Keeper
	AetraStakingPolicyKeeper	aetrastakingpolicykeeper.Keeper
	AetraEconomicsKeeper		aetraeconomicskeeper.Keeper
	AetraValidatorScoreKeeper	aetravalidatorscorekeeper.Keeper
}

func NewNativeKeepers(deps NativeKeeperDeps) NativeKeepers {
	repKeeper := reputationkeeper.NewKeeper(
		runtime.NewKVStoreService(deps.Keys[reputationtypes.StoreKey]),
		deps.GovAuthority,
	)
	fcKeeper := feecollectorkeeper.NewKeeper(
		deps.AppCodec,
		runtime.NewKVStoreService(deps.Keys[feecollectortypes.StoreKey]),
		deps.AccountKeeper,
		deps.BankKeeper,
		deps.GovAuthority,
	)
	return NativeKeepers{
		BurnKeeper: burnkeeper.NewKeeper(
			deps.AppCodec,
			runtime.NewKVStoreService(deps.Keys[burntypes.StoreKey]),
			deps.BankKeeper,
			deps.GovAuthority,
		),
		TreasuryKeeper: treasurykeeper.NewKeeper(
			deps.AppCodec,
			runtime.NewKVStoreService(deps.Keys[treasurytypes.StoreKey]),
			deps.AccountKeeper,
			deps.BankKeeper,
			deps.GovAuthority,
		),
		EmissionsKeeper: emissionskeeper.NewKeeper(
			deps.AppCodec,
			runtime.NewKVStoreService(deps.Keys[emissionstypes.StoreKey]),
			deps.GovAuthority,
		),
		MintAuthorityKeeper: mintauthoritykeeper.NewKeeper(
			runtime.NewKVStoreService(deps.Keys[mintauthoritytypes.StoreKey]),
			deps.BankKeeper,
			deps.GovAuthority,
		),
		DelegatorProtectionKeeper: delegatorprotectionkeeper.NewKeeper(
			runtime.NewKVStoreService(deps.Keys[delegatorprotectiontypes.StoreKey]),
			deps.GovAuthority,
		),
		ReputationKeeper:	repKeeper,
		PerformanceKeeper: performancekeeper.NewKeeper(
			runtime.NewKVStoreService(deps.Keys[performancetypes.StoreKey]),
			deps.GovAuthority,
		),
		DynamicCommissionKeeper: dynamiccommissionkeeper.NewKeeper(
			deps.AppCodec,
			runtime.NewKVStoreService(deps.Keys[dynamiccommissiontypes.StoreKey]),
			deps.GovAuthority,
		).WithReputationKeeper(validatorReputationAdapter{Keeper: repKeeper}),
		StakeConcentrationKeeper: stakeconcentrationkeeper.NewKeeper(
			deps.AppCodec,
			runtime.NewKVStoreService(deps.Keys[stakeconcentrationtypes.StoreKey]),
			deps.GovAuthority,
		),
		FeeCollectorKeeper:	fcKeeper,
		FeesKeeper: feeskeeper.NewKeeper(
			deps.AppCodec,
			runtime.NewKVStoreService(deps.Keys[feestypes.StoreKey]),
			deps.AccountKeeper,
			deps.BankKeeper,
			deps.DistrKeeper,
			deps.GovAuthority,
		).WithReputationReader(reputationReaderAdapter{Keeper: repKeeper}).WithFeeCollector(fcKeeper),
		AetraStakingPolicyKeeper:	aetrastakingpolicykeeper.NewPersistentKeeper(runtime.NewKVStoreService(deps.Keys[aetrastakingpolicytypes.StoreKey]), deps.GovAuthority),
		AetraEconomicsKeeper:		aetraeconomicskeeper.NewPersistentKeeper(runtime.NewKVStoreService(deps.Keys[aetraeconomicstypes.StoreKey]), deps.GovAuthority),
		AetraValidatorScoreKeeper:	aetravalidatorscorekeeper.NewPersistentKeeper(runtime.NewKVStoreService(deps.Keys[aetravalidatorscoretypes.StoreKey]), deps.GovAuthority),
	}
}
