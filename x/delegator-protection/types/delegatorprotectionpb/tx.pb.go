package delegatorprotectionpb

import (
	context "context"
	fmt "fmt"
	_ "github.com/cosmos/cosmos-sdk/types/msgservice"
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

type MsgSubmitDelegatorProtectionClaim struct {
	Delegator	string	`protobuf:"bytes,1,opt,name=delegator,proto3" json:"delegator,omitempty"`
	Validator	string	`protobuf:"bytes,2,opt,name=validator,proto3" json:"validator,omitempty"`
	LossAmount	string	`protobuf:"bytes,3,opt,name=loss_amount,json=lossAmount,proto3" json:"loss_amount,omitempty"`
	RequestedPayout	string	`protobuf:"bytes,4,opt,name=requested_payout,json=requestedPayout,proto3" json:"requested_payout,omitempty"`
	EligibilityHash	string	`protobuf:"bytes,5,opt,name=eligibility_hash,json=eligibilityHash,proto3" json:"eligibility_hash,omitempty"`
	Reason		string	`protobuf:"bytes,6,opt,name=reason,proto3" json:"reason,omitempty"`
	Epoch		uint64	`protobuf:"varint,7,opt,name=epoch,proto3" json:"epoch,omitempty"`
	Height		uint64	`protobuf:"varint,8,opt,name=height,proto3" json:"height,omitempty"`
}

func (m *MsgSubmitDelegatorProtectionClaim) Reset()		{ *m = MsgSubmitDelegatorProtectionClaim{} }
func (m *MsgSubmitDelegatorProtectionClaim) String() string	{ return proto.CompactTextString(m) }
func (*MsgSubmitDelegatorProtectionClaim) ProtoMessage()	{}
func (*MsgSubmitDelegatorProtectionClaim) Descriptor() ([]byte, []int) {
	return fileDescriptor_2a1bbe4fd57b5b40, []int{0}
}
func (m *MsgSubmitDelegatorProtectionClaim) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgSubmitDelegatorProtectionClaim) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgSubmitDelegatorProtectionClaim.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgSubmitDelegatorProtectionClaim) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgSubmitDelegatorProtectionClaim.Merge(m, src)
}
func (m *MsgSubmitDelegatorProtectionClaim) XXX_Size() int {
	return m.Size()
}
func (m *MsgSubmitDelegatorProtectionClaim) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgSubmitDelegatorProtectionClaim.DiscardUnknown(m)
}

var xxx_messageInfo_MsgSubmitDelegatorProtectionClaim proto.InternalMessageInfo

func (m *MsgSubmitDelegatorProtectionClaim) GetDelegator() string {
	if m != nil {
		return m.Delegator
	}
	return ""
}

func (m *MsgSubmitDelegatorProtectionClaim) GetValidator() string {
	if m != nil {
		return m.Validator
	}
	return ""
}

func (m *MsgSubmitDelegatorProtectionClaim) GetLossAmount() string {
	if m != nil {
		return m.LossAmount
	}
	return ""
}

func (m *MsgSubmitDelegatorProtectionClaim) GetRequestedPayout() string {
	if m != nil {
		return m.RequestedPayout
	}
	return ""
}

func (m *MsgSubmitDelegatorProtectionClaim) GetEligibilityHash() string {
	if m != nil {
		return m.EligibilityHash
	}
	return ""
}

func (m *MsgSubmitDelegatorProtectionClaim) GetReason() string {
	if m != nil {
		return m.Reason
	}
	return ""
}

func (m *MsgSubmitDelegatorProtectionClaim) GetEpoch() uint64 {
	if m != nil {
		return m.Epoch
	}
	return 0
}

func (m *MsgSubmitDelegatorProtectionClaim) GetHeight() uint64 {
	if m != nil {
		return m.Height
	}
	return 0
}

type MsgSubmitDelegatorProtectionClaimResponse struct {
	ClaimJson string `protobuf:"bytes,1,opt,name=claim_json,json=claimJson,proto3" json:"claim_json,omitempty"`
}

func (m *MsgSubmitDelegatorProtectionClaimResponse) Reset() {
	*m = MsgSubmitDelegatorProtectionClaimResponse{}
}
func (m *MsgSubmitDelegatorProtectionClaimResponse) String() string {
	return proto.CompactTextString(m)
}
func (*MsgSubmitDelegatorProtectionClaimResponse) ProtoMessage()	{}
func (*MsgSubmitDelegatorProtectionClaimResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_2a1bbe4fd57b5b40, []int{1}
}
func (m *MsgSubmitDelegatorProtectionClaimResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgSubmitDelegatorProtectionClaimResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgSubmitDelegatorProtectionClaimResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgSubmitDelegatorProtectionClaimResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgSubmitDelegatorProtectionClaimResponse.Merge(m, src)
}
func (m *MsgSubmitDelegatorProtectionClaimResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgSubmitDelegatorProtectionClaimResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgSubmitDelegatorProtectionClaimResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgSubmitDelegatorProtectionClaimResponse proto.InternalMessageInfo

func (m *MsgSubmitDelegatorProtectionClaimResponse) GetClaimJson() string {
	if m != nil {
		return m.ClaimJson
	}
	return ""
}

type MsgApproveDelegatorProtectionClaim struct {
	Authority	string	`protobuf:"bytes,1,opt,name=authority,proto3" json:"authority,omitempty"`
	ClaimId		string	`protobuf:"bytes,2,opt,name=claim_id,json=claimId,proto3" json:"claim_id,omitempty"`
	ApprovedPayout	string	`protobuf:"bytes,3,opt,name=approved_payout,json=approvedPayout,proto3" json:"approved_payout,omitempty"`
	Height		uint64	`protobuf:"varint,4,opt,name=height,proto3" json:"height,omitempty"`
}

func (m *MsgApproveDelegatorProtectionClaim) Reset()		{ *m = MsgApproveDelegatorProtectionClaim{} }
func (m *MsgApproveDelegatorProtectionClaim) String() string	{ return proto.CompactTextString(m) }
func (*MsgApproveDelegatorProtectionClaim) ProtoMessage()	{}
func (*MsgApproveDelegatorProtectionClaim) Descriptor() ([]byte, []int) {
	return fileDescriptor_2a1bbe4fd57b5b40, []int{2}
}
func (m *MsgApproveDelegatorProtectionClaim) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgApproveDelegatorProtectionClaim) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgApproveDelegatorProtectionClaim.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgApproveDelegatorProtectionClaim) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgApproveDelegatorProtectionClaim.Merge(m, src)
}
func (m *MsgApproveDelegatorProtectionClaim) XXX_Size() int {
	return m.Size()
}
func (m *MsgApproveDelegatorProtectionClaim) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgApproveDelegatorProtectionClaim.DiscardUnknown(m)
}

