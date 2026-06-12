package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	corestore "cosmossdk.io/core/store"

	"github.com/sovereign-l1/l1/x/evidence/types"
	"github.com/sovereign-l1/l1/x/internal/prototype"
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
	hooks		SlashingIntegrationHooks
}

type SlashingIntegrationHooks interface {
	RecordSlashingEvent(event types.SlashEvent) error
	JailValidator(validator string, tombstone bool, jailUntilHeight uint64, evidenceID string) error
	UnjailValidator(validator string, height uint64) error
	FreezeStake(validator string, evidenceID string, amount uint64, releaseHeight uint64) error
	ApplyReputationPenalty(validator string, evidenceID string, evidenceType string, height uint64) error
	ApplyInsuranceSlash(validator string, evidenceID string, amount uint64, height uint64) error
}

func NewKeeper() Keeper {
	return Keeper{genesis: DefaultGenesis()}
}

func NewPersistentKeeper(storeService corestore.KVStoreService) Keeper {
	return Keeper{genesis: DefaultGenesis(), storeService: storeService}
}

func (k *Keeper) SetSlashingIntegrationHooks(hooks SlashingIntegrationHooks) {
	k.hooks = hooks
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
		return errors.New("native evidence unsupported genesis version")
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

func (k *Keeper) SubmitEvidence(msg types.MsgSubmitEvidence) (types.EvidenceRecord, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.EvidenceRecord{}, err
	}
	if msg.Height == 0 {
		return types.EvidenceRecord{}, errors.New("native evidence submission height must be positive")
	}
	record := types.EvidenceRecord{
		EvidenceID:		msg.EvidenceID,
		Status:			types.StatusPending,
		EvidenceType:		msg.EvidenceType,
		AccusedValidator:	msg.AccusedValidator,
		Reporter:		msg.Reporter,
		ProofPayloadHash:	msg.ProofPayloadHash,
		PayloadSizeBytes:	msg.PayloadSizeBytes,
		SlashDecision: types.SlashDecision{
			FractionBps:	types.CanonicalSlashFraction(k.genesis.Params, msg.EvidenceType, msg.SlashFractionBps),
			Tombstone:	types.IsCriticalEvidenceType(msg.EvidenceType),
		},
		RewardDecision: types.RewardDecision{
			Reporter:	msg.Reporter,
			AmountNaet:	msg.RewardNaet,
		},
		SubmittedHeight:	msg.Height,
		UpdatedHeight:		msg.Height,
		ExpirationHeight:	msg.Height + k.genesis.Params.EvidenceTTLBlocks,
		RequiresReview:		msg.RequiresReview,
	}
	if record.RewardDecision.AmountNaet == 0 {
		record.RewardDecision.AmountNaet = k.genesis.Params.MaxReporterRewardNaet
	}
	if err := record.Validate(k.genesis.Params); err != nil {
		return types.EvidenceRecord{}, err
	}
	next := cloneGenesis(k.genesis)
	if _, _, found := findEvidence(next.State.Evidence, record.EvidenceID); found {
		return types.EvidenceRecord{}, errors.New("native evidence duplicate evidence id")
	}
	if _, found := findEvidenceByHash(next.State.Evidence, record.ProofPayloadHash); found {
		return types.EvidenceRecord{}, errors.New("native evidence duplicate proof payload hash")
	}
	next.State.Evidence = append(next.State.Evidence, record)
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return types.EvidenceRecord{}, err
	}
	k.genesis = next
	return record, nil
}

