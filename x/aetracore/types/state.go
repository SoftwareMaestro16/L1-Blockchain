package types

import (
	"errors"
	"fmt"
	"sort"
)

type ProofRoot struct {
	Height		uint64
	ZoneID		ZoneID
	RootType	RootType
	RootHash	string
	Source		string
	ZoneCount	uint32
}

type FinalityRoots struct {
	GlobalStateRoot		string
	GlobalMessageRoot	string
	ExecutionReceiptRoot	string
}

type RootSnapshot struct {
	Height			uint64
	GlobalStateRoot		ProofRoot
	GlobalMessageRoot	ProofRoot
	ExecutionReceiptRoot	ProofRoot
	ProofRoots		[]ProofRoot
	Finality		FinalityRoots
}

type CoreState struct {
	Params			AetraCoreParams
	Zones			[]ZoneDescriptor
	ZoneDescriptors		[]ZoneDescriptor
	ServiceDescriptors	[]ServiceDescriptor
	ShardLayouts		[]ShardLayout
	RoutingTables		[]RoutingTableCommitment
	ZoneCommitments		[]ZoneCommitment
	GlobalRoots		[]GlobalStateRoot
	RootSnapshots		[]RootSnapshot
	FinalityRecords		[]FinalityRecord
	ExportManifests		[]ExportManifest
}

type AetraCoreState = CoreState

func EmptyState(params ...AetraCoreParams) CoreState {
	stateParams := DefaultParams()
	if len(params) > 0 {
		stateParams = params[0]
	}
	return CoreState{
		Params:			stateParams,
		Zones:			[]ZoneDescriptor{},
		ZoneDescriptors:	[]ZoneDescriptor{},
		ServiceDescriptors:	[]ServiceDescriptor{},
		ShardLayouts:		[]ShardLayout{},
		RoutingTables:		[]RoutingTableCommitment{},
		ZoneCommitments:	[]ZoneCommitment{},
		GlobalRoots:		[]GlobalStateRoot{},
		RootSnapshots:		[]RootSnapshot{},
		FinalityRecords:	[]FinalityRecord{},
		ExportManifests:	[]ExportManifest{},
	}
}

func RegisterZone(state CoreState, descriptor ZoneDescriptor) (CoreState, error) {
	if err := state.Validate(); err != nil {
		return CoreState{}, err
	}
	descriptor = CanonicalZoneDescriptor(descriptor)
	if err := descriptor.Validate(state.Params); err != nil {
		return CoreState{}, err
	}
	if _, found := state.ZoneDescriptorByID(descriptor.ZoneID); found {
		return CoreState{}, fmt.Errorf("aetracore zone descriptor %s already registered", descriptor.ZoneID)
	}
	next := state.Clone()
	next.Zones = append(next.Zones, descriptor)
	next.ZoneDescriptors = append(next.ZoneDescriptors, descriptor)
	sortZoneDescriptors(next.Zones)
	sortZoneDescriptors(next.ZoneDescriptors)
	return next.Export(), next.Validate()
}

func RegisterZoneDescriptor(state CoreState, descriptor ZoneDescriptor) (CoreState, error) {
	return RegisterZone(state, descriptor)
}

func RegisterServiceDescriptor(state CoreState, descriptor ServiceDescriptor) (CoreState, error) {
	if err := state.Validate(); err != nil {
		return CoreState{}, err
	}
	descriptor = CanonicalServiceDescriptor(descriptor)
	if err := descriptor.Validate(); err != nil {
		return CoreState{}, err
	}
	if _, found := state.ZoneDescriptorByID(descriptor.ZoneID); !found {
		return CoreState{}, fmt.Errorf("aetracore service zone %s is not registered", descriptor.ZoneID)
	}
	if _, found := state.ServiceByID(descriptor.ServiceID); found {
		return CoreState{}, fmt.Errorf("aetracore service descriptor %s already registered", descriptor.ServiceID)
	}
	next := state.Clone()
	next.ServiceDescriptors = append(next.ServiceDescriptors, descriptor)
	sortServiceDescriptors(next.ServiceDescriptors)
	return next.Export(), next.Validate()
}

