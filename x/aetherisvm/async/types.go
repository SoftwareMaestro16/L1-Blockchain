package async

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	AddressDerivationDomain = "aetheris/async-contract/v1"
	BounceOpcode            = uint32(0xffff_fffe)
	RefundOpcode            = uint32(0xffff_fffd)
	ResultOK                = uint32(0)
	ResultNoDestination     = uint32(1)
	ResultExpired           = uint32(2)
	ResultExecutionFailed   = uint32(3)
	ResultLimitExceeded     = uint32(4)
	CodeHashLength          = 32
)

type Params struct {
	MaxMessagesPerTx           uint32
	MaxMessagesPerBlock        uint32
	MaxRecursionDepth          uint32
	MaxBodySize                uint32
	MaxStateSize               uint32
	MaxContractDeploysPerTx    uint32
	MaxContractDeploysPerBlock uint32
	MaxEmittedMessagesPerExec  uint32
	MaxStorageWritesPerExec    uint32
	ExecutionGasPerMessage     uint64
	StorageFeePerByte          sdkmath.Int
	ForwardingFee              sdkmath.Int
	ContractDeploymentCost     sdkmath.Int
}

type ContractAccount struct {
	Address     sdk.AccAddress
	CodeHash    []byte
	State       []byte
	BalanceNaet sdkmath.Int
	LogicalTime uint64
}

type MessageEnvelope struct {
	Source             sdk.AccAddress
	Destination        sdk.AccAddress
	Value              sdk.Coin
	Opcode             uint32
	QueryID            uint64
	Body               []byte
	Bounce             bool
	Bounced            bool
	CreatedLogicalTime uint64
	DeliverAtBlock     uint64
	// ExecutionBlockHeight is set only while a queued message is being delivered.
	ExecutionBlockHeight uint64
	DeadlineBlock        uint64
	GasLimit             uint64
	ForwardFee           sdk.Coin
	Depth                uint32
}

type QueuedMessage struct {
	TxIndex           uint64
	MessageIndex      uint32
	SourceLogicalTime uint64
	DestinationKey    string
	Sequence          uint64
	EnqueuedBlock     uint64
	Envelope          MessageEnvelope
}

type ExecutionResult struct {
	NewState      []byte
	Outgoing      []MessageEnvelope
	GasUsed       uint64
	StorageWrites uint32
	ResultCode    uint32
	Error         string
}

type ExecutionReceipt struct {
	Sequence       uint64
	Source         sdk.AccAddress
	Destination    sdk.AccAddress
	Opcode         uint32
	QueryID        uint64
	ResultCode     uint32
	GasUsed        uint64
	StorageFeeNaet sdkmath.Int
	ForwardFeeNaet sdkmath.Int
	Bounced        bool
	Error          string
}

type Observability struct {
	QueuedMessages    uint64
	ProcessedMessages uint64
	BouncedMessages   uint64
	RefundMessages    uint64
	FailedExecutions  uint64
	GasUsed           uint64
	QueueLag          uint64
}

type ExportedState struct {
	Params       Params
	Contracts    []ContractAccount
	Queue        []QueuedMessage
	Inbox        map[string][]QueuedMessage
	Outbox       map[string][]QueuedMessage
	Receipts     []ExecutionReceipt
	NextSequence uint64
	NextTxIndex  uint64
	BlockHeight  uint64
	Metrics      Observability
}

type Handler func(contract ContractAccount, msg MessageEnvelope) ExecutionResult

type Executor struct {
	params         Params
	contracts      map[string]ContractAccount
	queue          []QueuedMessage
	inbox          map[string][]QueuedMessage
	outbox         map[string][]QueuedMessage
	receipts       []ExecutionReceipt
	nextSequence   uint64
	nextTxIndex    uint64
	blockHeight    uint64
	metrics        Observability
	handlers       map[string]Handler
	deploysInBlock uint32
}

type DeploySpec struct {
	CodeHash    []byte
	Salt        []byte
	State       []byte
	BalanceNaet sdkmath.Int
}
