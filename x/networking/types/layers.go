package types

import (
	"errors"
	"fmt"
	"strings"
)

const (
	BaseTransportCometBFTP2P	= "COMETBFT_P2P"
	LargePayloadBytes		= uint64(DefaultMaxMessageBytes)
)

type NetworkLayer string

const (
	LayerL0Physical		NetworkLayer	= "L0_PHYSICAL_TRANSPORT"
	LayerL1Session		NetworkLayer	= "L1_SECURE_IDENTITY_SESSION"
	LayerL2Overlay		NetworkLayer	= "L2_OVERLAY_ROUTING"
	LayerL3Application	NetworkLayer	= "L3_APPLICATION_NETWORKING"
)

type DeterminismSource string

const (
	DeterminismNone				DeterminismSource	= ""
	DeterminismCommittedState		DeterminismSource	= "COMMITTED_STATE"
	DeterminismDeterministicProof		DeterminismSource	= "DETERMINISTIC_PROOF"
	DeterminismReplaySafeMessageID		DeterminismSource	= "REPLAY_SAFE_MESSAGE_ID"
	DeterminismSignedDiscoveryRecord	DeterminismSource	= "SIGNED_DISCOVERY_RECORD"
	DeterminismAdvisoryPeerMetric		DeterminismSource	= "ADVISORY_PEER_METRIC"
	DeterminismExternalNetworkCall		DeterminismSource	= "EXTERNAL_NETWORK_CALL"
)

type LayerSpec struct {
	Layer			NetworkLayer
	Extends			NetworkLayer
	TransportBaseline	string
	ConsensusCritical	bool
	Channels		[]ChannelClass
}

type NetworkMessage struct {
	Layer			NetworkLayer
	Channel			ChannelClass
	ConsensusEffect		bool
	DeterminismSource	DeterminismSource
	ReplaySafeID		string
	PayloadHash		string
	PayloadSizeBytes	uint64
	Chunked			bool
	CommitmentVerified	bool
	CommittedProofHash	string
	UsesLivePeerMetrics	bool
	UsesExternalNetworkCall	bool
}

type DiscoveryRecord struct {
	RecordID		string
	RecordType		DRTObjectType
	OwnerNodeID		string
	TargetID		string
	AdvertisementHash	string
	ZoneID			string
	ServiceID		string
	OverlayID		string
	ExpiresHeight		uint64
	Signature		[]byte
	ProofHash		string
	ProofHeight		uint64
	Record			NodeRecord
}

type PeerScoreUse struct {
	Metrics			PeerMetrics
	Score			PeerScore
	Committed		bool
	UsedForConsensus	bool
}

type RoutingDecisionUse struct {
	UsedForConsensus		bool
	DerivedFromCommittedState	bool
	DeterministicProofHash		string
}

type StateTransitionNetworkAccess struct {
	InStateTransition	bool
	ExternalCalls		[]string
}

func DefaultLayerStack() []LayerSpec {
	return []LayerSpec{
		{
			Layer:			LayerL0Physical,
			TransportBaseline:	BaseTransportCometBFTP2P,
			ConsensusCritical:	true,
			Channels:		[]ChannelClass{ChannelConsensus, ChannelMempool, ChannelBlock, ChannelStateSync},
		},
		{
			Layer:			LayerL1Session,
			Extends:		LayerL0Physical,
			TransportBaseline:	BaseTransportCometBFTP2P,
			Channels:		[]ChannelClass{ChannelDiscovery, ChannelRouting},
		},
		{
			Layer:			LayerL2Overlay,
			Extends:		LayerL1Session,
			TransportBaseline:	BaseTransportCometBFTP2P,
			Channels:		[]ChannelClass{ChannelRouting, ChannelExecution, ChannelData},
		},
		{
			Layer:			LayerL3Application,
			Extends:		LayerL2Overlay,
			TransportBaseline:	BaseTransportCometBFTP2P,
			Channels:		[]ChannelClass{ChannelExecution, ChannelService, ChannelData, ChannelDiscovery},
		},
	}
}