func RegisterShardLayout(state CoreState, layout ShardLayout) (CoreState, error) {
	if err := state.Validate(); err != nil {
		return CoreState{}, err
	}
	layout.ActiveShards = cloneShardDescriptors(layout.ActiveShards)
	sortShardDescriptors(layout.ActiveShards)
	if err := layout.ValidateHash(); err != nil {
		return CoreState{}, err
	}
	descriptor, found := state.ZoneDescriptorByID(layout.ZoneID)
	if !found {
		return CoreState{}, fmt.Errorf("aetracore shard layout zone %s is not registered", layout.ZoneID)
	}
	if !descriptor.Enabled {
		return CoreState{}, fmt.Errorf("aetracore shard layout zone %s is disabled", layout.ZoneID)
	}
	if uint32(len(layout.ActiveShards)) > descriptor.MaxShards {
		return CoreState{}, fmt.Errorf("aetracore shard layout exceeds zone max shards %d", descriptor.MaxShards)
	}
	for _, existing := range state.ShardLayouts {
		if existing.ZoneID == layout.ZoneID && existing.LayoutEpoch == layout.LayoutEpoch {
			return CoreState{}, errors.New("duplicate aetracore shard layout epoch")
		}
	}
	next := state.Clone()
	next.ShardLayouts = append(next.ShardLayouts, layout)
	sortShardLayouts(next.ShardLayouts)
	return next.Export(), next.Validate()
}

func CommitRoutingTable(state CoreState, table RoutingTableCommitment) (CoreState, error) {
	if err := state.Validate(); err != nil {
		return CoreState{}, err
	}
	table.Entries = cloneRoutingZoneEntries(table.Entries)
	sortRoutingZoneEntries(table.Entries)
	if err := table.ValidateHash(); err != nil {
		return CoreState{}, err
	}
	for _, entry := range table.Entries {
		layout, found := state.ShardLayoutByEpoch(entry.ZoneID, entry.LayoutEpoch)
		if !found {
			return CoreState{}, fmt.Errorf("aetracore routing table references missing layout %s/%d", entry.ZoneID, entry.LayoutEpoch)
		}
		if layout.LayoutHash != entry.LayoutHash {
			return CoreState{}, errors.New("aetracore routing table layout hash mismatch")
		}
		if uint32(len(layout.ActiveShards)) != entry.ActiveShards {
			return CoreState{}, errors.New("aetracore routing table active shard count mismatch")
		}
		descriptor, found := state.ZoneDescriptorByID(entry.ZoneID)
		if !found {
			return CoreState{}, fmt.Errorf("aetracore routing table zone %s is not registered", entry.ZoneID)
		}
		if !descriptor.Enabled {
			return CoreState{}, fmt.Errorf("aetracore routing table zone %s is disabled", entry.ZoneID)
		}
	}
	for _, existing := range state.RoutingTables {
		if existing.RoutingEpoch == table.RoutingEpoch {
			return CoreState{}, errors.New("duplicate aetracore routing table epoch")
		}
	}
	next := state.Clone()
	next.RoutingTables = append(next.RoutingTables, table)
	sortRoutingTables(next.RoutingTables)
	return next.Export(), next.Validate()
}

func AppendZoneCommitment(state CoreState, commitment ZoneCommitment) (CoreState, error) {
	if err := state.Validate(); err != nil {
		return CoreState{}, err
	}
	if err := commitment.ValidateHash(); err != nil {
		return CoreState{}, err
	}
	descriptor, found := state.ZoneDescriptorByID(ZoneID(commitment.ZoneID))
	if !found {
		return CoreState{}, fmt.Errorf("aetracore zone %s is not registered", commitment.ZoneID)
	}
	if !descriptor.Enabled {
		return CoreState{}, fmt.Errorf("aetracore zone %s is disabled", commitment.ZoneID)
	}
	for _, existing := range state.ZoneCommitments {
		if existing.Height == commitment.Height && existing.ZoneID == commitment.ZoneID {
			return CoreState{}, errors.New("duplicate aetracore zone commitment height")
		}
	}
	next := state.Clone()
	next.ZoneCommitments = append(next.ZoneCommitments, commitment)
	sortZoneCommitments(next.ZoneCommitments)
	return next.Export(), next.Validate()
}

func CommitBlockRoots(state CoreState, height uint64) (CoreState, RootSnapshot, error) {
	if height == 0 {
		return CoreState{}, RootSnapshot{}, errors.New("aetracore root snapshot height must be positive")
	}
	if err := state.Validate(); err != nil {
		return CoreState{}, RootSnapshot{}, err
	}
	zonesRoot, err := ComputeZoneCommitmentsRoot(height, state.CommitmentsAtHeight(height))
	if err != nil {
		return CoreState{}, RootSnapshot{}, err
	}
	messageRoot := hashParts("aetra-aek-global-message-root-v1", fmt.Sprint(height), zonesRoot)
	receiptRoot := hashParts("aetra-aek-execution-receipt-root-v1", fmt.Sprint(height), zonesRoot)
	contributions := RootContributions{
		IdentityRoot:	EmptyRootHash,
		StorageRoot:	EmptyRootHash,
		MessageRoot:	messageRoot,
		ReceiptsRoot:	receiptRoot,
		RoutingRoot:	EmptyRootHash,
		PaymentsRoot:	EmptyRootHash,
		ContractsRoot:	EmptyRootHash,
		VMRoot:		EmptyRootHash,
		ParamsHash:	ComputeAetraCoreParamsHash(state.Params),
	}
	return CommitBlockRootsWithContributions(state, height, contributions)
}

