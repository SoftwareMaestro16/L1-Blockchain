package types

import (
	"errors"
	"fmt"
	"sort"
)

const MigrationPathSpecVersion = uint64(1)

type MigrationPhaseID string
type MigrationTaskID string
type MigrationExitCriterionID string

const (
	MigrationPhaseBaselineHardening		MigrationPhaseID	= "phase-0-baseline-hardening"
	MigrationPhaseCoreCommitments		MigrationPhaseID	= "phase-1-core-commitments"
	MigrationPhaseMessageBus		MigrationPhaseID	= "phase-2-message-bus"
	MigrationPhaseZoneExtraction		MigrationPhaseID	= "phase-3-zone-extraction"
	MigrationPhaseShardingRuntime		MigrationPhaseID	= "phase-4-sharding-runtime"
	MigrationPhaseAVM20			MigrationPhaseID	= "phase-5-avm-2.0"
	MigrationPhaseIdentityPayments		MigrationPhaseID	= "phase-6-identity-payment-integration"
	MigrationPhasePerformanceHardening	MigrationPhaseID	= "phase-7-performance-hardening"

	MigrationTaskModuleBoundaryDocs		MigrationTaskID	= "module-boundary-documentation"
	MigrationTaskStateExportValidation	MigrationTaskID	= "state-export-validation"
	MigrationTaskDeterministicGenesis	MigrationTaskID	= "deterministic-genesis-import"
	MigrationTaskDynamicFeeBoundsTests	MigrationTaskID	= "dynamic-fee-bounds-tests"
	MigrationTaskLegacyModuleInvariants	MigrationTaskID	= "staking-slashing-bank-distribution-invariants"
	MigrationTaskStoreV2CompatibilityAudit	MigrationTaskID	= "store-v2-compatibility-audit"

	MigrationTaskAetraCoreModule		MigrationTaskID	= "aetracore-module"
	MigrationTaskDefaultZoneRegistry	MigrationTaskID	= "default-zone-registry"
	MigrationTaskDefaultZoneStateRoot	MigrationTaskID	= "default-zone-state-root"
	MigrationTaskEmptyMessageRoot		MigrationTaskID	= "empty-message-root"
	MigrationTaskProofRootRegistry		MigrationTaskID	= "proof-root-registry"
	MigrationTaskRootQueryAPIs		MigrationTaskID	= "root-query-apis"

	MigrationTaskMsgbusModule		MigrationTaskID	= "msgbus-module"
	MigrationTaskMessageEncodingAndIDs	MigrationTaskID	= "message-encoding-and-ids"
	MigrationTaskMessageStores		MigrationTaskID	= "inbox-outbox-receipt-stores"
	MigrationTaskLocalZoneExecution		MigrationTaskID	= "local-zone-message-execution"
	MigrationTaskExpiryBounceLogic		MigrationTaskID	= "expiry-and-bounce-logic"
	MigrationTaskMessageInclusionProofs	MigrationTaskID	= "message-inclusion-proofs"

	MigrationTaskExtractFinancialZone	MigrationTaskID	= "extract-financial-zone"
	MigrationTaskExtractIdentityZone	MigrationTaskID	= "extract-identity-zone"
	MigrationTaskExtractApplicationZone	MigrationTaskID	= "extract-application-zone"
	MigrationTaskZoneSpecificKeepers	MigrationTaskID	= "zone-specific-keepers-state-prefixes"
	MigrationTaskZoneLocalFeePolicies	MigrationTaskID	= "zone-local-fee-policies"
	MigrationTaskZoneExecutionSummaries	MigrationTaskID	= "zone-execution-summaries"

	MigrationTaskShardsModule			MigrationTaskID	= "shards-module"
	MigrationTaskShardLayoutDescriptors		MigrationTaskID	= "shard-layout-descriptors"
	MigrationTaskRouteKeyCalculation		MigrationTaskID	= "route-key-calculation"
	MigrationTaskPerShardMessageStores		MigrationTaskID	= "per-shard-inbox-outbox"
	MigrationTaskShardRootAggregation		MigrationTaskID	= "shard-root-aggregation"
	MigrationTaskSplitMergeScheduler		MigrationTaskID	= "split-merge-scheduler"
	MigrationTaskDeterministicShardMigration	MigrationTaskID	= "deterministic-shard-migration"

	MigrationTaskAVMBytecodeFormat		MigrationTaskID	= "avm-bytecode-format"
	MigrationTaskAVMInterpreter		MigrationTaskID	= "avm-interpreter"
	MigrationTaskAVMGasTable		MigrationTaskID	= "avm-gas-table"
	MigrationTaskContractStorageAdapter	MigrationTaskID	= "contract-storage-adapter"
	MigrationTaskContractMessageSyscalls	MigrationTaskID	= "contract-message-syscalls"
	MigrationTaskProofVerificationSyscalls	MigrationTaskID	= "proof-verification-syscalls"
	MigrationTaskABIRegistry		MigrationTaskID	= "abi-registry"

	MigrationTaskIdentityProofActivation	MigrationTaskID	= "identity-proof-activation"
	MigrationTaskCrossZoneIdentityLookup	MigrationTaskID	= "cross-zone-identity-lookup-messages"
	MigrationTaskPaymentChannelSettlement	MigrationTaskID	= "payment-channel-settlement"
	MigrationTaskConditionalPaymentRouting	MigrationTaskID	= "conditional-payment-routing"
	MigrationTaskPaymentProofAPIs		MigrationTaskID	= "payment-proof-apis"
	MigrationTaskWalletSDKIdentityPayment	MigrationTaskID	= "wallet-sdk-identity-payment-helpers"

	MigrationTaskBlockSTMZoneShardWorkloads	MigrationTaskID	= "blockstm-zone-shard-workloads"
	MigrationTaskConflictProfiling		MigrationTaskID	= "conflict-profiling"
	MigrationTaskStoreV2Benchmarks		MigrationTaskID	= "store-v2-benchmarks"
	MigrationTaskMempoolLanes		MigrationTaskID	= "mempool-lanes"
	MigrationTaskCongestionAwareRouting	MigrationTaskID	= "congestion-aware-routing"
	MigrationTaskAdaptiveSyncRecoveryTests	MigrationTaskID	= "adaptivesync-recovery-tests"
	MigrationTaskMultiZoneLoadSimulation	MigrationTaskID	= "multi-zone-traffic-load-simulation"

	MigrationExitSingleChainReproducibleExport	MigrationExitCriterionID	= "single-chain-state-reproducible-exportable"
	MigrationExitLegacyInvariantCoverage		MigrationExitCriterionID	= "legacy-module-invariant-coverage"
	MigrationExitSafePrefixMigration		MigrationExitCriterionID	= "safe-prefix-migration-upgrade-handlers"

	MigrationExitSingleZoneOperation	MigrationExitCriterionID	= "current-chain-operates-as-one-zone"
	MigrationExitAppHashCoreRoot		MigrationExitCriterionID	= "app-hash-includes-core-root-structure"
	MigrationExitProofRegistryRootMetadata	MigrationExitCriterionID	= "proof-registry-serves-root-metadata"

	MigrationExitMessagesFirstClassCommitted	MigrationExitCriterionID	= "messages-first-class-committed-objects"
	MigrationExitLocalAsyncDeterministic		MigrationExitCriterionID	= "local-async-messages-deterministic"
	MigrationExitMessageReceiptsProofQueryable	MigrationExitCriterionID	= "message-receipts-proof-queryable"

	MigrationExitFinancialZoneExecution	MigrationExitCriterionID	= "financial-modules-execute-in-financial-zone"
	MigrationExitIdentityZoneIsolated	MigrationExitCriterionID	= "identity-module-isolated-zone"
	MigrationExitZoneRootsPerBlock		MigrationExitCriterionID	= "zone-roots-committed-per-block"

	MigrationExitMultiShardZones				MigrationExitCriterionID	= "zones-run-with-multiple-shards"
	MigrationExitParallelShardWorkloads			MigrationExitCriterionID	= "independent-shard-workloads-parallel"
	MigrationExitInflightMessagesSurviveLayoutChanges	MigrationExitCriterionID	= "in-flight-messages-survive-layout-changes"

	MigrationExitContractZoneDeterministic		MigrationExitCriterionID	= "contract-zone-runs-deterministic-contracts"
	MigrationExitContractsEmitAsyncMessages		MigrationExitCriterionID	= "contracts-emit-async-messages"
	MigrationExitContractStateProofsAvailable	MigrationExitCriterionID	= "contract-state-proofs-available"

	MigrationExitNamesResolveProofBacked		MigrationExitCriterionID	= "names-resolve-proof-backed-identity-zone"
	MigrationExitPaymentsSettleTrustlessly		MigrationExitCriterionID	= "payments-settle-trustlessly-financial-zone"
	MigrationExitContractsUseIdentityPaymentsAsync	MigrationExitCriterionID	= "contracts-use-identity-payment-messages-async"

	MigrationExitParallelismScales			MigrationExitCriterionID	= "zone-shard-execution-scales-with-parallelism"
	MigrationExitStateSyncRecoversCommitments	MigrationExitCriterionID	= "state-sync-recovers-zone-shard-commitments"
	MigrationExitRoutingDeterministicCongestion	MigrationExitCriterionID	= "routing-deterministic-under-congestion"
)

type MigrationTaskDescriptor struct {
	PhaseID		MigrationPhaseID
	TaskID		MigrationTaskID
	Task		string
	Target		string
	Evidence	string
	DescriptorHash	string
}

type MigrationExitCriterion struct {
	PhaseID		MigrationPhaseID
	CriterionID	MigrationExitCriterionID
	Criterion	string
	Evidence	string
	DescriptorHash	string
}

type MigrationPhase struct {
	PhaseID		MigrationPhaseID
	Title		string
	Tasks		[]MigrationTaskDescriptor
	ExitCriteria	[]MigrationExitCriterion
	PhaseHash	string
}

type MigrationPathSpec struct {
	Version	uint64
	Phases	[]MigrationPhase
	Root	string
}

type BaselineHardeningEvidence struct {
	ModuleBoundaryDocsRoot		string
	StateExportManifestHash		string
	GenesisImportHash		string
	DynamicFeeBoundsTestHash	string
	LegacyInvariantRoot		string
	StoreV2AuditHash		string
	UpgradeHandlerPrefixHash	string
	StateReproducible		bool
	StateExportable			bool
	InvariantCoverage		bool
	PrefixMigrationSafe		bool
	EvidenceHash			string
}

type CoreCommitmentMigrationEvidence struct {
	AetraCoreModuleHash		string
	DefaultZoneDescriptorHash	string
	DefaultZoneStateRoot		string
	EmptyMessageRoot		string
	ProofRegistryRoot		string
	RootQueryAPIHash		string
	AppHashCoreRoot			string
	DefaultZoneID			ZoneID
	SingleZoneMode			bool
	ProofRegistryMetadata		bool
	EvidenceHash			string
}

type MessageBusMigrationEvidence struct {
	MsgbusModuleHash	string
	MessageCodecHash	string
	MessageIDDerivationHash	string
	InboxStoreRoot		string
	OutboxStoreRoot		string
	ReceiptStoreRoot	string
	LocalExecutionRoot	string
	ExpiryBounceRoot	string
	InclusionProofRoot	string
	MessagesCommitted	bool
	LocalAsyncDeterministic	bool
	ReceiptProofQueryable	bool
	EvidenceHash		string
}

