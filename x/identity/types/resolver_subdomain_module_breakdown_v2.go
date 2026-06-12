package types

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ResolverModuleStateObjectV2 string
type ResolverModuleMessageNameV2 string
type ResolverModuleQueryNameV2 string
type ResolverModuleFailureModeV2 string
type ResolverModuleIntegrationPointV2 string

const (
	ResolverModulePathV2	= "resolver-module"

	ResolverModuleStateUnifiedResolutionRecord	ResolverModuleStateObjectV2	= "UnifiedResolutionRecord"
	ResolverModuleStateContractTarget		ResolverModuleStateObjectV2	= "ContractTarget"
	ResolverModuleStateServiceEndpoint		ResolverModuleStateObjectV2	= "ServiceEndpoint"
	ResolverModuleStateInterfaceDescriptor		ResolverModuleStateObjectV2	= "InterfaceDescriptor"
	ResolverModuleStateRoutingMetadata		ResolverModuleStateObjectV2	= "RoutingMetadata"
	ResolverModuleStateExecutionHints		ResolverModuleStateObjectV2	= "ExecutionHints"
	ResolverModuleStateReverseResolutionRecord	ResolverModuleStateObjectV2	= "ReverseResolutionRecord"
	ResolverModuleStateResolverParams		ResolverModuleStateObjectV2	= "ResolverParams"

	ResolverModuleMsgSetResolver		ResolverModuleMessageNameV2	= "MsgSetResolver"
	ResolverModuleMsgUpdateResolverRecord	ResolverModuleMessageNameV2	= "MsgUpdateResolverRecord"
	ResolverModuleMsgBatchUpdateResolvers	ResolverModuleMessageNameV2	= "MsgBatchUpdateResolvers"
	ResolverModuleMsgSetReverseRecord	ResolverModuleMessageNameV2	= "MsgSetReverseRecord"
	ResolverModuleMsgVerifyReverseRecord	ResolverModuleMessageNameV2	= "MsgVerifyReverseRecord"
	ResolverModuleMsgClearResolverRecord	ResolverModuleMessageNameV2	= "MsgClearResolverRecord"

	ResolverModuleQueryResolver		ResolverModuleQueryNameV2	= "QueryResolver"
	ResolverModuleQueryResolvePrimary	ResolverModuleQueryNameV2	= "QueryResolvePrimary"
	ResolverModuleQueryResolveTarget	ResolverModuleQueryNameV2	= "QueryResolveTarget"
	ResolverModuleQueryResolveService	ResolverModuleQueryNameV2	= "QueryResolveService"
	ResolverModuleQueryResolveInterface	ResolverModuleQueryNameV2	= "QueryResolveInterface"
	ResolverModuleQueryResolveRoute		ResolverModuleQueryNameV2	= "QueryResolveRoute"
	ResolverModuleQueryReverse		ResolverModuleQueryNameV2	= "QueryReverse"
	ResolverModuleQueryVerifiedReverse	ResolverModuleQueryNameV2	= "QueryVerifiedReverse"

	ResolverModuleFailureUnauthorizedUpdate		ResolverModuleFailureModeV2	= "unauthorized_resolver_update"
	ResolverModuleFailureOversizedPayload		ResolverModuleFailureModeV2	= "oversized_record_payload"
	ResolverModuleFailureStaleExpectedVersion	ResolverModuleFailureModeV2	= "stale_expected_record_version"
	ResolverModuleFailureReverseForwardMismatch	ResolverModuleFailureModeV2	= "reverse_record_forward_mismatch"
	ResolverModuleFailureInterfaceHashMismatch	ResolverModuleFailureModeV2	= "interface_descriptor_hash_mismatch"
	ResolverModuleFailureTTLExceedsDomainExpiry	ResolverModuleFailureModeV2	= "resolver_ttl_exceeds_domain_expiry"

	ResolverModuleIntegrationIdentityCore		ResolverModuleIntegrationPointV2	= "identity_core"
	ResolverModuleIntegrationSubdomainModule	ResolverModuleIntegrationPointV2	= "subdomain_module"
	ResolverModuleIntegrationFeeModule		ResolverModuleIntegrationPointV2	= "fee_module"
	ResolverModuleIntegrationRoutingIntegration	ResolverModuleIntegrationPointV2	= "routing_integration_module"
	ResolverModuleIntegrationStoreV2		ResolverModuleIntegrationPointV2	= "store_v2"
)

type SubdomainModuleStateObjectV2 string
type SubdomainModuleMessageNameV2 string
type SubdomainModuleQueryNameV2 string
type SubdomainModuleFailureModeV2 string
type SubdomainModuleIntegrationPointV2 string

