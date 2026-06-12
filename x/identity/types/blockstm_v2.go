package types

import (
	"encoding/hex"
	"errors"
	"fmt"
)

const (
	IdentityStoreV2SpecFeeAccumulatorPrefix	= IdentityStoreV2Prefix + "/fee_accumulator"

	IdentityBlockSTMConflictNoneV2				= "none"
	IdentityBlockSTMConflictSameNameV2			= "same_name"
	IdentityBlockSTMConflictParentPolicyChildCreateV2	= "parent_policy_child_create"
	IdentityBlockSTMConflictTransferResolverUpdateV2	= "transfer_resolver_update"
	IdentityBlockSTMConflictReversePrimaryResolverV2	= "reverse_primary_resolver_update"
	IdentityBlockSTMConflictAuctionFinalizeLateRevealV2	= "auction_finalize_late_reveal"
	IdentityBlockSTMConflictBatchDuplicateNameHashV2	= "batch_duplicate_name_hash"
	IdentityBlockSTMConflictParentChildUnknownV2		= "parent_child_unknown"
	IdentityBlockSTMFeeAccumulatorPartitionModuleV2		= "identity"
	IdentityBlockSTMFeeAccumulatorPartitionBatchModuleV2	= "identity_batch"
)

type IdentityBlockSTMConflictClassV2 string

type IdentityBlockSTMPlanV2 struct {
	MessageName	string
	AccessSet	IdentityAccessSet
	ConflictClass	IdentityBlockSTMConflictClassV2
	FeeKey		string
	NameHashes	[]string
}

func IdentityStoreV2SpecFeeAccumulatorKey(blockHeight uint64, module string) (string, error) {
	if blockHeight == 0 {
		return "", errors.New("identity v2 BlockSTM fee accumulator height is required")
	}
	if err := validateStorePathSegmentV2("identity v2 BlockSTM fee accumulator module", module); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%020d/%s", IdentityStoreV2SpecFeeAccumulatorPrefix, blockHeight, module), nil
}

