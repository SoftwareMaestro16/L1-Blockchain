package types

import (
	context "context"
	fmt "fmt"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/cosmos-sdk/types/msgservice"
	_ "github.com/cosmos/gogoproto/gogoproto"
	grpc1 "github.com/cosmos/gogoproto/grpc"
	proto "github.com/cosmos/gogoproto/proto"
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

type MsgSubmitTreasurySpend struct {
	Proposer		string						`protobuf:"bytes,1,opt,name=proposer,proto3" json:"proposer,omitempty"`
	Recipient		string						`protobuf:"bytes,2,opt,name=recipient,proto3" json:"recipient,omitempty"`
	Amount			github_com_cosmos_cosmos_sdk_types.Coins	`protobuf:"bytes,3,rep,name=amount,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"amount"`
	Bucket			string						`protobuf:"bytes,4,opt,name=bucket,proto3" json:"bucket,omitempty"`
	Epoch			uint64						`protobuf:"varint,5,opt,name=epoch,proto3" json:"epoch,omitempty"`
	VestingStartEpoch	uint64						`protobuf:"varint,6,opt,name=vesting_start_epoch,json=vestingStartEpoch,proto3" json:"vesting_start_epoch,omitempty"`
	VestingEndEpoch		uint64						`protobuf:"varint,7,opt,name=vesting_end_epoch,json=vestingEndEpoch,proto3" json:"vesting_end_epoch,omitempty"`
	Metadata		string						`protobuf:"bytes,8,opt,name=metadata,proto3" json:"metadata,omitempty"`
}

func (m *MsgSubmitTreasurySpend) Reset()		{ *m = MsgSubmitTreasurySpend{} }
func (m *MsgSubmitTreasurySpend) String() string	{ return proto.CompactTextString(m) }
func (*MsgSubmitTreasurySpend) ProtoMessage()		{}
func (*MsgSubmitTreasurySpend) Descriptor() ([]byte, []int) {
	return fileDescriptor_61f1b768ba2508ee, []int{0}
}
func (m *MsgSubmitTreasurySpend) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgSubmitTreasurySpend) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgSubmitTreasurySpend.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgSubmitTreasurySpend) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgSubmitTreasurySpend.Merge(m, src)
}
func (m *MsgSubmitTreasurySpend) XXX_Size() int {
	return m.Size()
}
func (m *MsgSubmitTreasurySpend) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgSubmitTreasurySpend.DiscardUnknown(m)
}

var xxx_messageInfo_MsgSubmitTreasurySpend proto.InternalMessageInfo

func (m *MsgSubmitTreasurySpend) GetProposer() string {
	if m != nil {
		return m.Proposer
	}
	return ""
}

func (m *MsgSubmitTreasurySpend) GetRecipient() string {
	if m != nil {
		return m.Recipient
	}
	return ""
}

func (m *MsgSubmitTreasurySpend) GetAmount() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.Amount
	}
	return nil
}

func (m *MsgSubmitTreasurySpend) GetBucket() string {
	if m != nil {
		return m.Bucket
	}
	return ""
}

func (m *MsgSubmitTreasurySpend) GetEpoch() uint64 {
	if m != nil {
		return m.Epoch
	}
	return 0
}

func (m *MsgSubmitTreasurySpend) GetVestingStartEpoch() uint64 {
	if m != nil {
		return m.VestingStartEpoch
	}
	return 0
}

func (m *MsgSubmitTreasurySpend) GetVestingEndEpoch() uint64 {
	if m != nil {
		return m.VestingEndEpoch
	}
	return 0
}

func (m *MsgSubmitTreasurySpend) GetMetadata() string {
	if m != nil {
		return m.Metadata
	}
	return ""
}

type MsgSubmitTreasurySpendResponse struct {
	Spend TreasurySpend `protobuf:"bytes,1,opt,name=spend,proto3" json:"spend"`
}

func (m *MsgSubmitTreasurySpendResponse) Reset()		{ *m = MsgSubmitTreasurySpendResponse{} }
func (m *MsgSubmitTreasurySpendResponse) String() string	{ return proto.CompactTextString(m) }
func (*MsgSubmitTreasurySpendResponse) ProtoMessage()		{}
func (*MsgSubmitTreasurySpendResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_61f1b768ba2508ee, []int{1}
}
func (m *MsgSubmitTreasurySpendResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgSubmitTreasurySpendResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgSubmitTreasurySpendResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgSubmitTreasurySpendResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgSubmitTreasurySpendResponse.Merge(m, src)
}
func (m *MsgSubmitTreasurySpendResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgSubmitTreasurySpendResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgSubmitTreasurySpendResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgSubmitTreasurySpendResponse proto.InternalMessageInfo

func (m *MsgSubmitTreasurySpendResponse) GetSpend() TreasurySpend {
	if m != nil {
		return m.Spend
	}
	return TreasurySpend{}
}

type MsgApproveTreasurySpend struct {
	Authority	string	`protobuf:"bytes,1,opt,name=authority,proto3" json:"authority,omitempty"`
	SpendId		uint64	`protobuf:"varint,2,opt,name=spend_id,json=spendId,proto3" json:"spend_id,omitempty"`
	Metadata	string	`protobuf:"bytes,3,opt,name=metadata,proto3" json:"metadata,omitempty"`
}

func (m *MsgApproveTreasurySpend) Reset()		{ *m = MsgApproveTreasurySpend{} }
func (m *MsgApproveTreasurySpend) String() string	{ return proto.CompactTextString(m) }
func (*MsgApproveTreasurySpend) ProtoMessage()		{}
func (*MsgApproveTreasurySpend) Descriptor() ([]byte, []int) {
	return fileDescriptor_61f1b768ba2508ee, []int{2}
}
func (m *MsgApproveTreasurySpend) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgApproveTreasurySpend) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgApproveTreasurySpend.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgApproveTreasurySpend) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgApproveTreasurySpend.Merge(m, src)
}
func (m *MsgApproveTreasurySpend) XXX_Size() int {
	return m.Size()
}
func (m *MsgApproveTreasurySpend) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgApproveTreasurySpend.DiscardUnknown(m)
}

var xxx_messageInfo_MsgApproveTreasurySpend proto.InternalMessageInfo

func (m *MsgApproveTreasurySpend) GetAuthority() string {
	if m != nil {
		return m.Authority
	}
	return ""
}

