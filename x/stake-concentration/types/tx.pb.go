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

type MsgUpdateConcentrationParams struct {
	Authority	string	`protobuf:"bytes,1,opt,name=authority,proto3" json:"authority,omitempty"`
	Params		Params	`protobuf:"bytes,2,opt,name=params,proto3" json:"params"`
}

func (m *MsgUpdateConcentrationParams) Reset()		{ *m = MsgUpdateConcentrationParams{} }
func (m *MsgUpdateConcentrationParams) String() string	{ return proto.CompactTextString(m) }
func (*MsgUpdateConcentrationParams) ProtoMessage()	{}
func (*MsgUpdateConcentrationParams) Descriptor() ([]byte, []int) {
	return fileDescriptor_c88972287abfb5a7, []int{0}
}
func (m *MsgUpdateConcentrationParams) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgUpdateConcentrationParams) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgUpdateConcentrationParams.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgUpdateConcentrationParams) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgUpdateConcentrationParams.Merge(m, src)
}
func (m *MsgUpdateConcentrationParams) XXX_Size() int {
	return m.Size()
}
func (m *MsgUpdateConcentrationParams) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgUpdateConcentrationParams.DiscardUnknown(m)
}

var xxx_messageInfo_MsgUpdateConcentrationParams proto.InternalMessageInfo

func (m *MsgUpdateConcentrationParams) GetAuthority() string {
	if m != nil {
		return m.Authority
	}
	return ""
}

func (m *MsgUpdateConcentrationParams) GetParams() Params {
	if m != nil {
		return m.Params
	}
	return Params{}
}

type MsgUpdateConcentrationParamsResponse struct {
}

func (m *MsgUpdateConcentrationParamsResponse) Reset()		{ *m = MsgUpdateConcentrationParamsResponse{} }
func (m *MsgUpdateConcentrationParamsResponse) String() string	{ return proto.CompactTextString(m) }
func (*MsgUpdateConcentrationParamsResponse) ProtoMessage()	{}
func (*MsgUpdateConcentrationParamsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_c88972287abfb5a7, []int{1}
}
func (m *MsgUpdateConcentrationParamsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgUpdateConcentrationParamsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgUpdateConcentrationParamsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgUpdateConcentrationParamsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgUpdateConcentrationParamsResponse.Merge(m, src)
}
func (m *MsgUpdateConcentrationParamsResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgUpdateConcentrationParamsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgUpdateConcentrationParamsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgUpdateConcentrationParamsResponse proto.InternalMessageInfo

type MsgRecomputeConcentration struct {
	Authority	string			`protobuf:"bytes,1,opt,name=authority,proto3" json:"authority,omitempty"`
	Epoch		uint64			`protobuf:"varint,2,opt,name=epoch,proto3" json:"epoch,omitempty"`
	ValidatorSet	[]ValidatorPower	`protobuf:"bytes,3,rep,name=validator_set,json=validatorSet,proto3" json:"validator_set"`
}

func (m *MsgRecomputeConcentration) Reset()		{ *m = MsgRecomputeConcentration{} }
func (m *MsgRecomputeConcentration) String() string	{ return proto.CompactTextString(m) }
func (*MsgRecomputeConcentration) ProtoMessage()	{}
func (*MsgRecomputeConcentration) Descriptor() ([]byte, []int) {
	return fileDescriptor_c88972287abfb5a7, []int{2}
}
func (m *MsgRecomputeConcentration) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgRecomputeConcentration) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgRecomputeConcentration.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgRecomputeConcentration) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgRecomputeConcentration.Merge(m, src)
}
func (m *MsgRecomputeConcentration) XXX_Size() int {
	return m.Size()
}
func (m *MsgRecomputeConcentration) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgRecomputeConcentration.DiscardUnknown(m)
}

var xxx_messageInfo_MsgRecomputeConcentration proto.InternalMessageInfo

func (m *MsgRecomputeConcentration) GetAuthority() string {
	if m != nil {
		return m.Authority
	}
	return ""
}

func (m *MsgRecomputeConcentration) GetEpoch() uint64 {
	if m != nil {
		return m.Epoch
	}
	return 0
}

func (m *MsgRecomputeConcentration) GetValidatorSet() []ValidatorPower {
	if m != nil {
		return m.ValidatorSet
	}
	return nil
}

type MsgRecomputeConcentrationResponse struct {
	Network NetworkConcentration `protobuf:"bytes,1,opt,name=network,proto3" json:"network"`
}

