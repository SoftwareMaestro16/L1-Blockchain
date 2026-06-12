package avm

import (
	"math/big"
	"testing"
)

func newStateInit() *StateInit {
	return &StateInit{
		ABIVersion:		StateInitABIVersion,
		CodeHash:		[32]byte{1, 2, 3},
		DeployerAddress:	"AE:deployer1",
		ChainID:		"aetra-1",
		Namespace:		"default",
		Capabilities:		AllDeployCapabilities,
	}
}

func TestStateInitValidationValid(t *testing.T) {
	si := newStateInit()
	if err := si.Validate(); err != nil {
		t.Errorf("valid StateInit should pass: %v", err)
	}
}

func TestStateInitZeroDeployer(t *testing.T) {
	si := newStateInit()
	si.DeployerAddress = ""
	if err := si.Validate(); err != ErrZeroDeployer {
		t.Errorf("expected ErrZeroDeployer, got %v", err)
	}
}

func TestStateInitEmptyCodeHash(t *testing.T) {
	si := newStateInit()
	si.CodeHash = [32]byte{}
	if err := si.Validate(); err != ErrEmptyCodeHash {
		t.Errorf("expected ErrEmptyCodeHash, got %v", err)
	}
}

func TestStateInitInitDataTooLarge(t *testing.T) {
	si := newStateInit()
	si.InitData = make([]byte, MaxInitDataSize+1)
	if err := si.Validate(); err != ErrInitDataTooLarge {
		t.Errorf("expected ErrInitDataTooLarge, got %v", err)
	}
}

func TestStateInitSaltTooLarge(t *testing.T) {
	si := newStateInit()
	si.Salt = make([]byte, MaxSaltSize+1)
	if err := si.Validate(); err != ErrSaltTooLarge {
		t.Errorf("expected ErrSaltTooLarge, got %v", err)
	}
}

func TestStateInitInvalidABI(t *testing.T) {
	si := newStateInit()
	si.ABIVersion = 0
	if err := si.Validate(); err != ErrInvalidABI {
		t.Errorf("expected ErrInvalidABI, got %v", err)
	}
}

func TestStateInitEmptyChainID(t *testing.T) {
	si := newStateInit()
	si.ChainID = ""
	if err := si.Validate(); err != ErrEmptyChainID {
		t.Errorf("expected ErrEmptyChainID, got %v", err)
	}
}

func TestDeriveContractAddressDeterministic(t *testing.T) {
	si1 := newStateInit()
	si2 := newStateInit()

	addr1, err := DeriveContractAddress(si1)
	if err != nil {
		t.Fatalf("derive address 1: %v", err)
	}
	addr2, err := DeriveContractAddress(si2)
	if err != nil {
		t.Fatalf("derive address 2: %v", err)
	}

	if addr1.Internal != addr2.Internal {
		t.Errorf("same StateInit should produce same internal address: %s vs %s", addr1.Internal, addr2.Internal)
	}
	if addr1.External != addr2.External {
		t.Errorf("same StateInit should produce same external address: %s vs %s", addr1.External, addr2.External)
	}
	if addr1.rawHash != addr2.rawHash {
		t.Error("same StateInit should produce same raw hash")
	}
}

func TestInitDataChangesAddress(t *testing.T) {
	si1 := newStateInit()
	si2 := newStateInit()
	si2.InitData = []byte{1, 2, 3}

	addr1, _ := DeriveContractAddress(si1)
	addr2, _ := DeriveContractAddress(si2)

	if addr1.Internal == addr2.Internal {
		t.Error("different init data should produce different addresses")
	}
}

func TestSaltChangesAddress(t *testing.T) {
	si1 := newStateInit()
	si2 := newStateInit()
	si2.Salt = []byte("different_salt")

	addr1, _ := DeriveContractAddress(si1)
	addr2, _ := DeriveContractAddress(si2)

	if addr1.Internal == addr2.Internal {
		t.Error("different salt should produce different addresses")
	}
}

