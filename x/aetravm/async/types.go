package async

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	AddressDerivationDomain	= "aetra/async-contract/v1"
	BounceOpcode		= uint32(0xffff_fffe)
	RefundOpcode		= uint32(0xffff_fffd)
	ResultOK		= uint32(0)
	ResultNoDestination	= uint32(1)
	ResultExpired		= uint32(2)
	ResultExecutionFailed	= uint32(3)
	ResultLimitExceeded	= uint32(4)
	ResultBounceSuppressed	= uint32(5)
	ResultRefundSuppressed	= uint32(6)

	ResultForbiddenHostCall		= uint32(7)
	ResultInvalidJump		= uint32(8)
	ResultCallStackOverflow		= uint32(9)
	ResultContinuationNotFound	= uint32(10)
	ResultRecursionLimitExceeded	= uint32(11)
	ResultInvalidMemoryAccess	= uint32(12)
	ResultNullReference		= uint32(13)
	ResultInvalidChunkReference	= uint32(14)
	ResultCorruptedStateObject	= uint32(15)
	ResultDivisionByZero		= uint32(16)
	ResultInvalidShift		= uint32(17)
	ResultArithmeticUnderflow	= uint32(18)
	ResultGasLimitExceeded		= uint32(19)
	ResultGasReservationFailed	= uint32(20)
	ResultExecutionTimeout		= uint32(21)
	ResultStackOverflow		= uint32(22)
	ResultStackUnderflow		= uint32(23)
	ResultTypeCheckError		= uint32(24)
	ResultMessageRoutingFailed	= uint32(25)
	ResultQueueOverflow		= uint32(26)
	ResultShardUnavailable		= uint32(27)
	ResultInsufficientBalance	= uint32(28)
	ResultInsufficientGas		= uint32(29)
	ResultStateCorruption		= uint32(30)
	ResultStateVersionMismatch	= uint32(31)
	ResultSnapshotFailure		= uint32(32)
	ResultExplicitAbort		= uint32(33)
	ResultAssertionFailed		= uint32(34)
	ResultAccountStateTooBig	= uint32(35)
	ResultStorageRentDebt		= uint32(36)
	ResultInactiveFrozen		= uint32(37)
	ResultActionBudgetExceeded	= uint32(38)

	CodeHashLength	= 32

	ContractStatusActive	= "active"
	ContractStatusFrozen	= "frozen"

	QueueStatusPending	= "pending"
	QueueStatusExecuted	= "executed"
	QueueStatusFailed	= "failed"
	QueueStatusExpired	= "expired"
	QueueStatusBounced	= "bounced"

	ExecutionKindExternal	= "external"
	ExecutionKindInternal	= "internal"
	ExecutionKindBounced	= "bounced"
	ExecutionKindSystem	= "system"

	FailedPhaseValidation	= "validation"
	FailedPhaseDispatch	= "dispatch"
	FailedPhaseExecution	= "execution"
	FailedPhaseStorage	= "storage"
	FailedPhaseQueue	= "queue"

	EventCodeStored		= "avm.code_stored"
	EventContractDeployed	= "avm.contract_deployed"
	EventExternalExecuted	= "avm.external_executed"
	EventInternalExecuted	= "avm.internal_executed"
	EventMessageQueued	= "avm.message_queued"
	EventMessageBounced	= "avm.message_bounced"
	EventContractFrozen	= "avm.contract_frozen"
	EventContractUnfrozen	= "avm.contract_unfrozen"
	EventRentPaid		= "avm.rent_paid"
)

type Params struct {
	MaxMessagesPerTx		uint32
	MaxMessagesPerBlock		uint32
	MaxQueuedMessagesPerContract	uint32
	MaxProcessingAttempts		uint32
	MaxRecursionDepth		uint32
	MaxBodySize			uint32
	MaxStateSize			uint32
	MaxContractDeploysPerTx		uint32
	MaxContractDeploysPerBlock	uint32
	MaxEmittedMessagesPerExec	uint32
	MaxStorageWritesPerExec		uint32
	MaxActionsPerExecution		uint32
	MaxRetriesPerMessage		uint32
	DefaultRetryDelayBlocks		uint64
	MaxRetryDelayBlocks		uint64
	MaxDeadLetters			uint32
	ExecutionGasPerMessage		uint64
	StorageFeePerByte		sdkmath.Int
	ForwardingFee			sdkmath.Int
	ContractDeploymentCost		sdkmath.Int
}

type ContractAccount struct {
	Address			sdk.AccAddress
	CodeHash		[]byte
	State			[]byte
	BalanceNaet		sdkmath.Int
	LogicalTime		uint64
	Status			string
	StorageRentDebtNaet	sdkmath.Int
	LastStorageChargeHeight	uint64
}

