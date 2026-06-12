package types

import "testing"

func BenchmarkApplyMessage(b *testing.B) {
	state, msg := benchMeshFixture(b)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		next, receipt, err := ApplyMessage(state, msg, successResult(), 100)
		if err != nil {
			b.Fatal(err)
		}
		if receipt.Status != ReceiptStatusSuccess || len(next.Receipts) != 1 {
			b.Fatal("unexpected receipt state")
		}
	}
}

func BenchmarkMeshProofVerification(b *testing.B) {
	state, msg := benchMeshFixture(b)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := ValidateSourceProof(state, msg, 100); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSortMessages(b *testing.B) {
	state, msg := benchMeshFixture(b)
	commitment := state.FinalizedCommitments[0]
	messages := make([]MeshMessage, 0, 512)
	for i := 0; i < 512; i++ {
		next := msg
		next.Nonce = uint64(512 - i)
		next.Sequence = uint64(i % 32)
		next.MessageID = ""
		next.Proof = MeshProof{}
		next = benchMessage(b, next)
		next.Proof = BuildProof(next, commitment)
		messages = append(messages, next)
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ordered := SortMessages(messages)
		if len(ordered) != len(messages) {
			b.Fatal("message count changed")
		}
	}
}

func BenchmarkExportImportState(b *testing.B) {
	state, msg := benchMeshFixture(b)
	next, _, err := ApplyMessage(state, msg, successResult(), 100)
	if err != nil {
		b.Fatal(err)
	}
	exported := next.Export()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		imported, err := ImportState(exported)
		if err != nil {
			b.Fatal(err)
		}
		if len(imported.ReplayMarkers) != len(exported.ReplayMarkers) {
			b.Fatal("replay marker count changed")
		}
	}
}

func benchMeshFixture(b *testing.B) (MeshState, MeshMessage) {
	b.Helper()

	state := EmptyState(DefaultParams())
	var err error
	state, err = RegisterDestination(state, MeshDestination{
		ZoneID:		ZoneID("FINANCIAL_ZONE"),
		ShardID:	ShardID("0:0"),
		Active:		true,
	})
	if err != nil {
		b.Fatal(err)
	}
	state, err = RegisterDestination(state, MeshDestination{
		ZoneID:		ZoneID("CONTRACT_ZONE"),
		ShardID:	ShardID("0:1"),
		Active:		true,
	})
	if err != nil {
		b.Fatal(err)
	}

	commitment := FinalizedCommitment{
		ZoneID:		ZoneID("FINANCIAL_ZONE"),
		ShardID:	ShardID("0:0"),
		Height:		90,
		CommitmentHash:	HashParts("source-commitment", "financial", "0:0", "90"),
		MessageRoot:	HashParts("message-root", "financial", "90"),
		ReceiptRoot:	HashParts("receipt-root", "financial", "90"),
	}
	state, err = AddFinalizedCommitment(state, commitment)
	if err != nil {
		b.Fatal(err)
	}

	msg := benchMessage(b, MeshMessage{
		SourceZone:		ZoneID("FINANCIAL_ZONE"),
		SourceShard:		ShardID("0:0"),
		DestinationZone:	ZoneID("CONTRACT_ZONE"),
		DestinationShard:	ShardID("0:1"),
		Nonce:			7,
		Sender:			[]byte("orb1sender"),
		Recipient:		[]byte("contract1recipient"),
		AssetCommitment:	HashParts("asset", "100naet"),
		PayloadHash:		HashParts("payload", "execute"),
		TimeoutHeight:		150,
		Finality:		FinalityReference{Height: commitment.Height, CommitmentHash: commitment.CommitmentHash},
		Sequence:		3,
		SourceLogicalTime:	88,
	})
	msg.Proof = BuildProof(msg, commitment)
	return state, msg
}

func benchMessage(b *testing.B, msg MeshMessage) MeshMessage {
	b.Helper()
	out, err := NewMessage(msg)
	if err != nil {
		b.Fatal(err)
	}
	return out
}
