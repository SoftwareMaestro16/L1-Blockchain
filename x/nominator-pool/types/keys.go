package types

import "fmt"

const (
	ModuleName	= "nominator-pool"
	StoreKey	= ModuleName
)

func ValidatorKey(validator string) []byte {
	return []byte("staking/validator/" + validator)
}

func ValidatorScoreKey(validator string, epoch uint64) []byte {
	return []byte(fmt.Sprintf("staking/validator_score/%s/%020d", validator, epoch))
}

func PoolKey(poolID string) []byte {
	return []byte("staking/pool/" + poolID)
}

func PoolStorageDebtKey(poolID string) []byte {
	return []byte("staking/pool_storage_debt/" + poolID)
}

func PoolByContractUserKey(contractAddressUser string) []byte {
	return []byte("staking/pool_by_contract_user/" + contractAddressUser)
}

func PoolByContractRawKey(contractAddressRaw string) []byte {
	return []byte("staking/pool_by_contract_raw/" + contractAddressRaw)
}

func PoolShareKey(poolID string, owner string) []byte {
	return []byte("staking/pool_share/" + poolID + "/" + owner)
}

func PoolAllocationKey(poolID string, validator string) []byte {
	return []byte("staking/pool_allocation/" + poolID + "/" + validator)
}

func PoolUnbondingKey(poolID string, owner string, requestID string) []byte {
	return []byte("staking/pool_unbonding/" + poolID + "/" + owner + "/" + requestID)
}

func PoolRewardIndexKey(poolID string) []byte {
	return []byte("staking/pool_reward_index/" + poolID)
}

func RewardClaimKey(poolID string, owner string, epoch uint64) []byte {
	return []byte(fmt.Sprintf("staking/reward_claim/%s/%s/%020d", poolID, owner, epoch))
}

func EpochSnapshotKey(epoch uint64) []byte {
	return []byte(fmt.Sprintf("staking/snapshot/epoch/%020d", epoch))
}

func ValidatorSetSnapshotKey(heightOrEpoch uint64) []byte {
	return []byte(fmt.Sprintf("staking/snapshot/validator_set/%020d", heightOrEpoch))
}

func IdentityReputationKey(account string) []byte {
	return []byte("identity/reputation/" + account)
}
