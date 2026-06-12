package types

import (
	"errors"
	"fmt"
	"sort"
)

type ChannelID uint16

const (
	ChannelIDConsensus	ChannelID	= 0x01
	ChannelIDMempool	ChannelID	= 0x02
	ChannelIDBlock		ChannelID	= 0x03
	ChannelIDStateSync	ChannelID	= 0x04
	ChannelIDData		ChannelID	= 0x05
	ChannelIDExecution	ChannelID	= 0x06
	ChannelIDService	ChannelID	= 0x07
	ChannelIDRouting	ChannelID	= 0x08
	ChannelIDDiscovery	ChannelID	= 0x09
)

type BandwidthAccount struct {
	Channel		ChannelClass
	LimitBytes	uint64
	UsedBytes	uint64
}

type BandwidthLedger struct {
	Height		uint64
	Accounts	[]BandwidthAccount
}

type L0ChannelMetrics struct {
	Height			uint64
	Channel			ChannelClass
	ChannelID		ChannelID
	EnqueuedCount		uint64
	SentCount		uint64
	DroppedCount		uint64
	BytesEnqueued		uint64
	BytesSent		uint64
	ConsensusDelayBlocks	uint64
}

type L0AlertSeverity string

const (
	L0AlertInfo	L0AlertSeverity	= "INFO"
	L0AlertWarning	L0AlertSeverity	= "WARNING"
	L0AlertCritical	L0AlertSeverity	= "CRITICAL"
)

type L0Alert struct {
	Severity	L0AlertSeverity
	Channel		ChannelClass
	Code		string
	Message		string
}

type L0Schedule struct {
	Height	uint64
	Plans	[]PropagationPlan
	Dropped	[]TransportEnvelope
	Ledger	BandwidthLedger
	Metrics	[]L0ChannelMetrics
	Alerts	[]L0Alert
}

func ChannelIDForClass(channel ChannelClass) (ChannelID, error) {
	switch channel {
	case ChannelConsensus:
		return ChannelIDConsensus, nil
	case ChannelMempool:
		return ChannelIDMempool, nil
	case ChannelBlock:
		return ChannelIDBlock, nil
	case ChannelStateSync:
		return ChannelIDStateSync, nil
	case ChannelData:
		return ChannelIDData, nil
	case ChannelExecution:
		return ChannelIDExecution, nil
	case ChannelService:
		return ChannelIDService, nil
	case ChannelRouting:
		return ChannelIDRouting, nil
	case ChannelDiscovery:
		return ChannelIDDiscovery, nil
	default:
		return 0, fmt.Errorf("unknown networking channel %q", channel)
	}
}

func ChannelClassForID(id ChannelID) (ChannelClass, error) {
	switch id {
	case ChannelIDConsensus:
		return ChannelConsensus, nil
	case ChannelIDMempool:
		return ChannelMempool, nil
	case ChannelIDBlock:
		return ChannelBlock, nil
	case ChannelIDStateSync:
		return ChannelStateSync, nil
	case ChannelIDData:
		return ChannelData, nil
	case ChannelIDExecution:
		return ChannelExecution, nil
	case ChannelIDService:
		return ChannelService, nil
	case ChannelIDRouting:
		return ChannelRouting, nil
	case ChannelIDDiscovery:
		return ChannelDiscovery, nil
	default:
		return "", fmt.Errorf("unknown networking channel id %d", id)
	}
}

func NewBandwidthLedger(height uint64, bandwidth BandwidthPolicy, policies []ChannelPolicy) (BandwidthLedger, error) {
	if height == 0 {
		return BandwidthLedger{}, errors.New("networking L0 bandwidth height must be positive")
	}
	if err := bandwidth.Validate(); err != nil {
		return BandwidthLedger{}, err
	}
	if err := ValidateChannelPolicies(policies); err != nil {
		return BandwidthLedger{}, err
	}
	ledger := BandwidthLedger{
		Height:		height,
		Accounts:	make([]BandwidthAccount, len(policies)),
	}
	for i, policy := range policies {
		limit := (bandwidth.MaxOutboundBytesPerBlock * uint64(policy.BandwidthWeight)) / uint64(BasisPoints)
		if policy.Channel == ChannelConsensus {
			reserve := (bandwidth.MaxOutboundBytesPerBlock * uint64(bandwidth.ConsensusReserveBps)) / uint64(BasisPoints)
			if limit < reserve {
				limit = reserve
			}
		}
		if limit < policy.MaxMessageBytes {
			limit = policy.MaxMessageBytes
		}
		ledger.Accounts[i] = BandwidthAccount{Channel: policy.Channel, LimitBytes: limit}
	}
	sortBandwidthAccounts(ledger.Accounts)
	return ledger, ledger.Validate()
}

