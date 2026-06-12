package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestFeeDistributionBucketsSumExactlyAndEmitEvents(t *testing.T) {
	collected := sdk.NewInt64Coin(BondDenom, 10_003)
	result, err := DistributeFeeBuckets(collected, DefaultFeeDistributionParams(), CongestionSignals{
		BlockGasUtilizationBps: 5_000,
	}, 1_000)
	if err != nil {
		t.Fatalf("fee buckets should distribute: %v", err)
	}
	if !sumFeeBuckets(result.Buckets).Equal(collected.Amount) {
		t.Fatalf("fee bucket allocation must exactly sum to collected fees: %+v", result)
	}
	if !result.ValidatorDelegatorRewards.Amount.Equal(sdkmath.NewInt(8_003)) {
		t.Fatalf("integer dust should stay with validator rewards, got %s", result.ValidatorDelegatorRewards)
	}
	if !result.Burn.Amount.Equal(sdkmath.NewInt(500)) || result.BurnAllocationBps != 500 {
		t.Fatalf("burn allocation must be queryable, got %+v", result)
	}
	if !result.StateMaintenanceReserve.Amount.Equal(sdkmath.NewInt(300)) || result.StateMaintenanceBps != 300 {
		t.Fatalf("state reserve allocation must be queryable, got %+v", result)
	}
	if !result.SecurityReserve.Amount.Equal(sdkmath.NewInt(200)) || result.SecurityReserveBps != 200 {
		t.Fatalf("security reserve allocation must be queryable, got %+v", result)
	}
	if len(result.Events) != 5 {
		t.Fatalf("expected machine-readable accounting events for all buckets, got %d", len(result.Events))
	}
	for _, event := range result.Events {
		if event.Type != "fee_allocation" || event.Denom != BondDenom || event.Amount.IsNegative() {
			t.Fatalf("invalid accounting event: %+v", event)
		}
	}
}

func TestFeeDistributionRaisesBurnAndStateReserveButPreservesValidatorFloor(t *testing.T) {
	params := DefaultFeeDistributionParams()
	result, err := DistributeFeeBuckets(sdk.NewInt64Coin(BondDenom, 10_000), params, CongestionSignals{
		BlockGasUtilizationBps:	10_000,
		MempoolPressureBps:	10_000,
		FailedExecutionRateBps:	10_000,
	}, 10_000)
	if err != nil {
		t.Fatalf("fee buckets should distribute under high activity: %v", err)
	}
	if result.ValidatorDelegatorRewards.Amount.LT(sdkmath.NewInt(6_000)) {
		t.Fatalf("validator floor must be preserved, got %s", result.ValidatorDelegatorRewards)
	}
	if !result.ValidatorRewardFloorEnforced {
		t.Fatalf("expected validator reward floor enforcement")
	}
	if result.BurnAllocationBps <= params.BurnBaseBps {
		t.Fatalf("expected sustained high activity to increase burn allocation")
	}
	if result.StateMaintenanceBps <= params.StateMaintenanceBaseBps {
		t.Fatalf("expected state growth to increase state maintenance allocation")
	}
	if !sumFeeBuckets(result.Buckets).Equal(result.Collected.Amount) {
		t.Fatalf("high activity fee buckets must sum exactly")
	}
}

func TestAntiSpamScoringAndSurchargeAreDeterministicBoundedAndIncreasing(t *testing.T) {
	params := DefaultAntiSpamParams(DefaultParams())
	normal := AccountActivityWindow{Sender: "alice", WindowID: 1, TxCount: 2, FailedTxCount: 0, StateWriteCount: 1, DeploymentCount: 0}
	normalSurcharge, normalScore, err := AntiSpamSurchargeBps(normal, params)
	if err != nil {
		t.Fatalf("normal activity should score: %v", err)
	}
	if normalSurcharge != 0 || normalScore.ScoreBps != 0 {
		t.Fatalf("normal user should not receive surcharge, got score=%+v surcharge=%d", normalScore, normalSurcharge)
	}
	spam := AccountActivityWindow{Sender: "alice", WindowID: 1, TxCount: 100, FailedTxCount: 20, StateWriteCount: 40, DeploymentCount: 8}
	left, leftScore, err := AntiSpamSurchargeBps(spam, params)
	if err != nil {
		t.Fatalf("spam activity should score: %v", err)
	}
	right, rightScore, err := AntiSpamSurchargeBps(spam, params)
	if err != nil {
		t.Fatalf("spam activity should score twice: %v", err)
	}
	if left != right || leftScore.ScoreBps != rightScore.ScoreBps {
		t.Fatalf("anti-spam scoring must be deterministic")
	}
	if left != params.MaxTotalSurchargeBps || !leftScore.Bounded {
		t.Fatalf("spam surcharge must be bounded at cap, got score=%+v surcharge=%d", leftScore, left)
	}
}