func (k *Keeper) VoteEvidence(msg types.MsgVoteEvidence) (types.EvidenceRecord, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.EvidenceRecord{}, err
	}
	idx, record, found := findEvidence(k.genesis.State.Evidence, msg.EvidenceID)
	if !found {
		return types.EvidenceRecord{}, errors.New("native evidence record not found")
	}
	if record.Status != types.StatusPending {
		return types.EvidenceRecord{}, errors.New("native evidence vote requires pending evidence")
	}
	if msg.Height == 0 || msg.Height > record.ExpirationHeight {
		return types.EvidenceRecord{}, errors.New("native evidence vote height is outside active evidence window")
	}
	support := types.VoteSupportReject
	if msg.Accept {
		support = types.VoteSupportAccept
	}
	vote := types.EvidenceVote{
		Voter:		msg.Voter,
		Support:	support,
		VotingPowerBps:	msg.VotingPowerBps,
		Height:		msg.Height,
	}
	if err := vote.Validate(); err != nil {
		return types.EvidenceRecord{}, err
	}
	for _, existing := range record.Votes {
		if existing.Voter == vote.Voter {
			return types.EvidenceRecord{}, errors.New("native evidence duplicate vote")
		}
	}
	record.Votes = append(record.Votes, vote)
	record.UpdatedHeight = msg.Height
	next := cloneGenesis(k.genesis)
	next.State.Evidence[idx] = record
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return types.EvidenceRecord{}, err
	}
	k.genesis = next
	return record, nil
}

func (k *Keeper) FinalizeEvidence(msg types.MsgFinalizeEvidence) (types.EvidenceRecord, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.EvidenceRecord{}, err
	}
	idx, record, found := findEvidence(k.genesis.State.Evidence, msg.EvidenceID)
	if !found {
		return types.EvidenceRecord{}, errors.New("native evidence record not found")
	}
	if record.Status != types.StatusPending {
		return types.EvidenceRecord{}, errors.New("native evidence can only be finalized once")
	}
	if msg.Height == 0 {
		return types.EvidenceRecord{}, errors.New("native evidence finalization height must be positive")
	}
	if msg.Height > record.ExpirationHeight {
		record.Status = types.StatusExpired
		record.UpdatedHeight = msg.Height
		record.FinalizedHeight = msg.Height
		next := cloneGenesis(k.genesis)
		next.State.Evidence[idx] = record
		next.State = next.State.Normalize(next.Params)
		if err := next.Validate(); err != nil {
			return types.EvidenceRecord{}, err
		}
		k.genesis = next
		return record, nil
	}
	if record.RequiresReview && types.AcceptedVotingPowerBps(record.Votes) < k.genesis.Params.ReviewQuorumBps {
		record.Status = types.StatusRejected
		record.RejectionReason = "review quorum not reached"
		record.UpdatedHeight = msg.Height
		record.FinalizedHeight = msg.Height
		next := cloneGenesis(k.genesis)
		next.State.Evidence[idx] = record
		next.State = next.State.Normalize(next.Params)
		if err := next.Validate(); err != nil {
			return types.EvidenceRecord{}, err
		}
		k.genesis = next
		return record, nil
	}
	record.Status = types.StatusAccepted
	record.SlashDecision.Applied = true
	record.RewardDecision.Paid = true
	record.UpdatedHeight = msg.Height
	record.FinalizedHeight = msg.Height

	next := cloneGenesis(k.genesis)
	next.State.Evidence[idx] = record
	next.State.SlashEvents = append(next.State.SlashEvents, types.SlashEvent{
		EvidenceID:		record.EvidenceID,
		ValidatorAddress:	record.AccusedValidator,
		FractionBps:		record.SlashDecision.FractionBps,
		Tombstone:		record.SlashDecision.Tombstone,
		Height:			msg.Height,
	})
	next.State.ReporterRewards = append(next.State.ReporterRewards, types.ReporterReward{
		EvidenceID:	record.EvidenceID,
		Reporter:	record.Reporter,
		AmountNaet:	record.RewardDecision.AmountNaet,
		Paid:		true,
		Height:		msg.Height,
	})
	status := types.RegistryStatusJailed
	if record.SlashDecision.Tombstone {
		status = types.RegistryStatusTombstoned
		next.State.TombstonedValidators = append(next.State.TombstonedValidators, record.AccusedValidator)
	}
	next.State.RegistryUpdates = append(next.State.RegistryUpdates, types.RegistryUpdate{
		EvidenceID:		record.EvidenceID,
		ValidatorAddress:	record.AccusedValidator,
		Status:			status,
		Height:			msg.Height,
	})
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return types.EvidenceRecord{}, err
	}
	k.genesis = next
	return record, nil
}

