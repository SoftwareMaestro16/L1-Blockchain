package avm

import (
	"fmt"
	"testing"

	"github.com/sovereign-l1/l1/x/aetravm/chunk"
)

func TestContractVersionEquals(t *testing.T) {
	v1 := ContractVersion{SchemaVersion: 1, CodeVersion: 1, StateVersion: 1}
	v2 := ContractVersion{SchemaVersion: 1, CodeVersion: 1, StateVersion: 1}
	if !v1.Equals(v2) {
		t.Error("identical versions should be equal")
	}

	v3 := ContractVersion{SchemaVersion: 2, CodeVersion: 1, StateVersion: 1}
	if v1.Equals(v3) {
		t.Error("different versions should not be equal")
	}
}

func TestContractVersionIsZero(t *testing.T) {
	v0 := ContractVersion{}
	if !v0.IsZero() {
		t.Error("zero version should be zero")
	}
	v1 := ContractVersion{SchemaVersion: 1}
	if v1.IsZero() {
		t.Error("non-zero version should not be zero")
	}
}

func TestContractVersionEncodeDecode(t *testing.T) {
	v := ContractVersion{SchemaVersion: 3, CodeVersion: 5, StateVersion: 7}
	encoded := v.Encode()
	if len(encoded) != 12 {
		t.Errorf("expected 12 bytes, got %d", len(encoded))
	}

	decoded, err := DecodeContractVersion(encoded)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !v.Equals(decoded) {
		t.Errorf("round trip failed: %v vs %v", v, decoded)
	}
}

func TestContractVersionDecodeTooShort(t *testing.T) {
	_, err := DecodeContractVersion([]byte{1, 2, 3})
	if err == nil {
		t.Error("should reject too-short version data")
	}
}

func TestUpgradeAuthorityStrings(t *testing.T) {
	tests := map[UpgradeAuthority]string{
		AuthorityNone:		"NONE",
		AuthorityAdmin:		"ADMIN",
		AuthorityGovernance:	"GOVERNANCE",
		AuthoritySystem:	"SYSTEM",
	}
	for auth, expected := range tests {
		if auth.String() != expected {
			t.Errorf("expected %s, got %s", expected, auth.String())
		}
	}
}

func TestImmutableContractCannotUpgrade(t *testing.T) {
	state := &ContractState{
		Address:	"4:abc123",
		Version:	ContractVersion{SchemaVersion: 1, CodeVersion: 1, StateVersion: 1},
		CodeHash:	[32]byte{1},
		Admin:		"AE:admin1",
		Capabilities:	UpgradeFlagNone,
	}

	err := EnforceImmutability(state)
	if err != ErrContractImmutable {
		t.Errorf("expected ErrContractImmutable, got %v", err)
	}
}

func TestUpgradeableContractPassesImmutability(t *testing.T) {
	state := &ContractState{
		Address:	"4:abc123",
		Version:	ContractVersion{SchemaVersion: 1, CodeVersion: 1, StateVersion: 1},
		CodeHash:	[32]byte{1},
		Admin:		"AE:admin1",
		Capabilities:	UpgradeFlagAllowed,
	}

	err := EnforceImmutability(state)
	if err != nil {
		t.Errorf("upgradeable contract should pass: %v", err)
	}
}

func TestDisabledUpgradesBlocked(t *testing.T) {
	state := &ContractState{
		Address:		"4:abc123",
		Version:		ContractVersion{SchemaVersion: 1, CodeVersion: 1, StateVersion: 1},
		CodeHash:		[32]byte{1},
		Admin:			"AE:admin1",
		Capabilities:		UpgradeFlagAllowed,
		UpgradeDisabled:	true,
	}

	err := EnforceImmutability(state)
	if err != ErrUpgradeDisabled {
		t.Errorf("expected ErrUpgradeDisabled, got %v", err)
	}
}

