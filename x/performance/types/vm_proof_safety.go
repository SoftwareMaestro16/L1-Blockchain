package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	ProofSafetyVersionV1 = uint32(1)
)

type VMRemoteMutationMode string

const (
	VMRemoteMutationNone		VMRemoteMutationMode	= "none"
	VMRemoteMutationAsyncMessage	VMRemoteMutationMode	= "async_message"
	VMRemoteMutationSynchronous	VMRemoteMutationMode	= "synchronous"
)

type VMPromiseTimeoutPolicy struct {
	CreatedHeight	uint64
	TimeoutHeight	uint64
	DelayBlocks	uint64
	PolicyHash	string
}

type VMSafetyEnvelope struct {
	ExecutionID			string
	ZoneID				string
	ShardID				string
	ContractAddress			string
	GasMeteringEnabled		bool
	GasLimit			uint64
	GasConsumed			uint64
	StorageIterationLimit		uint32
	StorageIterations		uint32
	ProofVerificationCount		uint32
	ProofVerificationGasReserved	uint64
	ProofVerificationGasConsumed	uint64
	MessageCreationCount		uint32
	ReservedForwardingFee		string
	ForwardingFeeRequired		string
	RemoteMutationMode		VMRemoteMutationMode
	PromiseTimeout			VMPromiseTimeoutPolicy
	SafetyHash			string
}

type ProofRootType string

const (
	ProofRootAccount	ProofRootType	= "account"
	ProofRootMessage	ProofRootType	= "message"
	ProofRootZone		ProofRootType	= "zone"
	ProofRootIdentity	ProofRootType	= "identity"
	ProofRootResolver	ProofRootType	= "resolver"
	ProofRootContract	ProofRootType	= "contract"
	ProofRootStorage	ProofRootType	= "storage"
)

type TrustedHeaderBinding struct {
	Height		uint64
	HeaderHash	string
	AppHash		string
}

type UniversalProofSafetyEnvelope struct {
	ProofID			string
	ProofVersion		uint32
	SupportedVersions	[]uint32
	TrustedHeader		TrustedHeaderBinding
	ProofHeight		uint64
	ZoneID			string
	ShardID			string
	ObjectKey		string
	RootType		ProofRootType
	RootHash		string
	NonExistence		bool
	ExistenceProofHash	string
	AbsenceProofHash	string
	ProofHash		string
}

func BuildVMSafetyEnvelope(envelope VMSafetyEnvelope) (VMSafetyEnvelope, error) {
	envelope = envelope.Normalize()
	envelope.SafetyHash = ComputeVMSafetyEnvelopeHash(envelope)
	return envelope, envelope.Validate()
}

func (e VMSafetyEnvelope) Normalize() VMSafetyEnvelope {
	e.ExecutionID = strings.TrimSpace(e.ExecutionID)
	e.ZoneID = strings.TrimSpace(e.ZoneID)
	e.ShardID = strings.TrimSpace(e.ShardID)
	e.ContractAddress = strings.TrimSpace(e.ContractAddress)
	e.ReservedForwardingFee = strings.TrimSpace(e.ReservedForwardingFee)
	e.ForwardingFeeRequired = strings.TrimSpace(e.ForwardingFeeRequired)
	e.RemoteMutationMode = VMRemoteMutationMode(strings.TrimSpace(string(e.RemoteMutationMode)))
	e.PromiseTimeout = e.PromiseTimeout.Normalize()
	e.SafetyHash = normalizeLowerHex(e.SafetyHash)
	return e
}

func (e VMSafetyEnvelope) Validate() error {
	envelope := e.Normalize()
	if err := validateExecutionToken("VM safety execution id", envelope.ExecutionID); err != nil {
		return err
	}
	if err := validateExecutionToken("VM safety zone id", envelope.ZoneID); err != nil {
		return err
	}
	if err := validateExecutionToken("VM safety shard id", envelope.ShardID); err != nil {
		return err
	}
	if err := validateStoreV2Path("VM safety contract address", envelope.ContractAddress); err != nil {
		return err
	}
	if !envelope.GasMeteringEnabled {
		return errors.New("VM safety gas metering is mandatory")
	}
	if envelope.GasLimit == 0 {
		return errors.New("VM safety gas limit must be positive")
	}
	if envelope.GasConsumed > envelope.GasLimit {
		return errors.New("VM safety gas consumed exceeds limit")
	}
	if envelope.StorageIterationLimit == 0 {
		return errors.New("VM safety storage iteration limit must be positive")
	}
	if envelope.StorageIterations > envelope.StorageIterationLimit {
		return errors.New("VM safety storage iteration is unbounded")
	}
	if envelope.ProofVerificationCount > 0 {
		if envelope.ProofVerificationGasReserved == 0 {
			return errors.New("VM safety proof verification cost must be metered")
		}
		if envelope.ProofVerificationGasConsumed > envelope.ProofVerificationGasReserved {
			return errors.New("VM safety proof verification gas is under-metered")
		}
	}
	if envelope.MessageCreationCount > 0 {
		reserved, err := parsePerformanceNonNegativeInt("VM safety reserved forwarding fee", envelope.ReservedForwardingFee)
		if err != nil {
			return err
		}
		required, err := parsePerformanceNonNegativeInt("VM safety required forwarding fee", envelope.ForwardingFeeRequired)
		if err != nil {
			return err
		}
		if reserved.LT(required) || required.IsZero() {
			return errors.New("VM safety message creation requires reserved forwarding fee")
		}
	}
	if !IsVMRemoteMutationMode(envelope.RemoteMutationMode) {
		return errors.New("VM safety remote mutation mode is unsupported")
	}
	if envelope.RemoteMutationMode == VMRemoteMutationSynchronous {
		return errors.New("VM safety contract cannot synchronously mutate remote zone state")
	}
	if err := envelope.PromiseTimeout.Validate(); err != nil {
		return err
	}
	if envelope.SafetyHash != ComputeVMSafetyEnvelopeHash(envelope) {
		return errors.New("VM safety envelope hash mismatch")
	}
	return nil
}

