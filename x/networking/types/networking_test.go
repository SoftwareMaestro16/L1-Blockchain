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

func TestNodeIdentityDerivesFromKeysAndBindsRoles(t *testing.T) {
	salt := []byte("aetheris-test-network")
	privateKey := deterministicPrivateKey(0x12)
	leftAddressHash, err := HashNetworkAddresses([]string{"tcp://10.0.0.1:26656"})
	require.NoError(t, err)
	rightAddressHash, err := HashNetworkAddresses([]string{"tcp://10.0.0.2:26656"})
	require.NoError(t, err)

	left, err := SignNodeRecord(NodeRecord{
		Roles:                []NodeRole{NodeRoleFull},
		NetworkAddressesHash: leftAddressHash,
		ProtocolVersions:     []string{DefaultProtocolVersion},
		ExpiresHeight:        100,
	}, privateKey, salt)
	require.NoError(t, err)
	right, err := SignNodeRecord(NodeRecord{
		Roles:                []NodeRole{NodeRoleFull},
		NetworkAddressesHash: rightAddressHash,
		ProtocolVersions:     []string{DefaultProtocolVersion},
		ExpiresHeight:        100,
	}, privateKey, salt)
	require.NoError(t, err)

	require.Equal(t, ComputeNodeID(left.NodePubKey, salt), left.NodeID)
	require.Equal(t, left.NodeID, right.NodeID)
	require.NotEqual(t, left.NetworkAddressesHash, right.NetworkAddressesHash)

	tampered := left
	tampered.Roles = []NodeRole{NodeRoleFull, NodeRoleService}
	require.ErrorContains(t, tampered.Validate(salt, 10), "signature")

	validatorKey := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{0x13}, ed25519.SeedSize)).Public().(ed25519.PublicKey)
	validator, err := SignNodeRecord(NodeRecord{
		ValidatorPubKey:      validatorKey,
		Roles:                []NodeRole{NodeRoleValidator, NodeRoleFull},
		NetworkAddressesHash: leftAddressHash,
		ProtocolVersions:     []string{DefaultProtocolVersion},
		ExpiresHeight:        100,
	}, privateKey, salt)
	require.NoError(t, err)
	require.Equal(t, ComputeNodeID(validatorKey, salt), validator.NodeID)
	require.NotEqual(t, ComputeNodeID(validator.NodePubKey, salt), validator.NodeID)
}

func TestSignedIdentityTransitionRotatesNodeIdentity(t *testing.T) {
	salt := []byte("aetheris-test-network")
	oldPrivateKey := deterministicPrivateKey(0x91)
	newPrivateKey := deterministicPrivateKey(0x92)
	oldRecord := signedNodeRecord(t, 0x91, salt, 100, NodeRoleService)
	newRecord := signedNodeRecord(t, 0x92, salt, 100, NodeRoleService)
	remote := signedNodeRecord(t, 0x93, salt, 100, NodeRoleFull)

	transition, err := SignIdentityTransition(oldRecord, newRecord, oldPrivateKey, newPrivateKey, salt, 20, 80, []byte("identity-rotation"))
	require.NoError(t, err)
	require.NoError(t, ValidateIdentityTransition(oldRecord, newRecord, transition, salt, 20))
	require.NotEmpty(t, transition.TransitionID)

	tampered := NormalizeIdentityTransition(transition)
	tampered.NewSignature[0] ^= 0xff
	require.ErrorContains(t, ValidateIdentityTransition(oldRecord, newRecord, tampered, salt, 20), "new signature")

	tampered = NormalizeIdentityTransition(transition)
	tampered.ToRoles = []NodeRole{NodeRoleService, NodeRoleRouting}
	tampered.TransitionID = ComputeIdentityTransitionID(tampered)
	payload, err := tampered.SigningPayload()
	require.NoError(t, err)
	tampered.OldSignature = ed25519.Sign(oldPrivateKey, payload)
	tampered.NewSignature = ed25519.Sign(newPrivateKey, payload)
	require.ErrorContains(t, ValidateIdentityTransition(oldRecord, newRecord, tampered, salt, 20), "roles")

	state := EmptyState()
	state, err = RegisterNodeRecord(state, oldRecord, salt, 10)
	require.NoError(t, err)
	state, err = RegisterNodeRecord(state, remote, salt, 10)
	require.NoError(t, err)
	session, err := NegotiateSession(oldRecord, remote, testSessionRequest(oldRecord, remote, 11, 50, "rotation-session", nil))
	require.NoError(t, err)
	state, err = OpenSession(state, session, 12)
	require.NoError(t, err)
	state, err = RegisterRoleCommitment(state, RoleCommitment{
		NodeID:         oldRecord.NodeID,
		Role:           NodeRoleService,
		Bonded:         true,
		CommitmentHash: HashParts("old-service-commitment"),
		ExpiresHeight:  70,
	}, 13)
	require.NoError(t, err)

	state, err = ApplyIdentityTransition(state, transition, newRecord, salt, 20)
	require.NoError(t, err)
	require.False(t, containsNode(state.NodeRecords, oldRecord.NodeID))
	require.True(t, containsNode(state.NodeRecords, newRecord.NodeID))
	require.Empty(t, state.Sessions)
	require.Empty(t, state.RoleCommitments)
	require.Equal(t, []IdentityTransitionRecord{transition}, state.IdentityTransitions)
	require.NoError(t, state.Validate())
}

