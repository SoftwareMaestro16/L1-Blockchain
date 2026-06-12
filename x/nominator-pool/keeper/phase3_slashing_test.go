package keeper

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/x/internal/kvtest"
	"github.com/sovereign-l1/l1/x/internal/prototype"
	"github.com/sovereign-l1/l1/x/nominator-pool/types"
)

func TestPhase34SlashingParamsAreGenesisGovernanceParams(t *testing.T) {
	k := NewKeeper()
	params := k.ExportGenesis().Params
	require.Equal(t, uint32(5), params.DowntimeSlashBps)
	require.Equal(t, uint32(500), params.DoubleSignSlashBps)
	require.True(t, params.DoubleSignTombstone)

	params.DowntimeSlashBps = 7
	updated, err := k.UpdateParams(types.MsgUpdateParams{
		Authority:	prototype.DefaultAuthority,
		Params:		params,
		Height:		2,
	})
	require.NoError(t, err)
	require.Equal(t, uint32(7), updated.DowntimeSlashBps)

	params.DoubleSignTombstone = false
	_, err = k.UpdateParams(types.MsgUpdateParams{
		Authority:	prototype.DefaultAuthority,
		Params:		params,
		Height:		3,
	})
	require.ErrorContains(t, err, "double-sign slash must tombstone")
}

func TestPhase34DowntimeSlashJailsValidatorAndReducesPoolExposure(t *testing.T) {
	k, pool, validator := phase34PoolWithValidatorAllocation(t, "phase34-downtime", "a1", types.DefaultMinPoolDeposit)

	events, err := k.ApplyValidatorSlash(types.MsgApplyValidatorSlash{
		Authority:		prototype.DefaultAuthority,
		ValidatorAddress:	validator,
		Fault:			types.SlashingFaultDowntime,
		Epoch:			1,
		Height:			10,
	})
	require.NoError(t, err)
	require.Len(t, events, 1)
	expectedLoss := types.DefaultMinPoolDeposit * uint64(types.DefaultParams().DowntimeSlashBps) / uint64(types.MaxBasisPoints)
	require.Equal(t, expectedLoss, events[0].SlashingLoss)
	require.Equal(t, types.StateValidatorStatusJailed, events[0].ValidatorStatus)
	require.False(t, events[0].Tombstoned)

	exported := k.ExportGenesis()
	require.Equal(t, types.StateValidatorStatusJailed, exported.State.Validators[0].Status)
	require.Equal(t, types.DefaultParams().DowntimeSlashBps, exported.State.Validators[0].SlashingRiskBps)
	require.Equal(t, types.DefaultMinPoolDeposit-expectedLoss, exported.State.Pools[0].TotalBondedStake)
	require.Equal(t, types.DefaultMinPoolDeposit-expectedLoss, exported.State.Pools[0].Allocations[0].Amount)
	require.Equal(t, types.DefaultMinPoolDeposit-expectedLoss, exported.State.PoolValidatorAllocations[0].ActiveStake)
	require.Zero(t, exported.State.PoolValidatorAllocations[0].TargetWeightBps)
	require.Equal(t, events[0].PoolSlashIndexAfter, exported.State.Pools[0].SlashIndex)
	require.Equal(t, pool.PoolID, exported.State.ValidatorSlashEvents[0].PoolID)
}

func TestPhase34DoubleSignSlashTombstonesValidator(t *testing.T) {
	k, _, validator := phase34PoolWithValidatorAllocation(t, "phase34-double-sign", "a2", types.DefaultMinPoolDeposit)

	events, err := k.ApplyValidatorSlash(types.MsgApplyValidatorSlash{
		Authority:		prototype.DefaultAuthority,
		ValidatorAddress:	validator,
		Fault:			types.SlashingFaultDoubleSign,
		Epoch:			1,
		Height:			11,
	})
	require.NoError(t, err)
	require.Len(t, events, 1)
	expectedLoss := types.DefaultMinPoolDeposit * uint64(types.DefaultParams().DoubleSignSlashBps) / uint64(types.MaxBasisPoints)
	require.Equal(t, expectedLoss, events[0].SlashingLoss)
	require.Equal(t, types.StateValidatorStatusSlashed, events[0].ValidatorStatus)
	require.True(t, events[0].Tombstoned)

	exported := k.ExportGenesis()
	require.Equal(t, types.StateValidatorStatusSlashed, exported.State.Validators[0].Status)
	require.Equal(t, expectedLoss, exported.State.ValidatorSlashEvents[0].SlashingLoss)
}

