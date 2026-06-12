package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type RequiredTestCoverageKind string
type RequiredUnitTestCoverageID string
type RequiredIntegrationTestCoverageID string
type RequiredInvariantTestCoverageID string
type RequiredSimulationTestCoverageID string
type RequiredPerformanceTestCoverageID string

const (
	TestCoverageKindUnit		RequiredTestCoverageKind	= "unit"
	TestCoverageKindIntegration	RequiredTestCoverageKind	= "integration"
	TestCoverageKindInvariant	RequiredTestCoverageKind	= "invariant"
	TestCoverageKindSimulation	RequiredTestCoverageKind	= "simulation"
	TestCoverageKindPerformance	RequiredTestCoverageKind	= "performance"

	UnitCoverageMessageIDDerivation			RequiredUnitTestCoverageID	= "message-id-derivation"
	UnitCoverageMessageNonceValidation		RequiredUnitTestCoverageID	= "message-nonce-validation"
	UnitCoverageFIFOOrdering			RequiredUnitTestCoverageID	= "fifo-ordering"
	UnitCoverageBounceHandling			RequiredUnitTestCoverageID	= "bounce-handling"
	UnitCoverageExpiryHandling			RequiredUnitTestCoverageID	= "expiry-handling"
	UnitCoverageZoneDescriptorValidation		RequiredUnitTestCoverageID	= "zone-descriptor-validation"
	UnitCoverageServiceDescriptorValidation		RequiredUnitTestCoverageID	= "service-descriptor-validation"
	UnitCoverageStorageObjectHashValidation		RequiredUnitTestCoverageID	= "storage-object-hash-validation"
	UnitCoverageIdentityResolverOutputValidation	RequiredUnitTestCoverageID	= "identity-resolver-output-validation"
	UnitCoveragePaymentConditionValidation		RequiredUnitTestCoverageID	= "payment-condition-validation"
	UnitCoverageRoutingCostCalculation		RequiredUnitTestCoverageID	= "routing-cost-calculation"
	UnitCoverageRootEncoding			RequiredUnitTestCoverageID	= "root-encoding"

	IntegrationCoverageDefaultZoneMigration			RequiredIntegrationTestCoverageID	= "default-zone-migration"
	IntegrationCoverageFinancialZoneTransferViaMessage	RequiredIntegrationTestCoverageID	= "financial-zone-transfer-via-message"
	IntegrationCoverageIdentityZoneResolverUpdate		RequiredIntegrationTestCoverageID	= "identity-zone-resolver-update"
	IntegrationCoverageServiceRegistrationLookupProof	RequiredIntegrationTestCoverageID	= "service-registration-and-lookup-proof"
	IntegrationCoverageStorageObjectRegistrationChunkProof	RequiredIntegrationTestCoverageID	= "storage-object-registration-and-chunk-proof"
	IntegrationCoverageCrossZoneIdentityBoundPayment	RequiredIntegrationTestCoverageID	= "cross-zone-identity-bound-payment"
	IntegrationCoverageContractOutboundMessageFinancial	RequiredIntegrationTestCoverageID	= "contract-outbound-message-to-financial-zone"
	IntegrationCoverageApplicationSchedulerToContract	RequiredIntegrationTestCoverageID	= "application-scheduler-emits-message-to-contract-zone"

	InvariantCoverageGlobalRootIncludesEnabledZones		RequiredInvariantTestCoverageID	= "global-root-includes-enabled-zone-roots"
	InvariantCoverageZoneRootIncludesLocalModules		RequiredInvariantTestCoverageID	= "zone-root-includes-local-module-roots"
	InvariantCoverageConsumedMessageHasOneReceipt		RequiredInvariantTestCoverageID	= "consumed-message-has-one-receipt"
	InvariantCoverageMessageValueConserved			RequiredInvariantTestCoverageID	= "message-value-conserved-across-bounce-settlement"
	InvariantCoverageReplayTombstonesRejectDuplicates	RequiredInvariantTestCoverageID	= "replay-tombstones-reject-duplicates"
	InvariantCoverageActiveIdentityBindingValidOwner	RequiredInvariantTestCoverageID	= "active-identity-binding-valid-owner"
	InvariantCoverageStorageObjectSizeEqualsChunkSum	RequiredInvariantTestCoverageID	= "storage-object-size-equals-chunk-sum"
	InvariantCoveragePaymentSettlementWithinEscrow		RequiredInvariantTestCoverageID	= "payment-settlement-within-escrow"
	InvariantCoverageServiceInterfaceHashMatchesDesc	RequiredInvariantTestCoverageID	= "service-interface-hash-matches-descriptor"

	SimulationCoverageHighVolumePerSenderQueues		RequiredSimulationTestCoverageID	= "high-volume-per-sender-queues"
	SimulationCoverageCrossZoneCongestion			RequiredSimulationTestCoverageID	= "cross-zone-congestion"
	SimulationCoverageRoutingTableEpochChanges		RequiredSimulationTestCoverageID	= "routing-table-epoch-changes"
	SimulationCoverageServiceLookupCacheExpiry		RequiredSimulationTestCoverageID	= "service-lookup-cache-expiry"
	SimulationCoverageStorageLazyFetchReceiptLoad		RequiredSimulationTestCoverageID	= "storage-lazy-fetch-receipt-load"
	SimulationCoveragePaymentConditionTimeout		RequiredSimulationTestCoverageID	= "payment-condition-timeout"
	SimulationCoverageIdentityResolverChurn			RequiredSimulationTestCoverageID	= "identity-resolver-churn"
	SimulationCoverageMixedZoneExecutionUnderBlockSTM	RequiredSimulationTestCoverageID	= "mixed-zone-execution-under-blockstm"

	PerformanceCoverageMessageEnqueueThroughput		RequiredPerformanceTestCoverageID	= "message-enqueue-throughput"
	PerformanceCoverageMessageDequeueThroughput		RequiredPerformanceTestCoverageID	= "message-dequeue-throughput"
	PerformanceCoverageReceiptProofGenerationLatency	RequiredPerformanceTestCoverageID	= "receipt-proof-generation-latency"
	PerformanceCoverageServiceLookupLatency			RequiredPerformanceTestCoverageID	= "service-lookup-latency"
	PerformanceCoverageIdentityResolutionLatency		RequiredPerformanceTestCoverageID	= "identity-resolution-latency"
	PerformanceCoverageStorageProofGenerationLatency	RequiredPerformanceTestCoverageID	= "storage-proof-generation-latency"
	PerformanceCoveragePaymentSettlementThroughput		RequiredPerformanceTestCoverageID	= "payment-settlement-throughput"
	PerformanceCoverageRootAggregationCostPerZone		RequiredPerformanceTestCoverageID	= "root-aggregation-cost-per-zone"
	PerformanceCoverageExportImportTime			RequiredPerformanceTestCoverageID	= "export-import-time"
)

