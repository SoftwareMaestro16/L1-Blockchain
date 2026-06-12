package keeper

import (
	"context"
	"errors"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/x/internal/prototype"
	"github.com/sovereign-l1/l1/x/validator-election/types"
	validatorregistrytypes "github.com/sovereign-l1/l1/x/validator-registry/types"
)

func TestRegisteredValidatorElectionSyncsToActiveStakingSet(t *testing.T) {
	k := NewKeeper()
	operator := adapterAddress(0x11)
	registry := fakeRegistry{records: map[string]validatorregistrytypes.ValidatorRecord{
		operator: adapterValidator(operator, "ed25519:validator-a", validatorregistrytypes.StatusCandidate),
	}}
	insurance := fakeInsurance{eligible: map[string]bool{operator: true}}

	app, err := k.ApplyRegisteredValidator(prototype.DefaultAuthority, operator, 2, registry, insurance)
	require.NoError(t, err)
	require.Equal(t, operator, app.OperatorAddress)

	result, err := k.CommitElectionWithRegistryPolicy(types.MsgCommitElection{Authority: prototype.DefaultAuthority, Height: 90}, validatorregistrytypes.DefaultParams(), true)
	require.NoError(t, err)
	require.Len(t, result.NextSet, 1)

	transition, err := k.FinalizeElection(types.MsgFinalizeElection{Authority: prototype.DefaultAuthority, Height: 101})
	require.NoError(t, err)
	require.Equal(t, result.NextSet, transition.CurrentSet)
	require.Equal(t, transition.CurrentSet, k.CurrentValidatorSet())

	syncer := &recordingStakingSyncer{}
	updates, err := k.SyncCurrentSetToStaking(context.Background(), syncer)
	require.NoError(t, err)
	require.Equal(t, updates, syncer.updates)
	require.Len(t, updates, 1)
	require.Equal(t, operator, updates[0].OperatorAddress)
	require.Equal(t, uint64(1_000_000), updates[0].VotingPower)
	require.False(t, updates[0].Remove)
}

func TestRegisteredValidatorEligibilityRejectsUnderfundedAndUninsured(t *testing.T) {
	k := NewKeeper()
	operator := adapterAddress(0x22)
	record := adapterValidator(operator, "ed25519:validator-b", validatorregistrytypes.StatusCandidate)
	record.SelfBond = validatorregistrytypes.DefaultMinValidatorStake - 1
	registry := fakeRegistry{records: map[string]validatorregistrytypes.ValidatorRecord{operator: record}}

	_, err := k.ApplyRegisteredValidator(prototype.DefaultAuthority, operator, 2, registry, fakeInsurance{eligible: map[string]bool{operator: true}})
	require.ErrorContains(t, err, "self-stake")

	record.SelfBond = validatorregistrytypes.DefaultMinValidatorStake
	registry.records[operator] = record
	_, err = k.ApplyRegisteredValidator(prototype.DefaultAuthority, operator, 2, registry, fakeInsurance{})
	require.ErrorContains(t, err, "minimum insurance")

	strictParams := validatorregistrytypes.DefaultParams()
	strictParams.MinValidatorStake = validatorregistrytypes.DefaultMinValidatorStake + 1
	strictRegistry := fakeRegistry{
		records:	map[string]validatorregistrytypes.ValidatorRecord{operator: record},
		params:		strictParams,
	}
	_, err = k.ApplyRegisteredValidator(prototype.DefaultAuthority, operator, 2, strictRegistry, fakeInsurance{eligible: map[string]bool{operator: true}})
	require.ErrorContains(t, err, "minimum validator stake")
}

func TestRegisteredValidatorEligibilityRejectsJailedAndSlashed(t *testing.T) {
	k := NewKeeper()
	jailed := adapterAddress(0x33)
	slashed := adapterAddress(0x44)
	registry := fakeRegistry{records: map[string]validatorregistrytypes.ValidatorRecord{
		jailed:		adapterValidator(jailed, "ed25519:validator-jailed", validatorregistrytypes.StatusJailed),
		slashed:	adapterValidator(slashed, "ed25519:validator-slashed", validatorregistrytypes.StatusCandidate),
	}}
	slashedRecord := registry.records[slashed]
	slashedRecord.SlashingHistory = []validatorregistrytypes.SlashingEvent{{Height: 7, FractionBps: 100, Reason: "downtime"}}
	registry.records[slashed] = slashedRecord
	insurance := fakeInsurance{eligible: map[string]bool{jailed: true, slashed: true}}

	_, err := k.ApplyRegisteredValidator(prototype.DefaultAuthority, jailed, 2, registry, insurance)
	require.ErrorContains(t, err, "not eligible")
	_, err = k.ApplyRegisteredValidator(prototype.DefaultAuthority, slashed, 2, registry, insurance)
	require.ErrorContains(t, err, "slashed validator")
}

