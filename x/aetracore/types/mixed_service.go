package types

import (
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strings"
)

type MixedServiceFaultClass string
type MixedDisputeStatus string
type MixedSettlementStatus string
type MixedVerificationHookType string

const (
	MixedFaultLow		MixedServiceFaultClass	= "LOW"
	MixedFaultMedium	MixedServiceFaultClass	= "MEDIUM"
	MixedFaultHigh		MixedServiceFaultClass	= "HIGH"
	MixedFaultCritical	MixedServiceFaultClass	= "CRITICAL"

	MixedDisputeOpen	MixedDisputeStatus	= "OPEN"
	MixedDisputeRejected	MixedDisputeStatus	= "REJECTED"
	MixedDisputeProven	MixedDisputeStatus	= "PROVEN"
	MixedDisputeFallback	MixedDisputeStatus	= "FALLBACK"
	MixedDisputeExpired	MixedDisputeStatus	= "EXPIRED"

	MixedSettlementReleased		MixedSettlementStatus	= "RELEASED"
	MixedSettlementRefunded		MixedSettlementStatus	= "REFUNDED"
	MixedSettlementPenalized	MixedSettlementStatus	= "PENALIZED"
	MixedSettlementFallback		MixedSettlementStatus	= "FALLBACK_EXECUTED"

	MixedHookProofVerification	MixedVerificationHookType	= "PROOF_VERIFICATION"
	MixedHookRecompute		MixedVerificationHookType	= "RECOMPUTE"
	MixedHookFallbackExecution	MixedVerificationHookType	= "FALLBACK_EXECUTION"
)

type MixedServiceState struct {
	ServiceID			string
	ProviderKey			string
	PaymentDenom			string
	EscrowID			string
	EscrowAmount			string
	CollateralDenom			string
	CollateralAmount		string
	RequiredCollateralAmount	string
	FaultClass			MixedServiceFaultClass
	ChallengeWindow			uint64
	FallbackServiceID		string
	FallbackMethodID		string
	FallbackDeterministic		bool
	Anchors				[]MixedResultAnchor
	Disputes			[]MixedDispute
	Settlements			[]MixedSettlement
	StateHash			string
}

type MixedResultAnchor struct {
	AnchorID		string
	ServiceID		string
	CallID			string
	RequestCommitment	string
	ResultCommitment	string
	ReceiptCommitment	string
	ProviderKey		string
	Height			uint64
	ChallengeEndHeight	uint64
	PaymentAmount		string
	AnchorHash		string
}

type MixedChallengeMessage struct {
	AnchorID		string
	Challenger		string
	ChallengeCommitment	string
	OpenedHeight		uint64
	VerificationHook	MixedVerificationHook
}

type MixedDispute struct {
	DisputeID		string
	AnchorID		string
	ServiceID		string
	Challenger		string
	ChallengeCommitment	string
	OpenedHeight		uint64
	ResolveByHeight		uint64
	Status			MixedDisputeStatus
	VerificationHook	MixedVerificationHook
	DisputeHash		string
}

type MixedVerificationHook struct {
	HookType		MixedVerificationHookType
	TargetServiceID		string
	MethodID		string
	ProofCommitment		string
	Deterministic		bool
	ProofMeterGas		uint64
	ExpectedResultHash	string
	HookHash		string
}

type MixedDisputeResolution struct {
	DisputeID		string
	Resolver		string
	ResolvedHeight		uint64
	ProofAccepted		bool
	RecomputeMatches	bool
	FallbackExecuted	bool
	ResolutionHash		string
}

type MixedSettlement struct {
	SettlementID		string
	AnchorID		string
	DisputeID		string
	ServiceID		string
	ProviderKey		string
	Status			MixedSettlementStatus
	SettledHeight		uint64
	PaymentDenom		string
	PaymentAmount		string
	PaymentStatus		ServicePaymentStatus
	PenaltyDenom		string
	PenaltyAmount		string
	PenaltyRecipient	string
	SettlementHash		string
}

