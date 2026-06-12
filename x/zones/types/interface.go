package types

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
)

const (
	MaxZoneProofKeyLength	= 128
	MaxZoneProofPathItems	= 256
)

type ZoneReceiptStatus string
type ZoneProofKind string

const (
	ZoneReceiptStatusSuccess	ZoneReceiptStatus	= "SUCCESS"
	ZoneReceiptStatusFailed		ZoneReceiptStatus	= "FAILED"
	ZoneReceiptStatusBounced	ZoneReceiptStatus	= "BOUNCED"

	ZoneProofKindState	ZoneProofKind	= "STATE"
	ZoneProofKindInbox	ZoneProofKind	= "INBOX"
	ZoneProofKindOutbox	ZoneProofKind	= "OUTBOX"
	ZoneProofKindReceipt	ZoneProofKind	= "RECEIPT"
	ZoneProofKindRuntime	ZoneProofKind	= "RUNTIME"
	ZoneProofKindExport	ZoneProofKind	= "EXPORT"
)

type ZoneStateMachine interface {
	ZoneID() ZoneID
	ExecuteZoneBatch(context.Context, ZoneBatch) (ZoneBatchResult, error)
	ApplyInboundMessage(context.Context, ZoneMessage) (ZoneReceipt, error)
	ComputeZoneRoot(context.Context) (ZoneRoot, error)
	ExportZone(context.Context) (ZoneExport, error)
	ImportZone(context.Context, ZoneExport) error
	QueryZoneProof(context.Context, ZoneProofRequest) (ZoneProof, error)
}

type ZoneTransaction struct {
	ZoneID		ZoneID
	TxHash		string
	MessageType	string
	GasLimit	uint64
	PayloadHash	string
	Sequence	uint64
}

type ZoneBatch struct {
	ZoneID		ZoneID
	Height		uint64
	MempoolLaneID	string
	Transactions	[]ZoneTransaction
	InboundMessages	[]ZoneMessage
}

type ZoneBatchResult struct {
	ZoneID			ZoneID
	Height			uint64
	TransactionsExecuted	uint32
	InboundMessagesApplied	uint32
	GasConsumed		uint64
	OutboundMessages	[]ZoneMessage
	Receipts		[]ZoneReceipt
	ZoneRoot		ZoneRoot
	ExecutionSummary	ZoneExecutionSummary
}

type ZoneRoot struct {
	ZoneID			ZoneID
	Height			uint64
	ZoneStateRoot		string
	InboxRoot		string
	OutboxRoot		string
	ReceiptRoot		string
	EventRoot		string
	ExecutionResultRoot	string
	ProofRoot		string
	RootHash		string
}

type ZoneExport struct {
	ZoneID		ZoneID
	Height		uint64
	Runtime		ZoneRuntimeState
	Queues		ZoneMessageQueues
	Receipts	[]ZoneReceipt
	Proofs		[]ZoneProof
	Manifest	ZoneExportManifest
}

type ZoneExportManifest struct {
	ZoneID		ZoneID
	Height		uint64
	DescriptorRoot	string
	LayoutRoot	string
	CommitmentRoot	string
	ProofRoot	string
	StateRoot	string
	InboxRoot	string
	OutboxRoot	string
	ReceiptRoot	string
	EventRoot	string
	ExportHash	string
}

type ZoneReceipt struct {
	ZoneID		ZoneID
	Height		uint64
	ItemHash	string
	Status		ZoneReceiptStatus
	GasUsed		uint64
	ResultHash	string
	Sequence	uint64
	ReceiptHash	string
}

type ZoneGasMeter struct {
	ZoneID		ZoneID
	MaxGas		uint64
	GasUsed		uint64
	MaxMessages	uint32
	MessagesUsed	uint32
}

type ZoneMessageQueues struct {
	ZoneID	ZoneID
	Inbox	[]ZoneMessage
	Outbox	[]ZoneMessage
}

type ZoneProofRequest struct {
	ZoneID	ZoneID
	Height	uint64
	Kind	ZoneProofKind
	Key	string
	Root	string
	Limit	uint32
}

type ZoneProof struct {
	ZoneID		ZoneID
	Height		uint64
	Kind		ZoneProofKind
	Key		string
	Root		string
	ValueHash	string
	Path		[]string
	ProofHash	string
}

