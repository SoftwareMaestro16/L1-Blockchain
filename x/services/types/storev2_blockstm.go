package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type ServiceStoreV2LookupKind string
type ServiceStoreV2LargeValuePolicy string
type ServiceBlockSTMOperationKind string

const (
	ServiceStoreV2LookupPrimaryRead	ServiceStoreV2LookupKind	= "primary_read"
	ServiceStoreV2LookupPrefixQuery	ServiceStoreV2LookupKind	= "prefix_query"
	ServiceStoreV2LookupSingleton	ServiceStoreV2LookupKind	= "singleton"

	ServiceStoreV2LargeValueInlineIfSmall		ServiceStoreV2LargeValuePolicy	= "inline_if_small"
	ServiceStoreV2LargeValueHashCommittedOnNeed	ServiceStoreV2LargeValuePolicy	= "hash_committed_store_on_chain_when_needed"

	ServiceStoreV2DescriptorPrefix	= ServiceStorePrefix + "descriptor"
	ServiceStoreV2AnchorPrefix	= ServiceStorePrefix + "anchor"
	ServiceStoreV2InterfacePrefix	= ServiceStorePrefix + "interface"
	ServiceStoreV2CallPrefix	= ServiceStorePrefix + "call"
	ServiceStoreV2ReceiptPrefix	= ServiceStorePrefix + "receipt"
	ServiceStoreV2ProviderPrefix	= ServiceStorePrefix + "provider"
	ServiceStoreV2PaymentPrefix	= ServiceStorePrefix + "payment"
	ServiceStoreV2IndexPrefix	= ServiceStorePrefix + "index"
	ServiceStoreV2ParamsKey		= ServiceStorePrefix + "params"

	ServiceStoreV2OwnerIndexPrefix		= ServiceStoreV2IndexPrefix + "/owner"
	ServiceStoreV2IdentityIndexPrefix	= ServiceStoreV2IndexPrefix + "/identity"
	ServiceStoreV2ProviderIndexPrefix	= ServiceStoreV2IndexPrefix + "/provider"
	ServiceStoreV2MethodIndexPrefix		= ServiceStoreV2IndexPrefix + "/method"
	ServiceStoreV2ReceiptHeightPrefix	= ServiceStoreV2IndexPrefix + "/receipt_height"

	ServiceBlockSTMRegisterService		ServiceBlockSTMOperationKind	= "register_service"
	ServiceBlockSTMUpdateService		ServiceBlockSTMOperationKind	= "update_service"
	ServiceBlockSTMAnchorReceipt		ServiceBlockSTMOperationKind	= "anchor_receipt"
	ServiceBlockSTMUpdateProvider		ServiceBlockSTMOperationKind	= "update_provider"
	ServiceBlockSTMExecuteOnChainCall	ServiceBlockSTMOperationKind	= "execute_on_chain_call"
	ServiceBlockSTMSettlePayment		ServiceBlockSTMOperationKind	= "settle_payment"
	ServiceBlockSTMSlashProvider		ServiceBlockSTMOperationKind	= "slash_provider"
	ServiceBlockSTMDisputePayment		ServiceBlockSTMOperationKind	= "dispute_payment"
	ServiceBlockSTMUpdateInterface		ServiceBlockSTMOperationKind	= "update_interface"
	ServiceBlockSTMUpdateDescriptorIface	ServiceBlockSTMOperationKind	= "update_descriptor_interface"
)

type ServiceStoreV2Entry struct {
	Name			string
	Prefix			string
	Keeper			ServiceKeeperBoundaryName
	LookupKind		ServiceStoreV2LookupKind
	PartitionKey		string
	PrimaryReadBound	uint32
	PrefixQueryable		bool
	HeightIndexed		bool
	LargeValuePolicy	ServiceStoreV2LargeValuePolicy
	EntryHash		string
}

type ServiceStoreV2Layout struct {
	Entries		[]ServiceStoreV2Entry
	LayoutHash	string
}

