package types

import (
	context "context"
	fmt "fmt"
	_ "github.com/cosmos/gogoproto/gogoproto"
	grpc1 "github.com/cosmos/gogoproto/grpc"
	proto "github.com/cosmos/gogoproto/proto"
	_ "google.golang.org/genproto/googleapis/api/annotations"
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

type QueryValidatorConcentrationRequest struct {
	OperatorAddress string `protobuf:"bytes,1,opt,name=operator_address,json=operatorAddress,proto3" json:"operator_address,omitempty"`
}

func (m *QueryValidatorConcentrationRequest) Reset()		{ *m = QueryValidatorConcentrationRequest{} }
func (m *QueryValidatorConcentrationRequest) String() string	{ return proto.CompactTextString(m) }
func (*QueryValidatorConcentrationRequest) ProtoMessage()	{}
func (*QueryValidatorConcentrationRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_aebe53aeb54e269c, []int{0}
}
func (m *QueryValidatorConcentrationRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryValidatorConcentrationRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryValidatorConcentrationRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryValidatorConcentrationRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryValidatorConcentrationRequest.Merge(m, src)
}
func (m *QueryValidatorConcentrationRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryValidatorConcentrationRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryValidatorConcentrationRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryValidatorConcentrationRequest proto.InternalMessageInfo

func (m *QueryValidatorConcentrationRequest) GetOperatorAddress() string {
	if m != nil {
		return m.OperatorAddress
	}
	return ""
}

type QueryValidatorConcentrationResponse struct {
	Concentration ValidatorConcentration `protobuf:"bytes,1,opt,name=concentration,proto3" json:"concentration"`
}

func (m *QueryValidatorConcentrationResponse) Reset()		{ *m = QueryValidatorConcentrationResponse{} }
func (m *QueryValidatorConcentrationResponse) String() string	{ return proto.CompactTextString(m) }
func (*QueryValidatorConcentrationResponse) ProtoMessage()	{}
func (*QueryValidatorConcentrationResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_aebe53aeb54e269c, []int{1}
}
func (m *QueryValidatorConcentrationResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryValidatorConcentrationResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryValidatorConcentrationResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryValidatorConcentrationResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryValidatorConcentrationResponse.Merge(m, src)
}
func (m *QueryValidatorConcentrationResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryValidatorConcentrationResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryValidatorConcentrationResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryValidatorConcentrationResponse proto.InternalMessageInfo

func (m *QueryValidatorConcentrationResponse) GetConcentration() ValidatorConcentration {
	if m != nil {
		return m.Concentration
	}
	return ValidatorConcentration{}
}

type QueryNetworkConcentrationRequest struct {
}

func (m *QueryNetworkConcentrationRequest) Reset()		{ *m = QueryNetworkConcentrationRequest{} }
func (m *QueryNetworkConcentrationRequest) String() string	{ return proto.CompactTextString(m) }
func (*QueryNetworkConcentrationRequest) ProtoMessage()		{}
func (*QueryNetworkConcentrationRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_aebe53aeb54e269c, []int{2}
}
func (m *QueryNetworkConcentrationRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryNetworkConcentrationRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryNetworkConcentrationRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryNetworkConcentrationRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryNetworkConcentrationRequest.Merge(m, src)
}
func (m *QueryNetworkConcentrationRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryNetworkConcentrationRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryNetworkConcentrationRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryNetworkConcentrationRequest proto.InternalMessageInfo

type QueryNetworkConcentrationResponse struct {
	Network NetworkConcentration `protobuf:"bytes,1,opt,name=network,proto3" json:"network"`
}

func (m *QueryNetworkConcentrationResponse) Reset()		{ *m = QueryNetworkConcentrationResponse{} }
func (m *QueryNetworkConcentrationResponse) String() string	{ return proto.CompactTextString(m) }
func (*QueryNetworkConcentrationResponse) ProtoMessage()	{}
func (*QueryNetworkConcentrationResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_aebe53aeb54e269c, []int{3}
}
func (m *QueryNetworkConcentrationResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryNetworkConcentrationResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryNetworkConcentrationResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryNetworkConcentrationResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryNetworkConcentrationResponse.Merge(m, src)
}
func (m *QueryNetworkConcentrationResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryNetworkConcentrationResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryNetworkConcentrationResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryNetworkConcentrationResponse proto.InternalMessageInfo

func (m *QueryNetworkConcentrationResponse) GetNetwork() NetworkConcentration {
	if m != nil {
		return m.Network
	}
	return NetworkConcentration{}
}

type QueryConcentrationParamsRequest struct {
}

func (m *QueryConcentrationParamsRequest) Reset()		{ *m = QueryConcentrationParamsRequest{} }
func (m *QueryConcentrationParamsRequest) String() string	{ return proto.CompactTextString(m) }
func (*QueryConcentrationParamsRequest) ProtoMessage()		{}
func (*QueryConcentrationParamsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_aebe53aeb54e269c, []int{4}
}
func (m *QueryConcentrationParamsRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryConcentrationParamsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryConcentrationParamsRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryConcentrationParamsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryConcentrationParamsRequest.Merge(m, src)
}
func (m *QueryConcentrationParamsRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryConcentrationParamsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryConcentrationParamsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryConcentrationParamsRequest proto.InternalMessageInfo

type QueryConcentrationParamsResponse struct {
	Params Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
}

func (m *QueryConcentrationParamsResponse) Reset()		{ *m = QueryConcentrationParamsResponse{} }
func (m *QueryConcentrationParamsResponse) String() string	{ return proto.CompactTextString(m) }
func (*QueryConcentrationParamsResponse) ProtoMessage()		{}
func (*QueryConcentrationParamsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_aebe53aeb54e269c, []int{5}
}
func (m *QueryConcentrationParamsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryConcentrationParamsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryConcentrationParamsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryConcentrationParamsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryConcentrationParamsResponse.Merge(m, src)
}
func (m *QueryConcentrationParamsResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryConcentrationParamsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryConcentrationParamsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryConcentrationParamsResponse proto.InternalMessageInfo

func (m *QueryConcentrationParamsResponse) GetParams() Params {
	if m != nil {
		return m.Params
	}
	return Params{}
}

func init() {
	proto.RegisterType((*QueryValidatorConcentrationRequest)(nil), "l1.stakeconcentration.v1.QueryValidatorConcentrationRequest")
	proto.RegisterType((*QueryValidatorConcentrationResponse)(nil), "l1.stakeconcentration.v1.QueryValidatorConcentrationResponse")
	proto.RegisterType((*QueryNetworkConcentrationRequest)(nil), "l1.stakeconcentration.v1.QueryNetworkConcentrationRequest")
	proto.RegisterType((*QueryNetworkConcentrationResponse)(nil), "l1.stakeconcentration.v1.QueryNetworkConcentrationResponse")
	proto.RegisterType((*QueryConcentrationParamsRequest)(nil), "l1.stakeconcentration.v1.QueryConcentrationParamsRequest")
	proto.RegisterType((*QueryConcentrationParamsResponse)(nil), "l1.stakeconcentration.v1.QueryConcentrationParamsResponse")
}

func init() {
	proto.RegisterFile("l1/stakeconcentration/v1/query.proto", fileDescriptor_aebe53aeb54e269c)
}

var fileDescriptor_aebe53aeb54e269c = []byte{

	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x94, 0x4f, 0x6b, 0xd4, 0x40,
	0x18, 0xc6, 0x33, 0x52, 0x2b, 0x8e, 0x88, 0x32, 0x16, 0x29, 0x41, 0xd2, 0xec, 0x54, 0xa4, 0x3d,
	0x34, 0x63, 0x2a, 0x28, 0xad, 0xff, 0xb0, 0xde, 0x6b, 0xdd, 0x83, 0x07, 0x11, 0x64, 0x76, 0x77,
	0x88, 0xa1, 0xe9, 0x4c, 0x3a, 0x33, 0x1b, 0x2d, 0xe2, 0x45, 0xbf, 0x80, 0xe0, 0x77, 0xf1, 0xe0,
	0x27, 0xd8, 0xe3, 0x82, 0xb0, 0x78, 0x12, 0xd9, 0xf5, 0x83, 0x48, 0x26, 0x13, 0x30, 0x9a, 0xd9,
	0x85, 0xbd, 0x85, 0x77, 0xde, 0xe7, 0x79, 0x7f, 0xef, 0xcc, 0x43, 0xe0, 0xcd, 0x2c, 0x26, 0x4a,
	0xd3, 0x63, 0xd6, 0x17, 0xbc, 0xcf, 0xb8, 0x96, 0x54, 0xa7, 0x82, 0x93, 0x22, 0x26, 0xa7, 0x43,
	0x26, 0xcf, 0xa2, 0x5c, 0x0a, 0x2d, 0xd0, 0x7a, 0x16, 0x47, 0xff, 0x77, 0x45, 0x45, 0xec, 0xdf,
	0x48, 0x84, 0x48, 0x32, 0x46, 0x68, 0x9e, 0x12, 0xca, 0xb9, 0xd0, 0xe6, 0x44, 0x55, 0x3a, 0x7f,
	0x2d, 0x11, 0x89, 0x30, 0x9f, 0xa4, 0xfc, 0xb2, 0xd5, 0x5b, 0xce, 0x99, 0x09, 0xe3, 0x4c, 0xa5,
	0x56, 0x8d, 0x9f, 0x41, 0xfc, 0xbc, 0x84, 0x78, 0x41, 0xb3, 0x74, 0x40, 0xb5, 0x90, 0x4f, 0xff,
	0xee, 0xef, 0xb2, 0xd3, 0x21, 0x53, 0x1a, 0x6d, 0xc3, 0xab, 0x22, 0x67, 0xb2, 0x3c, 0x7f, 0x4d,
	0x07, 0x03, 0xc9, 0x94, 0x5a, 0x07, 0x21, 0xd8, 0xba, 0xd8, 0xbd, 0x52, 0xd7, 0x9f, 0x54, 0x65,
	0xfc, 0x09, 0xc0, 0xcd, 0xb9, 0x8e, 0x2a, 0x17, 0x5c, 0x31, 0xf4, 0x0a, 0x5e, 0x6e, 0xa0, 0x19,
	0xbf, 0x4b, 0xbb, 0xb7, 0x23, 0xd7, 0x35, 0x44, 0xed, 0x86, 0x07, 0x2b, 0xa3, 0x9f, 0x1b, 0x5e,
	0xb7, 0x69, 0x86, 0x31, 0x0c, 0x0d, 0xc4, 0x21, 0xd3, 0x6f, 0x85, 0x3c, 0x6e, 0x5b, 0x0a, 0x2b,
	0xd8, 0x99, 0xd3, 0x63, 0x31, 0x0f, 0xe1, 0x05, 0x5e, 0x9d, 0x5b, 0xc0, 0xc8, 0x0d, 0xd8, 0x66,
	0x64, 0xf1, 0x6a, 0x13, 0xdc, 0x81, 0x1b, 0x66, 0x68, 0xa3, 0xe9, 0x88, 0x4a, 0x7a, 0xa2, 0x6a,
	0xae, 0x9e, 0x65, 0x6f, 0x6d, 0xb1, 0x58, 0x8f, 0xe0, 0x6a, 0x6e, 0x2a, 0x96, 0x2a, 0x74, 0x53,
	0x55, 0x4a, 0xcb, 0x61, 0x55, 0xbb, 0x93, 0x15, 0x78, 0xde, 0x0c, 0x41, 0x13, 0x00, 0xaf, 0xb7,
	0xdf, 0x2c, 0x7a, 0xe0, 0x36, 0x5d, 0x9c, 0x19, 0xff, 0xe1, 0x92, 0xea, 0x6a, 0x43, 0xfc, 0xf8,
	0xe3, 0xf7, 0xdf, 0x5f, 0xce, 0xed, 0xa1, 0x7b, 0xc4, 0x99, 0xe4, 0xa2, 0x76, 0x50, 0xe4, 0xfd,
	0xbf, 0xf1, 0xfc, 0x80, 0xbe, 0x01, 0xb8, 0xd6, 0xf6, 0x22, 0x68, 0x7f, 0x01, 0xd8, 0x9c, 0xcc,
	0xf8, 0xf7, 0x97, 0xd2, 0xda, 0x95, 0xb6, 0xcd, 0x4a, 0x9b, 0xa8, 0xe3, 0x5e, 0xc9, 0xc6, 0x04,
	0x7d, 0x05, 0xf0, 0x5a, 0xcb, 0xfb, 0xa3, 0xbd, 0x05, 0xf3, 0xdd, 0xb1, 0xf2, 0xf7, 0x97, 0x91,
	0x5a, 0xf2, 0x2d, 0x43, 0x8e, 0x51, 0xe8, 0x26, 0xaf, 0x82, 0x75, 0x70, 0x34, 0x9a, 0x06, 0x60,
	0x3c, 0x0d, 0xc0, 0xaf, 0x69, 0x00, 0x3e, 0xcf, 0x02, 0x6f, 0x3c, 0x0b, 0xbc, 0x1f, 0xb3, 0xc0,
	0x7b, 0x79, 0x37, 0x49, 0xf5, 0x9b, 0x61, 0x2f, 0xea, 0x8b, 0x13, 0xa2, 0x44, 0xc1, 0x24, 0x4b,
	0x13, 0xbe, 0x93, 0xc5, 0xa5, 0xe5, 0xbb, 0xca, 0x74, 0xa7, 0xe9, 0xaa, 0xcf, 0x72, 0xa6, 0x7a,
	0xab, 0xe6, 0x47, 0x75, 0xe7, 0x4f, 0x00, 0x00, 0x00, 0xff, 0xff, 0xba, 0xb1, 0xb3, 0x57, 0x46,
	0x05, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// QueryClient is the client API for Query service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type QueryClient interface {
	ValidatorConcentration(ctx context.Context, in *QueryValidatorConcentrationRequest, opts ...grpc.CallOption) (*QueryValidatorConcentrationResponse, error)
	NetworkConcentration(ctx context.Context, in *QueryNetworkConcentrationRequest, opts ...grpc.CallOption) (*QueryNetworkConcentrationResponse, error)
	ConcentrationParams(ctx context.Context, in *QueryConcentrationParamsRequest, opts ...grpc.CallOption) (*QueryConcentrationParamsResponse, error)
}

type queryClient struct {
	cc grpc1.ClientConn
}

func NewQueryClient(cc grpc1.ClientConn) QueryClient {
	return &queryClient{cc}
}

func (c *queryClient) ValidatorConcentration(ctx context.Context, in *QueryValidatorConcentrationRequest, opts ...grpc.CallOption) (*QueryValidatorConcentrationResponse, error) {
	out := new(QueryValidatorConcentrationResponse)
	err := c.cc.Invoke(ctx, "/l1.stakeconcentration.v1.Query/ValidatorConcentration", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) NetworkConcentration(ctx context.Context, in *QueryNetworkConcentrationRequest, opts ...grpc.CallOption) (*QueryNetworkConcentrationResponse, error) {
	out := new(QueryNetworkConcentrationResponse)
	err := c.cc.Invoke(ctx, "/l1.stakeconcentration.v1.Query/NetworkConcentration", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) ConcentrationParams(ctx context.Context, in *QueryConcentrationParamsRequest, opts ...grpc.CallOption) (*QueryConcentrationParamsResponse, error) {
	out := new(QueryConcentrationParamsResponse)
	err := c.cc.Invoke(ctx, "/l1.stakeconcentration.v1.Query/ConcentrationParams", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QueryServer is the server API for Query service.
type QueryServer interface {
	ValidatorConcentration(context.Context, *QueryValidatorConcentrationRequest) (*QueryValidatorConcentrationResponse, error)
	NetworkConcentration(context.Context, *QueryNetworkConcentrationRequest) (*QueryNetworkConcentrationResponse, error)
	ConcentrationParams(context.Context, *QueryConcentrationParamsRequest) (*QueryConcentrationParamsResponse, error)
}

// UnimplementedQueryServer can be embedded to have forward compatible implementations.
type UnimplementedQueryServer struct {
}

func (*UnimplementedQueryServer) ValidatorConcentration(ctx context.Context, req *QueryValidatorConcentrationRequest) (*QueryValidatorConcentrationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ValidatorConcentration not implemented")
}
func (*UnimplementedQueryServer) NetworkConcentration(ctx context.Context, req *QueryNetworkConcentrationRequest) (*QueryNetworkConcentrationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NetworkConcentration not implemented")
}
func (*UnimplementedQueryServer) ConcentrationParams(ctx context.Context, req *QueryConcentrationParamsRequest) (*QueryConcentrationParamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ConcentrationParams not implemented")
}

func RegisterQueryServer(s grpc1.Server, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc, srv)
}

func _Query_ValidatorConcentration_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryValidatorConcentrationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).ValidatorConcentration(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.stakeconcentration.v1.Query/ValidatorConcentration",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).ValidatorConcentration(ctx, req.(*QueryValidatorConcentrationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_NetworkConcentration_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryNetworkConcentrationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).NetworkConcentration(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.stakeconcentration.v1.Query/NetworkConcentration",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).NetworkConcentration(ctx, req.(*QueryNetworkConcentrationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_ConcentrationParams_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryConcentrationParamsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).ConcentrationParams(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.stakeconcentration.v1.Query/ConcentrationParams",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).ConcentrationParams(ctx, req.(*QueryConcentrationParamsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var Query_serviceDesc = _Query_serviceDesc
var _Query_serviceDesc = grpc.ServiceDesc{
	ServiceName:	"l1.stakeconcentration.v1.Query",
	HandlerType:	(*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName:	"ValidatorConcentration",
			Handler:	_Query_ValidatorConcentration_Handler,
		},
		{
			MethodName:	"NetworkConcentration",
			Handler:	_Query_NetworkConcentration_Handler,
		},
		{
			MethodName:	"ConcentrationParams",
			Handler:	_Query_ConcentrationParams_Handler,
		},
	},
	Streams:	[]grpc.StreamDesc{},
	Metadata:	"l1/stakeconcentration/v1/query.proto",
}

func (m *QueryValidatorConcentrationRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryValidatorConcentrationRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryValidatorConcentrationRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.OperatorAddress) > 0 {
		i -= len(m.OperatorAddress)
		copy(dAtA[i:], m.OperatorAddress)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.OperatorAddress)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *QueryValidatorConcentrationResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryValidatorConcentrationResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryValidatorConcentrationResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.Concentration.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintQuery(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *QueryNetworkConcentrationRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryNetworkConcentrationRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryNetworkConcentrationRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *QueryNetworkConcentrationResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryNetworkConcentrationResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryNetworkConcentrationResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
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
		i = encodeVarintQuery(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *QueryConcentrationParamsRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryConcentrationParamsRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryConcentrationParamsRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *QueryConcentrationParamsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryConcentrationParamsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryConcentrationParamsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
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
		i = encodeVarintQuery(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func encodeVarintQuery(dAtA []byte, offset int, v uint64) int {
	offset -= sovQuery(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *QueryValidatorConcentrationRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.OperatorAddress)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *QueryValidatorConcentrationResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Concentration.Size()
	n += 1 + l + sovQuery(uint64(l))
	return n
}

func (m *QueryNetworkConcentrationRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *QueryNetworkConcentrationResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Network.Size()
	n += 1 + l + sovQuery(uint64(l))
	return n
}

func (m *QueryConcentrationParamsRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *QueryConcentrationParamsResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Params.Size()
	n += 1 + l + sovQuery(uint64(l))
	return n
}

func sovQuery(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozQuery(x uint64) (n int) {
	return sovQuery(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *QueryValidatorConcentrationRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
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
			return fmt.Errorf("proto: QueryValidatorConcentrationRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryValidatorConcentrationRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field OperatorAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
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
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.OperatorAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
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
func (m *QueryValidatorConcentrationResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
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
			return fmt.Errorf("proto: QueryValidatorConcentrationResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryValidatorConcentrationResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Concentration", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
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
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Concentration.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
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
func (m *QueryNetworkConcentrationRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
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
			return fmt.Errorf("proto: QueryNetworkConcentrationRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryNetworkConcentrationRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
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
func (m *QueryNetworkConcentrationResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
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
			return fmt.Errorf("proto: QueryNetworkConcentrationResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryNetworkConcentrationResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Network", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
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
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
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
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
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
func (m *QueryConcentrationParamsRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
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
			return fmt.Errorf("proto: QueryConcentrationParamsRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryConcentrationParamsRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
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
func (m *QueryConcentrationParamsResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
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
			return fmt.Errorf("proto: QueryConcentrationParamsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryConcentrationParamsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Params", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
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
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
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
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
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
func skipQuery(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowQuery
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
					return 0, ErrIntOverflowQuery
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
					return 0, ErrIntOverflowQuery
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
				return 0, ErrInvalidLengthQuery
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupQuery
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthQuery
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthQuery		= fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowQuery		= fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupQuery	= fmt.Errorf("proto: unexpected end of group")
)
