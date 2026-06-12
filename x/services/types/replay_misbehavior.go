package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

type ProviderMisbehaviorFaultClass string
type ProviderPenaltySource string

const (
	ProviderFaultInvalidResult		ProviderMisbehaviorFaultClass	= "invalid_result"
	ProviderFaultMissingResult		ProviderMisbehaviorFaultClass	= "missing_result"
	ProviderFaultLateResult			ProviderMisbehaviorFaultClass	= "late_result"
	ProviderFaultDoubleResponse		ProviderMisbehaviorFaultClass	= "double_response"
	ProviderFaultWrongInterfaceVersion	ProviderMisbehaviorFaultClass	= "wrong_interface_version"
	ProviderFaultInvalidProof		ProviderMisbehaviorFaultClass	= "invalid_proof"
	ProviderFaultAvailabilityViolation	ProviderMisbehaviorFaultClass	= "availability_violation"

	ProviderPenaltyCollateral	ProviderPenaltySource	= "provider_collateral"
	ProviderPenaltyServiceStake	ProviderPenaltySource	= "service_stake"
	ProviderPenaltyEscrowedPayment	ProviderPenaltySource	= "escrowed_payment"
	ProviderPenaltyReputationScore	ProviderPenaltySource	= "reputation_score"
)

type ServiceReplayProtectionProof struct {
	CallID			string
	ServiceID		string
	MethodID		string
	Caller			string
	Nonce			uint64
	IdempotencyKey		string
	PayloadHash		string
	DeadlineHeight		uint64
	ReplayEntryHash		string
	ReceiptTombstone	bool
	TombstoneHash		string
	ReplayIndexHash		string
	ControlHash		string
}

type ProviderMisbehaviorReport struct {
	ServiceID		string
	ProviderID		string
	CallID			string
	FaultClass		ProviderMisbehaviorFaultClass
	EvidenceHash		string
	ExpectedInterfaceHash	string
	ObservedInterfaceHash	string
	ProofHash		string
	ObservedHeight		uint64
	DeadlineHeight		uint64
	PenaltySources		[]ProviderPenaltySource
	CollateralSlashAmount	string
	ServiceStakeSlashAmount	string
	EscrowForfeitAmount	string
	ReputationDelta		int64
	ReportHash		string
}

type ServiceFaultProof struct {
	ServiceID		string
	ProviderID		string
	CallID			string
	FaultClass		ProviderMisbehaviorFaultClass
	ReportHash		string
	EvidenceHash		string
	ProofHash		string
	ExpectedInterfaceHash	string
	ObservedInterfaceHash	string
	ReceiptHash		string
	TombstoneHash		string
	SubmittedHeight		uint64
	FaultProofHash		string
}

type ProviderPenaltyRouteEntry struct {
	Source		ProviderPenaltySource
	Denom		string
	Amount		string
	ReputationDelta	int64
	Recipient	string
	RouteEntryHash	string
}

type ProviderPenaltyRoute struct {
	ServiceID	string
	ProviderID	string
	CallID		string
	FaultClass	ProviderMisbehaviorFaultClass
	ReportHash	string
	Entries		[]ProviderPenaltyRouteEntry
	RouteHash	string
}

type ServiceChallengeFlow struct {
	ChallengeID		string
	ServiceID		string
	ProviderID		string
	CallID			string
	FaultClass		ProviderMisbehaviorFaultClass
	DisputeMessageHash	string
	FaultProofHash		string
	PenaltyRouteHash	string
	OpenedHeight		uint64
	ChallengeEndHeight	uint64
	FlowHash		string
}

type ServiceReceiptFreshnessProof struct {
	CallID			string
	ReceiptHash		string
	TombstoneHash		string
	ReceiptHeight		uint64
	RetainUntilHeight	uint64
	CurrentHeight		uint64
	Stale			bool
	FreshnessHash		string
}

type ServiceSecurityImplementationBundle struct {
	ServiceID			string
	TrustModelLabel			ServiceTrustModelLabel
	ExecutionFailureBehaviorLabel	ServiceFailureBehaviorLabel
	CollateralMessageHash		string
	ChallengeFlowHash		string
	FaultProofHash			string
	PenaltyRouteHash		string
	ReplayControlHash		string
	ReceiptFreshnessHash		string
	ImplementationHash		string
}

