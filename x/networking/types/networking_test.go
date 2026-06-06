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

func TestSessionHandshakeStateMachineRejectsReplaysAndExpiredRecords(t *testing.T) {
	salt := []byte("aetheris-test-network")
	local := signedNodeRecord(t, 0x25, salt, 100, NodeRoleFull)
	remote := signedNodeRecord(t, 0x26, salt, 100, NodeRoleService)
	req := testSessionRequest(local, remote, 20, 60, "state-machine-handshake", []ChannelClass{ChannelConsensus, ChannelService})

	state, err := RunSessionHandshake(local, remote, req, salt, 20, nil)
	require.NoError(t, err)
	require.Equal(t, HandshakePhaseEstablished, state.Phase)
	require.Equal(t, state.Session.SessionID, ComputeSessionID(req.Normalize(), state.CipherSuite, state.ProtocolVersions, state.ChannelClasses))
	require.NotEmpty(t, state.ReplayID)
	require.NotEmpty(t, state.SessionKeys.KeyID)
	require.Equal(t, state.Session.SessionKeys, state.SessionKeys)

	replayed, err := RunSessionHandshake(local, remote, req, salt, 20, []string{state.ReplayID})
	require.ErrorContains(t, err, "replayed")
	require.Equal(t, HandshakePhaseRejected, replayed.Phase)
	require.Contains(t, replayed.RejectReason, "replayed")

	expiredRemote := signedNodeRecord(t, 0x27, salt, 19, NodeRoleService)
	expiredReq := testSessionRequest(local, expiredRemote, 20, 60, "expired-handshake", nil)
	rejected, err := RunSessionHandshake(local, expiredRemote, expiredReq, salt, 20, nil)
	require.ErrorContains(t, err, "expired")
	require.Equal(t, HandshakePhaseRejected, rejected.Phase)
}

func TestSessionKeyRotationUpdatesKeysWithoutChangingSessionID(t *testing.T) {
	salt := []byte("aetheris-test-network")
	local := signedNodeRecord(t, 0x28, salt, 100, NodeRoleFull)
	remote := signedNodeRecord(t, 0x29, salt, 100, NodeRoleService)
	handshake, err := RunSessionHandshake(local, remote, testSessionRequest(local, remote, 20, 80, "rotate-session-keys", nil), salt, 20, nil)
	require.NoError(t, err)

	rotated, err := RotateSessionKeys(handshake.Session, SessionKeyRotationRequest{
		SessionID:                handshake.Session.SessionID,
		NewLocalEphemeralPubKey:  bytes.Repeat([]byte{0xc3}, SessionEphemeralKeyBytes),
		NewRemoteEphemeralPubKey: bytes.Repeat([]byte{0xd4}, SessionEphemeralKeyBytes),
		NewSecretCommitmentHash:  HashParts("rotated-session-secret", handshake.Session.SessionID),
		RotatedAtHeight:          40,
		ExpiresHeight:            80,
		Nonce:                    []byte("rotation-nonce"),
	})
	require.NoError(t, err)
	require.Equal(t, handshake.Session.SessionID, rotated.SessionID)
	require.NotEqual(t, handshake.Session.SessionKeys.KeyID, rotated.SessionKeys.KeyID)
	require.Equal(t, uint64(40), rotated.SessionKeys.EstablishedHeight)
	for _, stream := range rotated.Streams {
		require.Equal(t, streamEncryptionContext(rotated.SessionKeys.KeyID, stream.StreamID), stream.EncryptionContext)
	}

	_, err = RotateSessionKeys(handshake.Session, SessionKeyRotationRequest{
		SessionID:                handshake.Session.SessionID,
		NewLocalEphemeralPubKey:  bytes.Repeat([]byte{0xc3}, SessionEphemeralKeyBytes),
		NewRemoteEphemeralPubKey: bytes.Repeat([]byte{0xd4}, SessionEphemeralKeyBytes),
		NewSecretCommitmentHash:  HashParts("rotated-session-secret", handshake.Session.SessionID),
		RotatedAtHeight:          81,
		ExpiresHeight:            90,
		Nonce:                    []byte("late-rotation"),
	})
	require.ErrorContains(t, err, "outside session range")
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

func TestDefaultOverlayDescriptorsCoverL2OverlayTypes(t *testing.T) {
	descriptors := DefaultOverlayDescriptors()
	require.NoError(t, ValidateOverlayDescriptors(descriptors, 0))
	require.Len(t, descriptors, 8)

	seen := make(map[OverlayType]bool)
	for _, desc := range descriptors {
		require.Equal(t, ComputeOverlayID(desc), desc.OverlayID)
		require.NotZero(t, desc.MinPeers)
		require.GreaterOrEqual(t, desc.MaxPeers, desc.MinPeers)
		require.LessOrEqual(t, desc.Fanout, desc.MaxPeers)
		require.True(t, IsQoSClass(desc.QoSClass))
		seen[desc.OverlayType] = true
	}

	for _, overlayType := range []OverlayType{
		OverlayTypeValidator,
		OverlayTypeZone,
		OverlayTypeExecution,
		OverlayTypeData,
		OverlayTypeService,
		OverlayTypeDiscovery,
		OverlayTypeStorage,
		OverlayTypeRouting,
	} {
		require.True(t, seen[overlayType], overlayType)
	}
}

func TestOverlayDescriptorRejectsInvalidMembershipQoSAndFanout(t *testing.T) {
	_, err := NewOverlayDescriptor(OverlayDescriptor{
		OverlayType: OverlayTypeService,
		PolicyHash:  HashParts("bad-service-overlay-qos"),
		Membership:  OverlayMembershipServiceAdvertisement,
		Routing:     RoutingStrategyLowLatencyAdvisory,
		MinPeers:    2,
		MaxPeers:    8,
		Fanout:      2,
		QoSClass:    QoSClassBulkData,
		Version:     1,
	})
	require.ErrorContains(t, err, "qos")

	_, err = NewOverlayDescriptor(OverlayDescriptor{
		OverlayType: OverlayTypeService,
		PolicyHash:  HashParts("bad-service-overlay-membership"),
		Membership:  OverlayMembershipRoutingRole,
		Routing:     RoutingStrategyLowLatencyAdvisory,
		MinPeers:    2,
		MaxPeers:    8,
		Fanout:      2,
		QoSClass:    QoSClassServiceCall,
		Version:     1,
	})
	require.ErrorContains(t, err, "membership")

	_, err = NewOverlayDescriptor(OverlayDescriptor{
		OverlayType: OverlayTypeData,
		PolicyHash:  HashParts("bad-data-overlay-fanout"),
		Membership:  OverlayMembershipDataProvider,
		Routing:     RoutingStrategyKBucket,
		MinPeers:    2,
		MaxPeers:    4,
		Fanout:      5,
		QoSClass:    QoSClassBulkData,
		Version:     1,
	})
	require.ErrorContains(t, err, "fanout")

	_, err = NewOverlayDescriptor(OverlayDescriptor{
		OverlayType: OverlayTypeValidator,
		PolicyHash:  HashParts("bad-validator-overlay-routing"),
		Membership:  OverlayMembershipValidatorSet,
		Routing:     RoutingStrategyLowLatencyAdvisory,
		MinPeers:    4,
		MaxPeers:    32,
		Fanout:      4,
		QoSClass:    QoSClassCriticalConsensus,
		Version:     1,
	})
	require.ErrorContains(t, err, "advisory")
}

func TestOverlayMembershipMatchesNodeRolesAndCapabilities(t *testing.T) {
	salt := []byte("aetheris-test-network")
	multiRole := signedNodeRecord(t, 0x35, salt, 100, NodeRoleFull, NodeRoleZoneExecution, NodeRoleService, NodeRoleStorageProvider, NodeRoleRouting)

	for _, overlayType := range []OverlayType{
		OverlayTypeZone,
		OverlayTypeExecution,
		OverlayTypeData,
		OverlayTypeService,
		OverlayTypeDiscovery,
		OverlayTypeStorage,
		OverlayTypeRouting,
	} {
		matches, err := NodeSatisfiesOverlayMembership(multiRole, defaultOverlayByType(t, overlayType))
		require.NoError(t, err)
		require.True(t, matches, overlayType)
	}

	matches, err := NodeSatisfiesOverlayMembership(multiRole, defaultOverlayByType(t, OverlayTypeValidator))
	require.NoError(t, err)
	require.False(t, matches)

	validatorKey := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{0x36}, ed25519.SeedSize)).Public().(ed25519.PublicKey)
	validatorPrivateKey := deterministicPrivateKey(0x37)
	addressHash, err := HashNetworkAddresses([]string{"tcp://127.0.0.1:26656"})
	require.NoError(t, err)
	validator, err := SignNodeRecord(NodeRecord{
		ValidatorPubKey:      validatorKey,
		Roles:                []NodeRole{NodeRoleValidator, NodeRoleFull},
		NetworkAddressesHash: addressHash,
		ZonesSupported:       []string{"APPLICATION_ZONE"},
		ProtocolVersions:     []string{DefaultProtocolVersion},
		ExpiresHeight:        100,
	}, validatorPrivateKey, salt)
	require.NoError(t, err)

	matches, err = NodeSatisfiesOverlayMembership(validator, defaultOverlayByType(t, OverlayTypeValidator))
	require.NoError(t, err)
	require.True(t, matches)
}

