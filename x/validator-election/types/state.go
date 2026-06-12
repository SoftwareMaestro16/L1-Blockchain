package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/x/internal/prototype"
	validatorregistrytypes "github.com/sovereign-l1/l1/x/validator-registry/types"
)

const (
	ApplicationStatusPending	= "pending"
	ApplicationStatusWithdrawn	= "withdrawn"
	ApplicationStatusCommitted	= "committed"

	ExitStatusPending	= "pending"
	ExitStatusCancelled	= "cancelled"
	ExitStatusFinalized	= "finalized"

	MaxCandidatesV1		= uint32(10_000)
	MaxValidatorSetSizeV1	= uint32(512)
	MaxTransitionHistoryV1	= uint32(512)
	DefaultElectionWindow	= uint64(100)
	DefaultWithdrawCutoff	= uint64(80)
	DefaultFrozenUnlock	= uint64(1_000)
	DefaultMaxPower		= uint64(1_000_000)
	DefaultMaxTotalPower	= uint64(100_000_000)
)

type Params struct {
	Authority		string
	MaxCandidates		uint32
	MaxValidatorSetSize	uint32
	MaxValidatorPower	uint64
	MaxTotalVotingPower	uint64
	ElectionWindowBlocks	uint64
	WithdrawDeadlineBlocks	uint64
	FrozenStakeUnlockBlocks	uint64
	MaxTransitionHistory	uint32
}

type ValidatorPower struct {
	OperatorAddress		string
	ConsensusPublicKey	string
	VotingPower		uint64
	ValidatorStatus		string
}

type ElectionWindow struct {
	StartHeight		uint64
	EndHeight		uint64
	WithdrawDeadlineHeight	uint64
}

type CandidateApplication struct {
	OperatorAddress		string
	ConsensusPublicKey	string
	RequestedPower		uint64
	SelfBond		uint64
	ValidatorStatus		string
	Status			string
	AppliedHeight		uint64
	UpdatedHeight		uint64
}

type FrozenStake struct {
	OperatorAddress	string
	Amount		uint64
	FrozenAtHeight	uint64
	UnlockHeight	uint64
	Released	bool
}

type PendingExit struct {
	OperatorAddress	string
	RequestedHeight	uint64
	Status		string
}

type ValidatorPowerCap struct {
	OperatorAddress	string
	MaxVotingPower	uint64
}

type ElectionResult struct {
	Epoch		uint64
	Height		uint64
	NextSet		[]ValidatorPower
	Committed	bool
	Finalized	bool
}

type RewardDistributionSnapshot struct {
	Epoch			uint64
	Height			uint64
	ValidatorPowers		[]ValidatorPower
	TotalVotingPower	uint64
}

type ValidatorSetTransition struct {
	Epoch		uint64
	Height		uint64
	PreviousSet	[]ValidatorPower
	CurrentSet	[]ValidatorPower
	NextSet		[]ValidatorPower
}

type State struct {
	PreviousValidatorSet		[]ValidatorPower
	CurrentValidatorSet		[]ValidatorPower
	NextValidatorSet		[]ValidatorPower
	ElectionEpoch			uint64
	ElectionWindow			ElectionWindow
	CandidateApplications		[]CandidateApplication
	FrozenStakes			[]FrozenStake
	PendingExits			[]PendingExit
	ValidatorPowerCaps		[]ValidatorPowerCap
	ElectionResults			[]ElectionResult
	RewardDistributionSnapshots	[]RewardDistributionSnapshot
	TransitionHistory		[]ValidatorSetTransition
}

type MsgApplyForValidatorSet struct {
	Authority	string
	Application	CandidateApplication
	Height		uint64
}

type MsgWithdrawApplication struct {
	Authority	string
	OperatorAddress	string
	Height		uint64
}

type MsgCommitElection struct {
	Authority	string
	Height		uint64
}

type MsgFinalizeElection struct {
	Authority	string
	Height		uint64
}

type MsgRequestValidatorExit struct {
	Authority	string
	OperatorAddress	string
	Height		uint64
}