func TestRegistryActiveValidatorCountPolicyIsEnforcedAtCommit(t *testing.T) {
	k := NewKeeper()
	gs := k.ExportGenesis()
	gs.Params.MaxTotalVotingPower = 1_000_000_000
	gs.Params.ElectionWindowBlocks = 1_000
	gs.Params.WithdrawDeadlineBlocks = 900
	gs.State.ElectionWindow.EndHeight = 1 + gs.Params.ElectionWindowBlocks
	gs.State.ElectionWindow.WithdrawDeadlineHeight = 1 + gs.Params.WithdrawDeadlineBlocks
	require.NoError(t, k.InitGenesis(gs))
	registry := fakeRegistry{records: map[string]validatorregistrytypes.ValidatorRecord{}}
	insurance := fakeInsurance{eligible: map[string]bool{}}
	for i := 0; i < 301; i++ {
		operator := adapterAddressFromIndex(i + 1)
		registry.records[operator] = adapterValidator(operator, fmt.Sprintf("ed25519:validator-%03d", i), validatorregistrytypes.StatusCandidate)
		insurance.eligible[operator] = true
		_, err := k.ApplyRegisteredValidator(prototype.DefaultAuthority, operator, uint64(2+i), registry, insurance)
		require.NoError(t, err)
	}
	_, err := k.CommitElectionWithRegistryPolicy(types.MsgCommitElection{Authority: prototype.DefaultAuthority, Height: 950}, validatorregistrytypes.DefaultParams(), true)
	require.ErrorContains(t, err, "exceeds configured maximum")

	one := NewKeeper()
	operator := adapterAddress(0x55)
	oneRegistry := fakeRegistry{records: map[string]validatorregistrytypes.ValidatorRecord{
		operator: adapterValidator(operator, "ed25519:validator-single", validatorregistrytypes.StatusCandidate),
	}}
	oneInsurance := fakeInsurance{eligible: map[string]bool{operator: true}}
	_, err = one.ApplyRegisteredValidator(prototype.DefaultAuthority, operator, 2, oneRegistry, oneInsurance)
	require.NoError(t, err)
	_, err = one.CommitElectionWithRegistryPolicy(types.MsgCommitElection{Authority: prototype.DefaultAuthority, Height: 90}, validatorregistrytypes.DefaultParams(), false)
	require.ErrorContains(t, err, "below configured minimum")
	_, err = one.CommitElectionWithRegistryPolicy(types.MsgCommitElection{Authority: prototype.DefaultAuthority, Height: 90}, validatorregistrytypes.DefaultParams(), true)
	require.NoError(t, err)
}

func TestElectionExportImportPreservesPreviousCurrentAndNextSet(t *testing.T) {
	source := NewKeeper()
	current := []types.ValidatorPower{{OperatorAddress: adapterAddress(0x11), ConsensusPublicKey: "ed25519:current", VotingPower: 7, ValidatorStatus: validatorregistrytypes.StatusActive}}
	previous := []types.ValidatorPower{{OperatorAddress: adapterAddress(0x22), ConsensusPublicKey: "ed25519:previous", VotingPower: 3, ValidatorStatus: validatorregistrytypes.StatusActive}}
	next := []types.ValidatorPower{{OperatorAddress: adapterAddress(0x33), ConsensusPublicKey: "ed25519:next", VotingPower: 5, ValidatorStatus: validatorregistrytypes.StatusCandidate}}
	gs := source.ExportGenesis()
	gs.State.PreviousValidatorSet = previous
	gs.State.CurrentValidatorSet = current
	gs.State.NextValidatorSet = next
	gs.State = gs.State.Normalize(gs.Params)
	require.NoError(t, source.InitGenesis(gs))

	exported := source.ExportGenesis()
	target := NewKeeper()
	require.NoError(t, target.InitGenesis(exported))
	require.Equal(t, types.SortValidatorSet(previous), target.PreviousValidatorSet())
	require.Equal(t, types.SortValidatorSet(current), target.CurrentValidatorSet())
	require.Equal(t, types.SortValidatorSet(next), target.NextValidatorSet())
}

