package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	postypes "github.com/sovereign-l1/l1/x/pos/types"
)

const (
	DefaultPerformanceOracleAuthority	= "4:0000000000000000000000000000000000000000000000000000000000000001"

	ReportSourceValidator	= "validator"
	ReportSourceObserver	= "observer"

	ChallengeStatusOpen	= "open"
	ChallengeStatusAccepted	= "accepted"
	ChallengeStatusRejected	= "rejected"
)

type PerformanceOracleParams struct {
	Authority			string
	MinScoreBps			uint32
	MaxScoreBps			uint32
	MaxLatencyMillis		uint64
	MaxResponseTimeMillis		uint64
	MaxReportsPerEpoch		uint32
	MaxChallengesPerEpoch		uint32
	MinReportsForAggregation	uint32
	OutlierTrimBps			uint32
	SlashableSourcesRequired	bool
}

type PerformanceReport struct {
	ReportID		string
	Epoch			uint64
	ValidatorAddress	string
	ReporterAddress		string
	Source			string
	UptimeSignedBlocks	uint64
	UptimeTotalBlocks	uint64
	LatencyMillis		uint64
	ResponseTimeMillis	uint64
	MissedBlocks		uint64
	MissedWindowBlocks	uint64
	PeerScoreBps		uint32
	SubmittedHeight		uint64
	Slashable		bool
	Challenged		bool
	ReportHash		string
}

type PerformanceAggregate struct {
	Epoch			uint64
	ValidatorAddress	string
	ReportCount		uint32
	UptimeScoreBps		uint32
	LatencyScoreBps		uint32
	ResponseTimeScoreBps	uint32
	MissedBlockScoreBps	uint32
	PeerScoreBps		uint32
	PerformanceScoreBps	uint32
	AggregationHash		string
}

type PerformanceEpoch struct {
	Epoch		uint64
	Finalized	bool
	Reports		[]PerformanceReport
	Aggregates	[]PerformanceAggregate
	Challenges	[]PerformanceChallenge
	EpochHash	string
}

type PerformanceChallenge struct {
	ChallengeID	string
	ReportID	string
	Challenger	string
	Reason		string
	Epoch		uint64
	Accepted	bool
	Status		string
	ChallengeHash	string
}

type PerformanceOracleState struct {
	Params			PerformanceOracleParams
	AggregationEpoch	uint64
	Epochs			[]PerformanceEpoch
}

type MsgSubmitPerformanceReport struct {
	Authority	string
	Report		PerformanceReport
}

type MsgFinalizePerformanceEpoch struct {
	Authority	string
	Epoch		uint64
}

type MsgChallengePerformanceReport struct {
	Authority	string
	Epoch		uint64
	ReportID	string
	Challenger	string
	Reason		string
	Accepted	bool
}

type QueryValidatorPerformanceRequest struct {
	Epoch			uint64
	ValidatorAddress	string
}

type QueryValidatorPerformanceResponse struct {
	Aggregate PerformanceAggregate
}

type QueryPerformanceEpochRequest struct {
	Epoch uint64
}

type QueryPerformanceEpochResponse struct {
	Epoch PerformanceEpoch
}

type QueryPerformanceReportsRequest struct {
	Epoch			uint64
	ValidatorAddress	string
}

type QueryPerformanceReportsResponse struct {
	Reports []PerformanceReport
}

type QueryPerformanceParamsResponse struct {
	Params PerformanceOracleParams
}

func DefaultPerformanceOracleParams() PerformanceOracleParams {
	return PerformanceOracleParams{
		Authority:			DefaultPerformanceOracleAuthority,
		MinScoreBps:			0,
		MaxScoreBps:			postypes.BasisPoints,
		MaxLatencyMillis:		5_000,
		MaxResponseTimeMillis:		10_000,
		MaxReportsPerEpoch:		10_000,
		MaxChallengesPerEpoch:		10_000,
		MinReportsForAggregation:	1,
		OutlierTrimBps:			1_000,
		SlashableSourcesRequired:	true,
	}
}

func NewPerformanceOracleState(params PerformanceOracleParams) (PerformanceOracleState, error) {
	if strings.TrimSpace(params.Authority) == "" {
		params = DefaultPerformanceOracleParams()
	}
	if err := params.Validate(); err != nil {
		return PerformanceOracleState{}, err
	}
	return PerformanceOracleState{Params: params}, nil
}

