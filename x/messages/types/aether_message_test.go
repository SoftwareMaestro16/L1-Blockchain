package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestAetherMessageDerivesCanonicalIDAndTrace(t *testing.T) {
	route := testAetherRoute(t, 70, 2)
	msg := testAetherMessage(t, route)

	require.NoError(t, msg.Validate())
	require.NoError(t, zonestypes.ValidateHash("aether msg id", msg.MsgID))
	require.NoError(t, zonestypes.ValidateHash("aether trace id", msg.TraceID))
	require.Equal(t, ComputeAetherMessageID(msg), msg.MsgID)

	encodedA, err := CanonicalAetherMessageBinary(msg)
	require.NoError(t, err)
	encodedB, err := CanonicalAetherMessageBinary(msg)
	require.NoError(t, err)
	require.Equal(t, encodedA, encodedB)

	changed := msg.Clone()
	changed.Payload = []byte("different-payload")
	changed.MsgID = ComputeAetherMessageID(changed)
	require.NotEqual(t, msg.MsgID, changed.MsgID)
}

func TestAetherMessageSupportsParentProofAndSignatureMetadata(t *testing.T) {
	route := testAetherRoute(t, 75, 1)
	parent := testAetherMessage(t, route)
	child, err := NewAetherMessage(AetherMessage{
		ParentMsgID:		parent.MsgID,
		TraceID:		parent.TraceID,
		Sender:			"contract/child",
		SenderZoneID:		zonestypes.ZoneIDContract,
		SenderShardID:		"contract-7",
		Receiver:		"module/payments",
		ReceiverZoneID:		zonestypes.ZoneIDFinancial,
		ReceiverShardID:	"financial-1",
		ValueNAET:		sdkmath.NewInt(5),
		Payload:		[]byte("settlement"),
		PayloadType:		"payment.settle",
		GasLimit:		200,
		GasPrice:		sdkmath.NewInt(2),
		ForwardingFee:		sdkmath.NewInt(3),
		ExpiryHeight:		150,
		Bounce:			true,
		ExecutionMode:		ExecutionModePromise,
		OrderingClass:		OrderingClassStrictTraceOrder,
		RouteCommitment:	route.RouteCommitment,
		AuthProof: AetherProof{
			ProofType:	"delegation",
			RootHash:	EmptyHash(),
			ProofHash:	testMessageHash("delegation-proof"),
		},
		StateProof: AetherProof{
			ProofType:	"payment_state",
			RootHash:	testMessageHash("payment-root"),
			ProofHash:	testMessageHash("payment-proof"),
		},
		CreatedAtHeight:	76,
		Nonce:			2,
		Signature: AetherSignature{
			Signer:		"service/router",
			SignatureHex:	"abcdef0123456789",
		},
	})
	require.NoError(t, err)
	require.Equal(t, parent.MsgID, child.ParentMsgID)
	require.Equal(t, parent.TraceID, child.TraceID)
	require.NoError(t, child.AuthProof.ValidateOptional("auth"))
	require.NoError(t, child.StateProof.ValidateOptional("state"))
	require.NoError(t, child.Signature.ValidateOptional())
}

func TestAetherMessageFromLegacyMessageUsesRouteMetadata(t *testing.T) {
	params := testMessageParams()
	legacy := testMessage(t, params, 44, 55, []byte("legacy"))
	route := testAetherRoute(t, 90, 3)

	aether, err := AetherMessageFromMessage(legacy, route, "", "")
	require.NoError(t, err)
	require.Equal(t, zonestypes.ZoneIDFinancial, aether.SenderZoneID)
	require.Equal(t, "financial-1", aether.SenderShardID)
	require.Equal(t, zonestypes.ZoneIDContract, aether.ReceiverZoneID)
	require.Equal(t, "contract-7", aether.ReceiverShardID)
	require.Equal(t, legacy.Value, aether.ValueNAET)
	require.Equal(t, legacy.Opcode, aether.PayloadType)
	require.Equal(t, route.RouteCommitment, aether.RouteCommitment)
	require.Equal(t, ExecutionModeAsync, aether.ExecutionMode)
	require.Equal(t, OrderingClassSenderOrdered, aether.OrderingClass)
	require.NoError(t, aether.Validate())
}

func TestAetherMessageRejectsMissingConsensusMetadata(t *testing.T) {
	route := testAetherRoute(t, 95, 1)
	msg := testAetherMessage(t, route)

	noRoute := msg.Clone()
	noRoute.RouteCommitment = ""
	require.ErrorContains(t, noRoute.Validate(), "route commitment")

	noNonce := msg.Clone()
	noNonce.Nonce = 0
	noNonce.MsgID = ComputeAetherMessageID(noNonce)
	require.ErrorContains(t, noNonce.Validate(), "nonce")

	badProof := msg.Clone()
	badProof.StateProof = AetherProof{ProofType: "state", RootHash: EmptyHash(), ProofHash: "bad"}
	badProof.MsgID = ComputeAetherMessageID(badProof)
	require.ErrorContains(t, badProof.Validate(), "proof")

	badSignature := msg.Clone()
	badSignature.Signature = AetherSignature{Signer: "service/router", SignatureHex: "not-hex"}
	badSignature.MsgID = ComputeAetherMessageID(badSignature)
	require.ErrorContains(t, badSignature.Validate(), "signature")
}

func testAetherMessage(t *testing.T, route UnifiedMessageRoute) AetherMessage {
	t.Helper()
	msg, err := NewAetherMessage(AetherMessage{
		Sender:			"account/alice",
		SenderZoneID:		zonestypes.ZoneIDFinancial,
		SenderShardID:		"financial-1",
		Receiver:		"contract/vault",
		ReceiverZoneID:		zonestypes.ZoneIDContract,
		ReceiverShardID:	"contract-7",
		ValueNAET:		sdkmath.NewInt(10),
		Payload:		[]byte("execute"),
		PayloadType:		"contract.execute",
		GasLimit:		100,
		GasPrice:		sdkmath.NewInt(2),
		ForwardingFee:		sdkmath.NewInt(3),
		ExpiryHeight:		120,
		Bounce:			true,
		ExecutionMode:		ExecutionModeAsync,
		OrderingClass:		OrderingClassSenderOrdered,
		RouteCommitment:	route.RouteCommitment,
		CreatedAtHeight:	70,
		Nonce:			1,
	})
	require.NoError(t, err)
	return msg
}

func testAetherRoute(t *testing.T, height uint64, finalityDelay uint64) UnifiedMessageRoute {
	t.Helper()
	route := UnifiedMessageRoute{
		SourceZoneID:		zonestypes.ZoneIDFinancial,
		SourceShardID:		"financial-1",
		DestinationZoneID:	zonestypes.ZoneIDContract,
		DestinationShardID:	"contract-7",
		ModuleRoute:		"contract/execute",
		OrderingClass:		OrderingClassSenderOrdered,
		ExecutionMode:		ExecutionModeAsync,
		CommittedHeight:	height,
		FinalityDelay:		finalityDelay,
	}
	route.DeliveryEligibleFrom = route.CommittedHeight + route.FinalityDelay
	route.RouteCommitment = ComputeUnifiedRouteCommitment(route)
	require.NoError(t, route.ValidateHash())
	return route
}

func testMessageHash(seed string) string {
	return hashParts("aether-message-test-hash", seed)
}
