package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type MigrationPhase string

const (
	MigrationPhase0BaselineHardening	MigrationPhase	= "phase_0_baseline_hardening"
	MigrationPhase1CoreCommitments		MigrationPhase	= "phase_1_core_commitments"
	MigrationPhase2MessageBus		MigrationPhase	= "phase_2_message_bus"
	MigrationPhase3ZoneExtraction		MigrationPhase	= "phase_3_zone_extraction"
	MigrationPhase4ShardingRuntime		MigrationPhase	= "phase_4_sharding_runtime"
	MigrationPhase5AVM20			MigrationPhase	= "phase_5_avm_2_0"
	MigrationPhase6IdentityPayments		MigrationPhase	= "phase_6_identity_payment_integration"
	MigrationPhase7Performance		MigrationPhase	= "phase_7_performance_hardening"
)

type GenesisImportCheck struct {
	ModuleName	string
	Active		bool
	Deterministic	bool
	ExportHash	string
	ImportHash	string
}

type ModuleInvariantCheck struct {
	ModuleName	string
	InvariantName	string
	Covered		bool
	Deterministic	bool
	EvidenceHash	string
}

type StatePrefixMigrationCheck struct {
	ModuleName	string
	OldPrefix	string
	NewPrefix	string
	MigrationHash	string
	ReversibleProof	string
	Safe		bool
}

type MigrationPhase0Input struct {
	ModuleBoundaryDocHash		string
	StateExportValidationHash	string
	ExportedAppHash			string
	ReplayedAppHash			string
	GenesisImports			[]GenesisImportCheck
	DynamicFeeBoundsTestHash	string
	InvariantChecks			[]ModuleInvariantCheck
	StoreV2CompatibilityHash	string
	PrefixMigrations		[]StatePrefixMigrationCheck
}

type RootQueryAPICheck struct {
	QueryName	string
	RootType	ProofRootType
	Available	bool
	ResponseHash	string
}

type ProofRootMetadataCheck struct {
	RootType	ProofRootType
	Height		uint64
	RootHash	string
	MetadataHash	string
}

type MigrationPhase1Input struct {
	AetraCoreModuleHash	string
	ZoneRegistryRoot	string
	ZoneCount		uint32
	DefaultZoneID		string
	DefaultZoneStateRoot	string
	MessageRoot		string
	EmptyQueueRoot		string
	ProofRegistryRoot	string
	RootQueryAPIs		[]RootQueryAPICheck
	ProofMetadata		[]ProofRootMetadataCheck
	AppHashIncludesCoreRoot	bool
	CoreRootHash		string
}

type MsgBusStoreCheck struct {
	StoreName	string
	RootHash	string
	Committed	bool
}

type MsgBusEncodingCheck struct {
	CodecHash		string
	MessageIDRoot		string
	DeterministicIDs	bool
}

type MsgBusExecutionCheck struct {
	ExecutionRoot	string
	Deterministic	bool
	ExecutedLocally	bool
}

type MsgBusSafetyCheck struct {
	ExpiryRoot		string
	BounceRoot		string
	InclusionProofRoot	string
	ReceiptsProofRoot	string
}

type MigrationPhase2Input struct {
	MsgBusModuleHash	string
	Encoding		MsgBusEncodingCheck
	Stores			[]MsgBusStoreCheck
	LocalExecution		MsgBusExecutionCheck
	Safety			MsgBusSafetyCheck
	FirstClassObjectRoot	string
}

type ZoneExtractionCheck struct {
	ZoneID			string
	Extracted		bool
	KeeperHash		string
	StatePrefixRoot		string
	FeePolicyHash		string
	ExecutionSummaryHash	string
	CommittedRoot		string
	Modules			[]string
}

type MigrationPhase3Input struct {
	FinancialZone				ZoneExtractionCheck
	IdentityZone				ZoneExtractionCheck
	ApplicationZone				ZoneExtractionCheck
	BankFeesContractAssetsAVMAMMInFinancial	bool
	IdentityIsolatedActivation		bool
	ZoneRootsCommittedPerBlock		bool
	ZoneCommitmentRoot			string
}

type ShardRuntimeDescriptorCheck struct {
	ShardID			string
	LayoutHash		string
	RouteKeyRoot		string
	InboxRoot		string
	OutboxRoot		string
	ShardRoot		string
	ParallelGroupHash	string
	Active			bool
}

type ShardSplitMergeSchedulerCheck struct {
	SchedulerRoot		string
	SplitDecisionRoot	string
	MergeDecisionRoot	string
	Deterministic		bool
}

type ShardMigrationCheck struct {
	MigrationRoot		string
	OldLayoutHash		string
	NewLayoutHash		string
	InFlightMessageRoot	string
	SurvivesLayoutChange	bool
	DeterministicMigration	bool
}

type MigrationPhase4Input struct {
	ShardsModuleHash		string
	ZoneID				string
	ShardLayoutDescriptorRoot	string
	RouteKeyCalculationHash		string
	ShardDescriptors		[]ShardRuntimeDescriptorCheck
	RootAggregationHash		string
	SplitMergeScheduler		ShardSplitMergeSchedulerCheck
	Migration			ShardMigrationCheck
	ZonesSupportMultipleShards	bool
	IndependentWorkloadsParallel	bool
	InFlightMessagesSurviveChange	bool
}

type AVM20ComponentCheck struct {
	ComponentName	string
	ComponentHash	string
	Implemented	bool
	Deterministic	bool
}

type AVM20SyscallCheck struct {
	SyscallName	string
	SyscallHash	string
	Metered		bool
	Enabled		bool
}

type MigrationPhase5Input struct {
	BytecodeFormat			AVM20ComponentCheck
	Interpreter			AVM20ComponentCheck
	GasTable			AVM20ComponentCheck
	ContractStorageAdapter		AVM20ComponentCheck
	ABIRegistry			AVM20ComponentCheck
	MessageSyscalls			[]AVM20SyscallCheck
	ProofVerificationSyscalls	[]AVM20SyscallCheck
	ContractZoneDeterministic	bool
	AsyncMessageEmissionRoot	string
	ContractStateProofRoot		string
	ContractZoneExecutionRoot	string
}

type IdentityPaymentFlowCheck struct {
	FlowName	string
	FlowRoot	string
	ProofBacked	bool
	Asynchronous	bool
	Deterministic	bool
	ZoneID		string
	MessageTypeHash	string
}

type WalletSDKHelperCheck struct {
	HelperName	string
	HelperHash	string
	Available	bool
	Deterministic	bool
}

type MigrationPhase6Input struct {
	AETIdentityProofRoot			string
	CrossZoneIdentityLookupRoot		string
	PaymentChannelSettlementRoot		string
	ConditionalPaymentRoutingRoot		string
	PaymentProofAPIRoot			string
	IdentityFlows				[]IdentityPaymentFlowCheck
	PaymentFlows				[]IdentityPaymentFlowCheck
	WalletSDKHelpers			[]WalletSDKHelperCheck
	NamesResolveThroughIdentityZone		bool
	PaymentsSettleThroughFinancialZone	bool
	ContractsUseAsyncIdentityPayments	bool
}

type PerformanceHardeningCheck struct {
	CheckName	string
	EvidenceHash	string
	Enabled		bool
	Deterministic	bool
}

type StoreV2BenchmarkCheck struct {
	BenchmarkName	string
	ResultHash	string
	Covered		bool
	Bounded		bool
}

type MigrationPhase7Input struct {
	BlockSTMWorkloadRoot			string
	ConflictProfilingRoot			string
	MempoolLaneRoot				string
	CongestionRoutingRoot			string
	AdaptiveSyncRecoveryRoot		string
	MultiZoneLoadSimulationRoot		string
	HardeningChecks				[]PerformanceHardeningCheck
	StoreV2Benchmarks			[]StoreV2BenchmarkCheck
	IndependentExecutionScalesParallel	bool
	StateSyncRecoversCommitments		bool
	RoutingDeterministicUnderCongestion	bool
}

type MigrationReadinessReport struct {
	Phase		MigrationPhase
	Passed		bool
	Failed		[]string
	Evidence	[]string
	ReportHash	string
}

