package avm

import (
	"errors"
	"fmt"

	contractstypes "github.com/sovereign-l1/l1/x/contracts/types"
)

const (
	DefaultMaxGasPerMessage	= 100_000_000
	DefaultMaxGasPerBlock	= 1_000_000_000

	GasRefundNone	GasRefundPolicy	= "none"

	ExitCodeOutOfGas	= contractstypes.ExitCodeOutOfGas
)

var ErrOutOfGas = errors.New("AVM out of gas")

type GasRefundPolicy string

type GasTable struct {
	BaseTx				uint64
	StoreCodeBase			uint64
	DeployBase			uint64
	ExecuteBase			uint64
	BytecodeVerificationPerByte	uint64
	CodeLoadPerByte			uint64
	InstructionCost			uint64
	MemoryAllocationPerByte		uint64
	MemoryAllocationPerPage		uint64
	StorageReadPerKeyByte		uint64
	StorageReadPerValueByte		uint64
	StorageWritePerKeyByte		uint64
	StorageWritePerValueByte	uint64
	StorageDelete			uint64
	HashFunction			uint64
	SignatureVerify			uint64
	EmitEvent			uint64
	SendInternalMessage		uint64
	SerializeDeserializePerByte	uint64
	QueueEnqueue			uint64
	QueueDequeue			uint64
	ReceiptWrite			uint64
	RefundPolicy			GasRefundPolicy
}

type GasMeter struct {
	limit		uint64
	used		uint64
	refunded	uint64
}

type OutOfGasError struct {
	Reason	string
	Limit	uint64
	Used	uint64
	Wanted	uint64
}

func DefaultGasTable() GasTable {
	return GasTable{
		BaseTx:				1_000,
		StoreCodeBase:			5_000,
		DeployBase:			10_000,
		ExecuteBase:			700,
		BytecodeVerificationPerByte:	2,
		CodeLoadPerByte:		1,
		InstructionCost:		10,
		MemoryAllocationPerByte:	1,
		MemoryAllocationPerPage:	64,
		StorageReadPerKeyByte:		3,
		StorageReadPerValueByte:	2,
		StorageWritePerKeyByte:		5,
		StorageWritePerValueByte:	8,
		StorageDelete:			200,
		HashFunction:			80,
		SignatureVerify:		5_000,
		EmitEvent:			100,
		SendInternalMessage:		250,
		SerializeDeserializePerByte:	1,
		QueueEnqueue:			50,
		QueueDequeue:			50,
		ReceiptWrite:			75,
		RefundPolicy:			GasRefundNone,
	}
}

func NewGasMeter(limit uint64) (*GasMeter, error) {
	if limit == 0 {
		return nil, errors.New("AVM gas limit must be positive")
	}
	return &GasMeter{limit: limit}, nil
}

func (m *GasMeter) ConsumeGas(reason string, amount uint64) error {
	if m == nil {
		return errors.New("AVM gas meter is nil")
	}
	if amount == 0 {
		return nil
	}
	if amount > ^uint64(0)-m.used {
		return OutOfGasError{Reason: reason, Limit: m.limit, Used: m.used, Wanted: amount}
	}
	next := m.used + amount
	if next > m.limit {
		m.used = m.limit
		return OutOfGasError{Reason: reason, Limit: m.limit, Used: m.used, Wanted: amount}
	}
	m.used = next
	return nil
}

func (m *GasMeter) RefundGas(reason string, amount uint64) error {
	if m == nil {
		return errors.New("AVM gas meter is nil")
	}
	if amount == 0 {
		return nil
	}
	if amount > m.used {
		return fmt.Errorf("AVM gas refund %q exceeds gas used", reason)
	}
	if amount > ^uint64(0)-m.refunded {
		return fmt.Errorf("AVM gas refund %q overflows refund counter", reason)
	}
	m.used -= amount
	m.refunded += amount
	return nil
}

func (m *GasMeter) GasUsed() uint64 {
	if m == nil {
		return 0
	}
	return m.used
}

func (m *GasMeter) GasRefunded() uint64 {
	if m == nil {
		return 0
	}
	return m.refunded
}

func (e OutOfGasError) Error() string {
	if e.Reason == "" {
		return ErrOutOfGas.Error()
	}
	return fmt.Sprintf("%s: %s", ErrOutOfGas, e.Reason)
}

func (e OutOfGasError) Is(target error) bool {
	return target == ErrOutOfGas
}

func (t GasTable) Validate() error {
	checks := []struct {
		name	string
		value	uint64
	}{
		{"base tx", t.BaseTx},
		{"store code base", t.StoreCodeBase},
		{"deploy base", t.DeployBase},
		{"execute base", t.ExecuteBase},
		{"bytecode verification per byte", t.BytecodeVerificationPerByte},
		{"code load per byte", t.CodeLoadPerByte},
		{"instruction cost", t.InstructionCost},
		{"memory allocation per byte", t.MemoryAllocationPerByte},
		{"memory allocation per page", t.MemoryAllocationPerPage},
		{"storage read per key byte", t.StorageReadPerKeyByte},
		{"storage read per value byte", t.StorageReadPerValueByte},
		{"storage write per key byte", t.StorageWritePerKeyByte},
		{"storage write per value byte", t.StorageWritePerValueByte},
		{"storage delete", t.StorageDelete},
		{"hash function", t.HashFunction},
		{"signature verify", t.SignatureVerify},
		{"emit event", t.EmitEvent},
		{"send internal message", t.SendInternalMessage},
		{"serialize deserialize per byte", t.SerializeDeserializePerByte},
		{"queue enqueue", t.QueueEnqueue},
		{"queue dequeue", t.QueueDequeue},
		{"receipt write", t.ReceiptWrite},
	}
	for _, check := range checks {
		if check.value == 0 {
			return fmt.Errorf("AVM gas table %s cost must be positive", check.name)
		}
	}
	if t.RefundPolicy == "" {
		return errors.New("AVM gas refund policy is required")
	}
	if t.RefundPolicy != GasRefundNone {
		return fmt.Errorf("AVM gas refund policy %q is unsupported", t.RefundPolicy)
	}
	return nil
}
