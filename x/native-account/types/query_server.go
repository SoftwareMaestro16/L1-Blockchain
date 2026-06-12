package types

import (
	"context"

	"github.com/cosmos/gogoproto/grpc"
	gogoproto "github.com/cosmos/gogoproto/proto"
	grpcgo "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type QueryServer interface {
	Account(context.Context, *QueryAccountRequest) (*QueryAccountResponse, error)
	AccountByRaw(context.Context, *QueryAccountByRawRequest) (*QueryAccountResponse, error)
	VirtualAccount(context.Context, *QueryVirtualAccountRequest) (*QueryVirtualAccountResponse, error)
	Params(context.Context, *QueryParamsRequest) (*QueryParamsResponse, error)
	AccountStatus(context.Context, *QueryAccountStatusRequest) (*QueryAccountStatusResponse, error)
}

type UnimplementedQueryServer struct{}

func (UnimplementedQueryServer) Account(context.Context, *QueryAccountRequest) (*QueryAccountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Account not implemented")
}
func (UnimplementedQueryServer) AccountByRaw(context.Context, *QueryAccountByRawRequest) (*QueryAccountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AccountByRaw not implemented")
}
func (UnimplementedQueryServer) VirtualAccount(context.Context, *QueryVirtualAccountRequest) (*QueryVirtualAccountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method VirtualAccount not implemented")
}
func (UnimplementedQueryServer) Params(context.Context, *QueryParamsRequest) (*QueryParamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Params not implemented")
}
func (UnimplementedQueryServer) AccountStatus(context.Context, *QueryAccountStatusRequest) (*QueryAccountStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AccountStatus not implemented")
}

func RegisterQueryServer(s grpc.Server, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc, srv)
}

func _Query_Account_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryAccountRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).Account(ctx, in)
	}
	info := &grpcgo.UnaryServerInfo{Server: srv, FullMethod: "/l1.nativeaccount.v1.Query/Account"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Account(ctx, req.(*QueryAccountRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_AccountByRaw_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryAccountByRawRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).AccountByRaw(ctx, in)
	}
	info := &grpcgo.UnaryServerInfo{Server: srv, FullMethod: "/l1.nativeaccount.v1.Query/AccountByRaw"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).AccountByRaw(ctx, req.(*QueryAccountByRawRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_VirtualAccount_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryVirtualAccountRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).VirtualAccount(ctx, in)
	}
	info := &grpcgo.UnaryServerInfo{Server: srv, FullMethod: "/l1.nativeaccount.v1.Query/VirtualAccount"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).VirtualAccount(ctx, req.(*QueryVirtualAccountRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_Params_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryParamsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).Params(ctx, in)
	}
	info := &grpcgo.UnaryServerInfo{Server: srv, FullMethod: "/l1.nativeaccount.v1.Query/Params"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Params(ctx, req.(*QueryParamsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_AccountStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryAccountStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).AccountStatus(ctx, in)
	}
	info := &grpcgo.UnaryServerInfo{Server: srv, FullMethod: "/l1.nativeaccount.v1.Query/AccountStatus"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).AccountStatus(ctx, req.(*QueryAccountStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var Query_serviceDesc = _Query_serviceDesc

var _Query_serviceDesc = grpcgo.ServiceDesc{
	ServiceName:	"l1.nativeaccount.v1.Query",
	HandlerType:	(*QueryServer)(nil),
	Methods: []grpcgo.MethodDesc{
		{MethodName: "Account", Handler: _Query_Account_Handler},
		{MethodName: "AccountByRaw", Handler: _Query_AccountByRaw_Handler},
		{MethodName: "VirtualAccount", Handler: _Query_VirtualAccount_Handler},
		{MethodName: "Params", Handler: _Query_Params_Handler},
		{MethodName: "AccountStatus", Handler: _Query_AccountStatus_Handler},
	},
	Streams:	[]grpcgo.StreamDesc{},
	Metadata:	"l1/nativeaccount/v1/query.proto",
}

func (m *QueryAccountRequest) Reset()		{ *m = QueryAccountRequest{} }
func (m *QueryAccountRequest) String() string	{ return gogoproto.CompactTextString(m) }
func (*QueryAccountRequest) ProtoMessage()	{}
func (*QueryAccountRequest) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountQuery, []int{0}
}

func (m *QueryAccountResponse) Reset()		{ *m = QueryAccountResponse{} }
func (m *QueryAccountResponse) String() string	{ return gogoproto.CompactTextString(m) }
func (*QueryAccountResponse) ProtoMessage()	{}
func (*QueryAccountResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountQuery, []int{1}
}

func (m *QueryAccountByRawRequest) Reset()		{ *m = QueryAccountByRawRequest{} }
func (m *QueryAccountByRawRequest) String() string	{ return gogoproto.CompactTextString(m) }
func (*QueryAccountByRawRequest) ProtoMessage()		{}
func (*QueryAccountByRawRequest) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountQuery, []int{2}
}

func (m *QueryVirtualAccountRequest) Reset()		{ *m = QueryVirtualAccountRequest{} }
func (m *QueryVirtualAccountRequest) String() string	{ return gogoproto.CompactTextString(m) }
func (*QueryVirtualAccountRequest) ProtoMessage()	{}
func (*QueryVirtualAccountRequest) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountQuery, []int{3}
}

func (m *QueryVirtualAccountResponse) Reset()		{ *m = QueryVirtualAccountResponse{} }
func (m *QueryVirtualAccountResponse) String() string	{ return gogoproto.CompactTextString(m) }
func (*QueryVirtualAccountResponse) ProtoMessage()	{}
func (*QueryVirtualAccountResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountQuery, []int{4}
}

func (m *QueryParamsRequest) Reset()		{ *m = QueryParamsRequest{} }
func (m *QueryParamsRequest) String() string	{ return gogoproto.CompactTextString(m) }
func (*QueryParamsRequest) ProtoMessage()	{}
func (*QueryParamsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountQuery, []int{5}
}

func (m *QueryParamsResponse) Reset()		{ *m = QueryParamsResponse{} }
func (m *QueryParamsResponse) String() string	{ return gogoproto.CompactTextString(m) }
func (*QueryParamsResponse) ProtoMessage()	{}
func (*QueryParamsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountQuery, []int{6}
}

func (m *QueryAccountStatusRequest) Reset()		{ *m = QueryAccountStatusRequest{} }
func (m *QueryAccountStatusRequest) String() string	{ return gogoproto.CompactTextString(m) }
func (*QueryAccountStatusRequest) ProtoMessage()	{}
func (*QueryAccountStatusRequest) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountQuery, []int{7}
}