func (k *Keeper) CancelExpiredEvidence(msg types.MsgCancelExpiredEvidence) (types.EvidenceRecord, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.EvidenceRecord{}, err
	}
	idx, record, found := findEvidence(k.genesis.State.Evidence, msg.EvidenceID)
	if !found {
		return types.EvidenceRecord{}, errors.New("native evidence record not found")
	}
	if record.Status != types.StatusPending {
		return types.EvidenceRecord{}, errors.New("native evidence cancel requires pending evidence")
	}
	if msg.Height == 0 || msg.Height <= record.ExpirationHeight {
		return types.EvidenceRecord{}, errors.New("native evidence has not expired")
	}
	record.Status = types.StatusExpired
	record.UpdatedHeight = msg.Height
	record.FinalizedHeight = msg.Height
	next := cloneGenesis(k.genesis)
	next.State.Evidence[idx] = record
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return types.EvidenceRecord{}, err
	}
	k.genesis = next
	return record, nil
}

func (k *Keeper) ProcessDoubleSignEvidence(msg types.MsgSubmitDoubleSignEvidence) (types.EvidenceRecord, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.EvidenceRecord{}, err
	}
	if err := validateDoubleSignEvidence(msg); err != nil {
		return types.EvidenceRecord{}, err
	}
	if containsString(k.genesis.State.TombstonedValidators, strings.TrimSpace(msg.AccusedValidator)) {
		return types.EvidenceRecord{}, errors.New("native evidence validator already tombstoned")
	}
	proofHash := proofHashParts("double-sign", msg.AccusedValidator, msg.VoteAHash, msg.VoteBHash, fmt.Sprint(msg.InfractionHeight))
	record := acceptedEvidenceRecord(k.genesis.Params, msg.EvidenceID, types.EvidenceTypeDoubleSign, msg.AccusedValidator, msg.Reporter, proofHash, msg.Height, k.genesis.Params.CriticalFaultSlashFractionBps, true)
	return k.acceptObjectiveEvidence(record, slashingPipeline{
		reason:		types.SlashingReasonDoubleSign,
		validatorStake:	msg.ValidatorStake,
		jailBlocks:	k.genesis.Params.DoubleSignJailBlocks,
		tombstone:	true,
		height:		msg.Height,
	})
}

func (k *Keeper) ProcessDowntimeEvidence(msg types.MsgSubmitDowntimeEvidence) (types.EvidenceRecord, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.EvidenceRecord{}, err
	}
	if err := validateDowntimeEvidence(msg); err != nil {
		return types.EvidenceRecord{}, err
	}
	if containsString(k.genesis.State.TombstonedValidators, strings.TrimSpace(msg.AccusedValidator)) {
		return types.EvidenceRecord{}, errors.New("native evidence validator already tombstoned")
	}
	proofHash := proofHashParts("downtime", msg.AccusedValidator, fmt.Sprint(msg.MissedBlocks), fmt.Sprint(msg.WindowBlocks), fmt.Sprint(msg.InfractionHeight))
	offenseCount := currentOffenseCount(k.genesis.State.OffenseCounters, msg.AccusedValidator, types.EvidenceTypeDowntime) + 1
	fraction := types.DowntimeSlashFraction(k.genesis.Params, offenseCount)
	record := acceptedEvidenceRecord(k.genesis.Params, msg.EvidenceID, types.EvidenceTypeDowntime, msg.AccusedValidator, msg.Reporter, proofHash, msg.Height, fraction, false)
	return k.acceptObjectiveEvidence(record, slashingPipeline{
		reason:		types.SlashingReasonDowntime,
		validatorStake:	msg.ValidatorStake,
		jailBlocks:	types.DowntimeJailBlocks(k.genesis.Params, offenseCount),
		tombstone:	false,
		height:		msg.Height,
	})
}

