package async

import (
	"errors"
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
		Params:			e.params,
		Contracts:		contracts,
		Queue:			cloneQueuedMessages(e.queue),
		Inbox:			cloneQueuedMap(e.inbox),
		Outbox:			cloneQueuedMap(e.outbox),
		DeadLetters:		cloneDeadLetters(e.deadLetters),
		Receipts:		cloneReceipts(e.receipts),
		NextSequence:		e.nextSequence,
		NextTxIndex:		e.nextTxIndex,
		NextDeadLetterSequence:	e.nextDeadLetterSequence,
		BlockHeight:		e.blockHeight,
		Metrics:		e.metrics,
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
	executor.deadLetters = cloneDeadLetters(exported.DeadLetters)
	executor.receipts = cloneReceipts(exported.Receipts)
	executor.nextSequence = exported.NextSequence
	executor.nextTxIndex = exported.NextTxIndex
	executor.nextDeadLetterSequence = exported.NextDeadLetterSequence
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
		if i > 0 && queuedMessageLess(queued, exported.Queue[i-1]) {
			return fmt.Errorf("queued messages must be sorted by scheduled/tx/message/logical/sequence/destination order")
		}
		if err := validateQueuedMessage(queued, exported.Params); err != nil {
			return fmt.Errorf("invalid queued message %d: %w", queued.Sequence, err)
		}
	}
	if err := validateQueuedMap("inbox", exported.Inbox, exported.Params, func(msg MessageEnvelope) string {
		return inboxKey(msg.Destination)
	}); err != nil {
		return err
	}
	if err := validateQueuedMap("outbox", exported.Outbox, exported.Params, func(msg MessageEnvelope) string {
		return outboxKey(msg.Source)
	}); err != nil {
		return err
	}
	if err := validateDeadLetters(exported.DeadLetters, exported.NextDeadLetterSequence, exported.Params); err != nil {
		return err
	}
	if err := validateReceipts(exported.Receipts); err != nil {
		return err
	}
	return nil
}

func validateQueuedMap(name string, view map[string][]QueuedMessage, params Params, ownerOf func(MessageEnvelope) string) error {
	owners := make([]string, 0, len(view))
	for owner := range view {
		owners = append(owners, owner)
	}
	sort.Strings(owners)
	for _, owner := range owners {
		if err := validateQueuedView(name, owner, view[owner], params, ownerOf); err != nil {
			return err
		}
	}
	return nil
}

func validateQueuedView(name, owner string, messages []QueuedMessage, params Params, ownerOf func(MessageEnvelope) string) error {
	if len(owner) == 0 {
		return fmt.Errorf("%s owner key must not be empty", name)
	}
	for i, queued := range messages {
		if ownerOf(queued.Envelope) != owner {
			return fmt.Errorf("%s message %d owner key drift: owner=%q actual=%q", name, queued.Sequence, owner, ownerOf(queued.Envelope))
		}
		if i > 0 && queuedMessageLess(queued, messages[i-1]) {
			return fmt.Errorf("%s messages must be sorted canonically", name)
		}
		if err := validateQueuedMessage(queued, params); err != nil {
			return fmt.Errorf("invalid %s message %d: %w", name, queued.Sequence, err)
		}
	}
	return nil
}

func validateDeadLetters(deadLetters []DeadLetter, nextSequence uint64, params Params) error {
	if len(deadLetters) > int(params.MaxDeadLetters) {
		return fmt.Errorf("dead letters must be <= %d", params.MaxDeadLetters)
	}
	seen := make(map[uint64]struct{}, len(deadLetters))
	for i, dead := range deadLetters {
		if _, exists := seen[dead.Sequence]; exists {
			return fmt.Errorf("duplicate dead letter sequence: %d", dead.Sequence)
		}
		seen[dead.Sequence] = struct{}{}
		if dead.Sequence >= nextSequence {
			return fmt.Errorf("dead letter sequence %d must be less than next_dead_letter_sequence %d", dead.Sequence, nextSequence)
		}
		if i > 0 && deadLetters[i-1].Sequence >= dead.Sequence {
			return errors.New("dead letters must be sorted canonically")
		}
		if dead.Envelope.ExecutionBlockHeight != 0 {
			return fmt.Errorf("dead letter %d execution block height must be zero", dead.Sequence)
		}
		if err := dead.Envelope.Validate(params); err != nil {
			return fmt.Errorf("invalid dead letter %d: %w", dead.Sequence, err)
		}
		if dead.Receipt.Sequence != dead.FailedSequence {
			return fmt.Errorf("dead letter %d receipt sequence drift", dead.Sequence)
		}
		if dead.RecordedBlock == 0 {
			return fmt.Errorf("dead letter %d recorded block must be positive", dead.Sequence)
		}
		if dead.Reason == "" {
			return fmt.Errorf("dead letter %d reason is required", dead.Sequence)
		}
	}
	return nil
}
