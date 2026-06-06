package types

import (
	"errors"
	"fmt"
	"strings"

	coretypes "github.com/sovereign-l1/l1/x/aethercore/types"
)

const (
	EventTypeCodeStored           = "contracts.code_stored"
	EventTypeContractInstantiated = "contracts.instantiated"
	EventTypeContractExecuted     = "contracts.executed"

	ErrInvalidParams    = "contracts_invalid_params"
	ErrInvalidGenesis   = "contracts_invalid_genesis"
	ErrContractNotFound = "contracts_not_found"
	ErrInvalidBytecode  = "contracts_invalid_bytecode"
	ErrExecutionFailed  = "contracts_execution_failed"
)

type Params struct {
	Enabled                 bool
	MaxCodeBytes            uint64
	MaxContractStorageBytes uint64
	MaxGasPerExecution      uint64
}

type GenesisState struct {
	Params    Params
	StateRoot string
}

type MsgStoreCode struct {
	Authority string
	CodeHash  string
	CodeBytes uint64
}

type StoreCodeResponse struct {
	CodeID    string
	StateRoot string
}

type QueryContractRequest struct {
	ContractAddress string
}

type QueryContractResponse struct {
	ContractAddress string
	StateRoot       string
	Found           bool
}

type MsgServer interface {
	StoreCode(MsgStoreCode) (StoreCodeResponse, error)
}

type QueryServer interface {
	Contract(QueryContractRequest) (QueryContractResponse, error)
	RootContribution() (coretypes.RootContribution, error)
}

func DefaultParams() Params {
	return Params{
		Enabled:                 true,
		MaxCodeBytes:            4 * 1024 * 1024,
		MaxContractStorageBytes: 64 * 1024 * 1024,
		MaxGasPerExecution:      100_000_000,
	}
}

func DefaultGenesis() GenesisState {
	gs := GenesisState{Params: DefaultParams()}
	gs.StateRoot = ComputeContractsStateRoot(gs)
	return gs
}

func (p Params) Validate() error {
	if p.MaxCodeBytes == 0 {
		return errors.New(ErrInvalidParams + ": max code bytes must be positive")
	}
	if p.MaxContractStorageBytes == 0 {
		return errors.New(ErrInvalidParams + ": max contract storage bytes must be positive")
	}
	if p.MaxGasPerExecution == 0 {
		return errors.New(ErrInvalidParams + ": max gas per execution must be positive")
	}
	return nil
}

func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	if err := coretypes.ValidateHash("contracts genesis state root", gs.StateRoot); err != nil {
		return err
	}
	if gs.StateRoot != ComputeContractsStateRoot(gs) {
		return errors.New(ErrInvalidGenesis + ": state root mismatch")
	}
	return nil
}

func RootContribution(gs GenesisState) (coretypes.RootContribution, error) {
	if err := gs.Validate(); err != nil {
		return coretypes.RootContribution{}, err
	}
	return coretypes.NewRootContribution(coretypes.RootType(ModuleName), ModuleName, gs.StateRoot)
}

func ComputeContractsStateRoot(gs GenesisState) string {
	return coretypes.DeterministicEmptyRootCommitment(coretypes.RootType(ModuleName), fmt.Sprintf(
		"enabled=%t/code=%020d/storage=%020d/gas=%020d",
		gs.Params.Enabled,
		gs.Params.MaxCodeBytes,
		gs.Params.MaxContractStorageBytes,
		gs.Params.MaxGasPerExecution,
	))
}

func ValidateContractAddress(address string) error {
	address = strings.TrimSpace(address)
	if address == "" {
		return errors.New(ErrContractNotFound + ": contract address is required")
	}
	if len(address) > 128 {
		return errors.New(ErrContractNotFound + ": contract address is too long")
	}
	return nil
}
