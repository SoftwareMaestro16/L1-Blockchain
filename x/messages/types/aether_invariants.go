package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"
)

const (
	MsgBusModuleName	= "msgbus"

	MsgBusEnvelopePrefix	= "msgbus/envelopes"
	MsgBusOutboxPrefix	= "msgbus/outbox"
	MsgBusInboxPrefix	= "msgbus/inbox"
	MsgBusReceiptPrefix	= "msgbus/receipts"
	MsgBusReplayPrefix	= "msgbus/replay"
	MsgBusEscrowPrefix	= "msgbus/escrow"
)

type AetherEscrowStatus string
type AetherProofKind string

const (
	AetherEscrowLocked	AetherEscrowStatus	= "locked"
	AetherEscrowReleased	AetherEscrowStatus	= "released"
	AetherEscrowRefunded	AetherEscrowStatus	= "refunded"
	AetherEscrowBounced	AetherEscrowStatus	= "bounced"

	AetherProofMessageInclusion	AetherProofKind	= "message_inclusion"
	AetherProofReceiptInclusion	AetherProofKind	= "receipt_inclusion"
)

type AetherValueEscrow struct {
	MsgID		string
	ValueLocked	sdkmath.Int
	FeeLocked	sdkmath.Int
	Status		AetherEscrowStatus
	ReceiptHash	string
	EscrowHash	string
}

type AetherMsgBusState struct {
	Outbox		[]AetherMessage
	Inbox		[]AetherMessage
	Receipts	[]AetherMessageReceipt
	Escrows		[]AetherValueEscrow
	ReplayMsgIDs	[]string
	MessageRoot	string
	ReceiptRoot	string
	StateRoot	string
}

type AetherInclusionProof struct {
	Kind		AetherProofKind
	MsgID		string
	Root		string
	ValueHash	string
	Path		[]string
	ProofHash	string
}

type AetherPayloadExecutionPolicy struct {
	NoExternalAPIs		bool
	NoWallClockTime		bool
	MeteredIteration	bool
	PolicyHash		string
}

func NewAetherMsgBusState(outbox []AetherMessage, inbox []AetherMessage, receipts []AetherMessageReceipt, escrows []AetherValueEscrow, replayMsgIDs []string) (AetherMsgBusState, error) {
	state := AetherMsgBusState{
		Outbox:		cloneAetherMessages(outbox),
		Inbox:		cloneAetherMessages(inbox),
		Receipts:	cloneAetherMessageReceipts(receipts),
		Escrows:	cloneAetherValueEscrows(escrows),
		ReplayMsgIDs:	append([]string(nil), replayMsgIDs...),
	}
	sortAetherMessages(state.Outbox)
	sortAetherMessages(state.Inbox)
	sortAetherMessageReceipts(state.Receipts)
	sortAetherValueEscrows(state.Escrows)
	sort.Strings(state.ReplayMsgIDs)
	if err := ValidateAetherMessageInvariants(state); err != nil {
		return AetherMsgBusState{}, err
	}
	state.MessageRoot = ComputeAetherGlobalMessageRoot(state.Outbox, state.Inbox)
	receiptRoot, err := ComputeAetherReceiptRoot(state.Receipts)
	if err != nil {
		return AetherMsgBusState{}, err
	}
	state.ReceiptRoot = receiptRoot
	state.StateRoot = ComputeAetherMsgBusStateRoot(state)
	return state, state.Validate()
}

func (s AetherMsgBusState) Validate() error {
	if err := ValidateAetherMessageInvariants(s); err != nil {
		return err
	}
	if s.MessageRoot != ComputeAetherGlobalMessageRoot(s.Outbox, s.Inbox) {
		return errors.New("aether msgbus message root mismatch")
	}
	receiptRoot, err := ComputeAetherReceiptRoot(s.Receipts)
	if err != nil {
		return err
	}
	if s.ReceiptRoot != receiptRoot {
		return errors.New("aether msgbus receipt root mismatch")
	}
	if s.StateRoot != ComputeAetherMsgBusStateRoot(s) {
		return errors.New("aether msgbus state root mismatch")
	}
	return nil
}

