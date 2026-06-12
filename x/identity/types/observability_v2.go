package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type IdentityObservabilityEventTypeV2 string
type IdentityObservabilityMetricNameV2 string
type IdentityObservabilityAlertTypeV2 string
type IdentityObservabilityAlertSeverityV2 string

const (
	IdentityEventDomainCommitted	IdentityObservabilityEventTypeV2	= "identity_domain_committed"
	IdentityEventDomainRegistered	IdentityObservabilityEventTypeV2	= "identity_domain_registered"
	IdentityEventDomainRenewed	IdentityObservabilityEventTypeV2	= "identity_domain_renewed"
	IdentityEventDomainTransferred	IdentityObservabilityEventTypeV2	= "identity_domain_transferred"
	IdentityEventDomainExpired	IdentityObservabilityEventTypeV2	= "identity_domain_expired"
	IdentityEventDomainReleased	IdentityObservabilityEventTypeV2	= "identity_domain_released"
	IdentityEventNFTBindingUpdated	IdentityObservabilityEventTypeV2	= "identity_nft_binding_updated"
	IdentityEventResolverUpdated	IdentityObservabilityEventTypeV2	= "identity_resolver_updated"
	IdentityEventReverseSet		IdentityObservabilityEventTypeV2	= "identity_reverse_set"
	IdentityEventReverseVerified	IdentityObservabilityEventTypeV2	= "identity_reverse_verified"
	IdentityEventReverseInvalidated	IdentityObservabilityEventTypeV2	= "identity_reverse_invalidated"
	IdentityEventSubdomainCreated	IdentityObservabilityEventTypeV2	= "identity_subdomain_created"
	IdentityEventDelegationCreated	IdentityObservabilityEventTypeV2	= "identity_delegation_created"
	IdentityEventDelegationRevoked	IdentityObservabilityEventTypeV2	= "identity_delegation_revoked"
	IdentityEventZonePolicyUpdated	IdentityObservabilityEventTypeV2	= "identity_zone_policy_updated"
	IdentityEventAuctionStarted	IdentityObservabilityEventTypeV2	= "identity_auction_started"
	IdentityEventBidCommitted	IdentityObservabilityEventTypeV2	= "identity_bid_committed"
	IdentityEventBidRevealed	IdentityObservabilityEventTypeV2	= "identity_bid_revealed"
	IdentityEventAuctionFinalized	IdentityObservabilityEventTypeV2	= "identity_auction_finalized"
	IdentityEventCacheInvalidated	IdentityObservabilityEventTypeV2	= "identity_cache_invalidated"
)

const (
	IdentityMetricActiveDomains			IdentityObservabilityMetricNameV2	= "identity_active_domains"
	IdentityMetricExpiredDomains			IdentityObservabilityMetricNameV2	= "identity_expired_domains"
	IdentityMetricRenewalWindowDomains		IdentityObservabilityMetricNameV2	= "identity_domains_in_renewal_window"
	IdentityMetricGracePeriodDomains		IdentityObservabilityMetricNameV2	= "identity_domains_in_grace_period"
	IdentityMetricResolverRecordCount		IdentityObservabilityMetricNameV2	= "identity_resolver_record_count"
	IdentityMetricAverageResolverPayloadSize	IdentityObservabilityMetricNameV2	= "identity_average_resolver_payload_size_bytes"
	IdentityMetricReverseRecordsVerified		IdentityObservabilityMetricNameV2	= "identity_reverse_records_verified"
	IdentityMetricReverseRecordsInvalidated		IdentityObservabilityMetricNameV2	= "identity_reverse_records_invalidated"
	IdentityMetricSubdomainsByDepth			IdentityObservabilityMetricNameV2	= "identity_subdomains_by_depth"
	IdentityMetricDelegationRecordsActive		IdentityObservabilityMetricNameV2	= "identity_delegation_records_active"
	IdentityMetricAuctionsActive			IdentityObservabilityMetricNameV2	= "identity_auctions_active"
	IdentityMetricCommitmentsActive			IdentityObservabilityMetricNameV2	= "identity_commitments_active"
	IdentityMetricBatchResolverUpdateSize		IdentityObservabilityMetricNameV2	= "identity_batch_resolver_update_size"
	IdentityMetricBlockSTMConflictRate		IdentityObservabilityMetricNameV2	= "identity_blockstm_conflict_rate_bps"
	IdentityMetricStoreV2DirectReadLatency		IdentityObservabilityMetricNameV2	= "identity_store_v2_direct_resolution_read_latency_us"
	IdentityMetricStoreV2RecursiveReadLatency	IdentityObservabilityMetricNameV2	= "identity_store_v2_recursive_resolution_read_latency_us"
	IdentityMetricStoreV2ResolverWriteLatency	IdentityObservabilityMetricNameV2	= "identity_store_v2_resolver_update_write_latency_us"
	IdentityMetricProofQueryLatency			IdentityObservabilityMetricNameV2	= "identity_proof_query_latency_us"
	IdentityMetricProofVerificationFailureCount	IdentityObservabilityMetricNameV2	= "identity_proof_verification_failure_count"
	IdentityMetricExpiryProcessingBacklog		IdentityObservabilityMetricNameV2	= "identity_expiry_processing_backlog"
)

const (
	IdentityMetricUnitCount		= "count"
	IdentityMetricUnitBytes		= "bytes"
	IdentityMetricUnitMicroseconds	= "us"
	IdentityMetricUnitBasisPoints	= "bps"
)

