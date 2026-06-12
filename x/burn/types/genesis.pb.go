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
	BurnedByDenom	[]BurnedByDenomEntry	`protobuf:"bytes,2,rep,name=burned_by_denom,json=burnedByDenom,proto3" json:"burned_by_denom"`
	BurnedByEpoch	[]BurnedByEpochEntry	`protobuf:"bytes,3,rep,name=burned_by_epoch,json=burnedByEpoch,proto3" json:"burned_by_epoch"`
	BurnReasons	[]BurnReason		`protobuf:"bytes,4,rep,name=burn_reasons,json=burnReasons,proto3" json:"burn_reasons"`
}

func (m *GenesisState) Reset()		{ *m = GenesisState{} }
func (m *GenesisState) String() string	{ return proto.CompactTextString(m) }
func (*GenesisState) ProtoMessage()	{}
func (*GenesisState) Descriptor() ([]byte, []int) {
	return fileDescriptor_94a1fee0f654dbcf, []int{0}
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

func (m *GenesisState) GetBurnedByDenom() []BurnedByDenomEntry {
	if m != nil {
		return m.BurnedByDenom
	}
	return nil
}

func (m *GenesisState) GetBurnedByEpoch() []BurnedByEpochEntry {
	if m != nil {
		return m.BurnedByEpoch
	}
	return nil
}

func (m *GenesisState) GetBurnReasons() []BurnReason {
	if m != nil {
		return m.BurnReasons
	}
	return nil
}

type Params struct {
	AllowedDenoms		[]string		`protobuf:"bytes,1,rep,name=allowed_denoms,json=allowedDenoms,proto3" json:"allowed_denoms,omitempty"`
	ProtocolBurnPermissions	[]BurnPermission	`protobuf:"bytes,2,rep,name=protocol_burn_permissions,json=protocolBurnPermissions,proto3" json:"protocol_burn_permissions"`
	MaxReasonBytes		uint32			`protobuf:"varint,3,opt,name=max_reason_bytes,json=maxReasonBytes,proto3" json:"max_reason_bytes,omitempty"`
}

func (m *Params) Reset()		{ *m = Params{} }
func (m *Params) String() string	{ return proto.CompactTextString(m) }
func (*Params) ProtoMessage()		{}
func (*Params) Descriptor() ([]byte, []int) {
	return fileDescriptor_94a1fee0f654dbcf, []int{1}
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

func (m *Params) GetAllowedDenoms() []string {
	if m != nil {
		return m.AllowedDenoms
	}
	return nil
}

func (m *Params) GetProtocolBurnPermissions() []BurnPermission {
	if m != nil {
		return m.ProtocolBurnPermissions
	}
	return nil
}

func (m *Params) GetMaxReasonBytes() uint32 {
	if m != nil {
		return m.MaxReasonBytes
	}
	return 0
}

type BurnPermission struct {
	ModuleName	string		`protobuf:"bytes,1,opt,name=module_name,json=moduleName,proto3" json:"module_name,omitempty"`
	AllowedDenoms	[]string	`protobuf:"bytes,2,rep,name=allowed_denoms,json=allowedDenoms,proto3" json:"allowed_denoms,omitempty"`
}

func (m *BurnPermission) Reset()		{ *m = BurnPermission{} }
func (m *BurnPermission) String() string	{ return proto.CompactTextString(m) }
func (*BurnPermission) ProtoMessage()		{}
func (*BurnPermission) Descriptor() ([]byte, []int) {
	return fileDescriptor_94a1fee0f654dbcf, []int{2}
}
func (m *BurnPermission) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *BurnPermission) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_BurnPermission.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *BurnPermission) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BurnPermission.Merge(m, src)
}
func (m *BurnPermission) XXX_Size() int {
	return m.Size()
}
func (m *BurnPermission) XXX_DiscardUnknown() {
	xxx_messageInfo_BurnPermission.DiscardUnknown(m)
}

var xxx_messageInfo_BurnPermission proto.InternalMessageInfo

func (m *BurnPermission) GetModuleName() string {
	if m != nil {
		return m.ModuleName
	}
	return ""
}

func (m *BurnPermission) GetAllowedDenoms() []string {
	if m != nil {
		return m.AllowedDenoms
	}
	return nil
}

