package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

const (
	AVMCrossZoneBounceDisabled	AVMCrossZoneBounceBehavior	= "disabled"
	AVMCrossZoneBounceAllowed	AVMCrossZoneBounceBehavior	= "allowed"
	AVMCrossZoneBounceRequired	AVMCrossZoneBounceBehavior	= "required"

	AVMCrossZoneProofNone		AVMCrossZoneProofRequirement	= "none"
	AVMCrossZoneProofAuth		AVMCrossZoneProofRequirement	= "auth"
	AVMCrossZoneProofState		AVMCrossZoneProofRequirement	= "state"
	AVMCrossZoneProofAuthAndState	AVMCrossZoneProofRequirement	= "auth_and_state"

	AVMCrossZoneValueNone		AVMCrossZoneValueAccounting	= "none"
	AVMCrossZoneValueEscrow		AVMCrossZoneValueAccounting	= "escrow"
	AVMCrossZoneValueMessage	AVMCrossZoneValueAccounting	= "message_value"
	AVMCrossZoneValueEscrowPlus	AVMCrossZoneValueAccounting	= "escrow_plus_message_value"

	AVMCrossZoneFailureNone		AVMCrossZoneFailureResolution	= "none"
	AVMCrossZoneFailureBounced	AVMCrossZoneFailureResolution	= "bounced"
	AVMCrossZoneFailureDeadLettered	AVMCrossZoneFailureResolution	= "dead_lettered"

	MaxAVMCrossZoneOpcodeLength	= 96
	MaxAVMCrossZoneRouteKey		= 160
)

type AVMCrossZoneBounceBehavior string
type AVMCrossZoneProofRequirement string
type AVMCrossZoneValueAccounting string
type AVMCrossZoneFailureResolution string

type AVMCrossZoneEscrowStatus string

const (
	AVMCrossZoneEscrowLocked	AVMCrossZoneEscrowStatus	= "locked"
	AVMCrossZoneEscrowReleased	AVMCrossZoneEscrowStatus	= "released"
	AVMCrossZoneEscrowRefunded	AVMCrossZoneEscrowStatus	= "refunded"
)

type AVMCrossZoneRoutePolicy struct {
	SourceZone		zonestypes.ZoneID
	DestinationZone		zonestypes.ZoneID
	GasPolicy		zonestypes.ZoneGasPolicy
	ExecutionBudget		zonestypes.ZoneExecutionBudget
	MessageFilter		zonestypes.ZoneMessageFilter
	AllowedOpcodes		[]string
	BounceBehavior		AVMCrossZoneBounceBehavior
	ProofRequirement	AVMCrossZoneProofRequirement
	ValueAccounting		AVMCrossZoneValueAccounting
	PolicyHash		string
}

type AVMZoneRouterRoute struct {
	RouteKey		string
	SourceZone		zonestypes.ZoneID
	DestinationZone		zonestypes.ZoneID
	PolicyHash		string
	OutputMessageRoot	string
	DestinationInboxRoot	string
	CrossZoneReceiptRoot	string
	ValueEscrowRoot		string
	BounceBehavior		AVMCrossZoneBounceBehavior
	ProofRequirement	AVMCrossZoneProofRequirement
	ValueAccounting		AVMCrossZoneValueAccounting
	RouteHash		string
}

type AVMZoneRouterTable struct {
	Height		uint64
	Routes		[]AVMZoneRouterRoute
	TableRoot	string
}

type AVMCrossZoneZoneRoots struct {
	ZoneID			zonestypes.ZoneID
	Height			uint64
	OutputMessageRoot	string
	DestinationInboxRoot	string
	CrossZoneReceiptRoot	string
	ValueEscrowRoot		string
	CrossZoneRootHash	string
}

type AVMAetraCoreZoneCommitmentSet struct {
	Height		uint64
	ZoneRoots	[]AVMZoneStateRoot
	CoreRoot	string
}

type AVMCrossZoneValueEscrowRecord struct {
	MessageID	string
	SourceZone	zonestypes.ZoneID
	DestinationZone	zonestypes.ZoneID
	AmountNAET	uint64
	RefundedNAET	uint64
	ReleasedNAET	uint64
	Status		AVMCrossZoneEscrowStatus
	RefundReceiptID	string
	EscrowHash	string
}

type AVMCrossZoneProofKind string

const (
	AVMCrossZoneProofRoute		AVMCrossZoneProofKind	= "route"
	AVMCrossZoneProofExecution	AVMCrossZoneProofKind	= "execution"
	AVMCrossZoneProofEscrow		AVMCrossZoneProofKind	= "escrow"
)

type AVMCrossZoneProof struct {
	Kind			AVMCrossZoneProofKind
	ZoneID			zonestypes.ZoneID
	MessageID		string
	RouteKey		string
	RouterTableRoot		string
	ZoneCrossRoot		string
	OutputMessageRoot	string
	DestinationInboxRoot	string
	CrossZoneReceiptRoot	string
	ValueEscrowRoot		string
	ReceiptID		string
	EscrowHash		string
	ProofHash		string
}

type AVMCrossZoneProofIndex struct {
	RouterTable	AVMZoneRouterTable
	ZoneRoots	[]AVMCrossZoneZoneRoots
	Executions	[]AVMCrossZoneExecution
	Escrows		[]AVMCrossZoneValueEscrowRecord
}

type AVMCrossZoneExecution struct {
	Message				AVMAsyncMessage
	RoutePolicy			AVMCrossZoneRoutePolicy
	DestinationQueueEntry		AVMZoneQueueEntry
	DestinationReceipt		AVMExecutionReceipt
	SourceOutputMessagesRoot	string
	DestinationReceiptRoot		string
	DirectStateWriteAttempted	bool
	ValueEscrowedNAET		uint64
	ValueAccountedNAET		uint64
	FailureResolution		AVMCrossZoneFailureResolution
	BounceMessageOptional		AVMAsyncMessage
	DeadLetterOptional		AVMDeadLetterRecord
	ExecutionHash			string
}

func NewAVMZoneRouterTable(table AVMZoneRouterTable) (AVMZoneRouterTable, error) {
	table = canonicalAVMZoneRouterTable(table)
	for i := range table.Routes {
		if table.Routes[i].RouteHash == "" {
			table.Routes[i].RouteHash = ComputeAVMZoneRouterRouteHash(table.Routes[i])
		}
	}
	table.TableRoot = ComputeAVMZoneRouterTableRoot(table)
	return table, table.Validate()
}

func NewAVMZoneRouterRoute(policy AVMCrossZoneRoutePolicy, roots AVMCrossZoneZoneRoots) (AVMZoneRouterRoute, error) {
	policy = canonicalAVMCrossZoneRoutePolicy(policy)
	if err := policy.Validate(); err != nil {
		return AVMZoneRouterRoute{}, err
	}
	roots = canonicalAVMCrossZoneZoneRoots(roots)
	if err := roots.Validate(); err != nil {
		return AVMZoneRouterRoute{}, err
	}
	if roots.ZoneID != policy.DestinationZone {
		return AVMZoneRouterRoute{}, errors.New("AVM zone router route roots must belong to destination zone")
	}
	route := AVMZoneRouterRoute{
		RouteKey:		AVMZoneRouterRouteKey(policy.SourceZone, policy.DestinationZone, policy.AllowedOpcodes),
		SourceZone:		policy.SourceZone,
		DestinationZone:	policy.DestinationZone,
		PolicyHash:		policy.PolicyHash,
		OutputMessageRoot:	roots.OutputMessageRoot,
		DestinationInboxRoot:	roots.DestinationInboxRoot,
		CrossZoneReceiptRoot:	roots.CrossZoneReceiptRoot,
		ValueEscrowRoot:	roots.ValueEscrowRoot,
		BounceBehavior:		policy.BounceBehavior,
		ProofRequirement:	policy.ProofRequirement,
		ValueAccounting:	policy.ValueAccounting,
	}
	route.RouteHash = ComputeAVMZoneRouterRouteHash(route)
	return route, route.Validate()
}