const (
	IdentityAlertNFTBindingMismatch			IdentityObservabilityAlertTypeV2	= "identity_alert_nft_binding_mismatch"
	IdentityAlertResolverPayloadNearMaximum		IdentityObservabilityAlertTypeV2	= "identity_alert_resolver_payload_near_maximum"
	IdentityAlertExpiryBacklogAboveThreshold	IdentityObservabilityAlertTypeV2	= "identity_alert_expiry_processing_backlog_above_threshold"
	IdentityAlertProofQueryFailureSpike		IdentityObservabilityAlertTypeV2	= "identity_alert_proof_query_failure_spike"
	IdentityAlertRegistrationSpamSpike		IdentityObservabilityAlertTypeV2	= "identity_alert_registration_spam_spike"
	IdentityAlertResolverUpdateSpamSpike		IdentityObservabilityAlertTypeV2	= "identity_alert_resolver_update_spam_spike"
	IdentityAlertAuctionFinalizationBacklog		IdentityObservabilityAlertTypeV2	= "identity_alert_auction_finalization_backlog"
	IdentityAlertBlockSTMConflictRateHigh		IdentityObservabilityAlertTypeV2	= "identity_alert_blockstm_conflict_rate_above_threshold"
	IdentityAlertStoreV2ReadLatencyHigh		IdentityObservabilityAlertTypeV2	= "identity_alert_store_v2_identity_read_latency_above_threshold"
	IdentityAlertReverseMismatchSpike		IdentityObservabilityAlertTypeV2	= "identity_alert_reverse_verification_mismatch_spike"
)

const (
	IdentityAlertSeverityWarning	IdentityObservabilityAlertSeverityV2	= "warning"
	IdentityAlertSeverityCritical	IdentityObservabilityAlertSeverityV2	= "critical"
)

type IdentityObservabilitySpecV2 struct {
	Events		[]IdentityObservabilityEventTypeV2
	Metrics		[]IdentityObservabilityMetricNameV2
	Alerts		[]IdentityObservabilityAlertTypeV2
	SpecHash	string
}

type IdentityObservabilityEventV2 struct {
	Type		IdentityObservabilityEventTypeV2
	Height		uint64
	Name		string
	NameHash	string
	Actor		string
	Attributes	map[string]string
	EventHash	string
}

type IdentityMetricSampleV2 struct {
	Name		IdentityObservabilityMetricNameV2
	Height		uint64
	Value		uint64
	Unit		string
	Labels		map[string]string
	SampleHash	string
}

type IdentityObservabilityMetricsInputV2 struct {
	State					IdentityState
	Height					uint64
	Delegations				[]DelegationRecordV2
	BatchResolverUpdateSize			uint64
	BlockSTMIdentityMessages		uint64
	BlockSTMConflicts			uint64
	StoreV2DirectReadLatencyMicros		uint64
	StoreV2RecursiveReadLatencyMicros	uint64
	StoreV2ResolverWriteLatencyMicros	uint64
	ProofQueryLatencyMicros			uint64
	ProofVerificationFailureCount		uint64
	ReverseRecordsInvalidated		uint64
}

type IdentityObservabilityMetricsSnapshotV2 struct {
	Height		uint64
	Metrics		[]IdentityMetricSampleV2
	SnapshotHash	string
}

type IdentityObservabilityAlertThresholdsV2 struct {
	ResolverPayloadNearMaxBps		uint64
	ExpiryBacklogThreshold			uint64
	ProofFailureSpikeThreshold		uint64
	RegistrationSpamSpikeThreshold		uint64
	ResolverUpdateSpamSpikeThreshold	uint64
	AuctionFinalizationBacklogThreshold	uint64
	BlockSTMConflictRateBpsThreshold	uint64
	StoreV2ReadLatencyMicrosThreshold	uint64
	ReverseMismatchSpikeThreshold		uint64
}

type IdentityObservabilityAlertInputV2 struct {
	State				IdentityState
	Snapshot			IdentityObservabilityMetricsSnapshotV2
	Height				uint64
	Thresholds			IdentityObservabilityAlertThresholdsV2
	RegistrationAttemptsInWindow	uint64
	ResolverUpdatesInWindow		uint64
	ReverseMismatchesInWindow	uint64
}

type IdentityObservabilityAlertV2 struct {
	Type		IdentityObservabilityAlertTypeV2
	Severity	IdentityObservabilityAlertSeverityV2
	Height		uint64
	ObservedValue	uint64
	Threshold	uint64
	Name		string
	NameHash	string
	Metric		IdentityObservabilityMetricNameV2
	Details		string
	AlertHash	string
}

func DefaultIdentityObservabilitySpecV2() (IdentityObservabilitySpecV2, error) {
	spec := IdentityObservabilitySpecV2{
		Events:		requiredIdentityObservabilityEventsV2(),
		Metrics:	requiredIdentityObservabilityMetricsV2(),
		Alerts:		requiredIdentityObservabilityAlertsV2(),
	}
	spec.SpecHash = ComputeIdentityObservabilitySpecHashV2(spec)
	return spec, ValidateIdentityObservabilitySpecV2(spec)
}

func ValidateIdentityObservabilitySpecV2(spec IdentityObservabilitySpecV2) error {
	if err := validateRequiredTypedSetV2("identity observability event", spec.Events, requiredIdentityObservabilityEventsV2(), IsIdentityObservabilityEventTypeV2); err != nil {
		return err
	}
	if err := validateRequiredTypedSetV2("identity observability metric", spec.Metrics, requiredIdentityObservabilityMetricsV2(), IsIdentityObservabilityMetricNameV2); err != nil {
		return err
	}
	if err := validateRequiredTypedSetV2("identity observability alert", spec.Alerts, requiredIdentityObservabilityAlertsV2(), IsIdentityObservabilityAlertTypeV2); err != nil {
		return err
	}
	if spec.SpecHash == "" || spec.SpecHash != ComputeIdentityObservabilitySpecHashV2(spec) {
		return errors.New("identity observability spec hash mismatch")
	}
	return nil
}

