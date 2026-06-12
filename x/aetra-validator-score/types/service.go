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

type MsgUpdateValidatorScoreParamsResponse struct{}
type MsgUpdateValidatorScoresResponse struct{}

type MsgServer interface {
	UpdateValidatorScoreParams(context.Context, *MsgUpdateValidatorScoreParams) (*MsgUpdateValidatorScoreParamsResponse, error)
	UpdateValidatorScores(context.Context, *MsgUpdateValidatorScores) (*MsgUpdateValidatorScoresResponse, error)
}

type UnimplementedMsgServer struct{}

func (UnimplementedMsgServer) UpdateValidatorScoreParams(context.Context, *MsgUpdateValidatorScoreParams) (*MsgUpdateValidatorScoreParamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateValidatorScoreParams not implemented")
}
func (UnimplementedMsgServer) UpdateValidatorScores(context.Context, *MsgUpdateValidatorScores) (*MsgUpdateValidatorScoresResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateValidatorScores not implemented")
}

type QueryServer interface {
	Params(context.Context, *QueryParamsRequest) (*QueryParamsResponse, error)
	ValidatorScore(context.Context, *QueryValidatorScoreRequest) (*QueryValidatorScoreResponse, error)
	PublicValidatorMetrics(context.Context, *QueryPublicValidatorMetricsRequest) (*QueryPublicValidatorMetricsResponse, error)
	AllValidatorScores(context.Context, *QueryAllValidatorScoresRequest) (*QueryAllValidatorScoresResponse, error)
}

func RegisterMsgServer(s grpc.Server, srv MsgServer)		{ s.RegisterService(&Msg_serviceDesc, srv) }
func RegisterQueryServer(s grpc.Server, srv QueryServer)	{ s.RegisterService(&Query_serviceDesc, srv) }

type serviceCall func(context.Context, interface{}, interface{}) (interface{}, error)

var Msg_serviceDesc = grpcgo.ServiceDesc{
	ServiceName:	"l1.aetravalidatorscore.v1.Msg",
	HandlerType:	(*MsgServer)(nil),
	Methods: []grpcgo.MethodDesc{
		methodDesc("UpdateValidatorScoreParams", serviceHandler("UpdateValidatorScoreParams", func() interface{} { return new(MsgUpdateValidatorScoreParams) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(MsgServer).UpdateValidatorScoreParams(ctx, req.(*MsgUpdateValidatorScoreParams))
		})),
		methodDesc("UpdateValidatorScores", serviceHandler("UpdateValidatorScores", func() interface{} { return new(MsgUpdateValidatorScores) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(MsgServer).UpdateValidatorScores(ctx, req.(*MsgUpdateValidatorScores))
		})),
	},
	Metadata:	"l1/aetravalidatorscore/v1/tx.proto",
}

var Query_serviceDesc = grpcgo.ServiceDesc{
	ServiceName:	"l1.aetravalidatorscore.v1.Query",
	HandlerType:	(*QueryServer)(nil),
	Methods: []grpcgo.MethodDesc{
		methodDesc("Params", serviceHandler("Params", func() interface{} { return new(QueryParamsRequest) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(QueryServer).Params(ctx, req.(*QueryParamsRequest))
		})),
		methodDesc("ValidatorScore", serviceHandler("ValidatorScore", func() interface{} { return new(QueryValidatorScoreRequest) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(QueryServer).ValidatorScore(ctx, req.(*QueryValidatorScoreRequest))
		})),
		methodDesc("PublicValidatorMetrics", serviceHandler("PublicValidatorMetrics", func() interface{} { return new(QueryPublicValidatorMetricsRequest) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(QueryServer).PublicValidatorMetrics(ctx, req.(*QueryPublicValidatorMetricsRequest))
		})),
		methodDesc("AllValidatorScores", serviceHandler("AllValidatorScores", func() interface{} { return new(QueryAllValidatorScoresRequest) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(QueryServer).AllValidatorScores(ctx, req.(*QueryAllValidatorScoresRequest))
		})),
	},
	Metadata:	"l1/aetravalidatorscore/v1/query.proto",
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
	gogoproto.RegisterFile("l1/aetravalidatorscore/v1/tx.proto", buildServiceFileDescriptor("l1/aetravalidatorscore/v1/tx.proto", "l1.aetravalidatorscore.v1", "Msg", []string{"MsgUpdateValidatorScoreParams", "MsgUpdateValidatorScoreParamsResponse", "MsgUpdateValidatorScores", "MsgUpdateValidatorScoresResponse"}, [][3]string{{"UpdateValidatorScoreParams", "MsgUpdateValidatorScoreParams", "MsgUpdateValidatorScoreParamsResponse"}, {"UpdateValidatorScores", "MsgUpdateValidatorScores", "MsgUpdateValidatorScoresResponse"}}))
	gogoproto.RegisterFile("l1/aetravalidatorscore/v1/query.proto", buildServiceFileDescriptor("l1/aetravalidatorscore/v1/query.proto", "l1.aetravalidatorscore.v1", "Query", []string{"QueryParamsRequest", "QueryParamsResponse", "QueryValidatorScoreRequest", "QueryValidatorScoreResponse", "QueryPublicValidatorMetricsRequest", "QueryPublicValidatorMetricsResponse", "QueryAllValidatorScoresRequest", "QueryAllValidatorScoresResponse"}, [][3]string{{"Params", "QueryParamsRequest", "QueryParamsResponse"}, {"ValidatorScore", "QueryValidatorScoreRequest", "QueryValidatorScoreResponse"}, {"PublicValidatorMetrics", "QueryPublicValidatorMetricsRequest", "QueryPublicValidatorMetricsResponse"}, {"AllValidatorScores", "QueryAllValidatorScoresRequest", "QueryAllValidatorScoresResponse"}}))
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
	gogoproto.RegisterType((*MsgUpdateValidatorScoreParams)(nil), "l1.aetravalidatorscore.v1.MsgUpdateValidatorScoreParams")
	gogoproto.RegisterType((*MsgUpdateValidatorScoreParamsResponse)(nil), "l1.aetravalidatorscore.v1.MsgUpdateValidatorScoreParamsResponse")
	gogoproto.RegisterType((*MsgUpdateValidatorScores)(nil), "l1.aetravalidatorscore.v1.MsgUpdateValidatorScores")
	gogoproto.RegisterType((*MsgUpdateValidatorScoresResponse)(nil), "l1.aetravalidatorscore.v1.MsgUpdateValidatorScoresResponse")
}

