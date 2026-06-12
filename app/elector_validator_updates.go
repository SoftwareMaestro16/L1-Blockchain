package app

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtcrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/x/validator-election/types"
)

func (app *L1App) applyElectionValidatorUpdates(req *abci.RequestFinalizeBlock, res *abci.ResponseFinalizeBlock) error {
	if res == nil {
		return nil
	}
	ctx := app.NewUncachedContext(false, cmtproto.Header{Height: req.Height, Time: req.Time})
	electionGenesis, err := app.ValidatorElectionKeeper.ExportGenesisState(ctx)
	if err != nil {
		return err
	}
	currentSet := types.SortValidatorSet(electionGenesis.State.CurrentValidatorSet)
	if len(currentSet) == 0 {
		return nil
	}

	electedByKey := make(map[string]abci.ValidatorUpdate, len(currentSet))
	for _, validator := range currentSet {
		update, err := electionValidatorPowerToABCIUpdate(validator)
		if err != nil {
			return err
		}
		electedByKey[validatorUpdateKey(update)] = update
	}

	zeroUpdates := make(map[string]abci.ValidatorUpdate)
	for _, update := range res.ValidatorUpdates {
		key := validatorUpdateKey(update)
		if _, elected := electedByKey[key]; !elected {
			update.Power = 0
			zeroUpdates[key] = update
		}
	}
	if err := app.addStakingValidatorRemovals(ctx, electedByKey, zeroUpdates); err != nil {
		return err
	}
	if err := addElectionSetRemovals(electionGenesis.State.PreviousValidatorSet, electedByKey, zeroUpdates); err != nil {
		return err
	}

	res.ValidatorUpdates = sortedValidatorUpdates(zeroUpdates, electedByKey)
	return nil
}

func (app *L1App) addStakingValidatorRemovals(ctx sdk.Context, electedByKey map[string]abci.ValidatorUpdate, zeroUpdates map[string]abci.ValidatorUpdate) error {
	validators, err := app.StakingKeeper.GetAllValidators(ctx)
	if err != nil {
		return err
	}
	for _, validator := range validators {
		pubKey, err := validator.TmConsPublicKey()
		if err != nil {
			return err
		}
		update := abci.ValidatorUpdate{PubKey: pubKey, Power: 0}
		key := validatorUpdateKey(update)
		if _, elected := electedByKey[key]; !elected {
			zeroUpdates[key] = update
		}
	}
	return nil
}

func addElectionSetRemovals(previous []types.ValidatorPower, electedByKey map[string]abci.ValidatorUpdate, zeroUpdates map[string]abci.ValidatorUpdate) error {
	for _, validator := range types.SortValidatorSet(previous) {
		update, err := electionValidatorPowerToABCIUpdate(validator)
		if err != nil {
			return err
		}
		key := validatorUpdateKey(update)
		if _, elected := electedByKey[key]; !elected {
			update.Power = 0
			zeroUpdates[key] = update
		}
	}
	return nil
}

func electionValidatorPowerToABCIUpdate(validator types.ValidatorPower) (abci.ValidatorUpdate, error) {
	key, err := parseElectionConsensusPublicKey(validator.ConsensusPublicKey)
	if err != nil {
		return abci.ValidatorUpdate{}, fmt.Errorf("validator election consensus key for %s: %w", validator.OperatorAddress, err)
	}
	if validator.VotingPower > math.MaxInt64 {
		return abci.ValidatorUpdate{}, errors.New("validator election voting power exceeds CometBFT int64 power")
	}
	return abci.ValidatorUpdate{
		PubKey:	cmtcrypto.PublicKey{Sum: &cmtcrypto.PublicKey_Ed25519{Ed25519: key}},
		Power:	int64(validator.VotingPower),
	}, nil
}

func parseElectionConsensusPublicKey(text string) ([]byte, error) {
	kind, value, found := strings.Cut(strings.TrimSpace(text), ":")
	if !found || kind != "ed25519" {
		return nil, errors.New("must use ed25519:<32-byte-hex-or-base64>")
	}
	value = strings.TrimSpace(value)
	key, err := hex.DecodeString(value)
	if err != nil || len(key) != 32 {
		key, err = base64.StdEncoding.DecodeString(value)
	}
	if err != nil || len(key) != 32 {
		return nil, errors.New("ed25519 public key must be exactly 32 bytes")
	}
	return append([]byte(nil), key...), nil
}

func validatorUpdateKey(update abci.ValidatorUpdate) string {
	return hex.EncodeToString(update.PubKey.GetEd25519())
}

func sortedValidatorUpdates(zeroUpdates, electedByKey map[string]abci.ValidatorUpdate) []abci.ValidatorUpdate {
	keys := make([]string, 0, len(zeroUpdates)+len(electedByKey))
	for key := range zeroUpdates {
		if _, elected := electedByKey[key]; !elected {
			keys = append(keys, key)
		}
	}
	for key := range electedByKey {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	out := make([]abci.ValidatorUpdate, 0, len(keys))
	for _, key := range keys {
		if update, found := electedByKey[key]; found {
			out = append(out, update)
			continue
		}
		out = append(out, zeroUpdates[key])
	}
	return out
}