func NewServiceReplayProtectionProof(ctx coretypes.ServiceConsensusContext, descriptor ServiceDescriptor, index ServiceCallReplayIndex, call UnifiedServiceCall, requireTombstone bool) (ServiceReplayProtectionProof, error) {
	if err := ctx.Validate(); err != nil {
		return ServiceReplayProtectionProof{}, err
	}
	descriptor = coretypes.CanonicalServiceDescriptor(descriptor)
	if err := ValidateUnifiedServiceCallForDescriptor(ctx, descriptor, call); err != nil {
		return ServiceReplayProtectionProof{}, err
	}
	if err := index.Validate(); err != nil {
		return ServiceReplayProtectionProof{}, err
	}
	entry, found := index.EntryByServiceCallerNonce(call.TargetService, call.Caller, call.Nonce)
	if !found {
		return ServiceReplayProtectionProof{}, errors.New("services replay protection missing caller nonce entry")
	}
	if entry.CallID != call.CallID || entry.IdempotencyKey != call.IdempotencyKey || entry.PayloadHash != call.PayloadHash || entry.DeadlineHeight != call.DeadlineHeight {
		return ServiceReplayProtectionProof{}, errors.New("services replay protection entry does not match call controls")
	}
	tombstone, tombstoneFound := index.TombstoneByCallID(call.CallID)
	if requireTombstone && !tombstoneFound {
		return ServiceReplayProtectionProof{}, errors.New("services replay protection requires receipt tombstone")
	}
	proof := ServiceReplayProtectionProof{
		CallID:			call.CallID,
		ServiceID:		call.TargetService,
		MethodID:		call.MethodID,
		Caller:			call.Caller,
		Nonce:			call.Nonce,
		IdempotencyKey:		call.IdempotencyKey,
		PayloadHash:		call.PayloadHash,
		DeadlineHeight:		call.DeadlineHeight,
		ReplayEntryHash:	entry.EntryHash,
		ReceiptTombstone:	tombstoneFound,
		ReplayIndexHash:	index.IndexHash,
	}
	if tombstoneFound {
		proof.TombstoneHash = tombstone.TombstoneHash
	}
	proof.ControlHash = ComputeServiceReplayProtectionProofHash(proof)
	return proof, proof.Validate()
}

func NewProviderMisbehaviorReport(report ProviderMisbehaviorReport) (ProviderMisbehaviorReport, error) {
	if report.ReportHash != "" {
		return ProviderMisbehaviorReport{}, errors.New("services provider misbehavior report hash must be empty before construction")
	}
	report = canonicalProviderMisbehaviorReport(report)
	if err := report.ValidateFormat(); err != nil {
		return ProviderMisbehaviorReport{}, err
	}
	report.ReportHash = ComputeProviderMisbehaviorReportHash(report)
	return report, report.Validate()
}

func NewServiceFaultProof(report ProviderMisbehaviorReport, receipt ServiceReceipt, tombstone ServiceReceiptTombstone, submittedHeight uint64) (ServiceFaultProof, error) {
	report = canonicalProviderMisbehaviorReport(report)
	if err := report.Validate(); err != nil {
		return ServiceFaultProof{}, err
	}
	if err := receipt.Validate(); err != nil {
		return ServiceFaultProof{}, err
	}
	if err := tombstone.Validate(); err != nil {
		return ServiceFaultProof{}, err
	}
	if report.CallID != "" && receipt.CallID != report.CallID {
		return ServiceFaultProof{}, errors.New("services fault proof receipt call mismatch")
	}
	if tombstone.CallID != receipt.CallID || tombstone.ReceiptHash != receipt.ReceiptHash {
		return ServiceFaultProof{}, errors.New("services fault proof tombstone mismatch")
	}
	if submittedHeight == 0 || submittedHeight < report.ObservedHeight {
		return ServiceFaultProof{}, errors.New("services fault proof submitted height is invalid")
	}
	proof := ServiceFaultProof{
		ServiceID:		report.ServiceID,
		ProviderID:		report.ProviderID,
		CallID:			receipt.CallID,
		FaultClass:		report.FaultClass,
		ReportHash:		report.ReportHash,
		EvidenceHash:		report.EvidenceHash,
		ProofHash:		report.ProofHash,
		ExpectedInterfaceHash:	report.ExpectedInterfaceHash,
		ObservedInterfaceHash:	report.ObservedInterfaceHash,
		ReceiptHash:		receipt.ReceiptHash,
		TombstoneHash:		tombstone.TombstoneHash,
		SubmittedHeight:	submittedHeight,
	}
	proof.FaultProofHash = ComputeServiceFaultProofHash(proof)
	return proof, proof.Validate()
}

func NewProviderPenaltyRoute(report ProviderMisbehaviorReport, denom, recipient string) (ProviderPenaltyRoute, error) {
	report = canonicalProviderMisbehaviorReport(report)
	if err := report.Validate(); err != nil {
		return ProviderPenaltyRoute{}, err
	}
	if err := validateInterfaceToken("services penalty route denom", denom); err != nil {
		return ProviderPenaltyRoute{}, err
	}
	if err := validateInterfaceToken("services penalty route recipient", recipient); err != nil {
		return ProviderPenaltyRoute{}, err
	}
	route := ProviderPenaltyRoute{
		ServiceID:	report.ServiceID,
		ProviderID:	report.ProviderID,
		CallID:		report.CallID,
		FaultClass:	report.FaultClass,
		ReportHash:	report.ReportHash,
	}
	for _, source := range report.PenaltySources {
		entry := ProviderPenaltyRouteEntry{Source: source, Denom: denom, Recipient: recipient}
		switch source {
		case ProviderPenaltyCollateral:
			entry.Amount = report.CollateralSlashAmount
		case ProviderPenaltyServiceStake:
			entry.Amount = report.ServiceStakeSlashAmount
		case ProviderPenaltyEscrowedPayment:
			entry.Amount = report.EscrowForfeitAmount
		case ProviderPenaltyReputationScore:
			entry.ReputationDelta = report.ReputationDelta
		}
		entry.RouteEntryHash = ComputeProviderPenaltyRouteEntryHash(entry)
		route.Entries = append(route.Entries, entry)
	}
	sortProviderPenaltyRouteEntries(route.Entries)
	route.RouteHash = ComputeProviderPenaltyRouteHash(route)
	return route, route.Validate()
}