func (k *Keeper) UnjailValidator(msg types.MsgUnjailValidator) error {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return err
	}
	if msg.Height == 0 {
		return errors.New("native evidence unjail height must be positive")
	}
	validator := strings.TrimSpace(msg.ValidatorAddress)
	if validator == "" {
		return errors.New("native evidence unjail validator is required")
	}
	idx, jail, found := findActiveJail(k.genesis.State.JailRecords, validator)
	if !found {
		return errors.New("native evidence validator is not jailed")
	}
	if jail.Tombstone {
		return errors.New("native evidence tombstoned validator cannot be unjailed")
	}
	if msg.Height < jail.JailedUntilHeight {
		return errors.New("native evidence validator cannot unjail before jail period ends")
	}
	jail.Active = false
	next := cloneGenesis(k.genesis)
	next.State.JailRecords[idx] = jail
	next.State.RegistryUpdates = append(next.State.RegistryUpdates, types.RegistryUpdate{
		EvidenceID:		jail.EvidenceID,
		ValidatorAddress:	validator,
		Status:			types.RegistryStatusCandidate,
		Height:			msg.Height,
	})
	next.State.IntegrationEvents = append(next.State.IntegrationEvents, integrationEvent(jail.EvidenceID, validator, "validator-registry", "unjail", msg.Height))
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return err
	}
	if k.hooks != nil {
		if err := k.hooks.UnjailValidator(validator, msg.Height); err != nil {
			return err
		}
	}
	k.genesis = next
	return nil
}

func (k Keeper) JailRecords() []types.JailRecord {
	return types.SortJailRecords(k.genesis.State.JailRecords)
}

func (k Keeper) FrozenStakes() []types.FrozenStake {
	return types.SortFrozenStakes(k.genesis.State.FrozenStakes)
}

func (k Keeper) OffenseCounters() []types.OffenseCounter {
	return types.SortOffenseCounters(k.genesis.State.OffenseCounters)
}

func (k Keeper) IntegrationEvents() []types.IntegrationEvent {
	return types.SortIntegrationEvents(k.genesis.State.IntegrationEvents)
}

func (k Keeper) ValidateActiveSetInvariant(activeValidators []string) error {
	return types.ValidateNoJailedValidatorInActiveSet(k.genesis.State, activeValidators)
}

func (k Keeper) Evidence(evidenceID string) (types.EvidenceRecord, bool) {
	_, record, found := findEvidence(k.genesis.State.Evidence, evidenceID)
	return record, found
}

func (k Keeper) EvidenceByValidator(validator string) []types.EvidenceRecord {
	out := []types.EvidenceRecord{}
	for _, record := range types.SortEvidence(k.genesis.State.Evidence) {
		if record.AccusedValidator == validator {
			out = append(out, record)
		}
	}
	return out
}

func (k Keeper) EvidenceByReporter(reporter string) []types.EvidenceRecord {
	out := []types.EvidenceRecord{}
	for _, record := range types.SortEvidence(k.genesis.State.Evidence) {
		if record.Reporter == reporter {
			out = append(out, record)
		}
	}
	return out
}

func (k Keeper) PendingEvidence() []types.EvidenceRecord {
	out := []types.EvidenceRecord{}
	for _, record := range types.SortEvidence(k.genesis.State.Evidence) {
		if record.Status == types.StatusPending {
			out = append(out, record)
		}
	}
	return out
}

func (k Keeper) EvidenceParams() types.Params {
	return k.genesis.Params
}

func (k Keeper) SlashEvents() []types.SlashEvent {
	return types.SortSlashEvents(k.genesis.State.SlashEvents)
}

func (k Keeper) ReporterRewards() []types.ReporterReward {
	return types.SortReporterRewards(k.genesis.State.ReporterRewards)
}

func (k Keeper) RegistryUpdates() []types.RegistryUpdate {
	return types.SortRegistryUpdates(k.genesis.State.RegistryUpdates)
}