var xxx_messageInfo_MsgApproveDelegatorProtectionClaim proto.InternalMessageInfo

func (m *MsgApproveDelegatorProtectionClaim) GetAuthority() string {
	if m != nil {
		return m.Authority
	}
	return ""
}

func (m *MsgApproveDelegatorProtectionClaim) GetClaimId() string {
	if m != nil {
		return m.ClaimId
	}
	return ""
}

func (m *MsgApproveDelegatorProtectionClaim) GetApprovedPayout() string {
	if m != nil {
		return m.ApprovedPayout
	}
	return ""
}

func (m *MsgApproveDelegatorProtectionClaim) GetHeight() uint64 {
	if m != nil {
		return m.Height
	}
	return 0
}

type MsgApproveDelegatorProtectionClaimResponse struct {
}

func (m *MsgApproveDelegatorProtectionClaimResponse) Reset() {
	*m = MsgApproveDelegatorProtectionClaimResponse{}
}
func (m *MsgApproveDelegatorProtectionClaimResponse) String() string {
	return proto.CompactTextString(m)
}
func (*MsgApproveDelegatorProtectionClaimResponse) ProtoMessage()	{}
func (*MsgApproveDelegatorProtectionClaimResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_2a1bbe4fd57b5b40, []int{3}
}
func (m *MsgApproveDelegatorProtectionClaimResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgApproveDelegatorProtectionClaimResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgApproveDelegatorProtectionClaimResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgApproveDelegatorProtectionClaimResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgApproveDelegatorProtectionClaimResponse.Merge(m, src)
}
func (m *MsgApproveDelegatorProtectionClaimResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgApproveDelegatorProtectionClaimResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgApproveDelegatorProtectionClaimResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgApproveDelegatorProtectionClaimResponse proto.InternalMessageInfo

type MsgRejectDelegatorProtectionClaim struct {
	Authority	string	`protobuf:"bytes,1,opt,name=authority,proto3" json:"authority,omitempty"`
	ClaimId		string	`protobuf:"bytes,2,opt,name=claim_id,json=claimId,proto3" json:"claim_id,omitempty"`
	Reason		string	`protobuf:"bytes,3,opt,name=reason,proto3" json:"reason,omitempty"`
	Height		uint64	`protobuf:"varint,4,opt,name=height,proto3" json:"height,omitempty"`
}

func (m *MsgRejectDelegatorProtectionClaim) Reset()		{ *m = MsgRejectDelegatorProtectionClaim{} }
func (m *MsgRejectDelegatorProtectionClaim) String() string	{ return proto.CompactTextString(m) }
func (*MsgRejectDelegatorProtectionClaim) ProtoMessage()	{}
func (*MsgRejectDelegatorProtectionClaim) Descriptor() ([]byte, []int) {
	return fileDescriptor_2a1bbe4fd57b5b40, []int{4}
}
func (m *MsgRejectDelegatorProtectionClaim) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgRejectDelegatorProtectionClaim) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgRejectDelegatorProtectionClaim.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgRejectDelegatorProtectionClaim) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgRejectDelegatorProtectionClaim.Merge(m, src)
}
func (m *MsgRejectDelegatorProtectionClaim) XXX_Size() int {
	return m.Size()
}
func (m *MsgRejectDelegatorProtectionClaim) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgRejectDelegatorProtectionClaim.DiscardUnknown(m)
}

var xxx_messageInfo_MsgRejectDelegatorProtectionClaim proto.InternalMessageInfo

func (m *MsgRejectDelegatorProtectionClaim) GetAuthority() string {
	if m != nil {
		return m.Authority
	}
	return ""
}

func (m *MsgRejectDelegatorProtectionClaim) GetClaimId() string {
	if m != nil {
		return m.ClaimId
	}
	return ""
}

func (m *MsgRejectDelegatorProtectionClaim) GetReason() string {
	if m != nil {
		return m.Reason
	}
	return ""
}

func (m *MsgRejectDelegatorProtectionClaim) GetHeight() uint64 {
	if m != nil {
		return m.Height
	}
	return 0
}

type MsgRejectDelegatorProtectionClaimResponse struct {
}

