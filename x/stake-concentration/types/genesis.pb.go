package types

import (
	fmt "fmt"
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
	Params	Params			`protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
	Network	NetworkConcentration	`protobuf:"bytes,2,opt,name=network,proto3" json:"network"`
}

func (m *GenesisState) Reset()		{ *m = GenesisState{} }
func (m *GenesisState) String() string	{ return proto.CompactTextString(m) }
func (*GenesisState) ProtoMessage()	{}
func (*GenesisState) Descriptor() ([]byte, []int) {
	return fileDescriptor_63cd6ab20ece2409, []int{0}
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

func (m *GenesisState) GetNetwork() NetworkConcentration {
	if m != nil {
		return m.Network
	}
	return NetworkConcentration{}
}

type Params struct {
	MaxVotingPowerBps		uint32	`protobuf:"varint,1,opt,name=max_voting_power_bps,json=maxVotingPowerBps,proto3" json:"max_voting_power_bps,omitempty"`
	SoftVotingPowerBps		uint32	`protobuf:"varint,2,opt,name=soft_voting_power_bps,json=softVotingPowerBps,proto3" json:"soft_voting_power_bps,omitempty"`
	MaxRewardReductionBps		uint32	`protobuf:"varint,3,opt,name=max_reward_reduction_bps,json=maxRewardReductionBps,proto3" json:"max_reward_reduction_bps,omitempty"`
	WarningThresholdBps		uint32	`protobuf:"varint,4,opt,name=warning_threshold_bps,json=warningThresholdBps,proto3" json:"warning_threshold_bps,omitempty"`
	DelegationRejectionEnabled	bool	`protobuf:"varint,5,opt,name=delegation_rejection_enabled,json=delegationRejectionEnabled,proto3" json:"delegation_rejection_enabled,omitempty"`
}

func (m *Params) Reset()		{ *m = Params{} }
func (m *Params) String() string	{ return proto.CompactTextString(m) }
func (*Params) ProtoMessage()		{}
func (*Params) Descriptor() ([]byte, []int) {
	return fileDescriptor_63cd6ab20ece2409, []int{1}
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

func (m *Params) GetMaxVotingPowerBps() uint32 {
	if m != nil {
		return m.MaxVotingPowerBps
	}
	return 0
}

func (m *Params) GetSoftVotingPowerBps() uint32 {
	if m != nil {
		return m.SoftVotingPowerBps
	}
	return 0
}

func (m *Params) GetMaxRewardReductionBps() uint32 {
	if m != nil {
		return m.MaxRewardReductionBps
	}
	return 0
}

func (m *Params) GetWarningThresholdBps() uint32 {
	if m != nil {
		return m.WarningThresholdBps
	}
	return 0
}

func (m *Params) GetDelegationRejectionEnabled() bool {
	if m != nil {
		return m.DelegationRejectionEnabled
	}
	return false
}

type ValidatorPower struct {
	OperatorAddress	string	`protobuf:"bytes,1,opt,name=operator_address,json=operatorAddress,proto3" json:"operator_address,omitempty"`
	VotingPower	uint64	`protobuf:"varint,2,opt,name=voting_power,json=votingPower,proto3" json:"voting_power,omitempty"`
}

func (m *ValidatorPower) Reset()		{ *m = ValidatorPower{} }
func (m *ValidatorPower) String() string	{ return proto.CompactTextString(m) }
func (*ValidatorPower) ProtoMessage()		{}
func (*ValidatorPower) Descriptor() ([]byte, []int) {
	return fileDescriptor_63cd6ab20ece2409, []int{2}
}
func (m *ValidatorPower) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ValidatorPower) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ValidatorPower.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *ValidatorPower) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ValidatorPower.Merge(m, src)
}
func (m *ValidatorPower) XXX_Size() int {
	return m.Size()
}
func (m *ValidatorPower) XXX_DiscardUnknown() {
	xxx_messageInfo_ValidatorPower.DiscardUnknown(m)
}

var xxx_messageInfo_ValidatorPower proto.InternalMessageInfo

func (m *ValidatorPower) GetOperatorAddress() string {
	if m != nil {
		return m.OperatorAddress
	}
	return ""
}

func (m *ValidatorPower) GetVotingPower() uint64 {
	if m != nil {
		return m.VotingPower
	}
	return 0
}

type ValidatorConcentration struct {
	OperatorAddress		string	`protobuf:"bytes,1,opt,name=operator_address,json=operatorAddress,proto3" json:"operator_address,omitempty"`
	VotingPower		uint64	`protobuf:"varint,2,opt,name=voting_power,json=votingPower,proto3" json:"voting_power,omitempty"`
	RawVotingPowerBps	uint32	`protobuf:"varint,3,opt,name=raw_voting_power_bps,json=rawVotingPowerBps,proto3" json:"raw_voting_power_bps,omitempty"`
	EffectiveVotingPowerBps	uint32	`protobuf:"varint,4,opt,name=effective_voting_power_bps,json=effectiveVotingPowerBps,proto3" json:"effective_voting_power_bps,omitempty"`
	AboveSoftCap		bool	`protobuf:"varint,5,opt,name=above_soft_cap,json=aboveSoftCap,proto3" json:"above_soft_cap,omitempty"`
	AboveHardCap		bool	`protobuf:"varint,6,opt,name=above_hard_cap,json=aboveHardCap,proto3" json:"above_hard_cap,omitempty"`
	DelegationAllowed	bool	`protobuf:"varint,7,opt,name=delegation_allowed,json=delegationAllowed,proto3" json:"delegation_allowed,omitempty"`
	RewardModifierBps	uint32	`protobuf:"varint,8,opt,name=reward_modifier_bps,json=rewardModifierBps,proto3" json:"reward_modifier_bps,omitempty"`
	Warning			string	`protobuf:"bytes,9,opt,name=warning,proto3" json:"warning,omitempty"`
}

func (m *ValidatorConcentration) Reset()		{ *m = ValidatorConcentration{} }
func (m *ValidatorConcentration) String() string	{ return proto.CompactTextString(m) }
func (*ValidatorConcentration) ProtoMessage()		{}
func (*ValidatorConcentration) Descriptor() ([]byte, []int) {
	return fileDescriptor_63cd6ab20ece2409, []int{3}
}
func (m *ValidatorConcentration) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ValidatorConcentration) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ValidatorConcentration.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *ValidatorConcentration) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ValidatorConcentration.Merge(m, src)
}
func (m *ValidatorConcentration) XXX_Size() int {
	return m.Size()
}
func (m *ValidatorConcentration) XXX_DiscardUnknown() {
	xxx_messageInfo_ValidatorConcentration.DiscardUnknown(m)
}

var xxx_messageInfo_ValidatorConcentration proto.InternalMessageInfo

func (m *ValidatorConcentration) GetOperatorAddress() string {
	if m != nil {
		return m.OperatorAddress
	}
	return ""
}

func (m *ValidatorConcentration) GetVotingPower() uint64 {
	if m != nil {
		return m.VotingPower
	}
	return 0
}

func (m *ValidatorConcentration) GetRawVotingPowerBps() uint32 {
	if m != nil {
		return m.RawVotingPowerBps
	}
	return 0
}

func (m *ValidatorConcentration) GetEffectiveVotingPowerBps() uint32 {
	if m != nil {
		return m.EffectiveVotingPowerBps
	}
	return 0
}

func (m *ValidatorConcentration) GetAboveSoftCap() bool {
	if m != nil {
		return m.AboveSoftCap
	}
	return false
}

func (m *ValidatorConcentration) GetAboveHardCap() bool {
	if m != nil {
		return m.AboveHardCap
	}
	return false
}

func (m *ValidatorConcentration) GetDelegationAllowed() bool {
	if m != nil {
		return m.DelegationAllowed
	}
	return false
}

func (m *ValidatorConcentration) GetRewardModifierBps() uint32 {
	if m != nil {
		return m.RewardModifierBps
	}
	return 0
}

func (m *ValidatorConcentration) GetWarning() string {
	if m != nil {
		return m.Warning
	}
	return ""
}

type NetworkConcentration struct {
	Epoch			uint64				`protobuf:"varint,1,opt,name=epoch,proto3" json:"epoch,omitempty"`
	TotalVotingPower	uint64				`protobuf:"varint,2,opt,name=total_voting_power,json=totalVotingPower,proto3" json:"total_voting_power,omitempty"`
	MaxValidatorPowerBps	uint32				`protobuf:"varint,3,opt,name=max_validator_power_bps,json=maxValidatorPowerBps,proto3" json:"max_validator_power_bps,omitempty"`
	TopThreePowerBps	uint32				`protobuf:"varint,4,opt,name=top_three_power_bps,json=topThreePowerBps,proto3" json:"top_three_power_bps,omitempty"`
	Validators		[]ValidatorConcentration	`protobuf:"bytes,5,rep,name=validators,proto3" json:"validators"`
	Warnings		[]string			`protobuf:"bytes,6,rep,name=warnings,proto3" json:"warnings,omitempty"`
	RecomputedHeight	int64				`protobuf:"varint,7,opt,name=recomputed_height,json=recomputedHeight,proto3" json:"recomputed_height,omitempty"`
}

func (m *NetworkConcentration) Reset()		{ *m = NetworkConcentration{} }
func (m *NetworkConcentration) String() string	{ return proto.CompactTextString(m) }
func (*NetworkConcentration) ProtoMessage()	{}
func (*NetworkConcentration) Descriptor() ([]byte, []int) {
	return fileDescriptor_63cd6ab20ece2409, []int{4}
}
func (m *NetworkConcentration) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *NetworkConcentration) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_NetworkConcentration.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *NetworkConcentration) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NetworkConcentration.Merge(m, src)
}
func (m *NetworkConcentration) XXX_Size() int {
	return m.Size()
}
func (m *NetworkConcentration) XXX_DiscardUnknown() {
	xxx_messageInfo_NetworkConcentration.DiscardUnknown(m)
}

var xxx_messageInfo_NetworkConcentration proto.InternalMessageInfo

func (m *NetworkConcentration) GetEpoch() uint64 {
	if m != nil {
		return m.Epoch
	}
	return 0
}

func (m *NetworkConcentration) GetTotalVotingPower() uint64 {
	if m != nil {
		return m.TotalVotingPower
	}
	return 0
}

func (m *NetworkConcentration) GetMaxValidatorPowerBps() uint32 {
	if m != nil {
		return m.MaxValidatorPowerBps
	}
	return 0
}

func (m *NetworkConcentration) GetTopThreePowerBps() uint32 {
	if m != nil {
		return m.TopThreePowerBps
	}
	return 0
}

func (m *NetworkConcentration) GetValidators() []ValidatorConcentration {
	if m != nil {
		return m.Validators
	}
	return nil
}

func (m *NetworkConcentration) GetWarnings() []string {
	if m != nil {
		return m.Warnings
	}
	return nil
}

func (m *NetworkConcentration) GetRecomputedHeight() int64 {
	if m != nil {
		return m.RecomputedHeight
	}
	return 0
}

func init() {
	proto.RegisterType((*GenesisState)(nil), "l1.stakeconcentration.v1.GenesisState")
	proto.RegisterType((*Params)(nil), "l1.stakeconcentration.v1.Params")
	proto.RegisterType((*ValidatorPower)(nil), "l1.stakeconcentration.v1.ValidatorPower")
	proto.RegisterType((*ValidatorConcentration)(nil), "l1.stakeconcentration.v1.ValidatorConcentration")
	proto.RegisterType((*NetworkConcentration)(nil), "l1.stakeconcentration.v1.NetworkConcentration")
}

func init() {
	proto.RegisterFile("l1/stakeconcentration/v1/genesis.proto", fileDescriptor_63cd6ab20ece2409)
}

var fileDescriptor_63cd6ab20ece2409 = []byte{

	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x54, 0x4d, 0x6f, 0xd3, 0x4a,
	0x14, 0xcd, 0x57, 0xd3, 0xf6, 0xb6, 0xaf, 0x2f, 0x99, 0xa6, 0xaf, 0x56, 0xf4, 0x94, 0x97, 0x17,
	0x21, 0x14, 0x04, 0xb1, 0x49, 0x11, 0xb0, 0x40, 0x42, 0xb4, 0x15, 0xa2, 0x1b, 0xaa, 0xca, 0x45,
	0x5d, 0xb0, 0xc0, 0x9a, 0xc4, 0x37, 0x8e, 0xa9, 0xe3, 0xb1, 0xc6, 0xd3, 0x38, 0xfc, 0x0b, 0xb6,
	0x08, 0x89, 0xdf, 0xd3, 0x65, 0x97, 0xac, 0x10, 0xb4, 0x7f, 0x04, 0xcd, 0x8c, 0x9d, 0xa4, 0x4d,
	0xbb, 0x63, 0x67, 0xcf, 0x39, 0xe7, 0xde, 0x99, 0x73, 0xee, 0x0c, 0xdc, 0x0f, 0xba, 0x56, 0x2c,
	0xe8, 0x29, 0xf6, 0x59, 0xd8, 0xc7, 0x50, 0x70, 0x2a, 0x7c, 0x16, 0x5a, 0xe3, 0xae, 0xe5, 0x61,
	0x88, 0xb1, 0x1f, 0x9b, 0x11, 0x67, 0x82, 0x11, 0x23, 0xe8, 0x9a, 0x8b, 0x3c, 0x73, 0xdc, 0xad,
	0xd7, 0x3c, 0xe6, 0x31, 0x45, 0xb2, 0xe4, 0x97, 0xe6, 0xb7, 0xbe, 0xe5, 0x61, 0xfd, 0x8d, 0xae,
	0x70, 0x2c, 0xa8, 0x40, 0xf2, 0x12, 0xca, 0x11, 0xe5, 0x74, 0x14, 0x1b, 0xf9, 0x66, 0xbe, 0xbd,
	0xb6, 0xd3, 0x34, 0xef, 0xaa, 0x68, 0x1e, 0x29, 0xde, 0x5e, 0xe9, 0xfc, 0xc7, 0x7f, 0x39, 0x3b,
	0x55, 0x91, 0x43, 0x58, 0x0e, 0x51, 0x24, 0x8c, 0x9f, 0x1a, 0x05, 0x55, 0xc0, 0xbc, 0xbb, 0xc0,
	0xa1, 0x26, 0xee, 0xcf, 0xaf, 0xa7, 0xe5, 0xb2, 0x22, 0xad, 0xaf, 0x05, 0x28, 0xeb, 0x46, 0xc4,
	0x82, 0xda, 0x88, 0x4e, 0x9c, 0x31, 0x13, 0x7e, 0xe8, 0x39, 0x11, 0x4b, 0x90, 0x3b, 0xbd, 0x48,
	0x6f, 0xf4, 0x2f, 0xbb, 0x3a, 0xa2, 0x93, 0x13, 0x05, 0x1d, 0x49, 0x64, 0x2f, 0x8a, 0x49, 0x17,
	0xb6, 0x62, 0x36, 0x10, 0x8b, 0x8a, 0x82, 0x52, 0x10, 0x09, 0xde, 0x90, 0x3c, 0x07, 0x43, 0xf6,
	0xe0, 0x98, 0x50, 0xee, 0x3a, 0x1c, 0xdd, 0xb3, 0xbe, 0xdc, 0x95, 0x52, 0x15, 0x95, 0x6a, 0x6b,
	0x44, 0x27, 0xb6, 0x82, 0xed, 0x0c, 0x95, 0xc2, 0x1d, 0xd8, 0x4a, 0x28, 0x0f, 0x65, 0x1f, 0x31,
	0xe4, 0x18, 0x0f, 0x59, 0xe0, 0x2a, 0x55, 0x49, 0xa9, 0x36, 0x53, 0xf0, 0x5d, 0x86, 0x49, 0xcd,
	0x2b, 0xf8, 0xd7, 0xc5, 0x00, 0x3d, 0x75, 0x70, 0x87, 0xe3, 0x47, 0xd4, 0xcd, 0x30, 0xa4, 0xbd,
	0x00, 0x5d, 0x63, 0xa9, 0x99, 0x6f, 0xaf, 0xd8, 0xf5, 0x19, 0xc7, 0xce, 0x28, 0xaf, 0x35, 0xa3,
	0xf5, 0x01, 0x36, 0x4e, 0x68, 0xe0, 0xbb, 0x54, 0x30, 0xae, 0xce, 0x40, 0x1e, 0x40, 0x85, 0x45,
	0xc8, 0xe5, 0x82, 0x43, 0x5d, 0x97, 0x63, 0xac, 0x0d, 0x5a, 0xb5, 0xff, 0xce, 0xd6, 0x77, 0xf5,
	0x32, 0xf9, 0x1f, 0xd6, 0xe7, 0x9d, 0x51, 0xae, 0x94, 0xec, 0xb5, 0xf1, 0xcc, 0x91, 0xd6, 0x97,
	0x22, 0xfc, 0x33, 0x6d, 0x70, 0x2d, 0xa7, 0x3f, 0xdb, 0x48, 0x66, 0xcb, 0x69, 0xb2, 0x98, 0x94,
	0xf6, 0xbc, 0xca, 0x69, 0x72, 0x23, 0xa8, 0x17, 0x50, 0xc7, 0xc1, 0x40, 0xba, 0x31, 0xc6, 0x45,
	0x99, 0x36, 0x7d, 0x7b, 0xca, 0xb8, 0x21, 0xbe, 0x07, 0x1b, 0xb4, 0xc7, 0xc6, 0xe8, 0xa8, 0xf1,
	0xe8, 0xd3, 0x28, 0xb5, 0x7a, 0x5d, 0xad, 0x1e, 0xb3, 0x81, 0xd8, 0xa7, 0xd1, 0x8c, 0x35, 0x94,
	0xb3, 0x20, 0x59, 0xe5, 0x39, 0xd6, 0x01, 0xe5, 0xae, 0x64, 0x75, 0x80, 0xcc, 0x85, 0x48, 0x83,
	0x80, 0x25, 0xe8, 0x1a, 0xcb, 0x8a, 0x59, 0x9d, 0x21, 0xbb, 0x1a, 0x20, 0x26, 0x6c, 0xa6, 0xc3,
	0x35, 0x62, 0xae, 0x3f, 0xf0, 0xd3, 0x0d, 0xaf, 0xa4, 0xe7, 0x54, 0xd0, 0xdb, 0x14, 0x91, 0x5b,
	0x35, 0x60, 0x39, 0x1d, 0x1d, 0x63, 0x55, 0xb9, 0x9b, 0xfd, 0xb6, 0x7e, 0x15, 0xa0, 0x76, 0xdb,
	0x0d, 0x22, 0x35, 0x58, 0xc2, 0x88, 0xf5, 0x87, 0x2a, 0x8e, 0x92, 0xad, 0x7f, 0xc8, 0x23, 0x20,
	0x82, 0x09, 0x1a, 0x38, 0xb7, 0x44, 0x51, 0x51, 0xc8, 0x9c, 0x49, 0xe4, 0x29, 0x6c, 0xab, 0xbb,
	0x96, 0x65, 0xbf, 0x10, 0x89, 0xbc, 0x8a, 0xd7, 0x47, 0x4f, 0xee, 0xb6, 0x03, 0x9b, 0x82, 0x45,
	0xea, 0x06, 0xe0, 0x42, 0x1c, 0x15, 0xc1, 0x22, 0x39, 0xff, 0x38, 0xa5, 0x9f, 0x00, 0x4c, 0x3b,
	0xc4, 0xc6, 0x52, 0xb3, 0xd8, 0x5e, 0xdb, 0x79, 0x7c, 0xf7, 0x7b, 0x71, 0xfb, 0x24, 0xa6, 0x2f,
	0xc6, 0x5c, 0x25, 0x52, 0x87, 0x95, 0xd4, 0xa5, 0xd8, 0x28, 0x37, 0x8b, 0xed, 0x55, 0x7b, 0xfa,
	0x4f, 0x1e, 0x42, 0x95, 0x63, 0x9f, 0x8d, 0xa2, 0x33, 0x81, 0xae, 0x33, 0x44, 0xdf, 0x1b, 0x0a,
	0x15, 0x57, 0xd1, 0xae, 0xcc, 0x80, 0x03, 0xb5, 0xbe, 0x77, 0x74, 0x7e, 0xd9, 0xc8, 0x5f, 0x5c,
	0x36, 0xf2, 0x3f, 0x2f, 0x1b, 0xf9, 0xcf, 0x57, 0x8d, 0xdc, 0xc5, 0x55, 0x23, 0xf7, 0xfd, 0xaa,
	0x91, 0x7b, 0xff, 0xcc, 0xf3, 0xc5, 0xf0, 0xac, 0x67, 0xf6, 0xd9, 0xc8, 0x8a, 0xd9, 0x18, 0x39,
	0xfa, 0x5e, 0xd8, 0x09, 0xba, 0x56, 0xd0, 0xb5, 0x26, 0xfa, 0xa9, 0xee, 0x5c, 0x7f, 0xab, 0xc5,
	0xa7, 0x08, 0xe3, 0x5e, 0x59, 0xbd, 0xbb, 0x4f, 0x7e, 0x07, 0x00, 0x00, 0xff, 0xff, 0x75, 0x09,
	0xbc, 0xf7, 0xd1, 0x05, 0x00, 0x00,
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
	{
		size, err := m.Network.MarshalToSizedBuffer(dAtA[:i])
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
	if m.DelegationRejectionEnabled {
		i--
		if m.DelegationRejectionEnabled {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x28
	}
	if m.WarningThresholdBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.WarningThresholdBps))
		i--
		dAtA[i] = 0x20
	}
	if m.MaxRewardReductionBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.MaxRewardReductionBps))
		i--
		dAtA[i] = 0x18
	}
	if m.SoftVotingPowerBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.SoftVotingPowerBps))
		i--
		dAtA[i] = 0x10
	}
	if m.MaxVotingPowerBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.MaxVotingPowerBps))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *ValidatorPower) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ValidatorPower) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *ValidatorPower) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.VotingPower != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.VotingPower))
		i--
		dAtA[i] = 0x10
	}
	if len(m.OperatorAddress) > 0 {
		i -= len(m.OperatorAddress)
		copy(dAtA[i:], m.OperatorAddress)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.OperatorAddress)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *ValidatorConcentration) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ValidatorConcentration) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *ValidatorConcentration) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Warning) > 0 {
		i -= len(m.Warning)
		copy(dAtA[i:], m.Warning)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.Warning)))
		i--
		dAtA[i] = 0x4a
	}
	if m.RewardModifierBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.RewardModifierBps))
		i--
		dAtA[i] = 0x40
	}
	if m.DelegationAllowed {
		i--
		if m.DelegationAllowed {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x38
	}
	if m.AboveHardCap {
		i--
		if m.AboveHardCap {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x30
	}
	if m.AboveSoftCap {
		i--
		if m.AboveSoftCap {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x28
	}
	if m.EffectiveVotingPowerBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.EffectiveVotingPowerBps))
		i--
		dAtA[i] = 0x20
	}
	if m.RawVotingPowerBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.RawVotingPowerBps))
		i--
		dAtA[i] = 0x18
	}
	if m.VotingPower != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.VotingPower))
		i--
		dAtA[i] = 0x10
	}
	if len(m.OperatorAddress) > 0 {
		i -= len(m.OperatorAddress)
		copy(dAtA[i:], m.OperatorAddress)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.OperatorAddress)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *NetworkConcentration) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *NetworkConcentration) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *NetworkConcentration) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.RecomputedHeight != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.RecomputedHeight))
		i--
		dAtA[i] = 0x38
	}
	if len(m.Warnings) > 0 {
		for iNdEx := len(m.Warnings) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.Warnings[iNdEx])
			copy(dAtA[i:], m.Warnings[iNdEx])
			i = encodeVarintGenesis(dAtA, i, uint64(len(m.Warnings[iNdEx])))
			i--
			dAtA[i] = 0x32
		}
	}
	if len(m.Validators) > 0 {
		for iNdEx := len(m.Validators) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Validators[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x2a
		}
	}
	if m.TopThreePowerBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.TopThreePowerBps))
		i--
		dAtA[i] = 0x20
	}
	if m.MaxValidatorPowerBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.MaxValidatorPowerBps))
		i--
		dAtA[i] = 0x18
	}
	if m.TotalVotingPower != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.TotalVotingPower))
		i--
		dAtA[i] = 0x10
	}
	if m.Epoch != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.Epoch))
		i--
		dAtA[i] = 0x8
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
	l = m.Network.Size()
	n += 1 + l + sovGenesis(uint64(l))
	return n
}

func (m *Params) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.MaxVotingPowerBps != 0 {
		n += 1 + sovGenesis(uint64(m.MaxVotingPowerBps))
	}
	if m.SoftVotingPowerBps != 0 {
		n += 1 + sovGenesis(uint64(m.SoftVotingPowerBps))
	}
	if m.MaxRewardReductionBps != 0 {
		n += 1 + sovGenesis(uint64(m.MaxRewardReductionBps))
	}
	if m.WarningThresholdBps != 0 {
		n += 1 + sovGenesis(uint64(m.WarningThresholdBps))
	}
	if m.DelegationRejectionEnabled {
		n += 2
	}
	return n
}

func (m *ValidatorPower) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.OperatorAddress)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	if m.VotingPower != 0 {
		n += 1 + sovGenesis(uint64(m.VotingPower))
	}
	return n
}

func (m *ValidatorConcentration) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.OperatorAddress)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	if m.VotingPower != 0 {
		n += 1 + sovGenesis(uint64(m.VotingPower))
	}
	if m.RawVotingPowerBps != 0 {
		n += 1 + sovGenesis(uint64(m.RawVotingPowerBps))
	}
	if m.EffectiveVotingPowerBps != 0 {
		n += 1 + sovGenesis(uint64(m.EffectiveVotingPowerBps))
	}
	if m.AboveSoftCap {
		n += 2
	}
	if m.AboveHardCap {
		n += 2
	}
	if m.DelegationAllowed {
		n += 2
	}
	if m.RewardModifierBps != 0 {
		n += 1 + sovGenesis(uint64(m.RewardModifierBps))
	}
	l = len(m.Warning)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	return n
}

func (m *NetworkConcentration) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Epoch != 0 {
		n += 1 + sovGenesis(uint64(m.Epoch))
	}
	if m.TotalVotingPower != 0 {
		n += 1 + sovGenesis(uint64(m.TotalVotingPower))
	}
	if m.MaxValidatorPowerBps != 0 {
		n += 1 + sovGenesis(uint64(m.MaxValidatorPowerBps))
	}
	if m.TopThreePowerBps != 0 {
		n += 1 + sovGenesis(uint64(m.TopThreePowerBps))
	}
	if len(m.Validators) > 0 {
		for _, e := range m.Validators {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.Warnings) > 0 {
		for _, s := range m.Warnings {
			l = len(s)
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if m.RecomputedHeight != 0 {
		n += 1 + sovGenesis(uint64(m.RecomputedHeight))
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
				return fmt.Errorf("proto: wrong wireType = %d for field Network", wireType)
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
			if err := m.Network.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxVotingPowerBps", wireType)
			}
			m.MaxVotingPowerBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxVotingPowerBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field SoftVotingPowerBps", wireType)
			}
			m.SoftVotingPowerBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.SoftVotingPowerBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxRewardReductionBps", wireType)
			}
			m.MaxRewardReductionBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxRewardReductionBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field WarningThresholdBps", wireType)
			}
			m.WarningThresholdBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.WarningThresholdBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field DelegationRejectionEnabled", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.DelegationRejectionEnabled = bool(v != 0)
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
func (m *ValidatorPower) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: ValidatorPower: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ValidatorPower: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field OperatorAddress", wireType)
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
			m.OperatorAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field VotingPower", wireType)
			}
			m.VotingPower = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.VotingPower |= uint64(b&0x7F) << shift
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
func (m *ValidatorConcentration) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: ValidatorConcentration: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ValidatorConcentration: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field OperatorAddress", wireType)
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
			m.OperatorAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field VotingPower", wireType)
			}
			m.VotingPower = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.VotingPower |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field RawVotingPowerBps", wireType)
			}
			m.RawVotingPowerBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.RawVotingPowerBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field EffectiveVotingPowerBps", wireType)
			}
			m.EffectiveVotingPowerBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.EffectiveVotingPowerBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field AboveSoftCap", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.AboveSoftCap = bool(v != 0)
		case 6:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field AboveHardCap", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.AboveHardCap = bool(v != 0)
		case 7:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field DelegationAllowed", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.DelegationAllowed = bool(v != 0)
		case 8:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field RewardModifierBps", wireType)
			}
			m.RewardModifierBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.RewardModifierBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 9:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Warning", wireType)
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
			m.Warning = string(dAtA[iNdEx:postIndex])
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
func (m *NetworkConcentration) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: NetworkConcentration: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: NetworkConcentration: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Epoch", wireType)
			}
			m.Epoch = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Epoch |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field TotalVotingPower", wireType)
			}
			m.TotalVotingPower = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.TotalVotingPower |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxValidatorPowerBps", wireType)
			}
			m.MaxValidatorPowerBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxValidatorPowerBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field TopThreePowerBps", wireType)
			}
			m.TopThreePowerBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.TopThreePowerBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Validators", wireType)
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
			m.Validators = append(m.Validators, ValidatorConcentration{})
			if err := m.Validators[len(m.Validators)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Warnings", wireType)
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
			m.Warnings = append(m.Warnings, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		case 7:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field RecomputedHeight", wireType)
			}
			m.RecomputedHeight = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.RecomputedHeight |= int64(b&0x7F) << shift
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
