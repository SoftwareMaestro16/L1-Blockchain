package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegisterZonesAndAggregateGlobalRoot(t *testing.T) {
	state := EmptyState()
	var err error
	for _, zone := range []ZoneDescriptor{
		testDescriptor(ZoneIDIdentity, ZoneTypeIdentity, "identity"),
		testDescriptor(ZoneIDFinancial, ZoneTypeFinancial, "financial"),
		testDescriptor(ZoneIDContract, ZoneTypeContract, "contract"),
	} {
		state, err = RegisterZoneDescriptor(state, zone)
		require.NoError(t, err)
	}
	state, err = RegisterServiceDescriptor(state, testService("identity-resolver", ZoneIDIdentity))
	require.NoError(t, err)

	for _, commitment := range []ZoneCommitment{
		testCommitment(t, 10, ZoneIDIdentity),
		testCommitment(t, 10, ZoneIDFinancial),
		testCommitment(t, 10, ZoneIDContract),
	} {
		state, err = AppendZoneCommitment(state, commitment)
		require.NoError(t, err)
	}

	root, err := BuildGlobalStateRoot(10, state, testContributions(10))
	require.NoError(t, err)
	require.NoError(t, root.ValidateHash())

	state, err = AppendGlobalRoot(state, root)
	require.NoError(t, err)
	manifest, err := NewExportManifest(root, testHash("app-hash"), state)
	require.NoError(t, err)
	state, err = AddExportManifest(state, manifest)
	require.NoError(t, err)

	require.Len(t, state.GlobalRoots, 1)
	require.Len(t, state.ExportManifests, 1)
	require.Equal(t, root.GlobalRoot, state.ExportManifests[0].GlobalRoot)
	require.NoError(t, state.Validate())
}

func TestServiceDescriptorsCoverRuntimeModels(t *testing.T) {
	state := EmptyState()
	var err error
	for _, zone := range []ZoneDescriptor{
		testDescriptor(ZoneIDIdentity, ZoneTypeIdentity, "identity"),
		testDescriptor(ZoneIDApplication, ZoneTypeApplication, "application"),
	} {
		state, err = RegisterZoneDescriptor(state, zone)
		require.NoError(t, err)
	}

	for _, service := range []ServiceDescriptor{
		testService("identity-resolver", ZoneIDIdentity),
		testOffChainService("indexer-feed", ZoneIDApplication),
		testMixedService("hybrid-storage", ZoneIDApplication),
		testFogMarketService("fog-compute", ZoneIDApplication),
	} {
		state, err = RegisterServiceDescriptor(state, service)
		require.NoError(t, err)
	}

	require.Len(t, state.ServiceDescriptors, 4)
	require.NoError(t, state.Validate())
	root, err := ComputeServiceRoot(state.ServiceDescriptors)
	require.NoError(t, err)
	require.NoError(t, ValidateHash("test service root", root))
}

func TestServiceDescriptorRejectsInterfaceHashMismatch(t *testing.T) {
	state := EmptyState()
	var err error
	state, err = RegisterZoneDescriptor(state, testDescriptor(ZoneIDIdentity, ZoneTypeIdentity, "identity"))
	require.NoError(t, err)

	service := testService("identity-resolver", ZoneIDIdentity)
	service.Interface.Methods[0].InputSchemaHash = testHash("mutated-input-schema")
	_, err = RegisterServiceDescriptor(state, service)
	require.ErrorContains(t, err, "interface hash mismatch")
}

func TestServiceDescriptorRejectsUnsafeOffChainService(t *testing.T) {
	state := EmptyState()
	var err error
	state, err = RegisterZoneDescriptor(state, testDescriptor(ZoneIDApplication, ZoneTypeApplication, "application"))
	require.NoError(t, err)

	service := testOffChainService("unsafe-indexer", ZoneIDApplication)
	service.Verification = ServiceVerificationDescriptor{
		TrustModel:	ServiceTrustFullyTrusted,
		Model:		ServiceVerificationAdvisory,
	}
	_, err = RegisterServiceDescriptor(state, service)
	require.ErrorContains(t, err, "signed, proof-backed, or economically constrained")
}

func TestServiceDescriptorRejectsMixedServiceWithoutChallengeOrFallback(t *testing.T) {
	state := EmptyState()
	var err error
	state, err = RegisterZoneDescriptor(state, testDescriptor(ZoneIDApplication, ZoneTypeApplication, "application"))
	require.NoError(t, err)

	service := testMixedService("hybrid-storage", ZoneIDApplication)
	service.Execution.ChallengeWindow = 0
	service.Verification.ChallengeWindow = 0
	service.Verification.FallbackServiceID = ""
	_, err = RegisterServiceDescriptor(state, service)
	require.ErrorContains(t, err, "challenge window")
}

func TestServiceByIDReturnsIsolatedDescriptor(t *testing.T) {
	state := EmptyState()
	var err error
	state, err = RegisterZoneDescriptor(state, testDescriptor(ZoneIDIdentity, ZoneTypeIdentity, "identity"))
	require.NoError(t, err)
	state, err = RegisterServiceDescriptor(state, testService("identity-resolver", ZoneIDIdentity))
	require.NoError(t, err)

	descriptor, found := state.ServiceByID("identity-resolver")
	require.True(t, found)
	descriptor.Interface.Methods[0].MethodID = "mutated"

	descriptor, found = state.ServiceByID("identity-resolver")
	require.True(t, found)
	require.Equal(t, "resolve", descriptor.Interface.Methods[0].MethodID)
}

