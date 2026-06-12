package types

import (
	"errors"
	"fmt"
	"sort"
)

const ImplementationBacklogSpecVersion = uint64(1)

type BacklogPriority string
type BacklogItemID string

const (
	BacklogPriorityHigh	BacklogPriority	= "high"
	BacklogPriorityMedium	BacklogPriority	= "medium"
	BacklogPriorityLower	BacklogPriority	= "lower"

	BacklogItemZoneDescriptorCommitmentLayout	BacklogItemID	= "define-zone-descriptor-zone-commitment-shard-layout"
	BacklogItemAetraCoreSkeleton			BacklogItemID	= "aetracore-skeleton"
	BacklogItemGlobalRootHierarchy			BacklogItemID	= "global-root-hierarchy"
	BacklogItemMsgbusEncoding			BacklogItemID	= "msgbus-message-encoding"
	BacklogItemLocalMessageStores			BacklogItemID	= "local-inbox-outbox-receipt-stores"
	BacklogItemStoreV2PrefixPlan			BacklogItemID	= "store-v2-key-prefix-plan"
	BacklogItemBlockSTMConflictTests		BacklogItemID	= "blockstm-zone-batch-conflict-tests"
	BacklogItemRoutingTableFormat			BacklogItemID	= "deterministic-routing-table-format"
	BacklogItemProofRegistrySchema			BacklogItemID	= "proof-registry-schema"

	BacklogItemExtractFinancialZone		BacklogItemID	= "extract-financial-zone"
	BacklogItemActivateIdentityZone		BacklogItemID	= "activate-identity-zone"
	BacklogItemPerZoneMempoolLanes		BacklogItemID	= "per-zone-mempool-lanes"
	BacklogItemPerShardFeeAccumulators	BacklogItemID	= "per-shard-fee-accumulators"
	BacklogItemShardSplitMergeScheduler	BacklogItemID	= "shard-split-merge-scheduler"
	BacklogItemAVMBytecodeGasTable		BacklogItemID	= "avm-bytecode-gas-table"
	BacklogItemPaymentSettlementState	BacklogItemID	= "payment-settlement-state"
	BacklogItemCrossZoneIdentityLookup	BacklogItemID	= "cross-zone-identity-lookup"

	BacklogItemDynamicRouteCapacityScoring	BacklogItemID	= "dynamic-route-capacity-scoring"
	BacklogItemVirtualPaymentChannels	BacklogItemID	= "virtual-payment-channels"
	BacklogItemAdvancedABIIntrospection	BacklogItemID	= "advanced-abi-introspection"
	BacklogItemVMNativeResolverContracts	BacklogItemID	= "vm-native-resolver-contracts"
	BacklogItemValidatorServiceMetadata	BacklogItemID	= "validator-operated-service-metadata"
	BacklogItemZoneStateRentPolicies	BacklogItemID	= "zone-specific-state-rent-policies"
)

type BacklogItem struct {
	Priority	BacklogPriority
	ItemID		BacklogItemID
	Task		string
	Target		string
	Acceptance	[]string
	DescriptorHash	string
}

type ImplementationBacklogSpec struct {
	Version	uint64
	Items	[]BacklogItem
	Root	string
}

func DefaultImplementationBacklogSpec() (ImplementationBacklogSpec, error) {
	return BuildImplementationBacklogSpec(ImplementationBacklogItems())
}

func BuildImplementationBacklogSpec(items []BacklogItem) (ImplementationBacklogSpec, error) {
	spec := ImplementationBacklogSpec{
		Version:	ImplementationBacklogSpecVersion,
		Items:		normalizeBacklogItems(items),
	}
	if err := spec.ValidateFormat(); err != nil {
		return ImplementationBacklogSpec{}, err
	}
	spec.Root = ComputeImplementationBacklogRoot(spec.Items)
	return spec, spec.Validate()
}

func ImplementationBacklogItems() []BacklogItem {
	items := make([]BacklogItem, 0, 23)
	items = append(items, HighPriorityBacklogItems()...)
	items = append(items, MediumPriorityBacklogItems()...)
	items = append(items, LowerPriorityBacklogItems()...)
	return items
}

