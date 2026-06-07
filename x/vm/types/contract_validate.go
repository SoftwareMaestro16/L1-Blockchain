package types

import (
	"bytes"
	"errors"
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/x/aetravm/async"
	"github.com/sovereign-l1/l1/x/aetravm/avm"
)

func (s ContractZoneState) Validate() error {
	if err := s.Policy.Validate(); err != nil {
		return err
	}
	if s.NextCodeID == 0 {
		return errors.New("contract next code id must be positive")
	}
	if s.NextReceiptSeq == 0 {
		return errors.New("contract next receipt sequence must be positive")
	}
	if err := validateCodes(s.Codes, s.Policy); err != nil {
		return err
	}
	if err := validateContracts(s.Contracts, s.Codes, s.Policy); err != nil {
		return err
	}
	if err := validateQueuedMessages(s.QueuedMessages, s.Policy); err != nil {
		return err
	}
	if err := validateReceipts(s.Receipts); err != nil {
		return err
	}
	if s.BlockMessageCount > s.Policy.Limits.MaxMessagesPerBlock {
		return fmt.Errorf("contract block message count must be <= %d", s.Policy.Limits.MaxMessagesPerBlock)
	}
	return nil
}

func validateCodes(codes []ContractCode, policy ContractZonePolicy) error {
	seen := make(map[uint64]struct{}, len(codes))
	for i, code := range codes {
		if err := validateCode(code, policy); err != nil {
			return err
		}
		if _, found := seen[code.CodeID]; found {
			return errors.New("duplicate contract code id")
		}
		seen[code.CodeID] = struct{}{}
		if i > 0 && codes[i-1].CodeID >= code.CodeID {
			return errors.New("contract codes must be sorted canonically")
		}
	}
	return nil
}

func validateCode(code ContractCode, policy ContractZonePolicy) error {
	if code.CodeID == 0 {
		return errors.New("contract code id must be positive")
	}
	if code.Runtime != RuntimeAVM {
		return errors.New("contract readiness spec only stores AVM code")
	}
	if err := validateContractAddress("contract code owner", code.Owner); err != nil {
		return err
	}
	if len(code.CodeHash) != 32 {
		return errors.New("contract code hash must be 32 bytes")
	}
	if code.UploadedAt == 0 {
		return errors.New("contract code upload height must be positive")
	}
	if code.EncodedBytes == 0 || code.EncodedBytes > policy.Limits.MaxCodeSizeBytes {
		return fmt.Errorf("contract encoded code bytes must be in 1..%d", policy.Limits.MaxCodeSizeBytes)
	}
	verifier, err := avm.NewVerifier(policy.Runtime.AVMParams)
	if err != nil {
		return err
	}
	if err := verifier.Verify(code.AVMModule); err != nil {
		return err
	}
	hash, err := avm.CodeHash(code.AVMModule)
	if err != nil {
		return err
	}
	if !bytes.Equal(code.CodeHash, hash[:]) {
		return errors.New("contract code hash mismatch")
	}
	return nil
}

func validateContracts(contracts []ContractInstance, codes []ContractCode, policy ContractZonePolicy) error {
	seen := make(map[string]struct{}, len(contracts))
	for i, contract := range contracts {
		if err := validateContract(contract, codes, policy); err != nil {
			return err
		}
		key := string(contract.Address)
		if _, found := seen[key]; found {
			return errors.New("duplicate contract address")
		}
		seen[key] = struct{}{}
		if i > 0 && compareAddress(contracts[i-1].Address, contract.Address) >= 0 {
			return errors.New("contracts must be sorted canonically")
		}
	}
	return nil
}

