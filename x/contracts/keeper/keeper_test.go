package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/app/addressing"
	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
	"github.com/sovereign-l1/l1/x/contracts/types"
)

func TestContractsKeeperGenesisExportImportInvariantsAndRootContribution(t *testing.T) {
	keeper := NewKeeper()
	require.NoError(t, keeper.ValidateInvariants())

	exported := keeper.ExportGenesis()
	require.NoError(t, exported.Validate())

	imported := NewKeeper()
	require.NoError(t, imported.InitGenesis(exported))
	require.Equal(t, exported, imported.ExportGenesis())

	root, err := imported.RootContribution()
	require.NoError(t, err)
	require.Equal(t, coretypes.RootType(types.ModuleName), root.RootType)
	require.Equal(t, types.ModuleName, root.ID)
	require.Equal(t, exported.StateRoot, root.RootHash)
	require.NoError(t, root.Validate())
}

func TestContractsKeeperTypedErrorsAndMsgQuerySurface(t *testing.T) {
	keeper := NewKeeper()
	authority := aeAddress("11")
	codeHash := coretypes.DeterministicEmptyRootCommitment(coretypes.RootType(types.ModuleName), "code")

	response, err := keeper.StoreCode(types.MsgStoreCode{Authority: authority, CodeHash: codeHash, CodeBytes: 128})
	require.NoError(t, err)
	require.Equal(t, codeHash, response.CodeID)
	require.NotEmpty(t, response.StateRoot)

	_, err = keeper.StoreCode(types.MsgStoreCode{Authority: authority, CodeHash: codeHash, CodeBytes: 0})
	require.ErrorContains(t, err, types.ErrInvalidBytecode)

	stateInit := types.NewStateInit(authority, codeHash, nil, "query", 0)
	contractAddress, _, err := types.DeriveContractAddressFromStateInit("", "", authority, stateInit, types.DefaultParams())
	require.NoError(t, err)
	query, err := keeper.Contract(types.QueryContractRequest{ContractAddress: contractAddress})
	require.NoError(t, err)
	require.False(t, query.Found)
	require.Equal(t, contractAddress, query.ContractAddress)

	_, err = keeper.Contract(types.QueryContractRequest{})
	require.ErrorContains(t, err, types.ErrContractNotFound)
}

func TestAVMExitCodesAreSmallStableAndNamed(t *testing.T) {
	require.Equal(t, uint32(0), types.ExitCodeOK)
	require.Equal(t, "ok", types.ExitCodeName(types.ExitCodeOK))
	require.Equal(t, "code_rejected", types.ExitCodeName(types.ExitCodeCodeRejected))
	require.Equal(t, "internal_bounce", types.ExitCodeName(types.ExitCodeInternalBounce))
	require.Less(t, types.ExitCodeInternalBounce, uint32(100))
	require.Equal(t, "unknown", types.ExitCodeName(105))
}

func TestStoreCodeAcceptsCanonicalAVMBytecodeAndRejectsNondeterminism(t *testing.T) {
	wallet := aeAddress("11")
	k := NewKeeperWithAccountStatus(testAccountStatus{wallet: accountStatusActive})
	bytecode := []byte("AVM1\nset key value\nemit ok")
	codeHash := types.CanonicalCodeHash(bytecode)

	response, err := k.StoreCode(types.MsgStoreCode{Authority: wallet, Bytecode: bytecode})
	require.NoError(t, err)
	require.Equal(t, codeHash, response.CodeID)
	exported := k.ExportGenesis()
	require.Len(t, exported.State.Codes, 1)
	require.Equal(t, bytecode, exported.State.Codes[0].Bytecode)
	require.Equal(t, uint64(len(bytecode)), exported.State.Codes[0].CodeBytes)
	require.Equal(t, codeHash, exported.State.Codes[0].CodeHash)

	_, err = k.StoreCode(types.MsgStoreCode{Authority: wallet, Bytecode: []byte("AVM1 time.now")})
	require.ErrorContains(t, err, types.ErrInvalidBytecode)
	_, err = k.StoreCode(types.MsgStoreCode{Authority: wallet, CodeHash: sha256Hex("wrong"), Bytecode: bytecode})
	require.ErrorContains(t, err, "canonical bytecode hash")
}

