package keeper

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"

	corestore "cosmossdk.io/core/store"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
	"github.com/sovereign-l1/l1/x/contracts/types"
)

const storageRentReserveModule = "feecollector_storage_rent_reserve"

var storageRentBaseDenom = "naet"

// BankKeeper defines the subset of bank functionality needed by the contracts keeper.
type BankKeeper interface {
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}

type Keeper struct {
	genesis			types.GenesisState
	storeService		corestore.KVStoreService
	accountStatusReader	AccountStatusReader
	bankKeeper		BankKeeper
	runtimeCtx		context.Context
	storageRentRateProvider	StorageRentRateProvider
}

const (
	accountStatusActive	= "active"
	accountStatusInactive	= "inactive"
	accountStatusFrozen	= "frozen"
)

var genesisKey = []byte{0x01}

type AccountStatusReader interface {
	AccountStatus(context.Context, string) (string, bool, error)
}

// StorageRentRateProvider queries the active storage rent rate from the storage-rent module.
type StorageRentRateProvider interface {
	StorageRentRatePerByteBlock() uint64
}

func NewKeeper() Keeper {
	return Keeper{genesis: types.DefaultGenesis()}
}

func NewPersistentKeeper(storeService corestore.KVStoreService) Keeper {
	return Keeper{genesis: types.DefaultGenesis(), storeService: storeService}
}

func NewKeeperWithAccountStatus(reader AccountStatusReader) Keeper {
	k := NewKeeper()
	k.accountStatusReader = reader
	return k
}

func (k Keeper) WithAccountStatusReader(reader AccountStatusReader) Keeper {
	k.accountStatusReader = reader
	return k
}

func (k Keeper) WithBankKeeper(bk BankKeeper) Keeper {
	k.bankKeeper = bk
	return k
}

func (k Keeper) WithStorageRentRateProvider(provider StorageRentRateProvider) Keeper {
	k.storageRentRateProvider = provider
	return k
}

func DefaultGenesis() types.GenesisState {
	return types.DefaultGenesis()
}

func (k *Keeper) InitGenesis(gs types.GenesisState) error {
	gs = types.RefreshStateRoot(gs)
	if err := gs.Validate(); err != nil {
		return err
	}
	k.genesis = gs
	return nil
}

func (k *Keeper) InitGenesisState(ctx context.Context, gs types.GenesisState) error {
	if err := k.InitGenesis(gs); err != nil {
		return err
	}
	k.runtimeCtx = ctx
	return k.writeGenesis(ctx)
}

func (k Keeper) ExportGenesis() types.GenesisState {
	return types.RefreshStateRoot(k.genesis)
}

func (k Keeper) ExportGenesisState(ctx context.Context) (types.GenesisState, error) {
	if k.storeService == nil {
		return k.ExportGenesis(), nil
	}
	if !reflect.DeepEqual(k.genesis, types.DefaultGenesis()) {
		return k.ExportGenesis(), nil
	}
	bz, err := k.storeService.OpenKVStore(ctx).Get(genesisKey)
	if err != nil {
		return types.GenesisState{}, err
	}
	if len(bz) == 0 {
		return types.DefaultGenesis(), nil
	}
	var gs types.GenesisState
	if err := json.Unmarshal(bz, &gs); err != nil {
		return types.GenesisState{}, err
	}
	gs = types.RefreshStateRoot(gs)
	if err := gs.Validate(); err != nil {
		return types.GenesisState{}, err
	}
	return gs, nil
}

func (k Keeper) Params() types.Params {
	return k.genesis.Params
}

func (k Keeper) Code(req types.QueryCodeRequest) (types.CodeRecord, bool, error) {
	if req.CodeID == "" {
		return types.CodeRecord{}, false, errors.New("contract code id is required")
	}
	code, found := findCode(k.genesis.State.Codes, req.CodeID)
	return code, found, nil
}

func (k Keeper) Codes(req types.QueryCodesRequest) ([]types.CodeRecord, error) {
	if err := types.ValidateQueryPagination(req.Pagination); err != nil {
		return nil, err
	}
	codes := k.genesis.State.Normalize().Codes
	if uint32(len(codes)) > req.Pagination.Limit {
		codes = codes[:req.Pagination.Limit]
	}
	return append([]types.CodeRecord(nil), codes...), nil
}

func (k Keeper) ValidateInvariants() error {
	return k.genesis.Validate()
}

func (k Keeper) RootContribution() (coretypes.RootContribution, error) {
	return types.RootContribution(k.genesis)
}

func (k Keeper) Migrate1to2State(ctx context.Context) error {
	_, err := k.ExportGenesisState(ctx)
	return err
}

func (k *Keeper) StoreCode(msg types.MsgStoreCode) (types.StoreCodeResponse, error) {
	if !k.genesis.Params.Enabled {
		return types.StoreCodeResponse{}, errors.New(types.ErrExecutionFailed + ": module disabled")
	}
	if err := types.ValidateUserFacingAEAddress("contract code authority", msg.Authority); err != nil {
		return types.StoreCodeResponse{}, err
	}
	if err := k.ensureActiveWallet(k.runtimeCtx, msg.Authority, "contract code store"); err != nil {
		return types.StoreCodeResponse{}, err
	}
	return k.storeCodeUnchecked(msg)
}

func (k *Keeper) storeCodeUnchecked(msg types.MsgStoreCode) (types.StoreCodeResponse, error) {
	if len(msg.Bytecode) > 0 {
		if err := types.ValidateAVMBytecode(k.genesis.Params, msg.Bytecode); err != nil {
			return types.StoreCodeResponse{}, err
		}
		codeHash := types.CanonicalCodeHash(msg.Bytecode)
		if msg.CodeHash != "" && msg.CodeHash != codeHash {
			return types.StoreCodeResponse{}, errors.New(types.ErrInvalidBytecode + ": code hash must match canonical bytecode hash")
		}
		msg.CodeHash = codeHash
		msg.CodeBytes = uint64(len(msg.Bytecode))
	}
	if msg.CodeBytes == 0 || msg.CodeBytes > k.genesis.Params.MaxCodeBytes {
		return types.StoreCodeResponse{}, errors.New(types.ErrInvalidBytecode + ": code size out of bounds")
	}
	if err := coretypes.ValidateHash("contracts code hash", msg.CodeHash); err != nil {
		return types.StoreCodeResponse{}, err
	}
	next := k.genesis
	next.State.Codes = upsertCode(next.State.Codes, types.CodeRecord{
		CodeID:		msg.CodeHash,
		CodeHash:	msg.CodeHash,
		CodeBytes:	msg.CodeBytes,
		Bytecode:	append([]byte(nil), msg.Bytecode...),
		Owner:		msg.Authority,
	})
	next = types.RefreshStateRoot(next)
	if err := next.Validate(); err != nil {
		return types.StoreCodeResponse{}, err
	}
	k.genesis = next
	return types.StoreCodeResponse{CodeID: msg.CodeHash, StateRoot: k.genesis.StateRoot}, nil
}

func (k *Keeper) StoreCodeState(ctx context.Context, msg types.MsgStoreCode) (types.StoreCodeResponse, error) {
	if !k.genesis.Params.Enabled {
		return types.StoreCodeResponse{}, errors.New(types.ErrExecutionFailed + ": module disabled")
	}
	if err := types.ValidateUserFacingAEAddress("contract code authority", msg.Authority); err != nil {
		return types.StoreCodeResponse{}, err
	}
	if err := k.ensureActiveWallet(ctx, msg.Authority, "contract code store"); err != nil {
		return types.StoreCodeResponse{}, err
	}
	res, err := k.storeCodeUnchecked(msg)
	if err != nil {
		return types.StoreCodeResponse{}, err
	}
	return res, k.writeGenesis(ctx)
}

