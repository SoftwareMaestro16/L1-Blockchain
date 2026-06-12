package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	configtypes "github.com/sovereign-l1/l1/x/config/types"
)

const (
	ProposalStatusVoting	= "voting"
	ProposalStatusPassed	= "passed"
	ProposalStatusRejected	= "rejected"
	ProposalStatusVetoed	= "vetoed"
	ProposalStatusExecuted	= "executed"

	VoteOptionYes		= "yes"
	VoteOptionNo		= "no"
	VoteOptionAbstain	= "abstain"
	VoteOptionVeto		= "veto"
)

type ConfigVotingParams struct {
	MaxProposals		uint32
	MaxVotes		uint32
	MaxSnapshotEntries	uint32
	QuorumBps		uint32
	ThresholdBps		uint32
	VetoThresholdBps	uint32
	VotingPeriod		uint64
	ExecutionDelay		uint64
	EmergencyDelay		uint64
	MaxMetadataBytes	uint32
	MaxConstitutionBytes	uint32
	BpsScale		uint32
	VetoAuthorities		[]string
}

type ConfigVotingState struct {
	Proposals	[]ConfigProposal
	Votes		[]ConfigVote
}

type ConfigProposal struct {
	ProposalID				string
	Title					string
	ConfigKey				string
	ConfigValue				string
	Operation				string
	SubmittedBy				string
	Status					string
	Metadata				string
	ConstitutionReference			string
	Emergency				bool
	RequiresConstitutionalException		bool
	SnapshotHeight				uint64
	SubmitHeight				uint64
	VotingEndHeight				uint64
	EarliestExecutionHeight			uint64
	ExecutedHeight				uint64
	TotalVotingPower			uint64
	VotingPowerSnapshot			[]VotingPowerSnapshotEntry
	ExpectedPreviousVersion			uint64
	AllowMissingExpectedPrevious		bool
	ExecutionConstitutionValidatedAt	uint64
}

type VotingPowerSnapshotEntry struct {
	Voter	string
	Power	uint64
}

type ConfigVote struct {
	ProposalID	string
	Voter		string
	Option		string
	Power		uint64
	Height		uint64
}

type VoteTally struct {
	YesPower	uint64
	NoPower		uint64
	AbstainPower	uint64
	VetoPower	uint64
	TotalVoted	uint64
	TotalPower	uint64
}

type MsgSubmitConfigProposal struct {
	Authority	string
	Proposal	ConfigProposal
}

type MsgVoteConfigProposal struct {
	Voter		string
	ProposalID	string
	Option		string
	Height		uint64
}

type MsgExecuteConfigProposal struct {
	Authority	string
	ProposalID	string
	Height		uint64
	ConfigState	configtypes.ConfigState
	ConfigParams	configtypes.Params
}

type MsgVetoConfigProposal struct {
	Authority	string
	ProposalID	string
	Reason		string
	Height		uint64
}

func DefaultConfigVotingParams() ConfigVotingParams {
	return ConfigVotingParams{
		MaxProposals:		10_000,
		MaxVotes:		1_000_000,
		MaxSnapshotEntries:	10_000,
		QuorumBps:		4_000,
		ThresholdBps:		5_000,
		VetoThresholdBps:	3_340,
		VotingPeriod:		10_000,
		ExecutionDelay:		100,
		EmergencyDelay:		1,
		MaxMetadataBytes:	4_096,
		MaxConstitutionBytes:	512,
		BpsScale:		10_000,
		VetoAuthorities:	[]string{},
	}
}

func EmptyConfigVotingState() ConfigVotingState {
	return ConfigVotingState{
		Proposals:	[]ConfigProposal{},
		Votes:		[]ConfigVote{},
	}
}

func (p ConfigVotingParams) Validate() error {
	if p.MaxProposals == 0 || p.MaxVotes == 0 || p.MaxSnapshotEntries == 0 {
		return errors.New("config voting state limits must be positive")
	}
	if p.BpsScale == 0 {
		return errors.New("config voting bps scale must be positive")
	}
	if p.QuorumBps > p.BpsScale || p.ThresholdBps > p.BpsScale || p.VetoThresholdBps > p.BpsScale {
		return errors.New("config voting bps thresholds exceed scale")
	}
	if p.QuorumBps == 0 || p.ThresholdBps == 0 || p.VotingPeriod == 0 {
		return errors.New("config voting quorum, threshold, and period must be positive")
	}
	if p.EmergencyDelay > p.ExecutionDelay {
		return errors.New("config voting emergency delay cannot exceed normal delay")
	}
	return nil
}

