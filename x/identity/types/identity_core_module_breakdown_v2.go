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

type IdentityCoreStateObjectV2 string
type IdentityCoreMessageNameV2 string
type IdentityCoreQueryNameV2 string
type IdentityCoreFailureModeV2 string
type IdentityCoreIntegrationPointV2 string

const (
	IdentityCoreModulePathV2	= "identity-core-module"

	IdentityCoreStateDomainRecord		IdentityCoreStateObjectV2	= "DomainRecord"
	IdentityCoreStateDomainCommitment	IdentityCoreStateObjectV2	= "DomainCommitment"
	IdentityCoreStateDomainNFTBinding	IdentityCoreStateObjectV2	= "DomainNFTBinding"
	IdentityCoreStateDomainLifecycleParams	IdentityCoreStateObjectV2	= "DomainLifecycleParams"
	IdentityCoreStateExpiryIndex		IdentityCoreStateObjectV2	= "ExpiryIndex"
	IdentityCoreStateOwnerIndex		IdentityCoreStateObjectV2	= "OwnerIndex"

	IdentityCoreMsgCommitRegistration	IdentityCoreMessageNameV2	= "MsgCommitRegistration"
	IdentityCoreMsgRevealRegistration	IdentityCoreMessageNameV2	= "MsgRevealRegistration"
	IdentityCoreMsgRegisterDirect		IdentityCoreMessageNameV2	= "MsgRegisterDirect"
	IdentityCoreMsgRenewDomain		IdentityCoreMessageNameV2	= "MsgRenewDomain"
	IdentityCoreMsgTransferDomain		IdentityCoreMessageNameV2	= "MsgTransferDomain"
	IdentityCoreMsgExpireDomain		IdentityCoreMessageNameV2	= "MsgExpireDomain"
	IdentityCoreMsgRepairNFTBinding		IdentityCoreMessageNameV2	= "MsgRepairNFTBinding"

	IdentityCoreQueryDomain			IdentityCoreQueryNameV2	= "QueryDomain"
	IdentityCoreQueryDomainByName		IdentityCoreQueryNameV2	= "QueryDomainByName"
	IdentityCoreQueryDomainsByOwner		IdentityCoreQueryNameV2	= "QueryDomainsByOwner"
	IdentityCoreQueryDomainNFTBinding	IdentityCoreQueryNameV2	= "QueryDomainNFTBinding"
	IdentityCoreQueryDomainLifecycle	IdentityCoreQueryNameV2	= "QueryDomainLifecycle"
	IdentityCoreQueryRegistrationPrice	IdentityCoreQueryNameV2	= "QueryRegistrationPrice"
	IdentityCoreQueryRenewalPrice		IdentityCoreQueryNameV2	= "QueryRenewalPrice"

	IdentityCoreFailureNFTOwnerMismatch		IdentityCoreFailureModeV2	= "registry_nft_owner_mismatch"
	IdentityCoreFailureMissingExpiryIndex		IdentityCoreFailureModeV2	= "expiry_index_missing_domain"
	IdentityCoreFailureCommitmentReplay		IdentityCoreFailureModeV2	= "commitment_replay"
	IdentityCoreFailureTransferResolverRace		IdentityCoreFailureModeV2	= "transfer_race_with_resolver_update"
	IdentityCoreFailureInvalidLifecycleChange	IdentityCoreFailureModeV2	= "incorrect_lifecycle_transition"

	IdentityCoreIntegrationNFTModule	IdentityCoreIntegrationPointV2	= "nft_module"
	IdentityCoreIntegrationBankModule	IdentityCoreIntegrationPointV2	= "bank_module"
	IdentityCoreIntegrationFeeModule	IdentityCoreIntegrationPointV2	= "fee_module"
	IdentityCoreIntegrationResolverModule	IdentityCoreIntegrationPointV2	= "resolver_module"
	IdentityCoreIntegrationStoreV2		IdentityCoreIntegrationPointV2	= "store_v2"
	IdentityCoreIntegrationBlockSTM		IdentityCoreIntegrationPointV2	= "blockstm"
)

type IdentityCoreFailureCoverageV2 struct {
	Mode		IdentityCoreFailureModeV2
	Guard		string
	StoreScope	string
}

type IdentityCoreLifecycleKeeperContractV2 struct {
	Responsibilities	[]string
	MessagesHandled		[]IdentityCoreMessageNameV2
	QueriesServed		[]IdentityCoreQueryNameV2
	NFTHooks		[]string
	Invariants		[]IdentityCoreFailureModeV2
	BlockSTMClasses		[]string
}