type BurnedByDenomEntry struct {
	Denom	string						`protobuf:"bytes,1,opt,name=denom,proto3" json:"denom,omitempty"`
	Amount	github_com_cosmos_cosmos_sdk_types.Coins	`protobuf:"bytes,2,rep,name=amount,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"amount"`
}

func (m *BurnedByDenomEntry) Reset()		{ *m = BurnedByDenomEntry{} }
func (m *BurnedByDenomEntry) String() string	{ return proto.CompactTextString(m) }
func (*BurnedByDenomEntry) ProtoMessage()	{}
func (*BurnedByDenomEntry) Descriptor() ([]byte, []int) {
	return fileDescriptor_94a1fee0f654dbcf, []int{3}
}
func (m *BurnedByDenomEntry) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *BurnedByDenomEntry) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_BurnedByDenomEntry.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *BurnedByDenomEntry) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BurnedByDenomEntry.Merge(m, src)
}
func (m *BurnedByDenomEntry) XXX_Size() int {
	return m.Size()
}
func (m *BurnedByDenomEntry) XXX_DiscardUnknown() {
	xxx_messageInfo_BurnedByDenomEntry.DiscardUnknown(m)
}

var xxx_messageInfo_BurnedByDenomEntry proto.InternalMessageInfo

func (m *BurnedByDenomEntry) GetDenom() string {
	if m != nil {
		return m.Denom
	}
	return ""
}

func (m *BurnedByDenomEntry) GetAmount() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.Amount
	}
	return nil
}

type BurnedByEpochEntry struct {
	Epoch	uint64						`protobuf:"varint,1,opt,name=epoch,proto3" json:"epoch,omitempty"`
	Amount	github_com_cosmos_cosmos_sdk_types.Coins	`protobuf:"bytes,2,rep,name=amount,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"amount"`
}

func (m *BurnedByEpochEntry) Reset()		{ *m = BurnedByEpochEntry{} }
func (m *BurnedByEpochEntry) String() string	{ return proto.CompactTextString(m) }
func (*BurnedByEpochEntry) ProtoMessage()	{}
func (*BurnedByEpochEntry) Descriptor() ([]byte, []int) {
	return fileDescriptor_94a1fee0f654dbcf, []int{4}
}
func (m *BurnedByEpochEntry) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *BurnedByEpochEntry) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_BurnedByEpochEntry.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *BurnedByEpochEntry) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BurnedByEpochEntry.Merge(m, src)
}
func (m *BurnedByEpochEntry) XXX_Size() int {
	return m.Size()
}
func (m *BurnedByEpochEntry) XXX_DiscardUnknown() {
	xxx_messageInfo_BurnedByEpochEntry.DiscardUnknown(m)
}

var xxx_messageInfo_BurnedByEpochEntry proto.InternalMessageInfo

func (m *BurnedByEpochEntry) GetEpoch() uint64 {
	if m != nil {
		return m.Epoch
	}
	return 0
}

func (m *BurnedByEpochEntry) GetAmount() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.Amount
	}
	return nil
}

type BurnReason struct {
	Id		uint64						`protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Epoch		uint64						`protobuf:"varint,2,opt,name=epoch,proto3" json:"epoch,omitempty"`
	Burner		string						`protobuf:"bytes,3,opt,name=burner,proto3" json:"burner,omitempty"`
	SourceModule	string						`protobuf:"bytes,4,opt,name=source_module,json=sourceModule,proto3" json:"source_module,omitempty"`
	Amount		github_com_cosmos_cosmos_sdk_types.Coins	`protobuf:"bytes,5,rep,name=amount,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"amount"`
	Reason		string						`protobuf:"bytes,6,opt,name=reason,proto3" json:"reason,omitempty"`
	Height		int64						`protobuf:"varint,7,opt,name=height,proto3" json:"height,omitempty"`
	Protocol	bool						`protobuf:"varint,8,opt,name=protocol,proto3" json:"protocol,omitempty"`
}

func (m *BurnReason) Reset()		{ *m = BurnReason{} }
func (m *BurnReason) String() string	{ return proto.CompactTextString(m) }
func (*BurnReason) ProtoMessage()	{}
func (*BurnReason) Descriptor() ([]byte, []int) {
	return fileDescriptor_94a1fee0f654dbcf, []int{5}
}
func (m *BurnReason) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *BurnReason) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_BurnReason.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *BurnReason) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BurnReason.Merge(m, src)
}
func (m *BurnReason) XXX_Size() int {
	return m.Size()
}
func (m *BurnReason) XXX_DiscardUnknown() {
	xxx_messageInfo_BurnReason.DiscardUnknown(m)
}

