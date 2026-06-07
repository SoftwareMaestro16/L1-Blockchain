package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultImplementationRoadmapCoversPhaseZeroThroughSeven(t *testing.T) {
	roadmap, err := DefaultImplementationRoadmap()
	require.NoError(t, err)
	require.NoError(t, roadmap.Validate())
	require.Len(t, roadmap.Phases, 8)
	require.Equal(t, RoadmapPhaseBaselineAudit, roadmap.Phases[0].PhaseID)
	require.Equal(t, uint32(0), roadmap.Phases[0].PhaseNumber)
	require.Equal(t, RoadmapPhaseKernelRootModel, roadmap.Phases[1].PhaseID)
	require.Equal(t, uint32(1), roadmap.Phases[1].PhaseNumber)
	require.Equal(t, RoadmapPhaseCrossZoneMessages, roadmap.Phases[2].PhaseID)
	require.Equal(t, uint32(2), roadmap.Phases[2].PhaseNumber)
	require.Equal(t, RoadmapPhaseCanonicalZones, roadmap.Phases[3].PhaseID)
	require.Equal(t, uint32(3), roadmap.Phases[3].PhaseNumber)
	require.Equal(t, RoadmapPhaseServiceStorageRouting, roadmap.Phases[4].PhaseID)
	require.Equal(t, uint32(4), roadmap.Phases[4].PhaseNumber)
	require.Equal(t, RoadmapPhaseIdentityPaymentIntegration, roadmap.Phases[5].PhaseID)
	require.Equal(t, uint32(5), roadmap.Phases[5].PhaseNumber)
	require.Equal(t, RoadmapPhaseVMRuntime, roadmap.Phases[6].PhaseID)
	require.Equal(t, uint32(6), roadmap.Phases[6].PhaseNumber)
	require.Equal(t, RoadmapPhasePerformanceHardening, roadmap.Phases[7].PhaseID)
	require.Equal(t, uint32(7), roadmap.Phases[7].PhaseNumber)
	require.Equal(t, ComputeImplementationRoadmapHash(roadmap), roadmap.RoadmapHash)

	phase0 := roadmap.Phases[0]
	requireRoadmapTask(t, phase0, "inventory-current-modules-state-keys")
	requireRoadmapTask(t, phase0, "identify-cross-module-direct-writes")
	requireRoadmapTask(t, phase0, "add-export-import-tests-current-state")
	requireRoadmapTask(t, phase0, "add-module-invariant-test-harness")
	requireRoadmapTask(t, phase0, "add-root-contribution-interface-design")
	requireRoadmapExit(t, phase0, "current-aetra-state-reproducible")
	requireRoadmapExit(t, phase0, "current-module-boundaries-documented")
	requireRoadmapExit(t, phase0, "migration-risk-list-complete")
	require.Len(t, phase0.Evidence.ModuleInventory, len(RequiredCosmosSDKModules()))
	require.True(t, phase0.Evidence.CrossModuleDirectWritesAudited)
	require.True(t, phase0.Evidence.ExportImportTestsAdded)
	require.True(t, phase0.Evidence.ModuleInvariantHarnessAdded)
	require.True(t, phase0.Evidence.RootContributionInterfaceDesign)
	require.True(t, phase0.Evidence.CurrentStateReproducible)
	require.True(t, phase0.Evidence.ModuleBoundariesDocumented)
	require.True(t, phase0.Evidence.MigrationRiskListComplete)

	phase1 := roadmap.Phases[1]
	requireRoadmapTask(t, phase1, "implement-x-aetracore")
	requireRoadmapTask(t, phase1, "implement-x-zones")
	requireRoadmapTask(t, phase1, "add-zone-registry")
	requireRoadmapTask(t, phase1, "add-root-contribution-interface")
	requireRoadmapTask(t, phase1, "add-global-state-root")
	requireRoadmapTask(t, phase1, "add-block-commitment-metadata-queries")
	requireRoadmapExit(t, phase1, "existing-chain-runs-as-default-zone")
	requireRoadmapExit(t, phase1, "global-root-includes-default-zone-root")
	requireRoadmapExit(t, phase1, "export-import-preserves-root-metadata")
	require.True(t, phase1.Evidence.AetraCoreModuleImplemented)
	require.True(t, phase1.Evidence.ZonesModuleImplemented)
	require.True(t, phase1.Evidence.ZoneRegistryImplemented)
	require.True(t, phase1.Evidence.GlobalStateRootImplemented)
	require.True(t, phase1.Evidence.BlockCommitmentMetadataQueries)
	require.True(t, phase1.Evidence.DefaultZoneRunnable)
	require.True(t, phase1.Evidence.DefaultZoneRootIncluded)
	require.True(t, phase1.Evidence.ExportImportPreservesRootMeta)

	phase2 := roadmap.Phases[2]
	requireRoadmapTask(t, phase2, "implement-x-messages")
	requireRoadmapTask(t, phase2, "add-message-envelope")
	requireRoadmapTask(t, phase2, "add-fifo-per-sender-queues")
	requireRoadmapTask(t, phase2, "add-nonce-and-replay-protection")
	requireRoadmapTask(t, phase2, "add-receipts")
	requireRoadmapTask(t, phase2, "add-bounce-and-expiry")
	requireRoadmapTask(t, phase2, "add-message-and-receipt-roots")
	requireRoadmapExit(t, phase2, "same-chain-async-messages-execute-deterministically")
	requireRoadmapExit(t, phase2, "message-inclusion-and-receipt-proofs-available")
	requireRoadmapExit(t, phase2, "replay-attempts-rejected")
	require.True(t, phase2.Evidence.MessagesModuleImplemented)
	require.True(t, phase2.Evidence.MessageEnvelopeAdded)
	require.True(t, phase2.Evidence.FIFOPerSenderQueuesAdded)
	require.True(t, phase2.Evidence.NonceReplayProtectionAdded)
	require.True(t, phase2.Evidence.MessageReceiptsAdded)
	require.True(t, phase2.Evidence.BounceAndExpiryAdded)
	require.True(t, phase2.Evidence.MessageReceiptRootsAdded)
	require.True(t, phase2.Evidence.SameChainAsyncDeterministic)
	require.True(t, phase2.Evidence.MessageReceiptProofsAvailable)
	require.True(t, phase2.Evidence.ReplayAttemptsRejected)

	phase3 := roadmap.Phases[3]
	requireRoadmapTask(t, phase3, "move-bank-fees-contract-assets-dex-into-financial-zone-boundary")
	requireRoadmapTask(t, phase3, "activate-identity-zone")
	requireRoadmapTask(t, phase3, "add-application-zone-scheduler-boundary")
	requireRoadmapTask(t, phase3, "add-contract-zone-skeleton")
	requireRoadmapTask(t, phase3, "add-zone-specific-queries-and-roots")
	requireRoadmapExit(t, phase3, "four-canonical-zones-exist")
	requireRoadmapExit(t, phase3, "each-zone-has-namespace-root-queue-msg-query-keeper")
	requireRoadmapExit(t, phase3, "cross-zone-state-mutation-only-through-messages")
	require.Equal(t, []ZoneID{ZoneIDApplication, ZoneIDContract, ZoneIDFinancial, ZoneIDIdentity}, roadmapCanonicalZoneIDs(phase3.Evidence.CanonicalZones))
	require.True(t, phase3.Evidence.FinancialZoneBoundaryMoved)
	require.True(t, phase3.Evidence.IdentityZoneActivated)
	require.True(t, phase3.Evidence.ApplicationZoneSchedulerBoundary)
	require.True(t, phase3.Evidence.ContractZoneSkeletonAdded)
	require.True(t, phase3.Evidence.ZoneSpecificQueriesRoots)
	require.True(t, phase3.Evidence.FourCanonicalZonesExist)
	require.True(t, phase3.Evidence.CanonicalZoneSurfacesComplete)
	require.True(t, phase3.Evidence.CrossZoneMutationMessagesOnly)
	for _, zone := range phase3.Evidence.CanonicalZones {
		require.True(t, zone.MessageQueue, zone.ZoneID)
		require.True(t, zone.MsgServer, zone.ZoneID)
		require.True(t, zone.QueryServer, zone.ZoneID)
		require.True(t, zone.Keeper, zone.ZoneID)
		require.NotEmpty(t, zone.StateNamespace)
		require.NotEmpty(t, zone.RootType)
	}

	phase4 := roadmap.Phases[4]
	requireRoadmapTask(t, phase4, "implement-x-services")
	requireRoadmapTask(t, phase4, "implement-x-storage")
	requireRoadmapTask(t, phase4, "implement-x-routing")
	requireRoadmapTask(t, phase4, "add-service-descriptors")
	requireRoadmapTask(t, phase4, "add-storage-object-commitments")
	requireRoadmapTask(t, phase4, "add-node-records-and-routing-table-epochs")
	requireRoadmapTask(t, phase4, "add-proof-attached-lookup-queries")
	requireRoadmapExit(t, phase4, "service-discovery-deterministic")
	requireRoadmapExit(t, phase4, "storage-commitments-proof-verifiable")
	requireRoadmapExit(t, phase4, "routing-table-committed-and-queryable")
	require.True(t, phase4.Evidence.ServicesModuleImplemented)
	require.True(t, phase4.Evidence.StorageModuleImplemented)
	require.True(t, phase4.Evidence.RoutingModuleImplemented)
	require.True(t, phase4.Evidence.ServiceDescriptorsAdded)
	require.True(t, phase4.Evidence.StorageObjectCommitmentsAdded)
	require.True(t, phase4.Evidence.NodeRecordsRoutingEpochsAdded)
	require.True(t, phase4.Evidence.ProofAttachedLookupQueriesAdded)
	require.True(t, phase4.Evidence.ServiceDiscoveryDeterministic)
	require.True(t, phase4.Evidence.StorageCommitmentsProofVerifiable)
	require.True(t, phase4.Evidence.RoutingTableCommittedQueryable)

	phase5 := roadmap.Phases[5]
	requireRoadmapTask(t, phase5, "upgrade-aet-resolver-outputs")
	requireRoadmapTask(t, phase5, "add-identity-graph")
	requireRoadmapTask(t, phase5, "add-cross-zone-identity-binding")
	requireRoadmapTask(t, phase5, "implement-x-payments")
	requireRoadmapTask(t, phase5, "add-payment-envelope")
	requireRoadmapTask(t, phase5, "add-conditional-transfers")
	requireRoadmapTask(t, phase5, "add-settlement-in-financial-zone")
	requireRoadmapExit(t, phase5, "identity-resolves-account-zone-service-contract-composite")
	requireRoadmapExit(t, phase5, "payments-settle-through-financial-zone")
	requireRoadmapExit(t, phase5, "payment-disputes-resolve-by-deterministic-replay")
	require.Equal(t, []string{"account", "composite", "contract", "service", "zone"}, phase5.Evidence.IdentityResolverOutputs)
	require.True(t, phase5.Evidence.AETResolverOutputsUpgraded)
	require.True(t, phase5.Evidence.IdentityGraphAdded)
	require.True(t, phase5.Evidence.CrossZoneIdentityBindingAdded)
	require.True(t, phase5.Evidence.PaymentsModuleImplemented)
	require.True(t, phase5.Evidence.PaymentEnvelopeAdded)
	require.True(t, phase5.Evidence.ConditionalTransfersAdded)
	require.True(t, phase5.Evidence.FinancialZoneSettlementAdded)
	require.True(t, phase5.Evidence.IdentityResolvesAllOutputTypes)
	require.True(t, phase5.Evidence.PaymentsSettleThroughFinancialZone)
	require.True(t, phase5.Evidence.PaymentDisputesDeterministicReplay)

	phase6 := roadmap.Phases[6]
	requireRoadmapTask(t, phase6, "implement-x-contracts")
	requireRoadmapTask(t, phase6, "add-avm-ready-bytecode-interface")
	requireRoadmapTask(t, phase6, "add-cosmwasm-adapter-boundary")
	requireRoadmapTask(t, phase6, "add-vm-storage-adapter")
	requireRoadmapTask(t, phase6, "add-vm-outbound-message-support")
	requireRoadmapTask(t, phase6, "add-contract-receipts-and-proofs")
	requireRoadmapExit(t, phase6, "contract-execution-message-driven")
	requireRoadmapExit(t, phase6, "contracts-cannot-directly-mutate-other-zones")
	requireRoadmapExit(t, phase6, "contract-state-root-proof-verifiable")
	require.True(t, phase6.Evidence.ContractsModuleImplemented)
	require.True(t, phase6.Evidence.AVMBytecodeInterfaceAdded)
	require.True(t, phase6.Evidence.CosmWasmAdapterBoundaryAdded)
	require.True(t, phase6.Evidence.VMStorageAdapterAdded)
	require.True(t, phase6.Evidence.VMOutboundMessageSupportAdded)
	require.True(t, phase6.Evidence.ContractReceiptsProofsAdded)
	require.True(t, phase6.Evidence.ContractExecutionMessageDriven)
	require.True(t, phase6.Evidence.ContractsNoDirectZoneMutation)
	require.True(t, phase6.Evidence.ContractStateRootProofVerifiable)

	phase7 := roadmap.Phases[7]
	requireRoadmapTask(t, phase7, "add-blockstm-aware-workload-grouping")
	requireRoadmapTask(t, phase7, "add-store-v2-optimization-for-root-heavy-reads")
	requireRoadmapTask(t, phase7, "add-queue-draining-benchmarks")
	requireRoadmapTask(t, phase7, "add-service-lookup-benchmarks")
	requireRoadmapTask(t, phase7, "add-storage-proof-benchmarks")
	requireRoadmapTask(t, phase7, "add-routing-simulation-tests")
	requireRoadmapTask(t, phase7, "add-adaptivesync-recovery-tests")
	requireRoadmapExit(t, phase7, "independent-zone-workloads-parallelize")
	requireRoadmapExit(t, phase7, "root-generation-remains-bounded")
	requireRoadmapExit(t, phase7, "nodes-recover-and-serve-proof-queries-after-sync")
	require.True(t, phase7.Evidence.BlockSTMAwareGroupingAdded)
	require.True(t, phase7.Evidence.StoreV2RootReadOptimizationAdded)
	require.True(t, phase7.Evidence.QueueDrainingBenchmarksAdded)
	require.True(t, phase7.Evidence.ServiceLookupBenchmarksAdded)
	require.True(t, phase7.Evidence.StorageProofBenchmarksAdded)
	require.True(t, phase7.Evidence.RoutingSimulationTestsAdded)
	require.True(t, phase7.Evidence.AdaptiveSyncRecoveryTestsAdded)
	require.True(t, phase7.Evidence.IndependentZoneWorkloadsParallelize)
	require.True(t, phase7.Evidence.RootGenerationBounded)
	require.True(t, phase7.Evidence.NodesRecoverServeProofQueriesAfterSync)
}