func (k *Keeper) DeployContract(msg types.MsgDeployContract) (types.InstantiateContractResponse, error) {
	return k.deployContract(k.runtimeCtx, msg)
}

func (k *Keeper) DeployContractState(ctx context.Context, msg types.MsgDeployContract) (types.InstantiateContractResponse, error) {
	res, err := k.deployContract(ctx, msg)
	if err != nil {
		return types.InstantiateContractResponse{}, err
	}
	return res, k.writeGenesis(ctx)
}

func (k *Keeper) deployContract(ctx context.Context, msg types.MsgDeployContract) (types.InstantiateContractResponse, error) {
	if err := msg.ValidateBasic(k.genesis.Params); err != nil {
		return types.InstantiateContractResponse{}, err
	}
	return k.instantiateContract(ctx, types.MsgInstantiateContract{
		Creator:	msg.Creator,
		CodeID:		msg.CodeID,
		ChainID:	msg.ChainID,
		Namespace:	msg.Namespace,
		StateInit:	msg.StateInit,
		InitMsg:	append([]byte(nil), msg.InitPayload...),
		Funds:		msg.InitialBalance,
		Admin:		msg.Admin,
		Salt:		msg.Salt,
		Upgradeable:	msg.Upgradeable,
		SystemOwned:	msg.SystemOwned,
		SchemaVersion:	msg.SchemaVersion,
		Height:		msg.Height,
	})
}

func (k *Keeper) ExecuteExternal(msg types.MsgExecuteExternal) (types.ExecuteContractResponse, error) {
	return k.executeExternal(k.runtimeCtx, msg)
}

func (k *Keeper) ExecuteExternalState(ctx context.Context, msg types.MsgExecuteExternal) (types.ExecuteContractResponse, error) {
	res, err := k.executeExternal(ctx, msg)
	if err != nil {
		return types.ExecuteContractResponse{}, err
	}
	return res, k.writeGenesis(ctx)
}

func (k *Keeper) executeExternal(ctx context.Context, msg types.MsgExecuteExternal) (types.ExecuteContractResponse, error) {
	if err := msg.ValidateBasic(k.genesis.Params); err != nil {
		return types.ExecuteContractResponse{}, err
	}
	if _, found := findContract(k.genesis.State.Contracts, msg.ContractAddress); !found && msg.StateInit != nil {
		user, _, err := types.DeriveContractAddressFromStateInit(msg.ChainID, msg.Namespace, msg.Sender, *msg.StateInit, k.genesis.Params)
		if err != nil {
			return types.ExecuteContractResponse{}, err
		}
		if user != msg.ContractAddress {
			return types.ExecuteContractResponse{}, errors.New(types.ErrContractNotFound + ": state init address does not match external execute target")
		}
		_, err = k.instantiateContract(ctx, types.MsgInstantiateContract{
			Creator:	msg.Sender,
			CodeID:		msg.StateInit.Normalize().CodeID,
			ChainID:	msg.ChainID,
			Namespace:	msg.Namespace,
			StateInit:	msg.StateInit,
			Height:		msg.Height,
		})
		if err != nil {
			return types.ExecuteContractResponse{}, err
		}
	}
	return k.executeContract(ctx, types.MsgExecuteContract{
		Sender:			msg.Sender,
		ContractAddress:	msg.ContractAddress,
		Msg:			append([]byte(nil), msg.Payload...),
		Funds:			msg.Funds,
		Height:			msg.Height,
	})
}

func (k *Keeper) ExecuteInternal(msg types.MsgExecuteInternal) (types.InternalMessage, error) {
	if err := msg.ValidateBasic(k.genesis.Params); err != nil {
		return types.InternalMessage{}, err
	}
	return k.ReceiveInternalMessage(types.MsgReceiveInternalMessage{
		SourceContractUser:	msg.Message.SourceContractUser,
		DestinationAccount:	msg.Message.DestinationAccount,
		Funds:			msg.Message.Funds,
		Opcode:			msg.Message.Opcode,
		QueryID:		msg.Message.QueryID,
		Body:			append([]byte(nil), msg.Message.Body...),
		Bounce:			msg.Message.Bounce,
		Deadline:		msg.Message.Deadline,
		GasLimit:		msg.Message.GasLimit,
		LogicalTime:		msg.Message.LogicalTime,
		MessageID:		msg.Message.MessageID,
		Height:			msg.Height,
	})
}

func (k *Keeper) SendInternalMessage(msg types.MsgSendInternalMessage) (types.InternalMessage, error) {
	if err := msg.ValidateBasic(k.genesis.Params); err != nil {
		return types.InternalMessage{}, err
	}
	return k.ReceiveInternalMessage(types.MsgReceiveInternalMessage{
		SourceContractUser:	msg.Message.SourceContractUser,
		DestinationAccount:	msg.Message.DestinationAccount,
		Funds:			msg.Message.Funds,
		Opcode:			msg.Message.Opcode,
		QueryID:		msg.Message.QueryID,
		Body:			append([]byte(nil), msg.Message.Body...),
		Bounce:			msg.Message.Bounce,
		Deadline:		msg.Message.Deadline,
		GasLimit:		msg.Message.GasLimit,
		LogicalTime:		msg.Message.LogicalTime,
		MessageID:		msg.Message.MessageID,
		Height:			msg.Height,
	})
}

func (k *Keeper) UpdateContractParams(msg types.MsgUpdateContractParams) error {
	if err := msg.ValidateBasic(); err != nil {
		return err
	}
	next := k.genesis
	next.Params = msg.Params
	next = types.RefreshStateRoot(next)
	if err := next.Validate(); err != nil {
		return err
	}
	k.genesis = next
	return nil
}

func (k Keeper) Contract(req types.QueryContractRequest) (types.QueryContractResponse, error) {
	if strings.TrimSpace(req.ContractAddress) == "" && req.StateInit != nil {
		user, _, err := types.DeriveContractAddressFromStateInit(req.ChainID, req.Namespace, req.Deployer, *req.StateInit, k.genesis.Params)
		if err != nil {
			return types.QueryContractResponse{}, err
		}
		req.ContractAddress = user
	}
	if err := types.ValidateContractAddress(req.ContractAddress); err != nil {
		return types.QueryContractResponse{}, err
	}
	contract, found := findContract(k.genesis.State.Contracts, req.ContractAddress)
	if !found && req.StateInit != nil {
		user, _, err := types.DeriveContractAddressFromStateInit(req.ChainID, req.Namespace, req.Deployer, *req.StateInit, k.genesis.Params)
		if err != nil {
			return types.QueryContractResponse{}, err
		}
		if user != req.ContractAddress {
			return types.QueryContractResponse{}, errors.New(types.ErrContractNotFound + ": state init address does not match query address")
		}
		return types.QueryContractResponse{ContractAddress: req.ContractAddress, StateRoot: k.genesis.StateRoot, Found: false, Virtual: true}, nil
	}
	if found {
		if err := types.EnsureContractLifecycleAction(contract, types.ContractLifecycleActionQuery); err != nil {
			return types.QueryContractResponse{}, err
		}
	}
	return types.QueryContractResponse{ContractAddress: req.ContractAddress, StateRoot: k.genesis.StateRoot, Found: found, Contract: contract}, nil
}

