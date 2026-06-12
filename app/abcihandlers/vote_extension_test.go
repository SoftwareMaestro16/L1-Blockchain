package abcihandlers

import (
	"encoding/json"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestVoteExtensionHandlerIsDeterministicAndRejectsTampering(t *testing.T) {
	handler := NewVoteExtensionHandler()
	req := &abci.RequestExtendVote{Height: 10, Hash: []byte("block-hash")}

	first, err := handler.ExtendVote()(sdk.Context{}, req)
	require.NoError(t, err)
	second, err := handler.ExtendVote()(sdk.Context{}, req)
	require.NoError(t, err)
	require.Equal(t, first.VoteExtension, second.VoteExtension)

	verify, err := handler.VerifyVoteExtension()(sdk.Context{}, &abci.RequestVerifyVoteExtension{
		Height:			req.Height,
		Hash:			req.Hash,
		ValidatorAddress:	[]byte("validator-a"),
		VoteExtension:		first.VoteExtension,
	})
	require.NoError(t, err)
	require.Equal(t, abci.ResponseVerifyVoteExtension_ACCEPT, verify.Status)

	rejected, err := handler.VerifyVoteExtension()(sdk.Context{}, &abci.RequestVerifyVoteExtension{
		Height:			req.Height + 1,
		Hash:			req.Hash,
		ValidatorAddress:	[]byte("validator-a"),
		VoteExtension:		first.VoteExtension,
	})
	require.NoError(t, err)
	require.Equal(t, abci.ResponseVerifyVoteExtension_REJECT, rejected.Status)
}

func TestVoteExtensionPolicyRejectsUnsignedOversizedAndUnknownKinds(t *testing.T) {
	handler := NewVoteExtensionHandler()
	req := &abci.RequestExtendVote{Height: 10, Hash: []byte("block-hash")}
	extended, err := handler.ExtendVote()(sdk.Context{}, req)
	require.NoError(t, err)
	require.LessOrEqual(t, len(extended.VoteExtension), MaxVoteExtensionBytes)

	unsigned, err := handler.VerifyVoteExtension()(sdk.Context{}, &abci.RequestVerifyVoteExtension{
		Height:		req.Height,
		Hash:		req.Hash,
		VoteExtension:	extended.VoteExtension,
	})
	require.NoError(t, err)
	require.Equal(t, abci.ResponseVerifyVoteExtension_REJECT, unsigned.Status)

	oversized, err := handler.VerifyVoteExtension()(sdk.Context{}, &abci.RequestVerifyVoteExtension{
		Height:			req.Height,
		Hash:			req.Hash,
		ValidatorAddress:	[]byte("validator-a"),
		VoteExtension:		make([]byte, MaxVoteExtensionBytes+1),
	})
	require.NoError(t, err)
	require.Equal(t, abci.ResponseVerifyVoteExtension_REJECT, oversized.Status)

	unknown := VoteExtension{
		Kind:	"large_nondeterministic_payload",
		Hash:	req.Hash,
		Height:	req.Height,
		Data:	DeterministicVoteExtensionData(req.Height, req.Hash),
	}
	unknownBz, err := json.Marshal(unknown)
	require.NoError(t, err)
	rejected, err := handler.VerifyVoteExtension()(sdk.Context{}, &abci.RequestVerifyVoteExtension{
		Height:			req.Height,
		Hash:			req.Hash,
		ValidatorAddress:	[]byte("validator-a"),
		VoteExtension:		unknownBz,
	})
	require.NoError(t, err)
	require.Equal(t, abci.ResponseVerifyVoteExtension_REJECT, rejected.Status)
}

func TestVoteExtensionAllowedKindsAreExplicitAndSmall(t *testing.T) {
	require.True(t, AllowedVoteExtensionKind(VoteExtensionKindValidatorTelemetrySummary))
	require.True(t, AllowedVoteExtensionKind(VoteExtensionKindOracleFutureExtension))
	require.True(t, AllowedVoteExtensionKind(VoteExtensionKindEncryptedMempoolShare))
	require.False(t, AllowedVoteExtensionKind("bulk_state_payload"))
	require.LessOrEqual(t, MaxVoteExtensionDataBytes, 128)
	require.LessOrEqual(t, MaxVoteExtensionBytes, 512)
}
