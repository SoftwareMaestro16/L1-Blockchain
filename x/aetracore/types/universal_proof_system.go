package types

import (
	"errors"
	"fmt"
	"sort"
)

const (
	BalanceProofRootType		RootType	= "balance"
	ZoneStateProofRootType		RootType	= "zone_state"
	ShardStateProofRootType		RootType	= "shard_state"
	DomainOwnershipProofRootType	RootType	= "domain_ownership"
	ContractStateProofRootType	RootType	= "contract_state"
	PaymentSettlementProofRootType	RootType	= "payment_settlement"
)

type UniversalProofObjective string

const (
	ProofObjectiveAccountState	UniversalProofObjective	= "account_state"
	ProofObjectiveBalanceState	UniversalProofObjective	= "balance_state"
	ProofObjectiveMessageInclusion	UniversalProofObjective	= "message_inclusion"
	ProofObjectiveMessageReceipt	UniversalProofObjective	= "message_receipt"
	ProofObjectiveZoneStateRoot	UniversalProofObjective	= "zone_state_root"
	ProofObjectiveShardStateRoot	UniversalProofObjective	= "shard_state_root"
	ProofObjectiveDomainOwnership	UniversalProofObjective	= "domain_ownership"
	ProofObjectiveResolverRecords	UniversalProofObjective	= "resolver_records"
	ProofObjectiveContractState	UniversalProofObjective	= "contract_state"
	ProofObjectivePaymentSettlement	UniversalProofObjective	= "payment_settlement"
)

type UniversalProofObjectiveDescriptor struct {
	Objective	UniversalProofObjective
	VerifiedObject	string
	RootType	RootType
	ZoneID		ZoneID
	RequiresShard	bool
}

type UniversalShardRootBranch struct {
	ZoneID		ZoneID
	ShardID		ShardID
	ShardRoot	string
}

type UniversalZoneRootBranch struct {
	ZoneID		ZoneID
	ZoneRoot	string
	ShardRootsRoot	string
	ShardRoots	[]UniversalShardRootBranch
}

type UniversalMessageRootBranch struct {
	ZoneOutboxRoot	string
	ZoneInboxRoot	string
	ReceiptRoot	string
	MessageRoot	string
}

type UniversalRootHierarchy struct {
	Height			uint64
	AetraCoreRoot		string
	GlobalZoneRoot		string
	GlobalMessageRoot	string
	AppHash			string
	Zones			[]UniversalZoneRootBranch
	Messages		UniversalMessageRootBranch
}

func SupportedUniversalProofObjectives() []UniversalProofObjectiveDescriptor {
	return []UniversalProofObjectiveDescriptor{
		{Objective: ProofObjectiveAccountState, VerifiedObject: "account metadata, signer state, nonce, permissions, and ownership bindings", RootType: AccountProofRootType},
		{Objective: ProofObjectiveBalanceState, VerifiedObject: "native and token balances under Financial Zone roots", RootType: BalanceProofRootType, ZoneID: ZoneIDFinancial, RequiresShard: true},
		{Objective: ProofObjectiveMessageInclusion, VerifiedObject: "source outbox or global message root inclusion for an AetherMessage", RootType: MessageProofRootType},
		{Objective: ProofObjectiveMessageReceipt, VerifiedObject: "destination receipt status, gas, fees, output messages, and write summary", RootType: ReceiptProofRootType},
		{Objective: ProofObjectiveZoneStateRoot, VerifiedObject: "zone commitment included in the global zone root", RootType: ZoneStateProofRootType},
		{Objective: ProofObjectiveShardStateRoot, VerifiedObject: "shard root included in a zone shard_roots_root", RootType: ShardStateProofRootType, RequiresShard: true},
		{Objective: ProofObjectiveDomainOwnership, VerifiedObject: ".aet domain ownership, NFT binding, delegation, or auction state", RootType: DomainOwnershipProofRootType, ZoneID: ZoneIDIdentity, RequiresShard: true},
		{Objective: ProofObjectiveResolverRecords, VerifiedObject: ".aet resolver values and reverse lookup records", RootType: ResolverProofRootType, ZoneID: ZoneIDIdentity, RequiresShard: true},
		{Objective: ProofObjectiveContractState, VerifiedObject: "contract code, instance metadata, storage values, ABI, and events", RootType: ContractStateProofRootType, ZoneID: ZoneIDContract, RequiresShard: true},
		{Objective: ProofObjectivePaymentSettlement, VerifiedObject: "payment channel, escrow, collateral, route, and settlement status", RootType: PaymentSettlementProofRootType, ZoneID: ZoneIDFinancial, RequiresShard: true},
	}
}

