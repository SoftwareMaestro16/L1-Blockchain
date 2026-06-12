package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/app/addressing"
)

const (
	ScoreMin	= uint8(0)
	ScoreMax	= uint8(100)

	LevelRestricted	= "restricted"
	LevelNew	= "new"
	LevelNormal	= "normal"
	LevelTrusted	= "trusted"
	LevelElite	= "elite"

	MaxDomainScore		= uint16(10)
	MaxContractScore	= uint16(15)

	DefaultReputationAuthority	= "4:0000000000000000000000000000000000000000000000000000000000000001"

	SubjectValidator	= "validator"
	SubjectReporter		= "reporter"
	SubjectAccount		= "account"

	ComponentMissedBlock	= "missed_block"
	ComponentSlashing	= "slashing"
	ComponentSpam		= "spam"
	ComponentUptime		= "uptime"
	ComponentRecovery	= "recovery"
	ComponentVolume		= "volume"
	ComponentStakeTime	= "stake_time"

	StakeReputationMaxScore			= uint16(100)
	DefaultStakeSecondsPerPoint		= uint64(3_600)
	DefaultJailedPoolExposureBps		= uint32(0)
	DefaultSlashedPoolExposureBps		= uint32(5_000)
	DefaultValidatorStakeBonusBps		= uint32(2_000)
	DefaultMaxStakeReputationRecords	= uint32(1_000_000)
)

type ReputationRecord struct {
	Account			sdk.AccAddress
	Score			uint8
	AgeScore		uint16
	StakingScore		uint16
	TxSuccessScore		uint16
	VolumeScore		uint16
	DomainScore		uint16
	ContractScore		uint16
	SpamPenalty		uint16
	FailedTxPenalty		uint16
	SlashPenalty		uint16
	LastUpdatedEpoch	uint64
}

type DecayParams struct {
	InactiveAfterEpochs	uint64
	DecayRatePerEpoch	uint8
}

type ReputationParams struct {
	Authority			string
	MinScore			uint8
	MaxScore			uint8
	MissedBlockPenalty		uint16
	SlashingPenalty			uint16
	UptimeReward			uint16
	RecoveryReward			uint16
	SlashingReducesScore		bool
	Decay				DecayParams
	MaxHistorySnapshots		uint32
	MaxEvents			uint32
	StakeSecondsPerPoint		uint64
	JailedPoolExposureBps		uint32
	SlashedPoolExposureBps		uint32
	ValidatorStakeBonusBps		uint32
	MaxStakeReputationRecords	uint32
}

type ReputationEvent struct {
	EventID		string
	SubjectType	string
	Subject		sdk.AccAddress
	Component	string
	Amount		uint16
	Reason		string
	Epoch		uint64
	ScoreBefore	uint8
	ScoreAfter	uint8
	EventHash	string
}

type ReputationSnapshot struct {
	Epoch		uint64
	ValidatorScores	[]ReputationScore
	ReporterScores	[]ReputationScore
	SnapshotHash	string
}

type ReputationScore struct {
	Account	sdk.AccAddress
	Score	uint8
}

type ReputationState struct {
	Params		ReputationParams
	Accounts	[]ReputationRecord
	Validators	[]ReputationRecord
	Reporters	[]ReputationRecord
	StakeRecords	[]StakeReputationRecord
	Snapshots	[]ReputationSnapshot
	PenaltyEvents	[]ReputationEvent
	RecoveryEvents	[]ReputationEvent
}

type StakePoolExposure struct {
	PoolID			string
	Shares			uint64
	TotalPoolShares		uint64
	PoolActiveStake		uint64
	EffectiveStake		uint64
	LastUpdatedUnix		uint64
	ValidatorJailed		bool
	ValidatorSlashed	bool
	ValidatorOperator	bool
	ValidatorBonusBps	uint32
	ValidatorBonusBlocked	bool
}

type StakeReputationRecord struct {
	Account				sdk.AccAddress
	AccountUser			string
	StakeWeightedSeconds		uint64
	ClaimedStakeWeightedSeconds	uint64
	ClaimedStakeReputation		uint16
	LastUpdatedUnix			uint64
	PoolExposures			[]StakePoolExposure
	NonTransferable			bool
}

type StakeReputationClaim struct {
	Account				sdk.AccAddress
	AccountUser			string
	StakeWeightedSeconds		uint64
	ClaimableStakeWeightedSeconds	uint64
	ReputationDelta			uint16
	ClaimedStakeReputation		uint16
	AccountScoreAfter		uint8
	ClaimHash			string
}

type MsgUpdateReputationParams struct {
	Authority	string
	Params		ReputationParams
}

type MsgApplyReputationPenalty struct {
	Authority	string
	SubjectType	string
	Subject		sdk.AccAddress
	Component	string
	Amount		uint16
	Reason		string
	Epoch		uint64
}

type MsgApplyReputationReward struct {
	Authority	string
	SubjectType	string
	Subject		sdk.AccAddress
	Component	string
	Amount		uint16
	Reason		string
	Epoch		uint64
}

type MsgRecomputeReputation struct {
	Authority	string
	SubjectType	string
	Subject		sdk.AccAddress
	Epoch		uint64
}

type MsgClaimStakeReputation struct {
	Authority		string
	Account			string
	PoolID			string
	PoolShares		uint64
	PoolTotalShares		uint64
	PoolActiveStake		uint64
	TimestampUnix		uint64
	ValidatorOperator	bool
	ValidatorJailed		bool
	ValidatorSlashed	bool
}

type QueryStakeReputationRequest struct {
	Account string
}

type QueryStakeReputationResponse struct {
	Record StakeReputationRecord
}

type QueryAccountReputationRequest struct {
	Account string
}

type QueryAccountReputationResponse struct {
	Record		ReputationRecord
	StakeReputation	StakeReputationRecord
}

type ReputationHistoryQuery struct {
	SubjectType	string
	Subject		sdk.AccAddress
	Limit		uint32
}

type ProgressiveLimits struct {
	MaxTxsPerBlock	uint32
	MaxTxGas	uint64
	MaxQueueMsgs	uint32
}

func DefaultReputationParams() ReputationParams {
	return ReputationParams{
		Authority:		DefaultReputationAuthority,
		MinScore:		ScoreMin,
		MaxScore:		ScoreMax,
		MissedBlockPenalty:	5,
		SlashingPenalty:	25,
		UptimeReward:		2,
		RecoveryReward:		3,
		SlashingReducesScore:	true,
		Decay: DecayParams{
			InactiveAfterEpochs:	10,
			DecayRatePerEpoch:	1,
		},
		MaxHistorySnapshots:		1024,
		MaxEvents:			4096,
		StakeSecondsPerPoint:		DefaultStakeSecondsPerPoint,
		JailedPoolExposureBps:		DefaultJailedPoolExposureBps,
		SlashedPoolExposureBps:		DefaultSlashedPoolExposureBps,
		ValidatorStakeBonusBps:		DefaultValidatorStakeBonusBps,
		MaxStakeReputationRecords:	DefaultMaxStakeReputationRecords,
	}
}