func CommitBlockRootsWithContributions(state CoreState, height uint64, contributions RootContributions) (CoreState, RootSnapshot, error) {
	if height == 0 {
		return CoreState{}, RootSnapshot{}, errors.New("aetracore root snapshot height must be positive")
	}
	if err := state.Validate(); err != nil {
		return CoreState{}, RootSnapshot{}, err
	}
	if err := contributions.Validate(); err != nil {
		return CoreState{}, RootSnapshot{}, err
	}
	for _, snapshot := range state.RootSnapshots {
		if snapshot.Height == height {
			return CoreState{}, RootSnapshot{}, errors.New("duplicate aetracore root snapshot height")
		}
	}
	commitments := state.CommitmentsAtHeight(height)
	if _, err := ComputeZoneCommitmentsRoot(height, commitments); err != nil {
		return CoreState{}, RootSnapshot{}, err
	}
	globalRoot, err := BuildGlobalStateRoot(height, state, contributions)
	if err != nil {
		return CoreState{}, RootSnapshot{}, err
	}
	globalRootHash := globalRoot.GlobalRoot
	proofRoots := make([]ProofRoot, len(commitments))
	for i, commitment := range commitments {
		proofRoots[i] = ProofRoot{
			Height:		height,
			ZoneID:		ZoneID(commitment.ZoneID),
			RootType:	ZoneCommitmentsRoot,
			RootHash:	commitment.CommitmentHash,
			Source:		"aetracore.zone_commitments",
		}
	}
	proofRoots = append(proofRoots, state.ShardLayoutProofRootsAtHeight(height)...)
	if table, found := state.LatestRoutingTableAtHeight(height); found {
		proofRoots = append(proofRoots, ProofRoot{
			Height:		height,
			RootType:	RoutingTableRootType,
			RootHash:	table.TableHash,
			Source:		"aetracore.routing_table",
			ZoneCount:	uint32(len(table.Entries)),
		})
	}
	snapshot := RootSnapshot{
		Height:	height,
		GlobalStateRoot: ProofRoot{
			Height:		height,
			RootType:	DefaultProofRootType,
			RootHash:	globalRootHash,
			ZoneCount:	uint32(len(commitments)),
		},
		GlobalMessageRoot: ProofRoot{
			Height:		height,
			RootType:	MessageProofRootType,
			RootHash:	contributions.MessageRoot,
		},
		ExecutionReceiptRoot: ProofRoot{
			Height:		height,
			RootType:	ReceiptProofRootType,
			RootHash:	contributions.ReceiptsRoot,
		},
		ProofRoots:	proofRoots,
		Finality: FinalityRoots{
			GlobalStateRoot:	globalRootHash,
			GlobalMessageRoot:	contributions.MessageRoot,
			ExecutionReceiptRoot:	contributions.ReceiptsRoot,
		},
	}
	if err := snapshot.Validate(); err != nil {
		return CoreState{}, RootSnapshot{}, err
	}
	next := state.Clone()
	next.GlobalRoots = append(next.GlobalRoots, globalRoot)
	next.RootSnapshots = append(next.RootSnapshots, snapshot)
	sortGlobalRoots(next.GlobalRoots)
	sortRootSnapshots(next.RootSnapshots)
	return next.Export(), snapshot, next.Validate()
}

func AppendGlobalRoot(state CoreState, root GlobalStateRoot) (CoreState, error) {
	if err := state.Validate(); err != nil {
		return CoreState{}, err
	}
	if err := root.ValidateHash(); err != nil {
		return CoreState{}, err
	}
	for _, existing := range state.GlobalRoots {
		if existing.Height == root.Height {
			return CoreState{}, errors.New("duplicate aetracore global root height")
		}
	}
	next := state.Clone()
	next.GlobalRoots = append(next.GlobalRoots, root)
	sortGlobalRoots(next.GlobalRoots)
	return next.Export(), next.Validate()
}

func AddExportManifest(state CoreState, manifest ExportManifest) (CoreState, error) {
	if err := state.Validate(); err != nil {
		return CoreState{}, err
	}
	if err := manifest.ValidateHash(); err != nil {
		return CoreState{}, err
	}
	root, found := state.GlobalRootByHeight(manifest.Height)
	if !found {
		return CoreState{}, fmt.Errorf("aetracore global root for height %d is not found", manifest.Height)
	}
	if manifest.GlobalRoot != root.GlobalRoot {
		return CoreState{}, errors.New("aetracore export manifest global root mismatch")
	}
	if err := ValidateExportImportRootChecks(state, manifest); err != nil {
		return CoreState{}, err
	}
	for _, existing := range state.ExportManifests {
		if existing.Height == manifest.Height {
			return CoreState{}, errors.New("duplicate aetracore export manifest height")
		}
	}
	next := state.Clone()
	next.ExportManifests = append(next.ExportManifests, manifest)
	sortExportManifests(next.ExportManifests)
	return next.Export(), next.Validate()
}

