package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

type ServiceRoadmapPhaseID string
type ServiceRoadmapTaskID string
type ServiceRoadmapExitCriterionID string

const (
	ServiceRoadmapPhaseSpecificationCompatibility	ServiceRoadmapPhaseID	= "phase_0_specification_and_compatibility"
	ServiceRoadmapPhaseCoreRegistry			ServiceRoadmapPhaseID	= "phase_1_core_registry"
	ServiceRoadmapPhaseInterfaceSystem		ServiceRoadmapPhaseID	= "phase_2_interface_system"
	ServiceRoadmapPhaseUnifiedCallsReceipts		ServiceRoadmapPhaseID	= "phase_3_unified_calls_and_receipts"
	ServiceRoadmapPhasePayments			ServiceRoadmapPhaseID	= "phase_4_payments"
	ServiceRoadmapPhaseOffChainMixedServices	ServiceRoadmapPhaseID	= "phase_5_off_chain_and_mixed_services"
	ServiceRoadmapPhaseFogMarketProviders		ServiceRoadmapPhaseID	= "phase_6_fog_market_providers"
	ServiceRoadmapPhaseSDKUXTooling			ServiceRoadmapPhaseID	= "phase_7_sdk_and_ux_tooling"
	ServiceRoadmapPhasePerformanceHardening		ServiceRoadmapPhaseID	= "phase_8_performance_hardening"

	ServiceRoadmapTaskAddAvailabilityCommitments		ServiceRoadmapTaskID	= "add_availability_commitments"
	ServiceRoadmapTaskAddCallbacksRetries			ServiceRoadmapTaskID	= "add_callbacks_and_retries"
	ServiceRoadmapTaskAddCallEnvelopeValidation		ServiceRoadmapTaskID	= "add_call_envelope_validation"
	ServiceRoadmapTaskAddCLICommandGeneration		ServiceRoadmapTaskID	= "add_cli_command_generation_from_interface_schema"
	ServiceRoadmapTaskAddCollateralStaking			ServiceRoadmapTaskID	= "add_collateral_staking"
	ServiceRoadmapTaskAddDeterministicReceipts		ServiceRoadmapTaskID	= "add_deterministic_receipts"
	ServiceRoadmapTaskAddEscrowSettlement			ServiceRoadmapTaskID	= "add_escrow_settlement"
	ServiceRoadmapTaskAddFallbackExecutionHooks		ServiceRoadmapTaskID	= "add_fallback_execution_hooks"
	ServiceRoadmapTaskAddExportImport			ServiceRoadmapTaskID	= "add_export_and_import"
	ServiceRoadmapTaskAddIdentityBindingPlaceholder		ServiceRoadmapTaskID	= "add_identity_binding_placeholder"
	ServiceRoadmapTaskAddInterfaceCallBuilder		ServiceRoadmapTaskID	= "add_interface_driven_call_builder"
	ServiceRoadmapTaskAddInterfaceHashValidation		ServiceRoadmapTaskID	= "add_interface_hash_validation"
	ServiceRoadmapTaskAddInterfaceProofQuery		ServiceRoadmapTaskID	= "add_interface_proof_query"
	ServiceRoadmapTaskAddInterfaceRegistration		ServiceRoadmapTaskID	= "add_interface_registration"
	ServiceRoadmapTaskAddMeteredUsageReceipt		ServiceRoadmapTaskID	= "add_metered_usage_receipt"
	ServiceRoadmapTaskAddMethodSchema			ServiceRoadmapTaskID	= "add_method_schema"
	ServiceRoadmapTaskAddNameOwnerIndexes			ServiceRoadmapTaskID	= "add_service_name_and_owner_indexes"
	ServiceRoadmapTaskAddNoncesIdempotency			ServiceRoadmapTaskID	= "add_nonces_and_idempotency"
	ServiceRoadmapTaskAddPaymentModelQuery			ServiceRoadmapTaskID	= "add_payment_model_query"
	ServiceRoadmapTaskAddPerCallPayment			ServiceRoadmapTaskID	= "add_per_call_payment"
	ServiceRoadmapTaskAddProviderCollateralPenalties	ServiceRoadmapTaskID	= "add_provider_collateral_penalties"
	ServiceRoadmapTaskAddProviderRegistry			ServiceRoadmapTaskID	= "add_provider_registry"
	ServiceRoadmapTaskAddProviderSelectionQuery		ServiceRoadmapTaskID	= "add_provider_selection_query"
	ServiceRoadmapTaskAddProofVerificationHelpers		ServiceRoadmapTaskID	= "add_proof_verification_helpers"
	ServiceRoadmapTaskAddProofQuery				ServiceRoadmapTaskID	= "add_service_proof_query"
	ServiceRoadmapTaskAddReputationCommitments		ServiceRoadmapTaskID	= "add_reputation_commitments"
	ServiceRoadmapTaskAddResultAnchoring			ServiceRoadmapTaskID	= "add_result_anchoring"
	ServiceRoadmapTaskAddReceiptProofBenchmarks		ServiceRoadmapTaskID	= "add_receipt_proof_generation_benchmarks"
	ServiceRoadmapTaskAddRegistryLookupBenchmarks		ServiceRoadmapTaskID	= "add_registry_lookup_benchmarks"
	ServiceRoadmapTaskAddServiceCallThroughputTests		ServiceRoadmapTaskID	= "add_service_call_throughput_tests"
	ServiceRoadmapTaskAddServiceResolverSDK			ServiceRoadmapTaskID	= "add_service_resolver_sdk"
	ServiceRoadmapTaskAddSDKInterfaceVerifier		ServiceRoadmapTaskID	= "add_sdk_interface_verifier"
	ServiceRoadmapTaskAddServiceRegistrationUpdate		ServiceRoadmapTaskID	= "add_service_registration_and_update"
	ServiceRoadmapTaskAddSignedRequestResponseFormat	ServiceRoadmapTaskID	= "add_signed_request_and_response_format"
	ServiceRoadmapTaskAddStoreV2Benchmarks			ServiceRoadmapTaskID	= "add_store_v2_read_write_benchmarks"
	ServiceRoadmapTaskAddBlockSTMConflictBenchmarks		ServiceRoadmapTaskID	= "add_blockstm_conflict_benchmarks"
	ServiceRoadmapTaskAddMixedDisputeLoadTests		ServiceRoadmapTaskID	= "add_mixed_service_dispute_load_tests"
	ServiceRoadmapTaskAddMixedServiceChallengeFlow		ServiceRoadmapTaskID	= "add_mixed_service_challenge_flow"
	ServiceRoadmapTaskAddWalletMetadataFormat		ServiceRoadmapTaskID	= "add_wallet_metadata_format"
	ServiceRoadmapTaskDefineCallEnvelope			ServiceRoadmapTaskID	= "define_call_envelope"
	ServiceRoadmapTaskDefinePaymentModelEnum		ServiceRoadmapTaskID	= "define_payment_model_enum"
	ServiceRoadmapTaskDefineReceiptFormat			ServiceRoadmapTaskID	= "define_receipt_format"
	ServiceRoadmapTaskDefineTrustVerificationEnums		ServiceRoadmapTaskID	= "define_trust_and_verification_model_enums"
	ServiceRoadmapTaskFinalizeDescriptorSchema		ServiceRoadmapTaskID	= "finalize_service_descriptor_schema"
	ServiceRoadmapTaskFinalizeInterfaceSchema		ServiceRoadmapTaskID	= "finalize_interface_schema_format"
	ServiceRoadmapTaskImplementServiceCallsModule		ServiceRoadmapTaskID	= "implement_x_servicecalls"
	ServiceRoadmapTaskImplementServiceInterfaceModule	ServiceRoadmapTaskID	= "implement_x_serviceinterface"
	ServiceRoadmapTaskImplementServicePaymentsModule	ServiceRoadmapTaskID	= "implement_x_servicepayments"
	ServiceRoadmapTaskImplementServiceProvidersModule	ServiceRoadmapTaskID	= "implement_x_serviceproviders"
	ServiceRoadmapTaskImplementServiceReceiptsModule	ServiceRoadmapTaskID	= "implement_x_servicereceipts"
	ServiceRoadmapTaskImplementServicesModule		ServiceRoadmapTaskID	= "implement_x_services"
	ServiceRoadmapTaskIntegrateBankFinancialZone		ServiceRoadmapTaskID	= "integrate_with_bank_or_financial_zone"
	ServiceRoadmapTaskMapExistingModules			ServiceRoadmapTaskID	= "map_existing_aetra_modules_to_on_chain_services"

	ServiceRoadmapExitClientsFetchVerifyInterfaces	ServiceRoadmapExitCriterionID	= "clients_can_fetch_and_verify_formal_service_interfaces"
	ServiceRoadmapExitCallsRequireSettlePayments	ServiceRoadmapExitCriterionID	= "calls_can_require_and_settle_service_payments"
	ServiceRoadmapExitClientEndToEndFlow		ServiceRoadmapExitCriterionID	= "client_can_resolve_service_fetch_interface_build_call_attach_payment_execute_and_verify_receipt"
	ServiceRoadmapExitClientsQueryProviderSet	ServiceRoadmapExitCriterionID	= "clients_can_query_provider_set_by_service"
	ServiceRoadmapExitCoreObjectsProto		ServiceRoadmapExitCriterionID	= "all_core_objects_have_protobuf_definitions"
	ServiceRoadmapExitDescriptorProofQueryable	ServiceRoadmapExitCriterionID	= "descriptors_are_proof_queryable"
	ServiceRoadmapExitExistingDescriptors		ServiceRoadmapExitCriterionID	= "existing_modules_can_expose_service_descriptors"
	ServiceRoadmapExitInterfaceVersioning		ServiceRoadmapExitCriterionID	= "interface_versioning_is_enforced"
	ServiceRoadmapExitLookupsLowLatency		ServiceRoadmapExitCriterionID	= "registry_and_interface_lookups_remain_low_latency"
	ServiceRoadmapExitManualABINotRequired		ServiceRoadmapExitCriterionID	= "manual_abi_coding_is_not_required_for_registered_services"
	ServiceRoadmapExitMethodSchemasPublished	ServiceRoadmapExitCriterionID	= "existing_modules_can_publish_method_schemas"
	ServiceRoadmapExitMixedResultsChallenged	ServiceRoadmapExitCriterionID	= "mixed_service_results_can_be_challenged"
	ServiceRoadmapExitOffChainResultsAnchored	ServiceRoadmapExitCriterionID	= "off_chain_service_results_can_be_anchored"
	ServiceRoadmapExitOnChainUnifiedEnvelope	ServiceRoadmapExitCriterionID	= "on_chain_services_can_be_called_through_unified_call_envelope"
	ServiceRoadmapExitPaymentModelKnownBeforeSign	ServiceRoadmapExitCriterionID	= "payment_model_is_known_before_signing"
	ServiceRoadmapExitPaymentTestCoverage		ServiceRoadmapExitCriterionID	= "escrow_and_metered_usage_are_test_covered"
	ServiceRoadmapExitProviderFaultsPenalized	ServiceRoadmapExitCriterionID	= "provider_faults_can_be_penalized_deterministically"
	ServiceRoadmapExitProviderFaultCollateralFlows	ServiceRoadmapExitCriterionID	= "fault_and_collateral_flows_are_deterministic"
	ServiceRoadmapExitProvidersAdvertiseStake	ServiceRoadmapExitCriterionID	= "providers_can_advertise_services_with_stake_and_interface_support"
	ServiceRoadmapExitProofReceiptBounded		ServiceRoadmapExitCriterionID	= "proof_and_receipt_generation_remain_bounded"
	ServiceRoadmapExitReceiptsCommittedProof	ServiceRoadmapExitCriterionID	= "receipts_are_committed_and_proof_queryable"
	ServiceRoadmapExitServicesParallelizeSafely	ServiceRoadmapExitCriterionID	= "independent_services_and_calls_parallelize_safely"
	ServiceRoadmapExitRegistryReproducible		ServiceRoadmapExitCriterionID	= "registry_state_is_reproducible"
	ServiceRoadmapExitReplayAttemptsRejected	ServiceRoadmapExitCriterionID	= "replay_attempts_are_rejected"
	ServiceRoadmapExitServiceDiscovery		ServiceRoadmapExitCriterionID	= "services_are_discoverable_by_id_owner_and_name"
	ServiceRoadmapExitSignableVectors		ServiceRoadmapExitCriterionID	= "all_signable_objects_have_canonical_encoding_test_vectors"
)

