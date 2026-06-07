package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

func TestServiceSecurityLabelsParseSpecValues(t *testing.T) {
	trust, err := ParseServiceTrustModelLabel("fully_trusted")
	require.NoError(t, err)
	require.Equal(t, coretypes.ServiceTrustFullyTrusted, trust)

	trust, err = ParseServiceTrustModelLabel("CONSENSUS_EXECUTED")
	require.NoError(t, err)
	require.Equal(t, coretypes.ServiceTrustConsensusExecuted, trust)

	failure, err := ParseServiceFailureBehaviorLabel("partial_settle")
	require.NoError(t, err)
	require.Equal(t, coretypes.ServiceFailurePartialSettle, failure)

	label, err := ServiceTrustModelSpecLabel(coretypes.ServiceTrustHybridChallengeable)
	require.NoError(t, err)
	require.Equal(t, ServiceTrustLabelHybridChallengeable, label)

	failureLabel, err := ServiceFailureBehaviorSpecLabel(coretypes.ServiceFailureFallbackOnChain)
	require.NoError(t, err)
	require.Equal(t, ServiceFailureLabelFallbackOnChain, failureLabel)
}

func TestServiceSecurityPolicyAllowsConsensusExecutedDeterministicGasService(t *testing.T) {
	descriptor := testUnifiedOnChainDescriptor()
	policy, err := NewServiceSecurityPolicy(descriptor, true)
	require.NoError(t, err)
	require.Equal(t, ServiceTrustLabelConsensusExecuted, policy.TrustModelLabel)
	require.True(t, policy.ConsensusCriticalAllowed)
	require.True(t, policy.RequiresDeterministicGas)
	require.Len(t, policy.MethodPolicies, 1)
	require.Equal(t, ServiceFailureLabelRevert, policy.ExecutionFailureBehaviorLabel)
	require.Equal(t, ServiceFailureLabelRetry, policy.MethodPolicies[0].FailureBehaviorLabel)
	require.NoError(t, policy.Validate())
}

func TestServiceSecurityPolicyRejectsFullyTrustedConsensusCriticalWithoutIndependentVerification(t *testing.T) {
	descriptor := testUnifiedOffChainDescriptor()
	policy, err := NewServiceSecurityPolicy(descriptor, false)
	require.NoError(t, err)
	require.True(t, policy.RequiresIndependentVerification)
	require.True(t, policy.ConsensusCriticalAllowed)

	_, err = NewServiceSecurityPolicy(descriptor, true)
	require.ErrorContains(t, err, "consensus-critical")
}

func TestServiceSecurityPolicyRequiresEconomicPenaltyRules(t *testing.T) {
	descriptor := testUnifiedOffChainDescriptor()
	descriptor.Verification.TrustModel = coretypes.ServiceTrustEconomicallySecured
	descriptor.Verification.Model = coretypes.ServiceVerificationEconomicCollateral
	descriptor.Verification.ProviderCollateralDenom = coretypes.NativeFeePolicyID
	descriptor.Verification.ProviderCollateralAmount = "100"
	descriptor.Verification.FaultPolicy = coretypes.ServiceFailureRefund

	_, err := NewServiceSecurityPolicy(descriptor, false)
	require.ErrorContains(t, err, "slash-provider")

	descriptor.Verification.FaultPolicy = coretypes.ServiceFailureSlashProvider
	policy, err := NewServiceSecurityPolicy(descriptor, false)
	require.NoError(t, err)
	require.True(t, policy.RequiresCollateralPenalty)
	require.Equal(t, ServiceFailureLabelSlashProvider, policy.FaultPolicyLabel)
	require.NoError(t, policy.Validate())
}

func TestServiceSecurityPolicyRequiresCryptographicProofFormat(t *testing.T) {
	descriptor := testUnifiedOffChainDescriptor()
	descriptor.Verification.TrustModel = coretypes.ServiceTrustCryptographicallyVerifiable
	descriptor.Verification.Model = coretypes.ServiceVerificationProofAnchored

	_, err := NewServiceSecurityPolicy(descriptor, false)
	require.ErrorContains(t, err, "proof format")

	descriptor.Verification.ProofFormat = "groth16-v1"
	policy, err := NewServiceSecurityPolicy(descriptor, false)
	require.NoError(t, err)
	require.True(t, policy.RequiresProofFormat)
	require.Equal(t, "groth16-v1", policy.ProofFormat)
}

func TestServiceSecurityPolicyRequiresHybridChallengeFallbackRule(t *testing.T) {
	descriptor := testInterfaceSystemDescriptor()
	_, err := NewServiceSecurityPolicy(descriptor, false)
	require.ErrorContains(t, err, "fallback rule")

	descriptor.Verification.FallbackServiceID = "identity-unified"
	policy, err := NewServiceSecurityPolicy(descriptor, false)
	require.NoError(t, err)
	require.True(t, policy.RequiresChallengeFallback)
	require.True(t, policy.HasFallbackRule())
	require.Equal(t, ServiceTrustLabelHybridChallengeable, policy.TrustModelLabel)
}

func TestServiceSecurityPolicyRequiresMethodFailureDeclaration(t *testing.T) {
	descriptor := testUnifiedOffChainDescriptor()
	descriptor.Interface.Methods[0].FailureBehavior = ""
	descriptor.Interface.InterfaceHash = coretypes.ComputeServiceInterfaceHash(descriptor.Interface)

	_, err := NewServiceSecurityPolicy(descriptor, false)
	require.ErrorContains(t, err, "failure behavior")
}
