package types

import (
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	MaxAetherMeshPayloadBytes	= MaxStreamMessageBytes
	MaxAetherMeshTTL		= uint64(10_000)
)

type AetherMeshMessageType string

const (
	MeshMessageConsensus	AetherMeshMessageType	= "consensus"
	MeshMessageTx		AetherMeshMessageType	= "tx"
	MeshMessageExecution	AetherMeshMessageType	= "execution"
	MeshMessageQuery	AetherMeshMessageType	= "query"
	MeshMessageService	AetherMeshMessageType	= "service"
	MeshMessageCrossZone	AetherMeshMessageType	= "cross_zone"
	MeshMessageStateSync	AetherMeshMessageType	= "state_sync"
	MeshMessageStorage	AetherMeshMessageType	= "storage"
	MeshMessageRouting	AetherMeshMessageType	= "routing"
)

type AetherMeshProof struct {
	ProofType	string
	ProofHash	string
	ProofHeight	uint64
}

type AetherMeshMessage struct {
	Type			AetherMeshMessageType
	Payload			[]byte
	Origin			string
	Destination		string
	Priority		uint32
	TTL			uint64
	MessageID		string
	OverlayID		string
	SourceZone		string
	DestinationZone		string
	Sequence		uint64
	PayloadHash		string
	RouteHint		RouteHint
	DeadlineHeight		uint64
	Signature		[]byte
	Proof			AetherMeshProof
	ConsensusEffect		bool
	DeterminismSource	DeterminismSource
}

type AetherMeshRouteRequest struct {
	Message			AetherMeshMessage
	SourceNodeID		string
	CandidatePeers		[]NodeRecord
	MembershipProofs	[]OverlayMembershipProof
	Graph			RoutingGraph
	CurrentHeight		uint64
}

type AetherMeshDelivery struct {
	Message	AetherMeshMessage
	Route	OverlayRoutePlan
	Channel	ChannelClass
}

func NewAetherMeshMessage(msg AetherMeshMessage) (AetherMeshMessage, error) {
	msg = NormalizeAetherMeshMessage(msg)
	if msg.PayloadHash == "" && len(msg.Payload) > 0 {
		msg.PayloadHash = hashBytes("aetra-mesh-payload-v1", msg.Payload)
	}
	if msg.MessageID == "" {
		msg.MessageID = ComputeAetherMeshMessageID(msg)
	}
	if err := msg.ValidateBasic(0); err != nil {
		return AetherMeshMessage{}, err
	}
	return msg, nil
}

func SignAetherMeshMessage(msg AetherMeshMessage, privateKey ed25519.PrivateKey) (AetherMeshMessage, error) {
	if len(privateKey) != ed25519.PrivateKeySize {
		return AetherMeshMessage{}, errors.New("networking mesh private key must be ed25519")
	}
	msg.Signature = nil
	msg, err := NewAetherMeshMessage(msg)
	if err != nil {
		return AetherMeshMessage{}, err
	}
	payload, err := msg.SigningPayload()
	if err != nil {
		return AetherMeshMessage{}, err
	}
	msg.Signature = ed25519.Sign(privateKey, payload)
	if err := msg.ValidateBasic(0); err != nil {
		return AetherMeshMessage{}, err
	}
	return msg, nil
}

func NormalizeAetherMeshMessage(msg AetherMeshMessage) AetherMeshMessage {
	msg.Type = AetherMeshMessageType(strings.ToLower(strings.TrimSpace(string(msg.Type))))
	msg.Payload = cloneBytes(msg.Payload)
	msg.Origin = normalizeHashText(msg.Origin)
	msg.Destination = normalizeHashText(msg.Destination)
	msg.MessageID = normalizeHashText(msg.MessageID)
	msg.OverlayID = normalizeHashText(msg.OverlayID)
	msg.SourceZone = strings.TrimSpace(msg.SourceZone)
	msg.DestinationZone = strings.TrimSpace(msg.DestinationZone)
	msg.PayloadHash = normalizeHashText(msg.PayloadHash)
	msg.RouteHint.ZoneID = strings.TrimSpace(msg.RouteHint.ZoneID)
	msg.RouteHint.ShardID = strings.TrimSpace(msg.RouteHint.ShardID)
	msg.RouteHint.ServiceID = strings.TrimSpace(msg.RouteHint.ServiceID)
	msg.RouteHint.StorageKeyHash = normalizeHashText(msg.RouteHint.StorageKeyHash)
	msg.RouteHint.DeterministicHintHash = normalizeHashText(msg.RouteHint.DeterministicHintHash)
	msg.Signature = cloneBytes(msg.Signature)
	msg.Proof.ProofType = strings.TrimSpace(msg.Proof.ProofType)
	msg.Proof.ProofHash = normalizeHashText(msg.Proof.ProofHash)
	if msg.DeterminismSource == "" && msg.ConsensusEffect {
		msg.DeterminismSource = DeterminismReplaySafeMessageID
	}
	return msg
}