func (s ConfigVotingState) Export() ConfigVotingState {
	out := ConfigVotingState{
		Proposals:	cloneProposals(s.Proposals),
		Votes:		cloneVotes(s.Votes),
	}
	SortProposals(out.Proposals)
	SortVotes(out.Votes)
	return out
}

func (s ConfigVotingState) Validate(params ConfigVotingParams) error {
	if err := params.Validate(); err != nil {
		return err
	}
	s = s.Export()
	if uint32(len(s.Proposals)) > params.MaxProposals {
		return errors.New("config voting proposal count exceeds limit")
	}
	if uint32(len(s.Votes)) > params.MaxVotes {
		return errors.New("config voting vote count exceeds limit")
	}
	proposals := map[string]ConfigProposal{}
	for _, proposal := range s.Proposals {
		if err := proposal.Validate(params); err != nil {
			return err
		}
		if _, found := proposals[proposal.ProposalID]; found {
			return fmt.Errorf("duplicate config proposal id %q", proposal.ProposalID)
		}
		proposals[proposal.ProposalID] = proposal
	}
	seenVotes := map[string]struct{}{}
	for _, vote := range s.Votes {
		if err := vote.Validate(); err != nil {
			return err
		}
		proposal, found := proposals[vote.ProposalID]
		if !found {
			return fmt.Errorf("config vote references unknown proposal %q", vote.ProposalID)
		}
		power, found := proposal.SnapshotPower(vote.Voter)
		if !found {
			return fmt.Errorf("config vote voter %q is not in snapshot", vote.Voter)
		}
		if vote.Power != power {
			return fmt.Errorf("config vote power for %q does not match snapshot", vote.Voter)
		}
		key := vote.ProposalID + "\x00" + vote.Voter
		if _, found := seenVotes[key]; found {
			return fmt.Errorf("duplicate config vote for voter %q", vote.Voter)
		}
		seenVotes[key] = struct{}{}
	}
	return nil
}

func (p ConfigProposal) Normalize(params ConfigVotingParams) ConfigProposal {
	p.ProposalID = strings.TrimSpace(p.ProposalID)
	p.Title = strings.TrimSpace(p.Title)
	p.ConfigKey = strings.TrimSpace(p.ConfigKey)
	p.Operation = strings.TrimSpace(p.Operation)
	if p.Operation == "" {
		p.Operation = configtypes.OperationSet
	}
	p.SubmittedBy = strings.TrimSpace(p.SubmittedBy)
	p.Status = strings.TrimSpace(p.Status)
	if p.Status == "" {
		p.Status = ProposalStatusVoting
	}
	p.Metadata = strings.TrimSpace(p.Metadata)
	p.ConstitutionReference = strings.TrimSpace(p.ConstitutionReference)
	p.VotingPowerSnapshot = cloneSnapshot(p.VotingPowerSnapshot)
	SortSnapshot(p.VotingPowerSnapshot)
	if p.VotingEndHeight == 0 && p.SubmitHeight != 0 {
		p.VotingEndHeight = p.SubmitHeight + params.VotingPeriod
	}
	if p.EarliestExecutionHeight == 0 && p.VotingEndHeight != 0 {
		delay := params.ExecutionDelay
		if p.Emergency {
			delay = params.EmergencyDelay
		}
		p.EarliestExecutionHeight = p.VotingEndHeight + delay
	}
	if p.TotalVotingPower == 0 {
		for _, entry := range p.VotingPowerSnapshot {
			p.TotalVotingPower += entry.Power
		}
	}
	return p
}