type IdentityCoreModuleBreakdownV2 struct {
	ModulePath		string
	Purpose			[]string
	StateObjects		[]IdentityCoreStateObjectV2
	Messages		[]IdentityCoreMessageNameV2
	Queries			[]IdentityCoreQueryNameV2
	FailureModes		[]IdentityCoreFailureCoverageV2
	IntegrationPoints	[]IdentityCoreIntegrationPointV2
	Keeper			IdentityCoreLifecycleKeeperContractV2
	StoreKeys		[]string
	BreakdownHash		string
}

type IdentityCoreIndexEntryV2 struct {
	Kind		string
	StoreKey	string
	Name		string
	NameHash	string
	Owner		string
	ExpiryHeight	uint64
}

type IdentityCoreDerivedIndexesV2 struct {
	ExpiryIndex	[]IdentityCoreIndexEntryV2
	OwnerIndex	[]IdentityCoreIndexEntryV2
}

type IdentityCoreInvariantReportV2 struct {
	Valid		bool
	Issues		[]string
	DerivedIndexes	IdentityCoreDerivedIndexesV2
	ReportHash	string
}

func DefaultIdentityCoreModuleBreakdownV2() (IdentityCoreModuleBreakdownV2, error) {
	breakdown := IdentityCoreModuleBreakdownV2{
		ModulePath:	IdentityCoreModulePathV2,
		Purpose: []string{
			"canonical_domain_lifecycle",
			"expiry_processing",
			"nft_backed_ownership",
			"registration_and_commit_reveal",
			"renewal_and_transfer",
			"store_v2_indexes",
		},
		StateObjects:	requiredIdentityCoreStateObjectsV2(),
		Messages:	requiredIdentityCoreMessagesV2(),
		Queries:	requiredIdentityCoreQueriesV2(),
		FailureModes: []IdentityCoreFailureCoverageV2{
			newIdentityCoreFailureCoverageV2(IdentityCoreFailureNFTOwnerMismatch, "IdentityState.Validate", IdentityStoreV2SpecNFTBindingsByNamePrefix),
			newIdentityCoreFailureCoverageV2(IdentityCoreFailureMissingExpiryIndex, "ValidateIdentityCoreStoreV2IndexesV2", IdentityStoreV2SpecExpiryIndexPrefix),
			newIdentityCoreFailureCoverageV2(IdentityCoreFailureCommitmentReplay, "CommitDomainRegistration", IdentityStoreV2SpecCommitmentsPrefix),
			newIdentityCoreFailureCoverageV2(IdentityCoreFailureTransferResolverRace, "TransferDomainNFT", IdentityStoreV2PendingResolverPrefix),
			newIdentityCoreFailureCoverageV2(IdentityCoreFailureInvalidLifecycleChange, "ApplyDomainLifecycleTransitionV2", IdentityStoreV2SpecDomainsPrefix),
		},
		IntegrationPoints:	requiredIdentityCoreIntegrationPointsV2(),
		Keeper: IdentityCoreLifecycleKeeperContractV2{
			Responsibilities: []string{
				"domain_activation",
				"domain_expiry",
				"domain_registration",
				"domain_renewal",
				"domain_transfer",
				"nft_binding_repair",
			},
			MessagesHandled:	requiredIdentityCoreMessagesV2(),
			QueriesServed:		requiredIdentityCoreQueriesV2(),
			NFTHooks: []string{
				"ApplyIdentityNFTTransferHookStateV2",
				"ApplyIdentityRegistryTransferHookStateV2",
				"RepairDomainNFTBindingInternalFailureV2",
			},
			Invariants: []IdentityCoreFailureModeV2{
				IdentityCoreFailureNFTOwnerMismatch,
				IdentityCoreFailureMissingExpiryIndex,
				IdentityCoreFailureCommitmentReplay,
				IdentityCoreFailureTransferResolverRace,
				IdentityCoreFailureInvalidLifecycleChange,
			},
			BlockSTMClasses: []string{
				"register_different_names",
				"renew_different_names",
				"transfer_different_names",
				"expire_partitioned_by_height",
				"same_name_version_conflict",
			},
		},
		StoreKeys: []string{
			IdentityStoreV2SpecDomainsPrefix,
			IdentityStoreV2SpecDomainNamesPrefix,
			IdentityStoreV2SpecCommitmentsPrefix,
			IdentityStoreV2SpecNFTBindingsPrefix,
			IdentityStoreV2SpecNFTBindingsByNamePrefix,
			IdentityStoreV2SpecExpiryIndexPrefix,
			IdentityStoreV2SpecOwnerIndexPrefix,
		},
	}
	return NewIdentityCoreModuleBreakdownV2(breakdown)
}

