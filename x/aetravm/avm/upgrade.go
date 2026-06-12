package avm

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/sovereign-l1/l1/x/aetravm/chunk"
	"lukechampine.com/blake3"
)

type ContractVersion struct {
	SchemaVersion	uint32
	CodeVersion	uint32
	StateVersion	uint32
}

func (v ContractVersion) Equals(other ContractVersion) bool {
	return v.SchemaVersion == other.SchemaVersion &&
		v.CodeVersion == other.CodeVersion &&
		v.StateVersion == other.StateVersion
}

func (v ContractVersion) IsZero() bool {
	return v.SchemaVersion == 0 && v.CodeVersion == 0 && v.StateVersion == 0
}

func (v ContractVersion) Encode() []byte {
	buf := make([]byte, 12)
	binary.BigEndian.PutUint32(buf[0:4], v.SchemaVersion)
	binary.BigEndian.PutUint32(buf[4:8], v.CodeVersion)
	binary.BigEndian.PutUint32(buf[8:12], v.StateVersion)
	return buf
}

func DecodeContractVersion(data []byte) (ContractVersion, error) {
	if len(data) < 12 {
		return ContractVersion{}, fmt.Errorf("AVM upgrade: version data too short: %d bytes", len(data))
	}
	return ContractVersion{
		SchemaVersion:	binary.BigEndian.Uint32(data[0:4]),
		CodeVersion:	binary.BigEndian.Uint32(data[4:8]),
		StateVersion:	binary.BigEndian.Uint32(data[8:12]),
	}, nil
}

type UpgradeAuthority uint8

const (
	AuthorityNone	UpgradeAuthority	= iota
	AuthorityAdmin
	AuthorityGovernance
	AuthoritySystem
)

func (a UpgradeAuthority) String() string {
	switch a {
	case AuthorityNone:
		return "NONE"
	case AuthorityAdmin:
		return "ADMIN"
	case AuthorityGovernance:
		return "GOVERNANCE"
	case AuthoritySystem:
		return "SYSTEM"
	default:
		return "UNKNOWN"
	}
}

type UpgradeCapability uint64

const (
	UpgradeFlagNone		UpgradeCapability	= 0
	UpgradeFlagAllowed	UpgradeCapability	= 1 << iota
	UpgradeFlagAdminOnly
	UpgradeFlagGovernanceOnly
	UpgradeFlagSystemOnly
	UpgradeFlagMigrationRequired
	UpgradeFlagDisableAfterFirst
)

func (c UpgradeCapability) IsUpgradeable() bool {
	return c&UpgradeFlagAllowed != 0
}

func (c UpgradeCapability) RequiresMigration() bool {
	return c&UpgradeFlagMigrationRequired != 0
}

type UpgradeMessageType uint8

const (
	MsgUpgradeContractCode	UpgradeMessageType	= iota
	MsgMigrateContractState
	MsgSetContractAdmin
	MsgDisableContractUpgrades
)

func (m UpgradeMessageType) String() string {
	switch m {
	case MsgUpgradeContractCode:
		return "upgrade_code"
	case MsgMigrateContractState:
		return "migrate_state"
	case MsgSetContractAdmin:
		return "set_admin"
	case MsgDisableContractUpgrades:
		return "disable_upgrades"
	default:
		return "unknown"
	}
}

type UpgradeMessage struct {
	Type			UpgradeMessageType
	ContractAddress		string
	Caller			string
	Authority		UpgradeAuthority
	NewCodeHash		[32]byte
	MigrationHandler	[32]byte
	TargetSchema		ContractVersion
	GasLimit		uint64
	Signature		[]byte
}

type ContractState struct {
	Address		string
	Version		ContractVersion
	CodeHash	[32]byte
	Admin		string
	Capabilities	UpgradeCapability
	StateRoot	*chunk.Chunk
	UpgradeCount	uint32
	UpgradeDisabled	bool
}

type MigrationHandler func(oldRoot *chunk.Chunk, oldSchema, newSchema ContractVersion, gasLimit uint64) (*chunk.Chunk, uint64, error)