func TestNetworkingStateRegistersAndPrunesOverlayDescriptors(t *testing.T) {
	state := EmptyState()
	desc, err := NewOverlayDescriptor(OverlayDescriptor{
		OverlayType:   OverlayTypeService,
		PolicyHash:    HashParts("temporary-service-overlay"),
		Membership:    OverlayMembershipServiceAdvertisement,
		Routing:       RoutingStrategyLowLatencyAdvisory,
		MinPeers:      2,
		MaxPeers:      16,
		Fanout:        4,
		QoSClass:      QoSClassServiceCall,
		ExpiresHeight: 30,
		Version:       2,
	})
	require.NoError(t, err)

	state, err = RegisterOverlayDescriptor(state, desc, 10)
	require.NoError(t, err)
	require.Contains(t, overlayIDs(state.OverlayDescriptors), desc.OverlayID)

	pruned, err := PruneExpired(state, 31)
	require.NoError(t, err)
	require.NotContains(t, overlayIDs(pruned.OverlayDescriptors), desc.OverlayID)
	require.NoError(t, pruned.Validate())
}

func TestOverlayRoutingFanoutClampsToEligiblePeers(t *testing.T) {
	desc := defaultOverlayByType(t, OverlayTypeService)

	fanout, err := PlanOverlayFanout(desc, 20)
	require.NoError(t, err)
	require.Equal(t, desc.Fanout, fanout)

	fanout, err = PlanOverlayFanout(desc, 4)
	require.NoError(t, err)
	require.Equal(t, uint32(4), fanout)

	_, err = PlanOverlayFanout(desc, 1)
	require.ErrorContains(t, err, "insufficient")
}

func TestOverlayMembershipProofAuthorizesServiceStakeAndSignedRecords(t *testing.T) {
	salt := []byte("aetheris-test-network")
	service := signedNodeRecord(t, 0x38, salt, 100, NodeRoleService)
	serviceDesc := defaultOverlayByType(t, OverlayTypeService)
	serviceProof := testOverlayMembershipProof(t, service, serviceDesc, MembershipProofServiceRegistration, OverlayMembershipModeServiceRegistry, 80)

	membership, err := AuthorizeOverlayMembership(service, serviceDesc, serviceProof, 20)
	require.NoError(t, err)
	require.Equal(t, service.NodeID, membership.NodeID)
	require.Equal(t, serviceDesc.OverlayID, membership.OverlayID)

	storage := signedNodeRecord(t, 0x39, salt, 100, NodeRoleStorageProvider)
	storageDesc := defaultOverlayByType(t, OverlayTypeStorage)
	stakeProof := testOverlayMembershipProof(t, storage, storageDesc, MembershipProofProviderStake, OverlayMembershipModeStakeBased, 80)
	_, err = AuthorizeOverlayMembership(storage, storageDesc, stakeProof, 20)
	require.NoError(t, err)

	routing := signedNodeRecord(t, 0x3a, salt, 100, NodeRoleRouting)
	routingDesc := defaultOverlayByType(t, OverlayTypeRouting)
	authProof := testOverlayMembershipProof(t, routing, routingDesc, MembershipProofSignedAuthorization, OverlayMembershipModeCryptographicAuth, 80)
	_, err = AuthorizeOverlayMembership(routing, routingDesc, authProof, 20)
	require.NoError(t, err)

	tampered := authProof
	tampered.Signature[0] ^= 0xff
	_, err = AuthorizeOverlayMembership(routing, routingDesc, tampered, 20)
	require.ErrorContains(t, err, "signature")
}

