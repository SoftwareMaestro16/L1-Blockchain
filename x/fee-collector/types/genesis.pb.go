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
	Balances		FeeBalances		`protobuf:"bytes,2,opt,name=balances,proto3" json:"balances"`
	PendingDistribution	PendingDistribution	`protobuf:"bytes,3,opt,name=pending_distribution,json=pendingDistribution,proto3" json:"pending_distribution"`
	FeeHistory		[]FeeHistoryEntry	`protobuf:"bytes,4,rep,name=fee_history,json=feeHistory,proto3" json:"fee_history"`
}

func (m *GenesisState) Reset()		{ *m = GenesisState{} }
func (m *GenesisState) String() string	{ return proto.CompactTextString(m) }
func (*GenesisState) ProtoMessage()	{}
func (*GenesisState) Descriptor() ([]byte, []int) {
	return fileDescriptor_027cc99771428dd9, []int{0}
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

func (m *GenesisState) GetBalances() FeeBalances {
	if m != nil {
		return m.Balances
	}
	return FeeBalances{}
}

func (m *GenesisState) GetPendingDistribution() PendingDistribution {
	if m != nil {
		return m.PendingDistribution
	}
	return PendingDistribution{}
}

func (m *GenesisState) GetFeeHistory() []FeeHistoryEntry {
	if m != nil {
		return m.FeeHistory
	}
	return nil
}

type Params struct {
	BaseDenom		string	`protobuf:"bytes,1,opt,name=base_denom,json=baseDenom,proto3" json:"base_denom,omitempty"`
	TreasuryBps		uint32	`protobuf:"varint,2,opt,name=treasury_bps,json=treasuryBps,proto3" json:"treasury_bps,omitempty"`
	ProtectionBps		uint32	`protobuf:"varint,3,opt,name=protection_bps,json=protectionBps,proto3" json:"protection_bps,omitempty"`
	ValidatorsBps		uint32	`protobuf:"varint,4,opt,name=validators_bps,json=validatorsBps,proto3" json:"validators_bps,omitempty"`
	BurnBps			uint32	`protobuf:"varint,5,opt,name=burn_bps,json=burnBps,proto3" json:"burn_bps,omitempty"`
	CollectorModule		string	`protobuf:"bytes,6,opt,name=collector_module,json=collectorModule,proto3" json:"collector_module,omitempty"`
	TreasuryModule		string	`protobuf:"bytes,7,opt,name=treasury_module,json=treasuryModule,proto3" json:"treasury_module,omitempty"`
	ProtectionModule	string	`protobuf:"bytes,8,opt,name=protection_module,json=protectionModule,proto3" json:"protection_module,omitempty"`
	ValidatorsModule	string	`protobuf:"bytes,9,opt,name=validators_module,json=validatorsModule,proto3" json:"validators_module,omitempty"`
}

func (m *Params) Reset()		{ *m = Params{} }
func (m *Params) String() string	{ return proto.CompactTextString(m) }
func (*Params) ProtoMessage()		{}
func (*Params) Descriptor() ([]byte, []int) {
	return fileDescriptor_027cc99771428dd9, []int{1}
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

func (m *Params) GetBaseDenom() string {
	if m != nil {
		return m.BaseDenom
	}
	return ""
}

func (m *Params) GetTreasuryBps() uint32 {
	if m != nil {
		return m.TreasuryBps
	}
	return 0
}

func (m *Params) GetProtectionBps() uint32 {
	if m != nil {
		return m.ProtectionBps
	}
	return 0
}

func (m *Params) GetValidatorsBps() uint32 {
	if m != nil {
		return m.ValidatorsBps
	}
	return 0
}

func (m *Params) GetBurnBps() uint32 {
	if m != nil {
		return m.BurnBps
	}
	return 0
}

func (m *Params) GetCollectorModule() string {
	if m != nil {
		return m.CollectorModule
	}
	return ""
}

func (m *Params) GetTreasuryModule() string {
	if m != nil {
		return m.TreasuryModule
	}
	return ""
}

func (m *Params) GetProtectionModule() string {
	if m != nil {
		return m.ProtectionModule
	}
	return ""
}

func (m *Params) GetValidatorsModule() string {
	if m != nil {
		return m.ValidatorsModule
	}
	return ""
}

type FeeBalances struct {
	GasFees			github_com_cosmos_cosmos_sdk_types.Coins	`protobuf:"bytes,1,rep,name=gas_fees,json=gasFees,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"gas_fees"`
	ForwardingFees		github_com_cosmos_cosmos_sdk_types.Coins	`protobuf:"bytes,2,rep,name=forwarding_fees,json=forwardingFees,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"forwarding_fees"`
	ProtocolFees		github_com_cosmos_cosmos_sdk_types.Coins	`protobuf:"bytes,3,rep,name=protocol_fees,json=protocolFees,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"protocol_fees"`
	TotalCollected		github_com_cosmos_cosmos_sdk_types.Coins	`protobuf:"bytes,4,rep,name=total_collected,json=totalCollected,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"total_collected"`
	TotalDistributed	github_com_cosmos_cosmos_sdk_types.Coins	`protobuf:"bytes,5,rep,name=total_distributed,json=totalDistributed,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"total_distributed"`
	TotalBurned		github_com_cosmos_cosmos_sdk_types.Coins	`protobuf:"bytes,6,rep,name=total_burned,json=totalBurned,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"total_burned"`
}

func (m *FeeBalances) Reset()		{ *m = FeeBalances{} }
func (m *FeeBalances) String() string	{ return proto.CompactTextString(m) }
func (*FeeBalances) ProtoMessage()	{}
func (*FeeBalances) Descriptor() ([]byte, []int) {
	return fileDescriptor_027cc99771428dd9, []int{2}
}
func (m *FeeBalances) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *FeeBalances) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_FeeBalances.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *FeeBalances) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FeeBalances.Merge(m, src)
}
func (m *FeeBalances) XXX_Size() int {
	return m.Size()
}
func (m *FeeBalances) XXX_DiscardUnknown() {
	xxx_messageInfo_FeeBalances.DiscardUnknown(m)
}