func NewIdentityObservabilityEventV2(event IdentityObservabilityEventV2) (IdentityObservabilityEventV2, error) {
	if event.EventHash != "" {
		return IdentityObservabilityEventV2{}, errors.New("identity observability event hash must be empty before construction")
	}
	if event.Name != "" {
		normalized, err := NormalizeAETDomain(event.Name)
		if err != nil {
			return IdentityObservabilityEventV2{}, err
		}
		event.Name = normalized
		nameHash, err := DomainRecordV2NameHash(normalized)
		if err != nil {
			return IdentityObservabilityEventV2{}, err
		}
		if event.NameHash != "" && event.NameHash != nameHash {
			return IdentityObservabilityEventV2{}, errors.New("identity observability event name_hash mismatch")
		}
		event.NameHash = nameHash
	}
	event.Attributes = cloneStringMapV2(event.Attributes)
	event.EventHash = ComputeIdentityObservabilityEventHashV2(event)
	return event, ValidateIdentityObservabilityEventV2(event)
}

func ValidateIdentityObservabilityEventV2(event IdentityObservabilityEventV2) error {
	if !IsIdentityObservabilityEventTypeV2(event.Type) {
		return fmt.Errorf("unsupported identity observability event type %q", event.Type)
	}
	if event.Height == 0 {
		return errors.New("identity observability event height is required")
	}
	if event.NameHash != "" {
		if err := validateHexHash("identity observability event name_hash", event.NameHash); err != nil {
			return err
		}
	}
	if event.Name != "" {
		normalized, err := NormalizeAETDomain(event.Name)
		if err != nil {
			return err
		}
		nameHash, err := DomainRecordV2NameHash(normalized)
		if err != nil {
			return err
		}
		if event.NameHash != nameHash {
			return errors.New("identity observability event name_hash mismatch")
		}
	}
	for key, value := range event.Attributes {
		if strings.TrimSpace(key) == "" {
			return errors.New("identity observability event attribute key is required")
		}
		if !isASCIIStringV2(key) || !isASCIIStringV2(value) {
			return errors.New("identity observability event attributes must be ASCII")
		}
	}
	if event.EventHash == "" || event.EventHash != ComputeIdentityObservabilityEventHashV2(event) {
		return errors.New("identity observability event hash mismatch")
	}
	return nil
}

func BuildIdentityObservabilityABCIEventV2(event IdentityObservabilityEventV2) (IdentityABCIEventV2, error) {
	if err := ValidateIdentityObservabilityEventV2(event); err != nil {
		return IdentityABCIEventV2{}, err
	}
	return IdentityABCIEventV2{
		Type:		string(event.Type),
		Height:		event.Height,
		Name:		event.Name,
		NameHash:	event.NameHash,
		Attributes:	observabilityAttributesV2(event),
	}, nil
}

func BuildIdentityObservabilityMetricsSnapshotV2(input IdentityObservabilityMetricsInputV2) (IdentityObservabilityMetricsSnapshotV2, error) {
	if input.Height == 0 {
		return IdentityObservabilityMetricsSnapshotV2{}, errors.New("identity observability metrics height is required")
	}
	state := input.State.Export()
	if err := state.Validate(); err != nil {
		return IdentityObservabilityMetricsSnapshotV2{}, err
	}
	params := normalizeIdentityParams(state.Params)
	counts := map[IdentityObservabilityMetricNameV2]uint64{
		IdentityMetricActiveDomains:			0,
		IdentityMetricExpiredDomains:			0,
		IdentityMetricRenewalWindowDomains:		0,
		IdentityMetricGracePeriodDomains:		0,
		IdentityMetricResolverRecordCount:		uint64(len(state.Resolvers)),
		IdentityMetricReverseRecordsVerified:		0,
		IdentityMetricDelegationRecordsActive:		0,
		IdentityMetricAuctionsActive:			0,
		IdentityMetricCommitmentsActive:		0,
		IdentityMetricExpiryProcessingBacklog:		0,
		IdentityMetricReverseRecordsInvalidated:	input.ReverseRecordsInvalidated,
		IdentityMetricBatchResolverUpdateSize:		input.BatchResolverUpdateSize,
		IdentityMetricProofVerificationFailureCount:	input.ProofVerificationFailureCount,
	}
	var resolverPayloadBytes uint64
	for _, domain := range state.Domains {
		status, err := DomainLifecycle(state, domain.Name, input.Height)
		if err != nil {
			return IdentityObservabilityMetricsSnapshotV2{}, err
		}
		switch status {
		case DomainLifecycleActive:
			counts[IdentityMetricActiveDomains]++
		case DomainLifecycleRenewalWindow:
			counts[IdentityMetricActiveDomains]++
			counts[IdentityMetricRenewalWindowDomains]++
		case DomainLifecycleExpired:
			counts[IdentityMetricExpiredDomains]++
			counts[IdentityMetricExpiryProcessingBacklog]++
			if input.Height < domain.ExpiryHeight+params.RenewalWindowBlocks {
				counts[IdentityMetricGracePeriodDomains]++
			}
		}
	}
	for _, commit := range state.Commits {
		if commit.ExpiresHeight > input.Height {
			counts[IdentityMetricCommitmentsActive]++
		}
	}
	for _, auction := range state.Auctions {
		if auction.Phase != AuctionPhaseFinalized {
			counts[IdentityMetricAuctionsActive]++
		}
	}
	for _, resolver := range state.Resolvers {
		resolverPayloadBytes += estimateResolverRecordPayloadBytesV2(resolver)
	}
	avgResolverPayload := uint64(0)
	if len(state.Resolvers) > 0 {
		avgResolverPayload = resolverPayloadBytes / uint64(len(state.Resolvers))
	}
	counts[IdentityMetricAverageResolverPayloadSize] = avgResolverPayload
	for _, reverse := range state.ReverseRecords {
		v2, err := reverseRecordV2FromLegacy(state, reverse, true)
		if err != nil {
			continue
		}
		if ValidateReverseResolutionRecordV2(state, v2, input.Height, nil) == nil {
			counts[IdentityMetricReverseRecordsVerified]++
		}
	}
	for _, delegation := range input.Delegations {
		if err := ValidateDelegationRecordV2(delegation); err == nil && input.Height < delegation.ExpiresAtHeight {
			counts[IdentityMetricDelegationRecordsActive]++
		}
	}
	conflictRate := uint64(0)
	if input.BlockSTMIdentityMessages > 0 {
		conflictRate = input.BlockSTMConflicts * 10_000 / input.BlockSTMIdentityMessages
	}
	counts[IdentityMetricBlockSTMConflictRate] = conflictRate
	counts[IdentityMetricStoreV2DirectReadLatency] = input.StoreV2DirectReadLatencyMicros
	counts[IdentityMetricStoreV2RecursiveReadLatency] = input.StoreV2RecursiveReadLatencyMicros
	counts[IdentityMetricStoreV2ResolverWriteLatency] = input.StoreV2ResolverWriteLatencyMicros
	counts[IdentityMetricProofQueryLatency] = input.ProofQueryLatencyMicros

	snapshot := IdentityObservabilityMetricsSnapshotV2{Height: input.Height}
	for _, metric := range requiredIdentityObservabilityMetricsV2() {
		unit := identityObservabilityMetricUnitV2(metric)
		if metric == IdentityMetricSubdomainsByDepth {
			depths := subdomainDepthCountsV2(state.Subdomains)
			if len(depths) == 0 {
				sample, err := NewIdentityMetricSampleV2(metric, input.Height, 0, unit, map[string]string{"depth": "0"})
				if err != nil {
					return IdentityObservabilityMetricsSnapshotV2{}, err
				}
				snapshot.Metrics = append(snapshot.Metrics, sample)
			}
			for _, depth := range sortedUintMapKeysV2(depths) {
				sample, err := NewIdentityMetricSampleV2(metric, input.Height, depths[depth], unit, map[string]string{"depth": fmt.Sprint(depth)})
				if err != nil {
					return IdentityObservabilityMetricsSnapshotV2{}, err
				}
				snapshot.Metrics = append(snapshot.Metrics, sample)
			}
			continue
		}
		sample, err := NewIdentityMetricSampleV2(metric, input.Height, counts[metric], unit, nil)
		if err != nil {
			return IdentityObservabilityMetricsSnapshotV2{}, err
		}
		snapshot.Metrics = append(snapshot.Metrics, sample)
	}
	snapshot.SnapshotHash = ComputeIdentityObservabilityMetricsSnapshotHashV2(snapshot)
	return snapshot, ValidateIdentityObservabilityMetricsSnapshotV2(snapshot)
}

