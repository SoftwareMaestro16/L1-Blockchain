package keeper

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"sort"

	corestore "cosmossdk.io/core/store"

	"github.com/sovereign-l1/l1/x/internal/prototype"
	"github.com/sovereign-l1/l1/x/validator-election/types"
	validatorregistrytypes "github.com/sovereign-l1/l1/x/validator-registry/types"
)

var genesisKey = []byte{0x01}

type GenesisState struct {
	Version uint64
	Params  types.Params
	State   types.State
}

type Keeper struct {
	genesis      GenesisState
	storeService corestore.KVStoreService
}

func NewKeeper() Keeper {
	return Keeper{genesis: DefaultGenesis()}
}

func NewPersistentKeeper(storeService corestore.KVStoreService) Keeper {
	return Keeper{genesis: DefaultGenesis(), storeService: storeService}
}

func DefaultGenesis() GenesisState {
	params := types.DefaultParams()
	return GenesisState{
		Version: prototype.CurrentGenesisVersion,
		Params:  params,
		State:   types.DefaultState(params).Normalize(params),
	}
}

func (gs GenesisState) Validate() error {
	if gs.Version != prototype.CurrentGenesisVersion {
		return errors.New("validator election unsupported genesis version")
	}
	return gs.State.Validate(gs.Params)
}

func (k *Keeper) InitGenesis(gs GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	k.genesis = cloneGenesis(gs)
	return nil
}

func (k *Keeper) InitGenesisState(ctx context.Context, gs GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	k.genesis = cloneGenesis(gs)
	if k.storeService == nil {
		return nil
	}
	bz, err := json.Marshal(cloneGenesis(gs))
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(genesisKey, bz)
}

func (k Keeper) ExportGenesis() GenesisState {
	return cloneGenesis(k.genesis)
}

func (k Keeper) ExportGenesisState(ctx context.Context) (GenesisState, error) {
	if k.storeService == nil {
		return k.ExportGenesis(), nil
	}
	if !reflect.DeepEqual(k.genesis, DefaultGenesis()) {
		return k.ExportGenesis(), nil
	}
	bz, err := k.storeService.OpenKVStore(ctx).Get(genesisKey)
	if err != nil {
		return GenesisState{}, err
	}
	if len(bz) == 0 {
		return DefaultGenesis(), nil
	}
	var gs GenesisState
	if err := json.Unmarshal(bz, &gs); err != nil {
		return GenesisState{}, err
	}
	if err := gs.Validate(); err != nil {
		return GenesisState{}, err
	}
	return cloneGenesis(gs), nil
}

func (k *Keeper) ApplyForValidatorSet(msg types.MsgApplyForValidatorSet) (types.CandidateApplication, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.CandidateApplication{}, err
	}
	if msg.Height == 0 || msg.Height < k.genesis.State.ElectionWindow.StartHeight || msg.Height > k.genesis.State.ElectionWindow.EndHeight {
		return types.CandidateApplication{}, errors.New("validator election application height is outside election window")
	}
	app := msg.Application.Normalize()
	app.AppliedHeight = msg.Height
	app.UpdatedHeight = msg.Height
	if app.ValidatorStatus == validatorregistrytypes.StatusJailed || app.ValidatorStatus == validatorregistrytypes.StatusTombstoned {
		return types.CandidateApplication{}, errors.New("validator election jailed or tombstoned validator cannot apply")
	}
	if err := app.Validate(); err != nil {
		return types.CandidateApplication{}, err
	}
	next := cloneGenesis(k.genesis)
	if existing, found := getApplication(next.State.CandidateApplications, app.OperatorAddress); found && existing.Status == types.ApplicationStatusPending {
		return types.CandidateApplication{}, errors.New("validator election candidate application already pending")
	}
	next.State.CandidateApplications = upsertApplication(next.State.CandidateApplications, app)
	next.State.FrozenStakes = append(next.State.FrozenStakes, types.FrozenStake{
		OperatorAddress: app.OperatorAddress,
		Amount:          app.SelfBond,
		FrozenAtHeight:  msg.Height,
		UnlockHeight:    msg.Height + k.genesis.Params.FrozenStakeUnlockBlocks,
	})
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return types.CandidateApplication{}, err
	}
	k.genesis = next
	return app, nil
}

