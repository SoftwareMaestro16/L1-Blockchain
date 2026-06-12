package keeper

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"

	corestore "cosmossdk.io/core/store"

	"github.com/sovereign-l1/l1/x/internal/prototype"
	"github.com/sovereign-l1/l1/x/reporter/types"
)

var genesisKey = []byte{0x01}

type GenesisState struct {
	Version	uint64
	Params	types.Params
	State	types.State
}

type Keeper struct {
	genesis		GenesisState
	storeService	corestore.KVStoreService
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
		Version:	prototype.CurrentGenesisVersion,
		Params:		params,
		State:		types.State{}.Normalize(params),
	}
}

func (gs GenesisState) Validate() error {
	if gs.Version != prototype.CurrentGenesisVersion {
		return errors.New("reporter unsupported genesis version")
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

func (k *Keeper) RegisterReporter(msg types.MsgRegisterReporter) (types.ReporterRecord, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.ReporterRecord{}, err
	}
	if msg.Height == 0 {
		return types.ReporterRecord{}, errors.New("reporter registration height must be positive")
	}
	if _, _, found := findReporter(k.genesis.State.Reporters, msg.ReporterAddress); found {
		return types.ReporterRecord{}, errors.New("reporter already registered")
	}
	reporter := types.ReporterRecord{
		ReporterAddress:	msg.ReporterAddress,
		ReporterScore:		k.genesis.Params.InitialScore,
		Status:			types.StatusActive,
	}
	if err := reporter.Validate(k.genesis.Params); err != nil {
		return types.ReporterRecord{}, err
	}
	next := cloneGenesis(k.genesis)
	next.State.Reporters = append(next.State.Reporters, reporter)
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return types.ReporterRecord{}, err
	}
	k.genesis = next
	return reporter, nil
}

func (k *Keeper) BondReporter(msg types.MsgBondReporter) (types.ReporterRecord, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.ReporterRecord{}, err
	}
	if msg.Height == 0 {
		return types.ReporterRecord{}, errors.New("reporter bond height must be positive")
	}
	if msg.Amount == 0 {
		return types.ReporterRecord{}, errors.New("reporter bond amount must be positive")
	}
	idx, reporter, found := findReporter(k.genesis.State.Reporters, msg.ReporterAddress)
	if !found {
		return types.ReporterRecord{}, errors.New("reporter not registered")
	}
	if reporter.Status == types.StatusJailed {
		return types.ReporterRecord{}, errors.New("jailed reporter cannot bond")
	}
	reporter.BondedAmount += msg.Amount
	reporter.Status = types.StatusActive
	reporter.UnbondingStartHeight = 0
	reporter.UnbondingCompleteHeight = 0
	next := cloneGenesis(k.genesis)
	next.State.Reporters[idx] = reporter
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return types.ReporterRecord{}, err
	}
	k.genesis = next
	return reporter, nil
}

func (k *Keeper) UnbondReporter(msg types.MsgUnbondReporter) (types.ReporterRecord, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.ReporterRecord{}, err
	}
	if msg.Height == 0 {
		return types.ReporterRecord{}, errors.New("reporter unbond height must be positive")
	}
	idx, reporter, found := findReporter(k.genesis.State.Reporters, msg.ReporterAddress)
	if !found {
		return types.ReporterRecord{}, errors.New("reporter not registered")
	}
	if reporter.Status == types.StatusJailed {
		return types.ReporterRecord{}, errors.New("jailed reporter cannot unbond")
	}
	if reporter.Status != types.StatusUnbonding {
		reporter.Status = types.StatusUnbonding
		reporter.UnbondingStartHeight = msg.Height
		reporter.UnbondingCompleteHeight = msg.Height + k.genesis.Params.ChallengePeriodBlocks
	} else {
		if msg.Height < reporter.UnbondingCompleteHeight {
			return types.ReporterRecord{}, errors.New("reporter unbonding challenge period has not elapsed")
		}
		reporter.Status = types.StatusActive
		reporter.BondedAmount = 0
		reporter.UnbondingStartHeight = 0
		reporter.UnbondingCompleteHeight = 0
	}
	next := cloneGenesis(k.genesis)
	next.State.Reporters[idx] = reporter
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return types.ReporterRecord{}, err
	}
	k.genesis = next
	return reporter, nil
}

