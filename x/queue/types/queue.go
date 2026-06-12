package types

import (
	"errors"
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/x/reputation/types"
)

const (
	ClassElite	= uint8(0)
	ClassTrusted	= uint8(1)
	ClassNormal	= uint8(2)
	ClassNew	= uint8(3)
	ClassRestricted	= uint8(4)
)

type QueueParams struct {
	MaxPerBlock		uint32
	MaxPerAccountQueued	uint32
	MaxPerContractQueued	uint32
	StarvationWindowHeights	uint64
}

type QueueItem struct {
	ScheduledHeight		uint64
	ReputationClass		uint8
	TxHeight		uint64
	TxIndex			uint32
	MessageIndex		uint32
	SourceLogicalTime	uint64
	Sequence		uint64
	Account			sdk.AccAddress
	Contract		sdk.AccAddress
	Payload			[]byte
	Attempts		uint32
	LastError		string
}

type Queue struct {
	params		QueueParams
	items		[]QueueItem
	nextSequence	uint64
	accountCounts	map[string]uint32
	contractCounts	map[string]uint32
	processed	uint64
	failed		uint64
}

type Observability struct {
	Queued			uint64
	Processed		uint64
	Failed			uint64
	Lag			uint64
	AccountsTracked		uint64
	ContractsTracked	uint64
}

func DefaultParams() QueueParams {
	return QueueParams{
		MaxPerBlock:			128,
		MaxPerAccountQueued:		64,
		MaxPerContractQueued:		128,
		StarvationWindowHeights:	100,
	}
}

func NewQueue(params QueueParams) (*Queue, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}
	return &Queue{
		params:		params,
		accountCounts:	make(map[string]uint32),
		contractCounts:	make(map[string]uint32),
	}, nil
}

func (p QueueParams) Validate() error {
	if p.MaxPerBlock == 0 {
		return errors.New("queue max per block must be positive")
	}
	if p.MaxPerAccountQueued == 0 {
		return errors.New("queue max per account must be positive")
	}
	if p.MaxPerContractQueued == 0 {
		return errors.New("queue max per contract must be positive")
	}
	if p.StarvationWindowHeights == 0 {
		return errors.New("queue starvation window must be positive")
	}
	return nil
}

func (q *Queue) Enqueue(item QueueItem) (QueueItem, error) {
	if err := item.Validate(); err != nil {
		return QueueItem{}, err
	}
	accountKey := string(item.Account)
	contractKey := string(item.Contract)
	if q.accountCounts[accountKey] >= q.params.MaxPerAccountQueued {
		return QueueItem{}, fmt.Errorf("account queued message limit %d reached", q.params.MaxPerAccountQueued)
	}
	if q.contractCounts[contractKey] >= q.params.MaxPerContractQueued {
		return QueueItem{}, fmt.Errorf("contract queued message limit %d reached", q.params.MaxPerContractQueued)
	}
	item.Sequence = q.nextSequence
	q.nextSequence++
	q.items = append(q.items, cloneItem(item))
	q.accountCounts[accountKey]++
	q.contractCounts[contractKey]++
	q.sort(item.ScheduledHeight)
	return cloneItem(item), nil
}

func (q *Queue) PopReady(height uint64) []QueueItem {
	q.sort(height)
	limit := q.params.MaxPerBlock
	out := make([]QueueItem, 0, limit)
	remaining := q.items[:0]
	for _, item := range q.items {
		if item.ScheduledHeight <= height && uint32(len(out)) < limit {
			out = append(out, cloneItem(item))
			q.decrement(item)
			q.processed++
			continue
		}
		remaining = append(remaining, item)
	}
	q.items = remaining
	return out
}

func (q *Queue) Retry(item QueueItem, nextHeight uint64, reason string) (QueueItem, error) {
	item.ScheduledHeight = nextHeight
	item.Attempts++
	item.LastError = reason
	return q.Enqueue(item)
}

func (q *Queue) Fail(item QueueItem, reason string) QueueItem {
	item.LastError = reason
	item.Attempts++
	q.failed++
	return cloneItem(item)
}