func TestPhase34SlashedStateExportImportAndLegacyMigrationCannotRecoverStake(t *testing.T) {
	ctx := context.Background()
	service := kvtest.NewStoreService()
	k := NewPersistentKeeper(service)
	user := aePoolAddress(t, "b1")
	validator := aePoolAddress(t, "b2")
	k.accountStatusReader = accountStatusFixture{user: accountStatusActive, validator: accountStatusActive}
	require.NoError(t, k.InitGenesisState(ctx, DefaultGenesis()))
	pool := createOfficialLiquidStakingPool(t, &k, "phase34-persist")
	require.NoError(t, phase34RegisterAndAllocate(t, &k, pool, user, validator, 2*types.DefaultMinPoolDeposit))

	events, err := k.ApplyValidatorSlash(types.MsgApplyValidatorSlash{
		Authority:		prototype.DefaultAuthority,
		ValidatorAddress:	validator,
		Fault:			types.SlashingFaultDoubleSign,
		Epoch:			2,
		Height:			20,
	})
	require.NoError(t, err)
	require.Len(t, events, 1)
	slashedExport, err := k.ExportGenesisState(ctx)
	require.NoError(t, err)
	slashedTotal := slashedExport.State.Pools[0].TotalBondedStake
	require.NotEqual(t, uint64(2*types.DefaultMinPoolDeposit), slashedTotal)

	imported := NewPersistentKeeper(kvtest.NewStoreService())
	imported.accountStatusReader = accountStatusFixture{user: accountStatusActive, validator: accountStatusActive}
	require.NoError(t, imported.InitGenesisState(ctx, slashedExport))
	roundTrip, err := imported.ExportGenesisState(ctx)
	require.NoError(t, err)
	require.Equal(t, slashedTotal, roundTrip.State.Pools[0].TotalBondedStake)
	require.Equal(t, slashedExport.State.ValidatorSlashEvents, roundTrip.State.ValidatorSlashEvents)
	require.Equal(t, slashedExport.State.PoolValidatorAllocations, roundTrip.State.PoolValidatorAllocations)

	legacyService := kvtest.NewStoreService()
	legacyBytes, err := json.Marshal(slashedExport)
	require.NoError(t, err)
	require.NoError(t, legacyService.RawStore().Set(genesisKey, legacyBytes))
	migrated := NewPersistentKeeper(legacyService)
	migratedExport, err := migrated.ExportGenesisState(ctx)
	require.NoError(t, err)
	require.Equal(t, slashedTotal, migratedExport.State.Pools[0].TotalBondedStake)
	require.Equal(t, slashedExport.State.ValidatorSlashEvents, migratedExport.State.ValidatorSlashEvents)
}

func phase34PoolWithValidatorAllocation(t *testing.T, poolID string, validatorHex string, amount uint64) (Keeper, types.NominatorPool, string) {
	t.Helper()
	user := aePoolAddress(t, "a0")
	validator := aePoolAddress(t, validatorHex)
	k := NewKeeperWithAccountStatus(accountStatusFixture{user: accountStatusActive, validator: accountStatusActive})
	pool := createOfficialLiquidStakingPool(t, &k, poolID)
	require.NoError(t, phase34RegisterAndAllocate(t, &k, pool, user, validator, amount))
	return k, pool, validator
}

func phase34RegisterAndAllocate(t *testing.T, k *Keeper, pool types.NominatorPool, user string, validator string, amount uint64) error {
	t.Helper()
	if _, err := k.RegisterValidator(types.MsgRegisterValidator{
		SignerAddress:		validator,
		ValidatorAddress:	validator,
		SelfStake:		types.DefaultMinValidatorStake,
		CommissionBps:		types.DefaultParams().DefaultValidatorCommissionBps,
		Height:			2,
	}); err != nil {
		return err
	}
	if _, err := k.DepositToStakingPool(types.MsgDepositToStakingPool{
		PoolID:		pool.PoolID,
		WalletAddress:	user,
		Amount:		amount,
		Height:		3,
	}); err != nil {
		return err
	}
	_, err := k.InjectPoolStake(types.MsgInjectPoolStake{
		CallerContractUser:	pool.ContractAddressUser,
		PoolID:			pool.PoolID,
		Allocations: []types.PoolAllocation{{
			ValidatorAddress:	validator,
			Amount:			amount,
			Height:			4,
		}},
		Height:	4,
	})
	return err
}
