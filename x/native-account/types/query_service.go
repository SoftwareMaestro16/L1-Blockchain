package types

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"

	"github.com/cosmos/gogoproto/grpc"
	gogoproto "github.com/cosmos/gogoproto/proto"
	grpcgo "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	proto2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/sovereign-l1/l1/x/internal/prototype"
)

type QueryAccountRequest struct {
	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
}

type QueryAccountResponse struct {
	Found       bool   `protobuf:"varint,1,opt,name=found,proto3" json:"found,omitempty"`
	Virtual     bool   `protobuf:"varint,2,opt,name=virtual,proto3" json:"virtual,omitempty"`
	AddressUser string `protobuf:"bytes,3,opt,name=address_user,json=addressUser,proto3" json:"address_user,omitempty"`
	AddressRaw  string `protobuf:"bytes,4,opt,name=address_raw,json=addressRaw,proto3" json:"address_raw,omitempty"`
	Status      string `protobuf:"bytes,5,opt,name=status,proto3" json:"status,omitempty"`
	AccountJSON string `protobuf:"bytes,6,opt,name=account_json,json=accountJSON,proto3" json:"account_json,omitempty"`
}

type QueryAccountByRawRequest struct {
	AddressRaw string `protobuf:"bytes,1,opt,name=address_raw,json=addressRaw,proto3" json:"address_raw,omitempty"`
}

type QueryVirtualAccountRequest struct {
	AddressUser string `protobuf:"bytes,1,opt,name=address_user,json=addressUser,proto3" json:"address_user,omitempty"`
}

type QueryVirtualAccountResponse struct {
	AddressUser       string `protobuf:"bytes,1,opt,name=address_user,json=addressUser,proto3" json:"address_user,omitempty"`
	AddressRaw        string `protobuf:"bytes,2,opt,name=address_raw,json=addressRaw,proto3" json:"address_raw,omitempty"`
	Status            string `protobuf:"bytes,3,opt,name=status,proto3" json:"status,omitempty"`
	Persistent        bool   `protobuf:"varint,4,opt,name=persistent,proto3" json:"persistent,omitempty"`
	StorageRentActive bool   `protobuf:"varint,5,opt,name=storage_rent_active,json=storageRentActive,proto3" json:"storage_rent_active,omitempty"`
}

type QueryParamsRequest struct{}

type QueryParamsResponse struct {
	DefaultQueryLimit uint64 `protobuf:"varint,1,opt,name=default_query_limit,json=defaultQueryLimit,proto3" json:"default_query_limit,omitempty"`
	MaxQueryLimit     uint64 `protobuf:"varint,2,opt,name=max_query_limit,json=maxQueryLimit,proto3" json:"max_query_limit,omitempty"`
}

type QueryAccountStatusRequest struct {
	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
}

type QueryAccountStatusResponse struct {
	AddressUser       string `protobuf:"bytes,1,opt,name=address_user,json=addressUser,proto3" json:"address_user,omitempty"`
	AddressRaw        string `protobuf:"bytes,2,opt,name=address_raw,json=addressRaw,proto3" json:"address_raw,omitempty"`
	Status            string `protobuf:"bytes,3,opt,name=status,proto3" json:"status,omitempty"`
	Persistent        bool   `protobuf:"varint,4,opt,name=persistent,proto3" json:"persistent,omitempty"`
	StorageRentActive bool   `protobuf:"varint,5,opt,name=storage_rent_active,json=storageRentActive,proto3" json:"storage_rent_active,omitempty"`
	StorageRentDebt   uint64 `protobuf:"varint,6,opt,name=storage_rent_debt,json=storageRentDebt,proto3" json:"storage_rent_debt,omitempty"`
}

func NewQueryAccountResponse(account Account, found, virtual bool) QueryAccountResponse {
	resp := QueryAccountResponse{Found: found, Virtual: virtual}
	if found {
		resp.AddressUser = account.AddressUser
		resp.AddressRaw = account.AddressRaw
		resp.Status = account.Status
		if bz, err := json.Marshal(account); err == nil {
			resp.AccountJSON = string(bz)
		}
	}
	return resp
}

func NewQueryVirtualAccountResponse(view VirtualAccountView) QueryVirtualAccountResponse {
	return QueryVirtualAccountResponse{
		AddressUser:       view.AddressUser,
		AddressRaw:        view.AddressRaw,
		Status:            view.Status,
		Persistent:        view.Persistent,
		StorageRentActive: view.StorageRentActive,
	}
}