const (
	SubdomainModulePathV2	= "subdomain-module"

	SubdomainModuleStateSubdomainRecord	SubdomainModuleStateObjectV2	= "SubdomainRecord"
	SubdomainModuleStateDelegationRecord	SubdomainModuleStateObjectV2	= "DelegationRecord"
	SubdomainModuleStateZonePolicy		SubdomainModuleStateObjectV2	= "ZonePolicy"
	SubdomainModuleStateSubdomainIndex	SubdomainModuleStateObjectV2	= "SubdomainIndex"
	SubdomainModuleStatePathCommitment	SubdomainModuleStateObjectV2	= "PathCommitment"

	SubdomainModuleMsgCreateSubdomain	SubdomainModuleMessageNameV2	= "MsgCreateSubdomain"
	SubdomainModuleMsgDelegateSubdomain	SubdomainModuleMessageNameV2	= "MsgDelegateSubdomain"
	SubdomainModuleMsgRevokeDelegation	SubdomainModuleMessageNameV2	= "MsgRevokeDelegation"
	SubdomainModuleMsgUpdateZonePolicy	SubdomainModuleMessageNameV2	= "MsgUpdateZonePolicy"
	SubdomainModuleMsgDetachSubdomain	SubdomainModuleMessageNameV2	= "MsgDetachSubdomain"
	SubdomainModuleMsgRenewSubdomain	SubdomainModuleMessageNameV2	= "MsgRenewSubdomain"

	SubdomainModuleQuerySubdomains			SubdomainModuleQueryNameV2	= "QuerySubdomains"
	SubdomainModuleQueryDelegations			SubdomainModuleQueryNameV2	= "QueryDelegations"
	SubdomainModuleQueryZonePolicy			SubdomainModuleQueryNameV2	= "QueryZonePolicy"
	SubdomainModuleQueryRecursivePath		SubdomainModuleQueryNameV2	= "QueryRecursivePath"
	SubdomainModuleQuerySubdomainAuthorization	SubdomainModuleQueryNameV2	= "QuerySubdomainAuthorization"

	SubdomainModuleFailureChildExpiryExceedsParent	SubdomainModuleFailureModeV2	= "child_expiry_exceeds_parent_expiry"
	SubdomainModuleFailureDelegateEscalation	SubdomainModuleFailureModeV2	= "delegate_escalates_permissions"
	SubdomainModuleFailureParentTransferStaleCache	SubdomainModuleFailureModeV2	= "parent_transfer_leaves_child_cache_stale"
	SubdomainModuleFailureInconsistentZonePolicy	SubdomainModuleFailureModeV2	= "zone_policy_inconsistent_child_rules"
	SubdomainModuleFailureDetachedMissingPayment	SubdomainModuleFailureModeV2	= "detached_subdomain_lacks_independent_payment"

	SubdomainModuleIntegrationIdentityCore		SubdomainModuleIntegrationPointV2	= "identity_core"
	SubdomainModuleIntegrationResolverModule	SubdomainModuleIntegrationPointV2	= "resolver_module"
	SubdomainModuleIntegrationFeeModule		SubdomainModuleIntegrationPointV2	= "fee_module"
	SubdomainModuleIntegrationProofModule		SubdomainModuleIntegrationPointV2	= "proof_module"
	SubdomainModuleIntegrationStoreV2		SubdomainModuleIntegrationPointV2	= "store_v2"
)

type ResolverModuleFailureCoverageV2 struct {
	Mode		ResolverModuleFailureModeV2
	Guard		string
	StoreScope	string
}

type ResolverModuleBreakdownV2 struct {
	ModulePath		string
	Purpose			[]string
	StateObjects		[]ResolverModuleStateObjectV2
	Messages		[]ResolverModuleMessageNameV2
	Queries			[]ResolverModuleQueryNameV2
	FailureModes		[]ResolverModuleFailureCoverageV2
	IntegrationPoints	[]ResolverModuleIntegrationPointV2
	BackingPrimitives	[]string
	StoreKeys		[]string
	BreakdownHash		string
}

type SubdomainModuleFailureCoverageV2 struct {
	Mode		SubdomainModuleFailureModeV2
	Guard		string
	StoreScope	string
}

type SubdomainModuleBreakdownV2 struct {
	ModulePath		string
	Purpose			[]string
	StateObjects		[]SubdomainModuleStateObjectV2
	Messages		[]SubdomainModuleMessageNameV2
	Queries			[]SubdomainModuleQueryNameV2
	FailureModes		[]SubdomainModuleFailureCoverageV2
	IntegrationPoints	[]SubdomainModuleIntegrationPointV2
	BackingPrimitives	[]string
	StoreKeys		[]string
	BreakdownHash		string
}

type SubdomainModuleIndexEntryV2 struct {
	ParentName	string
	ChildName	string
	ChildLabel	string
	ParentNameHash	string
	ChildNameHash	string
	StoreKey	string
	Detached	bool
}

type SubdomainModuleIndexReportV2 struct {
	Valid		bool
	Issues		[]string
	Index		[]SubdomainModuleIndexEntryV2
	ReportHash	string
}

func DefaultResolverModuleBreakdownV2() (ResolverModuleBreakdownV2, error) {
	breakdown := ResolverModuleBreakdownV2{
		ModulePath:	ResolverModulePathV2,
		Purpose: []string{
			"contract_target_resolution",
			"execution_hint_resolution",
			"interface_descriptor_resolution",
			"reverse_resolution",
			"routing_metadata_resolution",
			"service_endpoint_resolution",
			"unified_record_validation",
		},
		StateObjects:	requiredResolverModuleStateObjectsV2(),
		Messages:	requiredResolverModuleMessagesV2(),
		Queries:	requiredResolverModuleQueriesV2(),
		FailureModes: []ResolverModuleFailureCoverageV2{
			{Mode: ResolverModuleFailureInterfaceHashMismatch, Guard: "ValidateUnifiedResolutionRecordV2", StoreScope: IdentityStoreV2SpecResolversPrefix},
			{Mode: ResolverModuleFailureOversizedPayload, Guard: "ValidateUnifiedResolverPayloadSafetyV2", StoreScope: IdentityStoreV2SpecResolversPrefix},
			{Mode: ResolverModuleFailureReverseForwardMismatch, Guard: "ValidateReverseResolutionRecordV2", StoreScope: IdentityStoreV2SpecReversePrefix},
			{Mode: ResolverModuleFailureStaleExpectedVersion, Guard: "ValidateResolverRecordVersionForUpdateV2", StoreScope: IdentityStoreV2SpecResolversPrefix},
			{Mode: ResolverModuleFailureTTLExceedsDomainExpiry, Guard: "ValidateResolverModuleRecordV2", StoreScope: IdentityStoreV2SpecResolversPrefix},
			{Mode: ResolverModuleFailureUnauthorizedUpdate, Guard: "ValidateResolverUpdateAuthorizationV2", StoreScope: IdentityStoreV2SpecDelegationsPrefix},
		},
		IntegrationPoints:	requiredResolverModuleIntegrationPointsV2(),
		BackingPrimitives: []string{
			"BuildIdentityResolutionProofFormatV2",
			"BuildUnifiedResolutionRecordV2",
			"ExecuteBatchResolverUpdatesV2",
			"ValidateReverseResolutionRecordV2",
			"ValidateUnifiedResolutionRecordV2",
		},
		StoreKeys: []string{
			IdentityStoreV2SpecDelegationsPrefix,
			IdentityStoreV2SpecResolversPrefix,
			IdentityStoreV2SpecResolverIndexPrefix,
			IdentityStoreV2SpecReversePrefix,
		},
	}
	return NewResolverModuleBreakdownV2(breakdown)
}