func NewIdentityMetricSampleV2(name IdentityObservabilityMetricNameV2, height uint64, value uint64, unit string, labels map[string]string) (IdentityMetricSampleV2, error) {
	sample := IdentityMetricSampleV2{
		Name:	name,
		Height:	height,
		Value:	value,
		Unit:	unit,
		Labels:	cloneStringMapV2(labels),
	}
	sample.SampleHash = ComputeIdentityMetricSampleHashV2(sample)
	return sample, ValidateIdentityMetricSampleV2(sample)
}

func ValidateIdentityMetricSampleV2(sample IdentityMetricSampleV2) error {
	if !IsIdentityObservabilityMetricNameV2(sample.Name) {
		return fmt.Errorf("unsupported identity observability metric %q", sample.Name)
	}
	if sample.Height == 0 {
		return errors.New("identity observability metric height is required")
	}
	if expected := identityObservabilityMetricUnitV2(sample.Name); sample.Unit != expected {
		return fmt.Errorf("identity observability metric %s unit must be %s", sample.Name, expected)
	}
	for key, value := range sample.Labels {
		if strings.TrimSpace(key) == "" || !isASCIIStringV2(key) || !isASCIIStringV2(value) {
			return errors.New("identity observability metric labels must be non-empty ASCII")
		}
	}
	if sample.SampleHash == "" || sample.SampleHash != ComputeIdentityMetricSampleHashV2(sample) {
		return errors.New("identity observability metric sample hash mismatch")
	}
	return nil
}

func ValidateIdentityObservabilityMetricsSnapshotV2(snapshot IdentityObservabilityMetricsSnapshotV2) error {
	if snapshot.Height == 0 {
		return errors.New("identity observability metrics snapshot height is required")
	}
	if len(snapshot.Metrics) < len(requiredIdentityObservabilityMetricsV2()) {
		return errors.New("identity observability metrics snapshot missing required metrics")
	}
	seen := map[IdentityObservabilityMetricNameV2]bool{}
	for _, sample := range snapshot.Metrics {
		if err := ValidateIdentityMetricSampleV2(sample); err != nil {
			return err
		}
		seen[sample.Name] = true
	}
	for _, metric := range requiredIdentityObservabilityMetricsV2() {
		if !seen[metric] {
			return fmt.Errorf("identity observability metrics snapshot missing %s", metric)
		}
	}
	if snapshot.SnapshotHash == "" || snapshot.SnapshotHash != ComputeIdentityObservabilityMetricsSnapshotHashV2(snapshot) {
		return errors.New("identity observability metrics snapshot hash mismatch")
	}
	return nil
}

func DefaultIdentityObservabilityAlertThresholdsV2() IdentityObservabilityAlertThresholdsV2 {
	return IdentityObservabilityAlertThresholdsV2{
		ResolverPayloadNearMaxBps:		8_000,
		ExpiryBacklogThreshold:			100,
		ProofFailureSpikeThreshold:		10,
		RegistrationSpamSpikeThreshold:		100,
		ResolverUpdateSpamSpikeThreshold:	100,
		AuctionFinalizationBacklogThreshold:	10,
		BlockSTMConflictRateBpsThreshold:	2_000,
		StoreV2ReadLatencyMicrosThreshold:	50_000,
		ReverseMismatchSpikeThreshold:		10,
	}
}

