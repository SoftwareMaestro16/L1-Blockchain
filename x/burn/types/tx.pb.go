package types

import (
	context "context"
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/cosmos-sdk/types/msgservice"
	_ "github.com/cosmos/cosmos-sdk/types/tx/amino"
	_ "github.com/cosmos/gogoproto/gogoproto"
	grpc1 "github.com/cosmos/gogoproto/grpc"
	proto "github.com/cosmos/gogoproto/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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

type MsgBurnProtocolCoins struct {
	Authority	string						`protobuf:"bytes,1,opt,name=authority,proto3" json:"authority,omitempty"`
	SourceModule	string						`protobuf:"bytes,2,opt,name=source_module,json=sourceModule,proto3" json:"source_module,omitempty"`
	Amount		github_com_cosmos_cosmos_sdk_types.Coins	`protobuf:"bytes,3,rep,name=amount,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"amount"`
	Epoch		uint64						`protobuf:"varint,4,opt,name=epoch,proto3" json:"epoch,omitempty"`
	Reason		string						`protobuf:"bytes,5,opt,name=reason,proto3" json:"reason,omitempty"`
}

func (m *MsgBurnProtocolCoins) Reset()		{ *m = MsgBurnProtocolCoins{} }
func (m *MsgBurnProtocolCoins) String() string	{ return proto.CompactTextString(m) }
func (*MsgBurnProtocolCoins) ProtoMessage()	{}
func (*MsgBurnProtocolCoins) Descriptor() ([]byte, []int) {
	return fileDescriptor_441139a801c0ac00, []int{0}
}
func (m *MsgBurnProtocolCoins) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgBurnProtocolCoins) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgBurnProtocolCoins.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgBurnProtocolCoins) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgBurnProtocolCoins.Merge(m, src)
}
func (m *MsgBurnProtocolCoins) XXX_Size() int {
	return m.Size()
}
func (m *MsgBurnProtocolCoins) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgBurnProtocolCoins.DiscardUnknown(m)
}

var xxx_messageInfo_MsgBurnProtocolCoins proto.InternalMessageInfo

func (m *MsgBurnProtocolCoins) GetAuthority() string {
	if m != nil {
		return m.Authority
	}
	return ""
}

func (m *MsgBurnProtocolCoins) GetSourceModule() string {
	if m != nil {
		return m.SourceModule
	}
	return ""
}

func (m *MsgBurnProtocolCoins) GetAmount() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.Amount
	}
	return nil
}

func (m *MsgBurnProtocolCoins) GetEpoch() uint64 {
	if m != nil {
		return m.Epoch
	}
	return 0
}

func (m *MsgBurnProtocolCoins) GetReason() string {
	if m != nil {
		return m.Reason
	}
	return ""
}

type MsgBurnProtocolCoinsResponse struct {
	Burn BurnReason `protobuf:"bytes,1,opt,name=burn,proto3" json:"burn"`
}

func (m *MsgBurnProtocolCoinsResponse) Reset()		{ *m = MsgBurnProtocolCoinsResponse{} }
func (m *MsgBurnProtocolCoinsResponse) String() string	{ return proto.CompactTextString(m) }
func (*MsgBurnProtocolCoinsResponse) ProtoMessage()	{}
func (*MsgBurnProtocolCoinsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_441139a801c0ac00, []int{1}
}
func (m *MsgBurnProtocolCoinsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgBurnProtocolCoinsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgBurnProtocolCoinsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgBurnProtocolCoinsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgBurnProtocolCoinsResponse.Merge(m, src)
}
func (m *MsgBurnProtocolCoinsResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgBurnProtocolCoinsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgBurnProtocolCoinsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgBurnProtocolCoinsResponse proto.InternalMessageInfo

func (m *MsgBurnProtocolCoinsResponse) GetBurn() BurnReason {
	if m != nil {
		return m.Burn
	}
	return BurnReason{}
}

type MsgBurnUserCoins struct {
	Burner	string						`protobuf:"bytes,1,opt,name=burner,proto3" json:"burner,omitempty"`
	Amount	github_com_cosmos_cosmos_sdk_types.Coins	`protobuf:"bytes,2,rep,name=amount,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"amount"`
	Epoch	uint64						`protobuf:"varint,3,opt,name=epoch,proto3" json:"epoch,omitempty"`
	Reason	string						`protobuf:"bytes,4,opt,name=reason,proto3" json:"reason,omitempty"`
}

func (m *MsgBurnUserCoins) Reset()		{ *m = MsgBurnUserCoins{} }
func (m *MsgBurnUserCoins) String() string	{ return proto.CompactTextString(m) }
func (*MsgBurnUserCoins) ProtoMessage()		{}
func (*MsgBurnUserCoins) Descriptor() ([]byte, []int) {
	return fileDescriptor_441139a801c0ac00, []int{2}
}
func (m *MsgBurnUserCoins) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgBurnUserCoins) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgBurnUserCoins.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgBurnUserCoins) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgBurnUserCoins.Merge(m, src)
}
func (m *MsgBurnUserCoins) XXX_Size() int {
	return m.Size()
}
func (m *MsgBurnUserCoins) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgBurnUserCoins.DiscardUnknown(m)
}