func TestOversizedInitDataRejected(t *testing.T) {
	si := newStateInit()
	si.InitData = make([]byte, MaxInitDataSize+1)
	if err := si.Validate(); err == nil {
		t.Error("oversized init data should be rejected")
	}
}

func TestDuplicateDeploymentRejected(t *testing.T) {
	si := newStateInit()
	existing := map[string]ContractDeployState{}

	result, err := DeployContract(si, existing)
	if err != nil {
		t.Fatalf("first deploy should succeed: %v", err)
	}
	if result.Error != nil {
		t.Fatalf("first deploy should have no error: %v", result.Error)
	}

	existing[result.Address.Internal] = ContractDeployed
	result2, err := DeployContract(si, existing)
	if err != nil {
		t.Fatalf("duplicate deploy check: %v", err)
	}
	if result2.Error != ErrDuplicateDeployment {
		t.Errorf("expected ErrDuplicateDeployment, got %v", result2.Error)
	}
}

func TestCounterfactualAddressQuery(t *testing.T) {
	si := newStateInit()
	addr, err := DeriveContractAddress(si)
	if err != nil {
		t.Fatalf("derive address: %v", err)
	}

	cs := QueryContractState(*addr, false, false)
	if cs.State != ContractNotDeployed {
		t.Errorf("expected NOT_DEPLOYED, got %v", cs.State)
	}

	cs = QueryContractState(*addr, true, false)
	if cs.State != ContractDeployed {
		t.Errorf("expected DEPLOYED, got %v", cs.State)
	}

	cs = QueryContractState(*addr, true, true)
	if cs.State != ContractInitialized {
		t.Errorf("expected INITIALIZED, got %v", cs.State)
	}
}

func TestZeroDeployerRejected(t *testing.T) {
	si := newStateInit()
	si.DeployerAddress = ""
	_, err := DeriveContractAddress(si)
	if err == nil {
		t.Error("zero deployer should be rejected")
	}
}

func TestMalformedCodeHashRejected(t *testing.T) {
	si := newStateInit()
	si.CodeHash = [32]byte{}
	_, err := DeriveContractAddress(si)
	if err == nil {
		t.Error("empty code hash should be rejected")
	}
}

func TestHashStateInitDeterminism(t *testing.T) {
	si1 := newStateInit()
	si2 := newStateInit()

	h1, err := HashStateInit(si1)
	if err != nil {
		t.Fatalf("hash 1: %v", err)
	}
	h2, err := HashStateInit(si2)
	if err != nil {
		t.Fatalf("hash 2: %v", err)
	}

	if h1 != h2 {
		t.Error("same StateInit should produce same hash")
	}
}

func TestHashStateInitChangesOnMutation(t *testing.T) {
	si1 := newStateInit()
	si2 := newStateInit()
	si2.InitData = []byte{0xFF}

	h1, _ := HashStateInit(si1)
	h2, _ := HashStateInit(si2)

	if h1 == h2 {
		t.Error("different StateInit should produce different hashes")
	}
}

func TestExportImportRoundTrip(t *testing.T) {
	si := newStateInit()
	si.InitData = []byte{1, 2, 3, 4}
	si.Salt = []byte("test_salt")

	exported, err := ExportStateInit(si)
	if err != nil {
		t.Fatalf("export: %v", err)
	}

	imported, err := ImportStateInit(exported)
	if err != nil {
		t.Fatalf("import: %v", err)
	}

	if imported.ABIVersion != si.ABIVersion {
		t.Errorf("ABI version mismatch: %d vs %d", imported.ABIVersion, si.ABIVersion)
	}
	if imported.CodeHash != si.CodeHash {
		t.Error("code hash mismatch")
	}
	if imported.DeployerAddress != si.DeployerAddress {
		t.Errorf("deployer mismatch: %s vs %s", imported.DeployerAddress, si.DeployerAddress)
	}
	if imported.ChainID != si.ChainID {
		t.Errorf("chain ID mismatch: %s vs %s", imported.ChainID, si.ChainID)
	}
	if imported.Namespace != si.Namespace {
		t.Errorf("namespace mismatch: %s vs %s", imported.Namespace, si.Namespace)
	}
	if imported.InitialBalance != si.InitialBalance {
		t.Errorf("balance mismatch: %d vs %d", imported.InitialBalance, si.InitialBalance)
	}
	if imported.Capabilities.Flags != si.Capabilities.Flags {
		t.Errorf("capabilities mismatch: %d vs %d", imported.Capabilities.Flags, si.Capabilities.Flags)
	}
}

