package keeper

import (
	"context"
	"errors"
	"math"

	"github.com/sovereign-l1/l1/x/validator-election/types"
	validatorregistrytypes "github.com/sovereign-l1/l1/x/validator-registry/types"
)

type ValidatorRegistryReader interface {
	Validator(operator string) (validatorregistrytypes.ValidatorRecord, bool, error)
	ValidatorParams() validatorregistrytypes.Params
}

type ValidatorInsuranceReader interface {
	ValidateValidatorActivation(validatorAddress string) error
}

type StakingSetSyncer interface {
	ApplyValidatorSetUpdate(context.Context, StakingValidatorUpdate) error
}

type StakingValidatorUpdate struct {
	OperatorAddress		string
	ConsensusPublicKey	string
	VotingPower		uint64
	Remove			bool
}

func (k *Keeper) ApplyRegisteredValidator(authority, operator string, height uint64, registry ValidatorRegistryReader, insurance ValidatorInsuranceReader) (types.CandidateApplication, error) {
	if registry == nil {
		return types.CandidateApplication{}, errors.New("validator election registry reader is required")
	}
	record, found, err := registry.Validator(operator)
	if err != nil {
		return types.CandidateApplication{}, err
	}
	if !found {
		return types.CandidateApplication{}, errors.New("validator election registry validator not found")
	}
	if err := ValidateRegistryEligibility(record, registry.ValidatorParams(), insurance); err != nil {
		return types.CandidateApplication{}, err
	}
	return k.ApplyForValidatorSet(types.MsgApplyForValidatorSet{
		Authority:	authority,
		Application: types.CandidateApplication{
			OperatorAddress:	record.OperatorAddress,
			ConsensusPublicKey:	record.ConsensusPublicKey,
			RequestedPower:		registryVotingPower(record, k.genesis.Params),
			SelfBond:		registryVotingPower(record, k.genesis.Params),
			ValidatorStatus:	record.Status,
		},
		Height:	height,
	})
}

func ValidateRegistryEligibility(record validatorregistrytypes.ValidatorRecord, params validatorregistrytypes.Params, insurance ValidatorInsuranceReader) error {
	record = record.Normalize(params)
	if err := record.Validate(params); err != nil {
		return err
	}
	switch record.Status {
	case validatorregistrytypes.StatusCandidate, validatorregistrytypes.StatusActive:
	default:
		return errors.New("validator election registry validator status is not eligible")
	}
	if len(record.SlashingHistory) != 0 {
		return errors.New("validator election slashed validator is not eligible")
	}
	if insurance != nil {
		if err := insurance.ValidateValidatorActivation(record.OperatorAddress); err != nil {
			return err
		}
	}
	return nil
}

func (k *Keeper) CommitElectionWithRegistryPolicy(msg types.MsgCommitElection, registryParams validatorregistrytypes.Params, testnetOverride bool) (types.ElectionResult, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.ElectionResult{}, err
	}
	if msg.Height == 0 || msg.Height < k.genesis.State.ElectionWindow.WithdrawDeadlineHeight {
		return types.ElectionResult{}, errors.New("validator election cannot commit before withdrawal deadline")
	}
	nextSet := k.computeNextSet()
	if err := registryParams.ValidateActiveValidatorCount(uint32(len(nextSet)), testnetOverride); err != nil {
		return types.ElectionResult{}, err
	}
	result := types.ElectionResult{
		Epoch:		k.genesis.State.ElectionEpoch,
		Height:		msg.Height,
		NextSet:	types.SortValidatorSet(nextSet),
		Committed:	true,
	}
	next := cloneGenesis(k.genesis)
	next.State.NextValidatorSet = result.NextSet
	next.State.ElectionResults = upsertResult(next.State.ElectionResults, result)
	next.State.CandidateApplications = markPendingApplicationsCommitted(next.State.CandidateApplications, msg.Height)
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return types.ElectionResult{}, err
	}
	if err := k.saveGenesis(next); err != nil {
		return types.ElectionResult{}, err
	}
	return result, nil
}

func (k Keeper) SyncCurrentSetToStaking(ctx context.Context, syncer StakingSetSyncer) ([]StakingValidatorUpdate, error) {
	if syncer == nil {
		return nil, errors.New("validator election staking syncer is required")
	}
	current := types.SortValidatorSet(k.genesis.State.CurrentValidatorSet)
	previous := types.SortValidatorSet(k.genesis.State.PreviousValidatorSet)
	currentByOperator := make(map[string]types.ValidatorPower, len(current))
	for _, validator := range current {
		currentByOperator[validator.OperatorAddress] = validator
	}
	updates := make([]StakingValidatorUpdate, 0, len(current)+len(previous))
	for _, validator := range previous {
		if _, found := currentByOperator[validator.OperatorAddress]; found {
			continue
		}
		updates = append(updates, StakingValidatorUpdate{
			OperatorAddress:	validator.OperatorAddress,
			ConsensusPublicKey:	validator.ConsensusPublicKey,
			Remove:			true,
		})
	}
	for _, validator := range current {
		updates = append(updates, StakingValidatorUpdate{
			OperatorAddress:	validator.OperatorAddress,
			ConsensusPublicKey:	validator.ConsensusPublicKey,
			VotingPower:		validator.VotingPower,
		})
	}
	for _, update := range updates {
		if err := syncer.ApplyValidatorSetUpdate(ctx, update); err != nil {
			return nil, err
		}
	}
	return updates, nil
}

func registryVotingPower(record validatorregistrytypes.ValidatorRecord, params types.Params) uint64 {
	total := record.SelfBond + record.NominatorBond
	power := total / validatorregistrytypes.DefaultAETBaseUnits
	if power == 0 {
		power = 1
	}
	if power > params.MaxValidatorPower {
		return params.MaxValidatorPower
	}
	if power > math.MaxInt64 {
		return math.MaxInt64
	}
	return power
}