func NewServiceChallengeFlow(descriptor ServiceDescriptor, msg coretypes.MsgSubmitServiceDispute, proof ServiceFaultProof, route ProviderPenaltyRoute) (ServiceChallengeFlow, error) {
	descriptor = coretypes.CanonicalServiceDescriptor(descriptor)
	if err := descriptor.Validate(); err != nil {
		return ServiceChallengeFlow{}, err
	}
	if err := msg.ValidateBasic(); err != nil {
		return ServiceChallengeFlow{}, err
	}
	if err := proof.Validate(); err != nil {
		return ServiceChallengeFlow{}, err
	}
	if err := route.Validate(); err != nil {
		return ServiceChallengeFlow{}, err
	}
	if msg.ServiceID != descriptor.ServiceID || proof.ServiceID != descriptor.ServiceID || route.ServiceID != descriptor.ServiceID {
		return ServiceChallengeFlow{}, errors.New("services challenge flow service mismatch")
	}
	if msg.CallID != proof.CallID || route.CallID != proof.CallID {
		return ServiceChallengeFlow{}, errors.New("services challenge flow call mismatch")
	}
	if msg.ProviderID != proof.ProviderID || route.ProviderID != proof.ProviderID {
		return ServiceChallengeFlow{}, errors.New("services challenge flow provider mismatch")
	}
	challengeWindow := maxUint64(descriptor.Verification.ChallengeWindow, descriptor.Execution.ChallengeWindow)
	if challengeWindow == 0 {
		return ServiceChallengeFlow{}, errors.New("services challenge flow requires challenge window")
	}
	flow := ServiceChallengeFlow{
		ServiceID:		descriptor.ServiceID,
		ProviderID:		proof.ProviderID,
		CallID:			proof.CallID,
		FaultClass:		proof.FaultClass,
		DisputeMessageHash:	msg.MessageHash,
		FaultProofHash:		proof.FaultProofHash,
		PenaltyRouteHash:	route.RouteHash,
		OpenedHeight:		msg.OpenedHeight,
		ChallengeEndHeight:	msg.OpenedHeight + challengeWindow,
	}
	flow.ChallengeID = ComputeServiceChallengeID(flow)
	flow.FlowHash = ComputeServiceChallengeFlowHash(flow)
	return flow, flow.Validate()
}

func NewServiceReceiptFreshnessProof(tombstone ServiceReceiptTombstone, currentHeight uint64, requireFresh bool) (ServiceReceiptFreshnessProof, error) {
	if err := tombstone.Validate(); err != nil {
		return ServiceReceiptFreshnessProof{}, err
	}
	if currentHeight == 0 {
		return ServiceReceiptFreshnessProof{}, errors.New("services receipt freshness current height must be positive")
	}
	stale := currentHeight > tombstone.RetainUntilHeight
	if requireFresh && stale {
		return ServiceReceiptFreshnessProof{}, errors.New("services receipt freshness proof is stale")
	}
	proof := ServiceReceiptFreshnessProof{
		CallID:			tombstone.CallID,
		ReceiptHash:		tombstone.ReceiptHash,
		TombstoneHash:		tombstone.TombstoneHash,
		ReceiptHeight:		tombstone.ReceiptHeight,
		RetainUntilHeight:	tombstone.RetainUntilHeight,
		CurrentHeight:		currentHeight,
		Stale:			stale,
	}
	proof.FreshnessHash = ComputeServiceReceiptFreshnessProofHash(proof)
	return proof, proof.Validate()
}

