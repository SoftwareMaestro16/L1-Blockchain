package params

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultEconomicGovernanceSurfaceIsComplete(t *testing.T) {
	report := BuildEconomicGovernanceSurfaceReport(nil)
	require.True(t, report.Passed, report.Failed)
	require.Len(t, report.Parameters, 46)
	require.Equal(t, 8, report.RequiredInflation)
	require.Equal(t, 5, report.RequiredBurn)
	require.Equal(t, 8, report.RequiredFee)
	require.Equal(t, 10, report.RequiredValidator)
	require.Equal(t, 8, report.RequiredStorage)
	require.Equal(t, 7, report.RequiredSecurity)
	require.Equal(t, 8, report.CoveredInflation)
	require.Equal(t, 5, report.CoveredBurn)
	require.Equal(t, 8, report.CoveredFee)
	require.Equal(t, 10, report.CoveredValidator)
	require.Equal(t, 8, report.CoveredStorage)
	require.Equal(t, 7, report.CoveredSecurity)
	require.Equal(t, BasisPoints, report.InflationCoverageBps)
	require.Equal(t, BasisPoints, report.BurnCoverageBps)
	require.Equal(t, BasisPoints, report.FeeCoverageBps)
	require.Equal(t, BasisPoints, report.ValidatorCoverageBps)
	require.Equal(t, BasisPoints, report.StorageCoverageBps)
	require.Equal(t, BasisPoints, report.SecurityCoverageBps)
	require.Contains(t, report.GovernanceSummary, "inflation=8/8")
	require.Contains(t, report.GovernanceSummary, "burn=5/5")
	require.Contains(t, report.GovernanceSummary, "fee=8/8")
	require.Contains(t, report.GovernanceSummary, "validator=10/10")
	require.Contains(t, report.GovernanceSummary, "storage=8/8")
	require.Contains(t, report.GovernanceSummary, "security=7/7")

	for _, param := range report.Parameters {
		require.True(t, param.Required)
		require.True(t, param.Queryable)
		require.True(t, param.Bounded)
		require.True(t, param.ImpactReport)
		require.True(t, param.ChangeControlled)
		require.NotEmpty(t, param.Source)
		require.NotEmpty(t, param.Unit)
	}
}

func TestEconomicGovernanceSurfaceRejectsMissingAndDuplicateParameter(t *testing.T) {
	params := DefaultEconomicGovernanceParameters()
	params = append(params[:1], params[2:]...)
	params = append(params, params[0])

	report := BuildEconomicGovernanceSurfaceReport(params)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, GovernanceParamTargetInflation+":missing_required_governance_parameter")
	require.Contains(t, report.Failed, GovernanceParamMinimumInflation+":duplicate_governance_parameter")
	require.Less(t, report.InflationCoverageBps, BasisPoints)
}

func TestEconomicGovernanceSurfaceRequiresQueryableBoundedImpactControlledParams(t *testing.T) {
	params := DefaultEconomicGovernanceParameters()
	for i := range params {
		switch params[i].ID {
		case GovernanceParamFeeBurnAllocation:
			params[i].Queryable = false
		case GovernanceParamSlashingBurnAllocation:
			params[i].Bounded = false
		case GovernanceParamBurnCapPerEpoch:
			params[i].ImpactReport = false
		case GovernanceParamBurnActivationThreshold:
			params[i].ChangeControlled = false
		}
	}

	report := BuildEconomicGovernanceSurfaceReport(params)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, GovernanceParamFeeBurnAllocation+":not_queryable")
	require.Contains(t, report.Failed, GovernanceParamSlashingBurnAllocation+":not_bounded")
	require.Contains(t, report.Failed, GovernanceParamBurnCapPerEpoch+":impact_report_missing")
	require.Contains(t, report.Failed, GovernanceParamBurnActivationThreshold+":change_control_missing")
	require.Less(t, report.BurnCoverageBps, BasisPoints)
}

func TestEconomicGovernanceSurfaceRejectsInvalidBoundsAndMetadata(t *testing.T) {
	params := DefaultEconomicGovernanceParameters()
	for i := range params {
		switch params[i].ID {
		case GovernanceParamMinimumBaseFee:
			params[i].Source = ""
		case GovernanceParamMaximumBaseFee:
			params[i].Unit = ""
		case GovernanceParamTargetBlockUtilization:
			params[i].MinBps = 9_000
			params[i].MaxBps = 7_000
		case GovernanceParamMaxFeeAdjustmentPerWindow:
			params[i].MinBps = -1
		}
	}

	report := BuildEconomicGovernanceSurfaceReport(params)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, GovernanceParamMinimumBaseFee+":source_missing")
	require.Contains(t, report.Failed, GovernanceParamMaximumBaseFee+":unit_missing")
	require.Contains(t, report.Failed, GovernanceParamTargetBlockUtilization+":invalid_bound_order")
	require.Contains(t, report.Failed, GovernanceParamMaxFeeAdjustmentPerWindow+":negative_bounds")
	require.Less(t, report.FeeCoverageBps, BasisPoints)
}

func TestEconomicGovernanceSurfaceRejectsUnknownCategoryAndBlankID(t *testing.T) {
	params := DefaultEconomicGovernanceParameters()
	params = append(params,
		EconomicGovernanceParameter{ID: "unknown", Category: "other", Required: true, Queryable: true, Bounded: true, ImpactReport: true, ChangeControlled: true, Source: "x", Unit: "bps"},
		EconomicGovernanceParameter{Category: EconomicGovernanceCategoryInflation, Required: true},
	)

	report := BuildEconomicGovernanceSurfaceReport(params)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, "unknown:unknown_governance_category")
	require.Contains(t, report.Failed, "governance_parameter_id_required")
}

func TestEconomicGovernanceSurfaceCoversValidatorStorageAndSecurityParameters(t *testing.T) {
	params := DefaultEconomicGovernanceParameters()
	filtered := make([]EconomicGovernanceParameter, 0, len(params)-3)
	for _, param := range params {
		switch param.ID {
		case GovernanceParamEpochLength, GovernanceParamRentRate, GovernanceParamCircuitBreakerThresholds:
			continue
		default:
			filtered = append(filtered, param)
		}
	}

	report := BuildEconomicGovernanceSurfaceReport(filtered)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, GovernanceParamEpochLength+":missing_required_governance_parameter")
	require.Contains(t, report.Failed, GovernanceParamRentRate+":missing_required_governance_parameter")
	require.Contains(t, report.Failed, GovernanceParamCircuitBreakerThresholds+":missing_required_governance_parameter")
	require.Less(t, report.ValidatorCoverageBps, BasisPoints)
	require.Less(t, report.StorageCoverageBps, BasisPoints)
	require.Less(t, report.SecurityCoverageBps, BasisPoints)
}
