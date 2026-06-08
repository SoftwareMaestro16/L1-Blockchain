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

const (
	MsgActivateAccountTypeURL       = "/l1.nativeaccount.v1.MsgActivateAccount"
	MsgUpdateAuthPolicyTypeURL      = "/l1.nativeaccount.v1.MsgUpdateAuthPolicy"
	MsgRotateKeyTypeURL             = "/l1.nativeaccount.v1.MsgRotateKey"
	MsgRecoverAccountTypeURL        = "/l1.nativeaccount.v1.MsgRecoverAccount"
	MsgFreezeAccountTypeURL         = "/l1.nativeaccount.v1.MsgFreezeAccount"
	MsgPayStorageDebtTypeURL        = "/l1.nativeaccount.v1.MsgPayStorageDebt"
	MsgUnfreezeAccountTypeURL       = "/l1.nativeaccount.v1.MsgUnfreezeAccount"
	MsgUpdateAccountMetadataTypeURL = "/l1.nativeaccount.v1.MsgUpdateAccountMetadata"
)

type MsgActivateAccountResponse struct {
	AddressUser   string `protobuf:"bytes,1,opt,name=address_user,json=addressUser,proto3" json:"address_user,omitempty"`
	AddressRaw    string `protobuf:"bytes,2,opt,name=address_raw,json=addressRaw,proto3" json:"address_raw,omitempty"`
	AccountNumber uint64 `protobuf:"varint,3,opt,name=account_number,json=accountNumber,proto3" json:"account_number,omitempty"`
	Sequence      uint64 `protobuf:"varint,4,opt,name=sequence,proto3" json:"sequence,omitempty"`
}

type MsgUpdateAuthPolicyResponse struct {
	AddressUser string `protobuf:"bytes,1,opt,name=address_user,json=addressUser,proto3" json:"address_user,omitempty"`
	Sequence    uint64 `protobuf:"varint,2,opt,name=sequence,proto3" json:"sequence,omitempty"`
	Status      string `protobuf:"bytes,3,opt,name=status,proto3" json:"status,omitempty"`
}

type MsgRotateKeyResponse struct {
	AddressUser string `protobuf:"bytes,1,opt,name=address_user,json=addressUser,proto3" json:"address_user,omitempty"`
	Sequence    uint64 `protobuf:"varint,2,opt,name=sequence,proto3" json:"sequence,omitempty"`
	Status      string `protobuf:"bytes,3,opt,name=status,proto3" json:"status,omitempty"`
}

type MsgRecoverAccountResponse struct {
	AddressUser string `protobuf:"bytes,1,opt,name=address_user,json=addressUser,proto3" json:"address_user,omitempty"`
	Sequence    uint64 `protobuf:"varint,2,opt,name=sequence,proto3" json:"sequence,omitempty"`
	Status      string `protobuf:"bytes,3,opt,name=status,proto3" json:"status,omitempty"`
}

type MsgFreezeAccountResponse struct {
	AddressUser string `protobuf:"bytes,1,opt,name=address_user,json=addressUser,proto3" json:"address_user,omitempty"`
	Sequence    uint64 `protobuf:"varint,2,opt,name=sequence,proto3" json:"sequence,omitempty"`
	Status      string `protobuf:"bytes,3,opt,name=status,proto3" json:"status,omitempty"`
}

type MsgPayStorageDebtResponse struct {
	AddressUser     string `protobuf:"bytes,1,opt,name=address_user,json=addressUser,proto3" json:"address_user,omitempty"`
	StorageRentDebt uint64 `protobuf:"varint,2,opt,name=storage_rent_debt,json=storageRentDebt,proto3" json:"storage_rent_debt,omitempty"`
	Status          string `protobuf:"bytes,3,opt,name=status,proto3" json:"status,omitempty"`
}

type MsgUnfreezeAccountResponse struct {
	AddressUser     string `protobuf:"bytes,1,opt,name=address_user,json=addressUser,proto3" json:"address_user,omitempty"`
	StorageRentDebt uint64 `protobuf:"varint,2,opt,name=storage_rent_debt,json=storageRentDebt,proto3" json:"storage_rent_debt,omitempty"`
	Status          string `protobuf:"bytes,3,opt,name=status,proto3" json:"status,omitempty"`
}