func NewReputationState(params ReputationParams) (ReputationState, error) {
	if params.Authority == "" {
		params = DefaultReputationParams()
	}
	if err := params.Validate(); err != nil {
		return ReputationState{}, err
	}
	return ReputationState{Params: params}, nil
}

func ComputeScore(record ReputationRecord) uint8 {
	positive := record.AgeScore +
		record.StakingScore +
		record.TxSuccessScore +
		record.VolumeScore +
		minU16(record.DomainScore, MaxDomainScore) +
		minU16(record.ContractScore, MaxContractScore)
	negative := record.SpamPenalty + record.FailedTxPenalty + record.SlashPenalty
	if negative >= positive {
		return ScoreMin
	}
	net := positive - negative
	if net > uint16(ScoreMax) {
		return ScoreMax
	}
	return uint8(net)
}

func (params ReputationParams) Validate() error {
	if err := addressing.ValidateAuthorityAddress("reputation authority", params.Authority); err != nil {
		return err
	}
	if params.MinScore != ScoreMin {
		return fmt.Errorf("reputation min score must be %d", ScoreMin)
	}
	if params.MaxScore != ScoreMax {
		return fmt.Errorf("reputation max score must be %d", ScoreMax)
	}
	if params.MissedBlockPenalty == 0 {
		return errors.New("reputation missed block penalty must be positive")
	}
	if params.SlashingReducesScore && params.SlashingPenalty == 0 {
		return errors.New("reputation slashing penalty must be positive when slashing reduces score")
	}
	if params.UptimeReward == 0 {
		return errors.New("reputation uptime reward must be positive")
	}
	if params.RecoveryReward == 0 {
		return errors.New("reputation recovery reward must be positive")
	}
	if params.MaxHistorySnapshots == 0 {
		return errors.New("reputation max history snapshots must be positive")
	}
	if params.MaxEvents == 0 {
		return errors.New("reputation max events must be positive")
	}
	if params.StakeSecondsPerPoint == 0 {
		return errors.New("reputation stake seconds per point must be positive")
	}
	if params.JailedPoolExposureBps > 10_000 {
		return errors.New("reputation jailed pool exposure bps exceeds basis points")
	}
	if params.SlashedPoolExposureBps > 10_000 {
		return errors.New("reputation slashed pool exposure bps exceeds basis points")
	}
	if params.ValidatorStakeBonusBps > 10_000 {
		return errors.New("reputation validator stake bonus bps exceeds basis points")
	}
	if params.MaxStakeReputationRecords == 0 {
		return errors.New("reputation max stake records must be positive")
	}
	return nil
}

