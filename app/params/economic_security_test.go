package params

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestEconomicSecurityRoutesPenaltyAndRequestsReserveFunding(t *testing.T) {
	params := DefaultEconomicSecurityParams()
	params.SecurityReserveMinimumNaet = sdkmath.NewInt(1_000)
	params.SecurityReserveTargetNaet = sdkmath.NewInt(3_000)
	params.MaxReserveFundingRequestNaet = sdkmath.NewInt(5_000)
	params.ReporterRewardCapNaet = sdkmath.NewInt(100)

	report, err := EvaluateEconomicSecurityEpoch(baseEconomicSecurityInput(EconomicSecurityEpochInput{
		SlashingPenaltyNaet:		sdkmath.NewInt(1_000),
		EvidenceAccepted:		true,
		RequestedReporterRewardNaet:	sdkmath.NewInt(150),
		SecurityReserveBalanceNaet:	sdkmath.NewInt(500),
	}), params)
	require.NoError(t, err)
	require.True(t, report.Passed, report.Failed)
	require.Equal(t, sdkmath.NewInt(1_000), report.PenaltyRouting.PenaltyNaet)
	require.Equal(t, sdkmath.NewInt(200), report.PenaltyRouting.BurnNaet)
	require.Equal(t, sdkmath.NewInt(300), report.PenaltyRouting.TreasuryNaet)
	require.Equal(t, sdkmath.NewInt(100), report.PenaltyRouting.ReporterRewardNaet)
	require.Equal(t, sdkmath.NewInt(400), report.PenaltyRouting.ValidatorPoolNaet)
	require.Equal(t, sdkmath.NewInt(2_500), report.ReserveAccounting.FundingRequestNaet)
	require.Contains(t, report.GovernanceReport.AlertTypes, EconomicSecurityAlertReserveLow)
	require.Contains(t, report.GovernanceReport.ActiveRestrictions, "security_reserve_refill_required")
	require.NotEmpty(t, report.AuditLogs)
}

func TestEconomicSecurityRejectsDuplicateEvidenceReporterReward(t *testing.T) {
	report, err := EvaluateEconomicSecurityEpoch(baseEconomicSecurityInput(EconomicSecurityEpochInput{
		SlashingPenaltyNaet:		sdkmath.NewInt(1_000),
		EvidenceAccepted:		true,
		EvidenceDuplicate:		true,
		RequestedReporterRewardNaet:	sdkmath.NewInt(500),
	}), DefaultEconomicSecurityParams())
	require.NoError(t, err)
	require.True(t, report.Passed, report.Failed)
	require.True(t, report.PenaltyRouting.ReporterRewardNaet.IsZero())
	require.Contains(t, report.GovernanceReport.AlertTypes, EconomicSecurityAlertEvidenceDuplicate)
}

func TestEconomicSecurityDetectsReserveAccountingInvariantViolation(t *testing.T) {
	report, err := EvaluateEconomicSecurityEpoch(baseEconomicSecurityInput(EconomicSecurityEpochInput{
		SecurityReserveBalanceNaet:	sdkmath.NewInt(100),
		SecurityReserveOutflowNaet:	sdkmath.NewInt(1_000),
	}), DefaultEconomicSecurityParams())
	require.NoError(t, err)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, "security_reserve_accounting_negative")
	require.Contains(t, incidentReasons(report.InvariantEvents), "security_reserve_accounting_negative")
}

func TestEconomicSecurityCircuitBreakerRulesAreDeterministic(t *testing.T) {
	report, err := EvaluateEconomicSecurityEpoch(baseEconomicSecurityInput(EconomicSecurityEpochInput{
		BlockLoadBps:		HighCongestionLoadBps,
		FeeSpikeBps:		DefaultCircuitBreakerFeeSpikeBps + 1,
		ControllerDriftBps:	DefaultCircuitBreakerControllerDriftBps + 1,
		FailedTxRateBps:	DefaultCircuitBreakerFailedTxRateBps + 1,
		BurnToMintBps:		DeflationGuardBurnToMintBps + 1,
	}), DefaultEconomicSecurityParams())
	require.NoError(t, err)
	require.True(t, report.Passed, report.Failed)
	require.True(t, report.CircuitBreaker.Active)
	require.Contains(t, report.CircuitBreaker.Reasons, "fee_spike_abnormal")
	require.Contains(t, report.CircuitBreaker.Reasons, "controller_drift_abnormal")
	require.Contains(t, report.GovernanceReport.AlertTypes, EconomicSecurityAlertCircuitBreaker)
	for _, audit := range report.AuditLogs {
		require.True(t, audit.Deterministic)
	}
}

