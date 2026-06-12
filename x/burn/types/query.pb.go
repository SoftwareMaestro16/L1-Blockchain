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

type QueryBurnedByDenomRequest struct {
	Denom string `protobuf:"bytes,1,opt,name=denom,proto3" json:"denom,omitempty"`
}

func (m *QueryBurnedByDenomRequest) Reset()		{ *m = QueryBurnedByDenomRequest{} }
func (m *QueryBurnedByDenomRequest) String() string	{ return proto.CompactTextString(m) }
func (*QueryBurnedByDenomRequest) ProtoMessage()	{}
func (*QueryBurnedByDenomRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_72dc5f9e73dcf9a6, []int{0}
}
func (m *QueryBurnedByDenomRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryBurnedByDenomRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryBurnedByDenomRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryBurnedByDenomRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryBurnedByDenomRequest.Merge(m, src)
}
func (m *QueryBurnedByDenomRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryBurnedByDenomRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryBurnedByDenomRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryBurnedByDenomRequest proto.InternalMessageInfo

func (m *QueryBurnedByDenomRequest) GetDenom() string {
	if m != nil {
		return m.Denom
	}
	return ""
}

type QueryBurnedByDenomResponse struct {
	BurnedByDenom []BurnedByDenomEntry `protobuf:"bytes,1,rep,name=burned_by_denom,json=burnedByDenom,proto3" json:"burned_by_denom"`
}

func (m *QueryBurnedByDenomResponse) Reset()		{ *m = QueryBurnedByDenomResponse{} }
func (m *QueryBurnedByDenomResponse) String() string	{ return proto.CompactTextString(m) }
func (*QueryBurnedByDenomResponse) ProtoMessage()	{}
func (*QueryBurnedByDenomResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_72dc5f9e73dcf9a6, []int{1}
}
func (m *QueryBurnedByDenomResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryBurnedByDenomResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryBurnedByDenomResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryBurnedByDenomResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryBurnedByDenomResponse.Merge(m, src)
}
func (m *QueryBurnedByDenomResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryBurnedByDenomResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryBurnedByDenomResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryBurnedByDenomResponse proto.InternalMessageInfo

func (m *QueryBurnedByDenomResponse) GetBurnedByDenom() []BurnedByDenomEntry {
	if m != nil {
		return m.BurnedByDenom
	}
	return nil
}

type QueryBurnedByEpochRequest struct {
	Epoch uint64 `protobuf:"varint,1,opt,name=epoch,proto3" json:"epoch,omitempty"`
}

func (m *QueryBurnedByEpochRequest) Reset()		{ *m = QueryBurnedByEpochRequest{} }
func (m *QueryBurnedByEpochRequest) String() string	{ return proto.CompactTextString(m) }
func (*QueryBurnedByEpochRequest) ProtoMessage()	{}
func (*QueryBurnedByEpochRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_72dc5f9e73dcf9a6, []int{2}
}
func (m *QueryBurnedByEpochRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryBurnedByEpochRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryBurnedByEpochRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryBurnedByEpochRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryBurnedByEpochRequest.Merge(m, src)
}
func (m *QueryBurnedByEpochRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryBurnedByEpochRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryBurnedByEpochRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryBurnedByEpochRequest proto.InternalMessageInfo

func (m *QueryBurnedByEpochRequest) GetEpoch() uint64 {
	if m != nil {
		return m.Epoch
	}
	return 0
}

type QueryBurnedByEpochResponse struct {
	BurnedByEpoch []BurnedByEpochEntry `protobuf:"bytes,1,rep,name=burned_by_epoch,json=burnedByEpoch,proto3" json:"burned_by_epoch"`
}

func (m *QueryBurnedByEpochResponse) Reset()		{ *m = QueryBurnedByEpochResponse{} }
func (m *QueryBurnedByEpochResponse) String() string	{ return proto.CompactTextString(m) }
func (*QueryBurnedByEpochResponse) ProtoMessage()	{}
func (*QueryBurnedByEpochResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_72dc5f9e73dcf9a6, []int{3}
}
func (m *QueryBurnedByEpochResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryBurnedByEpochResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryBurnedByEpochResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryBurnedByEpochResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryBurnedByEpochResponse.Merge(m, src)
}
func (m *QueryBurnedByEpochResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryBurnedByEpochResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryBurnedByEpochResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryBurnedByEpochResponse proto.InternalMessageInfo

func (m *QueryBurnedByEpochResponse) GetBurnedByEpoch() []BurnedByEpochEntry {
	if m != nil {
		return m.BurnedByEpoch
	}
	return nil
}

type QueryBurnParamsRequest struct {
}

func (m *QueryBurnParamsRequest) Reset()		{ *m = QueryBurnParamsRequest{} }
func (m *QueryBurnParamsRequest) String() string	{ return proto.CompactTextString(m) }
func (*QueryBurnParamsRequest) ProtoMessage()		{}
func (*QueryBurnParamsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_72dc5f9e73dcf9a6, []int{4}
}
func (m *QueryBurnParamsRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryBurnParamsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryBurnParamsRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryBurnParamsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryBurnParamsRequest.Merge(m, src)
}
func (m *QueryBurnParamsRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryBurnParamsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryBurnParamsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryBurnParamsRequest proto.InternalMessageInfo

type QueryBurnParamsResponse struct {
	Params Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
}

func (m *QueryBurnParamsResponse) Reset()		{ *m = QueryBurnParamsResponse{} }
func (m *QueryBurnParamsResponse) String() string	{ return proto.CompactTextString(m) }
func (*QueryBurnParamsResponse) ProtoMessage()		{}
func (*QueryBurnParamsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_72dc5f9e73dcf9a6, []int{5}
}
func (m *QueryBurnParamsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryBurnParamsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryBurnParamsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryBurnParamsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryBurnParamsResponse.Merge(m, src)
}
func (m *QueryBurnParamsResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryBurnParamsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryBurnParamsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryBurnParamsResponse proto.InternalMessageInfo

func (m *QueryBurnParamsResponse) GetParams() Params {
	if m != nil {
		return m.Params
	}
	return Params{}
}

func init() {
	proto.RegisterType((*QueryBurnedByDenomRequest)(nil), "l1.burn.v1.QueryBurnedByDenomRequest")
	proto.RegisterType((*QueryBurnedByDenomResponse)(nil), "l1.burn.v1.QueryBurnedByDenomResponse")
	proto.RegisterType((*QueryBurnedByEpochRequest)(nil), "l1.burn.v1.QueryBurnedByEpochRequest")
	proto.RegisterType((*QueryBurnedByEpochResponse)(nil), "l1.burn.v1.QueryBurnedByEpochResponse")
	proto.RegisterType((*QueryBurnParamsRequest)(nil), "l1.burn.v1.QueryBurnParamsRequest")
	proto.RegisterType((*QueryBurnParamsResponse)(nil), "l1.burn.v1.QueryBurnParamsResponse")
}

func init()	{ proto.RegisterFile("l1/burn/v1/query.proto", fileDescriptor_72dc5f9e73dcf9a6) }

var fileDescriptor_72dc5f9e73dcf9a6 = []byte{

	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x93, 0x4f, 0x8f, 0xd2, 0x40,
	0x18, 0xc6, 0x5b, 0xf9, 0x93, 0x38, 0x84, 0x98, 0x4c, 0x08, 0x62, 0xd5, 0x6a, 0x4a, 0x54, 0x2e,
	0x76, 0x2c, 0x7e, 0x02, 0x1b, 0x39, 0xe9, 0x41, 0x7b, 0xf4, 0x42, 0x5a, 0x98, 0x94, 0x9a, 0x32,
	0x53, 0x3a, 0x2d, 0xb1, 0x57, 0x4d, 0x3c, 0x6f, 0xb2, 0x5f, 0x8a, 0x23, 0xc9, 0x5e, 0xf6, 0xb4,
	0xd9, 0xc0, 0x7e, 0x90, 0x4d, 0xa7, 0xc3, 0x32, 0x5d, 0x0a, 0xdc, 0xe6, 0x9d, 0xf7, 0x7d, 0xde,
	0xe7, 0xc7, 0x3c, 0x14, 0x74, 0x43, 0x0b, 0x79, 0x69, 0x4c, 0xd0, 0xd2, 0x42, 0x8b, 0x14, 0xc7,
	0x99, 0x19, 0xc5, 0x34, 0xa1, 0x10, 0x84, 0x96, 0x99, 0xdf, 0x9b, 0x4b, 0x4b, 0x7b, 0xe5, 0x53,
	0xea, 0x87, 0x18, 0xb9, 0x51, 0x80, 0x5c, 0x42, 0x68, 0xe2, 0x26, 0x01, 0x25, 0xac, 0x98, 0xd4,
	0x3a, 0x3e, 0xf5, 0x29, 0x3f, 0xa2, 0xfc, 0x24, 0x6e, 0x7b, 0xd2, 0x5e, 0x1f, 0x13, 0xcc, 0x02,
	0x31, 0x6f, 0x58, 0xe0, 0xc5, 0xcf, 0xdc, 0xc8, 0x4e, 0x63, 0x82, 0xa7, 0x76, 0xf6, 0x15, 0x13,
	0x3a, 0x77, 0xf0, 0x22, 0xc5, 0x2c, 0x81, 0x1d, 0xd0, 0x98, 0xe6, 0x75, 0x4f, 0x7d, 0xab, 0x0e,
	0x9e, 0x3a, 0x45, 0x61, 0xfc, 0x06, 0x5a, 0x95, 0x84, 0x45, 0x94, 0x30, 0x0c, 0xbf, 0x83, 0x67,
	0x1e, 0x6f, 0x8c, 0xbd, 0x6c, 0xbc, 0x53, 0xd7, 0x06, 0xad, 0xa1, 0x6e, 0xee, 0x7f, 0x84, 0x59,
	0xd2, 0x8e, 0x48, 0x12, 0x67, 0x76, 0x7d, 0x75, 0xf3, 0x46, 0x71, 0xda, 0x9e, 0xdc, 0x39, 0xc0,
	0x1b, 0x45, 0x74, 0x32, 0x93, 0xf0, 0x70, 0x5e, 0x73, 0xbc, 0xba, 0x53, 0x14, 0x07, 0x78, 0x42,
	0x52, 0x85, 0xb7, 0x53, 0x1f, 0xc5, 0xe3, 0xda, 0x4a, 0x3c, 0xde, 0x31, 0x7a, 0xa0, 0xfb, 0xe0,
	0xf5, 0xc3, 0x8d, 0xdd, 0x39, 0x13, 0x6c, 0xc6, 0x37, 0xf0, 0xfc, 0xa0, 0x23, 0x10, 0x3e, 0x81,
	0x66, 0xc4, 0x6f, 0x38, 0x77, 0x6b, 0x08, 0x65, 0xe7, 0x62, 0x56, 0xb8, 0x89, 0xb9, 0xe1, 0xff,
	0x1a, 0x68, 0xf0, 0x6d, 0xf0, 0x9f, 0x0a, 0xda, 0xa5, 0xb7, 0x83, 0xef, 0x64, 0xf5, 0xd1, 0x28,
	0xb5, 0xf7, 0xe7, 0xc6, 0x0a, 0x38, 0xa3, 0xff, 0xf7, 0xea, 0xee, 0xf2, 0xc9, 0x6b, 0xf8, 0x12,
	0x49, 0x7f, 0x99, 0x47, 0x81, 0x96, 0x28, 0xf8, 0x43, 0x9c, 0xa0, 0x90, 0x13, 0x3b, 0x41, 0x51,
	0x4a, 0xe9, 0x1c, 0x05, 0xcf, 0x0d, 0x2e, 0x00, 0xd8, 0xbf, 0x2e, 0x34, 0x2a, 0x57, 0x97, 0x42,
	0xd1, 0xfa, 0x27, 0x67, 0x84, 0xb7, 0xc6, 0xbd, 0x3b, 0x10, 0xca, 0xde, 0x45, 0x10, 0xf6, 0x97,
	0xd5, 0x46, 0x57, 0xd7, 0x1b, 0x5d, 0xbd, 0xdd, 0xe8, 0xea, 0xc5, 0x56, 0x57, 0xd6, 0x5b, 0x5d,
	0xb9, 0xde, 0xea, 0xca, 0xaf, 0x0f, 0x7e, 0x90, 0xcc, 0x52, 0xcf, 0x9c, 0xd0, 0x39, 0x62, 0x74,
	0x89, 0x63, 0x1c, 0xf8, 0xe4, 0x63, 0x68, 0xe5, 0x4b, 0xfe, 0x14, 0x6b, 0x92, 0x2c, 0xc2, 0xcc,
	0x6b, 0xf2, 0xef, 0xee, 0xf3, 0x7d, 0x00, 0x00, 0x00, 0xff, 0xff, 0x47, 0x0e, 0x39, 0xcd, 0xeb,
	0x03, 0x00, 0x00,
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
	BurnedByDenom(ctx context.Context, in *QueryBurnedByDenomRequest, opts ...grpc.CallOption) (*QueryBurnedByDenomResponse, error)
	BurnedByEpoch(ctx context.Context, in *QueryBurnedByEpochRequest, opts ...grpc.CallOption) (*QueryBurnedByEpochResponse, error)
	BurnParams(ctx context.Context, in *QueryBurnParamsRequest, opts ...grpc.CallOption) (*QueryBurnParamsResponse, error)
}

type queryClient struct {
	cc grpc1.ClientConn
}

func NewQueryClient(cc grpc1.ClientConn) QueryClient {
	return &queryClient{cc}
}

func (c *queryClient) BurnedByDenom(ctx context.Context, in *QueryBurnedByDenomRequest, opts ...grpc.CallOption) (*QueryBurnedByDenomResponse, error) {
	out := new(QueryBurnedByDenomResponse)
	err := c.cc.Invoke(ctx, "/l1.burn.v1.Query/BurnedByDenom", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) BurnedByEpoch(ctx context.Context, in *QueryBurnedByEpochRequest, opts ...grpc.CallOption) (*QueryBurnedByEpochResponse, error) {
	out := new(QueryBurnedByEpochResponse)
	err := c.cc.Invoke(ctx, "/l1.burn.v1.Query/BurnedByEpoch", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) BurnParams(ctx context.Context, in *QueryBurnParamsRequest, opts ...grpc.CallOption) (*QueryBurnParamsResponse, error) {
	out := new(QueryBurnParamsResponse)
	err := c.cc.Invoke(ctx, "/l1.burn.v1.Query/BurnParams", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QueryServer is the server API for Query service.
type QueryServer interface {
	BurnedByDenom(context.Context, *QueryBurnedByDenomRequest) (*QueryBurnedByDenomResponse, error)
	BurnedByEpoch(context.Context, *QueryBurnedByEpochRequest) (*QueryBurnedByEpochResponse, error)
	BurnParams(context.Context, *QueryBurnParamsRequest) (*QueryBurnParamsResponse, error)
}

// UnimplementedQueryServer can be embedded to have forward compatible implementations.
type UnimplementedQueryServer struct {
}

func (*UnimplementedQueryServer) BurnedByDenom(ctx context.Context, req *QueryBurnedByDenomRequest) (*QueryBurnedByDenomResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BurnedByDenom not implemented")
}
func (*UnimplementedQueryServer) BurnedByEpoch(ctx context.Context, req *QueryBurnedByEpochRequest) (*QueryBurnedByEpochResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BurnedByEpoch not implemented")
}
func (*UnimplementedQueryServer) BurnParams(ctx context.Context, req *QueryBurnParamsRequest) (*QueryBurnParamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BurnParams not implemented")
}

func RegisterQueryServer(s grpc1.Server, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc, srv)
}

func _Query_BurnedByDenom_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryBurnedByDenomRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).BurnedByDenom(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.burn.v1.Query/BurnedByDenom",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).BurnedByDenom(ctx, req.(*QueryBurnedByDenomRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_BurnedByEpoch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryBurnedByEpochRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).BurnedByEpoch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.burn.v1.Query/BurnedByEpoch",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).BurnedByEpoch(ctx, req.(*QueryBurnedByEpochRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_BurnParams_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryBurnParamsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).BurnParams(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.burn.v1.Query/BurnParams",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).BurnParams(ctx, req.(*QueryBurnParamsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var Query_serviceDesc = _Query_serviceDesc
var _Query_serviceDesc = grpc.ServiceDesc{
	ServiceName:	"l1.burn.v1.Query",
	HandlerType:	(*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName:	"BurnedByDenom",
			Handler:	_Query_BurnedByDenom_Handler,
		},
		{
			MethodName:	"BurnedByEpoch",
			Handler:	_Query_BurnedByEpoch_Handler,
		},
		{
			MethodName:	"BurnParams",
			Handler:	_Query_BurnParams_Handler,
		},
	},
	Streams:	[]grpc.StreamDesc{},
	Metadata:	"l1/burn/v1/query.proto",
}

func (m *QueryBurnedByDenomRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryBurnedByDenomRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryBurnedByDenomRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Denom) > 0 {
		i -= len(m.Denom)
		copy(dAtA[i:], m.Denom)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.Denom)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *QueryBurnedByDenomResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryBurnedByDenomResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryBurnedByDenomResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.BurnedByDenom) > 0 {
		for iNdEx := len(m.BurnedByDenom) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.BurnedByDenom[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintQuery(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *QueryBurnedByEpochRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryBurnedByEpochRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryBurnedByEpochRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Epoch != 0 {
		i = encodeVarintQuery(dAtA, i, uint64(m.Epoch))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *QueryBurnedByEpochResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryBurnedByEpochResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryBurnedByEpochResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.BurnedByEpoch) > 0 {
		for iNdEx := len(m.BurnedByEpoch) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.BurnedByEpoch[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintQuery(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *QueryBurnParamsRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryBurnParamsRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryBurnParamsRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *QueryBurnParamsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryBurnParamsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryBurnParamsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
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
func (m *QueryBurnedByDenomRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Denom)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *QueryBurnedByDenomResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.BurnedByDenom) > 0 {
		for _, e := range m.BurnedByDenom {
			l = e.Size()
			n += 1 + l + sovQuery(uint64(l))
		}
	}
	return n
}

func (m *QueryBurnedByEpochRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Epoch != 0 {
		n += 1 + sovQuery(uint64(m.Epoch))
	}
	return n
}

func (m *QueryBurnedByEpochResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.BurnedByEpoch) > 0 {
		for _, e := range m.BurnedByEpoch {
			l = e.Size()
			n += 1 + l + sovQuery(uint64(l))
		}
	}
	return n
}

func (m *QueryBurnParamsRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *QueryBurnParamsResponse) Size() (n int) {
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
func (m *QueryBurnedByDenomRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryBurnedByDenomRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryBurnedByDenomRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Denom", wireType)
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
			m.Denom = string(dAtA[iNdEx:postIndex])
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
func (m *QueryBurnedByDenomResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryBurnedByDenomResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryBurnedByDenomResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BurnedByDenom", wireType)
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
			m.BurnedByDenom = append(m.BurnedByDenom, BurnedByDenomEntry{})
			if err := m.BurnedByDenom[len(m.BurnedByDenom)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
func (m *QueryBurnedByEpochRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryBurnedByEpochRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryBurnedByEpochRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Epoch", wireType)
			}
			m.Epoch = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
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
func (m *QueryBurnedByEpochResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryBurnedByEpochResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryBurnedByEpochResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BurnedByEpoch", wireType)
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
			m.BurnedByEpoch = append(m.BurnedByEpoch, BurnedByEpochEntry{})
			if err := m.BurnedByEpoch[len(m.BurnedByEpoch)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
func (m *QueryBurnParamsRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryBurnParamsRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryBurnParamsRequest: illegal tag %d (wire type %d)", fieldNum, wire)
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
func (m *QueryBurnParamsResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryBurnParamsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryBurnParamsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
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
