package avm

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"

	"github.com/sovereign-l1/l1/x/aetravm/chunk"
	"lukechampine.com/blake3"
)

const (
	MaxInitDataSize		uint32	= 1 << 20
	MaxSaltSize		uint32	= 64
	MaxDependencyCount	uint32	= 256
	MaxStateInitSize	uint32	= 2 << 20
	StateInitABIVersion	uint32	= 1
)

type DeployCapFlag uint64

const (
	DeployCapStorage	DeployCapFlag	= 1 << iota
	DeployCapMessaging
	DeployCapHostFunc
	DeployCapQueryOnly
	DeployCapMintCoins
	DeployCapUpgrade
)

type DeployCapabilityMask struct {
	Flags uint64
}

func (c DeployCapabilityMask) Has(flag DeployCapFlag) bool {
	return c.Flags&uint64(flag) != 0
}

func (c DeployCapabilityMask) IsEmpty() bool {
	return c.Flags == 0
}

var AllDeployCapabilities = DeployCapabilityMask{Flags: ^uint64(0)}

type ContractDeployState uint8

const (
	ContractNotDeployed	ContractDeployState	= iota
	ContractDeployed
	ContractInitialized
)

func (s ContractDeployState) String() string {
	switch s {
	case ContractNotDeployed:
		return "NOT_DEPLOYED"
	case ContractDeployed:
		return "DEPLOYED"
	case ContractInitialized:
		return "INITIALIZED"
	default:
		return "UNKNOWN"
	}
}

// ---------------
// StateInit — Canonical Deployment Descriptor
// ---------------
// Any change in ANY field MUST change StateInitHash and ContractAddress.
type StateInit struct {
	ABIVersion		uint32
	CodeHash		[32]byte
	InitData		[]byte
	Salt			[]byte
	DeployerAddress		string
	ChainID			string
	Namespace		string
	DependencyHashes	[][32]byte
	InitialStateRoot	*chunk.Chunk
	InitialBalance		uint64
	Capabilities		DeployCapabilityMask
}

var (
	ErrZeroDeployer		= errors.New("AVM: deployer address must not be zero/empty")
	ErrEmptyCodeHash	= errors.New("AVM: code hash must not be empty")
	ErrInitDataTooLarge	= errors.New("AVM: init data exceeds maximum size")
	ErrSaltTooLarge		= errors.New("AVM: salt exceeds maximum size")
	ErrTooManyDeps		= errors.New("AVM: too many dependencies")
	ErrInvalidABI		= errors.New("AVM: invalid ABI version")
	ErrEmptyChainID		= errors.New("AVM: chain ID must not be empty")
	ErrStateInitTooLarge	= errors.New("AVM: StateInit encoded size exceeds maximum")
	ErrDuplicateDeployment	= errors.New("AVM: duplicate deployment")
)

// Validate performs validation before deployment.
func (si *StateInit) Validate() error {
	if si.DeployerAddress == "" {
		return ErrZeroDeployer
	}
	if si.CodeHash == [32]byte{} {
		return ErrEmptyCodeHash
	}
	if len(si.InitData) > int(MaxInitDataSize) {
		return ErrInitDataTooLarge
	}
	if len(si.Salt) > int(MaxSaltSize) {
		return ErrSaltTooLarge
	}
	if len(si.DependencyHashes) > int(MaxDependencyCount) {
		return ErrTooManyDeps
	}
	if si.ABIVersion == 0 || si.ABIVersion > 1000 {
		return ErrInvalidABI
	}
	if si.ChainID == "" {
		return ErrEmptyChainID
	}
	encoded, err := si.CanonicalEncode()
	if err != nil {
		return fmt.Errorf("AVM: encode for size check: %w", err)
	}
	if uint32(len(encoded)) > MaxStateInitSize {
		return ErrStateInitTooLarge
	}
	return nil
}