func TestZoneDescriptorCoversKernelSpecificationFields(t *testing.T) {
	descriptor := ZoneDescriptor{
		ZoneID:			ZoneIDApplication,
		ZoneName:		"application",
		ZoneType:		ZoneTypeApplication,
		Enabled:		true,
		StateVersion:		2,
		KeeperScope:		"application.keeper",
		MsgServerScope:		"application.msg",
		QueryServerScope:	"application.query",
		GasPolicyID:		DefaultGasPolicy,
		MessagePolicyID:	DefaultMessagePolicy,
		RootPrefix:		"zone/APPLICATION_ZONE",
		ShardLayoutEpoch:	3,
		UpgradeHeightOptional:	100,
		MessageCapabilities:	[]string{"async-outbox", "async-inbox"},
		ProofCapabilities:	[]string{"state", "receipt"},
		MaxShards:		4,
	}
	descriptor = CanonicalZoneDescriptor(descriptor)
	require.Equal(t, "application", descriptor.ModuleName)
	require.Equal(t, uint64(2), descriptor.StateMachineVersion)
	require.Equal(t, []string{"async-inbox", "async-outbox", "receipt", "state"}, descriptor.Capabilities)
	require.NoError(t, descriptor.Validate(TestnetParams()))
	spec, err := NewZoneDescriptorSpec(descriptor)
	require.NoError(t, err)
	require.NoError(t, spec.Validate())
	require.Equal(t, ZoneID(ZoneIDApplication), spec.ZoneID)
	require.Equal(t, ZoneTypeApplication, spec.ZoneType)
	require.Equal(t, "application", spec.ModuleName)
	require.True(t, spec.Enabled)
	require.Equal(t, uint64(2), spec.StateMachineVersion)
	require.Equal(t, DefaultMempoolPolicy, spec.MempoolPolicyID)
	require.Equal(t, NativeFeePolicyID, spec.FeePolicyID)
	require.Equal(t, uint64(3), spec.ShardLayoutEpoch)
	require.Equal(t, uint32(4), spec.MaxShards)
	require.Equal(t, []string{"async-inbox", "async-outbox"}, spec.MessageCapabilities)
	require.Equal(t, []string{"receipt", "state"}, spec.ProofCapabilities)
	require.Equal(t, uint64(100), spec.UpgradeHeightOptional)

	mutated := descriptor
	mutated.Capabilities = []string{"state", "receipt", "async-outbox", "async-inbox"}
	rootA, err := ComputeZoneDescriptorRoot([]ZoneDescriptor{descriptor}, TestnetParams())
	require.NoError(t, err)
	rootB, err := ComputeZoneDescriptorRoot([]ZoneDescriptor{mutated}, TestnetParams())
	require.NoError(t, err)
	require.Equal(t, rootA, rootB)
	require.Equal(t, ComputeZoneDescriptorHash(descriptor), ComputeZoneDescriptorHash(mutated))
}

func TestZoneCommitmentCoversKernelSpecificationFields(t *testing.T) {
	commitment := testCommitment(t, 21, ZoneIDFinancial)
	require.Equal(t, uint64(21), commitment.Height)
	require.Equal(t, ZoneID(ZoneIDFinancial), commitment.ZoneID)
	require.NotEmpty(t, commitment.StateRoot)
	require.NotEmpty(t, commitment.InboxRoot)
	require.NotEmpty(t, commitment.OutboxRoot)
	require.NotEmpty(t, commitment.ReceiptsRoot)
	require.NotEmpty(t, commitment.EventsRoot)
	require.NotEmpty(t, commitment.ShardRootsRoot)
	require.NotEmpty(t, commitment.ParamsHash)
	require.NotEmpty(t, commitment.ExecutionSummaryHash)
	require.NoError(t, commitment.ValidateHash())
	spec, err := NewZoneCommitmentSpec(commitment)
	require.NoError(t, err)
	require.NoError(t, spec.Validate())
	require.Equal(t, commitment.Height, spec.Height)
	require.Equal(t, commitment.ZoneID, spec.ZoneID)
	require.Equal(t, commitment.StateRoot, spec.ZoneStateRoot)
	require.Equal(t, commitment.OutboxRoot, spec.ZoneMessageOutboxRoot)
	require.Equal(t, commitment.InboxRoot, spec.ZoneMessageInboxRoot)
	require.Equal(t, commitment.ReceiptsRoot, spec.ZoneReceiptRoot)
	require.Equal(t, commitment.EventsRoot, spec.ZoneEventRoot)
	require.Equal(t, commitment.ShardRootsRoot, spec.ShardRootsRoot)
	require.Equal(t, commitment.ExecutionSummaryHash, spec.ExecutionSummaryHash)

	mutated := commitment
	mutated.ShardRootsRoot = testHash("mutated/shards")
	require.ErrorContains(t, mutated.ValidateHash(), "commitment hash mismatch")
}

func TestCoreExecutionPipelineSpecMatchesABCIPhases(t *testing.T) {
	pipeline, err := DefaultCoreExecutionPipelineSpec()
	require.NoError(t, err)
	require.NoError(t, pipeline.ValidateHash())
	require.Len(t, pipeline.Phases, 4)
	require.Equal(t, KernelPhasePrepareProposal, pipeline.Phases[0].Phase)
	require.Equal(t, KernelPhaseProcessProposal, pipeline.Phases[1].Phase)
	require.Equal(t, KernelPhaseFinalizeBlock, pipeline.Phases[2].Phase)
	require.Equal(t, KernelPhaseCommit, pipeline.Phases[3].Phase)
	require.Contains(t, pipeline.Phases[0].DeterministicWork, "group-transactions-by-zone-and-shard")
	require.Contains(t, pipeline.Phases[1].RejectionChecks, "wrong-message-delivery-order")
	require.Contains(t, pipeline.Phases[2].CommittedOutput, "global-message-root")
	require.Contains(t, pipeline.Phases[3].CommittedOutput, "delivery-eligibility")

	mutated := pipeline
	mutated.Phases[0], mutated.Phases[1] = mutated.Phases[1], mutated.Phases[0]
	require.ErrorContains(t, mutated.ValidateHash(), "phase order mismatch")
}

func TestCoreImplementationTasksMatchRoadmapAndReadiness(t *testing.T) {
	tasks, err := DefaultCoreImplementationTasks()
	require.NoError(t, err)
	require.Len(t, tasks, 11)
	require.Equal(t, CoreTaskPriorityP0, taskByID(t, tasks, CoreTaskZoneRegistry).Priority)
	require.Equal(t, CoreTaskPriorityP0, taskByID(t, tasks, CoreTaskProofRootRegistry).Priority)
	require.Equal(t, CoreTaskPriorityP1, taskByID(t, tasks, CoreTaskKeeperIntegration).Priority)
	require.Equal(t, CoreTaskPriorityP2, taskByID(t, tasks, CoreTaskOperationalExportImport).Priority)
	require.Contains(t, taskByID(t, tasks, CoreTaskInboundMessageScheduler).AcceptanceCriteria, "reject-reordered-messages")

	pipeline, err := DefaultCoreExecutionPipelineSpec()
	require.NoError(t, err)
	state := nextReadyState(t,
		[]ZoneID{ZoneIDFinancial, ZoneIDIdentity, ZoneIDApplication, ZoneIDContract},
		[]ZoneID{ZoneIDFinancial, ZoneIDIdentity, ZoneIDApplication, ZoneIDContract},
	)
	state, _ = appendTestFinalityRecord(t, state, 10)
	evidence := DeriveCoreImplementationEvidence(state, pipeline)
	readiness, err := AssessCoreImplementationReadiness(tasks, evidence)
	require.NoError(t, err)
	require.True(t, readiness.Ready)
	require.Empty(t, readiness.RequiredP0Missing)
	require.Contains(t, readiness.MissingTaskIDs, CoreTaskKeeperIntegration)
	require.Contains(t, readiness.MissingTaskIDs, CoreTaskOperationalExportImport)
	require.NotEmpty(t, readiness.ReadinessHash)

	mutated := tasks
	mutated[0].TaskHash = testHash("wrong-task-hash")
	_, err = AssessCoreImplementationReadiness(mutated, evidence)
	require.ErrorContains(t, err, "task hash mismatch")
}

