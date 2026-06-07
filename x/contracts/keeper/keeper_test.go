package keeper

import (
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

	contractAddress, _, err := types.DeriveContractAddress(authority, codeHash, "query")
	require.NoError(t, err)
	query, err := keeper.Contract(types.QueryContractRequest{ContractAddress: contractAddress})
	require.NoError(t, err)
	require.False(t, query.Found)
	require.Equal(t, contractAddress, query.ContractAddress)

	_, err = keeper.Contract(types.QueryContractRequest{})
	require.ErrorContains(t, err, types.ErrContractNotFound)
}

func TestWalletInstantiatesExecutesAndPassesFunds(t *testing.T) {
	wallet := aeAddress("11")
	k := NewKeeperWithAccountStatus(testAccountStatus{wallet: accountStatusActive})
	codeHash := storeContractCode(t, &k, wallet)

	created, err := k.InstantiateContract(types.MsgInstantiateContract{
		Creator:      wallet,
		CodeID:       codeHash,
		InitMsg:      []byte(`{"owner":"wallet"}`),
		Funds:        100,
		Admin:        wallet,
		Salt:         "contract-a",
		StorageBytes: 4,
		Height:       10,
	})
	require.NoError(t, err)
	require.Equal(t, wallet, created.Owner)
	require.Equal(t, wallet, created.Admin)
	require.Equal(t, uint64(100), created.Balance)
	require.True(t, stringsHasPrefix(created.ContractAddressUser, "AE"))
	require.True(t, stringsHasPrefix(created.ContractAddressRaw, "4:"))
	require.Equal(t, types.EventTypeContractInstantiated, created.Events[0].Type)
	require.Equal(t, created.ContractAddressUser, created.Events[0].Contract)
	require.Equal(t, created.ContractAddressRaw, created.Events[0].InternalRaw)

	executed, err := k.ExecuteContract(types.MsgExecuteContract{
		Sender:          wallet,
		ContractAddress: created.ContractAddressUser,
		Msg:             []byte(`{"transfer":1}`),
		Funds:           25,
		Height:          11,
	})
	require.NoError(t, err)
	require.Equal(t, created.ContractAddressUser, executed.ContractAddressUser)
	require.Equal(t, uint64(121), executed.Balance)
	require.Equal(t, types.EventTypeContractExecuted, executed.Events[0].Type)
	require.Equal(t, created.ContractAddressUser, executed.Events[0].Contract)
	require.Equal(t, created.ContractAddressRaw, executed.Events[0].InternalRaw)
}

func TestFrozenWalletCannotInstantiateOrExecuteUntilUnfrozen(t *testing.T) {
	wallet := aeAddress("11")
	status := testAccountStatus{wallet: accountStatusActive}
	k := NewKeeperWithAccountStatus(status)
	codeHash := storeContractCode(t, &k, wallet)
	created := instantiateContract(t, &k, wallet, codeHash, "contract-a", 10, 10, 2)

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
		Creator: wallet, CodeID: codeHash, Admin: rawAdmin, Salt: "raw-admin", Height: 9,
	})
	require.ErrorContains(t, err, "AE user-facing")

	created, err := k.InstantiateContract(types.MsgInstantiateContract{
		Creator: wallet, CodeID: codeHash, Admin: admin, Salt: "asset-contract", Height: 10,
	})
	require.NoError(t, err)
	require.True(t, stringsHasPrefix(created.Owner, "AE"))
	require.True(t, stringsHasPrefix(created.Admin, "AE"))

	err = k.SetAssetOwner(types.AssetOwnershipRecord{
		AssetType:           types.AssetTypeNFT,
		ContractAddressUser: created.ContractAddressUser,
		AssetID:             "nft-1",
		Owner:               assetOwner,
	})
	require.NoError(t, err)
	owner, err := k.AssetOwner(types.QueryAssetOwnerRequest{AssetType: types.AssetTypeNFT, ContractAddressUser: created.ContractAddressUser, AssetID: "nft-1"})
	require.NoError(t, err)
	require.True(t, owner.Found)
	require.Equal(t, assetOwner, owner.Owner)

	exported := k.ExportGenesis()
	require.NotEqual(t, "token-balance", exported.State.Contracts[0].Owner)
	require.Len(t, exported.State.AssetOwnership, 1)
}

func TestStakeReputationCannotBeRepresentedAsTokenOrNFTAsset(t *testing.T) {
	wallet := aeAddress("11")
	owner := aeAddress("12")
	k := NewKeeperWithAccountStatus(testAccountStatus{wallet: accountStatusActive})
	codeHash := storeContractCode(t, &k, wallet)
	created := instantiateContract(t, &k, wallet, codeHash, "reputation-asset", 10, 0, 0)

	err := k.SetAssetOwner(types.AssetOwnershipRecord{
		AssetType:           types.AssetTypeNFT,
		ContractAddressUser: created.ContractAddressUser,
		AssetID:             "stake_reputation/" + owner,
		Owner:               owner,
	})
	require.ErrorContains(t, err, "cannot be transferred as token or NFT")

	err = k.SetAssetOwner(types.AssetOwnershipRecord{
		AssetType:           types.AssetTypeToken,
		ContractAddressUser: created.ContractAddressUser,
		AssetID:             "reputation/stake/" + owner,
		Owner:               owner,
	})
	require.ErrorContains(t, err, "cannot be transferred as token or NFT")
}