func TestWalletInstantiatesExecutesAndPassesFunds(t *testing.T) {
	wallet := aeAddress("11")
	k := NewKeeperWithAccountStatus(testAccountStatus{wallet: accountStatusActive})
	codeHash := storeContractCode(t, &k, wallet)
	initMsg := []byte(`{"owner":"wallet"}`)
	initialFunds := uint64(500)

	created, err := k.InstantiateContract(types.MsgInstantiateContract{
		Creator:	wallet,
		CodeID:		codeHash,
		InitMsg:	initMsg,
		Funds:		initialFunds,
		Admin:		wallet,
		Salt:		"contract-a",
		Height:		10,
	})
	require.NoError(t, err)
	require.Equal(t, wallet, created.Owner)
	require.Equal(t, wallet, created.Admin)
	require.Equal(t, initialFunds, created.Balance)
	require.True(t, stringsHasPrefix(created.ContractAddressUser, "AE"))
	require.True(t, stringsHasPrefix(created.ContractAddressRaw, "4:"))
	require.Equal(t, types.EventTypeContractInstantiated, created.Events[0].Type)
	require.Equal(t, created.ContractAddressUser, created.Events[0].Contract)
	require.Equal(t, created.ContractAddressRaw, created.Events[0].InternalRaw)
	query, err := k.Contract(types.QueryContractRequest{ContractAddress: created.ContractAddressUser})
	require.NoError(t, err)
	require.Equal(t, contractStorageBytes(128, initMsg), query.Contract.StorageBytes)
	require.Equal(t, codeHash, query.Contract.CodeHash)
	require.NotEmpty(t, query.Contract.StateRoot)
	require.Equal(t, uint64(1), query.Contract.LogicalTime)

	execMsg := []byte(`{"transfer":1}`)
	executed, err := k.ExecuteContract(types.MsgExecuteContract{
		Sender:			wallet,
		ContractAddress:	created.ContractAddressUser,
		Msg:			execMsg,
		Funds:			25,
		Height:			11,
	})
	require.NoError(t, err)
	require.Equal(t, created.ContractAddressUser, executed.ContractAddressUser)
	require.Equal(t, initialFunds-contractStorageBytes(128, initMsg)+25, executed.Balance)
	require.Equal(t, types.EventTypeContractExecuted, executed.Events[0].Type)
	require.Equal(t, created.ContractAddressUser, executed.Events[0].Contract)
	require.Equal(t, created.ContractAddressRaw, executed.Events[0].InternalRaw)
	query, err = k.Contract(types.QueryContractRequest{ContractAddress: created.ContractAddressUser})
	require.NoError(t, err)
	require.Equal(t, contractStorageBytes(128, execMsg), query.Contract.StorageBytes)
	require.Equal(t, uint64(2), query.Contract.LogicalTime)
	require.Equal(t, types.ComputeContractStateRoot(query.Contract), query.Contract.StateRoot)
}

func TestContractUpgradeMigrationAndAdminPolicy(t *testing.T) {
	wallet := aeAddress("11")
	admin := aeAddress("22")
	other := aeAddress("33")
	k := NewKeeperWithAccountStatus(testAccountStatus{wallet: accountStatusActive, admin: accountStatusActive, other: accountStatusActive})
	codeV1 := storeContractCode(t, &k, wallet)
	codeV2Hash := sha256Hex("code-v2")
	_, err := k.StoreCode(types.MsgStoreCode{Authority: wallet, CodeHash: codeV2Hash, CodeBytes: 256})
	require.NoError(t, err)

	immutable, err := k.InstantiateContract(types.MsgInstantiateContract{
		Creator:	wallet, CodeID: codeV1, InitMsg: []byte("v1"), Admin: admin, Salt: "immutable", Height: 10,
	})
	require.NoError(t, err)
	_, err = k.UpgradeContractCode(types.MsgUpgradeContractCode{
		Actor:	admin, ContractAddress: immutable.ContractAddressUser, NewCodeID: codeV2Hash, MigrationHandler: "schema_only", Height: 11,
	})
	require.ErrorContains(t, err, "immutable")

	upgradeable, err := k.InstantiateContract(types.MsgInstantiateContract{
		Creator:	wallet, CodeID: codeV1, InitMsg: []byte("v1"), Admin: admin, Salt: "upgradeable", Upgradeable: true, SchemaVersion: 1, Height: 20,
	})
	require.NoError(t, err)
	_, err = k.UpgradeContractCode(types.MsgUpgradeContractCode{
		Actor:	other, ContractAddress: upgradeable.ContractAddressUser, NewCodeID: codeV2Hash, MigrationHandler: "schema_only", Height: 21,
	})
	require.ErrorContains(t, err, types.ErrUnauthorized)
	receipt, err := k.UpgradeContractCode(types.MsgUpgradeContractCode{
		Actor:	admin, ContractAddress: upgradeable.ContractAddressUser, NewCodeID: codeV2Hash, MigrationHandler: "schema_only", Height: 22,
	})
	require.NoError(t, err)
	require.Equal(t, "upgrade_code", receipt.Operation)
	query, err := k.Contract(types.QueryContractRequest{ContractAddress: upgradeable.ContractAddressUser})
	require.NoError(t, err)
	require.Equal(t, codeV2Hash, query.Contract.CodeID)
	require.Equal(t, uint64(1), query.Contract.StorageSchemaVersion)

	beforeRoot := query.Contract.StateRoot
	_, err = k.MigrateContractState(types.MsgMigrateContractState{
		Actor:	admin, ContractAddress: upgradeable.ContractAddressUser, FromSchemaVersion: 1, ToSchemaVersion: 2, MigrationHandler: "fail", Payload: []byte("bad"), Height: 23,
	})
	require.ErrorContains(t, err, "migration handler failed")
	rolledBack, err := k.Contract(types.QueryContractRequest{ContractAddress: upgradeable.ContractAddressUser})
	require.NoError(t, err)
	require.Equal(t, uint64(1), rolledBack.Contract.StorageSchemaVersion)
	require.Equal(t, beforeRoot, rolledBack.Contract.StateRoot)

	receipt, err = k.MigrateContractState(types.MsgMigrateContractState{
		Actor:	admin, ContractAddress: upgradeable.ContractAddressUser, FromSchemaVersion: 1, ToSchemaVersion: 2, MigrationHandler: "append", Payload: []byte(":v2"), Height: 24,
	})
	require.NoError(t, err)
	require.Equal(t, "migrate_state", receipt.Operation)
	migrated, err := k.Contract(types.QueryContractRequest{ContractAddress: upgradeable.ContractAddressUser})
	require.NoError(t, err)
	require.Equal(t, uint64(2), migrated.Contract.StorageSchemaVersion)
	require.Equal(t, []byte("v1:v2"), migrated.Contract.Data)

	newAdmin := aeAddress("44")
	_, err = k.SetContractAdmin(types.MsgSetContractAdmin{Actor: admin, ContractAddress: upgradeable.ContractAddressUser, NewAdmin: newAdmin, Height: 25})
	require.NoError(t, err)
	_, err = k.DisableContractUpgrades(types.MsgDisableContractUpgrades{Actor: newAdmin, ContractAddress: upgradeable.ContractAddressUser, Height: 26})
	require.NoError(t, err)
	_, err = k.UpgradeContractCode(types.MsgUpgradeContractCode{
		Actor:	newAdmin, ContractAddress: upgradeable.ContractAddressUser, NewCodeID: codeV1, MigrationHandler: "schema_only", Height: 27,
	})
	require.ErrorContains(t, err, "immutable")
}

