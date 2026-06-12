package types

import (
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

type XServiceProvidersStateObject string
type XServiceProvidersMessageName string
type XServiceProvidersQueryName string
type XServiceProvidersFailureMode string
type XServiceProvidersIntegrationPoint string

const (
	XServiceProvidersStateAvailabilityCommitment	XServiceProvidersStateObject	= "AvailabilityCommitment"
	XServiceProvidersStateProviderCollateral	XServiceProvidersStateObject	= "ProviderCollateral"
	XServiceProvidersStateProviderFault		XServiceProvidersStateObject	= "ProviderFault"
	XServiceProvidersStateProviderRecord		XServiceProvidersStateObject	= "ProviderRecord"
	XServiceProvidersStateProviderReputation	XServiceProvidersStateObject	= "ProviderReputation"

	XServiceProvidersMsgRegisterProvider			XServiceProvidersMessageName	= "MsgRegisterProvider"
	XServiceProvidersMsgStakeProviderCollateral		XServiceProvidersMessageName	= "MsgStakeProviderCollateral"
	XServiceProvidersMsgSubmitAvailabilityCommitment	XServiceProvidersMessageName	= "MsgSubmitAvailabilityCommitment"
	XServiceProvidersMsgSubmitProviderFault			XServiceProvidersMessageName	= "MsgSubmitProviderFault"
	XServiceProvidersMsgUnstakeProviderCollateral		XServiceProvidersMessageName	= "MsgUnstakeProviderCollateral"
	XServiceProvidersMsgUpdateProvider			XServiceProvidersMessageName	= "MsgUpdateProvider"

	XServiceProvidersQueryAvailabilityCommitment	XServiceProvidersQueryName	= "QueryAvailabilityCommitment"
	XServiceProvidersQueryProvider			XServiceProvidersQueryName	= "QueryProvider"
	XServiceProvidersQueryProviderCollateral	XServiceProvidersQueryName	= "QueryProviderCollateral"
	XServiceProvidersQueryProviderReputation	XServiceProvidersQueryName	= "QueryProviderReputation"
	XServiceProvidersQueryProvidersByService	XServiceProvidersQueryName	= "QueryProvidersByService"

	XServiceProvidersFailureAvailabilityCommitmentExpired	XServiceProvidersFailureMode	= "availability_commitment_expired"
	XServiceProvidersFailureCollateralInsufficient		XServiceProvidersFailureMode	= "collateral_insufficient"
	XServiceProvidersFailureFaultProofInvalid		XServiceProvidersFailureMode	= "fault_proof_invalid"
	XServiceProvidersFailureReputationNonDeterministic	XServiceProvidersFailureMode	= "reputation_update_not_deterministic"
	XServiceProvidersFailureUnsupportedInterface		XServiceProvidersFailureMode	= "unsupported_interface_advertised"

	XServiceProvidersIntegrationRoutingDiscovery		XServiceProvidersIntegrationPoint	= "routing_discovery_layer"
	XServiceProvidersIntegrationSlashingPenaltyRoute	XServiceProvidersIntegrationPoint	= "slashing_like_penalty_routing"
	XServiceProvidersIntegrationServicePayments		XServiceProvidersIntegrationPoint	= "x/servicepayments"
	XServiceProvidersIntegrationServices			XServiceProvidersIntegrationPoint	= "x/services"
)

type ProviderFault = ProviderMisbehaviorReport

type XServiceProvidersFailureCoverage struct {
	Mode	XServiceProvidersFailureMode
	Guard	string
	Scope	string
}

type XServiceProvidersModuleBreakdown struct {
	ModulePath		string
	Purpose			[]string
	StateObjects		[]XServiceProvidersStateObject
	Messages		[]XServiceProvidersMessageName
	Queries			[]XServiceProvidersQueryName
	FailureModes		[]XServiceProvidersFailureCoverage
	IntegrationPoints	[]XServiceProvidersIntegrationPoint
	BreakdownHash		string
}

type ProviderCollateral struct {
	ServiceID	string
	ProviderID	string
	Denom		string
	Amount		string
	StakedHeight	uint64
	UpdatedHeight	uint64
	CollateralHash	string
}

type AvailabilityCommitment struct {
	ServiceID	string
	ProviderID	string
	InterfaceHash	string
	EndpointHash	string
	WindowStart	uint64
	WindowEnd	uint64
	UptimeTargetBps	uint32
	RenewalNonce	uint64
	SignatureHash	string
	CommitmentHash	string
	RecordHash	string
}

type MsgSubmitAvailabilityCommitment struct {
	Authority	string
	Commitment	AvailabilityCommitment
	MessageHash	string
}

type MsgSubmitProviderFault struct {
	Authority	string
	Report		ProviderMisbehaviorReport
	Proof		ServiceFaultProof
	MessageHash	string
}

type QueryProvider struct {
	ProviderID string
}

type QueryProviderResponse struct {
	Provider	ProviderRecord
	Found		bool
}

type QueryProviderCollateral struct {
	ProviderID string
}

type QueryProviderCollateralResponse struct {
	Collateral	ProviderCollateral
	Found		bool
}

type QueryProviderReputation struct {
	ProviderID string
}

type QueryProviderReputationResponse struct {
	Reputation	ReputationRecord
	Found		bool
}

type QueryAvailabilityCommitment struct {
	ProviderID	string
	ServiceID	string
}

type QueryAvailabilityCommitmentResponse struct {
	Commitment	AvailabilityCommitment
	Found		bool
}

type ServiceProviderState struct {
	Providers	[]ProviderRecord
	Collaterals	[]ProviderCollateral
	Reputations	[]ReputationRecord
	Commitments	[]AvailabilityCommitment
	StateRootHash	string
	UpdatedHeight	uint64
}

func DefaultXServiceProvidersModuleBreakdown() (XServiceProvidersModuleBreakdown, error) {
	breakdown := XServiceProvidersModuleBreakdown{
		ModulePath:	ServiceModuleProviders,
		Purpose: []string{
			"availability_commitments",
			"collateral_management",
			"fog_market_provider_registry",
			"provider_faults",
			"reputation_updates",
		},
		StateObjects: []XServiceProvidersStateObject{
			XServiceProvidersStateAvailabilityCommitment,
			XServiceProvidersStateProviderCollateral,
			XServiceProvidersStateProviderFault,
			XServiceProvidersStateProviderRecord,
			XServiceProvidersStateProviderReputation,
		},
		Messages: []XServiceProvidersMessageName{
			XServiceProvidersMsgRegisterProvider,
			XServiceProvidersMsgStakeProviderCollateral,
			XServiceProvidersMsgSubmitAvailabilityCommitment,
			XServiceProvidersMsgSubmitProviderFault,
			XServiceProvidersMsgUnstakeProviderCollateral,
			XServiceProvidersMsgUpdateProvider,
		},
		Queries: []XServiceProvidersQueryName{
			XServiceProvidersQueryAvailabilityCommitment,
			XServiceProvidersQueryProvider,
			XServiceProvidersQueryProviderCollateral,
			XServiceProvidersQueryProviderReputation,
			XServiceProvidersQueryProvidersByService,
		},
		FailureModes: []XServiceProvidersFailureCoverage{
			newXServiceProvidersFailureCoverage(XServiceProvidersFailureAvailabilityCommitmentExpired, "ValidateAvailabilityCommitmentActive", ServiceStoreV2ProviderPrefix),
			newXServiceProvidersFailureCoverage(XServiceProvidersFailureCollateralInsufficient, "ValidateProviderCollateralSufficient", ServiceStoreV2ProviderPrefix),
			newXServiceProvidersFailureCoverage(XServiceProvidersFailureFaultProofInvalid, "ValidateProviderFaultProof", ServiceStoreV2ProviderPrefix),
			newXServiceProvidersFailureCoverage(XServiceProvidersFailureReputationNonDeterministic, "ApplyDeterministicProviderReputationUpdate", ServiceStoreV2ProviderPrefix),
			newXServiceProvidersFailureCoverage(XServiceProvidersFailureUnsupportedInterface, "ValidateProviderAdvertisesInterface", ServiceStoreV2ProviderPrefix),
		},
		IntegrationPoints: []XServiceProvidersIntegrationPoint{
			XServiceProvidersIntegrationRoutingDiscovery,
			XServiceProvidersIntegrationSlashingPenaltyRoute,
			XServiceProvidersIntegrationServicePayments,
			XServiceProvidersIntegrationServices,
		},
	}
	return NewXServiceProvidersModuleBreakdown(breakdown)
}

func NewXServiceProvidersModuleBreakdown(breakdown XServiceProvidersModuleBreakdown) (XServiceProvidersModuleBreakdown, error) {
	breakdown = canonicalXServiceProvidersModuleBreakdown(breakdown)
	if err := breakdown.ValidateFormat(); err != nil {
		return XServiceProvidersModuleBreakdown{}, err
	}
	breakdown.BreakdownHash = ComputeXServiceProvidersModuleBreakdownHash(breakdown)
	return breakdown, breakdown.Validate()
}

func NewProviderCollateral(record ProviderRecord, stakedHeight uint64) (ProviderCollateral, error) {
	record = coretypes.CanonicalProviderRecord(record)
	if err := record.Validate(); err != nil {
		return ProviderCollateral{}, err
	}
	collateral := ProviderCollateral{
		ServiceID:	record.ServiceID,
		ProviderID:	record.Provider.ProviderID,
		Denom:		record.Provider.CollateralDenom,
		Amount:		record.Provider.CollateralAmount,
		StakedHeight:	stakedHeight,
		UpdatedHeight:	record.Provider.UpdatedHeight,
	}
	collateral.CollateralHash = ComputeProviderCollateralHash(collateral)
	return collateral, collateral.Validate()
}

func NewAvailabilityCommitment(record ProviderRecord, interfaceHash string) (AvailabilityCommitment, error) {
	record = coretypes.CanonicalProviderRecord(record)
	if err := record.Validate(); err != nil {
		return AvailabilityCommitment{}, err
	}
	interfaceHash = strings.ToLower(strings.TrimSpace(interfaceHash))
	if err := coretypes.ValidateHash("x/serviceproviders availability interface hash", interfaceHash); err != nil {
		return AvailabilityCommitment{}, err
	}
	if err := ValidateProviderAdvertisesInterface(record, interfaceHash); err != nil {
		return AvailabilityCommitment{}, err
	}
	source := record.Provider.AvailabilityCommitment
	commitment := AvailabilityCommitment{
		ServiceID:		record.ServiceID,
		ProviderID:		record.Provider.ProviderID,
		InterfaceHash:		interfaceHash,
		EndpointHash:		source.EndpointHash,
		WindowStart:		source.WindowStart,
		WindowEnd:		source.WindowEnd,
		UptimeTargetBps:	source.UptimeTargetBps,
		RenewalNonce:		source.RenewalNonce,
		SignatureHash:		source.SignatureHash,
		CommitmentHash:		source.CommitmentHash,
	}
	commitment.RecordHash = ComputeAvailabilityCommitmentRecordHash(commitment)
	return commitment, commitment.Validate()
}

func NewMsgSubmitAvailabilityCommitment(authority string, commitment AvailabilityCommitment) (MsgSubmitAvailabilityCommitment, error) {
	msg := MsgSubmitAvailabilityCommitment{Authority: strings.TrimSpace(authority), Commitment: commitment}
	msg.MessageHash = ComputeMsgSubmitAvailabilityCommitmentHash(msg)
	return msg, msg.ValidateBasic()
}

func NewMsgSubmitProviderFault(authority string, report ProviderMisbehaviorReport, proof ServiceFaultProof) (MsgSubmitProviderFault, error) {
	msg := MsgSubmitProviderFault{Authority: strings.TrimSpace(authority), Report: report, Proof: proof}
	msg.MessageHash = ComputeMsgSubmitProviderFaultHash(msg)
	return msg, msg.ValidateBasic()
}

func BuildServiceProviderState(providers []ProviderRecord, collaterals []ProviderCollateral, reputations []ReputationRecord, commitments []AvailabilityCommitment, height uint64) (ServiceProviderState, error) {
	state := ServiceProviderState{
		Providers:	cloneProviderRecords(providers),
		Collaterals:	cloneProviderCollaterals(collaterals),
		Reputations:	cloneReputationRecords(reputations),
		Commitments:	cloneAvailabilityCommitments(commitments),
		UpdatedHeight:	height,
	}
	sortProviderRecords(state.Providers)
	sortProviderCollaterals(state.Collaterals)
	sortReputationRecords(state.Reputations)
	sortAvailabilityCommitments(state.Commitments)
	if err := state.ValidateFormat(); err != nil {
		return ServiceProviderState{}, err
	}
	state.StateRootHash = ComputeServiceProviderStateRootHash(state)
	return state, state.Validate()
}

func ValidateProviderCollateralSufficient(collateral ProviderCollateral, requiredDenom, requiredAmount string) error {
	if err := collateral.Validate(); err != nil {
		return err
	}
	requiredDenom = strings.TrimSpace(requiredDenom)
	requiredAmount = strings.TrimSpace(requiredAmount)
	if collateral.Denom != requiredDenom {
		return errors.New("x/serviceproviders collateral denom mismatch")
	}
	ok, err := decimalStringAtLeast(collateral.Amount, requiredAmount)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("x/serviceproviders collateral insufficient")
	}
	return nil
}

