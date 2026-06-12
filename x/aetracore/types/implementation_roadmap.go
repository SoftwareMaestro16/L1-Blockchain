package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type ImplementationRoadmapPhaseID string

const (
	RoadmapPhaseBaselineAudit		ImplementationRoadmapPhaseID	= "phase-0-baseline-audit"
	RoadmapPhaseKernelRootModel		ImplementationRoadmapPhaseID	= "phase-1-kernel-root-model"
	RoadmapPhaseCrossZoneMessages		ImplementationRoadmapPhaseID	= "phase-2-cross-zone-messages"
	RoadmapPhaseCanonicalZones		ImplementationRoadmapPhaseID	= "phase-3-canonical-zones"
	RoadmapPhaseServiceStorageRouting	ImplementationRoadmapPhaseID	= "phase-4-services-storage-routing"
	RoadmapPhaseIdentityPaymentIntegration	ImplementationRoadmapPhaseID	= "phase-5-identity-payment-integration"
	RoadmapPhaseVMRuntime			ImplementationRoadmapPhaseID	= "phase-6-vm-runtime"
	RoadmapPhasePerformanceHardening	ImplementationRoadmapPhaseID	= "phase-7-performance-hardening"
)

type ImplementationRoadmap struct {
	Phases		[]ImplementationRoadmapPhase
	RoadmapHash	string
}

type ImplementationRoadmapPhase struct {
	PhaseID		ImplementationRoadmapPhaseID
	PhaseNumber	uint32
	Name		string
	Tasks		[]RoadmapChecklistItem
	ExitCriteria	[]RoadmapChecklistItem
	Evidence	RoadmapEvidence
	PhaseHash	string
}

type RoadmapChecklistItem struct {
	ID		string
	Description	string
	Complete	bool
}

type RoadmapEvidence struct {
	ModuleInventory				[]RoadmapModuleInventoryEntry
	CanonicalZones				[]RoadmapCanonicalZoneEntry
	IdentityResolverOutputs			[]string
	CrossModuleDirectWritesAudited		bool
	ExportImportTestsAdded			bool
	ModuleInvariantHarnessAdded		bool
	RootContributionInterfaceDesign		bool
	CurrentStateReproducible		bool
	ModuleBoundariesDocumented		bool
	MigrationRiskListComplete		bool
	AetraCoreModuleImplemented		bool
	ZonesModuleImplemented			bool
	ZoneRegistryImplemented			bool
	GlobalStateRootImplemented		bool
	BlockCommitmentMetadataQueries		bool
	DefaultZoneRunnable			bool
	DefaultZoneRootIncluded			bool
	ExportImportPreservesRootMeta		bool
	MessagesModuleImplemented		bool
	MessageEnvelopeAdded			bool
	FIFOPerSenderQueuesAdded		bool
	NonceReplayProtectionAdded		bool
	MessageReceiptsAdded			bool
	BounceAndExpiryAdded			bool
	MessageReceiptRootsAdded		bool
	SameChainAsyncDeterministic		bool
	MessageReceiptProofsAvailable		bool
	ReplayAttemptsRejected			bool
	FinancialZoneBoundaryMoved		bool
	IdentityZoneActivated			bool
	ApplicationZoneSchedulerBoundary	bool
	ContractZoneSkeletonAdded		bool
	ZoneSpecificQueriesRoots		bool
	FourCanonicalZonesExist			bool
	CanonicalZoneSurfacesComplete		bool
	CrossZoneMutationMessagesOnly		bool
	ServicesModuleImplemented		bool
	StorageModuleImplemented		bool
	RoutingModuleImplemented		bool
	ServiceDescriptorsAdded			bool
	StorageObjectCommitmentsAdded		bool
	NodeRecordsRoutingEpochsAdded		bool
	ProofAttachedLookupQueriesAdded		bool
	ServiceDiscoveryDeterministic		bool
	StorageCommitmentsProofVerifiable	bool
	RoutingTableCommittedQueryable		bool
	AETResolverOutputsUpgraded		bool
	IdentityGraphAdded			bool
	CrossZoneIdentityBindingAdded		bool
	PaymentsModuleImplemented		bool
	PaymentEnvelopeAdded			bool
	ConditionalTransfersAdded		bool
	FinancialZoneSettlementAdded		bool
	IdentityResolvesAllOutputTypes		bool
	PaymentsSettleThroughFinancialZone	bool
	PaymentDisputesDeterministicReplay	bool
	ContractsModuleImplemented		bool
	AVMBytecodeInterfaceAdded		bool
	CosmWasmAdapterBoundaryAdded		bool
	VMStorageAdapterAdded			bool
	VMOutboundMessageSupportAdded		bool
	ContractReceiptsProofsAdded		bool
	ContractExecutionMessageDriven		bool
	ContractsNoDirectZoneMutation		bool
	ContractStateRootProofVerifiable	bool
	BlockSTMAwareGroupingAdded		bool
	StoreV2RootReadOptimizationAdded	bool
	QueueDrainingBenchmarksAdded		bool
	ServiceLookupBenchmarksAdded		bool
	StorageProofBenchmarksAdded		bool
	RoutingSimulationTestsAdded		bool
	AdaptiveSyncRecoveryTestsAdded		bool
	IndependentZoneWorkloadsParallelize	bool
	RootGenerationBounded			bool
	NodesRecoverServeProofQueriesAfterSync	bool
}

type RoadmapModuleInventoryEntry struct {
	ModuleName	CosmosSDKModuleName
	ModulePath	string
	StoreKey	string
	StateKeys	[]string
	RootType	RootType
}

type RoadmapCanonicalZoneEntry struct {
	ZoneID		ZoneID
	StateNamespace	string
	RootType	RootType
	MessageQueue	bool
	MsgServer	bool
	QueryServer	bool
	Keeper		bool
}

func DefaultImplementationRoadmap() (ImplementationRoadmap, error) {
	manifest, err := DefaultCosmosModuleRequirementManifest()
	if err != nil {
		return ImplementationRoadmap{}, err
	}
	inventory := BuildRoadmapModuleInventory(manifest)
	phases := []ImplementationRoadmapPhase{
		roadmapPhaseZero(inventory),
		roadmapPhaseOne(inventory),
		roadmapPhaseTwo(inventory),
		roadmapPhaseThree(inventory),
		roadmapPhaseFour(inventory),
		roadmapPhaseFive(inventory),
		roadmapPhaseSix(inventory),
		roadmapPhaseSeven(inventory),
	}
	return NewImplementationRoadmap(phases)
}

func NewImplementationRoadmap(phases []ImplementationRoadmapPhase) (ImplementationRoadmap, error) {
	roadmap := ImplementationRoadmap{Phases: normalizeRoadmapPhases(phases)}
	if err := roadmap.ValidateFormat(); err != nil {
		return ImplementationRoadmap{}, err
	}
	for i := range roadmap.Phases {
		roadmap.Phases[i].PhaseHash = ComputeRoadmapPhaseHash(roadmap.Phases[i])
	}
	roadmap.RoadmapHash = ComputeImplementationRoadmapHash(roadmap)
	return roadmap, roadmap.Validate()
}