func NewIdentityCoreModuleBreakdownV2(breakdown IdentityCoreModuleBreakdownV2) (IdentityCoreModuleBreakdownV2, error) {
	breakdown = canonicalIdentityCoreModuleBreakdownV2(breakdown)
	if err := breakdown.ValidateFormat(); err != nil {
		return IdentityCoreModuleBreakdownV2{}, err
	}
	breakdown.BreakdownHash = ComputeIdentityCoreModuleBreakdownHashV2(breakdown)
	return breakdown, breakdown.Validate()
}

func (breakdown IdentityCoreModuleBreakdownV2) ValidateFormat() error {
	if breakdown.ModulePath != IdentityCoreModulePathV2 {
		return errors.New("identity core breakdown must describe identity-core-module")
	}
	if err := validateIdentityCoreTokenSetV2("purpose", breakdown.Purpose, nil); err != nil {
		return err
	}
	if err := validateIdentityCoreEnumSetV2("state object", breakdown.StateObjects, requiredIdentityCoreStateObjectsV2(), IsIdentityCoreStateObjectV2); err != nil {
		return err
	}
	if err := validateIdentityCoreEnumSetV2("message", breakdown.Messages, requiredIdentityCoreMessagesV2(), IsIdentityCoreMessageNameV2); err != nil {
		return err
	}
	if err := validateIdentityCoreEnumSetV2("query", breakdown.Queries, requiredIdentityCoreQueriesV2(), IsIdentityCoreQueryNameV2); err != nil {
		return err
	}
	if err := validateIdentityCoreFailureCoverageV2(breakdown.FailureModes); err != nil {
		return err
	}
	if err := validateIdentityCoreEnumSetV2("integration", breakdown.IntegrationPoints, requiredIdentityCoreIntegrationPointsV2(), IsIdentityCoreIntegrationPointV2); err != nil {
		return err
	}
	if err := breakdown.Keeper.Validate(); err != nil {
		return err
	}
	if err := validateIdentityCoreStoreKeysV2(breakdown.StoreKeys); err != nil {
		return err
	}
	if breakdown.BreakdownHash != "" {
		return validateHexHash("identity core breakdown hash", breakdown.BreakdownHash)
	}
	return nil
}

func (breakdown IdentityCoreModuleBreakdownV2) Validate() error {
	breakdown = canonicalIdentityCoreModuleBreakdownV2(breakdown)
	if err := breakdown.ValidateFormat(); err != nil {
		return err
	}
	if breakdown.BreakdownHash == "" {
		return errors.New("identity core breakdown hash is required")
	}
	if breakdown.BreakdownHash != ComputeIdentityCoreModuleBreakdownHashV2(breakdown) {
		return errors.New("identity core breakdown hash mismatch")
	}
	return nil
}

func (keeper IdentityCoreLifecycleKeeperContractV2) Validate() error {
	keeper = canonicalIdentityCoreLifecycleKeeperContractV2(keeper)
	if err := validateIdentityCoreTokenSetV2("keeper responsibility", keeper.Responsibilities, nil); err != nil {
		return err
	}
	if err := validateIdentityCoreEnumSetV2("keeper message", keeper.MessagesHandled, requiredIdentityCoreMessagesV2(), IsIdentityCoreMessageNameV2); err != nil {
		return err
	}
	if err := validateIdentityCoreEnumSetV2("keeper query", keeper.QueriesServed, requiredIdentityCoreQueriesV2(), IsIdentityCoreQueryNameV2); err != nil {
		return err
	}
	if err := validateIdentityCoreTokenSetV2("nft hook", keeper.NFTHooks, []string{"ApplyIdentityNFTTransferHookStateV2", "ApplyIdentityRegistryTransferHookStateV2", "RepairDomainNFTBindingInternalFailureV2"}); err != nil {
		return err
	}
	if err := validateIdentityCoreEnumSetV2("keeper invariant", keeper.Invariants, requiredIdentityCoreFailureModesV2(), IsIdentityCoreFailureModeV2); err != nil {
		return err
	}
	return validateIdentityCoreTokenSetV2("blockstm class", keeper.BlockSTMClasses, []string{"expire_partitioned_by_height", "register_different_names", "renew_different_names", "same_name_version_conflict", "transfer_different_names"})
}

func (coverage IdentityCoreFailureCoverageV2) Validate() error {
	if !IsIdentityCoreFailureModeV2(coverage.Mode) {
		return fmt.Errorf("identity core unknown failure mode %q", coverage.Mode)
	}
	if strings.TrimSpace(coverage.Guard) == "" {
		return errors.New("identity core failure guard is required")
	}
	if !strings.HasPrefix(coverage.StoreScope, IdentityStoreV2Prefix+"/") {
		return fmt.Errorf("identity core failure scope %s must be under Store v2 identity prefix", coverage.StoreScope)
	}
	return nil
}

