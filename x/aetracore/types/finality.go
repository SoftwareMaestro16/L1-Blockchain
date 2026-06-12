package types

import (
	"errors"
	"fmt"
)

type FinalityRecord struct {
	Height			uint64
	AppHash			string
	GlobalStateRoot		string
	GlobalMessageRoot	string
	ExecutionReceiptRoot	string
	ZoneCommitmentsRoot	string
	RoutingTableRoot	string
	CommitmentCount		uint64
	ProofRootCount		uint64
	CrossZoneFinalityDelay	uint64
	EligibleDeliveryHeight	uint64
	FinalityHash		string
}

func NewFinalityRecord(snapshot RootSnapshot, appHash string, commitmentCount uint64, routingTableRoot string, crossZoneDelay uint64) (FinalityRecord, error) {
	if err := snapshot.Validate(); err != nil {
		return FinalityRecord{}, err
	}
	if err := ValidateHash("aetracore finality app hash", appHash); err != nil {
		return FinalityRecord{}, err
	}
	if routingTableRoot != "" {
		if err := ValidateHash("aetracore finality routing table root", routingTableRoot); err != nil {
			return FinalityRecord{}, err
		}
	}
	if crossZoneDelay == 0 {
		crossZoneDelay = 1
	}
	record := FinalityRecord{
		Height:			snapshot.Height,
		AppHash:		appHash,
		GlobalStateRoot:	snapshot.Finality.GlobalStateRoot,
		GlobalMessageRoot:	snapshot.Finality.GlobalMessageRoot,
		ExecutionReceiptRoot:	snapshot.Finality.ExecutionReceiptRoot,
		RoutingTableRoot:	routingTableRoot,
		CommitmentCount:	commitmentCount,
		ProofRootCount:		uint64(len(snapshot.ProofRoots)),
		CrossZoneFinalityDelay:	crossZoneDelay,
		EligibleDeliveryHeight:	snapshot.Height + crossZoneDelay,
	}
	if record.CommitmentCount == 0 {
		record.CommitmentCount = uint64(snapshot.GlobalStateRoot.ZoneCount)
	}
	record.ZoneCommitmentsRoot = computeSnapshotZoneCommitmentsRoot(snapshot)
	record.FinalityHash = ComputeFinalityRecordHash(record)
	return record, record.ValidateHash()
}

func NewFinalityRecordFromKernelFinalization(finalization KernelFinalization, appHash string) (FinalityRecord, error) {
	if err := finalization.Validate(); err != nil {
		return FinalityRecord{}, err
	}
	routingRoot := ""
	for _, root := range finalization.RootSnapshot.ProofRoots {
		if root.RootType == RoutingTableRootType && root.RootHash != "" {
			routingRoot = root.RootHash
			break
		}
	}
	return NewFinalityRecord(finalization.RootSnapshot, appHash, finalization.CommitmentCount, routingRoot, 1)
}

func (r FinalityRecord) ValidateFormat() error {
	if r.Height == 0 {
		return errors.New("aetracore finality record height must be positive")
	}
	for _, field := range []struct {
		name	string
		value	string
	}{
		{"aetracore finality app hash", r.AppHash},
		{"aetracore finality global state root", r.GlobalStateRoot},
		{"aetracore finality global message root", r.GlobalMessageRoot},
		{"aetracore finality execution receipt root", r.ExecutionReceiptRoot},
		{"aetracore finality zone commitments root", r.ZoneCommitmentsRoot},
	} {
		if err := ValidateHash(field.name, field.value); err != nil {
			return err
		}
	}
	if r.RoutingTableRoot != "" {
		if err := ValidateHash("aetracore finality routing table root", r.RoutingTableRoot); err != nil {
			return err
		}
	}
	if r.CommitmentCount == 0 {
		return errors.New("aetracore finality commitment count must be positive")
	}
	if r.CrossZoneFinalityDelay == 0 {
		return errors.New("aetracore finality delay must be positive")
	}
	if r.EligibleDeliveryHeight <= r.Height {
		return errors.New("aetracore finality eligible delivery height must follow height")
	}
	if r.EligibleDeliveryHeight != r.Height+r.CrossZoneFinalityDelay {
		return errors.New("aetracore finality eligible delivery height mismatch")
	}
	if r.FinalityHash != "" {
		return ValidateHash("aetracore finality record hash", r.FinalityHash)
	}
	return nil
}

func (r FinalityRecord) ValidateHash() error {
	if err := r.ValidateFormat(); err != nil {
		return err
	}
	expected := ComputeFinalityRecordHash(r)
	if r.FinalityHash != expected {
		return fmt.Errorf("aetracore finality record hash mismatch: expected %s", expected)
	}
	return nil
}

func ComputeFinalityRecordHash(record FinalityRecord) string {
	return hashParts(
		"aetra-aek-finality-record-v1",
		fmt.Sprint(record.Height),
		record.AppHash,
		record.GlobalStateRoot,
		record.GlobalMessageRoot,
		record.ExecutionReceiptRoot,
		record.ZoneCommitmentsRoot,
		record.RoutingTableRoot,
		fmt.Sprint(record.CommitmentCount),
		fmt.Sprint(record.ProofRootCount),
		fmt.Sprint(record.CrossZoneFinalityDelay),
		fmt.Sprint(record.EligibleDeliveryHeight),
	)
}

func computeSnapshotZoneCommitmentsRoot(snapshot RootSnapshot) string {
	for _, root := range snapshot.ProofRoots {
		if root.RootType == ZoneCommitmentsRoot && root.ZoneID == "" {
			return root.RootHash
		}
	}
	parts := []string{"aetra-aek-finality-zone-commitments-v1", fmt.Sprint(snapshot.Height), fmt.Sprint(len(snapshot.ProofRoots))}
	for _, root := range snapshot.ProofRoots {
		if root.RootType == ZoneCommitmentsRoot {
			parts = append(parts, string(root.ZoneID), root.RootHash)
		}
	}
	return hashParts(parts...)
}
