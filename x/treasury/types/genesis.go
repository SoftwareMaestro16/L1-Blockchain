package types

import (
	"fmt"
	"sort"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
)

func DefaultParams() Params {
	return Params{
		BaseDenom:		BaseDenom,
		TreasuryModule:		TreasuryModuleName,
		ReserveBps:		5_000,
		EcosystemBps:		3_000,
		ValidatorIncentivesBps:	1_500,
		BurnBps:		500,
		PerEpochSpendCap:	sdk.NewInt64Coin(BaseDenom, 1_000_000_000_000),
		MaxMetadataBytes:	DefaultMaxMetadataBytes,
	}
}

func DefaultAllocations() TreasuryAllocations {
	return TreasuryAllocations{
		ReserveBalance:			sdk.NewCoins(),
		EcosystemBalance:		sdk.NewCoins(),
		ValidatorIncentiveBalance:	sdk.NewCoins(),
		BurnBalance:			sdk.NewCoins(),
		TotalReceived:			sdk.NewCoins(),
		TotalSpent:			sdk.NewCoins(),
	}
}

func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:		DefaultParams(),
		Allocations:	DefaultAllocations(),
		Spends:		[]TreasurySpend{},
		EpochSpends:	[]EpochSpend{},
		NextSpendId:	1,
	}
}

func NormalizeParams(params Params) Params {
	if params.BaseDenom == "" {
		params.BaseDenom = BaseDenom
	}
	if params.TreasuryModule == "" {
		params.TreasuryModule = TreasuryModuleName
	}
	if params.MaxMetadataBytes == 0 {
		params.MaxMetadataBytes = DefaultMaxMetadataBytes
	}
	if params.PerEpochSpendCap.Denom == "" && params.PerEpochSpendCap.Amount.IsNil() {
		params.PerEpochSpendCap = sdk.NewInt64Coin(params.BaseDenom, 1_000_000_000_000)
	}
	return params
}

func (p Params) Validate() error {
	if p.BaseDenom != BaseDenom {
		return fmt.Errorf("base_denom must be %s", BaseDenom)
	}
	if p.TreasuryModule != TreasuryModuleName {
		return fmt.Errorf("treasury_module must be %s", TreasuryModuleName)
	}
	total := uint64(p.ReserveBps) + uint64(p.EcosystemBps) + uint64(p.ValidatorIncentivesBps) + uint64(p.BurnBps)
	if total != uint64(BasisPoints) {
		return fmt.Errorf("distribution proportions must sum to %d bps", BasisPoints)
	}
	if p.MaxMetadataBytes == 0 || p.MaxMetadataBytes > 4096 {
		return fmt.Errorf("max_metadata_bytes must be between 1 and 4096")
	}
	if err := validateCapCoin(p.BaseDenom, p.PerEpochSpendCap); err != nil {
		return err
	}
	seen := map[string]struct{}{}
	for _, recipient := range p.RecipientAllowlist {
		if err := aetraaddress.ValidateUserAddress("recipient_allowlist", recipient); err != nil {
			return err
		}
		if _, ok := seen[recipient]; ok {
			return fmt.Errorf("duplicate recipient allowlist address %s", recipient)
		}
		seen[recipient] = struct{}{}
	}
	return nil
}

func (a TreasuryAllocations) Validate(params Params) error {
	for name, coins := range map[string]sdk.Coins{
		"reserve_balance":		a.ReserveBalance,
		"ecosystem_balance":		a.EcosystemBalance,
		"validator_incentive_balance":	a.ValidatorIncentiveBalance,
		"burn_balance":			a.BurnBalance,
		"total_received":		a.TotalReceived,
		"total_spent":			a.TotalSpent,
	} {
		if err := ValidateTreasuryCoins(params, coins, true); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
	}
	return nil
}

func (a TreasuryAllocations) AccountingBalance() sdk.Coins {
	return a.ReserveBalance.Add(a.EcosystemBalance...).Add(a.ValidatorIncentiveBalance...).Add(a.BurnBalance...)
}

func (s TreasurySpend) Validate(params Params) error {
	if s.Id == 0 {
		return fmt.Errorf("spend id must be positive")
	}
	if err := aetraaddress.ValidateUserAddress("proposer", s.Proposer); err != nil {
		return err
	}
	if err := aetraaddress.ValidateUserAddress("recipient", s.Recipient); err != nil {
		return err
	}
	if err := ValidateTreasuryCoins(params, s.Amount, false); err != nil {
		return err
	}
	if !IsSpendableBucket(s.Bucket) {
		return fmt.Errorf("bucket %s is not spendable", s.Bucket)
	}
	if !IsSpendStatus(s.Status) {
		return fmt.Errorf("invalid spend status %s", s.Status)
	}
	if s.Epoch == 0 {
		return fmt.Errorf("spend epoch must be positive")
	}
	if len(s.Metadata) > int(params.MaxMetadataBytes) {
		return fmt.Errorf("metadata exceeds max_metadata_bytes")
	}
	if s.VestingEndEpoch != 0 && s.VestingStartEpoch != 0 && s.VestingEndEpoch < s.VestingStartEpoch {
		return fmt.Errorf("vesting_end_epoch must be greater than or equal to vesting_start_epoch")
	}
	return nil
}