func ValidateAetherMessageInvariants(state AetherMsgBusState) error {
	seenMessages := make(map[string]struct{}, len(state.Outbox)+len(state.Inbox))
	for _, msg := range append(cloneAetherMessages(state.Outbox), cloneAetherMessages(state.Inbox)...) {
		if err := msg.Validate(); err != nil {
			return err
		}
		if _, found := seenMessages[msg.MsgID]; found {
			return errors.New("aether invariant: msg_id must be globally unique")
		}
		seenMessages[msg.MsgID] = struct{}{}
	}
	seenReceipts := make(map[string]struct{}, len(state.Receipts))
	for _, receipt := range state.Receipts {
		if err := receipt.Validate(); err != nil {
			return err
		}
		if _, found := seenReceipts[receipt.MsgID]; found {
			return errors.New("aether invariant: duplicate receipt")
		}
		seenReceipts[receipt.MsgID] = struct{}{}
	}
	seenReplay := make(map[string]struct{}, len(state.ReplayMsgIDs))
	for _, msgID := range state.ReplayMsgIDs {
		msgID = strings.ToLower(strings.TrimSpace(msgID))
		if err := validateOptionalHash("aether replay message id", msgID); err != nil {
			return err
		}
		if _, found := seenReplay[msgID]; found {
			return errors.New("aether invariant: duplicate replay entry")
		}
		seenReplay[msgID] = struct{}{}
	}
	escrowsByMsg := make(map[string]AetherValueEscrow, len(state.Escrows))
	for _, escrow := range state.Escrows {
		if err := escrow.Validate(); err != nil {
			return err
		}
		if _, found := escrowsByMsg[escrow.MsgID]; found {
			return errors.New("aether invariant: duplicate value escrow")
		}
		escrowsByMsg[escrow.MsgID] = escrow
	}
	for _, msg := range state.Outbox {
		escrow, found := escrowsByMsg[msg.MsgID]
		if !found {
			return errors.New("aether invariant: source outbox requires value escrow")
		}
		if escrow.Status != AetherEscrowLocked {
			return errors.New("aether invariant: source outbox escrow must be locked")
		}
		if escrow.ValueLocked.LT(msg.ValueNAET) || escrow.FeeLocked.LT(msg.ForwardingFee) {
			return errors.New("aether invariant: escrow must cover message value and forwarding fee")
		}
	}
	for _, receipt := range state.Receipts {
		expiryHeight := messageExpiryHeight(state, receipt.MsgID)
		if expiryHeight != 0 && receipt.Status == ReceiptStatusExecuted && receipt.Height > expiryHeight {
			return errors.New("aether invariant: expired message cannot execute")
		}
	}
	return nil
}

func ApplyAetherInboundMessage(msg AetherMessage, height uint64, execute func(AetherMessage) (AetherMessageReceipt, error)) (AetherMessageReceipt, error) {
	if err := msg.Validate(); err != nil {
		return AetherMessageReceipt{}, err
	}
	if height == 0 {
		return AetherMessageReceipt{}, errors.New("aether inbound height must be positive")
	}
	if height > msg.ExpiryHeight {
		return ExpireAetherMessage(msg, height, EmptyHash(), hashParts("aether-expired-state-write-summary", msg.MsgID, fmt.Sprint(height)))
	}
	if execute == nil {
		return AetherMessageReceipt{}, errors.New("aether inbound executor is required")
	}
	receipt, err := execute(msg.Clone())
	if err != nil {
		return AetherReceiptFromMessage(msg, ReceiptStatusFailed, height, 0, sdkmath.ZeroInt(), nil, "ERR_EXECUTOR", EmptyHash(), hashParts("aether-failed-state-write-summary", msg.MsgID, fmt.Sprint(height)))
	}
	if receipt.MsgID != msg.MsgID {
		return AetherMessageReceipt{}, errors.New("aether inbound receipt message mismatch")
	}
	if receipt.Height == 0 {
		receipt.Height = height
		receipt.ReceiptHash = ""
		return NewAetherMessageReceipt(receipt)
	}
	return receipt, receipt.Validate()
}

