package types

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestZoneStateMachineInterfaceExecutesExportsImportsAndQueriesProofs(t *testing.T) {
	zone := testZone(ZoneIDContract, ZoneKindContract, VMPolicyAVM, 1)
	runtime, err := NewZoneRuntimeState(zone, hash("contract-state"), nil, DefaultZoneExecutionBudget(), DefaultZoneGasPolicy(), DefaultZoneMessageFilter())
	require.NoError(t, err)
	queues, err := NewZoneMessageQueues(ZoneIDContract, nil, nil)
	require.NoError(t, err)
	machine := &testZoneMachine{zoneID: ZoneIDContract, height: 22, runtime: runtime, queues: queues}

	batch, err := NewZoneBatch(ZoneIDContract, 22, []ZoneTransaction{
		{
			ZoneID:		ZoneIDContract,
			TxHash:		hash("tx-1"),
			MessageType:	"contract.execute",
			GasLimit:	10_000,
			PayloadHash:	hash("payload-1"),
			Sequence:	1,
		},
	}, []ZoneMessage{testZoneMessage(ZoneIDContract, "contract.inbound", 1, 5_000)})
	require.NoError(t, err)
	require.Equal(t, ZoneMempoolLane(ZoneIDContract), batch.MempoolLaneID)

	result, err := ExecuteZoneBatch(context.Background(), machine, batch)
	require.NoError(t, err)
	require.Equal(t, uint64(15_000), result.GasConsumed)
	require.Equal(t, uint32(1), result.TransactionsExecuted)
	require.Equal(t, uint32(1), result.InboundMessagesApplied)
	require.NoError(t, result.ZoneRoot.Validate())

	receipt, err := ApplyInboundMessage(context.Background(), machine, testZoneMessage(ZoneIDContract, "contract.inbound", 2, 1_000))
	require.NoError(t, err)
	require.Equal(t, ZoneReceiptStatusSuccess, receipt.Status)

	root, err := ComputeZoneRoot(context.Background(), machine)
	require.NoError(t, err)
	require.Equal(t, ZoneIDContract, root.ZoneID)

	exported, err := ExportZone(context.Background(), machine)
	require.NoError(t, err)
	require.NoError(t, ImportZone(context.Background(), machine, exported))

	req := ZoneProofRequest{ZoneID: ZoneIDContract, Height: 22, Kind: ZoneProofKindState, Key: "contract/state", Root: root.RootHash, Limit: 4}
	proof, err := QueryZoneProof(context.Background(), machine, req)
	require.NoError(t, err)
	require.Equal(t, req.Root, proof.Root)
}