func (p VMPromiseTimeoutPolicy) Normalize() VMPromiseTimeoutPolicy {
	p.PolicyHash = normalizeLowerHex(p.PolicyHash)
	return p
}

func (p VMPromiseTimeoutPolicy) Validate() error {
	policy := p.Normalize()
	if policy.CreatedHeight == 0 || policy.TimeoutHeight == 0 || policy.DelayBlocks == 0 {
		return errors.New("VM safety promise timeout heights and delay must be positive")
	}
	if policy.TimeoutHeight != policy.CreatedHeight+policy.DelayBlocks {
		return errors.New("VM safety promise timeout must be deterministic")
	}
	if policy.PolicyHash != ComputeVMPromiseTimeoutPolicyHash(policy) {
		return errors.New("VM safety promise timeout policy hash mismatch")
	}
	return nil
}

func IsVMRemoteMutationMode(mode VMRemoteMutationMode) bool {
	switch mode {
	case VMRemoteMutationNone, VMRemoteMutationAsyncMessage, VMRemoteMutationSynchronous:
		return true
	default:
		return false
	}
}

func BuildUniversalProofSafetyEnvelope(envelope UniversalProofSafetyEnvelope) (UniversalProofSafetyEnvelope, error) {
	envelope = envelope.Normalize()
	envelope.ProofHash = ComputeUniversalProofSafetyHash(envelope)
	return envelope, envelope.Validate()
}

func (e UniversalProofSafetyEnvelope) Normalize() UniversalProofSafetyEnvelope {
	e.ProofID = strings.TrimSpace(e.ProofID)
	e.SupportedVersions = append([]uint32(nil), e.SupportedVersions...)
	sort.SliceStable(e.SupportedVersions, func(i, j int) bool {
		return e.SupportedVersions[i] < e.SupportedVersions[j]
	})
	e.TrustedHeader = e.TrustedHeader.Normalize()
	e.ZoneID = strings.TrimSpace(e.ZoneID)
	e.ShardID = strings.TrimSpace(e.ShardID)
	e.ObjectKey = strings.TrimSpace(e.ObjectKey)
	e.RootType = ProofRootType(strings.TrimSpace(string(e.RootType)))
	e.RootHash = normalizeLowerHex(e.RootHash)
	e.ExistenceProofHash = normalizeLowerHex(e.ExistenceProofHash)
	e.AbsenceProofHash = normalizeLowerHex(e.AbsenceProofHash)
	e.ProofHash = normalizeLowerHex(e.ProofHash)
	return e
}