func (m *MsgRecomputeConcentrationResponse) Reset()		{ *m = MsgRecomputeConcentrationResponse{} }
func (m *MsgRecomputeConcentrationResponse) String() string	{ return proto.CompactTextString(m) }
func (*MsgRecomputeConcentrationResponse) ProtoMessage()	{}
func (*MsgRecomputeConcentrationResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_c88972287abfb5a7, []int{3}
}
func (m *MsgRecomputeConcentrationResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgRecomputeConcentrationResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgRecomputeConcentrationResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgRecomputeConcentrationResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgRecomputeConcentrationResponse.Merge(m, src)
}
func (m *MsgRecomputeConcentrationResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgRecomputeConcentrationResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgRecomputeConcentrationResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgRecomputeConcentrationResponse proto.InternalMessageInfo

func (m *MsgRecomputeConcentrationResponse) GetNetwork() NetworkConcentration {
	if m != nil {
		return m.Network
	}
	return NetworkConcentration{}
}

func init() {
	proto.RegisterType((*MsgUpdateConcentrationParams)(nil), "l1.stakeconcentration.v1.MsgUpdateConcentrationParams")
	proto.RegisterType((*MsgUpdateConcentrationParamsResponse)(nil), "l1.stakeconcentration.v1.MsgUpdateConcentrationParamsResponse")
	proto.RegisterType((*MsgRecomputeConcentration)(nil), "l1.stakeconcentration.v1.MsgRecomputeConcentration")
	proto.RegisterType((*MsgRecomputeConcentrationResponse)(nil), "l1.stakeconcentration.v1.MsgRecomputeConcentrationResponse")
}

func init()	{ proto.RegisterFile("l1/stakeconcentration/v1/tx.proto", fileDescriptor_c88972287abfb5a7) }

var fileDescriptor_c88972287abfb5a7 = []byte{

	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x93, 0xc1, 0x4b, 0x1b, 0x41,
	0x14, 0xc6, 0x33, 0x46, 0x2d, 0x4e, 0xda, 0x1e, 0x16, 0x69, 0xe3, 0x22, 0xdb, 0x18, 0x8a, 0x04,
	0xc1, 0x1d, 0x36, 0x82, 0x87, 0x16, 0x3c, 0xd8, 0x73, 0x24, 0xac, 0xb4, 0x87, 0x5e, 0xca, 0xb8,
	0x3e, 0x26, 0x8b, 0xbb, 0xfb, 0x96, 0x79, 0x93, 0xa8, 0xb7, 0xd2, 0x63, 0xe9, 0xa1, 0xd0, 0x3f,
	0xa4, 0x1e, 0xfa, 0x47, 0x78, 0xf4, 0xd8, 0x53, 0x29, 0xc9, 0xc1, 0x7f, 0xa3, 0xb8, 0x9b, 0xd5,
	0x5a, 0x3b, 0x11, 0xbc, 0xcd, 0x0c, 0xdf, 0xf7, 0xcd, 0xef, 0xe3, 0xf1, 0xf8, 0x5a, 0x12, 0x08,
	0x32, 0xf2, 0x08, 0x22, 0xcc, 0x22, 0xc8, 0x8c, 0x96, 0x26, 0xc6, 0x4c, 0x8c, 0x02, 0x61, 0x4e,
	0xfc, 0x5c, 0xa3, 0x41, 0xa7, 0x99, 0x04, 0xfe, 0x5d, 0x89, 0x3f, 0x0a, 0xdc, 0xe7, 0x11, 0x52,
	0x8a, 0x24, 0x52, 0x52, 0x57, 0x8e, 0x94, 0x54, 0x69, 0x71, 0x97, 0x15, 0x2a, 0x2c, 0x8e, 0xe2,
	0xea, 0x34, 0x7d, 0x5d, 0xb7, 0xfe, 0xa5, 0x20, 0x03, 0x8a, 0xa9, 0xd4, 0xb5, 0xbf, 0x30, 0xbe,
	0xda, 0x23, 0xf5, 0x36, 0x3f, 0x94, 0x06, 0xde, 0xfc, 0xad, 0xed, 0x4b, 0x2d, 0x53, 0x72, 0x56,
	0xf9, 0x92, 0x1c, 0x9a, 0x01, 0xea, 0xd8, 0x9c, 0x36, 0x59, 0x8b, 0x75, 0x96, 0xc2, 0x9b, 0x07,
	0x67, 0x87, 0x2f, 0xe6, 0x85, 0xae, 0x39, 0xd7, 0x62, 0x9d, 0x46, 0xb7, 0xe5, 0xdb, 0x0a, 0xf8,
	0x65, 0xde, 0xee, 0xfc, 0xf9, 0xaf, 0x17, 0xb5, 0x70, 0xea, 0x7a, 0xf5, 0xf4, 0xd3, 0xe5, 0xd9,
	0xc6, 0x4d, 0x5e, 0x7b, 0x9d, 0xbf, 0x9c, 0x45, 0x13, 0x02, 0xe5, 0x98, 0x11, 0xb4, 0x7f, 0x30,
	0xbe, 0xd2, 0x23, 0x15, 0x42, 0x84, 0x69, 0x3e, 0xfc, 0x47, 0x7b, 0x0f, 0xf3, 0x32, 0x5f, 0x80,
	0x1c, 0xa3, 0x41, 0x81, 0x3c, 0x1f, 0x96, 0x17, 0x67, 0x9f, 0x3f, 0x19, 0xc9, 0x24, 0x3e, 0x94,
	0x06, 0xf5, 0x07, 0x02, 0xd3, 0xac, 0xb7, 0xea, 0x9d, 0x46, 0xb7, 0x63, 0x2f, 0xf4, 0xae, 0x92,
	0xf7, 0xf1, 0x18, 0xf4, 0xb4, 0xd8, 0xe3, 0xeb, 0x90, 0x7d, 0x30, 0x77, 0xea, 0x11, 0x5f, 0xb3,
	0x52, 0x57, 0xdd, 0x9c, 0x3d, 0xfe, 0x28, 0x03, 0x73, 0x8c, 0xfa, 0xa8, 0x60, 0x6f, 0x74, 0x7d,
	0x3b, 0xc3, 0x5e, 0x29, 0xbc, 0x15, 0x34, 0x25, 0xa9, 0x42, 0xba, 0xdf, 0xe7, 0x78, 0xbd, 0x47,
	0xca, 0xf9, 0xc6, 0xf8, 0x8a, 0x7d, 0xce, 0xdb, 0xf6, 0x4f, 0x66, 0x4d, 0xc4, 0xdd, 0x79, 0x98,
	0xef, 0xba, 0xed, 0x67, 0xc6, 0x9f, 0x59, 0xc6, 0xb8, 0x35, 0x33, 0xfa, 0xff, 0x26, 0xf7, 0xf5,
	0x03, 0x4c, 0x15, 0x8c, 0xbb, 0xf0, 0xf1, 0xf2, 0x6c, 0x83, 0xed, 0xf6, 0xcf, 0xc7, 0x1e, 0xbb,
	0x18, 0x7b, 0xec, 0xf7, 0xd8, 0x63, 0x5f, 0x27, 0x5e, 0xed, 0x62, 0xe2, 0xd5, 0x7e, 0x4e, 0xbc,
	0xda, 0xfb, 0x6d, 0x15, 0x9b, 0xc1, 0xf0, 0xc0, 0x8f, 0x30, 0x15, 0x84, 0x23, 0xd0, 0x10, 0xab,
	0x6c, 0x33, 0x09, 0x44, 0x12, 0x88, 0x93, 0x72, 0xe1, 0x36, 0x6f, 0x6f, 0x9c, 0x39, 0xcd, 0x81,
	0x0e, 0x16, 0x8b, 0x6d, 0xdb, 0xfa, 0x13, 0x00, 0x00, 0xff, 0xff, 0xab, 0x5a, 0x6f, 0x88, 0x03,
	0x04, 0x00, 0x00,
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
	UpdateConcentrationParams(ctx context.Context, in *MsgUpdateConcentrationParams, opts ...grpc.CallOption) (*MsgUpdateConcentrationParamsResponse, error)
	RecomputeConcentration(ctx context.Context, in *MsgRecomputeConcentration, opts ...grpc.CallOption) (*MsgRecomputeConcentrationResponse, error)
}

type msgClient struct {
	cc grpc1.ClientConn
}

func NewMsgClient(cc grpc1.ClientConn) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) UpdateConcentrationParams(ctx context.Context, in *MsgUpdateConcentrationParams, opts ...grpc.CallOption) (*MsgUpdateConcentrationParamsResponse, error) {
	out := new(MsgUpdateConcentrationParamsResponse)
	err := c.cc.Invoke(ctx, "/l1.stakeconcentration.v1.Msg/UpdateConcentrationParams", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) RecomputeConcentration(ctx context.Context, in *MsgRecomputeConcentration, opts ...grpc.CallOption) (*MsgRecomputeConcentrationResponse, error) {
	out := new(MsgRecomputeConcentrationResponse)
	err := c.cc.Invoke(ctx, "/l1.stakeconcentration.v1.Msg/RecomputeConcentration", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MsgServer is the server API for Msg service.
type MsgServer interface {
	UpdateConcentrationParams(context.Context, *MsgUpdateConcentrationParams) (*MsgUpdateConcentrationParamsResponse, error)
	RecomputeConcentration(context.Context, *MsgRecomputeConcentration) (*MsgRecomputeConcentrationResponse, error)
}

// UnimplementedMsgServer can be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (*UnimplementedMsgServer) UpdateConcentrationParams(ctx context.Context, req *MsgUpdateConcentrationParams) (*MsgUpdateConcentrationParamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateConcentrationParams not implemented")
}
func (*UnimplementedMsgServer) RecomputeConcentration(ctx context.Context, req *MsgRecomputeConcentration) (*MsgRecomputeConcentrationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RecomputeConcentration not implemented")
}

func RegisterMsgServer(s grpc1.Server, srv MsgServer) {
	s.RegisterService(&_Msg_serviceDesc, srv)
}

func _Msg_UpdateConcentrationParams_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUpdateConcentrationParams)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).UpdateConcentrationParams(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.stakeconcentration.v1.Msg/UpdateConcentrationParams",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).UpdateConcentrationParams(ctx, req.(*MsgUpdateConcentrationParams))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_RecomputeConcentration_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgRecomputeConcentration)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).RecomputeConcentration(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.stakeconcentration.v1.Msg/RecomputeConcentration",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).RecomputeConcentration(ctx, req.(*MsgRecomputeConcentration))
	}
	return interceptor(ctx, in, info, handler)
}

