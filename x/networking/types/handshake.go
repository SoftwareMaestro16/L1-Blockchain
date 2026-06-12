package types

import (
	"errors"
	"fmt"
	"sort"
)

type HandshakePhase string

const (
	HandshakePhaseInit			HandshakePhase	= "INIT"
	HandshakePhaseIdentityVerified		HandshakePhase	= "IDENTITY_VERIFIED"
	HandshakePhaseProtocolsNegotiated	HandshakePhase	= "PROTOCOLS_NEGOTIATED"
	HandshakePhaseChannelsNegotiated	HandshakePhase	= "CHANNELS_NEGOTIATED"
	HandshakePhaseKeysEstablished		HandshakePhase	= "KEYS_ESTABLISHED"
	HandshakePhaseEstablished		HandshakePhase	= "ESTABLISHED"
	HandshakePhaseRejected			HandshakePhase	= "REJECTED"
)

type SessionHandshakeState struct {
	Phase			HandshakePhase
	ReplayID		string
	LocalNodeID		string
	RemoteNodeID		string
	CipherSuite		CipherSuite
	ProtocolVersions	[]string
	ChannelClasses		[]ChannelClass
	SessionKeys		SessionKeySet
	Session			SessionChannel
	RejectReason		string
}

type SessionKeyRotationRequest struct {
	SessionID			string
	NewLocalEphemeralPubKey		[]byte
	NewRemoteEphemeralPubKey	[]byte
	NewSecretCommitmentHash		string
	RotatedAtHeight			uint64
	ExpiresHeight			uint64
	Nonce				[]byte
}

func RunSessionHandshake(local, remote NodeRecord, req SessionRequest, networkSalt []byte, currentHeight uint64, seenReplayIDs []string) (SessionHandshakeState, error) {
	state := SessionHandshakeState{
		Phase:		HandshakePhaseInit,
		ReplayID:	ComputeHandshakeReplayID(req),
		LocalNodeID:	normalizeHashText(req.LocalNodeID),
		RemoteNodeID:	normalizeHashText(req.RemoteNodeID),
	}
	if hasString(seenReplayIDs, state.ReplayID) {
		state.Phase = HandshakePhaseRejected
		state.RejectReason = "replayed handshake"
		return state, errors.New("networking replayed session handshake")
	}
	local = NormalizeNodeRecord(local)
	remote = NormalizeNodeRecord(remote)
	req = req.Normalize()
	if err := local.Validate(networkSalt, currentHeight); err != nil {
		return rejectHandshake(state, err)
	}
	if err := remote.Validate(networkSalt, currentHeight); err != nil {
		return rejectHandshake(state, err)
	}
	if req.LocalNodeID != local.NodeID || req.RemoteNodeID != remote.NodeID {
		return rejectHandshake(state, errors.New("networking session handshake node mismatch"))
	}
	state.Phase = HandshakePhaseIdentityVerified

	cipher, err := chooseCipher(req.CipherSuites)
	if err != nil {
		return rejectHandshake(state, err)
	}
	protocols := intersectStrings(local.ProtocolVersions, remote.ProtocolVersions, req.ProtocolVersions)
	if len(protocols) == 0 {
		return rejectHandshake(state, errors.New("networking session handshake has no mutually supported protocol"))
	}
	state.CipherSuite = cipher
	state.ProtocolVersions = protocols
	state.Phase = HandshakePhaseProtocolsNegotiated

	channels := req.ChannelClasses
	if len(channels) == 0 {
		channels = normalizeChannels([]ChannelClass{ChannelConsensus, ChannelBlock, ChannelStateSync, ChannelExecution, ChannelMempool, ChannelService, ChannelRouting, ChannelDiscovery, ChannelData})
	}
	if err := validateChannels(channels); err != nil {
		return rejectHandshake(state, err)
	}
	state.ChannelClasses = append([]ChannelClass(nil), channels...)
	state.Phase = HandshakePhaseChannelsNegotiated

	keys, err := BuildSessionKeySet(req, cipher, protocols, channels)
	if err != nil {
		return rejectHandshake(state, err)
	}
	state.SessionKeys = keys
	state.Phase = HandshakePhaseKeysEstablished

	session, err := NegotiateSession(local, remote, req)
	if err != nil {
		return rejectHandshake(state, err)
	}
	state.Session = session
	state.Phase = HandshakePhaseEstablished
	return state, nil
}