type MsgCancelValidatorExit struct {
	Authority	string
	OperatorAddress	string
	Height		uint64
}

func DefaultParams() Params {
	return Params{
		Authority:			prototype.DefaultAuthority,
		MaxCandidates:			MaxCandidatesV1,
		MaxValidatorSetSize:		MaxValidatorSetSizeV1,
		MaxValidatorPower:		DefaultMaxPower,
		MaxTotalVotingPower:		DefaultMaxTotalPower,
		ElectionWindowBlocks:		DefaultElectionWindow,
		WithdrawDeadlineBlocks:		DefaultWithdrawCutoff,
		FrozenStakeUnlockBlocks:	DefaultFrozenUnlock,
		MaxTransitionHistory:		MaxTransitionHistoryV1,
	}
}

func DefaultState(params Params) State {
	return State{
		ElectionEpoch:	1,
		ElectionWindow: ElectionWindow{
			StartHeight:		1,
			EndHeight:		1 + params.ElectionWindowBlocks,
			WithdrawDeadlineHeight:	1 + params.WithdrawDeadlineBlocks,
		},
	}
}

func (p Params) Validate() error {
	if err := addressing.ValidateAuthorityAddress("validator election authority", p.Authority); err != nil {
		return err
	}
	if p.MaxCandidates == 0 || p.MaxCandidates > MaxCandidatesV1 {
		return fmt.Errorf("validator election max candidates must be between 1 and %d", MaxCandidatesV1)
	}
	if p.MaxValidatorSetSize == 0 || p.MaxValidatorSetSize > MaxValidatorSetSizeV1 {
		return fmt.Errorf("validator election max validator set size must be between 1 and %d", MaxValidatorSetSizeV1)
	}
	if p.MaxValidatorPower == 0 {
		return errors.New("validator election max validator power must be positive")
	}
	if p.MaxTotalVotingPower == 0 || p.MaxTotalVotingPower < p.MaxValidatorPower {
		return errors.New("validator election max total voting power must be positive and >= max validator power")
	}
	if p.ElectionWindowBlocks == 0 {
		return errors.New("validator election window must be positive")
	}
	if p.WithdrawDeadlineBlocks >= p.ElectionWindowBlocks {
		return errors.New("validator election withdraw deadline must be before window end")
	}
	if p.FrozenStakeUnlockBlocks == 0 {
		return errors.New("validator election frozen stake unlock blocks must be positive")
	}
	if p.MaxTransitionHistory == 0 || p.MaxTransitionHistory > MaxTransitionHistoryV1 {
		return fmt.Errorf("validator election max transition history must be between 1 and %d", MaxTransitionHistoryV1)
	}
	return nil
}

func (p Params) Authorize(authority string) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("validator election update authority", authority); err != nil {
		return err
	}
	if authority != p.Authority {
		return errors.New("validator election update requires governance authority")
	}
	return nil
}

func (s State) Validate(params Params) error {
	if err := params.Validate(); err != nil {
		return err
	}
	if err := s.ElectionWindow.Validate(); err != nil {
		return err
	}
	for label, set := range map[string][]ValidatorPower{
		"previous":	s.PreviousValidatorSet,
		"current":	s.CurrentValidatorSet,
		"next":		s.NextValidatorSet,
	} {
		if err := validateValidatorSet(label, set, params, label == "next"); err != nil {
			return err
		}
	}
	if uint32(len(s.CandidateApplications)) > params.MaxCandidates {
		return errors.New("validator election candidate limit exceeded")
	}
	if err := validateApplications(s.CandidateApplications); err != nil {
		return err
	}
	if err := validateFrozenStakes(s.FrozenStakes); err != nil {
		return err
	}
	if err := validatePendingExits(s.PendingExits); err != nil {
		return err
	}
	if err := validatePowerCaps(s.ValidatorPowerCaps, params); err != nil {
		return err
	}
	if err := validateResults(s.ElectionResults, params); err != nil {
		return err
	}
	if err := validateSnapshots(s.RewardDistributionSnapshots, params); err != nil {
		return err
	}
	if uint32(len(s.TransitionHistory)) > params.MaxTransitionHistory {
		return errors.New("validator election transition history limit exceeded")
	}
	for _, transition := range s.TransitionHistory {
		if transition.Epoch == 0 || transition.Height == 0 {
			return errors.New("validator election transition epoch and height must be positive")
		}
	}
	return nil
}