func AppendFinalityRecord(state CoreState, record FinalityRecord) (CoreState, error) {
	if err := state.Validate(); err != nil {
		return CoreState{}, err
	}
	if err := record.ValidateHash(); err != nil {
		return CoreState{}, err
	}
	snapshot, found := state.RootSnapshotAtHeight(record.Height)
	if !found {
		return CoreState{}, fmt.Errorf("aetracore finality record missing root snapshot at height %d", record.Height)
	}
	if record.GlobalStateRoot != snapshot.Finality.GlobalStateRoot {
		return CoreState{}, errors.New("aetracore finality record global root mismatch")
	}
	if record.GlobalMessageRoot != snapshot.Finality.GlobalMessageRoot {
		return CoreState{}, errors.New("aetracore finality record message root mismatch")
	}
	if record.ExecutionReceiptRoot != snapshot.Finality.ExecutionReceiptRoot {
		return CoreState{}, errors.New("aetracore finality record receipt root mismatch")
	}
	for _, existing := range state.FinalityRecords {
		if existing.Height == record.Height {
			return CoreState{}, errors.New("duplicate aetracore finality record height")
		}
	}
	next := state.Clone()
	next.FinalityRecords = append(next.FinalityRecords, record)
	sortFinalityRecords(next.FinalityRecords)
	return next.Export(), next.Validate()
}

func ImportState(state CoreState) (CoreState, error) {
	if err := state.Validate(); err != nil {
		return CoreState{}, err
	}
	return state.Export(), nil
}

func (s CoreState) Export() CoreState {
	out := s.Clone()
	sortZoneDescriptors(out.Zones)
	out.ZoneDescriptors = append([]ZoneDescriptor(nil), out.Zones...)
	out.ServiceDescriptors = cloneServiceDescriptors(out.ServiceDescriptors)
	sortServiceDescriptors(out.ServiceDescriptors)
	sortShardLayouts(out.ShardLayouts)
	sortRoutingTables(out.RoutingTables)
	sortZoneCommitments(out.ZoneCommitments)
	sortGlobalRoots(out.GlobalRoots)
	sortRootSnapshots(out.RootSnapshots)
	sortFinalityRecords(out.FinalityRecords)
	sortExportManifests(out.ExportManifests)
	return out
}

func (s CoreState) Clone() CoreState {
	zones := canonicalZones(s)
	out := CoreState{
		Params:			s.Params,
		Zones:			make([]ZoneDescriptor, len(zones)),
		ZoneDescriptors:	make([]ZoneDescriptor, len(zones)),
		ServiceDescriptors:	cloneServiceDescriptors(s.ServiceDescriptors),
		ShardLayouts:		cloneShardLayouts(s.ShardLayouts),
		RoutingTables:		cloneRoutingTables(s.RoutingTables),
		ZoneCommitments:	append([]ZoneCommitment(nil), s.ZoneCommitments...),
		GlobalRoots:		append([]GlobalStateRoot(nil), s.GlobalRoots...),
		RootSnapshots:		cloneRootSnapshots(s.RootSnapshots),
		FinalityRecords:	append([]FinalityRecord(nil), s.FinalityRecords...),
		ExportManifests:	append([]ExportManifest(nil), s.ExportManifests...),
	}
	for i, descriptor := range zones {
		out.Zones[i] = CanonicalZoneDescriptor(descriptor)
		out.ZoneDescriptors[i] = CanonicalZoneDescriptor(descriptor)
	}
	return out
}

func (s CoreState) Validate() error {
	if err := s.Params.Validate(); err != nil {
		return err
	}
	zones := canonicalZones(s)
	if err := validateZoneDescriptors(zones, s.Params); err != nil {
		return err
	}
	registeredZones := make(map[ZoneID]ZoneDescriptor, len(zones))
	for _, descriptor := range zones {
		registeredZones[descriptor.ZoneID] = descriptor
	}
	if err := validateServiceDescriptors(s.ServiceDescriptors, registeredZones); err != nil {
		return err
	}
	if err := validateShardLayouts(s.ShardLayouts, registeredZones); err != nil {
		return err
	}
	if err := validateRoutingTables(s.RoutingTables, s.ShardLayouts, registeredZones); err != nil {
		return err
	}
	if err := validateZoneCommitments(s.ZoneCommitments, registeredZones); err != nil {
		return err
	}
	if err := validateGlobalRoots(s.GlobalRoots); err != nil {
		return err
	}
	if err := validateRootSnapshots(s.RootSnapshots); err != nil {
		return err
	}
	if err := validateFinalityRecords(s.FinalityRecords, s.RootSnapshots); err != nil {
		return err
	}
	return validateExportManifests(s.ExportManifests, s.GlobalRoots)
}