func ComputeAetherMeshMessageID(msg AetherMeshMessage) string {
	msg = NormalizeAetherMeshMessage(msg)
	return HashParts(
		"aether-mesh-message",
		string(msg.Type),
		msg.Origin,
		msg.Destination,
		fmt.Sprintf("%d", msg.Priority),
		fmt.Sprintf("%d", msg.TTL),
		msg.OverlayID,
		msg.SourceZone,
		msg.DestinationZone,
		fmt.Sprintf("%d", msg.Sequence),
		msg.PayloadHash,
		msg.RouteHint.ZoneID,
		msg.RouteHint.ShardID,
		msg.RouteHint.ServiceID,
		msg.RouteHint.StorageKeyHash,
		msg.RouteHint.DeterministicHintHash,
		fmt.Sprintf("%d", msg.DeadlineHeight),
		msg.Proof.ProofHash,
		fmt.Sprintf("%d", msg.Proof.ProofHeight),
		fmt.Sprintf("%t", msg.ConsensusEffect),
		string(msg.DeterminismSource),
	)
}

func (m AetherMeshMessage) SigningPayload() ([]byte, error) {
	msg := NormalizeAetherMeshMessage(m)
	msg.Signature = nil
	return json.Marshal(msg)
}

func (m AetherMeshMessage) ValidateBasic(currentHeight uint64) error {
	msg := NormalizeAetherMeshMessage(m)
	if !IsAetherMeshMessageType(msg.Type) {
		return fmt.Errorf("unknown networking mesh message type %q", msg.Type)
	}
	if err := ValidateHash("networking mesh origin", msg.Origin); err != nil {
		return err
	}
	if err := ValidateHash("networking mesh destination", msg.Destination); err != nil {
		return err
	}
	if err := ValidateHash("networking mesh message id", msg.MessageID); err != nil {
		return err
	}
	if msg.MessageID != ComputeAetherMeshMessageID(msg) {
		return errors.New("networking mesh message id does not match payload")
	}
	if err := ValidateHash("networking mesh overlay id", msg.OverlayID); err != nil {
		return err
	}
	if msg.Sequence == 0 {
		return errors.New("networking mesh sequence must be positive")
	}
	if msg.TTL == 0 || msg.TTL > MaxAetherMeshTTL {
		return fmt.Errorf("networking mesh ttl must be between 1 and %d", MaxAetherMeshTTL)
	}
	if msg.DeadlineHeight > 0 && currentHeight > 0 && currentHeight > msg.DeadlineHeight {
		return errors.New("networking mesh message deadline expired")
	}
	if len(msg.Payload) == 0 || len(msg.Payload) > MaxAetherMeshPayloadBytes {
		return fmt.Errorf("networking mesh payload bytes must be between 1 and %d", MaxAetherMeshPayloadBytes)
	}
	if hashBytes("aetra-mesh-payload-v1", msg.Payload) != msg.PayloadHash {
		return errors.New("networking mesh payload hash mismatch")
	}
	if err := ValidateHash("networking mesh payload hash", msg.PayloadHash); err != nil {
		return err
	}
	if msg.RouteHint.StorageKeyHash != "" {
		if err := ValidateHash("networking mesh storage key hash", msg.RouteHint.StorageKeyHash); err != nil {
			return err
		}
	}
	if msg.RouteHint.DeterministicHintHash != "" {
		if err := ValidateHash("networking mesh deterministic route hint", msg.RouteHint.DeterministicHintHash); err != nil {
			return err
		}
	}
	if err := msg.Proof.Validate(msg.ConsensusEffect); err != nil {
		return err
	}
	if msg.ConsensusEffect && !IsConsensusSafeDeterminismSource(msg.DeterminismSource) {
		return fmt.Errorf("networking mesh consensus message requires deterministic source, got %q", msg.DeterminismSource)
	}
	if msg.ConsensusEffect && msg.Proof.ProofHash == "" && msg.Type == MeshMessageService {
		return errors.New("networking mesh service message cannot affect consensus without proof")
	}
	if len(msg.Signature) > 0 && len(msg.Signature) != ed25519.SignatureSize {
		return fmt.Errorf("networking mesh signature must be %d bytes", ed25519.SignatureSize)
	}
	return validateMeshZoneRequirements(msg)
}

func (p AetherMeshProof) Validate(required bool) error {
	proofHash := normalizeHashText(p.ProofHash)
	if proofHash == "" {
		if required {
			return errors.New("networking mesh proof is required")
		}
		return nil
	}
	if strings.TrimSpace(p.ProofType) == "" {
		return errors.New("networking mesh proof type is required")
	}
	if err := ValidateHash("networking mesh proof hash", proofHash); err != nil {
		return err
	}
	if p.ProofHeight == 0 {
		return errors.New("networking mesh proof height must be positive")
	}
	return nil
}

