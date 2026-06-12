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

type MsgUpdateStakingPolicyParamsResponse struct{}
type MsgRegisterValidatorIdentityResponse struct{}
type MsgAcknowledgeConcentrationWarningResponse struct{}

type MsgServer interface {
	UpdateStakingPolicyParams(context.Context, *MsgUpdateStakingPolicyParams) (*MsgUpdateStakingPolicyParamsResponse, error)
	RegisterValidatorIdentity(context.Context, *MsgRegisterValidatorIdentity) (*MsgRegisterValidatorIdentityResponse, error)
	AcknowledgeConcentrationWarning(context.Context, *MsgAcknowledgeConcentrationWarning) (*MsgAcknowledgeConcentrationWarningResponse, error)
}

type UnimplementedMsgServer struct{}

func (UnimplementedMsgServer) UpdateStakingPolicyParams(context.Context, *MsgUpdateStakingPolicyParams) (*MsgUpdateStakingPolicyParamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateStakingPolicyParams not implemented")
}
func (UnimplementedMsgServer) RegisterValidatorIdentity(context.Context, *MsgRegisterValidatorIdentity) (*MsgRegisterValidatorIdentityResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegisterValidatorIdentity not implemented")
}
func (UnimplementedMsgServer) AcknowledgeConcentrationWarning(context.Context, *MsgAcknowledgeConcentrationWarning) (*MsgAcknowledgeConcentrationWarningResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AcknowledgeConcentrationWarning not implemented")
}

type QueryServer interface {
	Params(context.Context, *QueryParamsRequest) (*QueryParamsResponse, error)
	ValidatorEffectivePower(context.Context, *QueryValidatorEffectivePowerRequest) (*QueryValidatorEffectivePowerResponse, error)
	ValidatorStake(context.Context, *QueryValidatorStakeRequest) (*QueryValidatorStakeResponse, error)
	TopNConcentration(context.Context, *QueryTopNConcentrationRequest) (*QueryTopNConcentrationResponse, error)
	ValidatorRewardMultiplier(context.Context, *QueryValidatorRewardMultiplierRequest) (*QueryValidatorRewardMultiplierResponse, error)
	DelegationWarningStatus(context.Context, *QueryDelegationWarningStatusRequest) (*QueryDelegationWarningStatusResponse, error)
}

func RegisterMsgServer(s grpc.Server, srv MsgServer)		{ s.RegisterService(&Msg_serviceDesc, srv) }
func RegisterQueryServer(s grpc.Server, srv QueryServer)	{ s.RegisterService(&Query_serviceDesc, srv) }

type serviceCall func(context.Context, interface{}, interface{}) (interface{}, error)

var Msg_serviceDesc = grpcgo.ServiceDesc{
	ServiceName:	"l1.aetrastakingpolicy.v1.Msg",
	HandlerType:	(*MsgServer)(nil),
	Methods: []grpcgo.MethodDesc{
		methodDesc("UpdateStakingPolicyParams", serviceHandler("UpdateStakingPolicyParams", func() interface{} { return new(MsgUpdateStakingPolicyParams) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(MsgServer).UpdateStakingPolicyParams(ctx, req.(*MsgUpdateStakingPolicyParams))
		})),
		methodDesc("RegisterValidatorIdentity", serviceHandler("RegisterValidatorIdentity", func() interface{} { return new(MsgRegisterValidatorIdentity) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(MsgServer).RegisterValidatorIdentity(ctx, req.(*MsgRegisterValidatorIdentity))
		})),
		methodDesc("AcknowledgeConcentrationWarning", serviceHandler("AcknowledgeConcentrationWarning", func() interface{} { return new(MsgAcknowledgeConcentrationWarning) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(MsgServer).AcknowledgeConcentrationWarning(ctx, req.(*MsgAcknowledgeConcentrationWarning))
		})),
	},
	Metadata:	"l1/aetrastakingpolicy/v1/tx.proto",
}