func (params PerformanceOracleParams) Validate() error {
	if strings.TrimSpace(params.Authority) == "" {
		return errors.New("performance oracle authority is required")
	}
	if params.MinScoreBps != 0 {
		return errors.New("performance oracle min score must be zero")
	}
	if params.MaxScoreBps == 0 || params.MaxScoreBps > postypes.BasisPoints {
		return errors.New("performance oracle max score must be between 1 and 10000 bps")
	}
	if params.MaxLatencyMillis == 0 {
		return errors.New("performance oracle max latency must be positive")
	}
	if params.MaxResponseTimeMillis == 0 {
		return errors.New("performance oracle max response time must be positive")
	}
	if params.MaxReportsPerEpoch == 0 {
		return errors.New("performance oracle max reports per epoch must be positive")
	}
	if params.MaxChallengesPerEpoch == 0 {
		return errors.New("performance oracle max challenges per epoch must be positive")
	}
	if params.MinReportsForAggregation == 0 {
		return errors.New("performance oracle min reports for aggregation must be positive")
	}
	if params.OutlierTrimBps > 4_900 {
		return errors.New("performance oracle outlier trim must be <= 4900 bps")
	}
	return nil
}

func (state PerformanceOracleState) Validate() error {
	state = NormalizePerformanceOracleState(state)
	if err := state.Params.Validate(); err != nil {
		return err
	}
	for _, epoch := range state.Epochs {
		if err := epoch.Validate(state.Params); err != nil {
			return err
		}
	}
	return nil
}

func ApplySubmitPerformanceReport(state PerformanceOracleState, msg MsgSubmitPerformanceReport) (PerformanceOracleState, error) {
	state = NormalizePerformanceOracleState(state)
	if err := authorizePerformanceOracle(state.Params, msg.Authority); err != nil {
		return PerformanceOracleState{}, err
	}
	report := normalizePerformanceReport(msg.Report)
	if err := report.Validate(state.Params); err != nil {
		return PerformanceOracleState{}, err
	}
	if state.Params.SlashableSourcesRequired && !report.Slashable {
		return PerformanceOracleState{}, errors.New("performance report source must be slashable")
	}
	epoch := state.epochByID(report.Epoch)
	if epoch.Finalized {
		return PerformanceOracleState{}, errors.New("performance epoch is finalized")
	}
	if uint32(len(epoch.Reports)) >= state.Params.MaxReportsPerEpoch {
		return PerformanceOracleState{}, errors.New("performance report limit reached")
	}
	for _, existing := range epoch.Reports {
		if existing.ReportID == report.ReportID {
			return PerformanceOracleState{}, errors.New("duplicate performance report")
		}
	}
	epoch.Reports = append(epoch.Reports, report)
	epoch = finalizeEpochHash(epoch)
	state = state.upsertEpoch(epoch)
	state.AggregationEpoch = maxU64(state.AggregationEpoch, report.Epoch)
	return NormalizePerformanceOracleState(state), state.Validate()
}

func ApplyChallengePerformanceReport(state PerformanceOracleState, msg MsgChallengePerformanceReport) (PerformanceOracleState, error) {
	state = NormalizePerformanceOracleState(state)
	if err := authorizePerformanceOracle(state.Params, msg.Authority); err != nil {
		return PerformanceOracleState{}, err
	}
	if msg.Epoch == 0 {
		return PerformanceOracleState{}, errors.New("performance challenge epoch is required")
	}
	if strings.TrimSpace(msg.ReportID) == "" {
		return PerformanceOracleState{}, errors.New("performance challenge report id is required")
	}
	if strings.TrimSpace(msg.Challenger) == "" {
		return PerformanceOracleState{}, errors.New("performance challenge challenger is required")
	}
	if strings.TrimSpace(msg.Reason) == "" {
		return PerformanceOracleState{}, errors.New("performance challenge reason is required")
	}
	epoch, found := state.findEpochByID(msg.Epoch)
	if !found {
		return PerformanceOracleState{}, errors.New("performance epoch not found")
	}
	if uint32(len(epoch.Challenges)) >= state.Params.MaxChallengesPerEpoch {
		return PerformanceOracleState{}, errors.New("performance challenge limit reached")
	}
	reportFound := false
	for i := range epoch.Reports {
		if epoch.Reports[i].ReportID == msg.ReportID {
			epoch.Reports[i].Challenged = true
			epoch.Reports[i].ReportHash = ComputePerformanceReportHash(epoch.Reports[i])
			reportFound = true
			break
		}
	}
	if !reportFound {
		return PerformanceOracleState{}, errors.New("performance report not found")
	}
	challenge := NewPerformanceChallenge(msg)
	epoch.Challenges = append(epoch.Challenges, challenge)
	epoch = finalizeEpochHash(epoch)
	state = state.upsertEpoch(epoch)
	return NormalizePerformanceOracleState(state), state.Validate()
}

