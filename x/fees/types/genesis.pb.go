package types

import (
	fmt "fmt"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3	// please upgrade the proto package

type GenesisState struct {
	Params			Params			`protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
	ProtocolFeeState	ProtocolFeeState	`protobuf:"bytes,2,opt,name=protocol_fee_state,json=protocolFeeState,proto3" json:"protocol_fee_state"`
	CongestionBps		uint32			`protobuf:"varint,3,opt,name=congestion_bps,json=congestionBps,proto3" json:"congestion_bps,omitempty"`
}

func (m *GenesisState) Reset()		{ *m = GenesisState{} }
func (m *GenesisState) String() string	{ return proto.CompactTextString(m) }
func (*GenesisState) ProtoMessage()	{}
func (*GenesisState) Descriptor() ([]byte, []int) {
	return fileDescriptor_e8fec93f5a3aedc0, []int{0}
}
func (m *GenesisState) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *GenesisState) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_GenesisState.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *GenesisState) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GenesisState.Merge(m, src)
}
func (m *GenesisState) XXX_Size() int {
	return m.Size()
}
func (m *GenesisState) XXX_DiscardUnknown() {
	xxx_messageInfo_GenesisState.DiscardUnknown(m)
}

var xxx_messageInfo_GenesisState proto.InternalMessageInfo

func (m *GenesisState) GetParams() Params {
	if m != nil {
		return m.Params
	}
	return Params{}
}

func (m *GenesisState) GetProtocolFeeState() ProtocolFeeState {
	if m != nil {
		return m.ProtocolFeeState
	}
	return ProtocolFeeState{}
}

func (m *GenesisState) GetCongestionBps() uint32 {
	if m != nil {
		return m.CongestionBps
	}
	return 0
}

type Params struct {
	AllowedFeeDenoms		[]string	`protobuf:"bytes,1,rep,name=allowed_fee_denoms,json=allowedFeeDenoms,proto3" json:"allowed_fee_denoms,omitempty"`
	ValidatorRewardsRatio		string		`protobuf:"bytes,2,opt,name=validator_rewards_ratio,json=validatorRewardsRatio,proto3" json:"validator_rewards_ratio,omitempty"`
	CommunityPoolRatio		string		`protobuf:"bytes,3,opt,name=community_pool_ratio,json=communityPoolRatio,proto3" json:"community_pool_ratio,omitempty"`
	MinFeeAmount			string		`protobuf:"bytes,4,opt,name=min_fee_amount,json=minFeeAmount,proto3" json:"min_fee_amount,omitempty"`
	FeeCollectorModule		string		`protobuf:"bytes,5,opt,name=fee_collector_module,json=feeCollectorModule,proto3" json:"fee_collector_module,omitempty"`
	ValidatorRewardsTarget		string		`protobuf:"bytes,6,opt,name=validator_rewards_target,json=validatorRewardsTarget,proto3" json:"validator_rewards_target,omitempty"`
	CommunityPoolTarget		string		`protobuf:"bytes,7,opt,name=community_pool_target,json=communityPoolTarget,proto3" json:"community_pool_target,omitempty"`
	BaseFeeAmount			string		`protobuf:"bytes,8,opt,name=base_fee_amount,json=baseFeeAmount,proto3" json:"base_fee_amount,omitempty"`
	MaxFeeAmount			string		`protobuf:"bytes,9,opt,name=max_fee_amount,json=maxFeeAmount,proto3" json:"max_fee_amount,omitempty"`
	TargetBlockUtilizationBps	uint32		`protobuf:"varint,10,opt,name=target_block_utilization_bps,json=targetBlockUtilizationBps,proto3" json:"target_block_utilization_bps,omitempty"`
	CongestionThresholdBps		uint32		`protobuf:"varint,11,opt,name=congestion_threshold_bps,json=congestionThresholdBps,proto3" json:"congestion_threshold_bps,omitempty"`
	MaxTxGas			uint64		`protobuf:"varint,12,opt,name=max_tx_gas,json=maxTxGas,proto3" json:"max_tx_gas,omitempty"`
	MaxBlockGas			uint64		`protobuf:"varint,13,opt,name=max_block_gas,json=maxBlockGas,proto3" json:"max_block_gas,omitempty"`
	MaxBlockTxs			uint64		`protobuf:"varint,14,opt,name=max_block_txs,json=maxBlockTxs,proto3" json:"max_block_txs,omitempty"`
	MaxSenderTxsPerBlock		uint64		`protobuf:"varint,15,opt,name=max_sender_txs_per_block,json=maxSenderTxsPerBlock,proto3" json:"max_sender_txs_per_block,omitempty"`
	StakeTxAllowanceStepAmount	string		`protobuf:"bytes,16,opt,name=stake_tx_allowance_step_amount,json=stakeTxAllowanceStepAmount,proto3" json:"stake_tx_allowance_step_amount,omitempty"`
	MaxSenderTxsPerBlockWithStake	uint64		`protobuf:"varint,17,opt,name=max_sender_txs_per_block_with_stake,json=maxSenderTxsPerBlockWithStake,proto3" json:"max_sender_txs_per_block_with_stake,omitempty"`
	FeePriorityWeightBps		uint32		`protobuf:"varint,18,opt,name=fee_priority_weight_bps,json=feePriorityWeightBps,proto3" json:"fee_priority_weight_bps,omitempty"`
	StakePriorityWeightBps		uint32		`protobuf:"varint,19,opt,name=stake_priority_weight_bps,json=stakePriorityWeightBps,proto3" json:"stake_priority_weight_bps,omitempty"`
	// target_transfer_fee_naet is the governance anchor for a standard transfer
	// fee (10,000,000 naet = 0.01 AET). All bounded surcharges and discounts are
	// expressed relative to this value.  Must be a governance/genesis parameter;
	// never hardcoded in binary logic.
	TargetTransferFeeAmount	string	`protobuf:"bytes,20,opt,name=target_transfer_fee_amount,json=targetTransferFeeAmount,proto3" json:"target_transfer_fee_amount,omitempty"`
	// congestion_surcharge_cap_bps caps the bounded_congestion_surcharge as a
	// fraction of target_transfer_fee_naet (in basis points).
	CongestionSurchargeCapBps	uint32	`protobuf:"varint,21,opt,name=congestion_surcharge_cap_bps,json=congestionSurchargeCapBps,proto3" json:"congestion_surcharge_cap_bps,omitempty"`
	// low_reputation_premium_cap_naet caps the low_reputation_premium addend.
	LowReputationPremiumCapAmount	string	`protobuf:"bytes,22,opt,name=low_reputation_premium_cap_amount,json=lowReputationPremiumCapAmount,proto3" json:"low_reputation_premium_cap_amount,omitempty"`
	// reputation_discount_cap_naet caps the reputation discount so fees never
	// reach zero.
	ReputationDiscountCapAmount	string	`protobuf:"bytes,23,opt,name=reputation_discount_cap_amount,json=reputationDiscountCapAmount,proto3" json:"reputation_discount_cap_amount,omitempty"`
	// storage_rent_byte_fee_amount is the fee per byte of new persistent state
	// created or enlarged by a transaction.
	StorageRentByteFeeAmount	string	`protobuf:"bytes,24,opt,name=storage_rent_byte_fee_amount,json=storageRentByteFeeAmount,proto3" json:"storage_rent_byte_fee_amount,omitempty"`
	// base_fee_per_gas_amount is the current_base_fee_per_gas_naet parameter
	// used in the gas component of the full fee formula.
	BaseFeePerGasAmount	string	`protobuf:"bytes,25,opt,name=base_fee_per_gas_amount,json=baseFeePerGasAmount,proto3" json:"base_fee_per_gas_amount,omitempty"`
	// byte_fee_amount is the per-byte component of the full fee formula.
	ByteFeeAmount	string	`protobuf:"bytes,26,opt,name=byte_fee_amount,json=byteFeeAmount,proto3" json:"byte_fee_amount,omitempty"`
	// message_fee_amount is the per-message component of the full fee formula.
	MessageFeeAmount	string	`protobuf:"bytes,27,opt,name=message_fee_amount,json=messageFeeAmount,proto3" json:"message_fee_amount,omitempty"`
}

func (m *Params) Reset()		{ *m = Params{} }
func (m *Params) String() string	{ return proto.CompactTextString(m) }
func (*Params) ProtoMessage()		{}
func (*Params) Descriptor() ([]byte, []int) {
	return fileDescriptor_e8fec93f5a3aedc0, []int{1}
}
func (m *Params) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Params) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Params.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Params) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Params.Merge(m, src)
}
func (m *Params) XXX_Size() int {
	return m.Size()
}
func (m *Params) XXX_DiscardUnknown() {
	xxx_messageInfo_Params.DiscardUnknown(m)
}

var xxx_messageInfo_Params proto.InternalMessageInfo

func (m *Params) GetAllowedFeeDenoms() []string {
	if m != nil {
		return m.AllowedFeeDenoms
	}
	return nil
}

func (m *Params) GetValidatorRewardsRatio() string {
	if m != nil {
		return m.ValidatorRewardsRatio
	}
	return ""
}

func (m *Params) GetCommunityPoolRatio() string {
	if m != nil {
		return m.CommunityPoolRatio
	}
	return ""
}

func (m *Params) GetMinFeeAmount() string {
	if m != nil {
		return m.MinFeeAmount
	}
	return ""
}

func (m *Params) GetFeeCollectorModule() string {
	if m != nil {
		return m.FeeCollectorModule
	}
	return ""
}

func (m *Params) GetValidatorRewardsTarget() string {
	if m != nil {
		return m.ValidatorRewardsTarget
	}
	return ""
}

func (m *Params) GetCommunityPoolTarget() string {
	if m != nil {
		return m.CommunityPoolTarget
	}
	return ""
}

func (m *Params) GetBaseFeeAmount() string {
	if m != nil {
		return m.BaseFeeAmount
	}
	return ""
}

func (m *Params) GetMaxFeeAmount() string {
	if m != nil {
		return m.MaxFeeAmount
	}
	return ""
}

func (m *Params) GetTargetBlockUtilizationBps() uint32 {
	if m != nil {
		return m.TargetBlockUtilizationBps
	}
	return 0
}

func (m *Params) GetCongestionThresholdBps() uint32 {
	if m != nil {
		return m.CongestionThresholdBps
	}
	return 0
}

func (m *Params) GetMaxTxGas() uint64 {
	if m != nil {
		return m.MaxTxGas
	}
	return 0
}

func (m *Params) GetMaxBlockGas() uint64 {
	if m != nil {
		return m.MaxBlockGas
	}
	return 0
}

func (m *Params) GetMaxBlockTxs() uint64 {
	if m != nil {
		return m.MaxBlockTxs
	}
	return 0
}

func (m *Params) GetMaxSenderTxsPerBlock() uint64 {
	if m != nil {
		return m.MaxSenderTxsPerBlock
	}
	return 0
}

func (m *Params) GetStakeTxAllowanceStepAmount() string {
	if m != nil {
		return m.StakeTxAllowanceStepAmount
	}
	return ""
}

func (m *Params) GetMaxSenderTxsPerBlockWithStake() uint64 {
	if m != nil {
		return m.MaxSenderTxsPerBlockWithStake
	}
	return 0
}

func (m *Params) GetFeePriorityWeightBps() uint32 {
	if m != nil {
		return m.FeePriorityWeightBps
	}
	return 0
}

func (m *Params) GetStakePriorityWeightBps() uint32 {
	if m != nil {
		return m.StakePriorityWeightBps
	}
	return 0
}

func (m *Params) GetTargetTransferFeeAmount() string {
	if m != nil {
		return m.TargetTransferFeeAmount
	}
	return ""
}

func (m *Params) GetCongestionSurchargeCapBps() uint32 {
	if m != nil {
		return m.CongestionSurchargeCapBps
	}
	return 0
}

func (m *Params) GetLowReputationPremiumCapAmount() string {
	if m != nil {
		return m.LowReputationPremiumCapAmount
	}
	return ""
}

func (m *Params) GetReputationDiscountCapAmount() string {
	if m != nil {
		return m.ReputationDiscountCapAmount
	}
	return ""
}

func (m *Params) GetStorageRentByteFeeAmount() string {
	if m != nil {
		return m.StorageRentByteFeeAmount
	}
	return ""
}

func (m *Params) GetBaseFeePerGasAmount() string {
	if m != nil {
		return m.BaseFeePerGasAmount
	}
	return ""
}

func (m *Params) GetByteFeeAmount() string {
	if m != nil {
		return m.ByteFeeAmount
	}
	return ""
}

func (m *Params) GetMessageFeeAmount() string {
	if m != nil {
		return m.MessageFeeAmount
	}
	return ""
}

type ProtocolFeeState struct {
	TotalCollected		github_com_cosmos_cosmos_sdk_types.Coins	`protobuf:"bytes,1,rep,name=total_collected,json=totalCollected,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"total_collected"`
	ValidatorRewards	github_com_cosmos_cosmos_sdk_types.Coins	`protobuf:"bytes,2,rep,name=validator_rewards,json=validatorRewards,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"validator_rewards"`
	CommunityPool		github_com_cosmos_cosmos_sdk_types.Coins	`protobuf:"bytes,3,rep,name=community_pool,json=communityPool,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"community_pool"`
}

func (m *ProtocolFeeState) Reset()		{ *m = ProtocolFeeState{} }
func (m *ProtocolFeeState) String() string	{ return proto.CompactTextString(m) }
func (*ProtocolFeeState) ProtoMessage()		{}
func (*ProtocolFeeState) Descriptor() ([]byte, []int) {
	return fileDescriptor_e8fec93f5a3aedc0, []int{2}
}
func (m *ProtocolFeeState) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ProtocolFeeState) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ProtocolFeeState.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *ProtocolFeeState) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ProtocolFeeState.Merge(m, src)
}
func (m *ProtocolFeeState) XXX_Size() int {
	return m.Size()
}
func (m *ProtocolFeeState) XXX_DiscardUnknown() {
	xxx_messageInfo_ProtocolFeeState.DiscardUnknown(m)
}