func (m *MsgApproveTreasurySpend) GetSpendId() uint64 {
	if m != nil {
		return m.SpendId
	}
	return 0
}

func (m *MsgApproveTreasurySpend) GetMetadata() string {
	if m != nil {
		return m.Metadata
	}
	return ""
}

type MsgApproveTreasurySpendResponse struct {
	Spend TreasurySpend `protobuf:"bytes,1,opt,name=spend,proto3" json:"spend"`
}

func (m *MsgApproveTreasurySpendResponse) Reset()		{ *m = MsgApproveTreasurySpendResponse{} }
func (m *MsgApproveTreasurySpendResponse) String() string	{ return proto.CompactTextString(m) }
func (*MsgApproveTreasurySpendResponse) ProtoMessage()		{}
func (*MsgApproveTreasurySpendResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_61f1b768ba2508ee, []int{3}
}
func (m *MsgApproveTreasurySpendResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgApproveTreasurySpendResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgApproveTreasurySpendResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgApproveTreasurySpendResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgApproveTreasurySpendResponse.Merge(m, src)
}
func (m *MsgApproveTreasurySpendResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgApproveTreasurySpendResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgApproveTreasurySpendResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgApproveTreasurySpendResponse proto.InternalMessageInfo

func (m *MsgApproveTreasurySpendResponse) GetSpend() TreasurySpend {
	if m != nil {
		return m.Spend
	}
	return TreasurySpend{}
}

type MsgRejectTreasurySpend struct {
	Authority	string	`protobuf:"bytes,1,opt,name=authority,proto3" json:"authority,omitempty"`
	SpendId		uint64	`protobuf:"varint,2,opt,name=spend_id,json=spendId,proto3" json:"spend_id,omitempty"`
	Metadata	string	`protobuf:"bytes,3,opt,name=metadata,proto3" json:"metadata,omitempty"`
}

func (m *MsgRejectTreasurySpend) Reset()		{ *m = MsgRejectTreasurySpend{} }
func (m *MsgRejectTreasurySpend) String() string	{ return proto.CompactTextString(m) }
func (*MsgRejectTreasurySpend) ProtoMessage()		{}
func (*MsgRejectTreasurySpend) Descriptor() ([]byte, []int) {
	return fileDescriptor_61f1b768ba2508ee, []int{4}
}
func (m *MsgRejectTreasurySpend) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgRejectTreasurySpend) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgRejectTreasurySpend.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgRejectTreasurySpend) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgRejectTreasurySpend.Merge(m, src)
}
func (m *MsgRejectTreasurySpend) XXX_Size() int {
	return m.Size()
}
func (m *MsgRejectTreasurySpend) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgRejectTreasurySpend.DiscardUnknown(m)
}

var xxx_messageInfo_MsgRejectTreasurySpend proto.InternalMessageInfo

func (m *MsgRejectTreasurySpend) GetAuthority() string {
	if m != nil {
		return m.Authority
	}
	return ""
}

func (m *MsgRejectTreasurySpend) GetSpendId() uint64 {
	if m != nil {
		return m.SpendId
	}
	return 0
}

func (m *MsgRejectTreasurySpend) GetMetadata() string {
	if m != nil {
		return m.Metadata
	}
	return ""
}

type MsgRejectTreasurySpendResponse struct {
	Spend TreasurySpend `protobuf:"bytes,1,opt,name=spend,proto3" json:"spend"`
}

func (m *MsgRejectTreasurySpendResponse) Reset()		{ *m = MsgRejectTreasurySpendResponse{} }
func (m *MsgRejectTreasurySpendResponse) String() string	{ return proto.CompactTextString(m) }
func (*MsgRejectTreasurySpendResponse) ProtoMessage()		{}
func (*MsgRejectTreasurySpendResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_61f1b768ba2508ee, []int{5}
}
func (m *MsgRejectTreasurySpendResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgRejectTreasurySpendResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgRejectTreasurySpendResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgRejectTreasurySpendResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgRejectTreasurySpendResponse.Merge(m, src)
}
func (m *MsgRejectTreasurySpendResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgRejectTreasurySpendResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgRejectTreasurySpendResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgRejectTreasurySpendResponse proto.InternalMessageInfo

func (m *MsgRejectTreasurySpendResponse) GetSpend() TreasurySpend {
	if m != nil {
		return m.Spend
	}
	return TreasurySpend{}
}

type MsgExecuteTreasurySpend struct {
	Authority	string	`protobuf:"bytes,1,opt,name=authority,proto3" json:"authority,omitempty"`
	SpendId		uint64	`protobuf:"varint,2,opt,name=spend_id,json=spendId,proto3" json:"spend_id,omitempty"`
	Epoch		uint64	`protobuf:"varint,3,opt,name=epoch,proto3" json:"epoch,omitempty"`
}

func (m *MsgExecuteTreasurySpend) Reset()		{ *m = MsgExecuteTreasurySpend{} }
func (m *MsgExecuteTreasurySpend) String() string	{ return proto.CompactTextString(m) }
func (*MsgExecuteTreasurySpend) ProtoMessage()		{}
func (*MsgExecuteTreasurySpend) Descriptor() ([]byte, []int) {
	return fileDescriptor_61f1b768ba2508ee, []int{6}
}
func (m *MsgExecuteTreasurySpend) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgExecuteTreasurySpend) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgExecuteTreasurySpend.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgExecuteTreasurySpend) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgExecuteTreasurySpend.Merge(m, src)
}
func (m *MsgExecuteTreasurySpend) XXX_Size() int {
	return m.Size()
}
func (m *MsgExecuteTreasurySpend) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgExecuteTreasurySpend.DiscardUnknown(m)
}

var xxx_messageInfo_MsgExecuteTreasurySpend proto.InternalMessageInfo

func (m *MsgExecuteTreasurySpend) GetAuthority() string {
	if m != nil {
		return m.Authority
	}
	return ""
}

func (m *MsgExecuteTreasurySpend) GetSpendId() uint64 {
	if m != nil {
		return m.SpendId
	}
	return 0
}

func (m *MsgExecuteTreasurySpend) GetEpoch() uint64 {
	if m != nil {
		return m.Epoch
	}
	return 0
}