func NewServiceSecurityImplementationBundle(descriptor ServiceDescriptor, collateralMsg coretypes.MsgStakeProviderCollateral, flow ServiceChallengeFlow, faultProof ServiceFaultProof, route ProviderPenaltyRoute, replayProof ServiceReplayProtectionProof, freshness ServiceReceiptFreshnessProof) (ServiceSecurityImplementationBundle, error) {
	descriptor = coretypes.CanonicalServiceDescriptor(descriptor)
	if err := descriptor.Validate(); err != nil {
		return ServiceSecurityImplementationBundle{}, err
	}
	if err := collateralMsg.ValidateBasic(); err != nil {
		return ServiceSecurityImplementationBundle{}, err
	}
	if err := flow.Validate(); err != nil {
		return ServiceSecurityImplementationBundle{}, err
	}
	if err := faultProof.Validate(); err != nil {
		return ServiceSecurityImplementationBundle{}, err
	}
	if err := route.Validate(); err != nil {
		return ServiceSecurityImplementationBundle{}, err
	}
	if err := replayProof.Validate(); err != nil {
		return ServiceSecurityImplementationBundle{}, err
	}
	if err := freshness.Validate(); err != nil {
		return ServiceSecurityImplementationBundle{}, err
	}
	if collateralMsg.ServiceID != descriptor.ServiceID || flow.ServiceID != descriptor.ServiceID || faultProof.ServiceID != descriptor.ServiceID || route.ServiceID != descriptor.ServiceID || replayProof.ServiceID != descriptor.ServiceID {
		return ServiceSecurityImplementationBundle{}, errors.New("services security implementation service mismatch")
	}
	if collateralMsg.ProviderID != flow.ProviderID || collateralMsg.ProviderID != faultProof.ProviderID || collateralMsg.ProviderID != route.ProviderID {
		return ServiceSecurityImplementationBundle{}, errors.New("services security implementation provider mismatch")
	}
	if flow.FaultProofHash != faultProof.FaultProofHash || flow.PenaltyRouteHash != route.RouteHash {
		return ServiceSecurityImplementationBundle{}, errors.New("services security implementation flow linkage mismatch")
	}
	trustLabel, err := ServiceTrustModelSpecLabel(descriptor.Verification.TrustModel)
	if err != nil {
		return ServiceSecurityImplementationBundle{}, err
	}
	failureLabel, err := ServiceFailureBehaviorSpecLabel(descriptor.Execution.FailureBehavior)
	if err != nil {
		return ServiceSecurityImplementationBundle{}, err
	}
	bundle := ServiceSecurityImplementationBundle{
		ServiceID:			descriptor.ServiceID,
		TrustModelLabel:		trustLabel,
		ExecutionFailureBehaviorLabel:	failureLabel,
		CollateralMessageHash:		collateralMsg.MessageHash,
		ChallengeFlowHash:		flow.FlowHash,
		FaultProofHash:			faultProof.FaultProofHash,
		PenaltyRouteHash:		route.RouteHash,
		ReplayControlHash:		replayProof.ControlHash,
		ReceiptFreshnessHash:		freshness.FreshnessHash,
	}
	bundle.ImplementationHash = ComputeServiceSecurityImplementationBundleHash(bundle)
	return bundle, bundle.Validate()
}

func (proof ServiceReplayProtectionProof) Validate() error {
	if err := coretypes.ValidateHash("services replay protection call id", proof.CallID); err != nil {
		return err
	}
	if err := validateInterfaceToken("services replay protection service id", proof.ServiceID); err != nil {
		return err
	}
	if err := validateInterfaceToken("services replay protection method id", proof.MethodID); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("services replay protection caller", proof.Caller); err != nil {
		return err
	}
	if proof.Nonce == 0 {
		return errors.New("services replay protection nonce must be positive")
	}
	if proof.IdempotencyKey == "" {
		return errors.New("services replay protection requires idempotency key")
	}
	if err := validateInterfaceToken("services replay protection idempotency key", proof.IdempotencyKey); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services replay protection payload hash", proof.PayloadHash); err != nil {
		return err
	}
	if proof.DeadlineHeight == 0 {
		return errors.New("services replay protection deadline height must be positive")
	}
	if err := coretypes.ValidateHash("services replay protection entry hash", proof.ReplayEntryHash); err != nil {
		return err
	}
	if proof.ReceiptTombstone {
		if err := coretypes.ValidateHash("services replay protection tombstone hash", proof.TombstoneHash); err != nil {
			return err
		}
	} else if proof.TombstoneHash != "" {
		return errors.New("services replay protection tombstone hash set without tombstone flag")
	}
	if err := coretypes.ValidateHash("services replay protection index hash", proof.ReplayIndexHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services replay protection control hash", proof.ControlHash); err != nil {
		return err
	}
	if expected := ComputeServiceReplayProtectionProofHash(proof); proof.ControlHash != expected {
		return fmt.Errorf("services replay protection control hash mismatch: expected %s", expected)
	}
	return nil
}

func (proof ServiceFaultProof) Validate() error {
	if err := validateInterfaceToken("services fault proof service id", proof.ServiceID); err != nil {
		return err
	}
	if err := validateInterfaceToken("services fault proof provider id", proof.ProviderID); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services fault proof call id", proof.CallID); err != nil {
		return err
	}
	if !IsProviderMisbehaviorFaultClass(proof.FaultClass) {
		return fmt.Errorf("services fault proof unknown class %q", proof.FaultClass)
	}
	if err := coretypes.ValidateHash("services fault proof report hash", proof.ReportHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services fault proof evidence hash", proof.EvidenceHash); err != nil {
		return err
	}
	if proof.FaultClass == ProviderFaultInvalidProof && proof.ProofHash == "" {
		return errors.New("services fault proof invalid-proof class requires proof hash")
	}
	if proof.ProofHash != "" {
		if err := coretypes.ValidateHash("services fault proof proof hash", proof.ProofHash); err != nil {
			return err
		}
	}
	if proof.ExpectedInterfaceHash != "" {
		if err := coretypes.ValidateHash("services fault proof expected interface hash", proof.ExpectedInterfaceHash); err != nil {
			return err
		}
	}
	if proof.ObservedInterfaceHash != "" {
		if err := coretypes.ValidateHash("services fault proof observed interface hash", proof.ObservedInterfaceHash); err != nil {
			return err
		}
	}
	if err := coretypes.ValidateHash("services fault proof receipt hash", proof.ReceiptHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services fault proof tombstone hash", proof.TombstoneHash); err != nil {
		return err
	}
	if proof.SubmittedHeight == 0 {
		return errors.New("services fault proof submitted height must be positive")
	}
	if err := coretypes.ValidateHash("services fault proof hash", proof.FaultProofHash); err != nil {
		return err
	}
	if expected := ComputeServiceFaultProofHash(proof); proof.FaultProofHash != expected {
		return fmt.Errorf("services fault proof hash mismatch: expected %s", expected)
	}
	return nil
}