var xxx_messageInfo_FeeBalances proto.InternalMessageInfo

func (m *FeeBalances) GetGasFees() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.GasFees
	}
	return nil
}

func (m *FeeBalances) GetForwardingFees() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.ForwardingFees
	}
	return nil
}

func (m *FeeBalances) GetProtocolFees() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.ProtocolFees
	}
	return nil
}

func (m *FeeBalances) GetTotalCollected() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.TotalCollected
	}
	return nil
}

func (m *FeeBalances) GetTotalDistributed() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.TotalDistributed
	}
	return nil
}

func (m *FeeBalances) GetTotalBurned() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.TotalBurned
	}
	return nil
}

type PendingDistribution struct {
	Treasury	github_com_cosmos_cosmos_sdk_types.Coins	`protobuf:"bytes,1,rep,name=treasury,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"treasury"`
	Protection	github_com_cosmos_cosmos_sdk_types.Coins	`protobuf:"bytes,2,rep,name=protection,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"protection"`
	Validators	github_com_cosmos_cosmos_sdk_types.Coins	`protobuf:"bytes,3,rep,name=validators,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"validators"`
	Burn		github_com_cosmos_cosmos_sdk_types.Coins	`protobuf:"bytes,4,rep,name=burn,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"burn"`
}

func (m *PendingDistribution) Reset()		{ *m = PendingDistribution{} }
func (m *PendingDistribution) String() string	{ return proto.CompactTextString(m) }
func (*PendingDistribution) ProtoMessage()	{}
func (*PendingDistribution) Descriptor() ([]byte, []int) {
	return fileDescriptor_027cc99771428dd9, []int{3}
}
func (m *PendingDistribution) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *PendingDistribution) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_PendingDistribution.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *PendingDistribution) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PendingDistribution.Merge(m, src)
}
func (m *PendingDistribution) XXX_Size() int {
	return m.Size()
}
func (m *PendingDistribution) XXX_DiscardUnknown() {
	xxx_messageInfo_PendingDistribution.DiscardUnknown(m)
}

var xxx_messageInfo_PendingDistribution proto.InternalMessageInfo

func (m *PendingDistribution) GetTreasury() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.Treasury
	}
	return nil
}

func (m *PendingDistribution) GetProtection() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.Protection
	}
	return nil
}

