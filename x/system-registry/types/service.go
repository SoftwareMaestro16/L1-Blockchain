package types

import (
	"bytes"
	"compress/gzip"
	"context"

	"github.com/cosmos/gogoproto/grpc"
	gogoproto "github.com/cosmos/gogoproto/proto"
	grpcgo "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	proto2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/sovereign-l1/l1/app/addressing"
)

type MsgRegisterSystemEntityResponse struct {
	Entity	SystemEntity		`protobuf:"bytes,1,opt,name=entity,proto3" json:"entity"`
	Event	SystemEntityEvent	`protobuf:"bytes,2,opt,name=event,proto3" json:"event"`
}
type MsgUpdateSystemEntityResponse struct {
	Entity	SystemEntity		`protobuf:"bytes,1,opt,name=entity,proto3" json:"entity"`
	Event	SystemEntityEvent	`protobuf:"bytes,2,opt,name=event,proto3" json:"event"`
}
type MsgPauseSystemEntityResponse struct {
	Entity	SystemEntity		`protobuf:"bytes,1,opt,name=entity,proto3" json:"entity"`
	Event	SystemEntityEvent	`protobuf:"bytes,2,opt,name=event,proto3" json:"event"`
}
type MsgResumeSystemEntityResponse struct {
	Entity	SystemEntity		`protobuf:"bytes,1,opt,name=entity,proto3" json:"entity"`
	Event	SystemEntityEvent	`protobuf:"bytes,2,opt,name=event,proto3" json:"event"`
}
type MsgDeprecateSystemEntityResponse struct {
	Entity	SystemEntity		`protobuf:"bytes,1,opt,name=entity,proto3" json:"entity"`
	Event	SystemEntityEvent	`protobuf:"bytes,2,opt,name=event,proto3" json:"event"`
}

type QueryParamsRequest struct{}
type QueryParamsResponse struct {
	Params Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
}
type QuerySystemEntitiesRequest struct {
	Offset	uint64	`protobuf:"varint,1,opt,name=offset,proto3" json:"offset,omitempty"`
	Limit	uint64	`protobuf:"varint,2,opt,name=limit,proto3" json:"limit,omitempty"`
}
type QuerySystemEntitiesResponse struct {
	Entities	[]SystemEntity	`protobuf:"bytes,1,rep,name=entities,proto3" json:"entities,omitempty"`
	Next		uint64		`protobuf:"varint,2,opt,name=next,proto3" json:"next,omitempty"`
	Total		uint64		`protobuf:"varint,3,opt,name=total,proto3" json:"total,omitempty"`
}
type QuerySystemEntityRequest struct {
	ModuleName string `protobuf:"bytes,1,opt,name=module_name,json=moduleName,proto3" json:"module_name,omitempty"`
}
type QuerySystemEntityResponse struct {
	Entity	SystemEntity	`protobuf:"bytes,1,opt,name=entity,proto3" json:"entity"`
	Found	bool		`protobuf:"varint,2,opt,name=found,proto3" json:"found,omitempty"`
}
type QueryReservedSystemAddressesRequest struct {
	Offset	uint64	`protobuf:"varint,1,opt,name=offset,proto3" json:"offset,omitempty"`
	Limit	uint64	`protobuf:"varint,2,opt,name=limit,proto3" json:"limit,omitempty"`
}
type QueryReservedSystemAddressesResponse struct {
	Addresses	[]addressing.SystemAddress	`protobuf:"bytes,1,rep,name=addresses,proto3" json:"addresses,omitempty"`
	Next		uint64				`protobuf:"varint,2,opt,name=next,proto3" json:"next,omitempty"`
	Total		uint64				`protobuf:"varint,3,opt,name=total,proto3" json:"total,omitempty"`
}
type QuerySystemAddressRequest struct {
	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
}
type QuerySystemAddressResponse struct {
	Address	addressing.SystemAddress	`protobuf:"bytes,1,opt,name=address,proto3" json:"address"`
	Found	bool				`protobuf:"varint,2,opt,name=found,proto3" json:"found,omitempty"`
}
type QueryDependencyGraphRequest struct{}
type QueryDependencyGraphResponse struct {
	Edges []DependencyEdge `protobuf:"bytes,1,rep,name=edges,proto3" json:"edges,omitempty"`
}

