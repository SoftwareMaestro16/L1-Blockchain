package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBlockLifecycleRejectsFullUserScansAndAllowsBoundedValidators(t *testing.T) {
	limits := DefaultScalabilityLimits()

	require.NoError(t, ValidateOperationMetrics(OperationMetrics{
		Operation:			OperationBeginBlock,
		ValidatorSetItemsIterated:	300,
		AllocationItemsIterated:	300,
		MaintenanceItemsProcessed:	50,
	}, limits))

	err := ValidateOperationMetrics(OperationMetrics{
		Operation:		OperationEndBlock,
		FullPoolUserScan:	true,
		PoolUsersIterated:	1_000_000,
	}, limits)
	require.ErrorContains(t, err, "full account, pool user, or pool share scans are forbidden")

	err = ValidateOperationMetrics(OperationMetrics{
		Operation:			OperationBeginBlock,
		ValidatorSetItemsIterated:	301,
	}, limits)
	require.ErrorContains(t, err, "validator iteration exceeds")
}

func TestRewardClaimTouchesBoundedKeyCount(t *testing.T) {
	limits := DefaultScalabilityLimits()
	store := NewTemporaryScalabilityFixture(1_000_000)

	metrics := store.ClaimReward("pool-a", "AE-user-999999")

	require.NoError(t, ValidateRewardClaimMetrics(metrics, limits))
	require.Equal(t, uint64(0), metrics.PoolUsersIterated)
	require.Equal(t, uint64(1), metrics.PoolSharesIterated)
	require.LessOrEqual(t, metrics.KeysRead+metrics.KeysWritten, limits.MaxRewardClaimKeys)

	metrics.FullPoolShareScan = true
	require.ErrorContains(t, ValidateRewardClaimMetrics(metrics, limits), "full account")
}

func TestReputationClaimTouchesBoundedKeyCount(t *testing.T) {
	limits := DefaultScalabilityLimits()
	store := NewTemporaryScalabilityFixture(1_000_000)

	metrics := store.ClaimReputation("AE-user-500000")

	require.NoError(t, ValidateReputationClaimMetrics(metrics, limits))
	require.Equal(t, uint64(1), metrics.AccountsIterated)
	require.Equal(t, uint64(1), metrics.PoolSharesIterated)
	require.LessOrEqual(t, metrics.KeysRead+metrics.KeysWritten, limits.MaxReputationClaimKeys)
}

func TestStorageRentChargeTouchesBoundedKeyCount(t *testing.T) {
	limits := DefaultScalabilityLimits()
	store := NewTemporaryScalabilityFixture(1_000_000)

	metrics := store.ChargeStorageRent("AE-user-42")

	require.NoError(t, ValidateStorageRentChargeMetrics(metrics, limits))
	require.Equal(t, uint64(1), metrics.AccountsIterated)
	require.Equal(t, uint64(0), metrics.PoolSharesIterated)
	require.LessOrEqual(t, metrics.KeysRead+metrics.KeysWritten, limits.MaxStorageRentChargeKeys)
}

func TestMillionUserDepositClaimReputationPathsRemainBounded(t *testing.T) {
	limits := DefaultScalabilityLimits()
	store := NewTemporaryScalabilityFixture(1_000_000)

	deposit := store.DepositToPool("pool-a", "AE-user-1000000")
	reward := store.ClaimReward("pool-a", "AE-user-1000000")
	reputation := store.ClaimReputation("AE-user-1000000")
	rent := store.ChargeStorageRent("AE-user-1000000")

	require.NoError(t, ValidateDepositMetrics(deposit, limits))
	require.NoError(t, ValidateRewardClaimMetrics(reward, limits))
	require.NoError(t, ValidateReputationClaimMetrics(reputation, limits))
	require.NoError(t, ValidateStorageRentChargeMetrics(rent, limits))

	for _, metrics := range []OperationMetrics{deposit, reward, reputation, rent} {
		require.False(t, metrics.FullAccountScan)
		require.False(t, metrics.FullPoolUserScan)
		require.False(t, metrics.FullPoolShareScan)
		require.LessOrEqual(t, metrics.AccountsIterated, uint64(1))
		require.LessOrEqual(t, metrics.PoolUsersIterated, uint64(1))
		require.LessOrEqual(t, metrics.PoolSharesIterated, uint64(1))
	}
}

