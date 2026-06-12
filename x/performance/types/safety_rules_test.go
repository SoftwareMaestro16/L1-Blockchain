package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConsensusDeterminismPolicyRejectsUnsafeSources(t *testing.T) {
	policy, err := BuildConsensusDeterminismPolicy(ConsensusDeterminismPolicy{
		ConsensusPathName:		"finalize_block",
		UsesConsensusTime:		true,
		SortsStateTransitionMaps:	true,
		EncodesProposalMempoolInputs:	true,
	})
	require.NoError(t, err)
	require.NoError(t, policy.Validate())

	external := policy
	external.UsesExternalAPIs = true
	external.DeterminismPolicyHash = ComputeConsensusDeterminismPolicyHash(external)
	require.ErrorContains(t, external.Validate(), "external APIs")

	localClock := policy
	localClock.UsesLocalClock = true
	localClock.UsesConsensusTime = false
	localClock.DeterminismPolicyHash = ComputeConsensusDeterminismPolicyHash(localClock)
	require.ErrorContains(t, localClock.Validate(), "local clock")

	randomShard := policy
	randomShard.UsesRandomShardPlacement = true
	randomShard.DeterminismPolicyHash = ComputeConsensusDeterminismPolicyHash(randomShard)
	require.ErrorContains(t, randomShard.Validate(), "random")

	floatMath := policy
	floatMath.UsesFloatingPointMath = true
	floatMath.DeterminismPolicyHash = ComputeConsensusDeterminismPolicyHash(floatMath)
	require.ErrorContains(t, floatMath.Validate(), "floating point")
}

func TestRoutingSafetyRequiresCommittedInputsAndDeterministicTieBreak(t *testing.T) {
	best := routingSafetyPath("path-best", "10", 2, 100, "a")
	worse := routingSafetyPath("path-worse", "10", 2, 500, "b")
	input, err := BuildRoutingSafetyInput(RoutingSafetyInput{
		RoutingEpoch:		3,
		RoutingTableHash:	hashStrings("routing-table"),
		RoutingMetricsRoot:	hashStrings("routing-metrics"),
		CandidatePaths:		[]RoutingSafetyPath{worse, best},
		SelectedPathID:		"path-best",
	})
	require.NoError(t, err)
	require.NoError(t, input.Validate())
	require.Equal(t, "path-best", input.CandidatePaths[0].PathID)

	missingMetrics := input
	missingMetrics.RoutingMetricsRoot = ""
	missingMetrics.SafetyHash = ComputeRoutingSafetyHash(missingMetrics)
	require.ErrorContains(t, missingMetrics.Validate(), "metrics")

	wrongSelection := input
	wrongSelection.SelectedPathID = "path-worse"
	wrongSelection.SafetyHash = ComputeRoutingSafetyHash(wrongSelection)
	require.ErrorContains(t, wrongSelection.Validate(), "best path")
}

func TestRoutingFailureAccountingPreservesValueAndRequiresReceiptForBurn(t *testing.T) {
	accounting := routingFailureAccounting("route-a", "100", "95", "5", hashStrings("receipt-a"), "bounced")
	require.NoError(t, accounting.Validate())

	noReceipt := routingFailureAccounting("route-b", "100", "95", "5", "", "bounced")
	require.ErrorContains(t, noReceipt.Validate(), "without receipt")

	mintingBounce := routingFailureAccounting("route-c", "100", "101", "0", hashStrings("receipt-c"), "bounced")
	require.ErrorContains(t, mintingBounce.Validate(), "extra value")
}

func TestShardLayoutTransitionRequiresEpochBoundaryMigrationRootAndDeliveryEpoch(t *testing.T) {
	msg := inFlightMessage("message-a", 6, 120)
	transition, err := BuildShardLayoutTransition(ShardLayoutTransition{
		ZoneID:			"financial",
		PreviousLayoutEpoch:	5,
		NextLayoutEpoch:	6,
		CurrentHeight:		100,
		ActivationHeight:	110,
		EpochBoundaryHeight:	110,
		SplitMergeDecisionHash:	hashStrings("split-merge-decision"),
		MigrationRoot:		hashStrings("migration-root"),
		OldLayoutHash:		hashStrings("old-layout"),
		NewLayoutHash:		hashStrings("new-layout"),
		ProofHorizon:		16,
		InFlightMessages:	[]InFlightShardMessage{msg},
	})
	require.NoError(t, err)
	require.NoError(t, transition.Validate())
	require.True(t, OldShardLayoutQueryable(126, transition.ActivationHeight, transition.ProofHorizon))
	require.False(t, OldShardLayoutQueryable(127, transition.ActivationHeight, transition.ProofHorizon))

	notBoundary := transition
	notBoundary.ActivationHeight = 111
	notBoundary.TransitionHash = ComputeShardLayoutTransitionHash(notBoundary)
	require.ErrorContains(t, notBoundary.Validate(), "epoch boundaries")

	wrongEpoch := transition
	wrongEpoch.InFlightMessages = append([]InFlightShardMessage{}, transition.InFlightMessages...)
	wrongEpoch.InFlightMessages[0].DeliveryEpoch = 5
	wrongEpoch.InFlightMessages[0].MessageHash = ComputeInFlightShardMessageHash(wrongEpoch.InFlightMessages[0])
	wrongEpoch.TransitionHash = ComputeShardLayoutTransitionHash(wrongEpoch)
	require.ErrorContains(t, wrongEpoch.Validate(), "delivery epoch")

	missingMigration := transition
	missingMigration.MigrationRoot = ""
	missingMigration.TransitionHash = ComputeShardLayoutTransitionHash(missingMigration)
	require.ErrorContains(t, missingMigration.Validate(), "migration root")
}

func routingSafetyPath(id, cost string, hops uint32, congestion uint32, tie string) RoutingSafetyPath {
	path := RoutingSafetyPath{
		PathID:		id,
		RouteHash:	hashStrings("route", id),
		Cost:		cost,
		HopCount:	hops,
		LiquidityBps:	9_000,
		CongestionBps:	congestion,
		TieBreakKey:	tie,
	}
	return path.Normalize()
}

func routingFailureAccounting(routeID, original, bounced, burned, receiptHash, status string) RoutingFailureAccounting {
	accounting := RoutingFailureAccounting{
		RouteID:	routeID,
		OriginalValue:	original,
		BouncedValue:	bounced,
		BurnedValue:	burned,
		ReceiptHash:	receiptHash,
		ReceiptStatus:	status,
	}
	accounting.AccountingHash = ComputeRoutingFailureAccountingHash(accounting)
	return accounting
}

func inFlightMessage(seed string, deliveryEpoch uint64, expiryHeight uint64) InFlightShardMessage {
	msg := InFlightShardMessage{
		MessageID:		hashStrings("in-flight", seed),
		SourceShardID:		"shard-a",
		DestinationShardID:	"shard-b",
		DeliveryEpoch:		deliveryEpoch,
		ExpiryHeight:		expiryHeight,
	}
	msg.MessageHash = ComputeInFlightShardMessageHash(msg)
	return msg
}