type RequiredTestCoverageSpec struct {
	ID		string
	Kind		RequiredTestCoverageKind
	ModuleName	CosmosSDKModuleName
	PhaseID		ImplementationRoadmapPhaseID
	CoverageTarget	string
	Assertions	[]string
	SpecHash	string
}

type RequiredTestCoverageManifest struct {
	UnitTests		[]RequiredTestCoverageSpec
	IntegrationTests	[]RequiredTestCoverageSpec
	InvariantTests		[]RequiredTestCoverageSpec
	SimulationTests		[]RequiredTestCoverageSpec
	PerformanceTests	[]RequiredTestCoverageSpec
	ManifestHash		string
}

func DefaultRequiredTestCoverageManifest() (RequiredTestCoverageManifest, error) {
	manifest := RequiredTestCoverageManifest{
		UnitTests: []RequiredTestCoverageSpec{
			testCoverageSpec(string(UnitCoverageMessageIDDerivation), TestCoverageKindUnit, CosmosModuleMessages, RoadmapPhaseCrossZoneMessages, "canonical message id derivation", "same canonical envelope derives identical id", "payload hash chain id sender nonce and source zone affect id"),
			testCoverageSpec(string(UnitCoverageMessageNonceValidation), TestCoverageKindUnit, CosmosModuleMessages, RoadmapPhaseCrossZoneMessages, "message nonce validation", "same sender nonce cannot be reused", "nonce scope includes source zone and sender"),
			testCoverageSpec(string(UnitCoverageFIFOOrdering), TestCoverageKindUnit, CosmosModuleMessages, RoadmapPhaseCrossZoneMessages, "FIFO per-sender queue ordering", "sender queue drains in sequence order", "different senders remain deterministic after canonical sorting"),
			testCoverageSpec(string(UnitCoverageBounceHandling), TestCoverageKindUnit, CosmosModuleMessages, RoadmapPhaseCrossZoneMessages, "message bounce handling", "failed message produces bounce metadata", "remaining value returns through message semantics"),
			testCoverageSpec(string(UnitCoverageExpiryHandling), TestCoverageKindUnit, CosmosModuleMessages, RoadmapPhaseCrossZoneMessages, "message expiry handling", "expired messages are skipped", "expiry creates receipt and replay tombstone"),
			testCoverageSpec(string(UnitCoverageZoneDescriptorValidation), TestCoverageKindUnit, CosmosModuleZones, RoadmapPhaseCanonicalZones, "zone descriptor validation", "zone id type scope and root prefix are validated", "invalid zone capabilities are rejected"),
			testCoverageSpec(string(UnitCoverageServiceDescriptorValidation), TestCoverageKindUnit, CosmosModuleServices, RoadmapPhaseServiceStorageRouting, "service descriptor validation", "service owner endpoint interface and ttl are validated", "descriptor hash changes when interface changes"),
			testCoverageSpec(string(UnitCoverageStorageObjectHashValidation), TestCoverageKindUnit, CosmosModuleStorage, RoadmapPhaseServiceStorageRouting, "storage object hash validation", "content hash commits to ordered chunk roots", "object size equals chunk sizes"),
			testCoverageSpec(string(UnitCoverageIdentityResolverOutputValidation), TestCoverageKindUnit, CosmosModuleIdentity, RoadmapPhaseIdentityPaymentIntegration, "identity resolver output validation", "account zone service contract and composite outputs validate", "invalid resolver output hashes are rejected"),
			testCoverageSpec(string(UnitCoveragePaymentConditionValidation), TestCoverageKindUnit, CosmosModulePayments, RoadmapPhaseIdentityPaymentIntegration, "payment condition validation", "condition hash and expiry are validated", "invalid conditional transfer proofs are rejected"),
			testCoverageSpec(string(UnitCoverageRoutingCostCalculation), TestCoverageKindUnit, CosmosModuleRouting, RoadmapPhaseServiceStorageRouting, "routing cost calculation", "integer weights produce deterministic cost", "route tie breaks lexicographically"),
			testCoverageSpec(string(UnitCoverageRootEncoding), TestCoverageKindUnit, CosmosModuleAetraCore, RoadmapPhaseKernelRootModel, "root encoding", "root order is canonical", "tampered root contribution changes global root"),
		},
		IntegrationTests: []RequiredTestCoverageSpec{
			testCoverageSpec(string(IntegrationCoverageDefaultZoneMigration), TestCoverageKindIntegration, CosmosModuleZones, RoadmapPhaseKernelRootModel, "default zone migration", "existing chain state runs as default zone", "export import preserves root metadata"),
			testCoverageSpec(string(IntegrationCoverageFinancialZoneTransferViaMessage), TestCoverageKindIntegration, CosmosModuleMessages, RoadmapPhaseCanonicalZones, "Financial Zone transfer via message", "message ingress applies transfer in Financial Zone", "receipt proves transfer execution"),
			testCoverageSpec(string(IntegrationCoverageIdentityZoneResolverUpdate), TestCoverageKindIntegration, CosmosModuleIdentity, RoadmapPhaseCanonicalZones, "Identity Zone resolver update", "resolver update mutates identity namespace only", "identity root changes deterministically"),
			testCoverageSpec(string(IntegrationCoverageServiceRegistrationLookupProof), TestCoverageKindIntegration, CosmosModuleServices, RoadmapPhaseServiceStorageRouting, "service registration and lookup proof", "registered service is found by deterministic index", "lookup response includes proof root"),
			testCoverageSpec(string(IntegrationCoverageStorageObjectRegistrationChunkProof), TestCoverageKindIntegration, CosmosModuleStorage, RoadmapPhaseServiceStorageRouting, "storage object registration and chunk proof", "object registration commits chunk roots", "chunk inclusion proof verifies object root"),
			testCoverageSpec(string(IntegrationCoverageCrossZoneIdentityBoundPayment), TestCoverageKindIntegration, CosmosModulePayments, RoadmapPhaseIdentityPaymentIntegration, "cross-zone identity-bound payment", "identity binding resolves payment recipient", "payment settles through Financial Zone"),
			testCoverageSpec(string(IntegrationCoverageContractOutboundMessageFinancial), TestCoverageKindIntegration, CosmosModuleContracts, RoadmapPhaseVMRuntime, "contract outbound message to Financial Zone", "contract emits outbound settlement message", "contract cannot directly mutate financial state"),
			testCoverageSpec(string(IntegrationCoverageApplicationSchedulerToContract), TestCoverageKindIntegration, CosmosModuleZones, RoadmapPhaseCanonicalZones, "Application scheduler emits message to Contract Zone", "scheduler emits ordered contract-zone message", "contract zone receipt is queryable"),
		},
		InvariantTests: []RequiredTestCoverageSpec{
			testCoverageSpec(string(InvariantCoverageGlobalRootIncludesEnabledZones), TestCoverageKindInvariant, CosmosModuleAetraCore, RoadmapPhaseKernelRootModel, "global root includes all enabled zone roots", "disabled zones are excluded from aggregation", "each enabled zone commitment participates in global root"),
			testCoverageSpec(string(InvariantCoverageZoneRootIncludesLocalModules), TestCoverageKindInvariant, CosmosModuleZones, RoadmapPhaseCanonicalZones, "zone root includes local module roots", "local module roots are ordered by module id", "tampered module root changes zone root"),
			testCoverageSpec(string(InvariantCoverageConsumedMessageHasOneReceipt), TestCoverageKindInvariant, CosmosModuleMessages, RoadmapPhaseCrossZoneMessages, "every consumed message has one receipt", "consumed message id maps to exactly one receipt", "missing and duplicate receipts are rejected"),
			testCoverageSpec(string(InvariantCoverageMessageValueConserved), TestCoverageKindInvariant, CosmosModulePayments, RoadmapPhaseIdentityPaymentIntegration, "message value is conserved across bounce and settlement", "bounced value plus charged fee equals source debit", "settled value equals destination credit plus fee"),
			testCoverageSpec(string(InvariantCoverageReplayTombstonesRejectDuplicates), TestCoverageKindInvariant, CosmosModuleMessages, RoadmapPhaseCrossZoneMessages, "replay tombstones reject duplicate messages", "duplicate message id is rejected after tombstone creation", "proof horizon preserves tombstone lookup"),
			testCoverageSpec(string(InvariantCoverageActiveIdentityBindingValidOwner), TestCoverageKindInvariant, CosmosModuleIdentity, RoadmapPhaseIdentityPaymentIntegration, "active identity binding has valid owner", "binding owner matches active identity record", "expired binding is unavailable for routing"),
			testCoverageSpec(string(InvariantCoverageStorageObjectSizeEqualsChunkSum), TestCoverageKindInvariant, CosmosModuleStorage, RoadmapPhaseServiceStorageRouting, "storage object size equals chunk sum", "registered object size equals ordered chunk descriptor sizes", "chunk sum overflow is rejected"),
			testCoverageSpec(string(InvariantCoveragePaymentSettlementWithinEscrow), TestCoverageKindInvariant, CosmosModulePayments, RoadmapPhaseIdentityPaymentIntegration, "payment settlement cannot exceed escrow", "settlement amount is bounded by escrow balance", "partial settlements preserve remaining escrow"),
			testCoverageSpec(string(InvariantCoverageServiceInterfaceHashMatchesDesc), TestCoverageKindInvariant, CosmosModuleServices, RoadmapPhaseServiceStorageRouting, "service interface hash matches descriptor", "interface descriptor bytes hash to advertised hash", "descriptor mismatch blocks registration"),
		},
		SimulationTests: []RequiredTestCoverageSpec{
			testCoverageSpec(string(SimulationCoverageHighVolumePerSenderQueues), TestCoverageKindSimulation, CosmosModuleMessages, RoadmapPhasePerformanceHardening, "high-volume per-sender queues", "large sender queues drain in nonce order", "bounded drain limit preserves deterministic backlog"),
			testCoverageSpec(string(SimulationCoverageCrossZoneCongestion), TestCoverageKindSimulation, CosmosModuleRouting, RoadmapPhasePerformanceHardening, "cross-zone congestion", "committed congestion snapshot affects route cost", "congested routes are deprioritized deterministically"),
			testCoverageSpec(string(SimulationCoverageRoutingTableEpochChanges), TestCoverageKindSimulation, CosmosModuleRouting, RoadmapPhasePerformanceHardening, "routing table epoch changes", "epoch transition preserves proofable old routes", "new epoch route selection is deterministic"),
			testCoverageSpec(string(SimulationCoverageServiceLookupCacheExpiry), TestCoverageKindSimulation, CosmosModuleServices, RoadmapPhasePerformanceHardening, "service lookup cache expiry", "expired cache entries are ignored", "fresh lookup result uses committed service index"),
			testCoverageSpec(string(SimulationCoverageStorageLazyFetchReceiptLoad), TestCoverageKindSimulation, CosmosModuleStorage, RoadmapPhasePerformanceHardening, "storage lazy fetch receipt load", "lazy fetch receipts remain bounded per block", "receipt root is stable under high load"),
			testCoverageSpec(string(SimulationCoveragePaymentConditionTimeout), TestCoverageKindSimulation, CosmosModulePayments, RoadmapPhasePerformanceHardening, "payment condition timeout", "expired conditions cannot settle", "timeout receipts release escrow deterministically"),
			testCoverageSpec(string(SimulationCoverageIdentityResolverChurn), TestCoverageKindSimulation, CosmosModuleIdentity, RoadmapPhasePerformanceHardening, "identity resolver churn", "rapid resolver updates produce deterministic identity root", "stale resolver proof is rejected after update"),
			testCoverageSpec(string(SimulationCoverageMixedZoneExecutionUnderBlockSTM), TestCoverageKindSimulation, CosmosModuleZones, RoadmapPhasePerformanceHardening, "mixed zone execution under BlockSTM", "independent zone workloads keep identical roots with BlockSTM grouping", "conflicting workloads fall back to deterministic serial order"),
		},
		PerformanceTests: []RequiredTestCoverageSpec{
			testCoverageSpec(string(PerformanceCoverageMessageEnqueueThroughput), TestCoverageKindPerformance, CosmosModuleMessages, RoadmapPhasePerformanceHardening, "message enqueue throughput", "enqueue benchmark reports messages per second", "enqueue remains bounded by configured gas and queue limits"),
			testCoverageSpec(string(PerformanceCoverageMessageDequeueThroughput), TestCoverageKindPerformance, CosmosModuleMessages, RoadmapPhasePerformanceHardening, "message dequeue throughput", "dequeue benchmark reports receipts per block", "dequeue respects bounded per-sender draining"),
			testCoverageSpec(string(PerformanceCoverageReceiptProofGenerationLatency), TestCoverageKindPerformance, CosmosModuleMessages, RoadmapPhasePerformanceHardening, "receipt proof generation latency", "receipt proof benchmark records p50 and p99 latency", "proof generation uses finalized receipt root"),
			testCoverageSpec(string(PerformanceCoverageServiceLookupLatency), TestCoverageKindPerformance, CosmosModuleServices, RoadmapPhasePerformanceHardening, "service lookup latency", "service lookup benchmark records deterministic index latency", "lookup latency includes proof attachment cost"),
			testCoverageSpec(string(PerformanceCoverageIdentityResolutionLatency), TestCoverageKindPerformance, CosmosModuleIdentity, RoadmapPhasePerformanceHardening, "identity resolution latency", "resolver benchmark records graph traversal latency", "resolution remains bounded by configured graph depth"),
			testCoverageSpec(string(PerformanceCoverageStorageProofGenerationLatency), TestCoverageKindPerformance, CosmosModuleStorage, RoadmapPhasePerformanceHardening, "storage proof generation latency", "storage proof benchmark records chunk proof latency", "proof generation uses committed object root"),
			testCoverageSpec(string(PerformanceCoveragePaymentSettlementThroughput), TestCoverageKindPerformance, CosmosModulePayments, RoadmapPhasePerformanceHardening, "payment settlement throughput", "settlement benchmark reports completed payments per block", "settlement throughput respects escrow and dispute constraints"),
			testCoverageSpec(string(PerformanceCoverageRootAggregationCostPerZone), TestCoverageKindPerformance, CosmosModuleAetraCore, RoadmapPhasePerformanceHardening, "root aggregation cost per zone", "aggregation benchmark records cost per enabled zone", "cost scales with enabled zone count and root size"),
			testCoverageSpec(string(PerformanceCoverageExportImportTime), TestCoverageKindPerformance, CosmosModuleAetraCore, RoadmapPhasePerformanceHardening, "export import time", "export benchmark records manifest generation time", "import benchmark verifies reproducible root metadata"),
		},
	}
	return NewRequiredTestCoverageManifest(manifest.UnitTests, manifest.IntegrationTests, manifest.InvariantTests, manifest.SimulationTests, manifest.PerformanceTests)
}