func TestZoneExecutionMachineLifecycleMatchesZoneInterfaceSpec(t *testing.T) {
	zone := testZone(ZoneIDApplication, ZoneKindApplication, VMPolicyAVM, 1)
	runtime, err := NewZoneRuntimeState(zone, hash("application-state"), nil, DefaultZoneExecutionBudget(), DefaultZoneGasPolicy(), DefaultZoneMessageFilter())
	require.NoError(t, err)
	queues, err := NewZoneMessageQueues(ZoneIDApplication, nil, nil)
	require.NoError(t, err)
	machine := &testZoneMachine{zoneID: ZoneIDApplication, height: 31, runtime: runtime, queues: queues}

	require.NoError(t, BeginZoneBlock(context.Background(), machine))

	tx := ZoneTransaction{
		ZoneID:		ZoneIDApplication,
		TxHash:		hash("application-tx"),
		MessageType:	"application.execute",
		GasLimit:	8_000,
		PayloadHash:	hash("application-payload"),
		Sequence:	7,
	}
	result, err := ExecuteTx(context.Background(), machine, tx)
	require.NoError(t, err)
	require.Equal(t, tx.TxHash, result.ItemHash)
	require.Equal(t, ZoneReceiptStatusSuccess, result.Status)

	msg := testZoneMessage(ZoneIDApplication, "application.inbound", 8, 2_000)
	receipt, err := ApplyMessage(context.Background(), machine, msg)
	require.NoError(t, err)
	require.Equal(t, msg.PayloadHash, receipt.ItemHash)

	summary, err := EndZoneBlock(context.Background(), machine)
	require.NoError(t, err)
	require.Equal(t, uint64(1), summary.TxCount)
	require.Equal(t, uint64(1), summary.InboundMessageCount)
	require.Equal(t, uint64(0), summary.OutboundMessageCount)
	require.Equal(t, uint64(10_000), summary.GasUsed)
	require.Equal(t, uint64(2), summary.StateWrites)
	require.Equal(t, uint64(1), summary.StateReads)
	require.Equal(t, uint32(1), summary.ShardsTouched)
	require.Equal(t, uint64(0), summary.FailedMessages)
	require.Equal(t, summary.StateRoot, summary.ZoneStateRoot)
	require.Equal(t, EmptyRootHash(), summary.EventRoot)
	require.Equal(t, uint32(1), summary.TransactionsExecuted)
	require.Equal(t, uint32(1), summary.InboundMessagesApplied)
	require.Equal(t, uint32(2), summary.ReceiptsProduced)
	require.Equal(t, uint64(10_000), summary.GasConsumed)
	require.Equal(t, ComputeZoneExecutionSummaryHash(summary), summary.SummaryHash)

	root, err := ComputeZoneRoot(context.Background(), machine)
	require.NoError(t, err)
	require.NoError(t, VerifyZoneExecutionSummaryOutputs(summary, root, machine.queues, machine.receipts))
	badSummary := summary
	badSummary.OutboundMessageCount = 9
	badSummary.SummaryHash = ComputeZoneExecutionSummaryHash(badSummary)
	require.ErrorContains(t, VerifyZoneExecutionSummaryOutputs(badSummary, root, machine.queues, machine.receipts), "counts")

	stateRoot, err := StateRoot(context.Background(), machine)
	require.NoError(t, err)
	require.Equal(t, runtime.StateRoot, stateRoot)
}

func TestZoneLocalGasQueuesAndProofValidation(t *testing.T) {
	budget := ZoneExecutionBudget{MaxGas: 100, MaxMessages: 2}
	meter, err := NewZoneGasMeter(ZoneIDFinancial, budget)
	require.NoError(t, err)
	meter, err = meter.Consume(60, 1)
	require.NoError(t, err)
	require.Equal(t, uint64(60), meter.GasUsed)
	_, err = meter.Consume(50, 1)
	require.ErrorContains(t, err, "gas used")

	queues, err := NewZoneMessageQueues(
		ZoneIDFinancial,
		[]ZoneMessage{testZoneMessage(ZoneIDFinancial, "bank.transfer", 2, 10), testZoneMessage(ZoneIDFinancial, "bank.transfer", 1, 10)},
		nil,
	)
	require.NoError(t, err)
	require.Equal(t, uint64(1), queues.Inbox[0].Sequence)
	require.NotEmpty(t, queues.QueueRoot())

	req := ZoneProofRequest{ZoneID: ZoneIDFinancial, Height: 3, Kind: ZoneProofKindInbox, Key: "balance/alice", Root: queues.InboxRoot(), Limit: 1}
	proof, err := NewZoneProof(req, hash("alice-balance"), []string{"leaf"})
	require.NoError(t, err)
	require.NoError(t, proof.ValidateFor(req))
	proof.Path = append(proof.Path, "extra")
	require.ErrorContains(t, proof.ValidateFor(req), "limit")
}