func ValidateIdentityCoreLifecycleTransitionV2(before DomainRecordV2, ctx DomainLifecycleTransitionContextV2) (DomainRecordV2, error) {
	after, err := ApplyDomainLifecycleTransitionV2(before, ctx)
	if err != nil {
		return DomainRecordV2{}, fmt.Errorf("%s: %w", IdentityCoreFailureInvalidLifecycleChange, err)
	}
	return after, nil
}

func ValidateIdentityCoreModuleInvariantsV2(state IdentityState, height uint64) (IdentityCoreInvariantReportV2, error) {
	if height == 0 {
		return IdentityCoreInvariantReportV2{}, errors.New("identity core invariant height must be positive")
	}
	indexes, err := BuildIdentityCoreDerivedIndexesV2(state)
	if err != nil {
		return IdentityCoreInvariantReportV2{}, err
	}
	report := IdentityCoreInvariantReportV2{Valid: true, DerivedIndexes: indexes}
	if err := state.Validate(); err != nil {
		report.addIssue(fmt.Sprintf("state_validation: %v", err))
	}
	report.merge(ValidateIdentityCoreStoreV2IndexesV2(state, indexes))
	for _, domain := range state.Domains {
		lifecycle, err := DomainLifecycle(state, domain.Name, height)
		if err != nil {
			report.addIssue(fmt.Sprintf("domain_lifecycle: %s: %v", domain.Name, err))
			continue
		}
		if lifecycle == DomainLifecycleActive || lifecycle == DomainLifecycleRenewalWindow {
			nft, found := findDomainNFTByID(state, domain.NFTID)
			if !found || !bytes.Equal(nft.Owner, domain.Owner) || nft.Domain != domain.Name {
				report.addIssue(fmt.Sprintf("%s: %s", IdentityCoreFailureNFTOwnerMismatch, domain.Name))
			}
		}
	}
	seenCommitments := map[string]struct{}{}
	for _, tombstone := range state.UsedCommitments {
		if _, found := seenCommitments[tombstone.CommitmentHash]; found {
			report.addIssue(fmt.Sprintf("%s: %s", IdentityCoreFailureCommitmentReplay, tombstone.CommitmentHash))
			continue
		}
		seenCommitments[tombstone.CommitmentHash] = struct{}{}
	}
	for _, intent := range state.PendingResolverUpdates {
		domain, found := findDomain(state, intent.Domain)
		if found && !bytes.Equal(domain.Owner, intent.Actor) {
			report.addIssue(fmt.Sprintf("%s: %s", IdentityCoreFailureTransferResolverRace, intent.Domain))
		}
	}
	report.finalize()
	return report, nil
}

func BuildIdentityCoreDerivedIndexesV2(state IdentityState) (IdentityCoreDerivedIndexesV2, error) {
	state = normalizeIdentityStateParams(state)
	indexes := IdentityCoreDerivedIndexesV2{
		ExpiryIndex:	make([]IdentityCoreIndexEntryV2, 0, len(state.Domains)),
		OwnerIndex:	make([]IdentityCoreIndexEntryV2, 0, len(state.Domains)),
	}
	for _, domain := range state.Domains {
		nameHash, err := DomainRecordV2NameHash(domain.Name)
		if err != nil {
			return IdentityCoreDerivedIndexesV2{}, err
		}
		expiryKey, err := IdentityStoreV2SpecExpiryIndexKey(domain.ExpiryHeight, domain.Name)
		if err != nil {
			return IdentityCoreDerivedIndexesV2{}, err
		}
		ownerKey, err := IdentityStoreV2SpecOwnerIndexKey(domain.Owner, domain.Name)
		if err != nil {
			return IdentityCoreDerivedIndexesV2{}, err
		}
		owner := hex.EncodeToString(domain.Owner)
		indexes.ExpiryIndex = append(indexes.ExpiryIndex, IdentityCoreIndexEntryV2{
			Kind:		"expiry_index",
			StoreKey:	expiryKey,
			Name:		domain.Name,
			NameHash:	nameHash,
			Owner:		owner,
			ExpiryHeight:	domain.ExpiryHeight,
		})
		indexes.OwnerIndex = append(indexes.OwnerIndex, IdentityCoreIndexEntryV2{
			Kind:		"owner_index",
			StoreKey:	ownerKey,
			Name:		domain.Name,
			NameHash:	nameHash,
			Owner:		owner,
			ExpiryHeight:	domain.ExpiryHeight,
		})
	}
	sortIdentityCoreIndexEntriesV2(indexes.ExpiryIndex)
	sortIdentityCoreIndexEntriesV2(indexes.OwnerIndex)
	return indexes, nil
}

