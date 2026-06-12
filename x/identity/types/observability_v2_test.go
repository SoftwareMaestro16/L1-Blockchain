package types

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIdentityObservabilitySpecV2CoversSection15(t *testing.T) {
	spec, err := DefaultIdentityObservabilitySpecV2()
	require.NoError(t, err)
	require.NoError(t, ValidateIdentityObservabilitySpecV2(spec))

	require.ElementsMatch(t, requiredIdentityObservabilityEventsV2(), spec.Events)
	require.ElementsMatch(t, requiredIdentityObservabilityMetricsV2(), spec.Metrics)
	require.ElementsMatch(t, requiredIdentityObservabilityAlertsV2(), spec.Alerts)

	for _, eventType := range []IdentityObservabilityEventTypeV2{
		IdentityEventDomainCommitted,
		IdentityEventDomainRegistered,
		IdentityEventDomainRenewed,
		IdentityEventDomainTransferred,
		IdentityEventDomainExpired,
		IdentityEventDomainReleased,
		IdentityEventNFTBindingUpdated,
		IdentityEventResolverUpdated,
		IdentityEventReverseSet,
		IdentityEventReverseVerified,
		IdentityEventReverseInvalidated,
		IdentityEventSubdomainCreated,
		IdentityEventDelegationCreated,
		IdentityEventDelegationRevoked,
		IdentityEventZonePolicyUpdated,
		IdentityEventAuctionStarted,
		IdentityEventBidCommitted,
		IdentityEventBidRevealed,
		IdentityEventAuctionFinalized,
		IdentityEventCacheInvalidated,
	} {
		require.Contains(t, spec.Events, eventType)
	}

	for _, metric := range []IdentityObservabilityMetricNameV2{
		IdentityMetricActiveDomains,
		IdentityMetricExpiredDomains,
		IdentityMetricRenewalWindowDomains,
		IdentityMetricGracePeriodDomains,
		IdentityMetricResolverRecordCount,
		IdentityMetricAverageResolverPayloadSize,
		IdentityMetricReverseRecordsVerified,
		IdentityMetricReverseRecordsInvalidated,
		IdentityMetricSubdomainsByDepth,
		IdentityMetricDelegationRecordsActive,
		IdentityMetricAuctionsActive,
		IdentityMetricCommitmentsActive,
		IdentityMetricBatchResolverUpdateSize,
		IdentityMetricBlockSTMConflictRate,
		IdentityMetricStoreV2DirectReadLatency,
		IdentityMetricStoreV2RecursiveReadLatency,
		IdentityMetricStoreV2ResolverWriteLatency,
		IdentityMetricProofQueryLatency,
		IdentityMetricProofVerificationFailureCount,
		IdentityMetricExpiryProcessingBacklog,
	} {
		require.Contains(t, spec.Metrics, metric)
	}

	for _, alertType := range []IdentityObservabilityAlertTypeV2{
		IdentityAlertNFTBindingMismatch,
		IdentityAlertResolverPayloadNearMaximum,
		IdentityAlertExpiryBacklogAboveThreshold,
		IdentityAlertProofQueryFailureSpike,
		IdentityAlertRegistrationSpamSpike,
		IdentityAlertResolverUpdateSpamSpike,
		IdentityAlertAuctionFinalizationBacklog,
		IdentityAlertBlockSTMConflictRateHigh,
		IdentityAlertStoreV2ReadLatencyHigh,
		IdentityAlertReverseMismatchSpike,
	} {
		require.Contains(t, spec.Alerts, alertType)
	}
}