func (s CoreState) ZoneDescriptorByID(id ZoneID) (ZoneDescriptor, bool) {
	for _, descriptor := range canonicalZones(s) {
		if descriptor.ZoneID == id {
			return CanonicalZoneDescriptor(descriptor), true
		}
	}
	return ZoneDescriptor{}, false
}

func (s CoreState) ShardLayoutByEpoch(zoneID ZoneID, layoutEpoch uint64) (ShardLayout, bool) {
	for _, layout := range s.ShardLayouts {
		if layout.ZoneID == zoneID && layout.LayoutEpoch == layoutEpoch {
			out := layout
			out.ActiveShards = cloneShardDescriptors(layout.ActiveShards)
			return out, true
		}
	}
	return ShardLayout{}, false
}

func (s CoreState) LatestShardLayout(zoneID ZoneID, height uint64) (ShardLayout, bool) {
	var latest ShardLayout
	found := false
	for _, layout := range s.ShardLayouts {
		if layout.ZoneID != zoneID || layout.ActivationHeight > height {
			continue
		}
		if !found || layout.LayoutEpoch > latest.LayoutEpoch {
			latest = layout
			found = true
		}
	}
	if !found {
		return ShardLayout{}, false
	}
	latest.ActiveShards = cloneShardDescriptors(latest.ActiveShards)
	return latest, true
}

func (s CoreState) RoutingTableByEpoch(epoch uint64) (RoutingTableCommitment, bool) {
	for _, table := range s.RoutingTables {
		if table.RoutingEpoch == epoch {
			out := table
			out.Entries = cloneRoutingZoneEntries(table.Entries)
			return out, true
		}
	}
	return RoutingTableCommitment{}, false
}

func (s CoreState) LatestRoutingTableAtHeight(height uint64) (RoutingTableCommitment, bool) {
	var latest RoutingTableCommitment
	found := false
	for _, table := range s.RoutingTables {
		if table.Height > height {
			continue
		}
		if !found || table.Height > latest.Height || table.Height == latest.Height && table.RoutingEpoch > latest.RoutingEpoch {
			latest = table
			found = true
		}
	}
	if !found {
		return RoutingTableCommitment{}, false
	}
	latest.Entries = cloneRoutingZoneEntries(latest.Entries)
	return latest, true
}

func (s CoreState) ShardLayoutProofRootsAtHeight(height uint64) []ProofRoot {
	zones := make([]ZoneID, 0)
	seen := make(map[ZoneID]struct{})
	for _, layout := range s.ShardLayouts {
		if layout.ActivationHeight > height {
			continue
		}
		if _, found := seen[layout.ZoneID]; found {
			continue
		}
		seen[layout.ZoneID] = struct{}{}
		zones = append(zones, layout.ZoneID)
	}
	sort.SliceStable(zones, func(i, j int) bool {
		return zones[i] < zones[j]
	})
	roots := make([]ProofRoot, 0, len(zones))
	for _, zoneID := range zones {
		layout, found := s.LatestShardLayout(zoneID, height)
		if !found {
			continue
		}
		roots = append(roots, ProofRoot{
			Height:		height,
			ZoneID:		zoneID,
			RootType:	ShardLayoutRootType,
			RootHash:	layout.LayoutHash,
			Source:		"aetracore.shard_layout",
			ZoneCount:	uint32(len(layout.ActiveShards)),
		})
	}
	return roots
}

func (s CoreState) ServiceByID(id string) (ServiceDescriptor, bool) {
	for _, descriptor := range s.ServiceDescriptors {
		if descriptor.ServiceID == id {
			return cloneServiceDescriptor(descriptor), true
		}
	}
	return ServiceDescriptor{}, false
}

func (s CoreState) GlobalRootByHeight(height uint64) (GlobalStateRoot, bool) {
	for _, root := range s.GlobalRoots {
		if root.Height == height {
			return root, true
		}
	}
	return GlobalStateRoot{}, false
}

func (s CoreState) LatestRootSnapshot() (RootSnapshot, bool) {
	if len(s.RootSnapshots) == 0 {
		return RootSnapshot{}, false
	}
	return s.RootSnapshots[len(s.RootSnapshots)-1], true
}

func (s CoreState) FinalityRecordAtHeight(height uint64) (FinalityRecord, bool) {
	for _, record := range s.FinalityRecords {
		if record.Height == height {
			return record, true
		}
	}
	return FinalityRecord{}, false
}