var xxx_messageInfo_ProtocolFeeState proto.InternalMessageInfo

func (m *ProtocolFeeState) GetTotalCollected() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.TotalCollected
	}
	return nil
}

func (m *ProtocolFeeState) GetValidatorRewards() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.ValidatorRewards
	}
	return nil
}

func (m *ProtocolFeeState) GetCommunityPool() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.CommunityPool
	}
	return nil
}

type ModuleBalance struct {
	ModuleName	string						`protobuf:"bytes,1,opt,name=module_name,json=moduleName,proto3" json:"module_name,omitempty"`
	Address		string						`protobuf:"bytes,2,opt,name=address,proto3" json:"address,omitempty"`
	Balance		github_com_cosmos_cosmos_sdk_types.Coins	`protobuf:"bytes,3,rep,name=balance,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"balance"`
}

func (m *ModuleBalance) Reset()		{ *m = ModuleBalance{} }
func (m *ModuleBalance) String() string	{ return proto.CompactTextString(m) }
func (*ModuleBalance) ProtoMessage()	{}
func (*ModuleBalance) Descriptor() ([]byte, []int) {
	return fileDescriptor_e8fec93f5a3aedc0, []int{3}
}
func (m *ModuleBalance) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ModuleBalance) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ModuleBalance.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *ModuleBalance) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ModuleBalance.Merge(m, src)
}
func (m *ModuleBalance) XXX_Size() int {
	return m.Size()
}
func (m *ModuleBalance) XXX_DiscardUnknown() {
	xxx_messageInfo_ModuleBalance.DiscardUnknown(m)
}