func TestRootReplayIdenticalAcrossNodes(t *testing.T) {
	nodeA := populatedState(t, []ZoneID{ZoneIDFinancial, ZoneIDContract})
	nodeB := populatedState(t, []ZoneID{ZoneIDContract, ZoneIDFinancial})

	rootA, err := BuildGlobalStateRoot(7, nodeA, testContributions(7))
	require.NoError(t, err)
	rootB, err := BuildGlobalStateRoot(7, nodeB, testContributions(7))
	require.NoError(t, err)

	nodeA, err = AppendGlobalRoot(nodeA, rootA)
	require.NoError(t, err)
	nodeB, err = AppendGlobalRoot(nodeB, rootB)
	require.NoError(t, err)

	require.Equal(t, rootA, rootB)
	require.Equal(t, nodeA.Export(), nodeB.Export())
}

func TestUnifiedStateCommitmentModelCommitsExtendedRootSet(t *testing.T) {
	state := populatedState(t, []ZoneID{ZoneIDFinancial, ZoneIDContract})
	root, err := BuildGlobalStateRoot(7, state, testContributions(7))
	require.NoError(t, err)
	require.NoError(t, root.ValidateHash())

	rootSet := root.RootSet()
	require.NoError(t, rootSet.Validate())
	require.Equal(t, root.ZonesRoot, rootSet.ZonesRoot)
	require.Equal(t, root.ServicesRoot, rootSet.ServicesRoot)
	require.Equal(t, root.IdentityRoot, rootSet.IdentityRoot)
	require.Equal(t, root.StorageRoot, rootSet.StorageRoot)
	require.Equal(t, root.MessageRoot, rootSet.MessageRoot)
	require.Equal(t, root.ReceiptsRoot, rootSet.ReceiptsRoot)
	require.Equal(t, root.RoutingRoot, rootSet.RoutingRoot)
	require.Equal(t, root.PaymentsRoot, rootSet.PaymentsRoot)
	require.Equal(t, root.ContractsRoot, rootSet.ContractsRoot)

	rootSetHash, err := ComputeUnifiedStateRootSetHash(rootSet)
	require.NoError(t, err)
	require.NoError(t, ValidateHash("test unified root set hash", rootSetHash))

	manifest, err := NewExportManifest(root, testHash("unified/app-hash"), state)
	require.NoError(t, err)
	require.Equal(t, root.RoutingRoot, manifest.RoutingRoot)
	require.Equal(t, root.ContractsRoot, manifest.ContractsRoot)
	require.NoError(t, manifest.ValidateHash())
}

func TestUnifiedStateCommitmentRejectsExtendedRootTampering(t *testing.T) {
	state := populatedState(t, []ZoneID{ZoneIDFinancial, ZoneIDContract})
	root, err := BuildGlobalStateRoot(7, state, testContributions(7))
	require.NoError(t, err)

	tamperedRouting := root
	tamperedRouting.RoutingRoot = testHash("wrong-routing-root")
	require.ErrorContains(t, tamperedRouting.ValidateHash(), "global root mismatch")

	tamperedContracts := root
	tamperedContracts.ContractsRoot = testHash("wrong-contracts-root")
	require.ErrorContains(t, tamperedContracts.ValidateHash(), "global root mismatch")

	invalidSet := root.RootSet()
	invalidSet.ContractsRoot = "not-a-root"
	_, err = ComputeUnifiedStateRootSetHash(invalidSet)
	require.ErrorContains(t, err, "contracts root")
}

func TestProposalScheduleGroupsByZoneAndShardDeterministically(t *testing.T) {
	items := []ProposalItem{
		testProposalItem(ZoneIDContract, "2", "c", 4, 15, 2),
		testProposalItem(ZoneIDFinancial, "0", "b", 2, 14, 1),
		testProposalItem(ZoneIDContract, "1", "a", 4, 13, 0),
		testProposalItem(ZoneIDFinancial, "0", "a", 1, 16, 3),
	}

	schedule, err := BuildProposalSchedule(9, items, TestnetParams())
	require.NoError(t, err)
	require.NoError(t, schedule.Validate())
	require.Equal(t, []ProposalGroup{
		{
			ZoneID:		ZoneIDContract,
			ShardID:	"1",
			Items:		[]ProposalItem{testProposalItem(ZoneIDContract, "1", "a", 4, 13, 0)},
		},
		{
			ZoneID:		ZoneIDContract,
			ShardID:	"2",
			Items:		[]ProposalItem{testProposalItem(ZoneIDContract, "2", "c", 4, 15, 2)},
		},
		{
			ZoneID:		ZoneIDFinancial,
			ShardID:	"0",
			Items: []ProposalItem{
				testProposalItem(ZoneIDFinancial, "0", "a", 1, 16, 3),
				testProposalItem(ZoneIDFinancial, "0", "b", 2, 14, 1),
			},
		},
	}, schedule.Groups)
}