func TestIdentityObservabilityEventV2CanonicalABCIEvent(t *testing.T) {
	event, err := NewIdentityObservabilityEventV2(IdentityObservabilityEventV2{
		Type:	IdentityEventResolverUpdated,
		Height:	20,
		Name:	"Alice.AET",
		Actor:	hex.EncodeToString(addr(1)),
		Attributes: map[string]string{
			"record_version":	"2",
			"scope":		string(DelegationScopeResolverUpdate),
		},
	})
	require.NoError(t, err)
	require.Equal(t, "alice.aet", event.Name)
	require.NoError(t, ValidateIdentityObservabilityEventV2(event))
	require.NotEmpty(t, event.EventHash)

	abci, err := BuildIdentityObservabilityABCIEventV2(event)
	require.NoError(t, err)
	require.Equal(t, string(IdentityEventResolverUpdated), abci.Type)
	require.Equal(t, event.NameHash, abci.NameHash)
	require.Contains(t, abci.Attributes, "record_version=2")
	require.Contains(t, abci.Attributes, "scope=resolver_update")
	require.Contains(t, abci.Attributes, "event_hash="+event.EventHash)

	tampered := event
	tampered.Attributes["record_version"] = "3"
	require.ErrorContains(t, ValidateIdentityObservabilityEventV2(tampered), "hash mismatch")
}

func TestIdentityObservabilityMetricsSnapshotV2(t *testing.T) {
	state := observabilityState(t)
	delegation, err := NewDelegationRecordV2("alice.aet", addr(7), DelegationScopeResolverUpdate, []string{ResolverKeyPrimary}, 80, 0, ResolverKeyPrimary, 15)
	require.NoError(t, err)

	snapshot, err := BuildIdentityObservabilityMetricsSnapshotV2(IdentityObservabilityMetricsInputV2{
		State:					state,
		Height:					50,
		Delegations:				[]DelegationRecordV2{delegation},
		BatchResolverUpdateSize:		3,
		BlockSTMIdentityMessages:		10,
		BlockSTMConflicts:			2,
		StoreV2DirectReadLatencyMicros:		11,
		StoreV2RecursiveReadLatencyMicros:	22,
		StoreV2ResolverWriteLatencyMicros:	33,
		ProofQueryLatencyMicros:		44,
		ProofVerificationFailureCount:		5,
		ReverseRecordsInvalidated:		1,
	})
	require.NoError(t, err)
	require.NoError(t, ValidateIdentityObservabilityMetricsSnapshotV2(snapshot))
	require.Equal(t, uint64(50), snapshot.Height)
	require.NotEmpty(t, snapshot.SnapshotHash)

	metrics := metricSamplesByNameV2(snapshot)
	require.Equal(t, uint64(3), metrics[IdentityMetricActiveDomains][0].Value)
	require.Equal(t, uint64(1), metrics[IdentityMetricExpiredDomains][0].Value)
	require.Equal(t, uint64(1), metrics[IdentityMetricRenewalWindowDomains][0].Value)
	require.Equal(t, uint64(1), metrics[IdentityMetricGracePeriodDomains][0].Value)
	require.Equal(t, uint64(1), metrics[IdentityMetricResolverRecordCount][0].Value)
	require.Greater(t, metrics[IdentityMetricAverageResolverPayloadSize][0].Value, uint64(0))
	require.Equal(t, uint64(1), metrics[IdentityMetricReverseRecordsVerified][0].Value)
	require.Equal(t, uint64(1), metrics[IdentityMetricReverseRecordsInvalidated][0].Value)
	require.Equal(t, uint64(1), metrics[IdentityMetricDelegationRecordsActive][0].Value)
	require.Equal(t, uint64(1), metrics[IdentityMetricAuctionsActive][0].Value)
	require.Equal(t, uint64(1), metrics[IdentityMetricCommitmentsActive][0].Value)
	require.Equal(t, uint64(3), metrics[IdentityMetricBatchResolverUpdateSize][0].Value)
	require.Equal(t, uint64(2_000), metrics[IdentityMetricBlockSTMConflictRate][0].Value)
	require.Equal(t, uint64(11), metrics[IdentityMetricStoreV2DirectReadLatency][0].Value)
	require.Equal(t, uint64(22), metrics[IdentityMetricStoreV2RecursiveReadLatency][0].Value)
	require.Equal(t, uint64(33), metrics[IdentityMetricStoreV2ResolverWriteLatency][0].Value)
	require.Equal(t, uint64(44), metrics[IdentityMetricProofQueryLatency][0].Value)
	require.Equal(t, uint64(5), metrics[IdentityMetricProofVerificationFailureCount][0].Value)
	require.Equal(t, uint64(1), metrics[IdentityMetricExpiryProcessingBacklog][0].Value)

	depthSamples := metrics[IdentityMetricSubdomainsByDepth]
	require.Len(t, depthSamples, 1)
	require.Equal(t, map[string]string{"depth": "1"}, depthSamples[0].Labels)
	require.Equal(t, uint64(1), depthSamples[0].Value)
	require.Equal(t, IdentityMetricUnitBasisPoints, metrics[IdentityMetricBlockSTMConflictRate][0].Unit)
	require.Equal(t, IdentityMetricUnitMicroseconds, metrics[IdentityMetricProofQueryLatency][0].Unit)
}