func ExecuteZoneBatch(ctx context.Context, machine ZoneStateMachine, batch ZoneBatch) (ZoneBatchResult, error) {
	if machine == nil {
		return ZoneBatchResult{}, errors.New("zone state machine is required")
	}
	if machine.ZoneID() != batch.ZoneID {
		return ZoneBatchResult{}, errors.New("zone batch machine route mismatch")
	}
	if err := batch.Validate(); err != nil {
		return ZoneBatchResult{}, err
	}
	result, err := machine.ExecuteZoneBatch(ctx, batch)
	if err != nil {
		return ZoneBatchResult{}, err
	}
	return result, result.Validate()
}

func ApplyInboundMessage(ctx context.Context, machine ZoneStateMachine, msg ZoneMessage) (ZoneReceipt, error) {
	if machine == nil {
		return ZoneReceipt{}, errors.New("zone state machine is required")
	}
	if err := msg.Validate(machine.ZoneID()); err != nil {
		return ZoneReceipt{}, err
	}
	receipt, err := machine.ApplyInboundMessage(ctx, msg)
	if err != nil {
		return ZoneReceipt{}, err
	}
	if receipt.ZoneID != machine.ZoneID() {
		return ZoneReceipt{}, errors.New("zone inbound receipt route mismatch")
	}
	return receipt, receipt.Validate()
}

func ComputeZoneRoot(ctx context.Context, machine ZoneStateMachine) (ZoneRoot, error) {
	if machine == nil {
		return ZoneRoot{}, errors.New("zone state machine is required")
	}
	root, err := machine.ComputeZoneRoot(ctx)
	if err != nil {
		return ZoneRoot{}, err
	}
	if root.ZoneID != machine.ZoneID() {
		return ZoneRoot{}, errors.New("zone root route mismatch")
	}
	return root, root.Validate()
}

func ExportZone(ctx context.Context, machine ZoneStateMachine) (ZoneExport, error) {
	if machine == nil {
		return ZoneExport{}, errors.New("zone state machine is required")
	}
	exported, err := machine.ExportZone(ctx)
	if err != nil {
		return ZoneExport{}, err
	}
	if exported.ZoneID != machine.ZoneID() {
		return ZoneExport{}, errors.New("zone export route mismatch")
	}
	return exported, exported.Validate()
}

func ImportZone(ctx context.Context, machine ZoneStateMachine, exported ZoneExport) error {
	if machine == nil {
		return errors.New("zone state machine is required")
	}
	if exported.ZoneID != machine.ZoneID() {
		return errors.New("zone import route mismatch")
	}
	if err := exported.Validate(); err != nil {
		return err
	}
	return machine.ImportZone(ctx, exported)
}

func QueryZoneProof(ctx context.Context, machine ZoneStateMachine, req ZoneProofRequest) (ZoneProof, error) {
	if machine == nil {
		return ZoneProof{}, errors.New("zone state machine is required")
	}
	if req.ZoneID != machine.ZoneID() {
		return ZoneProof{}, errors.New("zone proof query route mismatch")
	}
	if err := req.Validate(); err != nil {
		return ZoneProof{}, err
	}
	proof, err := machine.QueryZoneProof(ctx, req)
	if err != nil {
		return ZoneProof{}, err
	}
	return proof, proof.ValidateFor(req)
}

func NewZoneBatch(zoneID ZoneID, height uint64, transactions []ZoneTransaction, inbound []ZoneMessage) (ZoneBatch, error) {
	batch := ZoneBatch{
		ZoneID:			zoneID,
		Height:			height,
		MempoolLaneID:		ZoneMempoolLane(zoneID),
		Transactions:		cloneZoneTransactions(transactions),
		InboundMessages:	cloneZoneMessages(inbound),
	}
	return batch, batch.Validate()
}

func (b ZoneBatch) Validate() error {
	if err := ValidateZoneID(b.ZoneID); err != nil {
		return err
	}
	if b.Height == 0 {
		return errors.New("zone batch height must be positive")
	}
	if b.MempoolLaneID != ZoneMempoolLane(b.ZoneID) {
		return fmt.Errorf("zone batch mempool lane must be %q", ZoneMempoolLane(b.ZoneID))
	}
	for i, tx := range b.Transactions {
		if err := tx.Validate(b.ZoneID); err != nil {
			return err
		}
		if i > 0 && compareZoneTransactions(b.Transactions[i-1], tx) >= 0 {
			return errors.New("zone batch transactions must be sorted canonically")
		}
	}
	for i, msg := range b.InboundMessages {
		if err := msg.Validate(b.ZoneID); err != nil {
			return err
		}
		if i > 0 && compareZoneMessages(b.InboundMessages[i-1], msg) >= 0 {
			return errors.New("zone batch inbound messages must be sorted canonically")
		}
	}
	_, err := b.GasRequired()
	return err
}

