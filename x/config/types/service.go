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
)

type MsgSubmitConfigChangeResponse struct {
	Change ConfigChange `protobuf:"bytes,1,opt,name=change,proto3" json:"change"`
}
type MsgApproveConfigChangeResponse struct {
	Change ConfigChange `protobuf:"bytes,1,opt,name=change,proto3" json:"change"`
}
type MsgRejectConfigChangeResponse struct {
	Change ConfigChange `protobuf:"bytes,1,opt,name=change,proto3" json:"change"`
}
type MsgExecuteConfigChangeResponse struct {
	Entry	ConfigEntry	`protobuf:"bytes,1,opt,name=entry,proto3" json:"entry"`
	Change	ConfigChange	`protobuf:"bytes,2,opt,name=change,proto3" json:"change"`
}
type MsgCancelConfigChangeResponse struct {
	Change ConfigChange `protobuf:"bytes,1,opt,name=change,proto3" json:"change"`
}

type QueryParamsRequest struct{}
type QueryParamsResponse struct {
	Params Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
}
type QueryEntriesRequest struct {
	Offset	uint64	`protobuf:"varint,1,opt,name=offset,proto3" json:"offset,omitempty"`
	Limit	uint64	`protobuf:"varint,2,opt,name=limit,proto3" json:"limit,omitempty"`
}
type QueryEntriesResponse struct {
	Entries	[]ConfigEntry	`protobuf:"bytes,1,rep,name=entries,proto3" json:"entries,omitempty"`
	Next	uint64		`protobuf:"varint,2,opt,name=next,proto3" json:"next,omitempty"`
	Total	uint64		`protobuf:"varint,3,opt,name=total,proto3" json:"total,omitempty"`
}
type QueryEntryRequest struct {
	Key string `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
}
type QueryEntryResponse struct {
	Entry	ConfigEntry	`protobuf:"bytes,1,opt,name=entry,proto3" json:"entry"`
	Found	bool		`protobuf:"varint,2,opt,name=found,proto3" json:"found,omitempty"`
}
type QueryPendingChangesRequest struct {
	Offset	uint64	`protobuf:"varint,1,opt,name=offset,proto3" json:"offset,omitempty"`
	Limit	uint64	`protobuf:"varint,2,opt,name=limit,proto3" json:"limit,omitempty"`
}
type QueryPendingChangesResponse struct {
	Changes	[]ConfigChange	`protobuf:"bytes,1,rep,name=changes,proto3" json:"changes,omitempty"`
	Next	uint64		`protobuf:"varint,2,opt,name=next,proto3" json:"next,omitempty"`
	Total	uint64		`protobuf:"varint,3,opt,name=total,proto3" json:"total,omitempty"`
}
type QueryChangeRequest struct {
	ID string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}
type QueryChangeResponse struct {
	Change	ConfigChange	`protobuf:"bytes,1,opt,name=change,proto3" json:"change"`
	Found	bool		`protobuf:"varint,2,opt,name=found,proto3" json:"found,omitempty"`
}
type QueryEffectiveParamsRequest struct {
	Height	int64	`protobuf:"varint,1,opt,name=height,proto3" json:"height,omitempty"`
	Offset	uint64	`protobuf:"varint,2,opt,name=offset,proto3" json:"offset,omitempty"`
	Limit	uint64	`protobuf:"varint,3,opt,name=limit,proto3" json:"limit,omitempty"`
}
type QueryEffectiveParamsResponse struct {
	Entries	[]ConfigEntry	`protobuf:"bytes,1,rep,name=entries,proto3" json:"entries,omitempty"`
	Next	uint64		`protobuf:"varint,2,opt,name=next,proto3" json:"next,omitempty"`
	Total	uint64		`protobuf:"varint,3,opt,name=total,proto3" json:"total,omitempty"`
}

type MsgServer interface {
	SubmitConfigChange(context.Context, *MsgSubmitConfigChange) (*MsgSubmitConfigChangeResponse, error)
	ApproveConfigChange(context.Context, *MsgApproveConfigChange) (*MsgApproveConfigChangeResponse, error)
	RejectConfigChange(context.Context, *MsgRejectConfigChange) (*MsgRejectConfigChangeResponse, error)
	ExecuteConfigChange(context.Context, *MsgExecuteConfigChange) (*MsgExecuteConfigChangeResponse, error)
	CancelConfigChange(context.Context, *MsgCancelConfigChange) (*MsgCancelConfigChangeResponse, error)
}

type QueryServer interface {
	Params(context.Context, *QueryParamsRequest) (*QueryParamsResponse, error)
	Entries(context.Context, *QueryEntriesRequest) (*QueryEntriesResponse, error)
	Entry(context.Context, *QueryEntryRequest) (*QueryEntryResponse, error)
	PendingChanges(context.Context, *QueryPendingChangesRequest) (*QueryPendingChangesResponse, error)
	Change(context.Context, *QueryChangeRequest) (*QueryChangeResponse, error)
	EffectiveParams(context.Context, *QueryEffectiveParamsRequest) (*QueryEffectiveParamsResponse, error)
}

type UnimplementedMsgServer struct{}
type UnimplementedQueryServer struct{}

func (UnimplementedMsgServer) SubmitConfigChange(context.Context, *MsgSubmitConfigChange) (*MsgSubmitConfigChangeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitConfigChange not implemented")
}
func (UnimplementedMsgServer) ApproveConfigChange(context.Context, *MsgApproveConfigChange) (*MsgApproveConfigChangeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ApproveConfigChange not implemented")
}
func (UnimplementedMsgServer) RejectConfigChange(context.Context, *MsgRejectConfigChange) (*MsgRejectConfigChangeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RejectConfigChange not implemented")
}
func (UnimplementedMsgServer) ExecuteConfigChange(context.Context, *MsgExecuteConfigChange) (*MsgExecuteConfigChangeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ExecuteConfigChange not implemented")
}
func (UnimplementedMsgServer) CancelConfigChange(context.Context, *MsgCancelConfigChange) (*MsgCancelConfigChangeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CancelConfigChange not implemented")
}

func RegisterMsgServer(s grpc.Server, srv MsgServer)		{ s.RegisterService(&Msg_serviceDesc, srv) }
func RegisterQueryServer(s grpc.Server, srv QueryServer)	{ s.RegisterService(&Query_serviceDesc, srv) }

type serviceCall func(context.Context, interface{}, interface{}) (interface{}, error)

var Msg_serviceDesc = grpcgo.ServiceDesc{
	ServiceName:	"l1.config.v1.Msg",
	HandlerType:	(*MsgServer)(nil),
	Methods: []grpcgo.MethodDesc{
		methodDesc("SubmitConfigChange", serviceHandler("SubmitConfigChange", func() interface{} { return new(MsgSubmitConfigChange) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(MsgServer).SubmitConfigChange(ctx, req.(*MsgSubmitConfigChange))
		})),
		methodDesc("ApproveConfigChange", serviceHandler("ApproveConfigChange", func() interface{} { return new(MsgApproveConfigChange) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(MsgServer).ApproveConfigChange(ctx, req.(*MsgApproveConfigChange))
		})),
		methodDesc("RejectConfigChange", serviceHandler("RejectConfigChange", func() interface{} { return new(MsgRejectConfigChange) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(MsgServer).RejectConfigChange(ctx, req.(*MsgRejectConfigChange))
		})),
		methodDesc("ExecuteConfigChange", serviceHandler("ExecuteConfigChange", func() interface{} { return new(MsgExecuteConfigChange) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(MsgServer).ExecuteConfigChange(ctx, req.(*MsgExecuteConfigChange))
		})),
		methodDesc("CancelConfigChange", serviceHandler("CancelConfigChange", func() interface{} { return new(MsgCancelConfigChange) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(MsgServer).CancelConfigChange(ctx, req.(*MsgCancelConfigChange))
		})),
	},
	Metadata:	"l1/config/v1/tx.proto",
}

var Query_serviceDesc = grpcgo.ServiceDesc{
	ServiceName:	"l1.config.v1.Query",
	HandlerType:	(*QueryServer)(nil),
	Methods: []grpcgo.MethodDesc{
		methodDesc("Params", serviceHandler("Params", func() interface{} { return new(QueryParamsRequest) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(QueryServer).Params(ctx, req.(*QueryParamsRequest))
		})),
		methodDesc("Entries", serviceHandler("Entries", func() interface{} { return new(QueryEntriesRequest) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(QueryServer).Entries(ctx, req.(*QueryEntriesRequest))
		})),
		methodDesc("Entry", serviceHandler("Entry", func() interface{} { return new(QueryEntryRequest) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(QueryServer).Entry(ctx, req.(*QueryEntryRequest))
		})),
		methodDesc("PendingChanges", serviceHandler("PendingChanges", func() interface{} { return new(QueryPendingChangesRequest) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(QueryServer).PendingChanges(ctx, req.(*QueryPendingChangesRequest))
		})),
		methodDesc("Change", serviceHandler("Change", func() interface{} { return new(QueryChangeRequest) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(QueryServer).Change(ctx, req.(*QueryChangeRequest))
		})),
		methodDesc("EffectiveParams", serviceHandler("EffectiveParams", func() interface{} { return new(QueryEffectiveParamsRequest) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(QueryServer).EffectiveParams(ctx, req.(*QueryEffectiveParamsRequest))
		})),
	},
	Metadata:	"l1/config/v1/query.proto",
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
	gogoproto.RegisterFile("l1/config/v1/tx.proto", buildServiceFileDescriptor("l1/config/v1/tx.proto", "l1.config.v1", "Msg", []string{"MsgSubmitConfigChange", "MsgSubmitConfigChangeResponse", "MsgApproveConfigChange", "MsgApproveConfigChangeResponse", "MsgRejectConfigChange", "MsgRejectConfigChangeResponse", "MsgExecuteConfigChange", "MsgExecuteConfigChangeResponse", "MsgCancelConfigChange", "MsgCancelConfigChangeResponse"}, [][3]string{{"SubmitConfigChange", "MsgSubmitConfigChange", "MsgSubmitConfigChangeResponse"}, {"ApproveConfigChange", "MsgApproveConfigChange", "MsgApproveConfigChangeResponse"}, {"RejectConfigChange", "MsgRejectConfigChange", "MsgRejectConfigChangeResponse"}, {"ExecuteConfigChange", "MsgExecuteConfigChange", "MsgExecuteConfigChangeResponse"}, {"CancelConfigChange", "MsgCancelConfigChange", "MsgCancelConfigChangeResponse"}}))
	gogoproto.RegisterFile("l1/config/v1/query.proto", buildServiceFileDescriptor("l1/config/v1/query.proto", "l1.config.v1", "Query", []string{"QueryParamsRequest", "QueryParamsResponse", "QueryEntriesRequest", "QueryEntriesResponse", "QueryEntryRequest", "QueryEntryResponse", "QueryPendingChangesRequest", "QueryPendingChangesResponse", "QueryChangeRequest", "QueryChangeResponse", "QueryEffectiveParamsRequest", "QueryEffectiveParamsResponse"}, [][3]string{{"Params", "QueryParamsRequest", "QueryParamsResponse"}, {"Entries", "QueryEntriesRequest", "QueryEntriesResponse"}, {"Entry", "QueryEntryRequest", "QueryEntryResponse"}, {"PendingChanges", "QueryPendingChangesRequest", "QueryPendingChangesResponse"}, {"Change", "QueryChangeRequest", "QueryChangeResponse"}, {"EffectiveParams", "QueryEffectiveParamsRequest", "QueryEffectiveParamsResponse"}}))
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
		{(*MsgSubmitConfigChange)(nil), "l1.config.v1.MsgSubmitConfigChange"},
		{(*MsgSubmitConfigChangeResponse)(nil), "l1.config.v1.MsgSubmitConfigChangeResponse"},
		{(*MsgApproveConfigChange)(nil), "l1.config.v1.MsgApproveConfigChange"},
		{(*MsgApproveConfigChangeResponse)(nil), "l1.config.v1.MsgApproveConfigChangeResponse"},
		{(*MsgRejectConfigChange)(nil), "l1.config.v1.MsgRejectConfigChange"},
		{(*MsgRejectConfigChangeResponse)(nil), "l1.config.v1.MsgRejectConfigChangeResponse"},
		{(*MsgExecuteConfigChange)(nil), "l1.config.v1.MsgExecuteConfigChange"},
		{(*MsgExecuteConfigChangeResponse)(nil), "l1.config.v1.MsgExecuteConfigChangeResponse"},
		{(*MsgCancelConfigChange)(nil), "l1.config.v1.MsgCancelConfigChange"},
		{(*MsgCancelConfigChangeResponse)(nil), "l1.config.v1.MsgCancelConfigChangeResponse"},
		{(*QueryParamsRequest)(nil), "l1.config.v1.QueryParamsRequest"},
		{(*QueryParamsResponse)(nil), "l1.config.v1.QueryParamsResponse"},
		{(*QueryEntriesRequest)(nil), "l1.config.v1.QueryEntriesRequest"},
		{(*QueryEntriesResponse)(nil), "l1.config.v1.QueryEntriesResponse"},
		{(*QueryEntryRequest)(nil), "l1.config.v1.QueryEntryRequest"},
		{(*QueryEntryResponse)(nil), "l1.config.v1.QueryEntryResponse"},
		{(*QueryPendingChangesRequest)(nil), "l1.config.v1.QueryPendingChangesRequest"},
		{(*QueryPendingChangesResponse)(nil), "l1.config.v1.QueryPendingChangesResponse"},
		{(*QueryChangeRequest)(nil), "l1.config.v1.QueryChangeRequest"},
		{(*QueryChangeResponse)(nil), "l1.config.v1.QueryChangeResponse"},
		{(*QueryEffectiveParamsRequest)(nil), "l1.config.v1.QueryEffectiveParamsRequest"},
		{(*QueryEffectiveParamsResponse)(nil), "l1.config.v1.QueryEffectiveParamsResponse"},
	} {
		gogoproto.RegisterType(item.msg, item.name)
	}
}

func (m *MsgSubmitConfigChange) Reset()			{ *m = MsgSubmitConfigChange{} }
func (m *MsgApproveConfigChange) Reset()		{ *m = MsgApproveConfigChange{} }
func (m *MsgRejectConfigChange) Reset()			{ *m = MsgRejectConfigChange{} }
func (m *MsgExecuteConfigChange) Reset()		{ *m = MsgExecuteConfigChange{} }
func (m *MsgCancelConfigChange) Reset()			{ *m = MsgCancelConfigChange{} }
func (m *MsgSubmitConfigChangeResponse) Reset()		{ *m = MsgSubmitConfigChangeResponse{} }
func (m *MsgApproveConfigChangeResponse) Reset()	{ *m = MsgApproveConfigChangeResponse{} }
func (m *MsgRejectConfigChangeResponse) Reset()		{ *m = MsgRejectConfigChangeResponse{} }
func (m *MsgExecuteConfigChangeResponse) Reset()	{ *m = MsgExecuteConfigChangeResponse{} }
func (m *MsgCancelConfigChangeResponse) Reset()		{ *m = MsgCancelConfigChangeResponse{} }
func (m *QueryParamsRequest) Reset()			{ *m = QueryParamsRequest{} }
func (m *QueryParamsResponse) Reset()			{ *m = QueryParamsResponse{} }
func (m *QueryEntriesRequest) Reset()			{ *m = QueryEntriesRequest{} }
func (m *QueryEntriesResponse) Reset()			{ *m = QueryEntriesResponse{} }
func (m *QueryEntryRequest) Reset()			{ *m = QueryEntryRequest{} }
func (m *QueryEntryResponse) Reset()			{ *m = QueryEntryResponse{} }
func (m *QueryPendingChangesRequest) Reset()		{ *m = QueryPendingChangesRequest{} }
func (m *QueryPendingChangesResponse) Reset()		{ *m = QueryPendingChangesResponse{} }
func (m *QueryChangeRequest) Reset()			{ *m = QueryChangeRequest{} }
func (m *QueryChangeResponse) Reset()			{ *m = QueryChangeResponse{} }
func (m *QueryEffectiveParamsRequest) Reset()		{ *m = QueryEffectiveParamsRequest{} }
func (m *QueryEffectiveParamsResponse) Reset()		{ *m = QueryEffectiveParamsResponse{} }
func (m *MsgSubmitConfigChange) String() string		{ return gogoproto.CompactTextString(m) }
func (m *MsgApproveConfigChange) String() string	{ return gogoproto.CompactTextString(m) }
func (m *MsgRejectConfigChange) String() string		{ return gogoproto.CompactTextString(m) }
func (m *MsgExecuteConfigChange) String() string	{ return gogoproto.CompactTextString(m) }
func (m *MsgCancelConfigChange) String() string		{ return gogoproto.CompactTextString(m) }
func (m *MsgSubmitConfigChangeResponse) String() string {
	return gogoproto.CompactTextString(m)
}
func (m *MsgApproveConfigChangeResponse) String() string {
	return gogoproto.CompactTextString(m)
}
func (m *MsgRejectConfigChangeResponse) String() string {
	return gogoproto.CompactTextString(m)
}
func (m *MsgExecuteConfigChangeResponse) String() string {
	return gogoproto.CompactTextString(m)
}
func (m *MsgCancelConfigChangeResponse) String() string {
	return gogoproto.CompactTextString(m)
}
func (m *QueryParamsRequest) String() string		{ return gogoproto.CompactTextString(m) }
func (m *QueryParamsResponse) String() string		{ return gogoproto.CompactTextString(m) }
func (m *QueryEntriesRequest) String() string		{ return gogoproto.CompactTextString(m) }
func (m *QueryEntriesResponse) String() string		{ return gogoproto.CompactTextString(m) }
func (m *QueryEntryRequest) String() string		{ return gogoproto.CompactTextString(m) }
func (m *QueryEntryResponse) String() string		{ return gogoproto.CompactTextString(m) }
func (m *QueryPendingChangesRequest) String() string	{ return gogoproto.CompactTextString(m) }
func (m *QueryPendingChangesResponse) String() string	{ return gogoproto.CompactTextString(m) }
func (m *QueryChangeRequest) String() string		{ return gogoproto.CompactTextString(m) }
func (m *QueryChangeResponse) String() string		{ return gogoproto.CompactTextString(m) }
func (m *QueryEffectiveParamsRequest) String() string	{ return gogoproto.CompactTextString(m) }
func (m *QueryEffectiveParamsResponse) String() string	{ return gogoproto.CompactTextString(m) }

func (*MsgSubmitConfigChange) ProtoMessage()		{}
func (*MsgApproveConfigChange) ProtoMessage()		{}
func (*MsgRejectConfigChange) ProtoMessage()		{}
func (*MsgExecuteConfigChange) ProtoMessage()		{}
func (*MsgCancelConfigChange) ProtoMessage()		{}
func (*MsgSubmitConfigChangeResponse) ProtoMessage()	{}
func (*MsgApproveConfigChangeResponse) ProtoMessage()	{}
func (*MsgRejectConfigChangeResponse) ProtoMessage()	{}
func (*MsgExecuteConfigChangeResponse) ProtoMessage()	{}
func (*MsgCancelConfigChangeResponse) ProtoMessage()	{}
func (*QueryParamsRequest) ProtoMessage()		{}
func (*QueryParamsResponse) ProtoMessage()		{}
func (*QueryEntriesRequest) ProtoMessage()		{}
func (*QueryEntriesResponse) ProtoMessage()		{}
func (*QueryEntryRequest) ProtoMessage()		{}
func (*QueryEntryResponse) ProtoMessage()		{}
func (*QueryPendingChangesRequest) ProtoMessage()	{}
func (*QueryPendingChangesResponse) ProtoMessage()	{}
func (*QueryChangeRequest) ProtoMessage()		{}
func (*QueryChangeResponse) ProtoMessage()		{}
func (*QueryEffectiveParamsRequest) ProtoMessage()	{}
func (*QueryEffectiveParamsResponse) ProtoMessage()	{}
