package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
)

const (
	DefaultContractChainID		= "aetra"
	DefaultContractNamespace	= "contracts"
	DefaultInitialStorageRoot	= "0000000000000000000000000000000000000000000000000000000000000000"
	MaxContractSaltBytes		= 256
	MaxContractDependencies		= 32
)

type StateInit struct {
	ABIVersion		uint32
	CodeID			string
	CodeHash		string
	InitData		[]byte
	Salt			string
	SaltBytes		[]byte
	Owner			string
	Libraries		[]CodeDependency
	InitialStorageRoot	string
	InitialBalanceNAET	uint64
	Capabilities		[]string
}

type CodeDependency struct {
	CodeID		string
	CodeHash	string
}

func NewStateInit(owner, codeHash string, initData []byte, salt string, initialBalance uint64) StateInit {
	return StateInit{
		ABIVersion:		1,
		CodeID:			codeHash,
		CodeHash:		codeHash,
		InitData:		append([]byte(nil), initData...),
		Salt:			salt,
		Owner:			owner,
		InitialStorageRoot:	DefaultInitialStorageRoot,
		InitialBalanceNAET:	initialBalance,
	}
}

func (s StateInit) Normalize() StateInit {
	out := s
	out.CodeID = strings.TrimSpace(out.CodeID)
	out.CodeHash = strings.TrimSpace(out.CodeHash)
	out.InitData = append([]byte(nil), out.InitData...)
	out.Salt = strings.TrimSpace(out.Salt)
	out.SaltBytes = append([]byte(nil), out.SaltBytes...)
	out.Owner = strings.TrimSpace(out.Owner)
	out.InitialStorageRoot = strings.TrimSpace(out.InitialStorageRoot)
	if out.InitialStorageRoot == "" {
		out.InitialStorageRoot = DefaultInitialStorageRoot
	}
	out.Libraries = append([]CodeDependency(nil), out.Libraries...)
	for i := range out.Libraries {
		out.Libraries[i].CodeID = strings.TrimSpace(out.Libraries[i].CodeID)
		out.Libraries[i].CodeHash = strings.TrimSpace(out.Libraries[i].CodeHash)
	}
	sort.SliceStable(out.Libraries, func(i, j int) bool {
		if out.Libraries[i].CodeID != out.Libraries[j].CodeID {
			return out.Libraries[i].CodeID < out.Libraries[j].CodeID
		}
		return out.Libraries[i].CodeHash < out.Libraries[j].CodeHash
	})
	out.Capabilities = append([]string(nil), out.Capabilities...)
	for i := range out.Capabilities {
		out.Capabilities[i] = strings.TrimSpace(out.Capabilities[i])
	}
	sort.Strings(out.Capabilities)
	dedup := out.Capabilities[:0]
	for _, capability := range out.Capabilities {
		if capability == "" {
			continue
		}
		if len(dedup) == 0 || dedup[len(dedup)-1] != capability {
			dedup = append(dedup, capability)
		}
	}
	out.Capabilities = append([]string(nil), dedup...)
	return out
}

func (s StateInit) IsZero() bool {
	s = s.Normalize()
	return s.ABIVersion == 0 &&
		s.CodeID == "" &&
		s.CodeHash == "" &&
		len(s.InitData) == 0 &&
		s.Salt == "" &&
		len(s.SaltBytes) == 0 &&
		s.Owner == "" &&
		len(s.Libraries) == 0 &&
		s.InitialStorageRoot == DefaultInitialStorageRoot &&
		s.InitialBalanceNAET == 0 &&
		len(s.Capabilities) == 0
}

func (s StateInit) Validate(params Params) error {
	s = s.Normalize()
	if s.ABIVersion == 0 {
		return errors.New("state init ABI version must be positive")
	}
	if strings.TrimSpace(s.CodeID) == "" {
		return errors.New("state init code id is required")
	}
	if err := validateHashText("state init code hash", s.CodeHash); err != nil {
		return err
	}
	if uint64(len(s.InitData)) > params.MaxInitDataBytes {
		return errors.New("state init data exceeds maximum size")
	}
	if len(s.SaltBytesForAddress()) > int(params.MaxStateInitSaltBytes) {
		return errors.New("state init salt exceeds maximum size")
	}
	if uint32(len(s.Libraries)) > params.MaxStateInitDependencies {
		return errors.New("state init dependency count exceeds maximum")
	}
	if err := ValidateUserFacingAEAddress("state init owner", s.Owner); err != nil {
		return err
	}
	if err := validateHashText("state init storage root", s.InitialStorageRoot); err != nil {
		return err
	}
	seen := map[string]struct{}{}
	seenCodeID := map[string]struct{}{}
	for _, dep := range s.Libraries {
		if dep.CodeID == "" {
			return errors.New("state init dependency code id is required")
		}
		if err := validateHashText("state init dependency code hash", dep.CodeHash); err != nil {
			return err
		}
		key := dep.CodeID + "/" + dep.CodeHash
		if _, found := seen[key]; found {
			return errors.New("duplicate state init dependency")
		}
		if _, found := seenCodeID[dep.CodeID]; found {
			return errors.New("duplicate state init dependency code id")
		}
		seen[key] = struct{}{}
		seenCodeID[dep.CodeID] = struct{}{}
	}
	return nil
}

