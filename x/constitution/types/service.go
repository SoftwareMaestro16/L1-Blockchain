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

type MsgProposeConstitutionAmendmentResponse struct {
	Amendment Amendment `protobuf:"bytes,1,opt,name=amendment,proto3" json:"amendment"`
}
type MsgVoteConstitutionAmendmentResponse struct {
	Amendment Amendment `protobuf:"bytes,1,opt,name=amendment,proto3" json:"amendment"`
}
type MsgExecuteConstitutionAmendmentResponse struct {
	Constitution	Constitution	`protobuf:"bytes,1,opt,name=constitution,proto3" json:"constitution"`
	Amendment	Amendment	`protobuf:"bytes,2,opt,name=amendment,proto3" json:"amendment"`
}
type MsgCancelConstitutionAmendmentResponse struct {
	Amendment Amendment `protobuf:"bytes,1,opt,name=amendment,proto3" json:"amendment"`
}

type QueryParamsRequest struct{}
type QueryParamsResponse struct {
	Params Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
}
type QueryConstitutionRequest struct{}
type QueryConstitutionResponse struct {
	Constitution Constitution `protobuf:"bytes,1,opt,name=constitution,proto3" json:"constitution"`
}
type QueryPendingAmendmentsRequest struct {
	Offset	uint64	`protobuf:"varint,1,opt,name=offset,proto3" json:"offset,omitempty"`
	Limit	uint64	`protobuf:"varint,2,opt,name=limit,proto3" json:"limit,omitempty"`
}
type QueryPendingAmendmentsResponse struct {
	Amendments	[]Amendment	`protobuf:"bytes,1,rep,name=amendments,proto3" json:"amendments,omitempty"`
	Next		uint64		`protobuf:"varint,2,opt,name=next,proto3" json:"next,omitempty"`
	Total		uint64		`protobuf:"varint,3,opt,name=total,proto3" json:"total,omitempty"`
}
type QueryAmendmentRequest struct {
	ID string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}
type QueryAmendmentResponse struct {
	Amendment	Amendment	`protobuf:"bytes,1,opt,name=amendment,proto3" json:"amendment"`
	Found		bool		`protobuf:"varint,2,opt,name=found,proto3" json:"found,omitempty"`
}
type QueryProtectedLimitsRequest struct{}
type QueryProtectedLimitsResponse struct {
	Limits ProtectedLimits `protobuf:"bytes,1,opt,name=limits,proto3" json:"limits"`
}

type MsgServer interface {
	ProposeConstitutionAmendment(context.Context, *MsgProposeConstitutionAmendment) (*MsgProposeConstitutionAmendmentResponse, error)
	VoteConstitutionAmendment(context.Context, *MsgVoteConstitutionAmendment) (*MsgVoteConstitutionAmendmentResponse, error)
	ExecuteConstitutionAmendment(context.Context, *MsgExecuteConstitutionAmendment) (*MsgExecuteConstitutionAmendmentResponse, error)
	CancelConstitutionAmendment(context.Context, *MsgCancelConstitutionAmendment) (*MsgCancelConstitutionAmendmentResponse, error)
}

type QueryServer interface {
	Params(context.Context, *QueryParamsRequest) (*QueryParamsResponse, error)
	Constitution(context.Context, *QueryConstitutionRequest) (*QueryConstitutionResponse, error)
	PendingAmendments(context.Context, *QueryPendingAmendmentsRequest) (*QueryPendingAmendmentsResponse, error)
	Amendment(context.Context, *QueryAmendmentRequest) (*QueryAmendmentResponse, error)
	ProtectedLimits(context.Context, *QueryProtectedLimitsRequest) (*QueryProtectedLimitsResponse, error)
}

type UnimplementedMsgServer struct{}
type UnimplementedQueryServer struct{}

func (UnimplementedMsgServer) ProposeConstitutionAmendment(context.Context, *MsgProposeConstitutionAmendment) (*MsgProposeConstitutionAmendmentResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ProposeConstitutionAmendment not implemented")
}
func (UnimplementedMsgServer) VoteConstitutionAmendment(context.Context, *MsgVoteConstitutionAmendment) (*MsgVoteConstitutionAmendmentResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method VoteConstitutionAmendment not implemented")
}
func (UnimplementedMsgServer) ExecuteConstitutionAmendment(context.Context, *MsgExecuteConstitutionAmendment) (*MsgExecuteConstitutionAmendmentResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ExecuteConstitutionAmendment not implemented")
}
func (UnimplementedMsgServer) CancelConstitutionAmendment(context.Context, *MsgCancelConstitutionAmendment) (*MsgCancelConstitutionAmendmentResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CancelConstitutionAmendment not implemented")
}

