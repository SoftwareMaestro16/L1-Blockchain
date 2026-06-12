package types

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/x/aetravm/async"
)

const (
	StatusActive	= "active"
	StatusPaused	= "paused"
	StatusDeleted	= "deleted"
)

type Actor struct {
	Address		sdk.AccAddress
	CodeHash	[]byte
	StateRoot	[]byte
	LogicalTime	uint64
	MailboxStats	MailboxStats
	Status		string
}

type MailboxStats struct {
	Pending		uint64
	Processed	uint64
	Failed		uint64
}

type ActorState struct {
	Actor	Actor
	Mailbox	[]async.QueuedMessage
}

type StateTransition struct {
	ActorAddress	sdk.AccAddress
	PreviousRoot	[]byte
	NextRoot	[]byte
	Message		async.MessageEnvelope
}

type ExportedActors struct {
	Actors []ActorState
}

func NewActor(address sdk.AccAddress, codeHash []byte, state []byte) (Actor, error) {
	actor := Actor{
		Address:	append(sdk.AccAddress(nil), address...),
		CodeHash:	append([]byte(nil), codeHash...),
		StateRoot:	StateRoot(state),
		Status:		StatusActive,
	}
	if err := actor.Validate(); err != nil {
		return Actor{}, err
	}
	return actor, nil
}

func (a Actor) Validate() error {
	if err := addressing.RejectZeroAddress("actor address", a.Address); err != nil {
		return err
	}
	if len(a.CodeHash) != async.CodeHashLength {
		return fmt.Errorf("actor code_hash must be %d bytes", async.CodeHashLength)
	}
	if len(a.StateRoot) != 32 {
		return errors.New("actor state_root must be 32 bytes")
	}
	if !IsActorStatus(a.Status) {
		return fmt.Errorf("invalid actor status %q", a.Status)
	}
	return nil
}

func ApplyMessage(actor Actor, msg async.MessageEnvelope, nextState []byte) (Actor, StateTransition, error) {
	if err := actor.Validate(); err != nil {
		return Actor{}, StateTransition{}, err
	}
	if actor.Status != StatusActive {
		return Actor{}, StateTransition{}, errors.New("actor is not active")
	}
	if !actor.Address.Equals(msg.Destination) {
		return Actor{}, StateTransition{}, errors.New("message destination must match actor")
	}
	next := actor
	next.LogicalTime++
	next.StateRoot = StateRoot(nextState)
	next.MailboxStats.Processed++
	transition := StateTransition{
		ActorAddress:	append(sdk.AccAddress(nil), actor.Address...),
		PreviousRoot:	append([]byte(nil), actor.StateRoot...),
		NextRoot:	append([]byte(nil), next.StateRoot...),
		Message:	cloneMessage(msg),
	}
	return next, transition, nil
}

func ValidateCrossActorIsolation(transition StateTransition, actor Actor) error {
	if !transition.ActorAddress.Equals(actor.Address) {
		return errors.New("actor cannot mutate another actor state directly")
	}
	return nil
}

func ExportActorState(states []ActorState) (ExportedActors, error) {
	out := make([]ActorState, len(states))
	seen := make(map[string]struct{}, len(states))
	for i, state := range states {
		if err := state.Actor.Validate(); err != nil {
			return ExportedActors{}, err
		}
		key := string(state.Actor.Address)
		if _, ok := seen[key]; ok {
			return ExportedActors{}, errors.New("duplicate actor address")
		}
		seen[key] = struct{}{}
		out[i] = ActorState{
			Actor:		cloneActor(state.Actor),
			Mailbox:	cloneQueued(state.Mailbox),
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return string(out[i].Actor.Address) < string(out[j].Actor.Address)
	})
	return ExportedActors{Actors: out}, nil
}

func StateRoot(state []byte) []byte {
	sum := sha256.Sum256(state)
	return sum[:]
}

func IsActorStatus(status string) bool {
	switch status {
	case StatusActive, StatusPaused, StatusDeleted:
		return true
	default:
		return false
	}
}

func cloneActor(actor Actor) Actor {
	actor.Address = append(sdk.AccAddress(nil), actor.Address...)
	actor.CodeHash = append([]byte(nil), actor.CodeHash...)
	actor.StateRoot = append([]byte(nil), actor.StateRoot...)
	return actor
}

func cloneQueued(messages []async.QueuedMessage) []async.QueuedMessage {
	out := make([]async.QueuedMessage, len(messages))
	copy(out, messages)
	return out
}

func cloneMessage(msg async.MessageEnvelope) async.MessageEnvelope {
	msg.Source = append(sdk.AccAddress(nil), msg.Source...)
	msg.Destination = append(sdk.AccAddress(nil), msg.Destination...)
	msg.Body = append([]byte(nil), msg.Body...)
	return msg
}