func TestAdminUpgradeAuthorized(t *testing.T) {
	state := &ContractState{
		Address:	"4:abc123",
		Admin:		"AE:admin1",
		Capabilities:	UpgradeFlagAllowed,
	}
	msg := UpgradeMessage{
		Type:		MsgUpgradeContractCode,
		Caller:		"AE:admin1",
		Authority:	AuthorityAdmin,
	}

	if !ValidateUpgradeAuthority(state, msg) {
		t.Error("admin should be authorized to upgrade")
	}
}

func TestNonAdminUpgradeRejected(t *testing.T) {
	state := &ContractState{
		Address:	"4:abc123",
		Admin:		"AE:admin1",
		Capabilities:	UpgradeFlagAllowed,
	}
	msg := UpgradeMessage{
		Type:		MsgUpgradeContractCode,
		Caller:		"AE:attacker",
		Authority:	AuthorityAdmin,
	}

	if ValidateUpgradeAuthority(state, msg) {
		t.Error("non-admin should be rejected")
	}
}

func TestNoneAuthorityRejected(t *testing.T) {
	state := &ContractState{
		Address:	"4:abc123",
		Admin:		"AE:admin1",
		Capabilities:	UpgradeFlagAllowed,
	}
	msg := UpgradeMessage{
		Type:		MsgUpgradeContractCode,
		Caller:		"AE:admin1",
		Authority:	AuthorityNone,
	}

	if ValidateUpgradeAuthority(state, msg) {
		t.Error("NONE authority should be rejected")
	}
}

func TestCompatibleNoMigration(t *testing.T) {
	old := ContractVersion{SchemaVersion: 1, CodeVersion: 1, StateVersion: 1}
	new_ := ContractVersion{SchemaVersion: 1, CodeVersion: 2, StateVersion: 1}

	result := CheckStateCompatibility(old, new_)
	if result != CompatibleNoMigration {
		t.Errorf("same schema version should be CompatibleNoMigration, got %v", result)
	}
}

func TestCompatibleWithMigration(t *testing.T) {
	old := ContractVersion{SchemaVersion: 1, CodeVersion: 1, StateVersion: 1}
	new_ := ContractVersion{SchemaVersion: 2, CodeVersion: 2, StateVersion: 2}

	result := CheckStateCompatibility(old, new_)
	if result != CompatibleWithMigration {
		t.Errorf("higher schema should require migration, got %v", result)
	}
}

func TestIncompatibleDowngrade(t *testing.T) {
	old := ContractVersion{SchemaVersion: 3, CodeVersion: 1, StateVersion: 1}
	new_ := ContractVersion{SchemaVersion: 2, CodeVersion: 1, StateVersion: 1}

	result := CheckStateCompatibility(old, new_)
	if result != Incompatible {
		t.Errorf("schema downgrade should be incompatible, got %v", result)
	}
}

func TestMigrationSuccess(t *testing.T) {
	registry := NewMigrationRegistry()
	migrationCalled := false

	handler := func(oldRoot *chunk.Chunk, oldSchema, newSchema ContractVersion, gasLimit uint64) (*chunk.Chunk, uint64, error) {
		migrationCalled = true
		m := chunk.NewEmptyMap()
		b, _ := chunk.NewBuilder().SetTypeTag(chunk.TypeNormal).SetData([]byte("migrated"), 56).Build()
		newMap, _ := m.Put([]byte("migrated"), b)
		return newMap.Root(), 1000, nil
	}

	var codeHash [32]byte
	codeHash[0] = 1
	registry.Register(codeHash, handler, nil)

	m := chunk.NewEmptyMap()
	b, _ := chunk.NewBuilder().SetTypeTag(chunk.TypeNormal).SetData([]byte("initial"), 48).Build()
	m, _ = m.Put([]byte("init"), b)
	stateRoot := m.Root()

	state := &ContractState{
		Address:	"4:contract1",
		Version:	ContractVersion{SchemaVersion: 1, CodeVersion: 1, StateVersion: 1},
		CodeHash:	[32]byte{0},
		Admin:		"AE:admin1",
		Capabilities:	UpgradeFlagAllowed,
		StateRoot:	stateRoot,
	}

	engine := NewUpgradeEngine(registry)
	engine.RegisterVerifiedModule(codeHash)

	msg := UpgradeMessage{
		Type:			MsgUpgradeContractCode,
		ContractAddress:	"4:contract1",
		Caller:			"AE:admin1",
		Authority:		AuthorityAdmin,
		NewCodeHash:		codeHash,
		MigrationHandler:	codeHash,
		TargetSchema:		ContractVersion{SchemaVersion: 2, CodeVersion: 2, StateVersion: 2},
		GasLimit:		100000,
	}

	newState, receipt, err := engine.ProcessUpgrade(state, msg)
	if err != nil {
		t.Fatalf("upgrade failed: %v", err)
	}
	if !migrationCalled {
		t.Error("migration handler should have been called")
	}
	if !receipt.Success {
		t.Error("migration should succeed")
	}
	if receipt.RolledBack {
		t.Error("migration should not be rolled back")
	}
	if newState.UpgradeCount != 1 {
		t.Errorf("expected upgrade count 1, got %d", newState.UpgradeCount)
	}
}