type ServiceBlockSTMPartitionRule struct {
	StateFamily	string
	PartitionBy	[]string
	PartitionKey	string
	RuleHash	string
}

type ServiceBlockSTMOperation struct {
	Kind				ServiceBlockSTMOperationKind
	SubjectKey			string
	ReadPartitions			[]string
	WritePartitions			[]string
	ExpectedDescriptorVersion	uint64
	OperationHash			string
}

type ServiceBlockSTMStrategy struct {
	PartitionRules	[]ServiceBlockSTMPartitionRule
	StrategyHash	string
}

func DefaultServiceStoreV2Layout() (ServiceStoreV2Layout, error) {
	entries := []ServiceStoreV2Entry{
		newServiceStoreV2Entry("descriptor_by_service_id", ServiceStoreV2DescriptorPrefix, ServicesKeeperBoundary, ServiceStoreV2LookupPrimaryRead, "service_id", 1, false, false, ServiceStoreV2LargeValueInlineIfSmall),
		newServiceStoreV2Entry("anchor_by_service_id", ServiceStoreV2AnchorPrefix, ServicesKeeperBoundary, ServiceStoreV2LookupPrimaryRead, "service_id", 1, false, false, ServiceStoreV2LargeValueInlineIfSmall),
		newServiceStoreV2Entry("interface_by_interface_hash", ServiceStoreV2InterfacePrefix, InterfaceKeeperBoundary, ServiceStoreV2LookupPrimaryRead, "interface_hash", 1, false, false, ServiceStoreV2LargeValueHashCommittedOnNeed),
		newServiceStoreV2Entry("call_by_call_id", ServiceStoreV2CallPrefix, CallKeeperBoundary, ServiceStoreV2LookupPrimaryRead, "call_id", 1, false, false, ServiceStoreV2LargeValueInlineIfSmall),
		newServiceStoreV2Entry("receipt_by_call_id", ServiceStoreV2ReceiptPrefix, ReceiptKeeperBoundary, ServiceStoreV2LookupPrimaryRead, "call_id", 1, false, true, ServiceStoreV2LargeValueInlineIfSmall),
		newServiceStoreV2Entry("provider_by_provider_id", ServiceStoreV2ProviderPrefix, ProviderKeeperBoundary, ServiceStoreV2LookupPrimaryRead, "provider_id", 1, false, false, ServiceStoreV2LargeValueInlineIfSmall),
		newServiceStoreV2Entry("payment_by_escrow_or_stream_id", ServiceStoreV2PaymentPrefix, PaymentKeeperBoundary, ServiceStoreV2LookupPrimaryRead, "escrow_id_or_stream_id", 1, false, false, ServiceStoreV2LargeValueInlineIfSmall),
		newServiceStoreV2Entry("index_root", ServiceStoreV2IndexPrefix, ServicesKeeperBoundary, ServiceStoreV2LookupPrefixQuery, "index_name", 0, true, false, ServiceStoreV2LargeValueInlineIfSmall),
		newServiceStoreV2Entry("owner_service_index", ServiceStoreV2OwnerIndexPrefix, ServicesKeeperBoundary, ServiceStoreV2LookupPrefixQuery, "owner/service_id", 0, true, false, ServiceStoreV2LargeValueInlineIfSmall),
		newServiceStoreV2Entry("identity_service_index", ServiceStoreV2IdentityIndexPrefix, ServicesKeeperBoundary, ServiceStoreV2LookupPrefixQuery, "identity_name/service_id", 0, true, false, ServiceStoreV2LargeValueInlineIfSmall),
		newServiceStoreV2Entry("provider_service_index", ServiceStoreV2ProviderIndexPrefix, ProviderKeeperBoundary, ServiceStoreV2LookupPrefixQuery, "service_id/provider_id", 0, true, false, ServiceStoreV2LargeValueInlineIfSmall),
		newServiceStoreV2Entry("method_service_index", ServiceStoreV2MethodIndexPrefix, InterfaceKeeperBoundary, ServiceStoreV2LookupPrefixQuery, "interface_hash/method_id", 0, true, false, ServiceStoreV2LargeValueInlineIfSmall),
		newServiceStoreV2Entry("receipt_height_pruning_index", ServiceStoreV2ReceiptHeightPrefix, ReceiptKeeperBoundary, ServiceStoreV2LookupPrefixQuery, "height/call_id", 0, true, true, ServiceStoreV2LargeValueInlineIfSmall),
		newServiceStoreV2Entry("params_singleton", ServiceStoreV2ParamsKey, ServicesKeeperBoundary, ServiceStoreV2LookupSingleton, "params", 1, false, false, ServiceStoreV2LargeValueInlineIfSmall),
	}
	return NewServiceStoreV2Layout(entries)
}