func DefaultSubdomainModuleBreakdownV2() (SubdomainModuleBreakdownV2, error) {
	breakdown := SubdomainModuleBreakdownV2{
		ModulePath:	SubdomainModulePathV2,
		Purpose: []string{
			"delegated_partial_permissions",
			"detached_subdomain_lifecycle",
			"hierarchical_subdomain_creation",
			"parent_child_indexes",
			"recursive_path_commitments",
			"zone_policy_enforcement",
		},
		StateObjects:	requiredSubdomainModuleStateObjectsV2(),
		Messages:	requiredSubdomainModuleMessagesV2(),
		Queries:	requiredSubdomainModuleQueriesV2(),
		FailureModes: []SubdomainModuleFailureCoverageV2{
			{Mode: SubdomainModuleFailureChildExpiryExceedsParent, Guard: "ValidateSubdomainCreationV2", StoreScope: IdentityStoreV2SpecSubdomainsPrefix},
			{Mode: SubdomainModuleFailureDelegateEscalation, Guard: "ValidateDelegationDoesNotEscalateV2", StoreScope: IdentityStoreV2SpecDelegationsPrefix},
			{Mode: SubdomainModuleFailureDetachedMissingPayment, Guard: "ValidateSubdomainCreationV2", StoreScope: IdentityStoreV2SpecSubdomainsPrefix},
			{Mode: SubdomainModuleFailureInconsistentZonePolicy, Guard: "ValidateZonePolicyForSubdomainV2", StoreScope: IdentityStoreV2SpecSubdomainsPrefix},
			{Mode: SubdomainModuleFailureParentTransferStaleCache, Guard: "ValidateResolutionCacheRecordV2Use", StoreScope: IdentityStoreV2SpecResolutionCachePrefix},
		},
		IntegrationPoints:	requiredSubdomainModuleIntegrationPointsV2(),
		BackingPrimitives: []string{
			"BuildIdentityPathCommitmentV2",
			"BuildRecursivePolicyProofV2",
			"IssueSubdomainV2",
			"ValidateDelegationRecordV2Use",
			"ValidateZonePolicyV2",
		},
		StoreKeys: []string{
			IdentityStoreV2SpecDelegationsPrefix,
			IdentityStoreV2SpecResolutionCachePrefix,
			IdentityStoreV2SpecSubdomainsPrefix,
		},
	}
	return NewSubdomainModuleBreakdownV2(breakdown)
}

func NewResolverModuleBreakdownV2(breakdown ResolverModuleBreakdownV2) (ResolverModuleBreakdownV2, error) {
	breakdown = canonicalResolverModuleBreakdownV2(breakdown)
	if err := breakdown.ValidateFormat(); err != nil {
		return ResolverModuleBreakdownV2{}, err
	}
	breakdown.BreakdownHash = ComputeResolverModuleBreakdownHashV2(breakdown)
	return breakdown, breakdown.Validate()
}

func NewSubdomainModuleBreakdownV2(breakdown SubdomainModuleBreakdownV2) (SubdomainModuleBreakdownV2, error) {
	breakdown = canonicalSubdomainModuleBreakdownV2(breakdown)
	if err := breakdown.ValidateFormat(); err != nil {
		return SubdomainModuleBreakdownV2{}, err
	}
	breakdown.BreakdownHash = ComputeSubdomainModuleBreakdownHashV2(breakdown)
	return breakdown, breakdown.Validate()
}

func (breakdown ResolverModuleBreakdownV2) ValidateFormat() error {
	if breakdown.ModulePath != ResolverModulePathV2 {
		return errors.New("resolver module breakdown must describe resolver-module")
	}
	if err := validateBreakdownTokenSetV2("resolver purpose", breakdown.Purpose, nil); err != nil {
		return err
	}
	if err := validateResolverModuleEnumSetV2("state object", breakdown.StateObjects, requiredResolverModuleStateObjectsV2(), IsResolverModuleStateObjectV2); err != nil {
		return err
	}
	if err := validateResolverModuleEnumSetV2("message", breakdown.Messages, requiredResolverModuleMessagesV2(), IsResolverModuleMessageNameV2); err != nil {
		return err
	}
	if err := validateResolverModuleEnumSetV2("query", breakdown.Queries, requiredResolverModuleQueriesV2(), IsResolverModuleQueryNameV2); err != nil {
		return err
	}
	if err := validateResolverModuleFailuresV2(breakdown.FailureModes); err != nil {
		return err
	}
	if err := validateResolverModuleEnumSetV2("integration", breakdown.IntegrationPoints, requiredResolverModuleIntegrationPointsV2(), IsResolverModuleIntegrationPointV2); err != nil {
		return err
	}
	if err := validateBreakdownTokenSetV2("resolver backing primitive", breakdown.BackingPrimitives, []string{"BuildIdentityResolutionProofFormatV2", "BuildUnifiedResolutionRecordV2", "ExecuteBatchResolverUpdatesV2", "ValidateReverseResolutionRecordV2", "ValidateUnifiedResolutionRecordV2"}); err != nil {
		return err
	}
	if err := validateBreakdownStoreKeysV2("resolver", breakdown.StoreKeys, []string{IdentityStoreV2SpecDelegationsPrefix, IdentityStoreV2SpecResolversPrefix, IdentityStoreV2SpecResolverIndexPrefix, IdentityStoreV2SpecReversePrefix}); err != nil {
		return err
	}
	if breakdown.BreakdownHash != "" {
		return validateHexHash("resolver module breakdown hash", breakdown.BreakdownHash)
	}
	return nil
}

func (breakdown ResolverModuleBreakdownV2) Validate() error {
	breakdown = canonicalResolverModuleBreakdownV2(breakdown)
	if err := breakdown.ValidateFormat(); err != nil {
		return err
	}
	if breakdown.BreakdownHash == "" {
		return errors.New("resolver module breakdown hash is required")
	}
	if breakdown.BreakdownHash != ComputeResolverModuleBreakdownHashV2(breakdown) {
		return errors.New("resolver module breakdown hash mismatch")
	}
	return nil
}