type ServiceRoadmapTask struct {
	TaskID		ServiceRoadmapTaskID
	Module		string
	Artifact	string
	TaskHash	string
}

type ServiceRoadmapExitCriterion struct {
	CriterionID	ServiceRoadmapExitCriterionID
	Evidence	string
	Met		bool
	ExitHash	string
}

type ServiceCoreObjectDefinition struct {
	ObjectName	string
	ProtobufType	string
	CanonicalHash	string
	DefinitionHash	string
}

type ServiceSignableObjectVector struct {
	ObjectName		string
	CanonicalEncoding	string
	TestVectorHash		string
	VectorHash		string
}

type AetraModuleServiceMapping struct {
	ModuleName	string
	ModulePath	string
	ServiceID	string
	InterfaceHash	string
	DescriptorHash	string
	OnChain		bool
	MappingHash	string
}

type ServiceRoadmapPhase struct {
	PhaseID		ServiceRoadmapPhaseID
	Title		string
	Tasks		[]ServiceRoadmapTask
	ExitCriteria	[]ServiceRoadmapExitCriterion
	CoreObjects	[]ServiceCoreObjectDefinition
	SignableVectors	[]ServiceSignableObjectVector
	ModuleMappings	[]AetraModuleServiceMapping
	DependsOn	[]ServiceRoadmapPhaseID
	PhaseHash	string
}

type ServiceImplementationRoadmap struct {
	Phases		[]ServiceRoadmapPhase
	RoadmapHash	string
}