type MsgUpdateAccountMetadataResponse struct {
	AddressUser string `protobuf:"bytes,1,opt,name=address_user,json=addressUser,proto3" json:"address_user,omitempty"`
	Sequence    uint64 `protobuf:"varint,2,opt,name=sequence,proto3" json:"sequence,omitempty"`
	Status      string `protobuf:"bytes,3,opt,name=status,proto3" json:"status,omitempty"`
}

func (m *MsgActivateAccount) Reset()         { *m = MsgActivateAccount{} }
func (m *MsgActivateAccount) String() string { return gogoproto.CompactTextString(m) }
func (*MsgActivateAccount) ProtoMessage()    {}
func (*MsgActivateAccount) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountTx, []int{0}
}

func (m *MsgActivateAccountResponse) Reset()         { *m = MsgActivateAccountResponse{} }
func (m *MsgActivateAccountResponse) String() string { return gogoproto.CompactTextString(m) }
func (*MsgActivateAccountResponse) ProtoMessage()    {}
func (*MsgActivateAccountResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountTx, []int{1}
}

func (m *MsgUpdateAuthPolicy) Reset()         { *m = MsgUpdateAuthPolicy{} }
func (m *MsgUpdateAuthPolicy) String() string { return gogoproto.CompactTextString(m) }
func (*MsgUpdateAuthPolicy) ProtoMessage()    {}
func (*MsgUpdateAuthPolicy) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountTx, []int{2}
}

func (m *MsgUpdateAuthPolicyResponse) Reset()         { *m = MsgUpdateAuthPolicyResponse{} }
func (m *MsgUpdateAuthPolicyResponse) String() string { return gogoproto.CompactTextString(m) }
func (*MsgUpdateAuthPolicyResponse) ProtoMessage()    {}
func (*MsgUpdateAuthPolicyResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountTx, []int{3}
}

func (m *MsgRotateKey) Reset()         { *m = MsgRotateKey{} }
func (m *MsgRotateKey) String() string { return gogoproto.CompactTextString(m) }
func (*MsgRotateKey) ProtoMessage()    {}
func (*MsgRotateKey) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountTx, []int{4}
}

func (m *MsgRotateKeyResponse) Reset()         { *m = MsgRotateKeyResponse{} }
func (m *MsgRotateKeyResponse) String() string { return gogoproto.CompactTextString(m) }
func (*MsgRotateKeyResponse) ProtoMessage()    {}
func (*MsgRotateKeyResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountTx, []int{5}
}

func (m *MsgRecoverAccount) Reset()         { *m = MsgRecoverAccount{} }
func (m *MsgRecoverAccount) String() string { return gogoproto.CompactTextString(m) }
func (*MsgRecoverAccount) ProtoMessage()    {}
func (*MsgRecoverAccount) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountTx, []int{6}
}

func (m *MsgRecoverAccountResponse) Reset()         { *m = MsgRecoverAccountResponse{} }
func (m *MsgRecoverAccountResponse) String() string { return gogoproto.CompactTextString(m) }
func (*MsgRecoverAccountResponse) ProtoMessage()    {}
func (*MsgRecoverAccountResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountTx, []int{7}
}

func (m *MsgFreezeAccount) Reset()         { *m = MsgFreezeAccount{} }
func (m *MsgFreezeAccount) String() string { return gogoproto.CompactTextString(m) }
func (*MsgFreezeAccount) ProtoMessage()    {}
func (*MsgFreezeAccount) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountTx, []int{8}
}

func (m *MsgFreezeAccountResponse) Reset()         { *m = MsgFreezeAccountResponse{} }
func (m *MsgFreezeAccountResponse) String() string { return gogoproto.CompactTextString(m) }
func (*MsgFreezeAccountResponse) ProtoMessage()    {}
func (*MsgFreezeAccountResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountTx, []int{9}
}

func (m *MsgPayStorageDebt) Reset()         { *m = MsgPayStorageDebt{} }
func (m *MsgPayStorageDebt) String() string { return gogoproto.CompactTextString(m) }
func (*MsgPayStorageDebt) ProtoMessage()    {}
func (*MsgPayStorageDebt) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountTx, []int{10}
}