func (k Keeper) Contracts(req types.QueryContractsRequest) ([]types.Contract, error) {
	if err := types.ValidateQueryPagination(req.Pagination); err != nil {
		return nil, err
	}
	contracts := k.genesis.State.Normalize().Contracts
	if uint32(len(contracts)) > req.Pagination.Limit {
		contracts = contracts[:req.Pagination.Limit]
	}
	return append([]types.Contract(nil), contracts...), nil
}

func (k Keeper) ContractStorage(req types.QueryContractStorageRequest) ([]types.ContractStorageEntry, error) {
	if err := types.ValidateContractAddress(req.ContractAddress); err != nil {
		return nil, err
	}
	if err := types.ValidateQueryPagination(req.Pagination); err != nil {
		return nil, err
	}
	contract, found := findContract(k.genesis.State.Contracts, req.ContractAddress)
	if !found {
		return nil, errors.New(types.ErrContractNotFound + ": contract not found")
	}
	if err := types.EnsureContractLifecycleAction(contract, types.ContractLifecycleActionProofQuery); err != nil {
		return nil, err
	}
	entries := []types.ContractStorageEntry{{
		ContractAddress:	contract.AddressUser,
		Key:			[]byte("data"),
		Value:			append([]byte(nil), contract.Data...),
	}}
	out := make([]types.ContractStorageEntry, 0, len(entries))
	for _, entry := range entries {
		if len(req.KeyPrefix) != 0 && !bytes.HasPrefix(entry.Key, req.KeyPrefix) {
			continue
		}
		out = append(out, entry)
		if uint32(len(out)) == req.Pagination.Limit {
			break
		}
	}
	return out, nil
}

func (k Keeper) ContractReceipts(req types.QueryContractReceiptsRequest) ([]types.ContractReceipt, error) {
	if err := types.ValidateContractAddress(req.ContractAddress); err != nil {
		return nil, err
	}
	if err := types.ValidateQueryPagination(req.Pagination); err != nil {
		return nil, err
	}
	receipts := k.genesis.State.Normalize().Receipts
	out := make([]types.ContractReceipt, 0)
	contract, found := findContract(k.genesis.State.Contracts, req.ContractAddress)
	if !found {
		return nil, errors.New(types.ErrContractNotFound + ": contract not found")
	}
	if err := types.EnsureContractLifecycleAction(contract, types.ContractLifecycleActionQuery); err != nil {
		return nil, err
	}
	for _, receipt := range receipts {
		if receipt.ContractAddress != req.ContractAddress {
			continue
		}
		out = append(out, receipt)
		if uint32(len(out)) == req.Pagination.Limit {
			break
		}
	}
	return out, nil
}

func (k Keeper) ContractQueue(req types.QueryContractQueueRequest) ([]types.InternalMessage, error) {
	if err := types.ValidateContractAddress(req.ContractAddress); err != nil {
		return nil, err
	}
	if err := types.ValidateQueryPagination(req.Pagination); err != nil {
		return nil, err
	}
	queue := make([]types.InternalMessage, 0)
	contract, found := findContract(k.genesis.State.Contracts, req.ContractAddress)
	if !found {
		return nil, errors.New(types.ErrContractNotFound + ": contract not found")
	}
	if err := types.EnsureContractLifecycleAction(contract, types.ContractLifecycleActionQuery); err != nil {
		return nil, err
	}
	for _, msg := range k.genesis.State.Normalize().InternalMessages {
		if msg.SourceContractUser == req.ContractAddress || msg.DestinationAccount == req.ContractAddress {
			queue = append(queue, msg)
			if uint32(len(queue)) == req.Pagination.Limit {
				break
			}
		}
	}
	return queue, nil
}

func (k Keeper) ContractEvents(req types.QueryContractEventsRequest) error {
	if err := types.ValidateContractAddress(req.ContractAddress); err != nil {
		return err
	}
	return types.ValidateQueryPagination(req.Pagination)
}

func (k Keeper) ContractStateRoot(req types.QueryContractStateRootRequest) (string, error) {
	if err := types.ValidateContractAddress(req.ContractAddress); err != nil {
		return "", err
	}
	contract, found := findContract(k.genesis.State.Contracts, req.ContractAddress)
	if !found {
		return "", errors.New(types.ErrContractNotFound + ": contract not found")
	}
	if err := types.EnsureContractLifecycleAction(contract, types.ContractLifecycleActionProofQuery); err != nil {
		return "", err
	}
	return contract.StateRoot, nil
}

func (k *Keeper) InstantiateContract(msg types.MsgInstantiateContract) (types.InstantiateContractResponse, error) {
	return k.instantiateContract(k.runtimeCtx, msg)
}

func (k *Keeper) instantiateContract(ctx context.Context, msg types.MsgInstantiateContract) (types.InstantiateContractResponse, error) {
	if !k.genesis.Params.Enabled {
		return types.InstantiateContractResponse{}, errors.New(types.ErrExecutionFailed + ": module disabled")
	}
	if err := types.ValidateUserFacingAEAddress("contract creator", msg.Creator); err != nil {
		return types.InstantiateContractResponse{}, err
	}
	if err := k.ensureActiveWallet(ctx, msg.Creator, "contract instantiate"); err != nil {
		return types.InstantiateContractResponse{}, err
	}
	if msg.Height == 0 {
		return types.InstantiateContractResponse{}, errors.New("contract instantiate height must be positive")
	}
	code, found := findCode(k.genesis.State.Codes, msg.CodeID)
	if !found {
		return types.InstantiateContractResponse{}, errors.New(types.ErrContractNotFound + ": contract code not found")
	}
	if code.Owner != msg.Creator {
		return types.InstantiateContractResponse{}, errors.New(types.ErrUnauthorized + ": contract instantiate requires code owner")
	}
	stateInit, data, funds, err := k.stateInitForInstantiate(msg, code)
	if err != nil {
		return types.InstantiateContractResponse{}, err
	}
	admin := msg.Admin
	if admin == "" {
		admin = stateInit.Owner
	}
	if err := types.ValidateUserFacingAEAddress("contract admin", admin); err != nil {
		return types.InstantiateContractResponse{}, err
	}
	user, raw, err := types.DeriveContractAddressFromStateInit(msg.ChainID, msg.Namespace, msg.Creator, stateInit, k.genesis.Params)
	if err != nil {
		return types.InstantiateContractResponse{}, err
	}
	if _, found := findContract(k.genesis.State.Contracts, user); found {
		return types.InstantiateContractResponse{}, errors.New(types.ErrContractNotFound + ": contract address already exists")
	}
	storageBytes, err := contractStorageBytesForCode(code, data)
	if err != nil {
		return types.InstantiateContractResponse{}, err
	}
	if msg.StorageBytes != 0 && msg.StorageBytes != storageBytes {
		return types.InstantiateContractResponse{}, errors.New(types.ErrStorageRent + ": contract storage must equal code bytes plus data bytes")
	}
	if storageBytes > k.genesis.Params.MaxContractStorageBytes {
		return types.InstantiateContractResponse{}, errors.New(types.ErrStorageRent + ": contract storage exceeds configured limit")
	}
	stateInitHash, err := types.HashStateInit(stateInit)
	if err != nil {
		return types.InstantiateContractResponse{}, err
	}
	schemaVersion := msg.SchemaVersion
	if schemaVersion == 0 {
		schemaVersion = 1
	}
	contract := types.Contract{
		AddressUser:			user,
		AddressRaw:			raw,
		CodeID:				msg.CodeID,
		CodeHash:			code.CodeHash,
		StateInitHash:			stateInitHash,
		StateInit:			stateInit,
		Creator:			msg.Creator,
		Owner:				stateInit.Owner,
		Admin:				admin,
		Upgradeable:			msg.Upgradeable,
		SystemOwned:			msg.SystemOwned,
		StorageSchemaVersion:		schemaVersion,
		InitMsg:			append([]byte(nil), data...),
		Data:				append([]byte(nil), data...),
		Balance:			funds,
		Status:				types.ContractStatusActive,
		StorageBytes:			storageBytes,
		LastStorageChargeHeight:	msg.Height,
		LogicalTime:			1,
		CreatedHeight:			msg.Height,
		UpdatedHeight:			msg.Height,
	}
	contract.StateRoot = types.ComputeContractStateRoot(contract)
	next := k.genesis
	next.State.Contracts = append(next.State.Contracts, contract)
	next.State.Receipts = append(next.State.Receipts, newContractReceipt(contract.AddressUser, msg.Creator, "deploy", types.ExitCodeOK, funds, 0, contract.LogicalTime, msg.Height))
	next = types.RefreshStateRoot(next)
	if err := next.Validate(); err != nil {
		return types.InstantiateContractResponse{}, err
	}
	initialStorageFee := storageBytes * k.storageRentPerByteBlock()
	if err := k.collectRentPayment(ctx, msg.Creator, initialStorageFee); err != nil {
		return types.InstantiateContractResponse{}, err
	}
	k.genesis = next
	return types.InstantiateContractResponse{
		ContractAddressUser:	user,
		ContractAddressRaw:	raw,
		Owner:			contract.Owner,
		Admin:			contract.Admin,
		Balance:		contract.Balance,
		Events: []types.ContractEvent{{
			Type:		types.EventTypeContractInstantiated,
			Actor:		msg.Creator,
			Contract:	user,
			Amount:		funds,
			InternalRaw:	raw,
		}},
	}, nil
}