func (entry ProviderPenaltyRouteEntry) Validate() error {
	if !IsProviderPenaltySource(entry.Source) {
		return fmt.Errorf("services penalty route unknown source %q", entry.Source)
	}
	if err := validateInterfaceToken("services penalty route denom", entry.Denom); err != nil {
		return err
	}
	if err := validateInterfaceToken("services penalty route recipient", entry.Recipient); err != nil {
		return err
	}
	if entry.Source == ProviderPenaltyReputationScore {
		if entry.ReputationDelta >= 0 {
			return errors.New("services penalty route reputation delta must be negative")
		}
	} else if err := validatePositivePaymentAmount("services penalty route amount", entry.Amount); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services penalty route entry hash", entry.RouteEntryHash); err != nil {
		return err
	}
	if expected := ComputeProviderPenaltyRouteEntryHash(entry); entry.RouteEntryHash != expected {
		return fmt.Errorf("services penalty route entry hash mismatch: expected %s", expected)
	}
	return nil
}

func (route ProviderPenaltyRoute) Validate() error {
	if err := validateInterfaceToken("services penalty route service id", route.ServiceID); err != nil {
		return err
	}
	if err := validateInterfaceToken("services penalty route provider id", route.ProviderID); err != nil {
		return err
	}
	if route.CallID != "" {
		if err := coretypes.ValidateHash("services penalty route call id", route.CallID); err != nil {
			return err
		}
	}
	if !IsProviderMisbehaviorFaultClass(route.FaultClass) {
		return fmt.Errorf("services penalty route unknown fault class %q", route.FaultClass)
	}
	if err := coretypes.ValidateHash("services penalty route report hash", route.ReportHash); err != nil {
		return err
	}
	if err := validateProviderPenaltyRouteEntries(route.Entries); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services penalty route hash", route.RouteHash); err != nil {
		return err
	}
	if expected := ComputeProviderPenaltyRouteHash(route); route.RouteHash != expected {
		return fmt.Errorf("services penalty route hash mismatch: expected %s", expected)
	}
	return nil
}

func (flow ServiceChallengeFlow) Validate() error {
	if err := coretypes.ValidateHash("services challenge id", flow.ChallengeID); err != nil {
		return err
	}
	if err := validateInterfaceToken("services challenge service id", flow.ServiceID); err != nil {
		return err
	}
	if err := validateInterfaceToken("services challenge provider id", flow.ProviderID); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services challenge call id", flow.CallID); err != nil {
		return err
	}
	if !IsProviderMisbehaviorFaultClass(flow.FaultClass) {
		return fmt.Errorf("services challenge unknown fault class %q", flow.FaultClass)
	}
	if err := coretypes.ValidateHash("services challenge dispute message hash", flow.DisputeMessageHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services challenge fault proof hash", flow.FaultProofHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services challenge penalty route hash", flow.PenaltyRouteHash); err != nil {
		return err
	}
	if flow.OpenedHeight == 0 || flow.ChallengeEndHeight <= flow.OpenedHeight {
		return errors.New("services challenge heights are invalid")
	}
	if err := coretypes.ValidateHash("services challenge flow hash", flow.FlowHash); err != nil {
		return err
	}
	if expectedID := ComputeServiceChallengeID(flow); flow.ChallengeID != expectedID {
		return fmt.Errorf("services challenge id mismatch: expected %s", expectedID)
	}
	if expected := ComputeServiceChallengeFlowHash(flow); flow.FlowHash != expected {
		return fmt.Errorf("services challenge flow hash mismatch: expected %s", expected)
	}
	return nil
}

func (proof ServiceReceiptFreshnessProof) Validate() error {
	if err := coretypes.ValidateHash("services receipt freshness call id", proof.CallID); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services receipt freshness receipt hash", proof.ReceiptHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services receipt freshness tombstone hash", proof.TombstoneHash); err != nil {
		return err
	}
	if proof.ReceiptHeight == 0 || proof.RetainUntilHeight < proof.ReceiptHeight || proof.CurrentHeight == 0 {
		return errors.New("services receipt freshness heights are invalid")
	}
	if proof.Stale != (proof.CurrentHeight > proof.RetainUntilHeight) {
		return errors.New("services receipt freshness stale flag mismatch")
	}
	if err := coretypes.ValidateHash("services receipt freshness hash", proof.FreshnessHash); err != nil {
		return err
	}
	if expected := ComputeServiceReceiptFreshnessProofHash(proof); proof.FreshnessHash != expected {
		return fmt.Errorf("services receipt freshness hash mismatch: expected %s", expected)
	}
	return nil
}