func TestSystemOwnedContractUpgradeRequiresGovernanceAuthority(t *testing.T) {
	wallet := aeAddress("11")
	k := NewKeeperWithAccountStatus(testAccountStatus{wallet: accountStatusActive})
	codeV1 := storeContractCode(t, &k, wallet)
	codeV2 := sha256Hex("system-code-v2")
	_, err := k.StoreCode(types.MsgStoreCode{Authority: wallet, CodeHash: codeV2, CodeBytes: 256})
	require.NoError(t, err)
	created, err := k.InstantiateContract(types.MsgInstantiateContract{
		Creator:	wallet, CodeID: codeV1, InitMsg: []byte("sys"), Admin: wallet, Salt: "system-owned", Upgradeable: true, SystemOwned: true, Height: 10,
	})
	require.NoError(t, err)
	_, err = k.UpgradeContractCode(types.MsgUpgradeContractCode{
		Actor:	wallet, ContractAddress: created.ContractAddressUser, NewCodeID: codeV2, MigrationHandler: "schema_only", Height: 11,
	})
	require.ErrorContains(t, err, "governance authority")
	_, err = k.UpgradeContractCode(types.MsgUpgradeContractCode{
		Actor:	k.Params().Authority, ContractAddress: created.ContractAddressUser, NewCodeID: codeV2, MigrationHandler: "schema_only", Height: 12,
	})
	require.NoError(t, err)
}

func TestFrozenWalletCannotInstantiateOrExecuteUntilUnfrozen(t *testing.T) {
	wallet := aeAddress("11")
	status := testAccountStatus{wallet: accountStatusActive}
	k := NewKeeperWithAccountStatus(status)
	codeHash := storeContractCode(t, &k, wallet)
	created := instantiateContract(t, &k, wallet, codeHash, "contract-a", 10, 300, 2)

	status[wallet] = accountStatusFrozen
	k.accountStatusReader = status
	_, err := k.InstantiateContract(types.MsgInstantiateContract{Creator: wallet, CodeID: codeHash, Salt: "contract-b", Height: 11})
	require.ErrorContains(t, err, types.ErrAccountFrozen)

	_, err = k.ExecuteContract(types.MsgExecuteContract{Sender: wallet, ContractAddress: created.ContractAddressUser, Msg: []byte("blocked"), Height: 11})
	require.ErrorContains(t, err, types.ErrAccountFrozen)

	status[wallet] = accountStatusActive
	k.accountStatusReader = status
	_, err = k.ExecuteContract(types.MsgExecuteContract{Sender: wallet, ContractAddress: created.ContractAddressUser, Msg: []byte("ok"), Height: 11})
	require.NoError(t, err)
}

func TestFrozenContractRecoveryKeepsCodeDataAndBalance(t *testing.T) {
	wallet := aeAddress("11")
	k := NewKeeperWithAccountStatus(testAccountStatus{wallet: accountStatusActive})
	codeHash := storeContractCode(t, &k, wallet)
	created := instantiateContract(t, &k, wallet, codeHash, "rent", 10, 1, 100)

	_, err := k.ExecuteContract(types.MsgExecuteContract{Sender: wallet, ContractAddress: created.ContractAddressUser, Msg: []byte("too late"), Height: 200})
	require.ErrorContains(t, err, types.ErrStorageRent)

	frozen, err := k.Contract(types.QueryContractRequest{ContractAddress: created.ContractAddressUser})
	require.NoError(t, err)
	require.True(t, frozen.Found)
	require.Equal(t, types.ContractStatusFrozen, frozen.Contract.Status)
	require.Equal(t, codeHash, frozen.Contract.CodeID)
	require.Equal(t, []byte("init"), frozen.Contract.Data)
	require.Zero(t, frozen.Contract.Balance)
	require.NotZero(t, frozen.Contract.StorageRentDebt)

	topped, err := k.TopUpContract(types.MsgTopUpContract{Sender: wallet, ContractAddress: created.ContractAddressUser, Amount: frozen.Contract.StorageRentDebt + 50, Height: 201})
	require.NoError(t, err)
	require.Equal(t, codeHash, topped.CodeID)
	require.Equal(t, []byte("init"), topped.Data)

	paid, err := k.PayContractStorageDebt(types.MsgPayContractStorageDebt{Sender: wallet, ContractAddress: created.ContractAddressUser, Amount: frozen.Contract.StorageRentDebt, Height: 202})
	require.NoError(t, err)
	require.Zero(t, paid.StorageRentDebt)
	require.Equal(t, []byte("init"), paid.Data)

	_, err = k.ExecuteContract(types.MsgExecuteContract{Sender: wallet, ContractAddress: created.ContractAddressUser, Msg: []byte("still-frozen"), Height: 202})
	require.ErrorContains(t, err, types.ErrAccountFrozen)

	unfrozen, err := k.UnfreezeContract(types.MsgUnfreezeContract{Sender: wallet, ContractAddress: created.ContractAddressUser, Height: 203})
	require.NoError(t, err)
	require.Equal(t, types.ContractStatusActive, unfrozen.Status)
	require.Equal(t, codeHash, unfrozen.CodeID)
	require.Equal(t, []byte("init"), unfrozen.Data)
	require.NotZero(t, unfrozen.Balance)
}

