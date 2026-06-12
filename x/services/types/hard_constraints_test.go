package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultServiceHardConstraintsManifestCoversSection19(t *testing.T) {
	manifest, err := DefaultServiceHardConstraintsManifest()
	require.NoError(t, err)
	require.NoError(t, manifest.Validate())
	require.Len(t, manifest.Constraints, 9)

	required := map[ServiceHardConstraintID]bool{
		ServiceConstraintNoCentralizedBackendAssumption:	false,
		ServiceConstraintNoMessagingApplicationDependency:	false,
		ServiceConstraintNoMonolithicExecutionEngine:		false,
		ServiceConstraintNoManualABIForRegisteredServices:	false,
		ServiceConstraintNoExternalAPIInConsensusExecution:	false,
		ServiceConstraintNoNondeterministicStateTransition:	false,
		ServiceConstraintNoUnboundedRegistryScans:		false,
		ServiceConstraintNoUnmeteredProofVerification:		false,
		ServiceConstraintNoUnverifiedOffChainCanonicalState:	false,
	}
	consensusCritical := map[ServiceHardConstraintID]bool{
		ServiceConstraintNoExternalAPIInConsensusExecution:	true,
		ServiceConstraintNoNondeterministicStateTransition:	true,
		ServiceConstraintNoUnboundedRegistryScans:		true,
		ServiceConstraintNoUnmeteredProofVerification:		true,
		ServiceConstraintNoUnverifiedOffChainCanonicalState:	true,
	}
	for _, constraint := range manifest.Constraints {
		_, found := required[constraint.ConstraintID]
		require.Truef(t, found, "unexpected constraint %s", constraint.ConstraintID)
		required[constraint.ConstraintID] = true
		require.NotEmpty(t, constraint.Stages)
		require.NotEmpty(t, constraint.RequiredControls)
		require.Equal(t, consensusCritical[constraint.ConstraintID], constraint.ConsensusCritical)
		require.Equal(t, ComputeServiceHardConstraintHash(constraint), constraint.ConstraintHash)
	}
	for constraintID, found := range required {
		require.Truef(t, found, "missing constraint %s", constraintID)
	}
	require.Equal(t, ComputeServiceHardConstraintsManifestHash(manifest), manifest.ManifestHash)
}

func TestServiceHardConstraintsManifestRejectsMissingAndDuplicateConstraints(t *testing.T) {
	manifest, err := DefaultServiceHardConstraintsManifest()
	require.NoError(t, err)

	_, err = NewServiceHardConstraintsManifest(manifest.Constraints[1:])
	require.ErrorContains(t, err, "must include 9 constraints")

	constraints := append([]ServiceHardConstraint(nil), manifest.Constraints...)
	constraints[len(constraints)-1] = constraints[0]
	_, err = NewServiceHardConstraintsManifest(constraints)
	require.ErrorContains(t, err, "duplicate services hard constraint")
}

func TestServiceHardConstraintsManifestRejectsHashTampering(t *testing.T) {
	manifest, err := DefaultServiceHardConstraintsManifest()
	require.NoError(t, err)

	tamperedConstraint := manifest
	tamperedConstraint.Constraints = append([]ServiceHardConstraint(nil), manifest.Constraints...)
	tamperedConstraint.Constraints[0].RequiredControls = append([]string(nil), manifest.Constraints[0].RequiredControls...)
	tamperedConstraint.Constraints[0].RequiredControls[0] = "tampered_control"
	require.ErrorContains(t, tamperedConstraint.Validate(), "hash mismatch")

	tamperedManifest := manifest
	tamperedManifest.ManifestHash = testDistributedHash("tampered-hard-constraints-manifest")
	require.ErrorContains(t, tamperedManifest.Validate(), "manifest hash mismatch")
}

func TestServiceExecutionHardConstraintPolicyAcceptsSafePath(t *testing.T) {
	policy := ServiceExecutionHardConstraintPolicy{
		OffChainResultCanonical:		true,
		OffChainResultHasChallengeWindow:	true,
		OffChainResultHasExplicitTrustModel:	true,
	}
	require.NoError(t, policy.ValidateAgainstHardConstraints())
	require.True(t, policy.HasOffChainCanonicalizationGate())
}

func TestServiceExecutionHardConstraintPolicyRejectsForbiddenPaths(t *testing.T) {
	cases := []struct {
		name		string
		policy		ServiceExecutionHardConstraintPolicy
		constraint	ServiceHardConstraintID
	}{
		{
			name:		"centralized backend",
			policy:		ServiceExecutionHardConstraintPolicy{CentralizedBackendRequired: true},
			constraint:	ServiceConstraintNoCentralizedBackendAssumption,
		},
		{
			name:		"messaging app dependency",
			policy:		ServiceExecutionHardConstraintPolicy{MessagingApplicationRequired: true},
			constraint:	ServiceConstraintNoMessagingApplicationDependency,
		},
		{
			name:		"monolithic engine",
			policy:		ServiceExecutionHardConstraintPolicy{MonolithicExecutionEngine: true},
			constraint:	ServiceConstraintNoMonolithicExecutionEngine,
		},
		{
			name:		"manual abi",
			policy:		ServiceExecutionHardConstraintPolicy{ManualABIRequiredForRegisteredService: true},
			constraint:	ServiceConstraintNoManualABIForRegisteredServices,
		},
		{
			name:		"external api",
			policy:		ServiceExecutionHardConstraintPolicy{ExternalAPIRequiredInConsensus: true},
			constraint:	ServiceConstraintNoExternalAPIInConsensusExecution,
		},
		{
			name:		"nondeterministic transition",
			policy:		ServiceExecutionHardConstraintPolicy{NondeterministicStateTransition: true},
			constraint:	ServiceConstraintNoNondeterministicStateTransition,
		},
		{
			name:		"unbounded registry scan",
			policy:		ServiceExecutionHardConstraintPolicy{UnboundedRegistryScan: true},
			constraint:	ServiceConstraintNoUnboundedRegistryScans,
		},
		{
			name:		"unmetered proof verification",
			policy:		ServiceExecutionHardConstraintPolicy{UnmeteredProofVerification: true},
			constraint:	ServiceConstraintNoUnmeteredProofVerification,
		},
		{
			name:		"unverified off-chain canonical result",
			policy:		ServiceExecutionHardConstraintPolicy{OffChainResultCanonical: true},
			constraint:	ServiceConstraintNoUnverifiedOffChainCanonicalState,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.policy.ValidateAgainstHardConstraints()
			require.ErrorContains(t, err, string(tc.constraint))
		})
	}
}

func TestServiceHardConstraintClassifiers(t *testing.T) {
	require.True(t, IsServiceHardConstraintID(ServiceConstraintNoExternalAPIInConsensusExecution))
	require.False(t, IsServiceHardConstraintID(ServiceHardConstraintID("unknown")))
	require.True(t, IsServiceHardConstraintCategory(ServiceConstraintCategoryConsensus))
	require.False(t, IsServiceHardConstraintCategory(ServiceHardConstraintCategory("unknown")))
	require.True(t, IsServiceHardConstraintStage(ServiceConstraintStageFinalizeBlock))
	require.False(t, IsServiceHardConstraintStage(ServiceHardConstraintStage("unknown")))
}