func NewMixedServiceState(descriptor ServiceDescriptor, providerKey string, faultClass MixedServiceFaultClass, requiredCollateralAmount string) (MixedServiceState, error) {
	descriptor = CanonicalServiceDescriptor(descriptor)
	if err := descriptor.Validate(); err != nil {
		return MixedServiceState{}, err
	}
	if descriptor.ServiceType != ServiceTypeMixed {
		return MixedServiceState{}, errors.New("aetracore mixed state requires mixed service descriptor")
	}
	state := MixedServiceState{
		ServiceID:			descriptor.ServiceID,
		ProviderKey:			strings.TrimSpace(providerKey),
		PaymentDenom:			descriptor.Payment.Denom,
		EscrowID:			descriptor.Payment.EscrowID,
		EscrowAmount:			descriptor.Payment.Amount,
		CollateralDenom:		descriptor.Verification.ProviderCollateralDenom,
		CollateralAmount:		descriptor.Verification.ProviderCollateralAmount,
		RequiredCollateralAmount:	strings.TrimSpace(requiredCollateralAmount),
		FaultClass:			faultClass,
		ChallengeWindow:		descriptor.Verification.ChallengeWindow,
		FallbackServiceID:		descriptor.Verification.FallbackServiceID,
		FallbackDeterministic:		descriptor.Execution.FailureBehavior == ServiceFailureFallbackOnChain || descriptor.Verification.FallbackServiceID != "",
		Anchors:			[]MixedResultAnchor{},
		Disputes:			[]MixedDispute{},
		Settlements:			[]MixedSettlement{},
	}
	if state.ChallengeWindow == 0 {
		state.ChallengeWindow = descriptor.Execution.ChallengeWindow
	}
	if len(descriptor.Interface.Methods) > 0 {
		state.FallbackMethodID = descriptor.Interface.Methods[0].MethodID
	}
	state.StateHash = ComputeMixedServiceStateHash(state)
	return state, state.Validate()
}

func (state MixedServiceState) Validate() error {
	if err := validatePolicyID("aetracore mixed state service id", state.ServiceID); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore mixed state provider key", state.ProviderKey); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore mixed state payment denom", state.PaymentDenom); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore mixed state escrow id", state.EscrowID); err != nil {
		return err
	}
	if err := validateAmountString("aetracore mixed state escrow amount", state.EscrowAmount); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore mixed state collateral denom", state.CollateralDenom); err != nil {
		return err
	}
	if err := validateAmountString("aetracore mixed state collateral amount", state.CollateralAmount); err != nil {
		return err
	}
	if err := validateAmountString("aetracore mixed state required collateral", state.RequiredCollateralAmount); err != nil {
		return err
	}
	if !IsMixedFaultClass(state.FaultClass) {
		return fmt.Errorf("unknown aetracore mixed fault class %q", state.FaultClass)
	}
	if err := mixedAmountAtLeast("aetracore mixed provider collateral", state.CollateralAmount, state.RequiredCollateralAmount); err != nil {
		return err
	}
	if state.ChallengeWindow == 0 {
		return errors.New("aetracore mixed service challenge window must be explicit")
	}
	if !state.FallbackDeterministic {
		return errors.New("aetracore mixed fallback path must be deterministic")
	}
	if state.FallbackServiceID != "" {
		if err := validatePolicyID("aetracore mixed fallback service id", state.FallbackServiceID); err != nil {
			return err
		}
	}
	if state.FallbackMethodID != "" {
		if err := validatePolicyID("aetracore mixed fallback method id", state.FallbackMethodID); err != nil {
			return err
		}
	}
	if err := validateMixedAnchors(state.Anchors); err != nil {
		return err
	}
	if err := validateMixedDisputes(state.Disputes); err != nil {
		return err
	}
	if err := validateMixedSettlements(state.Settlements); err != nil {
		return err
	}
	if err := ValidateHash("aetracore mixed state hash", state.StateHash); err != nil {
		return err
	}
	if expected := ComputeMixedServiceStateHash(state); state.StateHash != expected {
		return fmt.Errorf("aetracore mixed state hash mismatch: expected %s", expected)
	}
	return nil
}

func AnchorMixedServiceResult(state MixedServiceState, anchor MixedResultAnchor) (MixedServiceState, MixedResultAnchor, error) {
	if err := state.Validate(); err != nil {
		return MixedServiceState{}, MixedResultAnchor{}, err
	}
	anchor = CanonicalMixedResultAnchor(anchor)
	if anchor.ServiceID == "" {
		anchor.ServiceID = state.ServiceID
	}
	if anchor.ProviderKey == "" {
		anchor.ProviderKey = state.ProviderKey
	}
	if anchor.ChallengeEndHeight == 0 && anchor.Height != 0 {
		anchor.ChallengeEndHeight = anchor.Height + state.ChallengeWindow
	}
	if anchor.PaymentAmount == "" {
		anchor.PaymentAmount = state.EscrowAmount
	}
	anchor.AnchorID = ComputeMixedResultAnchorID(anchor)
	anchor.AnchorHash = ComputeMixedResultAnchorHash(anchor)
	if err := anchor.ValidateForState(state); err != nil {
		return MixedServiceState{}, MixedResultAnchor{}, err
	}
	if _, found := state.AnchorByID(anchor.AnchorID); found {
		return MixedServiceState{}, MixedResultAnchor{}, fmt.Errorf("aetracore mixed result anchor %s already exists", anchor.AnchorID)
	}
	next := state.clone()
	next.Anchors = append(next.Anchors, anchor)
	sortMixedAnchors(next.Anchors)
	next.StateHash = ComputeMixedServiceStateHash(next)
	return next, anchor, next.Validate()
}

