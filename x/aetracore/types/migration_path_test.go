package types

import "testing"

func TestMigrationPathSpecCoversPhaseZeroThroughSeven(t *testing.T) {
	spec, err := DefaultMigrationPathSpec()
	if err != nil {
		t.Fatalf("default migration path spec: %v", err)
	}
	if err := spec.Validate(); err != nil {
		t.Fatalf("validate migration path spec: %v", err)
	}
	if err := ValidateMigrationPathCoverage(); err != nil {
		t.Fatalf("migration path coverage: %v", err)
	}
	if len(spec.Phases) != 8 {
		t.Fatalf("expected 8 phases, got %d", len(spec.Phases))
	}

	phaseByID := map[MigrationPhaseID]MigrationPhase{}
	for _, phase := range spec.Phases {
		phaseByID[phase.PhaseID] = phase
	}

	phase0 := phaseByID[MigrationPhaseBaselineHardening]
	if len(phase0.Tasks) != 6 {
		t.Fatalf("expected phase 0 to cover 6 tasks, got %d", len(phase0.Tasks))
	}
	if len(phase0.ExitCriteria) != 3 {
		t.Fatalf("expected phase 0 to cover 3 exit criteria, got %d", len(phase0.ExitCriteria))
	}

	phase1 := phaseByID[MigrationPhaseCoreCommitments]
	if len(phase1.Tasks) != 6 {
		t.Fatalf("expected phase 1 to cover 6 tasks, got %d", len(phase1.Tasks))
	}
	if len(phase1.ExitCriteria) != 3 {
		t.Fatalf("expected phase 1 to cover 3 exit criteria, got %d", len(phase1.ExitCriteria))
	}

	phase2 := phaseByID[MigrationPhaseMessageBus]
	if len(phase2.Tasks) != 6 {
		t.Fatalf("expected phase 2 to cover 6 tasks, got %d", len(phase2.Tasks))
	}
	if len(phase2.ExitCriteria) != 3 {
		t.Fatalf("expected phase 2 to cover 3 exit criteria, got %d", len(phase2.ExitCriteria))
	}

	phase3 := phaseByID[MigrationPhaseZoneExtraction]
	if len(phase3.Tasks) != 6 {
		t.Fatalf("expected phase 3 to cover 6 tasks, got %d", len(phase3.Tasks))
	}
	if len(phase3.ExitCriteria) != 3 {
		t.Fatalf("expected phase 3 to cover 3 exit criteria, got %d", len(phase3.ExitCriteria))
	}

	phase4 := phaseByID[MigrationPhaseShardingRuntime]
	if len(phase4.Tasks) != 7 {
		t.Fatalf("expected phase 4 to cover 7 tasks, got %d", len(phase4.Tasks))
	}
	if len(phase4.ExitCriteria) != 3 {
		t.Fatalf("expected phase 4 to cover 3 exit criteria, got %d", len(phase4.ExitCriteria))
	}

	phase5 := phaseByID[MigrationPhaseAVM20]
	if len(phase5.Tasks) != 7 {
		t.Fatalf("expected phase 5 to cover 7 tasks, got %d", len(phase5.Tasks))
	}
	if len(phase5.ExitCriteria) != 3 {
		t.Fatalf("expected phase 5 to cover 3 exit criteria, got %d", len(phase5.ExitCriteria))
	}

	phase6 := phaseByID[MigrationPhaseIdentityPayments]
	if len(phase6.Tasks) != 6 {
		t.Fatalf("expected phase 6 to cover 6 tasks, got %d", len(phase6.Tasks))
	}
	if len(phase6.ExitCriteria) != 3 {
		t.Fatalf("expected phase 6 to cover 3 exit criteria, got %d", len(phase6.ExitCriteria))
	}

	phase7 := phaseByID[MigrationPhasePerformanceHardening]
	if len(phase7.Tasks) != 7 {
		t.Fatalf("expected phase 7 to cover 7 tasks, got %d", len(phase7.Tasks))
	}
	if len(phase7.ExitCriteria) != 3 {
		t.Fatalf("expected phase 7 to cover 3 exit criteria, got %d", len(phase7.ExitCriteria))
	}
}