func (si *StateInit) CanonicalEncode() ([]byte, error) {
	buf := make([]byte, 0, 512)

	buf = binary.BigEndian.AppendUint32(buf, si.ABIVersion)
	buf = append(buf, si.CodeHash[:]...)

	buf = binary.BigEndian.AppendUint32(buf, uint32(len(si.InitData)))
	buf = append(buf, si.InitData...)

	buf = binary.BigEndian.AppendUint32(buf, uint32(len(si.Salt)))
	buf = append(buf, si.Salt...)

	deployerBytes := []byte(si.DeployerAddress)
	buf = binary.BigEndian.AppendUint32(buf, uint32(len(deployerBytes)))
	buf = append(buf, deployerBytes...)

	chainIDBytes := []byte(si.ChainID)
	buf = binary.BigEndian.AppendUint32(buf, uint32(len(chainIDBytes)))
	buf = append(buf, chainIDBytes...)

	nsBytes := []byte(si.Namespace)
	buf = binary.BigEndian.AppendUint32(buf, uint32(len(nsBytes)))
	buf = append(buf, nsBytes...)

	buf = binary.BigEndian.AppendUint32(buf, uint32(len(si.DependencyHashes)))
	for _, dep := range si.DependencyHashes {
		buf = append(buf, dep[:]...)
	}

	if si.InitialStateRoot != nil {
		buf = append(buf, si.InitialStateRoot.Hash()...)
	} else {
		buf = append(buf, make([]byte, 32)...)
	}

	buf = binary.BigEndian.AppendUint64(buf, si.InitialBalance)
	buf = binary.BigEndian.AppendUint64(buf, si.Capabilities.Flags)

	return buf, nil
}

func HashStateInit(si *StateInit) ([32]byte, error) {
	encoded, err := si.CanonicalEncode()
	if err != nil {
		return [32]byte{}, fmt.Errorf("AVM: hash StateInit: %w", err)
	}
	return blake3.Sum256(encoded), nil
}

type ContractAddress struct {
	Internal	string
	External	string
	rawHash		[32]byte
}

func (a ContractAddress) RawHash() [32]byte	{ return a.rawHash }

func DeriveContractAddress(si *StateInit) (*ContractAddress, error) {
	if err := si.Validate(); err != nil {
		return nil, fmt.Errorf("AVM: derive address: %w", err)
	}

	stateInitHash, err := HashStateInit(si)
	if err != nil {
		return nil, fmt.Errorf("AVM: derive address hash: %w", err)
	}

	h := blake3.New(32, nil)
	h.Write([]byte(si.ChainID))
	h.Write([]byte{0x00})
	h.Write([]byte(si.Namespace))
	h.Write([]byte{0x00})
	h.Write([]byte(si.DeployerAddress))
	h.Write([]byte{0x00})
	h.Write(si.CodeHash[:])
	h.Write(stateInitHash[:])
	h.Write(si.Salt)

	var rawHash [32]byte
	copy(rawHash[:], h.Sum(nil))

	internal := fmt.Sprintf("4:%x", rawHash)
	external := fmt.Sprintf("AE:%s", base58Encode(rawHash[:]))

	return &ContractAddress{
		Internal:	internal,
		External:	external,
		rawHash:	rawHash,
	}, nil
}

type CounterfactualState struct {
	Address		ContractAddress
	State		ContractDeployState
	InitData	*StateInit
}

func QueryContractState(addr ContractAddress, deployed bool, initialized bool) CounterfactualState {
	state := ContractNotDeployed
	if initialized {
		state = ContractInitialized
	} else if deployed {
		state = ContractDeployed
	}
	return CounterfactualState{
		Address:	addr,
		State:		state,
	}
}

type DeploymentResult struct {
	Address		ContractAddress
	StateInitHash	[32]byte
	State		ContractDeployState
	Error		error
}

func DeployContract(si *StateInit, existingAddresses map[string]ContractDeployState) (*DeploymentResult, error) {
	if err := si.Validate(); err != nil {
		return nil, err
	}

	addr, err := DeriveContractAddress(si)
	if err != nil {
		return nil, err
	}

	if state, ok := existingAddresses[addr.Internal]; ok {
		if state == ContractDeployed || state == ContractInitialized {
			return &DeploymentResult{
				Address:	*addr,
				State:		state,
				Error:		ErrDuplicateDeployment,
			}, nil
		}
	}

	if err := validateDependencyDAG(si.DependencyHashes); err != nil {
		return nil, fmt.Errorf("AVM: dependency validation: %w", err)
	}

	stateInitHash, err := HashStateInit(si)
	if err != nil {
		return nil, err
	}

	return &DeploymentResult{
		Address:	*addr,
		StateInitHash:	stateInitHash,
		State:		ContractDeployed,
	}, nil
}