func NewServiceStoreV2Layout(entries []ServiceStoreV2Entry) (ServiceStoreV2Layout, error) {
	layout := ServiceStoreV2Layout{Entries: normalizeServiceStoreV2Entries(entries)}
	if err := layout.ValidateFormat(); err != nil {
		return ServiceStoreV2Layout{}, err
	}
	layout.LayoutHash = ComputeServiceStoreV2LayoutHash(layout)
	return layout, layout.Validate()
}

func (layout ServiceStoreV2Layout) ValidateFormat() error {
	if len(layout.Entries) == 0 {
		return errors.New("services Store v2 layout entries are required")
	}
	seen := map[string]struct{}{}
	for _, entry := range layout.Entries {
		if err := entry.Validate(); err != nil {
			return err
		}
		if _, found := seen[entry.Prefix]; found {
			return fmt.Errorf("services Store v2 duplicate prefix %s", entry.Prefix)
		}
		seen[entry.Prefix] = struct{}{}
	}
	return nil
}

func (layout ServiceStoreV2Layout) Validate() error {
	if err := layout.ValidateFormat(); err != nil {
		return err
	}
	if err := validateServiceStoreV2RequiredPrefixes(layout.Entries); err != nil {
		return err
	}
	if err := validateServiceStoreV2PerformanceRules(layout.Entries); err != nil {
		return err
	}
	if err := validateInterfaceToken("services Store v2 layout hash", layout.LayoutHash); err != nil {
		return err
	}
	if expected := ComputeServiceStoreV2LayoutHash(layout); layout.LayoutHash != expected {
		return fmt.Errorf("services Store v2 layout hash mismatch: expected %s", expected)
	}
	return nil
}

func (entry ServiceStoreV2Entry) Validate() error {
	if err := validateInterfaceToken("services Store v2 entry name", entry.Name); err != nil {
		return err
	}
	if entry.Prefix != ServiceStoreV2ParamsKey {
		if !IsServiceStoreKey(entry.Prefix + "/_") {
			return fmt.Errorf("services Store v2 prefix %s must use services prefix", entry.Prefix)
		}
	} else if !IsServiceStoreKey(entry.Prefix) {
		return fmt.Errorf("services Store v2 key %s must use services prefix", entry.Prefix)
	}
	if !IsServiceKeeperBoundaryName(entry.Keeper) {
		return fmt.Errorf("services Store v2 unknown keeper %q", entry.Keeper)
	}
	if !IsServiceStoreV2LookupKind(entry.LookupKind) {
		return fmt.Errorf("services Store v2 unknown lookup kind %q", entry.LookupKind)
	}
	if err := validateInterfaceToken("services Store v2 partition key", entry.PartitionKey); err != nil {
		return err
	}
	if entry.LookupKind == ServiceStoreV2LookupPrimaryRead && entry.PrimaryReadBound != 1 {
		return fmt.Errorf("services Store v2 primary lookup %s must be one primary read", entry.Name)
	}
	if entry.LookupKind == ServiceStoreV2LookupPrefixQuery && !entry.PrefixQueryable {
		return fmt.Errorf("services Store v2 index %s must be prefix-queryable", entry.Name)
	}
	if !IsServiceStoreV2LargeValuePolicy(entry.LargeValuePolicy) {
		return fmt.Errorf("services Store v2 unknown large value policy %q", entry.LargeValuePolicy)
	}
	if err := validateInterfaceToken("services Store v2 entry hash", entry.EntryHash); err != nil {
		return err
	}
	if expected := ComputeServiceStoreV2EntryHash(entry); entry.EntryHash != expected {
		return fmt.Errorf("services Store v2 entry hash mismatch: expected %s", expected)
	}
	return nil
}