type ZoneExtractionMigrationEvidence struct {
	FinancialZoneRoot		string
	IdentityZoneRoot		string
	ApplicationZoneRoot		string
	ZoneKeeperRoot			string
	ZonePrefixRoot			string
	ZoneFeePolicyRoot		string
	ZoneExecutionSummaryRoot	string
	FinancialModulesRouted		bool
	IdentityIsolated		bool
	ZoneRootsCommittedPerBlock	bool
	EvidenceHash			string
}

type ShardingRuntimeMigrationEvidence struct {
	ShardsModuleHash	string
	ShardLayoutRoot		string
	RouteKeyCalculationRoot	string
	PerShardInboxRoot	string
	PerShardOutboxRoot	string
	ShardRootAggregate	string
	SplitMergeScheduleRoot	string
	ShardMigrationRoot	string
	MultiShardZones		bool
	ParallelShardWorkloads	bool
	InflightMessagesSafe	bool
	EvidenceHash		string
}

type AVM20MigrationEvidence struct {
	BytecodeFormatHash	string
	InterpreterRoot		string
	GasTableRoot		string
	ContractStorageRoot	string
	MessageSyscallRoot	string
	ProofSyscallRoot	string
	ABIRegistryRoot		string
	DeterministicContracts	bool
	AsyncMessages		bool
	ContractProofsAvailable	bool
	EvidenceHash		string
}

type IdentityPaymentIntegrationEvidence struct {
	IdentityProofRoot		string
	IdentityLookupMessageRoot	string
	PaymentChannelSettlementRoot	string
	ConditionalPaymentRouteRoot	string
	PaymentProofAPIRoot		string
	WalletSDKHelperRoot		string
	ProofBackedNames		bool
	TrustlessPayments		bool
	AsyncContractMessages		bool
	EvidenceHash			string
}

type PerformanceHardeningMigrationEvidence struct {
	BlockSTMWorkloadRoot		string
	ConflictProfileRoot		string
	StoreV2BenchmarkRoot		string
	MempoolLaneRoot			string
	CongestionRoutingRoot		string
	AdaptiveSyncRecoveryRoot	string
	LoadSimulationRoot		string
	ParallelismScales		bool
	StateSyncRecoversRoots		bool
	CongestionDeterministic		bool
	EvidenceHash			string
}

func DefaultMigrationPathSpec() (MigrationPathSpec, error) {
	return BuildMigrationPathSpec([]MigrationPhase{
		migrationPhase(MigrationPhaseBaselineHardening, "Phase 0: Baseline Hardening", MigrationPhase0Tasks(), MigrationPhase0ExitCriteria()),
		migrationPhase(MigrationPhaseCoreCommitments, "Phase 1: Core Commitments", MigrationPhase1Tasks(), MigrationPhase1ExitCriteria()),
		migrationPhase(MigrationPhaseMessageBus, "Phase 2: Message Bus", MigrationPhase2Tasks(), MigrationPhase2ExitCriteria()),
		migrationPhase(MigrationPhaseZoneExtraction, "Phase 3: Zone Extraction", MigrationPhase3Tasks(), MigrationPhase3ExitCriteria()),
		migrationPhase(MigrationPhaseShardingRuntime, "Phase 4: Sharding Runtime", MigrationPhase4Tasks(), MigrationPhase4ExitCriteria()),
		migrationPhase(MigrationPhaseAVM20, "Phase 5: AVM 2.0", MigrationPhase5Tasks(), MigrationPhase5ExitCriteria()),
		migrationPhase(MigrationPhaseIdentityPayments, "Phase 6: Identity and Payment Integration", MigrationPhase6Tasks(), MigrationPhase6ExitCriteria()),
		migrationPhase(MigrationPhasePerformanceHardening, "Phase 7: Performance Hardening", MigrationPhase7Tasks(), MigrationPhase7ExitCriteria()),
	})
}

func BuildMigrationPathSpec(phases []MigrationPhase) (MigrationPathSpec, error) {
	spec := MigrationPathSpec{
		Version:	MigrationPathSpecVersion,
		Phases:		normalizeMigrationPhases(phases),
	}
	if err := spec.ValidateFormat(); err != nil {
		return MigrationPathSpec{}, err
	}
	spec.Root = ComputeMigrationPathSpecRoot(spec.Phases)
	return spec, spec.Validate()
}

func MigrationPhase0Tasks() []MigrationTaskDescriptor {
	return []MigrationTaskDescriptor{
		migrationTask(MigrationPhaseBaselineHardening, MigrationTaskModuleBoundaryDocs, "Finalize current module boundary documentation.", "module boundary manifest", "zone/core boundary matrix;module ownership table"),
		migrationTask(MigrationPhaseBaselineHardening, MigrationTaskStateExportValidation, "Add state export validation.", "ExportManifest", "ExportManifest.ValidateHash;state root snapshots"),
		migrationTask(MigrationPhaseBaselineHardening, MigrationTaskDeterministicGenesis, "Add deterministic genesis import for all active modules.", "genesis import replay", "canonical module import order;genesis import hash"),
		migrationTask(MigrationPhaseBaselineHardening, MigrationTaskDynamicFeeBoundsTests, "Add dynamic fee bounds tests.", "fee module tests", "min/max fee bounds;forwarding fee escrow bounds"),
		migrationTask(MigrationPhaseBaselineHardening, MigrationTaskLegacyModuleInvariants, "Add staking, slashing, bank, and distribution invariants.", "legacy invariant registry", "staking;slashing;bank;distribution invariant root"),
		migrationTask(MigrationPhaseBaselineHardening, MigrationTaskStoreV2CompatibilityAudit, "Add Store v2 compatibility audit.", "Store v2 audit report", "prefix proof;bounded range scan;object store compatibility"),
	}
}

func MigrationPhase1Tasks() []MigrationTaskDescriptor {
	return []MigrationTaskDescriptor{
		migrationTask(MigrationPhaseCoreCommitments, MigrationTaskAetraCoreModule, "Implement x/aetracore.", "x/aetracore", "CoreState;ZoneDescriptor;ZoneCommitment;RootSnapshot"),
		migrationTask(MigrationPhaseCoreCommitments, MigrationTaskDefaultZoneRegistry, "Add zone registry with one default zone.", "zone registry", "one enabled default ZoneDescriptor"),
		migrationTask(MigrationPhaseCoreCommitments, MigrationTaskDefaultZoneStateRoot, "Commit default zone state root.", "GlobalStateRoot", "default ZoneCommitment.zone_state_root"),
		migrationTask(MigrationPhaseCoreCommitments, MigrationTaskEmptyMessageRoot, "Commit message root with empty queues.", "GlobalMessageRoot", "EmptyRootHash inbox and outbox queues"),
		migrationTask(MigrationPhaseCoreCommitments, MigrationTaskProofRootRegistry, "Add proof root registry.", "ProofRoot registry", "state;message;zone;receipt root metadata"),
		migrationTask(MigrationPhaseCoreCommitments, MigrationTaskRootQueryAPIs, "Add root query APIs.", "root query surface", "height-scoped root snapshot and proof metadata queries"),
	}
}

func MigrationPhase2Tasks() []MigrationTaskDescriptor {
	return []MigrationTaskDescriptor{
		migrationTask(MigrationPhaseMessageBus, MigrationTaskMsgbusModule, "Implement x/msgbus.", "x/msgbus", "message module;zone adapters;deterministic stores"),
		migrationTask(MigrationPhaseMessageBus, MigrationTaskMessageEncodingAndIDs, "Add message encoding and IDs.", "AetherMessage codec", "canonical encoding;payload hash;nonce;route commitment;msg_id"),
		migrationTask(MigrationPhaseMessageBus, MigrationTaskMessageStores, "Add inbox, outbox, and receipt stores.", "message stores", "zone and shard inbox/outbox/receipt prefixes with committed roots"),
		migrationTask(MigrationPhaseMessageBus, MigrationTaskLocalZoneExecution, "Add local-zone message execution.", "message executor", "local async delivery;ApplyInboundMessage;deterministic receipt"),
		migrationTask(MigrationPhaseMessageBus, MigrationTaskExpiryBounceLogic, "Add expiry and bounce logic.", "delivery executor", "expired, bounced, failed, refunded, and rejected receipts"),
		migrationTask(MigrationPhaseMessageBus, MigrationTaskMessageInclusionProofs, "Add message inclusion proofs.", "message proof queries", "source outbox, destination inbox, receipt, and global message proofs"),
	}
}

func MigrationPhase3Tasks() []MigrationTaskDescriptor {
	return []MigrationTaskDescriptor{
		migrationTask(MigrationPhaseZoneExtraction, MigrationTaskExtractFinancialZone, "Extract Financial Zone.", "Financial Zone adapter", "bank;fees;contract-assets;dex prefixes and roots"),
		migrationTask(MigrationPhaseZoneExtraction, MigrationTaskExtractIdentityZone, "Extract Identity Zone.", "Identity Zone adapter", "identity resolver;reverse lookup;delegation roots"),
		migrationTask(MigrationPhaseZoneExtraction, MigrationTaskExtractApplicationZone, "Extract Application Zone.", "Application Zone adapter", "workflows;scheduler;app queues;permission roots"),
		migrationTask(MigrationPhaseZoneExtraction, MigrationTaskZoneSpecificKeepers, "Add zone-specific keepers and state prefixes.", "zone keeper wiring", "keeper scopes;state prefixes;zone export manifests"),
		migrationTask(MigrationPhaseZoneExtraction, MigrationTaskZoneLocalFeePolicies, "Add zone-local fee policies.", "zone fee policy registry", "financial fee roots;zone fee policy IDs;aggregation records"),
		migrationTask(MigrationPhaseZoneExtraction, MigrationTaskZoneExecutionSummaries, "Add zone execution summaries.", "ZoneExecutionSummary", "tx counts;message counts;gas;roots;summary hash"),
	}
}

func MigrationPhase4Tasks() []MigrationTaskDescriptor {
	return []MigrationTaskDescriptor{
		migrationTask(MigrationPhaseShardingRuntime, MigrationTaskShardsModule, "Implement x/shards.", "x/shards", "shard layouts;metrics;migration executor;zone adapters"),
		migrationTask(MigrationPhaseShardingRuntime, MigrationTaskShardLayoutDescriptors, "Add shard layout descriptors.", "ShardLayout and ShardDescriptor", "active shards;assignment modes;activation height;layout hash"),
		migrationTask(MigrationPhaseShardingRuntime, MigrationTaskRouteKeyCalculation, "Add route-key calculation.", "RouteKeyToShard", "zone_id;state_key;layout_epoch;committed shard layout"),
		migrationTask(MigrationPhaseShardingRuntime, MigrationTaskPerShardMessageStores, "Add per-shard inbox and outbox.", "shard message stores", "zone/{zone_id}/shard/{shard_id}/inbox and outbox roots"),
		migrationTask(MigrationPhaseShardingRuntime, MigrationTaskShardRootAggregation, "Add shard root aggregation.", "zone root aggregation", "shard_state_root;shard_roots_root;ZoneCommitment"),
		migrationTask(MigrationPhaseShardingRuntime, MigrationTaskSplitMergeScheduler, "Add split and merge scheduler.", "layout transition planner", "committed metrics;future layout_epoch;no mid-block routing changes"),
		migrationTask(MigrationPhaseShardingRuntime, MigrationTaskDeterministicShardMigration, "Add deterministic shard migration.", "migration executor", "migration tasks;receipts;in-flight message delivery epoch"),
	}
}

