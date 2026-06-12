package types

import (
	"errors"
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/app/addressing"
)

const (
	EventTransfer			= "EventTransfer"
	EventMemoAttached		= "EventMemoAttached"
	EventDomainAuctionStarted	= "EventDomainAuctionStarted"
	EventDomainResolved		= "EventDomainResolved"
	EventContractMessageQueued	= "EventContractMessageQueued"
	EventContractMessageProcessed	= "EventContractMessageProcessed"
	EventReputationUpdated		= "EventReputationUpdated"
	EventFeeDistributed		= "EventFeeDistributed"

	CategoryProtocol	= "protocol"
	CategoryIndexer		= "indexer"
	CategoryContract	= "contract"
	CategoryMemo		= "memo"
	CategoryDomain		= "domain"
	CategoryReputation	= "reputation"
)

type ProtocolEvent struct {
	Type		string
	Category	string
	TxHash		[]byte
	Height		uint64
	Sequence	uint64
	Actor		sdk.AccAddress
	Attributes	[]Attribute
}

type Attribute struct {
	Key	string
	Value	string
}

func NewEvent(eventType, category string, txHash []byte, height uint64, sequence uint64, actor sdk.AccAddress, attrs []Attribute) (ProtocolEvent, error) {
	event := ProtocolEvent{
		Type:		eventType,
		Category:	category,
		TxHash:		append([]byte(nil), txHash...),
		Height:		height,
		Sequence:	sequence,
		Actor:		append(sdk.AccAddress(nil), actor...),
		Attributes:	CanonicalAttributes(attrs),
	}
	if err := event.Validate(); err != nil {
		return ProtocolEvent{}, err
	}
	return event, nil
}

func (e ProtocolEvent) Validate() error {
	if !IsEventType(e.Type) {
		return fmt.Errorf("invalid event type %q", e.Type)
	}
	if !IsEventCategory(e.Category) {
		return fmt.Errorf("invalid event category %q", e.Category)
	}
	if len(e.Actor) > 0 {
		if err := addressing.RejectZeroAddress("event actor", e.Actor); err != nil {
			return err
		}
	}
	for _, attr := range e.Attributes {
		if attr.Key == "" {
			return errors.New("event attribute key is required")
		}
	}
	if !AttributesSorted(e.Attributes) {
		return errors.New("event attributes must be sorted by key")
	}
	return nil
}

func CanonicalAttributes(attrs []Attribute) []Attribute {
	out := append([]Attribute(nil), attrs...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Key == out[j].Key {
			return out[i].Value < out[j].Value
		}
		return out[i].Key < out[j].Key
	})
	return out
}

func AttributesSorted(attrs []Attribute) bool {
	for i := 1; i < len(attrs); i++ {
		if attrs[i-1].Key > attrs[i].Key {
			return false
		}
		if attrs[i-1].Key == attrs[i].Key && attrs[i-1].Value > attrs[i].Value {
			return false
		}
	}
	return true
}

func SortEvents(events []ProtocolEvent) []ProtocolEvent {
	out := append([]ProtocolEvent(nil), events...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Height != out[j].Height {
			return out[i].Height < out[j].Height
		}
		if out[i].Sequence != out[j].Sequence {
			return out[i].Sequence < out[j].Sequence
		}
		if out[i].Type != out[j].Type {
			return out[i].Type < out[j].Type
		}
		return string(out[i].TxHash) < string(out[j].TxHash)
	})
	return out
}

func IsEventType(eventType string) bool {
	switch eventType {
	case EventTransfer,
		EventMemoAttached,
		EventDomainAuctionStarted,
		EventDomainResolved,
		EventContractMessageQueued,
		EventContractMessageProcessed,
		EventReputationUpdated,
		EventFeeDistributed:
		return true
	default:
		return false
	}
}

func IsEventCategory(category string) bool {
	switch category {
	case CategoryProtocol, CategoryIndexer, CategoryContract, CategoryMemo, CategoryDomain, CategoryReputation:
		return true
	default:
		return false
	}
}