func (k *Keeper) WithdrawApplication(msg types.MsgWithdrawApplication) (types.CandidateApplication, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.CandidateApplication{}, err
	}
	if msg.Height == 0 || msg.Height > k.genesis.State.ElectionWindow.WithdrawDeadlineHeight {
		return types.CandidateApplication{}, errors.New("validator election application withdrawal deadline has passed")
	}
	return k.transitionApplication(msg.OperatorAddress, func(app types.CandidateApplication) (types.CandidateApplication, error) {
		if app.Status != types.ApplicationStatusPending {
			return types.CandidateApplication{}, errors.New("validator election application is not pending")
		}
		app.Status = types.ApplicationStatusWithdrawn
		app.UpdatedHeight = msg.Height
		return app, nil
	})
}

func (k *Keeper) CommitElection(msg types.MsgCommitElection) (types.ElectionResult, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.ElectionResult{}, err
	}
	if msg.Height == 0 || msg.Height < k.genesis.State.ElectionWindow.WithdrawDeadlineHeight {
		return types.ElectionResult{}, errors.New("validator election cannot commit before withdrawal deadline")
	}
	nextSet := k.computeNextSet()
	result := types.ElectionResult{
		Epoch:     k.genesis.State.ElectionEpoch,
		Height:    msg.Height,
		NextSet:   types.SortValidatorSet(nextSet),
		Committed: true,
	}
	next := cloneGenesis(k.genesis)
	next.State.NextValidatorSet = result.NextSet
	next.State.ElectionResults = upsertResult(next.State.ElectionResults, result)
	next.State.CandidateApplications = markPendingApplicationsCommitted(next.State.CandidateApplications, msg.Height)
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return types.ElectionResult{}, err
	}
	k.genesis = next
	return result, nil
}

func (k *Keeper) FinalizeElection(msg types.MsgFinalizeElection) (types.ValidatorSetTransition, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.ValidatorSetTransition{}, err
	}
	if msg.Height == 0 || msg.Height < k.genesis.State.ElectionWindow.EndHeight {
		return types.ValidatorSetTransition{}, errors.New("validator election cannot finalize before window end")
	}
	if len(k.genesis.State.NextValidatorSet) == 0 && len(k.genesis.State.CandidateApplications) > 0 {
		return types.ValidatorSetTransition{}, errors.New("validator election must be committed before finalization")
	}
	transition := types.ValidatorSetTransition{
		Epoch:       k.genesis.State.ElectionEpoch,
		Height:      msg.Height,
		PreviousSet: types.SortValidatorSet(k.genesis.State.CurrentValidatorSet),
		CurrentSet:  types.SortValidatorSet(k.genesis.State.NextValidatorSet),
		NextSet:     []types.ValidatorPower{},
	}
	next := cloneGenesis(k.genesis)
	next.State.PreviousValidatorSet = transition.PreviousSet
	next.State.CurrentValidatorSet = transition.CurrentSet
	next.State.NextValidatorSet = []types.ValidatorPower{}
	next.State.ElectionResults = finalizeResults(next.State.ElectionResults, next.State.ElectionEpoch)
	next.State.RewardDistributionSnapshots = append(next.State.RewardDistributionSnapshots, types.RewardDistributionSnapshot{
		Epoch:            next.State.ElectionEpoch,
		Height:           msg.Height,
		ValidatorPowers:  transition.CurrentSet,
		TotalVotingPower: totalPower(transition.CurrentSet),
	})
	next.State.TransitionHistory = append(next.State.TransitionHistory, transition)
	next.State.ElectionEpoch++
	next.State.ElectionWindow = types.ElectionWindow{
		StartHeight:            msg.Height + 1,
		EndHeight:              msg.Height + 1 + next.Params.ElectionWindowBlocks,
		WithdrawDeadlineHeight: msg.Height + 1 + next.Params.WithdrawDeadlineBlocks,
	}
	next.State.CandidateApplications = []types.CandidateApplication{}
	next.State.PendingExits = finalizePendingExits(next.State.PendingExits)
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return types.ValidatorSetTransition{}, err
	}
	k.genesis = next
	return transition, nil
}

func (k *Keeper) RequestValidatorExit(msg types.MsgRequestValidatorExit) (types.PendingExit, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.PendingExit{}, err
	}
	if msg.Height == 0 {
		return types.PendingExit{}, errors.New("validator election exit height must be positive")
	}
	if _, _, found := findPendingExit(k.genesis.State.PendingExits, msg.OperatorAddress); found {
		return types.PendingExit{}, errors.New("validator election exit already pending")
	}
	exit := types.PendingExit{OperatorAddress: msg.OperatorAddress, RequestedHeight: msg.Height, Status: types.ExitStatusPending}
	next := cloneGenesis(k.genesis)
	next.State.PendingExits = append(next.State.PendingExits, exit)
	next.State.NextValidatorSet = removeFromSet(next.State.NextValidatorSet, msg.OperatorAddress)
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return types.PendingExit{}, err
	}
	k.genesis = next
	return exit, nil
}

