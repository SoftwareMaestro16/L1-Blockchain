package types

import (
	"bytes"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	appparams "github.com/sovereign-l1/l1/app/params"
	"github.com/sovereign-l1/l1/x/aetravm/async"
)

func TestActorLifecycleAndOneTransitionPerMessage(t *testing.T) {
	actor, err := NewActor(addr(1), codeHash(1), []byte("state1"))
	require.NoError(t, err)
	msg := testMessage(addr(9), actor.Address)
	next, transition, err := ApplyMessage(actor, msg, []byte("state2"))
	require.NoError(t, err)
	require.Equal(t, uint64(1), next.LogicalTime)
	require.Equal(t, uint64(1), next.MailboxStats.Processed)
	require.Equal(t, actor.StateRoot, transition.PreviousRoot)
	require.Equal(t, next.StateRoot, transition.NextRoot)
	require.NoError(t, ValidateCrossActorIsolation(transition, next))
}

func TestActorCannotMutateAnotherActorDirectly(t *testing.T) {
	actor, err := NewActor(addr(1), codeHash(1), []byte("a"))
	require.NoError(t, err)
	other, err := NewActor(addr(2), codeHash(2), []byte("b"))
	require.NoError(t, err)
	_, transition, err := ApplyMessage(actor, testMessage(addr(9), actor.Address), []byte("next"))
	require.NoError(t, err)
	require.ErrorContains(t, ValidateCrossActorIsolation(transition, other), "another actor")
}

func TestActorMailboxExport(t *testing.T) {
	actorB, err := NewActor(addr(2), codeHash(2), []byte("b"))
	require.NoError(t, err)
	actorA, err := NewActor(addr(1), codeHash(1), []byte("a"))
	require.NoError(t, err)
	exported, err := ExportActorState([]ActorState{
		{Actor: actorB, Mailbox: []async.QueuedMessage{{Sequence: 2}}},
		{Actor: actorA, Mailbox: []async.QueuedMessage{{Sequence: 1}}},
	})
	require.NoError(t, err)
	require.Len(t, exported.Actors, 2)
	require.Equal(t, actorA.Address, exported.Actors[0].Actor.Address)
	require.Equal(t, uint64(1), exported.Actors[0].Mailbox[0].Sequence)

	_, err = ExportActorState([]ActorState{{Actor: actorA}, {Actor: actorA}})
	require.ErrorContains(t, err, "duplicate actor")
}

func TestActorValidation(t *testing.T) {
	_, err := NewActor(make([]byte, 20), codeHash(1), nil)
	require.ErrorContains(t, err, "actor address")
	_, err = NewActor(addr(1), []byte{1}, nil)
	require.ErrorContains(t, err, "code_hash")
	actor, err := NewActor(addr(1), codeHash(1), nil)
	require.NoError(t, err)
	actor.Status = "bad"
	require.ErrorContains(t, actor.Validate(), "status")
}

func testMessage(source, destination sdk.AccAddress) async.MessageEnvelope {
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

func codeHash(seed byte) []byte {
	return bytes.Repeat([]byte{seed}, async.CodeHashLength)
}
