package types

import (
	"fmt"
	"testing"
)

func BenchmarkRegisterZones(b *testing.B) {
	zones := make([]Zone, 0, 128)
	for i := 0; i < 128; i++ {
		id := ZoneID(fmt.Sprintf("ZONE_%04d", i))
		zones = append(zones, testZone(id, ZoneKindApplication, VMPolicyAVM, 1))
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		state := EmptyState()
		var err error
		for _, zone := range zones {
			state, err = RegisterZone(state, zone)
			if err != nil {
				b.Fatal(err)
			}
		}
		if len(state.Zones) != len(zones) {
			b.Fatal("zone count changed")
		}
	}
}

func BenchmarkAppendCommitments(b *testing.B) {
	state, err := RegisterZone(EmptyState(), testZone(ZoneIDFinancial, ZoneKindFinancial, VMPolicyNativeModule, 1))
	if err != nil {
		b.Fatal(err)
	}
	commitments := make([]ZoneCommitment, 0, 256)
	previous := ""
	for i := 1; i <= 256; i++ {
		commitment := benchCommitment(b, ZoneIDFinancial, uint64(i), previous)
		commitments = append(commitments, commitment)
		previous = commitment.CommitmentHash
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		next := state
		for _, commitment := range commitments {
			next, err = AppendCommitment(next, commitment)
			if err != nil {
				b.Fatal(err)
			}
		}
		if len(next.Commitments) != len(commitments) {
			b.Fatal("commitment count changed")
		}
	}
}

func BenchmarkZoneCommitmentValidation(b *testing.B) {
	commitments := make([]ZoneCommitment, 0, 1024)
	previous := ""
	for i := 1; i <= 1024; i++ {
		commitment := benchCommitment(b, ZoneIDFinancial, uint64(i), previous)
		commitments = append(commitments, commitment)
		previous = commitment.CommitmentHash
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		commitment := commitments[i%len(commitments)]
		if err := commitment.ValidateHash(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkExportImportState(b *testing.B) {
	state := EmptyState()
	var err error
	for i := 0; i < 128; i++ {
		id := ZoneID(fmt.Sprintf("ZONE_%04d", i))
		state, err = RegisterZone(state, testZone(id, ZoneKindApplication, VMPolicyAVM, 1))
		if err != nil {
			b.Fatal(err)
		}
	}
	exported := state.Export()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		imported, err := ImportState(exported)
		if err != nil {
			b.Fatal(err)
		}
		if len(imported.Zones) != len(exported.Zones) {
			b.Fatal("zone count changed")
		}
	}
}

func benchCommitment(b *testing.B, id ZoneID, height uint64, previous string) ZoneCommitment {
	b.Helper()
	commitment, err := NewZoneCommitment(
		id,
		height,
		hash(fmt.Sprintf("%s-state-%020d", id, height)),
		hash(fmt.Sprintf("%s-receipt-%020d", id, height)),
		hash(fmt.Sprintf("%s-message-%020d", id, height)),
		hash(fmt.Sprintf("%s-execution-%020d", id, height)),
		previous,
	)
	if err != nil {
		b.Fatal(err)
	}
	return commitment
}