type MsgServer interface {
	RegisterSystemEntity(context.Context, *MsgRegisterSystemEntity) (*MsgRegisterSystemEntityResponse, error)
	UpdateSystemEntity(context.Context, *MsgUpdateSystemEntity) (*MsgUpdateSystemEntityResponse, error)
	PauseSystemEntity(context.Context, *MsgPauseSystemEntity) (*MsgPauseSystemEntityResponse, error)
	ResumeSystemEntity(context.Context, *MsgResumeSystemEntity) (*MsgResumeSystemEntityResponse, error)
	DeprecateSystemEntity(context.Context, *MsgDeprecateSystemEntity) (*MsgDeprecateSystemEntityResponse, error)
}

type QueryServer interface {
	Params(context.Context, *QueryParamsRequest) (*QueryParamsResponse, error)
	SystemEntities(context.Context, *QuerySystemEntitiesRequest) (*QuerySystemEntitiesResponse, error)
	SystemEntity(context.Context, *QuerySystemEntityRequest) (*QuerySystemEntityResponse, error)
	ReservedSystemAddresses(context.Context, *QueryReservedSystemAddressesRequest) (*QueryReservedSystemAddressesResponse, error)
	SystemAddress(context.Context, *QuerySystemAddressRequest) (*QuerySystemAddressResponse, error)
	DependencyGraph(context.Context, *QueryDependencyGraphRequest) (*QueryDependencyGraphResponse, error)
}

type UnimplementedMsgServer struct{}

func (UnimplementedMsgServer) RegisterSystemEntity(context.Context, *MsgRegisterSystemEntity) (*MsgRegisterSystemEntityResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegisterSystemEntity not implemented")
}
func (UnimplementedMsgServer) UpdateSystemEntity(context.Context, *MsgUpdateSystemEntity) (*MsgUpdateSystemEntityResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateSystemEntity not implemented")
}
func (UnimplementedMsgServer) PauseSystemEntity(context.Context, *MsgPauseSystemEntity) (*MsgPauseSystemEntityResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PauseSystemEntity not implemented")
}
func (UnimplementedMsgServer) ResumeSystemEntity(context.Context, *MsgResumeSystemEntity) (*MsgResumeSystemEntityResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ResumeSystemEntity not implemented")
}
func (UnimplementedMsgServer) DeprecateSystemEntity(context.Context, *MsgDeprecateSystemEntity) (*MsgDeprecateSystemEntityResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeprecateSystemEntity not implemented")
}

func RegisterMsgServer(s grpc.Server, srv MsgServer)		{ s.RegisterService(&Msg_serviceDesc, srv) }
func RegisterQueryServer(s grpc.Server, srv QueryServer)	{ s.RegisterService(&Query_serviceDesc, srv) }

type serviceCall func(context.Context, interface{}, interface{}) (interface{}, error)

var Msg_serviceDesc = grpcgo.ServiceDesc{
	ServiceName:	"l1.systemregistry.v1.Msg",
	HandlerType:	(*MsgServer)(nil),
	Methods: []grpcgo.MethodDesc{
		methodDesc("RegisterSystemEntity", serviceHandler("RegisterSystemEntity", func() interface{} { return new(MsgRegisterSystemEntity) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(MsgServer).RegisterSystemEntity(ctx, req.(*MsgRegisterSystemEntity))
		})),
		methodDesc("UpdateSystemEntity", serviceHandler("UpdateSystemEntity", func() interface{} { return new(MsgUpdateSystemEntity) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(MsgServer).UpdateSystemEntity(ctx, req.(*MsgUpdateSystemEntity))
		})),
		methodDesc("PauseSystemEntity", serviceHandler("PauseSystemEntity", func() interface{} { return new(MsgPauseSystemEntity) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(MsgServer).PauseSystemEntity(ctx, req.(*MsgPauseSystemEntity))
		})),
		methodDesc("ResumeSystemEntity", serviceHandler("ResumeSystemEntity", func() interface{} { return new(MsgResumeSystemEntity) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(MsgServer).ResumeSystemEntity(ctx, req.(*MsgResumeSystemEntity))
		})),
		methodDesc("DeprecateSystemEntity", serviceHandler("DeprecateSystemEntity", func() interface{} { return new(MsgDeprecateSystemEntity) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(MsgServer).DeprecateSystemEntity(ctx, req.(*MsgDeprecateSystemEntity))
		})),
	},
	Metadata:	"l1/systemregistry/v1/tx.proto",
}