func ApplyFinalizePerformanceEpoch(state PerformanceOracleState, msg MsgFinalizePerformanceEpoch) (PerformanceOracleState, error) {
	state = NormalizePerformanceOracleState(state)
	if err := authorizePerformanceOracle(state.Params, msg.Authority); err != nil {
		return PerformanceOracleState{}, err
	}
	if msg.Epoch == 0 {
		return PerformanceOracleState{}, errors.New("performance finalize epoch is required")
	}
	epoch, found := state.findEpochByID(msg.Epoch)
	if !found {
		return PerformanceOracleState{}, errors.New("performance epoch not found")
	}
	if epoch.Finalized {
		return PerformanceOracleState{}, errors.New("performance epoch already finalized")
	}
	aggregates, err := AggregatePerformanceEpoch(epoch.Reports, state.Params)
	if err != nil {
		return PerformanceOracleState{}, err
	}
	epoch.Aggregates = aggregates
	epoch.Finalized = true
	epoch = finalizeEpochHash(epoch)
	state = state.upsertEpoch(epoch)
	state.AggregationEpoch = maxU64(state.AggregationEpoch, msg.Epoch)
	return NormalizePerformanceOracleState(state), state.Validate()
}

func AggregatePerformanceEpoch(reports []PerformanceReport, params PerformanceOracleParams) ([]PerformanceAggregate, error) {
	params.Authority = strings.TrimSpace(params.Authority)
	if params.Authority == "" {
		params = DefaultPerformanceOracleParams()
	}
	if err := params.Validate(); err != nil {
		return nil, err
	}
	reports = normalizeReports(reports)
	byValidator := make(map[string][]PerformanceReport)
	for _, report := range reports {
		if err := report.Validate(params); err != nil {
			return nil, err
		}
		if report.Challenged {
			continue
		}
		byValidator[report.ValidatorAddress] = append(byValidator[report.ValidatorAddress], report)
	}
	validators := make([]string, 0, len(byValidator))
	for validator := range byValidator {
		validators = append(validators, validator)
	}
	sort.Strings(validators)
	aggregates := make([]PerformanceAggregate, 0, len(validators))
	for _, validator := range validators {
		group := byValidator[validator]
		if uint32(len(group)) < params.MinReportsForAggregation {
			continue
		}
		aggregate := aggregateValidatorReports(validator, group, params)
		aggregate.AggregationHash = ComputePerformanceAggregateHash(aggregate)
		if err := aggregate.Validate(params); err != nil {
			return nil, err
		}
		aggregates = append(aggregates, aggregate)
	}
	return aggregates, nil
}

func QueryValidatorPerformanceOracle(state PerformanceOracleState, req QueryValidatorPerformanceRequest) (QueryValidatorPerformanceResponse, error) {
	state = NormalizePerformanceOracleState(state)
	epoch, found := state.findEpochByID(req.Epoch)
	if !found {
		return QueryValidatorPerformanceResponse{}, errors.New("performance epoch not found")
	}
	validator := strings.TrimSpace(req.ValidatorAddress)
	if validator == "" {
		return QueryValidatorPerformanceResponse{}, errors.New("performance validator address is required")
	}
	for _, aggregate := range epoch.Aggregates {
		if aggregate.ValidatorAddress == validator {
			return QueryValidatorPerformanceResponse{Aggregate: aggregate}, nil
		}
	}
	return QueryValidatorPerformanceResponse{}, errors.New("performance aggregate not found")
}

func QueryPerformanceEpochOracle(state PerformanceOracleState, req QueryPerformanceEpochRequest) (QueryPerformanceEpochResponse, error) {
	state = NormalizePerformanceOracleState(state)
	epoch, found := state.findEpochByID(req.Epoch)
	if !found {
		return QueryPerformanceEpochResponse{}, errors.New("performance epoch not found")
	}
	return QueryPerformanceEpochResponse{Epoch: epoch}, nil
}