func TestImplementationRoadmapInventoryIsDerivedFromModuleManifest(t *testing.T) {
	manifest, err := DefaultCosmosModuleRequirementManifest()
	require.NoError(t, err)

	inventory := BuildRoadmapModuleInventory(manifest)
	require.Len(t, inventory, len(RequiredCosmosSDKModules()))
	require.Equal(t, RequiredCosmosSDKModules(), roadmapInventoryModuleNames(inventory))

	for _, entry := range inventory {
		require.NotEmpty(t, entry.ModulePath)
		require.Equal(t, string(entry.ModuleName), entry.StoreKey)
		require.Contains(t, entry.StateKeys, string(entry.ModuleName)+"/params")
		require.Contains(t, entry.StateKeys, string(entry.ModuleName)+"/genesis")
		require.Contains(t, entry.StateKeys, string(entry.ModuleName)+"/root")
		require.NotEmpty(t, entry.RootType)
	}
}

func TestImplementationRoadmapRejectsIncompletePhaseZeroEvidence(t *testing.T) {
	roadmap, err := DefaultImplementationRoadmap()
	require.NoError(t, err)

	missingInventory := roadmap
	missingInventory.Phases = append([]ImplementationRoadmapPhase(nil), roadmap.Phases...)
	missingInventory.Phases[0].Evidence.ModuleInventory = missingInventory.Phases[0].Evidence.ModuleInventory[1:]
	missingInventory.Phases[0].PhaseHash = ComputeRoadmapPhaseHash(missingInventory.Phases[0])
	missingInventory.RoadmapHash = ComputeImplementationRoadmapHash(missingInventory)
	require.ErrorContains(t, missingInventory.Validate(), "must include 9 required modules")

	noExportImportTests := roadmap
	noExportImportTests.Phases = append([]ImplementationRoadmapPhase(nil), roadmap.Phases...)
	noExportImportTests.Phases[0].Evidence.ExportImportTestsAdded = false
	noExportImportTests.Phases[0].PhaseHash = ComputeRoadmapPhaseHash(noExportImportTests.Phases[0])
	noExportImportTests.RoadmapHash = ComputeImplementationRoadmapHash(noExportImportTests)
	require.ErrorContains(t, noExportImportTests.Validate(), "export/import tests")

	incompleteExit := roadmap
	incompleteExit.Phases = append([]ImplementationRoadmapPhase(nil), roadmap.Phases...)
	incompleteExit.Phases[0].ExitCriteria[0].Complete = false
	incompleteExit.Phases[0].PhaseHash = ComputeRoadmapPhaseHash(incompleteExit.Phases[0])
	incompleteExit.RoadmapHash = ComputeImplementationRoadmapHash(incompleteExit)
	require.ErrorContains(t, incompleteExit.Validate(), "incomplete exit criteria")
}