var xxx_messageInfo_ModuleBalance proto.InternalMessageInfo

func (m *ModuleBalance) GetModuleName() string {
	if m != nil {
		return m.ModuleName
	}
	return ""
}

func (m *ModuleBalance) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func (m *ModuleBalance) GetBalance() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.Balance
	}
	return nil
}

func init() {
	proto.RegisterType((*GenesisState)(nil), "l1.fees.v1.GenesisState")
	proto.RegisterType((*Params)(nil), "l1.fees.v1.Params")
	proto.RegisterType((*ProtocolFeeState)(nil), "l1.fees.v1.ProtocolFeeState")
	proto.RegisterType((*ModuleBalance)(nil), "l1.fees.v1.ModuleBalance")
}

func init()	{ proto.RegisterFile("l1/fees/v1/genesis.proto", fileDescriptor_e8fec93f5a3aedc0) }

var fileDescriptor_e8fec93f5a3aedc0 = []byte{

	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xb4, 0x96, 0xcd, 0x6e, 0xdb, 0x46,
	0x10, 0xc7, 0xad, 0x38, 0x75, 0xe2, 0x95, 0x65, 0x2b, 0x1b, 0x7f, 0xd0, 0x8e, 0x23, 0xbb, 0xee,
	0x97, 0x0f, 0x89, 0x64, 0xb9, 0x1f, 0x68, 0x51, 0xa0, 0x85, 0xa5, 0xc0, 0x2e, 0x0a, 0xb4, 0x10,
	0x68, 0x15, 0x01, 0x7a, 0x21, 0x56, 0xe4, 0x98, 0x22, 0x4c, 0x72, 0x89, 0xdd, 0x95, 0x45, 0xf7,
	0x29, 0xfa, 0x18, 0x45, 0x6f, 0x3d, 0xf6, 0x0d, 0x72, 0xcc, 0xb1, 0xa7, 0xb6, 0xb0, 0xdf, 0xa1,
	0xe7, 0x62, 0x67, 0x97, 0x12, 0x25, 0xa7, 0xb7, 0xf4, 0x64, 0x79, 0xe7, 0xf7, 0x9f, 0x8f, 0xdd,
	0xd9, 0x1d, 0x12, 0x27, 0x6e, 0xb7, 0x2e, 0x00, 0x64, 0xeb, 0xaa, 0xdd, 0x0a, 0x21, 0x05, 0x19,
	0xc9, 0x66, 0x26, 0xb8, 0xe2, 0x94, 0xc4, 0xed, 0xa6, 0xb6, 0x34, 0xaf, 0xda, 0x3b, 0x0d, 0x9f,
	0xcb, 0x84, 0xcb, 0xd6, 0x80, 0x49, 0x68, 0x5d, 0xb5, 0x07, 0xa0, 0x58, 0xbb, 0xe5, 0xf3, 0x28,
	0x35, 0xec, 0xce, 0x7a, 0xc8, 0x43, 0x8e, 0x3f, 0x5b, 0xfa, 0x97, 0x59, 0x3d, 0xf8, 0xbd, 0x42,
	0x56, 0xce, 0x8c, 0xcf, 0x73, 0xc5, 0x14, 0xd0, 0x23, 0xb2, 0x94, 0x31, 0xc1, 0x12, 0xe9, 0x54,
	0xf6, 0x2b, 0x87, 0xd5, 0x63, 0xda, 0x9c, 0xc6, 0x68, 0xf6, 0xd0, 0xd2, 0xb9, 0xff, 0xea, 0xcf,
	0xbd, 0x05, 0xd7, 0x72, 0xb4, 0x47, 0x28, 0xfa, 0xf2, 0x79, 0xec, 0x5d, 0x00, 0x78, 0x52, 0xfb,
	0x71, 0xee, 0xa1, 0x7a, 0x77, 0x46, 0x6d, 0xa9, 0x53, 0x00, 0x8c, 0x65, 0xfd, 0xd4, 0xb3, 0xb9,
	0x75, 0xfa, 0x01, 0x59, 0xf5, 0x79, 0x1a, 0x82, 0x54, 0x11, 0x4f, 0xbd, 0x41, 0x26, 0x9d, 0xc5,
	0xfd, 0xca, 0x61, 0xcd, 0xad, 0x4d, 0x57, 0x3b, 0x99, 0x3c, 0xf8, 0xa5, 0x4a, 0x96, 0x4c, 0x46,
	0xf4, 0x19, 0xa1, 0x2c, 0x8e, 0xf9, 0x18, 0x02, 0x4c, 0x21, 0x80, 0x94, 0x63, 0x05, 0x8b, 0x87,
	0xcb, 0x6e, 0xdd, 0x5a, 0x4e, 0x01, 0x5e, 0xe0, 0x3a, 0xfd, 0x8c, 0x6c, 0x5d, 0xb1, 0x38, 0x0a,
	0x98, 0xe2, 0xc2, 0x13, 0x30, 0x66, 0x22, 0x90, 0x9e, 0x60, 0x2a, 0xe2, 0x98, 0xf6, 0xb2, 0xbb,
	0x31, 0x31, 0xbb, 0xc6, 0xea, 0x6a, 0x23, 0x3d, 0x22, 0xeb, 0x3e, 0x4f, 0x92, 0x51, 0x1a, 0xa9,
	0x6b, 0x2f, 0xe3, 0x3c, 0xb6, 0xa2, 0x45, 0x14, 0xd1, 0x89, 0xad, 0xc7, 0x79, 0x6c, 0x14, 0xef,
	0x93, 0xd5, 0x24, 0x4a, 0x31, 0x27, 0x96, 0xf0, 0x51, 0xaa, 0x9c, 0xfb, 0xc8, 0xae, 0x24, 0x51,
	0x7a, 0x0a, 0x70, 0x82, 0x6b, 0xda, 0xaf, 0x26, 0x7c, 0x1e, 0xc7, 0xe0, 0xeb, 0x9c, 0x12, 0x1e,
	0x8c, 0x62, 0x70, 0xde, 0x31, 0x7e, 0x2f, 0x00, 0xba, 0x85, 0xe9, 0x3b, 0xb4, 0xd0, 0xcf, 0x89,
	0x73, 0xb7, 0x02, 0xc5, 0x44, 0x08, 0xca, 0x59, 0x42, 0xd5, 0xe6, 0x7c, 0x09, 0x7d, 0xb4, 0xd2,
	0x63, 0xb2, 0x31, 0x57, 0x83, 0x95, 0x3d, 0x40, 0xd9, 0xe3, 0x99, 0x22, 0xac, 0xe6, 0x43, 0xb2,
	0xa6, 0xbb, 0xaa, 0x5c, 0xc6, 0x43, 0xa4, 0x6b, 0x7a, 0x79, 0x5a, 0x87, 0xae, 0x96, 0xe5, 0x65,
	0x6c, 0xd9, 0x56, 0xcb, 0xf2, 0x29, 0xf5, 0x35, 0xd9, 0x35, 0x21, 0xbd, 0x41, 0xcc, 0xfd, 0x4b,
	0x6f, 0xa4, 0xa2, 0x38, 0xfa, 0x89, 0x4d, 0xce, 0x9a, 0xe0, 0x59, 0x6f, 0x1b, 0xa6, 0xa3, 0x91,
	0x1f, 0xa6, 0x44, 0x27, 0x93, 0xba, 0xf8, 0x52, 0x7b, 0xa8, 0xa1, 0x00, 0x39, 0xe4, 0x71, 0x80,
	0xe2, 0x2a, 0x8a, 0x37, 0xa7, 0xf6, 0x7e, 0x61, 0xd6, 0xca, 0x5d, 0x42, 0x74, 0x82, 0x2a, 0xf7,
	0x42, 0x26, 0x9d, 0x95, 0xfd, 0xca, 0xe1, 0x7d, 0xf7, 0x61, 0xc2, 0xf2, 0x7e, 0x7e, 0xc6, 0x24,
	0x3d, 0x20, 0x35, 0x6d, 0x35, 0x59, 0x69, 0xa0, 0x86, 0x40, 0x35, 0x61, 0x39, 0xa6, 0x71, 0x87,
	0x51, 0xb9, 0x74, 0x56, 0x67, 0x99, 0x7e, 0xae, 0xdb, 0xcb, 0xd1, 0x8c, 0x84, 0x34, 0x00, 0xa1,
	0x21, 0x2f, 0x03, 0x61, 0x24, 0xce, 0x1a, 0xe2, 0xeb, 0x09, 0xcb, 0xcf, 0xd1, 0xdc, 0xcf, 0x65,
	0x0f, 0x04, 0x4a, 0x69, 0x87, 0x34, 0xa4, 0x62, 0x97, 0xa0, 0xf3, 0xc3, 0x9e, 0x65, 0xa9, 0xaf,
	0xaf, 0x13, 0x64, 0xc5, 0x76, 0xd6, 0x71, 0x3b, 0x77, 0x90, 0xea, 0xe7, 0x27, 0x05, 0x73, 0xae,
	0x20, 0xb3, 0x9b, 0xfb, 0x2d, 0x79, 0xef, 0xbf, 0x62, 0x7b, 0xe3, 0x48, 0x0d, 0x3d, 0xd4, 0x3a,
	0x8f, 0x30, 0x8d, 0xa7, 0x6f, 0x4a, 0xe3, 0x65, 0xa4, 0x86, 0xe7, 0x1a, 0xa2, 0x9f, 0x92, 0x2d,
	0x7d, 0x94, 0x99, 0x88, 0xb8, 0xd0, 0xdd, 0x32, 0x86, 0x28, 0x1c, 0x2a, 0xdc, 0x66, 0x8a, 0xdb,
	0xac, 0xbb, 0xb6, 0x67, 0xad, 0x2f, 0xd1, 0xa8, 0x37, 0xf9, 0x0b, 0xb2, 0x6d, 0xca, 0x78, 0x93,
	0xf0, 0xb1, 0x39, 0x1f, 0x04, 0xee, 0x4a, 0xbf, 0x24, 0x3b, 0xb6, 0x35, 0x94, 0x60, 0xa9, 0xbc,
	0x00, 0x51, 0x6e, 0xa6, 0x75, 0xac, 0x7e, 0xcb, 0x10, 0x7d, 0x0b, 0xcc, 0xf4, 0x55, 0xa9, 0x2d,
	0xe4, 0x48, 0xf8, 0x43, 0x4d, 0x7a, 0x3e, 0xcb, 0x30, 0xf4, 0x86, 0xe9, 0xab, 0x29, 0x73, 0x5e,
	0x20, 0x5d, 0x96, 0xe9, 0xe8, 0xdf, 0x90, 0x77, 0x63, 0x3e, 0xf6, 0x04, 0x64, 0x23, 0x65, 0xda,
	0x31, 0x13, 0x90, 0x44, 0xa3, 0x04, 0x5d, 0xd8, 0x24, 0x36, 0x31, 0x89, 0xa7, 0x31, 0x1f, 0xbb,
	0x13, 0xae, 0x67, 0xb0, 0x2e, 0x2b, 0x4e, 0xa1, 0x4b, 0x1a, 0x25, 0x2f, 0x41, 0x24, 0x7d, 0xbd,
	0x5c, 0x76, 0xb3, 0x85, 0x6e, 0x9e, 0x4c, 0xa9, 0x17, 0x16, 0x9a, 0x3a, 0xf9, 0x8a, 0xec, 0x4a,
	0xc5, 0x05, 0x0b, 0xc1, 0x13, 0x90, 0x2a, 0x6f, 0x70, 0xad, 0x66, 0xae, 0xa0, 0x83, 0x2e, 0x1c,
	0xcb, 0xb8, 0x90, 0xaa, 0xce, 0xb5, 0x2a, 0xdd, 0xc6, 0x4f, 0xc8, 0xd6, 0xe4, 0xd6, 0xea, 0x26,
	0x08, 0x99, 0x2c, 0xa4, 0xdb, 0xe6, 0xae, 0xdb, 0xdb, 0xdb, 0x03, 0x71, 0xc6, 0xa4, 0x55, 0xe9,
	0xbb, 0x3e, 0x17, 0x68, 0xc7, 0xde, 0xf5, 0x19, 0xef, 0xcf, 0x08, 0x4d, 0x40, 0x4a, 0x9d, 0x5d,
	0x09, 0x7d, 0x82, 0x68, 0xdd, 0x5a, 0x26, 0xf4, 0xc1, 0x3f, 0xf7, 0x48, 0x7d, 0xfe, 0xf9, 0xa7,
	0x8a, 0xac, 0x29, 0xae, 0x58, 0x5c, 0x3c, 0x7c, 0x10, 0xe0, 0x8b, 0x5d, 0x3d, 0xde, 0x6e, 0x9a,
	0x59, 0xd6, 0xd4, 0x09, 0x36, 0xed, 0x2c, 0x6b, 0x76, 0x79, 0x94, 0x76, 0x8e, 0xf4, 0xc8, 0xf8,
	0xf5, 0xaf, 0xbd, 0xc3, 0x30, 0x52, 0xc3, 0xd1, 0xa0, 0xe9, 0xf3, 0xa4, 0x65, 0x07, 0x9f, 0xf9,
	0xf3, 0x5c, 0x06, 0x97, 0x2d, 0x75, 0x9d, 0x81, 0x44, 0x81, 0x74, 0x57, 0x31, 0x46, 0xb7, 0x08,
	0x41, 0x73, 0xf2, 0xe8, 0xce, 0xd3, 0xe9, 0xdc, 0x7b, 0xfb, 0x71, 0xeb, 0xf3, 0x0f, 0x30, 0x15,
	0x7a, 0xac, 0x95, 0x9f, 0x5e, 0x67, 0xf1, 0xed, 0x87, 0xad, 0xcd, 0x3c, 0xe0, 0x07, 0xbf, 0x55,
	0x48, 0xcd, 0xcc, 0x8c, 0x0e, 0x8b, 0xf5, 0x5b, 0x41, 0xf7, 0x48, 0xd5, 0x8c, 0x17, 0x2f, 0x65,
	0x09, 0xe0, 0x94, 0x5f, 0x76, 0x89, 0x59, 0xfa, 0x9e, 0x25, 0x40, 0x1d, 0xf2, 0x80, 0x05, 0x81,
	0x00, 0x29, 0xed, 0x34, 0x2c, 0xfe, 0xa5, 0x40, 0x1e, 0x0c, 0x8c, 0x97, 0xff, 0x23, 0xf3, 0xc2,
	0x77, 0xe7, 0xe4, 0xd5, 0x4d, 0xa3, 0xf2, 0xfa, 0xa6, 0x51, 0xf9, 0xfb, 0xa6, 0x51, 0xf9, 0xf9,
	0xb6, 0xb1, 0xf0, 0xfa, 0xb6, 0xb1, 0xf0, 0xc7, 0x6d, 0x63, 0xe1, 0xc7, 0x8f, 0x4a, 0xce, 0x24,
	0xbf, 0x02, 0x01, 0x51, 0x98, 0x3e, 0x8f, 0xdb, 0xad, 0xb8, 0xdd, 0xca, 0xcd, 0x37, 0x12, 0x7a,
	0x1c, 0x2c, 0xe1, 0x37, 0xc5, 0xc7, 0xff, 0x06, 0x00, 0x00, 0xff, 0xff, 0x5a, 0x11, 0x51, 0x7d,
	0x3b, 0x09, 0x00, 0x00,
}

