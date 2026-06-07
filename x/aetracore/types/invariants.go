package types

import (
	"errors"
	"fmt"
)

func ValidateRootAggregationInvariants(state CoreState) error {
	state = state.Export()
	if err := state.Validate(); err != nil {
		return err
	}
	for _, root := range state.GlobalRoots {
		zonesRoot, err := ComputeZoneCommitmentsRoot(root.Height, state.CommitmentsAtHeight(root.Height))
		if err != nil {
			return err
		}
		if root.ZonesRoot != zonesRoot {
			return fmt.Errorf("aetracore invariant zones root mismatch at height %d", root.Height)
		}
		servicesRoot, err := ComputeServiceRoot(state.ServiceDescriptors)
		if err != nil {
			return err
		}
		if root.ServicesRoot != servicesRoot {
			return fmt.Errorf("aetracore invariant services root mismatch at height %d", root.Height)
		}
		if expected := ComputeGlobalStateRootHash(root); root.GlobalRoot != expected {
			return fmt.Errorf("aetracore invariant global root hash mismatch at height %d", root.Height)
		}
	}
	for _, snapshot := range state.RootSnapshots {
		root, found := state.GlobalRootByHeight(snapshot.Height)
		if !found {
			return fmt.Errorf("aetracore invariant snapshot missing global root at height %d", snapshot.Height)
		}
		if snapshot.Finality.GlobalStateRoot != root.GlobalRoot {
			return fmt.Errorf("aetracore invariant snapshot finality root mismatch at height %d", snapshot.Height)
		}
		if snapshot.Finality.GlobalMessageRoot != root.MessageRoot {
			return fmt.Errorf("aetracore invariant snapshot message root mismatch at height %d", snapshot.Height)
		}
		if snapshot.Finality.ExecutionReceiptRoot != root.ReceiptsRoot {
			return fmt.Errorf("aetracore invariant snapshot receipts root mismatch at height %d", snapshot.Height)
		}
		if snapshot.GlobalStateRoot.ZoneCount != uint32(len(state.CommitmentsAtHeight(snapshot.Height))) {
			return fmt.Errorf("aetracore invariant snapshot zone count mismatch at height %d", snapshot.Height)
		}
	}
	for _, record := range state.FinalityRecords {
		snapshot, found := state.RootSnapshotAtHeight(record.Height)
		if !found {
			return fmt.Errorf("aetracore invariant finality record missing snapshot at height %d", record.Height)
		}
		if record.GlobalStateRoot != snapshot.Finality.GlobalStateRoot {
			return fmt.Errorf("aetracore invariant finality record global root mismatch at height %d", record.Height)
		}
		if record.GlobalMessageRoot != snapshot.Finality.GlobalMessageRoot {
			return fmt.Errorf("aetracore invariant finality record message root mismatch at height %d", record.Height)
		}
		if record.ExecutionReceiptRoot != snapshot.Finality.ExecutionReceiptRoot {
			return fmt.Errorf("aetracore invariant finality record receipt root mismatch at height %d", record.Height)
		}
	}
	for _, manifest := range state.ExportManifests {
		if err := ValidateKernelImport(state, manifest); err != nil {
			return err
		}
	}
	return nil
}

func AssertReplayIdenticalRoots(left CoreState, right CoreState) error {
	left = left.Export()
	right = right.Export()
	if err := ValidateRootAggregationInvariants(left); err != nil {
		return err
	}
	if err := ValidateRootAggregationInvariants(right); err != nil {
		return err
	}
	if len(left.GlobalRoots) != len(right.GlobalRoots) {
		return errors.New("aetracore replay root count mismatch")
	}
	for i := range left.GlobalRoots {
		if left.GlobalRoots[i] != right.GlobalRoots[i] {
			return fmt.Errorf("aetracore replay global root mismatch at index %d", i)
		}
	}
	if len(left.RootSnapshots) != len(right.RootSnapshots) {
		return errors.New("aetracore replay snapshot count mismatch")
	}
	for i := range left.RootSnapshots {
		if left.RootSnapshots[i].Finality != right.RootSnapshots[i].Finality {
			return fmt.Errorf("aetracore replay finality root mismatch at index %d", i)
		}
	}
	if len(left.FinalityRecords) != len(right.FinalityRecords) {
		return errors.New("aetracore replay finality record count mismatch")
	}
	for i := range left.FinalityRecords {
		if left.FinalityRecords[i] != right.FinalityRecords[i] {
			return fmt.Errorf("aetracore replay finality record mismatch at index %d", i)
		}
	}
	return nil
}
