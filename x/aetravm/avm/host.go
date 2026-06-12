package avm

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"sort"

	"github.com/sovereign-l1/l1/x/aetravm/async"
)

const (
	MaxHostArgs	= 8
	MaxHostArgBytes	= 4096

	ExitCodeForbiddenHostCall	= async.ResultForbiddenHostCall
)

// HostGasCost specifies the gas cost model for a host function.
// Each function has a base cost plus per-byte and per-ref (Chunk reference) costs.
type HostGasCost struct {
	BaseCost	uint64
	PerByte		uint64
	PerRef		uint64
	MinCost		uint64
}

// HostGasCostFor returns the gas cost for calling a host function with given inputs.
func HostGasCostFor(host HostFunction, argBytes int, refCount int) uint64 {
	costs := DefaultHostFunctionCosts()
	perByteCost := hostPerByteCosts()
	perRefCost := hostPerRefCosts()

	cost := costs[host]
	cost += uint64(argBytes) * perByteCost[host]
	cost += uint64(refCount) * perRefCost[host]

	minCost := hostMinCosts()
	if min, ok := minCost[host]; ok && cost < min {
		return min
	}
	return cost
}

func hostPerByteCosts() map[HostFunction]uint64 {
	return map[HostFunction]uint64{
		HostSendInternal:	2,
		HostEmitEvent:		1,
		HostWriteStorage:	1,
		HostReadStorage:	1,
	}
}

func hostPerRefCosts() map[HostFunction]uint64 {
	return map[HostFunction]uint64{
		HostReadStorage:	10,
		HostWriteStorage:	25,
	}
}

func hostMinCosts() map[HostFunction]uint64 {
	return map[HostFunction]uint64{
		HostHashSHA256:		70,
		HostHashBLAKE3:		80,
		HostVerifyEd25519:	3_000,
		HostSendInternal:	200,
		HostWriteStorage:	50,
		HostReadStorage:	20,
		HostDeleteStorage:	150,
		HostParseAetraAddress:	30,
		HostFormatAetraAddress:	30,
	}
}

type HostFunctionSpec struct {
	ID		HostFunction
	Name		string
	MinArgs		uint16
	MaxArgs		uint16
	MaxArgBytes	uint32
	Forbidden	bool
	Class		HostFunctionClass
}

const (
	HostHashSHA256		HostFunction	= 101
	HostHashBLAKE3		HostFunction	= 102
	HostVerifyEd25519	HostFunction	= 103
	HostParseAetraAddress	HostFunction	= 104
	HostFormatAetraAddress	HostFunction	= 105
	HostDeleteStorage	HostFunction	= 106
	HostEmitEvent		HostFunction	= 107
	HostSendInternal	HostFunction	= 108
	HostGetBlockHeight	HostFunction	= 109
	HostGetChainID		HostFunction	= 110
	HostGetContractAddress	HostFunction	= 111
	HostGetCallerSource	HostFunction	= 112
	HostGetAttachedValue	HostFunction	= 113
	HostAbort		HostFunction	= 114

	HostWallClockTime	HostFunction	= 0xf001
	HostRandomness		HostFunction	= 0xf002
	HostFilesystem		HostFunction	= 0xf003
	HostNetwork		HostFunction	= 0xf004
	HostFloatingPoint	HostFunction	= 0xf005
	HostGoroutine		HostFunction	= 0xf006
	HostProcessEnv		HostFunction	= 0xf007
	HostNondeterministicMap	HostFunction	= 0xf008
)

func DefaultHostFunctionCosts() map[HostFunction]uint64 {
	return map[HostFunction]uint64{
		HostReadStorage:	25,
		HostWriteStorage:	60,
		HostEmitInternal:	150,
		HostInspectMsg:		10,
		HostBlockContext:	10,
		HostChargeGas:		1,
		HostReturn:		1,
		HostScheduleSelf:	150,
		HostHashSHA256:		80,
		HostHashBLAKE3:		90,
		HostVerifyEd25519:	5_000,
		HostParseAetraAddress:	40,
		HostFormatAetraAddress:	40,
		HostDeleteStorage:	200,
		HostEmitEvent:		100,
		HostSendInternal:	250,
		HostGetBlockHeight:	10,
		HostGetChainID:		10,
		HostGetContractAddress:	10,
		HostGetCallerSource:	10,
		HostGetAttachedValue:	10,
		HostAbort:		1,
	}
}

