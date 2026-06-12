package types

import (
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/x/internal/prototype"
)

const (
	StatusActive	= "active"
	StatusUnbonding	= "unbonding"
	StatusJailed	= "jailed"

	ReportStatusPending	= "pending"
	ReportStatusAccepted	= "accepted"
	ReportStatusRejected	= "rejected"
	ReportStatusMalicious	= "malicious"

	ReportTypeFault		= "fault"
	ReportTypeProof		= "proof"
	ReportTypeLatency	= "latency"
	ReportTypeAvailability	= "availability"

	MaxReportersV1		= uint32(100_000)
	MaxReportsV1		= uint32(1_000_000)
	MaxRewardHistoryV1	= uint32(10_000)
	MaxReportIDBytesV1	= uint32(96)
	MaxReportTypeBytesV1	= uint32(64)
	MaxSubjectBytesV1	= uint32(256)
	MaxPayloadHashBytesV1	= uint32(64)
	MaxPayloadBytesV1	= uint32(16_384)
	MaxBasisPoints		= uint32(10_000)

	DefaultMinBondAmount		= uint64(1_000_000)
	DefaultChallengePeriodBlocks	= uint64(1_000)
	DefaultMaliciousSlashBps	= uint32(5_000)
	DefaultInitialScore		= int64(1_000)
	DefaultAcceptedScoreDelta	= int64(10)
	DefaultRejectedScoreDelta	= int64(25)
	DefaultMaxRewardAmount		= uint64(1_000_000)
)

type Params struct {
	Authority		string
	MinBondAmount		uint64
	MaxReporters		uint32
	MaxReports		uint32
	MaxRewardHistory	uint32
	MaxReportIDBytes	uint32
	MaxReportTypeBytes	uint32
	MaxSubjectBytes		uint32
	MaxPayloadHashBytes	uint32
	MaxPayloadBytes		uint32
	ChallengePeriodBlocks	uint64
	MaliciousSlashBps	uint32
	InitialScore		int64
	AcceptedScoreDelta	int64
	RejectedScoreDelta	int64
	MaxRewardAmount		uint64
}

type State struct {
	Reporters	[]ReporterRecord
	Reports		[]ReportRecord
}

type ReporterRecord struct {
	ReporterAddress		string
	BondedAmount		uint64
	ReporterScore		int64
	AcceptedReports		uint64
	RejectedReports		uint64
	SlashedReporterBond	uint64
	Status			string
	UnbondingStartHeight	uint64
	UnbondingCompleteHeight	uint64
	RewardHistory		[]ReporterReward
}

type ReporterReward struct {
	ReportID	string
	Amount		uint64
	Claimed		bool
	CreatedAt	uint64
	ClaimedAt	uint64
}

type ReportRecord struct {
	ReportID		string
	ReporterAddress		string
	ReportType		string
	Subject			string
	PayloadHash		string
	PayloadSizeBytes	uint32
	Status			string
	SubmittedHeight		uint64
	FinalizedHeight		uint64
	RewardAmount		uint64
	RewardClaimed		bool
	SlashAmount		uint64
}

type MsgRegisterReporter struct {
	Authority	string
	ReporterAddress	string
	Height		uint64
}

type MsgBondReporter struct {
	Authority	string
	ReporterAddress	string
	Amount		uint64
	Height		uint64
}

type MsgUnbondReporter struct {
	Authority	string
	ReporterAddress	string
	Height		uint64
}

type MsgSubmitReport struct {
	Authority		string
	ReporterAddress		string
	ReportID		string
	ReportType		string
	Subject			string
	PayloadHash		string
	PayloadSizeBytes	uint32
	Accepted		bool
	Malicious		bool
	RewardAmount		uint64
	Height			uint64
}

type MsgClaimReporterReward struct {
	Authority	string
	ReporterAddress	string
	ReportID	string
	Height		uint64
}

