package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultServiceImplementationRoadmapCoversPhases0To8(t *testing.T) {
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
	require.Contains(t, phase0.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskMapExistingModules, ServiceModuleServices, "AetraModuleServiceMapping"))
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

	phase3, found := roadmap.PhaseByID(ServiceRoadmapPhaseUnifiedCallsReceipts)
	require.True(t, found)
	require.Contains(t, phase3.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskImplementServiceCallsModule, ServiceModuleCalls, "XServiceCallsModuleBreakdown"))
	require.Contains(t, phase3.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskImplementServiceReceiptsModule, ServiceModuleReceipts, "XServiceReceiptsModuleBreakdown"))
	require.Contains(t, phase3.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskAddCallEnvelopeValidation, ServiceModuleCalls, "ValidateUnifiedServiceCallForDescriptor"))
	require.Contains(t, phase3.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskAddNoncesIdempotency, ServiceModuleCalls, "ServiceCallReplayIndex"))
	require.Contains(t, phase3.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskAddCallbacksRetries, ServiceModuleCalls, "UnifiedServiceCallback/ServiceRetryPolicy"))
	require.Contains(t, phase3.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskAddDeterministicReceipts, ServiceModuleReceipts, "ReceiptRoot"))
	require.NoError(t, roadmap.ReadyForPhase(ServiceRoadmapPhaseUnifiedCallsReceipts))
	require.NoError(t, ValidateServiceRoadmapExitCriteria(phase3))

	phase4, found := roadmap.PhaseByID(ServiceRoadmapPhasePayments)
	require.True(t, found)
	require.Contains(t, phase4.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskImplementServicePaymentsModule, ServiceModulePayments, "XServicePaymentsModuleBreakdown"))
	require.Contains(t, phase4.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskAddPerCallPayment, ServiceModulePayments, "QuotePerCallPayment"))
	require.Contains(t, phase4.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskAddEscrowSettlement, ServiceModulePayments, "ServiceEscrow/PaymentSettlement"))
	require.Contains(t, phase4.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskAddMeteredUsageReceipt, ServiceModulePayments, "MeteredUsage"))
	require.Contains(t, phase4.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskAddPaymentModelQuery, ServiceModulePayments, "QueryPaymentModel"))
	require.Contains(t, phase4.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskIntegrateBankFinancialZone, ServiceModulePayments, "BuildFinancialZonePaymentRoute"))
	require.NoError(t, roadmap.ReadyForPhase(ServiceRoadmapPhasePayments))
	require.NoError(t, ValidateServiceRoadmapExitCriteria(phase4))

	phase5, found := roadmap.PhaseByID(ServiceRoadmapPhaseOffChainMixedServices)
	require.True(t, found)
	require.Contains(t, phase5.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskAddSignedRequestResponseFormat, ServiceModuleCalls, "ServiceCallEnvelope/ServiceReceipt"))
	require.Contains(t, phase5.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskAddResultAnchoring, ServiceModuleReceipts, "MsgAnchorReceipt"))
	require.Contains(t, phase5.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskAddMixedServiceChallengeFlow, ServiceModuleCalls, "NewServiceChallengeFlow"))
	require.Contains(t, phase5.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskAddFallbackExecutionHooks, ServiceModuleCalls, "ServiceChallengeFlow"))
	require.Contains(t, phase5.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskAddProviderCollateralPenalties, ServiceModuleProviders, "ProviderPenaltyRoute"))
	require.NoError(t, roadmap.ReadyForPhase(ServiceRoadmapPhaseOffChainMixedServices))
	require.NoError(t, ValidateServiceRoadmapExitCriteria(phase5))

	phase6, found := roadmap.PhaseByID(ServiceRoadmapPhaseFogMarketProviders)
	require.True(t, found)
	require.Contains(t, phase6.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskImplementServiceProvidersModule, ServiceModuleProviders, "XServiceProvidersModuleBreakdown"))
	require.Contains(t, phase6.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskAddProviderRegistry, ServiceModuleProviders, "ProviderRecord"))
	require.Contains(t, phase6.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskAddCollateralStaking, ServiceModuleProviders, "ProviderCollateral"))
	require.Contains(t, phase6.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskAddAvailabilityCommitments, ServiceModuleProviders, "AvailabilityCommitment"))
	require.Contains(t, phase6.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskAddReputationCommitments, ServiceModuleProviders, "ReputationRecord"))
	require.Contains(t, phase6.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskAddProviderSelectionQuery, ServiceModuleProviders, "QueryProvidersByService"))
	require.NoError(t, roadmap.ReadyForPhase(ServiceRoadmapPhaseFogMarketProviders))
	require.NoError(t, ValidateServiceRoadmapExitCriteria(phase6))

	phase7, found := roadmap.PhaseByID(ServiceRoadmapPhaseSDKUXTooling)
	require.True(t, found)
	require.Contains(t, phase7.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskAddServiceResolverSDK, ServiceModuleServices, "ServiceResolver"))
	require.Contains(t, phase7.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskAddInterfaceCallBuilder, ServiceModuleCalls, "MethodLevelCallBuilder"))
	require.Contains(t, phase7.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskAddCLICommandGeneration, ServiceModuleInterface, "CLIInterfaceCommandGenerator"))
	require.Contains(t, phase7.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskAddWalletMetadataFormat, ServiceModuleInterface, "WalletMetadataFormat"))
	require.Contains(t, phase7.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskAddProofVerificationHelpers, ServiceModuleReceipts, "ServiceReceiptProofRecord"))
	require.NoError(t, roadmap.ReadyForPhase(ServiceRoadmapPhaseSDKUXTooling))
	require.NoError(t, ValidateServiceRoadmapExitCriteria(phase7))

	phase8, found := roadmap.PhaseByID(ServiceRoadmapPhasePerformanceHardening)
	require.True(t, found)
	require.Contains(t, phase8.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskAddBlockSTMConflictBenchmarks, ServiceModuleCalls, "ServiceBlockSTMOperation"))
	require.Contains(t, phase8.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskAddStoreV2Benchmarks, ServiceModuleServices, "ServiceStoreV2Layout"))
	require.Contains(t, phase8.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskAddServiceCallThroughputTests, ServiceModuleCalls, "UnifiedServiceCall"))
	require.Contains(t, phase8.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskAddReceiptProofBenchmarks, ServiceModuleReceipts, "QueryReceiptProof"))
	require.Contains(t, phase8.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskAddRegistryLookupBenchmarks, ServiceModuleServices, "QueryService/QueryServiceInterface"))
	require.Contains(t, phase8.Tasks, newServiceRoadmapTask(ServiceRoadmapTaskAddMixedDisputeLoadTests, ServiceModuleCalls, "ServiceChallengeFlow"))
	require.NoError(t, roadmap.ReadyForPhase(ServiceRoadmapPhasePerformanceHardening))
	require.NoError(t, ValidateServiceRoadmapExitCriteria(phase8))
}

