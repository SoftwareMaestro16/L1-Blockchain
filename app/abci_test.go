package app

import (
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestVoteExtensionHandlerIsDeterministicAndTestOnly(t *testing.T) {
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
}
