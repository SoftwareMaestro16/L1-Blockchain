package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

type ServiceTrustModelLabel string
type ServiceFailureBehaviorLabel string

const (
	ServiceTrustLabelFullyTrusted			ServiceTrustModelLabel	= "fully_trusted"
	ServiceTrustLabelEconomicallySecured		ServiceTrustModelLabel	= "economically_secured"
	ServiceTrustLabelCryptographicallyVerifiable	ServiceTrustModelLabel	= "cryptographically_verifiable"
	ServiceTrustLabelConsensusExecuted		ServiceTrustModelLabel	= "consensus_executed"
	ServiceTrustLabelHybridChallengeable		ServiceTrustModelLabel	= "hybrid_challengeable"

	ServiceFailureLabelRevert		ServiceFailureBehaviorLabel	= "revert"
	ServiceFailureLabelRetry		ServiceFailureBehaviorLabel	= "retry"
	ServiceFailureLabelFallbackOnChain	ServiceFailureBehaviorLabel	= "fallback_on_chain"
	ServiceFailureLabelChallenge		ServiceFailureBehaviorLabel	= "challenge"
	ServiceFailureLabelSlashProvider	ServiceFailureBehaviorLabel	= "slash_provider"
	ServiceFailureLabelRefund		ServiceFailureBehaviorLabel	= "refund"
	ServiceFailureLabelPartialSettle	ServiceFailureBehaviorLabel	= "partial_settle"
)

type ServiceMethodSecurityPolicy struct {
	MethodID			string
	TrustModel			coretypes.ServiceTrustModel
	TrustModelLabel			ServiceTrustModelLabel
	VerificationModel		coretypes.ServiceVerificationModel
	FailureBehavior			coretypes.ServiceFailureBehavior
	FailureBehaviorLabel		ServiceFailureBehaviorLabel
	ConsensusCriticalAllowed	bool
	MethodSecurityHash		string
}

type ServiceSecurityPolicy struct {
	ServiceID			string
	ServiceType			coretypes.ServiceType
	TrustModel			coretypes.ServiceTrustModel
	TrustModelLabel			ServiceTrustModelLabel
	VerificationModel		coretypes.ServiceVerificationModel
	ProofFormat			string
	ProviderCollateralDenom		string
	ProviderCollateralAmount	string
	ChallengeWindow			uint64
	FallbackServiceID		string
	FaultPolicy			coretypes.ServiceFailureBehavior
	FaultPolicyLabel		ServiceFailureBehaviorLabel
	ExecutionFailureBehavior	coretypes.ServiceFailureBehavior
	ExecutionFailureBehaviorLabel	ServiceFailureBehaviorLabel
	ConsensusCriticalRequested	bool
	ConsensusCriticalAllowed	bool
	RequiresIndependentVerification	bool
	RequiresCollateralPenalty	bool
	RequiresProofFormat		bool
	RequiresDeterministicGas	bool
	RequiresChallengeFallback	bool
	MethodPolicies			[]ServiceMethodSecurityPolicy
	SecurityHash			string
}