var xxx_messageInfo_MsgBurnUserCoins proto.InternalMessageInfo

func (m *MsgBurnUserCoins) GetBurner() string {
	if m != nil {
		return m.Burner
	}
	return ""
}

func (m *MsgBurnUserCoins) GetAmount() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.Amount
	}
	return nil
}

func (m *MsgBurnUserCoins) GetEpoch() uint64 {
	if m != nil {
		return m.Epoch
	}
	return 0
}

func (m *MsgBurnUserCoins) GetReason() string {
	if m != nil {
		return m.Reason
	}
	return ""
}

type MsgBurnUserCoinsResponse struct {
	Burn BurnReason `protobuf:"bytes,1,opt,name=burn,proto3" json:"burn"`
}

func (m *MsgBurnUserCoinsResponse) Reset()		{ *m = MsgBurnUserCoinsResponse{} }
func (m *MsgBurnUserCoinsResponse) String() string	{ return proto.CompactTextString(m) }
func (*MsgBurnUserCoinsResponse) ProtoMessage()		{}
func (*MsgBurnUserCoinsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_441139a801c0ac00, []int{3}
}
func (m *MsgBurnUserCoinsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgBurnUserCoinsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgBurnUserCoinsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgBurnUserCoinsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgBurnUserCoinsResponse.Merge(m, src)
}
func (m *MsgBurnUserCoinsResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgBurnUserCoinsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgBurnUserCoinsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgBurnUserCoinsResponse proto.InternalMessageInfo

func (m *MsgBurnUserCoinsResponse) GetBurn() BurnReason {
	if m != nil {
		return m.Burn
	}
	return BurnReason{}
}

type MsgUpdateBurnParams struct {
	Authority	string	`protobuf:"bytes,1,opt,name=authority,proto3" json:"authority,omitempty"`
	Params		Params	`protobuf:"bytes,2,opt,name=params,proto3" json:"params"`
}

func (m *MsgUpdateBurnParams) Reset()		{ *m = MsgUpdateBurnParams{} }
func (m *MsgUpdateBurnParams) String() string	{ return proto.CompactTextString(m) }
func (*MsgUpdateBurnParams) ProtoMessage()	{}
func (*MsgUpdateBurnParams) Descriptor() ([]byte, []int) {
	return fileDescriptor_441139a801c0ac00, []int{4}
}
func (m *MsgUpdateBurnParams) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgUpdateBurnParams) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgUpdateBurnParams.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgUpdateBurnParams) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgUpdateBurnParams.Merge(m, src)
}
func (m *MsgUpdateBurnParams) XXX_Size() int {
	return m.Size()
}
func (m *MsgUpdateBurnParams) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgUpdateBurnParams.DiscardUnknown(m)
}

var xxx_messageInfo_MsgUpdateBurnParams proto.InternalMessageInfo

func (m *MsgUpdateBurnParams) GetAuthority() string {
	if m != nil {
		return m.Authority
	}
	return ""
}