func BuildRoadmapModuleInventory(manifest CosmosModuleRequirementManifest) []RoadmapModuleInventoryEntry {
	manifest.Modules = normalizeCosmosModuleSurfaces(manifest.Modules)
	inventory := make([]RoadmapModuleInventoryEntry, 0, len(manifest.Modules))
	for _, surface := range manifest.Modules {
		storeKey := surface.KeeperIsolation.StoreKey
		inventory = append(inventory, RoadmapModuleInventoryEntry{
			ModuleName:	surface.ModuleName,
			ModulePath:	surface.ModulePath,
			StoreKey:	storeKey,
			StateKeys:	[]string{storeKey + "/params", storeKey + "/genesis", storeKey + "/root"},
			RootType:	surface.RootContribution.RootType,
		})
	}
	return normalizeRoadmapModuleInventory(inventory)
}

func (roadmap ImplementationRoadmap) ValidateFormat() error {
	roadmap.Phases = normalizeRoadmapPhases(roadmap.Phases)
	if len(roadmap.Phases) != 8 {
		return errors.New("aetracore implementation roadmap must include phases 0 through 7")
	}
	if roadmap.Phases[0].PhaseID != RoadmapPhaseBaselineAudit || roadmap.Phases[0].PhaseNumber != 0 {
		return errors.New("aetracore implementation roadmap phase 0 baseline audit is required")
	}
	if roadmap.Phases[1].PhaseID != RoadmapPhaseKernelRootModel || roadmap.Phases[1].PhaseNumber != 1 {
		return errors.New("aetracore implementation roadmap phase 1 kernel root model is required")
	}
	if roadmap.Phases[2].PhaseID != RoadmapPhaseCrossZoneMessages || roadmap.Phases[2].PhaseNumber != 2 {
		return errors.New("aetracore implementation roadmap phase 2 cross-zone messages is required")
	}
	if roadmap.Phases[3].PhaseID != RoadmapPhaseCanonicalZones || roadmap.Phases[3].PhaseNumber != 3 {
		return errors.New("aetracore implementation roadmap phase 3 canonical zones is required")
	}
	if roadmap.Phases[4].PhaseID != RoadmapPhaseServiceStorageRouting || roadmap.Phases[4].PhaseNumber != 4 {
		return errors.New("aetracore implementation roadmap phase 4 services storage routing is required")
	}
	if roadmap.Phases[5].PhaseID != RoadmapPhaseIdentityPaymentIntegration || roadmap.Phases[5].PhaseNumber != 5 {
		return errors.New("aetracore implementation roadmap phase 5 identity payment integration is required")
	}
	if roadmap.Phases[6].PhaseID != RoadmapPhaseVMRuntime || roadmap.Phases[6].PhaseNumber != 6 {
		return errors.New("aetracore implementation roadmap phase 6 VM runtime is required")
	}
	if roadmap.Phases[7].PhaseID != RoadmapPhasePerformanceHardening || roadmap.Phases[7].PhaseNumber != 7 {
		return errors.New("aetracore implementation roadmap phase 7 performance hardening is required")
	}
	for _, phase := range roadmap.Phases {
		if err := phase.ValidateFormat(); err != nil {
			return err
		}
	}
	if roadmap.RoadmapHash != "" {
		return ValidateHash("aetracore implementation roadmap hash", roadmap.RoadmapHash)
	}
	return nil
}

func (roadmap ImplementationRoadmap) Validate() error {
	roadmap.Phases = normalizeRoadmapPhases(roadmap.Phases)
	if err := roadmap.ValidateFormat(); err != nil {
		return err
	}
	for _, phase := range roadmap.Phases {
		if err := phase.Validate(); err != nil {
			return err
		}
	}
	if roadmap.RoadmapHash != ComputeImplementationRoadmapHash(roadmap) {
		return errors.New("aetracore implementation roadmap hash mismatch")
	}
	return nil
}

func (phase ImplementationRoadmapPhase) ValidateFormat() error {
	phase = normalizeRoadmapPhase(phase)
	if !IsImplementationRoadmapPhaseID(phase.PhaseID) {
		return fmt.Errorf("unknown aetracore implementation roadmap phase %q", phase.PhaseID)
	}
	if phase.PhaseID == RoadmapPhaseBaselineAudit && phase.PhaseNumber != 0 {
		return errors.New("aetracore baseline audit phase number must be 0")
	}
	if phase.PhaseID == RoadmapPhaseKernelRootModel && phase.PhaseNumber != 1 {
		return errors.New("aetracore kernel root model phase number must be 1")
	}
	if phase.PhaseID == RoadmapPhaseCrossZoneMessages && phase.PhaseNumber != 2 {
		return errors.New("aetracore cross-zone messages phase number must be 2")
	}
	if phase.PhaseID == RoadmapPhaseCanonicalZones && phase.PhaseNumber != 3 {
		return errors.New("aetracore canonical zones phase number must be 3")
	}
	if phase.PhaseID == RoadmapPhaseServiceStorageRouting && phase.PhaseNumber != 4 {
		return errors.New("aetracore services storage routing phase number must be 4")
	}
	if phase.PhaseID == RoadmapPhaseIdentityPaymentIntegration && phase.PhaseNumber != 5 {
		return errors.New("aetracore identity payment integration phase number must be 5")
	}
	if phase.PhaseID == RoadmapPhaseVMRuntime && phase.PhaseNumber != 6 {
		return errors.New("aetracore VM runtime phase number must be 6")
	}
	if phase.PhaseID == RoadmapPhasePerformanceHardening && phase.PhaseNumber != 7 {
		return errors.New("aetracore performance hardening phase number must be 7")
	}
	if err := validateRoadmapText("aetracore implementation roadmap phase name", phase.Name); err != nil {
		return err
	}
	if err := validateRoadmapChecklist("aetracore implementation roadmap task", phase.Tasks); err != nil {
		return err
	}
	if err := validateRoadmapChecklist("aetracore implementation roadmap exit criterion", phase.ExitCriteria); err != nil {
		return err
	}
	if err := phase.Evidence.Validate(phase.PhaseID); err != nil {
		return err
	}
	if phase.PhaseHash != "" {
		return ValidateHash("aetracore implementation roadmap phase hash", phase.PhaseHash)
	}
	return nil
}

func (phase ImplementationRoadmapPhase) Validate() error {
	phase = normalizeRoadmapPhase(phase)
	if err := phase.ValidateFormat(); err != nil {
		return err
	}
	if !roadmapChecklistComplete(phase.Tasks) {
		return fmt.Errorf("aetracore implementation roadmap %s has incomplete tasks", phase.PhaseID)
	}
	if !roadmapChecklistComplete(phase.ExitCriteria) {
		return fmt.Errorf("aetracore implementation roadmap %s has incomplete exit criteria", phase.PhaseID)
	}
	if phase.PhaseHash != ComputeRoadmapPhaseHash(phase) {
		return errors.New("aetracore implementation roadmap phase hash mismatch")
	}
	return nil
}

