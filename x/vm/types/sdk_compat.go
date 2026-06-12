package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

const (
	LifecyclePrepareProposal	= "prepare_proposal"
	LifecycleProcessProposal	= "process_proposal"
	LifecycleFinalizeBlock		= "finalize_block"
	LifecycleCommit			= "commit"

	StateAccessStoreV2	= "store_v2"
	StateAccessKVStore	= "kvstore"

	DefaultAVMModuleName	= "vm"
	DefaultAVMKeeperName	= "VMKeeper"
	DefaultAVMMsgServerName	= "MsgServer"
	DefaultAVMStoreKey	= "vm"
)

type LifecycleStage string
type StateAccessKind string

type KVLayoutEntry struct {
	StoreKey	string
	Prefix		string
	Purpose		string
}

type SDKIntegrationBinding struct {
	ModuleName			string
	KeeperName			string
	MsgServerName			string
	StoreKey			string
	Lifecycle			[]LifecycleStage
	StateAccess			[]StateAccessKind
	KVLayout			[]KVLayoutEntry
	BlockSTMConflictPrefixes	[]string
	StakingFinalityRequired		bool
	CometBFTFinalityRequired	bool
}

type SDKDispatch struct {
	ZoneID		zonestypes.ZoneID
	MsgType		string
	Call		VMCall
	KVPrefix	string
	BlockSTMKey	string
	StakingPower	uint64
	ExecutionHeight	uint64
}

type FinalizeBlockPlan struct {
	Height		uint64
	Binding		SDKIntegrationBinding
	Dispatches	[]SDKDispatch
}

func DefaultAVMSDKBinding(zoneID zonestypes.ZoneID) SDKIntegrationBinding {
	return SDKIntegrationBinding{
		ModuleName:	DefaultAVMModuleName,
		KeeperName:	DefaultAVMKeeperName,
		MsgServerName:	DefaultAVMMsgServerName,
		StoreKey:	DefaultAVMStoreKey,
		Lifecycle: []LifecycleStage{
			LifecycleCommit,
			LifecycleFinalizeBlock,
			LifecyclePrepareProposal,
			LifecycleProcessProposal,
		},
		StateAccess: []StateAccessKind{
			StateAccessKVStore,
			StateAccessStoreV2,
		},
		KVLayout:			AVMKVStoreLayout(zoneID),
		BlockSTMConflictPrefixes:	[]string{ContractZoneKVPrefix(zoneID)},
		StakingFinalityRequired:	true,
		CometBFTFinalityRequired:	true,
	}
}

func AVMKVStoreLayout(zoneID zonestypes.ZoneID) []KVLayoutEntry {
	prefix := ContractZoneKVPrefix(zoneID)
	return []KVLayoutEntry{
		{StoreKey: DefaultAVMStoreKey, Prefix: prefix + "code/", Purpose: "contract-code"},
		{StoreKey: DefaultAVMStoreKey, Prefix: prefix + "instance/", Purpose: "contract-instance"},
		{StoreKey: DefaultAVMStoreKey, Prefix: prefix + "queue/", Purpose: "message-queue"},
		{StoreKey: DefaultAVMStoreKey, Prefix: prefix + "receipt/", Purpose: "execution-receipt"},
		{StoreKey: DefaultAVMStoreKey, Prefix: prefix + "root/", Purpose: "proof-root"},
	}
}

func ContractZoneKVPrefix(zoneID zonestypes.ZoneID) string {
	return "vm/" + string(zoneID) + "/"
}

func BuildFinalizeBlockPlan(height uint64, binding SDKIntegrationBinding, dispatches []SDKDispatch, policy RuntimePolicy) (FinalizeBlockPlan, error) {
	if height == 0 {
		return FinalizeBlockPlan{}, errors.New("FinalizeBlock height must be positive")
	}
	if err := binding.Validate(); err != nil {
		return FinalizeBlockPlan{}, err
	}
	out := make([]SDKDispatch, len(dispatches))
	for i, dispatch := range dispatches {
		dispatch.ExecutionHeight = height
		if dispatch.KVPrefix == "" {
			dispatch.KVPrefix = ContractZoneKVPrefix(dispatch.ZoneID)
		}
		if dispatch.BlockSTMKey == "" {
			dispatch.BlockSTMKey = dispatch.KVPrefix
		}
		if err := dispatch.Validate(policy); err != nil {
			return FinalizeBlockPlan{}, err
		}
		out[i] = dispatch
	}
	sort.SliceStable(out, func(i, j int) bool {
		return compareDispatch(out[i], out[j]) < 0
	})
	plan := FinalizeBlockPlan{Height: height, Binding: binding, Dispatches: out}
	return plan, plan.Validate(policy)
}