var Query_serviceDesc = grpcgo.ServiceDesc{
	ServiceName:	"l1.systemregistry.v1.Query",
	HandlerType:	(*QueryServer)(nil),
	Methods: []grpcgo.MethodDesc{
		methodDesc("Params", serviceHandler("Params", func() interface{} { return new(QueryParamsRequest) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(QueryServer).Params(ctx, req.(*QueryParamsRequest))
		})),
		methodDesc("SystemEntities", serviceHandler("SystemEntities", func() interface{} { return new(QuerySystemEntitiesRequest) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(QueryServer).SystemEntities(ctx, req.(*QuerySystemEntitiesRequest))
		})),
		methodDesc("SystemEntity", serviceHandler("SystemEntity", func() interface{} { return new(QuerySystemEntityRequest) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(QueryServer).SystemEntity(ctx, req.(*QuerySystemEntityRequest))
		})),
		methodDesc("ReservedSystemAddresses", serviceHandler("ReservedSystemAddresses", func() interface{} { return new(QueryReservedSystemAddressesRequest) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(QueryServer).ReservedSystemAddresses(ctx, req.(*QueryReservedSystemAddressesRequest))
		})),
		methodDesc("SystemAddress", serviceHandler("SystemAddress", func() interface{} { return new(QuerySystemAddressRequest) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(QueryServer).SystemAddress(ctx, req.(*QuerySystemAddressRequest))
		})),
		methodDesc("DependencyGraph", serviceHandler("DependencyGraph", func() interface{} { return new(QueryDependencyGraphRequest) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(QueryServer).DependencyGraph(ctx, req.(*QueryDependencyGraphRequest))
		})),
	},
	Metadata:	"l1/systemregistry/v1/query.proto",
}

func methodDesc(name string, handler grpcgo.MethodHandler) grpcgo.MethodDesc {
	return grpcgo.MethodDesc{MethodName: name, Handler: handler}
}

func serviceHandler(method string, newReq func() interface{}, call serviceCall) grpcgo.MethodHandler {
	return func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
		req := newReq()
		if err := dec(req); err != nil {
			return nil, err
		}
		if interceptor == nil {
			return call(ctx, srv, req)
		}
		return interceptor(ctx, req, &grpcgo.UnaryServerInfo{Server: srv, FullMethod: method}, func(ctx context.Context, request interface{}) (interface{}, error) {
			return call(ctx, srv, request)
		})
	}
}