func (m *PendingDistribution) GetValidators() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.Validators
	}
	return nil
}

func (m *PendingDistribution) GetBurn() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.Burn
	}
	return nil
}

type FeeHistoryEntry struct {
	Epoch			uint64						`protobuf:"varint,1,opt,name=epoch,proto3" json:"epoch,omitempty"`
	Collected		github_com_cosmos_cosmos_sdk_types.Coins	`protobuf:"bytes,2,rep,name=collected,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"collected"`
	Treasury		github_com_cosmos_cosmos_sdk_types.Coins	`protobuf:"bytes,3,rep,name=treasury,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"treasury"`
	Protection		github_com_cosmos_cosmos_sdk_types.Coins	`protobuf:"bytes,4,rep,name=protection,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"protection"`
	Validators		github_com_cosmos_cosmos_sdk_types.Coins	`protobuf:"bytes,5,rep,name=validators,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"validators"`
	Burn			github_com_cosmos_cosmos_sdk_types.Coins	`protobuf:"bytes,6,rep,name=burn,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"burn"`
	RoundingRemainder	github_com_cosmos_cosmos_sdk_types.Coins	`protobuf:"bytes,7,rep,name=rounding_remainder,json=roundingRemainder,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"rounding_remainder"`
	DistributedAtHeight	int64						`protobuf:"varint,8,opt,name=distributed_at_height,json=distributedAtHeight,proto3" json:"distributed_at_height,omitempty"`
}

func (m *FeeHistoryEntry) Reset()		{ *m = FeeHistoryEntry{} }
func (m *FeeHistoryEntry) String() string	{ return proto.CompactTextString(m) }
func (*FeeHistoryEntry) ProtoMessage()		{}
func (*FeeHistoryEntry) Descriptor() ([]byte, []int) {
	return fileDescriptor_027cc99771428dd9, []int{4}
}
func (m *FeeHistoryEntry) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *FeeHistoryEntry) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_FeeHistoryEntry.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *FeeHistoryEntry) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FeeHistoryEntry.Merge(m, src)
}
func (m *FeeHistoryEntry) XXX_Size() int {
	return m.Size()
}
func (m *FeeHistoryEntry) XXX_DiscardUnknown() {
	xxx_messageInfo_FeeHistoryEntry.DiscardUnknown(m)
}

var xxx_messageInfo_FeeHistoryEntry proto.InternalMessageInfo

func (m *FeeHistoryEntry) GetEpoch() uint64 {
	if m != nil {
		return m.Epoch
	}
	return 0
}

func (m *FeeHistoryEntry) GetCollected() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.Collected
	}
	return nil
}

func (m *FeeHistoryEntry) GetTreasury() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.Treasury
	}
	return nil
}

func (m *FeeHistoryEntry) GetProtection() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.Protection
	}
	return nil
}

func (m *FeeHistoryEntry) GetValidators() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.Validators
	}
	return nil
}

func (m *FeeHistoryEntry) GetBurn() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.Burn
	}
	return nil
}

func (m *FeeHistoryEntry) GetRoundingRemainder() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.RoundingRemainder
	}
	return nil
}

func (m *FeeHistoryEntry) GetDistributedAtHeight() int64 {
	if m != nil {
		return m.DistributedAtHeight
	}
	return 0
}

func init() {
	proto.RegisterType((*GenesisState)(nil), "l1.feecollector.v1.GenesisState")
	proto.RegisterType((*Params)(nil), "l1.feecollector.v1.Params")
	proto.RegisterType((*FeeBalances)(nil), "l1.feecollector.v1.FeeBalances")
	proto.RegisterType((*PendingDistribution)(nil), "l1.feecollector.v1.PendingDistribution")
	proto.RegisterType((*FeeHistoryEntry)(nil), "l1.feecollector.v1.FeeHistoryEntry")
}

func init()	{ proto.RegisterFile("l1/feecollector/v1/genesis.proto", fileDescriptor_027cc99771428dd9) }