func DefaultServiceBlockSTMStrategy() (ServiceBlockSTMStrategy, error) {
	rules := []ServiceBlockSTMPartitionRule{
		newServiceBlockSTMPartitionRule("descriptor", []string{"service_id"}, "service_id"),
		newServiceBlockSTMPartitionRule("receipt", []string{"call_id"}, "call_id"),
		newServiceBlockSTMPartitionRule("provider", []string{"provider_id"}, "provider_id"),
		newServiceBlockSTMPartitionRule("payment", []string{"escrow_id", "stream_id"}, "escrow_id_or_stream_id"),
		newServiceBlockSTMPartitionRule("service_local_state", []string{"service_id", "state_key"}, "service_id/state_key"),
		newServiceBlockSTMPartitionRule("interface", []string{"interface_hash"}, "interface_hash"),
	}
	return NewServiceBlockSTMStrategy(rules)
}

func NewServiceBlockSTMStrategy(rules []ServiceBlockSTMPartitionRule) (ServiceBlockSTMStrategy, error) {
	strategy := ServiceBlockSTMStrategy{PartitionRules: normalizeServiceBlockSTMPartitionRules(rules)}
	if err := strategy.ValidateFormat(); err != nil {
		return ServiceBlockSTMStrategy{}, err
	}
	strategy.StrategyHash = ComputeServiceBlockSTMStrategyHash(strategy)
	return strategy, strategy.Validate()
}

func (strategy ServiceBlockSTMStrategy) ValidateFormat() error {
	if len(strategy.PartitionRules) == 0 {
		return errors.New("services BlockSTM partition rules are required")
	}
	seen := map[string]struct{}{}
	for _, rule := range strategy.PartitionRules {
		if err := rule.Validate(); err != nil {
			return err
		}
		if _, found := seen[rule.StateFamily]; found {
			return fmt.Errorf("services BlockSTM duplicate state family %s", rule.StateFamily)
		}
		seen[rule.StateFamily] = struct{}{}
	}
	return nil
}

func (strategy ServiceBlockSTMStrategy) Validate() error {
	if err := strategy.ValidateFormat(); err != nil {
		return err
	}
	required := []string{"descriptor", "receipt", "provider", "payment", "service_local_state", "interface"}
	for _, family := range required {
		if _, found := strategy.ruleByFamily(family); !found {
			return fmt.Errorf("services BlockSTM missing partition family %s", family)
		}
	}
	if err := validateInterfaceToken("services BlockSTM strategy hash", strategy.StrategyHash); err != nil {
		return err
	}
	if expected := ComputeServiceBlockSTMStrategyHash(strategy); strategy.StrategyHash != expected {
		return fmt.Errorf("services BlockSTM strategy hash mismatch: expected %s", expected)
	}
	return nil
}

func (rule ServiceBlockSTMPartitionRule) Validate() error {
	if err := validateInterfaceToken("services BlockSTM state family", rule.StateFamily); err != nil {
		return err
	}
	if len(rule.PartitionBy) == 0 {
		return errors.New("services BlockSTM partition fields are required")
	}
	if err := validateSortedTokens("services BlockSTM partition field", rule.PartitionBy); err != nil {
		return err
	}
	if err := validateInterfaceToken("services BlockSTM partition key", rule.PartitionKey); err != nil {
		return err
	}
	if err := validateInterfaceToken("services BlockSTM rule hash", rule.RuleHash); err != nil {
		return err
	}
	if expected := ComputeServiceBlockSTMPartitionRuleHash(rule); rule.RuleHash != expected {
		return fmt.Errorf("services BlockSTM rule hash mismatch: expected %s", expected)
	}
	return nil
}

