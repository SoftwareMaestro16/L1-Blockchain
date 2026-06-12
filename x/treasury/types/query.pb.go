package types

import (
	context "context"
	fmt "fmt"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/gogoproto/gogoproto"
	grpc1 "github.com/cosmos/gogoproto/grpc"
	proto "github.com/cosmos/gogoproto/proto"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3	// please upgrade the proto package

type QueryTreasuryBalanceRequest struct {
}

func (m *QueryTreasuryBalanceRequest) Reset()		{ *m = QueryTreasuryBalanceRequest{} }
func (m *QueryTreasuryBalanceRequest) String() string	{ return proto.CompactTextString(m) }
func (*QueryTreasuryBalanceRequest) ProtoMessage()	{}
func (*QueryTreasuryBalanceRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_4947e03e23a1db6f, []int{0}
}
func (m *QueryTreasuryBalanceRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryTreasuryBalanceRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryTreasuryBalanceRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryTreasuryBalanceRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryTreasuryBalanceRequest.Merge(m, src)
}
func (m *QueryTreasuryBalanceRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryTreasuryBalanceRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryTreasuryBalanceRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryTreasuryBalanceRequest proto.InternalMessageInfo

type QueryTreasuryBalanceResponse struct {
	ModuleAccount		string						`protobuf:"bytes,1,opt,name=module_account,json=moduleAccount,proto3" json:"module_account,omitempty"`
	BankBalance		github_com_cosmos_cosmos_sdk_types.Coins	`protobuf:"bytes,2,rep,name=bank_balance,json=bankBalance,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"bank_balance"`
	AccountingBalance	github_com_cosmos_cosmos_sdk_types.Coins	`protobuf:"bytes,3,rep,name=accounting_balance,json=accountingBalance,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"accounting_balance"`
}

func (m *QueryTreasuryBalanceResponse) Reset()		{ *m = QueryTreasuryBalanceResponse{} }
func (m *QueryTreasuryBalanceResponse) String() string	{ return proto.CompactTextString(m) }
func (*QueryTreasuryBalanceResponse) ProtoMessage()	{}
func (*QueryTreasuryBalanceResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_4947e03e23a1db6f, []int{1}
}
func (m *QueryTreasuryBalanceResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryTreasuryBalanceResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryTreasuryBalanceResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryTreasuryBalanceResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryTreasuryBalanceResponse.Merge(m, src)
}
func (m *QueryTreasuryBalanceResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryTreasuryBalanceResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryTreasuryBalanceResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryTreasuryBalanceResponse proto.InternalMessageInfo

func (m *QueryTreasuryBalanceResponse) GetModuleAccount() string {
	if m != nil {
		return m.ModuleAccount
	}
	return ""
}

func (m *QueryTreasuryBalanceResponse) GetBankBalance() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.BankBalance
	}
	return nil
}

func (m *QueryTreasuryBalanceResponse) GetAccountingBalance() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.AccountingBalance
	}
	return nil
}

type QueryTreasuryAllocationsRequest struct {
}