func TestMigrationPathSpecRootCanonicalAndRejectsTamper(t *testing.T) {
	defaultSpec, err := DefaultMigrationPathSpec()
	if err != nil {
		t.Fatalf("default migration path spec: %v", err)
	}

	reordered, err := BuildMigrationPathSpec([]MigrationPhase{
		migrationPhase(MigrationPhaseAVM20, "Phase 5: AVM 2.0", MigrationPhase5Tasks(), MigrationPhase5ExitCriteria()),
		migrationPhase(MigrationPhaseZoneExtraction, "Phase 3: Zone Extraction", MigrationPhase3Tasks(), MigrationPhase3ExitCriteria()),
		migrationPhase(MigrationPhasePerformanceHardening, "Phase 7: Performance Hardening", MigrationPhase7Tasks(), MigrationPhase7ExitCriteria()),
		migrationPhase(MigrationPhaseBaselineHardening, "Phase 0: Baseline Hardening", MigrationPhase0Tasks(), MigrationPhase0ExitCriteria()),
		migrationPhase(MigrationPhaseShardingRuntime, "Phase 4: Sharding Runtime", MigrationPhase4Tasks(), MigrationPhase4ExitCriteria()),
		migrationPhase(MigrationPhaseIdentityPayments, "Phase 6: Identity and Payment Integration", MigrationPhase6Tasks(), MigrationPhase6ExitCriteria()),
		migrationPhase(MigrationPhaseMessageBus, "Phase 2: Message Bus", MigrationPhase2Tasks(), MigrationPhase2ExitCriteria()),
		migrationPhase(MigrationPhaseCoreCommitments, "Phase 1: Core Commitments", MigrationPhase1Tasks(), MigrationPhase1ExitCriteria()),
	})
	if err != nil {
		t.Fatalf("build reordered migration path spec: %v", err)
	}
	if reordered.Root != defaultSpec.Root {
		t.Fatalf("canonical migration root mismatch: %s != %s", reordered.Root, defaultSpec.Root)
	}

	if _, err := BuildMigrationPathSpec([]MigrationPhase{defaultSpec.Phases[0], defaultSpec.Phases[0]}); err == nil {
		t.Fatal("expected duplicate migration phases to fail")
	}

	tampered := defaultSpec
	tampered.Phases[0].Tasks[0].DescriptorHash = hashParts("tampered migration task")
	if err := tampered.Validate(); err == nil {
		t.Fatal("expected tampered migration task hash to fail")
	}
}

func TestBaselineHardeningEvidenceRequiresAllExitCriteria(t *testing.T) {
	evidence := validBaselineHardeningEvidence()
	if err := evidence.Validate(); err != nil {
		t.Fatalf("baseline evidence should validate: %v", err)
	}

	notExportable := evidence
	notExportable.StateExportable = false
	notExportable.EvidenceHash = ComputeBaselineHardeningEvidenceHash(notExportable)
	if err := notExportable.Validate(); err == nil {
		t.Fatal("expected non-exportable baseline evidence to fail")
	}

	missingInvariantCoverage := evidence
	missingInvariantCoverage.InvariantCoverage = false
	missingInvariantCoverage.EvidenceHash = ComputeBaselineHardeningEvidenceHash(missingInvariantCoverage)
	if err := missingInvariantCoverage.Validate(); err == nil {
		t.Fatal("expected missing invariant coverage to fail")
	}

	unsafePrefixMigration := evidence
	unsafePrefixMigration.PrefixMigrationSafe = false
	unsafePrefixMigration.EvidenceHash = ComputeBaselineHardeningEvidenceHash(unsafePrefixMigration)
	if err := unsafePrefixMigration.Validate(); err == nil {
		t.Fatal("expected unsafe prefix migration to fail")
	}

	tamperedHash := evidence
	tamperedHash.EvidenceHash = hashParts("different baseline evidence")
	if err := tamperedHash.Validate(); err == nil {
		t.Fatal("expected tampered baseline evidence hash to fail")
	}
}