func init() {
	registerServiceTypes()
	gogoproto.RegisterFile("l1/systemregistry/v1/tx.proto", buildServiceFileDescriptor("l1/systemregistry/v1/tx.proto", "l1.systemregistry.v1", "Msg", []string{"MsgRegisterSystemEntity", "MsgRegisterSystemEntityResponse", "MsgUpdateSystemEntity", "MsgUpdateSystemEntityResponse", "MsgPauseSystemEntity", "MsgPauseSystemEntityResponse", "MsgResumeSystemEntity", "MsgResumeSystemEntityResponse", "MsgDeprecateSystemEntity", "MsgDeprecateSystemEntityResponse"}, [][3]string{{"RegisterSystemEntity", "MsgRegisterSystemEntity", "MsgRegisterSystemEntityResponse"}, {"UpdateSystemEntity", "MsgUpdateSystemEntity", "MsgUpdateSystemEntityResponse"}, {"PauseSystemEntity", "MsgPauseSystemEntity", "MsgPauseSystemEntityResponse"}, {"ResumeSystemEntity", "MsgResumeSystemEntity", "MsgResumeSystemEntityResponse"}, {"DeprecateSystemEntity", "MsgDeprecateSystemEntity", "MsgDeprecateSystemEntityResponse"}}))
	gogoproto.RegisterFile("l1/systemregistry/v1/query.proto", buildServiceFileDescriptor("l1/systemregistry/v1/query.proto", "l1.systemregistry.v1", "Query", []string{"QueryParamsRequest", "QueryParamsResponse", "QuerySystemEntitiesRequest", "QuerySystemEntitiesResponse", "QuerySystemEntityRequest", "QuerySystemEntityResponse", "QueryReservedSystemAddressesRequest", "QueryReservedSystemAddressesResponse", "QuerySystemAddressRequest", "QuerySystemAddressResponse", "QueryDependencyGraphRequest", "QueryDependencyGraphResponse"}, [][3]string{{"Params", "QueryParamsRequest", "QueryParamsResponse"}, {"SystemEntities", "QuerySystemEntitiesRequest", "QuerySystemEntitiesResponse"}, {"SystemEntity", "QuerySystemEntityRequest", "QuerySystemEntityResponse"}, {"ReservedSystemAddresses", "QueryReservedSystemAddressesRequest", "QueryReservedSystemAddressesResponse"}, {"SystemAddress", "QuerySystemAddressRequest", "QuerySystemAddressResponse"}, {"DependencyGraph", "QueryDependencyGraphRequest", "QueryDependencyGraphResponse"}}))
}