func NewAVMCrossZoneZoneRoots(roots AVMCrossZoneZoneRoots) (AVMCrossZoneZoneRoots, error) {
	roots = canonicalAVMCrossZoneZoneRoots(roots)
	roots.CrossZoneRootHash = ComputeAVMCrossZoneZoneRootHash(roots)
	return roots, roots.Validate()
}

func NewAVMAetraCoreZoneCommitmentSet(set AVMAetraCoreZoneCommitmentSet) (AVMAetraCoreZoneCommitmentSet, error) {
	set = canonicalAVMAetraCoreZoneCommitmentSet(set)
	set.CoreRoot = ComputeAVMAetraCoreZoneCommitmentRoot(set)
	return set, set.Validate()
}

func NewAVMCrossZoneValueEscrowRecord(record AVMCrossZoneValueEscrowRecord) (AVMCrossZoneValueEscrowRecord, error) {
	record = canonicalAVMCrossZoneValueEscrowRecord(record)
	if record.Status == "" {
		record.Status = AVMCrossZoneEscrowLocked
	}
	record.EscrowHash = ComputeAVMCrossZoneValueEscrowHash(record)
	return record, record.Validate()
}

func RefundAVMCrossZoneEscrow(record AVMCrossZoneValueEscrowRecord, receipt AVMExecutionReceipt) (AVMCrossZoneValueEscrowRecord, error) {
	record = canonicalAVMCrossZoneValueEscrowRecord(record)
	if err := record.Validate(); err != nil {
		return AVMCrossZoneValueEscrowRecord{}, err
	}
	receipt = canonicalAVMExecutionReceipt(receipt)
	if err := receipt.Validate(); err != nil {
		return AVMCrossZoneValueEscrowRecord{}, err
	}
	if receipt.MessageID != record.MessageID || receipt.ZoneID != record.DestinationZone {
		return AVMCrossZoneValueEscrowRecord{}, errors.New("AVM cross-zone refund receipt mismatch")
	}
	if !isCrossZoneFailureStatus(receipt.Status) {
		return AVMCrossZoneValueEscrowRecord{}, errors.New("AVM cross-zone escrow refund requires failed terminal receipt")
	}
	if record.Status != AVMCrossZoneEscrowLocked {
		return AVMCrossZoneValueEscrowRecord{}, errors.New("AVM cross-zone escrow can only refund locked value")
	}
	record.Status = AVMCrossZoneEscrowRefunded
	record.RefundedNAET = record.AmountNAET
	record.RefundReceiptID = receipt.ReceiptID
	record.EscrowHash = ComputeAVMCrossZoneValueEscrowHash(record)
	return record, record.Validate()
}

func ReleaseAVMCrossZoneEscrow(record AVMCrossZoneValueEscrowRecord, receipt AVMExecutionReceipt) (AVMCrossZoneValueEscrowRecord, error) {
	record = canonicalAVMCrossZoneValueEscrowRecord(record)
	if err := record.Validate(); err != nil {
		return AVMCrossZoneValueEscrowRecord{}, err
	}
	receipt = canonicalAVMExecutionReceipt(receipt)
	if err := receipt.Validate(); err != nil {
		return AVMCrossZoneValueEscrowRecord{}, err
	}
	if receipt.MessageID != record.MessageID || receipt.ZoneID != record.DestinationZone {
		return AVMCrossZoneValueEscrowRecord{}, errors.New("AVM cross-zone release receipt mismatch")
	}
	if receipt.Status != AVMReceiptStatusExecuted {
		return AVMCrossZoneValueEscrowRecord{}, errors.New("AVM cross-zone escrow release requires executed receipt")
	}
	if record.Status != AVMCrossZoneEscrowLocked {
		return AVMCrossZoneValueEscrowRecord{}, errors.New("AVM cross-zone escrow can only release locked value")
	}
	record.Status = AVMCrossZoneEscrowReleased
	record.ReleasedNAET = record.AmountNAET
	record.EscrowHash = ComputeAVMCrossZoneValueEscrowHash(record)
	return record, record.Validate()
}

func QueryAVMCrossZoneProof(index AVMCrossZoneProofIndex, kind AVMCrossZoneProofKind, zoneID zonestypes.ZoneID, messageID string) (AVMCrossZoneProof, error) {
	index = canonicalAVMCrossZoneProofIndex(index)
	if err := index.RouterTable.Validate(); err != nil {
		return AVMCrossZoneProof{}, err
	}
	if !IsAVMCrossZoneProofKind(kind) {
		return AVMCrossZoneProof{}, fmt.Errorf("invalid AVM cross-zone proof kind %q", kind)
	}
	if err := zonestypes.ValidateZoneID(zoneID); err != nil {
		return AVMCrossZoneProof{}, err
	}
	if err := zonestypes.ValidateHash("AVM cross-zone proof message id", messageID); err != nil {
		return AVMCrossZoneProof{}, err
	}
	zoneRoots, found := findCrossZoneRoots(index.ZoneRoots, zoneID)
	if !found {
		return AVMCrossZoneProof{}, errors.New("AVM cross-zone proof zone roots not found")
	}
	execution, found := findCrossZoneExecution(index.Executions, messageID)
	if !found {
		return AVMCrossZoneProof{}, errors.New("AVM cross-zone proof execution not found")
	}
	routeKey := AVMZoneRouterRouteKey(execution.Message.SourceZone, execution.Message.DestinationZone, execution.RoutePolicy.AllowedOpcodes)
	if _, found := findZoneRouterRoute(index.RouterTable.Routes, routeKey); !found {
		return AVMCrossZoneProof{}, errors.New("AVM cross-zone proof route not found")
	}
	proof := AVMCrossZoneProof{
		Kind:			kind,
		ZoneID:			zoneID,
		MessageID:		messageID,
		RouteKey:		routeKey,
		RouterTableRoot:	index.RouterTable.TableRoot,
		ZoneCrossRoot:		zoneRoots.CrossZoneRootHash,
		OutputMessageRoot:	zoneRoots.OutputMessageRoot,
		DestinationInboxRoot:	zoneRoots.DestinationInboxRoot,
		CrossZoneReceiptRoot:	zoneRoots.CrossZoneReceiptRoot,
		ValueEscrowRoot:	zoneRoots.ValueEscrowRoot,
		ReceiptID:		execution.DestinationReceipt.ReceiptID,
	}
	if kind == AVMCrossZoneProofEscrow {
		escrow, found := findCrossZoneEscrow(index.Escrows, messageID)
		if !found {
			return AVMCrossZoneProof{}, errors.New("AVM cross-zone proof escrow not found")
		}
		proof.EscrowHash = escrow.EscrowHash
	}
	proof.ProofHash = ComputeAVMCrossZoneProofHash(proof)
	return proof, proof.Validate()
}

