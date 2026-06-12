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

type MsgUpdateEconomicsParamsResponse struct{}
type MsgApplyEpochEconomicsResponse struct{}

type MsgServer interface {
	UpdateEconomicsParams(context.Context, *MsgUpdateEconomicsParams) (*MsgUpdateEconomicsParamsResponse, error)
	ApplyEpochEconomics(context.Context, *MsgApplyEpochEconomics) (*MsgApplyEpochEconomicsResponse, error)
}

type UnimplementedMsgServer struct{}

func (UnimplementedMsgServer) UpdateEconomicsParams(context.Context, *MsgUpdateEconomicsParams) (*MsgUpdateEconomicsParamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateEconomicsParams not implemented")
}
func (UnimplementedMsgServer) ApplyEpochEconomics(context.Context, *MsgApplyEpochEconomics) (*MsgApplyEpochEconomicsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ApplyEpochEconomics not implemented")
}

type QueryServer interface {
	CurrentInflation(context.Context, *QueryCurrentInflationRequest) (*QueryCurrentInflationResponse, error)
	CurrentBondedRatio(context.Context, *QueryCurrentBondedRatioRequest) (*QueryCurrentBondedRatioResponse, error)
	EstimatedAPR(context.Context, *QueryEstimatedAPRRequest) (*QueryEstimatedAPRResponse, error)
	FeeSplitParams(context.Context, *QueryFeeSplitParamsRequest) (*QueryFeeSplitParamsResponse, error)
	BurnedSupply(context.Context, *QueryBurnedSupplyRequest) (*QueryBurnedSupplyResponse, error)
	TreasuryBalance(context.Context, *QueryTreasuryBalanceRequest) (*QueryTreasuryBalanceResponse, error)
	EpochRewardSummary(context.Context, *QueryEpochRewardSummaryRequest) (*QueryEpochRewardSummaryResponse, error)
}

func RegisterMsgServer(s grpc.Server, srv MsgServer)		{ s.RegisterService(&Msg_serviceDesc, srv) }
func RegisterQueryServer(s grpc.Server, srv QueryServer)	{ s.RegisterService(&Query_serviceDesc, srv) }

type serviceCall func(context.Context, interface{}, interface{}) (interface{}, error)

var Msg_serviceDesc = grpcgo.ServiceDesc{
	ServiceName:	"l1.aetraeconomics.v1.Msg",
	HandlerType:	(*MsgServer)(nil),
	Methods: []grpcgo.MethodDesc{
		methodDesc("UpdateEconomicsParams", serviceHandler("UpdateEconomicsParams", func() interface{} { return new(MsgUpdateEconomicsParams) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(MsgServer).UpdateEconomicsParams(ctx, req.(*MsgUpdateEconomicsParams))
		})),
		methodDesc("ApplyEpochEconomics", serviceHandler("ApplyEpochEconomics", func() interface{} { return new(MsgApplyEpochEconomics) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(MsgServer).ApplyEpochEconomics(ctx, req.(*MsgApplyEpochEconomics))
		})),
	},
	Metadata:	"l1/aetraeconomics/v1/tx.proto",
}

var Query_serviceDesc = grpcgo.ServiceDesc{
	ServiceName:	"l1.aetraeconomics.v1.Query",
	HandlerType:	(*QueryServer)(nil),
	Methods: []grpcgo.MethodDesc{
		methodDesc("CurrentInflation", serviceHandler("CurrentInflation", func() interface{} { return new(QueryCurrentInflationRequest) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(QueryServer).CurrentInflation(ctx, req.(*QueryCurrentInflationRequest))
		})),
		methodDesc("CurrentBondedRatio", serviceHandler("CurrentBondedRatio", func() interface{} { return new(QueryCurrentBondedRatioRequest) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(QueryServer).CurrentBondedRatio(ctx, req.(*QueryCurrentBondedRatioRequest))
		})),
		methodDesc("EstimatedAPR", serviceHandler("EstimatedAPR", func() interface{} { return new(QueryEstimatedAPRRequest) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(QueryServer).EstimatedAPR(ctx, req.(*QueryEstimatedAPRRequest))
		})),
		methodDesc("FeeSplitParams", serviceHandler("FeeSplitParams", func() interface{} { return new(QueryFeeSplitParamsRequest) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(QueryServer).FeeSplitParams(ctx, req.(*QueryFeeSplitParamsRequest))
		})),
		methodDesc("BurnedSupply", serviceHandler("BurnedSupply", func() interface{} { return new(QueryBurnedSupplyRequest) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(QueryServer).BurnedSupply(ctx, req.(*QueryBurnedSupplyRequest))
		})),
		methodDesc("TreasuryBalance", serviceHandler("TreasuryBalance", func() interface{} { return new(QueryTreasuryBalanceRequest) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(QueryServer).TreasuryBalance(ctx, req.(*QueryTreasuryBalanceRequest))
		})),
		methodDesc("EpochRewardSummary", serviceHandler("EpochRewardSummary", func() interface{} { return new(QueryEpochRewardSummaryRequest) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(QueryServer).EpochRewardSummary(ctx, req.(*QueryEpochRewardSummaryRequest))
		})),
	},
	Metadata:	"l1/aetraeconomics/v1/query.proto",
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
		info := &grpcgo.UnaryServerInfo{Server: srv, FullMethod: method}
		return interceptor(ctx, req, info, func(ctx context.Context, request interface{}) (interface{}, error) {
			return call(ctx, srv, request)
		})
	}
}