func EvaluateIdentityObservabilityAlertsV2(input IdentityObservabilityAlertInputV2) ([]IdentityObservabilityAlertV2, error) {
	if input.Height == 0 {
		return nil, errors.New("identity observability alert height is required")
	}
	state := input.State.Export()
	if input.Snapshot.Height != 0 {
		if err := ValidateIdentityObservabilityMetricsSnapshotV2(input.Snapshot); err != nil {
			return nil, err
		}
	}
	thresholds := normalizeIdentityObservabilityAlertThresholdsV2(input.Thresholds)
	alerts := make([]IdentityObservabilityAlertV2, 0)
	for _, domain := range state.Domains {
		nft, found := findDomainNFTByID(state, domain.NFTID)
		if !found || !addressesEqual(nft.Owner, domain.Owner) {
			alert, err := NewIdentityObservabilityAlertV2(IdentityObservabilityAlertV2{
				Type:		IdentityAlertNFTBindingMismatch,
				Severity:	IdentityAlertSeverityCritical,
				Height:		input.Height,
				ObservedValue:	1,
				Threshold:	0,
				Name:		domain.Name,
				Details:	"registry owner and nft owner diverged",
			})
			if err != nil {
				return nil, err
			}
			alerts = append(alerts, alert)
		}
	}
	for _, resolver := range state.Resolvers {
		payloadBytes := estimateResolverRecordPayloadBytesV2(resolver)
		payloadBps := payloadBytes * 10_000 / MaxUnifiedPayloadBytesV2
		if payloadBps >= thresholds.ResolverPayloadNearMaxBps {
			alert, err := NewIdentityObservabilityAlertV2(IdentityObservabilityAlertV2{
				Type:		IdentityAlertResolverPayloadNearMaximum,
				Severity:	IdentityAlertSeverityWarning,
				Height:		input.Height,
				ObservedValue:	payloadBps,
				Threshold:	thresholds.ResolverPayloadNearMaxBps,
				Name:		resolver.Domain,
				Metric:		IdentityMetricAverageResolverPayloadSize,
				Details:	"resolver payload is near maximum",
			})
			if err != nil {
				return nil, err
			}
			alerts = append(alerts, alert)
		}
	}
	metricValues := identityObservabilityMetricValuesV2(input.Snapshot)
	var err error
	if alerts, err = appendThresholdAlertV2(alerts, input.Height, IdentityAlertExpiryBacklogAboveThreshold, IdentityAlertSeverityWarning, IdentityMetricExpiryProcessingBacklog, metricValues[IdentityMetricExpiryProcessingBacklog], thresholds.ExpiryBacklogThreshold, "expiry processing backlog above threshold"); err != nil {
		return nil, err
	}
	if alerts, err = appendThresholdAlertV2(alerts, input.Height, IdentityAlertProofQueryFailureSpike, IdentityAlertSeverityWarning, IdentityMetricProofVerificationFailureCount, metricValues[IdentityMetricProofVerificationFailureCount], thresholds.ProofFailureSpikeThreshold, "proof query failure spike"); err != nil {
		return nil, err
	}
	if alerts, err = appendThresholdAlertV2(alerts, input.Height, IdentityAlertBlockSTMConflictRateHigh, IdentityAlertSeverityWarning, IdentityMetricBlockSTMConflictRate, metricValues[IdentityMetricBlockSTMConflictRate], thresholds.BlockSTMConflictRateBpsThreshold, "BlockSTM identity conflict rate above threshold"); err != nil {
		return nil, err
	}
	maxReadLatency := metricValues[IdentityMetricStoreV2DirectReadLatency]
	if recursive := metricValues[IdentityMetricStoreV2RecursiveReadLatency]; recursive > maxReadLatency {
		maxReadLatency = recursive
	}
	if alerts, err = appendThresholdAlertV2(alerts, input.Height, IdentityAlertStoreV2ReadLatencyHigh, IdentityAlertSeverityWarning, IdentityMetricStoreV2DirectReadLatency, maxReadLatency, thresholds.StoreV2ReadLatencyMicrosThreshold, "Store v2 identity read latency above threshold"); err != nil {
		return nil, err
	}
	if alerts, err = appendThresholdAlertV2(alerts, input.Height, IdentityAlertRegistrationSpamSpike, IdentityAlertSeverityWarning, "", input.RegistrationAttemptsInWindow, thresholds.RegistrationSpamSpikeThreshold, "registration spam spike"); err != nil {
		return nil, err
	}
	if alerts, err = appendThresholdAlertV2(alerts, input.Height, IdentityAlertResolverUpdateSpamSpike, IdentityAlertSeverityWarning, "", input.ResolverUpdatesInWindow, thresholds.ResolverUpdateSpamSpikeThreshold, "resolver update spam spike"); err != nil {
		return nil, err
	}
	if alerts, err = appendThresholdAlertV2(alerts, input.Height, IdentityAlertReverseMismatchSpike, IdentityAlertSeverityWarning, "", input.ReverseMismatchesInWindow, thresholds.ReverseMismatchSpikeThreshold, "reverse verification mismatch spike"); err != nil {
		return nil, err
	}
	auctionBacklog := auctionFinalizationBacklogV2(state, input.Height)
	if alerts, err = appendThresholdAlertV2(alerts, input.Height, IdentityAlertAuctionFinalizationBacklog, IdentityAlertSeverityWarning, IdentityMetricAuctionsActive, auctionBacklog, thresholds.AuctionFinalizationBacklogThreshold, "auction finalization backlog"); err != nil {
		return nil, err
	}
	sort.Slice(alerts, func(i, j int) bool {
		if alerts[i].Type != alerts[j].Type {
			return alerts[i].Type < alerts[j].Type
		}
		if alerts[i].NameHash != alerts[j].NameHash {
			return alerts[i].NameHash < alerts[j].NameHash
		}
		return alerts[i].AlertHash < alerts[j].AlertHash
	})
	return alerts, nil
}

