package types

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
)

type ShardAssignmentMode string

const (
	ShardAssignmentConsistentHash	ShardAssignmentMode	= "consistent_hash"
	ShardAssignmentKeyPrefix	ShardAssignmentMode	= "key_prefix"
	ShardAssignmentExplicit		ShardAssignmentMode	= "explicit"
)

type ShardPlacementOverride struct {
	ObjectKey		string
	ShardID			ShardID
	ReadOnlyReplicated	bool
}

type ShardRoutingInput struct {
	ZoneID			ZoneID
	StateKey		string
	ShardLayoutEpoch	uint64
	AssignmentMode		ShardAssignmentMode
	PlacementOverride	ShardID
}

type ShardRoute struct {
	ZoneID			ZoneID
	StateKey		string
	LayoutEpoch		uint64
	AssignmentMode		ShardAssignmentMode
	ShardID			ShardID
	ShardCount		uint32
	PlacementOverride	ShardID
	ReadOnlyReplicated	bool
	RouteHash		string
}

type ShardMetrics struct {
	ZoneID			ZoneID
	ShardID			ShardID
	Height			uint64
	GasUsed			uint64
	FeeCollected		uint64
	InboxBacklog		uint64
	OutboxBacklog		uint64
	WriteConflictCount	uint64
	StateSizeBytes		uint64
	ProofLatencyMicros	uint64
	ExecutionDelayMicros	uint64
	FailedDeliveryCount	uint64
	ExpiredMessageCount	uint64
	MetricsHash		string
}

type ShardDescriptor struct {
	ShardID			ShardID
	StatePrefix		string
	ParentShardID		ShardID
	ActivationHeight	uint64
	ValidatorSetHash	string
	Available		bool
	KeyPrefix		string
	HashRangeStart		uint64
	HashRangeEnd		uint64
	SystemShard		bool
}

type ShardLayout struct {
	ZoneID			ZoneID
	LayoutEpoch		uint64
	ActivationHeight	uint64
	RoutingSeedHash		string
	AssignmentMode		ShardAssignmentMode
	SystemShardID		ShardID
	PlacementOverrides	[]ShardPlacementOverride
	ReadOnlyReplicatedKeys	[]string
	ActiveShards		[]ShardDescriptor
	LayoutHash		string
}

type RoutingZoneEntry struct {
	ZoneID		ZoneID
	LayoutEpoch	uint64
	ActiveShards	uint32
	LayoutHash	string
}

type RoutingTableCommitment struct {
	RoutingEpoch	uint64
	Height		uint64
	Entries		[]RoutingZoneEntry
	TableHash	string
}

func NewShardLayout(zoneID ZoneID, layoutEpoch uint64, activationHeight uint64, routingSeedHash string, shards []ShardDescriptor) (ShardLayout, error) {
	layout := ShardLayout{
		ZoneID:			zoneID,
		LayoutEpoch:		layoutEpoch,
		ActivationHeight:	activationHeight,
		RoutingSeedHash:	routingSeedHash,
		AssignmentMode:		ShardAssignmentConsistentHash,
		ActiveShards:		cloneShardDescriptors(shards),
	}
	sortShardDescriptors(layout.ActiveShards)
	if err := layout.ValidateFormat(); err != nil {
		return ShardLayout{}, err
	}
	layout.LayoutHash = ComputeShardLayoutHash(layout)
	return layout, nil
}

func NewRoutingTableCommitment(routingEpoch uint64, height uint64, entries []RoutingZoneEntry) (RoutingTableCommitment, error) {
	table := RoutingTableCommitment{
		RoutingEpoch:	routingEpoch,
		Height:		height,
		Entries:	cloneRoutingZoneEntries(entries),
	}
	sortRoutingZoneEntries(table.Entries)
	if err := table.ValidateFormat(); err != nil {
		return RoutingTableCommitment{}, err
	}
	table.TableHash = ComputeRoutingTableHash(table)
	return table, nil
}

