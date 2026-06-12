package types

import (
	context "context"
	fmt "fmt"
	_ "github.com/cosmos/cosmos-sdk/types/msgservice"
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

type MsgUpdateEmissionsParams struct {
	Authority	string	`protobuf:"bytes,1,opt,name=authority,proto3" json:"authority,omitempty"`
	Params		Params	`protobuf:"bytes,2,opt,name=params,proto3" json:"params"`
}

func (m *MsgUpdateEmissionsParams) Reset()		{ *m = MsgUpdateEmissionsParams{} }
func (m *MsgUpdateEmissionsParams) String() string	{ return proto.CompactTextString(m) }
func (*MsgUpdateEmissionsParams) ProtoMessage()		{}
func (*MsgUpdateEmissionsParams) Descriptor() ([]byte, []int) {
	return fileDescriptor_dcb9d4efabd350c1, []int{0}
}
func (m *MsgUpdateEmissionsParams) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgUpdateEmissionsParams) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgUpdateEmissionsParams.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgUpdateEmissionsParams) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgUpdateEmissionsParams.Merge(m, src)
}
func (m *MsgUpdateEmissionsParams) XXX_Size() int {
	return m.Size()
}
func (m *MsgUpdateEmissionsParams) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgUpdateEmissionsParams.DiscardUnknown(m)
}

var xxx_messageInfo_MsgUpdateEmissionsParams proto.InternalMessageInfo

func (m *MsgUpdateEmissionsParams) GetAuthority() string {
	if m != nil {
		return m.Authority
	}
	return ""
}

func (m *MsgUpdateEmissionsParams) GetParams() Params {
	if m != nil {
		return m.Params
	}
	return Params{}
}

type MsgUpdateEmissionsParamsResponse struct {
}

func (m *MsgUpdateEmissionsParamsResponse) Reset()		{ *m = MsgUpdateEmissionsParamsResponse{} }
func (m *MsgUpdateEmissionsParamsResponse) String() string	{ return proto.CompactTextString(m) }
func (*MsgUpdateEmissionsParamsResponse) ProtoMessage()		{}
func (*MsgUpdateEmissionsParamsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_dcb9d4efabd350c1, []int{1}
}
func (m *MsgUpdateEmissionsParamsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgUpdateEmissionsParamsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgUpdateEmissionsParamsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgUpdateEmissionsParamsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgUpdateEmissionsParamsResponse.Merge(m, src)
}
func (m *MsgUpdateEmissionsParamsResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgUpdateEmissionsParamsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgUpdateEmissionsParamsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgUpdateEmissionsParamsResponse proto.InternalMessageInfo

type MsgFinalizeEmissionEpoch struct {
	Authority	string	`protobuf:"bytes,1,opt,name=authority,proto3" json:"authority,omitempty"`
	Epoch		uint64	`protobuf:"varint,2,opt,name=epoch,proto3" json:"epoch,omitempty"`
	StakingRatioBps	uint32	`protobuf:"varint,3,opt,name=staking_ratio_bps,json=stakingRatioBps,proto3" json:"staking_ratio_bps,omitempty"`
}

func (m *MsgFinalizeEmissionEpoch) Reset()		{ *m = MsgFinalizeEmissionEpoch{} }
func (m *MsgFinalizeEmissionEpoch) String() string	{ return proto.CompactTextString(m) }
func (*MsgFinalizeEmissionEpoch) ProtoMessage()		{}
func (*MsgFinalizeEmissionEpoch) Descriptor() ([]byte, []int) {
	return fileDescriptor_dcb9d4efabd350c1, []int{2}
}
func (m *MsgFinalizeEmissionEpoch) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgFinalizeEmissionEpoch) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgFinalizeEmissionEpoch.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgFinalizeEmissionEpoch) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgFinalizeEmissionEpoch.Merge(m, src)
}
func (m *MsgFinalizeEmissionEpoch) XXX_Size() int {
	return m.Size()
}
func (m *MsgFinalizeEmissionEpoch) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgFinalizeEmissionEpoch.DiscardUnknown(m)
}

var xxx_messageInfo_MsgFinalizeEmissionEpoch proto.InternalMessageInfo

func (m *MsgFinalizeEmissionEpoch) GetAuthority() string {
	if m != nil {
		return m.Authority
	}
	return ""
}

func (m *MsgFinalizeEmissionEpoch) GetEpoch() uint64 {
	if m != nil {
		return m.Epoch
	}
	return 0
}

func (m *MsgFinalizeEmissionEpoch) GetStakingRatioBps() uint32 {
	if m != nil {
		return m.StakingRatioBps
	}
	return 0
}

