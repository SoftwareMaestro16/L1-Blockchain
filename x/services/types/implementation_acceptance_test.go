package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultServiceImplementationAcceptanceManifestCoversSection20(t *testing.T) {
	manifest, err := DefaultServiceImplementationAcceptanceManifest()
	require.NoError(t, err)
	require.NoError(t, manifest.Validate())
	require.Len(t, manifest.Criteria, 12)

	required := map[ServiceImplementationAcceptanceID]bool{
		ServiceAcceptanceFirstClassRegistryObjects:	false,
		ServiceAcceptanceAllServiceTypesDefined:	false,
		ServiceAcceptanceFormalInterfaceBinding:	false,
		ServiceAcceptanceUnifiedCallEnvelope:		false,
		ServiceAcceptanceDiscoveryResolution:		false,
		ServiceAcceptancePaymentSettlementModes:	false,
		ServiceAcceptanceStorageDeclarations:		false,
		ServiceAcceptanceMixedServiceDisputes:		false,
		ServiceAcceptanceProviderRules:			false,
		ServiceAcceptanceCosmosModuleSurface:		false,
		ServiceAcceptanceStoreV2BlockSTMStrategy:	false,
		ServiceAcceptanceSDKExecutionFlow:		false,
	}
	for _, criterion := range manifest.Criteria {
		_, found := required[criterion.CriterionID]
		require.Truef(t, found, "unexpected criterion %s", criterion.CriterionID)
		required[criterion.CriterionID] = true
		require.NotEmpty(t, criterion.RequiredEvidence)
		require.NotEmpty(t, criterion.RequiredObjects)
		require.Equal(t, ComputeServiceImplementationAcceptanceCriterionHash(criterion), criterion.CriterionHash)
	}
	for criterionID, found := range required {
		require.Truef(t, found, "missing criterion %s", criterionID)
	}
	require.Equal(t, ComputeServiceImplementationAcceptanceManifestHash(manifest), manifest.ManifestHash)
}

func TestServiceImplementationAcceptanceManifestMapsRequiredObjects(t *testing.T) {
	manifest, err := DefaultServiceImplementationAcceptanceManifest()
	require.NoError(t, err)

	requireAcceptanceObjects(t, manifest, ServiceAcceptanceAllServiceTypesDefined, "fog_market", "mixed", "off_chain", "on_chain")
	requireAcceptanceObjects(t, manifest, ServiceAcceptanceUnifiedCallEnvelope, "PaymentEnvelope", "ServiceCall", "idempotency_key", "payload", "proof_requirement", "signature", "timeout")
	requireAcceptanceObjects(t, manifest, ServiceAcceptanceDiscoveryResolution, "identity_binding", "mesh_record", "registry", "signed_cache")
	requireAcceptanceObjects(t, manifest, ServiceAcceptancePaymentSettlementModes, "escrow", "metered", "on_chain", "prepaid", "streaming")
	requireAcceptanceObjects(t, manifest, ServiceAcceptanceStorageDeclarations, "distributed_off_chain", "ephemeral", "hybrid", "persistent_on_chain")
	requireAcceptanceObjects(t, manifest, ServiceAcceptanceCosmosModuleSurface, "Events", "Genesis", "Invariants", "Keeper", "MsgServer", "QueryServer", "RootCommitments", "TypedErrors")
	requireAcceptanceObjects(t, manifest, ServiceAcceptanceSDKExecutionFlow, "AttachPayment", "BuildCall", "ExecuteCall", "ResolveService", "VerifyInterface", "VerifyReceipt")
}

func TestServiceImplementationAcceptanceManifestRejectsMissingAndDuplicateCriteria(t *testing.T) {
	manifest, err := DefaultServiceImplementationAcceptanceManifest()
	require.NoError(t, err)

	_, err = NewServiceImplementationAcceptanceManifest(manifest.Criteria[1:])
	require.ErrorContains(t, err, "must include 12 criteria")

	criteria := append([]ServiceImplementationAcceptanceCriterion(nil), manifest.Criteria...)
	criteria[len(criteria)-1] = criteria[0]
	_, err = NewServiceImplementationAcceptanceManifest(criteria)
	require.ErrorContains(t, err, "duplicate services implementation acceptance criterion")
}

func TestServiceImplementationAcceptanceManifestRejectsHashTampering(t *testing.T) {
	manifest, err := DefaultServiceImplementationAcceptanceManifest()
	require.NoError(t, err)

	tamperedCriterion := manifest
	tamperedCriterion.Criteria = append([]ServiceImplementationAcceptanceCriterion(nil), manifest.Criteria...)
	tamperedCriterion.Criteria[0].RequiredObjects = append([]string(nil), manifest.Criteria[0].RequiredObjects...)
	tamperedCriterion.Criteria[0].RequiredObjects[0] = "tampered_object"
	require.ErrorContains(t, tamperedCriterion.Validate(), "hash mismatch")

	tamperedManifest := manifest
	tamperedManifest.ManifestHash = testDistributedHash("tampered-implementation-acceptance-manifest")
	require.ErrorContains(t, tamperedManifest.Validate(), "manifest hash mismatch")
}

func TestServiceImplementationPlanningGateRequiresAllCriteria(t *testing.T) {
	manifest, err := DefaultServiceImplementationAcceptanceManifest()
	require.NoError(t, err)

	require.NoError(t, ReadyServiceImplementationPlanningGate().ValidateReady(manifest))

	missingOne := ServiceImplementationPlanningGate{
		SatisfiedCriteria: requiredServiceImplementationAcceptanceIDs()[1:],
	}
	require.ErrorContains(t, missingOne.ValidateReady(manifest), "missing criterion")

	duplicate := ReadyServiceImplementationPlanningGate()
	duplicate.SatisfiedCriteria = append(duplicate.SatisfiedCriteria, duplicate.SatisfiedCriteria[0])
	require.ErrorContains(t, duplicate.ValidateReady(manifest), "duplicate services implementation acceptance gate criterion")
}

func TestServiceImplementationAcceptanceClassifiers(t *testing.T) {
	require.True(t, IsServiceImplementationAcceptanceID(ServiceAcceptanceSDKExecutionFlow))
	require.False(t, IsServiceImplementationAcceptanceID(ServiceImplementationAcceptanceID("unknown")))
	require.True(t, IsServiceImplementationAcceptanceCategory(ServiceAcceptanceCategoryCosmos))
	require.False(t, IsServiceImplementationAcceptanceCategory(ServiceImplementationAcceptanceCategory("unknown")))
	require.True(t, IsServiceImplementationAcceptanceEvidence(ServiceAcceptanceEvidenceIntegration))
	require.False(t, IsServiceImplementationAcceptanceEvidence(ServiceImplementationAcceptanceEvidence("unknown")))
}

func requireAcceptanceObjects(t *testing.T, manifest ServiceImplementationAcceptanceManifest, criterionID ServiceImplementationAcceptanceID, objects ...string) {
	t.Helper()
	var found ServiceImplementationAcceptanceCriterion
	for _, criterion := range manifest.Criteria {
		if criterion.CriterionID == criterionID {
			found = criterion
			break
		}
	}
	require.NotEmpty(t, found.CriterionID)
	objectSet := map[string]struct{}{}
	for _, object := range found.RequiredObjects {
		objectSet[object] = struct{}{}
	}
	for _, object := range objects {
		_, ok := objectSet[object]
		require.Truef(t, ok, "criterion %s missing object %s", criterionID, object)
	}
}
