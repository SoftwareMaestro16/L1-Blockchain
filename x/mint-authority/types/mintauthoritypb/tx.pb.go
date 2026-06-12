package mintauthoritypb

import (
	context "context"
	fmt "fmt"
	_ "github.com/cosmos/cosmos-sdk/types/msgservice"
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

type MsgMintProtocolCoins struct {
	Caller				string	`protobuf:"bytes,1,opt,name=caller,proto3" json:"caller,omitempty"`
	Recipient			string	`protobuf:"bytes,2,opt,name=recipient,proto3" json:"recipient,omitempty"`
	Denom				string	`protobuf:"bytes,3,opt,name=denom,proto3" json:"denom,omitempty"`
	Amount				string	`protobuf:"bytes,4,opt,name=amount,proto3" json:"amount,omitempty"`
	Epoch				uint64	`protobuf:"varint,5,opt,name=epoch,proto3" json:"epoch,omitempty"`
	Height				uint64	`protobuf:"varint,6,opt,name=height,proto3" json:"height,omitempty"`
	EmissionsDecisionJson		string	`protobuf:"bytes,7,opt,name=emissions_decision_json,json=emissionsDecisionJson,proto3" json:"emissions_decision_json,omitempty"`
	Emergency			bool	`protobuf:"varint,8,opt,name=emergency,proto3" json:"emergency,omitempty"`
	EmergencyAuthorizationJson	string	`protobuf:"bytes,9,opt,name=emergency_authorization_json,json=emergencyAuthorizationJson,proto3" json:"emergency_authorization_json,omitempty"`
}

func (m *MsgMintProtocolCoins) Reset()		{ *m = MsgMintProtocolCoins{} }
func (m *MsgMintProtocolCoins) String() string	{ return proto.CompactTextString(m) }
func (*MsgMintProtocolCoins) ProtoMessage()	{}
func (*MsgMintProtocolCoins) Descriptor() ([]byte, []int) {
	return fileDescriptor_a988a1188d8bd2f0, []int{0}
}
func (m *MsgMintProtocolCoins) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgMintProtocolCoins) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgMintProtocolCoins.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgMintProtocolCoins) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgMintProtocolCoins.Merge(m, src)
}
func (m *MsgMintProtocolCoins) XXX_Size() int {
	return m.Size()
}
func (m *MsgMintProtocolCoins) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgMintProtocolCoins.DiscardUnknown(m)
}

var xxx_messageInfo_MsgMintProtocolCoins proto.InternalMessageInfo

func (m *MsgMintProtocolCoins) GetCaller() string {
	if m != nil {
		return m.Caller
	}
	return ""
}

func (m *MsgMintProtocolCoins) GetRecipient() string {
	if m != nil {
		return m.Recipient
	}
	return ""
}

func (m *MsgMintProtocolCoins) GetDenom() string {
	if m != nil {
		return m.Denom
	}
	return ""
}

func (m *MsgMintProtocolCoins) GetAmount() string {
	if m != nil {
		return m.Amount
	}
	return ""
}

func (m *MsgMintProtocolCoins) GetEpoch() uint64 {
	if m != nil {
		return m.Epoch
	}
	return 0
}

func (m *MsgMintProtocolCoins) GetHeight() uint64 {
	if m != nil {
		return m.Height
	}
	return 0
}

func (m *MsgMintProtocolCoins) GetEmissionsDecisionJson() string {
	if m != nil {
		return m.EmissionsDecisionJson
	}
	return ""
}

func (m *MsgMintProtocolCoins) GetEmergency() bool {
	if m != nil {
		return m.Emergency
	}
	return false
}

func (m *MsgMintProtocolCoins) GetEmergencyAuthorizationJson() string {
	if m != nil {
		return m.EmergencyAuthorizationJson
	}
	return ""
}

