package types

import (
	"errors"
	"fmt"
	"sort"
)

type AetraNextCommitment struct {
	Height			uint64
	ZoneDescriptorRoot	string
	ZoneCommitmentsRoot	string
	ShardLayoutRoot		string
	RoutingTableRoot	string
	UniversalProofRoot	string
	GlobalStateRoot		string
	ParamsHash		string
	ArchitectureHash	string
}

func BuildAetraNextCommitment(height uint64, state CoreState, contributions RootContributions, resolverRoot string) (AetraNextCommitment, error) {
	if height == 0 {
		return AetraNextCommitment{}, errors.New("aetracore next architecture height must be positive")
	}
	if err := state.Params.RequireEnabled(); err != nil {
		return AetraNextCommitment{}, err
	}
	if err := state.Validate(); err != nil {
		return AetraNextCommitment{}, err
	}
	if err := contributions.Validate(); err != nil {
		return AetraNextCommitment{}, err
	}
	if err := ValidateHash("aetracore resolver root", resolverRoot); err != nil {
		return AetraNextCommitment{}, err
	}
	zones := canonicalZones(state)
	if err := ensureEnabledZones(zones); err != nil {
		return AetraNextCommitment{}, err
	}
	if err := ensureEnabledZoneCommitmentsAtHeight(height, zones, state.CommitmentsAtHeight(height)); err != nil {
		return AetraNextCommitment{}, err
	}
	if err := ensureNativeAETResolver(state); err != nil {
		return AetraNextCommitment{}, err
	}
	if err := ValidateAetraNextTopologyState(state, height); err != nil {
		return AetraNextCommitment{}, err
	}

	zoneDescriptorRoot, err := ComputeZoneDescriptorRoot(zones, state.Params)
	if err != nil {
		return AetraNextCommitment{}, err
	}
	zoneCommitmentsRoot, err := ComputeZoneCommitmentsRoot(height, state.CommitmentsAtHeight(height))
	if err != nil {
		return AetraNextCommitment{}, err
	}
	activeLayouts, err := latestExecutionZoneLayouts(height, state, zones)
	if err != nil {
		return AetraNextCommitment{}, err
	}
	shardLayoutRoot, err := ComputeActiveShardLayoutRoot(height, activeLayouts)
	if err != nil {
		return AetraNextCommitment{}, err
	}
	routingTable, found := state.LatestRoutingTableAtHeight(height)
	if !found {
		return AetraNextCommitment{}, errors.New("aetracore next architecture requires committed routing table")
	}
	if err := validateRoutingTableCoversLayouts(routingTable, activeLayouts); err != nil {
		return AetraNextCommitment{}, err
	}
	globalRoot, err := BuildGlobalStateRoot(height, state, contributions)
	if err != nil {
		return AetraNextCommitment{}, err
	}
	proofRegistryRoot, err := ComputeUniversalProofRegistryRoot(height, nextArchitectureProofRoots(
		height,
		globalRoot.GlobalRoot,
		zoneCommitmentsRoot,
		shardLayoutRoot,
		routingTable.TableHash,
		contributions,
		resolverRoot,
	))
	if err != nil {
		return AetraNextCommitment{}, err
	}

	commitment := AetraNextCommitment{
		Height:			height,
		ZoneDescriptorRoot:	zoneDescriptorRoot,
		ZoneCommitmentsRoot:	zoneCommitmentsRoot,
		ShardLayoutRoot:	shardLayoutRoot,
		RoutingTableRoot:	routingTable.TableHash,
		UniversalProofRoot:	proofRegistryRoot,
		GlobalStateRoot:	globalRoot.GlobalRoot,
		ParamsHash:		ComputeAetraCoreParamsHash(state.Params),
	}
	commitment.ArchitectureHash = ComputeAetraNextArchitectureHash(commitment)
	return commitment, commitment.ValidateHash()
}

func (c AetraNextCommitment) ValidateFormat() error {
	if c.Height == 0 {
		return errors.New("aetracore next architecture height must be positive")
	}
	for _, field := range []struct {
		name	string
		value	string
	}{
		{"aetracore next zone descriptor root", c.ZoneDescriptorRoot},
		{"aetracore next zone commitments root", c.ZoneCommitmentsRoot},
		{"aetracore next shard layout root", c.ShardLayoutRoot},
		{"aetracore next routing table root", c.RoutingTableRoot},
		{"aetracore next universal proof root", c.UniversalProofRoot},
		{"aetracore next global state root", c.GlobalStateRoot},
		{"aetracore next params hash", c.ParamsHash},
	} {
		if err := ValidateHash(field.name, field.value); err != nil {
			return err
		}
	}
	if c.ArchitectureHash != "" {
		return ValidateHash("aetracore next architecture hash", c.ArchitectureHash)
	}
	return nil
}