func BuildAetherMessageInclusionProof(messages []AetherMessage, msgID string, root string) (AetherInclusionProof, error) {
	ordered := cloneAetherMessages(messages)
	sortAetherMessages(ordered)
	if root == "" {
		root = ComputeAetherMessageListRoot(ordered)
	}
	proof := AetherInclusionProof{Kind: AetherProofMessageInclusion, MsgID: strings.ToLower(strings.TrimSpace(msgID)), Root: root}
	for _, msg := range ordered {
		if err := msg.Validate(); err != nil {
			return AetherInclusionProof{}, err
		}
		value := ComputeAetherMessageLeafHash(msg)
		proof.Path = append(proof.Path, value)
		if msg.MsgID == proof.MsgID {
			proof.ValueHash = value
		}
	}
	if proof.ValueHash == "" {
		return AetherInclusionProof{}, errors.New("aether message inclusion proof item not found")
	}
	proof.ProofHash = ComputeAetherInclusionProofHash(proof)
	return proof, proof.Validate()
}

func BuildAetherReceiptInclusionProof(receipts []AetherMessageReceipt, msgID string, root string) (AetherInclusionProof, error) {
	ordered := cloneAetherMessageReceipts(receipts)
	sortAetherMessageReceipts(ordered)
	if root == "" {
		var err error
		root, err = ComputeAetherReceiptRoot(ordered)
		if err != nil {
			return AetherInclusionProof{}, err
		}
	}
	proof := AetherInclusionProof{Kind: AetherProofReceiptInclusion, MsgID: strings.ToLower(strings.TrimSpace(msgID)), Root: root}
	for _, receipt := range ordered {
		if err := receipt.Validate(); err != nil {
			return AetherInclusionProof{}, err
		}
		proof.Path = append(proof.Path, receipt.ReceiptHash)
		if receipt.MsgID == proof.MsgID {
			proof.ValueHash = receipt.ReceiptHash
		}
	}
	if proof.ValueHash == "" {
		return AetherInclusionProof{}, errors.New("aether receipt inclusion proof item not found")
	}
	proof.ProofHash = ComputeAetherInclusionProofHash(proof)
	return proof, proof.Validate()
}

func ValidateAetherRouteReproducible(msg AetherMessage, plan AetherRoutePlan) error {
	if err := msg.Validate(); err != nil {
		return err
	}
	if err := plan.Validate(); err != nil {
		return err
	}
	if msg.RouteCommitment != plan.RouteCommitment {
		return errors.New("aether invariant: route commitment mismatch")
	}
	return nil
}

func (e AetherValueEscrow) Validate() error {
	e = normalizeAetherValueEscrow(e)
	if err := validateOptionalHash("aether escrow message id", e.MsgID); err != nil {
		return err
	}
	if e.ValueLocked.IsNil() || e.FeeLocked.IsNil() || e.ValueLocked.IsNegative() || e.FeeLocked.IsNegative() {
		return errors.New("aether escrow value and fee must be non-negative")
	}
	if !IsAetherEscrowStatus(e.Status) {
		return fmt.Errorf("unknown aether escrow status %q", e.Status)
	}
	if e.ReceiptHash != "" {
		if err := validateOptionalHash("aether escrow receipt hash", e.ReceiptHash); err != nil {
			return err
		}
	}
	if err := validateOptionalHash("aether escrow hash", e.EscrowHash); err != nil {
		return err
	}
	if e.EscrowHash != "" && e.EscrowHash != ComputeAetherValueEscrowHash(e) {
		return errors.New("aether escrow hash mismatch")
	}
	return nil
}

func (p AetherInclusionProof) Validate() error {
	if !IsAetherProofKind(p.Kind) {
		return fmt.Errorf("unknown aether proof kind %q", p.Kind)
	}
	if err := validateOptionalHash("aether proof message id", p.MsgID); err != nil {
		return err
	}
	if err := validateOptionalHash("aether proof root", p.Root); err != nil {
		return err
	}
	if err := validateOptionalHash("aether proof value hash", p.ValueHash); err != nil {
		return err
	}
	if len(p.Path) == 0 || len(p.Path) > MaxProofPathItems {
		return errors.New("aether proof path length is invalid")
	}
	for _, item := range p.Path {
		if err := validateOptionalHash("aether proof path item", item); err != nil {
			return err
		}
	}
	if err := validateOptionalHash("aether proof hash", p.ProofHash); err != nil {
		return err
	}
	if p.ProofHash != ComputeAetherInclusionProofHash(p) {
		return errors.New("aether proof hash mismatch")
	}
	return nil
}