func TestContractOwnersAdminsAndAssetQueriesUseAEAndRegistryState(t *testing.T) {
	wallet := aeAddress("11")
	admin := aeAddress("12")
	assetOwner := aeAddress("13")
	k := NewKeeperWithAccountStatus(testAccountStatus{wallet: accountStatusActive})
	codeHash := storeContractCode(t, &k, wallet)
	rawAdmin, err := types.RawAddressForUserAddress(admin)
	require.NoError(t, err)
	_, err = k.InstantiateContract(types.MsgInstantiateContract{
		Creator:	wallet, CodeID: codeHash, Admin: rawAdmin, Salt: "raw-admin", Height: 9,
	})
	require.ErrorContains(t, err, "AE user-facing")

	created, err := k.InstantiateContract(types.MsgInstantiateContract{
		Creator:	wallet, CodeID: codeHash, Admin: admin, Salt: "asset-contract", Height: 10,
	})
	require.NoError(t, err)
	require.True(t, stringsHasPrefix(created.Owner, "AE"))
	require.True(t, stringsHasPrefix(created.Admin, "AE"))

	err = k.SetAssetOwner(types.AssetOwnershipRecord{
		AssetType:		"contract_asset",
		ContractAddressUser:	created.ContractAddressUser,
		AssetID:		"asset-1",
		Owner:			assetOwner,
	})
	require.NoError(t, err)
	owner, err := k.AssetOwner(types.QueryAssetOwnerRequest{AssetType: "contract_asset", ContractAddressUser: created.ContractAddressUser, AssetID: "asset-1"})
	require.NoError(t, err)
	require.True(t, owner.Found)
	require.Equal(t, assetOwner, owner.Owner)

	exported := k.ExportGenesis()
	require.NotEqual(t, "token-balance", exported.State.Contracts[0].Owner)
	require.Len(t, exported.State.AssetOwnership, 1)
}

func TestAssetOwnershipRecordValidate(t *testing.T) {
	wallet := aeAddress("11")
	k := NewKeeperWithAccountStatus(testAccountStatus{wallet: accountStatusActive})
	codeHash := storeContractCode(t, &k, wallet)
	created := instantiateContract(t, &k, wallet, codeHash, "asset-record", 10, 0, 0)

	err := k.SetAssetOwner(types.AssetOwnershipRecord{
		AssetType:		"contract_asset",
		ContractAddressUser:	created.ContractAddressUser,
		AssetID:		"asset-1",
		Owner:			wallet,
	})
	require.NoError(t, err)
}

func TestOfficialLiquidStakingContractCapabilityAllowsNativeHookOnlyForAuthorizedContract(t *testing.T) {
	wallet := aeAddress("11")
	k := NewKeeperWithAccountStatus(testAccountStatus{wallet: accountStatusActive})
	codeHash := storeContractCode(t, &k, wallet)
	official := instantiateContract(t, &k, wallet, codeHash, "official-lst", 10, 1000, 0)
	unauthorized := instantiateContract(t, &k, wallet, codeHash, "other-contract", 11, 1000, 0)

	capability, err := k.GrantNativeStakingCapability(types.MsgGrantNativeStakingCapability{
		Authority:		types.DefaultParams().Authority,
		ContractAddressUser:	official.ContractAddressUser,
		ContractAddressRaw:	official.ContractAddressRaw,
		PoolID:			"official-pool",
		Height:			12,
	})
	require.NoError(t, err)
	require.Equal(t, official.ContractAddressUser, capability.ContractAddressUser)
	require.Equal(t, official.ContractAddressRaw, capability.ContractAddressRaw)

	injection, err := k.InjectNativeStaking(types.MsgInjectNativeStaking{
		CallerContractUser:	official.ContractAddressUser,
		CallerContractRaw:	official.ContractAddressRaw,
		PoolID:			"official-pool",
		Amount:			500,
		Height:			13,
	})
	require.NoError(t, err)
	require.Equal(t, official.ContractAddressUser, injection.ContractAddressUser)
	require.Equal(t, official.ContractAddressRaw, injection.ContractAddressRaw)

	_, err = k.InjectNativeStaking(types.MsgInjectNativeStaking{
		CallerContractUser:	unauthorized.ContractAddressUser,
		CallerContractRaw:	unauthorized.ContractAddressRaw,
		PoolID:			"official-pool",
		Amount:			1,
		Height:			14,
	})
	require.ErrorContains(t, err, types.ErrUnauthorized)
}