func (m *MsgRejectDelegatorProtectionClaimResponse) Reset() {
	*m = MsgRejectDelegatorProtectionClaimResponse{}
}
func (m *MsgRejectDelegatorProtectionClaimResponse) String() string {
	return proto.CompactTextString(m)
}
func (*MsgRejectDelegatorProtectionClaimResponse) ProtoMessage()	{}
func (*MsgRejectDelegatorProtectionClaimResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_2a1bbe4fd57b5b40, []int{5}
}
func (m *MsgRejectDelegatorProtectionClaimResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgRejectDelegatorProtectionClaimResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgRejectDelegatorProtectionClaimResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgRejectDelegatorProtectionClaimResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgRejectDelegatorProtectionClaimResponse.Merge(m, src)
}
func (m *MsgRejectDelegatorProtectionClaimResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgRejectDelegatorProtectionClaimResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgRejectDelegatorProtectionClaimResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgRejectDelegatorProtectionClaimResponse proto.InternalMessageInfo

type MsgClaimDelegatorCompensation struct {
	Delegator	string	`protobuf:"bytes,1,opt,name=delegator,proto3" json:"delegator,omitempty"`
	ClaimId		string	`protobuf:"bytes,2,opt,name=claim_id,json=claimId,proto3" json:"claim_id,omitempty"`
	Epoch		uint64	`protobuf:"varint,3,opt,name=epoch,proto3" json:"epoch,omitempty"`
	Height		uint64	`protobuf:"varint,4,opt,name=height,proto3" json:"height,omitempty"`
}

func (m *MsgClaimDelegatorCompensation) Reset()		{ *m = MsgClaimDelegatorCompensation{} }
func (m *MsgClaimDelegatorCompensation) String() string	{ return proto.CompactTextString(m) }
func (*MsgClaimDelegatorCompensation) ProtoMessage()	{}
func (*MsgClaimDelegatorCompensation) Descriptor() ([]byte, []int) {
	return fileDescriptor_2a1bbe4fd57b5b40, []int{6}
}
func (m *MsgClaimDelegatorCompensation) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgClaimDelegatorCompensation) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgClaimDelegatorCompensation.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgClaimDelegatorCompensation) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgClaimDelegatorCompensation.Merge(m, src)
}
func (m *MsgClaimDelegatorCompensation) XXX_Size() int {
	return m.Size()
}
func (m *MsgClaimDelegatorCompensation) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgClaimDelegatorCompensation.DiscardUnknown(m)
}

var xxx_messageInfo_MsgClaimDelegatorCompensation proto.InternalMessageInfo

func (m *MsgClaimDelegatorCompensation) GetDelegator() string {
	if m != nil {
		return m.Delegator
	}
	return ""
}

func (m *MsgClaimDelegatorCompensation) GetClaimId() string {
	if m != nil {
		return m.ClaimId
	}
	return ""
}

func (m *MsgClaimDelegatorCompensation) GetEpoch() uint64 {
	if m != nil {
		return m.Epoch
	}
	return 0
}

func (m *MsgClaimDelegatorCompensation) GetHeight() uint64 {
	if m != nil {
		return m.Height
	}
	return 0
}

type MsgClaimDelegatorCompensationResponse struct {
	PayoutJson string `protobuf:"bytes,1,opt,name=payout_json,json=payoutJson,proto3" json:"payout_json,omitempty"`
}

func (m *MsgClaimDelegatorCompensationResponse) Reset()		{ *m = MsgClaimDelegatorCompensationResponse{} }
func (m *MsgClaimDelegatorCompensationResponse) String() string	{ return proto.CompactTextString(m) }
func (*MsgClaimDelegatorCompensationResponse) ProtoMessage()	{}
func (*MsgClaimDelegatorCompensationResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_2a1bbe4fd57b5b40, []int{7}
}
func (m *MsgClaimDelegatorCompensationResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgClaimDelegatorCompensationResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgClaimDelegatorCompensationResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgClaimDelegatorCompensationResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgClaimDelegatorCompensationResponse.Merge(m, src)
}
func (m *MsgClaimDelegatorCompensationResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgClaimDelegatorCompensationResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgClaimDelegatorCompensationResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgClaimDelegatorCompensationResponse proto.InternalMessageInfo

func (m *MsgClaimDelegatorCompensationResponse) GetPayoutJson() string {
	if m != nil {
		return m.PayoutJson
	}
	return ""
}

type MsgUpdateProtectionParams struct {
	Authority	string	`protobuf:"bytes,1,opt,name=authority,proto3" json:"authority,omitempty"`
	ParamsJson	string	`protobuf:"bytes,2,opt,name=params_json,json=paramsJson,proto3" json:"params_json,omitempty"`
}

func (m *MsgUpdateProtectionParams) Reset()		{ *m = MsgUpdateProtectionParams{} }
func (m *MsgUpdateProtectionParams) String() string	{ return proto.CompactTextString(m) }
func (*MsgUpdateProtectionParams) ProtoMessage()	{}
func (*MsgUpdateProtectionParams) Descriptor() ([]byte, []int) {
	return fileDescriptor_2a1bbe4fd57b5b40, []int{8}
}
func (m *MsgUpdateProtectionParams) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgUpdateProtectionParams) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgUpdateProtectionParams.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgUpdateProtectionParams) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgUpdateProtectionParams.Merge(m, src)
}
func (m *MsgUpdateProtectionParams) XXX_Size() int {
	return m.Size()
}
func (m *MsgUpdateProtectionParams) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgUpdateProtectionParams.DiscardUnknown(m)
}

var xxx_messageInfo_MsgUpdateProtectionParams proto.InternalMessageInfo

func (m *MsgUpdateProtectionParams) GetAuthority() string {
	if m != nil {
		return m.Authority
	}
	return ""
}

func (m *MsgUpdateProtectionParams) GetParamsJson() string {
	if m != nil {
		return m.ParamsJson
	}
	return ""
}

type MsgUpdateProtectionParamsResponse struct {
}