func TestCoreCommitmentMigrationEvidenceRequiresDefaultZoneRootsAndProofRegistry(t *testing.T) {
	evidence := validCoreCommitmentMigrationEvidence()
	if err := evidence.Validate(); err != nil {
		t.Fatalf("core commitment evidence should validate: %v", err)
	}

	nonEmptyMessageRoot := evidence
	nonEmptyMessageRoot.EmptyMessageRoot = hashParts("not empty")
	nonEmptyMessageRoot.EvidenceHash = ComputeCoreCommitmentMigrationEvidenceHash(nonEmptyMessageRoot)
	if err := nonEmptyMessageRoot.Validate(); err == nil {
		t.Fatal("expected non-empty message root evidence to fail")
	}

	notSingleZone := evidence
	notSingleZone.SingleZoneMode = false
	notSingleZone.EvidenceHash = ComputeCoreCommitmentMigrationEvidenceHash(notSingleZone)
	if err := notSingleZone.Validate(); err == nil {
		t.Fatal("expected non single-zone evidence to fail")
	}

	missingMetadata := evidence
	missingMetadata.ProofRegistryMetadata = false
	missingMetadata.EvidenceHash = ComputeCoreCommitmentMigrationEvidenceHash(missingMetadata)
	if err := missingMetadata.Validate(); err == nil {
		t.Fatal("expected missing proof registry metadata to fail")
	}

	badZone := evidence
	badZone.DefaultZoneID = ZoneID("default")
	badZone.EvidenceHash = ComputeCoreCommitmentMigrationEvidenceHash(badZone)
	if err := badZone.Validate(); err == nil {
		t.Fatal("expected invalid default zone id to fail")
	}
}

func TestMessageBusMigrationEvidenceRequiresCommittedMessagesAndProofReceipts(t *testing.T) {
	evidence := validMessageBusMigrationEvidence()
	if err := evidence.Validate(); err != nil {
		t.Fatalf("message bus evidence should validate: %v", err)
	}

	notCommitted := evidence
	notCommitted.MessagesCommitted = false
	notCommitted.EvidenceHash = ComputeMessageBusMigrationEvidenceHash(notCommitted)
	if err := notCommitted.Validate(); err == nil {
		t.Fatal("expected uncommitted message evidence to fail")
	}

	notDeterministic := evidence
	notDeterministic.LocalAsyncDeterministic = false
	notDeterministic.EvidenceHash = ComputeMessageBusMigrationEvidenceHash(notDeterministic)
	if err := notDeterministic.Validate(); err == nil {
		t.Fatal("expected non-deterministic local async evidence to fail")
	}

	notQueryable := evidence
	notQueryable.ReceiptProofQueryable = false
	notQueryable.EvidenceHash = ComputeMessageBusMigrationEvidenceHash(notQueryable)
	if err := notQueryable.Validate(); err == nil {
		t.Fatal("expected non-queryable receipt evidence to fail")
	}

	tamperedHash := evidence
	tamperedHash.EvidenceHash = hashParts("different message bus evidence")
	if err := tamperedHash.Validate(); err == nil {
		t.Fatal("expected tampered message bus evidence hash to fail")
	}
}

func TestZoneExtractionMigrationEvidenceRequiresZoneIsolationAndPerBlockRoots(t *testing.T) {
	evidence := validZoneExtractionMigrationEvidence()
	if err := evidence.Validate(); err != nil {
		t.Fatalf("zone extraction evidence should validate: %v", err)
	}

	financialNotRouted := evidence
	financialNotRouted.FinancialModulesRouted = false
	financialNotRouted.EvidenceHash = ComputeZoneExtractionMigrationEvidenceHash(financialNotRouted)
	if err := financialNotRouted.Validate(); err == nil {
		t.Fatal("expected financial modules not routed evidence to fail")
	}

	identityNotIsolated := evidence
	identityNotIsolated.IdentityIsolated = false
	identityNotIsolated.EvidenceHash = ComputeZoneExtractionMigrationEvidenceHash(identityNotIsolated)
	if err := identityNotIsolated.Validate(); err == nil {
		t.Fatal("expected non-isolated identity evidence to fail")
	}

	missingRoots := evidence
	missingRoots.ZoneRootsCommittedPerBlock = false
	missingRoots.EvidenceHash = ComputeZoneExtractionMigrationEvidenceHash(missingRoots)
	if err := missingRoots.Validate(); err == nil {
		t.Fatal("expected missing per-block zone roots evidence to fail")
	}

	tamperedHash := evidence
	tamperedHash.EvidenceHash = hashParts("different zone extraction evidence")
	if err := tamperedHash.Validate(); err == nil {
		t.Fatal("expected tampered zone extraction evidence hash to fail")
	}
}

