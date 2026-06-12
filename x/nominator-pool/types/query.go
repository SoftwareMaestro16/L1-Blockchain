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

type QueryServer interface {
	NominatorPool(context.Context, *QueryNominatorPoolRequest) (*QueryNominatorPoolResponse, error)
	NominatorPools(context.Context, *QueryNominatorPoolsRequest) (*QueryNominatorPoolsResponse, error)
	PoolDelegator(context.Context, *QueryPoolDelegatorRequest) (*QueryPoolDelegatorResponse, error)
	PoolRewards(context.Context, *QueryPoolRewardsRequest) (*QueryPoolRewardsResponse, error)
	PoolShare(context.Context, *QueryPoolShareRequest) (*QueryPoolShareResponse, error)
	PoolAllocations(context.Context, *QueryPoolAllocationsRequest) (*QueryPoolAllocationsResponse, error)
	StakeReputation(context.Context, *QueryStakeReputationRequest) (*QueryStakeReputationResponse, error)
	AccountReputation(context.Context, *QueryAccountReputationRequest) (*QueryAccountReputationResponse, error)
	StakingRewards(context.Context, *QueryStakingRewardsRequest) (*QueryStakingRewardsResponse, error)
	StakingProof(context.Context, *QueryStakingProofRequest) (*QueryStakingProofResponse, error)
	PoolUnbondingQueue(context.Context, *QueryPoolUnbondingQueueRequest) (*QueryPoolUnbondingQueueResponse, error)
}

type UnimplementedQueryServer struct{}

func (UnimplementedQueryServer) NominatorPool(context.Context, *QueryNominatorPoolRequest) (*QueryNominatorPoolResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NominatorPool not implemented")
}
func (UnimplementedQueryServer) NominatorPools(context.Context, *QueryNominatorPoolsRequest) (*QueryNominatorPoolsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NominatorPools not implemented")
}
func (UnimplementedQueryServer) PoolDelegator(context.Context, *QueryPoolDelegatorRequest) (*QueryPoolDelegatorResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PoolDelegator not implemented")
}
func (UnimplementedQueryServer) PoolRewards(context.Context, *QueryPoolRewardsRequest) (*QueryPoolRewardsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PoolRewards not implemented")
}
func (UnimplementedQueryServer) PoolShare(context.Context, *QueryPoolShareRequest) (*QueryPoolShareResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PoolShare not implemented")
}
func (UnimplementedQueryServer) PoolAllocations(context.Context, *QueryPoolAllocationsRequest) (*QueryPoolAllocationsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PoolAllocations not implemented")
}
func (UnimplementedQueryServer) StakeReputation(context.Context, *QueryStakeReputationRequest) (*QueryStakeReputationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StakeReputation not implemented")
}
func (UnimplementedQueryServer) AccountReputation(context.Context, *QueryAccountReputationRequest) (*QueryAccountReputationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AccountReputation not implemented")
}
func (UnimplementedQueryServer) StakingRewards(context.Context, *QueryStakingRewardsRequest) (*QueryStakingRewardsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StakingRewards not implemented")
}
func (UnimplementedQueryServer) StakingProof(context.Context, *QueryStakingProofRequest) (*QueryStakingProofResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StakingProof not implemented")
}
func (UnimplementedQueryServer) PoolUnbondingQueue(context.Context, *QueryPoolUnbondingQueueRequest) (*QueryPoolUnbondingQueueResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PoolUnbondingQueue not implemented")
}

func RegisterQueryServer(s grpc.Server, srv QueryServer) {
	s.RegisterService(&Query_serviceDesc, srv)
}