func TestImplementationRoadmapRejectsIncompletePhaseOneEvidence(t *testing.T) {
	roadmap, err := DefaultImplementationRoadmap()
	require.NoError(t, err)

	noZones := roadmap
	noZones.Phases = append([]ImplementationRoadmapPhase(nil), roadmap.Phases...)
	noZones.Phases[1].Evidence.ZonesModuleImplemented = false
	noZones.Phases[1].PhaseHash = ComputeRoadmapPhaseHash(noZones.Phases[1])
	noZones.RoadmapHash = ComputeImplementationRoadmapHash(noZones)
	require.ErrorContains(t, noZones.Validate(), "x/zones")

	noGlobalRoot := roadmap
	noGlobalRoot.Phases = append([]ImplementationRoadmapPhase(nil), roadmap.Phases...)
	noGlobalRoot.Phases[1].Evidence.GlobalStateRootImplemented = false
	noGlobalRoot.Phases[1].PhaseHash = ComputeRoadmapPhaseHash(noGlobalRoot.Phases[1])
	noGlobalRoot.RoadmapHash = ComputeImplementationRoadmapHash(noGlobalRoot)
	require.ErrorContains(t, noGlobalRoot.Validate(), "GlobalStateRoot")

	noRootMetadataRoundTrip := roadmap
	noRootMetadataRoundTrip.Phases = append([]ImplementationRoadmapPhase(nil), roadmap.Phases...)
	noRootMetadataRoundTrip.Phases[1].Evidence.ExportImportPreservesRootMeta = false
	noRootMetadataRoundTrip.Phases[1].PhaseHash = ComputeRoadmapPhaseHash(noRootMetadataRoundTrip.Phases[1])
	noRootMetadataRoundTrip.RoadmapHash = ComputeImplementationRoadmapHash(noRootMetadataRoundTrip)
	require.ErrorContains(t, noRootMetadataRoundTrip.Validate(), "root metadata preservation")
}