var Query_serviceDesc = grpcgo.ServiceDesc{
	ServiceName:	"l1.aetrastakingpolicy.v1.Query",
	HandlerType:	(*QueryServer)(nil),
	Methods: []grpcgo.MethodDesc{
		methodDesc("Params", serviceHandler("Params", func() interface{} { return new(QueryParamsRequest) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(QueryServer).Params(ctx, req.(*QueryParamsRequest))
		})),
		methodDesc("ValidatorEffectivePower", serviceHandler("ValidatorEffectivePower", func() interface{} { return new(QueryValidatorEffectivePowerRequest) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(QueryServer).ValidatorEffectivePower(ctx, req.(*QueryValidatorEffectivePowerRequest))
		})),
		methodDesc("ValidatorStake", serviceHandler("ValidatorStake", func() interface{} { return new(QueryValidatorStakeRequest) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(QueryServer).ValidatorStake(ctx, req.(*QueryValidatorStakeRequest))
		})),
		methodDesc("TopNConcentration", serviceHandler("TopNConcentration", func() interface{} { return new(QueryTopNConcentrationRequest) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(QueryServer).TopNConcentration(ctx, req.(*QueryTopNConcentrationRequest))
		})),
		methodDesc("ValidatorRewardMultiplier", serviceHandler("ValidatorRewardMultiplier", func() interface{} { return new(QueryValidatorRewardMultiplierRequest) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(QueryServer).ValidatorRewardMultiplier(ctx, req.(*QueryValidatorRewardMultiplierRequest))
		})),
		methodDesc("DelegationWarningStatus", serviceHandler("DelegationWarningStatus", func() interface{} { return new(QueryDelegationWarningStatusRequest) }, func(ctx context.Context, srv interface{}, req interface{}) (interface{}, error) {
			return srv.(QueryServer).DelegationWarningStatus(ctx, req.(*QueryDelegationWarningStatusRequest))
		})),
	},
	Metadata:	"l1/aetrastakingpolicy/v1/query.proto",
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
	gogoproto.RegisterFile("l1/aetrastakingpolicy/v1/tx.proto", buildServiceFileDescriptor("l1/aetrastakingpolicy/v1/tx.proto", "l1.aetrastakingpolicy.v1", "Msg", []string{"MsgUpdateStakingPolicyParams", "MsgUpdateStakingPolicyParamsResponse", "MsgRegisterValidatorIdentity", "MsgRegisterValidatorIdentityResponse", "MsgAcknowledgeConcentrationWarning", "MsgAcknowledgeConcentrationWarningResponse"}, [][3]string{{"UpdateStakingPolicyParams", "MsgUpdateStakingPolicyParams", "MsgUpdateStakingPolicyParamsResponse"}, {"RegisterValidatorIdentity", "MsgRegisterValidatorIdentity", "MsgRegisterValidatorIdentityResponse"}, {"AcknowledgeConcentrationWarning", "MsgAcknowledgeConcentrationWarning", "MsgAcknowledgeConcentrationWarningResponse"}}))
	gogoproto.RegisterFile("l1/aetrastakingpolicy/v1/query.proto", buildServiceFileDescriptor("l1/aetrastakingpolicy/v1/query.proto", "l1.aetrastakingpolicy.v1", "Query", []string{"QueryParamsRequest", "QueryParamsResponse", "QueryValidatorEffectivePowerRequest", "QueryValidatorEffectivePowerResponse", "QueryValidatorStakeRequest", "QueryValidatorStakeResponse", "QueryTopNConcentrationRequest", "QueryTopNConcentrationResponse", "QueryValidatorRewardMultiplierRequest", "QueryValidatorRewardMultiplierResponse", "QueryDelegationWarningStatusRequest", "QueryDelegationWarningStatusResponse"}, [][3]string{{"Params", "QueryParamsRequest", "QueryParamsResponse"}, {"ValidatorEffectivePower", "QueryValidatorEffectivePowerRequest", "QueryValidatorEffectivePowerResponse"}, {"ValidatorStake", "QueryValidatorStakeRequest", "QueryValidatorStakeResponse"}, {"TopNConcentration", "QueryTopNConcentrationRequest", "QueryTopNConcentrationResponse"}, {"ValidatorRewardMultiplier", "QueryValidatorRewardMultiplierRequest", "QueryValidatorRewardMultiplierResponse"}, {"DelegationWarningStatus", "QueryDelegationWarningStatusRequest", "QueryDelegationWarningStatusResponse"}}))
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
	gogoproto.RegisterType((*MsgUpdateStakingPolicyParams)(nil), "l1.aetrastakingpolicy.v1.MsgUpdateStakingPolicyParams")
	gogoproto.RegisterType((*MsgUpdateStakingPolicyParamsResponse)(nil), "l1.aetrastakingpolicy.v1.MsgUpdateStakingPolicyParamsResponse")
	gogoproto.RegisterType((*MsgRegisterValidatorIdentity)(nil), "l1.aetrastakingpolicy.v1.MsgRegisterValidatorIdentity")
	gogoproto.RegisterType((*MsgRegisterValidatorIdentityResponse)(nil), "l1.aetrastakingpolicy.v1.MsgRegisterValidatorIdentityResponse")
	gogoproto.RegisterType((*MsgAcknowledgeConcentrationWarning)(nil), "l1.aetrastakingpolicy.v1.MsgAcknowledgeConcentrationWarning")
	gogoproto.RegisterType((*MsgAcknowledgeConcentrationWarningResponse)(nil), "l1.aetrastakingpolicy.v1.MsgAcknowledgeConcentrationWarningResponse")
}

func (m *MsgUpdateStakingPolicyParams) Reset()		{ *m = MsgUpdateStakingPolicyParams{} }
func (m *MsgRegisterValidatorIdentity) Reset()		{ *m = MsgRegisterValidatorIdentity{} }
func (m *MsgAcknowledgeConcentrationWarning) Reset()	{ *m = MsgAcknowledgeConcentrationWarning{} }
func (m *MsgUpdateStakingPolicyParamsResponse) Reset()	{ *m = MsgUpdateStakingPolicyParamsResponse{} }
func (m *MsgRegisterValidatorIdentityResponse) Reset()	{ *m = MsgRegisterValidatorIdentityResponse{} }
func (m *MsgAcknowledgeConcentrationWarningResponse) Reset() {
	*m = MsgAcknowledgeConcentrationWarningResponse{}
}

