package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultServiceImplementationRoadmapCoversPhases0To2(t *testing.T) {
	roadmap, err := DefaultServiceImplementationRoadmap()
	require.NoError(t, err)
	require.NoError(t, roadmap.Validate())
	require.NotEmpty(t, roadmap.RoadmapHash)

	phase0, found := roadmap.PhaseByID(ServiceRoadmapPhaseSpecificationCompatibility)
	require.True(t, found)
	require.Contains(t, phase0.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskFinalizeDescriptorSchema, ServiceModuleServices, "ServiceDescriptor"))
	require.Contains(t, phase0.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskFinalizeInterfaceSchema, ServiceModuleInterface, "ServiceInterface"))
	require.Contains(t, phase0.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskDefineCallEnvelope, ServiceModuleCalls, "UnifiedServiceCall"))
	require.Contains(t, phase0.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskDefineReceiptFormat, ServiceModuleReceipts, "ServiceReceipt"))
	require.Contains(t, phase0.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskDefinePaymentModelEnum, ServiceModulePayments, "ServicePaymentSettlementMode"))
	require.Contains(t, phase0.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskDefineTrustVerificationEnums, ServiceModuleServices, "ServiceTrustModel/ServiceVerificationModel"))
	require.Contains(t, phase0.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskMapExistingModules, ServiceModuleServices, "AetherisModuleServiceMapping"))
	require.NoError(t, ValidateServiceRoadmapExitCriteria(phase0))

	phase1, found := roadmap.PhaseByID(ServiceRoadmapPhaseCoreRegistry)
	require.True(t, found)
	require.Contains(t, phase1.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskImplementServicesModule, ServiceModuleServices, "XServicesModuleBreakdown"))
	require.Contains(t, phase1.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskAddServiceRegistrationUpdate, ServiceModuleServices, "MsgRegisterService/MsgUpdateService"))
	require.Contains(t, phase1.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskAddNameOwnerIndexes, ServiceModuleServices, "ServiceRegistryStateEntry"))
	require.Contains(t, phase1.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskAddIdentityBindingPlaceholder, ServiceModuleServices, "IdentityServiceBinding"))
	require.Contains(t, phase1.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskAddProofQuery, ServiceModuleServices, "ServiceRegistryProof"))
	require.Contains(t, phase1.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskAddExportImport, ServiceModuleServices, "ServiceRegistryState"))
	require.NoError(t, roadmap.ReadyForPhase(ServiceRoadmapPhaseCoreRegistry))
	require.NoError(t, ValidateServiceRoadmapExitCriteria(phase1))

	phase2, found := roadmap.PhaseByID(ServiceRoadmapPhaseInterfaceSystem)
	require.True(t, found)
	require.Contains(t, phase2.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskImplementServiceInterfaceModule, ServiceModuleInterface, "XServiceInterfaceModuleBreakdown"))
	require.Contains(t, phase2.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskAddInterfaceRegistration, ServiceModuleInterface, "MsgRegisterInterface"))
	require.Contains(t, phase2.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskAddMethodSchema, ServiceModuleInterface, "ServiceInterfaceMethodSchema"))
	require.Contains(t, phase2.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskAddInterfaceHashValidation, ServiceModuleInterface, "ComputeFormalServiceInterfaceHash"))
	require.Contains(t, phase2.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskAddInterfaceProofQuery, ServiceModuleInterface, "QueryInterfaceProof"))
	require.Contains(t, phase2.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskAddSDKInterfaceVerifier, ServiceModuleInterface, "SDKInterfaceVerifier"))
	require.NoError(t, roadmap.ReadyForPhase(ServiceRoadmapPhaseInterfaceSystem))
	require.NoError(t, ValidateServiceRoadmapExitCriteria(phase2))
}

