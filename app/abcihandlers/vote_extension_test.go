package abcihandlers

import (
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
		Height:        req.Height,
		Hash:          req.Hash,
		VoteExtension: first.VoteExtension,
	})
	require.NoError(t, err)
	require.Equal(t, abci.ResponseVerifyVoteExtension_ACCEPT, verify.Status)

	rejected, err := handler.VerifyVoteExtension()(sdk.Context{}, &abci.RequestVerifyVoteExtension{
		Height:        req.Height + 1,
		Hash:          req.Hash,
		VoteExtension: first.VoteExtension,
	})
	require.NoError(t, err)
	require.Equal(t, abci.ResponseVerifyVoteExtension_REJECT, rejected.Status)
}
