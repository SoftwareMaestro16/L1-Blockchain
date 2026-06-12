package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	ActorStatusActive	= "active"
	ActorStatusFrozen	= "frozen"
	ActorStatusDeleted	= "deleted"
	ActorStatusMigrated	= "migrated"

	RentStatusCurrent	= "current"
	RentStatusDue		= "due"
	RentStatusDelinquent	= "delinquent"

	DeletedValuePolicyReject	= "reject"
	DeletedValuePolicyRedirect	= "redirect"
	DeletedValuePolicyRefund	= "refund"

	MaxCapabilityBytes	= 128
)

type ActorRegistryParams struct {
	MaxActors		uint32
	MaxCapabilities		uint32
	DeletedValuePolicy	string
	DeletedValueRedirect	string
	AllowDeletedValueSend	bool
}

type ActorRegistryState struct {
	Actors		[]ActorRecord
	CodeStore	[]CodeRecord
	Migrations	[]ActorMigrationRecord
}

type CodeRecord struct {
	CodeHash	string
	RegisteredBy	string
	RegisteredAt	uint64
}

type ActorRecord struct {
	ActorID			string
	ContractAddress		string
	Owner			string
	CodeHash		string
	StorageRoot		string
	MailboxRoot		string
	Balance			uint64
	LogicalTime		uint64
	Status			string
	RentStatus		string
	LastActiveHeight	uint64
	Capabilities		[]string
	MigratedFrom		string
	MigratedTo		string
}

type ActorMigrationRecord struct {
	ActorID		string
	FromCodeHash	string
	ToCodeHash	string
	Height		uint64
	LogicalTime	uint64
}

type MsgRegisterActor struct {
	Authority	string
	Owner		string
	CodeHash	string
	Salt		string
	ActorID		string
	ContractAddress	string
	StorageRoot	string
	MailboxRoot	string
	Balance		uint64
	Height		uint64
	Capabilities	[]string
}

type MsgUpdateActorCode struct {
	Authority	string
	ActorID		string
	CodeHash	string
	Height		uint64
	LogicalTime	uint64
}

type MsgFreezeActor struct {
	Authority	string
	ActorID		string
	Height		uint64
}

type MsgUnfreezeActor struct {
	Authority	string
	ActorID		string
	Height		uint64
}

type MsgDeleteActor struct {
	Authority	string
	ActorID		string
	Height		uint64
}

type MsgMigrateActor struct {
	Authority	string
	ActorID		string
	NewCodeHash	string
	NewStorageRoot	string
	NewMailboxRoot	string
	NewActorID	string
	NewAddress	string
	Height		uint64
	LogicalTime	uint64
}

func DefaultActorRegistryParams() ActorRegistryParams {
	return ActorRegistryParams{
		MaxActors:		100_000,
		MaxCapabilities:	64,
		DeletedValuePolicy:	DeletedValuePolicyReject,
	}
}

func EmptyActorRegistryState() ActorRegistryState {
	return ActorRegistryState{Actors: []ActorRecord{}, CodeStore: []CodeRecord{}, Migrations: []ActorMigrationRecord{}}
}

func (p ActorRegistryParams) Validate() error {
	if p.MaxActors == 0 {
		return errors.New("actor registry max actors must be positive")
	}
	if p.MaxCapabilities == 0 {
		return errors.New("actor registry max capabilities must be positive")
	}
	if !IsDeletedValuePolicy(p.DeletedValuePolicy) {
		return errors.New("actor registry deleted value policy is invalid")
	}
	if p.DeletedValuePolicy == DeletedValuePolicyRedirect && p.DeletedValueRedirect == "" {
		return errors.New("actor registry deleted redirect target is required")
	}
	return nil
}

func (s ActorRegistryState) Export() ActorRegistryState {
	out := ActorRegistryState{
		Actors:		cloneActors(s.Actors),
		CodeStore:	cloneCodeStore(s.CodeStore),
		Migrations:	cloneMigrations(s.Migrations),
	}
	SortActors(out.Actors)
	SortCodeStore(out.CodeStore)
	SortMigrations(out.Migrations)
	return out
}