func ValidateIdentityCoreStoreV2IndexesV2(state IdentityState, observed IdentityCoreDerivedIndexesV2) IdentityCoreInvariantReportV2 {
	expected, err := BuildIdentityCoreDerivedIndexesV2(state)
	report := IdentityCoreInvariantReportV2{Valid: true, DerivedIndexes: observed}
	if err != nil {
		report.addIssue(err.Error())
		report.finalize()
		return report
	}
	compareIdentityCoreIndexEntriesV2(&report, string(IdentityCoreFailureMissingExpiryIndex), "expiry_index", expected.ExpiryIndex, observed.ExpiryIndex)
	compareIdentityCoreIndexEntriesV2(&report, "owner_index_missing_domain", "owner_index", expected.OwnerIndex, observed.OwnerIndex)
	report.finalize()
	return report
}

func ComputeIdentityCoreModuleBreakdownHashV2(breakdown IdentityCoreModuleBreakdownV2) string {
	breakdown = canonicalIdentityCoreModuleBreakdownV2(breakdown)
	parts := []string{"aetra-identity-core-module-breakdown-v2", breakdown.ModulePath}
	parts = append(parts, "purpose", fmt.Sprint(len(breakdown.Purpose)))
	parts = append(parts, breakdown.Purpose...)
	parts = append(parts, "state", fmt.Sprint(len(breakdown.StateObjects)))
	for _, object := range breakdown.StateObjects {
		parts = append(parts, string(object))
	}
	parts = append(parts, "messages", fmt.Sprint(len(breakdown.Messages)))
	for _, message := range breakdown.Messages {
		parts = append(parts, string(message))
	}
	parts = append(parts, "queries", fmt.Sprint(len(breakdown.Queries)))
	for _, query := range breakdown.Queries {
		parts = append(parts, string(query))
	}
	parts = append(parts, "failures", fmt.Sprint(len(breakdown.FailureModes)))
	for _, failure := range breakdown.FailureModes {
		parts = append(parts, string(failure.Mode), failure.Guard, failure.StoreScope)
	}
	parts = append(parts, "integrations", fmt.Sprint(len(breakdown.IntegrationPoints)))
	for _, point := range breakdown.IntegrationPoints {
		parts = append(parts, string(point))
	}
	parts = append(parts, "keeper")
	parts = append(parts, breakdown.Keeper.Responsibilities...)
	for _, message := range breakdown.Keeper.MessagesHandled {
		parts = append(parts, "keeper-message", string(message))
	}
	for _, query := range breakdown.Keeper.QueriesServed {
		parts = append(parts, "keeper-query", string(query))
	}
	parts = append(parts, breakdown.Keeper.NFTHooks...)
	for _, invariant := range breakdown.Keeper.Invariants {
		parts = append(parts, "keeper-invariant", string(invariant))
	}
	parts = append(parts, breakdown.Keeper.BlockSTMClasses...)
	parts = append(parts, "store", fmt.Sprint(len(breakdown.StoreKeys)))
	parts = append(parts, breakdown.StoreKeys...)
	return identityHash(parts...)
}

func ComputeIdentityCoreInvariantReportHashV2(report IdentityCoreInvariantReportV2) string {
	report.Issues = sortedIdentityCoreStringsV2(report.Issues)
	sortIdentityCoreIndexEntriesV2(report.DerivedIndexes.ExpiryIndex)
	sortIdentityCoreIndexEntriesV2(report.DerivedIndexes.OwnerIndex)
	parts := []string{"aetra-identity-core-invariant-report-v2", fmt.Sprintf("valid=%t", report.Valid)}
	parts = append(parts, report.Issues...)
	for _, entry := range report.DerivedIndexes.ExpiryIndex {
		parts = append(parts, entry.Kind, entry.StoreKey, entry.Name, entry.NameHash, entry.Owner, fmt.Sprint(entry.ExpiryHeight))
	}
	for _, entry := range report.DerivedIndexes.OwnerIndex {
		parts = append(parts, entry.Kind, entry.StoreKey, entry.Name, entry.NameHash, entry.Owner, fmt.Sprint(entry.ExpiryHeight))
	}
	return identityHash(parts...)
}

func IsIdentityCoreStateObjectV2(object IdentityCoreStateObjectV2) bool {
	switch object {
	case IdentityCoreStateDomainRecord, IdentityCoreStateDomainCommitment, IdentityCoreStateDomainNFTBinding, IdentityCoreStateDomainLifecycleParams, IdentityCoreStateExpiryIndex, IdentityCoreStateOwnerIndex:
		return true
	default:
		return false
	}
}