func QueryPerformanceReportsOracle(state PerformanceOracleState, req QueryPerformanceReportsRequest) (QueryPerformanceReportsResponse, error) {
	state = NormalizePerformanceOracleState(state)
	epoch, found := state.findEpochByID(req.Epoch)
	if !found {
		return QueryPerformanceReportsResponse{}, errors.New("performance epoch not found")
	}
	validator := strings.TrimSpace(req.ValidatorAddress)
	reports := make([]PerformanceReport, 0, len(epoch.Reports))
	for _, report := range epoch.Reports {
		if validator == "" || report.ValidatorAddress == validator {
			reports = append(reports, report)
		}
	}
	return QueryPerformanceReportsResponse{Reports: normalizeReports(reports)}, nil
}

func QueryPerformanceParamsOracle(state PerformanceOracleState) QueryPerformanceParamsResponse {
	return QueryPerformanceParamsResponse{Params: state.Params}
}

func ExportPerformanceOracleState(state PerformanceOracleState) (PerformanceOracleState, error) {
	state = NormalizePerformanceOracleState(state)
	if err := state.Validate(); err != nil {
		return PerformanceOracleState{}, err
	}
	return clonePerformanceOracleState(state), nil
}

func ImportPerformanceOracleState(exported PerformanceOracleState) (PerformanceOracleState, error) {
	exported = NormalizePerformanceOracleState(exported)
	if err := exported.Validate(); err != nil {
		return PerformanceOracleState{}, err
	}
	return clonePerformanceOracleState(exported), nil
}

func CheckPerformanceOracleInvariants(state PerformanceOracleState) error {
	state = NormalizePerformanceOracleState(state)
	return state.Validate()
}

func NewPerformanceReport(report PerformanceReport) (PerformanceReport, error) {
	report = normalizePerformanceReport(report)
	if report.ReportID == "" {
		report.ReportID = ComputePerformanceReportID(report)
	}
	if report.ReportHash == "" {
		report.ReportHash = ComputePerformanceReportHash(report)
	}
	if err := report.Validate(DefaultPerformanceOracleParams()); err != nil {
		return PerformanceReport{}, err
	}
	return report, nil
}

func (report PerformanceReport) Validate(params PerformanceOracleParams) error {
	report = normalizePerformanceReport(report)
	if report.Epoch == 0 {
		return errors.New("performance report epoch is required")
	}
	if report.ValidatorAddress == "" {
		return errors.New("performance report validator is required")
	}
	if report.ReporterAddress == "" {
		return errors.New("performance report reporter is required")
	}
	if report.Source != ReportSourceValidator && report.Source != ReportSourceObserver {
		return fmt.Errorf("unknown performance report source %q", report.Source)
	}
	if params.SlashableSourcesRequired && !report.Slashable {
		return errors.New("performance report source must be slashable")
	}
	if report.UptimeTotalBlocks == 0 {
		return errors.New("performance report uptime window is required")
	}
	if report.UptimeSignedBlocks > report.UptimeTotalBlocks {
		return errors.New("performance report signed blocks exceed uptime window")
	}
	if report.MissedBlocks > report.MissedWindowBlocks {
		return errors.New("performance report missed blocks exceed window")
	}
	if report.LatencyMillis > params.MaxLatencyMillis {
		return errors.New("performance report latency exceeds configured max")
	}
	if report.ResponseTimeMillis > params.MaxResponseTimeMillis {
		return errors.New("performance report response time exceeds configured max")
	}
	if report.PeerScoreBps > params.MaxScoreBps {
		return errors.New("performance report peer score exceeds max")
	}
	if report.SubmittedHeight == 0 {
		return errors.New("performance report submitted height is required")
	}
	if !isHex64(report.ReportID) {
		return errors.New("performance report id must be a hex hash")
	}
	if !isHex64(report.ReportHash) {
		return errors.New("performance report hash must be a hex hash")
	}
	if report.ReportID != ComputePerformanceReportID(report) {
		return errors.New("performance report id mismatch")
	}
	if report.ReportHash != ComputePerformanceReportHash(report) {
		return errors.New("performance report hash mismatch")
	}
	return nil
}

func (aggregate PerformanceAggregate) Validate(params PerformanceOracleParams) error {
	if aggregate.Epoch == 0 {
		return errors.New("performance aggregate epoch is required")
	}
	if strings.TrimSpace(aggregate.ValidatorAddress) == "" {
		return errors.New("performance aggregate validator is required")
	}
	if aggregate.ReportCount == 0 {
		return errors.New("performance aggregate report count is required")
	}
	for _, score := range []uint32{
		aggregate.UptimeScoreBps,
		aggregate.LatencyScoreBps,
		aggregate.ResponseTimeScoreBps,
		aggregate.MissedBlockScoreBps,
		aggregate.PeerScoreBps,
		aggregate.PerformanceScoreBps,
	} {
		if score < params.MinScoreBps || score > params.MaxScoreBps {
			return errors.New("performance aggregate score exceeds configured bounds")
		}
	}
	if !isHex64(aggregate.AggregationHash) {
		return errors.New("performance aggregate hash must be a hex hash")
	}
	if aggregate.AggregationHash != ComputePerformanceAggregateHash(aggregate) {
		return errors.New("performance aggregate hash mismatch")
	}
	return nil
}