func (state ReputationState) Validate() error {
	state = NormalizeReputationState(state)
	if err := state.Params.Validate(); err != nil {
		return err
	}
	if err := validateReputationRecords("account", state.Accounts); err != nil {
		return err
	}
	if err := validateReputationRecords("validator", state.Validators); err != nil {
		return err
	}
	if err := validateReputationRecords("reporter", state.Reporters); err != nil {
		return err
	}
	if uint32(len(state.StakeRecords)) > state.Params.MaxStakeReputationRecords {
		return errors.New("reputation stake record limit exceeded")
	}
	if err := validateStakeReputationRecords(state.StakeRecords); err != nil {
		return err
	}
	if uint32(len(state.Snapshots)) > state.Params.MaxHistorySnapshots {
		return errors.New("reputation snapshot history exceeds configured limit")
	}
	if uint32(len(state.PenaltyEvents)) > state.Params.MaxEvents {
		return errors.New("reputation penalty events exceed configured limit")
	}
	if uint32(len(state.RecoveryEvents)) > state.Params.MaxEvents {
		return errors.New("reputation recovery events exceed configured limit")
	}
	for _, snapshot := range state.Snapshots {
		if err := snapshot.Validate(); err != nil {
			return err
		}
	}
	for _, event := range state.PenaltyEvents {
		if err := event.Validate(); err != nil {
			return err
		}
	}
	for _, event := range state.RecoveryEvents {
		if err := event.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func ApplyUpdateReputationParams(state ReputationState, msg MsgUpdateReputationParams) (ReputationState, error) {
	state = NormalizeReputationState(state)
	if err := authorizeReputation(state.Params, msg.Authority); err != nil {
		return ReputationState{}, err
	}
	msg.Params.Authority = strings.TrimSpace(msg.Params.Authority)
	if msg.Params.Authority == "" {
		msg.Params.Authority = state.Params.Authority
	}
	if err := msg.Params.Validate(); err != nil {
		return ReputationState{}, err
	}
	state.Params = msg.Params
	return state, state.Validate()
}

func ApplyReputationPenalty(state ReputationState, msg MsgApplyReputationPenalty) (ReputationState, error) {
	state = NormalizeReputationState(state)
	if err := authorizeReputation(state.Params, msg.Authority); err != nil {
		return ReputationState{}, err
	}
	if err := validateSubject(msg.SubjectType, msg.Subject); err != nil {
		return ReputationState{}, err
	}
	amount := msg.Amount
	if amount == 0 {
		amount = defaultPenaltyAmount(state.Params, msg.Component)
	}
	if amount == 0 {
		return ReputationState{}, errors.New("reputation penalty amount must be positive")
	}
	record, found := state.recordBySubject(msg.SubjectType, msg.Subject)
	if !found {
		record = ApplyComputedScore(ReputationRecord{Account: cloneAddress(msg.Subject)})
	}
	before := record.Score
	record = applyPenaltyComponent(record, msg.Component, amount, msg.Epoch)
	if err := ValidateReputationRecord(record); err != nil {
		return ReputationState{}, err
	}
	if msg.Component == ComponentSlashing && state.Params.SlashingReducesScore && before > ScoreMin && record.Score >= before {
		return ReputationState{}, errors.New("reputation slashing penalty must reduce score")
	}
	state = state.upsertRecord(msg.SubjectType, record)
	event, err := NewReputationEvent(msg.SubjectType, msg.Subject, msg.Component, amount, msg.Reason, msg.Epoch, before, record.Score)
	if err != nil {
		return ReputationState{}, err
	}
	state.PenaltyEvents = append(state.PenaltyEvents, event)
	state.PenaltyEvents = trimEvents(state.PenaltyEvents, state.Params.MaxEvents)
	state = NormalizeReputationState(state)
	return state, state.Validate()
}

func ApplyReputationReward(state ReputationState, msg MsgApplyReputationReward) (ReputationState, error) {
	state = NormalizeReputationState(state)
	if err := authorizeReputation(state.Params, msg.Authority); err != nil {
		return ReputationState{}, err
	}
	if err := validateSubject(msg.SubjectType, msg.Subject); err != nil {
		return ReputationState{}, err
	}
	amount := msg.Amount
	if amount == 0 {
		amount = defaultRewardAmount(state.Params, msg.Component)
	}
	if amount == 0 {
		return ReputationState{}, errors.New("reputation reward amount must be positive")
	}
	record, found := state.recordBySubject(msg.SubjectType, msg.Subject)
	if !found {
		record = ApplyComputedScore(ReputationRecord{Account: cloneAddress(msg.Subject)})
	}
	before := record.Score
	record = applyRewardComponent(record, msg.Component, amount, msg.Epoch)
	if err := ValidateReputationRecord(record); err != nil {
		return ReputationState{}, err
	}
	state = state.upsertRecord(msg.SubjectType, record)
	event, err := NewReputationEvent(msg.SubjectType, msg.Subject, msg.Component, amount, msg.Reason, msg.Epoch, before, record.Score)
	if err != nil {
		return ReputationState{}, err
	}
	state.RecoveryEvents = append(state.RecoveryEvents, event)
	state.RecoveryEvents = trimEvents(state.RecoveryEvents, state.Params.MaxEvents)
	state = NormalizeReputationState(state)
	return state, state.Validate()
}

func ApplyRecomputeReputation(state ReputationState, msg MsgRecomputeReputation) (ReputationState, error) {
	state = NormalizeReputationState(state)
	if err := authorizeReputation(state.Params, msg.Authority); err != nil {
		return ReputationState{}, err
	}
	record, found := state.recordBySubject(msg.SubjectType, msg.Subject)
	if !found {
		return ReputationState{}, errors.New("reputation record not found")
	}
	record.Score = ComputeScore(record)
	record.LastUpdatedEpoch = msg.Epoch
	state = state.upsertRecord(msg.SubjectType, record)
	state = NormalizeReputationState(state)
	return state, state.Validate()
}

func SnapshotReputationEpoch(state ReputationState, epoch uint64) (ReputationState, ReputationSnapshot, error) {
	state = NormalizeReputationState(state)
	if epoch == 0 {
		return ReputationState{}, ReputationSnapshot{}, errors.New("reputation snapshot epoch must be positive")
	}
	snapshot := ReputationSnapshot{
		Epoch:			epoch,
		ValidatorScores:	scoresFromRecords(state.Validators),
		ReporterScores:		scoresFromRecords(state.Reporters),
	}
	snapshot.SnapshotHash = ComputeReputationSnapshotHash(snapshot)
	if err := snapshot.Validate(); err != nil {
		return ReputationState{}, ReputationSnapshot{}, err
	}
	state.Snapshots = append(state.Snapshots, snapshot)
	state.Snapshots = trimSnapshots(state.Snapshots, state.Params.MaxHistorySnapshots)
	state = NormalizeReputationState(state)
	return state, snapshot, state.Validate()
}

func QueryValidatorReputation(state ReputationState, validator sdk.AccAddress) (ReputationRecord, bool) {
	state = NormalizeReputationState(state)
	return state.recordBySubject(SubjectValidator, validator)
}

func QueryReporterReputation(state ReputationState, reporter sdk.AccAddress) (ReputationRecord, bool) {
	state = NormalizeReputationState(state)
	return state.recordBySubject(SubjectReporter, reporter)
}

func ApplyClaimStakeReputation(state ReputationState, msg MsgClaimStakeReputation) (ReputationState, StakeReputationClaim, error) {
	state = NormalizeReputationState(state)
	if err := authorizeReputation(state.Params, msg.Authority); err != nil {
		return ReputationState{}, StakeReputationClaim{}, err
	}
	account, err := addressing.ParseUserAddress("stake reputation account", msg.Account)
	if err != nil {
		return ReputationState{}, StakeReputationClaim{}, err
	}
	if strings.TrimSpace(msg.PoolID) == "" {
		return ReputationState{}, StakeReputationClaim{}, errors.New("stake reputation pool id is required")
	}
	if msg.TimestampUnix == 0 {
		return ReputationState{}, StakeReputationClaim{}, errors.New("stake reputation timestamp must be positive")
	}
	if msg.PoolShares > 0 && msg.PoolTotalShares == 0 {
		return ReputationState{}, StakeReputationClaim{}, errors.New("stake reputation total pool shares must be positive when shares are positive")
	}
	record, found := state.stakeRecord(account)
	if !found {
		record = NewStakeReputationRecord(account)
	}
	nextRecord, err := AccumulateStakeExposure(state.Params, record, StakePoolExposure{
		PoolID:			strings.TrimSpace(msg.PoolID),
		Shares:			msg.PoolShares,
		TotalPoolShares:	msg.PoolTotalShares,
		PoolActiveStake:	msg.PoolActiveStake,
		LastUpdatedUnix:	msg.TimestampUnix,
		ValidatorJailed:	msg.ValidatorJailed,
		ValidatorSlashed:	msg.ValidatorSlashed,
		ValidatorOperator:	msg.ValidatorOperator,
		ValidatorBonusBps:	state.Params.ValidatorStakeBonusBps,
	})
	if err != nil {
		return ReputationState{}, StakeReputationClaim{}, err
	}

	claimable := nextRecord.StakeWeightedSeconds - nextRecord.ClaimedStakeWeightedSeconds
	points := claimable / state.Params.StakeSecondsPerPoint
	if points > uint64(^uint16(0)) {
		points = uint64(^uint16(0))
	}
	remainingScore := uint64(0)
	if nextRecord.ClaimedStakeReputation < StakeReputationMaxScore {
		remainingScore = uint64(StakeReputationMaxScore - nextRecord.ClaimedStakeReputation)
	}
	if points > remainingScore {
		points = remainingScore
	}
	reputationDelta := uint16(points)
	if reputationDelta == 0 {
		state = state.upsertStakeRecord(nextRecord)
		return state, StakeReputationClaim{
			Account:			cloneAddress(account),
			AccountUser:			nextRecord.AccountUser,
			StakeWeightedSeconds:		nextRecord.StakeWeightedSeconds,
			ClaimableStakeWeightedSeconds:	claimable,
			ClaimedStakeReputation:		nextRecord.ClaimedStakeReputation,
			AccountScoreAfter:		accountScore(state, account),
			ClaimHash:			ComputeStakeReputationClaimHash(nextRecord, 0, 0),
		}, state.Validate()
	}
	nextRecord.ClaimedStakeWeightedSeconds += uint64(reputationDelta) * state.Params.StakeSecondsPerPoint
	nextRecord.ClaimedStakeReputation += reputationDelta

	accountRecord, found := state.recordBySubject(SubjectAccount, account)
	if !found {
		accountRecord = ApplyComputedScore(ReputationRecord{Account: cloneAddress(account)})
	}
	before := accountRecord.Score
	accountRecord.StakingScore = nextRecord.ClaimedStakeReputation
	accountRecord.LastUpdatedEpoch = msg.TimestampUnix
	accountRecord = ApplyComputedScore(accountRecord)
	if accountRecord.Score <= before && reputationDelta > 0 {
		return ReputationState{}, StakeReputationClaim{}, errors.New("stake-time reputation claim must increase account reputation")
	}
	state = state.upsertRecord(SubjectAccount, accountRecord)
	state = state.upsertStakeRecord(nextRecord)
	state = NormalizeReputationState(state)
	claim := StakeReputationClaim{
		Account:			cloneAddress(account),
		AccountUser:			nextRecord.AccountUser,
		StakeWeightedSeconds:		nextRecord.StakeWeightedSeconds,
		ClaimableStakeWeightedSeconds:	claimable,
		ReputationDelta:		reputationDelta,
		ClaimedStakeReputation:		nextRecord.ClaimedStakeReputation,
		AccountScoreAfter:		accountRecord.Score,
	}
	claim.ClaimHash = ComputeStakeReputationClaimHash(nextRecord, reputationDelta, accountRecord.Score)
	return state, claim, state.Validate()
}

func QueryStakeReputation(state ReputationState, account sdk.AccAddress) (StakeReputationRecord, bool) {
	state = NormalizeReputationState(state)
	return state.stakeRecord(account)
}

func QueryAccountReputation(state ReputationState, account sdk.AccAddress) (ReputationRecord, StakeReputationRecord, bool) {
	state = NormalizeReputationState(state)
	record, found := state.recordBySubject(SubjectAccount, account)
	stake, stakeFound := state.stakeRecord(account)
	return record, stake, found || stakeFound
}

func QueryReputationHistory(state ReputationState, query ReputationHistoryQuery) ([]ReputationSnapshot, []ReputationEvent, error) {
	state = NormalizeReputationState(state)
	if err := validateSubject(query.SubjectType, query.Subject); err != nil {
		return nil, nil, err
	}
	limit := query.Limit
	if limit == 0 || limit > state.Params.MaxHistorySnapshots {
		limit = state.Params.MaxHistorySnapshots
	}
	snapshots := make([]ReputationSnapshot, 0, len(state.Snapshots))
	for _, snapshot := range state.Snapshots {
		if snapshotContains(snapshot, query.SubjectType, query.Subject) {
			snapshots = append(snapshots, snapshot)
		}
	}
	if uint32(len(snapshots)) > limit {
		snapshots = snapshots[len(snapshots)-int(limit):]
	}
	events := make([]ReputationEvent, 0)
	for _, event := range append(append([]ReputationEvent(nil), state.PenaltyEvents...), state.RecoveryEvents...) {
		if event.SubjectType == query.SubjectType && addressKey(event.Subject) == addressKey(query.Subject) {
			events = append(events, event)
		}
	}
	sortEvents(events)
	return snapshots, events, nil
}

func QueryReputationParams(state ReputationState) ReputationParams {
	return state.Params
}

func ExportReputationState(state ReputationState) (ReputationState, error) {
	state = NormalizeReputationState(state)
	if err := state.Validate(); err != nil {
		return ReputationState{}, err
	}
	return cloneReputationState(state), nil
}

func ImportReputationState(exported ReputationState) (ReputationState, error) {
	exported = NormalizeReputationState(exported)
	if err := exported.Validate(); err != nil {
		return ReputationState{}, err
	}
	return cloneReputationState(exported), nil
}

func NewReputationEvent(subjectType string, subject sdk.AccAddress, component string, amount uint16, reason string, epoch uint64, before uint8, after uint8) (ReputationEvent, error) {
	event := ReputationEvent{
		SubjectType:	subjectType,
		Subject:	cloneAddress(subject),
		Component:	strings.TrimSpace(component),
		Amount:		amount,
		Reason:		strings.TrimSpace(reason),
		Epoch:		epoch,
		ScoreBefore:	before,
		ScoreAfter:	after,
	}
	if event.Reason == "" {
		event.Reason = event.Component
	}
	if err := event.ValidateFormat(); err != nil {
		return ReputationEvent{}, err
	}
	event.EventID = ComputeReputationEventID(event)
	event.EventHash = ComputeReputationEventHash(event)
	return event, event.Validate()
}

func (event ReputationEvent) ValidateFormat() error {
	if err := validateSubject(event.SubjectType, event.Subject); err != nil {
		return err
	}
	if !isReputationComponent(event.Component) {
		return fmt.Errorf("unknown reputation component %q", event.Component)
	}
	if event.Amount == 0 {
		return errors.New("reputation event amount must be positive")
	}
	if event.Epoch == 0 {
		return errors.New("reputation event epoch must be positive")
	}
	if strings.TrimSpace(event.Reason) == "" {
		return errors.New("reputation event reason is required")
	}
	if event.EventID != "" && !isHexHash(event.EventID) {
		return errors.New("reputation event id must be a hex hash")
	}
	if event.EventHash != "" && !isHexHash(event.EventHash) {
		return errors.New("reputation event hash must be a hex hash")
	}
	return nil
}

func (event ReputationEvent) Validate() error {
	if err := event.ValidateFormat(); err != nil {
		return err
	}
	if event.EventID != ComputeReputationEventID(event) {
		return errors.New("reputation event id mismatch")
	}
	if event.EventHash != ComputeReputationEventHash(event) {
		return errors.New("reputation event hash mismatch")
	}
	return nil
}

func (snapshot ReputationSnapshot) Validate() error {
	if snapshot.Epoch == 0 {
		return errors.New("reputation snapshot epoch must be positive")
	}
	if err := validateScores(snapshot.ValidatorScores); err != nil {
		return err
	}
	if err := validateScores(snapshot.ReporterScores); err != nil {
		return err
	}
	if snapshot.SnapshotHash == "" || !isHexHash(snapshot.SnapshotHash) {
		return errors.New("reputation snapshot hash must be a hex hash")
	}
	if snapshot.SnapshotHash != ComputeReputationSnapshotHash(snapshot) {
		return errors.New("reputation snapshot hash mismatch")
	}
	return nil
}

func CheckReputationInvariants(state ReputationState) error {
	state = NormalizeReputationState(state)
	return state.Validate()
}

func NormalizeReputationState(state ReputationState) ReputationState {
	if state.Params.Authority == "" {
		state.Params = DefaultReputationParams()
	}
	state.Params.Authority = strings.TrimSpace(state.Params.Authority)
	state.Accounts = normalizeRecords(state.Accounts)
	state.Validators = normalizeRecords(state.Validators)
	state.Reporters = normalizeRecords(state.Reporters)
	state.StakeRecords = normalizeStakeReputationRecords(state.StakeRecords)
	state.Snapshots = normalizeSnapshots(state.Snapshots)
	state.PenaltyEvents = normalizeEvents(state.PenaltyEvents)
	state.RecoveryEvents = normalizeEvents(state.RecoveryEvents)
	return state
}

func ApplyComputedScore(record ReputationRecord) ReputationRecord {
	record.Score = ComputeScore(record)
	return record
}

func ApplyInactivityDecay(score uint8, inactiveEpochs uint64, params DecayParams) uint8 {
	if inactiveEpochs <= params.InactiveAfterEpochs || params.DecayRatePerEpoch == 0 {
		return score
	}
	decayEpochs := inactiveEpochs - params.InactiveAfterEpochs
	decay := decayEpochs * uint64(params.DecayRatePerEpoch)
	if decay >= uint64(score) {
		return ScoreMin
	}
	return uint8(uint64(score) - decay)
}

func LevelForScore(score uint8) string {
	switch {
	case score < 20:
		return LevelRestricted
	case score < 50:
		return LevelNew
	case score < 80:
		return LevelNormal
	case score < 95:
		return LevelTrusted
	default:
		return LevelElite
	}
}

func ValidateReputationRecord(record ReputationRecord) error {
	if len(record.Account) == 0 {
		return errors.New("reputation account is required")
	}
	if err := addressing.RejectZeroAddress("reputation account", record.Account); err != nil {
		return err
	}
	expected := ComputeScore(record)
	if record.Score != expected {
		return fmt.Errorf("reputation score mismatch: expected %d got %d", expected, record.Score)
	}
	if record.DomainScore > MaxDomainScore {
		return fmt.Errorf("domain score must not exceed %d", MaxDomainScore)
	}
	if record.ContractScore > MaxContractScore {
		return fmt.Errorf("contract score must not exceed %d", MaxContractScore)
	}
	return nil
}

func LimitsForScore(score uint8) ProgressiveLimits {
	switch LevelForScore(score) {
	case LevelRestricted:
		return ProgressiveLimits{MaxTxsPerBlock: 1, MaxTxGas: 100_000, MaxQueueMsgs: 1}
	case LevelNew:
		return ProgressiveLimits{MaxTxsPerBlock: 5, MaxTxGas: 250_000, MaxQueueMsgs: 4}
	case LevelNormal:
		return ProgressiveLimits{MaxTxsPerBlock: 25, MaxTxGas: 1_000_000, MaxQueueMsgs: 16}
	case LevelTrusted:
		return ProgressiveLimits{MaxTxsPerBlock: 100, MaxTxGas: 2_000_000, MaxQueueMsgs: 64}
	default:
		return ProgressiveLimits{MaxTxsPerBlock: 250, MaxTxGas: 5_000_000, MaxQueueMsgs: 128}
	}
}

func IsDirectReputationPurchaseAllowed() bool {
	return false
}

func IsStakeReputationTransferableAsTokenNFTOrDomain() bool {
	return false
}

func NewStakeReputationRecord(account sdk.AccAddress) StakeReputationRecord {
	return StakeReputationRecord{
		Account:		cloneAddress(account),
		AccountUser:		addressing.FormatAccAddress(account),
		NonTransferable:	true,
	}
}

func AccumulateStakeExposure(params ReputationParams, record StakeReputationRecord, nextExposure StakePoolExposure) (StakeReputationRecord, error) {
	if err := params.Validate(); err != nil {
		return StakeReputationRecord{}, err
	}
	record = normalizeStakeReputationRecord(record)
	if err := record.Validate(); err != nil {
		return StakeReputationRecord{}, err
	}
	nextExposure.PoolID = strings.TrimSpace(nextExposure.PoolID)
	if nextExposure.PoolID == "" {
		return StakeReputationRecord{}, errors.New("stake reputation pool id is required")
	}
	if nextExposure.LastUpdatedUnix == 0 {
		return StakeReputationRecord{}, errors.New("stake reputation exposure timestamp must be positive")
	}
	idx, previous, found := findStakePoolExposure(record.PoolExposures, nextExposure.PoolID)
	if found {
		if nextExposure.LastUpdatedUnix < previous.LastUpdatedUnix {
			return StakeReputationRecord{}, errors.New("stake reputation exposure timestamp cannot move backwards")
		}
		duration := nextExposure.LastUpdatedUnix - previous.LastUpdatedUnix
		increment, err := stakeWeightedSeconds(previous.EffectiveStake, duration)
		if err != nil {
			return StakeReputationRecord{}, err
		}
		record.StakeWeightedSeconds, err = checkedAddUint64(record.StakeWeightedSeconds, increment)
		if err != nil {
			return StakeReputationRecord{}, err
		}
	} else if record.LastUpdatedUnix != 0 && nextExposure.LastUpdatedUnix < record.LastUpdatedUnix {
		return StakeReputationRecord{}, errors.New("stake reputation exposure timestamp cannot move backwards")
	}
	nextExposure.EffectiveStake = EffectivePoolStakeExposure(params, nextExposure)
	if nextExposure.ValidatorJailed || nextExposure.ValidatorSlashed {
		nextExposure.ValidatorBonusBlocked = true
		nextExposure.ValidatorBonusBps = 0
	}
	if !nextExposure.ValidatorOperator {
		nextExposure.ValidatorBonusBps = 0
	}
	if found {
		record.PoolExposures[idx] = nextExposure
	} else {
		record.PoolExposures = append(record.PoolExposures, nextExposure)
	}
	record.LastUpdatedUnix = maxUint64(record.LastUpdatedUnix, nextExposure.LastUpdatedUnix)
	record.PoolExposures = normalizeStakePoolExposures(record.PoolExposures)
	return record, record.Validate()
}

func EffectivePoolStakeExposure(params ReputationParams, exposure StakePoolExposure) uint64 {
	if exposure.Shares == 0 || exposure.TotalPoolShares == 0 || exposure.PoolActiveStake == 0 {
		return 0
	}
	effective := mulDivFloor(exposure.PoolActiveStake, exposure.Shares, exposure.TotalPoolShares)
	if exposure.ValidatorJailed {
		effective = mulDivFloor(effective, uint64(params.JailedPoolExposureBps), 10_000)
	} else if exposure.ValidatorSlashed {
		effective = mulDivFloor(effective, uint64(params.SlashedPoolExposureBps), 10_000)
	}
	if exposure.ValidatorOperator && !exposure.ValidatorJailed && !exposure.ValidatorSlashed {
		bonus := mulDivFloor(effective, uint64(params.ValidatorStakeBonusBps), 10_000)
		next, err := checkedAddUint64(effective, bonus)
		if err != nil {
			return ^uint64(0)
		}
		return next
	}
	return effective
}

func ValidateStakeReputationTransfer(record StakeReputationRecord, assetKind string) error {
	switch strings.TrimSpace(assetKind) {
	case "domain":
		return errors.New("stake reputation is account-owned and cannot be transferred as domain ownership")
	default:
		if !record.NonTransferable {
			return errors.New("stake reputation record must be marked non-transferable")
		}
		return nil
	}
}

func (state ReputationState) recordBySubject(subjectType string, subject sdk.AccAddress) (ReputationRecord, bool) {
	key := addressKey(subject)
	switch subjectType {
	case SubjectAccount:
		for _, record := range state.Accounts {
			if addressKey(record.Account) == key {
				return cloneRecord(record), true
			}
		}
	case SubjectValidator:
		for _, record := range state.Validators {
			if addressKey(record.Account) == key {
				return cloneRecord(record), true
			}
		}
	case SubjectReporter:
		for _, record := range state.Reporters {
			if addressKey(record.Account) == key {
				return cloneRecord(record), true
			}
		}
	}
	return ReputationRecord{}, false
}

func (state ReputationState) upsertRecord(subjectType string, record ReputationRecord) ReputationState {
	record = cloneRecord(record)
	switch subjectType {
	case SubjectAccount:
		state.Accounts = upsertRecord(state.Accounts, record)
	case SubjectValidator:
		state.Validators = upsertRecord(state.Validators, record)
	case SubjectReporter:
		state.Reporters = upsertRecord(state.Reporters, record)
	}
	return NormalizeReputationState(state)
}

func (state ReputationState) stakeRecord(account sdk.AccAddress) (StakeReputationRecord, bool) {
	key := addressKey(account)
	for _, record := range state.StakeRecords {
		if addressKey(record.Account) == key {
			return cloneStakeReputationRecord(record), true
		}
	}
	return StakeReputationRecord{}, false
}

func (state ReputationState) upsertStakeRecord(record StakeReputationRecord) ReputationState {
	key := addressKey(record.Account)
	next := make([]StakeReputationRecord, 0, len(state.StakeRecords)+1)
	replaced := false
	for _, existing := range state.StakeRecords {
		if addressKey(existing.Account) == key {
			next = append(next, cloneStakeReputationRecord(record))
			replaced = true
			continue
		}
		next = append(next, cloneStakeReputationRecord(existing))
	}
	if !replaced {
		next = append(next, cloneStakeReputationRecord(record))
	}
	state.StakeRecords = normalizeStakeReputationRecords(next)
	return NormalizeReputationState(state)
}

func authorizeReputation(params ReputationParams, authority string) error {
	if err := params.Validate(); err != nil {
		return err
	}
	authority = strings.TrimSpace(authority)
	if err := addressing.ValidateAuthorityAddress("reputation message authority", authority); err != nil {
		return err
	}
	if authority != params.Authority {
		return errors.New("reputation message requires authority")
	}
	return nil
}

func validateSubject(subjectType string, subject sdk.AccAddress) error {
	switch subjectType {
	case SubjectAccount, SubjectValidator, SubjectReporter:
	default:
		return fmt.Errorf("unknown reputation subject type %q", subjectType)
	}
	if len(subject) == 0 {
		return errors.New("reputation subject is required")
	}
	return addressing.RejectZeroAddress("reputation subject", subject)
}

func defaultPenaltyAmount(params ReputationParams, component string) uint16 {
	switch component {
	case ComponentMissedBlock:
		return params.MissedBlockPenalty
	case ComponentSlashing:
		return params.SlashingPenalty
	default:
		return 0
	}
}

func defaultRewardAmount(params ReputationParams, component string) uint16 {
	switch component {
	case ComponentUptime:
		return params.UptimeReward
	case ComponentRecovery:
		return params.RecoveryReward
	default:
		return 0
	}
}

func applyPenaltyComponent(record ReputationRecord, component string, amount uint16, epoch uint64) ReputationRecord {
	switch component {
	case ComponentMissedBlock:
		record.FailedTxPenalty = saturatingAddU16(record.FailedTxPenalty, amount)
	case ComponentSlashing:
		record.SlashPenalty = saturatingAddU16(record.SlashPenalty, amount)
	case ComponentSpam:
		record.SpamPenalty = saturatingAddU16(record.SpamPenalty, amount)
	default:
		record.FailedTxPenalty = saturatingAddU16(record.FailedTxPenalty, amount)
	}
	record.LastUpdatedEpoch = epoch
	return ApplyComputedScore(record)
}

func applyRewardComponent(record ReputationRecord, component string, amount uint16, epoch uint64) ReputationRecord {
	switch component {
	case ComponentUptime:
		record.TxSuccessScore = saturatingAddU16(record.TxSuccessScore, amount)
	case ComponentRecovery:
		record.AgeScore = saturatingAddU16(record.AgeScore, amount)
	case ComponentVolume:
		record.VolumeScore = saturatingAddU16(record.VolumeScore, amount)
	default:
		record.TxSuccessScore = saturatingAddU16(record.TxSuccessScore, amount)
	}
	record.LastUpdatedEpoch = epoch
	return ApplyComputedScore(record)
}

func isReputationComponent(component string) bool {
	switch component {
	case ComponentMissedBlock, ComponentSlashing, ComponentSpam, ComponentUptime, ComponentRecovery, ComponentVolume, ComponentStakeTime:
		return true
	default:
		return false
	}
}

func validateReputationRecords(label string, records []ReputationRecord) error {
	seen := make(map[string]struct{}, len(records))
	var previous string
	for i, record := range records {
		if err := ValidateReputationRecord(record); err != nil {
			return fmt.Errorf("%s reputation record invalid: %w", label, err)
		}
		key := addressKey(record.Account)
		if _, found := seen[key]; found {
			return fmt.Errorf("duplicate %s reputation record", label)
		}
		seen[key] = struct{}{}
		if i > 0 && previous >= key {
			return fmt.Errorf("%s reputation records must be sorted deterministically", label)
		}
		previous = key
	}
	return nil
}

func validateStakeReputationRecords(records []StakeReputationRecord) error {
	seen := make(map[string]struct{}, len(records))
	var previous string
	for i, record := range records {
		if err := record.Validate(); err != nil {
			return fmt.Errorf("stake reputation record invalid: %w", err)
		}
		key := addressKey(record.Account)
		if _, found := seen[key]; found {
			return errors.New("duplicate stake reputation record")
		}
		seen[key] = struct{}{}
		if i > 0 && previous >= key {
			return errors.New("stake reputation records must be sorted deterministically")
		}
		previous = key
	}
	return nil
}

func (record StakeReputationRecord) Validate() error {
	if len(record.Account) == 0 {
		return errors.New("stake reputation account is required")
	}
	if err := addressing.RejectZeroAddress("stake reputation account", record.Account); err != nil {
		return err
	}
	if strings.TrimSpace(record.AccountUser) == "" {
		return errors.New("stake reputation account user address is required")
	}
	parsed, err := addressing.ParseUserAddress("stake reputation account user", record.AccountUser)
	if err != nil {
		return err
	}
	if addressKey(parsed) != addressKey(record.Account) {
		return errors.New("stake reputation AE account does not match account bytes")
	}
	if !record.NonTransferable {
		return errors.New("stake reputation must be non-transferable")
	}
	if record.ClaimedStakeWeightedSeconds > record.StakeWeightedSeconds {
		return errors.New("claimed stake reputation seconds exceed accumulator")
	}
	if record.ClaimedStakeReputation > StakeReputationMaxScore {
		return errors.New("claimed stake reputation exceeds max score")
	}
	if err := validateStakePoolExposures(record.PoolExposures); err != nil {
		return err
	}
	return nil
}

func validateStakePoolExposures(exposures []StakePoolExposure) error {
	seen := map[string]struct{}{}
	previous := ""
	for i, exposure := range exposures {
		if strings.TrimSpace(exposure.PoolID) == "" {
			return errors.New("stake reputation pool id is required")
		}
		if _, found := seen[exposure.PoolID]; found {
			return errors.New("duplicate stake reputation pool exposure")
		}
		seen[exposure.PoolID] = struct{}{}
		if i > 0 && previous >= exposure.PoolID {
			return errors.New("stake reputation pool exposures must be sorted deterministically")
		}
		previous = exposure.PoolID
		if exposure.Shares > 0 && exposure.TotalPoolShares == 0 {
			return errors.New("stake reputation total pool shares must be positive when shares are positive")
		}
		if exposure.Shares > exposure.TotalPoolShares && exposure.TotalPoolShares > 0 {
			return errors.New("stake reputation shares cannot exceed total pool shares")
		}
		if exposure.ValidatorBonusBps > 10_000 {
			return errors.New("stake reputation validator bonus bps exceeds basis points")
		}
		if (exposure.ValidatorJailed || exposure.ValidatorSlashed) && exposure.ValidatorBonusBps > 0 {
			return errors.New("slashed or jailed validator cannot receive positive stake reputation bonus")
		}
		if (exposure.ValidatorJailed || exposure.ValidatorSlashed) && !exposure.ValidatorBonusBlocked {
			return errors.New("slashed or jailed validator bonus must be marked blocked")
		}
	}
	return nil
}

func validateScores(scores []ReputationScore) error {
	seen := make(map[string]struct{}, len(scores))
	var previous string
	for i, score := range scores {
		if err := addressing.RejectZeroAddress("reputation snapshot account", score.Account); err != nil {
			return err
		}
		if score.Score > ScoreMax {
			return errors.New("reputation snapshot score exceeds max")
		}
		key := addressKey(score.Account)
		if _, found := seen[key]; found {
			return errors.New("duplicate reputation snapshot score")
		}
		seen[key] = struct{}{}
		if i > 0 && previous >= key {
			return errors.New("reputation snapshot scores must be sorted deterministically")
		}
		previous = key
	}
	return nil
}

func scoresFromRecords(records []ReputationRecord) []ReputationScore {
	records = normalizeRecords(records)
	scores := make([]ReputationScore, len(records))
	for i, record := range records {
		scores[i] = ReputationScore{Account: cloneAddress(record.Account), Score: record.Score}
	}
	return scores
}

func snapshotContains(snapshot ReputationSnapshot, subjectType string, subject sdk.AccAddress) bool {
	key := addressKey(subject)
	var scores []ReputationScore
	if subjectType == SubjectValidator {
		scores = snapshot.ValidatorScores
	} else {
		scores = snapshot.ReporterScores
	}
	for _, score := range scores {
		if addressKey(score.Account) == key {
			return true
		}
	}
	return false
}

func upsertRecord(records []ReputationRecord, record ReputationRecord) []ReputationRecord {
	key := addressKey(record.Account)
	next := make([]ReputationRecord, 0, len(records)+1)
	replaced := false
	for _, existing := range records {
		if addressKey(existing.Account) == key {
			next = append(next, cloneRecord(record))
			replaced = true
			continue
		}
		next = append(next, cloneRecord(existing))
	}
	if !replaced {
		next = append(next, cloneRecord(record))
	}
	return normalizeRecords(next)
}

func normalizeRecords(records []ReputationRecord) []ReputationRecord {
	out := make([]ReputationRecord, len(records))
	for i, record := range records {
		out[i] = cloneRecord(record)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return addressKey(out[i].Account) < addressKey(out[j].Account)
	})
	return out
}

func normalizeStakeReputationRecords(records []StakeReputationRecord) []StakeReputationRecord {
	out := make([]StakeReputationRecord, len(records))
	for i, record := range records {
		out[i] = normalizeStakeReputationRecord(record)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return addressKey(out[i].Account) < addressKey(out[j].Account)
	})
	return out
}

