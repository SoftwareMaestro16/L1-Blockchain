package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAcceptanceCriteriaManifestCoversSection19(t *testing.T) {
	manifest, err := DefaultAcceptanceCriteriaManifest()
	require.NoError(t, err)
	require.NoError(t, manifest.Validate())
	require.Len(t, manifest.Criteria, 11)
	require.Equal(t, ComputeAcceptanceCriteriaManifestHash(manifest), manifest.ManifestHash)

	for _, id := range RequiredAcceptanceCriterionIDs() {
		criterion, found := AcceptanceCriterionByID(manifest, id)
		require.True(t, found, id)
		require.True(t, criterion.PlanningReady, id)
		require.True(t, IsImplementationRoadmapPhaseID(criterion.PhaseID), id)
		require.NotEmpty(t, criterion.Modules, id)
		require.NotEmpty(t, criterion.RootTypes, id)
		require.NotEmpty(t, criterion.Evidence, id)
		require.Equal(t, ComputeAcceptanceCriterionHash(criterion), criterion.CriterionHash)
		for _, moduleName := range criterion.Modules {
			require.True(t, IsRequiredCosmosSDKModule(moduleName), id)
		}
	}

	defaultZone, found := AcceptanceCriterionByID(manifest, AcceptanceAEKCoordinatesDefaultZone)
	require.True(t, found)
	require.Equal(t, RoadmapPhaseKernelRootModel, defaultZone.PhaseID)
	require.Equal(t, []CosmosSDKModuleName{CosmosModuleAetraCore, CosmosModuleZones}, defaultZone.Modules)
	require.Contains(t, defaultZone.Evidence, "AEK can register and coordinate a default execution zone")

	canonicalZones, found := AcceptanceCriterionByID(manifest, AcceptanceFourCanonicalZonesSpecified)
	require.True(t, found)
	require.Equal(t, RoadmapPhaseCanonicalZones, canonicalZones.PhaseID)
	require.Equal(t, []RootType{RootType("application"), RootType("contracts"), RootType("financial"), RootType("identity")}, canonicalZones.RootTypes)

	messaging, found := AcceptanceCriterionByID(manifest, AcceptanceCrossZoneMessagingSemantics)
	require.True(t, found)
	require.Equal(t, []RootType{MessageProofRootType, ReceiptProofRootType}, messaging.RootTypes)
	require.Contains(t, messaging.Evidence, "replay protection bounce handling and receipts are mandatory")

	globalRoot, found := AcceptanceCriterionByID(manifest, AcceptanceGlobalRootExposesAllRoots)
	require.True(t, found)
	require.Equal(t, []RootType{RootType("contracts"), RootType("identity"), MessageProofRootType, RootType("payments"), ReceiptProofRootType, RoutingTableRootType, RootType("services"), RootType("storage"), RootType("zones")}, globalRoot.RootTypes)

	modules, found := AcceptanceCriterionByID(manifest, AcceptanceModulesExportImportInvariantsTests)
	require.True(t, found)
	require.Equal(t, RequiredCosmosSDKModules(), modules.Modules)
	require.Contains(t, modules.Evidence, "all modules expose typed query interfaces")
}