func (p ConfigProposal) Validate(params ConfigVotingParams) error {
	p = p.Normalize(params)
	if p.ProposalID == "" || p.Title == "" || p.ConfigKey == "" || p.SubmittedBy == "" {
		return errors.New("config proposal id, title, key, and submitter are required")
	}
	if uint32(len(p.Metadata)) > params.MaxMetadataBytes || uint32(len(p.ConstitutionReference)) > params.MaxConstitutionBytes {
		return errors.New("config proposal metadata exceeds limit")
	}
	if !IsProposalStatus(p.Status) {
		return errors.New("config proposal status is invalid")
	}
	if !configtypes.IsChangeOperation(p.Operation) {
		return errors.New("config proposal operation is invalid")
	}
	if p.Operation == configtypes.OperationDelete && p.ConfigValue != "" {
		return errors.New("config delete proposal value must be empty")
	}
	if p.SubmitHeight == 0 || p.SnapshotHeight == 0 || p.VotingEndHeight <= p.SubmitHeight || p.EarliestExecutionHeight < p.VotingEndHeight {
		return errors.New("config proposal heights are invalid")
	}
	if p.ExecutedHeight != 0 && p.Status != ProposalStatusExecuted {
		return errors.New("only executed config proposal can have execution height")
	}
	if p.Status == ProposalStatusExecuted && p.ExecutedHeight == 0 {
		return errors.New("executed config proposal requires execution height")
	}
	if len(p.VotingPowerSnapshot) == 0 || uint32(len(p.VotingPowerSnapshot)) > params.MaxSnapshotEntries {
		return errors.New("config proposal snapshot size is invalid")
	}
	total := uint64(0)
	previous := ""
	for _, entry := range p.VotingPowerSnapshot {
		if err := entry.Validate(); err != nil {
			return err
		}
		if previous == entry.Voter {
			return fmt.Errorf("duplicate config voting snapshot voter %q", entry.Voter)
		}
		previous = entry.Voter
		total += entry.Power
	}
	if total == 0 || p.TotalVotingPower != total {
		return errors.New("config proposal total voting power must match snapshot")
	}
	if err := configtypes.ValidateConfigKey("config proposal key", p.ConfigKey, configtypes.MaxConfigKeyBytesV1); err != nil {
		return err
	}
	if uint32(len(p.ConfigValue)) > configtypes.MaxConfigValueBytesV1 {
		return errors.New("config proposal value exceeds limit")
	}
	return nil
}

func (p ConfigProposal) SnapshotPower(voter string) (uint64, bool) {
	voter = strings.TrimSpace(voter)
	idx := sort.Search(len(p.VotingPowerSnapshot), func(i int) bool {
		return p.VotingPowerSnapshot[i].Voter >= voter
	})
	if idx >= len(p.VotingPowerSnapshot) || p.VotingPowerSnapshot[idx].Voter != voter {
		return 0, false
	}
	return p.VotingPowerSnapshot[idx].Power, true
}

func (p ConfigProposal) ToConfigChange(authority string) configtypes.ConfigChange {
	return configtypes.ConfigChange{
		ID:					p.ProposalID,
		Key:					p.ConfigKey,
		Value:					p.ConfigValue,
		Operation:				p.Operation,
		Status:					configtypes.ChangeStatusPending,
		SubmittedBy:				authority,
		RequiresConstitutionalException:	p.RequiresConstitutionalException,
		ExpectedPreviousVersion:		p.ExpectedPreviousVersion,
		AllowMissingExpectedPreviousVersion:	p.AllowMissingExpectedPrevious,
		CreatedHeight:				int64(p.SubmitHeight),
		UpdatedHeight:				int64(p.SubmitHeight),
	}
}

func (e VotingPowerSnapshotEntry) Normalize() VotingPowerSnapshotEntry {
	e.Voter = strings.TrimSpace(e.Voter)
	return e
}

func (e VotingPowerSnapshotEntry) Validate() error {
	e = e.Normalize()
	if e.Voter == "" || e.Power == 0 {
		return errors.New("config voting snapshot voter and power are required")
	}
	return nil
}

func (v ConfigVote) Normalize() ConfigVote {
	v.ProposalID = strings.TrimSpace(v.ProposalID)
	v.Voter = strings.TrimSpace(v.Voter)
	v.Option = strings.TrimSpace(v.Option)
	return v
}