func UniversalProofObjectiveByID(objective UniversalProofObjective) (UniversalProofObjectiveDescriptor, bool) {
	for _, descriptor := range SupportedUniversalProofObjectives() {
		if descriptor.Objective == objective {
			return descriptor, true
		}
	}
	return UniversalProofObjectiveDescriptor{}, false
}

func ValidateProofRootForObjective(objective UniversalProofObjective, root ProofRoot) error {
	descriptor, found := UniversalProofObjectiveByID(objective)
	if !found {
		return fmt.Errorf("aetracore universal proof objective %q is not supported", objective)
	}
	if err := root.Validate(); err != nil {
		return err
	}
	if root.RootType != descriptor.RootType {
		return fmt.Errorf("aetracore proof objective %s requires root type %s", objective, descriptor.RootType)
	}
	if descriptor.ZoneID != "" && root.ZoneID != descriptor.ZoneID {
		return fmt.Errorf("aetracore proof objective %s requires zone %s", objective, descriptor.ZoneID)
	}
	return nil
}

func NewUniversalZoneRootBranch(height uint64, zoneID ZoneID, zoneRoot string, shards []UniversalShardRootBranch) (UniversalZoneRootBranch, error) {
	branch := UniversalZoneRootBranch{
		ZoneID:		zoneID,
		ZoneRoot:	zoneRoot,
		ShardRoots:	canonicalUniversalShardRootBranches(shards),
	}
	root, err := ComputeUniversalShardRootsRoot(height, zoneID, branch.ShardRoots)
	if err != nil {
		return UniversalZoneRootBranch{}, err
	}
	branch.ShardRootsRoot = root
	return branch, branch.Validate(height)
}

func NewUniversalMessageRootBranch(height uint64, outboxRoot string, inboxRoot string, receiptRoot string) (UniversalMessageRootBranch, error) {
	branch := UniversalMessageRootBranch{
		ZoneOutboxRoot:	outboxRoot,
		ZoneInboxRoot:	inboxRoot,
		ReceiptRoot:	receiptRoot,
	}
	if err := branch.ValidateFormat(); err != nil {
		return UniversalMessageRootBranch{}, err
	}
	branch.MessageRoot = ComputeUniversalMessageRoot(height, branch)
	return branch, branch.Validate(height)
}

func NewUniversalRootHierarchy(height uint64, aetherCoreRoot string, zones []UniversalZoneRootBranch, messages UniversalMessageRootBranch) (UniversalRootHierarchy, error) {
	hierarchy := UniversalRootHierarchy{
		Height:		height,
		AetraCoreRoot:	aetherCoreRoot,
		Zones:		canonicalUniversalZoneRootBranches(zones),
		Messages:	messages,
	}
	if err := hierarchy.ValidateFormatOnly(); err != nil {
		return UniversalRootHierarchy{}, err
	}
	globalZoneRoot, err := ComputeUniversalGlobalZoneRoot(height, hierarchy.Zones)
	if err != nil {
		return UniversalRootHierarchy{}, err
	}
	hierarchy.GlobalZoneRoot = globalZoneRoot
	hierarchy.GlobalMessageRoot = ComputeUniversalMessageRoot(height, hierarchy.Messages)
	hierarchy.AppHash = ComputeUniversalAppHash(hierarchy)
	return hierarchy, hierarchy.Validate()
}

func (b UniversalShardRootBranch) Validate(zoneID ZoneID) error {
	if b.ZoneID != zoneID {
		return fmt.Errorf("aetracore universal shard branch zone mismatch: expected %s", zoneID)
	}
	if err := ValidateZoneID(b.ZoneID); err != nil {
		return err
	}
	if err := ValidateShardID(b.ShardID); err != nil {
		return err
	}
	return ValidateHash("aetracore universal shard root", b.ShardRoot)
}

