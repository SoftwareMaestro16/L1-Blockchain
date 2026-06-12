package types

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/app/addressing"
)

const (
	MaxResolverLabels		= 8
	MaxResolverRecords		= 16
	MaxResolverKeyBytes		= 32
	MaxResolverMetadataBytes	= 512

	ResolverKeyPrimary	= "primary"
	ResolverKeyMetadata	= "metadata"
	ResolverKeyWallet	= "wallet"
	ResolverKeyContract	= "contract"
	ResolverKeyMultisig	= "multisig"
	ResolverKeyNFT		= "nft"
	ResolverKeyDEX		= "dex"

	ResolverEventSet	= "resolver_set"
	ResolverEventChanged	= "resolver_changed"
	ResolverEventReverse	= "resolver_reverse_set"
)

type ResolverRecord struct {
	Domain		string
	Owner		sdk.AccAddress
	Primary		sdk.AccAddress
	Contract	sdk.AccAddress
	ZoneEndpoint	string
	Records		map[string]sdk.AccAddress
	Metadata	[]byte
	UpdatedAtUnix	int64
}

type ResolverUpdate struct {
	Domain		string
	Primary		sdk.AccAddress
	Contract	sdk.AccAddress
	ZoneEndpoint	string
	Records		map[string]sdk.AccAddress
	Metadata	[]byte
	UpdatedAtUnix	int64
}

type ResolverGrant struct {
	Domain		string
	Owner		sdk.AccAddress
	Manager		sdk.AccAddress
	Keys		[]string
	ExpiresAtUnix	int64
}

type ResolverEvent struct {
	Type		string
	Domain		string
	Actor		sdk.AccAddress
	Keys		[]string
	UpdatedAtUnix	int64
}

type ReverseResolverRecord struct {
	Address		sdk.AccAddress
	Domain		string
	UpdatedAtUnix	int64
}

func ApplyResolverUpdate(existing *ResolverRecord, domainRecord DomainRecord, actor sdk.AccAddress, update ResolverUpdate, grant *ResolverGrant, nowUnix int64) (ResolverRecord, ResolverEvent, error) {
	if err := ValidateDomainUsableForResolver(domainRecord, update.Domain, nowUnix); err != nil {
		return ResolverRecord{}, ResolverEvent{}, err
	}
	changedKeys := ResolverUpdateKeys(update)
	if err := CanUpdateResolver(actor, domainRecord.Owner, update.Domain, changedKeys, grant, nowUnix); err != nil {
		return ResolverRecord{}, ResolverEvent{}, err
	}
	normalizedDomain, err := NormalizeResolverDomain(update.Domain)
	if err != nil {
		return ResolverRecord{}, ResolverEvent{}, err
	}
	record := ResolverRecord{
		Domain:		normalizedDomain,
		Owner:		append(sdk.AccAddress(nil), domainRecord.Owner...),
		Primary:	cloneAddress(update.Primary),
		Contract:	cloneAddress(update.Contract),
		ZoneEndpoint:	strings.TrimSpace(update.ZoneEndpoint),
		Records:	cloneResolverRecords(update.Records),
		Metadata:	append([]byte(nil), update.Metadata...),
		UpdatedAtUnix:	update.UpdatedAtUnix,
	}
	if existing != nil {
		if err := ValidateResolverRecordForDomain(*existing, domainRecord, nowUnix); err != nil {
			return ResolverRecord{}, ResolverEvent{}, err
		}
		if existing.Domain != record.Domain {
			return ResolverRecord{}, ResolverEvent{}, errors.New("resolver update domain cannot change existing record domain")
		}
	}
	if err := ValidateResolverRecordForDomain(record, domainRecord, nowUnix); err != nil {
		return ResolverRecord{}, ResolverEvent{}, err
	}
	eventType := ResolverEventSet
	if existing != nil {
		eventType = ResolverEventChanged
	}
	return record, ResolverEvent{
		Type:		eventType,
		Domain:		record.Domain,
		Actor:		cloneAddress(actor),
		Keys:		changedKeys,
		UpdatedAtUnix:	update.UpdatedAtUnix,
	}, nil
}

func ResolvePaymentTarget(record ResolverRecord, domainRecord DomainRecord, nowUnix int64) (sdk.AccAddress, error) {
	if err := ValidateResolverRecordForDomain(record, domainRecord, nowUnix); err != nil {
		return nil, err
	}
	if len(record.Primary) == 0 {
		return nil, errors.New("domain not resolved")
	}
	return cloneAddress(record.Primary), nil
}