func (breakdown SubdomainModuleBreakdownV2) ValidateFormat() error {
	if breakdown.ModulePath != SubdomainModulePathV2 {
		return errors.New("subdomain module breakdown must describe subdomain-module")
	}
	if err := validateBreakdownTokenSetV2("subdomain purpose", breakdown.Purpose, nil); err != nil {
		return err
	}
	if err := validateSubdomainModuleEnumSetV2("state object", breakdown.StateObjects, requiredSubdomainModuleStateObjectsV2(), IsSubdomainModuleStateObjectV2); err != nil {
		return err
	}
	if err := validateSubdomainModuleEnumSetV2("message", breakdown.Messages, requiredSubdomainModuleMessagesV2(), IsSubdomainModuleMessageNameV2); err != nil {
		return err
	}
	if err := validateSubdomainModuleEnumSetV2("query", breakdown.Queries, requiredSubdomainModuleQueriesV2(), IsSubdomainModuleQueryNameV2); err != nil {
		return err
	}
	if err := validateSubdomainModuleFailuresV2(breakdown.FailureModes); err != nil {
		return err
	}
	if err := validateSubdomainModuleEnumSetV2("integration", breakdown.IntegrationPoints, requiredSubdomainModuleIntegrationPointsV2(), IsSubdomainModuleIntegrationPointV2); err != nil {
		return err
	}
	if err := validateBreakdownTokenSetV2("subdomain backing primitive", breakdown.BackingPrimitives, []string{"BuildIdentityPathCommitmentV2", "BuildRecursivePolicyProofV2", "IssueSubdomainV2", "ValidateDelegationRecordV2Use", "ValidateZonePolicyV2"}); err != nil {
		return err
	}
	if err := validateBreakdownStoreKeysV2("subdomain", breakdown.StoreKeys, []string{IdentityStoreV2SpecDelegationsPrefix, IdentityStoreV2SpecResolutionCachePrefix, IdentityStoreV2SpecSubdomainsPrefix}); err != nil {
		return err
	}
	if breakdown.BreakdownHash != "" {
		return validateHexHash("subdomain module breakdown hash", breakdown.BreakdownHash)
	}
	return nil
}

func (breakdown SubdomainModuleBreakdownV2) Validate() error {
	breakdown = canonicalSubdomainModuleBreakdownV2(breakdown)
	if err := breakdown.ValidateFormat(); err != nil {
		return err
	}
	if breakdown.BreakdownHash == "" {
		return errors.New("subdomain module breakdown hash is required")
	}
	if breakdown.BreakdownHash != ComputeSubdomainModuleBreakdownHashV2(breakdown) {
		return errors.New("subdomain module breakdown hash mismatch")
	}
	return nil
}

func ValidateResolverModuleRecordV2(state IdentityState, name string, record UnifiedResolutionRecordV2, height uint64) error {
	if err := ValidateUnifiedResolutionRecordV2(record); err != nil {
		return err
	}
	domain, err := requireActiveDomain(state, name, height)
	if err != nil {
		return err
	}
	nameHash, err := DomainRecordV2NameHash(domain.Name)
	if err != nil {
		return err
	}
	if record.NameHash != nameHash {
		return errors.New("resolver module record name_hash does not match identity core domain")
	}
	if !bytes.Equal(record.Owner, domain.Owner) {
		return fmt.Errorf("%s: record owner must match identity core owner", ResolverModuleFailureUnauthorizedUpdate)
	}
	if record.RecordTTL == 0 {
		return errors.New("resolver module record ttl is required")
	}
	if record.UpdatedAtHeight+record.RecordTTL > domain.ExpiryHeight {
		return fmt.Errorf("%s: ttl height %d exceeds domain expiry %d", ResolverModuleFailureTTLExceedsDomainExpiry, record.UpdatedAtHeight+record.RecordTTL, domain.ExpiryHeight)
	}
	if EstimateUnifiedResolverPayloadBytesV2(record) > record.MaxPayloadBytes {
		return fmt.Errorf("%s: payload exceeds max_payload_bytes", ResolverModuleFailureOversizedPayload)
	}
	return nil
}

func ExecuteResolverModuleBatchUpdateV2(state IdentityState, msg MsgBatchUpdateResolversV2, options IdentityBatchResolverUpdateOptionsV2) (IdentityState, IdentityBatchResolverUpdateResponseV2, error) {
	next, response, err := ExecuteBatchResolverUpdatesV2(state, msg, options)
	if validationErr := ValidateBatchResolverUpdateResponseV2(response); validationErr != nil {
		if err != nil {
			return next, response, fmt.Errorf("%w; response validation: %v", err, validationErr)
		}
		return next, response, validationErr
	}
	return next, response, err
}

func BuildResolverModuleProofQueryV2(state IdentityState, chainID string, appHash string, name string, queryType IdentityProofQueryTypeV2, height uint64, ttl uint64, reverseAddress sdk.AccAddress) (IdentityResolutionProofFormatV2, error) {
	switch queryType {
	case IdentityProofQueryResolvePrimary, IdentityProofQueryResolveRecord, IdentityProofQueryResolveReverse:
	default:
		return IdentityResolutionProofFormatV2{}, fmt.Errorf("resolver module unsupported proof query type %q", queryType)
	}
	return BuildIdentityResolutionProofFormatV2(state, chainID, appHash, name, queryType, height, ttl, reverseAddress)
}

