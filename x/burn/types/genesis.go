package types

import (
	"fmt"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func DefaultParams() Params {
	return Params{
		AllowedDenoms:	[]string{BaseDenom},
		ProtocolBurnPermissions: []BurnPermission{{
			ModuleName:	ModuleName,
			AllowedDenoms:	[]string{BaseDenom},
		}},
		MaxReasonBytes:	DefaultMaxReasonBytes,
	}
}

func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:		DefaultParams(),
		BurnedByDenom:	[]BurnedByDenomEntry{},
		BurnedByEpoch:	[]BurnedByEpochEntry{},
		BurnReasons:	[]BurnReason{},
	}
}

func NormalizeParams(params Params) Params {
	if len(params.AllowedDenoms) == 0 {
		params.AllowedDenoms = []string{BaseDenom}
	}
	if len(params.ProtocolBurnPermissions) == 0 {
		params.ProtocolBurnPermissions = DefaultParams().ProtocolBurnPermissions
	}
	if params.MaxReasonBytes == 0 {
		params.MaxReasonBytes = DefaultMaxReasonBytes
	}
	for i := range params.ProtocolBurnPermissions {
		if len(params.ProtocolBurnPermissions[i].AllowedDenoms) == 0 {
			params.ProtocolBurnPermissions[i].AllowedDenoms = append([]string(nil), params.AllowedDenoms...)
		}
	}
	return params
}

func (p Params) Validate() error {
	if p.MaxReasonBytes == 0 || p.MaxReasonBytes > 4096 {
		return fmt.Errorf("max_reason_bytes must be between 1 and 4096")
	}
	if err := validateDenomList("allowed_denoms", p.AllowedDenoms); err != nil {
		return err
	}
	seenModules := map[string]struct{}{}
	for _, permission := range p.ProtocolBurnPermissions {
		if strings.TrimSpace(permission.ModuleName) == "" {
			return fmt.Errorf("burn permission module name must be set")
		}
		if _, ok := seenModules[permission.ModuleName]; ok {
			return fmt.Errorf("duplicate burn permission module %s", permission.ModuleName)
		}
		seenModules[permission.ModuleName] = struct{}{}
		if err := validateDenomList("permission allowed_denoms", permission.AllowedDenoms); err != nil {
			return err
		}
		for _, denom := range permission.AllowedDenoms {
			if !contains(p.AllowedDenoms, denom) {
				return fmt.Errorf("permission denom %s must be globally allowed", denom)
			}
		}
	}
	return nil
}

func (gs GenesisState) Validate() error {
	params := NormalizeParams(gs.Params)
	if err := params.Validate(); err != nil {
		return err
	}
	seenDenoms := map[string]struct{}{}
	for _, entry := range gs.BurnedByDenom {
		if _, ok := seenDenoms[entry.Denom]; ok {
			return fmt.Errorf("duplicate burned denom %s", entry.Denom)
		}
		seenDenoms[entry.Denom] = struct{}{}
		if err := entry.Validate(params); err != nil {
			return err
		}
	}
	seenEpochs := map[uint64]struct{}{}
	for _, entry := range gs.BurnedByEpoch {
		if _, ok := seenEpochs[entry.Epoch]; ok {
			return fmt.Errorf("duplicate burned epoch %d", entry.Epoch)
		}
		seenEpochs[entry.Epoch] = struct{}{}
		if err := entry.Validate(params); err != nil {
			return err
		}
	}
	seenReasons := map[uint64]struct{}{}
	for _, reason := range gs.BurnReasons {
		if _, ok := seenReasons[reason.Id]; ok {
			return fmt.Errorf("duplicate burn reason id %d", reason.Id)
		}
		seenReasons[reason.Id] = struct{}{}
		if err := reason.Validate(params); err != nil {
			return err
		}
	}
	return nil
}