func (m *MsgUpdateValidatorScoreParams) Reset()		{ *m = MsgUpdateValidatorScoreParams{} }
func (m *MsgUpdateValidatorScores) Reset()		{ *m = MsgUpdateValidatorScores{} }
func (m *MsgUpdateValidatorScoreParamsResponse) Reset()	{ *m = MsgUpdateValidatorScoreParamsResponse{} }
func (m *MsgUpdateValidatorScoresResponse) Reset()	{ *m = MsgUpdateValidatorScoresResponse{} }

func (m *MsgUpdateValidatorScoreParams) String() string	{ return gogoproto.CompactTextString(m) }
func (m *MsgUpdateValidatorScores) String() string	{ return gogoproto.CompactTextString(m) }
func (m *MsgUpdateValidatorScoreParamsResponse) String() string {
	return gogoproto.CompactTextString(m)
}
func (m *MsgUpdateValidatorScoresResponse) String() string	{ return gogoproto.CompactTextString(m) }

func (*MsgUpdateValidatorScoreParams) ProtoMessage()		{}
func (*MsgUpdateValidatorScores) ProtoMessage()			{}
func (*MsgUpdateValidatorScoreParamsResponse) ProtoMessage()	{}
func (*MsgUpdateValidatorScoresResponse) ProtoMessage()		{}

func (m *QueryParamsRequest) Reset()				{ *m = QueryParamsRequest{} }
func (m *QueryParamsResponse) Reset()				{ *m = QueryParamsResponse{} }
func (m *QueryValidatorScoreRequest) Reset()			{ *m = QueryValidatorScoreRequest{} }
func (m *QueryValidatorScoreResponse) Reset()			{ *m = QueryValidatorScoreResponse{} }
func (m *QueryPublicValidatorMetricsRequest) Reset()		{ *m = QueryPublicValidatorMetricsRequest{} }
func (m *QueryPublicValidatorMetricsResponse) Reset()		{ *m = QueryPublicValidatorMetricsResponse{} }
func (m *QueryAllValidatorScoresRequest) Reset()		{ *m = QueryAllValidatorScoresRequest{} }
func (m *QueryAllValidatorScoresResponse) Reset()		{ *m = QueryAllValidatorScoresResponse{} }
func (m *QueryParamsRequest) String() string			{ return gogoproto.CompactTextString(m) }
func (m *QueryParamsResponse) String() string			{ return gogoproto.CompactTextString(m) }
func (m *QueryValidatorScoreRequest) String() string		{ return gogoproto.CompactTextString(m) }
func (m *QueryValidatorScoreResponse) String() string		{ return gogoproto.CompactTextString(m) }
func (m *QueryPublicValidatorMetricsRequest) String() string	{ return gogoproto.CompactTextString(m) }
func (m *QueryPublicValidatorMetricsResponse) String() string {
	return gogoproto.CompactTextString(m)
}
func (m *QueryAllValidatorScoresRequest) String() string	{ return gogoproto.CompactTextString(m) }
func (m *QueryAllValidatorScoresResponse) String() string	{ return gogoproto.CompactTextString(m) }
func (*QueryParamsRequest) ProtoMessage()			{}
func (*QueryParamsResponse) ProtoMessage()			{}
func (*QueryValidatorScoreRequest) ProtoMessage()		{}
func (*QueryValidatorScoreResponse) ProtoMessage()		{}
func (*QueryPublicValidatorMetricsRequest) ProtoMessage()	{}
func (*QueryPublicValidatorMetricsResponse) ProtoMessage()	{}
func (*QueryAllValidatorScoresRequest) ProtoMessage()		{}
func (*QueryAllValidatorScoresResponse) ProtoMessage()		{}
