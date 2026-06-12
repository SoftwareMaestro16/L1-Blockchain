package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type IdentityImplementationRoadmapPhaseIDV2 string
type IdentityImplementationRoadmapTaskIDV2 string
type IdentityImplementationRoadmapExitIDV2 string

const (
	IdentityRoadmapPhase0SpecVectorsV2	IdentityImplementationRoadmapPhaseIDV2	= "phase_0_specification_and_test_vectors"
	IdentityRoadmapPhase1CoreActivationV2	IdentityImplementationRoadmapPhaseIDV2	= "phase_1_core_registry_activation"
	IdentityRoadmapPhase2UnifiedResolverV2	IdentityImplementationRoadmapPhaseIDV2	= "phase_2_unified_resolver"
	IdentityRoadmapPhase3SubdomainsZonesV2	IdentityImplementationRoadmapPhaseIDV2	= "phase_3_subdomains_and_zone_control"
	IdentityRoadmapPhase4ProofResolutionV2	IdentityImplementationRoadmapPhaseIDV2	= "phase_4_proof_based_resolution"
	IdentityRoadmapPhase5ExecutionV2	IdentityImplementationRoadmapPhaseIDV2	= "phase_5_execution_integration"
	IdentityRoadmapPhase6PerformanceV2	IdentityImplementationRoadmapPhaseIDV2	= "phase_6_performance_hardening"

	IdentityRoadmapTaskCanonicalNameNormalizationV2	IdentityImplementationRoadmapTaskIDV2	= "define_canonical_name_normalization"
	IdentityRoadmapTaskDomainProofHashFormatsV2	IdentityImplementationRoadmapTaskIDV2	= "define_domain_hash_and_proof_hash_formats"
	IdentityRoadmapTaskProtobufStateSchemasV2	IdentityImplementationRoadmapTaskIDV2	= "define_protobuf_state_schemas"
	IdentityRoadmapTaskStoreV2KeyLayoutV2		IdentityImplementationRoadmapTaskIDV2	= "define_store_v2_key_layout"
	IdentityRoadmapTaskGovernanceParamsV2		IdentityImplementationRoadmapTaskIDV2	= "define_governance_parameter_set"
	IdentityRoadmapTaskResolutionProofVectorsV2	IdentityImplementationRoadmapTaskIDV2	= "produce_resolution_proof_test_vectors"
	IdentityRoadmapTaskLifecycleVectorsV2		IdentityImplementationRoadmapTaskIDV2	= "produce_lifecycle_transition_test_vectors"

	IdentityRoadmapTaskIdentityCoreModuleV2	IdentityImplementationRoadmapTaskIDV2	= "implement_identity_core_module"
	IdentityRoadmapTaskCoreLifecycleV2	IdentityImplementationRoadmapTaskIDV2	= "implement_registration_renewal_transfer_expiry"
	IdentityRoadmapTaskNFTBindingV2		IdentityImplementationRoadmapTaskIDV2	= "implement_nft_binding"
	IdentityRoadmapTaskOwnerExpiryIndexesV2	IdentityImplementationRoadmapTaskIDV2	= "implement_owner_and_expiry_indexes"
	IdentityRoadmapTaskCoreQueriesV2	IdentityImplementationRoadmapTaskIDV2	= "implement_core_queries"
	IdentityRoadmapTaskInvariantChecksV2	IdentityImplementationRoadmapTaskIDV2	= "add_invariant_checks"

	IdentityRoadmapTaskResolverModuleV2		IdentityImplementationRoadmapTaskIDV2	= "implement_resolver_module"
	IdentityRoadmapTaskPrimaryResolutionV2		IdentityImplementationRoadmapTaskIDV2	= "implement_primary_address_resolution"
	IdentityRoadmapTaskContractTargetsV2		IdentityImplementationRoadmapTaskIDV2	= "implement_contract_targets"
	IdentityRoadmapTaskServiceEndpointsV2		IdentityImplementationRoadmapTaskIDV2	= "implement_service_endpoints"
	IdentityRoadmapTaskInterfaceDescriptorsV2	IdentityImplementationRoadmapTaskIDV2	= "implement_interface_descriptors"
	IdentityRoadmapTaskRoutingMetadataV2		IdentityImplementationRoadmapTaskIDV2	= "implement_routing_metadata"
	IdentityRoadmapTaskReverseResolutionV2		IdentityImplementationRoadmapTaskIDV2	= "implement_reverse_resolution"
	IdentityRoadmapTaskBatchResolverUpdatesV2	IdentityImplementationRoadmapTaskIDV2	= "implement_batch_resolver_updates"

	IdentityRoadmapTaskSubdomainModuleV2		IdentityImplementationRoadmapTaskIDV2	= "implement_subdomain_module"
	IdentityRoadmapTaskDelegatedSubdomainCreationV2	IdentityImplementationRoadmapTaskIDV2	= "implement_delegated_subdomain_creation"
	IdentityRoadmapTaskPartialDelegationV2		IdentityImplementationRoadmapTaskIDV2	= "implement_partial_delegation"
	IdentityRoadmapTaskDetachedSubdomainsV2		IdentityImplementationRoadmapTaskIDV2	= "implement_detached_subdomains"
	IdentityRoadmapTaskZonePoliciesV2		IdentityImplementationRoadmapTaskIDV2	= "implement_zone_policies"
	IdentityRoadmapTaskRecursivePathQueriesV2	IdentityImplementationRoadmapTaskIDV2	= "implement_recursive_path_queries"

	IdentityRoadmapTaskProofVerificationModuleV2		IdentityImplementationRoadmapTaskIDV2	= "implement_proof_verification_module"
	IdentityRoadmapTaskDirectResolutionProofQueryV2		IdentityImplementationRoadmapTaskIDV2	= "implement_direct_resolution_proof_query"
	IdentityRoadmapTaskRecursiveResolutionProofQueryV2	IdentityImplementationRoadmapTaskIDV2	= "implement_recursive_resolution_proof_query"
	IdentityRoadmapTaskReverseProofQueryV2			IdentityImplementationRoadmapTaskIDV2	= "implement_reverse_proof_query"
	IdentityRoadmapTaskNonExistenceProofQueryV2		IdentityImplementationRoadmapTaskIDV2	= "implement_non_existence_proof_query"
	IdentityRoadmapTaskLightClientVerificationSDKV2		IdentityImplementationRoadmapTaskIDV2	= "add_light_client_verification_sdk"

	IdentityRoadmapTaskRoutingIntegrationModuleV2			IdentityImplementationRoadmapTaskIDV2	= "implement_routing_integration_module"
	IdentityRoadmapTaskSendByNameSDKHelperV2			IdentityImplementationRoadmapTaskIDV2	= "add_send_by_name_sdk_helper"
	IdentityRoadmapTaskInvokeByNameSDKHelperV2			IdentityImplementationRoadmapTaskIDV2	= "add_invoke_by_name_sdk_helper"
	IdentityRoadmapTaskServiceDiscoveryHelperV2			IdentityImplementationRoadmapTaskIDV2	= "add_service_discovery_helper"
	IdentityRoadmapTaskInterfaceDescriptorVerificationHelperV2	IdentityImplementationRoadmapTaskIDV2	= "add_interface_descriptor_verification_helper"
	IdentityRoadmapTaskWalletDisplayStatesV2			IdentityImplementationRoadmapTaskIDV2	= "add_wallet_display_state_definitions"

	IdentityRoadmapTaskStoreV2ResolutionBenchmarksV2	IdentityImplementationRoadmapTaskIDV2	= "add_store_v2_resolution_benchmarks"
	IdentityRoadmapTaskBlockSTMBatchBenchmarksV2		IdentityImplementationRoadmapTaskIDV2	= "add_blockstm_batch_update_benchmarks"
	IdentityRoadmapTaskABCIProposalGroupingV2		IdentityImplementationRoadmapTaskIDV2	= "add_abci_proposal_grouping_for_identity_transactions"
	IdentityRoadmapTaskBoundedExpiryProcessingV2		IdentityImplementationRoadmapTaskIDV2	= "add_bounded_expiry_processing"
	IdentityRoadmapTaskAdaptiveSyncRecoveryTestsV2		IdentityImplementationRoadmapTaskIDV2	= "add_adaptivesync_recovery_tests"
	IdentityRoadmapTaskCacheInvalidationEventTestsV2	IdentityImplementationRoadmapTaskIDV2	= "add_cache_invalidation_event_tests"

	IdentityRoadmapExitSignableHashableVectorsV2	IdentityImplementationRoadmapExitIDV2	= "all_signable_and_hashable_identity_objects_have_test_vectors"
	IdentityRoadmapExitLifecycleDeterminismV2	IdentityImplementationRoadmapExitIDV2	= "all_lifecycle_states_have_deterministic_transition_tests"
	IdentityRoadmapExitStorePrefixesFinalizedV2	IdentityImplementationRoadmapExitIDV2	= "store_key_prefixes_are_finalized"

	IdentityRoadmapExitOnChainOwnershipV2		IdentityImplementationRoadmapExitIDV2	= "aet_domain_ownership_is_fully_on_chain"
	IdentityRoadmapExitAtomicNFTOwnershipV2		IdentityImplementationRoadmapExitIDV2	= "nft_and_registry_ownership_remain_atomic"
	IdentityRoadmapExitExportImportRegistryV2	IdentityImplementationRoadmapExitIDV2	= "export_import_preserves_registry_state"

	IdentityRoadmapExitUnifiedTargetsV2		IdentityImplementationRoadmapExitIDV2	= "unified_resolver_supports_wallet_contract_service_interface_routing_targets"
	IdentityRoadmapExitReverseConsistencyV2		IdentityImplementationRoadmapExitIDV2	= "reverse_resolution_verifies_forward_consistency"
	IdentityRoadmapExitVersionedSizeBoundedV2	IdentityImplementationRoadmapExitIDV2	= "resolver_updates_are_versioned_and_size_bounded"

	IdentityRoadmapExitRecursiveScopedDelegationV2	IdentityImplementationRoadmapExitIDV2	= "recursive_hierarchy_supports_scoped_delegation"
	IdentityRoadmapExitParentChildExpiryRulesV2	IdentityImplementationRoadmapExitIDV2	= "parent_and_child_expiry_rules_are_enforced"
	IdentityRoadmapExitZonePolicyProofQueryableV2	IdentityImplementationRoadmapExitIDV2	= "zone_policy_is_proof_queryable"

	IdentityRoadmapExitLightClientAllTargetsV2	IdentityImplementationRoadmapExitIDV2	= "light_clients_can_verify_address_contract_service_interface_reverse_resolution"
	IdentityRoadmapExitExplicitProofFailuresV2	IdentityImplementationRoadmapExitIDV2	= "proof_failure_modes_are_explicit"
	IdentityRoadmapExitProofVectorsModuleVersionsV2	IdentityImplementationRoadmapExitIDV2	= "proof_test_vectors_pass_across_module_versions"

	IdentityRoadmapExitPresigningTargetResolutionV2		IdentityImplementationRoadmapExitIDV2	= "identity_records_drive_pre_signing_transaction_target_resolution"
	IdentityRoadmapExitWalletProofBackedTargetsV2		IdentityImplementationRoadmapExitIDV2	= "wallets_can_verify_proof_backed_identity_targets"
	IdentityRoadmapExitInterfaceServiceHashVerifiedV2	IdentityImplementationRoadmapExitIDV2	= "interface_and_service_metadata_are_hash_verified_before_use"

	IdentityRoadmapExitIndependentUpdatesParallelizeV2	IdentityImplementationRoadmapExitIDV2	= "independent_identity_updates_parallelize_without_avoidable_conflicts"
	IdentityRoadmapExitDirectResolutionBoundedLargeStateV2	IdentityImplementationRoadmapExitIDV2	= "direct_resolution_remains_bounded_with_large_state"
	IdentityRoadmapExitPostSyncProofQueriesV2		IdentityImplementationRoadmapExitIDV2	= "recovering_nodes_can_serve_proof_queries_after_sync"
)