func (epoch PerformanceEpoch) Validate(params PerformanceOracleParams) error {
	epoch = normalizeEpoch(epoch)
	if epoch.Epoch == 0 {
		return errors.New("performance epoch is required")
	}
	if uint32(len(epoch.Reports)) > params.MaxReportsPerEpoch {
		return errors.New("performance epoch report limit exceeded")
	}
	if uint32(len(epoch.Challenges)) > params.MaxChallengesPerEpoch {
		return errors.New("performance epoch challenge limit exceeded")
	}
	for _, report := range epoch.Reports {
		if report.Epoch != epoch.Epoch {
			return errors.New("performance report epoch mismatch")
		}
		if err := report.Validate(params); err != nil {
			return err
		}
	}
	for _, aggregate := range epoch.Aggregates {
		if aggregate.Epoch != epoch.Epoch {
			return errors.New("performance aggregate epoch mismatch")
		}
		if err := aggregate.Validate(params); err != nil {
			return err
		}
	}
	for _, challenge := range epoch.Challenges {
		if challenge.Epoch != epoch.Epoch {
			return errors.New("performance challenge epoch mismatch")
		}
		if err := challenge.Validate(); err != nil {
			return err
		}
	}
	if epoch.EpochHash == "" || !isHex64(epoch.EpochHash) {
		return errors.New("performance epoch hash must be a hex hash")
	}
	if epoch.EpochHash != ComputePerformanceEpochHash(epoch) {
		return errors.New("performance epoch hash mismatch")
	}
	return nil
}

func NewPerformanceChallenge(msg MsgChallengePerformanceReport) PerformanceChallenge {
	challenge := PerformanceChallenge{
		ReportID:	strings.TrimSpace(msg.ReportID),
		Challenger:	strings.TrimSpace(msg.Challenger),
		Reason:		strings.TrimSpace(msg.Reason),
		Epoch:		msg.Epoch,
		Accepted:	msg.Accepted,
		Status:		ChallengeStatusRejected,
	}
	if msg.Accepted {
		challenge.Status = ChallengeStatusAccepted
	}
	challenge.ChallengeID = ComputePerformanceChallengeID(challenge)
	challenge.ChallengeHash = ComputePerformanceChallengeHash(challenge)
	return challenge
}

func (challenge PerformanceChallenge) Validate() error {
	challenge = normalizeChallenge(challenge)
	if challenge.ReportID == "" {
		return errors.New("performance challenge report id is required")
	}
	if challenge.Challenger == "" {
		return errors.New("performance challenge challenger is required")
	}
	if challenge.Reason == "" {
		return errors.New("performance challenge reason is required")
	}
	if challenge.Epoch == 0 {
		return errors.New("performance challenge epoch is required")
	}
	if challenge.Status != ChallengeStatusAccepted && challenge.Status != ChallengeStatusRejected && challenge.Status != ChallengeStatusOpen {
		return errors.New("performance challenge status is invalid")
	}
	if !isHex64(challenge.ChallengeID) {
		return errors.New("performance challenge id must be a hex hash")
	}
	if !isHex64(challenge.ChallengeHash) {
		return errors.New("performance challenge hash must be a hex hash")
	}
	if challenge.ChallengeID != ComputePerformanceChallengeID(challenge) {
		return errors.New("performance challenge id mismatch")
	}
	if challenge.ChallengeHash != ComputePerformanceChallengeHash(challenge) {
		return errors.New("performance challenge hash mismatch")
	}
	return nil
}

func NormalizePerformanceOracleState(state PerformanceOracleState) PerformanceOracleState {
	if strings.TrimSpace(state.Params.Authority) == "" {
		state.Params = DefaultPerformanceOracleParams()
	}
	state.Params.Authority = strings.TrimSpace(state.Params.Authority)
	state.Epochs = normalizeEpochs(state.Epochs)
	return state
}

func (state PerformanceOracleState) findEpochByID(epochID uint64) (PerformanceEpoch, bool) {
	for _, epoch := range state.Epochs {
		if epoch.Epoch == epochID {
			return epoch, true
		}
	}
	return PerformanceEpoch{}, false
}

