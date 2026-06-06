package types

import (
	"bytes"
	"crypto/ed25519"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNodeRecordSignatureIdentityAndExpiry(t *testing.T) {
	salt := []byte("aetheris-test-network")
	record := signedNodeRecord(t, 0x11, salt, 100, NodeRoleFull, NodeRoleRouting)

	require.NoError(t, record.Validate(salt, 99))

	expired := record
	require.ErrorContains(t, expired.Validate(salt, 101), "expired")

	tampered := record
	tampered.Roles = append(tampered.Roles, NodeRoleStorageProvider)
	require.ErrorContains(t, tampered.Validate(salt, 99), "signature")

	wrongSalt := []byte("wrong-network")
	require.ErrorContains(t, record.Validate(wrongSalt, 99), "node id")
}

func TestNetworkAddressHashCanonicalizesOffchainAddresses(t *testing.T) {
	left, err := HashNetworkAddresses([]string{
		"tcp://10.0.0.2:26656",
		"tcp://10.0.0.1:26656",
		"tcp://10.0.0.1:26656",
	})
	require.NoError(t, err)
	right, err := HashNetworkAddresses([]string{
		"tcp://10.0.0.1:26656",
		"tcp://10.0.0.2:26656",
	})
	require.NoError(t, err)

	require.Equal(t, left, right)
}

func TestSessionNegotiationCreatesDeterministicStreams(t *testing.T) {
	salt := []byte("aetheris-test-network")
	local := signedNodeRecord(t, 0x21, salt, 100, NodeRoleFull)
	remote := signedNodeRecord(t, 0x22, salt, 100, NodeRoleService)

	req := SessionRequest{
		LocalNodeID:      local.NodeID,
		RemoteNodeID:     remote.NodeID,
		ProtocolVersions: []string{DefaultProtocolVersion},
		ChannelClasses:   []ChannelClass{ChannelService, ChannelConsensus, ChannelData},
		OpenedHeight:     10,
		ExpiresHeight:    50,
		Nonce:            []byte("session-nonce"),
		QOSPolicy:        QoSPolicyConsensusFirst,
	}
	session, err := NegotiateSession(local, remote, req)
	require.NoError(t, err)
	require.NoError(t, session.Validate())
	require.Equal(t, ChannelConsensus, session.Streams[0].Channel)
	require.Equal(t, ChannelData, session.Streams[len(session.Streams)-1].Channel)

	again, err := NegotiateSession(local, remote, req)
	require.NoError(t, err)
	require.Equal(t, session, again)

	req.ProtocolVersions = []string{"unsupported"}
	_, err = NegotiateSession(local, remote, req)
	require.ErrorContains(t, err, "protocol")
}

func TestConsensusEnvelopeOutranksServiceAndBulkData(t *testing.T) {
	policies := DefaultChannelPolicies()
	service := TransportEnvelope{
		Channel:        ChannelService,
		SizeBytes:      512,
		EnqueuedHeight: 1,
		Sequence:       1,
		PayloadHash:    HashParts("service"),
	}
	data := TransportEnvelope{
		Channel:        ChannelData,
		SizeBytes:      2 << 20,
		EnqueuedHeight: 1,
		Sequence:       2,
		PayloadHash:    HashParts("data"),
	}
	consensus := TransportEnvelope{
		Channel:        ChannelConsensus,
		SizeBytes:      128,
		EnqueuedHeight: 100,
		Sequence:       99,
		PayloadHash:    HashParts("consensus"),
	}

	next, found, err := SelectNextEnvelope([]TransportEnvelope{service, data, consensus}, policies)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, ChannelConsensus, next.Channel)
}

func TestChunkPayloadRoundTripAndCorruptionDetection(t *testing.T) {
	payload := bytes.Repeat([]byte("aetheris-networking"), 512)
	chunks, err := ChunkPayload(payload, 257)
	require.NoError(t, err)
	require.Greater(t, len(chunks), 1)

	reordered := append([]PayloadChunk(nil), chunks...)
	for i, j := 0, len(reordered)-1; i < j; i, j = i+1, j-1 {
		reordered[i], reordered[j] = reordered[j], reordered[i]
	}
	decoded, err := ReassemblePayload(reordered)
	require.NoError(t, err)
	require.Equal(t, payload, decoded)

	reordered[0].Bytes[0] ^= 0xff
	_, err = ReassemblePayload(reordered)
	require.ErrorContains(t, err, "chunk hash")
}

