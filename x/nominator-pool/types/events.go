package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type StakingEventType string

const (
	EventAccountActivated		StakingEventType	= "AccountActivated"
	EventPoolStakeDeposited		StakingEventType	= "PoolStakeDeposited"
	EventPoolSharesMinted		StakingEventType	= "PoolSharesMinted"
	EventPoolAllocationUpdated	StakingEventType	= "PoolAllocationUpdated"
	EventPoolUnbondingRequested	StakingEventType	= "PoolUnbondingRequested"
	EventPoolUnbondingCompleted	StakingEventType	= "PoolUnbondingCompleted"
	EventPoolRewardsClaimed		StakingEventType	= "PoolRewardsClaimed"
	EventStakeReputationClaimed	StakingEventType	= "StakeReputationClaimed"
	EventValidatorRegistered	StakingEventType	= "ValidatorRegistered"
	EventValidatorUpdated		StakingEventType	= "ValidatorUpdated"
	EventAdvancedStakeDelegated	StakingEventType	= "AdvancedStakeDelegated"
	EventAdvancedStakeUndelegated	StakingEventType	= "AdvancedStakeUndelegated"
	EventAdvancedStakeRedelegated	StakingEventType	= "AdvancedStakeRedelegated"
)

type StakingEvent struct {
	Type			StakingEventType
	Actor			string
	PoolContract		string
	Validator		string
	Amount			uint64
	Shares			uint64
	Height			uint64
	Epoch			uint64
	Sequence		uint64
	StateKey		string
	ProofMetadataHash	string
	EventHash		string
}

type StakingReceipt struct {
	TxID		string
	Height		uint64
	Events		[]StakingEvent
	ReceiptHash	string
}

type StakingEventAttribute struct {
	Key	string
	Value	string
}

func NewStakingEvent(event StakingEvent) (StakingEvent, error) {
	event = normalizeStakingEvent(event)
	event.EventHash = ""
	if err := event.ValidateFormat(); err != nil {
		return StakingEvent{}, err
	}
	event.EventHash = ComputeStakingEventHash(event)
	return event, event.Validate()
}

func NewStakingReceipt(txID string, height uint64, events []StakingEvent) (StakingReceipt, error) {
	receipt := StakingReceipt{
		TxID:	strings.ToLower(strings.TrimSpace(txID)),
		Height:	height,
	}
	if receipt.TxID == "" {
		return StakingReceipt{}, errors.New("staking receipt tx id is required")
	}
	if containsSecretMaterial(receipt.TxID) {
		return StakingReceipt{}, errors.New("staking receipt must not include secret material")
	}
	if receipt.Height == 0 {
		return StakingReceipt{}, errors.New("staking receipt height must be positive")
	}
	ordered, err := DeterministicStakingEventOrder(events)
	if err != nil {
		return StakingReceipt{}, err
	}
	for idx, event := range ordered {
		if event.Height != receipt.Height {
			return StakingReceipt{}, errors.New("staking receipt event height mismatch")
		}
		if event.Sequence != uint64(idx) {
			return StakingReceipt{}, errors.New("staking receipt event sequences must be contiguous")
		}
	}
	receipt.Events = ordered
	receipt.ReceiptHash = ComputeStakingReceiptHash(receipt)
	return receipt, receipt.Validate()
}

func DeterministicStakingEventOrder(events []StakingEvent) ([]StakingEvent, error) {
	ordered := append([]StakingEvent(nil), events...)
	for idx, event := range ordered {
		normalized, err := NewStakingEvent(event)
		if err != nil {
			return nil, fmt.Errorf("staking event %d: %w", idx, err)
		}
		ordered[idx] = normalized
	}
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Height != ordered[j].Height {
			return ordered[i].Height < ordered[j].Height
		}
		if ordered[i].Sequence != ordered[j].Sequence {
			return ordered[i].Sequence < ordered[j].Sequence
		}
		if ordered[i].Type != ordered[j].Type {
			return ordered[i].Type < ordered[j].Type
		}
		return ordered[i].EventHash < ordered[j].EventHash
	})
	return ordered, nil
}

