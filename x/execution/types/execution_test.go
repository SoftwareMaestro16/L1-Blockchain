package types

import (
	"bytes"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	appparams "github.com/sovereign-l1/l1/app/params"
	"github.com/sovereign-l1/l1/x/aetravm/async"
	identitytypes "github.com/sovereign-l1/l1/x/identity/types"
	memotypes "github.com/sovereign-l1/l1/x/memo/types"
	reputationtypes "github.com/sovereign-l1/l1/x/reputation/types"
)

func TestCheckTxPipelineValidatesFeesMemoAndTrace(t *testing.T) {
	envelope := testEnvelope()
	result, err := CheckTx(envelope, DefaultPipelineParams())
	require.NoError(t, err)
	require.Equal(t, sdk.NewInt64Coin(appparams.BaseDenom, 2), result.FeeQuote.RequiredFee)
	require.Equal(t, appparams.BaseDenom, result.MemoFee.Denom)
	require.NotEmpty(t, result.Trace.Steps)
	require.Equal(t, StageCheckTxDecode, result.Trace.Steps[0].Stage)
	require.Equal(t, StageCheckTxStateless, result.Trace.Steps[len(result.Trace.Steps)-1].Stage)
}

func TestCheckTxRejectsInvalidMemoAndFee(t *testing.T) {
	envelope := testEnvelope()
	envelope.Memo.Memo = string([]byte{0xff})
	_, err := CheckTx(envelope, DefaultPipelineParams())
	require.ErrorContains(t, err, "UTF-8")

	envelope = testEnvelope()
	envelope.Fee = sdk.NewCoins(sdk.NewInt64Coin("testtoken", 1))
	_, err = CheckTx(envelope, DefaultPipelineParams())
	require.ErrorContains(t, err, "naet")
}

func TestDeliverTxIntegratesResolverReputationAsyncAndEvents(t *testing.T) {
	params := DefaultPipelineParams()
	envelope := testEnvelope()
	envelope.Route = RouteResolverPayment
	envelope.ResolverDomain = "alice.aet"
	envelope.TimestampUnix = 100
	domain := identitytypes.DomainRecord{
		Name:		"alice",
		TLD:		identitytypes.DomainTLD,
		Owner:		addr(8),
		ExpiryUnix:	1_000,
		NFTItemID:	identitytypes.DomainNFTItemID("alice"),
		Status:		identitytypes.DomainStatusActive,
		CreatedAtUnix:	1,
		UpdatedAtUnix:	2,
	}
	resolver := identitytypes.ResolverRecord{
		Domain:		"alice.aet",
		Owner:		addr(8),
		Primary:	addr(9),
		Records:	map[string]sdk.AccAddress{identitytypes.ResolverKeyWallet: addr(9)},
		UpdatedAtUnix:	3,
	}
	envelope.DomainRecord = &domain
	envelope.ResolverRecord = &resolver
	envelope.AsyncMessages = []async.MessageEnvelope{testAsyncMessage(envelope.Sender, addr(10))}

	result, err := DeliverTx(envelope, params)
	require.NoError(t, err)
	require.Equal(t, addr(9), result.ResolvedTarget)
	require.True(t, result.AsyncQueued)
	require.True(t, result.StateWrite)
	require.Contains(t, result.Events, memotypes.EventTypeMemoAttached)
	require.Contains(t, result.Events, identitytypes.ResolverEventSet)
	require.Equal(t, StageStateWrite, result.Trace.Steps[len(result.Trace.Steps)-1].Stage)
}

func TestDeliverTxRejectsLowReputationLimitsAndInvalidVMRoute(t *testing.T) {
	envelope := testEnvelope()
	identity := reputationtypes.NewIdentityReputation("AE" + envelope.Sender.String())
	identity.Score = 100
	envelope.Identity = identity
	envelope.SenderTxCount = 1
	_, err := DeliverTx(envelope, DefaultPipelineParams())
	require.ErrorContains(t, err, "tx rate limit")

	envelope = testEnvelope()
	envelope.Route = RouteContractCall
	envelope.VMRoute = "unknown"
	_, err = DeliverTx(envelope, DefaultPipelineParams())
	require.ErrorContains(t, err, "invalid VM route")
}

func TestDeterministicEventsSorted(t *testing.T) {
	envelope := testEnvelope()
	envelope.ModuleEvents = []string{"z", "a"}
	envelope.ResolverDomain = "alice.aet"
	events := DeterministicEvents(envelope)
	require.Equal(t, []string{"a", memotypes.EventTypeMemoAttached, identitytypes.ResolverEventSet, "z"}, events)
}

func testEnvelope() ExecutionEnvelope {
	identity := reputationtypes.NewIdentityReputation("AE" + addr(1).String())
	identity.Score = 2000
	identity.Confidence = 1000
	return ExecutionEnvelope{
		TxHash:		[]byte("tx"),
		Sender:		addr(1),
		Receiver:	addr(2),
		Route:		RouteBankTransfer,
		GasLimit:	200_000,
		Fee:		sdk.NewCoins(sdk.NewInt64Coin(appparams.BaseDenom, 10)),
		Memo:		memotypes.TxMetadata{Memo: "note", MemoVisible: true},
		Identity:	identity,
		SenderStake:	sdkmath.NewInt(1_000_000_000),
	}
}

func testAsyncMessage(source, destination sdk.AccAddress) async.MessageEnvelope {
	return async.MessageEnvelope{
		Source:		source,
		Destination:	destination,
		Value:		sdk.NewCoin(appparams.BaseDenom, sdkmath.ZeroInt()),
		Opcode:		1,
		QueryID:	1,
		Body:		[]byte("body"),
		Bounce:		true,
		GasLimit:	100_000,
		ForwardFee:	sdk.NewCoin(appparams.BaseDenom, async.DefaultParams().ForwardingFee),
	}
}

func addr(seed byte) sdk.AccAddress {
	return sdk.AccAddress(bytes.Repeat([]byte{seed}, 20))
}
