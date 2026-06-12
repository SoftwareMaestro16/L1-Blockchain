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

type MsgCreateNominatorPoolResponse struct {
	Pool NominatorPool `protobuf:"bytes,1,opt,name=pool,proto3" json:"pool"`
}

type MsgDepositToPoolResponse struct {
	Delegator DelegatorShare `protobuf:"bytes,1,opt,name=delegator,proto3" json:"delegator"`
}

type MsgRequestPoolWithdrawalResponse struct {
	Withdrawal PendingWithdrawal `protobuf:"bytes,1,opt,name=withdrawal,proto3" json:"withdrawal"`
}

type MsgCancelPoolWithdrawalResponse struct {
	Withdrawal PendingWithdrawal `protobuf:"bytes,1,opt,name=withdrawal,proto3" json:"withdrawal"`
}

type MsgDepositToStakingPoolResponse struct {
	PoolID		string	`protobuf:"bytes,1,opt,name=pool_id,json=poolId,proto3" json:"pool_id,omitempty"`
	OwnerAddress	string	`protobuf:"bytes,2,opt,name=owner_address,json=ownerAddress,proto3" json:"owner_address,omitempty"`
	Amount		uint64	`protobuf:"varint,3,opt,name=amount,proto3" json:"amount,omitempty"`
	Shares		uint64	`protobuf:"varint,4,opt,name=shares,proto3" json:"shares,omitempty"`
	Height		uint64	`protobuf:"varint,5,opt,name=height,proto3" json:"height,omitempty"`
	ReceiptToken	string	`protobuf:"bytes,6,opt,name=receipt_token,json=receiptToken,proto3" json:"receipt_token,omitempty"`
}

type MsgRequestPoolUnbondResponse struct {
	PoolID		string	`protobuf:"bytes,1,opt,name=pool_id,json=poolId,proto3" json:"pool_id,omitempty"`
	OwnerAddress	string	`protobuf:"bytes,2,opt,name=owner_address,json=ownerAddress,proto3" json:"owner_address,omitempty"`
	RequestID	string	`protobuf:"bytes,3,opt,name=request_id,json=requestId,proto3" json:"request_id,omitempty"`
	Shares		uint64	`protobuf:"varint,4,opt,name=shares,proto3" json:"shares,omitempty"`
	Amount		uint64	`protobuf:"varint,5,opt,name=amount,proto3" json:"amount,omitempty"`
	CompleteHeight	uint64	`protobuf:"varint,6,opt,name=complete_height,json=completeHeight,proto3" json:"complete_height,omitempty"`
}

type MsgWithdrawPoolStakeResponse struct {
	PoolID		string	`protobuf:"bytes,1,opt,name=pool_id,json=poolId,proto3" json:"pool_id,omitempty"`
	OwnerAddress	string	`protobuf:"bytes,2,opt,name=owner_address,json=ownerAddress,proto3" json:"owner_address,omitempty"`
	RequestID	string	`protobuf:"bytes,3,opt,name=request_id,json=requestId,proto3" json:"request_id,omitempty"`
	Amount		uint64	`protobuf:"varint,4,opt,name=amount,proto3" json:"amount,omitempty"`
	Height		uint64	`protobuf:"varint,5,opt,name=height,proto3" json:"height,omitempty"`
}

type MsgTopUpPoolReserveResponse struct {
	PoolID		string	`protobuf:"bytes,1,opt,name=pool_id,json=poolId,proto3" json:"pool_id,omitempty"`
	PayerAddress	string	`protobuf:"bytes,2,opt,name=payer_address,json=payerAddress,proto3" json:"payer_address,omitempty"`
	Amount		uint64	`protobuf:"varint,3,opt,name=amount,proto3" json:"amount,omitempty"`
	StorageDebtPaid	uint64	`protobuf:"varint,4,opt,name=storage_debt_paid,json=storageDebtPaid,proto3" json:"storage_debt_paid,omitempty"`
}

type MsgClaimPoolRewardsResponse struct {
	PoolID		string	`protobuf:"bytes,1,opt,name=pool_id,json=poolId,proto3" json:"pool_id,omitempty"`
	OwnerAddress	string	`protobuf:"bytes,2,opt,name=owner_address,json=ownerAddress,proto3" json:"owner_address,omitempty"`
	Amount		uint64	`protobuf:"varint,3,opt,name=amount,proto3" json:"amount,omitempty"`
	Epoch		uint64	`protobuf:"varint,4,opt,name=epoch,proto3" json:"epoch,omitempty"`
}

type MsgSyncPoolRewardsResponse struct {
	PoolUserRewards		uint64	`protobuf:"varint,1,opt,name=pool_user_rewards,json=poolUserRewards,proto3" json:"pool_user_rewards,omitempty"`
	ValidatorCommission	uint64	`protobuf:"varint,2,opt,name=validator_commission,json=validatorCommission,proto3" json:"validator_commission,omitempty"`
	PoolProtocolFee		uint64	`protobuf:"varint,3,opt,name=pool_protocol_fee,json=poolProtocolFee,proto3" json:"pool_protocol_fee,omitempty"`
	RewardIndexAfter	uint64	`protobuf:"varint,4,opt,name=reward_index_after,json=rewardIndexAfter,proto3" json:"reward_index_after,omitempty"`
}

type MsgClaimStakingRewardsResponse struct {
	RewardAmount uint64 `protobuf:"varint,1,opt,name=reward_amount,json=rewardAmount,proto3" json:"reward_amount,omitempty"`
}

type MsgClaimStakeReputationResponse struct {
	Account		string	`protobuf:"bytes,1,opt,name=account,proto3" json:"account,omitempty"`
	PoolID		string	`protobuf:"bytes,2,opt,name=pool_id,json=poolId,proto3" json:"pool_id,omitempty"`
	ReputationDelta	uint64	`protobuf:"varint,3,opt,name=reputation_delta,json=reputationDelta,proto3" json:"reputation_delta,omitempty"`
	ReputationScore	uint64	`protobuf:"varint,4,opt,name=reputation_score,json=reputationScore,proto3" json:"reputation_score,omitempty"`
}