func (state PerformanceOracleState) epochByID(epochID uint64) PerformanceEpoch {
	epoch, found := state.findEpochByID(epochID)
	if found {
		return epoch
	}
	return finalizeEpochHash(PerformanceEpoch{Epoch: epochID})
}

func (state PerformanceOracleState) upsertEpoch(epoch PerformanceEpoch) PerformanceOracleState {
	next := make([]PerformanceEpoch, 0, len(state.Epochs)+1)
	replaced := false
	for _, existing := range state.Epochs {
		if existing.Epoch == epoch.Epoch {
			next = append(next, epoch)
			replaced = true
			continue
		}
		next = append(next, existing)
	}
	if !replaced {
		next = append(next, epoch)
	}
	state.Epochs = normalizeEpochs(next)
	return state
}

func aggregateValidatorReports(validator string, reports []PerformanceReport, params PerformanceOracleParams) PerformanceAggregate {
	reports = normalizeReports(reports)
	epoch := reports[0].Epoch
	uptime := make([]uint32, 0, len(reports))
	latency := make([]uint32, 0, len(reports))
	response := make([]uint32, 0, len(reports))
	missed := make([]uint32, 0, len(reports))
	peer := make([]uint32, 0, len(reports))
	for _, report := range reports {
		uptime = append(uptime, ratioScore(report.UptimeSignedBlocks, report.UptimeTotalBlocks, params.MaxScoreBps))
		latency = append(latency, inverseDurationScore(report.LatencyMillis, params.MaxLatencyMillis, params.MaxScoreBps))
		response = append(response, inverseDurationScore(report.ResponseTimeMillis, params.MaxResponseTimeMillis, params.MaxScoreBps))
		missed = append(missed, inverseRatioScore(report.MissedBlocks, report.MissedWindowBlocks, params.MaxScoreBps))
		peer = append(peer, report.PeerScoreBps)
	}
	aggregate := PerformanceAggregate{
		Epoch:			epoch,
		ValidatorAddress:	validator,
		ReportCount:		uint32(len(reports)),
		UptimeScoreBps:		trimmedAverageBps(uptime, params.OutlierTrimBps),
		LatencyScoreBps:	trimmedAverageBps(latency, params.OutlierTrimBps),
		ResponseTimeScoreBps:	trimmedAverageBps(response, params.OutlierTrimBps),
		MissedBlockScoreBps:	trimmedAverageBps(missed, params.OutlierTrimBps),
		PeerScoreBps:		trimmedAverageBps(peer, params.OutlierTrimBps),
	}
	aggregate.PerformanceScoreBps = averageBps([]uint32{
		aggregate.UptimeScoreBps,
		aggregate.LatencyScoreBps,
		aggregate.ResponseTimeScoreBps,
		aggregate.MissedBlockScoreBps,
		aggregate.PeerScoreBps,
	})
	if aggregate.PerformanceScoreBps > params.MaxScoreBps {
		aggregate.PerformanceScoreBps = params.MaxScoreBps
	}
	return aggregate
}

func trimmedAverageBps(values []uint32, trimBps uint32) uint32 {
	if len(values) == 0 {
		return 0
	}
	values = append([]uint32(nil), values...)
	sort.SliceStable(values, func(i, j int) bool { return values[i] < values[j] })
	trim := int((uint64(len(values)) * uint64(trimBps)) / 10_000)
	if trim*2 >= len(values) {
		trim = 0
	}
	return averageBps(values[trim : len(values)-trim])
}

func averageBps(values []uint32) uint32 {
	if len(values) == 0 {
		return 0
	}
	var total uint64
	for _, value := range values {
		total += uint64(value)
	}
	return uint32(total / uint64(len(values)))
}

func ratioScore(numerator uint64, denominator uint64, max uint32) uint32 {
	if denominator == 0 {
		return 0
	}
	score := (uint64(max) * numerator) / denominator
	if score > uint64(max) {
		return max
	}
	return uint32(score)
}

func inverseRatioScore(numerator uint64, denominator uint64, max uint32) uint32 {
	if denominator == 0 {
		return max
	}
	ratio := ratioScore(numerator, denominator, max)
	if ratio >= max {
		return 0
	}
	return max - ratio
}

func inverseDurationScore(value uint64, maxDuration uint64, max uint32) uint32 {
	if maxDuration == 0 {
		return 0
	}
	if value >= maxDuration {
		return 0
	}
	return uint32((uint64(max) * (maxDuration - value)) / maxDuration)
}

