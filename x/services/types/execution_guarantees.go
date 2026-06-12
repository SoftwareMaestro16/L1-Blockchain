package types

import (
	"errors"
	"fmt"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

type ExecutionGuaranteeReport struct {
	ServiceID					string
	ServiceType					coretypes.ServiceType
	CallID						string
	InterfaceHash					string
	PaymentModelHash				string
	ReplayControlHash				string
	ReceiptHash					string
	PenaltyRouteHash				string
	DiscoveryCacheHash				string
	AvailabilityCommitmentHash			string
	OnChainPathsDeterministic			bool
	HybridResultsVerifiableOrChallengeable		bool
	OffChainCallsSignedAndReplayProtected		bool
	ProviderMisbehaviorEconomicallyPenalized	bool
	AnchoredReceiptsDeterministic			bool
	PaymentRulesKnownBeforeSigning			bool
	InterfaceHashKnownBeforeConstruction		bool
	EndpointAvailabilityBackedByCommitment		bool
	UIGenerationClientResponsibility		bool
	OffChainResultCorrectnessVerifiedOrSettled	bool
	CachedDiscoveryAuthoritative			bool
	GuaranteeHash					string
}

type ExecutionGuaranteeInput struct {
	Context				coretypes.ServiceConsensusContext
	Descriptor			ServiceDescriptor
	Call				UnifiedServiceCall
	ReplayProof			ServiceReplayProtectionProof
	Receipt				ServiceReceipt
	PaymentModel			ServicePaymentModel
	PenaltyRoute			ProviderPenaltyRoute
	DiscoveryCache			ServiceDiscoveryCacheRecord
	AvailabilityCommitmentHash	string
	ResultProofVerified		bool
	ChallengePeriodElapsed		bool
}

func NewExecutionGuaranteeReport(input ExecutionGuaranteeInput) (ExecutionGuaranteeReport, error) {
	if err := input.Context.Validate(); err != nil {
		return ExecutionGuaranteeReport{}, err
	}
	descriptor := coretypes.CanonicalServiceDescriptor(input.Descriptor)
	if err := descriptor.Validate(); err != nil {
		return ExecutionGuaranteeReport{}, err
	}
	if err := ValidateUnifiedServiceCallForDescriptor(input.Context, descriptor, input.Call); err != nil {
		return ExecutionGuaranteeReport{}, err
	}
	if err := input.ReplayProof.Validate(); err != nil {
		return ExecutionGuaranteeReport{}, err
	}
	if input.ReplayProof.CallID != input.Call.CallID || input.ReplayProof.ServiceID != descriptor.ServiceID || input.ReplayProof.MethodID != input.Call.MethodID {
		return ExecutionGuaranteeReport{}, errors.New("services execution guarantee replay proof mismatch")
	}
	if err := input.PaymentModel.Validate(); err != nil {
		return ExecutionGuaranteeReport{}, err
	}
	if input.PaymentModel.ServiceID != descriptor.ServiceID {
		return ExecutionGuaranteeReport{}, errors.New("services execution guarantee payment model service mismatch")
	}
	if input.PaymentModel.DefaultDenom != input.Call.Payment.Denom || input.PaymentModel.SettlementMode != descriptor.Payment.SettlementMode {
		return ExecutionGuaranteeReport{}, errors.New("services execution guarantee payment model mismatch")
	}
	receiptDeterministic := false
	if input.Receipt.ReceiptHash != "" {
		if input.Receipt.CallID != input.Call.CallID || input.Receipt.ServiceID != descriptor.ServiceID {
			return ExecutionGuaranteeReport{}, errors.New("services execution guarantee receipt mismatch")
		}
		if _, err := NewServiceReceiptCanonicalView(input.Receipt); err != nil {
			return ExecutionGuaranteeReport{}, err
		}
		receiptDeterministic = true
	}
	penalized := !serviceGuaranteeRequiresPenalty(descriptor)
	if input.PenaltyRoute.RouteHash != "" {
		if err := input.PenaltyRoute.Validate(); err != nil {
			return ExecutionGuaranteeReport{}, err
		}
		if input.PenaltyRoute.ServiceID != descriptor.ServiceID {
			return ExecutionGuaranteeReport{}, errors.New("services execution guarantee penalty route service mismatch")
		}
		penalized = true
	}
	cacheAuthoritative := false
	if input.DiscoveryCache.CacheHash != "" {
		if err := input.DiscoveryCache.ValidateFormat(); err != nil {
			return ExecutionGuaranteeReport{}, err
		}
		cacheAuthoritative = input.DiscoveryCache.ProofHeightOptional != 0 || input.DiscoveryCache.SignatureOptional != ""
	}
	availabilityBacked := input.AvailabilityCommitmentHash != ""
	if availabilityBacked {
		if err := coretypes.ValidateHash("services execution guarantee availability commitment hash", input.AvailabilityCommitmentHash); err != nil {
			return ExecutionGuaranteeReport{}, err
		}
	}
	report := ExecutionGuaranteeReport{
		ServiceID:					descriptor.ServiceID,
		ServiceType:					descriptor.ServiceType,
		CallID:						input.Call.CallID,
		InterfaceHash:					descriptor.Interface.InterfaceHash,
		PaymentModelHash:				input.PaymentModel.ModelHash,
		ReplayControlHash:				input.ReplayProof.ControlHash,
		ReceiptHash:					input.Receipt.ReceiptHash,
		PenaltyRouteHash:				input.PenaltyRoute.RouteHash,
		DiscoveryCacheHash:				input.DiscoveryCache.CacheHash,
		AvailabilityCommitmentHash:			input.AvailabilityCommitmentHash,
		OnChainPathsDeterministic:			descriptor.ServiceType != coretypes.ServiceTypeOnChain || serviceOnChainPathDeterministic(descriptor),
		HybridResultsVerifiableOrChallengeable:		descriptor.ServiceType != coretypes.ServiceTypeMixed || serviceHybridResultVerifiableOrChallengeable(descriptor),
		OffChainCallsSignedAndReplayProtected:		!serviceHasOffChainExecution(descriptor) || serviceCallSignedAndReplayProtected(input.Call, input.ReplayProof),
		ProviderMisbehaviorEconomicallyPenalized:	penalized,
		AnchoredReceiptsDeterministic:			receiptDeterministic,
		PaymentRulesKnownBeforeSigning:			input.PaymentModel.KnownBeforeSigning,
		InterfaceHashKnownBeforeConstruction:		input.Call.InterfaceHash == descriptor.Interface.InterfaceHash && input.Call.InterfaceHash != "",
		EndpointAvailabilityBackedByCommitment:		availabilityBacked,
		UIGenerationClientResponsibility:		true,
		OffChainResultCorrectnessVerifiedOrSettled:	!serviceHasOffChainExecution(descriptor) || input.ResultProofVerified || input.ChallengePeriodElapsed,
		CachedDiscoveryAuthoritative:			cacheAuthoritative,
	}
	report.GuaranteeHash = ComputeExecutionGuaranteeReportHash(report)
	return report, report.Validate()
}

func (report ExecutionGuaranteeReport) Validate() error {
	if err := validateInterfaceToken("services execution guarantee service id", report.ServiceID); err != nil {
		return err
	}
	if !coretypes.IsServiceType(report.ServiceType) {
		return fmt.Errorf("services execution guarantee unknown service type %q", report.ServiceType)
	}
	if err := coretypes.ValidateHash("services execution guarantee call id", report.CallID); err != nil {
		return err
	}
	for label, value := range map[string]string{
		"interface":		report.InterfaceHash,
		"payment model":	report.PaymentModelHash,
		"replay":		report.ReplayControlHash,
	} {
		if err := coretypes.ValidateHash("services execution guarantee "+label+" hash", value); err != nil {
			return err
		}
	}
	for label, value := range map[string]string{
		"receipt":	report.ReceiptHash,
		"penalty":	report.PenaltyRouteHash,
		"cache":	report.DiscoveryCacheHash,
		"availability":	report.AvailabilityCommitmentHash,
	} {
		if value == "" {
			continue
		}
		if err := coretypes.ValidateHash("services execution guarantee "+label+" hash", value); err != nil {
			return err
		}
	}
	if !report.OnChainPathsDeterministic {
		return errors.New("services execution guarantee requires deterministic on-chain paths")
	}
	if !report.HybridResultsVerifiableOrChallengeable {
		return errors.New("services execution guarantee requires verifiable or challengeable hybrid results")
	}
	if !report.OffChainCallsSignedAndReplayProtected {
		return errors.New("services execution guarantee requires signed replay-protected off-chain calls")
	}
	if !report.ProviderMisbehaviorEconomicallyPenalized {
		return errors.New("services execution guarantee requires economic provider penalties where declared")
	}
	if !report.AnchoredReceiptsDeterministic {
		return errors.New("services execution guarantee requires deterministic anchored receipts")
	}
	if !report.PaymentRulesKnownBeforeSigning {
		return errors.New("services execution guarantee requires payment rules before signing")
	}
	if !report.InterfaceHashKnownBeforeConstruction {
		return errors.New("services execution guarantee requires interface hash before construction")
	}
	if !report.UIGenerationClientResponsibility {
		return errors.New("services execution non-guarantee requires UI generation to remain client responsibility")
	}
	if !report.OffChainResultCorrectnessVerifiedOrSettled && serviceReportHasOffChainExecution(report.ServiceType) {
		return errors.New("services execution non-guarantee forbids assuming off-chain correctness before proof or challenge settlement")
	}
	if report.CachedDiscoveryAuthoritative && report.DiscoveryCacheHash == "" {
		return errors.New("services execution cache authority requires discovery cache hash")
	}
	if err := coretypes.ValidateHash("services execution guarantee hash", report.GuaranteeHash); err != nil {
		return err
	}
	if expected := ComputeExecutionGuaranteeReportHash(report); report.GuaranteeHash != expected {
		return fmt.Errorf("services execution guarantee hash mismatch: expected %s", expected)
	}
	return nil
}

func ComputeExecutionGuaranteeReportHash(report ExecutionGuaranteeReport) string {
	return servicesHashParts(
		"aetra-services-execution-guarantees-v1",
		report.ServiceID,
		string(report.ServiceType),
		report.CallID,
		report.InterfaceHash,
		report.PaymentModelHash,
		report.ReplayControlHash,
		report.ReceiptHash,
		report.PenaltyRouteHash,
		report.DiscoveryCacheHash,
		report.AvailabilityCommitmentHash,
		fmt.Sprint(report.OnChainPathsDeterministic),
		fmt.Sprint(report.HybridResultsVerifiableOrChallengeable),
		fmt.Sprint(report.OffChainCallsSignedAndReplayProtected),
		fmt.Sprint(report.ProviderMisbehaviorEconomicallyPenalized),
		fmt.Sprint(report.AnchoredReceiptsDeterministic),
		fmt.Sprint(report.PaymentRulesKnownBeforeSigning),
		fmt.Sprint(report.InterfaceHashKnownBeforeConstruction),
		fmt.Sprint(report.EndpointAvailabilityBackedByCommitment),
		fmt.Sprint(report.UIGenerationClientResponsibility),
		fmt.Sprint(report.OffChainResultCorrectnessVerifiedOrSettled),
		fmt.Sprint(report.CachedDiscoveryAuthoritative),
	)
}

func serviceOnChainPathDeterministic(descriptor ServiceDescriptor) bool {
	if !descriptor.Execution.Deterministic || descriptor.Verification.TrustModel != coretypes.ServiceTrustConsensusExecuted || descriptor.Verification.Model != coretypes.ServiceVerificationConsensusReceipt {
		return false
	}
	for _, method := range descriptor.Interface.Methods {
		if method.GasModel == "" {
			return false
		}
	}
	return true
}

func serviceHybridResultVerifiableOrChallengeable(descriptor ServiceDescriptor) bool {
	if descriptor.Verification.Model == coretypes.ServiceVerificationProofAnchored || descriptor.Verification.Model == coretypes.ServiceVerificationChallengeWindow || descriptor.Verification.Model == coretypes.ServiceVerificationEconomicCollateral {
		return true
	}
	return descriptor.Verification.ChallengeWindow != 0 || descriptor.Execution.ChallengeWindow != 0 || descriptor.Verification.FallbackServiceID != ""
}

func serviceCallSignedAndReplayProtected(call UnifiedServiceCall, proof ServiceReplayProtectionProof) bool {
	return call.SignatureHash != "" &&
		proof.ControlHash != "" &&
		proof.CallID == call.CallID &&
		proof.ServiceID == call.TargetService &&
		proof.MethodID == call.MethodID &&
		proof.Nonce == call.Nonce &&
		proof.IdempotencyKey == call.IdempotencyKey &&
		proof.PayloadHash == call.PayloadHash &&
		proof.DeadlineHeight == call.DeadlineHeight
}

func serviceHasOffChainExecution(descriptor ServiceDescriptor) bool {
	switch descriptor.ServiceType {
	case coretypes.ServiceTypeOffChain, coretypes.ServiceTypeMixed, coretypes.ServiceTypeFogMarket:
		return true
	default:
		return false
	}
}

func serviceReportHasOffChainExecution(serviceType coretypes.ServiceType) bool {
	switch serviceType {
	case coretypes.ServiceTypeOffChain, coretypes.ServiceTypeMixed, coretypes.ServiceTypeFogMarket:
		return true
	default:
		return false
	}
}

func serviceGuaranteeRequiresPenalty(descriptor ServiceDescriptor) bool {
	return descriptor.Verification.TrustModel == coretypes.ServiceTrustEconomicallySecured ||
		descriptor.ServiceType == coretypes.ServiceTypeFogMarket ||
		descriptor.Verification.FaultPolicy == coretypes.ServiceFailureSlashProvider
}