func (m *GenesisState) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GenesisState) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *GenesisState) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.CongestionBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.CongestionBps))
		i--
		dAtA[i] = 0x18
	}
	{
		size, err := m.ProtocolFeeState.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	{
		size, err := m.Params.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *Params) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Params) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Params) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.MessageFeeAmount) > 0 {
		i -= len(m.MessageFeeAmount)
		copy(dAtA[i:], m.MessageFeeAmount)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.MessageFeeAmount)))
		i--
		dAtA[i] = 0x1
		i--
		dAtA[i] = 0xda
	}
	if len(m.ByteFeeAmount) > 0 {
		i -= len(m.ByteFeeAmount)
		copy(dAtA[i:], m.ByteFeeAmount)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.ByteFeeAmount)))
		i--
		dAtA[i] = 0x1
		i--
		dAtA[i] = 0xd2
	}
	if len(m.BaseFeePerGasAmount) > 0 {
		i -= len(m.BaseFeePerGasAmount)
		copy(dAtA[i:], m.BaseFeePerGasAmount)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.BaseFeePerGasAmount)))
		i--
		dAtA[i] = 0x1
		i--
		dAtA[i] = 0xca
	}
	if len(m.StorageRentByteFeeAmount) > 0 {
		i -= len(m.StorageRentByteFeeAmount)
		copy(dAtA[i:], m.StorageRentByteFeeAmount)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.StorageRentByteFeeAmount)))
		i--
		dAtA[i] = 0x1
		i--
		dAtA[i] = 0xc2
	}
	if len(m.ReputationDiscountCapAmount) > 0 {
		i -= len(m.ReputationDiscountCapAmount)
		copy(dAtA[i:], m.ReputationDiscountCapAmount)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.ReputationDiscountCapAmount)))
		i--
		dAtA[i] = 0x1
		i--
		dAtA[i] = 0xba
	}
	if len(m.LowReputationPremiumCapAmount) > 0 {
		i -= len(m.LowReputationPremiumCapAmount)
		copy(dAtA[i:], m.LowReputationPremiumCapAmount)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.LowReputationPremiumCapAmount)))
		i--
		dAtA[i] = 0x1
		i--
		dAtA[i] = 0xb2
	}
	if m.CongestionSurchargeCapBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.CongestionSurchargeCapBps))
		i--
		dAtA[i] = 0x1
		i--
		dAtA[i] = 0xa8
	}
	if len(m.TargetTransferFeeAmount) > 0 {
		i -= len(m.TargetTransferFeeAmount)
		copy(dAtA[i:], m.TargetTransferFeeAmount)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.TargetTransferFeeAmount)))
		i--
		dAtA[i] = 0x1
		i--
		dAtA[i] = 0xa2
	}
	if m.StakePriorityWeightBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.StakePriorityWeightBps))
		i--
		dAtA[i] = 0x1
		i--
		dAtA[i] = 0x98
	}
	if m.FeePriorityWeightBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.FeePriorityWeightBps))
		i--
		dAtA[i] = 0x1
		i--
		dAtA[i] = 0x90
	}
	if m.MaxSenderTxsPerBlockWithStake != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.MaxSenderTxsPerBlockWithStake))
		i--
		dAtA[i] = 0x1
		i--
		dAtA[i] = 0x88
	}
	if len(m.StakeTxAllowanceStepAmount) > 0 {
		i -= len(m.StakeTxAllowanceStepAmount)
		copy(dAtA[i:], m.StakeTxAllowanceStepAmount)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.StakeTxAllowanceStepAmount)))
		i--
		dAtA[i] = 0x1
		i--
		dAtA[i] = 0x82
	}
	if m.MaxSenderTxsPerBlock != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.MaxSenderTxsPerBlock))
		i--
		dAtA[i] = 0x78
	}
	if m.MaxBlockTxs != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.MaxBlockTxs))
		i--
		dAtA[i] = 0x70
	}
	if m.MaxBlockGas != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.MaxBlockGas))
		i--
		dAtA[i] = 0x68
	}
	if m.MaxTxGas != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.MaxTxGas))
		i--
		dAtA[i] = 0x60
	}
	if m.CongestionThresholdBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.CongestionThresholdBps))
		i--
		dAtA[i] = 0x58
	}
	if m.TargetBlockUtilizationBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.TargetBlockUtilizationBps))
		i--
		dAtA[i] = 0x50
	}
	if len(m.MaxFeeAmount) > 0 {
		i -= len(m.MaxFeeAmount)
		copy(dAtA[i:], m.MaxFeeAmount)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.MaxFeeAmount)))
		i--
		dAtA[i] = 0x4a
	}
	if len(m.BaseFeeAmount) > 0 {
		i -= len(m.BaseFeeAmount)
		copy(dAtA[i:], m.BaseFeeAmount)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.BaseFeeAmount)))
		i--
		dAtA[i] = 0x42
	}
	if len(m.CommunityPoolTarget) > 0 {
		i -= len(m.CommunityPoolTarget)
		copy(dAtA[i:], m.CommunityPoolTarget)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.CommunityPoolTarget)))
		i--
		dAtA[i] = 0x3a
	}
	if len(m.ValidatorRewardsTarget) > 0 {
		i -= len(m.ValidatorRewardsTarget)
		copy(dAtA[i:], m.ValidatorRewardsTarget)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.ValidatorRewardsTarget)))
		i--
		dAtA[i] = 0x32
	}
	if len(m.FeeCollectorModule) > 0 {
		i -= len(m.FeeCollectorModule)
		copy(dAtA[i:], m.FeeCollectorModule)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.FeeCollectorModule)))
		i--
		dAtA[i] = 0x2a
	}
	if len(m.MinFeeAmount) > 0 {
		i -= len(m.MinFeeAmount)
		copy(dAtA[i:], m.MinFeeAmount)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.MinFeeAmount)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.CommunityPoolRatio) > 0 {
		i -= len(m.CommunityPoolRatio)
		copy(dAtA[i:], m.CommunityPoolRatio)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.CommunityPoolRatio)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.ValidatorRewardsRatio) > 0 {
		i -= len(m.ValidatorRewardsRatio)
		copy(dAtA[i:], m.ValidatorRewardsRatio)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.ValidatorRewardsRatio)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.AllowedFeeDenoms) > 0 {
		for iNdEx := len(m.AllowedFeeDenoms) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.AllowedFeeDenoms[iNdEx])
			copy(dAtA[i:], m.AllowedFeeDenoms[iNdEx])
			i = encodeVarintGenesis(dAtA, i, uint64(len(m.AllowedFeeDenoms[iNdEx])))
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *ProtocolFeeState) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ProtocolFeeState) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *ProtocolFeeState) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.CommunityPool) > 0 {
		for iNdEx := len(m.CommunityPool) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.CommunityPool[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x1a
		}
	}
	if len(m.ValidatorRewards) > 0 {
		for iNdEx := len(m.ValidatorRewards) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.ValidatorRewards[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
		}
	}
	if len(m.TotalCollected) > 0 {
		for iNdEx := len(m.TotalCollected) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.TotalCollected[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *ModuleBalance) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ModuleBalance) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *ModuleBalance) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Balance) > 0 {
		for iNdEx := len(m.Balance) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Balance[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x1a
		}
	}
	if len(m.Address) > 0 {
		i -= len(m.Address)
		copy(dAtA[i:], m.Address)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.Address)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.ModuleName) > 0 {
		i -= len(m.ModuleName)
		copy(dAtA[i:], m.ModuleName)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.ModuleName)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintGenesis(dAtA []byte, offset int, v uint64) int {
	offset -= sovGenesis(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *GenesisState) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Params.Size()
	n += 1 + l + sovGenesis(uint64(l))
	l = m.ProtocolFeeState.Size()
	n += 1 + l + sovGenesis(uint64(l))
	if m.CongestionBps != 0 {
		n += 1 + sovGenesis(uint64(m.CongestionBps))
	}
	return n
}

