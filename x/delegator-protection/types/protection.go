package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"
)

const (
	ModuleName	= "delegator-protection"
	StoreKey	= ModuleName

	DefaultProtectionAuthority	= "4:0000000000000000000000000000000000000000000000000000000000000001"
	DefaultBondDenom		= "naet"
	BasisPoints			= uint32(10_000)

	ClaimStatusSubmitted	= "submitted"
	ClaimStatusApproved	= "approved"
	ClaimStatusRejected	= "rejected"
	ClaimStatusPaid		= "paid"
)

type ProtectionParams struct {
	Authority		string
	Denom			string
	IncomingFeeShareBps	uint32
	TreasuryDistributionBps	uint32
	ReserveFloor		sdkmath.Int
	MaxPayoutPerEpoch	sdkmath.Int
	MinClaimLoss		sdkmath.Int
	MaxActiveClaims		uint32
	MaxPayoutHistory	uint32
	RequireEligibilityHash	bool
}

type ProtectionFund struct {
	Balance			sdkmath.Int
	IncomingFeeShare	uint32
	ReserveFloor		sdkmath.Int
	Denom			string
	FundHash		string
}

type ProtectionClaim struct {
	ClaimID		string
	Delegator	string
	Validator	string
	LossAmount	sdkmath.Int
	RequestedPayout	sdkmath.Int
	ApprovedPayout	sdkmath.Int
	EligibilityHash	string
	Reason		string
	Epoch		uint64
	SubmittedHeight	uint64
	ApprovedHeight	uint64
	Status		string
	Paid		bool
	ClaimHash	string
}

type ProtectionPayout struct {
	PayoutID	string
	ClaimID		string
	Delegator	string
	Amount		sdkmath.Int
	Epoch		uint64
	Height		uint64
	PayoutHash	string
}

type DelegatorProtectionState struct {
	Params	ProtectionParams
	Fund	ProtectionFund
	Claims	[]ProtectionClaim
	Payouts	[]ProtectionPayout
}

type FeeShareAllocation struct {
	CollectedFee	sdkmath.Int
	AddedToFund	sdkmath.Int
	Remainder	sdkmath.Int
	ShareBps	uint32
}

type MsgSubmitDelegatorProtectionClaim struct {
	Delegator	string
	Validator	string
	LossAmount	sdkmath.Int
	RequestedPayout	sdkmath.Int
	EligibilityHash	string
	Reason		string
	Epoch		uint64
	Height		uint64
}

type MsgApproveDelegatorProtectionClaim struct {
	Authority	string
	ClaimID		string
	ApprovedPayout	sdkmath.Int
	Height		uint64
}

type MsgRejectDelegatorProtectionClaim struct {
	Authority	string
	ClaimID		string
	Reason		string
	Height		uint64
}

type MsgClaimDelegatorCompensation struct {
	Delegator	string
	ClaimID		string
	Epoch		uint64
	Height		uint64
}

type MsgUpdateProtectionParams struct {
	Authority	string
	Params		ProtectionParams
}

type QueryProtectionClaimsRequest struct {
	Delegator	string
	Status		string
}

type QueryDelegatorCompensationRequest struct {
	Delegator string
}

func DefaultProtectionParams() ProtectionParams {
	return ProtectionParams{
		Authority:			DefaultProtectionAuthority,
		Denom:				DefaultBondDenom,
		IncomingFeeShareBps:		250,
		TreasuryDistributionBps:	1_000,
		ReserveFloor:			sdkmath.NewInt(1_000),
		MaxPayoutPerEpoch:		sdkmath.NewInt(10_000),
		MinClaimLoss:			sdkmath.NewInt(1),
		MaxActiveClaims:		10_000,
		MaxPayoutHistory:		10_000,
		RequireEligibilityHash:		true,
	}
}

func NewDelegatorProtectionState(params ProtectionParams) (DelegatorProtectionState, error) {
	if strings.TrimSpace(params.Authority) == "" {
		params = DefaultProtectionParams()
	}
	if err := params.Validate(); err != nil {
		return DelegatorProtectionState{}, err
	}
	fund := ProtectionFund{
		Balance:		sdkmath.ZeroInt(),
		IncomingFeeShare:	params.IncomingFeeShareBps,
		ReserveFloor:		params.ReserveFloor,
		Denom:			params.Denom,
	}
	fund.FundHash = ComputeProtectionFundHash(fund)
	return DelegatorProtectionState{Params: params, Fund: fund}, nil
}