type MsgMintProtocolCoinsResponse struct {
	EventJson string `protobuf:"bytes,1,opt,name=event_json,json=eventJson,proto3" json:"event_json,omitempty"`
}

func (m *MsgMintProtocolCoinsResponse) Reset()		{ *m = MsgMintProtocolCoinsResponse{} }
func (m *MsgMintProtocolCoinsResponse) String() string	{ return proto.CompactTextString(m) }
func (*MsgMintProtocolCoinsResponse) ProtoMessage()	{}
func (*MsgMintProtocolCoinsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_a988a1188d8bd2f0, []int{1}
}
func (m *MsgMintProtocolCoinsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgMintProtocolCoinsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgMintProtocolCoinsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgMintProtocolCoinsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgMintProtocolCoinsResponse.Merge(m, src)
}
func (m *MsgMintProtocolCoinsResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgMintProtocolCoinsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgMintProtocolCoinsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgMintProtocolCoinsResponse proto.InternalMessageInfo

func (m *MsgMintProtocolCoinsResponse) GetEventJson() string {
	if m != nil {
		return m.EventJson
	}
	return ""
}

type MsgUpdateMintAuthorityParams struct {
	Authority	string	`protobuf:"bytes,1,opt,name=authority,proto3" json:"authority,omitempty"`
	StateJson	string	`protobuf:"bytes,2,opt,name=state_json,json=stateJson,proto3" json:"state_json,omitempty"`
}

func (m *MsgUpdateMintAuthorityParams) Reset()		{ *m = MsgUpdateMintAuthorityParams{} }
func (m *MsgUpdateMintAuthorityParams) String() string	{ return proto.CompactTextString(m) }
func (*MsgUpdateMintAuthorityParams) ProtoMessage()	{}
func (*MsgUpdateMintAuthorityParams) Descriptor() ([]byte, []int) {
	return fileDescriptor_a988a1188d8bd2f0, []int{2}
}
func (m *MsgUpdateMintAuthorityParams) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgUpdateMintAuthorityParams) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgUpdateMintAuthorityParams.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgUpdateMintAuthorityParams) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgUpdateMintAuthorityParams.Merge(m, src)
}
func (m *MsgUpdateMintAuthorityParams) XXX_Size() int {
	return m.Size()
}
func (m *MsgUpdateMintAuthorityParams) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgUpdateMintAuthorityParams.DiscardUnknown(m)
}

var xxx_messageInfo_MsgUpdateMintAuthorityParams proto.InternalMessageInfo

func (m *MsgUpdateMintAuthorityParams) GetAuthority() string {
	if m != nil {
		return m.Authority
	}
	return ""
}

func (m *MsgUpdateMintAuthorityParams) GetStateJson() string {
	if m != nil {
		return m.StateJson
	}
	return ""
}

type MsgUpdateMintAuthorityParamsResponse struct {
}