func TestShardLayoutsAndRoutingTablesCommitDeterministically(t *testing.T) {
	nodeA := EmptyState()
	nodeB := EmptyState()
	var err error
	for _, zone := range []ZoneDescriptor{
		testDescriptor(ZoneIDFinancial, ZoneTypeFinancial, "financial"),
		testDescriptor(ZoneIDIdentity, ZoneTypeIdentity, "identity"),
	} {
		nodeA, err = RegisterZoneDescriptor(nodeA, zone)
		require.NoError(t, err)
	}
	for _, zone := range []ZoneDescriptor{
		testDescriptor(ZoneIDIdentity, ZoneTypeIdentity, "identity"),
		testDescriptor(ZoneIDFinancial, ZoneTypeFinancial, "financial"),
	} {
		nodeB, err = RegisterZoneDescriptor(nodeB, zone)
		require.NoError(t, err)
	}

	financial := testShardLayout(t, ZoneIDFinancial, 2, []ShardID{"1", "0"})
	identity := testShardLayout(t, ZoneIDIdentity, 1, []ShardID{"0"})
	for _, layout := range []ShardLayout{financial, identity} {
		nodeA, err = RegisterShardLayout(nodeA, layout)
		require.NoError(t, err)
	}
	for _, layout := range []ShardLayout{identity, financial} {
		nodeB, err = RegisterShardLayout(nodeB, layout)
		require.NoError(t, err)
	}

	tableA, err := BuildRoutingTableCommitment(3, 10, []ShardLayout{financial, identity})
	require.NoError(t, err)
	tableB, err := BuildRoutingTableCommitment(3, 10, []ShardLayout{identity, financial})
	require.NoError(t, err)
	require.Equal(t, tableA, tableB)

	nodeA, err = CommitRoutingTable(nodeA, tableA)
	require.NoError(t, err)
	nodeB, err = CommitRoutingTable(nodeB, tableB)
	require.NoError(t, err)
	require.Equal(t, nodeA.Export(), nodeB.Export())

	_, snapshot, err := CommitBlockRoots(nodeA, 10)
	require.NoError(t, err)
	require.True(t, hasProofRoot(snapshot, ShardLayoutRootType, financial.LayoutHash))
	require.True(t, hasProofRoot(snapshot, ShardLayoutRootType, identity.LayoutHash))
	require.True(t, hasProofRoot(snapshot, RoutingTableRootType, tableA.TableHash))
}

func TestRoutingTableRejectsLayoutMismatch(t *testing.T) {
	state := EmptyState()
	var err error
	state, err = RegisterZoneDescriptor(state, testDescriptor(ZoneIDFinancial, ZoneTypeFinancial, "financial"))
	require.NoError(t, err)
	layout := testShardLayout(t, ZoneIDFinancial, 1, []ShardID{"0"})
	state, err = RegisterShardLayout(state, layout)
	require.NoError(t, err)

	table, err := NewRoutingTableCommitment(2, 5, []RoutingZoneEntry{{
		ZoneID:		ZoneIDFinancial,
		LayoutEpoch:	layout.LayoutEpoch,
		ActiveShards:	2,
		LayoutHash:	layout.LayoutHash,
	}})
	require.NoError(t, err)
	_, err = CommitRoutingTable(state, table)
	require.ErrorContains(t, err, "active shard count mismatch")
}

func TestAetraNextCommitmentDeterministicAcrossNodes(t *testing.T) {
	nodeA := nextReadyState(t,
		[]ZoneID{ZoneIDFinancial, ZoneIDIdentity, ZoneIDApplication, ZoneIDContract},
		[]ZoneID{ZoneIDFinancial, ZoneIDIdentity, ZoneIDApplication, ZoneIDContract},
	)
	nodeB := nextReadyState(t,
		[]ZoneID{ZoneIDContract, ZoneIDApplication, ZoneIDIdentity, ZoneIDFinancial},
		[]ZoneID{ZoneIDContract, ZoneIDApplication, ZoneIDIdentity, ZoneIDFinancial},
	)

	commitmentA, err := BuildAetraNextCommitment(10, nodeA, testContributions(10), testHash("10/resolver"))
	require.NoError(t, err)
	require.NoError(t, commitmentA.ValidateHash())
	commitmentB, err := BuildAetraNextCommitment(10, nodeB, testContributions(10), testHash("10/resolver"))
	require.NoError(t, err)

	require.Equal(t, commitmentA, commitmentB)
	require.NoError(t, ValidateHash("next architecture hash", commitmentA.ArchitectureHash))
}

func TestAetraNextTopologyBootstrapMatchesArchitectureDiagram(t *testing.T) {
	plan, err := DefaultAetraNextTopology()
	require.NoError(t, err)
	require.NoError(t, plan.ValidateHash())
	require.Len(t, plan.Nodes, 9)
	require.Len(t, plan.Edges, 10)
	require.True(t, hasTopologyEdge(plan, topologyNodeCore, topologyNodeFinancial, topologyRelationSchedules))
	require.True(t, hasTopologyEdge(plan, topologyNodeCore, topologyNodeIdentity, topologyRelationSchedules))
	require.True(t, hasTopologyEdge(plan, topologyNodeCore, topologyNodeApplication, topologyRelationSchedules))
	require.True(t, hasTopologyEdge(plan, topologyNodeFinancialShards, topologyNodeContract, topologyRelationAsyncCall))
	require.True(t, hasTopologyEdge(plan, topologyNodeIdentityShards, topologyNodeContract, topologyRelationAsyncCall))
	require.True(t, hasTopologyEdge(plan, topologyNodeApplicationShards, topologyNodeContract, topologyRelationAsyncCall))
	require.True(t, hasTopologyEdge(plan, topologyNodeContract, topologyNodeContractShards, topologyRelationOwns))

	stateA, planA, err := BuildAetraNextTopologyState(TestnetParams(), 1, 1, 10)
	require.NoError(t, err)
	stateB, planB, err := BuildAetraNextTopologyState(TestnetParams(), 1, 1, 10)
	require.NoError(t, err)
	require.Equal(t, planA, planB)
	require.Equal(t, stateA, stateB)
	require.Equal(t, plan.TopologyHash, planA.TopologyHash)
	require.NoError(t, ValidateAetraNextTopologyState(stateA, 10))
	require.Len(t, stateA.Export().Zones, 4)
	require.Len(t, stateA.Export().ShardLayouts, 4)
	require.Len(t, stateA.Export().RoutingTables, 1)
}