func DefaultServiceImplementationRoadmap() (ServiceImplementationRoadmap, error) {
	phase0, err := NewServiceRoadmapPhase(ServiceRoadmapPhase{
		PhaseID:	ServiceRoadmapPhaseSpecificationCompatibility,
		Title:		"Specification and Compatibility",
		Tasks: []ServiceRoadmapTask{
			newServiceRoadmapTask(ServiceRoadmapTaskDefineCallEnvelope, ServiceModuleCalls, "UnifiedServiceCall"),
			newServiceRoadmapTask(ServiceRoadmapTaskDefinePaymentModelEnum, ServiceModulePayments, "ServicePaymentSettlementMode"),
			newServiceRoadmapTask(ServiceRoadmapTaskDefineReceiptFormat, ServiceModuleReceipts, "ServiceReceipt"),
			newServiceRoadmapTask(ServiceRoadmapTaskDefineTrustVerificationEnums, ServiceModuleServices, "ServiceTrustModel/ServiceVerificationModel"),
			newServiceRoadmapTask(ServiceRoadmapTaskFinalizeDescriptorSchema, ServiceModuleServices, "ServiceDescriptor"),
			newServiceRoadmapTask(ServiceRoadmapTaskFinalizeInterfaceSchema, ServiceModuleInterface, "ServiceInterface"),
			newServiceRoadmapTask(ServiceRoadmapTaskMapExistingModules, ServiceModuleServices, "AetraModuleServiceMapping"),
		},
		ExitCriteria: []ServiceRoadmapExitCriterion{
			newServiceRoadmapExitCriterion(ServiceRoadmapExitCoreObjectsProto, "ServiceCoreObjectDefinition", true),
			newServiceRoadmapExitCriterion(ServiceRoadmapExitExistingDescriptors, "AetraModuleServiceMapping", true),
			newServiceRoadmapExitCriterion(ServiceRoadmapExitSignableVectors, "ServiceSignableObjectVector", true),
		},
		CoreObjects:		defaultServiceCoreObjectDefinitions(),
		SignableVectors:	defaultServiceSignableObjectVectors(),
		ModuleMappings:		defaultAetraModuleServiceMappings(),
	})
	if err != nil {
		return ServiceImplementationRoadmap{}, err
	}
	phase1, err := NewServiceRoadmapPhase(ServiceRoadmapPhase{
		PhaseID:	ServiceRoadmapPhaseCoreRegistry,
		Title:		"Core Registry",
		Tasks: []ServiceRoadmapTask{
			newServiceRoadmapTask(ServiceRoadmapTaskAddExportImport, ServiceModuleServices, "ServiceRegistryState"),
			newServiceRoadmapTask(ServiceRoadmapTaskAddIdentityBindingPlaceholder, ServiceModuleServices, "IdentityServiceBinding"),
			newServiceRoadmapTask(ServiceRoadmapTaskAddNameOwnerIndexes, ServiceModuleServices, "ServiceRegistryStateEntry"),
			newServiceRoadmapTask(ServiceRoadmapTaskAddProofQuery, ServiceModuleServices, "ServiceRegistryProof"),
			newServiceRoadmapTask(ServiceRoadmapTaskAddServiceRegistrationUpdate, ServiceModuleServices, "MsgRegisterService/MsgUpdateService"),
			newServiceRoadmapTask(ServiceRoadmapTaskImplementServicesModule, ServiceModuleServices, "XServicesModuleBreakdown"),
		},
		ExitCriteria: []ServiceRoadmapExitCriterion{
			newServiceRoadmapExitCriterion(ServiceRoadmapExitDescriptorProofQueryable, "QueryServiceProof", true),
			newServiceRoadmapExitCriterion(ServiceRoadmapExitRegistryReproducible, "ServiceRegistryState.StateRoot", true),
			newServiceRoadmapExitCriterion(ServiceRoadmapExitServiceDiscovery, "QueryService/QueryServicesByOwner/QueryServiceByName", true),
		},
		DependsOn:	[]ServiceRoadmapPhaseID{ServiceRoadmapPhaseSpecificationCompatibility},
	})
	if err != nil {
		return ServiceImplementationRoadmap{}, err
	}
	phase2, err := NewServiceRoadmapPhase(ServiceRoadmapPhase{
		PhaseID:	ServiceRoadmapPhaseInterfaceSystem,
		Title:		"Interface System",
		Tasks: []ServiceRoadmapTask{
			newServiceRoadmapTask(ServiceRoadmapTaskAddInterfaceHashValidation, ServiceModuleInterface, "ComputeFormalServiceInterfaceHash"),
			newServiceRoadmapTask(ServiceRoadmapTaskAddInterfaceProofQuery, ServiceModuleInterface, "QueryInterfaceProof"),
			newServiceRoadmapTask(ServiceRoadmapTaskAddInterfaceRegistration, ServiceModuleInterface, "MsgRegisterInterface"),
			newServiceRoadmapTask(ServiceRoadmapTaskAddMethodSchema, ServiceModuleInterface, "ServiceInterfaceMethodSchema"),
			newServiceRoadmapTask(ServiceRoadmapTaskAddSDKInterfaceVerifier, ServiceModuleInterface, "SDKInterfaceVerifier"),
			newServiceRoadmapTask(ServiceRoadmapTaskImplementServiceInterfaceModule, ServiceModuleInterface, "XServiceInterfaceModuleBreakdown"),
		},
		ExitCriteria: []ServiceRoadmapExitCriterion{
			newServiceRoadmapExitCriterion(ServiceRoadmapExitClientsFetchVerifyInterfaces, "ServiceInterfaceVerifier", true),
			newServiceRoadmapExitCriterion(ServiceRoadmapExitInterfaceVersioning, "InterfaceVersion", true),
			newServiceRoadmapExitCriterion(ServiceRoadmapExitMethodSchemasPublished, "ServiceInterfaceMethodSchema", true),
		},
		DependsOn:	[]ServiceRoadmapPhaseID{ServiceRoadmapPhaseCoreRegistry},
	})
	if err != nil {
		return ServiceImplementationRoadmap{}, err
	}
	phase3, err := NewServiceRoadmapPhase(ServiceRoadmapPhase{
		PhaseID:	ServiceRoadmapPhaseUnifiedCallsReceipts,
		Title:		"Unified Calls and Receipts",
		Tasks: []ServiceRoadmapTask{
			newServiceRoadmapTask(ServiceRoadmapTaskAddCallbacksRetries, ServiceModuleCalls, "UnifiedServiceCallback/ServiceRetryPolicy"),
			newServiceRoadmapTask(ServiceRoadmapTaskAddCallEnvelopeValidation, ServiceModuleCalls, "ValidateUnifiedServiceCallForDescriptor"),
			newServiceRoadmapTask(ServiceRoadmapTaskAddDeterministicReceipts, ServiceModuleReceipts, "ReceiptRoot"),
			newServiceRoadmapTask(ServiceRoadmapTaskAddNoncesIdempotency, ServiceModuleCalls, "ServiceCallReplayIndex"),
			newServiceRoadmapTask(ServiceRoadmapTaskImplementServiceCallsModule, ServiceModuleCalls, "XServiceCallsModuleBreakdown"),
			newServiceRoadmapTask(ServiceRoadmapTaskImplementServiceReceiptsModule, ServiceModuleReceipts, "XServiceReceiptsModuleBreakdown"),
		},
		ExitCriteria: []ServiceRoadmapExitCriterion{
			newServiceRoadmapExitCriterion(ServiceRoadmapExitOnChainUnifiedEnvelope, "ValidateUnifiedServiceCallForDescriptor", true),
			newServiceRoadmapExitCriterion(ServiceRoadmapExitReceiptsCommittedProof, "QueryReceiptProof", true),
			newServiceRoadmapExitCriterion(ServiceRoadmapExitReplayAttemptsRejected, "ServiceCallReplayIndex", true),
		},
		DependsOn:	[]ServiceRoadmapPhaseID{ServiceRoadmapPhaseInterfaceSystem},
	})
	if err != nil {
		return ServiceImplementationRoadmap{}, err
	}
	phase4, err := NewServiceRoadmapPhase(ServiceRoadmapPhase{
		PhaseID:	ServiceRoadmapPhasePayments,
		Title:		"Payments",
		Tasks: []ServiceRoadmapTask{
			newServiceRoadmapTask(ServiceRoadmapTaskAddEscrowSettlement, ServiceModulePayments, "ServiceEscrow/PaymentSettlement"),
			newServiceRoadmapTask(ServiceRoadmapTaskAddMeteredUsageReceipt, ServiceModulePayments, "MeteredUsage"),
			newServiceRoadmapTask(ServiceRoadmapTaskAddPaymentModelQuery, ServiceModulePayments, "QueryPaymentModel"),
			newServiceRoadmapTask(ServiceRoadmapTaskAddPerCallPayment, ServiceModulePayments, "QuotePerCallPayment"),
			newServiceRoadmapTask(ServiceRoadmapTaskImplementServicePaymentsModule, ServiceModulePayments, "XServicePaymentsModuleBreakdown"),
			newServiceRoadmapTask(ServiceRoadmapTaskIntegrateBankFinancialZone, ServiceModulePayments, "BuildFinancialZonePaymentRoute"),
		},
		ExitCriteria: []ServiceRoadmapExitCriterion{
			newServiceRoadmapExitCriterion(ServiceRoadmapExitCallsRequireSettlePayments, "PaymentSettlement", true),
			newServiceRoadmapExitCriterion(ServiceRoadmapExitPaymentModelKnownBeforeSign, "ServicePaymentSignedModelSnapshot", true),
			newServiceRoadmapExitCriterion(ServiceRoadmapExitPaymentTestCoverage, "servicepayments_module_breakdown_test", true),
		},
		DependsOn:	[]ServiceRoadmapPhaseID{ServiceRoadmapPhaseUnifiedCallsReceipts},
	})
	if err != nil {
		return ServiceImplementationRoadmap{}, err
	}
	phase5, err := NewServiceRoadmapPhase(ServiceRoadmapPhase{
		PhaseID:	ServiceRoadmapPhaseOffChainMixedServices,
		Title:		"Off Chain and Mixed Services",
		Tasks: []ServiceRoadmapTask{
			newServiceRoadmapTask(ServiceRoadmapTaskAddFallbackExecutionHooks, ServiceModuleCalls, "ServiceChallengeFlow"),
			newServiceRoadmapTask(ServiceRoadmapTaskAddMixedServiceChallengeFlow, ServiceModuleCalls, "NewServiceChallengeFlow"),
			newServiceRoadmapTask(ServiceRoadmapTaskAddProviderCollateralPenalties, ServiceModuleProviders, "ProviderPenaltyRoute"),
			newServiceRoadmapTask(ServiceRoadmapTaskAddResultAnchoring, ServiceModuleReceipts, "MsgAnchorReceipt"),
			newServiceRoadmapTask(ServiceRoadmapTaskAddSignedRequestResponseFormat, ServiceModuleCalls, "ServiceCallEnvelope/ServiceReceipt"),
		},
		ExitCriteria: []ServiceRoadmapExitCriterion{
			newServiceRoadmapExitCriterion(ServiceRoadmapExitMixedResultsChallenged, "NewServiceChallengeFlow", true),
			newServiceRoadmapExitCriterion(ServiceRoadmapExitOffChainResultsAnchored, "MsgAnchorReceipt", true),
			newServiceRoadmapExitCriterion(ServiceRoadmapExitProviderFaultsPenalized, "ProviderPenaltyRoute", true),
		},
		DependsOn:	[]ServiceRoadmapPhaseID{ServiceRoadmapPhasePayments},
	})
	if err != nil {
		return ServiceImplementationRoadmap{}, err
	}
	phase6, err := NewServiceRoadmapPhase(ServiceRoadmapPhase{
		PhaseID:	ServiceRoadmapPhaseFogMarketProviders,
		Title:		"Fog Market Providers",
		Tasks: []ServiceRoadmapTask{
			newServiceRoadmapTask(ServiceRoadmapTaskAddAvailabilityCommitments, ServiceModuleProviders, "AvailabilityCommitment"),
			newServiceRoadmapTask(ServiceRoadmapTaskAddCollateralStaking, ServiceModuleProviders, "ProviderCollateral"),
			newServiceRoadmapTask(ServiceRoadmapTaskAddProviderRegistry, ServiceModuleProviders, "ProviderRecord"),
			newServiceRoadmapTask(ServiceRoadmapTaskAddProviderSelectionQuery, ServiceModuleProviders, "QueryProvidersByService"),
			newServiceRoadmapTask(ServiceRoadmapTaskAddReputationCommitments, ServiceModuleProviders, "ReputationRecord"),
			newServiceRoadmapTask(ServiceRoadmapTaskImplementServiceProvidersModule, ServiceModuleProviders, "XServiceProvidersModuleBreakdown"),
		},
		ExitCriteria: []ServiceRoadmapExitCriterion{
			newServiceRoadmapExitCriterion(ServiceRoadmapExitClientsQueryProviderSet, "QueryProvidersByService", true),
			newServiceRoadmapExitCriterion(ServiceRoadmapExitProviderFaultCollateralFlows, "ProviderCollateral/ProviderFault", true),
			newServiceRoadmapExitCriterion(ServiceRoadmapExitProvidersAdvertiseStake, "ValidateProviderAdvertisesInterface", true),
		},
		DependsOn:	[]ServiceRoadmapPhaseID{ServiceRoadmapPhaseOffChainMixedServices},
	})
	if err != nil {
		return ServiceImplementationRoadmap{}, err
	}
	phase7, err := NewServiceRoadmapPhase(ServiceRoadmapPhase{
		PhaseID:	ServiceRoadmapPhaseSDKUXTooling,
		Title:		"SDK and UX Tooling",
		Tasks: []ServiceRoadmapTask{
			newServiceRoadmapTask(ServiceRoadmapTaskAddCLICommandGeneration, ServiceModuleInterface, "CLIInterfaceCommandGenerator"),
			newServiceRoadmapTask(ServiceRoadmapTaskAddInterfaceCallBuilder, ServiceModuleCalls, "MethodLevelCallBuilder"),
			newServiceRoadmapTask(ServiceRoadmapTaskAddProofVerificationHelpers, ServiceModuleReceipts, "ServiceReceiptProofRecord"),
			newServiceRoadmapTask(ServiceRoadmapTaskAddServiceResolverSDK, ServiceModuleServices, "ServiceResolver"),
			newServiceRoadmapTask(ServiceRoadmapTaskAddWalletMetadataFormat, ServiceModuleInterface, "WalletMetadataFormat"),
		},
		ExitCriteria: []ServiceRoadmapExitCriterion{
			newServiceRoadmapExitCriterion(ServiceRoadmapExitClientEndToEndFlow, "ServiceResolver/MethodLevelCallBuilder/ServiceReceiptProofRecord", true),
			newServiceRoadmapExitCriterion(ServiceRoadmapExitManualABINotRequired, "CLIInterfaceCommandGenerator", true),
		},
		DependsOn:	[]ServiceRoadmapPhaseID{ServiceRoadmapPhaseFogMarketProviders},
	})
	if err != nil {
		return ServiceImplementationRoadmap{}, err
	}
	phase8, err := NewServiceRoadmapPhase(ServiceRoadmapPhase{
		PhaseID:	ServiceRoadmapPhasePerformanceHardening,
		Title:		"Performance Hardening",
		Tasks: []ServiceRoadmapTask{
			newServiceRoadmapTask(ServiceRoadmapTaskAddBlockSTMConflictBenchmarks, ServiceModuleCalls, "ServiceBlockSTMOperation"),
			newServiceRoadmapTask(ServiceRoadmapTaskAddMixedDisputeLoadTests, ServiceModuleCalls, "ServiceChallengeFlow"),
			newServiceRoadmapTask(ServiceRoadmapTaskAddReceiptProofBenchmarks, ServiceModuleReceipts, "QueryReceiptProof"),
			newServiceRoadmapTask(ServiceRoadmapTaskAddRegistryLookupBenchmarks, ServiceModuleServices, "QueryService/QueryServiceInterface"),
			newServiceRoadmapTask(ServiceRoadmapTaskAddServiceCallThroughputTests, ServiceModuleCalls, "UnifiedServiceCall"),
			newServiceRoadmapTask(ServiceRoadmapTaskAddStoreV2Benchmarks, ServiceModuleServices, "ServiceStoreV2Layout"),
		},
		ExitCriteria: []ServiceRoadmapExitCriterion{
			newServiceRoadmapExitCriterion(ServiceRoadmapExitLookupsLowLatency, "registry_interface_lookup_benchmarks", true),
			newServiceRoadmapExitCriterion(ServiceRoadmapExitProofReceiptBounded, "receipt_proof_generation_benchmarks", true),
			newServiceRoadmapExitCriterion(ServiceRoadmapExitServicesParallelizeSafely, "blockstm_conflict_benchmarks", true),
		},
		DependsOn:	[]ServiceRoadmapPhaseID{ServiceRoadmapPhaseSDKUXTooling},
	})
	if err != nil {
		return ServiceImplementationRoadmap{}, err
	}
	return NewServiceImplementationRoadmap([]ServiceRoadmapPhase{phase0, phase1, phase2, phase3, phase4, phase5, phase6, phase7, phase8})
}

