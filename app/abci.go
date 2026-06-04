package app

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strconv"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	DeterministicVoteExtensionsAppOption = "orbitalis.vote_extensions.deterministic_for_testing"
	voteExtensionDomain                  = "orbitalis/vote-extension/v1"
)

type (
	// VoteExtensionHandler is deterministic and must only be enabled through
	// DeterministicVoteExtensionsAppOption in test/dev networks.
	VoteExtensionHandler struct{}

	// VoteExtension defines the explicitly deterministic vote extension payload.
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
			Data:   deterministicVoteExtensionData(req.Hash, req.Height),
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

		case !bytes.Equal(ve.Data, deterministicVoteExtensionData(req.Hash, req.Height)):
			return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}, nil
		}

		return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_ACCEPT}, nil
	}
}

func deterministicVoteExtensionData(hash []byte, height int64) []byte {
	heightBz := make([]byte, 8)
	binary.BigEndian.PutUint64(heightBz, uint64(height))

	sum := sha256.New()
	sum.Write([]byte(voteExtensionDomain))
	sum.Write(heightBz)
	sum.Write(hash)
	return sum.Sum(nil)
}

func deterministicVoteExtensionsEnabled(appOpts servertypes.AppOptions) bool {
	if appOpts == nil {
		return false
	}

	switch value := appOpts.Get(DeterministicVoteExtensionsAppOption).(type) {
	case bool:
		return value
	case string:
		enabled, err := strconv.ParseBool(value)
		return err == nil && enabled
	default:
		return false
	}
}