func (q *Queue) Items() []QueueItem {
	out := make([]QueueItem, len(q.items))
	for i, item := range q.items {
		out[i] = cloneItem(item)
	}
	return out
}

func (q *Queue) Metrics(height uint64) Observability {
	lag := uint64(0)
	for _, item := range q.items {
		if item.ScheduledHeight <= height {
			lag++
		}
	}
	return Observability{
		Queued:			uint64(len(q.items)),
		Processed:		q.processed,
		Failed:			q.failed,
		Lag:			lag,
		AccountsTracked:	uint64(len(q.accountCounts)),
		ContractsTracked:	uint64(len(q.contractCounts)),
	}
}

func (q *Queue) sort(height uint64) {
	sort.SliceStable(q.items, func(i, j int) bool {
		return Less(q.items[i], q.items[j], height, q.params)
	})
}

func (q *Queue) decrement(item QueueItem) {
	if key := string(item.Account); q.accountCounts[key] > 0 {
		q.accountCounts[key]--
	}
	if key := string(item.Contract); q.contractCounts[key] > 0 {
		q.contractCounts[key]--
	}
}

func Less(left, right QueueItem, height uint64, params QueueParams) bool {
	leftKey := PriorityKey(left, height, params)
	rightKey := PriorityKey(right, height, params)
	return leftKey.Less(rightKey)
}

type PriorityKeyValue struct {
	ScheduledHeight		uint64
	ReputationClass		uint8
	TxHeight		uint64
	TxIndex			uint32
	MessageIndex		uint32
	SourceLogicalTime	uint64
	Sequence		uint64
}

func PriorityKey(item QueueItem, height uint64, params QueueParams) PriorityKeyValue {
	return PriorityKeyValue{
		ScheduledHeight:	item.ScheduledHeight,
		ReputationClass:	EffectiveReputationClass(item, height, params),
		TxHeight:		item.TxHeight,
		TxIndex:		item.TxIndex,
		MessageIndex:		item.MessageIndex,
		SourceLogicalTime:	item.SourceLogicalTime,
		Sequence:		item.Sequence,
	}
}

func (p PriorityKeyValue) Less(other PriorityKeyValue) bool {
	if p.ScheduledHeight != other.ScheduledHeight {
		return p.ScheduledHeight < other.ScheduledHeight
	}
	if p.ReputationClass != other.ReputationClass {
		return p.ReputationClass < other.ReputationClass
	}
	if p.TxHeight != other.TxHeight {
		return p.TxHeight < other.TxHeight
	}
	if p.TxIndex != other.TxIndex {
		return p.TxIndex < other.TxIndex
	}
	if p.MessageIndex != other.MessageIndex {
		return p.MessageIndex < other.MessageIndex
	}
	if p.SourceLogicalTime != other.SourceLogicalTime {
		return p.SourceLogicalTime < other.SourceLogicalTime
	}
	return p.Sequence < other.Sequence
}

func EffectiveReputationClass(item QueueItem, height uint64, params QueueParams) uint8 {
	if height >= item.ScheduledHeight && height-item.ScheduledHeight >= params.StarvationWindowHeights {
		return ClassElite
	}
	return item.ReputationClass
}

func ReputationClassForScore(score uint8) uint8 {
	switch types.LevelForScore(score) {
	case types.LevelElite:
		return ClassElite
	case types.LevelTrusted:
		return ClassTrusted
	case types.LevelNormal:
		return ClassNormal
	case types.LevelNew:
		return ClassNew
	default:
		return ClassRestricted
	}
}

func (i QueueItem) Validate() error {
	if err := addressing.RejectZeroAddress("queue account", i.Account); err != nil {
		return err
	}
	if err := addressing.RejectZeroAddress("queue contract", i.Contract); err != nil {
		return err
	}
	if i.ReputationClass > ClassRestricted {
		return fmt.Errorf("invalid reputation class %d", i.ReputationClass)
	}
	return nil
}

func cloneItem(item QueueItem) QueueItem {
	item.Account = append(sdk.AccAddress(nil), item.Account...)
	item.Contract = append(sdk.AccAddress(nil), item.Contract...)
	item.Payload = append([]byte(nil), item.Payload...)
	return item
}
