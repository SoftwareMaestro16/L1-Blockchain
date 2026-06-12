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
	Params		Params			`protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
	Allocations	TreasuryAllocations	`protobuf:"bytes,2,opt,name=allocations,proto3" json:"allocations"`
	Spends		[]TreasurySpend		`protobuf:"bytes,3,rep,name=spends,proto3" json:"spends"`
	EpochSpends	[]EpochSpend		`protobuf:"bytes,4,rep,name=epoch_spends,json=epochSpends,proto3" json:"epoch_spends"`
	NextSpendId	uint64			`protobuf:"varint,5,opt,name=next_spend_id,json=nextSpendId,proto3" json:"next_spend_id,omitempty"`
}

func (m *GenesisState) Reset()		{ *m = GenesisState{} }
func (m *GenesisState) String() string	{ return proto.CompactTextString(m) }
func (*GenesisState) ProtoMessage()	{}
func (*GenesisState) Descriptor() ([]byte, []int) {
	return fileDescriptor_56b9e07eec10d57c, []int{0}
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

func (m *GenesisState) GetAllocations() TreasuryAllocations {
	if m != nil {
		return m.Allocations
	}
	return TreasuryAllocations{}
}

func (m *GenesisState) GetSpends() []TreasurySpend {
	if m != nil {
		return m.Spends
	}
	return nil
}

func (m *GenesisState) GetEpochSpends() []EpochSpend {
	if m != nil {
		return m.EpochSpends
	}
	return nil
}

func (m *GenesisState) GetNextSpendId() uint64 {
	if m != nil {
		return m.NextSpendId
	}
	return 0
}

type Params struct {
	BaseDenom			string		`protobuf:"bytes,1,opt,name=base_denom,json=baseDenom,proto3" json:"base_denom,omitempty"`
	TreasuryModule			string		`protobuf:"bytes,2,opt,name=treasury_module,json=treasuryModule,proto3" json:"treasury_module,omitempty"`
	ReserveBps			uint32		`protobuf:"varint,3,opt,name=reserve_bps,json=reserveBps,proto3" json:"reserve_bps,omitempty"`
	EcosystemBps			uint32		`protobuf:"varint,4,opt,name=ecosystem_bps,json=ecosystemBps,proto3" json:"ecosystem_bps,omitempty"`
	ValidatorIncentivesBps		uint32		`protobuf:"varint,5,opt,name=validator_incentives_bps,json=validatorIncentivesBps,proto3" json:"validator_incentives_bps,omitempty"`
	BurnBps				uint32		`protobuf:"varint,6,opt,name=burn_bps,json=burnBps,proto3" json:"burn_bps,omitempty"`
	PerEpochSpendCap		types.Coin	`protobuf:"bytes,7,opt,name=per_epoch_spend_cap,json=perEpochSpendCap,proto3" json:"per_epoch_spend_cap"`
	RecipientAllowlistEnabled	bool		`protobuf:"varint,8,opt,name=recipient_allowlist_enabled,json=recipientAllowlistEnabled,proto3" json:"recipient_allowlist_enabled,omitempty"`
	RecipientAllowlist		[]string	`protobuf:"bytes,9,rep,name=recipient_allowlist,json=recipientAllowlist,proto3" json:"recipient_allowlist,omitempty"`
	MaxMetadataBytes		uint32		`protobuf:"varint,10,opt,name=max_metadata_bytes,json=maxMetadataBytes,proto3" json:"max_metadata_bytes,omitempty"`
}

func (m *Params) Reset()		{ *m = Params{} }
func (m *Params) String() string	{ return proto.CompactTextString(m) }
func (*Params) ProtoMessage()		{}
func (*Params) Descriptor() ([]byte, []int) {
	return fileDescriptor_56b9e07eec10d57c, []int{1}
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

func (m *Params) GetTreasuryModule() string {
	if m != nil {
		return m.TreasuryModule
	}
	return ""
}

func (m *Params) GetReserveBps() uint32 {
	if m != nil {
		return m.ReserveBps
	}
	return 0
}

func (m *Params) GetEcosystemBps() uint32 {
	if m != nil {
		return m.EcosystemBps
	}
	return 0
}

func (m *Params) GetValidatorIncentivesBps() uint32 {
	if m != nil {
		return m.ValidatorIncentivesBps
	}
	return 0
}

func (m *Params) GetBurnBps() uint32 {
	if m != nil {
		return m.BurnBps
	}
	return 0
}

func (m *Params) GetPerEpochSpendCap() types.Coin {
	if m != nil {
		return m.PerEpochSpendCap
	}
	return types.Coin{}
}

func (m *Params) GetRecipientAllowlistEnabled() bool {
	if m != nil {
		return m.RecipientAllowlistEnabled
	}
	return false
}

func (m *Params) GetRecipientAllowlist() []string {
	if m != nil {
		return m.RecipientAllowlist
	}
	return nil
}

func (m *Params) GetMaxMetadataBytes() uint32 {
	if m != nil {
		return m.MaxMetadataBytes
	}
	return 0
}

type TreasuryAllocations struct {
	ReserveBalance			github_com_cosmos_cosmos_sdk_types.Coins	`protobuf:"bytes,1,rep,name=reserve_balance,json=reserveBalance,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"reserve_balance"`
	EcosystemBalance		github_com_cosmos_cosmos_sdk_types.Coins	`protobuf:"bytes,2,rep,name=ecosystem_balance,json=ecosystemBalance,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"ecosystem_balance"`
	ValidatorIncentiveBalance	github_com_cosmos_cosmos_sdk_types.Coins	`protobuf:"bytes,3,rep,name=validator_incentive_balance,json=validatorIncentiveBalance,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"validator_incentive_balance"`
	BurnBalance			github_com_cosmos_cosmos_sdk_types.Coins	`protobuf:"bytes,4,rep,name=burn_balance,json=burnBalance,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"burn_balance"`
	TotalReceived			github_com_cosmos_cosmos_sdk_types.Coins	`protobuf:"bytes,5,rep,name=total_received,json=totalReceived,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"total_received"`
	TotalSpent			github_com_cosmos_cosmos_sdk_types.Coins	`protobuf:"bytes,6,rep,name=total_spent,json=totalSpent,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"total_spent"`
}