func (params ProtectionParams) Validate() error {
	if strings.TrimSpace(params.Authority) == "" {
		return errors.New("delegator protection authority is required")
	}
	if strings.TrimSpace(params.Denom) == "" {
		return errors.New("delegator protection denom is required")
	}
	if params.IncomingFeeShareBps > BasisPoints {
		return errors.New("delegator protection fee share exceeds 10000 bps")
	}
	if params.TreasuryDistributionBps > BasisPoints {
		return errors.New("delegator protection treasury distribution exceeds 10000 bps")
	}
	if params.IncomingFeeShareBps+params.TreasuryDistributionBps > BasisPoints {
		return errors.New("delegator protection fee share is incompatible with treasury distribution proportions")
	}
	if normalizeInt(params.ReserveFloor).IsNegative() {
		return errors.New("delegator protection reserve floor cannot be negative")
	}
	if !normalizeInt(params.MaxPayoutPerEpoch).IsPositive() {
		return errors.New("delegator protection max payout per epoch must be positive")
	}
	if !normalizeInt(params.MinClaimLoss).IsPositive() {
		return errors.New("delegator protection min claim loss must be positive")
	}
	if params.MaxActiveClaims == 0 {
		return errors.New("delegator protection max active claims must be positive")
	}
	if params.MaxPayoutHistory == 0 {
		return errors.New("delegator protection max payout history must be positive")
	}
	return nil
}

func (state DelegatorProtectionState) Validate() error {
	state = NormalizeDelegatorProtectionState(state)
	if err := state.Params.Validate(); err != nil {
		return err
	}
	if err := state.Fund.Validate(state.Params); err != nil {
		return err
	}
	if uint32(len(activeClaims(state.Claims))) > state.Params.MaxActiveClaims {
		return errors.New("delegator protection active claim queue exceeds configured max")
	}
	if uint32(len(state.Payouts)) > state.Params.MaxPayoutHistory {
		return errors.New("delegator protection payout history exceeds configured max")
	}
	seenClaims := make(map[string]struct{}, len(state.Claims))
	for _, claim := range state.Claims {
		if err := claim.Validate(state.Params); err != nil {
			return err
		}
		if _, found := seenClaims[claim.ClaimID]; found {
			return errors.New("duplicate delegator protection claim")
		}
		seenClaims[claim.ClaimID] = struct{}{}
	}
	paidClaims := make(map[string]struct{}, len(state.Payouts))
	epochPaid := make(map[uint64]sdkmath.Int)
	for _, payout := range state.Payouts {
		if err := payout.Validate(); err != nil {
			return err
		}
		if _, found := seenClaims[payout.ClaimID]; !found {
			return errors.New("delegator protection payout references unknown claim")
		}
		if _, found := paidClaims[payout.ClaimID]; found {
			return errors.New("delegator protection claim cannot be paid twice")
		}
		paidClaims[payout.ClaimID] = struct{}{}
		current := epochPaid[payout.Epoch]
		epochPaid[payout.Epoch] = normalizeInt(current).Add(payout.Amount)
	}
	for epoch, paid := range epochPaid {
		if paid.GT(state.Params.MaxPayoutPerEpoch) {
			return fmt.Errorf("delegator protection payout exceeds max per epoch %d", epoch)
		}
	}
	return nil
}

func (fund ProtectionFund) Validate(params ProtectionParams) error {
	fund = normalizeFund(fund)
	if normalizeInt(fund.Balance).IsNegative() {
		return errors.New("delegator protection fund cannot go negative")
	}
	if fund.IncomingFeeShare != params.IncomingFeeShareBps {
		return errors.New("delegator protection fund fee share mismatch")
	}
	if fund.Denom != params.Denom {
		return errors.New("delegator protection fund denom mismatch")
	}
	if !fund.ReserveFloor.Equal(params.ReserveFloor) {
		return errors.New("delegator protection fund reserve floor mismatch")
	}
	if !isHex64(fund.FundHash) {
		return errors.New("delegator protection fund hash must be hex")
	}
	if fund.FundHash != ComputeProtectionFundHash(fund) {
		return errors.New("delegator protection fund hash mismatch")
	}
	return nil
}