func TestNativeStakingCapabilityRejectsBadAuthorityAndFrozenContract(t *testing.T) {
	wallet := aeAddress("11")
	k := NewKeeperWithAccountStatus(testAccountStatus{wallet: accountStatusActive})
	codeHash := storeContractCode(t, &k, wallet)
	official := instantiateContract(t, &k, wallet, codeHash, "official-frozen", 10, 100, 0)

	_, err := k.GrantNativeStakingCapability(types.MsgGrantNativeStakingCapability{
		Authority:		wallet,
		ContractAddressUser:	official.ContractAddressUser,
		ContractAddressRaw:	official.ContractAddressRaw,
		PoolID:			"official-pool",
		Height:			11,
	})
	require.ErrorContains(t, err, types.ErrUnauthorized)

	_, err = k.GrantNativeStakingCapability(types.MsgGrantNativeStakingCapability{
		Authority:		types.DefaultParams().Authority,
		ContractAddressUser:	official.ContractAddressUser,
		ContractAddressRaw:	official.ContractAddressRaw,
		PoolID:			"official-pool",
		Height:			12,
	})
	require.NoError(t, err)

	gs := k.ExportGenesis()
	gs.State.Contracts[0].Status = types.ContractStatusFrozen
	require.NoError(t, k.InitGenesis(gs))

	_, err = k.InjectNativeStaking(types.MsgInjectNativeStaking{
		CallerContractUser:	official.ContractAddressUser,
		CallerContractRaw:	official.ContractAddressRaw,
		PoolID:			"official-pool",
		Amount:			500,
		Height:			13,
	})
	require.ErrorContains(t, err, types.ErrAccountFrozen)

	query, err := k.Contract(types.QueryContractRequest{ContractAddress: official.ContractAddressUser})
	require.NoError(t, err)
	require.True(t, query.Found)
	require.Equal(t, types.ContractStatusFrozen, query.Contract.Status)
}

func TestNativeStakingHookChargesStorageRentBeforeInjection(t *testing.T) {
	wallet := aeAddress("11")
	k := NewKeeperWithAccountStatus(testAccountStatus{wallet: accountStatusActive})
	codeHash := storeContractCode(t, &k, wallet)
	official := instantiateContract(t, &k, wallet, codeHash, "official-rent", 10, 1, 0)

	_, err := k.GrantNativeStakingCapability(types.MsgGrantNativeStakingCapability{
		Authority:		types.DefaultParams().Authority,
		ContractAddressUser:	official.ContractAddressUser,
		ContractAddressRaw:	official.ContractAddressRaw,
		PoolID:			"official-pool",
		Height:			11,
	})
	require.NoError(t, err)

	_, err = k.InjectNativeStaking(types.MsgInjectNativeStaking{
		CallerContractUser:	official.ContractAddressUser,
		CallerContractRaw:	official.ContractAddressRaw,
		PoolID:			"official-pool",
		Amount:			500,
		Height:			12,
	})
	require.ErrorContains(t, err, types.ErrStorageRent)

	query, err := k.Contract(types.QueryContractRequest{ContractAddress: official.ContractAddressUser})
	require.NoError(t, err)
	require.Equal(t, types.ContractStatusFrozenLimited, query.Contract.Status)
	require.Zero(t, query.Contract.Balance)
	require.Equal(t, []byte("init"), query.Contract.Data)
	require.NotZero(t, query.Contract.StorageRentDebt)
	require.Empty(t, k.ExportGenesis().State.NativeStakingInjects)

	_, err = k.ExecuteContract(types.MsgExecuteContract{Sender: wallet, ContractAddress: official.ContractAddressUser, Msg: []byte("blocked"), Height: 13})
	require.ErrorContains(t, err, types.ErrAccountFrozen)
}

func TestInternalMessagesAndExportImportAreDeterministic(t *testing.T) {
	wallet := aeAddress("11")
	destination := aeAddress("22")
	k := NewKeeperWithAccountStatus(testAccountStatus{wallet: accountStatusActive})
	codeHash := storeContractCode(t, &k, wallet)
	contract := instantiateContract(t, &k, wallet, codeHash, "internal", 10, 1000, 0)

	message, err := k.ReceiveInternalMessage(types.MsgReceiveInternalMessage{
		SourceContractUser:	contract.ContractAddressUser,
		DestinationAccount:	destination,
		Funds:			7,
		Body:			[]byte("hello"),
		Height:			11,
	})
	require.NoError(t, err)
	require.Equal(t, contract.ContractAddressUser, message.SourceContractUser)
	require.Equal(t, destination, message.DestinationAccount)

	exported := k.ExportGenesis()
	require.NoError(t, exported.Validate())
	roundTrip := NewKeeper()
	require.NoError(t, roundTrip.InitGenesis(exported))
	require.Equal(t, exported, roundTrip.ExportGenesis())
}

