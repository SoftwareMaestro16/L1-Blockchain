package types

import (
	"fmt"
	"sort"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

const (
	BucketValidatorRewards		= "validator_rewards"
	BucketTreasury			= "treasury"
	BucketDelegatorProtection	= "delegator_protection"
	BucketValidatorInsuranceReserve	= "validator_insurance_reserve"
	BucketEcosystemGrants		= "ecosystem_grants"
	BucketStorageRentReserve	= "storage_rent_reserve"
	BucketBurn			= "burn"
	BucketReporterRewards		= "reporter_rewards"
)

type ProtocolIncomePolicy struct {
	Scale	uint32				`json:"scale"`
	Buckets	[]ProtocolIncomeBucketRule	`json:"buckets"`
}

type ProtocolIncomeBucketRule struct {
	Bucket		string	`json:"bucket"`
	ModuleAccount	string	`json:"module_account"`
	Weight		uint32	`json:"weight"`
	AllowZeroWeight	bool	`json:"allow_zero_weight"`
	ConstitutionMin	uint32	`json:"constitution_min"`
	Burn		bool	`json:"burn"`
}

type ProtocolIncomeAllocation struct {
	Bucket		string
	ModuleAccount	string
	Amount		sdk.Coins
	Burn		bool
}

func DefaultProtocolIncomePolicy() ProtocolIncomePolicy {
	return ProtocolIncomePolicy{
		Scale:	BasisPoints,
		Buckets: []ProtocolIncomeBucketRule{
			{Bucket: BucketValidatorRewards, ModuleAccount: authtypes.FeeCollectorName, Weight: 3_800},
			{Bucket: BucketTreasury, ModuleAccount: TreasuryModuleName, Weight: 2_500},
			{Bucket: BucketDelegatorProtection, ModuleAccount: ProtectionModuleName, Weight: 1_000, ConstitutionMin: 500},
			{Bucket: BucketValidatorInsuranceReserve, ModuleAccount: ValidatorInsuranceModuleName, Weight: 800, ConstitutionMin: 200},
			{Bucket: BucketEcosystemGrants, ModuleAccount: EcosystemGrantsModuleName, Weight: 1_200},
			{Bucket: BucketStorageRentReserve, ModuleAccount: StorageRentReserveModuleName, Weight: 500, ConstitutionMin: 100},
			{Bucket: BucketBurn, ModuleAccount: BurnModuleName, Weight: 200, ConstitutionMin: 1, Burn: true},
			{Bucket: BucketReporterRewards, ModuleAccount: ReporterRewardsModuleName, Weight: 0, AllowZeroWeight: true},
		},
	}
}

func NormalizeProtocolIncomePolicy(policy ProtocolIncomePolicy) ProtocolIncomePolicy {
	if policy.Scale == 0 {
		policy.Scale = BasisPoints
	}
	if len(policy.Buckets) == 0 {
		policy.Buckets = DefaultProtocolIncomePolicy().Buckets
	}
	policy.Buckets = append([]ProtocolIncomeBucketRule(nil), policy.Buckets...)
	sort.SliceStable(policy.Buckets, func(i, j int) bool {
		return bucketOrder(policy.Buckets[i].Bucket) < bucketOrder(policy.Buckets[j].Bucket)
	})
	return policy
}

func (p ProtocolIncomePolicy) Validate() error {
	p = NormalizeProtocolIncomePolicy(p)
	if p.Scale != BasisPoints {
		return fmt.Errorf("protocol income scale must be %d", BasisPoints)
	}
	total := uint64(0)
	seen := map[string]struct{}{}
	for _, bucket := range p.Buckets {
		if bucket.Bucket == "" {
			return fmt.Errorf("protocol income bucket name is required")
		}
		if _, found := seen[bucket.Bucket]; found {
			return fmt.Errorf("duplicate protocol income bucket %s", bucket.Bucket)
		}
		seen[bucket.Bucket] = struct{}{}
		if bucket.ModuleAccount == "" {
			return fmt.Errorf("protocol income bucket %s requires module account", bucket.Bucket)
		}
		if bucket.Weight == 0 && !bucket.AllowZeroWeight {
			return fmt.Errorf("protocol income bucket %s has zero weight without explicit allowance", bucket.Bucket)
		}
		if bucket.Weight < bucket.ConstitutionMin {
			return fmt.Errorf("protocol income bucket %s violates constitutional minimum", bucket.Bucket)
		}
		total += uint64(bucket.Weight)
	}
	if total != uint64(p.Scale) {
		return fmt.Errorf("protocol income weights must sum to %d bps", p.Scale)
	}
	return nil
}

func SplitProtocolIncome(policy ProtocolIncomePolicy, fees sdk.Coins) ([]ProtocolIncomeAllocation, sdk.Coins, error) {
	policy = NormalizeProtocolIncomePolicy(policy)
	if err := policy.Validate(); err != nil {
		return nil, nil, ErrInvalidParams.Wrap(err.Error())
	}
	if err := ValidateFeeCoins(BaseDenom, fees); err != nil {
		return nil, nil, err
	}
	allocations := make([]ProtocolIncomeAllocation, len(policy.Buckets))
	for i, bucket := range policy.Buckets {
		allocations[i] = ProtocolIncomeAllocation{Bucket: bucket.Bucket, ModuleAccount: bucket.ModuleAccount, Amount: sdk.NewCoins(), Burn: bucket.Burn}
	}
	remainder := sdk.NewCoins()
	for _, fee := range fees {
		parts, rem := splitCoinByLargestRemainder(policy, fee)
		for i, amount := range parts {
			if amount.IsPositive() {
				allocations[i].Amount = allocations[i].Amount.Add(sdk.NewCoin(fee.Denom, amount))
			}
		}
		if rem.IsPositive() {
			remainder = remainder.Add(sdk.NewCoin(fee.Denom, rem))
		}
	}
	return allocations, remainder, nil
}

type fractionalRemainder struct {
	index		int
	remainder	sdkmath.Int
}

func splitCoinByLargestRemainder(policy ProtocolIncomePolicy, fee sdk.Coin) ([]sdkmath.Int, sdkmath.Int) {
	amounts := make([]sdkmath.Int, len(policy.Buckets))
	remainders := make([]fractionalRemainder, 0, len(policy.Buckets))
	allocated := sdkmath.ZeroInt()
	for i, bucket := range policy.Buckets {
		product := fee.Amount.MulRaw(int64(bucket.Weight))
		amount := product.QuoRaw(int64(policy.Scale))
		rem := product.ModRaw(int64(policy.Scale))
		amounts[i] = amount
		allocated = allocated.Add(amount)
		if bucket.Weight > 0 && rem.IsPositive() {
			remainders = append(remainders, fractionalRemainder{index: i, remainder: rem})
		}
	}
	leftover := fee.Amount.Sub(allocated)
	sort.SliceStable(remainders, func(i, j int) bool {
		if !remainders[i].remainder.Equal(remainders[j].remainder) {
			return remainders[i].remainder.GT(remainders[j].remainder)
		}
		return bucketOrder(policy.Buckets[remainders[i].index].Bucket) < bucketOrder(policy.Buckets[remainders[j].index].Bucket)
	})
	recordedRemainder := leftover
	for i := 0; leftover.IsPositive() && i < len(remainders); i++ {
		amounts[remainders[i].index] = amounts[remainders[i].index].Add(sdkmath.OneInt())
		leftover = leftover.Sub(sdkmath.OneInt())
	}
	return amounts, recordedRemainder
}

func bucketOrder(bucket string) int {
	switch bucket {
	case BucketValidatorRewards:
		return 0
	case BucketTreasury:
		return 1
	case BucketDelegatorProtection:
		return 2
	case BucketValidatorInsuranceReserve:
		return 3
	case BucketEcosystemGrants:
		return 4
	case BucketStorageRentReserve:
		return 5
	case BucketBurn:
		return 6
	case BucketReporterRewards:
		return 7
	default:
		return 1000
	}
}