func TestFinalityRecordCommitsCoreFinalityState(t *testing.T) {
	nodeA := nextReadyState(t,
		[]ZoneID{ZoneIDFinancial, ZoneIDIdentity, ZoneIDApplication, ZoneIDContract},
		[]ZoneID{ZoneIDFinancial, ZoneIDIdentity, ZoneIDApplication, ZoneIDContract},
	)
	nodeB := nextReadyState(t,
		[]ZoneID{ZoneIDContract, ZoneIDApplication, ZoneIDIdentity, ZoneIDFinancial},
		[]ZoneID{ZoneIDContract, ZoneIDApplication, ZoneIDIdentity, ZoneIDFinancial},
	)

	nodeA, recordA := appendTestFinalityRecord(t, nodeA, 10)
	nodeB, recordB := appendTestFinalityRecord(t, nodeB, 10)

	require.Equal(t, recordA, recordB)
	require.Equal(t, uint64(11), recordA.EligibleDeliveryHeight)
	require.Equal(t, recordA.GlobalStateRoot, nodeA.RootSnapshots[0].Finality.GlobalStateRoot)
	require.NoError(t, ValidateRootAggregationInvariants(nodeA))
	require.NoError(t, AssertReplayIdenticalRoots(nodeA, nodeB))

	stored, found := nodeA.FinalityRecordAtHeight(10)
	require.True(t, found)
	require.Equal(t, recordA, stored)
}

func TestAetraNextCommitmentRejectsMissingShardLayout(t *testing.T) {
	state := EmptyState(TestnetParams())
	var err error
	for _, zone := range DefaultAetraNextZoneDescriptors() {
		state, err = RegisterZoneDescriptor(state, zone)
		require.NoError(t, err)
	}
	service, err := DefaultAetraNextIdentityResolverService(1)
	require.NoError(t, err)
	state, err = RegisterServiceDescriptor(state, service)
	require.NoError(t, err)
	layouts, err := DefaultAetraNextShardLayouts(1)
	require.NoError(t, err)
	registeredLayouts := make([]ShardLayout, 0, len(layouts)-1)
	for _, layout := range layouts {
		if layout.ZoneID == ZoneIDContract {
			continue
		}
		state, err = RegisterShardLayout(state, layout)
		require.NoError(t, err)
		registeredLayouts = append(registeredLayouts, layout)
	}
	table, err := BuildRoutingTableCommitment(1, 10, registeredLayouts)
	require.NoError(t, err)
	state, err = CommitRoutingTable(state, table)
	require.NoError(t, err)
	for _, zone := range DefaultAetraNextZoneDescriptors() {
		state, err = AppendZoneCommitment(state, testCommitment(t, 10, zone.ZoneID))
		require.NoError(t, err)
	}

	_, err = BuildAetraNextCommitment(10, state, testContributions(10), testHash("10/resolver"))
	require.ErrorContains(t, err, "missing shard layout for zone CONTRACT_ZONE")
}

func TestAetraNextCommitmentRejectsStaleRoutingTable(t *testing.T) {
	state := nextReadyState(t,
		[]ZoneID{ZoneIDIdentity, ZoneIDFinancial, ZoneIDApplication, ZoneIDContract},
		[]ZoneID{ZoneIDIdentity, ZoneIDFinancial, ZoneIDApplication, ZoneIDContract},
	)
	financialV2 := testShardLayout(t, ZoneIDFinancial, 2, []ShardID{"0", "1", "2"})
	var err error
	state, err = RegisterShardLayout(state, financialV2)
	require.NoError(t, err)

	_, err = BuildAetraNextCommitment(10, state, testContributions(10), testHash("10/resolver"))
	require.ErrorContains(t, err, "layout epoch mismatch")
}

func TestUniversalProofRegistryRootCanonicalizesInput(t *testing.T) {
	roots := []ProofRoot{
		{Height: 8, RootType: MessageProofRootType, RootHash: testHash("messages"), Source: "aetracore.next.messages"},
		{Height: 8, RootType: AccountProofRootType, RootHash: testHash("accounts"), Source: "aetracore.next.accounts"},
		{Height: 8, RootType: ResolverProofRootType, RootHash: testHash("resolver"), Source: "aetracore.next.resolver"},
	}
	rootA, err := ComputeUniversalProofRegistryRoot(8, roots)
	require.NoError(t, err)
	rootB, err := ComputeUniversalProofRegistryRoot(8, []ProofRoot{roots[2], roots[0], roots[1]})
	require.NoError(t, err)
	require.Equal(t, rootA, rootB)
}

func TestProposalScheduleRequiresCommittedActiveShard(t *testing.T) {
	state := EmptyState()
	var err error
	state, err = RegisterZoneDescriptor(state, testDescriptor(ZoneIDFinancial, ZoneTypeFinancial, "financial"))
	require.NoError(t, err)
	layout := testShardLayout(t, ZoneIDFinancial, 1, []ShardID{"0"})
	state, err = RegisterShardLayout(state, layout)
	require.NoError(t, err)

	schedule, err := BuildProposalSchedule(9, []ProposalItem{
		testProposalItem(ZoneIDFinancial, "0", "accepted", 1, 1, 0),
	}, TestnetParams())
	require.NoError(t, err)
	require.NoError(t, ValidateProposalScheduleForState(schedule, state))

	schedule, err = BuildProposalSchedule(9, []ProposalItem{
		testProposalItem(ZoneIDFinancial, "1", "missing-shard", 1, 1, 0),
	}, TestnetParams())
	require.NoError(t, err)
	require.ErrorContains(t, ValidateProposalScheduleForState(schedule, state), "not active")
}

func TestExportImportRoundTripDeterministic(t *testing.T) {
	state := populatedState(t, []ZoneID{ZoneIDFinancial, ZoneIDContract})
	root, err := BuildGlobalStateRoot(7, state, testContributions(7))
	require.NoError(t, err)
	state, err = AppendGlobalRoot(state, root)
	require.NoError(t, err)

	imported, err := ImportState(state.Export())
	require.NoError(t, err)
	require.Equal(t, state.Export(), imported.Export())
}

func TestInvalidCoreStateRejected(t *testing.T) {
	state := EmptyState()
	bad := testDescriptor(ZoneIDFinancial, ZoneTypeFinancial, "financial")
	bad.StateMachineVersion = 0
	_, err := RegisterZoneDescriptor(state, bad)
	require.ErrorContains(t, err, "state machine version")

	_, err = AppendZoneCommitment(state, testCommitment(t, 1, ZoneIDFinancial))
	require.ErrorContains(t, err, "not registered")

	_, err = BuildProposalSchedule(1, []ProposalItem{testProposalItem("bad zone", "0", "x", 1, 1, 0)}, TestnetParams())
	require.ErrorContains(t, err, "zone id")
}