func validateContract(contract ContractInstance, codes []ContractCode, policy ContractZonePolicy) error {
	if err := validateContractAddress("contract address", contract.Address); err != nil {
		return err
	}
	if contract.CodeID == 0 {
		return errors.New("contract code id must be positive")
	}
	if _, found := findCode(ContractZoneState{Codes: codes}, contract.CodeID); !found {
		return errors.New("contract references missing code")
	}
	if contract.Runtime != RuntimeAVM {
		return errors.New("contract runtime must be AVM")
	}
	if err := validateContractAddress("contract owner", contract.Owner); err != nil {
		return err
	}
	if err := validateContractAddress("contract admin", contract.Admin); err != nil {
		return err
	}
	if contract.CreatedHeight == 0 || contract.UpdatedHeight < contract.CreatedHeight {
		return errors.New("contract heights are invalid")
	}
	for _, entry := range contract.Storage {
		if entry.Namespace != ContractNamespace(contract.Address) {
			return errors.New("contract storage namespace mismatch")
		}
	}
	return ValidateStorageEntries(contract.Storage, policy.Limits)
}

func validateQueuedMessages(messages []async.MessageEnvelope, policy ContractZonePolicy) error {
	if len(messages) > int(policy.Limits.MaxMessagesPerBlock) {
		return fmt.Errorf("contract queued messages must be <= %d", policy.Limits.MaxMessagesPerBlock)
	}
	for _, msg := range messages {
		if err := validateContractAddress("queued message source", msg.Source); err != nil {
			return err
		}
		if err := validateContractAddress("queued message destination", msg.Destination); err != nil {
			return err
		}
		if len(msg.Body) > int(async.DefaultParams().MaxBodySize) {
			return errors.New("queued message body exceeds async limit")
		}
	}
	return nil
}

func validateReceipts(receipts []ContractReceipt) error {
	for i, receipt := range receipts {
		if receipt.Sequence == 0 {
			return errors.New("contract receipt sequence must be positive")
		}
		if i > 0 && receipts[i-1].Sequence >= receipt.Sequence {
			return errors.New("contract receipts must be sorted canonically")
		}
		if err := validateContractAddress("contract receipt address", receipt.Contract); err != nil {
			return err
		}
	}
	return nil
}

func normalizeContractState(state ContractZoneState) ContractZoneState {
	if state.NextCodeID == 0 {
		state.NextCodeID = 1
	}
	if state.NextReceiptSeq == 0 {
		state.NextReceiptSeq = 1
	}
	return state
}

func findCode(state ContractZoneState, codeID uint64) (ContractCode, bool) {
	for _, code := range state.Codes {
		if code.CodeID == codeID {
			return cloneCode(code), true
		}
	}
	return ContractCode{}, false
}

func findContract(state ContractZoneState, address sdk.AccAddress) (ContractInstance, bool) {
	for _, contract := range state.Contracts {
		if bytes.Equal(contract.Address, address) {
			return cloneContract(contract), true
		}
	}
	return ContractInstance{}, false
}

func upsertContract(contracts []ContractInstance, contract ContractInstance) []ContractInstance {
	out := cloneContracts(contracts)
	for i := range out {
		if bytes.Equal(out[i].Address, contract.Address) {
			out[i] = cloneContract(contract)
			return out
		}
	}
	return append(out, cloneContract(contract))
}

func sortContractCodes(codes []ContractCode) {
	sort.SliceStable(codes, func(i, j int) bool { return codes[i].CodeID < codes[j].CodeID })
}

func sortContracts(contracts []ContractInstance) {
	sort.SliceStable(contracts, func(i, j int) bool {
		return compareAddress(contracts[i].Address, contracts[j].Address) < 0
	})
}

func sortQueuedMessages(messages []async.MessageEnvelope) {
	sort.SliceStable(messages, func(i, j int) bool {
		if messages[i].CreatedLogicalTime != messages[j].CreatedLogicalTime {
			return messages[i].CreatedLogicalTime < messages[j].CreatedLogicalTime
		}
		if messages[i].QueryID != messages[j].QueryID {
			return messages[i].QueryID < messages[j].QueryID
		}
		return compareAddress(messages[i].Destination, messages[j].Destination) < 0
	})
}

func sortReceipts(receipts []ContractReceipt) {
	sort.SliceStable(receipts, func(i, j int) bool { return receipts[i].Sequence < receipts[j].Sequence })
}

func compareAddress(left, right sdk.AccAddress) int {
	return bytes.Compare(left, right)
}