func BuildRoutingTableCommitment(routingEpoch uint64, height uint64, layouts []ShardLayout) (RoutingTableCommitment, error) {
	entries := make([]RoutingZoneEntry, len(layouts))
	for i, layout := range layouts {
		if err := layout.ValidateHash(); err != nil {
			return RoutingTableCommitment{}, err
		}
		entries[i] = RoutingZoneEntry{
			ZoneID:		layout.ZoneID,
			LayoutEpoch:	layout.LayoutEpoch,
			ActiveShards:	uint32(len(layout.ActiveShards)),
			LayoutHash:	layout.LayoutHash,
		}
	}
	return NewRoutingTableCommitment(routingEpoch, height, entries)
}

func (d ShardDescriptor) Validate() error {
	if err := ValidateShardID(d.ShardID); err != nil {
		return err
	}
	if err := validateToken("aetracore shard state prefix", d.StatePrefix, MaxScopeLength); err != nil {
		return err
	}
	if d.ParentShardID != "" {
		if err := ValidateShardID(d.ParentShardID); err != nil {
			return err
		}
	}
	if d.ActivationHeight == 0 {
		return errors.New("aetracore shard activation height must be positive")
	}
	if d.KeyPrefix != "" {
		if err := validateToken("aetracore shard key prefix", d.KeyPrefix, MaxScopeLength); err != nil {
			return err
		}
	}
	if d.HashRangeEnd < d.HashRangeStart {
		return errors.New("aetracore shard hash range end must not precede start")
	}
	return ValidateHash("aetracore shard validator set hash", d.ValidatorSetHash)
}

func (l ShardLayout) ValidateFormat() error {
	if err := ValidateZoneID(l.ZoneID); err != nil {
		return err
	}
	if l.LayoutEpoch == 0 {
		return errors.New("aetracore shard layout epoch must be positive")
	}
	if l.ActivationHeight == 0 {
		return errors.New("aetracore shard layout activation height must be positive")
	}
	if err := ValidateHash("aetracore shard layout routing seed", l.RoutingSeedHash); err != nil {
		return err
	}
	mode := l.AssignmentMode
	if mode == "" {
		mode = ShardAssignmentConsistentHash
	}
	if !IsShardAssignmentMode(mode) {
		return fmt.Errorf("unknown aetracore shard assignment mode %q", mode)
	}
	if len(l.ActiveShards) == 0 {
		return errors.New("aetracore shard layout requires active shards")
	}
	if err := validateShardDescriptors(l.ActiveShards); err != nil {
		return err
	}
	if l.SystemShardID != "" && !l.HasActiveShard(l.SystemShardID) {
		return errors.New("aetracore shard layout system shard must be active")
	}
	if err := validateShardPlacementOverrides(l.PlacementOverrides, l); err != nil {
		return err
	}
	if err := validateReadOnlyReplicatedKeys(l.ReadOnlyReplicatedKeys); err != nil {
		return err
	}
	if l.LayoutHash != "" {
		return ValidateHash("aetracore shard layout hash", l.LayoutHash)
	}
	return nil
}

func (l ShardLayout) ValidateHash() error {
	if err := l.ValidateFormat(); err != nil {
		return err
	}
	expected := ComputeShardLayoutHash(l)
	if l.LayoutHash != expected {
		return fmt.Errorf("aetracore shard layout hash mismatch: expected %s", expected)
	}
	return nil
}

func (e RoutingZoneEntry) Validate() error {
	if err := ValidateZoneID(e.ZoneID); err != nil {
		return err
	}
	if e.LayoutEpoch == 0 {
		return errors.New("aetracore routing entry layout epoch must be positive")
	}
	if e.ActiveShards == 0 {
		return errors.New("aetracore routing entry active shards must be positive")
	}
	return ValidateHash("aetracore routing entry layout hash", e.LayoutHash)
}

