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
	ClaimStatusPending	= "pending"
	ClaimStatusPaid		= "paid"
	ClaimStatusRejected	= "rejected"

	WithdrawalStatusNone		= ""
	WithdrawalStatusPending		= "pending"
	WithdrawalStatusCompleted	= "completed"

	DefaultMinimumInsurance		= uint64(1_000)
	DefaultWithdrawalLockBlocks	= uint64(1_000)
	DefaultSlashCoverageBps		= uint32(10_000)
	MaxValidatorsV1			= uint32(100_000)
	MaxClaimsV1			= uint32(200_000)
	MaxClaimIDBytesV1		= uint32(96)
	MaxReasonBytesV1		= uint32(512)
	MaxCoverageRulesV1		= uint32(128)
	MaxFaultTypeBytesV1		= uint32(128)
	MaxBasisPoints			= uint32(10_000)
)

type Params struct {
	Authority		string
	Enabled			bool
	MinimumInsurance	uint64
	WithdrawalLockBlocks	uint64
	DefaultSlashCoverageBps	uint32
	MaxValidators		uint32
	MaxClaims		uint32
	MaxClaimIDBytes		uint32
	MaxReasonBytes		uint32
	MaxCoverageRules	uint32
	MaxFaultTypeBytes	uint32
}

type State struct {
	Insurances	[]ValidatorInsurance
	Claims		[]InsuranceClaim
	CoverageRules	[]SlashCoverageRule
}

type ValidatorInsurance struct {
	ValidatorAddress	string
	Balance			uint64
	PendingWithdrawal	PendingInsuranceWithdrawal
	ValidatorStatus		string
}

type PendingInsuranceWithdrawal struct {
	Amount		uint64
	Recipient	string
	RequestHeight	uint64
	CompleteHeight	uint64
	Status		string
}

type InsuranceClaim struct {
	ClaimID			string
	ValidatorAddress	string
	Claimant		string
	Amount			uint64
	PayoutAmount		uint64
	Status			string
	Reason			string
	SubmittedHeight		uint64
	ResolvedHeight		uint64
	Paid			bool
}

type SlashCoverageRule struct {
	FaultType	string
	CoverageBps	uint32
}

type SlashCoverageResult struct {
	ValidatorAddress	string
	SlashAmount		uint64
	CoveredAmount		uint64
	RemainingPenalty	uint64
	CoverageBps		uint32
}

type MsgFundValidatorInsurance struct {
	Authority		string
	ValidatorAddress	string
	Funder			string
	Amount			uint64
	Height			uint64
}

type MsgWithdrawValidatorInsurance struct {
	Authority		string
	ValidatorAddress	string
	Recipient		string
	Amount			uint64
	Height			uint64
	ValidatorStatus		string
}

type MsgSubmitInsuranceClaim struct {
	Authority		string
	ClaimID			string
	ValidatorAddress	string
	Claimant		string
	Amount			uint64
	Reason			string
	Height			uint64
}

type MsgResolveInsuranceClaim struct {
	Authority	string
	ClaimID		string
	Approved	bool
	Height		uint64
}

func DefaultParams() Params {
	return Params{
		Authority:			prototype.DefaultAuthority,
		Enabled:			true,
		MinimumInsurance:		DefaultMinimumInsurance,
		WithdrawalLockBlocks:		DefaultWithdrawalLockBlocks,
		DefaultSlashCoverageBps:	DefaultSlashCoverageBps,
		MaxValidators:			MaxValidatorsV1,
		MaxClaims:			MaxClaimsV1,
		MaxClaimIDBytes:		MaxClaimIDBytesV1,
		MaxReasonBytes:			MaxReasonBytesV1,
		MaxCoverageRules:		MaxCoverageRulesV1,
		MaxFaultTypeBytes:		MaxFaultTypeBytesV1,
	}
}

func DefaultState(params Params) State {
	return State{
		CoverageRules: []SlashCoverageRule{{
			FaultType:	"default",
			CoverageBps:	params.DefaultSlashCoverageBps,
		}},
	}.Normalize(params)
}