type MsgDelegateToValidatorResponse struct{}
type MsgRegisterValidatorResponse struct {
	Validator	string	`protobuf:"bytes,1,opt,name=validator,proto3" json:"validator,omitempty"`
	Status		string	`protobuf:"bytes,2,opt,name=status,proto3" json:"status,omitempty"`
	SelfStake	uint64	`protobuf:"varint,3,opt,name=self_stake,json=selfStake,proto3" json:"self_stake,omitempty"`
	PoolStake	uint64	`protobuf:"varint,4,opt,name=pool_stake,json=poolStake,proto3" json:"pool_stake,omitempty"`
}
type MsgUpdateValidatorResponse struct {
	Validator	string	`protobuf:"bytes,1,opt,name=validator,proto3" json:"validator,omitempty"`
	Status		string	`protobuf:"bytes,2,opt,name=status,proto3" json:"status,omitempty"`
	SelfStake	uint64	`protobuf:"varint,3,opt,name=self_stake,json=selfStake,proto3" json:"self_stake,omitempty"`
	PoolStake	uint64	`protobuf:"varint,4,opt,name=pool_stake,json=poolStake,proto3" json:"pool_stake,omitempty"`
}
type MsgUpdateStakingParamsResponse struct{}

type MsgUpdatePoolCommissionResponse struct {
	Pool NominatorPool `protobuf:"bytes,1,opt,name=pool,proto3" json:"pool"`
}

type MsgChangePoolValidatorResponse struct {
	Pool NominatorPool `protobuf:"bytes,1,opt,name=pool,proto3" json:"pool"`
}

type MsgCreateOfficialLiquidStakingPoolResponse struct {
	PoolID			string	`protobuf:"bytes,1,opt,name=pool_id,json=poolId,proto3" json:"pool_id,omitempty"`
	ContractAddressUser	string	`protobuf:"bytes,2,opt,name=contract_address_user,json=contractAddressUser,proto3" json:"contract_address_user,omitempty"`
	ContractAddressRaw	string	`protobuf:"bytes,3,opt,name=contract_address_raw,json=contractAddressRaw,proto3" json:"contract_address_raw,omitempty"`
}

type MsgServer interface {
	CreateNominatorPool(context.Context, *MsgCreateNominatorPool) (*MsgCreateNominatorPoolResponse, error)
	DepositToPool(context.Context, *MsgDepositToPool) (*MsgDepositToPoolResponse, error)
	RequestPoolWithdrawal(context.Context, *MsgRequestPoolWithdrawal) (*MsgRequestPoolWithdrawalResponse, error)
	CancelPoolWithdrawal(context.Context, *MsgCancelPoolWithdrawal) (*MsgCancelPoolWithdrawalResponse, error)
	ClaimPoolRewards(context.Context, *MsgClaimPoolRewards) (*MsgClaimPoolRewardsResponse, error)
	SyncPoolRewards(context.Context, *MsgSyncPoolRewards) (*MsgSyncPoolRewardsResponse, error)
	ClaimStakingRewards(context.Context, *MsgClaimStakingRewards) (*MsgClaimStakingRewardsResponse, error)
	UpdatePoolCommission(context.Context, *MsgUpdatePoolCommission) (*MsgUpdatePoolCommissionResponse, error)
	ChangePoolValidator(context.Context, *MsgChangePoolValidator) (*MsgChangePoolValidatorResponse, error)
	DepositToStakingPool(context.Context, *MsgDepositToStakingPool) (*MsgDepositToStakingPoolResponse, error)
	RequestPoolUnbond(context.Context, *MsgRequestPoolUnbond) (*MsgRequestPoolUnbondResponse, error)
	WithdrawPoolStake(context.Context, *MsgWithdrawPoolStake) (*MsgWithdrawPoolStakeResponse, error)
	TopUpPoolReserve(context.Context, *MsgTopUpPoolReserve) (*MsgTopUpPoolReserveResponse, error)
	ClaimStakeReputation(context.Context, *MsgClaimStakeReputation) (*MsgClaimStakeReputationResponse, error)
	DelegateToValidator(context.Context, *MsgDelegateToValidator) (*MsgDelegateToValidatorResponse, error)
	RegisterValidator(context.Context, *MsgRegisterValidator) (*MsgRegisterValidatorResponse, error)
	UpdateValidator(context.Context, *MsgUpdateValidator) (*MsgUpdateValidatorResponse, error)
	UpdateStakingParams(context.Context, *MsgUpdateStakingParams) (*MsgUpdateStakingParamsResponse, error)
	CreateOfficialLiquidStakingPool(context.Context, *MsgCreateOfficialLiquidStakingPool) (*MsgCreateOfficialLiquidStakingPoolResponse, error)
}

type UnimplementedMsgServer struct{}

