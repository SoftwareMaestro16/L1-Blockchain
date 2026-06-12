package types

import (
	"errors"
	"fmt"
)

type ZoneExecutionSummary struct {
	Height			uint64
	ZoneID			ZoneID
	LocalTxCount		uint64
	InboundMessageCount	uint64
	OutboxMessageCount	uint64
	ReceiptCount		uint64
	FailedReceiptCount	uint64
	GasUsed			uint64
	MessageRoot		string
	ReceiptsRoot		string
	EventsRoot		string
	SummaryHash		string
}

func CollectZoneExecutionSummary(height uint64, zoneID ZoneID, envelopes []KernelMessageEnvelope, receipts []ExecutionReceipt, gasUsed uint64, eventsRoot string) (ZoneExecutionSummary, error) {
	if height == 0 {
		return ZoneExecutionSummary{}, errors.New("aetracore execution summary height must be positive")
	}
	if err := ValidateZoneID(zoneID); err != nil {
		return ZoneExecutionSummary{}, err
	}
	if eventsRoot == "" {
		eventsRoot = EmptyRootHash
	}
	if err := ValidateHash("aetracore execution summary events root", eventsRoot); err != nil {
		return ZoneExecutionSummary{}, err
	}
	zoneEnvelopes := make([]KernelMessageEnvelope, 0)
	for _, envelope := range normalizeKernelEnvelopes(envelopes) {
		switch {
		case envelope.DestinationZone == zoneID:
			zoneEnvelopes = append(zoneEnvelopes, envelope)
		case envelope.SourceZone == zoneID && envelope.DestinationZone != zoneID:
			zoneEnvelopes = append(zoneEnvelopes, envelope)
		}
	}
	zoneReceipts := make([]ExecutionReceipt, 0)
	for _, receipt := range receipts {
		if receipt.SourceZone == zoneID || receipt.DestinationZone == zoneID {
			zoneReceipts = append(zoneReceipts, receipt)
		}
	}
	receiptsRoot, err := ComputeExecutionReceiptsRoot(zoneReceipts)
	if err != nil {
		return ZoneExecutionSummary{}, err
	}
	summary := ZoneExecutionSummary{
		Height:			height,
		ZoneID:			zoneID,
		LocalTxCount:		countZoneLocalEnvelopes(zoneID, zoneEnvelopes),
		InboundMessageCount:	countZoneInboundEnvelopes(zoneID, zoneEnvelopes),
		OutboxMessageCount:	countZoneOutboxEnvelopes(zoneID, zoneEnvelopes),
		ReceiptCount:		uint64(len(zoneReceipts)),
		FailedReceiptCount:	countFailedExecutionReceipts(zoneReceipts),
		GasUsed:		gasUsed,
		MessageRoot:		ComputeKernelEnvelopeRoot(zoneEnvelopes),
		ReceiptsRoot:		receiptsRoot,
		EventsRoot:		eventsRoot,
	}
	summary.SummaryHash = ComputeZoneExecutionSummaryHash(summary)
	return summary, summary.Validate()
}

func NewZoneCommitmentFromSummary(summary ZoneExecutionSummary, stateRoot string, inboxRoot string, paramsHash string) (ZoneCommitment, error) {
	return NewZoneCommitmentFromSummaryWithShardRoots(summary, stateRoot, inboxRoot, EmptyRootHash, paramsHash)
}

func NewZoneCommitmentFromSummaryWithShardRoots(summary ZoneExecutionSummary, stateRoot string, inboxRoot string, shardRootsRoot string, paramsHash string) (ZoneCommitment, error) {
	if err := summary.Validate(); err != nil {
		return ZoneCommitment{}, err
	}
	return NewZoneCommitment(
		summary.Height,
		summary.ZoneID,
		stateRoot,
		inboxRoot,
		summary.MessageRoot,
		summary.ReceiptsRoot,
		summary.EventsRoot,
		shardRootsRoot,
		paramsHash,
		summary.SummaryHash,
	)
}

func (s ZoneExecutionSummary) Validate() error {
	if s.Height == 0 {
		return errors.New("aetracore execution summary height must be positive")
	}
	if err := ValidateZoneID(s.ZoneID); err != nil {
		return err
	}
	if err := ValidateHash("aetracore execution summary message root", s.MessageRoot); err != nil {
		return err
	}
	if err := ValidateHash("aetracore execution summary receipts root", s.ReceiptsRoot); err != nil {
		return err
	}
	if err := ValidateHash("aetracore execution summary events root", s.EventsRoot); err != nil {
		return err
	}
	if s.FailedReceiptCount > s.ReceiptCount {
		return errors.New("aetracore execution summary failed receipts exceed receipt count")
	}
	if err := ValidateHash("aetracore execution summary hash", s.SummaryHash); err != nil {
		return err
	}
	if expected := ComputeZoneExecutionSummaryHash(s); expected != s.SummaryHash {
		return fmt.Errorf("aetracore execution summary hash mismatch: expected %s", expected)
	}
	return nil
}

func ComputeZoneExecutionSummaryHash(summary ZoneExecutionSummary) string {
	return hashParts(
		"aetra-aek-zone-execution-summary-v1",
		fmt.Sprint(summary.Height),
		string(summary.ZoneID),
		fmt.Sprint(summary.LocalTxCount),
		fmt.Sprint(summary.InboundMessageCount),
		fmt.Sprint(summary.OutboxMessageCount),
		fmt.Sprint(summary.ReceiptCount),
		fmt.Sprint(summary.FailedReceiptCount),
		fmt.Sprint(summary.GasUsed),
		summary.MessageRoot,
		summary.ReceiptsRoot,
		summary.EventsRoot,
	)
}

func countZoneLocalEnvelopes(zoneID ZoneID, envelopes []KernelMessageEnvelope) uint64 {
	var count uint64
	for _, envelope := range envelopes {
		if envelope.Kind == KernelMessageLocalTx && envelope.DestinationZone == zoneID {
			count++
		}
	}
	return count
}

func countZoneInboundEnvelopes(zoneID ZoneID, envelopes []KernelMessageEnvelope) uint64 {
	var count uint64
	for _, envelope := range envelopes {
		if envelope.Kind == KernelMessageRoutedInbound && envelope.DestinationZone == zoneID {
			count++
		}
	}
	return count
}

func countZoneOutboxEnvelopes(zoneID ZoneID, envelopes []KernelMessageEnvelope) uint64 {
	var count uint64
	for _, envelope := range envelopes {
		if envelope.SourceZone == zoneID && envelope.DestinationZone != zoneID {
			count++
		}
	}
	return count
}

func countFailedExecutionReceipts(receipts []ExecutionReceipt) uint64 {
	var count uint64
	for _, receipt := range receipts {
		if receipt.Status != ReceiptStatusSuccess {
			count++
		}
	}
	return count
}