func (b ZoneBatch) GasRequired() (uint64, error) {
	var total uint64
	for _, tx := range b.Transactions {
		next, err := addZoneGas(total, tx.GasLimit)
		if err != nil {
			return 0, err
		}
		total = next
	}
	for _, msg := range b.InboundMessages {
		next, err := addZoneGas(total, msg.GasLimit)
		if err != nil {
			return 0, err
		}
		total = next
	}
	return total, nil
}

func (t ZoneTransaction) Validate(expectedZone ZoneID) error {
	if err := ValidateZoneID(t.ZoneID); err != nil {
		return err
	}
	if t.ZoneID != expectedZone {
		return fmt.Errorf("zone transaction belongs to %s, expected %s", t.ZoneID, expectedZone)
	}
	if err := ValidateHash("zone transaction hash", t.TxHash); err != nil {
		return err
	}
	if err := validateRuntimeToken("zone transaction message type", t.MessageType, MaxZoneMessageTypeLength); err != nil {
		return err
	}
	if t.GasLimit == 0 {
		return errors.New("zone transaction gas limit must be positive")
	}
	return ValidateHash("zone transaction payload hash", t.PayloadHash)
}

func (r ZoneBatchResult) Validate() error {
	if err := ValidateZoneID(r.ZoneID); err != nil {
		return err
	}
	if r.Height == 0 {
		return errors.New("zone batch result height must be positive")
	}
	if r.ZoneRoot.ZoneID != r.ZoneID || r.ZoneRoot.Height != r.Height {
		return errors.New("zone batch result root mismatch")
	}
	if err := r.ZoneRoot.Validate(); err != nil {
		return err
	}
	for _, msg := range r.OutboundMessages {
		if err := msg.Validate(r.ZoneID); err != nil {
			return err
		}
	}
	for _, receipt := range r.Receipts {
		if err := receipt.Validate(); err != nil {
			return err
		}
		if receipt.ZoneID != r.ZoneID || receipt.Height != r.Height {
			return errors.New("zone batch receipt route mismatch")
		}
	}
	if r.ExecutionSummary.SummaryHash != "" {
		if err := r.ExecutionSummary.Validate(); err != nil {
			return err
		}
		if r.ExecutionSummary.ZoneID != r.ZoneID || r.ExecutionSummary.Height != r.Height {
			return errors.New("zone batch execution summary route mismatch")
		}
		if r.ExecutionSummary.TxCount != uint64(r.TransactionsExecuted) ||
			r.ExecutionSummary.InboundMessageCount != uint64(r.InboundMessagesApplied) ||
			r.ExecutionSummary.OutboundMessageCount != uint64(len(r.OutboundMessages)) ||
			r.ExecutionSummary.GasUsed != r.GasConsumed ||
			r.ExecutionSummary.ZoneStateRoot != r.ZoneRoot.ZoneStateRoot ||
			r.ExecutionSummary.OutboxRoot != r.ZoneRoot.OutboxRoot ||
			r.ExecutionSummary.ReceiptRoot != r.ZoneRoot.ReceiptRoot ||
			r.ExecutionSummary.EventRoot != r.ZoneRoot.EventRoot {
			return errors.New("zone batch execution summary differs from committed outputs")
		}
	}
	return nil
}

func BuildZoneRoot(height uint64, runtime ZoneRuntimeState, queues ZoneMessageQueues) (ZoneRoot, error) {
	if height == 0 {
		return ZoneRoot{}, errors.New("zone root height must be positive")
	}
	if err := runtime.Validate(); err != nil {
		return ZoneRoot{}, err
	}
	if err := queues.Validate(); err != nil {
		return ZoneRoot{}, err
	}
	if runtime.ZoneID != queues.ZoneID {
		return ZoneRoot{}, errors.New("zone root queue route mismatch")
	}
	root := ZoneRoot{
		ZoneID:			runtime.ZoneID,
		Height:			height,
		ZoneStateRoot:		runtime.StateRoot,
		InboxRoot:		ComputeZoneMessageRoot(queues.Inbox),
		OutboxRoot:		ComputeZoneMessageRoot(queues.Outbox),
		ReceiptRoot:		runtime.ReceiptRoot,
		EventRoot:		EmptyRootHash(),
		ExecutionResultRoot:	runtime.ExecutionResultRoot,
		ProofRoot:		runtime.ProofRoot,
	}
	root.RootHash = ComputeZoneRootHash(root)
	return root, root.Validate()
}