type MsgFinalizeEmissionEpochResponse struct {
	EmissionEpoch EmissionEpoch `protobuf:"bytes,1,opt,name=emission_epoch,json=emissionEpoch,proto3" json:"emission_epoch"`
}

func (m *MsgFinalizeEmissionEpochResponse) Reset()		{ *m = MsgFinalizeEmissionEpochResponse{} }
func (m *MsgFinalizeEmissionEpochResponse) String() string	{ return proto.CompactTextString(m) }
func (*MsgFinalizeEmissionEpochResponse) ProtoMessage()		{}
func (*MsgFinalizeEmissionEpochResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_dcb9d4efabd350c1, []int{3}
}
func (m *MsgFinalizeEmissionEpochResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgFinalizeEmissionEpochResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgFinalizeEmissionEpochResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgFinalizeEmissionEpochResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgFinalizeEmissionEpochResponse.Merge(m, src)
}
func (m *MsgFinalizeEmissionEpochResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgFinalizeEmissionEpochResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgFinalizeEmissionEpochResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgFinalizeEmissionEpochResponse proto.InternalMessageInfo

func (m *MsgFinalizeEmissionEpochResponse) GetEmissionEpoch() EmissionEpoch {
	if m != nil {
		return m.EmissionEpoch
	}
	return EmissionEpoch{}
}

func init() {
	proto.RegisterType((*MsgUpdateEmissionsParams)(nil), "l1.emissions.v1.MsgUpdateEmissionsParams")
	proto.RegisterType((*MsgUpdateEmissionsParamsResponse)(nil), "l1.emissions.v1.MsgUpdateEmissionsParamsResponse")
	proto.RegisterType((*MsgFinalizeEmissionEpoch)(nil), "l1.emissions.v1.MsgFinalizeEmissionEpoch")
	proto.RegisterType((*MsgFinalizeEmissionEpochResponse)(nil), "l1.emissions.v1.MsgFinalizeEmissionEpochResponse")
}

func init()	{ proto.RegisterFile("l1/emissions/v1/tx.proto", fileDescriptor_dcb9d4efabd350c1) }

var fileDescriptor_dcb9d4efabd350c1 = []byte{

	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x92, 0xcd, 0xaa, 0xd3, 0x40,
	0x14, 0xc7, 0x33, 0xde, 0x0f, 0xb8, 0x23, 0xf7, 0x5e, 0x0c, 0xb7, 0x34, 0x04, 0x8d, 0x21, 0xab,
	0x5a, 0x34, 0x43, 0x2a, 0x6e, 0x5c, 0x16, 0xda, 0x8d, 0x14, 0x24, 0xe0, 0xc6, 0x4d, 0x49, 0xeb,
	0x30, 0x1d, 0x4c, 0x32, 0x43, 0xce, 0xa4, 0xb4, 0x6e, 0x14, 0x97, 0xae, 0x7c, 0x94, 0x3e, 0x46,
	0x97, 0x5d, 0xba, 0x12, 0x69, 0x17, 0x7d, 0x07, 0x57, 0x92, 0x8f, 0x7e, 0xd8, 0xa6, 0x7a, 0x77,
	0x93, 0x73, 0x7e, 0x39, 0xe7, 0x97, 0x7f, 0x06, 0x1b, 0xa1, 0x47, 0x68, 0xc4, 0x01, 0xb8, 0x88,
	0x81, 0x8c, 0x3d, 0xa2, 0x26, 0xae, 0x4c, 0x84, 0x12, 0xfa, 0x6d, 0xe8, 0xb9, 0xdb, 0x8e, 0x3b,
	0xf6, 0xcc, 0xfa, 0x50, 0x40, 0x24, 0x80, 0x44, 0xc0, 0x32, 0x30, 0x02, 0x56, 0x90, 0xe6, 0x1d,
	0x13, 0x4c, 0xe4, 0x47, 0x92, 0x9d, 0xca, 0xea, 0x93, 0xc3, 0xc9, 0x8c, 0xc6, 0x14, 0x38, 0x14,
	0x6d, 0xe7, 0x33, 0x36, 0x7a, 0xc0, 0xde, 0xc9, 0x0f, 0x81, 0xa2, 0x9d, 0x0d, 0xf6, 0x36, 0x48,
	0x82, 0x08, 0xf4, 0xc7, 0xf8, 0x2a, 0x48, 0xd5, 0x48, 0x24, 0x5c, 0x4d, 0x0d, 0x64, 0xa3, 0xc6,
	0x95, 0xbf, 0x2b, 0xe8, 0xaf, 0xf0, 0xa5, 0xcc, 0x39, 0xe3, 0x81, 0x8d, 0x1a, 0x0f, 0x5b, 0x75,
	0xf7, 0xc0, 0xd4, 0x2d, 0xc6, 0xb4, 0xcf, 0xe7, 0x3f, 0x9f, 0x6a, 0x7e, 0x09, 0xbf, 0xbe, 0xf9,
	0xba, 0x9e, 0x35, 0x77, 0x63, 0x1c, 0x07, 0xdb, 0xa7, 0x04, 0x7c, 0x0a, 0x52, 0xc4, 0x40, 0x9d,
	0x6f, 0x28, 0xb7, 0xec, 0xf2, 0x38, 0x08, 0xf9, 0xa7, 0x2d, 0xd6, 0x91, 0x62, 0x38, 0xfa, 0x8f,
	0xe5, 0x1d, 0xbe, 0xa0, 0x19, 0x96, 0x4b, 0x9e, 0xfb, 0xc5, 0x83, 0xde, 0xc4, 0x8f, 0x40, 0x05,
	0x1f, 0x79, 0xcc, 0xfa, 0x49, 0xa0, 0xb8, 0xe8, 0x0f, 0x24, 0x18, 0x67, 0x36, 0x6a, 0x5c, 0xfb,
	0xb7, 0x65, 0xc3, 0xcf, 0xea, 0x6d, 0x79, 0x2c, 0x2c, 0x72, 0xe1, 0x4a, 0x97, 0x8d, 0xb0, 0xfe,
	0x06, 0xdf, 0x6c, 0x92, 0xe8, 0x17, 0xeb, 0x51, 0x9e, 0x91, 0x75, 0x94, 0xd1, 0x5f, 0xef, 0x97,
	0x51, 0x5d, 0xd3, 0xfd, 0x62, 0xeb, 0x37, 0xc2, 0x67, 0x3d, 0x60, 0x7a, 0x8a, 0x6b, 0xd5, 0xff,
	0xe9, 0xd9, 0xd1, 0xd4, 0x53, 0x89, 0x9a, 0xde, 0xbd, 0xd1, 0xed, 0xb7, 0xa4, 0xb8, 0x56, 0x1d,
	0x7c, 0xe5, 0xda, 0x4a, 0xb4, 0x7a, 0xed, 0x3f, 0x23, 0x34, 0x2f, 0xbe, 0xac, 0x67, 0x4d, 0xd4,
	0xee, 0xce, 0x97, 0x16, 0x5a, 0x2c, 0x2d, 0xf4, 0x6b, 0x69, 0xa1, 0xef, 0x2b, 0x4b, 0x5b, 0xac,
	0x2c, 0xed, 0xc7, 0xca, 0xd2, 0xde, 0x3f, 0x67, 0x5c, 0x8d, 0xd2, 0x81, 0x3b, 0x14, 0x11, 0x01,
	0x31, 0xa6, 0x09, 0xe5, 0x2c, 0x7e, 0x11, 0x7a, 0x24, 0xf4, 0xc8, 0x64, 0xef, 0xca, 0xab, 0xa9,
	0xa4, 0x30, 0xb8, 0xcc, 0xaf, 0xfb, 0xcb, 0x3f, 0x01, 0x00, 0x00, 0xff, 0xff, 0x3e, 0x96, 0x5f,
	0xf4, 0x69, 0x03, 0x00, 0x00,
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
	UpdateEmissionsParams(ctx context.Context, in *MsgUpdateEmissionsParams, opts ...grpc.CallOption) (*MsgUpdateEmissionsParamsResponse, error)
	FinalizeEmissionEpoch(ctx context.Context, in *MsgFinalizeEmissionEpoch, opts ...grpc.CallOption) (*MsgFinalizeEmissionEpochResponse, error)
}

type msgClient struct {
	cc grpc1.ClientConn
}

func NewMsgClient(cc grpc1.ClientConn) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) UpdateEmissionsParams(ctx context.Context, in *MsgUpdateEmissionsParams, opts ...grpc.CallOption) (*MsgUpdateEmissionsParamsResponse, error) {
	out := new(MsgUpdateEmissionsParamsResponse)
	err := c.cc.Invoke(ctx, "/l1.emissions.v1.Msg/UpdateEmissionsParams", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) FinalizeEmissionEpoch(ctx context.Context, in *MsgFinalizeEmissionEpoch, opts ...grpc.CallOption) (*MsgFinalizeEmissionEpochResponse, error) {
	out := new(MsgFinalizeEmissionEpochResponse)
	err := c.cc.Invoke(ctx, "/l1.emissions.v1.Msg/FinalizeEmissionEpoch", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MsgServer is the server API for Msg service.
type MsgServer interface {
	UpdateEmissionsParams(context.Context, *MsgUpdateEmissionsParams) (*MsgUpdateEmissionsParamsResponse, error)
	FinalizeEmissionEpoch(context.Context, *MsgFinalizeEmissionEpoch) (*MsgFinalizeEmissionEpochResponse, error)
}

// UnimplementedMsgServer can be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (*UnimplementedMsgServer) UpdateEmissionsParams(ctx context.Context, req *MsgUpdateEmissionsParams) (*MsgUpdateEmissionsParamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateEmissionsParams not implemented")
}
func (*UnimplementedMsgServer) FinalizeEmissionEpoch(ctx context.Context, req *MsgFinalizeEmissionEpoch) (*MsgFinalizeEmissionEpochResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FinalizeEmissionEpoch not implemented")
}

func RegisterMsgServer(s grpc1.Server, srv MsgServer) {
	s.RegisterService(&_Msg_serviceDesc, srv)
}

func _Msg_UpdateEmissionsParams_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUpdateEmissionsParams)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).UpdateEmissionsParams(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.emissions.v1.Msg/UpdateEmissionsParams",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).UpdateEmissionsParams(ctx, req.(*MsgUpdateEmissionsParams))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_FinalizeEmissionEpoch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgFinalizeEmissionEpoch)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).FinalizeEmissionEpoch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.emissions.v1.Msg/FinalizeEmissionEpoch",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).FinalizeEmissionEpoch(ctx, req.(*MsgFinalizeEmissionEpoch))
	}
	return interceptor(ctx, in, info, handler)
}