func (l BandwidthLedger) Validate() error {
	if l.Height == 0 {
		return errors.New("networking L0 bandwidth ledger height must be positive")
	}
	seen := make(map[ChannelClass]struct{}, len(l.Accounts))
	for _, account := range l.Accounts {
		if !IsChannelClass(account.Channel) {
			return fmt.Errorf("unknown networking bandwidth channel %q", account.Channel)
		}
		if _, found := seen[account.Channel]; found {
			return errors.New("networking duplicate bandwidth account")
		}
		seen[account.Channel] = struct{}{}
		if account.LimitBytes == 0 {
			return errors.New("networking bandwidth account limit must be positive")
		}
		if account.UsedBytes > account.LimitBytes {
			return errors.New("networking bandwidth account exceeds limit")
		}
	}
	return nil
}

func AccountBandwidth(ledger BandwidthLedger, envelope TransportEnvelope) (BandwidthLedger, error) {
	ledger = cloneBandwidthLedger(ledger)
	if err := ledger.Validate(); err != nil {
		return BandwidthLedger{}, err
	}
	envelope = envelope.Normalize()
	if err := envelope.Validate(DefaultChannelPolicies()); err != nil {
		return BandwidthLedger{}, err
	}
	for i, account := range ledger.Accounts {
		if account.Channel != envelope.Channel {
			continue
		}
		if account.UsedBytes+envelope.SizeBytes > account.LimitBytes {
			return BandwidthLedger{}, fmt.Errorf("networking bandwidth exhausted for %s", envelope.Channel)
		}
		ledger.Accounts[i].UsedBytes += envelope.SizeBytes
		return ledger, nil
	}
	return BandwidthLedger{}, fmt.Errorf("networking missing bandwidth account for %s", envelope.Channel)
}

func ScheduleL0Propagation(adapter AetherNetworkingAdapter, envelopes []TransportEnvelope, peerCount uint32, score PeerScore, height uint64) (L0Schedule, error) {
	if err := ValidateAetherNetworkingAdapter(adapter); err != nil {
		return L0Schedule{}, err
	}
	ledger, err := NewBandwidthLedger(height, adapter.Bandwidth, DefaultChannelPolicies())
	if err != nil {
		return L0Schedule{}, err
	}
	ordered := SortTransportEnvelopes(envelopes, DefaultChannelPolicies())
	metrics := newL0Metrics(height, envelopes)
	schedule := L0Schedule{Height: height, Ledger: ledger}
	for _, envelope := range ordered {
		plan, err := PlanPropagation(adapter, envelope, peerCount, score)
		if err != nil {
			return L0Schedule{}, err
		}
		nextLedger, err := AccountBandwidth(schedule.Ledger, envelope)
		if err != nil {
			schedule.Dropped = append(schedule.Dropped, envelope.Normalize())
			metrics = markL0Dropped(metrics, envelope)
			continue
		}
		schedule.Ledger = nextLedger
		schedule.Plans = append(schedule.Plans, plan)
		metrics = markL0Sent(metrics, envelope)
	}
	schedule.Metrics = sortL0Metrics(metrics)
	schedule.Alerts = EvaluateL0Alerts(schedule.Metrics)
	return schedule, schedule.Validate()
}

func (s L0Schedule) Validate() error {
	if s.Height == 0 {
		return errors.New("networking L0 schedule height must be positive")
	}
	if err := s.Ledger.Validate(); err != nil {
		return err
	}
	for _, plan := range s.Plans {
		if plan.Envelope.Channel == ChannelConsensus && !plan.HandledByCometBFT {
			return errors.New("networking L0 consensus propagation must remain handled by CometBFT")
		}
	}
	for _, alert := range s.Alerts {
		if !IsL0AlertSeverity(alert.Severity) {
			return fmt.Errorf("unknown networking L0 alert severity %q", alert.Severity)
		}
	}
	return nil
}