func (r ZoneRoot) Validate() error {
	r = canonicalZoneRoot(r)
	if err := ValidateZoneID(r.ZoneID); err != nil {
		return err
	}
	if r.Height == 0 {
		return errors.New("zone root height must be positive")
	}
	for _, item := range []struct {
		name	string
		value	string
	}{
		{name: "zone state root", value: r.ZoneStateRoot},
		{name: "zone inbox root", value: r.InboxRoot},
		{name: "zone outbox root", value: r.OutboxRoot},
		{name: "zone receipt root", value: r.ReceiptRoot},
		{name: "zone event root", value: r.EventRoot},
		{name: "zone execution result root", value: r.ExecutionResultRoot},
		{name: "zone proof root", value: r.ProofRoot},
		{name: "zone root hash", value: r.RootHash},
	} {
		if err := ValidateHash(item.name, item.value); err != nil {
			return err
		}
	}
	if r.RootHash != ComputeZoneRootHash(r) {
		return errors.New("zone root hash mismatch")
	}
	return nil
}

func (e ZoneExport) Validate() error {
	if err := ValidateZoneID(e.ZoneID); err != nil {
		return err
	}
	if e.Height == 0 {
		return errors.New("zone export height must be positive")
	}
	if err := e.Runtime.Validate(); err != nil {
		return err
	}
	if e.Runtime.ZoneID != e.ZoneID {
		return errors.New("zone export runtime route mismatch")
	}
	if err := e.Queues.Validate(); err != nil {
		return err
	}
	if e.Queues.ZoneID != e.ZoneID {
		return errors.New("zone export queue route mismatch")
	}
	for _, receipt := range e.Receipts {
		if err := receipt.Validate(); err != nil {
			return err
		}
		if receipt.ZoneID != e.ZoneID {
			return errors.New("zone export receipt route mismatch")
		}
	}
	for _, proof := range e.Proofs {
		if err := proof.Validate(); err != nil {
			return err
		}
		if proof.ZoneID != e.ZoneID {
			return errors.New("zone export proof route mismatch")
		}
	}
	if e.Manifest.ExportHash != "" {
		if err := e.Manifest.Validate(); err != nil {
			return err
		}
		if e.Manifest.ZoneID != e.ZoneID || e.Manifest.Height != e.Height {
			return errors.New("zone export manifest route mismatch")
		}
		if e.Manifest.StateRoot != e.Runtime.StateRoot ||
			e.Manifest.InboxRoot != e.Queues.InboxRoot() ||
			e.Manifest.OutboxRoot != e.Queues.OutboxRoot() ||
			e.Manifest.ReceiptRoot != ComputeZoneReceiptRoot(e.Receipts) ||
			e.Manifest.ProofRoot != ComputeZoneProofCollectionRoot(e.Proofs) {
			return errors.New("zone export manifest does not reproduce exported roots")
		}
	}
	return nil
}

func NewZoneExportManifest(manifest ZoneExportManifest) (ZoneExportManifest, error) {
	if manifest.ExportHash != "" {
		return ZoneExportManifest{}, errors.New("zone export manifest hash must be empty before construction")
	}
	if manifest.EventRoot == "" {
		manifest.EventRoot = EmptyRootHash()
	}
	if err := manifest.ValidateFormat(); err != nil {
		return ZoneExportManifest{}, err
	}
	manifest.ExportHash = ComputeZoneExportManifestHash(manifest)
	return manifest, manifest.Validate()
}

func BuildZoneExportManifest(exported ZoneExport, descriptorRoot string, layoutRoot string, commitmentRoot string, eventRoot string) (ZoneExportManifest, error) {
	if eventRoot == "" {
		eventRoot = EmptyRootHash()
	}
	return NewZoneExportManifest(ZoneExportManifest{
		ZoneID:		exported.ZoneID,
		Height:		exported.Height,
		DescriptorRoot:	descriptorRoot,
		LayoutRoot:	layoutRoot,
		CommitmentRoot:	commitmentRoot,
		ProofRoot:	ComputeZoneProofCollectionRoot(exported.Proofs),
		StateRoot:	exported.Runtime.StateRoot,
		InboxRoot:	exported.Queues.InboxRoot(),
		OutboxRoot:	exported.Queues.OutboxRoot(),
		ReceiptRoot:	ComputeZoneReceiptRoot(exported.Receipts),
		EventRoot:	eventRoot,
	})
}

