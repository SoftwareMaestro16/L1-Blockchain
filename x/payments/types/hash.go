package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"sort"
)

const HashHexLength = 64

func HashParts(parts ...string) string {
	h := sha256.New()
	writeString(h, "aetheris-payments-hash-parts-v1")
	for _, part := range parts {
		writeString(h, part)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeStateHash(state ChannelState) string {
	state = state.Normalize()
	h := sha256.New()
	writeString(h, "aetheris-payment-channel-state-v1")
	writeString(h, state.ChainID)
	writeString(h, state.ChannelID)
	writeString(h, string(state.ChannelType))
	writeString(h, state.Denom)
	writeUint64(h, state.Epoch)
	writeUint64(h, state.Nonce)
	writeString(h, state.PreviousStateHash)
	for _, balance := range state.Balances {
		writeString(h, balance.Participant)
		writeString(h, balance.Amount)
	}
	for _, condition := range state.Conditions {
		writeString(h, condition.ConditionID)
		writeString(h, string(condition.ConditionType))
		writeString(h, condition.Payer)
		writeString(h, condition.Payee)
		writeString(h, condition.Amount)
		writeString(h, condition.HashLock)
		writeUint64(h, condition.TimeoutHeight)
		writeUint64(h, condition.NonceStart)
		writeUint64(h, condition.NonceEnd)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeSignatureHash(signer, stateHash string) string {
	h := sha256.New()
	writeString(h, "aetheris-payment-state-signature-v1")
	writeString(h, signer)
	writeString(h, stateHash)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeVirtualChannelAnchor(vc VirtualChannel) string {
	vc = vc.Normalize()
	h := sha256.New()
	writeString(h, "aetheris-virtual-payment-channel-anchor-v1")
	writeString(h, vc.VirtualChannelID)
	for _, id := range vc.ParentChannelIDs {
		writeString(h, id)
	}
	for _, endpoint := range vc.Endpoints {
		writeString(h, endpoint)
	}
	writeString(h, vc.Capacity)
	writeUint64(h, vc.ExpiresHeight)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeSettlementHash(settlement SettlementRecord) string {
	settlement = settlement.Normalize()
	h := sha256.New()
	writeString(h, "aetheris-payment-settlement-v1")
	writeString(h, settlement.ChannelID)
	writeString(h, settlement.StateHash)
	writeUint64(h, settlement.Nonce)
	writeUint64(h, settlement.SettledHeight)
	writeString(h, settlement.SettlementFeeDenom)
	writeString(h, settlement.SettlementFee)
	for _, balance := range settlement.FinalBalances {
		writeString(h, balance.Participant)
		writeString(h, balance.Amount)
	}
	for _, penalty := range settlement.Penalties {
		writeString(h, penalty.Offender)
		writeString(h, penalty.Recipient)
		writeString(h, penalty.Denom)
		writeString(h, penalty.Amount)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeBatchRoot(operations []SettlementOperation) string {
	ordered := SortSettlementOperations(operations)
	h := sha256.New()
	writeString(h, "aetheris-payment-settlement-batch-v1")
	for _, op := range ordered {
		writeString(h, op.OperationID)
		writeString(h, string(op.OperationType))
		writeString(h, op.ChannelID)
		writeUint64(h, op.Nonce)
		writeString(h, op.StateHash)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ValidateHash(fieldName, value string) error {
	if len(value) != HashHexLength {
		return fmt.Errorf("%s must be %d lowercase hex chars", fieldName, HashHexLength)
	}
	for _, r := range value {
		if r >= '0' && r <= '9' || r >= 'a' && r <= 'f' {
			continue
		}
		return fmt.Errorf("%s must be %d lowercase hex chars", fieldName, HashHexLength)
	}
	return nil
}

func writeString(w interface{ Write([]byte) (int, error) }, value string) {
	writeUint64(w, uint64(len(value)))
	_, _ = w.Write([]byte(value))
}

func writeUint64(w interface{ Write([]byte) (int, error) }, value uint64) {
	var bz [8]byte
	binary.BigEndian.PutUint64(bz[:], value)
	_, _ = w.Write(bz[:])
}

func compareString(left, right string) int {
	return bytes.Compare([]byte(left), []byte(right))
}

func sortStrings(values []string) {
	sort.SliceStable(values, func(i, j int) bool {
		return values[i] < values[j]
	})
}