var Msg_serviceDesc = _Msg_serviceDesc
var _Msg_serviceDesc = grpc.ServiceDesc{
	ServiceName:	"l1.emissions.v1.Msg",
	HandlerType:	(*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName:	"UpdateEmissionsParams",
			Handler:	_Msg_UpdateEmissionsParams_Handler,
		},
		{
			MethodName:	"FinalizeEmissionEpoch",
			Handler:	_Msg_FinalizeEmissionEpoch_Handler,
		},
	},
	Streams:	[]grpc.StreamDesc{},
	Metadata:	"l1/emissions/v1/tx.proto",
}

func (m *MsgUpdateEmissionsParams) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgUpdateEmissionsParams) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgUpdateEmissionsParams) MarshalToSizedBuffer(dAtA []byte) (int, error) {
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

func (m *MsgUpdateEmissionsParamsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgUpdateEmissionsParamsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgUpdateEmissionsParamsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *MsgFinalizeEmissionEpoch) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgFinalizeEmissionEpoch) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgFinalizeEmissionEpoch) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.StakingRatioBps != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.StakingRatioBps))
		i--
		dAtA[i] = 0x18
	}
	if m.Epoch != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.Epoch))
		i--
		dAtA[i] = 0x10
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

func (m *MsgFinalizeEmissionEpochResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgFinalizeEmissionEpochResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgFinalizeEmissionEpochResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.EmissionEpoch.MarshalToSizedBuffer(dAtA[:i])
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
func (m *MsgUpdateEmissionsParams) Size() (n int) {
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

func (m *MsgUpdateEmissionsParamsResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *MsgFinalizeEmissionEpoch) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Authority)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if m.Epoch != 0 {
		n += 1 + sovTx(uint64(m.Epoch))
	}
	if m.StakingRatioBps != 0 {
		n += 1 + sovTx(uint64(m.StakingRatioBps))
	}
	return n
}

func (m *MsgFinalizeEmissionEpochResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.EmissionEpoch.Size()
	n += 1 + l + sovTx(uint64(l))
	return n
}

func sovTx(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozTx(x uint64) (n int) {
	return sovTx(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *MsgUpdateEmissionsParams) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgUpdateEmissionsParams: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgUpdateEmissionsParams: illegal tag %d (wire type %d)", fieldNum, wire)
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
func (m *MsgUpdateEmissionsParamsResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgUpdateEmissionsParamsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgUpdateEmissionsParamsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
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
func (m *MsgFinalizeEmissionEpoch) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgFinalizeEmissionEpoch: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgFinalizeEmissionEpoch: illegal tag %d (wire type %d)", fieldNum, wire)
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
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field StakingRatioBps", wireType)
			}
			m.StakingRatioBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.StakingRatioBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
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
func (m *MsgFinalizeEmissionEpochResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgFinalizeEmissionEpochResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgFinalizeEmissionEpochResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field EmissionEpoch", wireType)
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
			if err := m.EmissionEpoch.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