func TestAntiSpamAdmissionUsesCompatibleMempoolAndExecutionRules(t *testing.T) {
	params := DefaultAntiSpamParams(DefaultParams())
	activity := AccountActivityWindow{Sender: "sender-a", WindowID: 7, TxCount: 30, FailedTxCount: 6, StateWriteCount: 10, DeploymentCount: 1}
	base := sdk.NewInt64Coin(BondDenom, 100)
	paid := sdk.NewInt64Coin(BondDenom, 250)
	mempool, err := ValidateAntiSpamFee(AntiSpamFeeInput{
		Mode:			AdmissionModeMempool,
		BaseRequiredFee:	base,
		PaidFee:		paid,
		Activity:		activity,
		Params:			params,
		Executable:		true,
	})
	if err != nil {
		t.Fatalf("mempool anti-spam fee should validate: %v", err)
	}
	execution, err := ValidateAntiSpamFee(AntiSpamFeeInput{
		Mode:			AdmissionModeBlockExecution,
		BaseRequiredFee:	base,
		PaidFee:		paid,
		Activity:		activity,
		Params:			params,
		Executable:		true,
	})
	if err != nil {
		t.Fatalf("execution anti-spam fee should validate: %v", err)
	}
	if !mempool.RequiredFee.Equal(execution.RequiredFee) || mempool.SurchargeBps != execution.SurchargeBps {
		t.Fatalf("mempool and block execution must use compatible final fee rules: mempool=%+v execution=%+v", mempool, execution)
	}
	if !mempool.Accepted || !execution.Accepted {
		t.Fatalf("paid fee should satisfy anti-spam requirement")
	}
}

func TestAntiSpamFeeBypassAttemptsAreRejected(t *testing.T) {
	params := DefaultAntiSpamParams(DefaultParams())
	activity := AccountActivityWindow{Sender: "sender-b", WindowID: 8, TxCount: 2, FailedTxCount: 10, StateWriteCount: 1, DeploymentCount: 0}
	decision, err := ValidateAntiSpamFee(AntiSpamFeeInput{
		Mode:			AdmissionModeMempool,
		BaseRequiredFee:	sdk.NewInt64Coin(BondDenom, 100),
		PaidFee:		sdk.NewInt64Coin(BondDenom, 100),
		Activity:		activity,
		Params:			params,
		Executable:		true,
	})
	if err != nil {
		t.Fatalf("underpaid bypass attempt should return decision, not panic: %v", err)
	}
	if decision.Accepted || decision.Reason != "fee_below_anti_spam_requirement" {
		t.Fatalf("underpaid spam should be rejected, got %+v", decision)
	}
	_, err = ValidateAntiSpamFee(AntiSpamFeeInput{
		Mode:			AdmissionModeMempool,
		BaseRequiredFee:	sdk.NewInt64Coin("uatom", 100),
		PaidFee:		sdk.NewInt64Coin(BondDenom, 1_000),
		Activity:		activity,
		Params:			params,
		Executable:		true,
	})
	if err == nil {
		t.Fatalf("wrong-denom base fee bypass should fail")
	}
	_, err = ValidateAntiSpamFee(AntiSpamFeeInput{
		Mode:			AdmissionModeMempool,
		BaseRequiredFee:	sdk.NewInt64Coin(BondDenom, 100),
		PaidFee:		sdk.NewInt64Coin("uatom", 1_000),
		Activity:		activity,
		Params:			params,
		Executable:		true,
	})
	if err == nil {
		t.Fatalf("wrong-denom paid fee bypass should fail")
	}
}