func (UnimplementedMsgServer) CreateNominatorPool(context.Context, *MsgCreateNominatorPool) (*MsgCreateNominatorPoolResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateNominatorPool not implemented")
}
func (UnimplementedMsgServer) DepositToPool(context.Context, *MsgDepositToPool) (*MsgDepositToPoolResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DepositToPool not implemented")
}
func (UnimplementedMsgServer) RequestPoolWithdrawal(context.Context, *MsgRequestPoolWithdrawal) (*MsgRequestPoolWithdrawalResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RequestPoolWithdrawal not implemented")
}
func (UnimplementedMsgServer) CancelPoolWithdrawal(context.Context, *MsgCancelPoolWithdrawal) (*MsgCancelPoolWithdrawalResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CancelPoolWithdrawal not implemented")
}
func (UnimplementedMsgServer) DepositToStakingPool(context.Context, *MsgDepositToStakingPool) (*MsgDepositToStakingPoolResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DepositToStakingPool not implemented")
}
func (UnimplementedMsgServer) RequestPoolUnbond(context.Context, *MsgRequestPoolUnbond) (*MsgRequestPoolUnbondResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RequestPoolUnbond not implemented")
}
func (UnimplementedMsgServer) WithdrawPoolStake(context.Context, *MsgWithdrawPoolStake) (*MsgWithdrawPoolStakeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method WithdrawPoolStake not implemented")
}
func (UnimplementedMsgServer) TopUpPoolReserve(context.Context, *MsgTopUpPoolReserve) (*MsgTopUpPoolReserveResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TopUpPoolReserve not implemented")
}
func (UnimplementedMsgServer) ClaimPoolRewards(context.Context, *MsgClaimPoolRewards) (*MsgClaimPoolRewardsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ClaimPoolRewards not implemented")
}
func (UnimplementedMsgServer) SyncPoolRewards(context.Context, *MsgSyncPoolRewards) (*MsgSyncPoolRewardsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SyncPoolRewards not implemented")
}
func (UnimplementedMsgServer) ClaimStakingRewards(context.Context, *MsgClaimStakingRewards) (*MsgClaimStakingRewardsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ClaimStakingRewards not implemented")
}
func (UnimplementedMsgServer) ClaimStakeReputation(context.Context, *MsgClaimStakeReputation) (*MsgClaimStakeReputationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ClaimStakeReputation not implemented")
}
func (UnimplementedMsgServer) DelegateToValidator(context.Context, *MsgDelegateToValidator) (*MsgDelegateToValidatorResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DelegateToValidator not implemented")
}
func (UnimplementedMsgServer) RegisterValidator(context.Context, *MsgRegisterValidator) (*MsgRegisterValidatorResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegisterValidator not implemented")
}
func (UnimplementedMsgServer) UpdateValidator(context.Context, *MsgUpdateValidator) (*MsgUpdateValidatorResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateValidator not implemented")
}
func (UnimplementedMsgServer) UpdateStakingParams(context.Context, *MsgUpdateStakingParams) (*MsgUpdateStakingParamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateStakingParams not implemented")
}
func (UnimplementedMsgServer) UpdatePoolCommission(context.Context, *MsgUpdatePoolCommission) (*MsgUpdatePoolCommissionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdatePoolCommission not implemented")
}
func (UnimplementedMsgServer) ChangePoolValidator(context.Context, *MsgChangePoolValidator) (*MsgChangePoolValidatorResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ChangePoolValidator not implemented")
}
func (UnimplementedMsgServer) CreateOfficialLiquidStakingPool(context.Context, *MsgCreateOfficialLiquidStakingPool) (*MsgCreateOfficialLiquidStakingPoolResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateOfficialLiquidStakingPool not implemented")
}

func RegisterMsgServer(s grpc.Server, srv MsgServer)	{ s.RegisterService(&Msg_serviceDesc, srv) }

var Msg_serviceDesc = grpcgo.ServiceDesc{
	ServiceName:	"l1.nominatorpool.v1.Msg",
	HandlerType:	(*MsgServer)(nil),
	Methods: []grpcgo.MethodDesc{
		msgMethod("CreateNominatorPool", _Msg_CreateNominatorPool_Handler),
		msgMethod("DepositToPool", _Msg_DepositToPool_Handler),
		msgMethod("RequestPoolWithdrawal", _Msg_RequestPoolWithdrawal_Handler),
		msgMethod("CancelPoolWithdrawal", _Msg_CancelPoolWithdrawal_Handler),
		msgMethod("ClaimPoolRewards", _Msg_ClaimPoolRewards_Handler),
		msgMethod("SyncPoolRewards", _Msg_SyncPoolRewards_Handler),
		msgMethod("ClaimStakingRewards", _Msg_ClaimStakingRewards_Handler),
		msgMethod("UpdatePoolCommission", _Msg_UpdatePoolCommission_Handler),
		msgMethod("ChangePoolValidator", _Msg_ChangePoolValidator_Handler),
		msgMethod("DepositToStakingPool", _Msg_DepositToStakingPool_Handler),
		msgMethod("RequestPoolUnbond", _Msg_RequestPoolUnbond_Handler),
		msgMethod("WithdrawPoolStake", _Msg_WithdrawPoolStake_Handler),
		msgMethod("TopUpPoolReserve", _Msg_TopUpPoolReserve_Handler),
		msgMethod("ClaimStakeReputation", _Msg_ClaimStakeReputation_Handler),
		msgMethod("DelegateToValidator", _Msg_DelegateToValidator_Handler),
		msgMethod("RegisterValidator", _Msg_RegisterValidator_Handler),
		msgMethod("UpdateValidator", _Msg_UpdateValidator_Handler),
		msgMethod("UpdateStakingParams", _Msg_UpdateStakingParams_Handler),
		msgMethod("CreateOfficialLiquidStakingPool", _Msg_CreateOfficialLiquidStakingPool_Handler),
	},
	Streams:	[]grpcgo.StreamDesc{},
	Metadata:	"l1/nominatorpool/v1/tx.proto",
}

func msgMethod(name string, handler grpcgo.MethodHandler) grpcgo.MethodDesc {
	return grpcgo.MethodDesc{MethodName: name, Handler: handler}
}