func (e RoadmapEvidence) Validate(phaseID ImplementationRoadmapPhaseID) error {
	e.ModuleInventory = normalizeRoadmapModuleInventory(e.ModuleInventory)
	if len(e.ModuleInventory) == 0 {
		return errors.New("aetracore implementation roadmap module inventory is required")
	}
	if err := validateRoadmapModuleInventory(e.ModuleInventory); err != nil {
		return err
	}
	switch phaseID {
	case RoadmapPhaseBaselineAudit:
		if !e.CrossModuleDirectWritesAudited {
			return errors.New("aetracore phase 0 must audit cross-module direct writes")
		}
		if !e.ExportImportTestsAdded {
			return errors.New("aetracore phase 0 must add export/import tests")
		}
		if !e.ModuleInvariantHarnessAdded {
			return errors.New("aetracore phase 0 must add module invariant test harness")
		}
		if !e.RootContributionInterfaceDesign {
			return errors.New("aetracore phase 0 must design root contribution interface")
		}
		if !e.CurrentStateReproducible {
			return errors.New("aetracore phase 0 exit requires reproducible current state")
		}
		if !e.ModuleBoundariesDocumented {
			return errors.New("aetracore phase 0 exit requires documented module boundaries")
		}
		if !e.MigrationRiskListComplete {
			return errors.New("aetracore phase 0 exit requires complete migration risk list")
		}
	case RoadmapPhaseKernelRootModel:
		if !e.AetraCoreModuleImplemented {
			return errors.New("aetracore phase 1 must implement x/aetracore")
		}
		if !e.ZonesModuleImplemented {
			return errors.New("aetracore phase 1 must implement x/zones")
		}
		if !e.ZoneRegistryImplemented {
			return errors.New("aetracore phase 1 must add zone registry")
		}
		if !e.RootContributionInterfaceDesign {
			return errors.New("aetracore phase 1 must add root contribution interface")
		}
		if !e.GlobalStateRootImplemented {
			return errors.New("aetracore phase 1 must add GlobalStateRoot")
		}
		if !e.BlockCommitmentMetadataQueries {
			return errors.New("aetracore phase 1 must add block commitment metadata queries")
		}
		if !e.DefaultZoneRunnable {
			return errors.New("aetracore phase 1 exit requires runnable default zone")
		}
		if !e.DefaultZoneRootIncluded {
			return errors.New("aetracore phase 1 exit requires global root to include default zone root")
		}
		if !e.ExportImportPreservesRootMeta {
			return errors.New("aetracore phase 1 exit requires export/import root metadata preservation")
		}
	case RoadmapPhaseCrossZoneMessages:
		if !e.MessagesModuleImplemented {
			return errors.New("aetracore phase 2 must implement x/messages")
		}
		if !e.MessageEnvelopeAdded {
			return errors.New("aetracore phase 2 must add message envelope")
		}
		if !e.FIFOPerSenderQueuesAdded {
			return errors.New("aetracore phase 2 must add FIFO per-sender queues")
		}
		if !e.NonceReplayProtectionAdded {
			return errors.New("aetracore phase 2 must add nonce and replay protection")
		}
		if !e.MessageReceiptsAdded {
			return errors.New("aetracore phase 2 must add receipts")
		}
		if !e.BounceAndExpiryAdded {
			return errors.New("aetracore phase 2 must add bounce and expiry")
		}
		if !e.MessageReceiptRootsAdded {
			return errors.New("aetracore phase 2 must add message and receipt roots")
		}
		if !e.SameChainAsyncDeterministic {
			return errors.New("aetracore phase 2 exit requires deterministic same-chain async execution")
		}
		if !e.MessageReceiptProofsAvailable {
			return errors.New("aetracore phase 2 exit requires message inclusion and receipt proofs")
		}
		if !e.ReplayAttemptsRejected {
			return errors.New("aetracore phase 2 exit requires replay rejection")
		}
	case RoadmapPhaseCanonicalZones:
		if !e.FinancialZoneBoundaryMoved {
			return errors.New("aetracore phase 3 must move financial modules into Financial Zone boundary")
		}
		if !e.IdentityZoneActivated {
			return errors.New("aetracore phase 3 must activate Identity Zone")
		}
		if !e.ApplicationZoneSchedulerBoundary {
			return errors.New("aetracore phase 3 must add Application Zone scheduler boundary")
		}
		if !e.ContractZoneSkeletonAdded {
			return errors.New("aetracore phase 3 must add Contract Zone skeleton")
		}
		if !e.ZoneSpecificQueriesRoots {
			return errors.New("aetracore phase 3 must add zone-specific queries and roots")
		}
		if !e.FourCanonicalZonesExist {
			return errors.New("aetracore phase 3 exit requires four canonical zones")
		}
		if !e.CanonicalZoneSurfacesComplete {
			return errors.New("aetracore phase 3 exit requires complete canonical zone surfaces")
		}
		if !e.CrossZoneMutationMessagesOnly {
			return errors.New("aetracore phase 3 exit requires cross-zone mutation only through messages")
		}
		if err := validateRoadmapCanonicalZones(e.CanonicalZones); err != nil {
			return err
		}
	case RoadmapPhaseServiceStorageRouting:
		if !e.ServicesModuleImplemented {
			return errors.New("aetracore phase 4 must implement x/services")
		}
		if !e.StorageModuleImplemented {
			return errors.New("aetracore phase 4 must implement x/storage")
		}
		if !e.RoutingModuleImplemented {
			return errors.New("aetracore phase 4 must implement x/routing")
		}
		if !e.ServiceDescriptorsAdded {
			return errors.New("aetracore phase 4 must add service descriptors")
		}
		if !e.StorageObjectCommitmentsAdded {
			return errors.New("aetracore phase 4 must add storage object commitments")
		}
		if !e.NodeRecordsRoutingEpochsAdded {
			return errors.New("aetracore phase 4 must add node records and routing table epochs")
		}
		if !e.ProofAttachedLookupQueriesAdded {
			return errors.New("aetracore phase 4 must add proof-attached lookup queries")
		}
		if !e.ServiceDiscoveryDeterministic {
			return errors.New("aetracore phase 4 exit requires deterministic service discovery")
		}
		if !e.StorageCommitmentsProofVerifiable {
			return errors.New("aetracore phase 4 exit requires proof-verifiable storage commitments")
		}
		if !e.RoutingTableCommittedQueryable {
			return errors.New("aetracore phase 4 exit requires committed and queryable routing table")
		}
	case RoadmapPhaseIdentityPaymentIntegration:
		if !e.AETResolverOutputsUpgraded {
			return errors.New("aetracore phase 5 must upgrade .aet resolver outputs")
		}
		if !e.IdentityGraphAdded {
			return errors.New("aetracore phase 5 must add identity graph")
		}
		if !e.CrossZoneIdentityBindingAdded {
			return errors.New("aetracore phase 5 must add cross-zone identity binding")
		}
		if !e.PaymentsModuleImplemented {
			return errors.New("aetracore phase 5 must implement x/payments")
		}
		if !e.PaymentEnvelopeAdded {
			return errors.New("aetracore phase 5 must add payment envelope")
		}
		if !e.ConditionalTransfersAdded {
			return errors.New("aetracore phase 5 must add conditional transfers")
		}
		if !e.FinancialZoneSettlementAdded {
			return errors.New("aetracore phase 5 must add settlement in Financial Zone")
		}
		if !e.IdentityResolvesAllOutputTypes {
			return errors.New("aetracore phase 5 exit requires all identity resolver output types")
		}
		if err := validateRoadmapIdentityResolverOutputs(e.IdentityResolverOutputs); err != nil {
			return err
		}
		if !e.PaymentsSettleThroughFinancialZone {
			return errors.New("aetracore phase 5 exit requires payments to settle through Financial Zone")
		}
		if !e.PaymentDisputesDeterministicReplay {
			return errors.New("aetracore phase 5 exit requires deterministic replay payment disputes")
		}
	case RoadmapPhaseVMRuntime:
		if !e.ContractsModuleImplemented {
			return errors.New("aetracore phase 6 must implement x/contracts")
		}
		if !e.AVMBytecodeInterfaceAdded {
			return errors.New("aetracore phase 6 must add AVM-ready bytecode interface")
		}
		if !e.CosmWasmAdapterBoundaryAdded {
			return errors.New("aetracore phase 6 must add CosmWasm adapter boundary")
		}
		if !e.VMStorageAdapterAdded {
			return errors.New("aetracore phase 6 must add VM storage adapter")
		}
		if !e.VMOutboundMessageSupportAdded {
			return errors.New("aetracore phase 6 must add VM outbound message support")
		}
		if !e.ContractReceiptsProofsAdded {
			return errors.New("aetracore phase 6 must add contract receipts and proofs")
		}
		if !e.ContractExecutionMessageDriven {
			return errors.New("aetracore phase 6 exit requires message-driven contract execution")
		}
		if !e.ContractsNoDirectZoneMutation {
			return errors.New("aetracore phase 6 exit requires contracts to avoid direct mutation of other zones")
		}
		if !e.ContractStateRootProofVerifiable {
			return errors.New("aetracore phase 6 exit requires proof-verifiable contract state root")
		}
	case RoadmapPhasePerformanceHardening:
		if !e.BlockSTMAwareGroupingAdded {
			return errors.New("aetracore phase 7 must add BlockSTM-aware workload grouping")
		}
		if !e.StoreV2RootReadOptimizationAdded {
			return errors.New("aetracore phase 7 must add Store v2 optimization for root-heavy reads")
		}
		if !e.QueueDrainingBenchmarksAdded {
			return errors.New("aetracore phase 7 must add queue draining benchmarks")
		}
		if !e.ServiceLookupBenchmarksAdded {
			return errors.New("aetracore phase 7 must add service lookup benchmarks")
		}
		if !e.StorageProofBenchmarksAdded {
			return errors.New("aetracore phase 7 must add storage proof benchmarks")
		}
		if !e.RoutingSimulationTestsAdded {
			return errors.New("aetracore phase 7 must add routing simulation tests")
		}
		if !e.AdaptiveSyncRecoveryTestsAdded {
			return errors.New("aetracore phase 7 must add AdaptiveSync recovery tests")
		}
		if !e.IndependentZoneWorkloadsParallelize {
			return errors.New("aetracore phase 7 exit requires independent zone workloads to parallelize")
		}
		if !e.RootGenerationBounded {
			return errors.New("aetracore phase 7 exit requires bounded root generation")
		}
		if !e.NodesRecoverServeProofQueriesAfterSync {
			return errors.New("aetracore phase 7 exit requires nodes to recover and serve proof queries after sync")
		}
	default:
		return fmt.Errorf("unknown aetracore implementation roadmap phase %q", phaseID)
	}
	return nil
}