func (s ActorRegistryState) Validate(params ActorRegistryParams) error {
	if err := params.Validate(); err != nil {
		return err
	}
	s = s.Export()
	if uint32(len(s.Actors)) > params.MaxActors {
		return errors.New("actor registry actor count exceeds limit")
	}
	codeHashes := map[string]struct{}{}
	for _, code := range s.CodeStore {
		if err := code.Validate(); err != nil {
			return err
		}
		if _, found := codeHashes[code.CodeHash]; found {
			return fmt.Errorf("duplicate actor registry code hash %q", code.CodeHash)
		}
		codeHashes[code.CodeHash] = struct{}{}
	}
	seenActorIDs := map[string]struct{}{}
	seenAddresses := map[string]struct{}{}
	for _, actor := range s.Actors {
		if err := actor.Validate(params); err != nil {
			return err
		}
		if _, found := codeHashes[actor.CodeHash]; !found {
			return fmt.Errorf("actor registry code hash %q is missing from AVM code store", actor.CodeHash)
		}
		if _, found := seenActorIDs[actor.ActorID]; found {
			return fmt.Errorf("duplicate actor id %q", actor.ActorID)
		}
		if _, found := seenAddresses[actor.ContractAddress]; found {
			return fmt.Errorf("duplicate actor contract address %q", actor.ContractAddress)
		}
		seenActorIDs[actor.ActorID] = struct{}{}
		seenAddresses[actor.ContractAddress] = struct{}{}
	}
	for _, migration := range s.Migrations {
		if err := migration.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (c CodeRecord) Normalize() CodeRecord {
	c.CodeHash = strings.TrimSpace(c.CodeHash)
	c.RegisteredBy = strings.TrimSpace(c.RegisteredBy)
	return c
}

func (c CodeRecord) Validate() error {
	c = c.Normalize()
	if err := ValidateHash("actor registry code hash", c.CodeHash); err != nil {
		return err
	}
	if c.RegisteredBy == "" {
		return errors.New("actor registry code registered_by is required")
	}
	if c.RegisteredAt == 0 {
		return errors.New("actor registry code registered_at must be positive")
	}
	return nil
}

func (a ActorRecord) Normalize() ActorRecord {
	a.ActorID = strings.TrimSpace(a.ActorID)
	a.ContractAddress = strings.TrimSpace(a.ContractAddress)
	a.Owner = strings.TrimSpace(a.Owner)
	a.CodeHash = strings.TrimSpace(a.CodeHash)
	a.StorageRoot = strings.TrimSpace(a.StorageRoot)
	a.MailboxRoot = strings.TrimSpace(a.MailboxRoot)
	a.Status = strings.TrimSpace(a.Status)
	a.RentStatus = strings.TrimSpace(a.RentStatus)
	a.MigratedFrom = strings.TrimSpace(a.MigratedFrom)
	a.MigratedTo = strings.TrimSpace(a.MigratedTo)
	a.Capabilities = normalizeStrings(a.Capabilities)
	return a
}

func (a ActorRecord) Validate(params ActorRegistryParams) error {
	a = a.Normalize()
	if err := ValidateHash("actor id", a.ActorID); err != nil {
		return err
	}
	if err := ValidateHash("actor contract address", a.ContractAddress); err != nil {
		return err
	}
	if a.Owner == "" {
		return errors.New("actor owner is required")
	}
	if err := ValidateHash("actor code hash", a.CodeHash); err != nil {
		return err
	}
	if err := ValidateHash("actor storage root", a.StorageRoot); err != nil {
		return err
	}
	if err := ValidateHash("actor mailbox root", a.MailboxRoot); err != nil {
		return err
	}
	if !IsActorStatus(a.Status) {
		return errors.New("actor status is invalid")
	}
	if !IsRentStatus(a.RentStatus) {
		return errors.New("actor rent status is invalid")
	}
	if a.LastActiveHeight == 0 {
		return errors.New("actor last active height must be positive")
	}
	if uint32(len(a.Capabilities)) > params.MaxCapabilities {
		return errors.New("actor capabilities exceed limit")
	}
	for _, capability := range a.Capabilities {
		if len(capability) > MaxCapabilityBytes {
			return errors.New("actor capability exceeds max bytes")
		}
	}
	return nil
}

func (m ActorMigrationRecord) Validate() error {
	if err := ValidateHash("actor migration actor id", strings.TrimSpace(m.ActorID)); err != nil {
		return err
	}
	if err := ValidateHash("actor migration from code hash", strings.TrimSpace(m.FromCodeHash)); err != nil {
		return err
	}
	if err := ValidateHash("actor migration to code hash", strings.TrimSpace(m.ToCodeHash)); err != nil {
		return err
	}
	if m.Height == 0 || m.LogicalTime == 0 {
		return errors.New("actor migration height and logical time must be positive")
	}
	return nil
}

func DeriveActorID(owner, codeHash, salt string) string {
	return hashParts("aetra-actor-id-v1", strings.TrimSpace(owner), strings.TrimSpace(codeHash), strings.TrimSpace(salt))
}

func DeriveContractAddress(actorID string) string {
	return hashParts("aetra-contract-address-v1", strings.TrimSpace(actorID))
}

func DefaultRoot(seed string) string {
	return hashParts("aetra-empty-root-v1", seed)
}

func CanExecuteNormalMessage(actor ActorRecord) bool {
	return actor.Normalize().Status == ActorStatusActive
}

func CanReceiveValue(actor ActorRecord, params ActorRegistryParams) bool {
	actor = actor.Normalize()
	if actor.Status != ActorStatusDeleted {
		return true
	}
	return params.AllowDeletedValueSend || params.DeletedValuePolicy == DeletedValuePolicyRedirect || params.DeletedValuePolicy == DeletedValuePolicyRefund
}

func NextLogicalTime(current, requested uint64) (uint64, error) {
	if requested == 0 {
		return current + 1, nil
	}
	if requested <= current {
		return 0, errors.New("actor logical time must monotonically increase")
	}
	return requested, nil
}

func IsActorStatus(status string) bool {
	switch status {
	case ActorStatusActive, ActorStatusFrozen, ActorStatusDeleted, ActorStatusMigrated:
		return true
	default:
		return false
	}
}

func IsRentStatus(status string) bool {
	switch status {
	case RentStatusCurrent, RentStatusDue, RentStatusDelinquent:
		return true
	default:
		return false
	}
}

func IsDeletedValuePolicy(policy string) bool {
	switch policy {
	case DeletedValuePolicyReject, DeletedValuePolicyRedirect, DeletedValuePolicyRefund:
		return true
	default:
		return false
	}
}

func ValidateHash(name, value string) error {
	value = strings.TrimSpace(value)
	if len(value) != 64 {
		return fmt.Errorf("%s must be 32-byte hex", name)
	}
	if _, err := hex.DecodeString(value); err != nil {
		return fmt.Errorf("%s must be hex: %w", name, err)
	}
	return nil
}

func SortActors(actors []ActorRecord) {
	sort.SliceStable(actors, func(i, j int) bool { return actors[i].Normalize().ActorID < actors[j].Normalize().ActorID })
}

func SortCodeStore(codes []CodeRecord) {
	sort.SliceStable(codes, func(i, j int) bool { return codes[i].Normalize().CodeHash < codes[j].Normalize().CodeHash })
}

func SortMigrations(migrations []ActorMigrationRecord) {
	sort.SliceStable(migrations, func(i, j int) bool {
		if migrations[i].ActorID != migrations[j].ActorID {
			return migrations[i].ActorID < migrations[j].ActorID
		}
		return migrations[i].LogicalTime < migrations[j].LogicalTime
	})
}

func normalizeStrings(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, found := seen[value]; found {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func cloneActors(actors []ActorRecord) []ActorRecord {
	out := make([]ActorRecord, len(actors))
	for i, actor := range actors {
		out[i] = actor.Normalize()
	}
	return out
}

func cloneCodeStore(codes []CodeRecord) []CodeRecord {
	out := make([]CodeRecord, len(codes))
	for i, code := range codes {
		out[i] = code.Normalize()
	}
	return out
}

func cloneMigrations(migrations []ActorMigrationRecord) []ActorMigrationRecord {
	out := append([]ActorMigrationRecord(nil), migrations...)
	for i := range out {
		out[i].ActorID = strings.TrimSpace(out[i].ActorID)
		out[i].FromCodeHash = strings.TrimSpace(out[i].FromCodeHash)
		out[i].ToCodeHash = strings.TrimSpace(out[i].ToCodeHash)
	}
	return out
}

func hashParts(parts ...string) string {
	h := sha256.New()
	for _, part := range parts {
		h.Write([]byte(part))
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}
