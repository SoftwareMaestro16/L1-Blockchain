package app

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/app/addressing"
	contractstypes "github.com/sovereign-l1/l1/x/contracts/types"
	nativeaccounttypes "github.com/sovereign-l1/l1/x/native-account/types"
)

func TestAVMRuntimeAppLevelDeployExecuteQueueStorageReceiptsAndExportImport(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false).WithBlockHeight(100)
	account := nativeAccountActivateViaRoute(t, app, ctx, nativeAccountModuleTestPubKey())

	require.Contains(t, app.ModuleManager.Modules, contractstypes.ModuleName)
	require.Contains(t, app.keys, contractstypes.StoreKey)
	require.NotNil(t, app.ContractsKeeper)
	require.NotNil(t, app.MsgServiceRouter().Handler(&contractstypes.MsgStoreCode{}))
	require.NotNil(t, app.MsgServiceRouter().Handler(&contractstypes.MsgDeployContract{}))
	require.NotNil(t, app.MsgServiceRouter().Handler(&contractstypes.MsgExecuteExternal{}))
	require.NotNil(t, app.MsgServiceRouter().Handler(&contractstypes.MsgSendInternalMessage{}))
	require.NotNil(t, app.GRPCQueryRouter().Route("/l1.contracts.v1.Query/ContractStorage"))
	require.NotNil(t, app.GRPCQueryRouter().Route("/l1.contracts.v1.Query/ContractReceipts"))

	storeRoute := app.MsgServiceRouter().Handler(&contractstypes.MsgStoreCode{})
	bytecode := []byte("AVM1 app-runtime deterministic")
	_, err := storeRoute(ctx, &contractstypes.MsgStoreCode{
		Authority:	account.AddressUser,
		Bytecode:	bytecode,
	})
	require.NoError(t, err)
	codeID := contractstypes.CanonicalCodeHash(bytecode)
	code, found, err := app.ContractsKeeper.Code(contractstypes.QueryCodeRequest{CodeID: codeID})
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, codeID, code.CodeID)

	deployRoute := app.MsgServiceRouter().Handler(&contractstypes.MsgDeployContract{})
	_, err = deployRoute(ctx.WithBlockHeight(101), &contractstypes.MsgDeployContract{
		Creator:	account.AddressUser,
		CodeID:		codeID,
		InitPayload:	[]byte("init"),
		InitialBalance:	1_000,
		Admin:		account.AddressUser,
		Salt:		"app-runtime",
		Height:		101,
	})
	require.NoError(t, err)
	contracts, err := app.ContractsKeeper.Contracts(contractstypes.QueryContractsRequest{Pagination: contractstypes.PageRequest{Limit: 10}})
	require.NoError(t, err)
	require.Len(t, contracts, 1)
	deployed := contractstypes.InstantiateContractResponse{
		ContractAddressUser:	contracts[0].AddressUser,
		ContractAddressRaw:	contracts[0].AddressRaw,
	}
	require.True(t, bytes.HasPrefix([]byte(deployed.ContractAddressUser), []byte("AE")))
	require.True(t, bytes.HasPrefix([]byte(deployed.ContractAddressRaw), []byte("4:")))

	executeRoute := app.MsgServiceRouter().Handler(&contractstypes.MsgExecuteExternal{})
	_, err = executeRoute(ctx.WithBlockHeight(102), &contractstypes.MsgExecuteExternal{
		Sender:			account.AddressUser,
		ContractAddress:	deployed.ContractAddressUser,
		Payload:		[]byte("call"),
		Funds:			25,
		GasLimit:		app.ContractsKeeper.Params().MaxGasPerExecution,
		Height:			102,
	})
	require.NoError(t, err)
	executed, err := app.ContractsKeeper.Contract(contractstypes.QueryContractRequest{ContractAddress: deployed.ContractAddressUser})
	require.NoError(t, err)
	require.Equal(t, deployed.ContractAddressUser, executed.Contract.AddressUser)

	destination := appAVMRuntimeAddress(0x77)
	internalRoute := app.MsgServiceRouter().Handler(&contractstypes.MsgSendInternalMessage{})
	_, err = internalRoute(ctx.WithBlockHeight(103), &contractstypes.MsgSendInternalMessage{
		Message: contractstypes.InternalMessage{
			SourceContractUser:	deployed.ContractAddressUser,
			DestinationAccount:	destination,
			Funds:			7,
			Opcode:			1,
			QueryID:		2,
			Body:			[]byte("internal"),
			GasLimit:		100,
			LogicalTime:		3,
			Height:			103,
		},
		Height:	103,
	})
	require.NoError(t, err)
	queue, err := app.ContractsKeeper.ContractQueue(contractstypes.QueryContractQueueRequest{
		ContractAddress:	deployed.ContractAddressUser,
		Pagination:		contractstypes.PageRequest{Limit: 10},
	})
	require.NoError(t, err)
	require.Len(t, queue, 1)

	storage, err := app.ContractsKeeper.ContractStorage(contractstypes.QueryContractStorageRequest{
		ContractAddress:	deployed.ContractAddressUser,
		Pagination:		contractstypes.PageRequest{Limit: 10},
	})
	require.NoError(t, err)
	require.Equal(t, []byte("data"), storage[0].Key)
	require.Equal(t, []byte("call"), storage[0].Value)

	receipts, err := app.ContractsKeeper.ContractReceipts(contractstypes.QueryContractReceiptsRequest{
		ContractAddress:	deployed.ContractAddressUser,
		Pagination:		contractstypes.PageRequest{Limit: 10},
	})
	require.NoError(t, err)
	require.Len(t, receipts, 3)
	require.Equal(t, "deploy", receipts[0].Operation)
	require.Equal(t, "execute", receipts[1].Operation)
	require.Equal(t, "internal_message_queued", receipts[2].Operation)
	require.NoError(t, app.RunAppInvariant(ctx, AppInvariantAVMQueueReceipts))

	exported, err := app.ContractsKeeper.ExportGenesisState(ctx)
	require.NoError(t, err)
	restarted := Setup(t, false)
	restartedCtx := restarted.NewContext(false).WithBlockHeight(200)
	require.NoError(t, restarted.ContractsKeeper.InitGenesisState(restartedCtx, exported))
	roundTrip, err := restarted.ContractsKeeper.ContractStorage(contractstypes.QueryContractStorageRequest{
		ContractAddress:	deployed.ContractAddressUser,
		Pagination:		contractstypes.PageRequest{Limit: 10},
	})
	require.NoError(t, err)
	require.Equal(t, storage, roundTrip)
	roundTripReceipts, err := restarted.ContractsKeeper.ContractReceipts(contractstypes.QueryContractReceiptsRequest{
		ContractAddress:	deployed.ContractAddressUser,
		Pagination:		contractstypes.PageRequest{Limit: 10},
	})
	require.NoError(t, err)
	require.Equal(t, receipts, roundTripReceipts)
}

