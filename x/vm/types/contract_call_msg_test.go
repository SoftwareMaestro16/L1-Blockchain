package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/x/aetravm/avm"
)

func TestMsgContractCallBuildsExecutableContractCall(t *testing.T) {
	state, code := uploadTestCode(t, counterSpecModule(), govAddr(), 10)
	state, contract, _, err := InstantiateAVMContract(state, govAddr(), code.CodeID, govAddr(), []byte("call-msg"), nil, 11, state.Policy.GasModel.DeployGas)
	require.NoError(t, err)

	msg := testMsgContractCall(contract.Address, state.Policy.GasModel.ExecuteGas)
	admission := testContractCallAdmission(msg, avm.EntryReceiveExternal)
	call, err := BuildContractCallFromMsg(msg, state, admission)
	require.NoError(t, err)
	require.Equal(t, msg.Caller, call.Actor)
	require.Equal(t, msg.ContractAddr, call.Contract)
	require.Equal(t, avm.EntryReceiveExternal, call.Entrypoint)
	require.Equal(t, msg.Args, call.Body)

	next, receipt, err := ExecuteAVMContract(state, call, admission.CreatedHeight)
	require.NoError(t, err)
	require.Equal(t, msg.ContractAddr, receipt.Contract)
	require.NotEqual(t, state.Export(), next.Export())
}

func TestMsgContractCallRejectsMissingContractDisabledMethodEscrowAndGas(t *testing.T) {
	state, code := uploadTestCode(t, counterSpecModule(), govAddr(), 10)
	state, contract, _, err := InstantiateAVMContract(state, govAddr(), code.CodeID, govAddr(), []byte("call-msg"), nil, 11, state.Policy.GasModel.DeployGas)
	require.NoError(t, err)

	msg := testMsgContractCall(contract.Address, state.Policy.GasModel.ExecuteGas)
	admission := testContractCallAdmission(msg, avm.EntryReceiveExternal)

	missing := msg
	missing.ContractAddr = userAddr(9)
	require.ErrorContains(t, missing.Validate(state, admission), "exist")

	disabledMethod := admission.Normalize()
	disabledMethod.Methods[0].Enabled = false
	require.ErrorContains(t, msg.Validate(state, disabledMethod), "method")

	missingEscrow := admission
	missingEscrow.Escrows = nil
	require.ErrorContains(t, msg.Validate(state, missingEscrow), "escrow")

	overGas := msg
	overGas.GasLimit = state.Policy.GasModel.ExecuteGas + 1
	require.ErrorContains(t, overGas.Validate(state, admission), "gas limit")
}

func testMsgContractCall(contractAddr sdkAddr, gasLimit uint64) MsgContractCall {
	return MsgContractCall{
		Caller:			govAddr(),
		ContractAddr:		contractAddr,
		Method:			"counter.increment",
		Args:			[]byte("increment"),
		Funds:			sdkmath.NewInt(5),
		GasLimit:		gasLimit,
		ReplyToOptional:	govAddr(),
		ExpiryHeight:		30,
	}
}

func testContractCallAdmission(msg MsgContractCall, entrypoint avm.Entrypoint) ContractCallAdmission {
	return ContractCallAdmission{
		CreatedHeight:	20,
		Methods: []ContractMethodAdmission{{
			ContractAddr:	msg.ContractAddr,
			Method:		msg.Method,
			Entrypoint:	entrypoint,
			Enabled:	true,
		}},
		Escrows: []ContractFundsEscrow{{
			Caller:		msg.Caller,
			ContractAddr:	msg.ContractAddr,
			Amount:		msg.Funds,
			ExpiryHeight:	msg.ExpiryHeight,
			Escrowed:	true,
		}},
		MaxArgsBytes:	128,
	}
}