func (w ElectionWindow) Validate() error {
	if w.StartHeight == 0 || w.EndHeight <= w.StartHeight {
		return errors.New("validator election window heights are invalid")
	}
	if w.WithdrawDeadlineHeight < w.StartHeight || w.WithdrawDeadlineHeight >= w.EndHeight {
		return errors.New("validator election withdraw deadline is outside election window")
	}
	return nil
}

func (v ValidatorPower) Validate(params Params, rejectJailed bool) error {
	if err := addressing.ValidateAuthorityAddress("validator election operator address", v.OperatorAddress); err != nil {
		return err
	}
	if strings.TrimSpace(v.ConsensusPublicKey) == "" {
		return errors.New("validator election consensus public key must be non-empty")
	}
	if v.VotingPower == 0 || v.VotingPower > params.MaxValidatorPower {
		return errors.New("validator election voting power must be positive and bounded")
	}
	if rejectJailed && (v.ValidatorStatus == validatorregistrytypes.StatusJailed || v.ValidatorStatus == validatorregistrytypes.StatusTombstoned) {
		return errors.New("validator election next set cannot include jailed or tombstoned validators")
	}
	if strings.TrimSpace(v.ValidatorStatus) == "" {
		return errors.New("validator election validator status must be non-empty")
	}
	return nil
}

func (a CandidateApplication) Normalize() CandidateApplication {
	a.OperatorAddress = strings.TrimSpace(a.OperatorAddress)
	a.ConsensusPublicKey = strings.TrimSpace(a.ConsensusPublicKey)
	a.ValidatorStatus = strings.TrimSpace(a.ValidatorStatus)
	if a.ValidatorStatus == "" {
		a.ValidatorStatus = validatorregistrytypes.StatusCandidate
	}
	a.Status = strings.TrimSpace(a.Status)
	if a.Status == "" {
		a.Status = ApplicationStatusPending
	}
	return a
}

func (a CandidateApplication) Validate() error {
	a = a.Normalize()
	if err := addressing.ValidateAuthorityAddress("validator election candidate operator", a.OperatorAddress); err != nil {
		return err
	}
	if a.ConsensusPublicKey == "" {
		return errors.New("validator election candidate consensus key must be non-empty")
	}
	if a.RequestedPower == 0 {
		return errors.New("validator election candidate requested power must be positive")
	}
	if a.SelfBond == 0 {
		return errors.New("validator election candidate self bond must be positive")
	}
	if a.AppliedHeight == 0 || a.UpdatedHeight == 0 {
		return errors.New("validator election candidate heights must be positive")
	}
	if !IsApplicationStatus(a.Status) {
		return errors.New("validator election candidate status is invalid")
	}
	return nil
}

func (s State) Normalize(params Params) State {
	s.PreviousValidatorSet = SortValidatorSet(s.PreviousValidatorSet)
	s.CurrentValidatorSet = SortValidatorSet(s.CurrentValidatorSet)
	s.NextValidatorSet = SortValidatorSet(s.NextValidatorSet)
	s.CandidateApplications = SortApplications(normalizeApplications(s.CandidateApplications))
	s.FrozenStakes = SortFrozenStakes(s.FrozenStakes)
	s.PendingExits = SortPendingExits(s.PendingExits)
	s.ValidatorPowerCaps = SortPowerCaps(s.ValidatorPowerCaps)
	s.ElectionResults = SortResults(s.ElectionResults)
	s.RewardDistributionSnapshots = SortSnapshots(s.RewardDistributionSnapshots)
	s.TransitionHistory = SortTransitions(s.TransitionHistory)
	if uint32(len(s.TransitionHistory)) > params.MaxTransitionHistory {
		s.TransitionHistory = s.TransitionHistory[len(s.TransitionHistory)-int(params.MaxTransitionHistory):]
	}
	return s
}