func BuildSubdomainModuleIndexV2(state IdentityState) (SubdomainModuleIndexReportV2, error) {
	if err := state.Validate(); err != nil {
		return SubdomainModuleIndexReportV2{}, err
	}
	report := SubdomainModuleIndexReportV2{Valid: true}
	for _, subdomain := range state.Subdomains {
		parentHash, err := DomainRecordV2NameHash(subdomain.ParentName)
		if err != nil {
			return SubdomainModuleIndexReportV2{}, err
		}
		childHash, err := DomainRecordV2NameHash(subdomain.Name)
		if err != nil {
			return SubdomainModuleIndexReportV2{}, err
		}
		label, err := subdomainChildLabelV2(subdomain.ParentName, subdomain.Name)
		if err != nil {
			report.addIssue(err.Error())
			continue
		}
		key, err := IdentityStoreV2SpecSubdomainKey(subdomain.ParentName, label)
		if err != nil {
			report.addIssue(err.Error())
			continue
		}
		report.Index = append(report.Index, SubdomainModuleIndexEntryV2{
			ParentName:	subdomain.ParentName,
			ChildName:	subdomain.Name,
			ChildLabel:	label,
			ParentNameHash:	parentHash,
			ChildNameHash:	childHash,
			StoreKey:	key,
			Detached:	subdomain.Detached,
		})
	}
	sort.Slice(report.Index, func(i, j int) bool { return report.Index[i].StoreKey < report.Index[j].StoreKey })
	report.finalize()
	return report, nil
}

func ValidateSubdomainModulePathPolicyV2(rootName string, targetName string, sourceVersion uint64, parentEpoch uint64, childEpoch uint64, policies []ZonePolicyV2) (IdentityPathCommitmentV2, RecursivePolicyProofV2, error) {
	path, err := CanonicalResolutionPathV2(targetName)
	if err != nil {
		return IdentityPathCommitmentV2{}, RecursivePolicyProofV2{}, err
	}
	if len(path.Path) == 0 || path.Path[0] != rootName {
		return IdentityPathCommitmentV2{}, RecursivePolicyProofV2{}, errors.New("subdomain module recursive path root mismatch")
	}
	commitment, err := BuildIdentityPathCommitmentV2(path, sourceVersion, parentEpoch, childEpoch)
	if err != nil {
		return IdentityPathCommitmentV2{}, RecursivePolicyProofV2{}, err
	}
	proof, err := BuildRecursivePolicyProofV2(rootName, targetName, commitment, policies)
	if err != nil {
		return IdentityPathCommitmentV2{}, RecursivePolicyProofV2{}, err
	}
	return commitment, proof, nil
}

func ComputeResolverModuleBreakdownHashV2(breakdown ResolverModuleBreakdownV2) string {
	breakdown = canonicalResolverModuleBreakdownV2(breakdown)
	parts := []string{"aetra-resolver-module-breakdown-v2", breakdown.ModulePath}
	parts = appendBreakdownStringsV2(parts, "purpose", breakdown.Purpose)
	for _, value := range breakdown.StateObjects {
		parts = append(parts, "state", string(value))
	}
	for _, value := range breakdown.Messages {
		parts = append(parts, "message", string(value))
	}
	for _, value := range breakdown.Queries {
		parts = append(parts, "query", string(value))
	}
	for _, failure := range breakdown.FailureModes {
		parts = append(parts, "failure", string(failure.Mode), failure.Guard, failure.StoreScope)
	}
	for _, value := range breakdown.IntegrationPoints {
		parts = append(parts, "integration", string(value))
	}
	parts = appendBreakdownStringsV2(parts, "primitive", breakdown.BackingPrimitives)
	parts = appendBreakdownStringsV2(parts, "store", breakdown.StoreKeys)
	return identityHash(parts...)
}

func ComputeSubdomainModuleBreakdownHashV2(breakdown SubdomainModuleBreakdownV2) string {
	breakdown = canonicalSubdomainModuleBreakdownV2(breakdown)
	parts := []string{"aetra-subdomain-module-breakdown-v2", breakdown.ModulePath}
	parts = appendBreakdownStringsV2(parts, "purpose", breakdown.Purpose)
	for _, value := range breakdown.StateObjects {
		parts = append(parts, "state", string(value))
	}
	for _, value := range breakdown.Messages {
		parts = append(parts, "message", string(value))
	}
	for _, value := range breakdown.Queries {
		parts = append(parts, "query", string(value))
	}
	for _, failure := range breakdown.FailureModes {
		parts = append(parts, "failure", string(failure.Mode), failure.Guard, failure.StoreScope)
	}
	for _, value := range breakdown.IntegrationPoints {
		parts = append(parts, "integration", string(value))
	}
	parts = appendBreakdownStringsV2(parts, "primitive", breakdown.BackingPrimitives)
	parts = appendBreakdownStringsV2(parts, "store", breakdown.StoreKeys)
	return identityHash(parts...)
}

func ComputeSubdomainModuleIndexReportHashV2(report SubdomainModuleIndexReportV2) string {
	sort.Strings(report.Issues)
	sort.Slice(report.Index, func(i, j int) bool { return report.Index[i].StoreKey < report.Index[j].StoreKey })
	parts := []string{"aetra-subdomain-module-index-report-v2", fmt.Sprintf("valid=%t", report.Valid)}
	parts = append(parts, report.Issues...)
	for _, entry := range report.Index {
		parts = append(parts, entry.ParentName, entry.ChildName, entry.ChildLabel, entry.ParentNameHash, entry.ChildNameHash, entry.StoreKey, fmt.Sprintf("detached=%t", entry.Detached))
	}
	return identityHash(parts...)
}

func IsResolverModuleStateObjectV2(value ResolverModuleStateObjectV2) bool {
	switch value {
	case ResolverModuleStateUnifiedResolutionRecord, ResolverModuleStateContractTarget, ResolverModuleStateServiceEndpoint, ResolverModuleStateInterfaceDescriptor, ResolverModuleStateRoutingMetadata, ResolverModuleStateExecutionHints, ResolverModuleStateReverseResolutionRecord, ResolverModuleStateResolverParams:
		return true
	default:
		return false
	}
}

func IsResolverModuleMessageNameV2(value ResolverModuleMessageNameV2) bool {
	switch value {
	case ResolverModuleMsgSetResolver, ResolverModuleMsgUpdateResolverRecord, ResolverModuleMsgBatchUpdateResolvers, ResolverModuleMsgSetReverseRecord, ResolverModuleMsgVerifyReverseRecord, ResolverModuleMsgClearResolverRecord:
		return true
	default:
		return false
	}
}