func BuildMigrationPhase0Readiness(input MigrationPhase0Input) MigrationReadinessReport {
	input = input.Normalize()
	failed := make([]string, 0)
	evidence := make([]string, 0)
	if err := validateHexHash("migration phase 0 module boundary documentation hash", input.ModuleBoundaryDocHash); err != nil {
		failed = append(failed, "module_boundary_documentation")
	} else {
		evidence = append(evidence, "module_boundary_documentation:"+input.ModuleBoundaryDocHash)
	}
	if err := validateHexHash("migration phase 0 state export validation hash", input.StateExportValidationHash); err != nil {
		failed = append(failed, "state_export_validation")
	} else {
		evidence = append(evidence, "state_export_validation:"+input.StateExportValidationHash)
	}
	if input.ExportedAppHash == "" || input.ReplayedAppHash == "" || input.ExportedAppHash != input.ReplayedAppHash {
		failed = append(failed, "single_chain_state_not_reproducible")
	} else if err := validateHexHash("migration phase 0 exported app hash", input.ExportedAppHash); err != nil {
		failed = append(failed, "single_chain_state_hash_invalid")
	} else {
		evidence = append(evidence, "reproducible_export:"+input.ExportedAppHash)
	}
	if len(input.GenesisImports) == 0 {
		failed = append(failed, "active_module_genesis_imports_missing")
	}
	for _, check := range input.GenesisImports {
		if err := check.Validate(); err != nil {
			failed = append(failed, "genesis_import:"+check.ModuleName)
		} else if check.Active {
			evidence = append(evidence, "genesis_import:"+check.ModuleName+":"+check.ImportHash)
		}
	}
	if err := validateHexHash("migration phase 0 dynamic fee bounds test hash", input.DynamicFeeBoundsTestHash); err != nil {
		failed = append(failed, "dynamic_fee_bounds_tests")
	} else {
		evidence = append(evidence, "dynamic_fee_bounds_tests:"+input.DynamicFeeBoundsTestHash)
	}
	if err := validateRequiredInvariantCoverage(input.InvariantChecks); err != nil {
		failed = append(failed, "module_invariant_coverage")
	} else {
		evidence = append(evidence, "module_invariant_coverage:"+hashInvariantChecks(input.InvariantChecks))
	}
	if err := validateHexHash("migration phase 0 Store v2 compatibility audit hash", input.StoreV2CompatibilityHash); err != nil {
		failed = append(failed, "store_v2_compatibility_audit")
	} else {
		evidence = append(evidence, "store_v2_compatibility_audit:"+input.StoreV2CompatibilityHash)
	}
	if len(input.PrefixMigrations) == 0 {
		failed = append(failed, "upgrade_prefix_migrations_missing")
	}
	for _, migration := range input.PrefixMigrations {
		if err := migration.Validate(); err != nil {
			failed = append(failed, "prefix_migration:"+migration.ModuleName)
		} else {
			evidence = append(evidence, "prefix_migration:"+migration.ModuleName+":"+migration.MigrationHash)
		}
	}
	report := MigrationReadinessReport{
		Phase:		MigrationPhase0BaselineHardening,
		Passed:		len(failed) == 0,
		Failed:		normalizeStringSet(failed),
		Evidence:	normalizeStringSet(evidence),
	}
	report.ReportHash = ComputeMigrationReadinessReportHash(report)
	return report
}

func BuildMigrationPhase1Readiness(input MigrationPhase1Input) MigrationReadinessReport {
	input = input.Normalize()
	failed := make([]string, 0)
	evidence := make([]string, 0)
	for _, item := range []struct {
		label	string
		hash	string
	}{
		{"aetracore_module", input.AetraCoreModuleHash},
		{"zone_registry_root", input.ZoneRegistryRoot},
		{"default_zone_state_root", input.DefaultZoneStateRoot},
		{"message_root", input.MessageRoot},
		{"empty_queue_root", input.EmptyQueueRoot},
		{"proof_registry_root", input.ProofRegistryRoot},
		{"core_root_hash", input.CoreRootHash},
	} {
		if err := validateHexHash("migration phase 1 "+item.label, item.hash); err != nil {
			failed = append(failed, item.label)
		} else {
			evidence = append(evidence, item.label+":"+item.hash)
		}
	}
	if input.ZoneCount != 1 {
		failed = append(failed, "default_zone_count")
	}
	if strings.TrimSpace(input.DefaultZoneID) == "" {
		failed = append(failed, "default_zone_id")
	}
	if input.MessageRoot != input.EmptyQueueRoot {
		failed = append(failed, "message_root_not_empty_queue")
	}
	if !input.AppHashIncludesCoreRoot {
		failed = append(failed, "app_hash_missing_core_root")
	}
	if err := validateRootQueryAPIs(input.RootQueryAPIs); err != nil {
		failed = append(failed, "root_query_apis")
	} else {
		evidence = append(evidence, "root_query_apis:"+hashRootQueryAPIs(input.RootQueryAPIs))
	}
	if err := validateProofRootMetadata(input.ProofMetadata); err != nil {
		failed = append(failed, "proof_registry_metadata")
	} else {
		evidence = append(evidence, "proof_registry_metadata:"+hashProofMetadata(input.ProofMetadata))
	}
	report := MigrationReadinessReport{
		Phase:		MigrationPhase1CoreCommitments,
		Passed:		len(failed) == 0,
		Failed:		normalizeStringSet(failed),
		Evidence:	normalizeStringSet(evidence),
	}
	report.ReportHash = ComputeMigrationReadinessReportHash(report)
	return report
}

func BuildMigrationPhase2Readiness(input MigrationPhase2Input) MigrationReadinessReport {
	input = input.Normalize()
	failed := make([]string, 0)
	evidence := make([]string, 0)
	for _, item := range []struct {
		label	string
		hash	string
	}{
		{"msgbus_module", input.MsgBusModuleHash},
		{"first_class_message_objects", input.FirstClassObjectRoot},
	} {
		if err := validateHexHash("migration phase 2 "+item.label, item.hash); err != nil {
			failed = append(failed, item.label)
		} else {
			evidence = append(evidence, item.label+":"+item.hash)
		}
	}
	if err := input.Encoding.Validate(); err != nil {
		failed = append(failed, "message_encoding_and_ids")
	} else {
		evidence = append(evidence, "message_encoding_and_ids:"+input.Encoding.MessageIDRoot)
	}
	if err := validateMsgBusStores(input.Stores); err != nil {
		failed = append(failed, "inbox_outbox_receipt_stores")
	} else {
		evidence = append(evidence, "message_stores:"+hashMsgBusStores(input.Stores))
	}
	if err := input.LocalExecution.Validate(); err != nil {
		failed = append(failed, "local_zone_message_execution")
	} else {
		evidence = append(evidence, "local_zone_message_execution:"+input.LocalExecution.ExecutionRoot)
	}
	if err := input.Safety.Validate(); err != nil {
		failed = append(failed, "expiry_bounce_inclusion_receipt_proofs")
	} else {
		evidence = append(evidence, "message_safety:"+hashMsgBusSafety(input.Safety))
	}
	report := MigrationReadinessReport{
		Phase:		MigrationPhase2MessageBus,
		Passed:		len(failed) == 0,
		Failed:		normalizeStringSet(failed),
		Evidence:	normalizeStringSet(evidence),
	}
	report.ReportHash = ComputeMigrationReadinessReportHash(report)
	return report
}

func BuildMigrationPhase3Readiness(input MigrationPhase3Input) MigrationReadinessReport {
	input = input.Normalize()
	failed := make([]string, 0)
	evidence := make([]string, 0)
	for _, zone := range []ZoneExtractionCheck{input.FinancialZone, input.IdentityZone, input.ApplicationZone} {
		if err := zone.Validate(); err != nil {
			failed = append(failed, "zone_extraction:"+zone.ZoneID)
		} else {
			evidence = append(evidence, "zone_extraction:"+zone.ZoneID+":"+zone.CommittedRoot)
		}
	}
	if !input.BankFeesContractAssetsAVMAMMInFinancial {
		failed = append(failed, "financial_zone_modules")
	}
	if !input.IdentityIsolatedActivation {
		failed = append(failed, "identity_zone_isolated_activation")
	}
	if !input.ZoneRootsCommittedPerBlock {
		failed = append(failed, "zone_roots_committed_per_block")
	}
	if err := validateHexHash("migration phase 3 zone commitment root", input.ZoneCommitmentRoot); err != nil {
		failed = append(failed, "zone_commitment_root")
	} else {
		evidence = append(evidence, "zone_commitment_root:"+input.ZoneCommitmentRoot)
	}
	report := MigrationReadinessReport{
		Phase:		MigrationPhase3ZoneExtraction,
		Passed:		len(failed) == 0,
		Failed:		normalizeStringSet(failed),
		Evidence:	normalizeStringSet(evidence),
	}
	report.ReportHash = ComputeMigrationReadinessReportHash(report)
	return report
}

