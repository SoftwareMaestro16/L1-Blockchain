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

const MsgActivateAccountTypeURL = "/l1.nativeaccount.v1.MsgActivateAccount"

type MsgActivateAccountResponse struct {
	AddressUser   string `protobuf:"bytes,1,opt,name=address_user,json=addressUser,proto3" json:"address_user,omitempty"`
	AddressRaw    string `protobuf:"bytes,2,opt,name=address_raw,json=addressRaw,proto3" json:"address_raw,omitempty"`
	AccountNumber uint64 `protobuf:"varint,3,opt,name=account_number,json=accountNumber,proto3" json:"account_number,omitempty"`
	Sequence      uint64 `protobuf:"varint,4,opt,name=sequence,proto3" json:"sequence,omitempty"`
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

type MsgServer interface {
	ActivateAccount(context.Context, *MsgActivateAccount) (*MsgActivateAccountResponse, error)
}

type UnimplementedMsgServer struct{}

func (UnimplementedMsgServer) ActivateAccount(context.Context, *MsgActivateAccount) (*MsgActivateAccountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ActivateAccount not implemented")
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

var Msg_serviceDesc = _Msg_serviceDesc

var _Msg_serviceDesc = grpcgo.ServiceDesc{
	ServiceName: "l1.nativeaccount.v1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []grpcgo.MethodDesc{
		{
			MethodName: "ActivateAccount",
			Handler:    _Msg_ActivateAccount_Handler,
		},
	},
	Streams:  []grpcgo.StreamDesc{},
	Metadata: "l1/nativeaccount/v1/tx.proto",
}

func init() {
	gogoproto.RegisterType((*MsgActivateAccount)(nil), "l1.nativeaccount.v1.MsgActivateAccount")
	gogoproto.RegisterType((*MsgActivateAccountResponse)(nil), "l1.nativeaccount.v1.MsgActivateAccountResponse")
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
		},
		Service: []*descriptorpb.ServiceDescriptorProto{
			{
				Name: descriptorString("Msg"),
				Method: []*descriptorpb.MethodDescriptorProto{
					{
						Name:       descriptorString("ActivateAccount"),
						InputType:  descriptorString(".l1.nativeaccount.v1.MsgActivateAccount"),
						OutputType: descriptorString(".l1.nativeaccount.v1.MsgActivateAccountResponse"),
					},
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

func descriptorField(name string, number int32, typ descriptorpb.FieldDescriptorProto_Type) *descriptorpb.FieldDescriptorProto {
	label := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL
	return &descriptorpb.FieldDescriptorProto{
		Name:   descriptorString(name),
		Number: descriptorInt32(number),
		Label:  &label,
		Type:   &typ,
	}
}

func descriptorString(value string) *string {
	return &value
}

func descriptorInt32(value int32) *int32 {
	return &value
}