func ValidateProviderAdvertisesInterface(record ProviderRecord, interfaceHash string) error {
	record = coretypes.CanonicalProviderRecord(record)
	if err := record.Validate(); err != nil {
		return err
	}
	interfaceHash = strings.ToLower(strings.TrimSpace(interfaceHash))
	if err := coretypes.ValidateHash("x/serviceproviders advertised interface hash", interfaceHash); err != nil {
		return err
	}
	for _, supported := range record.Provider.SupportedInterfaces {
		if supported == interfaceHash {
			return nil
		}
	}
	return errors.New("x/serviceproviders provider advertises unsupported interface")
}

func ValidateAvailabilityCommitmentActive(commitment AvailabilityCommitment, height uint64) error {
	if err := commitment.Validate(); err != nil {
		return err
	}
	if height == 0 {
		return errors.New("x/serviceproviders availability check height must be positive")
	}
	if height > commitment.WindowEnd {
		return errors.New("x/serviceproviders availability commitment expired")
	}
	if height < commitment.WindowStart {
		return errors.New("x/serviceproviders availability commitment is not active yet")
	}
	return nil
}

func ValidateProviderFaultProof(report ProviderMisbehaviorReport, proof ServiceFaultProof) error {
	report = canonicalProviderMisbehaviorReport(report)
	if err := report.Validate(); err != nil {
		return err
	}
	if err := proof.Validate(); err != nil {
		return err
	}
	if proof.ReportHash != report.ReportHash ||
		proof.ServiceID != report.ServiceID ||
		proof.ProviderID != report.ProviderID ||
		proof.CallID != report.CallID ||
		proof.FaultClass != report.FaultClass ||
		proof.EvidenceHash != report.EvidenceHash ||
		proof.ProofHash != report.ProofHash ||
		proof.ExpectedInterfaceHash != report.ExpectedInterfaceHash ||
		proof.ObservedInterfaceHash != report.ObservedInterfaceHash {
		return errors.New("x/serviceproviders fault proof invalid")
	}
	return nil
}