func MigrationPhase5Tasks() []MigrationTaskDescriptor {
	return []MigrationTaskDescriptor{
		migrationTask(MigrationPhaseAVM20, MigrationTaskAVMBytecodeFormat, "Implement AVM bytecode format.", "AVM bytecode codec", "canonical;versioned;hashable;malformed opcode rejection"),
		migrationTask(MigrationPhaseAVM20, MigrationTaskAVMInterpreter, "Implement interpreter.", "AVM runtime", "deterministic stack execution;bounded memory;replay-safe outputs"),
		migrationTask(MigrationPhaseAVM20, MigrationTaskAVMGasTable, "Implement gas table.", "AVM metering profile", "opcode;memory;storage;proof;message;event gas costs"),
		migrationTask(MigrationPhaseAVM20, MigrationTaskContractStorageAdapter, "Implement contract storage adapter.", "Contract Zone Store v2 adapter", "contract prefixes;bounded reads and writes;storage root"),
		migrationTask(MigrationPhaseAVM20, MigrationTaskContractMessageSyscalls, "Implement contract message syscalls.", "MSG_* syscalls", "prepaid async AetherMessage outputs;no remote mutation"),
		migrationTask(MigrationPhaseAVM20, MigrationTaskProofVerificationSyscalls, "Implement proof verification syscalls.", "VERIFY_* syscalls", "metered Merkle, message, zone, and signature proof checks"),
		migrationTask(MigrationPhaseAVM20, MigrationTaskABIRegistry, "Implement ABI registry.", "ABI descriptor registry", "canonical ABI descriptors;interface hash;code ID binding"),
	}
}

func MigrationPhase6Tasks() []MigrationTaskDescriptor {
	return []MigrationTaskDescriptor{
		migrationTask(MigrationPhaseIdentityPayments, MigrationTaskIdentityProofActivation, "Activate .aet identity proofs.", "Identity Zone proof API", "domain ownership;resolver;reverse;delegation;expiry proof roots"),
		migrationTask(MigrationPhaseIdentityPayments, MigrationTaskCrossZoneIdentityLookup, "Add cross-zone identity lookup messages.", "MsgResolveIdentity and MsgIdentityResolutionResult", "async requests;proof-required replies;receipt-backed results"),
		migrationTask(MigrationPhaseIdentityPayments, MigrationTaskPaymentChannelSettlement, "Add payment channel settlement.", "Financial Zone payment channels", "collateral escrow;latest signed state;challenge period;settlement proof"),
		migrationTask(MigrationPhaseIdentityPayments, MigrationTaskConditionalPaymentRouting, "Add conditional payment routing.", "conditional payment routes", "hash locks;timeouts;route commitments;linked condition settlement"),
		migrationTask(MigrationPhaseIdentityPayments, MigrationTaskPaymentProofAPIs, "Add payment proof APIs.", "payment proof queries", "channel;condition;route;settlement;dispute;receipt proofs"),
		migrationTask(MigrationPhaseIdentityPayments, MigrationTaskWalletSDKIdentityPayment, "Add wallet SDK helpers for identity and payment flows.", "wallet SDK", "send-by-name;invoke-by-name;proof-bound payment routes"),
	}
}

func MigrationPhase7Tasks() []MigrationTaskDescriptor {
	return []MigrationTaskDescriptor{
		migrationTask(MigrationPhasePerformanceHardening, MigrationTaskBlockSTMZoneShardWorkloads, "Enable BlockSTM workloads for zone and shard batches.", "BlockSTM execution planner", "disjoint zone and shard batches;conflict keys;parallel execution metrics"),
		migrationTask(MigrationPhasePerformanceHardening, MigrationTaskConflictProfiling, "Add conflict profiling.", "conflict profiler", "state access conflicts;hot object counters;committed conflict profile root"),
		migrationTask(MigrationPhasePerformanceHardening, MigrationTaskStoreV2Benchmarks, "Add Store v2 benchmarks.", "Store v2 benchmark suite", "balance;identity;contract;message;payment;AVM AMM;proof generation benchmarks"),
		migrationTask(MigrationPhasePerformanceHardening, MigrationTaskMempoolLanes, "Add mempool lanes.", "zonemempool lanes", "zone lanes;shard sublanes;message class priority;DoS limits"),
		migrationTask(MigrationPhasePerformanceHardening, MigrationTaskCongestionAwareRouting, "Add congestion-aware routing.", "routing cost model", "committed congestion metrics;deterministic scoring;bounded fairness"),
		migrationTask(MigrationPhasePerformanceHardening, MigrationTaskAdaptiveSyncRecoveryTests, "Add AdaptiveSync recovery tests.", "state sync recovery", "zone and shard commitment replay;root snapshot recovery"),
		migrationTask(MigrationPhasePerformanceHardening, MigrationTaskMultiZoneLoadSimulation, "Add load simulation for multi-zone traffic.", "load simulator", "cross-zone messages;payments;identity;contracts;congestion and expiry traffic"),
	}
}

func MigrationPhase0ExitCriteria() []MigrationExitCriterion {
	return []MigrationExitCriterion{
		migrationExitCriterion(MigrationPhaseBaselineHardening, MigrationExitSingleChainReproducibleExport, "Existing single-chain state is reproducible and exportable.", "state export manifest and genesis import hashes match after replay"),
		migrationExitCriterion(MigrationPhaseBaselineHardening, MigrationExitLegacyInvariantCoverage, "Existing modules have invariant coverage.", "staking, slashing, bank, and distribution invariants are present in the invariant root"),
		migrationExitCriterion(MigrationPhaseBaselineHardening, MigrationExitSafePrefixMigration, "Upgrade handlers can migrate state prefixes safely.", "upgrade handler prefix migration plan is hash-committed and replay-safe"),
	}
}

func MigrationPhase1ExitCriteria() []MigrationExitCriterion {
	return []MigrationExitCriterion{
		migrationExitCriterion(MigrationPhaseCoreCommitments, MigrationExitSingleZoneOperation, "Current chain operates as one zone.", "exactly one enabled default zone is committed in the registry"),
		migrationExitCriterion(MigrationPhaseCoreCommitments, MigrationExitAppHashCoreRoot, "App hash includes core root structure.", "app hash binds aether core, zone, message, receipt, and proof roots"),
		migrationExitCriterion(MigrationPhaseCoreCommitments, MigrationExitProofRegistryRootMetadata, "Proof registry serves root metadata.", "proof root metadata is queryable by height and root type"),
	}
}

func MigrationPhase2ExitCriteria() []MigrationExitCriterion {
	return []MigrationExitCriterion{
		migrationExitCriterion(MigrationPhaseMessageBus, MigrationExitMessagesFirstClassCommitted, "Messages are first-class committed objects.", "message envelopes, inboxes, outboxes, receipts, and replay state commit under deterministic roots"),
		migrationExitCriterion(MigrationPhaseMessageBus, MigrationExitLocalAsyncDeterministic, "Local async messages execute deterministically.", "local-zone async delivery produces identical receipts and roots on replay"),
		migrationExitCriterion(MigrationPhaseMessageBus, MigrationExitMessageReceiptsProofQueryable, "Message receipts are proof-queryable.", "receipt query returns height-scoped proof metadata under receipt and message roots"),
	}
}

func MigrationPhase3ExitCriteria() []MigrationExitCriterion {
	return []MigrationExitCriterion{
		migrationExitCriterion(MigrationPhaseZoneExtraction, MigrationExitFinancialZoneExecution, "Existing bank, fees, contract-assets, and AVM AMM contracts execute in Financial Zone.", "financial zone roots include bank, fee, contract-assets, and AVM AMM state transitions"),
		migrationExitCriterion(MigrationPhaseZoneExtraction, MigrationExitIdentityZoneIsolated, "Identity module can activate as isolated Identity Zone.", "identity state writes are restricted to identity zone prefixes and roots"),
		migrationExitCriterion(MigrationPhaseZoneExtraction, MigrationExitZoneRootsPerBlock, "Zone roots are committed per block.", "each finalized height commits zone roots, summaries, inbox, outbox, receipt, and event roots"),
	}
}

func MigrationPhase4ExitCriteria() []MigrationExitCriterion {
	return []MigrationExitCriterion{
		migrationExitCriterion(MigrationPhaseShardingRuntime, MigrationExitMultiShardZones, "Zones can run with multiple shards.", "zone descriptors reference committed shard layouts with more than one active shard"),
		migrationExitCriterion(MigrationPhaseShardingRuntime, MigrationExitParallelShardWorkloads, "Independent shard workloads execute in parallel.", "disjoint shard batches carry separate roots and no shared write locks"),
		migrationExitCriterion(MigrationPhaseShardingRuntime, MigrationExitInflightMessagesSurviveLayoutChanges, "In-flight messages survive shard layout changes.", "messages retain source metadata and deliver by committed delivery epoch after split or merge"),
	}
}

func MigrationPhase5ExitCriteria() []MigrationExitCriterion {
	return []MigrationExitCriterion{
		migrationExitCriterion(MigrationPhaseAVM20, MigrationExitContractZoneDeterministic, "Contract Zone runs deterministic contracts.", "same bytecode, state root, context, and input produce identical roots, receipts, gas, events, and messages"),
		migrationExitCriterion(MigrationPhaseAVM20, MigrationExitContractsEmitAsyncMessages, "Contracts can emit async messages.", "MSG_* syscalls emit prepaid AetherMessage outbox entries without remote state mutation"),
		migrationExitCriterion(MigrationPhaseAVM20, MigrationExitContractStateProofsAvailable, "Contract state proofs are available.", "code, instance, storage, ABI, event, inbox, and outbox proofs bind to Contract Zone roots"),
	}
}

func MigrationPhase6ExitCriteria() []MigrationExitCriterion {
	return []MigrationExitCriterion{
		migrationExitCriterion(MigrationPhaseIdentityPayments, MigrationExitNamesResolveProofBacked, "Names resolve through proof-backed Identity Zone data.", "resolver replies include Identity Zone proof roots or committed identity receipts"),
		migrationExitCriterion(MigrationPhaseIdentityPayments, MigrationExitPaymentsSettleTrustlessly, "Payments settle trustlessly through Financial Zone.", "payment channels, conditions, disputes, refunds, and settlement proofs bind to Financial Zone roots"),
		migrationExitCriterion(MigrationPhaseIdentityPayments, MigrationExitContractsUseIdentityPaymentsAsync, "Contracts can use identity and payment messages asynchronously.", "AVM contracts emit identity lookup and payment route messages and consume receipt-backed replies"),
	}
}