func (s StateInit) SaltBytesForAddress() []byte {
	s = s.Normalize()
	if len(s.SaltBytes) > 0 {
		return append([]byte(nil), s.SaltBytes...)
	}
	return []byte(s.Salt)
}

func (s StateInit) InitDataHash() string {
	sum := sha256.Sum256(s.Normalize().InitData)
	return hex.EncodeToString(sum[:])
}

func CanonicalStateInitBytes(stateInit StateInit) ([]byte, error) {
	s := stateInit.Normalize()
	var buf bytes.Buffer
	buf.WriteString("aetra-state-init-v1")
	writeU32(&buf, s.ABIVersion)
	writeString(&buf, s.CodeID)
	writeString(&buf, s.CodeHash)
	writeBytes(&buf, s.InitData)
	writeString(&buf, s.Salt)
	writeBytes(&buf, s.SaltBytes)
	writeString(&buf, s.Owner)
	writeString(&buf, s.InitialStorageRoot)
	writeU64(&buf, s.InitialBalanceNAET)
	writeU32(&buf, uint32(len(s.Libraries)))
	for _, dep := range s.Libraries {
		writeString(&buf, dep.CodeID)
		writeString(&buf, dep.CodeHash)
	}
	writeU32(&buf, uint32(len(s.Capabilities)))
	for _, capability := range s.Capabilities {
		writeString(&buf, capability)
	}
	return buf.Bytes(), nil
}

func HashStateInit(stateInit StateInit) (string, error) {
	encoded, err := CanonicalStateInitBytes(stateInit)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:]), nil
}

func DeriveContractAddress(chainID, namespace, deployer, codeHash, initDataHash string, salt []byte) (string, string, error) {
	chainID = strings.TrimSpace(chainID)
	namespace = strings.TrimSpace(namespace)
	if chainID == "" {
		chainID = DefaultContractChainID
	}
	if namespace == "" {
		namespace = DefaultContractNamespace
	}
	if err := ValidateUserFacingAEAddress("contract deployer", deployer); err != nil {
		return "", "", err
	}
	rawDeployer, err := RawAddressForUserAddress(deployer)
	if err != nil {
		return "", "", err
	}
	if rawDeployer == addressing.ZeroRawAddress {
		return "", "", errors.New("contract deployer must not be zero address")
	}
	if err := validateHashText("contract code hash", codeHash); err != nil {
		return "", "", err
	}
	if err := validateHashText("contract init data hash", initDataHash); err != nil {
		return "", "", err
	}
	var buf bytes.Buffer
	buf.WriteString("aetra-contract-address-v2")
	writeString(&buf, chainID)
	writeString(&buf, namespace)
	writeString(&buf, deployer)
	writeString(&buf, codeHash)
	writeString(&buf, initDataHash)
	writeBytes(&buf, append([]byte(nil), salt...))
	sum := sha256.Sum256(buf.Bytes())
	user, err := addressing.FormatUserFriendly(sum[:])
	if err != nil {
		return "", "", err
	}
	return user, addressing.Format(sum[:]), nil
}

func DeriveContractAddressFromStateInit(chainID, namespace, deployer string, stateInit StateInit, params Params) (string, string, error) {
	if err := stateInit.Validate(params); err != nil {
		return "", "", err
	}
	normalized := stateInit.Normalize()
	return DeriveContractAddress(chainID, namespace, deployer, normalized.CodeHash, normalized.InitDataHash(), normalized.SaltBytesForAddress())
}

func writeU32(buf *bytes.Buffer, value uint32) {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], value)
	buf.Write(b[:])
}

func writeU64(buf *bytes.Buffer, value uint64) {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], value)
	buf.Write(b[:])
}

func writeString(buf *bytes.Buffer, value string) {
	writeBytes(buf, []byte(value))
}

func writeBytes(buf *bytes.Buffer, value []byte) {
	if uint64(len(value)) > uint64(^uint32(0)) {
		panic(fmt.Sprintf("canonical field too large: %d", len(value)))
	}
	writeU32(buf, uint32(len(value)))
	buf.Write(value)
}

func sha256Sum(value []byte) []byte {
	sum := sha256.Sum256(value)
	return sum[:]
}
