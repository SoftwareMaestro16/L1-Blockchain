package types

import (
	"fmt"
	"strings"
)

const (
	ModuleName	= "payments"
	StoreKey	= ModuleName

	PaymentsKeyChannelPrefix		= "channel"
	PaymentsKeyParticipantIndexPrefix	= "participant_index"
	PaymentsKeyPendingCloseIndexPrefix	= "pending_close"
	PaymentsKeyConditionIndexPrefix		= "condition"
	PaymentsKeyRoutingAdIndexPrefix		= "routing_ad"
	PaymentsKeySettlementTombstonePrefix	= "settlement_tombstone"
	PaymentsKeySettlementPrefix		= "settlement"
	PaymentsKeyCustodyPrefix		= "custody"
	PaymentsKeyBlockAccumulatorPrefix	= "block_accumulator"

	StoreV2MigrationVersion	= uint64(2)

	StoreV2KeyChannelsPrefix		= "channels"
	StoreV2KeyChannelStatesPrefix		= "channel_states"
	StoreV2KeyPendingClosesPrefix		= "pending_closes"
	StoreV2KeyConditionsPrefix		= "conditions"
	StoreV2KeyVirtualChannelsPrefix		= "virtual_channels"
	StoreV2KeyParticipantChannelsPrefix	= "participant_channels"
	StoreV2KeySettlementTombstonesPrefix	= "settlement_tombstones"
	StoreV2KeyFeeAccumulatorsPrefix		= "fee_accumulators"
	StoreV2KeyFraudProofsPrefix		= "fraud_proofs"
	StoreV2KeyAdaptiveSnapshotsPrefix	= "adaptive_snapshots"
	StoreV2KeyActiveDisputesPrefix		= "active_disputes"
	StoreV2KeyPendingFinalizationsPrefix	= "pending_finalizations"
	StoreV2KeyWatcherReplayEventsPrefix	= "watcher_replay_events"
)

func PaymentChannelKey(channelID string) string {
	return paymentKey(PaymentsKeyChannelPrefix, normalizeHash(channelID))
}

func PaymentParticipantIndexKey(participant, channelID string) string {
	return paymentKey(PaymentsKeyParticipantIndexPrefix, strings.TrimSpace(participant), normalizeHash(channelID))
}

func PaymentPendingCloseIndexKey(channelID string) string {
	return paymentKey(PaymentsKeyPendingCloseIndexPrefix, normalizeHash(channelID))
}

func PaymentConditionIndexKey(channelID, conditionID string) string {
	return paymentKey(PaymentsKeyConditionIndexPrefix, normalizeHash(channelID), normalizeHash(conditionID))
}

func PaymentRoutingAdvertisementIndexKey(channelID string) string {
	return paymentKey(PaymentsKeyRoutingAdIndexPrefix, normalizeHash(channelID))
}

func PaymentSettlementTombstoneKey(channelID string) string {
	return paymentKey(PaymentsKeySettlementTombstonePrefix, normalizeHash(channelID))
}

func PaymentSettlementKey(channelID string) string {
	return paymentKey(PaymentsKeySettlementPrefix, normalizeHash(channelID))
}

func PaymentCustodyKey(channelID string) string {
	return paymentKey(PaymentsKeyCustodyPrefix, normalizeHash(channelID))
}

func PaymentBlockAccumulatorKey(blockHeight uint64) string {
	return paymentKey(PaymentsKeyBlockAccumulatorPrefix, fmt.Sprintf("%020d", blockHeight))
}

func StoreV2ChannelKey(channelID string) string {
	return paymentKey(StoreV2KeyChannelsPrefix, normalizeHash(channelID))
}

func StoreV2ChannelStateKey(channelID string, nonce uint64) string {
	return paymentKey(StoreV2KeyChannelStatesPrefix, normalizeHash(channelID), fmt.Sprintf("%020d", nonce))
}

func StoreV2PendingCloseKey(channelID string) string {
	return paymentKey(StoreV2KeyPendingClosesPrefix, normalizeHash(channelID))
}

func StoreV2ConditionKey(conditionID string) string {
	return paymentKey(StoreV2KeyConditionsPrefix, normalizeHash(conditionID))
}

func StoreV2VirtualChannelKey(virtualChannelID string) string {
	return paymentKey(StoreV2KeyVirtualChannelsPrefix, normalizeHash(virtualChannelID))
}

func StoreV2ParticipantChannelKey(address, channelID string) string {
	return paymentKey(StoreV2KeyParticipantChannelsPrefix, strings.TrimSpace(address), normalizeHash(channelID))
}

func StoreV2ParticipantChannelPrefix(address string) string {
	return paymentKey(StoreV2KeyParticipantChannelsPrefix, strings.TrimSpace(address))
}

func StoreV2SettlementTombstoneKey(channelID string) string {
	return paymentKey(StoreV2KeySettlementTombstonesPrefix, normalizeHash(channelID))
}

func StoreV2FeeAccumulatorKey(blockOrEpoch, bucket string) string {
	return paymentKey(StoreV2KeyFeeAccumulatorsPrefix, strings.TrimSpace(blockOrEpoch), strings.TrimSpace(bucket))
}

func StoreV2FraudProofKey(proofID string) string {
	return paymentKey(StoreV2KeyFraudProofsPrefix, normalizeHash(proofID))
}

func StoreV2AdaptiveSnapshotKey(height uint64) string {
	return paymentKey(StoreV2KeyAdaptiveSnapshotsPrefix, fmt.Sprintf("%020d", height))
}

func StoreV2ActiveDisputeKey(channelID string) string {
	return paymentKey(StoreV2KeyActiveDisputesPrefix, normalizeHash(channelID))
}

func StoreV2PendingFinalizationKey(channelID string) string {
	return paymentKey(StoreV2KeyPendingFinalizationsPrefix, normalizeHash(channelID))
}

func StoreV2WatcherReplayEventKey(height uint64, eventID string) string {
	return paymentKey(StoreV2KeyWatcherReplayEventsPrefix, fmt.Sprintf("%020d", height), normalizeHash(eventID))
}

func paymentKey(parts ...string) string {
	out := []string{ModuleName}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return strings.Join(out, "/")
}