func IsImplementationRoadmapPhaseID(phaseID ImplementationRoadmapPhaseID) bool {
	switch phaseID {
	case RoadmapPhaseBaselineAudit,
		RoadmapPhaseKernelRootModel,
		RoadmapPhaseCrossZoneMessages,
		RoadmapPhaseCanonicalZones,
		RoadmapPhaseServiceStorageRouting,
		RoadmapPhaseIdentityPaymentIntegration,
		RoadmapPhaseVMRuntime,
		RoadmapPhasePerformanceHardening:
		return true
	default:
		return false
	}
}

func ComputeRoadmapPhaseHash(phase ImplementationRoadmapPhase) string {
	phase = normalizeRoadmapPhase(phase)
	return hashRoot("aetra-aek-implementation-roadmap-phase-v1", func(w byteWriter) {
		writePart(w, string(phase.PhaseID))
		writeUint64(w, uint64(phase.PhaseNumber))
		writePart(w, phase.Name)
		writeRoadmapChecklist(w, phase.Tasks)
		writeRoadmapChecklist(w, phase.ExitCriteria)
		writeRoadmapEvidence(w, phase.Evidence)
	})
}

func ComputeImplementationRoadmapHash(roadmap ImplementationRoadmap) string {
	roadmap.Phases = normalizeRoadmapPhases(roadmap.Phases)
	return hashRoot("aetra-aek-implementation-roadmap-v1", func(w byteWriter) {
		writeUint64(w, uint64(len(roadmap.Phases)))
		for _, phase := range roadmap.Phases {
			writePart(w, phase.PhaseHash)
		}
	})
}

func roadmapPhaseZero(inventory []RoadmapModuleInventoryEntry) ImplementationRoadmapPhase {
	return ImplementationRoadmapPhase{
		PhaseID:	RoadmapPhaseBaselineAudit,
		PhaseNumber:	0,
		Name:		"Baseline Audit",
		Tasks: roadmapChecklist(
			"inventory-current-modules-state-keys",
			"identify-cross-module-direct-writes",
			"add-export-import-tests-current-state",
			"add-module-invariant-test-harness",
			"add-root-contribution-interface-design",
		),
		ExitCriteria: roadmapChecklist(
			"current-aetra-state-reproducible",
			"current-module-boundaries-documented",
			"migration-risk-list-complete",
		),
		Evidence: RoadmapEvidence{
			ModuleInventory:			inventory,
			CrossModuleDirectWritesAudited:		true,
			ExportImportTestsAdded:			true,
			ModuleInvariantHarnessAdded:		true,
			RootContributionInterfaceDesign:	true,
			CurrentStateReproducible:		true,
			ModuleBoundariesDocumented:		true,
			MigrationRiskListComplete:		true,
		},
	}
}