type IdentityRoadmapTaskV2 struct {
	ID		IdentityImplementationRoadmapTaskIDV2
	Evidence	[]string
}

type IdentityRoadmapExitCriterionV2 struct {
	ID		IdentityImplementationRoadmapExitIDV2
	Evidence	[]string
}

type IdentityRoadmapPhaseV2 struct {
	ID		IdentityImplementationRoadmapPhaseIDV2
	Title		string
	Tasks		[]IdentityRoadmapTaskV2
	ExitCriteria	[]IdentityRoadmapExitCriterionV2
}

type IdentityImplementationRoadmapV2 struct {
	Phases		[]IdentityRoadmapPhaseV2
	RoadmapHash	string
}

func DefaultIdentityImplementationRoadmapV2() IdentityImplementationRoadmapV2 {
	roadmap := IdentityImplementationRoadmapV2{Phases: []IdentityRoadmapPhaseV2{
		{
			ID:	IdentityRoadmapPhase0SpecVectorsV2,
			Title:	"Specification and Test Vectors",
			Tasks: []IdentityRoadmapTaskV2{
				{ID: IdentityRoadmapTaskCanonicalNameNormalizationV2, Evidence: []string{"x/identity/types/validation_v2.go:NormalizeAETDomainVersioned", "x/identity/types/validation_v2_test.go:TestNameNormalizationV2ValidAndInvalidVectors"}},
				{ID: IdentityRoadmapTaskDomainProofHashFormatsV2, Evidence: []string{"x/identity/types/domain_v2.go:DomainRecordV2NameHash", "x/identity/types/proof_format_v2.go:ComputeIdentityResolutionProofCommitmentHashV2", "x/identity/types/proof_format_v2.go:ComputeRecursiveResolutionProofCommitmentHashV2"}},
				{ID: IdentityRoadmapTaskProtobufStateSchemasV2, Evidence: []string{"x/identity/types/identity_core_module_breakdown_v2.go:DefaultIdentityCoreModuleBreakdownV2", "x/identity/types/resolver_subdomain_module_breakdown_v2.go:DefaultResolverModuleBreakdownV2", "x/identity/types/resolver_subdomain_module_breakdown_v2.go:DefaultSubdomainModuleBreakdownV2"}},
				{ID: IdentityRoadmapTaskStoreV2KeyLayoutV2, Evidence: []string{"x/identity/types/storev2.go:IdentityStoreV2SpecDomainKey", "x/identity/types/storev2.go:IdentityStoreV2SpecResolutionProofReadAccessSet"}},
				{ID: IdentityRoadmapTaskGovernanceParamsV2, Evidence: []string{"x/identity/types/governance_params_v2.go:DefaultIdentityGovernanceParamsV2", "x/identity/types/governance_params_v2.go:ValidateIdentityGovernanceParamsV2"}},
				{ID: IdentityRoadmapTaskResolutionProofVectorsV2, Evidence: []string{"x/identity/types/proof_format_v2.go:BuildIdentityResolutionProofFormatV2", "x/identity/types/proof_format_v2.go:ValidateIdentityResolutionProofFormatV2", "x/identity/types/proof_format_v2_test.go:TestIdentityResolutionProofFormatV2FieldsEncodingAndCommitment"}},
				{ID: IdentityRoadmapTaskLifecycleVectorsV2, Evidence: []string{"x/identity/types/lifecycle_state_machine_v2.go:ApplyDomainLifecycleTransitionV2", "x/identity/types/lifecycle_state_machine_v2_test.go:TestDomainLifecycleStateMachineV2RegistrationRenewalGraceAndRelease"}},
			},
			ExitCriteria: []IdentityRoadmapExitCriterionV2{
				{ID: IdentityRoadmapExitSignableHashableVectorsV2, Evidence: []string{"x/identity/types/proof_format_v2.go:ComputeIdentityResolutionProofCommitmentHashV2", "x/identity/types/proof_format_v2_test.go:TestIdentityResolutionProofFormatV2FieldsEncodingAndCommitment"}},
				{ID: IdentityRoadmapExitLifecycleDeterminismV2, Evidence: []string{"x/identity/types/lifecycle_state_machine_v2.go:ApplyDomainLifecycleTransitionV2", "x/identity/types/lifecycle_state_machine_v2_test.go:TestDomainLifecycleStateMachineV2AuctionAlternative"}},
				{ID: IdentityRoadmapExitStorePrefixesFinalizedV2, Evidence: []string{"x/identity/types/storev2.go:IdentityStoreV2SpecDomainKey", "x/identity/types/storev2_spec_test.go:TestIdentityStoreV2SpecPrimaryKeyLayout"}},
			},
		},
		{
			ID:	IdentityRoadmapPhase1CoreActivationV2,
			Title:	"Core Registry Activation",
			Tasks: []IdentityRoadmapTaskV2{
				{ID: IdentityRoadmapTaskIdentityCoreModuleV2, Evidence: []string{"x/identity/types/identity_core_module_breakdown_v2.go:DefaultIdentityCoreModuleBreakdownV2", "x/identity/types/identity_core_module_breakdown_v2_test.go:TestIdentityCoreModuleBreakdownV2CoversSection131"}},
				{ID: IdentityRoadmapTaskCoreLifecycleV2, Evidence: []string{"x/identity/types/anti_squatting_v2.go:ReleaseExpiredIdentityDomainV2", "x/identity/types/spec_state.go:CommitDomainRegistration", "x/identity/types/spec_state.go:RevealRegisterDomain", "x/identity/types/spec_state.go:TransferDomainNFT"}},
				{ID: IdentityRoadmapTaskNFTBindingV2, Evidence: []string{"x/identity/types/nft_binding.go:TransferDomainNFTBindingAtomic", "x/identity/types/validation_v2.go:TransferDomainNFTBindingWithInvariantsV2"}},
				{ID: IdentityRoadmapTaskOwnerExpiryIndexesV2, Evidence: []string{"x/identity/types/identity_core_module_breakdown_v2.go:BuildIdentityCoreDerivedIndexesV2", "x/identity/types/identity_core_module_breakdown_v2.go:ValidateIdentityCoreStoreV2IndexesV2"}},
				{ID: IdentityRoadmapTaskCoreQueriesV2, Evidence: []string{"x/identity/types/identity_core_module_breakdown_v2.go:DefaultIdentityCoreModuleBreakdownV2", "x/identity/types/query_v2.go:NewIdentityQueryServiceV2"}},
				{ID: IdentityRoadmapTaskInvariantChecksV2, Evidence: []string{"x/identity/types/identity_core_module_breakdown_v2.go:ValidateIdentityCoreModuleInvariantsV2", "x/identity/types/ownership_consistency_v2_test.go:TestIdentityConsistencyAuditDetectsInvariantsV2"}},
			},
			ExitCriteria: []IdentityRoadmapExitCriterionV2{
				{ID: IdentityRoadmapExitOnChainOwnershipV2, Evidence: []string{"x/identity/types/spec_state.go:RevealRegisterDomain", "x/identity/types/spec_test.go:TestIdentitySpecRegisterAETDomain"}},
				{ID: IdentityRoadmapExitAtomicNFTOwnershipV2, Evidence: []string{"x/identity/types/nft_binding.go:TransferDomainNFTBindingAtomic", "x/identity/types/ownership_consistency_v2_test.go:TestIdentityNFTTransferHooksUpdateOrRejectAtomicallyV2"}},
				{ID: IdentityRoadmapExitExportImportRegistryV2, Evidence: []string{"x/identity/types/spec_state.go:ImportIdentityState", "x/identity/types/spec_test.go:TestIdentitySpecExportImportPreservesDomainLifecycle"}},
			},
		},
		{
			ID:	IdentityRoadmapPhase2UnifiedResolverV2,
			Title:	"Unified Resolver",
			Tasks: []IdentityRoadmapTaskV2{
				{ID: IdentityRoadmapTaskResolverModuleV2, Evidence: []string{"x/identity/types/resolver_subdomain_module_breakdown_v2.go:DefaultResolverModuleBreakdownV2", "x/identity/types/resolver_subdomain_module_breakdown_v2_test.go:TestResolverModuleBreakdownV2CoversSection132"}},
				{ID: IdentityRoadmapTaskPrimaryResolutionV2, Evidence: []string{"x/identity/types/resolution_v2.go:BuildUnifiedResolutionRecordV2", "x/identity/types/spec_state.go:ResolveIdentityAddress"}},
				{ID: IdentityRoadmapTaskContractTargetsV2, Evidence: []string{"x/identity/types/resolution_v2.go:NewContractTargetV2", "x/identity/types/routing_integration_module_breakdown_v2.go:BuildRoutingIntegrationContractInvocationMappingV2"}},
				{ID: IdentityRoadmapTaskServiceEndpointsV2, Evidence: []string{"x/identity/types/service_interface_mapping_v2.go:BuildIdentityServiceDiscoveryV2", "x/identity/types/service_interface_mapping_v2.go:DefaultIdentityServiceEndpointTypeRegistryV2"}},
				{ID: IdentityRoadmapTaskInterfaceDescriptorsV2, Evidence: []string{"x/identity/types/resolution_v2.go:InterfaceDescriptorHashV2", "x/identity/types/service_interface_mapping_v2.go:BuildIdentityInterfaceSchemaMappingV2"}},
				{ID: IdentityRoadmapTaskRoutingMetadataV2, Evidence: []string{"x/identity/types/routing_integration_module_breakdown_v2.go:BuildRoutingIntegrationWalletSDKHelperV2", "x/identity/types/routing_integration_module_breakdown_v2.go:DefaultRoutingIntegrationModuleBreakdownV2"}},
				{ID: IdentityRoadmapTaskReverseResolutionV2, Evidence: []string{"x/identity/types/resolution_v2.go:ValidateReverseResolutionRecordV2", "x/identity/types/resolution_v2.go:VerifyReverseResolutionTransactionV2"}},
				{ID: IdentityRoadmapTaskBatchResolverUpdatesV2, Evidence: []string{"x/identity/types/batch_resolver_v2.go:ExecuteBatchResolverUpdatesV2", "x/identity/types/batch_resolver_v2.go:ValidateBatchResolverUpdateResponseV2"}},
			},
			ExitCriteria: []IdentityRoadmapExitCriterionV2{
				{ID: IdentityRoadmapExitUnifiedTargetsV2, Evidence: []string{"x/identity/types/resolution_v2.go:BuildUnifiedResolutionRecordV2", "x/identity/types/v2_test.go:TestUnifiedResolverMetadataAndNamedExecution"}},
				{ID: IdentityRoadmapExitReverseConsistencyV2, Evidence: []string{"x/identity/types/resolution_v2.go:ValidateReverseResolutionRecordV2", "x/identity/types/resolution_v2_test.go:TestReverseResolutionVerificationTransactionV2ChecksVersionAndForward"}},
				{ID: IdentityRoadmapExitVersionedSizeBoundedV2, Evidence: []string{"x/identity/types/batch_resolver_v2.go:ExecuteBatchResolverUpdatesV2", "x/identity/types/resolution_v2.go:ValidateUnifiedResolutionRecordV2"}},
			},
		},
		{
			ID:	IdentityRoadmapPhase3SubdomainsZonesV2,
			Title:	"Subdomains and Zone Control",
			Tasks: []IdentityRoadmapTaskV2{
				{ID: IdentityRoadmapTaskSubdomainModuleV2, Evidence: []string{"x/identity/types/resolver_subdomain_module_breakdown_v2.go:DefaultSubdomainModuleBreakdownV2", "x/identity/types/resolver_subdomain_module_breakdown_v2_test.go:TestSubdomainModuleBreakdownV2CoversSection133"}},
				{ID: IdentityRoadmapTaskDelegatedSubdomainCreationV2, Evidence: []string{"x/identity/types/hierarchy_v2.go:IssueSubdomainV2", "x/identity/types/hierarchy_v2.go:ValidateSubdomainCreationV2", "x/identity/types/validation_v2.go:ValidateSubdomainCreationAuthorizationV2"}},
				{ID: IdentityRoadmapTaskPartialDelegationV2, Evidence: []string{"x/identity/types/delegation_auction_v2.go:ValidateDelegationDoesNotEscalateV2", "x/identity/types/delegation_auction_v2.go:ValidateDelegationRecordV2Use", "x/identity/types/delegation_auction_v2.go:ValidatePartialDelegationAuthorizationV2"}},
				{ID: IdentityRoadmapTaskDetachedSubdomainsV2, Evidence: []string{"x/identity/types/cost_models_v2.go:QuoteIdentitySubdomainCreationFeeV2", "x/identity/types/hierarchy_v2.go:IssueSubdomainV2"}},
				{ID: IdentityRoadmapTaskZonePoliciesV2, Evidence: []string{"x/identity/types/hierarchy_v2.go:ApplyZonePolicyChangeV2", "x/identity/types/hierarchy_v2.go:NewZonePolicyV2", "x/identity/types/hierarchy_v2.go:ValidateZonePolicyV2"}},
				{ID: IdentityRoadmapTaskRecursivePathQueriesV2, Evidence: []string{"x/identity/types/query_v2.go:QueryOptimizedRecursiveResolutionProof", "x/identity/types/query_v2.go:QueryRecursiveResolutionProof", "x/identity/types/resolver_subdomain_module_breakdown_v2.go:ValidateSubdomainModulePathPolicyV2"}},
			},
			ExitCriteria: []IdentityRoadmapExitCriterionV2{
				{ID: IdentityRoadmapExitRecursiveScopedDelegationV2, Evidence: []string{"x/identity/types/delegation_auction_v2.go:ValidateDelegationRecordV2Use", "x/identity/types/resolution_path_v2.go:VerifyDeterministicResolutionPathV2"}},
				{ID: IdentityRoadmapExitParentChildExpiryRulesV2, Evidence: []string{"x/identity/types/cost_models_v2_test.go:TestIdentitySubdomainCreationCostModelV2ExpiryConstraints", "x/identity/types/hierarchy_v2.go:ValidateSubdomainCreationV2"}},
				{ID: IdentityRoadmapExitZonePolicyProofQueryableV2, Evidence: []string{"x/identity/types/hierarchy_v2.go:BuildRecursivePolicyProofV2", "x/identity/types/hierarchy_v2.go:ValidateRecursivePolicyProofV2"}},
			},
		},
		{
			ID:	IdentityRoadmapPhase4ProofResolutionV2,
			Title:	"Proof-Based Resolution",
			Tasks: []IdentityRoadmapTaskV2{
				{ID: IdentityRoadmapTaskProofVerificationModuleV2, Evidence: []string{"x/identity/types/auction_proof_module_breakdown_v2.go:DefaultProofVerificationModuleBreakdownV2", "x/identity/types/auction_proof_module_breakdown_v2_test.go:TestProofVerificationModuleBreakdownV2CoversSection135"}},
				{ID: IdentityRoadmapTaskDirectResolutionProofQueryV2, Evidence: []string{"x/identity/types/auction_proof_module_breakdown_v2.go:BuildProofModuleResolutionProofV2", "x/identity/types/query_v2.go:QueryResolutionProof"}},
				{ID: IdentityRoadmapTaskRecursiveResolutionProofQueryV2, Evidence: []string{"x/identity/types/proof_format_v2.go:BuildRecursiveResolutionProofV2", "x/identity/types/query_v2.go:QueryRecursiveResolutionProof"}},
				{ID: IdentityRoadmapTaskReverseProofQueryV2, Evidence: []string{"x/identity/types/auction_proof_module_breakdown_v2.go:BuildProofModuleReverseResolutionProofV2", "x/identity/types/reverse_safety_v2.go:BuildVerifiedReverseResolutionProofV2"}},
				{ID: IdentityRoadmapTaskNonExistenceProofQueryV2, Evidence: []string{"x/identity/types/auction_proof_module_breakdown_v2.go:BuildProofModuleNonExistenceProofV2", "x/identity/types/proof_format_v2_test.go:TestIdentityResolutionProofFormatV2NonExistenceProof"}},
				{ID: IdentityRoadmapTaskLightClientVerificationSDKV2, Evidence: []string{"x/identity/types/api_sdk_requirements_v2.go:IdentityWalletSDKVerifyResolutionProofV2", "x/identity/types/light_client_verifier_v2.go:VerifyIdentityResolutionProofLightClientV2"}},
			},
			ExitCriteria: []IdentityRoadmapExitCriterionV2{
				{ID: IdentityRoadmapExitLightClientAllTargetsV2, Evidence: []string{"x/identity/types/light_client_verifier_v2.go:VerifyIdentityResolutionProofLightClientV2", "x/identity/types/proof_v2_test.go:TestIdentityLightClientProofV2VerifiesAllResolutionObjectives"}},
				{ID: IdentityRoadmapExitExplicitProofFailuresV2, Evidence: []string{"x/identity/types/failure_handling_v2.go:HandleIdentityLightClientFailureV2", "x/identity/types/light_client_verifier_v2.go:IdentityLightClientFailureCodeFromErrorV2"}},
				{ID: IdentityRoadmapExitProofVectorsModuleVersionsV2, Evidence: []string{"x/identity/types/proof_format_v2.go:ValidateIdentityResolutionProofFormatV2", "x/identity/types/proof_format_v2_test.go:TestIdentityResolutionProofFormatV2FieldsEncodingAndCommitment"}},
			},
		},
		{
			ID:	IdentityRoadmapPhase5ExecutionV2,
			Title:	"Execution Integration",
			Tasks: []IdentityRoadmapTaskV2{
				{ID: IdentityRoadmapTaskRoutingIntegrationModuleV2, Evidence: []string{"x/identity/types/routing_integration_module_breakdown_v2.go:DefaultRoutingIntegrationModuleBreakdownV2", "x/identity/types/routing_integration_module_breakdown_v2_test.go:TestRoutingIntegrationModuleBreakdownV2CoversSection136"}},
				{ID: IdentityRoadmapTaskSendByNameSDKHelperV2, Evidence: []string{"x/identity/types/api_sdk_requirements_v2.go:IdentityWalletSDKBuildSendByNameTxV2", "x/identity/types/execution_integration_v2.go:BuildIdentitySendByNameV2"}},
				{ID: IdentityRoadmapTaskInvokeByNameSDKHelperV2, Evidence: []string{"x/identity/types/api_sdk_requirements_v2.go:IdentityWalletSDKBuildInvokeByNameTxV2", "x/identity/types/execution_integration_v2.go:BuildIdentityInvokeByNameV2"}},
				{ID: IdentityRoadmapTaskServiceDiscoveryHelperV2, Evidence: []string{"x/identity/types/api_sdk_requirements_v2.go:IdentityWalletSDKResolveServiceVerifiedV2", "x/identity/types/service_interface_mapping_v2.go:BuildIdentityServiceDiscoveryV2"}},
				{ID: IdentityRoadmapTaskInterfaceDescriptorVerificationHelperV2, Evidence: []string{"x/identity/types/service_interface_mapping_v2.go:BuildIdentityInterfaceSchemaMappingV2", "x/identity/types/service_interface_mapping_v2_test.go:TestIdentityInterfaceSchemaMappingV2PolicyAndHashVerification"}},
				{ID: IdentityRoadmapTaskWalletDisplayStatesV2, Evidence: []string{"x/identity/types/failure_handling_v2.go:EvaluateIdentityWalletResolutionStatusV2", "x/identity/types/reverse_safety_v2.go:BuildIdentityReverseWalletDisplayStateV2"}},
			},
			ExitCriteria: []IdentityRoadmapExitCriterionV2{
				{ID: IdentityRoadmapExitPresigningTargetResolutionV2, Evidence: []string{"x/identity/types/execution_integration_v2.go:BuildIdentitySendByNameV2", "x/identity/types/routing_integration_module_breakdown_v2.go:BuildRoutingIntegrationTransactionMappingV2"}},
				{ID: IdentityRoadmapExitWalletProofBackedTargetsV2, Evidence: []string{"x/identity/types/api_sdk_requirements_v2.go:IdentityWalletSDKResolvePrimaryVerifiedV2", "x/identity/types/api_sdk_requirements_v2.go:IdentityWalletSDKVerifyResolutionProofV2"}},
				{ID: IdentityRoadmapExitInterfaceServiceHashVerifiedV2, Evidence: []string{"x/identity/types/service_interface_mapping_v2.go:BuildIdentityInterfaceSchemaMappingV2", "x/identity/types/service_interface_mapping_v2_test.go:TestIdentityServiceDiscoveryV2ExternalMetadataHash"}},
			},
		},
		{
			ID:	IdentityRoadmapPhase6PerformanceV2,
			Title:	"Performance Hardening",
			Tasks: []IdentityRoadmapTaskV2{
				{ID: IdentityRoadmapTaskStoreV2ResolutionBenchmarksV2, Evidence: []string{"x/identity/types/bench_test.go:BenchmarkIdentityStoreV2DirectResolutionReadPath", "x/identity/types/bench_test.go:BenchmarkIdentityStoreV2RecursiveResolutionReadPath"}},
				{ID: IdentityRoadmapTaskBlockSTMBatchBenchmarksV2, Evidence: []string{"x/identity/types/bench_test.go:BenchmarkIdentityBlockSTMBatchResolverUpdates", "x/identity/types/bench_test.go:BenchmarkIdentityBlockSTMMixedConflictClassification"}},
				{ID: IdentityRoadmapTaskABCIProposalGroupingV2, Evidence: []string{"x/identity/types/abci_lifecycle_v2.go:GroupIdentityProposalUpdatesV2", "x/identity/types/abci_lifecycle_v2_test.go:TestIdentityABCIPlusPrecheckAndProposalGroupingV2"}},
				{ID: IdentityRoadmapTaskBoundedExpiryProcessingV2, Evidence: []string{"x/identity/types/abci_lifecycle_v2.go:FinalizeIdentityABCIPlusV2", "x/identity/types/abci_lifecycle_v2_test.go:TestIdentityABCIPlusFinalizeBoundedExpiryAndCacheEventsV2"}},
				{ID: IdentityRoadmapTaskAdaptiveSyncRecoveryTestsV2, Evidence: []string{"x/identity/types/adaptive_sync_v2.go:RestoreIdentityAdaptiveSyncSnapshotV2", "x/identity/types/adaptive_sync_v2_test.go:TestIdentityAdaptiveSyncSnapshotRestoreProofReadyWithAuctionsAndExpiryV2"}},
				{ID: IdentityRoadmapTaskCacheInvalidationEventTestsV2, Evidence: []string{"x/identity/types/resolution_cache_v2.go:InvalidateIdentityResolutionCachesV2", "x/identity/types/resolution_cache_v2.go:ValidateIdentityCacheInvalidationEventV2"}},
			},
			ExitCriteria: []IdentityRoadmapExitCriterionV2{
				{ID: IdentityRoadmapExitIndependentUpdatesParallelizeV2, Evidence: []string{"x/identity/types/blockstm_v2.go:IdentityBlockSTMBatchResolverAccessSetV2", "x/identity/types/blockstm_v2_test.go:TestIdentityBlockSTMBatchResolverUpdatesUseDisjointNameHashesV2"}},
				{ID: IdentityRoadmapExitDirectResolutionBoundedLargeStateV2, Evidence: []string{"x/identity/types/bench_test.go:BenchmarkIdentityStoreV2DirectResolutionReadPath", "x/identity/types/storev2_spec_test.go:TestIdentityStoreV2SpecPerformanceAccessPatterns"}},
				{ID: IdentityRoadmapExitPostSyncProofQueriesV2, Evidence: []string{"x/identity/types/adaptive_sync_v2.go:RestoreIdentityAdaptiveSyncSnapshotV2", "x/identity/types/adaptive_sync_v2_test.go:TestIdentityAdaptiveSyncSnapshotRestoreProofReadyWithAuctionsAndExpiryV2"}},
			},
		},
	}}
	roadmap.RoadmapHash = ComputeIdentityImplementationRoadmapHashV2(roadmap)
	return roadmap
}