func NewServiceBlockSTMOperation(kind ServiceBlockSTMOperationKind, ids map[string]string, expectedDescriptorVersion uint64) (ServiceBlockSTMOperation, error) {
	cleanIDs := cleanServiceBlockSTMIDs(ids)
	op := ServiceBlockSTMOperation{Kind: kind, ExpectedDescriptorVersion: expectedDescriptorVersion}
	switch kind {
	case ServiceBlockSTMRegisterService:
		op.SubjectKey = cleanIDs["service_id"]
		op.WritePartitions = []string{serviceBlockSTMPartition("descriptor", cleanIDs["service_id"])}
	case ServiceBlockSTMUpdateService:
		op.SubjectKey = cleanIDs["service_id"]
		op.ReadPartitions = []string{serviceBlockSTMPartition("descriptor", cleanIDs["service_id"])}
		op.WritePartitions = []string{serviceBlockSTMPartition("descriptor", cleanIDs["service_id"])}
	case ServiceBlockSTMAnchorReceipt:
		op.SubjectKey = cleanIDs["call_id"]
		op.WritePartitions = []string{serviceBlockSTMPartition("receipt", cleanIDs["call_id"])}
	case ServiceBlockSTMUpdateProvider:
		op.SubjectKey = cleanIDs["provider_id"]
		op.WritePartitions = []string{serviceBlockSTMPartition("provider", cleanIDs["provider_id"])}
	case ServiceBlockSTMExecuteOnChainCall:
		op.SubjectKey = cleanIDs["call_id"]
		op.ReadPartitions = []string{serviceBlockSTMPartition("descriptor", cleanIDs["service_id"])}
		op.WritePartitions = []string{
			serviceBlockSTMPartition("receipt", cleanIDs["call_id"]),
			serviceBlockSTMPartition("service_local_state", cleanIDs["service_id"]+"/"+cleanIDs["state_key"]),
		}
	case ServiceBlockSTMSettlePayment:
		op.SubjectKey = firstNonEmpty(cleanIDs["escrow_id"], cleanIDs["stream_id"])
		op.WritePartitions = []string{serviceBlockSTMPartition("payment", op.SubjectKey)}
	case ServiceBlockSTMSlashProvider:
		op.SubjectKey = cleanIDs["provider_id"]
		op.WritePartitions = []string{serviceBlockSTMPartition("provider", cleanIDs["provider_id"])}
	case ServiceBlockSTMDisputePayment:
		op.SubjectKey = cleanIDs["call_id"]
		op.ReadPartitions = []string{serviceBlockSTMPartition("receipt", cleanIDs["call_id"])}
		op.WritePartitions = []string{serviceBlockSTMPartition("payment", firstNonEmpty(cleanIDs["escrow_id"], cleanIDs["stream_id"]))}
	case ServiceBlockSTMUpdateInterface:
		op.SubjectKey = cleanIDs["interface_hash"]
		op.WritePartitions = []string{serviceBlockSTMPartition("interface", cleanIDs["interface_hash"])}
	case ServiceBlockSTMUpdateDescriptorIface:
		op.SubjectKey = cleanIDs["service_id"]
		op.ReadPartitions = []string{serviceBlockSTMPartition("interface", cleanIDs["interface_hash"])}
		op.WritePartitions = []string{serviceBlockSTMPartition("descriptor", cleanIDs["service_id"])}
	default:
		return ServiceBlockSTMOperation{}, fmt.Errorf("services BlockSTM unknown operation %q", kind)
	}
	op.ReadPartitions = sortedStrings(op.ReadPartitions)
	op.WritePartitions = sortedStrings(op.WritePartitions)
	op.OperationHash = ComputeServiceBlockSTMOperationHash(op)
	return op, op.Validate()
}

