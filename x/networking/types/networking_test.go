package types

import (
	"bytes"
	"crypto/ed25519"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNodeRecordSignatureIdentityAndExpiry(t *testing.T) {
	salt := []byte("aetra-test-network")
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
	salt := []byte("aetra-test-network")
	privateKey := deterministicPrivateKey(0x12)
	leftAddressHash, err := HashNetworkAddresses([]string{"tcp://10.0.0.1:26656"})
	require.NoError(t, err)
	rightAddressHash, err := HashNetworkAddresses([]string{"tcp://10.0.0.2:26656"})
	require.NoError(t, err)

	left, err := SignNodeRecord(NodeRecord{
		Roles:			[]NodeRole{NodeRoleFull},
		NetworkAddressesHash:	leftAddressHash,
		ProtocolVersions:	[]string{DefaultProtocolVersion},
		ExpiresHeight:		100,
	}, privateKey, salt)
	require.NoError(t, err)
	right, err := SignNodeRecord(NodeRecord{
		Roles:			[]NodeRole{NodeRoleFull},
		NetworkAddressesHash:	rightAddressHash,
		ProtocolVersions:	[]string{DefaultProtocolVersion},
		ExpiresHeight:		100,
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
		ValidatorPubKey:	validatorKey,
		Roles:			[]NodeRole{NodeRoleValidator, NodeRoleFull},
		NetworkAddressesHash:	leftAddressHash,
		ProtocolVersions:	[]string{DefaultProtocolVersion},
		ExpiresHeight:		100,
	}, privateKey, salt)
	require.NoError(t, err)
	require.Equal(t, ComputeNodeID(validatorKey, salt), validator.NodeID)
	require.NotEqual(t, ComputeNodeID(validator.NodePubKey, salt), validator.NodeID)
}

func TestSignedIdentityTransitionRotatesNodeIdentity(t *testing.T) {
	salt := []byte("aetra-test-network")
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
		NodeID:		oldRecord.NodeID,
		Role:		NodeRoleService,
		Bonded:		true,
		CommitmentHash:	HashParts("old-service-commitment"),
		ExpiresHeight:	70,
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

	salt := []byte("aetra-test-network")
	record := signedNodeRecord(t, 0x19, salt, 100, NodeRoleFull)
	record.NetworkAddressesHash = left
	payload, err := record.SigningPayload()
	require.NoError(t, err)
	record.Signature = ed25519.Sign(deterministicPrivateKey(0x19), payload)
	require.NoError(t, VerifyNodeRecordAddresses(record, addresses))
	require.ErrorContains(t, VerifyNodeRecordAddresses(record, []string{"tcp://10.0.0.3:26656"}), "address list")
}

func TestSessionNegotiationCreatesDeterministicStreams(t *testing.T) {
	salt := []byte("aetra-test-network")
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
	salt := []byte("aetra-test-network")
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
	salt := []byte("aetra-test-network")
	local := signedNodeRecord(t, 0x28, salt, 100, NodeRoleFull)
	remote := signedNodeRecord(t, 0x29, salt, 100, NodeRoleService)
	handshake, err := RunSessionHandshake(local, remote, testSessionRequest(local, remote, 20, 80, "rotate-session-keys", nil), salt, 20, nil)
	require.NoError(t, err)

	rotated, err := RotateSessionKeys(handshake.Session, SessionKeyRotationRequest{
		SessionID:			handshake.Session.SessionID,
		NewLocalEphemeralPubKey:	bytes.Repeat([]byte{0xc3}, SessionEphemeralKeyBytes),
		NewRemoteEphemeralPubKey:	bytes.Repeat([]byte{0xd4}, SessionEphemeralKeyBytes),
		NewSecretCommitmentHash:	HashParts("rotated-session-secret", handshake.Session.SessionID),
		RotatedAtHeight:		40,
		ExpiresHeight:			80,
		Nonce:				[]byte("rotation-nonce"),
	})
	require.NoError(t, err)
	require.Equal(t, handshake.Session.SessionID, rotated.SessionID)
	require.NotEqual(t, handshake.Session.SessionKeys.KeyID, rotated.SessionKeys.KeyID)
	require.Equal(t, uint64(40), rotated.SessionKeys.EstablishedHeight)
	for _, stream := range rotated.Streams {
		require.Equal(t, streamEncryptionContext(rotated.SessionKeys.KeyID, stream.StreamID), stream.EncryptionContext)
	}

	_, err = RotateSessionKeys(handshake.Session, SessionKeyRotationRequest{
		SessionID:			handshake.Session.SessionID,
		NewLocalEphemeralPubKey:	bytes.Repeat([]byte{0xc3}, SessionEphemeralKeyBytes),
		NewRemoteEphemeralPubKey:	bytes.Repeat([]byte{0xd4}, SessionEphemeralKeyBytes),
		NewSecretCommitmentHash:	HashParts("rotated-session-secret", handshake.Session.SessionID),
		RotatedAtHeight:		81,
		ExpiresHeight:			90,
		Nonce:				[]byte("late-rotation"),
	})
	require.ErrorContains(t, err, "outside session range")
}

func TestMultiplexedStreamsEnforceEncryptionCapacityAndResetPolicy(t *testing.T) {
	salt := []byte("aetra-test-network")
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
		Channel:	ChannelService,
		SizeBytes:	512,
		EnqueuedHeight:	1,
		Sequence:	1,
		PayloadHash:	HashParts("service"),
	}
	data := TransportEnvelope{
		Channel:	ChannelData,
		SizeBytes:	2 << 20,
		EnqueuedHeight:	1,
		Sequence:	2,
		PayloadHash:	HashParts("data"),
	}
	consensus := TransportEnvelope{
		Channel:	ChannelConsensus,
		SizeBytes:	128,
		EnqueuedHeight:	100,
		Sequence:	99,
		PayloadHash:	HashParts("consensus"),
	}

	next, found, err := SelectNextEnvelope([]TransportEnvelope{service, data, consensus}, policies)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, ChannelConsensus, next.Channel)
}

func TestL0ChannelIDsAreStableAndRoundTrip(t *testing.T) {
	expected := map[ChannelClass]ChannelID{
		ChannelConsensus:	ChannelIDConsensus,
		ChannelMempool:		ChannelIDMempool,
		ChannelBlock:		ChannelIDBlock,
		ChannelStateSync:	ChannelIDStateSync,
		ChannelData:		ChannelIDData,
		ChannelExecution:	ChannelIDExecution,
		ChannelService:		ChannelIDService,
		ChannelRouting:		ChannelIDRouting,
		ChannelDiscovery:	ChannelIDDiscovery,
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
	payload := bytes.Repeat([]byte("aetra-networking"), 512)
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

func TestRL2TransferIDValidationAndChannelMapping(t *testing.T) {
	source := HashParts("rl2-node", "source")
	target := HashParts("rl2-node", "target")
	payloadRoot := HashParts("rl2-payload", "root")
	cases := []struct {
		payloadType	RL2PayloadType
		channel		ChannelClass
	}{
		{RL2PayloadLargeBlock, ChannelBlock},
		{RL2PayloadBlockChunk, ChannelBlock},
		{RL2PayloadStateSyncStream, ChannelStateSync},
		{RL2PayloadZoneSnapshot, ChannelStateSync},
		{RL2PayloadExecutionResult, ChannelExecution},
		{RL2PayloadStorageObject, ChannelData},
		{RL2PayloadProofSet, ChannelData},
	}

	for _, tc := range cases {
		t.Run(string(tc.payloadType), func(t *testing.T) {
			transfer, err := NewRL2Transfer(RL2Transfer{
				SourceNode:	source,
				TargetNode:	target,
				PayloadType:	tc.payloadType,
				PayloadRoot:	payloadRoot,
				ChunkCount:	4,
				ChunkSize:	1024,
				Priority:	DefaultRL2Priority(tc.payloadType),
				DeadlineHeight:	100,
			})
			require.NoError(t, err)
			require.Equal(t, ComputeRL2TransferID(transfer), transfer.TransferID)
			require.Equal(t, RL2FECNone, transfer.FECPolicy)
			require.Equal(t, tc.channel, RL2ChannelForPayloadType(tc.payloadType))

			envelope, err := transfer.TransportEnvelope(20, 7)
			require.NoError(t, err)
			require.Equal(t, tc.channel, envelope.Channel)
			require.Equal(t, transfer.ChunkSize, envelope.SizeBytes)
			require.Equal(t, transfer.PayloadRoot, envelope.PayloadHash)
		})
	}
}

func TestRL2TransferRejectsExpiredInvalidRootAndChunkBounds(t *testing.T) {
	source := HashParts("rl2-node", "source")
	target := HashParts("rl2-node", "target")
	payloadRoot := HashParts("rl2-payload", "root")
	valid, err := NewRL2Transfer(RL2Transfer{
		SourceNode:	source,
		TargetNode:	target,
		PayloadType:	RL2PayloadStateSyncStream,
		PayloadRoot:	payloadRoot,
		ChunkCount:	2,
		ChunkSize:	512,
		Priority:	DefaultRL2Priority(RL2PayloadStateSyncStream),
		DeadlineHeight:	50,
	})
	require.NoError(t, err)

	require.ErrorContains(t, valid.Validate(51), "expired")

	_, err = NewRL2Transfer(RL2Transfer{
		SourceNode:	source,
		TargetNode:	target,
		PayloadType:	RL2PayloadStateSyncStream,
		PayloadRoot:	"bad-root",
		ChunkCount:	2,
		ChunkSize:	512,
		Priority:	DefaultRL2Priority(RL2PayloadStateSyncStream),
	})
	require.ErrorContains(t, err, "payload root")

	_, err = NewRL2Transfer(RL2Transfer{
		SourceNode:	source,
		TargetNode:	target,
		PayloadType:	RL2PayloadStateSyncStream,
		PayloadRoot:	payloadRoot,
		ChunkSize:	512,
		Priority:	DefaultRL2Priority(RL2PayloadStateSyncStream),
	})
	require.ErrorContains(t, err, "chunk count")

	_, err = NewRL2Transfer(RL2Transfer{
		SourceNode:	source,
		TargetNode:	target,
		PayloadType:	RL2PayloadStateSyncStream,
		PayloadRoot:	payloadRoot,
		ChunkCount:	2,
		ChunkSize:	MaxChunkBytes + 1,
		Priority:	DefaultRL2Priority(RL2PayloadStateSyncStream),
	})
	require.ErrorContains(t, err, "chunk size")

	withBadResume := valid
	withBadResume.ResumeToken = "not-a-hash"
	require.ErrorContains(t, withBadResume.Validate(10), "resume token")
}

func TestRL2TransferBuildsFromChunksResumeTokenAndPropagationPlan(t *testing.T) {
	source := HashParts("rl2-node", "source")
	target := HashParts("rl2-node", "target")
	payload := bytes.Repeat([]byte("rl2-state-sync-stream"), 128)
	chunks, err := ChunkPayload(payload, 128)
	require.NoError(t, err)
	require.Greater(t, len(chunks), 2)

	transfer, err := NewRL2TransferFromChunks(
		source,
		target,
		RL2PayloadStorageObject,
		chunks,
		DefaultRL2Priority(RL2PayloadStorageObject),
		80,
		RL2FECXORParity,
	)
	require.NoError(t, err)
	chunkRoot, err := ComputeRL2ChunkRoot(chunks)
	require.NoError(t, err)
	require.Equal(t, chunkRoot, transfer.PayloadRoot)
	require.Equal(t, chunks[0].Total, transfer.ChunkCount)
	require.Equal(t, RL2FECXORParity, transfer.FECPolicy)

	progress, err := NewRL2TransferProgress(transfer, []uint32{2, 0})
	require.NoError(t, err)
	sameProgress, err := NewRL2TransferProgress(transfer, []uint32{0, 2})
	require.NoError(t, err)
	require.Equal(t, []uint32{0, 2}, progress.ReceivedChunks)
	require.Equal(t, sameProgress.ResumeToken, progress.ResumeToken)

	_, err = NewRL2TransferProgress(transfer, []uint32{0, 0})
	require.ErrorContains(t, err, "unique")
	_, err = NewRL2TransferProgress(transfer, []uint32{transfer.ChunkCount})
	require.ErrorContains(t, err, "out of range")

	plan, err := PlanRL2Transfer(DefaultAetherNetworkingAdapter(), transfer, 20, 9, 12, PeerScore{ScoreBps: 9_000})
	require.NoError(t, err)
	require.Equal(t, ChannelData, plan.Envelope.Channel)
	require.False(t, plan.HandledByCometBFT)
	require.True(t, plan.UsesAdvisoryPeerMetric)
	require.Greater(t, plan.AdapterFanout, uint32(0))
}

func TestRL2ChunkDescriptorsVerifyOrderedMerkleRootAndChunkBytes(t *testing.T) {
	source := HashParts("rl2-node", "source")
	target := HashParts("rl2-node", "target")
	payload := bytes.Repeat([]byte("rl2-proof-set"), 96)
	chunks, err := ChunkPayload(payload, 64)
	require.NoError(t, err)
	require.Greater(t, len(chunks), 2)

	transfer, err := NewRL2TransferFromChunks(
		source,
		target,
		RL2PayloadProofSet,
		chunks,
		DefaultRL2Priority(RL2PayloadProofSet),
		0,
		RL2FECNone,
	)
	require.NoError(t, err)

	descriptors, err := NewRL2ChunkDescriptors(transfer, chunks)
	require.NoError(t, err)
	require.Len(t, descriptors, len(chunks))
	require.NoError(t, ValidateRL2ChunkDescriptors(transfer, descriptors))

	for i, descriptor := range descriptors {
		require.Equal(t, uint32(i), descriptor.ChunkIndex)
		require.NotEmpty(t, descriptor.ProofPath)
		require.NoError(t, VerifyRL2Chunk(transfer, descriptor, chunks[i]))
	}

	tampered := descriptors[1]
	tampered.ChunkHash = HashParts("rl2", "wrong-chunk")
	require.ErrorContains(t, VerifyRL2ChunkProof(tampered, transfer.PayloadRoot, transfer.ChunkCount), "root mismatch")

	badRange := descriptors
	badRange[1].RangeStart++
	require.ErrorContains(t, ValidateRL2ChunkDescriptors(transfer, badRange), "contiguous")
}

func TestRL2MissingChunksRequestUsesVerifiedBitmapResumeToken(t *testing.T) {
	source := HashParts("rl2-node", "source")
	target := HashParts("rl2-node", "target")
	payload := bytes.Repeat([]byte{0xa7}, 256)
	chunks, err := ChunkPayload(payload, 64)
	require.NoError(t, err)

	transfer, err := NewRL2TransferFromChunks(
		source,
		target,
		RL2PayloadZoneSnapshot,
		chunks,
		DefaultRL2Priority(RL2PayloadZoneSnapshot),
		120,
		RL2FECReedSolomon,
	)
	require.NoError(t, err)

	progress, err := NewRL2TransferProgress(transfer, []uint32{0, 2})
	require.NoError(t, err)
	require.Len(t, progress.VerifiedBitmap, int(transfer.ChunkCount))
	require.True(t, progress.VerifiedBitmap[0])
	require.True(t, progress.VerifiedBitmap[2])
	require.False(t, progress.VerifiedBitmap[1])

	request, err := NewRL2ChunkRequest(transfer, []uint32{0, 2})
	require.NoError(t, err)
	require.Equal(t, transfer.TransferID, request.TransferID)
	require.Equal(t, progress.ResumeToken, request.ResumeToken)
	require.Equal(t, []uint32{1, 3}, request.MissingIndexes)
}

func TestRL2StreamingPlanAdaptsBandwidthBackpressureParallelismAndFEC(t *testing.T) {
	source := HashParts("rl2-node", "source")
	target := HashParts("rl2-node", "target")
	payload := bytes.Repeat([]byte("rl2-streaming"), 256)
	chunks, err := ChunkPayload(payload, 128)
	require.NoError(t, err)

	transfer, err := NewRL2TransferFromChunks(
		source,
		target,
		RL2PayloadStateSyncStream,
		chunks,
		DefaultRL2Priority(RL2PayloadStateSyncStream),
		0,
		RL2FECXORParity,
	)
	require.NoError(t, err)

	plan, err := PlanRL2Streaming(transfer, PeerScore{ScoreBps: 5_000}, 1024, 0, 4)
	require.NoError(t, err)
	require.Equal(t, ChannelStateSync, plan.Channel)
	require.Equal(t, transfer.Priority, plan.PriorityLane)
	require.Equal(t, uint32(4), plan.ParallelStreams)
	require.Equal(t, uint64(512), plan.ChunkBudgetBytes)
	require.True(t, plan.FECEnabled)
	require.False(t, plan.BackpressureActive)
	require.Greater(t, plan.MaxInFlightChunks, uint32(1))

	congested, err := PlanRL2Streaming(transfer, PeerScore{ScoreBps: 9_000}, 1024, 1024, 4)
	require.NoError(t, err)
	require.True(t, congested.BackpressureActive)
	require.Equal(t, uint32(1), congested.ParallelStreams)
	require.Equal(t, uint32(1), congested.MaxInFlightChunks)
}

func TestRL2OfferStateMachineResumesInterruptedTransferAndCompletes(t *testing.T) {
	source := HashParts("rl2-node", "source")
	target := HashParts("rl2-node", "target")
	payload := bytes.Repeat([]byte{0x42}, 512)
	chunks, err := ChunkPayload(payload, 128)
	require.NoError(t, err)

	offer, descriptors, err := NewRL2TransferOfferFromChunks(
		source,
		target,
		RL2PayloadStateSyncStream,
		chunks,
		DefaultRL2Priority(RL2PayloadStateSyncStream),
		10,
		100,
		RL2FECReedSolomon,
		PeerScore{ScoreBps: 8_000},
		1024,
		4,
	)
	require.NoError(t, err)
	require.Equal(t, ComputeRL2OfferID(offer), offer.OfferID)
	require.NoError(t, offer.Validate(10))

	session, err := AcceptRL2TransferOffer(offer, nil, 11)
	require.NoError(t, err)
	require.Equal(t, RL2StateAccepted, session.State)

	session, err = StartRL2Transfer(session, PeerScore{ScoreBps: 8_000}, 1024, 0)
	require.NoError(t, err)
	require.Equal(t, RL2StateStreaming, session.State)

	session, err = AcceptRL2Chunk(session, descriptors[0], chunks[0], 12)
	require.NoError(t, err)
	session, err = AcceptRL2Chunk(session, descriptors[1], chunks[1], 13)
	require.NoError(t, err)
	require.Equal(t, RL2StateStreaming, session.State)
	require.Equal(t, []uint32{0, 1}, session.Progress.ReceivedChunks)

	signal, err := NewRL2BackpressureSignal(offer.Transfer, 2048, 512, session.Progress.ReceivedChunks)
	require.NoError(t, err)
	require.True(t, signal.PauseRequested)
	session, err = PauseRL2Transfer(session, signal, 14)
	require.NoError(t, err)
	require.Equal(t, RL2StatePaused, session.State)

	request, err := NewRL2ChunkRequest(offer.Transfer, session.Progress.ReceivedChunks)
	require.NoError(t, err)
	require.Equal(t, []uint32{2, 3}, request.MissingIndexes)
	session, err = ResumeRL2Transfer(session, session.Progress.ReceivedChunks, 15)
	require.NoError(t, err)
	require.Equal(t, RL2StateResumed, session.State)

	session, err = StartRL2Transfer(session, PeerScore{ScoreBps: 9_000}, 2048, 0)
	require.NoError(t, err)
	for _, index := range request.MissingIndexes {
		session, err = AcceptRL2Chunk(session, descriptors[index], chunks[index], 16+uint64(index))
		require.NoError(t, err)
	}
	require.Equal(t, RL2StateVerified, session.State)

	session, decoded, err := VerifyRL2TransferCompletion(session, descriptors, chunks, 20)
	require.NoError(t, err)
	require.Equal(t, RL2StateCompleted, session.State)
	require.Equal(t, payload, decoded)
	require.True(t, IsRL2TerminalState(session.State))
}

func TestRL2StateMachineClassifiesInvalidChunksAndRootMismatch(t *testing.T) {
	source := HashParts("rl2-node", "source")
	target := HashParts("rl2-node", "target")
	payload := bytes.Repeat([]byte{0x24}, 256)
	chunks, err := ChunkPayload(payload, 64)
	require.NoError(t, err)

	offer, descriptors, err := NewRL2TransferOfferFromChunks(
		source,
		target,
		RL2PayloadStorageObject,
		chunks,
		DefaultRL2Priority(RL2PayloadStorageObject),
		10,
		100,
		RL2FECNone,
		PeerScore{ScoreBps: 7_000},
		1024,
		2,
	)
	require.NoError(t, err)

	session, err := AcceptRL2TransferOffer(offer, nil, 11)
	require.NoError(t, err)
	session, err = StartRL2Transfer(session, PeerScore{ScoreBps: 7_000}, 1024, 0)
	require.NoError(t, err)

	corruptChunk := chunks[0]
	corruptChunk.Bytes[0] ^= 0xff
	session, err = AcceptRL2Chunk(session, descriptors[0], corruptChunk, 12)
	require.NoError(t, err)
	require.Equal(t, RL2StateInvalidChunk, session.State)
	require.Contains(t, session.FailureReason, "chunk hash")
	require.True(t, IsRL2FailureState(session.State))

	session, err = AcceptRL2TransferOffer(offer, nil, 11)
	require.NoError(t, err)
	session, err = StartRL2Transfer(session, PeerScore{ScoreBps: 7_000}, 1024, 0)
	require.NoError(t, err)
	badDescriptor := descriptors[1]
	badDescriptor.ChunkHash = HashParts("rl2", "different-chunk")
	session, err = AcceptRL2Chunk(session, badDescriptor, chunks[1], 12)
	require.NoError(t, err)
	require.Equal(t, RL2StateRootMismatch, session.State)
	require.Contains(t, session.FailureReason, "root mismatch")

	session, err = AcceptRL2TransferOffer(offer, nil, 11)
	require.NoError(t, err)
	session, err = FailRL2Transfer(session, RL2StatePeerDisconnected, "remote closed session", 12)
	require.NoError(t, err)
	require.Equal(t, RL2StatePeerDisconnected, session.State)
	require.True(t, IsRL2TerminalState(session.State))
}

func TestRL2AdaptiveChunkSizingAndOfferValidation(t *testing.T) {
	size, err := RecommendRL2ChunkSize(4096, PeerScore{ScoreBps: 5_000}, 2048, 128, 1024)
	require.NoError(t, err)
	require.Equal(t, uint64(256), size)

	_, err = RecommendRL2ChunkSize(4096, PeerScore{ScoreBps: BasisPoints + 1}, 2048, 128, 1024)
	require.ErrorContains(t, err, "score")

	source := HashParts("rl2-node", "source")
	target := HashParts("rl2-node", "target")
	payload := bytes.Repeat([]byte{0x11}, 256)
	chunks, err := ChunkPayload(payload, 64)
	require.NoError(t, err)
	offer, _, err := NewRL2TransferOfferFromChunks(
		source,
		target,
		RL2PayloadExecutionResult,
		chunks,
		DefaultRL2Priority(RL2PayloadExecutionResult),
		10,
		20,
		RL2FECNone,
		PeerScore{ScoreBps: 5_000},
		512,
		2,
	)
	require.NoError(t, err)
	require.ErrorContains(t, offer.Validate(21), "expired")

	tampered := offer
	tampered.SuggestedChunkSize = offer.Transfer.ChunkSize + 1
	tampered.OfferID = ComputeRL2OfferID(tampered)
	require.ErrorContains(t, tampered.Validate(10), "suggested chunk size")
}

func TestNetworkingStateRegistersNodesAndSessionsCanonically(t *testing.T) {
	salt := []byte("aetra-test-network")
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
		OverlayType:	OverlayTypeService,
		PolicyHash:	HashParts("bad-service-overlay-qos"),
		Membership:	OverlayMembershipServiceAdvertisement,
		Routing:	RoutingStrategyLowLatencyAdvisory,
		MinPeers:	2,
		MaxPeers:	8,
		Fanout:		2,
		QoSClass:	QoSClassBulkData,
		Version:	1,
	})
	require.ErrorContains(t, err, "qos")

	_, err = NewOverlayDescriptor(OverlayDescriptor{
		OverlayType:	OverlayTypeService,
		PolicyHash:	HashParts("bad-service-overlay-membership"),
		Membership:	OverlayMembershipRoutingRole,
		Routing:	RoutingStrategyLowLatencyAdvisory,
		MinPeers:	2,
		MaxPeers:	8,
		Fanout:		2,
		QoSClass:	QoSClassServiceCall,
		Version:	1,
	})
	require.ErrorContains(t, err, "membership")

	_, err = NewOverlayDescriptor(OverlayDescriptor{
		OverlayType:	OverlayTypeData,
		PolicyHash:	HashParts("bad-data-overlay-fanout"),
		Membership:	OverlayMembershipDataProvider,
		Routing:	RoutingStrategyKBucket,
		MinPeers:	2,
		MaxPeers:	4,
		Fanout:		5,
		QoSClass:	QoSClassBulkData,
		Version:	1,
	})
	require.ErrorContains(t, err, "fanout")

	_, err = NewOverlayDescriptor(OverlayDescriptor{
		OverlayType:	OverlayTypeValidator,
		PolicyHash:	HashParts("bad-validator-overlay-routing"),
		Membership:	OverlayMembershipValidatorSet,
		Routing:	RoutingStrategyLowLatencyAdvisory,
		MinPeers:	4,
		MaxPeers:	32,
		Fanout:		4,
		QoSClass:	QoSClassCriticalConsensus,
		Version:	1,
	})
	require.ErrorContains(t, err, "advisory")
}

func TestOverlayMembershipMatchesNodeRolesAndCapabilities(t *testing.T) {
	salt := []byte("aetra-test-network")
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
		ValidatorPubKey:	validatorKey,
		Roles:			[]NodeRole{NodeRoleValidator, NodeRoleFull},
		NetworkAddressesHash:	addressHash,
		ZonesSupported:		[]string{"APPLICATION_ZONE"},
		ProtocolVersions:	[]string{DefaultProtocolVersion},
		ExpiresHeight:		100,
	}, validatorPrivateKey, salt)
	require.NoError(t, err)

	matches, err = NodeSatisfiesOverlayMembership(validator, defaultOverlayByType(t, OverlayTypeValidator))
	require.NoError(t, err)
	require.True(t, matches)
}