func NewServiceImplementationRoadmap(phases []ServiceRoadmapPhase) (ServiceImplementationRoadmap, error) {
	roadmap := ServiceImplementationRoadmap{Phases: cloneServiceRoadmapPhases(phases)}
	sortServiceRoadmapPhases(roadmap.Phases)
	if err := roadmap.ValidateFormat(); err != nil {
		return ServiceImplementationRoadmap{}, err
	}
	roadmap.RoadmapHash = ComputeServiceImplementationRoadmapHash(roadmap)
	return roadmap, roadmap.Validate()
}

func NewServiceRoadmapPhase(phase ServiceRoadmapPhase) (ServiceRoadmapPhase, error) {
	phase = canonicalServiceRoadmapPhase(phase)
	if err := phase.ValidateFormat(); err != nil {
		return ServiceRoadmapPhase{}, err
	}
	phase.PhaseHash = ComputeServiceRoadmapPhaseHash(phase)
	return phase, phase.Validate()
}

func ValidateServiceRoadmapExitCriteria(phase ServiceRoadmapPhase) error {
	if err := phase.Validate(); err != nil {
		return err
	}
	for _, criterion := range phase.ExitCriteria {
		if !criterion.Met {
			return fmt.Errorf("services roadmap exit criterion %s is not met", criterion.CriterionID)
		}
	}
	switch phase.PhaseID {
	case ServiceRoadmapPhaseSpecificationCompatibility:
		if err := validateServiceCoreObjectDefinitions(phase.CoreObjects); err != nil {
			return err
		}
		if err := validateServiceSignableObjectVectors(phase.SignableVectors); err != nil {
			return err
		}
		return ValidateAetraModuleServiceMappings(phase.ModuleMappings)
	case ServiceRoadmapPhaseCoreRegistry:
		return phase.requireTasks(ServiceRoadmapTaskImplementServicesModule, ServiceRoadmapTaskAddServiceRegistrationUpdate, ServiceRoadmapTaskAddNameOwnerIndexes, ServiceRoadmapTaskAddProofQuery, ServiceRoadmapTaskAddExportImport)
	case ServiceRoadmapPhaseInterfaceSystem:
		return phase.requireTasks(ServiceRoadmapTaskImplementServiceInterfaceModule, ServiceRoadmapTaskAddInterfaceRegistration, ServiceRoadmapTaskAddMethodSchema, ServiceRoadmapTaskAddInterfaceHashValidation, ServiceRoadmapTaskAddInterfaceProofQuery, ServiceRoadmapTaskAddSDKInterfaceVerifier)
	case ServiceRoadmapPhaseUnifiedCallsReceipts:
		return phase.requireTasks(ServiceRoadmapTaskImplementServiceCallsModule, ServiceRoadmapTaskImplementServiceReceiptsModule, ServiceRoadmapTaskAddCallEnvelopeValidation, ServiceRoadmapTaskAddNoncesIdempotency, ServiceRoadmapTaskAddCallbacksRetries, ServiceRoadmapTaskAddDeterministicReceipts)
	case ServiceRoadmapPhasePayments:
		return phase.requireTasks(ServiceRoadmapTaskImplementServicePaymentsModule, ServiceRoadmapTaskAddPerCallPayment, ServiceRoadmapTaskAddEscrowSettlement, ServiceRoadmapTaskAddMeteredUsageReceipt, ServiceRoadmapTaskAddPaymentModelQuery, ServiceRoadmapTaskIntegrateBankFinancialZone)
	case ServiceRoadmapPhaseOffChainMixedServices:
		return phase.requireTasks(ServiceRoadmapTaskAddSignedRequestResponseFormat, ServiceRoadmapTaskAddResultAnchoring, ServiceRoadmapTaskAddMixedServiceChallengeFlow, ServiceRoadmapTaskAddFallbackExecutionHooks, ServiceRoadmapTaskAddProviderCollateralPenalties)
	case ServiceRoadmapPhaseFogMarketProviders:
		return phase.requireTasks(ServiceRoadmapTaskImplementServiceProvidersModule, ServiceRoadmapTaskAddProviderRegistry, ServiceRoadmapTaskAddCollateralStaking, ServiceRoadmapTaskAddAvailabilityCommitments, ServiceRoadmapTaskAddReputationCommitments, ServiceRoadmapTaskAddProviderSelectionQuery)
	case ServiceRoadmapPhaseSDKUXTooling:
		return phase.requireTasks(ServiceRoadmapTaskAddServiceResolverSDK, ServiceRoadmapTaskAddInterfaceCallBuilder, ServiceRoadmapTaskAddCLICommandGeneration, ServiceRoadmapTaskAddWalletMetadataFormat, ServiceRoadmapTaskAddProofVerificationHelpers)
	case ServiceRoadmapPhasePerformanceHardening:
		return phase.requireTasks(ServiceRoadmapTaskAddBlockSTMConflictBenchmarks, ServiceRoadmapTaskAddStoreV2Benchmarks, ServiceRoadmapTaskAddServiceCallThroughputTests, ServiceRoadmapTaskAddReceiptProofBenchmarks, ServiceRoadmapTaskAddRegistryLookupBenchmarks, ServiceRoadmapTaskAddMixedDisputeLoadTests)
	default:
		return fmt.Errorf("services roadmap unknown phase %q", phase.PhaseID)
	}
}