func (anchor MixedResultAnchor) ValidateForState(state MixedServiceState) error {
	if err := anchor.Validate(); err != nil {
		return err
	}
	if anchor.ServiceID != state.ServiceID {
		return errors.New("aetracore mixed result anchor service mismatch")
	}
	if anchor.ProviderKey != state.ProviderKey {
		return errors.New("aetracore mixed result anchor provider mismatch")
	}
	if anchor.ChallengeEndHeight != anchor.Height+state.ChallengeWindow {
		return errors.New("aetracore mixed result anchor challenge period mismatch")
	}
	return mixedAmountAtMost("aetracore mixed result payment", anchor.PaymentAmount, state.EscrowAmount)
}

func (anchor MixedResultAnchor) Validate() error {
	if err := ValidateHash("aetracore mixed result anchor id", anchor.AnchorID); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore mixed result service id", anchor.ServiceID); err != nil {
		return err
	}
	if err := validateOptionalHash("aetracore mixed result call id", anchor.CallID); err != nil {
		return err
	}
	if err := ValidateHash("aetracore mixed request commitment", anchor.RequestCommitment); err != nil {
		return err
	}
	if err := ValidateHash("aetracore mixed result commitment", anchor.ResultCommitment); err != nil {
		return err
	}
	if err := ValidateHash("aetracore mixed receipt commitment", anchor.ReceiptCommitment); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore mixed result provider key", anchor.ProviderKey); err != nil {
		return err
	}
	if anchor.Height == 0 {
		return errors.New("aetracore mixed result anchor height must be positive")
	}
	if anchor.ChallengeEndHeight <= anchor.Height {
		return errors.New("aetracore mixed result challenge end must exceed anchor height")
	}
	if err := validateAmountString("aetracore mixed result payment amount", anchor.PaymentAmount); err != nil {
		return err
	}
	if err := ValidateHash("aetracore mixed result anchor hash", anchor.AnchorHash); err != nil {
		return err
	}
	if expected := ComputeMixedResultAnchorHash(anchor); anchor.AnchorHash != expected {
		return fmt.Errorf("aetracore mixed result anchor hash mismatch: expected %s", expected)
	}
	return nil
}

func OpenMixedServiceChallenge(state MixedServiceState, message MixedChallengeMessage) (MixedServiceState, MixedDispute, error) {
	if err := state.Validate(); err != nil {
		return MixedServiceState{}, MixedDispute{}, err
	}
	anchor, found := state.AnchorByID(message.AnchorID)
	if !found {
		return MixedServiceState{}, MixedDispute{}, fmt.Errorf("aetracore mixed anchor %s not found", message.AnchorID)
	}
	if message.OpenedHeight == 0 || message.OpenedHeight > anchor.ChallengeEndHeight {
		return MixedServiceState{}, MixedDispute{}, errors.New("aetracore mixed challenge is outside challenge window")
	}
	hook := CanonicalMixedVerificationHook(message.VerificationHook)
	if hook.TargetServiceID == "" {
		hook.TargetServiceID = state.FallbackServiceID
	}
	if hook.MethodID == "" {
		hook.MethodID = state.FallbackMethodID
	}
	hook.HookHash = ComputeMixedVerificationHookHash(hook)
	dispute := MixedDispute{
		AnchorID:		message.AnchorID,
		ServiceID:		state.ServiceID,
		Challenger:		strings.TrimSpace(message.Challenger),
		ChallengeCommitment:	strings.ToLower(strings.TrimSpace(message.ChallengeCommitment)),
		OpenedHeight:		message.OpenedHeight,
		ResolveByHeight:	anchor.ChallengeEndHeight,
		Status:			MixedDisputeOpen,
		VerificationHook:	hook,
	}
	dispute.DisputeID = ComputeMixedDisputeID(dispute)
	dispute.DisputeHash = ComputeMixedDisputeHash(dispute)
	if err := dispute.ValidateForState(state); err != nil {
		return MixedServiceState{}, MixedDispute{}, err
	}
	if _, found := state.DisputeByID(dispute.DisputeID); found {
		return MixedServiceState{}, MixedDispute{}, fmt.Errorf("aetracore mixed dispute %s already exists", dispute.DisputeID)
	}
	next := state.clone()
	next.Disputes = append(next.Disputes, dispute)
	sortMixedDisputes(next.Disputes)
	next.StateHash = ComputeMixedServiceStateHash(next)
	return next, dispute, next.Validate()
}