var xxx_messageInfo_BurnReason proto.InternalMessageInfo

func (m *BurnReason) GetId() uint64 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *BurnReason) GetEpoch() uint64 {
	if m != nil {
		return m.Epoch
	}
	return 0
}

func (m *BurnReason) GetBurner() string {
	if m != nil {
		return m.Burner
	}
	return ""
}

func (m *BurnReason) GetSourceModule() string {
	if m != nil {
		return m.SourceModule
	}
	return ""
}

func (m *BurnReason) GetAmount() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.Amount
	}
	return nil
}

func (m *BurnReason) GetReason() string {
	if m != nil {
		return m.Reason
	}
	return ""
}

func (m *BurnReason) GetHeight() int64 {
	if m != nil {
		return m.Height
	}
	return 0
}

func (m *BurnReason) GetProtocol() bool {
	if m != nil {
		return m.Protocol
	}
	return false
}

func init() {
	proto.RegisterType((*GenesisState)(nil), "l1.burn.v1.GenesisState")
	proto.RegisterType((*Params)(nil), "l1.burn.v1.Params")
	proto.RegisterType((*BurnPermission)(nil), "l1.burn.v1.BurnPermission")
	proto.RegisterType((*BurnedByDenomEntry)(nil), "l1.burn.v1.BurnedByDenomEntry")
	proto.RegisterType((*BurnedByEpochEntry)(nil), "l1.burn.v1.BurnedByEpochEntry")
	proto.RegisterType((*BurnReason)(nil), "l1.burn.v1.BurnReason")
}

func init()	{ proto.RegisterFile("l1/burn/v1/genesis.proto", fileDescriptor_94a1fee0f654dbcf) }

