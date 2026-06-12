package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"
)

type PaymentLockedByChannelType struct {
	ChannelType	ChannelType
	Amount		string
}

type PaymentObservabilityMetrics struct {
	Height					uint64
	ActiveChannels				uint64
	PendingCloses				uint64
	ActiveDisputes				uint64
	FinalizableChannels			uint64
	SettledChannelsPerBlock			uint64
	AverageChannelLifetime			uint64
	TotalLockedNaet				string
	LockedNaetByChannelType			[]PaymentLockedByChannelType
	ConditionalPromisesActive		uint64
	ConditionalPromisesExpired		uint64
	ConditionalPromisesResolved		uint64
	VirtualChannelsActive			uint64
	RoutingAdvertisementsActive		uint64
	FraudProofsSubmitted			uint64
	FraudProofsAccepted			uint64
	FraudProofsRejected			uint64
	PenaltiesApplied			uint64
	ReporterRewardsPaid			uint64
	SettlementFeesCollected			string
	ChannelOpenFeeAverage			string
	DisputeInclusionLatency			uint64
	ChallengePeriodNearExpiryCount		uint64
	BlockSTMConflictRateBps			uint64
	StoreV2PaymentModuleReadLatencyOps	uint64
	StoreV2PaymentModuleWriteLatencyOps	uint64
	MetricsHash				string
}

