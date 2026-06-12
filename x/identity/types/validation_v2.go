package types

import (
	"errors"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	NameNormalizationVersionV2		uint64	= 1
	MinSupportedNameNormalizationVersionV2	uint64	= NameNormalizationVersionV2
	MaxSupportedNameNormalizationVersionV2	uint64	= NameNormalizationVersionV2

	ResolverRecoveryMetadataKeyV2	= "recovery"
)

type NameNormalizationResultV2 struct {
	Input		string
	NormalizedName	string
	NameHash	string
	Version		uint64
	Labels		[]string
}

func ValidateNameNormalizationVersionV2(version uint64) error {
	if version < MinSupportedNameNormalizationVersionV2 || version > MaxSupportedNameNormalizationVersionV2 {
		return fmt.Errorf("unsupported identity name normalization version %d", version)
	}
	return nil
}

func NormalizeAETDomainVersioned(name string, version uint64) (NameNormalizationResultV2, error) {
	if err := ValidateNameNormalizationVersionV2(version); err != nil {
		return NameNormalizationResultV2{}, err
	}
	if name == "" {
		return NameNormalizationResultV2{}, errors.New("identity v2 domain name is required")
	}
	if strings.TrimSpace(name) != name {
		return NameNormalizationResultV2{}, errors.New("identity v2 domain name must not have surrounding whitespace")
	}
	if !isASCIIStringV2(name) {
		return NameNormalizationResultV2{}, errors.New("identity v2 domain name must be ASCII")
	}
	if name != strings.ToLower(name) {
		return NameNormalizationResultV2{}, errors.New("identity v2 domain name must be lowercase")
	}
	if !strings.HasSuffix(name, DomainTLD) {
		return NameNormalizationResultV2{}, fmt.Errorf("identity v2 domain name must end with %s", DomainTLD)
	}
	if len(name) > MaxDomainFullBytes {
		return NameNormalizationResultV2{}, fmt.Errorf("identity v2 domain name must be <= %d bytes", MaxDomainFullBytes)
	}

	labelsPart := strings.TrimSuffix(name, DomainTLD)
	if labelsPart == "" {
		return NameNormalizationResultV2{}, errors.New("identity v2 domain label is required")
	}
	if strings.HasPrefix(labelsPart, ".") || strings.HasSuffix(labelsPart, ".") {
		return NameNormalizationResultV2{}, errors.New("identity v2 domain has leading or trailing separator")
	}

	labels := strings.Split(labelsPart, ".")
	if len(labels) > MaxDomainLabels {
		return NameNormalizationResultV2{}, fmt.Errorf("identity v2 domain must not exceed %d labels", MaxDomainLabels)
	}
	for _, label := range labels {
		if label == "" {
			return NameNormalizationResultV2{}, errors.New("identity v2 domain contains empty label")
		}
		if err := validateDomainLabel(label); err != nil {
			return NameNormalizationResultV2{}, err
		}
		if err := validateIdentityV2LabelSpoofing(label); err != nil {
			return NameNormalizationResultV2{}, err
		}
	}

	nameHash, err := DomainRecordV2NameHash(name)
	if err != nil {
		return NameNormalizationResultV2{}, err
	}
	return NameNormalizationResultV2{
		Input:		name,
		NormalizedName:	name,
		NameHash:	nameHash,
		Version:	version,
		Labels:		append([]string(nil), labels...),
	}, nil
}

func MigrateNameNormalizationVersionV2(result NameNormalizationResultV2, targetVersion uint64) (NameNormalizationResultV2, error) {
	if err := ValidateNameNormalizationVersionV2(result.Version); err != nil {
		return NameNormalizationResultV2{}, err
	}
	if err := ValidateNameNormalizationVersionV2(targetVersion); err != nil {
		return NameNormalizationResultV2{}, err
	}
	if result.Version == targetVersion {
		return NormalizeAETDomainVersioned(result.NormalizedName, targetVersion)
	}
	return NameNormalizationResultV2{}, fmt.Errorf("identity v2 name normalization migration from %d to %d is not supported", result.Version, targetVersion)
}