func (claim ProtectionClaim) Validate(params ProtectionParams) error {
	claim = normalizeClaim(claim)
	if claim.ClaimID == "" || !isHex64(claim.ClaimID) {
		return errors.New("delegator protection claim id must be hex")
	}
	if strings.TrimSpace(claim.Delegator) == "" {
		return errors.New("delegator protection claim delegator is required")
	}
	if strings.TrimSpace(claim.Validator) == "" {
		return errors.New("delegator protection claim validator is required")
	}
	if !claim.LossAmount.IsPositive() || claim.LossAmount.LT(params.MinClaimLoss) {
		return errors.New("delegator protection claim loss is below eligibility rules")
	}
	if !claim.RequestedPayout.IsPositive() {
		return errors.New("delegator protection requested payout must be positive")
	}
	if claim.RequestedPayout.GT(claim.LossAmount) {
		return errors.New("delegator protection requested payout exceeds loss")
	}
	if claim.ApprovedPayout.IsNegative() {
		return errors.New("delegator protection approved payout cannot be negative")
	}
	if claim.ApprovedPayout.GT(claim.RequestedPayout) {
		return errors.New("delegator protection approved payout exceeds request")
	}
	if params.RequireEligibilityHash && !isHex64(claim.EligibilityHash) {
		return errors.New("delegator protection eligibility hash is required")
	}
	if claim.Epoch == 0 {
		return errors.New("delegator protection claim epoch is required")
	}
	if claim.SubmittedHeight == 0 {
		return errors.New("delegator protection claim submitted height is required")
	}
	switch claim.Status {
	case ClaimStatusSubmitted, ClaimStatusApproved, ClaimStatusRejected, ClaimStatusPaid:
	default:
		return fmt.Errorf("delegator protection claim status %q is invalid", claim.Status)
	}
	if claim.Paid && claim.Status != ClaimStatusPaid {
		return errors.New("delegator protection paid claim must have paid status")
	}
	if !claim.Paid && claim.Status == ClaimStatusPaid {
		return errors.New("delegator protection paid status requires paid flag")
	}
	if !isHex64(claim.ClaimHash) {
		return errors.New("delegator protection claim hash must be hex")
	}
	if claim.ClaimHash != ComputeProtectionClaimHash(claim) {
		return errors.New("delegator protection claim hash mismatch")
	}
	return nil
}

func (payout ProtectionPayout) Validate() error {
	payout = normalizePayout(payout)
	if !isHex64(payout.PayoutID) {
		return errors.New("delegator protection payout id must be hex")
	}
	if !isHex64(payout.ClaimID) {
		return errors.New("delegator protection payout claim id must be hex")
	}
	if strings.TrimSpace(payout.Delegator) == "" {
		return errors.New("delegator protection payout delegator is required")
	}
	if !payout.Amount.IsPositive() {
		return errors.New("delegator protection payout amount must be positive")
	}
	if payout.Epoch == 0 {
		return errors.New("delegator protection payout epoch is required")
	}
	if payout.Height == 0 {
		return errors.New("delegator protection payout height is required")
	}
	if !isHex64(payout.PayoutHash) {
		return errors.New("delegator protection payout hash must be hex")
	}
	if payout.PayoutID != ComputeProtectionPayoutID(payout) {
		return errors.New("delegator protection payout id mismatch")
	}
	if payout.PayoutHash != ComputeProtectionPayoutHash(payout) {
		return errors.New("delegator protection payout hash mismatch")
	}
	return nil
}