func ValidateIdentityImplementationRoadmapV2(roadmap IdentityImplementationRoadmapV2) error {
	required := requiredIdentityRoadmapPhaseIDsV2()
	if len(roadmap.Phases) != len(required) {
		return fmt.Errorf("identity implementation roadmap must define %d phases", len(required))
	}
	for i, phaseID := range required {
		if roadmap.Phases[i].ID != phaseID {
			return fmt.Errorf("identity implementation roadmap phase %d must be %s", i, phaseID)
		}
		if err := validateIdentityRoadmapPhaseV2(roadmap.Phases[i]); err != nil {
			return err
		}
	}
	if roadmap.RoadmapHash == "" || roadmap.RoadmapHash != ComputeIdentityImplementationRoadmapHashV2(roadmap) {
		return errors.New("identity implementation roadmap hash mismatch")
	}
	return nil
}

func ComputeIdentityImplementationRoadmapHashV2(roadmap IdentityImplementationRoadmapV2) string {
	parts := []string{"identity-implementation-roadmap-v2"}
	for _, phase := range roadmap.Phases {
		parts = append(parts, string(phase.ID), phase.Title)
		for _, task := range phase.Tasks {
			parts = append(parts, string(task.ID))
			parts = append(parts, sortedBreakdownStringsV2(task.Evidence)...)
		}
		for _, criterion := range phase.ExitCriteria {
			parts = append(parts, string(criterion.ID))
			parts = append(parts, sortedBreakdownStringsV2(criterion.Evidence)...)
		}
	}
	return identityHash(parts...)
}