func (e BurnedByDenomEntry) Validate(params Params) error {
	if e.Denom == "" {
		return fmt.Errorf("burned denom must be set")
	}
	if len(e.Amount) != 1 || e.Amount[0].Denom != e.Denom {
		return fmt.Errorf("burned denom entry must contain only %s", e.Denom)
	}
	return ValidateBurnCoins(params, e.Amount)
}

func (e BurnedByEpochEntry) Validate(params Params) error {
	if e.Epoch == 0 {
		return fmt.Errorf("burned epoch must be positive")
	}
	return ValidateBurnCoins(params, e.Amount)
}

func (r BurnReason) Validate(params Params) error {
	if r.Id == 0 {
		return fmt.Errorf("burn reason id must be positive")
	}
	if r.Epoch == 0 {
		return fmt.Errorf("burn reason epoch must be positive")
	}
	if len(r.Reason) > int(params.MaxReasonBytes) {
		return fmt.Errorf("burn reason exceeds max_reason_bytes")
	}
	if r.Protocol && strings.TrimSpace(r.SourceModule) == "" {
		return fmt.Errorf("protocol burn source module must be set")
	}
	if !r.Protocol && strings.TrimSpace(r.Burner) == "" {
		return fmt.Errorf("user burn burner must be set")
	}
	return ValidateBurnCoins(params, r.Amount)
}

func ValidateBurnCoins(params Params, amount sdk.Coins) error {
	params = NormalizeParams(params)
	if err := params.Validate(); err != nil {
		return ErrInvalidParams.Wrap(err.Error())
	}
	if amount.Empty() {
		return ErrInvalidBurn.Wrap("burn amount must be positive")
	}
	if !amount.IsValid() {
		return ErrInvalidBurn.Wrapf("invalid burn amount: %s", amount)
	}
	for _, coin := range amount {
		if coin.IsNil() || !coin.IsPositive() {
			return ErrInvalidBurn.Wrapf("burn coin must be positive: %s", coin)
		}
		if !contains(params.AllowedDenoms, coin.Denom) {
			return ErrInvalidBurn.Wrapf("burn denom %s is not approved", coin.Denom)
		}
	}
	return nil
}

func ValidateProtocolBurn(params Params, sourceModule string, amount sdk.Coins) error {
	if err := ValidateBurnCoins(params, amount); err != nil {
		return err
	}
	for _, permission := range params.ProtocolBurnPermissions {
		if permission.ModuleName != sourceModule {
			continue
		}
		for _, coin := range amount {
			if !contains(permission.AllowedDenoms, coin.Denom) {
				return ErrUnauthorized.Wrapf("module %s cannot burn denom %s", sourceModule, coin.Denom)
			}
		}
		return nil
	}
	return ErrUnauthorized.Wrapf("module %s is not allowed to burn protocol coins", sourceModule)
}

func SortBurnedByDenom(in []BurnedByDenomEntry) []BurnedByDenomEntry {
	out := append([]BurnedByDenomEntry(nil), in...)
	sort.Slice(out, func(i, j int) bool { return out[i].Denom < out[j].Denom })
	return out
}

func SortBurnedByEpoch(in []BurnedByEpochEntry) []BurnedByEpochEntry {
	out := append([]BurnedByEpochEntry(nil), in...)
	sort.Slice(out, func(i, j int) bool { return out[i].Epoch < out[j].Epoch })
	return out
}

func SortBurnReasons(in []BurnReason) []BurnReason {
	out := append([]BurnReason(nil), in...)
	sort.Slice(out, func(i, j int) bool { return out[i].Id < out[j].Id })
	return out
}

func validateDenomList(name string, denoms []string) error {
	if len(denoms) == 0 {
		return fmt.Errorf("%s must not be empty", name)
	}
	seen := map[string]struct{}{}
	for _, denom := range denoms {
		if err := sdk.ValidateDenom(denom); err != nil {
			return fmt.Errorf("invalid %s denom %s: %w", name, denom, err)
		}
		if _, ok := seen[denom]; ok {
			return fmt.Errorf("duplicate %s denom %s", name, denom)
		}
		seen[denom] = struct{}{}
	}
	return nil
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