func ApplyProtocolFeeShare(state DelegatorProtectionState, collectedFee sdkmath.Int) (DelegatorProtectionState, FeeShareAllocation, error) {
	state = NormalizeDelegatorProtectionState(state)
	if err := state.Validate(); err != nil {
		return DelegatorProtectionState{}, FeeShareAllocation{}, err
	}
	collectedFee = normalizeInt(collectedFee)
	if collectedFee.IsNegative() {
		return DelegatorProtectionState{}, FeeShareAllocation{}, errors.New("delegator protection collected fee cannot be negative")
	}
	added := collectedFee.MulRaw(int64(state.Params.IncomingFeeShareBps)).QuoRaw(int64(BasisPoints))
	state.Fund.Balance = state.Fund.Balance.Add(added)
	state.Fund.FundHash = ComputeProtectionFundHash(state.Fund)
	allocation := FeeShareAllocation{
		CollectedFee:	collectedFee,
		AddedToFund:	added,
		Remainder:	collectedFee.Sub(added),
		ShareBps:	state.Params.IncomingFeeShareBps,
	}
	state = NormalizeDelegatorProtectionState(state)
	return state, allocation, state.Validate()
}

func ApplySubmitDelegatorProtectionClaim(state DelegatorProtectionState, msg MsgSubmitDelegatorProtectionClaim) (DelegatorProtectionState, ProtectionClaim, error) {
	state = NormalizeDelegatorProtectionState(state)
	if err := state.Validate(); err != nil {
		return DelegatorProtectionState{}, ProtectionClaim{}, err
	}
	claim := ProtectionClaim{
		Delegator:		strings.TrimSpace(msg.Delegator),
		Validator:		strings.TrimSpace(msg.Validator),
		LossAmount:		normalizeInt(msg.LossAmount),
		RequestedPayout:	normalizeInt(msg.RequestedPayout),
		EligibilityHash:	strings.ToLower(strings.TrimSpace(msg.EligibilityHash)),
		Reason:			strings.TrimSpace(msg.Reason),
		Epoch:			msg.Epoch,
		SubmittedHeight:	msg.Height,
		Status:			ClaimStatusSubmitted,
	}
	if claim.RequestedPayout.IsNil() || !claim.RequestedPayout.IsPositive() {
		claim.RequestedPayout = claim.LossAmount
	}
	claim.ClaimID = ComputeProtectionClaimID(claim)
	claim.ClaimHash = ComputeProtectionClaimHash(claim)
	if err := claim.Validate(state.Params); err != nil {
		return DelegatorProtectionState{}, ProtectionClaim{}, err
	}
	for _, existing := range state.Claims {
		if existing.ClaimID == claim.ClaimID {
			return DelegatorProtectionState{}, ProtectionClaim{}, errors.New("duplicate delegator protection claim")
		}
	}
	state.Claims = append(state.Claims, claim)
	state = NormalizeDelegatorProtectionState(state)
	return state, claim, state.Validate()
}

func ApplyApproveDelegatorProtectionClaim(state DelegatorProtectionState, msg MsgApproveDelegatorProtectionClaim) (DelegatorProtectionState, error) {
	state = NormalizeDelegatorProtectionState(state)
	if err := authorize(state.Params, msg.Authority); err != nil {
		return DelegatorProtectionState{}, err
	}
	approved := normalizeInt(msg.ApprovedPayout)
	if !approved.IsPositive() {
		return DelegatorProtectionState{}, errors.New("delegator protection approved payout must be positive")
	}
	index, claim, found := state.claimByID(msg.ClaimID)
	if !found {
		return DelegatorProtectionState{}, errors.New("delegator protection claim not found")
	}
	if claim.Status != ClaimStatusSubmitted {
		return DelegatorProtectionState{}, errors.New("delegator protection claim is not submitted")
	}
	if approved.GT(claim.RequestedPayout) {
		return DelegatorProtectionState{}, errors.New("delegator protection approved payout exceeds request")
	}
	claim.ApprovedPayout = approved
	claim.ApprovedHeight = msg.Height
	claim.Status = ClaimStatusApproved
	claim.ClaimHash = ComputeProtectionClaimHash(claim)
	state.Claims[index] = claim
	state = NormalizeDelegatorProtectionState(state)
	return state, state.Validate()
}