func (p AetherPayloadExecutionPolicy) Validate() error {
	if !p.NoExternalAPIs {
		return errors.New("aether payload policy must prohibit external APIs")
	}
	if !p.NoWallClockTime {
		return errors.New("aether payload policy must prohibit wall-clock time")
	}
	if !p.MeteredIteration {
		return errors.New("aether payload policy must require metered iteration")
	}
	if p.PolicyHash != "" {
		if err := validateOptionalHash("aether payload policy hash", p.PolicyHash); err != nil {
			return err
		}
		if p.PolicyHash != ComputeAetherPayloadExecutionPolicyHash(p) {
			return errors.New("aether payload policy hash mismatch")
		}
	}
	return nil
}

func NewAetherValueEscrow(escrow AetherValueEscrow) (AetherValueEscrow, error) {
	if escrow.EscrowHash != "" {
		return AetherValueEscrow{}, errors.New("aether escrow hash must be empty before construction")
	}
	escrow = normalizeAetherValueEscrow(escrow)
	if err := escrow.Validate(); err != nil {
		return AetherValueEscrow{}, err
	}
	escrow.EscrowHash = ComputeAetherValueEscrowHash(escrow)
	return escrow, escrow.Validate()
}

func ComputeAetherGlobalMessageRoot(outbox []AetherMessage, inbox []AetherMessage) string {
	return hashParts("aetra-aether-global-message-root-v1", ComputeAetherMessageListRoot(outbox), ComputeAetherMessageListRoot(inbox))
}

func ComputeAetherMessageListRoot(messages []AetherMessage) string {
	ordered := cloneAetherMessages(messages)
	sortAetherMessages(ordered)
	parts := []string{"aetra-aether-message-list-root-v1", fmt.Sprint(len(ordered))}
	for _, msg := range ordered {
		parts = append(parts, ComputeAetherMessageLeafHash(msg))
	}
	return hashParts(parts...)
}

func ComputeAetherMessageLeafHash(msg AetherMessage) string {
	return hashParts("aetra-aether-message-leaf-v1", msg.MsgID, ComputeAetherPayloadHash(msg.Payload), msg.RouteCommitment, msg.ValueNAET.String(), fmt.Sprint(msg.CreatedAtHeight), fmt.Sprint(msg.Nonce))
}

func ComputeAetherMsgBusStateRoot(state AetherMsgBusState) string {
	escrowRoot := ComputeAetherEscrowRoot(state.Escrows)
	replayRoot := ComputeAetherReplayRoot(state.ReplayMsgIDs)
	return hashParts("aetra-aether-msgbus-state-root-v1", state.MessageRoot, state.ReceiptRoot, escrowRoot, replayRoot)
}

func ComputeAetherEscrowRoot(escrows []AetherValueEscrow) string {
	ordered := cloneAetherValueEscrows(escrows)
	sortAetherValueEscrows(ordered)
	parts := []string{"aetra-aether-escrow-root-v1", fmt.Sprint(len(ordered))}
	for _, escrow := range ordered {
		parts = append(parts, escrow.EscrowHash)
	}
	return hashParts(parts...)
}

func ComputeAetherReplayRoot(msgIDs []string) string {
	ordered := append([]string(nil), msgIDs...)
	sort.Strings(ordered)
	parts := []string{"aetra-aether-replay-root-v1", fmt.Sprint(len(ordered))}
	parts = append(parts, ordered...)
	return hashParts(parts...)
}

func ComputeAetherValueEscrowHash(escrow AetherValueEscrow) string {
	escrow = normalizeAetherValueEscrow(escrow)
	return hashParts("aetra-aether-value-escrow-v1", escrow.MsgID, escrow.ValueLocked.String(), escrow.FeeLocked.String(), string(escrow.Status), escrow.ReceiptHash)
}

func ComputeAetherInclusionProofHash(proof AetherInclusionProof) string {
	parts := []string{"aetra-aether-inclusion-proof-v1", string(proof.Kind), proof.MsgID, proof.Root, proof.ValueHash}
	parts = append(parts, proof.Path...)
	return hashParts(parts...)
}

func ComputeAetherPayloadExecutionPolicyHash(policy AetherPayloadExecutionPolicy) string {
	return hashParts("aetra-aether-payload-policy-v1", fmt.Sprint(policy.NoExternalAPIs), fmt.Sprint(policy.NoWallClockTime), fmt.Sprint(policy.MeteredIteration))
}