func (k Keeper) TombstonedValidators() []string {
	return append([]string(nil), k.genesis.State.TombstonedValidators...)
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

func findEvidence(records []types.EvidenceRecord, evidenceID string) (int, types.EvidenceRecord, bool) {
	for idx, record := range records {
		if record.EvidenceID == evidenceID {
			return idx, record, true
		}
	}
	return -1, types.EvidenceRecord{}, false
}

func findEvidenceByHash(records []types.EvidenceRecord, hash string) (types.EvidenceRecord, bool) {
	for _, record := range records {
		if record.ProofPayloadHash == hash {
			return record, true
		}
	}
	return types.EvidenceRecord{}, false
}

type slashingPipeline struct {
	reason		string
	validatorStake	uint64
	jailBlocks	uint64
	tombstone	bool
	height		uint64
}

func (k *Keeper) acceptObjectiveEvidence(record types.EvidenceRecord, pipeline slashingPipeline) (types.EvidenceRecord, error) {
	record.AccusedValidator = strings.TrimSpace(record.AccusedValidator)
	record.Reporter = strings.TrimSpace(record.Reporter)
	if pipeline.height == 0 {
		return types.EvidenceRecord{}, errors.New("native evidence objective finalization height must be positive")
	}
	if pipeline.validatorStake == 0 {
		return types.EvidenceRecord{}, errors.New("native evidence validator stake must be positive for slashing")
	}
	if record.FinalizedHeight != pipeline.height {
		return types.EvidenceRecord{}, errors.New("native evidence finalized height mismatch")
	}
	if err := record.Validate(k.genesis.Params); err != nil {
		return types.EvidenceRecord{}, err
	}
	next := cloneGenesis(k.genesis)
	if _, _, found := findEvidence(next.State.Evidence, record.EvidenceID); found {
		return types.EvidenceRecord{}, errors.New("native evidence duplicate evidence id")
	}
	if _, found := findEvidenceByHash(next.State.Evidence, record.ProofPayloadHash); found {
		return types.EvidenceRecord{}, errors.New("native evidence duplicate proof payload hash")
	}
	offenseCount := uint64(0)
	next.State.OffenseCounters, offenseCount = upsertOffenseCounter(next.State.OffenseCounters, record.AccusedValidator, record.EvidenceType, pipeline.height)
	slashAmount := slashAmount(pipeline.validatorStake, record.SlashDecision.FractionBps)
	releaseHeight := uint64(0)
	if !pipeline.tombstone {
		releaseHeight = pipeline.height + k.genesis.Params.FrozenStakeBlocks
	}
	jailUntil := uint64(0)
	if !pipeline.tombstone {
		jailUntil = pipeline.height + pipeline.jailBlocks
	}
	record.SlashDecision.Tombstone = pipeline.tombstone
	record.SlashDecision.Applied = true
	record.RewardDecision.Paid = true

	slashEvent := types.SlashEvent{
		EvidenceID:		record.EvidenceID,
		ValidatorAddress:	record.AccusedValidator,
		FractionBps:		record.SlashDecision.FractionBps,
		Tombstone:		pipeline.tombstone,
		Height:			pipeline.height,
		Reason:			pipeline.reason,
		OffenseCount:		offenseCount,
		FrozenStake:		slashAmount,
		JailUntilHeight:	jailUntil,
	}
	status := types.RegistryStatusJailed
	if pipeline.tombstone {
		status = types.RegistryStatusTombstoned
		next.State.TombstonedValidators = append(next.State.TombstonedValidators, record.AccusedValidator)
	}
	next.State.Evidence = append(next.State.Evidence, record)
	next.State.SlashEvents = append(next.State.SlashEvents, slashEvent)
	next.State.ReporterRewards = append(next.State.ReporterRewards, types.ReporterReward{
		EvidenceID:	record.EvidenceID,
		Reporter:	record.Reporter,
		AmountNaet:	record.RewardDecision.AmountNaet,
		Paid:		true,
		Height:		pipeline.height,
	})
	next.State.RegistryUpdates = append(next.State.RegistryUpdates, types.RegistryUpdate{
		EvidenceID:		record.EvidenceID,
		ValidatorAddress:	record.AccusedValidator,
		Status:			status,
		Height:			pipeline.height,
	})
	next.State.JailRecords = append(next.State.JailRecords, types.JailRecord{
		EvidenceID:		record.EvidenceID,
		ValidatorAddress:	record.AccusedValidator,
		Reason:			pipeline.reason,
		JailedAtHeight:		pipeline.height,
		JailedUntilHeight:	jailUntil,
		Tombstone:		pipeline.tombstone,
		Active:			true,
	})
	next.State.FrozenStakes = append(next.State.FrozenStakes, types.FrozenStake{
		EvidenceID:		record.EvidenceID,
		ValidatorAddress:	record.AccusedValidator,
		Amount:			slashAmount,
		FrozenAtHeight:		pipeline.height,
		ReleaseHeight:		releaseHeight,
		Reason:			pipeline.reason,
	})
	next.State.IntegrationEvents = append(next.State.IntegrationEvents,
		integrationEvent(record.EvidenceID, record.AccusedValidator, "validator-registry", status, pipeline.height),
		integrationEvent(record.EvidenceID, record.AccusedValidator, "reputation", "slash-penalty", pipeline.height),
		integrationEvent(record.EvidenceID, record.AccusedValidator, "validator-insurance", "slash-claim", pipeline.height),
	)
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return types.EvidenceRecord{}, err
	}
	if err := k.runSlashingHooks(slashEvent, releaseHeight, status, record.EvidenceType); err != nil {
		return types.EvidenceRecord{}, err
	}
	k.genesis = next
	return record, nil
}