func ApplyRejectDelegatorProtectionClaim(state DelegatorProtectionState, msg MsgRejectDelegatorProtectionClaim) (DelegatorProtectionState, error) {
	state = NormalizeDelegatorProtectionState(state)
	if err := authorize(state.Params, msg.Authority); err != nil {
		return DelegatorProtectionState{}, err
	}
	index, claim, found := state.claimByID(msg.ClaimID)
	if !found {
		return DelegatorProtectionState{}, errors.New("delegator protection claim not found")
	}
	if claim.Paid {
		return DelegatorProtectionState{}, errors.New("delegator protection paid claim cannot be rejected")
	}
	claim.Status = ClaimStatusRejected
	claim.Reason = strings.TrimSpace(msg.Reason)
	if claim.Reason == "" {
		claim.Reason = "rejected"
	}
	claim.ApprovedHeight = msg.Height
	claim.ClaimHash = ComputeProtectionClaimHash(claim)
	state.Claims[index] = claim
	state = NormalizeDelegatorProtectionState(state)
	return state, state.Validate()
}

func ApplyClaimDelegatorCompensation(state DelegatorProtectionState, msg MsgClaimDelegatorCompensation) (DelegatorProtectionState, ProtectionPayout, error) {
	state = NormalizeDelegatorProtectionState(state)
	index, claim, found := state.claimByID(msg.ClaimID)
	if !found {
		return DelegatorProtectionState{}, ProtectionPayout{}, errors.New("delegator protection claim not found")
	}
	if claim.Delegator != strings.TrimSpace(msg.Delegator) {
		return DelegatorProtectionState{}, ProtectionPayout{}, errors.New("delegator protection claim delegator mismatch")
	}
	if claim.Status != ClaimStatusApproved || claim.Paid {
		return DelegatorProtectionState{}, ProtectionPayout{}, errors.New("delegator protection claim is not payable")
	}
	if msg.Epoch == 0 || msg.Height == 0 {
		return DelegatorProtectionState{}, ProtectionPayout{}, errors.New("delegator protection payout epoch and height are required")
	}
	amount := claim.ApprovedPayout
	if amount.GT(state.Params.MaxPayoutPerEpoch) {
		return DelegatorProtectionState{}, ProtectionPayout{}, errors.New("delegator protection payout exceeds max per epoch")
	}
	paidThisEpoch := state.paidInEpoch(msg.Epoch)
	if paidThisEpoch.Add(amount).GT(state.Params.MaxPayoutPerEpoch) {
		return DelegatorProtectionState{}, ProtectionPayout{}, errors.New("delegator protection payout exceeds max per epoch")
	}
	if state.Fund.Balance.Sub(amount).LT(state.Params.ReserveFloor) {
		return DelegatorProtectionState{}, ProtectionPayout{}, errors.New("delegator protection fund below reserve floor")
	}
	state.Fund.Balance = state.Fund.Balance.Sub(amount)
	state.Fund.FundHash = ComputeProtectionFundHash(state.Fund)
	claim.Status = ClaimStatusPaid
	claim.Paid = true
	claim.ClaimHash = ComputeProtectionClaimHash(claim)
	state.Claims[index] = claim
	payout := ProtectionPayout{
		ClaimID:	claim.ClaimID,
		Delegator:	claim.Delegator,
		Amount:		amount,
		Epoch:		msg.Epoch,
		Height:		msg.Height,
	}
	payout.PayoutID = ComputeProtectionPayoutID(payout)
	payout.PayoutHash = ComputeProtectionPayoutHash(payout)
	state.Payouts = append(state.Payouts, payout)
	state = NormalizeDelegatorProtectionState(state)
	return state, payout, state.Validate()
}

func ApplyUpdateProtectionParams(state DelegatorProtectionState, msg MsgUpdateProtectionParams) (DelegatorProtectionState, error) {
	state = NormalizeDelegatorProtectionState(state)
	if err := authorize(state.Params, msg.Authority); err != nil {
		return DelegatorProtectionState{}, err
	}
	next := msg.Params
	if strings.TrimSpace(next.Authority) == "" {
		next.Authority = state.Params.Authority
	}
	if strings.TrimSpace(next.Denom) == "" {
		next.Denom = state.Params.Denom
	}
	if err := next.Validate(); err != nil {
		return DelegatorProtectionState{}, err
	}
	state.Params = next
	state.Fund.IncomingFeeShare = next.IncomingFeeShareBps
	state.Fund.ReserveFloor = next.ReserveFloor
	state.Fund.Denom = next.Denom
	state.Fund.FundHash = ComputeProtectionFundHash(state.Fund)
	state = NormalizeDelegatorProtectionState(state)
	return state, state.Validate()
}