func TestImplementationRoadmapRejectsIncompletePhaseTwoEvidence(t *testing.T) {
	roadmap, err := DefaultImplementationRoadmap()
	require.NoError(t, err)

	noEnvelope := roadmap
	noEnvelope.Phases = append([]ImplementationRoadmapPhase(nil), roadmap.Phases...)
	noEnvelope.Phases[2].Evidence.MessageEnvelopeAdded = false
	noEnvelope.Phases[2].PhaseHash = ComputeRoadmapPhaseHash(noEnvelope.Phases[2])
	noEnvelope.RoadmapHash = ComputeImplementationRoadmapHash(noEnvelope)
	require.ErrorContains(t, noEnvelope.Validate(), "message envelope")

	noProofs := roadmap
	noProofs.Phases = append([]ImplementationRoadmapPhase(nil), roadmap.Phases...)
	noProofs.Phases[2].Evidence.MessageReceiptProofsAvailable = false
	noProofs.Phases[2].PhaseHash = ComputeRoadmapPhaseHash(noProofs.Phases[2])
	noProofs.RoadmapHash = ComputeImplementationRoadmapHash(noProofs)
	require.ErrorContains(t, noProofs.Validate(), "message inclusion and receipt proofs")

	noReplayRejection := roadmap
	noReplayRejection.Phases = append([]ImplementationRoadmapPhase(nil), roadmap.Phases...)
	noReplayRejection.Phases[2].Evidence.ReplayAttemptsRejected = false
	noReplayRejection.Phases[2].PhaseHash = ComputeRoadmapPhaseHash(noReplayRejection.Phases[2])
	noReplayRejection.RoadmapHash = ComputeImplementationRoadmapHash(noReplayRejection)
	require.ErrorContains(t, noReplayRejection.Validate(), "replay rejection")
}

