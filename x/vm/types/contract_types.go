package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/x/aetravm/async"
	"github.com/sovereign-l1/l1/x/aetravm/avm"
)

const (
	UploadModeGovernanceOnly	= "governance-only"
	UploadModeAllowlistTestnet	= "allowlist-testnet"

	InstantiateModeCodeOwnerOnly	= "code-owner-only"
	InstantiateModeEverybody	= "everybody"
)

type GasModel struct {
	DeployGas		uint64
	ExecuteGas		uint64
	QueryGas		uint64
	StorageWriteGas		uint64
	MessageForwardingGas	uint64
}

type ContractLimits struct {
	MaxCodeSizeBytes	uint64
	MaxStateSizeBytes	uint64
	MaxStorageKeyBytes	uint32
	MaxStorageValueBytes	uint64
	MaxQueryResponseBytes	uint64
	MaxQueryDepth		uint32
	MaxEmittedMessages	uint32
	MaxMessagesPerBlock	uint32
}

type ContractHostPolicy struct {
	ConsensusRandomnessEnabled	bool
	ExternalAPIsEnabled		bool
	LocalTimeEnabled		bool
	CrossContractStateMutation	bool
}

type ContractZonePolicy struct {
	Runtime			RuntimePolicy
	GovernanceAuthority	sdk.AccAddress
	UploadMode		string
	UploadAllowlist		[]sdk.AccAddress
	InstantiateMode		string
	MigrationsEnabled	bool
	GasModel		GasModel
	Limits			ContractLimits
	HostPolicy		ContractHostPolicy
	TestnetAllowlist	bool
}

type ContractCode struct {
	CodeID		uint64
	Runtime		string
	CodeHash	[]byte
	Owner		sdk.AccAddress
	AVMModule	avm.Module
	UploadedAt	uint64
	EncodedBytes	uint64
}

type StorageEntry struct {
	Namespace	string
	Key		string
	Value		[]byte
}

type ContractInstance struct {
	Address		sdk.AccAddress
	CodeID		uint64
	Runtime		string
	Owner		sdk.AccAddress
	Admin		sdk.AccAddress
	Storage		[]StorageEntry
	CreatedHeight	uint64
	UpdatedHeight	uint64
}

type ContractCall struct {
	Actor			sdk.AccAddress
	Contract		sdk.AccAddress
	Entrypoint		avm.Entrypoint
	GasLimit		uint64
	QueryDepth		uint32
	QueryResponseBytes	uint64
	Body			[]byte
	EmitDestination		sdk.AccAddress
}

type ContractReceipt struct {
	Sequence	uint64
	Contract	sdk.AccAddress
	Entrypoint	avm.Entrypoint
	ResultCode	uint32
	GasUsed		uint64
	StorageWrites	uint32
	EmittedMessages	uint32
	ReturnValue	uint64
	QueryResponse	[]byte
	Error		string
}

type ContractZoneState struct {
	Policy			ContractZonePolicy
	Codes			[]ContractCode
	Contracts		[]ContractInstance
	QueuedMessages		[]async.MessageEnvelope
	Receipts		[]ContractReceipt
	NextCodeID		uint64
	NextReceiptSeq		uint64
	BlockMessageCount	uint32
}