func normalizeStakeReputationRecord(record StakeReputationRecord) StakeReputationRecord {
	record.Account = cloneAddress(record.Account)
	if len(record.Account) > 0 {
		record.AccountUser = addressing.FormatAccAddress(record.Account)
	}
	record.NonTransferable = true
	record.PoolExposures = normalizeStakePoolExposures(record.PoolExposures)
	return record
}

func normalizeStakePoolExposures(exposures []StakePoolExposure) []StakePoolExposure {
	out := append([]StakePoolExposure(nil), exposures...)
	for i := range out {
		out[i].PoolID = strings.TrimSpace(out[i].PoolID)
		if out[i].ValidatorJailed || out[i].ValidatorSlashed {
			out[i].ValidatorBonusBps = 0
			out[i].ValidatorBonusBlocked = true
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].PoolID < out[j].PoolID })
	return out
}

func normalizeSnapshots(snapshots []ReputationSnapshot) []ReputationSnapshot {
	out := make([]ReputationSnapshot, len(snapshots))
	for i, snapshot := range snapshots {
		snapshot.ValidatorScores = normalizeScores(snapshot.ValidatorScores)
		snapshot.ReporterScores = normalizeScores(snapshot.ReporterScores)
		if snapshot.SnapshotHash == "" {
			snapshot.SnapshotHash = ComputeReputationSnapshotHash(snapshot)
		}
		out[i] = snapshot
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Epoch < out[j].Epoch
	})
	return out
}