func (m *TreasuryAllocations) Reset()		{ *m = TreasuryAllocations{} }
func (m *TreasuryAllocations) String() string	{ return proto.CompactTextString(m) }
func (*TreasuryAllocations) ProtoMessage()	{}
func (*TreasuryAllocations) Descriptor() ([]byte, []int) {
	return fileDescriptor_56b9e07eec10d57c, []int{2}
}
func (m *TreasuryAllocations) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *TreasuryAllocations) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_TreasuryAllocations.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *TreasuryAllocations) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TreasuryAllocations.Merge(m, src)
}
func (m *TreasuryAllocations) XXX_Size() int {
	return m.Size()
}
func (m *TreasuryAllocations) XXX_DiscardUnknown() {
	xxx_messageInfo_TreasuryAllocations.DiscardUnknown(m)
}

var xxx_messageInfo_TreasuryAllocations proto.InternalMessageInfo

func (m *TreasuryAllocations) GetReserveBalance() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.ReserveBalance
	}
	return nil
}

func (m *TreasuryAllocations) GetEcosystemBalance() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.EcosystemBalance
	}
	return nil
}

func (m *TreasuryAllocations) GetValidatorIncentiveBalance() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.ValidatorIncentiveBalance
	}
	return nil
}

func (m *TreasuryAllocations) GetBurnBalance() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.BurnBalance
	}
	return nil
}

func (m *TreasuryAllocations) GetTotalReceived() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.TotalReceived
	}
	return nil
}

func (m *TreasuryAllocations) GetTotalSpent() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.TotalSpent
	}
	return nil
}

type TreasurySpend struct {
	Id			uint64						`protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Proposer		string						`protobuf:"bytes,2,opt,name=proposer,proto3" json:"proposer,omitempty"`
	Recipient		string						`protobuf:"bytes,3,opt,name=recipient,proto3" json:"recipient,omitempty"`
	Amount			github_com_cosmos_cosmos_sdk_types.Coins	`protobuf:"bytes,4,rep,name=amount,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"amount"`
	Bucket			string						`protobuf:"bytes,5,opt,name=bucket,proto3" json:"bucket,omitempty"`
	Status			string						`protobuf:"bytes,6,opt,name=status,proto3" json:"status,omitempty"`
	Epoch			uint64						`protobuf:"varint,7,opt,name=epoch,proto3" json:"epoch,omitempty"`
	VestingStartEpoch	uint64						`protobuf:"varint,8,opt,name=vesting_start_epoch,json=vestingStartEpoch,proto3" json:"vesting_start_epoch,omitempty"`
	VestingEndEpoch		uint64						`protobuf:"varint,9,opt,name=vesting_end_epoch,json=vestingEndEpoch,proto3" json:"vesting_end_epoch,omitempty"`
	Metadata		string						`protobuf:"bytes,10,opt,name=metadata,proto3" json:"metadata,omitempty"`
	CreatedHeight		int64						`protobuf:"varint,11,opt,name=created_height,json=createdHeight,proto3" json:"created_height,omitempty"`
	UpdatedHeight		int64						`protobuf:"varint,12,opt,name=updated_height,json=updatedHeight,proto3" json:"updated_height,omitempty"`
	ExecutedHeight		int64						`protobuf:"varint,13,opt,name=executed_height,json=executedHeight,proto3" json:"executed_height,omitempty"`
}