func MigrationPhase7ExitCriteria() []MigrationExitCriterion {
	return []MigrationExitCriterion{
		migrationExitCriterion(MigrationPhasePerformanceHardening, MigrationExitParallelismScales, "Independent zone and shard execution scales with available parallelism.", "BlockSTM and load simulation metrics show disjoint batches scaling without global write locks"),
		migrationExitCriterion(MigrationPhasePerformanceHardening, MigrationExitStateSyncRecoversCommitments, "State sync can recover zone and shard commitments.", "AdaptiveSync recovery tests replay root snapshots, zone roots, shard roots, and proof registry metadata"),
		migrationExitCriterion(MigrationPhasePerformanceHardening, MigrationExitRoutingDeterministicCongestion, "Routing remains deterministic under congestion.", "routing decisions use committed congestion metrics and deterministic tie-breaks under load"),
	}
}

func (s MigrationPathSpec) Normalize() MigrationPathSpec {
	if s.Version == 0 {
		s.Version = MigrationPathSpecVersion
	}
	s.Phases = normalizeMigrationPhases(s.Phases)
	s.Root = normalizePerformanceHash(s.Root)
	return s
}

func (s MigrationPathSpec) ValidateFormat() error {
	s = s.Normalize()
	if s.Version != MigrationPathSpecVersion {
		return fmt.Errorf("aetracore migration path spec version must be %d", MigrationPathSpecVersion)
	}
	if len(s.Phases) != 8 {
		return errors.New("aetracore migration path spec requires phases 0 through 7")
	}
	seen := make(map[MigrationPhaseID]struct{}, len(s.Phases))
	var previous MigrationPhaseID
	for i, phase := range s.Phases {
		if err := phase.Validate(); err != nil {
			return err
		}
		if _, found := seen[phase.PhaseID]; found {
			return fmt.Errorf("duplicate aetracore migration phase %s", phase.PhaseID)
		}
		seen[phase.PhaseID] = struct{}{}
		if i > 0 && previous >= phase.PhaseID {
			return errors.New("aetracore migration phases must be sorted canonically")
		}
		previous = phase.PhaseID
	}
	if _, found := seen[MigrationPhaseBaselineHardening]; !found {
		return errors.New("aetracore migration path missing phase 0 baseline hardening")
	}
	if _, found := seen[MigrationPhaseCoreCommitments]; !found {
		return errors.New("aetracore migration path missing phase 1 core commitments")
	}
	if _, found := seen[MigrationPhaseMessageBus]; !found {
		return errors.New("aetracore migration path missing phase 2 message bus")
	}
	if _, found := seen[MigrationPhaseZoneExtraction]; !found {
		return errors.New("aetracore migration path missing phase 3 zone extraction")
	}
	if _, found := seen[MigrationPhaseShardingRuntime]; !found {
		return errors.New("aetracore migration path missing phase 4 sharding runtime")
	}
	if _, found := seen[MigrationPhaseAVM20]; !found {
		return errors.New("aetracore migration path missing phase 5 AVM 2.0")
	}
	if _, found := seen[MigrationPhaseIdentityPayments]; !found {
		return errors.New("aetracore migration path missing phase 6 identity and payment integration")
	}
	if _, found := seen[MigrationPhasePerformanceHardening]; !found {
		return errors.New("aetracore migration path missing phase 7 performance hardening")
	}
	if s.Root != "" {
		if err := ValidateHash("aetracore migration path spec root", s.Root); err != nil {
			return err
		}
	}
	return nil
}

func (s MigrationPathSpec) Validate() error {
	s = s.Normalize()
	if err := s.ValidateFormat(); err != nil {
		return err
	}
	if s.Root == "" {
		return errors.New("aetracore migration path spec root is required")
	}
	expected := ComputeMigrationPathSpecRoot(s.Phases)
	if s.Root != expected {
		return fmt.Errorf("aetracore migration path spec root mismatch: expected %s", expected)
	}
	return nil
}

func (p MigrationPhase) Normalize() MigrationPhase {
	p.Title = compactPerformanceText(p.Title)
	p.Tasks = normalizeMigrationTasks(p.Tasks)
	p.ExitCriteria = normalizeMigrationExitCriteria(p.ExitCriteria)
	p.PhaseHash = normalizePerformanceHash(p.PhaseHash)
	return p
}

func (p MigrationPhase) ValidateFormat() error {
	p = p.Normalize()
	if !IsMigrationPhaseID(p.PhaseID) {
		return fmt.Errorf("unknown aetracore migration phase %q", p.PhaseID)
	}
	if p.Title == "" {
		return errors.New("aetracore migration phase title is required")
	}
	if len(p.Tasks) == 0 || len(p.ExitCriteria) == 0 {
		return errors.New("aetracore migration phase requires tasks and exit criteria")
	}
	seenTasks := make(map[MigrationTaskID]struct{}, len(p.Tasks))
	var previousTask MigrationTaskID
	for i, task := range p.Tasks {
		if err := task.Validate(); err != nil {
			return err
		}
		if task.PhaseID != p.PhaseID {
			return errors.New("aetracore migration task phase mismatch")
		}
		if _, found := seenTasks[task.TaskID]; found {
			return fmt.Errorf("duplicate aetracore migration task %s", task.TaskID)
		}
		seenTasks[task.TaskID] = struct{}{}
		if i > 0 && previousTask >= task.TaskID {
			return errors.New("aetracore migration tasks must be sorted canonically")
		}
		previousTask = task.TaskID
	}
	seenCriteria := make(map[MigrationExitCriterionID]struct{}, len(p.ExitCriteria))
	var previousCriterion MigrationExitCriterionID
	for i, criterion := range p.ExitCriteria {
		if err := criterion.Validate(); err != nil {
			return err
		}
		if criterion.PhaseID != p.PhaseID {
			return errors.New("aetracore migration exit criterion phase mismatch")
		}
		if _, found := seenCriteria[criterion.CriterionID]; found {
			return fmt.Errorf("duplicate aetracore migration exit criterion %s", criterion.CriterionID)
		}
		seenCriteria[criterion.CriterionID] = struct{}{}
		if i > 0 && previousCriterion >= criterion.CriterionID {
			return errors.New("aetracore migration exit criteria must be sorted canonically")
		}
		previousCriterion = criterion.CriterionID
	}
	if p.PhaseHash != "" {
		if err := ValidateHash("aetracore migration phase hash", p.PhaseHash); err != nil {
			return err
		}
	}
	return nil
}

func (p MigrationPhase) Validate() error {
	p = p.Normalize()
	if err := p.ValidateFormat(); err != nil {
		return err
	}
	if p.PhaseHash == "" {
		return errors.New("aetracore migration phase hash is required")
	}
	expected := ComputeMigrationPhaseHash(p)
	if p.PhaseHash != expected {
		return fmt.Errorf("aetracore migration phase hash mismatch: expected %s", expected)
	}
	return nil
}

func (d MigrationTaskDescriptor) Normalize() MigrationTaskDescriptor {
	d.Task = compactPerformanceText(d.Task)
	d.Target = compactPerformanceText(d.Target)
	d.Evidence = compactPerformanceText(d.Evidence)
	d.DescriptorHash = normalizePerformanceHash(d.DescriptorHash)
	return d
}

func (d MigrationTaskDescriptor) ValidateFormat() error {
	d = d.Normalize()
	if !IsMigrationPhaseID(d.PhaseID) {
		return fmt.Errorf("unknown aetracore migration task phase %q", d.PhaseID)
	}
	if !IsMigrationTaskID(d.PhaseID, d.TaskID) {
		return fmt.Errorf("unknown aetracore migration task %q", d.TaskID)
	}
	if d.Task == "" || d.Target == "" || d.Evidence == "" {
		return errors.New("aetracore migration task requires task, target, and evidence")
	}
	if d.DescriptorHash != "" {
		if err := ValidateHash("aetracore migration task descriptor hash", d.DescriptorHash); err != nil {
			return err
		}
	}
	return nil
}

func (d MigrationTaskDescriptor) Validate() error {
	d = d.Normalize()
	if err := d.ValidateFormat(); err != nil {
		return err
	}
	if d.DescriptorHash == "" {
		return errors.New("aetracore migration task descriptor hash is required")
	}
	expected := ComputeMigrationTaskHash(d)
	if d.DescriptorHash != expected {
		return fmt.Errorf("aetracore migration task descriptor hash mismatch: expected %s", expected)
	}
	return nil
}

func (c MigrationExitCriterion) Normalize() MigrationExitCriterion {
	c.Criterion = compactPerformanceText(c.Criterion)
	c.Evidence = compactPerformanceText(c.Evidence)
	c.DescriptorHash = normalizePerformanceHash(c.DescriptorHash)
	return c
}

func (c MigrationExitCriterion) ValidateFormat() error {
	c = c.Normalize()
	if !IsMigrationPhaseID(c.PhaseID) {
		return fmt.Errorf("unknown aetracore migration exit criterion phase %q", c.PhaseID)
	}
	if !IsMigrationExitCriterionID(c.PhaseID, c.CriterionID) {
		return fmt.Errorf("unknown aetracore migration exit criterion %q", c.CriterionID)
	}
	if c.Criterion == "" || c.Evidence == "" {
		return errors.New("aetracore migration exit criterion requires criterion and evidence")
	}
	if c.DescriptorHash != "" {
		if err := ValidateHash("aetracore migration exit criterion descriptor hash", c.DescriptorHash); err != nil {
			return err
		}
	}
	return nil
}

func (c MigrationExitCriterion) Validate() error {
	c = c.Normalize()
	if err := c.ValidateFormat(); err != nil {
		return err
	}
	if c.DescriptorHash == "" {
		return errors.New("aetracore migration exit criterion descriptor hash is required")
	}
	expected := ComputeMigrationExitCriterionHash(c)
	if c.DescriptorHash != expected {
		return fmt.Errorf("aetracore migration exit criterion descriptor hash mismatch: expected %s", expected)
	}
	return nil
}

func (e BaselineHardeningEvidence) Normalize() BaselineHardeningEvidence {
	e.ModuleBoundaryDocsRoot = normalizePerformanceHash(e.ModuleBoundaryDocsRoot)
	e.StateExportManifestHash = normalizePerformanceHash(e.StateExportManifestHash)
	e.GenesisImportHash = normalizePerformanceHash(e.GenesisImportHash)
	e.DynamicFeeBoundsTestHash = normalizePerformanceHash(e.DynamicFeeBoundsTestHash)
	e.LegacyInvariantRoot = normalizePerformanceHash(e.LegacyInvariantRoot)
	e.StoreV2AuditHash = normalizePerformanceHash(e.StoreV2AuditHash)
	e.UpgradeHandlerPrefixHash = normalizePerformanceHash(e.UpgradeHandlerPrefixHash)
	e.EvidenceHash = normalizePerformanceHash(e.EvidenceHash)
	return e
}