func TestOverlayRoutingPipelineClassifiesAndSelectsServiceProviders(t *testing.T) {
	salt := []byte("aetheris-test-network")
	source := signedNodeRecord(t, 0x3b, salt, 100, NodeRoleFull)
	left := signedNodeRecord(t, 0x3c, salt, 100, NodeRoleService)
	right := signedNodeRecord(t, 0x3d, salt, 100, NodeRoleService)
	left.ServicesSupported = []string{"state-sync"}
	right.ServicesSupported = []string{"execution-stream"}
	desc, err := NewOverlayDescriptor(OverlayDescriptor{
		OverlayType: OverlayTypeService,
		PolicyHash:  HashParts("service-provider-routing"),
		Membership:  OverlayMembershipServiceAdvertisement,
		Routing:     RoutingStrategyServiceProvider,
		MinPeers:    2,
		MaxPeers:    8,
		Fanout:      2,
		QoSClass:    QoSClassServiceCall,
		Version:     1,
	})
	require.NoError(t, err)
	proofs := []OverlayMembershipProof{
		testOverlayMembershipProof(t, left, desc, MembershipProofServiceRegistration, OverlayMembershipModeServiceRegistry, 80),
		testOverlayMembershipProof(t, right, desc, MembershipProofServiceRegistration, OverlayMembershipModeServiceRegistry, 80),
	}
	msg, err := NewNetworkMessage(NetworkMessage{
		Layer:            LayerL3Application,
		Channel:          ChannelService,
		PayloadHash:      HashParts("service-route-payload"),
		PayloadSizeBytes: 512,
	})
	require.NoError(t, err)

	plan, err := BuildOverlayRoute(OverlayRoutingRequest{
		Message:          msg,
		SourceNodeID:     source.NodeID,
		CandidatePeers:   []NodeRecord{left, right},
		MembershipProofs: proofs,
		Graph:            RoutingGraph{OverlayID: desc.OverlayID, Version: 1},
		Hint:             RouteHint{ServiceID: "execution-stream"},
		CurrentHeight:    20,
	}, []OverlayDescriptor{desc})
	require.NoError(t, err)
	require.Equal(t, OverlayTypeService, plan.OverlayType)
	require.Equal(t, RoutingStrategyServiceProvider, plan.Strategy)
	require.Equal(t, right.NodeID, plan.TargetNodeIDs[0])
}

func TestOverlayRoutingConsensusSafetyRequiresCommittedRoutingTable(t *testing.T) {
	salt := []byte("aetheris-test-network")
	source := signedNodeRecord(t, 0x3e, salt, 100, NodeRoleZoneExecution)
	slow := signedNodeRecord(t, 0x3f, salt, 100, NodeRoleZoneExecution)
	fast := signedNodeRecord(t, 0x40, salt, 100, NodeRoleZoneExecution)
	desc, err := NewOverlayDescriptor(OverlayDescriptor{
		OverlayType: OverlayTypeExecution,
		PolicyHash:  HashParts("committed-latency-routing"),
		Membership:  OverlayMembershipExecutionRole,
		Routing:     RoutingStrategyShortestLatencyPath,
		MinPeers:    2,
		MaxPeers:    8,
		Fanout:      2,
		QoSClass:    QoSClassExecutionMessage,
		Version:     1,
	})
	require.NoError(t, err)
	proofs := []OverlayMembershipProof{
		testOverlayMembershipProof(t, slow, desc, MembershipProofZoneAssignment, OverlayMembershipModeZoneAssignment, 80),
		testOverlayMembershipProof(t, fast, desc, MembershipProofZoneAssignment, OverlayMembershipModeZoneAssignment, 80),
	}
	msg, err := NewNetworkMessage(NetworkMessage{
		Layer:             LayerL2Overlay,
		Channel:           ChannelExecution,
		ConsensusEffect:   true,
		DeterminismSource: DeterminismCommittedState,
		PayloadHash:       HashParts("execution-route-payload"),
		PayloadSizeBytes:  512,
	})
	require.NoError(t, err)
	graph := RoutingGraph{
		OverlayID: desc.OverlayID,
		Version:   1,
		Edges: []RoutingEdge{
			{FromNodeID: source.NodeID, ToNodeID: slow.NodeID, LatencyMillis: 90, Priority: 2},
			{FromNodeID: source.NodeID, ToNodeID: fast.NodeID, LatencyMillis: 10, Priority: 1},
		},
	}

	_, err = BuildOverlayRoute(OverlayRoutingRequest{
		Message:          msg,
		SourceNodeID:     source.NodeID,
		CandidatePeers:   []NodeRecord{slow, fast},
		MembershipProofs: proofs,
		Graph:            graph,
		CurrentHeight:    20,
	}, []OverlayDescriptor{desc})
	require.ErrorContains(t, err, "committed routing")

	graph.Committed = true
	plan, err := BuildOverlayRoute(OverlayRoutingRequest{
		Message:          msg,
		SourceNodeID:     source.NodeID,
		CandidatePeers:   []NodeRecord{slow, fast},
		MembershipProofs: proofs,
		Graph:            graph,
		CurrentHeight:    20,
	}, []OverlayDescriptor{desc})
	require.NoError(t, err)
	require.True(t, plan.UsesCommittedRoutingTable)
	require.False(t, plan.UsesNodeLocalAdaptation)
	require.Equal(t, fast.NodeID, plan.TargetNodeIDs[0])
}