func TestNetworkingStateRegistersAndPrunesOverlayDescriptors(t *testing.T) {
	state := EmptyState()
	desc, err := NewOverlayDescriptor(OverlayDescriptor{
		OverlayType:	OverlayTypeService,
		PolicyHash:	HashParts("temporary-service-overlay"),
		Membership:	OverlayMembershipServiceAdvertisement,
		Routing:	RoutingStrategyLowLatencyAdvisory,
		MinPeers:	2,
		MaxPeers:	16,
		Fanout:		4,
		QoSClass:	QoSClassServiceCall,
		ExpiresHeight:	30,
		Version:	2,
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
	salt := []byte("aetra-test-network")
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
	salt := []byte("aetra-test-network")
	source := signedNodeRecord(t, 0x3b, salt, 100, NodeRoleFull)
	left := signedNodeRecord(t, 0x3c, salt, 100, NodeRoleService)
	right := signedNodeRecord(t, 0x3d, salt, 100, NodeRoleService)
	left.ServicesSupported = []string{"state-sync"}
	right.ServicesSupported = []string{"execution-stream"}
	desc, err := NewOverlayDescriptor(OverlayDescriptor{
		OverlayType:	OverlayTypeService,
		PolicyHash:	HashParts("service-provider-routing"),
		Membership:	OverlayMembershipServiceAdvertisement,
		Routing:	RoutingStrategyServiceProvider,
		MinPeers:	2,
		MaxPeers:	8,
		Fanout:		2,
		QoSClass:	QoSClassServiceCall,
		Version:	1,
	})
	require.NoError(t, err)
	proofs := []OverlayMembershipProof{
		testOverlayMembershipProof(t, left, desc, MembershipProofServiceRegistration, OverlayMembershipModeServiceRegistry, 80),
		testOverlayMembershipProof(t, right, desc, MembershipProofServiceRegistration, OverlayMembershipModeServiceRegistry, 80),
	}
	msg, err := NewNetworkMessage(NetworkMessage{
		Layer:			LayerL3Application,
		Channel:		ChannelService,
		PayloadHash:		HashParts("service-route-payload"),
		PayloadSizeBytes:	512,
	})
	require.NoError(t, err)

	plan, err := BuildOverlayRoute(OverlayRoutingRequest{
		Message:		msg,
		SourceNodeID:		source.NodeID,
		CandidatePeers:		[]NodeRecord{left, right},
		MembershipProofs:	proofs,
		Graph:			RoutingGraph{OverlayID: desc.OverlayID, Version: 1},
		Hint:			RouteHint{ServiceID: "execution-stream"},
		CurrentHeight:		20,
	}, []OverlayDescriptor{desc})
	require.NoError(t, err)
	require.Equal(t, OverlayTypeService, plan.OverlayType)
	require.Equal(t, RoutingStrategyServiceProvider, plan.Strategy)
	require.Equal(t, right.NodeID, plan.TargetNodeIDs[0])
}

func TestOverlayRoutingConsensusSafetyRequiresCommittedRoutingTable(t *testing.T) {
	salt := []byte("aetra-test-network")
	source := signedNodeRecord(t, 0x3e, salt, 100, NodeRoleZoneExecution)
	slow := signedNodeRecord(t, 0x3f, salt, 100, NodeRoleZoneExecution)
	fast := signedNodeRecord(t, 0x40, salt, 100, NodeRoleZoneExecution)
	desc, err := NewOverlayDescriptor(OverlayDescriptor{
		OverlayType:	OverlayTypeExecution,
		PolicyHash:	HashParts("committed-latency-routing"),
		Membership:	OverlayMembershipExecutionRole,
		Routing:	RoutingStrategyShortestLatencyPath,
		MinPeers:	2,
		MaxPeers:	8,
		Fanout:		2,
		QoSClass:	QoSClassExecutionMessage,
		Version:	1,
	})
	require.NoError(t, err)
	proofs := []OverlayMembershipProof{
		testOverlayMembershipProof(t, slow, desc, MembershipProofZoneAssignment, OverlayMembershipModeZoneAssignment, 80),
		testOverlayMembershipProof(t, fast, desc, MembershipProofZoneAssignment, OverlayMembershipModeZoneAssignment, 80),
	}
	msg, err := NewNetworkMessage(NetworkMessage{
		Layer:			LayerL2Overlay,
		Channel:		ChannelExecution,
		ConsensusEffect:	true,
		DeterminismSource:	DeterminismCommittedState,
		PayloadHash:		HashParts("execution-route-payload"),
		PayloadSizeBytes:	512,
	})
	require.NoError(t, err)
	graph := RoutingGraph{
		OverlayID:	desc.OverlayID,
		Version:	1,
		Edges: []RoutingEdge{
			{FromNodeID: source.NodeID, ToNodeID: slow.NodeID, LatencyMillis: 90, Priority: 2},
			{FromNodeID: source.NodeID, ToNodeID: fast.NodeID, LatencyMillis: 10, Priority: 1},
		},
	}

	_, err = BuildOverlayRoute(OverlayRoutingRequest{
		Message:		msg,
		SourceNodeID:		source.NodeID,
		CandidatePeers:		[]NodeRecord{slow, fast},
		MembershipProofs:	proofs,
		Graph:			graph,
		CurrentHeight:		20,
	}, []OverlayDescriptor{desc})
	require.ErrorContains(t, err, "committed routing")

	graph.Committed = true
	plan, err := BuildOverlayRoute(OverlayRoutingRequest{
		Message:		msg,
		SourceNodeID:		source.NodeID,
		CandidatePeers:		[]NodeRecord{slow, fast},
		MembershipProofs:	proofs,
		Graph:			graph,
		CurrentHeight:		20,
	}, []OverlayDescriptor{desc})
	require.NoError(t, err)
	require.True(t, plan.UsesCommittedRoutingTable)
	require.False(t, plan.UsesNodeLocalAdaptation)
	require.Equal(t, fast.NodeID, plan.TargetNodeIDs[0])
}

func TestAdaptiveOverlayGraphBuildsPeerSetsAndPreservesGlobalDiversity(t *testing.T) {
	salt := []byte("aetra-test-network")
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
		OverlayID:	desc.OverlayID,
		LocalNodeID:	HashParts("local-adaptive-node"),
		RoutingEpoch:	1,
		RandomSet:	[]AdaptivePeer{peerA, peerB},
		FallbackSet:	[]AdaptivePeer{peerA},
		FastSet:	[]AdaptivePeer{peerA},
		StableSet:	[]AdaptivePeer{peerB},
		ZoneSet:	[]AdaptivePeer{peerA, peerB},
		ServiceSet:	[]AdaptivePeer{},
		PolicyHash:	HashParts("bad-adaptive-policy"),
	}
	require.ErrorContains(t, graph.Validate(desc), "zone peers")

	graph.ZoneSet = nil
	graph.RandomSet = nil
	require.ErrorContains(t, graph.Validate(desc), "eclipse")
}

func TestPeerScoreDecayIsBoundedByPolicy(t *testing.T) {
	decayed, err := DecayPeerScore(PeerScore{ScoreBps: 9_000}, 3, PeerScoreDecayPolicy{
		MaxDecayBpsPerEpoch:	1_000,
		MinScoreBps:		4_000,
	})
	require.NoError(t, err)
	require.Equal(t, uint32(6_000), decayed.ScoreBps)

	decayed, err = DecayPeerScore(PeerScore{ScoreBps: 4_500}, 5, PeerScoreDecayPolicy{
		MaxDecayBpsPerEpoch:	1_000,
		MinScoreBps:		4_000,
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
		RoutingEpoch:	3,
		OverlayRoots: []OverlayRouteRoot{
			{OverlayID: desc.OverlayID, RootHash: HashParts("service-overlay-root")},
		},
		ZoneRouteRoot:		HashParts("zone-route-root"),
		ServiceRouteRoot:	HashParts("service-route-root"),
		PeerClassRoot:		HashParts("peer-class-root"),
		CongestionSnapshotRoot:	HashParts("congestion-snapshot-root"),
		PolicyHash:		desc.PolicyHash,
	})
	require.NoError(t, err)
	require.NoError(t, commitment.Validate())
	require.NotEmpty(t, ComputeRoutingTableCommitmentHash(commitment))

	require.NoError(t, ValidateRoutingTableUse(RoutingTableUse{
		Commitment:			commitment,
		UsedForPhysicalForwarding:	true,
	}))
	require.ErrorContains(t, ValidateRoutingTableUse(RoutingTableUse{
		Commitment:			commitment,
		UsedForExecutionScheduling:	true,
	}), "execution scheduling")
	require.NoError(t, ValidateRoutingTableUse(RoutingTableUse{
		Commitment:			commitment,
		Committed:			true,
		UsedForExecutionScheduling:	true,
	}))

	tampered := commitment
	tampered.OverlayRoots[0].RootHash = "not-a-hash"
	require.ErrorContains(t, tampered.Validate(), "lowercase hex")
}

func TestOverlayMembershipManagerRegistersMembersCanonically(t *testing.T) {
	salt := []byte("aetra-test-network")
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
	salt := []byte("aetra-test-network")
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
	salt := []byte("aetra-test-network")
	source := signedNodeRecord(t, 0x4e, salt, 100, NodeRoleFull)
	service := signedNodeRecord(t, 0x4f, salt, 100, NodeRoleService)
	fallbackA := signedNodeRecord(t, 0x50, salt, 100, NodeRoleFull)
	fallbackB := signedNodeRecord(t, 0x53, salt, 100, NodeRoleFull)
	fallbackC := signedNodeRecord(t, 0x54, salt, 100, NodeRoleFull)
	desc, err := NewOverlayDescriptor(OverlayDescriptor{
		OverlayType:	OverlayTypeService,
		PolicyHash:	HashParts("partitioned-service-overlay"),
		Membership:	OverlayMembershipServiceAdvertisement,
		Routing:	RoutingStrategyServiceProvider,
		MinPeers:	3,
		MaxPeers:	8,
		Fanout:		3,
		QoSClass:	QoSClassServiceCall,
		Version:	1,
	})
	require.NoError(t, err)
	msg, err := NewNetworkMessage(NetworkMessage{
		Layer:			LayerL3Application,
		Channel:		ChannelService,
		PayloadHash:		HashParts("partitioned-service-payload"),
		PayloadSizeBytes:	512,
	})
	require.NoError(t, err)
	fallbackGraph, err := BuildAdaptiveOverlayGraph(desc, source.NodeID, []AdaptivePeer{
		testGlobalAdaptivePeer(t, fallbackA, 7_000, 80, 8_000),
		testGlobalAdaptivePeer(t, fallbackB, 6_500, 90, 8_500),
		testGlobalAdaptivePeer(t, fallbackC, 6_000, 110, 7_500),
	}, 11, HashParts("fallback-partition-policy"))
	require.NoError(t, err)

	plan, err := BuildOverlayRouteWithFallback(OverlayRoutingRequest{
		Message:		msg,
		SourceNodeID:		source.NodeID,
		CandidatePeers:		[]NodeRecord{service},
		MembershipProofs:	[]OverlayMembershipProof{testOverlayMembershipProof(t, service, desc, MembershipProofServiceRegistration, OverlayMembershipModeServiceRegistry, 80)},
		Graph:			RoutingGraph{OverlayID: desc.OverlayID, Version: 1},
		Hint:			RouteHint{ServiceID: "execution-stream"},
		CurrentHeight:		20,
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
		Message:	consensusMsg,
		SourceNodeID:	source.NodeID,
		CandidatePeers:	[]NodeRecord{service},
		Graph:		RoutingGraph{OverlayID: desc.OverlayID, Version: 1},
		CurrentHeight:	20,
	}, []OverlayDescriptor{desc}, fallbackGraph)
	require.Error(t, err)
}

func TestAdaptivePeerRotationBoundsChurnAndKeepsStablePeers(t *testing.T) {
	salt := []byte("aetra-test-network")
	local := signedNodeRecord(t, 0x55, salt, 100, NodeRoleService)
	desc := defaultOverlayByType(t, OverlayTypeService)
	stable := testAdaptivePeer(t, signedNodeRecord(t, 0x56, salt, 100, NodeRoleService), 9_000, 20, 9_900, true)
	stable.ZonesSupported = nil
	stable.Services = nil
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
		MeshMessageConsensus:	ChannelConsensus,
		MeshMessageTx:		ChannelMempool,
		MeshMessageExecution:	ChannelExecution,
		MeshMessageQuery:	ChannelService,
		MeshMessageService:	ChannelService,
		MeshMessageCrossZone:	ChannelExecution,
		MeshMessageStateSync:	ChannelStateSync,
		MeshMessageStorage:	ChannelData,
		MeshMessageRouting:	ChannelRouting,
	}

	require.Len(t, AetherMeshMessageTypes(), len(expected))
	for messageType, channel := range expected {
		require.True(t, IsAetherMeshMessageType(messageType))
		require.Equal(t, channel, channelForMeshMessageType(messageType))
	}
}

func TestAetherMeshMessageSignsAndRejectsTampering(t *testing.T) {
	salt := []byte("aetra-test-network")
	originKey := deterministicPrivateKey(0x5b)
	origin := signedNodeRecord(t, 0x5b, salt, 100, NodeRoleService)
	destination := signedNodeRecord(t, 0x5c, salt, 100, NodeRoleService)
	desc := defaultOverlayByType(t, OverlayTypeService)

	msg, err := SignAetherMeshMessage(AetherMeshMessage{
		Type:			MeshMessageService,
		Payload:		[]byte("service-call-payload"),
		Origin:			origin.NodeID,
		Destination:		destination.NodeID,
		Priority:		PriorityForChannel(ChannelService),
		TTL:			50,
		OverlayID:		desc.OverlayID,
		Sequence:		1,
		RouteHint:		RouteHint{ServiceID: "execution-stream"},
		DeadlineHeight:		70,
		ConsensusEffect:	false,
	}, originKey)
	require.NoError(t, err)
	require.NoError(t, VerifyAetherMeshMessageSignature(msg, originKey.Public().(ed25519.PublicKey), 20))

	again, err := NewAetherMeshMessage(AetherMeshMessage{
		Type:			MeshMessageService,
		Payload:		[]byte("service-call-payload"),
		Origin:			origin.NodeID,
		Destination:		destination.NodeID,
		Priority:		PriorityForChannel(ChannelService),
		TTL:			50,
		OverlayID:		desc.OverlayID,
		Sequence:		1,
		RouteHint:		RouteHint{ServiceID: "execution-stream"},
		DeadlineHeight:		70,
		ConsensusEffect:	false,
	})
	require.NoError(t, err)
	require.Equal(t, msg.MessageID, again.MessageID)

	tampered := msg
	tampered.Payload = []byte("tampered")
	require.ErrorContains(t, VerifyAetherMeshMessageSignature(tampered, originKey.Public().(ed25519.PublicKey), 20), "payload hash")
}