func IsIdentityCoreMessageNameV2(message IdentityCoreMessageNameV2) bool {
	switch message {
	case IdentityCoreMsgCommitRegistration, IdentityCoreMsgRevealRegistration, IdentityCoreMsgRegisterDirect, IdentityCoreMsgRenewDomain, IdentityCoreMsgTransferDomain, IdentityCoreMsgExpireDomain, IdentityCoreMsgRepairNFTBinding:
		return true
	default:
		return false
	}
}

func IsIdentityCoreQueryNameV2(query IdentityCoreQueryNameV2) bool {
	switch query {
	case IdentityCoreQueryDomain, IdentityCoreQueryDomainByName, IdentityCoreQueryDomainsByOwner, IdentityCoreQueryDomainNFTBinding, IdentityCoreQueryDomainLifecycle, IdentityCoreQueryRegistrationPrice, IdentityCoreQueryRenewalPrice:
		return true
	default:
		return false
	}
}

func IsIdentityCoreFailureModeV2(mode IdentityCoreFailureModeV2) bool {
	switch mode {
	case IdentityCoreFailureNFTOwnerMismatch, IdentityCoreFailureMissingExpiryIndex, IdentityCoreFailureCommitmentReplay, IdentityCoreFailureTransferResolverRace, IdentityCoreFailureInvalidLifecycleChange:
		return true
	default:
		return false
	}
}

func IsIdentityCoreIntegrationPointV2(point IdentityCoreIntegrationPointV2) bool {
	switch point {
	case IdentityCoreIntegrationNFTModule, IdentityCoreIntegrationBankModule, IdentityCoreIntegrationFeeModule, IdentityCoreIntegrationResolverModule, IdentityCoreIntegrationStoreV2, IdentityCoreIntegrationBlockSTM:
		return true
	default:
		return false
	}
}

func newIdentityCoreFailureCoverageV2(mode IdentityCoreFailureModeV2, guard, storeScope string) IdentityCoreFailureCoverageV2 {
	return IdentityCoreFailureCoverageV2{Mode: mode, Guard: guard, StoreScope: storeScope}
}

func requiredIdentityCoreStateObjectsV2() []IdentityCoreStateObjectV2 {
	return []IdentityCoreStateObjectV2{IdentityCoreStateDomainCommitment, IdentityCoreStateDomainLifecycleParams, IdentityCoreStateDomainNFTBinding, IdentityCoreStateDomainRecord, IdentityCoreStateExpiryIndex, IdentityCoreStateOwnerIndex}
}

func requiredIdentityCoreMessagesV2() []IdentityCoreMessageNameV2 {
	return []IdentityCoreMessageNameV2{IdentityCoreMsgCommitRegistration, IdentityCoreMsgExpireDomain, IdentityCoreMsgRegisterDirect, IdentityCoreMsgRenewDomain, IdentityCoreMsgRepairNFTBinding, IdentityCoreMsgRevealRegistration, IdentityCoreMsgTransferDomain}
}

func requiredIdentityCoreQueriesV2() []IdentityCoreQueryNameV2 {
	return []IdentityCoreQueryNameV2{IdentityCoreQueryDomain, IdentityCoreQueryDomainByName, IdentityCoreQueryDomainLifecycle, IdentityCoreQueryDomainNFTBinding, IdentityCoreQueryDomainsByOwner, IdentityCoreQueryRegistrationPrice, IdentityCoreQueryRenewalPrice}
}

func requiredIdentityCoreFailureModesV2() []IdentityCoreFailureModeV2 {
	return []IdentityCoreFailureModeV2{IdentityCoreFailureCommitmentReplay, IdentityCoreFailureMissingExpiryIndex, IdentityCoreFailureInvalidLifecycleChange, IdentityCoreFailureNFTOwnerMismatch, IdentityCoreFailureTransferResolverRace}
}

func requiredIdentityCoreIntegrationPointsV2() []IdentityCoreIntegrationPointV2 {
	return []IdentityCoreIntegrationPointV2{IdentityCoreIntegrationBankModule, IdentityCoreIntegrationBlockSTM, IdentityCoreIntegrationFeeModule, IdentityCoreIntegrationNFTModule, IdentityCoreIntegrationResolverModule, IdentityCoreIntegrationStoreV2}
}