func NewRequiredTestCoverageManifest(unitTests, integrationTests, invariantTests, simulationTests, performanceTests []RequiredTestCoverageSpec) (RequiredTestCoverageManifest, error) {
	manifest := RequiredTestCoverageManifest{
		UnitTests:		normalizeTestCoverageSpecs(unitTests, TestCoverageKindUnit),
		IntegrationTests:	normalizeTestCoverageSpecs(integrationTests, TestCoverageKindIntegration),
		InvariantTests:		normalizeTestCoverageSpecs(invariantTests, TestCoverageKindInvariant),
		SimulationTests:	normalizeTestCoverageSpecs(simulationTests, TestCoverageKindSimulation),
		PerformanceTests:	normalizeTestCoverageSpecs(performanceTests, TestCoverageKindPerformance),
	}
	if err := manifest.ValidateFormat(); err != nil {
		return RequiredTestCoverageManifest{}, err
	}
	for i := range manifest.UnitTests {
		manifest.UnitTests[i].SpecHash = ComputeRequiredTestCoverageSpecHash(manifest.UnitTests[i])
	}
	for i := range manifest.IntegrationTests {
		manifest.IntegrationTests[i].SpecHash = ComputeRequiredTestCoverageSpecHash(manifest.IntegrationTests[i])
	}
	for i := range manifest.InvariantTests {
		manifest.InvariantTests[i].SpecHash = ComputeRequiredTestCoverageSpecHash(manifest.InvariantTests[i])
	}
	for i := range manifest.SimulationTests {
		manifest.SimulationTests[i].SpecHash = ComputeRequiredTestCoverageSpecHash(manifest.SimulationTests[i])
	}
	for i := range manifest.PerformanceTests {
		manifest.PerformanceTests[i].SpecHash = ComputeRequiredTestCoverageSpecHash(manifest.PerformanceTests[i])
	}
	manifest.ManifestHash = ComputeRequiredTestCoverageManifestHash(manifest)
	return manifest, manifest.Validate()
}