func (m *TreasurySpend) Reset()		{ *m = TreasurySpend{} }
func (m *TreasurySpend) String() string	{ return proto.CompactTextString(m) }
func (*TreasurySpend) ProtoMessage()	{}
func (*TreasurySpend) Descriptor() ([]byte, []int) {
	return fileDescriptor_56b9e07eec10d57c, []int{3}
}
func (m *TreasurySpend) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *TreasurySpend) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_TreasurySpend.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *TreasurySpend) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TreasurySpend.Merge(m, src)
}
func (m *TreasurySpend) XXX_Size() int {
	return m.Size()
}
func (m *TreasurySpend) XXX_DiscardUnknown() {
	xxx_messageInfo_TreasurySpend.DiscardUnknown(m)
}

var xxx_messageInfo_TreasurySpend proto.InternalMessageInfo

func (m *TreasurySpend) GetId() uint64 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *TreasurySpend) GetProposer() string {
	if m != nil {
		return m.Proposer
	}
	return ""
}

func (m *TreasurySpend) GetRecipient() string {
	if m != nil {
		return m.Recipient
	}
	return ""
}

func (m *TreasurySpend) GetAmount() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.Amount
	}
	return nil
}

func (m *TreasurySpend) GetBucket() string {
	if m != nil {
		return m.Bucket
	}
	return ""
}

func (m *TreasurySpend) GetStatus() string {
	if m != nil {
		return m.Status
	}
	return ""
}

func (m *TreasurySpend) GetEpoch() uint64 {
	if m != nil {
		return m.Epoch
	}
	return 0
}

func (m *TreasurySpend) GetVestingStartEpoch() uint64 {
	if m != nil {
		return m.VestingStartEpoch
	}
	return 0
}

func (m *TreasurySpend) GetVestingEndEpoch() uint64 {
	if m != nil {
		return m.VestingEndEpoch
	}
	return 0
}

func (m *TreasurySpend) GetMetadata() string {
	if m != nil {
		return m.Metadata
	}
	return ""
}

func (m *TreasurySpend) GetCreatedHeight() int64 {
	if m != nil {
		return m.CreatedHeight
	}
	return 0
}

func (m *TreasurySpend) GetUpdatedHeight() int64 {
	if m != nil {
		return m.UpdatedHeight
	}
	return 0
}

func (m *TreasurySpend) GetExecutedHeight() int64 {
	if m != nil {
		return m.ExecutedHeight
	}
	return 0
}

type EpochSpend struct {
	Epoch	uint64						`protobuf:"varint,1,opt,name=epoch,proto3" json:"epoch,omitempty"`
	Spent	github_com_cosmos_cosmos_sdk_types.Coins	`protobuf:"bytes,2,rep,name=spent,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"spent"`
}

func (m *EpochSpend) Reset()		{ *m = EpochSpend{} }
func (m *EpochSpend) String() string	{ return proto.CompactTextString(m) }
func (*EpochSpend) ProtoMessage()	{}
func (*EpochSpend) Descriptor() ([]byte, []int) {
	return fileDescriptor_56b9e07eec10d57c, []int{4}
}
func (m *EpochSpend) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *EpochSpend) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_EpochSpend.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *EpochSpend) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EpochSpend.Merge(m, src)
}
func (m *EpochSpend) XXX_Size() int {
	return m.Size()
}
func (m *EpochSpend) XXX_DiscardUnknown() {
	xxx_messageInfo_EpochSpend.DiscardUnknown(m)
}

var xxx_messageInfo_EpochSpend proto.InternalMessageInfo

func (m *EpochSpend) GetEpoch() uint64 {
	if m != nil {
		return m.Epoch
	}
	return 0
}

func (m *EpochSpend) GetSpent() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.Spent
	}
	return nil
}

func init() {
	proto.RegisterType((*GenesisState)(nil), "l1.treasury.v1.GenesisState")
	proto.RegisterType((*Params)(nil), "l1.treasury.v1.Params")
	proto.RegisterType((*TreasuryAllocations)(nil), "l1.treasury.v1.TreasuryAllocations")
	proto.RegisterType((*TreasurySpend)(nil), "l1.treasury.v1.TreasurySpend")
	proto.RegisterType((*EpochSpend)(nil), "l1.treasury.v1.EpochSpend")
}

func init()	{ proto.RegisterFile("l1/treasury/v1/genesis.proto", fileDescriptor_56b9e07eec10d57c) }