func ComputeHandshakeReplayID(req SessionRequest) string {
	req = req.Normalize()
	return HashParts(
		"session-handshake-replay",
		req.LocalNodeID,
		req.RemoteNodeID,
		fmt.Sprintf("%d", req.HandshakeVersion),
		fmt.Sprintf("%d", req.OpenedHeight),
		fmt.Sprintf("%d", req.ExpiresHeight),
		string(req.Nonce),
		string(req.LocalEphemeralPubKey),
		string(req.RemoteEphemeralPubKey),
		req.SessionSecretCommitmentHash,
	)
}

func RotateSessionKeys(session SessionChannel, req SessionKeyRotationRequest) (SessionChannel, error) {
	if err := session.Validate(); err != nil {
		return SessionChannel{}, err
	}
	req.SessionID = normalizeHashText(req.SessionID)
	req.NewSecretCommitmentHash = normalizeHashText(req.NewSecretCommitmentHash)
	req.NewLocalEphemeralPubKey = cloneBytes(req.NewLocalEphemeralPubKey)
	req.NewRemoteEphemeralPubKey = cloneBytes(req.NewRemoteEphemeralPubKey)
	req.Nonce = cloneBytes(req.Nonce)
	if req.SessionID != session.SessionID {
		return SessionChannel{}, errors.New("networking session key rotation session mismatch")
	}
	if req.RotatedAtHeight <= session.OpenedHeight || req.RotatedAtHeight > session.ExpiresHeight {
		return SessionChannel{}, errors.New("networking session key rotation height is outside session range")
	}
	if req.ExpiresHeight <= req.RotatedAtHeight || req.ExpiresHeight > session.ExpiresHeight {
		return SessionChannel{}, errors.New("networking session key rotation expiry is outside session range")
	}
	if len(req.Nonce) == 0 || len(req.Nonce) > MaxNonceBytes {
		return SessionChannel{}, fmt.Errorf("networking session key rotation nonce must be between 1 and %d bytes", MaxNonceBytes)
	}
	rotationRequest := SessionRequest{
		LocalNodeID:			session.LocalNodeID,
		RemoteNodeID:			session.RemoteNodeID,
		HandshakeVersion:		session.HandshakeVersion,
		CipherSuites:			[]CipherSuite{session.CipherSuite},
		ProtocolVersions:		append([]string(nil), session.ProtocolVersions...),
		ChannelClasses:			streamChannels(session.Streams),
		LocalEphemeralPubKey:		req.NewLocalEphemeralPubKey,
		RemoteEphemeralPubKey:		req.NewRemoteEphemeralPubKey,
		SessionSecretCommitmentHash:	req.NewSecretCommitmentHash,
		OpenedHeight:			req.RotatedAtHeight,
		ExpiresHeight:			req.ExpiresHeight,
		Nonce:				req.Nonce,
		QOSPolicy:			session.QOSPolicy,
	}
	keys, err := BuildSessionKeySet(rotationRequest, session.CipherSuite, session.ProtocolVersions, rotationRequest.ChannelClasses)
	if err != nil {
		return SessionChannel{}, err
	}
	if keys.KeyID == session.SessionKeys.KeyID {
		return SessionChannel{}, errors.New("networking session key rotation must produce a new key id")
	}
	next := cloneSession(session)
	next.SessionKeys = keys
	for i := range next.Streams {
		next.Streams[i].EncryptionContext = streamEncryptionContext(keys.KeyID, next.Streams[i].StreamID)
	}
	if err := next.Validate(); err != nil {
		return SessionChannel{}, err
	}
	return next, nil
}

func IsHandshakePhase(phase HandshakePhase) bool {
	switch phase {
	case HandshakePhaseInit, HandshakePhaseIdentityVerified, HandshakePhaseProtocolsNegotiated, HandshakePhaseChannelsNegotiated, HandshakePhaseKeysEstablished, HandshakePhaseEstablished, HandshakePhaseRejected:
		return true
	default:
		return false
	}
}

func rejectHandshake(state SessionHandshakeState, err error) (SessionHandshakeState, error) {
	state.Phase = HandshakePhaseRejected
	state.RejectReason = err.Error()
	return state, err
}

func streamChannels(streams []StreamSpec) []ChannelClass {
	out := make([]ChannelClass, 0, len(streams))
	seen := make(map[ChannelClass]struct{}, len(streams))
	for _, stream := range streams {
		if _, found := seen[stream.Channel]; found {
			continue
		}
		seen[stream.Channel] = struct{}{}
		out = append(out, stream.Channel)
	}
	sort.SliceStable(out, func(i, j int) bool {
		leftPriority := PriorityForChannel(out[i])
		rightPriority := PriorityForChannel(out[j])
		if leftPriority != rightPriority {
			return leftPriority < rightPriority
		}
		return channelSortRank(out[i]) < channelSortRank(out[j])
	})
	return out
}

func hasString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}
