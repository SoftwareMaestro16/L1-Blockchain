package async

import (
	"fmt"
	"sort"
)

func (e *Executor) ExportState() ExportedState {
	contracts := make([]ContractAccount, 0, len(e.contracts))
	for _, contract := range e.contracts {
		contracts = append(contracts, cloneContract(contract))
	}
	sort.Slice(contracts, func(i, j int) bool {
		return string(contracts[i].Address) < string(contracts[j].Address)
	})
	return ExportedState{
		Params:       e.params,
		Contracts:    contracts,
		Queue:        cloneQueuedMessages(e.queue),
		Inbox:        cloneQueuedMap(e.inbox),
		Outbox:       cloneQueuedMap(e.outbox),
		Receipts:     cloneReceipts(e.receipts),
		NextSequence: e.nextSequence,
		NextTxIndex:  e.nextTxIndex,
		BlockHeight:  e.blockHeight,
		Metrics:      e.metrics,
	}
}

func ImportState(exported ExportedState) (*Executor, error) {
	executor, err := NewExecutor(exported.Params)
	if err != nil {
		return nil, err
	}
	if err := ValidateExportedState(exported); err != nil {
		return nil, err
	}
	executor.queue = cloneQueuedMessages(exported.Queue)
	executor.inbox = cloneQueuedMap(exported.Inbox)
	executor.outbox = cloneQueuedMap(exported.Outbox)
	executor.receipts = cloneReceipts(exported.Receipts)
	executor.nextSequence = exported.NextSequence
	executor.nextTxIndex = exported.NextTxIndex
	executor.blockHeight = exported.BlockHeight
	executor.metrics = exported.Metrics
	for _, contract := range exported.Contracts {
		if err := contract.Validate(exported.Params); err != nil {
			return nil, err
		}
		executor.contracts[string(contract.Address)] = cloneContract(contract)
	}
	return executor, nil
}

func ValidateExportedState(exported ExportedState) error {
	if err := exported.Params.Validate(); err != nil {
		return err
	}
	seenContracts := make(map[string]struct{}, len(exported.Contracts))
	for _, contract := range exported.Contracts {
		if err := contract.Validate(exported.Params); err != nil {
			return err
		}
		key := string(contract.Address)
		if _, exists := seenContracts[key]; exists {
			return fmt.Errorf("duplicate contract address: %s", contract.Address.String())
		}
		seenContracts[key] = struct{}{}
	}
	seenSequences := make(map[uint64]struct{}, len(exported.Queue))
	for i, queued := range exported.Queue {
		if _, exists := seenSequences[queued.Sequence]; exists {
			return fmt.Errorf("duplicate queued message sequence: %d", queued.Sequence)
		}
		seenSequences[queued.Sequence] = struct{}{}
		if queued.Sequence >= exported.NextSequence {
			return fmt.Errorf("queued message sequence %d must be less than next_sequence %d", queued.Sequence, exported.NextSequence)
		}
		if queued.TxIndex >= exported.NextTxIndex {
			return fmt.Errorf("queued message tx_index %d must be less than next_tx_index %d", queued.TxIndex, exported.NextTxIndex)
		}
		if queued.SourceLogicalTime != queued.Envelope.CreatedLogicalTime {
			return fmt.Errorf("queued message %d source logical time drift", queued.Sequence)
		}
		if queued.DestinationKey != string(queued.Envelope.Destination) {
			return fmt.Errorf("queued message %d destination key drift", queued.Sequence)
		}
		if queued.Envelope.ExecutionBlockHeight != 0 {
			return fmt.Errorf("queued message %d execution block height must be zero", queued.Sequence)
		}
		if i > 0 && queuedMessageLess(queued, exported.Queue[i-1]) {
			return fmt.Errorf("queued messages must be sorted by tx/message/logical/destination/sequence order")
		}
		if err := queued.Envelope.Validate(exported.Params); err != nil {
			return fmt.Errorf("invalid queued message %d: %w", queued.Sequence, err)
		}
	}
	if err := validateQueuedMap("inbox", exported.Inbox, exported.Params); err != nil {
		return err
	}
	if err := validateQueuedMap("outbox", exported.Outbox, exported.Params); err != nil {
		return err
	}
	return nil
}

func validateQueuedMap(name string, view map[string][]QueuedMessage, params Params) error {
	owners := make([]string, 0, len(view))
	for owner := range view {
		owners = append(owners, owner)
	}
	sort.Strings(owners)
	for _, owner := range owners {
		if err := validateQueuedView(name, owner, view[owner], params); err != nil {
			return err
		}
	}
	return nil
}

func validateQueuedView(name, owner string, messages []QueuedMessage, params Params) error {
	if len(owner) == 0 {
		return fmt.Errorf("%s owner key must not be empty", name)
	}
	for _, queued := range messages {
		if queued.Envelope.ExecutionBlockHeight != 0 {
			return fmt.Errorf("%s message %d execution block height must be zero", name, queued.Sequence)
		}
		if err := queued.Envelope.Validate(params); err != nil {
			return fmt.Errorf("invalid %s message %d: %w", name, queued.Sequence, err)
		}
	}
	return nil
}