func buildServiceFileDescriptor(path, pkg, service string, messageNames []string, methods [][3]string) []byte {
	messages := make([]*descriptorpb.DescriptorProto, 0, len(messageNames))
	for _, name := range messageNames {
		messages = append(messages, &descriptorpb.DescriptorProto{Name: stringPtr(name)})
	}
	md := make([]*descriptorpb.MethodDescriptorProto, 0, len(methods))
	for _, method := range methods {
		md = append(md, &descriptorpb.MethodDescriptorProto{Name: stringPtr(method[0]), InputType: stringPtr("." + pkg + "." + method[1]), OutputType: stringPtr("." + pkg + "." + method[2])})
	}
	fd := &descriptorpb.FileDescriptorProto{Name: stringPtr(path), Package: stringPtr(pkg), Syntax: stringPtr("proto3"), MessageType: messages, Service: []*descriptorpb.ServiceDescriptorProto{{Name: stringPtr(service), Method: md}}}
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

func stringPtr(value string) *string	{ return &value }

func registerServiceTypes() {
	for _, item := range []struct {
		msg	gogoproto.Message
		name	string
	}{
		{(*MsgRegisterSystemEntity)(nil), "l1.systemregistry.v1.MsgRegisterSystemEntity"},
		{(*MsgRegisterSystemEntityResponse)(nil), "l1.systemregistry.v1.MsgRegisterSystemEntityResponse"},
		{(*MsgUpdateSystemEntity)(nil), "l1.systemregistry.v1.MsgUpdateSystemEntity"},
		{(*MsgUpdateSystemEntityResponse)(nil), "l1.systemregistry.v1.MsgUpdateSystemEntityResponse"},
		{(*MsgPauseSystemEntity)(nil), "l1.systemregistry.v1.MsgPauseSystemEntity"},
		{(*MsgPauseSystemEntityResponse)(nil), "l1.systemregistry.v1.MsgPauseSystemEntityResponse"},
		{(*MsgResumeSystemEntity)(nil), "l1.systemregistry.v1.MsgResumeSystemEntity"},
		{(*MsgResumeSystemEntityResponse)(nil), "l1.systemregistry.v1.MsgResumeSystemEntityResponse"},
		{(*MsgDeprecateSystemEntity)(nil), "l1.systemregistry.v1.MsgDeprecateSystemEntity"},
		{(*MsgDeprecateSystemEntityResponse)(nil), "l1.systemregistry.v1.MsgDeprecateSystemEntityResponse"},
		{(*QueryParamsRequest)(nil), "l1.systemregistry.v1.QueryParamsRequest"},
		{(*QueryParamsResponse)(nil), "l1.systemregistry.v1.QueryParamsResponse"},
		{(*QuerySystemEntitiesRequest)(nil), "l1.systemregistry.v1.QuerySystemEntitiesRequest"},
		{(*QuerySystemEntitiesResponse)(nil), "l1.systemregistry.v1.QuerySystemEntitiesResponse"},
		{(*QuerySystemEntityRequest)(nil), "l1.systemregistry.v1.QuerySystemEntityRequest"},
		{(*QuerySystemEntityResponse)(nil), "l1.systemregistry.v1.QuerySystemEntityResponse"},
		{(*QueryReservedSystemAddressesRequest)(nil), "l1.systemregistry.v1.QueryReservedSystemAddressesRequest"},
		{(*QueryReservedSystemAddressesResponse)(nil), "l1.systemregistry.v1.QueryReservedSystemAddressesResponse"},
		{(*QuerySystemAddressRequest)(nil), "l1.systemregistry.v1.QuerySystemAddressRequest"},
		{(*QuerySystemAddressResponse)(nil), "l1.systemregistry.v1.QuerySystemAddressResponse"},
		{(*QueryDependencyGraphRequest)(nil), "l1.systemregistry.v1.QueryDependencyGraphRequest"},
		{(*QueryDependencyGraphResponse)(nil), "l1.systemregistry.v1.QueryDependencyGraphResponse"},
	} {
		gogoproto.RegisterType(item.msg, item.name)
	}
}

func (m *MsgRegisterSystemEntity) Reset()		{ *m = MsgRegisterSystemEntity{} }
func (m *MsgUpdateSystemEntity) Reset()			{ *m = MsgUpdateSystemEntity{} }
func (m *MsgPauseSystemEntity) Reset()			{ *m = MsgPauseSystemEntity{} }
func (m *MsgResumeSystemEntity) Reset()			{ *m = MsgResumeSystemEntity{} }
func (m *MsgDeprecateSystemEntity) Reset()		{ *m = MsgDeprecateSystemEntity{} }
func (m *MsgRegisterSystemEntityResponse) Reset()	{ *m = MsgRegisterSystemEntityResponse{} }
func (m *MsgUpdateSystemEntityResponse) Reset()		{ *m = MsgUpdateSystemEntityResponse{} }
func (m *MsgPauseSystemEntityResponse) Reset()		{ *m = MsgPauseSystemEntityResponse{} }
func (m *MsgResumeSystemEntityResponse) Reset()		{ *m = MsgResumeSystemEntityResponse{} }
func (m *MsgDeprecateSystemEntityResponse) Reset()	{ *m = MsgDeprecateSystemEntityResponse{} }
func (m *QueryParamsRequest) Reset()			{ *m = QueryParamsRequest{} }
func (m *QueryParamsResponse) Reset()			{ *m = QueryParamsResponse{} }
func (m *QuerySystemEntitiesRequest) Reset()		{ *m = QuerySystemEntitiesRequest{} }
func (m *QuerySystemEntitiesResponse) Reset()		{ *m = QuerySystemEntitiesResponse{} }
func (m *QuerySystemEntityRequest) Reset()		{ *m = QuerySystemEntityRequest{} }
func (m *QuerySystemEntityResponse) Reset()		{ *m = QuerySystemEntityResponse{} }
func (m *QueryReservedSystemAddressesRequest) Reset()	{ *m = QueryReservedSystemAddressesRequest{} }
func (m *QueryReservedSystemAddressesResponse) Reset() {
	*m = QueryReservedSystemAddressesResponse{}
}
func (m *QuerySystemAddressRequest) Reset()			{ *m = QuerySystemAddressRequest{} }
func (m *QuerySystemAddressResponse) Reset()			{ *m = QuerySystemAddressResponse{} }
func (m *QueryDependencyGraphRequest) Reset()			{ *m = QueryDependencyGraphRequest{} }
func (m *QueryDependencyGraphResponse) Reset()			{ *m = QueryDependencyGraphResponse{} }
func (m *MsgRegisterSystemEntity) String() string		{ return gogoproto.CompactTextString(m) }
func (m *MsgUpdateSystemEntity) String() string			{ return gogoproto.CompactTextString(m) }
func (m *MsgPauseSystemEntity) String() string			{ return gogoproto.CompactTextString(m) }
func (m *MsgResumeSystemEntity) String() string			{ return gogoproto.CompactTextString(m) }
func (m *MsgDeprecateSystemEntity) String() string		{ return gogoproto.CompactTextString(m) }
func (m *MsgRegisterSystemEntityResponse) String() string	{ return gogoproto.CompactTextString(m) }
func (m *MsgUpdateSystemEntityResponse) String() string		{ return gogoproto.CompactTextString(m) }
func (m *MsgPauseSystemEntityResponse) String() string		{ return gogoproto.CompactTextString(m) }
func (m *MsgResumeSystemEntityResponse) String() string		{ return gogoproto.CompactTextString(m) }
func (m *MsgDeprecateSystemEntityResponse) String() string	{ return gogoproto.CompactTextString(m) }
func (m *QueryParamsRequest) String() string			{ return gogoproto.CompactTextString(m) }
func (m *QueryParamsResponse) String() string			{ return gogoproto.CompactTextString(m) }
func (m *QuerySystemEntitiesRequest) String() string		{ return gogoproto.CompactTextString(m) }
func (m *QuerySystemEntitiesResponse) String() string		{ return gogoproto.CompactTextString(m) }
func (m *QuerySystemEntityRequest) String() string		{ return gogoproto.CompactTextString(m) }
func (m *QuerySystemEntityResponse) String() string		{ return gogoproto.CompactTextString(m) }
func (m *QueryReservedSystemAddressesRequest) String() string	{ return gogoproto.CompactTextString(m) }
func (m *QueryReservedSystemAddressesResponse) String() string {
	return gogoproto.CompactTextString(m)
}
func (m *QuerySystemAddressRequest) String() string	{ return gogoproto.CompactTextString(m) }
func (m *QuerySystemAddressResponse) String() string	{ return gogoproto.CompactTextString(m) }
func (m *QueryDependencyGraphRequest) String() string	{ return gogoproto.CompactTextString(m) }
func (m *QueryDependencyGraphResponse) String() string	{ return gogoproto.CompactTextString(m) }

func (*MsgRegisterSystemEntity) ProtoMessage()			{}
func (*MsgUpdateSystemEntity) ProtoMessage()			{}
func (*MsgPauseSystemEntity) ProtoMessage()			{}
func (*MsgResumeSystemEntity) ProtoMessage()			{}
func (*MsgDeprecateSystemEntity) ProtoMessage()			{}
func (*MsgRegisterSystemEntityResponse) ProtoMessage()		{}
func (*MsgUpdateSystemEntityResponse) ProtoMessage()		{}
func (*MsgPauseSystemEntityResponse) ProtoMessage()		{}
func (*MsgResumeSystemEntityResponse) ProtoMessage()		{}
func (*MsgDeprecateSystemEntityResponse) ProtoMessage()		{}
func (*QueryParamsRequest) ProtoMessage()			{}
func (*QueryParamsResponse) ProtoMessage()			{}
func (*QuerySystemEntitiesRequest) ProtoMessage()		{}
func (*QuerySystemEntitiesResponse) ProtoMessage()		{}
func (*QuerySystemEntityRequest) ProtoMessage()			{}
func (*QuerySystemEntityResponse) ProtoMessage()		{}
func (*QueryReservedSystemAddressesRequest) ProtoMessage()	{}
func (*QueryReservedSystemAddressesResponse) ProtoMessage()	{}
func (*QuerySystemAddressRequest) ProtoMessage()		{}
func (*QuerySystemAddressResponse) ProtoMessage()		{}
func (*QueryDependencyGraphRequest) ProtoMessage()		{}
func (*QueryDependencyGraphResponse) ProtoMessage()		{}