var Query_serviceDesc = grpcgo.ServiceDesc{
	ServiceName:	"l1.nominatorpool.v1.Query",
	HandlerType:	(*QueryServer)(nil),
	Methods: []grpcgo.MethodDesc{
		queryMethod("NominatorPool", _Query_NominatorPool_Handler),
		queryMethod("NominatorPools", _Query_NominatorPools_Handler),
		queryMethod("PoolDelegator", _Query_PoolDelegator_Handler),
		queryMethod("PoolRewards", _Query_PoolRewards_Handler),
		queryMethod("PoolShare", _Query_PoolShare_Handler),
		queryMethod("PoolAllocations", _Query_PoolAllocations_Handler),
		queryMethod("StakeReputation", _Query_StakeReputation_Handler),
		queryMethod("AccountReputation", _Query_AccountReputation_Handler),
		queryMethod("StakingRewards", _Query_StakingRewards_Handler),
		queryMethod("StakingProof", _Query_StakingProof_Handler),
		queryMethod("PoolUnbondingQueue", _Query_PoolUnbondingQueue_Handler),
	},
	Streams:	[]grpcgo.StreamDesc{},
	Metadata:	"l1/nominatorpool/v1/query.proto",
}

func queryMethod(name string, handler grpcgo.MethodHandler) grpcgo.MethodDesc {
	return grpcgo.MethodDesc{MethodName: name, Handler: handler}
}

func _Query_NominatorPool_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
	return queryHandler(ctx, srv, dec, interceptor, "NominatorPool", new(QueryNominatorPoolRequest), func(ctx context.Context, srv QueryServer, req interface{}) (interface{}, error) {
		return srv.NominatorPool(ctx, req.(*QueryNominatorPoolRequest))
	})
}
func _Query_NominatorPools_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
	return queryHandler(ctx, srv, dec, interceptor, "NominatorPools", new(QueryNominatorPoolsRequest), func(ctx context.Context, srv QueryServer, req interface{}) (interface{}, error) {
		return srv.NominatorPools(ctx, req.(*QueryNominatorPoolsRequest))
	})
}
func _Query_PoolDelegator_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
	return queryHandler(ctx, srv, dec, interceptor, "PoolDelegator", new(QueryPoolDelegatorRequest), func(ctx context.Context, srv QueryServer, req interface{}) (interface{}, error) {
		return srv.PoolDelegator(ctx, req.(*QueryPoolDelegatorRequest))
	})
}
func _Query_PoolRewards_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
	return queryHandler(ctx, srv, dec, interceptor, "PoolRewards", new(QueryPoolRewardsRequest), func(ctx context.Context, srv QueryServer, req interface{}) (interface{}, error) {
		return srv.PoolRewards(ctx, req.(*QueryPoolRewardsRequest))
	})
}
func _Query_PoolShare_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
	return queryHandler(ctx, srv, dec, interceptor, "PoolShare", new(QueryPoolShareRequest), func(ctx context.Context, srv QueryServer, req interface{}) (interface{}, error) {
		return srv.PoolShare(ctx, req.(*QueryPoolShareRequest))
	})
}
func _Query_PoolAllocations_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
	return queryHandler(ctx, srv, dec, interceptor, "PoolAllocations", new(QueryPoolAllocationsRequest), func(ctx context.Context, srv QueryServer, req interface{}) (interface{}, error) {
		return srv.PoolAllocations(ctx, req.(*QueryPoolAllocationsRequest))
	})
}
func _Query_StakeReputation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
	return queryHandler(ctx, srv, dec, interceptor, "StakeReputation", new(QueryStakeReputationRequest), func(ctx context.Context, srv QueryServer, req interface{}) (interface{}, error) {
		return srv.StakeReputation(ctx, req.(*QueryStakeReputationRequest))
	})
}
func _Query_AccountReputation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
	return queryHandler(ctx, srv, dec, interceptor, "AccountReputation", new(QueryAccountReputationRequest), func(ctx context.Context, srv QueryServer, req interface{}) (interface{}, error) {
		return srv.AccountReputation(ctx, req.(*QueryAccountReputationRequest))
	})
}
func _Query_StakingRewards_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
	return queryHandler(ctx, srv, dec, interceptor, "StakingRewards", new(QueryStakingRewardsRequest), func(ctx context.Context, srv QueryServer, req interface{}) (interface{}, error) {
		return srv.StakingRewards(ctx, req.(*QueryStakingRewardsRequest))
	})
}
func _Query_StakingProof_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
	return queryHandler(ctx, srv, dec, interceptor, "StakingProof", new(QueryStakingProofRequest), func(ctx context.Context, srv QueryServer, req interface{}) (interface{}, error) {
		return srv.StakingProof(ctx, req.(*QueryStakingProofRequest))
	})
}
func _Query_PoolUnbondingQueue_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
	return queryHandler(ctx, srv, dec, interceptor, "PoolUnbondingQueue", new(QueryPoolUnbondingQueueRequest), func(ctx context.Context, srv QueryServer, req interface{}) (interface{}, error) {
		return srv.PoolUnbondingQueue(ctx, req.(*QueryPoolUnbondingQueueRequest))
	})
}