func TestServiceImplementationRoadmapRejectsMissingPhaseSurface(t *testing.T) {
	roadmap, err := DefaultServiceImplementationRoadmap()
	require.NoError(t, err)
	roadmap.Phases = roadmap.Phases[:2]
	roadmap.RoadmapHash = ComputeServiceImplementationRoadmapHash(roadmap)
	require.ErrorContains(t, roadmap.Validate(), "phases 0, 1, and 2")

	roadmap, err = DefaultServiceImplementationRoadmap()
	require.NoError(t, err)
	phase1, found := roadmap.PhaseByID(ServiceRoadmapPhaseCoreRegistry)
	require.True(t, found)
	phase1.Tasks = removeServiceRoadmapTaskForTest(phase1.Tasks, ServiceRoadmapTaskAddProofQuery)
	phase1, err = NewServiceRoadmapPhase(phase1)
	require.NoError(t, err)
	require.ErrorContains(t, ValidateServiceRoadmapExitCriteria(phase1), "missing task")
}

func TestServiceRoadmapPhase0ExitCriteriaEnforceCompatibilityArtifacts(t *testing.T) {
	roadmap, err := DefaultServiceImplementationRoadmap()
	require.NoError(t, err)
	phase0, found := roadmap.PhaseByID(ServiceRoadmapPhaseSpecificationCompatibility)
	require.True(t, found)

	require.NoError(t, validateServiceCoreObjectDefinitions(phase0.CoreObjects))
	require.NoError(t, validateServiceSignableObjectVectors(phase0.SignableVectors))
	require.NoError(t, ValidateAetherisModuleServiceMappings(phase0.ModuleMappings))

	phase0.CoreObjects = phase0.CoreObjects[1:]
	phase0, err = NewServiceRoadmapPhase(phase0)
	require.NoError(t, err)
	require.NoError(t, ValidateServiceRoadmapExitCriteria(phase0))

	mapping := phase0.ModuleMappings[0]
	mapping.OnChain = false
	mapping.MappingHash = ComputeAetherisModuleServiceMappingHash(mapping)
	phase0.ModuleMappings[0] = mapping
	_, err = NewServiceRoadmapPhase(phase0)
	require.ErrorContains(t, err, "on-chain services")
}

func TestServiceRoadmapReadyForPhaseRequiresMetDependencies(t *testing.T) {
	roadmap, err := DefaultServiceImplementationRoadmap()
	require.NoError(t, err)
	phase0, found := roadmap.PhaseByID(ServiceRoadmapPhaseSpecificationCompatibility)
	require.True(t, found)
	phase0.ExitCriteria[0].Met = false
	phase0.ExitCriteria[0].ExitHash = ComputeServiceRoadmapExitCriterionHash(phase0.ExitCriteria[0])
	phase0, err = NewServiceRoadmapPhase(phase0)
	require.NoError(t, err)
	for i := range roadmap.Phases {
		if roadmap.Phases[i].PhaseID == ServiceRoadmapPhaseSpecificationCompatibility {
			roadmap.Phases[i] = phase0
		}
	}
	roadmap, err = NewServiceImplementationRoadmap(roadmap.Phases)
	require.NoError(t, err)
	require.ErrorContains(t, roadmap.ReadyForPhase(ServiceRoadmapPhaseCoreRegistry), "not met")
}

func TestServiceRoadmapCanonicalHashesDetectTampering(t *testing.T) {
	task := newServiceRoadmapTask(ServiceRoadmapTaskDefineCallEnvelope, ServiceModuleCalls, "UnifiedServiceCall")
	task.Artifact = "OtherArtifact"
	require.ErrorContains(t, task.Validate(), "hash mismatch")

	vector := newServiceSignableObjectVector("ServiceDescriptor", "proto3_canonical_json", testInterfaceHash("roadmap/vector"))
	vector.CanonicalEncoding = "amino"
	require.ErrorContains(t, vector.Validate(), "hash mismatch")

	mapping := newAetherisModuleServiceMapping("dex", "x/dex", "dex-service", true)
	mapping.ServiceID = "other-service"
	require.ErrorContains(t, mapping.Validate(), "hash mismatch")
}

func removeServiceRoadmapTaskForTest(tasks []ServiceRoadmapTask, target ServiceRoadmapTaskID) []ServiceRoadmapTask {
	out := make([]ServiceRoadmapTask, 0, len(tasks))
	for _, task := range tasks {
		if task.TaskID != target {
			out = append(out, task)
		}
	}
	return out
}