func ApplyDeterministicProviderReputationUpdate(previous ReputationRecord, report ProviderMisbehaviorReport, height uint64, expectedRecordHash string) (ReputationRecord, error) {
	previous = coretypes.CanonicalReputationRecord(previous)
	if err := previous.Validate(); err != nil {
		return ReputationRecord{}, err
	}
	report = canonicalProviderMisbehaviorReport(report)
	if err := report.Validate(); err != nil {
		return ReputationRecord{}, err
	}
	if previous.ProviderID != report.ProviderID {
		return ReputationRecord{}, errors.New("x/serviceproviders reputation provider mismatch")
	}
	if height == 0 || height < previous.UpdatedHeight || height < report.ObservedHeight {
		return ReputationRecord{}, errors.New("x/serviceproviders reputation update height is invalid")
	}
	next := previous
	next.Failures++
	next.UpdatedHeight = height
	if report.ReputationDelta < 0 {
		delta := uint64(-report.ReputationDelta)
		if delta > next.Score {
			next.Score = 0
		} else {
			next.Score -= delta
		}
	}
	next.RecordHash = coretypes.ComputeReputationRecordHash(next)
	if expectedRecordHash != "" && strings.ToLower(strings.TrimSpace(expectedRecordHash)) != next.RecordHash {
		return ReputationRecord{}, errors.New("x/serviceproviders reputation update not deterministic")
	}
	return next, next.Validate()
}