func (m *MsgPayStorageDebtResponse) Reset()         { *m = MsgPayStorageDebtResponse{} }
func (m *MsgPayStorageDebtResponse) String() string { return gogoproto.CompactTextString(m) }
func (*MsgPayStorageDebtResponse) ProtoMessage()    {}
func (*MsgPayStorageDebtResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountTx, []int{11}
}

func (m *MsgUnfreezeAccount) Reset()         { *m = MsgUnfreezeAccount{} }
func (m *MsgUnfreezeAccount) String() string { return gogoproto.CompactTextString(m) }
func (*MsgUnfreezeAccount) ProtoMessage()    {}
func (*MsgUnfreezeAccount) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountTx, []int{12}
}

func (m *MsgUnfreezeAccountResponse) Reset()         { *m = MsgUnfreezeAccountResponse{} }
func (m *MsgUnfreezeAccountResponse) String() string { return gogoproto.CompactTextString(m) }
func (*MsgUnfreezeAccountResponse) ProtoMessage()    {}
func (*MsgUnfreezeAccountResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountTx, []int{13}
}

func (m *MsgUpdateAccountMetadata) Reset()         { *m = MsgUpdateAccountMetadata{} }
func (m *MsgUpdateAccountMetadata) String() string { return gogoproto.CompactTextString(m) }
func (*MsgUpdateAccountMetadata) ProtoMessage()    {}
func (*MsgUpdateAccountMetadata) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountTx, []int{14}
}

func (m *MsgUpdateAccountMetadataResponse) Reset()         { *m = MsgUpdateAccountMetadataResponse{} }
func (m *MsgUpdateAccountMetadataResponse) String() string { return gogoproto.CompactTextString(m) }
func (*MsgUpdateAccountMetadataResponse) ProtoMessage()    {}
func (*MsgUpdateAccountMetadataResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountTx, []int{15}
}

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
	info := &grpcgo.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/l1.nativeaccount.v1.Msg/ActivateAccount",
	}
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
	ServiceName: "l1.nativeaccount.v1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []grpcgo.MethodDesc{
		{
			MethodName: "ActivateAccount",
			Handler:    _Msg_ActivateAccount_Handler,
		},
		{MethodName: "UpdateAuthPolicy", Handler: _Msg_UpdateAuthPolicy_Handler},
		{MethodName: "RotateKey", Handler: _Msg_RotateKey_Handler},
		{MethodName: "RecoverAccount", Handler: _Msg_RecoverAccount_Handler},
		{MethodName: "FreezeAccount", Handler: _Msg_FreezeAccount_Handler},
		{MethodName: "PayStorageDebt", Handler: _Msg_PayStorageDebt_Handler},
		{MethodName: "UnfreezeAccount", Handler: _Msg_UnfreezeAccount_Handler},
		{MethodName: "UpdateAccountMetadata", Handler: _Msg_UpdateAccountMetadata_Handler},
	},
	Streams:  []grpcgo.StreamDesc{},
	Metadata: "l1/nativeaccount/v1/tx.proto",
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

var fileDescriptorNativeAccountTx = buildNativeAccountTxFileDescriptor()