func authorizePerformanceOracle(params PerformanceOracleParams, authority string) error {
	if err := params.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(authority) != params.Authority {
		return errors.New("performance oracle message requires authority")
	}
	return nil
}

func normalizePerformanceReport(report PerformanceReport) PerformanceReport {
	report.ReportID = strings.ToLower(strings.TrimSpace(report.ReportID))
	report.ValidatorAddress = strings.TrimSpace(report.ValidatorAddress)
	report.ReporterAddress = strings.TrimSpace(report.ReporterAddress)
	report.Source = strings.TrimSpace(report.Source)
	report.ReportHash = strings.ToLower(strings.TrimSpace(report.ReportHash))
	if report.ReportID == "" {
		report.ReportID = ComputePerformanceReportID(report)
	}
	if report.ReportHash == "" {
		report.ReportHash = ComputePerformanceReportHash(report)
	}
	return report
}

func normalizeReports(reports []PerformanceReport) []PerformanceReport {
	out := make([]PerformanceReport, len(reports))
	for i, report := range reports {
		out[i] = normalizePerformanceReport(report)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Epoch != out[j].Epoch {
			return out[i].Epoch < out[j].Epoch
		}
		if out[i].ValidatorAddress != out[j].ValidatorAddress {
			return out[i].ValidatorAddress < out[j].ValidatorAddress
		}
		if out[i].ReporterAddress != out[j].ReporterAddress {
			return out[i].ReporterAddress < out[j].ReporterAddress
		}
		return out[i].ReportID < out[j].ReportID
	})
	return out
}

func normalizeAggregate(aggregate PerformanceAggregate) PerformanceAggregate {
	aggregate.ValidatorAddress = strings.TrimSpace(aggregate.ValidatorAddress)
	aggregate.AggregationHash = strings.ToLower(strings.TrimSpace(aggregate.AggregationHash))
	if aggregate.AggregationHash == "" {
		aggregate.AggregationHash = ComputePerformanceAggregateHash(aggregate)
	}
	return aggregate
}

func normalizeAggregates(aggregates []PerformanceAggregate) []PerformanceAggregate {
	out := make([]PerformanceAggregate, len(aggregates))
	for i, aggregate := range aggregates {
		out[i] = normalizeAggregate(aggregate)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Epoch != out[j].Epoch {
			return out[i].Epoch < out[j].Epoch
		}
		return out[i].ValidatorAddress < out[j].ValidatorAddress
	})
	return out
}

func normalizeChallenge(challenge PerformanceChallenge) PerformanceChallenge {
	challenge.ChallengeID = strings.ToLower(strings.TrimSpace(challenge.ChallengeID))
	challenge.ReportID = strings.TrimSpace(challenge.ReportID)
	challenge.Challenger = strings.TrimSpace(challenge.Challenger)
	challenge.Reason = strings.TrimSpace(challenge.Reason)
	challenge.Status = strings.TrimSpace(challenge.Status)
	challenge.ChallengeHash = strings.ToLower(strings.TrimSpace(challenge.ChallengeHash))
	if challenge.Status == "" {
		challenge.Status = ChallengeStatusRejected
		if challenge.Accepted {
			challenge.Status = ChallengeStatusAccepted
		}
	}
	if challenge.ChallengeID == "" {
		challenge.ChallengeID = ComputePerformanceChallengeID(challenge)
	}
	if challenge.ChallengeHash == "" {
		challenge.ChallengeHash = ComputePerformanceChallengeHash(challenge)
	}
	return challenge
}

func normalizeChallenges(challenges []PerformanceChallenge) []PerformanceChallenge {
	out := make([]PerformanceChallenge, len(challenges))
	for i, challenge := range challenges {
		out[i] = normalizeChallenge(challenge)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Epoch != out[j].Epoch {
			return out[i].Epoch < out[j].Epoch
		}
		return out[i].ChallengeID < out[j].ChallengeID
	})
	return out
}

func normalizeEpoch(epoch PerformanceEpoch) PerformanceEpoch {
	epoch.Reports = normalizeReports(epoch.Reports)
	epoch.Aggregates = normalizeAggregates(epoch.Aggregates)
	epoch.Challenges = normalizeChallenges(epoch.Challenges)
	epoch.EpochHash = strings.ToLower(strings.TrimSpace(epoch.EpochHash))
	if epoch.EpochHash == "" {
		epoch.EpochHash = ComputePerformanceEpochHash(epoch)
	}
	return epoch
}