func (k *Keeper) CancelValidatorExit(msg types.MsgCancelValidatorExit) (types.PendingExit, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.PendingExit{}, err
	}
	idx, exit, found := findPendingExit(k.genesis.State.PendingExits, msg.OperatorAddress)
	if !found {
		return types.PendingExit{}, errors.New("validator election pending exit not found")
	}
	if exit.Status != types.ExitStatusPending {
		return types.PendingExit{}, errors.New("validator election exit is not pending")
	}
	exit.Status = types.ExitStatusCancelled
	next := cloneGenesis(k.genesis)
	next.State.PendingExits[idx] = exit
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return types.PendingExit{}, err
	}
	k.genesis = next
	return exit, nil
}

func (k *Keeper) ReleaseFrozenStake(operator string, height uint64) (types.FrozenStake, error) {
	for idx, stake := range k.genesis.State.FrozenStakes {
		if stake.OperatorAddress != operator || stake.Released {
			continue
		}
		if height < stake.UnlockHeight {
			return types.FrozenStake{}, errors.New("validator election frozen stake is still locked")
		}
		stake.Released = true
		next := cloneGenesis(k.genesis)
		next.State.FrozenStakes[idx] = stake
		next.State = next.State.Normalize(next.Params)
		if err := next.Validate(); err != nil {
			return types.FrozenStake{}, err
		}
		k.genesis = next
		return stake, nil
	}
	return types.FrozenStake{}, errors.New("validator election frozen stake not found")
}

func (k Keeper) PreviousValidatorSet() []types.ValidatorPower {
	return types.SortValidatorSet(k.genesis.State.PreviousValidatorSet)
}
func (k Keeper) CurrentValidatorSet() []types.ValidatorPower {
	return types.SortValidatorSet(k.genesis.State.CurrentValidatorSet)
}
func (k Keeper) NextValidatorSet() []types.ValidatorPower {
	return types.SortValidatorSet(k.genesis.State.NextValidatorSet)
}
func (k Keeper) Election() types.State { return cloneGenesis(k.genesis).State }
func (k Keeper) ElectionCandidates() []types.CandidateApplication {
	return types.SortApplications(k.genesis.State.CandidateApplications)
}
func (k Keeper) FrozenStake(operator string) ([]types.FrozenStake, error) {
	out := []types.FrozenStake{}
	for _, stake := range types.SortFrozenStakes(k.genesis.State.FrozenStakes) {
		if stake.OperatorAddress == operator {
			out = append(out, stake)
		}
	}
	return out, nil
}
func (k Keeper) ValidatorSetTransition(epoch uint64) (types.ValidatorSetTransition, bool) {
	for _, transition := range types.SortTransitions(k.genesis.State.TransitionHistory) {
		if transition.Epoch == epoch {
			return transition, true
		}
	}
	return types.ValidatorSetTransition{}, false
}

type Migrator struct{ keeper *Keeper }

func NewMigrator(k *Keeper) Migrator  { return Migrator{keeper: k} }
func (m Migrator) Migrate1to2() error { return m.keeper.ExportGenesis().Validate() }
func (k Keeper) Migrate1to2State(ctx context.Context) error {
	_, err := k.ExportGenesisState(ctx)
	return err
}

func (k Keeper) computeNextSet() []types.ValidatorPower {
	caps := map[string]uint64{}
	for _, cap := range k.genesis.State.ValidatorPowerCaps {
		caps[cap.OperatorAddress] = cap.MaxVotingPower
	}
	exiting := map[string]struct{}{}
	for _, exit := range k.genesis.State.PendingExits {
		if exit.Status == types.ExitStatusPending {
			exiting[exit.OperatorAddress] = struct{}{}
		}
	}
	candidates := []types.CandidateApplication{}
	for _, app := range k.genesis.State.CandidateApplications {
		app = app.Normalize()
		if app.Status != types.ApplicationStatusPending {
			continue
		}
		if app.ValidatorStatus == validatorregistrytypes.StatusJailed || app.ValidatorStatus == validatorregistrytypes.StatusTombstoned {
			continue
		}
		if _, found := exiting[app.OperatorAddress]; found {
			continue
		}
		candidates = append(candidates, app)
	}
	sort.Slice(candidates, func(i, j int) bool {
		left := types.CandidateRank(candidates[i], caps[candidates[i].OperatorAddress], k.genesis.Params)
		right := types.CandidateRank(candidates[j], caps[candidates[j].OperatorAddress], k.genesis.Params)
		if left.VotingPower == right.VotingPower {
			return left.OperatorAddress < right.OperatorAddress
		}
		return left.VotingPower > right.VotingPower
	})
	nextSet := []types.ValidatorPower{}
	total := uint64(0)
	for _, candidate := range candidates {
		power := types.CandidateRank(candidate, caps[candidate.OperatorAddress], k.genesis.Params)
		if power.VotingPower == 0 || total > k.genesis.Params.MaxTotalVotingPower-power.VotingPower {
			continue
		}
		nextSet = append(nextSet, power)
		total += power.VotingPower
		if uint32(len(nextSet)) == k.genesis.Params.MaxValidatorSetSize {
			break
		}
	}
	return nextSet
}