func NewIdentityObservabilityAlertV2(alert IdentityObservabilityAlertV2) (IdentityObservabilityAlertV2, error) {
	if alert.AlertHash != "" {
		return IdentityObservabilityAlertV2{}, errors.New("identity observability alert hash must be empty before construction")
	}
	if alert.Name != "" {
		normalized, err := NormalizeAETDomain(alert.Name)
		if err != nil {
			return IdentityObservabilityAlertV2{}, err
		}
		alert.Name = normalized
		nameHash, err := DomainRecordV2NameHash(normalized)
		if err != nil {
			return IdentityObservabilityAlertV2{}, err
		}
		if alert.NameHash != "" && alert.NameHash != nameHash {
			return IdentityObservabilityAlertV2{}, errors.New("identity observability alert name_hash mismatch")
		}
		alert.NameHash = nameHash
	}
	alert.AlertHash = ComputeIdentityObservabilityAlertHashV2(alert)
	return alert, ValidateIdentityObservabilityAlertV2(alert)
}

func ValidateIdentityObservabilityAlertV2(alert IdentityObservabilityAlertV2) error {
	if !IsIdentityObservabilityAlertTypeV2(alert.Type) {
		return fmt.Errorf("unsupported identity observability alert type %q", alert.Type)
	}
	switch alert.Severity {
	case IdentityAlertSeverityWarning, IdentityAlertSeverityCritical:
	default:
		return fmt.Errorf("unsupported identity observability alert severity %q", alert.Severity)
	}
	if alert.Height == 0 {
		return errors.New("identity observability alert height is required")
	}
	if alert.NameHash != "" {
		if err := validateHexHash("identity observability alert name_hash", alert.NameHash); err != nil {
			return err
		}
	}
	if alert.Name != "" {
		normalized, err := NormalizeAETDomain(alert.Name)
		if err != nil {
			return err
		}
		nameHash, err := DomainRecordV2NameHash(normalized)
		if err != nil {
			return err
		}
		if alert.NameHash != nameHash {
			return errors.New("identity observability alert name_hash mismatch")
		}
	}
	if alert.Metric != "" && !IsIdentityObservabilityMetricNameV2(alert.Metric) {
		return fmt.Errorf("unsupported identity observability alert metric %q", alert.Metric)
	}
	if alert.Details == "" || !isASCIIStringV2(alert.Details) {
		return errors.New("identity observability alert details are required ASCII")
	}
	if alert.AlertHash == "" || alert.AlertHash != ComputeIdentityObservabilityAlertHashV2(alert) {
		return errors.New("identity observability alert hash mismatch")
	}
	return nil
}

func ComputeIdentityObservabilitySpecHashV2(spec IdentityObservabilitySpecV2) string {
	parts := []string{"identity-observability-spec-v2"}
	for _, event := range sortedIdentityObservabilityEventsV2(spec.Events) {
		parts = append(parts, "event", string(event))
	}
	for _, metric := range sortedIdentityObservabilityMetricsV2(spec.Metrics) {
		parts = append(parts, "metric", string(metric))
	}
	for _, alert := range sortedIdentityObservabilityAlertsV2(spec.Alerts) {
		parts = append(parts, "alert", string(alert))
	}
	return identityHash(parts...)
}

func ComputeIdentityObservabilityEventHashV2(event IdentityObservabilityEventV2) string {
	parts := []string{"identity-observability-event-v2", string(event.Type), fmt.Sprint(event.Height), event.Name, event.NameHash, event.Actor}
	for _, key := range sortedStringMapKeysV2(event.Attributes) {
		parts = append(parts, key, event.Attributes[key])
	}
	return identityHash(parts...)
}

func ComputeIdentityMetricSampleHashV2(sample IdentityMetricSampleV2) string {
	parts := []string{"identity-observability-metric-sample-v2", string(sample.Name), fmt.Sprint(sample.Height), fmt.Sprint(sample.Value), sample.Unit}
	for _, key := range sortedStringMapKeysV2(sample.Labels) {
		parts = append(parts, key, sample.Labels[key])
	}
	return identityHash(parts...)
}

func ComputeIdentityObservabilityMetricsSnapshotHashV2(snapshot IdentityObservabilityMetricsSnapshotV2) string {
	parts := []string{"identity-observability-metrics-snapshot-v2", fmt.Sprint(snapshot.Height)}
	samples := append([]IdentityMetricSampleV2(nil), snapshot.Metrics...)
	sort.Slice(samples, func(i, j int) bool {
		if samples[i].Name != samples[j].Name {
			return samples[i].Name < samples[j].Name
		}
		return samples[i].SampleHash < samples[j].SampleHash
	})
	for _, sample := range samples {
		parts = append(parts, sample.SampleHash)
	}
	return identityHash(parts...)
}

func ComputeIdentityObservabilityAlertHashV2(alert IdentityObservabilityAlertV2) string {
	return identityHash("identity-observability-alert-v2", string(alert.Type), string(alert.Severity), fmt.Sprint(alert.Height), fmt.Sprint(alert.ObservedValue), fmt.Sprint(alert.Threshold), alert.Name, alert.NameHash, string(alert.Metric), alert.Details)
}