func TestMigrationRollbackOnFailure(t *testing.T) {
	registry := NewMigrationRegistry()

	handler := func(oldRoot *chunk.Chunk, oldSchema, newSchema ContractVersion, gasLimit uint64) (*chunk.Chunk, uint64, error) {
		return nil, 500, fmt.Errorf("migration failed: schema incompatibility")
	}

	var codeHash [32]byte
	codeHash[0] = 2
	registry.Register(codeHash, handler, nil)

	m := chunk.NewEmptyMap()
	b, _ := chunk.NewBuilder().SetTypeTag(chunk.TypeNormal).SetData([]byte("original"), 56).Build()
	m, _ = m.Put([]byte("state"), b)
	originalRoot := m.Root()

	state := &ContractState{
		Address:	"4:contract1",
		Version:	ContractVersion{SchemaVersion: 1, CodeVersion: 1, StateVersion: 1},
		CodeHash:	[32]byte{0},
		Admin:		"AE:admin1",
		Capabilities:	UpgradeFlagAllowed,
		StateRoot:	originalRoot,
	}

	engine := NewUpgradeEngine(registry)
	engine.RegisterVerifiedModule(codeHash)

	msg := UpgradeMessage{
		Type:			MsgUpgradeContractCode,
		ContractAddress:	"4:contract1",
		Caller:			"AE:admin1",
		Authority:		AuthorityAdmin,
		NewCodeHash:		codeHash,
		MigrationHandler:	codeHash,
		TargetSchema:		ContractVersion{SchemaVersion: 2, CodeVersion: 2, StateVersion: 2},
		GasLimit:		100000,
	}

	_, receipt, err := engine.ProcessUpgrade(state, msg)
	if err != nil {
		t.Fatalf("upgrade should not error even on migration failure: %v", err)
	}
	if receipt.Success {
		t.Error("migration failure should produce unsuccessful receipt")
	}
	if !receipt.RolledBack {
		t.Error("failed migration should be rolled back")
	}
}

func TestMigrationHandlerMissing(t *testing.T) {
	registry := NewMigrationRegistry()

	m := chunk.NewEmptyMap()
	b, _ := chunk.NewBuilder().SetTypeTag(chunk.TypeSystem).SetData([]byte{}, 0).Build()
	m, _ = m.Put([]byte("__init__"), b)
	stateRoot := m.Root()

	state := &ContractState{
		Address:	"4:contract1",
		Version:	ContractVersion{SchemaVersion: 1, CodeVersion: 1, StateVersion: 1},
		CodeHash:	[32]byte{0},
		Admin:		"AE:admin1",
		Capabilities:	UpgradeFlagAllowed,
		StateRoot:	stateRoot,
	}

	var codeHash [32]byte
	codeHash[0] = 3

	engine := NewUpgradeEngine(registry)
	engine.RegisterVerifiedModule(codeHash)

	msg := UpgradeMessage{
		Type:			MsgUpgradeContractCode,
		ContractAddress:	"4:contract1",
		Caller:			"AE:admin1",
		Authority:		AuthorityAdmin,
		NewCodeHash:		codeHash,
		TargetSchema:		ContractVersion{SchemaVersion: 2, CodeVersion: 2, StateVersion: 2},
		GasLimit:		100000,
	}

	_, _, err := engine.ProcessUpgrade(state, msg)
	if err != ErrMigrationHandlerMissing {
		t.Errorf("expected ErrMigrationHandlerMissing, got %v", err)
	}
}