func (dispute MixedDispute) ValidateForState(state MixedServiceState) error {
	if err := dispute.Validate(); err != nil {
		return err
	}
	if dispute.ServiceID != state.ServiceID {
		return errors.New("aetracore mixed dispute service mismatch")
	}
	if _, found := state.AnchorByID(dispute.AnchorID); !found {
		return errors.New("aetracore mixed dispute references unknown anchor")
	}
	return nil
}

func (dispute MixedDispute) Validate() error {
	if err := ValidateHash("aetracore mixed dispute id", dispute.DisputeID); err != nil {
		return err
	}
	if err := ValidateHash("aetracore mixed dispute anchor id", dispute.AnchorID); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore mixed dispute service id", dispute.ServiceID); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore mixed dispute challenger", dispute.Challenger); err != nil {
		return err
	}
	if err := ValidateHash("aetracore mixed challenge commitment", dispute.ChallengeCommitment); err != nil {
		return err
	}
	if dispute.OpenedHeight == 0 || dispute.ResolveByHeight <= dispute.OpenedHeight {
		return errors.New("aetracore mixed dispute challenge period is invalid")
	}
	if !IsMixedDisputeStatus(dispute.Status) {
		return fmt.Errorf("unknown aetracore mixed dispute status %q", dispute.Status)
	}
	if err := dispute.VerificationHook.Validate(); err != nil {
		return err
	}
	if err := ValidateHash("aetracore mixed dispute hash", dispute.DisputeHash); err != nil {
		return err
	}
	if expected := ComputeMixedDisputeHash(dispute); dispute.DisputeHash != expected {
		return fmt.Errorf("aetracore mixed dispute hash mismatch: expected %s", expected)
	}
	return nil
}

func ResolveMixedServiceChallenge(state MixedServiceState, resolution MixedDisputeResolution) (MixedServiceState, MixedSettlement, error) {
	if err := state.Validate(); err != nil {
		return MixedServiceState{}, MixedSettlement{}, err
	}
	dispute, found := state.DisputeByID(resolution.DisputeID)
	if !found {
		return MixedServiceState{}, MixedSettlement{}, fmt.Errorf("aetracore mixed dispute %s not found", resolution.DisputeID)
	}
	if dispute.Status != MixedDisputeOpen {
		return MixedServiceState{}, MixedSettlement{}, errors.New("aetracore mixed dispute is not open")
	}
	if resolution.ResolvedHeight == 0 || resolution.ResolvedHeight > dispute.ResolveByHeight {
		return MixedServiceState{}, MixedSettlement{}, errors.New("aetracore mixed dispute resolution is outside challenge window")
	}
	if err := resolution.Validate(); err != nil {
		return MixedServiceState{}, MixedSettlement{}, err
	}
	next := state.clone()
	for i := range next.Disputes {
		if next.Disputes[i].DisputeID == dispute.DisputeID {
			next.Disputes[i].Status = mixedDisputeStatusForResolution(resolution)
			next.Disputes[i].DisputeHash = ComputeMixedDisputeHash(next.Disputes[i])
			dispute = next.Disputes[i]
			break
		}
	}
	anchor, _ := next.AnchorByID(dispute.AnchorID)
	settlement := settlementForResolvedDispute(next, anchor, dispute, resolution)
	next.Settlements = append(next.Settlements, settlement)
	sortMixedSettlements(next.Settlements)
	next.StateHash = ComputeMixedServiceStateHash(next)
	return next, settlement, next.Validate()
}

func (resolution MixedDisputeResolution) Validate() error {
	if err := ValidateHash("aetracore mixed dispute resolution id", resolution.DisputeID); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore mixed dispute resolver", resolution.Resolver); err != nil {
		return err
	}
	if resolution.ResolvedHeight == 0 {
		return errors.New("aetracore mixed dispute resolution height must be positive")
	}
	if resolution.ResolutionHash == "" {
		return nil
	}
	if err := ValidateHash("aetracore mixed dispute resolution hash", resolution.ResolutionHash); err != nil {
		return err
	}
	if expected := ComputeMixedDisputeResolutionHash(resolution); resolution.ResolutionHash != expected {
		return fmt.Errorf("aetracore mixed dispute resolution hash mismatch: expected %s", expected)
	}
	return nil
}