func TestServiceImplementationRoadmapRejectsMissingPhaseSurface(t *testing.T) {
	roadmap, err := DefaultServiceImplementationRoadmap()
	require.NoError(t, err)
	roadmap.Phases = roadmap.Phases[:2]
	roadmap.RoadmapHash = ComputeServiceImplementationRoadmapHash(roadmap)
	require.ErrorContains(t, roadmap.Validate(), "phases 0, 1, 2, 3, 4, 5, 6, 7, and 8")

	roadmap, err = DefaultServiceImplementationRoadmap()
	require.NoError(t, err)
	phase1, found := roadmap.PhaseByID(ServiceRoadmapPhaseCoreRegistry)
	require.True(t, found)
	phase1.Tasks = removeServiceRoadmapTaskForTest(phase1.Tasks, ServiceRoadmapTaskAddProofQuery)
	phase1, err = NewServiceRoadmapPhase(phase1)
	require.NoError(t, err)
	require.ErrorContains(t, ValidateServiceRoadmapExitCriteria(phase1), "missing task")

	roadmap, err = DefaultServiceImplementationRoadmap()
	require.NoError(t, err)
	phase4, found := roadmap.PhaseByID(ServiceRoadmapPhasePayments)
	require.True(t, found)
	phase4.Tasks = removeServiceRoadmapTaskForTest(phase4.Tasks, ServiceRoadmapTaskAddEscrowSettlement)
	phase4, err = NewServiceRoadmapPhase(phase4)
	require.NoError(t, err)
	require.ErrorContains(t, ValidateServiceRoadmapExitCriteria(phase4), "missing task")

	roadmap, err = DefaultServiceImplementationRoadmap()
	require.NoError(t, err)
	phase5, found := roadmap.PhaseByID(ServiceRoadmapPhaseOffChainMixedServices)
	require.True(t, found)
	phase5.Tasks = removeServiceRoadmapTaskForTest(phase5.Tasks, ServiceRoadmapTaskAddMixedServiceChallengeFlow)
	phase5, err = NewServiceRoadmapPhase(phase5)
	require.NoError(t, err)
	require.ErrorContains(t, ValidateServiceRoadmapExitCriteria(phase5), "missing task")

	roadmap, err = DefaultServiceImplementationRoadmap()
	require.NoError(t, err)
	phase6, found := roadmap.PhaseByID(ServiceRoadmapPhaseFogMarketProviders)
	require.True(t, found)
	phase6.Tasks = removeServiceRoadmapTaskForTest(phase6.Tasks, ServiceRoadmapTaskAddCollateralStaking)
	phase6, err = NewServiceRoadmapPhase(phase6)
	require.NoError(t, err)
	require.ErrorContains(t, ValidateServiceRoadmapExitCriteria(phase6), "missing task")

	roadmap, err = DefaultServiceImplementationRoadmap()
	require.NoError(t, err)
	phase7, found := roadmap.PhaseByID(ServiceRoadmapPhaseSDKUXTooling)
	require.True(t, found)
	phase7.Tasks = removeServiceRoadmapTaskForTest(phase7.Tasks, ServiceRoadmapTaskAddInterfaceCallBuilder)
	phase7, err = NewServiceRoadmapPhase(phase7)
	require.NoError(t, err)
	require.ErrorContains(t, ValidateServiceRoadmapExitCriteria(phase7), "missing task")

	roadmap, err = DefaultServiceImplementationRoadmap()
	require.NoError(t, err)
	phase8, found := roadmap.PhaseByID(ServiceRoadmapPhasePerformanceHardening)
	require.True(t, found)
	phase8.Tasks = removeServiceRoadmapTaskForTest(phase8.Tasks, ServiceRoadmapTaskAddStoreV2Benchmarks)
	phase8, err = NewServiceRoadmapPhase(phase8)
	require.NoError(t, err)
	require.ErrorContains(t, ValidateServiceRoadmapExitCriteria(phase8), "missing task")
}

func TestServiceRoadmapPhase0ExitCriteriaEnforceCompatibilityArtifacts(t *testing.T) {
	roadmap, err := DefaultServiceImplementationRoadmap()
	require.NoError(t, err)
	phase0, found := roadmap.PhaseByID(ServiceRoadmapPhaseSpecificationCompatibility)
	require.True(t, found)

	require.NoError(t, validateServiceCoreObjectDefinitions(phase0.CoreObjects))
	require.NoError(t, validateServiceSignableObjectVectors(phase0.SignableVectors))
	require.NoError(t, ValidateAetraModuleServiceMappings(phase0.ModuleMappings))

	phase0.CoreObjects = phase0.CoreObjects[1:]
	phase0, err = NewServiceRoadmapPhase(phase0)
	require.NoError(t, err)
	require.NoError(t, ValidateServiceRoadmapExitCriteria(phase0))

	mapping := phase0.ModuleMappings[0]
	mapping.OnChain = false
	mapping.MappingHash = ComputeAetraModuleServiceMappingHash(mapping)
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

	mapping := newAetraModuleServiceMapping("avm-dex-contract", "avm-dex-contract", "avm-dex-contract-service", true)
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