func TestEconomicSecurityAlertsForStateStakeAndConcentration(t *testing.T) {
	params := DefaultEconomicSecurityParams()
	params.MaxStateGrowthBytes = 1_000
	params.MaxStakeMovementBps = 500
	report, err := EvaluateEconomicSecurityEpoch(baseEconomicSecurityInput(EconomicSecurityEpochInput{
		ValidatorConcentrationBps:	params.MaxValidatorConcentrationBps + 1,
		TopNConcentrationBps:		params.MaxTopNConcentrationBps + 1,
		StateGrowthBytes:		2_000,
		StakeInflowBps:			700,
	}), params)
	require.NoError(t, err)
	require.True(t, report.Passed, report.Failed)
	require.Contains(t, report.GovernanceReport.AlertTypes, EconomicSecurityAlertConcentration)
	require.Contains(t, report.GovernanceReport.AlertTypes, EconomicSecurityAlertTopNConcentration)
	require.Contains(t, report.GovernanceReport.AlertTypes, EconomicSecurityAlertStateGrowth)
	require.Contains(t, report.GovernanceReport.AlertTypes, EconomicSecurityAlertStakeMovement)
	require.Contains(t, incidentTypes(report.InvariantEvents), EconomicSecurityAlertStateGrowth)
}

func TestEconomicSecurityRejectsMisconfiguredRoutingRatios(t *testing.T) {
	params := DefaultEconomicSecurityParams()
	params.SlashingBurnRatioBps = 7_000
	params.SlashingTreasuryRatioBps = 4_000
	_, err := EvaluateEconomicSecurityEpoch(baseEconomicSecurityInput(EconomicSecurityEpochInput{}), params)
	require.ErrorContains(t, err, "slashing burn and treasury ratios exceed 100%")
}

func baseEconomicSecurityInput(override EconomicSecurityEpochInput) EconomicSecurityEpochInput {
	input := EconomicSecurityEpochInput{
		EpochID:			7,
		BlockHeight:			700,
		ValidatorConcentrationBps:	MaxTopValidatorConcentrationBps,
		TopNConcentrationBps:		DefaultTopNConcentrationThresholdBps,
		BlockLoadBps:			DefaultTargetLoadBps,
		BurnToMintBps:			BasisPoints,
		SecurityReserveBalanceNaet:	sdkmath.NewInt(DefaultSecurityReserveTargetNaet),
		GovernanceThresholdVersion:	DefaultSecurityGovernanceVersion,
	}
	if override.EpochID != 0 {
		input.EpochID = override.EpochID
	}
	if override.BlockHeight != 0 {
		input.BlockHeight = override.BlockHeight
	}
	if !override.SlashingPenaltyNaet.IsNil() {
		input.SlashingPenaltyNaet = override.SlashingPenaltyNaet
	}
	input.EvidenceAccepted = override.EvidenceAccepted
	input.EvidenceDuplicate = override.EvidenceDuplicate
	if !override.RequestedReporterRewardNaet.IsNil() {
		input.RequestedReporterRewardNaet = override.RequestedReporterRewardNaet
	}
	if override.ValidatorConcentrationBps != 0 {
		input.ValidatorConcentrationBps = override.ValidatorConcentrationBps
	}
	if override.TopNConcentrationBps != 0 {
		input.TopNConcentrationBps = override.TopNConcentrationBps
	}
	if override.BlockLoadBps != 0 {
		input.BlockLoadBps = override.BlockLoadBps
	}
	if override.FeeSpikeBps != 0 {
		input.FeeSpikeBps = override.FeeSpikeBps
	}
	if override.ControllerDriftBps != 0 {
		input.ControllerDriftBps = override.ControllerDriftBps
	}
	if override.FailedTxRateBps != 0 {
		input.FailedTxRateBps = override.FailedTxRateBps
	}
	if override.BurnToMintBps != 0 {
		input.BurnToMintBps = override.BurnToMintBps
	}
	if override.StateGrowthBytes != 0 {
		input.StateGrowthBytes = override.StateGrowthBytes
	}
	if override.StakeInflowBps != 0 {
		input.StakeInflowBps = override.StakeInflowBps
	}
	if override.StakeOutflowBps != 0 {
		input.StakeOutflowBps = override.StakeOutflowBps
	}
	if !override.SecurityReserveBalanceNaet.IsNil() {
		input.SecurityReserveBalanceNaet = override.SecurityReserveBalanceNaet
	}
	if !override.SecurityReserveInflowNaet.IsNil() {
		input.SecurityReserveInflowNaet = override.SecurityReserveInflowNaet
	}
	if !override.SecurityReserveOutflowNaet.IsNil() {
		input.SecurityReserveOutflowNaet = override.SecurityReserveOutflowNaet
	}
	if override.GovernanceThresholdVersion != "" {
		input.GovernanceThresholdVersion = override.GovernanceThresholdVersion
	}
	return input
}

func incidentReasons(events []EconomicSecurityIncidentEvent) []string {
	reasons := make([]string, 0, len(events))
	for _, event := range events {
		reasons = append(reasons, event.Reason)
	}
	return reasons
}

func incidentTypes(events []EconomicSecurityIncidentEvent) []string {
	types := make([]string, 0, len(events))
	for _, event := range events {
		types = append(types, event.Type)
	}
	return types
}