func (k *Keeper) transitionApplication(operator string, mutate func(types.CandidateApplication) (types.CandidateApplication, error)) (types.CandidateApplication, error) {
	idx, app, found := findApplication(k.genesis.State.CandidateApplications, operator)
	if !found {
		return types.CandidateApplication{}, errors.New("validator election candidate application not found")
	}
	updated, err := mutate(app)
	if err != nil {
		return types.CandidateApplication{}, err
	}
	next := cloneGenesis(k.genesis)
	next.State.CandidateApplications[idx] = updated
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return types.CandidateApplication{}, err
	}
	k.genesis = next
	return updated.Normalize(), nil
}

func cloneGenesis(gs GenesisState) GenesisState {
	gs.State = gs.State.Normalize(gs.Params)
	return gs
}

func getApplication(apps []types.CandidateApplication, operator string) (types.CandidateApplication, bool) {
	_, app, found := findApplication(apps, operator)
	return app, found
}

func findApplication(apps []types.CandidateApplication, operator string) (int, types.CandidateApplication, bool) {
	for idx, app := range apps {
		if app.OperatorAddress == operator {
			return idx, app, true
		}
	}
	return -1, types.CandidateApplication{}, false
}

func upsertApplication(apps []types.CandidateApplication, app types.CandidateApplication) []types.CandidateApplication {
	next := append([]types.CandidateApplication(nil), apps...)
	for idx, current := range next {
		if current.OperatorAddress == app.OperatorAddress {
			next[idx] = app
			return types.SortApplications(next)
		}
	}
	return types.SortApplications(append(next, app))
}

func markPendingApplicationsCommitted(apps []types.CandidateApplication, height uint64) []types.CandidateApplication {
	next := append([]types.CandidateApplication(nil), apps...)
	for idx, app := range next {
		if app.Status == types.ApplicationStatusPending {
			app.Status = types.ApplicationStatusCommitted
			app.UpdatedHeight = height
			next[idx] = app
		}
	}
	return types.SortApplications(next)
}

func upsertResult(results []types.ElectionResult, result types.ElectionResult) []types.ElectionResult {
	next := append([]types.ElectionResult(nil), results...)
	for idx, current := range next {
		if current.Epoch == result.Epoch {
			next[idx] = result
			return types.SortResults(next)
		}
	}
	return types.SortResults(append(next, result))
}

func finalizeResults(results []types.ElectionResult, epoch uint64) []types.ElectionResult {
	next := append([]types.ElectionResult(nil), results...)
	for idx, result := range next {
		if result.Epoch == epoch {
			result.Finalized = true
			next[idx] = result
		}
	}
	return types.SortResults(next)
}

func finalizePendingExits(exits []types.PendingExit) []types.PendingExit {
	next := append([]types.PendingExit(nil), exits...)
	for idx, exit := range next {
		if exit.Status == types.ExitStatusPending {
			exit.Status = types.ExitStatusFinalized
			next[idx] = exit
		}
	}
	return types.SortPendingExits(next)
}

func findPendingExit(exits []types.PendingExit, operator string) (int, types.PendingExit, bool) {
	for idx, exit := range exits {
		if exit.OperatorAddress == operator {
			return idx, exit, true
		}
	}
	return -1, types.PendingExit{}, false
}

func removeFromSet(set []types.ValidatorPower, operator string) []types.ValidatorPower {
	next := []types.ValidatorPower{}
	for _, value := range set {
		if value.OperatorAddress != operator {
			next = append(next, value)
		}
	}
	return types.SortValidatorSet(next)
}

func totalPower(set []types.ValidatorPower) uint64 {
	total := uint64(0)
	for _, value := range set {
		total += value.VotingPower
	}
	return total
}