func (v ConfigVote) Validate() error {
	v = v.Normalize()
	if v.ProposalID == "" || v.Voter == "" || v.Height == 0 {
		return errors.New("config vote proposal, voter, and height are required")
	}
	if v.Power == 0 {
		return errors.New("config vote power must be positive")
	}
	if !IsVoteOption(v.Option) {
		return errors.New("config vote option is invalid")
	}
	return nil
}

func Tally(proposal ConfigProposal, votes []ConfigVote) VoteTally {
	proposal = proposal.Normalize(DefaultConfigVotingParams())
	tally := VoteTally{TotalPower: proposal.TotalVotingPower}
	filtered := make([]ConfigVote, 0, len(votes))
	for _, vote := range votes {
		if vote.ProposalID == proposal.ProposalID {
			filtered = append(filtered, vote.Normalize())
		}
	}
	SortVotes(filtered)
	for _, vote := range filtered {
		switch vote.Option {
		case VoteOptionYes:
			tally.YesPower += vote.Power
		case VoteOptionNo:
			tally.NoPower += vote.Power
		case VoteOptionAbstain:
			tally.AbstainPower += vote.Power
		case VoteOptionVeto:
			tally.VetoPower += vote.Power
		}
		tally.TotalVoted += vote.Power
	}
	return tally
}

func (t VoteTally) HasQuorum(params ConfigVotingParams) bool {
	return meetsBps(t.TotalVoted, t.TotalPower, params.QuorumBps, params.BpsScale)
}

func (t VoteTally) HasThreshold(params ConfigVotingParams) bool {
	nonAbstain := t.YesPower + t.NoPower + t.VetoPower
	return meetsBps(t.YesPower, nonAbstain, params.ThresholdBps, params.BpsScale)
}

func (t VoteTally) HasVeto(params ConfigVotingParams) bool {
	return meetsBps(t.VetoPower, t.TotalPower, params.VetoThresholdBps, params.BpsScale)
}

func IsProposalStatus(status string) bool {
	switch status {
	case ProposalStatusVoting, ProposalStatusPassed, ProposalStatusRejected, ProposalStatusVetoed, ProposalStatusExecuted:
		return true
	default:
		return false
	}
}

func IsVoteOption(option string) bool {
	switch option {
	case VoteOptionYes, VoteOptionNo, VoteOptionAbstain, VoteOptionVeto:
		return true
	default:
		return false
	}
}

func SortProposals(proposals []ConfigProposal) {
	sort.SliceStable(proposals, func(i, j int) bool {
		if proposals[i].SubmitHeight != proposals[j].SubmitHeight {
			return proposals[i].SubmitHeight < proposals[j].SubmitHeight
		}
		return proposals[i].ProposalID < proposals[j].ProposalID
	})
}

func SortVotes(votes []ConfigVote) {
	sort.SliceStable(votes, func(i, j int) bool {
		if votes[i].ProposalID != votes[j].ProposalID {
			return votes[i].ProposalID < votes[j].ProposalID
		}
		return votes[i].Voter < votes[j].Voter
	})
}

func SortSnapshot(snapshot []VotingPowerSnapshotEntry) {
	sort.SliceStable(snapshot, func(i, j int) bool { return snapshot[i].Voter < snapshot[j].Voter })
}

func cloneProposals(proposals []ConfigProposal) []ConfigProposal {
	out := append([]ConfigProposal(nil), proposals...)
	params := DefaultConfigVotingParams()
	for i := range out {
		out[i] = out[i].Normalize(params)
	}
	return out
}

func cloneVotes(votes []ConfigVote) []ConfigVote {
	out := append([]ConfigVote(nil), votes...)
	for i := range out {
		out[i] = out[i].Normalize()
	}
	return out
}

func cloneSnapshot(snapshot []VotingPowerSnapshotEntry) []VotingPowerSnapshotEntry {
	out := append([]VotingPowerSnapshotEntry(nil), snapshot...)
	for i := range out {
		out[i] = out[i].Normalize()
	}
	return out
}

func meetsBps(numerator, denominator uint64, bps, scale uint32) bool {
	if denominator == 0 {
		return false
	}
	return numerator*uint64(scale) >= denominator*uint64(bps)
}