func (bundle ServiceSecurityImplementationBundle) Validate() error {
	if err := validateInterfaceToken("services security implementation service id", bundle.ServiceID); err != nil {
		return err
	}
	if bundle.TrustModelLabel == "" {
		return errors.New("services security implementation trust label is required")
	}
	if bundle.ExecutionFailureBehaviorLabel == "" {
		return errors.New("services security implementation failure label is required")
	}
	for label, value := range map[string]string{
		"collateral message":	bundle.CollateralMessageHash,
		"challenge flow":	bundle.ChallengeFlowHash,
		"fault proof":		bundle.FaultProofHash,
		"penalty route":	bundle.PenaltyRouteHash,
		"replay control":	bundle.ReplayControlHash,
		"receipt freshness":	bundle.ReceiptFreshnessHash,
		"implementation":	bundle.ImplementationHash,
	} {
		if err := coretypes.ValidateHash("services security implementation "+label+" hash", value); err != nil {
			return err
		}
	}
	if expected := ComputeServiceSecurityImplementationBundleHash(bundle); bundle.ImplementationHash != expected {
		return fmt.Errorf("services security implementation hash mismatch: expected %s", expected)
	}
	return nil
}

func (report ProviderMisbehaviorReport) ValidateFormat() error {
	report = canonicalProviderMisbehaviorReport(report)
	if err := validateInterfaceToken("services provider fault service id", report.ServiceID); err != nil {
		return err
	}
	if err := validateInterfaceToken("services provider fault provider id", report.ProviderID); err != nil {
		return err
	}
	if report.CallID != "" {
		if err := coretypes.ValidateHash("services provider fault call id", report.CallID); err != nil {
			return err
		}
	}
	if !IsProviderMisbehaviorFaultClass(report.FaultClass) {
		return fmt.Errorf("services provider fault unknown class %q", report.FaultClass)
	}
	if err := coretypes.ValidateHash("services provider fault evidence hash", report.EvidenceHash); err != nil {
		return err
	}
	if report.ExpectedInterfaceHash != "" {
		if err := coretypes.ValidateHash("services provider fault expected interface hash", report.ExpectedInterfaceHash); err != nil {
			return err
		}
	}
	if report.ObservedInterfaceHash != "" {
		if err := coretypes.ValidateHash("services provider fault observed interface hash", report.ObservedInterfaceHash); err != nil {
			return err
		}
	}
	if report.ProofHash != "" {
		if err := coretypes.ValidateHash("services provider fault proof hash", report.ProofHash); err != nil {
			return err
		}
	}
	if report.ObservedHeight == 0 {
		return errors.New("services provider fault observed height must be positive")
	}
	if err := validateProviderPenaltySources(report.PenaltySources); err != nil {
		return err
	}
	if stringInPenaltySources(report.PenaltySources, ProviderPenaltyCollateral) {
		if err := validatePositivePaymentAmount("services provider fault collateral slash", report.CollateralSlashAmount); err != nil {
			return err
		}
	}
	if stringInPenaltySources(report.PenaltySources, ProviderPenaltyServiceStake) {
		if err := validatePositivePaymentAmount("services provider fault service stake slash", report.ServiceStakeSlashAmount); err != nil {
			return err
		}
	}
	if stringInPenaltySources(report.PenaltySources, ProviderPenaltyEscrowedPayment) {
		if err := validatePositivePaymentAmount("services provider fault escrow forfeit", report.EscrowForfeitAmount); err != nil {
			return err
		}
	}
	if stringInPenaltySources(report.PenaltySources, ProviderPenaltyReputationScore) && report.ReputationDelta >= 0 {
		return errors.New("services provider fault reputation penalty must be negative")
	}
	return validateProviderFaultClassRules(report)
}

func (report ProviderMisbehaviorReport) Validate() error {
	if err := report.ValidateFormat(); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services provider fault report hash", report.ReportHash); err != nil {
		return err
	}
	if expected := ComputeProviderMisbehaviorReportHash(report); report.ReportHash != expected {
		return fmt.Errorf("services provider fault report hash mismatch: expected %s", expected)
	}
	return nil
}

func IsProviderMisbehaviorFaultClass(faultClass ProviderMisbehaviorFaultClass) bool {
	switch faultClass {
	case ProviderFaultInvalidResult,
		ProviderFaultMissingResult,
		ProviderFaultLateResult,
		ProviderFaultDoubleResponse,
		ProviderFaultWrongInterfaceVersion,
		ProviderFaultInvalidProof,
		ProviderFaultAvailabilityViolation:
		return true
	default:
		return false
	}
}

func IsProviderPenaltySource(source ProviderPenaltySource) bool {
	switch source {
	case ProviderPenaltyCollateral,
		ProviderPenaltyServiceStake,
		ProviderPenaltyEscrowedPayment,
		ProviderPenaltyReputationScore:
		return true
	default:
		return false
	}
}