func (k *Keeper) SubmitReport(msg types.MsgSubmitReport) (types.ReportRecord, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.ReportRecord{}, err
	}
	if msg.Height == 0 {
		return types.ReportRecord{}, errors.New("report submission height must be positive")
	}
	if msg.Accepted && msg.Malicious {
		return types.ReportRecord{}, errors.New("report cannot be both accepted and malicious")
	}
	reporterIdx, reporter, found := findReporter(k.genesis.State.Reporters, msg.ReporterAddress)
	if !found {
		return types.ReportRecord{}, errors.New("reporter not registered")
	}
	if reporter.Status != types.StatusActive {
		return types.ReportRecord{}, errors.New("reporter must be active to submit slashable reports")
	}
	if reporter.BondedAmount < k.genesis.Params.MinBondAmount {
		return types.ReportRecord{}, errors.New("reporter must be bonded to submit slashable reports")
	}
	if _, _, found := findReport(k.genesis.State.Reports, msg.ReportID); found {
		return types.ReportRecord{}, errors.New("report already submitted")
	}
	status := types.ReportStatusPending
	finalizedHeight := uint64(0)
	rewardAmount := uint64(0)
	slashAmount := uint64(0)
	if msg.Accepted {
		status = types.ReportStatusAccepted
		finalizedHeight = msg.Height
		rewardAmount = msg.RewardAmount
		if rewardAmount == 0 {
			rewardAmount = k.genesis.Params.MaxRewardAmount
		}
		reporter.AcceptedReports++
		reporter.ReporterScore += k.genesis.Params.AcceptedScoreDelta
		reporter.RewardHistory = append(reporter.RewardHistory, types.ReporterReward{
			ReportID:	msg.ReportID,
			Amount:		rewardAmount,
			CreatedAt:	msg.Height,
		})
	} else if msg.Malicious {
		status = types.ReportStatusMalicious
		finalizedHeight = msg.Height
		slashAmount = types.SlashAmount(reporter.BondedAmount, k.genesis.Params.MaliciousSlashBps)
		reporter.RejectedReports++
		reporter.SlashedReporterBond += slashAmount
		reporter.BondedAmount -= slashAmount
		reporter.ReporterScore -= k.genesis.Params.RejectedScoreDelta
		if reporter.BondedAmount < k.genesis.Params.MinBondAmount {
			reporter.Status = types.StatusJailed
		}
	}
	report := types.ReportRecord{
		ReportID:		msg.ReportID,
		ReporterAddress:	msg.ReporterAddress,
		ReportType:		msg.ReportType,
		Subject:		msg.Subject,
		PayloadHash:		msg.PayloadHash,
		PayloadSizeBytes:	msg.PayloadSizeBytes,
		Status:			status,
		SubmittedHeight:	msg.Height,
		FinalizedHeight:	finalizedHeight,
		RewardAmount:		rewardAmount,
		SlashAmount:		slashAmount,
	}
	if err := report.Validate(k.genesis.Params); err != nil {
		return types.ReportRecord{}, err
	}
	if err := reporter.Validate(k.genesis.Params); err != nil {
		return types.ReportRecord{}, err
	}
	next := cloneGenesis(k.genesis)
	next.State.Reporters[reporterIdx] = reporter
	next.State.Reports = append(next.State.Reports, report)
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return types.ReportRecord{}, err
	}
	k.genesis = next
	return report, nil
}

func (k *Keeper) ClaimReporterReward(msg types.MsgClaimReporterReward) (types.ReporterReward, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.ReporterReward{}, err
	}
	if msg.Height == 0 {
		return types.ReporterReward{}, errors.New("reporter reward claim height must be positive")
	}
	reportIdx, report, found := findReport(k.genesis.State.Reports, msg.ReportID)
	if !found {
		return types.ReporterReward{}, errors.New("report not found")
	}
	if report.ReporterAddress != msg.ReporterAddress {
		return types.ReporterReward{}, errors.New("reporter reward claim address mismatch")
	}
	if report.Status != types.ReportStatusAccepted {
		return types.ReporterReward{}, errors.New("only accepted report rewards can be claimed")
	}
	if report.RewardClaimed {
		return types.ReporterReward{}, errors.New("reporter reward already claimed")
	}
	reporterIdx, reporter, found := findReporter(k.genesis.State.Reporters, msg.ReporterAddress)
	if !found {
		return types.ReporterReward{}, errors.New("reporter not registered")
	}
	rewardIdx, reward, found := findReward(reporter.RewardHistory, msg.ReportID)
	if !found {
		return types.ReporterReward{}, errors.New("reporter reward not found")
	}
	if reward.Claimed {
		return types.ReporterReward{}, errors.New("reporter reward already claimed")
	}
	reward.Claimed = true
	reward.ClaimedAt = msg.Height
	report.RewardClaimed = true
	next := cloneGenesis(k.genesis)
	next.State.Reporters[reporterIdx].RewardHistory[rewardIdx] = reward
	next.State.Reports[reportIdx] = report
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return types.ReporterReward{}, err
	}
	k.genesis = next
	return reward, nil
}

func (k Keeper) Reporter(address string) (types.ReporterRecord, bool) {
	_, reporter, found := findReporter(k.genesis.State.Reporters, address)
	return reporter, found
}

func (k Keeper) Reporters() []types.ReporterRecord {
	return types.SortReporters(k.genesis.State.Reporters)
}

func (k Keeper) ReporterReports(address string) []types.ReportRecord {
	out := []types.ReportRecord{}
	for _, report := range types.SortReports(k.genesis.State.Reports) {
		if report.ReporterAddress == address {
			out = append(out, report)
		}
	}
	return out
}

func (k Keeper) ReporterRewards(address string) []types.ReporterReward {
	_, reporter, found := findReporter(k.genesis.State.Reporters, address)
	if !found {
		return []types.ReporterReward{}
	}
	return types.SortRewards(reporter.RewardHistory)
}

type Migrator struct{ keeper *Keeper }

func NewMigrator(k *Keeper) Migrator	{ return Migrator{keeper: k} }
func (m Migrator) Migrate1to2() error	{ return m.keeper.ExportGenesis().Validate() }
func (k Keeper) Migrate1to2State(ctx context.Context) error {
	_, err := k.ExportGenesisState(ctx)
	return err
}

func cloneGenesis(gs GenesisState) GenesisState {
	gs.State = gs.State.Normalize(gs.Params)
	return gs
}

func findReporter(reporters []types.ReporterRecord, address string) (int, types.ReporterRecord, bool) {
	for idx, reporter := range reporters {
		if reporter.ReporterAddress == address {
			return idx, reporter, true
		}
	}
	return -1, types.ReporterRecord{}, false
}

func findReport(reports []types.ReportRecord, reportID string) (int, types.ReportRecord, bool) {
	for idx, report := range reports {
		if report.ReportID == reportID {
			return idx, report, true
		}
	}
	return -1, types.ReportRecord{}, false
}

func findReward(rewards []types.ReporterReward, reportID string) (int, types.ReporterReward, bool) {
	for idx, reward := range rewards {
		if reward.ReportID == reportID {
			return idx, reward, true
		}
	}
	return -1, types.ReporterReward{}, false
}