func normalizeScores(scores []ReputationScore) []ReputationScore {
	out := make([]ReputationScore, len(scores))
	for i, score := range scores {
		out[i] = ReputationScore{Account: cloneAddress(score.Account), Score: score.Score}
	}
	sort.SliceStable(out, func(i, j int) bool {
		return addressKey(out[i].Account) < addressKey(out[j].Account)
	})
	return out
}

func normalizeEvents(events []ReputationEvent) []ReputationEvent {
	out := make([]ReputationEvent, len(events))
	for i, event := range events {
		event.SubjectType = strings.TrimSpace(event.SubjectType)
		event.Subject = cloneAddress(event.Subject)
		event.Component = strings.TrimSpace(event.Component)
		event.Reason = strings.TrimSpace(event.Reason)
		if event.EventID == "" {
			event.EventID = ComputeReputationEventID(event)
		}
		if event.EventHash == "" {
			event.EventHash = ComputeReputationEventHash(event)
		}
		out[i] = event
	}
	sortEvents(out)
	return out
}

func sortEvents(events []ReputationEvent) {
	sort.SliceStable(events, func(i, j int) bool {
		if events[i].Epoch != events[j].Epoch {
			return events[i].Epoch < events[j].Epoch
		}
		return events[i].EventID < events[j].EventID
	})
}