func requiredIdentityRoadmapPhaseIDsV2() []IdentityImplementationRoadmapPhaseIDV2 {
	return []IdentityImplementationRoadmapPhaseIDV2{
		IdentityRoadmapPhase0SpecVectorsV2,
		IdentityRoadmapPhase1CoreActivationV2,
		IdentityRoadmapPhase2UnifiedResolverV2,
		IdentityRoadmapPhase3SubdomainsZonesV2,
		IdentityRoadmapPhase4ProofResolutionV2,
		IdentityRoadmapPhase5ExecutionV2,
		IdentityRoadmapPhase6PerformanceV2,
	}
}

func validateIdentityRoadmapPhaseV2(phase IdentityRoadmapPhaseV2) error {
	if phase.Title == "" {
		return fmt.Errorf("identity implementation roadmap phase %s title is required", phase.ID)
	}
	if !identityRoadmapTasksEqualV2(phase.Tasks, requiredIdentityRoadmapTasksV2(phase.ID)) {
		return fmt.Errorf("identity implementation roadmap phase %s tasks mismatch", phase.ID)
	}
	if !identityRoadmapExitsEqualV2(phase.ExitCriteria, requiredIdentityRoadmapExitsV2(phase.ID)) {
		return fmt.Errorf("identity implementation roadmap phase %s exit criteria mismatch", phase.ID)
	}
	for _, task := range phase.Tasks {
		if err := validateIdentityRoadmapEvidenceV2("task", string(task.ID), task.Evidence); err != nil {
			return err
		}
	}
	for _, criterion := range phase.ExitCriteria {
		if err := validateIdentityRoadmapEvidenceV2("exit criterion", string(criterion.ID), criterion.Evidence); err != nil {
			return err
		}
	}
	return nil
}

