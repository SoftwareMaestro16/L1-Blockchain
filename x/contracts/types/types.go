package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
	"github.com/sovereign-l1/l1/x/internal/prototype"
)

const (
	EventTypeCodeStored		= "contracts.code_stored"
	EventTypeContractInstantiated	= "contracts.instantiated"
	EventTypeContractExecuted	= "contracts.executed"

	ErrInvalidParams	= "contracts_invalid_params"
	ErrInvalidGenesis	= "contracts_invalid_genesis"
	ErrContractNotFound	= "contracts_not_found"
	ErrInvalidBytecode	= "contracts_invalid_bytecode"
	ErrExecutionFailed	= "contracts_execution_failed"
)

type Params struct {
	Authority			string
	Enabled				bool
	MaxCodeBytes			uint64
	MaxContractStorageBytes		uint64
	MaxGasPerExecution		uint64
	StorageRentPerByteBlock		uint64
	MaxInitDataBytes		uint64
	MaxStateInitSaltBytes		uint64
	MaxStateInitDependencies	uint32
}

type GenesisState struct {
	Params		Params
	State		State
	StateRoot	string
}

type MsgStoreCode struct {
	Authority	string
	CodeHash	string
	CodeBytes	uint64
	Bytecode	[]byte
}

type StoreCodeResponse struct {
	CodeID		string
	StateRoot	string
}

type QueryContractRequest struct {
	ContractAddress	string
	ChainID		string
	Namespace	string
	Deployer	string
	StateInit	*StateInit
}

type QueryContractResponse struct {
	ContractAddress	string
	StateRoot	string
	Found		bool
	Virtual		bool
	Contract	Contract
}

type MsgServer interface {
	StoreCode(MsgStoreCode) (StoreCodeResponse, error)
	DeployContract(MsgDeployContract) (InstantiateContractResponse, error)
	ExecuteExternal(MsgExecuteExternal) (ExecuteContractResponse, error)
	ExecuteInternal(MsgExecuteInternal) (InternalMessage, error)
	SendInternalMessage(MsgSendInternalMessage) (InternalMessage, error)
	UpgradeContractCode(MsgUpgradeContractCode) (ContractReceipt, error)
	MigrateContractState(MsgMigrateContractState) (ContractReceipt, error)
	SetContractAdmin(MsgSetContractAdmin) (ContractReceipt, error)
	DisableContractUpgrades(MsgDisableContractUpgrades) (ContractReceipt, error)
	UpdateContractParams(MsgUpdateContractParams) error
}

type QueryServer interface {
	Params() Params
	Code(QueryCodeRequest) (CodeRecord, bool, error)
	Codes(QueryCodesRequest) ([]CodeRecord, error)
	Contract(QueryContractRequest) (QueryContractResponse, error)
	Contracts(QueryContractsRequest) ([]Contract, error)
	ContractStorage(QueryContractStorageRequest) ([]ContractStorageEntry, error)
	ContractReceipts(QueryContractReceiptsRequest) ([]ContractReceipt, error)
	ContractQueue(QueryContractQueueRequest) ([]InternalMessage, error)
	ContractEvents(QueryContractEventsRequest) error
	ContractStateRoot(QueryContractStateRootRequest) (string, error)
	RootContribution() (coretypes.RootContribution, error)
}

func DefaultParams() Params {
	return Params{
		Authority:			prototype.DefaultAuthority,
		Enabled:			true,
		MaxCodeBytes:			4 * 1024 * 1024,
		MaxContractStorageBytes:	64 * 1024 * 1024,
		MaxGasPerExecution:		100_000_000,
		StorageRentPerByteBlock:	1,
		MaxInitDataBytes:		MaxContractPayloadBytes,
		MaxStateInitSaltBytes:		MaxContractSaltBytes,
		MaxStateInitDependencies:	MaxContractDependencies,
	}
}

func DefaultGenesis() GenesisState {
	gs := GenesisState{Params: DefaultParams()}
	gs.State = gs.State.Normalize()
	gs.StateRoot = ComputeContractsStateRoot(gs)
	return gs
}

func (p Params) Validate() error {
	if strings.TrimSpace(p.Authority) == "" {
		return errors.New(ErrInvalidParams + ": authority is required")
	}
	if p.MaxCodeBytes == 0 {
		return errors.New(ErrInvalidParams + ": max code bytes must be positive")
	}
	if p.MaxContractStorageBytes == 0 {
		return errors.New(ErrInvalidParams + ": max contract storage bytes must be positive")
	}
	if p.MaxGasPerExecution == 0 {
		return errors.New(ErrInvalidParams + ": max gas per execution must be positive")
	}
	if p.MaxInitDataBytes == 0 {
		return errors.New(ErrInvalidParams + ": max init data bytes must be positive")
	}
	if p.MaxStateInitSaltBytes == 0 {
		return errors.New(ErrInvalidParams + ": max state init salt bytes must be positive")
	}
	if p.MaxStateInitDependencies == 0 {
		return errors.New(ErrInvalidParams + ": max state init dependencies must be positive")
	}
	return nil
}

func (p Params) Authorize(authority string) error {
	if strings.TrimSpace(authority) != p.Authority {
		return errors.New(ErrUnauthorized + ": authority mismatch")
	}
	return nil
}

func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	if err := gs.State.Validate(gs.Params); err != nil {
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
	stateJSON, err := json.Marshal(gs.State.Normalize())
	if err != nil {
		panic(err)
	}
	return coretypes.DeterministicEmptyRootCommitment(coretypes.RootType(ModuleName), fmt.Sprintf(
		"authority=%s/enabled=%t/code=%020d/storage=%020d/gas=%020d/rent=%020d/init=%020d/salt=%020d/deps=%010d/state=%s",
		gs.Params.Authority,
		gs.Params.Enabled,
		gs.Params.MaxCodeBytes,
		gs.Params.MaxContractStorageBytes,
		gs.Params.MaxGasPerExecution,
		gs.Params.StorageRentPerByteBlock,
		gs.Params.MaxInitDataBytes,
		gs.Params.MaxStateInitSaltBytes,
		gs.Params.MaxStateInitDependencies,
		string(stateJSON),
	))
}

func ValidateContractAddress(address string) error {
	address = strings.TrimSpace(address)
	if address == "" {
		return errors.New(ErrContractNotFound + ": contract address is required")
	}
	return ValidateUserFacingAEAddress("contract address", address)
}