func IdentityBlockSTMAccessSetV2(msg IdentityMsgV2, blockHeight uint64) (IdentityBlockSTMPlanV2, error) {
	if err := ValidateIdentityMsgV2(msg); err != nil {
		return IdentityBlockSTMPlanV2{}, err
	}
	feeKey, err := IdentityStoreV2SpecFeeAccumulatorKey(blockHeight, IdentityBlockSTMFeeAccumulatorPartitionModuleV2)
	if err != nil {
		return IdentityBlockSTMPlanV2{}, err
	}
	plan := IdentityBlockSTMPlanV2{
		MessageName:	msg.IdentityMessageName(),
		ConflictClass:	IdentityBlockSTMConflictNoneV2,
		FeeKey:		feeKey,
	}
	switch m := msg.(type) {
	case MsgCommitRegistrationV2:
		set, hashes, err := blockSTMNameLifecycleSetV2(m.Name, m.NameHash, m.Auth.Signer, false, false)
		return plan.with(set, hashes, err)
	case MsgRevealRegistrationV2:
		set, hashes, err := blockSTMNameLifecycleSetV2(m.Name, m.NameHash, m.Auth.Signer, true, false)
		return plan.with(set, hashes, err)
	case MsgRegisterDirectV2:
		set, hashes, err := blockSTMNameLifecycleSetV2(m.Name, m.NameHash, m.Owner, true, false)
		return plan.with(set, hashes, err)
	case MsgRenewDomainV2:
		set, hashes, err := blockSTMNameLifecycleSetV2(m.Name, m.NameHash, m.Auth.Signer, false, false)
		return plan.with(set, hashes, err)
	case MsgTransferDomainV2:
		set, hashes, err := blockSTMNameLifecycleSetV2(m.Name, m.NameHash, m.NewOwner, true, true)
		plan.ConflictClass = IdentityBlockSTMConflictTransferResolverUpdateV2
		return plan.with(set, hashes, err)
	case MsgSetResolverV2:
		set, hashes, err := blockSTMResolverUpdateSetV2(m.Name, m.NameHash, false)
		return plan.with(set, hashes, err)
	case MsgUpdateResolverRecordV2:
		set, hashes, err := blockSTMResolverUpdateSetV2(m.Name, m.NameHash, resolverPatchTouchesPrimaryV2(m.Patch))
		if resolverPatchTouchesPrimaryV2(m.Patch) {
			plan.ConflictClass = IdentityBlockSTMConflictReversePrimaryResolverV2
		}
		return plan.with(set, hashes, err)
	case MsgSetReverseRecordV2:
		set, hashes, err := blockSTMReverseSetV2(m.Record)
		return plan.with(set, hashes, err)
	case MsgVerifyReverseRecordV2:
		set, hashes, err := blockSTMReverseVerifySetV2(m.Record)
		plan.ConflictClass = IdentityBlockSTMConflictReversePrimaryResolverV2
		return plan.with(set, hashes, err)
	case MsgCreateSubdomainV2:
		set, hashes, err := blockSTMSubdomainCreateSetV2(m.ParentName, m.ParentNameHash, m.Label)
		plan.ConflictClass = IdentityBlockSTMConflictParentPolicyChildCreateV2
		return plan.with(set, hashes, err)
	case MsgDelegateSubdomainV2:
		set, hashes, err := blockSTMDelegationSetV2(m.Delegation)
		return plan.with(set, hashes, err)
	case MsgRevokeDelegationV2:
		set, hashes, err := blockSTMDelegationRevokeSetV2(m.Name, m.NameHash, m.Delegate, m.Scope)
		return plan.with(set, hashes, err)
	case MsgStartAuctionV2:
		set, hashes, err := blockSTMAuctionSetV2(m.AuctionID(), m.NameHash)
		return plan.with(set, hashes, err)
	case MsgCommitBidV2:
		set, hashes, err := blockSTMAuctionBidSetV2(m.AuctionID, m.NameHash, m.Auth.Signer, m.CommitmentHash)
		return plan.with(set, hashes, err)
	case MsgRevealBidV2:
		set, hashes, err := blockSTMAuctionBidSetV2(m.AuctionID, m.NameHash, m.Auth.Signer, m.CommitmentHash)
		plan.ConflictClass = IdentityBlockSTMConflictAuctionFinalizeLateRevealV2
		return plan.with(set, hashes, err)
	case MsgFinalizeAuctionV2:
		set, hashes, err := blockSTMAuctionSetV2(m.AuctionID, m.NameHash)
		plan.ConflictClass = IdentityBlockSTMConflictAuctionFinalizeLateRevealV2
		return plan.with(set, hashes, err)
	case MsgExpireDomainV2:
		set, hashes, err := blockSTMNameLifecycleSetV2(m.Name, m.NameHash, m.Auth.Signer, false, false)
		return plan.with(set, hashes, err)
	case MsgBatchUpdateResolversV2:
		feeKey, err := IdentityStoreV2SpecFeeAccumulatorKey(blockHeight, IdentityBlockSTMFeeAccumulatorPartitionBatchModuleV2)
		if err != nil {
			return IdentityBlockSTMPlanV2{}, err
		}
		plan.FeeKey = feeKey
		set, hashes, err := IdentityBlockSTMBatchResolverAccessSetV2(m)
		return plan.with(set, hashes, err)
	case MsgBatchRenewDomainsV2:
		set, hashes, err := blockSTMBatchRenewSetV2(m)
		return plan.with(set, hashes, err)
	case MsgInvalidateResolutionCacheV2:
		set := newIdentityAccessSet([]string{IdentityStoreV2SpecResolutionCachePrefix + "/" + m.NameHash + "/" + m.ResolutionPathHash}, []string{IdentityStoreV2SpecResolutionCachePrefix + "/" + m.NameHash + "/" + m.ResolutionPathHash})
		return plan.with(set, []string{m.NameHash}, nil)
	default:
		return IdentityBlockSTMPlanV2{}, fmt.Errorf("unsupported identity v2 BlockSTM message %T", msg)
	}
}

func (p IdentityBlockSTMPlanV2) with(set IdentityAccessSet, nameHashes []string, err error) (IdentityBlockSTMPlanV2, error) {
	if err != nil {
		return IdentityBlockSTMPlanV2{}, err
	}
	p.AccessSet = newIdentityAccessSet(set.Reads, set.Writes)
	p.NameHashes = sortedUniqueStrings(nameHashes)
	return p, nil
}