func TestContractsTypedMsgAndQueryServiceSurface(t *testing.T) {
	wallet := aeAddress("11")
	destination := aeAddress("22")
	k := NewKeeperWithAccountStatus(testAccountStatus{wallet: accountStatusActive})
	bytecode := []byte("AVM1 typed-service")
	stored, err := k.StoreCode(types.MsgStoreCode{Authority: wallet, Bytecode: bytecode})
	require.NoError(t, err)

	code, found, err := k.Code(types.QueryCodeRequest{CodeID: stored.CodeID})
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, types.CanonicalCodeHash(bytecode), code.CodeHash)
	codes, err := k.Codes(types.QueryCodesRequest{Pagination: types.PageRequest{Limit: 1}})
	require.NoError(t, err)
	require.Len(t, codes, 1)

	deployed, err := k.DeployContract(types.MsgDeployContract{
		Creator:	wallet,
		CodeID:		stored.CodeID,
		Salt:		"typed",
		InitPayload:	[]byte("init"),
		InitialBalance:	1_000,
		Admin:		wallet,
		Height:		20,
	})
	require.NoError(t, err)
	executed, err := k.ExecuteExternal(types.MsgExecuteExternal{
		Sender:			wallet,
		ContractAddress:	deployed.ContractAddressUser,
		Payload:		[]byte("call"),
		GasLimit:		k.Params().MaxGasPerExecution,
		Height:			21,
	})
	require.NoError(t, err)
	require.Equal(t, deployed.ContractAddressUser, executed.ContractAddressUser)

	stateRoot, err := k.ContractStateRoot(types.QueryContractStateRootRequest{ContractAddress: deployed.ContractAddressUser})
	require.NoError(t, err)
	require.NotEmpty(t, stateRoot)
	contracts, err := k.Contracts(types.QueryContractsRequest{Pagination: types.PageRequest{Limit: 1}})
	require.NoError(t, err)
	require.Len(t, contracts, 1)

	internal, err := k.SendInternalMessage(types.MsgSendInternalMessage{
		Message: types.InternalMessage{
			SourceContractUser:	deployed.ContractAddressUser,
			DestinationAccount:	destination,
			Funds:			5,
			Opcode:			7,
			QueryID:		9,
			Body:			[]byte("internal"),
			Bounce:			true,
			Deadline:		25,
			GasLimit:		100,
			LogicalTime:		3,
		},
		Height:	22,
	})
	require.NoError(t, err)
	require.NotEmpty(t, internal.MessageID)
	require.Equal(t, types.ComputeInternalMessageID(internal), internal.MessageID)
	queue, err := k.ContractQueue(types.QueryContractQueueRequest{ContractAddress: deployed.ContractAddressUser, Pagination: types.PageRequest{Limit: 10}})
	require.NoError(t, err)
	require.Equal(t, []types.InternalMessage{internal}, queue)
	storage, err := k.ContractStorage(types.QueryContractStorageRequest{ContractAddress: deployed.ContractAddressUser, Pagination: types.PageRequest{Limit: 1}})
	require.NoError(t, err)
	require.Equal(t, []types.ContractStorageEntry{{
		ContractAddress:	deployed.ContractAddressUser,
		Key:			[]byte("data"),
		Value:			[]byte("call"),
	}}, storage)
	receipts, err := k.ContractReceipts(types.QueryContractReceiptsRequest{ContractAddress: deployed.ContractAddressUser, Pagination: types.PageRequest{Limit: 10}})
	require.NoError(t, err)
	require.Len(t, receipts, 3)
	require.Equal(t, "deploy", receipts[0].Operation)
	require.Equal(t, "execute", receipts[1].Operation)
	require.Equal(t, "internal_message_queued", receipts[2].Operation)
	require.NoError(t, k.ContractEvents(types.QueryContractEventsRequest{ContractAddress: deployed.ContractAddressUser, Pagination: types.PageRequest{Limit: 1}}))
}

func TestStateInitCounterfactualDeployVirtualQueryAndExternalAttachment(t *testing.T) {
	wallet := aeAddress("11")
	k := NewKeeperWithAccountStatus(testAccountStatus{wallet: accountStatusActive})
	codeHash := storeContractCode(t, &k, wallet)
	stateInit := types.NewStateInit(wallet, codeHash, []byte("init"), "counterfactual", 1_000)
	expectedUser, expectedRaw, err := types.DeriveContractAddressFromStateInit("chain-a", "zone-a", wallet, stateInit, k.Params())
	require.NoError(t, err)

	virtual, err := k.Contract(types.QueryContractRequest{
		ChainID:	"chain-a",
		Namespace:	"zone-a",
		Deployer:	wallet,
		StateInit:	&stateInit,
	})
	require.NoError(t, err)
	require.False(t, virtual.Found)
	require.True(t, virtual.Virtual)
	require.Equal(t, expectedUser, virtual.ContractAddress)

	deployed, err := k.DeployContract(types.MsgDeployContract{
		Creator:	wallet,
		CodeID:		codeHash,
		ChainID:	"chain-a",
		Namespace:	"zone-a",
		StateInit:	&stateInit,
		InitPayload:	[]byte("init"),
		InitialBalance:	1_000,
		Height:		20,
	})
	require.NoError(t, err)
	require.Equal(t, expectedUser, deployed.ContractAddressUser)
	require.Equal(t, expectedRaw, deployed.ContractAddressRaw)

	_, err = k.DeployContract(types.MsgDeployContract{
		Creator:	wallet,
		CodeID:		codeHash,
		ChainID:	"chain-a",
		Namespace:	"zone-a",
		StateInit:	&stateInit,
		InitPayload:	[]byte("init"),
		InitialBalance:	1_000,
		Height:		21,
	})
	require.ErrorContains(t, err, "already exists")

	lazyInit := types.NewStateInit(wallet, codeHash, []byte("lazy-init"), "lazy", 1_000)
	lazyAddress, _, err := types.DeriveContractAddressFromStateInit("chain-a", "zone-a", wallet, lazyInit, k.Params())
	require.NoError(t, err)
	executed, err := k.ExecuteExternal(types.MsgExecuteExternal{
		Sender:			wallet,
		ContractAddress:	lazyAddress,
		ChainID:		"chain-a",
		Namespace:		"zone-a",
		StateInit:		&lazyInit,
		Payload:		[]byte("call"),
		GasLimit:		k.Params().MaxGasPerExecution,
		Height:			22,
	})
	require.NoError(t, err)
	require.Equal(t, lazyAddress, executed.ContractAddressUser)

	query, err := k.Contract(types.QueryContractRequest{ContractAddress: lazyAddress})
	require.NoError(t, err)
	require.True(t, query.Found)
	require.Equal(t, []byte("call"), query.Contract.Data)
	require.NotEmpty(t, query.Contract.StateInitHash)
	stateInitHash, err := types.HashStateInit(lazyInit)
	require.NoError(t, err)
	require.Equal(t, stateInitHash, query.Contract.StateInitHash)

	exported := k.ExportGenesis()
	imported := NewKeeper()
	require.NoError(t, imported.InitGenesis(exported))
	roundTrip, err := imported.Contract(types.QueryContractRequest{ContractAddress: lazyAddress})
	require.NoError(t, err)
	require.Equal(t, query.Contract.StateInitHash, roundTrip.Contract.StateInitHash)
}