func (t RoutingTableCommitment) ValidateFormat() error {
	if t.RoutingEpoch == 0 {
		return errors.New("aetracore routing table epoch must be positive")
	}
	if t.Height == 0 {
		return errors.New("aetracore routing table height must be positive")
	}
	if len(t.Entries) == 0 {
		return errors.New("aetracore routing table requires entries")
	}
	if err := validateRoutingZoneEntries(t.Entries); err != nil {
		return err
	}
	if t.TableHash != "" {
		return ValidateHash("aetracore routing table hash", t.TableHash)
	}
	return nil
}

func (t RoutingTableCommitment) ValidateHash() error {
	if err := t.ValidateFormat(); err != nil {
		return err
	}
	expected := ComputeRoutingTableHash(t)
	if t.TableHash != expected {
		return fmt.Errorf("aetracore routing table hash mismatch: expected %s", expected)
	}
	return nil
}

func ComputeShardLayoutHash(layout ShardLayout) string {
	ordered := cloneShardDescriptors(layout.ActiveShards)
	sortShardDescriptors(ordered)
	overrides := cloneShardPlacementOverrides(layout.PlacementOverrides)
	sortShardPlacementOverrides(overrides)
	readOnly := append([]string(nil), layout.ReadOnlyReplicatedKeys...)
	sort.Strings(readOnly)
	mode := layout.AssignmentMode
	if mode == "" {
		mode = ShardAssignmentConsistentHash
	}
	parts := []string{
		"aetra-aek-shard-layout-v1",
		string(layout.ZoneID),
		fmt.Sprint(layout.LayoutEpoch),
		fmt.Sprint(layout.ActivationHeight),
		layout.RoutingSeedHash,
		string(mode),
		string(layout.SystemShardID),
		fmt.Sprint(len(ordered)),
	}
	for _, shard := range ordered {
		parts = append(parts,
			string(shard.ShardID),
			shard.StatePrefix,
			string(shard.ParentShardID),
			fmt.Sprint(shard.ActivationHeight),
			shard.ValidatorSetHash,
			fmt.Sprint(shard.Available),
			shard.KeyPrefix,
			fmt.Sprint(shard.HashRangeStart),
			fmt.Sprint(shard.HashRangeEnd),
			fmt.Sprint(shard.SystemShard),
		)
	}
	parts = append(parts, fmt.Sprint(len(overrides)))
	for _, override := range overrides {
		parts = append(parts, override.ObjectKey, string(override.ShardID), fmt.Sprint(override.ReadOnlyReplicated))
	}
	parts = append(parts, fmt.Sprint(len(readOnly)))
	parts = append(parts, readOnly...)
	return hashParts(parts...)
}