func validateDependencyDAG(deps [][32]byte) error {
	seen := make(map[[32]byte]bool)
	for _, dep := range deps {
		if seen[dep] {
			return fmt.Errorf("AVM: duplicate dependency %x", dep[:8])
		}
		seen[dep] = true
	}
	return nil
}

func ExportStateInit(si *StateInit) ([]byte, error) {
	return si.CanonicalEncode()
}

func ImportStateInit(data []byte) (*StateInit, error) {
	if len(data) < 4+32+4 {
		return nil, fmt.Errorf("AVM: data too short for StateInit")
	}

	offset := 0

	abiVersion := binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4

	var codeHash [32]byte
	copy(codeHash[:], data[offset:offset+32])
	offset += 32

	initDataLen := binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4
	if uint32(len(data)-offset) < initDataLen {
		return nil, fmt.Errorf("AVM: truncated init data")
	}
	initData := make([]byte, initDataLen)
	copy(initData, data[offset:offset+int(initDataLen)])
	offset += int(initDataLen)

	saltLen := binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4
	if uint32(len(data)-offset) < saltLen {
		return nil, fmt.Errorf("AVM: truncated salt")
	}
	salt := make([]byte, saltLen)
	copy(salt, data[offset:offset+int(saltLen)])
	offset += int(saltLen)

	deployerLen := binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4
	if uint32(len(data)-offset) < deployerLen {
		return nil, fmt.Errorf("AVM: truncated deployer address")
	}
	deployer := string(data[offset : offset+int(deployerLen)])
	offset += int(deployerLen)

	chainIDLen := binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4
	if uint32(len(data)-offset) < chainIDLen {
		return nil, fmt.Errorf("AVM: truncated chain ID")
	}
	chainID := string(data[offset : offset+int(chainIDLen)])
	offset += int(chainIDLen)

	nsLen := binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4
	if uint32(len(data)-offset) < nsLen {
		return nil, fmt.Errorf("AVM: truncated namespace")
	}
	namespace := string(data[offset : offset+int(nsLen)])
	offset += int(nsLen)

	depCount := binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4
	if depCount > MaxDependencyCount {
		return nil, fmt.Errorf("AVM: too many dependencies in import: %d", depCount)
	}
	deps := make([][32]byte, depCount)
	for i := uint32(0); i < depCount; i++ {
		if len(data)-offset < 32 {
			return nil, fmt.Errorf("AVM: truncated dependency hash %d", i)
		}
		copy(deps[i][:], data[offset:offset+32])
		offset += 32
	}

	var stateRootHash [32]byte
	copy(stateRootHash[:], data[offset:offset+32])
	offset += 32

	initialBalance := binary.BigEndian.Uint64(data[offset : offset+8])
	offset += 8

	capFlags := binary.BigEndian.Uint64(data[offset : offset+8])
	offset += 8

	var initialStateRoot *chunk.Chunk
	if stateRootHash != [32]byte{} {
		b := chunk.NewBuilder()
		b.SetTypeTag(chunk.TypeSystem)
		b.SetData(stateRootHash[:], 256)
		initialStateRoot, _ = b.Build()
	}

	return &StateInit{
		ABIVersion:		abiVersion,
		CodeHash:		codeHash,
		InitData:		initData,
		Salt:			salt,
		DeployerAddress:	deployer,
		ChainID:		chainID,
		Namespace:		namespace,
		DependencyHashes:	deps,
		InitialStateRoot:	initialStateRoot,
		InitialBalance:		initialBalance,
		Capabilities:		DeployCapabilityMask{Flags: capFlags},
	}, nil
}

var base58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

func base58Encode(input []byte) string {
	leadingZeros := 0
	for _, b := range input {
		if b == 0 {
			leadingZeros++
		} else {
			break
		}
	}

	num := new(big.Int).SetBytes(input)
	base := new(big.Int).SetUint64(58)
	zero := new(big.Int)
	mod := new(big.Int)

	var result []byte
	for num.Cmp(zero) > 0 {
		num.DivMod(num, base, mod)
		result = append(result, base58Alphabet[mod.Uint64()])
	}

	for i := 0; i < leadingZeros; i++ {
		result = append(result, '1')
	}

	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return string(result)
}