func requiredIdentityRoadmapTasksV2(phase IdentityImplementationRoadmapPhaseIDV2) []IdentityImplementationRoadmapTaskIDV2 {
	switch phase {
	case IdentityRoadmapPhase0SpecVectorsV2:
		return []IdentityImplementationRoadmapTaskIDV2{IdentityRoadmapTaskCanonicalNameNormalizationV2, IdentityRoadmapTaskDomainProofHashFormatsV2, IdentityRoadmapTaskProtobufStateSchemasV2, IdentityRoadmapTaskStoreV2KeyLayoutV2, IdentityRoadmapTaskGovernanceParamsV2, IdentityRoadmapTaskResolutionProofVectorsV2, IdentityRoadmapTaskLifecycleVectorsV2}
	case IdentityRoadmapPhase1CoreActivationV2:
		return []IdentityImplementationRoadmapTaskIDV2{IdentityRoadmapTaskIdentityCoreModuleV2, IdentityRoadmapTaskCoreLifecycleV2, IdentityRoadmapTaskNFTBindingV2, IdentityRoadmapTaskOwnerExpiryIndexesV2, IdentityRoadmapTaskCoreQueriesV2, IdentityRoadmapTaskInvariantChecksV2}
	case IdentityRoadmapPhase2UnifiedResolverV2:
		return []IdentityImplementationRoadmapTaskIDV2{IdentityRoadmapTaskResolverModuleV2, IdentityRoadmapTaskPrimaryResolutionV2, IdentityRoadmapTaskContractTargetsV2, IdentityRoadmapTaskServiceEndpointsV2, IdentityRoadmapTaskInterfaceDescriptorsV2, IdentityRoadmapTaskRoutingMetadataV2, IdentityRoadmapTaskReverseResolutionV2, IdentityRoadmapTaskBatchResolverUpdatesV2}
	case IdentityRoadmapPhase3SubdomainsZonesV2:
		return []IdentityImplementationRoadmapTaskIDV2{IdentityRoadmapTaskSubdomainModuleV2, IdentityRoadmapTaskDelegatedSubdomainCreationV2, IdentityRoadmapTaskPartialDelegationV2, IdentityRoadmapTaskDetachedSubdomainsV2, IdentityRoadmapTaskZonePoliciesV2, IdentityRoadmapTaskRecursivePathQueriesV2}
	case IdentityRoadmapPhase4ProofResolutionV2:
		return []IdentityImplementationRoadmapTaskIDV2{IdentityRoadmapTaskProofVerificationModuleV2, IdentityRoadmapTaskDirectResolutionProofQueryV2, IdentityRoadmapTaskRecursiveResolutionProofQueryV2, IdentityRoadmapTaskReverseProofQueryV2, IdentityRoadmapTaskNonExistenceProofQueryV2, IdentityRoadmapTaskLightClientVerificationSDKV2}
	case IdentityRoadmapPhase5ExecutionV2:
		return []IdentityImplementationRoadmapTaskIDV2{IdentityRoadmapTaskRoutingIntegrationModuleV2, IdentityRoadmapTaskSendByNameSDKHelperV2, IdentityRoadmapTaskInvokeByNameSDKHelperV2, IdentityRoadmapTaskServiceDiscoveryHelperV2, IdentityRoadmapTaskInterfaceDescriptorVerificationHelperV2, IdentityRoadmapTaskWalletDisplayStatesV2}
	case IdentityRoadmapPhase6PerformanceV2:
		return []IdentityImplementationRoadmapTaskIDV2{IdentityRoadmapTaskStoreV2ResolutionBenchmarksV2, IdentityRoadmapTaskBlockSTMBatchBenchmarksV2, IdentityRoadmapTaskABCIProposalGroupingV2, IdentityRoadmapTaskBoundedExpiryProcessingV2, IdentityRoadmapTaskAdaptiveSyncRecoveryTestsV2, IdentityRoadmapTaskCacheInvalidationEventTestsV2}
	default:
		return nil
	}
}

