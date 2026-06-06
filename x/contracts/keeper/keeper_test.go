package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	coretypes "github.com/sovereign-l1/l1/x/aethercore/types"
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
	codeHash := coretypes.DeterministicEmptyRootCommitment(coretypes.RootType(types.ModuleName), "code")

	response, err := keeper.StoreCode(types.MsgStoreCode{Authority: "authority", CodeHash: codeHash, CodeBytes: 128})
	require.NoError(t, err)
	require.Equal(t, codeHash, response.CodeID)
	require.NotEmpty(t, response.StateRoot)

	_, err = keeper.StoreCode(types.MsgStoreCode{Authority: "authority", CodeHash: codeHash, CodeBytes: 0})
	require.ErrorContains(t, err, types.ErrInvalidBytecode)

	query, err := keeper.Contract(types.QueryContractRequest{ContractAddress: "contract1"})
	require.NoError(t, err)
	require.False(t, query.Found)
	require.Equal(t, "contract1", query.ContractAddress)

	_, err = keeper.Contract(types.QueryContractRequest{})
	require.ErrorContains(t, err, types.ErrContractNotFound)
}