func TestMessageReceiptAndProofRootsValidate(t *testing.T) {
	messageRoot, err := NewGlobalMessageRoot(1, testHash("inbox"), testHash("outbox"), 2)
	require.NoError(t, err)
	require.NoError(t, messageRoot.ValidateHash())

	receiptRoot, err := NewExecutionReceiptRoot(1, testHash("receipts"), 2)
	require.NoError(t, err)
	require.NoError(t, receiptRoot.Validate())

	proofRoot, err := NewProofRoot(1, MessageProofRootType, messageRoot.MessageRoot, "aetracore.global_messages")
	require.NoError(t, err)
	require.NoError(t, proofRoot.Validate())
}

func populatedState(t *testing.T, order []ZoneID) AetraCoreState {
	t.Helper()
	state := EmptyState()
	var err error
	for _, zoneID := range order {
		var zoneType ZoneType
		var name string
		switch zoneID {
		case ZoneIDFinancial:
			zoneType = ZoneTypeFinancial
			name = "financial"
		case ZoneIDContract:
			zoneType = ZoneTypeContract
			name = "contract"
		default:
			t.Fatalf("unsupported zone %s", zoneID)
		}
		state, err = RegisterZoneDescriptor(state, testDescriptor(zoneID, zoneType, name))
		require.NoError(t, err)
		state, err = AppendZoneCommitment(state, testCommitment(t, 7, zoneID))
		require.NoError(t, err)
	}
	return state
}

func nextReadyState(t *testing.T, zoneOrder []ZoneID, layoutOrder []ZoneID) AetraCoreState {
	t.Helper()
	state := EmptyState(TestnetParams())
	var err error
	for _, zoneID := range zoneOrder {
		switch zoneID {
		case ZoneIDFinancial:
			state, err = RegisterZoneDescriptor(state, testDescriptor(ZoneIDFinancial, ZoneTypeFinancial, "financial"))
		case ZoneIDIdentity:
			state, err = RegisterZoneDescriptor(state, testDescriptor(ZoneIDIdentity, ZoneTypeIdentity, "identity"))
		case ZoneIDApplication:
			state, err = RegisterZoneDescriptor(state, testDescriptor(ZoneIDApplication, ZoneTypeApplication, "application"))
		case ZoneIDContract:
			state, err = RegisterZoneDescriptor(state, testDescriptor(ZoneIDContract, ZoneTypeContract, "contract"))
		default:
			t.Fatalf("unsupported next zone %s", zoneID)
		}
		require.NoError(t, err)
	}
	state, err = RegisterServiceDescriptor(state, testService("identity-resolver", ZoneIDIdentity))
	require.NoError(t, err)

	layoutsByZone := map[ZoneID]ShardLayout{
		ZoneIDFinancial:	testShardLayout(t, ZoneIDFinancial, 1, []ShardID{"1", "0"}),
		ZoneIDIdentity:		testShardLayout(t, ZoneIDIdentity, 1, []ShardID{"0"}),
		ZoneIDApplication:	testShardLayout(t, ZoneIDApplication, 1, []ShardID{"0"}),
		ZoneIDContract:		testShardLayout(t, ZoneIDContract, 1, []ShardID{"1", "0"}),
	}
	layouts := make([]ShardLayout, 0, len(layoutOrder))
	for _, zoneID := range layoutOrder {
		layout := layoutsByZone[zoneID]
		state, err = RegisterShardLayout(state, layout)
		require.NoError(t, err)
		layouts = append(layouts, layout)
	}
	table, err := BuildRoutingTableCommitment(1, 10, layouts)
	require.NoError(t, err)
	state, err = CommitRoutingTable(state, table)
	require.NoError(t, err)
	for _, zoneID := range zoneOrder {
		state, err = AppendZoneCommitment(state, testCommitment(t, 10, zoneID))
		require.NoError(t, err)
	}
	return state
}

func testDescriptor(id ZoneID, zoneType ZoneType, moduleName string) ZoneDescriptor {
	return ZoneDescriptor{
		ZoneID:			id,
		ZoneType:		zoneType,
		ModuleName:		moduleName,
		Enabled:		true,
		StateMachineVersion:	1,
		MempoolPolicyID:	DefaultMempoolPolicy,
		FeePolicyID:		NativeFeePolicyID,
		ShardLayoutEpoch:	1,
		MaxShards:		4,
		MessageCapabilities:	[]string{"async-inbox", "async-outbox"},
		ProofCapabilities:	[]string{"account", "message", "receipt"},
	}
}

func testService(serviceID string, zoneID ZoneID) ServiceDescriptor {
	interfaceID := "l1.identity.v1.Query"
	return ServiceDescriptor{
		ServiceID:		serviceID,
		Owner:			DefaultAuthority,
		ServiceType:		ServiceTypeOnChain,
		ZoneID:			zoneID,
		InterfaceID:		interfaceID,
		EndpointKey:		"identity.query",
		Version:		1,
		AvailabilityHash:	testHash(serviceID + "/availability"),
		Enabled:		true,
		Status:			ServiceStatusActive,
		ExpiryHeight:		100,
		CreatedHeight:		1,
		UpdatedHeight:		1,
		Interface: testServiceInterface(interfaceID, []ServiceMethodDescriptor{
			testServiceMethod("resolve", ServiceMethodSync, ServiceVerificationConsensusReceipt, DefaultGasPolicy, ServiceFailureRevert),
		}),
		Execution: ServiceExecutionDescriptor{
			Location:		ServiceLocationModule,
			Target:			"identity.query",
			ModuleRoute:		"identity",
			Mode:			ExecutionModeSync,
			Deterministic:		true,
			FailureBehavior:	ServiceFailureRevert,
		},
		Discovery: ServiceDiscoveryDescriptor{
			ServiceName:		serviceID,
			IdentityName:		"identity.aet",
			MetadataHash:		testHash(serviceID + "/metadata"),
			CacheExpiryHeight:	90,
			SignaturePolicy:	"owner-signature-v1",
		},
		Payment: ServicePaymentDescriptor{
			SettlementMode:	ServicePaymentOnChain,
			Denom:		NativeFeePolicyID,
			Amount:		"0",
			PricingUnit:	ServicePricingPerCall,
		},
		Storage: ServiceStorageDescriptor{
			Model:		ServiceStorageOnChain,
			StateRootType:	StateProofRootType,
			ProofRequired:	true,
		},
		Verification: ServiceVerificationDescriptor{
			TrustModel:	ServiceTrustConsensusExecuted,
			Model:		ServiceVerificationConsensusReceipt,
		},
	}
}