func RegisterMsgServer(s grpc.Server, srv MsgServer)		{ s.RegisterService(&Msg_serviceDesc, srv) }
func RegisterQueryServer(s grpc.Server, srv QueryServer)	{ s.RegisterService(&Query_serviceDesc, srv) }

type serviceCall func(context.Context, interface{}, interface{}) (interface{}, error)

var Msg_serviceDesc = grpcgo.ServiceDesc{
	ServiceName:	"l1.constitution.v1.Msg",
	HandlerType:	(*MsgServer)(nil),
	Methods: []grpcgo.MethodDesc{
		methodDesc("ProposeConstitutionAmendment", serviceHandler("ProposeConstitutionAmendment", func() interface{} { return new(MsgProposeConstitutionAmendment) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(MsgServer).ProposeConstitutionAmendment(ctx, req.(*MsgProposeConstitutionAmendment))
		})),
		methodDesc("VoteConstitutionAmendment", serviceHandler("VoteConstitutionAmendment", func() interface{} { return new(MsgVoteConstitutionAmendment) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(MsgServer).VoteConstitutionAmendment(ctx, req.(*MsgVoteConstitutionAmendment))
		})),
		methodDesc("ExecuteConstitutionAmendment", serviceHandler("ExecuteConstitutionAmendment", func() interface{} { return new(MsgExecuteConstitutionAmendment) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(MsgServer).ExecuteConstitutionAmendment(ctx, req.(*MsgExecuteConstitutionAmendment))
		})),
		methodDesc("CancelConstitutionAmendment", serviceHandler("CancelConstitutionAmendment", func() interface{} { return new(MsgCancelConstitutionAmendment) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(MsgServer).CancelConstitutionAmendment(ctx, req.(*MsgCancelConstitutionAmendment))
		})),
	},
	Metadata:	"l1/constitution/v1/tx.proto",
}