func HighPriorityBacklogItems() []BacklogItem {
	return []BacklogItem{
		backlogItem(BacklogPriorityHigh, BacklogItemZoneDescriptorCommitmentLayout, "Define ZoneDescriptor, ZoneCommitment, and ShardLayout.", "x/aetracore types", []string{"zone descriptor validation", "zone commitment hash", "shard layout hash", "canonical export"}),
		backlogItem(BacklogPriorityHigh, BacklogItemAetraCoreSkeleton, "Implement x/aetracore skeleton.", "x/aetracore module", []string{"module name", "keeper shell", "params", "genesis import export"}),
		backlogItem(BacklogPriorityHigh, BacklogItemGlobalRootHierarchy, "Implement global root hierarchy.", "root aggregation", []string{"aether core root", "global zone root", "global message root", "proof root registry"}),
		backlogItem(BacklogPriorityHigh, BacklogItemMsgbusEncoding, "Implement x/msgbus message encoding.", "x/msgbus codec", []string{"canonical AetherMessage envelope", "message ID derivation", "receipt hash"}),
		backlogItem(BacklogPriorityHigh, BacklogItemLocalMessageStores, "Implement local inbox, outbox, and receipt stores.", "message stores", []string{"inbox prefix", "outbox prefix", "receipt prefix", "deterministic ordering"}),
		backlogItem(BacklogPriorityHigh, BacklogItemStoreV2PrefixPlan, "Add Store v2 key prefix plan.", "state prefix plan", []string{"core prefixes", "zone prefixes", "shard prefixes", "proof scope mapping"}),
		backlogItem(BacklogPriorityHigh, BacklogItemBlockSTMConflictTests, "Add BlockSTM conflict tests for zone batches.", "BlockSTM tests", []string{"disjoint zone batches", "same object conflicts", "cross-zone message conversion"}),
		backlogItem(BacklogPriorityHigh, BacklogItemRoutingTableFormat, "Add deterministic routing table format.", "routing table state", []string{"routing epoch", "adjacency", "capacity metrics", "path commitment"}),
		backlogItem(BacklogPriorityHigh, BacklogItemProofRegistrySchema, "Add proof registry schema.", "x/proofregistry types", []string{"root metadata", "history window", "proof query envelope", "failure codes"}),
	}
}

func MediumPriorityBacklogItems() []BacklogItem {
	return []BacklogItem{
		backlogItem(BacklogPriorityMedium, BacklogItemExtractFinancialZone, "Extract Financial Zone.", "Financial Zone adapter", []string{"bank routing", "fees routing", "contract-assets routing", "AVM AMM routing"}),
		backlogItem(BacklogPriorityMedium, BacklogItemActivateIdentityZone, "Activate Identity Zone.", "Identity Zone adapter", []string{".aet registry", "resolver roots", "reverse lookup", "delegation state"}),
		backlogItem(BacklogPriorityMedium, BacklogItemPerZoneMempoolLanes, "Implement per-zone mempool lanes.", "zonemempool", []string{"zone lane key", "shard sublane key", "message priority class"}),
		backlogItem(BacklogPriorityMedium, BacklogItemPerShardFeeAccumulators, "Implement per-shard fee accumulators.", "zonefees", []string{"shard fee bucket", "end-block aggregation", "zone fee root"}),
		backlogItem(BacklogPriorityMedium, BacklogItemShardSplitMergeScheduler, "Implement shard split and merge scheduler.", "x/shards scheduler", []string{"committed metrics", "future layout epoch", "migration task"}),
		backlogItem(BacklogPriorityMedium, BacklogItemAVMBytecodeGasTable, "Implement AVM bytecode and gas table.", "x/avm", []string{"bytecode codec", "opcode gas", "memory gas", "storage gas"}),
		backlogItem(BacklogPriorityMedium, BacklogItemPaymentSettlementState, "Implement payment settlement state.", "x/payments", []string{"channel state", "condition state", "settlement proof", "payment receipt root"}),
		backlogItem(BacklogPriorityMedium, BacklogItemCrossZoneIdentityLookup, "Implement cross-zone identity lookup.", "identity messages", []string{"lookup request", "resolution result", "proof optional", "expiry handling"}),
	}
}