func TestAetherMeshCrossZoneAndConsensusProofRules(t *testing.T) {
	salt := []byte("aetra-test-network")
	origin := signedNodeRecord(t, 0x5d, salt, 100, NodeRoleZoneExecution)
	destination := signedNodeRecord(t, 0x5e, salt, 100, NodeRoleZoneExecution)
	desc := defaultOverlayByType(t, OverlayTypeExecution)

	_, err := NewAetherMeshMessage(AetherMeshMessage{
		Type:			MeshMessageCrossZone,
		Payload:		[]byte("cross-zone"),
		Origin:			origin.NodeID,
		Destination:		destination.NodeID,
		Priority:		PriorityForChannel(ChannelExecution),
		TTL:			25,
		OverlayID:		desc.OverlayID,
		SourceZone:		"APPLICATION_ZONE",
		DestinationZone:	"APPLICATION_ZONE",
		Sequence:		1,
	})
	require.ErrorContains(t, err, "different zones")

	_, err = NewAetherMeshMessage(AetherMeshMessage{
		Type:			MeshMessageService,
		Payload:		[]byte("consensus-service"),
		Origin:			origin.NodeID,
		Destination:		destination.NodeID,
		Priority:		PriorityForChannel(ChannelService),
		TTL:			25,
		OverlayID:		defaultOverlayByType(t, OverlayTypeService).OverlayID,
		Sequence:		1,
		ConsensusEffect:	true,
		DeterminismSource:	DeterminismCommittedState,
	})
	require.ErrorContains(t, err, "proof")

	msg, err := NewAetherMeshMessage(AetherMeshMessage{
		Type:			MeshMessageCrossZone,
		Payload:		[]byte("cross-zone"),
		Origin:			origin.NodeID,
		Destination:		destination.NodeID,
		Priority:		PriorityForChannel(ChannelExecution),
		TTL:			25,
		OverlayID:		desc.OverlayID,
		SourceZone:		"APPLICATION_ZONE",
		DestinationZone:	"FINANCIAL_ZONE",
		Sequence:		2,
		ConsensusEffect:	true,
		DeterminismSource:	DeterminismDeterministicProof,
		Proof: AetherMeshProof{
			ProofType:	"zone-commitment",
			ProofHash:	HashParts("cross-zone-proof"),
			ProofHeight:	20,
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
	salt := []byte("aetra-test-network")
	source := signedNodeRecord(t, 0x5f, salt, 100, NodeRoleFull)
	left := signedNodeRecord(t, 0x60, salt, 100, NodeRoleService)
	right := signedNodeRecord(t, 0x62, salt, 100, NodeRoleService)
	desc := defaultOverlayByType(t, OverlayTypeService)
	msg, err := NewAetherMeshMessage(AetherMeshMessage{
		Type:		MeshMessageService,
		Payload:	[]byte("mesh-service-route"),
		Origin:		source.NodeID,
		Destination:	right.NodeID,
		Priority:	PriorityForChannel(ChannelService),
		TTL:		30,
		OverlayID:	desc.OverlayID,
		Sequence:	1,
		RouteHint:	RouteHint{ServiceID: "execution-stream"},
		DeadlineHeight:	90,
	})
	require.NoError(t, err)

	delivery, err := RouteAetherMeshMessage(AetherMeshRouteRequest{
		Message:	msg,
		SourceNodeID:	source.NodeID,
		CandidatePeers:	[]NodeRecord{left, right},
		MembershipProofs: []OverlayMembershipProof{
			testOverlayMembershipProof(t, left, desc, MembershipProofServiceRegistration, OverlayMembershipModeServiceRegistry, 80),
			testOverlayMembershipProof(t, right, desc, MembershipProofServiceRegistration, OverlayMembershipModeServiceRegistry, 80),
		},
		Graph: RoutingGraph{
			OverlayID:	desc.OverlayID,
			Version:	1,
			Edges: []RoutingEdge{
				{FromNodeID: source.NodeID, ToNodeID: left.NodeID, LatencyMillis: 25, Priority: 1},
				{FromNodeID: source.NodeID, ToNodeID: right.NodeID, LatencyMillis: 15, Priority: 0},
			},
		},
		CurrentHeight:	20,
	}, []OverlayDescriptor{desc})
	require.NoError(t, err)
	require.Equal(t, ChannelService, delivery.Channel)
	require.Equal(t, desc.OverlayID, delivery.Route.OverlayID)
	require.False(t, delivery.Route.FallbackUsed)
	require.NotEmpty(t, delivery.Route.TargetNodeIDs)
}

func TestExecutionZoneMessageRequiresCommittedScheduleForConsensusOrder(t *testing.T) {
	salt := []byte("aetra-test-network")
	origin := signedNodeRecord(t, 0x63, salt, 100, NodeRoleZoneExecution)
	destination := signedNodeRecord(t, 0x64, salt, 100, NodeRoleZoneExecution)
	desc := defaultOverlayByType(t, OverlayTypeExecution)
	mesh, err := NewAetherMeshMessage(AetherMeshMessage{
		Type:			MeshMessageExecution,
		Payload:		[]byte("execution-zone-payload"),
		Origin:			origin.NodeID,
		Destination:		destination.NodeID,
		Priority:		PriorityForChannel(ChannelExecution),
		TTL:			40,
		OverlayID:		desc.OverlayID,
		DestinationZone:	"APPLICATION_ZONE",
		Sequence:		7,
		ConsensusEffect:	true,
		DeterminismSource:	DeterminismCommittedState,
		Proof: AetherMeshProof{
			ProofType:	"committed-schedule",
			ProofHash:	HashParts("execution-proof"),
			ProofHeight:	30,
		},
	})
	require.NoError(t, err)
	uncommitted, err := NewExecutionMessageSchedule(ExecutionMessageSchedule{
		ZoneID:			"APPLICATION_ZONE",
		ShardID:		"shard-1",
		RoutingClass:		ExecutionRoutingExecutionOverlay,
		Ordered:		true,
		MessageIDs:		[]string{mesh.MessageID},
		FirstZoneSequence:	7,
		LastZoneSequence:	7,
	})
	require.NoError(t, err)

	_, err = NewExecutionZoneMessage(ExecutionZoneMessage{
		Message:		mesh,
		RoutingClass:		ExecutionRoutingExecutionOverlay,
		ZoneID:			"APPLICATION_ZONE",
		ShardID:		"shard-1",
		ExecutionOverlayID:	desc.OverlayID,
		ZoneSequence:		7,
		NetworkDeliveryOrdinal:	99,
		ConsensusScheduleOrder:	1,
	}, uncommitted)
	require.ErrorContains(t, err, "committed schedule")

	committed, err := NewExecutionMessageSchedule(ExecutionMessageSchedule{
		ZoneID:			"APPLICATION_ZONE",
		ShardID:		"shard-1",
		RoutingClass:		ExecutionRoutingExecutionOverlay,
		Committed:		true,
		Ordered:		true,
		MessageIDs:		[]string{mesh.MessageID},
		FirstZoneSequence:	7,
		LastZoneSequence:	7,
	})
	require.NoError(t, err)
	executionMsg, err := NewExecutionZoneMessage(ExecutionZoneMessage{
		Message:		mesh,
		RoutingClass:		ExecutionRoutingExecutionOverlay,
		ZoneID:			"APPLICATION_ZONE",
		ShardID:		"shard-1",
		ExecutionOverlayID:	desc.OverlayID,
		ZoneSequence:		7,
		NetworkDeliveryOrdinal:	99,
		ConsensusScheduleOrder:	1,
	}, committed)
	require.NoError(t, err)
	require.NotEqual(t, executionMsg.NetworkDeliveryOrdinal, executionMsg.ConsensusScheduleOrder)
	require.Equal(t, committed.ScheduleID, executionMsg.ConsensusScheduleID)
}

func TestExecutionZoneMessageSupportsAsyncParallelBlockSTMGroups(t *testing.T) {
	salt := []byte("aetra-test-network")
	origin := signedNodeRecord(t, 0x65, salt, 100, NodeRoleZoneExecution)
	destination := signedNodeRecord(t, 0x66, salt, 100, NodeRoleZoneExecution)
	desc := defaultOverlayByType(t, OverlayTypeExecution)
	mesh, err := NewAetherMeshMessage(AetherMeshMessage{
		Type:			MeshMessageExecution,
		Payload:		[]byte("async-execution"),
		Origin:			origin.NodeID,
		Destination:		destination.NodeID,
		Priority:		PriorityForChannel(ChannelExecution),
		TTL:			40,
		OverlayID:		desc.OverlayID,
		DestinationZone:	"APPLICATION_ZONE",
		Sequence:		8,
	})
	require.NoError(t, err)

	_, err = NewExecutionZoneMessage(ExecutionZoneMessage{
		Message:		mesh,
		RoutingClass:		ExecutionRoutingShard,
		ZoneID:			"APPLICATION_ZONE",
		ShardID:		"shard-2",
		ExecutionOverlayID:	desc.OverlayID,
		ExecutionGroupID:	HashParts("async-group"),
		ZoneSequence:		8,
		Async:			true,
		ParallelZoneExecution:	true,
	}, ExecutionMessageSchedule{})
	require.ErrorContains(t, err, "BlockSTM")

	executionMsg, err := NewExecutionZoneMessage(ExecutionZoneMessage{
		Message:		mesh,
		RoutingClass:		ExecutionRoutingShard,
		ZoneID:			"APPLICATION_ZONE",
		ShardID:		"shard-2",
		ExecutionOverlayID:	desc.OverlayID,
		ExecutionGroupID:	HashParts("async-group"),
		BlockSTMGroupID:	HashParts("blockstm-group"),
		ZoneSequence:		8,
		Async:			true,
		ParallelZoneExecution:	true,
	}, ExecutionMessageSchedule{})
	require.NoError(t, err)
	require.True(t, executionMsg.Async)
	require.True(t, executionMsg.ParallelZoneExecution)
}

func TestCrossZoneMessageRequiresSequenceExpiryAndReplayProtection(t *testing.T) {
	msg := CrossZoneMessage{
		SourceZone:		"APPLICATION_ZONE",
		DestinationZone:	"FINANCIAL_ZONE",
		MessageHash:		HashParts("cross-zone-message"),
		ExpiryHeight:		90,
		ReceiptPolicy:		ReceiptPolicyOnExecution,
		ProofRequired:		true,
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
		SourceZone:		"APPLICATION_ZONE",
		DestinationZone:	"FINANCIAL_ZONE",
		SourceSequence:		11,
		MessageHash:		HashParts("cross-zone-message"),
		Status:			CrossZoneReceiptExecuted,
		ReceiptPolicy:		ReceiptPolicyAlways,
		ProofHash:		HashParts("cross-zone-receipt-proof"),
		ReceiptHeight:		35,
		RollbackSafe:		true,
		ProofQueryable:		true,
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

func TestOverlayMessageQueueOrdersMessagesAndRejectsReplay(t *testing.T) {
	salt := []byte("aetra-test-network")
	origin := signedNodeRecord(t, 0x67, salt, 100, NodeRoleService)
	destination := signedNodeRecord(t, 0x68, salt, 100, NodeRoleService)
	desc := defaultOverlayByType(t, OverlayTypeService)
	queue, err := NewOverlayMessageQueue(desc.OverlayID, 8)
	require.NoError(t, err)
	lowPriority := testMeshMessage(t, MeshMessageService, desc.OverlayID, origin.NodeID, destination.NodeID, 5, 2)
	highPriority := testMeshMessage(t, MeshMessageService, desc.OverlayID, origin.NodeID, destination.NodeID, 1, 1)

	queue, err = EnqueueOverlayMessage(queue, lowPriority, 20)
	require.NoError(t, err)
	queue, err = EnqueueOverlayMessage(queue, highPriority, 20)
	require.NoError(t, err)
	_, err = EnqueueOverlayMessage(queue, highPriority, 20)
	require.ErrorContains(t, err, "replay")

	next, messages, err := DequeueOverlayMessages(queue, 1)
	require.NoError(t, err)
	require.Len(t, messages, 1)
	require.Equal(t, highPriority.MessageID, messages[0].MessageID)
	require.Len(t, next.Messages, 1)
	require.Contains(t, next.SeenMessageIDs, highPriority.MessageID)
}

func TestCrossZoneSequenceTrackerAssignsAndRejectsReplayAndGaps(t *testing.T) {
	tracker := CrossZoneSequenceTracker{}
	var seq uint64
	var err error
	tracker, seq, err = NextCrossZoneSequence(tracker, "APPLICATION_ZONE", "FINANCIAL_ZONE")
	require.NoError(t, err)
	require.Equal(t, uint64(1), seq)
	tracker, seq, err = NextCrossZoneSequence(tracker, "APPLICATION_ZONE", "FINANCIAL_ZONE")
	require.NoError(t, err)
	require.Equal(t, uint64(2), seq)

	first, err := NewCrossZoneMessage(CrossZoneMessage{
		SourceZone:		"APPLICATION_ZONE",
		DestinationZone:	"FINANCIAL_ZONE",
		SourceSequence:		1,
		MessageHash:		HashParts("cz-sequence-1"),
		ExpiryHeight:		100,
		ReceiptPolicy:		ReceiptPolicyAlways,
	})
	require.NoError(t, err)
	tracker, err = AcceptCrossZoneSequence(tracker, first, true, 20)
	require.NoError(t, err)
	_, err = AcceptCrossZoneSequence(tracker, first, true, 21)
	require.ErrorContains(t, err, "replay")

	gap, err := NewCrossZoneMessage(CrossZoneMessage{
		SourceZone:		"APPLICATION_ZONE",
		DestinationZone:	"FINANCIAL_ZONE",
		SourceSequence:		3,
		MessageHash:		HashParts("cz-sequence-3"),
		ExpiryHeight:		100,
		ReceiptPolicy:		ReceiptPolicyAlways,
	})
	require.NoError(t, err)
	_, err = AcceptCrossZoneSequence(tracker, gap, true, 22)
	require.ErrorContains(t, err, "gap")
	_, err = AcceptCrossZoneSequence(tracker, gap, false, 22)
	require.NoError(t, err)
}

func TestReceiptDeliveryProtocolAcknowledgesAndFeedsMetrics(t *testing.T) {
	receipt, err := NewCrossZoneReceipt(CrossZoneReceipt{
		SourceZone:		"APPLICATION_ZONE",
		DestinationZone:	"FINANCIAL_ZONE",
		SourceSequence:		12,
		MessageHash:		HashParts("receipt-delivery-message"),
		Status:			CrossZoneReceiptDelivered,
		ReceiptPolicy:		ReceiptPolicyOnDelivery,
		ProofHash:		HashParts("receipt-delivery-proof"),
		ReceiptHeight:		44,
		RollbackSafe:		true,
		ProofQueryable:		true,
	})
	require.NoError(t, err)
	delivery, err := NewReceiptDelivery(receipt, HashParts("receipt-destination-node"), 45)
	require.NoError(t, err)
	require.Equal(t, ReceiptDeliveryPending, delivery.State)
	acked, err := AckReceiptDelivery(delivery, HashParts("receipt-ack"))
	require.NoError(t, err)
	require.Equal(t, ReceiptDeliveryAcknowledged, acked.State)

	metrics, err := EvaluateL3Metrics(nil, []ReceiptDelivery{acked}, nil, nil, nil)
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	require.Equal(t, uint64(1), metrics[0].ReceiptCount)
	require.Equal(t, uint64(1), metrics[0].DeliveredCount)
}

func TestQueryResponseProofAttachmentIsRequiredAndMetriced(t *testing.T) {
	_, err := NewQueryResponseProof(QueryResponseProof{
		RequestID:	HashParts("query-request"),
		Responder:	HashParts("query-responder"),
		PayloadHash:	HashParts("query-payload"),
		ResponseHeight:	50,
	})
	require.ErrorContains(t, err, "proof")

	response, err := NewQueryResponseProof(QueryResponseProof{
		RequestID:	HashParts("query-request"),
		Responder:	HashParts("query-responder"),
		PayloadHash:	HashParts("query-payload"),
		Proof: AetherMeshProof{
			ProofType:	"iavl-query-proof",
			ProofHash:	HashParts("query-proof"),
			ProofHeight:	49,
		},
		ResponseHeight:	50,
	})
	require.NoError(t, err)
	require.Equal(t, ComputeQueryResponseID(response), response.ResponseID)
	metrics, err := EvaluateL3Metrics(nil, nil, []QueryResponseProof{response}, nil, nil)
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	require.Equal(t, uint64(1), metrics[0].QueryProofCount)
}

func TestL3MetricsAccountQueuesReplayDropsAndExpiry(t *testing.T) {
	salt := []byte("aetra-test-network")
	origin := signedNodeRecord(t, 0x69, salt, 100, NodeRoleService)
	destination := signedNodeRecord(t, 0x6a, salt, 100, NodeRoleService)
	desc := defaultOverlayByType(t, OverlayTypeService)
	queue, err := NewOverlayMessageQueue(desc.OverlayID, 8)
	require.NoError(t, err)
	msg := testMeshMessage(t, MeshMessageService, desc.OverlayID, origin.NodeID, destination.NodeID, 5, 3)
	queue, err = EnqueueOverlayMessage(queue, msg, 20)
	require.NoError(t, err)

	metrics, err := EvaluateL3Metrics([]OverlayMessageQueue{queue}, nil, nil, map[string]uint64{desc.OverlayID: 2}, map[string]uint64{desc.OverlayID: 1})
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	require.Equal(t, desc.OverlayID, metrics[0].OverlayID)
	require.Equal(t, uint64(1), metrics[0].QueuedCount)
	require.Equal(t, uint64(2), metrics[0].ReplayDropCount)
	require.Equal(t, uint64(1), metrics[0].ExpiredCount)
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
		Channel:	ChannelConsensus,
		SizeBytes:	128,
		EnqueuedHeight:	100,
		Sequence:	1,
		PayloadHash:	HashParts("consensus-vote"),
	}
	plan, err := PlanPropagation(adapter, consensus, 20, PeerScore{ScoreBps: BasisPoints})
	require.NoError(t, err)
	require.True(t, plan.HandledByCometBFT)
	require.Zero(t, plan.AdapterFanout)
	require.False(t, plan.UsesAdvisoryPeerMetric)

	service := TransportEnvelope{
		Channel:	ChannelService,
		SizeBytes:	512,
		EnqueuedHeight:	100,
		Sequence:	2,
		PayloadHash:	HashParts("service-message"),
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
		Layer:			LayerL3Application,
		Channel:		ChannelService,
		ConsensusEffect:	true,
		DeterminismSource:	DeterminismDeterministicProof,
		PayloadHash:		HashParts("payload"),
		PayloadSizeBytes:	128,
		CommittedProofHash:	proofHash,
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
		Layer:			LayerL3Application,
		Channel:		ChannelData,
		PayloadHash:		HashParts("large-payload"),
		PayloadSizeBytes:	LargePayloadBytes + 1,
	}

	require.ErrorContains(t, msg.ValidateHardRules(), "large payloads")

	msg.Chunked = true
	msg.CommitmentVerified = true
	require.NoError(t, msg.ValidateHardRules())
}

func TestDiscoveryRecordMustBeSignedExpiringAndProofChecked(t *testing.T) {
	salt := []byte("aetra-test-network")
	record := signedNodeRecord(t, 0x51, salt, 100, NodeRoleStateSync)
	discovery := DiscoveryRecord{
		Record:		record,
		ProofHash:	HashParts("optional-discovery-proof"),
		ProofHeight:	90,
	}

	require.NoError(t, discovery.Validate(salt, 95))

	discovery.ProofHeight = 101
	require.ErrorContains(t, discovery.Validate(salt, 0), "outlive")

	discovery = DiscoveryRecord{Record: record}
	discovery.Record.Signature[0] ^= 0xff
	require.ErrorContains(t, discovery.Validate(salt, 95), "signature")
}

func TestDistributedRoutingTableIndexesLeaseBasedDiscoveryObjects(t *testing.T) {
	salt := []byte("aetra-test-network")
	overlay := defaultOverlayByType(t, OverlayTypeService)
	serviceLow := signedNodeRecordWithCapabilities(t, 0x52, salt, 100, []NodeRole{NodeRoleService}, nil, []string{"svc.payments"})
	serviceHigh := signedNodeRecordWithCapabilities(t, 0x53, salt, 100, []NodeRole{NodeRoleService}, nil, []string{"svc.payments"})
	zone := signedNodeRecordWithCapabilities(t, 0x54, salt, 100, []NodeRole{NodeRoleZoneExecution}, []string{"zone-a"}, nil)
	storage := signedNodeRecord(t, 0x55, salt, 100, NodeRoleStorageProvider)
	routing := signedNodeRecord(t, 0x56, salt, 100, NodeRoleRouting)
	full := signedNodeRecord(t, 0x57, salt, 100, NodeRoleFull)

	table := EmptyDistributedRoutingTable()
	ads := []DRTAdvertisement{
		testDRTAdvertisement(DRTObjectNode, full, "", "", "", "", 1, 5_000, 10, 90),
		testDRTAdvertisement(DRTObjectExecutionZone, zone, "", "zone-a", "", "", 10, 7_000, 10, 90),
		testDRTAdvertisement(DRTObjectServiceEndpoint, serviceLow, overlay.OverlayID, "", "svc.payments", HashParts("endpoint", "svc-low"), 100, 8_000, 10, 90),
		testDRTAdvertisement(DRTObjectServiceEndpoint, serviceHigh, overlay.OverlayID, "", "svc.payments", HashParts("endpoint", "svc-high"), 1_000, 7_000, 10, 90),
		testDRTAdvertisement(DRTObjectRPCEndpoint, full, "", "", "", HashParts("endpoint", "rpc"), 50, 6_000, 10, 90),
		testDRTAdvertisement(DRTObjectStorageProvider, storage, "", "", "", HashParts("endpoint", "storage"), 500, 8_500, 10, 90),
		testDRTAdvertisement(DRTObjectRoutingEntryPoint, routing, "", "", "", HashParts("endpoint", "routing"), 250, 9_000, 10, 90),
		testDRTAdvertisement(DRTObjectOverlayMembershipRecord, serviceHigh, overlay.OverlayID, "", "svc.payments", "", 1_000, 7_000, 10, 90),
		testDRTAdvertisement(DRTObjectStreamProvider, full, "", "", "", HashParts("endpoint", "stream"), 25, 5_000, 10, 90),
	}

	var err error
	for _, ad := range ads {
		table, err = table.Add(ad, salt, 20)
		require.NoError(t, err)
	}
	require.Len(t, table.Advertisements, len(ads))
	require.NoError(t, table.Validate(salt, 20))

	serviceResults := table.Query(DRTQuery{
		ObjectType:	DRTObjectServiceEndpoint,
		OverlayID:	overlay.OverlayID,
		ServiceID:	"svc.payments",
		CurrentHeight:	20,
	})
	require.Len(t, serviceResults, 2)
	require.Equal(t, serviceHigh.NodeID, serviceResults[0].Discovery.Record.NodeID)
	require.Greater(t, serviceResults[0].StakeWeight, serviceResults[1].StakeWeight)

	require.Len(t, table.Query(DRTQuery{ObjectType: DRTObjectExecutionZone, ZoneID: "zone-a", CurrentHeight: 20}), 1)
	require.Len(t, table.Query(DRTQuery{ObjectType: DRTObjectStorageProvider, MinStakeWeight: 100, CurrentHeight: 20}), 1)
	require.Len(t, table.Query(DRTQuery{ObjectType: DRTObjectStreamProvider, CurrentHeight: 20}), 1)

	root, err := ComputeDRTIndexRoot(table.Advertisements)
	require.NoError(t, err)
	require.NoError(t, ValidateHash("networking DRT test root", root))

	pruned := table.Prune(91)
	require.Empty(t, pruned.Advertisements)
}

func TestDRTRejectsExpiredTamperedAndRoleMismatchedAdvertisements(t *testing.T) {
	salt := []byte("aetra-test-network")
	service := signedNodeRecordWithCapabilities(t, 0x58, salt, 100, []NodeRole{NodeRoleService}, nil, []string{"svc.search"})
	full := signedNodeRecord(t, 0x59, salt, 100, NodeRoleFull)
	table := EmptyDistributedRoutingTable()

	expired := testDRTAdvertisement(DRTObjectServiceEndpoint, service, "", "", "svc.search", HashParts("endpoint", "expired"), 1, 1_000, 10, 20)
	_, err := table.Add(expired, salt, 21)
	require.ErrorContains(t, err, "expired")

	outliving := testDRTAdvertisement(DRTObjectServiceEndpoint, service, "", "", "svc.search", HashParts("endpoint", "outlive"), 1, 1_000, 10, 101)
	_, err = table.Add(outliving, salt, 20)
	require.ErrorContains(t, err, "outlive")

	tampered := testDRTAdvertisement(DRTObjectServiceEndpoint, service, "", "", "svc.search", HashParts("endpoint", "tamper"), 1, 1_000, 10, 90)
	tampered.Discovery.Record.Signature[0] ^= 0xff
	_, err = table.Add(tampered, salt, 20)
	require.ErrorContains(t, err, "signature")

	wrongRole := testDRTAdvertisement(DRTObjectStorageProvider, full, "", "", "", HashParts("endpoint", "storage"), 1, 1_000, 10, 90)
	_, err = table.Add(wrongRole, salt, 20)
	require.ErrorContains(t, err, "storage provider role")

	missingEndpoint := testDRTAdvertisement(DRTObjectServiceEndpoint, service, "", "", "svc.search", "", 1, 1_000, 10, 90)
	_, err = table.Add(missingEndpoint, salt, 20)
	require.ErrorContains(t, err, "endpoint hash")
}

func TestDRTBucketsAndOverlayNativeDiscoveryAreDeterministic(t *testing.T) {
	salt := []byte("aetra-test-network")
	overlay := defaultOverlayByType(t, OverlayTypeRouting)
	local := signedNodeRecord(t, 0x5a, salt, 100, NodeRoleFull)
	left := signedNodeRecord(t, 0x5b, salt, 100, NodeRoleRouting)
	right := signedNodeRecord(t, 0x5c, salt, 100, NodeRoleRouting)
	table := EmptyDistributedRoutingTable()

	var err error
	for _, record := range []NodeRecord{left, right} {
		table, err = table.Add(testDRTAdvertisement(DRTObjectRoutingEntryPoint, record, overlay.OverlayID, "", "", HashParts("endpoint", record.NodeID), 1_000, 8_000, 10, 90), salt, 20)
		require.NoError(t, err)
	}

	overlayResults := table.Query(DRTQuery{ObjectType: DRTObjectRoutingEntryPoint, OverlayID: overlay.OverlayID, CurrentHeight: 20})
	require.Len(t, overlayResults, 2)

	buckets, err := table.Buckets(local.NodeID, DRTObjectRoutingEntryPoint, 8, 20)
	require.NoError(t, err)
	sameBuckets, err := table.Buckets(local.NodeID, DRTObjectRoutingEntryPoint, 8, 20)
	require.NoError(t, err)
	require.Equal(t, buckets, sameBuckets)

	total := 0
	for _, bucket := range buckets {
		total += len(bucket.Advertisements)
	}
	require.Equal(t, 2, total)
}

func TestSignedDiscoveryRecordStoreFindRenewAndRevoke(t *testing.T) {
	salt := []byte("aetra-test-network")
	owner := signedNodeRecordWithCapabilities(t, 0x5d, salt, 120, []NodeRole{NodeRoleService}, nil, []string{"svc.payments"})
	targetID := HashParts("discovery-target", "svc.payments")
	endpointHash := HashParts("endpoint", "svc.payments.primary")
	record := testSignedDiscoveryObjectRecord(t, owner, 0x5d, salt, DRTObjectServiceEndpoint, targetID, endpointHash, "", "svc.payments", "", 80)

	require.Equal(t, ComputeDiscoveryRecordID(record), record.RecordID)
	require.NoError(t, record.Validate(salt, 20))

	tampered := record
	tampered.Signature = cloneBytes(record.Signature)
	tampered.Signature[0] ^= 0xff
	require.ErrorContains(t, tampered.Validate(salt, 20), "signature")

	table := EmptyDistributedRoutingTable()
	table, err := table.Store(record, salt, 20)
	require.NoError(t, err)
	require.Len(t, table.FindService("svc.payments", 20), 1)
	require.Len(t, table.FindEndpoint(endpointHash, 20), 1)
	require.Empty(t, table.FindService("svc.payments", 81))

	renewed, err := RenewDiscoveryRecord(record, 100, deterministicPrivateKey(0x5d), salt)
	require.NoError(t, err)
	table, err = table.UpdateLease(renewed, salt, 30)
	require.NoError(t, err)
	require.Len(t, table.Records, 1)
	require.Equal(t, uint64(100), table.FindService("svc.payments", 90)[0].ExpiresHeight)

	_, err = table.UpdateLease(record, salt, 31)
	require.ErrorContains(t, err, "extend expiry")

	revocation, err := NewDiscoveryRevocation(renewed, deterministicPrivateKey(0x5d), 95)
	require.NoError(t, err)
	table, err = table.Revoke(revocation, salt, 95)
	require.NoError(t, err)
	require.Empty(t, table.FindService("svc.payments", 95))
	require.Len(t, table.Revocations, 1)

	_, err = table.Store(renewed, salt, 96)
	require.ErrorContains(t, err, "revoked")
}

func TestDiscoveryQueryOperationsCoverNodeZoneStorageAndEndpoint(t *testing.T) {
	salt := []byte("aetra-test-network")
	node := signedNodeRecord(t, 0x5e, salt, 100, NodeRoleFull)
	zone := signedNodeRecordWithCapabilities(t, 0x5f, salt, 100, []NodeRole{NodeRoleZoneExecution}, []string{"zone-b"}, nil)
	storage := signedNodeRecord(t, 0x60, salt, 100, NodeRoleStorageProvider)
	table := EmptyDistributedRoutingTable()

	nodeRecord := testSignedDiscoveryObjectRecord(t, node, 0x5e, salt, DRTObjectNode, node.NodeID, HashParts("advertisement", "node"), "", "", "", 80)
	zoneRecord := testSignedDiscoveryObjectRecord(t, zone, 0x5f, salt, DRTObjectExecutionZone, HashParts("target", "zone-b"), HashParts("advertisement", "zone-b"), "zone-b", "", "", 80)
	storageRecord := testSignedDiscoveryObjectRecord(t, storage, 0x60, salt, DRTObjectStorageProvider, storage.NodeID, HashParts("endpoint", "storage"), "", "", "", 80)

	var err error
	for _, record := range []DiscoveryRecord{nodeRecord, zoneRecord, storageRecord} {
		table, err = table.Store(record, salt, 20)
		require.NoError(t, err)
	}

	require.Len(t, table.FindNode(node.NodeID, 20), 1)
	require.Len(t, table.FindZone("zone-b", 20), 1)
	require.Len(t, table.FindStorageProvider(20), 1)
	require.Len(t, table.FindEndpoint(HashParts("endpoint", "storage"), 20), 1)

	_, err = NewSignedDiscoveryRecord(DiscoveryRecord{
		RecordType:		DRTObjectStorageProvider,
		OwnerNodeID:		node.NodeID,
		TargetID:		node.NodeID,
		AdvertisementHash:	HashParts("endpoint", "bad-storage"),
		ExpiresHeight:		80,
		Record:			node,
	}, deterministicPrivateKey(0x5e), salt)
	require.ErrorContains(t, err, "storage provider role")
}

func TestDiscoveryResponseSignsResultsAndProofAttachment(t *testing.T) {
	salt := []byte("aetra-test-network")
	source := signedNodeRecord(t, 0x62, salt, 120, NodeRoleRouting)
	service := signedNodeRecordWithCapabilities(t, 0x63, salt, 120, []NodeRole{NodeRoleService}, nil, []string{"svc.search"})
	record := testSignedDiscoveryObjectRecord(t, service, 0x63, salt, DRTObjectServiceEndpoint, HashParts("target", "svc.search"), HashParts("endpoint", "svc.search"), "", "svc.search", "", 90)
	table := EmptyDistributedRoutingTable()
	var err error
	table, err = table.Store(record, salt, 20)
	require.NoError(t, err)

	query := DRTQuery{ObjectType: DRTObjectServiceEndpoint, ServiceID: "svc.search", CurrentHeight: 20}
	advisory, err := BuildDiscoveryResponse(table, query, source, deterministicPrivateKey(0x62), salt, DiscoveryOnChainProof{}, 20)
	require.NoError(t, err)
	require.True(t, advisory.AdvisoryOnly)
	require.Len(t, advisory.MatchedRecords, 1)
	require.Equal(t, ComputeDiscoveryResponseID(advisory), advisory.ResponseID)
	require.NoError(t, advisory.Validate(source.NodePubKey, salt, 20))

	resultHash, err := ComputeDiscoveryResponseResultHash(advisory.MatchedRecords)
	require.NoError(t, err)
	stateRoot := HashParts("state-root", "services")
	proof := DiscoveryOnChainProof{
		ProofHeight:	20,
		StateRoot:	stateRoot,
		ProofHash:	ComputeDiscoveryOnChainProofHash(resultHash, stateRoot, 20),
	}
	proven, err := BuildDiscoveryResponse(table, query, source, deterministicPrivateKey(0x62), salt, proof, 20)
	require.NoError(t, err)
	require.False(t, proven.AdvisoryOnly)
	require.NoError(t, proven.Validate(source.NodePubKey, salt, 20))

	forged := proven
	forged.SourceSignature = cloneBytes(proven.SourceSignature)
	forged.SourceSignature[0] ^= 0xff
	require.ErrorContains(t, forged.Validate(source.NodePubKey, salt, 20), "source signature")

	badProof := proven
	badProof.OnChainProof.ProofHash = HashParts("wrong-proof")
	badProof.ResponseID = ComputeDiscoveryResponseID(badProof)
	payload, err := DiscoveryResponseSigningPayload(badProof)
	require.NoError(t, err)
	badProof.SourceSignature = ed25519.Sign(deterministicPrivateKey(0x62), payload)
	require.ErrorContains(t, badProof.Validate(source.NodePubKey, salt, 20), "proof mismatch")
}

func TestDiscoveryResponseRejectsExpiredForgedAndReplayedRecords(t *testing.T) {
	salt := []byte("aetra-test-network")
	source := signedNodeRecord(t, 0x64, salt, 120, NodeRoleRouting)
	service := signedNodeRecordWithCapabilities(t, 0x65, salt, 120, []NodeRole{NodeRoleService}, nil, []string{"svc.replay"})
	record := testSignedDiscoveryObjectRecord(t, service, 0x65, salt, DRTObjectServiceEndpoint, HashParts("target", "svc.replay"), HashParts("endpoint", "svc.replay"), "", "svc.replay", "", 40)
	table := EmptyDistributedRoutingTable()
	var err error
	table, err = table.Store(record, salt, 20)
	require.NoError(t, err)

	response, err := BuildDiscoveryResponse(table, DRTQuery{ObjectType: DRTObjectServiceEndpoint, ServiceID: "svc.replay", CurrentHeight: 20}, source, deterministicPrivateKey(0x64), salt, DiscoveryOnChainProof{}, 20)
	require.NoError(t, err)
	require.NoError(t, response.Validate(source.NodePubKey, salt, 39))
	require.ErrorContains(t, response.Validate(source.NodePubKey, salt, 41), "expired")

	replayed := response
	replayed.GeneratedHeight = 45
	replayed.ResponseID = ComputeDiscoveryResponseID(replayed)
	payload, err := DiscoveryResponseSigningPayload(replayed)
	require.NoError(t, err)
	replayed.SourceSignature = ed25519.Sign(deterministicPrivateKey(0x64), payload)
	require.ErrorContains(t, replayed.Validate(source.NodePubKey, salt, 45), "heights")

	forgedRecord := response
	forgedRecord.MatchedRecords = cloneDiscoveryRecords(response.MatchedRecords)
	forgedRecord.MatchedRecords[0].Signature = cloneBytes(forgedRecord.MatchedRecords[0].Signature)
	forgedRecord.MatchedRecords[0].Signature[0] ^= 0xff
	resultHash, err := ComputeDiscoveryResponseResultHash(forgedRecord.MatchedRecords)
	require.NoError(t, err)
	forgedRecord.ResultHash = resultHash
	forgedRecord.ResponseID = ComputeDiscoveryResponseID(forgedRecord)
	payload, err = DiscoveryResponseSigningPayload(forgedRecord)
	require.NoError(t, err)
	forgedRecord.SourceSignature = ed25519.Sign(deterministicPrivateKey(0x64), payload)
	require.ErrorContains(t, forgedRecord.Validate(source.NodePubKey, salt, 20), "signature")
}

func TestBroadcastMessageSignsDeduplicatesAndRejectsForgedOrExpired(t *testing.T) {
	salt := []byte("aetra-test-network")
	originKey := deterministicPrivateKey(0x66)
	origin := signedNodeRecord(t, 0x66, salt, 100, NodeRoleRouting)
	desc := defaultOverlayByType(t, OverlayTypeRouting)

	msg, err := SignBroadcastMessage(BroadcastMessage{
		OverlayID:	desc.OverlayID,
		PayloadHash:	HashParts("broadcast", "payload"),
		PayloadType:	BroadcastPayloadRouting,
		Height:		10,
		TTL:		8,
		Priority:	PriorityForChannel(ChannelRouting),
		FanoutPolicy: BroadcastFanoutPolicy{
			TreeFanout:	2,
			GossipFanout:	3,
			OverlayBound:	true,
		},
	}, originKey, salt)
	require.NoError(t, err)
	require.Equal(t, origin.NodeID, msg.OriginNode)
	require.NoError(t, VerifyBroadcastMessageSignature(msg, origin.NodePubKey, salt, 12))

	deduper := BroadcastDeduper{}
	deduper, accepted, err := deduper.Accept(msg)
	require.NoError(t, err)
	require.True(t, accepted)
	deduper, accepted, err = deduper.Accept(msg)
	require.NoError(t, err)
	require.False(t, accepted)
	require.Len(t, deduper.SeenKeys, 1)

	forged := msg
	forged.Signature = cloneBytes(msg.Signature)
	forged.Signature[0] ^= 0xff
	require.ErrorContains(t, VerifyBroadcastMessageSignature(forged, origin.NodePubKey, salt, 12), "signature")
	require.ErrorContains(t, msg.ValidateBasic(19), "expired")
}

func TestBroadcastForwardingUsesTreeThenGossipFallbackAndOverlayFanout(t *testing.T) {
	salt := []byte("aetra-test-network")
	originKey := deterministicPrivateKey(0x67)
	origin := signedNodeRecord(t, 0x67, salt, 100, NodeRoleRouting)
	local := signedNodeRecord(t, 0x68, salt, 100, NodeRoleRouting)
	peerA := signedNodeRecord(t, 0x69, salt, 100, NodeRoleRouting)
	peerB := signedNodeRecord(t, 0x6a, salt, 100, NodeRoleRouting)
	peerC := signedNodeRecord(t, 0x6b, salt, 100, NodeRoleRouting)
	peerD := signedNodeRecord(t, 0x6c, salt, 100, NodeRoleRouting)
	desc := defaultOverlayByType(t, OverlayTypeRouting)

	msg, err := SignBroadcastMessage(BroadcastMessage{
		OverlayID:	desc.OverlayID,
		PayloadHash:	HashParts("broadcast", "tree"),
		PayloadType:	BroadcastPayloadRouting,
		Height:		20,
		TTL:		5,
		Priority:	1,
		FanoutPolicy: BroadcastFanoutPolicy{
			TreeFanout:	desc.Fanout + 10,
			GossipFanout:	desc.Fanout + 10,
			OverlayBound:	true,
		},
	}, originKey, salt)
	require.NoError(t, err)

	graph := RoutingGraph{
		OverlayID:	desc.OverlayID,
		Version:	1,
		Edges: []RoutingEdge{
			{FromNodeID: local.NodeID, ToNodeID: peerB.NodeID, LatencyMillis: 50, Weight: 7_000, Priority: 2},
			{FromNodeID: local.NodeID, ToNodeID: peerA.NodeID, LatencyMillis: 10, Weight: 9_000, Priority: 1},
		},
	}
	graph.GraphHash = ComputeRoutingGraphHash(graph)
	candidates := []string{origin.NodeID, local.NodeID, peerA.NodeID, peerB.NodeID, peerC.NodeID, peerD.NodeID}
	deduper, plan, err := PlanBroadcastForwarding(msg, desc, graph, local.NodeID, candidates, BroadcastDeduper{}, 20)
	require.NoError(t, err)
	require.Equal(t, []string{peerA.NodeID, peerB.NodeID}, plan.TreeTargets)
	require.ElementsMatch(t, []string{peerC.NodeID, peerD.NodeID}, plan.GossipTargets)
	require.True(t, plan.FallbackUsed)
	require.Equal(t, uint32(4), plan.TTLRemaining)
	require.Len(t, deduper.SeenKeys, 1)

	_, duplicate, err := PlanBroadcastForwarding(msg, desc, graph, local.NodeID, candidates, deduper, 20)
	require.NoError(t, err)
	require.Empty(t, duplicate.TreeTargets)
	require.Empty(t, duplicate.GossipTargets)
	require.Equal(t, msg.BroadcastID, duplicate.BroadcastID)
}

func TestBroadcastPriorityOrderingAndOverlayMismatch(t *testing.T) {
	salt := []byte("aetra-test-network")
	key := deterministicPrivateKey(0x6d)
	serviceDesc := defaultOverlayByType(t, OverlayTypeService)
	routingDesc := defaultOverlayByType(t, OverlayTypeRouting)
	low, err := SignBroadcastMessage(BroadcastMessage{
		OverlayID:	serviceDesc.OverlayID,
		PayloadHash:	HashParts("broadcast", "low"),
		PayloadType:	BroadcastPayloadService,
		Height:		30,
		TTL:		5,
		Priority:	5,
	}, key, salt)
	require.NoError(t, err)
	high, err := SignBroadcastMessage(BroadcastMessage{
		OverlayID:	serviceDesc.OverlayID,
		PayloadHash:	HashParts("broadcast", "high"),
		PayloadType:	BroadcastPayloadService,
		Height:		31,
		TTL:		5,
		Priority:	1,
	}, key, salt)
	require.NoError(t, err)

	ordered := SortBroadcastMessages([]BroadcastMessage{low, high})
	require.Equal(t, high.BroadcastID, ordered[0].BroadcastID)
	_, _, err = PlanBroadcastForwarding(high, routingDesc, RoutingGraph{OverlayID: routingDesc.OverlayID, Version: 1}, HashParts("local"), nil, BroadcastDeduper{}, 31)
	require.ErrorContains(t, err, "overlay mismatch")
}

func TestBroadcastDedupCacheDropsDuplicatesDetectsConflictsAndPrunes(t *testing.T) {
	salt := []byte("aetra-test-network")
	key := deterministicPrivateKey(0x6e)
	peer := signedNodeRecord(t, 0x6f, salt, 100, NodeRoleRouting)
	desc := defaultOverlayByType(t, OverlayTypeRouting)
	msg, err := SignBroadcastMessage(BroadcastMessage{
		OverlayID:	desc.OverlayID,
		PayloadHash:	HashParts("broadcast", "dedupe"),
		PayloadType:	BroadcastPayloadRouting,
		Height:		40,
		TTL:		10,
		Priority:	2,
	}, key, salt)
	require.NoError(t, err)

	cache := NewBroadcastDedupCache(2)
	cache, decision, err := cache.Accept(msg, peer.NodeID, 40)
	require.NoError(t, err)
	require.True(t, decision.Accepted)
	require.Len(t, cache.Entries, 1)
	require.Equal(t, msg.BroadcastID, cache.Entries[0].BroadcastID)
	require.Equal(t, msg.PayloadHash, cache.Entries[0].PayloadHash)

	cache, decision, err = cache.Accept(msg, peer.NodeID, 41)
	require.NoError(t, err)
	require.True(t, decision.DroppedDuplicate)

	conflict := msg
	conflict.PayloadHash = HashParts("broadcast", "conflicting-payload")
	conflict.Signature = nil
	cache, decision, err = cache.Accept(conflict, peer.NodeID, 41)
	require.NoError(t, err)
	require.False(t, decision.Accepted)
	require.NotEmpty(t, decision.FaultEvidence.EvidenceHash)
	require.Equal(t, msg.BroadcastID, decision.FaultEvidence.BroadcastID)
	require.Equal(t, msg.PayloadHash, decision.FaultEvidence.ExpectedPayloadHash)
	require.Equal(t, conflict.PayloadHash, decision.FaultEvidence.ConflictingPayloadHash)

	pruned := cache.Prune(43)
	require.Empty(t, pruned.Entries)
	require.Len(t, pruned.Faults, 1)
}

func TestBlockHeaderFirstPropagationVerifiesChunksProofsAndReconstructs(t *testing.T) {
	salt := []byte("aetra-test-network")
	proposerKey := deterministicPrivateKey(0x70)
	validatorKey := deterministicPrivateKey(0x71).Public().(ed25519.PublicKey)
	addressHash, err := HashNetworkAddresses([]string{"tcp://127.0.0.70:26656"})
	require.NoError(t, err)
	proposer, err := SignNodeRecord(NodeRecord{
		ValidatorPubKey:	validatorKey,
		Roles:			[]NodeRole{NodeRoleValidator},
		NetworkAddressesHash:	addressHash,
		ProtocolVersions:	[]string{DefaultProtocolVersion},
		ExpiresHeight:		100,
	}, proposerKey, salt)
	require.NoError(t, err)

	blockBytes := bytes.Repeat([]byte("aetra-block-body"), 32)
	chunks, err := ChunkPayload(blockBytes, 96)
	require.NoError(t, err)
	chunkRoot, err := ComputeRL2ChunkRoot(chunks)
	require.NoError(t, err)
	proofHashes := []string{HashParts("block-proof", "commit"), HashParts("block-proof", "availability")}
	proofRoot := HashParts(append([]string{"block-proof-set"}, proofHashes...)...)
	headerHash := HashParts("block-header", "height-42")
	blockRoot := ComputeBlockRoot(headerHash, chunkRoot, proofRoot)
	header, err := NewBlockBroadcastHeader(BlockBroadcastHeader{
		Height:				42,
		ProposerNodeID:			proposer.NodeID,
		HeaderHash:			headerHash,
		ChunkSetRoot:			chunkRoot,
		ProofSetRoot:			proofRoot,
		BlockRoot:			blockRoot,
		ChunkCount:			uint32(len(chunks)),
		AvailabilityMetadataHash:	HashParts("availability", "metadata"),
	})
	require.NoError(t, err)
	proofSet := BlockProofSet{BlockID: header.BlockID, ProofRoot: proofRoot, ProofHashes: proofHashes}
	metadata, metadataRoot, err := NewBlockChunkMetadata(header.BlockID, chunks)
	require.NoError(t, err)
	require.Equal(t, header.ChunkSetRoot, metadataRoot)

	session, err := StartBlockPropagation(header, proofSet, proposer, 42)
	require.NoError(t, err)
	_, err = ReconstructBlock(session, metadata)
	require.ErrorContains(t, err, "all chunks")

	corrupt := chunks[0]
	corrupt.Bytes = cloneBytes(corrupt.Bytes)
	corrupt.Bytes[0] ^= 0xff
	_, err = AcceptBlockChunk(session, metadata[0], corrupt)
	require.ErrorContains(t, err, "chunk hash")

	for i, chunk := range chunks {
		session, err = AcceptBlockChunk(session, metadata[i], chunk)
		require.NoError(t, err)
	}
	require.Len(t, session.ReceivedChunks, len(chunks))
	reconstructed, err := ReconstructBlock(session, metadata)
	require.NoError(t, err)
	require.Equal(t, blockBytes, reconstructed)

	badProof := proofSet
	badProof.ProofRoot = HashParts("wrong-proof-root")
	_, err = StartBlockPropagation(header, badProof, proposer, 42)
	require.ErrorContains(t, err, "proof set root")
}

func TestStreamSessionFromSpecOpensAndAccountsFlowControl(t *testing.T) {
	salt := []byte("aetra-test-network")
	local := signedNodeRecord(t, 0x72, salt, 100, NodeRoleFull, NodeRoleStateSync)
	remote := signedNodeRecord(t, 0x73, salt, 100, NodeRoleService, NodeRoleStorageProvider)
	session, err := NegotiateVerifiedSession(local, remote, testSessionRequest(local, remote, 10, 80, "stream-session", []ChannelClass{
		ChannelConsensus,
		ChannelStateSync,
		ChannelData,
	}), salt, 10)
	require.NoError(t, err)

	var dataStream StreamSpec
	for _, stream := range session.Streams {
		if stream.Channel == ChannelData {
			dataStream = stream
			break
		}
	}
	require.NotEmpty(t, dataStream.StreamID)

	stream, err := StreamSessionFromSpec(session, dataStream, StreamingPayloadStateSync, 512, 4)
	require.NoError(t, err)
	require.Equal(t, StreamStateOpening, stream.State)
	require.Equal(t, ComputeStreamSessionID(stream), stream.StreamID)

	stream, err = OpenStreamSession(stream)
	require.NoError(t, err)
	plan, err := PlanStreamChunks(stream, 4096)
	require.NoError(t, err)
	require.Equal(t, uint64(512), plan.ChunkSize)
	require.Equal(t, uint32(4), plan.MaxInFlightChunks)
	require.False(t, plan.Backpressure)

	stream, window, err := RecordStreamBytesSent(stream, stream.FlowControlWindow)
	require.NoError(t, err)
	require.True(t, window.Backpressure)
	require.Zero(t, window.AvailableWindow)

	_, _, err = RecordStreamBytesSent(stream, 1)
	require.ErrorContains(t, err, "flow control")

	stream, window, err = AcknowledgeStreamBytes(stream, stream.FlowControlWindow/2)
	require.NoError(t, err)
	require.Equal(t, stream.FlowControlWindow/2, window.AvailableWindow)
	require.False(t, window.Backpressure)
}

func TestStreamSessionPauseResumeDrainCloseAndFail(t *testing.T) {
	stream, err := NewStreamSession(StreamSession{
		SessionID:		HashParts("parent-session"),
		PayloadType:		StreamingPayloadStorageObject,
		Priority:		PriorityForChannel(ChannelData),
		FlowControlWindow:	2048,
		ChunkSize:		512,
		Parallelism:		2,
	})
	require.NoError(t, err)

	stream, err = OpenStreamSession(stream)
	require.NoError(t, err)
	stream, _, err = RecordStreamBytesSent(stream, 1024)
	require.NoError(t, err)

	paused, err := PauseStreamSession(stream)
	require.NoError(t, err)
	_, err = PlanStreamChunks(paused, 1024)
	require.ErrorContains(t, err, "active or draining")

	stream, err = ResumeStreamSession(paused)
	require.NoError(t, err)
	_, err = CloseStreamSession(stream)
	require.ErrorContains(t, err, "draining")

	stream, err = DrainStreamSession(stream)
	require.NoError(t, err)
	_, err = CloseStreamSession(stream)
	require.ErrorContains(t, err, "unacknowledged")

	stream, _, err = AcknowledgeStreamBytes(stream, stream.BytesSent)
	require.NoError(t, err)
	stream, err = CloseStreamSession(stream)
	require.NoError(t, err)
	require.Equal(t, StreamStateClosed, stream.State)

	failing, err := NewStreamSession(StreamSession{
		SessionID:		HashParts("parent-session-fail"),
		PayloadType:		StreamingPayloadProofBundle,
		Priority:		PriorityForChannel(ChannelData),
		FlowControlWindow:	1024,
		ChunkSize:		256,
	})
	require.NoError(t, err)
	failing, err = OpenStreamSession(failing)
	require.NoError(t, err)
	failing, err = FailStreamSession(failing)
	require.NoError(t, err)
	require.Equal(t, StreamStateFailed, failing.State)
}

func TestStreamPayloadTypeMapsToRL2AndRejectsInvalidBounds(t *testing.T) {
	cases := map[StreamingPayloadType]struct {
		rl2	RL2PayloadType
		channel	ChannelClass
	}{
		StreamingPayloadStateSync:		{rl2: RL2PayloadStateSyncStream, channel: ChannelStateSync},
		StreamingPayloadZoneSnapshot:		{rl2: RL2PayloadZoneSnapshot, channel: ChannelData},
		StreamingPayloadBlockPropagation:	{rl2: RL2PayloadLargeBlock, channel: ChannelBlock},
		StreamingPayloadExecutionReceipts:	{rl2: RL2PayloadExecutionResult, channel: ChannelExecution},
		StreamingPayloadStorageObject:		{rl2: RL2PayloadStorageObject, channel: ChannelData},
		StreamingPayloadProofBundle:		{rl2: RL2PayloadProofSet, channel: ChannelData},
		StreamingPayloadHistoricalQueryRange:	{rl2: RL2PayloadStorageObject, channel: ChannelData},
	}

	for payloadType, expected := range cases {
		rl2Type, err := StreamingPayloadRL2Type(payloadType)
		require.NoError(t, err)
		require.Equal(t, expected.rl2, rl2Type)

		channel, err := StreamingPayloadChannel(payloadType)
		require.NoError(t, err)
		require.Equal(t, expected.channel, channel)
	}

	_, err := NewStreamSession(StreamSession{
		SessionID:		HashParts("invalid-stream-window"),
		PayloadType:		StreamingPayloadBlockPropagation,
		Priority:		PriorityForChannel(ChannelBlock),
		FlowControlWindow:	128,
		ChunkSize:		256,
	})
	require.ErrorContains(t, err, "chunk size exceeds")

	_, err = NewStreamSession(StreamSession{
		SessionID:		HashParts("invalid-stream-ack"),
		PayloadType:		StreamingPayloadExecutionReceipts,
		Priority:		PriorityForChannel(ChannelExecution),
		FlowControlWindow:	1024,
		ChunkSize:		256,
		BytesSent:		1,
		BytesAcknowledged:	2,
		State:			StreamStateActive,
	})
	require.ErrorContains(t, err, "acknowledged bytes exceed")

	_, err = NewStreamSession(StreamSession{
		SessionID:		HashParts("invalid-stream-opening"),
		PayloadType:		StreamingPayloadProofBundle,
		Priority:		PriorityForChannel(ChannelData),
		FlowControlWindow:	1024,
		ChunkSize:		256,
		BytesSent:		1,
	})
	require.ErrorContains(t, err, "opening")
}

func TestStreamBackpressureSignalsControlWindowAndState(t *testing.T) {
	stream, err := NewStreamSession(StreamSession{
		SessionID:		HashParts("backpressure-parent"),
		PayloadType:		StreamingPayloadExecutionReceipts,
		Priority:		PriorityForChannel(ChannelExecution),
		FlowControlWindow:	2048,
		ChunkSize:		512,
		Parallelism:		2,
	})
	require.NoError(t, err)
	stream, err = OpenStreamSession(stream)
	require.NoError(t, err)
	stream, _, err = RecordStreamBytesSent(stream, 1024)
	require.NoError(t, err)

	_, _, err = ApplyStreamBackpressure(stream, StreamBackpressureFrame{
		StreamID:		stream.StreamID,
		Signal:			StreamSignalWindowUpdate,
		FlowControlWindow:	512,
	})
	require.ErrorContains(t, err, "below in-flight")

	stream, window, err := ApplyStreamBackpressure(stream, StreamBackpressureFrame{
		StreamID:		stream.StreamID,
		Signal:			StreamSignalWindowUpdate,
		FlowControlWindow:	1536,
		CumulativeAcknowledge:	512,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1024), window.AvailableWindow)
	require.Equal(t, uint64(512), stream.BytesAcknowledged)

	stream, window, err = ApplyStreamBackpressure(stream, StreamBackpressureFrame{
		StreamID:	stream.StreamID,
		Signal:		StreamSignalPause,
	})
	require.NoError(t, err)
	require.Equal(t, StreamStatePaused, stream.State)
	require.True(t, window.Backpressure)

	ordered, err := SortStreamPriorityLanes([]StreamSession{
		mustTestStreamSession(t, StreamingPayloadStorageObject, PriorityForChannel(ChannelData)),
		stream,
		mustTestStreamSession(t, StreamingPayloadBlockPropagation, PriorityForChannel(ChannelBlock)),
	})
	require.NoError(t, err)
	require.Equal(t, StreamingPayloadBlockPropagation, ordered[0].PayloadType)
	require.Equal(t, StreamingPayloadExecutionReceipts, ordered[1].PayloadType)
	require.Equal(t, StreamingPayloadStorageObject, ordered[2].PayloadType)

	stream, _, err = ApplyStreamBackpressure(stream, StreamBackpressureFrame{
		StreamID:	stream.StreamID,
		Signal:		StreamSignalResume,
	})
	require.NoError(t, err)
	require.Equal(t, StreamStateActive, stream.State)

	stream, _, err = AcknowledgeStreamBytes(stream, stream.BytesSent)
	require.NoError(t, err)
	stream, window, err = ApplyStreamBackpressure(stream, StreamBackpressureFrame{
		StreamID:		stream.StreamID,
		Signal:			StreamSignalSlowDown,
		SuggestedChunkSize:	256,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(256), stream.ChunkSize)
	require.False(t, window.Backpressure)

	stream, _, err = ApplyStreamBackpressure(stream, StreamBackpressureFrame{
		StreamID:	stream.StreamID,
		Signal:		StreamSignalCancel,
	})
	require.NoError(t, err)
	require.Equal(t, StreamStateFailed, stream.State)
}

func TestStreamAdaptiveChunkingAndMetricsRespectBoundaries(t *testing.T) {
	stream, err := NewStreamSession(StreamSession{
		SessionID:		HashParts("adaptive-stream-parent"),
		PayloadType:		StreamingPayloadStateSync,
		Priority:		PriorityForChannel(ChannelStateSync),
		FlowControlWindow:	4096,
		ChunkSize:		512,
		Parallelism:		4,
	})
	require.NoError(t, err)
	stream, err = OpenStreamSession(stream)
	require.NoError(t, err)

	next, err := RecommendStreamChunkSize(stream, StreamAdaptiveChunkInputs{
		ObservedThroughputBps:	32 << 20,
		LossRateBps:		50,
		PeerScoreBps:		9_000,
		PayloadPriority:	stream.Priority,
		StreamClass:		QoSClassStateSync,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1024), next)
	stream, err = ApplyStreamChunkSize(stream, next)
	require.NoError(t, err)
	require.Equal(t, uint64(1024), stream.ChunkSize)

	stream, _, err = RecordStreamBytesSent(stream, 1024)
	require.NoError(t, err)
	stream, _, err = AcknowledgeStreamBytes(stream, 512)
	require.NoError(t, err)
	_, err = ApplyStreamChunkSize(stream, 512)
	require.ErrorContains(t, err, "chunk boundary")

	metrics, err := ComputeStreamMetrics(stream, 2_000, 3, 1)
	require.NoError(t, err)
	require.Equal(t, uint64(256), metrics.ThroughputBytesBps)
	require.Equal(t, uint32(5_000), metrics.CompletionBps)
	require.Equal(t, uint64(512), metrics.InFlightBytes)
	require.Equal(t, uint64(3), metrics.StallCount)

	stream, _, err = AcknowledgeStreamBytes(stream, stream.BytesSent)
	require.NoError(t, err)
	next, err = RecommendStreamChunkSize(stream, StreamAdaptiveChunkInputs{
		ObservedThroughputBps:	1 << 20,
		LossRateBps:		800,
		PeerScoreBps:		4_500,
		PayloadPriority:	stream.Priority,
		StreamClass:		QoSClassStateSync,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(512), next)
}

func TestStreamPartialRecoveryFetchAndStableReassemblyRoot(t *testing.T) {
	payload := bytes.Repeat([]byte("stream-payload"), 128)
	rootA, err := StreamReassemblyRoot(payload)
	require.NoError(t, err)
	chunksA, err := ChunkPayload(payload, 256)
	require.NoError(t, err)
	chunksB, err := ChunkPayload(payload, 512)
	require.NoError(t, err)
	rootB, err := StreamReassemblyRoot(payload)
	require.NoError(t, err)
	require.Equal(t, rootA, rootB)
	require.Equal(t, chunksA[0].PayloadHash, chunksB[0].PayloadHash)
	require.NoError(t, VerifyStreamChunkHash(chunksA[0]))

	corrupt := chunksA[0]
	corrupt.Bytes = cloneBytes(corrupt.Bytes)
	corrupt.Bytes[0] ^= 0xff
	require.ErrorContains(t, VerifyStreamChunkHash(corrupt), "chunk hash")

	stream, err := NewStreamSession(StreamSession{
		SessionID:		HashParts("partial-recovery-parent"),
		PayloadType:		StreamingPayloadStorageObject,
		Priority:		PriorityForChannel(ChannelData),
		FlowControlWindow:	2048,
		ChunkSize:		256,
		Parallelism:		3,
	})
	require.NoError(t, err)
	stream, err = OpenStreamSession(stream)
	require.NoError(t, err)

	verified := []bool{true, false, true, false, false, false, false}
	plan, err := PlanParallelStreamFetch(stream, uint64(len(payload)), verified, []string{"peer-a", "peer-b"})
	require.NoError(t, err)
	require.True(t, plan.RecoveryResume)
	require.Equal(t, uint32(7), plan.TotalChunks)
	require.Len(t, plan.Requests, 3)
	require.Equal(t, uint32(1), plan.Requests[0].ChunkIndex)
	require.Equal(t, "peer-a", plan.Requests[0].AssignedPeer)
	require.Equal(t, uint32(3), plan.Requests[1].ChunkIndex)
	require.Equal(t, "peer-b", plan.Requests[1].AssignedPeer)
	require.Equal(t, uint32(4), plan.Requests[2].ChunkIndex)
	require.Equal(t, "peer-a", plan.Requests[2].AssignedPeer)
}

func TestNetworkSecurityPolicyDetectsThreatsAndProtectsConsensus(t *testing.T) {
	policy := DefaultNetworkSecurityPolicy()
	require.NoError(t, policy.Validate())

	score := PeerScore{ScoreBps: 8_500, ReliabilityBps: 9_000, ThroughputBps: 8_000}
	decision, err := EvaluateNetworkSecurity(score, PeerSecurityObservation{
		PeerNodeID:		HashParts("security-peer"),
		InvalidMessages:	policy.MaxInvalidMessages + 1,
		DuplicateMessages:	policy.MaxInvalidMessages + 2,
		CorruptChunks:		1,
		ForgedAdvertisements:	1,
		BytesThisEpoch:		policy.MaxBytesPerEpoch + 1,
		SybilClusterPeers:	3,
		CrossZoneReplayCount:	1,
		ConflictingBroadcasts:	1,
	}, policy)
	require.NoError(t, err)
	require.False(t, decision.Accepted)
	require.True(t, decision.Quarantine)
	require.True(t, decision.RotatePeer)
	require.True(t, decision.DropMessage)
	require.True(t, decision.DowngradeQoS)
	require.True(t, decision.ConsensusIsolated)
	require.Less(t, decision.Score.ScoreBps, score.ScoreBps)
	require.Contains(t, decision.Threats, ThreatSpamFlood)
	require.Contains(t, decision.Threats, ThreatBandwidthExhaustion)
	require.Contains(t, decision.Threats, ThreatChunkCorruption)
	require.Contains(t, decision.Controls, ControlReplayProtection)
	require.Contains(t, decision.Controls, ControlQoSIsolation)

	ok, err := CheckChannelRateLimit(ChannelRateLimit{Channel: ChannelDiscovery, MaxBytes: 1024, WindowHeight: 1, DropOnExceeded: true}, ChannelRateUsage{
		Channel:	ChannelDiscovery,
		Bytes:		2048,
		WindowStart:	10,
		WindowEnd:	10,
	})
	require.NoError(t, err)
	require.False(t, ok)

	broken := policy
	broken.RequiredControls = []SecurityControl{ControlPeerReputation}
	require.ErrorContains(t, broken.Validate(), "missing required control")
}

func TestNetworkSecurityReplayChannelBindingAndQoSIsolation(t *testing.T) {
	cache := NewReplayProtectionCache(4)
	replayID := HashParts("cross-zone-message", "zone-a", "zone-b", "1")
	var accepted bool
	var err error
	cache, accepted, err = cache.Accept(replayID, 10)
	require.NoError(t, err)
	require.True(t, accepted)
	cache, accepted, err = cache.Accept(replayID, 11)
	require.NoError(t, err)
	require.False(t, accepted)
	cache, accepted, err = cache.Accept(HashParts("cross-zone-message", "zone-a", "zone-b", "2"), 20)
	require.NoError(t, err)
	require.True(t, accepted)
	require.Len(t, cache.Entries, 1)

	salt := []byte("aetra-test-network")
	local := signedNodeRecord(t, 0x74, salt, 100, NodeRoleFull)
	remote := signedNodeRecord(t, 0x75, salt, 100, NodeRoleService)
	session, err := NegotiateVerifiedSession(local, remote, testSessionRequest(local, remote, 10, 80, "security-channel", nil), salt, 10)
	require.NoError(t, err)
	require.NoError(t, ValidateSecurityChannelBinding(session, remote, session.Streams[0].StreamID))
	require.ErrorContains(t, ValidateSecurityChannelBinding(session, local, session.Streams[0].StreamID), "remote node mismatch")
	require.ErrorContains(t, ValidateSecurityChannelBinding(session, remote, "missing-stream"), "not bound")

	require.NoError(t, ValidateSecurityQoSIsolation(DefaultQoSClassPolicies()))
	broken := append([]QoSClassPolicy(nil), DefaultQoSClassPolicies()...)
	for i := range broken {
		if broken[i].Class == QoSClassBulkData {
			broken[i].Priority = 2
		}
	}
	require.ErrorContains(t, ValidateSecurityQoSIsolation(broken), "bulk traffic")
}

func TestNetworkSecurityAuthenticatesDiscoveryOverlayAndChunks(t *testing.T) {
	salt := []byte("aetra-test-network")
	owner := signedNodeRecordWithCapabilities(t, 0x76, salt, 100, []NodeRole{NodeRoleService}, nil, []string{"svc.secure"})
	record := testSignedDiscoveryObjectRecord(t, owner, 0x76, salt, DRTObjectServiceEndpoint, HashParts("target", "svc.secure"), HashParts("endpoint", "svc.secure"), "", "svc.secure", "", 90)
	require.NoError(t, ValidateSecurityDiscoveryRecord(record, salt, 20))

	forged := record
	forged.AdvertisementHash = HashParts("endpoint", "forged")
	require.ErrorContains(t, ValidateSecurityDiscoveryRecord(forged, salt, 20), "record id")

	expired := record
	expired.ExpiresHeight = 19
	expired.RecordID = ComputeDiscoveryRecordID(expired)
	require.ErrorContains(t, ValidateSecurityDiscoveryRecord(expired, salt, 20), "expired")

	serviceDesc := testDefaultOverlayDescriptor(t, OverlayTypeService)
	serviceMsg := testMeshMessage(t, MeshMessageService, serviceDesc.OverlayID, owner.NodeID, HashParts("security-destination"), PriorityForChannel(ChannelService), 1)
	require.NoError(t, ValidateOverlayIsolation(serviceDesc, serviceMsg))
	badMsg := serviceMsg
	badMsg.Type = MeshMessageExecution
	badMsg.MessageID = ComputeAetherMeshMessageID(badMsg)
	require.ErrorContains(t, ValidateOverlayIsolation(serviceDesc, badMsg), "non-service")

	payload := bytes.Repeat([]byte("secure-chunk"), 64)
	chunks, err := ChunkPayload(payload, 128)
	require.NoError(t, err)
	transfer, err := NewRL2TransferFromChunks(owner.NodeID, HashParts("security-target"), RL2PayloadStorageObject, chunks, PriorityForChannel(ChannelData), 0, RL2FECNone)
	require.NoError(t, err)
	descriptors, err := NewRL2ChunkDescriptors(transfer, chunks)
	require.NoError(t, err)
	require.NoError(t, VerifySecurityChunk(transfer, descriptors[0], chunks[0]))

	corrupt := chunks[0]
	corrupt.Bytes = cloneBytes(corrupt.Bytes)
	corrupt.Bytes[0] ^= 0xff
	require.ErrorContains(t, VerifySecurityChunk(transfer, descriptors[0], corrupt), "chunk hash")
}

func TestNetworkSecurityPeerDiversityAndWithheldBlockChunks(t *testing.T) {
	salt := []byte("aetra-test-network")
	peerA := signedNodeRecordWithCapabilities(t, 0x77, salt, 100, []NodeRole{NodeRoleFull}, []string{"zone-a"}, nil)
	peerB := peerA
	peerC := signedNodeRecordWithCapabilities(t, 0x78, salt, 100, []NodeRole{NodeRoleFull}, []string{"zone-b"}, nil)

	report, err := EvaluatePeerDiversity([]NodeRecord{peerA, peerB, peerC}, DefaultNetworkSecurityPolicy())
	require.NoError(t, err)
	require.Equal(t, uint32(3), report.TotalPeers)
	require.Equal(t, uint32(2), report.UniqueNodeIDs)
	require.True(t, report.SybilRisk)

	proposerKey := deterministicPrivateKey(0x79)
	validatorKey := deterministicPrivateKey(0x7a).Public().(ed25519.PublicKey)
	addressHash, err := HashNetworkAddresses([]string{"tcp://127.0.0.79:26656"})
	require.NoError(t, err)
	proposer, err := SignNodeRecord(NodeRecord{
		ValidatorPubKey:	validatorKey,
		Roles:			[]NodeRole{NodeRoleValidator},
		NetworkAddressesHash:	addressHash,
		ProtocolVersions:	[]string{DefaultProtocolVersion},
		ExpiresHeight:		100,
	}, proposerKey, salt)
	require.NoError(t, err)
	blockBytes := bytes.Repeat([]byte("security-block"), 32)
	chunks, err := ChunkPayload(blockBytes, 96)
	require.NoError(t, err)
	chunkRoot, err := ComputeRL2ChunkRoot(chunks)
	require.NoError(t, err)
	proofHashes := []string{HashParts("security-proof", "commit")}
	proofRoot := HashParts(append([]string{"block-proof-set"}, proofHashes...)...)
	headerHash := HashParts("security-header", "42")
	header, err := NewBlockBroadcastHeader(BlockBroadcastHeader{
		Height:				42,
		ProposerNodeID:			proposer.NodeID,
		HeaderHash:			headerHash,
		ChunkSetRoot:			chunkRoot,
		ProofSetRoot:			proofRoot,
		BlockRoot:			ComputeBlockRoot(headerHash, chunkRoot, proofRoot),
		ChunkCount:			uint32(len(chunks)),
		AvailabilityMetadataHash:	HashParts("security-availability"),
	})
	require.NoError(t, err)
	session, err := StartBlockPropagation(header, BlockProofSet{BlockID: header.BlockID, ProofRoot: proofRoot, ProofHashes: proofHashes}, proposer, 42)
	require.NoError(t, err)
	withheld, err := DetectWithheldBlockChunks(session, 45, DefaultNetworkSecurityPolicy())
	require.NoError(t, err)
	require.True(t, withheld)
}

func TestPeerReputationIsAdvisoryUntilEvidenceCommittedAndDecayBounded(t *testing.T) {
	input := PeerReputationInput{
		PeerNodeID:			HashParts("reputation-peer"),
		ValidMessages:			95,
		InvalidMessages:		5,
		LatencyMillis:			50,
		ThroughputBytesPerSec:		32 << 20,
		CorrectChunks:			20,
		CorruptChunks:			1,
		ValidDiscoveryResponses:	9,
		InvalidDiscoveryResponses:	1,
		ValidServiceResponses:		8,
		InvalidServiceResponses:	2,
		Timeouts:			1,
		DuplicateBroadcasts:		2,
		ConflictingBroadcasts:		0,
		ElapsedEpochs:			2,
		DecayPolicy: PeerScoreDecayPolicy{
			MaxDecayBpsPerEpoch:	300,
			MinScoreBps:		4_000,
		},
	}
	decision, err := ComputePeerReputation(input)
	require.NoError(t, err)
	require.True(t, decision.LocalAdvisory)
	require.False(t, decision.ConsensusEligible)
	require.Greater(t, decision.PenaltyBps, uint32(0))
	require.Equal(t, uint32(600), decision.DecayAppliedBps)
	require.NoError(t, ValidateReputationConsensusUse(decision, false))
	require.ErrorContains(t, ValidateReputationConsensusUse(decision, true), "advisory")

	input.UsedForConsensus = true
	_, err = ComputePeerReputation(input)
	require.ErrorContains(t, err, "committed evidence")

	input.CommittedEvidence = true
	input.EvidenceHash = HashParts("reputation-evidence", input.PeerNodeID)
	input.EvidenceHeight = 55
	committed, err := ComputePeerReputation(input)
	require.NoError(t, err)
	require.False(t, committed.LocalAdvisory)
	require.True(t, committed.ConsensusEligible)
	require.NoError(t, ValidateReputationConsensusUse(committed, true))

	input.ElapsedEpochs = 100
	bounded, err := ComputePeerReputation(input)
	require.NoError(t, err)
	require.Equal(t, uint32(4_000), bounded.Score.ScoreBps)
}

func TestEclipseResistancePlanMaintainsDiversityAndProofBackedRouting(t *testing.T) {
	graph := AdaptiveOverlayGraph{
		OverlayID:	HashParts("eclipse-overlay"),
		LocalNodeID:	HashParts("eclipse-local"),
		RoutingEpoch:	7,
		PolicyHash:	HashParts("eclipse-policy"),
		RandomSet: []AdaptivePeer{
			testSecurityAdaptivePeer("random-a", []NodeRole{NodeRoleFull}, []string{"zone-a"}),
			testSecurityAdaptivePeer("random-b", []NodeRole{NodeRoleFull}, []string{"zone-b"}),
		},
		FallbackSet: []AdaptivePeer{
			testSecurityAdaptivePeer("fallback-a", []NodeRole{NodeRoleFull}, []string{"zone-c"}),
		},
		StableSet: []AdaptivePeer{
			testSecurityAdaptivePeer("validator-a", []NodeRole{NodeRoleValidator}, []string{"zone-a"}),
			testSecurityAdaptivePeer("validator-b", []NodeRole{NodeRoleValidator}, []string{"zone-b"}),
		},
		ZoneSet: []AdaptivePeer{
			testSecurityAdaptivePeer("zone-a", []NodeRole{NodeRoleZoneExecution}, []string{"zone-a"}),
			testSecurityAdaptivePeer("zone-b", []NodeRole{NodeRoleZoneExecution}, []string{"zone-b"}),
		},
	}
	records := []DiscoveryRecord{
		{
			RecordType:	DRTObjectRoutingEntryPoint,
			OwnerNodeID:	graph.StableSet[0].NodeID,
			ProofHash:	HashParts("proof-backed-routing", "a"),
			ProofHeight:	7,
		},
		{
			RecordType:	DRTObjectRoutingEntryPoint,
			OwnerNodeID:	graph.StableSet[1].NodeID,
		},
	}
	plan, err := BuildEclipseResistancePlan(graph, records, DefaultEclipseResistancePolicy(), 7)
	require.NoError(t, err)
	require.True(t, plan.RandomSetMaintained)
	require.True(t, plan.ValidatorDiversity)
	require.True(t, plan.ZoneDiversity)
	require.True(t, plan.DiscoverySourcesRotated)
	require.True(t, plan.IdentityClusterLimited)
	require.True(t, plan.ProofBackedCriticalRoute)
	require.Equal(t, []string{graph.StableSet[0].NodeID}, plan.CriticalRoutingPeers)
	require.NoError(t, ValidateEclipseResistancePlan(plan))

	clustered := graph
	clustered.FastSet = []AdaptivePeer{
		testSecurityAdaptivePeer("cluster-a", []NodeRole{NodeRoleFull}, []string{"cluster-zone"}),
		testSecurityAdaptivePeer("cluster-b", []NodeRole{NodeRoleFull}, []string{"cluster-zone"}),
		testSecurityAdaptivePeer("cluster-c", []NodeRole{NodeRoleFull}, []string{"cluster-zone"}),
	}
	badPlan, err := BuildEclipseResistancePlan(clustered, nil, DefaultEclipseResistancePolicy(), 8)
	require.NoError(t, err)
	require.False(t, badPlan.IdentityClusterLimited)
	require.NotEmpty(t, badPlan.PeersToDrop)
	require.ErrorContains(t, ValidateEclipseResistancePlan(badPlan), "identity cluster")
}

func TestSpamResistanceSignedEnvelopeRateLimitsAndResourceBackedAds(t *testing.T) {
	salt := []byte("aetra-test-network")
	privateKey := deterministicPrivateKey(0x7b)
	signer, err := SignNodeRecord(NodeRecord{
		Roles:			[]NodeRole{NodeRoleService},
		NetworkAddressesHash:	HashParts("network-addresses", "signed-envelope"),
		ServicesSupported:	[]string{"svc.secure"},
		ProtocolVersions:	[]string{DefaultProtocolVersion},
		ExpiresHeight:		100,
	}, privateKey, salt)
	require.NoError(t, err)

	envelope := testEnvelope(ChannelService, 1024, 20, 1, "signed-security-envelope")
	signed, err := SignSecurityEnvelope(envelope, signer, privateKey, 20)
	require.NoError(t, err)
	require.NoError(t, signed.Validate(signer.NodePubKey))
	tampered := signed
	tampered.Envelope.PayloadHash = HashParts("tampered-envelope")
	require.ErrorContains(t, tampered.Validate(signer.NodePubKey), "signature")

	policy := DefaultNetworkSecurityPolicy()
	decision, err := EvaluatePeerRateLimit(policy, PeerRateUsage{
		PeerNodeID:	signer.NodeID,
		Channel:	ChannelService,
		Messages:	policy.MaxPeerMessagesPerWindow + 1,
		Bytes:		9 << 20,
		WindowStart:	20,
		WindowEnd:	20,
	})
	require.NoError(t, err)
	require.False(t, decision.Allowed)
	require.True(t, decision.ExceededMessages)
	require.True(t, decision.ExceededBytes)
	require.True(t, decision.ThrottlePeer)

	req := testSessionRequest(signer, signedNodeRecord(t, 0x7c, salt, 100, NodeRoleFull), 20, 40, "cost-limited", nil)
	report, err := EvaluateHandshakeCost(req, policy)
	require.NoError(t, err)
	require.True(t, report.Accepted)
	tightPolicy := policy
	tightPolicy.MaxHandshakeCostUnits = 1
	report, err = EvaluateHandshakeCost(req, tightPolicy)
	require.NoError(t, err)
	require.False(t, report.Accepted)

	require.NoError(t, ValidatePayloadSize(envelope, policy))
	tightPayloadPolicy := policy
	tightPayloadPolicy.MaxPayloadBytes = 512
	require.ErrorContains(t, ValidatePayloadSize(envelope, tightPayloadPolicy), "payload size")

	record := testSignedDiscoveryObjectRecord(t, signer, 0x7b, salt, DRTObjectServiceEndpoint, HashParts("target", "svc.secure"), HashParts("endpoint", "svc.secure"), "", "svc.secure", "", 90)
	ad := testDRTAdvertisement(DRTObjectServiceEndpoint, signer, "", "", "svc.secure", HashParts("endpoint", "svc.secure"), policy.MinServiceStakeWeight, 8_000, 20, 90)
	ad.Discovery = record
	ad.ObjectID = ComputeDRTObjectID(ad)
	ad.AdvertisementID = ComputeDRTAdvertisementID(ad)
	require.NoError(t, ValidateResourceBackedAdvertisement(ad, salt, 20, policy, false))
	ad.StakeWeight = policy.MinServiceStakeWeight - 1
	ad.AdvertisementID = ComputeDRTAdvertisementID(ad)
	require.ErrorContains(t, ValidateResourceBackedAdvertisement(ad, salt, 20, policy, false), "stake backed")
}

func TestSpamResistanceChunkLimitsDuplicateSuppressionAndSimulations(t *testing.T) {
	policy := DefaultNetworkSecurityPolicy()
	source := HashParts("spam-source")
	target := HashParts("spam-target")
	payload := bytes.Repeat([]byte("spam-chunk"), 64)
	chunks, err := ChunkPayload(payload, 64)
	require.NoError(t, err)
	transfer, err := NewRL2TransferFromChunks(source, target, RL2PayloadStorageObject, chunks, PriorityForChannel(ChannelData), 0, RL2FECNone)
	require.NoError(t, err)
	request, err := NewRL2ChunkRequest(transfer, []uint32{0})
	require.NoError(t, err)
	require.NoError(t, ValidateChunkRequestLimit(request, transfer, policy))

	badRequest := request
	badRequest.MissingIndexes = []uint32{1, 1}
	require.ErrorContains(t, ValidateChunkRequestLimit(badRequest, transfer, policy), "duplicate")
	tightPolicy := policy
	tightPolicy.MaxChunkRequestsPerWindow = 1
	badRequest.MissingIndexes = []uint32{1, 2}
	require.ErrorContains(t, ValidateChunkRequestLimit(badRequest, transfer, tightPolicy), "limit")

	origin := signedNodeRecord(t, 0x7d, []byte("aetra-test-network"), 100, NodeRoleFull)
	msg, err := SignBroadcastMessage(BroadcastMessage{
		OriginNode:	origin.NodeID,
		OverlayID:	HashParts("spam-overlay"),
		PayloadHash:	HashParts("spam-payload"),
		PayloadType:	BroadcastPayloadService,
		TTL:		5,
		Priority:	PriorityForChannel(ChannelService),
		FanoutPolicy: BroadcastFanoutPolicy{
			TreeFanout:	1,
			GossipFanout:	2,
			OverlayBound:	true,
		},
		Height:	20,
	}, deterministicPrivateKey(0x7d), []byte("aetra-test-network"))
	require.NoError(t, err)
	cache := NewBroadcastDedupCache(4)
	cache, accepted, err := SuppressDuplicateBroadcast(cache, msg, origin.NodeID, 20)
	require.NoError(t, err)
	require.True(t, accepted)
	_, accepted, err = SuppressDuplicateBroadcast(cache, msg, origin.NodeID, 20)
	require.NoError(t, err)
	require.False(t, accepted)

	spamResult, err := SimulateSpamResistance(policy, PeerRateUsage{
		PeerNodeID:	origin.NodeID,
		Channel:	ChannelService,
		Messages:	policy.MaxPeerMessagesPerWindow + 10,
		Bytes:		10 << 20,
		WindowStart:	20,
		WindowEnd:	20,
	}, []BroadcastMessage{msg, msg}, origin.NodeID, 20)
	require.NoError(t, err)
	require.True(t, spamResult.Throttled)
	require.Contains(t, spamResult.Threats, ThreatSpamFlood)
	require.Contains(t, spamResult.Threats, ThreatBandwidthExhaustion)

	conflict := msg
	conflict.PayloadHash = HashParts("spam-conflicting-payload")
	conflict.Signature = nil
	conflict.BroadcastID = msg.BroadcastID
	conflict.Signature = msg.Signature
	manipulation, err := SimulateRoutingManipulation([]BroadcastMessage{msg, conflict}, origin.NodeID, 20)
	require.NoError(t, err)
	require.Equal(t, uint64(1), manipulation.FaultsDetected)
	require.Contains(t, manipulation.Threats, ThreatRoutingManipulation)

	graph := AdaptiveOverlayGraph{
		OverlayID:	HashParts("spam-eclipse-overlay"),
		LocalNodeID:	HashParts("spam-eclipse-local"),
		RoutingEpoch:	1,
		PolicyHash:	HashParts("spam-eclipse-policy"),
		RandomSet: []AdaptivePeer{
			testSecurityAdaptivePeer("only-random", []NodeRole{NodeRoleFull}, []string{"zone-a"}),
		},
		FallbackSet: []AdaptivePeer{
			testSecurityAdaptivePeer("fallback", []NodeRole{NodeRoleFull}, []string{"zone-a"}),
		},
	}
	_, threats, err := SimulateEclipseResistance(graph, nil, DefaultEclipseResistancePolicy(), 1)
	require.NoError(t, err)
	require.Contains(t, threats, ThreatEclipseAttack)
}

func TestPerformanceModelValidatesOptimizationTargets(t *testing.T) {
	blockSession := testPerformanceBlockSession(t)
	streamPlan := testPerformanceStreamPlan()
	zoneGraph := RoutingGraph{
		OverlayID:	HashParts("performance-zone-overlay"),
		Version:	1,
		Edges: []RoutingEdge{
			{FromNodeID: HashParts("performance-source"), ToNodeID: HashParts("performance-zone-peer-a"), LatencyMillis: 25, Weight: 9_000, Priority: 1, ZoneID: "zone-a"},
			{FromNodeID: HashParts("performance-source"), ToNodeID: HashParts("performance-zone-peer-b"), LatencyMillis: 35, Weight: 8_500, Priority: 2, ZoneID: "zone-a"},
		},
	}
	zoneGraph.GraphHash = ComputeRoutingGraphHash(zoneGraph)

	plan, err := BuildPerformanceModelPlan(PerformanceModelInput{
		PeerCount:			4096,
		DiscoveryBranchingFactor:	16,
		OverlayDescriptors:		DefaultOverlayDescriptors(),
		RoutingGraphs:			[]RoutingGraph{zoneGraph},
		BlockSession:			blockSession,
		StreamPlan:			streamPlan,
		QoSPolicies:			DefaultQoSClassPolicies(),
		ZoneID:				"zone-a",
		MaxZoneLatencyMillis:		50,
	})
	require.NoError(t, err)
	require.Equal(t, uint32(3), plan.DiscoveryHops)
	require.GreaterOrEqual(t, plan.OverlayConcurrency, uint32(2))
	require.False(t, plan.GlobalBroadcastOnly)
	require.True(t, plan.ShardLocalExecution)
	require.True(t, plan.ZoneIsolated)
	require.True(t, plan.HeaderFirstBlockPropagation)
	require.True(t, plan.ParallelChunkStreaming)
	require.True(t, plan.ServiceTrafficIsolated)
	require.Contains(t, plan.SatisfiedOptimizationGoals, PerformanceGoalParallelPropagation)
	require.Contains(t, plan.SatisfiedOptimizationGoals, PerformanceGoalNoGlobalBroadcastOnly)
	require.Contains(t, plan.SatisfiedTargetProperties, PerformanceTargetLogDiscovery)
	require.Contains(t, plan.SatisfiedTargetProperties, PerformanceTargetParallelChunks)
	require.NoError(t, ValidatePerformanceModelPlan(plan))

	plan.GlobalBroadcastOnly = true
	require.ErrorContains(t, ValidatePerformanceModelPlan(plan), "global broadcast")
}

func TestPerformanceModelBoundsFanoutLatencyStreamingAndQoS(t *testing.T) {
	desc := testDefaultOverlayDescriptor(t, OverlayTypeService)
	fanout, err := ValidateBoundedOverlayFanout(desc, 100)
	require.NoError(t, err)
	require.LessOrEqual(t, fanout, desc.Fanout)

	msg := BroadcastMessage{
		BroadcastID:	HashParts("performance-broadcast"),
		OriginNode:	HashParts("performance-origin"),
		OverlayID:	desc.OverlayID,
		PayloadHash:	HashParts("performance-payload"),
		PayloadType:	BroadcastPayloadService,
		Height:		10,
		TTL:		10,
		Priority:	PriorityForChannel(ChannelService),
		FanoutPolicy:	BroadcastFanoutPolicy{TreeFanout: desc.Fanout + 10, GossipFanout: desc.Fanout + 20, OverlayBound: true},
	}
	require.NoError(t, ValidateBoundedBroadcastFanout(msg, desc))
	require.NoError(t, ValidateHeaderFirstPerformance(testPerformanceBlockSession(t)))
	require.NoError(t, ValidateParallelChunkPerformance(testPerformanceStreamPlan()))

	graph := RoutingGraph{
		OverlayID:	HashParts("performance-latency-overlay"),
		Version:	1,
		Edges: []RoutingEdge{
			{FromNodeID: HashParts("latency-source"), ToNodeID: HashParts("latency-fast"), LatencyMillis: 20, ZoneID: "zone-fast"},
			{FromNodeID: HashParts("latency-source"), ToNodeID: HashParts("latency-slow"), LatencyMillis: 300, ZoneID: "zone-fast"},
		},
	}
	latency, isolated, err := EvaluateZoneLocalLatency([]RoutingGraph{graph}, "zone-fast", 100)
	require.NoError(t, err)
	require.Equal(t, uint64(300), latency)
	require.False(t, isolated)

	require.NoError(t, ValidatePerformanceQoSIsolation(DefaultQoSClassPolicies()))
	broken := append([]QoSClassPolicy(nil), DefaultQoSClassPolicies()...)
	for i := range broken {
		if broken[i].Class == QoSClassServiceCall {
			broken[i].Priority = 0
		}
	}
	require.ErrorContains(t, ValidatePerformanceQoSIsolation(broken), "priority inversion")

	_, err = EstimateDiscoveryHops(100, 1)
	require.ErrorContains(t, err, "branching")
}

func TestPerformanceMetricsSnapshotAggregatesOverlayBandwidthAndScores(t *testing.T) {
	salt := []byte("aetra-test-network")
	storage := signedNodeRecord(t, 0x7e, salt, 100, NodeRoleStorageProvider)
	service := signedNodeRecord(t, 0x7f, salt, 100, NodeRoleService)
	full := signedNodeRecord(t, 0x80, salt, 100, NodeRoleFull)
	overlayID := HashParts("performance-metrics-overlay")
	streamPlan := testPerformanceStreamPlan()
	streamMetric := StreamMetrics{
		StreamID:		streamPlan.StreamID,
		PayloadType:		StreamingPayloadStorageObject,
		State:			StreamStateActive,
		BytesSent:		1024,
		BytesAcknowledged:	768,
		InFlightBytes:		256,
		AvailableWindow:	1024,
		ThroughputBytesBps:	4096,
		StallCount:		2,
		CompletionBps:		7_500,
	}
	snapshot, err := BuildPerformanceMetricsSnapshot(PerformanceMetricsInput{
		NodeRecords:	[]NodeRecord{storage, service, full},
		OverlayMemberships: []OverlayMembershipRecord{
			{OverlayID: overlayID, NodeID: storage.NodeID},
			{OverlayID: overlayID, NodeID: service.NodeID},
		},
		MessageLatencies: []PropagationLatencySample{
			{OverlayID: overlayID, MessageID: HashParts("msg", "1"), LatencyMillis: 20},
			{OverlayID: overlayID, MessageID: HashParts("msg", "2"), LatencyMillis: 40},
		},
		RouteFailures: []RouteFailureSample{
			{OverlayID: overlayID, Attempts: 10, Failures: 1},
		},
		BlockSession:			testPerformanceBlockSession(t),
		BlockHeaderLatencyMillis:	15,
		BlockReconstructionMillis:	25,
		BlockBytes:			2048,
		ChunkAttempts:			10,
		ChunkRetries:			1,
		StreamMetrics:			[]StreamMetrics{streamMetric},
		StreamPlans:			[]StreamParallelFetchPlan{streamPlan},
		DiscoveryLatencies:		[]uint64{10, 20, 30},
		CrossZoneDeliveries: []CrossZoneDeliverySample{
			{SourceZone: "zone-a", DestinationZone: "zone-b", Sequence: 1, LatencyMillis: 50},
			{SourceZone: "zone-a", DestinationZone: "zone-b", Sequence: 2, LatencyMillis: 70},
		},
		ChannelMetrics: []L0ChannelMetrics{
			{Channel: ChannelConsensus, EnqueuedCount: 1, SentCount: 1, BytesEnqueued: 1000, BytesSent: 1000},
			{Channel: ChannelService, EnqueuedCount: 2, SentCount: 2, BytesEnqueued: 2000, BytesSent: 1500},
		},
		PeerScores:	[]PeerScore{{ScoreBps: 9_000}, {ScoreBps: 6_000}, {ScoreBps: 3_000}},
	})
	require.NoError(t, err)
	require.Len(t, snapshot.PeerCountByRole, 3)
	require.Equal(t, uint64(2), snapshot.OverlayMetrics[0].MembershipSize)
	require.Equal(t, uint64(30), snapshot.OverlayMetrics[0].MessagePropagationLatency.AverageMillis)
	require.Equal(t, uint32(1_000), snapshot.OverlayMetrics[0].RouteFailureRateBps)
	require.Equal(t, uint64(15), snapshot.BlockBenchmark.HeaderLatencyMillis)
	require.Equal(t, uint64(2), snapshot.ChunkBenchmarks[0].StallCount)
	require.Equal(t, uint32(1_000), snapshot.ChunkBenchmarks[0].RetryRateBps)
	require.Equal(t, uint64(20), snapshot.DiscoveryQueryLatency.AverageMillis)
	require.Equal(t, uint64(60), snapshot.CrossZoneDeliveryLatency.AverageMillis)
	require.True(t, snapshot.ServiceTrafficIsolated)
	require.Equal(t, uint32(1_000), snapshot.RouteFailureRateBps)
	require.Equal(t, uint32(6_000), snapshot.PeerScoreDistribution.AverageBps)
	require.Equal(t, uint64(1), snapshot.PeerScoreDistribution.LowCount)
	require.Equal(t, uint64(1), snapshot.PeerScoreDistribution.MidCount)
	require.Equal(t, uint64(1), snapshot.PeerScoreDistribution.HighCount)
}

func TestPerformanceMetricsRejectInvalidBenchmarksAndServiceIsolationFailures(t *testing.T) {
	_, err := BenchmarkBlockPropagation(testPerformanceBlockSession(t), 0, 10, 1024)
	require.ErrorContains(t, err, "header latency")

	_, err = BenchmarkChunkStreaming(nil, []StreamParallelFetchPlan{testPerformanceStreamPlan()}, 1, 2)
	require.ErrorContains(t, err, "retries exceed")

	_, err = ComputeOverlayPerformanceMetrics(nil, []PropagationLatencySample{{OverlayID: HashParts("overlay"), LatencyMillis: 0}}, nil)
	require.ErrorContains(t, err, "latency")

	_, err = ComputeRouteFailureRate([]RouteFailureSample{{OverlayID: HashParts("overlay"), Attempts: 1, Failures: 2}})
	require.ErrorContains(t, err, "failures exceed")

	_, err = ComputePeerScoreDistribution([]PeerScore{{ScoreBps: BasisPoints + 1}})
	require.ErrorContains(t, err, "peer score")

	require.ErrorContains(t, ValidateServiceTrafficIsolationFromMetrics([]L0ChannelMetrics{
		{Channel: ChannelConsensus, EnqueuedCount: 1, DroppedCount: 1, BytesEnqueued: 100},
		{Channel: ChannelService, EnqueuedCount: 1, SentCount: 1, BytesEnqueued: 100, BytesSent: 100},
	}), "consensus traffic")

	bandwidth, err := ComputeChannelBandwidthMetrics([]L0ChannelMetrics{
		{Channel: ChannelService, EnqueuedCount: 2, SentCount: 1, DroppedCount: 1, BytesEnqueued: 2000, BytesSent: 1000},
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1000), bandwidth[0].BytesDropped)
	require.Equal(t, uint32(5_000), bandwidth[0].UsageBps)
}

func TestCosmosCometBFTCompatibilityPlanPreservesRequiredSurfaces(t *testing.T) {
	plan := DefaultCosmosCometBFTCompatibilityPlan()
	require.NoError(t, ValidateCosmosCometBFTCompatibilityPlan(plan))

	plan.PreserveCometBFTConsensusMessages = false
	require.ErrorContains(t, ValidateCosmosCometBFTCompatibilityPlan(plan), "consensus messages")

	plan = DefaultCosmosCometBFTCompatibilityPlan()
	plan.Surfaces = plan.Surfaces[:len(plan.Surfaces)-1]
	require.ErrorContains(t, ValidateCosmosCometBFTCompatibilityPlan(plan), "state_sync_snapshots")
}

func TestABCICompatibilityPipelineUsesCommittedSchedulesAndRoots(t *testing.T) {
	overlayID := HashParts("abci-compat-overlay")
	origin := HashParts("abci-compat-origin")
	destination := HashParts("abci-compat-destination")
	mesh, err := NewAetherMeshMessage(AetherMeshMessage{
		Type:			MeshMessageExecution,
		Payload:		[]byte("abci-compat-execution-message"),
		Origin:			origin,
		Destination:		destination,
		Priority:		2,
		TTL:			50,
		OverlayID:		overlayID,
		DestinationZone:	"zone-a",
		Sequence:		10,
	})
	require.NoError(t, err)
	schedule, err := NewExecutionMessageSchedule(ExecutionMessageSchedule{
		ZoneID:			"zone-a",
		ShardID:		"shard-1",
		RoutingClass:		ExecutionRoutingExecutionOverlay,
		Committed:		true,
		Ordered:		true,
		MessageIDs:		[]string{mesh.MessageID},
		FirstZoneSequence:	10,
		LastZoneSequence:	10,
	})
	require.NoError(t, err)
	hint := ANAProposalHint{
		HintID:			HashParts("abci-compat-hint"),
		ScheduleID:		schedule.ScheduleID,
		ScheduleHash:		schedule.ScheduleHash,
		ZoneID:			schedule.ZoneID,
		ShardID:		schedule.ShardID,
		DeterminismSource:	DeterminismCommittedState,
		CommittedStateDerived:	true,
		UsedForOrdering:	true,
		Priority:		1,
		BlockSTMGroupID:	HashParts("abci-compat-blockstm"),
	}

	prepared, err := BuildPrepareProposalCompatibility(ABCIPrepareProposalInput{
		Height:		50,
		Adapter:	DefaultAetherNetworkingAdapter(),
		Schedules:	[]ExecutionMessageSchedule{schedule},
		Hints:		[]ANAProposalHint{hint},
	})
	require.NoError(t, err)
	require.Equal(t, ABCIPrepareProposal, prepared.Phase)
	require.Equal(t, uint64(1), prepared.MessageCount)
	require.Equal(t, hint.BlockSTMGroupID, prepared.Groups[0].BlockSTMGroupID)
	require.Equal(t, ComputeABCIProposalScheduleRoot(prepared.Groups), prepared.ScheduleRoot)
	require.Equal(t, ComputeABCIOrderingCommitment(prepared.Groups), prepared.OrderingCommitment)

	processed, err := ProcessProposalCompatibility(ABCIProcessProposalInput{
		Height:				50,
		Proposal:			prepared,
		ExpectedScheduleRoot:		prepared.ScheduleRoot,
		ExpectedOrderingCommitment:	prepared.OrderingCommitment,
		VerifiesOrderingCommitment:	true,
	})
	require.NoError(t, err)
	require.Equal(t, ABCIProcessProposal, processed.Phase)

	execMsg := ExecutionZoneMessage{
		Message:		mesh,
		RoutingClass:		ExecutionRoutingExecutionOverlay,
		ZoneID:			schedule.ZoneID,
		ShardID:		schedule.ShardID,
		ExecutionOverlayID:	overlayID,
		ZoneSequence:		10,
		ConsensusScheduleID:	schedule.ScheduleID,
		ConsensusScheduleHash:	schedule.ScheduleHash,
		ConsensusScheduleOrder:	1,
		DeterministicOrdering:	true,
		BlockSTMGroupID:	hint.BlockSTMGroupID,
	}
	result, err := FinalizeBlockCompatibility(ABCIFinalizeBlockInput{
		Height:			50,
		Proposal:		processed,
		ExecutionMessages:	[]ExecutionZoneMessage{execMsg},
	})
	require.NoError(t, err)
	require.Equal(t, []string{mesh.MessageID}, result.ExecutedMessageIDs)
	require.Equal(t, prepared.ScheduleRoot, result.ScheduleRoot)
	require.NotEmpty(t, result.ExecutionRoot)
	require.NotEmpty(t, result.ReceiptsRoot)
}

func TestABCICompatibilityRejectsPeerLocalOrderingUncommittedAndLiveState(t *testing.T) {
	overlayID := HashParts("abci-reject-overlay")
	origin := HashParts("abci-reject-origin")
	destination := HashParts("abci-reject-destination")
	mesh, err := NewAetherMeshMessage(AetherMeshMessage{
		Type:			MeshMessageExecution,
		Payload:		[]byte("abci-reject-execution-message"),
		Origin:			origin,
		Destination:		destination,
		Priority:		2,
		TTL:			50,
		OverlayID:		overlayID,
		DestinationZone:	"zone-b",
		Sequence:		11,
	})
	require.NoError(t, err)
	schedule, err := NewExecutionMessageSchedule(ExecutionMessageSchedule{
		ZoneID:			"zone-b",
		ShardID:		"shard-1",
		RoutingClass:		ExecutionRoutingExecutionOverlay,
		Committed:		true,
		Ordered:		true,
		MessageIDs:		[]string{mesh.MessageID},
		FirstZoneSequence:	11,
		LastZoneSequence:	11,
	})
	require.NoError(t, err)

	_, err = BuildPrepareProposalCompatibility(ABCIPrepareProposalInput{
		Height:		60,
		Adapter:	DefaultAetherNetworkingAdapter(),
		Schedules:	[]ExecutionMessageSchedule{schedule},
		Hints: []ANAProposalHint{{
			HintID:			HashParts("abci-peer-local-hint"),
			ScheduleID:		schedule.ScheduleID,
			ScheduleHash:		schedule.ScheduleHash,
			ZoneID:			schedule.ZoneID,
			ShardID:		schedule.ShardID,
			DeterminismSource:	DeterminismAdvisoryPeerMetric,
			CommittedStateDerived:	false,
			PeerLocal:		true,
			UsedForOrdering:	true,
		}},
	})
	require.ErrorContains(t, err, "peer-local")

	uncommitted := schedule
	uncommitted.Committed = false
	uncommitted.ScheduleHash = ComputeExecutionMessageScheduleHash(uncommitted)
	uncommitted.ScheduleID = ComputeExecutionMessageScheduleID(uncommitted)
	_, err = BuildPrepareProposalCompatibility(ABCIPrepareProposalInput{
		Height:		60,
		Adapter:	DefaultAetherNetworkingAdapter(),
		Schedules:	[]ExecutionMessageSchedule{uncommitted},
	})
	require.ErrorContains(t, err, "committed deterministic")

	prepared, err := BuildPrepareProposalCompatibility(ABCIPrepareProposalInput{
		Height:		60,
		Adapter:	DefaultAetherNetworkingAdapter(),
		Schedules:	[]ExecutionMessageSchedule{schedule},
	})
	require.NoError(t, err)
	_, err = ProcessProposalCompatibility(ABCIProcessProposalInput{
		Height:				60,
		Proposal:			prepared,
		ExpectedScheduleRoot:		prepared.ScheduleRoot,
		ExpectedOrderingCommitment:	prepared.OrderingCommitment,
		DependsOnPeerLocalData:		true,
		VerifiesOrderingCommitment:	true,
	})
	require.ErrorContains(t, err, "peer-local")

	_, err = ProcessProposalCompatibility(ABCIProcessProposalInput{
		Height:				60,
		Proposal:			prepared,
		ExpectedScheduleRoot:		prepared.ScheduleRoot,
		ExpectedOrderingCommitment:	prepared.OrderingCommitment,
		VerifiesOrderingCommitment:	false,
	})
	require.ErrorContains(t, err, "ordering commitments")

	_, err = FinalizeBlockCompatibility(ABCIFinalizeBlockInput{
		Height:			60,
		Proposal:		prepared,
		LiveNetworkStateRead:	true,
	})
	require.ErrorContains(t, err, "live network state")

	uncommittedExec := ExecutionZoneMessage{
		Message:		mesh,
		RoutingClass:		ExecutionRoutingExecutionOverlay,
		ZoneID:			schedule.ZoneID,
		ShardID:		schedule.ShardID,
		ExecutionOverlayID:	overlayID,
		ZoneSequence:		11,
	}
	_, err = FinalizeBlockCompatibility(ABCIFinalizeBlockInput{
		Height:			60,
		Proposal:		prepared,
		ExecutionMessages:	[]ExecutionZoneMessage{uncommittedExec},
	})
	require.ErrorContains(t, err, "committed messages only")
}

func TestBlockSTMNetworkAssistGroupsZoneShardAndQueuesCrossZone(t *testing.T) {
	overlayID := HashParts("blockstm-api-overlay")
	origin := HashParts("blockstm-api-origin")
	destination := HashParts("blockstm-api-destination")
	txID := HashParts("blockstm-api-tx")
	crossMesh, err := NewAetherMeshMessage(AetherMeshMessage{
		Type:			MeshMessageCrossZone,
		Payload:		[]byte("blockstm-cross-zone-message"),
		Origin:			origin,
		Destination:		destination,
		Priority:		PriorityForChannel(ChannelExecution),
		TTL:			50,
		OverlayID:		overlayID,
		SourceZone:		"zone-a",
		DestinationZone:	"zone-b",
		Sequence:		12,
	})
	require.NoError(t, err)
	schedule, err := NewExecutionMessageSchedule(ExecutionMessageSchedule{
		ZoneID:			"zone-b",
		ShardID:		"shard-7",
		RoutingClass:		ExecutionRoutingExecutionOverlay,
		Committed:		true,
		Ordered:		true,
		TransactionIDs:		[]string{txID},
		MessageIDs:		[]string{crossMesh.MessageID},
		FirstZoneSequence:	12,
		LastZoneSequence:	12,
	})
	require.NoError(t, err)
	hint := ANAProposalHint{
		HintID:			HashParts("blockstm-api-hint"),
		ScheduleID:		schedule.ScheduleID,
		ScheduleHash:		schedule.ScheduleHash,
		ZoneID:			schedule.ZoneID,
		ShardID:		schedule.ShardID,
		DeterminismSource:	DeterminismDeterministicProof,
		CommittedStateDerived:	true,
		UsedForOrdering:	true,
		Priority:		PriorityForChannel(ChannelExecution),
		BlockSTMGroupID:	HashParts("blockstm-api-group"),
		DeterministicHintProof:	HashParts("blockstm-api-route-proof"),
	}
	cross := ExecutionZoneMessage{
		Message:		crossMesh,
		RoutingClass:		ExecutionRoutingExecutionOverlay,
		ZoneID:			"zone-b",
		ShardID:		"shard-7",
		ExecutionOverlayID:	overlayID,
		ZoneSequence:		12,
		CrossZone: CrossZoneMessage{
			SourceZone:		"zone-a",
			DestinationZone:	"zone-b",
			SourceSequence:		12,
			MessageHash:		crossMesh.PayloadHash,
			ExpiryHeight:		90,
			ReceiptPolicy:		ReceiptPolicyOnExecution,
		},
	}

	plan, err := BuildBlockSTMNetworkAssistPlan(BlockSTMNetworkAssistInput{
		Height:			60,
		Schedules:		[]ExecutionMessageSchedule{schedule},
		Hints:			[]ANAProposalHint{hint},
		CrossZoneMessages:	[]ExecutionZoneMessage{cross},
	})
	require.NoError(t, err)
	require.True(t, plan.PrioritizesExecutionOverlayTraffic)
	require.True(t, plan.PropagatesZoneShardRouteHints)
	require.True(t, plan.DeliversCrossZoneExecutionQueues)
	require.False(t, plan.DecidesCommittedConflicts)
	require.Len(t, plan.Groups, 1)
	require.Equal(t, "zone-b", plan.Groups[0].RouteHint.ZoneID)
	require.Equal(t, "shard-7", plan.Groups[0].RouteHint.ShardID)
	require.Equal(t, hint.DeterministicHintProof, plan.Groups[0].RouteHint.DeterministicHintHash)
	require.Equal(t, PriorityForChannel(ChannelExecution), plan.Groups[0].Priority)
	require.Len(t, plan.CrossZoneDeliveries, 1)
	require.Equal(t, crossMesh.MessageID, plan.CrossZoneDeliveries[0].MessageID)
	require.Equal(t, "zone-b", plan.CrossZoneDeliveries[0].DestinationZone)

	_, err = BuildBlockSTMNetworkAssistPlan(BlockSTMNetworkAssistInput{
		Height:				60,
		Schedules:			[]ExecutionMessageSchedule{schedule},
		DecidesCommittedConflicts:	true,
	})
	require.ErrorContains(t, err, "conflicts")
}

func TestNetworkingAPIIntegrationBuildsDiagnosticsProofsAndRouteHints(t *testing.T) {
	require.NoError(t, DefaultNetworkingQueryServiceDescriptor().Validate())
	broken := DefaultNetworkingQueryServiceDescriptor()
	broken.RouteHintEndpoint = false
	require.ErrorContains(t, broken.Validate(), "route hints")

	salt := []byte("aetra-api-network")
	service := signedNodeRecordWithCapabilities(t, 0x83, salt, 100, []NodeRole{NodeRoleService, NodeRoleFull}, []string{"zone-api"}, []string{"svc.api"})
	stateSync := signedNodeRecordWithCapabilities(t, 0x84, salt, 100, []NodeRole{NodeRoleStateSync}, []string{"zone-api"}, []string{"state-sync"})
	nodeResponse, err := BuildNodeNetworkingQueryResponse(NodeNetworkingQueryRequest{
		CurrentHeight:	80,
		NetworkSalt:	salt,
		Role:		NodeRoleService,
		ZoneID:		"zone-api",
		ServiceID:	"svc.api",
	}, []NodeRecord{service, stateSync})
	require.NoError(t, err)
	require.Len(t, nodeResponse.Nodes, 1)
	require.Equal(t, service.NodeID, nodeResponse.Nodes[0].NodeID)
	require.NotEmpty(t, nodeResponse.ResultHash)

	desc := testDefaultOverlayDescriptor(t, OverlayTypeExecution)
	graph := NormalizeRoutingGraph(RoutingGraph{
		OverlayID:	desc.OverlayID,
		Version:	1,
		Committed:	true,
		Edges: []RoutingEdge{{
			FromNodeID:	service.NodeID,
			ToNodeID:	stateSync.NodeID,
			Weight:		1,
			Priority:	PriorityForChannel(ChannelExecution),
			ZoneID:		"zone-api",
		}},
	})
	overlayResponse, err := BuildOverlayDiagnosticsResponse(OverlayDiagnosticsRequest{
		OverlayID:	desc.OverlayID,
		CurrentHeight:	80,
	}, []OverlayDescriptor{desc}, []OverlayMembershipRecord{{
		OverlayID:	desc.OverlayID,
		NodeID:		service.NodeID,
		ProofID:	HashParts("api-membership-proof"),
		Membership:	desc.Membership,
		Mode:		OverlayMembershipModeZoneAssignment,
		JoinedHeight:	70,
		ExpiresHeight:	100,
	}}, graph, []L3OverlayMetrics{{OverlayID: desc.OverlayID, QueuedCount: 3}}, []RouteFailureSample{{OverlayID: desc.OverlayID, Attempts: 10, Failures: 1}})
	require.NoError(t, err)
	require.Equal(t, uint64(1), overlayResponse.MembershipSize)
	require.Equal(t, uint64(3), overlayResponse.QueuedMessages)
	require.Equal(t, uint32(1_000), overlayResponse.RouteFailureRateBps)

	stream, err := NewStreamSession(StreamSession{
		SessionID:		HashParts("api-session"),
		PayloadType:		StreamingPayloadStateSync,
		Priority:		PriorityForChannel(ChannelStateSync),
		FlowControlWindow:	4096,
		ChunkSize:		1024,
		Parallelism:		2,
		BytesSent:		2048,
		BytesAcknowledged:	1024,
		State:			StreamStateActive,
	})
	require.NoError(t, err)
	streamMetrics, err := ComputeStreamMetrics(stream, 1000, 2, 1)
	require.NoError(t, err)
	streamResponse, err := BuildStreamDiagnosticsResponse(StreamDiagnosticsRequest{
		StreamID:	stream.StreamID,
		PayloadType:	StreamingPayloadStateSync,
	}, []StreamSession{stream}, []StreamMetrics{streamMetrics})
	require.NoError(t, err)
	require.Equal(t, stream.StreamID, streamResponse.StreamID)
	require.Equal(t, uint64(2), streamResponse.StallCount)
	require.NotEmpty(t, streamResponse.ResultHash)

	table := EmptyDistributedRoutingTable()
	record := testSignedDiscoveryObjectRecord(t, service, 0x83, salt, DRTObjectServiceEndpoint, HashParts("api-target"), HashParts("api-ad"), "", "svc.api", "", 100)
	table, err = table.Store(record, salt, 80)
	require.NoError(t, err)
	discoveryResultHash, err := ComputeDiscoveryResponseResultHash([]DiscoveryRecord{record})
	require.NoError(t, err)
	stateRoot := HashParts("api-state-root")
	proofResponse, err := BuildDiscoveryProofAPIResponse(DiscoveryProofAPIRequest{
		Query:		DRTQuery{ObjectType: DRTObjectServiceEndpoint, ServiceID: "svc.api", Limit: 4},
		CurrentHeight:	80,
		OnChainProof: DiscoveryOnChainProof{
			ProofHash:	ComputeDiscoveryOnChainProofHash(discoveryResultHash, stateRoot, 80),
			ProofHeight:	80,
			StateRoot:	stateRoot,
		},
	}, table, service, deterministicPrivateKey(0x83), salt)
	require.NoError(t, err)
	require.False(t, proofResponse.Response.AdvisoryOnly)
	require.Equal(t, proofResponse.Response.ResultHash, proofResponse.ResultHash)

	networkMsg, err := NewNetworkMessage(NetworkMessage{
		Layer:			LayerL3Application,
		Channel:		ChannelExecution,
		ConsensusEffect:	false,
		DeterminismSource:	DeterminismNone,
		PayloadHash:		HashParts("api-route-payload"),
		PayloadSizeBytes:	512,
	})
	require.NoError(t, err)
	routeResponse, err := BuildRouteHintAPIResponse(RouteHintAPIRequest{
		Message:	networkMsg,
		Descriptor:	desc,
		Graph:		graph,
		ClientZoneID:	"zone-api",
		ClientShardID:	"shard-api",
	})
	require.NoError(t, err)
	require.True(t, routeResponse.AdvisoryOnly)
	require.Equal(t, "zone-api", routeResponse.Hint.ZoneID)
	require.Equal(t, "shard-api", routeResponse.Hint.ShardID)
	require.NotEmpty(t, routeResponse.ResultHash)
}

func TestNetworkingComponentMapCoversNodeComponentsAndSupportModules(t *testing.T) {
	componentMap := DefaultNetworkingComponentMap()
	require.NoError(t, ValidateNetworkingComponentMap(componentMap))
	require.Len(t, componentMap.NodeComponents, 7)
	require.Len(t, componentMap.SupportModules, 5)
	require.Equal(t, ComputeNetworkingComponentMapRoot(componentMap), componentMap.MapRoot)

	ana, found, err := ComponentByName(componentMap, ComponentANA)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, LayerL0Physical, ana.Layer)
	require.True(t, ana.RuntimeOnly)
	require.True(t, ana.AdvisoryUntilCommitted)
	require.False(t, ana.WritesCommittedState)
	require.Contains(t, ana.Channels, ChannelConsensus)

	mesh, found, err := ComponentByName(componentMap, ComponentMesh)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, LayerL3Application, mesh.Layer)
	require.Contains(t, mesh.DependsOn, ComponentOverlayMgr)
	require.Contains(t, mesh.DependsOn, ComponentRL2)

	messages, found, err := SupportModuleByName(componentMap, SupportModuleMessages)
	require.NoError(t, err)
	require.True(t, found)
	require.True(t, messages.OwnsCommittedState)
	require.True(t, messages.ConsumesNetworkProofs)
	require.Contains(t, messages.StateObjects, "cross-zone message receipts")
	require.Contains(t, messages.StateObjects, "replay protection")
}

func TestNetworkingComponentMapRejectsUnsafeBoundariesAndMissingLinks(t *testing.T) {
	componentMap := DefaultNetworkingComponentMap()
	componentMap.NodeComponents[0].WritesCommittedState = true
	componentMap.MapRoot = ComputeNetworkingComponentMapRoot(componentMap)
	require.ErrorContains(t, ValidateNetworkingComponentMap(componentMap), "committed state")

	componentMap = DefaultNetworkingComponentMap()
	for i := range componentMap.NodeComponents {
		if componentMap.NodeComponents[i].Component == ComponentMesh {
			componentMap.NodeComponents[i].DependsOn = []NodeSideComponent{ComponentOverlayMgr, "missing"}
		}
	}
	componentMap.MapRoot = ComputeNetworkingComponentMapRoot(componentMap)
	require.ErrorContains(t, ValidateNetworkingComponentMap(componentMap), "missing dependency")

	componentMap = DefaultNetworkingComponentMap()
	componentMap.SupportModules[0].AllowsExternalNetworkCall = true
	componentMap.MapRoot = ComputeNetworkingComponentMapRoot(componentMap)
	require.ErrorContains(t, ValidateNetworkingComponentMap(componentMap), "external network calls")

	componentMap = DefaultNetworkingComponentMap()
	filteredModules := make([]OnChainSupportModuleSpec, 0, len(componentMap.SupportModules)-1)
	for _, module := range componentMap.SupportModules {
		if module.Module != SupportModuleMessages {
			filteredModules = append(filteredModules, module)
		}
	}
	componentMap.SupportModules = filteredModules
	componentMap.MapRoot = ComputeNetworkingComponentMapRoot(componentMap)
	require.ErrorContains(t, ValidateNetworkingComponentMap(componentMap), "x/messages")

	componentMap = DefaultNetworkingComponentMap()
	componentMap.MapRoot = HashParts("wrong-component-map-root")
	require.ErrorContains(t, ValidateNetworkingComponentMap(componentMap), "root mismatch")
}

func TestXNetworkStateKeysMessagesAndQueries(t *testing.T) {
	salt := []byte("aetra-x-network-state")
	node := signedNodeRecord(t, 0x85, salt, 100, NodeRoleFull, NodeRoleRouting)
	service := signedNodeRecordWithCapabilities(t, 0x86, salt, 100, []NodeRole{NodeRoleService}, []string{"zone-net"}, []string{"svc.net"})
	state := EmptyState()
	var err error
	state, err = RegisterNodeRecord(state, node, salt, 10)
	require.NoError(t, err)
	state, err = RegisterNodeRecord(state, service, salt, 10)
	require.NoError(t, err)
	state, err = RegisterRoleCommitment(state, RoleCommitment{
		NodeID:		service.NodeID,
		Role:		NodeRoleService,
		Bonded:		true,
		CommitmentHash:	HashParts("x-network-role-commitment"),
		ExpiresHeight:	90,
	}, 11)
	require.NoError(t, err)
	discovery := testSignedDiscoveryObjectRecord(t, service, 0x86, salt, DRTObjectServiceEndpoint, HashParts("x-network-discovery-target"), HashParts("x-network-discovery-ad"), "", "svc.net", "", 90)
	evidence, err := NewNetworkEvidenceRecord(NetworkEvidenceRecord{
		EvidenceType:	NetworkEvidenceConflictingBroadcast,
		ReporterNodeID:	node.NodeID,
		SubjectNodeID:	service.NodeID,
		EvidenceHash:	HashParts("x-network-evidence-hash"),
		EvidenceHeight:	20,
		PayloadBytes:	512,
		Committed:	true,
	})
	require.NoError(t, err)
	reputation := NetworkReputationRecord{
		NodeID:			service.NodeID,
		Score:			PeerScore{ScoreBps: 8_000, LatencyBps: 8_500, ReliabilityBps: 9_000, ThroughputBps: 7_500, PenaltyBps: 250},
		LastUpdatedHeight:	20,
		EvidenceHash:		evidence.EvidenceHash,
		ConsensusEligible:	true,
	}
	params := DefaultXNetworkParams(salt)
	xstate, err := NewXNetworkState(params, state, []DiscoveryRecord{discovery}, []NetworkReputationRecord{reputation}, []NetworkEvidenceRecord{evidence})
	require.NoError(t, err)
	require.NotEmpty(t, xstate.StateRoot)

	keys, err := BuildXNetworkStateKeys(xstate)
	require.NoError(t, err)
	require.Equal(t, "network/params", keys.ParamsKey)
	require.Contains(t, keys.NodeKeys, "network/nodes/"+service.NodeID)
	require.Contains(t, keys.RoleKeys, "network/roles/SERVICE_NODE/"+service.NodeID)
	require.Contains(t, keys.DiscoveryKeys, "network/discovery/"+discovery.RecordID)
	require.Contains(t, keys.ReputationKeys, "network/reputation/"+service.NodeID)
	require.Contains(t, keys.EvidenceKeys, "network/evidence/"+evidence.EvidenceID)
	require.Equal(t, ComputeXNetworkStateKeysRoot(keys), keys.StateKeysRoot)
	require.NoError(t, ValidateXNetworkStateKey(XNetworkRoleStateKey(NodeRoleService, service.NodeID)))

	require.NoError(t, MsgRegisterNodeRequest{SignerNodeID: node.NodeID, Record: node, NetworkSalt: salt, CurrentHeight: 10}.ValidateBasic())
	require.NoError(t, MsgUpdateNodeRequest{SignerNodeID: service.NodeID, Record: service, NetworkSalt: salt, CurrentHeight: 10}.ValidateBasic())
	require.NoError(t, MsgRenewNodeRequest{SignerNodeID: service.NodeID, Record: service, NetworkSalt: salt, CurrentHeight: 10}.ValidateBasic())
	require.NoError(t, MsgRevokeNodeRequest{SignerNodeID: service.NodeID, NodeID: service.NodeID, ReasonHash: HashParts("x-network-revoke-reason"), CurrentHeight: 20}.ValidateBasic())
	require.NoError(t, MsgSubmitNetworkEvidenceRequest{SignerNodeID: node.NodeID, Evidence: evidence}.ValidateBasic(params))

	foundNode, found, err := QueryNodeFromXNetworkState(xstate, QueryNodeRequest{NodeID: service.NodeID})
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, service.NodeID, foundNode.NodeID)
	services, err := QueryNodesByRoleFromXNetworkState(xstate, QueryNodesByRoleRequest{Role: NodeRoleService, CurrentHeight: 20})
	require.NoError(t, err)
	require.Len(t, services, 1)
	overlay, found, err := QueryOverlayFromXNetworkState(xstate, QueryOverlayRequest{OverlayID: state.OverlayDescriptors[0].OverlayID})
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, state.OverlayDescriptors[0].OverlayID, overlay.OverlayID)
	foundDiscovery, found, err := QueryDiscoveryRecordFromXNetworkState(xstate, QueryDiscoveryRecordRequest{RecordID: discovery.RecordID})
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, discovery.RecordID, foundDiscovery.RecordID)
	foundParams, err := QueryNetworkParamsFromXNetworkState(xstate, QueryNetworkParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, params.NetworkSaltHash, foundParams.NetworkSaltHash)
	foundEvidence, found, err := QueryNetworkEvidenceFromXNetworkState(xstate, QueryNetworkEvidenceRequest{EvidenceID: evidence.EvidenceID})
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, evidence.EvidenceID, foundEvidence.EvidenceID)
}