func SetReverseResolution(domainRecord DomainRecord, resolver ResolverRecord, actor sdk.AccAddress, address sdk.AccAddress, nowUnix int64) (ReverseResolverRecord, ResolverEvent, error) {
	if err := ValidateResolverRecordForDomain(resolver, domainRecord, nowUnix); err != nil {
		return ReverseResolverRecord{}, ResolverEvent{}, err
	}
	if len(actor) == 0 || len(address) == 0 {
		return ReverseResolverRecord{}, ResolverEvent{}, errors.New("reverse resolution actor and address are required")
	}
	if err := addressing.RejectZeroAddress("reverse resolution actor", actor); err != nil {
		return ReverseResolverRecord{}, ResolverEvent{}, err
	}
	if err := addressing.RejectZeroAddress("reverse resolution address", address); err != nil {
		return ReverseResolverRecord{}, ResolverEvent{}, err
	}
	if !bytes.Equal(actor, address) {
		return ReverseResolverRecord{}, ResolverEvent{}, errors.New("reverse resolution actor must control address")
	}
	if !ResolverRecordContainsAddress(resolver, address) {
		return ReverseResolverRecord{}, ResolverEvent{}, errors.New("resolver does not point to reverse address")
	}
	reverse := ReverseResolverRecord{
		Address:	cloneAddress(address),
		Domain:		resolver.Domain,
		UpdatedAtUnix:	nowUnix,
	}
	event := ResolverEvent{
		Type:		ResolverEventReverse,
		Domain:		resolver.Domain,
		Actor:		cloneAddress(actor),
		Keys:		[]string{ResolverKeyPrimary},
		UpdatedAtUnix:	nowUnix,
	}
	return reverse, event, nil
}

func ValidateResolverRecordForDomain(record ResolverRecord, domainRecord DomainRecord, nowUnix int64) error {
	if err := ValidateResolverRecord(record); err != nil {
		return err
	}
	if err := ValidateDomainUsableForResolver(domainRecord, record.Domain, nowUnix); err != nil {
		return err
	}
	if !bytes.Equal(record.Owner, domainRecord.Owner) {
		return errors.New("resolver owner must match registry owner")
	}
	return nil
}

func ValidateResolverRecord(record ResolverRecord) error {
	if err := ValidateResolverDomain(record.Domain); err != nil {
		return err
	}
	if len(record.Owner) == 0 {
		return errors.New("resolver owner is required")
	}
	if err := addressing.RejectZeroAddress("resolver owner", record.Owner); err != nil {
		return err
	}
	if len(record.Primary) > 0 {
		if err := addressing.RejectZeroAddress("resolver primary", record.Primary); err != nil {
			return err
		}
	}
	if len(record.Contract) > 0 {
		if err := addressing.RejectZeroAddress("resolver contract", record.Contract); err != nil {
			return err
		}
	}
	if strings.TrimSpace(record.ZoneEndpoint) != record.ZoneEndpoint {
		return errors.New("resolver zone endpoint must not have surrounding whitespace")
	}
	if len(record.ZoneEndpoint) > MaxResolverMetadataBytes {
		return fmt.Errorf("resolver zone endpoint must not exceed %d bytes", MaxResolverMetadataBytes)
	}
	if len(record.Records) > MaxResolverRecords {
		return fmt.Errorf("resolver records must not exceed %d", MaxResolverRecords)
	}
	for _, key := range sortedResolverKeys(record.Records) {
		if err := ValidateResolverKey(key); err != nil {
			return err
		}
		target := record.Records[key]
		if len(target) == 0 {
			return fmt.Errorf("resolver target for %q is required", key)
		}
		if err := addressing.RejectZeroAddress("resolver target "+key, target); err != nil {
			return err
		}
	}
	if len(record.Metadata) > MaxResolverMetadataBytes {
		return fmt.Errorf("resolver metadata must not exceed %d bytes", MaxResolverMetadataBytes)
	}
	if record.UpdatedAtUnix < 0 {
		return errors.New("resolver updated_at must be non-negative")
	}
	return nil
}

func ValidateDomainUsableForResolver(record DomainRecord, resolverDomain string, nowUnix int64) error {
	if err := ValidateDomainRecord(record); err != nil {
		return err
	}
	if record.Status != DomainStatusActive {
		return errors.New("domain is not active")
	}
	if record.ExpiryUnix <= nowUnix {
		return errors.New("domain is expired")
	}
	base, err := BaseDomainFromResolverDomain(resolverDomain)
	if err != nil {
		return err
	}
	if base != record.Name {
		return errors.New("resolver domain base does not match registry domain")
	}
	return nil
}

func CanUpdateResolver(actor sdk.AccAddress, owner sdk.AccAddress, domain string, keys []string, grant *ResolverGrant, nowUnix int64) error {
	if len(actor) == 0 {
		return errors.New("resolver update actor is required")
	}
	if err := addressing.RejectZeroAddress("resolver update actor", actor); err != nil {
		return err
	}
	if bytes.Equal(actor, owner) {
		return nil
	}
	if grant == nil {
		return errors.New("unauthorized resolver update")
	}
	if err := ValidateResolverGrant(*grant, nowUnix); err != nil {
		return err
	}
	normalizedDomain, err := NormalizeResolverDomain(domain)
	if err != nil {
		return err
	}
	if grant.Domain != normalizedDomain || !bytes.Equal(grant.Owner, owner) || !bytes.Equal(grant.Manager, actor) {
		return errors.New("unauthorized resolver update")
	}
	allowed := make(map[string]struct{}, len(grant.Keys))
	for _, key := range grant.Keys {
		allowed[key] = struct{}{}
	}
	for _, key := range keys {
		if _, ok := allowed[key]; !ok {
			return fmt.Errorf("resolver grant does not allow key %q", key)
		}
	}
	return nil
}

