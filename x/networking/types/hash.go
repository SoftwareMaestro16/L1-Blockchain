package types

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"sort"
)

func ComputeNodeID(identityPubKey, networkSalt []byte) string {
	h := sha256.New()
	writeString(h, "aetra-node-id-v1")
	writeBytes(h, identityPubKey)
	writeBytes(h, networkSalt)
	return hex.EncodeToString(h.Sum(nil))
}

func HashNetworkAddresses(addresses []string) (string, error) {
	normalized, err := normalizeStringSet("network address", addresses, MaxNetworkAddressBytes)
	if err != nil {
		return "", err
	}
	h := sha256.New()
	writeString(h, "aetra-network-addresses-v1")
	for _, address := range normalized {
		writeString(h, address)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func ComputeTransportEnvelopeID(envelope TransportEnvelope) string {
	h := sha256.New()
	writeString(h, "aetra-transport-envelope-v1")
	writeString(h, string(envelope.Channel))
	writeUint64(h, envelope.SizeBytes)
	writeUint64(h, envelope.EnqueuedHeight)
	writeUint64(h, envelope.Sequence)
	writeString(h, envelope.PayloadHash)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeSessionID(req SessionRequest, cipher CipherSuite, protocols []string, channels []ChannelClass) string {
	h := sha256.New()
	writeString(h, "aetra-session-v1")
	writeString(h, ComputeSessionTranscriptHash(req, cipher, protocols, channels))
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeSessionTranscriptHash(req SessionRequest, cipher CipherSuite, protocols []string, channels []ChannelClass) string {
	req = req.Normalize()
	h := sha256.New()
	writeString(h, "aetra-session-transcript-v1")
	writeString(h, req.LocalNodeID)
	writeString(h, req.RemoteNodeID)
	writeUint64(h, uint64(req.HandshakeVersion))
	writeString(h, string(cipher))
	for _, protocol := range protocols {
		writeString(h, protocol)
	}
	for _, channel := range channels {
		writeString(h, string(channel))
	}
	writeUint64(h, req.OpenedHeight)
	writeUint64(h, req.ExpiresHeight)
	writeBytes(h, req.Nonce)
	writeBytes(h, req.LocalEphemeralPubKey)
	writeBytes(h, req.RemoteEphemeralPubKey)
	writeString(h, req.SessionSecretCommitmentHash)
	return hex.EncodeToString(h.Sum(nil))
}

func HashParts(parts ...string) string {
	h := sha256.New()
	writeString(h, "aetra-networking-hash-parts-v1")
	for _, part := range parts {
		writeString(h, part)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func hashBytes(domain string, bz []byte) string {
	h := sha256.New()
	writeString(h, domain)
	writeBytes(h, bz)
	return hex.EncodeToString(h.Sum(nil))
}

func writeString(w interface{ Write([]byte) (int, error) }, value string) {
	writeUint64(w, uint64(len(value)))
	_, _ = w.Write([]byte(value))
}

func writeBytes(w interface{ Write([]byte) (int, error) }, value []byte) {
	writeUint64(w, uint64(len(value)))
	_, _ = w.Write(value)
}

func writeUint64(w interface{ Write([]byte) (int, error) }, value uint64) {
	var bz [8]byte
	binary.BigEndian.PutUint64(bz[:], value)
	_, _ = w.Write(bz[:])
}

func cloneBytes(in []byte) []byte {
	if len(in) == 0 {
		return nil
	}
	return append([]byte(nil), in...)
}

func sortStrings(values []string) {
	sort.SliceStable(values, func(i, j int) bool {
		return values[i] < values[j]
	})
}