func (k *Keeper) InstantiateContractState(ctx context.Context, msg types.MsgInstantiateContract) (types.InstantiateContractResponse, error) {
	res, err := k.instantiateContract(ctx, msg)
	if err != nil {
		return types.InstantiateContractResponse{}, err
	}
	return res, k.writeGenesis(ctx)
}

func (k *Keeper) UpgradeContractCode(msg types.MsgUpgradeContractCode) (types.ContractReceipt, error) {
	if err := msg.ValidateBasic(k.genesis.Params); err != nil {
		return types.ContractReceipt{}, err
	}
	idx, contract, found := findContractWithIndex(k.genesis.State.Contracts, msg.ContractAddress)
	if !found {
		return types.ContractReceipt{}, errors.New(types.ErrContractNotFound + ": contract not found")
	}
	if err := types.EnsureContractLifecycleAction(contract, types.ContractLifecycleActionUpgradeMigrate); err != nil {
		return types.ContractReceipt{}, err
	}
	if err := k.authorizeContractUpgradeActor(contract, msg.Actor); err != nil {
		return types.ContractReceipt{}, err
	}
	if !contract.Upgradeable || contract.UpgradesDisabled {
		return types.ContractReceipt{}, errors.New(types.ErrUnauthorized + ": contract is immutable")
	}
	code, found := findCode(k.genesis.State.Codes, msg.NewCodeID)
	if !found {
		return types.ContractReceipt{}, errors.New(types.ErrContractNotFound + ": upgrade code not found")
	}
	if code.CodeHash != contract.CodeHash && strings.TrimSpace(msg.MigrationHandler) == "" {
		return types.ContractReceipt{}, errors.New(types.ErrExecutionFailed + ": code hash change requires migration handler")
	}
	nextContract := contract
	nextContract.CodeID = code.CodeID
	nextContract.CodeHash = code.CodeHash
	storageBytes, err := contractStorageBytesForCode(code, nextContract.Data)
	if err != nil {
		return types.ContractReceipt{}, err
	}
	if storageBytes > k.genesis.Params.MaxContractStorageBytes {
		return types.ContractReceipt{}, errors.New(types.ErrStorageRent + ": contract storage exceeds configured limit")
	}
	if storageBytes > contract.StorageBytes {
		diff := storageBytes - contract.StorageBytes
		extraFee := diff * k.storageRentPerByteBlock()
		if err := k.collectRentPayment(k.runtimeCtx, msg.Actor, extraFee); err != nil {
			return types.ContractReceipt{}, err
		}
	}
	nextContract.StorageBytes = storageBytes
	nextContract.LogicalTime++
	nextContract.UpdatedHeight = msg.Height
	nextContract.StateRoot = types.ComputeContractStateRoot(nextContract)
	receipt := newContractReceipt(nextContract.AddressUser, receiptActor(contract, msg.Actor), "upgrade_code", types.ExitCodeOK, 0, 0, nextContract.LogicalTime, msg.Height)
	next := k.genesis
	next.State.Contracts[idx] = nextContract
	next.State.Receipts = append(next.State.Receipts, receipt)
	next = types.RefreshStateRoot(next)
	if err := next.Validate(); err != nil {
		return types.ContractReceipt{}, err
	}
	k.genesis = next
	return receipt, nil
}

func (k *Keeper) MigrateContractState(msg types.MsgMigrateContractState) (types.ContractReceipt, error) {
	if err := msg.ValidateBasic(k.genesis.Params); err != nil {
		return types.ContractReceipt{}, err
	}
	idx, contract, found := findContractWithIndex(k.genesis.State.Contracts, msg.ContractAddress)
	if !found {
		return types.ContractReceipt{}, errors.New(types.ErrContractNotFound + ": contract not found")
	}
	if err := types.EnsureContractLifecycleAction(contract, types.ContractLifecycleActionUpgradeMigrate); err != nil {
		return types.ContractReceipt{}, err
	}
	if err := k.authorizeContractUpgradeActor(contract, msg.Actor); err != nil {
		return types.ContractReceipt{}, err
	}
	if !contract.Upgradeable || contract.UpgradesDisabled {
		return types.ContractReceipt{}, errors.New(types.ErrUnauthorized + ": contract is immutable")
	}
	if contract.StorageSchemaVersion != msg.FromSchemaVersion {
		return types.ContractReceipt{}, errors.New(types.ErrExecutionFailed + ": contract migration schema version mismatch")
	}
	nextContract := contract
	data, err := applyContractMigration(nextContract.Data, msg.MigrationHandler, msg.Payload)
	if err != nil {
		return types.ContractReceipt{}, err
	}
	nextContract.Data = data
	nextContract.StorageSchemaVersion = msg.ToSchemaVersion
	storageBytes, err := k.contractStorageBytes(nextContract)
	if err != nil {
		return types.ContractReceipt{}, err
	}
	if storageBytes > k.genesis.Params.MaxContractStorageBytes {
		return types.ContractReceipt{}, errors.New(types.ErrStorageRent + ": migrated contract storage exceeds configured limit")
	}
	if storageBytes > contract.StorageBytes {
		diff := storageBytes - contract.StorageBytes
		extraFee := diff * k.storageRentPerByteBlock()
		if err := k.collectRentPayment(k.runtimeCtx, msg.Actor, extraFee); err != nil {
			return types.ContractReceipt{}, err
		}
	}
	nextContract.StorageBytes = storageBytes
	nextContract.LogicalTime++
	nextContract.UpdatedHeight = msg.Height
	nextContract.StateRoot = types.ComputeContractStateRoot(nextContract)
	receipt := newContractReceipt(nextContract.AddressUser, receiptActor(contract, msg.Actor), "migrate_state", types.ExitCodeOK, msg.ToSchemaVersion, 0, nextContract.LogicalTime, msg.Height)
	next := k.genesis
	next.State.Contracts[idx] = nextContract
	next.State.Receipts = append(next.State.Receipts, receipt)
	next = types.RefreshStateRoot(next)
	if err := next.Validate(); err != nil {
		return types.ContractReceipt{}, err
	}
	k.genesis = next
	return receipt, nil
}