func (c AetraNextCommitment) ValidateHash() error {
	if err := c.ValidateFormat(); err != nil {
		return err
	}
	expected := ComputeAetraNextArchitectureHash(c)
	if c.ArchitectureHash != expected {
		return fmt.Errorf("aetracore next architecture hash mismatch: expected %s", expected)
	}
	return nil
}

func ComputeAetraNextArchitectureHash(c AetraNextCommitment) string {
	return hashParts(
		"aetra-next-architecture-v1",
		fmt.Sprint(c.Height),
		c.ZoneDescriptorRoot,
		c.ZoneCommitmentsRoot,
		c.ShardLayoutRoot,
		c.RoutingTableRoot,
		c.UniversalProofRoot,
		c.GlobalStateRoot,
		c.ParamsHash,
	)
}

func ComputeActiveShardLayoutRoot(height uint64, layouts []ShardLayout) (string, error) {
	if height == 0 {
		return "", errors.New("aetracore active shard layout root height must be positive")
	}
	ordered := cloneShardLayouts(layouts)
	sortShardLayouts(ordered)
	parts := []string{"aetra-next-active-shard-layout-root-v1", fmt.Sprint(height), fmt.Sprint(len(ordered))}
	var previous ShardLayout
	for i, layout := range ordered {
		if err := layout.ValidateHash(); err != nil {
			return "", err
		}
		if layout.ActivationHeight > height {
			return "", errors.New("aetracore active shard layout activates after root height")
		}
		if i > 0 && compareShardLayouts(previous, layout) >= 0 {
			return "", errors.New("aetracore active shard layouts must be sorted canonically")
		}
		parts = append(parts, layout.LayoutHash)
		previous = layout
	}
	return hashParts(parts...), nil
}

func ComputeUniversalProofRegistryRoot(height uint64, roots []ProofRoot) (string, error) {
	if height == 0 {
		return "", errors.New("aetracore universal proof registry height must be positive")
	}
	ordered := append([]ProofRoot(nil), roots...)
	sortProofRoots(ordered)
	parts := []string{"aetra-next-universal-proof-registry-v1", fmt.Sprint(height), fmt.Sprint(len(ordered))}
	var previous ProofRoot
	for i, root := range ordered {
		if root.Height != height {
			return "", errors.New("aetracore universal proof registry contains different height")
		}
		if err := root.Validate(); err != nil {
			return "", err
		}
		if i > 0 && compareProofRoots(previous, root) >= 0 {
			return "", errors.New("aetracore universal proof roots must be sorted canonically")
		}
		parts = append(parts,
			string(root.RootType),
			string(root.ZoneID),
			root.RootHash,
			root.Source,
			fmt.Sprint(root.ZoneCount),
		)
		previous = root
	}
	return hashParts(parts...), nil
}

func latestExecutionZoneLayouts(height uint64, state CoreState, zones []ZoneDescriptor) ([]ShardLayout, error) {
	layouts := make([]ShardLayout, 0)
	for _, zone := range zones {
		zone = CanonicalZoneDescriptor(zone)
		if !zone.Enabled || zone.ZoneID == ZoneIDAetraCore {
			continue
		}
		layout, found := state.LatestShardLayout(zone.ZoneID, height)
		if !found {
			return nil, fmt.Errorf("aetracore next architecture missing shard layout for zone %s", zone.ZoneID)
		}
		layouts = append(layouts, layout)
	}
	sortShardLayouts(layouts)
	return layouts, nil
}

func validateRoutingTableCoversLayouts(table RoutingTableCommitment, layouts []ShardLayout) error {
	if err := table.ValidateHash(); err != nil {
		return err
	}
	entries := make(map[ZoneID]RoutingZoneEntry, len(table.Entries))
	for _, entry := range table.Entries {
		entries[entry.ZoneID] = entry
	}
	for _, layout := range layouts {
		entry, found := entries[layout.ZoneID]
		if !found {
			return fmt.Errorf("aetracore routing table missing zone %s", layout.ZoneID)
		}
		if entry.LayoutEpoch != layout.LayoutEpoch {
			return fmt.Errorf("aetracore routing table layout epoch mismatch for zone %s", layout.ZoneID)
		}
		if entry.LayoutHash != layout.LayoutHash {
			return fmt.Errorf("aetracore routing table layout hash mismatch for zone %s", layout.ZoneID)
		}
		if entry.ActiveShards != uint32(len(layout.ActiveShards)) {
			return fmt.Errorf("aetracore routing table active shard mismatch for zone %s", layout.ZoneID)
		}
	}
	return nil
}