func TestAdaptiveOverlayGraphBuildsPeerSetsAndPreservesGlobalDiversity(t *testing.T) {
	salt := []byte("aetheris-test-network")
	local := signedNodeRecord(t, 0x43, salt, 100, NodeRoleService)
	desc := defaultOverlayByType(t, OverlayTypeService)
	peers := []AdaptivePeer{
		testAdaptivePeer(t, signedNodeRecord(t, 0x44, salt, 100, NodeRoleService), 9_500, 11, 9_100, true),
		testAdaptivePeer(t, signedNodeRecord(t, 0x45, salt, 100, NodeRoleService), 8_800, 24, 9_800, true),
		testAdaptivePeer(t, signedNodeRecord(t, 0x46, salt, 100, NodeRoleService), 8_200, 50, 8_900, false),
		testAdaptivePeer(t, signedNodeRecord(t, 0x47, salt, 100, NodeRoleService), 7_500, 80, 9_700, false),
		testAdaptivePeer(t, signedNodeRecord(t, 0x48, salt, 100, NodeRoleFull), 7_000, 120, 8_000, false),
	}
	peers[len(peers)-1].ZonesSupported = nil
	peers[len(peers)-1].Services = nil

	graph, err := BuildAdaptiveOverlayGraph(desc, local.NodeID, peers, 7, HashParts("adaptive-policy"))
	require.NoError(t, err)
	require.NoError(t, graph.Validate(desc))
	require.NotEmpty(t, graph.FastSet)
	require.NotEmpty(t, graph.StableSet)
	require.NotEmpty(t, graph.RandomSet)
	require.NotEmpty(t, graph.ZoneSet)
	require.NotEmpty(t, graph.ServiceSet)
	require.NotEmpty(t, graph.FallbackSet)
	require.Equal(t, peers[0].NodeID, graph.FastSet[0].NodeID)
	require.Equal(t, peers[1].NodeID, graph.StableSet[0].NodeID)
	require.GreaterOrEqual(t, distinctAdaptivePeerBuckets(graph.RandomSet), uint32(2))
	require.NoError(t, ValidateAdaptiveOverlayGraphUse(graph, false))
	require.ErrorContains(t, ValidateAdaptiveOverlayGraphUse(graph, true), "advisory")
	graph.LivePeerScoresCommitted = true
	require.NoError(t, ValidateAdaptiveOverlayGraphUse(graph, true))
}

func TestAdaptiveOverlayGraphRejectsEclipseAndZoneReplacement(t *testing.T) {
	desc := defaultOverlayByType(t, OverlayTypeService)
	peerA := AdaptivePeer{NodeID: HashParts("aa-peer-a"), ScoreBps: 8_000, ReliabilityBps: 8_000, ZonesSupported: []string{"APPLICATION_ZONE"}}
	peerB := AdaptivePeer{NodeID: HashParts("aa-peer-b"), ScoreBps: 7_000, ReliabilityBps: 7_000, ZonesSupported: []string{"APPLICATION_ZONE"}}
	graph := AdaptiveOverlayGraph{
		OverlayID:    desc.OverlayID,
		LocalNodeID:  HashParts("local-adaptive-node"),
		RoutingEpoch: 1,
		RandomSet:    []AdaptivePeer{peerA, peerB},
		FallbackSet:  []AdaptivePeer{peerA},
		FastSet:      []AdaptivePeer{peerA},
		StableSet:    []AdaptivePeer{peerB},
		ZoneSet:      []AdaptivePeer{peerA, peerB},
		ServiceSet:   []AdaptivePeer{},
		PolicyHash:   HashParts("bad-adaptive-policy"),
	}
	require.ErrorContains(t, graph.Validate(desc), "zone peers")

	graph.ZoneSet = nil
	graph.RandomSet = nil
	require.ErrorContains(t, graph.Validate(desc), "eclipse")
}

func TestPeerScoreDecayIsBoundedByPolicy(t *testing.T) {
	decayed, err := DecayPeerScore(PeerScore{ScoreBps: 9_000}, 3, PeerScoreDecayPolicy{
		MaxDecayBpsPerEpoch: 1_000,
		MinScoreBps:         4_000,
	})
	require.NoError(t, err)
	require.Equal(t, uint32(6_000), decayed.ScoreBps)

	decayed, err = DecayPeerScore(PeerScore{ScoreBps: 4_500}, 5, PeerScoreDecayPolicy{
		MaxDecayBpsPerEpoch: 1_000,
		MinScoreBps:         4_000,
	})
	require.NoError(t, err)
	require.Equal(t, uint32(4_000), decayed.ScoreBps)

	_, err = DecayPeerScore(PeerScore{ScoreBps: 9_000}, 1, PeerScoreDecayPolicy{
		MaxDecayBpsPerEpoch: BasisPoints + 1,
	})
	require.ErrorContains(t, err, "decay")
}

func TestRoutingTableCommitmentGuardsExecutionScheduling(t *testing.T) {
	desc := defaultOverlayByType(t, OverlayTypeService)
	commitment, err := NewRoutingTableCommitment(RoutingTableCommitment{
		RoutingEpoch: 3,
		OverlayRoots: []OverlayRouteRoot{
			{OverlayID: desc.OverlayID, RootHash: HashParts("service-overlay-root")},
		},
		ZoneRouteRoot:          HashParts("zone-route-root"),
		ServiceRouteRoot:       HashParts("service-route-root"),
		PeerClassRoot:          HashParts("peer-class-root"),
		CongestionSnapshotRoot: HashParts("congestion-snapshot-root"),
		PolicyHash:             desc.PolicyHash,
	})
	require.NoError(t, err)
	require.NoError(t, commitment.Validate())
	require.NotEmpty(t, ComputeRoutingTableCommitmentHash(commitment))

	require.NoError(t, ValidateRoutingTableUse(RoutingTableUse{
		Commitment:                commitment,
		UsedForPhysicalForwarding: true,
	}))
	require.ErrorContains(t, ValidateRoutingTableUse(RoutingTableUse{
		Commitment:                 commitment,
		UsedForExecutionScheduling: true,
	}), "execution scheduling")
	require.NoError(t, ValidateRoutingTableUse(RoutingTableUse{
		Commitment:                 commitment,
		Committed:                  true,
		UsedForExecutionScheduling: true,
	}))

	tampered := commitment
	tampered.OverlayRoots[0].RootHash = "not-a-hash"
	require.ErrorContains(t, tampered.Validate(), "lowercase hex")
}

func TestOverlayMembershipManagerRegistersMembersCanonically(t *testing.T) {
	salt := []byte("aetheris-test-network")
	desc := defaultOverlayByType(t, OverlayTypeService)
	record := signedNodeRecord(t, 0x49, salt, 100, NodeRoleService)
	proof := testOverlayMembershipProof(t, record, desc, MembershipProofServiceRegistration, OverlayMembershipModeServiceRegistry, 80)

	manager, err := NewOverlayMembershipManager(DefaultOverlayDescriptors())
	require.NoError(t, err)
	manager, membership, err := manager.Join(record, proof, 20)
	require.NoError(t, err)
	require.Equal(t, record.NodeID, membership.NodeID)
	require.Equal(t, []string{record.NodeID}, manager.Members(desc.OverlayID, 20))
	require.Empty(t, manager.Members(desc.OverlayID, 81))
	require.NoError(t, manager.Validate(20))
}

