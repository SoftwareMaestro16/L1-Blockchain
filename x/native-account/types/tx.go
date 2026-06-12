package types

import (
	gogoproto "github.com/cosmos/gogoproto/proto"
)

const (
	MsgActivateAccountTypeURL	= "/l1.nativeaccount.v1.MsgActivateAccount"
	MsgUpdateAuthPolicyTypeURL	= "/l1.nativeaccount.v1.MsgUpdateAuthPolicy"
	MsgRotateKeyTypeURL		= "/l1.nativeaccount.v1.MsgRotateKey"
	MsgRecoverAccountTypeURL	= "/l1.nativeaccount.v1.MsgRecoverAccount"
	MsgFreezeAccountTypeURL		= "/l1.nativeaccount.v1.MsgFreezeAccount"
	MsgPayStorageDebtTypeURL	= "/l1.nativeaccount.v1.MsgPayStorageDebt"
	MsgUnfreezeAccountTypeURL	= "/l1.nativeaccount.v1.MsgUnfreezeAccount"
	MsgUpdateAccountMetadataTypeURL	= "/l1.nativeaccount.v1.MsgUpdateAccountMetadata"
)

type MsgActivateAccountResponse struct {
	AddressUser	string	`protobuf:"bytes,1,opt,name=address_user,json=addressUser,proto3" json:"address_user,omitempty"`
	AddressRaw	string	`protobuf:"bytes,2,opt,name=address_raw,json=addressRaw,proto3" json:"address_raw,omitempty"`
	AccountNumber	uint64	`protobuf:"varint,3,opt,name=account_number,json=accountNumber,proto3" json:"account_number,omitempty"`
	Sequence	uint64	`protobuf:"varint,4,opt,name=sequence,proto3" json:"sequence,omitempty"`
}

type MsgUpdateAuthPolicyResponse struct {
	AddressUser	string	`protobuf:"bytes,1,opt,name=address_user,json=addressUser,proto3" json:"address_user,omitempty"`
	Sequence	uint64	`protobuf:"varint,2,opt,name=sequence,proto3" json:"sequence,omitempty"`
	Status		string	`protobuf:"bytes,3,opt,name=status,proto3" json:"status,omitempty"`
}

type MsgRotateKeyResponse struct {
	AddressUser	string	`protobuf:"bytes,1,opt,name=address_user,json=addressUser,proto3" json:"address_user,omitempty"`
	Sequence	uint64	`protobuf:"varint,2,opt,name=sequence,proto3" json:"sequence,omitempty"`
	Status		string	`protobuf:"bytes,3,opt,name=status,proto3" json:"status,omitempty"`
}

type MsgRecoverAccountResponse struct {
	AddressUser	string	`protobuf:"bytes,1,opt,name=address_user,json=addressUser,proto3" json:"address_user,omitempty"`
	Sequence	uint64	`protobuf:"varint,2,opt,name=sequence,proto3" json:"sequence,omitempty"`
	Status		string	`protobuf:"bytes,3,opt,name=status,proto3" json:"status,omitempty"`
}

type MsgFreezeAccountResponse struct {
	AddressUser	string	`protobuf:"bytes,1,opt,name=address_user,json=addressUser,proto3" json:"address_user,omitempty"`
	Sequence	uint64	`protobuf:"varint,2,opt,name=sequence,proto3" json:"sequence,omitempty"`
	Status		string	`protobuf:"bytes,3,opt,name=status,proto3" json:"status,omitempty"`
}

type MsgPayStorageDebtResponse struct {
	AddressUser	string	`protobuf:"bytes,1,opt,name=address_user,json=addressUser,proto3" json:"address_user,omitempty"`
	StorageRentDebt	uint64	`protobuf:"varint,2,opt,name=storage_rent_debt,json=storageRentDebt,proto3" json:"storage_rent_debt,omitempty"`
	Status		string	`protobuf:"bytes,3,opt,name=status,proto3" json:"status,omitempty"`
}

type MsgUnfreezeAccountResponse struct {
	AddressUser	string	`protobuf:"bytes,1,opt,name=address_user,json=addressUser,proto3" json:"address_user,omitempty"`
	StorageRentDebt	uint64	`protobuf:"varint,2,opt,name=storage_rent_debt,json=storageRentDebt,proto3" json:"storage_rent_debt,omitempty"`
	Status		string	`protobuf:"bytes,3,opt,name=status,proto3" json:"status,omitempty"`
}

type MsgUpdateAccountMetadataResponse struct {
	AddressUser	string	`protobuf:"bytes,1,opt,name=address_user,json=addressUser,proto3" json:"address_user,omitempty"`
	Sequence	uint64	`protobuf:"varint,2,opt,name=sequence,proto3" json:"sequence,omitempty"`
	Status		string	`protobuf:"bytes,3,opt,name=status,proto3" json:"status,omitempty"`
}

func (m *MsgActivateAccount) Reset()		{ *m = MsgActivateAccount{} }
func (m *MsgActivateAccount) String() string	{ return gogoproto.CompactTextString(m) }
func (*MsgActivateAccount) ProtoMessage()	{}
func (*MsgActivateAccount) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountTx, []int{0}
}

func (m *MsgActivateAccountResponse) Reset()		{ *m = MsgActivateAccountResponse{} }
func (m *MsgActivateAccountResponse) String() string	{ return gogoproto.CompactTextString(m) }
func (*MsgActivateAccountResponse) ProtoMessage()	{}
func (*MsgActivateAccountResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountTx, []int{1}
}

