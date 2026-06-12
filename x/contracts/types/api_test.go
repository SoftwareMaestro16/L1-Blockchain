package types

import (
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/app/addressing"
)

func TestContractsTxAPIValidationRejectsMalformedAddressesAndBounds(t *testing.T) {
	params := DefaultParams()
	sender := contractAPIAddress(0x11)
	stateInit := NewStateInit(sender, strings.Repeat("a", 64), nil, "api", 0)
	contract, _, err := DeriveContractAddressFromStateInit("", "", sender, stateInit, params)
	require.NoError(t, err)

	require.NoError(t, MsgDeployContract{
		Creator:	sender,
		CodeID:		strings.Repeat("a", 64),
		InitPayload:	[]byte("init"),
		Admin:		sender,
		Height:		1,
	}.ValidateBasic(params))
	require.ErrorContains(t, MsgDeployContract{Creator: sender, CodeID: "code", InitPayload: make([]byte, MaxContractPayloadBytes+1), Height: 1}.ValidateBasic(params), "payload")
	require.ErrorContains(t, MsgDeployContract{Creator: sender, CodeID: "code", Metadata: make([]byte, MaxContractMetadataBytes+1), Height: 1}.ValidateBasic(params), "metadata")

	require.NoError(t, MsgExecuteExternal{
		Sender:			sender,
		ContractAddress:	contract,
		Payload:		[]byte("call"),
		GasLimit:		params.MaxGasPerExecution,
		Height:			2,
	}.ValidateBasic(params))
	require.ErrorContains(t, MsgExecuteExternal{Sender: sender, ContractAddress: contract, GasLimit: params.MaxGasPerExecution + 1, Height: 2}.ValidateBasic(params), "gas limit")
	require.Error(t, MsgExecuteExternal{Sender: sender, ContractAddress: "4:" + strings.Repeat("00", 32), GasLimit: 1, Height: 2}.ValidateBasic(params))
}

func TestContractsStoreCodeAndQueryAPIValidation(t *testing.T) {
	params := DefaultParams()
	sender := contractAPIAddress(0x22)
	bytecode := []byte("AVM1 deterministic")
	require.NoError(t, MsgStoreCode{Authority: sender, Bytecode: bytecode}.ValidateBasic(params))
	require.ErrorContains(t, MsgStoreCode{Authority: sender, Bytecode: []byte("AVM1 random")}.ValidateBasic(params), ErrInvalidBytecode)

	require.NoError(t, ValidateQueryPagination(PageRequest{Limit: MaxContractQueryLimit}))
	require.ErrorContains(t, ValidateQueryPagination(PageRequest{}), "query limit")
	require.ErrorContains(t, ValidateQueryPagination(PageRequest{Limit: MaxContractQueryLimit + 1}), "query limit")
}

func contractAPIAddress(fill byte) string {
	bz := make([]byte, 20)
	for i := range bz {
		bz[i] = fill
	}
	return addressing.FormatAccAddress(sdk.AccAddress(bz))
}