func TestRoutingGraphBuilderProducesDeterministicEdges(t *testing.T) {
	salt := []byte("aetheris-test-network")
	local := signedNodeRecord(t, 0x4a, salt, 100, NodeRoleService)
	desc := defaultOverlayByType(t, OverlayTypeService)
	peers := []AdaptivePeer{
		testAdaptivePeer(t, signedNodeRecord(t, 0x4b, salt, 100, NodeRoleService), 9_000, 15, 9_000, true),
		testAdaptivePeer(t, signedNodeRecord(t, 0x4c, salt, 100, NodeRoleService), 8_000, 25, 8_000, true),
		testAdaptivePeer(t, signedNodeRecord(t, 0x4d, salt, 100, NodeRoleFull), 7_000, 50, 7_500, false),
	}
	peers[2].ZonesSupported = nil
	peers[2].Services = nil
	manager, err := NewPeerSetManager(desc, local.NodeID, peers, 9, HashParts("peer-set-manager-policy"))
	require.NoError(t, err)

	graph, err := manager.RoutingGraph(true, HashParts("deterministic-route-hint"))
	require.NoError(t, err)
	require.True(t, graph.Committed)
	require.NotEmpty(t, graph.Edges)
	require.Equal(t, ComputeRoutingGraphHash(graph), graph.GraphHash)

	again, err := manager.RoutingGraph(true, HashParts("deterministic-route-hint"))
	require.NoError(t, err)
	require.Equal(t, graph, again)
}

func TestOverlayPartitionUsesFallbackRouteForPhysicalDelivery(t *testing.T) {
	salt := []byte("aetheris-test-network")
	source := signedNodeRecord(t, 0x4e, salt, 100, NodeRoleFull)
	service := signedNodeRecord(t, 0x4f, salt, 100, NodeRoleService)
	fallbackA := signedNodeRecord(t, 0x50, salt, 100, NodeRoleFull)
	fallbackB := signedNodeRecord(t, 0x53, salt, 100, NodeRoleFull)
	fallbackC := signedNodeRecord(t, 0x54, salt, 100, NodeRoleFull)
	desc, err := NewOverlayDescriptor(OverlayDescriptor{
		OverlayType: OverlayTypeService,
		PolicyHash:  HashParts("partitioned-service-overlay"),
		Membership:  OverlayMembershipServiceAdvertisement,
		Routing:     RoutingStrategyServiceProvider,
		MinPeers:    3,
		MaxPeers:    8,
		Fanout:      3,
		QoSClass:    QoSClassServiceCall,
		Version:     1,
	})
	require.NoError(t, err)
	msg, err := NewNetworkMessage(NetworkMessage{
		Layer:            LayerL3Application,
		Channel:          ChannelService,
		PayloadHash:      HashParts("partitioned-service-payload"),
		PayloadSizeBytes: 512,
	})
	require.NoError(t, err)
	fallbackGraph, err := BuildAdaptiveOverlayGraph(desc, source.NodeID, []AdaptivePeer{
		testGlobalAdaptivePeer(t, fallbackA, 7_000, 80, 8_000),
		testGlobalAdaptivePeer(t, fallbackB, 6_500, 90, 8_500),
		testGlobalAdaptivePeer(t, fallbackC, 6_000, 110, 7_500),
	}, 11, HashParts("fallback-partition-policy"))
	require.NoError(t, err)

	plan, err := BuildOverlayRouteWithFallback(OverlayRoutingRequest{
		Message:          msg,
		SourceNodeID:     source.NodeID,
		CandidatePeers:   []NodeRecord{service},
		MembershipProofs: []OverlayMembershipProof{testOverlayMembershipProof(t, service, desc, MembershipProofServiceRegistration, OverlayMembershipModeServiceRegistry, 80)},
		Graph:            RoutingGraph{OverlayID: desc.OverlayID, Version: 1},
		Hint:             RouteHint{ServiceID: "execution-stream"},
		CurrentHeight:    20,
	}, []OverlayDescriptor{desc}, fallbackGraph)
	require.NoError(t, err)
	require.True(t, plan.FallbackUsed)
	require.Equal(t, RoutingStrategyProbabilisticGossip, plan.Strategy)
	require.NotEmpty(t, plan.TargetNodeIDs)

	consensusMsg := msg
	consensusMsg.ConsensusEffect = true
	consensusMsg.DeterminismSource = DeterminismCommittedState
	consensusMsg.ReplaySafeID = ComputeNetworkMessageID(consensusMsg)
	_, err = BuildOverlayRouteWithFallback(OverlayRoutingRequest{
		Message:        consensusMsg,
		SourceNodeID:   source.NodeID,
		CandidatePeers: []NodeRecord{service},
		Graph:          RoutingGraph{OverlayID: desc.OverlayID, Version: 1},
		CurrentHeight:  20,
	}, []OverlayDescriptor{desc}, fallbackGraph)
	require.Error(t, err)
}

func TestAdaptivePeerRotationBoundsChurnAndKeepsStablePeers(t *testing.T) {
	salt := []byte("aetheris-test-network")
	local := signedNodeRecord(t, 0x55, salt, 100, NodeRoleService)
	desc := defaultOverlayByType(t, OverlayTypeService)
	stable := testAdaptivePeer(t, signedNodeRecord(t, 0x56, salt, 100, NodeRoleService), 9_000, 20, 9_900, true)
	oldA := testAdaptivePeer(t, signedNodeRecord(t, 0x57, salt, 100, NodeRoleService), 8_000, 40, 8_000, false)
	oldB := testAdaptivePeer(t, signedNodeRecord(t, 0x58, salt, 100, NodeRoleService), 7_500, 60, 7_500, false)
	oldC := testAdaptivePeer(t, signedNodeRecord(t, 0x59, salt, 100, NodeRoleFull), 7_000, 100, 7_000, false)
	oldC.ZonesSupported = nil
	oldC.Services = nil
	stable.LastSeenHeight = 19
	oldA.LastSeenHeight = 1
	oldB.LastSeenHeight = 1
	oldC.LastSeenHeight = 1
	manager, err := NewPeerSetManager(desc, local.NodeID, []AdaptivePeer{stable, oldA, oldB, oldC}, 10, HashParts("rotation-policy"))
	require.NoError(t, err)
	manager.RotationPolicy = PeerRotationPolicy{MaxRotatedPeersBps: 2_500, StaleAfterEpochs: 4}

	candidate := testAdaptivePeer(t, signedNodeRecord(t, 0x5a, salt, 100, NodeRoleService), 8_700, 30, 8_700, false)
	rotated, err := manager.Rotate([]AdaptivePeer{candidate}, 20)
	require.NoError(t, err)
	require.Contains(t, adaptivePeerIDs(uniqueAdaptivePeers(rotated.Graph.FastSet, rotated.Graph.StableSet, rotated.Graph.RandomSet, rotated.Graph.FallbackSet)), stable.NodeID)
	require.Contains(t, adaptivePeerIDs(uniqueAdaptivePeers(rotated.Graph.FastSet, rotated.Graph.StableSet, rotated.Graph.RandomSet, rotated.Graph.FallbackSet)), candidate.NodeID)
	require.NoError(t, rotated.Graph.Validate(desc))
}