func normalizeEpochs(epochs []PerformanceEpoch) []PerformanceEpoch {
	out := make([]PerformanceEpoch, len(epochs))
	for i, epoch := range epochs {
		out[i] = normalizeEpoch(epoch)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Epoch < out[j].Epoch
	})
	return out
}

func finalizeEpochHash(epoch PerformanceEpoch) PerformanceEpoch {
	epoch = normalizeEpoch(epoch)
	epoch.EpochHash = ComputePerformanceEpochHash(epoch)
	return epoch
}

func clonePerformanceOracleState(state PerformanceOracleState) PerformanceOracleState {
	state = NormalizePerformanceOracleState(state)
	return PerformanceOracleState{
		Params:			state.Params,
		AggregationEpoch:	state.AggregationEpoch,
		Epochs:			append([]PerformanceEpoch(nil), state.Epochs...),
	}
}

func ComputePerformanceReportID(report PerformanceReport) string {
	report.ReportID = ""
	report.ReportHash = ""
	return performanceHashParts("performance-report-id-v1",
		fmt.Sprint(report.Epoch),
		report.ValidatorAddress,
		report.ReporterAddress,
		report.Source,
		fmt.Sprint(report.SubmittedHeight),
	)
}

func ComputePerformanceReportHash(report PerformanceReport) string {
	return performanceHashParts("performance-report-v1",
		report.ReportID,
		fmt.Sprint(report.Epoch),
		report.ValidatorAddress,
		report.ReporterAddress,
		report.Source,
		fmt.Sprint(report.UptimeSignedBlocks),
		fmt.Sprint(report.UptimeTotalBlocks),
		fmt.Sprint(report.LatencyMillis),
		fmt.Sprint(report.ResponseTimeMillis),
		fmt.Sprint(report.MissedBlocks),
		fmt.Sprint(report.MissedWindowBlocks),
		fmt.Sprint(report.PeerScoreBps),
		fmt.Sprint(report.SubmittedHeight),
		fmt.Sprint(report.Slashable),
		fmt.Sprint(report.Challenged),
	)
}

func ComputePerformanceAggregateHash(aggregate PerformanceAggregate) string {
	aggregate.AggregationHash = ""
	return performanceHashParts("performance-aggregate-v1",
		fmt.Sprint(aggregate.Epoch),
		aggregate.ValidatorAddress,
		fmt.Sprint(aggregate.ReportCount),
		fmt.Sprint(aggregate.UptimeScoreBps),
		fmt.Sprint(aggregate.LatencyScoreBps),
		fmt.Sprint(aggregate.ResponseTimeScoreBps),
		fmt.Sprint(aggregate.MissedBlockScoreBps),
		fmt.Sprint(aggregate.PeerScoreBps),
		fmt.Sprint(aggregate.PerformanceScoreBps),
	)
}

func ComputePerformanceChallengeID(challenge PerformanceChallenge) string {
	challenge.ChallengeID = ""
	challenge.ChallengeHash = ""
	return performanceHashParts("performance-challenge-id-v1",
		fmt.Sprint(challenge.Epoch),
		challenge.ReportID,
		challenge.Challenger,
		challenge.Reason,
	)
}

func ComputePerformanceChallengeHash(challenge PerformanceChallenge) string {
	return performanceHashParts("performance-challenge-v1",
		challenge.ChallengeID,
		challenge.ReportID,
		challenge.Challenger,
		challenge.Reason,
		fmt.Sprint(challenge.Epoch),
		fmt.Sprint(challenge.Accepted),
		challenge.Status,
	)
}

func ComputePerformanceEpochHash(epoch PerformanceEpoch) string {
	epoch.Reports = normalizeReports(epoch.Reports)
	epoch.Aggregates = normalizeAggregates(epoch.Aggregates)
	epoch.Challenges = normalizeChallenges(epoch.Challenges)
	parts := []string{"performance-epoch-v1", fmt.Sprint(epoch.Epoch), fmt.Sprint(epoch.Finalized)}
	for _, report := range epoch.Reports {
		parts = append(parts, report.ReportHash)
	}
	for _, aggregate := range epoch.Aggregates {
		parts = append(parts, aggregate.AggregationHash)
	}
	for _, challenge := range epoch.Challenges {
		parts = append(parts, challenge.ChallengeHash)
	}
	return performanceHashParts(parts...)
}

func performanceHashParts(parts ...string) string {
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

func isHex64(value string) bool {
	if len(value) != 64 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func maxU64(left uint64, right uint64) uint64 {
	if left > right {
		return left
	}
	return right
}