func canonicalIdentityCoreModuleBreakdownV2(breakdown IdentityCoreModuleBreakdownV2) IdentityCoreModuleBreakdownV2 {
	breakdown.ModulePath = strings.TrimSpace(breakdown.ModulePath)
	breakdown.Purpose = sortedIdentityCoreStringsV2(breakdown.Purpose)
	breakdown.StateObjects = sortedIdentityCoreTypedStringsV2(breakdown.StateObjects)
	breakdown.Messages = sortedIdentityCoreTypedStringsV2(breakdown.Messages)
	breakdown.Queries = sortedIdentityCoreTypedStringsV2(breakdown.Queries)
	breakdown.FailureModes = sortedIdentityCoreFailureCoverageV2(breakdown.FailureModes)
	breakdown.IntegrationPoints = sortedIdentityCoreTypedStringsV2(breakdown.IntegrationPoints)
	breakdown.Keeper = canonicalIdentityCoreLifecycleKeeperContractV2(breakdown.Keeper)
	breakdown.StoreKeys = sortedIdentityCoreStringsV2(breakdown.StoreKeys)
	breakdown.BreakdownHash = strings.TrimSpace(breakdown.BreakdownHash)
	return breakdown
}

func canonicalIdentityCoreLifecycleKeeperContractV2(keeper IdentityCoreLifecycleKeeperContractV2) IdentityCoreLifecycleKeeperContractV2 {
	keeper.Responsibilities = sortedIdentityCoreStringsV2(keeper.Responsibilities)
	keeper.MessagesHandled = sortedIdentityCoreTypedStringsV2(keeper.MessagesHandled)
	keeper.QueriesServed = sortedIdentityCoreTypedStringsV2(keeper.QueriesServed)
	keeper.NFTHooks = sortedIdentityCoreStringsV2(keeper.NFTHooks)
	keeper.Invariants = sortedIdentityCoreTypedStringsV2(keeper.Invariants)
	keeper.BlockSTMClasses = sortedIdentityCoreStringsV2(keeper.BlockSTMClasses)
	return keeper
}

func validateIdentityCoreFailureCoverageV2(coverage []IdentityCoreFailureCoverageV2) error {
	if len(coverage) != len(requiredIdentityCoreFailureModesV2()) {
		return fmt.Errorf("identity core expected %d failure modes", len(requiredIdentityCoreFailureModesV2()))
	}
	seen := map[IdentityCoreFailureModeV2]struct{}{}
	for _, item := range coverage {
		if err := item.Validate(); err != nil {
			return err
		}
		if _, found := seen[item.Mode]; found {
			return fmt.Errorf("identity core duplicate failure mode %s", item.Mode)
		}
		seen[item.Mode] = struct{}{}
	}
	for _, required := range requiredIdentityCoreFailureModesV2() {
		if _, found := seen[required]; !found {
			return fmt.Errorf("identity core missing failure mode %s", required)
		}
	}
	return nil
}

func validateIdentityCoreStoreKeysV2(keys []string) error {
	required := []string{
		IdentityStoreV2SpecCommitmentsPrefix,
		IdentityStoreV2SpecDomainNamesPrefix,
		IdentityStoreV2SpecDomainsPrefix,
		IdentityStoreV2SpecExpiryIndexPrefix,
		IdentityStoreV2SpecNFTBindingsByNamePrefix,
		IdentityStoreV2SpecNFTBindingsPrefix,
		IdentityStoreV2SpecOwnerIndexPrefix,
	}
	return validateIdentityCoreTokenSetV2("store key", keys, required)
}

func validateIdentityCoreEnumSetV2[T ~string](label string, values []T, required []T, allowed func(T) bool) error {
	if len(values) != len(required) {
		return fmt.Errorf("identity core expected %d %s entries", len(required), label)
	}
	seen := map[T]struct{}{}
	previous := ""
	for _, value := range values {
		if !allowed(value) {
			return fmt.Errorf("identity core unknown %s %q", label, value)
		}
		if _, found := seen[value]; found {
			return fmt.Errorf("identity core duplicate %s %q", label, value)
		}
		if previous != "" && previous >= string(value) {
			return fmt.Errorf("identity core %s entries must be sorted", label)
		}
		seen[value] = struct{}{}
		previous = string(value)
	}
	for _, requiredValue := range required {
		if _, found := seen[requiredValue]; !found {
			return fmt.Errorf("identity core missing %s %q", label, requiredValue)
		}
	}
	return nil
}