func QueryProviderFromState(state ServiceProviderState, query QueryProvider) (QueryProviderResponse, error) {
	if err := state.Validate(); err != nil {
		return QueryProviderResponse{}, err
	}
	if err := validateInterfaceToken("x/serviceproviders query provider id", query.ProviderID); err != nil {
		return QueryProviderResponse{}, err
	}
	for _, provider := range state.Providers {
		if provider.Provider.ProviderID == query.ProviderID {
			return QueryProviderResponse{Provider: provider, Found: true}, nil
		}
	}
	return QueryProviderResponse{}, nil
}

func QueryProviderCollateralFromState(state ServiceProviderState, query QueryProviderCollateral) (QueryProviderCollateralResponse, error) {
	if err := state.Validate(); err != nil {
		return QueryProviderCollateralResponse{}, err
	}
	if err := validateInterfaceToken("x/serviceproviders query collateral provider id", query.ProviderID); err != nil {
		return QueryProviderCollateralResponse{}, err
	}
	for _, collateral := range state.Collaterals {
		if collateral.ProviderID == query.ProviderID {
			return QueryProviderCollateralResponse{Collateral: collateral, Found: true}, nil
		}
	}
	return QueryProviderCollateralResponse{}, nil
}

func QueryProviderReputationFromState(state ServiceProviderState, query QueryProviderReputation) (QueryProviderReputationResponse, error) {
	if err := state.Validate(); err != nil {
		return QueryProviderReputationResponse{}, err
	}
	if err := validateInterfaceToken("x/serviceproviders query reputation provider id", query.ProviderID); err != nil {
		return QueryProviderReputationResponse{}, err
	}
	for _, reputation := range state.Reputations {
		if reputation.ProviderID == query.ProviderID {
			return QueryProviderReputationResponse{Reputation: reputation, Found: true}, nil
		}
	}
	return QueryProviderReputationResponse{}, nil
}

func QueryAvailabilityCommitmentFromState(state ServiceProviderState, query QueryAvailabilityCommitment) (QueryAvailabilityCommitmentResponse, error) {
	if err := state.Validate(); err != nil {
		return QueryAvailabilityCommitmentResponse{}, err
	}
	if err := validateInterfaceToken("x/serviceproviders query commitment provider id", query.ProviderID); err != nil {
		return QueryAvailabilityCommitmentResponse{}, err
	}
	for _, commitment := range state.Commitments {
		if commitment.ProviderID == query.ProviderID && (query.ServiceID == "" || commitment.ServiceID == query.ServiceID) {
			return QueryAvailabilityCommitmentResponse{Commitment: commitment, Found: true}, nil
		}
	}
	return QueryAvailabilityCommitmentResponse{}, nil
}

func (breakdown XServiceProvidersModuleBreakdown) ValidateFormat() error {
	if breakdown.ModulePath != ServiceModuleProviders {
		return errors.New("x/serviceproviders module path must be x/serviceproviders")
	}
	if err := validateSortedTokens("x/serviceproviders purpose", breakdown.Purpose); err != nil {
		return err
	}
	if err := validateXServiceProvidersEnumSet("state", breakdown.StateObjects, requiredXServiceProvidersStates(), IsXServiceProvidersStateObject); err != nil {
		return err
	}
	if err := validateXServiceProvidersEnumSet("message", breakdown.Messages, requiredXServiceProvidersMessages(), IsXServiceProvidersMessageName); err != nil {
		return err
	}
	if err := validateXServiceProvidersEnumSet("query", breakdown.Queries, requiredXServiceProvidersQueries(), IsXServiceProvidersQueryName); err != nil {
		return err
	}
	if err := validateXServiceProvidersFailureCoverages(breakdown.FailureModes); err != nil {
		return err
	}
	if err := validateXServiceProvidersEnumSet("integration", breakdown.IntegrationPoints, requiredXServiceProvidersIntegrations(), IsXServiceProvidersIntegrationPoint); err != nil {
		return err
	}
	if breakdown.BreakdownHash != "" {
		return coretypes.ValidateHash("x/serviceproviders breakdown hash", breakdown.BreakdownHash)
	}
	return nil
}