var fileDescriptor_56b9e07eec10d57c = []byte{

	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xb4, 0x56, 0xdd, 0x8e, 0xdb, 0x44,
	0x14, 0x5e, 0xe7, 0xaf, 0xf1, 0xc9, 0xcf, 0xb6, 0xb3, 0xd5, 0xca, 0x9b, 0xb6, 0xd9, 0x28, 0x15,
	0x6a, 0x04, 0xd4, 0x26, 0x85, 0x0b, 0x24, 0x24, 0xa4, 0xee, 0xb2, 0x82, 0x0a, 0x15, 0x21, 0x2f,
	0x57, 0xdc, 0x58, 0x13, 0xfb, 0x28, 0x6b, 0xd5, 0xf6, 0x58, 0x9e, 0x89, 0xc9, 0x3e, 0x00, 0x57,
	0xdc, 0xf0, 0x1c, 0xbc, 0x00, 0xaf, 0xd0, 0xcb, 0x5e, 0x22, 0x2e, 0x00, 0xed, 0x8a, 0x27, 0xe0,
	0x05, 0xd0, 0xfc, 0xd8, 0x49, 0xdb, 0x85, 0xab, 0xf4, 0x2a, 0x9e, 0xef, 0x7c, 0xe7, 0x7c, 0x33,
	0x67, 0xce, 0x7c, 0x0a, 0xdc, 0x4f, 0xe6, 0x9e, 0x28, 0x90, 0xf2, 0x55, 0x71, 0xe9, 0x95, 0x73,
	0x6f, 0x89, 0x19, 0xf2, 0x98, 0xbb, 0x79, 0xc1, 0x04, 0x23, 0xc3, 0x64, 0xee, 0x56, 0x51, 0xb7,
	0x9c, 0x8f, 0xc6, 0x21, 0xe3, 0x29, 0xe3, 0xde, 0x82, 0x72, 0xf4, 0xca, 0xf9, 0x02, 0x05, 0x9d,
	0x7b, 0x21, 0x8b, 0x33, 0xcd, 0x1f, 0xdd, 0x5d, 0xb2, 0x25, 0x53, 0x9f, 0x9e, 0xfc, 0xd2, 0xe8,
	0xf4, 0xd7, 0x06, 0xf4, 0xbf, 0xd4, 0x75, 0xcf, 0x05, 0x15, 0x48, 0x3e, 0x81, 0x4e, 0x4e, 0x0b,
	0x9a, 0x72, 0xc7, 0x9a, 0x58, 0xb3, 0xde, 0x93, 0x43, 0xf7, 0x75, 0x1d, 0xf7, 0x5b, 0x15, 0x3d,
	0x69, 0xbd, 0xfc, 0xe3, 0x78, 0xcf, 0x37, 0x5c, 0xf2, 0x35, 0xf4, 0x68, 0x92, 0xb0, 0x90, 0x8a,
	0x98, 0x65, 0xdc, 0x69, 0xa8, 0xd4, 0x87, 0x6f, 0xa6, 0x7e, 0x67, 0xbe, 0x9f, 0x6e, 0xa8, 0xa6,
	0xce, 0x76, 0x36, 0xf9, 0x0c, 0x3a, 0x3c, 0xc7, 0x2c, 0xe2, 0x4e, 0x73, 0xd2, 0x9c, 0xf5, 0x9e,
	0x3c, 0xf8, 0xaf, 0x3a, 0xe7, 0x92, 0x55, 0xed, 0x44, 0xa7, 0x90, 0x53, 0xe8, 0x63, 0xce, 0xc2,
	0x8b, 0xc0, 0x94, 0x68, 0xa9, 0x12, 0xa3, 0x37, 0x4b, 0x9c, 0x49, 0xce, 0x76, 0x7e, 0x0f, 0x6b,
	0x84, 0x93, 0x29, 0x0c, 0x32, 0x5c, 0x0b, 0x5d, 0x23, 0x88, 0x23, 0xa7, 0x3d, 0xb1, 0x66, 0x2d,
	0xbf, 0x27, 0x41, 0x45, 0x79, 0x16, 0x4d, 0x7f, 0x6f, 0x42, 0x47, 0xf7, 0x82, 0x3c, 0x00, 0x90,
	0x5d, 0x0f, 0x22, 0xcc, 0x58, 0xaa, 0xfa, 0x66, 0xfb, 0xb6, 0x44, 0xbe, 0x90, 0x00, 0x79, 0x04,
	0xfb, 0x95, 0x74, 0x90, 0xb2, 0x68, 0x95, 0xa0, 0x6a, 0x90, 0xed, 0x0f, 0x2b, 0xf8, 0xb9, 0x42,
	0xc9, 0x31, 0xf4, 0x0a, 0xe4, 0x58, 0x94, 0x18, 0x2c, 0x72, 0x79, 0x7a, 0x6b, 0x36, 0xf0, 0xc1,
	0x40, 0x27, 0x39, 0x27, 0x0f, 0x61, 0x80, 0x21, 0xe3, 0x97, 0x5c, 0x60, 0xaa, 0x28, 0x2d, 0x45,
	0xe9, 0xd7, 0xa0, 0x24, 0x7d, 0x0a, 0x4e, 0x49, 0x93, 0x38, 0xa2, 0x82, 0x15, 0x41, 0x9c, 0x85,
	0x98, 0x89, 0xb8, 0x44, 0xae, 0xf8, 0x6d, 0xc5, 0x3f, 0xac, 0xe3, 0xcf, 0xea, 0xb0, 0xcc, 0x3c,
	0x82, 0xee, 0x62, 0x55, 0x64, 0x8a, 0xd9, 0x51, 0xcc, 0x5b, 0x72, 0x2d, 0x43, 0xdf, 0xc0, 0x41,
	0x8e, 0x45, 0xb0, 0xd5, 0xda, 0x20, 0xa4, 0xb9, 0x73, 0x4b, 0x5d, 0xf4, 0x91, 0xab, 0x67, 0xcf,
	0x95, 0x67, 0x76, 0xcd, 0xec, 0xb9, 0xa7, 0x2c, 0xce, 0x4c, 0x73, 0x6f, 0xe7, 0x58, 0x6c, 0x3a,
	0x7e, 0x4a, 0x73, 0xf2, 0x39, 0xdc, 0x2b, 0x30, 0x8c, 0xf3, 0x18, 0x33, 0x11, 0xc8, 0xcb, 0xff,
	0x21, 0x89, 0xb9, 0x08, 0x30, 0xa3, 0x8b, 0x04, 0x23, 0xa7, 0x3b, 0xb1, 0x66, 0x5d, 0xff, 0xa8,
	0xa6, 0x3c, 0xad, 0x18, 0x67, 0x9a, 0x40, 0x3c, 0x38, 0xb8, 0x21, 0xdf, 0xb1, 0x27, 0xcd, 0x99,
	0xed, 0x93, 0xb7, 0xf3, 0xc8, 0x87, 0x40, 0x52, 0xba, 0x0e, 0x52, 0x14, 0x34, 0xa2, 0x82, 0x06,
	0x8b, 0x4b, 0x81, 0xdc, 0x01, 0x75, 0xca, 0xdb, 0x29, 0x5d, 0x3f, 0x37, 0x81, 0x13, 0x89, 0x4f,
	0xff, 0x6e, 0xc3, 0xc1, 0x0d, 0xd3, 0x4a, 0x04, 0xec, 0xd7, 0x37, 0x44, 0x13, 0x9a, 0x85, 0xe8,
	0x58, 0x6a, 0xc0, 0xfe, 0xa7, 0x05, 0x1f, 0xc9, 0x16, 0xfc, 0xf2, 0xe7, 0xf1, 0x6c, 0x19, 0x8b,
	0x8b, 0xd5, 0xc2, 0x0d, 0x59, 0xea, 0x99, 0xb7, 0xaa, 0x7f, 0x1e, 0xf3, 0xe8, 0x85, 0x27, 0x2e,
	0x73, 0xe4, 0x2a, 0x81, 0xfb, 0xc3, 0xea, 0xca, 0xb5, 0x04, 0x59, 0xc3, 0x9d, 0xad, 0x6b, 0x37,
	0xba, 0x8d, 0xdd, 0xeb, 0xde, 0xde, 0xcc, 0x91, 0x51, 0xfe, 0xc9, 0x82, 0x7b, 0x37, 0x0c, 0x53,
	0xbd, 0x89, 0xe6, 0xee, 0x37, 0x71, 0xf4, 0xf6, 0x70, 0x56, 0xbb, 0xc9, 0xa0, 0xaf, 0xe7, 0xd3,
	0xa8, 0xb7, 0x76, 0xaf, 0xde, 0x53, 0x03, 0x6f, 0xf4, 0x0a, 0x18, 0x0a, 0x26, 0x68, 0x12, 0x14,
	0x18, 0x62, 0x5c, 0xa2, 0xf4, 0x81, 0x9d, 0x2b, 0x0e, 0x94, 0x84, 0x6f, 0x14, 0x48, 0x02, 0x3d,
	0xad, 0x29, 0x1f, 0x99, 0x70, 0x3a, 0xbb, 0x17, 0x04, 0x55, 0x5f, 0xbe, 0x44, 0x31, 0xfd, 0xa7,
	0x09, 0x83, 0xd7, 0xdc, 0x94, 0x0c, 0xa1, 0x11, 0x47, 0xca, 0xc3, 0x5a, 0x7e, 0x23, 0x8e, 0xc8,
	0x08, 0xba, 0x79, 0xc1, 0x72, 0xc6, 0xb1, 0x30, 0xae, 0x55, 0xaf, 0xc9, 0x7d, 0xb0, 0xeb, 0x97,
	0xa6, 0xdc, 0xca, 0xf6, 0x37, 0x00, 0x09, 0xa1, 0x43, 0x53, 0xb6, 0xca, 0xc4, 0xbb, 0xb8, 0x27,
	0x53, 0x9a, 0x1c, 0x42, 0x67, 0xb1, 0x0a, 0x5f, 0xa0, 0x50, 0xd6, 0x66, 0xfb, 0x66, 0x25, 0x71,
	0x2e, 0xa8, 0x58, 0x69, 0x23, 0xb3, 0x7d, 0xb3, 0x22, 0x77, 0xa1, 0xad, 0x3c, 0x4c, 0x39, 0x57,
	0xcb, 0xd7, 0x0b, 0xe2, 0xc2, 0x41, 0x89, 0x5c, 0xc4, 0xd9, 0x32, 0xe0, 0x82, 0x16, 0x42, 0xfb,
	0x9c, 0x72, 0xa1, 0x96, 0x7f, 0xc7, 0x84, 0xce, 0x65, 0x44, 0xb9, 0x18, 0x79, 0x1f, 0x2a, 0x30,
	0x90, 0x4e, 0xa8, 0xd9, 0xb6, 0x62, 0xef, 0x9b, 0xc0, 0x59, 0x16, 0x69, 0xee, 0x08, 0xba, 0x95,
	0xe9, 0x28, 0xbb, 0xb1, 0xfd, 0x7a, 0x4d, 0xde, 0x83, 0x61, 0x58, 0x20, 0x15, 0x18, 0x05, 0x17,
	0x18, 0x2f, 0x2f, 0x84, 0xd3, 0x9b, 0x58, 0xb3, 0xa6, 0x3f, 0x30, 0xe8, 0x57, 0x0a, 0x94, 0xb4,
	0x55, 0x1e, 0x6d, 0xd3, 0xfa, 0x9a, 0x66, 0x50, 0x43, 0x7b, 0x04, 0xfb, 0xb8, 0xc6, 0x70, 0xb5,
	0xc5, 0x1b, 0x28, 0xde, 0xb0, 0x82, 0x35, 0x71, 0xfa, 0xa3, 0x05, 0xb0, 0xb1, 0xe3, 0x4d, 0x4f,
	0xac, 0xed, 0x9e, 0x50, 0x68, 0xeb, 0x11, 0x7c, 0x07, 0x46, 0xa3, 0x2b, 0x9f, 0x9c, 0xbd, 0xbc,
	0x1a, 0x5b, 0xaf, 0xae, 0xc6, 0xd6, 0x5f, 0x57, 0x63, 0xeb, 0xe7, 0xeb, 0xf1, 0xde, 0xab, 0xeb,
	0xf1, 0xde, 0x6f, 0xd7, 0xe3, 0xbd, 0xef, 0x3f, 0xd8, 0x2a, 0xc5, 0x59, 0x89, 0x05, 0xc6, 0xcb,
	0xec, 0x71, 0x32, 0xf7, 0x92, 0xb9, 0xb7, 0xde, 0xfc, 0x29, 0x52, 0x35, 0x17, 0x1d, 0xf5, 0x57,
	0xe6, 0xe3, 0x7f, 0x03, 0x00, 0x00, 0xff, 0xff, 0x35, 0xe9, 0x6c, 0x90, 0x30, 0x09, 0x00, 0x00,
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
	if m.NextSpendId != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.NextSpendId))
		i--
		dAtA[i] = 0x28
	}
	if len(m.EpochSpends) > 0 {
		for iNdEx := len(m.EpochSpends) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.EpochSpends[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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
	if len(m.Spends) > 0 {
		for iNdEx := len(m.Spends) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Spends[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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
	{
		size, err := m.Allocations.MarshalToSizedBuffer(dAtA[:i])
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
	if m.MaxMetadataBytes != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.MaxMetadataBytes))
		i--
		dAtA[i] = 0x50
	}
	if len(m.RecipientAllowlist) > 0 {
		for iNdEx := len(m.RecipientAllowlist) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.RecipientAllowlist[iNdEx])
			copy(dAtA[i:], m.RecipientAllowlist[iNdEx])
			i = encodeVarintGenesis(dAtA, i, uint64(len(m.RecipientAllowlist[iNdEx])))
			i--
			dAtA[i] = 0x4a
		}
	}
	if m.RecipientAllowlistEnabled {
		i--
		if m.RecipientAllowlistEnabled {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x40
	}
	{
		size, err := m.PerEpochSpendCap.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x3a
	if m.BurnBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.BurnBps))
		i--
		dAtA[i] = 0x30
	}
	if m.ValidatorIncentivesBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.ValidatorIncentivesBps))
		i--
		dAtA[i] = 0x28
	}
	if m.EcosystemBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.EcosystemBps))
		i--
		dAtA[i] = 0x20
	}
	if m.ReserveBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.ReserveBps))
		i--
		dAtA[i] = 0x18
	}
	if len(m.TreasuryModule) > 0 {
		i -= len(m.TreasuryModule)
		copy(dAtA[i:], m.TreasuryModule)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.TreasuryModule)))
		i--
		dAtA[i] = 0x12
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