func IdentityBlockSTMConflictClassifyV2(left IdentityBlockSTMPlanV2, right IdentityBlockSTMPlanV2) IdentityBlockSTMConflictClassV2 {
	if !left.AccessSet.Conflicts(right.AccessSet) {
		return IdentityBlockSTMConflictNoneV2
	}
	if left.ConflictClass == IdentityBlockSTMConflictParentPolicyChildCreateV2 || right.ConflictClass == IdentityBlockSTMConflictParentPolicyChildCreateV2 {
		return IdentityBlockSTMConflictParentPolicyChildCreateV2
	}
	if left.ConflictClass == IdentityBlockSTMConflictTransferResolverUpdateV2 || right.ConflictClass == IdentityBlockSTMConflictTransferResolverUpdateV2 {
		return IdentityBlockSTMConflictTransferResolverUpdateV2
	}
	if left.ConflictClass == IdentityBlockSTMConflictReversePrimaryResolverV2 || right.ConflictClass == IdentityBlockSTMConflictReversePrimaryResolverV2 {
		return IdentityBlockSTMConflictReversePrimaryResolverV2
	}
	if left.ConflictClass == IdentityBlockSTMConflictAuctionFinalizeLateRevealV2 || right.ConflictClass == IdentityBlockSTMConflictAuctionFinalizeLateRevealV2 {
		return IdentityBlockSTMConflictAuctionFinalizeLateRevealV2
	}
	if sameStringSetIntersectsV2(left.NameHashes, right.NameHashes) {
		return IdentityBlockSTMConflictSameNameV2
	}
	return IdentityBlockSTMConflictParentChildUnknownV2
}

func IdentityBlockSTMValidateVersionedUpdateV2(currentVersion uint64, expectedVersion uint64) error {
	return ValidateResolverRecordVersionForUpdateV2(currentVersion, expectedVersion)
}

func IdentityBlockSTMBatchResolverAccessSetV2(msg MsgBatchUpdateResolversV2) (IdentityAccessSet, []string, error) {
	if err := msg.ValidateBasic(); err != nil {
		return IdentityAccessSet{}, nil, err
	}
	reads := make([]string, 0, len(msg.Updates))
	writes := make([]string, 0, len(msg.Updates))
	hashes := make([]string, 0, len(msg.Updates))
	for _, update := range msg.Updates {
		hash, err := validateIdentityTxNameOrHashV2(update.Name, update.NameHash)
		if err != nil {
			return IdentityAccessSet{}, nil, err
		}
		domainKey, resolverKey, err := blockSTMNameResolverKeysV2(update.Name, hash)
		if err != nil {
			return IdentityAccessSet{}, nil, err
		}
		reads = append(reads, domainKey, resolverKey)
		writes = append(writes, resolverKey)
		hashes = append(hashes, hash)
	}
	return newIdentityAccessSet(reads, writes), sortedUniqueStrings(hashes), nil
}

func blockSTMNameLifecycleSetV2(name string, nameHash string, owner []byte, includeNFT bool, includeOwnerIndex bool) (IdentityAccessSet, []string, error) {
	hash, err := validateIdentityTxNameOrHashV2(name, nameHash)
	if err != nil {
		return IdentityAccessSet{}, nil, err
	}
	domainKey, err := IdentityStoreV2SpecDomainKeyByHash(hash)
	if err != nil {
		return IdentityAccessSet{}, nil, err
	}
	reads := []string{domainKey}
	writes := []string{domainKey}
	if includeNFT && name != "" {
		nftKey, err := IdentityStoreV2SpecNFTBindingByNameKey(name)
		if err != nil {
			return IdentityAccessSet{}, nil, err
		}
		reads = append(reads, nftKey)
		writes = append(writes, nftKey)
	}
	if includeOwnerIndex && len(owner) != 0 && name != "" {
		ownerKey, err := IdentityStoreV2SpecOwnerIndexKey(owner, name)
		if err != nil {
			return IdentityAccessSet{}, nil, err
		}
		writes = append(writes, ownerKey)
	}
	return newIdentityAccessSet(reads, writes), []string{hash}, nil
}

