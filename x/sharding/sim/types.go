package sim

import loadtypes "github.com/sovereign-l1/l1/x/load/types"

const (
	MasterchainID	int32	= -1
	BaseWorkchain	int32	= 0
	BaseShardID	string	= ""
	FeeDenomNaet	string	= "naet"
)

type Validator struct {
	Address	string
	Power	int64
}

type WorkchainConfig struct {
	ID			int32
	AllowedVMs		[]string
	FeeDenom		string
	AddressFormat		string
	GenesisStateHash	string
	UpgradePolicy		string
}

type ShardID struct {
	WorkchainID	int32
	Prefix		string
}

type ShardState struct {
	ID			ShardID
	Height			uint64
	StateRoot		string
	MessageQueueRoot	string
	ReceiptRoot		string
	ValidatorSubset		[]string
	Queue			[]CrossShardMessage
	Receipts		map[string]Receipt
	Available		bool
}

type ShardHeader struct {
	ShardID			ShardID
	Height			uint64
	StateRoot		string
	MessageQueueRoot	string
	ReceiptRoot		string
	ValidatorSubset		[]string
	Available		bool
	Commitment		string
}

type CrossShardMessage struct {
	Source		ShardID
	Destination	ShardID
	MessageID	string
	Nonce		uint64
	Payload		[]byte
	RoutingKey	[]byte
	Proof		string
	Timeout		uint64
	Bounce		bool
	Bounced		bool
}

type Receipt struct {
	MessageID	string
	Source		ShardID
	Destination	ShardID
	Success		bool
	Height		uint64
	ResultCode	uint32
	Proof		string
}

type EquivocationEvidence struct {
	Validator	string
	ShardID		ShardID
	Height		uint64
	LeftRoot	string
	RightRoot	string
}

type WorkchainLoadState struct {
	WorkchainID			int32
	EMA				loadtypes.EMAState
	LastLoadScoreBps		uint32
	LastLoadBand			loadtypes.LoadBand
	ActiveShardCount		uint32
	TargetShardCount		uint32
	CooldownBlocks			uint64
	BelowTargetSinceHeight		uint64
	RoutingEpoch			uint64
	LastUpdateHeight		uint64
	LastValidatorEpochHeight	uint64
}

type MasterchainState struct {
	Height			uint64
	Validators		[]Validator
	StakingSnapshot		map[string]int64
	Workchains		map[int32]WorkchainConfig
	Shards			map[string]ShardState
	Headers			map[string]ShardHeader
	CrossShardReceipts	map[string]Receipt
	LoadStates		map[int32]WorkchainLoadState
	ConfigUpdates		[]string
	Evidence		[]EquivocationEvidence
	FinalityLag		uint64
	RandomnessSeed		string
}

type Simulator struct {
	state		MasterchainState
	processed	map[string]struct{}
	pendingReceipt	map[string]CrossShardMessage
}