func TestImplementationRoadmapRejectsIncompletePhaseThreeEvidence(t *testing.T) {
	roadmap, err := DefaultImplementationRoadmap()
	require.NoError(t, err)

	noFinancialBoundary := roadmap
	noFinancialBoundary.Phases = append([]ImplementationRoadmapPhase(nil), roadmap.Phases...)
	noFinancialBoundary.Phases[3].Evidence.FinancialZoneBoundaryMoved = false
	noFinancialBoundary.Phases[3].PhaseHash = ComputeRoadmapPhaseHash(noFinancialBoundary.Phases[3])
	noFinancialBoundary.RoadmapHash = ComputeImplementationRoadmapHash(noFinancialBoundary)
	require.ErrorContains(t, noFinancialBoundary.Validate(), "Financial Zone boundary")

	missingZone := roadmap
	missingZone.Phases = append([]ImplementationRoadmapPhase(nil), roadmap.Phases...)
	missingZone.Phases[3].Evidence.CanonicalZones = missingZone.Phases[3].Evidence.CanonicalZones[1:]
	missingZone.Phases[3].PhaseHash = ComputeRoadmapPhaseHash(missingZone.Phases[3])
	missingZone.RoadmapHash = ComputeImplementationRoadmapHash(missingZone)
	require.ErrorContains(t, missingZone.Validate(), "must include 4 zones")

	missingSurface := roadmap
	missingSurface.Phases = append([]ImplementationRoadmapPhase(nil), roadmap.Phases...)
	missingSurface.Phases[3].Evidence.CanonicalZones[0].MsgServer = false
	missingSurface.Phases[3].PhaseHash = ComputeRoadmapPhaseHash(missingSurface.Phases[3])
	missingSurface.RoadmapHash = ComputeImplementationRoadmapHash(missingSurface)
	require.ErrorContains(t, missingSurface.Validate(), "MsgServer")

	directMutation := roadmap
	directMutation.Phases = append([]ImplementationRoadmapPhase(nil), roadmap.Phases...)
	directMutation.Phases[3].Evidence.CrossZoneMutationMessagesOnly = false
	directMutation.Phases[3].PhaseHash = ComputeRoadmapPhaseHash(directMutation.Phases[3])
	directMutation.RoadmapHash = ComputeImplementationRoadmapHash(directMutation)
	require.ErrorContains(t, directMutation.Validate(), "cross-zone mutation only through messages")
}

