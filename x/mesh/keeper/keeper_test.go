package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/x/internal/prototype"
	meshtypes "github.com/sovereign-l1/l1/x/mesh/types"
)

func TestDefaultGenesisValidates(t *testing.T) {
	gs := DefaultGenesis()

	require.NoError(t, gs.Validate())
	require.False(t, gs.Params.Enabled)
	require.Empty(t, gs.State.Destinations)
}

func TestInitGenesisRejectsCorruptedState(t *testing.T) {
	keeper := NewKeeper()
	gs := DefaultGenesis()
	gs.Version = prototype.CurrentGenesisVersion + 1
	require.ErrorContains(t, keeper.InitGenesis(gs), "unsupported genesis version")

	gs = DefaultGenesis()
	destination := keeperDestination("FINANCIAL_ZONE", "0:0")
	gs.State.Destinations = []meshtypes.MeshDestination{destination, destination}
	require.ErrorContains(t, keeper.InitGenesis(gs), "duplicate mesh destination")
}

func TestFeatureDisabledRejectsMutatingMessages(t *testing.T) {
	keeper := NewKeeper()

	err := keeper.RegisterDestination(keeperDestination("FINANCIAL_ZONE", "0:0"))
	require.ErrorContains(t, err, "disabled")
}

func TestUnauthorizedAuthorityRejected(t *testing.T) {
	keeper := NewKeeper()

	err := keeper.UpdateParams("4:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", prototype.TestnetParams())
	require.ErrorContains(t, err, "authority")
}

func TestRegisterCommitmentAndApplyMessage(t *testing.T) {
	keeper, msg := keeperMeshFixture(t)

	receipt, err := keeper.ApplyMessage(msg, keeperSuccessResult(), 100)
	require.NoError(t, err)
	require.Equal(t, meshtypes.ReceiptStatusSuccess, receipt.Status)

	exported := keeper.ExportGenesis()
	require.NoError(t, exported.Validate())
	require.Len(t, exported.State.ReplayMarkers, 1)
	require.Len(t, exported.State.Receipts, 1)
}

func TestApplyMessageRejectsReplay(t *testing.T) {
	keeper, msg := keeperMeshFixture(t)
	_, err := keeper.ApplyMessage(msg, keeperSuccessResult(), 100)
	require.NoError(t, err)

	_, err = keeper.ApplyMessage(msg, keeperSuccessResult(), 101)
	require.ErrorContains(t, err, "replay")
}

func TestReceiptsQueryPaginationBoundedAndMalformedSafe(t *testing.T) {
	keeper, msg := keeperMeshFixture(t)
	params := prototype.TestnetParams()
	params.DefaultQueryLimit = 1
	params.MaxQueryLimit = 1
	require.NoError(t, keeper.UpdateParams(prototype.DefaultAuthority, params))
	_, err := keeper.ApplyMessage(msg, keeperSuccessResult(), 100)
	require.NoError(t, err)

	receipts, page, err := keeper.Receipts(nil)
	require.NoError(t, err)
	require.Len(t, receipts, 1)
	require.Zero(t, page.NextOffset)

	_, _, err = keeper.Receipts(&prototype.PageRequest{Offset: 99, Limit: 1})
	require.ErrorContains(t, err, "offset")

	_, _, err = keeper.Receipts(&prototype.PageRequest{Limit: params.MaxQueryLimit + 1})
	require.ErrorContains(t, err, "limit")
}

func TestExportImportRoundTripDeterministic(t *testing.T) {
	keeper, msg := keeperMeshFixture(t)
	_, err := keeper.ApplyMessage(msg, keeperSuccessResult(), 100)
	require.NoError(t, err)

	exported := keeper.ExportGenesis()
	imported := NewKeeper()
	require.NoError(t, imported.InitGenesis(exported))

	require.Equal(t, exported, imported.ExportGenesis())
}

func TestMigrationNoopPathValidatesCurrentState(t *testing.T) {
	keeper := NewKeeper()

	err := NewMigrator(&keeper).Migrate1to2()
	require.NoError(t, err)
}

func keeperMeshFixture(t *testing.T) (Keeper, meshtypes.MeshMessage) {
	t.Helper()

	keeper := NewKeeper()
	require.NoError(t, keeper.UpdateParams(prototype.DefaultAuthority, prototype.TestnetParams()))
	require.NoError(t, keeper.RegisterDestination(keeperDestination("CONTRACT_ZONE", "0:1")))
	require.NoError(t, keeper.RegisterDestination(keeperDestination("FINANCIAL_ZONE", "0:0")))

	commitment := meshtypes.FinalizedCommitment{
		ZoneID:		meshtypes.ZoneID("FINANCIAL_ZONE"),
		ShardID:	meshtypes.ShardID("0:0"),
		Height:		90,
		CommitmentHash:	meshtypes.HashParts("source-commitment", "financial", "0:0", "90"),
		MessageRoot:	meshtypes.HashParts("message-root", "financial", "90"),
		ReceiptRoot:	meshtypes.HashParts("receipt-root", "financial", "90"),
	}
	require.NoError(t, keeper.AddFinalizedCommitment(commitment))

	msg, err := meshtypes.NewMessage(meshtypes.MeshMessage{
		SourceZone:		meshtypes.ZoneID("FINANCIAL_ZONE"),
		SourceShard:		meshtypes.ShardID("0:0"),
		DestinationZone:	meshtypes.ZoneID("CONTRACT_ZONE"),
		DestinationShard:	meshtypes.ShardID("0:1"),
		Nonce:			7,
		Sender:			[]byte("orb1sender"),
		Recipient:		[]byte("contract1recipient"),
		AssetCommitment:	meshtypes.HashParts("asset", "100naet"),
		PayloadHash:		meshtypes.HashParts("payload", "execute"),
		TimeoutHeight:		150,
		Finality:		meshtypes.FinalityReference{Height: commitment.Height, CommitmentHash: commitment.CommitmentHash},
		Sequence:		3,
		SourceLogicalTime:	88,
	})
	require.NoError(t, err)
	msg.Proof = meshtypes.BuildProof(msg, commitment)
	return keeper, msg
}

func keeperDestination(zone string, shard string) meshtypes.MeshDestination {
	return meshtypes.MeshDestination{
		ZoneID:		meshtypes.ZoneID(zone),
		ShardID:	meshtypes.ShardID(shard),
		Active:		true,
	}
}

func keeperSuccessResult() meshtypes.ExecutionResult {
	return meshtypes.ExecutionResult{
		Success:	true,
		Code:		0,
		ResultHash:	meshtypes.HashParts("execution", "success"),
	}
}