func (s CoreState) CommitmentsAtHeight(height uint64) []ZoneCommitment {
	out := make([]ZoneCommitment, 0)
	for _, commitment := range s.ZoneCommitments {
		if commitment.Height == height {
			out = append(out, commitment)
		}
	}
	sortZoneCommitments(out)
	return out
}

func (r ProofRoot) Validate() error {
	if r.Height == 0 {
		return errors.New("aetracore proof root height must be positive")
	}
	if err := validateToken("aetracore proof root type", string(r.RootType), MaxScopeLength); err != nil {
		return err
	}
	if err := ValidateHash("aetracore proof root hash", r.RootHash); err != nil {
		return err
	}
	if r.ZoneID != "" {
		if err := ValidateZoneID(r.ZoneID); err != nil {
			return err
		}
	}
	if r.Source != "" {
		return validateToken("aetracore proof root source", r.Source, MaxScopeLength)
	}
	return nil
}

func (s RootSnapshot) Validate() error {
	if s.Height == 0 {
		return errors.New("aetracore root snapshot height must be positive")
	}
	if err := s.GlobalStateRoot.Validate(); err != nil {
		return err
	}
	if err := s.GlobalMessageRoot.Validate(); err != nil {
		return err
	}
	if err := s.ExecutionReceiptRoot.Validate(); err != nil {
		return err
	}
	if s.Finality.GlobalStateRoot != s.GlobalStateRoot.RootHash {
		return errors.New("aetracore finality global state root mismatch")
	}
	if s.Finality.GlobalMessageRoot != s.GlobalMessageRoot.RootHash {
		return errors.New("aetracore finality global message root mismatch")
	}
	if s.Finality.ExecutionReceiptRoot != s.ExecutionReceiptRoot.RootHash {
		return errors.New("aetracore finality receipt root mismatch")
	}
	for _, root := range s.ProofRoots {
		if err := root.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func canonicalZones(s CoreState) []ZoneDescriptor {
	if len(s.Zones) > 0 {
		return append([]ZoneDescriptor(nil), s.Zones...)
	}
	return append([]ZoneDescriptor(nil), s.ZoneDescriptors...)
}

func validateZoneDescriptors(descriptors []ZoneDescriptor, params AetraCoreParams) error {
	var previous ZoneID
	seen := make(map[ZoneID]struct{}, len(descriptors))
	for i, descriptor := range descriptors {
		descriptor = CanonicalZoneDescriptor(descriptor)
		if err := descriptor.Validate(params); err != nil {
			return err
		}
		if _, found := seen[descriptor.ZoneID]; found {
			return fmt.Errorf("duplicate aetracore zone descriptor %s", descriptor.ZoneID)
		}
		seen[descriptor.ZoneID] = struct{}{}
		if i > 0 && previous >= descriptor.ZoneID {
			return errors.New("aetracore zone descriptors must be sorted canonically by zone id")
		}
		previous = descriptor.ZoneID
	}
	return nil
}

func validateServiceDescriptors(descriptors []ServiceDescriptor, registeredZones map[ZoneID]ZoneDescriptor) error {
	var previous string
	seen := make(map[string]struct{}, len(descriptors))
	for i, descriptor := range descriptors {
		descriptor = CanonicalServiceDescriptor(descriptor)
		if err := descriptor.Validate(); err != nil {
			return err
		}
		if _, found := registeredZones[descriptor.ZoneID]; !found {
			return fmt.Errorf("aetracore service zone %s is not registered", descriptor.ZoneID)
		}
		if _, found := seen[descriptor.ServiceID]; found {
			return fmt.Errorf("duplicate aetracore service descriptor %s", descriptor.ServiceID)
		}
		seen[descriptor.ServiceID] = struct{}{}
		if i > 0 && previous >= descriptor.ServiceID {
			return errors.New("aetracore service descriptors must be sorted canonically by service id")
		}
		previous = descriptor.ServiceID
	}
	return nil
}

func validateShardLayouts(layouts []ShardLayout, registeredZones map[ZoneID]ZoneDescriptor) error {
	var previous ShardLayout
	seen := make(map[string]struct{}, len(layouts))
	for i, layout := range layouts {
		if err := layout.ValidateHash(); err != nil {
			return err
		}
		descriptor, found := registeredZones[layout.ZoneID]
		if !found {
			return fmt.Errorf("aetracore shard layout zone %s is not registered", layout.ZoneID)
		}
		if !descriptor.Enabled {
			return fmt.Errorf("aetracore shard layout zone %s is disabled", layout.ZoneID)
		}
		if uint32(len(layout.ActiveShards)) > descriptor.MaxShards {
			return fmt.Errorf("aetracore shard layout exceeds zone max shards %d", descriptor.MaxShards)
		}
		key := fmt.Sprintf("%s/%020d", layout.ZoneID, layout.LayoutEpoch)
		if _, found := seen[key]; found {
			return errors.New("duplicate aetracore shard layout epoch")
		}
		seen[key] = struct{}{}
		if i > 0 && compareShardLayouts(previous, layout) >= 0 {
			return errors.New("aetracore shard layouts must be sorted canonically")
		}
		previous = layout
	}
	return nil
}

func validateRoutingTables(tables []RoutingTableCommitment, layouts []ShardLayout, registeredZones map[ZoneID]ZoneDescriptor) error {
	var previous uint64
	seen := make(map[uint64]struct{}, len(tables))
	for i, table := range tables {
		if err := table.ValidateHash(); err != nil {
			return err
		}
		if _, found := seen[table.RoutingEpoch]; found {
			return errors.New("duplicate aetracore routing table epoch")
		}
		seen[table.RoutingEpoch] = struct{}{}
		if i > 0 && previous >= table.RoutingEpoch {
			return errors.New("aetracore routing tables must be sorted canonically by epoch")
		}
		for _, entry := range table.Entries {
			descriptor, found := registeredZones[entry.ZoneID]
			if !found {
				return fmt.Errorf("aetracore routing table zone %s is not registered", entry.ZoneID)
			}
			if !descriptor.Enabled {
				return fmt.Errorf("aetracore routing table zone %s is disabled", entry.ZoneID)
			}
			layout, found := shardLayoutByEpoch(layouts, entry.ZoneID, entry.LayoutEpoch)
			if !found {
				return fmt.Errorf("aetracore routing table references missing layout %s/%d", entry.ZoneID, entry.LayoutEpoch)
			}
			if layout.LayoutHash != entry.LayoutHash {
				return errors.New("aetracore routing table layout hash mismatch")
			}
			if uint32(len(layout.ActiveShards)) != entry.ActiveShards {
				return errors.New("aetracore routing table active shard count mismatch")
			}
		}
		previous = table.RoutingEpoch
	}
	return nil
}

func validateZoneCommitments(commitments []ZoneCommitment, registeredZones map[ZoneID]ZoneDescriptor) error {
	var previous ZoneCommitment
	seen := make(map[string]struct{}, len(commitments))
	for i, commitment := range commitments {
		if err := commitment.ValidateHash(); err != nil {
			return err
		}
		descriptor, found := registeredZones[ZoneID(commitment.ZoneID)]
		if !found {
			return fmt.Errorf("aetracore zone %s is not registered", commitment.ZoneID)
		}
		if !descriptor.Enabled {
			return fmt.Errorf("aetracore zone %s is disabled", commitment.ZoneID)
		}
		key := fmt.Sprintf("%020d/%s", commitment.Height, commitment.ZoneID)
		if _, found := seen[key]; found {
			return errors.New("duplicate aetracore zone commitment height")
		}
		seen[key] = struct{}{}
		if i > 0 && compareZoneCommitments(previous, commitment) >= 0 {
			return errors.New("aetracore zone commitments must be sorted canonically")
		}
		previous = commitment
	}
	return nil
}

func compareShardLayouts(left, right ShardLayout) int {
	if left.ZoneID < right.ZoneID {
		return -1
	}
	if left.ZoneID > right.ZoneID {
		return 1
	}
	if left.LayoutEpoch < right.LayoutEpoch {
		return -1
	}
	if left.LayoutEpoch > right.LayoutEpoch {
		return 1
	}
	return 0
}

func shardLayoutByEpoch(layouts []ShardLayout, zoneID ZoneID, layoutEpoch uint64) (ShardLayout, bool) {
	for _, layout := range layouts {
		if layout.ZoneID == zoneID && layout.LayoutEpoch == layoutEpoch {
			return layout, true
		}
	}
	return ShardLayout{}, false
}

func validateGlobalRoots(roots []GlobalStateRoot) error {
	var previous uint64
	seen := make(map[uint64]struct{}, len(roots))
	for i, root := range roots {
		if err := root.ValidateHash(); err != nil {
			return err
		}
		if _, found := seen[root.Height]; found {
			return errors.New("duplicate aetracore global root height")
		}
		seen[root.Height] = struct{}{}
		if i > 0 && previous >= root.Height {
			return errors.New("aetracore global roots must be sorted canonically by height")
		}
		previous = root.Height
	}
	return nil
}

func validateRootSnapshots(snapshots []RootSnapshot) error {
	var previous uint64
	seen := make(map[uint64]struct{}, len(snapshots))
	for i, snapshot := range snapshots {
		if err := snapshot.Validate(); err != nil {
			return err
		}
		if _, found := seen[snapshot.Height]; found {
			return errors.New("duplicate aetracore root snapshot height")
		}
		seen[snapshot.Height] = struct{}{}
		if i > 0 && previous >= snapshot.Height {
			return errors.New("aetracore root snapshots must be sorted canonically by height")
		}
		previous = snapshot.Height
	}
	return nil
}

func validateFinalityRecords(records []FinalityRecord, snapshots []RootSnapshot) error {
	var previous uint64
	seen := make(map[uint64]struct{}, len(records))
	for i, record := range records {
		if err := record.ValidateHash(); err != nil {
			return err
		}
		if _, found := seen[record.Height]; found {
			return errors.New("duplicate aetracore finality record height")
		}
		seen[record.Height] = struct{}{}
		if i > 0 && previous >= record.Height {
			return errors.New("aetracore finality records must be sorted canonically by height")
		}
		snapshot, found := rootSnapshotByHeight(snapshots, record.Height)
		if !found {
			return fmt.Errorf("aetracore finality record missing root snapshot at height %d", record.Height)
		}
		if record.GlobalStateRoot != snapshot.Finality.GlobalStateRoot {
			return fmt.Errorf("aetracore finality record global root mismatch at height %d", record.Height)
		}
		if record.GlobalMessageRoot != snapshot.Finality.GlobalMessageRoot {
			return fmt.Errorf("aetracore finality record message root mismatch at height %d", record.Height)
		}
		if record.ExecutionReceiptRoot != snapshot.Finality.ExecutionReceiptRoot {
			return fmt.Errorf("aetracore finality record receipt root mismatch at height %d", record.Height)
		}
		previous = record.Height
	}
	return nil
}

func validateExportManifests(manifests []ExportManifest, roots []GlobalStateRoot) error {
	var previous uint64
	seen := make(map[uint64]struct{}, len(manifests))
	for i, manifest := range manifests {
		if err := manifest.ValidateHash(); err != nil {
			return err
		}
		if _, found := seen[manifest.Height]; found {
			return errors.New("duplicate aetracore export manifest height")
		}
		seen[manifest.Height] = struct{}{}
		if i > 0 && previous >= manifest.Height {
			return errors.New("aetracore export manifests must be sorted canonically by height")
		}
		root, found := globalRootByHeight(roots, manifest.Height)
		if !found {
			return fmt.Errorf("aetracore global root for height %d is not found", manifest.Height)
		}
		if manifest.GlobalRoot != root.GlobalRoot {
			return errors.New("aetracore export manifest global root mismatch")
		}
		previous = manifest.Height
	}
	return nil
}

func sortZoneDescriptors(descriptors []ZoneDescriptor) {
	sort.SliceStable(descriptors, func(i, j int) bool {
		return descriptors[i].ZoneID < descriptors[j].ZoneID
	})
}

func sortServiceDescriptors(descriptors []ServiceDescriptor) {
	sort.SliceStable(descriptors, func(i, j int) bool {
		return descriptors[i].ServiceID < descriptors[j].ServiceID
	})
}

func sortZoneCommitments(commitments []ZoneCommitment) {
	sort.SliceStable(commitments, func(i, j int) bool {
		return compareZoneCommitments(commitments[i], commitments[j]) < 0
	})
}

func sortGlobalRoots(roots []GlobalStateRoot) {
	sort.SliceStable(roots, func(i, j int) bool {
		return roots[i].Height < roots[j].Height
	})
}

func sortRootSnapshots(snapshots []RootSnapshot) {
	sort.SliceStable(snapshots, func(i, j int) bool {
		return snapshots[i].Height < snapshots[j].Height
	})
}

func sortFinalityRecords(records []FinalityRecord) {
	sort.SliceStable(records, func(i, j int) bool {
		return records[i].Height < records[j].Height
	})
}

func sortExportManifests(manifests []ExportManifest) {
	sort.SliceStable(manifests, func(i, j int) bool {
		return manifests[i].Height < manifests[j].Height
	})
}

func cloneRootSnapshots(snapshots []RootSnapshot) []RootSnapshot {
	out := make([]RootSnapshot, len(snapshots))
	for i, snapshot := range snapshots {
		out[i] = snapshot
		out[i].ProofRoots = append([]ProofRoot(nil), snapshot.ProofRoots...)
	}
	return out
}

func globalRootByHeight(roots []GlobalStateRoot, height uint64) (GlobalStateRoot, bool) {
	for _, root := range roots {
		if root.Height == height {
			return root, true
		}
	}
	return GlobalStateRoot{}, false
}

func rootSnapshotByHeight(snapshots []RootSnapshot, height uint64) (RootSnapshot, bool) {
	for _, snapshot := range snapshots {
		if snapshot.Height == height {
			out := snapshot
			out.ProofRoots = append([]ProofRoot(nil), snapshot.ProofRoots...)
			return out, true
		}
	}
	return RootSnapshot{}, false
}