func blockSTMResolverUpdateSetV2(name string, nameHash string, touchPrimary bool) (IdentityAccessSet, []string, error) {
	hash, err := validateIdentityTxNameOrHashV2(name, nameHash)
	if err != nil {
		return IdentityAccessSet{}, nil, err
	}
	domainKey, resolverKey, err := blockSTMNameResolverKeysV2(name, hash)
	if err != nil {
		return IdentityAccessSet{}, nil, err
	}
	reads := []string{domainKey, resolverKey}
	writes := []string{resolverKey}
	if touchPrimary {
		writes = append(writes, IdentityStoreV2SpecReversePrefix+"/primary/"+hash)
	}
	return newIdentityAccessSet(reads, writes), []string{hash}, nil
}

func blockSTMReverseSetV2(record ReverseResolutionRecordV2) (IdentityAccessSet, []string, error) {
	if err := ValidateReverseResolutionRecordV2Format(record); err != nil {
		return IdentityAccessSet{}, nil, err
	}
	reverseKey, err := IdentityStoreV2SpecReverseKey(record.Address)
	if err != nil {
		return IdentityAccessSet{}, nil, err
	}
	return newIdentityAccessSet([]string{reverseKey}, []string{reverseKey}), []string{record.NameHash}, nil
}

func blockSTMReverseVerifySetV2(record ReverseResolutionRecordV2) (IdentityAccessSet, []string, error) {
	if err := ValidateReverseResolutionRecordV2Format(record); err != nil {
		return IdentityAccessSet{}, nil, err
	}
	reverseKey, err := IdentityStoreV2SpecReverseKey(record.Address)
	if err != nil {
		return IdentityAccessSet{}, nil, err
	}
	domainKey, err := IdentityStoreV2SpecDomainKeyByHash(record.NameHash)
	if err != nil {
		return IdentityAccessSet{}, nil, err
	}
	resolverKey := IdentityStoreV2SpecResolversPrefix + "/" + record.NameHash
	primaryGuard := IdentityStoreV2SpecReversePrefix + "/primary/" + record.NameHash
	return newIdentityAccessSet([]string{reverseKey, domainKey, resolverKey, primaryGuard}, []string{reverseKey}), []string{record.NameHash}, nil
}

func blockSTMSubdomainCreateSetV2(parentName string, parentHash string, label string) (IdentityAccessSet, []string, error) {
	hash, err := validateIdentityTxNameOrHashV2(parentName, parentHash)
	if err != nil {
		return IdentityAccessSet{}, nil, err
	}
	if err := validateDomainLabel(label); err != nil {
		return IdentityAccessSet{}, nil, err
	}
	parentKey, err := IdentityStoreV2SpecDomainKeyByHash(hash)
	if err != nil {
		return IdentityAccessSet{}, nil, err
	}
	subdomainKey, err := IdentityStoreV2SpecSubdomainKey(parentName, label)
	if err != nil {
		return IdentityAccessSet{}, nil, err
	}
	parentPolicyKey := blockSTMParentPolicyGuardKeyV2(hash)
	return newIdentityAccessSet([]string{parentKey, parentPolicyKey}, []string{subdomainKey}), []string{hash}, nil
}

func blockSTMDelegationSetV2(record DelegationRecordV2) (IdentityAccessSet, []string, error) {
	if err := ValidateDelegationRecordV2(record); err != nil {
		return IdentityAccessSet{}, nil, err
	}
	key := fmt.Sprintf("%s/%s/%s/%s", IdentityStoreV2SpecDelegationsPrefix, record.NameHash, hex.EncodeToString(record.Delegate), record.Scope)
	policyKey := blockSTMParentPolicyGuardKeyV2(record.NameHash)
	return newIdentityAccessSet([]string{key}, []string{key, policyKey}), []string{record.NameHash}, nil
}

func blockSTMDelegationRevokeSetV2(name string, nameHash string, delegate []byte, scope DelegationScopeV2) (IdentityAccessSet, []string, error) {
	hash, err := validateIdentityTxNameOrHashV2(name, nameHash)
	if err != nil {
		return IdentityAccessSet{}, nil, err
	}
	if err := validateSpecAddress("identity v2 BlockSTM delegation delegate", delegate); err != nil {
		return IdentityAccessSet{}, nil, err
	}
	if err := validateDelegationScopeV2(scope); err != nil {
		return IdentityAccessSet{}, nil, err
	}
	key := fmt.Sprintf("%s/%s/%s/%s", IdentityStoreV2SpecDelegationsPrefix, hash, hex.EncodeToString(delegate), scope)
	policyKey := blockSTMParentPolicyGuardKeyV2(hash)
	return newIdentityAccessSet([]string{key}, []string{key, policyKey}), []string{hash}, nil
}