func TestIdentityObservabilityMetricsRejectMissingAndBadUnitsV2(t *testing.T) {
	sample, err := NewIdentityMetricSampleV2(IdentityMetricProofQueryLatency, 10, 5, IdentityMetricUnitCount, nil)
	require.ErrorContains(t, err, "unit must be us")
	require.NotEmpty(t, sample.SampleHash)

	valid, err := NewIdentityMetricSampleV2(IdentityMetricProofQueryLatency, 10, 5, IdentityMetricUnitMicroseconds, nil)
	require.NoError(t, err)
	snapshot := IdentityObservabilityMetricsSnapshotV2{
		Height:		10,
		Metrics:	[]IdentityMetricSampleV2{valid},
		SnapshotHash:	identityHash("wrong"),
	}
	require.ErrorContains(t, ValidateIdentityObservabilityMetricsSnapshotV2(snapshot), "missing required metrics")
}

func TestIdentityObservabilityAlertsV2EvaluateRequiredAlerts(t *testing.T) {
	validState := observabilityState(t)
	snapshot, err := BuildIdentityObservabilityMetricsSnapshotV2(IdentityObservabilityMetricsInputV2{
		State:					validState,
		Height:					70,
		BlockSTMIdentityMessages:		10,
		BlockSTMConflicts:			5,
		StoreV2DirectReadLatencyMicros:		200,
		StoreV2RecursiveReadLatencyMicros:	300,
		ProofQueryLatencyMicros:		400,
		ProofVerificationFailureCount:		6,
	})
	require.NoError(t, err)

	broken := validState.Clone()
	require.NotEmpty(t, broken.DomainNFTs)
	broken.DomainNFTs[0].Owner = addr(99)
	require.NotEmpty(t, broken.Resolvers)
	broken.Resolvers[0].Metadata = []byte(strings.Repeat("a", MaxUnifiedPayloadBytesV2))

	alerts, err := EvaluateIdentityObservabilityAlertsV2(IdentityObservabilityAlertInputV2{
		State:				broken,
		Snapshot:			snapshot,
		Height:				70,
		RegistrationAttemptsInWindow:	2,
		ResolverUpdatesInWindow:	2,
		ReverseMismatchesInWindow:	2,
		Thresholds: IdentityObservabilityAlertThresholdsV2{
			ResolverPayloadNearMaxBps:		8_000,
			ExpiryBacklogThreshold:			1,
			ProofFailureSpikeThreshold:		1,
			RegistrationSpamSpikeThreshold:		1,
			ResolverUpdateSpamSpikeThreshold:	1,
			AuctionFinalizationBacklogThreshold:	1,
			BlockSTMConflictRateBpsThreshold:	1,
			StoreV2ReadLatencyMicrosThreshold:	1,
			ReverseMismatchSpikeThreshold:		1,
		},
	})
	require.NoError(t, err)
	byType := alertTypesV2(alerts)
	for _, alertType := range requiredIdentityObservabilityAlertsV2() {
		require.Contains(t, byType, alertType)
	}
	require.Equal(t, IdentityAlertSeverityCritical, byType[IdentityAlertNFTBindingMismatch].Severity)
	require.Equal(t, IdentityAlertSeverityWarning, byType[IdentityAlertResolverPayloadNearMaximum].Severity)
	require.Equal(t, uint64(5_000), byType[IdentityAlertBlockSTMConflictRateHigh].ObservedValue)
	require.Equal(t, uint64(300), byType[IdentityAlertStoreV2ReadLatencyHigh].ObservedValue)
	for _, alert := range alerts {
		require.NoError(t, ValidateIdentityObservabilityAlertV2(alert))
		require.NotEmpty(t, alert.AlertHash)
	}
}