func (e BaselineHardeningEvidence) ValidateFormat() error {
	e = e.Normalize()
	hashes := []struct {
		name	string
		value	string
	}{
		{"aetracore migration module boundary docs root", e.ModuleBoundaryDocsRoot},
		{"aetracore migration state export manifest hash", e.StateExportManifestHash},
		{"aetracore migration genesis import hash", e.GenesisImportHash},
		{"aetracore migration dynamic fee bounds test hash", e.DynamicFeeBoundsTestHash},
		{"aetracore migration legacy invariant root", e.LegacyInvariantRoot},
		{"aetracore migration Store v2 audit hash", e.StoreV2AuditHash},
		{"aetracore migration upgrade handler prefix hash", e.UpgradeHandlerPrefixHash},
	}
	for _, item := range hashes {
		if err := ValidateHash(item.name, item.value); err != nil {
			return err
		}
	}
	if !e.StateReproducible || !e.StateExportable {
		return errors.New("aetracore migration baseline evidence requires reproducible and exportable state")
	}
	if !e.InvariantCoverage {
		return errors.New("aetracore migration baseline evidence requires invariant coverage")
	}
	if !e.PrefixMigrationSafe {
		return errors.New("aetracore migration baseline evidence requires safe prefix migration")
	}
	if e.EvidenceHash != "" {
		if err := ValidateHash("aetracore migration baseline evidence hash", e.EvidenceHash); err != nil {
			return err
		}
	}
	return nil
}

func (e BaselineHardeningEvidence) Validate() error {
	e = e.Normalize()
	if err := e.ValidateFormat(); err != nil {
		return err
	}
	if e.EvidenceHash == "" {
		return errors.New("aetracore migration baseline evidence hash is required")
	}
	expected := ComputeBaselineHardeningEvidenceHash(e)
	if e.EvidenceHash != expected {
		return fmt.Errorf("aetracore migration baseline evidence hash mismatch: expected %s", expected)
	}
	return nil
}

func (e CoreCommitmentMigrationEvidence) Normalize() CoreCommitmentMigrationEvidence {
	e.AetraCoreModuleHash = normalizePerformanceHash(e.AetraCoreModuleHash)
	e.DefaultZoneDescriptorHash = normalizePerformanceHash(e.DefaultZoneDescriptorHash)
	e.DefaultZoneStateRoot = normalizePerformanceHash(e.DefaultZoneStateRoot)
	e.EmptyMessageRoot = normalizePerformanceHash(e.EmptyMessageRoot)
	e.ProofRegistryRoot = normalizePerformanceHash(e.ProofRegistryRoot)
	e.RootQueryAPIHash = normalizePerformanceHash(e.RootQueryAPIHash)
	e.AppHashCoreRoot = normalizePerformanceHash(e.AppHashCoreRoot)
	e.EvidenceHash = normalizePerformanceHash(e.EvidenceHash)
	return e
}

func (e CoreCommitmentMigrationEvidence) ValidateFormat() error {
	e = e.Normalize()
	hashes := []struct {
		name	string
		value	string
	}{
		{"aetracore migration module hash", e.AetraCoreModuleHash},
		{"aetracore migration default zone descriptor hash", e.DefaultZoneDescriptorHash},
		{"aetracore migration default zone state root", e.DefaultZoneStateRoot},
		{"aetracore migration empty message root", e.EmptyMessageRoot},
		{"aetracore migration proof registry root", e.ProofRegistryRoot},
		{"aetracore migration root query API hash", e.RootQueryAPIHash},
		{"aetracore migration app hash core root", e.AppHashCoreRoot},
	}
	for _, item := range hashes {
		if err := ValidateHash(item.name, item.value); err != nil {
			return err
		}
	}
	if err := validateZoneID(string(e.DefaultZoneID)); err != nil {
		return err
	}
	if e.EmptyMessageRoot != EmptyRootHash {
		return errors.New("aetracore migration core evidence requires empty message queues to use EmptyRootHash")
	}
	if !e.SingleZoneMode {
		return errors.New("aetracore migration core evidence requires single-zone operation")
	}
	if !e.ProofRegistryMetadata {
		return errors.New("aetracore migration core evidence requires proof registry root metadata")
	}
	if e.EvidenceHash != "" {
		if err := ValidateHash("aetracore migration core evidence hash", e.EvidenceHash); err != nil {
			return err
		}
	}
	return nil
}

func (e CoreCommitmentMigrationEvidence) Validate() error {
	e = e.Normalize()
	if err := e.ValidateFormat(); err != nil {
		return err
	}
	if e.EvidenceHash == "" {
		return errors.New("aetracore migration core evidence hash is required")
	}
	expected := ComputeCoreCommitmentMigrationEvidenceHash(e)
	if e.EvidenceHash != expected {
		return fmt.Errorf("aetracore migration core evidence hash mismatch: expected %s", expected)
	}
	return nil
}

func (e MessageBusMigrationEvidence) Normalize() MessageBusMigrationEvidence {
	e.MsgbusModuleHash = normalizePerformanceHash(e.MsgbusModuleHash)
	e.MessageCodecHash = normalizePerformanceHash(e.MessageCodecHash)
	e.MessageIDDerivationHash = normalizePerformanceHash(e.MessageIDDerivationHash)
	e.InboxStoreRoot = normalizePerformanceHash(e.InboxStoreRoot)
	e.OutboxStoreRoot = normalizePerformanceHash(e.OutboxStoreRoot)
	e.ReceiptStoreRoot = normalizePerformanceHash(e.ReceiptStoreRoot)
	e.LocalExecutionRoot = normalizePerformanceHash(e.LocalExecutionRoot)
	e.ExpiryBounceRoot = normalizePerformanceHash(e.ExpiryBounceRoot)
	e.InclusionProofRoot = normalizePerformanceHash(e.InclusionProofRoot)
	e.EvidenceHash = normalizePerformanceHash(e.EvidenceHash)
	return e
}

func (e MessageBusMigrationEvidence) ValidateFormat() error {
	e = e.Normalize()
	hashes := []struct {
		name	string
		value	string
	}{
		{"aetracore migration msgbus module hash", e.MsgbusModuleHash},
		{"aetracore migration message codec hash", e.MessageCodecHash},
		{"aetracore migration message ID derivation hash", e.MessageIDDerivationHash},
		{"aetracore migration inbox store root", e.InboxStoreRoot},
		{"aetracore migration outbox store root", e.OutboxStoreRoot},
		{"aetracore migration receipt store root", e.ReceiptStoreRoot},
		{"aetracore migration local execution root", e.LocalExecutionRoot},
		{"aetracore migration expiry bounce root", e.ExpiryBounceRoot},
		{"aetracore migration inclusion proof root", e.InclusionProofRoot},
	}
	for _, item := range hashes {
		if err := ValidateHash(item.name, item.value); err != nil {
			return err
		}
	}
	if !e.MessagesCommitted {
		return errors.New("aetracore migration message bus evidence requires first-class committed messages")
	}
	if !e.LocalAsyncDeterministic {
		return errors.New("aetracore migration message bus evidence requires deterministic local async execution")
	}
	if !e.ReceiptProofQueryable {
		return errors.New("aetracore migration message bus evidence requires proof-queryable receipts")
	}
	if e.EvidenceHash != "" {
		if err := ValidateHash("aetracore migration message bus evidence hash", e.EvidenceHash); err != nil {
			return err
		}
	}
	return nil
}

func (e MessageBusMigrationEvidence) Validate() error {
	e = e.Normalize()
	if err := e.ValidateFormat(); err != nil {
		return err
	}
	if e.EvidenceHash == "" {
		return errors.New("aetracore migration message bus evidence hash is required")
	}
	expected := ComputeMessageBusMigrationEvidenceHash(e)
	if e.EvidenceHash != expected {
		return fmt.Errorf("aetracore migration message bus evidence hash mismatch: expected %s", expected)
	}
	return nil
}

func (e ZoneExtractionMigrationEvidence) Normalize() ZoneExtractionMigrationEvidence {
	e.FinancialZoneRoot = normalizePerformanceHash(e.FinancialZoneRoot)
	e.IdentityZoneRoot = normalizePerformanceHash(e.IdentityZoneRoot)
	e.ApplicationZoneRoot = normalizePerformanceHash(e.ApplicationZoneRoot)
	e.ZoneKeeperRoot = normalizePerformanceHash(e.ZoneKeeperRoot)
	e.ZonePrefixRoot = normalizePerformanceHash(e.ZonePrefixRoot)
	e.ZoneFeePolicyRoot = normalizePerformanceHash(e.ZoneFeePolicyRoot)
	e.ZoneExecutionSummaryRoot = normalizePerformanceHash(e.ZoneExecutionSummaryRoot)
	e.EvidenceHash = normalizePerformanceHash(e.EvidenceHash)
	return e
}

func (e ZoneExtractionMigrationEvidence) ValidateFormat() error {
	e = e.Normalize()
	hashes := []struct {
		name	string
		value	string
	}{
		{"aetracore migration financial zone root", e.FinancialZoneRoot},
		{"aetracore migration identity zone root", e.IdentityZoneRoot},
		{"aetracore migration application zone root", e.ApplicationZoneRoot},
		{"aetracore migration zone keeper root", e.ZoneKeeperRoot},
		{"aetracore migration zone prefix root", e.ZonePrefixRoot},
		{"aetracore migration zone fee policy root", e.ZoneFeePolicyRoot},
		{"aetracore migration zone execution summary root", e.ZoneExecutionSummaryRoot},
	}
	for _, item := range hashes {
		if err := ValidateHash(item.name, item.value); err != nil {
			return err
		}
	}
	if !e.FinancialModulesRouted {
		return errors.New("aetracore migration zone extraction evidence requires financial modules routed to Financial Zone")
	}
	if !e.IdentityIsolated {
		return errors.New("aetracore migration zone extraction evidence requires isolated Identity Zone")
	}
	if !e.ZoneRootsCommittedPerBlock {
		return errors.New("aetracore migration zone extraction evidence requires zone roots committed per block")
	}
	if e.EvidenceHash != "" {
		if err := ValidateHash("aetracore migration zone extraction evidence hash", e.EvidenceHash); err != nil {
			return err
		}
	}
	return nil
}

func (e ZoneExtractionMigrationEvidence) Validate() error {
	e = e.Normalize()
	if err := e.ValidateFormat(); err != nil {
		return err
	}
	if e.EvidenceHash == "" {
		return errors.New("aetracore migration zone extraction evidence hash is required")
	}
	expected := ComputeZoneExtractionMigrationEvidenceHash(e)
	if e.EvidenceHash != expected {
		return fmt.Errorf("aetracore migration zone extraction evidence hash mismatch: expected %s", expected)
	}
	return nil
}

func (e ShardingRuntimeMigrationEvidence) Normalize() ShardingRuntimeMigrationEvidence {
	e.ShardsModuleHash = normalizePerformanceHash(e.ShardsModuleHash)
	e.ShardLayoutRoot = normalizePerformanceHash(e.ShardLayoutRoot)
	e.RouteKeyCalculationRoot = normalizePerformanceHash(e.RouteKeyCalculationRoot)
	e.PerShardInboxRoot = normalizePerformanceHash(e.PerShardInboxRoot)
	e.PerShardOutboxRoot = normalizePerformanceHash(e.PerShardOutboxRoot)
	e.ShardRootAggregate = normalizePerformanceHash(e.ShardRootAggregate)
	e.SplitMergeScheduleRoot = normalizePerformanceHash(e.SplitMergeScheduleRoot)
	e.ShardMigrationRoot = normalizePerformanceHash(e.ShardMigrationRoot)
	e.EvidenceHash = normalizePerformanceHash(e.EvidenceHash)
	return e
}