var Msg_serviceDesc = _Msg_serviceDesc
var _Msg_serviceDesc = grpc.ServiceDesc{
	ServiceName:	"l1.stakeconcentration.v1.Msg",
	HandlerType:	(*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName:	"UpdateConcentrationParams",
			Handler:	_Msg_UpdateConcentrationParams_Handler,
		},
		{
			MethodName:	"RecomputeConcentration",
			Handler:	_Msg_RecomputeConcentration_Handler,
		},
	},
	Streams:	[]grpc.StreamDesc{},
	Metadata:	"l1/stakeconcentration/v1/tx.proto",
}

func (m *MsgUpdateConcentrationParams) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgUpdateConcentrationParams) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgUpdateConcentrationParams) MarshalToSizedBuffer(dAtA []byte) (int, error) {
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

func (m *MsgUpdateConcentrationParamsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgUpdateConcentrationParamsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgUpdateConcentrationParamsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *MsgRecomputeConcentration) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgRecomputeConcentration) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgRecomputeConcentration) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.ValidatorSet) > 0 {
		for iNdEx := len(m.ValidatorSet) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.ValidatorSet[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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

func (m *MsgRecomputeConcentrationResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgRecomputeConcentrationResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgRecomputeConcentrationResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
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
func (m *MsgUpdateConcentrationParams) Size() (n int) {
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

func (m *MsgUpdateConcentrationParamsResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *MsgRecomputeConcentration) Size() (n int) {
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
	if len(m.ValidatorSet) > 0 {
		for _, e := range m.ValidatorSet {
			l = e.Size()
			n += 1 + l + sovTx(uint64(l))
		}
	}
	return n
}

func (m *MsgRecomputeConcentrationResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Network.Size()
	n += 1 + l + sovTx(uint64(l))
	return n
}

func sovTx(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozTx(x uint64) (n int) {
	return sovTx(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *MsgUpdateConcentrationParams) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgUpdateConcentrationParams: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgUpdateConcentrationParams: illegal tag %d (wire type %d)", fieldNum, wire)
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
func (m *MsgUpdateConcentrationParamsResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgUpdateConcentrationParamsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgUpdateConcentrationParamsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
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
func (m *MsgRecomputeConcentration) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgRecomputeConcentration: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgRecomputeConcentration: illegal tag %d (wire type %d)", fieldNum, wire)
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
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ValidatorSet", wireType)
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
			m.ValidatorSet = append(m.ValidatorSet, ValidatorPower{})
			if err := m.ValidatorSet[len(m.ValidatorSet)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
func (m *MsgRecomputeConcentrationResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgRecomputeConcentrationResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgRecomputeConcentrationResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Network", wireType)
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
			if err := m.Network.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