func ValidateAetraModuleServiceMappings(mappings []AetraModuleServiceMapping) error {
	if len(mappings) == 0 {
		return errors.New("services roadmap module service mappings are required")
	}
	seen := map[string]struct{}{}
	previous := ""
	ordered := cloneAetraModuleServiceMappings(mappings)
	sortAetraModuleServiceMappings(ordered)
	for _, mapping := range ordered {
		if err := mapping.Validate(); err != nil {
			return err
		}
		if !mapping.OnChain {
			return errors.New("services roadmap existing module mappings must expose on-chain services")
		}
		if previous != "" && previous >= mapping.ModuleName {
			return errors.New("services roadmap module mappings must be sorted canonically")
		}
		previous = mapping.ModuleName
		if _, found := seen[mapping.ServiceID]; found {
			return fmt.Errorf("services roadmap duplicate service mapping %s", mapping.ServiceID)
		}
		seen[mapping.ServiceID] = struct{}{}
	}
	return nil
}

func (roadmap ServiceImplementationRoadmap) PhaseByID(phaseID ServiceRoadmapPhaseID) (ServiceRoadmapPhase, bool) {
	for _, phase := range roadmap.Phases {
		if phase.PhaseID == phaseID {
			return phase, true
		}
	}
	return ServiceRoadmapPhase{}, false
}

func (roadmap ServiceImplementationRoadmap) ReadyForPhase(phaseID ServiceRoadmapPhaseID) error {
	if err := roadmap.Validate(); err != nil {
		return err
	}
	phase, found := roadmap.PhaseByID(phaseID)
	if !found {
		return fmt.Errorf("services roadmap phase %s not found", phaseID)
	}
	for _, dependencyID := range phase.DependsOn {
		dependency, found := roadmap.PhaseByID(dependencyID)
		if !found {
			return fmt.Errorf("services roadmap dependency %s not found", dependencyID)
		}
		if err := ValidateServiceRoadmapExitCriteria(dependency); err != nil {
			return err
		}
	}
	return nil
}

func (roadmap ServiceImplementationRoadmap) ValidateFormat() error {
	if err := validateServiceRoadmapPhases(roadmap.Phases); err != nil {
		return err
	}
	if roadmap.RoadmapHash != "" {
		return coretypes.ValidateHash("services roadmap hash", roadmap.RoadmapHash)
	}
	return nil
}

func (roadmap ServiceImplementationRoadmap) Validate() error {
	if err := roadmap.ValidateFormat(); err != nil {
		return err
	}
	if roadmap.RoadmapHash == "" {
		return errors.New("services roadmap hash is required")
	}
	if expected := ComputeServiceImplementationRoadmapHash(roadmap); roadmap.RoadmapHash != expected {
		return fmt.Errorf("services roadmap hash mismatch: expected %s", expected)
	}
	return nil
}

func (phase ServiceRoadmapPhase) ValidateFormat() error {
	if !IsServiceRoadmapPhaseID(phase.PhaseID) {
		return fmt.Errorf("services roadmap unknown phase %q", phase.PhaseID)
	}
	if err := validateRoadmapText("services roadmap title", phase.Title); err != nil {
		return err
	}
	if err := validateServiceRoadmapTasks(phase.Tasks); err != nil {
		return err
	}
	if err := validateServiceRoadmapExitCriteria(phase.ExitCriteria); err != nil {
		return err
	}
	if err := validateServiceRoadmapDependencies(phase.PhaseID, phase.DependsOn); err != nil {
		return err
	}
	switch phase.PhaseID {
	case ServiceRoadmapPhaseSpecificationCompatibility:
		if err := validateServiceCoreObjectDefinitions(phase.CoreObjects); err != nil {
			return err
		}
		if err := validateServiceSignableObjectVectors(phase.SignableVectors); err != nil {
			return err
		}
		if err := ValidateAetraModuleServiceMappings(phase.ModuleMappings); err != nil {
			return err
		}
	default:
		if len(phase.CoreObjects) != 0 || len(phase.SignableVectors) != 0 || len(phase.ModuleMappings) != 0 {
			return errors.New("services roadmap compatibility artifacts belong to phase 0")
		}
	}
	if phase.PhaseHash != "" {
		return coretypes.ValidateHash("services roadmap phase hash", phase.PhaseHash)
	}
	return nil
}

func (phase ServiceRoadmapPhase) Validate() error {
	if err := phase.ValidateFormat(); err != nil {
		return err
	}
	if phase.PhaseHash == "" {
		return errors.New("services roadmap phase hash is required")
	}
	if expected := ComputeServiceRoadmapPhaseHash(phase); phase.PhaseHash != expected {
		return fmt.Errorf("services roadmap phase hash mismatch: expected %s", expected)
	}
	return nil
}

func (phase ServiceRoadmapPhase) requireTasks(required ...ServiceRoadmapTaskID) error {
	seen := map[ServiceRoadmapTaskID]struct{}{}
	for _, task := range phase.Tasks {
		seen[task.TaskID] = struct{}{}
	}
	for _, taskID := range required {
		if _, found := seen[taskID]; !found {
			return fmt.Errorf("services roadmap phase %s missing task %s", phase.PhaseID, taskID)
		}
	}
	return nil
}