func TestZoneExportManifestRejectsNonReproducibleImport(t *testing.T) {
	zone := testZone(ZoneIDIdentity, ZoneKindIdentity, VMPolicyNativeModule, 1)
	runtime, err := NewZoneRuntimeState(zone, hash("identity-state"), nil, DefaultZoneExecutionBudget(), DefaultZoneGasPolicy(), DefaultZoneMessageFilter())
	require.NoError(t, err)
	queues, err := NewZoneMessageQueues(ZoneIDIdentity, []ZoneMessage{testZoneMessage(ZoneIDIdentity, "identity.lookup", 1, 100)}, nil)
	require.NoError(t, err)
	machine := &testZoneMachine{zoneID: ZoneIDIdentity, height: 44, runtime: runtime, queues: queues}

	exported, err := ExportZone(context.Background(), machine)
	require.NoError(t, err)
	manifest, err := BuildZoneExportManifest(exported, hash("descriptor"), hash("layout"), hash("commitment"), hash("events"))
	require.NoError(t, err)
	exported.Manifest = manifest
	require.NoError(t, exported.Validate())
	require.NoError(t, ValidateZoneImportReproducible(exported, manifest))

	badManifest := manifest
	badManifest.StateRoot = hash("mutated-state")
	badManifest.ExportHash = ComputeZoneExportManifestHash(badManifest)
	require.ErrorContains(t, ValidateZoneImportReproducible(exported, badManifest), "reproduce")
}

type testZoneMachine struct {
	zoneID		ZoneID
	height		uint64
	runtime		ZoneRuntimeState
	queues		ZoneMessageQueues
	receipts	[]ZoneReceipt
	began		bool
	txs		uint32
	messages	uint32
	gas		uint64
}

func (m *testZoneMachine) ZoneID() ZoneID {
	return m.zoneID
}

func (m *testZoneMachine) ExecuteZoneBatch(_ context.Context, batch ZoneBatch) (ZoneBatchResult, error) {
	gas, err := batch.GasRequired()
	if err != nil {
		return ZoneBatchResult{}, err
	}
	meter, err := NewZoneGasMeter(m.zoneID, m.runtime.Budget)
	if err != nil {
		return ZoneBatchResult{}, err
	}
	meter, err = meter.Consume(gas, uint32(len(batch.Transactions)+len(batch.InboundMessages)))
	if err != nil {
		return ZoneBatchResult{}, err
	}
	m.runtime.Budget = meter.Budget()
	for _, msg := range batch.InboundMessages {
		m.queues, err = m.queues.EnqueueInbox(msg)
		if err != nil {
			return ZoneBatchResult{}, err
		}
	}
	receipt, err := NewZoneReceipt(ZoneReceipt{
		ZoneID:		m.zoneID,
		Height:		batch.Height,
		ItemHash:	batch.Transactions[0].TxHash,
		Status:		ZoneReceiptStatusSuccess,
		GasUsed:	gas,
		ResultHash:	hash("zone-result"),
		Sequence:	1,
	})
	if err != nil {
		return ZoneBatchResult{}, err
	}
	m.receipts = append(m.receipts, receipt)
	m.runtime.ReceiptRoot = ComputeZoneReceiptRoot(m.receipts)
	m.runtime.ExecutionResultRoot = ComputeZoneExecutionResultRoot(m.receipts)
	m.runtime.ProofRoot = ComputeZoneRuntimeProofRoot(m.runtime)
	root, err := BuildZoneRoot(batch.Height, m.runtime, m.queues)
	if err != nil {
		return ZoneBatchResult{}, err
	}
	return ZoneBatchResult{
		ZoneID:			m.zoneID,
		Height:			batch.Height,
		TransactionsExecuted:	uint32(len(batch.Transactions)),
		InboundMessagesApplied:	uint32(len(batch.InboundMessages)),
		GasConsumed:		gas,
		Receipts:		[]ZoneReceipt{receipt},
		ZoneRoot:		root,
	}, nil
}

func (m *testZoneMachine) BeginZoneBlock(context.Context) error {
	m.began = true
	m.txs = 0
	m.messages = 0
	m.gas = 0
	return nil
}