func (breakdown XServiceProvidersModuleBreakdown) Validate() error {
	if err := breakdown.ValidateFormat(); err != nil {
		return err
	}
	if breakdown.BreakdownHash == "" {
		return errors.New("x/serviceproviders breakdown hash is required")
	}
	if expected := ComputeXServiceProvidersModuleBreakdownHash(breakdown); breakdown.BreakdownHash != expected {
		return fmt.Errorf("x/serviceproviders breakdown hash mismatch: expected %s", expected)
	}
	return nil
}

func (coverage XServiceProvidersFailureCoverage) Validate() error {
	if !IsXServiceProvidersFailureMode(coverage.Mode) {
		return fmt.Errorf("x/serviceproviders unknown failure mode %q", coverage.Mode)
	}
	if err := validateInterfaceToken("x/serviceproviders failure guard", coverage.Guard); err != nil {
		return err
	}
	if !IsServiceStoreKey(coverage.Scope + "/_") {
		return errors.New("x/serviceproviders failure scope must use services store prefix")
	}
	return nil
}

func (collateral ProviderCollateral) Validate() error {
	if err := validateInterfaceToken("x/serviceproviders collateral service id", collateral.ServiceID); err != nil {
		return err
	}
	if err := validateInterfaceToken("x/serviceproviders collateral provider id", collateral.ProviderID); err != nil {
		return err
	}
	if err := validateInterfaceToken("x/serviceproviders collateral denom", collateral.Denom); err != nil {
		return err
	}
	if err := validatePositivePaymentAmount("x/serviceproviders collateral amount", collateral.Amount); err != nil {
		return err
	}
	if collateral.StakedHeight == 0 || collateral.UpdatedHeight < collateral.StakedHeight {
		return errors.New("x/serviceproviders collateral heights are invalid")
	}
	if err := coretypes.ValidateHash("x/serviceproviders collateral hash", collateral.CollateralHash); err != nil {
		return err
	}
	if expected := ComputeProviderCollateralHash(collateral); collateral.CollateralHash != expected {
		return fmt.Errorf("x/serviceproviders collateral hash mismatch: expected %s", expected)
	}
	return nil
}

