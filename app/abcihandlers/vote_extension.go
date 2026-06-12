package abcihandlers

import (
	"bytes"
	"encoding/json"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type VoteExtensionHandler struct{}

type VoteExtension struct {
	Kind	string
	Hash	[]byte
	Height	int64
	Data	[]byte
}

const (
	VoteExtensionKindValidatorTelemetrySummary	= "validator_telemetry_summary"
	VoteExtensionKindOracleFutureExtension		= "oracle_future_extension"
	VoteExtensionKindEncryptedMempoolShare		= "encrypted_mempool_share"

	MaxVoteExtensionBytes		= 512
	MaxVoteExtensionDataBytes	= 128
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
			Kind:	VoteExtensionKindValidatorTelemetrySummary,
			Hash:	req.Hash,
			Height:	req.Height,
			Data:	DeterministicVoteExtensionData(req.Height, req.Hash),
		}

		bz, err := json.Marshal(ve)
		if err != nil {
			return nil, fmt.Errorf("failed to encode vote extension: %w", err)
		}
		if len(bz) > MaxVoteExtensionBytes {
			return nil, fmt.Errorf("vote extension exceeds %d bytes", MaxVoteExtensionBytes)
		}

		return &abci.ResponseExtendVote{VoteExtension: bz}, nil
	}
}

func (h *VoteExtensionHandler) VerifyVoteExtension() sdk.VerifyVoteExtensionHandler {
	return func(_ sdk.Context, req *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error) {
		var ve VoteExtension

		if len(req.ValidatorAddress) == 0 {
			return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}, nil
		}
		if len(req.VoteExtension) == 0 || len(req.VoteExtension) > MaxVoteExtensionBytes {
			return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}, nil
		}
		if err := json.Unmarshal(req.VoteExtension, &ve); err != nil {
			return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}, nil
		}

		switch {
		case !AllowedVoteExtensionKind(ve.Kind):
			return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}, nil
		case len(ve.Data) > MaxVoteExtensionDataBytes:
			return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}, nil
		case req.Height != ve.Height:
			return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}, nil
		case !bytes.Equal(req.Hash, ve.Hash):
			return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}, nil
		case !bytes.Equal(ve.Data, DeterministicVoteExtensionData(req.Height, req.Hash)):
			return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_REJECT}, nil
		}

		return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_ACCEPT}, nil
	}
}

func AllowedVoteExtensionKind(kind string) bool {
	switch kind {
	case VoteExtensionKindValidatorTelemetrySummary,
		VoteExtensionKindOracleFutureExtension,
		VoteExtensionKindEncryptedMempoolShare:
		return true
	default:
		return false
	}
}

func DeterministicVoteExtensionData(height int64, hash []byte) []byte {
	data := make([]byte, 0, len(hash)+32)
	data = append(data, []byte(fmt.Sprintf("aetra-test-vote-extension:%d:", height))...)
	data = append(data, hash...)
	return data
}