func (m *TreasuryAllocations) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *TreasuryAllocations) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *TreasuryAllocations) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.TotalSpent) > 0 {
		for iNdEx := len(m.TotalSpent) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.TotalSpent[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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
	if len(m.TotalReceived) > 0 {
		for iNdEx := len(m.TotalReceived) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.TotalReceived[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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
	if len(m.BurnBalance) > 0 {
		for iNdEx := len(m.BurnBalance) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.BurnBalance[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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
	if len(m.ValidatorIncentiveBalance) > 0 {
		for iNdEx := len(m.ValidatorIncentiveBalance) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.ValidatorIncentiveBalance[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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
	if len(m.EcosystemBalance) > 0 {
		for iNdEx := len(m.EcosystemBalance) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.EcosystemBalance[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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
	if len(m.ReserveBalance) > 0 {
		for iNdEx := len(m.ReserveBalance) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.ReserveBalance[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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

func (m *TreasurySpend) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *TreasurySpend) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *TreasurySpend) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.ExecutedHeight != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.ExecutedHeight))
		i--
		dAtA[i] = 0x68
	}
	if m.UpdatedHeight != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.UpdatedHeight))
		i--
		dAtA[i] = 0x60
	}
	if m.CreatedHeight != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.CreatedHeight))
		i--
		dAtA[i] = 0x58
	}
	if len(m.Metadata) > 0 {
		i -= len(m.Metadata)
		copy(dAtA[i:], m.Metadata)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.Metadata)))
		i--
		dAtA[i] = 0x52
	}
	if m.VestingEndEpoch != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.VestingEndEpoch))
		i--
		dAtA[i] = 0x48
	}
	if m.VestingStartEpoch != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.VestingStartEpoch))
		i--
		dAtA[i] = 0x40
	}
	if m.Epoch != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.Epoch))
		i--
		dAtA[i] = 0x38
	}
	if len(m.Status) > 0 {
		i -= len(m.Status)
		copy(dAtA[i:], m.Status)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.Status)))
		i--
		dAtA[i] = 0x32
	}
	if len(m.Bucket) > 0 {
		i -= len(m.Bucket)
		copy(dAtA[i:], m.Bucket)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.Bucket)))
		i--
		dAtA[i] = 0x2a
	}
	if len(m.Amount) > 0 {
		for iNdEx := len(m.Amount) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Amount[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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
	if len(m.Recipient) > 0 {
		i -= len(m.Recipient)
		copy(dAtA[i:], m.Recipient)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.Recipient)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Proposer) > 0 {
		i -= len(m.Proposer)
		copy(dAtA[i:], m.Proposer)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.Proposer)))
		i--
		dAtA[i] = 0x12
	}
	if m.Id != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.Id))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *EpochSpend) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *EpochSpend) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *EpochSpend) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Spent) > 0 {
		for iNdEx := len(m.Spent) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Spent[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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
	l = m.Allocations.Size()
	n += 1 + l + sovGenesis(uint64(l))
	if len(m.Spends) > 0 {
		for _, e := range m.Spends {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.EpochSpends) > 0 {
		for _, e := range m.EpochSpends {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if m.NextSpendId != 0 {
		n += 1 + sovGenesis(uint64(m.NextSpendId))
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
	l = len(m.TreasuryModule)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	if m.ReserveBps != 0 {
		n += 1 + sovGenesis(uint64(m.ReserveBps))
	}
	if m.EcosystemBps != 0 {
		n += 1 + sovGenesis(uint64(m.EcosystemBps))
	}
	if m.ValidatorIncentivesBps != 0 {
		n += 1 + sovGenesis(uint64(m.ValidatorIncentivesBps))
	}
	if m.BurnBps != 0 {
		n += 1 + sovGenesis(uint64(m.BurnBps))
	}
	l = m.PerEpochSpendCap.Size()
	n += 1 + l + sovGenesis(uint64(l))
	if m.RecipientAllowlistEnabled {
		n += 2
	}
	if len(m.RecipientAllowlist) > 0 {
		for _, s := range m.RecipientAllowlist {
			l = len(s)
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if m.MaxMetadataBytes != 0 {
		n += 1 + sovGenesis(uint64(m.MaxMetadataBytes))
	}
	return n
}

func (m *TreasuryAllocations) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.ReserveBalance) > 0 {
		for _, e := range m.ReserveBalance {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.EcosystemBalance) > 0 {
		for _, e := range m.EcosystemBalance {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.ValidatorIncentiveBalance) > 0 {
		for _, e := range m.ValidatorIncentiveBalance {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.BurnBalance) > 0 {
		for _, e := range m.BurnBalance {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.TotalReceived) > 0 {
		for _, e := range m.TotalReceived {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.TotalSpent) > 0 {
		for _, e := range m.TotalSpent {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	return n
}

func (m *TreasurySpend) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Id != 0 {
		n += 1 + sovGenesis(uint64(m.Id))
	}
	l = len(m.Proposer)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	l = len(m.Recipient)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	if len(m.Amount) > 0 {
		for _, e := range m.Amount {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	l = len(m.Bucket)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	l = len(m.Status)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	if m.Epoch != 0 {
		n += 1 + sovGenesis(uint64(m.Epoch))
	}
	if m.VestingStartEpoch != 0 {
		n += 1 + sovGenesis(uint64(m.VestingStartEpoch))
	}
	if m.VestingEndEpoch != 0 {
		n += 1 + sovGenesis(uint64(m.VestingEndEpoch))
	}
	l = len(m.Metadata)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	if m.CreatedHeight != 0 {
		n += 1 + sovGenesis(uint64(m.CreatedHeight))
	}
	if m.UpdatedHeight != 0 {
		n += 1 + sovGenesis(uint64(m.UpdatedHeight))
	}
	if m.ExecutedHeight != 0 {
		n += 1 + sovGenesis(uint64(m.ExecutedHeight))
	}
	return n
}

func (m *EpochSpend) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Epoch != 0 {
		n += 1 + sovGenesis(uint64(m.Epoch))
	}
	if len(m.Spent) > 0 {
		for _, e := range m.Spent {
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
				return fmt.Errorf("proto: wrong wireType = %d for field Allocations", wireType)
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
			if err := m.Allocations.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Spends", wireType)
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
			m.Spends = append(m.Spends, TreasurySpend{})
			if err := m.Spends[len(m.Spends)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field EpochSpends", wireType)
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
			m.EpochSpends = append(m.EpochSpends, EpochSpend{})
			if err := m.EpochSpends[len(m.EpochSpends)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field NextSpendId", wireType)
			}
			m.NextSpendId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.NextSpendId |= uint64(b&0x7F) << shift
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
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ReserveBps", wireType)
			}
			m.ReserveBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ReserveBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field EcosystemBps", wireType)
			}
			m.EcosystemBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.EcosystemBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ValidatorIncentivesBps", wireType)
			}
			m.ValidatorIncentivesBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ValidatorIncentivesBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 6:
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
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PerEpochSpendCap", wireType)
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
			if err := m.PerEpochSpendCap.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 8:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field RecipientAllowlistEnabled", wireType)
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
			m.RecipientAllowlistEnabled = bool(v != 0)
		case 9:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RecipientAllowlist", wireType)
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
			m.RecipientAllowlist = append(m.RecipientAllowlist, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		case 10:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxMetadataBytes", wireType)
			}
			m.MaxMetadataBytes = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxMetadataBytes |= uint32(b&0x7F) << shift
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
func (m *TreasuryAllocations) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: TreasuryAllocations: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: TreasuryAllocations: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ReserveBalance", wireType)
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
			m.ReserveBalance = append(m.ReserveBalance, types.Coin{})
			if err := m.ReserveBalance[len(m.ReserveBalance)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field EcosystemBalance", wireType)
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
			m.EcosystemBalance = append(m.EcosystemBalance, types.Coin{})
			if err := m.EcosystemBalance[len(m.EcosystemBalance)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ValidatorIncentiveBalance", wireType)
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
			m.ValidatorIncentiveBalance = append(m.ValidatorIncentiveBalance, types.Coin{})
			if err := m.ValidatorIncentiveBalance[len(m.ValidatorIncentiveBalance)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BurnBalance", wireType)
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
			m.BurnBalance = append(m.BurnBalance, types.Coin{})
			if err := m.BurnBalance[len(m.BurnBalance)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TotalReceived", wireType)
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
			m.TotalReceived = append(m.TotalReceived, types.Coin{})
			if err := m.TotalReceived[len(m.TotalReceived)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TotalSpent", wireType)
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
			m.TotalSpent = append(m.TotalSpent, types.Coin{})
			if err := m.TotalSpent[len(m.TotalSpent)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
func (m *TreasurySpend) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: TreasurySpend: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: TreasurySpend: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Id", wireType)
			}
			m.Id = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Id |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Proposer", wireType)
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
			m.Proposer = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Recipient", wireType)
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
			m.Recipient = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Amount", wireType)
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
			m.Amount = append(m.Amount, types.Coin{})
			if err := m.Amount[len(m.Amount)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Bucket", wireType)
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
			m.Bucket = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Status", wireType)
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
			m.Status = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 7:
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
		case 8:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field VestingStartEpoch", wireType)
			}
			m.VestingStartEpoch = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.VestingStartEpoch |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 9:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field VestingEndEpoch", wireType)
			}
			m.VestingEndEpoch = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.VestingEndEpoch |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 10:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Metadata", wireType)
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
			m.Metadata = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 11:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field CreatedHeight", wireType)
			}
			m.CreatedHeight = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.CreatedHeight |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 12:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field UpdatedHeight", wireType)
			}
			m.UpdatedHeight = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.UpdatedHeight |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 13:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ExecutedHeight", wireType)
			}
			m.ExecutedHeight = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ExecutedHeight |= int64(b&0x7F) << shift
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
func (m *EpochSpend) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: EpochSpend: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: EpochSpend: illegal tag %d (wire type %d)", fieldNum, wire)
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
				return fmt.Errorf("proto: wrong wireType = %d for field Spent", wireType)
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
			m.Spent = append(m.Spent, types.Coin{})
			if err := m.Spent[len(m.Spent)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