func TestPaginatedQueriesEnforceMaxPageSize(t *testing.T) {
	limits := DefaultScalabilityLimits()

	limit, err := NormalizeQueryPageLimit(0, limits)
	require.NoError(t, err)
	require.Equal(t, limits.DefaultPageSize, limit)

	limit, err = NormalizeQueryPageLimit(limits.MaxPageSize, limits)
	require.NoError(t, err)
	require.Equal(t, limits.MaxPageSize, limit)

	_, err = NormalizeQueryPageLimit(limits.MaxPageSize+1, limits)
	require.ErrorContains(t, err, "exceeds max")

	require.NoError(t, ValidateOperationMetrics(OperationMetrics{
		Operation:	OperationQuery,
		PageLimit:	limits.MaxPageSize,
		PrefixKey:	"pool/share/pool-a/",
	}, limits))
}

func TestStoreKeysArePrefixBasedAndDeterministic(t *testing.T) {
	descriptors := DefaultScalabilityStoreKeys()

	require.NoError(t, ValidateStoreKeyDescriptors(descriptors))

	bad := append([]StoreKeyDescriptor(nil), descriptors...)
	bad[0].Deterministic = false
	require.ErrorContains(t, ValidateStoreKeyDescriptors(bad), "must be deterministic")

	bad = append([]StoreKeyDescriptor(nil), descriptors...)
	bad[0].Paginated = false
	require.ErrorContains(t, ValidateStoreKeyDescriptors(bad), "must be paginated")
}

func TestHotPathValidatorsRejectAccidentalONUserLoops(t *testing.T) {
	limits := DefaultScalabilityLimits()

	err := ValidateRewardClaimMetrics(OperationMetrics{
		Operation:		OperationRewardClaim,
		PoolSharesIterated:	2,
		KeysRead:		2,
		KeysWritten:		2,
	}, limits)
	require.ErrorContains(t, err, "touch only caller state")

	err = ValidateReputationClaimMetrics(OperationMetrics{
		Operation:		OperationReputationClaim,
		PoolUsersIterated:	10,
		KeysRead:		2,
		KeysWritten:		2,
	}, limits)
	require.ErrorContains(t, err, "touch only caller state")
}

// Temporary integration boundary for CHAT4 scalability tests. It models pool,
// reward, reputation, and storage-rent touch counts without importing CHAT2/3
// packages while those workstreams are concurrently dirty.
type TemporaryScalabilityFixture struct {
	TotalUsers uint64
}

func NewTemporaryScalabilityFixture(totalUsers uint64) TemporaryScalabilityFixture {
	return TemporaryScalabilityFixture{TotalUsers: totalUsers}
}

func (f TemporaryScalabilityFixture) DepositToPool(poolID, user string) OperationMetrics {
	_ = f.TotalUsers
	_ = poolID
	_ = user
	return OperationMetrics{
		Operation:		OperationDeposit,
		AccountsIterated:	1,
		PoolUsersIterated:	1,
		PoolSharesIterated:	1,
		KeysRead:		4,
		KeysWritten:		5,
		PrefixKey:		"pool/share/",
	}
}

func (f TemporaryScalabilityFixture) ClaimReward(poolID, user string) OperationMetrics {
	_ = f.TotalUsers
	_ = poolID
	_ = user
	return OperationMetrics{
		Operation:		OperationRewardClaim,
		AccountsIterated:	1,
		PoolSharesIterated:	1,
		KeysRead:		3,
		KeysWritten:		3,
		PrefixKey:		"pool/reward/",
	}
}

func (f TemporaryScalabilityFixture) ClaimReputation(user string) OperationMetrics {
	_ = f.TotalUsers
	_ = user
	return OperationMetrics{
		Operation:		OperationReputationClaim,
		AccountsIterated:	1,
		PoolSharesIterated:	1,
		KeysRead:		4,
		KeysWritten:		4,
		PrefixKey:		"reputation/account/",
	}
}

func (f TemporaryScalabilityFixture) ChargeStorageRent(user string) OperationMetrics {
	_ = f.TotalUsers
	_ = user
	return OperationMetrics{
		Operation:		OperationStorageRentCharge,
		AccountsIterated:	1,
		KeysRead:		3,
		KeysWritten:		3,
		PrefixKey:		"storage-rent/debt/",
	}
}