func TestAetherMeshMessageTypesMapToChannels(t *testing.T) {
	expected := map[AetherMeshMessageType]ChannelClass{
		MeshMessageConsensus: ChannelConsensus,
		MeshMessageTx:        ChannelMempool,
		MeshMessageExecution: ChannelExecution,
		MeshMessageQuery:     ChannelService,
		MeshMessageService:   ChannelService,
		MeshMessageCrossZone: ChannelExecution,
		MeshMessageStateSync: ChannelStateSync,
		MeshMessageStorage:   ChannelData,
		MeshMessageRouting:   ChannelRouting,
	}

	require.Len(t, AetherMeshMessageTypes(), len(expected))
	for messageType, channel := range expected {
		require.True(t, IsAetherMeshMessageType(messageType))
		require.Equal(t, channel, channelForMeshMessageType(messageType))
	}
}

func TestAetherMeshMessageSignsAndRejectsTampering(t *testing.T) {
	salt := []byte("aetheris-test-network")
	originKey := deterministicPrivateKey(0x5b)
	origin := signedNodeRecord(t, 0x5b, salt, 100, NodeRoleService)
	destination := signedNodeRecord(t, 0x5c, salt, 100, NodeRoleService)
	desc := defaultOverlayByType(t, OverlayTypeService)

	msg, err := SignAetherMeshMessage(AetherMeshMessage{
		Type:            MeshMessageService,
		Payload:         []byte("service-call-payload"),
		Origin:          origin.NodeID,
		Destination:     destination.NodeID,
		Priority:        PriorityForChannel(ChannelService),
		TTL:             50,
		OverlayID:       desc.OverlayID,
		Sequence:        1,
		RouteHint:       RouteHint{ServiceID: "execution-stream"},
		DeadlineHeight:  70,
		ConsensusEffect: false,
	}, originKey)
	require.NoError(t, err)
	require.NoError(t, VerifyAetherMeshMessageSignature(msg, originKey.Public().(ed25519.PublicKey), 20))

	again, err := NewAetherMeshMessage(AetherMeshMessage{
		Type:            MeshMessageService,
		Payload:         []byte("service-call-payload"),
		Origin:          origin.NodeID,
		Destination:     destination.NodeID,
		Priority:        PriorityForChannel(ChannelService),
		TTL:             50,
		OverlayID:       desc.OverlayID,
		Sequence:        1,
		RouteHint:       RouteHint{ServiceID: "execution-stream"},
		DeadlineHeight:  70,
		ConsensusEffect: false,
	})
	require.NoError(t, err)
	require.Equal(t, msg.MessageID, again.MessageID)

	tampered := msg
	tampered.Payload = []byte("tampered")
	require.ErrorContains(t, VerifyAetherMeshMessageSignature(tampered, originKey.Public().(ed25519.PublicKey), 20), "payload hash")
}

func TestAetherMeshCrossZoneAndConsensusProofRules(t *testing.T) {
	salt := []byte("aetheris-test-network")
	origin := signedNodeRecord(t, 0x5d, salt, 100, NodeRoleZoneExecution)
	destination := signedNodeRecord(t, 0x5e, salt, 100, NodeRoleZoneExecution)
	desc := defaultOverlayByType(t, OverlayTypeExecution)

	_, err := NewAetherMeshMessage(AetherMeshMessage{
		Type:            MeshMessageCrossZone,
		Payload:         []byte("cross-zone"),
		Origin:          origin.NodeID,
		Destination:     destination.NodeID,
		Priority:        PriorityForChannel(ChannelExecution),
		TTL:             25,
		OverlayID:       desc.OverlayID,
		SourceZone:      "APPLICATION_ZONE",
		DestinationZone: "APPLICATION_ZONE",
		Sequence:        1,
	})
	require.ErrorContains(t, err, "different zones")

	_, err = NewAetherMeshMessage(AetherMeshMessage{
		Type:              MeshMessageService,
		Payload:           []byte("consensus-service"),
		Origin:            origin.NodeID,
		Destination:       destination.NodeID,
		Priority:          PriorityForChannel(ChannelService),
		TTL:               25,
		OverlayID:         defaultOverlayByType(t, OverlayTypeService).OverlayID,
		Sequence:          1,
		ConsensusEffect:   true,
		DeterminismSource: DeterminismCommittedState,
	})
	require.ErrorContains(t, err, "proof")

	msg, err := NewAetherMeshMessage(AetherMeshMessage{
		Type:              MeshMessageCrossZone,
		Payload:           []byte("cross-zone"),
		Origin:            origin.NodeID,
		Destination:       destination.NodeID,
		Priority:          PriorityForChannel(ChannelExecution),
		TTL:               25,
		OverlayID:         desc.OverlayID,
		SourceZone:        "APPLICATION_ZONE",
		DestinationZone:   "FINANCIAL_ZONE",
		Sequence:          2,
		ConsensusEffect:   true,
		DeterminismSource: DeterminismDeterministicProof,
		Proof: AetherMeshProof{
			ProofType:   "zone-commitment",
			ProofHash:   HashParts("cross-zone-proof"),
			ProofHeight: 20,
		},
	})
	require.NoError(t, err)
	base, err := msg.ToNetworkMessage()
	require.NoError(t, err)
	require.Equal(t, ChannelExecution, base.Channel)
	require.True(t, base.ConsensusEffect)
	require.Equal(t, msg.MessageID, base.ReplaySafeID)
}