func (e EpochSpend) Validate(params Params) error {
	if e.Epoch == 0 {
		return fmt.Errorf("epoch spend epoch must be positive")
	}
	return ValidateTreasuryCoins(params, e.Spent, true)
}

func (gs GenesisState) Validate() error {
	params := NormalizeParams(gs.Params)
	if err := params.Validate(); err != nil {
		return err
	}
	if err := gs.Allocations.Validate(params); err != nil {
		return err
	}
	seenSpends := map[uint64]struct{}{}
	maxSpendID := uint64(0)
	for _, spend := range gs.Spends {
		if _, ok := seenSpends[spend.Id]; ok {
			return fmt.Errorf("duplicate spend id %d", spend.Id)
		}
		seenSpends[spend.Id] = struct{}{}
		if spend.Id > maxSpendID {
			maxSpendID = spend.Id
		}
		if err := spend.Validate(params); err != nil {
			return err
		}
	}
	seenEpochs := map[uint64]struct{}{}
	for _, epochSpend := range gs.EpochSpends {
		if _, ok := seenEpochs[epochSpend.Epoch]; ok {
			return fmt.Errorf("duplicate epoch spend %d", epochSpend.Epoch)
		}
		seenEpochs[epochSpend.Epoch] = struct{}{}
		if err := epochSpend.Validate(params); err != nil {
			return err
		}
	}
	nextID := gs.NextSpendId
	if nextID == 0 {
		nextID = 1
	}
	if nextID <= maxSpendID {
		return fmt.Errorf("next_spend_id must be greater than existing spend ids")
	}
	return nil
}

func ValidateTreasuryCoins(params Params, amount sdk.Coins, allowEmpty bool) error {
	params = NormalizeParams(params)
	if !allowEmpty && amount.Empty() {
		return ErrInvalidSpend.Wrap("amount must be positive")
	}
	if !amount.IsValid() {
		return ErrInvalidSpend.Wrapf("invalid amount %s", amount)
	}
	for _, coin := range amount {
		if coin.IsNil() || coin.IsNegative() || (!allowEmpty && coin.IsZero()) {
			return ErrInvalidSpend.Wrapf("coin must be positive: %s", coin)
		}
		if coin.Denom != params.BaseDenom {
			return ErrInvalidSpend.Wrapf("denom %s is not treasury base denom", coin.Denom)
		}
	}
	return nil
}

func IsBucket(bucket string) bool {
	switch bucket {
	case BucketReserve, BucketEcosystem, BucketValidatorIncentives, BucketBurn:
		return true
	default:
		return false
	}
}

func IsSpendableBucket(bucket string) bool {
	switch bucket {
	case BucketReserve, BucketEcosystem, BucketValidatorIncentives:
		return true
	default:
		return false
	}
}

func IsSpendStatus(status string) bool {
	switch status {
	case StatusPending, StatusApproved, StatusRejected, StatusExecuted, StatusCanceled:
		return true
	default:
		return false
	}
}

func SpendIsTerminal(status string) bool {
	switch status {
	case StatusRejected, StatusExecuted, StatusCanceled:
		return true
	default:
		return false
	}
}

func SortSpends(in []TreasurySpend) []TreasurySpend {
	out := append([]TreasurySpend(nil), in...)
	sort.Slice(out, func(i, j int) bool { return out[i].Id < out[j].Id })
	return out
}

func SortEpochSpends(in []EpochSpend) []EpochSpend {
	out := append([]EpochSpend(nil), in...)
	sort.Slice(out, func(i, j int) bool { return out[i].Epoch < out[j].Epoch })
	return out
}

func IsRecipientAllowed(params Params, recipient string) bool {
	if !params.RecipientAllowlistEnabled {
		return true
	}
	for _, allowed := range params.RecipientAllowlist {
		if allowed == recipient {
			return true
		}
	}
	return false
}

func validateCapCoin(baseDenom string, coin sdk.Coin) error {
	if coin.Denom != baseDenom {
		return fmt.Errorf("per_epoch_spend_cap denom must be %s", baseDenom)
	}
	if coin.Amount.IsNil() || coin.Amount.IsNegative() {
		return fmt.Errorf("per_epoch_spend_cap must be non-negative")
	}
	return nil
}

func BpsAmount(amount sdkmath.Int, bps uint32) sdkmath.Int {
	if amount.IsZero() || bps == 0 {
		return sdkmath.ZeroInt()
	}
	return amount.MulRaw(int64(bps)).QuoRaw(int64(BasisPoints))
}