func NewServiceSecurityPolicy(descriptor ServiceDescriptor, consensusCritical bool) (ServiceSecurityPolicy, error) {
	descriptor = coretypes.CanonicalServiceDescriptor(descriptor)
	if err := descriptor.Validate(); err != nil {
		return ServiceSecurityPolicy{}, err
	}
	policy := ServiceSecurityPolicy{
		ServiceID:				descriptor.ServiceID,
		ServiceType:				descriptor.ServiceType,
		TrustModel:				descriptor.Verification.TrustModel,
		TrustModelLabel:			MustServiceTrustModelLabel(descriptor.Verification.TrustModel),
		VerificationModel:			descriptor.Verification.Model,
		ProofFormat:				descriptor.Verification.ProofFormat,
		ProviderCollateralDenom:		descriptor.Verification.ProviderCollateralDenom,
		ProviderCollateralAmount:		descriptor.Verification.ProviderCollateralAmount,
		ChallengeWindow:			maxUint64(descriptor.Verification.ChallengeWindow, descriptor.Execution.ChallengeWindow),
		FallbackServiceID:			descriptor.Verification.FallbackServiceID,
		FaultPolicy:				descriptor.Verification.FaultPolicy,
		FaultPolicyLabel:			serviceFailureBehaviorLabelOrEmpty(descriptor.Verification.FaultPolicy),
		ExecutionFailureBehavior:		descriptor.Execution.FailureBehavior,
		ExecutionFailureBehaviorLabel:		MustServiceFailureBehaviorLabel(descriptor.Execution.FailureBehavior),
		ConsensusCriticalRequested:		consensusCritical,
		RequiresIndependentVerification:	descriptor.Verification.TrustModel == coretypes.ServiceTrustFullyTrusted,
		RequiresCollateralPenalty:		descriptor.Verification.TrustModel == coretypes.ServiceTrustEconomicallySecured,
		RequiresProofFormat:			descriptor.Verification.TrustModel == coretypes.ServiceTrustCryptographicallyVerifiable,
		RequiresDeterministicGas:		descriptor.Verification.TrustModel == coretypes.ServiceTrustConsensusExecuted,
		RequiresChallengeFallback:		descriptor.Verification.TrustModel == coretypes.ServiceTrustHybridChallengeable,
	}
	policy.ConsensusCriticalAllowed = serviceConsensusCriticalAllowed(descriptor, consensusCritical)
	for _, method := range descriptor.Interface.Methods {
		methodPolicy, err := NewServiceMethodSecurityPolicy(descriptor, method, consensusCritical)
		if err != nil {
			return ServiceSecurityPolicy{}, err
		}
		policy.MethodPolicies = append(policy.MethodPolicies, methodPolicy)
	}
	sortServiceMethodSecurityPolicies(policy.MethodPolicies)
	policy.SecurityHash = ComputeServiceSecurityPolicyHash(policy)
	return policy, policy.Validate()
}

func NewServiceMethodSecurityPolicy(descriptor ServiceDescriptor, method coretypes.ServiceMethodDescriptor, consensusCritical bool) (ServiceMethodSecurityPolicy, error) {
	if method.FailureBehavior == "" {
		return ServiceMethodSecurityPolicy{}, fmt.Errorf("services security method %s must declare failure behavior", method.MethodID)
	}
	if !coretypes.IsServiceFailureBehavior(method.FailureBehavior) {
		return ServiceMethodSecurityPolicy{}, fmt.Errorf("services security method %s has unknown failure behavior %q", method.MethodID, method.FailureBehavior)
	}
	policy := ServiceMethodSecurityPolicy{
		MethodID:			method.MethodID,
		TrustModel:			descriptor.Verification.TrustModel,
		TrustModelLabel:		MustServiceTrustModelLabel(descriptor.Verification.TrustModel),
		VerificationModel:		method.VerificationModel,
		FailureBehavior:		method.FailureBehavior,
		FailureBehaviorLabel:		MustServiceFailureBehaviorLabel(method.FailureBehavior),
		ConsensusCriticalAllowed:	serviceConsensusCriticalAllowed(descriptor, consensusCritical),
	}
	policy.MethodSecurityHash = ComputeServiceMethodSecurityPolicyHash(policy)
	return policy, policy.Validate()
}

func ValidateServiceSecurityModel(descriptor ServiceDescriptor, consensusCritical bool) error {
	_, err := NewServiceSecurityPolicy(descriptor, consensusCritical)
	return err
}