func QueryProtectionFund(state DelegatorProtectionState) ProtectionFund {
	return NormalizeDelegatorProtectionState(state).Fund
}

func QueryProtectionClaims(state DelegatorProtectionState, req QueryProtectionClaimsRequest) []ProtectionClaim {
	state = NormalizeDelegatorProtectionState(state)
	delegator := strings.TrimSpace(req.Delegator)
	status := strings.TrimSpace(req.Status)
	out := make([]ProtectionClaim, 0, len(state.Claims))
	for _, claim := range state.Claims {
		if delegator != "" && claim.Delegator != delegator {
			continue
		}
		if status != "" && claim.Status != status {
			continue
		}
		out = append(out, claim)
	}
	return normalizeClaims(out)
}

func QueryDelegatorCompensation(state DelegatorProtectionState, req QueryDelegatorCompensationRequest) []ProtectionPayout {
	state = NormalizeDelegatorProtectionState(state)
	delegator := strings.TrimSpace(req.Delegator)
	out := make([]ProtectionPayout, 0, len(state.Payouts))
	for _, payout := range state.Payouts {
		if delegator == "" || payout.Delegator == delegator {
			out = append(out, payout)
		}
	}
	return normalizePayouts(out)
}

func QueryProtectionParams(state DelegatorProtectionState) ProtectionParams {
	return NormalizeDelegatorProtectionState(state).Params
}

func ExportDelegatorProtectionState(state DelegatorProtectionState) (DelegatorProtectionState, error) {
	state = NormalizeDelegatorProtectionState(state)
	if err := state.Validate(); err != nil {
		return DelegatorProtectionState{}, err
	}
	return cloneState(state), nil
}

func ImportDelegatorProtectionState(exported DelegatorProtectionState) (DelegatorProtectionState, error) {
	exported = NormalizeDelegatorProtectionState(exported)
	if err := exported.Validate(); err != nil {
		return DelegatorProtectionState{}, err
	}
	return cloneState(exported), nil
}

func CheckDelegatorProtectionInvariants(state DelegatorProtectionState) error {
	state = NormalizeDelegatorProtectionState(state)
	return state.Validate()
}

func NormalizeDelegatorProtectionState(state DelegatorProtectionState) DelegatorProtectionState {
	if strings.TrimSpace(state.Params.Authority) == "" {
		state.Params = DefaultProtectionParams()
	}
	state.Params.Authority = strings.TrimSpace(state.Params.Authority)
	state.Params.Denom = strings.TrimSpace(state.Params.Denom)
	state.Params.ReserveFloor = normalizeInt(state.Params.ReserveFloor)
	state.Params.MaxPayoutPerEpoch = normalizeInt(state.Params.MaxPayoutPerEpoch)
	state.Params.MinClaimLoss = normalizeInt(state.Params.MinClaimLoss)
	state.Fund = normalizeFund(state.Fund)
	if state.Fund.Denom == "" {
		state.Fund.Denom = state.Params.Denom
	}
	if state.Fund.IncomingFeeShare == 0 {
		state.Fund.IncomingFeeShare = state.Params.IncomingFeeShareBps
	}
	if state.Fund.ReserveFloor.IsNil() {
		state.Fund.ReserveFloor = state.Params.ReserveFloor
	}
	if state.Fund.FundHash == "" {
		state.Fund.FundHash = ComputeProtectionFundHash(state.Fund)
	}
	state.Claims = normalizeClaims(state.Claims)
	state.Payouts = normalizePayouts(state.Payouts)
	return state
}

func (state DelegatorProtectionState) claimByID(claimID string) (int, ProtectionClaim, bool) {
	claimID = strings.ToLower(strings.TrimSpace(claimID))
	for i, claim := range state.Claims {
		if claim.ClaimID == claimID {
			return i, claim, true
		}
	}
	return -1, ProtectionClaim{}, false
}

