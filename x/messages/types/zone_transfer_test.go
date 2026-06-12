package types

import (
	"encoding/hex"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestMsgZoneTransferBuildsCanonicalMessagesEscrowAndReceipt(t *testing.T) {
	params := testZoneTransferParams()
	transfer := testZoneTransfer()
	admission := testZoneTransferAdmission(transfer)

	msg, err := NewMessageFromZoneTransfer(transfer, admission, params)
	require.NoError(t, err)
	require.Equal(t, transfer.SourceZoneID, msg.SourceZone)
	require.Equal(t, transfer.DestinationZoneID, msg.DestinationZone)
	require.Equal(t, ZoneTransferPayloadType, msg.Opcode)
	require.Equal(t, transfer.Amount, msg.Value)
	require.Equal(t, transfer.ForwardingFee, msg.FeeLimit)
	require.Equal(t, ZoneTransferAuthScope, msg.AuthScope)
	require.True(t, msg.Bounce)
	require.Contains(t, string(msg.Payload), transfer.Denom)

	again, err := NewMessageFromZoneTransfer(transfer, admission, params)
	require.NoError(t, err)
	require.Equal(t, msg.MessageID, again.MessageID)
	require.Equal(t, msg.PayloadHash, again.PayloadHash)

	aether, err := NewAetherMessageFromZoneTransfer(transfer, admission, params)
	require.NoError(t, err)
	require.Equal(t, hex.EncodeToString(transfer.FromAddress), aether.Sender)
	require.Equal(t, hex.EncodeToString(transfer.ToAddress), aether.Receiver)
	require.Equal(t, transfer.SourceZoneID, aether.SenderZoneID)
	require.Equal(t, transfer.DestinationZoneID, aether.ReceiverZoneID)
	require.Equal(t, transfer.Amount, aether.ValueNAET)
	require.Equal(t, ComputeZoneTransferRouteCommitment(transfer, admission), aether.RouteCommitment)

	result, err := BuildZoneTransferAcceptedResult(transfer, admission, params, admission.CreatedHeight)
	require.NoError(t, err)
	require.Equal(t, msg.MessageID, result.Message.MessageID)
	require.Equal(t, aether.MsgID, result.AetherMessage.MsgID)
	require.Equal(t, aether.MsgID, result.Escrow.MsgID)
	require.Equal(t, AetherEscrowLocked, result.Escrow.Status)
	require.Equal(t, ReceiptStatusAccepted, result.Receipt.Status)

	state, err := NewAetherMsgBusState([]AetherMessage{result.AetherMessage}, nil, []AetherMessageReceipt{result.Receipt}, []AetherValueEscrow{result.Escrow}, nil)
	require.NoError(t, err)
	require.NoError(t, state.Validate())
	require.NotEmpty(t, state.MessageRoot)
	require.NotEmpty(t, state.ReceiptRoot)
}

func TestMessageKeeperSubmitsZoneTransferToOutbox(t *testing.T) {
	params := testZoneTransferParams()
	keeper, err := NewMessageKeeper(params)
	require.NoError(t, err)
	transfer := testZoneTransfer()
	admission := testZoneTransferAdmission(transfer)

	keeper, response, err := keeper.SubmitZoneTransfer(transfer, admission)
	require.NoError(t, err)
	require.Len(t, response.MessageID, MessageIDBytes)
	require.Len(t, keeper.State().Outbox, 1)
	queued := keeper.State().Outbox[0].Message
	require.Equal(t, response.MessageID, queued.MessageID)
	require.Equal(t, ZoneTransferPayloadType, queued.Opcode)
	require.Equal(t, ZoneTransferRouteID(transfer), queued.RouteID)
}

func TestMsgZoneTransferRejectsInvalidAdmissionAndRouting(t *testing.T) {
	params := testZoneTransferParams()
	transfer := testZoneTransfer()
	admission := testZoneTransferAdmission(transfer)

	insufficient := admission
	insufficient.SourceSpendable = sdkmath.NewInt(9)
	require.ErrorContains(t, transfer.Validate(insufficient, params), "source balance")

	badDenom := admission
	badDenom.RoutableDenoms = nil
	require.ErrorContains(t, transfer.Validate(badDenom, params), "routable denoms")

	notRoutable := admission
	notRoutable.RoutableDenoms = append([]RoutableDenom(nil), admission.RoutableDenoms...)
	notRoutable.RoutableDenoms[0].Denom = "factory/test/other"
	require.ErrorContains(t, transfer.Validate(notRoutable, params), "not routable")

	lowFee := transfer
	lowFee.ForwardingFee = sdkmath.ZeroInt()
	require.ErrorContains(t, lowFee.Validate(admission, params), "forwarding fee")

	disabled := admission
	disabled.EnabledZones = []zonestypes.ZoneID{transfer.SourceZoneID}
	require.ErrorContains(t, transfer.Validate(disabled, params), "destination zone")

	expired := transfer
	expired.ExpiryHeight = admission.CreatedHeight - 1
	require.ErrorContains(t, expired.Validate(admission, params), "expiry")

	notEscrowed := admission
	notEscrowed.SourceEscrowed = false
	require.ErrorContains(t, transfer.Validate(notEscrowed, params), "escrowed")
}

func testZoneTransferParams() MessageParams {
	params := testMessageParams()
	params.MaxPayloadSize = 512
	params.MaxGasLimit = 10_000
	params.MinFeeLimit = sdkmath.NewInt(1)
	params.ParamsHash = EmptyHash()
	return params
}

func testZoneTransfer() MsgZoneTransfer {
	return MsgZoneTransfer{
		FromAddress:		addr(10),
		ToAddress:		addr(11),
		SourceZoneID:		zonestypes.ZoneIDFinancial,
		DestinationZoneID:	zonestypes.ZoneIDContract,
		Amount:			sdkmath.NewInt(100),
		Denom:			"naet",
		GasLimit:		100,
		ForwardingFee:		sdkmath.NewInt(3),
		ExpiryHeight:		120,
		MemoHashOptional:	testMessageHash("zone-transfer-memo"),
	}
}

func testZoneTransferAdmission(transfer MsgZoneTransfer) ZoneTransferAdmission {
	return ZoneTransferAdmission{
		CreatedHeight:		20,
		Nonce:			77,
		SourceSequence:		78,
		SourceSpendable:	sdkmath.NewInt(1000),
		EnabledZones:		[]zonestypes.ZoneID{transfer.DestinationZoneID, transfer.SourceZoneID},
		SourceShardID:		"financial-shard-1",
		DestinationShardID:	"contract-shard-2",
		MaxDeliveryWindow:	200,
		SourceEscrowed:		true,
		RoutableDenoms: []RoutableDenom{{
			Denom:			transfer.Denom,
			SourceZoneID:		transfer.SourceZoneID,
			DestinationZoneID:	transfer.DestinationZoneID,
			AuthorityPath:		"financial/contract-assets",
		}},
		MinimumRouteFees: []ZoneTransferRouteFee{{
			SourceZoneID:		transfer.SourceZoneID,
			DestinationZoneID:	transfer.DestinationZoneID,
			Denom:			transfer.Denom,
			MinimumFee:		sdkmath.NewInt(2),
		}},
	}
}
