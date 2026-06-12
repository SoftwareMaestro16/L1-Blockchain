package types

import (
	"bytes"
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/x/aetravm/avm"
)

const (
	DefaultDeployGas		uint64	= 500_000
	DefaultExecuteGas		uint64	= 250_000
	DefaultQueryGas			uint64	= 50_000
	DefaultStorageWriteGas		uint64	= 1_000
	DefaultMessageForwardingGas	uint64	= 10_000

	DefaultMaxStateSizeBytes	uint64	= 256 * 1024
	DefaultMaxStorageKeyBytes	uint32	= 128
	DefaultMaxStorageValueBytes	uint64	= 8 * 1024
	DefaultMaxEmittedMessages	uint32	= 16
	DefaultMaxContractMessages	uint32	= 1_000
	DefaultContractGovernanceSeed	byte	= 1
)

func DefaultContractZonePolicy() ContractZonePolicy {
	runtime := DefaultRuntimePolicy()
	return ContractZonePolicy{
		Runtime:		runtime,
		GovernanceAuthority:	filledAddress(DefaultContractGovernanceSeed),
		UploadMode:		UploadModeGovernanceOnly,
		InstantiateMode:	InstantiateModeCodeOwnerOnly,
		MigrationsEnabled:	true,
		GasModel: GasModel{
			DeployGas:		DefaultDeployGas,
			ExecuteGas:		DefaultExecuteGas,
			QueryGas:		DefaultQueryGas,
			StorageWriteGas:	DefaultStorageWriteGas,
			MessageForwardingGas:	DefaultMessageForwardingGas,
		},
		Limits: ContractLimits{
			MaxCodeSizeBytes:	uint64(runtime.AVMParams.MaxCodeBytes),
			MaxStateSizeBytes:	DefaultMaxStateSizeBytes,
			MaxStorageKeyBytes:	DefaultMaxStorageKeyBytes,
			MaxStorageValueBytes:	DefaultMaxStorageValueBytes,
			MaxQueryResponseBytes:	runtime.CosmWasmPolicy.MaxQueryResponseBytes,
			MaxQueryDepth:		runtime.CosmWasmPolicy.MaxQueryDepth,
			MaxEmittedMessages:	DefaultMaxEmittedMessages,
			MaxMessagesPerBlock:	DefaultMaxContractMessages,
		},
	}
}

func (p ContractZonePolicy) Validate() error {
	if err := ValidateRuntimePolicy(p.Runtime); err != nil {
		return err
	}
	if err := validateContractAddress("contract governance authority", p.GovernanceAuthority); err != nil {
		return err
	}
	if p.UploadMode != UploadModeGovernanceOnly && p.UploadMode != UploadModeAllowlistTestnet {
		return fmt.Errorf("invalid contract upload mode %q", p.UploadMode)
	}
	if p.UploadMode == UploadModeAllowlistTestnet && !p.TestnetAllowlist {
		return errors.New("contract upload allowlist is testnet-only")
	}
	if p.UploadMode == UploadModeAllowlistTestnet && len(p.UploadAllowlist) == 0 {
		return errors.New("contract upload allowlist must not be empty")
	}
	for _, actor := range p.UploadAllowlist {
		if err := validateContractAddress("contract upload allowlist actor", actor); err != nil {
			return err
		}
	}
	if p.InstantiateMode != InstantiateModeCodeOwnerOnly && p.InstantiateMode != InstantiateModeEverybody {
		return fmt.Errorf("invalid contract instantiate mode %q", p.InstantiateMode)
	}
	if err := p.GasModel.Validate(); err != nil {
		return err
	}
	if err := p.Limits.Validate(); err != nil {
		return err
	}
	if p.HostPolicy.LocalTimeEnabled || p.HostPolicy.ExternalAPIsEnabled || p.HostPolicy.CrossContractStateMutation {
		return errors.New("contract host policy enables nondeterministic or unsafe host capability")
	}
	if p.HostPolicy.ConsensusRandomnessEnabled {
		return errors.New("contract consensus randomness is not enabled in readiness spec")
	}
	return nil
}

func (g GasModel) Validate() error {
	if g.DeployGas == 0 || g.ExecuteGas == 0 || g.QueryGas == 0 || g.StorageWriteGas == 0 || g.MessageForwardingGas == 0 {
		return errors.New("contract gas model values must be positive")
	}
	return nil
}

func (l ContractLimits) Validate() error {
	if l.MaxCodeSizeBytes == 0 {
		return errors.New("contract max code size must be positive")
	}
	if l.MaxStateSizeBytes == 0 {
		return errors.New("contract max state size must be positive")
	}
	if l.MaxStorageKeyBytes == 0 || l.MaxStorageKeyBytes > avm.MaxKeySize {
		return fmt.Errorf("contract max storage key bytes must be in 1..%d", avm.MaxKeySize)
	}
	if l.MaxStorageValueBytes == 0 {
		return errors.New("contract max storage value bytes must be positive")
	}
	if l.MaxQueryResponseBytes == 0 {
		return errors.New("contract max query response bytes must be positive")
	}
	if l.MaxQueryDepth == 0 {
		return errors.New("contract max query depth must be positive")
	}
	if l.MaxEmittedMessages == 0 {
		return errors.New("contract max emitted messages must be positive")
	}
	if l.MaxMessagesPerBlock == 0 {
		return errors.New("contract max messages per block must be positive")
	}
	return nil
}

func CanUploadContract(actor sdk.AccAddress, policy ContractZonePolicy) error {
	if err := policy.Validate(); err != nil {
		return err
	}
	if err := validateContractAddress("contract upload actor", actor); err != nil {
		return err
	}
	if bytes.Equal(actor, policy.GovernanceAuthority) {
		return nil
	}
	if policy.UploadMode != UploadModeAllowlistTestnet {
		return errors.New("contract upload requires governance authority")
	}
	for _, allowed := range policy.UploadAllowlist {
		if bytes.Equal(actor, allowed) {
			return nil
		}
	}
	return errors.New("contract upload actor is not allowlisted")
}

func CanInstantiateContract(actor sdk.AccAddress, code ContractCode, policy ContractZonePolicy) error {
	if err := policy.Validate(); err != nil {
		return err
	}
	if err := validateContractAddress("contract instantiate actor", actor); err != nil {
		return err
	}
	if policy.InstantiateMode == InstantiateModeEverybody {
		return nil
	}
	if !bytes.Equal(actor, code.Owner) {
		return errors.New("contract instantiate requires code owner")
	}
	return nil
}

func CanMigrateContract(actor sdk.AccAddress, contract ContractInstance, policy ContractZonePolicy) error {
	if err := policy.Validate(); err != nil {
		return err
	}
	if !policy.MigrationsEnabled {
		return errors.New("contract migrations are disabled by governance")
	}
	if err := validateContractAddress("contract migrate actor", actor); err != nil {
		return err
	}
	if err := validateContractAddress("contract admin", contract.Admin); err != nil {
		return err
	}
	if !bytes.Equal(actor, contract.Admin) {
		return errors.New("contract migrate requires contract admin")
	}
	return nil
}

func validateContractAddress(field string, addr sdk.AccAddress) error {
	if len(addr) == 0 {
		return fmt.Errorf("%s is required", field)
	}
	return addressing.RejectZeroAddress(field, addr)
}

func filledAddress(fill byte) sdk.AccAddress {
	out := make([]byte, 20)
	for i := range out {
		out[i] = fill
	}
	return sdk.AccAddress(out)
}