type MsgExecuteTreasurySpendResponse struct {
	Spend TreasurySpend `protobuf:"bytes,1,opt,name=spend,proto3" json:"spend"`
}

func (m *MsgExecuteTreasurySpendResponse) Reset()		{ *m = MsgExecuteTreasurySpendResponse{} }
func (m *MsgExecuteTreasurySpendResponse) String() string	{ return proto.CompactTextString(m) }
func (*MsgExecuteTreasurySpendResponse) ProtoMessage()		{}
func (*MsgExecuteTreasurySpendResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_61f1b768ba2508ee, []int{7}
}
func (m *MsgExecuteTreasurySpendResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgExecuteTreasurySpendResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgExecuteTreasurySpendResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgExecuteTreasurySpendResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgExecuteTreasurySpendResponse.Merge(m, src)
}
func (m *MsgExecuteTreasurySpendResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgExecuteTreasurySpendResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgExecuteTreasurySpendResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgExecuteTreasurySpendResponse proto.InternalMessageInfo

func (m *MsgExecuteTreasurySpendResponse) GetSpend() TreasurySpend {
	if m != nil {
		return m.Spend
	}
	return TreasurySpend{}
}

type MsgCancelTreasurySpend struct {
	Actor		string	`protobuf:"bytes,1,opt,name=actor,proto3" json:"actor,omitempty"`
	SpendId		uint64	`protobuf:"varint,2,opt,name=spend_id,json=spendId,proto3" json:"spend_id,omitempty"`
	Metadata	string	`protobuf:"bytes,3,opt,name=metadata,proto3" json:"metadata,omitempty"`
}

func (m *MsgCancelTreasurySpend) Reset()		{ *m = MsgCancelTreasurySpend{} }
func (m *MsgCancelTreasurySpend) String() string	{ return proto.CompactTextString(m) }
func (*MsgCancelTreasurySpend) ProtoMessage()		{}
func (*MsgCancelTreasurySpend) Descriptor() ([]byte, []int) {
	return fileDescriptor_61f1b768ba2508ee, []int{8}
}
func (m *MsgCancelTreasurySpend) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgCancelTreasurySpend) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgCancelTreasurySpend.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgCancelTreasurySpend) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgCancelTreasurySpend.Merge(m, src)
}
func (m *MsgCancelTreasurySpend) XXX_Size() int {
	return m.Size()
}
func (m *MsgCancelTreasurySpend) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgCancelTreasurySpend.DiscardUnknown(m)
}

var xxx_messageInfo_MsgCancelTreasurySpend proto.InternalMessageInfo

func (m *MsgCancelTreasurySpend) GetActor() string {
	if m != nil {
		return m.Actor
	}
	return ""
}

func (m *MsgCancelTreasurySpend) GetSpendId() uint64 {
	if m != nil {
		return m.SpendId
	}
	return 0
}

func (m *MsgCancelTreasurySpend) GetMetadata() string {
	if m != nil {
		return m.Metadata
	}
	return ""
}

type MsgCancelTreasurySpendResponse struct {
	Spend TreasurySpend `protobuf:"bytes,1,opt,name=spend,proto3" json:"spend"`
}

func (m *MsgCancelTreasurySpendResponse) Reset()		{ *m = MsgCancelTreasurySpendResponse{} }
func (m *MsgCancelTreasurySpendResponse) String() string	{ return proto.CompactTextString(m) }
func (*MsgCancelTreasurySpendResponse) ProtoMessage()		{}
func (*MsgCancelTreasurySpendResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_61f1b768ba2508ee, []int{9}
}
func (m *MsgCancelTreasurySpendResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgCancelTreasurySpendResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgCancelTreasurySpendResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgCancelTreasurySpendResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgCancelTreasurySpendResponse.Merge(m, src)
}
func (m *MsgCancelTreasurySpendResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgCancelTreasurySpendResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgCancelTreasurySpendResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgCancelTreasurySpendResponse proto.InternalMessageInfo

func (m *MsgCancelTreasurySpendResponse) GetSpend() TreasurySpend {
	if m != nil {
		return m.Spend
	}
	return TreasurySpend{}
}

type MsgUpdateTreasuryParams struct {
	Authority	string	`protobuf:"bytes,1,opt,name=authority,proto3" json:"authority,omitempty"`
	Params		Params	`protobuf:"bytes,2,opt,name=params,proto3" json:"params"`
}

func (m *MsgUpdateTreasuryParams) Reset()		{ *m = MsgUpdateTreasuryParams{} }
func (m *MsgUpdateTreasuryParams) String() string	{ return proto.CompactTextString(m) }
func (*MsgUpdateTreasuryParams) ProtoMessage()		{}
func (*MsgUpdateTreasuryParams) Descriptor() ([]byte, []int) {
	return fileDescriptor_61f1b768ba2508ee, []int{10}
}
func (m *MsgUpdateTreasuryParams) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgUpdateTreasuryParams) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgUpdateTreasuryParams.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgUpdateTreasuryParams) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgUpdateTreasuryParams.Merge(m, src)
}
func (m *MsgUpdateTreasuryParams) XXX_Size() int {
	return m.Size()
}
func (m *MsgUpdateTreasuryParams) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgUpdateTreasuryParams.DiscardUnknown(m)
}

var xxx_messageInfo_MsgUpdateTreasuryParams proto.InternalMessageInfo

func (m *MsgUpdateTreasuryParams) GetAuthority() string {
	if m != nil {
		return m.Authority
	}
	return ""
}

func (m *MsgUpdateTreasuryParams) GetParams() Params {
	if m != nil {
		return m.Params
	}
	return Params{}
}

type MsgUpdateTreasuryParamsResponse struct {
}