func (m *MsgUpdateProtectionParamsResponse) Reset()		{ *m = MsgUpdateProtectionParamsResponse{} }
func (m *MsgUpdateProtectionParamsResponse) String() string	{ return proto.CompactTextString(m) }
func (*MsgUpdateProtectionParamsResponse) ProtoMessage()	{}
func (*MsgUpdateProtectionParamsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_2a1bbe4fd57b5b40, []int{9}
}
func (m *MsgUpdateProtectionParamsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgUpdateProtectionParamsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgUpdateProtectionParamsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgUpdateProtectionParamsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgUpdateProtectionParamsResponse.Merge(m, src)
}
func (m *MsgUpdateProtectionParamsResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgUpdateProtectionParamsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgUpdateProtectionParamsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgUpdateProtectionParamsResponse proto.InternalMessageInfo

func init() {
	proto.RegisterType((*MsgSubmitDelegatorProtectionClaim)(nil), "l1.delegatorprotection.v1.MsgSubmitDelegatorProtectionClaim")
	proto.RegisterType((*MsgSubmitDelegatorProtectionClaimResponse)(nil), "l1.delegatorprotection.v1.MsgSubmitDelegatorProtectionClaimResponse")
	proto.RegisterType((*MsgApproveDelegatorProtectionClaim)(nil), "l1.delegatorprotection.v1.MsgApproveDelegatorProtectionClaim")
	proto.RegisterType((*MsgApproveDelegatorProtectionClaimResponse)(nil), "l1.delegatorprotection.v1.MsgApproveDelegatorProtectionClaimResponse")
	proto.RegisterType((*MsgRejectDelegatorProtectionClaim)(nil), "l1.delegatorprotection.v1.MsgRejectDelegatorProtectionClaim")
	proto.RegisterType((*MsgRejectDelegatorProtectionClaimResponse)(nil), "l1.delegatorprotection.v1.MsgRejectDelegatorProtectionClaimResponse")
	proto.RegisterType((*MsgClaimDelegatorCompensation)(nil), "l1.delegatorprotection.v1.MsgClaimDelegatorCompensation")
	proto.RegisterType((*MsgClaimDelegatorCompensationResponse)(nil), "l1.delegatorprotection.v1.MsgClaimDelegatorCompensationResponse")
	proto.RegisterType((*MsgUpdateProtectionParams)(nil), "l1.delegatorprotection.v1.MsgUpdateProtectionParams")
	proto.RegisterType((*MsgUpdateProtectionParamsResponse)(nil), "l1.delegatorprotection.v1.MsgUpdateProtectionParamsResponse")
}

func init() {
	proto.RegisterFile("l1/delegatorprotection/v1/tx.proto", fileDescriptor_2a1bbe4fd57b5b40)
}

var fileDescriptor_2a1bbe4fd57b5b40 = []byte{

	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x56, 0xc1, 0x4f, 0x13, 0x4f,
	0x14, 0x66, 0x68, 0x29, 0xf0, 0x48, 0xe0, 0x97, 0xcd, 0x2f, 0xb8, 0x6c, 0x64, 0xc1, 0x35, 0x46,
	0x40, 0xe9, 0xa6, 0xea, 0xc1, 0x18, 0x4c, 0x44, 0x30, 0x41, 0x92, 0x26, 0xa4, 0xc6, 0x8b, 0x97,
	0x66, 0xda, 0x4e, 0x76, 0x87, 0xec, 0xee, 0xac, 0x3b, 0xd3, 0x86, 0xde, 0x8c, 0x67, 0x0f, 0x9e,
	0x8c, 0x07, 0x8f, 0x26, 0x5e, 0x0c, 0xe1, 0xcf, 0xf0, 0xc8, 0xd1, 0xa3, 0x81, 0x03, 0xff, 0x86,
	0x99, 0xd9, 0xee, 0x14, 0x63, 0xd9, 0x25, 0xe0, 0x71, 0xbe, 0x79, 0xdf, 0x9b, 0xf7, 0xbe, 0xf9,
	0xde, 0x64, 0xc0, 0x09, 0x6a, 0x6e, 0x87, 0x04, 0xc4, 0xc3, 0x82, 0x25, 0x71, 0xc2, 0x04, 0x69,
	0x0b, 0xca, 0x22, 0xb7, 0x57, 0x73, 0xc5, 0x41, 0x55, 0x02, 0xcc, 0x58, 0x08, 0x6a, 0xd5, 0x11,
	0x31, 0xd5, 0x5e, 0xcd, 0xba, 0xd1, 0x66, 0x3c, 0x64, 0xdc, 0x0d, 0xb9, 0x27, 0x29, 0x21, 0xf7,
	0x52, 0x8e, 0xf3, 0x7d, 0x1c, 0x6e, 0xd5, 0xb9, 0xf7, 0xaa, 0xdb, 0x0a, 0xa9, 0xd8, 0xce, 0xc8,
	0x7b, 0x9a, 0xbc, 0x15, 0x60, 0x1a, 0x1a, 0x37, 0x61, 0x5a, 0x27, 0x36, 0xd1, 0x32, 0x5a, 0x99,
	0x6e, 0x0c, 0x01, 0xb9, 0xdb, 0xc3, 0x01, 0xed, 0xa8, 0xdd, 0xf1, 0x74, 0x57, 0x03, 0xc6, 0x12,
	0xcc, 0x04, 0x8c, 0xf3, 0x26, 0x0e, 0x59, 0x37, 0x12, 0x66, 0x49, 0xed, 0x83, 0x84, 0x36, 0x15,
	0x62, 0xac, 0xc2, 0x7f, 0x09, 0x79, 0xdb, 0x25, 0x5c, 0x90, 0x4e, 0x33, 0xc6, 0x7d, 0xd6, 0x15,
	0x66, 0x59, 0x45, 0xcd, 0x69, 0x7c, 0x4f, 0xc1, 0x32, 0x94, 0x04, 0xd4, 0xa3, 0x2d, 0x1a, 0x50,
	0xd1, 0x6f, 0xfa, 0x98, 0xfb, 0xe6, 0x44, 0x1a, 0x7a, 0x0e, 0xdf, 0xc1, 0xdc, 0x37, 0xe6, 0xa1,
	0x92, 0x10, 0xcc, 0x59, 0x64, 0x56, 0x54, 0xc0, 0x60, 0x65, 0xfc, 0x0f, 0x13, 0x24, 0x66, 0x6d,
	0xdf, 0x9c, 0x5c, 0x46, 0x2b, 0xe5, 0x46, 0xba, 0x90, 0xd1, 0x3e, 0xa1, 0x9e, 0x2f, 0xcc, 0x29,
	0x05, 0x0f, 0x56, 0x4f, 0x66, 0xdf, 0x9f, 0x1d, 0xad, 0x0d, 0x5b, 0x75, 0x76, 0x61, 0xb5, 0x50,
	0xad, 0x06, 0xe1, 0x31, 0x8b, 0x38, 0x31, 0x16, 0x01, 0xda, 0x12, 0x68, 0xee, 0xcb, 0x32, 0x06,
	0xb2, 0x29, 0x64, 0x97, 0xb3, 0xc8, 0x39, 0x44, 0xe0, 0xd4, 0xb9, 0xb7, 0x19, 0xc7, 0x09, 0xeb,
	0x91, 0x3c, 0xed, 0x71, 0x57, 0xf8, 0x2c, 0xa1, 0xa2, 0x9f, 0x25, 0xd1, 0x80, 0xb1, 0x00, 0x53,
	0xe9, 0x19, 0xb4, 0x33, 0x90, 0x7e, 0x52, 0xad, 0x5f, 0x76, 0x8c, 0xbb, 0x30, 0x87, 0xd3, 0xdc,
	0x5a, 0xd6, 0x54, 0xfc, 0xd9, 0x0c, 0x1e, 0xa8, 0x3a, 0x6c, 0xbe, 0x3c, 0xa2, 0x79, 0x7d, 0x96,
	0x73, 0x1f, 0xd6, 0x8a, 0xeb, 0xcd, 0xba, 0x77, 0xbe, 0x20, 0xe5, 0xac, 0x06, 0xd9, 0x27, 0x6d,
	0xf1, 0xef, 0xbb, 0x1b, 0xde, 0x6f, 0xe9, 0x8f, 0xfb, 0xbd, 0x6c, 0x33, 0xf7, 0xd4, 0x4d, 0xe6,
	0x57, 0xa7, 0x7b, 0xf9, 0x84, 0x60, 0xb1, 0xce, 0x3d, 0x05, 0xea, 0xe0, 0x2d, 0x16, 0xc6, 0x24,
	0xe2, 0x58, 0x86, 0x17, 0x4c, 0x48, 0x4e, 0x1f, 0xda, 0x8f, 0xa5, 0xd1, 0x7e, 0x2c, 0xe7, 0xfa,
	0x71, 0x07, 0xee, 0xe4, 0xd6, 0xa5, 0xbd, 0xb8, 0x04, 0x33, 0xa9, 0x07, 0xce, 0x9b, 0x11, 0x52,
	0x48, 0xb9, 0x71, 0x1f, 0x16, 0xea, 0xdc, 0x7b, 0x1d, 0x77, 0xb0, 0x20, 0x43, 0x19, 0xf6, 0x70,
	0x82, 0x43, 0x5e, 0x70, 0x4b, 0x2a, 0xb7, 0x8c, 0x4b, 0x73, 0x8f, 0x67, 0xb9, 0x25, 0x24, 0x73,
	0xff, 0xa5, 0xfd, 0x6d, 0xe5, 0x8c, 0xd1, 0x67, 0x65, 0x15, 0x3f, 0x38, 0xac, 0x40, 0xa9, 0xce,
	0x3d, 0xe3, 0x2b, 0x02, 0xbb, 0xe0, 0x79, 0xda, 0xa8, 0x5e, 0xf8, 0xf2, 0x55, 0x0b, 0xc7, 0xd5,
	0xda, 0xbe, 0x0e, 0x5b, 0x0b, 0xfc, 0x0d, 0xc1, 0x52, 0xd1, 0x28, 0x3f, 0xcd, 0x3f, 0xa9, 0x80,
	0x6e, 0xbd, 0xb8, 0x16, 0x5d, 0x57, 0x2a, 0x05, 0x2d, 0x98, 0xca, 0x02, 0x41, 0xf3, 0xd9, 0x45,
	0x82, 0x5e, 0x6e, 0xe6, 0x8c, 0xcf, 0x08, 0xac, 0x9c, 0x81, 0x7b, 0x9c, 0x7f, 0xc8, 0xc5, 0x4c,
	0xeb, 0xd9, 0x55, 0x99, 0xba, 0xb4, 0x0f, 0x08, 0xe6, 0x2f, 0x98, 0x94, 0x47, 0xf9, 0xc9, 0x47,
	0xb3, 0xac, 0x8d, 0xab, 0xb0, 0xb2, 0x72, 0xac, 0x89, 0x77, 0x67, 0x47, 0x6b, 0xe8, 0xb9, 0xf7,
	0xe3, 0xc4, 0x46, 0xc7, 0x27, 0x36, 0xfa, 0x75, 0x62, 0xa3, 0x8f, 0xa7, 0xf6, 0xd8, 0xf1, 0xa9,
	0x3d, 0xf6, 0xf3, 0xd4, 0x1e, 0x7b, 0x53, 0xf7, 0xa8, 0xf0, 0xbb, 0xad, 0x6a, 0x9b, 0x85, 0x2e,
	0x67, 0x3d, 0x92, 0x10, 0xea, 0x45, 0xeb, 0x41, 0xcd, 0x0d, 0x6a, 0xee, 0xc1, 0xf0, 0x5b, 0xb1,
	0x7e, 0xee, 0x5f, 0x21, 0xfa, 0x31, 0xe1, 0xa3, 0x7e, 0x1c, 0x71, 0xab, 0x55, 0x51, 0x5f, 0x87,
	0x87, 0xbf, 0x03, 0x00, 0x00, 0xff, 0xff, 0xe2, 0x70, 0x93, 0xec, 0x94, 0x08, 0x00, 0x00,
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
	SubmitDelegatorProtectionClaim(ctx context.Context, in *MsgSubmitDelegatorProtectionClaim, opts ...grpc.CallOption) (*MsgSubmitDelegatorProtectionClaimResponse, error)
	ApproveDelegatorProtectionClaim(ctx context.Context, in *MsgApproveDelegatorProtectionClaim, opts ...grpc.CallOption) (*MsgApproveDelegatorProtectionClaimResponse, error)
	RejectDelegatorProtectionClaim(ctx context.Context, in *MsgRejectDelegatorProtectionClaim, opts ...grpc.CallOption) (*MsgRejectDelegatorProtectionClaimResponse, error)
	ClaimDelegatorCompensation(ctx context.Context, in *MsgClaimDelegatorCompensation, opts ...grpc.CallOption) (*MsgClaimDelegatorCompensationResponse, error)
	UpdateProtectionParams(ctx context.Context, in *MsgUpdateProtectionParams, opts ...grpc.CallOption) (*MsgUpdateProtectionParamsResponse, error)
}

type msgClient struct {
	cc grpc1.ClientConn
}

func NewMsgClient(cc grpc1.ClientConn) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) SubmitDelegatorProtectionClaim(ctx context.Context, in *MsgSubmitDelegatorProtectionClaim, opts ...grpc.CallOption) (*MsgSubmitDelegatorProtectionClaimResponse, error) {
	out := new(MsgSubmitDelegatorProtectionClaimResponse)
	err := c.cc.Invoke(ctx, "/l1.delegatorprotection.v1.Msg/SubmitDelegatorProtectionClaim", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) ApproveDelegatorProtectionClaim(ctx context.Context, in *MsgApproveDelegatorProtectionClaim, opts ...grpc.CallOption) (*MsgApproveDelegatorProtectionClaimResponse, error) {
	out := new(MsgApproveDelegatorProtectionClaimResponse)
	err := c.cc.Invoke(ctx, "/l1.delegatorprotection.v1.Msg/ApproveDelegatorProtectionClaim", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) RejectDelegatorProtectionClaim(ctx context.Context, in *MsgRejectDelegatorProtectionClaim, opts ...grpc.CallOption) (*MsgRejectDelegatorProtectionClaimResponse, error) {
	out := new(MsgRejectDelegatorProtectionClaimResponse)
	err := c.cc.Invoke(ctx, "/l1.delegatorprotection.v1.Msg/RejectDelegatorProtectionClaim", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) ClaimDelegatorCompensation(ctx context.Context, in *MsgClaimDelegatorCompensation, opts ...grpc.CallOption) (*MsgClaimDelegatorCompensationResponse, error) {
	out := new(MsgClaimDelegatorCompensationResponse)
	err := c.cc.Invoke(ctx, "/l1.delegatorprotection.v1.Msg/ClaimDelegatorCompensation", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) UpdateProtectionParams(ctx context.Context, in *MsgUpdateProtectionParams, opts ...grpc.CallOption) (*MsgUpdateProtectionParamsResponse, error) {
	out := new(MsgUpdateProtectionParamsResponse)
	err := c.cc.Invoke(ctx, "/l1.delegatorprotection.v1.Msg/UpdateProtectionParams", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MsgServer is the server API for Msg service.
type MsgServer interface {
	SubmitDelegatorProtectionClaim(context.Context, *MsgSubmitDelegatorProtectionClaim) (*MsgSubmitDelegatorProtectionClaimResponse, error)
	ApproveDelegatorProtectionClaim(context.Context, *MsgApproveDelegatorProtectionClaim) (*MsgApproveDelegatorProtectionClaimResponse, error)
	RejectDelegatorProtectionClaim(context.Context, *MsgRejectDelegatorProtectionClaim) (*MsgRejectDelegatorProtectionClaimResponse, error)
	ClaimDelegatorCompensation(context.Context, *MsgClaimDelegatorCompensation) (*MsgClaimDelegatorCompensationResponse, error)
	UpdateProtectionParams(context.Context, *MsgUpdateProtectionParams) (*MsgUpdateProtectionParamsResponse, error)
}

// UnimplementedMsgServer can be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (*UnimplementedMsgServer) SubmitDelegatorProtectionClaim(ctx context.Context, req *MsgSubmitDelegatorProtectionClaim) (*MsgSubmitDelegatorProtectionClaimResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitDelegatorProtectionClaim not implemented")
}
func (*UnimplementedMsgServer) ApproveDelegatorProtectionClaim(ctx context.Context, req *MsgApproveDelegatorProtectionClaim) (*MsgApproveDelegatorProtectionClaimResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ApproveDelegatorProtectionClaim not implemented")
}
func (*UnimplementedMsgServer) RejectDelegatorProtectionClaim(ctx context.Context, req *MsgRejectDelegatorProtectionClaim) (*MsgRejectDelegatorProtectionClaimResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RejectDelegatorProtectionClaim not implemented")
}
func (*UnimplementedMsgServer) ClaimDelegatorCompensation(ctx context.Context, req *MsgClaimDelegatorCompensation) (*MsgClaimDelegatorCompensationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ClaimDelegatorCompensation not implemented")
}
func (*UnimplementedMsgServer) UpdateProtectionParams(ctx context.Context, req *MsgUpdateProtectionParams) (*MsgUpdateProtectionParamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateProtectionParams not implemented")
}

func RegisterMsgServer(s grpc1.Server, srv MsgServer) {
	s.RegisterService(&_Msg_serviceDesc, srv)
}

func _Msg_SubmitDelegatorProtectionClaim_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgSubmitDelegatorProtectionClaim)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).SubmitDelegatorProtectionClaim(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.delegatorprotection.v1.Msg/SubmitDelegatorProtectionClaim",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).SubmitDelegatorProtectionClaim(ctx, req.(*MsgSubmitDelegatorProtectionClaim))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_ApproveDelegatorProtectionClaim_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgApproveDelegatorProtectionClaim)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).ApproveDelegatorProtectionClaim(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.delegatorprotection.v1.Msg/ApproveDelegatorProtectionClaim",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).ApproveDelegatorProtectionClaim(ctx, req.(*MsgApproveDelegatorProtectionClaim))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_RejectDelegatorProtectionClaim_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgRejectDelegatorProtectionClaim)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).RejectDelegatorProtectionClaim(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.delegatorprotection.v1.Msg/RejectDelegatorProtectionClaim",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).RejectDelegatorProtectionClaim(ctx, req.(*MsgRejectDelegatorProtectionClaim))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_ClaimDelegatorCompensation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgClaimDelegatorCompensation)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).ClaimDelegatorCompensation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.delegatorprotection.v1.Msg/ClaimDelegatorCompensation",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).ClaimDelegatorCompensation(ctx, req.(*MsgClaimDelegatorCompensation))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_UpdateProtectionParams_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUpdateProtectionParams)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).UpdateProtectionParams(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.delegatorprotection.v1.Msg/UpdateProtectionParams",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).UpdateProtectionParams(ctx, req.(*MsgUpdateProtectionParams))
	}
	return interceptor(ctx, in, info, handler)
}