func SettleMixedServiceResult(state MixedServiceState, anchorID string, currentHeight uint64) (MixedServiceState, MixedSettlement, error) {
	if err := state.Validate(); err != nil {
		return MixedServiceState{}, MixedSettlement{}, err
	}
	anchor, found := state.AnchorByID(anchorID)
	if !found {
		return MixedServiceState{}, MixedSettlement{}, fmt.Errorf("aetracore mixed anchor %s not found", anchorID)
	}
	if currentHeight <= anchor.ChallengeEndHeight {
		return MixedServiceState{}, MixedSettlement{}, errors.New("aetracore mixed settlement is still in challenge window")
	}
	if state.HasOpenDispute(anchorID) {
		return MixedServiceState{}, MixedSettlement{}, errors.New("aetracore mixed settlement has open dispute")
	}
	if state.HasSettlement(anchorID) {
		return MixedServiceState{}, MixedSettlement{}, errors.New("aetracore mixed anchor is already settled")
	}
	settlement := MixedSettlement{
		AnchorID:	anchor.AnchorID,
		ServiceID:	state.ServiceID,
		ProviderKey:	state.ProviderKey,
		Status:		MixedSettlementReleased,
		SettledHeight:	currentHeight,
		PaymentDenom:	state.PaymentDenom,
		PaymentAmount:	anchor.PaymentAmount,
		PaymentStatus:	ServicePaymentStatusSettled,
	}
	settlement.SettlementID = ComputeMixedSettlementID(settlement)
	settlement.SettlementHash = ComputeMixedSettlementHash(settlement)
	next := state.clone()
	next.Settlements = append(next.Settlements, settlement)
	sortMixedSettlements(next.Settlements)
	next.StateHash = ComputeMixedServiceStateHash(next)
	return next, settlement, next.Validate()
}

func CanonicalMixedResultAnchor(anchor MixedResultAnchor) MixedResultAnchor {
	anchor.ServiceID = strings.TrimSpace(anchor.ServiceID)
	anchor.CallID = strings.ToLower(strings.TrimSpace(anchor.CallID))
	anchor.RequestCommitment = strings.ToLower(strings.TrimSpace(anchor.RequestCommitment))
	anchor.ResultCommitment = strings.ToLower(strings.TrimSpace(anchor.ResultCommitment))
	anchor.ReceiptCommitment = strings.ToLower(strings.TrimSpace(anchor.ReceiptCommitment))
	anchor.ProviderKey = strings.TrimSpace(anchor.ProviderKey)
	anchor.PaymentAmount = strings.TrimSpace(anchor.PaymentAmount)
	return anchor
}

func CanonicalMixedVerificationHook(hook MixedVerificationHook) MixedVerificationHook {
	hook.TargetServiceID = strings.TrimSpace(hook.TargetServiceID)
	hook.MethodID = strings.TrimSpace(hook.MethodID)
	hook.ProofCommitment = strings.ToLower(strings.TrimSpace(hook.ProofCommitment))
	hook.ExpectedResultHash = strings.ToLower(strings.TrimSpace(hook.ExpectedResultHash))
	return hook
}

func (hook MixedVerificationHook) Validate() error {
	if !IsMixedVerificationHookType(hook.HookType) {
		return fmt.Errorf("unknown aetracore mixed verification hook type %q", hook.HookType)
	}
	if hook.TargetServiceID != "" {
		if err := validatePolicyID("aetracore mixed verification target service", hook.TargetServiceID); err != nil {
			return err
		}
	}
	if hook.MethodID != "" {
		if err := validatePolicyID("aetracore mixed verification method", hook.MethodID); err != nil {
			return err
		}
	}
	if hook.ProofCommitment != "" {
		if err := ValidateHash("aetracore mixed verification proof commitment", hook.ProofCommitment); err != nil {
			return err
		}
	}
	if hook.ExpectedResultHash != "" {
		if err := ValidateHash("aetracore mixed verification expected result", hook.ExpectedResultHash); err != nil {
			return err
		}
	}
	if hook.HookType == MixedHookFallbackExecution && !hook.Deterministic {
		return errors.New("aetracore mixed fallback verification hook must be deterministic")
	}
	if hook.HookType != MixedHookFallbackExecution && hook.ProofMeterGas == 0 {
		return errors.New("aetracore mixed proof or recompute hook must be metered")
	}
	if err := ValidateHash("aetracore mixed verification hook hash", hook.HookHash); err != nil {
		return err
	}
	if expected := ComputeMixedVerificationHookHash(hook); hook.HookHash != expected {
		return fmt.Errorf("aetracore mixed verification hook hash mismatch: expected %s", expected)
	}
	return nil
}

func (state MixedServiceState) AnchorByID(anchorID string) (MixedResultAnchor, bool) {
	for _, anchor := range state.Anchors {
		if anchor.AnchorID == anchorID {
			return anchor, true
		}
	}
	return MixedResultAnchor{}, false
}

func (state MixedServiceState) DisputeByID(disputeID string) (MixedDispute, bool) {
	for _, dispute := range state.Disputes {
		if dispute.DisputeID == disputeID {
			return dispute, true
		}
	}
	return MixedDispute{}, false
}