func NewAVMCrossZoneRoutePolicy(policy AVMCrossZoneRoutePolicy) (AVMCrossZoneRoutePolicy, error) {
	policy = canonicalAVMCrossZoneRoutePolicy(policy)
	if policy.GasPolicy.Denom == "" {
		policy.GasPolicy = zonestypes.DefaultZoneGasPolicy()
	}
	if policy.ExecutionBudget.MaxGas == 0 && policy.ExecutionBudget.MaxMessages == 0 {
		policy.ExecutionBudget = zonestypes.DefaultZoneExecutionBudget()
	}
	if len(policy.MessageFilter.AllowedMessageTypes) == 0 {
		policy.MessageFilter = zonestypes.DefaultZoneMessageFilter()
	}
	if policy.BounceBehavior == "" {
		policy.BounceBehavior = AVMCrossZoneBounceAllowed
	}
	if policy.ProofRequirement == "" {
		policy.ProofRequirement = AVMCrossZoneProofNone
	}
	if policy.ValueAccounting == "" {
		policy.ValueAccounting = AVMCrossZoneValueNone
	}
	policy.PolicyHash = ComputeAVMCrossZoneRoutePolicyHash(policy)
	return policy, policy.Validate()
}

func AdmitAVMCrossZoneMessage(queue AVMZoneQueue, msg AVMAsyncMessage, height uint64, maxDepth uint32, policy AVMCrossZoneRoutePolicy) (AVMZoneQueue, AVMZoneQueueEntry, error) {
	msg = canonicalAVMAsyncMessage(msg)
	if err := msg.Validate(); err != nil {
		return AVMZoneQueue{}, AVMZoneQueueEntry{}, err
	}
	policy = canonicalAVMCrossZoneRoutePolicy(policy)
	if err := policy.Validate(); err != nil {
		return AVMZoneQueue{}, AVMZoneQueueEntry{}, err
	}
	if err := ValidateAVMCrossZoneRoute(msg, policy); err != nil {
		return AVMZoneQueue{}, AVMZoneQueueEntry{}, err
	}
	if queue.ZoneID != msg.DestinationZone {
		return AVMZoneQueue{}, AVMZoneQueueEntry{}, errors.New("AVM cross-zone destination queue mismatch")
	}
	return AdmitAVMZoneQueueMessage(queue, msg, height, maxDepth)
}

func NewAVMCrossZoneExecution(execution AVMCrossZoneExecution) (AVMCrossZoneExecution, error) {
	execution = canonicalAVMCrossZoneExecution(execution)
	execution.ExecutionHash = ComputeAVMCrossZoneExecutionHash(execution)
	return execution, execution.Validate()
}

func ValidateAVMCrossZoneRoute(msg AVMAsyncMessage, policy AVMCrossZoneRoutePolicy) error {
	msg = canonicalAVMAsyncMessage(msg)
	if err := msg.Validate(); err != nil {
		return err
	}
	policy = canonicalAVMCrossZoneRoutePolicy(policy)
	if err := policy.Validate(); err != nil {
		return err
	}
	if msg.SourceZone == msg.DestinationZone {
		return errors.New("AVM cross-zone route requires distinct source and destination zones")
	}
	if msg.SourceZone != policy.SourceZone {
		return errors.New("AVM cross-zone route source zone mismatch")
	}
	if msg.DestinationZone != policy.DestinationZone {
		return errors.New("AVM cross-zone route destination zone mismatch")
	}
	if !policy.MessageFilter.Allows(msg.PayloadType) {
		return fmt.Errorf("AVM cross-zone payload type %q is not allowed by destination filter", msg.PayloadType)
	}
	if !crossZoneOpcodeAllowed(policy.AllowedOpcodes, msg.PayloadType) {
		return fmt.Errorf("AVM cross-zone opcode %q is not allowed", msg.PayloadType)
	}
	if err := validateCrossZoneProofs(msg, policy.ProofRequirement); err != nil {
		return err
	}
	if msg.ValueNAET > 0 && policy.ValueAccounting == AVMCrossZoneValueNone {
		return errors.New("AVM cross-zone value transfer requires escrow or message value accounting")
	}
	return nil
}

func (p AVMCrossZoneRoutePolicy) Validate() error {
	p = canonicalAVMCrossZoneRoutePolicy(p)
	if err := zonestypes.ValidateZoneID(p.SourceZone); err != nil {
		return err
	}
	if err := zonestypes.ValidateZoneID(p.DestinationZone); err != nil {
		return err
	}
	if p.SourceZone == p.DestinationZone {
		return errors.New("AVM cross-zone policy requires distinct source and destination zones")
	}
	if err := p.GasPolicy.Validate(); err != nil {
		return err
	}
	if err := p.ExecutionBudget.Validate(); err != nil {
		return err
	}
	if err := p.MessageFilter.Validate(); err != nil {
		return err
	}
	if !IsAVMCrossZoneBounceBehavior(p.BounceBehavior) {
		return fmt.Errorf("invalid AVM cross-zone bounce behavior %q", p.BounceBehavior)
	}
	if !IsAVMCrossZoneProofRequirement(p.ProofRequirement) {
		return fmt.Errorf("invalid AVM cross-zone proof requirement %q", p.ProofRequirement)
	}
	if !IsAVMCrossZoneValueAccounting(p.ValueAccounting) {
		return fmt.Errorf("invalid AVM cross-zone value accounting %q", p.ValueAccounting)
	}
	if err := validateCrossZoneOpcodes(p.AllowedOpcodes); err != nil {
		return err
	}
	if p.PolicyHash == "" {
		return errors.New("AVM cross-zone route policy hash is required")
	}
	if err := zonestypes.ValidateHash("AVM cross-zone route policy hash", p.PolicyHash); err != nil {
		return err
	}
	if p.PolicyHash != ComputeAVMCrossZoneRoutePolicyHash(p) {
		return errors.New("AVM cross-zone route policy hash mismatch")
	}
	return nil
}

func (r AVMZoneRouterRoute) Validate() error {
	r = canonicalAVMZoneRouterRoute(r)
	if err := validateRouterOptionalToken("AVM zone router route key", r.RouteKey, MaxAVMCrossZoneRouteKey); err != nil {
		return err
	}
	if r.RouteKey == "" {
		return errors.New("AVM zone router route key is required")
	}
	if err := zonestypes.ValidateZoneID(r.SourceZone); err != nil {
		return err
	}
	if err := zonestypes.ValidateZoneID(r.DestinationZone); err != nil {
		return err
	}
	if r.SourceZone == r.DestinationZone {
		return errors.New("AVM zone router route requires distinct zones")
	}
	for _, item := range []struct {
		name	string
		value	string
	}{
		{name: "AVM zone router policy hash", value: r.PolicyHash},
		{name: "AVM zone router output message root", value: r.OutputMessageRoot},
		{name: "AVM zone router destination inbox root", value: r.DestinationInboxRoot},
		{name: "AVM zone router cross-zone receipt root", value: r.CrossZoneReceiptRoot},
		{name: "AVM zone router value escrow root", value: r.ValueEscrowRoot},
		{name: "AVM zone router route hash", value: r.RouteHash},
	} {
		if err := zonestypes.ValidateHash(item.name, item.value); err != nil {
			return err
		}
	}
	if !IsAVMCrossZoneBounceBehavior(r.BounceBehavior) {
		return fmt.Errorf("invalid AVM zone router bounce behavior %q", r.BounceBehavior)
	}
	if !IsAVMCrossZoneProofRequirement(r.ProofRequirement) {
		return fmt.Errorf("invalid AVM zone router proof requirement %q", r.ProofRequirement)
	}
	if !IsAVMCrossZoneValueAccounting(r.ValueAccounting) {
		return fmt.Errorf("invalid AVM zone router value accounting %q", r.ValueAccounting)
	}
	if r.RouteHash != ComputeAVMZoneRouterRouteHash(r) {
		return errors.New("AVM zone router route hash mismatch")
	}
	return nil
}