func SortValidatorSet(values []ValidatorPower) []ValidatorPower {
	out := append([]ValidatorPower(nil), values...)
	sort.Slice(out, func(i, j int) bool { return out[i].OperatorAddress < out[j].OperatorAddress })
	return out
}

func SortApplications(values []CandidateApplication) []CandidateApplication {
	out := append([]CandidateApplication(nil), values...)
	sort.Slice(out, func(i, j int) bool { return out[i].OperatorAddress < out[j].OperatorAddress })
	return out
}

func SortFrozenStakes(values []FrozenStake) []FrozenStake {
	out := append([]FrozenStake(nil), values...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].OperatorAddress == out[j].OperatorAddress {
			return out[i].UnlockHeight < out[j].UnlockHeight
		}
		return out[i].OperatorAddress < out[j].OperatorAddress
	})
	return out
}

func SortPendingExits(values []PendingExit) []PendingExit {
	out := append([]PendingExit(nil), values...)
	sort.Slice(out, func(i, j int) bool { return out[i].OperatorAddress < out[j].OperatorAddress })
	return out
}

func SortPowerCaps(values []ValidatorPowerCap) []ValidatorPowerCap {
	out := append([]ValidatorPowerCap(nil), values...)
	sort.Slice(out, func(i, j int) bool { return out[i].OperatorAddress < out[j].OperatorAddress })
	return out
}

func SortResults(values []ElectionResult) []ElectionResult {
	out := append([]ElectionResult(nil), values...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Epoch == out[j].Epoch {
			return out[i].Height < out[j].Height
		}
		return out[i].Epoch < out[j].Epoch
	})
	return out
}

func SortSnapshots(values []RewardDistributionSnapshot) []RewardDistributionSnapshot {
	out := append([]RewardDistributionSnapshot(nil), values...)
	sort.Slice(out, func(i, j int) bool { return out[i].Epoch < out[j].Epoch })
	return out
}

func SortTransitions(values []ValidatorSetTransition) []ValidatorSetTransition {
	out := append([]ValidatorSetTransition(nil), values...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Epoch == out[j].Epoch {
			return out[i].Height < out[j].Height
		}
		return out[i].Epoch < out[j].Epoch
	})
	return out
}

func IsApplicationStatus(status string) bool {
	switch status {
	case ApplicationStatusPending, ApplicationStatusWithdrawn, ApplicationStatusCommitted:
		return true
	default:
		return false
	}
}

func IsExitStatus(status string) bool {
	switch status {
	case ExitStatusPending, ExitStatusCancelled, ExitStatusFinalized:
		return true
	default:
		return false
	}
}

func CandidateRank(app CandidateApplication, cap uint64, params Params) ValidatorPower {
	power := app.RequestedPower
	if app.SelfBond < power {
		power = app.SelfBond
	}
	if cap > 0 && power > cap {
		power = cap
	}
	if power > params.MaxValidatorPower {
		power = params.MaxValidatorPower
	}
	return ValidatorPower{
		OperatorAddress:	app.OperatorAddress,
		ConsensusPublicKey:	app.ConsensusPublicKey,
		VotingPower:		power,
		ValidatorStatus:	app.ValidatorStatus,
	}
}

func validateValidatorSet(label string, values []ValidatorPower, params Params, rejectJailed bool) error {
	if uint32(len(values)) > params.MaxValidatorSetSize {
		return fmt.Errorf("validator election %s set exceeds max size", label)
	}
	seen := map[string]struct{}{}
	total := uint64(0)
	sorted := SortValidatorSet(values)
	for i, value := range values {
		if value.OperatorAddress != sorted[i].OperatorAddress {
			return fmt.Errorf("validator election %s set must be sorted deterministically", label)
		}
		if _, found := seen[value.OperatorAddress]; found {
			return fmt.Errorf("validator election %s set has duplicate operator", label)
		}
		seen[value.OperatorAddress] = struct{}{}
		if err := value.Validate(params, rejectJailed); err != nil {
			return err
		}
		if total > params.MaxTotalVotingPower-value.VotingPower {
			return errors.New("validator election total voting power exceeds configured max")
		}
		total += value.VotingPower
	}
	return nil
}