func TestIdentityObservabilityAlertsV2StayQuietBelowThresholds(t *testing.T) {
	state := observabilityState(t)
	snapshot, err := BuildIdentityObservabilityMetricsSnapshotV2(IdentityObservabilityMetricsInputV2{
		State:				state,
		Height:				50,
		BlockSTMIdentityMessages:	100,
		BlockSTMConflicts:		1,
	})
	require.NoError(t, err)

	alerts, err := EvaluateIdentityObservabilityAlertsV2(IdentityObservabilityAlertInputV2{
		State:		state,
		Snapshot:	snapshot,
		Height:		50,
		Thresholds: IdentityObservabilityAlertThresholdsV2{
			ResolverPayloadNearMaxBps:		10_000,
			ExpiryBacklogThreshold:			100,
			ProofFailureSpikeThreshold:		100,
			RegistrationSpamSpikeThreshold:		100,
			ResolverUpdateSpamSpikeThreshold:	100,
			AuctionFinalizationBacklogThreshold:	100,
			BlockSTMConflictRateBpsThreshold:	1_000,
			StoreV2ReadLatencyMicrosThreshold:	100,
			ReverseMismatchSpikeThreshold:		100,
		},
	})
	require.NoError(t, err)
	require.Empty(t, alerts)
}

func observabilityState(t *testing.T) IdentityState {
	t.Helper()
	state := routingIntegrationState(t)
	state, _, err := SetIdentityReverse(state, addr(2), addr(2), "alice.aet", 13)
	require.NoError(t, err)
	bob, _ := registerSpecDomain(t, "bob", addr(4), "salt-bob", 10)
	bobDomain, found := findDomain(bob, "bob.aet")
	require.True(t, found)
	bobDomain.ExpiryHeight = 55
	state.Domains = upsertDomain(state.Domains, bobDomain)
	state.DomainNFTs = append(state.DomainNFTs, bob.DomainNFTs...)
	expired, _ := registerSpecDomain(t, "old", addr(5), "salt-old", 10)
	oldDomain, found := findDomain(expired, "old.aet")
	require.True(t, found)
	oldDomain.ExpiryHeight = 45
	state.Domains = upsertDomain(state.Domains, oldDomain)
	state.DomainNFTs = append(state.DomainNFTs, expired.DomainNFTs...)
	state, _, err = IssueSubdomainV2(state, SubdomainCreationPolicyV2{
		ParentName:		"alice.aet",
		Label:			"api",
		Actor:			addr(1),
		ChildOwner:		addr(1),
		Height:			20,
		ChildExpiryHeight:	90,
		DelegationType:		SubdomainDelegationOwnerControlledV2,
	})
	require.NoError(t, err)
	commit, err := CommitDomainRegistration(state, "commit.aet", addr(6), identityHash("commit"), 45)
	require.NoError(t, err)
	state = commit
	state.Auctions = append(state.Auctions, Auction{Name: "auction.aet", CommitStartHeight: 40, RevealStartHeight: 45, RevealEndHeight: 60, Phase: AuctionPhaseCommit})
	state.Params = normalizeIdentityParams(state.Params)
	state.Params.RenewalWindowBlocks = 10
	state = state.Export()
	require.NoError(t, state.Validate())
	return state
}

func metricSamplesByNameV2(snapshot IdentityObservabilityMetricsSnapshotV2) map[IdentityObservabilityMetricNameV2][]IdentityMetricSampleV2 {
	out := map[IdentityObservabilityMetricNameV2][]IdentityMetricSampleV2{}
	for _, sample := range snapshot.Metrics {
		out[sample.Name] = append(out[sample.Name], sample)
	}
	return out
}

func alertTypesV2(alerts []IdentityObservabilityAlertV2) map[IdentityObservabilityAlertTypeV2]IdentityObservabilityAlertV2 {
	out := map[IdentityObservabilityAlertTypeV2]IdentityObservabilityAlertV2{}
	for _, alert := range alerts {
		out[alert.Type] = alert
	}
	return out
}