func (t AVMZoneRouterTable) Validate() error {
	t = canonicalAVMZoneRouterTable(t)
	if t.Height == 0 {
		return errors.New("AVM zone router table height must be positive")
	}
	if len(t.Routes) == 0 {
		return errors.New("AVM zone router table must contain routes")
	}
	seen := make(map[string]struct{}, len(t.Routes))
	for i, route := range t.Routes {
		if err := route.Validate(); err != nil {
			return err
		}
		if _, found := seen[route.RouteKey]; found {
			return fmt.Errorf("duplicate AVM zone router route %q", route.RouteKey)
		}
		seen[route.RouteKey] = struct{}{}
		if i > 0 && t.Routes[i-1].RouteKey >= route.RouteKey {
			return errors.New("AVM zone router table routes must be sorted canonically")
		}
	}
	if t.TableRoot == "" {
		return errors.New("AVM zone router table root is required")
	}
	if err := zonestypes.ValidateHash("AVM zone router table root", t.TableRoot); err != nil {
		return err
	}
	if t.TableRoot != ComputeAVMZoneRouterTableRoot(t) {
		return errors.New("AVM zone router table root mismatch")
	}
	return nil
}

func (r AVMCrossZoneZoneRoots) Validate() error {
	r = canonicalAVMCrossZoneZoneRoots(r)
	if err := zonestypes.ValidateZoneID(r.ZoneID); err != nil {
		return err
	}
	if r.Height == 0 {
		return errors.New("AVM cross-zone roots height must be positive")
	}
	for _, item := range []struct {
		name	string
		value	string
	}{
		{name: "AVM cross-zone output message root", value: r.OutputMessageRoot},
		{name: "AVM cross-zone destination inbox root", value: r.DestinationInboxRoot},
		{name: "AVM cross-zone receipt root", value: r.CrossZoneReceiptRoot},
		{name: "AVM cross-zone value escrow root", value: r.ValueEscrowRoot},
		{name: "AVM cross-zone root hash", value: r.CrossZoneRootHash},
	} {
		if err := zonestypes.ValidateHash(item.name, item.value); err != nil {
			return err
		}
	}
	if r.CrossZoneRootHash != ComputeAVMCrossZoneZoneRootHash(r) {
		return errors.New("AVM cross-zone roots hash mismatch")
	}
	return nil
}

func (s AVMAetraCoreZoneCommitmentSet) Validate() error {
	s = canonicalAVMAetraCoreZoneCommitmentSet(s)
	if s.Height == 0 {
		return errors.New("Aether core zone commitment height must be positive")
	}
	if len(s.ZoneRoots) == 0 {
		return errors.New("Aether core zone commitment set must contain zone roots")
	}
	seen := make(map[zonestypes.ZoneID]struct{}, len(s.ZoneRoots))
	for i, root := range s.ZoneRoots {
		if err := root.Validate(); err != nil {
			return err
		}
		if root.Height != s.Height {
			return errors.New("Aether core zone commitment height drift")
		}
		if _, found := seen[root.ZoneID]; found {
			return fmt.Errorf("duplicate Aether core zone root %q", root.ZoneID)
		}
		seen[root.ZoneID] = struct{}{}
		if i > 0 && s.ZoneRoots[i-1].ZoneID >= root.ZoneID {
			return errors.New("Aether core zone roots must be sorted canonically")
		}
	}
	if s.CoreRoot == "" {
		return errors.New("Aether core zone commitment root is required")
	}
	if err := zonestypes.ValidateHash("Aether core zone commitment root", s.CoreRoot); err != nil {
		return err
	}
	if s.CoreRoot != ComputeAVMAetraCoreZoneCommitmentRoot(s) {
		return errors.New("Aether core zone commitment root mismatch")
	}
	return nil
}

func (r AVMCrossZoneValueEscrowRecord) Validate() error {
	r = canonicalAVMCrossZoneValueEscrowRecord(r)
	if err := zonestypes.ValidateHash("AVM cross-zone escrow message id", r.MessageID); err != nil {
		return err
	}
	if err := zonestypes.ValidateZoneID(r.SourceZone); err != nil {
		return err
	}
	if err := zonestypes.ValidateZoneID(r.DestinationZone); err != nil {
		return err
	}
	if r.SourceZone == r.DestinationZone {
		return errors.New("AVM cross-zone escrow requires distinct zones")
	}
	if r.AmountNAET == 0 {
		return errors.New("AVM cross-zone escrow amount must be positive")
	}
	if !IsAVMCrossZoneEscrowStatus(r.Status) {
		return fmt.Errorf("invalid AVM cross-zone escrow status %q", r.Status)
	}
	if r.RefundedNAET > r.AmountNAET || r.ReleasedNAET > r.AmountNAET {
		return errors.New("AVM cross-zone escrow accounting exceeds amount")
	}
	if r.RefundedNAET > 0 && r.ReleasedNAET > 0 {
		return errors.New("AVM cross-zone escrow cannot both refund and release value")
	}
	switch r.Status {
	case AVMCrossZoneEscrowLocked:
		if r.RefundedNAET != 0 || r.ReleasedNAET != 0 || r.RefundReceiptID != "" {
			return errors.New("locked AVM cross-zone escrow must not refund or release value")
		}
	case AVMCrossZoneEscrowRefunded:
		if r.RefundedNAET != r.AmountNAET {
			return errors.New("refunded AVM cross-zone escrow must refund full amount")
		}
		if err := zonestypes.ValidateHash("AVM cross-zone escrow refund receipt id", r.RefundReceiptID); err != nil {
			return err
		}
	case AVMCrossZoneEscrowReleased:
		if r.ReleasedNAET != r.AmountNAET {
			return errors.New("released AVM cross-zone escrow must release full amount")
		}
		if r.RefundReceiptID != "" {
			return errors.New("released AVM cross-zone escrow must not have refund receipt")
		}
	}
	if r.EscrowHash == "" {
		return errors.New("AVM cross-zone escrow hash is required")
	}
	if err := zonestypes.ValidateHash("AVM cross-zone escrow hash", r.EscrowHash); err != nil {
		return err
	}
	if r.EscrowHash != ComputeAVMCrossZoneValueEscrowHash(r) {
		return errors.New("AVM cross-zone escrow hash mismatch")
	}
	return nil
}

