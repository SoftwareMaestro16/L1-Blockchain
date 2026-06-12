package types

import (
	"encoding/hex"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestMsgCrossZoneCallMatchesSectionTwelveThree(t *testing.T) {
	params := testMessageParams()
	call := testCrossZoneCall()
	admission := testCrossZoneAdmission(call)

	msg, err := NewMessageFromCrossZoneCall(call, admission, params)
	require.NoError(t, err)
	require.Equal(t, call.Caller, msg.Sender)
	require.Equal(t, call.Callee, msg.Recipient)
	require.Equal(t, call.SourceZoneID, msg.SourceZone)
	require.Equal(t, call.DestinationZoneID, msg.DestinationZone)
	require.Equal(t, call.PayloadType, msg.Opcode)
	require.Equal(t, call.Payload, msg.Payload)
	require.Equal(t, call.ValueNAET, msg.Value)
	require.Equal(t, call.GasLimit, msg.GasLimit)
	require.Equal(t, call.ForwardingFee, msg.FeeLimit)
	require.Equal(t, call.ExpiryHeight, msg.Deadline)
	require.Equal(t, "cross-zone-call/"+ReplyModeCaller, msg.AuthScope)

	oversized := call
	oversized.Payload = make([]byte, params.MaxPayloadSize+1)
	require.ErrorContains(t, oversized.Validate(admission, params), "payload size")

	unsupportedPayload := call
	unsupportedPayload.PayloadType = "unsupported.payload"
	require.ErrorContains(t, unsupportedPayload.Validate(admission, params), "payload type")

	missingEscrow := admission
	missingEscrow.Escrows = nil
	require.ErrorContains(t, call.Validate(missingEscrow, params), "escrow")

	unsupportedReply := call
	unsupportedReply.ReplyMode = "stream"
	require.ErrorContains(t, unsupportedReply.Validate(admission, params), "reply mode")
}

func TestMsgContractCallBuildsCanonicalMessageAetherEnvelopeAndReceipt(t *testing.T) {
	params := testContractCallParams()
	call := testMessageLayerContractCall()
	admission := testMessageLayerContractAdmission(call)

	msg, err := NewMessageFromContractCall(call, admission, params)
	require.NoError(t, err)
	require.Equal(t, zonestypes.ZoneIDContract, msg.SourceZone)
	require.Equal(t, zonestypes.ZoneIDContract, msg.DestinationZone)
	require.Equal(t, call.Caller, msg.Sender)
	require.Equal(t, call.ContractAddr, msg.Recipient)
	require.Equal(t, ContractCallPayloadType, msg.Opcode)
	require.Equal(t, call.Funds, msg.Value)
	require.Equal(t, params.MinFeeLimit, msg.FeeLimit)
	require.Equal(t, call.GasLimit, msg.GasLimit)
	require.True(t, msg.Bounce)
	require.Contains(t, string(msg.Payload), call.Method)

	again, err := NewMessageFromContractCall(call, admission, params)
	require.NoError(t, err)
	require.Equal(t, msg.MessageID, again.MessageID)
	require.Equal(t, msg.PayloadHash, again.PayloadHash)

	aether, err := NewAetherMessageFromContractCall(call, admission, params)
	require.NoError(t, err)
	require.Equal(t, hex.EncodeToString(call.Caller), aether.Sender)
	require.Equal(t, hex.EncodeToString(call.ContractAddr), aether.Receiver)
	require.Equal(t, zonestypes.ZoneIDContract, aether.SenderZoneID)
	require.Equal(t, zonestypes.ZoneIDContract, aether.ReceiverZoneID)
	require.Equal(t, call.Funds, aether.ValueNAET)
	require.Equal(t, ContractCallPayloadType, aether.PayloadType)
	require.Equal(t, OrderingClassObjectOrdered, aether.OrderingClass)
	require.Equal(t, ComputeContractCallRouteCommitment(call, admission), aether.RouteCommitment)

	escrow, err := NewAetherValueEscrow(AetherValueEscrow{
		MsgID:		aether.MsgID,
		ValueLocked:	aether.ValueNAET,
		FeeLocked:	aether.ForwardingFee,
		Status:		AetherEscrowLocked,
	})
	require.NoError(t, err)
	receipt, err := AetherReceiptFromMessage(aether, ReceiptStatusAccepted, admission.CreatedHeight, 0, sdkmath.ZeroInt(), nil, "", EmptyHash(), hashParts("contract-call-writes", aether.MsgID))
	require.NoError(t, err)
	state, err := NewAetherMsgBusState([]AetherMessage{aether}, nil, []AetherMessageReceipt{receipt}, []AetherValueEscrow{escrow}, nil)
	require.NoError(t, err)
	require.NoError(t, state.Validate())
}

func TestMessageKeeperSubmitsContractCallToOutbox(t *testing.T) {
	params := testContractCallParams()
	keeper, err := NewMessageKeeper(params)
	require.NoError(t, err)
	call := testMessageLayerContractCall()
	admission := testMessageLayerContractAdmission(call)

	keeper, response, err := keeper.SubmitContractCall(call, admission)
	require.NoError(t, err)
	require.Len(t, response.MessageID, MessageIDBytes)
	require.Len(t, keeper.State().Outbox, 1)
	queued := keeper.State().Outbox[0].Message
	require.Equal(t, response.MessageID, queued.MessageID)
	require.Equal(t, ContractCallPayloadType, queued.Opcode)
	require.Equal(t, ContractCallRouteID(call), queued.RouteID)
}

func TestMsgContractCallRejectsSectionTwelveFourValidationFailures(t *testing.T) {
	params := testContractCallParams()
	call := testMessageLayerContractCall()
	admission := testMessageLayerContractAdmission(call)

	missing := admission
	missing.ContractExists = false
	require.ErrorContains(t, call.Validate(missing, params), "does not exist")

	disabled := admission
	disabled.ContractEnabled = false
	require.ErrorContains(t, call.Validate(disabled, params), "disabled")

	badSelector := admission
	badSelector.EnabledMethods = append([]ContractCallMethod(nil), admission.EnabledMethods...)
	badSelector.EnabledMethods[0].SelectorHash = hashParts("wrong-selector")
	require.ErrorContains(t, call.Validate(badSelector, params), "selector")

	missingEscrow := admission
	missingEscrow.FundEscrows = nil
	require.ErrorContains(t, call.Validate(missingEscrow, params), "escrowed")

	overGas := call
	overGas.GasLimit = admission.MaxGasLimit + 1
	require.ErrorContains(t, overGas.Validate(admission, params), "gas limit")

	oversizedArgs := call
	oversizedArgs.Args = make([]byte, admission.MaxArgsBytes+1)
	require.ErrorContains(t, oversizedArgs.Validate(admission, params), "args size")
}

func testContractCallParams() MessageParams {
	params := testMessageParams()
	params.MaxPayloadSize = 1024
	params.MinGasLimit = 10
	params.MaxGasLimit = 10_000
	params.MinFeeLimit = sdkmath.NewInt(1)
	params.ParamsHash = EmptyHash()
	return params
}

func testMessageLayerContractCall() MsgContractCall {
	return MsgContractCall{
		Caller:			addr(20),
		ContractAddr:		addr(21),
		Method:			"counter.increment",
		Args:			[]byte("increment"),
		Funds:			sdkmath.NewInt(5),
		GasLimit:		500,
		ReplyToOptional:	addr(22),
		ExpiryHeight:		80,
	}
}

func testMessageLayerContractAdmission(call MsgContractCall) ContractCallAdmission {
	return ContractCallAdmission{
		CreatedHeight:		30,
		Nonce:			33,
		SourceSequence:		34,
		ContractExists:		true,
		ContractEnabled:	true,
		MaxArgsBytes:		128,
		MinGasLimit:		10,
		MaxGasLimit:		1_000,
		ContractShardID:	"contract-shard-1",
		ReplyShardID:		"contract-shard-2",
		EnabledMethods: []ContractCallMethod{{
			ContractAddr:	call.ContractAddr,
			Method:		call.Method,
			SelectorHash:	ComputeContractMethodSelectorHash(call.Method),
			Enabled:	true,
		}},
		FundEscrows: []ContractCallFundsEscrow{{
			Caller:		call.Caller,
			ContractAddr:	call.ContractAddr,
			Amount:		call.Funds,
			ExpiryHeight:	call.ExpiryHeight,
			Escrowed:	true,
		}},
	}
}