func (b SDKIntegrationBinding) Validate() error {
	for _, item := range []struct {
		name	string
		value	string
	}{
		{name: "SDK module name", value: b.ModuleName},
		{name: "SDK keeper name", value: b.KeeperName},
		{name: "SDK MsgServer name", value: b.MsgServerName},
		{name: "SDK store key", value: b.StoreKey},
	} {
		if err := validateSDKToken(item.name, item.value); err != nil {
			return err
		}
	}
	if !containsLifecycle(b.Lifecycle, LifecycleFinalizeBlock) {
		return errors.New("AVM SDK binding must include FinalizeBlock lifecycle")
	}
	if !containsStateAccess(b.StateAccess, StateAccessStoreV2) {
		return errors.New("AVM SDK binding must support Store v2 state access")
	}
	if !containsStateAccess(b.StateAccess, StateAccessKVStore) {
		return errors.New("AVM SDK binding must support KVStore-compatible state access")
	}
	if len(b.KVLayout) == 0 {
		return errors.New("AVM SDK binding must define KVStore layout")
	}
	if err := validateLifecycle(b.Lifecycle); err != nil {
		return err
	}
	if err := validateStateAccess(b.StateAccess); err != nil {
		return err
	}
	if err := validateKVLayout(b.StoreKey, b.KVLayout); err != nil {
		return err
	}
	if err := validateConflictPrefixes(b.BlockSTMConflictPrefixes); err != nil {
		return err
	}
	if !b.StakingFinalityRequired {
		return errors.New("AVM SDK binding must require staking-secured execution")
	}
	if !b.CometBFTFinalityRequired {
		return errors.New("AVM SDK binding must require CometBFT finality")
	}
	return nil
}

func (d SDKDispatch) Validate(policy RuntimePolicy) error {
	if err := zonestypes.ValidateZoneID(d.ZoneID); err != nil {
		return err
	}
	if err := validateSDKToken("SDK dispatch message type", d.MsgType); err != nil {
		return err
	}
	if d.KVPrefix != ContractZoneKVPrefix(d.ZoneID) {
		return fmt.Errorf("SDK dispatch KV prefix must be %q", ContractZoneKVPrefix(d.ZoneID))
	}
	if strings.TrimSpace(d.BlockSTMKey) != d.BlockSTMKey || d.BlockSTMKey == "" {
		return errors.New("SDK dispatch BlockSTM key is required")
	}
	if !strings.HasPrefix(d.BlockSTMKey, d.KVPrefix) {
		return errors.New("SDK dispatch BlockSTM key must stay within zone KV prefix")
	}
	if d.ExecutionHeight == 0 {
		return errors.New("SDK dispatch execution height must be positive")
	}
	if d.StakingPower == 0 {
		return errors.New("SDK dispatch requires staking voting power")
	}
	return ValidateVMCall(d.Call, policy)
}

func (p FinalizeBlockPlan) Validate(policy RuntimePolicy) error {
	if p.Height == 0 {
		return errors.New("FinalizeBlock plan height must be positive")
	}
	if err := p.Binding.Validate(); err != nil {
		return err
	}
	seenKeys := make([]string, 0, len(p.Dispatches))
	for i, dispatch := range p.Dispatches {
		if dispatch.ExecutionHeight != p.Height {
			return errors.New("FinalizeBlock dispatch height drift")
		}
		if err := dispatch.Validate(policy); err != nil {
			return err
		}
		if i > 0 && compareDispatch(p.Dispatches[i-1], dispatch) >= 0 {
			return errors.New("FinalizeBlock dispatches must be sorted canonically")
		}
		for _, existing := range seenKeys {
			if prefixesOverlap(existing, dispatch.BlockSTMKey) {
				return fmt.Errorf("BlockSTM conflict key %q overlaps %q", dispatch.BlockSTMKey, existing)
			}
		}
		seenKeys = append(seenKeys, dispatch.BlockSTMKey)
	}
	return nil
}