func (b UniversalZoneRootBranch) Validate(height uint64) error {
	if err := ValidateZoneID(b.ZoneID); err != nil {
		return err
	}
	if err := ValidateHash("aetracore universal zone root", b.ZoneRoot); err != nil {
		return err
	}
	if err := ValidateHash("aetracore universal shard roots root", b.ShardRootsRoot); err != nil {
		return err
	}
	expected, err := ComputeUniversalShardRootsRoot(height, b.ZoneID, b.ShardRoots)
	if err != nil {
		return err
	}
	if b.ShardRootsRoot != expected {
		return fmt.Errorf("aetracore universal shard roots root mismatch: expected %s", expected)
	}
	return nil
}

func (b UniversalMessageRootBranch) ValidateFormat() error {
	if err := ValidateHash("aetracore universal zone outbox root", b.ZoneOutboxRoot); err != nil {
		return err
	}
	if err := ValidateHash("aetracore universal zone inbox root", b.ZoneInboxRoot); err != nil {
		return err
	}
	if err := ValidateHash("aetracore universal receipt root", b.ReceiptRoot); err != nil {
		return err
	}
	if b.MessageRoot != "" {
		return ValidateHash("aetracore universal message root", b.MessageRoot)
	}
	return nil
}

func (b UniversalMessageRootBranch) Validate(height uint64) error {
	if height == 0 {
		return errors.New("aetracore universal message root height must be positive")
	}
	if err := b.ValidateFormat(); err != nil {
		return err
	}
	expected := ComputeUniversalMessageRoot(height, b)
	if b.MessageRoot != expected {
		return fmt.Errorf("aetracore universal message root mismatch: expected %s", expected)
	}
	return nil
}

func (h UniversalRootHierarchy) ValidateFormatOnly() error {
	if h.Height == 0 {
		return errors.New("aetracore universal root hierarchy height must be positive")
	}
	if err := ValidateHash("aetracore universal aether core root", h.AetraCoreRoot); err != nil {
		return err
	}
	if len(h.Zones) == 0 {
		return errors.New("aetracore universal root hierarchy requires zone roots")
	}
	return h.Messages.Validate(h.Height)
}

func (h UniversalRootHierarchy) Validate() error {
	if err := h.ValidateFormatOnly(); err != nil {
		return err
	}
	expectedZoneRoot, err := ComputeUniversalGlobalZoneRoot(h.Height, h.Zones)
	if err != nil {
		return err
	}
	if h.GlobalZoneRoot != expectedZoneRoot {
		return fmt.Errorf("aetracore universal global zone root mismatch: expected %s", expectedZoneRoot)
	}
	expectedMessageRoot := ComputeUniversalMessageRoot(h.Height, h.Messages)
	if h.GlobalMessageRoot != expectedMessageRoot {
		return fmt.Errorf("aetracore universal global message root mismatch: expected %s", expectedMessageRoot)
	}
	expectedAppHash := ComputeUniversalAppHash(h)
	if h.AppHash != expectedAppHash {
		return fmt.Errorf("aetracore universal app hash mismatch: expected %s", expectedAppHash)
	}
	return nil
}

func (h UniversalRootHierarchy) ValidateRequiredZones(required []ZoneID) error {
	if err := h.Validate(); err != nil {
		return err
	}
	seen := make(map[ZoneID]struct{}, len(h.Zones))
	for _, zone := range h.Zones {
		seen[zone.ZoneID] = struct{}{}
	}
	for _, zoneID := range required {
		if err := ValidateZoneID(zoneID); err != nil {
			return err
		}
		if _, found := seen[zoneID]; !found {
			return fmt.Errorf("aetracore universal root hierarchy missing required zone %s", zoneID)
		}
	}
	return nil
}

func RequiredAetraNextProofZones() []ZoneID {
	return []ZoneID{ZoneIDFinancial, ZoneIDIdentity, ZoneIDApplication, ZoneIDContract}
}

func ComputeUniversalShardRootsRoot(height uint64, zoneID ZoneID, shards []UniversalShardRootBranch) (string, error) {
	if height == 0 {
		return "", errors.New("aetracore universal shard roots height must be positive")
	}
	if err := ValidateZoneID(zoneID); err != nil {
		return "", err
	}
	if len(shards) == 0 {
		return "", errors.New("aetracore universal shard roots require at least one shard")
	}
	ordered := canonicalUniversalShardRootBranches(shards)
	seen := make(map[ShardID]struct{}, len(ordered))
	return hashRoot("aetra-next-universal-shard-roots-v1", func(w byteWriter) {
		writeUint64(w, height)
		writePart(w, string(zoneID))
		writeUint64(w, uint64(len(ordered)))
		for _, shard := range ordered {
			if _, found := seen[shard.ShardID]; found {
				writePart(w, "duplicate")
				continue
			}
			seen[shard.ShardID] = struct{}{}
			if err := shard.Validate(zoneID); err != nil {
				writePart(w, "invalid")
				continue
			}
			writePart(w, string(shard.ShardID))
			writePart(w, shard.ShardRoot)
		}
	}), validateUniversalShardRootBranches(zoneID, ordered)
}