func _Msg_CreateNominatorPool_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
	return msgHandler(ctx, srv, dec, interceptor, "CreateNominatorPool", new(MsgCreateNominatorPool), func(ctx context.Context, srv MsgServer, req interface{}) (interface{}, error) {
		return srv.CreateNominatorPool(ctx, req.(*MsgCreateNominatorPool))
	})
}
func _Msg_DepositToPool_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
	return msgHandler(ctx, srv, dec, interceptor, "DepositToPool", new(MsgDepositToPool), func(ctx context.Context, srv MsgServer, req interface{}) (interface{}, error) {
		return srv.DepositToPool(ctx, req.(*MsgDepositToPool))
	})
}
func _Msg_RequestPoolWithdrawal_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
	return msgHandler(ctx, srv, dec, interceptor, "RequestPoolWithdrawal", new(MsgRequestPoolWithdrawal), func(ctx context.Context, srv MsgServer, req interface{}) (interface{}, error) {
		return srv.RequestPoolWithdrawal(ctx, req.(*MsgRequestPoolWithdrawal))
	})
}
func _Msg_CancelPoolWithdrawal_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
	return msgHandler(ctx, srv, dec, interceptor, "CancelPoolWithdrawal", new(MsgCancelPoolWithdrawal), func(ctx context.Context, srv MsgServer, req interface{}) (interface{}, error) {
		return srv.CancelPoolWithdrawal(ctx, req.(*MsgCancelPoolWithdrawal))
	})
}
func _Msg_SyncPoolRewards_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
	return msgHandler(ctx, srv, dec, interceptor, "SyncPoolRewards", new(MsgSyncPoolRewards), func(ctx context.Context, srv MsgServer, req interface{}) (interface{}, error) {
		return srv.SyncPoolRewards(ctx, req.(*MsgSyncPoolRewards))
	})
}
func _Msg_ClaimStakingRewards_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
	return msgHandler(ctx, srv, dec, interceptor, "ClaimStakingRewards", new(MsgClaimStakingRewards), func(ctx context.Context, srv MsgServer, req interface{}) (interface{}, error) {
		return srv.ClaimStakingRewards(ctx, req.(*MsgClaimStakingRewards))
	})
}
func _Msg_UpdatePoolCommission_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
	return msgHandler(ctx, srv, dec, interceptor, "UpdatePoolCommission", new(MsgUpdatePoolCommission), func(ctx context.Context, srv MsgServer, req interface{}) (interface{}, error) {
		return srv.UpdatePoolCommission(ctx, req.(*MsgUpdatePoolCommission))
	})
}
func _Msg_ChangePoolValidator_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
	return msgHandler(ctx, srv, dec, interceptor, "ChangePoolValidator", new(MsgChangePoolValidator), func(ctx context.Context, srv MsgServer, req interface{}) (interface{}, error) {
		return srv.ChangePoolValidator(ctx, req.(*MsgChangePoolValidator))
	})
}
func _Msg_DepositToStakingPool_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
	return msgHandler(ctx, srv, dec, interceptor, "DepositToStakingPool", new(MsgDepositToStakingPool), func(ctx context.Context, srv MsgServer, req interface{}) (interface{}, error) {
		return srv.DepositToStakingPool(ctx, req.(*MsgDepositToStakingPool))
	})
}
func _Msg_RequestPoolUnbond_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
	return msgHandler(ctx, srv, dec, interceptor, "RequestPoolUnbond", new(MsgRequestPoolUnbond), func(ctx context.Context, srv MsgServer, req interface{}) (interface{}, error) {
		return srv.RequestPoolUnbond(ctx, req.(*MsgRequestPoolUnbond))
	})
}
func _Msg_WithdrawPoolStake_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
	return msgHandler(ctx, srv, dec, interceptor, "WithdrawPoolStake", new(MsgWithdrawPoolStake), func(ctx context.Context, srv MsgServer, req interface{}) (interface{}, error) {
		return srv.WithdrawPoolStake(ctx, req.(*MsgWithdrawPoolStake))
	})
}
func _Msg_TopUpPoolReserve_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
	return msgHandler(ctx, srv, dec, interceptor, "TopUpPoolReserve", new(MsgTopUpPoolReserve), func(ctx context.Context, srv MsgServer, req interface{}) (interface{}, error) {
		return srv.TopUpPoolReserve(ctx, req.(*MsgTopUpPoolReserve))
	})
}
func _Msg_ClaimPoolRewards_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
	return msgHandler(ctx, srv, dec, interceptor, "ClaimPoolRewards", new(MsgClaimPoolRewards), func(ctx context.Context, srv MsgServer, req interface{}) (interface{}, error) {
		return srv.ClaimPoolRewards(ctx, req.(*MsgClaimPoolRewards))
	})
}
func _Msg_ClaimStakeReputation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
	return msgHandler(ctx, srv, dec, interceptor, "ClaimStakeReputation", new(MsgClaimStakeReputation), func(ctx context.Context, srv MsgServer, req interface{}) (interface{}, error) {
		return srv.ClaimStakeReputation(ctx, req.(*MsgClaimStakeReputation))
	})
}
func _Msg_DelegateToValidator_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
	return msgHandler(ctx, srv, dec, interceptor, "DelegateToValidator", new(MsgDelegateToValidator), func(ctx context.Context, srv MsgServer, req interface{}) (interface{}, error) {
		return srv.DelegateToValidator(ctx, req.(*MsgDelegateToValidator))
	})
}
func _Msg_RegisterValidator_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
	return msgHandler(ctx, srv, dec, interceptor, "RegisterValidator", new(MsgRegisterValidator), func(ctx context.Context, srv MsgServer, req interface{}) (interface{}, error) {
		return srv.RegisterValidator(ctx, req.(*MsgRegisterValidator))
	})
}
func _Msg_UpdateValidator_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
	return msgHandler(ctx, srv, dec, interceptor, "UpdateValidator", new(MsgUpdateValidator), func(ctx context.Context, srv MsgServer, req interface{}) (interface{}, error) {
		return srv.UpdateValidator(ctx, req.(*MsgUpdateValidator))
	})
}
func _Msg_UpdateStakingParams_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
	return msgHandler(ctx, srv, dec, interceptor, "UpdateStakingParams", new(MsgUpdateStakingParams), func(ctx context.Context, srv MsgServer, req interface{}) (interface{}, error) {
		return srv.UpdateStakingParams(ctx, req.(*MsgUpdateStakingParams))
	})
}
func _Msg_CreateOfficialLiquidStakingPool_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
	return msgHandler(ctx, srv, dec, interceptor, "CreateOfficialLiquidStakingPool", new(MsgCreateOfficialLiquidStakingPool), func(ctx context.Context, srv MsgServer, req interface{}) (interface{}, error) {
		return srv.CreateOfficialLiquidStakingPool(ctx, req.(*MsgCreateOfficialLiquidStakingPool))
	})
}

func msgHandler(ctx context.Context, srv interface{}, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor, method string, req interface{}, call func(context.Context, MsgServer, interface{}) (interface{}, error)) (interface{}, error) {
	if err := dec(req); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return call(ctx, srv.(MsgServer), req)
	}
	info := &grpcgo.UnaryServerInfo{Server: srv, FullMethod: "/l1.nominatorpool.v1.Msg/" + method}
	handler := func(ctx context.Context, request interface{}) (interface{}, error) {
		return call(ctx, srv.(MsgServer), request)
	}
	return interceptor(ctx, req, info, handler)
}

func init() {
	registerMsgTypes()
	gogoproto.RegisterFile("l1/nominatorpool/v1/tx.proto", fileDescriptorNominatorPoolTx)
}