func VerifyAetherMeshMessageSignature(msg AetherMeshMessage, pubKey ed25519.PublicKey, currentHeight uint64) error {
	msg = NormalizeAetherMeshMessage(msg)
	if err := msg.ValidateBasic(currentHeight); err != nil {
		return err
	}
	if len(pubKey) != ed25519.PublicKeySize {
		return fmt.Errorf("networking mesh public key must be %d bytes", ed25519.PublicKeySize)
	}
	if len(msg.Signature) != ed25519.SignatureSize {
		return fmt.Errorf("networking mesh signature must be %d bytes", ed25519.SignatureSize)
	}
	payload, err := msg.SigningPayload()
	if err != nil {
		return err
	}
	if !ed25519.Verify(pubKey, payload, msg.Signature) {
		return errors.New("networking mesh signature verification failed")
	}
	return nil
}

func (m AetherMeshMessage) ToNetworkMessage() (NetworkMessage, error) {
	msg := NormalizeAetherMeshMessage(m)
	if err := msg.ValidateBasic(0); err != nil {
		return NetworkMessage{}, err
	}
	return NewNetworkMessage(NetworkMessage{
		Layer:			LayerL3Application,
		Channel:		channelForMeshMessageType(msg.Type),
		ConsensusEffect:	msg.ConsensusEffect,
		DeterminismSource:	msg.DeterminismSource,
		ReplaySafeID:		msg.MessageID,
		PayloadHash:		msg.PayloadHash,
		PayloadSizeBytes:	uint64(len(msg.Payload)),
		Chunked:		uint64(len(msg.Payload)) > LargePayloadBytes,
		CommitmentVerified:	msg.Proof.ProofHash != "",
		CommittedProofHash:	msg.Proof.ProofHash,
	})
}

func RouteAetherMeshMessage(req AetherMeshRouteRequest, descriptors []OverlayDescriptor) (AetherMeshDelivery, error) {
	msg := NormalizeAetherMeshMessage(req.Message)
	if err := msg.ValidateBasic(req.CurrentHeight); err != nil {
		return AetherMeshDelivery{}, err
	}
	base, err := msg.ToNetworkMessage()
	if err != nil {
		return AetherMeshDelivery{}, err
	}
	plan, err := BuildOverlayRoute(OverlayRoutingRequest{
		Message:		base,
		SourceNodeID:		req.SourceNodeID,
		CandidatePeers:		req.CandidatePeers,
		MembershipProofs:	req.MembershipProofs,
		Graph:			req.Graph,
		Hint:			msg.RouteHint,
		CurrentHeight:		req.CurrentHeight,
	}, descriptors)
	if err != nil {
		return AetherMeshDelivery{}, err
	}
	if plan.OverlayID != msg.OverlayID {
		return AetherMeshDelivery{}, errors.New("networking mesh route overlay mismatch")
	}
	return AetherMeshDelivery{
		Message:	msg,
		Route:		plan,
		Channel:	channelForMeshMessageType(msg.Type),
	}, nil
}

func IsAetherMeshMessageType(messageType AetherMeshMessageType) bool {
	switch messageType {
	case MeshMessageConsensus,
		MeshMessageTx,
		MeshMessageExecution,
		MeshMessageQuery,
		MeshMessageService,
		MeshMessageCrossZone,
		MeshMessageStateSync,
		MeshMessageStorage,
		MeshMessageRouting:
		return true
	default:
		return false
	}
}

func AetherMeshMessageTypes() []AetherMeshMessageType {
	out := []AetherMeshMessageType{
		MeshMessageConsensus,
		MeshMessageTx,
		MeshMessageExecution,
		MeshMessageQuery,
		MeshMessageService,
		MeshMessageCrossZone,
		MeshMessageStateSync,
		MeshMessageStorage,
		MeshMessageRouting,
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func channelForMeshMessageType(messageType AetherMeshMessageType) ChannelClass {
	switch messageType {
	case MeshMessageConsensus:
		return ChannelConsensus
	case MeshMessageTx:
		return ChannelMempool
	case MeshMessageExecution, MeshMessageCrossZone:
		return ChannelExecution
	case MeshMessageQuery, MeshMessageService:
		return ChannelService
	case MeshMessageStateSync:
		return ChannelStateSync
	case MeshMessageStorage:
		return ChannelData
	case MeshMessageRouting:
		return ChannelRouting
	default:
		return ChannelService
	}
}

func validateMeshZoneRequirements(msg AetherMeshMessage) error {
	switch msg.Type {
	case MeshMessageExecution:
		if msg.DestinationZone == "" {
			return errors.New("networking mesh execution message requires destination zone")
		}
	case MeshMessageCrossZone:
		if msg.SourceZone == "" || msg.DestinationZone == "" {
			return errors.New("networking mesh cross-zone message requires source and destination zones")
		}
		if msg.SourceZone == msg.DestinationZone {
			return errors.New("networking mesh cross-zone message requires different zones")
		}
	}
	return nil
}