var Query_serviceDesc = grpcgo.ServiceDesc{
	ServiceName:	"l1.constitution.v1.Query",
	HandlerType:	(*QueryServer)(nil),
	Methods: []grpcgo.MethodDesc{
		methodDesc("Params", serviceHandler("Params", func() interface{} { return new(QueryParamsRequest) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(QueryServer).Params(ctx, req.(*QueryParamsRequest))
		})),
		methodDesc("Constitution", serviceHandler("Constitution", func() interface{} { return new(QueryConstitutionRequest) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(QueryServer).Constitution(ctx, req.(*QueryConstitutionRequest))
		})),
		methodDesc("PendingAmendments", serviceHandler("PendingAmendments", func() interface{} { return new(QueryPendingAmendmentsRequest) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(QueryServer).PendingAmendments(ctx, req.(*QueryPendingAmendmentsRequest))
		})),
		methodDesc("Amendment", serviceHandler("Amendment", func() interface{} { return new(QueryAmendmentRequest) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(QueryServer).Amendment(ctx, req.(*QueryAmendmentRequest))
		})),
		methodDesc("ProtectedLimits", serviceHandler("ProtectedLimits", func() interface{} { return new(QueryProtectedLimitsRequest) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(QueryServer).ProtectedLimits(ctx, req.(*QueryProtectedLimitsRequest))
		})),
	},
	Metadata:	"l1/constitution/v1/query.proto",
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
	gogoproto.RegisterFile("l1/constitution/v1/tx.proto", buildServiceFileDescriptor("l1/constitution/v1/tx.proto", "l1.constitution.v1", "Msg", []string{"MsgProposeConstitutionAmendment", "MsgProposeConstitutionAmendmentResponse", "MsgVoteConstitutionAmendment", "MsgVoteConstitutionAmendmentResponse", "MsgExecuteConstitutionAmendment", "MsgExecuteConstitutionAmendmentResponse", "MsgCancelConstitutionAmendment", "MsgCancelConstitutionAmendmentResponse"}, [][3]string{{"ProposeConstitutionAmendment", "MsgProposeConstitutionAmendment", "MsgProposeConstitutionAmendmentResponse"}, {"VoteConstitutionAmendment", "MsgVoteConstitutionAmendment", "MsgVoteConstitutionAmendmentResponse"}, {"ExecuteConstitutionAmendment", "MsgExecuteConstitutionAmendment", "MsgExecuteConstitutionAmendmentResponse"}, {"CancelConstitutionAmendment", "MsgCancelConstitutionAmendment", "MsgCancelConstitutionAmendmentResponse"}}))
	gogoproto.RegisterFile("l1/constitution/v1/query.proto", buildServiceFileDescriptor("l1/constitution/v1/query.proto", "l1.constitution.v1", "Query", []string{"QueryParamsRequest", "QueryParamsResponse", "QueryConstitutionRequest", "QueryConstitutionResponse", "QueryPendingAmendmentsRequest", "QueryPendingAmendmentsResponse", "QueryAmendmentRequest", "QueryAmendmentResponse", "QueryProtectedLimitsRequest", "QueryProtectedLimitsResponse"}, [][3]string{{"Params", "QueryParamsRequest", "QueryParamsResponse"}, {"Constitution", "QueryConstitutionRequest", "QueryConstitutionResponse"}, {"PendingAmendments", "QueryPendingAmendmentsRequest", "QueryPendingAmendmentsResponse"}, {"Amendment", "QueryAmendmentRequest", "QueryAmendmentResponse"}, {"ProtectedLimits", "QueryProtectedLimitsRequest", "QueryProtectedLimitsResponse"}}))
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
		{(*MsgProposeConstitutionAmendment)(nil), "l1.constitution.v1.MsgProposeConstitutionAmendment"},
		{(*MsgProposeConstitutionAmendmentResponse)(nil), "l1.constitution.v1.MsgProposeConstitutionAmendmentResponse"},
		{(*MsgVoteConstitutionAmendment)(nil), "l1.constitution.v1.MsgVoteConstitutionAmendment"},
		{(*MsgVoteConstitutionAmendmentResponse)(nil), "l1.constitution.v1.MsgVoteConstitutionAmendmentResponse"},
		{(*MsgExecuteConstitutionAmendment)(nil), "l1.constitution.v1.MsgExecuteConstitutionAmendment"},
		{(*MsgExecuteConstitutionAmendmentResponse)(nil), "l1.constitution.v1.MsgExecuteConstitutionAmendmentResponse"},
		{(*MsgCancelConstitutionAmendment)(nil), "l1.constitution.v1.MsgCancelConstitutionAmendment"},
		{(*MsgCancelConstitutionAmendmentResponse)(nil), "l1.constitution.v1.MsgCancelConstitutionAmendmentResponse"},
		{(*QueryParamsRequest)(nil), "l1.constitution.v1.QueryParamsRequest"},
		{(*QueryParamsResponse)(nil), "l1.constitution.v1.QueryParamsResponse"},
		{(*QueryConstitutionRequest)(nil), "l1.constitution.v1.QueryConstitutionRequest"},
		{(*QueryConstitutionResponse)(nil), "l1.constitution.v1.QueryConstitutionResponse"},
		{(*QueryPendingAmendmentsRequest)(nil), "l1.constitution.v1.QueryPendingAmendmentsRequest"},
		{(*QueryPendingAmendmentsResponse)(nil), "l1.constitution.v1.QueryPendingAmendmentsResponse"},
		{(*QueryAmendmentRequest)(nil), "l1.constitution.v1.QueryAmendmentRequest"},
		{(*QueryAmendmentResponse)(nil), "l1.constitution.v1.QueryAmendmentResponse"},
		{(*QueryProtectedLimitsRequest)(nil), "l1.constitution.v1.QueryProtectedLimitsRequest"},
		{(*QueryProtectedLimitsResponse)(nil), "l1.constitution.v1.QueryProtectedLimitsResponse"},
	} {
		gogoproto.RegisterType(item.msg, item.name)
	}
}