func TestNetworkAddressHashCanonicalizesOffchainAddresses(t *testing.T) {
	addresses := []string{
		"tcp://10.0.0.2:26656",
		"tcp://10.0.0.1:26656",
		"tcp://10.0.0.1:26656",
	}
	left, err := HashNetworkAddresses(addresses)
	require.NoError(t, err)
	right, err := HashNetworkAddresses([]string{
		"tcp://10.0.0.1:26656",
		"tcp://10.0.0.2:26656",
	})
	require.NoError(t, err)

	require.Equal(t, left, right)

	salt := []byte("aetheris-test-network")
	record := signedNodeRecord(t, 0x19, salt, 100, NodeRoleFull)
	record.NetworkAddressesHash = left
	payload, err := record.SigningPayload()
	require.NoError(t, err)
	record.Signature = ed25519.Sign(deterministicPrivateKey(0x19), payload)
	require.NoError(t, VerifyNodeRecordAddresses(record, addresses))
	require.ErrorContains(t, VerifyNodeRecordAddresses(record, []string{"tcp://10.0.0.3:26656"}), "address list")
}

func TestSessionNegotiationCreatesDeterministicStreams(t *testing.T) {
	salt := []byte("aetheris-test-network")
	local := signedNodeRecord(t, 0x21, salt, 100, NodeRoleFull)
	remote := signedNodeRecord(t, 0x22, salt, 100, NodeRoleService)

	req := testSessionRequest(local, remote, 10, 50, "session-nonce", []ChannelClass{ChannelService, ChannelConsensus, ChannelData})
	req.QOSPolicy = QoSPolicyConsensusFirst
	session, err := NegotiateVerifiedSession(local, remote, req, salt, 10)
	require.NoError(t, err)
	require.NoError(t, session.Validate())
	require.Equal(t, ChannelConsensus, session.Streams[0].Channel)
	require.Equal(t, ChannelData, session.Streams[len(session.Streams)-1].Channel)
	require.Equal(t, session.CipherSuite, session.SessionKeys.CipherSuite)
	require.Equal(t, session.OpenedHeight, session.SessionKeys.EstablishedHeight)
	require.NotEmpty(t, session.SessionKeys.TranscriptHash)

	again, err := NegotiateVerifiedSession(local, remote, req, salt, 10)
	require.NoError(t, err)
	require.Equal(t, session, again)

	req.ProtocolVersions = []string{"unsupported"}
	_, err = NegotiateVerifiedSession(local, remote, req, salt, 10)
	require.ErrorContains(t, err, "protocol")

	req = testSessionRequest(local, remote, 10, 101, "expired-session", nil)
	_, err = NegotiateVerifiedSession(local, remote, req, salt, 101)
	require.ErrorContains(t, err, "expired")
}

func TestMultiplexedStreamsEnforceEncryptionCapacityAndResetPolicy(t *testing.T) {
	salt := []byte("aetheris-test-network")
	local := signedNodeRecord(t, 0x23, salt, 100, NodeRoleFull)
	remote := signedNodeRecord(t, 0x24, salt, 100, NodeRoleService)
	session, err := NegotiateVerifiedSession(local, remote, testSessionRequest(local, remote, 10, 50, "multiplexed-streams", nil), salt, 10)
	require.NoError(t, err)

	var consensus StreamSpec
	var service StreamSpec
	var bulk StreamSpec
	for _, stream := range session.Streams {
		require.NotEmpty(t, stream.EncryptionContext)
		switch stream.Channel {
		case ChannelConsensus:
			consensus = stream
		case ChannelService:
			service = stream
		case ChannelData:
			bulk = stream
		}
	}
	require.GreaterOrEqual(t, consensus.FlowControlWindow, uint64(DefaultFlowWindowBytes))
	require.Greater(t, service.Priority, consensus.Priority)
	require.GreaterOrEqual(t, bulk.FlowControlWindow, bulk.MaxMessageBytes)

	broken := append([]StreamSpec(nil), session.Streams...)
	for i := range broken {
		if broken[i].Channel == ChannelService {
			broken[i].Priority = consensus.Priority
		}
	}
	require.ErrorContains(t, ValidateStreamSet(broken, DefaultQoSClassPolicies()), "outrank consensus")

	decision, err := ResetStream(session, service.StreamID, StreamResetKeepSession)
	require.NoError(t, err)
	require.False(t, decision.SessionClosed)
	require.Len(t, decision.RemainingStreams, len(session.Streams)-1)

	decision, err = ResetStream(session, service.StreamID, StreamResetCloseSession)
	require.NoError(t, err)
	require.True(t, decision.SessionClosed)
}