func TestNetworkingStateRegistersNodesAndSessionsCanonically(t *testing.T) {
	salt := []byte("aetheris-test-network")
	local := signedNodeRecord(t, 0x31, salt, 100, NodeRoleFull)
	remote := signedNodeRecord(t, 0x32, salt, 100, NodeRoleService)

	state := EmptyState()
	var err error
	state, err = RegisterNodeRecord(state, remote, salt, 10)
	require.NoError(t, err)
	state, err = RegisterNodeRecord(state, local, salt, 10)
	require.NoError(t, err)
	require.Len(t, state.NodeRecords, 2)
	require.Less(t, state.NodeRecords[0].NodeID, state.NodeRecords[1].NodeID)

	session, err := NegotiateSession(local, remote, SessionRequest{
		LocalNodeID:      local.NodeID,
		RemoteNodeID:     remote.NodeID,
		ProtocolVersions: []string{DefaultProtocolVersion},
		OpenedHeight:     11,
		ExpiresHeight:    50,
		Nonce:            []byte("nonce"),
	})
	require.NoError(t, err)
	state, err = OpenSession(state, session, 12)
	require.NoError(t, err)
	require.Len(t, state.Sessions, 1)

	_, err = OpenSession(state, session, 12)
	require.ErrorContains(t, err, "already exists")

	pruned, err := PruneExpired(state, 101)
	require.NoError(t, err)
	require.Empty(t, pruned.NodeRecords)
	require.Empty(t, pruned.Sessions)
}

func TestLayerStackPreservesCometBFTBaselineAndExtensionOrder(t *testing.T) {
	stack := DefaultLayerStack()
	require.NoError(t, ValidateLayerStack(stack))
	require.Equal(t, LayerL0Physical, stack[0].Layer)
	require.Equal(t, BaseTransportCometBFTP2P, stack[0].TransportBaseline)
	require.True(t, stack[0].ConsensusCritical)

	broken := cloneLayerStack(stack)
	broken[1].Extends = LayerL3Application
	require.ErrorContains(t, ValidateLayerStack(broken), "extend")

	broken = cloneLayerStack(stack)
	broken[0].TransportBaseline = "CUSTOM_P2P"
	require.ErrorContains(t, ValidateLayerStack(broken), "CometBFT")

	broken = cloneLayerStack(stack)
	broken[2].ConsensusCritical = true
	require.ErrorContains(t, ValidateLayerStack(broken), "upper layers")
}

func TestDefaultANAValidatesCometBFTBaselineAndResponsibilities(t *testing.T) {
	adapter := DefaultAetherNetworkingAdapter()

	require.NoError(t, ValidateAetherNetworkingAdapter(adapter))
	require.True(t, hasBaseCapability(adapter.BaseCapabilities, BaseCapabilityConsensusGossip))
	require.True(t, hasANAResponsibility(adapter.Responsibilities, ANAResponsibilityPeerScoring))
	require.Equal(t, BaseTransportCometBFTP2P, adapter.BaselineTransport)
	require.True(t, adapter.ValidateRoleAdvertisements)

	broken := cloneAdapter(adapter)
	broken.BaselineTransport = "CUSTOM_P2P"
	require.ErrorContains(t, ValidateAetherNetworkingAdapter(broken), "CometBFT")

	broken = cloneAdapter(adapter)
	broken.BaseCapabilities = broken.BaseCapabilities[1:]
	require.ErrorContains(t, ValidateAetherNetworkingAdapter(broken), "missing capability")

	broken = cloneAdapter(adapter)
	broken.Responsibilities = broken.Responsibilities[1:]
	require.ErrorContains(t, ValidateAetherNetworkingAdapter(broken), "missing responsibility")
}

func TestANAGuardrailsRejectConsensusReplacementAndCommittedPeerMetrics(t *testing.T) {
	adapter := DefaultAetherNetworkingAdapter()

	broken := cloneAdapter(adapter)
	broken.ChangesConsensusValidity = true
	require.ErrorContains(t, ValidateAetherNetworkingAdapter(broken), "validity")

	broken = cloneAdapter(adapter)
	broken.HidesConsensusMessages = true
	require.ErrorContains(t, ValidateAetherNetworkingAdapter(broken), "hide")

	broken = cloneAdapter(adapter)
	broken.ReplacesCometBFTConsensusGossip = true
	require.ErrorContains(t, ValidateAetherNetworkingAdapter(broken), "replace")

	broken = cloneAdapter(adapter)
	broken.PeerMetricsAffectCommittedState = true
	require.ErrorContains(t, ValidateAetherNetworkingAdapter(broken), "advisory")
}