func TestImplementationRoadmapRejectsIncompletePhaseFourEvidence(t *testing.T) {
	roadmap, err := DefaultImplementationRoadmap()
	require.NoError(t, err)

	noServices := roadmap
	noServices.Phases = append([]ImplementationRoadmapPhase(nil), roadmap.Phases...)
	noServices.Phases[4].Evidence.ServicesModuleImplemented = false
	noServices.Phases[4].PhaseHash = ComputeRoadmapPhaseHash(noServices.Phases[4])
	noServices.RoadmapHash = ComputeImplementationRoadmapHash(noServices)
	require.ErrorContains(t, noServices.Validate(), "x/services")

	noStorageCommitments := roadmap
	noStorageCommitments.Phases = append([]ImplementationRoadmapPhase(nil), roadmap.Phases...)
	noStorageCommitments.Phases[4].Evidence.StorageObjectCommitmentsAdded = false
	noStorageCommitments.Phases[4].PhaseHash = ComputeRoadmapPhaseHash(noStorageCommitments.Phases[4])
	noStorageCommitments.RoadmapHash = ComputeImplementationRoadmapHash(noStorageCommitments)
	require.ErrorContains(t, noStorageCommitments.Validate(), "storage object commitments")

	noProofLookup := roadmap
	noProofLookup.Phases = append([]ImplementationRoadmapPhase(nil), roadmap.Phases...)
	noProofLookup.Phases[4].Evidence.ProofAttachedLookupQueriesAdded = false
	noProofLookup.Phases[4].PhaseHash = ComputeRoadmapPhaseHash(noProofLookup.Phases[4])
	noProofLookup.RoadmapHash = ComputeImplementationRoadmapHash(noProofLookup)
	require.ErrorContains(t, noProofLookup.Validate(), "proof-attached lookup queries")

	notQueryable := roadmap
	notQueryable.Phases = append([]ImplementationRoadmapPhase(nil), roadmap.Phases...)
	notQueryable.Phases[4].Evidence.RoutingTableCommittedQueryable = false
	notQueryable.Phases[4].PhaseHash = ComputeRoadmapPhaseHash(notQueryable.Phases[4])
	notQueryable.RoadmapHash = ComputeImplementationRoadmapHash(notQueryable)
	require.ErrorContains(t, notQueryable.Validate(), "committed and queryable routing table")
}