func TestInternalMessageChargesStorageRentBeforeSend(t *testing.T) {
	wallet := aeAddress("11")
	destination := aeAddress("22")
	k := NewKeeperWithAccountStatus(testAccountStatus{wallet: accountStatusActive})
	codeHash := storeContractCode(t, &k, wallet)
	contract := instantiateContract(t, &k, wallet, codeHash, "internal-rent", 10, 1, 0)

	_, err := k.ReceiveInternalMessage(types.MsgReceiveInternalMessage{
		SourceContractUser:	contract.ContractAddressUser,
		DestinationAccount:	destination,
		Funds:			7,
		Body:			[]byte("hello"),
		Height:			12,
	})
	require.ErrorContains(t, err, types.ErrStorageRent)

	query, err := k.Contract(types.QueryContractRequest{ContractAddress: contract.ContractAddressUser})
	require.NoError(t, err)
	require.Equal(t, types.ContractStatusFrozen, query.Contract.Status)
	require.Empty(t, k.ExportGenesis().State.InternalMessages)
}

func TestAVMContractLifecycleStateMachineEnforced(t *testing.T) {
	wallet := aeAddress("11")
	other := aeAddress("22")
	k := NewKeeperWithAccountStatus(testAccountStatus{wallet: accountStatusActive, other: accountStatusActive})
	codeHash := storeContractCode(t, &k, wallet)
	codeV2 := sha256Hex("lifecycle-code-v2")
	_, err := k.StoreCode(types.MsgStoreCode{Authority: wallet, CodeHash: codeV2, CodeBytes: 256})
	require.NoError(t, err)

	active := instantiateContract(t, &k, wallet, codeHash, "lifecycle-active", 10, 500, 0)
	executed, err := k.ExecuteContract(types.MsgExecuteContract{
		Sender:			wallet,
		ContractAddress:	active.ContractAddressUser,
		Msg:			[]byte("active-call"),
		Height:			11,
	})
	require.NoError(t, err)
	require.Equal(t, active.ContractAddressUser, executed.ContractAddressUser)

	frozen := instantiateContract(t, &k, wallet, codeHash, "lifecycle-frozen", 20, 500, 0)
	setContractLifecycle(t, &k, frozen.ContractAddressUser, types.ContractStatusFrozen, func(contract *types.Contract) {
		contract.StorageRentDebt = 25
	})
	frozenQuery, err := k.Contract(types.QueryContractRequest{ContractAddress: frozen.ContractAddressUser})
	require.NoError(t, err)
	beforeFrozen := frozenQuery.Contract

	_, err = k.ExecuteContract(types.MsgExecuteContract{Sender: wallet, ContractAddress: frozen.ContractAddressUser, Msg: []byte("blocked"), Height: 21})
	require.ErrorContains(t, err, types.ErrAccountFrozen)
	_, err = k.ReceiveInternalMessage(types.MsgReceiveInternalMessage{SourceContractUser: active.ContractAddressUser, DestinationAccount: frozen.ContractAddressUser, Height: 22})
	require.ErrorContains(t, err, types.ErrAccountFrozen)

	topped, err := k.TopUpContract(types.MsgTopUpContract{Sender: wallet, ContractAddress: frozen.ContractAddressUser, Amount: 25, Height: 23})
	require.NoError(t, err)
	require.Equal(t, beforeFrozen.CodeID, topped.CodeID)
	require.Equal(t, beforeFrozen.Data, topped.Data)
	require.Equal(t, beforeFrozen.StateRoot, topped.StateRoot)
	paid, err := k.PayContractStorageDebt(types.MsgPayContractStorageDebt{Sender: wallet, ContractAddress: frozen.ContractAddressUser, Amount: 25, Height: 24})
	require.NoError(t, err)
	require.Zero(t, paid.StorageRentDebt)
	unfrozen, err := k.UnfreezeContract(types.MsgUnfreezeContract{Sender: wallet, ContractAddress: frozen.ContractAddressUser, Height: 25})
	require.NoError(t, err)
	require.Equal(t, types.ContractStatusActive, unfrozen.Status)
	require.Equal(t, paid.CodeID, unfrozen.CodeID)
	require.Equal(t, paid.Data, unfrozen.Data)
	require.Equal(t, paid.Balance, unfrozen.Balance)
	require.Equal(t, paid.StateRoot, unfrozen.StateRoot)

	frozenLimited := instantiateContract(t, &k, wallet, codeHash, "lifecycle-frozen-limited", 30, 500, 0)
	setContractLifecycle(t, &k, frozenLimited.ContractAddressUser, types.ContractStatusFrozenLimited, func(contract *types.Contract) {
		contract.Upgradeable = true
		contract.StorageRentDebt = 10
	})
	_, err = k.ExecuteContract(types.MsgExecuteContract{Sender: wallet, ContractAddress: frozenLimited.ContractAddressUser, Msg: []byte("blocked"), Height: 31})
	require.ErrorContains(t, err, types.ErrAccountFrozen)
	_, err = k.ReceiveInternalMessage(types.MsgReceiveInternalMessage{SourceContractUser: frozenLimited.ContractAddressUser, DestinationAccount: other, Height: 32})
	require.ErrorContains(t, err, types.ErrAccountFrozen)
	_, err = k.UpgradeContractCode(types.MsgUpgradeContractCode{
		Actor:	wallet, ContractAddress: frozenLimited.ContractAddressUser, NewCodeID: codeV2, MigrationHandler: "schema_only", Height: 33,
	})
	require.ErrorContains(t, err, types.ErrAccountFrozen)
	_, err = k.TopUpContract(types.MsgTopUpContract{Sender: wallet, ContractAddress: frozenLimited.ContractAddressUser, Amount: 10, Height: 34})
	require.NoError(t, err)
	_, err = k.PayContractStorageDebt(types.MsgPayContractStorageDebt{Sender: wallet, ContractAddress: frozenLimited.ContractAddressUser, Amount: 10, Height: 35})
	require.NoError(t, err)
	_, err = k.UnfreezeContract(types.MsgUnfreezeContract{Sender: wallet, ContractAddress: frozenLimited.ContractAddressUser, Height: 36})
	require.NoError(t, err)

	archived := instantiateContract(t, &k, wallet, codeHash, "lifecycle-archived", 40, 500, 0)
	setContractLifecycle(t, &k, archived.ContractAddressUser, types.ContractStatusArchived, func(contract *types.Contract) {
		contract.Upgradeable = true
	})
	_, err = k.ExecuteContract(types.MsgExecuteContract{Sender: wallet, ContractAddress: archived.ContractAddressUser, Msg: []byte("blocked"), Height: 41})
	require.ErrorContains(t, err, types.ErrContractLifecycle)
	_, err = k.ReceiveInternalMessage(types.MsgReceiveInternalMessage{SourceContractUser: archived.ContractAddressUser, DestinationAccount: other, Height: 42})
	require.ErrorContains(t, err, types.ErrContractLifecycle)
	_, err = k.MigrateContractState(types.MsgMigrateContractState{
		Actor:	wallet, ContractAddress: archived.ContractAddressUser, FromSchemaVersion: 1, ToSchemaVersion: 2, MigrationHandler: "schema_only", Height: 43,
	})
	require.ErrorContains(t, err, types.ErrContractLifecycle)
	archivedRoot, err := k.ContractStateRoot(types.QueryContractStateRootRequest{ContractAddress: archived.ContractAddressUser})
	require.NoError(t, err)
	require.NotEmpty(t, archivedRoot)

	deleted := instantiateContract(t, &k, wallet, codeHash, "lifecycle-deleted", 50, 500, 0)
	setContractLifecycle(t, &k, deleted.ContractAddressUser, types.ContractStatusDeleted, func(contract *types.Contract) {
		contract.Balance = 0
		contract.StorageRentDebt = 0
	})
	_, err = k.ExecuteContract(types.MsgExecuteContract{Sender: wallet, ContractAddress: deleted.ContractAddressUser, Msg: []byte("blocked"), Height: 51})
	require.ErrorContains(t, err, types.ErrContractLifecycle)
	_, err = k.TopUpContract(types.MsgTopUpContract{Sender: wallet, ContractAddress: deleted.ContractAddressUser, Amount: 1, Height: 52})
	require.ErrorContains(t, err, types.ErrContractLifecycle)
	deletedQuery, err := k.Contract(types.QueryContractRequest{ContractAddress: deleted.ContractAddressUser})
	require.NoError(t, err)
	require.True(t, deletedQuery.Found)
	require.Equal(t, types.ContractStatusDeleted, deletedQuery.Contract.Status)
}