func IsResolverModuleQueryNameV2(value ResolverModuleQueryNameV2) bool {
	switch value {
	case ResolverModuleQueryResolver, ResolverModuleQueryResolvePrimary, ResolverModuleQueryResolveTarget, ResolverModuleQueryResolveService, ResolverModuleQueryResolveInterface, ResolverModuleQueryResolveRoute, ResolverModuleQueryReverse, ResolverModuleQueryVerifiedReverse:
		return true
	default:
		return false
	}
}

func IsResolverModuleFailureModeV2(value ResolverModuleFailureModeV2) bool {
	switch value {
	case ResolverModuleFailureUnauthorizedUpdate, ResolverModuleFailureOversizedPayload, ResolverModuleFailureStaleExpectedVersion, ResolverModuleFailureReverseForwardMismatch, ResolverModuleFailureInterfaceHashMismatch, ResolverModuleFailureTTLExceedsDomainExpiry:
		return true
	default:
		return false
	}
}

func IsResolverModuleIntegrationPointV2(value ResolverModuleIntegrationPointV2) bool {
	switch value {
	case ResolverModuleIntegrationIdentityCore, ResolverModuleIntegrationSubdomainModule, ResolverModuleIntegrationFeeModule, ResolverModuleIntegrationRoutingIntegration, ResolverModuleIntegrationStoreV2:
		return true
	default:
		return false
	}
}

func IsSubdomainModuleStateObjectV2(value SubdomainModuleStateObjectV2) bool {
	switch value {
	case SubdomainModuleStateSubdomainRecord, SubdomainModuleStateDelegationRecord, SubdomainModuleStateZonePolicy, SubdomainModuleStateSubdomainIndex, SubdomainModuleStatePathCommitment:
		return true
	default:
		return false
	}
}

func IsSubdomainModuleMessageNameV2(value SubdomainModuleMessageNameV2) bool {
	switch value {
	case SubdomainModuleMsgCreateSubdomain, SubdomainModuleMsgDelegateSubdomain, SubdomainModuleMsgRevokeDelegation, SubdomainModuleMsgUpdateZonePolicy, SubdomainModuleMsgDetachSubdomain, SubdomainModuleMsgRenewSubdomain:
		return true
	default:
		return false
	}
}

func IsSubdomainModuleQueryNameV2(value SubdomainModuleQueryNameV2) bool {
	switch value {
	case SubdomainModuleQuerySubdomains, SubdomainModuleQueryDelegations, SubdomainModuleQueryZonePolicy, SubdomainModuleQueryRecursivePath, SubdomainModuleQuerySubdomainAuthorization:
		return true
	default:
		return false
	}
}

func IsSubdomainModuleFailureModeV2(value SubdomainModuleFailureModeV2) bool {
	switch value {
	case SubdomainModuleFailureChildExpiryExceedsParent, SubdomainModuleFailureDelegateEscalation, SubdomainModuleFailureParentTransferStaleCache, SubdomainModuleFailureInconsistentZonePolicy, SubdomainModuleFailureDetachedMissingPayment:
		return true
	default:
		return false
	}
}

func IsSubdomainModuleIntegrationPointV2(value SubdomainModuleIntegrationPointV2) bool {
	switch value {
	case SubdomainModuleIntegrationIdentityCore, SubdomainModuleIntegrationResolverModule, SubdomainModuleIntegrationFeeModule, SubdomainModuleIntegrationProofModule, SubdomainModuleIntegrationStoreV2:
		return true
	default:
		return false
	}
}

func requiredResolverModuleStateObjectsV2() []ResolverModuleStateObjectV2 {
	return []ResolverModuleStateObjectV2{ResolverModuleStateContractTarget, ResolverModuleStateExecutionHints, ResolverModuleStateInterfaceDescriptor, ResolverModuleStateResolverParams, ResolverModuleStateReverseResolutionRecord, ResolverModuleStateRoutingMetadata, ResolverModuleStateServiceEndpoint, ResolverModuleStateUnifiedResolutionRecord}
}

func requiredResolverModuleMessagesV2() []ResolverModuleMessageNameV2 {
	return []ResolverModuleMessageNameV2{ResolverModuleMsgBatchUpdateResolvers, ResolverModuleMsgClearResolverRecord, ResolverModuleMsgSetResolver, ResolverModuleMsgSetReverseRecord, ResolverModuleMsgUpdateResolverRecord, ResolverModuleMsgVerifyReverseRecord}
}

func requiredResolverModuleQueriesV2() []ResolverModuleQueryNameV2 {
	return []ResolverModuleQueryNameV2{ResolverModuleQueryResolver, ResolverModuleQueryResolveInterface, ResolverModuleQueryResolvePrimary, ResolverModuleQueryResolveRoute, ResolverModuleQueryResolveService, ResolverModuleQueryResolveTarget, ResolverModuleQueryReverse, ResolverModuleQueryVerifiedReverse}
}

func requiredResolverModuleFailuresV2() []ResolverModuleFailureModeV2 {
	return []ResolverModuleFailureModeV2{ResolverModuleFailureInterfaceHashMismatch, ResolverModuleFailureOversizedPayload, ResolverModuleFailureReverseForwardMismatch, ResolverModuleFailureStaleExpectedVersion, ResolverModuleFailureTTLExceedsDomainExpiry, ResolverModuleFailureUnauthorizedUpdate}
}

func requiredResolverModuleIntegrationPointsV2() []ResolverModuleIntegrationPointV2 {
	return []ResolverModuleIntegrationPointV2{ResolverModuleIntegrationFeeModule, ResolverModuleIntegrationIdentityCore, ResolverModuleIntegrationRoutingIntegration, ResolverModuleIntegrationStoreV2, ResolverModuleIntegrationSubdomainModule}
}

func requiredSubdomainModuleStateObjectsV2() []SubdomainModuleStateObjectV2 {
	return []SubdomainModuleStateObjectV2{SubdomainModuleStateDelegationRecord, SubdomainModuleStatePathCommitment, SubdomainModuleStateSubdomainIndex, SubdomainModuleStateSubdomainRecord, SubdomainModuleStateZonePolicy}
}