func RouteKeyToShard(layout ShardLayout, input ShardRoutingInput) (ShardRoute, error) {
	layout.ActiveShards = cloneShardDescriptors(layout.ActiveShards)
	sortShardDescriptors(layout.ActiveShards)
	if err := layout.ValidateHash(); err != nil {
		return ShardRoute{}, err
	}
	if input.ZoneID != layout.ZoneID {
		return ShardRoute{}, errors.New("aetracore shard route zone mismatch")
	}
	if input.ShardLayoutEpoch != layout.LayoutEpoch {
		return ShardRoute{}, errors.New("aetracore shard route layout epoch mismatch")
	}
	if err := validateToken("aetracore shard route state key", input.StateKey, MaxScopeLength); err != nil {
		return ShardRoute{}, err
	}
	mode := input.AssignmentMode
	if mode == "" {
		mode = layout.AssignmentMode
	}
	if mode == "" {
		mode = ShardAssignmentConsistentHash
	}
	if !IsShardAssignmentMode(mode) {
		return ShardRoute{}, fmt.Errorf("unknown aetracore shard assignment mode %q", mode)
	}
	if input.PlacementOverride != "" {
		if !layout.HasActiveShard(input.PlacementOverride) {
			return ShardRoute{}, errors.New("aetracore shard route placement override is not active")
		}
		return newShardRoute(layout, input, ShardAssignmentExplicit, input.PlacementOverride, input.PlacementOverride, isReadOnlyReplicatedKey(layout, input.StateKey))
	}
	if override, found := shardPlacementOverrideFor(layout, input.StateKey); found {
		return newShardRoute(layout, input, ShardAssignmentExplicit, override.ShardID, override.ShardID, override.ReadOnlyReplicated)
	}
	if isReadOnlyReplicatedKey(layout, input.StateKey) {
		systemShard, err := systemShardID(layout)
		if err != nil {
			return ShardRoute{}, err
		}
		return newShardRoute(layout, input, ShardAssignmentExplicit, systemShard, "", true)
	}
	switch mode {
	case ShardAssignmentExplicit:
		systemShard, err := systemShardID(layout)
		if err != nil {
			return ShardRoute{}, err
		}
		return newShardRoute(layout, input, mode, systemShard, "", false)
	case ShardAssignmentKeyPrefix:
		shardID, err := routeKeyPrefixShard(layout, input.StateKey)
		if err != nil {
			return ShardRoute{}, err
		}
		return newShardRoute(layout, input, mode, shardID, "", false)
	case ShardAssignmentConsistentHash:
		shardID, err := routeConsistentHashShard(layout, input.StateKey)
		if err != nil {
			return ShardRoute{}, err
		}
		return newShardRoute(layout, input, mode, shardID, "", false)
	default:
		return ShardRoute{}, fmt.Errorf("unknown aetracore shard assignment mode %q", mode)
	}
}

func (r ShardRoute) ValidateHash() error {
	if err := ValidateZoneID(r.ZoneID); err != nil {
		return err
	}
	if r.LayoutEpoch == 0 {
		return errors.New("aetracore shard route layout epoch must be positive")
	}
	if !IsShardAssignmentMode(r.AssignmentMode) {
		return fmt.Errorf("unknown aetracore shard route mode %q", r.AssignmentMode)
	}
	if err := ValidateShardID(r.ShardID); err != nil {
		return err
	}
	if r.ShardCount == 0 {
		return errors.New("aetracore shard route shard count must be positive")
	}
	if err := validateToken("aetracore shard route state key", r.StateKey, MaxScopeLength); err != nil {
		return err
	}
	if r.PlacementOverride != "" {
		if err := ValidateShardID(r.PlacementOverride); err != nil {
			return err
		}
	}
	if err := ValidateHash("aetracore shard route hash", r.RouteHash); err != nil {
		return err
	}
	if r.RouteHash != ComputeShardRouteHash(r) {
		return errors.New("aetracore shard route hash mismatch")
	}
	return nil
}

func NewShardMetrics(metrics ShardMetrics) (ShardMetrics, error) {
	if metrics.MetricsHash != "" {
		return ShardMetrics{}, errors.New("aetracore shard metrics hash must be empty before construction")
	}
	if err := metrics.ValidateFormat(); err != nil {
		return ShardMetrics{}, err
	}
	metrics.MetricsHash = ComputeShardMetricsHash(metrics)
	return metrics, metrics.ValidateHash()
}

func (m ShardMetrics) ValidateFormat() error {
	if err := ValidateZoneID(m.ZoneID); err != nil {
		return err
	}
	if err := ValidateShardID(m.ShardID); err != nil {
		return err
	}
	if m.Height == 0 {
		return errors.New("aetracore shard metrics height must be positive")
	}
	if m.MetricsHash != "" {
		return ValidateHash("aetracore shard metrics hash", m.MetricsHash)
	}
	return nil
}

func (m ShardMetrics) ValidateHash() error {
	if err := m.ValidateFormat(); err != nil {
		return err
	}
	if m.MetricsHash != ComputeShardMetricsHash(m) {
		return errors.New("aetracore shard metrics hash mismatch")
	}
	return nil
}