func ValidateZoneImportReproducible(exported ZoneExport, expected ZoneExportManifest) error {
	if err := exported.Validate(); err != nil {
		return err
	}
	if err := expected.Validate(); err != nil {
		return err
	}
	actual, err := BuildZoneExportManifest(exported, expected.DescriptorRoot, expected.LayoutRoot, expected.CommitmentRoot, expected.EventRoot)
	if err != nil {
		return err
	}
	if actual.ExportHash != expected.ExportHash {
		return errors.New("zone import does not reproduce descriptor, layout, commitment, and proof roots")
	}
	return nil
}

func (m ZoneExportManifest) ValidateFormat() error {
	if err := ValidateZoneID(m.ZoneID); err != nil {
		return err
	}
	if m.Height == 0 {
		return errors.New("zone export manifest height must be positive")
	}
	for _, item := range []struct {
		name	string
		value	string
	}{
		{name: "zone export descriptor root", value: m.DescriptorRoot},
		{name: "zone export layout root", value: m.LayoutRoot},
		{name: "zone export commitment root", value: m.CommitmentRoot},
		{name: "zone export proof root", value: m.ProofRoot},
		{name: "zone export state root", value: m.StateRoot},
		{name: "zone export inbox root", value: m.InboxRoot},
		{name: "zone export outbox root", value: m.OutboxRoot},
		{name: "zone export receipt root", value: m.ReceiptRoot},
		{name: "zone export event root", value: m.EventRoot},
	} {
		if err := ValidateHash(item.name, item.value); err != nil {
			return err
		}
	}
	if m.ExportHash != "" {
		return ValidateHash("zone export manifest hash", m.ExportHash)
	}
	return nil
}

func (m ZoneExportManifest) Validate() error {
	if err := m.ValidateFormat(); err != nil {
		return err
	}
	if m.ExportHash == "" {
		return errors.New("zone export manifest hash is required")
	}
	if m.ExportHash != ComputeZoneExportManifestHash(m) {
		return errors.New("zone export manifest hash mismatch")
	}
	return nil
}

func NewZoneReceipt(receipt ZoneReceipt) (ZoneReceipt, error) {
	if receipt.ReceiptHash != "" {
		return ZoneReceipt{}, errors.New("zone receipt hash must be empty before construction")
	}
	if err := receipt.ValidateFormat(); err != nil {
		return ZoneReceipt{}, err
	}
	receipt.ReceiptHash = ComputeZoneReceiptHash(receipt)
	return receipt, receipt.Validate()
}

func (r ZoneReceipt) ValidateFormat() error {
	if err := ValidateZoneID(r.ZoneID); err != nil {
		return err
	}
	if r.Height == 0 {
		return errors.New("zone receipt height must be positive")
	}
	if err := ValidateHash("zone receipt item hash", r.ItemHash); err != nil {
		return err
	}
	if !IsZoneReceiptStatus(r.Status) {
		return fmt.Errorf("unknown zone receipt status %q", r.Status)
	}
	if err := ValidateHash("zone receipt result hash", r.ResultHash); err != nil {
		return err
	}
	if r.ReceiptHash != "" {
		return ValidateHash("zone receipt hash", r.ReceiptHash)
	}
	return nil
}

func (r ZoneReceipt) Validate() error {
	if err := r.ValidateFormat(); err != nil {
		return err
	}
	if r.ReceiptHash == "" {
		return errors.New("zone receipt hash is required")
	}
	if r.ReceiptHash != ComputeZoneReceiptHash(r) {
		return errors.New("zone receipt hash mismatch")
	}
	return nil
}

func NewZoneGasMeter(zoneID ZoneID, budget ZoneExecutionBudget) (ZoneGasMeter, error) {
	if err := ValidateZoneID(zoneID); err != nil {
		return ZoneGasMeter{}, err
	}
	if err := budget.Validate(); err != nil {
		return ZoneGasMeter{}, err
	}
	meter := ZoneGasMeter{
		ZoneID:		zoneID,
		MaxGas:		budget.MaxGas,
		GasUsed:	budget.GasUsed,
		MaxMessages:	budget.MaxMessages,
		MessagesUsed:	budget.MessagesUsed,
	}
	return meter, meter.Validate()
}

func (m ZoneGasMeter) Validate() error {
	if err := ValidateZoneID(m.ZoneID); err != nil {
		return err
	}
	return m.Budget().Validate()
}

func (m ZoneGasMeter) Consume(gas uint64, messages uint32) (ZoneGasMeter, error) {
	budget, err := m.Budget().Consume(gas, messages)
	if err != nil {
		return ZoneGasMeter{}, err
	}
	next := m
	next.GasUsed = budget.GasUsed
	next.MessagesUsed = budget.MessagesUsed
	return next, next.Validate()
}