func TestOfficialLiquidStakingContractCapabilityAllowsNativeHookOnlyForAuthorizedContract(t *testing.T) {
	wallet := aeAddress("11")
	k := NewKeeperWithAccountStatus(testAccountStatus{wallet: accountStatusActive})
	codeHash := storeContractCode(t, &k, wallet)
	official := instantiateContract(t, &k, wallet, codeHash, "official-lst", 10, 100, 0)
	unauthorized := instantiateContract(t, &k, wallet, codeHash, "other-contract", 11, 100, 0)

	capability, err := k.GrantNativeStakingCapability(types.MsgGrantNativeStakingCapability{
		Authority:           types.DefaultParams().Authority,
		ContractAddressUser: official.ContractAddressUser,
		ContractAddressRaw:  official.ContractAddressRaw,
		PoolID:              "official-pool",
		Height:              12,
	})
	require.NoError(t, err)
	require.Equal(t, official.ContractAddressUser, capability.ContractAddressUser)
	require.Equal(t, official.ContractAddressRaw, capability.ContractAddressRaw)

	injection, err := k.InjectNativeStaking(types.MsgInjectNativeStaking{
		CallerContractUser: official.ContractAddressUser,
		CallerContractRaw:  official.ContractAddressRaw,
		PoolID:             "official-pool",
		Amount:             500,
		Height:             13,
	})
	require.NoError(t, err)
	require.Equal(t, official.ContractAddressUser, injection.ContractAddressUser)
	require.Equal(t, official.ContractAddressRaw, injection.ContractAddressRaw)

	_, err = k.InjectNativeStaking(types.MsgInjectNativeStaking{
		CallerContractUser: unauthorized.ContractAddressUser,
		CallerContractRaw:  unauthorized.ContractAddressRaw,
		PoolID:             "official-pool",
		Amount:             1,
		Height:             14,
	})
	require.ErrorContains(t, err, types.ErrUnauthorized)
}

func TestNativeStakingCapabilityRejectsBadAuthorityAndFrozenContract(t *testing.T) {
	wallet := aeAddress("11")
	k := NewKeeperWithAccountStatus(testAccountStatus{wallet: accountStatusActive})
	codeHash := storeContractCode(t, &k, wallet)
	official := instantiateContract(t, &k, wallet, codeHash, "official-frozen", 10, 100, 0)

	_, err := k.GrantNativeStakingCapability(types.MsgGrantNativeStakingCapability{
		Authority:           wallet,
		ContractAddressUser: official.ContractAddressUser,
		ContractAddressRaw:  official.ContractAddressRaw,
		PoolID:              "official-pool",
		Height:              11,
	})
	require.ErrorContains(t, err, types.ErrUnauthorized)

	_, err = k.GrantNativeStakingCapability(types.MsgGrantNativeStakingCapability{
		Authority:           types.DefaultParams().Authority,
		ContractAddressUser: official.ContractAddressUser,
		ContractAddressRaw:  official.ContractAddressRaw,
		PoolID:              "official-pool",
		Height:              12,
	})
	require.NoError(t, err)

	gs := k.ExportGenesis()
	gs.State.Contracts[0].Status = types.ContractStatusFrozen
	require.NoError(t, k.InitGenesis(gs))

	_, err = k.InjectNativeStaking(types.MsgInjectNativeStaking{
		CallerContractUser: official.ContractAddressUser,
		CallerContractRaw:  official.ContractAddressRaw,
		PoolID:             "official-pool",
		Amount:             500,
		Height:             13,
	})
	require.ErrorContains(t, err, types.ErrAccountFrozen)

	query, err := k.Contract(types.QueryContractRequest{ContractAddress: official.ContractAddressUser})
	require.NoError(t, err)
	require.True(t, query.Found)
	require.Equal(t, types.ContractStatusFrozen, query.Contract.Status)
}

func TestInternalMessagesAndExportImportAreDeterministic(t *testing.T) {
	wallet := aeAddress("11")
	destination := aeAddress("22")
	k := NewKeeperWithAccountStatus(testAccountStatus{wallet: accountStatusActive})
	codeHash := storeContractCode(t, &k, wallet)
	contract := instantiateContract(t, &k, wallet, codeHash, "internal", 10, 100, 0)

	message, err := k.ReceiveInternalMessage(types.MsgReceiveInternalMessage{
		SourceContractUser: contract.ContractAddressUser,
		DestinationAccount: destination,
		Funds:              7,
		Body:               []byte("hello"),
		Height:             11,
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

type testAccountStatus map[string]string

func (s testAccountStatus) AccountStatus(address string) (string, bool) {
	status, found := s[address]
	return status, found
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
		Creator:      owner,
		CodeID:       codeHash,
		InitMsg:      []byte("init"),
		Funds:        funds,
		Admin:        owner,
		Salt:         salt,
		StorageBytes: storageBytes,
		Height:       height,
	})
	require.NoError(t, err)
	return created
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