func (state DelegatorProtectionState) paidInEpoch(epoch uint64) sdkmath.Int {
	total := sdkmath.ZeroInt()
	for _, payout := range state.Payouts {
		if payout.Epoch == epoch {
			total = total.Add(payout.Amount)
		}
	}
	return total
}

func authorize(params ProtectionParams, authority string) error {
	if err := params.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(authority) != params.Authority {
		return errors.New("delegator protection message requires authority")
	}
	return nil
}

func activeClaims(claims []ProtectionClaim) []ProtectionClaim {
	out := make([]ProtectionClaim, 0, len(claims))
	for _, claim := range claims {
		if claim.Status == ClaimStatusSubmitted || claim.Status == ClaimStatusApproved {
			out = append(out, claim)
		}
	}
	return out
}

func normalizeFund(fund ProtectionFund) ProtectionFund {
	fund.Balance = normalizeInt(fund.Balance)
	fund.ReserveFloor = normalizeInt(fund.ReserveFloor)
	fund.Denom = strings.TrimSpace(fund.Denom)
	fund.FundHash = strings.ToLower(strings.TrimSpace(fund.FundHash))
	return fund
}

func normalizeClaim(claim ProtectionClaim) ProtectionClaim {
	claim.ClaimID = strings.ToLower(strings.TrimSpace(claim.ClaimID))
	claim.Delegator = strings.TrimSpace(claim.Delegator)
	claim.Validator = strings.TrimSpace(claim.Validator)
	claim.LossAmount = normalizeInt(claim.LossAmount)
	claim.RequestedPayout = normalizeInt(claim.RequestedPayout)
	claim.ApprovedPayout = normalizeInt(claim.ApprovedPayout)
	claim.EligibilityHash = strings.ToLower(strings.TrimSpace(claim.EligibilityHash))
	claim.Reason = strings.TrimSpace(claim.Reason)
	claim.Status = strings.TrimSpace(claim.Status)
	if claim.Status == "" {
		claim.Status = ClaimStatusSubmitted
	}
	claim.ClaimHash = strings.ToLower(strings.TrimSpace(claim.ClaimHash))
	if claim.ClaimID == "" {
		claim.ClaimID = ComputeProtectionClaimID(claim)
	}
	if claim.ClaimHash == "" {
		claim.ClaimHash = ComputeProtectionClaimHash(claim)
	}
	return claim
}

func normalizeClaims(claims []ProtectionClaim) []ProtectionClaim {
	out := make([]ProtectionClaim, len(claims))
	for i, claim := range claims {
		out[i] = normalizeClaim(claim)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Epoch != out[j].Epoch {
			return out[i].Epoch < out[j].Epoch
		}
		return out[i].ClaimID < out[j].ClaimID
	})
	return out
}

func normalizePayout(payout ProtectionPayout) ProtectionPayout {
	payout.PayoutID = strings.ToLower(strings.TrimSpace(payout.PayoutID))
	payout.ClaimID = strings.ToLower(strings.TrimSpace(payout.ClaimID))
	payout.Delegator = strings.TrimSpace(payout.Delegator)
	payout.Amount = normalizeInt(payout.Amount)
	payout.PayoutHash = strings.ToLower(strings.TrimSpace(payout.PayoutHash))
	if payout.PayoutID == "" {
		payout.PayoutID = ComputeProtectionPayoutID(payout)
	}
	if payout.PayoutHash == "" {
		payout.PayoutHash = ComputeProtectionPayoutHash(payout)
	}
	return payout
}

func normalizePayouts(payouts []ProtectionPayout) []ProtectionPayout {
	out := make([]ProtectionPayout, len(payouts))
	for i, payout := range payouts {
		out[i] = normalizePayout(payout)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Epoch != out[j].Epoch {
			return out[i].Epoch < out[j].Epoch
		}
		return out[i].PayoutID < out[j].PayoutID
	})
	return out
}

func cloneState(state DelegatorProtectionState) DelegatorProtectionState {
	state = NormalizeDelegatorProtectionState(state)
	return DelegatorProtectionState{
		Params:		state.Params,
		Fund:		state.Fund,
		Claims:		append([]ProtectionClaim(nil), state.Claims...),
		Payouts:	append([]ProtectionPayout(nil), state.Payouts...),
	}
}