func (m *MsgUpdateAuthPolicy) Reset()		{ *m = MsgUpdateAuthPolicy{} }
func (m *MsgUpdateAuthPolicy) String() string	{ return gogoproto.CompactTextString(m) }
func (*MsgUpdateAuthPolicy) ProtoMessage()	{}
func (*MsgUpdateAuthPolicy) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountTx, []int{2}
}

func (m *MsgUpdateAuthPolicyResponse) Reset()		{ *m = MsgUpdateAuthPolicyResponse{} }
func (m *MsgUpdateAuthPolicyResponse) String() string	{ return gogoproto.CompactTextString(m) }
func (*MsgUpdateAuthPolicyResponse) ProtoMessage()	{}
func (*MsgUpdateAuthPolicyResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountTx, []int{3}
}

func (m *MsgRotateKey) Reset()		{ *m = MsgRotateKey{} }
func (m *MsgRotateKey) String() string	{ return gogoproto.CompactTextString(m) }
func (*MsgRotateKey) ProtoMessage()	{}
func (*MsgRotateKey) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountTx, []int{4}
}

func (m *MsgRotateKeyResponse) Reset()		{ *m = MsgRotateKeyResponse{} }
func (m *MsgRotateKeyResponse) String() string	{ return gogoproto.CompactTextString(m) }
func (*MsgRotateKeyResponse) ProtoMessage()	{}
func (*MsgRotateKeyResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountTx, []int{5}
}

func (m *MsgRecoverAccount) Reset()		{ *m = MsgRecoverAccount{} }
func (m *MsgRecoverAccount) String() string	{ return gogoproto.CompactTextString(m) }
func (*MsgRecoverAccount) ProtoMessage()	{}
func (*MsgRecoverAccount) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountTx, []int{6}
}

func (m *MsgRecoverAccountResponse) Reset()		{ *m = MsgRecoverAccountResponse{} }
func (m *MsgRecoverAccountResponse) String() string	{ return gogoproto.CompactTextString(m) }
func (*MsgRecoverAccountResponse) ProtoMessage()	{}
func (*MsgRecoverAccountResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountTx, []int{7}
}

func (m *MsgFreezeAccount) Reset()		{ *m = MsgFreezeAccount{} }
func (m *MsgFreezeAccount) String() string	{ return gogoproto.CompactTextString(m) }
func (*MsgFreezeAccount) ProtoMessage()		{}
func (*MsgFreezeAccount) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountTx, []int{8}
}

func (m *MsgFreezeAccountResponse) Reset()		{ *m = MsgFreezeAccountResponse{} }
func (m *MsgFreezeAccountResponse) String() string	{ return gogoproto.CompactTextString(m) }
func (*MsgFreezeAccountResponse) ProtoMessage()		{}
func (*MsgFreezeAccountResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountTx, []int{9}
}

func (m *MsgPayStorageDebt) Reset()		{ *m = MsgPayStorageDebt{} }
func (m *MsgPayStorageDebt) String() string	{ return gogoproto.CompactTextString(m) }
func (*MsgPayStorageDebt) ProtoMessage()	{}
func (*MsgPayStorageDebt) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountTx, []int{10}
}

func (m *MsgPayStorageDebtResponse) Reset()		{ *m = MsgPayStorageDebtResponse{} }
func (m *MsgPayStorageDebtResponse) String() string	{ return gogoproto.CompactTextString(m) }
func (*MsgPayStorageDebtResponse) ProtoMessage()	{}
func (*MsgPayStorageDebtResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountTx, []int{11}
}

func (m *MsgUnfreezeAccount) Reset()		{ *m = MsgUnfreezeAccount{} }
func (m *MsgUnfreezeAccount) String() string	{ return gogoproto.CompactTextString(m) }
func (*MsgUnfreezeAccount) ProtoMessage()	{}
func (*MsgUnfreezeAccount) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountTx, []int{12}
}

func (m *MsgUnfreezeAccountResponse) Reset()		{ *m = MsgUnfreezeAccountResponse{} }
func (m *MsgUnfreezeAccountResponse) String() string	{ return gogoproto.CompactTextString(m) }
func (*MsgUnfreezeAccountResponse) ProtoMessage()	{}
func (*MsgUnfreezeAccountResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountTx, []int{13}
}

func (m *MsgUpdateAccountMetadata) Reset()		{ *m = MsgUpdateAccountMetadata{} }
func (m *MsgUpdateAccountMetadata) String() string	{ return gogoproto.CompactTextString(m) }
func (*MsgUpdateAccountMetadata) ProtoMessage()		{}
func (*MsgUpdateAccountMetadata) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountTx, []int{14}
}

func (m *MsgUpdateAccountMetadataResponse) Reset()		{ *m = MsgUpdateAccountMetadataResponse{} }
func (m *MsgUpdateAccountMetadataResponse) String() string	{ return gogoproto.CompactTextString(m) }
func (*MsgUpdateAccountMetadataResponse) ProtoMessage()		{}
func (*MsgUpdateAccountMetadataResponse) Descriptor() ([]byte, []int) {
	return fileDescriptorNativeAccountTx, []int{15}
}