func requiredSubdomainModuleMessagesV2() []SubdomainModuleMessageNameV2 {
	return []SubdomainModuleMessageNameV2{SubdomainModuleMsgCreateSubdomain, SubdomainModuleMsgDelegateSubdomain, SubdomainModuleMsgDetachSubdomain, SubdomainModuleMsgRenewSubdomain, SubdomainModuleMsgRevokeDelegation, SubdomainModuleMsgUpdateZonePolicy}
}

func requiredSubdomainModuleQueriesV2() []SubdomainModuleQueryNameV2 {
	return []SubdomainModuleQueryNameV2{SubdomainModuleQueryDelegations, SubdomainModuleQueryRecursivePath, SubdomainModuleQuerySubdomainAuthorization, SubdomainModuleQuerySubdomains, SubdomainModuleQueryZonePolicy}
}

func requiredSubdomainModuleFailuresV2() []SubdomainModuleFailureModeV2 {
	return []SubdomainModuleFailureModeV2{SubdomainModuleFailureChildExpiryExceedsParent, SubdomainModuleFailureDelegateEscalation, SubdomainModuleFailureDetachedMissingPayment, SubdomainModuleFailureInconsistentZonePolicy, SubdomainModuleFailureParentTransferStaleCache}
}

func requiredSubdomainModuleIntegrationPointsV2() []SubdomainModuleIntegrationPointV2 {
	return []SubdomainModuleIntegrationPointV2{SubdomainModuleIntegrationFeeModule, SubdomainModuleIntegrationIdentityCore, SubdomainModuleIntegrationProofModule, SubdomainModuleIntegrationResolverModule, SubdomainModuleIntegrationStoreV2}
}

func canonicalResolverModuleBreakdownV2(breakdown ResolverModuleBreakdownV2) ResolverModuleBreakdownV2 {
	breakdown.ModulePath = strings.TrimSpace(breakdown.ModulePath)
	breakdown.Purpose = sortedBreakdownStringsV2(breakdown.Purpose)
	breakdown.StateObjects = sortedBreakdownTypedStringsV2(breakdown.StateObjects)
	breakdown.Messages = sortedBreakdownTypedStringsV2(breakdown.Messages)
	breakdown.Queries = sortedBreakdownTypedStringsV2(breakdown.Queries)
	breakdown.FailureModes = sortedResolverModuleFailuresV2(breakdown.FailureModes)
	breakdown.IntegrationPoints = sortedBreakdownTypedStringsV2(breakdown.IntegrationPoints)
	breakdown.BackingPrimitives = sortedBreakdownStringsV2(breakdown.BackingPrimitives)
	breakdown.StoreKeys = sortedBreakdownStringsV2(breakdown.StoreKeys)
	breakdown.BreakdownHash = strings.TrimSpace(breakdown.BreakdownHash)
	return breakdown
}

func canonicalSubdomainModuleBreakdownV2(breakdown SubdomainModuleBreakdownV2) SubdomainModuleBreakdownV2 {
	breakdown.ModulePath = strings.TrimSpace(breakdown.ModulePath)
	breakdown.Purpose = sortedBreakdownStringsV2(breakdown.Purpose)
	breakdown.StateObjects = sortedBreakdownTypedStringsV2(breakdown.StateObjects)
	breakdown.Messages = sortedBreakdownTypedStringsV2(breakdown.Messages)
	breakdown.Queries = sortedBreakdownTypedStringsV2(breakdown.Queries)
	breakdown.FailureModes = sortedSubdomainModuleFailuresV2(breakdown.FailureModes)
	breakdown.IntegrationPoints = sortedBreakdownTypedStringsV2(breakdown.IntegrationPoints)
	breakdown.BackingPrimitives = sortedBreakdownStringsV2(breakdown.BackingPrimitives)
	breakdown.StoreKeys = sortedBreakdownStringsV2(breakdown.StoreKeys)
	breakdown.BreakdownHash = strings.TrimSpace(breakdown.BreakdownHash)
	return breakdown
}

func validateResolverModuleFailuresV2(values []ResolverModuleFailureCoverageV2) error {
	if len(values) != len(requiredResolverModuleFailuresV2()) {
		return fmt.Errorf("resolver module expected %d failure modes", len(requiredResolverModuleFailuresV2()))
	}
	seen := map[ResolverModuleFailureModeV2]struct{}{}
	for _, value := range values {
		if !IsResolverModuleFailureModeV2(value.Mode) {
			return fmt.Errorf("resolver module unknown failure mode %q", value.Mode)
		}
		if strings.TrimSpace(value.Guard) == "" || !strings.HasPrefix(value.StoreScope, IdentityStoreV2Prefix+"/") {
			return fmt.Errorf("resolver module failure %s has invalid guard or store scope", value.Mode)
		}
		if _, found := seen[value.Mode]; found {
			return fmt.Errorf("resolver module duplicate failure mode %s", value.Mode)
		}
		seen[value.Mode] = struct{}{}
	}
	for _, required := range requiredResolverModuleFailuresV2() {
		if _, found := seen[required]; !found {
			return fmt.Errorf("resolver module missing failure mode %s", required)
		}
	}
	return nil
}

func validateSubdomainModuleFailuresV2(values []SubdomainModuleFailureCoverageV2) error {
	if len(values) != len(requiredSubdomainModuleFailuresV2()) {
		return fmt.Errorf("subdomain module expected %d failure modes", len(requiredSubdomainModuleFailuresV2()))
	}
	seen := map[SubdomainModuleFailureModeV2]struct{}{}
	for _, value := range values {
		if !IsSubdomainModuleFailureModeV2(value.Mode) {
			return fmt.Errorf("subdomain module unknown failure mode %q", value.Mode)
		}
		if strings.TrimSpace(value.Guard) == "" || !strings.HasPrefix(value.StoreScope, IdentityStoreV2Prefix+"/") {
			return fmt.Errorf("subdomain module failure %s has invalid guard or store scope", value.Mode)
		}
		if _, found := seen[value.Mode]; found {
			return fmt.Errorf("subdomain module duplicate failure mode %s", value.Mode)
		}
		seen[value.Mode] = struct{}{}
	}
	for _, required := range requiredSubdomainModuleFailuresV2() {
		if _, found := seen[required]; !found {
			return fmt.Errorf("subdomain module missing failure mode %s", required)
		}
	}
	return nil
}

