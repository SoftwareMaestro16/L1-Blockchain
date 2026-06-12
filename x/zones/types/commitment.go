package types

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
)

type ZoneCommitment struct {
	ZoneID			ZoneID
	ZoneHeight		uint64
	StateRoot		string
	ReceiptRoot		string
	MessageRoot		string
	ExecutionResultRoot	string
	PreviousCommitment	string
	CommitmentHash		string
}

func NewZoneCommitment(
	zoneID ZoneID,
	zoneHeight uint64,
	stateRoot string,
	receiptRoot string,
	messageRoot string,
	executionResultRoot string,
	previousCommitment string,
) (ZoneCommitment, error) {
	commitment := ZoneCommitment{
		ZoneID:			zoneID,
		ZoneHeight:		zoneHeight,
		StateRoot:		stateRoot,
		ReceiptRoot:		receiptRoot,
		MessageRoot:		messageRoot,
		ExecutionResultRoot:	executionResultRoot,
		PreviousCommitment:	previousCommitment,
	}
	if err := commitment.ValidateFormat(); err != nil {
		return ZoneCommitment{}, err
	}
	commitment.CommitmentHash = ComputeCommitmentHash(commitment)
	return commitment, nil
}

func (c ZoneCommitment) ValidateFormat() error {
	if err := ValidateZoneID(c.ZoneID); err != nil {
		return err
	}
	if c.ZoneHeight == 0 {
		return errors.New("zone commitment height must be positive")
	}
	if err := ValidateHash("zone state root", c.StateRoot); err != nil {
		return err
	}
	if err := ValidateHash("zone receipt root", c.ReceiptRoot); err != nil {
		return err
	}
	if err := ValidateHash("zone message root", c.MessageRoot); err != nil {
		return err
	}
	if err := ValidateHash("zone execution result root", c.ExecutionResultRoot); err != nil {
		return err
	}
	if c.PreviousCommitment != "" {
		if err := ValidateHash("zone previous commitment", c.PreviousCommitment); err != nil {
			return err
		}
	}
	if c.CommitmentHash != "" {
		if err := ValidateHash("zone commitment hash", c.CommitmentHash); err != nil {
			return err
		}
	}
	return nil
}

func (c ZoneCommitment) ValidateHash() error {
	if err := c.ValidateFormat(); err != nil {
		return err
	}
	expected := ComputeCommitmentHash(c)
	if c.CommitmentHash != expected {
		return fmt.Errorf("zone commitment hash mismatch: expected %s", expected)
	}
	return nil
}

func ComputeCommitmentHash(c ZoneCommitment) string {
	h := sha256.New()
	writePart(h, "aetra-zone-commitment-v1")
	writePart(h, string(c.ZoneID))
	writeUint64(h, c.ZoneHeight)
	writePart(h, c.StateRoot)
	writePart(h, c.ReceiptRoot)
	writePart(h, c.MessageRoot)
	writePart(h, c.ExecutionResultRoot)
	writePart(h, c.PreviousCommitment)
	return hex.EncodeToString(h.Sum(nil))
}

type byteWriter interface {
	Write([]byte) (int, error)
}

func writePart(w byteWriter, value string) {
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], uint64(len(value)))
	_, _ = w.Write(length[:])
	_, _ = w.Write([]byte(value))
}

func writeUint64(w byteWriter, value uint64) {
	var bz [8]byte
	binary.BigEndian.PutUint64(bz[:], value)
	_, _ = w.Write(bz[:])
}
