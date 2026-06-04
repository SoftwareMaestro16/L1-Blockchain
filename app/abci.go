package app

import (
	"bytes"
	"encoding/json"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	// VoteExtensionHandler defines a deterministic test-only vote extension handler for L1App.
	//
	// NOTE: This implementation is solely used for targeted tests. Production
	// app wiring must not install it unless a real vote extension protocol exists.
	VoteExtensionHandler struct{}

	// VoteExtension defines the structure used to create a dummy vote extension.
	VoteExtension struct {
		Hash   []byte
		Height int64
		Data   []byte
	}
)

func NewVoteExtensionHandler() *VoteExtensionHandler {
	return &VoteExtensionHandler{}
}

func (h *VoteExtensionHandler) SetHandlers(bApp *baseapp.BaseApp) {
	bApp.SetExtendVoteHandler(h.ExtendVote())
	bApp.SetVerifyVoteExtensionHandler(h.VerifyVoteExtension())
}

func (h *VoteExtensionHandler) ExtendVote() sdk.ExtendVoteHandler {
	return func(_ sdk.Context, req *abci.RequestExtendVote) (*abci.ResponseExtendVote, error) {
		ve := VoteExtension{
			Hash:   req.Hash,
			Height: req.Height,
			Data:   deterministicVoteExtensionData(req.Height, req.Hash),
		}

		bz, err := json.Marshal(ve)
		if err != nil {
			return nil, fmt.Errorf("failed to encode vote extension: %w", err)
		}

		return &abci.ResponseExtendVote{VoteExtension: bz}, nil
	}
}

func (h *VoteExtensionHandler) VerifyVoteExtension() sdk.VerifyVoteExtensionHandler {
	return func(ctx sdk.Context, req *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error) {
		var ve VoteExtension

		if err := json.Unmarshal(req.VoteExtension, &ve); err != nil {
			return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}, nil
		}

		switch {
		case req.Height != ve.Height:
			return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}, nil

		case !bytes.Equal(req.Hash, ve.Hash):
			return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}, nil

		case !bytes.Equal(ve.Data, deterministicVoteExtensionData(req.Height, req.Hash)):
			return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}, nil
		}

		return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_ACCEPT}, nil
	}
}

func deterministicVoteExtensionData(height int64, hash []byte) []byte {
	data := make([]byte, 0, len(hash)+32)
	data = append(data, []byte(fmt.Sprintf("orbitalis-test-vote-extension:%d:", height))...)
	data = append(data, hash...)
	return data
}