func (policy ServiceSecurityPolicy) Validate() error {
	if err := validateInterfaceToken("services security policy service id", policy.ServiceID); err != nil {
		return err
	}
	if !coretypes.IsServiceType(policy.ServiceType) {
		return fmt.Errorf("services security policy unknown service type %q", policy.ServiceType)
	}
	if !coretypes.IsServiceTrustModel(policy.TrustModel) {
		return fmt.Errorf("services security policy unknown trust model %q", policy.TrustModel)
	}
	if policy.TrustModelLabel != MustServiceTrustModelLabel(policy.TrustModel) {
		return errors.New("services security policy trust label mismatch")
	}
	if !coretypes.IsServiceVerificationModel(policy.VerificationModel) {
		return fmt.Errorf("services security policy unknown verification model %q", policy.VerificationModel)
	}
	if policy.ProofFormat != "" {
		if err := validateInterfaceToken("services security policy proof format", policy.ProofFormat); err != nil {
			return err
		}
	}
	if policy.ProviderCollateralDenom != "" {
		if err := validateInterfaceToken("services security policy collateral denom", policy.ProviderCollateralDenom); err != nil {
			return err
		}
	}
	if policy.ProviderCollateralAmount != "" {
		if err := validateInterfaceToken("services security policy collateral amount", policy.ProviderCollateralAmount); err != nil {
			return err
		}
	}
	if policy.FallbackServiceID != "" {
		if err := validateInterfaceToken("services security policy fallback service id", policy.FallbackServiceID); err != nil {
			return err
		}
	}
	if policy.FaultPolicy != "" {
		if !coretypes.IsServiceFailureBehavior(policy.FaultPolicy) {
			return fmt.Errorf("services security policy unknown fault policy %q", policy.FaultPolicy)
		}
		if policy.FaultPolicyLabel != MustServiceFailureBehaviorLabel(policy.FaultPolicy) {
			return errors.New("services security policy fault label mismatch")
		}
	}
	if !coretypes.IsServiceFailureBehavior(policy.ExecutionFailureBehavior) {
		return fmt.Errorf("services security policy unknown execution failure behavior %q", policy.ExecutionFailureBehavior)
	}
	if policy.ExecutionFailureBehaviorLabel != MustServiceFailureBehaviorLabel(policy.ExecutionFailureBehavior) {
		return errors.New("services security policy execution failure label mismatch")
	}
	if len(policy.MethodPolicies) == 0 {
		return errors.New("services security policy requires method policies")
	}
	if err := validateServiceMethodSecurityPolicies(policy.MethodPolicies); err != nil {
		return err
	}
	if policy.ConsensusCriticalRequested && !policy.ConsensusCriticalAllowed {
		return errors.New("services security policy forbids consensus-critical use")
	}
	switch policy.TrustModel {
	case coretypes.ServiceTrustFullyTrusted:
		if policy.ConsensusCriticalRequested && !serviceVerificationIsIndependent(policy.VerificationModel) {
			return errors.New("services fully trusted consensus-critical service requires independent verification")
		}
	case coretypes.ServiceTrustEconomicallySecured:
		if policy.ProviderCollateralDenom == "" || policy.ProviderCollateralAmount == "" || policy.FaultPolicy != coretypes.ServiceFailureSlashProvider {
			return errors.New("services economically secured service requires collateral and slash-provider penalty")
		}
	case coretypes.ServiceTrustCryptographicallyVerifiable:
		if policy.ProofFormat == "" {
			return errors.New("services cryptographically verifiable service requires proof format")
		}
	case coretypes.ServiceTrustConsensusExecuted:
		if !policy.RequiresDeterministicGas || policy.VerificationModel != coretypes.ServiceVerificationConsensusReceipt {
			return errors.New("services consensus executed service requires deterministic consensus receipt verification")
		}
	case coretypes.ServiceTrustHybridChallengeable:
		if policy.ChallengeWindow == 0 || !policy.HasFallbackRule() {
			return errors.New("services hybrid challengeable service requires challenge period and fallback rule")
		}
	}
	if err := coretypes.ValidateHash("services security policy hash", policy.SecurityHash); err != nil {
		return err
	}
	if expected := ComputeServiceSecurityPolicyHash(policy); policy.SecurityHash != expected {
		return fmt.Errorf("services security policy hash mismatch: expected %s", expected)
	}
	return nil
}

func (policy ServiceSecurityPolicy) HasFallbackRule() bool {
	if policy.FallbackServiceID != "" || policy.ExecutionFailureBehavior == coretypes.ServiceFailureFallbackOnChain {
		return true
	}
	for _, method := range policy.MethodPolicies {
		if method.FailureBehavior == coretypes.ServiceFailureFallbackOnChain {
			return true
		}
	}
	return false
}

func (policy ServiceMethodSecurityPolicy) Validate() error {
	if err := validateInterfaceToken("services security method id", policy.MethodID); err != nil {
		return err
	}
	if !coretypes.IsServiceTrustModel(policy.TrustModel) {
		return fmt.Errorf("services security method unknown trust model %q", policy.TrustModel)
	}
	if policy.TrustModelLabel != MustServiceTrustModelLabel(policy.TrustModel) {
		return errors.New("services security method trust label mismatch")
	}
	if !coretypes.IsServiceVerificationModel(policy.VerificationModel) {
		return fmt.Errorf("services security method unknown verification model %q", policy.VerificationModel)
	}
	if !coretypes.IsServiceFailureBehavior(policy.FailureBehavior) {
		return fmt.Errorf("services security method unknown failure behavior %q", policy.FailureBehavior)
	}
	if policy.FailureBehaviorLabel != MustServiceFailureBehaviorLabel(policy.FailureBehavior) {
		return errors.New("services security method failure label mismatch")
	}
	if err := coretypes.ValidateHash("services security method hash", policy.MethodSecurityHash); err != nil {
		return err
	}
	if expected := ComputeServiceMethodSecurityPolicyHash(policy); policy.MethodSecurityHash != expected {
		return fmt.Errorf("services security method hash mismatch: expected %s", expected)
	}
	return nil
}