func DefaultParams() Params {
	return Params{
		Authority:		prototype.DefaultAuthority,
		MinBondAmount:		DefaultMinBondAmount,
		MaxReporters:		MaxReportersV1,
		MaxReports:		MaxReportsV1,
		MaxRewardHistory:	MaxRewardHistoryV1,
		MaxReportIDBytes:	MaxReportIDBytesV1,
		MaxReportTypeBytes:	MaxReportTypeBytesV1,
		MaxSubjectBytes:	MaxSubjectBytesV1,
		MaxPayloadHashBytes:	MaxPayloadHashBytesV1,
		MaxPayloadBytes:	MaxPayloadBytesV1,
		ChallengePeriodBlocks:	DefaultChallengePeriodBlocks,
		MaliciousSlashBps:	DefaultMaliciousSlashBps,
		InitialScore:		DefaultInitialScore,
		AcceptedScoreDelta:	DefaultAcceptedScoreDelta,
		RejectedScoreDelta:	DefaultRejectedScoreDelta,
		MaxRewardAmount:	DefaultMaxRewardAmount,
	}
}

func (p Params) Validate() error {
	if err := addressing.ValidateAuthorityAddress("reporter authority", p.Authority); err != nil {
		return err
	}
	if p.MinBondAmount == 0 {
		return errors.New("reporter minimum bond must be positive")
	}
	if p.MaxReporters == 0 || p.MaxReporters > MaxReportersV1 {
		return fmt.Errorf("reporter max reporters must be between 1 and %d", MaxReportersV1)
	}
	if p.MaxReports == 0 || p.MaxReports > MaxReportsV1 {
		return fmt.Errorf("reporter max reports must be between 1 and %d", MaxReportsV1)
	}
	if p.MaxRewardHistory == 0 || p.MaxRewardHistory > MaxRewardHistoryV1 {
		return fmt.Errorf("reporter max reward history must be between 1 and %d", MaxRewardHistoryV1)
	}
	if p.MaxReportIDBytes == 0 || p.MaxReportIDBytes > MaxReportIDBytesV1 {
		return fmt.Errorf("reporter max report id bytes must be between 1 and %d", MaxReportIDBytesV1)
	}
	if p.MaxReportTypeBytes == 0 || p.MaxReportTypeBytes > MaxReportTypeBytesV1 {
		return fmt.Errorf("reporter max report type bytes must be between 1 and %d", MaxReportTypeBytesV1)
	}
	if p.MaxSubjectBytes == 0 || p.MaxSubjectBytes > MaxSubjectBytesV1 {
		return fmt.Errorf("reporter max subject bytes must be between 1 and %d", MaxSubjectBytesV1)
	}
	if p.MaxPayloadHashBytes == 0 || p.MaxPayloadHashBytes > MaxPayloadHashBytesV1 {
		return fmt.Errorf("reporter max payload hash bytes must be between 1 and %d", MaxPayloadHashBytesV1)
	}
	if p.MaxPayloadBytes == 0 || p.MaxPayloadBytes > MaxPayloadBytesV1 {
		return fmt.Errorf("reporter max payload bytes must be between 1 and %d", MaxPayloadBytesV1)
	}
	if p.ChallengePeriodBlocks == 0 {
		return errors.New("reporter challenge period must be positive")
	}
	if p.MaliciousSlashBps == 0 || p.MaliciousSlashBps > MaxBasisPoints {
		return fmt.Errorf("reporter malicious slash bps must be within 1..%d", MaxBasisPoints)
	}
	if p.AcceptedScoreDelta < 0 || p.RejectedScoreDelta < 0 {
		return errors.New("reporter score deltas cannot be negative")
	}
	return nil
}

func (p Params) Authorize(authority string) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("reporter update authority", authority); err != nil {
		return err
	}
	if authority != p.Authority {
		return errors.New("reporter update requires governance authority")
	}
	return nil
}