func RequiredUnitTestCoverageIDs() []RequiredUnitTestCoverageID {
	return []RequiredUnitTestCoverageID{
		UnitCoverageMessageIDDerivation,
		UnitCoverageMessageNonceValidation,
		UnitCoverageFIFOOrdering,
		UnitCoverageBounceHandling,
		UnitCoverageExpiryHandling,
		UnitCoverageZoneDescriptorValidation,
		UnitCoverageServiceDescriptorValidation,
		UnitCoverageStorageObjectHashValidation,
		UnitCoverageIdentityResolverOutputValidation,
		UnitCoveragePaymentConditionValidation,
		UnitCoverageRoutingCostCalculation,
		UnitCoverageRootEncoding,
	}
}

func RequiredIntegrationTestCoverageIDs() []RequiredIntegrationTestCoverageID {
	return []RequiredIntegrationTestCoverageID{
		IntegrationCoverageDefaultZoneMigration,
		IntegrationCoverageFinancialZoneTransferViaMessage,
		IntegrationCoverageIdentityZoneResolverUpdate,
		IntegrationCoverageServiceRegistrationLookupProof,
		IntegrationCoverageStorageObjectRegistrationChunkProof,
		IntegrationCoverageCrossZoneIdentityBoundPayment,
		IntegrationCoverageContractOutboundMessageFinancial,
		IntegrationCoverageApplicationSchedulerToContract,
	}
}