func (p AVMCrossZoneProof) Validate() error {
	p = canonicalAVMCrossZoneProof(p)
	if !IsAVMCrossZoneProofKind(p.Kind) {
		return fmt.Errorf("invalid AVM cross-zone proof kind %q", p.Kind)
	}
	if err := zonestypes.ValidateZoneID(p.ZoneID); err != nil {
		return err
	}
	if err := validateRouterOptionalToken("AVM cross-zone proof route key", p.RouteKey, MaxAVMCrossZoneRouteKey); err != nil {
		return err
	}
	if p.RouteKey == "" {
		return errors.New("AVM cross-zone proof route key is required")
	}
	for _, item := range []struct {
		name	string
		value	string
	}{
		{name: "AVM cross-zone proof message id", value: p.MessageID},
		{name: "AVM cross-zone proof router table root", value: p.RouterTableRoot},
		{name: "AVM cross-zone proof zone root", value: p.ZoneCrossRoot},
		{name: "AVM cross-zone proof output message root", value: p.OutputMessageRoot},
		{name: "AVM cross-zone proof destination inbox root", value: p.DestinationInboxRoot},
		{name: "AVM cross-zone proof receipt root", value: p.CrossZoneReceiptRoot},
		{name: "AVM cross-zone proof value escrow root", value: p.ValueEscrowRoot},
		{name: "AVM cross-zone proof receipt id", value: p.ReceiptID},
		{name: "AVM cross-zone proof hash", value: p.ProofHash},
	} {
		if err := zonestypes.ValidateHash(item.name, item.value); err != nil {
			return err
		}
	}
	if p.Kind == AVMCrossZoneProofEscrow {
		if err := zonestypes.ValidateHash("AVM cross-zone proof escrow hash", p.EscrowHash); err != nil {
			return err
		}
	}
	if p.ProofHash != ComputeAVMCrossZoneProofHash(p) {
		return errors.New("AVM cross-zone proof hash mismatch")
	}
	return nil
}

func (e AVMCrossZoneExecution) Validate() error {
	e = canonicalAVMCrossZoneExecution(e)
	if err := ValidateAVMCrossZoneRoute(e.Message, e.RoutePolicy); err != nil {
		return err
	}
	if err := e.DestinationQueueEntry.Validate(); err != nil {
		return err
	}
	if e.DestinationQueueEntry.ZoneID != e.Message.DestinationZone ||
		e.DestinationQueueEntry.MessageID != e.Message.ID {
		return errors.New("AVM cross-zone destination queue entry mismatch")
	}
	if err := e.DestinationReceipt.Validate(); err != nil {
		return err
	}
	if e.DestinationReceipt.MessageID != e.Message.ID ||
		e.DestinationReceipt.ZoneID != e.Message.DestinationZone {
		return errors.New("AVM cross-zone destination receipt mismatch")
	}
	if e.DirectStateWriteAttempted {
		return errors.New("AVM cross-zone execution forbids direct state writes")
	}
	for _, item := range []struct {
		name	string
		value	string
	}{
		{name: "AVM cross-zone source output messages root", value: e.SourceOutputMessagesRoot},
		{name: "AVM cross-zone destination receipt root", value: e.DestinationReceiptRoot},
	} {
		if err := zonestypes.ValidateHash(item.name, item.value); err != nil {
			return err
		}
	}
	if err := e.validateValueAccounting(); err != nil {
		return err
	}
	if err := e.validateFailureResolution(); err != nil {
		return err
	}
	if e.ExecutionHash == "" {
		return errors.New("AVM cross-zone execution hash is required")
	}
	if err := zonestypes.ValidateHash("AVM cross-zone execution hash", e.ExecutionHash); err != nil {
		return err
	}
	if e.ExecutionHash != ComputeAVMCrossZoneExecutionHash(e) {
		return errors.New("AVM cross-zone execution hash mismatch")
	}
	return nil
}

func (e AVMCrossZoneExecution) validateValueAccounting() error {
	if e.Message.ValueNAET == 0 {
		return nil
	}
	switch e.RoutePolicy.ValueAccounting {
	case AVMCrossZoneValueEscrow:
		if e.ValueEscrowedNAET != e.Message.ValueNAET {
			return errors.New("AVM cross-zone escrow value accounting mismatch")
		}
	case AVMCrossZoneValueMessage:
		if e.ValueAccountedNAET != e.Message.ValueNAET {
			return errors.New("AVM cross-zone message value accounting mismatch")
		}
	case AVMCrossZoneValueEscrowPlus:
		if e.ValueEscrowedNAET+e.ValueAccountedNAET < e.Message.ValueNAET || e.ValueEscrowedNAET+e.ValueAccountedNAET < e.ValueEscrowedNAET {
			return errors.New("AVM cross-zone combined value accounting mismatch")
		}
	default:
		return errors.New("AVM cross-zone value transfer requires escrow or message value accounting")
	}
	return nil
}

func (e AVMCrossZoneExecution) validateFailureResolution() error {
	status := e.DestinationReceipt.Status
	if !isCrossZoneFailureStatus(status) {
		if e.FailureResolution != AVMCrossZoneFailureNone {
			return errors.New("AVM cross-zone successful execution must not carry failure resolution")
		}
		return nil
	}
	switch e.FailureResolution {
	case AVMCrossZoneFailureBounced:
		if e.RoutePolicy.BounceBehavior == AVMCrossZoneBounceDisabled {
			return errors.New("AVM cross-zone bounce is disabled")
		}
		return validateCrossZoneBounceMessage(e.Message, e.BounceMessageOptional)
	case AVMCrossZoneFailureDeadLettered:
		if e.RoutePolicy.BounceBehavior == AVMCrossZoneBounceRequired && e.Message.BounceFlag {
			return errors.New("AVM cross-zone failure must bounce when bounce is required")
		}
		if e.DeadLetterOptional.MessageID != e.Message.ID {
			return errors.New("AVM cross-zone dead letter message mismatch")
		}
		if e.DeadLetterOptional.ZoneID != e.Message.DestinationZone {
			return errors.New("AVM cross-zone dead letter zone mismatch")
		}
		if e.DestinationReceipt.Status == AVMReceiptStatusDeadLettered {
			return e.DeadLetterOptional.ValidateWithReceipt(e.DestinationReceipt)
		}
		return e.DeadLetterOptional.Validate()
	default:
		return errors.New("failed AVM cross-zone execution must bounce or dead-letter")
	}
}

func IsAVMCrossZoneBounceBehavior(value AVMCrossZoneBounceBehavior) bool {
	switch value {
	case AVMCrossZoneBounceDisabled, AVMCrossZoneBounceAllowed, AVMCrossZoneBounceRequired:
		return true
	default:
		return false
	}
}

func IsAVMCrossZoneProofRequirement(value AVMCrossZoneProofRequirement) bool {
	switch value {
	case AVMCrossZoneProofNone, AVMCrossZoneProofAuth, AVMCrossZoneProofState, AVMCrossZoneProofAuthAndState:
		return true
	default:
		return false
	}
}

func IsAVMCrossZoneValueAccounting(value AVMCrossZoneValueAccounting) bool {
	switch value {
	case AVMCrossZoneValueNone, AVMCrossZoneValueEscrow, AVMCrossZoneValueMessage, AVMCrossZoneValueEscrowPlus:
		return true
	default:
		return false
	}
}

func IsAVMCrossZoneEscrowStatus(value AVMCrossZoneEscrowStatus) bool {
	switch value {
	case AVMCrossZoneEscrowLocked, AVMCrossZoneEscrowReleased, AVMCrossZoneEscrowRefunded:
		return true
	default:
		return false
	}
}