func (m ZoneGasMeter) Budget() ZoneExecutionBudget {
	return ZoneExecutionBudget{
		MaxGas:		m.MaxGas,
		GasUsed:	m.GasUsed,
		MaxMessages:	m.MaxMessages,
		MessagesUsed:	m.MessagesUsed,
	}
}

func NewZoneMessageQueues(zoneID ZoneID, inbox []ZoneMessage, outbox []ZoneMessage) (ZoneMessageQueues, error) {
	queues := ZoneMessageQueues{
		ZoneID:	zoneID,
		Inbox:	cloneZoneMessages(inbox),
		Outbox:	cloneZoneMessages(outbox),
	}
	return queues, queues.Validate()
}

func (q ZoneMessageQueues) Validate() error {
	if err := ValidateZoneID(q.ZoneID); err != nil {
		return err
	}
	if err := validateZoneMessageList("zone inbox", q.ZoneID, q.Inbox); err != nil {
		return err
	}
	return validateZoneMessageList("zone outbox", q.ZoneID, q.Outbox)
}

func (q ZoneMessageQueues) EnqueueInbox(msg ZoneMessage) (ZoneMessageQueues, error) {
	if err := q.Validate(); err != nil {
		return ZoneMessageQueues{}, err
	}
	if err := msg.Validate(q.ZoneID); err != nil {
		return ZoneMessageQueues{}, err
	}
	next := q.Clone()
	next.Inbox = append(next.Inbox, msg)
	sortZoneMessages(next.Inbox)
	return next, next.Validate()
}

func (q ZoneMessageQueues) EnqueueOutbox(msg ZoneMessage) (ZoneMessageQueues, error) {
	if err := q.Validate(); err != nil {
		return ZoneMessageQueues{}, err
	}
	if err := msg.Validate(q.ZoneID); err != nil {
		return ZoneMessageQueues{}, err
	}
	next := q.Clone()
	next.Outbox = append(next.Outbox, msg)
	sortZoneMessages(next.Outbox)
	return next, next.Validate()
}

func (q ZoneMessageQueues) Clone() ZoneMessageQueues {
	return ZoneMessageQueues{
		ZoneID:	q.ZoneID,
		Inbox:	cloneZoneMessages(q.Inbox),
		Outbox:	cloneZoneMessages(q.Outbox),
	}
}

func (q ZoneMessageQueues) InboxRoot() string {
	return ComputeZoneMessageRoot(q.Inbox)
}

func (q ZoneMessageQueues) OutboxRoot() string {
	return ComputeZoneMessageRoot(q.Outbox)
}

func (q ZoneMessageQueues) QueueRoot() string {
	return ComputeZoneMessageQueuesRoot(q.Inbox, q.Outbox)
}

func (r ZoneProofRequest) Validate() error {
	if err := ValidateZoneID(r.ZoneID); err != nil {
		return err
	}
	if r.Height == 0 {
		return errors.New("zone proof request height must be positive")
	}
	if !IsZoneProofKind(r.Kind) {
		return fmt.Errorf("unknown zone proof kind %q", r.Kind)
	}
	if err := validateRuntimeToken("zone proof key", r.Key, MaxZoneProofKeyLength); err != nil {
		return err
	}
	if err := ValidateHash("zone proof root", r.Root); err != nil {
		return err
	}
	if r.Limit > MaxZoneProofPathItems {
		return fmt.Errorf("zone proof path limit must be <= %d", MaxZoneProofPathItems)
	}
	return nil
}

func NewZoneProof(req ZoneProofRequest, valueHash string, path []string) (ZoneProof, error) {
	if err := req.Validate(); err != nil {
		return ZoneProof{}, err
	}
	proof := ZoneProof{
		ZoneID:		req.ZoneID,
		Height:		req.Height,
		Kind:		req.Kind,
		Key:		req.Key,
		Root:		req.Root,
		ValueHash:	valueHash,
		Path:		append([]string(nil), path...),
	}
	if err := proof.ValidateFormat(); err != nil {
		return ZoneProof{}, err
	}
	proof.ProofHash = ComputeZoneProofHash(proof)
	return proof, proof.ValidateFor(req)
}