func ValidateLayerStack(stack []LayerSpec) error {
	if len(stack) != 4 {
		return errors.New("networking layer stack must define exactly L0-L3")
	}
	expected := []struct {
		layer	NetworkLayer
		extends	NetworkLayer
	}{
		{LayerL0Physical, ""},
		{LayerL1Session, LayerL0Physical},
		{LayerL2Overlay, LayerL1Session},
		{LayerL3Application, LayerL2Overlay},
	}
	for i, spec := range stack {
		if spec.Layer != expected[i].layer {
			return fmt.Errorf("networking layer %d must be %s", i, expected[i].layer)
		}
		if spec.Extends != expected[i].extends {
			return fmt.Errorf("networking layer %s must extend %s", spec.Layer, expected[i].extends)
		}
		if strings.TrimSpace(spec.TransportBaseline) != BaseTransportCometBFTP2P {
			return errors.New("networking layer stack must preserve CometBFT P2P as baseline transport")
		}
		if len(spec.Channels) == 0 {
			return fmt.Errorf("networking layer %s requires channels", spec.Layer)
		}
		if err := validateChannels(spec.Channels); err != nil {
			return err
		}
		if spec.Layer == LayerL0Physical && !spec.ConsensusCritical {
			return errors.New("networking L0 CometBFT transport must remain consensus-critical")
		}
		if spec.Layer != LayerL0Physical && spec.ConsensusCritical {
			return errors.New("networking upper layers must not replace consensus-critical L0 transport")
		}
	}
	return nil
}

func NewNetworkMessage(msg NetworkMessage) (NetworkMessage, error) {
	msg = msg.Normalize()
	if msg.ReplaySafeID == "" {
		msg.ReplaySafeID = ComputeNetworkMessageID(msg)
	}
	if err := msg.ValidateHardRules(); err != nil {
		return NetworkMessage{}, err
	}
	return msg, nil
}

func (m NetworkMessage) Normalize() NetworkMessage {
	m.ReplaySafeID = strings.ToLower(strings.TrimSpace(m.ReplaySafeID))
	m.PayloadHash = strings.ToLower(strings.TrimSpace(m.PayloadHash))
	m.CommittedProofHash = strings.ToLower(strings.TrimSpace(m.CommittedProofHash))
	return m
}

func (m NetworkMessage) ValidateHardRules() error {
	msg := m.Normalize()
	if !IsNetworkLayer(msg.Layer) {
		return fmt.Errorf("unknown networking layer %q", msg.Layer)
	}
	if !IsChannelClass(msg.Channel) {
		return fmt.Errorf("unknown networking message channel %q", msg.Channel)
	}
	if err := ValidateHash("networking message payload hash", msg.PayloadHash); err != nil {
		return err
	}
	if msg.PayloadSizeBytes == 0 {
		return errors.New("networking message payload size must be positive")
	}
	if msg.PayloadSizeBytes > LargePayloadBytes && (!msg.Chunked || !msg.CommitmentVerified) {
		return errors.New("networking large payloads must be chunked and commitment-verified")
	}
	if msg.UsesExternalNetworkCall {
		return errors.New("networking message must not embed external network calls")
	}
	if msg.UsesLivePeerMetrics && msg.ConsensusEffect {
		return errors.New("networking live peer metrics are advisory until committed")
	}
	if !msg.ConsensusEffect {
		return nil
	}
	if err := ValidateHash("networking message replay-safe id", msg.ReplaySafeID); err != nil {
		return err
	}
	if !IsConsensusSafeDeterminismSource(msg.DeterminismSource) {
		return fmt.Errorf("networking consensus-effect message requires deterministic committed source, got %q", msg.DeterminismSource)
	}
	if msg.DeterminismSource == DeterminismDeterministicProof {
		if err := ValidateHash("networking committed proof hash", msg.CommittedProofHash); err != nil {
			return err
		}
	}
	if msg.Layer == LayerL3Application && msg.Channel == ChannelService && msg.CommittedProofHash == "" {
		return errors.New("networking L3 service traffic cannot affect consensus without committed proof")
	}
	return nil
}