func testOffChainService(serviceID string, zoneID ZoneID) ServiceDescriptor {
	interfaceID := "l1.indexer.v1.Query"
	return ServiceDescriptor{
		ServiceID:		serviceID,
		Owner:			DefaultAuthority,
		ServiceType:		ServiceTypeOffChain,
		ZoneID:			zoneID,
		InterfaceID:		interfaceID,
		EndpointKey:		"indexer.query",
		Version:		1,
		AvailabilityHash:	testHash(serviceID + "/availability"),
		Enabled:		true,
		Status:			ServiceStatusActive,
		ExpiryHeight:		120,
		CreatedHeight:		1,
		UpdatedHeight:		1,
		Interface: testServiceInterface(interfaceID, []ServiceMethodDescriptor{
			testServiceMethod("query", ServiceMethodAsync, ServiceVerificationSignedResult, "", ServiceFailureRetry),
		}),
		Execution: ServiceExecutionDescriptor{
			Location:		ServiceLocationExternal,
			Target:			"indexer.query",
			Endpoint:		"https://indexer.aetra.local/v1",
			Mode:			ExecutionModeAsync,
			FailureBehavior:	ServiceFailureRetry,
			ResultExpiry:		30,
		},
		Discovery: ServiceDiscoveryDescriptor{
			ServiceName:		serviceID,
			MetadataHash:		testHash(serviceID + "/metadata"),
			CacheExpiryHeight:	100,
			SignaturePolicy:	"provider-signature-v1",
		},
		Payment: ServicePaymentDescriptor{
			SettlementMode:	ServicePaymentPrepaid,
			Denom:		NativeFeePolicyID,
			Amount:		"1",
			PricingUnit:	ServicePricingPerCall,
		},
		Storage: ServiceStorageDescriptor{
			Model: ServiceStorageDistributedOffChain,
		},
		Verification: ServiceVerificationDescriptor{
			TrustModel:			ServiceTrustFullyTrusted,
			Model:				ServiceVerificationSignedResult,
			RequestSigningRequired:		true,
			ResponseSigningRequired:	true,
		},
	}
}

func testMixedService(serviceID string, zoneID ZoneID) ServiceDescriptor {
	interfaceID := "l1.storage.v1.Mixed"
	return ServiceDescriptor{
		ServiceID:		serviceID,
		Owner:			DefaultAuthority,
		ServiceType:		ServiceTypeMixed,
		ZoneID:			zoneID,
		InterfaceID:		interfaceID,
		EndpointKey:		"storage.hybrid",
		Version:		1,
		AvailabilityHash:	testHash(serviceID + "/availability"),
		Enabled:		true,
		Status:			ServiceStatusActive,
		ExpiryHeight:		180,
		CreatedHeight:		1,
		UpdatedHeight:		1,
		Interface: testServiceInterface(interfaceID, []ServiceMethodDescriptor{
			testServiceMethod("put", ServiceMethodAsync, ServiceVerificationChallengeWindow, "", ServiceFailureChallenge),
		}),
		Execution: ServiceExecutionDescriptor{
			Location:		ServiceLocationHybrid,
			Target:			"storage.hybrid",
			Endpoint:		"https://storage.aetra.local/v1",
			Mode:			ExecutionModeAsync,
			FailureBehavior:	ServiceFailureChallenge,
			ResultExpiry:		40,
			ChallengeWindow:	20,
		},
		Discovery: ServiceDiscoveryDescriptor{
			ServiceName:		serviceID,
			MetadataHash:		testHash(serviceID + "/metadata"),
			CacheExpiryHeight:	150,
			SignaturePolicy:	"owner-and-provider-signature-v1",
		},
		Payment: ServicePaymentDescriptor{
			SettlementMode:	ServicePaymentEscrow,
			Denom:		NativeFeePolicyID,
			Amount:		"5",
			PricingUnit:	ServicePricingPerByte,
			EscrowRequired:	true,
			EscrowID:	"storage-escrow",
		},
		Storage: ServiceStorageDescriptor{
			Model:		ServiceStorageHybridCommitment,
			CommitmentHash:	testHash(serviceID + "/storage-commitment"),
			ProofRequired:	true,
		},
		Verification: ServiceVerificationDescriptor{
			TrustModel:		ServiceTrustHybridChallengeable,
			Model:			ServiceVerificationChallengeWindow,
			ChallengeWindow:	20,
			FaultPolicy:		ServiceFailureChallenge,
		},
	}
}

func testFogMarketService(serviceID string, zoneID ZoneID) ServiceDescriptor {
	interfaceID := "l1.fog.v1.Compute"
	return ServiceDescriptor{
		ServiceID:		serviceID,
		Owner:			DefaultAuthority,
		ServiceType:		ServiceTypeFogMarket,
		ZoneID:			zoneID,
		InterfaceID:		interfaceID,
		EndpointKey:		"fog.compute",
		Version:		1,
		AvailabilityHash:	testHash(serviceID + "/availability"),
		Enabled:		true,
		Status:			ServiceStatusActive,
		ExpiryHeight:		200,
		CreatedHeight:		1,
		UpdatedHeight:		1,
		Interface: testServiceInterface(interfaceID, []ServiceMethodDescriptor{
			testServiceMethod("run", ServiceMethodAsync, ServiceVerificationEconomicCollateral, "", ServiceFailureSlashProvider),
		}),
		Execution: ServiceExecutionDescriptor{
			Location:		ServiceLocationProviderPool,
			Target:			"fog.compute",
			ProviderPoolID:		"compute-pool",
			Mode:			ExecutionModeAsync,
			FailureBehavior:	ServiceFailureSlashProvider,
			ResultExpiry:		25,
		},
		Discovery: ServiceDiscoveryDescriptor{
			ServiceName:		serviceID,
			ProviderRoot:		testHash(serviceID + "/providers"),
			MetadataHash:		testHash(serviceID + "/metadata"),
			CacheExpiryHeight:	180,
			SignaturePolicy:	"provider-set-signature-v1",
		},
		Payment: ServicePaymentDescriptor{
			SettlementMode:	ServicePaymentMetered,
			Denom:		NativeFeePolicyID,
			Amount:		"2",
			PricingUnit:	ServicePricingPerComputeUnit,
			MeterID:	"compute-meter",
		},
		Storage: ServiceStorageDescriptor{
			Model: ServiceStorageEphemeral,
		},
		Verification: ServiceVerificationDescriptor{
			TrustModel:			ServiceTrustEconomicallySecured,
			Model:				ServiceVerificationEconomicCollateral,
			ProviderCollateralDenom:	NativeFeePolicyID,
			ProviderCollateralAmount:	"100",
			FaultPolicy:			ServiceFailureSlashProvider,
		},
	}
}