func (m *MsgUpdateMintAuthorityParamsResponse) Reset()		{ *m = MsgUpdateMintAuthorityParamsResponse{} }
func (m *MsgUpdateMintAuthorityParamsResponse) String() string	{ return proto.CompactTextString(m) }
func (*MsgUpdateMintAuthorityParamsResponse) ProtoMessage()	{}
func (*MsgUpdateMintAuthorityParamsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_a988a1188d8bd2f0, []int{3}
}
func (m *MsgUpdateMintAuthorityParamsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgUpdateMintAuthorityParamsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgUpdateMintAuthorityParamsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgUpdateMintAuthorityParamsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgUpdateMintAuthorityParamsResponse.Merge(m, src)
}
func (m *MsgUpdateMintAuthorityParamsResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgUpdateMintAuthorityParamsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgUpdateMintAuthorityParamsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgUpdateMintAuthorityParamsResponse proto.InternalMessageInfo

func init() {
	proto.RegisterType((*MsgMintProtocolCoins)(nil), "l1.mintauthority.v1.MsgMintProtocolCoins")
	proto.RegisterType((*MsgMintProtocolCoinsResponse)(nil), "l1.mintauthority.v1.MsgMintProtocolCoinsResponse")
	proto.RegisterType((*MsgUpdateMintAuthorityParams)(nil), "l1.mintauthority.v1.MsgUpdateMintAuthorityParams")
	proto.RegisterType((*MsgUpdateMintAuthorityParamsResponse)(nil), "l1.mintauthority.v1.MsgUpdateMintAuthorityParamsResponse")
}

func init()	{ proto.RegisterFile("l1/mintauthority/v1/tx.proto", fileDescriptor_a988a1188d8bd2f0) }

var fileDescriptor_a988a1188d8bd2f0 = []byte{

	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x53, 0x41, 0x8b, 0xd3, 0x40,
	0x14, 0xee, 0x74, 0xb7, 0x75, 0x3b, 0x82, 0x60, 0x5c, 0xdd, 0x18, 0x6a, 0x28, 0x45, 0xa4, 0x2e,
	0x6c, 0x42, 0x14, 0x04, 0x05, 0xc1, 0xaa, 0x27, 0xa1, 0xb0, 0x14, 0xbc, 0xe8, 0xa1, 0xa4, 0xe9,
	0x23, 0x19, 0xcd, 0xcc, 0xc4, 0xbc, 0x69, 0xd9, 0x7a, 0x12, 0x6f, 0xe2, 0xc5, 0x9f, 0xb2, 0x3f,
	0xc3, 0x8b, 0xb0, 0x47, 0x8f, 0xd2, 0x1e, 0xf6, 0x6f, 0xc8, 0x4c, 0xd2, 0xd4, 0xc5, 0x56, 0xdc,
	0x5b, 0xbf, 0xef, 0x7d, 0xef, 0xeb, 0xcb, 0xf7, 0xe6, 0xd1, 0x76, 0x1a, 0xf8, 0x9c, 0x09, 0x15,
	0x4e, 0x55, 0x22, 0x73, 0xa6, 0xe6, 0xfe, 0x2c, 0xf0, 0xd5, 0x89, 0x97, 0xe5, 0x52, 0x49, 0xeb,
	0x46, 0x1a, 0x78, 0x17, 0xaa, 0xde, 0x2c, 0x70, 0x0e, 0x22, 0x89, 0x5c, 0xa2, 0xcf, 0x31, 0xd6,
	0x62, 0x8e, 0x71, 0xa1, 0xee, 0xfe, 0xa8, 0xd3, 0xfd, 0x01, 0xc6, 0x03, 0x26, 0xd4, 0xb1, 0x26,
	0x22, 0x99, 0xbe, 0x90, 0x4c, 0xa0, 0x75, 0x8b, 0x36, 0xa3, 0x30, 0x4d, 0x21, 0xb7, 0x49, 0x87,
	0xf4, 0x5a, 0xc3, 0x12, 0x59, 0x6d, 0xda, 0xca, 0x21, 0x62, 0x19, 0x03, 0xa1, 0xec, 0xba, 0x29,
	0xad, 0x09, 0x6b, 0x9f, 0x36, 0x26, 0x20, 0x24, 0xb7, 0x77, 0x4c, 0xa5, 0x00, 0xda, 0x2b, 0xe4,
	0x72, 0x2a, 0x94, 0xbd, 0x5b, 0x78, 0x15, 0x48, 0xab, 0x21, 0x93, 0x51, 0x62, 0x37, 0x3a, 0xa4,
	0xb7, 0x3b, 0x2c, 0x80, 0x56, 0x27, 0xc0, 0xe2, 0x44, 0xd9, 0x4d, 0x43, 0x97, 0xc8, 0x7a, 0x44,
	0x0f, 0x80, 0x33, 0x44, 0x26, 0x05, 0x8e, 0x26, 0x10, 0x31, 0xfd, 0x6b, 0xf4, 0x0e, 0xa5, 0xb0,
	0xaf, 0x18, 0xdb, 0x9b, 0x55, 0xf9, 0x65, 0x59, 0x7d, 0x85, 0x52, 0xe8, 0x89, 0x81, 0x43, 0x1e,
	0x83, 0x88, 0xe6, 0xf6, 0x5e, 0x87, 0xf4, 0xf6, 0x86, 0x6b, 0xc2, 0x7a, 0x46, 0xdb, 0x15, 0x18,
	0x95, 0x99, 0x7d, 0x0c, 0x55, 0x65, 0xdd, 0x32, 0xd6, 0x4e, 0xa5, 0xe9, 0xff, 0x29, 0xd1, 0xfe,
	0x4f, 0xae, 0x7e, 0x3e, 0x3f, 0x3d, 0x2c, 0xe3, 0xe9, 0x3e, 0xa5, 0xed, 0x4d, 0x71, 0x0e, 0x01,
	0x33, 0x29, 0x10, 0xac, 0x3b, 0x94, 0xc2, 0x0c, 0x84, 0x2a, 0xcc, 0x8b, 0x68, 0x5b, 0x86, 0xd1,
	0x5e, 0xdd, 0xf7, 0xa6, 0xfd, 0x75, 0x36, 0x09, 0x15, 0x68, 0x93, 0xfe, 0x6a, 0x89, 0xc7, 0x61,
	0x1e, 0x72, 0xd4, 0xdf, 0x52, 0xed, 0x75, 0xd5, 0x5d, 0x11, 0xda, 0x1c, 0x55, 0xa8, 0xa0, 0x30,
	0x2f, 0x97, 0x63, 0x18, 0x33, 0xe8, 0x35, 0x3d, 0xe8, 0x5a, 0xde, 0xbd, 0x47, 0xef, 0xfe, 0xeb,
	0xcf, 0x56, 0x33, 0x3f, 0xf8, 0x5a, 0xa7, 0x3b, 0x03, 0x8c, 0xad, 0x0f, 0xf4, 0xfa, 0xdf, 0xef,
	0xe4, 0xbe, 0xb7, 0xe1, 0xbd, 0x79, 0x9b, 0x32, 0x70, 0x82, 0xff, 0x96, 0x56, 0x71, 0x7d, 0x21,
	0xf4, 0xf6, 0xf6, 0x34, 0xb6, 0x1a, 0x6e, 0x6d, 0x71, 0x1e, 0x5f, 0xba, 0x65, 0x35, 0x8b, 0xd3,
	0xf8, 0x74, 0x7e, 0x7a, 0x48, 0x9e, 0xbf, 0xfd, 0xbe, 0x70, 0xc9, 0xd9, 0xc2, 0x25, 0xbf, 0x16,
	0x2e, 0xf9, 0xb6, 0x74, 0x6b, 0x67, 0x4b, 0xb7, 0xf6, 0x73, 0xe9, 0xd6, 0xde, 0xf4, 0x63, 0xa6,
	0x92, 0xe9, 0xd8, 0x8b, 0x24, 0xf7, 0x51, 0xce, 0x20, 0x07, 0x16, 0x8b, 0xa3, 0x34, 0xf0, 0xd3,
	0xc0, 0x3f, 0x31, 0x17, 0x7b, 0xb4, 0x3e, 0x59, 0x35, 0xcf, 0x00, 0x2f, 0x9e, 0x71, 0x36, 0x1e,
	0x37, 0xcd, 0x55, 0x3e, 0xfc, 0x1d, 0x00, 0x00, 0xff, 0xff, 0xbf, 0x41, 0xd9, 0x09, 0xe3, 0x03,
	0x00, 0x00,
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
	MintProtocolCoins(ctx context.Context, in *MsgMintProtocolCoins, opts ...grpc.CallOption) (*MsgMintProtocolCoinsResponse, error)
	UpdateMintAuthorityParams(ctx context.Context, in *MsgUpdateMintAuthorityParams, opts ...grpc.CallOption) (*MsgUpdateMintAuthorityParamsResponse, error)
}

type msgClient struct {
	cc grpc1.ClientConn
}

func NewMsgClient(cc grpc1.ClientConn) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) MintProtocolCoins(ctx context.Context, in *MsgMintProtocolCoins, opts ...grpc.CallOption) (*MsgMintProtocolCoinsResponse, error) {
	out := new(MsgMintProtocolCoinsResponse)
	err := c.cc.Invoke(ctx, "/l1.mintauthority.v1.Msg/MintProtocolCoins", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) UpdateMintAuthorityParams(ctx context.Context, in *MsgUpdateMintAuthorityParams, opts ...grpc.CallOption) (*MsgUpdateMintAuthorityParamsResponse, error) {
	out := new(MsgUpdateMintAuthorityParamsResponse)
	err := c.cc.Invoke(ctx, "/l1.mintauthority.v1.Msg/UpdateMintAuthorityParams", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MsgServer is the server API for Msg service.
type MsgServer interface {
	MintProtocolCoins(context.Context, *MsgMintProtocolCoins) (*MsgMintProtocolCoinsResponse, error)
	UpdateMintAuthorityParams(context.Context, *MsgUpdateMintAuthorityParams) (*MsgUpdateMintAuthorityParamsResponse, error)
}

// UnimplementedMsgServer can be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (*UnimplementedMsgServer) MintProtocolCoins(ctx context.Context, req *MsgMintProtocolCoins) (*MsgMintProtocolCoinsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MintProtocolCoins not implemented")
}
func (*UnimplementedMsgServer) UpdateMintAuthorityParams(ctx context.Context, req *MsgUpdateMintAuthorityParams) (*MsgUpdateMintAuthorityParamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateMintAuthorityParams not implemented")
}

func RegisterMsgServer(s grpc1.Server, srv MsgServer) {
	s.RegisterService(&_Msg_serviceDesc, srv)
}

func _Msg_MintProtocolCoins_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgMintProtocolCoins)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).MintProtocolCoins(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.mintauthority.v1.Msg/MintProtocolCoins",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).MintProtocolCoins(ctx, req.(*MsgMintProtocolCoins))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_UpdateMintAuthorityParams_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUpdateMintAuthorityParams)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).UpdateMintAuthorityParams(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.mintauthority.v1.Msg/UpdateMintAuthorityParams",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).UpdateMintAuthorityParams(ctx, req.(*MsgUpdateMintAuthorityParams))
	}
	return interceptor(ctx, in, info, handler)
}

