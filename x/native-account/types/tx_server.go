package types

import (
	"context"

	"github.com/cosmos/gogoproto/grpc"
	gogoproto "github.com/cosmos/gogoproto/proto"
	grpcgo "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MsgServer interface {
	ActivateAccount(context.Context, *MsgActivateAccount) (*MsgActivateAccountResponse, error)
	UpdateAuthPolicy(context.Context, *MsgUpdateAuthPolicy) (*MsgUpdateAuthPolicyResponse, error)
	RotateKey(context.Context, *MsgRotateKey) (*MsgRotateKeyResponse, error)
	RecoverAccount(context.Context, *MsgRecoverAccount) (*MsgRecoverAccountResponse, error)
	FreezeAccount(context.Context, *MsgFreezeAccount) (*MsgFreezeAccountResponse, error)
	PayStorageDebt(context.Context, *MsgPayStorageDebt) (*MsgPayStorageDebtResponse, error)
	UnfreezeAccount(context.Context, *MsgUnfreezeAccount) (*MsgUnfreezeAccountResponse, error)
	UpdateAccountMetadata(context.Context, *MsgUpdateAccountMetadata) (*MsgUpdateAccountMetadataResponse, error)
}

type UnimplementedMsgServer struct{}

func (UnimplementedMsgServer) ActivateAccount(context.Context, *MsgActivateAccount) (*MsgActivateAccountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ActivateAccount not implemented")
}
func (UnimplementedMsgServer) UpdateAuthPolicy(context.Context, *MsgUpdateAuthPolicy) (*MsgUpdateAuthPolicyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateAuthPolicy not implemented")
}
func (UnimplementedMsgServer) RotateKey(context.Context, *MsgRotateKey) (*MsgRotateKeyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RotateKey not implemented")
}
func (UnimplementedMsgServer) RecoverAccount(context.Context, *MsgRecoverAccount) (*MsgRecoverAccountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RecoverAccount not implemented")
}
func (UnimplementedMsgServer) FreezeAccount(context.Context, *MsgFreezeAccount) (*MsgFreezeAccountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FreezeAccount not implemented")
}
func (UnimplementedMsgServer) PayStorageDebt(context.Context, *MsgPayStorageDebt) (*MsgPayStorageDebtResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PayStorageDebt not implemented")
}
func (UnimplementedMsgServer) UnfreezeAccount(context.Context, *MsgUnfreezeAccount) (*MsgUnfreezeAccountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UnfreezeAccount not implemented")
}
func (UnimplementedMsgServer) UpdateAccountMetadata(context.Context, *MsgUpdateAccountMetadata) (*MsgUpdateAccountMetadataResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateAccountMetadata not implemented")
}

func RegisterMsgServer(s grpc.Server, srv MsgServer) {
	s.RegisterService(&_Msg_serviceDesc, srv)
}

func _Msg_ActivateAccount_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgActivateAccount)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).ActivateAccount(ctx, in)
	}
	info := &grpcgo.UnaryServerInfo{Server: srv, FullMethod: "/l1.nativeaccount.v1.Msg/ActivateAccount"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).ActivateAccount(ctx, req.(*MsgActivateAccount))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_UpdateAuthPolicy_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUpdateAuthPolicy)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).UpdateAuthPolicy(ctx, in)
	}
	info := &grpcgo.UnaryServerInfo{Server: srv, FullMethod: "/l1.nativeaccount.v1.Msg/UpdateAuthPolicy"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).UpdateAuthPolicy(ctx, req.(*MsgUpdateAuthPolicy))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_RotateKey_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgRotateKey)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).RotateKey(ctx, in)
	}
	info := &grpcgo.UnaryServerInfo{Server: srv, FullMethod: "/l1.nativeaccount.v1.Msg/RotateKey"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).RotateKey(ctx, req.(*MsgRotateKey))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_RecoverAccount_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgRecoverAccount)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).RecoverAccount(ctx, in)
	}
	info := &grpcgo.UnaryServerInfo{Server: srv, FullMethod: "/l1.nativeaccount.v1.Msg/RecoverAccount"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).RecoverAccount(ctx, req.(*MsgRecoverAccount))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_FreezeAccount_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgFreezeAccount)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).FreezeAccount(ctx, in)
	}
	info := &grpcgo.UnaryServerInfo{Server: srv, FullMethod: "/l1.nativeaccount.v1.Msg/FreezeAccount"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).FreezeAccount(ctx, req.(*MsgFreezeAccount))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_PayStorageDebt_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgPayStorageDebt)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).PayStorageDebt(ctx, in)
	}
	info := &grpcgo.UnaryServerInfo{Server: srv, FullMethod: "/l1.nativeaccount.v1.Msg/PayStorageDebt"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).PayStorageDebt(ctx, req.(*MsgPayStorageDebt))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_UnfreezeAccount_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUnfreezeAccount)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).UnfreezeAccount(ctx, in)
	}
	info := &grpcgo.UnaryServerInfo{Server: srv, FullMethod: "/l1.nativeaccount.v1.Msg/UnfreezeAccount"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).UnfreezeAccount(ctx, req.(*MsgUnfreezeAccount))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_UpdateAccountMetadata_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpcgo.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUpdateAccountMetadata)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).UpdateAccountMetadata(ctx, in)
	}
	info := &grpcgo.UnaryServerInfo{Server: srv, FullMethod: "/l1.nativeaccount.v1.Msg/UpdateAccountMetadata"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).UpdateAccountMetadata(ctx, req.(*MsgUpdateAccountMetadata))
	}
	return interceptor(ctx, in, info, handler)
}