func (task ServiceRoadmapTask) Validate() error {
	if !IsServiceRoadmapTaskID(task.TaskID) {
		return fmt.Errorf("services roadmap unknown task %q", task.TaskID)
	}
	if err := validateInterfaceToken("services roadmap task module", task.Module); err != nil {
		return err
	}
	if err := validateRoadmapText("services roadmap task artifact", task.Artifact); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services roadmap task hash", task.TaskHash); err != nil {
		return err
	}
	if expected := ComputeServiceRoadmapTaskHash(task); task.TaskHash != expected {
		return fmt.Errorf("services roadmap task hash mismatch: expected %s", expected)
	}
	return nil
}

func (criterion ServiceRoadmapExitCriterion) Validate() error {
	if !IsServiceRoadmapExitCriterionID(criterion.CriterionID) {
		return fmt.Errorf("services roadmap unknown exit criterion %q", criterion.CriterionID)
	}
	if err := validateRoadmapText("services roadmap exit evidence", criterion.Evidence); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services roadmap exit hash", criterion.ExitHash); err != nil {
		return err
	}
	if expected := ComputeServiceRoadmapExitCriterionHash(criterion); criterion.ExitHash != expected {
		return fmt.Errorf("services roadmap exit hash mismatch: expected %s", expected)
	}
	return nil
}

func (definition ServiceCoreObjectDefinition) Validate() error {
	if err := validateInterfaceToken("services roadmap core object", definition.ObjectName); err != nil {
		return err
	}
	if err := validateInterfaceToken("services roadmap protobuf type", definition.ProtobufType); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services roadmap canonical hash", definition.CanonicalHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services roadmap definition hash", definition.DefinitionHash); err != nil {
		return err
	}
	if expected := ComputeServiceCoreObjectDefinitionHash(definition); definition.DefinitionHash != expected {
		return fmt.Errorf("services roadmap definition hash mismatch: expected %s", expected)
	}
	return nil
}

func (vector ServiceSignableObjectVector) Validate() error {
	if err := validateInterfaceToken("services roadmap signable object", vector.ObjectName); err != nil {
		return err
	}
	if err := validateInterfaceToken("services roadmap canonical encoding", vector.CanonicalEncoding); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services roadmap test vector hash", vector.TestVectorHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services roadmap vector hash", vector.VectorHash); err != nil {
		return err
	}
	if expected := ComputeServiceSignableObjectVectorHash(vector); vector.VectorHash != expected {
		return fmt.Errorf("services roadmap vector hash mismatch: expected %s", expected)
	}
	return nil
}

func (mapping AetraModuleServiceMapping) Validate() error {
	if err := validateInterfaceToken("services roadmap module name", mapping.ModuleName); err != nil {
		return err
	}
	if err := validateInterfaceToken("services roadmap module path", mapping.ModulePath); err != nil {
		return err
	}
	if err := validateInterfaceToken("services roadmap service id", mapping.ServiceID); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services roadmap interface hash", mapping.InterfaceHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services roadmap descriptor hash", mapping.DescriptorHash); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("services roadmap mapping hash", mapping.MappingHash); err != nil {
		return err
	}
	if expected := ComputeAetraModuleServiceMappingHash(mapping); mapping.MappingHash != expected {
		return fmt.Errorf("services roadmap mapping hash mismatch: expected %s", expected)
	}
	return nil
}

func ComputeServiceImplementationRoadmapHash(roadmap ServiceImplementationRoadmap) string {
	phases := cloneServiceRoadmapPhases(roadmap.Phases)
	sortServiceRoadmapPhases(phases)
	parts := []string{"aetra-services-implementation-roadmap-v1", fmt.Sprint(len(phases))}
	for _, phase := range phases {
		parts = append(parts, phase.PhaseHash)
	}
	return servicesHashParts(parts...)
}

func ComputeServiceRoadmapPhaseHash(phase ServiceRoadmapPhase) string {
	phase = canonicalServiceRoadmapPhase(phase)
	parts := []string{"aetra-services-roadmap-phase-v1", string(phase.PhaseID), phase.Title}
	for _, task := range phase.Tasks {
		parts = append(parts, "task", task.TaskHash)
	}
	for _, criterion := range phase.ExitCriteria {
		parts = append(parts, "exit", criterion.ExitHash)
	}
	for _, definition := range phase.CoreObjects {
		parts = append(parts, "core", definition.DefinitionHash)
	}
	for _, vector := range phase.SignableVectors {
		parts = append(parts, "vector", vector.VectorHash)
	}
	for _, mapping := range phase.ModuleMappings {
		parts = append(parts, "mapping", mapping.MappingHash)
	}
	for _, dependency := range phase.DependsOn {
		parts = append(parts, "depends", string(dependency))
	}
	return servicesHashParts(parts...)
}

func ComputeServiceRoadmapTaskHash(task ServiceRoadmapTask) string {
	return servicesHashParts("aetra-services-roadmap-task-v1", string(task.TaskID), task.Module, task.Artifact)
}

func ComputeServiceRoadmapExitCriterionHash(criterion ServiceRoadmapExitCriterion) string {
	return servicesHashParts("aetra-services-roadmap-exit-v1", string(criterion.CriterionID), criterion.Evidence, fmt.Sprint(criterion.Met))
}

func ComputeServiceCoreObjectDefinitionHash(definition ServiceCoreObjectDefinition) string {
	return servicesHashParts("aetra-services-roadmap-core-object-v1", definition.ObjectName, definition.ProtobufType, definition.CanonicalHash)
}

func ComputeServiceSignableObjectVectorHash(vector ServiceSignableObjectVector) string {
	return servicesHashParts("aetra-services-roadmap-signable-vector-v1", vector.ObjectName, vector.CanonicalEncoding, vector.TestVectorHash)
}

func ComputeAetraModuleServiceMappingHash(mapping AetraModuleServiceMapping) string {
	return servicesHashParts("aetra-services-roadmap-module-mapping-v1", mapping.ModuleName, mapping.ModulePath, mapping.ServiceID, mapping.InterfaceHash, mapping.DescriptorHash, fmt.Sprint(mapping.OnChain))
}

func IsServiceRoadmapPhaseID(phaseID ServiceRoadmapPhaseID) bool {
	switch phaseID {
	case ServiceRoadmapPhaseSpecificationCompatibility, ServiceRoadmapPhaseCoreRegistry, ServiceRoadmapPhaseInterfaceSystem, ServiceRoadmapPhaseUnifiedCallsReceipts, ServiceRoadmapPhasePayments, ServiceRoadmapPhaseOffChainMixedServices, ServiceRoadmapPhaseFogMarketProviders, ServiceRoadmapPhaseSDKUXTooling, ServiceRoadmapPhasePerformanceHardening:
		return true
	default:
		return false
	}
}

func IsServiceRoadmapTaskID(taskID ServiceRoadmapTaskID) bool {
	switch taskID {
	case ServiceRoadmapTaskAddAvailabilityCommitments, ServiceRoadmapTaskAddBlockSTMConflictBenchmarks, ServiceRoadmapTaskAddCallbacksRetries,
		ServiceRoadmapTaskAddCallEnvelopeValidation, ServiceRoadmapTaskAddCLICommandGeneration, ServiceRoadmapTaskAddCollateralStaking,
		ServiceRoadmapTaskAddDeterministicReceipts, ServiceRoadmapTaskAddEscrowSettlement, ServiceRoadmapTaskAddFallbackExecutionHooks,
		ServiceRoadmapTaskAddExportImport, ServiceRoadmapTaskAddIdentityBindingPlaceholder, ServiceRoadmapTaskAddInterfaceHashValidation,
		ServiceRoadmapTaskAddInterfaceCallBuilder, ServiceRoadmapTaskAddInterfaceProofQuery, ServiceRoadmapTaskAddInterfaceRegistration,
		ServiceRoadmapTaskAddMethodSchema, ServiceRoadmapTaskAddMeteredUsageReceipt, ServiceRoadmapTaskAddMixedDisputeLoadTests,
		ServiceRoadmapTaskAddMixedServiceChallengeFlow,
		ServiceRoadmapTaskAddNameOwnerIndexes, ServiceRoadmapTaskAddNoncesIdempotency, ServiceRoadmapTaskAddPaymentModelQuery,
		ServiceRoadmapTaskAddPerCallPayment, ServiceRoadmapTaskAddProviderCollateralPenalties, ServiceRoadmapTaskAddProviderRegistry,
		ServiceRoadmapTaskAddProviderSelectionQuery, ServiceRoadmapTaskAddProofQuery, ServiceRoadmapTaskAddProofVerificationHelpers,
		ServiceRoadmapTaskAddReceiptProofBenchmarks, ServiceRoadmapTaskAddRegistryLookupBenchmarks, ServiceRoadmapTaskAddReputationCommitments,
		ServiceRoadmapTaskAddResultAnchoring, ServiceRoadmapTaskAddSDKInterfaceVerifier, ServiceRoadmapTaskAddServiceCallThroughputTests,
		ServiceRoadmapTaskAddServiceRegistrationUpdate, ServiceRoadmapTaskAddServiceResolverSDK, ServiceRoadmapTaskAddStoreV2Benchmarks,
		ServiceRoadmapTaskAddWalletMetadataFormat, ServiceRoadmapTaskDefineCallEnvelope, ServiceRoadmapTaskDefinePaymentModelEnum,
		ServiceRoadmapTaskDefineReceiptFormat, ServiceRoadmapTaskDefineTrustVerificationEnums, ServiceRoadmapTaskFinalizeDescriptorSchema,
		ServiceRoadmapTaskFinalizeInterfaceSchema, ServiceRoadmapTaskImplementServiceCallsModule, ServiceRoadmapTaskImplementServiceInterfaceModule,
		ServiceRoadmapTaskImplementServicePaymentsModule, ServiceRoadmapTaskImplementServiceProvidersModule, ServiceRoadmapTaskImplementServiceReceiptsModule,
		ServiceRoadmapTaskImplementServicesModule, ServiceRoadmapTaskIntegrateBankFinancialZone, ServiceRoadmapTaskMapExistingModules,
		ServiceRoadmapTaskAddSignedRequestResponseFormat:
		return true
	default:
		return false
	}
}