func (s State) Validate(params Params) error {
	if err := params.Validate(); err != nil {
		return err
	}
	if uint32(len(s.Reporters)) > params.MaxReporters {
		return errors.New("reporter count limit exceeded")
	}
	if uint32(len(s.Reports)) > params.MaxReports {
		return errors.New("report count limit exceeded")
	}
	reporters := map[string]ReporterRecord{}
	for _, reporter := range s.Reporters {
		if err := reporter.Validate(params); err != nil {
			return err
		}
		if _, found := reporters[reporter.ReporterAddress]; found {
			return fmt.Errorf("duplicate reporter %s", reporter.ReporterAddress)
		}
		reporters[reporter.ReporterAddress] = reporter
	}
	reports := map[string]struct{}{}
	rewardReports := map[string]struct{}{}
	for _, reporter := range s.Reporters {
		for _, reward := range reporter.RewardHistory {
			if _, found := rewardReports[reward.ReportID]; found {
				return fmt.Errorf("duplicate reporter reward for report %s", reward.ReportID)
			}
			rewardReports[reward.ReportID] = struct{}{}
		}
	}
	for _, report := range s.Reports {
		if err := report.Validate(params); err != nil {
			return err
		}
		if _, found := reports[report.ReportID]; found {
			return fmt.Errorf("duplicate report %s", report.ReportID)
		}
		reports[report.ReportID] = struct{}{}
		reporter, found := reporters[report.ReporterAddress]
		if !found {
			return fmt.Errorf("report %s references unknown reporter", report.ReportID)
		}
		if report.Status == ReportStatusAccepted {
			if _, found := rewardReports[report.ReportID]; !found {
				return fmt.Errorf("accepted report %s missing reward history", report.ReportID)
			}
			for _, reward := range reporter.RewardHistory {
				if reward.ReportID == report.ReportID && reward.Claimed != report.RewardClaimed {
					return fmt.Errorf("report %s reward claim state mismatch", report.ReportID)
				}
			}
		}
		if report.Status == ReportStatusMalicious && report.SlashAmount == 0 {
			return fmt.Errorf("malicious report %s must slash reporter bond", report.ReportID)
		}
	}
	return nil
}

func (r ReporterRecord) Validate(params Params) error {
	if err := addressing.ValidateAuthorityAddress("reporter address", r.ReporterAddress); err != nil {
		return err
	}
	if !isReporterStatus(r.Status) {
		return fmt.Errorf("unsupported reporter status %q", r.Status)
	}
	if r.Status == StatusActive && r.UnbondingCompleteHeight != 0 {
		return errors.New("active reporter cannot have unbonding completion height")
	}
	if r.Status == StatusUnbonding && (r.UnbondingStartHeight == 0 || r.UnbondingCompleteHeight <= r.UnbondingStartHeight) {
		return errors.New("unbonding reporter heights are invalid")
	}
	if uint32(len(r.RewardHistory)) > params.MaxRewardHistory {
		return errors.New("reporter reward history limit exceeded")
	}
	rewards := map[string]struct{}{}
	for _, reward := range r.RewardHistory {
		if err := reward.Validate(params); err != nil {
			return err
		}
		if _, found := rewards[reward.ReportID]; found {
			return fmt.Errorf("duplicate reward for report %s", reward.ReportID)
		}
		rewards[reward.ReportID] = struct{}{}
	}
	return nil
}

func (r ReporterReward) Validate(params Params) error {
	if err := validateID("reporter reward report id", r.ReportID, params.MaxReportIDBytes); err != nil {
		return err
	}
	if r.Amount == 0 || r.Amount > params.MaxRewardAmount {
		return errors.New("reporter reward amount outside configured bounds")
	}
	if r.CreatedAt == 0 {
		return errors.New("reporter reward creation height must be positive")
	}
	if r.Claimed && r.ClaimedAt == 0 {
		return errors.New("claimed reporter reward must have claimed height")
	}
	if !r.Claimed && r.ClaimedAt != 0 {
		return errors.New("unclaimed reporter reward cannot have claimed height")
	}
	return nil
}