func IsIdentityObservabilityEventTypeV2(eventType IdentityObservabilityEventTypeV2) bool {
	switch eventType {
	case IdentityEventDomainCommitted, IdentityEventDomainRegistered, IdentityEventDomainRenewed,
		IdentityEventDomainTransferred, IdentityEventDomainExpired, IdentityEventDomainReleased,
		IdentityEventNFTBindingUpdated, IdentityEventResolverUpdated, IdentityEventReverseSet,
		IdentityEventReverseVerified, IdentityEventReverseInvalidated, IdentityEventSubdomainCreated,
		IdentityEventDelegationCreated, IdentityEventDelegationRevoked, IdentityEventZonePolicyUpdated,
		IdentityEventAuctionStarted, IdentityEventBidCommitted, IdentityEventBidRevealed,
		IdentityEventAuctionFinalized, IdentityEventCacheInvalidated:
		return true
	default:
		return false
	}
}

func IsIdentityObservabilityMetricNameV2(metric IdentityObservabilityMetricNameV2) bool {
	switch metric {
	case IdentityMetricActiveDomains, IdentityMetricExpiredDomains, IdentityMetricRenewalWindowDomains,
		IdentityMetricGracePeriodDomains, IdentityMetricResolverRecordCount, IdentityMetricAverageResolverPayloadSize,
		IdentityMetricReverseRecordsVerified, IdentityMetricReverseRecordsInvalidated, IdentityMetricSubdomainsByDepth,
		IdentityMetricDelegationRecordsActive, IdentityMetricAuctionsActive, IdentityMetricCommitmentsActive,
		IdentityMetricBatchResolverUpdateSize, IdentityMetricBlockSTMConflictRate, IdentityMetricStoreV2DirectReadLatency,
		IdentityMetricStoreV2RecursiveReadLatency, IdentityMetricStoreV2ResolverWriteLatency, IdentityMetricProofQueryLatency,
		IdentityMetricProofVerificationFailureCount, IdentityMetricExpiryProcessingBacklog:
		return true
	default:
		return false
	}
}

func IsIdentityObservabilityAlertTypeV2(alertType IdentityObservabilityAlertTypeV2) bool {
	switch alertType {
	case IdentityAlertNFTBindingMismatch, IdentityAlertResolverPayloadNearMaximum, IdentityAlertExpiryBacklogAboveThreshold,
		IdentityAlertProofQueryFailureSpike, IdentityAlertRegistrationSpamSpike, IdentityAlertResolverUpdateSpamSpike,
		IdentityAlertAuctionFinalizationBacklog, IdentityAlertBlockSTMConflictRateHigh, IdentityAlertStoreV2ReadLatencyHigh,
		IdentityAlertReverseMismatchSpike:
		return true
	default:
		return false
	}
}

func requiredIdentityObservabilityEventsV2() []IdentityObservabilityEventTypeV2 {
	return []IdentityObservabilityEventTypeV2{
		IdentityEventAuctionFinalized,
		IdentityEventAuctionStarted,
		IdentityEventBidCommitted,
		IdentityEventBidRevealed,
		IdentityEventCacheInvalidated,
		IdentityEventDelegationCreated,
		IdentityEventDelegationRevoked,
		IdentityEventDomainCommitted,
		IdentityEventDomainExpired,
		IdentityEventDomainRegistered,
		IdentityEventDomainReleased,
		IdentityEventDomainRenewed,
		IdentityEventDomainTransferred,
		IdentityEventNFTBindingUpdated,
		IdentityEventResolverUpdated,
		IdentityEventReverseInvalidated,
		IdentityEventReverseSet,
		IdentityEventReverseVerified,
		IdentityEventSubdomainCreated,
		IdentityEventZonePolicyUpdated,
	}
}

func requiredIdentityObservabilityMetricsV2() []IdentityObservabilityMetricNameV2 {
	return []IdentityObservabilityMetricNameV2{
		IdentityMetricActiveDomains,
		IdentityMetricAuctionsActive,
		IdentityMetricAverageResolverPayloadSize,
		IdentityMetricBatchResolverUpdateSize,
		IdentityMetricBlockSTMConflictRate,
		IdentityMetricCommitmentsActive,
		IdentityMetricDelegationRecordsActive,
		IdentityMetricExpiredDomains,
		IdentityMetricExpiryProcessingBacklog,
		IdentityMetricGracePeriodDomains,
		IdentityMetricProofQueryLatency,
		IdentityMetricProofVerificationFailureCount,
		IdentityMetricRenewalWindowDomains,
		IdentityMetricResolverRecordCount,
		IdentityMetricReverseRecordsInvalidated,
		IdentityMetricReverseRecordsVerified,
		IdentityMetricStoreV2DirectReadLatency,
		IdentityMetricStoreV2RecursiveReadLatency,
		IdentityMetricStoreV2ResolverWriteLatency,
		IdentityMetricSubdomainsByDepth,
	}
}

func requiredIdentityObservabilityAlertsV2() []IdentityObservabilityAlertTypeV2 {
	return []IdentityObservabilityAlertTypeV2{
		IdentityAlertAuctionFinalizationBacklog,
		IdentityAlertBlockSTMConflictRateHigh,
		IdentityAlertExpiryBacklogAboveThreshold,
		IdentityAlertNFTBindingMismatch,
		IdentityAlertProofQueryFailureSpike,
		IdentityAlertRegistrationSpamSpike,
		IdentityAlertResolverPayloadNearMaximum,
		IdentityAlertResolverUpdateSpamSpike,
		IdentityAlertReverseMismatchSpike,
		IdentityAlertStoreV2ReadLatencyHigh,
	}
}

func identityObservabilityMetricUnitV2(metric IdentityObservabilityMetricNameV2) string {
	switch metric {
	case IdentityMetricAverageResolverPayloadSize:
		return IdentityMetricUnitBytes
	case IdentityMetricBlockSTMConflictRate:
		return IdentityMetricUnitBasisPoints
	case IdentityMetricStoreV2DirectReadLatency, IdentityMetricStoreV2RecursiveReadLatency, IdentityMetricStoreV2ResolverWriteLatency, IdentityMetricProofQueryLatency:
		return IdentityMetricUnitMicroseconds
	default:
		return IdentityMetricUnitCount
	}
}