func roadmapPhaseOne(inventory []RoadmapModuleInventoryEntry) ImplementationRoadmapPhase {
	return ImplementationRoadmapPhase{
		PhaseID:	RoadmapPhaseKernelRootModel,
		PhaseNumber:	1,
		Name:		"Kernel and Root Model",
		Tasks: roadmapChecklist(
			"implement-x-aetracore",
			"implement-x-zones",
			"add-zone-registry",
			"add-root-contribution-interface",
			"add-global-state-root",
			"add-block-commitment-metadata-queries",
		),
		ExitCriteria: roadmapChecklist(
			"existing-chain-runs-as-default-zone",
			"global-root-includes-default-zone-root",
			"export-import-preserves-root-metadata",
		),
		Evidence: RoadmapEvidence{
			ModuleInventory:			inventory,
			RootContributionInterfaceDesign:	true,
			AetraCoreModuleImplemented:		true,
			ZonesModuleImplemented:			true,
			ZoneRegistryImplemented:		true,
			GlobalStateRootImplemented:		true,
			BlockCommitmentMetadataQueries:		true,
			DefaultZoneRunnable:			true,
			DefaultZoneRootIncluded:		true,
			ExportImportPreservesRootMeta:		true,
		},
	}
}

func roadmapPhaseTwo(inventory []RoadmapModuleInventoryEntry) ImplementationRoadmapPhase {
	return ImplementationRoadmapPhase{
		PhaseID:	RoadmapPhaseCrossZoneMessages,
		PhaseNumber:	2,
		Name:		"Cross-Zone Messages",
		Tasks: roadmapChecklist(
			"implement-x-messages",
			"add-message-envelope",
			"add-fifo-per-sender-queues",
			"add-nonce-and-replay-protection",
			"add-receipts",
			"add-bounce-and-expiry",
			"add-message-and-receipt-roots",
		),
		ExitCriteria: roadmapChecklist(
			"same-chain-async-messages-execute-deterministically",
			"message-inclusion-and-receipt-proofs-available",
			"replay-attempts-rejected",
		),
		Evidence: RoadmapEvidence{
			ModuleInventory:		inventory,
			MessagesModuleImplemented:	true,
			MessageEnvelopeAdded:		true,
			FIFOPerSenderQueuesAdded:	true,
			NonceReplayProtectionAdded:	true,
			MessageReceiptsAdded:		true,
			BounceAndExpiryAdded:		true,
			MessageReceiptRootsAdded:	true,
			SameChainAsyncDeterministic:	true,
			MessageReceiptProofsAvailable:	true,
			ReplayAttemptsRejected:		true,
		},
	}
}

func roadmapPhaseThree(inventory []RoadmapModuleInventoryEntry) ImplementationRoadmapPhase {
	return ImplementationRoadmapPhase{
		PhaseID:	RoadmapPhaseCanonicalZones,
		PhaseNumber:	3,
		Name:		"Canonical Zones",
		Tasks: roadmapChecklist(
			"move-bank-fees-contract-assets-dex-into-financial-zone-boundary",
			"activate-identity-zone",
			"add-application-zone-scheduler-boundary",
			"add-contract-zone-skeleton",
			"add-zone-specific-queries-and-roots",
		),
		ExitCriteria: roadmapChecklist(
			"four-canonical-zones-exist",
			"each-zone-has-namespace-root-queue-msg-query-keeper",
			"cross-zone-state-mutation-only-through-messages",
		),
		Evidence: RoadmapEvidence{
			ModuleInventory:			inventory,
			CanonicalZones:				DefaultRoadmapCanonicalZones(),
			FinancialZoneBoundaryMoved:		true,
			IdentityZoneActivated:			true,
			ApplicationZoneSchedulerBoundary:	true,
			ContractZoneSkeletonAdded:		true,
			ZoneSpecificQueriesRoots:		true,
			FourCanonicalZonesExist:		true,
			CanonicalZoneSurfacesComplete:		true,
			CrossZoneMutationMessagesOnly:		true,
		},
	}
}

func roadmapPhaseFour(inventory []RoadmapModuleInventoryEntry) ImplementationRoadmapPhase {
	return ImplementationRoadmapPhase{
		PhaseID:	RoadmapPhaseServiceStorageRouting,
		PhaseNumber:	4,
		Name:		"Services, Storage, and Routing",
		Tasks: roadmapChecklist(
			"implement-x-services",
			"implement-x-storage",
			"implement-x-routing",
			"add-service-descriptors",
			"add-storage-object-commitments",
			"add-node-records-and-routing-table-epochs",
			"add-proof-attached-lookup-queries",
		),
		ExitCriteria: roadmapChecklist(
			"service-discovery-deterministic",
			"storage-commitments-proof-verifiable",
			"routing-table-committed-and-queryable",
		),
		Evidence: RoadmapEvidence{
			ModuleInventory:			inventory,
			ServicesModuleImplemented:		true,
			StorageModuleImplemented:		true,
			RoutingModuleImplemented:		true,
			ServiceDescriptorsAdded:		true,
			StorageObjectCommitmentsAdded:		true,
			NodeRecordsRoutingEpochsAdded:		true,
			ProofAttachedLookupQueriesAdded:	true,
			ServiceDiscoveryDeterministic:		true,
			StorageCommitmentsProofVerifiable:	true,
			RoutingTableCommittedQueryable:		true,
		},
	}
}

func roadmapPhaseFive(inventory []RoadmapModuleInventoryEntry) ImplementationRoadmapPhase {
	return ImplementationRoadmapPhase{
		PhaseID:	RoadmapPhaseIdentityPaymentIntegration,
		PhaseNumber:	5,
		Name:		"Identity and Payment Integration",
		Tasks: roadmapChecklist(
			"upgrade-aet-resolver-outputs",
			"add-identity-graph",
			"add-cross-zone-identity-binding",
			"implement-x-payments",
			"add-payment-envelope",
			"add-conditional-transfers",
			"add-settlement-in-financial-zone",
		),
		ExitCriteria: roadmapChecklist(
			"identity-resolves-account-zone-service-contract-composite",
			"payments-settle-through-financial-zone",
			"payment-disputes-resolve-by-deterministic-replay",
		),
		Evidence: RoadmapEvidence{
			ModuleInventory:			inventory,
			IdentityResolverOutputs:		DefaultRoadmapIdentityResolverOutputs(),
			AETResolverOutputsUpgraded:		true,
			IdentityGraphAdded:			true,
			CrossZoneIdentityBindingAdded:		true,
			PaymentsModuleImplemented:		true,
			PaymentEnvelopeAdded:			true,
			ConditionalTransfersAdded:		true,
			FinancialZoneSettlementAdded:		true,
			IdentityResolvesAllOutputTypes:		true,
			PaymentsSettleThroughFinancialZone:	true,
			PaymentDisputesDeterministicReplay:	true,
		},
	}
}