func (e ShardingRuntimeMigrationEvidence) ValidateFormat() error {
	e = e.Normalize()
	hashes := []struct {
		name	string
		value	string
	}{
		{"aetracore migration shards module hash", e.ShardsModuleHash},
		{"aetracore migration shard layout root", e.ShardLayoutRoot},
		{"aetracore migration route key calculation root", e.RouteKeyCalculationRoot},
		{"aetracore migration per-shard inbox root", e.PerShardInboxRoot},
		{"aetracore migration per-shard outbox root", e.PerShardOutboxRoot},
		{"aetracore migration shard root aggregate", e.ShardRootAggregate},
		{"aetracore migration split merge schedule root", e.SplitMergeScheduleRoot},
		{"aetracore migration shard migration root", e.ShardMigrationRoot},
	}
	for _, item := range hashes {
		if err := ValidateHash(item.name, item.value); err != nil {
			return err
		}
	}
	if !e.MultiShardZones {
		return errors.New("aetracore migration sharding evidence requires multi-shard zones")
	}
	if !e.ParallelShardWorkloads {
		return errors.New("aetracore migration sharding evidence requires parallel independent shard workloads")
	}
	if !e.InflightMessagesSafe {
		return errors.New("aetracore migration sharding evidence requires in-flight messages to survive layout changes")
	}
	if e.EvidenceHash != "" {
		if err := ValidateHash("aetracore migration sharding evidence hash", e.EvidenceHash); err != nil {
			return err
		}
	}
	return nil
}

func (e ShardingRuntimeMigrationEvidence) Validate() error {
	e = e.Normalize()
	if err := e.ValidateFormat(); err != nil {
		return err
	}
	if e.EvidenceHash == "" {
		return errors.New("aetracore migration sharding evidence hash is required")
	}
	expected := ComputeShardingRuntimeMigrationEvidenceHash(e)
	if e.EvidenceHash != expected {
		return fmt.Errorf("aetracore migration sharding evidence hash mismatch: expected %s", expected)
	}
	return nil
}

func (e AVM20MigrationEvidence) Normalize() AVM20MigrationEvidence {
	e.BytecodeFormatHash = normalizePerformanceHash(e.BytecodeFormatHash)
	e.InterpreterRoot = normalizePerformanceHash(e.InterpreterRoot)
	e.GasTableRoot = normalizePerformanceHash(e.GasTableRoot)
	e.ContractStorageRoot = normalizePerformanceHash(e.ContractStorageRoot)
	e.MessageSyscallRoot = normalizePerformanceHash(e.MessageSyscallRoot)
	e.ProofSyscallRoot = normalizePerformanceHash(e.ProofSyscallRoot)
	e.ABIRegistryRoot = normalizePerformanceHash(e.ABIRegistryRoot)
	e.EvidenceHash = normalizePerformanceHash(e.EvidenceHash)
	return e
}

func (e AVM20MigrationEvidence) ValidateFormat() error {
	e = e.Normalize()
	hashes := []struct {
		name	string
		value	string
	}{
		{"aetracore migration AVM bytecode format hash", e.BytecodeFormatHash},
		{"aetracore migration AVM interpreter root", e.InterpreterRoot},
		{"aetracore migration AVM gas table root", e.GasTableRoot},
		{"aetracore migration contract storage root", e.ContractStorageRoot},
		{"aetracore migration contract message syscall root", e.MessageSyscallRoot},
		{"aetracore migration proof syscall root", e.ProofSyscallRoot},
		{"aetracore migration ABI registry root", e.ABIRegistryRoot},
	}
	for _, item := range hashes {
		if err := ValidateHash(item.name, item.value); err != nil {
			return err
		}
	}
	if !e.DeterministicContracts {
		return errors.New("aetracore migration AVM evidence requires deterministic contract execution")
	}
	if !e.AsyncMessages {
		return errors.New("aetracore migration AVM evidence requires async contract messages")
	}
	if !e.ContractProofsAvailable {
		return errors.New("aetracore migration AVM evidence requires contract state proofs")
	}
	if e.EvidenceHash != "" {
		if err := ValidateHash("aetracore migration AVM evidence hash", e.EvidenceHash); err != nil {
			return err
		}
	}
	return nil
}

func (e AVM20MigrationEvidence) Validate() error {
	e = e.Normalize()
	if err := e.ValidateFormat(); err != nil {
		return err
	}
	if e.EvidenceHash == "" {
		return errors.New("aetracore migration AVM evidence hash is required")
	}
	expected := ComputeAVM20MigrationEvidenceHash(e)
	if e.EvidenceHash != expected {
		return fmt.Errorf("aetracore migration AVM evidence hash mismatch: expected %s", expected)
	}
	return nil
}

func (e IdentityPaymentIntegrationEvidence) Normalize() IdentityPaymentIntegrationEvidence {
	e.IdentityProofRoot = normalizePerformanceHash(e.IdentityProofRoot)
	e.IdentityLookupMessageRoot = normalizePerformanceHash(e.IdentityLookupMessageRoot)
	e.PaymentChannelSettlementRoot = normalizePerformanceHash(e.PaymentChannelSettlementRoot)
	e.ConditionalPaymentRouteRoot = normalizePerformanceHash(e.ConditionalPaymentRouteRoot)
	e.PaymentProofAPIRoot = normalizePerformanceHash(e.PaymentProofAPIRoot)
	e.WalletSDKHelperRoot = normalizePerformanceHash(e.WalletSDKHelperRoot)
	e.EvidenceHash = normalizePerformanceHash(e.EvidenceHash)
	return e
}

func (e IdentityPaymentIntegrationEvidence) ValidateFormat() error {
	e = e.Normalize()
	hashes := []struct {
		name	string
		value	string
	}{
		{"aetracore migration identity proof root", e.IdentityProofRoot},
		{"aetracore migration identity lookup message root", e.IdentityLookupMessageRoot},
		{"aetracore migration payment channel settlement root", e.PaymentChannelSettlementRoot},
		{"aetracore migration conditional payment route root", e.ConditionalPaymentRouteRoot},
		{"aetracore migration payment proof API root", e.PaymentProofAPIRoot},
		{"aetracore migration wallet SDK helper root", e.WalletSDKHelperRoot},
	}
	for _, item := range hashes {
		if err := ValidateHash(item.name, item.value); err != nil {
			return err
		}
	}
	if !e.ProofBackedNames {
		return errors.New("aetracore migration identity/payment evidence requires proof-backed name resolution")
	}
	if !e.TrustlessPayments {
		return errors.New("aetracore migration identity/payment evidence requires trustless Financial Zone settlement")
	}
	if !e.AsyncContractMessages {
		return errors.New("aetracore migration identity/payment evidence requires async contract identity and payment messages")
	}
	if e.EvidenceHash != "" {
		if err := ValidateHash("aetracore migration identity/payment evidence hash", e.EvidenceHash); err != nil {
			return err
		}
	}
	return nil
}

func (e IdentityPaymentIntegrationEvidence) Validate() error {
	e = e.Normalize()
	if err := e.ValidateFormat(); err != nil {
		return err
	}
	if e.EvidenceHash == "" {
		return errors.New("aetracore migration identity/payment evidence hash is required")
	}
	expected := ComputeIdentityPaymentIntegrationEvidenceHash(e)
	if e.EvidenceHash != expected {
		return fmt.Errorf("aetracore migration identity/payment evidence hash mismatch: expected %s", expected)
	}
	return nil
}

func (e PerformanceHardeningMigrationEvidence) Normalize() PerformanceHardeningMigrationEvidence {
	e.BlockSTMWorkloadRoot = normalizePerformanceHash(e.BlockSTMWorkloadRoot)
	e.ConflictProfileRoot = normalizePerformanceHash(e.ConflictProfileRoot)
	e.StoreV2BenchmarkRoot = normalizePerformanceHash(e.StoreV2BenchmarkRoot)
	e.MempoolLaneRoot = normalizePerformanceHash(e.MempoolLaneRoot)
	e.CongestionRoutingRoot = normalizePerformanceHash(e.CongestionRoutingRoot)
	e.AdaptiveSyncRecoveryRoot = normalizePerformanceHash(e.AdaptiveSyncRecoveryRoot)
	e.LoadSimulationRoot = normalizePerformanceHash(e.LoadSimulationRoot)
	e.EvidenceHash = normalizePerformanceHash(e.EvidenceHash)
	return e
}

func (e PerformanceHardeningMigrationEvidence) ValidateFormat() error {
	e = e.Normalize()
	hashes := []struct {
		name	string
		value	string
	}{
		{"aetracore migration BlockSTM workload root", e.BlockSTMWorkloadRoot},
		{"aetracore migration conflict profile root", e.ConflictProfileRoot},
		{"aetracore migration Store v2 benchmark root", e.StoreV2BenchmarkRoot},
		{"aetracore migration mempool lane root", e.MempoolLaneRoot},
		{"aetracore migration congestion routing root", e.CongestionRoutingRoot},
		{"aetracore migration AdaptiveSync recovery root", e.AdaptiveSyncRecoveryRoot},
		{"aetracore migration load simulation root", e.LoadSimulationRoot},
	}
	for _, item := range hashes {
		if err := ValidateHash(item.name, item.value); err != nil {
			return err
		}
	}
	if !e.ParallelismScales {
		return errors.New("aetracore migration performance evidence requires parallelism scaling")
	}
	if !e.StateSyncRecoversRoots {
		return errors.New("aetracore migration performance evidence requires state sync commitment recovery")
	}
	if !e.CongestionDeterministic {
		return errors.New("aetracore migration performance evidence requires deterministic congestion routing")
	}
	if e.EvidenceHash != "" {
		if err := ValidateHash("aetracore migration performance evidence hash", e.EvidenceHash); err != nil {
			return err
		}
	}
	return nil
}

func (e PerformanceHardeningMigrationEvidence) Validate() error {
	e = e.Normalize()
	if err := e.ValidateFormat(); err != nil {
		return err
	}
	if e.EvidenceHash == "" {
		return errors.New("aetracore migration performance evidence hash is required")
	}
	expected := ComputePerformanceHardeningMigrationEvidenceHash(e)
	if e.EvidenceHash != expected {
		return fmt.Errorf("aetracore migration performance evidence hash mismatch: expected %s", expected)
	}
	return nil
}