func TestDisableUpgrades(t *testing.T) {
	state := &ContractState{
		Address:	"4:contract1",
		Admin:		"AE:admin1",
		Capabilities:	UpgradeFlagAllowed,
	}

	msg := UpgradeMessage{
		Type:		MsgDisableContractUpgrades,
		Caller:		"AE:admin1",
		Authority:	AuthorityAdmin,
	}

	engine := NewUpgradeEngine(NewMigrationRegistry())
	newState, err := engine.DisableUpgrades(state, msg)
	if err != nil {
		t.Fatalf("disable upgrades: %v", err)
	}
	if !newState.UpgradeDisabled {
		t.Error("upgrades should be disabled")
	}

	err2 := EnforceImmutability(newState)
	if err2 != ErrUpgradeDisabled {
		t.Error("disabled contract should not be upgradeable")
	}
}

func TestSetAdmin(t *testing.T) {
	state := &ContractState{
		Address:	"4:contract1",
		Admin:		"AE:admin1",
		Capabilities:	UpgradeFlagAllowed,
	}

	msg := UpgradeMessage{
		Type:		MsgSetContractAdmin,
		Caller:		"AE:admin1",
		Authority:	AuthorityAdmin,
	}

	engine := NewUpgradeEngine(NewMigrationRegistry())
	newState, err := engine.SetAdmin(state, msg)
	if err != nil {
		t.Fatalf("set admin: %v", err)
	}
	if newState.Admin != "AE:admin1" {
		t.Errorf("expected admin to remain AE:admin1, got %s", newState.Admin)
	}
}

func TestSetAdminNonAdminRejected(t *testing.T) {
	state := &ContractState{
		Address:	"4:contract1",
		Admin:		"AE:admin1",
		Capabilities:	UpgradeFlagAllowed,
	}

	msg := UpgradeMessage{
		Type:		MsgSetContractAdmin,
		Caller:		"AE:attacker",
		Authority:	AuthorityAdmin,
	}

	engine := NewUpgradeEngine(NewMigrationRegistry())
	_, err := engine.SetAdmin(state, msg)
	if err == nil {
		t.Error("non-admin should not be able to set admin")
	}
}

func TestSystemOverrideRequiresSystemAuthority(t *testing.T) {
	state := &ContractState{
		Address:	"4:system1",
		Admin:		"AE:system",
		Capabilities:	UpgradeFlagAllowed | UpgradeFlagSystemOnly,
	}

	err := ValidateSystemUpgrade(state, UpgradeMessage{Authority: AuthorityAdmin})
	if err != ErrSystemAuthorityRequired {
		t.Errorf("expected ErrSystemAuthorityRequired, got %v", err)
	}
}

func TestSystemOverrideRequiresSignature(t *testing.T) {
	state := &ContractState{
		Address:	"4:system1",
		Admin:		"AE:system",
		Capabilities:	UpgradeFlagAllowed | UpgradeFlagSystemOnly,
	}

	err := ValidateSystemUpgrade(state, UpgradeMessage{
		Authority:	AuthoritySystem,
		Signature:	nil,
	})
	if err != ErrGovernanceSignatureInvalid {
		t.Errorf("expected ErrGovernanceSignatureInvalid, got %v", err)
	}
}

func TestSystemOverrideWithValidSignature(t *testing.T) {
	state := &ContractState{
		Address:	"4:system1",
		Admin:		"AE:system",
		Capabilities:	UpgradeFlagAllowed | UpgradeFlagSystemOnly,
	}

	err := ValidateSystemUpgrade(state, UpgradeMessage{
		Authority:	AuthoritySystem,
		Signature:	[]byte{1, 2, 3, 4},
	})
	if err != nil {
		t.Errorf("system override with valid signature should pass: %v", err)
	}
}