func testServiceInterface(interfaceID string, methods []ServiceMethodDescriptor) ServiceInterfaceDescriptor {
	descriptor := ServiceInterfaceDescriptor{
		InterfaceID:	interfaceID,
		InterfaceName:	interfaceID,
		Version:	1,
		SchemaEncoding:	"json-schema-v1",
		Methods:	methods,
		Events:		[]string{"service.receipt"},
		Errors:		[]string{"service.error"},
		AuthModel:	"aetra-account",
		PaymentModel:	"naet-fixed",
		MetadataHash:	testHash(interfaceID + "/metadata"),
		CreatedHeight:	1,
	}
	descriptor = CanonicalServiceInterfaceDescriptor(descriptor)
	descriptor.InterfaceHash = ComputeServiceInterfaceHash(descriptor)
	return descriptor
}

func testServiceMethod(methodID string, executionType ServiceMethodExecutionType, verificationModel ServiceVerificationModel, gasModel string, failureBehavior ServiceFailureBehavior) ServiceMethodDescriptor {
	return ServiceMethodDescriptor{
		MethodID:		methodID,
		Name:			methodID,
		InputSchemaHash:	testHash(methodID + "/input"),
		OutputSchemaHash:	testHash(methodID + "/output"),
		ExecutionType:		executionType,
		RequiredPaymentModel:	"naet-fixed",
		GasModel:		gasModel,
		VerificationModel:	verificationModel,
		TimeoutHeightDelta:	10,
		IdempotencyRequired:	true,
		FailureBehavior:	failureBehavior,
	}
}

func testCommitment(t *testing.T, height uint64, zoneID ZoneID) ZoneCommitment {
	t.Helper()
	commitment, err := NewZoneCommitment(
		height,
		zoneID,
		testHash(fmt.Sprintf("%d/%s/state", height, zoneID)),
		testHash(fmt.Sprintf("%d/%s/inbox", height, zoneID)),
		testHash(fmt.Sprintf("%d/%s/outbox", height, zoneID)),
		testHash(fmt.Sprintf("%d/%s/receipts", height, zoneID)),
		testHash(fmt.Sprintf("%d/%s/events", height, zoneID)),
		testHash(fmt.Sprintf("%d/%s/shards", height, zoneID)),
		testHash(fmt.Sprintf("%d/%s/params", height, zoneID)),
		testHash(fmt.Sprintf("%d/%s/execution", height, zoneID)),
	)
	require.NoError(t, err)
	return commitment
}

func testShardLayout(t *testing.T, zoneID ZoneID, epoch uint64, shardIDs []ShardID) ShardLayout {
	t.Helper()
	shards := make([]ShardDescriptor, len(shardIDs))
	for i, shardID := range shardIDs {
		shards[i] = ShardDescriptor{
			ShardID:		shardID,
			StatePrefix:		fmt.Sprintf("zone/%s/shard/%s", zoneID, shardID),
			ActivationHeight:	1,
			ValidatorSetHash:	testHash(fmt.Sprintf("%s/%s/validators", zoneID, shardID)),
			Available:		true,
		}
	}
	layout, err := NewShardLayout(zoneID, epoch, 1, testHash(fmt.Sprintf("%s/%d/routing-seed", zoneID, epoch)), shards)
	require.NoError(t, err)
	return layout
}

func testContributions(height uint64) RootContributions {
	return RootContributions{
		IdentityRoot:	testHash(fmt.Sprintf("%d/identity", height)),
		StorageRoot:	testHash(fmt.Sprintf("%d/storage", height)),
		MessageRoot:	testHash(fmt.Sprintf("%d/messages", height)),
		ReceiptsRoot:	testHash(fmt.Sprintf("%d/receipts", height)),
		RoutingRoot:	testHash(fmt.Sprintf("%d/routing", height)),
		PaymentsRoot:	testHash(fmt.Sprintf("%d/payments", height)),
		ContractsRoot:	testHash(fmt.Sprintf("%d/contracts", height)),
		VMRoot:		testHash(fmt.Sprintf("%d/vm", height)),
		ParamsHash:	testHash(fmt.Sprintf("%d/params", height)),
	}
}

func testProposalItem(zoneID ZoneID, shardID ShardID, seed string, priority uint32, height uint64, txIndex uint32) ProposalItem {
	return ProposalItem{
		ZoneID:			zoneID,
		ShardID:		shardID,
		TxHash:			testHash(seed),
		PriorityClass:		priority,
		AdmissionHeight:	height,
		TxIndex:		txIndex,
	}
}

func hasProofRoot(snapshot RootSnapshot, rootType RootType, hash string) bool {
	for _, root := range snapshot.ProofRoots {
		if root.RootType == rootType && root.RootHash == hash {
			return true
		}
	}
	return false
}

func hasTopologyEdge(plan AetraNextTopologyPlan, from string, to string, relation string) bool {
	for _, edge := range plan.Edges {
		if edge.FromNodeID == from && edge.ToNodeID == to && edge.Relation == relation {
			return true
		}
	}
	return false
}

func appendTestFinalityRecord(t *testing.T, state AetraCoreState, height uint64) (AetraCoreState, FinalityRecord) {
	t.Helper()
	next, snapshot, err := CommitBlockRootsWithContributions(state, height, testContributions(height))
	require.NoError(t, err)
	table, found := next.LatestRoutingTableAtHeight(height)
	require.True(t, found)
	record, err := NewFinalityRecord(snapshot, testHash(fmt.Sprintf("%d/apphash", height)), uint64(len(next.CommitmentsAtHeight(height))), table.TableHash, next.Params.CrossZoneFinalityDelay)
	require.NoError(t, err)
	next, err = AppendFinalityRecord(next, record)
	require.NoError(t, err)
	return next, record
}

func taskByID(t *testing.T, tasks []CoreImplementationTaskSpec, taskID CoreImplementationTaskID) CoreImplementationTaskSpec {
	t.Helper()
	for _, task := range tasks {
		if task.TaskID == taskID {
			return task
		}
	}
	t.Fatalf("missing task %s", taskID)
	return CoreImplementationTaskSpec{}
}

func testHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