func (k *Keeper) SetContractAdmin(msg types.MsgSetContractAdmin) (types.ContractReceipt, error) {
	if err := msg.ValidateBasic(k.genesis.Params); err != nil {
		return types.ContractReceipt{}, err
	}
	idx, contract, found := findContractWithIndex(k.genesis.State.Contracts, msg.ContractAddress)
	if !found {
		return types.ContractReceipt{}, errors.New(types.ErrContractNotFound + ": contract not found")
	}
	if err := types.EnsureContractLifecycleAction(contract, types.ContractLifecycleActionUpgradeMigrate); err != nil {
		return types.ContractReceipt{}, err
	}
	if err := k.authorizeContractUpgradeActor(contract, msg.Actor); err != nil {
		return types.ContractReceipt{}, err
	}
	contract.Admin = msg.NewAdmin
	contract.LogicalTime++
	contract.UpdatedHeight = msg.Height
	contract.StateRoot = types.ComputeContractStateRoot(contract)
	receipt := newContractReceipt(contract.AddressUser, receiptActor(contract, msg.Actor), "set_admin", types.ExitCodeOK, 0, 0, contract.LogicalTime, msg.Height)
	next := k.genesis
	next.State.Contracts[idx] = contract
	next.State.Receipts = append(next.State.Receipts, receipt)
	next = types.RefreshStateRoot(next)
	if err := next.Validate(); err != nil {
		return types.ContractReceipt{}, err
	}
	k.genesis = next
	return receipt, nil
}

func (k *Keeper) DisableContractUpgrades(msg types.MsgDisableContractUpgrades) (types.ContractReceipt, error) {
	if err := msg.ValidateBasic(k.genesis.Params); err != nil {
		return types.ContractReceipt{}, err
	}
	idx, contract, found := findContractWithIndex(k.genesis.State.Contracts, msg.ContractAddress)
	if !found {
		return types.ContractReceipt{}, errors.New(types.ErrContractNotFound + ": contract not found")
	}
	if err := types.EnsureContractLifecycleAction(contract, types.ContractLifecycleActionUpgradeMigrate); err != nil {
		return types.ContractReceipt{}, err
	}
	if err := k.authorizeContractUpgradeActor(contract, msg.Actor); err != nil {
		return types.ContractReceipt{}, err
	}
	contract.Upgradeable = false
	contract.UpgradesDisabled = true
	contract.LogicalTime++
	contract.UpdatedHeight = msg.Height
	contract.StateRoot = types.ComputeContractStateRoot(contract)
	receipt := newContractReceipt(contract.AddressUser, receiptActor(contract, msg.Actor), "disable_upgrades", types.ExitCodeOK, 0, 0, contract.LogicalTime, msg.Height)
	next := k.genesis
	next.State.Contracts[idx] = contract
	next.State.Receipts = append(next.State.Receipts, receipt)
	next = types.RefreshStateRoot(next)
	if err := next.Validate(); err != nil {
		return types.ContractReceipt{}, err
	}
	k.genesis = next
	return receipt, nil
}

func (k Keeper) stateInitForInstantiate(msg types.MsgInstantiateContract, code types.CodeRecord) (types.StateInit, []byte, uint64, error) {
	if msg.StateInit == nil {
		stateInit := types.NewStateInit(msg.Creator, code.CodeHash, msg.InitMsg, msg.Salt, msg.Funds).Normalize()
		if err := stateInit.Validate(k.genesis.Params); err != nil {
			return types.StateInit{}, nil, 0, err
		}
		return stateInit, append([]byte(nil), stateInit.InitData...), stateInit.InitialBalanceNAET, nil
	}
	stateInit := msg.StateInit.Normalize()
	if err := stateInit.Validate(k.genesis.Params); err != nil {
		return types.StateInit{}, nil, 0, err
	}
	if stateInit.CodeID != msg.CodeID {
		return types.StateInit{}, nil, 0, errors.New("state init code id must match instantiate code id")
	}
	if stateInit.CodeHash != code.CodeHash {
		return types.StateInit{}, nil, 0, errors.New("state init code hash must match stored code")
	}
	if len(msg.InitMsg) != 0 && !bytes.Equal(msg.InitMsg, stateInit.InitData) {
		return types.StateInit{}, nil, 0, errors.New("state init data must match instantiate init message")
	}
	if msg.Funds != 0 && msg.Funds != stateInit.InitialBalanceNAET {
		return types.StateInit{}, nil, 0, errors.New("state init initial balance must match instantiate funds")
	}
	if msg.Salt != "" && msg.Salt != stateInit.Salt && !bytes.Equal([]byte(msg.Salt), stateInit.SaltBytesForAddress()) {
		return types.StateInit{}, nil, 0, errors.New("state init salt must match instantiate salt")
	}
	return stateInit, append([]byte(nil), stateInit.InitData...), stateInit.InitialBalanceNAET, nil
}

func (k *Keeper) ExecuteContract(msg types.MsgExecuteContract) (types.ExecuteContractResponse, error) {
	return k.executeContract(k.runtimeCtx, msg)
}

func (k *Keeper) executeContract(ctx context.Context, msg types.MsgExecuteContract) (types.ExecuteContractResponse, error) {
	if !k.genesis.Params.Enabled {
		return types.ExecuteContractResponse{}, errors.New(types.ErrExecutionFailed + ": module disabled")
	}
	if err := types.ValidateUserFacingAEAddress("contract execute sender", msg.Sender); err != nil {
		return types.ExecuteContractResponse{}, err
	}
	if err := k.ensureActiveWallet(ctx, msg.Sender, "contract execute"); err != nil {
		return types.ExecuteContractResponse{}, err
	}
	if msg.Height == 0 {
		return types.ExecuteContractResponse{}, errors.New("contract execute height must be positive")
	}
	idx, contract, found := findContractWithIndex(k.genesis.State.Contracts, msg.ContractAddress)
	if !found {
		return types.ExecuteContractResponse{}, errors.New(types.ErrContractNotFound + ": contract not found")
	}
	if err := types.EnsureContractLifecycleAction(contract, types.ContractLifecycleActionExecuteExternal); err != nil {
		return types.ExecuteContractResponse{}, err
	}
	contract, err := k.chargeContractRentAt(ctx, idx, contract, msg.Height)
	if err != nil {
		return types.ExecuteContractResponse{}, errors.New(types.ErrStorageRent + ": contract has storage rent debt")
	}
	balance, err := checkedAdd(contract.Balance, msg.Funds, "contract balance overflow")
	if err != nil {
		return types.ExecuteContractResponse{}, err
	}
	contract.Balance = balance
	contract.Data = append([]byte(nil), msg.Msg...)
	contract.LogicalTime++
	storageBytes, err := k.contractStorageBytes(contract)
	if err != nil {
		return types.ExecuteContractResponse{}, err
	}
	if storageBytes > k.genesis.Params.MaxContractStorageBytes {
		return types.ExecuteContractResponse{}, errors.New(types.ErrStorageRent + ": contract storage exceeds configured limit")
	}
	contract.StorageBytes = storageBytes
	contract.UpdatedHeight = msg.Height
	contract.StateRoot = types.ComputeContractStateRoot(contract)
	next := k.genesis
	next.State.Contracts[idx] = contract
	next.State.Receipts = append(next.State.Receipts, newContractReceipt(contract.AddressUser, msg.Sender, "execute", types.ExitCodeOK, msg.Funds, 0, contract.LogicalTime, msg.Height))
	next = types.RefreshStateRoot(next)
	if err := next.Validate(); err != nil {
		return types.ExecuteContractResponse{}, err
	}
	k.genesis = next
	return types.ExecuteContractResponse{
		ContractAddressUser:	contract.AddressUser,
		Owner:			contract.Owner,
		Balance:		contract.Balance,
		Events: []types.ContractEvent{{
			Type:		types.EventTypeContractExecuted,
			Actor:		msg.Sender,
			Contract:	contract.AddressUser,
			Amount:		msg.Funds,
			InternalRaw:	contract.AddressRaw,
		}},
	}, nil
}

