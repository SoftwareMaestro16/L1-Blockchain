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

	session, err := networkingtypes.NegotiateSession(local, remote, networkingtypes.SessionRequest{
		LocalNodeID:      local.NodeID,
		RemoteNodeID:     remote.NodeID,
		ProtocolVersions: []string{networkingtypes.DefaultProtocolVersion},
		OpenedHeight:     20,
		ExpiresHeight:    50,
		Nonce:            []byte("keeper-session"),
	})
	require.NoError(t, err)
	require.NoError(t, k.OpenSession(session, 21))

	sessions, _, err := k.Sessions(nil)
	require.NoError(t, err)
	require.Equal(t, []networkingtypes.SessionChannel{session}, sessions)

	require.NoError(t, k.RegisterRoleCommitment(networkingtypes.RoleCommitment{
		NodeID:         remote.NodeID,
		Role:           networkingtypes.NodeRoleService,
		Bonded:         true,
		CommitmentHash: networkingtypes.HashParts("keeper-service-role"),
		ExpiresHeight:  80,
	}, 22))
}

func signedKeeperNode(t *testing.T, seed byte, salt []byte, expiresHeight uint64, roles ...networkingtypes.NodeRole) networkingtypes.NodeRecord {
	t.Helper()

	privateKey := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{seed}, ed25519.SeedSize))
	addressHash, err := networkingtypes.HashNetworkAddresses([]string{"tcp://127.0.0.1:26656"})
	require.NoError(t, err)
	record, err := networkingtypes.SignNodeRecord(networkingtypes.NodeRecord{
		Roles:                roles,
		NetworkAddressesHash: addressHash,
		ProtocolVersions:     []string{networkingtypes.DefaultProtocolVersion},
		ExpiresHeight:        expiresHeight,
	}, privateKey, salt)
	require.NoError(t, err)
	return record
}