func (m *MsgUpdateBurnParams) GetParams() Params {
	if m != nil {
		return m.Params
	}
	return Params{}
}

type MsgUpdateBurnParamsResponse struct {
}

func (m *MsgUpdateBurnParamsResponse) Reset()		{ *m = MsgUpdateBurnParamsResponse{} }
func (m *MsgUpdateBurnParamsResponse) String() string	{ return proto.CompactTextString(m) }
func (*MsgUpdateBurnParamsResponse) ProtoMessage()	{}
func (*MsgUpdateBurnParamsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_441139a801c0ac00, []int{5}
}
func (m *MsgUpdateBurnParamsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgUpdateBurnParamsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgUpdateBurnParamsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgUpdateBurnParamsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgUpdateBurnParamsResponse.Merge(m, src)
}
func (m *MsgUpdateBurnParamsResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgUpdateBurnParamsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgUpdateBurnParamsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgUpdateBurnParamsResponse proto.InternalMessageInfo

func init() {
	proto.RegisterType((*MsgBurnProtocolCoins)(nil), "l1.burn.v1.MsgBurnProtocolCoins")
	proto.RegisterType((*MsgBurnProtocolCoinsResponse)(nil), "l1.burn.v1.MsgBurnProtocolCoinsResponse")
	proto.RegisterType((*MsgBurnUserCoins)(nil), "l1.burn.v1.MsgBurnUserCoins")
	proto.RegisterType((*MsgBurnUserCoinsResponse)(nil), "l1.burn.v1.MsgBurnUserCoinsResponse")
	proto.RegisterType((*MsgUpdateBurnParams)(nil), "l1.burn.v1.MsgUpdateBurnParams")
	proto.RegisterType((*MsgUpdateBurnParamsResponse)(nil), "l1.burn.v1.MsgUpdateBurnParamsResponse")
}

func init()	{ proto.RegisterFile("l1/burn/v1/tx.proto", fileDescriptor_441139a801c0ac00) }

var fileDescriptor_441139a801c0ac00 = []byte{

	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xbc, 0x54, 0x31, 0x6b, 0xdb, 0x40,
	0x14, 0xb6, 0x6c, 0xc7, 0x90, 0x4b, 0x03, 0x89, 0x62, 0x52, 0x45, 0x49, 0x15, 0xe3, 0x16, 0x22,
	0x02, 0xd6, 0x59, 0x2e, 0x74, 0xc8, 0x16, 0x77, 0xad, 0x21, 0x28, 0x64, 0x29, 0x05, 0x23, 0xcb,
	0x87, 0x2c, 0x2a, 0xe9, 0xc4, 0x9d, 0x64, 0x92, 0xad, 0xed, 0xd8, 0xa9, 0x3f, 0xa1, 0x73, 0x97,
	0x7a, 0x28, 0xf4, 0x2f, 0x64, 0x0c, 0x9d, 0x3a, 0xb5, 0xc5, 0x1e, 0xfc, 0x37, 0xca, 0x9d, 0xce,
	0xb6, 0x62, 0x8b, 0x1a, 0x5a, 0xe8, 0x62, 0xeb, 0xbe, 0xef, 0xbd, 0x77, 0xef, 0x7d, 0xef, 0x93,
	0xc0, 0x9e, 0x6f, 0xc2, 0x5e, 0x42, 0x42, 0x38, 0x34, 0x61, 0x7c, 0x6d, 0x44, 0x04, 0xc7, 0x58,
	0x06, 0xbe, 0x69, 0x30, 0xd0, 0x18, 0x9a, 0xea, 0xae, 0x1d, 0x78, 0x21, 0x86, 0xfc, 0x37, 0xa5,
	0x55, 0xcd, 0xc1, 0x34, 0xc0, 0x14, 0xf6, 0x6c, 0x8a, 0xe0, 0xd0, 0xec, 0xa1, 0xd8, 0x36, 0xa1,
	0x83, 0xbd, 0x50, 0xf0, 0x0f, 0x05, 0x1f, 0x50, 0x97, 0x95, 0x0d, 0xa8, 0x2b, 0x88, 0x83, 0x94,
	0xe8, 0xf2, 0x13, 0x4c, 0x0f, 0x82, 0xaa, 0xba, 0xd8, 0xc5, 0x29, 0xce, 0x9e, 0x04, 0xaa, 0x64,
	0xba, 0x73, 0x51, 0x88, 0xa8, 0x27, 0xe2, 0xeb, 0x5f, 0x8b, 0xa0, 0xda, 0xa1, 0x6e, 0x3b, 0x21,
	0xe1, 0x05, 0x03, 0x1c, 0xec, 0x3f, 0xc7, 0x5e, 0x48, 0xe5, 0x67, 0x60, 0xd3, 0x4e, 0xe2, 0x01,
	0x26, 0x5e, 0x7c, 0xa3, 0x48, 0x35, 0x49, 0xdf, 0x6c, 0x2b, 0xdf, 0xbe, 0x34, 0xaa, 0xe2, 0xb6,
	0xf3, 0x7e, 0x9f, 0x20, 0x4a, 0x2f, 0x63, 0xe2, 0x85, 0xae, 0xb5, 0x08, 0x95, 0x1f, 0x83, 0x6d,
	0x8a, 0x13, 0xe2, 0xa0, 0x6e, 0x80, 0xfb, 0x89, 0x8f, 0x94, 0x22, 0xcb, 0xb5, 0x1e, 0xa4, 0x60,
	0x87, 0x63, 0xb2, 0x03, 0x2a, 0x76, 0x80, 0x93, 0x30, 0x56, 0x4a, 0xb5, 0x92, 0xbe, 0xd5, 0x3a,
	0x30, 0x44, 0x59, 0x26, 0x85, 0x21, 0xa4, 0x30, 0x58, 0x23, 0xed, 0xe6, 0xed, 0x8f, 0xe3, 0xc2,
	0xa7, 0x9f, 0xc7, 0xba, 0xeb, 0xc5, 0x83, 0xa4, 0x67, 0x38, 0x38, 0x10, 0x13, 0x8b, 0xbf, 0x06,
	0xed, 0xbf, 0x86, 0xf1, 0x4d, 0x84, 0x28, 0x4f, 0xa0, 0x96, 0x28, 0x2d, 0x57, 0xc1, 0x06, 0x8a,
	0xb0, 0x33, 0x50, 0xca, 0x35, 0x49, 0x2f, 0x5b, 0xe9, 0x41, 0xde, 0x07, 0x15, 0x82, 0x6c, 0x8a,
	0x43, 0x65, 0x83, 0x37, 0x26, 0x4e, 0x67, 0xf0, 0xdd, 0x74, 0x74, 0xba, 0x98, 0xe3, 0xfd, 0x74,
	0x74, 0x7a, 0x34, 0x53, 0x2d, 0x4f, 0xa0, 0xfa, 0x05, 0x38, 0xca, 0xc3, 0x2d, 0x44, 0x23, 0x1c,
	0x52, 0x24, 0x37, 0x41, 0x99, 0x25, 0x73, 0xed, 0xb6, 0x5a, 0xfb, 0xc6, 0xc2, 0x0b, 0x06, 0x4b,
	0xb2, 0xf8, 0xb5, 0xed, 0x32, 0x1b, 0xcf, 0xe2, 0x91, 0xf5, 0xb7, 0x45, 0xb0, 0x23, 0x4a, 0x5e,
	0x51, 0x44, 0xd2, 0x3d, 0x34, 0x41, 0x85, 0x91, 0x88, 0xac, 0x5d, 0x82, 0x88, 0xcb, 0x88, 0x5b,
	0xfc, 0x0f, 0xe2, 0x96, 0xf2, 0xc5, 0x2d, 0xdf, 0x13, 0x57, 0x67, 0xe2, 0x8a, 0xfe, 0x98, 0xb2,
	0xca, 0x92, 0xb2, 0xf3, 0x71, 0xeb, 0x2f, 0x80, 0xb2, 0x8c, 0xfd, 0x83, 0xa2, 0x9f, 0x25, 0xb0,
	0xd7, 0xa1, 0xee, 0x55, 0xd4, 0xb7, 0x63, 0xc4, 0x57, 0x65, 0x13, 0x3b, 0xf8, 0x7b, 0x73, 0x37,
	0x41, 0x25, 0xe2, 0x15, 0xb8, 0xab, 0xb7, 0x5a, 0x72, 0xb6, 0x87, 0xb4, 0xb6, 0xb8, 0x5f, 0xc4,
	0x9d, 0x19, 0xab, 0xb6, 0x3a, 0xcc, 0x0c, 0xbf, 0xdc, 0x59, 0xfd, 0x11, 0x38, 0xcc, 0x81, 0x67,
	0x12, 0xb4, 0x3e, 0x16, 0x41, 0xa9, 0x43, 0x5d, 0xb9, 0x0b, 0x76, 0x57, 0x5f, 0xd9, 0x5a, 0xb6,
	0x9b, 0x3c, 0x6f, 0xaa, 0xfa, 0xba, 0x88, 0xb9, 0xd6, 0x97, 0x60, 0xfb, 0xbe, 0x0f, 0x8f, 0x72,
	0x52, 0xe7, 0xac, 0xfa, 0xe4, 0x4f, 0xec, 0xbc, 0xe8, 0x2b, 0xb0, 0xb3, 0xb2, 0x8a, 0xe3, 0xa5,
	0xcc, 0xe5, 0x00, 0xf5, 0x64, 0x4d, 0xc0, 0xac, 0xba, 0xba, 0xf1, 0x66, 0x3a, 0x3a, 0x95, 0xda,
	0xe7, 0xb7, 0x63, 0x4d, 0xba, 0x1b, 0x6b, 0xd2, 0xaf, 0xb1, 0x26, 0x7d, 0x98, 0x68, 0x85, 0xbb,
	0x89, 0x56, 0xf8, 0x3e, 0xd1, 0x0a, 0x2f, 0x4f, 0x32, 0x2e, 0xa7, 0x78, 0x88, 0x08, 0xf2, 0xdc,
	0xb0, 0xe1, 0x9b, 0xd0, 0x37, 0xe1, 0x75, 0xba, 0x12, 0x6e, 0xf5, 0x5e, 0x85, 0x7f, 0x1b, 0x9f,
	0xfe, 0x0e, 0x00, 0x00, 0xff, 0xff, 0x14, 0xa2, 0xc8, 0xbe, 0xd5, 0x05, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// MsgClient is the client API for Msg service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type MsgClient interface {
	BurnProtocolCoins(ctx context.Context, in *MsgBurnProtocolCoins, opts ...grpc.CallOption) (*MsgBurnProtocolCoinsResponse, error)
	BurnUserCoins(ctx context.Context, in *MsgBurnUserCoins, opts ...grpc.CallOption) (*MsgBurnUserCoinsResponse, error)
	UpdateBurnParams(ctx context.Context, in *MsgUpdateBurnParams, opts ...grpc.CallOption) (*MsgUpdateBurnParamsResponse, error)
}

type msgClient struct {
	cc grpc1.ClientConn
}

func NewMsgClient(cc grpc1.ClientConn) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) BurnProtocolCoins(ctx context.Context, in *MsgBurnProtocolCoins, opts ...grpc.CallOption) (*MsgBurnProtocolCoinsResponse, error) {
	out := new(MsgBurnProtocolCoinsResponse)
	err := c.cc.Invoke(ctx, "/l1.burn.v1.Msg/BurnProtocolCoins", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) BurnUserCoins(ctx context.Context, in *MsgBurnUserCoins, opts ...grpc.CallOption) (*MsgBurnUserCoinsResponse, error) {
	out := new(MsgBurnUserCoinsResponse)
	err := c.cc.Invoke(ctx, "/l1.burn.v1.Msg/BurnUserCoins", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) UpdateBurnParams(ctx context.Context, in *MsgUpdateBurnParams, opts ...grpc.CallOption) (*MsgUpdateBurnParamsResponse, error) {
	out := new(MsgUpdateBurnParamsResponse)
	err := c.cc.Invoke(ctx, "/l1.burn.v1.Msg/UpdateBurnParams", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MsgServer is the server API for Msg service.
type MsgServer interface {
	BurnProtocolCoins(context.Context, *MsgBurnProtocolCoins) (*MsgBurnProtocolCoinsResponse, error)
	BurnUserCoins(context.Context, *MsgBurnUserCoins) (*MsgBurnUserCoinsResponse, error)
	UpdateBurnParams(context.Context, *MsgUpdateBurnParams) (*MsgUpdateBurnParamsResponse, error)
}

// UnimplementedMsgServer can be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (*UnimplementedMsgServer) BurnProtocolCoins(ctx context.Context, req *MsgBurnProtocolCoins) (*MsgBurnProtocolCoinsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BurnProtocolCoins not implemented")
}
func (*UnimplementedMsgServer) BurnUserCoins(ctx context.Context, req *MsgBurnUserCoins) (*MsgBurnUserCoinsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BurnUserCoins not implemented")
}
func (*UnimplementedMsgServer) UpdateBurnParams(ctx context.Context, req *MsgUpdateBurnParams) (*MsgUpdateBurnParamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateBurnParams not implemented")
}

func RegisterMsgServer(s grpc1.Server, srv MsgServer) {
	s.RegisterService(&_Msg_serviceDesc, srv)
}

func _Msg_BurnProtocolCoins_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgBurnProtocolCoins)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).BurnProtocolCoins(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.burn.v1.Msg/BurnProtocolCoins",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).BurnProtocolCoins(ctx, req.(*MsgBurnProtocolCoins))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_BurnUserCoins_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgBurnUserCoins)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).BurnUserCoins(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.burn.v1.Msg/BurnUserCoins",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).BurnUserCoins(ctx, req.(*MsgBurnUserCoins))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_UpdateBurnParams_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUpdateBurnParams)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).UpdateBurnParams(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.burn.v1.Msg/UpdateBurnParams",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).UpdateBurnParams(ctx, req.(*MsgUpdateBurnParams))
	}
	return interceptor(ctx, in, info, handler)
}