func BuildMigrationPhase4Readiness(input MigrationPhase4Input) MigrationReadinessReport {
	input = input.Normalize()
	failed := make([]string, 0)
	evidence := make([]string, 0)
	for _, item := range []struct {
		label	string
		hash	string
	}{
		{"shards_module", input.ShardsModuleHash},
		{"shard_layout_descriptors", input.ShardLayoutDescriptorRoot},
		{"route_key_calculation", input.RouteKeyCalculationHash},
		{"shard_root_aggregation", input.RootAggregationHash},
	} {
		if err := validateHexHash("migration phase 4 "+item.label, item.hash); err != nil {
			failed = append(failed, item.label)
		} else {
			evidence = append(evidence, item.label+":"+item.hash)
		}
	}
	if err := validateExecutionToken("migration phase 4 zone id", input.ZoneID); err != nil {
		failed = append(failed, "zone_id")
	}
	if err := validateShardRuntimeDescriptors(input.ShardDescriptors); err != nil {
		failed = append(failed, "per_shard_runtime_descriptors")
	} else {
		evidence = append(evidence, "per_shard_runtime_descriptors:"+hashShardRuntimeDescriptors(input.ShardDescriptors))
	}
	if !input.ZonesSupportMultipleShards || activeShardCount(input.ShardDescriptors) < 2 {
		failed = append(failed, "zones_support_multiple_shards")
	}
	if !input.IndependentWorkloadsParallel {
		failed = append(failed, "independent_shard_workloads_parallel")
	}
	if err := input.SplitMergeScheduler.Validate(); err != nil {
		failed = append(failed, "split_merge_scheduler")
	} else {
		evidence = append(evidence, "split_merge_scheduler:"+hashShardSplitMergeScheduler(input.SplitMergeScheduler))
	}
	if err := input.Migration.Validate(); err != nil {
		failed = append(failed, "deterministic_shard_migration")
	} else {
		evidence = append(evidence, "deterministic_shard_migration:"+hashShardMigration(input.Migration))
	}
	if !input.InFlightMessagesSurviveChange {
		failed = append(failed, "in_flight_messages_survive_layout_changes")
	}
	report := MigrationReadinessReport{
		Phase:		MigrationPhase4ShardingRuntime,
		Passed:		len(failed) == 0,
		Failed:		normalizeStringSet(failed),
		Evidence:	normalizeStringSet(evidence),
	}
	report.ReportHash = ComputeMigrationReadinessReportHash(report)
	return report
}

func BuildMigrationPhase5Readiness(input MigrationPhase5Input) MigrationReadinessReport {
	input = input.Normalize()
	failed := make([]string, 0)
	evidence := make([]string, 0)
	for _, component := range []AVM20ComponentCheck{
		input.BytecodeFormat,
		input.Interpreter,
		input.GasTable,
		input.ContractStorageAdapter,
		input.ABIRegistry,
	} {
		if err := component.Validate(); err != nil {
			failed = append(failed, "avm_component:"+component.ComponentName)
		} else {
			evidence = append(evidence, "avm_component:"+component.ComponentName+":"+component.ComponentHash)
		}
	}
	if err := validateAVM20Syscalls("message", input.MessageSyscalls); err != nil {
		failed = append(failed, "contract_message_syscalls")
	} else {
		evidence = append(evidence, "contract_message_syscalls:"+hashAVM20Syscalls("message", input.MessageSyscalls))
	}
	if err := validateAVM20Syscalls("proof", input.ProofVerificationSyscalls); err != nil {
		failed = append(failed, "proof_verification_syscalls")
	} else {
		evidence = append(evidence, "proof_verification_syscalls:"+hashAVM20Syscalls("proof", input.ProofVerificationSyscalls))
	}
	for _, item := range []struct {
		label	string
		hash	string
	}{
		{"async_message_emission_root", input.AsyncMessageEmissionRoot},
		{"contract_state_proof_root", input.ContractStateProofRoot},
		{"contract_zone_execution_root", input.ContractZoneExecutionRoot},
	} {
		if err := validateHexHash("migration phase 5 "+item.label, item.hash); err != nil {
			failed = append(failed, item.label)
		} else {
			evidence = append(evidence, item.label+":"+item.hash)
		}
	}
	if !input.ContractZoneDeterministic {
		failed = append(failed, "contract_zone_deterministic_execution")
	}
	report := MigrationReadinessReport{
		Phase:		MigrationPhase5AVM20,
		Passed:		len(failed) == 0,
		Failed:		normalizeStringSet(failed),
		Evidence:	normalizeStringSet(evidence),
	}
	report.ReportHash = ComputeMigrationReadinessReportHash(report)
	return report
}

func BuildMigrationPhase6Readiness(input MigrationPhase6Input) MigrationReadinessReport {
	input = input.Normalize()
	failed := make([]string, 0)
	evidence := make([]string, 0)
	for _, item := range []struct {
		label	string
		hash	string
	}{
		{"aet_identity_proofs", input.AETIdentityProofRoot},
		{"cross_zone_identity_lookup_messages", input.CrossZoneIdentityLookupRoot},
		{"payment_channel_settlement", input.PaymentChannelSettlementRoot},
		{"conditional_payment_routing", input.ConditionalPaymentRoutingRoot},
		{"payment_proof_apis", input.PaymentProofAPIRoot},
	} {
		if err := validateHexHash("migration phase 6 "+item.label, item.hash); err != nil {
			failed = append(failed, item.label)
		} else {
			evidence = append(evidence, item.label+":"+item.hash)
		}
	}
	if err := validateIdentityPaymentFlows("identity", input.IdentityFlows); err != nil {
		failed = append(failed, "proof_backed_identity_flows")
	} else {
		evidence = append(evidence, "proof_backed_identity_flows:"+hashIdentityPaymentFlows("identity", input.IdentityFlows))
	}
	if err := validateIdentityPaymentFlows("payment", input.PaymentFlows); err != nil {
		failed = append(failed, "trustless_payment_flows")
	} else {
		evidence = append(evidence, "trustless_payment_flows:"+hashIdentityPaymentFlows("payment", input.PaymentFlows))
	}
	if err := validateWalletSDKHelpers(input.WalletSDKHelpers); err != nil {
		failed = append(failed, "wallet_sdk_helpers")
	} else {
		evidence = append(evidence, "wallet_sdk_helpers:"+hashWalletSDKHelpers(input.WalletSDKHelpers))
	}
	if !input.NamesResolveThroughIdentityZone {
		failed = append(failed, "names_resolve_through_identity_zone")
	}
	if !input.PaymentsSettleThroughFinancialZone {
		failed = append(failed, "payments_settle_through_financial_zone")
	}
	if !input.ContractsUseAsyncIdentityPayments {
		failed = append(failed, "contracts_use_async_identity_payment_messages")
	}
	report := MigrationReadinessReport{
		Phase:		MigrationPhase6IdentityPayments,
		Passed:		len(failed) == 0,
		Failed:		normalizeStringSet(failed),
		Evidence:	normalizeStringSet(evidence),
	}
	report.ReportHash = ComputeMigrationReadinessReportHash(report)
	return report
}

