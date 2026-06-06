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
		TrustModel: ServiceTrustFullyTrusted,
		Model:      ServiceVerificationAdvisory,
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
		ZoneID:                ZoneIDApplication,
		ZoneName:              "application",
		ZoneType:              ZoneTypeApplication,
		Enabled:               true,
		StateVersion:          2,
		KeeperScope:           "application.keeper",
		MsgServerScope:        "application.msg",
		QueryServerScope:      "application.query",
		GasPolicyID:           DefaultGasPolicy,
		MessagePolicyID:       DefaultMessagePolicy,
		RootPrefix:            "zone/APPLICATION_ZONE",
		UpgradeHeightOptional: 100,
		Capabilities:          []string{"receipt", "state", "async-outbox", "async-inbox"},
		MaxShards:             4,
	}
	descriptor = CanonicalZoneDescriptor(descriptor)
	require.Equal(t, "application", descriptor.ModuleName)
	require.Equal(t, uint64(2), descriptor.StateMachineVersion)
	require.Equal(t, []string{"async-inbox", "async-outbox", "receipt", "state"}, descriptor.Capabilities)
	require.NoError(t, descriptor.Validate(TestnetParams()))

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
	require.NotEmpty(t, commitment.ParamsHash)
	require.NotEmpty(t, commitment.ExecutionSummaryHash)
	require.NoError(t, commitment.ValidateHash())

	mutated := commitment
	mutated.EventsRoot = testHash("mutated/events")
	require.ErrorContains(t, mutated.ValidateHash(), "commitment hash mismatch")
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
			ZoneID:  ZoneIDContract,
			ShardID: "1",
			Items:   []ProposalItem{testProposalItem(ZoneIDContract, "1", "a", 4, 13, 0)},
		},
		{
			ZoneID:  ZoneIDContract,
			ShardID: "2",
			Items:   []ProposalItem{testProposalItem(ZoneIDContract, "2", "c", 4, 15, 2)},
		},
		{
			ZoneID:  ZoneIDFinancial,
			ShardID: "0",
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
		ZoneID:       ZoneIDFinancial,
		LayoutEpoch:  layout.LayoutEpoch,
		ActiveShards: 2,
		LayoutHash:   layout.LayoutHash,
	}})
	require.NoError(t, err)
	_, err = CommitRoutingTable(state, table)
	require.ErrorContains(t, err, "active shard count mismatch")
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

	proofRoot, err := NewProofRoot(1, MessageProofRootType, messageRoot.MessageRoot, "aethercore.global_messages")
	require.NoError(t, err)
	require.NoError(t, proofRoot.Validate())
}