var Msg_serviceDesc = _Msg_serviceDesc
var _Msg_serviceDesc = grpc.ServiceDesc{
	ServiceName:	"l1.delegatorprotection.v1.Msg",
	HandlerType:	(*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName:	"SubmitDelegatorProtectionClaim",
			Handler:	_Msg_SubmitDelegatorProtectionClaim_Handler,
		},
		{
			MethodName:	"ApproveDelegatorProtectionClaim",
			Handler:	_Msg_ApproveDelegatorProtectionClaim_Handler,
		},
		{
			MethodName:	"RejectDelegatorProtectionClaim",
			Handler:	_Msg_RejectDelegatorProtectionClaim_Handler,
		},
		{
			MethodName:	"ClaimDelegatorCompensation",
			Handler:	_Msg_ClaimDelegatorCompensation_Handler,
		},
		{
			MethodName:	"UpdateProtectionParams",
			Handler:	_Msg_UpdateProtectionParams_Handler,
		},
	},
	Streams:	[]grpc.StreamDesc{},
	Metadata:	"l1/delegatorprotection/v1/tx.proto",
}

func (m *MsgSubmitDelegatorProtectionClaim) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgSubmitDelegatorProtectionClaim) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgSubmitDelegatorProtectionClaim) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Height != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.Height))
		i--
		dAtA[i] = 0x40
	}
	if m.Epoch != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.Epoch))
		i--
		dAtA[i] = 0x38
	}
	if len(m.Reason) > 0 {
		i -= len(m.Reason)
		copy(dAtA[i:], m.Reason)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Reason)))
		i--
		dAtA[i] = 0x32
	}
	if len(m.EligibilityHash) > 0 {
		i -= len(m.EligibilityHash)
		copy(dAtA[i:], m.EligibilityHash)
		i = encodeVarintTx(dAtA, i, uint64(len(m.EligibilityHash)))
		i--
		dAtA[i] = 0x2a
	}
	if len(m.RequestedPayout) > 0 {
		i -= len(m.RequestedPayout)
		copy(dAtA[i:], m.RequestedPayout)
		i = encodeVarintTx(dAtA, i, uint64(len(m.RequestedPayout)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.LossAmount) > 0 {
		i -= len(m.LossAmount)
		copy(dAtA[i:], m.LossAmount)
		i = encodeVarintTx(dAtA, i, uint64(len(m.LossAmount)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Validator) > 0 {
		i -= len(m.Validator)
		copy(dAtA[i:], m.Validator)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Validator)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Delegator) > 0 {
		i -= len(m.Delegator)
		copy(dAtA[i:], m.Delegator)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Delegator)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MsgSubmitDelegatorProtectionClaimResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgSubmitDelegatorProtectionClaimResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgSubmitDelegatorProtectionClaimResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.ClaimJson) > 0 {
		i -= len(m.ClaimJson)
		copy(dAtA[i:], m.ClaimJson)
		i = encodeVarintTx(dAtA, i, uint64(len(m.ClaimJson)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MsgApproveDelegatorProtectionClaim) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgApproveDelegatorProtectionClaim) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgApproveDelegatorProtectionClaim) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Height != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.Height))
		i--
		dAtA[i] = 0x20
	}
	if len(m.ApprovedPayout) > 0 {
		i -= len(m.ApprovedPayout)
		copy(dAtA[i:], m.ApprovedPayout)
		i = encodeVarintTx(dAtA, i, uint64(len(m.ApprovedPayout)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.ClaimId) > 0 {
		i -= len(m.ClaimId)
		copy(dAtA[i:], m.ClaimId)
		i = encodeVarintTx(dAtA, i, uint64(len(m.ClaimId)))
		i--
		dAtA[i] = 0x12
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

func (m *MsgApproveDelegatorProtectionClaimResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgApproveDelegatorProtectionClaimResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgApproveDelegatorProtectionClaimResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *MsgRejectDelegatorProtectionClaim) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgRejectDelegatorProtectionClaim) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgRejectDelegatorProtectionClaim) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Height != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.Height))
		i--
		dAtA[i] = 0x20
	}
	if len(m.Reason) > 0 {
		i -= len(m.Reason)
		copy(dAtA[i:], m.Reason)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Reason)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.ClaimId) > 0 {
		i -= len(m.ClaimId)
		copy(dAtA[i:], m.ClaimId)
		i = encodeVarintTx(dAtA, i, uint64(len(m.ClaimId)))
		i--
		dAtA[i] = 0x12
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

func (m *MsgRejectDelegatorProtectionClaimResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgRejectDelegatorProtectionClaimResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgRejectDelegatorProtectionClaimResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *MsgClaimDelegatorCompensation) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgClaimDelegatorCompensation) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgClaimDelegatorCompensation) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Height != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.Height))
		i--
		dAtA[i] = 0x20
	}
	if m.Epoch != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.Epoch))
		i--
		dAtA[i] = 0x18
	}
	if len(m.ClaimId) > 0 {
		i -= len(m.ClaimId)
		copy(dAtA[i:], m.ClaimId)
		i = encodeVarintTx(dAtA, i, uint64(len(m.ClaimId)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Delegator) > 0 {
		i -= len(m.Delegator)
		copy(dAtA[i:], m.Delegator)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Delegator)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MsgClaimDelegatorCompensationResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgClaimDelegatorCompensationResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgClaimDelegatorCompensationResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.PayoutJson) > 0 {
		i -= len(m.PayoutJson)
		copy(dAtA[i:], m.PayoutJson)
		i = encodeVarintTx(dAtA, i, uint64(len(m.PayoutJson)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MsgUpdateProtectionParams) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgUpdateProtectionParams) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgUpdateProtectionParams) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.ParamsJson) > 0 {
		i -= len(m.ParamsJson)
		copy(dAtA[i:], m.ParamsJson)
		i = encodeVarintTx(dAtA, i, uint64(len(m.ParamsJson)))
		i--
		dAtA[i] = 0x12
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

func (m *MsgUpdateProtectionParamsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgUpdateProtectionParamsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgUpdateProtectionParamsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
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
func (m *MsgSubmitDelegatorProtectionClaim) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Delegator)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.Validator)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.LossAmount)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.RequestedPayout)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.EligibilityHash)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.Reason)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if m.Epoch != 0 {
		n += 1 + sovTx(uint64(m.Epoch))
	}
	if m.Height != 0 {
		n += 1 + sovTx(uint64(m.Height))
	}
	return n
}