func EvaluateL0Alerts(metrics []L0ChannelMetrics) []L0Alert {
	alerts := make([]L0Alert, 0)
	for _, metric := range metrics {
		if metric.Channel == ChannelConsensus && metric.DroppedCount > 0 {
			alerts = append(alerts, L0Alert{
				Severity:	L0AlertCritical,
				Channel:	metric.Channel,
				Code:		"CONSENSUS_TRAFFIC_DROPPED",
				Message:	"consensus traffic must not be dropped or delayed by adapter scheduling",
			})
		}
		if metric.Channel == ChannelConsensus && metric.ConsensusDelayBlocks > 0 {
			alerts = append(alerts, L0Alert{
				Severity:	L0AlertCritical,
				Channel:	metric.Channel,
				Code:		"CONSENSUS_TRAFFIC_DELAYED",
				Message:	"consensus traffic delay was observed at L0",
			})
		}
		if metric.Channel != ChannelConsensus && metric.DroppedCount > 0 {
			alerts = append(alerts, L0Alert{
				Severity:	L0AlertWarning,
				Channel:	metric.Channel,
				Code:		"NON_CONSENSUS_BACKPRESSURE",
				Message:	"non-consensus traffic hit channel bandwidth limits",
			})
		}
	}
	sort.SliceStable(alerts, func(i, j int) bool {
		if alerts[i].Severity != alerts[j].Severity {
			return l0SeverityRank(alerts[i].Severity) < l0SeverityRank(alerts[j].Severity)
		}
		if alerts[i].Channel != alerts[j].Channel {
			return alerts[i].Channel < alerts[j].Channel
		}
		return alerts[i].Code < alerts[j].Code
	})
	return alerts
}

func IsL0AlertSeverity(severity L0AlertSeverity) bool {
	switch severity {
	case L0AlertInfo, L0AlertWarning, L0AlertCritical:
		return true
	default:
		return false
	}
}

func l0SeverityRank(severity L0AlertSeverity) uint8 {
	switch severity {
	case L0AlertCritical:
		return 0
	case L0AlertWarning:
		return 1
	case L0AlertInfo:
		return 2
	default:
		return 255
	}
}

func newL0Metrics(height uint64, envelopes []TransportEnvelope) []L0ChannelMetrics {
	byChannel := make(map[ChannelClass]L0ChannelMetrics)
	for _, policy := range DefaultChannelPolicies() {
		id, _ := ChannelIDForClass(policy.Channel)
		byChannel[policy.Channel] = L0ChannelMetrics{Height: height, Channel: policy.Channel, ChannelID: id}
	}
	for _, envelope := range envelopes {
		envelope = envelope.Normalize()
		metric := byChannel[envelope.Channel]
		metric.EnqueuedCount++
		metric.BytesEnqueued += envelope.SizeBytes
		byChannel[envelope.Channel] = metric
	}
	out := make([]L0ChannelMetrics, 0, len(byChannel))
	for _, metric := range byChannel {
		out = append(out, metric)
	}
	return sortL0Metrics(out)
}

func markL0Sent(metrics []L0ChannelMetrics, envelope TransportEnvelope) []L0ChannelMetrics {
	return updateL0Metric(metrics, envelope, func(metric L0ChannelMetrics) L0ChannelMetrics {
		metric.SentCount++
		metric.BytesSent += envelope.SizeBytes
		return metric
	})
}

func markL0Dropped(metrics []L0ChannelMetrics, envelope TransportEnvelope) []L0ChannelMetrics {
	return updateL0Metric(metrics, envelope, func(metric L0ChannelMetrics) L0ChannelMetrics {
		metric.DroppedCount++
		if envelope.Channel == ChannelConsensus {
			metric.ConsensusDelayBlocks++
		}
		return metric
	})
}

func updateL0Metric(metrics []L0ChannelMetrics, envelope TransportEnvelope, update func(L0ChannelMetrics) L0ChannelMetrics) []L0ChannelMetrics {
	for i, metric := range metrics {
		if metric.Channel == envelope.Channel {
			metrics[i] = update(metric)
			return metrics
		}
	}
	id, _ := ChannelIDForClass(envelope.Channel)
	metrics = append(metrics, update(L0ChannelMetrics{Height: 0, Channel: envelope.Channel, ChannelID: id}))
	return metrics
}

func sortL0Metrics(metrics []L0ChannelMetrics) []L0ChannelMetrics {
	out := append([]L0ChannelMetrics(nil), metrics...)
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].ChannelID < out[j].ChannelID
	})
	return out
}

func cloneBandwidthLedger(ledger BandwidthLedger) BandwidthLedger {
	ledger.Accounts = append([]BandwidthAccount(nil), ledger.Accounts...)
	return ledger
}

func sortBandwidthAccounts(accounts []BandwidthAccount) {
	sort.SliceStable(accounts, func(i, j int) bool {
		leftID, _ := ChannelIDForClass(accounts[i].Channel)
		rightID, _ := ChannelIDForClass(accounts[j].Channel)
		return leftID < rightID
	})
}