func (p ZoneProof) ValidateFormat() error {
	if err := ValidateZoneID(p.ZoneID); err != nil {
		return err
	}
	if p.Height == 0 {
		return errors.New("zone proof height must be positive")
	}
	if !IsZoneProofKind(p.Kind) {
		return fmt.Errorf("unknown zone proof kind %q", p.Kind)
	}
	if err := validateRuntimeToken("zone proof key", p.Key, MaxZoneProofKeyLength); err != nil {
		return err
	}
	if err := ValidateHash("zone proof root", p.Root); err != nil {
		return err
	}
	if err := ValidateHash("zone proof value hash", p.ValueHash); err != nil {
		return err
	}
	if len(p.Path) > MaxZoneProofPathItems {
		return fmt.Errorf("zone proof path must be <= %d items", MaxZoneProofPathItems)
	}
	for _, item := range p.Path {
		if err := validateRuntimeToken("zone proof path item", item, MaxZoneProofKeyLength); err != nil {
			return err
		}
	}
	if p.ProofHash != "" {
		return ValidateHash("zone proof hash", p.ProofHash)
	}
	return nil
}

func (p ZoneProof) Validate() error {
	if err := p.ValidateFormat(); err != nil {
		return err
	}
	if p.ProofHash == "" {
		return errors.New("zone proof hash is required")
	}
	if p.ProofHash != ComputeZoneProofHash(p) {
		return errors.New("zone proof hash mismatch")
	}
	return nil
}

func (p ZoneProof) ValidateFor(req ZoneProofRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}
	if err := p.ValidateFormat(); err != nil {
		return err
	}
	if p.ZoneID != req.ZoneID || p.Height != req.Height || p.Kind != req.Kind || p.Key != req.Key || p.Root != req.Root {
		return errors.New("zone proof response does not match request")
	}
	if req.Limit > 0 && uint32(len(p.Path)) > req.Limit {
		return errors.New("zone proof response exceeds request path limit")
	}
	if p.ProofHash == "" {
		return errors.New("zone proof hash is required")
	}
	if p.ProofHash != ComputeZoneProofHash(p) {
		return errors.New("zone proof hash mismatch")
	}
	return nil
}

func ZoneMempoolLane(zoneID ZoneID) string {
	return "zone:" + string(zoneID)
}

func ComputeZoneRootHash(root ZoneRoot) string {
	root = canonicalZoneRoot(root)
	return hashRuntimeParts(
		"aetra-zone-root-v1",
		string(root.ZoneID),
		fmt.Sprint(root.Height),
		root.ZoneStateRoot,
		root.InboxRoot,
		root.OutboxRoot,
		root.ReceiptRoot,
		root.EventRoot,
		root.ExecutionResultRoot,
		root.ProofRoot,
	)
}

func ComputeZoneExportManifestHash(manifest ZoneExportManifest) string {
	if manifest.EventRoot == "" {
		manifest.EventRoot = EmptyRootHash()
	}
	return hashRuntimeParts(
		"aetra-zone-export-manifest-v1",
		string(manifest.ZoneID),
		fmt.Sprint(manifest.Height),
		manifest.DescriptorRoot,
		manifest.LayoutRoot,
		manifest.CommitmentRoot,
		manifest.ProofRoot,
		manifest.StateRoot,
		manifest.InboxRoot,
		manifest.OutboxRoot,
		manifest.ReceiptRoot,
		manifest.EventRoot,
	)
}

func ComputeZoneReceiptHash(receipt ZoneReceipt) string {
	return hashRuntimeParts(
		"aetra-zone-receipt-v1",
		string(receipt.ZoneID),
		fmt.Sprint(receipt.Height),
		receipt.ItemHash,
		string(receipt.Status),
		fmt.Sprint(receipt.GasUsed),
		receipt.ResultHash,
		fmt.Sprint(receipt.Sequence),
	)
}