func ParseServiceTrustModelLabel(value string) (coretypes.ServiceTrustModel, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(ServiceTrustLabelFullyTrusted), strings.ToLower(string(coretypes.ServiceTrustFullyTrusted)):
		return coretypes.ServiceTrustFullyTrusted, nil
	case string(ServiceTrustLabelEconomicallySecured), strings.ToLower(string(coretypes.ServiceTrustEconomicallySecured)):
		return coretypes.ServiceTrustEconomicallySecured, nil
	case string(ServiceTrustLabelCryptographicallyVerifiable), strings.ToLower(string(coretypes.ServiceTrustCryptographicallyVerifiable)):
		return coretypes.ServiceTrustCryptographicallyVerifiable, nil
	case string(ServiceTrustLabelConsensusExecuted), strings.ToLower(string(coretypes.ServiceTrustConsensusExecuted)):
		return coretypes.ServiceTrustConsensusExecuted, nil
	case string(ServiceTrustLabelHybridChallengeable), strings.ToLower(string(coretypes.ServiceTrustHybridChallengeable)):
		return coretypes.ServiceTrustHybridChallengeable, nil
	default:
		return "", fmt.Errorf("services unknown trust model label %q", value)
	}
}

func ParseServiceFailureBehaviorLabel(value string) (coretypes.ServiceFailureBehavior, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(ServiceFailureLabelRevert), strings.ToLower(string(coretypes.ServiceFailureRevert)):
		return coretypes.ServiceFailureRevert, nil
	case string(ServiceFailureLabelRetry), strings.ToLower(string(coretypes.ServiceFailureRetry)):
		return coretypes.ServiceFailureRetry, nil
	case string(ServiceFailureLabelFallbackOnChain), strings.ToLower(string(coretypes.ServiceFailureFallbackOnChain)):
		return coretypes.ServiceFailureFallbackOnChain, nil
	case string(ServiceFailureLabelChallenge), strings.ToLower(string(coretypes.ServiceFailureChallenge)):
		return coretypes.ServiceFailureChallenge, nil
	case string(ServiceFailureLabelSlashProvider), strings.ToLower(string(coretypes.ServiceFailureSlashProvider)):
		return coretypes.ServiceFailureSlashProvider, nil
	case string(ServiceFailureLabelRefund), strings.ToLower(string(coretypes.ServiceFailureRefund)):
		return coretypes.ServiceFailureRefund, nil
	case string(ServiceFailureLabelPartialSettle), strings.ToLower(string(coretypes.ServiceFailurePartialSettle)):
		return coretypes.ServiceFailurePartialSettle, nil
	default:
		return "", fmt.Errorf("services unknown failure behavior label %q", value)
	}
}

func MustServiceTrustModelLabel(model coretypes.ServiceTrustModel) ServiceTrustModelLabel {
	label, err := ServiceTrustModelSpecLabel(model)
	if err != nil {
		panic(err)
	}
	return label
}

func ServiceTrustModelSpecLabel(model coretypes.ServiceTrustModel) (ServiceTrustModelLabel, error) {
	switch model {
	case coretypes.ServiceTrustFullyTrusted:
		return ServiceTrustLabelFullyTrusted, nil
	case coretypes.ServiceTrustEconomicallySecured:
		return ServiceTrustLabelEconomicallySecured, nil
	case coretypes.ServiceTrustCryptographicallyVerifiable:
		return ServiceTrustLabelCryptographicallyVerifiable, nil
	case coretypes.ServiceTrustConsensusExecuted:
		return ServiceTrustLabelConsensusExecuted, nil
	case coretypes.ServiceTrustHybridChallengeable:
		return ServiceTrustLabelHybridChallengeable, nil
	default:
		return "", fmt.Errorf("services unknown trust model %q", model)
	}
}

func MustServiceFailureBehaviorLabel(behavior coretypes.ServiceFailureBehavior) ServiceFailureBehaviorLabel {
	label, err := ServiceFailureBehaviorSpecLabel(behavior)
	if err != nil {
		panic(err)
	}
	return label
}