func TestShardingRuntimeMigrationEvidenceRequiresMultiShardParallelAndSafeMessages(t *testing.T) {
	evidence := validShardingRuntimeMigrationEvidence()
	if err := evidence.Validate(); err != nil {
		t.Fatalf("sharding runtime evidence should validate: %v", err)
	}

	notMultiShard := evidence
	notMultiShard.MultiShardZones = false
	notMultiShard.EvidenceHash = ComputeShardingRuntimeMigrationEvidenceHash(notMultiShard)
	if err := notMultiShard.Validate(); err == nil {
		t.Fatal("expected non multi-shard evidence to fail")
	}

	notParallel := evidence
	notParallel.ParallelShardWorkloads = false
	notParallel.EvidenceHash = ComputeShardingRuntimeMigrationEvidenceHash(notParallel)
	if err := notParallel.Validate(); err == nil {
		t.Fatal("expected non-parallel shard workload evidence to fail")
	}

	unsafeMessages := evidence
	unsafeMessages.InflightMessagesSafe = false
	unsafeMessages.EvidenceHash = ComputeShardingRuntimeMigrationEvidenceHash(unsafeMessages)
	if err := unsafeMessages.Validate(); err == nil {
		t.Fatal("expected unsafe in-flight message evidence to fail")
	}

	tamperedHash := evidence
	tamperedHash.EvidenceHash = hashParts("different sharding runtime evidence")
	if err := tamperedHash.Validate(); err == nil {
		t.Fatal("expected tampered sharding runtime evidence hash to fail")
	}
}

func TestAVM20MigrationEvidenceRequiresDeterministicContractsMessagesAndProofs(t *testing.T) {
	evidence := validAVM20MigrationEvidence()
	if err := evidence.Validate(); err != nil {
		t.Fatalf("AVM 2.0 evidence should validate: %v", err)
	}

	notDeterministic := evidence
	notDeterministic.DeterministicContracts = false
	notDeterministic.EvidenceHash = ComputeAVM20MigrationEvidenceHash(notDeterministic)
	if err := notDeterministic.Validate(); err == nil {
		t.Fatal("expected non-deterministic AVM evidence to fail")
	}

	noAsyncMessages := evidence
	noAsyncMessages.AsyncMessages = false
	noAsyncMessages.EvidenceHash = ComputeAVM20MigrationEvidenceHash(noAsyncMessages)
	if err := noAsyncMessages.Validate(); err == nil {
		t.Fatal("expected missing async messages evidence to fail")
	}

	noProofs := evidence
	noProofs.ContractProofsAvailable = false
	noProofs.EvidenceHash = ComputeAVM20MigrationEvidenceHash(noProofs)
	if err := noProofs.Validate(); err == nil {
		t.Fatal("expected missing contract proofs evidence to fail")
	}

	tamperedHash := evidence
	tamperedHash.EvidenceHash = hashParts("different AVM evidence")
	if err := tamperedHash.Validate(); err == nil {
		t.Fatal("expected tampered AVM evidence hash to fail")
	}
}

func TestIdentityPaymentIntegrationEvidenceRequiresProofNamesSettlementAndAsyncMessages(t *testing.T) {
	evidence := validIdentityPaymentIntegrationEvidence()
	if err := evidence.Validate(); err != nil {
		t.Fatalf("identity/payment evidence should validate: %v", err)
	}

	noProofNames := evidence
	noProofNames.ProofBackedNames = false
	noProofNames.EvidenceHash = ComputeIdentityPaymentIntegrationEvidenceHash(noProofNames)
	if err := noProofNames.Validate(); err == nil {
		t.Fatal("expected missing proof-backed names evidence to fail")
	}

	noTrustlessPayments := evidence
	noTrustlessPayments.TrustlessPayments = false
	noTrustlessPayments.EvidenceHash = ComputeIdentityPaymentIntegrationEvidenceHash(noTrustlessPayments)
	if err := noTrustlessPayments.Validate(); err == nil {
		t.Fatal("expected missing trustless payment settlement evidence to fail")
	}

	noAsyncMessages := evidence
	noAsyncMessages.AsyncContractMessages = false
	noAsyncMessages.EvidenceHash = ComputeIdentityPaymentIntegrationEvidenceHash(noAsyncMessages)
	if err := noAsyncMessages.Validate(); err == nil {
		t.Fatal("expected missing async contract identity/payment message evidence to fail")
	}

	tamperedHash := evidence
	tamperedHash.EvidenceHash = hashParts("different identity payment evidence")
	if err := tamperedHash.Validate(); err == nil {
		t.Fatal("expected tampered identity/payment evidence hash to fail")
	}
}