func BuildMigrationPhase7Readiness(input MigrationPhase7Input) MigrationReadinessReport {
	input = input.Normalize()
	failed := make([]string, 0)
	evidence := make([]string, 0)
	for _, item := range []struct {
		label	string
		hash	string
	}{
		{"blockstm_workloads", input.BlockSTMWorkloadRoot},
		{"conflict_profiling", input.ConflictProfilingRoot},
		{"mempool_lanes", input.MempoolLaneRoot},
		{"congestion_aware_routing", input.CongestionRoutingRoot},
		{"adaptive_sync_recovery_tests", input.AdaptiveSyncRecoveryRoot},
		{"multi_zone_load_simulation", input.MultiZoneLoadSimulationRoot},
	} {
		if err := validateHexHash("migration phase 7 "+item.label, item.hash); err != nil {
			failed = append(failed, item.label)
		} else {
			evidence = append(evidence, item.label+":"+item.hash)
		}
	}
	if err := validatePerformanceHardeningChecks(input.HardeningChecks); err != nil {
		failed = append(failed, "performance_hardening_checks")
	} else {
		evidence = append(evidence, "performance_hardening_checks:"+hashPerformanceHardeningChecks(input.HardeningChecks))
	}
	if err := validateStoreV2Benchmarks(input.StoreV2Benchmarks); err != nil {
		failed = append(failed, "store_v2_benchmarks")
	} else {
		evidence = append(evidence, "store_v2_benchmarks:"+hashStoreV2Benchmarks(input.StoreV2Benchmarks))
	}
	if !input.IndependentExecutionScalesParallel {
		failed = append(failed, "independent_execution_scales_parallel")
	}
	if !input.StateSyncRecoversCommitments {
		failed = append(failed, "state_sync_recovers_zone_shard_commitments")
	}
	if !input.RoutingDeterministicUnderCongestion {
		failed = append(failed, "routing_deterministic_under_congestion")
	}
	report := MigrationReadinessReport{
		Phase:		MigrationPhase7Performance,
		Passed:		len(failed) == 0,
		Failed:		normalizeStringSet(failed),
		Evidence:	normalizeStringSet(evidence),
	}
	report.ReportHash = ComputeMigrationReadinessReportHash(report)
	return report
}

func (i MigrationPhase0Input) Normalize() MigrationPhase0Input {
	i.ModuleBoundaryDocHash = normalizeLowerHex(i.ModuleBoundaryDocHash)
	i.StateExportValidationHash = normalizeLowerHex(i.StateExportValidationHash)
	i.ExportedAppHash = normalizeLowerHex(i.ExportedAppHash)
	i.ReplayedAppHash = normalizeLowerHex(i.ReplayedAppHash)
	i.DynamicFeeBoundsTestHash = normalizeLowerHex(i.DynamicFeeBoundsTestHash)
	i.StoreV2CompatibilityHash = normalizeLowerHex(i.StoreV2CompatibilityHash)
	for idx := range i.GenesisImports {
		i.GenesisImports[idx] = i.GenesisImports[idx].Normalize()
	}
	sort.SliceStable(i.GenesisImports, func(left, right int) bool {
		return i.GenesisImports[left].ModuleName < i.GenesisImports[right].ModuleName
	})
	for idx := range i.InvariantChecks {
		i.InvariantChecks[idx] = i.InvariantChecks[idx].Normalize()
	}
	sort.SliceStable(i.InvariantChecks, func(left, right int) bool {
		return invariantKey(i.InvariantChecks[left]) < invariantKey(i.InvariantChecks[right])
	})
	for idx := range i.PrefixMigrations {
		i.PrefixMigrations[idx] = i.PrefixMigrations[idx].Normalize()
	}
	sort.SliceStable(i.PrefixMigrations, func(left, right int) bool {
		return i.PrefixMigrations[left].ModuleName < i.PrefixMigrations[right].ModuleName
	})
	return i
}

func (i MigrationPhase1Input) Normalize() MigrationPhase1Input {
	i.AetraCoreModuleHash = normalizeLowerHex(i.AetraCoreModuleHash)
	i.ZoneRegistryRoot = normalizeLowerHex(i.ZoneRegistryRoot)
	i.DefaultZoneID = strings.TrimSpace(i.DefaultZoneID)
	i.DefaultZoneStateRoot = normalizeLowerHex(i.DefaultZoneStateRoot)
	i.MessageRoot = normalizeLowerHex(i.MessageRoot)
	i.EmptyQueueRoot = normalizeLowerHex(i.EmptyQueueRoot)
	i.ProofRegistryRoot = normalizeLowerHex(i.ProofRegistryRoot)
	i.CoreRootHash = normalizeLowerHex(i.CoreRootHash)
	for idx := range i.RootQueryAPIs {
		i.RootQueryAPIs[idx] = i.RootQueryAPIs[idx].Normalize()
	}
	sort.SliceStable(i.RootQueryAPIs, func(left, right int) bool {
		return i.RootQueryAPIs[left].QueryName < i.RootQueryAPIs[right].QueryName
	})
	for idx := range i.ProofMetadata {
		i.ProofMetadata[idx] = i.ProofMetadata[idx].Normalize()
	}
	sort.SliceStable(i.ProofMetadata, func(left, right int) bool {
		return string(i.ProofMetadata[left].RootType) < string(i.ProofMetadata[right].RootType)
	})
	return i
}

func (i MigrationPhase2Input) Normalize() MigrationPhase2Input {
	i.MsgBusModuleHash = normalizeLowerHex(i.MsgBusModuleHash)
	i.Encoding = i.Encoding.Normalize()
	for idx := range i.Stores {
		i.Stores[idx] = i.Stores[idx].Normalize()
	}
	sort.SliceStable(i.Stores, func(left, right int) bool {
		return i.Stores[left].StoreName < i.Stores[right].StoreName
	})
	i.LocalExecution = i.LocalExecution.Normalize()
	i.Safety = i.Safety.Normalize()
	i.FirstClassObjectRoot = normalizeLowerHex(i.FirstClassObjectRoot)
	return i
}

func (i MigrationPhase3Input) Normalize() MigrationPhase3Input {
	i.FinancialZone = i.FinancialZone.Normalize()
	i.IdentityZone = i.IdentityZone.Normalize()
	i.ApplicationZone = i.ApplicationZone.Normalize()
	i.ZoneCommitmentRoot = normalizeLowerHex(i.ZoneCommitmentRoot)
	return i
}

func (i MigrationPhase4Input) Normalize() MigrationPhase4Input {
	i.ShardsModuleHash = normalizeLowerHex(i.ShardsModuleHash)
	i.ZoneID = strings.TrimSpace(i.ZoneID)
	i.ShardLayoutDescriptorRoot = normalizeLowerHex(i.ShardLayoutDescriptorRoot)
	i.RouteKeyCalculationHash = normalizeLowerHex(i.RouteKeyCalculationHash)
	for idx := range i.ShardDescriptors {
		i.ShardDescriptors[idx] = i.ShardDescriptors[idx].Normalize()
	}
	sort.SliceStable(i.ShardDescriptors, func(left, right int) bool {
		return i.ShardDescriptors[left].ShardID < i.ShardDescriptors[right].ShardID
	})
	i.RootAggregationHash = normalizeLowerHex(i.RootAggregationHash)
	i.SplitMergeScheduler = i.SplitMergeScheduler.Normalize()
	i.Migration = i.Migration.Normalize()
	return i
}

func (i MigrationPhase5Input) Normalize() MigrationPhase5Input {
	i.BytecodeFormat = i.BytecodeFormat.Normalize()
	i.Interpreter = i.Interpreter.Normalize()
	i.GasTable = i.GasTable.Normalize()
	i.ContractStorageAdapter = i.ContractStorageAdapter.Normalize()
	i.ABIRegistry = i.ABIRegistry.Normalize()
	for idx := range i.MessageSyscalls {
		i.MessageSyscalls[idx] = i.MessageSyscalls[idx].Normalize()
	}
	sort.SliceStable(i.MessageSyscalls, func(left, right int) bool {
		return i.MessageSyscalls[left].SyscallName < i.MessageSyscalls[right].SyscallName
	})
	for idx := range i.ProofVerificationSyscalls {
		i.ProofVerificationSyscalls[idx] = i.ProofVerificationSyscalls[idx].Normalize()
	}
	sort.SliceStable(i.ProofVerificationSyscalls, func(left, right int) bool {
		return i.ProofVerificationSyscalls[left].SyscallName < i.ProofVerificationSyscalls[right].SyscallName
	})
	i.AsyncMessageEmissionRoot = normalizeLowerHex(i.AsyncMessageEmissionRoot)
	i.ContractStateProofRoot = normalizeLowerHex(i.ContractStateProofRoot)
	i.ContractZoneExecutionRoot = normalizeLowerHex(i.ContractZoneExecutionRoot)
	return i
}