func (e UniversalProofSafetyEnvelope) Validate() error {
	envelope := e.Normalize()
	if err := validateExecutionToken("proof safety proof id", envelope.ProofID); err != nil {
		return err
	}
	if envelope.ProofVersion == 0 {
		return errors.New("proof safety proof version must be positive")
	}
	if !supportsProofVersion(envelope.SupportedVersions, envelope.ProofVersion) {
		return errors.New("proof safety unsupported proof version")
	}
	if err := envelope.TrustedHeader.Validate(); err != nil {
		return err
	}
	if envelope.ProofHeight == 0 {
		return errors.New("proof safety proof height must be positive")
	}
	if envelope.ProofHeight != envelope.TrustedHeader.Height {
		return errors.New("proof safety proof must bind to trusted header height")
	}
	if err := validateExecutionToken("proof safety zone id", envelope.ZoneID); err != nil {
		return err
	}
	if err := validateExecutionToken("proof safety shard id", envelope.ShardID); err != nil {
		return err
	}
	if err := validateStoreV2Path("proof safety object key", envelope.ObjectKey); err != nil {
		return err
	}
	if !IsProofRootType(envelope.RootType) {
		return errors.New("proof safety root type is unsupported")
	}
	if err := validateHexHash("proof safety root hash", envelope.RootHash); err != nil {
		return err
	}
	if envelope.NonExistence {
		if envelope.AbsenceProofHash == "" {
			return errors.New("proof safety non-existence proofs must be explicit")
		}
		if envelope.ExistenceProofHash != "" {
			return errors.New("proof safety non-existence proof must not include existence proof")
		}
		if err := validateHexHash("proof safety absence proof hash", envelope.AbsenceProofHash); err != nil {
			return err
		}
	} else {
		if envelope.ExistenceProofHash == "" {
			return errors.New("proof safety existence proof hash is required")
		}
		if err := validateHexHash("proof safety existence proof hash", envelope.ExistenceProofHash); err != nil {
			return err
		}
	}
	if envelope.ProofHash != ComputeUniversalProofSafetyHash(envelope) {
		return errors.New("proof safety proof hash mismatch")
	}
	return nil
}

func (h TrustedHeaderBinding) Normalize() TrustedHeaderBinding {
	h.HeaderHash = normalizeLowerHex(h.HeaderHash)
	h.AppHash = normalizeLowerHex(h.AppHash)
	return h
}

func (h TrustedHeaderBinding) Validate() error {
	header := h.Normalize()
	if header.Height == 0 {
		return errors.New("proof safety trusted header height must be positive")
	}
	if err := validateHexHash("proof safety trusted header hash", header.HeaderHash); err != nil {
		return err
	}
	return validateHexHash("proof safety trusted app hash", header.AppHash)
}

func IsProofRootType(rootType ProofRootType) bool {
	switch rootType {
	case ProofRootAccount, ProofRootMessage, ProofRootZone, ProofRootIdentity, ProofRootResolver, ProofRootContract, ProofRootStorage:
		return true
	default:
		return false
	}
}

func ComputeVMSafetyEnvelopeHash(envelope VMSafetyEnvelope) string {
	envelope = envelope.Normalize()
	return hashStrings(
		"vm-safety-envelope",
		envelope.ExecutionID,
		envelope.ZoneID,
		envelope.ShardID,
		envelope.ContractAddress,
		fmt.Sprintf("%t", envelope.GasMeteringEnabled),
		fmt.Sprintf("%020d", envelope.GasLimit),
		fmt.Sprintf("%020d", envelope.GasConsumed),
		fmt.Sprintf("%020d", uint64(envelope.StorageIterationLimit)),
		fmt.Sprintf("%020d", uint64(envelope.StorageIterations)),
		fmt.Sprintf("%020d", uint64(envelope.ProofVerificationCount)),
		fmt.Sprintf("%020d", envelope.ProofVerificationGasReserved),
		fmt.Sprintf("%020d", envelope.ProofVerificationGasConsumed),
		fmt.Sprintf("%020d", uint64(envelope.MessageCreationCount)),
		envelope.ReservedForwardingFee,
		envelope.ForwardingFeeRequired,
		string(envelope.RemoteMutationMode),
		envelope.PromiseTimeout.PolicyHash,
	)
}

func ComputeVMPromiseTimeoutPolicyHash(policy VMPromiseTimeoutPolicy) string {
	policy = policy.Normalize()
	return hashStrings(
		"vm-promise-timeout-policy",
		fmt.Sprintf("%020d", policy.CreatedHeight),
		fmt.Sprintf("%020d", policy.TimeoutHeight),
		fmt.Sprintf("%020d", policy.DelayBlocks),
	)
}

func ComputeUniversalProofSafetyHash(envelope UniversalProofSafetyEnvelope) string {
	envelope = envelope.Normalize()
	parts := []string{
		"universal-proof-safety",
		envelope.ProofID,
		fmt.Sprintf("%020d", uint64(envelope.ProofVersion)),
		fmt.Sprintf("%020d", envelope.ProofHeight),
		envelope.TrustedHeader.HeaderHash,
		envelope.TrustedHeader.AppHash,
		envelope.ZoneID,
		envelope.ShardID,
		envelope.ObjectKey,
		string(envelope.RootType),
		envelope.RootHash,
		fmt.Sprintf("%t", envelope.NonExistence),
		envelope.ExistenceProofHash,
		envelope.AbsenceProofHash,
	}
	for _, version := range envelope.SupportedVersions {
		parts = append(parts, fmt.Sprintf("%020d", uint64(version)))
	}
	return hashStrings(parts...)
}

func supportsProofVersion(supported []uint32, version uint32) bool {
	if len(supported) == 0 {
		supported = []uint32{ProofSafetyVersionV1}
	}
	for _, candidate := range supported {
		if candidate == version {
			return true
		}
	}
	return false
}