func (state MixedServiceState) HasOpenDispute(anchorID string) bool {
	for _, dispute := range state.Disputes {
		if dispute.AnchorID == anchorID && dispute.Status == MixedDisputeOpen {
			return true
		}
	}
	return false
}

func (state MixedServiceState) HasSettlement(anchorID string) bool {
	for _, settlement := range state.Settlements {
		if settlement.AnchorID == anchorID {
			return true
		}
	}
	return false
}

func (state MixedServiceState) clone() MixedServiceState {
	state.Anchors = append([]MixedResultAnchor(nil), state.Anchors...)
	state.Disputes = append([]MixedDispute(nil), state.Disputes...)
	state.Settlements = append([]MixedSettlement(nil), state.Settlements...)
	return state
}

func ComputeMixedResultAnchorID(anchor MixedResultAnchor) string {
	return hashParts("aetra-aek-mixed-result-anchor-id-v1", anchor.ServiceID, anchor.CallID, anchor.RequestCommitment, anchor.ProviderKey)
}

func ComputeMixedResultAnchorHash(anchor MixedResultAnchor) string {
	return hashParts(
		"aetra-aek-mixed-result-anchor-v1",
		anchor.AnchorID,
		anchor.ServiceID,
		anchor.CallID,
		anchor.RequestCommitment,
		anchor.ResultCommitment,
		anchor.ReceiptCommitment,
		anchor.ProviderKey,
		fmt.Sprint(anchor.Height),
		fmt.Sprint(anchor.ChallengeEndHeight),
		anchor.PaymentAmount,
	)
}

func ComputeMixedDisputeID(dispute MixedDispute) string {
	return hashParts("aetra-aek-mixed-dispute-id-v1", dispute.AnchorID, dispute.Challenger, dispute.ChallengeCommitment)
}

func ComputeMixedDisputeHash(dispute MixedDispute) string {
	return hashParts(
		"aetra-aek-mixed-dispute-v1",
		dispute.DisputeID,
		dispute.AnchorID,
		dispute.ServiceID,
		dispute.Challenger,
		dispute.ChallengeCommitment,
		fmt.Sprint(dispute.OpenedHeight),
		fmt.Sprint(dispute.ResolveByHeight),
		string(dispute.Status),
		dispute.VerificationHook.HookHash,
	)
}

func ComputeMixedVerificationHookHash(hook MixedVerificationHook) string {
	return hashParts(
		"aetra-aek-mixed-verification-hook-v1",
		string(hook.HookType),
		hook.TargetServiceID,
		hook.MethodID,
		hook.ProofCommitment,
		fmt.Sprint(hook.Deterministic),
		fmt.Sprint(hook.ProofMeterGas),
		hook.ExpectedResultHash,
	)
}

func ComputeMixedDisputeResolutionHash(resolution MixedDisputeResolution) string {
	return hashParts(
		"aetra-aek-mixed-dispute-resolution-v1",
		resolution.DisputeID,
		resolution.Resolver,
		fmt.Sprint(resolution.ResolvedHeight),
		fmt.Sprint(resolution.ProofAccepted),
		fmt.Sprint(resolution.RecomputeMatches),
		fmt.Sprint(resolution.FallbackExecuted),
	)
}

func ComputeMixedSettlementID(settlement MixedSettlement) string {
	return hashParts("aetra-aek-mixed-settlement-id-v1", settlement.AnchorID, settlement.DisputeID, string(settlement.Status))
}

func ComputeMixedSettlementHash(settlement MixedSettlement) string {
	return hashParts(
		"aetra-aek-mixed-settlement-v1",
		settlement.SettlementID,
		settlement.AnchorID,
		settlement.DisputeID,
		settlement.ServiceID,
		settlement.ProviderKey,
		string(settlement.Status),
		fmt.Sprint(settlement.SettledHeight),
		settlement.PaymentDenom,
		settlement.PaymentAmount,
		string(settlement.PaymentStatus),
		settlement.PenaltyDenom,
		settlement.PenaltyAmount,
		settlement.PenaltyRecipient,
	)
}

func ComputeMixedServiceStateHash(state MixedServiceState) string {
	parts := []string{
		"aetra-aek-mixed-service-state-v1",
		state.ServiceID,
		state.ProviderKey,
		state.PaymentDenom,
		state.EscrowID,
		state.EscrowAmount,
		state.CollateralDenom,
		state.CollateralAmount,
		state.RequiredCollateralAmount,
		string(state.FaultClass),
		fmt.Sprint(state.ChallengeWindow),
		state.FallbackServiceID,
		state.FallbackMethodID,
		fmt.Sprint(state.FallbackDeterministic),
		fmt.Sprint(len(state.Anchors)),
	}
	for _, anchor := range state.Anchors {
		parts = append(parts, anchor.AnchorHash)
	}
	parts = append(parts, fmt.Sprint(len(state.Disputes)))
	for _, dispute := range state.Disputes {
		parts = append(parts, dispute.DisputeHash)
	}
	parts = append(parts, fmt.Sprint(len(state.Settlements)))
	for _, settlement := range state.Settlements {
		parts = append(parts, settlement.SettlementHash)
	}
	return hashParts(parts...)
}