func (m *Params) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.AllowedFeeDenoms) > 0 {
		for _, s := range m.AllowedFeeDenoms {
			l = len(s)
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	l = len(m.ValidatorRewardsRatio)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	l = len(m.CommunityPoolRatio)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	l = len(m.MinFeeAmount)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	l = len(m.FeeCollectorModule)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	l = len(m.ValidatorRewardsTarget)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	l = len(m.CommunityPoolTarget)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	l = len(m.BaseFeeAmount)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	l = len(m.MaxFeeAmount)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	if m.TargetBlockUtilizationBps != 0 {
		n += 1 + sovGenesis(uint64(m.TargetBlockUtilizationBps))
	}
	if m.CongestionThresholdBps != 0 {
		n += 1 + sovGenesis(uint64(m.CongestionThresholdBps))
	}
	if m.MaxTxGas != 0 {
		n += 1 + sovGenesis(uint64(m.MaxTxGas))
	}
	if m.MaxBlockGas != 0 {
		n += 1 + sovGenesis(uint64(m.MaxBlockGas))
	}
	if m.MaxBlockTxs != 0 {
		n += 1 + sovGenesis(uint64(m.MaxBlockTxs))
	}
	if m.MaxSenderTxsPerBlock != 0 {
		n += 1 + sovGenesis(uint64(m.MaxSenderTxsPerBlock))
	}
	l = len(m.StakeTxAllowanceStepAmount)
	if l > 0 {
		n += 2 + l + sovGenesis(uint64(l))
	}
	if m.MaxSenderTxsPerBlockWithStake != 0 {
		n += 2 + sovGenesis(uint64(m.MaxSenderTxsPerBlockWithStake))
	}
	if m.FeePriorityWeightBps != 0 {
		n += 2 + sovGenesis(uint64(m.FeePriorityWeightBps))
	}
	if m.StakePriorityWeightBps != 0 {
		n += 2 + sovGenesis(uint64(m.StakePriorityWeightBps))
	}
	l = len(m.TargetTransferFeeAmount)
	if l > 0 {
		n += 2 + l + sovGenesis(uint64(l))
	}
	if m.CongestionSurchargeCapBps != 0 {
		n += 2 + sovGenesis(uint64(m.CongestionSurchargeCapBps))
	}
	l = len(m.LowReputationPremiumCapAmount)
	if l > 0 {
		n += 2 + l + sovGenesis(uint64(l))
	}
	l = len(m.ReputationDiscountCapAmount)
	if l > 0 {
		n += 2 + l + sovGenesis(uint64(l))
	}
	l = len(m.StorageRentByteFeeAmount)
	if l > 0 {
		n += 2 + l + sovGenesis(uint64(l))
	}
	l = len(m.BaseFeePerGasAmount)
	if l > 0 {
		n += 2 + l + sovGenesis(uint64(l))
	}
	l = len(m.ByteFeeAmount)
	if l > 0 {
		n += 2 + l + sovGenesis(uint64(l))
	}
	l = len(m.MessageFeeAmount)
	if l > 0 {
		n += 2 + l + sovGenesis(uint64(l))
	}
	return n
}

func (m *ProtocolFeeState) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.TotalCollected) > 0 {
		for _, e := range m.TotalCollected {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.ValidatorRewards) > 0 {
		for _, e := range m.ValidatorRewards {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.CommunityPool) > 0 {
		for _, e := range m.CommunityPool {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	return n
}

func (m *ModuleBalance) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ModuleName)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	l = len(m.Address)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	if len(m.Balance) > 0 {
		for _, e := range m.Balance {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	return n
}

func sovGenesis(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozGenesis(x uint64) (n int) {
	return sovGenesis(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *GenesisState) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: GenesisState: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GenesisState: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Params", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Params.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ProtocolFeeState", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.ProtocolFeeState.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field CongestionBps", wireType)
			}
			m.CongestionBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.CongestionBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *Params) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Params: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Params: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AllowedFeeDenoms", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.AllowedFeeDenoms = append(m.AllowedFeeDenoms, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ValidatorRewardsRatio", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ValidatorRewardsRatio = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field CommunityPoolRatio", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.CommunityPoolRatio = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MinFeeAmount", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.MinFeeAmount = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field FeeCollectorModule", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.FeeCollectorModule = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ValidatorRewardsTarget", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ValidatorRewardsTarget = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field CommunityPoolTarget", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.CommunityPoolTarget = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 8:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BaseFeeAmount", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.BaseFeeAmount = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 9:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxFeeAmount", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.MaxFeeAmount = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 10:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field TargetBlockUtilizationBps", wireType)
			}
			m.TargetBlockUtilizationBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.TargetBlockUtilizationBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 11:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field CongestionThresholdBps", wireType)
			}
			m.CongestionThresholdBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.CongestionThresholdBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 12:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxTxGas", wireType)
			}
			m.MaxTxGas = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxTxGas |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 13:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxBlockGas", wireType)
			}
			m.MaxBlockGas = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxBlockGas |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 14:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxBlockTxs", wireType)
			}
			m.MaxBlockTxs = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxBlockTxs |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 15:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxSenderTxsPerBlock", wireType)
			}
			m.MaxSenderTxsPerBlock = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxSenderTxsPerBlock |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 16:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field StakeTxAllowanceStepAmount", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.StakeTxAllowanceStepAmount = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 17:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxSenderTxsPerBlockWithStake", wireType)
			}
			m.MaxSenderTxsPerBlockWithStake = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxSenderTxsPerBlockWithStake |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 18:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field FeePriorityWeightBps", wireType)
			}
			m.FeePriorityWeightBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.FeePriorityWeightBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 19:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field StakePriorityWeightBps", wireType)
			}
			m.StakePriorityWeightBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.StakePriorityWeightBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 20:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TargetTransferFeeAmount", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.TargetTransferFeeAmount = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 21:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field CongestionSurchargeCapBps", wireType)
			}
			m.CongestionSurchargeCapBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.CongestionSurchargeCapBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 22:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field LowReputationPremiumCapAmount", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.LowReputationPremiumCapAmount = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 23:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ReputationDiscountCapAmount", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ReputationDiscountCapAmount = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 24:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field StorageRentByteFeeAmount", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.StorageRentByteFeeAmount = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 25:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BaseFeePerGasAmount", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.BaseFeePerGasAmount = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 26:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ByteFeeAmount", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ByteFeeAmount = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 27:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MessageFeeAmount", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.MessageFeeAmount = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *ProtocolFeeState) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: ProtocolFeeState: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ProtocolFeeState: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TotalCollected", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.TotalCollected = append(m.TotalCollected, types.Coin{})
			if err := m.TotalCollected[len(m.TotalCollected)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ValidatorRewards", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ValidatorRewards = append(m.ValidatorRewards, types.Coin{})
			if err := m.ValidatorRewards[len(m.ValidatorRewards)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field CommunityPool", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.CommunityPool = append(m.CommunityPool, types.Coin{})
			if err := m.CommunityPool[len(m.CommunityPool)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *ModuleBalance) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: ModuleBalance: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ModuleBalance: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ModuleName", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ModuleName = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Address", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Address = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Balance", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Balance = append(m.Balance, types.Coin{})
			if err := m.Balance[len(m.Balance)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipGenesis(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowGenesis
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthGenesis
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupGenesis
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthGenesis
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthGenesis		= fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowGenesis		= fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupGenesis	= fmt.Errorf("proto: unexpected end of group")
)