func HostFunctionRegistry() map[HostFunction]HostFunctionSpec {
	specs := []HostFunctionSpec{
		{ID: HostReadStorage, Name: "read_storage", MinArgs: 1, MaxArgs: 1, MaxArgBytes: MaxKeySize, Class: ClassEffectful},
		{ID: HostWriteStorage, Name: "write_storage", MinArgs: 2, MaxArgs: 2, MaxArgBytes: MaxHostArgBytes, Class: ClassEffectful},
		{ID: HostEmitInternal, Name: "emit_internal", MinArgs: 1, MaxArgs: 1, MaxArgBytes: MaxHostArgBytes, Class: ClassEffectful},
		{ID: HostInspectMsg, Name: "inspect_message", MinArgs: 0, MaxArgs: 0, MaxArgBytes: 0, Class: ClassPure},
		{ID: HostBlockContext, Name: "block_context", MinArgs: 0, MaxArgs: 0, MaxArgBytes: 0, Class: ClassPure},
		{ID: HostChargeGas, Name: "charge_gas", MinArgs: 0, MaxArgs: 0, MaxArgBytes: 0, Class: ClassPure},
		{ID: HostReturn, Name: "return", MinArgs: 0, MaxArgs: 0, MaxArgBytes: 0, Class: ClassPure},
		{ID: HostScheduleSelf, Name: "schedule_self", MinArgs: 1, MaxArgs: 1, MaxArgBytes: MaxHostArgBytes, Class: ClassEffectful},
		{ID: HostHashSHA256, Name: "hash_sha256", MinArgs: 1, MaxArgs: 1, MaxArgBytes: MaxHostArgBytes, Class: ClassPure},
		{ID: HostHashBLAKE3, Name: "hash_blake3", MinArgs: 1, MaxArgs: 1, MaxArgBytes: MaxHostArgBytes, Class: ClassPure},
		{ID: HostVerifyEd25519, Name: "verify_ed25519", MinArgs: 3, MaxArgs: 3, MaxArgBytes: MaxHostArgBytes, Class: ClassPure},
		{ID: HostParseAetraAddress, Name: "parse_aetra_address", MinArgs: 1, MaxArgs: 1, MaxArgBytes: 128, Class: ClassPure},
		{ID: HostFormatAetraAddress, Name: "format_aetra_address", MinArgs: 1, MaxArgs: 1, MaxArgBytes: 32, Class: ClassPure},
		{ID: HostDeleteStorage, Name: "delete_storage", MinArgs: 1, MaxArgs: 1, MaxArgBytes: MaxKeySize, Class: ClassEffectful},
		{ID: HostEmitEvent, Name: "emit_event", MinArgs: 1, MaxArgs: 1, MaxArgBytes: MaxHostArgBytes, Class: ClassEffectful},
		{ID: HostSendInternal, Name: "send_internal_message", MinArgs: 1, MaxArgs: 1, MaxArgBytes: MaxHostArgBytes, Class: ClassEffectful},
		{ID: HostGetBlockHeight, Name: "get_block_height", MinArgs: 0, MaxArgs: 0, MaxArgBytes: 0, Class: ClassPure},
		{ID: HostGetChainID, Name: "get_chain_id", MinArgs: 0, MaxArgs: 0, MaxArgBytes: 0, Class: ClassPure},
		{ID: HostGetContractAddress, Name: "get_contract_address", MinArgs: 0, MaxArgs: 0, MaxArgBytes: 0, Class: ClassPure},
		{ID: HostGetCallerSource, Name: "get_caller_source", MinArgs: 0, MaxArgs: 0, MaxArgBytes: 0, Class: ClassPure},
		{ID: HostGetAttachedValue, Name: "get_attached_value", MinArgs: 0, MaxArgs: 0, MaxArgBytes: 0, Class: ClassPure},
		{ID: HostAbort, Name: "abort", MinArgs: 1, MaxArgs: 1, MaxArgBytes: 8, Class: ClassEffectful},
		{ID: HostWallClockTime, Name: "wall_clock_time", Forbidden: true, Class: ClassPure},
		{ID: HostRandomness, Name: "randomness", Forbidden: true, Class: ClassPure},
		{ID: HostFilesystem, Name: "filesystem", Forbidden: true, Class: ClassEffectful},
		{ID: HostNetwork, Name: "network", Forbidden: true, Class: ClassEffectful},
		{ID: HostFloatingPoint, Name: "floating_point", Forbidden: true, Class: ClassPure},
		{ID: HostGoroutine, Name: "goroutine", Forbidden: true, Class: ClassEffectful},
		{ID: HostProcessEnv, Name: "process_env", Forbidden: true, Class: ClassPure},
		{ID: HostNondeterministicMap, Name: "nondeterministic_map_iteration", Forbidden: true, Class: ClassEffectful},
	}
	out := make(map[HostFunction]HostFunctionSpec, len(specs))
	for _, spec := range specs {
		out[spec.ID] = spec
	}
	return out
}