func queryHandler(ctx context.Context, srv interface{}, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor, method string, req interface{}, call func(context.Context, QueryServer, interface{}) (interface{}, error)) (interface{}, error) {
	if err := dec(req); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return call(ctx, srv.(QueryServer), req)
	}
	info := &grpcgo.UnaryServerInfo{Server: srv, FullMethod: "/l1.nominatorpool.v1.Query/" + method}
	handler := func(ctx context.Context, request interface{}) (interface{}, error) {
		return call(ctx, srv.(QueryServer), request)
	}
	return interceptor(ctx, req, info, handler)
}

func init() {
	registerQueryTypes()
	gogoproto.RegisterFile("l1/nominatorpool/v1/query.proto", fileDescriptorNominatorPoolQuery)
}

var queryMessageNames = []string{
	"QueryNominatorPoolRequest",
	"QueryNominatorPoolResponse",
	"QueryNominatorPoolsRequest",
	"QueryNominatorPoolsResponse",
	"QueryPoolDelegatorRequest",
	"QueryPoolDelegatorResponse",
	"QueryPoolRewardsRequest",
	"QueryPoolRewardsResponse",
	"QueryPoolShareRequest",
	"QueryPoolShareResponse",
	"QueryPoolAllocationsRequest",
	"QueryPoolAllocationsResponse",
	"QueryStakeReputationRequest",
	"QueryStakeReputationResponse",
	"QueryAccountReputationRequest",
	"QueryAccountReputationResponse",
	"QueryStakingRewardsRequest",
	"QueryStakingRewardsResponse",
	"QueryStakingProofRequest",
	"QueryStakingProofResponse",
	"QueryPoolUnbondingQueueRequest",
	"QueryPoolUnbondingQueueResponse",
}

var fileDescriptorNominatorPoolQuery = buildNominatorPoolQueryFileDescriptor()