type MigrationRegistry struct {
	handlers	map[[32]byte]MigrationHandler
	schemas		map[string][]ContractVersion
}

func NewMigrationRegistry() *MigrationRegistry {
	return &MigrationRegistry{
		handlers:	make(map[[32]byte]MigrationHandler),
		schemas:	make(map[string][]ContractVersion),
	}
}

func (r *MigrationRegistry) Register(codeHash [32]byte, handler MigrationHandler, schemas []ContractVersion) {
	r.handlers[codeHash] = handler
	r.schemas[string(codeHash[:])] = schemas
}

func (r *MigrationRegistry) Get(codeHash [32]byte) (MigrationHandler, bool) {
	h, ok := r.handlers[codeHash]
	return h, ok
}

func (r *MigrationRegistry) Has(codeHash [32]byte) bool {
	_, ok := r.handlers[codeHash]
	return ok
}

type CompatibilityResult uint8

const (
	CompatibleNoMigration	CompatibilityResult	= iota
	CompatibleWithMigration
	Incompatible
)

func CheckStateCompatibility(oldSchema, newSchema ContractVersion) CompatibilityResult {
	if oldSchema.SchemaVersion == newSchema.SchemaVersion {
		return CompatibleNoMigration
	}
	if newSchema.SchemaVersion > oldSchema.SchemaVersion {
		return CompatibleWithMigration
	}
	return Incompatible
}

var (
	ErrContractImmutable		= errors.New("AVM upgrade: contract is permanently immutable")
	ErrUnauthorizedUpgrade		= errors.New("AVM upgrade: unauthorized upgrade authority")
	ErrUpgradeDisabled		= errors.New("AVM upgrade: upgrades have been disabled")
	ErrMigrationHandlerMissing	= errors.New("AVM upgrade: migration handler required but missing")
	ErrStateIncompatible		= errors.New("AVM upgrade: state schema incompatible, migration required")
	ErrInvalidCodeHash		= errors.New("AVM upgrade: invalid code hash")
	ErrSystemAuthorityRequired	= errors.New("AVM upgrade: system authority required for this operation")
	ErrGovernanceSignatureInvalid	= errors.New("AVM upgrade: governance signature invalid")
)

func ValidateUpgrade(state *ContractState, msg UpgradeMessage) error {
	if state.UpgradeDisabled {
		return ErrUpgradeDisabled
	}

	if !state.Capabilities.IsUpgradeable() {
		return ErrContractImmutable
	}

	if !ValidateUpgradeAuthority(state, msg) {
		return ErrUnauthorizedUpgrade
	}

	if msg.NewCodeHash == [32]byte{} {
		return ErrInvalidCodeHash
	}

	return nil
}

func ValidateUpgradeAuthority(state *ContractState, msg UpgradeMessage) bool {
	switch msg.Authority {
	case AuthorityNone:
		return false
	case AuthorityAdmin:
		return msg.Caller == state.Admin
	case AuthorityGovernance:
		return len(msg.Signature) > 0
	case AuthoritySystem:
		return len(msg.Signature) > 0 && state.Capabilities&UpgradeFlagSystemOnly != 0
	default:
		return false
	}
}

type MigrationResult struct {
	OldStateRoot	*chunk.Chunk
	NewStateRoot	*chunk.Chunk
	OldVersion	ContractVersion
	NewVersion	ContractVersion
	GasUsed		uint64
	Success		bool
	RolledBack	bool
}

func ExecuteMigration(
	state *ContractState,
	handler MigrationHandler,
	newCodeHash [32]byte,
	newSchema ContractVersion,
	gasLimit uint64,
) (*MigrationResult, error) {
	if handler == nil {
		return nil, ErrMigrationHandlerMissing
	}

	oldRoot := state.StateRoot
	oldVersion := state.Version

	newRoot, gasUsed, err := handler(oldRoot, oldVersion, newSchema, gasLimit)
	if err != nil || newRoot == nil {
		return &MigrationResult{
			OldStateRoot:	oldRoot,
			NewStateRoot:	oldRoot,
			OldVersion:	oldVersion,
			NewVersion:	oldVersion,
			GasUsed:	gasUsed,
			Success:	false,
			RolledBack:	true,
		}, nil
	}

	return &MigrationResult{
		OldStateRoot:	oldRoot,
		NewStateRoot:	newRoot,
		OldVersion:	oldVersion,
		NewVersion:	newSchema,
		GasUsed:	gasUsed,
		Success:	true,
		RolledBack:	false,
	}, nil
}