func (m *MsgProposeConstitutionAmendment) Reset()	{ *m = MsgProposeConstitutionAmendment{} }
func (m *MsgVoteConstitutionAmendment) Reset()		{ *m = MsgVoteConstitutionAmendment{} }
func (m *MsgExecuteConstitutionAmendment) Reset()	{ *m = MsgExecuteConstitutionAmendment{} }
func (m *MsgCancelConstitutionAmendment) Reset()	{ *m = MsgCancelConstitutionAmendment{} }
func (m *MsgProposeConstitutionAmendmentResponse) Reset() {
	*m = MsgProposeConstitutionAmendmentResponse{}
}
func (m *MsgVoteConstitutionAmendmentResponse) Reset()	{ *m = MsgVoteConstitutionAmendmentResponse{} }
func (m *MsgExecuteConstitutionAmendmentResponse) Reset() {
	*m = MsgExecuteConstitutionAmendmentResponse{}
}
func (m *MsgCancelConstitutionAmendmentResponse) Reset() {
	*m = MsgCancelConstitutionAmendmentResponse{}
}
func (m *QueryParamsRequest) Reset()			{ *m = QueryParamsRequest{} }
func (m *QueryParamsResponse) Reset()			{ *m = QueryParamsResponse{} }
func (m *QueryConstitutionRequest) Reset()		{ *m = QueryConstitutionRequest{} }
func (m *QueryConstitutionResponse) Reset()		{ *m = QueryConstitutionResponse{} }
func (m *QueryPendingAmendmentsRequest) Reset()		{ *m = QueryPendingAmendmentsRequest{} }
func (m *QueryPendingAmendmentsResponse) Reset()	{ *m = QueryPendingAmendmentsResponse{} }
func (m *QueryAmendmentRequest) Reset()			{ *m = QueryAmendmentRequest{} }
func (m *QueryAmendmentResponse) Reset()		{ *m = QueryAmendmentResponse{} }
func (m *QueryProtectedLimitsRequest) Reset()		{ *m = QueryProtectedLimitsRequest{} }
func (m *QueryProtectedLimitsResponse) Reset()		{ *m = QueryProtectedLimitsResponse{} }

func (m *MsgProposeConstitutionAmendment) String() string	{ return gogoproto.CompactTextString(m) }
func (m *MsgVoteConstitutionAmendment) String() string		{ return gogoproto.CompactTextString(m) }
func (m *MsgExecuteConstitutionAmendment) String() string	{ return gogoproto.CompactTextString(m) }
func (m *MsgCancelConstitutionAmendment) String() string	{ return gogoproto.CompactTextString(m) }
func (m *MsgProposeConstitutionAmendmentResponse) String() string {
	return gogoproto.CompactTextString(m)
}
func (m *MsgVoteConstitutionAmendmentResponse) String() string {
	return gogoproto.CompactTextString(m)
}
func (m *MsgExecuteConstitutionAmendmentResponse) String() string {
	return gogoproto.CompactTextString(m)
}
func (m *MsgCancelConstitutionAmendmentResponse) String() string {
	return gogoproto.CompactTextString(m)
}
func (m *QueryParamsRequest) String() string		{ return gogoproto.CompactTextString(m) }
func (m *QueryParamsResponse) String() string		{ return gogoproto.CompactTextString(m) }
func (m *QueryConstitutionRequest) String() string	{ return gogoproto.CompactTextString(m) }
func (m *QueryConstitutionResponse) String() string	{ return gogoproto.CompactTextString(m) }
func (m *QueryPendingAmendmentsRequest) String() string	{ return gogoproto.CompactTextString(m) }
func (m *QueryPendingAmendmentsResponse) String() string {
	return gogoproto.CompactTextString(m)
}
func (m *QueryAmendmentRequest) String() string		{ return gogoproto.CompactTextString(m) }
func (m *QueryAmendmentResponse) String() string	{ return gogoproto.CompactTextString(m) }
func (m *QueryProtectedLimitsRequest) String() string	{ return gogoproto.CompactTextString(m) }
func (m *QueryProtectedLimitsResponse) String() string {
	return gogoproto.CompactTextString(m)
}

func (*MsgProposeConstitutionAmendment) ProtoMessage()		{}
func (*MsgVoteConstitutionAmendment) ProtoMessage()		{}
func (*MsgExecuteConstitutionAmendment) ProtoMessage()		{}
func (*MsgCancelConstitutionAmendment) ProtoMessage()		{}
func (*MsgProposeConstitutionAmendmentResponse) ProtoMessage()	{}
func (*MsgVoteConstitutionAmendmentResponse) ProtoMessage()	{}
func (*MsgExecuteConstitutionAmendmentResponse) ProtoMessage()	{}
func (*MsgCancelConstitutionAmendmentResponse) ProtoMessage()	{}
func (*QueryParamsRequest) ProtoMessage()			{}
func (*QueryParamsResponse) ProtoMessage()			{}
func (*QueryConstitutionRequest) ProtoMessage()			{}
func (*QueryConstitutionResponse) ProtoMessage()		{}
func (*QueryPendingAmendmentsRequest) ProtoMessage()		{}
func (*QueryPendingAmendmentsResponse) ProtoMessage()		{}
func (*QueryAmendmentRequest) ProtoMessage()			{}
func (*QueryAmendmentResponse) ProtoMessage()			{}
func (*QueryProtectedLimitsRequest) ProtoMessage()		{}
func (*QueryProtectedLimitsResponse) ProtoMessage()		{}