func ServiceFailureBehaviorSpecLabel(behavior coretypes.ServiceFailureBehavior) (ServiceFailureBehaviorLabel, error) {
	switch behavior {
	case coretypes.ServiceFailureRevert:
		return ServiceFailureLabelRevert, nil
	case coretypes.ServiceFailureRetry:
		return ServiceFailureLabelRetry, nil
	case coretypes.ServiceFailureFallbackOnChain:
		return ServiceFailureLabelFallbackOnChain, nil
	case coretypes.ServiceFailureChallenge:
		return ServiceFailureLabelChallenge, nil
	case coretypes.ServiceFailureSlashProvider:
		return ServiceFailureLabelSlashProvider, nil
	case coretypes.ServiceFailureRefund:
		return ServiceFailureLabelRefund, nil
	case coretypes.ServiceFailurePartialSettle:
		return ServiceFailureLabelPartialSettle, nil
	default:
		return "", fmt.Errorf("services unknown failure behavior %q", behavior)
	}
}

func ComputeServiceSecurityPolicyHash(policy ServiceSecurityPolicy) string {
	methods := append([]ServiceMethodSecurityPolicy(nil), policy.MethodPolicies...)
	sortServiceMethodSecurityPolicies(methods)
	parts := []string{
		"aetra-services-security-policy-v1",
		policy.ServiceID,
		string(policy.ServiceType),
		string(policy.TrustModel),
		string(policy.TrustModelLabel),
		string(policy.VerificationModel),
		policy.ProofFormat,
		policy.ProviderCollateralDenom,
		policy.ProviderCollateralAmount,
		fmt.Sprint(policy.ChallengeWindow),
		policy.FallbackServiceID,
		string(policy.FaultPolicy),
		string(policy.FaultPolicyLabel),
		string(policy.ExecutionFailureBehavior),
		string(policy.ExecutionFailureBehaviorLabel),
		fmt.Sprint(policy.ConsensusCriticalRequested),
		fmt.Sprint(policy.ConsensusCriticalAllowed),
		fmt.Sprint(policy.RequiresIndependentVerification),
		fmt.Sprint(policy.RequiresCollateralPenalty),
		fmt.Sprint(policy.RequiresProofFormat),
		fmt.Sprint(policy.RequiresDeterministicGas),
		fmt.Sprint(policy.RequiresChallengeFallback),
		fmt.Sprint(len(methods)),
	}
	for _, method := range methods {
		parts = append(parts, method.MethodSecurityHash)
	}
	return servicesHashParts(parts...)
}

func ComputeServiceMethodSecurityPolicyHash(policy ServiceMethodSecurityPolicy) string {
	return servicesHashParts(
		"aetra-services-method-security-policy-v1",
		policy.MethodID,
		string(policy.TrustModel),
		string(policy.TrustModelLabel),
		string(policy.VerificationModel),
		string(policy.FailureBehavior),
		string(policy.FailureBehaviorLabel),
		fmt.Sprint(policy.ConsensusCriticalAllowed),
	)
}

func serviceFailureBehaviorLabelOrEmpty(behavior coretypes.ServiceFailureBehavior) ServiceFailureBehaviorLabel {
	if behavior == "" {
		return ""
	}
	return MustServiceFailureBehaviorLabel(behavior)
}

func serviceConsensusCriticalAllowed(descriptor ServiceDescriptor, consensusCritical bool) bool {
	if !consensusCritical {
		return true
	}
	if descriptor.Verification.TrustModel == coretypes.ServiceTrustFullyTrusted && !serviceVerificationIsIndependent(descriptor.Verification.Model) {
		return false
	}
	return true
}

func serviceVerificationIsIndependent(model coretypes.ServiceVerificationModel) bool {
	switch model {
	case coretypes.ServiceVerificationConsensusReceipt,
		coretypes.ServiceVerificationProofAnchored,
		coretypes.ServiceVerificationChallengeWindow,
		coretypes.ServiceVerificationEconomicCollateral:
		return true
	default:
		return false
	}
}

func validateServiceMethodSecurityPolicies(policies []ServiceMethodSecurityPolicy) error {
	previous := ""
	for _, policy := range policies {
		if err := policy.Validate(); err != nil {
			return err
		}
		if previous != "" && previous >= policy.MethodID {
			return errors.New("services security method policies must be sorted canonically")
		}
		previous = policy.MethodID
	}
	return nil
}

func sortServiceMethodSecurityPolicies(policies []ServiceMethodSecurityPolicy) {
	sort.SliceStable(policies, func(i, j int) bool { return policies[i].MethodID < policies[j].MethodID })
}

func maxUint64(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}