func AllowedHostFunctions() []HostFunction {
	registry := HostFunctionRegistry()
	hosts := make([]HostFunction, 0, len(registry))
	for host, spec := range registry {
		if !spec.Forbidden {
			hosts = append(hosts, host)
		}
	}
	sort.Slice(hosts, func(i, j int) bool { return hosts[i] < hosts[j] })
	return hosts
}

func IsForbiddenHostFunction(host HostFunction) bool {
	spec, ok := HostFunctionRegistry()[host]
	return ok && spec.Forbidden
}

func ValidateHostImport(host HostFunction, caps CapabilityMask) error {
	spec, ok := HostFunctionRegistry()[host]
	if !ok {
		return fmt.Errorf("AVM host function %d is unknown", host)
	}
	if spec.Forbidden {
		return fmt.Errorf("AVM host function %q is forbidden", spec.Name)
	}

	switch host {
	case HostHashSHA256, HostHashBLAKE3, HostVerifyEd25519:
		if !caps.Crypto {
			return errors.New("missing crypto capability")
		}
	case HostGetBlockHeight, HostGetChainID, HostGetContractAddress, HostGetCallerSource, HostGetAttachedValue:
		if !caps.Chain {
			return errors.New("missing chain capability")
		}
	case HostSendInternal, HostEmitEvent:
		if !caps.Messaging {
			return errors.New("missing messaging capability")
		}
	case HostReadStorage, HostWriteStorage, HostDeleteStorage:
		if !caps.Storage {
			return errors.New("missing storage capability")
		}
	}
	return nil
}

func ValidateHostCall(host HostFunction, encodedArgs []byte) ([][]byte, error) {
	spec, ok := HostFunctionRegistry()[host]
	if !ok {
		return nil, fmt.Errorf("AVM host call %d is unknown", host)
	}
	if spec.Forbidden {
		return nil, fmt.Errorf("AVM host call %q is forbidden", spec.Name)
	}
	args, err := DecodeHostArgs(encodedArgs)
	if err != nil {
		return nil, err
	}
	if len(args) < int(spec.MinArgs) || len(args) > int(spec.MaxArgs) {
		return nil, fmt.Errorf("AVM host call %q expects %d-%d args", spec.Name, spec.MinArgs, spec.MaxArgs)
	}
	for _, arg := range args {
		if len(arg) > int(spec.MaxArgBytes) {
			return nil, fmt.Errorf("AVM host call %q arg must be <= %d bytes", spec.Name, spec.MaxArgBytes)
		}
	}
	if host == HostVerifyEd25519 {
		if len(args[0]) != ed25519.PublicKeySize {
			return nil, errors.New("AVM ed25519 public key must be 32 bytes")
		}
		if len(args[1]) != ed25519.SignatureSize {
			return nil, errors.New("AVM ed25519 signature must be 64 bytes")
		}
	}
	if host == HostAbort && len(args[0]) != 8 {
		return nil, errors.New("AVM abort exit code must be u64")
	}
	return args, nil
}

func EncodeHostArgs(args ...[]byte) ([]byte, error) {
	if len(args) > MaxHostArgs {
		return nil, fmt.Errorf("AVM host arg count must be <= %d", MaxHostArgs)
	}
	buf := bytes.NewBuffer(nil)
	writeU16(buf, uint16(len(args)))
	for _, arg := range args {
		if len(arg) > MaxHostArgBytes {
			return nil, fmt.Errorf("AVM host arg must be <= %d bytes", MaxHostArgBytes)
		}
		writeU32(buf, uint32(len(arg)))
		buf.Write(arg)
	}
	return buf.Bytes(), nil
}