type MessageEnvelope struct {
	Source			sdk.AccAddress
	Destination		sdk.AccAddress
	Value			sdk.Coin
	Opcode			uint32
	QueryID			uint64
	Body			[]byte
	Bounce			bool
	Bounced			bool
	CreatedLogicalTime	uint64
	DeliverAtBlock		uint64
	RetryCount		uint32
	MaxRetries		uint32
	RetryDelayBlocks	uint64
	// ExecutionBlockHeight is set only while a queued message is being delivered.
	ExecutionBlockHeight	uint64
	DeadlineBlock		uint64
	GasLimit		uint64
	ForwardFee		sdk.Coin
	Depth			uint32
	RefundOfSequence	uint64
}

type QueuedMessage struct {
	MessageID		[]byte
	TxHeight		uint64
	TxIndex			uint64
	MessageIndex		uint32
	SourceLogicalTime	uint64
	DestinationKey		string
	Sequence		uint64
	EnqueuedBlock		uint64
	CreatedHeight		uint64
	ScheduledHeight		uint64
	Attempts		uint32
	Status			string
	Envelope		MessageEnvelope
}

type DeadLetter struct {
	Sequence	uint64
	FailedSequence	uint64
	RecordedBlock	uint64
	Envelope	MessageEnvelope
	Receipt		ExecutionReceipt
	Reason		string
}

type ExecutionResult struct {
	NewState	[]byte
	Outgoing	[]MessageEnvelope
	GasUsed		uint64
	StorageWrites	uint32
	ResultCode	uint32
	Error		string
}

type ExecutionReceipt struct {
	Sequence		uint64
	ReceiptID		string
	TxHash			string
	MessageID		[]byte
	ExecutionKind		string
	ContractAddress		sdk.AccAddress
	Caller			sdk.AccAddress
	Source			sdk.AccAddress
	Destination		sdk.AccAddress
	Opcode			uint32
	QueryID			uint64
	GasLimit		uint64
	ResultCode		uint32
	ExitCode		uint32
	ExitReason		string
	FailedPhase		string
	GasUsed			uint64
	StorageFeeNaet		sdkmath.Int
	ForwardFeeNaet		sdkmath.Int
	FeeChargedNaet		sdkmath.Int
	ValueInNaet		sdkmath.Int
	ValueOutNaet		sdkmath.Int
	StateRootBefore		string
	StateRootAfter		string
	EmittedMessageIDs	[][]byte
	Events			[]AVMEvent
	Bounced			bool
	BounceCreated		bool
	RefundCreated		bool
	Refunded		bool
	RefundAmountNaet	sdkmath.Int
	RefundFeeNaet		sdkmath.Int
	RefundOfSequence	uint64
	RefundReason		string
	QueueStatus		string
	RetryCount		uint32
	RetryScheduled		bool
	Height			uint64
	LogicalTime		uint64
	StateCommitted		bool
	Error			string
}

type AVMEvent struct {
	Type		string
	Attributes	[]AVMEventAttribute
}

type AVMEventAttribute struct {
	Key	string
	Value	string
}

type Observability struct {
	QueuedMessages		uint64
	ProcessedMessages	uint64
	BouncedMessages		uint64
	RefundMessages		uint64
	RetriedMessages		uint64
	DeadLetterMessages	uint64
	FailedExecutions	uint64
	GasUsed			uint64
	QueueLag		uint64
	DeploymentCostsNaet	uint64
}

type ExportedState struct {
	Params			Params
	Contracts		[]ContractAccount
	Queue			[]QueuedMessage
	Inbox			map[string][]QueuedMessage
	Outbox			map[string][]QueuedMessage
	DeadLetters		[]DeadLetter
	Receipts		[]ExecutionReceipt
	NextSequence		uint64
	NextTxIndex		uint64
	NextDeadLetterSequence	uint64
	BlockHeight		uint64
	Metrics			Observability
}

type Handler func(contract ContractAccount, msg MessageEnvelope) ExecutionResult

type Executor struct {
	params			Params
	contracts		map[string]ContractAccount
	queue			[]QueuedMessage
	inbox			map[string][]QueuedMessage
	outbox			map[string][]QueuedMessage
	deadLetters		[]DeadLetter
	receipts		[]ExecutionReceipt
	nextSequence		uint64
	nextTxIndex		uint64
	nextDeadLetterSequence	uint64
	blockHeight		uint64
	metrics			Observability
	handlers		map[string]Handler
	deploysInBlock		uint32
}

type DeploySpec struct {
	CodeHash	[]byte
	Salt		[]byte
	State		[]byte
	BalanceNaet	sdkmath.Int
}