func (m *MsgSubmitDelegatorProtectionClaimResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ClaimJson)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	return n
}

func (m *MsgApproveDelegatorProtectionClaim) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Authority)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.ClaimId)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.ApprovedPayout)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if m.Height != 0 {
		n += 1 + sovTx(uint64(m.Height))
	}
	return n
}

func (m *MsgApproveDelegatorProtectionClaimResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *MsgRejectDelegatorProtectionClaim) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Authority)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.ClaimId)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.Reason)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if m.Height != 0 {
		n += 1 + sovTx(uint64(m.Height))
	}
	return n
}

func (m *MsgRejectDelegatorProtectionClaimResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *MsgClaimDelegatorCompensation) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Delegator)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.ClaimId)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if m.Epoch != 0 {
		n += 1 + sovTx(uint64(m.Epoch))
	}
	if m.Height != 0 {
		n += 1 + sovTx(uint64(m.Height))
	}
	return n
}

func (m *MsgClaimDelegatorCompensationResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.PayoutJson)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	return n
}

func (m *MsgUpdateProtectionParams) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Authority)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.ParamsJson)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	return n
}

func (m *MsgUpdateProtectionParamsResponse) Size() (n int) {
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
func (m *MsgSubmitDelegatorProtectionClaim) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgSubmitDelegatorProtectionClaim: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgSubmitDelegatorProtectionClaim: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Delegator", wireType)
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
			m.Delegator = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Validator", wireType)
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
			m.Validator = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field LossAmount", wireType)
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
			m.LossAmount = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RequestedPayout", wireType)
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
			m.RequestedPayout = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field EligibilityHash", wireType)
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
			m.EligibilityHash = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Reason", wireType)
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
			m.Reason = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 7:
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
		case 8:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Height", wireType)
			}
			m.Height = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Height |= uint64(b&0x7F) << shift
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
func (m *MsgSubmitDelegatorProtectionClaimResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgSubmitDelegatorProtectionClaimResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgSubmitDelegatorProtectionClaimResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ClaimJson", wireType)
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
			m.ClaimJson = string(dAtA[iNdEx:postIndex])
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
func (m *MsgApproveDelegatorProtectionClaim) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgApproveDelegatorProtectionClaim: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgApproveDelegatorProtectionClaim: illegal tag %d (wire type %d)", fieldNum, wire)
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
				return fmt.Errorf("proto: wrong wireType = %d for field ClaimId", wireType)
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
			m.ClaimId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ApprovedPayout", wireType)
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
			m.ApprovedPayout = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Height", wireType)
			}
			m.Height = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Height |= uint64(b&0x7F) << shift
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
func (m *MsgApproveDelegatorProtectionClaimResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgApproveDelegatorProtectionClaimResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgApproveDelegatorProtectionClaimResponse: illegal tag %d (wire type %d)", fieldNum, wire)
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
func (m *MsgRejectDelegatorProtectionClaim) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgRejectDelegatorProtectionClaim: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgRejectDelegatorProtectionClaim: illegal tag %d (wire type %d)", fieldNum, wire)
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
				return fmt.Errorf("proto: wrong wireType = %d for field ClaimId", wireType)
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
			m.ClaimId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Reason", wireType)
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
			m.Reason = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Height", wireType)
			}
			m.Height = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Height |= uint64(b&0x7F) << shift
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
func (m *MsgRejectDelegatorProtectionClaimResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgRejectDelegatorProtectionClaimResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgRejectDelegatorProtectionClaimResponse: illegal tag %d (wire type %d)", fieldNum, wire)
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
func (m *MsgClaimDelegatorCompensation) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgClaimDelegatorCompensation: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgClaimDelegatorCompensation: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Delegator", wireType)
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
			m.Delegator = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ClaimId", wireType)
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
			m.ClaimId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
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
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Height", wireType)
			}
			m.Height = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Height |= uint64(b&0x7F) << shift
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
func (m *MsgClaimDelegatorCompensationResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgClaimDelegatorCompensationResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgClaimDelegatorCompensationResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PayoutJson", wireType)
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
			m.PayoutJson = string(dAtA[iNdEx:postIndex])
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
func (m *MsgUpdateProtectionParams) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgUpdateProtectionParams: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgUpdateProtectionParams: illegal tag %d (wire type %d)", fieldNum, wire)
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
				return fmt.Errorf("proto: wrong wireType = %d for field ParamsJson", wireType)
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
			m.ParamsJson = string(dAtA[iNdEx:postIndex])
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
func (m *MsgUpdateProtectionParamsResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgUpdateProtectionParamsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgUpdateProtectionParamsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
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