func RequiredInvariantTestCoverageIDs() []RequiredInvariantTestCoverageID {
	return []RequiredInvariantTestCoverageID{
		InvariantCoverageGlobalRootIncludesEnabledZones,
		InvariantCoverageZoneRootIncludesLocalModules,
		InvariantCoverageConsumedMessageHasOneReceipt,
		InvariantCoverageMessageValueConserved,
		InvariantCoverageReplayTombstonesRejectDuplicates,
		InvariantCoverageActiveIdentityBindingValidOwner,
		InvariantCoverageStorageObjectSizeEqualsChunkSum,
		InvariantCoveragePaymentSettlementWithinEscrow,
		InvariantCoverageServiceInterfaceHashMatchesDesc,
	}
}

func RequiredSimulationTestCoverageIDs() []RequiredSimulationTestCoverageID {
	return []RequiredSimulationTestCoverageID{
		SimulationCoverageHighVolumePerSenderQueues,
		SimulationCoverageCrossZoneCongestion,
		SimulationCoverageRoutingTableEpochChanges,
		SimulationCoverageServiceLookupCacheExpiry,
		SimulationCoverageStorageLazyFetchReceiptLoad,
		SimulationCoveragePaymentConditionTimeout,
		SimulationCoverageIdentityResolverChurn,
		SimulationCoverageMixedZoneExecutionUnderBlockSTM,
	}
}