func buildNativeAccountTxFileDescriptor() []byte {
	const path = "l1/nativeaccount/v1/tx.proto"
	fd := &descriptorpb.FileDescriptorProto{
		Name:    descriptorString(path),
		Package: descriptorString("l1.nativeaccount.v1"),
		Syntax:  descriptorString("proto3"),
		Options: &descriptorpb.FileOptions{
			GoPackage: descriptorString("github.com/sovereign-l1/l1/x/native-account/types"),
		},
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: descriptorString("MsgActivateAccount"),
				Field: []*descriptorpb.FieldDescriptorProto{
					descriptorField("address_user", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING),
					descriptorField("address_raw", 2, descriptorpb.FieldDescriptorProto_TYPE_STRING),
					descriptorField("public_key_type", 3, descriptorpb.FieldDescriptorProto_TYPE_STRING),
					descriptorField("public_key_hex", 4, descriptorpb.FieldDescriptorProto_TYPE_STRING),
					descriptorField("fee_paid", 5, descriptorpb.FieldDescriptorProto_TYPE_UINT64),
				},
			},
			{
				Name: descriptorString("MsgActivateAccountResponse"),
				Field: []*descriptorpb.FieldDescriptorProto{
					descriptorField("address_user", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING),
					descriptorField("address_raw", 2, descriptorpb.FieldDescriptorProto_TYPE_STRING),
					descriptorField("account_number", 3, descriptorpb.FieldDescriptorProto_TYPE_UINT64),
					descriptorField("sequence", 4, descriptorpb.FieldDescriptorProto_TYPE_UINT64),
				},
			},
			{
				Name: descriptorString("MsgUpdateAuthPolicy"),
				Field: []*descriptorpb.FieldDescriptorProto{
					descriptorField("account_user", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING),
					descriptorMessageField("new_auth_policy", 2, ".l1.nativeaccount.v1.AuthPolicy", false),
					descriptorRepeatedField("signers", 3, descriptorpb.FieldDescriptorProto_TYPE_STRING),
					descriptorField("current_height", 4, descriptorpb.FieldDescriptorProto_TYPE_UINT64),
				},
			},
			responseDescriptor("MsgUpdateAuthPolicyResponse", false),
			{
				Name: descriptorString("MsgRotateKey"),
				Field: []*descriptorpb.FieldDescriptorProto{
					descriptorField("account_user", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING),
					descriptorField("old_key_id", 2, descriptorpb.FieldDescriptorProto_TYPE_STRING),
					descriptorMessageField("new_key", 3, ".l1.nativeaccount.v1.AuthKey", false),
					descriptorRepeatedField("signers", 4, descriptorpb.FieldDescriptorProto_TYPE_STRING),
					descriptorField("current_height", 5, descriptorpb.FieldDescriptorProto_TYPE_UINT64),
				},
			},
			responseDescriptor("MsgRotateKeyResponse", false),
			messageWithSignersDescriptor("MsgRecoverAccount"),
			responseDescriptor("MsgRecoverAccountResponse", false),
			messageWithSignersDescriptor("MsgFreezeAccount"),
			responseDescriptor("MsgFreezeAccountResponse", false),
			{
				Name: descriptorString("MsgPayStorageDebt"),
				Field: []*descriptorpb.FieldDescriptorProto{
					descriptorField("account_user", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING),
					descriptorField("amount", 2, descriptorpb.FieldDescriptorProto_TYPE_UINT64),
					descriptorRepeatedField("signers", 3, descriptorpb.FieldDescriptorProto_TYPE_STRING),
					descriptorField("current_height", 4, descriptorpb.FieldDescriptorProto_TYPE_UINT64),
				},
			},
			responseDescriptor("MsgPayStorageDebtResponse", true),
			{
				Name: descriptorString("MsgUnfreezeAccount"),
				Field: []*descriptorpb.FieldDescriptorProto{
					descriptorField("account_user", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING),
					descriptorRepeatedField("signers", 2, descriptorpb.FieldDescriptorProto_TYPE_STRING),
					descriptorField("current_height", 3, descriptorpb.FieldDescriptorProto_TYPE_UINT64),
					descriptorField("storage_debt_paid", 4, descriptorpb.FieldDescriptorProto_TYPE_BOOL),
					descriptorField("other_freeze_reason", 5, descriptorpb.FieldDescriptorProto_TYPE_BOOL),
				},
			},
			responseDescriptor("MsgUnfreezeAccountResponse", true),
			{
				Name: descriptorString("MsgUpdateAccountMetadata"),
				Field: []*descriptorpb.FieldDescriptorProto{
					descriptorField("account_user", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING),
					descriptorMessageField("metadata", 2, ".l1.nativeaccount.v1.AccountMetadata", false),
					descriptorRepeatedField("signers", 3, descriptorpb.FieldDescriptorProto_TYPE_STRING),
					descriptorField("current_height", 4, descriptorpb.FieldDescriptorProto_TYPE_UINT64),
				},
			},
			responseDescriptor("MsgUpdateAccountMetadataResponse", false),
			authKeyDescriptor(),
			authWeightDescriptor(),
			recoveryPolicyDescriptor(),
			timelockPolicyDescriptor(),
			spendingLimitDescriptor(),
			authPolicyDescriptor(),
			accountMetadataDescriptor(),
		},
		Service: []*descriptorpb.ServiceDescriptorProto{
			{
				Name: descriptorString("Msg"),
				Method: []*descriptorpb.MethodDescriptorProto{
					methodDescriptor("ActivateAccount", "MsgActivateAccount", "MsgActivateAccountResponse"),
					methodDescriptor("UpdateAuthPolicy", "MsgUpdateAuthPolicy", "MsgUpdateAuthPolicyResponse"),
					methodDescriptor("RotateKey", "MsgRotateKey", "MsgRotateKeyResponse"),
					methodDescriptor("RecoverAccount", "MsgRecoverAccount", "MsgRecoverAccountResponse"),
					methodDescriptor("FreezeAccount", "MsgFreezeAccount", "MsgFreezeAccountResponse"),
					methodDescriptor("PayStorageDebt", "MsgPayStorageDebt", "MsgPayStorageDebtResponse"),
					methodDescriptor("UnfreezeAccount", "MsgUnfreezeAccount", "MsgUnfreezeAccountResponse"),
					methodDescriptor("UpdateAccountMetadata", "MsgUpdateAccountMetadata", "MsgUpdateAccountMetadataResponse"),
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

func methodDescriptor(name, input, output string) *descriptorpb.MethodDescriptorProto {
	return &descriptorpb.MethodDescriptorProto{
		Name:       descriptorString(name),
		InputType:  descriptorString(".l1.nativeaccount.v1." + input),
		OutputType: descriptorString(".l1.nativeaccount.v1." + output),
	}
}

func responseDescriptor(name string, includeDebt bool) *descriptorpb.DescriptorProto {
	fields := []*descriptorpb.FieldDescriptorProto{
		descriptorField("address_user", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING),
	}
	if includeDebt {
		fields = append(fields,
			descriptorField("storage_rent_debt", 2, descriptorpb.FieldDescriptorProto_TYPE_UINT64),
			descriptorField("status", 3, descriptorpb.FieldDescriptorProto_TYPE_STRING),
		)
	} else {
		fields = append(fields,
			descriptorField("sequence", 2, descriptorpb.FieldDescriptorProto_TYPE_UINT64),
			descriptorField("status", 3, descriptorpb.FieldDescriptorProto_TYPE_STRING),
		)
	}
	return &descriptorpb.DescriptorProto{Name: descriptorString(name), Field: fields}
}

func messageWithSignersDescriptor(name string) *descriptorpb.DescriptorProto {
	return &descriptorpb.DescriptorProto{
		Name: descriptorString(name),
		Field: []*descriptorpb.FieldDescriptorProto{
			descriptorField("account_user", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING),
			descriptorRepeatedField("signers", 2, descriptorpb.FieldDescriptorProto_TYPE_STRING),
			descriptorField("current_height", 3, descriptorpb.FieldDescriptorProto_TYPE_UINT64),
		},
	}
}

func authKeyDescriptor() *descriptorpb.DescriptorProto {
	return &descriptorpb.DescriptorProto{
		Name: descriptorString("AuthKey"),
		Field: []*descriptorpb.FieldDescriptorProto{
			descriptorField("id", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING),
			descriptorField("public_key", 2, descriptorpb.FieldDescriptorProto_TYPE_STRING),
			descriptorField("role", 3, descriptorpb.FieldDescriptorProto_TYPE_STRING),
		},
	}
}

func authWeightDescriptor() *descriptorpb.DescriptorProto {
	return &descriptorpb.DescriptorProto{
		Name: descriptorString("AuthWeight"),
		Field: []*descriptorpb.FieldDescriptorProto{
			descriptorField("key_id", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING),
			descriptorField("weight", 2, descriptorpb.FieldDescriptorProto_TYPE_UINT64),
		},
	}
}

func recoveryPolicyDescriptor() *descriptorpb.DescriptorProto {
	return &descriptorpb.DescriptorProto{
		Name: descriptorString("RecoveryPolicy"),
		Field: []*descriptorpb.FieldDescriptorProto{
			descriptorRepeatedField("keys", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING),
			descriptorField("threshold", 2, descriptorpb.FieldDescriptorProto_TYPE_UINT64),
			descriptorField("timelock_end_height", 3, descriptorpb.FieldDescriptorProto_TYPE_UINT64),
		},
	}
}

func timelockPolicyDescriptor() *descriptorpb.DescriptorProto {
	return &descriptorpb.DescriptorProto{
		Name: descriptorString("TimelockPolicy"),
		Field: []*descriptorpb.FieldDescriptorProto{
			descriptorField("auth_policy_update_end_height", 1, descriptorpb.FieldDescriptorProto_TYPE_UINT64),
			descriptorField("recovery_end_height", 2, descriptorpb.FieldDescriptorProto_TYPE_UINT64),
		},
	}
}

func spendingLimitDescriptor() *descriptorpb.DescriptorProto {
	return &descriptorpb.DescriptorProto{
		Name: descriptorString("SpendingLimit"),
		Field: []*descriptorpb.FieldDescriptorProto{
			descriptorField("operation", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING),
			descriptorField("max_amount", 2, descriptorpb.FieldDescriptorProto_TYPE_UINT64),
		},
	}
}

func authPolicyDescriptor() *descriptorpb.DescriptorProto {
	return &descriptorpb.DescriptorProto{
		Name: descriptorString("AuthPolicy"),
		Field: []*descriptorpb.FieldDescriptorProto{
			descriptorField("version", 1, descriptorpb.FieldDescriptorProto_TYPE_UINT64),
			descriptorField("mode", 2, descriptorpb.FieldDescriptorProto_TYPE_STRING),
			descriptorMessageField("keys", 3, ".l1.nativeaccount.v1.AuthKey", true),
			descriptorField("threshold", 4, descriptorpb.FieldDescriptorProto_TYPE_UINT64),
			descriptorMessageField("weights", 5, ".l1.nativeaccount.v1.AuthWeight", true),
			descriptorMessageField("recovery_policy", 6, ".l1.nativeaccount.v1.RecoveryPolicy", false),
			descriptorMessageField("timelock", 7, ".l1.nativeaccount.v1.TimelockPolicy", false),
			descriptorMessageField("spending_limits", 8, ".l1.nativeaccount.v1.SpendingLimit", true),
		},
	}
}

func accountMetadataDescriptor() *descriptorpb.DescriptorProto {
	return &descriptorpb.DescriptorProto{
		Name: descriptorString("AccountMetadata"),
		Field: []*descriptorpb.FieldDescriptorProto{
			descriptorField("metadata_hash", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING),
			descriptorField("display_name_hash", 2, descriptorpb.FieldDescriptorProto_TYPE_STRING),
			descriptorField("domain_alias", 3, descriptorpb.FieldDescriptorProto_TYPE_STRING),
			descriptorField("created_height", 4, descriptorpb.FieldDescriptorProto_TYPE_UINT64),
		},
	}
}

func descriptorField(name string, number int32, typ descriptorpb.FieldDescriptorProto_Type) *descriptorpb.FieldDescriptorProto {
	label := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL
	return &descriptorpb.FieldDescriptorProto{
		Name:   descriptorString(name),
		Number: descriptorInt32(number),
		Label:  &label,
		Type:   &typ,
	}
}

func descriptorRepeatedField(name string, number int32, typ descriptorpb.FieldDescriptorProto_Type) *descriptorpb.FieldDescriptorProto {
	field := descriptorField(name, number, typ)
	label := descriptorpb.FieldDescriptorProto_LABEL_REPEATED
	field.Label = &label
	return field
}

func descriptorMessageField(name string, number int32, typeName string, repeated bool) *descriptorpb.FieldDescriptorProto {
	field := descriptorField(name, number, descriptorpb.FieldDescriptorProto_TYPE_MESSAGE)
	field.TypeName = descriptorString(typeName)
	if repeated {
		label := descriptorpb.FieldDescriptorProto_LABEL_REPEATED
		field.Label = &label
	}
	return field
}

func descriptorString(value string) *string {
	return &value
}

func descriptorInt32(value int32) *int32 {
	return &value
}