func DecodeHostArgs(encoded []byte) ([][]byte, error) {
	if len(encoded) == 0 {
		return nil, nil
	}
	reader := bytes.NewReader(encoded)
	count, err := readU16(reader)
	if err != nil {
		return nil, err
	}
	if count > MaxHostArgs {
		return nil, fmt.Errorf("AVM host arg count must be <= %d", MaxHostArgs)
	}
	args := make([][]byte, count)
	for i := range args {
		length, err := readU32(reader)
		if err != nil {
			return nil, err
		}
		if length > MaxHostArgBytes {
			return nil, fmt.Errorf("AVM host arg must be <= %d bytes", MaxHostArgBytes)
		}
		args[i] = make([]byte, length)
		if length > 0 {
			if _, err := io.ReadFull(reader, args[i]); err != nil {
				return nil, err
			}
		}
	}
	if reader.Len() != 0 {
		return nil, errors.New("AVM host args have trailing data")
	}
	return args, nil
}

// RandomBeacon produces deterministic entropy for contract-level randomness.
// Uses only consensus inputs — NO process randomness, wall-clock, or external entropy.
//
//	random = SHA256(previous_state_root || block_entropy || message_hash || domain)
//
// The result is deterministic per block: all validators produce identical values.
func RandomBeacon(previousStateRoot, blockEntropy, messageHash, domain []byte) []byte {
	h := sha256.New()
	h.Write(previousStateRoot)
	h.Write(blockEntropy)
	h.Write(messageHash)
	h.Write(domain)
	return h.Sum(nil)
}

// ExecutionIsolationBoundary defines the isolation guarantees for ExecutionFrame.
//
// Invariants:
//   - Each ExecutionFrame is isolated: no shared mutable memory between frames
//   - All dependencies passed via Chunks (immutable, content-addressed)
//   - No global VM state mutation allowed
//   - Host functions execute inside sandboxed ExecutionFrame
//   - All inputs must come from Chunk or BlockContext
//
// This ensures:
//   - Deterministic execution across all validators
//   - No side-channel communication between contracts
//   - Replay-safe execution model
type ExecutionIsolationBoundary struct {
	FrameID		uint64
	ParentFrameID	uint64
	AllowedReads	[]string	// Chunk hashes readable by this frame
	AllowedWrites	[]string	// Chunk hashes writable by this frame
}

// ValidateIsolation checks that a host call respects isolation boundaries.
func ValidateIsolation(boundary *ExecutionIsolationBoundary, host HostFunction, args [][]byte) error {

	if IsPureHostFunction(host) {
		return nil
	}

	switch host {
	case HostReadStorage:
		if len(args) < 1 {
			return errors.New("read_storage requires key argument")
		}

	case HostWriteStorage:
		if len(args) < 2 {
			return errors.New("write_storage requires key and value arguments")
		}

	case HostDeleteStorage:
		if len(args) < 1 {
			return errors.New("delete_storage requires key argument")
		}

	}

	return nil
}

// IsPureHostFunction returns true if the host function has no side effects.
func IsPureHostFunction(host HostFunction) bool {
	spec, ok := HostFunctionRegistry()[host]
	if !ok {
		return false
	}
	return spec.Class == ClassPure
}

// HostErrorBoundary defines how host function failures are handled.
//
// Invariants:
//   - Host function failure MUST produce VM exit code (mapped)
//   - Failure MUST NOT corrupt state
//   - Failure MUST NOT partially apply effects
//   - All EFFECTFUL operations are staged, not applied immediately
//
// Error mapping:
//   - Invalid arguments → ExitCodeValidationFailed (1)
//   - Forbidden host call → ExitCodeForbiddenHostCall (7)
//   - Capability violation → ExitCodeUnauthorized (2)
//   - Resource exhaustion → ExitCodeOutOfGas (8)
type HostErrorBoundary struct {
	LastHostCall	HostFunction
	LastError	error
	MappedExitCode	uint32
}

// MapHostErrorToExitCode maps a host function error to a VM exit code.
func MapHostErrorToExitCode(err error) uint32 {
	if err == nil {
		return 0
	}

	errStr := err.Error()

	if bytes.Contains([]byte(errStr), []byte("forbidden")) {
		return 7
	}
	if bytes.Contains([]byte(errStr), []byte("capability")) {
		return 2
	}
	if bytes.Contains([]byte(errStr), []byte("gas")) {
		return 8
	}

	return 1
}