var fileDescriptor_94a1fee0f654dbcf = []byte{

	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xbc, 0x54, 0xcd, 0x6e, 0xd3, 0x4c,
	0x14, 0x8d, 0x9d, 0xd4, 0x5f, 0x73, 0xf3, 0xf3, 0xa1, 0x51, 0x55, 0xdc, 0x2c, 0x9c, 0x28, 0x08,
	0xe1, 0x4d, 0xed, 0xba, 0x3c, 0x00, 0x22, 0x50, 0xb1, 0x01, 0x54, 0x99, 0x0d, 0x42, 0x48, 0xd6,
	0xd8, 0x1e, 0x39, 0x16, 0xb6, 0x27, 0xf2, 0x38, 0x21, 0x79, 0x0a, 0x78, 0x01, 0x24, 0xd6, 0xbc,
	0x00, 0xaf, 0xd0, 0x65, 0x97, 0xac, 0x00, 0x25, 0x2f, 0x82, 0xe6, 0xa7, 0xf9, 0x29, 0x61, 0x09,
	0xab, 0xf8, 0x9e, 0x7b, 0xe7, 0xcc, 0x39, 0x73, 0x8f, 0x02, 0x66, 0xe6, 0xb9, 0xe1, 0xb4, 0x2c,
	0xdc, 0x99, 0xe7, 0x26, 0xa4, 0x20, 0x2c, 0x65, 0xce, 0xa4, 0xa4, 0x15, 0x45, 0x90, 0x79, 0x0e,
	0xef, 0x38, 0x33, 0xaf, 0x67, 0x45, 0x94, 0xe5, 0x94, 0xb9, 0x21, 0x66, 0xc4, 0x9d, 0x79, 0x21,
	0xa9, 0xb0, 0xe7, 0x46, 0x34, 0x2d, 0xe4, 0x6c, 0xef, 0x28, 0xa1, 0x09, 0x15, 0x9f, 0x2e, 0xff,
	0x92, 0xe8, 0xf0, 0x93, 0x0e, 0xed, 0x67, 0x92, 0xf3, 0x55, 0x85, 0x2b, 0x82, 0xce, 0xc0, 0x98,
	0xe0, 0x12, 0xe7, 0xcc, 0xd4, 0x06, 0x9a, 0xdd, 0x3a, 0x47, 0xce, 0xe6, 0x0e, 0xe7, 0x52, 0x74,
	0x46, 0x8d, 0xab, 0xef, 0xfd, 0x9a, 0xaf, 0xe6, 0xd0, 0x73, 0xf8, 0x9f, 0xf7, 0x49, 0x1c, 0x84,
	0x8b, 0x20, 0x26, 0x05, 0xcd, 0x4d, 0x7d, 0x50, 0xb7, 0x5b, 0xe7, 0xd6, 0xf6, 0xd1, 0x91, 0x18,
	0x19, 0x2d, 0x9e, 0xf2, 0x81, 0x8b, 0xa2, 0x2a, 0x17, 0x8a, 0xa6, 0x13, 0x6e, 0x77, 0x76, 0xd9,
	0xc8, 0x84, 0x46, 0x63, 0xb3, 0xfe, 0x67, 0xb6, 0x0b, 0x3e, 0xb0, 0x97, 0x4d, 0x74, 0xd0, 0x23,
	0x68, 0x73, 0x20, 0x28, 0x09, 0x66, 0xb4, 0x60, 0x66, 0x43, 0x50, 0x1d, 0xdf, 0xa6, 0xf2, 0x45,
	0x5b, 0x51, 0xb4, 0xc2, 0x35, 0xc2, 0x86, 0x5f, 0x35, 0x30, 0xa4, 0x6b, 0x74, 0x1f, 0xba, 0x38,
	0xcb, 0xe8, 0x7b, 0x12, 0x4b, 0x97, 0xfc, 0x85, 0xea, 0x76, 0xd3, 0xef, 0x28, 0x54, 0xe8, 0x67,
	0xe8, 0x2d, 0x9c, 0x88, 0xa7, 0x8d, 0x68, 0x16, 0x88, 0xbb, 0x27, 0xa4, 0xcc, 0x53, 0xc6, 0x52,
	0x7e, 0xbf, 0x7c, 0x98, 0xde, 0xed, 0xfb, 0x2f, 0xd7, 0x23, 0x4a, 0xc3, 0xdd, 0x1b, 0x8a, 0xdd,
	0x2e, 0x43, 0x36, 0xdc, 0xc9, 0xf1, 0x5c, 0xf9, 0x09, 0xc2, 0x45, 0x45, 0x98, 0x59, 0x1f, 0x68,
	0x76, 0xc7, 0xef, 0xe6, 0x78, 0xae, 0x7c, 0x70, 0x74, 0xf8, 0x1a, 0xba, 0xbb, 0x87, 0x51, 0x1f,
	0x5a, 0x39, 0x8d, 0xa7, 0x19, 0x09, 0x0a, 0x9c, 0x13, 0xb1, 0xdf, 0xa6, 0x0f, 0x12, 0x7a, 0x89,
	0x73, 0xb2, 0xc7, 0xa1, 0xbe, 0xc7, 0xe1, 0xf0, 0x83, 0x06, 0xe8, 0xf7, 0x75, 0xa2, 0x23, 0x38,
	0x90, 0xdb, 0x97, 0xc4, 0xb2, 0x40, 0x11, 0x18, 0x38, 0xa7, 0xd3, 0xa2, 0x52, 0xde, 0x4f, 0x1c,
	0x99, 0x53, 0x87, 0xe7, 0xd4, 0x51, 0x39, 0x75, 0x9e, 0xd0, 0xb4, 0x18, 0x9d, 0x71, 0xeb, 0x5f,
	0x7e, 0xf4, 0xed, 0x24, 0xad, 0xc6, 0xd3, 0xd0, 0x89, 0x68, 0xee, 0xaa, 0x50, 0xcb, 0x9f, 0x53,
	0x16, 0xbf, 0x73, 0xab, 0xc5, 0x84, 0x30, 0x71, 0x80, 0xf9, 0x8a, 0x7a, 0x47, 0xd1, 0x26, 0x12,
	0x5c, 0x91, 0x4c, 0x10, 0x57, 0xd4, 0xf0, 0x65, 0xf1, 0x6f, 0x14, 0x7d, 0xd6, 0x01, 0x36, 0xc9,
	0x42, 0x5d, 0xd0, 0xd3, 0x58, 0xc9, 0xd0, 0xd3, 0x78, 0xa3, 0x4c, 0xdf, 0x56, 0x76, 0x0c, 0x86,
	0x88, 0x6f, 0x29, 0x56, 0xda, 0xf4, 0x55, 0x85, 0xee, 0x41, 0x87, 0xd1, 0x69, 0x19, 0x91, 0x40,
	0x2e, 0xcb, 0x6c, 0x88, 0x76, 0x5b, 0x82, 0x2f, 0x04, 0xb6, 0x65, 0xeb, 0xe0, 0xaf, 0xd9, 0xe2,
	0x0a, 0x65, 0xf4, 0x4c, 0x43, 0x2a, 0x94, 0x15, 0xc7, 0xc7, 0x24, 0x4d, 0xc6, 0x95, 0xf9, 0xdf,
	0x40, 0xb3, 0xeb, 0xbe, 0xaa, 0x50, 0x0f, 0x0e, 0x6f, 0x92, 0x6c, 0x1e, 0x0e, 0x34, 0xfb, 0xd0,
	0x5f, 0xd7, 0xa3, 0xc7, 0x57, 0x4b, 0x4b, 0xbb, 0x5e, 0x5a, 0xda, 0xcf, 0xa5, 0xa5, 0x7d, 0x5c,
	0x59, 0xb5, 0xeb, 0x95, 0x55, 0xfb, 0xb6, 0xb2, 0x6a, 0x6f, 0x1e, 0x6c, 0xe9, 0x62, 0x74, 0x46,
	0x4a, 0x92, 0x26, 0xc5, 0x69, 0xe6, 0xb9, 0x99, 0xe7, 0xce, 0xe5, 0x5f, 0xa1, 0x10, 0x17, 0x1a,
	0x82, 0xec, 0xe1, 0xaf, 0x00, 0x00, 0x00, 0xff, 0xff, 0x87, 0xce, 0x34, 0x90, 0x22, 0x05, 0x00,
	0x00,
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
	if len(m.BurnReasons) > 0 {
		for iNdEx := len(m.BurnReasons) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.BurnReasons[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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
	if len(m.BurnedByEpoch) > 0 {
		for iNdEx := len(m.BurnedByEpoch) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.BurnedByEpoch[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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
	if len(m.BurnedByDenom) > 0 {
		for iNdEx := len(m.BurnedByDenom) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.BurnedByDenom[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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
	if m.MaxReasonBytes != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.MaxReasonBytes))
		i--
		dAtA[i] = 0x18
	}
	if len(m.ProtocolBurnPermissions) > 0 {
		for iNdEx := len(m.ProtocolBurnPermissions) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.ProtocolBurnPermissions[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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
	if len(m.AllowedDenoms) > 0 {
		for iNdEx := len(m.AllowedDenoms) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.AllowedDenoms[iNdEx])
			copy(dAtA[i:], m.AllowedDenoms[iNdEx])
			i = encodeVarintGenesis(dAtA, i, uint64(len(m.AllowedDenoms[iNdEx])))
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *BurnPermission) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *BurnPermission) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *BurnPermission) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.AllowedDenoms) > 0 {
		for iNdEx := len(m.AllowedDenoms) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.AllowedDenoms[iNdEx])
			copy(dAtA[i:], m.AllowedDenoms[iNdEx])
			i = encodeVarintGenesis(dAtA, i, uint64(len(m.AllowedDenoms[iNdEx])))
			i--
			dAtA[i] = 0x12
		}
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

func (m *BurnedByDenomEntry) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *BurnedByDenomEntry) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *BurnedByDenomEntry) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
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
			dAtA[i] = 0x12
		}
	}
	if len(m.Denom) > 0 {
		i -= len(m.Denom)
		copy(dAtA[i:], m.Denom)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.Denom)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *BurnedByEpochEntry) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *BurnedByEpochEntry) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *BurnedByEpochEntry) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
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