func (k *Keeper) runSlashingHooks(event types.SlashEvent, releaseHeight uint64, status string, evidenceType string) error {
	if k.hooks == nil {
		return nil
	}
	if err := k.hooks.RecordSlashingEvent(event); err != nil {
		return err
	}
	if err := k.hooks.JailValidator(event.ValidatorAddress, status == types.RegistryStatusTombstoned, event.JailUntilHeight, event.EvidenceID); err != nil {
		return err
	}
	if err := k.hooks.FreezeStake(event.ValidatorAddress, event.EvidenceID, event.FrozenStake, releaseHeight); err != nil {
		return err
	}
	if err := k.hooks.ApplyReputationPenalty(event.ValidatorAddress, event.EvidenceID, evidenceType, event.Height); err != nil {
		return err
	}
	if err := k.hooks.ApplyInsuranceSlash(event.ValidatorAddress, event.EvidenceID, event.FrozenStake, event.Height); err != nil {
		return err
	}
	return nil
}

func acceptedEvidenceRecord(params types.Params, id string, evidenceType string, validator string, reporter string, proofHash string, height uint64, fractionBps uint32, tombstone bool) types.EvidenceRecord {
	reward := params.MaxReporterRewardNaet
	return types.EvidenceRecord{
		EvidenceID:		strings.TrimSpace(id),
		Status:			types.StatusAccepted,
		EvidenceType:		evidenceType,
		AccusedValidator:	strings.TrimSpace(validator),
		Reporter:		strings.TrimSpace(reporter),
		ProofPayloadHash:	proofHash,
		PayloadSizeBytes:	uint32(len(proofHash)),
		SlashDecision: types.SlashDecision{
			FractionBps:	fractionBps,
			Tombstone:	tombstone,
			Applied:	true,
		},
		RewardDecision: types.RewardDecision{
			Reporter:	strings.TrimSpace(reporter),
			AmountNaet:	reward,
			Paid:		true,
		},
		SubmittedHeight:	height,
		UpdatedHeight:		height,
		ExpirationHeight:	height + params.EvidenceTTLBlocks,
		FinalizedHeight:	height,
	}
}