func ComputeNetworkMessageID(msg NetworkMessage) string {
	msg = msg.Normalize()
	return HashParts(
		"network-message",
		string(msg.Layer),
		string(msg.Channel),
		fmt.Sprintf("%t", msg.ConsensusEffect),
		string(msg.DeterminismSource),
		msg.PayloadHash,
		fmt.Sprintf("%d", msg.PayloadSizeBytes),
		msg.CommittedProofHash,
	)
}

func (d DiscoveryRecord) Validate(networkSalt []byte, currentHeight uint64) error {
	if IsObjectDiscoveryRecord(d) {
		return ValidateSignedDiscoveryRecord(d, networkSalt, currentHeight)
	}
	if err := d.Record.Validate(networkSalt, currentHeight); err != nil {
		return err
	}
	if d.ProofHash == "" {
		return nil
	}
	if err := ValidateHash("networking discovery proof hash", strings.ToLower(strings.TrimSpace(d.ProofHash))); err != nil {
		return err
	}
	if d.ProofHeight == 0 {
		return errors.New("networking discovery proof height must be positive")
	}
	if d.ProofHeight > d.Record.ExpiresHeight {
		return errors.New("networking discovery proof cannot outlive node record")
	}
	if currentHeight > 0 && d.ProofHeight > currentHeight {
		return errors.New("networking discovery proof height is in the future")
	}
	return nil
}

func ValidatePeerScoreUse(use PeerScoreUse) error {
	score, err := ComputePeerScore(use.Metrics)
	if err != nil {
		return err
	}
	if use.Score != (PeerScore{}) && use.Score != score {
		return errors.New("networking peer score does not match metrics")
	}
	if use.UsedForConsensus && !use.Committed {
		return errors.New("networking peer scoring is advisory until committed")
	}
	return nil
}

func ValidateRoutingDecisionUse(use RoutingDecisionUse) error {
	if !use.UsedForConsensus {
		return nil
	}
	if use.DerivedFromCommittedState {
		return nil
	}
	if err := ValidateHash("networking routing deterministic proof", strings.ToLower(strings.TrimSpace(use.DeterministicProofHash))); err != nil {
		return err
	}
	return nil
}

func ValidateStateTransitionNetworkAccess(access StateTransitionNetworkAccess) error {
	if !access.InStateTransition {
		return nil
	}
	if len(access.ExternalCalls) == 0 {
		return nil
	}
	return errors.New("networking external network calls are forbidden inside state transition execution")
}

func IsNetworkLayer(layer NetworkLayer) bool {
	switch layer {
	case LayerL0Physical, LayerL1Session, LayerL2Overlay, LayerL3Application:
		return true
	default:
		return false
	}
}

func IsDeterminismSource(source DeterminismSource) bool {
	switch source {
	case DeterminismNone,
		DeterminismCommittedState,
		DeterminismDeterministicProof,
		DeterminismReplaySafeMessageID,
		DeterminismSignedDiscoveryRecord,
		DeterminismAdvisoryPeerMetric,
		DeterminismExternalNetworkCall:
		return true
	default:
		return false
	}
}

func IsConsensusSafeDeterminismSource(source DeterminismSource) bool {
	switch source {
	case DeterminismCommittedState,
		DeterminismDeterministicProof,
		DeterminismReplaySafeMessageID,
		DeterminismSignedDiscoveryRecord:
		return true
	default:
		return false
	}
}

func cloneLayerStack(stack []LayerSpec) []LayerSpec {
	out := make([]LayerSpec, len(stack))
	for i, spec := range stack {
		spec.Channels = append([]ChannelClass(nil), spec.Channels...)
		out[i] = spec
	}
	return out
}
