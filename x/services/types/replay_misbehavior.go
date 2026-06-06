package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
	coretypes "github.com/sovereign-l1/l1/x/aethercore/types"
)

type ProviderMisbehaviorFaultClass string
type ProviderPenaltySource string

const (
	ProviderFaultInvalidResult         ProviderMisbehaviorFaultClass = "invalid_result"
	ProviderFaultMissingResult         ProviderMisbehaviorFaultClass = "missing_result"
	ProviderFaultLateResult            ProviderMisbehaviorFaultClass = "late_result"
	ProviderFaultDoubleResponse        ProviderMisbehaviorFaultClass = "double_response"
	ProviderFaultWrongInterfaceVersion ProviderMisbehaviorFaultClass = "wrong_interface_version"
	ProviderFaultInvalidProof          ProviderMisbehaviorFaultClass = "invalid_proof"
	ProviderFaultAvailabilityViolation ProviderMisbehaviorFaultClass = "availability_violation"

	ProviderPenaltyCollateral      ProviderPenaltySource = "provider_collateral"
	ProviderPenaltyServiceStake    ProviderPenaltySource = "service_stake"
	ProviderPenaltyEscrowedPayment ProviderPenaltySource = "escrowed_payment"
	ProviderPenaltyReputationScore ProviderPenaltySource = "reputation_score"
)

type ServiceReplayProtectionProof struct {
	CallID           string
	ServiceID        string
	MethodID         string
	Caller           string
	Nonce            uint64
	IdempotencyKey   string
	PayloadHash      string
	DeadlineHeight   uint64
	ReplayEntryHash  string
	ReceiptTombstone bool
	TombstoneHash    string
	ReplayIndexHash  string
	ControlHash      string
}

type ProviderMisbehaviorReport struct {
	ServiceID               string
	ProviderID              string
	CallID                  string
	FaultClass              ProviderMisbehaviorFaultClass
	EvidenceHash            string
	ExpectedInterfaceHash   string
	ObservedInterfaceHash   string
	ProofHash               string
	ObservedHeight          uint64
	DeadlineHeight          uint64
	PenaltySources          []ProviderPenaltySource
	CollateralSlashAmount   string
	ServiceStakeSlashAmount string
	EscrowForfeitAmount     string
	ReputationDelta         int64
	ReportHash              string
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
		CallID:           call.CallID,
		ServiceID:        call.TargetService,
		MethodID:         call.MethodID,
		Caller:           call.Caller,
		Nonce:            call.Nonce,
		IdempotencyKey:   call.IdempotencyKey,
		PayloadHash:      call.PayloadHash,
		DeadlineHeight:   call.DeadlineHeight,
		ReplayEntryHash:  entry.EntryHash,
		ReceiptTombstone: tombstoneFound,
		ReplayIndexHash:  index.IndexHash,
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
		"aetheris-services-replay-protection-proof-v1",
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
		"aetheris-services-provider-misbehavior-v1",
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