func TestAetherMeshRouteUsesOverlayAndServicePeers(t *testing.T) {
	salt := []byte("aetheris-test-network")
	source := signedNodeRecord(t, 0x5f, salt, 100, NodeRoleFull)
	left := signedNodeRecord(t, 0x60, salt, 100, NodeRoleService)
	right := signedNodeRecord(t, 0x62, salt, 100, NodeRoleService)
	desc := defaultOverlayByType(t, OverlayTypeService)
	msg, err := NewAetherMeshMessage(AetherMeshMessage{
		Type:           MeshMessageService,
		Payload:        []byte("mesh-service-route"),
		Origin:         source.NodeID,
		Destination:    right.NodeID,
		Priority:       PriorityForChannel(ChannelService),
		TTL:            30,
		OverlayID:      desc.OverlayID,
		Sequence:       1,
		RouteHint:      RouteHint{ServiceID: "execution-stream"},
		DeadlineHeight: 90,
	})
	require.NoError(t, err)

	delivery, err := RouteAetherMeshMessage(AetherMeshRouteRequest{
		Message:        msg,
		SourceNodeID:   source.NodeID,
		CandidatePeers: []NodeRecord{left, right},
		MembershipProofs: []OverlayMembershipProof{
			testOverlayMembershipProof(t, left, desc, MembershipProofServiceRegistration, OverlayMembershipModeServiceRegistry, 80),
			testOverlayMembershipProof(t, right, desc, MembershipProofServiceRegistration, OverlayMembershipModeServiceRegistry, 80),
		},
		Graph: RoutingGraph{
			OverlayID: desc.OverlayID,
			Version:   1,
			Edges: []RoutingEdge{
				{FromNodeID: source.NodeID, ToNodeID: left.NodeID, LatencyMillis: 25, Priority: 1},
				{FromNodeID: source.NodeID, ToNodeID: right.NodeID, LatencyMillis: 15, Priority: 0},
			},
		},
		CurrentHeight: 20,
	}, []OverlayDescriptor{desc})
	require.NoError(t, err)
	require.Equal(t, ChannelService, delivery.Channel)
	require.Equal(t, desc.OverlayID, delivery.Route.OverlayID)
	require.False(t, delivery.Route.FallbackUsed)
	require.NotEmpty(t, delivery.Route.TargetNodeIDs)
}

func TestExecutionZoneMessageRequiresCommittedScheduleForConsensusOrder(t *testing.T) {
	salt := []byte("aetheris-test-network")
	origin := signedNodeRecord(t, 0x63, salt, 100, NodeRoleZoneExecution)
	destination := signedNodeRecord(t, 0x64, salt, 100, NodeRoleZoneExecution)
	desc := defaultOverlayByType(t, OverlayTypeExecution)
	mesh, err := NewAetherMeshMessage(AetherMeshMessage{
		Type:              MeshMessageExecution,
		Payload:           []byte("execution-zone-payload"),
		Origin:            origin.NodeID,
		Destination:       destination.NodeID,
		Priority:          PriorityForChannel(ChannelExecution),
		TTL:               40,
		OverlayID:         desc.OverlayID,
		DestinationZone:   "APPLICATION_ZONE",
		Sequence:          7,
		ConsensusEffect:   true,
		DeterminismSource: DeterminismCommittedState,
		Proof: AetherMeshProof{
			ProofType:   "committed-schedule",
			ProofHash:   HashParts("execution-proof"),
			ProofHeight: 30,
		},
	})
	require.NoError(t, err)
	uncommitted, err := NewExecutionMessageSchedule(ExecutionMessageSchedule{
		ZoneID:            "APPLICATION_ZONE",
		ShardID:           "shard-1",
		RoutingClass:      ExecutionRoutingExecutionOverlay,
		Ordered:           true,
		MessageIDs:        []string{mesh.MessageID},
		FirstZoneSequence: 7,
		LastZoneSequence:  7,
	})
	require.NoError(t, err)

	_, err = NewExecutionZoneMessage(ExecutionZoneMessage{
		Message:                mesh,
		RoutingClass:           ExecutionRoutingExecutionOverlay,
		ZoneID:                 "APPLICATION_ZONE",
		ShardID:                "shard-1",
		ExecutionOverlayID:     desc.OverlayID,
		ZoneSequence:           7,
		NetworkDeliveryOrdinal: 99,
		ConsensusScheduleOrder: 1,
	}, uncommitted)
	require.ErrorContains(t, err, "committed schedule")

	committed, err := NewExecutionMessageSchedule(ExecutionMessageSchedule{
		ZoneID:            "APPLICATION_ZONE",
		ShardID:           "shard-1",
		RoutingClass:      ExecutionRoutingExecutionOverlay,
		Committed:         true,
		Ordered:           true,
		MessageIDs:        []string{mesh.MessageID},
		FirstZoneSequence: 7,
		LastZoneSequence:  7,
	})
	require.NoError(t, err)
	executionMsg, err := NewExecutionZoneMessage(ExecutionZoneMessage{
		Message:                mesh,
		RoutingClass:           ExecutionRoutingExecutionOverlay,
		ZoneID:                 "APPLICATION_ZONE",
		ShardID:                "shard-1",
		ExecutionOverlayID:     desc.OverlayID,
		ZoneSequence:           7,
		NetworkDeliveryOrdinal: 99,
		ConsensusScheduleOrder: 1,
	}, committed)
	require.NoError(t, err)
	require.NotEqual(t, executionMsg.NetworkDeliveryOrdinal, executionMsg.ConsensusScheduleOrder)
	require.Equal(t, committed.ScheduleID, executionMsg.ConsensusScheduleID)
}

func TestExecutionZoneMessageSupportsAsyncParallelBlockSTMGroups(t *testing.T) {
	salt := []byte("aetheris-test-network")
	origin := signedNodeRecord(t, 0x65, salt, 100, NodeRoleZoneExecution)
	destination := signedNodeRecord(t, 0x66, salt, 100, NodeRoleZoneExecution)
	desc := defaultOverlayByType(t, OverlayTypeExecution)
	mesh, err := NewAetherMeshMessage(AetherMeshMessage{
		Type:            MeshMessageExecution,
		Payload:         []byte("async-execution"),
		Origin:          origin.NodeID,
		Destination:     destination.NodeID,
		Priority:        PriorityForChannel(ChannelExecution),
		TTL:             40,
		OverlayID:       desc.OverlayID,
		DestinationZone: "APPLICATION_ZONE",
		Sequence:        8,
	})
	require.NoError(t, err)

	_, err = NewExecutionZoneMessage(ExecutionZoneMessage{
		Message:               mesh,
		RoutingClass:          ExecutionRoutingShard,
		ZoneID:                "APPLICATION_ZONE",
		ShardID:               "shard-2",
		ExecutionOverlayID:    desc.OverlayID,
		ExecutionGroupID:      HashParts("async-group"),
		ZoneSequence:          8,
		Async:                 true,
		ParallelZoneExecution: true,
	}, ExecutionMessageSchedule{})
	require.ErrorContains(t, err, "BlockSTM")

	executionMsg, err := NewExecutionZoneMessage(ExecutionZoneMessage{
		Message:               mesh,
		RoutingClass:          ExecutionRoutingShard,
		ZoneID:                "APPLICATION_ZONE",
		ShardID:               "shard-2",
		ExecutionOverlayID:    desc.OverlayID,
		ExecutionGroupID:      HashParts("async-group"),
		BlockSTMGroupID:       HashParts("blockstm-group"),
		ZoneSequence:          8,
		Async:                 true,
		ParallelZoneExecution: true,
	}, ExecutionMessageSchedule{})
	require.NoError(t, err)
	require.True(t, executionMsg.Async)
	require.True(t, executionMsg.ParallelZoneExecution)
}