var Msg_serviceDesc = _Msg_serviceDesc

var _Msg_serviceDesc = grpcgo.ServiceDesc{
	ServiceName:	"l1.nativeaccount.v1.Msg",
	HandlerType:	(*MsgServer)(nil),
	Methods: []grpcgo.MethodDesc{
		{MethodName: "ActivateAccount", Handler: _Msg_ActivateAccount_Handler},
		{MethodName: "UpdateAuthPolicy", Handler: _Msg_UpdateAuthPolicy_Handler},
		{MethodName: "RotateKey", Handler: _Msg_RotateKey_Handler},
		{MethodName: "RecoverAccount", Handler: _Msg_RecoverAccount_Handler},
		{MethodName: "FreezeAccount", Handler: _Msg_FreezeAccount_Handler},
		{MethodName: "PayStorageDebt", Handler: _Msg_PayStorageDebt_Handler},
		{MethodName: "UnfreezeAccount", Handler: _Msg_UnfreezeAccount_Handler},
		{MethodName: "UpdateAccountMetadata", Handler: _Msg_UpdateAccountMetadata_Handler},
	},
	Streams:	[]grpcgo.StreamDesc{},
	Metadata:	"l1/nativeaccount/v1/tx.proto",
}

func init() {
	gogoproto.RegisterType((*MsgActivateAccount)(nil), "l1.nativeaccount.v1.MsgActivateAccount")
	gogoproto.RegisterType((*MsgActivateAccountResponse)(nil), "l1.nativeaccount.v1.MsgActivateAccountResponse")
	gogoproto.RegisterType((*MsgUpdateAuthPolicy)(nil), "l1.nativeaccount.v1.MsgUpdateAuthPolicy")
	gogoproto.RegisterType((*MsgUpdateAuthPolicyResponse)(nil), "l1.nativeaccount.v1.MsgUpdateAuthPolicyResponse")
	gogoproto.RegisterType((*MsgRotateKey)(nil), "l1.nativeaccount.v1.MsgRotateKey")
	gogoproto.RegisterType((*MsgRotateKeyResponse)(nil), "l1.nativeaccount.v1.MsgRotateKeyResponse")
	gogoproto.RegisterType((*MsgRecoverAccount)(nil), "l1.nativeaccount.v1.MsgRecoverAccount")
	gogoproto.RegisterType((*MsgRecoverAccountResponse)(nil), "l1.nativeaccount.v1.MsgRecoverAccountResponse")
	gogoproto.RegisterType((*MsgFreezeAccount)(nil), "l1.nativeaccount.v1.MsgFreezeAccount")
	gogoproto.RegisterType((*MsgFreezeAccountResponse)(nil), "l1.nativeaccount.v1.MsgFreezeAccountResponse")
	gogoproto.RegisterType((*MsgPayStorageDebt)(nil), "l1.nativeaccount.v1.MsgPayStorageDebt")
	gogoproto.RegisterType((*MsgPayStorageDebtResponse)(nil), "l1.nativeaccount.v1.MsgPayStorageDebtResponse")
	gogoproto.RegisterType((*MsgUnfreezeAccount)(nil), "l1.nativeaccount.v1.MsgUnfreezeAccount")
	gogoproto.RegisterType((*MsgUnfreezeAccountResponse)(nil), "l1.nativeaccount.v1.MsgUnfreezeAccountResponse")
	gogoproto.RegisterType((*MsgUpdateAccountMetadata)(nil), "l1.nativeaccount.v1.MsgUpdateAccountMetadata")
	gogoproto.RegisterType((*MsgUpdateAccountMetadataResponse)(nil), "l1.nativeaccount.v1.MsgUpdateAccountMetadataResponse")
	gogoproto.RegisterFile("l1/nativeaccount/v1/tx.proto", fileDescriptorNativeAccountTx)
}