type MigrationReceipt struct {
	SchemaVersionBefore	ContractVersion
	SchemaVersionAfter	ContractVersion
	StateRootBefore		[]byte
	StateRootAfter		[]byte
	MigrationGasUsed	uint64
	Success			bool
	RolledBack		bool
	AuthorityType		UpgradeAuthority
	MigrationHandlerHash	[32]byte
	CodeHashBefore		[32]byte
	CodeHashAfter		[32]byte
	ContractAddress		string
	UpgradeCount		uint32
}

func (r *MigrationReceipt) CanonicalEncode() []byte {
	buf := make([]byte, 0, 256)
	buf = append(buf, r.SchemaVersionBefore.Encode()...)
	buf = append(buf, r.SchemaVersionAfter.Encode()...)
	buf = binary.BigEndian.AppendUint16(buf, uint16(len(r.StateRootBefore)))
	buf = append(buf, r.StateRootBefore...)
	buf = binary.BigEndian.AppendUint16(buf, uint16(len(r.StateRootAfter)))
	buf = append(buf, r.StateRootAfter...)
	buf = binary.BigEndian.AppendUint64(buf, r.MigrationGasUsed)
	if r.Success {
		buf = append(buf, 0x01)
	} else {
		buf = append(buf, 0x00)
	}
	if r.RolledBack {
		buf = append(buf, 0x01)
	} else {
		buf = append(buf, 0x00)
	}
	buf = append(buf, byte(r.AuthorityType))
	buf = append(buf, r.MigrationHandlerHash[:]...)
	buf = append(buf, r.CodeHashBefore[:]...)
	buf = append(buf, r.CodeHashAfter[:]...)
	buf = binary.BigEndian.AppendUint32(buf, uint32(len(r.ContractAddress)))
	buf = append(buf, r.ContractAddress...)
	buf = binary.BigEndian.AppendUint32(buf, r.UpgradeCount)
	return buf
}

func MigrationReceiptHash(receipt *MigrationReceipt) [32]byte {
	encoded := receipt.CanonicalEncode()
	return blake3.Sum256(encoded)
}

type UpgradeEngine struct {
	registry	*MigrationRegistry
	moduleHash	map[[32]byte]bool
}

func NewUpgradeEngine(registry *MigrationRegistry) *UpgradeEngine {
	return &UpgradeEngine{
		registry:	registry,
		moduleHash:	make(map[[32]byte]bool),
	}
}

func (e *UpgradeEngine) RegisterVerifiedModule(codeHash [32]byte) {
	e.moduleHash[codeHash] = true
}

func (e *UpgradeEngine) IsModuleVerified(codeHash [32]byte) bool {
	return e.moduleHash[codeHash]
}