func TestExportImportPreservesDerivedAddress(t *testing.T) {
	si := newStateInit()
	si.InitData = []byte{10, 20, 30}
	si.Salt = []byte("round_trip_salt")

	addr1, err := DeriveContractAddress(si)
	if err != nil {
		t.Fatalf("derive address: %v", err)
	}

	exported, _ := ExportStateInit(si)
	imported, _ := ImportStateInit(exported)

	addr2, err := DeriveContractAddress(imported)
	if err != nil {
		t.Fatalf("derive imported address: %v", err)
	}

	if addr1.Internal != addr2.Internal {
		t.Errorf("address not preserved after round trip: %s vs %s", addr1.Internal, addr2.Internal)
	}
}

func TestContractAddressFormat(t *testing.T) {
	si := newStateInit()
	addr, err := DeriveContractAddress(si)
	if err != nil {
		t.Fatalf("derive address: %v", err)
	}

	if len(addr.Internal) < 4 || addr.Internal[:2] != "4:" {
		t.Errorf("internal address should start with '4:', got %s", addr.Internal[:4])
	}

	if len(addr.External) < 3 || addr.External[:3] != "AE:" {
		t.Errorf("external address should start with 'AE:', got %s", addr.External[:3])
	}

	if addr.rawHash == [32]byte{} {
		t.Error("raw hash should not be zero")
	}
}

func TestDifferentNamespaceDifferentAddress(t *testing.T) {
	si1 := newStateInit()
	si1.Namespace = "ns1"
	si2 := newStateInit()
	si2.Namespace = "ns2"

	addr1, _ := DeriveContractAddress(si1)
	addr2, _ := DeriveContractAddress(si2)

	if addr1.Internal == addr2.Internal {
		t.Error("different namespaces should produce different addresses")
	}
}

func TestDependencyDAGValidation(t *testing.T) {
	si := newStateInit()
	si.DependencyHashes = [][32]byte{
		{1}, {2}, {1},
	}

	err := validateDependencyDAG(si.DependencyHashes)
	if err == nil {
		t.Error("duplicate dependencies should fail DAG validation")
	}
}

func TestBase58Encoding(t *testing.T) {
	input := big.NewInt(123456789).Bytes()
	result := base58Encode(input)
	if len(result) == 0 {
		t.Error("base58 encode should produce non-empty result")
	}
}

func TestDeployCapabilityMask(t *testing.T) {
	mask := AllDeployCapabilities
	if !mask.Has(DeployCapStorage) {
		t.Error("all capabilities should include storage")
	}
	if mask.IsEmpty() {
		t.Error("all capabilities should not be empty")
	}

	empty := DeployCapabilityMask{}
	if !empty.IsEmpty() {
		t.Error("zero flags should be empty")
	}
	if empty.Has(DeployCapStorage) {
		t.Error("empty mask should not have storage")
	}
}

func TestContractDeployStateStrings(t *testing.T) {
	tests := map[ContractDeployState]string{
		ContractNotDeployed:	"NOT_DEPLOYED",
		ContractDeployed:	"DEPLOYED",
		ContractInitialized:	"INITIALIZED",
	}
	for state, expected := range tests {
		if state.String() != expected {
			t.Errorf("expected %s, got %s", expected, state.String())
		}
	}
}