func ComputeShardRouteHash(route ShardRoute) string {
	return hashParts(
		"aetra-aek-shard-route-v1",
		string(route.ZoneID),
		route.StateKey,
		fmt.Sprint(route.LayoutEpoch),
		string(route.AssignmentMode),
		string(route.ShardID),
		fmt.Sprint(route.ShardCount),
		string(route.PlacementOverride),
		fmt.Sprint(route.ReadOnlyReplicated),
	)
}

func ComputeShardMetricsHash(metrics ShardMetrics) string {
	return hashParts(
		"aetra-aek-shard-metrics-v1",
		string(metrics.ZoneID),
		string(metrics.ShardID),
		fmt.Sprint(metrics.Height),
		fmt.Sprint(metrics.GasUsed),
		fmt.Sprint(metrics.FeeCollected),
		fmt.Sprint(metrics.InboxBacklog),
		fmt.Sprint(metrics.OutboxBacklog),
		fmt.Sprint(metrics.WriteConflictCount),
		fmt.Sprint(metrics.StateSizeBytes),
		fmt.Sprint(metrics.ProofLatencyMicros),
		fmt.Sprint(metrics.ExecutionDelayMicros),
		fmt.Sprint(metrics.FailedDeliveryCount),
		fmt.Sprint(metrics.ExpiredMessageCount),
	)
}

func IsShardAssignmentMode(mode ShardAssignmentMode) bool {
	switch mode {
	case ShardAssignmentConsistentHash, ShardAssignmentKeyPrefix, ShardAssignmentExplicit:
		return true
	default:
		return false
	}
}

func ComputeRoutingTableHash(table RoutingTableCommitment) string {
	ordered := cloneRoutingZoneEntries(table.Entries)
	sortRoutingZoneEntries(ordered)
	parts := []string{
		"aetra-aek-routing-table-v1",
		fmt.Sprint(table.RoutingEpoch),
		fmt.Sprint(table.Height),
		fmt.Sprint(len(ordered)),
	}
	for _, entry := range ordered {
		parts = append(parts,
			string(entry.ZoneID),
			fmt.Sprint(entry.LayoutEpoch),
			fmt.Sprint(entry.ActiveShards),
			entry.LayoutHash,
		)
	}
	return hashParts(parts...)
}

func validateShardDescriptors(shards []ShardDescriptor) error {
	var previous ShardID
	seen := make(map[ShardID]struct{}, len(shards))
	for i, shard := range shards {
		if err := shard.Validate(); err != nil {
			return err
		}
		if _, found := seen[shard.ShardID]; found {
			return fmt.Errorf("duplicate aetracore shard id %s", shard.ShardID)
		}
		seen[shard.ShardID] = struct{}{}
		if i > 0 && previous >= shard.ShardID {
			return errors.New("aetracore shard descriptors must be sorted canonically")
		}
		previous = shard.ShardID
	}
	return nil
}

func validateRoutingZoneEntries(entries []RoutingZoneEntry) error {
	var previous ZoneID
	seen := make(map[ZoneID]struct{}, len(entries))
	for i, entry := range entries {
		if err := entry.Validate(); err != nil {
			return err
		}
		if _, found := seen[entry.ZoneID]; found {
			return fmt.Errorf("duplicate aetracore routing zone entry %s", entry.ZoneID)
		}
		seen[entry.ZoneID] = struct{}{}
		if i > 0 && previous >= entry.ZoneID {
			return errors.New("aetracore routing zone entries must be sorted canonically")
		}
		previous = entry.ZoneID
	}
	return nil
}