func (e *UpgradeEngine) ProcessUpgrade(
	state *ContractState,
	msg UpgradeMessage,
) (*ContractState, *MigrationReceipt, error) {
	if err := ValidateUpgrade(state, msg); err != nil {
		return nil, nil, err
	}

	if !e.IsModuleVerified(msg.NewCodeHash) {
		return nil, nil, ErrInvalidCodeHash
	}

	compat := CheckStateCompatibility(state.Version, msg.TargetSchema)
	if compat == Incompatible {
		return nil, nil, ErrStateIncompatible
	}

	handler, _ := e.registry.Get(msg.NewCodeHash)
	if compat == CompatibleWithMigration && handler == nil {
		return nil, nil, ErrMigrationHandlerMissing
	}

	var newStateRoot *chunk.Chunk
	var gasUsed uint64
	var migrated bool
	var rolledBack bool

	oldRoot := state.StateRoot

	if handler != nil {
		result, err := ExecuteMigration(state, handler, msg.NewCodeHash, msg.TargetSchema, msg.GasLimit)
		if err != nil {
			return nil, nil, err
		}

		if result.Success {
			newStateRoot = result.NewStateRoot
			gasUsed = result.GasUsed
			migrated = true
		} else {
			rolledBack = result.RolledBack
			newStateRoot = oldRoot
		}
	} else {
		newStateRoot = oldRoot
	}

	var stateRootBefore, stateRootAfter []byte
	if oldRoot != nil {
		stateRootBefore = oldRoot.Hash()
	} else {
		stateRootBefore = make([]byte, 32)
	}
	if newStateRoot != nil {
		stateRootAfter = newStateRoot.Hash()
	} else {
		stateRootAfter = make([]byte, 32)
	}

	receipt := &MigrationReceipt{
		SchemaVersionBefore:	state.Version,
		SchemaVersionAfter:	msg.TargetSchema,
		StateRootBefore:	stateRootBefore,
		StateRootAfter:		stateRootAfter,
		MigrationGasUsed:	gasUsed,
		Success:		!rolledBack,
		RolledBack:		rolledBack,
		AuthorityType:		msg.Authority,
		MigrationHandlerHash:	msg.MigrationHandler,
		CodeHashBefore:		state.CodeHash,
		CodeHashAfter:		msg.NewCodeHash,
		ContractAddress:	state.Address,
		UpgradeCount:		state.UpgradeCount + 1,
	}

	newState := &ContractState{
		Address:		state.Address,
		Version:		msg.TargetSchema,
		CodeHash:		msg.NewCodeHash,
		Admin:			state.Admin,
		Capabilities:		state.Capabilities,
		StateRoot:		newStateRoot,
		UpgradeCount:		state.UpgradeCount + 1,
		UpgradeDisabled:	state.UpgradeDisabled,
	}

	if migrated && !rolledBack {
		newState.Version = msg.TargetSchema
	}

	return newState, receipt, nil
}

func (e *UpgradeEngine) DisableUpgrades(state *ContractState, msg UpgradeMessage) (*ContractState, error) {
	if !ValidateUpgradeAuthority(state, msg) {
		return nil, ErrUnauthorizedUpgrade
	}

	newState := &ContractState{
		Address:		state.Address,
		Version:		state.Version,
		CodeHash:		state.CodeHash,
		Admin:			state.Admin,
		Capabilities:		state.Capabilities,
		StateRoot:		state.StateRoot,
		UpgradeCount:		state.UpgradeCount,
		UpgradeDisabled:	true,
	}
	return newState, nil
}

func (e *UpgradeEngine) SetAdmin(state *ContractState, msg UpgradeMessage) (*ContractState, error) {
	if msg.Authority != AuthorityAdmin && msg.Authority != AuthorityGovernance {
		return nil, ErrUnauthorizedUpgrade
	}
	if msg.Caller != state.Admin && msg.Authority != AuthorityGovernance {
		return nil, ErrUnauthorizedUpgrade
	}

	newState := &ContractState{
		Address:		state.Address,
		Version:		state.Version,
		CodeHash:		state.CodeHash,
		Admin:			msg.Caller,
		Capabilities:		state.Capabilities,
		StateRoot:		state.StateRoot,
		UpgradeCount:		state.UpgradeCount,
		UpgradeDisabled:	state.UpgradeDisabled,
	}
	return newState, nil
}

func ValidateSystemUpgrade(state *ContractState, msg UpgradeMessage) error {
	if msg.Authority != AuthoritySystem {
		return ErrSystemAuthorityRequired
	}
	if len(msg.Signature) == 0 {
		return ErrGovernanceSignatureInvalid
	}
	return nil
}

func EnforceImmutability(state *ContractState) error {
	if !state.Capabilities.IsUpgradeable() {
		return ErrContractImmutable
	}
	if state.UpgradeDisabled {
		return ErrUpgradeDisabled
	}
	return nil
}