func ValidateMigrationPathCoverage() error {
	spec, err := DefaultMigrationPathSpec()
	if err != nil {
		return err
	}
	requiredTasks := map[MigrationPhaseID][]MigrationTaskID{
		MigrationPhaseBaselineHardening: {
			MigrationTaskModuleBoundaryDocs,
			MigrationTaskStateExportValidation,
			MigrationTaskDeterministicGenesis,
			MigrationTaskDynamicFeeBoundsTests,
			MigrationTaskLegacyModuleInvariants,
			MigrationTaskStoreV2CompatibilityAudit,
		},
		MigrationPhaseCoreCommitments: {
			MigrationTaskAetraCoreModule,
			MigrationTaskDefaultZoneRegistry,
			MigrationTaskDefaultZoneStateRoot,
			MigrationTaskEmptyMessageRoot,
			MigrationTaskProofRootRegistry,
			MigrationTaskRootQueryAPIs,
		},
		MigrationPhaseMessageBus: {
			MigrationTaskMsgbusModule,
			MigrationTaskMessageEncodingAndIDs,
			MigrationTaskMessageStores,
			MigrationTaskLocalZoneExecution,
			MigrationTaskExpiryBounceLogic,
			MigrationTaskMessageInclusionProofs,
		},
		MigrationPhaseZoneExtraction: {
			MigrationTaskExtractFinancialZone,
			MigrationTaskExtractIdentityZone,
			MigrationTaskExtractApplicationZone,
			MigrationTaskZoneSpecificKeepers,
			MigrationTaskZoneLocalFeePolicies,
			MigrationTaskZoneExecutionSummaries,
		},
		MigrationPhaseShardingRuntime: {
			MigrationTaskShardsModule,
			MigrationTaskShardLayoutDescriptors,
			MigrationTaskRouteKeyCalculation,
			MigrationTaskPerShardMessageStores,
			MigrationTaskShardRootAggregation,
			MigrationTaskSplitMergeScheduler,
			MigrationTaskDeterministicShardMigration,
		},
		MigrationPhaseAVM20: {
			MigrationTaskAVMBytecodeFormat,
			MigrationTaskAVMInterpreter,
			MigrationTaskAVMGasTable,
			MigrationTaskContractStorageAdapter,
			MigrationTaskContractMessageSyscalls,
			MigrationTaskProofVerificationSyscalls,
			MigrationTaskABIRegistry,
		},
		MigrationPhaseIdentityPayments: {
			MigrationTaskIdentityProofActivation,
			MigrationTaskCrossZoneIdentityLookup,
			MigrationTaskPaymentChannelSettlement,
			MigrationTaskConditionalPaymentRouting,
			MigrationTaskPaymentProofAPIs,
			MigrationTaskWalletSDKIdentityPayment,
		},
		MigrationPhasePerformanceHardening: {
			MigrationTaskBlockSTMZoneShardWorkloads,
			MigrationTaskConflictProfiling,
			MigrationTaskStoreV2Benchmarks,
			MigrationTaskMempoolLanes,
			MigrationTaskCongestionAwareRouting,
			MigrationTaskAdaptiveSyncRecoveryTests,
			MigrationTaskMultiZoneLoadSimulation,
		},
	}
	requiredCriteria := map[MigrationPhaseID][]MigrationExitCriterionID{
		MigrationPhaseBaselineHardening: {
			MigrationExitSingleChainReproducibleExport,
			MigrationExitLegacyInvariantCoverage,
			MigrationExitSafePrefixMigration,
		},
		MigrationPhaseCoreCommitments: {
			MigrationExitSingleZoneOperation,
			MigrationExitAppHashCoreRoot,
			MigrationExitProofRegistryRootMetadata,
		},
		MigrationPhaseMessageBus: {
			MigrationExitMessagesFirstClassCommitted,
			MigrationExitLocalAsyncDeterministic,
			MigrationExitMessageReceiptsProofQueryable,
		},
		MigrationPhaseZoneExtraction: {
			MigrationExitFinancialZoneExecution,
			MigrationExitIdentityZoneIsolated,
			MigrationExitZoneRootsPerBlock,
		},
		MigrationPhaseShardingRuntime: {
			MigrationExitMultiShardZones,
			MigrationExitParallelShardWorkloads,
			MigrationExitInflightMessagesSurviveLayoutChanges,
		},
		MigrationPhaseAVM20: {
			MigrationExitContractZoneDeterministic,
			MigrationExitContractsEmitAsyncMessages,
			MigrationExitContractStateProofsAvailable,
		},
		MigrationPhaseIdentityPayments: {
			MigrationExitNamesResolveProofBacked,
			MigrationExitPaymentsSettleTrustlessly,
			MigrationExitContractsUseIdentityPaymentsAsync,
		},
		MigrationPhasePerformanceHardening: {
			MigrationExitParallelismScales,
			MigrationExitStateSyncRecoversCommitments,
			MigrationExitRoutingDeterministicCongestion,
		},
	}
	phaseByID := make(map[MigrationPhaseID]MigrationPhase, len(spec.Phases))
	for _, phase := range spec.Phases {
		phaseByID[phase.PhaseID] = phase
	}
	for phaseID, tasks := range requiredTasks {
		phase, found := phaseByID[phaseID]
		if !found {
			return fmt.Errorf("aetracore migration coverage missing phase %s", phaseID)
		}
		seen := make(map[MigrationTaskID]struct{}, len(phase.Tasks))
		for _, task := range phase.Tasks {
			seen[task.TaskID] = struct{}{}
		}
		for _, taskID := range tasks {
			if _, found := seen[taskID]; !found {
				return fmt.Errorf("aetracore migration coverage missing task %s", taskID)
			}
		}
	}
	for phaseID, criteria := range requiredCriteria {
		phase, found := phaseByID[phaseID]
		if !found {
			return fmt.Errorf("aetracore migration coverage missing phase %s", phaseID)
		}
		seen := make(map[MigrationExitCriterionID]struct{}, len(phase.ExitCriteria))
		for _, criterion := range phase.ExitCriteria {
			seen[criterion.CriterionID] = struct{}{}
		}
		for _, criterionID := range criteria {
			if _, found := seen[criterionID]; !found {
				return fmt.Errorf("aetracore migration coverage missing exit criterion %s", criterionID)
			}
		}
	}
	return nil
}

func IsMigrationPhaseID(id MigrationPhaseID) bool {
	return id == MigrationPhaseBaselineHardening || id == MigrationPhaseCoreCommitments || id == MigrationPhaseMessageBus || id == MigrationPhaseZoneExtraction || id == MigrationPhaseShardingRuntime || id == MigrationPhaseAVM20 || id == MigrationPhaseIdentityPayments || id == MigrationPhasePerformanceHardening
}

func IsMigrationTaskID(phaseID MigrationPhaseID, taskID MigrationTaskID) bool {
	for _, task := range phaseTasksForID(phaseID) {
		if task == taskID {
			return true
		}
	}
	return false
}

func IsMigrationExitCriterionID(phaseID MigrationPhaseID, criterionID MigrationExitCriterionID) bool {
	for _, criterion := range phaseExitCriteriaForID(phaseID) {
		if criterion == criterionID {
			return true
		}
	}
	return false
}

func ComputeMigrationPathSpecRoot(phases []MigrationPhase) string {
	phases = normalizeMigrationPhases(phases)
	return hashRoot("aetra-aek-migration-path-spec-v1", func(w byteWriter) {
		writeUint64(w, uint64(len(phases)))
		for _, phase := range phases {
			writePart(w, string(phase.PhaseID))
			writePart(w, phase.PhaseHash)
		}
	})
}

func ComputeMigrationPhaseHash(phase MigrationPhase) string {
	phase = phase.Normalize()
	return hashRoot("aetra-aek-migration-phase-v1", func(w byteWriter) {
		writePart(w, string(phase.PhaseID))
		writePart(w, phase.Title)
		writeUint64(w, uint64(len(phase.Tasks)))
		for _, task := range phase.Tasks {
			writePart(w, task.DescriptorHash)
		}
		writeUint64(w, uint64(len(phase.ExitCriteria)))
		for _, criterion := range phase.ExitCriteria {
			writePart(w, criterion.DescriptorHash)
		}
	})
}

func ComputeMigrationTaskHash(task MigrationTaskDescriptor) string {
	task = task.Normalize()
	return hashRoot("aetra-aek-migration-task-v1", func(w byteWriter) {
		writePart(w, string(task.PhaseID))
		writePart(w, string(task.TaskID))
		writePart(w, task.Task)
		writePart(w, task.Target)
		writePart(w, task.Evidence)
	})
}

func ComputeMigrationExitCriterionHash(criterion MigrationExitCriterion) string {
	criterion = criterion.Normalize()
	return hashRoot("aetra-aek-migration-exit-criterion-v1", func(w byteWriter) {
		writePart(w, string(criterion.PhaseID))
		writePart(w, string(criterion.CriterionID))
		writePart(w, criterion.Criterion)
		writePart(w, criterion.Evidence)
	})
}

func ComputeBaselineHardeningEvidenceHash(e BaselineHardeningEvidence) string {
	e = e.Normalize()
	return hashRoot("aetra-aek-migration-baseline-evidence-v1", func(w byteWriter) {
		writePart(w, e.ModuleBoundaryDocsRoot)
		writePart(w, e.StateExportManifestHash)
		writePart(w, e.GenesisImportHash)
		writePart(w, e.DynamicFeeBoundsTestHash)
		writePart(w, e.LegacyInvariantRoot)
		writePart(w, e.StoreV2AuditHash)
		writePart(w, e.UpgradeHandlerPrefixHash)
		writeBoolPart(w, e.StateReproducible)
		writeBoolPart(w, e.StateExportable)
		writeBoolPart(w, e.InvariantCoverage)
		writeBoolPart(w, e.PrefixMigrationSafe)
	})
}

func ComputeCoreCommitmentMigrationEvidenceHash(e CoreCommitmentMigrationEvidence) string {
	e = e.Normalize()
	return hashRoot("aetra-aek-migration-core-commitment-evidence-v1", func(w byteWriter) {
		writePart(w, e.AetraCoreModuleHash)
		writePart(w, e.DefaultZoneDescriptorHash)
		writePart(w, e.DefaultZoneStateRoot)
		writePart(w, e.EmptyMessageRoot)
		writePart(w, e.ProofRegistryRoot)
		writePart(w, e.RootQueryAPIHash)
		writePart(w, e.AppHashCoreRoot)
		writePart(w, string(e.DefaultZoneID))
		writeBoolPart(w, e.SingleZoneMode)
		writeBoolPart(w, e.ProofRegistryMetadata)
	})
}

func ComputeMessageBusMigrationEvidenceHash(e MessageBusMigrationEvidence) string {
	e = e.Normalize()
	return hashRoot("aetra-aek-migration-message-bus-evidence-v1", func(w byteWriter) {
		writePart(w, e.MsgbusModuleHash)
		writePart(w, e.MessageCodecHash)
		writePart(w, e.MessageIDDerivationHash)
		writePart(w, e.InboxStoreRoot)
		writePart(w, e.OutboxStoreRoot)
		writePart(w, e.ReceiptStoreRoot)
		writePart(w, e.LocalExecutionRoot)
		writePart(w, e.ExpiryBounceRoot)
		writePart(w, e.InclusionProofRoot)
		writeBoolPart(w, e.MessagesCommitted)
		writeBoolPart(w, e.LocalAsyncDeterministic)
		writeBoolPart(w, e.ReceiptProofQueryable)
	})
}

func ComputeZoneExtractionMigrationEvidenceHash(e ZoneExtractionMigrationEvidence) string {
	e = e.Normalize()
	return hashRoot("aetra-aek-migration-zone-extraction-evidence-v1", func(w byteWriter) {
		writePart(w, e.FinancialZoneRoot)
		writePart(w, e.IdentityZoneRoot)
		writePart(w, e.ApplicationZoneRoot)
		writePart(w, e.ZoneKeeperRoot)
		writePart(w, e.ZonePrefixRoot)
		writePart(w, e.ZoneFeePolicyRoot)
		writePart(w, e.ZoneExecutionSummaryRoot)
		writeBoolPart(w, e.FinancialModulesRouted)
		writeBoolPart(w, e.IdentityIsolated)
		writeBoolPart(w, e.ZoneRootsCommittedPerBlock)
	})
}