func (m *MsgUpdateStakingPolicyParams) String() string		{ return gogoproto.CompactTextString(m) }
func (m *MsgRegisterValidatorIdentity) String() string		{ return gogoproto.CompactTextString(m) }
func (m *MsgAcknowledgeConcentrationWarning) String() string	{ return gogoproto.CompactTextString(m) }
func (m *MsgUpdateStakingPolicyParamsResponse) String() string	{ return gogoproto.CompactTextString(m) }
func (m *MsgRegisterValidatorIdentityResponse) String() string	{ return gogoproto.CompactTextString(m) }
func (m *MsgAcknowledgeConcentrationWarningResponse) String() string {
	return gogoproto.CompactTextString(m)
}

func (*MsgUpdateStakingPolicyParams) ProtoMessage()			{}
func (*MsgRegisterValidatorIdentity) ProtoMessage()			{}
func (*MsgAcknowledgeConcentrationWarning) ProtoMessage()		{}
func (*MsgUpdateStakingPolicyParamsResponse) ProtoMessage()		{}
func (*MsgRegisterValidatorIdentityResponse) ProtoMessage()		{}
func (*MsgAcknowledgeConcentrationWarningResponse) ProtoMessage()	{}

func (m *QueryParamsRequest) Reset()			{ *m = QueryParamsRequest{} }
func (m *QueryParamsResponse) Reset()			{ *m = QueryParamsResponse{} }
func (m *QueryValidatorEffectivePowerRequest) Reset()	{ *m = QueryValidatorEffectivePowerRequest{} }
func (m *QueryValidatorEffectivePowerResponse) Reset()	{ *m = QueryValidatorEffectivePowerResponse{} }
func (m *QueryValidatorStakeRequest) Reset()		{ *m = QueryValidatorStakeRequest{} }
func (m *QueryValidatorStakeResponse) Reset()		{ *m = QueryValidatorStakeResponse{} }
func (m *QueryTopNConcentrationRequest) Reset()		{ *m = QueryTopNConcentrationRequest{} }
func (m *QueryTopNConcentrationResponse) Reset()	{ *m = QueryTopNConcentrationResponse{} }
func (m *QueryValidatorRewardMultiplierRequest) Reset()	{ *m = QueryValidatorRewardMultiplierRequest{} }
func (m *QueryValidatorRewardMultiplierResponse) Reset() {
	*m = QueryValidatorRewardMultiplierResponse{}
}
func (m *QueryDelegationWarningStatusRequest) Reset()		{ *m = QueryDelegationWarningStatusRequest{} }
func (m *QueryDelegationWarningStatusResponse) Reset()		{ *m = QueryDelegationWarningStatusResponse{} }
func (m *QueryParamsRequest) String() string			{ return gogoproto.CompactTextString(m) }
func (m *QueryParamsResponse) String() string			{ return gogoproto.CompactTextString(m) }
func (m *QueryValidatorEffectivePowerRequest) String() string	{ return gogoproto.CompactTextString(m) }
func (m *QueryValidatorEffectivePowerResponse) String() string	{ return gogoproto.CompactTextString(m) }
func (m *QueryValidatorStakeRequest) String() string		{ return gogoproto.CompactTextString(m) }
func (m *QueryValidatorStakeResponse) String() string		{ return gogoproto.CompactTextString(m) }
func (m *QueryTopNConcentrationRequest) String() string		{ return gogoproto.CompactTextString(m) }
func (m *QueryTopNConcentrationResponse) String() string	{ return gogoproto.CompactTextString(m) }
func (m *QueryValidatorRewardMultiplierRequest) String() string {
	return gogoproto.CompactTextString(m)
}
func (m *QueryValidatorRewardMultiplierResponse) String() string {
	return gogoproto.CompactTextString(m)
}
func (m *QueryDelegationWarningStatusRequest) String() string	{ return gogoproto.CompactTextString(m) }
func (m *QueryDelegationWarningStatusResponse) String() string	{ return gogoproto.CompactTextString(m) }
func (*QueryParamsRequest) ProtoMessage()			{}
func (*QueryParamsResponse) ProtoMessage()			{}
func (*QueryValidatorEffectivePowerRequest) ProtoMessage()	{}
func (*QueryValidatorEffectivePowerResponse) ProtoMessage()	{}
func (*QueryValidatorStakeRequest) ProtoMessage()		{}
func (*QueryValidatorStakeResponse) ProtoMessage()		{}
func (*QueryTopNConcentrationRequest) ProtoMessage()		{}
func (*QueryTopNConcentrationResponse) ProtoMessage()		{}
func (*QueryValidatorRewardMultiplierRequest) ProtoMessage()	{}
func (*QueryValidatorRewardMultiplierResponse) ProtoMessage()	{}
func (*QueryDelegationWarningStatusRequest) ProtoMessage()	{}
func (*QueryDelegationWarningStatusResponse) ProtoMessage()	{}