func (k *Keeper) ExecuteContractState(ctx context.Context, msg types.MsgExecuteContract) (types.ExecuteContractResponse, error) {
	res, err := k.executeContract(ctx, msg)
	if err != nil {
		return types.ExecuteContractResponse{}, err
	}
	return res, k.writeGenesis(ctx)
}

func (k *Keeper) TopUpContract(msg types.MsgTopUpContract) (types.Contract, error) {
	if err := types.ValidateUserFacingAEAddress("contract top-up sender", msg.Sender); err != nil {
		return types.Contract{}, err
	}
	if msg.Amount == 0 || msg.Height == 0 {
		return types.Contract{}, errors.New("contract top-up amount and height must be positive")
	}
	idx, contract, found := findContractWithIndex(k.genesis.State.Contracts, msg.ContractAddress)
	if !found {
		return types.Contract{}, errors.New(types.ErrContractNotFound + ": contract not found")
	}
	if err := types.EnsureContractLifecycleAction(contract, types.ContractLifecycleActionReceiveTopUp); err != nil {
		return types.Contract{}, err
	}
	balance, err := checkedAdd(contract.Balance, msg.Amount, "contract top-up balance overflow")
	if err != nil {
		return types.Contract{}, err
	}
	contract.Balance = balance
	contract.UpdatedHeight = msg.Height
	next := k.genesis
	next.State.Contracts[idx] = contract
	next = types.RefreshStateRoot(next)
	if err := next.Validate(); err != nil {
		return types.Contract{}, err
	}
	k.genesis = next
	return contract, nil
}

func (k *Keeper) TopUpContractState(ctx context.Context, msg types.MsgTopUpContract) (types.Contract, error) {
	contract, err := k.TopUpContract(msg)
	if err != nil {
		return types.Contract{}, err
	}
	return contract, k.writeGenesis(ctx)
}

func (k *Keeper) PayContractStorageDebt(msg types.MsgPayContractStorageDebt) (types.Contract, error) {
	if err := types.ValidateUserFacingAEAddress("contract rent payer", msg.Sender); err != nil {
		return types.Contract{}, err
	}
	if msg.Amount == 0 || msg.Height == 0 {
		return types.Contract{}, errors.New("contract storage debt payment amount and height must be positive")
	}
	idx, contract, found := findContractWithIndex(k.genesis.State.Contracts, msg.ContractAddress)
	if !found {
		return types.Contract{}, errors.New(types.ErrContractNotFound + ": contract not found")
	}
	if err := types.EnsureContractLifecycleAction(contract, types.ContractLifecycleActionPayRentDebt); err != nil {
		return types.Contract{}, err
	}
	if msg.Amount >= contract.StorageRentDebt {
		contract.StorageRentDebt = 0
	} else {
		contract.StorageRentDebt -= msg.Amount
	}
	contract.UpdatedHeight = msg.Height
	next := k.genesis
	next.State.Contracts[idx] = contract
	next = types.RefreshStateRoot(next)
	if err := next.Validate(); err != nil {
		return types.Contract{}, err
	}
	k.genesis = next
	return contract, nil
}

func (k *Keeper) PayContractStorageDebtState(ctx context.Context, msg types.MsgPayContractStorageDebt) (types.Contract, error) {
	contract, err := k.PayContractStorageDebt(msg)
	if err != nil {
		return types.Contract{}, err
	}
	return contract, k.writeGenesis(ctx)
}

func (k *Keeper) UnfreezeContract(msg types.MsgUnfreezeContract) (types.Contract, error) {
	return k.unfreezeContract(k.runtimeCtx, msg)
}

func (k *Keeper) UnfreezeContractState(ctx context.Context, msg types.MsgUnfreezeContract) (types.Contract, error) {
	contract, err := k.unfreezeContract(ctx, msg)
	if err != nil {
		return types.Contract{}, err
	}
	return contract, k.writeGenesis(ctx)
}

func (k *Keeper) unfreezeContract(ctx context.Context, msg types.MsgUnfreezeContract) (types.Contract, error) {
	if err := types.ValidateUserFacingAEAddress("contract unfreeze sender", msg.Sender); err != nil {
		return types.Contract{}, err
	}
	if err := k.ensureActiveWallet(ctx, msg.Sender, "contract unfreeze"); err != nil {
		return types.Contract{}, err
	}
	if msg.Height == 0 {
		return types.Contract{}, errors.New("contract unfreeze height must be positive")
	}
	idx, contract, found := findContractWithIndex(k.genesis.State.Contracts, msg.ContractAddress)
	if !found {
		return types.Contract{}, errors.New(types.ErrContractNotFound + ": contract not found")
	}
	if err := types.EnsureContractLifecycleAction(contract, types.ContractLifecycleActionUnfreeze); err != nil {
		return types.Contract{}, err
	}
	if contract.StorageRentDebt > 0 {
		return types.Contract{}, errors.New(types.ErrStorageRent + ": contract storage rent debt must be paid before unfreeze")
	}
	contract.Status = types.ContractStatusActive
	contract.LastStorageChargeHeight = msg.Height
	contract.UpdatedHeight = msg.Height
	next := k.genesis
	next.State.Contracts[idx] = contract
	next = types.RefreshStateRoot(next)
	if err := next.Validate(); err != nil {
		return types.Contract{}, err
	}
	k.genesis = next
	return contract, nil
}

func (k *Keeper) GrantNativeStakingCapability(msg types.MsgGrantNativeStakingCapability) (types.ContractCapability, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.ContractCapability{}, err
	}
	if msg.Height == 0 {
		return types.ContractCapability{}, errors.New("contract capability height must be positive")
	}
	if _, found := findContract(k.genesis.State.Contracts, msg.ContractAddressUser); !found {
		return types.ContractCapability{}, errors.New(types.ErrContractNotFound + ": contract not found")
	}
	capability := types.ContractCapability{
		ContractAddressUser:	msg.ContractAddressUser,
		ContractAddressRaw:	msg.ContractAddressRaw,
		Capability:		types.NativeStakingCapability,
		PoolID:			msg.PoolID,
		GrantedHeight:		msg.Height,
	}
	if err := capability.Validate(); err != nil {
		return types.ContractCapability{}, err
	}
	next := k.genesis
	next.State.StakingCapabilities = upsertCapability(next.State.StakingCapabilities, capability)
	next = types.RefreshStateRoot(next)
	if err := next.Validate(); err != nil {
		return types.ContractCapability{}, err
	}
	k.genesis = next
	return capability, nil
}