var txMessageNames = []string{
	"MsgCreateNominatorPool",
	"MsgCreateNominatorPoolResponse",
	"MsgDepositToPool",
	"MsgDepositToPoolResponse",
	"MsgRequestPoolWithdrawal",
	"MsgRequestPoolWithdrawalResponse",
	"MsgCancelPoolWithdrawal",
	"MsgCancelPoolWithdrawalResponse",
	"MsgSyncPoolRewards",
	"MsgSyncPoolRewardsResponse",
	"MsgClaimStakingRewards",
	"MsgClaimStakingRewardsResponse",
	"MsgUpdatePoolCommission",
	"MsgUpdatePoolCommissionResponse",
	"MsgChangePoolValidator",
	"MsgChangePoolValidatorResponse",
	"MsgDepositToStakingPool",
	"MsgDepositToStakingPoolResponse",
	"MsgRequestPoolUnbond",
	"MsgRequestPoolUnbondResponse",
	"MsgWithdrawPoolStake",
	"MsgWithdrawPoolStakeResponse",
	"MsgTopUpPoolReserve",
	"MsgTopUpPoolReserveResponse",
	"MsgClaimPoolRewards",
	"MsgClaimPoolRewardsResponse",
	"MsgClaimStakeReputation",
	"MsgClaimStakeReputationResponse",
	"MsgDelegateToValidator",
	"MsgDelegateToValidatorResponse",
	"MsgRegisterValidator",
	"MsgRegisterValidatorResponse",
	"MsgUpdateValidator",
	"MsgUpdateValidatorResponse",
	"MsgUpdateStakingParams",
	"MsgUpdateStakingParamsResponse",
	"MsgCreateOfficialLiquidStakingPool",
	"MsgCreateOfficialLiquidStakingPoolResponse",
}

var fileDescriptorNominatorPoolTx = buildNominatorPoolTxFileDescriptor()