func (m *testZoneMachine) ExecuteTx(_ context.Context, tx ZoneTransaction) (ExecutionResult, error) {
	meter, err := NewZoneGasMeter(m.zoneID, m.runtime.Budget)
	if err != nil {
		return ExecutionResult{}, err
	}
	meter, err = meter.Consume(tx.GasLimit, 1)
	if err != nil {
		return ExecutionResult{}, err
	}
	m.runtime.Budget = meter.Budget()
	receipt, err := NewZoneReceipt(ZoneReceipt{
		ZoneID:		m.zoneID,
		Height:		m.height,
		ItemHash:	tx.TxHash,
		Status:		ZoneReceiptStatusSuccess,
		GasUsed:	tx.GasLimit,
		ResultHash:	hash("tx-result"),
		Sequence:	tx.Sequence,
	})
	if err != nil {
		return ExecutionResult{}, err
	}
	m.receipts = append(m.receipts, receipt)
	m.txs++
	m.gas += tx.GasLimit
	m.refreshRuntimeRoots()
	return receipt, nil
}

func (m *testZoneMachine) ApplyMessage(ctx context.Context, msg ZoneMessage) (MessageReceipt, error) {
	receipt, err := m.ApplyInboundMessage(ctx, msg)
	if err != nil {
		return MessageReceipt{}, err
	}
	m.receipts = append(m.receipts, receipt)
	m.messages++
	m.gas += msg.GasLimit
	m.refreshRuntimeRoots()
	return receipt, nil
}

func (m *testZoneMachine) EndZoneBlock(context.Context) (ZoneExecutionSummary, error) {
	return BuildZoneExecutionSummary(m.height, m.runtime, m.queues, m.txs, m.messages, m.receipts, m.gas)
}

func (m *testZoneMachine) StateRoot(context.Context) (string, error) {
	return m.runtime.StateRoot, nil
}

func (m *testZoneMachine) ApplyInboundMessage(_ context.Context, msg ZoneMessage) (ZoneReceipt, error) {
	var err error
	m.queues, err = m.queues.EnqueueInbox(msg)
	if err != nil {
		return ZoneReceipt{}, err
	}
	return NewZoneReceipt(ZoneReceipt{
		ZoneID:		m.zoneID,
		Height:		m.height,
		ItemHash:	msg.PayloadHash,
		Status:		ZoneReceiptStatusSuccess,
		GasUsed:	msg.GasLimit,
		ResultHash:	hash("inbound-result"),
		Sequence:	msg.Sequence,
	})
}

func (m *testZoneMachine) ComputeZoneRoot(context.Context) (ZoneRoot, error) {
	m.runtime.ProofRoot = ComputeZoneRuntimeProofRoot(m.runtime)
	return BuildZoneRoot(m.height, m.runtime, m.queues)
}

func (m *testZoneMachine) ExportZone(context.Context) (ZoneExport, error) {
	root, err := m.ComputeZoneRoot(context.Background())
	if err != nil {
		return ZoneExport{}, err
	}
	req := ZoneProofRequest{ZoneID: m.zoneID, Height: m.height, Kind: ZoneProofKindExport, Key: "export", Root: root.RootHash}
	proof, err := NewZoneProof(req, hash("export-value"), []string{"export"})
	if err != nil {
		return ZoneExport{}, err
	}
	return ZoneExport{ZoneID: m.zoneID, Height: m.height, Runtime: m.runtime, Queues: m.queues, Receipts: m.receipts, Proofs: []ZoneProof{proof}}, nil
}

func (m *testZoneMachine) ImportZone(_ context.Context, exported ZoneExport) error {
	m.runtime = exported.Runtime
	m.queues = exported.Queues.Clone()
	m.receipts = append([]ZoneReceipt(nil), exported.Receipts...)
	return nil
}

func (m *testZoneMachine) QueryZoneProof(_ context.Context, req ZoneProofRequest) (ZoneProof, error) {
	return NewZoneProof(req, hash("proof-value"), []string{"leaf"})
}

func (m *testZoneMachine) refreshRuntimeRoots() {
	m.runtime.ReceiptRoot = ComputeZoneReceiptRoot(m.receipts)
	m.runtime.ExecutionResultRoot = ComputeZoneExecutionResultRoot(m.receipts)
	m.runtime.ProofRoot = ComputeZoneRuntimeProofRoot(m.runtime)
}