func IsMixedFaultClass(faultClass MixedServiceFaultClass) bool {
	switch faultClass {
	case MixedFaultLow, MixedFaultMedium, MixedFaultHigh, MixedFaultCritical:
		return true
	default:
		return false
	}
}

func IsMixedDisputeStatus(status MixedDisputeStatus) bool {
	switch status {
	case MixedDisputeOpen, MixedDisputeRejected, MixedDisputeProven, MixedDisputeFallback, MixedDisputeExpired:
		return true
	default:
		return false
	}
}

func IsMixedSettlementStatus(status MixedSettlementStatus) bool {
	switch status {
	case MixedSettlementReleased, MixedSettlementRefunded, MixedSettlementPenalized, MixedSettlementFallback:
		return true
	default:
		return false
	}
}

func IsMixedVerificationHookType(hookType MixedVerificationHookType) bool {
	switch hookType {
	case MixedHookProofVerification, MixedHookRecompute, MixedHookFallbackExecution:
		return true
	default:
		return false
	}
}

func validateMixedAnchors(anchors []MixedResultAnchor) error {
	var previous string
	seen := make(map[string]struct{}, len(anchors))
	for _, anchor := range anchors {
		if err := anchor.Validate(); err != nil {
			return err
		}
		if _, found := seen[anchor.AnchorID]; found {
			return fmt.Errorf("duplicate aetracore mixed result anchor %s", anchor.AnchorID)
		}
		seen[anchor.AnchorID] = struct{}{}
		if previous != "" && previous >= anchor.AnchorID {
			return errors.New("aetracore mixed result anchors must be sorted canonically")
		}
		previous = anchor.AnchorID
	}
	return nil
}

func validateMixedDisputes(disputes []MixedDispute) error {
	var previous string
	seen := make(map[string]struct{}, len(disputes))
	for _, dispute := range disputes {
		if err := dispute.Validate(); err != nil {
			return err
		}
		if _, found := seen[dispute.DisputeID]; found {
			return fmt.Errorf("duplicate aetracore mixed dispute %s", dispute.DisputeID)
		}
		seen[dispute.DisputeID] = struct{}{}
		if previous != "" && previous >= dispute.DisputeID {
			return errors.New("aetracore mixed disputes must be sorted canonically")
		}
		previous = dispute.DisputeID
	}
	return nil
}

func validateMixedSettlements(settlements []MixedSettlement) error {
	var previous string
	seen := make(map[string]struct{}, len(settlements))
	for _, settlement := range settlements {
		if err := settlement.Validate(); err != nil {
			return err
		}
		if _, found := seen[settlement.SettlementID]; found {
			return fmt.Errorf("duplicate aetracore mixed settlement %s", settlement.SettlementID)
		}
		seen[settlement.SettlementID] = struct{}{}
		if previous != "" && previous >= settlement.SettlementID {
			return errors.New("aetracore mixed settlements must be sorted canonically")
		}
		previous = settlement.SettlementID
	}
	return nil
}

func (settlement MixedSettlement) Validate() error {
	if err := ValidateHash("aetracore mixed settlement id", settlement.SettlementID); err != nil {
		return err
	}
	if err := ValidateHash("aetracore mixed settlement anchor id", settlement.AnchorID); err != nil {
		return err
	}
	if err := validateOptionalHash("aetracore mixed settlement dispute id", settlement.DisputeID); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore mixed settlement service id", settlement.ServiceID); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore mixed settlement provider key", settlement.ProviderKey); err != nil {
		return err
	}
	if !IsMixedSettlementStatus(settlement.Status) {
		return fmt.Errorf("unknown aetracore mixed settlement status %q", settlement.Status)
	}
	if settlement.SettledHeight == 0 {
		return errors.New("aetracore mixed settlement height must be positive")
	}
	if err := validatePolicyID("aetracore mixed settlement payment denom", settlement.PaymentDenom); err != nil {
		return err
	}
	if err := validateAmountString("aetracore mixed settlement payment amount", settlement.PaymentAmount); err != nil {
		return err
	}
	if !IsServicePaymentStatus(settlement.PaymentStatus) {
		return fmt.Errorf("unknown aetracore mixed settlement payment status %q", settlement.PaymentStatus)
	}
	if settlement.PenaltyDenom != "" {
		if err := validatePolicyID("aetracore mixed settlement penalty denom", settlement.PenaltyDenom); err != nil {
			return err
		}
	}
	if settlement.PenaltyAmount != "" {
		if err := validateAmountString("aetracore mixed settlement penalty amount", settlement.PenaltyAmount); err != nil {
			return err
		}
	}
	if settlement.PenaltyRecipient != "" {
		if err := validatePolicyID("aetracore mixed settlement penalty recipient", settlement.PenaltyRecipient); err != nil {
			return err
		}
	}
	if err := ValidateHash("aetracore mixed settlement hash", settlement.SettlementHash); err != nil {
		return err
	}
	if expected := ComputeMixedSettlementHash(settlement); settlement.SettlementHash != expected {
		return fmt.Errorf("aetracore mixed settlement hash mismatch: expected %s", expected)
	}
	return nil
}