var Msg_serviceDesc = _Msg_serviceDesc
var _Msg_serviceDesc = grpc.ServiceDesc{
	ServiceName:	"l1.burn.v1.Msg",
	HandlerType:	(*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName:	"BurnProtocolCoins",
			Handler:	_Msg_BurnProtocolCoins_Handler,
		},
		{
			MethodName:	"BurnUserCoins",
			Handler:	_Msg_BurnUserCoins_Handler,
		},
		{
			MethodName:	"UpdateBurnParams",
			Handler:	_Msg_UpdateBurnParams_Handler,
		},
	},
	Streams:	[]grpc.StreamDesc{},
	Metadata:	"l1/burn/v1/tx.proto",
}

func (m *MsgBurnProtocolCoins) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgBurnProtocolCoins) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgBurnProtocolCoins) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Reason) > 0 {
		i -= len(m.Reason)
		copy(dAtA[i:], m.Reason)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Reason)))
		i--
		dAtA[i] = 0x2a
	}
	if m.Epoch != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.Epoch))
		i--
		dAtA[i] = 0x20
	}
	if len(m.Amount) > 0 {
		for iNdEx := len(m.Amount) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Amount[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintTx(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x1a
		}
	}
	if len(m.SourceModule) > 0 {
		i -= len(m.SourceModule)
		copy(dAtA[i:], m.SourceModule)
		i = encodeVarintTx(dAtA, i, uint64(len(m.SourceModule)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Authority) > 0 {
		i -= len(m.Authority)
		copy(dAtA[i:], m.Authority)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Authority)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MsgBurnProtocolCoinsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgBurnProtocolCoinsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgBurnProtocolCoinsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.Burn.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintTx(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *MsgBurnUserCoins) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgBurnUserCoins) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgBurnUserCoins) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Reason) > 0 {
		i -= len(m.Reason)
		copy(dAtA[i:], m.Reason)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Reason)))
		i--
		dAtA[i] = 0x22
	}
	if m.Epoch != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.Epoch))
		i--
		dAtA[i] = 0x18
	}
	if len(m.Amount) > 0 {
		for iNdEx := len(m.Amount) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Amount[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintTx(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
		}
	}
	if len(m.Burner) > 0 {
		i -= len(m.Burner)
		copy(dAtA[i:], m.Burner)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Burner)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MsgBurnUserCoinsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgBurnUserCoinsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgBurnUserCoinsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.Burn.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintTx(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *MsgUpdateBurnParams) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgUpdateBurnParams) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgUpdateBurnParams) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.Params.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintTx(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	if len(m.Authority) > 0 {
		i -= len(m.Authority)
		copy(dAtA[i:], m.Authority)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Authority)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MsgUpdateBurnParamsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgUpdateBurnParamsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgUpdateBurnParamsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func encodeVarintTx(dAtA []byte, offset int, v uint64) int {
	offset -= sovTx(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *MsgBurnProtocolCoins) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Authority)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.SourceModule)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if len(m.Amount) > 0 {
		for _, e := range m.Amount {
			l = e.Size()
			n += 1 + l + sovTx(uint64(l))
		}
	}
	if m.Epoch != 0 {
		n += 1 + sovTx(uint64(m.Epoch))
	}
	l = len(m.Reason)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	return n
}

func (m *MsgBurnProtocolCoinsResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Burn.Size()
	n += 1 + l + sovTx(uint64(l))
	return n
}

func (m *MsgBurnUserCoins) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Burner)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if len(m.Amount) > 0 {
		for _, e := range m.Amount {
			l = e.Size()
			n += 1 + l + sovTx(uint64(l))
		}
	}
	if m.Epoch != 0 {
		n += 1 + sovTx(uint64(m.Epoch))
	}
	l = len(m.Reason)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	return n
}

func (m *MsgBurnUserCoinsResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Burn.Size()
	n += 1 + l + sovTx(uint64(l))
	return n
}

func (m *MsgUpdateBurnParams) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Authority)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = m.Params.Size()
	n += 1 + l + sovTx(uint64(l))
	return n
}

func (m *MsgUpdateBurnParamsResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func sovTx(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozTx(x uint64) (n int) {
	return sovTx(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *MsgBurnProtocolCoins) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
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
			return fmt.Errorf("proto: MsgBurnProtocolCoins: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgBurnProtocolCoins: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Authority", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Authority = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SourceModule", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.SourceModule = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Amount", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Amount = append(m.Amount, types.Coin{})
			if err := m.Amount[len(m.Amount)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Epoch", wireType)
			}
			m.Epoch = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Reason", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Reason = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
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
func (m *MsgBurnProtocolCoinsResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
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
			return fmt.Errorf("proto: MsgBurnProtocolCoinsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgBurnProtocolCoinsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Burn", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Burn.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
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
func (m *MsgBurnUserCoins) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
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
			return fmt.Errorf("proto: MsgBurnUserCoins: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgBurnUserCoins: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Burner", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Burner = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Amount", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Amount = append(m.Amount, types.Coin{})
			if err := m.Amount[len(m.Amount)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Epoch", wireType)
			}
			m.Epoch = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Reason", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Reason = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
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
func (m *MsgBurnUserCoinsResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
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
			return fmt.Errorf("proto: MsgBurnUserCoinsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgBurnUserCoinsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Burn", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Burn.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
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
func (m *MsgUpdateBurnParams) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
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
			return fmt.Errorf("proto: MsgUpdateBurnParams: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgUpdateBurnParams: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Authority", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Authority = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Params", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Params.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
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
func (m *MsgUpdateBurnParamsResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
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
			return fmt.Errorf("proto: MsgUpdateBurnParamsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgUpdateBurnParamsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
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
func skipTx(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowTx
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
					return 0, ErrIntOverflowTx
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
					return 0, ErrIntOverflowTx
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
				return 0, ErrInvalidLengthTx
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupTx
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthTx
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthTx		= fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowTx		= fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupTx	= fmt.Errorf("proto: unexpected end of group")
)