func (k *Keeper) InjectNativeStaking(msg types.MsgInjectNativeStaking) (types.NativeStakingInjectionRecord, error) {
	if msg.Amount == 0 || msg.Height == 0 {
		return types.NativeStakingInjectionRecord{}, errors.New("native staking injection amount and height must be positive")
	}
	if err := types.ValidateAddressPair("native staking caller contract", msg.CallerContractUser, msg.CallerContractRaw); err != nil {
		return types.NativeStakingInjectionRecord{}, err
	}
	idx, contract, found := findContractWithIndex(k.genesis.State.Contracts, msg.CallerContractUser)
	if !found {
		return types.NativeStakingInjectionRecord{}, errors.New(types.ErrContractNotFound + ": contract not found")
	}
	if err := types.EnsureContractLifecycleAction(contract, types.ContractLifecycleActionExecuteExternal); err != nil {
		return types.NativeStakingInjectionRecord{}, err
	}
	if _, err := k.chargeContractRentAt(k.runtimeCtx, idx, contract, msg.Height); err != nil {
		return types.NativeStakingInjectionRecord{}, errors.New(types.ErrStorageRent + ": contract has storage rent debt")
	}
	if !hasCapability(k.genesis.State.StakingCapabilities, msg.CallerContractUser, msg.PoolID) {
		return types.NativeStakingInjectionRecord{}, errors.New(types.ErrUnauthorized + ": contract lacks native staking capability")
	}
	record := types.NativeStakingInjectionRecord{
		ContractAddressUser:	msg.CallerContractUser,
		ContractAddressRaw:	msg.CallerContractRaw,
		PoolID:			msg.PoolID,
		Amount:			msg.Amount,
		Height:			msg.Height,
	}
	next := k.genesis
	next.State.NativeStakingInjects = append(next.State.NativeStakingInjects, record)
	next = types.RefreshStateRoot(next)
	if err := next.Validate(); err != nil {
		return types.NativeStakingInjectionRecord{}, err
	}
	k.genesis = next
	return record, nil
}

func (k *Keeper) ReceiveInternalMessage(msg types.MsgReceiveInternalMessage) (types.InternalMessage, error) {
	record := types.InternalMessage{
		SourceContractUser:	msg.SourceContractUser,
		DestinationAccount:	msg.DestinationAccount,
		Funds:			msg.Funds,
		Opcode:			msg.Opcode,
		QueryID:		msg.QueryID,
		Body:			append([]byte(nil), msg.Body...),
		Bounce:			msg.Bounce,
		Deadline:		msg.Deadline,
		GasLimit:		msg.GasLimit,
		LogicalTime:		msg.LogicalTime,
		MessageID:		msg.MessageID,
		Height:			msg.Height,
	}
	if record.LogicalTime == 0 {
		record.LogicalTime = msg.Height
	}
	if record.MessageID == "" {
		record.MessageID = types.ComputeInternalMessageID(record)
	}
	if err := record.Validate(); err != nil {
		return types.InternalMessage{}, err
	}
	idx, contract, found := findContractWithIndex(k.genesis.State.Contracts, msg.SourceContractUser)
	if !found {
		return types.InternalMessage{}, errors.New(types.ErrContractNotFound + ": source contract not found")
	}
	if err := types.EnsureContractLifecycleAction(contract, types.ContractLifecycleActionEmitInternalMessage); err != nil {
		return types.InternalMessage{}, err
	}
	if _, destination, found := findContractWithIndex(k.genesis.State.Contracts, msg.DestinationAccount); found {
		if err := types.EnsureContractLifecycleAction(destination, types.ContractLifecycleActionReceiveInternal); err != nil {
			return types.InternalMessage{}, err
		}
	}
	if _, err := k.chargeContractRentAt(k.runtimeCtx, idx, contract, msg.Height); err != nil {
		return types.InternalMessage{}, errors.New(types.ErrStorageRent + ": contract has storage rent debt")
	}
	next := k.genesis
	next.State.InternalMessages = append(next.State.InternalMessages, record)
	next.State.Receipts = append(next.State.Receipts, newContractReceipt(record.SourceContractUser, record.SourceContractUser, "internal_message_queued", types.ExitCodeOK, record.Funds, record.GasLimit, record.LogicalTime, record.Height))
	next = types.RefreshStateRoot(next)
	if err := next.Validate(); err != nil {
		return types.InternalMessage{}, err
	}
	k.genesis = next
	return record, nil
}

func newContractReceipt(contractAddress, actor, operation string, exitCode uint32, amount, gasUsed, logicalTime, height uint64) types.ContractReceipt {
	receipt := types.ContractReceipt{
		ContractAddress:	contractAddress,
		Actor:			actor,
		Operation:		operation,
		ExitCode:		exitCode,
		Amount:			amount,
		GasUsed:		gasUsed,
		LogicalTime:		logicalTime,
		Height:			height,
	}
	receipt.ReceiptID = types.ComputeContractReceiptID(receipt)
	return receipt
}

func (k Keeper) AssetOwner(req types.QueryAssetOwnerRequest) (types.QueryAssetOwnerResponse, error) {
	if req.AssetType == "" {
		return types.QueryAssetOwnerResponse{}, fmt.Errorf("contract asset type must not be empty")
	}
	if err := types.ValidateUserFacingAEAddress("asset contract address", req.ContractAddressUser); err != nil {
		return types.QueryAssetOwnerResponse{}, err
	}
	if req.AssetID == "" {
		return types.QueryAssetOwnerResponse{}, errors.New("asset id is required")
	}
	for _, asset := range k.genesis.State.AssetOwnership {
		if asset.AssetType == req.AssetType && asset.ContractAddressUser == req.ContractAddressUser && asset.AssetID == req.AssetID {
			return types.QueryAssetOwnerResponse{Owner: asset.Owner, Found: true}, nil
		}
	}
	return types.QueryAssetOwnerResponse{}, nil
}

func (k *Keeper) SetAssetOwner(record types.AssetOwnershipRecord) error {
	if err := record.Validate(); err != nil {
		return err
	}
	next := k.genesis
	next.State.AssetOwnership = upsertAsset(next.State.AssetOwnership, record)
	next = types.RefreshStateRoot(next)
	if err := next.Validate(); err != nil {
		return err
	}
	k.genesis = next
	return nil
}

func (k *Keeper) ensureActiveWallet(ctx context.Context, address string, operation string) error {
	if k.accountStatusReader == nil {
		return nil
	}
	status, found, err := k.accountStatusReader.AccountStatus(ctx, address)
	if err != nil {
		return err
	}
	if !found || status == accountStatusInactive {
		return fmt.Errorf("%s: %s", operation, types.ErrAccountInactive)
	}
	if status == accountStatusFrozen {
		return fmt.Errorf("%s: %s", operation, types.ErrAccountFrozen)
	}
	if status != accountStatusActive {
		return fmt.Errorf("%s: unsupported account status %q", operation, status)
	}
	return nil
}

func (k *Keeper) chargeContractRentAt(ctx context.Context, idx int, contract types.Contract, height uint64) (types.Contract, error) {
	prevBalance := contract.Balance
	contract, changed, err := k.chargeRent(contract, height)
	if err != nil {
		return types.Contract{}, err
	}
	if contract.StorageRentDebt > 0 {
		contract.Status = k.storageRentFrozenStatus(contract)
		if err := k.persistContractAt(idx, contract); err != nil {
			return types.Contract{}, err
		}
		return contract, errors.New(types.ErrStorageRent + ": contract has storage rent debt")
	}
	if changed {
		rentCharged := prevBalance - contract.Balance
		if err := k.chargeRentToReserve(ctx, contract, rentCharged); err != nil {
			return types.Contract{}, err
		}
		if err := k.persistContractAt(idx, contract); err != nil {
			return types.Contract{}, err
		}
	}
	return contract, nil
}

func (k Keeper) storageRentFrozenStatus(contract types.Contract) string {
	if contract.Status == types.ContractStatusFrozenLimited || hasAnyNativeStakingCapability(k.genesis.State.StakingCapabilities, contract.AddressUser) {
		return types.ContractStatusFrozenLimited
	}
	return types.ContractStatusFrozen
}