func (op ServiceBlockSTMOperation) Validate() error {
	if !IsServiceBlockSTMOperationKind(op.Kind) {
		return fmt.Errorf("services BlockSTM unknown operation %q", op.Kind)
	}
	if err := validateInterfaceToken("services BlockSTM subject key", op.SubjectKey); err != nil {
		return err
	}
	if len(op.ReadPartitions)+len(op.WritePartitions) == 0 {
		return errors.New("services BlockSTM operation must declare read or write partitions")
	}
	if err := validateSortedTokens("services BlockSTM read partition", op.ReadPartitions); err != nil {
		return err
	}
	if err := validateSortedTokens("services BlockSTM write partition", op.WritePartitions); err != nil {
		return err
	}
	if op.Kind == ServiceBlockSTMUpdateService && op.ExpectedDescriptorVersion == 0 {
		return errors.New("services BlockSTM descriptor updates require expected version")
	}
	if err := validateInterfaceToken("services BlockSTM operation hash", op.OperationHash); err != nil {
		return err
	}
	if expected := ComputeServiceBlockSTMOperationHash(op); op.OperationHash != expected {
		return fmt.Errorf("services BlockSTM operation hash mismatch: expected %s", expected)
	}
	return nil
}

func ServiceBlockSTMOperationsConflict(left, right ServiceBlockSTMOperation) bool {
	leftWrites := map[string]struct{}{}
	for _, partition := range left.WritePartitions {
		leftWrites[partition] = struct{}{}
	}
	rightWrites := map[string]struct{}{}
	for _, partition := range right.WritePartitions {
		rightWrites[partition] = struct{}{}
		if _, found := leftWrites[partition]; found {
			return true
		}
	}
	for _, partition := range left.ReadPartitions {
		if _, found := rightWrites[partition]; found {
			return true
		}
	}
	for _, partition := range right.ReadPartitions {
		if _, found := leftWrites[partition]; found {
			return true
		}
	}
	return false
}

func IsServiceStoreV2LookupKind(kind ServiceStoreV2LookupKind) bool {
	switch kind {
	case ServiceStoreV2LookupPrimaryRead, ServiceStoreV2LookupPrefixQuery, ServiceStoreV2LookupSingleton:
		return true
	default:
		return false
	}
}

func IsServiceStoreV2LargeValuePolicy(policy ServiceStoreV2LargeValuePolicy) bool {
	switch policy {
	case ServiceStoreV2LargeValueInlineIfSmall, ServiceStoreV2LargeValueHashCommittedOnNeed:
		return true
	default:
		return false
	}
}

func IsServiceBlockSTMOperationKind(kind ServiceBlockSTMOperationKind) bool {
	switch kind {
	case ServiceBlockSTMRegisterService, ServiceBlockSTMUpdateService, ServiceBlockSTMAnchorReceipt, ServiceBlockSTMUpdateProvider, ServiceBlockSTMExecuteOnChainCall, ServiceBlockSTMSettlePayment, ServiceBlockSTMSlashProvider, ServiceBlockSTMDisputePayment, ServiceBlockSTMUpdateInterface, ServiceBlockSTMUpdateDescriptorIface:
		return true
	default:
		return false
	}
}

func ComputeServiceStoreV2EntryHash(entry ServiceStoreV2Entry) string {
	return servicesHashParts(
		"aetra-services-store-v2-entry-v1",
		entry.Name,
		entry.Prefix,
		string(entry.Keeper),
		string(entry.LookupKind),
		entry.PartitionKey,
		fmt.Sprint(entry.PrimaryReadBound),
		fmt.Sprint(entry.PrefixQueryable),
		fmt.Sprint(entry.HeightIndexed),
		string(entry.LargeValuePolicy),
	)
}

