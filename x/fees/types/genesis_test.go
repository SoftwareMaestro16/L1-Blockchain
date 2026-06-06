package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/sovereign-l1/l1/app/params"
)

const fixtureTestAssetDenom = "testtoken"

func TestDefaultParamsValidate(t *testing.T) {
	params := DefaultParams()
	if len(params.AllowedFeeDenoms) != 1 || params.AllowedFeeDenoms[0] != appparams.BaseDenom {
		t.Fatalf("expected only naet as default fee denom: %v", params.AllowedFeeDenoms)
	}
	if params.MinFeeAmount != "1" {
		t.Fatalf("expected min fee amount 1, got %q", params.MinFeeAmount)
	}
	if params.BaseFeeAmount != "1" || params.MaxFeeAmount != "1000" {
		t.Fatalf("expected capped low-fee defaults, got base=%q max=%q", params.BaseFeeAmount, params.MaxFeeAmount)
	}
	if params.FeeCollectorModule != FeeCollectorModuleName {
		t.Fatalf("expected fee collector %s, got %q", FeeCollectorModuleName, params.FeeCollectorModule)
	}
	if err := params.Validate(); err != nil {
		t.Fatalf("default params should validate: %v", err)
	}
}

func TestParamsRejectInvalidAllowedFeeDenoms(t *testing.T) {
	tests := map[string][]string{
		"empty list":       {},
		"non native denom": {"uatom"},
		"duplicate native": {appparams.BaseDenom, appparams.BaseDenom},
		"mixed denoms":     {appparams.BaseDenom, fixtureTestAssetDenom},
	}

	for name, denoms := range tests {
		t.Run(name, func(t *testing.T) {
			params := DefaultParams()
			params.AllowedFeeDenoms = denoms
			if err := params.Validate(); err == nil {
				t.Fatal("expected invalid allowed fee denoms to fail")
			}
		})
	}
}

func TestParamsRejectInvalidFeeSplitRatios(t *testing.T) {
	tests := map[string]func(*Params){
		"malformed validator ratio": func(params *Params) {
			params.ValidatorRewardsRatio = "not-a-decimal"
		},
		"malformed community ratio": func(params *Params) {
			params.CommunityPoolRatio = "not-a-decimal"
		},
		"negative ratio": func(params *Params) {
			params.ValidatorRewardsRatio = "-0.1"
			params.CommunityPoolRatio = "1.1"
		},
		"sum not one": func(params *Params) {
			params.ValidatorRewardsRatio = "0.80"
			params.CommunityPoolRatio = "0.10"
		},
		"ratio greater than one": func(params *Params) {
			params.ValidatorRewardsRatio = "1.01"
			params.CommunityPoolRatio = "-0.01"
		},
	}

	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			params := DefaultParams()
			mutate(&params)
			if err := params.Validate(); err == nil {
				t.Fatal("expected invalid fee split params to fail")
			}
		})
	}
}

func TestParamsRejectUnsafeProtocolFeeConfig(t *testing.T) {
	tests := map[string]func(*Params){
		"zero min fee": func(params *Params) {
			params.MinFeeAmount = "0"
		},
		"malformed min fee": func(params *Params) {
			params.MinFeeAmount = "not-an-int"
		},
		"min fee above v1 cap": func(params *Params) {
			params.MinFeeAmount = "1000000000000000001"
		},
		"unsafe fee collector": func(params *Params) {
			params.FeeCollectorModule = ModuleName
		},
		"unsafe validator target": func(params *Params) {
			params.ValidatorRewardsTarget = CommunityPoolTarget
		},
		"unsafe community target": func(params *Params) {
			params.CommunityPoolTarget = ValidatorRewardsTarget
		},
		"duplicate denom": func(params *Params) {
			params.AllowedFeeDenoms = []string{BondDenom, BondDenom}
		},
		"base below min": func(params *Params) {
			params.MinFeeAmount = "10"
			params.BaseFeeAmount = "9"
		},
		"max below base": func(params *Params) {
			params.BaseFeeAmount = "10"
			params.MaxFeeAmount = "9"
		},
		"target utilization too high": func(params *Params) {
			params.TargetBlockUtilizationBps = 10000
		},
		"congestion below target": func(params *Params) {
			params.CongestionThresholdBps = params.TargetBlockUtilizationBps
		},
		"tx gas above block gas": func(params *Params) {
			params.MaxTxGas = params.MaxBlockGas + 1
		},
		"sender stake limit below base": func(params *Params) {
			params.MaxSenderTxsPerBlockWithStake = params.MaxSenderTxsPerBlock - 1
		},
		"priority weights not normalized": func(params *Params) {
			params.FeePriorityWeightBps = 1
			params.StakePriorityWeightBps = 1
		},
	}

	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			params := DefaultParams()
			mutate(&params)
			if err := params.Validate(); err == nil {
				t.Fatal("expected invalid protocol fee params to fail")
			}
		})
	}
}