func TestAcceptanceCriteriaManifestRejectsMissingDuplicateAndMalformedCriteria(t *testing.T) {
	manifest, err := DefaultAcceptanceCriteriaManifest()
	require.NoError(t, err)

	missing := manifest
	missing.Criteria = append([]AcceptanceCriterion(nil), manifest.Criteria[1:]...)
	missing.ManifestHash = ComputeAcceptanceCriteriaManifestHash(missing)
	require.ErrorContains(t, missing.Validate(), "must include 11 required criteria")

	duplicate := manifest
	duplicate.Criteria = append([]AcceptanceCriterion(nil), manifest.Criteria...)
	duplicate.Criteria[len(duplicate.Criteria)-1] = duplicate.Criteria[0]
	duplicate.ManifestHash = ComputeAcceptanceCriteriaManifestHash(duplicate)
	require.ErrorContains(t, duplicate.Validate(), "duplicate")

	unknownModule := manifest
	unknownModule.Criteria = append([]AcceptanceCriterion(nil), manifest.Criteria...)
	unknownModule.Criteria[0].Modules = []CosmosSDKModuleName{CosmosSDKModuleName("dex")}
	unknownModule.Criteria[0].CriterionHash = ComputeAcceptanceCriterionHash(unknownModule.Criteria[0])
	unknownModule.ManifestHash = ComputeAcceptanceCriteriaManifestHash(unknownModule)
	require.ErrorContains(t, unknownModule.Validate(), "unknown module")

	unknownPhase := manifest
	unknownPhase.Criteria = append([]AcceptanceCriterion(nil), manifest.Criteria...)
	unknownPhase.Criteria[0].PhaseID = ImplementationRoadmapPhaseID("phase-99")
	unknownPhase.Criteria[0].CriterionHash = ComputeAcceptanceCriterionHash(unknownPhase.Criteria[0])
	unknownPhase.ManifestHash = ComputeAcceptanceCriteriaManifestHash(unknownPhase)
	require.ErrorContains(t, unknownPhase.Validate(), "unknown roadmap phase")

	notReady := manifest
	notReady.Criteria = append([]AcceptanceCriterion(nil), manifest.Criteria...)
	notReady.Criteria[0].PlanningReady = false
	notReady.Criteria[0].CriterionHash = ComputeAcceptanceCriterionHash(notReady.Criteria[0])
	notReady.ManifestHash = ComputeAcceptanceCriteriaManifestHash(notReady)
	require.ErrorContains(t, notReady.Validate(), "planning ready")

	noEvidence := manifest
	noEvidence.Criteria = append([]AcceptanceCriterion(nil), manifest.Criteria...)
	noEvidence.Criteria[0].Evidence = nil
	noEvidence.Criteria[0].CriterionHash = ComputeAcceptanceCriterionHash(noEvidence.Criteria[0])
	noEvidence.ManifestHash = ComputeAcceptanceCriteriaManifestHash(noEvidence)
	require.ErrorContains(t, noEvidence.Validate(), "evidence is required")

	hashMismatch := manifest
	hashMismatch.Criteria = append([]AcceptanceCriterion(nil), manifest.Criteria...)
	hashMismatch.Criteria[0].Evidence = append([]string(nil), hashMismatch.Criteria[0].Evidence...)
	hashMismatch.Criteria[0].Evidence[0] = "tampered evidence"
	hashMismatch.ManifestHash = ComputeAcceptanceCriteriaManifestHash(hashMismatch)
	require.ErrorContains(t, hashMismatch.Validate(), "criterion hash mismatch")
}

func TestAcceptanceCriteriaManifestHashIsCanonical(t *testing.T) {
	manifest, err := DefaultAcceptanceCriteriaManifest()
	require.NoError(t, err)

	reversed := reverseAcceptanceCriteria(manifest.Criteria)
	for i := range reversed {
		reversed[i].Modules = reverseModules(reversed[i].Modules)
		reversed[i].RootTypes = reverseRootTypes(reversed[i].RootTypes)
		reversed[i].Evidence = reverseStrings(reversed[i].Evidence)
		reversed[i].CriterionHash = ""
	}
	reordered, err := NewAcceptanceCriteriaManifest(reversed)
	require.NoError(t, err)
	require.Equal(t, manifest.ManifestHash, reordered.ManifestHash)
	require.Equal(t, manifest.Criteria, reordered.Criteria)

	tampered := manifest
	tampered.Criteria = append([]AcceptanceCriterion(nil), manifest.Criteria...)
	tampered.Criteria[0].Evidence = append([]string(nil), tampered.Criteria[0].Evidence...)
	tampered.Criteria[0].Evidence[0] = "different evidence"
	tampered.Criteria[0].CriterionHash = ComputeAcceptanceCriterionHash(tampered.Criteria[0])
	tampered.ManifestHash = ComputeAcceptanceCriteriaManifestHash(tampered)
	require.NoError(t, tampered.Validate())
	require.NotEqual(t, manifest.ManifestHash, tampered.ManifestHash)
}

func reverseAcceptanceCriteria(criteria []AcceptanceCriterion) []AcceptanceCriterion {
	out := make([]AcceptanceCriterion, len(criteria))
	for i := range criteria {
		out[i] = criteria[len(criteria)-1-i]
	}
	return out
}

func reverseRootTypes(rootTypes []RootType) []RootType {
	out := make([]RootType, len(rootTypes))
	for i := range rootTypes {
		out[i] = rootTypes[len(rootTypes)-1-i]
	}
	return out
}
