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
	Params			Params				`protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
	Commissions		[]ValidatorCommission		`protobuf:"bytes,2,rep,name=commissions,proto3" json:"commissions"`
	CommissionHistory	[]CommissionHistoryEntry	`protobuf:"bytes,3,rep,name=commission_history,json=commissionHistory,proto3" json:"commission_history"`
}

func (m *GenesisState) Reset()		{ *m = GenesisState{} }
func (m *GenesisState) String() string	{ return proto.CompactTextString(m) }
func (*GenesisState) ProtoMessage()	{}
func (*GenesisState) Descriptor() ([]byte, []int) {
	return fileDescriptor_f455ad5e736ca72e, []int{0}
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

func (m *GenesisState) GetCommissions() []ValidatorCommission {
	if m != nil {
		return m.Commissions
	}
	return nil
}

func (m *GenesisState) GetCommissionHistory() []CommissionHistoryEntry {
	if m != nil {
		return m.CommissionHistory
	}
	return nil
}

type Params struct {
	CommissionFloorBps		uint32	`protobuf:"varint,1,opt,name=commission_floor_bps,json=commissionFloorBps,proto3" json:"commission_floor_bps,omitempty"`
	CommissionCeilingBps		uint32	`protobuf:"varint,2,opt,name=commission_ceiling_bps,json=commissionCeilingBps,proto3" json:"commission_ceiling_bps,omitempty"`
	MaxRateChangeBps		uint32	`protobuf:"varint,3,opt,name=max_rate_change_bps,json=maxRateChangeBps,proto3" json:"max_rate_change_bps,omitempty"`
	HighPerformanceThresholdBps	uint32	`protobuf:"varint,4,opt,name=high_performance_threshold_bps,json=highPerformanceThresholdBps,proto3" json:"high_performance_threshold_bps,omitempty"`
	LowPerformanceThresholdBps	uint32	`protobuf:"varint,5,opt,name=low_performance_threshold_bps,json=lowPerformanceThresholdBps,proto3" json:"low_performance_threshold_bps,omitempty"`
	HighReputationThresholdBps	uint32	`protobuf:"varint,6,opt,name=high_reputation_threshold_bps,json=highReputationThresholdBps,proto3" json:"high_reputation_threshold_bps,omitempty"`
	LowReputationThresholdBps	uint32	`protobuf:"varint,7,opt,name=low_reputation_threshold_bps,json=lowReputationThresholdBps,proto3" json:"low_reputation_threshold_bps,omitempty"`
	PerformanceBonusBps		uint32	`protobuf:"varint,8,opt,name=performance_bonus_bps,json=performanceBonusBps,proto3" json:"performance_bonus_bps,omitempty"`
	PerformancePenaltyBps		uint32	`protobuf:"varint,9,opt,name=performance_penalty_bps,json=performancePenaltyBps,proto3" json:"performance_penalty_bps,omitempty"`
	ReputationBonusBps		uint32	`protobuf:"varint,10,opt,name=reputation_bonus_bps,json=reputationBonusBps,proto3" json:"reputation_bonus_bps,omitempty"`
	ReputationPenaltyBps		uint32	`protobuf:"varint,11,opt,name=reputation_penalty_bps,json=reputationPenaltyBps,proto3" json:"reputation_penalty_bps,omitempty"`
}

func (m *Params) Reset()		{ *m = Params{} }
func (m *Params) String() string	{ return proto.CompactTextString(m) }
func (*Params) ProtoMessage()		{}
func (*Params) Descriptor() ([]byte, []int) {
	return fileDescriptor_f455ad5e736ca72e, []int{1}
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

func (m *Params) GetCommissionFloorBps() uint32 {
	if m != nil {
		return m.CommissionFloorBps
	}
	return 0
}

func (m *Params) GetCommissionCeilingBps() uint32 {
	if m != nil {
		return m.CommissionCeilingBps
	}
	return 0
}

func (m *Params) GetMaxRateChangeBps() uint32 {
	if m != nil {
		return m.MaxRateChangeBps
	}
	return 0
}

func (m *Params) GetHighPerformanceThresholdBps() uint32 {
	if m != nil {
		return m.HighPerformanceThresholdBps
	}
	return 0
}

func (m *Params) GetLowPerformanceThresholdBps() uint32 {
	if m != nil {
		return m.LowPerformanceThresholdBps
	}
	return 0
}

func (m *Params) GetHighReputationThresholdBps() uint32 {
	if m != nil {
		return m.HighReputationThresholdBps
	}
	return 0
}

func (m *Params) GetLowReputationThresholdBps() uint32 {
	if m != nil {
		return m.LowReputationThresholdBps
	}
	return 0
}

func (m *Params) GetPerformanceBonusBps() uint32 {
	if m != nil {
		return m.PerformanceBonusBps
	}
	return 0
}

func (m *Params) GetPerformancePenaltyBps() uint32 {
	if m != nil {
		return m.PerformancePenaltyBps
	}
	return 0
}

func (m *Params) GetReputationBonusBps() uint32 {
	if m != nil {
		return m.ReputationBonusBps
	}
	return 0
}

func (m *Params) GetReputationPenaltyBps() uint32 {
	if m != nil {
		return m.ReputationPenaltyBps
	}
	return 0
}

type ValidatorCommission struct {
	ValidatorAddress	string	`protobuf:"bytes,1,opt,name=validator_address,json=validatorAddress,proto3" json:"validator_address,omitempty"`
	BaseCommissionBps	uint32	`protobuf:"varint,2,opt,name=base_commission_bps,json=baseCommissionBps,proto3" json:"base_commission_bps,omitempty"`
	EffectiveCommissionBps	uint32	`protobuf:"varint,3,opt,name=effective_commission_bps,json=effectiveCommissionBps,proto3" json:"effective_commission_bps,omitempty"`
	PerformanceModifierBps	int32	`protobuf:"varint,4,opt,name=performance_modifier_bps,json=performanceModifierBps,proto3" json:"performance_modifier_bps,omitempty"`
	ReputationModifierBps	int32	`protobuf:"varint,5,opt,name=reputation_modifier_bps,json=reputationModifierBps,proto3" json:"reputation_modifier_bps,omitempty"`
	CommissionFloorBps	uint32	`protobuf:"varint,6,opt,name=commission_floor_bps,json=commissionFloorBps,proto3" json:"commission_floor_bps,omitempty"`
	CommissionCeilingBps	uint32	`protobuf:"varint,7,opt,name=commission_ceiling_bps,json=commissionCeilingBps,proto3" json:"commission_ceiling_bps,omitempty"`
	LastUpdateHeight	uint64	`protobuf:"varint,8,opt,name=last_update_height,json=lastUpdateHeight,proto3" json:"last_update_height,omitempty"`
	Jailed			bool	`protobuf:"varint,9,opt,name=jailed,proto3" json:"jailed,omitempty"`
}

func (m *ValidatorCommission) Reset()		{ *m = ValidatorCommission{} }
func (m *ValidatorCommission) String() string	{ return proto.CompactTextString(m) }
func (*ValidatorCommission) ProtoMessage()	{}
func (*ValidatorCommission) Descriptor() ([]byte, []int) {
	return fileDescriptor_f455ad5e736ca72e, []int{2}
}
func (m *ValidatorCommission) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ValidatorCommission) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ValidatorCommission.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *ValidatorCommission) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ValidatorCommission.Merge(m, src)
}
func (m *ValidatorCommission) XXX_Size() int {
	return m.Size()
}
func (m *ValidatorCommission) XXX_DiscardUnknown() {
	xxx_messageInfo_ValidatorCommission.DiscardUnknown(m)
}

var xxx_messageInfo_ValidatorCommission proto.InternalMessageInfo

func (m *ValidatorCommission) GetValidatorAddress() string {
	if m != nil {
		return m.ValidatorAddress
	}
	return ""
}

func (m *ValidatorCommission) GetBaseCommissionBps() uint32 {
	if m != nil {
		return m.BaseCommissionBps
	}
	return 0
}

func (m *ValidatorCommission) GetEffectiveCommissionBps() uint32 {
	if m != nil {
		return m.EffectiveCommissionBps
	}
	return 0
}

func (m *ValidatorCommission) GetPerformanceModifierBps() int32 {
	if m != nil {
		return m.PerformanceModifierBps
	}
	return 0
}

func (m *ValidatorCommission) GetReputationModifierBps() int32 {
	if m != nil {
		return m.ReputationModifierBps
	}
	return 0
}

func (m *ValidatorCommission) GetCommissionFloorBps() uint32 {
	if m != nil {
		return m.CommissionFloorBps
	}
	return 0
}

func (m *ValidatorCommission) GetCommissionCeilingBps() uint32 {
	if m != nil {
		return m.CommissionCeilingBps
	}
	return 0
}

func (m *ValidatorCommission) GetLastUpdateHeight() uint64 {
	if m != nil {
		return m.LastUpdateHeight
	}
	return 0
}

func (m *ValidatorCommission) GetJailed() bool {
	if m != nil {
		return m.Jailed
	}
	return false
}

type CommissionHistoryEntry struct {
	ValidatorAddress	string	`protobuf:"bytes,1,opt,name=validator_address,json=validatorAddress,proto3" json:"validator_address,omitempty"`
	Height			uint64	`protobuf:"varint,2,opt,name=height,proto3" json:"height,omitempty"`
	BaseCommissionBps	uint32	`protobuf:"varint,3,opt,name=base_commission_bps,json=baseCommissionBps,proto3" json:"base_commission_bps,omitempty"`
	EffectiveCommissionBps	uint32	`protobuf:"varint,4,opt,name=effective_commission_bps,json=effectiveCommissionBps,proto3" json:"effective_commission_bps,omitempty"`
	PerformanceModifierBps	int32	`protobuf:"varint,5,opt,name=performance_modifier_bps,json=performanceModifierBps,proto3" json:"performance_modifier_bps,omitempty"`
	ReputationModifierBps	int32	`protobuf:"varint,6,opt,name=reputation_modifier_bps,json=reputationModifierBps,proto3" json:"reputation_modifier_bps,omitempty"`
	Jailed			bool	`protobuf:"varint,7,opt,name=jailed,proto3" json:"jailed,omitempty"`
	Reason			string	`protobuf:"bytes,8,opt,name=reason,proto3" json:"reason,omitempty"`
}

func (m *CommissionHistoryEntry) Reset()		{ *m = CommissionHistoryEntry{} }
func (m *CommissionHistoryEntry) String() string	{ return proto.CompactTextString(m) }
func (*CommissionHistoryEntry) ProtoMessage()		{}
func (*CommissionHistoryEntry) Descriptor() ([]byte, []int) {
	return fileDescriptor_f455ad5e736ca72e, []int{3}
}
func (m *CommissionHistoryEntry) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *CommissionHistoryEntry) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_CommissionHistoryEntry.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *CommissionHistoryEntry) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CommissionHistoryEntry.Merge(m, src)
}
func (m *CommissionHistoryEntry) XXX_Size() int {
	return m.Size()
}
func (m *CommissionHistoryEntry) XXX_DiscardUnknown() {
	xxx_messageInfo_CommissionHistoryEntry.DiscardUnknown(m)
}

var xxx_messageInfo_CommissionHistoryEntry proto.InternalMessageInfo

func (m *CommissionHistoryEntry) GetValidatorAddress() string {
	if m != nil {
		return m.ValidatorAddress
	}
	return ""
}

func (m *CommissionHistoryEntry) GetHeight() uint64 {
	if m != nil {
		return m.Height
	}
	return 0
}

func (m *CommissionHistoryEntry) GetBaseCommissionBps() uint32 {
	if m != nil {
		return m.BaseCommissionBps
	}
	return 0
}

func (m *CommissionHistoryEntry) GetEffectiveCommissionBps() uint32 {
	if m != nil {
		return m.EffectiveCommissionBps
	}
	return 0
}

func (m *CommissionHistoryEntry) GetPerformanceModifierBps() int32 {
	if m != nil {
		return m.PerformanceModifierBps
	}
	return 0
}

func (m *CommissionHistoryEntry) GetReputationModifierBps() int32 {
	if m != nil {
		return m.ReputationModifierBps
	}
	return 0
}

func (m *CommissionHistoryEntry) GetJailed() bool {
	if m != nil {
		return m.Jailed
	}
	return false
}

func (m *CommissionHistoryEntry) GetReason() string {
	if m != nil {
		return m.Reason
	}
	return ""
}

func init() {
	proto.RegisterType((*GenesisState)(nil), "l1.dynamiccommission.v1.GenesisState")
	proto.RegisterType((*Params)(nil), "l1.dynamiccommission.v1.Params")
	proto.RegisterType((*ValidatorCommission)(nil), "l1.dynamiccommission.v1.ValidatorCommission")
	proto.RegisterType((*CommissionHistoryEntry)(nil), "l1.dynamiccommission.v1.CommissionHistoryEntry")
}

func init() {
	proto.RegisterFile("l1/dynamiccommission/v1/genesis.proto", fileDescriptor_f455ad5e736ca72e)
}

var fileDescriptor_f455ad5e736ca72e = []byte{

	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x95, 0x4f, 0x6f, 0xd3, 0x3e,
	0x18, 0xc7, 0xfb, 0x37, 0xdb, 0xdc, 0xdf, 0x4f, 0xda, 0xd2, 0xad, 0x2b, 0x03, 0xba, 0xa9, 0x12,
	0xd2, 0x24, 0xb6, 0x66, 0x1d, 0x03, 0x71, 0x41, 0x68, 0xad, 0x80, 0x5d, 0x10, 0x53, 0x18, 0x1c,
	0xb8, 0x44, 0x6e, 0xe2, 0x26, 0x46, 0x4e, 0x1c, 0xd9, 0x6e, 0xb7, 0xbe, 0x04, 0x6e, 0xbc, 0xac,
	0x1d, 0x38, 0xec, 0xc8, 0x09, 0xa1, 0xed, 0xc4, 0xab, 0x00, 0xd9, 0xc9, 0x1a, 0x8f, 0x2d, 0x95,
	0xe8, 0xad, 0xf6, 0xf7, 0xf9, 0x7c, 0xfd, 0xc4, 0xcf, 0xf3, 0xd4, 0xe0, 0x11, 0xe9, 0x5a, 0xde,
	0x24, 0x82, 0x21, 0x76, 0x5d, 0x1a, 0x86, 0x98, 0x73, 0x4c, 0x23, 0x6b, 0xdc, 0xb5, 0x7c, 0x14,
	0x21, 0x8e, 0x79, 0x27, 0x66, 0x54, 0x50, 0x73, 0x9d, 0x74, 0x3b, 0xb7, 0xc2, 0x3a, 0xe3, 0xee,
	0xc6, 0xaa, 0x4f, 0x7d, 0xaa, 0x62, 0x2c, 0xf9, 0x2b, 0x09, 0x6f, 0x7f, 0x29, 0x81, 0xff, 0xde,
	0x24, 0x06, 0xef, 0x05, 0x14, 0xc8, 0x7c, 0x01, 0x8c, 0x18, 0x32, 0x18, 0xf2, 0x66, 0x71, 0xab,
	0xb8, 0x5d, 0xdb, 0xdf, 0xec, 0xe4, 0x18, 0x76, 0x8e, 0x55, 0x58, 0xaf, 0x72, 0xfe, 0x63, 0xb3,
	0x60, 0xa7, 0x90, 0x79, 0x02, 0x6a, 0x59, 0x14, 0x6f, 0x96, 0xb6, 0xca, 0xdb, 0xb5, 0xfd, 0x9d,
	0x5c, 0x8f, 0x8f, 0x90, 0x60, 0x0f, 0x0a, 0xca, 0xfa, 0xd3, 0xed, 0xd4, 0x50, 0xb7, 0x31, 0x3d,
	0x60, 0x66, 0x4b, 0x27, 0xc0, 0x5c, 0x50, 0x36, 0x69, 0x96, 0x95, 0xb9, 0x95, 0x6b, 0x9e, 0x79,
	0x1e, 0x25, 0xc4, 0xab, 0x48, 0xb0, 0x49, 0xea, 0xbf, 0xe2, 0xfe, 0xad, 0xb6, 0x7f, 0x57, 0x80,
	0x91, 0x7c, 0x94, 0xb9, 0x07, 0x56, 0xb5, 0x03, 0x87, 0x84, 0x52, 0xe6, 0x0c, 0xe2, 0xe4, 0x4e,
	0xfe, 0xb7, 0xb5, 0x64, 0x5e, 0x4b, 0xa9, 0x17, 0x73, 0xf3, 0x00, 0x34, 0x34, 0xc2, 0x45, 0x98,
	0xe0, 0xc8, 0x57, 0x4c, 0x49, 0x31, 0x9a, 0x5f, 0x3f, 0x11, 0x25, 0xb5, 0x0b, 0xea, 0x21, 0x3c,
	0x73, 0x18, 0x14, 0xc8, 0x71, 0x03, 0x18, 0xf9, 0x48, 0x21, 0x65, 0x85, 0x2c, 0x87, 0xf0, 0xcc,
	0x86, 0x02, 0xf5, 0x95, 0x20, 0xc3, 0xfb, 0xa0, 0x15, 0x60, 0x3f, 0x70, 0x62, 0xc4, 0x86, 0x94,
	0x85, 0x30, 0x72, 0x91, 0x23, 0x02, 0x86, 0x78, 0x40, 0x89, 0xa7, 0xc8, 0x8a, 0x22, 0xef, 0xcb,
	0xa8, 0xe3, 0x2c, 0xe8, 0xe4, 0x3a, 0x46, 0x9a, 0x1c, 0x82, 0x87, 0x84, 0x9e, 0xce, 0xf0, 0xa8,
	0x2a, 0x8f, 0x0d, 0x42, 0x4f, 0x67, 0x58, 0xa8, 0x3c, 0x18, 0x8a, 0x47, 0x02, 0x0a, 0xf9, 0xc5,
	0x37, 0x2d, 0x8c, 0xc4, 0x42, 0x06, 0xd9, 0xd3, 0x98, 0x1b, 0x16, 0x2f, 0xc1, 0x03, 0x99, 0x45,
	0xae, 0xc3, 0x82, 0x72, 0xb8, 0x47, 0xe8, 0x69, 0x8e, 0xc1, 0x3e, 0x58, 0xd3, 0x3f, 0x61, 0x40,
	0xa3, 0x11, 0x57, 0xe4, 0xa2, 0x22, 0xeb, 0x9a, 0xd8, 0x93, 0x9a, 0x64, 0x9e, 0x81, 0x75, 0x9d,
	0x89, 0x51, 0x04, 0x89, 0x98, 0x28, 0x6a, 0x49, 0x51, 0xba, 0xe5, 0x71, 0xa2, 0x4a, 0x6e, 0x0f,
	0xac, 0x6a, 0x89, 0x66, 0x47, 0x81, 0xa4, 0x1d, 0x32, 0x6d, 0x7a, 0xd2, 0x01, 0x68, 0x68, 0x84,
	0x7e, 0x50, 0x2d, 0x69, 0x87, 0x4c, 0xcd, 0xce, 0x69, 0x7f, 0x2b, 0x83, 0xfa, 0x1d, 0x23, 0x61,
	0x3e, 0x06, 0x2b, 0xe3, 0xeb, 0x6d, 0x07, 0x7a, 0x1e, 0x43, 0x3c, 0xe9, 0xc5, 0x25, 0x7b, 0x79,
	0x2a, 0x1c, 0x26, 0xfb, 0x66, 0x07, 0xd4, 0x07, 0x90, 0x23, 0x47, 0x6b, 0xc7, 0xac, 0x0d, 0x57,
	0xa4, 0xa4, 0x0d, 0x5b, 0xcc, 0xcd, 0xe7, 0xa0, 0x89, 0x86, 0x43, 0xe4, 0x0a, 0x3c, 0xbe, 0x05,
	0x25, 0x8d, 0xd8, 0x98, 0xea, 0xb7, 0x48, 0xfd, 0x3a, 0x43, 0xea, 0xe1, 0x21, 0x46, 0x6c, 0xda,
	0x88, 0x55, 0xbb, 0xa1, 0xe9, 0x6f, 0x53, 0x39, 0x2d, 0x84, 0x76, 0x3d, 0x37, 0xc0, 0xaa, 0x02,
	0xd7, 0x32, 0x59, 0xe7, 0xf2, 0xe6, 0xd2, 0x98, 0x63, 0x2e, 0x17, 0x66, 0xcc, 0xe5, 0x0e, 0x30,
	0x09, 0xe4, 0xc2, 0x19, 0xc5, 0x9e, 0x1c, 0xcd, 0x00, 0x61, 0x3f, 0x10, 0xaa, 0xb3, 0x2a, 0xf6,
	0xb2, 0x54, 0x3e, 0x28, 0xe1, 0x48, 0xed, 0x9b, 0x0d, 0x60, 0x7c, 0x86, 0x98, 0x20, 0x4f, 0x75,
	0xd1, 0xa2, 0x9d, 0xae, 0xda, 0xbf, 0x4a, 0xa0, 0x71, 0xf7, 0x9f, 0xd0, 0xbf, 0x55, 0xb4, 0x01,
	0x8c, 0x34, 0x83, 0x92, 0xca, 0x20, 0x5d, 0xe5, 0x55, 0xba, 0x3c, 0x4f, 0xa5, 0x2b, 0x73, 0x57,
	0xba, 0x3a, 0x6f, 0xa5, 0x8d, 0x59, 0x95, 0xce, 0xee, 0x74, 0x41, 0xbf, 0x53, 0xb9, 0xcf, 0x10,
	0xe4, 0x34, 0x52, 0xd5, 0x58, 0xb2, 0xd3, 0x55, 0xef, 0xdd, 0xf9, 0x65, 0xab, 0x78, 0x71, 0xd9,
	0x2a, 0xfe, 0xbc, 0x6c, 0x15, 0xbf, 0x5e, 0xb5, 0x0a, 0x17, 0x57, 0xad, 0xc2, 0xf7, 0xab, 0x56,
	0xe1, 0xd3, 0x53, 0x1f, 0x8b, 0x60, 0x34, 0xe8, 0xb8, 0x34, 0xb4, 0x38, 0x1d, 0x23, 0x86, 0xb0,
	0x1f, 0xed, 0x92, 0xae, 0x45, 0xba, 0xd6, 0xd9, 0xf5, 0x93, 0xba, 0xab, 0xbd, 0xa9, 0x62, 0x12,
	0x23, 0x3e, 0x30, 0xd4, 0x03, 0xf9, 0xe4, 0x4f, 0x00, 0x00, 0x00, 0xff, 0xff, 0x5b, 0x80, 0xaf,
	0x59, 0x78, 0x07, 0x00, 0x00,
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
	if len(m.CommissionHistory) > 0 {
		for iNdEx := len(m.CommissionHistory) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.CommissionHistory[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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
	if len(m.Commissions) > 0 {
		for iNdEx := len(m.Commissions) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Commissions[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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
	if m.ReputationPenaltyBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.ReputationPenaltyBps))
		i--
		dAtA[i] = 0x58
	}
	if m.ReputationBonusBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.ReputationBonusBps))
		i--
		dAtA[i] = 0x50
	}
	if m.PerformancePenaltyBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.PerformancePenaltyBps))
		i--
		dAtA[i] = 0x48
	}
	if m.PerformanceBonusBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.PerformanceBonusBps))
		i--
		dAtA[i] = 0x40
	}
	if m.LowReputationThresholdBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.LowReputationThresholdBps))
		i--
		dAtA[i] = 0x38
	}
	if m.HighReputationThresholdBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.HighReputationThresholdBps))
		i--
		dAtA[i] = 0x30
	}
	if m.LowPerformanceThresholdBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.LowPerformanceThresholdBps))
		i--
		dAtA[i] = 0x28
	}
	if m.HighPerformanceThresholdBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.HighPerformanceThresholdBps))
		i--
		dAtA[i] = 0x20
	}
	if m.MaxRateChangeBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.MaxRateChangeBps))
		i--
		dAtA[i] = 0x18
	}
	if m.CommissionCeilingBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.CommissionCeilingBps))
		i--
		dAtA[i] = 0x10
	}
	if m.CommissionFloorBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.CommissionFloorBps))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *ValidatorCommission) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ValidatorCommission) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *ValidatorCommission) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Jailed {
		i--
		if m.Jailed {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x48
	}
	if m.LastUpdateHeight != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.LastUpdateHeight))
		i--
		dAtA[i] = 0x40
	}
	if m.CommissionCeilingBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.CommissionCeilingBps))
		i--
		dAtA[i] = 0x38
	}
	if m.CommissionFloorBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.CommissionFloorBps))
		i--
		dAtA[i] = 0x30
	}
	if m.ReputationModifierBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.ReputationModifierBps))
		i--
		dAtA[i] = 0x28
	}
	if m.PerformanceModifierBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.PerformanceModifierBps))
		i--
		dAtA[i] = 0x20
	}
	if m.EffectiveCommissionBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.EffectiveCommissionBps))
		i--
		dAtA[i] = 0x18
	}
	if m.BaseCommissionBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.BaseCommissionBps))
		i--
		dAtA[i] = 0x10
	}
	if len(m.ValidatorAddress) > 0 {
		i -= len(m.ValidatorAddress)
		copy(dAtA[i:], m.ValidatorAddress)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.ValidatorAddress)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *CommissionHistoryEntry) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *CommissionHistoryEntry) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *CommissionHistoryEntry) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Reason) > 0 {
		i -= len(m.Reason)
		copy(dAtA[i:], m.Reason)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.Reason)))
		i--
		dAtA[i] = 0x42
	}
	if m.Jailed {
		i--
		if m.Jailed {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x38
	}
	if m.ReputationModifierBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.ReputationModifierBps))
		i--
		dAtA[i] = 0x30
	}
	if m.PerformanceModifierBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.PerformanceModifierBps))
		i--
		dAtA[i] = 0x28
	}
	if m.EffectiveCommissionBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.EffectiveCommissionBps))
		i--
		dAtA[i] = 0x20
	}
	if m.BaseCommissionBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.BaseCommissionBps))
		i--
		dAtA[i] = 0x18
	}
	if m.Height != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.Height))
		i--
		dAtA[i] = 0x10
	}
	if len(m.ValidatorAddress) > 0 {
		i -= len(m.ValidatorAddress)
		copy(dAtA[i:], m.ValidatorAddress)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.ValidatorAddress)))
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
	if len(m.Commissions) > 0 {
		for _, e := range m.Commissions {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.CommissionHistory) > 0 {
		for _, e := range m.CommissionHistory {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	return n
}

func (m *Params) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.CommissionFloorBps != 0 {
		n += 1 + sovGenesis(uint64(m.CommissionFloorBps))
	}
	if m.CommissionCeilingBps != 0 {
		n += 1 + sovGenesis(uint64(m.CommissionCeilingBps))
	}
	if m.MaxRateChangeBps != 0 {
		n += 1 + sovGenesis(uint64(m.MaxRateChangeBps))
	}
	if m.HighPerformanceThresholdBps != 0 {
		n += 1 + sovGenesis(uint64(m.HighPerformanceThresholdBps))
	}
	if m.LowPerformanceThresholdBps != 0 {
		n += 1 + sovGenesis(uint64(m.LowPerformanceThresholdBps))
	}
	if m.HighReputationThresholdBps != 0 {
		n += 1 + sovGenesis(uint64(m.HighReputationThresholdBps))
	}
	if m.LowReputationThresholdBps != 0 {
		n += 1 + sovGenesis(uint64(m.LowReputationThresholdBps))
	}
	if m.PerformanceBonusBps != 0 {
		n += 1 + sovGenesis(uint64(m.PerformanceBonusBps))
	}
	if m.PerformancePenaltyBps != 0 {
		n += 1 + sovGenesis(uint64(m.PerformancePenaltyBps))
	}
	if m.ReputationBonusBps != 0 {
		n += 1 + sovGenesis(uint64(m.ReputationBonusBps))
	}
	if m.ReputationPenaltyBps != 0 {
		n += 1 + sovGenesis(uint64(m.ReputationPenaltyBps))
	}
	return n
}

func (m *ValidatorCommission) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ValidatorAddress)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	if m.BaseCommissionBps != 0 {
		n += 1 + sovGenesis(uint64(m.BaseCommissionBps))
	}
	if m.EffectiveCommissionBps != 0 {
		n += 1 + sovGenesis(uint64(m.EffectiveCommissionBps))
	}
	if m.PerformanceModifierBps != 0 {
		n += 1 + sovGenesis(uint64(m.PerformanceModifierBps))
	}
	if m.ReputationModifierBps != 0 {
		n += 1 + sovGenesis(uint64(m.ReputationModifierBps))
	}
	if m.CommissionFloorBps != 0 {
		n += 1 + sovGenesis(uint64(m.CommissionFloorBps))
	}
	if m.CommissionCeilingBps != 0 {
		n += 1 + sovGenesis(uint64(m.CommissionCeilingBps))
	}
	if m.LastUpdateHeight != 0 {
		n += 1 + sovGenesis(uint64(m.LastUpdateHeight))
	}
	if m.Jailed {
		n += 2
	}
	return n
}

func (m *CommissionHistoryEntry) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ValidatorAddress)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	if m.Height != 0 {
		n += 1 + sovGenesis(uint64(m.Height))
	}
	if m.BaseCommissionBps != 0 {
		n += 1 + sovGenesis(uint64(m.BaseCommissionBps))
	}
	if m.EffectiveCommissionBps != 0 {
		n += 1 + sovGenesis(uint64(m.EffectiveCommissionBps))
	}
	if m.PerformanceModifierBps != 0 {
		n += 1 + sovGenesis(uint64(m.PerformanceModifierBps))
	}
	if m.ReputationModifierBps != 0 {
		n += 1 + sovGenesis(uint64(m.ReputationModifierBps))
	}
	if m.Jailed {
		n += 2
	}
	l = len(m.Reason)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
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
				return fmt.Errorf("proto: wrong wireType = %d for field Commissions", wireType)
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
			m.Commissions = append(m.Commissions, ValidatorCommission{})
			if err := m.Commissions[len(m.Commissions)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field CommissionHistory", wireType)
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
			m.CommissionHistory = append(m.CommissionHistory, CommissionHistoryEntry{})
			if err := m.CommissionHistory[len(m.CommissionHistory)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
				return fmt.Errorf("proto: wrong wireType = %d for field CommissionFloorBps", wireType)
			}
			m.CommissionFloorBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.CommissionFloorBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field CommissionCeilingBps", wireType)
			}
			m.CommissionCeilingBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.CommissionCeilingBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxRateChangeBps", wireType)
			}
			m.MaxRateChangeBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxRateChangeBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field HighPerformanceThresholdBps", wireType)
			}
			m.HighPerformanceThresholdBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.HighPerformanceThresholdBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field LowPerformanceThresholdBps", wireType)
			}
			m.LowPerformanceThresholdBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.LowPerformanceThresholdBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 6:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field HighReputationThresholdBps", wireType)
			}
			m.HighReputationThresholdBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.HighReputationThresholdBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 7:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field LowReputationThresholdBps", wireType)
			}
			m.LowReputationThresholdBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.LowReputationThresholdBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 8:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field PerformanceBonusBps", wireType)
			}
			m.PerformanceBonusBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.PerformanceBonusBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 9:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field PerformancePenaltyBps", wireType)
			}
			m.PerformancePenaltyBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.PerformancePenaltyBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 10:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ReputationBonusBps", wireType)
			}
			m.ReputationBonusBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ReputationBonusBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 11:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ReputationPenaltyBps", wireType)
			}
			m.ReputationPenaltyBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ReputationPenaltyBps |= uint32(b&0x7F) << shift
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
func (m *ValidatorCommission) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: ValidatorCommission: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ValidatorCommission: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ValidatorAddress", wireType)
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
			m.ValidatorAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field BaseCommissionBps", wireType)
			}
			m.BaseCommissionBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.BaseCommissionBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field EffectiveCommissionBps", wireType)
			}
			m.EffectiveCommissionBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.EffectiveCommissionBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field PerformanceModifierBps", wireType)
			}
			m.PerformanceModifierBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.PerformanceModifierBps |= int32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ReputationModifierBps", wireType)
			}
			m.ReputationModifierBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ReputationModifierBps |= int32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 6:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field CommissionFloorBps", wireType)
			}
			m.CommissionFloorBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.CommissionFloorBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 7:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field CommissionCeilingBps", wireType)
			}
			m.CommissionCeilingBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.CommissionCeilingBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 8:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field LastUpdateHeight", wireType)
			}
			m.LastUpdateHeight = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.LastUpdateHeight |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 9:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Jailed", wireType)
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
			m.Jailed = bool(v != 0)
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
func (m *CommissionHistoryEntry) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: CommissionHistoryEntry: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: CommissionHistoryEntry: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ValidatorAddress", wireType)
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
			m.ValidatorAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Height", wireType)
			}
			m.Height = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Height |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field BaseCommissionBps", wireType)
			}
			m.BaseCommissionBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.BaseCommissionBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field EffectiveCommissionBps", wireType)
			}
			m.EffectiveCommissionBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.EffectiveCommissionBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field PerformanceModifierBps", wireType)
			}
			m.PerformanceModifierBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.PerformanceModifierBps |= int32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 6:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ReputationModifierBps", wireType)
			}
			m.ReputationModifierBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ReputationModifierBps |= int32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 7:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Jailed", wireType)
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
			m.Jailed = bool(v != 0)
		case 8:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Reason", wireType)
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
			m.Reason = string(dAtA[iNdEx:postIndex])
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