func TestXNetworkStateRejectsInvalidKeysMessagesAndConsensusReputation(t *testing.T) {
	salt := []byte("aetra-x-network-invalid")
	node := signedNodeRecord(t, 0x87, salt, 100, NodeRoleFull)
	params := DefaultXNetworkParams(salt)
	state, err := NewXNetworkState(params, EmptyState(), nil, nil, nil)
	require.NoError(t, err)
	state.StateRoot = HashParts("wrong-x-network-root")
	require.ErrorContains(t, state.Validate(), "root mismatch")

	require.ErrorContains(t, ValidateXNetworkStateKey("network/roles/BAD/"+node.NodeID), "unknown")
	require.ErrorContains(t, MsgRegisterNodeRequest{SignerNodeID: HashParts("wrong-signer"), Record: node, NetworkSalt: salt, CurrentHeight: 10}.ValidateBasic(), "signer")
	require.ErrorContains(t, MsgRevokeNodeRequest{SignerNodeID: node.NodeID, NodeID: HashParts("other-node"), ReasonHash: HashParts("reason"), CurrentHeight: 20}.ValidateBasic(), "signer")

	evidence, err := NewNetworkEvidenceRecord(NetworkEvidenceRecord{
		EvidenceType:	NetworkEvidenceChunkCorruption,
		ReporterNodeID:	node.NodeID,
		SubjectNodeID:	HashParts("subject"),
		EvidenceHash:	HashParts("evidence"),
		EvidenceHeight:	20,
		PayloadBytes:	params.MaxEvidenceBytes + 1,
	})
	require.ErrorContains(t, err, "payload bytes")
	require.Empty(t, evidence.EvidenceID)

	badReputation := NetworkReputationRecord{
		NodeID:			node.NodeID,
		Score:			PeerScore{ScoreBps: BasisPoints + 1},
		LastUpdatedHeight:	20,
	}
	_, err = NewXNetworkState(params, EmptyState(), nil, []NetworkReputationRecord{badReputation}, nil)
	require.ErrorContains(t, err, "peer score")

	consensusReputation := NetworkReputationRecord{
		NodeID:			node.NodeID,
		Score:			PeerScore{ScoreBps: 5_000},
		LastUpdatedHeight:	20,
		ConsensusEligible:	true,
	}
	_, err = NewXNetworkState(params, EmptyState(), nil, []NetworkReputationRecord{consensusReputation}, nil)
	require.ErrorContains(t, err, "evidence hash")
}