func TestImplementationRoadmapRejectsIncompletePhaseFiveEvidence(t *testing.T) {
	roadmap, err := DefaultImplementationRoadmap()
	require.NoError(t, err)

	noResolverOutputUpgrade := roadmap
	noResolverOutputUpgrade.Phases = append([]ImplementationRoadmapPhase(nil), roadmap.Phases...)
	noResolverOutputUpgrade.Phases[5].Evidence.AETResolverOutputsUpgraded = false
	noResolverOutputUpgrade.Phases[5].PhaseHash = ComputeRoadmapPhaseHash(noResolverOutputUpgrade.Phases[5])
	noResolverOutputUpgrade.RoadmapHash = ComputeImplementationRoadmapHash(noResolverOutputUpgrade)
	require.ErrorContains(t, noResolverOutputUpgrade.Validate(), ".aet resolver outputs")

	missingOutput := roadmap
	missingOutput.Phases = append([]ImplementationRoadmapPhase(nil), roadmap.Phases...)
	missingOutput.Phases[5].Evidence.IdentityResolverOutputs = []string{"account", "composite", "contract", "service"}
	missingOutput.Phases[5].PhaseHash = ComputeRoadmapPhaseHash(missingOutput.Phases[5])
	missingOutput.RoadmapHash = ComputeImplementationRoadmapHash(missingOutput)
	require.ErrorContains(t, missingOutput.Validate(), "must include 5 output types")

	noFinancialSettlement := roadmap
	noFinancialSettlement.Phases = append([]ImplementationRoadmapPhase(nil), roadmap.Phases...)
	noFinancialSettlement.Phases[5].Evidence.PaymentsSettleThroughFinancialZone = false
	noFinancialSettlement.Phases[5].PhaseHash = ComputeRoadmapPhaseHash(noFinancialSettlement.Phases[5])
	noFinancialSettlement.RoadmapHash = ComputeImplementationRoadmapHash(noFinancialSettlement)
	require.ErrorContains(t, noFinancialSettlement.Validate(), "settle through Financial Zone")

	noDeterministicDisputes := roadmap
	noDeterministicDisputes.Phases = append([]ImplementationRoadmapPhase(nil), roadmap.Phases...)
	noDeterministicDisputes.Phases[5].Evidence.PaymentDisputesDeterministicReplay = false
	noDeterministicDisputes.Phases[5].PhaseHash = ComputeRoadmapPhaseHash(noDeterministicDisputes.Phases[5])
	noDeterministicDisputes.RoadmapHash = ComputeImplementationRoadmapHash(noDeterministicDisputes)
	require.ErrorContains(t, noDeterministicDisputes.Validate(), "deterministic replay payment disputes")
}

func TestImplementationRoadmapRejectsIncompletePhaseSixEvidence(t *testing.T) {
	roadmap, err := DefaultImplementationRoadmap()
	require.NoError(t, err)

	noContracts := roadmap
	noContracts.Phases = append([]ImplementationRoadmapPhase(nil), roadmap.Phases...)
	noContracts.Phases[6].Evidence.ContractsModuleImplemented = false
	noContracts.Phases[6].PhaseHash = ComputeRoadmapPhaseHash(noContracts.Phases[6])
	noContracts.RoadmapHash = ComputeImplementationRoadmapHash(noContracts)
	require.ErrorContains(t, noContracts.Validate(), "x/contracts")

	noCosmWasmBoundary := roadmap
	noCosmWasmBoundary.Phases = append([]ImplementationRoadmapPhase(nil), roadmap.Phases...)
	noCosmWasmBoundary.Phases[6].Evidence.CosmWasmAdapterBoundaryAdded = false
	noCosmWasmBoundary.Phases[6].PhaseHash = ComputeRoadmapPhaseHash(noCosmWasmBoundary.Phases[6])
	noCosmWasmBoundary.RoadmapHash = ComputeImplementationRoadmapHash(noCosmWasmBoundary)
	require.ErrorContains(t, noCosmWasmBoundary.Validate(), "CosmWasm adapter boundary")

	directZoneMutation := roadmap
	directZoneMutation.Phases = append([]ImplementationRoadmapPhase(nil), roadmap.Phases...)
	directZoneMutation.Phases[6].Evidence.ContractsNoDirectZoneMutation = false
	directZoneMutation.Phases[6].PhaseHash = ComputeRoadmapPhaseHash(directZoneMutation.Phases[6])
	directZoneMutation.RoadmapHash = ComputeImplementationRoadmapHash(directZoneMutation)
	require.ErrorContains(t, directZoneMutation.Validate(), "direct mutation of other zones")

	noStateProof := roadmap
	noStateProof.Phases = append([]ImplementationRoadmapPhase(nil), roadmap.Phases...)
	noStateProof.Phases[6].Evidence.ContractStateRootProofVerifiable = false
	noStateProof.Phases[6].PhaseHash = ComputeRoadmapPhaseHash(noStateProof.Phases[6])
	noStateProof.RoadmapHash = ComputeImplementationRoadmapHash(noStateProof)
	require.ErrorContains(t, noStateProof.Validate(), "proof-verifiable contract state root")
}

