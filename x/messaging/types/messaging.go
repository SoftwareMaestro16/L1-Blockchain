package types

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/app/addressing"
	appparams "github.com/sovereign-l1/l1/app/params"
	"github.com/sovereign-l1/l1/x/aetravm/async"
)

const MessageIDBytes = 32

type Message struct {
	ID		[]byte
	Source		sdk.AccAddress
	Destination	sdk.AccAddress
	ValueNaet	sdkmath.Int
	Opcode		uint32
	QueryID		uint64
	Body		[]byte
	Bounce		bool
	Deadline	uint64
	GasLimit	uint64
	CreatedLT	uint64
}

func NewMessage(source, destination sdk.AccAddress, valueNaet sdkmath.Int, opcode uint32, queryID uint64, body []byte, bounce bool, deadline uint64, gasLimit uint64, createdLT uint64) (Message, error) {
	msg := Message{
		Source:		append(sdk.AccAddress(nil), source...),
		Destination:	append(sdk.AccAddress(nil), destination...),
		ValueNaet:	valueNaet,
		Opcode:		opcode,
		QueryID:	queryID,
		Body:		append([]byte(nil), body...),
		Bounce:		bounce,
		Deadline:	deadline,
		GasLimit:	gasLimit,
		CreatedLT:	createdLT,
	}
	msg.ID = MessageID(msg)
	if err := msg.Validate(async.DefaultParams()); err != nil {
		return Message{}, err
	}
	return msg, nil
}

func (m Message) Validate(params async.Params) error {
	if len(m.ID) > 0 && len(m.ID) != MessageIDBytes {
		return fmt.Errorf("message id must be %d bytes", MessageIDBytes)
	}
	if err := addressing.RejectZeroAddress("message source", m.Source); err != nil {
		return err
	}
	if err := addressing.RejectZeroAddress("message destination", m.Destination); err != nil {
		return err
	}
	if m.ValueNaet.IsNil() || m.ValueNaet.IsNegative() {
		return errors.New("message value_naet must be non-negative")
	}
	if len(m.Body) > int(params.MaxBodySize) {
		return fmt.Errorf("message body size must be <= %d", params.MaxBodySize)
	}
	if m.GasLimit == 0 {
		return errors.New("message gas_limit must be positive")
	}
	return nil
}

func (m Message) Envelope(params async.Params) (async.MessageEnvelope, error) {
	if err := m.Validate(params); err != nil {
		return async.MessageEnvelope{}, err
	}
	return async.MessageEnvelope{
		Source:			append(sdk.AccAddress(nil), m.Source...),
		Destination:		append(sdk.AccAddress(nil), m.Destination...),
		Value:			sdk.NewCoin(appparams.BaseDenom, m.ValueNaet),
		Opcode:			m.Opcode,
		QueryID:		m.QueryID,
		Body:			append([]byte(nil), m.Body...),
		Bounce:			m.Bounce,
		DeadlineBlock:		m.Deadline,
		GasLimit:		m.GasLimit,
		CreatedLogicalTime:	m.CreatedLT,
		ForwardFee:		sdk.NewCoin(appparams.BaseDenom, params.ForwardingFee),
	}, nil
}

func EnqueueMessage(executor *async.Executor, params async.Params, msg Message) error {
	envelope, err := msg.Envelope(params)
	if err != nil {
		return err
	}
	return executor.EnqueueTxMessages([]async.MessageEnvelope{envelope})
}

func DeliverMessages(executor *async.Executor, height uint64) ([]async.ExecutionReceipt, error) {
	return executor.ProcessBlock(height)
}

func MessageID(m Message) []byte {
	h := sha256.New()
	writeBytes(h.Write, m.Source)
	writeBytes(h.Write, m.Destination)
	writeBytes(h.Write, []byte(m.ValueNaet.String()))
	writeU32(h.Write, m.Opcode)
	writeU64(h.Write, m.QueryID)
	writeBytes(h.Write, m.Body)
	writeBool(h.Write, m.Bounce)
	writeU64(h.Write, m.Deadline)
	writeU64(h.Write, m.GasLimit)
	writeU64(h.Write, m.CreatedLT)
	return h.Sum(nil)
}

func writeBytes(write func([]byte) (int, error), bz []byte) {
	writeU32(write, uint32(len(bz)))
	_, _ = write(bz)
}

func writeBool(write func([]byte) (int, error), value bool) {
	if value {
		_, _ = write([]byte{1})
		return
	}
	_, _ = write([]byte{0})
}

func writeU32(write func([]byte) (int, error), value uint32) {
	var out [4]byte
	binary.BigEndian.PutUint32(out[:], value)
	_, _ = write(out[:])
}

func writeU64(write func([]byte) (int, error), value uint64) {
	var out [8]byte
	binary.BigEndian.PutUint64(out[:], value)
	_, _ = write(out[:])
}