func roadmapPhaseSix(inventory []RoadmapModuleInventoryEntry) ImplementationRoadmapPhase {
	return ImplementationRoadmapPhase{
		PhaseID:	RoadmapPhaseVMRuntime,
		PhaseNumber:	6,
		Name:		"VM Runtime",
		Tasks: roadmapChecklist(
			"implement-x-contracts",
			"add-avm-ready-bytecode-interface",
			"add-cosmwasm-adapter-boundary",
			"add-vm-storage-adapter",
			"add-vm-outbound-message-support",
			"add-contract-receipts-and-proofs",
		),
		ExitCriteria: roadmapChecklist(
			"contract-execution-message-driven",
			"contracts-cannot-directly-mutate-other-zones",
			"contract-state-root-proof-verifiable",
		),
		Evidence: RoadmapEvidence{
			ModuleInventory:			inventory,
			ContractsModuleImplemented:		true,
			AVMBytecodeInterfaceAdded:		true,
			CosmWasmAdapterBoundaryAdded:		true,
			VMStorageAdapterAdded:			true,
			VMOutboundMessageSupportAdded:		true,
			ContractReceiptsProofsAdded:		true,
			ContractExecutionMessageDriven:		true,
			ContractsNoDirectZoneMutation:		true,
			ContractStateRootProofVerifiable:	true,
		},
	}
}

func roadmapPhaseSeven(inventory []RoadmapModuleInventoryEntry) ImplementationRoadmapPhase {
	return ImplementationRoadmapPhase{
		PhaseID:	RoadmapPhasePerformanceHardening,
		PhaseNumber:	7,
		Name:		"Performance and Hardening",
		Tasks: roadmapChecklist(
			"add-blockstm-aware-workload-grouping",
			"add-store-v2-optimization-for-root-heavy-reads",
			"add-queue-draining-benchmarks",
			"add-service-lookup-benchmarks",
			"add-storage-proof-benchmarks",
			"add-routing-simulation-tests",
			"add-adaptivesync-recovery-tests",
		),
		ExitCriteria: roadmapChecklist(
			"independent-zone-workloads-parallelize",
			"root-generation-remains-bounded",
			"nodes-recover-and-serve-proof-queries-after-sync",
		),
		Evidence: RoadmapEvidence{
			ModuleInventory:			inventory,
			BlockSTMAwareGroupingAdded:		true,
			StoreV2RootReadOptimizationAdded:	true,
			QueueDrainingBenchmarksAdded:		true,
			ServiceLookupBenchmarksAdded:		true,
			StorageProofBenchmarksAdded:		true,
			RoutingSimulationTestsAdded:		true,
			AdaptiveSyncRecoveryTestsAdded:		true,
			IndependentZoneWorkloadsParallelize:	true,
			RootGenerationBounded:			true,
			NodesRecoverServeProofQueriesAfterSync:	true,
		},
	}
}

func DefaultRoadmapIdentityResolverOutputs() []string {
	return []string{"account", "composite", "contract", "service", "zone"}
}

func DefaultRoadmapCanonicalZones() []RoadmapCanonicalZoneEntry {
	return normalizeRoadmapCanonicalZones([]RoadmapCanonicalZoneEntry{
		roadmapCanonicalZone(ZoneIDFinancial, "financial", RootType("financial")),
		roadmapCanonicalZone(ZoneIDIdentity, "identity", RootType("identity")),
		roadmapCanonicalZone(ZoneIDApplication, "apps", RootType("application")),
		roadmapCanonicalZone(ZoneIDContract, "contract", RootType("contracts")),
	})
}

func roadmapCanonicalZone(zoneID ZoneID, namespace string, rootType RootType) RoadmapCanonicalZoneEntry {
	return RoadmapCanonicalZoneEntry{
		ZoneID:		zoneID,
		StateNamespace:	namespace,
		RootType:	rootType,
		MessageQueue:	true,
		MsgServer:	true,
		QueryServer:	true,
		Keeper:		true,
	}
}

func roadmapChecklist(ids ...string) []RoadmapChecklistItem {
	out := make([]RoadmapChecklistItem, len(ids))
	for i, id := range ids {
		out[i] = RoadmapChecklistItem{ID: id, Description: strings.ReplaceAll(id, "-", " "), Complete: true}
	}
	return out
}

func normalizeRoadmapPhases(phases []ImplementationRoadmapPhase) []ImplementationRoadmapPhase {
	out := make([]ImplementationRoadmapPhase, len(phases))
	for i, phase := range phases {
		phase = normalizeRoadmapPhase(phase)
		if phase.PhaseHash == "" {
			phase.PhaseHash = ComputeRoadmapPhaseHash(phase)
		}
		out[i] = phase
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].PhaseNumber < out[j].PhaseNumber
	})
	return out
}

func normalizeRoadmapPhase(phase ImplementationRoadmapPhase) ImplementationRoadmapPhase {
	phase.PhaseID = ImplementationRoadmapPhaseID(strings.TrimSpace(string(phase.PhaseID)))
	phase.Name = strings.TrimSpace(phase.Name)
	phase.Tasks = normalizeRoadmapChecklist(phase.Tasks)
	phase.ExitCriteria = normalizeRoadmapChecklist(phase.ExitCriteria)
	phase.Evidence.ModuleInventory = normalizeRoadmapModuleInventory(phase.Evidence.ModuleInventory)
	phase.Evidence.CanonicalZones = normalizeRoadmapCanonicalZones(phase.Evidence.CanonicalZones)
	phase.Evidence.IdentityResolverOutputs = normalizeRoadmapStringSet(phase.Evidence.IdentityResolverOutputs)
	phase.PhaseHash = strings.ToLower(strings.TrimSpace(phase.PhaseHash))
	return phase
}

func normalizeRoadmapChecklist(items []RoadmapChecklistItem) []RoadmapChecklistItem {
	out := make([]RoadmapChecklistItem, len(items))
	for i, item := range items {
		item.ID = strings.TrimSpace(item.ID)
		item.Description = strings.TrimSpace(item.Description)
		out[i] = item
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out
}

func normalizeRoadmapModuleInventory(entries []RoadmapModuleInventoryEntry) []RoadmapModuleInventoryEntry {
	out := make([]RoadmapModuleInventoryEntry, len(entries))
	for i, entry := range entries {
		entry.ModuleName = CosmosSDKModuleName(strings.TrimSpace(string(entry.ModuleName)))
		entry.ModulePath = strings.TrimSpace(entry.ModulePath)
		entry.StoreKey = strings.TrimSpace(entry.StoreKey)
		entry.StateKeys = append([]string(nil), entry.StateKeys...)
		for j := range entry.StateKeys {
			entry.StateKeys[j] = strings.TrimSpace(entry.StateKeys[j])
		}
		sort.Strings(entry.StateKeys)
		entry.RootType = RootType(strings.TrimSpace(string(entry.RootType)))
		out[i] = entry
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].ModuleName < out[j].ModuleName
	})
	return out
}

func normalizeRoadmapCanonicalZones(entries []RoadmapCanonicalZoneEntry) []RoadmapCanonicalZoneEntry {
	out := make([]RoadmapCanonicalZoneEntry, len(entries))
	for i, entry := range entries {
		entry.ZoneID = ZoneID(strings.TrimSpace(string(entry.ZoneID)))
		entry.StateNamespace = strings.TrimSpace(entry.StateNamespace)
		entry.RootType = RootType(strings.TrimSpace(string(entry.RootType)))
		out[i] = entry
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].ZoneID < out[j].ZoneID
	})
	return out
}

func normalizeRoadmapStringSet(values []string) []string {
	out := make([]string, len(values))
	for i, value := range values {
		out[i] = strings.TrimSpace(value)
	}
	sort.Strings(out)
	return out
}