func TestImplementationRoadmapRejectsIncompletePhaseSevenEvidence(t *testing.T) {
	roadmap, err := DefaultImplementationRoadmap()
	require.NoError(t, err)

	noBlockSTM := roadmap
	noBlockSTM.Phases = append([]ImplementationRoadmapPhase(nil), roadmap.Phases...)
	noBlockSTM.Phases[7].Evidence.BlockSTMAwareGroupingAdded = false
	noBlockSTM.Phases[7].PhaseHash = ComputeRoadmapPhaseHash(noBlockSTM.Phases[7])
	noBlockSTM.RoadmapHash = ComputeImplementationRoadmapHash(noBlockSTM)
	require.ErrorContains(t, noBlockSTM.Validate(), "BlockSTM-aware workload grouping")

	noStorageBench := roadmap
	noStorageBench.Phases = append([]ImplementationRoadmapPhase(nil), roadmap.Phases...)
	noStorageBench.Phases[7].Evidence.StorageProofBenchmarksAdded = false
	noStorageBench.Phases[7].PhaseHash = ComputeRoadmapPhaseHash(noStorageBench.Phases[7])
	noStorageBench.RoadmapHash = ComputeImplementationRoadmapHash(noStorageBench)
	require.ErrorContains(t, noStorageBench.Validate(), "storage proof benchmarks")

	noParallelism := roadmap
	noParallelism.Phases = append([]ImplementationRoadmapPhase(nil), roadmap.Phases...)
	noParallelism.Phases[7].Evidence.IndependentZoneWorkloadsParallelize = false
	noParallelism.Phases[7].PhaseHash = ComputeRoadmapPhaseHash(noParallelism.Phases[7])
	noParallelism.RoadmapHash = ComputeImplementationRoadmapHash(noParallelism)
	require.ErrorContains(t, noParallelism.Validate(), "independent zone workloads to parallelize")

	noSyncProofQueries := roadmap
	noSyncProofQueries.Phases = append([]ImplementationRoadmapPhase(nil), roadmap.Phases...)
	noSyncProofQueries.Phases[7].Evidence.NodesRecoverServeProofQueriesAfterSync = false
	noSyncProofQueries.Phases[7].PhaseHash = ComputeRoadmapPhaseHash(noSyncProofQueries.Phases[7])
	noSyncProofQueries.RoadmapHash = ComputeImplementationRoadmapHash(noSyncProofQueries)
	require.ErrorContains(t, noSyncProofQueries.Validate(), "serve proof queries after sync")
}

func TestImplementationRoadmapHashIsCanonical(t *testing.T) {
	roadmap, err := DefaultImplementationRoadmap()
	require.NoError(t, err)

	reversed, err := NewImplementationRoadmap([]ImplementationRoadmapPhase{
		roadmap.Phases[7],
		roadmap.Phases[6],
		roadmap.Phases[5],
		roadmap.Phases[4],
		roadmap.Phases[3],
		roadmap.Phases[2],
		roadmap.Phases[1],
		roadmap.Phases[0],
	})
	require.NoError(t, err)
	require.Equal(t, roadmap.RoadmapHash, reversed.RoadmapHash)
	require.Equal(t, roadmap.Phases, reversed.Phases)

	tampered := roadmap
	tampered.Phases = append([]ImplementationRoadmapPhase(nil), roadmap.Phases...)
	tampered.Phases[0].Tasks[0].Description = "tampered baseline task"
	tampered.Phases[0].PhaseHash = ComputeRoadmapPhaseHash(tampered.Phases[0])
	tampered.RoadmapHash = ComputeImplementationRoadmapHash(tampered)
	require.NotEqual(t, roadmap.RoadmapHash, tampered.RoadmapHash)
	require.NoError(t, tampered.Validate())
}

func requireRoadmapTask(t *testing.T, phase ImplementationRoadmapPhase, id string) {
	t.Helper()
	for _, task := range phase.Tasks {
		if task.ID == id {
			require.True(t, task.Complete)
			return
		}
	}
	t.Fatalf("missing roadmap task %s", id)
}

func requireRoadmapExit(t *testing.T, phase ImplementationRoadmapPhase, id string) {
	t.Helper()
	for _, criterion := range phase.ExitCriteria {
		if criterion.ID == id {
			require.True(t, criterion.Complete)
			return
		}
	}
	t.Fatalf("missing roadmap exit criterion %s", id)
}

func roadmapInventoryModuleNames(entries []RoadmapModuleInventoryEntry) []CosmosSDKModuleName {
	entries = normalizeRoadmapModuleInventory(entries)
	out := make([]CosmosSDKModuleName, len(entries))
	for i, entry := range entries {
		out[i] = entry.ModuleName
	}
	return out
}

func roadmapCanonicalZoneIDs(entries []RoadmapCanonicalZoneEntry) []ZoneID {
	entries = normalizeRoadmapCanonicalZones(entries)
	out := make([]ZoneID, len(entries))
	for i, entry := range entries {
		out[i] = entry.ZoneID
	}
	return out
}