func BuildPaymentObservabilityMetrics(state PaymentsState, fraud FraudProofVerificationState, profile BlockSTMConflictProfile, currentHeight uint64, nearExpiryThreshold uint64) (PaymentObservabilityMetrics, error) {
	if currentHeight == 0 {
		return PaymentObservabilityMetrics{}, errors.New("payments observability metrics height must be positive")
	}
	state = state.Export()
	if err := state.Validate(); err != nil {
		return PaymentObservabilityMetrics{}, err
	}
	fraud = fraud.Export()
	if err := fraud.Validate(); err != nil {
		return PaymentObservabilityMetrics{}, err
	}
	layout, err := BuildStoreV2Layout(state)
	if err != nil {
		return PaymentObservabilityMetrics{}, err
	}
	alerts, err := MonitorNearExpiryDisputes(state, currentHeight, nearExpiryThreshold)
	if err != nil {
		return PaymentObservabilityMetrics{}, err
	}
	metrics := PaymentObservabilityMetrics{
		Height:				currentHeight,
		VirtualChannelsActive:		uint64(len(state.VirtualChannels)),
		FraudProofsAccepted:		uint64(len(fraud.EvidenceRecords)),
		PenaltiesApplied:		uint64(len(fraud.PenaltyRecords)),
		ChallengePeriodNearExpiryCount:	uint64(len(alerts)),
	}
	lockedByType := map[ChannelType]sdkmath.Int{}
	totalLocked := sdkmath.ZeroInt()
	openFeeTotal := sdkmath.ZeroInt()
	openFeeCount := uint64(0)
	feeTotal := sdkmath.ZeroInt()
	lifetimeTotal := uint64(0)
	lifetimeCount := uint64(0)
	for _, channel := range state.Channels {
		channel = channel.Normalize()
		switch channel.Status {
		case ChannelStatusOpen:
			metrics.ActiveChannels++
		case ChannelStatusPendingClose:
			metrics.PendingCloses++
			if channel.PendingClose.DisputeCount > 0 || channel.Finality == ChannelFinalityInDispute {
				metrics.ActiveDisputes++
			}
			if finality := FinalityAfterPendingClose(channel, currentHeight); finality == ChannelFinalityFinalizable {
				metrics.FinalizableChannels++
			}
		case ChannelStatusSettled:
			if channel.Finality == ChannelFinalityPenalized {
				metrics.PenaltiesApplied++
			}
		}
		if channel.RoutingAdvertised {
			metrics.RoutingAdvertisementsActive++
		}
		for _, condition := range channel.LatestState.Normalize().Conditions {
			condition = condition.Normalize()
			if !conditionWasClaimed(condition.ConditionID, state.ConditionClaims) && (condition.TimeoutHeight == 0 || condition.TimeoutHeight > currentHeight) {
				metrics.ConditionalPromisesActive++
			}
		}
	}
	for _, lock := range state.CustodyLocks {
		lock = lock.Normalize()
		if lock.Denom != NativeDenom {
			continue
		}
		amount, err := parseNonNegativeInt("payments observability locked naet", lock.Amount)
		if err != nil {
			return PaymentObservabilityMetrics{}, err
		}
		totalLocked = totalLocked.Add(amount)
		if channel, found := state.ChannelByID(lock.ChannelID); found {
			channelType := channel.Normalize().ChannelType
			current, ok := lockedByType[channelType]
			if !ok {
				current = sdkmath.ZeroInt()
			}
			lockedByType[channelType] = current.Add(amount)
		}
	}
	for _, settlement := range state.Settlements {
		settlement = settlement.Normalize()
		if settlement.SettledHeight == currentHeight {
			metrics.SettledChannelsPerBlock++
		}
		if channel, found := state.ChannelByID(settlement.ChannelID); found && settlement.SettledHeight >= channel.OpenHeight {
			lifetimeTotal += settlement.SettledHeight - channel.OpenHeight
			lifetimeCount++
		}
	}
	for _, claim := range state.ConditionClaims {
		claim = claim.Normalize()
		if claim.PreimageHash == "" || strings.Contains(claim.EvidenceHash, "expiry") {
			metrics.ConditionalPromisesExpired++
		} else {
			metrics.ConditionalPromisesResolved++
		}
	}
	for _, charge := range state.FeeCharges {
		charge = charge.Normalize()
		amount, err := parseNonNegativeInt("payments observability fee charge", charge.Amount)
		if err != nil {
			return PaymentObservabilityMetrics{}, err
		}
		feeTotal = feeTotal.Add(amount)
		if charge.FeeClass == PaymentFeeClassChannelOpen {
			openFeeTotal = openFeeTotal.Add(amount)
			openFeeCount++
		}
	}
	for _, reward := range fraud.ReporterRewards {
		if reward.Normalize().Claimed {
			metrics.ReporterRewardsPaid++
		}
	}
	rejected := uint64(0)
	for _, event := range state.Events {
		if event.Normalize().EventType == string(PaymentAPIEventFraudProofRejected) {
			rejected++
		}
	}
	metrics.FraudProofsRejected = rejected
	metrics.FraudProofsSubmitted = metrics.FraudProofsAccepted + metrics.FraudProofsRejected
	metrics.TotalLockedNaet = totalLocked.String()
	metrics.LockedNaetByChannelType = lockedNaetByChannelTypeSlice(lockedByType)
	metrics.SettlementFeesCollected = feeTotal.String()
	if openFeeCount > 0 {
		metrics.ChannelOpenFeeAverage = openFeeTotal.QuoRaw(int64(openFeeCount)).String()
	} else {
		metrics.ChannelOpenFeeAverage = "0"
	}
	if lifetimeCount > 0 {
		metrics.AverageChannelLifetime = lifetimeTotal / lifetimeCount
	}
	metrics.DisputeInclusionLatency = averageDisputeInclusionLatency(state.InclusionLatencies)
	metrics.BlockSTMConflictRateBps = blockSTMConflictRateBps(profile)
	metrics.StoreV2PaymentModuleReadLatencyOps = storeV2ReadLatencyOps(layout)
	metrics.StoreV2PaymentModuleWriteLatencyOps = storeV2WriteLatencyOps(layout)
	metrics = metrics.Normalize()
	metrics.MetricsHash = ComputePaymentObservabilityMetricsHash(metrics)
	return metrics, metrics.Validate()
}