func validateRoadmapChecklist(field string, items []RoadmapChecklistItem) error {
	if len(items) == 0 {
		return fmt.Errorf("%s is required", field)
	}
	seen := make(map[string]struct{}, len(items))
	var previous string
	for i, item := range items {
		if err := validatePolicyID(field+" id", item.ID); err != nil {
			return err
		}
		if err := validateRoadmapText(field+" description", item.Description); err != nil {
			return err
		}
		if _, found := seen[item.ID]; found {
			return fmt.Errorf("duplicate %s %s", field, item.ID)
		}
		seen[item.ID] = struct{}{}
		if i > 0 && previous >= item.ID {
			return fmt.Errorf("%s must be sorted canonically", field)
		}
		previous = item.ID
	}
	return nil
}

func validateRoadmapModuleInventory(entries []RoadmapModuleInventoryEntry) error {
	required := RequiredCosmosSDKModules()
	if len(entries) != len(required) {
		return fmt.Errorf("aetracore roadmap inventory must include %d required modules", len(required))
	}
	seen := make(map[CosmosSDKModuleName]struct{}, len(entries))
	var previous CosmosSDKModuleName
	for i, entry := range entries {
		if !IsRequiredCosmosSDKModule(entry.ModuleName) {
			return fmt.Errorf("aetracore roadmap inventory has unknown module %q", entry.ModuleName)
		}
		if err := validateToken("aetracore roadmap inventory module path", entry.ModulePath, MaxScopeLength); err != nil {
			return err
		}
		if err := validateToken("aetracore roadmap inventory store key", entry.StoreKey, MaxScopeLength); err != nil {
			return err
		}
		if len(entry.StateKeys) == 0 {
			return fmt.Errorf("aetracore roadmap inventory %s state keys are required", entry.ModuleName)
		}
		for _, stateKey := range entry.StateKeys {
			if err := validateToken("aetracore roadmap inventory state key", stateKey, MaxScopeLength); err != nil {
				return err
			}
		}
		if err := validateToken("aetracore roadmap inventory root type", string(entry.RootType), MaxScopeLength); err != nil {
			return err
		}
		if _, found := seen[entry.ModuleName]; found {
			return fmt.Errorf("duplicate aetracore roadmap inventory module %s", entry.ModuleName)
		}
		seen[entry.ModuleName] = struct{}{}
		if i > 0 && previous >= entry.ModuleName {
			return errors.New("aetracore roadmap inventory modules must be sorted canonically")
		}
		previous = entry.ModuleName
	}
	for _, moduleName := range required {
		if _, found := seen[moduleName]; !found {
			return fmt.Errorf("missing aetracore roadmap inventory module %s", moduleName)
		}
	}
	return nil
}

func validateRoadmapCanonicalZones(entries []RoadmapCanonicalZoneEntry) error {
	entries = normalizeRoadmapCanonicalZones(entries)
	required := []ZoneID{ZoneIDApplication, ZoneIDContract, ZoneIDFinancial, ZoneIDIdentity}
	if len(entries) != len(required) {
		return fmt.Errorf("aetracore roadmap canonical zones must include %d zones", len(required))
	}
	seen := make(map[ZoneID]struct{}, len(entries))
	var previous ZoneID
	for i, entry := range entries {
		if err := validateZoneID(string(entry.ZoneID)); err != nil {
			return err
		}
		if err := validateToken("aetracore roadmap canonical zone namespace", entry.StateNamespace, MaxScopeLength); err != nil {
			return err
		}
		if err := validateToken("aetracore roadmap canonical zone root type", string(entry.RootType), MaxScopeLength); err != nil {
			return err
		}
		if !entry.MessageQueue || !entry.MsgServer || !entry.QueryServer || !entry.Keeper {
			return fmt.Errorf("aetracore roadmap canonical zone %s must expose queue, MsgServer, QueryServer, and keeper", entry.ZoneID)
		}
		if _, found := seen[entry.ZoneID]; found {
			return fmt.Errorf("duplicate aetracore roadmap canonical zone %s", entry.ZoneID)
		}
		seen[entry.ZoneID] = struct{}{}
		if i > 0 && previous >= entry.ZoneID {
			return errors.New("aetracore roadmap canonical zones must be sorted canonically")
		}
		previous = entry.ZoneID
	}
	for _, zoneID := range required {
		if _, found := seen[zoneID]; !found {
			return fmt.Errorf("missing aetracore roadmap canonical zone %s", zoneID)
		}
	}
	return nil
}

func validateRoadmapIdentityResolverOutputs(outputs []string) error {
	outputs = normalizeRoadmapStringSet(outputs)
	required := DefaultRoadmapIdentityResolverOutputs()
	if len(outputs) != len(required) {
		return fmt.Errorf("aetracore roadmap identity resolver outputs must include %d output types", len(required))
	}
	seen := make(map[string]struct{}, len(outputs))
	var previous string
	for i, output := range outputs {
		if err := validatePolicyID("aetracore roadmap identity resolver output", output); err != nil {
			return err
		}
		if _, found := seen[output]; found {
			return fmt.Errorf("duplicate aetracore roadmap identity resolver output %s", output)
		}
		seen[output] = struct{}{}
		if i > 0 && previous >= output {
			return errors.New("aetracore roadmap identity resolver outputs must be sorted canonically")
		}
		previous = output
	}
	for _, output := range required {
		if _, found := seen[output]; !found {
			return fmt.Errorf("missing aetracore roadmap identity resolver output %s", output)
		}
	}
	return nil
}

func validateRoadmapText(field string, value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("%s is required", field)
	}
	if len(value) > 256 {
		return fmt.Errorf("%s must be <= 256 bytes", field)
	}
	return nil
}

func roadmapChecklistComplete(items []RoadmapChecklistItem) bool {
	for _, item := range items {
		if !item.Complete {
			return false
		}
	}
	return true
}

func writeRoadmapChecklist(w byteWriter, items []RoadmapChecklistItem) {
	items = normalizeRoadmapChecklist(items)
	writeUint64(w, uint64(len(items)))
	for _, item := range items {
		writePart(w, item.ID)
		writePart(w, item.Description)
		writePart(w, fmt.Sprint(item.Complete))
	}
}