func ComputeServiceReplayProtectionProofHash(proof ServiceReplayProtectionProof) string {
	return servicesHashParts(
		"aetra-services-replay-protection-proof-v1",
		proof.CallID,
		proof.ServiceID,
		proof.MethodID,
		proof.Caller,
		fmt.Sprint(proof.Nonce),
		proof.IdempotencyKey,
		proof.PayloadHash,
		fmt.Sprint(proof.DeadlineHeight),
		proof.ReplayEntryHash,
		fmt.Sprint(proof.ReceiptTombstone),
		proof.TombstoneHash,
		proof.ReplayIndexHash,
	)
}

func ComputeProviderMisbehaviorReportHash(report ProviderMisbehaviorReport) string {
	report = canonicalProviderMisbehaviorReport(report)
	parts := []string{
		"aetra-services-provider-misbehavior-v1",
		report.ServiceID,
		report.ProviderID,
		report.CallID,
		string(report.FaultClass),
		report.EvidenceHash,
		report.ExpectedInterfaceHash,
		report.ObservedInterfaceHash,
		report.ProofHash,
		fmt.Sprint(report.ObservedHeight),
		fmt.Sprint(report.DeadlineHeight),
		fmt.Sprint(report.ReputationDelta),
		report.CollateralSlashAmount,
		report.ServiceStakeSlashAmount,
		report.EscrowForfeitAmount,
		fmt.Sprint(len(report.PenaltySources)),
	}
	for _, source := range report.PenaltySources {
		parts = append(parts, string(source))
	}
	return servicesHashParts(parts...)
}

func ComputeServiceFaultProofHash(proof ServiceFaultProof) string {
	return servicesHashParts(
		"aetra-services-fault-proof-v1",
		proof.ServiceID,
		proof.ProviderID,
		proof.CallID,
		string(proof.FaultClass),
		proof.ReportHash,
		proof.EvidenceHash,
		proof.ProofHash,
		proof.ExpectedInterfaceHash,
		proof.ObservedInterfaceHash,
		proof.ReceiptHash,
		proof.TombstoneHash,
		fmt.Sprint(proof.SubmittedHeight),
	)
}

func ComputeProviderPenaltyRouteEntryHash(entry ProviderPenaltyRouteEntry) string {
	return servicesHashParts(
		"aetra-services-penalty-route-entry-v1",
		string(entry.Source),
		entry.Denom,
		entry.Amount,
		fmt.Sprint(entry.ReputationDelta),
		entry.Recipient,
	)
}

func ComputeProviderPenaltyRouteHash(route ProviderPenaltyRoute) string {
	entries := append([]ProviderPenaltyRouteEntry(nil), route.Entries...)
	sortProviderPenaltyRouteEntries(entries)
	parts := []string{
		"aetra-services-penalty-route-v1",
		route.ServiceID,
		route.ProviderID,
		route.CallID,
		string(route.FaultClass),
		route.ReportHash,
		fmt.Sprint(len(entries)),
	}
	for _, entry := range entries {
		parts = append(parts, entry.RouteEntryHash)
	}
	return servicesHashParts(parts...)
}

func ComputeServiceChallengeID(flow ServiceChallengeFlow) string {
	return servicesHashParts("aetra-services-challenge-id-v1", flow.ServiceID, flow.CallID, flow.ProviderID, string(flow.FaultClass), flow.DisputeMessageHash)
}

func ComputeServiceChallengeFlowHash(flow ServiceChallengeFlow) string {
	return servicesHashParts(
		"aetra-services-challenge-flow-v1",
		flow.ChallengeID,
		flow.ServiceID,
		flow.ProviderID,
		flow.CallID,
		string(flow.FaultClass),
		flow.DisputeMessageHash,
		flow.FaultProofHash,
		flow.PenaltyRouteHash,
		fmt.Sprint(flow.OpenedHeight),
		fmt.Sprint(flow.ChallengeEndHeight),
	)
}

func ComputeServiceReceiptFreshnessProofHash(proof ServiceReceiptFreshnessProof) string {
	return servicesHashParts(
		"aetra-services-receipt-freshness-v1",
		proof.CallID,
		proof.ReceiptHash,
		proof.TombstoneHash,
		fmt.Sprint(proof.ReceiptHeight),
		fmt.Sprint(proof.RetainUntilHeight),
		fmt.Sprint(proof.CurrentHeight),
		fmt.Sprint(proof.Stale),
	)
}

func ComputeServiceSecurityImplementationBundleHash(bundle ServiceSecurityImplementationBundle) string {
	return servicesHashParts(
		"aetra-services-security-implementation-v1",
		bundle.ServiceID,
		string(bundle.TrustModelLabel),
		string(bundle.ExecutionFailureBehaviorLabel),
		bundle.CollateralMessageHash,
		bundle.ChallengeFlowHash,
		bundle.FaultProofHash,
		bundle.PenaltyRouteHash,
		bundle.ReplayControlHash,
		bundle.ReceiptFreshnessHash,
	)
}