func IsServiceRoadmapExitCriterionID(criterionID ServiceRoadmapExitCriterionID) bool {
	switch criterionID {
	case ServiceRoadmapExitClientsFetchVerifyInterfaces, ServiceRoadmapExitCallsRequireSettlePayments, ServiceRoadmapExitClientEndToEndFlow, ServiceRoadmapExitClientsQueryProviderSet,
		ServiceRoadmapExitCoreObjectsProto, ServiceRoadmapExitDescriptorProofQueryable, ServiceRoadmapExitExistingDescriptors,
		ServiceRoadmapExitInterfaceVersioning, ServiceRoadmapExitLookupsLowLatency, ServiceRoadmapExitManualABINotRequired,
		ServiceRoadmapExitMethodSchemasPublished, ServiceRoadmapExitMixedResultsChallenged,
		ServiceRoadmapExitOffChainResultsAnchored, ServiceRoadmapExitOnChainUnifiedEnvelope, ServiceRoadmapExitPaymentModelKnownBeforeSign,
		ServiceRoadmapExitPaymentTestCoverage, ServiceRoadmapExitProviderFaultCollateralFlows, ServiceRoadmapExitProviderFaultsPenalized,
		ServiceRoadmapExitProvidersAdvertiseStake, ServiceRoadmapExitProofReceiptBounded, ServiceRoadmapExitReceiptsCommittedProof,
		ServiceRoadmapExitRegistryReproducible, ServiceRoadmapExitReplayAttemptsRejected, ServiceRoadmapExitServiceDiscovery,
		ServiceRoadmapExitServicesParallelizeSafely, ServiceRoadmapExitSignableVectors:
		return true
	default:
		return false
	}
}

func newServiceRoadmapTask(taskID ServiceRoadmapTaskID, module, artifact string) ServiceRoadmapTask {
	task := ServiceRoadmapTask{TaskID: taskID, Module: strings.TrimSpace(module), Artifact: strings.TrimSpace(artifact)}
	task.TaskHash = ComputeServiceRoadmapTaskHash(task)
	return task
}

func newServiceRoadmapExitCriterion(criterionID ServiceRoadmapExitCriterionID, evidence string, met bool) ServiceRoadmapExitCriterion {
	criterion := ServiceRoadmapExitCriterion{CriterionID: criterionID, Evidence: strings.TrimSpace(evidence), Met: met}
	criterion.ExitHash = ComputeServiceRoadmapExitCriterionHash(criterion)
	return criterion
}

func newServiceCoreObjectDefinition(objectName, protobufType, canonicalHash string) ServiceCoreObjectDefinition {
	definition := ServiceCoreObjectDefinition{ObjectName: objectName, ProtobufType: protobufType, CanonicalHash: canonicalHash}
	definition.DefinitionHash = ComputeServiceCoreObjectDefinitionHash(definition)
	return definition
}

func newServiceSignableObjectVector(objectName, encoding, testVectorHash string) ServiceSignableObjectVector {
	vector := ServiceSignableObjectVector{ObjectName: objectName, CanonicalEncoding: encoding, TestVectorHash: testVectorHash}
	vector.VectorHash = ComputeServiceSignableObjectVectorHash(vector)
	return vector
}

func newAetraModuleServiceMapping(moduleName, modulePath, serviceID string, onChain bool) AetraModuleServiceMapping {
	mapping := AetraModuleServiceMapping{
		ModuleName:	moduleName,
		ModulePath:	modulePath,
		ServiceID:	serviceID,
		InterfaceHash:	servicesHashParts("aetra-services-roadmap-interface-v1", moduleName, modulePath),
		DescriptorHash:	servicesHashParts("aetra-services-roadmap-descriptor-v1", moduleName, serviceID),
		OnChain:	onChain,
	}
	mapping.MappingHash = ComputeAetraModuleServiceMappingHash(mapping)
	return mapping
}

func defaultServiceCoreObjectDefinitions() []ServiceCoreObjectDefinition {
	return []ServiceCoreObjectDefinition{
		newServiceCoreObjectDefinition("PaymentEnvelope", "aetra.services.v1.PaymentEnvelope", servicesHashParts("roadmap/core/payment-envelope")),
		newServiceCoreObjectDefinition("ServiceCallEnvelope", "aetra.services.v1.ServiceCallEnvelope", servicesHashParts("roadmap/core/service-call-envelope")),
		newServiceCoreObjectDefinition("ServiceDescriptor", "aetra.services.v1.ServiceDescriptor", servicesHashParts("roadmap/core/service-descriptor")),
		newServiceCoreObjectDefinition("ServiceInterface", "aetra.services.v1.ServiceInterface", servicesHashParts("roadmap/core/service-interface")),
		newServiceCoreObjectDefinition("ServiceReceipt", "aetra.services.v1.ServiceReceipt", servicesHashParts("roadmap/core/service-receipt")),
	}
}

func defaultServiceSignableObjectVectors() []ServiceSignableObjectVector {
	return []ServiceSignableObjectVector{
		newServiceSignableObjectVector("PaymentEnvelope", "proto3_canonical_json", servicesHashParts("roadmap/vector/payment-envelope")),
		newServiceSignableObjectVector("ServiceAdvertisement", "proto3_canonical_json", servicesHashParts("roadmap/vector/service-advertisement")),
		newServiceSignableObjectVector("ServiceCallEnvelope", "proto3_canonical_json", servicesHashParts("roadmap/vector/service-call-envelope")),
		newServiceSignableObjectVector("ServiceDescriptor", "proto3_canonical_json", servicesHashParts("roadmap/vector/service-descriptor")),
		newServiceSignableObjectVector("ServiceReceipt", "proto3_canonical_json", servicesHashParts("roadmap/vector/service-receipt")),
	}
}

func defaultAetraModuleServiceMappings() []AetraModuleServiceMapping {
	return []AetraModuleServiceMapping{
		newAetraModuleServiceMapping("aetracore", "x/aetracore", "aetracore-service", true),
		newAetraModuleServiceMapping("avm-dex-contract", "avm-dex-contract", "avm-dex-contract-service", true),
		newAetraModuleServiceMapping("fees", "x/fees", "fees-service", true),
		newAetraModuleServiceMapping("identity", "x/identity", "identity-service", true),
		newAetraModuleServiceMapping("payments", "x/payments", "payments-service", true),
		newAetraModuleServiceMapping("pos", "x/pos", "pos-service", true),
	}
}