func (m *QueryAccountStatusResponse) Reset()		{ *m = QueryAccountStatusResponse{} }
func (m *QueryAccountStatusResponse) String() string	{ return gogoproto.CompactTextString(m) }
func (*QueryAccountStatusResponse) ProtoMessage()	{}
func (*QueryAccountStatusResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountQuery, []int{8}
}

func init() {
	gogoproto.RegisterType((*QueryAccountRequest)(nil), "l1.nativeaccount.v1.QueryAccountRequest")
	gogoproto.RegisterType((*QueryAccountResponse)(nil), "l1.nativeaccount.v1.QueryAccountResponse")
	gogoproto.RegisterType((*QueryAccountByRawRequest)(nil), "l1.nativeaccount.v1.QueryAccountByRawRequest")
	gogoproto.RegisterType((*QueryVirtualAccountRequest)(nil), "l1.nativeaccount.v1.QueryVirtualAccountRequest")
	gogoproto.RegisterType((*QueryVirtualAccountResponse)(nil), "l1.nativeaccount.v1.QueryVirtualAccountResponse")
	gogoproto.RegisterType((*QueryParamsRequest)(nil), "l1.nativeaccount.v1.QueryParamsRequest")
	gogoproto.RegisterType((*QueryParamsResponse)(nil), "l1.nativeaccount.v1.QueryParamsResponse")
	gogoproto.RegisterType((*QueryAccountStatusRequest)(nil), "l1.nativeaccount.v1.QueryAccountStatusRequest")
	gogoproto.RegisterType((*QueryAccountStatusResponse)(nil), "l1.nativeaccount.v1.QueryAccountStatusResponse")
	gogoproto.RegisterFile("l1/nativeaccount/v1/query.proto", fileDescriptorNativeAccountQuery)
}