func (p Params) Validate() error {
	if err := addressing.ValidateAuthorityAddress("validator insurance authority", p.Authority); err != nil {
		return err
	}
	if p.WithdrawalLockBlocks == 0 {
		return errors.New("validator insurance withdrawal lock period must be positive")
	}
	if p.DefaultSlashCoverageBps > MaxBasisPoints {
		return errors.New("validator insurance default slash coverage exceeds 100%")
	}
	if p.MaxValidators == 0 || p.MaxValidators > MaxValidatorsV1 {
		return fmt.Errorf("validator insurance max validators must be between 1 and %d", MaxValidatorsV1)
	}
	if p.MaxClaims == 0 || p.MaxClaims > MaxClaimsV1 {
		return fmt.Errorf("validator insurance max claims must be between 1 and %d", MaxClaimsV1)
	}
	if p.MaxClaimIDBytes == 0 || p.MaxClaimIDBytes > MaxClaimIDBytesV1 {
		return fmt.Errorf("validator insurance max claim id bytes must be between 1 and %d", MaxClaimIDBytesV1)
	}
	if p.MaxReasonBytes == 0 || p.MaxReasonBytes > MaxReasonBytesV1 {
		return fmt.Errorf("validator insurance max reason bytes must be between 1 and %d", MaxReasonBytesV1)
	}
	if p.MaxCoverageRules == 0 || p.MaxCoverageRules > MaxCoverageRulesV1 {
		return fmt.Errorf("validator insurance max coverage rules must be between 1 and %d", MaxCoverageRulesV1)
	}
	if p.MaxFaultTypeBytes == 0 || p.MaxFaultTypeBytes > MaxFaultTypeBytesV1 {
		return fmt.Errorf("validator insurance max fault type bytes must be between 1 and %d", MaxFaultTypeBytesV1)
	}
	return nil
}

func (p Params) Authorize(authority string) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("validator insurance update authority", authority); err != nil {
		return err
	}
	if authority != p.Authority {
		return errors.New("validator insurance update requires governance authority")
	}
	return nil
}

func (s State) Validate(params Params) error {
	if err := params.Validate(); err != nil {
		return err
	}
	if uint32(len(s.Insurances)) > params.MaxValidators {
		return errors.New("validator insurance validator limit exceeded")
	}
	if uint32(len(s.Claims)) > params.MaxClaims {
		return errors.New("validator insurance claim limit exceeded")
	}
	if uint32(len(s.CoverageRules)) > params.MaxCoverageRules {
		return errors.New("validator insurance coverage rule limit exceeded")
	}
	seenValidators := map[string]struct{}{}
	for _, insurance := range s.Insurances {
		if err := insurance.Validate(params); err != nil {
			return err
		}
		if _, found := seenValidators[insurance.ValidatorAddress]; found {
			return fmt.Errorf("duplicate validator insurance %s", insurance.ValidatorAddress)
		}
		seenValidators[insurance.ValidatorAddress] = struct{}{}
	}
	seenClaims := map[string]struct{}{}
	for _, claim := range s.Claims {
		if err := claim.Validate(params); err != nil {
			return err
		}
		if _, found := seenClaims[claim.ClaimID]; found {
			return fmt.Errorf("duplicate validator insurance claim %s", claim.ClaimID)
		}
		seenClaims[claim.ClaimID] = struct{}{}
	}
	seenRules := map[string]struct{}{}
	for _, rule := range s.CoverageRules {
		if err := rule.Validate(params); err != nil {
			return err
		}
		key := strings.ToLower(rule.FaultType)
		if _, found := seenRules[key]; found {
			return fmt.Errorf("duplicate validator insurance coverage rule %s", rule.FaultType)
		}
		seenRules[key] = struct{}{}
	}
	return nil
}

func (i ValidatorInsurance) Validate(params Params) error {
	if err := addressing.ValidateAuthorityAddress("validator insurance validator", i.ValidatorAddress); err != nil {
		return err
	}
	if err := i.PendingWithdrawal.Validate(); err != nil {
		return err
	}
	if i.ValidatorStatus != "" && !isValidatorStatus(i.ValidatorStatus) {
		return fmt.Errorf("unsupported validator insurance validator status %q", i.ValidatorStatus)
	}
	if params.Enabled && i.ValidatorStatus == validatorregistrytypes.StatusActive && i.Balance < params.MinimumInsurance {
		return errors.New("active validator insurance below minimum requirement")
	}
	return nil
}

