package keeper

import (
	"bytes"
	"crypto/ed25519"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/x/internal/prototype"
	networkingtypes "github.com/sovereign-l1/l1/x/networking/types"
)

func TestDefaultGenesisIsDisabledAndValid(t *testing.T) {
	gs := DefaultGenesis()

	require.NoError(t, gs.Validate())
	require.False(t, gs.Params.Enabled)
	require.NotEmpty(t, gs.State.ChannelPolicies)
	require.NotEmpty(t, gs.State.OverlayDescriptors)
	require.Empty(t, gs.State.NodeRecords)
	require.Empty(t, gs.State.Sessions)
}

func TestKeeperFeatureGateRejectsNetworkingMutationWhenDisabled(t *testing.T) {
	k := NewKeeper()

	err := k.RegisterNodeRecord(networkingtypes.NodeRecord{}, []byte("salt"), 1)
	require.ErrorContains(t, err, "disabled")
}

func TestKeeperRegistersNodeAndSessionWhenEnabled(t *testing.T) {
	k := NewKeeper()
	gs := DefaultGenesis()
	gs.Params = prototype.TestnetParams()
	require.NoError(t, k.InitGenesis(gs))

	salt := []byte("keeper-network")
	local := signedKeeperNode(t, 0x41, salt, 100, networkingtypes.NodeRoleFull)
	remote := signedKeeperNode(t, 0x42, salt, 100, networkingtypes.NodeRoleService)

	require.NoError(t, k.RegisterNodeRecord(remote, salt, 10))
	require.NoError(t, k.RegisterNodeRecord(local, salt, 10))

	records, page, err := k.NodeRecords(nil)
	require.NoError(t, err)
	require.Zero(t, page.NextOffset)
	require.Len(t, records, 2)

	session, err := networkingtypes.NegotiateSession(local, remote, keeperSessionRequest(local, remote, 20, 50, "keeper-session"))
	require.NoError(t, err)
	require.NoError(t, k.OpenSession(session, 21))

	sessions, _, err := k.Sessions(nil)
	require.NoError(t, err)
	expected := networkingtypes.EmptyState()
	expected.NodeRecords = []networkingtypes.NodeRecord{local, remote}
	expected.Sessions = []networkingtypes.SessionChannel{session}
	expected = expected.Export()
	require.NoError(t, expected.Validate())
	require.Equal(t, expected.Sessions, sessions)

	require.NoError(t, k.RegisterRoleCommitment(networkingtypes.RoleCommitment{
		NodeID:		remote.NodeID,
		Role:		networkingtypes.NodeRoleService,
		Bonded:		true,
		CommitmentHash:	networkingtypes.HashParts("keeper-service-role"),
		ExpiresHeight:	80,
	}, 22))

	overlay, err := networkingtypes.NewOverlayDescriptor(networkingtypes.OverlayDescriptor{
		OverlayType:	networkingtypes.OverlayTypeService,
		PolicyHash:	networkingtypes.HashParts("keeper-service-overlay"),
		Membership:	networkingtypes.OverlayMembershipServiceAdvertisement,
		Routing:	networkingtypes.RoutingStrategyLowLatencyAdvisory,
		MinPeers:	2,
		MaxPeers:	16,
		Fanout:		4,
		QoSClass:	networkingtypes.QoSClassServiceCall,
		ExpiresHeight:	90,
		Version:	2,
	})
	require.NoError(t, err)
	require.NoError(t, k.RegisterOverlayDescriptor(overlay, 23))
	overlays, _, err := k.OverlayDescriptors(nil)
	require.NoError(t, err)
	require.Contains(t, keeperOverlayIDs(overlays), overlay.OverlayID)
}

func TestKeeperAppliesSignedIdentityTransition(t *testing.T) {
	k := NewKeeper()
	gs := DefaultGenesis()
	gs.Params = prototype.TestnetParams()
	require.NoError(t, k.InitGenesis(gs))

	salt := []byte("keeper-network")
	oldPrivateKey := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{0x51}, ed25519.SeedSize))
	newPrivateKey := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{0x52}, ed25519.SeedSize))
	oldRecord := signedKeeperNode(t, 0x51, salt, 100, networkingtypes.NodeRoleService)
	newRecord := signedKeeperNode(t, 0x52, salt, 100, networkingtypes.NodeRoleService)
	require.NoError(t, k.RegisterNodeRecord(oldRecord, salt, 10))

	transition, err := networkingtypes.SignIdentityTransition(oldRecord, newRecord, oldPrivateKey, newPrivateKey, salt, 20, 80, []byte("keeper-identity-rotation"))
	require.NoError(t, err)
	require.NoError(t, k.ApplyIdentityTransition(transition, newRecord, salt, 20))

	records, _, err := k.NodeRecords(nil)
	require.NoError(t, err)
	require.Len(t, records, 1)
	require.Equal(t, newRecord.NodeID, records[0].NodeID)
	require.Len(t, k.ExportGenesis().State.IdentityTransitions, 1)
}

func signedKeeperNode(t *testing.T, seed byte, salt []byte, expiresHeight uint64, roles ...networkingtypes.NodeRole) networkingtypes.NodeRecord {
	t.Helper()

	privateKey := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{seed}, ed25519.SeedSize))
	addressHash, err := networkingtypes.HashNetworkAddresses([]string{"tcp://127.0.0.1:26656"})
	require.NoError(t, err)
	record, err := networkingtypes.SignNodeRecord(networkingtypes.NodeRecord{
		Roles:			roles,
		NetworkAddressesHash:	addressHash,
		ProtocolVersions:	[]string{networkingtypes.DefaultProtocolVersion},
		ExpiresHeight:		expiresHeight,
	}, privateKey, salt)
	require.NoError(t, err)
	return record
}

func keeperSessionRequest(local, remote networkingtypes.NodeRecord, openedHeight, expiresHeight uint64, nonce string) networkingtypes.SessionRequest {
	return networkingtypes.SessionRequest{
		LocalNodeID:			local.NodeID,
		RemoteNodeID:			remote.NodeID,
		ProtocolVersions:		[]string{networkingtypes.DefaultProtocolVersion},
		LocalEphemeralPubKey:		bytes.Repeat([]byte{0xc1}, networkingtypes.SessionEphemeralKeyBytes),
		RemoteEphemeralPubKey:		bytes.Repeat([]byte{0xd2}, networkingtypes.SessionEphemeralKeyBytes),
		SessionSecretCommitmentHash:	networkingtypes.HashParts("keeper-session-secret", local.NodeID, remote.NodeID, nonce),
		OpenedHeight:			openedHeight,
		ExpiresHeight:			expiresHeight,
		Nonce:				[]byte(nonce),
	}
}

func keeperOverlayIDs(descriptors []networkingtypes.OverlayDescriptor) []string {
	out := make([]string, len(descriptors))
	for i, desc := range descriptors {
		out[i] = networkingtypes.NormalizeOverlayDescriptor(desc).OverlayID
	}
	return out
}