func buildNominatorPoolTxFileDescriptor() []byte {
	messages := make([]*descriptorpb.DescriptorProto, 0, len(txMessageNames))
	for _, name := range txMessageNames {
		messages = append(messages, &descriptorpb.DescriptorProto{Name: descriptorString(name)})
	}
	methods := []*descriptorpb.MethodDescriptorProto{
		txMethod("CreateNominatorPool", "MsgCreateNominatorPool", "MsgCreateNominatorPoolResponse"),
		txMethod("DepositToPool", "MsgDepositToPool", "MsgDepositToPoolResponse"),
		txMethod("RequestPoolWithdrawal", "MsgRequestPoolWithdrawal", "MsgRequestPoolWithdrawalResponse"),
		txMethod("CancelPoolWithdrawal", "MsgCancelPoolWithdrawal", "MsgCancelPoolWithdrawalResponse"),
		txMethod("SyncPoolRewards", "MsgSyncPoolRewards", "MsgSyncPoolRewardsResponse"),
		txMethod("ClaimStakingRewards", "MsgClaimStakingRewards", "MsgClaimStakingRewardsResponse"),
		txMethod("UpdatePoolCommission", "MsgUpdatePoolCommission", "MsgUpdatePoolCommissionResponse"),
		txMethod("ChangePoolValidator", "MsgChangePoolValidator", "MsgChangePoolValidatorResponse"),
		txMethod("DepositToStakingPool", "MsgDepositToStakingPool", "MsgDepositToStakingPoolResponse"),
		txMethod("RequestPoolUnbond", "MsgRequestPoolUnbond", "MsgRequestPoolUnbondResponse"),
		txMethod("WithdrawPoolStake", "MsgWithdrawPoolStake", "MsgWithdrawPoolStakeResponse"),
		txMethod("TopUpPoolReserve", "MsgTopUpPoolReserve", "MsgTopUpPoolReserveResponse"),
		txMethod("ClaimPoolRewards", "MsgClaimPoolRewards", "MsgClaimPoolRewardsResponse"),
		txMethod("ClaimStakeReputation", "MsgClaimStakeReputation", "MsgClaimStakeReputationResponse"),
		txMethod("DelegateToValidator", "MsgDelegateToValidator", "MsgDelegateToValidatorResponse"),
		txMethod("RegisterValidator", "MsgRegisterValidator", "MsgRegisterValidatorResponse"),
		txMethod("UpdateValidator", "MsgUpdateValidator", "MsgUpdateValidatorResponse"),
		txMethod("UpdateStakingParams", "MsgUpdateStakingParams", "MsgUpdateStakingParamsResponse"),
		txMethod("CreateOfficialLiquidStakingPool", "MsgCreateOfficialLiquidStakingPool", "MsgCreateOfficialLiquidStakingPoolResponse"),
	}
	fd := &descriptorpb.FileDescriptorProto{
		Name:		descriptorString("l1/nominatorpool/v1/tx.proto"),
		Package:	descriptorString("l1.nominatorpool.v1"),
		Syntax:		descriptorString("proto3"),
		MessageType:	messages,
		Service:	[]*descriptorpb.ServiceDescriptorProto{{Name: descriptorString("Msg"), Method: methods}},
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

func txMethod(name, input, output string) *descriptorpb.MethodDescriptorProto {
	return &descriptorpb.MethodDescriptorProto{
		Name:		descriptorString(name),
		InputType:	descriptorString(".l1.nominatorpool.v1." + input),
		OutputType:	descriptorString(".l1.nominatorpool.v1." + output),
	}
}

func descriptorString(value string) *string	{ return &value }

func registerMsgTypes() {
	gogoproto.RegisterType((*MsgCreateNominatorPool)(nil), "l1.nominatorpool.v1.MsgCreateNominatorPool")
	gogoproto.RegisterType((*MsgCreateNominatorPoolResponse)(nil), "l1.nominatorpool.v1.MsgCreateNominatorPoolResponse")
	gogoproto.RegisterType((*MsgDepositToPool)(nil), "l1.nominatorpool.v1.MsgDepositToPool")
	gogoproto.RegisterType((*MsgDepositToPoolResponse)(nil), "l1.nominatorpool.v1.MsgDepositToPoolResponse")
	gogoproto.RegisterType((*MsgRequestPoolWithdrawal)(nil), "l1.nominatorpool.v1.MsgRequestPoolWithdrawal")
	gogoproto.RegisterType((*MsgRequestPoolWithdrawalResponse)(nil), "l1.nominatorpool.v1.MsgRequestPoolWithdrawalResponse")
	gogoproto.RegisterType((*MsgCancelPoolWithdrawal)(nil), "l1.nominatorpool.v1.MsgCancelPoolWithdrawal")
	gogoproto.RegisterType((*MsgCancelPoolWithdrawalResponse)(nil), "l1.nominatorpool.v1.MsgCancelPoolWithdrawalResponse")
	gogoproto.RegisterType((*MsgSyncPoolRewards)(nil), "l1.nominatorpool.v1.MsgSyncPoolRewards")
	gogoproto.RegisterType((*MsgSyncPoolRewardsResponse)(nil), "l1.nominatorpool.v1.MsgSyncPoolRewardsResponse")
	gogoproto.RegisterType((*MsgClaimStakingRewards)(nil), "l1.nominatorpool.v1.MsgClaimStakingRewards")
	gogoproto.RegisterType((*MsgClaimStakingRewardsResponse)(nil), "l1.nominatorpool.v1.MsgClaimStakingRewardsResponse")
	gogoproto.RegisterType((*MsgUpdatePoolCommission)(nil), "l1.nominatorpool.v1.MsgUpdatePoolCommission")
	gogoproto.RegisterType((*MsgUpdatePoolCommissionResponse)(nil), "l1.nominatorpool.v1.MsgUpdatePoolCommissionResponse")
	gogoproto.RegisterType((*MsgChangePoolValidator)(nil), "l1.nominatorpool.v1.MsgChangePoolValidator")
	gogoproto.RegisterType((*MsgChangePoolValidatorResponse)(nil), "l1.nominatorpool.v1.MsgChangePoolValidatorResponse")
	gogoproto.RegisterType((*MsgDepositToStakingPool)(nil), "l1.nominatorpool.v1.MsgDepositToStakingPool")
	gogoproto.RegisterType((*MsgDepositToStakingPoolResponse)(nil), "l1.nominatorpool.v1.MsgDepositToStakingPoolResponse")
	gogoproto.RegisterType((*MsgRequestPoolUnbond)(nil), "l1.nominatorpool.v1.MsgRequestPoolUnbond")
	gogoproto.RegisterType((*MsgRequestPoolUnbondResponse)(nil), "l1.nominatorpool.v1.MsgRequestPoolUnbondResponse")
	gogoproto.RegisterType((*MsgWithdrawPoolStake)(nil), "l1.nominatorpool.v1.MsgWithdrawPoolStake")
	gogoproto.RegisterType((*MsgWithdrawPoolStakeResponse)(nil), "l1.nominatorpool.v1.MsgWithdrawPoolStakeResponse")
	gogoproto.RegisterType((*MsgTopUpPoolReserve)(nil), "l1.nominatorpool.v1.MsgTopUpPoolReserve")
	gogoproto.RegisterType((*MsgTopUpPoolReserveResponse)(nil), "l1.nominatorpool.v1.MsgTopUpPoolReserveResponse")
	gogoproto.RegisterType((*MsgClaimPoolRewards)(nil), "l1.nominatorpool.v1.MsgClaimPoolRewards")
	gogoproto.RegisterType((*MsgClaimPoolRewardsResponse)(nil), "l1.nominatorpool.v1.MsgClaimPoolRewardsResponse")
	gogoproto.RegisterType((*MsgClaimStakeReputation)(nil), "l1.nominatorpool.v1.MsgClaimStakeReputation")
	gogoproto.RegisterType((*MsgClaimStakeReputationResponse)(nil), "l1.nominatorpool.v1.MsgClaimStakeReputationResponse")
	gogoproto.RegisterType((*MsgDelegateToValidator)(nil), "l1.nominatorpool.v1.MsgDelegateToValidator")
	gogoproto.RegisterType((*MsgDelegateToValidatorResponse)(nil), "l1.nominatorpool.v1.MsgDelegateToValidatorResponse")
	gogoproto.RegisterType((*MsgRegisterValidator)(nil), "l1.nominatorpool.v1.MsgRegisterValidator")
	gogoproto.RegisterType((*MsgRegisterValidatorResponse)(nil), "l1.nominatorpool.v1.MsgRegisterValidatorResponse")
	gogoproto.RegisterType((*MsgUpdateValidator)(nil), "l1.nominatorpool.v1.MsgUpdateValidator")
	gogoproto.RegisterType((*MsgUpdateValidatorResponse)(nil), "l1.nominatorpool.v1.MsgUpdateValidatorResponse")
	gogoproto.RegisterType((*MsgUpdateStakingParams)(nil), "l1.nominatorpool.v1.MsgUpdateStakingParams")
	gogoproto.RegisterType((*MsgUpdateStakingParamsResponse)(nil), "l1.nominatorpool.v1.MsgUpdateStakingParamsResponse")
	gogoproto.RegisterType((*MsgCreateOfficialLiquidStakingPool)(nil), "l1.nominatorpool.v1.MsgCreateOfficialLiquidStakingPool")
	gogoproto.RegisterType((*MsgCreateOfficialLiquidStakingPoolResponse)(nil), "l1.nominatorpool.v1.MsgCreateOfficialLiquidStakingPoolResponse")
}

func (m *MsgCreateNominatorPool) Reset()		{ *m = MsgCreateNominatorPool{} }
func (m *MsgCreateNominatorPoolResponse) Reset()	{ *m = MsgCreateNominatorPoolResponse{} }
func (m *MsgDepositToPool) Reset()			{ *m = MsgDepositToPool{} }
func (m *MsgDepositToPoolResponse) Reset()		{ *m = MsgDepositToPoolResponse{} }
func (m *MsgRequestPoolWithdrawal) Reset()		{ *m = MsgRequestPoolWithdrawal{} }
func (m *MsgRequestPoolWithdrawalResponse) Reset()	{ *m = MsgRequestPoolWithdrawalResponse{} }
func (m *MsgCancelPoolWithdrawal) Reset()		{ *m = MsgCancelPoolWithdrawal{} }
func (m *MsgCancelPoolWithdrawalResponse) Reset()	{ *m = MsgCancelPoolWithdrawalResponse{} }
func (m *MsgSyncPoolRewards) Reset()			{ *m = MsgSyncPoolRewards{} }
func (m *MsgSyncPoolRewardsResponse) Reset()		{ *m = MsgSyncPoolRewardsResponse{} }
func (m *MsgClaimStakingRewards) Reset()		{ *m = MsgClaimStakingRewards{} }
func (m *MsgClaimStakingRewardsResponse) Reset()	{ *m = MsgClaimStakingRewardsResponse{} }
func (m *MsgUpdatePoolCommission) Reset()		{ *m = MsgUpdatePoolCommission{} }
func (m *MsgUpdatePoolCommissionResponse) Reset()	{ *m = MsgUpdatePoolCommissionResponse{} }
func (m *MsgChangePoolValidator) Reset()		{ *m = MsgChangePoolValidator{} }
func (m *MsgChangePoolValidatorResponse) Reset()	{ *m = MsgChangePoolValidatorResponse{} }
func (m *MsgDepositToStakingPool) Reset()		{ *m = MsgDepositToStakingPool{} }
func (m *MsgDepositToStakingPoolResponse) Reset()	{ *m = MsgDepositToStakingPoolResponse{} }
func (m *MsgRequestPoolUnbond) Reset()			{ *m = MsgRequestPoolUnbond{} }
func (m *MsgRequestPoolUnbondResponse) Reset()		{ *m = MsgRequestPoolUnbondResponse{} }
func (m *MsgWithdrawPoolStake) Reset()			{ *m = MsgWithdrawPoolStake{} }
func (m *MsgWithdrawPoolStakeResponse) Reset()		{ *m = MsgWithdrawPoolStakeResponse{} }
func (m *MsgTopUpPoolReserve) Reset()			{ *m = MsgTopUpPoolReserve{} }
func (m *MsgTopUpPoolReserveResponse) Reset()		{ *m = MsgTopUpPoolReserveResponse{} }
func (m *MsgClaimPoolRewards) Reset()			{ *m = MsgClaimPoolRewards{} }
func (m *MsgClaimPoolRewardsResponse) Reset()		{ *m = MsgClaimPoolRewardsResponse{} }
func (m *MsgClaimStakeReputation) Reset()		{ *m = MsgClaimStakeReputation{} }
func (m *MsgClaimStakeReputationResponse) Reset()	{ *m = MsgClaimStakeReputationResponse{} }
func (m *MsgDelegateToValidator) Reset()		{ *m = MsgDelegateToValidator{} }
func (m *MsgDelegateToValidatorResponse) Reset()	{ *m = MsgDelegateToValidatorResponse{} }
func (m *MsgRegisterValidator) Reset()			{ *m = MsgRegisterValidator{} }
func (m *MsgRegisterValidatorResponse) Reset()		{ *m = MsgRegisterValidatorResponse{} }
func (m *MsgUpdateValidator) Reset()			{ *m = MsgUpdateValidator{} }
func (m *MsgUpdateValidatorResponse) Reset()		{ *m = MsgUpdateValidatorResponse{} }
func (m *MsgUpdateStakingParams) Reset()		{ *m = MsgUpdateStakingParams{} }
func (m *MsgUpdateStakingParamsResponse) Reset()	{ *m = MsgUpdateStakingParamsResponse{} }
func (m *MsgCreateOfficialLiquidStakingPool) Reset()	{ *m = MsgCreateOfficialLiquidStakingPool{} }
func (m *MsgCreateOfficialLiquidStakingPoolResponse) Reset() {
	*m = MsgCreateOfficialLiquidStakingPoolResponse{}
}

func (m *MsgCreateNominatorPool) String() string		{ return gogoproto.CompactTextString(m) }
func (m *MsgCreateNominatorPoolResponse) String() string	{ return gogoproto.CompactTextString(m) }
func (m *MsgDepositToPool) String() string			{ return gogoproto.CompactTextString(m) }
func (m *MsgDepositToPoolResponse) String() string		{ return gogoproto.CompactTextString(m) }
func (m *MsgRequestPoolWithdrawal) String() string		{ return gogoproto.CompactTextString(m) }
func (m *MsgRequestPoolWithdrawalResponse) String() string	{ return gogoproto.CompactTextString(m) }
func (m *MsgCancelPoolWithdrawal) String() string		{ return gogoproto.CompactTextString(m) }
func (m *MsgCancelPoolWithdrawalResponse) String() string	{ return gogoproto.CompactTextString(m) }
func (m *MsgSyncPoolRewards) String() string			{ return gogoproto.CompactTextString(m) }
func (m *MsgSyncPoolRewardsResponse) String() string		{ return gogoproto.CompactTextString(m) }
func (m *MsgClaimStakingRewards) String() string		{ return gogoproto.CompactTextString(m) }
func (m *MsgClaimStakingRewardsResponse) String() string	{ return gogoproto.CompactTextString(m) }
func (m *MsgUpdatePoolCommission) String() string		{ return gogoproto.CompactTextString(m) }
func (m *MsgUpdatePoolCommissionResponse) String() string	{ return gogoproto.CompactTextString(m) }
func (m *MsgChangePoolValidator) String() string		{ return gogoproto.CompactTextString(m) }
func (m *MsgChangePoolValidatorResponse) String() string	{ return gogoproto.CompactTextString(m) }
func (m *MsgDepositToStakingPool) String() string		{ return gogoproto.CompactTextString(m) }
func (m *MsgDepositToStakingPoolResponse) String() string	{ return gogoproto.CompactTextString(m) }
func (m *MsgRequestPoolUnbond) String() string			{ return gogoproto.CompactTextString(m) }
func (m *MsgRequestPoolUnbondResponse) String() string		{ return gogoproto.CompactTextString(m) }
func (m *MsgWithdrawPoolStake) String() string			{ return gogoproto.CompactTextString(m) }
func (m *MsgWithdrawPoolStakeResponse) String() string		{ return gogoproto.CompactTextString(m) }
func (m *MsgTopUpPoolReserve) String() string			{ return gogoproto.CompactTextString(m) }
func (m *MsgTopUpPoolReserveResponse) String() string		{ return gogoproto.CompactTextString(m) }
func (m *MsgClaimPoolRewards) String() string			{ return gogoproto.CompactTextString(m) }
func (m *MsgClaimPoolRewardsResponse) String() string		{ return gogoproto.CompactTextString(m) }
func (m *MsgClaimStakeReputation) String() string		{ return gogoproto.CompactTextString(m) }
func (m *MsgClaimStakeReputationResponse) String() string	{ return gogoproto.CompactTextString(m) }
func (m *MsgDelegateToValidator) String() string		{ return gogoproto.CompactTextString(m) }
func (m *MsgDelegateToValidatorResponse) String() string	{ return gogoproto.CompactTextString(m) }
func (m *MsgRegisterValidator) String() string			{ return gogoproto.CompactTextString(m) }
func (m *MsgRegisterValidatorResponse) String() string		{ return gogoproto.CompactTextString(m) }
func (m *MsgUpdateValidator) String() string			{ return gogoproto.CompactTextString(m) }
func (m *MsgUpdateValidatorResponse) String() string		{ return gogoproto.CompactTextString(m) }
func (m *MsgUpdateStakingParams) String() string		{ return gogoproto.CompactTextString(m) }
func (m *MsgUpdateStakingParamsResponse) String() string	{ return gogoproto.CompactTextString(m) }
func (m *MsgCreateOfficialLiquidStakingPool) String() string	{ return gogoproto.CompactTextString(m) }
func (m *MsgCreateOfficialLiquidStakingPoolResponse) String() string {
	return gogoproto.CompactTextString(m)
}

func (*MsgCreateNominatorPool) ProtoMessage()				{}
func (*MsgCreateNominatorPoolResponse) ProtoMessage()			{}
func (*MsgDepositToPool) ProtoMessage()					{}
func (*MsgDepositToPoolResponse) ProtoMessage()				{}
func (*MsgRequestPoolWithdrawal) ProtoMessage()				{}
func (*MsgRequestPoolWithdrawalResponse) ProtoMessage()			{}
func (*MsgCancelPoolWithdrawal) ProtoMessage()				{}
func (*MsgCancelPoolWithdrawalResponse) ProtoMessage()			{}
func (*MsgSyncPoolRewards) ProtoMessage()				{}
func (*MsgSyncPoolRewardsResponse) ProtoMessage()			{}
func (*MsgClaimStakingRewards) ProtoMessage()				{}
func (*MsgClaimStakingRewardsResponse) ProtoMessage()			{}
func (*MsgUpdatePoolCommission) ProtoMessage()				{}
func (*MsgUpdatePoolCommissionResponse) ProtoMessage()			{}
func (*MsgChangePoolValidator) ProtoMessage()				{}
func (*MsgChangePoolValidatorResponse) ProtoMessage()			{}
func (*MsgDepositToStakingPool) ProtoMessage()				{}
func (*MsgDepositToStakingPoolResponse) ProtoMessage()			{}
func (*MsgRequestPoolUnbond) ProtoMessage()				{}
func (*MsgRequestPoolUnbondResponse) ProtoMessage()			{}
func (*MsgWithdrawPoolStake) ProtoMessage()				{}
func (*MsgWithdrawPoolStakeResponse) ProtoMessage()			{}
func (*MsgTopUpPoolReserve) ProtoMessage()				{}
func (*MsgTopUpPoolReserveResponse) ProtoMessage()			{}
func (*MsgClaimPoolRewards) ProtoMessage()				{}
func (*MsgClaimPoolRewardsResponse) ProtoMessage()			{}
func (*MsgClaimStakeReputation) ProtoMessage()				{}
func (*MsgClaimStakeReputationResponse) ProtoMessage()			{}
func (*MsgDelegateToValidator) ProtoMessage()				{}
func (*MsgDelegateToValidatorResponse) ProtoMessage()			{}
func (*MsgRegisterValidator) ProtoMessage()				{}
func (*MsgRegisterValidatorResponse) ProtoMessage()			{}
func (*MsgUpdateValidator) ProtoMessage()				{}
func (*MsgUpdateValidatorResponse) ProtoMessage()			{}
func (*MsgUpdateStakingParams) ProtoMessage()				{}
func (*MsgUpdateStakingParamsResponse) ProtoMessage()			{}
func (*MsgCreateOfficialLiquidStakingPool) ProtoMessage()		{}
func (*MsgCreateOfficialLiquidStakingPoolResponse) ProtoMessage()	{}

func (*MsgCreateNominatorPool) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolTx, []int{0}
}
func (*MsgCreateNominatorPoolResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolTx, []int{1}
}
func (*MsgDepositToPool) Descriptor() ([]byte, []int)	{ return fileDescriptorNominatorPoolTx, []int{2} }
func (*MsgDepositToPoolResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolTx, []int{3}
}
func (*MsgRequestPoolWithdrawal) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolTx, []int{4}
}
func (*MsgRequestPoolWithdrawalResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolTx, []int{5}
}
func (*MsgCancelPoolWithdrawal) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolTx, []int{6}
}
func (*MsgCancelPoolWithdrawalResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolTx, []int{7}
}
func (*MsgSyncPoolRewards) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolTx, []int{8}
}
func (*MsgSyncPoolRewardsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolTx, []int{9}
}
func (*MsgClaimStakingRewards) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolTx, []int{10}
}
func (*MsgClaimStakingRewardsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolTx, []int{11}
}
func (*MsgUpdatePoolCommission) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolTx, []int{12}
}
func (*MsgUpdatePoolCommissionResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolTx, []int{13}
}
func (*MsgChangePoolValidator) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolTx, []int{14}
}
func (*MsgChangePoolValidatorResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolTx, []int{15}
}
func (*MsgDepositToStakingPool) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolTx, []int{16}
}
func (*MsgDepositToStakingPoolResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolTx, []int{17}
}
func (*MsgRequestPoolUnbond) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolTx, []int{18}
}
func (*MsgRequestPoolUnbondResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolTx, []int{19}
}
func (*MsgWithdrawPoolStake) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolTx, []int{20}
}
func (*MsgWithdrawPoolStakeResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolTx, []int{21}
}
func (*MsgTopUpPoolReserve) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolTx, []int{22}
}
func (*MsgTopUpPoolReserveResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolTx, []int{23}
}
func (*MsgClaimPoolRewards) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolTx, []int{24}
}
func (*MsgClaimPoolRewardsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolTx, []int{25}
}
func (*MsgClaimStakeReputation) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolTx, []int{26}
}
func (*MsgClaimStakeReputationResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolTx, []int{27}
}
func (*MsgDelegateToValidator) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolTx, []int{28}
}
func (*MsgDelegateToValidatorResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolTx, []int{29}
}
func (*MsgRegisterValidator) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolTx, []int{30}
}
func (*MsgRegisterValidatorResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolTx, []int{31}
}
func (*MsgUpdateValidator) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolTx, []int{32}
}
func (*MsgUpdateValidatorResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolTx, []int{33}
}
func (*MsgUpdateStakingParams) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolTx, []int{34}
}
func (*MsgUpdateStakingParamsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolTx, []int{35}
}
func (*MsgCreateOfficialLiquidStakingPool) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolTx, []int{36}
}
func (*MsgCreateOfficialLiquidStakingPoolResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNominatorPoolTx, []int{37}
}