func buildNominatorPoolQueryFileDescriptor() []byte {
	messages := make([]*descriptorpb.DescriptorProto, 0, len(queryMessageNames))
	for _, name := range queryMessageNames {
		messages = append(messages, &descriptorpb.DescriptorProto{Name: descriptorString(name)})
	}
	methods := []*descriptorpb.MethodDescriptorProto{
		queryDescriptorMethod("NominatorPool", "QueryNominatorPoolRequest", "QueryNominatorPoolResponse"),
		queryDescriptorMethod("NominatorPools", "QueryNominatorPoolsRequest", "QueryNominatorPoolsResponse"),
		queryDescriptorMethod("PoolDelegator", "QueryPoolDelegatorRequest", "QueryPoolDelegatorResponse"),
		queryDescriptorMethod("PoolRewards", "QueryPoolRewardsRequest", "QueryPoolRewardsResponse"),
		queryDescriptorMethod("PoolShare", "QueryPoolShareRequest", "QueryPoolShareResponse"),
		queryDescriptorMethod("PoolAllocations", "QueryPoolAllocationsRequest", "QueryPoolAllocationsResponse"),
		queryDescriptorMethod("StakeReputation", "QueryStakeReputationRequest", "QueryStakeReputationResponse"),
		queryDescriptorMethod("AccountReputation", "QueryAccountReputationRequest", "QueryAccountReputationResponse"),
		queryDescriptorMethod("StakingRewards", "QueryStakingRewardsRequest", "QueryStakingRewardsResponse"),
		queryDescriptorMethod("StakingProof", "QueryStakingProofRequest", "QueryStakingProofResponse"),
		queryDescriptorMethod("PoolUnbondingQueue", "QueryPoolUnbondingQueueRequest", "QueryPoolUnbondingQueueResponse"),
	}
	fd := &descriptorpb.FileDescriptorProto{
		Name:		descriptorString("l1/nominatorpool/v1/query.proto"),
		Package:	descriptorString("l1.nominatorpool.v1"),
		Syntax:		descriptorString("proto3"),
		MessageType:	messages,
		Service:	[]*descriptorpb.ServiceDescriptorProto{{Name: descriptorString("Query"), Method: methods}},
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

func queryDescriptorMethod(name, input, output string) *descriptorpb.MethodDescriptorProto {
	return &descriptorpb.MethodDescriptorProto{
		Name:		descriptorString(name),
		InputType:	descriptorString(".l1.nominatorpool.v1." + input),
		OutputType:	descriptorString(".l1.nominatorpool.v1." + output),
	}
}

func registerQueryTypes() {
	gogoproto.RegisterType((*QueryNominatorPoolRequest)(nil), "l1.nominatorpool.v1.QueryNominatorPoolRequest")
	gogoproto.RegisterType((*QueryNominatorPoolResponse)(nil), "l1.nominatorpool.v1.QueryNominatorPoolResponse")
	gogoproto.RegisterType((*QueryNominatorPoolsRequest)(nil), "l1.nominatorpool.v1.QueryNominatorPoolsRequest")
	gogoproto.RegisterType((*QueryNominatorPoolsResponse)(nil), "l1.nominatorpool.v1.QueryNominatorPoolsResponse")
	gogoproto.RegisterType((*QueryPoolDelegatorRequest)(nil), "l1.nominatorpool.v1.QueryPoolDelegatorRequest")
	gogoproto.RegisterType((*QueryPoolDelegatorResponse)(nil), "l1.nominatorpool.v1.QueryPoolDelegatorResponse")
	gogoproto.RegisterType((*QueryPoolRewardsRequest)(nil), "l1.nominatorpool.v1.QueryPoolRewardsRequest")
	gogoproto.RegisterType((*QueryPoolRewardsResponse)(nil), "l1.nominatorpool.v1.QueryPoolRewardsResponse")
	gogoproto.RegisterType((*QueryPoolShareRequest)(nil), "l1.nominatorpool.v1.QueryPoolShareRequest")
	gogoproto.RegisterType((*QueryPoolShareResponse)(nil), "l1.nominatorpool.v1.QueryPoolShareResponse")
	gogoproto.RegisterType((*QueryPoolAllocationsRequest)(nil), "l1.nominatorpool.v1.QueryPoolAllocationsRequest")
	gogoproto.RegisterType((*QueryPoolAllocationsResponse)(nil), "l1.nominatorpool.v1.QueryPoolAllocationsResponse")
	gogoproto.RegisterType((*QueryStakeReputationRequest)(nil), "l1.nominatorpool.v1.QueryStakeReputationRequest")
	gogoproto.RegisterType((*QueryStakeReputationResponse)(nil), "l1.nominatorpool.v1.QueryStakeReputationResponse")
	gogoproto.RegisterType((*QueryAccountReputationRequest)(nil), "l1.nominatorpool.v1.QueryAccountReputationRequest")
	gogoproto.RegisterType((*QueryAccountReputationResponse)(nil), "l1.nominatorpool.v1.QueryAccountReputationResponse")
	gogoproto.RegisterType((*QueryStakingRewardsRequest)(nil), "l1.nominatorpool.v1.QueryStakingRewardsRequest")
	gogoproto.RegisterType((*QueryStakingRewardsResponse)(nil), "l1.nominatorpool.v1.QueryStakingRewardsResponse")
	gogoproto.RegisterType((*QueryStakingProofRequest)(nil), "l1.nominatorpool.v1.QueryStakingProofRequest")
	gogoproto.RegisterType((*QueryStakingProofResponse)(nil), "l1.nominatorpool.v1.QueryStakingProofResponse")
	gogoproto.RegisterType((*QueryPoolUnbondingQueueRequest)(nil), "l1.nominatorpool.v1.QueryPoolUnbondingQueueRequest")
	gogoproto.RegisterType((*QueryPoolUnbondingQueueResponse)(nil), "l1.nominatorpool.v1.QueryPoolUnbondingQueueResponse")
}

func (m *QueryNominatorPoolRequest) Reset()		{ *m = QueryNominatorPoolRequest{} }
func (m *QueryNominatorPoolResponse) Reset()		{ *m = QueryNominatorPoolResponse{} }
func (m *QueryNominatorPoolsRequest) Reset()		{ *m = QueryNominatorPoolsRequest{} }
func (m *QueryNominatorPoolsResponse) Reset()		{ *m = QueryNominatorPoolsResponse{} }
func (m *QueryPoolDelegatorRequest) Reset()		{ *m = QueryPoolDelegatorRequest{} }
func (m *QueryPoolDelegatorResponse) Reset()		{ *m = QueryPoolDelegatorResponse{} }
func (m *QueryPoolRewardsRequest) Reset()		{ *m = QueryPoolRewardsRequest{} }
func (m *QueryPoolRewardsResponse) Reset()		{ *m = QueryPoolRewardsResponse{} }
func (m *QueryPoolShareRequest) Reset()			{ *m = QueryPoolShareRequest{} }
func (m *QueryPoolShareResponse) Reset()		{ *m = QueryPoolShareResponse{} }
func (m *QueryPoolAllocationsRequest) Reset()		{ *m = QueryPoolAllocationsRequest{} }
func (m *QueryPoolAllocationsResponse) Reset()		{ *m = QueryPoolAllocationsResponse{} }
func (m *QueryStakeReputationRequest) Reset()		{ *m = QueryStakeReputationRequest{} }
func (m *QueryStakeReputationResponse) Reset()		{ *m = QueryStakeReputationResponse{} }
func (m *QueryAccountReputationRequest) Reset()		{ *m = QueryAccountReputationRequest{} }
func (m *QueryAccountReputationResponse) Reset()	{ *m = QueryAccountReputationResponse{} }
func (m *QueryStakingRewardsRequest) Reset()		{ *m = QueryStakingRewardsRequest{} }
func (m *QueryStakingRewardsResponse) Reset()		{ *m = QueryStakingRewardsResponse{} }
func (m *QueryStakingProofRequest) Reset()		{ *m = QueryStakingProofRequest{} }
func (m *QueryStakingProofResponse) Reset()		{ *m = QueryStakingProofResponse{} }
func (m *QueryPoolUnbondingQueueRequest) Reset()	{ *m = QueryPoolUnbondingQueueRequest{} }
func (m *QueryPoolUnbondingQueueResponse) Reset()	{ *m = QueryPoolUnbondingQueueResponse{} }
func (m *QueryNominatorPoolRequest) String() string	{ return gogoproto.CompactTextString(m) }
func (m *QueryNominatorPoolResponse) String() string	{ return gogoproto.CompactTextString(m) }
func (m *QueryNominatorPoolsRequest) String() string	{ return gogoproto.CompactTextString(m) }
func (m *QueryNominatorPoolsResponse) String() string {
	return gogoproto.CompactTextString(m)
}
func (m *QueryPoolDelegatorRequest) String() string	{ return gogoproto.CompactTextString(m) }
func (m *QueryPoolDelegatorResponse) String() string	{ return gogoproto.CompactTextString(m) }
func (m *QueryPoolRewardsRequest) String() string	{ return gogoproto.CompactTextString(m) }
func (m *QueryPoolRewardsResponse) String() string	{ return gogoproto.CompactTextString(m) }
func (m *QueryPoolShareRequest) String() string		{ return gogoproto.CompactTextString(m) }
func (m *QueryPoolShareResponse) String() string	{ return gogoproto.CompactTextString(m) }
func (m *QueryPoolAllocationsRequest) String() string	{ return gogoproto.CompactTextString(m) }
func (m *QueryPoolAllocationsResponse) String() string {
	return gogoproto.CompactTextString(m)
}
func (m *QueryStakeReputationRequest) String() string	{ return gogoproto.CompactTextString(m) }
func (m *QueryStakeReputationResponse) String() string	{ return gogoproto.CompactTextString(m) }
func (m *QueryAccountReputationRequest) String() string {
	return gogoproto.CompactTextString(m)
}
func (m *QueryAccountReputationResponse) String() string {
	return gogoproto.CompactTextString(m)
}
func (m *QueryStakingRewardsRequest) String() string	{ return gogoproto.CompactTextString(m) }
func (m *QueryStakingRewardsResponse) String() string {
	return gogoproto.CompactTextString(m)
}
func (m *QueryStakingProofRequest) String() string	{ return gogoproto.CompactTextString(m) }
func (m *QueryStakingProofResponse) String() string	{ return gogoproto.CompactTextString(m) }
func (m *QueryPoolUnbondingQueueRequest) String() string {
	return gogoproto.CompactTextString(m)
}
func (m *QueryPoolUnbondingQueueResponse) String() string {
	return gogoproto.CompactTextString(m)
}

func (*QueryNominatorPoolRequest) ProtoMessage()	{}
func (*QueryNominatorPoolResponse) ProtoMessage()	{}
func (*QueryNominatorPoolsRequest) ProtoMessage()	{}
func (*QueryNominatorPoolsResponse) ProtoMessage()	{}
func (*QueryPoolDelegatorRequest) ProtoMessage()	{}
func (*QueryPoolDelegatorResponse) ProtoMessage()	{}
func (*QueryPoolRewardsRequest) ProtoMessage()		{}
func (*QueryPoolRewardsResponse) ProtoMessage()		{}
func (*QueryPoolShareRequest) ProtoMessage()		{}
func (*QueryPoolShareResponse) ProtoMessage()		{}
func (*QueryPoolAllocationsRequest) ProtoMessage()	{}
func (*QueryPoolAllocationsResponse) ProtoMessage()	{}
func (*QueryStakeReputationRequest) ProtoMessage()	{}
func (*QueryStakeReputationResponse) ProtoMessage()	{}
func (*QueryAccountReputationRequest) ProtoMessage()	{}
func (*QueryAccountReputationResponse) ProtoMessage()	{}
func (*QueryStakingRewardsRequest) ProtoMessage()	{}
func (*QueryStakingRewardsResponse) ProtoMessage()	{}
func (*QueryStakingProofRequest) ProtoMessage()		{}
func (*QueryStakingProofResponse) ProtoMessage()	{}
func (*QueryPoolUnbondingQueueRequest) ProtoMessage()	{}
func (*QueryPoolUnbondingQueueResponse) ProtoMessage()	{}

func (*QueryNominatorPoolRequest) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolQuery, []int{0}
}
func (*QueryNominatorPoolResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolQuery, []int{1}
}
func (*QueryNominatorPoolsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolQuery, []int{2}
}
func (*QueryNominatorPoolsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolQuery, []int{3}
}
func (*QueryPoolDelegatorRequest) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolQuery, []int{4}
}
func (*QueryPoolDelegatorResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolQuery, []int{5}
}
func (*QueryPoolRewardsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolQuery, []int{6}
}
func (*QueryPoolRewardsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolQuery, []int{7}
}
func (*QueryPoolShareRequest) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolQuery, []int{8}
}
func (*QueryPoolShareResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolQuery, []int{9}
}
func (*QueryPoolAllocationsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolQuery, []int{10}
}
func (*QueryPoolAllocationsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolQuery, []int{11}
}
func (*QueryStakeReputationRequest) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolQuery, []int{12}
}
func (*QueryStakeReputationResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolQuery, []int{13}
}
func (*QueryAccountReputationRequest) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolQuery, []int{14}
}
func (*QueryAccountReputationResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolQuery, []int{15}
}
func (*QueryStakingRewardsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolQuery, []int{16}
}
func (*QueryStakingRewardsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolQuery, []int{17}
}
func (*QueryStakingProofRequest) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolQuery, []int{18}
}
func (*QueryStakingProofResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolQuery, []int{19}
}
func (*QueryPoolUnbondingQueueRequest) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolQuery, []int{20}
}
func (*QueryPoolUnbondingQueueResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolQuery, []int{21}
}