func validateShardPlacementOverrides(overrides []ShardPlacementOverride, layout ShardLayout) error {
	ordered := cloneShardPlacementOverrides(overrides)
	sortShardPlacementOverrides(ordered)
	var previous string
	for i, override := range ordered {
		if err := validateToken("aetracore shard placement object key", override.ObjectKey, MaxScopeLength); err != nil {
			return err
		}
		if err := ValidateShardID(override.ShardID); err != nil {
			return err
		}
		if !layout.HasActiveShard(override.ShardID) {
			return errors.New("aetracore shard placement override shard must be active")
		}
		if i > 0 && previous >= override.ObjectKey {
			return errors.New("aetracore shard placement overrides must be sorted canonically")
		}
		previous = override.ObjectKey
	}
	return nil
}

func validateReadOnlyReplicatedKeys(keys []string) error {
	ordered := append([]string(nil), keys...)
	sort.Strings(ordered)
	var previous string
	for i, key := range ordered {
		if err := validateToken("aetracore read-only replicated key", key, MaxScopeLength); err != nil {
			return err
		}
		if i > 0 && previous >= key {
			return errors.New("aetracore read-only replicated keys must be sorted canonically")
		}
		previous = key
	}
	return nil
}

func sortShardDescriptors(shards []ShardDescriptor) {
	sort.SliceStable(shards, func(i, j int) bool {
		return shards[i].ShardID < shards[j].ShardID
	})
}

func sortShardLayouts(layouts []ShardLayout) {
	sort.SliceStable(layouts, func(i, j int) bool {
		if layouts[i].ZoneID == layouts[j].ZoneID {
			return layouts[i].LayoutEpoch < layouts[j].LayoutEpoch
		}
		return layouts[i].ZoneID < layouts[j].ZoneID
	})
}

func sortRoutingZoneEntries(entries []RoutingZoneEntry) {
	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].ZoneID < entries[j].ZoneID
	})
}

func sortRoutingTables(tables []RoutingTableCommitment) {
	sort.SliceStable(tables, func(i, j int) bool {
		return tables[i].RoutingEpoch < tables[j].RoutingEpoch
	})
}

func cloneShardDescriptors(shards []ShardDescriptor) []ShardDescriptor {
	out := make([]ShardDescriptor, len(shards))
	copy(out, shards)
	return out
}

func cloneShardLayouts(layouts []ShardLayout) []ShardLayout {
	out := make([]ShardLayout, len(layouts))
	for i, layout := range layouts {
		out[i] = layout
		out[i].ActiveShards = cloneShardDescriptors(layout.ActiveShards)
		out[i].PlacementOverrides = cloneShardPlacementOverrides(layout.PlacementOverrides)
		out[i].ReadOnlyReplicatedKeys = append([]string(nil), layout.ReadOnlyReplicatedKeys...)
	}
	return out
}

func cloneShardPlacementOverrides(overrides []ShardPlacementOverride) []ShardPlacementOverride {
	out := make([]ShardPlacementOverride, len(overrides))
	copy(out, overrides)
	return out
}

func cloneRoutingZoneEntries(entries []RoutingZoneEntry) []RoutingZoneEntry {
	out := make([]RoutingZoneEntry, len(entries))
	copy(out, entries)
	return out
}

func cloneRoutingTables(tables []RoutingTableCommitment) []RoutingTableCommitment {
	out := make([]RoutingTableCommitment, len(tables))
	for i, table := range tables {
		out[i] = table
		out[i].Entries = cloneRoutingZoneEntries(table.Entries)
	}
	return out
}

func ExportShardLayouts(layouts []ShardLayout) []ShardLayout {
	out := cloneShardLayouts(layouts)
	sortShardLayouts(out)
	return out
}

func ExportRoutingTables(tables []RoutingTableCommitment) []RoutingTableCommitment {
	out := cloneRoutingTables(tables)
	sortRoutingTables(out)
	return out
}

func sortShardPlacementOverrides(overrides []ShardPlacementOverride) {
	sort.SliceStable(overrides, func(i, j int) bool { return overrides[i].ObjectKey < overrides[j].ObjectKey })
}