func trimEvents(events []ReputationEvent, limit uint32) []ReputationEvent {
	events = normalizeEvents(events)
	if limit == 0 || uint32(len(events)) <= limit {
		return events
	}
	return append([]ReputationEvent(nil), events[len(events)-int(limit):]...)
}

func trimSnapshots(snapshots []ReputationSnapshot, limit uint32) []ReputationSnapshot {
	snapshots = normalizeSnapshots(snapshots)
	if limit == 0 || uint32(len(snapshots)) <= limit {
		return snapshots
	}
	return append([]ReputationSnapshot(nil), snapshots[len(snapshots)-int(limit):]...)
}

func cloneReputationState(state ReputationState) ReputationState {
	state = NormalizeReputationState(state)
	return ReputationState{
		Params:		state.Params,
		Accounts:	append([]ReputationRecord(nil), state.Accounts...),
		Validators:	append([]ReputationRecord(nil), state.Validators...),
		Reporters:	append([]ReputationRecord(nil), state.Reporters...),
		StakeRecords:	cloneStakeReputationRecords(state.StakeRecords),
		Snapshots:	append([]ReputationSnapshot(nil), state.Snapshots...),
		PenaltyEvents:	append([]ReputationEvent(nil), state.PenaltyEvents...),
		RecoveryEvents:	append([]ReputationEvent(nil), state.RecoveryEvents...),
	}
}