func (i MigrationPhase6Input) Normalize() MigrationPhase6Input {
	i.AETIdentityProofRoot = normalizeLowerHex(i.AETIdentityProofRoot)
	i.CrossZoneIdentityLookupRoot = normalizeLowerHex(i.CrossZoneIdentityLookupRoot)
	i.PaymentChannelSettlementRoot = normalizeLowerHex(i.PaymentChannelSettlementRoot)
	i.ConditionalPaymentRoutingRoot = normalizeLowerHex(i.ConditionalPaymentRoutingRoot)
	i.PaymentProofAPIRoot = normalizeLowerHex(i.PaymentProofAPIRoot)
	for idx := range i.IdentityFlows {
		i.IdentityFlows[idx] = i.IdentityFlows[idx].Normalize()
	}
	sort.SliceStable(i.IdentityFlows, func(left, right int) bool {
		return i.IdentityFlows[left].FlowName < i.IdentityFlows[right].FlowName
	})
	for idx := range i.PaymentFlows {
		i.PaymentFlows[idx] = i.PaymentFlows[idx].Normalize()
	}
	sort.SliceStable(i.PaymentFlows, func(left, right int) bool {
		return i.PaymentFlows[left].FlowName < i.PaymentFlows[right].FlowName
	})
	for idx := range i.WalletSDKHelpers {
		i.WalletSDKHelpers[idx] = i.WalletSDKHelpers[idx].Normalize()
	}
	sort.SliceStable(i.WalletSDKHelpers, func(left, right int) bool {
		return i.WalletSDKHelpers[left].HelperName < i.WalletSDKHelpers[right].HelperName
	})
	return i
}

func (i MigrationPhase7Input) Normalize() MigrationPhase7Input {
	i.BlockSTMWorkloadRoot = normalizeLowerHex(i.BlockSTMWorkloadRoot)
	i.ConflictProfilingRoot = normalizeLowerHex(i.ConflictProfilingRoot)
	i.MempoolLaneRoot = normalizeLowerHex(i.MempoolLaneRoot)
	i.CongestionRoutingRoot = normalizeLowerHex(i.CongestionRoutingRoot)
	i.AdaptiveSyncRecoveryRoot = normalizeLowerHex(i.AdaptiveSyncRecoveryRoot)
	i.MultiZoneLoadSimulationRoot = normalizeLowerHex(i.MultiZoneLoadSimulationRoot)
	for idx := range i.HardeningChecks {
		i.HardeningChecks[idx] = i.HardeningChecks[idx].Normalize()
	}
	sort.SliceStable(i.HardeningChecks, func(left, right int) bool {
		return i.HardeningChecks[left].CheckName < i.HardeningChecks[right].CheckName
	})
	for idx := range i.StoreV2Benchmarks {
		i.StoreV2Benchmarks[idx] = i.StoreV2Benchmarks[idx].Normalize()
	}
	sort.SliceStable(i.StoreV2Benchmarks, func(left, right int) bool {
		return i.StoreV2Benchmarks[left].BenchmarkName < i.StoreV2Benchmarks[right].BenchmarkName
	})
	return i
}

func (c GenesisImportCheck) Normalize() GenesisImportCheck {
	c.ModuleName = strings.TrimSpace(c.ModuleName)
	c.ExportHash = normalizeLowerHex(c.ExportHash)
	c.ImportHash = normalizeLowerHex(c.ImportHash)
	return c
}

func (c GenesisImportCheck) Validate() error {
	check := c.Normalize()
	if err := validateExecutionToken("migration genesis import module", check.ModuleName); err != nil {
		return err
	}
	if !check.Active {
		return nil
	}
	if !check.Deterministic {
		return errors.New("migration active module genesis import must be deterministic")
	}
	if err := validateHexHash("migration genesis export hash", check.ExportHash); err != nil {
		return err
	}
	if err := validateHexHash("migration genesis import hash", check.ImportHash); err != nil {
		return err
	}
	if check.ExportHash != check.ImportHash {
		return errors.New("migration genesis import hash must reproduce export hash")
	}
	return nil
}

func (c ModuleInvariantCheck) Normalize() ModuleInvariantCheck {
	c.ModuleName = strings.TrimSpace(c.ModuleName)
	c.InvariantName = strings.TrimSpace(c.InvariantName)
	c.EvidenceHash = normalizeLowerHex(c.EvidenceHash)
	return c
}

func (c ModuleInvariantCheck) Validate() error {
	check := c.Normalize()
	if err := validateExecutionToken("migration invariant module", check.ModuleName); err != nil {
		return err
	}
	if err := validateExecutionToken("migration invariant name", check.InvariantName); err != nil {
		return err
	}
	if !check.Covered || !check.Deterministic {
		return errors.New("migration invariant must be covered and deterministic")
	}
	return validateHexHash("migration invariant evidence hash", check.EvidenceHash)
}

func (c StatePrefixMigrationCheck) Normalize() StatePrefixMigrationCheck {
	c.ModuleName = strings.TrimSpace(c.ModuleName)
	c.OldPrefix = strings.TrimSpace(c.OldPrefix)
	c.NewPrefix = strings.TrimSpace(c.NewPrefix)
	c.MigrationHash = normalizeLowerHex(c.MigrationHash)
	c.ReversibleProof = normalizeLowerHex(c.ReversibleProof)
	return c
}

func (c StatePrefixMigrationCheck) Validate() error {
	check := c.Normalize()
	if err := validateExecutionToken("migration prefix module", check.ModuleName); err != nil {
		return err
	}
	if check.OldPrefix == "" || check.NewPrefix == "" || check.OldPrefix == check.NewPrefix {
		return errors.New("migration state prefixes must be non-empty and changed")
	}
	if !check.Safe {
		return errors.New("migration state prefix migration must be marked safe")
	}
	if err := validateHexHash("migration prefix migration hash", check.MigrationHash); err != nil {
		return err
	}
	return validateHexHash("migration prefix reversible proof", check.ReversibleProof)
}

func (c RootQueryAPICheck) Normalize() RootQueryAPICheck {
	c.QueryName = strings.TrimSpace(c.QueryName)
	c.ResponseHash = normalizeLowerHex(c.ResponseHash)
	return c
}

func (c RootQueryAPICheck) Validate() error {
	check := c.Normalize()
	if err := validateExecutionToken("migration root query API", check.QueryName); err != nil {
		return err
	}
	if !check.Available {
		return errors.New("migration root query API must be available")
	}
	if !IsProofRootType(check.RootType) {
		return errors.New("migration root query API root type is unsupported")
	}
	return validateHexHash("migration root query response hash", check.ResponseHash)
}

func (c ProofRootMetadataCheck) Normalize() ProofRootMetadataCheck {
	c.RootHash = normalizeLowerHex(c.RootHash)
	c.MetadataHash = normalizeLowerHex(c.MetadataHash)
	return c
}

func (c ProofRootMetadataCheck) Validate() error {
	check := c.Normalize()
	if !IsProofRootType(check.RootType) {
		return errors.New("migration proof metadata root type is unsupported")
	}
	if check.Height == 0 {
		return errors.New("migration proof metadata height must be positive")
	}
	if err := validateHexHash("migration proof metadata root hash", check.RootHash); err != nil {
		return err
	}
	return validateHexHash("migration proof metadata hash", check.MetadataHash)
}

func (c MsgBusStoreCheck) Normalize() MsgBusStoreCheck {
	c.StoreName = strings.TrimSpace(c.StoreName)
	c.RootHash = normalizeLowerHex(c.RootHash)
	return c
}

func (c MsgBusStoreCheck) Validate() error {
	check := c.Normalize()
	if err := validateExecutionToken("migration msgbus store name", check.StoreName); err != nil {
		return err
	}
	if !check.Committed {
		return errors.New("migration msgbus store must be committed")
	}
	return validateHexHash("migration msgbus store root", check.RootHash)
}