func (m *MsgUpdateTreasuryParamsResponse) Reset()		{ *m = MsgUpdateTreasuryParamsResponse{} }
func (m *MsgUpdateTreasuryParamsResponse) String() string	{ return proto.CompactTextString(m) }
func (*MsgUpdateTreasuryParamsResponse) ProtoMessage()		{}
func (*MsgUpdateTreasuryParamsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_61f1b768ba2508ee, []int{11}
}
func (m *MsgUpdateTreasuryParamsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgUpdateTreasuryParamsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgUpdateTreasuryParamsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgUpdateTreasuryParamsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgUpdateTreasuryParamsResponse.Merge(m, src)
}
func (m *MsgUpdateTreasuryParamsResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgUpdateTreasuryParamsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgUpdateTreasuryParamsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgUpdateTreasuryParamsResponse proto.InternalMessageInfo

func init() {
	proto.RegisterType((*MsgSubmitTreasurySpend)(nil), "l1.treasury.v1.MsgSubmitTreasurySpend")
	proto.RegisterType((*MsgSubmitTreasurySpendResponse)(nil), "l1.treasury.v1.MsgSubmitTreasurySpendResponse")
	proto.RegisterType((*MsgApproveTreasurySpend)(nil), "l1.treasury.v1.MsgApproveTreasurySpend")
	proto.RegisterType((*MsgApproveTreasurySpendResponse)(nil), "l1.treasury.v1.MsgApproveTreasurySpendResponse")
	proto.RegisterType((*MsgRejectTreasurySpend)(nil), "l1.treasury.v1.MsgRejectTreasurySpend")
	proto.RegisterType((*MsgRejectTreasurySpendResponse)(nil), "l1.treasury.v1.MsgRejectTreasurySpendResponse")
	proto.RegisterType((*MsgExecuteTreasurySpend)(nil), "l1.treasury.v1.MsgExecuteTreasurySpend")
	proto.RegisterType((*MsgExecuteTreasurySpendResponse)(nil), "l1.treasury.v1.MsgExecuteTreasurySpendResponse")
	proto.RegisterType((*MsgCancelTreasurySpend)(nil), "l1.treasury.v1.MsgCancelTreasurySpend")
	proto.RegisterType((*MsgCancelTreasurySpendResponse)(nil), "l1.treasury.v1.MsgCancelTreasurySpendResponse")
	proto.RegisterType((*MsgUpdateTreasuryParams)(nil), "l1.treasury.v1.MsgUpdateTreasuryParams")
	proto.RegisterType((*MsgUpdateTreasuryParamsResponse)(nil), "l1.treasury.v1.MsgUpdateTreasuryParamsResponse")
}

func init()	{ proto.RegisterFile("l1/treasury/v1/tx.proto", fileDescriptor_61f1b768ba2508ee) }

var fileDescriptor_61f1b768ba2508ee = []byte{

	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xc4, 0x56, 0x4f, 0x6f, 0xd3, 0x30,
	0x14, 0x6f, 0xd6, 0x3f, 0xeb, 0x3c, 0x31, 0x44, 0x5a, 0x6d, 0x5d, 0x34, 0xb2, 0xd1, 0x03, 0x54,
	0x43, 0x4b, 0x48, 0xe1, 0x02, 0x37, 0x36, 0xf5, 0xc0, 0x61, 0x12, 0xca, 0xe0, 0x02, 0x48, 0x95,
	0x9b, 0x58, 0x59, 0x58, 0x13, 0x9b, 0xd8, 0xa9, 0xba, 0x03, 0x08, 0xf1, 0x09, 0xf8, 0x1c, 0x5c,
	0xd8, 0xc7, 0xd8, 0x71, 0x47, 0x4e, 0x80, 0xd6, 0xc3, 0x2e, 0x7c, 0x08, 0x14, 0xdb, 0x6d, 0xd7,
	0xe2, 0x95, 0x09, 0x55, 0xe2, 0xd4, 0x3c, 0xbf, 0x9f, 0xdf, 0xaf, 0xbf, 0xf7, 0xf3, 0xb3, 0x0c,
	0xd6, 0xba, 0x8e, 0xcd, 0x12, 0x04, 0x69, 0x9a, 0x1c, 0xdb, 0x3d, 0xc7, 0x66, 0x7d, 0x8b, 0x24,
	0x98, 0x61, 0x7d, 0xa5, 0xeb, 0x58, 0xc3, 0x84, 0xd5, 0x73, 0x0c, 0xd3, 0xc3, 0x34, 0xc2, 0xd4,
	0xee, 0x40, 0x8a, 0xec, 0x9e, 0xd3, 0x41, 0x0c, 0x3a, 0xb6, 0x87, 0xc3, 0x58, 0xe0, 0x8d, 0x35,
	0x99, 0x8f, 0x68, 0x90, 0xd5, 0x89, 0x68, 0x20, 0x13, 0xd5, 0x00, 0x07, 0x98, 0x7f, 0xda, 0xd9,
	0x97, 0x5c, 0xdd, 0x98, 0xe2, 0x0d, 0x50, 0x8c, 0x68, 0x48, 0x45, 0xb6, 0xfe, 0x6b, 0x01, 0xac,
	0xee, 0xd3, 0xe0, 0x20, 0xed, 0x44, 0x21, 0x7b, 0x21, 0x61, 0x07, 0x04, 0xc5, 0xbe, 0x6e, 0x80,
	0x32, 0x49, 0x30, 0xc1, 0x14, 0x25, 0x35, 0x6d, 0x4b, 0x6b, 0x2c, 0xb9, 0xa3, 0x58, 0xdf, 0x00,
	0x4b, 0x09, 0xf2, 0x42, 0x12, 0xa2, 0x98, 0xd5, 0x16, 0x78, 0x72, 0xbc, 0xa0, 0x7b, 0xa0, 0x04,
	0x23, 0x9c, 0xc6, 0xac, 0x96, 0xdf, 0xca, 0x37, 0x96, 0x9b, 0xeb, 0x96, 0xf8, 0xcb, 0x56, 0x26,
	0xc9, 0x92, 0x92, 0xac, 0x3d, 0x1c, 0xc6, 0xbb, 0x0f, 0x4e, 0xbf, 0x6f, 0xe6, 0xbe, 0xfc, 0xd8,
	0x6c, 0x04, 0x21, 0x3b, 0x4c, 0x3b, 0x96, 0x87, 0x23, 0x5b, 0xea, 0x13, 0x3f, 0x3b, 0xd4, 0x3f,
	0xb2, 0xd9, 0x31, 0x41, 0x94, 0x6f, 0xa0, 0xae, 0x2c, 0xad, 0xaf, 0x82, 0x52, 0x27, 0xf5, 0x8e,
	0x10, 0xab, 0x15, 0x38, 0xbf, 0x8c, 0xf4, 0x2a, 0x28, 0x22, 0x82, 0xbd, 0xc3, 0x5a, 0x71, 0x4b,
	0x6b, 0x14, 0x5c, 0x11, 0xe8, 0x16, 0xa8, 0xf4, 0x10, 0x65, 0x61, 0x1c, 0xb4, 0x29, 0x83, 0x09,
	0x6b, 0x0b, 0x4c, 0x89, 0x63, 0x6e, 0xc9, 0xd4, 0x41, 0x96, 0x69, 0x71, 0xfc, 0x36, 0x18, 0x2e,
	0xb6, 0x51, 0xec, 0x4b, 0xf4, 0x22, 0x47, 0xdf, 0x94, 0x89, 0x56, 0xec, 0x0b, 0xac, 0x01, 0xca,
	0x11, 0x62, 0xd0, 0x87, 0x0c, 0xd6, 0xca, 0xa2, 0x51, 0xc3, 0xf8, 0xc9, 0x8d, 0x4f, 0x17, 0x27,
	0xdb, 0xa3, 0xbe, 0xd5, 0x5f, 0x03, 0x53, 0xdd, 0x6d, 0x17, 0x51, 0x82, 0x63, 0x8a, 0xf4, 0xc7,
	0xa0, 0x48, 0xb3, 0x05, 0xde, 0xf2, 0xe5, 0xe6, 0x6d, 0x6b, 0xf2, 0x74, 0x58, 0x13, 0xbb, 0x76,
	0x0b, 0x59, 0xfb, 0x5c, 0xb1, 0xa3, 0xfe, 0x01, 0xac, 0xed, 0xd3, 0xe0, 0x29, 0x21, 0x09, 0xee,
	0xa1, 0x49, 0x2f, 0x37, 0xc0, 0x12, 0x4c, 0xd9, 0x21, 0x4e, 0x42, 0x76, 0x2c, 0xcd, 0x1c, 0x2f,
	0xe8, 0xeb, 0xa0, 0xcc, 0x2b, 0xb4, 0x43, 0x9f, 0x9b, 0x59, 0x70, 0x17, 0x79, 0xfc, 0xcc, 0x9f,
	0xd0, 0x96, 0x9f, 0xd2, 0xb6, 0x92, 0x69, 0x1b, 0x97, 0xa9, 0xbf, 0x01, 0x9b, 0x57, 0xf0, 0xcf,
	0x43, 0xdd, 0x7b, 0x7e, 0x50, 0x5d, 0xf4, 0x16, 0x79, 0xec, 0x3f, 0x88, 0x13, 0xce, 0x29, 0xe8,
	0xe7, 0xa1, 0xad, 0xcf, 0x9d, 0x6b, 0xf5, 0x91, 0x97, 0xb2, 0x79, 0x39, 0x37, 0x9a, 0x83, 0xfc,
	0xa5, 0x39, 0xb8, 0xc2, 0x33, 0x15, 0xf3, 0x3c, 0x74, 0xbd, 0xe3, 0x9e, 0xed, 0xc1, 0xd8, 0x43,
	0xdd, 0x49, 0x59, 0x55, 0x50, 0x84, 0x1e, 0xc3, 0xc3, 0x9b, 0x45, 0x04, 0xff, 0xea, 0x15, 0xc8,
	0x44, 0x89, 0x12, 0xd2, 0x27, 0x05, 0xe5, 0x7c, 0xce, 0x60, 0xe6, 0xd3, 0x4b, 0xe2, 0xc3, 0x71,
	0xb3, 0x9e, 0xc3, 0x04, 0x46, 0xf4, 0x2f, 0x3e, 0x3d, 0x02, 0x25, 0xc2, 0x71, 0x5c, 0xd6, 0x72,
	0x73, 0x75, 0x9a, 0x54, 0x54, 0x91, 0x6c, 0x12, 0xfb, 0x87, 0x59, 0x77, 0xb8, 0x59, 0x2a, 0xfa,
	0xa1, 0xb8, 0xe6, 0xd7, 0x22, 0xc8, 0xef, 0xd3, 0x40, 0x8f, 0x40, 0x45, 0x75, 0xa7, 0xdf, 0x9d,
	0xe6, 0x55, 0xdf, 0x46, 0x86, 0x75, 0x3d, 0xdc, 0xa8, 0xa7, 0x04, 0x54, 0x95, 0xf7, 0xce, 0x3d,
	0x45, 0x1d, 0x15, 0xd0, 0xb0, 0xaf, 0x09, 0x1c, 0x31, 0x46, 0xa0, 0xa2, 0xba, 0x0b, 0x54, 0x02,
	0x15, 0x38, 0xa5, 0xc0, 0x59, 0xc3, 0x4d, 0x40, 0x55, 0x39, 0x9e, 0x2a, 0x81, 0x2a, 0xa0, 0x52,
	0xe0, 0xcc, 0xb1, 0x8b, 0x40, 0x45, 0x35, 0x38, 0x2a, 0x81, 0x0a, 0x9c, 0x52, 0xe0, 0xac, 0xa9,
	0x20, 0xa0, 0xaa, 0x3c, 0xd7, 0x2a, 0x81, 0x2a, 0xa0, 0x52, 0xe0, 0xac, 0xa3, 0x6a, 0x14, 0x3f,
	0x5e, 0x9c, 0x6c, 0x6b, 0xbb, 0xad, 0xd3, 0x73, 0x53, 0x3b, 0x3b, 0x37, 0xb5, 0x9f, 0xe7, 0xa6,
	0xf6, 0x79, 0x60, 0xe6, 0xce, 0x06, 0x66, 0xee, 0xdb, 0xc0, 0xcc, 0xbd, 0xba, 0x7f, 0xe9, 0x4d,
	0x40, 0x71, 0x0f, 0x25, 0x28, 0x0c, 0xe2, 0x9d, 0xae, 0x63, 0x77, 0x1d, 0xbb, 0x3f, 0x7e, 0xd3,
	0xf0, 0xc7, 0x41, 0xa7, 0xc4, 0xdf, 0x33, 0x0f, 0x7f, 0x07, 0x00, 0x00, 0xff, 0xff, 0xf6, 0x8d,
	0x79, 0x77, 0x67, 0x09, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// MsgClient is the client API for Msg service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type MsgClient interface {
	SubmitTreasurySpend(ctx context.Context, in *MsgSubmitTreasurySpend, opts ...grpc.CallOption) (*MsgSubmitTreasurySpendResponse, error)
	ApproveTreasurySpend(ctx context.Context, in *MsgApproveTreasurySpend, opts ...grpc.CallOption) (*MsgApproveTreasurySpendResponse, error)
	RejectTreasurySpend(ctx context.Context, in *MsgRejectTreasurySpend, opts ...grpc.CallOption) (*MsgRejectTreasurySpendResponse, error)
	ExecuteTreasurySpend(ctx context.Context, in *MsgExecuteTreasurySpend, opts ...grpc.CallOption) (*MsgExecuteTreasurySpendResponse, error)
	CancelTreasurySpend(ctx context.Context, in *MsgCancelTreasurySpend, opts ...grpc.CallOption) (*MsgCancelTreasurySpendResponse, error)
	UpdateTreasuryParams(ctx context.Context, in *MsgUpdateTreasuryParams, opts ...grpc.CallOption) (*MsgUpdateTreasuryParamsResponse, error)
}

type msgClient struct {
	cc grpc1.ClientConn
}

func NewMsgClient(cc grpc1.ClientConn) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) SubmitTreasurySpend(ctx context.Context, in *MsgSubmitTreasurySpend, opts ...grpc.CallOption) (*MsgSubmitTreasurySpendResponse, error) {
	out := new(MsgSubmitTreasurySpendResponse)
	err := c.cc.Invoke(ctx, "/l1.treasury.v1.Msg/SubmitTreasurySpend", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) ApproveTreasurySpend(ctx context.Context, in *MsgApproveTreasurySpend, opts ...grpc.CallOption) (*MsgApproveTreasurySpendResponse, error) {
	out := new(MsgApproveTreasurySpendResponse)
	err := c.cc.Invoke(ctx, "/l1.treasury.v1.Msg/ApproveTreasurySpend", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) RejectTreasurySpend(ctx context.Context, in *MsgRejectTreasurySpend, opts ...grpc.CallOption) (*MsgRejectTreasurySpendResponse, error) {
	out := new(MsgRejectTreasurySpendResponse)
	err := c.cc.Invoke(ctx, "/l1.treasury.v1.Msg/RejectTreasurySpend", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) ExecuteTreasurySpend(ctx context.Context, in *MsgExecuteTreasurySpend, opts ...grpc.CallOption) (*MsgExecuteTreasurySpendResponse, error) {
	out := new(MsgExecuteTreasurySpendResponse)
	err := c.cc.Invoke(ctx, "/l1.treasury.v1.Msg/ExecuteTreasurySpend", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) CancelTreasurySpend(ctx context.Context, in *MsgCancelTreasurySpend, opts ...grpc.CallOption) (*MsgCancelTreasurySpendResponse, error) {
	out := new(MsgCancelTreasurySpendResponse)
	err := c.cc.Invoke(ctx, "/l1.treasury.v1.Msg/CancelTreasurySpend", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) UpdateTreasuryParams(ctx context.Context, in *MsgUpdateTreasuryParams, opts ...grpc.CallOption) (*MsgUpdateTreasuryParamsResponse, error) {
	out := new(MsgUpdateTreasuryParamsResponse)
	err := c.cc.Invoke(ctx, "/l1.treasury.v1.Msg/UpdateTreasuryParams", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MsgServer is the server API for Msg service.
type MsgServer interface {
	SubmitTreasurySpend(context.Context, *MsgSubmitTreasurySpend) (*MsgSubmitTreasurySpendResponse, error)
	ApproveTreasurySpend(context.Context, *MsgApproveTreasurySpend) (*MsgApproveTreasurySpendResponse, error)
	RejectTreasurySpend(context.Context, *MsgRejectTreasurySpend) (*MsgRejectTreasurySpendResponse, error)
	ExecuteTreasurySpend(context.Context, *MsgExecuteTreasurySpend) (*MsgExecuteTreasurySpendResponse, error)
	CancelTreasurySpend(context.Context, *MsgCancelTreasurySpend) (*MsgCancelTreasurySpendResponse, error)
	UpdateTreasuryParams(context.Context, *MsgUpdateTreasuryParams) (*MsgUpdateTreasuryParamsResponse, error)
}

// UnimplementedMsgServer can be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (*UnimplementedMsgServer) SubmitTreasurySpend(ctx context.Context, req *MsgSubmitTreasurySpend) (*MsgSubmitTreasurySpendResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitTreasurySpend not implemented")
}
func (*UnimplementedMsgServer) ApproveTreasurySpend(ctx context.Context, req *MsgApproveTreasurySpend) (*MsgApproveTreasurySpendResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ApproveTreasurySpend not implemented")
}
func (*UnimplementedMsgServer) RejectTreasurySpend(ctx context.Context, req *MsgRejectTreasurySpend) (*MsgRejectTreasurySpendResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RejectTreasurySpend not implemented")
}
func (*UnimplementedMsgServer) ExecuteTreasurySpend(ctx context.Context, req *MsgExecuteTreasurySpend) (*MsgExecuteTreasurySpendResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ExecuteTreasurySpend not implemented")
}
func (*UnimplementedMsgServer) CancelTreasurySpend(ctx context.Context, req *MsgCancelTreasurySpend) (*MsgCancelTreasurySpendResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CancelTreasurySpend not implemented")
}
func (*UnimplementedMsgServer) UpdateTreasuryParams(ctx context.Context, req *MsgUpdateTreasuryParams) (*MsgUpdateTreasuryParamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateTreasuryParams not implemented")
}

func RegisterMsgServer(s grpc1.Server, srv MsgServer) {
	s.RegisterService(&_Msg_serviceDesc, srv)
}

func _Msg_SubmitTreasurySpend_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgSubmitTreasurySpend)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).SubmitTreasurySpend(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.treasury.v1.Msg/SubmitTreasurySpend",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).SubmitTreasurySpend(ctx, req.(*MsgSubmitTreasurySpend))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_ApproveTreasurySpend_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgApproveTreasurySpend)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).ApproveTreasurySpend(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.treasury.v1.Msg/ApproveTreasurySpend",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).ApproveTreasurySpend(ctx, req.(*MsgApproveTreasurySpend))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_RejectTreasurySpend_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgRejectTreasurySpend)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).RejectTreasurySpend(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.treasury.v1.Msg/RejectTreasurySpend",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).RejectTreasurySpend(ctx, req.(*MsgRejectTreasurySpend))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_ExecuteTreasurySpend_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgExecuteTreasurySpend)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).ExecuteTreasurySpend(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.treasury.v1.Msg/ExecuteTreasurySpend",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).ExecuteTreasurySpend(ctx, req.(*MsgExecuteTreasurySpend))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_CancelTreasurySpend_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgCancelTreasurySpend)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).CancelTreasurySpend(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.treasury.v1.Msg/CancelTreasurySpend",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).CancelTreasurySpend(ctx, req.(*MsgCancelTreasurySpend))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_UpdateTreasuryParams_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUpdateTreasuryParams)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).UpdateTreasuryParams(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.treasury.v1.Msg/UpdateTreasuryParams",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).UpdateTreasuryParams(ctx, req.(*MsgUpdateTreasuryParams))
	}
	return interceptor(ctx, in, info, handler)
}