func TestNetworkingImplementationRoadmapValidatesPhasesTasksAndExitCriteria(t *testing.T) {
	roadmap := DefaultNetworkingImplementationRoadmap()
	require.NoError(t, ValidateNetworkingImplementationRoadmap(roadmap))
	require.Len(t, roadmap.Phases, 9)
	require.Equal(t, ComputeNetworkingRoadmapRoot(roadmap), roadmap.RoadmapRoot)

	phase0 := roadmap.Phases[0]
	require.Equal(t, RoadmapPhaseBaselineInstrumentation, phase0.Phase)
	require.Contains(t, phase0.Tasks, RoadmapTaskInventoryCometBFTP2P)
	require.Contains(t, phase0.Tasks, RoadmapTaskNetworkParameterSchema)
	require.Contains(t, phase0.ExitCriteria, ExitCurrentBehaviorMeasurable)

	broken := roadmap
	broken.Phases[0].Tasks = broken.Phases[0].Tasks[:len(broken.Phases[0].Tasks)-1]
	broken.RoadmapRoot = ComputeNetworkingRoadmapRoot(broken)
	require.ErrorContains(t, ValidateNetworkingImplementationRoadmap(broken), "missing task")

	broken = roadmap
	broken.RoadmapRoot = HashParts("wrong-roadmap-root")
	require.ErrorContains(t, ValidateNetworkingImplementationRoadmap(broken), "root mismatch")
}