func (k *Keeper) persistContractAt(idx int, contract types.Contract) error {
	if idx < 0 || idx >= len(k.genesis.State.Contracts) {
		return errors.New(types.ErrContractNotFound + ": contract index out of bounds")
	}
	next := k.genesis
	next.State.Contracts[idx] = contract
	next = types.RefreshStateRoot(next)
	if err := next.Validate(); err != nil {
		return err
	}
	k.genesis = next
	return nil
}

func (k Keeper) storageRentPerByteBlock() uint64 {
	if k.storageRentRateProvider != nil {
		return k.storageRentRateProvider.StorageRentRatePerByteBlock()
	}
	return k.genesis.Params.StorageRentPerByteBlock
}

func (k Keeper) chargeRent(contract types.Contract, height uint64) (types.Contract, bool, error) {
	if height < contract.LastStorageChargeHeight {
		return types.Contract{}, false, errors.New(types.ErrStorageRent + ": contract storage rent height must be monotonic")
	}
	if height <= contract.LastStorageChargeHeight || contract.StorageBytes == 0 || k.storageRentPerByteBlock() == 0 {
		return contract, false, nil
	}
	blocks := height - contract.LastStorageChargeHeight
	charge, err := checkedMul(blocks, contract.StorageBytes, "contract storage rent overflow")
	if err != nil {
		return types.Contract{}, false, err
	}
	charge, err = checkedMul(charge, k.storageRentPerByteBlock(), "contract storage rent overflow")
	if err != nil {
		return types.Contract{}, false, err
	}
	if contract.Balance >= charge {
		contract.Balance -= charge
	} else {
		unpaid := charge - contract.Balance
		debt, err := checkedAdd(contract.StorageRentDebt, unpaid, "contract storage rent debt overflow")
		if err != nil {
			return types.Contract{}, false, err
		}
		contract.StorageRentDebt = debt
		contract.Balance = 0
	}
	contract.LastStorageChargeHeight = height
	return contract, true, nil
}

func (k Keeper) contractStorageBytes(contract types.Contract) (uint64, error) {
	code, found := findCode(k.genesis.State.Codes, contract.CodeID)
	if !found {
		return 0, errors.New(types.ErrContractNotFound + ": contract code not found")
	}
	return contractStorageBytesForCode(code, contract.Data)
}

func (k *Keeper) collectRentPayment(ctx context.Context, payer string, amount uint64) error {
	if k.bankKeeper == nil || ctx == nil {
		return nil
	}
	payerAddr, err := aetraaddress.ParseAccAddress(payer)
	if err != nil {
		return err
	}
	coin := sdk.NewCoins(sdk.NewCoin(storageRentBaseDenom, sdkmath.NewInt(int64(amount))))
	return k.bankKeeper.SendCoinsFromAccountToModule(ctx, payerAddr, storageRentReserveModule, coin)
}

func (k *Keeper) chargeRentToReserve(ctx context.Context, contract types.Contract, amount uint64) error {
	if k.bankKeeper == nil || amount == 0 {
		return nil
	}
	return k.collectRentPayment(ctx, contract.Creator, amount)
}

func (k Keeper) authorizeContractUpgradeActor(contract types.Contract, actor string) error {
	actor = strings.TrimSpace(actor)
	if contract.SystemOwned {
		if actor != k.genesis.Params.Authority {
			return errors.New(types.ErrUnauthorized + ": system contract upgrade requires governance authority")
		}
		return nil
	}
	if err := types.ValidateUserFacingAEAddress("contract upgrade actor", actor); err != nil {
		return err
	}
	if actor != contract.Admin {
		return errors.New(types.ErrUnauthorized + ": contract upgrade requires admin")
	}
	return nil
}

func receiptActor(contract types.Contract, actor string) string {
	if contract.SystemOwned {
		return ""
	}
	return actor
}

func applyContractMigration(current []byte, handler string, payload []byte) ([]byte, error) {
	switch strings.TrimSpace(handler) {
	case "schema_only":
		return append([]byte(nil), current...), nil
	case "replace":
		return append([]byte(nil), payload...), nil
	case "append":
		out := append([]byte(nil), current...)
		out = append(out, payload...)
		return out, nil
	case "fail":
		return nil, errors.New(types.ErrExecutionFailed + ": contract migration handler failed")
	default:
		return nil, errors.New(types.ErrExecutionFailed + ": unsupported contract migration handler")
	}
}

func contractStorageBytesForCode(code types.CodeRecord, data []byte) (uint64, error) {
	dataBytes := uint64(len(data))
	return checkedAdd(code.CodeBytes, dataBytes, "contract storage size overflow")
}

func checkedAdd(left, right uint64, message string) (uint64, error) {
	if left > math.MaxUint64-right {
		return 0, errors.New(message)
	}
	return left + right, nil
}

func checkedMul(left, right uint64, message string) (uint64, error) {
	if left != 0 && right > math.MaxUint64/left {
		return 0, errors.New(message)
	}
	return left * right, nil
}

func upsertCode(codes []types.CodeRecord, code types.CodeRecord) []types.CodeRecord {
	out := append([]types.CodeRecord(nil), codes...)
	for i := range out {
		if out[i].CodeID == code.CodeID {
			out[i] = code
			return out
		}
	}
	return append(out, code)
}

func upsertCapability(caps []types.ContractCapability, cap types.ContractCapability) []types.ContractCapability {
	out := append([]types.ContractCapability(nil), caps...)
	for i := range out {
		if out[i].ContractAddressUser == cap.ContractAddressUser && out[i].PoolID == cap.PoolID && out[i].Capability == cap.Capability {
			out[i] = cap
			return out
		}
	}
	return append(out, cap)
}

func upsertAsset(assets []types.AssetOwnershipRecord, record types.AssetOwnershipRecord) []types.AssetOwnershipRecord {
	out := append([]types.AssetOwnershipRecord(nil), assets...)
	for i := range out {
		if out[i].AssetType == record.AssetType && out[i].ContractAddressUser == record.ContractAddressUser && out[i].AssetID == record.AssetID {
			out[i] = record
			return out
		}
	}
	return append(out, record)
}

func findCode(codes []types.CodeRecord, codeID string) (types.CodeRecord, bool) {
	for _, code := range codes {
		if code.CodeID == codeID {
			return code, true
		}
	}
	return types.CodeRecord{}, false
}

func findContract(contracts []types.Contract, address string) (types.Contract, bool) {
	_, contract, found := findContractWithIndex(contracts, address)
	return contract, found
}

func findContractWithIndex(contracts []types.Contract, address string) (int, types.Contract, bool) {
	for idx, contract := range contracts {
		if contract.AddressUser == address {
			return idx, contract, true
		}
	}
	return -1, types.Contract{}, false
}

func hasCapability(caps []types.ContractCapability, contract string, poolID string) bool {
	for _, cap := range caps {
		if cap.ContractAddressUser == contract && cap.PoolID == poolID && cap.Capability == types.NativeStakingCapability {
			return true
		}
	}
	return false
}

func hasAnyNativeStakingCapability(caps []types.ContractCapability, contract string) bool {
	for _, cap := range caps {
		if cap.ContractAddressUser == contract && cap.Capability == types.NativeStakingCapability {
			return true
		}
	}
	return false
}

func (k Keeper) writeGenesis(ctx context.Context) error {
	if k.storeService == nil {
		return nil
	}
	gs := types.RefreshStateRoot(k.genesis)
	if err := gs.Validate(); err != nil {
		return err
	}
	bz, err := json.Marshal(gs)
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(genesisKey, bz)
}