func ComputeProtectionFundHash(fund ProtectionFund) string {
	fund = normalizeFund(fund)
	fund.FundHash = ""
	return protectionHashParts("delegator-protection-fund-v1", fund.Denom, fund.Balance.String(), fmt.Sprint(fund.IncomingFeeShare), fund.ReserveFloor.String())
}

func ComputeProtectionClaimID(claim ProtectionClaim) string {
	claim.Delegator = strings.TrimSpace(claim.Delegator)
	claim.Validator = strings.TrimSpace(claim.Validator)
	claim.LossAmount = normalizeInt(claim.LossAmount)
	claim.RequestedPayout = normalizeInt(claim.RequestedPayout)
	claim.EligibilityHash = strings.ToLower(strings.TrimSpace(claim.EligibilityHash))
	return protectionHashParts("delegator-protection-claim-id-v1", claim.Delegator, claim.Validator, claim.LossAmount.String(), claim.RequestedPayout.String(), claim.EligibilityHash, fmt.Sprint(claim.Epoch), fmt.Sprint(claim.SubmittedHeight))
}

func ComputeProtectionClaimHash(claim ProtectionClaim) string {
	claim = normalizeClaimForHash(claim)
	return protectionHashParts("delegator-protection-claim-v1", claim.ClaimID, claim.Delegator, claim.Validator, claim.LossAmount.String(), claim.RequestedPayout.String(), claim.ApprovedPayout.String(), claim.EligibilityHash, claim.Reason, fmt.Sprint(claim.Epoch), fmt.Sprint(claim.SubmittedHeight), fmt.Sprint(claim.ApprovedHeight), claim.Status, fmt.Sprint(claim.Paid))
}

func ComputeProtectionPayoutID(payout ProtectionPayout) string {
	payout.ClaimID = strings.ToLower(strings.TrimSpace(payout.ClaimID))
	payout.Delegator = strings.TrimSpace(payout.Delegator)
	payout.Amount = normalizeInt(payout.Amount)
	return protectionHashParts("delegator-protection-payout-id-v1", payout.ClaimID, payout.Delegator, payout.Amount.String(), fmt.Sprint(payout.Epoch), fmt.Sprint(payout.Height))
}

func ComputeProtectionPayoutHash(payout ProtectionPayout) string {
	payout = normalizePayoutForHash(payout)
	return protectionHashParts("delegator-protection-payout-v1", payout.PayoutID, payout.ClaimID, payout.Delegator, payout.Amount.String(), fmt.Sprint(payout.Epoch), fmt.Sprint(payout.Height))
}

func normalizeClaimForHash(claim ProtectionClaim) ProtectionClaim {
	claim.ClaimID = strings.ToLower(strings.TrimSpace(claim.ClaimID))
	claim.Delegator = strings.TrimSpace(claim.Delegator)
	claim.Validator = strings.TrimSpace(claim.Validator)
	claim.LossAmount = normalizeInt(claim.LossAmount)
	claim.RequestedPayout = normalizeInt(claim.RequestedPayout)
	claim.ApprovedPayout = normalizeInt(claim.ApprovedPayout)
	claim.EligibilityHash = strings.ToLower(strings.TrimSpace(claim.EligibilityHash))
	claim.Reason = strings.TrimSpace(claim.Reason)
	claim.Status = strings.TrimSpace(claim.Status)
	if claim.Status == "" {
		claim.Status = ClaimStatusSubmitted
	}
	claim.ClaimHash = ""
	return claim
}

func normalizePayoutForHash(payout ProtectionPayout) ProtectionPayout {
	payout.PayoutID = strings.ToLower(strings.TrimSpace(payout.PayoutID))
	payout.ClaimID = strings.ToLower(strings.TrimSpace(payout.ClaimID))
	payout.Delegator = strings.TrimSpace(payout.Delegator)
	payout.Amount = normalizeInt(payout.Amount)
	payout.PayoutHash = ""
	return payout
}

func protectionHashParts(parts ...string) string {
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

func normalizeInt(value sdkmath.Int) sdkmath.Int {
	if value.IsNil() {
		return sdkmath.ZeroInt()
	}
	return value
}

func isHex64(value string) bool {
	if len(value) != 64 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}