func init() {
	registerServiceTypes()
	gogoproto.RegisterFile("l1/aetraeconomics/v1/tx.proto", buildServiceFileDescriptor("l1/aetraeconomics/v1/tx.proto", "l1.aetraeconomics.v1", "Msg", []string{"MsgUpdateEconomicsParams", "MsgUpdateEconomicsParamsResponse", "MsgApplyEpochEconomics", "MsgApplyEpochEconomicsResponse"}, [][3]string{{"UpdateEconomicsParams", "MsgUpdateEconomicsParams", "MsgUpdateEconomicsParamsResponse"}, {"ApplyEpochEconomics", "MsgApplyEpochEconomics", "MsgApplyEpochEconomicsResponse"}}))
	gogoproto.RegisterFile("l1/aetraeconomics/v1/query.proto", buildServiceFileDescriptor("l1/aetraeconomics/v1/query.proto", "l1.aetraeconomics.v1", "Query", []string{"QueryCurrentInflationRequest", "QueryCurrentInflationResponse", "QueryCurrentBondedRatioRequest", "QueryCurrentBondedRatioResponse", "QueryEstimatedAPRRequest", "QueryEstimatedAPRResponse", "QueryFeeSplitParamsRequest", "QueryFeeSplitParamsResponse", "QueryBurnedSupplyRequest", "QueryBurnedSupplyResponse", "QueryTreasuryBalanceRequest", "QueryTreasuryBalanceResponse", "QueryEpochRewardSummaryRequest", "QueryEpochRewardSummaryResponse"}, [][3]string{{"CurrentInflation", "QueryCurrentInflationRequest", "QueryCurrentInflationResponse"}, {"CurrentBondedRatio", "QueryCurrentBondedRatioRequest", "QueryCurrentBondedRatioResponse"}, {"EstimatedAPR", "QueryEstimatedAPRRequest", "QueryEstimatedAPRResponse"}, {"FeeSplitParams", "QueryFeeSplitParamsRequest", "QueryFeeSplitParamsResponse"}, {"BurnedSupply", "QueryBurnedSupplyRequest", "QueryBurnedSupplyResponse"}, {"TreasuryBalance", "QueryTreasuryBalanceRequest", "QueryTreasuryBalanceResponse"}, {"EpochRewardSummary", "QueryEpochRewardSummaryRequest", "QueryEpochRewardSummaryResponse"}}))
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
	gogoproto.RegisterType((*MsgUpdateEconomicsParams)(nil), "l1.aetraeconomics.v1.MsgUpdateEconomicsParams")
	gogoproto.RegisterType((*MsgUpdateEconomicsParamsResponse)(nil), "l1.aetraeconomics.v1.MsgUpdateEconomicsParamsResponse")
	gogoproto.RegisterType((*MsgApplyEpochEconomics)(nil), "l1.aetraeconomics.v1.MsgApplyEpochEconomics")
	gogoproto.RegisterType((*MsgApplyEpochEconomicsResponse)(nil), "l1.aetraeconomics.v1.MsgApplyEpochEconomicsResponse")
}

func (m *MsgUpdateEconomicsParams) Reset()		{ *m = MsgUpdateEconomicsParams{} }
func (m *MsgApplyEpochEconomics) Reset()		{ *m = MsgApplyEpochEconomics{} }
func (m *MsgUpdateEconomicsParamsResponse) Reset()	{ *m = MsgUpdateEconomicsParamsResponse{} }
func (m *MsgApplyEpochEconomicsResponse) Reset()	{ *m = MsgApplyEpochEconomicsResponse{} }

func (m *MsgUpdateEconomicsParams) String() string		{ return gogoproto.CompactTextString(m) }
func (m *MsgApplyEpochEconomics) String() string		{ return gogoproto.CompactTextString(m) }
func (m *MsgUpdateEconomicsParamsResponse) String() string	{ return gogoproto.CompactTextString(m) }
func (m *MsgApplyEpochEconomicsResponse) String() string	{ return gogoproto.CompactTextString(m) }