func TestPerformanceHardeningMigrationEvidenceRequiresScalingSyncAndCongestionDeterminism(t *testing.T) {
	evidence := validPerformanceHardeningMigrationEvidence()
	if err := evidence.Validate(); err != nil {
		t.Fatalf("performance evidence should validate: %v", err)
	}

	noScaling := evidence
	noScaling.ParallelismScales = false
	noScaling.EvidenceHash = ComputePerformanceHardeningMigrationEvidenceHash(noScaling)
	if err := noScaling.Validate(); err == nil {
		t.Fatal("expected missing parallelism scaling evidence to fail")
	}

	noRecovery := evidence
	noRecovery.StateSyncRecoversRoots = false
	noRecovery.EvidenceHash = ComputePerformanceHardeningMigrationEvidenceHash(noRecovery)
	if err := noRecovery.Validate(); err == nil {
		t.Fatal("expected missing state sync recovery evidence to fail")
	}

	nondeterministicRouting := evidence
	nondeterministicRouting.CongestionDeterministic = false
	nondeterministicRouting.EvidenceHash = ComputePerformanceHardeningMigrationEvidenceHash(nondeterministicRouting)
	if err := nondeterministicRouting.Validate(); err == nil {
		t.Fatal("expected non-deterministic congestion routing evidence to fail")
	}

	tamperedHash := evidence
	tamperedHash.EvidenceHash = hashParts("different performance evidence")
	if err := tamperedHash.Validate(); err == nil {
		t.Fatal("expected tampered performance evidence hash to fail")
	}
}

func validBaselineHardeningEvidence() BaselineHardeningEvidence {
	evidence := BaselineHardeningEvidence{
		ModuleBoundaryDocsRoot:		hashParts("module boundary docs"),
		StateExportManifestHash:	hashParts("export manifest"),
		GenesisImportHash:		hashParts("genesis import"),
		DynamicFeeBoundsTestHash:	hashParts("dynamic fee bounds tests"),
		LegacyInvariantRoot:		hashParts("staking slashing bank distribution invariants"),
		StoreV2AuditHash:		hashParts("store v2 compatibility audit"),
		UpgradeHandlerPrefixHash:	hashParts("upgrade handler prefix migration"),
		StateReproducible:		true,
		StateExportable:		true,
		InvariantCoverage:		true,
		PrefixMigrationSafe:		true,
	}
	evidence.EvidenceHash = ComputeBaselineHardeningEvidenceHash(evidence)
	return evidence
}

func validCoreCommitmentMigrationEvidence() CoreCommitmentMigrationEvidence {
	evidence := CoreCommitmentMigrationEvidence{
		AetraCoreModuleHash:		hashParts("x/aetracore"),
		DefaultZoneDescriptorHash:	hashParts("default zone descriptor"),
		DefaultZoneStateRoot:		hashParts("default zone state root"),
		EmptyMessageRoot:		EmptyRootHash,
		ProofRegistryRoot:		hashParts("proof root registry"),
		RootQueryAPIHash:		hashParts("root query APIs"),
		AppHashCoreRoot:		hashParts("app hash includes core root"),
		DefaultZoneID:			ZoneID("DEFAULT"),
		SingleZoneMode:			true,
		ProofRegistryMetadata:		true,
	}
	evidence.EvidenceHash = ComputeCoreCommitmentMigrationEvidenceHash(evidence)
	return evidence
}

func validMessageBusMigrationEvidence() MessageBusMigrationEvidence {
	evidence := MessageBusMigrationEvidence{
		MsgbusModuleHash:		hashParts("x/msgbus"),
		MessageCodecHash:		hashParts("canonical message encoding"),
		MessageIDDerivationHash:	hashParts("message id derivation"),
		InboxStoreRoot:			hashParts("inbox store root"),
		OutboxStoreRoot:		hashParts("outbox store root"),
		ReceiptStoreRoot:		hashParts("receipt store root"),
		LocalExecutionRoot:		hashParts("local zone message execution"),
		ExpiryBounceRoot:		hashParts("expiry and bounce logic"),
		InclusionProofRoot:		hashParts("message inclusion proof"),
		MessagesCommitted:		true,
		LocalAsyncDeterministic:	true,
		ReceiptProofQueryable:		true,
	}
	evidence.EvidenceHash = ComputeMessageBusMigrationEvidenceHash(evidence)
	return evidence
}