func TestNetworkingRoadmapReadinessForPhasesZeroToEight(t *testing.T) {
	evidence := testRoadmapEvidence(t)

	phase0, err := EvaluateRoadmapPhaseReadiness(RoadmapPhaseBaselineInstrumentation, evidence)
	require.NoError(t, err)
	require.True(t, phase0.Ready)
	require.Contains(t, phase0.SatisfiedExitCriteria, ExitCurrentBehaviorMeasurable)
	require.Contains(t, phase0.SatisfiedExitCriteria, ExitBaselineMetricsExist)
	require.NotEmpty(t, phase0.ReportHash)

	phase1, err := EvaluateRoadmapPhaseReadiness(RoadmapPhaseAetherNetworkingAdapter, evidence)
	require.NoError(t, err)
	require.True(t, phase1.Ready)
	require.Contains(t, phase1.SatisfiedExitCriteria, ExitConsensusProtectedPriority)
	require.Contains(t, phase1.SatisfiedExitCriteria, ExitPeerScoringChannelMetrics)
	require.Contains(t, phase1.SatisfiedExitCriteria, ExitServiceCannotStarveConsensus)

	phase2, err := EvaluateRoadmapPhaseReadiness(RoadmapPhaseNodeIdentitySessions, evidence)
	require.NoError(t, err)
	require.True(t, phase2.Ready)
	require.Contains(t, phase2.SatisfiedExitCriteria, ExitCryptographicNodeAuth)
	require.Contains(t, phase2.SatisfiedExitCriteria, ExitLogicalStreamsShareSession)
	require.Contains(t, phase2.SatisfiedExitCriteria, ExitExpiredForgedRecordsRejected)

	phase3, err := EvaluateRoadmapPhaseReadiness(RoadmapPhaseOverlayRouting, evidence)
	require.NoError(t, err)
	require.True(t, phase3.Ready)
	require.Contains(t, phase3.SatisfiedExitCriteria, ExitOverlayJoinSupported)
	require.Contains(t, phase3.SatisfiedExitCriteria, ExitCommittedRoutesReproducible)
	require.Contains(t, phase3.SatisfiedExitCriteria, ExitPeerRotationConnectivity)

	phase4, err := EvaluateRoadmapPhaseReadiness(RoadmapPhaseRL2Streaming, evidence)
	require.NoError(t, err)
	require.True(t, phase4.Ready)
	require.Contains(t, phase4.SatisfiedExitCriteria, ExitChunkedStreamingPayloads)
	require.Contains(t, phase4.SatisfiedExitCriteria, ExitInterruptedTransfersResume)
	require.Contains(t, phase4.SatisfiedExitCriteria, ExitInvalidChunksRejected)

	phase5, err := EvaluateRoadmapPhaseReadiness(RoadmapPhaseDiscoveryLayer, evidence)
	require.NoError(t, err)
	require.True(t, phase5.Ready)
	require.Contains(t, phase5.SatisfiedExitCriteria, ExitDiscoveryObjectsDiscoverable)
	require.Contains(t, phase5.SatisfiedExitCriteria, ExitDiscoveryRecordsExpireVerify)
	require.Contains(t, phase5.SatisfiedExitCriteria, ExitForgedExpiredRecordsRejected)

	phase6, err := EvaluateRoadmapPhaseReadiness(RoadmapPhaseHybridBroadcast, evidence)
	require.NoError(t, err)
	require.True(t, phase6.Ready)
	require.Contains(t, phase6.SatisfiedExitCriteria, ExitBlocksHeaderChunksProofSet)
	require.Contains(t, phase6.SatisfiedExitCriteria, ExitDuplicateConflictingHandled)
	require.Contains(t, phase6.SatisfiedExitCriteria, ExitFallbackGossipResilient)

	phase7, err := EvaluateRoadmapPhaseReadiness(RoadmapPhaseAetherMesh, evidence)
	require.NoError(t, err)
	require.True(t, phase7.Ready)
	require.Contains(t, phase7.SatisfiedExitCriteria, ExitL3MessageClassesSupported)
	require.Contains(t, phase7.SatisfiedExitCriteria, ExitCrossZoneDeliverySemantics)
	require.Contains(t, phase7.SatisfiedExitCriteria, ExitReceiptsVisibleProofQueryable)

	phase8, err := EvaluateRoadmapPhaseReadiness(RoadmapPhaseSecurityLoadHardening, evidence)
	require.NoError(t, err)
	require.True(t, phase8.Ready)
	require.Contains(t, phase8.SatisfiedExitCriteria, ExitMaliciousPeersIsolated)
	require.Contains(t, phase8.SatisfiedExitCriteria, ExitCriticalChannelsUnderFlood)
	require.Contains(t, phase8.SatisfiedExitCriteria, ExitDiscoveryPoisoningDetected)
}

func TestNetworkingRoadmapReadinessRejectsMissingEvidence(t *testing.T) {
	evidence := testRoadmapEvidence(t)
	evidence.PerformanceSnapshot.ChannelBandwidth = nil
	phase0, err := EvaluateRoadmapPhaseReadiness(RoadmapPhaseBaselineInstrumentation, evidence)
	require.NoError(t, err)
	require.False(t, phase0.Ready)
	require.NotContains(t, phase0.SatisfiedExitCriteria, ExitBaselineMetricsExist)

	evidence = testRoadmapEvidence(t)
	evidence.PerformanceSnapshot.ServiceTrafficIsolated = false
	phase1, err := EvaluateRoadmapPhaseReadiness(RoadmapPhaseAetherNetworkingAdapter, evidence)
	require.NoError(t, err)
	require.False(t, phase1.Ready)
	require.NotContains(t, phase1.SatisfiedExitCriteria, ExitServiceCannotStarveConsensus)

	evidence = testRoadmapEvidence(t)
	evidence.HandshakeReplayRejected = false
	phase2, err := EvaluateRoadmapPhaseReadiness(RoadmapPhaseNodeIdentitySessions, evidence)
	require.NoError(t, err)
	require.False(t, phase2.Ready)
	require.NotContains(t, phase2.SatisfiedExitCriteria, ExitExpiredForgedRecordsRejected)

	evidence = testRoadmapEvidence(t)
	evidence.PeerRotationPreserved = false
	phase3, err := EvaluateRoadmapPhaseReadiness(RoadmapPhaseOverlayRouting, evidence)
	require.NoError(t, err)
	require.False(t, phase3.Ready)
	require.NotContains(t, phase3.SatisfiedExitCriteria, ExitPeerRotationConnectivity)

	evidence = testRoadmapEvidence(t)
	evidence.RL2InvalidChunkRejected = false
	phase4, err := EvaluateRoadmapPhaseReadiness(RoadmapPhaseRL2Streaming, evidence)
	require.NoError(t, err)
	require.False(t, phase4.Ready)
	require.NotContains(t, phase4.SatisfiedExitCriteria, ExitInvalidChunksRejected)

	evidence = testRoadmapEvidence(t)
	evidence.DiscoveryExpiredRejected = false
	phase5, err := EvaluateRoadmapPhaseReadiness(RoadmapPhaseDiscoveryLayer, evidence)
	require.NoError(t, err)
	require.False(t, phase5.Ready)
	require.NotContains(t, phase5.SatisfiedExitCriteria, ExitForgedExpiredRecordsRejected)

	evidence = testRoadmapEvidence(t)
	evidence.BroadcastConflictHandled = false
	phase6, err := EvaluateRoadmapPhaseReadiness(RoadmapPhaseHybridBroadcast, evidence)
	require.NoError(t, err)
	require.False(t, phase6.Ready)
	require.NotContains(t, phase6.SatisfiedExitCriteria, ExitDuplicateConflictingHandled)

	evidence = testRoadmapEvidence(t)
	evidence.CrossZoneExactlyOnce = false
	phase7, err := EvaluateRoadmapPhaseReadiness(RoadmapPhaseAetherMesh, evidence)
	require.NoError(t, err)
	require.False(t, phase7.Ready)
	require.NotContains(t, phase7.SatisfiedExitCriteria, ExitCrossZoneDeliverySemantics)

	evidence = testRoadmapEvidence(t)
	evidence.DiscoveryPoisoningDetected = false
	phase8, err := EvaluateRoadmapPhaseReadiness(RoadmapPhaseSecurityLoadHardening, evidence)
	require.NoError(t, err)
	require.False(t, phase8.Ready)
	require.NotContains(t, phase8.SatisfiedExitCriteria, ExitDiscoveryPoisoningDetected)
}

func TestNetworkingAcceptanceCriteriaDefineImplementationPlanningGate(t *testing.T) {
	spec := DefaultNetworkingAcceptanceCriteriaSpec()
	require.NoError(t, ValidateNetworkingAcceptanceCriteriaSpec(spec))
	require.Len(t, spec.Criteria, 11)
	require.NotEmpty(t, spec.SpecRoot)
	require.Contains(t, networkingAcceptanceCriterionIDs(spec.Criteria), AcceptanceCriterionL0CometBFTProtected)
	require.Contains(t, networkingAcceptanceCriterionIDs(spec.Criteria), AcceptanceCriterionANAChannelQoS)
	require.Contains(t, networkingAcceptanceCriterionIDs(spec.Criteria), AcceptanceCriterionRL2Streaming)
	require.Contains(t, networkingAcceptanceCriterionIDs(spec.Criteria), AcceptanceCriterionSecurityControls)
	require.Contains(t, networkingAcceptanceCriterionIDs(spec.Criteria), AcceptanceCriterionRequiredTestCoverage)

	missing := spec
	missing.Criteria = append([]NetworkingAcceptanceCriterion(nil), spec.Criteria[:len(spec.Criteria)-1]...)
	missing.SpecRoot = ComputeNetworkingAcceptanceCriteriaRoot(missing)
	require.ErrorContains(t, ValidateNetworkingAcceptanceCriteriaSpec(missing), "must define 11 criteria")

	unknown := spec
	unknown.Criteria = append([]NetworkingAcceptanceCriterion(nil), spec.Criteria...)
	unknown.Criteria[0] = NetworkingAcceptanceCriterion("ship_without_cometbft")
	unknown.SpecRoot = ComputeNetworkingAcceptanceCriteriaRoot(unknown)
	require.ErrorContains(t, ValidateNetworkingAcceptanceCriteriaSpec(unknown), "unknown networking acceptance criterion")
}

func TestNetworkingImplementationPlanningReadinessRequiresAllAcceptanceEvidence(t *testing.T) {
	spec := DefaultNetworkingAcceptanceCriteriaSpec()
	evidence := testNetworkingAcceptanceEvidence()
	report, err := EvaluateNetworkingImplementationPlanningReadiness(spec, evidence)
	require.NoError(t, err)
	require.True(t, report.Ready)
	require.Empty(t, report.Missing)
	require.Empty(t, report.Rejected)
	require.NotEmpty(t, report.ReadinessHash)

	missing := make([]NetworkingAcceptanceEvidence, 0, len(evidence)-1)
	for _, item := range evidence {
		if item.Criterion != AcceptanceCriterionDRTDiscovery {
			missing = append(missing, item)
		}
	}
	report, err = EvaluateNetworkingImplementationPlanningReadiness(spec, missing)
	require.NoError(t, err)
	require.False(t, report.Ready)
	require.Contains(t, report.Missing, AcceptanceCriterionDRTDiscovery)

	rejected := append([]NetworkingAcceptanceEvidence(nil), evidence...)
	for i := range rejected {
		if rejected[i].Criterion == AcceptanceCriterionSecurityControls {
			rejected[i].Accepted = false
			break
		}
	}
	report, err = EvaluateNetworkingImplementationPlanningReadiness(spec, rejected)
	require.NoError(t, err)
	require.False(t, report.Ready)
	require.Contains(t, report.Rejected, AcceptanceCriterionSecurityControls)

	emptyEvidence := append([]NetworkingAcceptanceEvidence(nil), evidence...)
	for i := range emptyEvidence {
		if emptyEvidence[i].Criterion == AcceptanceCriterionRequiredTestCoverage {
			emptyEvidence[i].Evidence = nil
			break
		}
	}
	report, err = EvaluateNetworkingImplementationPlanningReadiness(spec, emptyEvidence)
	require.NoError(t, err)
	require.False(t, report.Ready)
	require.Contains(t, report.Rejected, AcceptanceCriterionRequiredTestCoverage)
}

func TestRequiredNetworkingTestCoverageValidatesUnitAndIntegrationMatrix(t *testing.T) {
	coverage := DefaultRequiredNetworkingTestCoverage()
	require.NoError(t, ValidateRequiredNetworkingTestCoverage(coverage))
	require.NotEmpty(t, ComputeNetworkingTestCoverageRoot(coverage))

	var unitCount, integrationCount, securityCount, performanceCount int
	for _, spec := range coverage {
		switch spec.Category {
		case NetworkingTestCoverageUnit:
			unitCount++
		case NetworkingTestCoverageIntegration:
			integrationCount++
		case NetworkingTestCoverageSecurity:
			securityCount++
		case NetworkingTestCoveragePerformance:
			performanceCount++
		}
	}
	require.Equal(t, 10, unitCount)
	require.Equal(t, 9, integrationCount)
	require.Equal(t, 9, securityCount)
	require.Equal(t, 9, performanceCount)
	require.Contains(t, requiredCoverageTests(coverage), RequiredTestNodeIDDerivation)
	require.Contains(t, requiredCoverageTests(coverage), RequiredTestBroadcastDeduplication)
	require.Contains(t, requiredCoverageTests(coverage), RequiredTestHeaderFirstPropagation)
	require.Contains(t, requiredCoverageTests(coverage), RequiredTestCrossZoneReplaySecurity)
	require.Contains(t, requiredCoverageTests(coverage), RequiredTestPeerRotationStability)

	missing := append([]NetworkingTestCoverageSpec(nil), coverage[:len(coverage)-1]...)
	require.ErrorContains(t, ValidateRequiredNetworkingTestCoverage(missing), "must define 37 areas")

	wrongCategory := append([]NetworkingTestCoverageSpec(nil), coverage...)
	for i := range wrongCategory {
		if wrongCategory[i].Test == RequiredTestCrossZoneDelivery {
			wrongCategory[i].Category = NetworkingTestCoverageUnit
			break
		}
	}
	require.ErrorContains(t, ValidateRequiredNetworkingTestCoverage(wrongCategory), "must be integration")
}

func TestNetworkingTestCoverageReportRequiresAllRequiredEvidence(t *testing.T) {
	evidence := testRequiredNetworkingCoverageEvidence()
	report, err := EvaluateNetworkingTestCoverage(evidence)
	require.NoError(t, err)
	require.True(t, report.Ready)
	require.Empty(t, report.Missing)
	require.Empty(t, report.Failed)
	require.NotEmpty(t, report.ReportHash)

	missing := make([]NetworkingTestCoverageEvidence, 0, len(evidence)-1)
	for _, item := range evidence {
		if item.Test != RequiredTestBroadcastDeduplication {
			missing = append(missing, item)
		}
	}
	report, err = EvaluateNetworkingTestCoverage(missing)
	require.NoError(t, err)
	require.False(t, report.Ready)
	require.Contains(t, report.Missing, RequiredTestBroadcastDeduplication)

	failed := append([]NetworkingTestCoverageEvidence(nil), evidence...)
	for i := range failed {
		if failed[i].Test == RequiredTestCrossZoneDelivery {
			failed[i].Passed = false
			break
		}
	}
	report, err = EvaluateNetworkingTestCoverage(failed)
	require.NoError(t, err)
	require.False(t, report.Ready)
	require.Contains(t, report.Failed, RequiredTestCrossZoneDelivery)
}

func TestNetworkingObservabilitySpecCoversRequiredMetricsAndEvents(t *testing.T) {
	spec := DefaultNetworkingObservabilitySpec()
	require.NoError(t, ValidateNetworkingObservabilitySpec(spec))
	require.Len(t, spec.Metrics, 16)
	require.Len(t, spec.Events, 13)
	require.Len(t, spec.Alerts, 9)
	require.NotEmpty(t, spec.SpecRoot)
	require.Contains(t, spec.Metrics, ObservableMetricActivePeers)
	require.Contains(t, spec.Metrics, ObservableMetricRoutingFailureCount)
	require.Contains(t, spec.Events, ObservableEventNetworkNodeRegistered)
	require.Contains(t, spec.Events, ObservableEventNetworkRouteFailed)
	require.Contains(t, spec.Alerts, ObservableAlertConsensusChannelLatencyAboveThreshold)
	require.Contains(t, spec.Alerts, ObservableAlertEclipseRiskPeerDiversityLow)

	missingMetric := spec
	missingMetric.Metrics = append([]NetworkingObservableMetric(nil), spec.Metrics[:len(spec.Metrics)-1]...)
	missingMetric.SpecRoot = ComputeNetworkingObservabilitySpecRoot(missingMetric)
	require.ErrorContains(t, ValidateNetworkingObservabilitySpec(missingMetric), "must define 16 metrics")

	unknownEvent := spec
	unknownEvent.Events = append([]NetworkingObservableEvent(nil), spec.Events...)
	unknownEvent.Events[0] = NetworkingObservableEvent("network_unknown")
	unknownEvent.SpecRoot = ComputeNetworkingObservabilitySpecRoot(unknownEvent)
	require.ErrorContains(t, ValidateNetworkingObservabilitySpec(unknownEvent), "unknown networking observability event")

	missingAlert := spec
	missingAlert.Alerts = append([]NetworkingObservableAlert(nil), spec.Alerts[:len(spec.Alerts)-1]...)
	missingAlert.SpecRoot = ComputeNetworkingObservabilitySpecRoot(missingAlert)
	require.ErrorContains(t, ValidateNetworkingObservabilitySpec(missingAlert), "must define 9 alerts")

	unknownAlert := spec
	unknownAlert.Alerts = append([]NetworkingObservableAlert(nil), spec.Alerts...)
	unknownAlert.Alerts[0] = NetworkingObservableAlert("network_unknown_alert")
	unknownAlert.SpecRoot = ComputeNetworkingObservabilitySpecRoot(unknownAlert)
	require.ErrorContains(t, ValidateNetworkingObservabilitySpec(unknownAlert), "unknown networking observability alert")
}

func TestNetworkingObservabilityReportRequiresSamplesAndEvents(t *testing.T) {
	spec := DefaultNetworkingObservabilitySpec()
	metrics := testNetworkingObservabilityMetrics()
	events := testNetworkingObservabilityEvents()
	report, err := BuildNetworkingObservabilityReport(spec, metrics, events)
	require.NoError(t, err)
	require.True(t, report.Ready)
	require.Empty(t, report.MissingMetrics)
	require.Empty(t, report.MissingEvents)
	require.NotEmpty(t, report.ReportHash)

	missingMetric := make([]NetworkingMetricSample, 0, len(metrics)-1)
	for _, sample := range metrics {
		if sample.Metric != ObservableMetricBroadcastDedupHitRate {
			missingMetric = append(missingMetric, sample)
		}
	}
	report, err = BuildNetworkingObservabilityReport(spec, missingMetric, events)
	require.NoError(t, err)
	require.False(t, report.Ready)
	require.Contains(t, report.MissingMetrics, ObservableMetricBroadcastDedupHitRate)

	missingEvent := make([]NetworkingEventRecord, 0, len(events)-1)
	for _, event := range events {
		if event.Event != ObservableEventNetworkRouteFailed {
			missingEvent = append(missingEvent, event)
		}
	}
	report, err = BuildNetworkingObservabilityReport(spec, metrics, missingEvent)
	require.NoError(t, err)
	require.False(t, report.Ready)
	require.Contains(t, report.MissingEvents, ObservableEventNetworkRouteFailed)

	tampered := append([]NetworkingEventRecord(nil), events...)
	tampered[0].EventID = HashParts("wrong-event-id")
	_, err = BuildNetworkingObservabilityReport(spec, metrics, tampered)
	require.ErrorContains(t, err, "event id mismatch")
}

func TestNetworkingAlertRulesAndSignalsCoverSectionSeventeenThree(t *testing.T) {
	rules := DefaultNetworkingAlertRules()
	require.NoError(t, ValidateNetworkingAlertRules(rules))
	require.Len(t, rules, 9)
	require.Contains(t, networkingAlertRuleIDs(rules), ObservableAlertConsensusChannelLatencyAboveThreshold)
	require.Contains(t, networkingAlertRuleIDs(rules), ObservableAlertDiscoveryPoisoningAttempt)
	require.Contains(t, networkingAlertRuleIDs(rules), ObservableAlertEclipseRiskPeerDiversityLow)

	missing := append([]NetworkingAlertRule(nil), rules[:len(rules)-1]...)
	require.ErrorContains(t, ValidateNetworkingAlertRules(missing), "must define 9 alerts")

	invalid := append([]NetworkingAlertRule(nil), rules...)
	invalid[0].Threshold = 0
	require.ErrorContains(t, ValidateNetworkingAlertRules(invalid), "threshold must be positive")

	signals := testNetworkingAlertSignals(rules)
	report, err := BuildNetworkingAlertReport(rules, signals)
	require.NoError(t, err)
	require.True(t, report.Ready)
	require.Empty(t, report.MissingAlerts)
	require.NotEmpty(t, report.ReportHash)

	withoutDiscovery := make([]NetworkingAlertSignal, 0, len(signals)-1)
	for _, signal := range signals {
		if signal.Alert != ObservableAlertDiscoveryPoisoningAttempt {
			withoutDiscovery = append(withoutDiscovery, signal)
		}
	}
	report, err = BuildNetworkingAlertReport(rules, withoutDiscovery)
	require.NoError(t, err)
	require.False(t, report.Ready)
	require.Contains(t, report.MissingAlerts, ObservableAlertDiscoveryPoisoningAttempt)

	tampered := append([]NetworkingAlertSignal(nil), signals...)
	tampered[0].TriggerID = HashParts("wrong-alert-trigger")
	_, err = BuildNetworkingAlertReport(rules, tampered)
	require.ErrorContains(t, err, "trigger id mismatch")
}

func TestAdvisorySignalsCannotDriveConsensusUntilCommitted(t *testing.T) {
	metrics := PeerMetrics{LatencyMillis: 25, ReliabilityBps: 9_900, ThroughputBytesPerSec: 32 << 20}
	score, err := ComputePeerScore(metrics)
	require.NoError(t, err)

	require.NoError(t, ValidatePeerScoreUse(PeerScoreUse{Metrics: metrics, Score: score}))
	require.ErrorContains(t, ValidatePeerScoreUse(PeerScoreUse{
		Metrics:		metrics,
		Score:			score,
		UsedForConsensus:	true,
	}), "advisory")
	require.NoError(t, ValidatePeerScoreUse(PeerScoreUse{
		Metrics:		metrics,
		Score:			score,
		Committed:		true,
		UsedForConsensus:	true,
	}))
}

func TestRoutingAndStateTransitionHardRules(t *testing.T) {
	require.ErrorContains(t, ValidateRoutingDecisionUse(RoutingDecisionUse{UsedForConsensus: true}), "lowercase hex")
	require.NoError(t, ValidateRoutingDecisionUse(RoutingDecisionUse{
		UsedForConsensus:		true,
		DerivedFromCommittedState:	true,
	}))
	require.NoError(t, ValidateRoutingDecisionUse(RoutingDecisionUse{
		UsedForConsensus:	true,
		DeterministicProofHash:	HashParts("routing-proof"),
	}))

	require.NoError(t, ValidateStateTransitionNetworkAccess(StateTransitionNetworkAccess{
		InStateTransition: true,
	}))
	require.ErrorContains(t, ValidateStateTransitionNetworkAccess(StateTransitionNetworkAccess{
		InStateTransition:	true,
		ExternalCalls:		[]string{"https://example.invalid"},
	}), "forbidden")
}

func TestNetworkingNonGoalsDefineSectionEighteenBoundary(t *testing.T) {
	spec := DefaultNetworkingNonGoalSpec()
	require.NoError(t, ValidateNetworkingNonGoalSpec(spec))
	require.Len(t, spec.NonGoals, 7)
	require.NotEmpty(t, spec.SpecRoot)
	require.Contains(t, networkingNonGoalIDs(spec.NonGoals), NetworkingNonGoalApplicationLogic)
	require.Contains(t, networkingNonGoalIDs(spec.NonGoals), NetworkingNonGoalReplaceCometBFTConsensus)
	require.Contains(t, networkingNonGoalIDs(spec.NonGoals), NetworkingNonGoalExternalDiscoveryServices)
	require.Contains(t, networkingNonGoalIDs(spec.NonGoals), NetworkingNonGoalLiveMetricsConsensusAuthority)
	require.Contains(t, networkingNonGoalIDs(spec.NonGoals), NetworkingNonGoalOffChainServiceConsensusLogic)

	missing := spec
	missing.NonGoals = append([]NetworkingNonGoal(nil), spec.NonGoals[:len(spec.NonGoals)-1]...)
	missing.SpecRoot = ComputeNetworkingNonGoalSpecRoot(missing)
	require.ErrorContains(t, ValidateNetworkingNonGoalSpec(missing), "must define 7 non-goals")

	unknown := spec
	unknown.NonGoals = append([]NetworkingNonGoal(nil), spec.NonGoals...)
	unknown.NonGoals[0] = NetworkingNonGoal("ship_social_network")
	unknown.SpecRoot = ComputeNetworkingNonGoalSpecRoot(unknown)
	require.ErrorContains(t, ValidateNetworkingNonGoalSpec(unknown), "unknown networking non-goal")
}