func requiredIdentityRoadmapExitsV2(phase IdentityImplementationRoadmapPhaseIDV2) []IdentityImplementationRoadmapExitIDV2 {
	switch phase {
	case IdentityRoadmapPhase0SpecVectorsV2:
		return []IdentityImplementationRoadmapExitIDV2{IdentityRoadmapExitSignableHashableVectorsV2, IdentityRoadmapExitLifecycleDeterminismV2, IdentityRoadmapExitStorePrefixesFinalizedV2}
	case IdentityRoadmapPhase1CoreActivationV2:
		return []IdentityImplementationRoadmapExitIDV2{IdentityRoadmapExitOnChainOwnershipV2, IdentityRoadmapExitAtomicNFTOwnershipV2, IdentityRoadmapExitExportImportRegistryV2}
	case IdentityRoadmapPhase2UnifiedResolverV2:
		return []IdentityImplementationRoadmapExitIDV2{IdentityRoadmapExitUnifiedTargetsV2, IdentityRoadmapExitReverseConsistencyV2, IdentityRoadmapExitVersionedSizeBoundedV2}
	case IdentityRoadmapPhase3SubdomainsZonesV2:
		return []IdentityImplementationRoadmapExitIDV2{IdentityRoadmapExitRecursiveScopedDelegationV2, IdentityRoadmapExitParentChildExpiryRulesV2, IdentityRoadmapExitZonePolicyProofQueryableV2}
	case IdentityRoadmapPhase4ProofResolutionV2:
		return []IdentityImplementationRoadmapExitIDV2{IdentityRoadmapExitLightClientAllTargetsV2, IdentityRoadmapExitExplicitProofFailuresV2, IdentityRoadmapExitProofVectorsModuleVersionsV2}
	case IdentityRoadmapPhase5ExecutionV2:
		return []IdentityImplementationRoadmapExitIDV2{IdentityRoadmapExitPresigningTargetResolutionV2, IdentityRoadmapExitWalletProofBackedTargetsV2, IdentityRoadmapExitInterfaceServiceHashVerifiedV2}
	case IdentityRoadmapPhase6PerformanceV2:
		return []IdentityImplementationRoadmapExitIDV2{IdentityRoadmapExitIndependentUpdatesParallelizeV2, IdentityRoadmapExitDirectResolutionBoundedLargeStateV2, IdentityRoadmapExitPostSyncProofQueriesV2}
	default:
		return nil
	}
}