func validateDoubleSignEvidence(msg types.MsgSubmitDoubleSignEvidence) error {
	if msg.Height == 0 || msg.InfractionHeight == 0 || msg.Height < msg.InfractionHeight {
		return errors.New("native evidence double-sign heights are invalid")
	}
	if msg.ValidatorStake == 0 {
		return errors.New("native evidence double-sign validator stake must be positive")
	}
	if strings.TrimSpace(msg.VoteAHash) == strings.TrimSpace(msg.VoteBHash) {
		return errors.New("native evidence double-sign votes must be distinct")
	}
	if err := validateHexHash("native evidence double-sign vote A", msg.VoteAHash); err != nil {
		return err
	}
	return validateHexHash("native evidence double-sign vote B", msg.VoteBHash)
}

func validateDowntimeEvidence(msg types.MsgSubmitDowntimeEvidence) error {
	if msg.Height == 0 || msg.InfractionHeight == 0 || msg.Height < msg.InfractionHeight {
		return errors.New("native evidence downtime heights are invalid")
	}
	if msg.ValidatorStake == 0 {
		return errors.New("native evidence downtime validator stake must be positive")
	}
	if msg.WindowBlocks == 0 || msg.MissedBlocks == 0 || msg.MissedBlocks > msg.WindowBlocks {
		return errors.New("native evidence downtime missed/window blocks are invalid")
	}
	return nil
}

func validateHexHash(field, value string) error {
	value = strings.TrimSpace(value)
	if len(value) != 64 {
		return fmt.Errorf("%s must be 32-byte hex", field)
	}
	if _, err := hex.DecodeString(value); err != nil {
		return fmt.Errorf("%s must be hex: %w", field, err)
	}
	return nil
}

func proofHashParts(parts ...string) string {
	hash := sha256.New()
	for _, part := range parts {
		_, _ = hash.Write([]byte(strings.TrimSpace(part)))
		_, _ = hash.Write([]byte{0})
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func upsertOffenseCounter(counters []types.OffenseCounter, validator string, evidenceType string, height uint64) ([]types.OffenseCounter, uint64) {
	validator = strings.TrimSpace(validator)
	evidenceType = strings.TrimSpace(evidenceType)
	next := append([]types.OffenseCounter(nil), counters...)
	for idx, counter := range next {
		if counter.ValidatorAddress == validator && counter.EvidenceType == evidenceType {
			next[idx].Count++
			next[idx].LastHeight = height
			return types.SortOffenseCounters(next), next[idx].Count
		}
	}
	counter := types.OffenseCounter{ValidatorAddress: validator, EvidenceType: evidenceType, Count: 1, LastHeight: height}
	next = append(next, counter)
	return types.SortOffenseCounters(next), 1
}

func currentOffenseCount(counters []types.OffenseCounter, validator string, evidenceType string) uint64 {
	validator = strings.TrimSpace(validator)
	evidenceType = strings.TrimSpace(evidenceType)
	for _, counter := range counters {
		if counter.ValidatorAddress == validator && counter.EvidenceType == evidenceType {
			return counter.Count
		}
	}
	return 0
}

func slashAmount(stake uint64, fractionBps uint32) uint64 {
	fraction := uint64(fractionBps)
	if fraction == 0 {
		return 0
	}
	if stake > ^uint64(0)/fraction {
		return stake
	}
	amount := stake * fraction / uint64(types.MaxBasisPoints)
	if amount == 0 && stake > 0 && fractionBps > 0 {
		return 1
	}
	return amount
}

func integrationEvent(evidenceID string, validator string, target string, action string, height uint64) types.IntegrationEvent {
	return types.IntegrationEvent{
		EvidenceID:		evidenceID,
		ValidatorAddress:	validator,
		Target:			target,
		Action:			action,
		Height:			height,
	}
}

func findActiveJail(records []types.JailRecord, validator string) (int, types.JailRecord, bool) {
	for idx, record := range records {
		if record.ValidatorAddress == validator && record.Active {
			return idx, record, true
		}
	}
	return -1, types.JailRecord{}, false
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}