func TestANAPropagationKeepsConsensusOnCometBFTAndFanoutForService(t *testing.T) {
	adapter := DefaultAetherNetworkingAdapter()
	consensus := TransportEnvelope{
		Channel:        ChannelConsensus,
		SizeBytes:      128,
		EnqueuedHeight: 100,
		Sequence:       1,
		PayloadHash:    HashParts("consensus-vote"),
	}
	plan, err := PlanPropagation(adapter, consensus, 20, PeerScore{ScoreBps: BasisPoints})
	require.NoError(t, err)
	require.True(t, plan.HandledByCometBFT)
	require.Zero(t, plan.AdapterFanout)
	require.False(t, plan.UsesAdvisoryPeerMetric)

	service := TransportEnvelope{
		Channel:        ChannelService,
		SizeBytes:      512,
		EnqueuedHeight: 100,
		Sequence:       2,
		PayloadHash:    HashParts("service-message"),
	}
	plan, err = PlanPropagation(adapter, service, 20, PeerScore{ScoreBps: 9_000})
	require.NoError(t, err)
	require.False(t, plan.HandledByCometBFT)
	require.True(t, plan.UsesAdvisoryPeerMetric)
	require.GreaterOrEqual(t, plan.AdapterFanout, adapter.Fanout.MinFanout)
	require.LessOrEqual(t, plan.AdapterFanout, adapter.Fanout.MaxFanout)
	require.Less(t, plan.Priority, uint32(100))
}

func TestANAMultiplexingDoesNotLetServiceOutrankConsensus(t *testing.T) {
	adapter := DefaultAetherNetworkingAdapter()
	for i := range adapter.ChannelBindings {
		if adapter.ChannelBindings[i].Channel == ChannelService {
			adapter.ChannelBindings[i].Priority = 0
		}
	}

	require.ErrorContains(t, ValidateAetherNetworkingAdapter(adapter), "consensus priority")
}

func TestNetworkMessageHardRules(t *testing.T) {
	proofHash := HashParts("committed-service-proof")
	msg, err := NewNetworkMessage(NetworkMessage{
		Layer:              LayerL3Application,
		Channel:            ChannelService,
		ConsensusEffect:    true,
		DeterminismSource:  DeterminismDeterministicProof,
		PayloadHash:        HashParts("payload"),
		PayloadSizeBytes:   128,
		CommittedProofHash: proofHash,
	})
	require.NoError(t, err)
	require.NoError(t, msg.ValidateHardRules())
	require.NotEmpty(t, msg.ReplaySafeID)

	missingProof := msg
	missingProof.CommittedProofHash = ""
	missingProof.ReplaySafeID = ComputeNetworkMessageID(missingProof)
	require.ErrorContains(t, missingProof.ValidateHardRules(), "committed proof")

	advisoryMetric := msg
	advisoryMetric.DeterminismSource = DeterminismAdvisoryPeerMetric
	advisoryMetric.ReplaySafeID = ComputeNetworkMessageID(advisoryMetric)
	require.ErrorContains(t, advisoryMetric.ValidateHardRules(), "deterministic committed")

	externalCall := msg
	externalCall.UsesExternalNetworkCall = true
	require.ErrorContains(t, externalCall.ValidateHardRules(), "external network calls")
}

func TestLargeNetworkPayloadRequiresChunkedVerifiedCommitment(t *testing.T) {
	msg := NetworkMessage{
		Layer:            LayerL3Application,
		Channel:          ChannelData,
		PayloadHash:      HashParts("large-payload"),
		PayloadSizeBytes: LargePayloadBytes + 1,
	}

	require.ErrorContains(t, msg.ValidateHardRules(), "large payloads")

	msg.Chunked = true
	msg.CommitmentVerified = true
	require.NoError(t, msg.ValidateHardRules())
}