func (commitment AvailabilityCommitment) Validate() error {
	if err := validateInterfaceToken("x/serviceproviders availability service id", commitment.ServiceID); err != nil {
		return err
	}
	if err := validateInterfaceToken("x/serviceproviders availability provider id", commitment.ProviderID); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("x/serviceproviders availability interface hash", commitment.InterfaceHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("x/serviceproviders availability endpoint hash", commitment.EndpointHash); err != nil {
		return err
	}
	if commitment.WindowStart == 0 || commitment.WindowEnd <= commitment.WindowStart {
		return errors.New("x/serviceproviders availability window is invalid")
	}
	if commitment.UptimeTargetBps == 0 || commitment.UptimeTargetBps > 10_000 {
		return errors.New("x/serviceproviders availability target must be 1..10000 bps")
	}
	if commitment.RenewalNonce == 0 {
		return errors.New("x/serviceproviders availability renewal nonce must be positive")
	}
	if err := coretypes.ValidateHash("x/serviceproviders availability signature hash", commitment.SignatureHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("x/serviceproviders availability commitment hash", commitment.CommitmentHash); err != nil {
		return err
	}
	source := coretypes.FogAvailabilityCommitment{
		EndpointHash:		commitment.EndpointHash,
		WindowStart:		commitment.WindowStart,
		WindowEnd:		commitment.WindowEnd,
		UptimeTargetBps:	commitment.UptimeTargetBps,
		RenewalNonce:		commitment.RenewalNonce,
		SignatureHash:		commitment.SignatureHash,
		CommitmentHash:		commitment.CommitmentHash,
	}
	if expected := coretypes.ComputeFogAvailabilityCommitmentHash(source); commitment.CommitmentHash != expected {
		return fmt.Errorf("x/serviceproviders availability commitment hash mismatch: expected %s", expected)
	}
	if err := coretypes.ValidateHash("x/serviceproviders availability record hash", commitment.RecordHash); err != nil {
		return err
	}
	if expected := ComputeAvailabilityCommitmentRecordHash(commitment); commitment.RecordHash != expected {
		return fmt.Errorf("x/serviceproviders availability record hash mismatch: expected %s", expected)
	}
	return nil
}

func (msg MsgSubmitAvailabilityCommitment) ValidateBasic() error {
	if err := addressing.ValidateAuthorityAddress("x/serviceproviders submit availability authority", msg.Authority); err != nil {
		return err
	}
	if err := msg.Commitment.Validate(); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("x/serviceproviders submit availability message hash", msg.MessageHash); err != nil {
		return err
	}
	if expected := ComputeMsgSubmitAvailabilityCommitmentHash(msg); msg.MessageHash != expected {
		return fmt.Errorf("x/serviceproviders submit availability message hash mismatch: expected %s", expected)
	}
	return nil
}

func (msg MsgSubmitProviderFault) ValidateBasic() error {
	if err := addressing.ValidateAuthorityAddress("x/serviceproviders submit fault authority", msg.Authority); err != nil {
		return err
	}
	if err := ValidateProviderFaultProof(msg.Report, msg.Proof); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("x/serviceproviders submit fault message hash", msg.MessageHash); err != nil {
		return err
	}
	if expected := ComputeMsgSubmitProviderFaultHash(msg); msg.MessageHash != expected {
		return fmt.Errorf("x/serviceproviders submit fault message hash mismatch: expected %s", expected)
	}
	return nil
}

func (state ServiceProviderState) ValidateFormat() error {
	if state.UpdatedHeight == 0 {
		return errors.New("x/serviceproviders state updated height must be positive")
	}
	if err := validateProviderRecordsForServiceProviderState(state.Providers); err != nil {
		return err
	}
	if err := validateProviderCollaterals(state.Collaterals); err != nil {
		return err
	}
	if err := validateReputationRecordsForServiceProviderState(state.Reputations); err != nil {
		return err
	}
	if err := validateAvailabilityCommitments(state.Commitments); err != nil {
		return err
	}
	if state.StateRootHash != "" {
		return coretypes.ValidateHash("x/serviceproviders state root", state.StateRootHash)
	}
	return nil
}

func (state ServiceProviderState) Validate() error {
	if err := state.ValidateFormat(); err != nil {
		return err
	}
	if state.StateRootHash == "" {
		return errors.New("x/serviceproviders state root is required")
	}
	if expected := ComputeServiceProviderStateRootHash(state); state.StateRootHash != expected {
		return fmt.Errorf("x/serviceproviders state root mismatch: expected %s", expected)
	}
	return nil
}

func ComputeXServiceProvidersModuleBreakdownHash(breakdown XServiceProvidersModuleBreakdown) string {
	breakdown = canonicalXServiceProvidersModuleBreakdown(breakdown)
	parts := []string{"aetra-x-serviceproviders-breakdown-v1", breakdown.ModulePath}
	parts = appendStringParts(parts, "purpose", breakdown.Purpose)
	for _, state := range breakdown.StateObjects {
		parts = append(parts, "state", string(state))
	}
	for _, msg := range breakdown.Messages {
		parts = append(parts, "message", string(msg))
	}
	for _, query := range breakdown.Queries {
		parts = append(parts, "query", string(query))
	}
	for _, failure := range breakdown.FailureModes {
		parts = append(parts, "failure", string(failure.Mode), failure.Guard, failure.Scope)
	}
	for _, integration := range breakdown.IntegrationPoints {
		parts = append(parts, "integration", string(integration))
	}
	return servicesHashParts(parts...)
}

func ComputeProviderCollateralHash(collateral ProviderCollateral) string {
	return servicesHashParts(
		"aetra-x-serviceproviders-collateral-v1",
		collateral.ServiceID,
		collateral.ProviderID,
		collateral.Denom,
		collateral.Amount,
		fmt.Sprint(collateral.StakedHeight),
		fmt.Sprint(collateral.UpdatedHeight),
	)
}

func ComputeAvailabilityCommitmentRecordHash(commitment AvailabilityCommitment) string {
	return servicesHashParts(
		"aetra-x-serviceproviders-availability-record-v1",
		commitment.ServiceID,
		commitment.ProviderID,
		commitment.InterfaceHash,
		commitment.CommitmentHash,
	)
}

func ComputeMsgSubmitAvailabilityCommitmentHash(msg MsgSubmitAvailabilityCommitment) string {
	return servicesHashParts("aetra-x-serviceproviders-msg-submit-availability-v1", msg.Authority, msg.Commitment.RecordHash)
}

func ComputeMsgSubmitProviderFaultHash(msg MsgSubmitProviderFault) string {
	return servicesHashParts("aetra-x-serviceproviders-msg-submit-fault-v1", msg.Authority, msg.Report.ReportHash, msg.Proof.FaultProofHash)
}

func ComputeServiceProviderStateRootHash(state ServiceProviderState) string {
	state.Providers = cloneProviderRecords(state.Providers)
	state.Collaterals = cloneProviderCollaterals(state.Collaterals)
	state.Reputations = cloneReputationRecords(state.Reputations)
	state.Commitments = cloneAvailabilityCommitments(state.Commitments)
	sortProviderRecords(state.Providers)
	sortProviderCollaterals(state.Collaterals)
	sortReputationRecords(state.Reputations)
	sortAvailabilityCommitments(state.Commitments)
	parts := []string{"aetra-x-serviceproviders-state-root-v1", fmt.Sprint(state.UpdatedHeight)}
	for _, provider := range state.Providers {
		parts = append(parts, "provider", provider.RecordHash)
	}
	for _, collateral := range state.Collaterals {
		parts = append(parts, "collateral", collateral.CollateralHash)
	}
	for _, reputation := range state.Reputations {
		parts = append(parts, "reputation", reputation.RecordHash)
	}
	for _, commitment := range state.Commitments {
		parts = append(parts, "availability", commitment.RecordHash)
	}
	return servicesHashParts(parts...)
}

func IsXServiceProvidersStateObject(value XServiceProvidersStateObject) bool {
	switch value {
	case XServiceProvidersStateAvailabilityCommitment, XServiceProvidersStateProviderCollateral, XServiceProvidersStateProviderFault, XServiceProvidersStateProviderRecord, XServiceProvidersStateProviderReputation:
		return true
	default:
		return false
	}
}

func IsXServiceProvidersMessageName(value XServiceProvidersMessageName) bool {
	switch value {
	case XServiceProvidersMsgRegisterProvider, XServiceProvidersMsgStakeProviderCollateral, XServiceProvidersMsgSubmitAvailabilityCommitment, XServiceProvidersMsgSubmitProviderFault, XServiceProvidersMsgUnstakeProviderCollateral, XServiceProvidersMsgUpdateProvider:
		return true
	default:
		return false
	}
}

func IsXServiceProvidersQueryName(value XServiceProvidersQueryName) bool {
	switch value {
	case XServiceProvidersQueryAvailabilityCommitment, XServiceProvidersQueryProvider, XServiceProvidersQueryProviderCollateral, XServiceProvidersQueryProviderReputation, XServiceProvidersQueryProvidersByService:
		return true
	default:
		return false
	}
}

func IsXServiceProvidersFailureMode(value XServiceProvidersFailureMode) bool {
	switch value {
	case XServiceProvidersFailureAvailabilityCommitmentExpired, XServiceProvidersFailureCollateralInsufficient, XServiceProvidersFailureFaultProofInvalid, XServiceProvidersFailureReputationNonDeterministic, XServiceProvidersFailureUnsupportedInterface:
		return true
	default:
		return false
	}
}

func IsXServiceProvidersIntegrationPoint(value XServiceProvidersIntegrationPoint) bool {
	switch value {
	case XServiceProvidersIntegrationRoutingDiscovery, XServiceProvidersIntegrationSlashingPenaltyRoute, XServiceProvidersIntegrationServicePayments, XServiceProvidersIntegrationServices:
		return true
	default:
		return false
	}
}

func newXServiceProvidersFailureCoverage(mode XServiceProvidersFailureMode, guard, scope string) XServiceProvidersFailureCoverage {
	return XServiceProvidersFailureCoverage{Mode: mode, Guard: guard, Scope: scope}
}

func canonicalXServiceProvidersModuleBreakdown(breakdown XServiceProvidersModuleBreakdown) XServiceProvidersModuleBreakdown {
	breakdown.ModulePath = strings.TrimSpace(breakdown.ModulePath)
	sort.Strings(breakdown.Purpose)
	sort.SliceStable(breakdown.StateObjects, func(i, j int) bool { return breakdown.StateObjects[i] < breakdown.StateObjects[j] })
	sort.SliceStable(breakdown.Messages, func(i, j int) bool { return breakdown.Messages[i] < breakdown.Messages[j] })
	sort.SliceStable(breakdown.Queries, func(i, j int) bool { return breakdown.Queries[i] < breakdown.Queries[j] })
	sort.SliceStable(breakdown.FailureModes, func(i, j int) bool { return breakdown.FailureModes[i].Mode < breakdown.FailureModes[j].Mode })
	sort.SliceStable(breakdown.IntegrationPoints, func(i, j int) bool { return breakdown.IntegrationPoints[i] < breakdown.IntegrationPoints[j] })
	breakdown.BreakdownHash = strings.ToLower(strings.TrimSpace(breakdown.BreakdownHash))
	return breakdown
}

func validateXServiceProvidersEnumSet[T ~string](label string, values []T, required []T, allowed func(T) bool) error {
	if len(values) != len(required) {
		return fmt.Errorf("x/serviceproviders expected %d %s entries", len(required), label)
	}
	seen := map[T]struct{}{}
	previous := ""
	for _, value := range values {
		if !allowed(value) {
			return fmt.Errorf("x/serviceproviders unknown %s %q", label, value)
		}
		current := string(value)
		if previous != "" && previous >= current {
			return fmt.Errorf("x/serviceproviders %s entries must be sorted canonically", label)
		}
		previous = current
		if _, found := seen[value]; found {
			return fmt.Errorf("x/serviceproviders duplicate %s %s", label, value)
		}
		seen[value] = struct{}{}
	}
	for _, value := range required {
		if _, found := seen[value]; !found {
			return fmt.Errorf("x/serviceproviders missing %s %s", label, value)
		}
	}
	return nil
}

func validateXServiceProvidersFailureCoverages(values []XServiceProvidersFailureCoverage) error {
	required := requiredXServiceProvidersFailures()
	if len(values) != len(required) {
		return fmt.Errorf("x/serviceproviders expected %d failure entries", len(required))
	}
	seen := map[XServiceProvidersFailureMode]struct{}{}
	previous := ""
	for _, value := range values {
		if err := value.Validate(); err != nil {
			return err
		}
		current := string(value.Mode)
		if previous != "" && previous >= current {
			return errors.New("x/serviceproviders failure entries must be sorted canonically")
		}
		previous = current
		if _, found := seen[value.Mode]; found {
			return fmt.Errorf("x/serviceproviders duplicate failure %s", value.Mode)
		}
		seen[value.Mode] = struct{}{}
	}
	for _, value := range required {
		if _, found := seen[value]; !found {
			return fmt.Errorf("x/serviceproviders missing failure %s", value)
		}
	}
	return nil
}

func requiredXServiceProvidersStates() []XServiceProvidersStateObject {
	return []XServiceProvidersStateObject{XServiceProvidersStateAvailabilityCommitment, XServiceProvidersStateProviderCollateral, XServiceProvidersStateProviderFault, XServiceProvidersStateProviderRecord, XServiceProvidersStateProviderReputation}
}

func requiredXServiceProvidersMessages() []XServiceProvidersMessageName {
	return []XServiceProvidersMessageName{XServiceProvidersMsgRegisterProvider, XServiceProvidersMsgStakeProviderCollateral, XServiceProvidersMsgSubmitAvailabilityCommitment, XServiceProvidersMsgSubmitProviderFault, XServiceProvidersMsgUnstakeProviderCollateral, XServiceProvidersMsgUpdateProvider}
}

func requiredXServiceProvidersQueries() []XServiceProvidersQueryName {
	return []XServiceProvidersQueryName{XServiceProvidersQueryAvailabilityCommitment, XServiceProvidersQueryProvider, XServiceProvidersQueryProviderCollateral, XServiceProvidersQueryProviderReputation, XServiceProvidersQueryProvidersByService}
}

func requiredXServiceProvidersFailures() []XServiceProvidersFailureMode {
	return []XServiceProvidersFailureMode{XServiceProvidersFailureAvailabilityCommitmentExpired, XServiceProvidersFailureCollateralInsufficient, XServiceProvidersFailureFaultProofInvalid, XServiceProvidersFailureReputationNonDeterministic, XServiceProvidersFailureUnsupportedInterface}
}

func requiredXServiceProvidersIntegrations() []XServiceProvidersIntegrationPoint {
	return []XServiceProvidersIntegrationPoint{XServiceProvidersIntegrationRoutingDiscovery, XServiceProvidersIntegrationSlashingPenaltyRoute, XServiceProvidersIntegrationServicePayments, XServiceProvidersIntegrationServices}
}

func validateProviderRecordsForServiceProviderState(records []ProviderRecord) error {
	previous := ""
	for _, record := range records {
		if err := record.Validate(); err != nil {
			return err
		}
		if previous != "" && previous >= record.Provider.ProviderID {
			return errors.New("x/serviceproviders provider records must be sorted canonically")
		}
		previous = record.Provider.ProviderID
	}
	return nil
}

func validateProviderCollaterals(collaterals []ProviderCollateral) error {
	previous := ""
	for _, collateral := range collaterals {
		if err := collateral.Validate(); err != nil {
			return err
		}
		if previous != "" && previous >= collateral.ProviderID {
			return errors.New("x/serviceproviders collaterals must be sorted canonically")
		}
		previous = collateral.ProviderID
	}
	return nil
}

func validateReputationRecordsForServiceProviderState(records []ReputationRecord) error {
	previous := ""
	for _, record := range records {
		if err := record.Validate(); err != nil {
			return err
		}
		if previous != "" && previous >= record.ProviderID {
			return errors.New("x/serviceproviders reputations must be sorted canonically")
		}
		previous = record.ProviderID
	}
	return nil
}

func validateAvailabilityCommitments(commitments []AvailabilityCommitment) error {
	previous := ""
	for _, commitment := range commitments {
		if err := commitment.Validate(); err != nil {
			return err
		}
		key := commitment.ProviderID + "/" + commitment.ServiceID
		if previous != "" && previous >= key {
			return errors.New("x/serviceproviders availability commitments must be sorted canonically")
		}
		previous = key
	}
	return nil
}

func decimalStringAtLeast(actual, required string) (bool, error) {
	actualInt, ok := new(big.Int).SetString(strings.TrimSpace(actual), 10)
	if !ok || actualInt.Sign() <= 0 {
		return false, errors.New("x/serviceproviders actual amount must be positive decimal")
	}
	requiredInt, ok := new(big.Int).SetString(strings.TrimSpace(required), 10)
	if !ok || requiredInt.Sign() <= 0 {
		return false, errors.New("x/serviceproviders required amount must be positive decimal")
	}
	return actualInt.Cmp(requiredInt) >= 0, nil
}

func cloneProviderRecords(records []ProviderRecord) []ProviderRecord {
	out := make([]ProviderRecord, len(records))
	copy(out, records)
	for i := range out {
		out[i].Provider.SupportedInterfaces = append([]string(nil), out[i].Provider.SupportedInterfaces...)
	}
	return out
}

func cloneProviderCollaterals(collaterals []ProviderCollateral) []ProviderCollateral {
	out := make([]ProviderCollateral, len(collaterals))
	copy(out, collaterals)
	return out
}

func cloneReputationRecords(records []ReputationRecord) []ReputationRecord {
	out := make([]ReputationRecord, len(records))
	copy(out, records)
	return out
}

func cloneAvailabilityCommitments(commitments []AvailabilityCommitment) []AvailabilityCommitment {
	out := make([]AvailabilityCommitment, len(commitments))
	copy(out, commitments)
	return out
}

func sortProviderRecords(records []ProviderRecord) {
	sort.SliceStable(records, func(i, j int) bool { return records[i].Provider.ProviderID < records[j].Provider.ProviderID })
}

func sortProviderCollaterals(collaterals []ProviderCollateral) {
	sort.SliceStable(collaterals, func(i, j int) bool { return collaterals[i].ProviderID < collaterals[j].ProviderID })
}

func sortReputationRecords(records []ReputationRecord) {
	sort.SliceStable(records, func(i, j int) bool { return records[i].ProviderID < records[j].ProviderID })
}

func sortAvailabilityCommitments(commitments []AvailabilityCommitment) {
	sort.SliceStable(commitments, func(i, j int) bool {
		return commitments[i].ProviderID+"/"+commitments[i].ServiceID < commitments[j].ProviderID+"/"+commitments[j].ServiceID
	})
}