func TestAVMRuntimeRejectsFrozenAccountsReservedZeroAndRecoversFrozenContracts(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false).WithBlockHeight(300)
	account := nativeAccountActivateViaRoute(t, app, ctx, nativeAccountModuleTestPubKey())
	codeID := appAVMRuntimeStoreCode(t, app, ctx, account.AddressUser)
	contract := appAVMRuntimeDeploy(t, app, ctx.WithBlockHeight(301), account.AddressUser, codeID, "frozen-runtime", 1, 301)

	account.Status = nativeaccounttypes.AccountStatusFrozen
	require.NoError(t, app.NativeAccountKeeper.SetAccount(ctx, account))
	executeRoute := app.MsgServiceRouter().Handler(&contractstypes.MsgExecuteExternal{})
	_, err := executeRoute(ctx.WithBlockHeight(302), &contractstypes.MsgExecuteExternal{
		Sender:			account.AddressUser,
		ContractAddress:	contract.ContractAddressUser,
		Payload:		[]byte("blocked"),
		GasLimit:		app.ContractsKeeper.Params().MaxGasPerExecution,
		Height:			302,
	})
	require.ErrorContains(t, err, contractstypes.ErrAccountFrozen)

	_, err = app.ContractsKeeper.StoreCodeState(ctx, contractstypes.MsgStoreCode{
		Authority:	addressing.SystemAddressAETMintUserFriendly,
		Bytecode:	[]byte("AVM1 reserved"),
	})
	require.ErrorContains(t, err, "reserved system address")
	_, err = app.ContractsKeeper.StoreCodeState(ctx, contractstypes.MsgStoreCode{
		Authority:	addressing.ZeroUserFriendly,
		Bytecode:	[]byte("AVM1 zero"),
	})
	require.ErrorContains(t, err, "zero address")

	account.Status = nativeaccounttypes.AccountStatusActive
	require.NoError(t, app.NativeAccountKeeper.SetAccount(ctx, account))
	_, err = executeRoute(ctx.WithBlockHeight(400), &contractstypes.MsgExecuteExternal{
		Sender:			account.AddressUser,
		ContractAddress:	contract.ContractAddressUser,
		Payload:		[]byte("rent-freeze"),
		GasLimit:		app.ContractsKeeper.Params().MaxGasPerExecution,
		Height:			400,
	})
	require.ErrorContains(t, err, contractstypes.ErrStorageRent)
	frozen, err := app.ContractsKeeper.Contract(contractstypes.QueryContractRequest{ContractAddress: contract.ContractAddressUser})
	require.NoError(t, err)
	require.Equal(t, contractstypes.ContractStatusFrozen, frozen.Contract.Status)
	require.NotZero(t, frozen.Contract.StorageRentDebt)

	_, err = app.ContractsKeeper.TopUpContractState(ctx.WithBlockHeight(401), contractstypes.MsgTopUpContract{
		Sender:			account.AddressUser,
		ContractAddress:	contract.ContractAddressUser,
		Amount:			frozen.Contract.StorageRentDebt + 100,
		Height:			401,
	})
	require.NoError(t, err)
	_, err = app.ContractsKeeper.PayContractStorageDebtState(ctx.WithBlockHeight(402), contractstypes.MsgPayContractStorageDebt{
		Sender:			account.AddressUser,
		ContractAddress:	contract.ContractAddressUser,
		Amount:			frozen.Contract.StorageRentDebt,
		Height:			402,
	})
	require.NoError(t, err)
	unfrozen, err := app.ContractsKeeper.UnfreezeContractState(ctx.WithBlockHeight(403), contractstypes.MsgUnfreezeContract{
		Sender:			account.AddressUser,
		ContractAddress:	contract.ContractAddressUser,
		Height:			403,
	})
	require.NoError(t, err)
	require.Equal(t, contractstypes.ContractStatusActive, unfrozen.Status)
}