func LowerPriorityBacklogItems() []BacklogItem {
	return []BacklogItem{
		backlogItem(BacklogPriorityLower, BacklogItemDynamicRouteCapacityScoring, "Implement dynamic route capacity scoring.", "routing cost model", []string{"committed capacity", "congestion score", "deterministic tie-break"}),
		backlogItem(BacklogPriorityLower, BacklogItemVirtualPaymentChannels, "Implement virtual payment channels.", "x/payments virtual channels", []string{"underlying channel proof", "route commitment", "capacity proof"}),
		backlogItem(BacklogPriorityLower, BacklogItemAdvancedABIIntrospection, "Implement advanced ABI introspection.", "x/avm ABI registry", []string{"interface hash", "method descriptors", "event descriptors", "gas hints"}),
		backlogItem(BacklogPriorityLower, BacklogItemVMNativeResolverContracts, "Implement VM-native resolver contracts.", "identity resolver adapter", []string{"bounded resolver execution", "native ownership check", "resolver output proof"}),
		backlogItem(BacklogPriorityLower, BacklogItemValidatorServiceMetadata, "Implement validator-operated service metadata.", "service registry", []string{"validator service record", "availability hash", "proof anchored metadata"}),
		backlogItem(BacklogPriorityLower, BacklogItemZoneStateRentPolicies, "Implement zone-specific state rent policies.", "zone rent policy", []string{"zone rent params", "storage metering", "expiry or renewal rule"}),
	}
}

func (s ImplementationBacklogSpec) Normalize() ImplementationBacklogSpec {
	if s.Version == 0 {
		s.Version = ImplementationBacklogSpecVersion
	}
	s.Items = normalizeBacklogItems(s.Items)
	s.Root = normalizePerformanceHash(s.Root)
	return s
}

func (s ImplementationBacklogSpec) ValidateFormat() error {
	s = s.Normalize()
	if s.Version != ImplementationBacklogSpecVersion {
		return fmt.Errorf("aetracore implementation backlog spec version must be %d", ImplementationBacklogSpecVersion)
	}
	if len(s.Items) == 0 {
		return errors.New("aetracore implementation backlog requires items")
	}
	seen := make(map[BacklogItemID]struct{}, len(s.Items))
	var previousPriority BacklogPriority
	var previousItem BacklogItemID
	for i, item := range s.Items {
		if err := item.Validate(); err != nil {
			return err
		}
		if _, found := seen[item.ItemID]; found {
			return fmt.Errorf("duplicate aetracore implementation backlog item %s", item.ItemID)
		}
		seen[item.ItemID] = struct{}{}
		if i > 0 {
			if backlogPriorityRank(previousPriority) > backlogPriorityRank(item.Priority) {
				return errors.New("aetracore implementation backlog priorities must be sorted canonically")
			}
			if previousPriority == item.Priority && previousItem >= item.ItemID {
				return errors.New("aetracore implementation backlog items must be sorted canonically")
			}
		}
		previousPriority = item.Priority
		previousItem = item.ItemID
	}
	if s.Root != "" {
		if err := ValidateHash("aetracore implementation backlog root", s.Root); err != nil {
			return err
		}
	}
	return nil
}

func (s ImplementationBacklogSpec) Validate() error {
	s = s.Normalize()
	if err := s.ValidateFormat(); err != nil {
		return err
	}
	if s.Root == "" {
		return errors.New("aetracore implementation backlog root is required")
	}
	expected := ComputeImplementationBacklogRoot(s.Items)
	if s.Root != expected {
		return fmt.Errorf("aetracore implementation backlog root mismatch: expected %s", expected)
	}
	return nil
}