func (m *QueryTreasuryAllocationsRequest) Reset()		{ *m = QueryTreasuryAllocationsRequest{} }
func (m *QueryTreasuryAllocationsRequest) String() string	{ return proto.CompactTextString(m) }
func (*QueryTreasuryAllocationsRequest) ProtoMessage()		{}
func (*QueryTreasuryAllocationsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_4947e03e23a1db6f, []int{2}
}
func (m *QueryTreasuryAllocationsRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryTreasuryAllocationsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryTreasuryAllocationsRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryTreasuryAllocationsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryTreasuryAllocationsRequest.Merge(m, src)
}
func (m *QueryTreasuryAllocationsRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryTreasuryAllocationsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryTreasuryAllocationsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryTreasuryAllocationsRequest proto.InternalMessageInfo

type QueryTreasuryAllocationsResponse struct {
	Allocations TreasuryAllocations `protobuf:"bytes,1,opt,name=allocations,proto3" json:"allocations"`
}

func (m *QueryTreasuryAllocationsResponse) Reset()		{ *m = QueryTreasuryAllocationsResponse{} }
func (m *QueryTreasuryAllocationsResponse) String() string	{ return proto.CompactTextString(m) }
func (*QueryTreasuryAllocationsResponse) ProtoMessage()		{}
func (*QueryTreasuryAllocationsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_4947e03e23a1db6f, []int{3}
}
func (m *QueryTreasuryAllocationsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryTreasuryAllocationsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryTreasuryAllocationsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryTreasuryAllocationsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryTreasuryAllocationsResponse.Merge(m, src)
}
func (m *QueryTreasuryAllocationsResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryTreasuryAllocationsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryTreasuryAllocationsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryTreasuryAllocationsResponse proto.InternalMessageInfo

func (m *QueryTreasuryAllocationsResponse) GetAllocations() TreasuryAllocations {
	if m != nil {
		return m.Allocations
	}
	return TreasuryAllocations{}
}

type QueryTreasurySpendRequest struct {
	SpendId uint64 `protobuf:"varint,1,opt,name=spend_id,json=spendId,proto3" json:"spend_id,omitempty"`
}

func (m *QueryTreasurySpendRequest) Reset()		{ *m = QueryTreasurySpendRequest{} }
func (m *QueryTreasurySpendRequest) String() string	{ return proto.CompactTextString(m) }
func (*QueryTreasurySpendRequest) ProtoMessage()	{}
func (*QueryTreasurySpendRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_4947e03e23a1db6f, []int{4}
}
func (m *QueryTreasurySpendRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryTreasurySpendRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryTreasurySpendRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryTreasurySpendRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryTreasurySpendRequest.Merge(m, src)
}
func (m *QueryTreasurySpendRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryTreasurySpendRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryTreasurySpendRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryTreasurySpendRequest proto.InternalMessageInfo

func (m *QueryTreasurySpendRequest) GetSpendId() uint64 {
	if m != nil {
		return m.SpendId
	}
	return 0
}

type QueryTreasurySpendResponse struct {
	Spend TreasurySpend `protobuf:"bytes,1,opt,name=spend,proto3" json:"spend"`
}

func (m *QueryTreasurySpendResponse) Reset()		{ *m = QueryTreasurySpendResponse{} }
func (m *QueryTreasurySpendResponse) String() string	{ return proto.CompactTextString(m) }
func (*QueryTreasurySpendResponse) ProtoMessage()	{}
func (*QueryTreasurySpendResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_4947e03e23a1db6f, []int{5}
}
func (m *QueryTreasurySpendResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryTreasurySpendResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryTreasurySpendResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryTreasurySpendResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryTreasurySpendResponse.Merge(m, src)
}
func (m *QueryTreasurySpendResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryTreasurySpendResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryTreasurySpendResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryTreasurySpendResponse proto.InternalMessageInfo

func (m *QueryTreasurySpendResponse) GetSpend() TreasurySpend {
	if m != nil {
		return m.Spend
	}
	return TreasurySpend{}
}

type QueryTreasurySpendsRequest struct {
	Status string `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
}

func (m *QueryTreasurySpendsRequest) Reset()		{ *m = QueryTreasurySpendsRequest{} }
func (m *QueryTreasurySpendsRequest) String() string	{ return proto.CompactTextString(m) }
func (*QueryTreasurySpendsRequest) ProtoMessage()	{}
func (*QueryTreasurySpendsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_4947e03e23a1db6f, []int{6}
}
func (m *QueryTreasurySpendsRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryTreasurySpendsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryTreasurySpendsRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryTreasurySpendsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryTreasurySpendsRequest.Merge(m, src)
}
func (m *QueryTreasurySpendsRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryTreasurySpendsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryTreasurySpendsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryTreasurySpendsRequest proto.InternalMessageInfo

func (m *QueryTreasurySpendsRequest) GetStatus() string {
	if m != nil {
		return m.Status
	}
	return ""
}

type QueryTreasurySpendsResponse struct {
	Spends []TreasurySpend `protobuf:"bytes,1,rep,name=spends,proto3" json:"spends"`
}

func (m *QueryTreasurySpendsResponse) Reset()		{ *m = QueryTreasurySpendsResponse{} }
func (m *QueryTreasurySpendsResponse) String() string	{ return proto.CompactTextString(m) }
func (*QueryTreasurySpendsResponse) ProtoMessage()	{}
func (*QueryTreasurySpendsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_4947e03e23a1db6f, []int{7}
}
func (m *QueryTreasurySpendsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryTreasurySpendsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryTreasurySpendsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryTreasurySpendsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryTreasurySpendsResponse.Merge(m, src)
}
func (m *QueryTreasurySpendsResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryTreasurySpendsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryTreasurySpendsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryTreasurySpendsResponse proto.InternalMessageInfo

func (m *QueryTreasurySpendsResponse) GetSpends() []TreasurySpend {
	if m != nil {
		return m.Spends
	}
	return nil
}

type QueryTreasuryParamsRequest struct {
}

func (m *QueryTreasuryParamsRequest) Reset()		{ *m = QueryTreasuryParamsRequest{} }
func (m *QueryTreasuryParamsRequest) String() string	{ return proto.CompactTextString(m) }
func (*QueryTreasuryParamsRequest) ProtoMessage()	{}
func (*QueryTreasuryParamsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_4947e03e23a1db6f, []int{8}
}
func (m *QueryTreasuryParamsRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryTreasuryParamsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryTreasuryParamsRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryTreasuryParamsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryTreasuryParamsRequest.Merge(m, src)
}
func (m *QueryTreasuryParamsRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryTreasuryParamsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryTreasuryParamsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryTreasuryParamsRequest proto.InternalMessageInfo

type QueryTreasuryParamsResponse struct {
	Params Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
}

func (m *QueryTreasuryParamsResponse) Reset()		{ *m = QueryTreasuryParamsResponse{} }
func (m *QueryTreasuryParamsResponse) String() string	{ return proto.CompactTextString(m) }
func (*QueryTreasuryParamsResponse) ProtoMessage()	{}
func (*QueryTreasuryParamsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_4947e03e23a1db6f, []int{9}
}
func (m *QueryTreasuryParamsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryTreasuryParamsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryTreasuryParamsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryTreasuryParamsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryTreasuryParamsResponse.Merge(m, src)
}
func (m *QueryTreasuryParamsResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryTreasuryParamsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryTreasuryParamsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryTreasuryParamsResponse proto.InternalMessageInfo

func (m *QueryTreasuryParamsResponse) GetParams() Params {
	if m != nil {
		return m.Params
	}
	return Params{}
}

func init() {
	proto.RegisterType((*QueryTreasuryBalanceRequest)(nil), "l1.treasury.v1.QueryTreasuryBalanceRequest")
	proto.RegisterType((*QueryTreasuryBalanceResponse)(nil), "l1.treasury.v1.QueryTreasuryBalanceResponse")
	proto.RegisterType((*QueryTreasuryAllocationsRequest)(nil), "l1.treasury.v1.QueryTreasuryAllocationsRequest")
	proto.RegisterType((*QueryTreasuryAllocationsResponse)(nil), "l1.treasury.v1.QueryTreasuryAllocationsResponse")
	proto.RegisterType((*QueryTreasurySpendRequest)(nil), "l1.treasury.v1.QueryTreasurySpendRequest")
	proto.RegisterType((*QueryTreasurySpendResponse)(nil), "l1.treasury.v1.QueryTreasurySpendResponse")
	proto.RegisterType((*QueryTreasurySpendsRequest)(nil), "l1.treasury.v1.QueryTreasurySpendsRequest")
	proto.RegisterType((*QueryTreasurySpendsResponse)(nil), "l1.treasury.v1.QueryTreasurySpendsResponse")
	proto.RegisterType((*QueryTreasuryParamsRequest)(nil), "l1.treasury.v1.QueryTreasuryParamsRequest")
	proto.RegisterType((*QueryTreasuryParamsResponse)(nil), "l1.treasury.v1.QueryTreasuryParamsResponse")
}

func init()	{ proto.RegisterFile("l1/treasury/v1/query.proto", fileDescriptor_4947e03e23a1db6f) }

var fileDescriptor_4947e03e23a1db6f = []byte{

	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x95, 0x4f, 0x6f, 0xd3, 0x3e,
	0x18, 0xc7, 0x9b, 0xfd, 0xe9, 0xef, 0x87, 0xcb, 0x86, 0x30, 0x68, 0xac, 0xd9, 0x96, 0x6e, 0x99,
	0x90, 0x36, 0xc6, 0xe2, 0x65, 0x4c, 0x48, 0x88, 0xd3, 0x8a, 0x38, 0x20, 0x2e, 0xd0, 0x21, 0x21,
	0xed, 0x32, 0x39, 0xa9, 0x15, 0xa2, 0xa5, 0x76, 0x16, 0x27, 0x15, 0x03, 0x71, 0xe1, 0xc6, 0x01,
	0x09, 0xc4, 0x99, 0x37, 0xc0, 0x8d, 0x77, 0xb1, 0xe3, 0x24, 0x2e, 0x9c, 0x00, 0xad, 0xbc, 0x10,
	0x14, 0xdb, 0xd9, 0x9a, 0x90, 0x75, 0x39, 0x70, 0x6a, 0x6c, 0x3f, 0xcf, 0xf3, 0xfd, 0xf8, 0xf1,
	0xd7, 0x2e, 0xd0, 0x03, 0x1b, 0xc5, 0x11, 0xc1, 0x3c, 0x89, 0x0e, 0x51, 0xdf, 0x46, 0x07, 0x09,
	0x89, 0x0e, 0xad, 0x30, 0x62, 0x31, 0x83, 0xd3, 0x81, 0x6d, 0x65, 0x6b, 0x56, 0xdf, 0xd6, 0x0d,
	0x97, 0xf1, 0x1e, 0xe3, 0xc8, 0xc1, 0x9c, 0xa0, 0xbe, 0xed, 0x90, 0x18, 0xdb, 0xc8, 0x65, 0x3e,
	0x95, 0xf1, 0xfa, 0x75, 0x8f, 0x79, 0x4c, 0x7c, 0xa2, 0xf4, 0x4b, 0xcd, 0xce, 0x7b, 0x8c, 0x79,
	0x01, 0x41, 0x38, 0xf4, 0x11, 0xa6, 0x94, 0xc5, 0x38, 0xf6, 0x19, 0xe5, 0xd9, 0x6a, 0x41, 0xdf,
	0x23, 0x94, 0x70, 0x5f, 0xad, 0x9a, 0x0b, 0x60, 0xee, 0x69, 0x0a, 0xf4, 0x4c, 0x45, 0xb4, 0x71,
	0x80, 0xa9, 0x4b, 0x3a, 0xe4, 0x20, 0x21, 0x3c, 0x36, 0xbf, 0x8e, 0x81, 0xf9, 0xf2, 0x75, 0x1e,
	0x32, 0xca, 0x09, 0xbc, 0x09, 0xa6, 0x7b, 0xac, 0x9b, 0x04, 0x64, 0x0f, 0xbb, 0x2e, 0x4b, 0x68,
	0x3c, 0xab, 0x2d, 0x6a, 0x2b, 0x97, 0x3a, 0x53, 0x72, 0x76, 0x5b, 0x4e, 0x42, 0x0a, 0x2e, 0x3b,
	0x98, 0xee, 0xef, 0x39, 0x32, 0x7d, 0x76, 0x6c, 0x71, 0x7c, 0xa5, 0xb1, 0xd9, 0xb4, 0xe4, 0x7e,
	0xad, 0x74, 0xbf, 0x96, 0xda, 0xaf, 0xf5, 0x80, 0xf9, 0xb4, 0xbd, 0x71, 0xf4, 0xa3, 0x55, 0xfb,
	0xf2, 0xb3, 0xb5, 0xe2, 0xf9, 0xf1, 0x8b, 0xc4, 0xb1, 0x5c, 0xd6, 0x43, 0xaa, 0x39, 0xf2, 0x67,
	0x9d, 0x77, 0xf7, 0x51, 0x7c, 0x18, 0x12, 0x2e, 0x12, 0x78, 0xa7, 0x91, 0x0a, 0x28, 0x3c, 0xf8,
	0x0a, 0x40, 0xc5, 0xe3, 0x53, 0xef, 0x54, 0x75, 0xfc, 0xdf, 0xab, 0x5e, 0x3d, 0x93, 0x51, 0xda,
	0xe6, 0x12, 0x68, 0xe5, 0x5a, 0xb6, 0x1d, 0x04, 0xcc, 0x95, 0x47, 0x92, 0xb5, 0x95, 0x81, 0xc5,
	0xf3, 0x43, 0x54, 0x67, 0x1f, 0x83, 0x06, 0x3e, 0x9b, 0x16, 0x6d, 0x6d, 0x6c, 0x2e, 0x5b, 0x79,
	0xc7, 0x58, 0x25, 0x15, 0xda, 0x13, 0xe9, 0x2e, 0x3a, 0xc3, 0xd9, 0xe6, 0x5d, 0xd0, 0xcc, 0x09,
	0xee, 0x84, 0x84, 0x76, 0x15, 0x0d, 0x6c, 0x82, 0xff, 0x79, 0x3a, 0xde, 0xf3, 0xbb, 0x42, 0x66,
	0xa2, 0xf3, 0x9f, 0x18, 0x3f, 0xea, 0x9a, 0xcf, 0x81, 0x5e, 0x96, 0xa7, 0x10, 0xef, 0x81, 0x49,
	0x11, 0xa8, 0xe0, 0x16, 0xce, 0x83, 0x13, 0x59, 0x0a, 0x4b, 0x66, 0x98, 0x5b, 0x65, 0x85, 0xb3,
	0xfe, 0xc0, 0x19, 0x50, 0xe7, 0x31, 0x8e, 0x13, 0xae, 0xdc, 0xa4, 0x46, 0xe6, 0x6e, 0xc1, 0xad,
	0x59, 0x96, 0xe2, 0xb9, 0x0f, 0xea, 0xa2, 0x7a, 0x9a, 0x36, 0x5e, 0x15, 0x48, 0xa5, 0x98, 0xf3,
	0x05, 0xa2, 0x27, 0x38, 0xc2, 0xbd, 0xd3, 0x13, 0xdb, 0x29, 0x28, 0x67, 0xab, 0x4a, 0x79, 0x0b,
	0xd4, 0x43, 0x31, 0xa3, 0x5a, 0x31, 0x53, 0x54, 0x96, 0xf1, 0x99, 0xa4, 0x8c, 0xdd, 0x1c, 0x4c,
	0x82, 0x49, 0x51, 0x15, 0xbe, 0xd7, 0xc0, 0x95, 0xc2, 0x15, 0x83, 0x6b, 0xc5, 0x1a, 0x23, 0x2e,
	0xaa, 0x7e, 0xbb, 0x5a, 0xb0, 0xc4, 0x35, 0x5b, 0x6f, 0xbf, 0xfd, 0xfe, 0x34, 0xd6, 0x84, 0x37,
	0x50, 0xe1, 0x71, 0x50, 0x37, 0x05, 0x7e, 0xd6, 0xc0, 0xb5, 0x12, 0x6b, 0x41, 0x34, 0x52, 0xe6,
	0x6f, 0xa7, 0xeb, 0x1b, 0xd5, 0x13, 0x14, 0xdb, 0xb2, 0x60, 0x5b, 0x80, 0x73, 0x45, 0xb6, 0x21,
	0x3f, 0xc3, 0x8f, 0x1a, 0x98, 0xca, 0x1d, 0x26, 0x5c, 0x1d, 0x29, 0x34, 0xec, 0x77, 0xfd, 0x56,
	0x95, 0x50, 0x45, 0xb3, 0x2a, 0x68, 0x96, 0xe1, 0x52, 0x91, 0x46, 0xba, 0x06, 0xbd, 0xce, 0x6e,
	0xce, 0x1b, 0xf8, 0x4e, 0x03, 0xd3, 0x79, 0x63, 0xc2, 0x0a, 0x4a, 0xa7, 0x9d, 0x5a, 0xab, 0x14,
	0xab, 0xb0, 0x0c, 0x81, 0x35, 0x0b, 0x67, 0xca, 0xb1, 0x72, 0x2c, 0xd2, 0x7a, 0x17, 0xb0, 0xe4,
	0xdc, 0x7e, 0x01, 0x4b, 0xde, 0xfb, 0xe7, 0xb3, 0x48, 0x97, 0xb7, 0x1f, 0x1e, 0x9d, 0x18, 0xda,
	0xf1, 0x89, 0xa1, 0xfd, 0x3a, 0x31, 0xb4, 0x0f, 0x03, 0xa3, 0x76, 0x3c, 0x30, 0x6a, 0xdf, 0x07,
	0x46, 0x6d, 0x77, 0x6d, 0xe8, 0x99, 0xe5, 0xac, 0x4f, 0x22, 0xe2, 0x7b, 0x74, 0x3d, 0xb0, 0xd3,
	0x42, 0x2f, 0xcf, 0x4a, 0x89, 0xf7, 0xd6, 0xa9, 0x8b, 0x3f, 0xac, 0x3b, 0x7f, 0x02, 0x00, 0x00,
	0xff, 0xff, 0xd8, 0xc9, 0xcb, 0x1d, 0x50, 0x07, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// QueryClient is the client API for Query service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type QueryClient interface {
	TreasuryBalance(ctx context.Context, in *QueryTreasuryBalanceRequest, opts ...grpc.CallOption) (*QueryTreasuryBalanceResponse, error)
	TreasuryAllocations(ctx context.Context, in *QueryTreasuryAllocationsRequest, opts ...grpc.CallOption) (*QueryTreasuryAllocationsResponse, error)
	TreasurySpend(ctx context.Context, in *QueryTreasurySpendRequest, opts ...grpc.CallOption) (*QueryTreasurySpendResponse, error)
	TreasurySpends(ctx context.Context, in *QueryTreasurySpendsRequest, opts ...grpc.CallOption) (*QueryTreasurySpendsResponse, error)
	TreasuryParams(ctx context.Context, in *QueryTreasuryParamsRequest, opts ...grpc.CallOption) (*QueryTreasuryParamsResponse, error)
}

type queryClient struct {
	cc grpc1.ClientConn
}

func NewQueryClient(cc grpc1.ClientConn) QueryClient {
	return &queryClient{cc}
}

func (c *queryClient) TreasuryBalance(ctx context.Context, in *QueryTreasuryBalanceRequest, opts ...grpc.CallOption) (*QueryTreasuryBalanceResponse, error) {
	out := new(QueryTreasuryBalanceResponse)
	err := c.cc.Invoke(ctx, "/l1.treasury.v1.Query/TreasuryBalance", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) TreasuryAllocations(ctx context.Context, in *QueryTreasuryAllocationsRequest, opts ...grpc.CallOption) (*QueryTreasuryAllocationsResponse, error) {
	out := new(QueryTreasuryAllocationsResponse)
	err := c.cc.Invoke(ctx, "/l1.treasury.v1.Query/TreasuryAllocations", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) TreasurySpend(ctx context.Context, in *QueryTreasurySpendRequest, opts ...grpc.CallOption) (*QueryTreasurySpendResponse, error) {
	out := new(QueryTreasurySpendResponse)
	err := c.cc.Invoke(ctx, "/l1.treasury.v1.Query/TreasurySpend", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) TreasurySpends(ctx context.Context, in *QueryTreasurySpendsRequest, opts ...grpc.CallOption) (*QueryTreasurySpendsResponse, error) {
	out := new(QueryTreasurySpendsResponse)
	err := c.cc.Invoke(ctx, "/l1.treasury.v1.Query/TreasurySpends", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) TreasuryParams(ctx context.Context, in *QueryTreasuryParamsRequest, opts ...grpc.CallOption) (*QueryTreasuryParamsResponse, error) {
	out := new(QueryTreasuryParamsResponse)
	err := c.cc.Invoke(ctx, "/l1.treasury.v1.Query/TreasuryParams", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QueryServer is the server API for Query service.
type QueryServer interface {
	TreasuryBalance(context.Context, *QueryTreasuryBalanceRequest) (*QueryTreasuryBalanceResponse, error)
	TreasuryAllocations(context.Context, *QueryTreasuryAllocationsRequest) (*QueryTreasuryAllocationsResponse, error)
	TreasurySpend(context.Context, *QueryTreasurySpendRequest) (*QueryTreasurySpendResponse, error)
	TreasurySpends(context.Context, *QueryTreasurySpendsRequest) (*QueryTreasurySpendsResponse, error)
	TreasuryParams(context.Context, *QueryTreasuryParamsRequest) (*QueryTreasuryParamsResponse, error)
}

// UnimplementedQueryServer can be embedded to have forward compatible implementations.
type UnimplementedQueryServer struct {
}

func (*UnimplementedQueryServer) TreasuryBalance(ctx context.Context, req *QueryTreasuryBalanceRequest) (*QueryTreasuryBalanceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TreasuryBalance not implemented")
}
func (*UnimplementedQueryServer) TreasuryAllocations(ctx context.Context, req *QueryTreasuryAllocationsRequest) (*QueryTreasuryAllocationsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TreasuryAllocations not implemented")
}
func (*UnimplementedQueryServer) TreasurySpend(ctx context.Context, req *QueryTreasurySpendRequest) (*QueryTreasurySpendResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TreasurySpend not implemented")
}
func (*UnimplementedQueryServer) TreasurySpends(ctx context.Context, req *QueryTreasurySpendsRequest) (*QueryTreasurySpendsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TreasurySpends not implemented")
}
func (*UnimplementedQueryServer) TreasuryParams(ctx context.Context, req *QueryTreasuryParamsRequest) (*QueryTreasuryParamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TreasuryParams not implemented")
}

func RegisterQueryServer(s grpc1.Server, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc, srv)
}

func _Query_TreasuryBalance_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryTreasuryBalanceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).TreasuryBalance(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.treasury.v1.Query/TreasuryBalance",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).TreasuryBalance(ctx, req.(*QueryTreasuryBalanceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_TreasuryAllocations_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryTreasuryAllocationsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).TreasuryAllocations(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.treasury.v1.Query/TreasuryAllocations",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).TreasuryAllocations(ctx, req.(*QueryTreasuryAllocationsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_TreasurySpend_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryTreasurySpendRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).TreasurySpend(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.treasury.v1.Query/TreasurySpend",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).TreasurySpend(ctx, req.(*QueryTreasurySpendRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_TreasurySpends_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryTreasurySpendsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).TreasurySpends(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.treasury.v1.Query/TreasurySpends",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).TreasurySpends(ctx, req.(*QueryTreasurySpendsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_TreasuryParams_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryTreasuryParamsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).TreasuryParams(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.treasury.v1.Query/TreasuryParams",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).TreasuryParams(ctx, req.(*QueryTreasuryParamsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var Query_serviceDesc = _Query_serviceDesc
var _Query_serviceDesc = grpc.ServiceDesc{
	ServiceName:	"l1.treasury.v1.Query",
	HandlerType:	(*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName:	"TreasuryBalance",
			Handler:	_Query_TreasuryBalance_Handler,
		},
		{
			MethodName:	"TreasuryAllocations",
			Handler:	_Query_TreasuryAllocations_Handler,
		},
		{
			MethodName:	"TreasurySpend",
			Handler:	_Query_TreasurySpend_Handler,
		},
		{
			MethodName:	"TreasurySpends",
			Handler:	_Query_TreasurySpends_Handler,
		},
		{
			MethodName:	"TreasuryParams",
			Handler:	_Query_TreasuryParams_Handler,
		},
	},
	Streams:	[]grpc.StreamDesc{},
	Metadata:	"l1/treasury/v1/query.proto",
}

func (m *QueryTreasuryBalanceRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryTreasuryBalanceRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryTreasuryBalanceRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *QueryTreasuryBalanceResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryTreasuryBalanceResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryTreasuryBalanceResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.AccountingBalance) > 0 {
		for iNdEx := len(m.AccountingBalance) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.AccountingBalance[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintQuery(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x1a
		}
	}
	if len(m.BankBalance) > 0 {
		for iNdEx := len(m.BankBalance) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.BankBalance[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintQuery(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
		}
	}
	if len(m.ModuleAccount) > 0 {
		i -= len(m.ModuleAccount)
		copy(dAtA[i:], m.ModuleAccount)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.ModuleAccount)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *QueryTreasuryAllocationsRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryTreasuryAllocationsRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryTreasuryAllocationsRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *QueryTreasuryAllocationsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryTreasuryAllocationsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryTreasuryAllocationsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.Allocations.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintQuery(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *QueryTreasurySpendRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryTreasurySpendRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryTreasurySpendRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.SpendId != 0 {
		i = encodeVarintQuery(dAtA, i, uint64(m.SpendId))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *QueryTreasurySpendResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryTreasurySpendResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryTreasurySpendResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.Spend.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintQuery(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *QueryTreasurySpendsRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryTreasurySpendsRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryTreasurySpendsRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Status) > 0 {
		i -= len(m.Status)
		copy(dAtA[i:], m.Status)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.Status)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *QueryTreasurySpendsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryTreasurySpendsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryTreasurySpendsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Spends) > 0 {
		for iNdEx := len(m.Spends) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Spends[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintQuery(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *QueryTreasuryParamsRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryTreasuryParamsRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryTreasuryParamsRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *QueryTreasuryParamsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryTreasuryParamsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryTreasuryParamsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.Params.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintQuery(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func encodeVarintQuery(dAtA []byte, offset int, v uint64) int {
	offset -= sovQuery(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *QueryTreasuryBalanceRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *QueryTreasuryBalanceResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ModuleAccount)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	if len(m.BankBalance) > 0 {
		for _, e := range m.BankBalance {
			l = e.Size()
			n += 1 + l + sovQuery(uint64(l))
		}
	}
	if len(m.AccountingBalance) > 0 {
		for _, e := range m.AccountingBalance {
			l = e.Size()
			n += 1 + l + sovQuery(uint64(l))
		}
	}
	return n
}

func (m *QueryTreasuryAllocationsRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *QueryTreasuryAllocationsResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Allocations.Size()
	n += 1 + l + sovQuery(uint64(l))
	return n
}

func (m *QueryTreasurySpendRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.SpendId != 0 {
		n += 1 + sovQuery(uint64(m.SpendId))
	}
	return n
}

func (m *QueryTreasurySpendResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Spend.Size()
	n += 1 + l + sovQuery(uint64(l))
	return n
}

func (m *QueryTreasurySpendsRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Status)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *QueryTreasurySpendsResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Spends) > 0 {
		for _, e := range m.Spends {
			l = e.Size()
			n += 1 + l + sovQuery(uint64(l))
		}
	}
	return n
}

func (m *QueryTreasuryParamsRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *QueryTreasuryParamsResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Params.Size()
	n += 1 + l + sovQuery(uint64(l))
	return n
}

func sovQuery(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozQuery(x uint64) (n int) {
	return sovQuery(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *QueryTreasuryBalanceRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: QueryTreasuryBalanceRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryTreasuryBalanceRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *QueryTreasuryBalanceResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: QueryTreasuryBalanceResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryTreasuryBalanceResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ModuleAccount", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ModuleAccount = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BankBalance", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.BankBalance = append(m.BankBalance, types.Coin{})
			if err := m.BankBalance[len(m.BankBalance)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AccountingBalance", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.AccountingBalance = append(m.AccountingBalance, types.Coin{})
			if err := m.AccountingBalance[len(m.AccountingBalance)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *QueryTreasuryAllocationsRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: QueryTreasuryAllocationsRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryTreasuryAllocationsRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *QueryTreasuryAllocationsResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: QueryTreasuryAllocationsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryTreasuryAllocationsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Allocations", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Allocations.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *QueryTreasurySpendRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: QueryTreasurySpendRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryTreasurySpendRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field SpendId", wireType)
			}
			m.SpendId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.SpendId |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *QueryTreasurySpendResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: QueryTreasurySpendResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryTreasurySpendResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Spend", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Spend.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *QueryTreasurySpendsRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: QueryTreasurySpendsRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryTreasurySpendsRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Status", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Status = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *QueryTreasurySpendsResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: QueryTreasurySpendsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryTreasurySpendsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Spends", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Spends = append(m.Spends, TreasurySpend{})
			if err := m.Spends[len(m.Spends)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *QueryTreasuryParamsRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: QueryTreasuryParamsRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryTreasuryParamsRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *QueryTreasuryParamsResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: QueryTreasuryParamsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryTreasuryParamsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Params", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Params.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipQuery(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthQuery
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupQuery
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthQuery
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthQuery		= fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowQuery		= fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupQuery	= fmt.Errorf("proto: unexpected end of group")
)