func cloneRecord(record ReputationRecord) ReputationRecord {
	record.Account = cloneAddress(record.Account)
	return record
}

func cloneStakeReputationRecords(records []StakeReputationRecord) []StakeReputationRecord {
	out := make([]StakeReputationRecord, len(records))
	for i, record := range records {
		out[i] = cloneStakeReputationRecord(record)
	}
	return out
}

func cloneStakeReputationRecord(record StakeReputationRecord) StakeReputationRecord {
	record.Account = cloneAddress(record.Account)
	record.PoolExposures = append([]StakePoolExposure(nil), record.PoolExposures...)
	return record
}

func cloneAddress(address sdk.AccAddress) sdk.AccAddress {
	return append(sdk.AccAddress(nil), address...)
}

func addressKey(address sdk.AccAddress) string {
	return hex.EncodeToString(address)
}

func ComputeReputationEventID(event ReputationEvent) string {
	return hashParts(
		"reputation-event-id-v1",
		event.SubjectType,
		addressKey(event.Subject),
		event.Component,
		fmt.Sprint(event.Amount),
		event.Reason,
		fmt.Sprint(event.Epoch),
		fmt.Sprint(event.ScoreBefore),
		fmt.Sprint(event.ScoreAfter),
	)
}

func ComputeReputationEventHash(event ReputationEvent) string {
	return hashParts(
		"reputation-event-v1",
		event.EventID,
		event.SubjectType,
		addressKey(event.Subject),
		event.Component,
		fmt.Sprint(event.Amount),
		event.Reason,
		fmt.Sprint(event.Epoch),
		fmt.Sprint(event.ScoreBefore),
		fmt.Sprint(event.ScoreAfter),
	)
}