func validateApplications(values []CandidateApplication) error {
	seen := map[string]struct{}{}
	sorted := SortApplications(normalizeApplications(values))
	for i, value := range normalizeApplications(values) {
		if value.OperatorAddress != sorted[i].OperatorAddress {
			return errors.New("validator election applications must be sorted deterministically")
		}
		if _, found := seen[value.OperatorAddress]; found {
			return errors.New("validator election duplicate candidate application")
		}
		seen[value.OperatorAddress] = struct{}{}
		if err := value.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func validateFrozenStakes(values []FrozenStake) error {
	seen := map[string]struct{}{}
	for _, value := range values {
		if err := addressing.ValidateAuthorityAddress("validator election frozen stake operator", value.OperatorAddress); err != nil {
			return err
		}
		if value.Amount == 0 || value.FrozenAtHeight == 0 || value.UnlockHeight <= value.FrozenAtHeight {
			return errors.New("validator election frozen stake amount and heights are invalid")
		}
		key := value.OperatorAddress + ":" + fmt.Sprint(value.UnlockHeight)
		if _, found := seen[key]; found {
			return errors.New("validator election duplicate frozen stake")
		}
		seen[key] = struct{}{}
	}
	return nil
}

func validatePendingExits(values []PendingExit) error {
	seen := map[string]struct{}{}
	for _, value := range values {
		if err := addressing.ValidateAuthorityAddress("validator election pending exit operator", value.OperatorAddress); err != nil {
			return err
		}
		if value.RequestedHeight == 0 || !IsExitStatus(value.Status) {
			return errors.New("validator election pending exit is invalid")
		}
		if _, found := seen[value.OperatorAddress]; found {
			return errors.New("validator election duplicate pending exit")
		}
		seen[value.OperatorAddress] = struct{}{}
	}
	return nil
}

func validatePowerCaps(values []ValidatorPowerCap, params Params) error {
	seen := map[string]struct{}{}
	for _, value := range values {
		if err := addressing.ValidateAuthorityAddress("validator election power cap operator", value.OperatorAddress); err != nil {
			return err
		}
		if value.MaxVotingPower == 0 || value.MaxVotingPower > params.MaxValidatorPower {
			return errors.New("validator election power cap must be positive and bounded")
		}
		if _, found := seen[value.OperatorAddress]; found {
			return errors.New("validator election duplicate power cap")
		}
		seen[value.OperatorAddress] = struct{}{}
	}
	return nil
}

func validateResults(values []ElectionResult, params Params) error {
	for _, value := range values {
		if value.Epoch == 0 || value.Height == 0 {
			return errors.New("validator election result epoch and height must be positive")
		}
		if err := validateValidatorSet("result next", value.NextSet, params, true); err != nil {
			return err
		}
	}
	return nil
}

func validateSnapshots(values []RewardDistributionSnapshot, params Params) error {
	for _, value := range values {
		if value.Epoch == 0 || value.Height == 0 {
			return errors.New("validator election reward snapshot epoch and height must be positive")
		}
		if err := validateValidatorSet("reward snapshot", value.ValidatorPowers, params, false); err != nil {
			return err
		}
		total := uint64(0)
		for _, power := range value.ValidatorPowers {
			total += power.VotingPower
		}
		if total != value.TotalVotingPower {
			return errors.New("validator election reward snapshot total voting power mismatch")
		}
	}
	return nil
}

func normalizeApplications(values []CandidateApplication) []CandidateApplication {
	out := make([]CandidateApplication, 0, len(values))
	for _, value := range values {
		out = append(out, value.Normalize())
	}
	return out
}