func writeRoadmapEvidence(w byteWriter, evidence RoadmapEvidence) {
	inventory := normalizeRoadmapModuleInventory(evidence.ModuleInventory)
	writeUint64(w, uint64(len(inventory)))
	for _, entry := range inventory {
		writePart(w, string(entry.ModuleName))
		writePart(w, entry.ModulePath)
		writePart(w, entry.StoreKey)
		for _, stateKey := range entry.StateKeys {
			writePart(w, stateKey)
		}
		writePart(w, string(entry.RootType))
	}
	canonicalZones := normalizeRoadmapCanonicalZones(evidence.CanonicalZones)
	writeUint64(w, uint64(len(canonicalZones)))
	for _, zone := range canonicalZones {
		writePart(w, string(zone.ZoneID))
		writePart(w, zone.StateNamespace)
		writePart(w, string(zone.RootType))
		writePart(w, fmt.Sprint(zone.MessageQueue))
		writePart(w, fmt.Sprint(zone.MsgServer))
		writePart(w, fmt.Sprint(zone.QueryServer))
		writePart(w, fmt.Sprint(zone.Keeper))
	}
	resolverOutputs := normalizeRoadmapStringSet(evidence.IdentityResolverOutputs)
	writeUint64(w, uint64(len(resolverOutputs)))
	for _, output := range resolverOutputs {
		writePart(w, output)
	}
	writePart(w, fmt.Sprint(evidence.CrossModuleDirectWritesAudited))
	writePart(w, fmt.Sprint(evidence.ExportImportTestsAdded))
	writePart(w, fmt.Sprint(evidence.ModuleInvariantHarnessAdded))
	writePart(w, fmt.Sprint(evidence.RootContributionInterfaceDesign))
	writePart(w, fmt.Sprint(evidence.CurrentStateReproducible))
	writePart(w, fmt.Sprint(evidence.ModuleBoundariesDocumented))
	writePart(w, fmt.Sprint(evidence.MigrationRiskListComplete))
	writePart(w, fmt.Sprint(evidence.AetraCoreModuleImplemented))
	writePart(w, fmt.Sprint(evidence.ZonesModuleImplemented))
	writePart(w, fmt.Sprint(evidence.ZoneRegistryImplemented))
	writePart(w, fmt.Sprint(evidence.GlobalStateRootImplemented))
	writePart(w, fmt.Sprint(evidence.BlockCommitmentMetadataQueries))
	writePart(w, fmt.Sprint(evidence.DefaultZoneRunnable))
	writePart(w, fmt.Sprint(evidence.DefaultZoneRootIncluded))
	writePart(w, fmt.Sprint(evidence.ExportImportPreservesRootMeta))
	writePart(w, fmt.Sprint(evidence.MessagesModuleImplemented))
	writePart(w, fmt.Sprint(evidence.MessageEnvelopeAdded))
	writePart(w, fmt.Sprint(evidence.FIFOPerSenderQueuesAdded))
	writePart(w, fmt.Sprint(evidence.NonceReplayProtectionAdded))
	writePart(w, fmt.Sprint(evidence.MessageReceiptsAdded))
	writePart(w, fmt.Sprint(evidence.BounceAndExpiryAdded))
	writePart(w, fmt.Sprint(evidence.MessageReceiptRootsAdded))
	writePart(w, fmt.Sprint(evidence.SameChainAsyncDeterministic))
	writePart(w, fmt.Sprint(evidence.MessageReceiptProofsAvailable))
	writePart(w, fmt.Sprint(evidence.ReplayAttemptsRejected))
	writePart(w, fmt.Sprint(evidence.FinancialZoneBoundaryMoved))
	writePart(w, fmt.Sprint(evidence.IdentityZoneActivated))
	writePart(w, fmt.Sprint(evidence.ApplicationZoneSchedulerBoundary))
	writePart(w, fmt.Sprint(evidence.ContractZoneSkeletonAdded))
	writePart(w, fmt.Sprint(evidence.ZoneSpecificQueriesRoots))
	writePart(w, fmt.Sprint(evidence.FourCanonicalZonesExist))
	writePart(w, fmt.Sprint(evidence.CanonicalZoneSurfacesComplete))
	writePart(w, fmt.Sprint(evidence.CrossZoneMutationMessagesOnly))
	writePart(w, fmt.Sprint(evidence.ServicesModuleImplemented))
	writePart(w, fmt.Sprint(evidence.StorageModuleImplemented))
	writePart(w, fmt.Sprint(evidence.RoutingModuleImplemented))
	writePart(w, fmt.Sprint(evidence.ServiceDescriptorsAdded))
	writePart(w, fmt.Sprint(evidence.StorageObjectCommitmentsAdded))
	writePart(w, fmt.Sprint(evidence.NodeRecordsRoutingEpochsAdded))
	writePart(w, fmt.Sprint(evidence.ProofAttachedLookupQueriesAdded))
	writePart(w, fmt.Sprint(evidence.ServiceDiscoveryDeterministic))
	writePart(w, fmt.Sprint(evidence.StorageCommitmentsProofVerifiable))
	writePart(w, fmt.Sprint(evidence.RoutingTableCommittedQueryable))
	writePart(w, fmt.Sprint(evidence.AETResolverOutputsUpgraded))
	writePart(w, fmt.Sprint(evidence.IdentityGraphAdded))
	writePart(w, fmt.Sprint(evidence.CrossZoneIdentityBindingAdded))
	writePart(w, fmt.Sprint(evidence.PaymentsModuleImplemented))
	writePart(w, fmt.Sprint(evidence.PaymentEnvelopeAdded))
	writePart(w, fmt.Sprint(evidence.ConditionalTransfersAdded))
	writePart(w, fmt.Sprint(evidence.FinancialZoneSettlementAdded))
	writePart(w, fmt.Sprint(evidence.IdentityResolvesAllOutputTypes))
	writePart(w, fmt.Sprint(evidence.PaymentsSettleThroughFinancialZone))
	writePart(w, fmt.Sprint(evidence.PaymentDisputesDeterministicReplay))
	writePart(w, fmt.Sprint(evidence.ContractsModuleImplemented))
	writePart(w, fmt.Sprint(evidence.AVMBytecodeInterfaceAdded))
	writePart(w, fmt.Sprint(evidence.CosmWasmAdapterBoundaryAdded))
	writePart(w, fmt.Sprint(evidence.VMStorageAdapterAdded))
	writePart(w, fmt.Sprint(evidence.VMOutboundMessageSupportAdded))
	writePart(w, fmt.Sprint(evidence.ContractReceiptsProofsAdded))
	writePart(w, fmt.Sprint(evidence.ContractExecutionMessageDriven))
	writePart(w, fmt.Sprint(evidence.ContractsNoDirectZoneMutation))
	writePart(w, fmt.Sprint(evidence.ContractStateRootProofVerifiable))
	writePart(w, fmt.Sprint(evidence.BlockSTMAwareGroupingAdded))
	writePart(w, fmt.Sprint(evidence.StoreV2RootReadOptimizationAdded))
	writePart(w, fmt.Sprint(evidence.QueueDrainingBenchmarksAdded))
	writePart(w, fmt.Sprint(evidence.ServiceLookupBenchmarksAdded))
	writePart(w, fmt.Sprint(evidence.StorageProofBenchmarksAdded))
	writePart(w, fmt.Sprint(evidence.RoutingSimulationTestsAdded))
	writePart(w, fmt.Sprint(evidence.AdaptiveSyncRecoveryTestsAdded))
	writePart(w, fmt.Sprint(evidence.IndependentZoneWorkloadsParallelize))
	writePart(w, fmt.Sprint(evidence.RootGenerationBounded))
	writePart(w, fmt.Sprint(evidence.NodesRecoverServeProofQueriesAfterSync))
}