func TestNetworkingScopeBoundaryRejectsNonGoalViolations(t *testing.T) {
	boundary := DefaultNetworkingScopeBoundary()
	require.NoError(t, ValidateNetworkingScopeBoundary(boundary))
	require.NotEmpty(t, boundary.BoundaryRoot)

	cases := []struct {
		name	string
		mutate	func(NetworkingScopeBoundary) NetworkingScopeBoundary
		wantErr	string
	}{
		{
			name:	"application logic",
			mutate: func(boundary NetworkingScopeBoundary) NetworkingScopeBoundary {
				boundary.ImplementsApplicationLogic = true
				boundary.BoundaryRoot = ComputeNetworkingScopeBoundaryRoot(boundary)
				return boundary
			},
			wantErr:	"application logic",
		},
		{
			name:	"replace cometbft",
			mutate: func(boundary NetworkingScopeBoundary) NetworkingScopeBoundary {
				boundary.ReplacesCometBFTConsensus = true
				boundary.BoundaryRoot = ComputeNetworkingScopeBoundaryRoot(boundary)
				return boundary
			},
			wantErr:	"replace CometBFT consensus",
		},
		{
			name:	"centralized routing",
			mutate: func(boundary NetworkingScopeBoundary) NetworkingScopeBoundary {
				boundary.RequiresCentralizedRouting = true
				boundary.BoundaryRoot = ComputeNetworkingScopeBoundaryRoot(boundary)
				return boundary
			},
			wantErr:	"centralized routing",
		},
		{
			name:	"external discovery",
			mutate: func(boundary NetworkingScopeBoundary) NetworkingScopeBoundary {
				boundary.RequiresExternalDiscoveryServices = true
				boundary.BoundaryRoot = ComputeNetworkingScopeBoundaryRoot(boundary)
				return boundary
			},
			wantErr:	"external discovery services",
		},
		{
			name:	"messaging social",
			mutate: func(boundary NetworkingScopeBoundary) NetworkingScopeBoundary {
				boundary.IntroducesMessagingSocialLayer = true
				boundary.BoundaryRoot = ComputeNetworkingScopeBoundaryRoot(boundary)
				return boundary
			},
			wantErr:	"messaging or social network",
		},
		{
			name:	"live metrics authoritative",
			mutate: func(boundary NetworkingScopeBoundary) NetworkingScopeBoundary {
				boundary.LiveMetricsConsensusAuthoritative = true
				boundary.BoundaryRoot = ComputeNetworkingScopeBoundaryRoot(boundary)
				return boundary
			},
			wantErr:	"live network metrics consensus-authoritative",
		},
		{
			name:	"offchain service logic",
			mutate: func(boundary NetworkingScopeBoundary) NetworkingScopeBoundary {
				boundary.OffChainServiceLogicInConsensus = true
				boundary.BoundaryRoot = ComputeNetworkingScopeBoundaryRoot(boundary)
				return boundary
			},
			wantErr:	"off-chain service logic inside consensus",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.ErrorContains(t, ValidateNetworkingScopeBoundary(tc.mutate(boundary)), tc.wantErr)
		})
	}

	tampered := boundary
	tampered.BoundaryRoot = HashParts("wrong-non-goal-boundary")
	require.ErrorContains(t, ValidateNetworkingScopeBoundary(tampered), "boundary root mismatch")
}

func TestNetworkRoleConsensusScopeRequiresBondedCommitment(t *testing.T) {
	salt := []byte("aetra-test-network")
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
		NodeID:		record.NodeID,
		Role:		NodeRoleService,
		Bonded:		true,
		CommitmentHash:	HashParts("service-role-commitment"),
		ExpiresHeight:	80,
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
	salt := []byte("aetra-test-network")
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
	salt := []byte("aetra-test-network")
	privateKey := deterministicPrivateKey(0x71)
	validatorKey := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{0x72}, ed25519.SeedSize)).Public().(ed25519.PublicKey)
	addressHash, err := HashNetworkAddresses([]string{"tcp://127.0.0.1:26656"})
	require.NoError(t, err)
	record, err := SignNodeRecord(NodeRecord{
		ValidatorPubKey:	validatorKey,
		Roles:			[]NodeRole{NodeRoleValidator, NodeRoleFull},
		NetworkAddressesHash:	addressHash,
		ProtocolVersions:	[]string{DefaultProtocolVersion},
		ExpiresHeight:		100,
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
		NodeID:		record.NodeID,
		Role:		NodeRoleValidator,
		Bonded:		true,
		CommitmentHash:	HashParts("validator-role-commitment"),
		ExpiresHeight:	80,
	}, 10)
	require.ErrorContains(t, err, "validator role")
}

func TestRoleCommitmentRejectsUnbondedUnadvertisedAndOutlivingRecords(t *testing.T) {
	salt := []byte("aetra-test-network")
	record := signedNodeRecord(t, 0x81, salt, 100, NodeRoleService)
	state := EmptyState()
	var err error
	state, err = RegisterNodeRecord(state, record, salt, 10)
	require.NoError(t, err)

	_, err = RegisterRoleCommitment(state, RoleCommitment{
		NodeID:		record.NodeID,
		Role:		NodeRoleService,
		CommitmentHash:	HashParts("unbonded"),
		ExpiresHeight:	80,
	}, 10)
	require.ErrorContains(t, err, "bonded")

	_, err = RegisterRoleCommitment(state, RoleCommitment{
		NodeID:		record.NodeID,
		Role:		NodeRoleRouting,
		Bonded:		true,
		CommitmentHash:	HashParts("not-advertised"),
		ExpiresHeight:	80,
	}, 10)
	require.ErrorContains(t, err, "advertised")

	_, err = RegisterRoleCommitment(state, RoleCommitment{
		NodeID:		record.NodeID,
		Role:		NodeRoleService,
		Bonded:		true,
		CommitmentHash:	HashParts("outlive"),
		ExpiresHeight:	101,
	}, 10)
	require.ErrorContains(t, err, "outlive")
}

func testEnvelope(channel ChannelClass, sizeBytes uint64, enqueuedHeight uint64, sequence uint64, label string) TransportEnvelope {
	return TransportEnvelope{
		Channel:	channel,
		SizeBytes:	sizeBytes,
		EnqueuedHeight:	enqueuedHeight,
		Sequence:	sequence,
		PayloadHash:	HashParts(label),
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
		OverlayID:	desc.OverlayID,
		NodeID:		record.NodeID,
		ProofType:	proofType,
		Mode:		mode,
		Membership:	desc.Membership,
		ProofHash:	HashParts("overlay-membership-proof", record.NodeID, desc.OverlayID, string(proofType)),
		AuthorityHash:	HashParts("overlay-membership-authority", desc.OverlayID),
		ExpiresHeight:	expiresHeight,
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
		LatencyMillis:	latencyMillis,
		ReliabilityBps:	reliabilityBps,
	}
	return AdaptivePeerFromNodeRecord(record, score, metrics, committed, 20)
}

func testMeshMessage(t *testing.T, messageType AetherMeshMessageType, overlayID, origin, destination string, priority uint32, sequence uint64) AetherMeshMessage {
	t.Helper()

	msg, err := NewAetherMeshMessage(AetherMeshMessage{
		Type:		messageType,
		Payload:	[]byte(fmt.Sprintf("mesh-%s-%d", messageType, sequence)),
		Origin:		origin,
		Destination:	destination,
		Priority:	priority,
		TTL:		50,
		OverlayID:	overlayID,
		Sequence:	sequence,
	})
	require.NoError(t, err)
	return msg
}

func testGlobalAdaptivePeer(t *testing.T, record NodeRecord, scoreBps uint32, latencyMillis uint64, reliabilityBps uint32) AdaptivePeer {
	t.Helper()

	peer := testAdaptivePeer(t, record, scoreBps, latencyMillis, reliabilityBps, false)
	peer.ZonesSupported = nil
	peer.Services = nil
	return peer
}

func testSecurityAdaptivePeer(label string, roles []NodeRole, zones []string) AdaptivePeer {
	return AdaptivePeer{
		NodeID:			HashParts("security-adaptive-peer", label),
		ScoreBps:		8_000,
		LatencyMillis:		25,
		ReliabilityBps:		8_500,
		Roles:			roles,
		ZonesSupported:		zones,
		LastSeenHeight:		7,
		LastScoreHeight:	7,
	}
}

func testPerformanceBlockSession(t *testing.T) BlockPropagationSession {
	t.Helper()

	chunks, err := ChunkPayload(bytes.Repeat([]byte("performance-block"), 16), 64)
	require.NoError(t, err)
	chunkRoot, err := ComputeRL2ChunkRoot(chunks)
	require.NoError(t, err)
	proofRoot := HashParts("performance-proof-root")
	headerHash := HashParts("performance-header")
	header, err := NewBlockBroadcastHeader(BlockBroadcastHeader{
		Height:				12,
		ProposerNodeID:			HashParts("performance-proposer"),
		HeaderHash:			headerHash,
		ChunkSetRoot:			chunkRoot,
		ProofSetRoot:			proofRoot,
		BlockRoot:			ComputeBlockRoot(headerHash, chunkRoot, proofRoot),
		ChunkCount:			uint32(len(chunks)),
		AvailabilityMetadataHash:	HashParts("performance-availability"),
	})
	require.NoError(t, err)
	return BlockPropagationSession{
		Header:		header,
		ProofSet:	BlockProofSet{BlockID: header.BlockID, ProofRoot: proofRoot, ProofHashes: []string{HashParts("performance-proof")}},
		VerifiedBitmap:	make([]bool, header.ChunkCount),
	}
}

func testPerformanceStreamPlan() StreamParallelFetchPlan {
	streamID := HashParts("performance-stream")
	return StreamParallelFetchPlan{
		StreamID:	streamID,
		ChunkSize:	256,
		PayloadBytes:	1024,
		TotalChunks:	4,
		Requests: []StreamFetchRequest{
			{StreamID: streamID, ChunkIndex: 0, RangeStart: 0, RangeEnd: 256, ChunkSize: 256, AssignedPeer: "peer-a"},
			{StreamID: streamID, ChunkIndex: 1, RangeStart: 256, RangeEnd: 512, ChunkSize: 256, AssignedPeer: "peer-b"},
		},
	}
}

func testRoadmapEvidence(t *testing.T) NetworkingRoadmapEvidence {
	t.Helper()

	salt := []byte("aetra-roadmap")
	local := signedNodeRecord(t, 0x88, salt, 100, NodeRoleFull, NodeRoleRouting)
	remote := signedNodeRecordWithCapabilities(t, 0x89, salt, 100, []NodeRole{NodeRoleService, NodeRoleRouting}, []string{"zone-roadmap"}, []string{"svc.roadmap"})
	adapter := DefaultAetherNetworkingAdapter()
	envelopes := []TransportEnvelope{
		testEnvelope(ChannelConsensus, 512, 10, 1, "roadmap-consensus"),
		testEnvelope(ChannelMempool, 512, 10, 2, "roadmap-mempool"),
		testEnvelope(ChannelService, 512, 10, 3, "roadmap-service"),
	}
	schedule, err := ScheduleL0Propagation(adapter, envelopes, 4, PeerScore{ScoreBps: 8_000}, 10)
	require.NoError(t, err)
	channelBandwidth, err := ComputeChannelBandwidthMetrics(schedule.Metrics)
	require.NoError(t, err)

	session, err := NegotiateSession(local, remote, testSessionRequest(local, remote, 10, 80, "roadmap-session", []ChannelClass{ChannelConsensus, ChannelService, ChannelData}))
	require.NoError(t, err)
	discovery := testSignedDiscoveryObjectRecord(t, remote, 0x89, salt, DRTObjectServiceEndpoint, HashParts("roadmap-target"), HashParts("roadmap-ad"), "", "svc.roadmap", "", 90)
	descriptors := DefaultOverlayDescriptors()
	memberships := make([]OverlayMembershipRecord, 0, 5)
	for _, overlayType := range []OverlayType{OverlayTypeValidator, OverlayTypeZone, OverlayTypeService, OverlayTypeData, OverlayTypeDiscovery} {
		desc := testDefaultOverlayDescriptor(t, overlayType)
		memberships = append(memberships, OverlayMembershipRecord{
			OverlayID:	desc.OverlayID,
			NodeID:		remote.NodeID,
			ProofID:	HashParts("roadmap-membership-proof", string(overlayType)),
			Membership:	desc.Membership,
			Mode:		OverlayMembershipModeCryptographicAuth,
			JoinedHeight:	10,
			ExpiresHeight:	90,
		})
	}
	serviceDesc := testDefaultOverlayDescriptor(t, OverlayTypeService)
	peerA := testAdaptivePeer(t, local, 8_000, 20, 9_000, true)
	peerB := testAdaptivePeer(t, remote, 8_500, 15, 9_250, true)
	peerC := AdaptivePeer{
		NodeID:			HashParts("roadmap-peer-c"),
		ScoreBps:		7_500,
		LatencyMillis:		30,
		ReliabilityBps:		8_500,
		Roles:			[]NodeRole{NodeRoleFull},
		LastSeenHeight:		10,
		LastScoreHeight:	10,
	}
	adaptiveGraph := NormalizeAdaptiveOverlayGraph(AdaptiveOverlayGraph{
		OverlayID:	serviceDesc.OverlayID,
		LocalNodeID:	local.NodeID,
		RoutingEpoch:	10,
		RandomSet:	[]AdaptivePeer{peerA, peerC},
		ServiceSet:	[]AdaptivePeer{peerB},
		FallbackSet:	[]AdaptivePeer{peerA},
		PolicyHash:	HashParts("roadmap-adaptive-policy"),
	})
	routingGraph := NormalizeRoutingGraph(RoutingGraph{
		OverlayID:		serviceDesc.OverlayID,
		Version:		10,
		Committed:		true,
		DeterministicHintHash:	HashParts("roadmap-routing-hint"),
		Edges: []RoutingEdge{{
			FromNodeID:	local.NodeID,
			ToNodeID:	remote.NodeID,
			LatencyMillis:	20,
			Weight:		1,
			Priority:	PriorityForChannel(ChannelService),
			ZoneID:		"zone-roadmap",
		}},
	})
	commitment, err := NewRoutingTableCommitment(RoutingTableCommitment{
		RoutingEpoch:		10,
		OverlayRoots:		[]OverlayRouteRoot{{OverlayID: serviceDesc.OverlayID, RootHash: routingGraph.GraphHash}},
		ZoneRouteRoot:		HashParts("roadmap-zone-route-root"),
		ServiceRouteRoot:	HashParts("roadmap-service-route-root"),
		PeerClassRoot:		HashParts("roadmap-peer-class-root"),
		CongestionSnapshotRoot:	HashParts("roadmap-congestion-root"),
		PolicyHash:		HashParts("roadmap-routing-policy"),
	})
	require.NoError(t, err)
	chunks, err := ChunkPayload(bytes.Repeat([]byte("roadmap-rl2"), 128), 64)
	require.NoError(t, err)
	offer, rl2Descriptors, err := NewRL2TransferOfferFromChunks(local.NodeID, remote.NodeID, RL2PayloadLargeBlock, chunks, PriorityForChannel(ChannelBlock), 10, 80, RL2FECNone, PeerScore{ScoreBps: 9_000}, 4096, 2)
	require.NoError(t, err)
	rl2Session, err := AcceptRL2TransferOffer(offer, []uint32{0}, 11)
	require.NoError(t, err)
	rl2Session, err = StartRL2Transfer(rl2Session, PeerScore{ScoreBps: 9_000}, 4096, 0)
	require.NoError(t, err)
	signal, err := NewRL2BackpressureSignal(offer.Transfer, 2048, 512, []uint32{0})
	require.NoError(t, err)
	paused, err := PauseRL2Transfer(rl2Session, signal, 12)
	require.NoError(t, err)
	resumed, err := ResumeRL2Transfer(paused, []uint32{0}, 13)
	require.NoError(t, err)
	badChunk := chunks[0]
	badChunk.ChunkHash = HashParts("roadmap-invalid-chunk")
	invalidChunkErr := VerifyRL2Chunk(offer.Transfer, rl2Descriptors[0], badChunk)
	zoneOwner := signedNodeRecordWithCapabilities(t, 0x8a, salt, 100, []NodeRole{NodeRoleZoneExecution}, []string{"zone-roadmap"}, nil)
	rpcOwner := signedNodeRecord(t, 0x8b, salt, 100, NodeRoleFull)
	storageOwner := signedNodeRecord(t, 0x8c, salt, 100, NodeRoleStorageProvider)
	nodeDiscovery := testSignedDiscoveryObjectRecord(t, local, 0x88, salt, DRTObjectNode, local.NodeID, HashParts("roadmap-node-ad"), "", "", "", 90)
	zoneDiscovery := testSignedDiscoveryObjectRecord(t, zoneOwner, 0x8a, salt, DRTObjectExecutionZone, HashParts("roadmap-zone-target"), HashParts("roadmap-zone-ad"), "zone-roadmap", "", "", 90)
	rpcDiscovery := testSignedDiscoveryObjectRecord(t, rpcOwner, 0x8b, salt, DRTObjectRPCEndpoint, rpcOwner.NodeID, HashParts("roadmap-rpc-endpoint"), "", "", "", 90)
	storageDiscovery := testSignedDiscoveryObjectRecord(t, storageOwner, 0x8c, salt, DRTObjectStorageProvider, storageOwner.NodeID, HashParts("roadmap-storage-endpoint"), "", "", "", 90)
	discoveryTable := EmptyDistributedRoutingTable()
	for _, record := range []DiscoveryRecord{nodeDiscovery, discovery, zoneDiscovery, rpcDiscovery, storageDiscovery} {
		discoveryTable, err = discoveryTable.Store(record, salt, 20)
		require.NoError(t, err)
	}
	renewedDiscovery, err := RenewDiscoveryRecord(discovery, 95, deterministicPrivateKey(0x89), salt)
	require.NoError(t, err)
	discoveryTable, err = discoveryTable.UpdateLease(renewedDiscovery, salt, 30)
	require.NoError(t, err)
	resultHash, err := ComputeDiscoveryResponseResultHash(discoveryTable.FindService("svc.roadmap", 30))
	require.NoError(t, err)
	stateRoot := HashParts("roadmap-discovery-state-root")
	discoveryProof := DiscoveryOnChainProof{
		ProofHeight:	30,
		StateRoot:	stateRoot,
		ProofHash:	ComputeDiscoveryOnChainProofHash(resultHash, stateRoot, 30),
	}
	discoveryResponse, err := BuildDiscoveryResponse(discoveryTable, DRTQuery{ObjectType: DRTObjectServiceEndpoint, ServiceID: "svc.roadmap", CurrentHeight: 30}, local, deterministicPrivateKey(0x88), salt, discoveryProof, 30)
	require.NoError(t, err)
	require.NoError(t, discoveryResponse.Validate(local.NodePubKey, salt, 30))
	forgedDiscovery := discovery
	forgedDiscovery.Signature = cloneBytes(forgedDiscovery.Signature)
	forgedDiscovery.Signature[0] ^= 0xff
	forgedDiscoveryErr := ValidateSignedDiscoveryRecord(forgedDiscovery, salt, 20)
	expiredDiscoveryErr := ValidateSignedDiscoveryRecord(discovery, salt, 91)
	broadcastOrigin := signedNodeRecord(t, 0x8d, salt, 100, NodeRoleRouting)
	broadcastMsg, err := SignBroadcastMessage(BroadcastMessage{
		OverlayID:	serviceDesc.OverlayID,
		PayloadHash:	HashParts("roadmap-broadcast-payload"),
		PayloadType:	BroadcastPayloadService,
		Height:		20,
		TTL:		64,
		Priority:	PriorityForChannel(ChannelService),
		FanoutPolicy: BroadcastFanoutPolicy{
			TreeFanout:	serviceDesc.Fanout + 10,
			GossipFanout:	serviceDesc.Fanout + 10,
			OverlayBound:	true,
		},
	}, deterministicPrivateKey(0x8d), salt)
	require.NoError(t, err)
	broadcastGraph := RoutingGraph{
		OverlayID:	serviceDesc.OverlayID,
		Version:	11,
		Edges: []RoutingEdge{
			{FromNodeID: local.NodeID, ToNodeID: remote.NodeID, LatencyMillis: 10, Weight: 9_000, Priority: 1},
			{FromNodeID: local.NodeID, ToNodeID: peerC.NodeID, LatencyMillis: 20, Weight: 8_000, Priority: 2},
		},
	}
	broadcastGraph.GraphHash = ComputeRoutingGraphHash(broadcastGraph)
	_, broadcastPlan, err := PlanBroadcastForwarding(broadcastMsg, serviceDesc, broadcastGraph, local.NodeID, []string{local.NodeID, broadcastOrigin.NodeID, remote.NodeID, peerC.NodeID, HashParts("roadmap-broadcast-peer-d")}, BroadcastDeduper{}, 20)
	require.NoError(t, err)
	broadcastCache := NewBroadcastDedupCache(64)
	broadcastCache, acceptDecision, err := broadcastCache.Accept(broadcastMsg, remote.NodeID, 20)
	require.NoError(t, err)
	require.True(t, acceptDecision.Accepted)
	broadcastCache, duplicateDecision, err := broadcastCache.Accept(broadcastMsg, remote.NodeID, 21)
	require.NoError(t, err)
	conflictMsg := broadcastMsg
	conflictMsg.PayloadHash = HashParts("roadmap-conflicting-broadcast-payload")
	broadcastCache, conflictDecision, err := broadcastCache.Accept(conflictMsg, remote.NodeID, 21)
	require.NoError(t, err)
	executionDesc := testDefaultOverlayDescriptor(t, OverlayTypeExecution)
	dataDesc := testDefaultOverlayDescriptor(t, OverlayTypeData)
	executionMesh, err := NewAetherMeshMessage(AetherMeshMessage{
		Type:			MeshMessageExecution,
		Payload:		[]byte("roadmap-execution-message"),
		Origin:			local.NodeID,
		Destination:		remote.NodeID,
		Priority:		PriorityForChannel(ChannelExecution),
		TTL:			40,
		OverlayID:		executionDesc.OverlayID,
		DestinationZone:	"zone-roadmap",
		Sequence:		1,
		ConsensusEffect:	true,
		DeterminismSource:	DeterminismCommittedState,
		Proof: AetherMeshProof{
			ProofType:	"roadmap-execution-schedule",
			ProofHash:	HashParts("roadmap-execution-proof"),
			ProofHeight:	30,
		},
	})
	require.NoError(t, err)
	serviceMesh, err := NewAetherMeshMessage(AetherMeshMessage{
		Type:		MeshMessageService,
		Payload:	[]byte("roadmap-service-message"),
		Origin:		local.NodeID,
		Destination:	remote.NodeID,
		Priority:	PriorityForChannel(ChannelService),
		TTL:		40,
		OverlayID:	serviceDesc.OverlayID,
		Sequence:	2,
		RouteHint:	RouteHint{ServiceID: "svc.roadmap"},
		DeadlineHeight:	90,
	})
	require.NoError(t, err)
	queryMesh, err := NewAetherMeshMessage(AetherMeshMessage{
		Type:		MeshMessageQuery,
		Payload:	[]byte("roadmap-query-message"),
		Origin:		local.NodeID,
		Destination:	remote.NodeID,
		Priority:	PriorityForChannel(ChannelService),
		TTL:		40,
		OverlayID:	serviceDesc.OverlayID,
		Sequence:	3,
	})
	require.NoError(t, err)
	storageMesh, err := NewAetherMeshMessage(AetherMeshMessage{
		Type:		MeshMessageStorage,
		Payload:	[]byte("roadmap-storage-message"),
		Origin:		local.NodeID,
		Destination:	storageOwner.NodeID,
		Priority:	PriorityForChannel(ChannelData),
		TTL:		40,
		OverlayID:	dataDesc.OverlayID,
		Sequence:	4,
		RouteHint:	RouteHint{StorageKeyHash: HashParts("roadmap-storage-key")},
	})
	require.NoError(t, err)
	crossZoneMesh, err := NewAetherMeshMessage(AetherMeshMessage{
		Type:			MeshMessageCrossZone,
		Payload:		[]byte("roadmap-cross-zone-message"),
		Origin:			local.NodeID,
		Destination:		remote.NodeID,
		Priority:		PriorityForChannel(ChannelExecution),
		TTL:			40,
		OverlayID:		executionDesc.OverlayID,
		SourceZone:		"zone-roadmap",
		DestinationZone:	"zone-remote",
		Sequence:		5,
		ConsensusEffect:	true,
		DeterminismSource:	DeterminismDeterministicProof,
		Proof: AetherMeshProof{
			ProofType:	"roadmap-cross-zone-proof",
			ProofHash:	HashParts("roadmap-cross-zone-proof"),
			ProofHeight:	30,
		},
	})
	require.NoError(t, err)
	meshRoute := func(msg AetherMeshMessage, desc OverlayDescriptor, channel ChannelClass, target string) AetherMeshDelivery {
		return AetherMeshDelivery{
			Message:	msg,
			Channel:	channel,
			Route: OverlayRoutePlan{
				MessageID:	msg.MessageID,
				OverlayID:	desc.OverlayID,
				OverlayType:	desc.OverlayType,
				Strategy:	desc.Routing,
				TargetNodeIDs:	[]string{target},
			},
		}
	}
	crossZoneMsg, err := NewCrossZoneMessage(CrossZoneMessage{
		SourceZone:		"zone-roadmap",
		DestinationZone:	"zone-remote",
		SourceSequence:		1,
		MessageHash:		crossZoneMesh.PayloadHash,
		ExpiryHeight:		100,
		ReceiptPolicy:		ReceiptPolicyAlways,
		ProofRequired:		true,
	})
	require.NoError(t, err)
	crossZoneTracker := CrossZoneSequenceTracker{}
	crossZoneTracker, err = AcceptCrossZoneSequence(crossZoneTracker, crossZoneMsg, true, 30)
	require.NoError(t, err)
	_, replayErr := AcceptCrossZoneSequence(crossZoneTracker, crossZoneMsg, true, 31)
	crossZoneReceipt, err := NewCrossZoneReceipt(CrossZoneReceipt{
		SourceZone:		crossZoneMsg.SourceZone,
		DestinationZone:	crossZoneMsg.DestinationZone,
		SourceSequence:		crossZoneMsg.SourceSequence,
		MessageHash:		crossZoneMsg.MessageHash,
		Status:			CrossZoneReceiptExecuted,
		ReceiptPolicy:		ReceiptPolicyAlways,
		ProofHash:		HashParts("roadmap-cross-zone-receipt-proof"),
		ReceiptHeight:		35,
		RollbackSafe:		true,
		ProofQueryable:		true,
	})
	require.NoError(t, err)
	receiptDelivery, err := NewReceiptDelivery(crossZoneReceipt, remote.NodeID, 36)
	require.NoError(t, err)
	receiptDelivery, err = AckReceiptDelivery(receiptDelivery, HashParts("roadmap-receipt-ack"))
	require.NoError(t, err)
	queryResponseProof, err := NewQueryResponseProof(QueryResponseProof{
		RequestID:	queryMesh.MessageID,
		Responder:	remote.NodeID,
		PayloadHash:	HashParts("roadmap-query-response"),
		Proof: AetherMeshProof{
			ProofType:	"roadmap-query-proof",
			ProofHash:	HashParts("roadmap-query-proof"),
			ProofHeight:	36,
		},
		ResponseHeight:	37,
	})
	require.NoError(t, err)
	l3Metrics, err := EvaluateL3Metrics(nil, []ReceiptDelivery{receiptDelivery}, []QueryResponseProof{queryResponseProof}, nil, nil)
	require.NoError(t, err)
	securityPolicy := DefaultNetworkSecurityPolicy()
	securityDecision, err := EvaluateNetworkSecurity(PeerScore{ScoreBps: 8_500, ReliabilityBps: 9_000, ThroughputBps: 8_000}, PeerSecurityObservation{
		PeerNodeID:		remote.NodeID,
		InvalidMessages:	securityPolicy.MaxInvalidMessages + 1,
		DuplicateMessages:	securityPolicy.MaxInvalidMessages + 2,
		ConflictingBroadcasts:	1,
		CorruptChunks:		1,
		ForgedAdvertisements:	1,
		BytesThisEpoch:		securityPolicy.MaxBytesPerEpoch + 1,
		SybilClusterPeers:	1,
		CrossZoneReplayCount:	1,
	}, securityPolicy)
	require.NoError(t, err)
	reputationDecision, err := ComputePeerReputation(PeerReputationInput{
		PeerNodeID:			remote.NodeID,
		ValidMessages:			90,
		InvalidMessages:		10,
		LatencyMillis:			50,
		ThroughputBytesPerSec:		32 << 20,
		CorrectChunks:			10,
		CorruptChunks:			1,
		ValidDiscoveryResponses:	9,
		InvalidDiscoveryResponses:	1,
		ValidServiceResponses:		9,
		InvalidServiceResponses:	1,
		Timeouts:			1,
		DuplicateBroadcasts:		1,
		ConflictingBroadcasts:		1,
	})
	require.NoError(t, err)
	eclipsePlan, err := BuildEclipseResistancePlan(AdaptiveOverlayGraph{
		OverlayID:	serviceDesc.OverlayID,
		LocalNodeID:	local.NodeID,
		RoutingEpoch:	30,
		PolicyHash:	HashParts("roadmap-security-eclipse-policy"),
		RandomSet: []AdaptivePeer{
			testSecurityAdaptivePeer("roadmap-random-a", []NodeRole{NodeRoleFull}, []string{"zone-a"}),
			testSecurityAdaptivePeer("roadmap-random-b", []NodeRole{NodeRoleFull}, []string{"zone-b"}),
		},
		FallbackSet: []AdaptivePeer{
			testSecurityAdaptivePeer("roadmap-fallback", []NodeRole{NodeRoleFull}, []string{"zone-c"}),
		},
		StableSet: []AdaptivePeer{
			testSecurityAdaptivePeer("roadmap-validator-a", []NodeRole{NodeRoleValidator}, []string{"zone-a"}),
			testSecurityAdaptivePeer("roadmap-validator-b", []NodeRole{NodeRoleValidator}, []string{"zone-b"}),
		},
		ZoneSet: []AdaptivePeer{
			testSecurityAdaptivePeer("roadmap-zone-a", []NodeRole{NodeRoleZoneExecution}, []string{"zone-a"}),
			testSecurityAdaptivePeer("roadmap-zone-b", []NodeRole{NodeRoleZoneExecution}, []string{"zone-b"}),
		},
	}, []DiscoveryRecord{{RecordType: DRTObjectRoutingEntryPoint, OwnerNodeID: HashParts("security-adaptive-peer", "roadmap-validator-a"), ProofHash: HashParts("roadmap-proof-backed-route"), ProofHeight: 30}}, DefaultEclipseResistancePolicy(), 30)
	require.NoError(t, err)
	_, eclipseThreats, err := SimulateEclipseResistance(AdaptiveOverlayGraph{
		OverlayID:	HashParts("roadmap-bad-eclipse-overlay"),
		LocalNodeID:	HashParts("roadmap-bad-eclipse-local"),
		RoutingEpoch:	31,
		PolicyHash:	HashParts("roadmap-bad-eclipse-policy"),
		RandomSet:	[]AdaptivePeer{testSecurityAdaptivePeer("roadmap-only-random", []NodeRole{NodeRoleFull}, []string{"zone-a"})},
		FallbackSet:	[]AdaptivePeer{testSecurityAdaptivePeer("roadmap-only-fallback", []NodeRole{NodeRoleFull}, []string{"zone-a"})},
	}, nil, DefaultEclipseResistancePolicy(), 31)
	require.NoError(t, err)
	spamSimulation, err := SimulateSpamResistance(securityPolicy, PeerRateUsage{
		PeerNodeID:	remote.NodeID,
		Channel:	ChannelService,
		Messages:	securityPolicy.MaxPeerMessagesPerWindow + 10,
		Bytes:		10 << 20,
		WindowStart:	30,
		WindowEnd:	30,
	}, []BroadcastMessage{broadcastMsg, broadcastMsg}, remote.NodeID, 30)
	require.NoError(t, err)
	routingManipulation, err := SimulateRoutingManipulation([]BroadcastMessage{broadcastMsg, conflictMsg}, remote.NodeID, 30)
	require.NoError(t, err)
	return NetworkingRoadmapEvidence{
		CometBFTInventory:	adapter,
		PerformanceSnapshot: PerformanceMetricsSnapshot{
			ChannelBandwidth:	channelBandwidth,
			BlockBenchmark: BlockPropagationBenchmark{
				HeaderLatencyMillis:	10,
				ReconstructionMillis:	20,
				ChunkCount:		2,
				HeaderFirst:		true,
			},
			PeerScoreDistribution: PeerScoreDistributionMetric{
				Count:		2,
				AverageBps:	8_000,
			},
			MessagePropagationLatency: PerformanceLatencySummary{
				Count:		2,
				MinMillis:	5,
				MaxMillis:	15,
				AverageMillis:	10,
			},
			ServiceTrafficIsolated:	true,
		},
		L0Schedule:			schedule,
		XNetworkParams:			DefaultXNetworkParams(salt),
		Session:			session,
		NodeRecords:			[]NodeRecord{local, remote},
		SignedDiscoveryRecords:		[]DiscoveryRecord{discovery, nodeDiscovery, zoneDiscovery, rpcDiscovery, storageDiscovery},
		HandshakeReplayRejected:	true,
		KeyRotationAvailable:		true,
		OverlayDescriptors:		descriptors,
		OverlayMemberships:		memberships,
		AdaptiveGraph:			adaptiveGraph,
		RoutingGraph:			routingGraph,
		RoutingTableUse: RoutingTableUse{
			Commitment:			commitment,
			Committed:			true,
			UsedForExecutionScheduling:	true,
		},
		PeerRotationPreserved:		true,
		RL2Offer:			offer,
		RL2ChunkDescriptors:		rl2Descriptors,
		RL2Session:			resumed,
		RL2StreamingPlan:		resumed.StreamingPlan,
		RL2PayloadTypes:		[]RL2PayloadType{RL2PayloadLargeBlock, RL2PayloadStateSyncStream, RL2PayloadProofSet},
		RL2BackpressureSignal:		signal,
		RL2InvalidChunkRejected:	invalidChunkErr != nil,
		RL2InterruptedResumed:		true,
		DiscoveryTable:			discoveryTable,
		DiscoveryResponse:		discoveryResponse,
		DiscoveryObjectTypes:		[]DRTObjectType{DRTObjectNode, DRTObjectExecutionZone, DRTObjectServiceEndpoint, DRTObjectRPCEndpoint, DRTObjectStorageProvider},
		DiscoveryLeaseRenewed:		renewedDiscovery.ExpiresHeight == 95,
		DiscoveryForgedRejected:	forgedDiscoveryErr != nil,
		DiscoveryExpiredRejected:	expiredDiscoveryErr != nil,
		BroadcastMessage:		broadcastMsg,
		BroadcastPlan:			broadcastPlan,
		BroadcastDedupCache:		broadcastCache,
		BroadcastDuplicateHandled:	duplicateDecision.DroppedDuplicate,
		BroadcastConflictHandled:	conflictDecision.FaultEvidence.EvidenceHash != "",
		BlockSession:			testPerformanceBlockSession(t),
		ParallelChunkPlan:		testPerformanceStreamPlan(),
		GossipFallbackUsed:		broadcastPlan.FallbackUsed,
		MeshMessages:			[]AetherMeshMessage{executionMesh, serviceMesh, queryMesh, storageMesh, crossZoneMesh},
		MeshDeliveries: []AetherMeshDelivery{
			meshRoute(executionMesh, executionDesc, ChannelExecution, remote.NodeID),
			meshRoute(serviceMesh, serviceDesc, ChannelService, remote.NodeID),
			meshRoute(crossZoneMesh, executionDesc, ChannelExecution, remote.NodeID),
		},
		CrossZoneTracker:		crossZoneTracker,
		CrossZoneReceipt:		crossZoneReceipt,
		ReceiptDelivery:		receiptDelivery,
		QueryResponseProof:		queryResponseProof,
		L3Metrics:			l3Metrics,
		CrossZoneAtLeastOnce:		true,
		CrossZoneExactlyOnce:		replayErr != nil,
		SecurityPolicy:			securityPolicy,
		SecurityDecision:		securityDecision,
		ReputationDecision:		reputationDecision,
		EclipsePlan:			eclipsePlan,
		EclipseThreats:			eclipseThreats,
		SpamSimulation:			spamSimulation,
		RoutingManipulation:		routingManipulation,
		BandwidthExhaustionDetected:	containsNetworkThreat(spamSimulation.Threats, ThreatBandwidthExhaustion),
		ChunkCorruptionDetected:	containsNetworkThreat(securityDecision.Threats, ThreatChunkCorruption),
		DiscoveryPoisoningDetected:	forgedDiscoveryErr != nil && discoveryResponse.OnChainProof.ProofHash != "",
		CriticalChannelsAvailable:	true,
	}
}