func populatedState(t *testing.T, order []ZoneID) AetherCoreState {
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

func testDescriptor(id ZoneID, zoneType ZoneType, moduleName string) ZoneDescriptor {
	return ZoneDescriptor{
		ZoneID:              id,
		ZoneType:            zoneType,
		ModuleName:          moduleName,
		Enabled:             true,
		StateMachineVersion: 1,
		MempoolPolicyID:     DefaultMempoolPolicy,
		FeePolicyID:         NativeFeePolicyID,
		ShardLayoutEpoch:    1,
		MaxShards:           4,
		MessageCapabilities: []string{"async-inbox", "async-outbox"},
		ProofCapabilities:   []string{"account", "message", "receipt"},
	}
}

func testService(serviceID string, zoneID ZoneID) ServiceDescriptor {
	interfaceID := "l1.identity.v1.Query"
	return ServiceDescriptor{
		ServiceID:        serviceID,
		Owner:            DefaultAuthority,
		ServiceType:      ServiceTypeOnChain,
		ZoneID:           zoneID,
		InterfaceID:      interfaceID,
		EndpointKey:      "identity.query",
		Version:          1,
		AvailabilityHash: testHash(serviceID + "/availability"),
		Enabled:          true,
		Status:           ServiceStatusActive,
		ExpiryHeight:     100,
		CreatedHeight:    1,
		UpdatedHeight:    1,
		Interface: testServiceInterface(interfaceID, []ServiceMethodDescriptor{
			testServiceMethod("resolve", ServiceMethodSync, ServiceVerificationConsensusReceipt, DefaultGasPolicy, ServiceFailureRevert),
		}),
		Execution: ServiceExecutionDescriptor{
			Location:        ServiceLocationModule,
			Target:          "identity.query",
			ModuleRoute:     "identity",
			Mode:            ExecutionModeSync,
			Deterministic:   true,
			FailureBehavior: ServiceFailureRevert,
		},
		Discovery: ServiceDiscoveryDescriptor{
			ServiceName:       serviceID,
			IdentityName:      "identity.aet",
			MetadataHash:      testHash(serviceID + "/metadata"),
			CacheExpiryHeight: 90,
			SignaturePolicy:   "owner-signature-v1",
		},
		Payment: ServicePaymentDescriptor{
			SettlementMode: ServicePaymentOnChain,
			Denom:          NativeFeePolicyID,
			Amount:         "0",
			PricingUnit:    ServicePricingPerCall,
		},
		Storage: ServiceStorageDescriptor{
			Model:         ServiceStorageOnChain,
			StateRootType: StateProofRootType,
			ProofRequired: true,
		},
		Verification: ServiceVerificationDescriptor{
			TrustModel: ServiceTrustConsensusExecuted,
			Model:      ServiceVerificationConsensusReceipt,
		},
	}
}

func testOffChainService(serviceID string, zoneID ZoneID) ServiceDescriptor {
	interfaceID := "l1.indexer.v1.Query"
	return ServiceDescriptor{
		ServiceID:        serviceID,
		Owner:            DefaultAuthority,
		ServiceType:      ServiceTypeOffChain,
		ZoneID:           zoneID,
		InterfaceID:      interfaceID,
		EndpointKey:      "indexer.query",
		Version:          1,
		AvailabilityHash: testHash(serviceID + "/availability"),
		Enabled:          true,
		Status:           ServiceStatusActive,
		ExpiryHeight:     120,
		CreatedHeight:    1,
		UpdatedHeight:    1,
		Interface: testServiceInterface(interfaceID, []ServiceMethodDescriptor{
			testServiceMethod("query", ServiceMethodAsync, ServiceVerificationSignedResult, "", ServiceFailureRetry),
		}),
		Execution: ServiceExecutionDescriptor{
			Location:        ServiceLocationExternal,
			Target:          "indexer.query",
			Endpoint:        "https://indexer.aetheris.local/v1",
			Mode:            ExecutionModeAsync,
			FailureBehavior: ServiceFailureRetry,
			ResultExpiry:    30,
		},
		Discovery: ServiceDiscoveryDescriptor{
			ServiceName:       serviceID,
			MetadataHash:      testHash(serviceID + "/metadata"),
			CacheExpiryHeight: 100,
			SignaturePolicy:   "provider-signature-v1",
		},
		Payment: ServicePaymentDescriptor{
			SettlementMode: ServicePaymentPrepaid,
			Denom:          NativeFeePolicyID,
			Amount:         "1",
			PricingUnit:    ServicePricingPerCall,
		},
		Storage: ServiceStorageDescriptor{
			Model: ServiceStorageDistributedOffChain,
		},
		Verification: ServiceVerificationDescriptor{
			TrustModel:              ServiceTrustFullyTrusted,
			Model:                   ServiceVerificationSignedResult,
			RequestSigningRequired:  true,
			ResponseSigningRequired: true,
		},
	}
}

func testMixedService(serviceID string, zoneID ZoneID) ServiceDescriptor {
	interfaceID := "l1.storage.v1.Mixed"
	return ServiceDescriptor{
		ServiceID:        serviceID,
		Owner:            DefaultAuthority,
		ServiceType:      ServiceTypeMixed,
		ZoneID:           zoneID,
		InterfaceID:      interfaceID,
		EndpointKey:      "storage.hybrid",
		Version:          1,
		AvailabilityHash: testHash(serviceID + "/availability"),
		Enabled:          true,
		Status:           ServiceStatusActive,
		ExpiryHeight:     180,
		CreatedHeight:    1,
		UpdatedHeight:    1,
		Interface: testServiceInterface(interfaceID, []ServiceMethodDescriptor{
			testServiceMethod("put", ServiceMethodAsync, ServiceVerificationChallengeWindow, "", ServiceFailureChallenge),
		}),
		Execution: ServiceExecutionDescriptor{
			Location:        ServiceLocationHybrid,
			Target:          "storage.hybrid",
			Endpoint:        "https://storage.aetheris.local/v1",
			Mode:            ExecutionModeAsync,
			FailureBehavior: ServiceFailureChallenge,
			ResultExpiry:    40,
			ChallengeWindow: 20,
		},
		Discovery: ServiceDiscoveryDescriptor{
			ServiceName:       serviceID,
			MetadataHash:      testHash(serviceID + "/metadata"),
			CacheExpiryHeight: 150,
			SignaturePolicy:   "owner-and-provider-signature-v1",
		},
		Payment: ServicePaymentDescriptor{
			SettlementMode: ServicePaymentEscrow,
			Denom:          NativeFeePolicyID,
			Amount:         "5",
			PricingUnit:    ServicePricingPerByte,
			EscrowRequired: true,
			EscrowID:       "storage-escrow",
		},
		Storage: ServiceStorageDescriptor{
			Model:          ServiceStorageHybridCommitment,
			CommitmentHash: testHash(serviceID + "/storage-commitment"),
			ProofRequired:  true,
		},
		Verification: ServiceVerificationDescriptor{
			TrustModel:      ServiceTrustHybridChallengeable,
			Model:           ServiceVerificationChallengeWindow,
			ChallengeWindow: 20,
			FaultPolicy:     ServiceFailureChallenge,
		},
	}
}

func testFogMarketService(serviceID string, zoneID ZoneID) ServiceDescriptor {
	interfaceID := "l1.fog.v1.Compute"
	return ServiceDescriptor{
		ServiceID:        serviceID,
		Owner:            DefaultAuthority,
		ServiceType:      ServiceTypeFogMarket,
		ZoneID:           zoneID,
		InterfaceID:      interfaceID,
		EndpointKey:      "fog.compute",
		Version:          1,
		AvailabilityHash: testHash(serviceID + "/availability"),
		Enabled:          true,
		Status:           ServiceStatusActive,
		ExpiryHeight:     200,
		CreatedHeight:    1,
		UpdatedHeight:    1,
		Interface: testServiceInterface(interfaceID, []ServiceMethodDescriptor{
			testServiceMethod("run", ServiceMethodAsync, ServiceVerificationEconomicCollateral, "", ServiceFailureSlashProvider),
		}),
		Execution: ServiceExecutionDescriptor{
			Location:        ServiceLocationProviderPool,
			Target:          "fog.compute",
			ProviderPoolID:  "compute-pool",
			Mode:            ExecutionModeAsync,
			FailureBehavior: ServiceFailureSlashProvider,
			ResultExpiry:    25,
		},
		Discovery: ServiceDiscoveryDescriptor{
			ServiceName:       serviceID,
			ProviderRoot:      testHash(serviceID + "/providers"),
			MetadataHash:      testHash(serviceID + "/metadata"),
			CacheExpiryHeight: 180,
			SignaturePolicy:   "provider-set-signature-v1",
		},
		Payment: ServicePaymentDescriptor{
			SettlementMode: ServicePaymentMetered,
			Denom:          NativeFeePolicyID,
			Amount:         "2",
			PricingUnit:    ServicePricingPerComputeUnit,
			MeterID:        "compute-meter",
		},
		Storage: ServiceStorageDescriptor{
			Model: ServiceStorageEphemeral,
		},
		Verification: ServiceVerificationDescriptor{
			TrustModel:               ServiceTrustEconomicallySecured,
			Model:                    ServiceVerificationEconomicCollateral,
			ProviderCollateralDenom:  NativeFeePolicyID,
			ProviderCollateralAmount: "100",
			FaultPolicy:              ServiceFailureSlashProvider,
		},
	}
}

func testServiceInterface(interfaceID string, methods []ServiceMethodDescriptor) ServiceInterfaceDescriptor {
	descriptor := ServiceInterfaceDescriptor{
		InterfaceID:    interfaceID,
		InterfaceName:  interfaceID,
		Version:        1,
		SchemaEncoding: "json-schema-v1",
		Methods:        methods,
		Events:         []string{"service.receipt"},
		Errors:         []string{"service.error"},
		AuthModel:      "aetheris-account",
		PaymentModel:   "naet-fixed",
		MetadataHash:   testHash(interfaceID + "/metadata"),
		CreatedHeight:  1,
	}
	descriptor = CanonicalServiceInterfaceDescriptor(descriptor)
	descriptor.InterfaceHash = ComputeServiceInterfaceHash(descriptor)
	return descriptor
}

func testServiceMethod(methodID string, executionType ServiceMethodExecutionType, verificationModel ServiceVerificationModel, gasModel string, failureBehavior ServiceFailureBehavior) ServiceMethodDescriptor {
	return ServiceMethodDescriptor{
		MethodID:             methodID,
		Name:                 methodID,
		InputSchemaHash:      testHash(methodID + "/input"),
		OutputSchemaHash:     testHash(methodID + "/output"),
		ExecutionType:        executionType,
		RequiredPaymentModel: "naet-fixed",
		GasModel:             gasModel,
		VerificationModel:    verificationModel,
		TimeoutHeightDelta:   10,
		IdempotencyRequired:  true,
		FailureBehavior:      failureBehavior,
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
			ShardID:          shardID,
			StatePrefix:      fmt.Sprintf("zone/%s/shard/%s", zoneID, shardID),
			ActivationHeight: 1,
			ValidatorSetHash: testHash(fmt.Sprintf("%s/%s/validators", zoneID, shardID)),
			Available:        true,
		}
	}
	layout, err := NewShardLayout(zoneID, epoch, 1, testHash(fmt.Sprintf("%s/%d/routing-seed", zoneID, epoch)), shards)
	require.NoError(t, err)
	return layout
}

func testContributions(height uint64) RootContributions {
	return RootContributions{
		IdentityRoot: testHash(fmt.Sprintf("%d/identity", height)),
		StorageRoot:  testHash(fmt.Sprintf("%d/storage", height)),
		MessageRoot:  testHash(fmt.Sprintf("%d/messages", height)),
		ReceiptsRoot: testHash(fmt.Sprintf("%d/receipts", height)),
		PaymentsRoot: testHash(fmt.Sprintf("%d/payments", height)),
		VMRoot:       testHash(fmt.Sprintf("%d/vm", height)),
		ParamsHash:   testHash(fmt.Sprintf("%d/params", height)),
	}
}

func testProposalItem(zoneID ZoneID, shardID ShardID, seed string, priority uint32, height uint64, txIndex uint32) ProposalItem {
	return ProposalItem{
		ZoneID:          zoneID,
		ShardID:         shardID,
		TxHash:          testHash(seed),
		PriorityClass:   priority,
		AdmissionHeight: height,
		TxIndex:         txIndex,
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

func testHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