type testAccountStatus map[string]string

func (s testAccountStatus) AccountStatus(_ context.Context, address string) (string, bool, error) {
	status, found := s[address]
	return status, found, nil
}

func storeContractCode(t *testing.T, k *Keeper, owner string) string {
	t.Helper()
	sum := sha256Hex("code/" + owner)
	response, err := k.StoreCode(types.MsgStoreCode{Authority: owner, CodeHash: sum, CodeBytes: 128})
	require.NoError(t, err)
	return response.CodeID
}

func instantiateContract(t *testing.T, k *Keeper, owner string, codeHash string, salt string, height uint64, funds uint64, storageBytes uint64) types.InstantiateContractResponse {
	t.Helper()
	created, err := k.InstantiateContract(types.MsgInstantiateContract{
		Creator:	owner,
		CodeID:		codeHash,
		InitMsg:	[]byte("init"),
		Funds:		funds,
		Admin:		owner,
		Salt:		salt,
		StorageBytes:	0,
		Height:		height,
	})
	require.NoError(t, err)
	return created
}

func setContractLifecycle(t *testing.T, k *Keeper, contractAddress string, status string, mutate func(*types.Contract)) {
	t.Helper()
	gs := k.ExportGenesis()
	for i := range gs.State.Contracts {
		if gs.State.Contracts[i].AddressUser != contractAddress {
			continue
		}
		gs.State.Contracts[i].Status = status
		if mutate != nil {
			mutate(&gs.State.Contracts[i])
		}
		require.NoError(t, k.InitGenesis(gs))
		return
	}
	t.Fatalf("contract %s not found", contractAddress)
}

func contractStorageBytes(codeBytes uint64, data []byte) uint64 {
	return codeBytes + uint64(len(data))
}

func aeAddress(hexByte string) string {
	bz, err := hex.DecodeString(strings.Repeat(hexByte, 20))
	if err != nil {
		panic(err)
	}
	return addressing.FormatAccAddress(sdk.AccAddress(bz))
}

func sha256Hex(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])
}

func stringsHasPrefix(text string, prefix string) bool {
	return strings.HasPrefix(text, prefix)
}