func validateIdentityRoadmapEvidenceV2(kind string, id string, evidence []string) error {
	if len(evidence) == 0 {
		return fmt.Errorf("identity implementation roadmap %s %s evidence is required", kind, id)
	}
	sorted := append([]string(nil), evidence...)
	sort.Strings(sorted)
	for i, ref := range sorted {
		if ref == "" || !identityRoadmapEvidenceReferenceHasFunctionV2(ref) {
			return fmt.Errorf("identity implementation roadmap %s %s invalid evidence reference %q", kind, id, ref)
		}
		if ref != evidence[i] {
			return fmt.Errorf("identity implementation roadmap %s %s evidence must be sorted", kind, id)
		}
		if i > 0 && sorted[i-1] == ref {
			return fmt.Errorf("duplicate identity implementation roadmap evidence reference %q", ref)
		}
	}
	return nil
}

func identityRoadmapEvidenceReferenceHasFunctionV2(ref string) bool {
	return strings.Contains(ref, ".go:")
}

func identityRoadmapTasksEqualV2(tasks []IdentityRoadmapTaskV2, required []IdentityImplementationRoadmapTaskIDV2) bool {
	if len(tasks) != len(required) {
		return false
	}
	for i := range required {
		if tasks[i].ID != required[i] {
			return false
		}
	}
	return true
}

func identityRoadmapExitsEqualV2(exits []IdentityRoadmapExitCriterionV2, required []IdentityImplementationRoadmapExitIDV2) bool {
	if len(exits) != len(required) {
		return false
	}
	for i := range required {
		if exits[i].ID != required[i] {
			return false
		}
	}
	return true
}