func validateLifecycle(stages []LifecycleStage) error {
	if len(stages) == 0 {
		return errors.New("SDK lifecycle stages are required")
	}
	seen := make(map[LifecycleStage]struct{}, len(stages))
	var previous LifecycleStage
	for i, stage := range stages {
		if !isLifecycleStage(stage) {
			return fmt.Errorf("unknown SDK lifecycle stage %q", stage)
		}
		if _, found := seen[stage]; found {
			return fmt.Errorf("duplicate SDK lifecycle stage %q", stage)
		}
		seen[stage] = struct{}{}
		if i > 0 && previous >= stage {
			return errors.New("SDK lifecycle stages must be sorted canonically")
		}
		previous = stage
	}
	return nil
}

func validateStateAccess(access []StateAccessKind) error {
	if len(access) == 0 {
		return errors.New("SDK state access list is required")
	}
	seen := make(map[StateAccessKind]struct{}, len(access))
	var previous StateAccessKind
	for i, item := range access {
		if !isStateAccessKind(item) {
			return fmt.Errorf("unknown SDK state access %q", item)
		}
		if _, found := seen[item]; found {
			return fmt.Errorf("duplicate SDK state access %q", item)
		}
		seen[item] = struct{}{}
		if i > 0 && previous >= item {
			return errors.New("SDK state access list must be sorted canonically")
		}
		previous = item
	}
	return nil
}

func validateKVLayout(storeKey string, layout []KVLayoutEntry) error {
	seen := make(map[string]struct{}, len(layout))
	var previous string
	for i, entry := range layout {
		if entry.StoreKey != storeKey {
			return errors.New("KV layout store key mismatch")
		}
		if err := validateSDKToken("KV layout purpose", entry.Purpose); err != nil {
			return err
		}
		if strings.TrimSpace(entry.Prefix) != entry.Prefix || entry.Prefix == "" {
			return errors.New("KV layout prefix is required")
		}
		if _, found := seen[entry.Prefix]; found {
			return fmt.Errorf("duplicate KV layout prefix %q", entry.Prefix)
		}
		seen[entry.Prefix] = struct{}{}
		if i > 0 && previous >= entry.Prefix {
			return errors.New("KV layout must be sorted canonically")
		}
		previous = entry.Prefix
	}
	return nil
}

func validateConflictPrefixes(prefixes []string) error {
	if len(prefixes) == 0 {
		return errors.New("BlockSTM conflict prefixes are required")
	}
	seen := make(map[string]struct{}, len(prefixes))
	var previous string
	for i, prefix := range prefixes {
		if strings.TrimSpace(prefix) != prefix || prefix == "" {
			return errors.New("BlockSTM conflict prefix is required")
		}
		if _, found := seen[prefix]; found {
			return fmt.Errorf("duplicate BlockSTM conflict prefix %q", prefix)
		}
		seen[prefix] = struct{}{}
		if i > 0 && previous >= prefix {
			return errors.New("BlockSTM conflict prefixes must be sorted canonically")
		}
		previous = prefix
	}
	return nil
}

func containsLifecycle(stages []LifecycleStage, target LifecycleStage) bool {
	for _, stage := range stages {
		if stage == target {
			return true
		}
	}
	return false
}

func containsStateAccess(access []StateAccessKind, target StateAccessKind) bool {
	for _, item := range access {
		if item == target {
			return true
		}
	}
	return false
}

func isLifecycleStage(stage LifecycleStage) bool {
	switch stage {
	case LifecycleCommit, LifecycleFinalizeBlock, LifecyclePrepareProposal, LifecycleProcessProposal:
		return true
	default:
		return false
	}
}

func isStateAccessKind(kind StateAccessKind) bool {
	switch kind {
	case StateAccessKVStore, StateAccessStoreV2:
		return true
	default:
		return false
	}
}

func compareDispatch(left, right SDKDispatch) int {
	if left.ZoneID < right.ZoneID {
		return -1
	}
	if left.ZoneID > right.ZoneID {
		return 1
	}
	if left.MsgType < right.MsgType {
		return -1
	}
	if left.MsgType > right.MsgType {
		return 1
	}
	if left.BlockSTMKey < right.BlockSTMKey {
		return -1
	}
	if left.BlockSTMKey > right.BlockSTMKey {
		return 1
	}
	return 0
}

func prefixesOverlap(left, right string) bool {
	return strings.HasPrefix(left, right) || strings.HasPrefix(right, left)
}

func validateSDKToken(fieldName, value string) error {
	if strings.TrimSpace(value) != value || value == "" {
		return fmt.Errorf("%s is required and must not have surrounding whitespace", fieldName)
	}
	for _, r := range value {
		if r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_' || r == '-' || r == '.' || r == ':' || r == '/' {
			continue
		}
		return fmt.Errorf("%s contains invalid character", fieldName)
	}
	return nil
}