func validateResolverModuleEnumSetV2[T ~string](label string, values []T, required []T, allowed func(T) bool) error {
	return validateModuleEnumSetV2("resolver module", label, values, required, allowed)
}

func validateSubdomainModuleEnumSetV2[T ~string](label string, values []T, required []T, allowed func(T) bool) error {
	return validateModuleEnumSetV2("subdomain module", label, values, required, allowed)
}

func validateModuleEnumSetV2[T ~string](module string, label string, values []T, required []T, allowed func(T) bool) error {
	if len(values) != len(required) {
		return fmt.Errorf("%s expected %d %s entries", module, len(required), label)
	}
	seen := map[T]struct{}{}
	previous := ""
	for _, value := range values {
		if !allowed(value) {
			return fmt.Errorf("%s unknown %s %q", module, label, value)
		}
		if _, found := seen[value]; found {
			return fmt.Errorf("%s duplicate %s %q", module, label, value)
		}
		if previous != "" && previous >= string(value) {
			return fmt.Errorf("%s %s entries must be sorted", module, label)
		}
		seen[value] = struct{}{}
		previous = string(value)
	}
	for _, requiredValue := range required {
		if _, found := seen[requiredValue]; !found {
			return fmt.Errorf("%s missing %s %q", module, label, requiredValue)
		}
	}
	return nil
}

func validateBreakdownTokenSetV2(label string, values []string, required []string) error {
	if len(values) == 0 {
		return fmt.Errorf("%s entries are required", label)
	}
	if len(required) > 0 && len(values) != len(required) {
		return fmt.Errorf("%s expected %d entries", label, len(required))
	}
	seen := map[string]struct{}{}
	previous := ""
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			return fmt.Errorf("%s entry is required", label)
		}
		if _, found := seen[value]; found {
			return fmt.Errorf("%s duplicate entry %q", label, value)
		}
		if previous != "" && previous >= value {
			return fmt.Errorf("%s entries must be sorted", label)
		}
		seen[value] = struct{}{}
		previous = value
	}
	for _, requiredValue := range required {
		if _, found := seen[requiredValue]; !found {
			return fmt.Errorf("%s missing entry %q", label, requiredValue)
		}
	}
	return nil
}

func validateBreakdownStoreKeysV2(module string, values []string, required []string) error {
	if err := validateBreakdownTokenSetV2(module+" store key", values, required); err != nil {
		return err
	}
	for _, value := range values {
		if !strings.HasPrefix(value, IdentityStoreV2Prefix+"/") {
			return fmt.Errorf("%s store key %s must be under Store v2 identity prefix", module, value)
		}
	}
	return nil
}

func sortedResolverModuleFailuresV2(values []ResolverModuleFailureCoverageV2) []ResolverModuleFailureCoverageV2 {
	out := append([]ResolverModuleFailureCoverageV2(nil), values...)
	for i := range out {
		out[i].Guard = strings.TrimSpace(out[i].Guard)
		out[i].StoreScope = strings.TrimSpace(out[i].StoreScope)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Mode < out[j].Mode })
	return out
}

func sortedSubdomainModuleFailuresV2(values []SubdomainModuleFailureCoverageV2) []SubdomainModuleFailureCoverageV2 {
	out := append([]SubdomainModuleFailureCoverageV2(nil), values...)
	for i := range out {
		out[i].Guard = strings.TrimSpace(out[i].Guard)
		out[i].StoreScope = strings.TrimSpace(out[i].StoreScope)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Mode < out[j].Mode })
	return out
}

func sortedBreakdownTypedStringsV2[T ~string](values []T) []T {
	out := append([]T(nil), values...)
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func sortedBreakdownStringsV2(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, strings.TrimSpace(value))
	}
	sort.Strings(out)
	return out
}

func appendBreakdownStringsV2(parts []string, label string, values []string) []string {
	parts = append(parts, label, fmt.Sprint(len(values)))
	return append(parts, values...)
}

func subdomainChildLabelV2(parentName string, childName string) (string, error) {
	parent, err := NormalizeAETDomain(parentName)
	if err != nil {
		return "", err
	}
	child, err := NormalizeAETDomain(childName)
	if err != nil {
		return "", err
	}
	suffix := "." + parent
	if !strings.HasSuffix(child, suffix) {
		return "", errors.New("subdomain module child is not below parent")
	}
	label := strings.TrimSuffix(child, suffix)
	if strings.Contains(label, ".") {
		return "", errors.New("subdomain module index requires immediate child label")
	}
	return label, validateDomainLabel(label)
}

func (report *SubdomainModuleIndexReportV2) addIssue(issue string) {
	report.Valid = false
	report.Issues = append(report.Issues, issue)
}

func (report *SubdomainModuleIndexReportV2) finalize() {
	sort.Strings(report.Issues)
	report.Valid = len(report.Issues) == 0
	report.ReportHash = ComputeSubdomainModuleIndexReportHashV2(*report)
}

func ResolverModuleReverseStoreKeyV2(record ReverseResolutionRecordV2) (string, error) {
	if err := ValidateReverseResolutionRecordV2Format(record); err != nil {
		return "", err
	}
	return IdentityStoreV2SpecReverseKey(record.Address)
}

func SubdomainModuleDelegationStoreKeyV2(name string, delegate sdk.AccAddress, scope DelegationScopeV2) (string, error) {
	return IdentityStoreV2SpecDelegationKey(name, delegate, scope)
}

func SubdomainModuleOwnerHexV2(owner sdk.AccAddress) (string, error) {
	if err := validateSpecAddress("subdomain module owner", owner); err != nil {
		return "", err
	}
	return hex.EncodeToString(owner), nil
}