func ComputeZoneReceiptRoot(receipts []ZoneReceipt) string {
	ordered := cloneZoneReceipts(receipts)
	h := sha256.New()
	writeRuntimePart(h, "aetra-zone-receipt-root-v1")
	writeRuntimeUint64(h, uint64(len(ordered)))
	for _, receipt := range ordered {
		writeRuntimePart(h, receipt.ReceiptHash)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeZoneExecutionResultRoot(receipts []ZoneReceipt) string {
	ordered := cloneZoneReceipts(receipts)
	h := sha256.New()
	writeRuntimePart(h, "aetra-zone-execution-result-root-v1")
	writeRuntimeUint64(h, uint64(len(ordered)))
	for _, receipt := range ordered {
		writeRuntimePart(h, receipt.ItemHash)
		writeRuntimePart(h, string(receipt.Status))
		writeRuntimeUint64(h, receipt.GasUsed)
		writeRuntimePart(h, receipt.ResultHash)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeZoneMessageQueuesRoot(inbox []ZoneMessage, outbox []ZoneMessage) string {
	return hashRuntimeParts(
		"aetra-zone-message-queues-v1",
		ComputeZoneMessageRoot(inbox),
		ComputeZoneMessageRoot(outbox),
	)
}

func ComputeZoneProofCollectionRoot(proofs []ZoneProof) string {
	ordered := append([]ZoneProof(nil), proofs...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Kind != ordered[j].Kind {
			return ordered[i].Kind < ordered[j].Kind
		}
		if ordered[i].Key != ordered[j].Key {
			return ordered[i].Key < ordered[j].Key
		}
		return ordered[i].ProofHash < ordered[j].ProofHash
	})
	parts := []string{"aetra-zone-proof-collection-root-v1", fmt.Sprint(len(ordered))}
	for _, proof := range ordered {
		parts = append(parts, proof.ProofHash)
	}
	return hashRuntimeParts(parts...)
}

func ComputeZoneProofHash(proof ZoneProof) string {
	h := sha256.New()
	writeRuntimePart(h, "aetra-zone-proof-v1")
	writeRuntimePart(h, string(proof.ZoneID))
	writeRuntimeUint64(h, proof.Height)
	writeRuntimePart(h, string(proof.Kind))
	writeRuntimePart(h, proof.Key)
	writeRuntimePart(h, proof.Root)
	writeRuntimePart(h, proof.ValueHash)
	writeRuntimeUint64(h, uint64(len(proof.Path)))
	for _, item := range proof.Path {
		writeRuntimePart(h, item)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func IsZoneReceiptStatus(status ZoneReceiptStatus) bool {
	switch status {
	case ZoneReceiptStatusSuccess, ZoneReceiptStatusFailed, ZoneReceiptStatusBounced:
		return true
	default:
		return false
	}
}

func IsZoneProofKind(kind ZoneProofKind) bool {
	switch kind {
	case ZoneProofKindState, ZoneProofKindInbox, ZoneProofKindOutbox, ZoneProofKindReceipt, ZoneProofKindRuntime, ZoneProofKindExport:
		return true
	default:
		return false
	}
}

func canonicalZoneRoot(root ZoneRoot) ZoneRoot {
	if root.EventRoot == "" {
		root.EventRoot = EmptyRootHash()
	}
	return root
}

func validateZoneMessageList(fieldName string, zoneID ZoneID, messages []ZoneMessage) error {
	for i, msg := range messages {
		if err := msg.Validate(zoneID); err != nil {
			return fmt.Errorf("%s: %w", fieldName, err)
		}
		if i > 0 && compareZoneMessages(messages[i-1], msg) >= 0 {
			return fmt.Errorf("%s must be sorted canonically", fieldName)
		}
	}
	return nil
}

func cloneZoneTransactions(transactions []ZoneTransaction) []ZoneTransaction {
	out := append([]ZoneTransaction(nil), transactions...)
	sort.SliceStable(out, func(i, j int) bool {
		return compareZoneTransactions(out[i], out[j]) < 0
	})
	return out
}

func compareZoneTransactions(left, right ZoneTransaction) int {
	if left.Sequence < right.Sequence {
		return -1
	}
	if left.Sequence > right.Sequence {
		return 1
	}
	if left.TxHash < right.TxHash {
		return -1
	}
	if left.TxHash > right.TxHash {
		return 1
	}
	if left.MessageType < right.MessageType {
		return -1
	}
	if left.MessageType > right.MessageType {
		return 1
	}
	return 0
}

func cloneZoneReceipts(receipts []ZoneReceipt) []ZoneReceipt {
	out := append([]ZoneReceipt(nil), receipts...)
	sort.SliceStable(out, func(i, j int) bool {
		return compareZoneReceipts(out[i], out[j]) < 0
	})
	return out
}

func compareZoneReceipts(left, right ZoneReceipt) int {
	if left.Sequence < right.Sequence {
		return -1
	}
	if left.Sequence > right.Sequence {
		return 1
	}
	if left.ItemHash < right.ItemHash {
		return -1
	}
	if left.ItemHash > right.ItemHash {
		return 1
	}
	return 0
}

func addZoneGas(left uint64, right uint64) (uint64, error) {
	if right > ^uint64(0)-left {
		return 0, errors.New("zone gas overflow")
	}
	return left + right, nil
}

func hashRuntimeParts(parts ...string) string {
	h := sha256.New()
	for _, part := range parts {
		writeRuntimePart(h, part)
	}
	return hex.EncodeToString(h.Sum(nil))
}