func ValidatePreTransferOwnershipV2(record DomainRecordV2, binding DomainNFTBinding, actor sdk.AccAddress, height uint64) error {
	if height == 0 {
		return errors.New("identity v2 ownership transfer height is required")
	}
	if err := validateSpecAddress("identity v2 ownership transfer actor", actor); err != nil {
		return err
	}
	if !addressesEqual(actor, record.Owner) {
		return errors.New("identity v2 ownership transfer requires current registry owner")
	}
	if err := validateTransferBindingAgreementV2(record, binding, height); err != nil {
		return err
	}
	if requiresUnexpiredDomainRecordV2(record.Status) && record.ExpiryHeight <= height {
		return errors.New("identity v2 ownership transfer requires unexpired active domain")
	}
	return nil
}

func ValidatePostTransferOwnershipV2(beforeRecord DomainRecordV2, beforeBinding DomainNFTBinding, afterRecord DomainRecordV2, afterBinding DomainNFTBinding, newOwner sdk.AccAddress, height uint64) error {
	if err := validateSpecAddress("identity v2 ownership transfer new owner", newOwner); err != nil {
		return err
	}
	if beforeRecord.NameHash != afterRecord.NameHash || beforeBinding.NameHash != afterBinding.NameHash || afterRecord.NameHash != afterBinding.NameHash {
		return errors.New("identity v2 ownership transfer changed name_hash")
	}
	if beforeRecord.NFTClassID != afterRecord.NFTClassID || beforeRecord.NFTItemID != afterRecord.NFTItemID {
		return errors.New("identity v2 ownership transfer changed nft identity")
	}
	if !addressesEqual(afterRecord.Owner, newOwner) || !addressesEqual(afterBinding.Owner, newOwner) {
		return errors.New("identity v2 ownership transfer must atomically update registry and nft owner")
	}
	if afterRecord.UpdatedAtHeight != height || afterBinding.LastVerifiedHeight != height {
		return errors.New("identity v2 ownership transfer height mismatch")
	}
	return validateTransferBindingAgreementV2(afterRecord, afterBinding, height)
}

func TransferDomainNFTBindingWithInvariantsV2(record DomainRecordV2, binding DomainNFTBinding, actor sdk.AccAddress, newOwner sdk.AccAddress, height uint64) (DomainRecordV2, DomainNFTBinding, error) {
	if err := ValidatePreTransferOwnershipV2(record, binding, actor, height); err != nil {
		return DomainRecordV2{}, DomainNFTBinding{}, err
	}
	nextRecord, nextBinding, err := TransferDomainNFTBindingAtomic(record, binding, newOwner, height)
	if err != nil {
		return DomainRecordV2{}, DomainNFTBinding{}, err
	}
	if err := ValidatePostTransferOwnershipV2(record, binding, nextRecord, nextBinding, newOwner, height); err != nil {
		return DomainRecordV2{}, DomainNFTBinding{}, err
	}
	return nextRecord, nextBinding, nil
}

func RepairDomainNFTBindingInternalFailureV2(record DomainRecordV2, binding DomainNFTBinding, nftModuleOwner sdk.AccAddress, height uint64) (DomainRecordV2, DomainNFTBinding, error) {
	if record.NameHash != binding.NameHash {
		return DomainRecordV2{}, DomainNFTBinding{}, errors.New("identity v2 internal repair binding name_hash mismatch")
	}
	if addressesEqual(record.Owner, binding.Owner) && addressesEqual(record.Owner, nftModuleOwner) {
		return DomainRecordV2{}, DomainNFTBinding{}, errors.New("identity v2 internal repair requires broken nft binding")
	}
	return RepairDomainNFTBinding(record, binding, nftModuleOwner, nftModuleOwner, height)
}