func canonicalServiceRoadmapPhase(phase ServiceRoadmapPhase) ServiceRoadmapPhase {
	phase.Title = strings.TrimSpace(phase.Title)
	phase.Tasks = cloneServiceRoadmapTasks(phase.Tasks)
	phase.ExitCriteria = cloneServiceRoadmapExitCriteria(phase.ExitCriteria)
	phase.CoreObjects = cloneServiceCoreObjectDefinitions(phase.CoreObjects)
	phase.SignableVectors = cloneServiceSignableObjectVectors(phase.SignableVectors)
	phase.ModuleMappings = cloneAetraModuleServiceMappings(phase.ModuleMappings)
	phase.DependsOn = append([]ServiceRoadmapPhaseID(nil), phase.DependsOn...)
	sortServiceRoadmapTasks(phase.Tasks)
	sortServiceRoadmapExitCriteria(phase.ExitCriteria)
	sortServiceCoreObjectDefinitions(phase.CoreObjects)
	sortServiceSignableObjectVectors(phase.SignableVectors)
	sortAetraModuleServiceMappings(phase.ModuleMappings)
	sort.SliceStable(phase.DependsOn, func(i, j int) bool { return phase.DependsOn[i] < phase.DependsOn[j] })
	phase.PhaseHash = strings.ToLower(strings.TrimSpace(phase.PhaseHash))
	return phase
}

func validateServiceRoadmapPhases(phases []ServiceRoadmapPhase) error {
	if len(phases) != 9 {
		return errors.New("services roadmap requires phases 0, 1, 2, 3, 4, 5, 6, 7, and 8")
	}
	seen := map[ServiceRoadmapPhaseID]struct{}{}
	previous := ""
	for _, phase := range phases {
		if err := phase.Validate(); err != nil {
			return err
		}
		current := string(phase.PhaseID)
		if previous != "" && previous >= current {
			return errors.New("services roadmap phases must be sorted canonically")
		}
		previous = current
		seen[phase.PhaseID] = struct{}{}
	}
	for _, required := range []ServiceRoadmapPhaseID{ServiceRoadmapPhaseSpecificationCompatibility, ServiceRoadmapPhaseCoreRegistry, ServiceRoadmapPhaseInterfaceSystem, ServiceRoadmapPhaseUnifiedCallsReceipts, ServiceRoadmapPhasePayments, ServiceRoadmapPhaseOffChainMixedServices, ServiceRoadmapPhaseFogMarketProviders, ServiceRoadmapPhaseSDKUXTooling, ServiceRoadmapPhasePerformanceHardening} {
		if _, found := seen[required]; !found {
			return fmt.Errorf("services roadmap missing phase %s", required)
		}
	}
	return nil
}

func validateServiceRoadmapTasks(tasks []ServiceRoadmapTask) error {
	if len(tasks) == 0 {
		return errors.New("services roadmap tasks are required")
	}
	previous := ""
	seen := map[ServiceRoadmapTaskID]struct{}{}
	for _, task := range tasks {
		if err := task.Validate(); err != nil {
			return err
		}
		current := string(task.TaskID)
		if previous != "" && previous >= current {
			return errors.New("services roadmap tasks must be sorted canonically")
		}
		previous = current
		if _, found := seen[task.TaskID]; found {
			return fmt.Errorf("services roadmap duplicate task %s", task.TaskID)
		}
		seen[task.TaskID] = struct{}{}
	}
	return nil
}

func validateServiceRoadmapExitCriteria(criteria []ServiceRoadmapExitCriterion) error {
	if len(criteria) == 0 {
		return errors.New("services roadmap exit criteria are required")
	}
	previous := ""
	seen := map[ServiceRoadmapExitCriterionID]struct{}{}
	for _, criterion := range criteria {
		if err := criterion.Validate(); err != nil {
			return err
		}
		current := string(criterion.CriterionID)
		if previous != "" && previous >= current {
			return errors.New("services roadmap exit criteria must be sorted canonically")
		}
		previous = current
		if _, found := seen[criterion.CriterionID]; found {
			return fmt.Errorf("services roadmap duplicate exit criterion %s", criterion.CriterionID)
		}
		seen[criterion.CriterionID] = struct{}{}
	}
	return nil
}

func validateServiceRoadmapDependencies(phaseID ServiceRoadmapPhaseID, dependencies []ServiceRoadmapPhaseID) error {
	previous := ""
	for _, dependency := range dependencies {
		if !IsServiceRoadmapPhaseID(dependency) {
			return fmt.Errorf("services roadmap unknown dependency %q", dependency)
		}
		if dependency >= phaseID {
			return errors.New("services roadmap phase dependency must point to an earlier phase")
		}
		current := string(dependency)
		if previous != "" && previous >= current {
			return errors.New("services roadmap dependencies must be sorted canonically")
		}
		previous = current
	}
	return nil
}

func validateServiceCoreObjectDefinitions(definitions []ServiceCoreObjectDefinition) error {
	if len(definitions) == 0 {
		return errors.New("services roadmap core object definitions are required")
	}
	previous := ""
	for _, definition := range definitions {
		if err := definition.Validate(); err != nil {
			return err
		}
		if previous != "" && previous >= definition.ObjectName {
			return errors.New("services roadmap core object definitions must be sorted canonically")
		}
		previous = definition.ObjectName
	}
	return nil
}

func validateServiceSignableObjectVectors(vectors []ServiceSignableObjectVector) error {
	if len(vectors) == 0 {
		return errors.New("services roadmap signable object vectors are required")
	}
	previous := ""
	for _, vector := range vectors {
		if err := vector.Validate(); err != nil {
			return err
		}
		if previous != "" && previous >= vector.ObjectName {
			return errors.New("services roadmap signable vectors must be sorted canonically")
		}
		previous = vector.ObjectName
	}
	return nil
}

func validateRoadmapText(fieldName, value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	if len(value) > 160 {
		return fmt.Errorf("%s must be <= 160 bytes", fieldName)
	}
	for _, r := range value {
		if r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r >= '0' && r <= '9' ||
			r == '_' || r == '-' || r == '.' || r == ':' || r == '/' || r == ' ' {
			continue
		}
		return fmt.Errorf("%s contains invalid character", fieldName)
	}
	return nil
}

func cloneServiceRoadmapPhases(phases []ServiceRoadmapPhase) []ServiceRoadmapPhase {
	out := make([]ServiceRoadmapPhase, len(phases))
	for i, phase := range phases {
		out[i] = canonicalServiceRoadmapPhase(phase)
	}
	return out
}

func cloneServiceRoadmapTasks(tasks []ServiceRoadmapTask) []ServiceRoadmapTask {
	out := make([]ServiceRoadmapTask, len(tasks))
	copy(out, tasks)
	return out
}

func cloneServiceRoadmapExitCriteria(criteria []ServiceRoadmapExitCriterion) []ServiceRoadmapExitCriterion {
	out := make([]ServiceRoadmapExitCriterion, len(criteria))
	copy(out, criteria)
	return out
}

func cloneServiceCoreObjectDefinitions(definitions []ServiceCoreObjectDefinition) []ServiceCoreObjectDefinition {
	out := make([]ServiceCoreObjectDefinition, len(definitions))
	copy(out, definitions)
	return out
}

func cloneServiceSignableObjectVectors(vectors []ServiceSignableObjectVector) []ServiceSignableObjectVector {
	out := make([]ServiceSignableObjectVector, len(vectors))
	copy(out, vectors)
	return out
}

func cloneAetraModuleServiceMappings(mappings []AetraModuleServiceMapping) []AetraModuleServiceMapping {
	out := make([]AetraModuleServiceMapping, len(mappings))
	copy(out, mappings)
	return out
}

func sortServiceRoadmapPhases(phases []ServiceRoadmapPhase) {
	sort.SliceStable(phases, func(i, j int) bool { return phases[i].PhaseID < phases[j].PhaseID })
}

func sortServiceRoadmapTasks(tasks []ServiceRoadmapTask) {
	sort.SliceStable(tasks, func(i, j int) bool { return tasks[i].TaskID < tasks[j].TaskID })
}

func sortServiceRoadmapExitCriteria(criteria []ServiceRoadmapExitCriterion) {
	sort.SliceStable(criteria, func(i, j int) bool { return criteria[i].CriterionID < criteria[j].CriterionID })
}

func sortServiceCoreObjectDefinitions(definitions []ServiceCoreObjectDefinition) {
	sort.SliceStable(definitions, func(i, j int) bool { return definitions[i].ObjectName < definitions[j].ObjectName })
}

func sortServiceSignableObjectVectors(vectors []ServiceSignableObjectVector) {
	sort.SliceStable(vectors, func(i, j int) bool { return vectors[i].ObjectName < vectors[j].ObjectName })
}

func sortAetraModuleServiceMappings(mappings []AetraModuleServiceMapping) {
	sort.SliceStable(mappings, func(i, j int) bool { return mappings[i].ModuleName < mappings[j].ModuleName })
}