func ComputeServiceStoreV2LayoutHash(layout ServiceStoreV2Layout) string {
	entries := normalizeServiceStoreV2Entries(layout.Entries)
	parts := []string{"aetra-services-store-v2-layout-v1", fmt.Sprint(len(entries))}
	for _, entry := range entries {
		parts = append(parts, entry.EntryHash)
	}
	return servicesHashParts(parts...)
}

func ComputeServiceBlockSTMPartitionRuleHash(rule ServiceBlockSTMPartitionRule) string {
	return servicesHashParts("aetra-services-blockstm-partition-rule-v1", rule.StateFamily, strings.Join(rule.PartitionBy, ","), rule.PartitionKey)
}

func ComputeServiceBlockSTMStrategyHash(strategy ServiceBlockSTMStrategy) string {
	rules := normalizeServiceBlockSTMPartitionRules(strategy.PartitionRules)
	parts := []string{"aetra-services-blockstm-strategy-v1", fmt.Sprint(len(rules))}
	for _, rule := range rules {
		parts = append(parts, rule.RuleHash)
	}
	return servicesHashParts(parts...)
}

func ComputeServiceBlockSTMOperationHash(op ServiceBlockSTMOperation) string {
	return servicesHashParts(
		"aetra-services-blockstm-operation-v1",
		string(op.Kind),
		op.SubjectKey,
		strings.Join(op.ReadPartitions, ","),
		strings.Join(op.WritePartitions, ","),
		fmt.Sprint(op.ExpectedDescriptorVersion),
	)
}

func newServiceStoreV2Entry(name, prefix string, keeper ServiceKeeperBoundaryName, lookup ServiceStoreV2LookupKind, partitionKey string, primaryReads uint32, prefixQueryable, heightIndexed bool, largeValuePolicy ServiceStoreV2LargeValuePolicy) ServiceStoreV2Entry {
	entry := ServiceStoreV2Entry{
		Name:			name,
		Prefix:			prefix,
		Keeper:			keeper,
		LookupKind:		lookup,
		PartitionKey:		partitionKey,
		PrimaryReadBound:	primaryReads,
		PrefixQueryable:	prefixQueryable,
		HeightIndexed:		heightIndexed,
		LargeValuePolicy:	largeValuePolicy,
	}
	entry.EntryHash = ComputeServiceStoreV2EntryHash(entry)
	return entry
}

func newServiceBlockSTMPartitionRule(family string, partitionBy []string, partitionKey string) ServiceBlockSTMPartitionRule {
	rule := ServiceBlockSTMPartitionRule{
		StateFamily:	family,
		PartitionBy:	sortedStrings(partitionBy),
		PartitionKey:	partitionKey,
	}
	rule.RuleHash = ComputeServiceBlockSTMPartitionRuleHash(rule)
	return rule
}

func normalizeServiceStoreV2Entries(entries []ServiceStoreV2Entry) []ServiceStoreV2Entry {
	out := append([]ServiceStoreV2Entry(nil), entries...)
	for i := range out {
		out[i].Name = strings.TrimSpace(out[i].Name)
		out[i].Prefix = strings.Trim(strings.TrimSpace(out[i].Prefix), "/")
		out[i].PartitionKey = strings.TrimSpace(out[i].PartitionKey)
		out[i].EntryHash = strings.TrimSpace(out[i].EntryHash)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Prefix < out[j].Prefix })
	return out
}

func normalizeServiceBlockSTMPartitionRules(rules []ServiceBlockSTMPartitionRule) []ServiceBlockSTMPartitionRule {
	out := append([]ServiceBlockSTMPartitionRule(nil), rules...)
	for i := range out {
		out[i].StateFamily = strings.TrimSpace(out[i].StateFamily)
		out[i].PartitionBy = sortedStrings(out[i].PartitionBy)
		out[i].PartitionKey = strings.TrimSpace(out[i].PartitionKey)
		out[i].RuleHash = strings.TrimSpace(out[i].RuleHash)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].StateFamily < out[j].StateFamily })
	return out
}