func (i BacklogItem) Normalize() BacklogItem {
	i.Task = compactPerformanceText(i.Task)
	i.Target = compactPerformanceText(i.Target)
	i.Acceptance = normalizeBacklogAcceptance(i.Acceptance)
	i.DescriptorHash = normalizePerformanceHash(i.DescriptorHash)
	return i
}

func (i BacklogItem) ValidateFormat() error {
	i = i.Normalize()
	if !IsBacklogPriority(i.Priority) {
		return fmt.Errorf("unknown aetracore implementation backlog priority %q", i.Priority)
	}
	if !IsBacklogItemID(i.Priority, i.ItemID) {
		return fmt.Errorf("unknown aetracore implementation backlog item %q for priority %s", i.ItemID, i.Priority)
	}
	if i.Task == "" || i.Target == "" {
		return errors.New("aetracore implementation backlog item requires task and target")
	}
	if len(i.Acceptance) == 0 {
		return errors.New("aetracore implementation backlog item requires acceptance criteria")
	}
	if i.DescriptorHash != "" {
		if err := ValidateHash("aetracore implementation backlog descriptor hash", i.DescriptorHash); err != nil {
			return err
		}
	}
	return nil
}

func (i BacklogItem) Validate() error {
	i = i.Normalize()
	if err := i.ValidateFormat(); err != nil {
		return err
	}
	if i.DescriptorHash == "" {
		return errors.New("aetracore implementation backlog descriptor hash is required")
	}
	expected := ComputeBacklogItemHash(i)
	if i.DescriptorHash != expected {
		return fmt.Errorf("aetracore implementation backlog descriptor hash mismatch: expected %s", expected)
	}
	return nil
}

func ValidateImplementationBacklogCoverage() error {
	spec, err := DefaultImplementationBacklogSpec()
	if err != nil {
		return err
	}
	required := map[BacklogPriority][]BacklogItemID{
		BacklogPriorityHigh: {
			BacklogItemZoneDescriptorCommitmentLayout,
			BacklogItemAetraCoreSkeleton,
			BacklogItemGlobalRootHierarchy,
			BacklogItemMsgbusEncoding,
			BacklogItemLocalMessageStores,
			BacklogItemStoreV2PrefixPlan,
			BacklogItemBlockSTMConflictTests,
			BacklogItemRoutingTableFormat,
			BacklogItemProofRegistrySchema,
		},
		BacklogPriorityMedium: {
			BacklogItemExtractFinancialZone,
			BacklogItemActivateIdentityZone,
			BacklogItemPerZoneMempoolLanes,
			BacklogItemPerShardFeeAccumulators,
			BacklogItemShardSplitMergeScheduler,
			BacklogItemAVMBytecodeGasTable,
			BacklogItemPaymentSettlementState,
			BacklogItemCrossZoneIdentityLookup,
		},
		BacklogPriorityLower: {
			BacklogItemDynamicRouteCapacityScoring,
			BacklogItemVirtualPaymentChannels,
			BacklogItemAdvancedABIIntrospection,
			BacklogItemVMNativeResolverContracts,
			BacklogItemValidatorServiceMetadata,
			BacklogItemZoneStateRentPolicies,
		},
	}
	itemsByPriority := make(map[BacklogPriority]map[BacklogItemID]struct{}, len(required))
	for _, item := range spec.Items {
		if _, found := itemsByPriority[item.Priority]; !found {
			itemsByPriority[item.Priority] = map[BacklogItemID]struct{}{}
		}
		itemsByPriority[item.Priority][item.ItemID] = struct{}{}
	}
	for priority, itemIDs := range required {
		seen := itemsByPriority[priority]
		for _, itemID := range itemIDs {
			if _, found := seen[itemID]; !found {
				return fmt.Errorf("aetracore implementation backlog coverage missing %s item %s", priority, itemID)
			}
		}
	}
	return nil
}

func ComputeImplementationBacklogRoot(items []BacklogItem) string {
	items = normalizeBacklogItems(items)
	return hashRoot("aetra-aek-implementation-backlog-v1", func(w byteWriter) {
		writeUint64(w, uint64(len(items)))
		for _, item := range items {
			writePart(w, string(item.Priority))
			writePart(w, string(item.ItemID))
			writePart(w, item.DescriptorHash)
		}
	})
}