func (event StakingEvent) ValidateFormat() error {
	event = normalizeStakingEvent(event)
	if !isStakingEventType(event.Type) {
		return fmt.Errorf("unsupported staking event type %q", event.Type)
	}
	if event.Height == 0 {
		return errors.New("staking event height must be positive")
	}
	if event.StateKey == "" {
		return errors.New("staking event state key is required")
	}
	if err := ValidateUserFacingAEAddress("staking event actor", event.Actor); err != nil {
		return err
	}
	if event.PoolContract != "" {
		if err := ValidateUserFacingAEAddress("staking event pool contract", event.PoolContract); err != nil {
			return err
		}
	}
	if event.Validator != "" {
		if err := ValidateUserFacingAEAddress("staking event validator", event.Validator); err != nil {
			return err
		}
	}
	if requiresPoolContract(event.Type) && event.PoolContract == "" {
		return errors.New("staking event pool contract is required")
	}
	if requiresValidator(event.Type) && event.Validator == "" {
		return errors.New("staking event validator is required")
	}
	if !allowsValidator(event.Type) && event.Validator != "" {
		return errors.New("staking event validator is only allowed for allocation, validator, or advanced staking paths")
	}
	if requiresAmount(event.Type) && event.Amount == 0 {
		return errors.New("staking event amount is required")
	}
	if requiresShares(event.Type) && event.Shares == 0 {
		return errors.New("staking event shares are required")
	}
	if containsSecretMaterial(event.Actor, event.PoolContract, event.Validator, event.StateKey, event.ProofMetadataHash, string(event.Type)) {
		return errors.New("staking event must not include secret material")
	}
	return nil
}

func (event StakingEvent) Validate() error {
	event = normalizeStakingEvent(event)
	if err := event.ValidateFormat(); err != nil {
		return err
	}
	if event.EventHash == "" {
		return errors.New("staking event hash is required")
	}
	if event.EventHash != ComputeStakingEventHash(event) {
		return errors.New("staking event hash mismatch")
	}
	return nil
}

func (receipt StakingReceipt) Validate() error {
	receipt.TxID = strings.ToLower(strings.TrimSpace(receipt.TxID))
	if receipt.TxID == "" {
		return errors.New("staking receipt tx id is required")
	}
	if receipt.Height == 0 {
		return errors.New("staking receipt height must be positive")
	}
	if len(receipt.Events) == 0 {
		return errors.New("staking receipt requires events")
	}
	for idx, event := range receipt.Events {
		if err := event.Validate(); err != nil {
			return err
		}
		if event.Height != receipt.Height || event.Sequence != uint64(idx) {
			return errors.New("staking receipt events are not deterministically ordered")
		}
	}
	if receipt.ReceiptHash == "" {
		return errors.New("staking receipt hash is required")
	}
	if receipt.ReceiptHash != ComputeStakingReceiptHash(receipt) {
		return errors.New("staking receipt hash mismatch")
	}
	return nil
}

func (event StakingEvent) OrderedAttributes() []StakingEventAttribute {
	event = normalizeStakingEvent(event)
	attrs := []StakingEventAttribute{
		{Key: "event_type", Value: string(event.Type)},
		{Key: "actor", Value: event.Actor},
	}
	if event.PoolContract != "" {
		attrs = append(attrs, StakingEventAttribute{Key: "pool_contract", Value: event.PoolContract})
	}
	if event.Validator != "" {
		attrs = append(attrs, StakingEventAttribute{Key: "validator", Value: event.Validator})
	}
	if event.Amount > 0 {
		attrs = append(attrs, StakingEventAttribute{Key: "amount", Value: fmt.Sprint(event.Amount)})
	}
	if event.Shares > 0 {
		attrs = append(attrs, StakingEventAttribute{Key: "shares", Value: fmt.Sprint(event.Shares)})
	}
	if event.Epoch > 0 {
		attrs = append(attrs, StakingEventAttribute{Key: "epoch", Value: fmt.Sprint(event.Epoch)})
	}
	attrs = append(attrs,
		StakingEventAttribute{Key: "height", Value: fmt.Sprint(event.Height)},
		StakingEventAttribute{Key: "sequence", Value: fmt.Sprint(event.Sequence)},
		StakingEventAttribute{Key: "state_key", Value: event.StateKey},
	)
	if event.ProofMetadataHash != "" {
		attrs = append(attrs, StakingEventAttribute{Key: "proof_metadata_hash", Value: event.ProofMetadataHash})
	}
	if event.EventHash != "" {
		attrs = append(attrs, StakingEventAttribute{Key: "event_hash", Value: event.EventHash})
	}
	return attrs
}

func AccountActivationEventStateKey(account string) string {
	return "accounts/" + strings.TrimSpace(account) + "/activation"
}

func AdvancedStakeEventStateKey(actor string, validator string) string {
	return "staking/advanced_delegation/" + strings.TrimSpace(actor) + "/" + strings.TrimSpace(validator)
}