func validateServiceStoreV2RequiredPrefixes(entries []ServiceStoreV2Entry) error {
	required := []string{
		ServiceStoreV2DescriptorPrefix,
		ServiceStoreV2AnchorPrefix,
		ServiceStoreV2InterfacePrefix,
		ServiceStoreV2CallPrefix,
		ServiceStoreV2ReceiptPrefix,
		ServiceStoreV2ProviderPrefix,
		ServiceStoreV2PaymentPrefix,
		ServiceStoreV2IndexPrefix,
		ServiceStoreV2ParamsKey,
	}
	for _, prefix := range required {
		if _, found := serviceStoreV2EntryByPrefix(entries, prefix); !found {
			return fmt.Errorf("services Store v2 missing required prefix %s", prefix)
		}
	}
	return nil
}

func validateServiceStoreV2PerformanceRules(entries []ServiceStoreV2Entry) error {
	descriptor, found := serviceStoreV2EntryByPrefix(entries, ServiceStoreV2DescriptorPrefix)
	if !found || descriptor.LookupKind != ServiceStoreV2LookupPrimaryRead || descriptor.PartitionKey != "service_id" || descriptor.PrimaryReadBound != 1 {
		return errors.New("services Store v2 descriptor lookup by service_id must be one primary read")
	}
	iface, found := serviceStoreV2EntryByPrefix(entries, ServiceStoreV2InterfacePrefix)
	if !found || iface.LookupKind != ServiceStoreV2LookupPrimaryRead || iface.PartitionKey != "interface_hash" || iface.PrimaryReadBound != 1 {
		return errors.New("services Store v2 interface lookup by interface_hash must be one primary read")
	}
	for _, prefix := range []string{ServiceStoreV2OwnerIndexPrefix, ServiceStoreV2IdentityIndexPrefix, ServiceStoreV2ProviderIndexPrefix, ServiceStoreV2MethodIndexPrefix} {
		entry, found := serviceStoreV2EntryByPrefix(entries, prefix)
		if !found || !entry.PrefixQueryable {
			return fmt.Errorf("services Store v2 index %s must be prefix-queryable", prefix)
		}
	}
	receipt, found := serviceStoreV2EntryByPrefix(entries, ServiceStoreV2ReceiptPrefix)
	if !found || !receipt.HeightIndexed {
		return errors.New("services Store v2 receipts must be height-indexed for pruning")
	}
	receiptHeight, found := serviceStoreV2EntryByPrefix(entries, ServiceStoreV2ReceiptHeightPrefix)
	if !found || !receiptHeight.HeightIndexed || !receiptHeight.PrefixQueryable {
		return errors.New("services Store v2 receipt height index is required for pruning")
	}
	if iface.LargeValuePolicy != ServiceStoreV2LargeValueHashCommittedOnNeed {
		return errors.New("services Store v2 large schemas must be committed by hash and stored only when needed on-chain")
	}
	return nil
}

func serviceStoreV2EntryByPrefix(entries []ServiceStoreV2Entry, prefix string) (ServiceStoreV2Entry, bool) {
	for _, entry := range entries {
		if entry.Prefix == prefix {
			return entry, true
		}
	}
	return ServiceStoreV2Entry{}, false
}

func (strategy ServiceBlockSTMStrategy) ruleByFamily(family string) (ServiceBlockSTMPartitionRule, bool) {
	for _, rule := range strategy.PartitionRules {
		if rule.StateFamily == family {
			return rule, true
		}
	}
	return ServiceBlockSTMPartitionRule{}, false
}

func serviceBlockSTMPartition(family, key string) string {
	return family + "/" + strings.Trim(strings.TrimSpace(key), "/")
}

func cleanServiceBlockSTMIDs(ids map[string]string) map[string]string {
	out := map[string]string{}
	for key, value := range ids {
		out[strings.TrimSpace(key)] = strings.Trim(strings.TrimSpace(value), "/")
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
