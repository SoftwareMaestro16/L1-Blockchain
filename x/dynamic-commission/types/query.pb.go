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

type QueryValidatorCommissionRequest struct {
	ValidatorAddress string `protobuf:"bytes,1,opt,name=validator_address,json=validatorAddress,proto3" json:"validator_address,omitempty"`
}

func (m *QueryValidatorCommissionRequest) Reset()		{ *m = QueryValidatorCommissionRequest{} }
func (m *QueryValidatorCommissionRequest) String() string	{ return proto.CompactTextString(m) }
func (*QueryValidatorCommissionRequest) ProtoMessage()		{}
func (*QueryValidatorCommissionRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_1b86f633401d69ec, []int{0}
}
func (m *QueryValidatorCommissionRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryValidatorCommissionRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryValidatorCommissionRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryValidatorCommissionRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryValidatorCommissionRequest.Merge(m, src)
}
func (m *QueryValidatorCommissionRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryValidatorCommissionRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryValidatorCommissionRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryValidatorCommissionRequest proto.InternalMessageInfo

func (m *QueryValidatorCommissionRequest) GetValidatorAddress() string {
	if m != nil {
		return m.ValidatorAddress
	}
	return ""
}

type QueryValidatorCommissionResponse struct {
	Commission ValidatorCommission `protobuf:"bytes,1,opt,name=commission,proto3" json:"commission"`
}

func (m *QueryValidatorCommissionResponse) Reset()		{ *m = QueryValidatorCommissionResponse{} }
func (m *QueryValidatorCommissionResponse) String() string	{ return proto.CompactTextString(m) }
func (*QueryValidatorCommissionResponse) ProtoMessage()		{}
func (*QueryValidatorCommissionResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_1b86f633401d69ec, []int{1}
}
func (m *QueryValidatorCommissionResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryValidatorCommissionResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryValidatorCommissionResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryValidatorCommissionResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryValidatorCommissionResponse.Merge(m, src)
}
func (m *QueryValidatorCommissionResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryValidatorCommissionResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryValidatorCommissionResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryValidatorCommissionResponse proto.InternalMessageInfo

func (m *QueryValidatorCommissionResponse) GetCommission() ValidatorCommission {
	if m != nil {
		return m.Commission
	}
	return ValidatorCommission{}
}

type QueryCommissionHistoryRequest struct {
	ValidatorAddress string `protobuf:"bytes,1,opt,name=validator_address,json=validatorAddress,proto3" json:"validator_address,omitempty"`
}

func (m *QueryCommissionHistoryRequest) Reset()		{ *m = QueryCommissionHistoryRequest{} }
func (m *QueryCommissionHistoryRequest) String() string	{ return proto.CompactTextString(m) }
func (*QueryCommissionHistoryRequest) ProtoMessage()	{}
func (*QueryCommissionHistoryRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_1b86f633401d69ec, []int{2}
}
func (m *QueryCommissionHistoryRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryCommissionHistoryRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryCommissionHistoryRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryCommissionHistoryRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryCommissionHistoryRequest.Merge(m, src)
}
func (m *QueryCommissionHistoryRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryCommissionHistoryRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryCommissionHistoryRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryCommissionHistoryRequest proto.InternalMessageInfo

func (m *QueryCommissionHistoryRequest) GetValidatorAddress() string {
	if m != nil {
		return m.ValidatorAddress
	}
	return ""
}

type QueryCommissionHistoryResponse struct {
	History []CommissionHistoryEntry `protobuf:"bytes,1,rep,name=history,proto3" json:"history"`
}

func (m *QueryCommissionHistoryResponse) Reset()		{ *m = QueryCommissionHistoryResponse{} }
func (m *QueryCommissionHistoryResponse) String() string	{ return proto.CompactTextString(m) }
func (*QueryCommissionHistoryResponse) ProtoMessage()		{}
func (*QueryCommissionHistoryResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_1b86f633401d69ec, []int{3}
}
func (m *QueryCommissionHistoryResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryCommissionHistoryResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryCommissionHistoryResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryCommissionHistoryResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryCommissionHistoryResponse.Merge(m, src)
}
func (m *QueryCommissionHistoryResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryCommissionHistoryResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryCommissionHistoryResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryCommissionHistoryResponse proto.InternalMessageInfo

func (m *QueryCommissionHistoryResponse) GetHistory() []CommissionHistoryEntry {
	if m != nil {
		return m.History
	}
	return nil
}

type QueryCommissionParamsRequest struct {
}

func (m *QueryCommissionParamsRequest) Reset()		{ *m = QueryCommissionParamsRequest{} }
func (m *QueryCommissionParamsRequest) String() string	{ return proto.CompactTextString(m) }
func (*QueryCommissionParamsRequest) ProtoMessage()	{}
func (*QueryCommissionParamsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_1b86f633401d69ec, []int{4}
}
func (m *QueryCommissionParamsRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryCommissionParamsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryCommissionParamsRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryCommissionParamsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryCommissionParamsRequest.Merge(m, src)
}
func (m *QueryCommissionParamsRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryCommissionParamsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryCommissionParamsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryCommissionParamsRequest proto.InternalMessageInfo

type QueryCommissionParamsResponse struct {
	Params Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
}

func (m *QueryCommissionParamsResponse) Reset()		{ *m = QueryCommissionParamsResponse{} }
func (m *QueryCommissionParamsResponse) String() string	{ return proto.CompactTextString(m) }
func (*QueryCommissionParamsResponse) ProtoMessage()	{}
func (*QueryCommissionParamsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_1b86f633401d69ec, []int{5}
}
func (m *QueryCommissionParamsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryCommissionParamsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryCommissionParamsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryCommissionParamsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryCommissionParamsResponse.Merge(m, src)
}
func (m *QueryCommissionParamsResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryCommissionParamsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryCommissionParamsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryCommissionParamsResponse proto.InternalMessageInfo

func (m *QueryCommissionParamsResponse) GetParams() Params {
	if m != nil {
		return m.Params
	}
	return Params{}
}

func init() {
	proto.RegisterType((*QueryValidatorCommissionRequest)(nil), "l1.dynamiccommission.v1.QueryValidatorCommissionRequest")
	proto.RegisterType((*QueryValidatorCommissionResponse)(nil), "l1.dynamiccommission.v1.QueryValidatorCommissionResponse")
	proto.RegisterType((*QueryCommissionHistoryRequest)(nil), "l1.dynamiccommission.v1.QueryCommissionHistoryRequest")
	proto.RegisterType((*QueryCommissionHistoryResponse)(nil), "l1.dynamiccommission.v1.QueryCommissionHistoryResponse")
	proto.RegisterType((*QueryCommissionParamsRequest)(nil), "l1.dynamiccommission.v1.QueryCommissionParamsRequest")
	proto.RegisterType((*QueryCommissionParamsResponse)(nil), "l1.dynamiccommission.v1.QueryCommissionParamsResponse")
}

func init() {
	proto.RegisterFile("l1/dynamiccommission/v1/query.proto", fileDescriptor_1b86f633401d69ec)
}

var fileDescriptor_1b86f633401d69ec = []byte{

	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x94, 0xc1, 0x6a, 0x13, 0x41,
	0x18, 0xc7, 0x33, 0x5a, 0x2b, 0x4e, 0x2f, 0xed, 0x28, 0x58, 0x96, 0xba, 0x89, 0x2b, 0x62, 0x41,
	0xbb, 0xc3, 0x46, 0x5a, 0xf5, 0x20, 0xd2, 0xa8, 0x28, 0x22, 0x56, 0x73, 0xf0, 0xe0, 0x41, 0x99,
	0x26, 0xc3, 0x76, 0x60, 0x77, 0x66, 0x33, 0x33, 0x59, 0x5c, 0xc4, 0x8b, 0x4f, 0x20, 0xf8, 0x16,
	0x3e, 0x49, 0x2f, 0x42, 0x41, 0x10, 0x2f, 0x8a, 0x24, 0x3e, 0x88, 0x64, 0x76, 0x92, 0xd4, 0x6e,
	0x47, 0x9b, 0xdc, 0x96, 0x6f, 0xbe, 0xef, 0xff, 0xfd, 0x7f, 0x33, 0x7f, 0x16, 0x5e, 0x49, 0x22,
	0xdc, 0x2d, 0x38, 0x49, 0x59, 0xa7, 0x23, 0xd2, 0x94, 0x29, 0xc5, 0x04, 0xc7, 0x79, 0x84, 0x7b,
	0x7d, 0x2a, 0x8b, 0x30, 0x93, 0x42, 0x0b, 0x74, 0x31, 0x89, 0xc2, 0x4a, 0x53, 0x98, 0x47, 0xde,
	0x5a, 0x2c, 0x44, 0x9c, 0x50, 0x4c, 0x32, 0x86, 0x09, 0xe7, 0x42, 0x13, 0xcd, 0x04, 0x57, 0xe5,
	0x98, 0x77, 0x21, 0x16, 0xb1, 0x30, 0x9f, 0x78, 0xf4, 0x65, 0xab, 0x57, 0x5d, 0x1b, 0x63, 0xca,
	0xa9, 0x62, 0x76, 0x38, 0x78, 0x06, 0xeb, 0x2f, 0x46, 0x16, 0x5e, 0x92, 0x84, 0x75, 0x89, 0x16,
	0xf2, 0xfe, 0xa4, 0xb9, 0x4d, 0x7b, 0x7d, 0xaa, 0x34, 0xba, 0x0e, 0x57, 0xf2, 0xf1, 0xe9, 0x1b,
	0xd2, 0xed, 0x4a, 0xaa, 0xd4, 0x2a, 0x68, 0x80, 0xf5, 0x73, 0xed, 0xe5, 0xc9, 0xc1, 0x76, 0x59,
	0x0f, 0x72, 0xd8, 0x70, 0xeb, 0xa9, 0x4c, 0x70, 0x45, 0x51, 0x1b, 0xc2, 0xa9, 0x25, 0xa3, 0xb4,
	0xd4, 0xbc, 0x11, 0x3a, 0xe0, 0xc3, 0x63, 0x94, 0x5a, 0x0b, 0xfb, 0x3f, 0xeb, 0xb5, 0xf6, 0x21,
	0x95, 0xe0, 0x29, 0xbc, 0x64, 0xf6, 0x4e, 0x9b, 0x1e, 0x33, 0xa5, 0x85, 0x2c, 0xe6, 0xa2, 0xe8,
	0x41, 0xdf, 0xa5, 0x66, 0x19, 0x76, 0xe0, 0xd9, 0xbd, 0xb2, 0xb4, 0x0a, 0x1a, 0xa7, 0xd7, 0x97,
	0x9a, 0xd8, 0x09, 0x50, 0x11, 0x79, 0xc8, 0xb5, 0x2c, 0x2c, 0xc3, 0x58, 0x25, 0xf0, 0xe1, 0xda,
	0x91, 0x95, 0xcf, 0x89, 0x24, 0xa9, 0xb2, 0xfe, 0x83, 0xd7, 0x15, 0xc0, 0xf1, 0xb9, 0x75, 0x74,
	0x17, 0x2e, 0x66, 0xa6, 0x62, 0x6f, 0xb4, 0xee, 0x34, 0x54, 0x0e, 0x5a, 0x03, 0x76, 0xa8, 0xf9,
	0x63, 0x01, 0x9e, 0x31, 0x0b, 0xd0, 0x37, 0x00, 0xcf, 0x1f, 0x73, 0xe9, 0xe8, 0xb6, 0x53, 0xf0,
	0x3f, 0x09, 0xf2, 0xee, 0xcc, 0x31, 0x59, 0x52, 0x05, 0x4f, 0x3e, 0x7c, 0xfd, 0xfd, 0xe9, 0xd4,
	0x03, 0xd4, 0xc2, 0xae, 0x3c, 0x4f, 0x1e, 0x4f, 0xe1, 0x77, 0x95, 0x17, 0x7e, 0x8f, 0xa7, 0xcd,
	0xe8, 0x0b, 0x80, 0x2b, 0x95, 0xc7, 0x40, 0x5b, 0xff, 0x36, 0xe7, 0x0a, 0x94, 0x77, 0x6b, 0xe6,
	0x39, 0x8b, 0xf4, 0xc8, 0x20, 0x6d, 0xa3, 0x7b, 0xf3, 0x22, 0xd9, 0xc8, 0xa0, 0xcf, 0x00, 0x2e,
	0x1f, 0x8d, 0x03, 0xda, 0x3c, 0xa9, 0xad, 0xbf, 0xe2, 0xe5, 0x6d, 0xcd, 0x3a, 0x66, 0x61, 0xae,
	0x19, 0x98, 0xcb, 0xa8, 0xee, 0x84, 0x29, 0xf3, 0xd5, 0xda, 0xd9, 0x1f, 0xf8, 0xe0, 0x60, 0xe0,
	0x83, 0x5f, 0x03, 0x1f, 0x7c, 0x1c, 0xfa, 0xb5, 0x83, 0xa1, 0x5f, 0xfb, 0x3e, 0xf4, 0x6b, 0xaf,
	0x36, 0x63, 0xa6, 0xf7, 0xfa, 0xbb, 0x61, 0x47, 0xa4, 0x58, 0x89, 0x9c, 0x4a, 0xca, 0x62, 0xbe,
	0x91, 0x44, 0x23, 0xc5, 0xb7, 0x63, 0xcd, 0x8d, 0x43, 0xa2, 0xba, 0xc8, 0xa8, 0xda, 0x5d, 0x34,
	0x3f, 0xb0, 0x9b, 0x7f, 0x02, 0x00, 0x00, 0xff, 0xff, 0x73, 0x78, 0x8a, 0x08, 0x5b, 0x05, 0x00,
	0x00,
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
	ValidatorCommission(ctx context.Context, in *QueryValidatorCommissionRequest, opts ...grpc.CallOption) (*QueryValidatorCommissionResponse, error)
	CommissionHistory(ctx context.Context, in *QueryCommissionHistoryRequest, opts ...grpc.CallOption) (*QueryCommissionHistoryResponse, error)
	CommissionParams(ctx context.Context, in *QueryCommissionParamsRequest, opts ...grpc.CallOption) (*QueryCommissionParamsResponse, error)
}

type queryClient struct {
	cc grpc1.ClientConn
}

func NewQueryClient(cc grpc1.ClientConn) QueryClient {
	return &queryClient{cc}
}

func (c *queryClient) ValidatorCommission(ctx context.Context, in *QueryValidatorCommissionRequest, opts ...grpc.CallOption) (*QueryValidatorCommissionResponse, error) {
	out := new(QueryValidatorCommissionResponse)
	err := c.cc.Invoke(ctx, "/l1.dynamiccommission.v1.Query/ValidatorCommission", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) CommissionHistory(ctx context.Context, in *QueryCommissionHistoryRequest, opts ...grpc.CallOption) (*QueryCommissionHistoryResponse, error) {
	out := new(QueryCommissionHistoryResponse)
	err := c.cc.Invoke(ctx, "/l1.dynamiccommission.v1.Query/CommissionHistory", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) CommissionParams(ctx context.Context, in *QueryCommissionParamsRequest, opts ...grpc.CallOption) (*QueryCommissionParamsResponse, error) {
	out := new(QueryCommissionParamsResponse)
	err := c.cc.Invoke(ctx, "/l1.dynamiccommission.v1.Query/CommissionParams", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QueryServer is the server API for Query service.
type QueryServer interface {
	ValidatorCommission(context.Context, *QueryValidatorCommissionRequest) (*QueryValidatorCommissionResponse, error)
	CommissionHistory(context.Context, *QueryCommissionHistoryRequest) (*QueryCommissionHistoryResponse, error)
	CommissionParams(context.Context, *QueryCommissionParamsRequest) (*QueryCommissionParamsResponse, error)
}

// UnimplementedQueryServer can be embedded to have forward compatible implementations.
type UnimplementedQueryServer struct {
}

func (*UnimplementedQueryServer) ValidatorCommission(ctx context.Context, req *QueryValidatorCommissionRequest) (*QueryValidatorCommissionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ValidatorCommission not implemented")
}
func (*UnimplementedQueryServer) CommissionHistory(ctx context.Context, req *QueryCommissionHistoryRequest) (*QueryCommissionHistoryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CommissionHistory not implemented")
}
func (*UnimplementedQueryServer) CommissionParams(ctx context.Context, req *QueryCommissionParamsRequest) (*QueryCommissionParamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CommissionParams not implemented")
}

func RegisterQueryServer(s grpc1.Server, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc, srv)
}

func _Query_ValidatorCommission_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryValidatorCommissionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).ValidatorCommission(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.dynamiccommission.v1.Query/ValidatorCommission",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).ValidatorCommission(ctx, req.(*QueryValidatorCommissionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_CommissionHistory_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryCommissionHistoryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).CommissionHistory(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.dynamiccommission.v1.Query/CommissionHistory",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).CommissionHistory(ctx, req.(*QueryCommissionHistoryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_CommissionParams_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryCommissionParamsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).CommissionParams(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.dynamiccommission.v1.Query/CommissionParams",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).CommissionParams(ctx, req.(*QueryCommissionParamsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var Query_serviceDesc = _Query_serviceDesc
var _Query_serviceDesc = grpc.ServiceDesc{
	ServiceName:	"l1.dynamiccommission.v1.Query",
	HandlerType:	(*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName:	"ValidatorCommission",
			Handler:	_Query_ValidatorCommission_Handler,
		},
		{
			MethodName:	"CommissionHistory",
			Handler:	_Query_CommissionHistory_Handler,
		},
		{
			MethodName:	"CommissionParams",
			Handler:	_Query_CommissionParams_Handler,
		},
	},
	Streams:	[]grpc.StreamDesc{},
	Metadata:	"l1/dynamiccommission/v1/query.proto",
}

func (m *QueryValidatorCommissionRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryValidatorCommissionRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryValidatorCommissionRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.ValidatorAddress) > 0 {
		i -= len(m.ValidatorAddress)
		copy(dAtA[i:], m.ValidatorAddress)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.ValidatorAddress)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *QueryValidatorCommissionResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryValidatorCommissionResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryValidatorCommissionResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.Commission.MarshalToSizedBuffer(dAtA[:i])
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

func (m *QueryCommissionHistoryRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryCommissionHistoryRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryCommissionHistoryRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.ValidatorAddress) > 0 {
		i -= len(m.ValidatorAddress)
		copy(dAtA[i:], m.ValidatorAddress)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.ValidatorAddress)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *QueryCommissionHistoryResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryCommissionHistoryResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryCommissionHistoryResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.History) > 0 {
		for iNdEx := len(m.History) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.History[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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

func (m *QueryCommissionParamsRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryCommissionParamsRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryCommissionParamsRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *QueryCommissionParamsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryCommissionParamsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryCommissionParamsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
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
func (m *QueryValidatorCommissionRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ValidatorAddress)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *QueryValidatorCommissionResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Commission.Size()
	n += 1 + l + sovQuery(uint64(l))
	return n
}

func (m *QueryCommissionHistoryRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ValidatorAddress)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *QueryCommissionHistoryResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.History) > 0 {
		for _, e := range m.History {
			l = e.Size()
			n += 1 + l + sovQuery(uint64(l))
		}
	}
	return n
}

func (m *QueryCommissionParamsRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *QueryCommissionParamsResponse) Size() (n int) {
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
func (m *QueryValidatorCommissionRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryValidatorCommissionRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryValidatorCommissionRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ValidatorAddress", wireType)
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
			m.ValidatorAddress = string(dAtA[iNdEx:postIndex])
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
func (m *QueryValidatorCommissionResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryValidatorCommissionResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryValidatorCommissionResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Commission", wireType)
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
			if err := m.Commission.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
func (m *QueryCommissionHistoryRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryCommissionHistoryRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryCommissionHistoryRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ValidatorAddress", wireType)
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
			m.ValidatorAddress = string(dAtA[iNdEx:postIndex])
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
func (m *QueryCommissionHistoryResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryCommissionHistoryResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryCommissionHistoryResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field History", wireType)
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
			m.History = append(m.History, CommissionHistoryEntry{})
			if err := m.History[len(m.History)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
func (m *QueryCommissionParamsRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryCommissionParamsRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryCommissionParamsRequest: illegal tag %d (wire type %d)", fieldNum, wire)
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
func (m *QueryCommissionParamsResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryCommissionParamsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryCommissionParamsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
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