func ComputeUniversalGlobalZoneRoot(height uint64, zones []UniversalZoneRootBranch) (string, error) {
	if height == 0 {
		return "", errors.New("aetracore universal global zone root height must be positive")
	}
	if len(zones) == 0 {
		return "", errors.New("aetracore universal global zone root requires zone roots")
	}
	ordered := canonicalUniversalZoneRootBranches(zones)
	if err := validateUniversalZoneRootBranches(height, ordered); err != nil {
		return "", err
	}
	return hashRoot("aetra-next-universal-global-zone-root-v1", func(w byteWriter) {
		writeUint64(w, height)
		writeUint64(w, uint64(len(ordered)))
		for _, zone := range ordered {
			writePart(w, string(zone.ZoneID))
			writePart(w, zone.ZoneRoot)
			writePart(w, zone.ShardRootsRoot)
		}
	}), nil
}

func ComputeUniversalMessageRoot(height uint64, branch UniversalMessageRootBranch) string {
	return hashRoot("aetra-next-universal-message-root-v1", func(w byteWriter) {
		writeUint64(w, height)
		writePart(w, branch.ZoneOutboxRoot)
		writePart(w, branch.ZoneInboxRoot)
		writePart(w, branch.ReceiptRoot)
	})
}

func ComputeUniversalAppHash(h UniversalRootHierarchy) string {
	return hashRoot("aetra-next-universal-app-hash-v1", func(w byteWriter) {
		writeUint64(w, h.Height)
		writePart(w, h.AetraCoreRoot)
		writePart(w, h.GlobalZoneRoot)
		writePart(w, h.GlobalMessageRoot)
	})
}

func validateUniversalShardRootBranches(zoneID ZoneID, shards []UniversalShardRootBranch) error {
	var previous ShardID
	seen := make(map[ShardID]struct{}, len(shards))
	for i, shard := range shards {
		if err := shard.Validate(zoneID); err != nil {
			return err
		}
		if _, found := seen[shard.ShardID]; found {
			return fmt.Errorf("duplicate aetracore universal shard root %s", shard.ShardID)
		}
		seen[shard.ShardID] = struct{}{}
		if i > 0 && previous >= shard.ShardID {
			return errors.New("aetracore universal shard roots must be sorted canonically")
		}
		previous = shard.ShardID
	}
	return nil
}

func validateUniversalZoneRootBranches(height uint64, zones []UniversalZoneRootBranch) error {
	var previous ZoneID
	seen := make(map[ZoneID]struct{}, len(zones))
	for i, zone := range zones {
		if err := zone.Validate(height); err != nil {
			return err
		}
		if _, found := seen[zone.ZoneID]; found {
			return fmt.Errorf("duplicate aetracore universal zone root %s", zone.ZoneID)
		}
		seen[zone.ZoneID] = struct{}{}
		if i > 0 && previous >= zone.ZoneID {
			return errors.New("aetracore universal zone roots must be sorted canonically")
		}
		previous = zone.ZoneID
	}
	return nil
}

func canonicalUniversalShardRootBranches(shards []UniversalShardRootBranch) []UniversalShardRootBranch {
	out := append([]UniversalShardRootBranch(nil), shards...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].ZoneID == out[j].ZoneID {
			return out[i].ShardID < out[j].ShardID
		}
		return out[i].ZoneID < out[j].ZoneID
	})
	return out
}

func canonicalUniversalZoneRootBranches(zones []UniversalZoneRootBranch) []UniversalZoneRootBranch {
	out := append([]UniversalZoneRootBranch(nil), zones...)
	for i := range out {
		out[i].ShardRoots = canonicalUniversalShardRootBranches(out[i].ShardRoots)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].ZoneID < out[j].ZoneID
	})
	return out
}