func ComputePaymentObservabilityMetricsHash(metrics PaymentObservabilityMetrics) string {
	metrics = metrics.Normalize()
	parts := []string{
		"payments-observability-metrics-v1",
		fmt.Sprintf("%020d", metrics.Height),
		fmt.Sprintf("%020d", metrics.ActiveChannels),
		fmt.Sprintf("%020d", metrics.PendingCloses),
		fmt.Sprintf("%020d", metrics.ActiveDisputes),
		fmt.Sprintf("%020d", metrics.FinalizableChannels),
		fmt.Sprintf("%020d", metrics.SettledChannelsPerBlock),
		fmt.Sprintf("%020d", metrics.AverageChannelLifetime),
		metrics.TotalLockedNaet,
		fmt.Sprintf("%020d", metrics.ConditionalPromisesActive),
		fmt.Sprintf("%020d", metrics.ConditionalPromisesExpired),
		fmt.Sprintf("%020d", metrics.ConditionalPromisesResolved),
		fmt.Sprintf("%020d", metrics.VirtualChannelsActive),
		fmt.Sprintf("%020d", metrics.RoutingAdvertisementsActive),
		fmt.Sprintf("%020d", metrics.FraudProofsSubmitted),
		fmt.Sprintf("%020d", metrics.FraudProofsAccepted),
		fmt.Sprintf("%020d", metrics.FraudProofsRejected),
		fmt.Sprintf("%020d", metrics.PenaltiesApplied),
		fmt.Sprintf("%020d", metrics.ReporterRewardsPaid),
		metrics.SettlementFeesCollected,
		metrics.ChannelOpenFeeAverage,
		fmt.Sprintf("%020d", metrics.DisputeInclusionLatency),
		fmt.Sprintf("%020d", metrics.ChallengePeriodNearExpiryCount),
		fmt.Sprintf("%020d", metrics.BlockSTMConflictRateBps),
		fmt.Sprintf("%020d", metrics.StoreV2PaymentModuleReadLatencyOps),
		fmt.Sprintf("%020d", metrics.StoreV2PaymentModuleWriteLatencyOps),
	}
	for _, item := range metrics.LockedNaetByChannelType {
		item = item.Normalize()
		parts = append(parts, string(item.ChannelType), item.Amount)
	}
	return HashParts(parts...)
}

func (m PaymentObservabilityMetrics) Normalize() PaymentObservabilityMetrics {
	m.TotalLockedNaet = strings.TrimSpace(m.TotalLockedNaet)
	if m.TotalLockedNaet == "" {
		m.TotalLockedNaet = "0"
	}
	m.SettlementFeesCollected = strings.TrimSpace(m.SettlementFeesCollected)
	if m.SettlementFeesCollected == "" {
		m.SettlementFeesCollected = "0"
	}
	m.ChannelOpenFeeAverage = strings.TrimSpace(m.ChannelOpenFeeAverage)
	if m.ChannelOpenFeeAverage == "" {
		m.ChannelOpenFeeAverage = "0"
	}
	for i := range m.LockedNaetByChannelType {
		m.LockedNaetByChannelType[i] = m.LockedNaetByChannelType[i].Normalize()
	}
	sort.SliceStable(m.LockedNaetByChannelType, func(i, j int) bool {
		return m.LockedNaetByChannelType[i].ChannelType < m.LockedNaetByChannelType[j].ChannelType
	})
	m.MetricsHash = normalizeOptionalHash(m.MetricsHash)
	return m
}