func ValidateResolverGrant(grant ResolverGrant, nowUnix int64) error {
	if err := ValidateResolverDomain(grant.Domain); err != nil {
		return err
	}
	if len(grant.Owner) == 0 || len(grant.Manager) == 0 {
		return errors.New("resolver grant owner and manager are required")
	}
	if err := addressing.RejectZeroAddress("resolver grant owner", grant.Owner); err != nil {
		return err
	}
	if err := addressing.RejectZeroAddress("resolver grant manager", grant.Manager); err != nil {
		return err
	}
	if grant.ExpiresAtUnix <= nowUnix {
		return errors.New("resolver grant expired")
	}
	if len(grant.Keys) == 0 {
		return errors.New("resolver grant keys are required")
	}
	for _, key := range grant.Keys {
		if err := ValidateResolverGrantKey(key); err != nil {
			return err
		}
	}
	return nil
}

func ResolverUpdateKeys(update ResolverUpdate) []string {
	keys := make([]string, 0, len(update.Records)+2)
	if len(update.Primary) > 0 {
		keys = append(keys, ResolverKeyPrimary)
	}
	if len(update.Contract) > 0 {
		keys = append(keys, ResolverKeyContract)
	}
	if strings.TrimSpace(update.ZoneEndpoint) != "" {
		keys = append(keys, "zone_endpoint")
	}
	if len(update.Metadata) > 0 {
		keys = append(keys, ResolverKeyMetadata)
	}
	keys = append(keys, sortedResolverKeys(update.Records)...)
	sort.Strings(keys)
	return keys
}

func ValidateResolverGrantKey(key string) error {
	if key == ResolverKeyPrimary || key == ResolverKeyMetadata {
		return nil
	}
	return ValidateResolverKey(key)
}

func ValidateResolverKey(key string) error {
	if key == "" {
		return errors.New("resolver key is required")
	}
	if len(key) > MaxResolverKeyBytes {
		return fmt.Errorf("resolver key must not exceed %d bytes", MaxResolverKeyBytes)
	}
	for i := 0; i < len(key); i++ {
		c := key[i]
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_' {
			continue
		}
		return fmt.Errorf("resolver key contains unsupported character %q", c)
	}
	return nil
}

func NormalizeResolverDomain(domain string) (string, error) {
	trimmed := strings.TrimSpace(domain)
	if trimmed == "" {
		return "", errors.New("resolver domain is required")
	}
	lower := strings.ToLower(trimmed)
	if trimmed != lower {
		return "", errors.New("resolver domain must be normalized lowercase ASCII")
	}
	if !strings.HasSuffix(lower, DomainTLD) {
		return "", fmt.Errorf("resolver domain must end with %s", DomainTLD)
	}
	labelsPart := strings.TrimSuffix(lower, DomainTLD)
	if labelsPart == "" {
		return "", errors.New("resolver domain label is required")
	}
	labels := strings.Split(labelsPart, ".")
	if len(labels) > MaxResolverLabels {
		return "", fmt.Errorf("resolver domain must not exceed %d labels", MaxResolverLabels)
	}
	for _, label := range labels {
		if err := validateDomainLabel(label); err != nil {
			return "", err
		}
	}
	return lower, nil
}

func ValidateResolverDomain(domain string) error {
	_, err := NormalizeResolverDomain(domain)
	return err
}

func BaseDomainFromResolverDomain(domain string) (string, error) {
	normalized, err := NormalizeResolverDomain(domain)
	if err != nil {
		return "", err
	}
	labelsPart := strings.TrimSuffix(normalized, DomainTLD)
	labels := strings.Split(labelsPart, ".")
	return labels[len(labels)-1], nil
}

func ResolverRecordContainsAddress(record ResolverRecord, address sdk.AccAddress) bool {
	if bytes.Equal(record.Primary, address) {
		return true
	}
	for _, key := range sortedResolverKeys(record.Records) {
		if bytes.Equal(record.Records[key], address) {
			return true
		}
	}
	return false
}

func sortedResolverKeys(records map[string]sdk.AccAddress) []string {
	keys := make([]string, 0, len(records))
	for key := range records {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func cloneResolverRecords(records map[string]sdk.AccAddress) map[string]sdk.AccAddress {
	if len(records) == 0 {
		return nil
	}
	out := make(map[string]sdk.AccAddress, len(records))
	for _, key := range sortedResolverKeys(records) {
		out[key] = cloneAddress(records[key])
	}
	return out
}

func cloneAddress(addr sdk.AccAddress) sdk.AccAddress {
	if len(addr) == 0 {
		return nil
	}
	return append(sdk.AccAddress(nil), addr...)
}
