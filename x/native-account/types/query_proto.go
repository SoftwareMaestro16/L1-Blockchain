package types

import (
	"bytes"
	"compress/gzip"

	proto2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

var fileDescriptorNativeAccountQuery = buildNativeAccountQueryFileDescriptor()

func buildNativeAccountQueryFileDescriptor() []byte {
	const path = "l1/nativeaccount/v1/query.proto"
	fd := &descriptorpb.FileDescriptorProto{
		Name:		descriptorString(path),
		Package:	descriptorString("l1.nativeaccount.v1"),
		Syntax:		descriptorString("proto3"),
		Options: &descriptorpb.FileOptions{
			GoPackage: descriptorString("github.com/sovereign-l1/l1/x/native-account/types"),
		},
		MessageType: []*descriptorpb.DescriptorProto{
			{Name: descriptorString("QueryAccountRequest"), Field: []*descriptorpb.FieldDescriptorProto{descriptorField("address", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING)}},
			{Name: descriptorString("QueryAccountResponse"), Field: []*descriptorpb.FieldDescriptorProto{
				descriptorField("found", 1, descriptorpb.FieldDescriptorProto_TYPE_BOOL),
				descriptorField("virtual", 2, descriptorpb.FieldDescriptorProto_TYPE_BOOL),
				descriptorField("address_user", 3, descriptorpb.FieldDescriptorProto_TYPE_STRING),
				descriptorField("address_raw", 4, descriptorpb.FieldDescriptorProto_TYPE_STRING),
				descriptorField("status", 5, descriptorpb.FieldDescriptorProto_TYPE_STRING),
				descriptorField("account_json", 6, descriptorpb.FieldDescriptorProto_TYPE_STRING),
			}},
			{Name: descriptorString("QueryAccountByRawRequest"), Field: []*descriptorpb.FieldDescriptorProto{descriptorField("address_raw", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING)}},
			{Name: descriptorString("QueryVirtualAccountRequest"), Field: []*descriptorpb.FieldDescriptorProto{descriptorField("address_user", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING)}},
			virtualAccountResponseDescriptor("QueryVirtualAccountResponse", false),
			{Name: descriptorString("QueryParamsRequest")},
			{Name: descriptorString("QueryParamsResponse"), Field: []*descriptorpb.FieldDescriptorProto{
				descriptorField("default_query_limit", 1, descriptorpb.FieldDescriptorProto_TYPE_UINT64),
				descriptorField("max_query_limit", 2, descriptorpb.FieldDescriptorProto_TYPE_UINT64),
			}},
			{Name: descriptorString("QueryAccountStatusRequest"), Field: []*descriptorpb.FieldDescriptorProto{descriptorField("address", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING)}},
			virtualAccountResponseDescriptor("QueryAccountStatusResponse", true),
		},
		Service: []*descriptorpb.ServiceDescriptorProto{
			{
				Name:	descriptorString("Query"),
				Method: []*descriptorpb.MethodDescriptorProto{
					queryMethodDescriptor("Account", "QueryAccountRequest", "QueryAccountResponse"),
					queryMethodDescriptor("AccountByRaw", "QueryAccountByRawRequest", "QueryAccountResponse"),
					queryMethodDescriptor("VirtualAccount", "QueryVirtualAccountRequest", "QueryVirtualAccountResponse"),
					queryMethodDescriptor("Params", "QueryParamsRequest", "QueryParamsResponse"),
					queryMethodDescriptor("AccountStatus", "QueryAccountStatusRequest", "QueryAccountStatusResponse"),
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

func queryMethodDescriptor(name, input, output string) *descriptorpb.MethodDescriptorProto {
	return &descriptorpb.MethodDescriptorProto{
		Name:		descriptorString(name),
		InputType:	descriptorString(".l1.nativeaccount.v1." + input),
		OutputType:	descriptorString(".l1.nativeaccount.v1." + output),
	}
}

func virtualAccountResponseDescriptor(name string, includeDebt bool) *descriptorpb.DescriptorProto {
	fields := []*descriptorpb.FieldDescriptorProto{
		descriptorField("address_user", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING),
		descriptorField("address_raw", 2, descriptorpb.FieldDescriptorProto_TYPE_STRING),
		descriptorField("status", 3, descriptorpb.FieldDescriptorProto_TYPE_STRING),
		descriptorField("persistent", 4, descriptorpb.FieldDescriptorProto_TYPE_BOOL),
		descriptorField("storage_rent_active", 5, descriptorpb.FieldDescriptorProto_TYPE_BOOL),
	}
	if includeDebt {
		fields = append(fields, descriptorField("storage_rent_debt", 6, descriptorpb.FieldDescriptorProto_TYPE_UINT64))
	}
	return &descriptorpb.DescriptorProto{Name: descriptorString(name), Field: fields}
}