func (w PendingInsuranceWithdrawal) Validate() error {
	if w.Status == WithdrawalStatusNone {
		if w.Amount != 0 || w.Recipient != "" || w.RequestHeight != 0 || w.CompleteHeight != 0 {
			return errors.New("empty validator insurance withdrawal must not carry values")
		}
		return nil
	}
	if w.Status != WithdrawalStatusPending && w.Status != WithdrawalStatusCompleted {
		return fmt.Errorf("unsupported validator insurance withdrawal status %q", w.Status)
	}
	if w.Amount == 0 || w.RequestHeight == 0 || w.CompleteHeight <= w.RequestHeight {
		return errors.New("validator insurance withdrawal amount and heights are invalid")
	}
	return addressing.ValidateAuthorityAddress("validator insurance withdrawal recipient", w.Recipient)
}

func (c InsuranceClaim) Validate(params Params) error {
	if err := validateClaimID(c.ClaimID, params.MaxClaimIDBytes); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("validator insurance claim validator", c.ValidatorAddress); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("validator insurance claimant", c.Claimant); err != nil {
		return err
	}
	if c.Amount == 0 {
		return errors.New("validator insurance claim amount must be positive")
	}
	if !isClaimStatus(c.Status) {
		return fmt.Errorf("unsupported validator insurance claim status %q", c.Status)
	}
	if uint32(len(c.Reason)) > params.MaxReasonBytes {
		return errors.New("validator insurance claim reason exceeds limit")
	}
	if c.SubmittedHeight == 0 {
		return errors.New("validator insurance claim submitted height must be positive")
	}
	if c.Status == ClaimStatusPending {
		if c.Paid || c.PayoutAmount != 0 || c.ResolvedHeight != 0 {
			return errors.New("pending validator insurance claim must not be paid or resolved")
		}
		return nil
	}
	if c.ResolvedHeight == 0 {
		return errors.New("resolved validator insurance claim must have resolved height")
	}
	if c.Status == ClaimStatusPaid && (!c.Paid || c.PayoutAmount == 0 || c.PayoutAmount > c.Amount) {
		return errors.New("paid validator insurance claim payout is invalid")
	}
	if c.Status == ClaimStatusRejected && (c.Paid || c.PayoutAmount != 0) {
		return errors.New("rejected validator insurance claim must not be paid")
	}
	return nil
}

func (r SlashCoverageRule) Validate(params Params) error {
	faultType := strings.TrimSpace(r.FaultType)
	if faultType == "" {
		return errors.New("validator insurance coverage rule fault type is required")
	}
	if uint32(len(faultType)) > params.MaxFaultTypeBytes || strings.ContainsAny(faultType, "\t\r\n") {
		return errors.New("validator insurance coverage rule fault type is invalid")
	}
	if r.CoverageBps > MaxBasisPoints {
		return errors.New("validator insurance coverage rule exceeds 100%")
	}
	return nil
}

func (s State) Normalize(params Params) State {
	s.Insurances = SortInsurances(s.Insurances)
	s.Claims = SortClaims(s.Claims)
	s.CoverageRules = SortCoverageRules(s.CoverageRules)
	return s
}

func SortInsurances(values []ValidatorInsurance) []ValidatorInsurance {
	out := append([]ValidatorInsurance(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].ValidatorAddress < out[j].ValidatorAddress })
	return out
}

func SortClaims(values []InsuranceClaim) []InsuranceClaim {
	out := append([]InsuranceClaim(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].ClaimID < out[j].ClaimID })
	return out
}

func SortCoverageRules(values []SlashCoverageRule) []SlashCoverageRule {
	out := append([]SlashCoverageRule(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return strings.ToLower(out[i].FaultType) < strings.ToLower(out[j].FaultType) })
	return out
}

func validateClaimID(value string, maxBytes uint32) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return errors.New("validator insurance claim id is required")
	}
	if uint32(len(value)) > maxBytes || strings.ContainsAny(value, " \t\r\n") {
		return errors.New("validator insurance claim id is invalid")
	}
	return nil
}

func isClaimStatus(status string) bool {
	return status == ClaimStatusPending || status == ClaimStatusPaid || status == ClaimStatusRejected
}

func isValidatorStatus(status string) bool {
	return status == validatorregistrytypes.StatusCandidate ||
		status == validatorregistrytypes.StatusActive ||
		status == validatorregistrytypes.StatusJailed ||
		status == validatorregistrytypes.StatusTombstoned ||
		status == validatorregistrytypes.StatusRetired
}