func ComputeBacklogItemHash(item BacklogItem) string {
	item = item.Normalize()
	return hashRoot("aetra-aek-implementation-backlog-item-v1", func(w byteWriter) {
		writePart(w, string(item.Priority))
		writePart(w, string(item.ItemID))
		writePart(w, item.Task)
		writePart(w, item.Target)
		writeUint64(w, uint64(len(item.Acceptance)))
		for _, criterion := range item.Acceptance {
			writePart(w, criterion)
		}
	})
}

func IsBacklogPriority(priority BacklogPriority) bool {
	return priority == BacklogPriorityHigh || priority == BacklogPriorityMedium || priority == BacklogPriorityLower
}

func IsBacklogItemID(priority BacklogPriority, itemID BacklogItemID) bool {
	for _, known := range backlogItemIDsForPriority(priority) {
		if known == itemID {
			return true
		}
	}
	return false
}

func backlogItem(priority BacklogPriority, itemID BacklogItemID, task, target string, acceptance []string) BacklogItem {
	item := BacklogItem{
		Priority:	priority,
		ItemID:		itemID,
		Task:		task,
		Target:		target,
		Acceptance:	normalizeBacklogAcceptance(acceptance),
	}
	item.DescriptorHash = ComputeBacklogItemHash(item)
	return item
}

func normalizeBacklogItems(items []BacklogItem) []BacklogItem {
	normalized := make([]BacklogItem, len(items))
	for i, item := range items {
		normalized[i] = item.Normalize()
	}
	sort.Slice(normalized, func(i, j int) bool {
		left := normalized[i]
		right := normalized[j]
		if backlogPriorityRank(left.Priority) != backlogPriorityRank(right.Priority) {
			return backlogPriorityRank(left.Priority) < backlogPriorityRank(right.Priority)
		}
		return left.ItemID < right.ItemID
	})
	return normalized
}

func normalizeBacklogAcceptance(values []string) []string {
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		normalized := compactPerformanceText(value)
		if normalized == "" {
			continue
		}
		if _, found := seen[normalized]; found {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	sort.Strings(out)
	return out
}

func backlogPriorityRank(priority BacklogPriority) int {
	switch priority {
	case BacklogPriorityHigh:
		return 0
	case BacklogPriorityMedium:
		return 1
	case BacklogPriorityLower:
		return 2
	default:
		return 99
	}
}

func backlogItemIDsForPriority(priority BacklogPriority) []BacklogItemID {
	switch priority {
	case BacklogPriorityHigh:
		return []BacklogItemID{
			BacklogItemZoneDescriptorCommitmentLayout,
			BacklogItemAetraCoreSkeleton,
			BacklogItemGlobalRootHierarchy,
			BacklogItemMsgbusEncoding,
			BacklogItemLocalMessageStores,
			BacklogItemStoreV2PrefixPlan,
			BacklogItemBlockSTMConflictTests,
			BacklogItemRoutingTableFormat,
			BacklogItemProofRegistrySchema,
		}
	case BacklogPriorityMedium:
		return []BacklogItemID{
			BacklogItemExtractFinancialZone,
			BacklogItemActivateIdentityZone,
			BacklogItemPerZoneMempoolLanes,
			BacklogItemPerShardFeeAccumulators,
			BacklogItemShardSplitMergeScheduler,
			BacklogItemAVMBytecodeGasTable,
			BacklogItemPaymentSettlementState,
			BacklogItemCrossZoneIdentityLookup,
		}
	case BacklogPriorityLower:
		return []BacklogItemID{
			BacklogItemDynamicRouteCapacityScoring,
			BacklogItemVirtualPaymentChannels,
			BacklogItemAdvancedABIIntrospection,
			BacklogItemVMNativeResolverContracts,
			BacklogItemValidatorServiceMetadata,
			BacklogItemZoneStateRentPolicies,
		}
	default:
		return nil
	}
}