func NewQueryParamsResponse(params prototype.Params) QueryParamsResponse {
	return QueryParamsResponse{DefaultQueryLimit: params.DefaultQueryLimit, MaxQueryLimit: params.MaxQueryLimit}
}

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
	ServiceName: "l1.nativeaccount.v1.Query",
	HandlerType: (*QueryServer)(nil),
	Methods: []grpcgo.MethodDesc{
		{MethodName: "Account", Handler: _Query_Account_Handler},
		{MethodName: "AccountByRaw", Handler: _Query_AccountByRaw_Handler},
		{MethodName: "VirtualAccount", Handler: _Query_VirtualAccount_Handler},
		{MethodName: "Params", Handler: _Query_Params_Handler},
		{MethodName: "AccountStatus", Handler: _Query_AccountStatus_Handler},
	},
	Streams:  []grpcgo.StreamDesc{},
	Metadata: "l1/nativeaccount/v1/query.proto",
}

func (m *QueryAccountRequest) Reset()         { *m = QueryAccountRequest{} }
func (m *QueryAccountRequest) String() string { return gogoproto.CompactTextString(m) }
func (*QueryAccountRequest) ProtoMessage()    {}
func (*QueryAccountRequest) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountQuery, []int{0}
}

func (m *QueryAccountResponse) Reset()         { *m = QueryAccountResponse{} }
func (m *QueryAccountResponse) String() string { return gogoproto.CompactTextString(m) }
func (*QueryAccountResponse) ProtoMessage()    {}
func (*QueryAccountResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountQuery, []int{1}
}

func (m *QueryAccountByRawRequest) Reset()         { *m = QueryAccountByRawRequest{} }
func (m *QueryAccountByRawRequest) String() string { return gogoproto.CompactTextString(m) }
func (*QueryAccountByRawRequest) ProtoMessage()    {}
func (*QueryAccountByRawRequest) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountQuery, []int{2}
}

func (m *QueryVirtualAccountRequest) Reset()         { *m = QueryVirtualAccountRequest{} }
func (m *QueryVirtualAccountRequest) String() string { return gogoproto.CompactTextString(m) }
func (*QueryVirtualAccountRequest) ProtoMessage()    {}
func (*QueryVirtualAccountRequest) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountQuery, []int{3}
}

func (m *QueryVirtualAccountResponse) Reset()         { *m = QueryVirtualAccountResponse{} }
func (m *QueryVirtualAccountResponse) String() string { return gogoproto.CompactTextString(m) }
func (*QueryVirtualAccountResponse) ProtoMessage()    {}
func (*QueryVirtualAccountResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountQuery, []int{4}
}

func (m *QueryParamsRequest) Reset()         { *m = QueryParamsRequest{} }
func (m *QueryParamsRequest) String() string { return gogoproto.CompactTextString(m) }
func (*QueryParamsRequest) ProtoMessage()    {}
func (*QueryParamsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountQuery, []int{5}
}

func (m *QueryParamsResponse) Reset()         { *m = QueryParamsResponse{} }
func (m *QueryParamsResponse) String() string { return gogoproto.CompactTextString(m) }
func (*QueryParamsResponse) ProtoMessage()    {}
func (*QueryParamsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountQuery, []int{6}
}

func (m *QueryAccountStatusRequest) Reset()         { *m = QueryAccountStatusRequest{} }
func (m *QueryAccountStatusRequest) String() string { return gogoproto.CompactTextString(m) }
func (*QueryAccountStatusRequest) ProtoMessage()    {}
func (*QueryAccountStatusRequest) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountQuery, []int{7}
}

func (m *QueryAccountStatusResponse) Reset()         { *m = QueryAccountStatusResponse{} }
func (m *QueryAccountStatusResponse) String() string { return gogoproto.CompactTextString(m) }
func (*QueryAccountStatusResponse) ProtoMessage()    {}
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

var fileDescriptorNativeAccountQuery = buildNativeAccountQueryFileDescriptor()