func validZoneExtractionMigrationEvidence() ZoneExtractionMigrationEvidence {
	evidence := ZoneExtractionMigrationEvidence{
		FinancialZoneRoot:		hashParts("financial zone root"),
		IdentityZoneRoot:		hashParts("identity zone root"),
		ApplicationZoneRoot:		hashParts("application zone root"),
		ZoneKeeperRoot:			hashParts("zone keeper root"),
		ZonePrefixRoot:			hashParts("zone prefix root"),
		ZoneFeePolicyRoot:		hashParts("zone fee policy root"),
		ZoneExecutionSummaryRoot:	hashParts("zone execution summary root"),
		FinancialModulesRouted:		true,
		IdentityIsolated:		true,
		ZoneRootsCommittedPerBlock:	true,
	}
	evidence.EvidenceHash = ComputeZoneExtractionMigrationEvidenceHash(evidence)
	return evidence
}

func validShardingRuntimeMigrationEvidence() ShardingRuntimeMigrationEvidence {
	evidence := ShardingRuntimeMigrationEvidence{
		ShardsModuleHash:		hashParts("x/shards"),
		ShardLayoutRoot:		hashParts("shard layout descriptors"),
		RouteKeyCalculationRoot:	hashParts("route key calculation"),
		PerShardInboxRoot:		hashParts("per shard inbox root"),
		PerShardOutboxRoot:		hashParts("per shard outbox root"),
		ShardRootAggregate:		hashParts("shard root aggregation"),
		SplitMergeScheduleRoot:		hashParts("split merge scheduler"),
		ShardMigrationRoot:		hashParts("deterministic shard migration"),
		MultiShardZones:		true,
		ParallelShardWorkloads:		true,
		InflightMessagesSafe:		true,
	}
	evidence.EvidenceHash = ComputeShardingRuntimeMigrationEvidenceHash(evidence)
	return evidence
}

func validAVM20MigrationEvidence() AVM20MigrationEvidence {
	evidence := AVM20MigrationEvidence{
		BytecodeFormatHash:		hashParts("AVM bytecode format"),
		InterpreterRoot:		hashParts("AVM interpreter"),
		GasTableRoot:			hashParts("AVM gas table"),
		ContractStorageRoot:		hashParts("contract storage adapter"),
		MessageSyscallRoot:		hashParts("contract message syscalls"),
		ProofSyscallRoot:		hashParts("proof verification syscalls"),
		ABIRegistryRoot:		hashParts("ABI registry"),
		DeterministicContracts:		true,
		AsyncMessages:			true,
		ContractProofsAvailable:	true,
	}
	evidence.EvidenceHash = ComputeAVM20MigrationEvidenceHash(evidence)
	return evidence
}

func validIdentityPaymentIntegrationEvidence() IdentityPaymentIntegrationEvidence {
	evidence := IdentityPaymentIntegrationEvidence{
		IdentityProofRoot:		hashParts("identity proof activation"),
		IdentityLookupMessageRoot:	hashParts("cross zone identity lookup messages"),
		PaymentChannelSettlementRoot:	hashParts("payment channel settlement"),
		ConditionalPaymentRouteRoot:	hashParts("conditional payment routing"),
		PaymentProofAPIRoot:		hashParts("payment proof APIs"),
		WalletSDKHelperRoot:		hashParts("wallet SDK identity payment helpers"),
		ProofBackedNames:		true,
		TrustlessPayments:		true,
		AsyncContractMessages:		true,
	}
	evidence.EvidenceHash = ComputeIdentityPaymentIntegrationEvidenceHash(evidence)
	return evidence
}

func validPerformanceHardeningMigrationEvidence() PerformanceHardeningMigrationEvidence {
	evidence := PerformanceHardeningMigrationEvidence{
		BlockSTMWorkloadRoot:		hashParts("BlockSTM zone shard workloads"),
		ConflictProfileRoot:		hashParts("conflict profiling"),
		StoreV2BenchmarkRoot:		hashParts("Store v2 benchmarks"),
		MempoolLaneRoot:		hashParts("mempool lanes"),
		CongestionRoutingRoot:		hashParts("congestion aware routing"),
		AdaptiveSyncRecoveryRoot:	hashParts("AdaptiveSync recovery tests"),
		LoadSimulationRoot:		hashParts("multi zone traffic load simulation"),
		ParallelismScales:		true,
		StateSyncRecoversRoots:		true,
		CongestionDeterministic:	true,
	}
	evidence.EvidenceHash = ComputePerformanceHardeningMigrationEvidenceHash(evidence)
	return evidence
}