func (c MsgBusEncodingCheck) Normalize() MsgBusEncodingCheck {
	c.CodecHash = normalizeLowerHex(c.CodecHash)
	c.MessageIDRoot = normalizeLowerHex(c.MessageIDRoot)
	return c
}

func (c MsgBusEncodingCheck) Validate() error {
	check := c.Normalize()
	if !check.DeterministicIDs {
		return errors.New("migration msgbus message ids must be deterministic")
	}
	if err := validateHexHash("migration msgbus codec hash", check.CodecHash); err != nil {
		return err
	}
	return validateHexHash("migration msgbus message id root", check.MessageIDRoot)
}

func (c MsgBusExecutionCheck) Normalize() MsgBusExecutionCheck {
	c.ExecutionRoot = normalizeLowerHex(c.ExecutionRoot)
	return c
}

func (c MsgBusExecutionCheck) Validate() error {
	check := c.Normalize()
	if !check.Deterministic || !check.ExecutedLocally {
		return errors.New("migration msgbus local async execution must be deterministic and local")
	}
	return validateHexHash("migration msgbus local execution root", check.ExecutionRoot)
}

func (c MsgBusSafetyCheck) Normalize() MsgBusSafetyCheck {
	c.ExpiryRoot = normalizeLowerHex(c.ExpiryRoot)
	c.BounceRoot = normalizeLowerHex(c.BounceRoot)
	c.InclusionProofRoot = normalizeLowerHex(c.InclusionProofRoot)
	c.ReceiptsProofRoot = normalizeLowerHex(c.ReceiptsProofRoot)
	return c
}

func (c MsgBusSafetyCheck) Validate() error {
	check := c.Normalize()
	for _, item := range []struct {
		name	string
		value	string
	}{
		{"migration msgbus expiry root", check.ExpiryRoot},
		{"migration msgbus bounce root", check.BounceRoot},
		{"migration msgbus inclusion proof root", check.InclusionProofRoot},
		{"migration msgbus receipts proof root", check.ReceiptsProofRoot},
	} {
		if err := validateHexHash(item.name, item.value); err != nil {
			return err
		}
	}
	return nil
}

func (c ZoneExtractionCheck) Normalize() ZoneExtractionCheck {
	c.ZoneID = strings.TrimSpace(c.ZoneID)
	c.KeeperHash = normalizeLowerHex(c.KeeperHash)
	c.StatePrefixRoot = normalizeLowerHex(c.StatePrefixRoot)
	c.FeePolicyHash = normalizeLowerHex(c.FeePolicyHash)
	c.ExecutionSummaryHash = normalizeLowerHex(c.ExecutionSummaryHash)
	c.CommittedRoot = normalizeLowerHex(c.CommittedRoot)
	c.Modules = normalizeStringSet(c.Modules)
	return c
}

func (c ZoneExtractionCheck) Validate() error {
	check := c.Normalize()
	if err := validateExecutionToken("migration extracted zone id", check.ZoneID); err != nil {
		return err
	}
	if !check.Extracted {
		return errors.New("migration zone must be extracted")
	}
	if len(check.Modules) == 0 {
		return errors.New("migration extracted zone requires modules")
	}
	for _, item := range []struct {
		name	string
		value	string
	}{
		{"migration zone keeper hash", check.KeeperHash},
		{"migration zone state prefix root", check.StatePrefixRoot},
		{"migration zone fee policy hash", check.FeePolicyHash},
		{"migration zone execution summary hash", check.ExecutionSummaryHash},
		{"migration zone committed root", check.CommittedRoot},
	} {
		if err := validateHexHash(item.name, item.value); err != nil {
			return err
		}
	}
	return nil
}

func (c ShardRuntimeDescriptorCheck) Normalize() ShardRuntimeDescriptorCheck {
	c.ShardID = strings.TrimSpace(c.ShardID)
	c.LayoutHash = normalizeLowerHex(c.LayoutHash)
	c.RouteKeyRoot = normalizeLowerHex(c.RouteKeyRoot)
	c.InboxRoot = normalizeLowerHex(c.InboxRoot)
	c.OutboxRoot = normalizeLowerHex(c.OutboxRoot)
	c.ShardRoot = normalizeLowerHex(c.ShardRoot)
	c.ParallelGroupHash = normalizeLowerHex(c.ParallelGroupHash)
	return c
}

func (c ShardRuntimeDescriptorCheck) Validate() error {
	check := c.Normalize()
	if err := validateExecutionToken("migration shard id", check.ShardID); err != nil {
		return err
	}
	if !check.Active {
		return errors.New("migration shard descriptor must be active")
	}
	for _, item := range []struct {
		name	string
		value	string
	}{
		{"migration shard layout hash", check.LayoutHash},
		{"migration shard route key root", check.RouteKeyRoot},
		{"migration shard inbox root", check.InboxRoot},
		{"migration shard outbox root", check.OutboxRoot},
		{"migration shard state root", check.ShardRoot},
		{"migration shard parallel group hash", check.ParallelGroupHash},
	} {
		if err := validateHexHash(item.name, item.value); err != nil {
			return err
		}
	}
	return nil
}

func (c ShardSplitMergeSchedulerCheck) Normalize() ShardSplitMergeSchedulerCheck {
	c.SchedulerRoot = normalizeLowerHex(c.SchedulerRoot)
	c.SplitDecisionRoot = normalizeLowerHex(c.SplitDecisionRoot)
	c.MergeDecisionRoot = normalizeLowerHex(c.MergeDecisionRoot)
	return c
}

func (c ShardSplitMergeSchedulerCheck) Validate() error {
	check := c.Normalize()
	if !check.Deterministic {
		return errors.New("migration shard split merge scheduler must be deterministic")
	}
	for _, item := range []struct {
		name	string
		value	string
	}{
		{"migration shard scheduler root", check.SchedulerRoot},
		{"migration shard split decision root", check.SplitDecisionRoot},
		{"migration shard merge decision root", check.MergeDecisionRoot},
	} {
		if err := validateHexHash(item.name, item.value); err != nil {
			return err
		}
	}
	return nil
}

func (c ShardMigrationCheck) Normalize() ShardMigrationCheck {
	c.MigrationRoot = normalizeLowerHex(c.MigrationRoot)
	c.OldLayoutHash = normalizeLowerHex(c.OldLayoutHash)
	c.NewLayoutHash = normalizeLowerHex(c.NewLayoutHash)
	c.InFlightMessageRoot = normalizeLowerHex(c.InFlightMessageRoot)
	return c
}

func (c ShardMigrationCheck) Validate() error {
	check := c.Normalize()
	if !check.DeterministicMigration {
		return errors.New("migration shard migration must be deterministic")
	}
	if !check.SurvivesLayoutChange {
		return errors.New("migration shard migration must preserve in-flight messages")
	}
	for _, item := range []struct {
		name	string
		value	string
	}{
		{"migration shard migration root", check.MigrationRoot},
		{"migration shard old layout hash", check.OldLayoutHash},
		{"migration shard new layout hash", check.NewLayoutHash},
		{"migration shard in-flight message root", check.InFlightMessageRoot},
	} {
		if err := validateHexHash(item.name, item.value); err != nil {
			return err
		}
	}
	if check.OldLayoutHash == check.NewLayoutHash {
		return errors.New("migration shard layout hashes must change")
	}
	return nil
}

func (c AVM20ComponentCheck) Normalize() AVM20ComponentCheck {
	c.ComponentName = strings.TrimSpace(c.ComponentName)
	c.ComponentHash = normalizeLowerHex(c.ComponentHash)
	return c
}

func (c AVM20ComponentCheck) Validate() error {
	check := c.Normalize()
	if err := validateExecutionToken("migration AVM component name", check.ComponentName); err != nil {
		return err
	}
	if !check.Implemented || !check.Deterministic {
		return errors.New("migration AVM component must be implemented and deterministic")
	}
	return validateHexHash("migration AVM component hash", check.ComponentHash)
}