func TestDynamicFeeFormulaIsBoundedAndNonAuction(t *testing.T) {
	params := DefaultParams()
	base, err := params.BaseFeeInt()
	if err != nil {
		t.Fatalf("base fee should parse: %v", err)
	}
	maxFee, err := params.MaxFeeInt()
	if err != nil {
		t.Fatalf("max fee should parse: %v", err)
	}

	underTarget := DynamicFeeAmount(base, maxFee, params.TargetBlockUtilizationBps, params.TargetBlockUtilizationBps)
	if !underTarget.Equal(base) {
		t.Fatalf("expected target utilization to keep base fee, got %s", underTarget)
	}
	congested := DynamicFeeAmount(base, maxFee, params.TargetBlockUtilizationBps, 9000)
	if !congested.GT(base) {
		t.Fatalf("expected congestion to raise fee above base")
	}
	full := DynamicFeeAmount(base, maxFee, params.TargetBlockUtilizationBps, 10000)
	if !full.Equal(maxFee) {
		t.Fatalf("expected full block to hit hard cap %s, got %s", maxFee, full)
	}
}

func TestQuoteFeeIncludesEconomicControlSurface(t *testing.T) {
	params := DefaultParams()
	quote, err := QuoteFee(params, params.MaxTxGas, params.MaxBlockGas-params.MaxTxGas)
	if err != nil {
		t.Fatalf("quote should compute: %v", err)
	}
	if !quote.Congested || !quote.AtHardCap {
		t.Fatalf("expected congested hard-cap quote, got %+v", quote)
	}
	if quote.EconomicControl.BurnRatioBps != appparams.CongestedBurnRatioBps {
		t.Fatalf("expected congested burn ratio, got %+v", quote.EconomicControl)
	}
	if !quote.EconomicControl.RateLimited {
		t.Fatalf("expected rate-limit economic signal")
	}
	if quote.EconomicControl.ActivityInflationDeltaBps >= 0 {
		t.Fatalf("expected network activity to reduce inflation pressure, got %+v", quote.EconomicControl)
	}
}

func TestValidateAdmissionRejectsSpamWithoutUnboundedFeeEscalation(t *testing.T) {
	params := DefaultParams()
	quote, err := ValidateAdmission(params, AdmissionInput{
		Fee:              sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 1000)),
		GasLimit:         params.MaxTxGas,
		BlockGasConsumed: params.MaxBlockGas - params.MaxTxGas,
		BlockTxCount:     1,
		SenderTxCount:    1,
		SenderStake:      sdkmath.ZeroInt(),
	})
	if err != nil {
		t.Fatalf("max capped fee should be accepted at full utilization: %v", err)
	}
	if !quote.AtHardCap || quote.RequiredFee.Amount.Int64() != 1000 {
		t.Fatalf("expected hard cap quote, got %+v", quote)
	}

	_, err = ValidateAdmission(params, AdmissionInput{
		Fee:              sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 1001)),
		GasLimit:         100_000,
		BlockGasConsumed: 0,
		BlockTxCount:     1,
		SenderTxCount:    1,
		SenderStake:      sdkmath.ZeroInt(),
	})
	if err == nil {
		t.Fatal("expected over-cap fee to be rejected")
	}

	_, err = ValidateAdmission(params, AdmissionInput{
		Fee:              sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 1)),
		GasLimit:         params.MaxTxGas + 1,
		BlockGasConsumed: 0,
		BlockTxCount:     1,
		SenderTxCount:    1,
		SenderStake:      sdkmath.ZeroInt(),
	})
	if err == nil {
		t.Fatal("expected over-sized tx gas to be rejected")
	}

	_, err = ValidateAdmission(params, AdmissionInput{
		Fee:              sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 1)),
		GasLimit:         100_000,
		BlockGasConsumed: 0,
		BlockTxCount:     1,
		SenderTxCount:    params.MaxSenderTxsPerBlock,
		SenderStake:      sdkmath.ZeroInt(),
	})
	if err == nil {
		t.Fatal("expected sender rate limit to be rejected")
	}
}