func ComputeReputationSnapshotHash(snapshot ReputationSnapshot) string {
	snapshot.ValidatorScores = normalizeScores(snapshot.ValidatorScores)
	snapshot.ReporterScores = normalizeScores(snapshot.ReporterScores)
	parts := []string{"reputation-snapshot-v1", fmt.Sprint(snapshot.Epoch)}
	for _, score := range snapshot.ValidatorScores {
		parts = append(parts, "validator", addressKey(score.Account), fmt.Sprint(score.Score))
	}
	for _, score := range snapshot.ReporterScores {
		parts = append(parts, "reporter", addressKey(score.Account), fmt.Sprint(score.Score))
	}
	return hashParts(parts...)
}

func ComputeStakeReputationClaimHash(record StakeReputationRecord, delta uint16, accountScoreAfter uint8) string {
	record = normalizeStakeReputationRecord(record)
	parts := []string{
		"stake-reputation-claim-v1",
		addressKey(record.Account),
		record.AccountUser,
		fmt.Sprint(record.StakeWeightedSeconds),
		fmt.Sprint(record.ClaimedStakeWeightedSeconds),
		fmt.Sprint(record.ClaimedStakeReputation),
		fmt.Sprint(delta),
		fmt.Sprint(accountScoreAfter),
	}
	for _, exposure := range record.PoolExposures {
		parts = append(parts,
			exposure.PoolID,
			fmt.Sprint(exposure.Shares),
			fmt.Sprint(exposure.TotalPoolShares),
			fmt.Sprint(exposure.PoolActiveStake),
			fmt.Sprint(exposure.EffectiveStake),
			fmt.Sprint(exposure.LastUpdatedUnix),
			fmt.Sprint(exposure.ValidatorJailed),
			fmt.Sprint(exposure.ValidatorSlashed),
			fmt.Sprint(exposure.ValidatorOperator),
			fmt.Sprint(exposure.ValidatorBonusBps),
			fmt.Sprint(exposure.ValidatorBonusBlocked),
		)
	}
	return hashParts(parts...)
}

func hashParts(parts ...string) string {
	h := sha256.New()
	for _, part := range parts {
		data := []byte(part)
		var lenBuf [8]byte
		for i := uint(0); i < 8; i++ {
			lenBuf[7-i] = byte(uint64(len(data)) >> (i * 8))
		}
		h.Write(lenBuf[:])
		h.Write(data)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func isHexHash(value string) bool {
	if len(value) != 64 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func saturatingAddU16(a, b uint16) uint16 {
	sum := uint32(a) + uint32(b)
	if sum > uint32(^uint16(0)) {
		return ^uint16(0)
	}
	return uint16(sum)
}

func checkedAddUint64(a, b uint64) (uint64, error) {
	if ^uint64(0)-a < b {
		return 0, errors.New("reputation stake-time accumulator overflow")
	}
	return a + b, nil
}

func stakeWeightedSeconds(stake, duration uint64) (uint64, error) {
	if stake == 0 || duration == 0 {
		return 0, nil
	}
	if stake > ^uint64(0)/duration {
		return 0, errors.New("reputation stake-time accumulator overflow")
	}
	return stake * duration, nil
}

func mulDivFloor(value, multiplier, denominator uint64) uint64 {
	if value == 0 || multiplier == 0 || denominator == 0 {
		return 0
	}
	if value > ^uint64(0)/multiplier {
		return ^uint64(0)
	}
	return value * multiplier / denominator
}

func maxUint64(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

func findStakePoolExposure(exposures []StakePoolExposure, poolID string) (int, StakePoolExposure, bool) {
	poolID = strings.TrimSpace(poolID)
	for idx, exposure := range exposures {
		if exposure.PoolID == poolID {
			return idx, exposure, true
		}
	}
	return -1, StakePoolExposure{}, false
}

func accountScore(state ReputationState, account sdk.AccAddress) uint8 {
	record, found := state.recordBySubject(SubjectAccount, account)
	if !found {
		return ScoreMin
	}
	return record.Score
}

func minU16(a, b uint16) uint16 {
	if a < b {
		return a
	}
	return b
}