func adaptivePeerIDs(peers []AdaptivePeer) []string {
	out := make([]string, len(peers))
	for i, peer := range peers {
		out[i] = normalizeHashText(peer.NodeID)
	}
	sortStrings(out)
	return out
}

func requiredCoverageTests(specs []NetworkingTestCoverageSpec) []NetworkingRequiredTest {
	out := make([]NetworkingRequiredTest, len(specs))
	for i, spec := range specs {
		out[i] = spec.Test
	}
	sortRequiredTests(out)
	return out
}

func networkingAcceptanceCriterionIDs(criteria []NetworkingAcceptanceCriterion) []NetworkingAcceptanceCriterion {
	out := append([]NetworkingAcceptanceCriterion(nil), criteria...)
	sortNetworkingAcceptanceCriteria(out)
	return out
}

func testNetworkingAcceptanceEvidence() []NetworkingAcceptanceEvidence {
	return []NetworkingAcceptanceEvidence{
		{
			Criterion:	AcceptanceCriterionL0CometBFTProtected,
			Evidence:	[]string{"TestLayerStackPreservesCometBFTBaselineAndExtensionOrder", "TestL0ScheduleKeepsConsensusAheadOfServiceAndBulkTraffic"},
			Accepted:	true,
		},
		{
			Criterion:	AcceptanceCriterionANAChannelQoS,
			Evidence:	[]string{"TestDefaultANAValidatesCometBFTBaselineAndResponsibilities", "TestANAPropagationKeepsConsensusOnCometBFTAndFanoutForService"},
			Accepted:	true,
		},
		{
			Criterion:	AcceptanceCriterionL1IdentitySessions,
			Evidence:	[]string{"TestNodeRecordSignatureIdentityAndExpiry", "TestSessionHandshakeRejectsReplayExpiredRecordsAndMismatches"},
			Accepted:	true,
		},
		{
			Criterion:	AcceptanceCriterionL2Overlays,
			Evidence:	[]string{"TestOverlayManagerFormsZoneAndServiceOverlays", "TestRoutingGraphBuildsDeterministicCommittedRoutesAndFallback"},
			Accepted:	true,
		},
		{
			Criterion:	AcceptanceCriterionL3AetherMesh,
			Evidence:	[]string{"TestAetherMeshMessageTypesMapToChannels", "TestAetherMeshCrossZoneAndConsensusProofRules"},
			Accepted:	true,
		},
		{
			Criterion:	AcceptanceCriterionRL2Streaming,
			Evidence:	[]string{"TestRL2ChunkDescriptorsVerifyOrderedMerkleRootAndChunkBytes", "TestRL2InterruptedTransferResumesAndRejectsInvalidChunks"},
			Accepted:	true,
		},
		{
			Criterion:	AcceptanceCriterionDRTDiscovery,
			Evidence:	[]string{"TestDiscoveryRecordMustBeSignedExpiringAndProofChecked", "TestSignedDiscoveryRecordStoreFindRenewAndRevoke"},
			Accepted:	true,
		},
		{
			Criterion:	AcceptanceCriterionHybridBroadcast,
			Evidence:	[]string{"TestBroadcastMessageSignsDeduplicatesAndRejectsForgedOrExpired", "TestBroadcastDedupCacheDropsDuplicatesDetectsConflictsAndPrunes"},
			Accepted:	true,
		},
		{
			Criterion:	AcceptanceCriterionSecurityControls,
			Evidence:	[]string{"TestNetworkSecurityReplayChannelBindingAndQoSIsolation", "TestNetworkSecuritySimulationsCoverEclipseSpamAndRoutingManipulation"},
			Accepted:	true,
		},
		{
			Criterion:	AcceptanceCriterionCosmosABCIIntegration,
			Evidence:	[]string{"TestCosmosCometBFTCompatibilityPlanPreservesRequiredSurfaces", "TestABCIIntegrationPlanKeepsLiveNetworkStateOutOfFinalizeBlock"},
			Accepted:	true,
		},
		{
			Criterion:	AcceptanceCriterionRequiredTestCoverage,
			Evidence:	[]string{"TestRequiredNetworkingTestCoverageValidatesUnitAndIntegrationMatrix", "TestNetworkingTestCoverageReportRequiresAllRequiredEvidence"},
			Accepted:	true,
		},
	}
}

func testNetworkingObservabilityMetrics() []NetworkingMetricSample {
	return []NetworkingMetricSample{
		{Metric: ObservableMetricActivePeers, Value: 8, Height: 50},
		{Metric: ObservableMetricPeersByRole, Labels: []string{"role=SERVICE_NODE"}, Value: 2, Height: 50},
		{Metric: ObservableMetricActiveSessions, Value: 3, Height: 50},
		{Metric: ObservableMetricStreamsByChannelType, Labels: []string{"channel=EXECUTION_CHANNEL"}, Value: 2, Height: 50},
		{Metric: ObservableMetricPerChannelBandwidth, Labels: []string{"channel=SERVICE_CHANNEL"}, Value: 4096, Height: 50},
		{Metric: ObservableMetricPeerScore, Labels: []string{"node=fast"}, Value: 8500, Height: 50},
		{Metric: ObservableMetricOverlaySize, Labels: []string{"overlay=zone"}, Value: 5, Height: 50},
		{Metric: ObservableMetricOverlayChurn, Labels: []string{"overlay=zone"}, Value: 1, Height: 50},
		{Metric: ObservableMetricDiscoveryQueryLatency, Value: 30, Height: 50},
		{Metric: ObservableMetricBroadcastDedupHitRate, Value: 5000, Height: 50},
		{Metric: ObservableMetricRL2TransferThroughput, Value: 1 << 20, Height: 50},
		{Metric: ObservableMetricRL2ChunkRetryRate, Value: 2, Height: 50},
		{Metric: ObservableMetricBlockPropagationLatency, Value: 15, Height: 50},
		{Metric: ObservableMetricCrossZoneMessageDeliveryLatency, Value: 60, Height: 50},
		{Metric: ObservableMetricServiceTrafficVolume, Value: 8192, Height: 50},
		{Metric: ObservableMetricRoutingFailureCount, Value: 1, Height: 50},
	}
}

func testNetworkingObservabilityEvents() []NetworkingEventRecord {
	nodeID := HashParts("observability-node")
	peerID := HashParts("observability-peer")
	overlayID := HashParts("observability-overlay")
	transferID := HashParts("observability-transfer")
	messageID := HashParts("observability-message")
	evidenceHash := HashParts("observability-evidence")
	return []NetworkingEventRecord{
		NewNetworkingEventRecord(ObservableEventNetworkNodeRegistered, nodeID, "", "", "", "", "", 50),
		NewNetworkingEventRecord(ObservableEventNetworkSessionOpened, peerID, "", ChannelConsensus, "", "", "", 51),
		NewNetworkingEventRecord(ObservableEventNetworkSessionClosed, peerID, "", ChannelConsensus, "", "", "", 52),
		NewNetworkingEventRecord(ObservableEventNetworkPeerScoreUpdated, peerID, "", "", "", "", evidenceHash, 53),
		NewNetworkingEventRecord(ObservableEventNetworkOverlayJoined, nodeID, overlayID, ChannelRouting, "", "", "", 54),
		NewNetworkingEventRecord(ObservableEventNetworkOverlayLeft, nodeID, overlayID, ChannelRouting, "", "", "", 55),
		NewNetworkingEventRecord(ObservableEventNetworkDiscoveryRecordStored, nodeID, overlayID, ChannelDiscovery, "", "", evidenceHash, 56),
		NewNetworkingEventRecord(ObservableEventNetworkDiscoveryRecordExpired, nodeID, overlayID, ChannelDiscovery, "", "", evidenceHash, 57),
		NewNetworkingEventRecord(ObservableEventNetworkRL2TransferStarted, nodeID, overlayID, ChannelData, transferID, "", "", 58),
		NewNetworkingEventRecord(ObservableEventNetworkRL2TransferCompleted, nodeID, overlayID, ChannelData, transferID, "", "", 59),
		NewNetworkingEventRecord(ObservableEventNetworkInvalidChunk, peerID, overlayID, ChannelData, transferID, "", evidenceHash, 60),
		NewNetworkingEventRecord(ObservableEventNetworkBroadcastConflict, peerID, overlayID, ChannelBlock, "", messageID, evidenceHash, 61),
		NewNetworkingEventRecord(ObservableEventNetworkRouteFailed, peerID, overlayID, ChannelRouting, "", messageID, evidenceHash, 62),
	}
}

func networkingAlertRuleIDs(rules []NetworkingAlertRule) []NetworkingObservableAlert {
	out := make([]NetworkingObservableAlert, len(rules))
	for i, rule := range rules {
		out[i] = NormalizeNetworkingAlertRule(rule).Alert
	}
	sortObservableAlerts(out)
	return out
}

func networkingNonGoalIDs(nonGoals []NetworkingNonGoal) []NetworkingNonGoal {
	out := append([]NetworkingNonGoal(nil), nonGoals...)
	sortNetworkingNonGoals(out)
	return out
}

func testNetworkingAlertSignals(rules []NetworkingAlertRule) []NetworkingAlertSignal {
	nodeID := HashParts("observability-alert-node")
	overlayID := HashParts("observability-alert-overlay")
	signals := make([]NetworkingAlertSignal, 0, len(rules))
	for i, rule := range NormalizeNetworkingAlertRules(rules) {
		observed := rule.Threshold + 1
		if rule.Condition == NetworkingAlertConditionBelowThreshold {
			observed = rule.Threshold - 1
		}
		var sourceMetric NetworkingObservableMetric
		var sourceEvent NetworkingObservableEvent
		if len(rule.SourceMetrics) > 0 {
			sourceMetric = rule.SourceMetrics[0]
		}
		if len(rule.SourceEvents) > 0 {
			sourceEvent = rule.SourceEvents[0]
		}
		signals = append(signals, NewNetworkingAlertSignal(rule, sourceMetric, sourceEvent, nodeID, overlayID, observed, uint64(70+i)))
	}
	return signals
}

func testRequiredNetworkingCoverageEvidence() []NetworkingTestCoverageEvidence {
	return []NetworkingTestCoverageEvidence{
		{
			Test:		RequiredTestNodeIDDerivation,
			Category:	NetworkingTestCoverageUnit,
			TestNames:	[]string{"TestNodeRecordSignatureIdentityAndExpiry"},
			Passed:		true,
		},
		{
			Test:		RequiredTestNodeRecordSignature,
			Category:	NetworkingTestCoverageUnit,
			TestNames:	[]string{"TestNodeRecordSignatureIdentityAndExpiry", "TestANAValidatesSignedPeerRoleAdvertisements"},
			Passed:		true,
		},
		{
			Test:		RequiredTestSessionHandshake,
			Category:	NetworkingTestCoverageUnit,
			TestNames:	[]string{"TestSessionHandshakeRejectsReplayExpiredRecordsAndMismatches", "TestSessionHandshakeNegotiatesEncryptedMultiplexedStreams"},
			Passed:		true,
		},
		{
			Test:		RequiredTestStreamPriority,
			Category:	NetworkingTestCoverageUnit,
			TestNames:	[]string{"TestAetherMeshMessageTypesMapToChannels", "TestStreamPayloadTypeMapsToRL2AndRejectsInvalidBounds"},
			Passed:		true,
		},
		{
			Test:		RequiredTestOverlayMembership,
			Category:	NetworkingTestCoverageUnit,
			TestNames:	[]string{"TestOverlayMembershipProofAuthorizesServiceStakeAndSignedRecords", "TestOverlayManagerFormsZoneAndServiceOverlays"},
			Passed:		true,
		},
		{
			Test:		RequiredTestRouteCost,
			Category:	NetworkingTestCoverageUnit,
			TestNames:	[]string{"TestRoutingGraphBuildsDeterministicCommittedRoutesAndFallback", "TestAetherMeshRouteUsesOverlayAndServicePeers"},
			Passed:		true,
		},
		{
			Test:		RequiredTestNetworkMessageID,
			Category:	NetworkingTestCoverageUnit,
			TestNames:	[]string{"TestNetworkMessageHardRulesRequireReplaySafeCommitments", "TestAetherMeshCrossZoneAndConsensusProofRules"},
			Passed:		true,
		},
		{
			Test:		RequiredTestDiscoveryRecordExpiry,
			Category:	NetworkingTestCoverageUnit,
			TestNames:	[]string{"TestDiscoveryRecordMustBeSignedExpiringAndProofChecked", "TestSignedDiscoveryRecordStoreFindRenewAndRevoke"},
			Passed:		true,
		},
		{
			Test:		RequiredTestChunkHashVerification,
			Category:	NetworkingTestCoverageUnit,
			TestNames:	[]string{"TestChunkPayloadRoundTripAndCorruptionDetection", "TestRL2ChunkDescriptorsVerifyOrderedMerkleRootAndChunkBytes"},
			Passed:		true,
		},
		{
			Test:		RequiredTestBroadcastDeduplication,
			Category:	NetworkingTestCoverageUnit,
			TestNames:	[]string{"TestBroadcastMessageSignsDeduplicatesAndRejectsForgedOrExpired", "TestBroadcastDedupCacheDropsDuplicatesDetectsConflictsAndPrunes"},
			Passed:		true,
		},
		{
			Test:		RequiredTestCometBFTANAConsensus,
			Category:	NetworkingTestCoverageIntegration,
			TestNames:	[]string{"TestANAPropagationKeepsConsensusOnCometBFTAndFanoutForService", "TestL0ScheduleKeepsConsensusAheadOfServiceAndBulkTraffic"},
			Passed:		true,
		},
		{
			Test:		RequiredTestMultiplexedSessionStreams,
			Category:	NetworkingTestCoverageIntegration,
			TestNames:	[]string{"TestSessionHandshakeNegotiatesEncryptedMultiplexedStreams", "TestNetworkSecurityReplayChannelBindingAndQoSIsolation"},
			Passed:		true,
		},
		{
			Test:		RequiredTestZoneOverlayFormation,
			Category:	NetworkingTestCoverageIntegration,
			TestNames:	[]string{"TestOverlayManagerFormsZoneAndServiceOverlays", "TestOverlayRouteBuildsZoneLocalAndFallbackPlans"},
			Passed:		true,
		},
		{
			Test:		RequiredTestServiceOverlayFormation,
			Category:	NetworkingTestCoverageIntegration,
			TestNames:	[]string{"TestOverlayManagerFormsZoneAndServiceOverlays", "TestAetherMeshRouteUsesOverlayAndServicePeers"},
			Passed:		true,
		},
		{
			Test:		RequiredTestCrossZoneDelivery,
			Category:	NetworkingTestCoverageIntegration,
			TestNames:	[]string{"TestCrossZoneSequenceTrackerAssignsAndRejectsReplayAndGaps", "TestReceiptDeliveryProtocolAcknowledgesAndFeedsMetrics"},
			Passed:		true,
		},
		{
			Test:		RequiredTestRL2BlockChunkTransfer,
			Category:	NetworkingTestCoverageIntegration,
			TestNames:	[]string{"TestRL2OfferStateMachineResumesInterruptedTransferAndCompletes", "TestBlockHeaderFirstPropagationVerifiesChunksProofsAndReconstructs"},
			Passed:		true,
		},
		{
			Test:		RequiredTestResumableStateSnapshot,
			Category:	NetworkingTestCoverageIntegration,
			TestNames:	[]string{"TestRL2MissingChunksRequestUsesVerifiedBitmapResumeToken", "TestRL2OfferStateMachineResumesInterruptedTransferAndCompletes"},
			Passed:		true,
		},
		{
			Test:		RequiredTestDiscoveryProofLookup,
			Category:	NetworkingTestCoverageIntegration,
			TestNames:	[]string{"TestDiscoveryResponseSignsResultsAndProofAttachment", "TestNetworkingAPIIntegrationBuildsDiagnosticsProofsAndRouteHints"},
			Passed:		true,
		},
		{
			Test:		RequiredTestHeaderFirstPropagation,
			Category:	NetworkingTestCoverageIntegration,
			TestNames:	[]string{"TestBlockHeaderFirstPropagationVerifiesChunksProofsAndReconstructs", "TestBroadcastForwardingUsesTreeThenGossipFallbackAndOverlayFanout"},
			Passed:		true,
		},
		{
			Test:		RequiredTestReplayedHandshake,
			Category:	NetworkingTestCoverageSecurity,
			TestNames:	[]string{"TestSessionHandshakeRejectsReplayExpiredRecordsAndMismatches", "TestNetworkSecurityReplayChannelBindingAndQoSIsolation"},
			Passed:		true,
		},
		{
			Test:		RequiredTestForgedNodeAdvertisement,
			Category:	NetworkingTestCoverageSecurity,
			TestNames:	[]string{"TestANAValidatesSignedPeerRoleAdvertisements", "TestNetworkSecurityAuthenticatesDiscoveryOverlayAndChunks"},
			Passed:		true,
		},
		{
			Test:		RequiredTestExpiredDiscoverySecurity,
			Category:	NetworkingTestCoverageSecurity,
			TestNames:	[]string{"TestDiscoveryResponseRejectsExpiredForgedAndReplayedRecords", "TestNetworkSecurityAuthenticatesDiscoveryOverlayAndChunks"},
			Passed:		true,
		},
		{
			Test:		RequiredTestConflictingBroadcast,
			Category:	NetworkingTestCoverageSecurity,
			TestNames:	[]string{"TestBroadcastDedupCacheDropsDuplicatesDetectsConflictsAndPrunes", "TestSpamResistanceChunkLimitsDuplicateSuppressionAndSimulations"},
			Passed:		true,
		},
		{
			Test:		RequiredTestInvalidChunkSecurity,
			Category:	NetworkingTestCoverageSecurity,
			TestNames:	[]string{"TestRL2StateMachineClassifiesInvalidChunksAndRootMismatch", "TestNetworkSecurityAuthenticatesDiscoveryOverlayAndChunks"},
			Passed:		true,
		},
		{
			Test:		RequiredTestEclipsePeerSetSimulation,
			Category:	NetworkingTestCoverageSecurity,
			TestNames:	[]string{"TestEclipseResistancePlanMaintainsDiversityAndProofBackedRouting", "TestSpamResistanceChunkLimitsDuplicateSuppressionAndSimulations"},
			Passed:		true,
		},
		{
			Test:		RequiredTestServiceSpamFlood,
			Category:	NetworkingTestCoverageSecurity,
			TestNames:	[]string{"TestSpamResistanceSignedEnvelopeRateLimitsAndResourceBackedAds", "TestSpamResistanceChunkLimitsDuplicateSuppressionAndSimulations"},
			Passed:		true,
		},
		{
			Test:		RequiredTestConsensusUnderBulkLoad,
			Category:	NetworkingTestCoverageSecurity,
			TestNames:	[]string{"TestL0ScheduleKeepsConsensusAheadOfServiceAndBulkTraffic", "TestPerformanceMetricsRejectInvalidBenchmarksAndServiceIsolationFailures"},
			Passed:		true,
		},
		{
			Test:		RequiredTestCrossZoneReplaySecurity,
			Category:	NetworkingTestCoverageSecurity,
			TestNames:	[]string{"TestCrossZoneMessageRequiresSequenceExpiryAndReplayProtection", "TestCrossZoneSequenceTrackerAssignsAndRejectsReplayAndGaps"},
			Passed:		true,
		},
		{
			Test:		RequiredTestBlockHeaderLatency,
			Category:	NetworkingTestCoveragePerformance,
			TestNames:	[]string{"TestPerformanceMetricsSnapshotAggregatesOverlayBandwidthAndScores", "TestPerformanceModelBoundsFanoutLatencyStreamingAndQoS"},
			Passed:		true,
		},
		{
			Test:		RequiredTestBlockReconstructionTime,
			Category:	NetworkingTestCoveragePerformance,
			TestNames:	[]string{"TestPerformanceMetricsSnapshotAggregatesOverlayBandwidthAndScores", "TestBlockHeaderFirstPropagationVerifiesChunksProofsAndReconstructs"},
			Passed:		true,
		},
		{
			Test:		RequiredTestChunkStreamingThroughput,
			Category:	NetworkingTestCoveragePerformance,
			TestNames:	[]string{"TestPerformanceMetricsSnapshotAggregatesOverlayBandwidthAndScores", "TestPerformanceModelBoundsFanoutLatencyStreamingAndQoS"},
			Passed:		true,
		},
		{
			Test:		RequiredTestDiscoveryQueryLatency,
			Category:	NetworkingTestCoveragePerformance,
			TestNames:	[]string{"TestPerformanceMetricsSnapshotAggregatesOverlayBandwidthAndScores", "TestDiscoveryResponseSignsResultsAndProofAttachment"},
			Passed:		true,
		},
		{
			Test:		RequiredTestOverlayJoinLatency,
			Category:	NetworkingTestCoveragePerformance,
			TestNames:	[]string{"TestPerformanceMetricsSnapshotAggregatesOverlayBandwidthAndScores", "TestOverlayManagerFormsZoneAndServiceOverlays"},
			Passed:		true,
		},
		{
			Test:		RequiredTestCrossZonePropagation,
			Category:	NetworkingTestCoveragePerformance,
			TestNames:	[]string{"TestPerformanceMetricsSnapshotAggregatesOverlayBandwidthAndScores", "TestReceiptDeliveryProtocolAcknowledgesAndFeedsMetrics"},
			Passed:		true,
		},
		{
			Test:		RequiredTestServiceTrafficThroughput,
			Category:	NetworkingTestCoveragePerformance,
			TestNames:	[]string{"TestPerformanceMetricsSnapshotAggregatesOverlayBandwidthAndScores", "TestPerformanceMetricsRejectInvalidBenchmarksAndServiceIsolationFailures"},
			Passed:		true,
		},
		{
			Test:		RequiredTestConsensusMixedLoadLatency,
			Category:	NetworkingTestCoveragePerformance,
			TestNames:	[]string{"TestL0ScheduleKeepsConsensusAheadOfServiceAndBulkTraffic", "TestPerformanceMetricsRejectInvalidBenchmarksAndServiceIsolationFailures"},
			Passed:		true,
		},
		{
			Test:		RequiredTestPeerRotationStability,
			Category:	NetworkingTestCoveragePerformance,
			TestNames:	[]string{"TestAdaptiveOverlayGraphRejectsEclipseAndZoneReplacement", "TestAdaptivePeerRotationBoundsChurnAndKeepsStablePeers"},
			Passed:		true,
		},
	}
}

func testSessionRequest(local, remote NodeRecord, openedHeight, expiresHeight uint64, nonce string, channels []ChannelClass) SessionRequest {
	return SessionRequest{
		LocalNodeID:			local.NodeID,
		RemoteNodeID:			remote.NodeID,
		ProtocolVersions:		[]string{DefaultProtocolVersion},
		ChannelClasses:			channels,
		LocalEphemeralPubKey:		bytes.Repeat([]byte{0xa1}, SessionEphemeralKeyBytes),
		RemoteEphemeralPubKey:		bytes.Repeat([]byte{0xb2}, SessionEphemeralKeyBytes),
		SessionSecretCommitmentHash:	HashParts("session-secret", local.NodeID, remote.NodeID, nonce),
		OpenedHeight:			openedHeight,
		ExpiresHeight:			expiresHeight,
		Nonce:				[]byte(nonce),
	}
}

func mustTestStreamSession(t *testing.T, payloadType StreamingPayloadType, priority uint32) StreamSession {
	t.Helper()

	stream, err := NewStreamSession(StreamSession{
		SessionID:		HashParts("priority-stream", string(payloadType)),
		PayloadType:		payloadType,
		Priority:		priority,
		FlowControlWindow:	2048,
		ChunkSize:		512,
		Parallelism:		1,
	})
	require.NoError(t, err)
	stream, err = OpenStreamSession(stream)
	require.NoError(t, err)
	return stream
}

func testDefaultOverlayDescriptor(t *testing.T, overlayType OverlayType) OverlayDescriptor {
	t.Helper()

	for _, desc := range DefaultOverlayDescriptors() {
		if desc.OverlayType == overlayType {
			return desc
		}
	}
	t.Fatalf("missing default overlay descriptor %s", overlayType)
	return OverlayDescriptor{}
}

func signedNodeRecord(t *testing.T, seed byte, salt []byte, expiresHeight uint64, roles ...NodeRole) NodeRecord {
	t.Helper()

	privateKey := deterministicPrivateKey(seed)
	addressHash, err := HashNetworkAddresses([]string{"tcp://127.0.0.1:26656"})
	require.NoError(t, err)
	record, err := SignNodeRecord(NodeRecord{
		Roles:			roles,
		NetworkAddressesHash:	addressHash,
		ZonesSupported:		[]string{"APPLICATION_ZONE", "FINANCIAL_ZONE"},
		ServicesSupported:	[]string{"state-sync", "execution-stream"},
		ProtocolVersions:	[]string{DefaultProtocolVersion},
		ExpiresHeight:		expiresHeight,
	}, privateKey, salt)
	require.NoError(t, err)
	return record
}

func signedNodeRecordWithCapabilities(t *testing.T, seed byte, salt []byte, expiresHeight uint64, roles []NodeRole, zones []string, services []string) NodeRecord {
	t.Helper()

	privateKey := deterministicPrivateKey(seed)
	addressHash, err := HashNetworkAddresses([]string{fmt.Sprintf("tcp://127.0.0.%d:26656", seed)})
	require.NoError(t, err)
	record, err := SignNodeRecord(NodeRecord{
		Roles:			roles,
		NetworkAddressesHash:	addressHash,
		ZonesSupported:		zones,
		ServicesSupported:	services,
		ProtocolVersions:	[]string{DefaultProtocolVersion},
		ExpiresHeight:		expiresHeight,
	}, privateKey, salt)
	require.NoError(t, err)
	return record
}

func testDRTAdvertisement(objectType DRTObjectType, record NodeRecord, overlayID, zoneID, serviceID, endpointHash string, stakeWeight uint64, peerScoreBps uint32, leaseStart, leaseExpires uint64) DRTAdvertisement {
	return DRTAdvertisement{
		ObjectType:	objectType,
		Discovery: DiscoveryRecord{
			Record: record,
		},
		OverlayID:		overlayID,
		ZoneID:			zoneID,
		ServiceID:		serviceID,
		EndpointHash:		endpointHash,
		StakeWeight:		stakeWeight,
		PeerScoreBps:		peerScoreBps,
		LeaseStartHeight:	leaseStart,
		LeaseExpireHeight:	leaseExpires,
	}
}

func testSignedDiscoveryObjectRecord(t *testing.T, owner NodeRecord, seed byte, salt []byte, recordType DRTObjectType, targetID, advertisementHash, zoneID, serviceID, overlayID string, expiresHeight uint64) DiscoveryRecord {
	t.Helper()

	record, err := NewSignedDiscoveryRecord(DiscoveryRecord{
		RecordType:		recordType,
		OwnerNodeID:		owner.NodeID,
		TargetID:		targetID,
		AdvertisementHash:	advertisementHash,
		ZoneID:			zoneID,
		ServiceID:		serviceID,
		OverlayID:		overlayID,
		ExpiresHeight:		expiresHeight,
		Record:			owner,
	}, deterministicPrivateKey(seed), salt)
	require.NoError(t, err)
	return record
}

func deterministicPrivateKey(fill byte) ed25519.PrivateKey {
	seed := bytes.Repeat([]byte{fill}, ed25519.SeedSize)
	return ed25519.NewKeyFromSeed(seed)
}