func ValidateResolverUpdateAuthorizationV2(domain DomainRecordV2, actor sdk.AccAddress, delegation *DelegationRecordV2, recordKey string, height uint64) error {
	if err := validateSpecAddress("identity v2 resolver update actor", actor); err != nil {
		return err
	}
	if recordKey == "" {
		return errors.New("identity v2 resolver update record key is required")
	}
	if domain.Flags&DomainRecordV2FlagRestricted != 0 {
		return errors.New("identity v2 broken nft binding blocks resolver changes until repaired")
	}
	if isExpiredForResolverUpdateV2(domain, height) {
		if !addressesEqual(actor, domain.Owner) || recordKey != ResolverRecoveryMetadataKeyV2 {
			return errors.New("identity v2 expired domain owner cannot update resolver except recovery metadata")
		}
		return nil
	}
	if addressesEqual(actor, domain.Owner) {
		return nil
	}
	if delegation == nil {
		return errors.New("identity v2 resolver update requires owner or delegated permission")
	}
	if !addressesEqual(actor, delegation.Delegate) {
		return errors.New("identity v2 resolver update delegate mismatch")
	}
	if delegation.NameHash != domain.NameHash {
		return errors.New("identity v2 resolver update delegation name_hash mismatch")
	}
	return ValidateDelegationRecordV2Use(*delegation, DelegationScopeResolverUpdate, recordKey, recordKey, 0, height)
}

func ValidateSubdomainCreationAuthorizationV2(parent DomainRecordV2, actor sdk.AccAddress, delegation *DelegationRecordV2, childLabel string, subtreeDepth uint8, height uint64) error {
	if err := validateSpecAddress("identity v2 subdomain create actor", actor); err != nil {
		return err
	}
	if err := validateDomainLabel(childLabel); err != nil {
		return err
	}
	if isExpiredForResolverUpdateV2(parent, height) {
		return errors.New("identity v2 subdomain creation requires active parent")
	}
	if addressesEqual(actor, parent.Owner) {
		return nil
	}
	if delegation == nil {
		return errors.New("identity v2 subdomain creation requires parent owner or scoped delegate")
	}
	if !addressesEqual(actor, delegation.Delegate) {
		return errors.New("identity v2 subdomain creation delegate mismatch")
	}
	if delegation.NameHash != parent.NameHash {
		return errors.New("identity v2 subdomain creation delegation name_hash mismatch")
	}
	return ValidateDelegationRecordV2Use(*delegation, DelegationScopeSubdomainCreate, "create", childLabel, subtreeDepth, height)
}

func validateTransferBindingAgreementV2(record DomainRecordV2, binding DomainNFTBinding, height uint64) error {
	if record.NameHash != binding.NameHash {
		return errors.New("identity v2 ownership binding name_hash mismatch")
	}
	if record.NFTClassID != binding.NFTClassID || record.NFTItemID != binding.NFTItemID {
		return errors.New("identity v2 ownership binding nft item mismatch")
	}
	if requiresDomainRecordV2NFTAgreement(record.Status) && !addressesEqual(record.Owner, binding.Owner) {
		return errors.New("identity v2 registry owner must equal nft owner")
	}
	return ValidateDomainNFTBinding(binding, DomainNFTBindingContext{
		RegistryOwner:	record.Owner,
		NFTModuleOwner:	binding.Owner,
		CurrentHeight:	height,
	})
}

func validateIdentityV2LabelSpoofing(label string) error {
	if identityV2ReservedLabel(label) {
		return fmt.Errorf("identity v2 domain label %q is reserved", label)
	}
	if strings.HasPrefix(label, "xn--") {
		return errors.New("identity v2 domain label must not use punycode spoofing prefix")
	}
	if strings.Contains(label, "--") {
		return errors.New("identity v2 domain label must not contain repeated hyphen spoofing pattern")
	}
	if strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") || strings.HasPrefix(label, "_") || strings.HasSuffix(label, "_") {
		return errors.New("identity v2 domain label must not start or end with separator")
	}
	return nil
}

func identityV2ReservedLabel(label string) bool {
	switch label {
	case "aet", "admin", "root", "null", "undefined":
		return true
	default:
		return false
	}
}

func isASCIIStringV2(value string) bool {
	for i := 0; i < len(value); i++ {
		if value[i] > 0x7f {
			return false
		}
	}
	return true
}

func isExpiredForResolverUpdateV2(domain DomainRecordV2, height uint64) bool {
	if domain.Status == DomainRecordV2Expired || domain.Status == DomainRecordV2GraceLocked || domain.Status == DomainRecordV2Released || domain.Status == DomainRecordV2Revoked {
		return true
	}
	return requiresUnexpiredDomainRecordV2(domain.Status) && domain.ExpiryHeight <= height
}