func TestStakingSyncUpdatesAreDeterministic(t *testing.T) {
	k := NewKeeper()
	gs := k.ExportGenesis()
	removed := adapterAddress(0x44)
	stays := adapterAddress(0x22)
	added := adapterAddress(0x11)
	gs.State.PreviousValidatorSet = []types.ValidatorPower{
		{OperatorAddress: removed, ConsensusPublicKey: "ed25519:removed", VotingPower: 9, ValidatorStatus: validatorregistrytypes.StatusActive},
		{OperatorAddress: stays, ConsensusPublicKey: "ed25519:stays", VotingPower: 4, ValidatorStatus: validatorregistrytypes.StatusActive},
	}
	gs.State.CurrentValidatorSet = []types.ValidatorPower{
		{OperatorAddress: stays, ConsensusPublicKey: "ed25519:stays", VotingPower: 4, ValidatorStatus: validatorregistrytypes.StatusActive},
		{OperatorAddress: added, ConsensusPublicKey: "ed25519:added", VotingPower: 8, ValidatorStatus: validatorregistrytypes.StatusActive},
	}
	gs.State = gs.State.Normalize(gs.Params)
	require.NoError(t, k.InitGenesis(gs))

	syncer := &recordingStakingSyncer{}
	updates, err := k.SyncCurrentSetToStaking(context.Background(), syncer)
	require.NoError(t, err)
	require.Equal(t, []StakingValidatorUpdate{
		{OperatorAddress: removed, ConsensusPublicKey: "ed25519:removed", Remove: true},
		{OperatorAddress: added, ConsensusPublicKey: "ed25519:added", VotingPower: 8},
		{OperatorAddress: stays, ConsensusPublicKey: "ed25519:stays", VotingPower: 4},
	}, updates)
	require.Equal(t, updates, syncer.updates)
}

type fakeRegistry struct {
	records	map[string]validatorregistrytypes.ValidatorRecord
	params	validatorregistrytypes.Params
}

func (f fakeRegistry) Validator(operator string) (validatorregistrytypes.ValidatorRecord, bool, error) {
	record, found := f.records[operator]
	return record, found, nil
}

func (f fakeRegistry) ValidatorParams() validatorregistrytypes.Params {
	if f.params.MinValidatorStake == 0 {
		return validatorregistrytypes.DefaultParams()
	}
	return f.params
}

type fakeInsurance struct {
	eligible map[string]bool
}

func (f fakeInsurance) ValidateValidatorActivation(validatorAddress string) error {
	if f.eligible[validatorAddress] {
		return nil
	}
	return errors.New("validator activation requires minimum insurance")
}

type recordingStakingSyncer struct {
	updates []StakingValidatorUpdate
}

func (r *recordingStakingSyncer) ApplyValidatorSetUpdate(_ context.Context, update StakingValidatorUpdate) error {
	r.updates = append(r.updates, update)
	return nil
}

func adapterValidator(operator, consensusKey, status string) validatorregistrytypes.ValidatorRecord {
	return validatorregistrytypes.ValidatorRecord{
		OperatorAddress:	operator,
		ConsensusPublicKey:	consensusKey,
		TreasuryAddress:	adapterAddress(0x66),
		WithdrawalAddress:	adapterAddress(0x77),
		EmergencyAddress:	adapterAddress(0x88),
		CommissionPolicy:	validatorregistrytypes.DefaultCommissionPolicy(),
		Status:			status,
		SelfBond:		validatorregistrytypes.DefaultMinValidatorStake,
	}
}

func adapterAddress(fill byte) string {
	return addressing.FormatAccAddress(sdk.AccAddress(adapterBytes(fill)))
}

func adapterAddressFromIndex(value int) string {
	out := make([]byte, 20)
	for i := len(out) - 1; i >= 0 && value > 0; i-- {
		out[i] = byte(value)
		value >>= 8
	}
	return addressing.FormatAccAddress(sdk.AccAddress(out))
}

func adapterBytes(fill byte) []byte {
	out := make([]byte, 20)
	for i := range out {
		out[i] = fill
	}
	return out
}