func routeKeyPrefixShard(layout ShardLayout, stateKey string) (ShardID, error) {
	var selected ShardDescriptor
	for _, shard := range layout.ActiveShards {
		if !shard.Available {
			continue
		}
		prefix := shard.KeyPrefix
		if prefix == "" {
			prefix = shard.StatePrefix
		}
		if prefix == "" || !strings.HasPrefix(stateKey, prefix) {
			continue
		}
		if selected.ShardID == "" || len(prefix) > len(selected.KeyPrefix) {
			selected = shard
		}
	}
	if selected.ShardID == "" {
		return "", errors.New("aetracore shard prefix route has no matching shard")
	}
	return selected.ShardID, nil
}

func routeConsistentHashShard(layout ShardLayout, stateKey string) (ShardID, error) {
	active := availableShardDescriptors(layout.ActiveShards)
	if len(active) == 0 {
		return "", errors.New("aetracore shard route requires available shards")
	}
	routeHash := hashParts("aetra-aek-route-key-v1", string(layout.ZoneID), fmt.Sprint(layout.LayoutEpoch), layout.RoutingSeedHash, stateKey)
	bytes, err := hex.DecodeString(routeHash[:16])
	if err != nil {
		return "", err
	}
	slot := binary.BigEndian.Uint64(bytes)
	for _, shard := range active {
		if shard.HashRangeEnd == 0 && shard.HashRangeStart == 0 {
			continue
		}
		if slot >= shard.HashRangeStart && slot <= shard.HashRangeEnd {
			return shard.ShardID, nil
		}
	}
	return active[slot%uint64(len(active))].ShardID, nil
}

func systemShardID(layout ShardLayout) (ShardID, error) {
	if layout.SystemShardID != "" {
		return layout.SystemShardID, nil
	}
	for _, shard := range layout.ActiveShards {
		if shard.Available && shard.SystemShard {
			return shard.ShardID, nil
		}
	}
	active := availableShardDescriptors(layout.ActiveShards)
	if len(active) == 0 {
		return "", errors.New("aetracore shard route requires available shards")
	}
	return active[0].ShardID, nil
}

func shardPlacementOverrideFor(layout ShardLayout, stateKey string) (ShardPlacementOverride, bool) {
	overrides := cloneShardPlacementOverrides(layout.PlacementOverrides)
	sortShardPlacementOverrides(overrides)
	for _, override := range overrides {
		if override.ObjectKey == stateKey || strings.HasPrefix(stateKey, override.ObjectKey+"/") {
			return override, true
		}
	}
	return ShardPlacementOverride{}, false
}

func isReadOnlyReplicatedKey(layout ShardLayout, stateKey string) bool {
	keys := append([]string(nil), layout.ReadOnlyReplicatedKeys...)
	sort.Strings(keys)
	for _, key := range keys {
		if key == stateKey || strings.HasPrefix(stateKey, key+"/") {
			return true
		}
	}
	return false
}

func availableShardDescriptors(shards []ShardDescriptor) []ShardDescriptor {
	ordered := cloneShardDescriptors(shards)
	sortShardDescriptors(ordered)
	out := make([]ShardDescriptor, 0, len(ordered))
	for _, shard := range ordered {
		if shard.Available {
			out = append(out, shard)
		}
	}
	return out
}

func newShardRoute(layout ShardLayout, input ShardRoutingInput, mode ShardAssignmentMode, shardID ShardID, placementOverride ShardID, readOnly bool) (ShardRoute, error) {
	route := ShardRoute{
		ZoneID:			layout.ZoneID,
		StateKey:		input.StateKey,
		LayoutEpoch:		layout.LayoutEpoch,
		AssignmentMode:		mode,
		ShardID:		shardID,
		ShardCount:		uint32(len(availableShardDescriptors(layout.ActiveShards))),
		PlacementOverride:	placementOverride,
		ReadOnlyReplicated:	readOnly,
	}
	route.RouteHash = ComputeShardRouteHash(route)
	return route, route.ValidateHash()
}