func observabilityAttributesV2(event IdentityObservabilityEventV2) []string {
	out := make([]string, 0, len(event.Attributes)+2)
	for _, key := range sortedStringMapKeysV2(event.Attributes) {
		out = append(out, key+"="+event.Attributes[key])
	}
	out = append(out, "event_hash="+event.EventHash)
	if event.Actor != "" {
		out = append(out, "actor="+event.Actor)
	}
	sort.Strings(out)
	return out
}

func estimateResolverRecordPayloadBytesV2(record ResolverRecord) uint64 {
	total := uint64(len(record.Domain) + len(record.Owner) + len(record.Primary) + len(record.Contract) + len(record.ZoneEndpoint) + len(record.Metadata))
	for key, address := range record.Records {
		total += uint64(len(key) + len(address))
	}
	return total
}

func subdomainDepthCountsV2(records []SubdomainRecord) map[uint64]uint64 {
	out := map[uint64]uint64{}
	for _, record := range records {
		normalized, err := NormalizeAETDomain(record.Name)
		if err != nil {
			continue
		}
		labels := strings.Split(strings.TrimSuffix(normalized, ".aet"), ".")
		depth := uint64(0)
		if len(labels) > 0 {
			depth = uint64(len(labels) - 1)
		}
		out[depth]++
	}
	return out
}

func sortedUintMapKeysV2(values map[uint64]uint64) []uint64 {
	keys := make([]uint64, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

func sortedIdentityObservabilityEventsV2(values []IdentityObservabilityEventTypeV2) []IdentityObservabilityEventTypeV2 {
	out := append([]IdentityObservabilityEventTypeV2(nil), values...)
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func sortedIdentityObservabilityMetricsV2(values []IdentityObservabilityMetricNameV2) []IdentityObservabilityMetricNameV2 {
	out := append([]IdentityObservabilityMetricNameV2(nil), values...)
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func sortedIdentityObservabilityAlertsV2(values []IdentityObservabilityAlertTypeV2) []IdentityObservabilityAlertTypeV2 {
	out := append([]IdentityObservabilityAlertTypeV2(nil), values...)
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func normalizeIdentityObservabilityAlertThresholdsV2(thresholds IdentityObservabilityAlertThresholdsV2) IdentityObservabilityAlertThresholdsV2 {
	defaults := DefaultIdentityObservabilityAlertThresholdsV2()
	if thresholds.ResolverPayloadNearMaxBps == 0 {
		thresholds.ResolverPayloadNearMaxBps = defaults.ResolverPayloadNearMaxBps
	}
	if thresholds.ExpiryBacklogThreshold == 0 {
		thresholds.ExpiryBacklogThreshold = defaults.ExpiryBacklogThreshold
	}
	if thresholds.ProofFailureSpikeThreshold == 0 {
		thresholds.ProofFailureSpikeThreshold = defaults.ProofFailureSpikeThreshold
	}
	if thresholds.RegistrationSpamSpikeThreshold == 0 {
		thresholds.RegistrationSpamSpikeThreshold = defaults.RegistrationSpamSpikeThreshold
	}
	if thresholds.ResolverUpdateSpamSpikeThreshold == 0 {
		thresholds.ResolverUpdateSpamSpikeThreshold = defaults.ResolverUpdateSpamSpikeThreshold
	}
	if thresholds.AuctionFinalizationBacklogThreshold == 0 {
		thresholds.AuctionFinalizationBacklogThreshold = defaults.AuctionFinalizationBacklogThreshold
	}
	if thresholds.BlockSTMConflictRateBpsThreshold == 0 {
		thresholds.BlockSTMConflictRateBpsThreshold = defaults.BlockSTMConflictRateBpsThreshold
	}
	if thresholds.StoreV2ReadLatencyMicrosThreshold == 0 {
		thresholds.StoreV2ReadLatencyMicrosThreshold = defaults.StoreV2ReadLatencyMicrosThreshold
	}
	if thresholds.ReverseMismatchSpikeThreshold == 0 {
		thresholds.ReverseMismatchSpikeThreshold = defaults.ReverseMismatchSpikeThreshold
	}
	return thresholds
}

func identityObservabilityMetricValuesV2(snapshot IdentityObservabilityMetricsSnapshotV2) map[IdentityObservabilityMetricNameV2]uint64 {
	out := map[IdentityObservabilityMetricNameV2]uint64{}
	for _, sample := range snapshot.Metrics {
		if sample.Value > out[sample.Name] {
			out[sample.Name] = sample.Value
		}
	}
	return out
}

func appendThresholdAlertV2(alerts []IdentityObservabilityAlertV2, height uint64, alertType IdentityObservabilityAlertTypeV2, severity IdentityObservabilityAlertSeverityV2, metric IdentityObservabilityMetricNameV2, observed uint64, threshold uint64, details string) ([]IdentityObservabilityAlertV2, error) {
	if observed < threshold {
		return alerts, nil
	}
	alert, err := NewIdentityObservabilityAlertV2(IdentityObservabilityAlertV2{
		Type:		alertType,
		Severity:	severity,
		Height:		height,
		ObservedValue:	observed,
		Threshold:	threshold,
		Metric:		metric,
		Details:	details,
	})
	if err != nil {
		return nil, err
	}
	return append(alerts, alert), nil
}

func auctionFinalizationBacklogV2(state IdentityState, height uint64) uint64 {
	var backlog uint64
	for _, auction := range state.Auctions {
		if auction.Phase != AuctionPhaseFinalized && auction.RevealEndHeight != 0 && auction.RevealEndHeight <= height {
			backlog++
		}
	}
	return backlog
}