func validateIdentityCoreTokenSetV2(label string, values []string, required []string) error {
	if len(values) == 0 {
		return fmt.Errorf("identity core %s entries are required", label)
	}
	if len(required) > 0 && len(values) != len(required) {
		return fmt.Errorf("identity core expected %d %s entries", len(required), label)
	}
	seen := map[string]struct{}{}
	previous := ""
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			return fmt.Errorf("identity core %s entry is required", label)
		}
		if _, found := seen[value]; found {
			return fmt.Errorf("identity core duplicate %s %q", label, value)
		}
		if previous != "" && previous >= value {
			return fmt.Errorf("identity core %s entries must be sorted", label)
		}
		seen[value] = struct{}{}
		previous = value
	}
	for _, requiredValue := range required {
		if _, found := seen[requiredValue]; !found {
			return fmt.Errorf("identity core missing %s %q", label, requiredValue)
		}
	}
	return nil
}

func compareIdentityCoreIndexEntriesV2(report *IdentityCoreInvariantReportV2, mode string, kind string, expected []IdentityCoreIndexEntryV2, observed []IdentityCoreIndexEntryV2) {
	observedByKey := make(map[string]IdentityCoreIndexEntryV2, len(observed))
	for _, entry := range observed {
		if entry.Kind != kind {
			report.addIssue(fmt.Sprintf("%s invalid index kind %s for %s", mode, entry.Kind, entry.StoreKey))
			continue
		}
		observedByKey[entry.StoreKey] = entry
	}
	for _, expectedEntry := range expected {
		observedEntry, found := observedByKey[expectedEntry.StoreKey]
		if !found {
			report.addIssue(fmt.Sprintf("%s: %s", mode, expectedEntry.Name))
			continue
		}
		if observedEntry != expectedEntry {
			report.addIssue(fmt.Sprintf("%s stale index entry: %s", mode, expectedEntry.Name))
		}
	}
}

func (report *IdentityCoreInvariantReportV2) merge(other IdentityCoreInvariantReportV2) {
	for _, issue := range other.Issues {
		report.addIssue(issue)
	}
}

func (report *IdentityCoreInvariantReportV2) addIssue(issue string) {
	report.Valid = false
	report.Issues = append(report.Issues, issue)
}

func (report *IdentityCoreInvariantReportV2) finalize() {
	report.Issues = sortedIdentityCoreStringsV2(report.Issues)
	report.Valid = len(report.Issues) == 0
	report.ReportHash = ComputeIdentityCoreInvariantReportHashV2(*report)
}

func sortedIdentityCoreFailureCoverageV2(values []IdentityCoreFailureCoverageV2) []IdentityCoreFailureCoverageV2 {
	out := append([]IdentityCoreFailureCoverageV2(nil), values...)
	for i := range out {
		out[i].Guard = strings.TrimSpace(out[i].Guard)
		out[i].StoreScope = strings.TrimSpace(out[i].StoreScope)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Mode == out[j].Mode {
			return out[i].Guard < out[j].Guard
		}
		return out[i].Mode < out[j].Mode
	})
	return out
}

func sortedIdentityCoreTypedStringsV2[T ~string](values []T) []T {
	out := append([]T(nil), values...)
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func sortedIdentityCoreStringsV2(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, strings.TrimSpace(value))
	}
	sort.Strings(out)
	return out
}

func sortIdentityCoreIndexEntriesV2(entries []IdentityCoreIndexEntryV2) {
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].StoreKey == entries[j].StoreKey {
			return entries[i].Name < entries[j].Name
		}
		return entries[i].StoreKey < entries[j].StoreKey
	})
}

func IdentityCoreStoreV2KeyExamplesV2(name string, owner sdk.AccAddress, commitmentHash string) ([]string, error) {
	domainKey, err := IdentityStoreV2SpecDomainKey(name)
	if err != nil {
		return nil, err
	}
	domainNameKey, err := IdentityStoreV2SpecDomainNameKey(name)
	if err != nil {
		return nil, err
	}
	commitmentKey, err := IdentityStoreV2SpecCommitmentKey(commitmentHash)
	if err != nil {
		return nil, err
	}
	nftID, err := DomainNFTID(name)
	if err != nil {
		return nil, err
	}
	nftKey, err := IdentityStoreV2SpecNFTBindingKey(DomainNFTClassID, nftID)
	if err != nil {
		return nil, err
	}
	nftByNameKey, err := IdentityStoreV2SpecNFTBindingByNameKey(name)
	if err != nil {
		return nil, err
	}
	expiryKey, err := IdentityStoreV2SpecExpiryIndexKey(DefaultRegistrationPeriodBlocks, name)
	if err != nil {
		return nil, err
	}
	ownerKey, err := IdentityStoreV2SpecOwnerIndexKey(owner, name)
	if err != nil {
		return nil, err
	}
	return []string{commitmentKey, domainKey, domainNameKey, expiryKey, nftByNameKey, nftKey, ownerKey}, nil
}