func MsgBusEnvelopeKey(msgID string) (string, error) {
	return msgBusHashKey(MsgBusEnvelopePrefix, "message id", msgID)
}

func MsgBusOutboxKey(zoneID string, shardID string, msgID string) (string, error) {
	return msgBusScopedKey(MsgBusOutboxPrefix, zoneID, shardID, msgID)
}

func MsgBusInboxKey(zoneID string, shardID string, msgID string) (string, error) {
	return msgBusScopedKey(MsgBusInboxPrefix, zoneID, shardID, msgID)
}

func MsgBusReceiptKey(msgID string) (string, error) {
	return msgBusHashKey(MsgBusReceiptPrefix, "message id", msgID)
}

func MsgBusReplayKey(msgID string) (string, error) {
	return msgBusHashKey(MsgBusReplayPrefix, "message id", msgID)
}

func MsgBusEscrowKey(msgID string) (string, error) {
	return msgBusHashKey(MsgBusEscrowPrefix, "message id", msgID)
}

func IsAetherEscrowStatus(status AetherEscrowStatus) bool {
	switch status {
	case AetherEscrowLocked, AetherEscrowReleased, AetherEscrowRefunded, AetherEscrowBounced:
		return true
	default:
		return false
	}
}

func IsAetherProofKind(kind AetherProofKind) bool {
	switch kind {
	case AetherProofMessageInclusion, AetherProofReceiptInclusion:
		return true
	default:
		return false
	}
}

func normalizeAetherValueEscrow(escrow AetherValueEscrow) AetherValueEscrow {
	escrow.MsgID = strings.ToLower(strings.TrimSpace(escrow.MsgID))
	if escrow.ValueLocked.IsNil() {
		escrow.ValueLocked = sdkmath.ZeroInt()
	}
	if escrow.FeeLocked.IsNil() {
		escrow.FeeLocked = sdkmath.ZeroInt()
	}
	escrow.ReceiptHash = strings.ToLower(strings.TrimSpace(escrow.ReceiptHash))
	escrow.EscrowHash = strings.ToLower(strings.TrimSpace(escrow.EscrowHash))
	return escrow
}

func messageExpiryHeight(state AetherMsgBusState, msgID string) uint64 {
	for _, msg := range append(cloneAetherMessages(state.Outbox), cloneAetherMessages(state.Inbox)...) {
		if msg.MsgID == msgID {
			return msg.ExpiryHeight
		}
	}
	return 0
}

func msgBusHashKey(prefix string, fieldName string, hash string) (string, error) {
	hash = strings.ToLower(strings.TrimSpace(hash))
	if err := validateOptionalHash("msgbus "+fieldName, hash); err != nil {
		return "", err
	}
	return prefix + "/" + hash, nil
}

func msgBusScopedKey(prefix string, zoneID string, shardID string, msgID string) (string, error) {
	if err := validateToken("msgbus zone id", zoneID, MaxAetherAddressLength); err != nil {
		return "", err
	}
	if err := validateToken("msgbus shard id", shardID, MaxShardIDLength); err != nil {
		return "", err
	}
	msgID = strings.ToLower(strings.TrimSpace(msgID))
	if err := validateOptionalHash("msgbus message id", msgID); err != nil {
		return "", err
	}
	return prefix + "/" + zoneID + "/" + shardID + "/" + msgID, nil
}

func cloneAetherMessages(messages []AetherMessage) []AetherMessage {
	out := make([]AetherMessage, len(messages))
	for i, msg := range messages {
		out[i] = msg.Clone()
	}
	return out
}

func sortAetherMessages(messages []AetherMessage) {
	sort.SliceStable(messages, func(i, j int) bool { return messages[i].MsgID < messages[j].MsgID })
}

func cloneAetherValueEscrows(escrows []AetherValueEscrow) []AetherValueEscrow {
	out := make([]AetherValueEscrow, len(escrows))
	copy(out, escrows)
	return out
}

func sortAetherValueEscrows(escrows []AetherValueEscrow) {
	sort.SliceStable(escrows, func(i, j int) bool { return escrows[i].MsgID < escrows[j].MsgID })
}