func RequiredPerformanceTestCoverageIDs() []RequiredPerformanceTestCoverageID {
	return []RequiredPerformanceTestCoverageID{
		PerformanceCoverageMessageEnqueueThroughput,
		PerformanceCoverageMessageDequeueThroughput,
		PerformanceCoverageReceiptProofGenerationLatency,
		PerformanceCoverageServiceLookupLatency,
		PerformanceCoverageIdentityResolutionLatency,
		PerformanceCoverageStorageProofGenerationLatency,
		PerformanceCoveragePaymentSettlementThroughput,
		PerformanceCoverageRootAggregationCostPerZone,
		PerformanceCoverageExportImportTime,
	}
}

func (manifest RequiredTestCoverageManifest) ValidateFormat() error {
	manifest.UnitTests = normalizeTestCoverageSpecs(manifest.UnitTests, TestCoverageKindUnit)
	manifest.IntegrationTests = normalizeTestCoverageSpecs(manifest.IntegrationTests, TestCoverageKindIntegration)
	manifest.InvariantTests = normalizeTestCoverageSpecs(manifest.InvariantTests, TestCoverageKindInvariant)
	manifest.SimulationTests = normalizeTestCoverageSpecs(manifest.SimulationTests, TestCoverageKindSimulation)
	manifest.PerformanceTests = normalizeTestCoverageSpecs(manifest.PerformanceTests, TestCoverageKindPerformance)
	if err := validateCoverageSpecSet("aetracore unit test coverage", manifest.UnitTests, requiredUnitCoverageIDStrings(), TestCoverageKindUnit); err != nil {
		return err
	}
	if err := validateCoverageSpecSet("aetracore integration test coverage", manifest.IntegrationTests, requiredIntegrationCoverageIDStrings(), TestCoverageKindIntegration); err != nil {
		return err
	}
	if err := validateCoverageSpecSet("aetracore invariant test coverage", manifest.InvariantTests, requiredInvariantCoverageIDStrings(), TestCoverageKindInvariant); err != nil {
		return err
	}
	if err := validateCoverageSpecSet("aetracore simulation test coverage", manifest.SimulationTests, requiredSimulationCoverageIDStrings(), TestCoverageKindSimulation); err != nil {
		return err
	}
	if err := validateCoverageSpecSet("aetracore performance test coverage", manifest.PerformanceTests, requiredPerformanceCoverageIDStrings(), TestCoverageKindPerformance); err != nil {
		return err
	}
	if manifest.ManifestHash != "" {
		return ValidateHash("aetracore required test coverage manifest hash", manifest.ManifestHash)
	}
	return nil
}

func (manifest RequiredTestCoverageManifest) Validate() error {
	manifest.UnitTests = normalizeTestCoverageSpecs(manifest.UnitTests, TestCoverageKindUnit)
	manifest.IntegrationTests = normalizeTestCoverageSpecs(manifest.IntegrationTests, TestCoverageKindIntegration)
	manifest.InvariantTests = normalizeTestCoverageSpecs(manifest.InvariantTests, TestCoverageKindInvariant)
	manifest.SimulationTests = normalizeTestCoverageSpecs(manifest.SimulationTests, TestCoverageKindSimulation)
	manifest.PerformanceTests = normalizeTestCoverageSpecs(manifest.PerformanceTests, TestCoverageKindPerformance)
	if err := manifest.ValidateFormat(); err != nil {
		return err
	}
	for _, spec := range manifest.UnitTests {
		if err := spec.Validate(); err != nil {
			return err
		}
	}
	for _, spec := range manifest.IntegrationTests {
		if err := spec.Validate(); err != nil {
			return err
		}
	}
	for _, spec := range manifest.InvariantTests {
		if err := spec.Validate(); err != nil {
			return err
		}
	}
	for _, spec := range manifest.SimulationTests {
		if err := spec.Validate(); err != nil {
			return err
		}
	}
	for _, spec := range manifest.PerformanceTests {
		if err := spec.Validate(); err != nil {
			return err
		}
	}
	if manifest.ManifestHash != ComputeRequiredTestCoverageManifestHash(manifest) {
		return errors.New("aetracore required test coverage manifest hash mismatch")
	}
	return nil
}