func TestStakeWeightingIncreasesRateLimitAndPriorityWithoutFeeAuction(t *testing.T) {
	params := DefaultParams()
	baseLimit, err := SenderTxLimit(params, sdkmath.ZeroInt())
	if err != nil {
		t.Fatalf("base sender limit should compute: %v", err)
	}
	stakedLimit, err := SenderTxLimit(params, sdkmath.NewInt(10_000_000_000))
	if err != nil {
		t.Fatalf("staked sender limit should compute: %v", err)
	}
	if stakedLimit <= baseLimit {
		t.Fatalf("expected stake to increase sender allowance: base=%d staked=%d", baseLimit, stakedLimit)
	}

	required := sdk.NewInt64Coin(BondDenom, 10)
	scoreAtRequired, err := PriorityScore(params, required, required, sdkmath.ZeroInt())
	if err != nil {
		t.Fatalf("priority should compute: %v", err)
	}
	scoreOverpay, err := PriorityScore(params, sdk.NewInt64Coin(BondDenom, 1000), required, sdkmath.ZeroInt())
	if err != nil {
		t.Fatalf("priority should compute: %v", err)
	}
	if scoreOverpay != scoreAtRequired {
		t.Fatalf("overpay must not increase priority: required=%d overpay=%d", scoreAtRequired, scoreOverpay)
	}
}

func TestSplitFeesRoundingPreservesTotal(t *testing.T) {
	params := DefaultParams()
	validator, community, err := SplitFees(params, sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 1)))
	if err != nil {
		t.Fatalf("split fees should not fail: %v", err)
	}
	if !validator.Equal(sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 1))) {
		t.Fatalf("expected validator side to receive integer dust, got %s", validator)
	}
	if !community.Empty() {
		t.Fatalf("expected community side to truncate to zero, got %s", community)
	}
	if !validator.Add(community...).Equal(sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 1))) {
		t.Fatal("split must preserve total")
	}
}

func TestProtocolFeeStateValidateRejectsAccountingMismatch(t *testing.T) {
	state := DefaultProtocolFeeState()
	state.TotalCollected = sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 100))
	state.ValidatorRewards = sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 80))
	state.CommunityPool = sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 10))
	if err := state.Validate(); err == nil {
		t.Fatal("expected accounting mismatch to fail")
	}
}

func TestProtocolFeeStateValidateRejectsWrongDenom(t *testing.T) {
	state := DefaultProtocolFeeState()
	state.TotalCollected = sdk.NewCoins(sdk.NewInt64Coin("uatom", 100))
	state.ValidatorRewards = sdk.NewCoins(sdk.NewInt64Coin("uatom", 100))
	if err := state.Validate(); err == nil {
		t.Fatal("expected wrong accounting denom to fail")
	}
}

func TestGenesisValidateIncludesProtocolFeeState(t *testing.T) {
	gs := DefaultGenesisState()
	if err := gs.Validate(); err != nil {
		t.Fatalf("default genesis should validate: %v", err)
	}

	gs.ProtocolFeeState.TotalCollected = sdk.NewCoins(sdk.NewInt64Coin(BondDenom, 10))
	if err := gs.Validate(); err == nil {
		t.Fatal("expected malformed protocol fee state to fail")
	}
}
