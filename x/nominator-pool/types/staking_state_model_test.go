package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/app/addressing"
)

func TestStakingStateKeyGoldenValues(t *testing.T) {
	validator := aeAddressFromRaw(t, rawAddress("11"))
	owner := aeAddressFromRaw(t, rawAddress("22"))
	contractRaw := rawAddress("66")
	contractUser := aeAddressFromRaw(t, contractRaw)

	require.Equal(t, "staking/validator/"+validator, string(ValidatorKey(validator)))
	require.Equal(t, "staking/validator_score/"+validator+"/00000000000000000007", string(ValidatorScoreKey(validator, 7)))
	require.Equal(t, "staking/pool/pool-a", string(PoolKey("pool-a")))
	require.Equal(t, "staking/pool_storage_debt/pool-a", string(PoolStorageDebtKey("pool-a")))
	require.Equal(t, "staking/pool_by_contract_user/"+contractUser, string(PoolByContractUserKey(contractUser)))
	require.Equal(t, "staking/pool_by_contract_raw/"+contractRaw, string(PoolByContractRawKey(contractRaw)))
	require.Equal(t, "staking/pool_share/pool-a/"+owner, string(PoolShareKey("pool-a", owner)))
	require.Equal(t, "staking/pool_allocation/pool-a/"+validator, string(PoolAllocationKey("pool-a", validator)))
	require.Equal(t, "staking/pool_unbonding/pool-a/"+owner+"/req-1", string(PoolUnbondingKey("pool-a", owner, "req-1")))
	require.Equal(t, "staking/pool_reward_index/pool-a", string(PoolRewardIndexKey("pool-a")))
	require.Equal(t, "staking/reward_claim/pool-a/"+owner+"/00000000000000000009", string(RewardClaimKey("pool-a", owner, 9)))
	require.Equal(t, "staking/snapshot/epoch/00000000000000000003", string(EpochSnapshotKey(3)))
	require.Equal(t, "staking/snapshot/validator_set/00000000000000000123", string(ValidatorSetSnapshotKey(123)))
}

func TestLiquidStakingPoolAddressPairValidation(t *testing.T) {
	params := DefaultParams()
	raw := rawAddress("66")
	pool := validLiquidPool(t, "pool-a", raw)
	require.NoError(t, pool.Validate(params))

	pool.ContractAddressRaw = rawAddress("67")
	require.ErrorContains(t, pool.Validate(params), "same account")

	user := aeAddressFromRaw(t, raw)
	roundTripRaw, err := RawAddressForUserAddress(user)
	require.NoError(t, err)
	require.Equal(t, raw, roundTripRaw)
}

func TestPoolShareRejectsRawUserFacingOwner(t *testing.T) {
	share := PoolShare{
		Owner:			rawAddress("22"),
		PoolID:			"pool-a",
		Shares:			10,
		PrincipalAmount:	10,
		CreatedHeight:		1,
		UpdatedHeight:		1,
	}
	require.ErrorContains(t, share.Validate(DefaultParams()), "AE user-facing")
}

func TestPoolAllocationRequiresActiveEligibleValidator(t *testing.T) {
	params := DefaultParams()
	validator := validStateValidator(t, "11")
	allocation := PoolValidatorAllocation{
		PoolID:			"pool-a",
		Validator:		validator.Address,
		TargetWeightBps:	params.MinPoolValidatorAllocationBps,
		ActiveStake:		10,
		PerformanceScore:	9_000,
		CommissionBps:		1_000,
		SlashingRiskBps:	100,
		UpdatedHeight:		2,
	}
	require.NoError(t, allocation.Validate(params, validator))

	validator.Status = StateValidatorStatusJailed
	require.ErrorContains(t, allocation.Validate(params, validator), "non-active validator must have zero target weight")
}

func TestPaginatedStateQueriesAreDeterministic(t *testing.T) {
	ownerA := aeAddressFromRaw(t, rawAddress("22"))
	ownerB := aeAddressFromRaw(t, rawAddress("33"))
	shares := []PoolShare{
		{Owner: ownerA, PoolID: "pool-b", Shares: 1, PrincipalAmount: 1, CreatedHeight: 1, UpdatedHeight: 1},
		{Owner: ownerB, PoolID: "pool-a", Shares: 1, PrincipalAmount: 1, CreatedHeight: 1, UpdatedHeight: 1},
		{Owner: ownerA, PoolID: "pool-a", Shares: 1, PrincipalAmount: 1, CreatedHeight: 1, UpdatedHeight: 1},
	}
	ownerPage := PaginatePoolSharesByOwner(shares, ownerA, 0, 1)
	require.Len(t, ownerPage, 1)
	require.Equal(t, "pool-a", ownerPage[0].PoolID)

	validatorA := aeAddressFromRaw(t, rawAddress("11"))
	validatorB := aeAddressFromRaw(t, rawAddress("12"))
	allocations := []PoolValidatorAllocation{
		{PoolID: "pool-a", Validator: validatorB, TargetWeightBps: 25, UpdatedHeight: 1},
		{PoolID: "pool-b", Validator: validatorA, TargetWeightBps: 25, UpdatedHeight: 1},
		{PoolID: "pool-a", Validator: validatorA, TargetWeightBps: 25, UpdatedHeight: 1},
	}
	allocationPage := PaginatePoolAllocationsByPool(allocations, "pool-a", 1, 1)
	require.Len(t, allocationPage, 1)
	require.Equal(t, validatorB, allocationPage[0].Validator)

	validators := []Validator{validStateValidator(t, "33"), validStateValidator(t, "11"), validStateValidator(t, "22")}
	validatorPage := PaginateValidators(validators, 1, 1)
	require.Len(t, validatorPage, 1)
	require.Equal(t, aeAddressFromRaw(t, rawAddress("22")), validatorPage[0].Address)
}