func TestDiscoveryRecordMustBeSignedExpiringAndProofChecked(t *testing.T) {
	salt := []byte("aetheris-test-network")
	record := signedNodeRecord(t, 0x51, salt, 100, NodeRoleStateSync)
	discovery := DiscoveryRecord{
		Record:      record,
		ProofHash:   HashParts("optional-discovery-proof"),
		ProofHeight: 90,
	}

	require.NoError(t, discovery.Validate(salt, 95))

	discovery.ProofHeight = 101
	require.ErrorContains(t, discovery.Validate(salt, 0), "outlive")

	discovery = DiscoveryRecord{Record: record}
	discovery.Record.Signature[0] ^= 0xff
	require.ErrorContains(t, discovery.Validate(salt, 95), "signature")
}

func TestAdvisorySignalsCannotDriveConsensusUntilCommitted(t *testing.T) {
	metrics := PeerMetrics{LatencyMillis: 25, ReliabilityBps: 9_900, ThroughputBytesPerSec: 32 << 20}
	score, err := ComputePeerScore(metrics)
	require.NoError(t, err)

	require.NoError(t, ValidatePeerScoreUse(PeerScoreUse{Metrics: metrics, Score: score}))
	require.ErrorContains(t, ValidatePeerScoreUse(PeerScoreUse{
		Metrics:          metrics,
		Score:            score,
		UsedForConsensus: true,
	}), "advisory")
	require.NoError(t, ValidatePeerScoreUse(PeerScoreUse{
		Metrics:          metrics,
		Score:            score,
		Committed:        true,
		UsedForConsensus: true,
	}))
}

func TestRoutingAndStateTransitionHardRules(t *testing.T) {
	require.ErrorContains(t, ValidateRoutingDecisionUse(RoutingDecisionUse{UsedForConsensus: true}), "lowercase hex")
	require.NoError(t, ValidateRoutingDecisionUse(RoutingDecisionUse{
		UsedForConsensus:          true,
		DerivedFromCommittedState: true,
	}))
	require.NoError(t, ValidateRoutingDecisionUse(RoutingDecisionUse{
		UsedForConsensus:       true,
		DeterministicProofHash: HashParts("routing-proof"),
	}))

	require.NoError(t, ValidateStateTransitionNetworkAccess(StateTransitionNetworkAccess{
		InStateTransition: true,
	}))
	require.ErrorContains(t, ValidateStateTransitionNetworkAccess(StateTransitionNetworkAccess{
		InStateTransition: true,
		ExternalCalls:     []string{"https://example.invalid"},
	}), "forbidden")
}

func TestNetworkRoleConsensusScopeRequiresBondedCommitment(t *testing.T) {
	salt := []byte("aetheris-test-network")
	record := signedNodeRecord(t, 0x61, salt, 100, NodeRoleService, NodeRoleRouting, NodeRoleStorageProvider)

	scopes, err := RoleScopes(record, nil, 10)
	require.NoError(t, err)
	require.Len(t, scopes, 3)
	for _, scope := range scopes {
		require.True(t, scope.Advertised)
		require.False(t, scope.Committed)
		require.False(t, scope.ConsensusCritical)
	}

	state := EmptyState()
	state, err = RegisterNodeRecord(state, record, salt, 10)
	require.NoError(t, err)
	state, err = RegisterRoleCommitment(state, RoleCommitment{
		NodeID:         record.NodeID,
		Role:           NodeRoleService,
		Bonded:         true,
		CommitmentHash: HashParts("service-role-commitment"),
		ExpiresHeight:  80,
	}, 20)
	require.NoError(t, err)

	scopes, err = RoleScopes(record, state.RoleCommitments, 20)
	require.NoError(t, err)
	var serviceScope RoleScope
	for _, scope := range scopes {
		if scope.Role == NodeRoleService {
			serviceScope = scope
			break
		}
	}
	require.True(t, serviceScope.Committed)
	require.True(t, serviceScope.ConsensusCritical)

	pruned, err := PruneExpired(state, 81)
	require.NoError(t, err)
	scopes, err = RoleScopes(record, pruned.RoleCommitments, 81)
	require.NoError(t, err)
	for _, scope := range scopes {
		require.False(t, scope.ConsensusCritical)
	}
}