var Msg_serviceDesc = _Msg_serviceDesc
var _Msg_serviceDesc = grpc.ServiceDesc{
	ServiceName:	"l1.treasury.v1.Msg",
	HandlerType:	(*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName:	"SubmitTreasurySpend",
			Handler:	_Msg_SubmitTreasurySpend_Handler,
		},
		{
			MethodName:	"ApproveTreasurySpend",
			Handler:	_Msg_ApproveTreasurySpend_Handler,
		},
		{
			MethodName:	"RejectTreasurySpend",
			Handler:	_Msg_RejectTreasurySpend_Handler,
		},
		{
			MethodName:	"ExecuteTreasurySpend",
			Handler:	_Msg_ExecuteTreasurySpend_Handler,
		},
		{
			MethodName:	"CancelTreasurySpend",
			Handler:	_Msg_CancelTreasurySpend_Handler,
		},
		{
			MethodName:	"UpdateTreasuryParams",
			Handler:	_Msg_UpdateTreasuryParams_Handler,
		},
	},
	Streams:	[]grpc.StreamDesc{},
	Metadata:	"l1/treasury/v1/tx.proto",
}

func (m *MsgSubmitTreasurySpend) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgSubmitTreasurySpend) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgSubmitTreasurySpend) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Metadata) > 0 {
		i -= len(m.Metadata)
		copy(dAtA[i:], m.Metadata)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Metadata)))
		i--
		dAtA[i] = 0x42
	}
	if m.VestingEndEpoch != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.VestingEndEpoch))
		i--
		dAtA[i] = 0x38
	}
	if m.VestingStartEpoch != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.VestingStartEpoch))
		i--
		dAtA[i] = 0x30
	}
	if m.Epoch != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.Epoch))
		i--
		dAtA[i] = 0x28
	}
	if len(m.Bucket) > 0 {
		i -= len(m.Bucket)
		copy(dAtA[i:], m.Bucket)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Bucket)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.Amount) > 0 {
		for iNdEx := len(m.Amount) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Amount[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintTx(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x1a
		}
	}
	if len(m.Recipient) > 0 {
		i -= len(m.Recipient)
		copy(dAtA[i:], m.Recipient)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Recipient)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Proposer) > 0 {
		i -= len(m.Proposer)
		copy(dAtA[i:], m.Proposer)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Proposer)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MsgSubmitTreasurySpendResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgSubmitTreasurySpendResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgSubmitTreasurySpendResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
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
		i = encodeVarintTx(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *MsgApproveTreasurySpend) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgApproveTreasurySpend) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgApproveTreasurySpend) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Metadata) > 0 {
		i -= len(m.Metadata)
		copy(dAtA[i:], m.Metadata)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Metadata)))
		i--
		dAtA[i] = 0x1a
	}
	if m.SpendId != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.SpendId))
		i--
		dAtA[i] = 0x10
	}
	if len(m.Authority) > 0 {
		i -= len(m.Authority)
		copy(dAtA[i:], m.Authority)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Authority)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MsgApproveTreasurySpendResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgApproveTreasurySpendResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgApproveTreasurySpendResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
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
		i = encodeVarintTx(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *MsgRejectTreasurySpend) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgRejectTreasurySpend) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgRejectTreasurySpend) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Metadata) > 0 {
		i -= len(m.Metadata)
		copy(dAtA[i:], m.Metadata)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Metadata)))
		i--
		dAtA[i] = 0x1a
	}
	if m.SpendId != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.SpendId))
		i--
		dAtA[i] = 0x10
	}
	if len(m.Authority) > 0 {
		i -= len(m.Authority)
		copy(dAtA[i:], m.Authority)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Authority)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MsgRejectTreasurySpendResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgRejectTreasurySpendResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgRejectTreasurySpendResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
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
		i = encodeVarintTx(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *MsgExecuteTreasurySpend) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgExecuteTreasurySpend) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgExecuteTreasurySpend) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Epoch != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.Epoch))
		i--
		dAtA[i] = 0x18
	}
	if m.SpendId != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.SpendId))
		i--
		dAtA[i] = 0x10
	}
	if len(m.Authority) > 0 {
		i -= len(m.Authority)
		copy(dAtA[i:], m.Authority)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Authority)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MsgExecuteTreasurySpendResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgExecuteTreasurySpendResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgExecuteTreasurySpendResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
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
		i = encodeVarintTx(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *MsgCancelTreasurySpend) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgCancelTreasurySpend) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgCancelTreasurySpend) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Metadata) > 0 {
		i -= len(m.Metadata)
		copy(dAtA[i:], m.Metadata)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Metadata)))
		i--
		dAtA[i] = 0x1a
	}
	if m.SpendId != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.SpendId))
		i--
		dAtA[i] = 0x10
	}
	if len(m.Actor) > 0 {
		i -= len(m.Actor)
		copy(dAtA[i:], m.Actor)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Actor)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MsgCancelTreasurySpendResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgCancelTreasurySpendResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgCancelTreasurySpendResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
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
		i = encodeVarintTx(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *MsgUpdateTreasuryParams) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgUpdateTreasuryParams) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgUpdateTreasuryParams) MarshalToSizedBuffer(dAtA []byte) (int, error) {
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
		i = encodeVarintTx(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	if len(m.Authority) > 0 {
		i -= len(m.Authority)
		copy(dAtA[i:], m.Authority)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Authority)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MsgUpdateTreasuryParamsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgUpdateTreasuryParamsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgUpdateTreasuryParamsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func encodeVarintTx(dAtA []byte, offset int, v uint64) int {
	offset -= sovTx(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *MsgSubmitTreasurySpend) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Proposer)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.Recipient)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if len(m.Amount) > 0 {
		for _, e := range m.Amount {
			l = e.Size()
			n += 1 + l + sovTx(uint64(l))
		}
	}
	l = len(m.Bucket)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if m.Epoch != 0 {
		n += 1 + sovTx(uint64(m.Epoch))
	}
	if m.VestingStartEpoch != 0 {
		n += 1 + sovTx(uint64(m.VestingStartEpoch))
	}
	if m.VestingEndEpoch != 0 {
		n += 1 + sovTx(uint64(m.VestingEndEpoch))
	}
	l = len(m.Metadata)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	return n
}

func (m *MsgSubmitTreasurySpendResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Spend.Size()
	n += 1 + l + sovTx(uint64(l))
	return n
}

func (m *MsgApproveTreasurySpend) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Authority)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if m.SpendId != 0 {
		n += 1 + sovTx(uint64(m.SpendId))
	}
	l = len(m.Metadata)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	return n
}

func (m *MsgApproveTreasurySpendResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Spend.Size()
	n += 1 + l + sovTx(uint64(l))
	return n
}

func (m *MsgRejectTreasurySpend) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Authority)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if m.SpendId != 0 {
		n += 1 + sovTx(uint64(m.SpendId))
	}
	l = len(m.Metadata)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	return n
}

func (m *MsgRejectTreasurySpendResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Spend.Size()
	n += 1 + l + sovTx(uint64(l))
	return n
}

func (m *MsgExecuteTreasurySpend) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Authority)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if m.SpendId != 0 {
		n += 1 + sovTx(uint64(m.SpendId))
	}
	if m.Epoch != 0 {
		n += 1 + sovTx(uint64(m.Epoch))
	}
	return n
}

func (m *MsgExecuteTreasurySpendResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Spend.Size()
	n += 1 + l + sovTx(uint64(l))
	return n
}

func (m *MsgCancelTreasurySpend) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Actor)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if m.SpendId != 0 {
		n += 1 + sovTx(uint64(m.SpendId))
	}
	l = len(m.Metadata)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	return n
}

func (m *MsgCancelTreasurySpendResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Spend.Size()
	n += 1 + l + sovTx(uint64(l))
	return n
}

func (m *MsgUpdateTreasuryParams) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Authority)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = m.Params.Size()
	n += 1 + l + sovTx(uint64(l))
	return n
}

func (m *MsgUpdateTreasuryParamsResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func sovTx(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozTx(x uint64) (n int) {
	return sovTx(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *MsgSubmitTreasurySpend) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
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
			return fmt.Errorf("proto: MsgSubmitTreasurySpend: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgSubmitTreasurySpend: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Proposer", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Proposer = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Recipient", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Recipient = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Amount", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Amount = append(m.Amount, types.Coin{})
			if err := m.Amount[len(m.Amount)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Bucket", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Bucket = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Epoch", wireType)
			}
			m.Epoch = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Epoch |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 6:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field VestingStartEpoch", wireType)
			}
			m.VestingStartEpoch = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.VestingStartEpoch |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 7:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field VestingEndEpoch", wireType)
			}
			m.VestingEndEpoch = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.VestingEndEpoch |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 8:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Metadata", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Metadata = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
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
func (m *MsgSubmitTreasurySpendResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
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
			return fmt.Errorf("proto: MsgSubmitTreasurySpendResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgSubmitTreasurySpendResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Spend", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTx
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
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
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
func (m *MsgApproveTreasurySpend) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
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
			return fmt.Errorf("proto: MsgApproveTreasurySpend: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgApproveTreasurySpend: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Authority", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Authority = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field SpendId", wireType)
			}
			m.SpendId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Metadata", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Metadata = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
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
func (m *MsgApproveTreasurySpendResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
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
			return fmt.Errorf("proto: MsgApproveTreasurySpendResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgApproveTreasurySpendResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Spend", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTx
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
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
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
func (m *MsgRejectTreasurySpend) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
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
			return fmt.Errorf("proto: MsgRejectTreasurySpend: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgRejectTreasurySpend: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Authority", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Authority = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field SpendId", wireType)
			}
			m.SpendId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Metadata", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Metadata = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
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
func (m *MsgRejectTreasurySpendResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
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
			return fmt.Errorf("proto: MsgRejectTreasurySpendResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgRejectTreasurySpendResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Spend", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTx
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
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
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
func (m *MsgExecuteTreasurySpend) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
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
			return fmt.Errorf("proto: MsgExecuteTreasurySpend: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgExecuteTreasurySpend: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Authority", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Authority = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field SpendId", wireType)
			}
			m.SpendId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Epoch", wireType)
			}
			m.Epoch = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Epoch |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
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
func (m *MsgExecuteTreasurySpendResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
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
			return fmt.Errorf("proto: MsgExecuteTreasurySpendResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgExecuteTreasurySpendResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Spend", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTx
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
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
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
func (m *MsgCancelTreasurySpend) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
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
			return fmt.Errorf("proto: MsgCancelTreasurySpend: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgCancelTreasurySpend: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Actor", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Actor = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field SpendId", wireType)
			}
			m.SpendId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Metadata", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Metadata = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
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
func (m *MsgCancelTreasurySpendResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
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
			return fmt.Errorf("proto: MsgCancelTreasurySpendResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgCancelTreasurySpendResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Spend", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTx
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
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
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
func (m *MsgUpdateTreasuryParams) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
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
			return fmt.Errorf("proto: MsgUpdateTreasuryParams: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgUpdateTreasuryParams: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Authority", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Authority = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Params", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTx
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
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
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
func (m *MsgUpdateTreasuryParamsResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
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
			return fmt.Errorf("proto: MsgUpdateTreasuryParamsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgUpdateTreasuryParamsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
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
func skipTx(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowTx
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
					return 0, ErrIntOverflowTx
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
					return 0, ErrIntOverflowTx
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
				return 0, ErrInvalidLengthTx
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupTx
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthTx
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthTx		= fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowTx		= fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupTx	= fmt.Errorf("proto: unexpected end of group")
)
