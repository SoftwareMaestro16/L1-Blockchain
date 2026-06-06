package params

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultEconomicOpenDesignDecisionsAreExplicitlyTracked(t *testing.T) {
	report := BuildEconomicOpenDesignDecisionReport(nil)
	require.True(t, report.Passed, report.Failed)
	require.Len(t, report.Decisions, 7)
	require.Equal(t, 7, report.Required)
	require.Equal(t, 7, report.Covered)
	require.Equal(t, BasisPoints, report.CoverageBps)
	require.Contains(t, report.Summary, "open_design_decisions=7/7")

	ids := make(map[string]struct{}, len(report.Decisions))
	for _, decision := range report.Decisions {
		ids[decision.ID] = struct{}{}
		require.Equal(t, OpenDesignDecisionStatusOpen, decision.Status)
		require.Empty(t, decision.Resolution)
		require.NotEmpty(t, decision.Question)
		require.Len(t, decision.Options, 3)
		require.NotEmpty(t, decision.LinkedGovernanceParams)
		require.NotEmpty(t, decision.RequiredBeforeActivation)
		require.True(t, decision.Queryable)
		require.True(t, decision.GovernanceRequired)
		require.True(t, decision.SimulationRequired)
		require.True(t, decision.ImplementationMustNotPick)
	}

	require.Contains(t, ids, OpenDesignDecisionRewardDampeningScope)
	require.Contains(t, ids, OpenDesignDecisionStateRentActivation)
	require.Contains(t, ids, OpenDesignDecisionDeleteRefundPayment)
	require.Contains(t, ids, OpenDesignDecisionBootstrapFunding)
	require.Contains(t, ids, OpenDesignDecisionRiskYieldLocation)
	require.Contains(t, ids, OpenDesignDecisionFeeBucketWeights)
	require.Contains(t, ids, OpenDesignDecisionSecurityReserveSpend)
}

func TestEconomicOpenDesignDecisionReportRejectsMissingAndDuplicateDecisions(t *testing.T) {
	decisions := DefaultEconomicOpenDesignDecisions()
	decisions = append(decisions[:1], decisions[2:]...)
	decisions = append(decisions, decisions[0])

	report := BuildEconomicOpenDesignDecisionReport(decisions)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, OpenDesignDecisionStateRentActivation+":missing_required_open_design_decision")
	require.Contains(t, report.Failed, OpenDesignDecisionRewardDampeningScope+":duplicate_open_design_decision")
	require.Less(t, report.CoverageBps, BasisPoints)
}

func TestEconomicOpenDesignDecisionReportRejectsPrematureResolution(t *testing.T) {
	decisions := DefaultEconomicOpenDesignDecisions()
	for i := range decisions {
		if decisions[i].ID == OpenDesignDecisionFeeBucketWeights {
			decisions[i].Status = "resolved"
			decisions[i].Resolution = "controller_adjusted_within_bounds"
		}
	}

	report := BuildEconomicOpenDesignDecisionReport(decisions)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, OpenDesignDecisionFeeBucketWeights+":status_must_remain_open")
	require.Contains(t, report.Failed, OpenDesignDecisionFeeBucketWeights+":resolution_must_be_empty_until_governance_decision")
}

func TestEconomicOpenDesignDecisionReportRequiresGovernanceLinksAndActivationGate(t *testing.T) {
	decisions := DefaultEconomicOpenDesignDecisions()
	for i := range decisions {
		switch decisions[i].ID {
		case OpenDesignDecisionStateRentActivation:
			decisions[i].LinkedGovernanceParams = nil
		case OpenDesignDecisionRiskYieldLocation:
			decisions[i].LinkedGovernanceParams = []string{"unknown_param"}
		case OpenDesignDecisionSecurityReserveSpend:
			decisions[i].RequiredBeforeActivation = ""
		}
	}

	report := BuildEconomicOpenDesignDecisionReport(decisions)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, OpenDesignDecisionStateRentActivation+":governance_param_link_missing")
	require.Contains(t, report.Failed, OpenDesignDecisionRiskYieldLocation+":unknown_linked_governance_param:unknown_param")
	require.Contains(t, report.Failed, OpenDesignDecisionSecurityReserveSpend+":activation_gate_missing")
	require.Less(t, report.CoverageBps, BasisPoints)
}

func TestEconomicOpenDesignDecisionReportRequiresQueryableGovernedSimulatedDecisions(t *testing.T) {
	decisions := DefaultEconomicOpenDesignDecisions()
	for i := range decisions {
		switch decisions[i].ID {
		case OpenDesignDecisionRewardDampeningScope:
			decisions[i].Queryable = false
		case OpenDesignDecisionBootstrapFunding:
			decisions[i].GovernanceRequired = false
		case OpenDesignDecisionDeleteRefundPayment:
			decisions[i].SimulationRequired = false
		case OpenDesignDecisionStateRentActivation:
			decisions[i].ImplementationMustNotPick = false
		}
	}

	report := BuildEconomicOpenDesignDecisionReport(decisions)
	require.False(t, report.Passed)
	require.Contains(t, report.Failed, OpenDesignDecisionRewardDampeningScope+":not_queryable")
	require.Contains(t, report.Failed, OpenDesignDecisionBootstrapFunding+":governance_required_missing")
	require.Contains(t, report.Failed, OpenDesignDecisionDeleteRefundPayment+":simulation_required_missing")
	require.Contains(t, report.Failed, OpenDesignDecisionStateRentActivation+":implementation_must_not_pick_missing")
}
