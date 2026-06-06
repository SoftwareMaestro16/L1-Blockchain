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
	ScoreMin = uint8(0)
	ScoreMax = uint8(100)

	LevelRestricted = "restricted"
	LevelNew        = "new"
	LevelNormal     = "normal"
	LevelTrusted    = "trusted"
	LevelElite      = "elite"

	MaxDomainScore   = uint16(10)
	MaxContractScore = uint16(15)

	DefaultReputationAuthority = "4:0000000000000000000000000000000000000000000000000000000000000001"

	SubjectValidator = "validator"
	SubjectReporter  = "reporter"

	ComponentMissedBlock = "missed_block"
	ComponentSlashing    = "slashing"
	ComponentSpam        = "spam"
	ComponentUptime      = "uptime"
	ComponentRecovery    = "recovery"
	ComponentVolume      = "volume"
)

type ReputationRecord struct {
	Account          sdk.AccAddress
	Score            uint8
	AgeScore         uint16
	StakingScore     uint16
	TxSuccessScore   uint16
	VolumeScore      uint16
	DomainScore      uint16
	ContractScore    uint16
	SpamPenalty      uint16
	FailedTxPenalty  uint16
	SlashPenalty     uint16
	LastUpdatedEpoch uint64
}

type DecayParams struct {
	InactiveAfterEpochs uint64
	DecayRatePerEpoch   uint8
}

type ReputationParams struct {
	Authority            string
	MinScore             uint8
	MaxScore             uint8
	MissedBlockPenalty   uint16
	SlashingPenalty      uint16
	UptimeReward         uint16
	RecoveryReward       uint16
	SlashingReducesScore bool
	Decay                DecayParams
	MaxHistorySnapshots  uint32
	MaxEvents            uint32
}

type ReputationEvent struct {
	EventID     string
	SubjectType string
	Subject     sdk.AccAddress
	Component   string
	Amount      uint16
	Reason      string
	Epoch       uint64
	ScoreBefore uint8
	ScoreAfter  uint8
	EventHash   string
}

type ReputationSnapshot struct {
	Epoch           uint64
	ValidatorScores []ReputationScore
	ReporterScores  []ReputationScore
	SnapshotHash    string
}

type ReputationScore struct {
	Account sdk.AccAddress
	Score   uint8
}

type ReputationState struct {
	Params         ReputationParams
	Validators     []ReputationRecord
	Reporters      []ReputationRecord
	Snapshots      []ReputationSnapshot
	PenaltyEvents  []ReputationEvent
	RecoveryEvents []ReputationEvent
}

type MsgUpdateReputationParams struct {
	Authority string
	Params    ReputationParams
}

type MsgApplyReputationPenalty struct {
	Authority   string
	SubjectType string
	Subject     sdk.AccAddress
	Component   string
	Amount      uint16
	Reason      string
	Epoch       uint64
}

type MsgApplyReputationReward struct {
	Authority   string
	SubjectType string
	Subject     sdk.AccAddress
	Component   string
	Amount      uint16
	Reason      string
	Epoch       uint64
}

type MsgRecomputeReputation struct {
	Authority   string
	SubjectType string
	Subject     sdk.AccAddress
	Epoch       uint64
}

type ReputationHistoryQuery struct {
	SubjectType string
	Subject     sdk.AccAddress
	Limit       uint32
}

type ProgressiveLimits struct {
	MaxTxsPerBlock uint32
	MaxTxGas       uint64
	MaxQueueMsgs   uint32
}

func DefaultReputationParams() ReputationParams {
	return ReputationParams{
		Authority:            DefaultReputationAuthority,
		MinScore:             ScoreMin,
		MaxScore:             ScoreMax,
		MissedBlockPenalty:   5,
		SlashingPenalty:      25,
		UptimeReward:         2,
		RecoveryReward:       3,
		SlashingReducesScore: true,
		Decay: DecayParams{
			InactiveAfterEpochs: 10,
			DecayRatePerEpoch:   1,
		},
		MaxHistorySnapshots: 1024,
		MaxEvents:           4096,
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
	return nil
}

func (state ReputationState) Validate() error {
	state = NormalizeReputationState(state)
	if err := state.Params.Validate(); err != nil {
		return err
	}
	if err := validateReputationRecords("validator", state.Validators); err != nil {
		return err
	}
	if err := validateReputationRecords("reporter", state.Reporters); err != nil {
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
		Epoch:           epoch,
		ValidatorScores: scoresFromRecords(state.Validators),
		ReporterScores:  scoresFromRecords(state.Reporters),
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
		SubjectType: subjectType,
		Subject:     cloneAddress(subject),
		Component:   strings.TrimSpace(component),
		Amount:      amount,
		Reason:      strings.TrimSpace(reason),
		Epoch:       epoch,
		ScoreBefore: before,
		ScoreAfter:  after,
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
	state.Validators = normalizeRecords(state.Validators)
	state.Reporters = normalizeRecords(state.Reporters)
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

func (state ReputationState) recordBySubject(subjectType string, subject sdk.AccAddress) (ReputationRecord, bool) {
	key := addressKey(subject)
	switch subjectType {
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
	case SubjectValidator:
		state.Validators = upsertRecord(state.Validators, record)
	case SubjectReporter:
		state.Reporters = upsertRecord(state.Reporters, record)
	}
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
	case SubjectValidator, SubjectReporter:
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
	case ComponentMissedBlock, ComponentSlashing, ComponentSpam, ComponentUptime, ComponentRecovery, ComponentVolume:
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
		Params:         state.Params,
		Validators:     append([]ReputationRecord(nil), state.Validators...),
		Reporters:      append([]ReputationRecord(nil), state.Reporters...),
		Snapshots:      append([]ReputationSnapshot(nil), state.Snapshots...),
		PenaltyEvents:  append([]ReputationEvent(nil), state.PenaltyEvents...),
		RecoveryEvents: append([]ReputationEvent(nil), state.RecoveryEvents...),
	}
}

func cloneRecord(record ReputationRecord) ReputationRecord {
	record.Account = cloneAddress(record.Account)
	return record
}

func cloneAddress(address sdk.AccAddress) sdk.AccAddress {
	return append(sdk.AccAddress(nil), address...)
}

func addressKey(address sdk.AccAddress) string {
	return hex.EncodeToString(address)
}

func ComputeReputationEventID(event ReputationEvent) string {
	event.EventID = ""
	event.EventHash = ""
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

func minU16(a, b uint16) uint16 {
	if a < b {
		return a
	}
	return b
}