func ensureEnabledZoneCommitmentsAtHeight(height uint64, zones []ZoneDescriptor, commitments []ZoneCommitment) error {
	seen := make(map[ZoneID]struct{}, len(commitments))
	for _, commitment := range commitments {
		if commitment.Height == height {
			seen[commitment.ZoneID] = struct{}{}
		}
	}
	for _, zone := range zones {
		zone = CanonicalZoneDescriptor(zone)
		if !zone.Enabled {
			continue
		}
		if _, found := seen[zone.ZoneID]; !found {
			return fmt.Errorf("aetracore next architecture missing zone commitment for %s", zone.ZoneID)
		}
	}
	return nil
}

func ensureEnabledZones(zones []ZoneDescriptor) error {
	for _, zone := range zones {
		if CanonicalZoneDescriptor(zone).Enabled {
			return nil
		}
	}
	return errors.New("aetracore next architecture requires at least one enabled zone")
}

func ensureNativeAETResolver(state CoreState) error {
	identity, found := state.ZoneDescriptorByID(ZoneIDIdentity)
	if !found || !identity.Enabled {
		return errors.New("aetracore next architecture requires enabled identity zone for .aet resolver")
	}
	for _, service := range state.ServiceDescriptors {
		if service.ZoneID == ZoneIDIdentity &&
			service.Enabled &&
			service.ServiceType == ServiceTypeOnChain &&
			service.Discovery.IdentityName == "identity.aet" {
			return nil
		}
	}
	return errors.New("aetracore next architecture requires native .aet identity resolver service")
}

func nextArchitectureProofRoots(
	height uint64,
	globalStateRoot string,
	zoneCommitmentsRoot string,
	shardLayoutRoot string,
	routingTableRoot string,
	contributions RootContributions,
	resolverRoot string,
) []ProofRoot {
	return []ProofRoot{
		{Height: height, RootType: StateProofRootType, RootHash: globalStateRoot, Source: "aetracore.next.state"},
		{Height: height, RootType: ZoneCommitmentsRoot, RootHash: zoneCommitmentsRoot, Source: "aetracore.next.zones"},
		{Height: height, RootType: ShardLayoutRootType, RootHash: shardLayoutRoot, Source: "aetracore.next.shards"},
		{Height: height, RootType: RoutingTableRootType, RootHash: routingTableRoot, Source: "aetracore.next.routing"},
		{Height: height, RootType: AccountProofRootType, RootHash: contributions.StorageRoot, Source: "aetracore.next.accounts"},
		{Height: height, RootType: MessageProofRootType, RootHash: contributions.MessageRoot, Source: "aetracore.next.messages"},
		{Height: height, RootType: ReceiptProofRootType, RootHash: contributions.ReceiptsRoot, Source: "aetracore.next.receipts"},
		{Height: height, RootType: IdentityProofRootType, RootHash: contributions.IdentityRoot, Source: "aetracore.next.identity"},
		{Height: height, RootType: ResolverProofRootType, RootHash: resolverRoot, Source: "aetracore.next.resolver"},
		{Height: height, RootType: PaymentsProofRootType, RootHash: contributions.PaymentsRoot, Source: "aetracore.next.payments"},
		{Height: height, RootType: VMProofRootType, RootHash: contributions.VMRoot, Source: "aetracore.next.vm"},
	}
}

func sortProofRoots(roots []ProofRoot) {
	sort.SliceStable(roots, func(i, j int) bool {
		return compareProofRoots(roots[i], roots[j]) < 0
	})
}

func compareProofRoots(left, right ProofRoot) int {
	for _, pair := range [][2]string{
		{string(left.RootType), string(right.RootType)},
		{string(left.ZoneID), string(right.ZoneID)},
		{left.Source, right.Source},
		{left.RootHash, right.RootHash},
	} {
		if pair[0] < pair[1] {
			return -1
		}
		if pair[0] > pair[1] {
			return 1
		}
	}
	if left.Height < right.Height {
		return -1
	}
	if left.Height > right.Height {
		return 1
	}
	if left.ZoneCount < right.ZoneCount {
		return -1
	}
	if left.ZoneCount > right.ZoneCount {
		return 1
	}
	return 0
}