func ComputeShardingRuntimeMigrationEvidenceHash(e ShardingRuntimeMigrationEvidence) string {
	e = e.Normalize()
	return hashRoot("aetra-aek-migration-sharding-runtime-evidence-v1", func(w byteWriter) {
		writePart(w, e.ShardsModuleHash)
		writePart(w, e.ShardLayoutRoot)
		writePart(w, e.RouteKeyCalculationRoot)
		writePart(w, e.PerShardInboxRoot)
		writePart(w, e.PerShardOutboxRoot)
		writePart(w, e.ShardRootAggregate)
		writePart(w, e.SplitMergeScheduleRoot)
		writePart(w, e.ShardMigrationRoot)
		writeBoolPart(w, e.MultiShardZones)
		writeBoolPart(w, e.ParallelShardWorkloads)
		writeBoolPart(w, e.InflightMessagesSafe)
	})
}

func ComputeAVM20MigrationEvidenceHash(e AVM20MigrationEvidence) string {
	e = e.Normalize()
	return hashRoot("aetra-aek-migration-avm-2.0-evidence-v1", func(w byteWriter) {
		writePart(w, e.BytecodeFormatHash)
		writePart(w, e.InterpreterRoot)
		writePart(w, e.GasTableRoot)
		writePart(w, e.ContractStorageRoot)
		writePart(w, e.MessageSyscallRoot)
		writePart(w, e.ProofSyscallRoot)
		writePart(w, e.ABIRegistryRoot)
		writeBoolPart(w, e.DeterministicContracts)
		writeBoolPart(w, e.AsyncMessages)
		writeBoolPart(w, e.ContractProofsAvailable)
	})
}

func ComputeIdentityPaymentIntegrationEvidenceHash(e IdentityPaymentIntegrationEvidence) string {
	e = e.Normalize()
	return hashRoot("aetra-aek-migration-identity-payment-evidence-v1", func(w byteWriter) {
		writePart(w, e.IdentityProofRoot)
		writePart(w, e.IdentityLookupMessageRoot)
		writePart(w, e.PaymentChannelSettlementRoot)
		writePart(w, e.ConditionalPaymentRouteRoot)
		writePart(w, e.PaymentProofAPIRoot)
		writePart(w, e.WalletSDKHelperRoot)
		writeBoolPart(w, e.ProofBackedNames)
		writeBoolPart(w, e.TrustlessPayments)
		writeBoolPart(w, e.AsyncContractMessages)
	})
}

func ComputePerformanceHardeningMigrationEvidenceHash(e PerformanceHardeningMigrationEvidence) string {
	e = e.Normalize()
	return hashRoot("aetra-aek-migration-performance-hardening-evidence-v1", func(w byteWriter) {
		writePart(w, e.BlockSTMWorkloadRoot)
		writePart(w, e.ConflictProfileRoot)
		writePart(w, e.StoreV2BenchmarkRoot)
		writePart(w, e.MempoolLaneRoot)
		writePart(w, e.CongestionRoutingRoot)
		writePart(w, e.AdaptiveSyncRecoveryRoot)
		writePart(w, e.LoadSimulationRoot)
		writeBoolPart(w, e.ParallelismScales)
		writeBoolPart(w, e.StateSyncRecoversRoots)
		writeBoolPart(w, e.CongestionDeterministic)
	})
}

func migrationPhase(phaseID MigrationPhaseID, title string, tasks []MigrationTaskDescriptor, criteria []MigrationExitCriterion) MigrationPhase {
	phase := MigrationPhase{
		PhaseID:	phaseID,
		Title:		title,
		Tasks:		normalizeMigrationTasks(tasks),
		ExitCriteria:	normalizeMigrationExitCriteria(criteria),
	}
	phase.PhaseHash = ComputeMigrationPhaseHash(phase)
	return phase
}

func migrationTask(phaseID MigrationPhaseID, taskID MigrationTaskID, task, target, evidence string) MigrationTaskDescriptor {
	descriptor := MigrationTaskDescriptor{
		PhaseID:	phaseID,
		TaskID:		taskID,
		Task:		task,
		Target:		target,
		Evidence:	evidence,
	}
	descriptor.DescriptorHash = ComputeMigrationTaskHash(descriptor)
	return descriptor
}

func migrationExitCriterion(phaseID MigrationPhaseID, criterionID MigrationExitCriterionID, criterion, evidence string) MigrationExitCriterion {
	descriptor := MigrationExitCriterion{
		PhaseID:	phaseID,
		CriterionID:	criterionID,
		Criterion:	criterion,
		Evidence:	evidence,
	}
	descriptor.DescriptorHash = ComputeMigrationExitCriterionHash(descriptor)
	return descriptor
}

func normalizeMigrationPhases(phases []MigrationPhase) []MigrationPhase {
	normalized := make([]MigrationPhase, len(phases))
	for i, phase := range phases {
		normalized[i] = phase.Normalize()
	}
	sort.Slice(normalized, func(i, j int) bool {
		return normalized[i].PhaseID < normalized[j].PhaseID
	})
	return normalized
}

func normalizeMigrationTasks(tasks []MigrationTaskDescriptor) []MigrationTaskDescriptor {
	normalized := make([]MigrationTaskDescriptor, len(tasks))
	for i, task := range tasks {
		normalized[i] = task.Normalize()
	}
	sort.Slice(normalized, func(i, j int) bool {
		return normalized[i].TaskID < normalized[j].TaskID
	})
	return normalized
}

func normalizeMigrationExitCriteria(criteria []MigrationExitCriterion) []MigrationExitCriterion {
	normalized := make([]MigrationExitCriterion, len(criteria))
	for i, criterion := range criteria {
		normalized[i] = criterion.Normalize()
	}
	sort.Slice(normalized, func(i, j int) bool {
		return normalized[i].CriterionID < normalized[j].CriterionID
	})
	return normalized
}

func phaseTasksForID(phaseID MigrationPhaseID) []MigrationTaskID {
	switch phaseID {
	case MigrationPhaseBaselineHardening:
		return []MigrationTaskID{
			MigrationTaskModuleBoundaryDocs,
			MigrationTaskStateExportValidation,
			MigrationTaskDeterministicGenesis,
			MigrationTaskDynamicFeeBoundsTests,
			MigrationTaskLegacyModuleInvariants,
			MigrationTaskStoreV2CompatibilityAudit,
		}
	case MigrationPhaseCoreCommitments:
		return []MigrationTaskID{
			MigrationTaskAetraCoreModule,
			MigrationTaskDefaultZoneRegistry,
			MigrationTaskDefaultZoneStateRoot,
			MigrationTaskEmptyMessageRoot,
			MigrationTaskProofRootRegistry,
			MigrationTaskRootQueryAPIs,
		}
	case MigrationPhaseMessageBus:
		return []MigrationTaskID{
			MigrationTaskMsgbusModule,
			MigrationTaskMessageEncodingAndIDs,
			MigrationTaskMessageStores,
			MigrationTaskLocalZoneExecution,
			MigrationTaskExpiryBounceLogic,
			MigrationTaskMessageInclusionProofs,
		}
	case MigrationPhaseZoneExtraction:
		return []MigrationTaskID{
			MigrationTaskExtractFinancialZone,
			MigrationTaskExtractIdentityZone,
			MigrationTaskExtractApplicationZone,
			MigrationTaskZoneSpecificKeepers,
			MigrationTaskZoneLocalFeePolicies,
			MigrationTaskZoneExecutionSummaries,
		}
	case MigrationPhaseShardingRuntime:
		return []MigrationTaskID{
			MigrationTaskShardsModule,
			MigrationTaskShardLayoutDescriptors,
			MigrationTaskRouteKeyCalculation,
			MigrationTaskPerShardMessageStores,
			MigrationTaskShardRootAggregation,
			MigrationTaskSplitMergeScheduler,
			MigrationTaskDeterministicShardMigration,
		}
	case MigrationPhaseAVM20:
		return []MigrationTaskID{
			MigrationTaskAVMBytecodeFormat,
			MigrationTaskAVMInterpreter,
			MigrationTaskAVMGasTable,
			MigrationTaskContractStorageAdapter,
			MigrationTaskContractMessageSyscalls,
			MigrationTaskProofVerificationSyscalls,
			MigrationTaskABIRegistry,
		}
	case MigrationPhaseIdentityPayments:
		return []MigrationTaskID{
			MigrationTaskIdentityProofActivation,
			MigrationTaskCrossZoneIdentityLookup,
			MigrationTaskPaymentChannelSettlement,
			MigrationTaskConditionalPaymentRouting,
			MigrationTaskPaymentProofAPIs,
			MigrationTaskWalletSDKIdentityPayment,
		}
	case MigrationPhasePerformanceHardening:
		return []MigrationTaskID{
			MigrationTaskBlockSTMZoneShardWorkloads,
			MigrationTaskConflictProfiling,
			MigrationTaskStoreV2Benchmarks,
			MigrationTaskMempoolLanes,
			MigrationTaskCongestionAwareRouting,
			MigrationTaskAdaptiveSyncRecoveryTests,
			MigrationTaskMultiZoneLoadSimulation,
		}
	default:
		return nil
	}
}

func phaseExitCriteriaForID(phaseID MigrationPhaseID) []MigrationExitCriterionID {
	switch phaseID {
	case MigrationPhaseBaselineHardening:
		return []MigrationExitCriterionID{
			MigrationExitSingleChainReproducibleExport,
			MigrationExitLegacyInvariantCoverage,
			MigrationExitSafePrefixMigration,
		}
	case MigrationPhaseCoreCommitments:
		return []MigrationExitCriterionID{
			MigrationExitSingleZoneOperation,
			MigrationExitAppHashCoreRoot,
			MigrationExitProofRegistryRootMetadata,
		}
	case MigrationPhaseMessageBus:
		return []MigrationExitCriterionID{
			MigrationExitMessagesFirstClassCommitted,
			MigrationExitLocalAsyncDeterministic,
			MigrationExitMessageReceiptsProofQueryable,
		}
	case MigrationPhaseZoneExtraction:
		return []MigrationExitCriterionID{
			MigrationExitFinancialZoneExecution,
			MigrationExitIdentityZoneIsolated,
			MigrationExitZoneRootsPerBlock,
		}
	case MigrationPhaseShardingRuntime:
		return []MigrationExitCriterionID{
			MigrationExitMultiShardZones,
			MigrationExitParallelShardWorkloads,
			MigrationExitInflightMessagesSurviveLayoutChanges,
		}
	case MigrationPhaseAVM20:
		return []MigrationExitCriterionID{
			MigrationExitContractZoneDeterministic,
			MigrationExitContractsEmitAsyncMessages,
			MigrationExitContractStateProofsAvailable,
		}
	case MigrationPhaseIdentityPayments:
		return []MigrationExitCriterionID{
			MigrationExitNamesResolveProofBacked,
			MigrationExitPaymentsSettleTrustlessly,
			MigrationExitContractsUseIdentityPaymentsAsync,
		}
	case MigrationPhasePerformanceHardening:
		return []MigrationExitCriterionID{
			MigrationExitParallelismScales,
			MigrationExitStateSyncRecoversCommitments,
			MigrationExitRoutingDeterministicCongestion,
		}
	default:
		return nil
	}
}

func writeBoolPart(w byteWriter, value bool) {
	if value {
		writePart(w, "true")
		return
	}
	writePart(w, "false")
}