func (c AVM20SyscallCheck) Normalize() AVM20SyscallCheck {
	c.SyscallName = strings.TrimSpace(c.SyscallName)
	c.SyscallHash = normalizeLowerHex(c.SyscallHash)
	return c
}

func (c AVM20SyscallCheck) Validate() error {
	check := c.Normalize()
	if err := validateExecutionToken("migration AVM syscall name", check.SyscallName); err != nil {
		return err
	}
	if !check.Enabled || !check.Metered {
		return errors.New("migration AVM syscall must be enabled and metered")
	}
	return validateHexHash("migration AVM syscall hash", check.SyscallHash)
}

func (c IdentityPaymentFlowCheck) Normalize() IdentityPaymentFlowCheck {
	c.FlowName = strings.TrimSpace(c.FlowName)
	c.FlowRoot = normalizeLowerHex(c.FlowRoot)
	c.ZoneID = strings.TrimSpace(c.ZoneID)
	c.MessageTypeHash = normalizeLowerHex(c.MessageTypeHash)
	return c
}

func (c IdentityPaymentFlowCheck) Validate() error {
	check := c.Normalize()
	if err := validateExecutionToken("migration identity payment flow name", check.FlowName); err != nil {
		return err
	}
	if err := validateExecutionToken("migration identity payment flow zone id", check.ZoneID); err != nil {
		return err
	}
	if !check.ProofBacked || !check.Asynchronous || !check.Deterministic {
		return errors.New("migration identity payment flow must be proof-backed, asynchronous, and deterministic")
	}
	if err := validateHexHash("migration identity payment flow root", check.FlowRoot); err != nil {
		return err
	}
	return validateHexHash("migration identity payment flow message type hash", check.MessageTypeHash)
}

func (c WalletSDKHelperCheck) Normalize() WalletSDKHelperCheck {
	c.HelperName = strings.TrimSpace(c.HelperName)
	c.HelperHash = normalizeLowerHex(c.HelperHash)
	return c
}

func (c WalletSDKHelperCheck) Validate() error {
	check := c.Normalize()
	if err := validateExecutionToken("migration wallet SDK helper name", check.HelperName); err != nil {
		return err
	}
	if !check.Available || !check.Deterministic {
		return errors.New("migration wallet SDK helper must be available and deterministic")
	}
	return validateHexHash("migration wallet SDK helper hash", check.HelperHash)
}

func (c PerformanceHardeningCheck) Normalize() PerformanceHardeningCheck {
	c.CheckName = strings.TrimSpace(c.CheckName)
	c.EvidenceHash = normalizeLowerHex(c.EvidenceHash)
	return c
}

func (c PerformanceHardeningCheck) Validate() error {
	check := c.Normalize()
	if err := validateExecutionToken("migration performance hardening check name", check.CheckName); err != nil {
		return err
	}
	if !check.Enabled || !check.Deterministic {
		return errors.New("migration performance hardening check must be enabled and deterministic")
	}
	return validateHexHash("migration performance hardening evidence hash", check.EvidenceHash)
}

func (c StoreV2BenchmarkCheck) Normalize() StoreV2BenchmarkCheck {
	c.BenchmarkName = strings.TrimSpace(c.BenchmarkName)
	c.ResultHash = normalizeLowerHex(c.ResultHash)
	return c
}

func (c StoreV2BenchmarkCheck) Validate() error {
	check := c.Normalize()
	if err := validateExecutionToken("migration Store v2 benchmark name", check.BenchmarkName); err != nil {
		return err
	}
	if !check.Covered || !check.Bounded {
		return errors.New("migration Store v2 benchmark must be covered and bounded")
	}
	return validateHexHash("migration Store v2 benchmark result hash", check.ResultHash)
}

func (r MigrationReadinessReport) Validate() error {
	if r.Phase != MigrationPhase0BaselineHardening &&
		r.Phase != MigrationPhase1CoreCommitments &&
		r.Phase != MigrationPhase2MessageBus &&
		r.Phase != MigrationPhase3ZoneExtraction &&
		r.Phase != MigrationPhase4ShardingRuntime &&
		r.Phase != MigrationPhase5AVM20 &&
		r.Phase != MigrationPhase6IdentityPayments &&
		r.Phase != MigrationPhase7Performance {
		return errors.New("migration readiness phase is unsupported")
	}
	if r.Passed && len(r.Failed) > 0 {
		return errors.New("migration readiness passed report must not include failures")
	}
	if len(r.Evidence) == 0 {
		return errors.New("migration readiness evidence is required")
	}
	if r.ReportHash != ComputeMigrationReadinessReportHash(r) {
		return errors.New("migration readiness report hash mismatch")
	}
	return nil
}

func validateMsgBusStores(checks []MsgBusStoreCheck) error {
	required := map[string]struct{}{
		"inbox":	{},
		"outbox":	{},
		"receipt":	{},
	}
	for _, check := range checks {
		if err := check.Validate(); err != nil {
			return err
		}
		if _, found := required[check.StoreName]; found {
			delete(required, check.StoreName)
		}
	}
	if len(required) > 0 {
		return fmt.Errorf("migration missing msgbus stores: %v", sortedMapKeys(required))
	}
	return nil
}

func validateShardRuntimeDescriptors(checks []ShardRuntimeDescriptorCheck) error {
	seen := make(map[string]struct{}, len(checks))
	for _, check := range checks {
		if err := check.Validate(); err != nil {
			return err
		}
		check = check.Normalize()
		if _, exists := seen[check.ShardID]; exists {
			return fmt.Errorf("migration duplicate shard descriptor: %s", check.ShardID)
		}
		seen[check.ShardID] = struct{}{}
	}
	if len(seen) == 0 {
		return errors.New("migration shard descriptors are required")
	}
	return nil
}

func activeShardCount(checks []ShardRuntimeDescriptorCheck) int {
	count := 0
	for _, check := range checks {
		if check.Normalize().Active {
			count++
		}
	}
	return count
}

func validateAVM20Syscalls(kind string, checks []AVM20SyscallCheck) error {
	seen := make(map[string]struct{}, len(checks))
	for _, check := range checks {
		if err := check.Validate(); err != nil {
			return err
		}
		check = check.Normalize()
		if _, exists := seen[check.SyscallName]; exists {
			return fmt.Errorf("migration duplicate AVM %s syscall: %s", kind, check.SyscallName)
		}
		seen[check.SyscallName] = struct{}{}
	}
	if len(seen) == 0 {
		return fmt.Errorf("migration AVM %s syscalls are required", kind)
	}
	return nil
}

func validateIdentityPaymentFlows(kind string, checks []IdentityPaymentFlowCheck) error {
	seen := make(map[string]struct{}, len(checks))
	for _, check := range checks {
		if err := check.Validate(); err != nil {
			return err
		}
		check = check.Normalize()
		if _, exists := seen[check.FlowName]; exists {
			return fmt.Errorf("migration duplicate %s flow: %s", kind, check.FlowName)
		}
		seen[check.FlowName] = struct{}{}
	}
	if len(seen) == 0 {
		return fmt.Errorf("migration %s flows are required", kind)
	}
	return nil
}

func validateWalletSDKHelpers(checks []WalletSDKHelperCheck) error {
	required := map[string]struct{}{
		"identity_lookup":	{},
		"payment_route":	{},
	}
	for _, check := range checks {
		if err := check.Validate(); err != nil {
			return err
		}
		check = check.Normalize()
		if _, found := required[check.HelperName]; found {
			delete(required, check.HelperName)
		}
	}
	if len(required) > 0 {
		return fmt.Errorf("migration missing wallet SDK helpers: %v", sortedMapKeys(required))
	}
	return nil
}

func validatePerformanceHardeningChecks(checks []PerformanceHardeningCheck) error {
	required := map[string]struct{}{
		"blockstm_zone_shard_batches":	{},
		"conflict_profiling":		{},
		"mempool_lanes":		{},
		"congestion_aware_routing":	{},
		"adaptive_sync_recovery":	{},
		"multi_zone_load_simulation":	{},
	}
	for _, check := range checks {
		if err := check.Validate(); err != nil {
			return err
		}
		check = check.Normalize()
		if _, found := required[check.CheckName]; found {
			delete(required, check.CheckName)
		}
	}
	if len(required) > 0 {
		return fmt.Errorf("migration missing performance hardening checks: %v", sortedMapKeys(required))
	}
	return nil
}