func blockSTMAuctionSetV2(auctionID string, nameHash string) (IdentityAccessSet, []string, error) {
	if err := validateIdentityTxAuctionIDOrNameHashV2(auctionID, nameHash); err != nil {
		return IdentityAccessSet{}, nil, err
	}
	key := IdentityStoreV2SpecAuctionsPrefix + "/" + auctionID
	return newIdentityAccessSet([]string{key}, []string{key}), []string{nameHash}, nil
}

func blockSTMAuctionBidSetV2(auctionID string, nameHash string, bidder []byte, commitmentHash string) (IdentityAccessSet, []string, error) {
	set, hashes, err := blockSTMAuctionSetV2(auctionID, nameHash)
	if err != nil {
		return IdentityAccessSet{}, nil, err
	}
	if err := validateSpecAddress("identity v2 BlockSTM auction bidder", bidder); err != nil {
		return IdentityAccessSet{}, nil, err
	}
	if err := validateHexHash("identity v2 BlockSTM bid commitment", commitmentHash); err != nil {
		return IdentityAccessSet{}, nil, err
	}
	bidKey := fmt.Sprintf("%s/%s/bids/%s/%s", IdentityStoreV2SpecAuctionsPrefix, auctionID, hex.EncodeToString(bidder), commitmentHash)
	return newIdentityAccessSet(append(set.Reads, bidKey), append(set.Writes, bidKey)), hashes, nil
}

func blockSTMBatchRenewSetV2(msg MsgBatchRenewDomainsV2) (IdentityAccessSet, []string, error) {
	if err := msg.ValidateBasic(); err != nil {
		return IdentityAccessSet{}, nil, err
	}
	reads := make([]string, 0, len(msg.Renewals))
	writes := make([]string, 0, len(msg.Renewals))
	hashes := make([]string, 0, len(msg.Renewals))
	for _, renewal := range msg.Renewals {
		hash, err := validateIdentityTxNameOrHashV2(renewal.Name, renewal.NameHash)
		if err != nil {
			return IdentityAccessSet{}, nil, err
		}
		key, err := IdentityStoreV2SpecDomainKeyByHash(hash)
		if err != nil {
			return IdentityAccessSet{}, nil, err
		}
		reads = append(reads, key)
		writes = append(writes, key)
		hashes = append(hashes, hash)
	}
	return newIdentityAccessSet(reads, writes), sortedUniqueStrings(hashes), nil
}

func blockSTMNameResolverKeysV2(name string, nameHash string) (string, string, error) {
	if name != "" {
		domainKey, err := IdentityStoreV2SpecDomainKey(name)
		if err != nil {
			return "", "", err
		}
		resolverKey, err := IdentityStoreV2SpecResolverKey(name)
		if err != nil {
			return "", "", err
		}
		return domainKey, resolverKey, nil
	}
	if err := validateHexHash("identity v2 BlockSTM name hash", nameHash); err != nil {
		return "", "", err
	}
	return IdentityStoreV2SpecDomainsPrefix + "/" + nameHash, IdentityStoreV2SpecResolversPrefix + "/" + nameHash, nil
}

func resolverPatchTouchesPrimaryV2(patch ResolverPatch) bool {
	return len(patch.Primary) != 0 || patch.ClearPrimary
}

func sameStringSetIntersectsV2(left []string, right []string) bool {
	seen := stringSet(left)
	for _, value := range right {
		if _, found := seen[value]; found {
			return true
		}
	}
	return false
}

func blockSTMParentPolicyGuardKeyV2(nameHash string) string {
	return IdentityStoreV2SpecDelegationsPrefix + "/" + nameHash + "/policy"
}

func (m MsgStartAuctionV2) AuctionID() string {
	if m.Name != "" {
		return identityHash("identity-v2-auction", m.Name, fmt.Sprintf("%020d", m.Auth.Nonce))
	}
	return identityHash("identity-v2-auction", m.NameHash, fmt.Sprintf("%020d", m.Auth.Nonce))
}