func TestANAValidatesSignedPeerRoleAdvertisements(t *testing.T) {
	salt := []byte("aetheris-test-network")
	record := signedNodeRecord(t, 0x66, salt, 100, NodeRoleService)
	adapter := DefaultAetherNetworkingAdapter()
	discovery := DiscoveryRecord{Record: record}

	scopes, err := ValidatePeerRoleAdvertisement(adapter, discovery, nil, salt, 10)
	require.NoError(t, err)
	require.Len(t, scopes, 1)
	require.Equal(t, NodeRoleService, scopes[0].Role)
	require.False(t, scopes[0].ConsensusCritical)

	discovery.Record.Signature[0] ^= 0xff
	_, err = ValidatePeerRoleAdvertisement(adapter, discovery, nil, salt, 10)
	require.ErrorContains(t, err, "signature")
}

func TestValidatorRoleIsConsensusCriticalWithoutRoleCommitment(t *testing.T) {
	salt := []byte("aetheris-test-network")
	privateKey := deterministicPrivateKey(0x71)
	validatorKey := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{0x72}, ed25519.SeedSize)).Public().(ed25519.PublicKey)
	addressHash, err := HashNetworkAddresses([]string{"tcp://127.0.0.1:26656"})
	require.NoError(t, err)
	record, err := SignNodeRecord(NodeRecord{
		ValidatorPubKey:      validatorKey,
		Roles:                []NodeRole{NodeRoleValidator, NodeRoleFull},
		NetworkAddressesHash: addressHash,
		ProtocolVersions:     []string{DefaultProtocolVersion},
		ExpiresHeight:        100,
	}, privateKey, salt)
	require.NoError(t, err)

	scopes, err := RoleScopes(record, nil, 10)
	require.NoError(t, err)
	var validatorScope RoleScope
	for _, scope := range scopes {
		if scope.Role == NodeRoleValidator {
			validatorScope = scope
			break
		}
	}
	require.True(t, validatorScope.ConsensusCritical)

	_, err = RegisterRoleCommitment(EmptyState(), RoleCommitment{
		NodeID:         record.NodeID,
		Role:           NodeRoleValidator,
		Bonded:         true,
		CommitmentHash: HashParts("validator-role-commitment"),
		ExpiresHeight:  80,
	}, 10)
	require.ErrorContains(t, err, "validator role")
}

func TestRoleCommitmentRejectsUnbondedUnadvertisedAndOutlivingRecords(t *testing.T) {
	salt := []byte("aetheris-test-network")
	record := signedNodeRecord(t, 0x81, salt, 100, NodeRoleService)
	state := EmptyState()
	var err error
	state, err = RegisterNodeRecord(state, record, salt, 10)
	require.NoError(t, err)

	_, err = RegisterRoleCommitment(state, RoleCommitment{
		NodeID:         record.NodeID,
		Role:           NodeRoleService,
		CommitmentHash: HashParts("unbonded"),
		ExpiresHeight:  80,
	}, 10)
	require.ErrorContains(t, err, "bonded")

	_, err = RegisterRoleCommitment(state, RoleCommitment{
		NodeID:         record.NodeID,
		Role:           NodeRoleRouting,
		Bonded:         true,
		CommitmentHash: HashParts("not-advertised"),
		ExpiresHeight:  80,
	}, 10)
	require.ErrorContains(t, err, "advertised")

	_, err = RegisterRoleCommitment(state, RoleCommitment{
		NodeID:         record.NodeID,
		Role:           NodeRoleService,
		Bonded:         true,
		CommitmentHash: HashParts("outlive"),
		ExpiresHeight:  101,
	}, 10)
	require.ErrorContains(t, err, "outlive")
}

func signedNodeRecord(t *testing.T, seed byte, salt []byte, expiresHeight uint64, roles ...NodeRole) NodeRecord {
	t.Helper()

	privateKey := deterministicPrivateKey(seed)
	addressHash, err := HashNetworkAddresses([]string{"tcp://127.0.0.1:26656"})
	require.NoError(t, err)
	record, err := SignNodeRecord(NodeRecord{
		Roles:                roles,
		NetworkAddressesHash: addressHash,
		ZonesSupported:       []string{"APPLICATION_ZONE", "FINANCIAL_ZONE"},
		ServicesSupported:    []string{"state-sync", "execution-stream"},
		ProtocolVersions:     []string{DefaultProtocolVersion},
		ExpiresHeight:        expiresHeight,
	}, privateKey, salt)
	require.NoError(t, err)
	return record
}

func deterministicPrivateKey(fill byte) ed25519.PrivateKey {
	seed := bytes.Repeat([]byte{fill}, ed25519.SeedSize)
	return ed25519.NewKeyFromSeed(seed)
}
