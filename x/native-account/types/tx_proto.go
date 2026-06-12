package types

import (
	"bytes"
	"compress/gzip"

	proto2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

var fileDescriptorNativeAccountTx = buildNativeAccountTxFileDescriptor()

func buildNativeAccountTxFileDescriptor() []byte {
	const path = "l1/nativeaccount/v1/tx.proto"
	fd := &descriptorpb.FileDescriptorProto{
		Name:		descriptorString(path),
		Package:	descriptorString("l1.nativeaccount.v1"),
		Syntax:		descriptorString("proto3"),
		Options: &descriptorpb.FileOptions{
			GoPackage: descriptorString("github.com/sovereign-l1/l1/x/native-account/types"),
		},
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name:	descriptorString("MsgActivateAccount"),
				Field: []*descriptorpb.FieldDescriptorProto{
					descriptorField("address_user", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING),
					descriptorField("address_raw", 2, descriptorpb.FieldDescriptorProto_TYPE_STRING),
					descriptorField("public_key_type", 3, descriptorpb.FieldDescriptorProto_TYPE_STRING),
					descriptorField("public_key_hex", 4, descriptorpb.FieldDescriptorProto_TYPE_STRING),
					descriptorField("fee_paid", 5, descriptorpb.FieldDescriptorProto_TYPE_UINT64),
				},
			},
			{
				Name:	descriptorString("MsgActivateAccountResponse"),
				Field: []*descriptorpb.FieldDescriptorProto{
					descriptorField("address_user", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING),
					descriptorField("address_raw", 2, descriptorpb.FieldDescriptorProto_TYPE_STRING),
					descriptorField("account_number", 3, descriptorpb.FieldDescriptorProto_TYPE_UINT64),
					descriptorField("sequence", 4, descriptorpb.FieldDescriptorProto_TYPE_UINT64),
				},
			},
			{
				Name:	descriptorString("MsgUpdateAuthPolicy"),
				Field: []*descriptorpb.FieldDescriptorProto{
					descriptorField("account_user", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING),
					descriptorMessageField("new_auth_policy", 2, ".l1.nativeaccount.v1.AuthPolicy", false),
					descriptorRepeatedField("signers", 3, descriptorpb.FieldDescriptorProto_TYPE_STRING),
					descriptorField("current_height", 4, descriptorpb.FieldDescriptorProto_TYPE_UINT64),
				},
			},
			responseDescriptor("MsgUpdateAuthPolicyResponse", false),
			{
				Name:	descriptorString("MsgRotateKey"),
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
				Name:	descriptorString("MsgPayStorageDebt"),
				Field: []*descriptorpb.FieldDescriptorProto{
					descriptorField("account_user", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING),
					descriptorField("amount", 2, descriptorpb.FieldDescriptorProto_TYPE_UINT64),
					descriptorRepeatedField("signers", 3, descriptorpb.FieldDescriptorProto_TYPE_STRING),
					descriptorField("current_height", 4, descriptorpb.FieldDescriptorProto_TYPE_UINT64),
				},
			},
			responseDescriptor("MsgPayStorageDebtResponse", true),
			{
				Name:	descriptorString("MsgUnfreezeAccount"),
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
				Name:	descriptorString("MsgUpdateAccountMetadata"),
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
				Name:	descriptorString("Msg"),
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
		Name:		descriptorString(name),
		InputType:	descriptorString(".l1.nativeaccount.v1." + input),
		OutputType:	descriptorString(".l1.nativeaccount.v1." + output),
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
		Name:	descriptorString(name),
		Field: []*descriptorpb.FieldDescriptorProto{
			descriptorField("account_user", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING),
			descriptorRepeatedField("signers", 2, descriptorpb.FieldDescriptorProto_TYPE_STRING),
			descriptorField("current_height", 3, descriptorpb.FieldDescriptorProto_TYPE_UINT64),
		},
	}
}

func authKeyDescriptor() *descriptorpb.DescriptorProto {
	return &descriptorpb.DescriptorProto{
		Name:	descriptorString("AuthKey"),
		Field: []*descriptorpb.FieldDescriptorProto{
			descriptorField("id", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING),
			descriptorField("public_key", 2, descriptorpb.FieldDescriptorProto_TYPE_STRING),
			descriptorField("role", 3, descriptorpb.FieldDescriptorProto_TYPE_STRING),
		},
	}
}

func authWeightDescriptor() *descriptorpb.DescriptorProto {
	return &descriptorpb.DescriptorProto{
		Name:	descriptorString("AuthWeight"),
		Field: []*descriptorpb.FieldDescriptorProto{
			descriptorField("key_id", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING),
			descriptorField("weight", 2, descriptorpb.FieldDescriptorProto_TYPE_UINT64),
		},
	}
}

func recoveryPolicyDescriptor() *descriptorpb.DescriptorProto {
	return &descriptorpb.DescriptorProto{
		Name:	descriptorString("RecoveryPolicy"),
		Field: []*descriptorpb.FieldDescriptorProto{
			descriptorRepeatedField("keys", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING),
			descriptorField("threshold", 2, descriptorpb.FieldDescriptorProto_TYPE_UINT64),
			descriptorField("timelock_end_height", 3, descriptorpb.FieldDescriptorProto_TYPE_UINT64),
		},
	}
}

func timelockPolicyDescriptor() *descriptorpb.DescriptorProto {
	return &descriptorpb.DescriptorProto{
		Name:	descriptorString("TimelockPolicy"),
		Field: []*descriptorpb.FieldDescriptorProto{
			descriptorField("auth_policy_update_end_height", 1, descriptorpb.FieldDescriptorProto_TYPE_UINT64),
			descriptorField("recovery_end_height", 2, descriptorpb.FieldDescriptorProto_TYPE_UINT64),
		},
	}
}

func spendingLimitDescriptor() *descriptorpb.DescriptorProto {
	return &descriptorpb.DescriptorProto{
		Name:	descriptorString("SpendingLimit"),
		Field: []*descriptorpb.FieldDescriptorProto{
			descriptorField("operation", 1, descriptorpb.FieldDescriptorProto_TYPE_STRING),
			descriptorField("max_amount", 2, descriptorpb.FieldDescriptorProto_TYPE_UINT64),
		},
	}
}

func authPolicyDescriptor() *descriptorpb.DescriptorProto {
	return &descriptorpb.DescriptorProto{
		Name:	descriptorString("AuthPolicy"),
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
		Name:	descriptorString("AccountMetadata"),
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
		Name:	descriptorString(name),
		Number:	descriptorInt32(number),
		Label:	&label,
		Type:	&typ,
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