func validateProviderFaultClassRules(report ProviderMisbehaviorReport) error {
	switch report.FaultClass {
	case ProviderFaultInvalidResult:
		if !stringInPenaltySources(report.PenaltySources, ProviderPenaltyCollateral) && !stringInPenaltySources(report.PenaltySources, ProviderPenaltyEscrowedPayment) {
			return errors.New("services invalid result fault requires collateral or escrow penalty")
		}
	case ProviderFaultMissingResult:
		if report.CallID == "" || report.DeadlineHeight == 0 || report.ObservedHeight <= report.DeadlineHeight {
			return errors.New("services missing result fault requires expired call deadline")
		}
	case ProviderFaultLateResult:
		if report.CallID == "" || report.DeadlineHeight == 0 || report.ObservedHeight <= report.DeadlineHeight {
			return errors.New("services late result fault requires observed height after deadline")
		}
	case ProviderFaultDoubleResponse:
		if report.CallID == "" {
			return errors.New("services double response fault requires call id")
		}
	case ProviderFaultWrongInterfaceVersion:
		if report.ExpectedInterfaceHash == "" || report.ObservedInterfaceHash == "" || report.ExpectedInterfaceHash == report.ObservedInterfaceHash {
			return errors.New("services wrong interface version fault requires mismatched interface hashes")
		}
	case ProviderFaultInvalidProof:
		if report.ProofHash == "" {
			return errors.New("services invalid proof fault requires proof hash")
		}
	case ProviderFaultAvailabilityViolation:
		if !stringInPenaltySources(report.PenaltySources, ProviderPenaltyReputationScore) {
			return errors.New("services availability violation requires reputation penalty")
		}
	}
	return nil
}

func validateProviderPenaltyRouteEntries(entries []ProviderPenaltyRouteEntry) error {
	if len(entries) == 0 {
		return errors.New("services penalty route requires entries")
	}
	previous := ProviderPenaltySource("")
	seen := map[ProviderPenaltySource]struct{}{}
	for _, entry := range entries {
		if err := entry.Validate(); err != nil {
			return err
		}
		if _, found := seen[entry.Source]; found {
			return fmt.Errorf("services penalty route duplicate source %s", entry.Source)
		}
		seen[entry.Source] = struct{}{}
		if previous != "" && previous >= entry.Source {
			return errors.New("services penalty route entries must be sorted canonically")
		}
		previous = entry.Source
	}
	return nil
}

func sortProviderPenaltyRouteEntries(entries []ProviderPenaltyRouteEntry) {
	sort.SliceStable(entries, func(i, j int) bool { return entries[i].Source < entries[j].Source })
}

func validateProviderPenaltySources(sources []ProviderPenaltySource) error {
	if len(sources) == 0 {
		return errors.New("services provider fault requires penalty source")
	}
	previous := ProviderPenaltySource("")
	seen := map[ProviderPenaltySource]struct{}{}
	for _, source := range sources {
		if !IsProviderPenaltySource(source) {
			return fmt.Errorf("services provider fault unknown penalty source %q", source)
		}
		if _, found := seen[source]; found {
			return fmt.Errorf("services provider fault duplicate penalty source %s", source)
		}
		seen[source] = struct{}{}
		if previous != "" && previous >= source {
			return errors.New("services provider fault penalty sources must be sorted canonically")
		}
		previous = source
	}
	return nil
}

func canonicalProviderMisbehaviorReport(report ProviderMisbehaviorReport) ProviderMisbehaviorReport {
	report.ServiceID = strings.TrimSpace(report.ServiceID)
	report.ProviderID = strings.TrimSpace(report.ProviderID)
	report.CallID = strings.ToLower(strings.TrimSpace(report.CallID))
	report.FaultClass = ProviderMisbehaviorFaultClass(strings.ToLower(strings.TrimSpace(string(report.FaultClass))))
	report.EvidenceHash = strings.ToLower(strings.TrimSpace(report.EvidenceHash))
	report.ExpectedInterfaceHash = strings.ToLower(strings.TrimSpace(report.ExpectedInterfaceHash))
	report.ObservedInterfaceHash = strings.ToLower(strings.TrimSpace(report.ObservedInterfaceHash))
	report.ProofHash = strings.ToLower(strings.TrimSpace(report.ProofHash))
	report.CollateralSlashAmount = strings.TrimSpace(report.CollateralSlashAmount)
	report.ServiceStakeSlashAmount = strings.TrimSpace(report.ServiceStakeSlashAmount)
	report.EscrowForfeitAmount = strings.TrimSpace(report.EscrowForfeitAmount)
	report.ReportHash = strings.ToLower(strings.TrimSpace(report.ReportHash))
	report.PenaltySources = append([]ProviderPenaltySource(nil), report.PenaltySources...)
	for i := range report.PenaltySources {
		report.PenaltySources[i] = ProviderPenaltySource(strings.ToLower(strings.TrimSpace(string(report.PenaltySources[i]))))
	}
	sort.SliceStable(report.PenaltySources, func(i, j int) bool { return report.PenaltySources[i] < report.PenaltySources[j] })
	return report
}

func stringInPenaltySources(sources []ProviderPenaltySource, target ProviderPenaltySource) bool {
	for _, source := range sources {
		if source == target {
			return true
		}
	}
	return false
}
