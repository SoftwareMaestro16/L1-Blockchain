package types

import (
	"errors"
	"fmt"
)

type GlobalMessageRoot struct {
	Height		uint64
	InboxRoot	string
	OutboxRoot	string
	MessageRoot	string
	ZoneCount	uint64
}

type ExecutionReceiptRoot struct {
	Height		uint64
	ReceiptsRoot	string
	ZoneCount	uint64
}

func NewGlobalMessageRoot(height uint64, inboxRoot string, outboxRoot string, zoneCount uint64) (GlobalMessageRoot, error) {
	root := GlobalMessageRoot{
		Height:		height,
		InboxRoot:	inboxRoot,
		OutboxRoot:	outboxRoot,
		ZoneCount:	zoneCount,
	}
	if err := root.ValidateFormat(); err != nil {
		return GlobalMessageRoot{}, err
	}
	root.MessageRoot = ComputeGlobalMessageRootHash(root)
	return root, nil
}

func NewExecutionReceiptRoot(height uint64, receiptsRoot string, zoneCount uint64) (ExecutionReceiptRoot, error) {
	root := ExecutionReceiptRoot{Height: height, ReceiptsRoot: receiptsRoot, ZoneCount: zoneCount}
	if err := root.Validate(); err != nil {
		return ExecutionReceiptRoot{}, err
	}
	return root, nil
}

func NewProofRoot(height uint64, rootType RootType, rootHash string, source string) (ProofRoot, error) {
	root := ProofRoot{Height: height, RootType: rootType, RootHash: rootHash, Source: source}
	if err := root.Validate(); err != nil {
		return ProofRoot{}, err
	}
	return root, nil
}

func (r GlobalMessageRoot) ValidateFormat() error {
	if r.Height == 0 {
		return errors.New("aetracore global message root height must be positive")
	}
	if r.ZoneCount == 0 {
		return errors.New("aetracore global message root zone count must be positive")
	}
	if err := ValidateHash("aetracore global inbox root", r.InboxRoot); err != nil {
		return err
	}
	if err := ValidateHash("aetracore global outbox root", r.OutboxRoot); err != nil {
		return err
	}
	if r.MessageRoot != "" {
		return ValidateHash("aetracore global message root", r.MessageRoot)
	}
	return nil
}

func (r GlobalMessageRoot) ValidateHash() error {
	if err := r.ValidateFormat(); err != nil {
		return err
	}
	expected := ComputeGlobalMessageRootHash(r)
	if r.MessageRoot != expected {
		return fmt.Errorf("aetracore global message root mismatch: expected %s", expected)
	}
	return nil
}

func (r ExecutionReceiptRoot) Validate() error {
	if r.Height == 0 {
		return errors.New("aetracore execution receipt root height must be positive")
	}
	if r.ZoneCount == 0 {
		return errors.New("aetracore execution receipt root zone count must be positive")
	}
	return ValidateHash("aetracore execution receipt root", r.ReceiptsRoot)
}

func ComputeGlobalMessageRootHash(r GlobalMessageRoot) string {
	return hashRoot("aetra-aek-global-message-root-v1", func(w byteWriter) {
		writeUint64(w, r.Height)
		writePart(w, r.InboxRoot)
		writePart(w, r.OutboxRoot)
		writeUint64(w, r.ZoneCount)
	})
}