func validateStoreV2Benchmarks(checks []StoreV2BenchmarkCheck) error {
	required := map[string]struct{}{
		"direct_balance_read":			{},
		"direct_identity_resolution":		{},
		"recursive_identity_resolution":	{},
		"contract_storage_read_write":		{},
		"message_enqueue_dequeue":		{},
		"payment_channel_settle":		{},
		"proof_generation":			{},
	}
	for _, check := range checks {
		if err := check.Validate(); err != nil {
			return err
		}
		check = check.Normalize()
		if _, found := required[check.BenchmarkName]; found {
			delete(required, check.BenchmarkName)
		}
	}
	if len(required) > 0 {
		return fmt.Errorf("migration missing Store v2 benchmarks: %v", sortedMapKeys(required))
	}
	return nil
}

func validateRequiredInvariantCoverage(checks []ModuleInvariantCheck) error {
	required := map[string]struct{}{
		"staking":	{},
		"slashing":	{},
		"bank":		{},
		"distribution":	{},
	}
	for _, check := range checks {
		if err := check.Validate(); err != nil {
			return err
		}
		if _, found := required[check.ModuleName]; found {
			delete(required, check.ModuleName)
		}
	}
	if len(required) > 0 {
		return fmt.Errorf("migration missing required invariants: %v", sortedMapKeys(required))
	}
	return nil
}

func validateRootQueryAPIs(checks []RootQueryAPICheck) error {
	required := map[ProofRootType]struct{}{
		ProofRootZone:		{},
		ProofRootMessage:	{},
		ProofRootStorage:	{},
	}
	for _, check := range checks {
		if err := check.Validate(); err != nil {
			return err
		}
		if _, found := required[check.RootType]; found {
			delete(required, check.RootType)
		}
	}
	if len(required) > 0 {
		return fmt.Errorf("migration missing root query APIs: %v", required)
	}
	return nil
}

func validateProofRootMetadata(checks []ProofRootMetadataCheck) error {
	required := map[ProofRootType]struct{}{
		ProofRootZone:		{},
		ProofRootMessage:	{},
		ProofRootStorage:	{},
	}
	for _, check := range checks {
		if err := check.Validate(); err != nil {
			return err
		}
		if _, found := required[check.RootType]; found {
			delete(required, check.RootType)
		}
	}
	if len(required) > 0 {
		return fmt.Errorf("migration missing proof metadata: %v", required)
	}
	return nil
}

func ComputeMigrationReadinessReportHash(report MigrationReadinessReport) string {
	failed := normalizeStringSet(report.Failed)
	evidence := normalizeStringSet(report.Evidence)
	parts := []string{"migration-readiness-report", string(report.Phase), fmt.Sprintf("%t", report.Passed)}
	parts = append(parts, failed...)
	parts = append(parts, evidence...)
	return hashStrings(parts...)
}

func hashInvariantChecks(checks []ModuleInvariantCheck) string {
	parts := []string{"migration-invariant-checks"}
	for _, check := range checks {
		check = check.Normalize()
		parts = append(parts, check.ModuleName, check.InvariantName, check.EvidenceHash)
	}
	return hashStrings(parts...)
}

func hashRootQueryAPIs(checks []RootQueryAPICheck) string {
	parts := []string{"migration-root-query-apis"}
	for _, check := range checks {
		check = check.Normalize()
		parts = append(parts, check.QueryName, string(check.RootType), check.ResponseHash)
	}
	return hashStrings(parts...)
}

func hashProofMetadata(checks []ProofRootMetadataCheck) string {
	parts := []string{"migration-proof-metadata"}
	for _, check := range checks {
		check = check.Normalize()
		parts = append(parts, string(check.RootType), fmt.Sprintf("%020d", check.Height), check.RootHash, check.MetadataHash)
	}
	return hashStrings(parts...)
}

func hashMsgBusStores(checks []MsgBusStoreCheck) string {
	parts := []string{"migration-msgbus-stores"}
	for _, check := range checks {
		check = check.Normalize()
		parts = append(parts, check.StoreName, check.RootHash, fmt.Sprintf("%t", check.Committed))
	}
	return hashStrings(parts...)
}

func hashMsgBusSafety(check MsgBusSafetyCheck) string {
	check = check.Normalize()
	return hashStrings("migration-msgbus-safety", check.ExpiryRoot, check.BounceRoot, check.InclusionProofRoot, check.ReceiptsProofRoot)
}

func hashShardRuntimeDescriptors(checks []ShardRuntimeDescriptorCheck) string {
	parts := []string{"migration-shard-runtime-descriptors"}
	for _, check := range checks {
		check = check.Normalize()
		parts = append(
			parts,
			check.ShardID,
			check.LayoutHash,
			check.RouteKeyRoot,
			check.InboxRoot,
			check.OutboxRoot,
			check.ShardRoot,
			check.ParallelGroupHash,
			fmt.Sprintf("%t", check.Active),
		)
	}
	return hashStrings(parts...)
}

func hashShardSplitMergeScheduler(check ShardSplitMergeSchedulerCheck) string {
	check = check.Normalize()
	return hashStrings("migration-shard-split-merge-scheduler", check.SchedulerRoot, check.SplitDecisionRoot, check.MergeDecisionRoot, fmt.Sprintf("%t", check.Deterministic))
}

func hashShardMigration(check ShardMigrationCheck) string {
	check = check.Normalize()
	return hashStrings(
		"migration-shard-migration",
		check.MigrationRoot,
		check.OldLayoutHash,
		check.NewLayoutHash,
		check.InFlightMessageRoot,
		fmt.Sprintf("%t", check.SurvivesLayoutChange),
		fmt.Sprintf("%t", check.DeterministicMigration),
	)
}

func hashAVM20Syscalls(kind string, checks []AVM20SyscallCheck) string {
	parts := []string{"migration-avm20-syscalls", kind}
	for _, check := range checks {
		check = check.Normalize()
		parts = append(parts, check.SyscallName, check.SyscallHash, fmt.Sprintf("%t", check.Metered), fmt.Sprintf("%t", check.Enabled))
	}
	return hashStrings(parts...)
}

func hashIdentityPaymentFlows(kind string, checks []IdentityPaymentFlowCheck) string {
	parts := []string{"migration-identity-payment-flows", kind}
	for _, check := range checks {
		check = check.Normalize()
		parts = append(
			parts,
			check.FlowName,
			check.FlowRoot,
			check.ZoneID,
			check.MessageTypeHash,
			fmt.Sprintf("%t", check.ProofBacked),
			fmt.Sprintf("%t", check.Asynchronous),
			fmt.Sprintf("%t", check.Deterministic),
		)
	}
	return hashStrings(parts...)
}

func hashWalletSDKHelpers(checks []WalletSDKHelperCheck) string {
	parts := []string{"migration-wallet-sdk-helpers"}
	for _, check := range checks {
		check = check.Normalize()
		parts = append(parts, check.HelperName, check.HelperHash, fmt.Sprintf("%t", check.Available), fmt.Sprintf("%t", check.Deterministic))
	}
	return hashStrings(parts...)
}

func hashPerformanceHardeningChecks(checks []PerformanceHardeningCheck) string {
	parts := []string{"migration-performance-hardening-checks"}
	for _, check := range checks {
		check = check.Normalize()
		parts = append(parts, check.CheckName, check.EvidenceHash, fmt.Sprintf("%t", check.Enabled), fmt.Sprintf("%t", check.Deterministic))
	}
	return hashStrings(parts...)
}

func hashStoreV2Benchmarks(checks []StoreV2BenchmarkCheck) string {
	parts := []string{"migration-store-v2-benchmarks"}
	for _, check := range checks {
		check = check.Normalize()
		parts = append(parts, check.BenchmarkName, check.ResultHash, fmt.Sprintf("%t", check.Covered), fmt.Sprintf("%t", check.Bounded))
	}
	return hashStrings(parts...)
}

func invariantKey(check ModuleInvariantCheck) string {
	return check.ModuleName + "/" + check.InvariantName
}

func sortedMapKeys[T any](values map[string]T) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