func (r ReportRecord) Validate(params Params) error {
	if err := validateID("report id", r.ReportID, params.MaxReportIDBytes); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("report reporter address", r.ReporterAddress); err != nil {
		return err
	}
	if !isReportType(r.ReportType) {
		return fmt.Errorf("unsupported report type %q", r.ReportType)
	}
	if err := validateBoundedText("report subject", r.Subject, params.MaxSubjectBytes); err != nil {
		return err
	}
	if err := validatePayloadHash(r.PayloadHash, params.MaxPayloadHashBytes); err != nil {
		return err
	}
	if r.PayloadSizeBytes == 0 || r.PayloadSizeBytes > params.MaxPayloadBytes {
		return errors.New("report payload size outside configured bounds")
	}
	if !isReportStatus(r.Status) {
		return fmt.Errorf("unsupported report status %q", r.Status)
	}
	if r.SubmittedHeight == 0 {
		return errors.New("report submitted height must be positive")
	}
	if r.FinalizedHeight != 0 && r.FinalizedHeight < r.SubmittedHeight {
		return errors.New("report finalized height cannot precede submission")
	}
	if r.Status == ReportStatusPending && (r.FinalizedHeight != 0 || r.RewardAmount != 0 || r.SlashAmount != 0 || r.RewardClaimed) {
		return errors.New("pending report cannot have final effects")
	}
	if r.Status == ReportStatusAccepted && (r.FinalizedHeight == 0 || r.RewardAmount == 0 || r.RewardAmount > params.MaxRewardAmount || r.SlashAmount != 0) {
		return errors.New("accepted report reward state is invalid")
	}
	if (r.Status == ReportStatusRejected || r.Status == ReportStatusMalicious) && r.RewardAmount != 0 {
		return errors.New("rejected report cannot carry reward")
	}
	if r.Status != ReportStatusAccepted && r.RewardClaimed {
		return errors.New("only accepted report reward can be claimed")
	}
	return nil
}

func (s State) Normalize(params Params) State {
	s.Reporters = SortReporters(s.Reporters)
	for idx := range s.Reporters {
		s.Reporters[idx].RewardHistory = SortRewards(s.Reporters[idx].RewardHistory)
	}
	s.Reports = SortReports(s.Reports)
	return s
}

func SortReporters(values []ReporterRecord) []ReporterRecord {
	out := append([]ReporterRecord(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].ReporterAddress < out[j].ReporterAddress })
	return out
}

func SortReports(values []ReportRecord) []ReportRecord {
	out := append([]ReportRecord(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].ReportID < out[j].ReportID })
	return out
}

func SortRewards(values []ReporterReward) []ReporterReward {
	out := append([]ReporterReward(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].ReportID < out[j].ReportID })
	return out
}

func validateID(field, value string, maxBytes uint32) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("%s is required", field)
	}
	if uint32(len(value)) > maxBytes || strings.ContainsAny(value, " \t\r\n") {
		return fmt.Errorf("%s must be non-blank, whitespace-free, and within configured length", field)
	}
	return nil
}

func validateBoundedText(field, value string, maxBytes uint32) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("%s is required", field)
	}
	if uint32(len(value)) > maxBytes {
		return fmt.Errorf("%s exceeds configured length", field)
	}
	return nil
}

func validatePayloadHash(value string, maxBytes uint32) error {
	if err := validateID("report payload hash", value, maxBytes); err != nil {
		return err
	}
	if len(value)%2 != 0 {
		return errors.New("report payload hash must be even-length hex")
	}
	if _, err := hex.DecodeString(value); err != nil {
		return fmt.Errorf("report payload hash must be hex: %w", err)
	}
	return nil
}

func isReporterStatus(status string) bool {
	return status == StatusActive || status == StatusUnbonding || status == StatusJailed
}

func isReportStatus(status string) bool {
	return status == ReportStatusPending || status == ReportStatusAccepted || status == ReportStatusRejected || status == ReportStatusMalicious
}

func isReportType(reportType string) bool {
	switch reportType {
	case ReportTypeFault, ReportTypeProof, ReportTypeLatency, ReportTypeAvailability:
		return true
	default:
		return false
	}
}

func SlashAmount(bond uint64, bps uint32) uint64 {
	amount := bond * uint64(bps) / uint64(MaxBasisPoints)
	if amount == 0 && bond > 0 {
		return 1
	}
	if amount > bond {
		return bond
	}
	return amount
}