func TestCrossZoneMessageRequiresSequenceExpiryAndReplayProtection(t *testing.T) {
	msg := CrossZoneMessage{
		SourceZone:      "APPLICATION_ZONE",
		DestinationZone: "FINANCIAL_ZONE",
		MessageHash:     HashParts("cross-zone-message"),
		ExpiryHeight:    90,
		ReceiptPolicy:   ReceiptPolicyOnExecution,
		ProofRequired:   true,
	}
	_, err := NewCrossZoneMessage(msg)
	require.ErrorContains(t, err, "source sequence")

	msg.SourceSequence = 11
	created, err := NewCrossZoneMessage(msg)
	require.NoError(t, err)
	require.NotEmpty(t, ComputeCrossZoneExecutionKey(created))

	guard := CrossZoneReplayGuard{}
	guard, err = guard.Accept(created, 20)
	require.NoError(t, err)
	require.Len(t, guard.ExecutedKeys, 1)
	_, err = guard.Accept(created, 21)
	require.ErrorContains(t, err, "already executed")

	_, err = guard.Accept(created, 91)
	require.ErrorContains(t, err, "expired")
}

func TestCrossZoneReceiptIsRollbackSafeAndProofQueryable(t *testing.T) {
	receipt := CrossZoneReceipt{
		SourceZone:      "APPLICATION_ZONE",
		DestinationZone: "FINANCIAL_ZONE",
		SourceSequence:  11,
		MessageHash:     HashParts("cross-zone-message"),
		Status:          CrossZoneReceiptExecuted,
		ReceiptPolicy:   ReceiptPolicyAlways,
		ProofHash:       HashParts("cross-zone-receipt-proof"),
		ReceiptHeight:   35,
		RollbackSafe:    true,
		ProofQueryable:  true,
	}
	created, err := NewCrossZoneReceipt(receipt)
	require.NoError(t, err)
	require.Equal(t, ComputeCrossZoneReceiptID(created), created.ReceiptID)

	broken := created
	broken.RollbackSafe = false
	broken.ReceiptID = ComputeCrossZoneReceiptID(broken)
	require.ErrorContains(t, broken.Validate(), "rollback-safe")

	bounced := receipt
	bounced.Status = CrossZoneReceiptBounced
	_, err = NewCrossZoneReceipt(bounced)
	require.ErrorContains(t, err, "bounced")

	bounced.Bounced = true
	_, err = NewCrossZoneReceipt(bounced)
	require.NoError(t, err)
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

func defaultOverlayByType(t *testing.T, overlayType OverlayType) OverlayDescriptor {
	t.Helper()

	for _, desc := range DefaultOverlayDescriptors() {
		if desc.OverlayType == overlayType {
			return desc
		}
	}
	t.Fatalf("missing default overlay %s", overlayType)
	return OverlayDescriptor{}
}

func overlayIDs(descriptors []OverlayDescriptor) []string {
	out := make([]string, len(descriptors))
	for i, desc := range descriptors {
		out[i] = NormalizeOverlayDescriptor(desc).OverlayID
	}
	return out
}

func testOverlayMembershipProof(t *testing.T, record NodeRecord, desc OverlayDescriptor, proofType MembershipProofType, mode OverlayMembershipMode, expiresHeight uint64) OverlayMembershipProof {
	t.Helper()

	record = NormalizeNodeRecord(record)
	desc = NormalizeOverlayDescriptor(desc)
	proof := OverlayMembershipProof{
		OverlayID:     desc.OverlayID,
		NodeID:        record.NodeID,
		ProofType:     proofType,
		Mode:          mode,
		Membership:    desc.Membership,
		ProofHash:     HashParts("overlay-membership-proof", record.NodeID, desc.OverlayID, string(proofType)),
		AuthorityHash: HashParts("overlay-membership-authority", desc.OverlayID),
		ExpiresHeight: expiresHeight,
	}
	if len(record.ZonesSupported) > 0 {
		proof.ZoneID = record.ZonesSupported[0]
	}
	if len(record.ServicesSupported) > 0 {
		proof.ServiceID = record.ServicesSupported[0]
	}
	if proofType == MembershipProofProviderStake {
		proof.StakeAmount = 1_000
		proof.Committed = true
	}
	if proofType == MembershipProofValidatorSet {
		proof.Committed = true
		proof.Deterministic = true
	}
	if proofType == MembershipProofSignedAuthorization {
		signed, err := SignOverlayMembershipProof(proof, deterministicPrivateKey(0x7a))
		require.NoError(t, err)
		return signed
	}
	created, err := NewOverlayMembershipProof(proof)
	require.NoError(t, err)
	return created
}

func testAdaptivePeer(t *testing.T, record NodeRecord, scoreBps uint32, latencyMillis uint64, reliabilityBps uint32, committed bool) AdaptivePeer {
	t.Helper()

	score := PeerScore{ScoreBps: scoreBps}
	metrics := PeerMetrics{
		LatencyMillis:  latencyMillis,
		ReliabilityBps: reliabilityBps,
	}
	return AdaptivePeerFromNodeRecord(record, score, metrics, committed, 20)
}

func testGlobalAdaptivePeer(t *testing.T, record NodeRecord, scoreBps uint32, latencyMillis uint64, reliabilityBps uint32) AdaptivePeer {
	t.Helper()

	peer := testAdaptivePeer(t, record, scoreBps, latencyMillis, reliabilityBps, false)
	peer.ZonesSupported = nil
	peer.Services = nil
	return peer
}

func adaptivePeerIDs(peers []AdaptivePeer) []string {
	out := make([]string, len(peers))
	for i, peer := range peers {
		out[i] = normalizeHashText(peer.NodeID)
	}
	sortStrings(out)
	return out
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