func IsAVMCrossZoneProofKind(value AVMCrossZoneProofKind) bool {
	switch value {
	case AVMCrossZoneProofRoute, AVMCrossZoneProofExecution, AVMCrossZoneProofEscrow:
		return true
	default:
		return false
	}
}

func AVMZoneRouterRouteKey(sourceZone, destinationZone zonestypes.ZoneID, opcodes []string) string {
	canonical := append([]string(nil), opcodes...)
	for i := range canonical {
		canonical[i] = strings.TrimSpace(canonical[i])
	}
	sort.Strings(canonical)
	return fmt.Sprintf("%s/%s/%s", sourceZone, destinationZone, strings.Join(canonical, "."))
}

func ComputeAVMZoneOutputMessageRoot(zoneID zonestypes.ZoneID, messages []AVMAsyncMessage) string {
	out := cloneCrossZoneMessages(messages)
	sort.SliceStable(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-cross-zone-output-messages-v1")
	writeEnginePart(h, string(zoneID))
	writeEngineUint64(h, uint64(len(out)))
	for _, msg := range out {
		writeEnginePart(h, msg.ID)
		writeEnginePart(h, string(msg.SourceZone))
		writeEnginePart(h, string(msg.DestinationZone))
		writeEngineUint64(h, msg.ValueNAET)
		writeEnginePart(h, msg.PayloadHash)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMDestinationInboxRoot(zoneID zonestypes.ZoneID, entries []AVMZoneQueueEntry) string {
	out := append([]AVMZoneQueueEntry(nil), entries...)
	for i := range out {
		out[i] = canonicalAVMZoneQueueEntry(out[i])
	}
	sort.SliceStable(out, func(i, j int) bool { return compareAVMQueueEntries(out[i], out[j]) < 0 })
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-cross-zone-destination-inbox-v1")
	writeEnginePart(h, string(zoneID))
	writeEngineUint64(h, uint64(len(out)))
	for _, entry := range out {
		writeEnginePart(h, entry.MessageID)
		writeEnginePart(h, string(entry.Lane))
		writeEnginePart(h, entry.SortKey)
		writeEngineUint64(h, entry.ScheduledHeight)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMCrossZoneReceiptRoot(zoneID zonestypes.ZoneID, receipts []AVMExecutionReceipt) string {
	out := append([]AVMExecutionReceipt(nil), receipts...)
	for i := range out {
		out[i] = canonicalAVMExecutionReceipt(out[i])
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ReceiptID < out[j].ReceiptID })
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-cross-zone-receipts-v1")
	writeEnginePart(h, string(zoneID))
	writeEngineUint64(h, uint64(len(out)))
	for _, receipt := range out {
		writeEnginePart(h, receipt.ReceiptID)
		writeEnginePart(h, receipt.MessageID)
		writeEnginePart(h, string(receipt.Status))
		writeEngineUint64(h, receipt.GasUsed)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMCrossZoneEscrowRoot(zoneID zonestypes.ZoneID, records []AVMCrossZoneValueEscrowRecord) string {
	out := append([]AVMCrossZoneValueEscrowRecord(nil), records...)
	for i := range out {
		out[i] = canonicalAVMCrossZoneValueEscrowRecord(out[i])
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].MessageID < out[j].MessageID })
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-cross-zone-escrow-root-v1")
	writeEnginePart(h, string(zoneID))
	writeEngineUint64(h, uint64(len(out)))
	for _, record := range out {
		writeEnginePart(h, record.MessageID)
		writeEnginePart(h, string(record.Status))
		writeEngineUint64(h, record.AmountNAET)
		writeEngineUint64(h, record.RefundedNAET)
		writeEngineUint64(h, record.ReleasedNAET)
		writeEnginePart(h, record.EscrowHash)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMZoneRouterRouteHash(route AVMZoneRouterRoute) string {
	route = canonicalAVMZoneRouterRoute(route)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-zone-router-route-v1")
	writeEnginePart(h, route.RouteKey)
	writeEnginePart(h, string(route.SourceZone))
	writeEnginePart(h, string(route.DestinationZone))
	writeEnginePart(h, route.PolicyHash)
	writeEnginePart(h, route.OutputMessageRoot)
	writeEnginePart(h, route.DestinationInboxRoot)
	writeEnginePart(h, route.CrossZoneReceiptRoot)
	writeEnginePart(h, route.ValueEscrowRoot)
	writeEnginePart(h, string(route.BounceBehavior))
	writeEnginePart(h, string(route.ProofRequirement))
	writeEnginePart(h, string(route.ValueAccounting))
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMZoneRouterTableRoot(table AVMZoneRouterTable) string {
	table = canonicalAVMZoneRouterTable(table)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-zone-router-table-v1")
	writeEngineUint64(h, table.Height)
	writeEngineUint64(h, uint64(len(table.Routes)))
	for _, route := range table.Routes {
		writeEnginePart(h, route.RouteHash)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMCrossZoneZoneRootHash(roots AVMCrossZoneZoneRoots) string {
	roots = canonicalAVMCrossZoneZoneRoots(roots)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-cross-zone-roots-v1")
	writeEnginePart(h, string(roots.ZoneID))
	writeEngineUint64(h, roots.Height)
	writeEnginePart(h, roots.OutputMessageRoot)
	writeEnginePart(h, roots.DestinationInboxRoot)
	writeEnginePart(h, roots.CrossZoneReceiptRoot)
	writeEnginePart(h, roots.ValueEscrowRoot)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMAetraCoreZoneCommitmentRoot(set AVMAetraCoreZoneCommitmentSet) string {
	set = canonicalAVMAetraCoreZoneCommitmentSet(set)
	h := sha256.New()
	writeEnginePart(h, "aetra-core-zone-commitments-v1")
	writeEngineUint64(h, set.Height)
	writeEngineUint64(h, uint64(len(set.ZoneRoots)))
	for _, root := range set.ZoneRoots {
		writeEnginePart(h, string(root.ZoneID))
		writeEnginePart(h, root.RootHash)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMCrossZoneValueEscrowHash(record AVMCrossZoneValueEscrowRecord) string {
	record = canonicalAVMCrossZoneValueEscrowRecord(record)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-cross-zone-value-escrow-v1")
	writeEnginePart(h, record.MessageID)
	writeEnginePart(h, string(record.SourceZone))
	writeEnginePart(h, string(record.DestinationZone))
	writeEngineUint64(h, record.AmountNAET)
	writeEngineUint64(h, record.RefundedNAET)
	writeEngineUint64(h, record.ReleasedNAET)
	writeEnginePart(h, string(record.Status))
	writeEnginePart(h, record.RefundReceiptID)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMCrossZoneProofHash(proof AVMCrossZoneProof) string {
	proof = canonicalAVMCrossZoneProof(proof)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-cross-zone-proof-v1")
	writeEnginePart(h, string(proof.Kind))
	writeEnginePart(h, string(proof.ZoneID))
	writeEnginePart(h, proof.MessageID)
	writeEnginePart(h, proof.RouteKey)
	writeEnginePart(h, proof.RouterTableRoot)
	writeEnginePart(h, proof.ZoneCrossRoot)
	writeEnginePart(h, proof.OutputMessageRoot)
	writeEnginePart(h, proof.DestinationInboxRoot)
	writeEnginePart(h, proof.CrossZoneReceiptRoot)
	writeEnginePart(h, proof.ValueEscrowRoot)
	writeEnginePart(h, proof.ReceiptID)
	writeEnginePart(h, proof.EscrowHash)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMCrossZoneRoutePolicyHash(policy AVMCrossZoneRoutePolicy) string {
	policy = canonicalAVMCrossZoneRoutePolicy(policy)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-cross-zone-route-policy-v1")
	writeEnginePart(h, string(policy.SourceZone))
	writeEnginePart(h, string(policy.DestinationZone))
	writeEnginePart(h, policy.GasPolicy.Denom)
	writeEngineUint64(h, policy.GasPolicy.MaxGasPerBlock)
	writeEngineUint64(h, policy.GasPolicy.MaxGasPerMessage)
	writeEngineUint64(h, policy.ExecutionBudget.MaxGas)
	writeEngineUint64(h, policy.ExecutionBudget.GasUsed)
	writeEngineUint64(h, uint64(policy.ExecutionBudget.MaxMessages))
	writeEngineUint64(h, uint64(policy.ExecutionBudget.MessagesUsed))
	writeEngineUint64(h, uint64(len(policy.MessageFilter.AllowedMessageTypes)))
	for _, item := range policy.MessageFilter.AllowedMessageTypes {
		writeEnginePart(h, item)
	}
	writeEngineUint64(h, uint64(len(policy.AllowedOpcodes)))
	for _, opcode := range policy.AllowedOpcodes {
		writeEnginePart(h, opcode)
	}
	writeEnginePart(h, string(policy.BounceBehavior))
	writeEnginePart(h, string(policy.ProofRequirement))
	writeEnginePart(h, string(policy.ValueAccounting))
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMCrossZoneExecutionHash(execution AVMCrossZoneExecution) string {
	execution = canonicalAVMCrossZoneExecution(execution)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-cross-zone-execution-v1")
	writeEnginePart(h, execution.Message.ID)
	writeEnginePart(h, execution.RoutePolicy.PolicyHash)
	writeEnginePart(h, execution.DestinationQueueEntry.SortKey)
	writeEnginePart(h, execution.DestinationReceipt.ReceiptHash)
	writeEnginePart(h, execution.SourceOutputMessagesRoot)
	writeEnginePart(h, execution.DestinationReceiptRoot)
	writeEngineBool(h, execution.DirectStateWriteAttempted)
	writeEngineUint64(h, execution.ValueEscrowedNAET)
	writeEngineUint64(h, execution.ValueAccountedNAET)
	writeEnginePart(h, string(execution.FailureResolution))
	writeEnginePart(h, execution.BounceMessageOptional.ID)
	writeEnginePart(h, execution.DeadLetterOptional.ReceiptID)
	return hex.EncodeToString(h.Sum(nil))
}

func canonicalAVMCrossZoneRoutePolicy(policy AVMCrossZoneRoutePolicy) AVMCrossZoneRoutePolicy {
	policy.PolicyHash = strings.TrimSpace(policy.PolicyHash)
	policy.AllowedOpcodes = append([]string(nil), policy.AllowedOpcodes...)
	for i := range policy.AllowedOpcodes {
		policy.AllowedOpcodes[i] = strings.TrimSpace(policy.AllowedOpcodes[i])
	}
	sort.Strings(policy.AllowedOpcodes)
	policy.MessageFilter.AllowedMessageTypes = append([]string(nil), policy.MessageFilter.AllowedMessageTypes...)
	for i := range policy.MessageFilter.AllowedMessageTypes {
		policy.MessageFilter.AllowedMessageTypes[i] = strings.TrimSpace(policy.MessageFilter.AllowedMessageTypes[i])
	}
	sort.Strings(policy.MessageFilter.AllowedMessageTypes)
	if len(policy.MessageFilter.AllowedMessageTypes) == 1 && policy.MessageFilter.AllowedMessageTypes[0] == "*" {
		return policy
	}
	return policy
}

func canonicalAVMZoneRouterRoute(route AVMZoneRouterRoute) AVMZoneRouterRoute {
	route.RouteKey = strings.TrimSpace(route.RouteKey)
	route.PolicyHash = strings.TrimSpace(route.PolicyHash)
	route.OutputMessageRoot = strings.TrimSpace(route.OutputMessageRoot)
	route.DestinationInboxRoot = strings.TrimSpace(route.DestinationInboxRoot)
	route.CrossZoneReceiptRoot = strings.TrimSpace(route.CrossZoneReceiptRoot)
	route.ValueEscrowRoot = strings.TrimSpace(route.ValueEscrowRoot)
	route.RouteHash = strings.TrimSpace(route.RouteHash)
	return route
}

func canonicalAVMZoneRouterTable(table AVMZoneRouterTable) AVMZoneRouterTable {
	table.TableRoot = strings.TrimSpace(table.TableRoot)
	table.Routes = append([]AVMZoneRouterRoute(nil), table.Routes...)
	for i := range table.Routes {
		table.Routes[i] = canonicalAVMZoneRouterRoute(table.Routes[i])
	}
	sort.SliceStable(table.Routes, func(i, j int) bool {
		return table.Routes[i].RouteKey < table.Routes[j].RouteKey
	})
	return table
}

func canonicalAVMCrossZoneZoneRoots(roots AVMCrossZoneZoneRoots) AVMCrossZoneZoneRoots {
	roots.OutputMessageRoot = strings.TrimSpace(roots.OutputMessageRoot)
	roots.DestinationInboxRoot = strings.TrimSpace(roots.DestinationInboxRoot)
	roots.CrossZoneReceiptRoot = strings.TrimSpace(roots.CrossZoneReceiptRoot)
	roots.ValueEscrowRoot = strings.TrimSpace(roots.ValueEscrowRoot)
	roots.CrossZoneRootHash = strings.TrimSpace(roots.CrossZoneRootHash)
	return roots
}

func canonicalAVMAetraCoreZoneCommitmentSet(set AVMAetraCoreZoneCommitmentSet) AVMAetraCoreZoneCommitmentSet {
	set.CoreRoot = strings.TrimSpace(set.CoreRoot)
	set.ZoneRoots = append([]AVMZoneStateRoot(nil), set.ZoneRoots...)
	for i := range set.ZoneRoots {
		set.ZoneRoots[i] = canonicalAVMZoneStateRoot(set.ZoneRoots[i])
	}
	sort.SliceStable(set.ZoneRoots, func(i, j int) bool {
		return set.ZoneRoots[i].ZoneID < set.ZoneRoots[j].ZoneID
	})
	return set
}

func canonicalAVMCrossZoneValueEscrowRecord(record AVMCrossZoneValueEscrowRecord) AVMCrossZoneValueEscrowRecord {
	record.MessageID = strings.TrimSpace(record.MessageID)
	record.RefundReceiptID = strings.TrimSpace(record.RefundReceiptID)
	record.EscrowHash = strings.TrimSpace(record.EscrowHash)
	return record
}

func canonicalAVMCrossZoneProof(proof AVMCrossZoneProof) AVMCrossZoneProof {
	proof.MessageID = strings.TrimSpace(proof.MessageID)
	proof.RouteKey = strings.TrimSpace(proof.RouteKey)
	proof.RouterTableRoot = strings.TrimSpace(proof.RouterTableRoot)
	proof.ZoneCrossRoot = strings.TrimSpace(proof.ZoneCrossRoot)
	proof.OutputMessageRoot = strings.TrimSpace(proof.OutputMessageRoot)
	proof.DestinationInboxRoot = strings.TrimSpace(proof.DestinationInboxRoot)
	proof.CrossZoneReceiptRoot = strings.TrimSpace(proof.CrossZoneReceiptRoot)
	proof.ValueEscrowRoot = strings.TrimSpace(proof.ValueEscrowRoot)
	proof.ReceiptID = strings.TrimSpace(proof.ReceiptID)
	proof.EscrowHash = strings.TrimSpace(proof.EscrowHash)
	proof.ProofHash = strings.TrimSpace(proof.ProofHash)
	return proof
}

func canonicalAVMCrossZoneProofIndex(index AVMCrossZoneProofIndex) AVMCrossZoneProofIndex {
	index.RouterTable = canonicalAVMZoneRouterTable(index.RouterTable)
	index.ZoneRoots = append([]AVMCrossZoneZoneRoots(nil), index.ZoneRoots...)
	for i := range index.ZoneRoots {
		index.ZoneRoots[i] = canonicalAVMCrossZoneZoneRoots(index.ZoneRoots[i])
	}
	sort.SliceStable(index.ZoneRoots, func(i, j int) bool {
		return index.ZoneRoots[i].ZoneID < index.ZoneRoots[j].ZoneID
	})
	index.Executions = append([]AVMCrossZoneExecution(nil), index.Executions...)
	for i := range index.Executions {
		index.Executions[i] = canonicalAVMCrossZoneExecution(index.Executions[i])
	}
	sort.SliceStable(index.Executions, func(i, j int) bool {
		return index.Executions[i].Message.ID < index.Executions[j].Message.ID
	})
	index.Escrows = append([]AVMCrossZoneValueEscrowRecord(nil), index.Escrows...)
	for i := range index.Escrows {
		index.Escrows[i] = canonicalAVMCrossZoneValueEscrowRecord(index.Escrows[i])
	}
	sort.SliceStable(index.Escrows, func(i, j int) bool {
		return index.Escrows[i].MessageID < index.Escrows[j].MessageID
	})
	return index
}

func canonicalAVMCrossZoneExecution(execution AVMCrossZoneExecution) AVMCrossZoneExecution {
	execution.Message = canonicalAVMAsyncMessage(execution.Message)
	execution.RoutePolicy = canonicalAVMCrossZoneRoutePolicy(execution.RoutePolicy)
	execution.DestinationQueueEntry = canonicalAVMZoneQueueEntry(execution.DestinationQueueEntry)
	execution.DestinationReceipt = canonicalAVMExecutionReceipt(execution.DestinationReceipt)
	execution.SourceOutputMessagesRoot = strings.TrimSpace(execution.SourceOutputMessagesRoot)
	execution.DestinationReceiptRoot = strings.TrimSpace(execution.DestinationReceiptRoot)
	execution.BounceMessageOptional = canonicalAVMAsyncMessage(execution.BounceMessageOptional)
	execution.DeadLetterOptional = canonicalAVMDeadLetterRecord(execution.DeadLetterOptional)
	execution.ExecutionHash = strings.TrimSpace(execution.ExecutionHash)
	if execution.FailureResolution == "" {
		execution.FailureResolution = AVMCrossZoneFailureNone
	}
	return execution
}

func cloneCrossZoneMessages(messages []AVMAsyncMessage) []AVMAsyncMessage {
	out := append([]AVMAsyncMessage(nil), messages...)
	for i := range out {
		out[i] = canonicalAVMAsyncMessage(out[i])
	}
	return out
}

func findCrossZoneRoots(roots []AVMCrossZoneZoneRoots, zoneID zonestypes.ZoneID) (AVMCrossZoneZoneRoots, bool) {
	for _, root := range roots {
		if root.ZoneID == zoneID {
			return root, true
		}
	}
	return AVMCrossZoneZoneRoots{}, false
}

func findCrossZoneExecution(executions []AVMCrossZoneExecution, messageID string) (AVMCrossZoneExecution, bool) {
	for _, execution := range executions {
		if execution.Message.ID == messageID {
			return execution, true
		}
	}
	return AVMCrossZoneExecution{}, false
}

func findCrossZoneEscrow(records []AVMCrossZoneValueEscrowRecord, messageID string) (AVMCrossZoneValueEscrowRecord, bool) {
	for _, record := range records {
		if record.MessageID == messageID {
			return record, true
		}
	}
	return AVMCrossZoneValueEscrowRecord{}, false
}

func findZoneRouterRoute(routes []AVMZoneRouterRoute, routeKey string) (AVMZoneRouterRoute, bool) {
	for _, route := range routes {
		if route.RouteKey == routeKey {
			return route, true
		}
	}
	return AVMZoneRouterRoute{}, false
}

func validateCrossZoneOpcodes(opcodes []string) error {
	if len(opcodes) == 0 {
		return errors.New("AVM cross-zone policy must define allowed opcodes")
	}
	seen := make(map[string]struct{}, len(opcodes))
	for i, opcode := range opcodes {
		if err := validateRouterOptionalToken("AVM cross-zone opcode", opcode, MaxAVMCrossZoneOpcodeLength); err != nil {
			return err
		}
		if opcode == "" {
			return errors.New("AVM cross-zone opcode is required")
		}
		if _, found := seen[opcode]; found {
			return fmt.Errorf("duplicate AVM cross-zone opcode %q", opcode)
		}
		seen[opcode] = struct{}{}
		if i > 0 && opcodes[i-1] >= opcode {
			return errors.New("AVM cross-zone opcodes must be sorted canonically")
		}
	}
	return nil
}

func validateCrossZoneProofs(msg AVMAsyncMessage, requirement AVMCrossZoneProofRequirement) error {
	switch requirement {
	case AVMCrossZoneProofNone:
		return nil
	case AVMCrossZoneProofAuth:
		if msg.AuthProofOptional == "" {
			return errors.New("AVM cross-zone route requires auth proof")
		}
	case AVMCrossZoneProofState:
		if msg.StateProofOptional == "" {
			return errors.New("AVM cross-zone route requires state proof")
		}
	case AVMCrossZoneProofAuthAndState:
		if msg.AuthProofOptional == "" || msg.StateProofOptional == "" {
			return errors.New("AVM cross-zone route requires auth and state proofs")
		}
	default:
		return fmt.Errorf("invalid AVM cross-zone proof requirement %q", requirement)
	}
	return nil
}

func crossZoneOpcodeAllowed(opcodes []string, opcode string) bool {
	for _, allowed := range opcodes {
		if allowed == opcode {
			return true
		}
	}
	return false
}

func validateCrossZoneBounceMessage(original AVMAsyncMessage, bounce AVMAsyncMessage) error {
	bounce = canonicalAVMAsyncMessage(bounce)
	if err := bounce.Validate(); err != nil {
		return err
	}
	if bounce.SourceZone != original.DestinationZone || bounce.DestinationZone != original.SourceZone {
		return errors.New("AVM cross-zone bounce message must reverse source and destination zones")
	}
	if !strings.Contains(bounce.RouteHintOptional, original.ID) {
		return errors.New("AVM cross-zone bounce message must reference original message")
	}
	return nil
}

func isCrossZoneFailureStatus(status AVMReceiptStatus) bool {
	switch status {
	case AVMReceiptStatusFailed, AVMReceiptStatusBounced, AVMReceiptStatusDeadLettered:
		return true
	default:
		return false
	}
}