var Msg_serviceDesc = _Msg_serviceDesc
var _Msg_serviceDesc = grpc.ServiceDesc{
	ServiceName:	"l1.mintauthority.v1.Msg",
	HandlerType:	(*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName:	"MintProtocolCoins",
			Handler:	_Msg_MintProtocolCoins_Handler,
		},
		{
			MethodName:	"UpdateMintAuthorityParams",
			Handler:	_Msg_UpdateMintAuthorityParams_Handler,
		},
	},
	Streams:	[]grpc.StreamDesc{},
	Metadata:	"l1/mintauthority/v1/tx.proto",
}

func (m *MsgMintProtocolCoins) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgMintProtocolCoins) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgMintProtocolCoins) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.EmergencyAuthorizationJson) > 0 {
		i -= len(m.EmergencyAuthorizationJson)
		copy(dAtA[i:], m.EmergencyAuthorizationJson)
		i = encodeVarintTx(dAtA, i, uint64(len(m.EmergencyAuthorizationJson)))
		i--
		dAtA[i] = 0x4a
	}
	if m.Emergency {
		i--
		if m.Emergency {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x40
	}
	if len(m.EmissionsDecisionJson) > 0 {
		i -= len(m.EmissionsDecisionJson)
		copy(dAtA[i:], m.EmissionsDecisionJson)
		i = encodeVarintTx(dAtA, i, uint64(len(m.EmissionsDecisionJson)))
		i--
		dAtA[i] = 0x3a
	}
	if m.Height != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.Height))
		i--
		dAtA[i] = 0x30
	}
	if m.Epoch != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.Epoch))
		i--
		dAtA[i] = 0x28
	}
	if len(m.Amount) > 0 {
		i -= len(m.Amount)
		copy(dAtA[i:], m.Amount)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Amount)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.Denom) > 0 {
		i -= len(m.Denom)
		copy(dAtA[i:], m.Denom)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Denom)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Recipient) > 0 {
		i -= len(m.Recipient)
		copy(dAtA[i:], m.Recipient)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Recipient)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Caller) > 0 {
		i -= len(m.Caller)
		copy(dAtA[i:], m.Caller)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Caller)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MsgMintProtocolCoinsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgMintProtocolCoinsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgMintProtocolCoinsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.EventJson) > 0 {
		i -= len(m.EventJson)
		copy(dAtA[i:], m.EventJson)
		i = encodeVarintTx(dAtA, i, uint64(len(m.EventJson)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MsgUpdateMintAuthorityParams) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgUpdateMintAuthorityParams) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgUpdateMintAuthorityParams) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.StateJson) > 0 {
		i -= len(m.StateJson)
		copy(dAtA[i:], m.StateJson)
		i = encodeVarintTx(dAtA, i, uint64(len(m.StateJson)))
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

func (m *MsgUpdateMintAuthorityParamsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgUpdateMintAuthorityParamsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgUpdateMintAuthorityParamsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
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
func (m *MsgMintProtocolCoins) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Caller)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.Recipient)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.Denom)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.Amount)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if m.Epoch != 0 {
		n += 1 + sovTx(uint64(m.Epoch))
	}
	if m.Height != 0 {
		n += 1 + sovTx(uint64(m.Height))
	}
	l = len(m.EmissionsDecisionJson)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if m.Emergency {
		n += 2
	}
	l = len(m.EmergencyAuthorizationJson)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	return n
}

func (m *MsgMintProtocolCoinsResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.EventJson)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	return n
}

func (m *MsgUpdateMintAuthorityParams) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Authority)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.StateJson)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	return n
}

func (m *MsgUpdateMintAuthorityParamsResponse) Size() (n int) {
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
func (m *MsgMintProtocolCoins) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgMintProtocolCoins: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgMintProtocolCoins: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Caller", wireType)
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
			m.Caller = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Recipient", wireType)
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
			m.Recipient = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Denom", wireType)
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
			m.Denom = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Amount", wireType)
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
			m.Amount = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
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
		case 6:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Height", wireType)
			}
			m.Height = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field EmissionsDecisionJson", wireType)
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
			m.EmissionsDecisionJson = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 8:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Emergency", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
			m.Emergency = bool(v != 0)
		case 9:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field EmergencyAuthorizationJson", wireType)
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
			m.EmergencyAuthorizationJson = string(dAtA[iNdEx:postIndex])
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
func (m *MsgMintProtocolCoinsResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgMintProtocolCoinsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgMintProtocolCoinsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field EventJson", wireType)
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
			m.EventJson = string(dAtA[iNdEx:postIndex])
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
func (m *MsgUpdateMintAuthorityParams) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgUpdateMintAuthorityParams: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgUpdateMintAuthorityParams: illegal tag %d (wire type %d)", fieldNum, wire)
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
				return fmt.Errorf("proto: wrong wireType = %d for field StateJson", wireType)
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
			m.StateJson = string(dAtA[iNdEx:postIndex])
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
func (m *MsgUpdateMintAuthorityParamsResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgUpdateMintAuthorityParamsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgUpdateMintAuthorityParamsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
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