func (m *BurnReason) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *BurnReason) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *BurnReason) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Protocol {
		i--
		if m.Protocol {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x40
	}
	if m.Height != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.Height))
		i--
		dAtA[i] = 0x38
	}
	if len(m.Reason) > 0 {
		i -= len(m.Reason)
		copy(dAtA[i:], m.Reason)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.Reason)))
		i--
		dAtA[i] = 0x32
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
			dAtA[i] = 0x2a
		}
	}
	if len(m.SourceModule) > 0 {
		i -= len(m.SourceModule)
		copy(dAtA[i:], m.SourceModule)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.SourceModule)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.Burner) > 0 {
		i -= len(m.Burner)
		copy(dAtA[i:], m.Burner)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.Burner)))
		i--
		dAtA[i] = 0x1a
	}
	if m.Epoch != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.Epoch))
		i--
		dAtA[i] = 0x10
	}
	if m.Id != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.Id))
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
	if len(m.BurnedByDenom) > 0 {
		for _, e := range m.BurnedByDenom {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.BurnedByEpoch) > 0 {
		for _, e := range m.BurnedByEpoch {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.BurnReasons) > 0 {
		for _, e := range m.BurnReasons {
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
	if len(m.AllowedDenoms) > 0 {
		for _, s := range m.AllowedDenoms {
			l = len(s)
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.ProtocolBurnPermissions) > 0 {
		for _, e := range m.ProtocolBurnPermissions {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if m.MaxReasonBytes != 0 {
		n += 1 + sovGenesis(uint64(m.MaxReasonBytes))
	}
	return n
}

func (m *BurnPermission) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ModuleName)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	if len(m.AllowedDenoms) > 0 {
		for _, s := range m.AllowedDenoms {
			l = len(s)
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	return n
}

func (m *BurnedByDenomEntry) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Denom)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	if len(m.Amount) > 0 {
		for _, e := range m.Amount {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	return n
}

func (m *BurnedByEpochEntry) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Epoch != 0 {
		n += 1 + sovGenesis(uint64(m.Epoch))
	}
	if len(m.Amount) > 0 {
		for _, e := range m.Amount {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	return n
}

func (m *BurnReason) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Id != 0 {
		n += 1 + sovGenesis(uint64(m.Id))
	}
	if m.Epoch != 0 {
		n += 1 + sovGenesis(uint64(m.Epoch))
	}
	l = len(m.Burner)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	l = len(m.SourceModule)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	if len(m.Amount) > 0 {
		for _, e := range m.Amount {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	l = len(m.Reason)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	if m.Height != 0 {
		n += 1 + sovGenesis(uint64(m.Height))
	}
	if m.Protocol {
		n += 2
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
				return fmt.Errorf("proto: wrong wireType = %d for field BurnedByDenom", wireType)
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
			m.BurnedByDenom = append(m.BurnedByDenom, BurnedByDenomEntry{})
			if err := m.BurnedByDenom[len(m.BurnedByDenom)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BurnedByEpoch", wireType)
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
			m.BurnedByEpoch = append(m.BurnedByEpoch, BurnedByEpochEntry{})
			if err := m.BurnedByEpoch[len(m.BurnedByEpoch)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BurnReasons", wireType)
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
			m.BurnReasons = append(m.BurnReasons, BurnReason{})
			if err := m.BurnReasons[len(m.BurnReasons)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
				return fmt.Errorf("proto: wrong wireType = %d for field AllowedDenoms", wireType)
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
			m.AllowedDenoms = append(m.AllowedDenoms, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ProtocolBurnPermissions", wireType)
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
			m.ProtocolBurnPermissions = append(m.ProtocolBurnPermissions, BurnPermission{})
			if err := m.ProtocolBurnPermissions[len(m.ProtocolBurnPermissions)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxReasonBytes", wireType)
			}
			m.MaxReasonBytes = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxReasonBytes |= uint32(b&0x7F) << shift
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
func (m *BurnPermission) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: BurnPermission: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: BurnPermission: illegal tag %d (wire type %d)", fieldNum, wire)
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
				return fmt.Errorf("proto: wrong wireType = %d for field AllowedDenoms", wireType)
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
			m.AllowedDenoms = append(m.AllowedDenoms, string(dAtA[iNdEx:postIndex]))
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
func (m *BurnedByDenomEntry) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: BurnedByDenomEntry: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: BurnedByDenomEntry: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Denom", wireType)
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
			m.Denom = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
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
func (m *BurnedByEpochEntry) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: BurnedByEpochEntry: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: BurnedByEpochEntry: illegal tag %d (wire type %d)", fieldNum, wire)
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
func (m *BurnReason) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: BurnReason: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: BurnReason: illegal tag %d (wire type %d)", fieldNum, wire)
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
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Burner", wireType)
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
			m.Burner = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SourceModule", wireType)
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
			m.SourceModule = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
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
		case 6:
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
		case 7:
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
				m.Height |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 8:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Protocol", wireType)
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
			m.Protocol = bool(v != 0)
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
