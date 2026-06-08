package keeper

import (
	"errors"
	"fmt"
	"math"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
	"github.com/sovereign-l1/l1/x/contracts/types"
)

type Keeper struct {
	genesis             types.GenesisState
	accountStatusReader AccountStatusReader
}

const (
	accountStatusActive   = "active"
	accountStatusInactive = "inactive"
	accountStatusFrozen   = "frozen"
)

// AccountStatusReader is a temporary integration boundary for CHAT 1 native-account wiring.
// It keeps contract auth/freeze checks local until the account keeper interface is finalized.
type AccountStatusReader interface {
	AccountStatus(address string) (string, bool)
}

func NewKeeper() Keeper {
	return Keeper{genesis: types.DefaultGenesis()}
}

func NewKeeperWithAccountStatus(reader AccountStatusReader) Keeper {
	k := NewKeeper()
	k.accountStatusReader = reader
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

func (k Keeper) ExportGenesis() types.GenesisState {
	return types.RefreshStateRoot(k.genesis)
}

func (k Keeper) Params() types.Params {
	return k.genesis.Params
}

func (k Keeper) ValidateInvariants() error {
	return k.genesis.Validate()
}

func (k Keeper) RootContribution() (coretypes.RootContribution, error) {
	return types.RootContribution(k.genesis)
}

func (k *Keeper) StoreCode(msg types.MsgStoreCode) (types.StoreCodeResponse, error) {
	if !k.genesis.Params.Enabled {
		return types.StoreCodeResponse{}, errors.New(types.ErrExecutionFailed + ": module disabled")
	}
	if err := types.ValidateUserFacingAEAddress("contract code authority", msg.Authority); err != nil {
		return types.StoreCodeResponse{}, err
	}
	if err := k.ensureActiveWallet(msg.Authority, "contract code store"); err != nil {
		return types.StoreCodeResponse{}, err
	}
	if msg.CodeBytes == 0 || msg.CodeBytes > k.genesis.Params.MaxCodeBytes {
		return types.StoreCodeResponse{}, errors.New(types.ErrInvalidBytecode + ": code size out of bounds")
	}
	if err := coretypes.ValidateHash("contracts code hash", msg.CodeHash); err != nil {
		return types.StoreCodeResponse{}, err
	}
	next := k.genesis
	next.State.Codes = upsertCode(next.State.Codes, types.CodeRecord{
		CodeID:    msg.CodeHash,
		CodeHash:  msg.CodeHash,
		CodeBytes: msg.CodeBytes,
		Owner:     msg.Authority,
	})
	next = types.RefreshStateRoot(next)
	if err := next.Validate(); err != nil {
		return types.StoreCodeResponse{}, err
	}
	k.genesis = next
	return types.StoreCodeResponse{CodeID: msg.CodeHash, StateRoot: k.genesis.StateRoot}, nil
}

func (k Keeper) Contract(req types.QueryContractRequest) (types.QueryContractResponse, error) {
	if err := types.ValidateContractAddress(req.ContractAddress); err != nil {
		return types.QueryContractResponse{}, err
	}
	contract, found := findContract(k.genesis.State.Contracts, req.ContractAddress)
	return types.QueryContractResponse{ContractAddress: req.ContractAddress, StateRoot: k.genesis.StateRoot, Found: found, Contract: contract}, nil
}

func (k *Keeper) InstantiateContract(msg types.MsgInstantiateContract) (types.InstantiateContractResponse, error) {
	if !k.genesis.Params.Enabled {
		return types.InstantiateContractResponse{}, errors.New(types.ErrExecutionFailed + ": module disabled")
	}
	if err := types.ValidateUserFacingAEAddress("contract creator", msg.Creator); err != nil {
		return types.InstantiateContractResponse{}, err
	}
	if err := k.ensureActiveWallet(msg.Creator, "contract instantiate"); err != nil {
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
	admin := msg.Admin
	if admin == "" {
		admin = msg.Creator
	}
	if err := types.ValidateUserFacingAEAddress("contract admin", admin); err != nil {
		return types.InstantiateContractResponse{}, err
	}
	user, raw, err := types.DeriveContractAddress(msg.Creator, msg.CodeID, msg.Salt)
	if err != nil {
		return types.InstantiateContractResponse{}, err
	}
	if _, found := findContract(k.genesis.State.Contracts, user); found {
		return types.InstantiateContractResponse{}, errors.New(types.ErrContractNotFound + ": contract address already exists")
	}
	data := append([]byte(nil), msg.InitMsg...)
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
	contract := types.Contract{
		AddressUser:             user,
		AddressRaw:              raw,
		CodeID:                  msg.CodeID,
		Creator:                 msg.Creator,
		Owner:                   msg.Creator,
		Admin:                   admin,
		InitMsg:                 append([]byte(nil), data...),
		Data:                    append([]byte(nil), data...),
		Balance:                 msg.Funds,
		Status:                  types.ContractStatusActive,
		StorageBytes:            storageBytes,
		LastStorageChargeHeight: msg.Height,
		CreatedHeight:           msg.Height,
		UpdatedHeight:           msg.Height,
	}
	next := k.genesis
	next.State.Contracts = append(next.State.Contracts, contract)
	next = types.RefreshStateRoot(next)
	if err := next.Validate(); err != nil {
		return types.InstantiateContractResponse{}, err
	}
	k.genesis = next
	return types.InstantiateContractResponse{
		ContractAddressUser: user,
		ContractAddressRaw:  raw,
		Owner:               contract.Owner,
		Admin:               contract.Admin,
		Balance:             contract.Balance,
		Events: []types.ContractEvent{{
			Type:        types.EventTypeContractInstantiated,
			Actor:       msg.Creator,
			Contract:    user,
			Amount:      msg.Funds,
			InternalRaw: raw,
		}},
	}, nil
}

func (k *Keeper) ExecuteContract(msg types.MsgExecuteContract) (types.ExecuteContractResponse, error) {
	if !k.genesis.Params.Enabled {
		return types.ExecuteContractResponse{}, errors.New(types.ErrExecutionFailed + ": module disabled")
	}
	if err := types.ValidateUserFacingAEAddress("contract execute sender", msg.Sender); err != nil {
		return types.ExecuteContractResponse{}, err
	}
	if err := k.ensureActiveWallet(msg.Sender, "contract execute"); err != nil {
		return types.ExecuteContractResponse{}, err
	}
	if msg.Height == 0 {
		return types.ExecuteContractResponse{}, errors.New("contract execute height must be positive")
	}
	idx, contract, found := findContractWithIndex(k.genesis.State.Contracts, msg.ContractAddress)
	if !found {
		return types.ExecuteContractResponse{}, errors.New(types.ErrContractNotFound + ": contract not found")
	}
	if contract.Status == types.ContractStatusFrozen {
		return types.ExecuteContractResponse{}, errors.New(types.ErrAccountFrozen + ": frozen contract cannot execute normal calls")
	}
	contract, err := k.chargeContractRentAt(idx, contract, msg.Height)
	if err != nil {
		return types.ExecuteContractResponse{}, errors.New(types.ErrStorageRent + ": contract has storage rent debt")
	}
	balance, err := checkedAdd(contract.Balance, msg.Funds, "contract balance overflow")
	if err != nil {
		return types.ExecuteContractResponse{}, err
	}
	contract.Balance = balance
	contract.Data = append([]byte(nil), msg.Msg...)
	storageBytes, err := k.contractStorageBytes(contract)
	if err != nil {
		return types.ExecuteContractResponse{}, err
	}
	if storageBytes > k.genesis.Params.MaxContractStorageBytes {
		return types.ExecuteContractResponse{}, errors.New(types.ErrStorageRent + ": contract storage exceeds configured limit")
	}
	contract.StorageBytes = storageBytes
	contract.UpdatedHeight = msg.Height
	next := k.genesis
	next.State.Contracts[idx] = contract
	next = types.RefreshStateRoot(next)
	if err := next.Validate(); err != nil {
		return types.ExecuteContractResponse{}, err
	}
	k.genesis = next
	return types.ExecuteContractResponse{
		ContractAddressUser: contract.AddressUser,
		Owner:               contract.Owner,
		Balance:             contract.Balance,
		Events: []types.ContractEvent{{
			Type:        types.EventTypeContractExecuted,
			Actor:       msg.Sender,
			Contract:    contract.AddressUser,
			Amount:      msg.Funds,
			InternalRaw: contract.AddressRaw,
		}},
	}, nil
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

func (k *Keeper) UnfreezeContract(msg types.MsgUnfreezeContract) (types.Contract, error) {
	if err := types.ValidateUserFacingAEAddress("contract unfreeze sender", msg.Sender); err != nil {
		return types.Contract{}, err
	}
	if err := k.ensureActiveWallet(msg.Sender, "contract unfreeze"); err != nil {
		return types.Contract{}, err
	}
	if msg.Height == 0 {
		return types.Contract{}, errors.New("contract unfreeze height must be positive")
	}
	idx, contract, found := findContractWithIndex(k.genesis.State.Contracts, msg.ContractAddress)
	if !found {
		return types.Contract{}, errors.New(types.ErrContractNotFound + ": contract not found")
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
		ContractAddressUser: msg.ContractAddressUser,
		ContractAddressRaw:  msg.ContractAddressRaw,
		Capability:          types.NativeStakingCapability,
		PoolID:              msg.PoolID,
		GrantedHeight:       msg.Height,
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
	if contract.Status != types.ContractStatusActive {
		return types.NativeStakingInjectionRecord{}, errors.New(types.ErrAccountFrozen + ": frozen contract cannot inject native staking")
	}
	contract, err := k.chargeContractRentAt(idx, contract, msg.Height)
	if err != nil {
		return types.NativeStakingInjectionRecord{}, errors.New(types.ErrStorageRent + ": contract has storage rent debt")
	}
	if !hasCapability(k.genesis.State.StakingCapabilities, msg.CallerContractUser, msg.PoolID) {
		return types.NativeStakingInjectionRecord{}, errors.New(types.ErrUnauthorized + ": contract lacks native staking capability")
	}
	record := types.NativeStakingInjectionRecord{
		ContractAddressUser: msg.CallerContractUser,
		ContractAddressRaw:  msg.CallerContractRaw,
		PoolID:              msg.PoolID,
		Amount:              msg.Amount,
		Height:              msg.Height,
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
		SourceContractUser: msg.SourceContractUser,
		DestinationAccount: msg.DestinationAccount,
		Funds:              msg.Funds,
		Body:               append([]byte(nil), msg.Body...),
		Height:             msg.Height,
	}
	if err := record.Validate(); err != nil {
		return types.InternalMessage{}, err
	}
	idx, contract, found := findContractWithIndex(k.genesis.State.Contracts, msg.SourceContractUser)
	if !found {
		return types.InternalMessage{}, errors.New(types.ErrContractNotFound + ": source contract not found")
	}
	if contract.Status != types.ContractStatusActive {
		return types.InternalMessage{}, errors.New(types.ErrAccountFrozen + ": frozen contract cannot send internal messages")
	}
	if _, err := k.chargeContractRentAt(idx, contract, msg.Height); err != nil {
		return types.InternalMessage{}, errors.New(types.ErrStorageRent + ": contract has storage rent debt")
	}
	next := k.genesis
	next.State.InternalMessages = append(next.State.InternalMessages, record)
	next = types.RefreshStateRoot(next)
	if err := next.Validate(); err != nil {
		return types.InternalMessage{}, err
	}
	k.genesis = next
	return record, nil
}

func (k Keeper) AssetOwner(req types.QueryAssetOwnerRequest) (types.QueryAssetOwnerResponse, error) {
	if req.AssetType != types.AssetTypeToken && req.AssetType != types.AssetTypeNFT && req.AssetType != types.AssetTypeDEX {
		return types.QueryAssetOwnerResponse{}, fmt.Errorf("unsupported contract asset type %q", req.AssetType)
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

func (k *Keeper) ensureActiveWallet(address string, operation string) error {
	if k.accountStatusReader == nil {
		return nil
	}
	status, found := k.accountStatusReader.AccountStatus(address)
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

func (k *Keeper) chargeContractRentAt(idx int, contract types.Contract, height uint64) (types.Contract, error) {
	contract, changed, err := k.chargeRent(contract, height)
	if err != nil {
		return types.Contract{}, err
	}
	if contract.StorageRentDebt > 0 {
		contract.Status = types.ContractStatusFrozen
		if err := k.persistContractAt(idx, contract); err != nil {
			return types.Contract{}, err
		}
		return contract, errors.New(types.ErrStorageRent + ": contract has storage rent debt")
	}
	if changed {
		if err := k.persistContractAt(idx, contract); err != nil {
			return types.Contract{}, err
		}
	}
	return contract, nil
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

func (k Keeper) chargeRent(contract types.Contract, height uint64) (types.Contract, bool, error) {
	if height < contract.LastStorageChargeHeight {
		return types.Contract{}, false, errors.New(types.ErrStorageRent + ": contract storage rent height must be monotonic")
	}
	if height <= contract.LastStorageChargeHeight || contract.StorageBytes == 0 || k.genesis.Params.StorageRentPerByteBlock == 0 {
		return contract, false, nil
	}
	blocks := height - contract.LastStorageChargeHeight
	charge, err := checkedMul(blocks, contract.StorageBytes, "contract storage rent overflow")
	if err != nil {
		return types.Contract{}, false, err
	}
	charge, err = checkedMul(charge, k.genesis.Params.StorageRentPerByteBlock, "contract storage rent overflow")
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