func (*MsgUpdateEconomicsParams) ProtoMessage()		{}
func (*MsgApplyEpochEconomics) ProtoMessage()		{}
func (*MsgUpdateEconomicsParamsResponse) ProtoMessage()	{}
func (*MsgApplyEpochEconomicsResponse) ProtoMessage()	{}

func (m *QueryCurrentInflationRequest) Reset()			{ *m = QueryCurrentInflationRequest{} }
func (m *QueryCurrentInflationResponse) Reset()			{ *m = QueryCurrentInflationResponse{} }
func (m *QueryCurrentBondedRatioRequest) Reset()		{ *m = QueryCurrentBondedRatioRequest{} }
func (m *QueryCurrentBondedRatioResponse) Reset()		{ *m = QueryCurrentBondedRatioResponse{} }
func (m *QueryEstimatedAPRRequest) Reset()			{ *m = QueryEstimatedAPRRequest{} }
func (m *QueryEstimatedAPRResponse) Reset()			{ *m = QueryEstimatedAPRResponse{} }
func (m *QueryFeeSplitParamsRequest) Reset()			{ *m = QueryFeeSplitParamsRequest{} }
func (m *QueryFeeSplitParamsResponse) Reset()			{ *m = QueryFeeSplitParamsResponse{} }
func (m *QueryBurnedSupplyRequest) Reset()			{ *m = QueryBurnedSupplyRequest{} }
func (m *QueryBurnedSupplyResponse) Reset()			{ *m = QueryBurnedSupplyResponse{} }
func (m *QueryTreasuryBalanceRequest) Reset()			{ *m = QueryTreasuryBalanceRequest{} }
func (m *QueryTreasuryBalanceResponse) Reset()			{ *m = QueryTreasuryBalanceResponse{} }
func (m *QueryEpochRewardSummaryRequest) Reset()		{ *m = QueryEpochRewardSummaryRequest{} }
func (m *QueryEpochRewardSummaryResponse) Reset()		{ *m = QueryEpochRewardSummaryResponse{} }
func (m *QueryCurrentInflationRequest) String() string		{ return gogoproto.CompactTextString(m) }
func (m *QueryCurrentInflationResponse) String() string		{ return gogoproto.CompactTextString(m) }
func (m *QueryCurrentBondedRatioRequest) String() string	{ return gogoproto.CompactTextString(m) }
func (m *QueryCurrentBondedRatioResponse) String() string	{ return gogoproto.CompactTextString(m) }
func (m *QueryEstimatedAPRRequest) String() string		{ return gogoproto.CompactTextString(m) }
func (m *QueryEstimatedAPRResponse) String() string		{ return gogoproto.CompactTextString(m) }
func (m *QueryFeeSplitParamsRequest) String() string		{ return gogoproto.CompactTextString(m) }
func (m *QueryFeeSplitParamsResponse) String() string		{ return gogoproto.CompactTextString(m) }
func (m *QueryBurnedSupplyRequest) String() string		{ return gogoproto.CompactTextString(m) }
func (m *QueryBurnedSupplyResponse) String() string		{ return gogoproto.CompactTextString(m) }
func (m *QueryTreasuryBalanceRequest) String() string		{ return gogoproto.CompactTextString(m) }
func (m *QueryTreasuryBalanceResponse) String() string		{ return gogoproto.CompactTextString(m) }
func (m *QueryEpochRewardSummaryRequest) String() string	{ return gogoproto.CompactTextString(m) }
func (m *QueryEpochRewardSummaryResponse) String() string	{ return gogoproto.CompactTextString(m) }
func (*QueryCurrentInflationRequest) ProtoMessage()		{}
func (*QueryCurrentInflationResponse) ProtoMessage()		{}
func (*QueryCurrentBondedRatioRequest) ProtoMessage()		{}
func (*QueryCurrentBondedRatioResponse) ProtoMessage()		{}
func (*QueryEstimatedAPRRequest) ProtoMessage()			{}
func (*QueryEstimatedAPRResponse) ProtoMessage()		{}
func (*QueryFeeSplitParamsRequest) ProtoMessage()		{}
func (*QueryFeeSplitParamsResponse) ProtoMessage()		{}
func (*QueryBurnedSupplyRequest) ProtoMessage()			{}
func (*QueryBurnedSupplyResponse) ProtoMessage()		{}
func (*QueryTreasuryBalanceRequest) ProtoMessage()		{}
func (*QueryTreasuryBalanceResponse) ProtoMessage()		{}
func (*QueryEpochRewardSummaryRequest) ProtoMessage()		{}
func (*QueryEpochRewardSummaryResponse) ProtoMessage()		{}