func TestQoSClassPoliciesForbidConsensusInversionAndDowngradeServiceOnly(t *testing.T) {
	policies := DefaultQoSClassPolicies()
	require.NoError(t, ValidateQoSClassPolicies(policies))

	broken := append([]QoSClassPolicy(nil), policies...)
	for i := range broken {
		if broken[i].Class == QoSClassCriticalConsensus {
			broken[i].Priority = PriorityForChannel(ChannelData)
		}
	}
	require.ErrorContains(t, ValidateQoSClassPolicies(broken), "priority inversion")

	broken = append([]QoSClassPolicy(nil), policies...)
	for i := range broken {
		if broken[i].Class == QoSClassBulkData {
			broken[i].Backpressure = false
		}
	}
	require.ErrorContains(t, ValidateQoSClassPolicies(broken), "backpressure")

	decision := EvaluatePeerServiceQuota(2<<20, 1<<20)
	require.True(t, decision.DowngradeServiceTraffic)
	require.False(t, decision.DisconnectConsensus)
	require.Equal(t, QoSClassServiceCall, decision.Class)
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

func TestL0ChannelIDsAreStableAndRoundTrip(t *testing.T) {
	expected := map[ChannelClass]ChannelID{
		ChannelConsensus: ChannelIDConsensus,
		ChannelMempool:   ChannelIDMempool,
		ChannelBlock:     ChannelIDBlock,
		ChannelStateSync: ChannelIDStateSync,
		ChannelData:      ChannelIDData,
		ChannelExecution: ChannelIDExecution,
		ChannelService:   ChannelIDService,
		ChannelRouting:   ChannelIDRouting,
		ChannelDiscovery: ChannelIDDiscovery,
	}

	for channel, id := range expected {
		gotID, err := ChannelIDForClass(channel)
		require.NoError(t, err)
		require.Equal(t, id, gotID)

		gotChannel, err := ChannelClassForID(id)
		require.NoError(t, err)
		require.Equal(t, channel, gotChannel)
	}

	_, err := ChannelClassForID(ChannelID(0xff))
	require.ErrorContains(t, err, "unknown")
}

func TestL0PriorityPolicyMatchesChannelClassOrder(t *testing.T) {
	require.Less(t, PriorityForChannel(ChannelConsensus), PriorityForChannel(ChannelBlock))
	require.Less(t, PriorityForChannel(ChannelBlock), PriorityForChannel(ChannelStateSync))
	require.Less(t, PriorityForChannel(ChannelStateSync), PriorityForChannel(ChannelExecution))
	require.Less(t, PriorityForChannel(ChannelExecution), PriorityForChannel(ChannelMempool))
	require.Less(t, PriorityForChannel(ChannelMempool), PriorityForChannel(ChannelService))
	require.Equal(t, PriorityForChannel(ChannelService), PriorityForChannel(ChannelDiscovery))
	require.Equal(t, PriorityForChannel(ChannelService), PriorityForChannel(ChannelRouting))
	require.Less(t, PriorityForChannel(ChannelService), PriorityForChannel(ChannelData))
}

func TestL0BandwidthLedgerAccountsByChannel(t *testing.T) {
	adapter := DefaultAetherNetworkingAdapter()
	ledger, err := NewBandwidthLedger(42, adapter.Bandwidth, DefaultChannelPolicies())
	require.NoError(t, err)
	require.NoError(t, ledger.Validate())

	consensusAccount := bandwidthAccountForChannel(t, ledger, ChannelConsensus)
	reserve := adapter.Bandwidth.MaxOutboundBytesPerBlock * uint64(adapter.Bandwidth.ConsensusReserveBps) / uint64(BasisPoints)
	require.GreaterOrEqual(t, consensusAccount.LimitBytes, reserve)

	envelope := testEnvelope(ChannelService, 512, 42, 1, "service-bandwidth")
	nextLedger, err := AccountBandwidth(ledger, envelope)
	require.NoError(t, err)
	serviceAccount := bandwidthAccountForChannel(t, nextLedger, ChannelService)
	require.Equal(t, envelope.SizeBytes, serviceAccount.UsedBytes)
}

func TestL0ScheduleKeepsConsensusAheadOfServiceAndBulkTraffic(t *testing.T) {
	adapter := DefaultAetherNetworkingAdapter()
	adapter.Bandwidth.MaxOutboundBytesPerBlock = 8 << 20
	adapter.Bandwidth.ConsensusReserveBps = 5_000

	envelopes := []TransportEnvelope{
		testEnvelope(ChannelService, DefaultMaxMessageBytes, 10, 1, "service-1"),
		testEnvelope(ChannelData, MaxStreamMessageBytes, 10, 2, "bulk-data"),
		testEnvelope(ChannelDiscovery, 16<<10, 10, 3, "discovery-1"),
		testEnvelope(ChannelConsensus, 256, 99, 99, "consensus-vote"),
		testEnvelope(ChannelExecution, 512, 10, 4, "execution-receipt"),
	}

	schedule, err := ScheduleL0Propagation(adapter, envelopes, 20, PeerScore{ScoreBps: BasisPoints}, 100)
	require.NoError(t, err)
	require.NoError(t, schedule.Validate())
	require.NotEmpty(t, schedule.Plans)
	require.Equal(t, ChannelConsensus, schedule.Plans[0].Envelope.Channel)
	require.True(t, schedule.Plans[0].HandledByCometBFT)

	for _, dropped := range schedule.Dropped {
		require.NotEqual(t, ChannelConsensus, dropped.Channel)
	}
	consensusMetrics := l0MetricsForChannel(t, schedule.Metrics, ChannelConsensus)
	require.Equal(t, uint64(1), consensusMetrics.SentCount)
	require.Zero(t, consensusMetrics.DroppedCount)
	require.Zero(t, consensusMetrics.ConsensusDelayBlocks)
	require.NotContains(t, l0AlertCodes(schedule.Alerts), "CONSENSUS_TRAFFIC_DROPPED")
}

func TestL0AlertsEscalateConsensusDropsAboveBackpressure(t *testing.T) {
	alerts := EvaluateL0Alerts([]L0ChannelMetrics{
		{Channel: ChannelService, ChannelID: ChannelIDService, DroppedCount: 1},
		{Channel: ChannelConsensus, ChannelID: ChannelIDConsensus, DroppedCount: 1},
	})

	require.Len(t, alerts, 2)
	require.Equal(t, L0AlertCritical, alerts[0].Severity)
	require.Equal(t, "CONSENSUS_TRAFFIC_DROPPED", alerts[0].Code)
	require.Equal(t, L0AlertWarning, alerts[1].Severity)
	require.Equal(t, "NON_CONSENSUS_BACKPRESSURE", alerts[1].Code)
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

	session, err := NegotiateSession(local, remote, testSessionRequest(local, remote, 11, 50, "nonce", nil))
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

func testEnvelope(channel ChannelClass, sizeBytes uint64, enqueuedHeight uint64, sequence uint64, label string) TransportEnvelope {
	return TransportEnvelope{
		Channel:        channel,
		SizeBytes:      sizeBytes,
		EnqueuedHeight: enqueuedHeight,
		Sequence:       sequence,
		PayloadHash:    HashParts(label),
	}
}

func bandwidthAccountForChannel(t *testing.T, ledger BandwidthLedger, channel ChannelClass) BandwidthAccount {
	t.Helper()

	for _, account := range ledger.Accounts {
		if account.Channel == channel {
			return account
		}
	}
	t.Fatalf("missing bandwidth account for %s", channel)
	return BandwidthAccount{}
}

func l0MetricsForChannel(t *testing.T, metrics []L0ChannelMetrics, channel ChannelClass) L0ChannelMetrics {
	t.Helper()

	for _, metric := range metrics {
		if metric.Channel == channel {
			return metric
		}
	}
	t.Fatalf("missing L0 metrics for %s", channel)
	return L0ChannelMetrics{}
}

func l0AlertCodes(alerts []L0Alert) []string {
	codes := make([]string, len(alerts))
	for i, alert := range alerts {
		codes[i] = alert.Code
	}
	return codes
}

func testSessionRequest(local, remote NodeRecord, openedHeight, expiresHeight uint64, nonce string, channels []ChannelClass) SessionRequest {
	return SessionRequest{
		LocalNodeID:                 local.NodeID,
		RemoteNodeID:                remote.NodeID,
		ProtocolVersions:            []string{DefaultProtocolVersion},
		ChannelClasses:              channels,
		LocalEphemeralPubKey:        bytes.Repeat([]byte{0xa1}, SessionEphemeralKeyBytes),
		RemoteEphemeralPubKey:       bytes.Repeat([]byte{0xb2}, SessionEphemeralKeyBytes),
		SessionSecretCommitmentHash: HashParts("session-secret", local.NodeID, remote.NodeID, nonce),
		OpenedHeight:                openedHeight,
		ExpiresHeight:               expiresHeight,
		Nonce:                       []byte(nonce),
	}
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