func TestUpgradeMessageTypeStrings(t *testing.T) {
	tests := map[UpgradeMessageType]string{
		MsgUpgradeContractCode:		"upgrade_code",
		MsgMigrateContractState:	"migrate_state",
		MsgSetContractAdmin:		"set_admin",
		MsgDisableContractUpgrades:	"disable_upgrades",
	}
	for mt, expected := range tests {
		if mt.String() != expected {
			t.Errorf("expected %s, got %s", expected, mt.String())
		}
	}
}

func TestMigrationReceiptCanonicalEncode(t *testing.T) {
	receipt := &MigrationReceipt{
		SchemaVersionBefore:	ContractVersion{SchemaVersion: 1, CodeVersion: 1, StateVersion: 1},
		SchemaVersionAfter:	ContractVersion{SchemaVersion: 2, CodeVersion: 2, StateVersion: 2},
		StateRootBefore:	make([]byte, 32),
		StateRootAfter:		make([]byte, 32),
		Success:		true,
		RolledBack:		false,
		AuthorityType:		AuthorityAdmin,
		ContractAddress:	"4:contract1",
		UpgradeCount:		1,
	}

	encoded := receipt.CanonicalEncode()
	if len(encoded) == 0 {
		t.Error("receipt canonical encoding should not be empty")
	}
}

func TestMigrationReceiptHashDeterministic(t *testing.T) {
	receipt := &MigrationReceipt{
		SchemaVersionBefore:	ContractVersion{SchemaVersion: 1, CodeVersion: 1, StateVersion: 1},
		SchemaVersionAfter:	ContractVersion{SchemaVersion: 2, CodeVersion: 2, StateVersion: 2},
		StateRootBefore:	make([]byte, 32),
		StateRootAfter:		make([]byte, 32),
		Success:		true,
		AuthorityType:		AuthorityAdmin,
		ContractAddress:	"4:contract1",
	}

	h1 := MigrationReceiptHash(receipt)
	h2 := MigrationReceiptHash(receipt)
	if h1 != h2 {
		t.Error("same receipt should produce same hash")
	}
}

func TestMigrationReceiptHashChangesOnMutation(t *testing.T) {
	r1 := &MigrationReceipt{
		SchemaVersionBefore:	ContractVersion{SchemaVersion: 1, CodeVersion: 1},
		SchemaVersionAfter:	ContractVersion{SchemaVersion: 2, CodeVersion: 2},
		StateRootBefore:	make([]byte, 32),
		StateRootAfter:		make([]byte, 32),
		Success:		true,
		AuthorityType:		AuthorityAdmin,
		ContractAddress:	"4:contract1",
	}
	r2 := &MigrationReceipt{
		SchemaVersionBefore:	ContractVersion{SchemaVersion: 1, CodeVersion: 1},
		SchemaVersionAfter:	ContractVersion{SchemaVersion: 3, CodeVersion: 3},
		StateRootBefore:	make([]byte, 32),
		StateRootAfter:		make([]byte, 32),
		Success:		true,
		AuthorityType:		AuthorityAdmin,
		ContractAddress:	"4:contract1",
	}

	h1 := MigrationReceiptHash(r1)
	h2 := MigrationReceiptHash(r2)
	if h1 == h2 {
		t.Error("different receipts should produce different hashes")
	}
}

func TestUpgradeCapabilityFlags(t *testing.T) {
	none := UpgradeFlagNone
	if none.IsUpgradeable() {
		t.Error("none flag should not be upgradeable")
	}
	if none.RequiresMigration() {
		t.Error("none flag should not require migration")
	}

	allowed := UpgradeFlagAllowed
	if !allowed.IsUpgradeable() {
		t.Error("allowed flag should be upgradeable")
	}

	required := UpgradeFlagAllowed | UpgradeFlagMigrationRequired
	if !required.RequiresMigration() {
		t.Error("migration required flag should require migration")
	}
}