var fileDescriptor_027cc99771428dd9 = []byte{

	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xbc, 0x56, 0xcd, 0x6e, 0xdb, 0x46,
	0x10, 0xd6, 0xff, 0xcf, 0x48, 0xb1, 0xec, 0xb5, 0x0b, 0x30, 0x06, 0x2a, 0xbb, 0x2a, 0x8a, 0xb8,
	0x28, 0x4c, 0x86, 0xee, 0xa5, 0xd7, 0x28, 0x6e, 0x1a, 0x04, 0x28, 0x50, 0xa8, 0xb7, 0x5e, 0xd8,
	0x25, 0x39, 0xa2, 0x88, 0x50, 0x5c, 0x62, 0x77, 0xa5, 0x46, 0x7d, 0x8a, 0x3e, 0x47, 0xdf, 0x20,
	0x87, 0xde, 0x73, 0xcc, 0xb1, 0xa7, 0xfe, 0xd8, 0x40, 0x9f, 0xa3, 0xd8, 0x5d, 0x52, 0x64, 0x6b,
	0xe7, 0x26, 0xf9, 0x24, 0x71, 0xf6, 0xdb, 0xef, 0x9b, 0x9d, 0xf9, 0x86, 0x5c, 0x38, 0x4f, 0x5c,
	0x67, 0x8e, 0x18, 0xb0, 0x24, 0xc1, 0x40, 0x32, 0xee, 0xac, 0x5d, 0x27, 0xc2, 0x14, 0x45, 0x2c,
	0xec, 0x8c, 0x33, 0xc9, 0x08, 0x49, 0x5c, 0xbb, 0x8a, 0xb0, 0xd7, 0xee, 0xe9, 0x38, 0x60, 0x62,
	0xc9, 0x84, 0xe3, 0x53, 0x81, 0xce, 0xda, 0xf5, 0x51, 0x52, 0xd7, 0x09, 0x58, 0x9c, 0x9a, 0x3d,
	0xa7, 0x27, 0x11, 0x8b, 0x98, 0xfe, 0xeb, 0xa8, 0x7f, 0x26, 0x3a, 0x79, 0xdb, 0x80, 0xe1, 0x37,
	0x86, 0xfb, 0x7b, 0x49, 0x25, 0x92, 0xaf, 0xa0, 0x93, 0x51, 0x4e, 0x97, 0xc2, 0xaa, 0x9f, 0xd7,
	0x2f, 0x06, 0x57, 0xa7, 0xf6, 0x5d, 0x2d, 0xfb, 0x3b, 0x8d, 0x98, 0xb6, 0xde, 0xfd, 0x71, 0x56,
	0x9b, 0xe5, 0x78, 0xf2, 0x0c, 0x7a, 0x3e, 0x4d, 0x68, 0x1a, 0xa0, 0xb0, 0x1a, 0x7a, 0xef, 0xd9,
	0x7d, 0x7b, 0x5f, 0x20, 0x4e, 0x73, 0x58, 0x4e, 0xb0, 0xdd, 0x46, 0x7e, 0x84, 0x93, 0x0c, 0xd3,
	0x30, 0x4e, 0x23, 0x2f, 0x8c, 0x85, 0xe4, 0xb1, 0xbf, 0x92, 0x31, 0x4b, 0xad, 0xa6, 0xa6, 0x7b,
	0x72, 0x6f, 0x2a, 0x06, 0x7f, 0x5d, 0x81, 0xe7, 0xb4, 0xc7, 0xd9, 0xdd, 0x25, 0xf2, 0x0a, 0x06,
	0x73, 0x44, 0x6f, 0x11, 0x0b, 0xc9, 0xf8, 0xc6, 0x6a, 0x9d, 0x37, 0x2f, 0x06, 0x57, 0x9f, 0x7e,
	0x20, 0xcf, 0x97, 0x06, 0xf5, 0x75, 0x2a, 0xf9, 0x26, 0x27, 0x85, 0xf9, 0x36, 0x3c, 0xf9, 0xbb,
	0x01, 0x1d, 0x53, 0x09, 0xf2, 0x31, 0x80, 0xaa, 0xbb, 0x17, 0x62, 0xca, 0x96, 0xba, 0x72, 0xfd,
	0x59, 0x5f, 0x45, 0xae, 0x55, 0x80, 0x7c, 0x02, 0x43, 0xc9, 0x91, 0x8a, 0x15, 0xdf, 0x78, 0x7e,
	0x66, 0xca, 0xf3, 0x68, 0x36, 0x28, 0x62, 0xd3, 0x4c, 0x90, 0xcf, 0xe0, 0x40, 0x75, 0x04, 0x03,
	0x95, 0xa6, 0x06, 0x35, 0x35, 0xe8, 0x51, 0x19, 0xcd, 0x61, 0x6b, 0x9a, 0xc4, 0x21, 0x95, 0x8c,
	0x0b, 0x0d, 0x6b, 0x19, 0x58, 0x19, 0x55, 0xb0, 0xc7, 0xd0, 0xf3, 0x57, 0xdc, 0xf0, 0xb4, 0x35,
	0xa0, 0xab, 0x9e, 0xd5, 0xd2, 0xe7, 0x70, 0xb8, 0x3d, 0xa7, 0xb7, 0x64, 0xe1, 0x2a, 0x41, 0xab,
	0xa3, 0x13, 0x1e, 0x6d, 0xe3, 0xdf, 0xea, 0x30, 0x79, 0x02, 0xa3, 0x6d, 0xda, 0x39, 0xb2, 0xab,
	0x91, 0x07, 0x45, 0x38, 0x07, 0x7e, 0x01, 0x47, 0x95, 0xe4, 0x73, 0x68, 0x4f, 0x43, 0x0f, 0xcb,
	0x85, 0x12, 0x5c, 0x39, 0x42, 0x0e, 0xee, 0x1b, 0x70, 0xb9, 0x60, 0xc0, 0x93, 0xdf, 0xda, 0x30,
	0xa8, 0x38, 0x86, 0xcc, 0xa1, 0x17, 0x51, 0xe1, 0xcd, 0x11, 0x95, 0x41, 0x55, 0xf3, 0x1e, 0xdb,
	0xc6, 0xf8, 0xb6, 0x2a, 0xb7, 0x9d, 0x1b, 0xdf, 0x7e, 0xce, 0xe2, 0x74, 0xfa, 0x54, 0xb5, 0xec,
	0xd7, 0x3f, 0xcf, 0x2e, 0xa2, 0x58, 0x2e, 0x56, 0xbe, 0x1d, 0xb0, 0xa5, 0x93, 0x4f, 0x89, 0xf9,
	0xb9, 0x14, 0xe1, 0x6b, 0x47, 0x6e, 0x32, 0x14, 0x7a, 0x83, 0x98, 0x75, 0x23, 0x2a, 0x5e, 0x20,
	0x0a, 0x22, 0x61, 0x34, 0x67, 0xfc, 0x27, 0xca, 0xb5, 0x19, 0xb5, 0x5c, 0x63, 0xf7, 0x72, 0x07,
	0xa5, 0x86, 0x56, 0xcd, 0x40, 0xb7, 0x9b, 0x05, 0x2c, 0x31, 0x9a, 0xcd, 0xdd, 0x6b, 0x0e, 0x0b,
	0x85, 0xe2, 0x9c, 0x92, 0x49, 0x9a, 0x78, 0x79, 0xef, 0x31, 0xcc, 0x67, 0x62, 0xb7, 0xe7, 0xd4,
	0x1a, 0xcf, 0x0b, 0x09, 0xf2, 0x06, 0x8e, 0x8c, 0xea, 0x76, 0xca, 0x31, 0xb4, 0xda, 0xbb, 0xd7,
	0x3d, 0xd4, 0x2a, 0xd7, 0xa5, 0x08, 0x49, 0x61, 0x68, 0x94, 0xd5, 0x38, 0x60, 0x68, 0x75, 0x76,
	0x2f, 0x3a, 0xd0, 0x02, 0x53, 0xcd, 0x3f, 0x79, 0xdb, 0x84, 0xe3, 0x7b, 0x5e, 0x51, 0x24, 0x82,
	0x5e, 0x31, 0x43, 0xfb, 0xf0, 0xf1, 0x96, 0x9c, 0xbc, 0x06, 0x28, 0x27, 0x70, 0x1f, 0x1e, 0xae,
	0xd0, 0x2b, 0xb1, 0x72, 0x82, 0xf7, 0x61, 0xde, 0x0a, 0x3d, 0xf1, 0xa0, 0xa5, 0x9a, 0xb8, 0x0f,
	0xbf, 0x6a, 0xe2, 0xc9, 0x3f, 0x6d, 0x18, 0xfd, 0xef, 0x2b, 0x40, 0x4e, 0xa0, 0x8d, 0x19, 0x0b,
	0x16, 0xfa, 0x1d, 0xdf, 0x9a, 0x99, 0x07, 0x12, 0x43, 0xbf, 0x9c, 0x9f, 0x3d, 0xd4, 0xb8, 0x64,
	0xff, 0x8f, 0x71, 0x9a, 0x0f, 0x67, 0x9c, 0xd6, 0x43, 0x1a, 0xa7, 0xfd, 0x30, 0xc6, 0xe9, 0xec,
	0xc9, 0x38, 0xe4, 0x67, 0x20, 0x9c, 0xad, 0xcc, 0x3d, 0x86, 0xe3, 0x92, 0xc6, 0x69, 0x88, 0xdc,
	0xea, 0xee, 0x5e, 0xee, 0xa8, 0x90, 0x99, 0x15, 0x2a, 0xe4, 0x0a, 0x3e, 0xaa, 0xbc, 0x54, 0x3d,
	0x2a, 0xbd, 0x05, 0xc6, 0xd1, 0x42, 0xea, 0xcf, 0x71, 0x73, 0x76, 0x5c, 0x59, 0x7c, 0x26, 0x5f,
	0xea, 0xa5, 0xe9, 0xab, 0x77, 0x37, 0xe3, 0xfa, 0xfb, 0x9b, 0x71, 0xfd, 0xaf, 0x9b, 0x71, 0xfd,
	0x97, 0xdb, 0x71, 0xed, 0xfd, 0xed, 0xb8, 0xf6, 0xfb, 0xed, 0xb8, 0xf6, 0xc3, 0xd3, 0x4a, 0x2a,
	0x82, 0xad, 0x91, 0x63, 0x1c, 0xa5, 0x97, 0x89, 0xeb, 0x24, 0xae, 0xf3, 0x46, 0x5d, 0x52, 0x2f,
	0xcb, 0x5b, 0xaa, 0x4e, 0xcc, 0xef, 0xe8, 0xcf, 0xcb, 0x97, 0xff, 0x06, 0x00, 0x00, 0xff, 0xff,
	0x1e, 0xe0, 0x87, 0xd0, 0xc5, 0x0a, 0x00, 0x00,
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
	if len(m.FeeHistory) > 0 {
		for iNdEx := len(m.FeeHistory) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.FeeHistory[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x22
		}
	}
	{
		size, err := m.PendingDistribution.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x1a
	{
		size, err := m.Balances.MarshalToSizedBuffer(dAtA[:i])
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
	if len(m.ValidatorsModule) > 0 {
		i -= len(m.ValidatorsModule)
		copy(dAtA[i:], m.ValidatorsModule)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.ValidatorsModule)))
		i--
		dAtA[i] = 0x4a
	}
	if len(m.ProtectionModule) > 0 {
		i -= len(m.ProtectionModule)
		copy(dAtA[i:], m.ProtectionModule)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.ProtectionModule)))
		i--
		dAtA[i] = 0x42
	}
	if len(m.TreasuryModule) > 0 {
		i -= len(m.TreasuryModule)
		copy(dAtA[i:], m.TreasuryModule)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.TreasuryModule)))
		i--
		dAtA[i] = 0x3a
	}
	if len(m.CollectorModule) > 0 {
		i -= len(m.CollectorModule)
		copy(dAtA[i:], m.CollectorModule)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.CollectorModule)))
		i--
		dAtA[i] = 0x32
	}
	if m.BurnBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.BurnBps))
		i--
		dAtA[i] = 0x28
	}
	if m.ValidatorsBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.ValidatorsBps))
		i--
		dAtA[i] = 0x20
	}
	if m.ProtectionBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.ProtectionBps))
		i--
		dAtA[i] = 0x18
	}
	if m.TreasuryBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.TreasuryBps))
		i--
		dAtA[i] = 0x10
	}
	if len(m.BaseDenom) > 0 {
		i -= len(m.BaseDenom)
		copy(dAtA[i:], m.BaseDenom)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.BaseDenom)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *FeeBalances) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *FeeBalances) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *FeeBalances) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.TotalBurned) > 0 {
		for iNdEx := len(m.TotalBurned) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.TotalBurned[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x32
		}
	}
	if len(m.TotalDistributed) > 0 {
		for iNdEx := len(m.TotalDistributed) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.TotalDistributed[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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
			dAtA[i] = 0x22
		}
	}
	if len(m.ProtocolFees) > 0 {
		for iNdEx := len(m.ProtocolFees) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.ProtocolFees[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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
	if len(m.ForwardingFees) > 0 {
		for iNdEx := len(m.ForwardingFees) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.ForwardingFees[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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
	if len(m.GasFees) > 0 {
		for iNdEx := len(m.GasFees) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.GasFees[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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

func (m *PendingDistribution) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *PendingDistribution) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *PendingDistribution) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Burn) > 0 {
		for iNdEx := len(m.Burn) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Burn[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x22
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
			dAtA[i] = 0x1a
		}
	}
	if len(m.Protection) > 0 {
		for iNdEx := len(m.Protection) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Protection[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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
	if len(m.Treasury) > 0 {
		for iNdEx := len(m.Treasury) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Treasury[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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

func (m *FeeHistoryEntry) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *FeeHistoryEntry) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *FeeHistoryEntry) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.DistributedAtHeight != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.DistributedAtHeight))
		i--
		dAtA[i] = 0x40
	}
	if len(m.RoundingRemainder) > 0 {
		for iNdEx := len(m.RoundingRemainder) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.RoundingRemainder[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x3a
		}
	}
	if len(m.Burn) > 0 {
		for iNdEx := len(m.Burn) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Burn[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
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
	if len(m.Protection) > 0 {
		for iNdEx := len(m.Protection) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Protection[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x22
		}
	}
	if len(m.Treasury) > 0 {
		for iNdEx := len(m.Treasury) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Treasury[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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
	if len(m.Collected) > 0 {
		for iNdEx := len(m.Collected) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Collected[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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
	l = m.Balances.Size()
	n += 1 + l + sovGenesis(uint64(l))
	l = m.PendingDistribution.Size()
	n += 1 + l + sovGenesis(uint64(l))
	if len(m.FeeHistory) > 0 {
		for _, e := range m.FeeHistory {
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
	l = len(m.BaseDenom)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	if m.TreasuryBps != 0 {
		n += 1 + sovGenesis(uint64(m.TreasuryBps))
	}
	if m.ProtectionBps != 0 {
		n += 1 + sovGenesis(uint64(m.ProtectionBps))
	}
	if m.ValidatorsBps != 0 {
		n += 1 + sovGenesis(uint64(m.ValidatorsBps))
	}
	if m.BurnBps != 0 {
		n += 1 + sovGenesis(uint64(m.BurnBps))
	}
	l = len(m.CollectorModule)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	l = len(m.TreasuryModule)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	l = len(m.ProtectionModule)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	l = len(m.ValidatorsModule)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	return n
}

func (m *FeeBalances) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.GasFees) > 0 {
		for _, e := range m.GasFees {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.ForwardingFees) > 0 {
		for _, e := range m.ForwardingFees {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.ProtocolFees) > 0 {
		for _, e := range m.ProtocolFees {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.TotalCollected) > 0 {
		for _, e := range m.TotalCollected {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.TotalDistributed) > 0 {
		for _, e := range m.TotalDistributed {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.TotalBurned) > 0 {
		for _, e := range m.TotalBurned {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	return n
}

func (m *PendingDistribution) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Treasury) > 0 {
		for _, e := range m.Treasury {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.Protection) > 0 {
		for _, e := range m.Protection {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.Validators) > 0 {
		for _, e := range m.Validators {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.Burn) > 0 {
		for _, e := range m.Burn {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	return n
}

func (m *FeeHistoryEntry) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Epoch != 0 {
		n += 1 + sovGenesis(uint64(m.Epoch))
	}
	if len(m.Collected) > 0 {
		for _, e := range m.Collected {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.Treasury) > 0 {
		for _, e := range m.Treasury {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.Protection) > 0 {
		for _, e := range m.Protection {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.Validators) > 0 {
		for _, e := range m.Validators {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.Burn) > 0 {
		for _, e := range m.Burn {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.RoundingRemainder) > 0 {
		for _, e := range m.RoundingRemainder {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if m.DistributedAtHeight != 0 {
		n += 1 + sovGenesis(uint64(m.DistributedAtHeight))
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
				return fmt.Errorf("proto: wrong wireType = %d for field Balances", wireType)
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
			if err := m.Balances.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PendingDistribution", wireType)
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
			if err := m.PendingDistribution.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field FeeHistory", wireType)
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
			m.FeeHistory = append(m.FeeHistory, FeeHistoryEntry{})
			if err := m.FeeHistory[len(m.FeeHistory)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BaseDenom", wireType)
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
			m.BaseDenom = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field TreasuryBps", wireType)
			}
			m.TreasuryBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.TreasuryBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ProtectionBps", wireType)
			}
			m.ProtectionBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ProtectionBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ValidatorsBps", wireType)
			}
			m.ValidatorsBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ValidatorsBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field BurnBps", wireType)
			}
			m.BurnBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.BurnBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field CollectorModule", wireType)
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
			m.CollectorModule = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TreasuryModule", wireType)
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
			m.TreasuryModule = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 8:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ProtectionModule", wireType)
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
			m.ProtectionModule = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 9:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ValidatorsModule", wireType)
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
			m.ValidatorsModule = string(dAtA[iNdEx:postIndex])
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
func (m *FeeBalances) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: FeeBalances: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: FeeBalances: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field GasFees", wireType)
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
			m.GasFees = append(m.GasFees, types.Coin{})
			if err := m.GasFees[len(m.GasFees)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ForwardingFees", wireType)
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
			m.ForwardingFees = append(m.ForwardingFees, types.Coin{})
			if err := m.ForwardingFees[len(m.ForwardingFees)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ProtocolFees", wireType)
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
			m.ProtocolFees = append(m.ProtocolFees, types.Coin{})
			if err := m.ProtocolFees[len(m.ProtocolFees)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
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
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TotalDistributed", wireType)
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
			m.TotalDistributed = append(m.TotalDistributed, types.Coin{})
			if err := m.TotalDistributed[len(m.TotalDistributed)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TotalBurned", wireType)
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
			m.TotalBurned = append(m.TotalBurned, types.Coin{})
			if err := m.TotalBurned[len(m.TotalBurned)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
func (m *PendingDistribution) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: PendingDistribution: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: PendingDistribution: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Treasury", wireType)
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
			m.Treasury = append(m.Treasury, types.Coin{})
			if err := m.Treasury[len(m.Treasury)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Protection", wireType)
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
			m.Protection = append(m.Protection, types.Coin{})
			if err := m.Protection[len(m.Protection)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
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
			m.Validators = append(m.Validators, types.Coin{})
			if err := m.Validators[len(m.Validators)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Burn", wireType)
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
			m.Burn = append(m.Burn, types.Coin{})
			if err := m.Burn[len(m.Burn)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
func (m *FeeHistoryEntry) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: FeeHistoryEntry: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: FeeHistoryEntry: illegal tag %d (wire type %d)", fieldNum, wire)
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
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Collected", wireType)
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
			m.Collected = append(m.Collected, types.Coin{})
			if err := m.Collected[len(m.Collected)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Treasury", wireType)
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
			m.Treasury = append(m.Treasury, types.Coin{})
			if err := m.Treasury[len(m.Treasury)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Protection", wireType)
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
			m.Protection = append(m.Protection, types.Coin{})
			if err := m.Protection[len(m.Protection)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
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
			m.Validators = append(m.Validators, types.Coin{})
			if err := m.Validators[len(m.Validators)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Burn", wireType)
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
			m.Burn = append(m.Burn, types.Coin{})
			if err := m.Burn[len(m.Burn)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RoundingRemainder", wireType)
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
			m.RoundingRemainder = append(m.RoundingRemainder, types.Coin{})
			if err := m.RoundingRemainder[len(m.RoundingRemainder)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 8:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field DistributedAtHeight", wireType)
			}
			m.DistributedAtHeight = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.DistributedAtHeight |= int64(b&0x7F) << shift
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