func (spec RequiredTestCoverageSpec) ValidateFormat() error {
	spec = normalizeTestCoverageSpec(spec, spec.Kind)
	if spec.Kind != TestCoverageKindUnit && spec.Kind != TestCoverageKindIntegration && spec.Kind != TestCoverageKindInvariant && spec.Kind != TestCoverageKindSimulation && spec.Kind != TestCoverageKindPerformance {
		return fmt.Errorf("unknown aetracore required test coverage kind %q", spec.Kind)
	}
	if err := validatePolicyID("aetracore required test coverage id", spec.ID); err != nil {
		return err
	}
	if !IsRequiredCosmosSDKModule(spec.ModuleName) {
		return fmt.Errorf("aetracore required test coverage %s references unknown module %s", spec.ID, spec.ModuleName)
	}
	if !IsImplementationRoadmapPhaseID(spec.PhaseID) {
		return fmt.Errorf("aetracore required test coverage %s references unknown roadmap phase %s", spec.ID, spec.PhaseID)
	}
	if err := validateRoadmapText("aetracore required test coverage target", spec.CoverageTarget); err != nil {
		return err
	}
	if err := validateCoverageAssertions(spec.ID, spec.Assertions); err != nil {
		return err
	}
	if spec.SpecHash != "" {
		return ValidateHash("aetracore required test coverage spec hash", spec.SpecHash)
	}
	return nil
}

func (spec RequiredTestCoverageSpec) Validate() error {
	spec = normalizeTestCoverageSpec(spec, spec.Kind)
	if err := spec.ValidateFormat(); err != nil {
		return err
	}
	if spec.SpecHash != ComputeRequiredTestCoverageSpecHash(spec) {
		return errors.New("aetracore required test coverage spec hash mismatch")
	}
	return nil
}

func RequiredUnitTestCoverageByID(manifest RequiredTestCoverageManifest, id RequiredUnitTestCoverageID) (RequiredTestCoverageSpec, bool) {
	return requiredTestCoverageByID(manifest.UnitTests, string(id))
}

func RequiredIntegrationTestCoverageByID(manifest RequiredTestCoverageManifest, id RequiredIntegrationTestCoverageID) (RequiredTestCoverageSpec, bool) {
	return requiredTestCoverageByID(manifest.IntegrationTests, string(id))
}

func RequiredInvariantTestCoverageByID(manifest RequiredTestCoverageManifest, id RequiredInvariantTestCoverageID) (RequiredTestCoverageSpec, bool) {
	return requiredTestCoverageByID(manifest.InvariantTests, string(id))
}

func RequiredSimulationTestCoverageByID(manifest RequiredTestCoverageManifest, id RequiredSimulationTestCoverageID) (RequiredTestCoverageSpec, bool) {
	return requiredTestCoverageByID(manifest.SimulationTests, string(id))
}

func RequiredPerformanceTestCoverageByID(manifest RequiredTestCoverageManifest, id RequiredPerformanceTestCoverageID) (RequiredTestCoverageSpec, bool) {
	return requiredTestCoverageByID(manifest.PerformanceTests, string(id))
}

func ComputeRequiredTestCoverageSpecHash(spec RequiredTestCoverageSpec) string {
	spec = normalizeTestCoverageSpec(spec, spec.Kind)
	return hashRoot("aetra-aek-required-test-coverage-spec-v1", func(w byteWriter) {
		writePart(w, spec.ID)
		writePart(w, string(spec.Kind))
		writePart(w, string(spec.ModuleName))
		writePart(w, string(spec.PhaseID))
		writePart(w, spec.CoverageTarget)
		writeStringParts(w, spec.Assertions)
	})
}

func ComputeRequiredTestCoverageManifestHash(manifest RequiredTestCoverageManifest) string {
	manifest.UnitTests = normalizeTestCoverageSpecs(manifest.UnitTests, TestCoverageKindUnit)
	manifest.IntegrationTests = normalizeTestCoverageSpecs(manifest.IntegrationTests, TestCoverageKindIntegration)
	manifest.InvariantTests = normalizeTestCoverageSpecs(manifest.InvariantTests, TestCoverageKindInvariant)
	manifest.SimulationTests = normalizeTestCoverageSpecs(manifest.SimulationTests, TestCoverageKindSimulation)
	manifest.PerformanceTests = normalizeTestCoverageSpecs(manifest.PerformanceTests, TestCoverageKindPerformance)
	return hashRoot("aetra-aek-required-test-coverage-manifest-v3", func(w byteWriter) {
		writeUint64(w, uint64(len(manifest.UnitTests)))
		for _, spec := range manifest.UnitTests {
			writePart(w, spec.SpecHash)
		}
		writeUint64(w, uint64(len(manifest.IntegrationTests)))
		for _, spec := range manifest.IntegrationTests {
			writePart(w, spec.SpecHash)
		}
		writeUint64(w, uint64(len(manifest.InvariantTests)))
		for _, spec := range manifest.InvariantTests {
			writePart(w, spec.SpecHash)
		}
		writeUint64(w, uint64(len(manifest.SimulationTests)))
		for _, spec := range manifest.SimulationTests {
			writePart(w, spec.SpecHash)
		}
		writeUint64(w, uint64(len(manifest.PerformanceTests)))
		for _, spec := range manifest.PerformanceTests {
			writePart(w, spec.SpecHash)
		}
	})
}