func TestSnapshotExportIsDeterministicAndStateRoundTrips(t *testing.T) {
	params := DefaultParams()
	owner := aeAddressFromRaw(t, rawAddress("22"))
	validator := validStateValidator(t, "11")
	pool := validLiquidPool(t, "pool-a", rawAddress("66"))
	state := State{
		Validators:			[]Validator{validator},
		ValidatorPerformanceScores:	[]ValidatorPerformanceScore{{Validator: validator.Address, Epoch: 1, ScoreBps: 9_000}},
		ValidatorCommissions:		[]ValidatorCommission{{Validator: validator.Address, Epoch: 1, RateBps: 1_000}},
		ValidatorSlashingRisks:		[]ValidatorSlashingRisk{{Validator: validator.Address, Epoch: 1, RiskBps: 100}},
		ValidatorAllocationLimits:	[]ValidatorAllocationLimit{{Validator: validator.Address, Epoch: 1, LimitBps: params.MaxPoolValidatorAllocationBps}},
		LiquidStakingPools:		[]LiquidStakingPool{pool},
		PoolShares: []PoolShare{{
			Owner:			owner,
			PoolID:			pool.PoolID,
			Shares:			10,
			PrincipalAmount:	10,
			CreatedHeight:		1,
			UpdatedHeight:		2,
		}},
		PoolValidatorAllocations: []PoolValidatorAllocation{{
			PoolID:			pool.PoolID,
			Validator:		validator.Address,
			TargetWeightBps:	params.MinPoolValidatorAllocationBps,
			ActiveStake:		10,
			PerformanceScore:	9_000,
			CommissionBps:		1_000,
			SlashingRiskBps:	100,
			UpdatedHeight:		2,
		}},
		PoolUnbondingRequests: []PoolUnbondingRequest{{
			PoolID:		pool.PoolID,
			Owner:		owner,
			RequestID:	"req-1",
			Shares:		1,
			Amount:		1,
			RequestHeight:	3,
			CompleteHeight:	4,
			Status:		WithdrawalStatusPending,
		}},
		PoolRewardIndexes:	[]PoolRewardIndex{{PoolID: pool.PoolID, RewardIndex: 7, Epoch: 1}},
		RewardClaims:		[]RewardClaim{{PoolID: pool.PoolID, Owner: owner, Epoch: 1, Amount: 2}},
		EpochStakingSnapshots:	[]EpochStakingSnapshot{{Epoch: 1, TotalActiveStake: 10, TotalPools: 1, ValidatorCount: 1, SnapshotHash: "hash-a"}},
		ValidatorSetSnapshots:	[]ValidatorSetSnapshot{{HeightOrEpoch: 10, Validators: []string{validator.Address}, TotalPower: 10, SnapshotHash: "hash-b"}},
	}
	normalized := state.Normalize(params)
	require.NoError(t, normalized.Validate(params))

	first, err := json.Marshal(normalized)
	require.NoError(t, err)
	second, err := json.Marshal(state.Normalize(params))
	require.NoError(t, err)
	require.Equal(t, string(first), string(second))

	var imported State
	require.NoError(t, json.Unmarshal(first, &imported))
	require.Equal(t, normalized, imported.Normalize(params))
	require.Equal(t, pool.StorageRentDebt, imported.LiquidStakingPools[0].StorageRentDebt)
	require.Equal(t, normalized.ValidatorPerformanceScores, imported.Normalize(params).ValidatorPerformanceScores)
	require.Equal(t, normalized.ValidatorCommissions, imported.Normalize(params).ValidatorCommissions)
	require.Equal(t, normalized.ValidatorSlashingRisks, imported.Normalize(params).ValidatorSlashingRisks)
	require.Equal(t, normalized.ValidatorAllocationLimits, imported.Normalize(params).ValidatorAllocationLimits)
}

func validLiquidPool(t *testing.T, poolID string, raw string) LiquidStakingPool {
	t.Helper()
	return LiquidStakingPool{
		PoolID:				poolID,
		ContractAddressUser:		aeAddressFromRaw(t, raw),
		ContractAddressRaw:		raw,
		ReceiptToken:			"aet-liquid-staking-share-v1",
		TotalDeposited:			100,
		TotalActiveStake:		70,
		TotalUnbonding:			30,
		TotalShares:			100,
		RewardIndex:			7,
		AllocationEpoch:		1,
		LastStorageChargeHeight:	10,
		StorageRentDebt:		5,
		RentPayerPolicy:		RentPayerPolicyPoolReserve,
		Status:				PoolStatusActive,
	}
}

func validStateValidator(t *testing.T, hexByte string) Validator {
	t.Helper()
	params := DefaultParams()
	return Validator{
		Address:		aeAddressFromRaw(t, rawAddress(hexByte)),
		SelfStake:		params.PoolBackedValidatorMinSelfStake,
		NominatorStake:		params.PoolBackedValidatorMaxNominatorStake,
		Status:			StateValidatorStatusActive,
		PerformanceScore:	9_000,
		CommissionBps:		1_000,
		SlashingRiskBps:	100,
		AllocationLimitBps:	params.MaxPoolValidatorAllocationBps,
		UpdatedHeight:		1,
	}
}

func rawAddress(hexByte string) string {
	return "4:000000000000000000000000" + stringsRepeat(hexByte, 20)
}

func aeAddressFromRaw(t *testing.T, raw string) string {
	t.Helper()
	bz, err := addressing.Parse(raw)
	require.NoError(t, err)
	user, err := addressing.FormatUserFriendly(bz)
	require.NoError(t, err)
	return user
}

func stringsRepeat(value string, count int) string {
	out := ""
	for i := 0; i < count; i++ {
		out += value
	}
	return out
}