func sortMixedAnchors(anchors []MixedResultAnchor) {
	sort.SliceStable(anchors, func(i, j int) bool { return anchors[i].AnchorID < anchors[j].AnchorID })
}

func sortMixedDisputes(disputes []MixedDispute) {
	sort.SliceStable(disputes, func(i, j int) bool { return disputes[i].DisputeID < disputes[j].DisputeID })
}

func sortMixedSettlements(settlements []MixedSettlement) {
	sort.SliceStable(settlements, func(i, j int) bool { return settlements[i].SettlementID < settlements[j].SettlementID })
}

func mixedDisputeStatusForResolution(resolution MixedDisputeResolution) MixedDisputeStatus {
	if resolution.FallbackExecuted {
		return MixedDisputeFallback
	}
	if resolution.ProofAccepted || !resolution.RecomputeMatches {
		return MixedDisputeProven
	}
	return MixedDisputeRejected
}

func settlementForResolvedDispute(state MixedServiceState, anchor MixedResultAnchor, dispute MixedDispute, resolution MixedDisputeResolution) MixedSettlement {
	status := MixedSettlementReleased
	paymentStatus := ServicePaymentStatusSettled
	penaltyDenom := ""
	penaltyAmount := ""
	penaltyRecipient := ""
	if resolution.FallbackExecuted {
		status = MixedSettlementFallback
		paymentStatus = ServicePaymentStatusEscrowed
	} else if resolution.ProofAccepted || !resolution.RecomputeMatches {
		status = MixedSettlementPenalized
		paymentStatus = ServicePaymentStatusRefunded
		penaltyDenom = state.CollateralDenom
		penaltyAmount = state.RequiredCollateralAmount
		penaltyRecipient = dispute.Challenger
	}
	settlement := MixedSettlement{
		AnchorID:		anchor.AnchorID,
		DisputeID:		dispute.DisputeID,
		ServiceID:		state.ServiceID,
		ProviderKey:		state.ProviderKey,
		Status:			status,
		SettledHeight:		resolution.ResolvedHeight,
		PaymentDenom:		state.PaymentDenom,
		PaymentAmount:		anchor.PaymentAmount,
		PaymentStatus:		paymentStatus,
		PenaltyDenom:		penaltyDenom,
		PenaltyAmount:		penaltyAmount,
		PenaltyRecipient:	penaltyRecipient,
	}
	settlement.SettlementID = ComputeMixedSettlementID(settlement)
	settlement.SettlementHash = ComputeMixedSettlementHash(settlement)
	return settlement
}

func mixedAmountAtLeast(fieldName, amount, minimum string) error {
	cmp, err := mixedCompareAmount(amount, minimum)
	if err != nil {
		return fmt.Errorf("%s: %w", fieldName, err)
	}
	if cmp < 0 {
		return fmt.Errorf("%s must cover required amount %s", fieldName, minimum)
	}
	return nil
}

func mixedAmountAtMost(fieldName, amount, maximum string) error {
	cmp, err := mixedCompareAmount(amount, maximum)
	if err != nil {
		return fmt.Errorf("%s: %w", fieldName, err)
	}
	if cmp > 0 {
		return fmt.Errorf("%s must not exceed %s", fieldName, maximum)
	}
	return nil
}

func mixedCompareAmount(left, right string) (int, error) {
	if err := validateAmountString("left amount", left); err != nil {
		return 0, err
	}
	if err := validateAmountString("right amount", right); err != nil {
		return 0, err
	}
	leftInt, ok := new(big.Int).SetString(left, 10)
	if !ok {
		return 0, errors.New("invalid left amount")
	}
	rightInt, ok := new(big.Int).SetString(right, 10)
	if !ok {
		return 0, errors.New("invalid right amount")
	}
	return leftInt.Cmp(rightInt), nil
}