func TestAVMRuntimeDeterminismGateRejectsNondeterminismAndKeepsStableRoots(t *testing.T) {
	params := contractstypes.DefaultParams()
	require.NoError(t, contractstypes.ValidateAVMBytecode(params, []byte("AVM1 set key value")))
	require.ErrorContains(t, contractstypes.ValidateAVMBytecode(params, []byte("AVM1 random")), contractstypes.ErrInvalidBytecode)

	app := Setup(t, false)
	ctx := app.NewContext(false).WithBlockHeight(500)
	account := nativeAccountActivateViaRoute(t, app, ctx, nativeAccountModuleTestPubKey())
	codeID := appAVMRuntimeStoreCode(t, app, ctx, account.AddressUser)
	contract := appAVMRuntimeDeploy(t, app, ctx.WithBlockHeight(501), account.AddressUser, codeID, "determinism", 1_000, 501)

	first, err := app.ContractsKeeper.ContractStateRoot(contractstypes.QueryContractStateRootRequest{ContractAddress: contract.ContractAddressUser})
	require.NoError(t, err)
	second, err := app.ContractsKeeper.ContractStateRoot(contractstypes.QueryContractStateRootRequest{ContractAddress: contract.ContractAddressUser})
	require.NoError(t, err)
	require.Equal(t, first, second)
}

func appAVMRuntimeStoreCode(t *testing.T, app *L1App, ctx sdk.Context, authority string) string {
	t.Helper()
	resp, err := app.ContractsKeeper.StoreCodeState(ctx, contractstypes.MsgStoreCode{
		Authority:	authority,
		Bytecode:	[]byte("AVM1 app-runtime helper"),
	})
	require.NoError(t, err)
	return resp.CodeID
}

func appAVMRuntimeDeploy(t *testing.T, app *L1App, ctx sdk.Context, creator string, codeID string, salt string, initialBalance uint64, height uint64) contractstypes.InstantiateContractResponse {
	t.Helper()
	resp, err := app.ContractsKeeper.DeployContractState(ctx, contractstypes.MsgDeployContract{
		Creator:	creator,
		CodeID:		codeID,
		InitPayload:	[]byte("init"),
		InitialBalance:	initialBalance,
		Admin:		creator,
		Salt:		salt,
		Height:		height,
	})
	require.NoError(t, err)
	return resp
}

func appAVMRuntimeAddress(fill byte) string {
	bz := bytes.Repeat([]byte{fill}, 20)
	return addressing.FormatAccAddress(sdk.AccAddress(bz))
}