func AdvancedStakeRedelegationEventStateKey(actor string, sourceValidator string, targetValidator string) string {
	return "staking/advanced_redelegation/" + strings.TrimSpace(actor) + "/" + strings.TrimSpace(sourceValidator) + "/" + strings.TrimSpace(targetValidator)
}

func ComputeStakingEventHash(event StakingEvent) string {
	event.EventHash = ""
	event = normalizeStakingEvent(event)
	return hashProofMetadataParts(
		"staking-event-v1",
		string(event.Type),
		event.Actor,
		event.PoolContract,
		event.Validator,
		fmt.Sprint(event.Amount),
		fmt.Sprint(event.Shares),
		fmt.Sprint(event.Height),
		fmt.Sprint(event.Epoch),
		fmt.Sprint(event.Sequence),
		event.StateKey,
		event.ProofMetadataHash,
	)
}

func ComputeStakingReceiptHash(receipt StakingReceipt) string {
	parts := []string{
		"staking-receipt-v1",
		strings.ToLower(strings.TrimSpace(receipt.TxID)),
		fmt.Sprint(receipt.Height),
	}
	for _, event := range receipt.Events {
		parts = append(parts, event.EventHash)
	}
	return hashProofMetadataParts(parts...)
}

func isStakingEventType(eventType StakingEventType) bool {
	switch eventType {
	case EventAccountActivated,
		EventPoolStakeDeposited,
		EventPoolSharesMinted,
		EventPoolAllocationUpdated,
		EventPoolUnbondingRequested,
		EventPoolUnbondingCompleted,
		EventPoolRewardsClaimed,
		EventStakeReputationClaimed,
		EventValidatorRegistered,
		EventValidatorUpdated,
		EventAdvancedStakeDelegated,
		EventAdvancedStakeUndelegated,
		EventAdvancedStakeRedelegated:
		return true
	default:
		return false
	}
}

func requiresPoolContract(eventType StakingEventType) bool {
	switch eventType {
	case EventPoolStakeDeposited,
		EventPoolSharesMinted,
		EventPoolAllocationUpdated,
		EventPoolUnbondingRequested,
		EventPoolUnbondingCompleted,
		EventPoolRewardsClaimed,
		EventStakeReputationClaimed:
		return true
	default:
		return false
	}
}

func requiresValidator(eventType StakingEventType) bool {
	switch eventType {
	case EventPoolAllocationUpdated,
		EventValidatorRegistered,
		EventValidatorUpdated,
		EventAdvancedStakeDelegated,
		EventAdvancedStakeUndelegated,
		EventAdvancedStakeRedelegated:
		return true
	default:
		return false
	}
}

func allowsValidator(eventType StakingEventType) bool {
	return requiresValidator(eventType)
}

func requiresAmount(eventType StakingEventType) bool {
	switch eventType {
	case EventPoolStakeDeposited,
		EventPoolAllocationUpdated,
		EventPoolUnbondingCompleted,
		EventPoolRewardsClaimed,
		EventStakeReputationClaimed,
		EventValidatorRegistered,
		EventAdvancedStakeDelegated,
		EventAdvancedStakeUndelegated,
		EventAdvancedStakeRedelegated:
		return true
	default:
		return false
	}
}

func requiresShares(eventType StakingEventType) bool {
	switch eventType {
	case EventPoolStakeDeposited,
		EventPoolSharesMinted,
		EventPoolUnbondingRequested:
		return true
	default:
		return false
	}
}

func normalizeStakingEvent(event StakingEvent) StakingEvent {
	event.Type = StakingEventType(strings.TrimSpace(string(event.Type)))
	event.Actor = strings.TrimSpace(event.Actor)
	event.PoolContract = strings.TrimSpace(event.PoolContract)
	event.Validator = strings.TrimSpace(event.Validator)
	event.StateKey = strings.TrimSpace(event.StateKey)
	event.ProofMetadataHash = strings.ToLower(strings.TrimSpace(event.ProofMetadataHash))
	event.EventHash = strings.ToLower(strings.TrimSpace(event.EventHash))
	return event
}

func containsSecretMaterial(values ...string) bool {
	for _, value := range values {
		lower := strings.ToLower(value)
		for _, marker := range []string{"private_key", "private key", "seed_phrase", "seed phrase", "mnemonic", "secret"} {
			if strings.Contains(lower, marker) {
				return true
			}
		}
	}
	return false
}