func buildNativeAccountQueryFileDescriptor() []byte {
	const path = "l1/nativeaccount/v1/query.proto"
	fd := &descriptorpb.FileDescriptorProto{
		Name:    descriptorString(path),
		Package: descriptorString("l1.nativeaccount.v1"),
		Syntax:  descriptorString("proto3"),
		Options: &descriptorpb.FileOptions{
			GoPackage: descriptorString("github.com/sovereign-l1/l1/x/native-account/types"),
		},
		MessageType: []*descriptorpb.DescriptorProto{
			{Name: descriptorString("QueryAccountRequest"), Field: []*descriptorpb.FieldDescriptorProto{descriptorField("address", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING)}},
			{Name: descriptorString("QueryAccountResponse"), Field: []*descriptorpb.FieldDescriptorProto{
				descriptorField("found", 1, descriptorpb.FieldDescriptorProto_TYPE_BOOL),
				descriptorField("virtual", 2, descriptorpb.FieldDescriptorProto_TYPE_BOOL),
				descriptorField("address_user", 3, descriptorpb.FieldDescriptorProto_TYPE_STRING),
				descriptorField("address_raw", 4, descriptorpb.FieldDescriptorProto_TYPE_STRING),
				descriptorField("status", 5, descriptorpb.FieldDescriptorProto_TYPE_STRING),
				descriptorField("account_json", 6, descriptorpb.FieldDescriptorProto_TYPE_STRING),
			}},
			{Name: descriptorString("QueryAccountByRawRequest"), Field: []*descriptorpb.FieldDescriptorProto{descriptorField("address_raw", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING)}},
			{Name: descriptorString("QueryVirtualAccountRequest"), Field: []*descriptorpb.FieldDescriptorProto{descriptorField("address_user", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING)}},
			virtualAccountResponseDescriptor("QueryVirtualAccountResponse", false),
			{Name: descriptorString("QueryParamsRequest")},
			{Name: descriptorString("QueryParamsResponse"), Field: []*descriptorpb.FieldDescriptorProto{
				descriptorField("default_query_limit", 1, descriptorpb.FieldDescriptorProto_TYPE_UINT64),
				descriptorField("max_query_limit", 2, descriptorpb.FieldDescriptorProto_TYPE_UINT64),
			}},
			{Name: descriptorString("QueryAccountStatusRequest"), Field: []*descriptorpb.FieldDescriptorProto{descriptorField("address", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING)}},
			virtualAccountResponseDescriptor("QueryAccountStatusResponse", true),
		},
		Service: []*descriptorpb.ServiceDescriptorProto{
			{
				Name: descriptorString("Query"),
				Method: []*descriptorpb.MethodDescriptorProto{
					queryMethodDescriptor("Account", "QueryAccountRequest", "QueryAccountResponse"),
					queryMethodDescriptor("AccountByRaw", "QueryAccountByRawRequest", "QueryAccountResponse"),
					queryMethodDescriptor("VirtualAccount", "QueryVirtualAccountRequest", "QueryVirtualAccountResponse"),
					queryMethodDescriptor("Params", "QueryParamsRequest", "QueryParamsResponse"),
					queryMethodDescriptor("AccountStatus", "QueryAccountStatusRequest", "QueryAccountStatusResponse"),
				},
			},
		},
	}
	raw, err := proto2.Marshal(fd)
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write(raw); err != nil {
		panic(err)
	}
	if err := zw.Close(); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func queryMethodDescriptor(name, input, output string) *descriptorpb.MethodDescriptorProto {
	return &descriptorpb.MethodDescriptorProto{
		Name:       descriptorString(name),
		InputType:  descriptorString(".l1.nativeaccount.v1." + input),
		OutputType: descriptorString(".l1.nativeaccount.v1." + output),
	}
}

func virtualAccountResponseDescriptor(name string, includeDebt bool) *descriptorpb.DescriptorProto {
	fields := []*descriptorpb.FieldDescriptorProto{
		descriptorField("address_user", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING),
		descriptorField("address_raw", 2, descriptorpb.FieldDescriptorProto_TYPE_STRING),
		descriptorField("status", 3, descriptorpb.FieldDescriptorProto_TYPE_STRING),
		descriptorField("persistent", 4, descriptorpb.FieldDescriptorProto_TYPE_BOOL),
		descriptorField("storage_rent_active", 5, descriptorpb.FieldDescriptorProto_TYPE_BOOL),
	}
	if includeDebt {
		fields = append(fields, descriptorField("storage_rent_debt", 6, descriptorpb.FieldDescriptorProto_TYPE_UINT64))
	}
	return &descriptorpb.DescriptorProto{Name: descriptorString(name), Field: fields}
}