func testCoverageSpec(id string, kind RequiredTestCoverageKind, moduleName CosmosSDKModuleName, phaseID ImplementationRoadmapPhaseID, target string, assertions ...string) RequiredTestCoverageSpec {
	return RequiredTestCoverageSpec{
		ID:		id,
		Kind:		kind,
		ModuleName:	moduleName,
		PhaseID:	phaseID,
		CoverageTarget:	target,
		Assertions:	assertions,
	}
}

func normalizeTestCoverageSpecs(specs []RequiredTestCoverageSpec, kind RequiredTestCoverageKind) []RequiredTestCoverageSpec {
	out := make([]RequiredTestCoverageSpec, len(specs))
	for i, spec := range specs {
		spec = normalizeTestCoverageSpec(spec, kind)
		if spec.SpecHash == "" {
			spec.SpecHash = ComputeRequiredTestCoverageSpecHash(spec)
		}
		out[i] = spec
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out
}

func normalizeTestCoverageSpec(spec RequiredTestCoverageSpec, defaultKind RequiredTestCoverageKind) RequiredTestCoverageSpec {
	spec.ID = strings.TrimSpace(spec.ID)
	if spec.Kind == "" {
		spec.Kind = defaultKind
	}
	spec.Kind = RequiredTestCoverageKind(strings.TrimSpace(string(spec.Kind)))
	spec.ModuleName = CosmosSDKModuleName(strings.TrimSpace(string(spec.ModuleName)))
	spec.PhaseID = ImplementationRoadmapPhaseID(strings.TrimSpace(string(spec.PhaseID)))
	spec.CoverageTarget = strings.TrimSpace(spec.CoverageTarget)
	spec.Assertions = normalizeRoadmapStringSet(spec.Assertions)
	spec.SpecHash = strings.ToLower(strings.TrimSpace(spec.SpecHash))
	return spec
}

func validateCoverageSpecSet(field string, specs []RequiredTestCoverageSpec, required []string, kind RequiredTestCoverageKind) error {
	if len(specs) != len(required) {
		return fmt.Errorf("%s must include %d required coverage areas", field, len(required))
	}
	seen := make(map[string]struct{}, len(specs))
	var previous string
	for i, spec := range specs {
		if spec.Kind != kind {
			return fmt.Errorf("%s %s must have kind %s", field, spec.ID, kind)
		}
		if err := spec.ValidateFormat(); err != nil {
			return err
		}
		if _, found := seen[spec.ID]; found {
			return fmt.Errorf("duplicate %s %s", field, spec.ID)
		}
		seen[spec.ID] = struct{}{}
		if i > 0 && previous >= spec.ID {
			return fmt.Errorf("%s must be sorted canonically", field)
		}
		previous = spec.ID
	}
	for _, id := range required {
		if _, found := seen[id]; !found {
			return fmt.Errorf("missing %s %s", field, id)
		}
	}
	return nil
}

func validateCoverageAssertions(id string, assertions []string) error {
	if len(assertions) == 0 {
		return fmt.Errorf("aetracore required test coverage %s assertions are required", id)
	}
	seen := make(map[string]struct{}, len(assertions))
	var previous string
	for i, assertion := range assertions {
		if err := validateRoadmapText("aetracore required test coverage assertion", assertion); err != nil {
			return err
		}
		if _, found := seen[assertion]; found {
			return fmt.Errorf("duplicate aetracore required test coverage assertion %s", assertion)
		}
		seen[assertion] = struct{}{}
		if i > 0 && previous >= assertion {
			return errors.New("aetracore required test coverage assertions must be sorted canonically")
		}
		previous = assertion
	}
	return nil
}

func requiredTestCoverageByID(specs []RequiredTestCoverageSpec, id string) (RequiredTestCoverageSpec, bool) {
	for _, spec := range specs {
		if spec.ID == id {
			return spec, true
		}
	}
	return RequiredTestCoverageSpec{}, false
}

func requiredUnitCoverageIDStrings() []string {
	ids := RequiredUnitTestCoverageIDs()
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = string(id)
	}
	sort.Strings(out)
	return out
}

func requiredIntegrationCoverageIDStrings() []string {
	ids := RequiredIntegrationTestCoverageIDs()
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = string(id)
	}
	sort.Strings(out)
	return out
}

func requiredInvariantCoverageIDStrings() []string {
	ids := RequiredInvariantTestCoverageIDs()
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = string(id)
	}
	sort.Strings(out)
	return out
}

func requiredSimulationCoverageIDStrings() []string {
	ids := RequiredSimulationTestCoverageIDs()
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = string(id)
	}
	sort.Strings(out)
	return out
}

func requiredPerformanceCoverageIDStrings() []string {
	ids := RequiredPerformanceTestCoverageIDs()
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = string(id)
	}
	sort.Strings(out)
	return out
}

func writeStringParts(w byteWriter, values []string) {
	values = normalizeRoadmapStringSet(values)
	writeUint64(w, uint64(len(values)))
	for _, value := range values {
		writePart(w, value)
	}
}