func (m PaymentObservabilityMetrics) Validate() error {
	metrics := m.Normalize()
	if metrics.Height == 0 {
		return errors.New("payments observability metrics height must be positive")
	}
	if err := validateNonNegativeInt("payments observability total locked naet", metrics.TotalLockedNaet); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments observability fees collected", metrics.SettlementFeesCollected); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments observability open fee average", metrics.ChannelOpenFeeAverage); err != nil {
		return err
	}
	if metrics.FraudProofsSubmitted < metrics.FraudProofsAccepted+metrics.FraudProofsRejected {
		return errors.New("payments observability fraud submitted count is inconsistent")
	}
	seenTypes := map[ChannelType]struct{}{}
	for _, item := range metrics.LockedNaetByChannelType {
		item = item.Normalize()
		if !IsChannelType(item.ChannelType) {
			return fmt.Errorf("unknown payments observability channel type %q", item.ChannelType)
		}
		if _, duplicate := seenTypes[item.ChannelType]; duplicate {
			return fmt.Errorf("duplicate payments observability channel type %q", item.ChannelType)
		}
		seenTypes[item.ChannelType] = struct{}{}
		if err := validateNonNegativeInt("payments observability locked by type", item.Amount); err != nil {
			return err
		}
	}
	if err := ValidateHash("payments observability metrics hash", metrics.MetricsHash); err != nil {
		return err
	}
	if expected := ComputePaymentObservabilityMetricsHash(metrics); metrics.MetricsHash != expected {
		return errors.New("payments observability metrics hash mismatch")
	}
	return nil
}

func (l PaymentLockedByChannelType) Normalize() PaymentLockedByChannelType {
	l.Amount = strings.TrimSpace(l.Amount)
	if l.Amount == "" {
		l.Amount = "0"
	}
	return l
}

func lockedNaetByChannelTypeSlice(values map[ChannelType]sdkmath.Int) []PaymentLockedByChannelType {
	out := make([]PaymentLockedByChannelType, 0, len(values))
	for channelType, amount := range values {
		out = append(out, PaymentLockedByChannelType{ChannelType: channelType, Amount: amount.String()}.Normalize())
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ChannelType < out[j].ChannelType })
	return out
}

func conditionWasClaimed(conditionID string, claims []ConditionClaimRecord) bool {
	conditionID = normalizeHash(conditionID)
	for _, claim := range claims {
		if claim.Normalize().ConditionID == conditionID {
			return true
		}
	}
	return false
}

func averageDisputeInclusionLatency(records []SettlementInclusionLatency) uint64 {
	total := uint64(0)
	count := uint64(0)
	for _, record := range records {
		record = record.Normalize()
		if record.Operation == SettlementArbitrationDispute || record.Operation == SettlementArbitrationFraudProof {
			total += record.LatencyBlocks
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return total / count
}

func blockSTMConflictRateBps(profile BlockSTMConflictProfile) uint64 {
	planCount := uint64(len(profile.Plans))
	if planCount < 2 {
		return 0
	}
	pairs := planCount * (planCount - 1) / 2
	conflictingPairs := map[string]struct{}{}
	for _, conflict := range profile.Conflicts {
		conflict = conflict.Normalize()
		left := conflict.LeftOperationID
		right := conflict.RightOperationID
		if left > right {
			left, right = right, left
		}
		conflictingPairs[left+"|"+right] = struct{}{}
	}
	return uint64(len(conflictingPairs)) * 10_000 / pairs
}

func storeV2ReadLatencyOps(layout StoreV2Layout) uint64 {
	layout = layout.Normalize()
	return uint64(len(layout.Channels) + len(layout.ChannelStates) + len(layout.PendingCloses) + len(layout.Conditions) + len(layout.VirtualChannels) + len(layout.ParticipantChannels) + len(layout.SettlementTombstones) + len(layout.FeeAccumulators) + len(layout.FraudProofs))
}

func storeV2WriteLatencyOps(layout StoreV2Layout) uint64 {
	layout = layout.Normalize()
	return uint64(len(layout.Channels)*2 + len(layout.ChannelStates) + len(layout.PendingCloses) + len(layout.Conditions) + len(layout.VirtualChannels) + len(layout.ParticipantChannels) + len(layout.SettlementTombstones) + len(layout.FeeAccumulators) + len(layout.FraudProofs))
}
