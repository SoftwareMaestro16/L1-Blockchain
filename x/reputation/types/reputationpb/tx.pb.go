package reputationpb

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

type MsgUpdateReputationParams struct {
	Authority	string	`protobuf:"bytes,1,opt,name=authority,proto3" json:"authority,omitempty"`
	ParamsJson	string	`protobuf:"bytes,2,opt,name=params_json,json=paramsJson,proto3" json:"params_json,omitempty"`
}

func (m *MsgUpdateReputationParams) Reset()		{ *m = MsgUpdateReputationParams{} }
func (m *MsgUpdateReputationParams) String() string	{ return proto.CompactTextString(m) }
func (*MsgUpdateReputationParams) ProtoMessage()	{}
func (*MsgUpdateReputationParams) Descriptor() ([]byte, []int) {
	return fileDescriptor_77efba301b074cfb, []int{0}
}
func (m *MsgUpdateReputationParams) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgUpdateReputationParams) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgUpdateReputationParams.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgUpdateReputationParams) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgUpdateReputationParams.Merge(m, src)
}
func (m *MsgUpdateReputationParams) XXX_Size() int {
	return m.Size()
}
func (m *MsgUpdateReputationParams) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgUpdateReputationParams.DiscardUnknown(m)
}

var xxx_messageInfo_MsgUpdateReputationParams proto.InternalMessageInfo

func (m *MsgUpdateReputationParams) GetAuthority() string {
	if m != nil {
		return m.Authority
	}
	return ""
}

func (m *MsgUpdateReputationParams) GetParamsJson() string {
	if m != nil {
		return m.ParamsJson
	}
	return ""
}

type MsgUpdateReputationParamsResponse struct {
}

func (m *MsgUpdateReputationParamsResponse) Reset()		{ *m = MsgUpdateReputationParamsResponse{} }
func (m *MsgUpdateReputationParamsResponse) String() string	{ return proto.CompactTextString(m) }
func (*MsgUpdateReputationParamsResponse) ProtoMessage()	{}
func (*MsgUpdateReputationParamsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_77efba301b074cfb, []int{1}
}
func (m *MsgUpdateReputationParamsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgUpdateReputationParamsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgUpdateReputationParamsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgUpdateReputationParamsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgUpdateReputationParamsResponse.Merge(m, src)
}
func (m *MsgUpdateReputationParamsResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgUpdateReputationParamsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgUpdateReputationParamsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgUpdateReputationParamsResponse proto.InternalMessageInfo

type MsgApplyReputationPenalty struct {
	Authority	string	`protobuf:"bytes,1,opt,name=authority,proto3" json:"authority,omitempty"`
	SubjectType	string	`protobuf:"bytes,2,opt,name=subject_type,json=subjectType,proto3" json:"subject_type,omitempty"`
	Subject		string	`protobuf:"bytes,3,opt,name=subject,proto3" json:"subject,omitempty"`
	Component	string	`protobuf:"bytes,4,opt,name=component,proto3" json:"component,omitempty"`
	Amount		uint32	`protobuf:"varint,5,opt,name=amount,proto3" json:"amount,omitempty"`
	Reason		string	`protobuf:"bytes,6,opt,name=reason,proto3" json:"reason,omitempty"`
	Epoch		uint64	`protobuf:"varint,7,opt,name=epoch,proto3" json:"epoch,omitempty"`
}

func (m *MsgApplyReputationPenalty) Reset()		{ *m = MsgApplyReputationPenalty{} }
func (m *MsgApplyReputationPenalty) String() string	{ return proto.CompactTextString(m) }
func (*MsgApplyReputationPenalty) ProtoMessage()	{}
func (*MsgApplyReputationPenalty) Descriptor() ([]byte, []int) {
	return fileDescriptor_77efba301b074cfb, []int{2}
}
func (m *MsgApplyReputationPenalty) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgApplyReputationPenalty) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgApplyReputationPenalty.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgApplyReputationPenalty) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgApplyReputationPenalty.Merge(m, src)
}
func (m *MsgApplyReputationPenalty) XXX_Size() int {
	return m.Size()
}
func (m *MsgApplyReputationPenalty) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgApplyReputationPenalty.DiscardUnknown(m)
}

var xxx_messageInfo_MsgApplyReputationPenalty proto.InternalMessageInfo

func (m *MsgApplyReputationPenalty) GetAuthority() string {
	if m != nil {
		return m.Authority
	}
	return ""
}

func (m *MsgApplyReputationPenalty) GetSubjectType() string {
	if m != nil {
		return m.SubjectType
	}
	return ""
}

func (m *MsgApplyReputationPenalty) GetSubject() string {
	if m != nil {
		return m.Subject
	}
	return ""
}

func (m *MsgApplyReputationPenalty) GetComponent() string {
	if m != nil {
		return m.Component
	}
	return ""
}

func (m *MsgApplyReputationPenalty) GetAmount() uint32 {
	if m != nil {
		return m.Amount
	}
	return 0
}

func (m *MsgApplyReputationPenalty) GetReason() string {
	if m != nil {
		return m.Reason
	}
	return ""
}

func (m *MsgApplyReputationPenalty) GetEpoch() uint64 {
	if m != nil {
		return m.Epoch
	}
	return 0
}

type MsgApplyReputationPenaltyResponse struct {
	RecordJson string `protobuf:"bytes,1,opt,name=record_json,json=recordJson,proto3" json:"record_json,omitempty"`
}

func (m *MsgApplyReputationPenaltyResponse) Reset()		{ *m = MsgApplyReputationPenaltyResponse{} }
func (m *MsgApplyReputationPenaltyResponse) String() string	{ return proto.CompactTextString(m) }
func (*MsgApplyReputationPenaltyResponse) ProtoMessage()	{}
func (*MsgApplyReputationPenaltyResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_77efba301b074cfb, []int{3}
}
func (m *MsgApplyReputationPenaltyResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgApplyReputationPenaltyResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgApplyReputationPenaltyResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgApplyReputationPenaltyResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgApplyReputationPenaltyResponse.Merge(m, src)
}
func (m *MsgApplyReputationPenaltyResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgApplyReputationPenaltyResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgApplyReputationPenaltyResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgApplyReputationPenaltyResponse proto.InternalMessageInfo

func (m *MsgApplyReputationPenaltyResponse) GetRecordJson() string {
	if m != nil {
		return m.RecordJson
	}
	return ""
}

type MsgApplyReputationReward struct {
	Authority	string	`protobuf:"bytes,1,opt,name=authority,proto3" json:"authority,omitempty"`
	SubjectType	string	`protobuf:"bytes,2,opt,name=subject_type,json=subjectType,proto3" json:"subject_type,omitempty"`
	Subject		string	`protobuf:"bytes,3,opt,name=subject,proto3" json:"subject,omitempty"`
	Component	string	`protobuf:"bytes,4,opt,name=component,proto3" json:"component,omitempty"`
	Amount		uint32	`protobuf:"varint,5,opt,name=amount,proto3" json:"amount,omitempty"`
	Reason		string	`protobuf:"bytes,6,opt,name=reason,proto3" json:"reason,omitempty"`
	Epoch		uint64	`protobuf:"varint,7,opt,name=epoch,proto3" json:"epoch,omitempty"`
}

func (m *MsgApplyReputationReward) Reset()		{ *m = MsgApplyReputationReward{} }
func (m *MsgApplyReputationReward) String() string	{ return proto.CompactTextString(m) }
func (*MsgApplyReputationReward) ProtoMessage()		{}
func (*MsgApplyReputationReward) Descriptor() ([]byte, []int) {
	return fileDescriptor_77efba301b074cfb, []int{4}
}
func (m *MsgApplyReputationReward) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgApplyReputationReward) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgApplyReputationReward.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgApplyReputationReward) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgApplyReputationReward.Merge(m, src)
}
func (m *MsgApplyReputationReward) XXX_Size() int {
	return m.Size()
}
func (m *MsgApplyReputationReward) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgApplyReputationReward.DiscardUnknown(m)
}

var xxx_messageInfo_MsgApplyReputationReward proto.InternalMessageInfo

func (m *MsgApplyReputationReward) GetAuthority() string {
	if m != nil {
		return m.Authority
	}
	return ""
}

func (m *MsgApplyReputationReward) GetSubjectType() string {
	if m != nil {
		return m.SubjectType
	}
	return ""
}

func (m *MsgApplyReputationReward) GetSubject() string {
	if m != nil {
		return m.Subject
	}
	return ""
}

func (m *MsgApplyReputationReward) GetComponent() string {
	if m != nil {
		return m.Component
	}
	return ""
}

func (m *MsgApplyReputationReward) GetAmount() uint32 {
	if m != nil {
		return m.Amount
	}
	return 0
}

func (m *MsgApplyReputationReward) GetReason() string {
	if m != nil {
		return m.Reason
	}
	return ""
}

func (m *MsgApplyReputationReward) GetEpoch() uint64 {
	if m != nil {
		return m.Epoch
	}
	return 0
}

type MsgApplyReputationRewardResponse struct {
	RecordJson string `protobuf:"bytes,1,opt,name=record_json,json=recordJson,proto3" json:"record_json,omitempty"`
}

func (m *MsgApplyReputationRewardResponse) Reset()		{ *m = MsgApplyReputationRewardResponse{} }
func (m *MsgApplyReputationRewardResponse) String() string	{ return proto.CompactTextString(m) }
func (*MsgApplyReputationRewardResponse) ProtoMessage()		{}
func (*MsgApplyReputationRewardResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_77efba301b074cfb, []int{5}
}
func (m *MsgApplyReputationRewardResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgApplyReputationRewardResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgApplyReputationRewardResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgApplyReputationRewardResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgApplyReputationRewardResponse.Merge(m, src)
}
func (m *MsgApplyReputationRewardResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgApplyReputationRewardResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgApplyReputationRewardResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgApplyReputationRewardResponse proto.InternalMessageInfo

func (m *MsgApplyReputationRewardResponse) GetRecordJson() string {
	if m != nil {
		return m.RecordJson
	}
	return ""
}

type MsgRecomputeReputation struct {
	Authority	string	`protobuf:"bytes,1,opt,name=authority,proto3" json:"authority,omitempty"`
	SubjectType	string	`protobuf:"bytes,2,opt,name=subject_type,json=subjectType,proto3" json:"subject_type,omitempty"`
	Subject		string	`protobuf:"bytes,3,opt,name=subject,proto3" json:"subject,omitempty"`
	Epoch		uint64	`protobuf:"varint,4,opt,name=epoch,proto3" json:"epoch,omitempty"`
}

func (m *MsgRecomputeReputation) Reset()		{ *m = MsgRecomputeReputation{} }
func (m *MsgRecomputeReputation) String() string	{ return proto.CompactTextString(m) }
func (*MsgRecomputeReputation) ProtoMessage()		{}
func (*MsgRecomputeReputation) Descriptor() ([]byte, []int) {
	return fileDescriptor_77efba301b074cfb, []int{6}
}
func (m *MsgRecomputeReputation) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgRecomputeReputation) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgRecomputeReputation.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgRecomputeReputation) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgRecomputeReputation.Merge(m, src)
}
func (m *MsgRecomputeReputation) XXX_Size() int {
	return m.Size()
}
func (m *MsgRecomputeReputation) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgRecomputeReputation.DiscardUnknown(m)
}

var xxx_messageInfo_MsgRecomputeReputation proto.InternalMessageInfo

func (m *MsgRecomputeReputation) GetAuthority() string {
	if m != nil {
		return m.Authority
	}
	return ""
}

func (m *MsgRecomputeReputation) GetSubjectType() string {
	if m != nil {
		return m.SubjectType
	}
	return ""
}

func (m *MsgRecomputeReputation) GetSubject() string {
	if m != nil {
		return m.Subject
	}
	return ""
}

func (m *MsgRecomputeReputation) GetEpoch() uint64 {
	if m != nil {
		return m.Epoch
	}
	return 0
}

type MsgRecomputeReputationResponse struct {
	RecordJson string `protobuf:"bytes,1,opt,name=record_json,json=recordJson,proto3" json:"record_json,omitempty"`
}

func (m *MsgRecomputeReputationResponse) Reset()		{ *m = MsgRecomputeReputationResponse{} }
func (m *MsgRecomputeReputationResponse) String() string	{ return proto.CompactTextString(m) }
func (*MsgRecomputeReputationResponse) ProtoMessage()		{}
func (*MsgRecomputeReputationResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_77efba301b074cfb, []int{7}
}
func (m *MsgRecomputeReputationResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgRecomputeReputationResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgRecomputeReputationResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgRecomputeReputationResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgRecomputeReputationResponse.Merge(m, src)
}
func (m *MsgRecomputeReputationResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgRecomputeReputationResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgRecomputeReputationResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgRecomputeReputationResponse proto.InternalMessageInfo

func (m *MsgRecomputeReputationResponse) GetRecordJson() string {
	if m != nil {
		return m.RecordJson
	}
	return ""
}

func init() {
	proto.RegisterType((*MsgUpdateReputationParams)(nil), "l1.reputation.v1.MsgUpdateReputationParams")
	proto.RegisterType((*MsgUpdateReputationParamsResponse)(nil), "l1.reputation.v1.MsgUpdateReputationParamsResponse")
	proto.RegisterType((*MsgApplyReputationPenalty)(nil), "l1.reputation.v1.MsgApplyReputationPenalty")
	proto.RegisterType((*MsgApplyReputationPenaltyResponse)(nil), "l1.reputation.v1.MsgApplyReputationPenaltyResponse")
	proto.RegisterType((*MsgApplyReputationReward)(nil), "l1.reputation.v1.MsgApplyReputationReward")
	proto.RegisterType((*MsgApplyReputationRewardResponse)(nil), "l1.reputation.v1.MsgApplyReputationRewardResponse")
	proto.RegisterType((*MsgRecomputeReputation)(nil), "l1.reputation.v1.MsgRecomputeReputation")
	proto.RegisterType((*MsgRecomputeReputationResponse)(nil), "l1.reputation.v1.MsgRecomputeReputationResponse")
}

func init()	{ proto.RegisterFile("l1/reputation/v1/tx.proto", fileDescriptor_77efba301b074cfb) }

var fileDescriptor_77efba301b074cfb = []byte{

	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xdc, 0x55, 0x3d, 0x6f, 0xd3, 0x40,
	0x18, 0xce, 0x91, 0x8f, 0xaa, 0x6f, 0x01, 0x21, 0x03, 0xc1, 0xb5, 0x90, 0x49, 0xc3, 0x12, 0x05,
	0x61, 0xe3, 0x76, 0xeb, 0x56, 0x60, 0x42, 0x8a, 0x84, 0xac, 0xb2, 0xb0, 0x54, 0x8e, 0x73, 0x72,
	0x12, 0xd9, 0x77, 0xc7, 0xdd, 0x39, 0xad, 0x3b, 0x21, 0x7e, 0x01, 0x2b, 0xff, 0xa2, 0x3f, 0x83,
	0xb1, 0x23, 0x23, 0x4a, 0x86, 0x0a, 0x26, 0x7e, 0x02, 0xf2, 0x47, 0x9c, 0x88, 0xda, 0xc4, 0x0c,
	0x2c, 0x1d, 0xdf, 0xe7, 0x9e, 0x7b, 0xdf, 0xe7, 0x79, 0x74, 0xaf, 0x0d, 0xbb, 0xbe, 0x65, 0x72,
	0xcc, 0x42, 0xe9, 0xc8, 0x09, 0x25, 0xe6, 0xcc, 0x32, 0xe5, 0x99, 0xc1, 0x38, 0x95, 0x54, 0xb9,
	0xe7, 0x5b, 0xc6, 0xea, 0xc8, 0x98, 0x59, 0xda, 0x23, 0x97, 0x8a, 0x80, 0x0a, 0x33, 0x10, 0x5e,
	0xcc, 0x0c, 0x84, 0x97, 0x52, 0xbb, 0x53, 0xd8, 0x1d, 0x08, 0xef, 0x1d, 0x1b, 0x39, 0x12, 0xdb,
	0xf9, 0x95, 0xb7, 0x0e, 0x77, 0x02, 0xa1, 0x3c, 0x86, 0x6d, 0x27, 0x94, 0x63, 0xca, 0x27, 0x32,
	0x52, 0x51, 0x07, 0xf5, 0xb6, 0xed, 0x15, 0xa0, 0x3c, 0x81, 0x1d, 0x96, 0xf0, 0x4e, 0xa6, 0x82,
	0x12, 0xf5, 0x56, 0x72, 0x0e, 0x29, 0xf4, 0x46, 0x50, 0x72, 0x78, 0xf7, 0xd3, 0xd5, 0x45, 0x7f,
	0x75, 0xa1, 0xfb, 0x14, 0xf6, 0x4a, 0x67, 0xd9, 0x58, 0x30, 0x4a, 0x04, 0xee, 0xfe, 0x44, 0x89,
	0xa2, 0x23, 0xc6, 0xfc, 0x68, 0x8d, 0x84, 0x89, 0xe3, 0xcb, 0x68, 0x83, 0xa2, 0x3d, 0xb8, 0x2d,
	0xc2, 0xe1, 0x14, 0xbb, 0xf2, 0x44, 0x46, 0x0c, 0x67, 0x92, 0x76, 0x32, 0xec, 0x38, 0x62, 0x58,
	0x51, 0x61, 0x2b, 0x2b, 0xd5, 0x7a, 0x72, 0xba, 0x2c, 0xe3, 0xd6, 0x2e, 0x0d, 0x18, 0x25, 0x98,
	0x48, 0xb5, 0x91, 0xb6, 0xce, 0x01, 0xa5, 0x0d, 0x2d, 0x27, 0xa0, 0x21, 0x91, 0x6a, 0xb3, 0x83,
	0x7a, 0x77, 0xec, 0xac, 0x8a, 0x71, 0x8e, 0x9d, 0xd8, 0x7f, 0x2b, 0xb9, 0x92, 0x55, 0xca, 0x03,
	0x68, 0x62, 0x46, 0xdd, 0xb1, 0xba, 0xd5, 0x41, 0xbd, 0x86, 0x9d, 0x16, 0xd7, 0x12, 0x79, 0x9d,
	0x24, 0x52, 0xec, 0x75, 0x99, 0x48, 0x9c, 0x33, 0xc7, 0x2e, 0xe5, 0xa3, 0x34, 0xe7, 0xd4, 0x35,
	0xa4, 0x50, 0x9c, 0x73, 0xf7, 0x07, 0x02, 0xf5, 0x7a, 0x1b, 0x1b, 0x9f, 0x3a, 0x7c, 0x74, 0xd3,
	0x12, 0x7b, 0x05, 0x9d, 0x32, 0xab, 0xd5, 0x03, 0xfb, 0x82, 0xa0, 0x3d, 0x10, 0x9e, 0x8d, 0x63,
	0xb5, 0xe1, 0xfa, 0x63, 0xfc, 0x9f, 0x71, 0xe5, 0x06, 0x1b, 0x7f, 0x33, 0x78, 0x04, 0x7a, 0xb1,
	0xb4, 0xca, 0xf6, 0xf6, 0x7f, 0xd5, 0xa1, 0x3e, 0x10, 0x9e, 0x72, 0x0e, 0xed, 0x92, 0xc5, 0x7e,
	0x66, 0xfc, 0xf9, 0x85, 0x30, 0x4a, 0x37, 0x53, 0x3b, 0xf8, 0x07, 0x72, 0x2e, 0xf2, 0x1c, 0xda,
	0x25, 0x2b, 0x5c, 0x3c, 0xbb, 0x98, 0x5c, 0x32, 0x7b, 0xc3, 0xc2, 0x9c, 0xc2, 0xc3, 0xe2, 0x5d,
	0xe8, 0x57, 0xe9, 0x96, 0x72, 0xb5, 0xfd, 0xea, 0xdc, 0x7c, 0xf0, 0x07, 0xb8, 0x5f, 0xf4, 0xa6,
	0x7a, 0x85, 0xad, 0x0a, 0x98, 0xda, 0x8b, 0xaa, 0xcc, 0xe5, 0x48, 0xad, 0xf9, 0xf1, 0xea, 0xa2,
	0x8f, 0x5e, 0x1e, 0x7f, 0x9d, 0xeb, 0xe8, 0x72, 0xae, 0xa3, 0xef, 0x73, 0x1d, 0x7d, 0x5e, 0xe8,
	0xb5, 0xcb, 0x85, 0x5e, 0xfb, 0xb6, 0xd0, 0x6b, 0xef, 0x0f, 0xbd, 0x89, 0x1c, 0x87, 0x43, 0xc3,
	0xa5, 0x81, 0x29, 0xe8, 0x0c, 0x73, 0x3c, 0xf1, 0xc8, 0x73, 0xdf, 0x32, 0x7d, 0xcb, 0x3c, 0x5b,
	0xff, 0x81, 0xc4, 0x0f, 0x5a, 0xac, 0x01, 0x6c, 0x38, 0x6c, 0x25, 0xff, 0x88, 0x83, 0xdf, 0x01,
	0x00, 0x00, 0xff, 0xff, 0x48, 0xc1, 0xc3, 0x2f, 0x6b, 0x06, 0x00, 0x00,
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
	UpdateReputationParams(ctx context.Context, in *MsgUpdateReputationParams, opts ...grpc.CallOption) (*MsgUpdateReputationParamsResponse, error)
	ApplyReputationPenalty(ctx context.Context, in *MsgApplyReputationPenalty, opts ...grpc.CallOption) (*MsgApplyReputationPenaltyResponse, error)
	ApplyReputationReward(ctx context.Context, in *MsgApplyReputationReward, opts ...grpc.CallOption) (*MsgApplyReputationRewardResponse, error)
	RecomputeReputation(ctx context.Context, in *MsgRecomputeReputation, opts ...grpc.CallOption) (*MsgRecomputeReputationResponse, error)
}

type msgClient struct {
	cc grpc1.ClientConn
}

func NewMsgClient(cc grpc1.ClientConn) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) UpdateReputationParams(ctx context.Context, in *MsgUpdateReputationParams, opts ...grpc.CallOption) (*MsgUpdateReputationParamsResponse, error) {
	out := new(MsgUpdateReputationParamsResponse)
	err := c.cc.Invoke(ctx, "/l1.reputation.v1.Msg/UpdateReputationParams", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) ApplyReputationPenalty(ctx context.Context, in *MsgApplyReputationPenalty, opts ...grpc.CallOption) (*MsgApplyReputationPenaltyResponse, error) {
	out := new(MsgApplyReputationPenaltyResponse)
	err := c.cc.Invoke(ctx, "/l1.reputation.v1.Msg/ApplyReputationPenalty", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) ApplyReputationReward(ctx context.Context, in *MsgApplyReputationReward, opts ...grpc.CallOption) (*MsgApplyReputationRewardResponse, error) {
	out := new(MsgApplyReputationRewardResponse)
	err := c.cc.Invoke(ctx, "/l1.reputation.v1.Msg/ApplyReputationReward", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) RecomputeReputation(ctx context.Context, in *MsgRecomputeReputation, opts ...grpc.CallOption) (*MsgRecomputeReputationResponse, error) {
	out := new(MsgRecomputeReputationResponse)
	err := c.cc.Invoke(ctx, "/l1.reputation.v1.Msg/RecomputeReputation", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MsgServer is the server API for Msg service.
type MsgServer interface {
	UpdateReputationParams(context.Context, *MsgUpdateReputationParams) (*MsgUpdateReputationParamsResponse, error)
	ApplyReputationPenalty(context.Context, *MsgApplyReputationPenalty) (*MsgApplyReputationPenaltyResponse, error)
	ApplyReputationReward(context.Context, *MsgApplyReputationReward) (*MsgApplyReputationRewardResponse, error)
	RecomputeReputation(context.Context, *MsgRecomputeReputation) (*MsgRecomputeReputationResponse, error)
}

// UnimplementedMsgServer can be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (*UnimplementedMsgServer) UpdateReputationParams(ctx context.Context, req *MsgUpdateReputationParams) (*MsgUpdateReputationParamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateReputationParams not implemented")
}
func (*UnimplementedMsgServer) ApplyReputationPenalty(ctx context.Context, req *MsgApplyReputationPenalty) (*MsgApplyReputationPenaltyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ApplyReputationPenalty not implemented")
}
func (*UnimplementedMsgServer) ApplyReputationReward(ctx context.Context, req *MsgApplyReputationReward) (*MsgApplyReputationRewardResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ApplyReputationReward not implemented")
}
func (*UnimplementedMsgServer) RecomputeReputation(ctx context.Context, req *MsgRecomputeReputation) (*MsgRecomputeReputationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RecomputeReputation not implemented")
}

func RegisterMsgServer(s grpc1.Server, srv MsgServer) {
	s.RegisterService(&_Msg_serviceDesc, srv)
}

func _Msg_UpdateReputationParams_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUpdateReputationParams)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).UpdateReputationParams(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.reputation.v1.Msg/UpdateReputationParams",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).UpdateReputationParams(ctx, req.(*MsgUpdateReputationParams))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_ApplyReputationPenalty_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgApplyReputationPenalty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).ApplyReputationPenalty(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.reputation.v1.Msg/ApplyReputationPenalty",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).ApplyReputationPenalty(ctx, req.(*MsgApplyReputationPenalty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_ApplyReputationReward_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgApplyReputationReward)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).ApplyReputationReward(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.reputation.v1.Msg/ApplyReputationReward",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).ApplyReputationReward(ctx, req.(*MsgApplyReputationReward))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_RecomputeReputation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgRecomputeReputation)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).RecomputeReputation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.reputation.v1.Msg/RecomputeReputation",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).RecomputeReputation(ctx, req.(*MsgRecomputeReputation))
	}
	return interceptor(ctx, in, info, handler)
}

var Msg_serviceDesc = _Msg_serviceDesc
var _Msg_serviceDesc = grpc.ServiceDesc{
	ServiceName:	"l1.reputation.v1.Msg",
	HandlerType:	(*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName:	"UpdateReputationParams",
			Handler:	_Msg_UpdateReputationParams_Handler,
		},
		{
			MethodName:	"ApplyReputationPenalty",
			Handler:	_Msg_ApplyReputationPenalty_Handler,
		},
		{
			MethodName:	"ApplyReputationReward",
			Handler:	_Msg_ApplyReputationReward_Handler,
		},
		{
			MethodName:	"RecomputeReputation",
			Handler:	_Msg_RecomputeReputation_Handler,
		},
	},
	Streams:	[]grpc.StreamDesc{},
	Metadata:	"l1/reputation/v1/tx.proto",
}

func (m *MsgUpdateReputationParams) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgUpdateReputationParams) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgUpdateReputationParams) MarshalToSizedBuffer(dAtA []byte) (int, error) {
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

func (m *MsgUpdateReputationParamsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgUpdateReputationParamsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgUpdateReputationParamsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *MsgApplyReputationPenalty) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgApplyReputationPenalty) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgApplyReputationPenalty) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
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
	if m.Amount != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.Amount))
		i--
		dAtA[i] = 0x28
	}
	if len(m.Component) > 0 {
		i -= len(m.Component)
		copy(dAtA[i:], m.Component)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Component)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.Subject) > 0 {
		i -= len(m.Subject)
		copy(dAtA[i:], m.Subject)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Subject)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.SubjectType) > 0 {
		i -= len(m.SubjectType)
		copy(dAtA[i:], m.SubjectType)
		i = encodeVarintTx(dAtA, i, uint64(len(m.SubjectType)))
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

func (m *MsgApplyReputationPenaltyResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgApplyReputationPenaltyResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgApplyReputationPenaltyResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.RecordJson) > 0 {
		i -= len(m.RecordJson)
		copy(dAtA[i:], m.RecordJson)
		i = encodeVarintTx(dAtA, i, uint64(len(m.RecordJson)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MsgApplyReputationReward) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgApplyReputationReward) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgApplyReputationReward) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
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
	if m.Amount != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.Amount))
		i--
		dAtA[i] = 0x28
	}
	if len(m.Component) > 0 {
		i -= len(m.Component)
		copy(dAtA[i:], m.Component)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Component)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.Subject) > 0 {
		i -= len(m.Subject)
		copy(dAtA[i:], m.Subject)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Subject)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.SubjectType) > 0 {
		i -= len(m.SubjectType)
		copy(dAtA[i:], m.SubjectType)
		i = encodeVarintTx(dAtA, i, uint64(len(m.SubjectType)))
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

func (m *MsgApplyReputationRewardResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgApplyReputationRewardResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgApplyReputationRewardResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.RecordJson) > 0 {
		i -= len(m.RecordJson)
		copy(dAtA[i:], m.RecordJson)
		i = encodeVarintTx(dAtA, i, uint64(len(m.RecordJson)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MsgRecomputeReputation) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgRecomputeReputation) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgRecomputeReputation) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Epoch != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.Epoch))
		i--
		dAtA[i] = 0x20
	}
	if len(m.Subject) > 0 {
		i -= len(m.Subject)
		copy(dAtA[i:], m.Subject)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Subject)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.SubjectType) > 0 {
		i -= len(m.SubjectType)
		copy(dAtA[i:], m.SubjectType)
		i = encodeVarintTx(dAtA, i, uint64(len(m.SubjectType)))
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

func (m *MsgRecomputeReputationResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgRecomputeReputationResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgRecomputeReputationResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.RecordJson) > 0 {
		i -= len(m.RecordJson)
		copy(dAtA[i:], m.RecordJson)
		i = encodeVarintTx(dAtA, i, uint64(len(m.RecordJson)))
		i--
		dAtA[i] = 0xa
	}
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
func (m *MsgUpdateReputationParams) Size() (n int) {
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

func (m *MsgUpdateReputationParamsResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *MsgApplyReputationPenalty) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Authority)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.SubjectType)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.Subject)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.Component)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if m.Amount != 0 {
		n += 1 + sovTx(uint64(m.Amount))
	}
	l = len(m.Reason)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if m.Epoch != 0 {
		n += 1 + sovTx(uint64(m.Epoch))
	}
	return n
}

func (m *MsgApplyReputationPenaltyResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.RecordJson)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	return n
}

func (m *MsgApplyReputationReward) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Authority)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.SubjectType)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.Subject)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.Component)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if m.Amount != 0 {
		n += 1 + sovTx(uint64(m.Amount))
	}
	l = len(m.Reason)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if m.Epoch != 0 {
		n += 1 + sovTx(uint64(m.Epoch))
	}
	return n
}

func (m *MsgApplyReputationRewardResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.RecordJson)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	return n
}

func (m *MsgRecomputeReputation) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Authority)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.SubjectType)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.Subject)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if m.Epoch != 0 {
		n += 1 + sovTx(uint64(m.Epoch))
	}
	return n
}

func (m *MsgRecomputeReputationResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.RecordJson)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	return n
}

func sovTx(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozTx(x uint64) (n int) {
	return sovTx(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *MsgUpdateReputationParams) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgUpdateReputationParams: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgUpdateReputationParams: illegal tag %d (wire type %d)", fieldNum, wire)
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
func (m *MsgUpdateReputationParamsResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgUpdateReputationParamsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgUpdateReputationParamsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
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
func (m *MsgApplyReputationPenalty) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgApplyReputationPenalty: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgApplyReputationPenalty: illegal tag %d (wire type %d)", fieldNum, wire)
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
				return fmt.Errorf("proto: wrong wireType = %d for field SubjectType", wireType)
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
			m.SubjectType = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Subject", wireType)
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
			m.Subject = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Component", wireType)
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
			m.Component = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Amount", wireType)
			}
			m.Amount = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Amount |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
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
func (m *MsgApplyReputationPenaltyResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgApplyReputationPenaltyResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgApplyReputationPenaltyResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RecordJson", wireType)
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
			m.RecordJson = string(dAtA[iNdEx:postIndex])
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
func (m *MsgApplyReputationReward) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgApplyReputationReward: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgApplyReputationReward: illegal tag %d (wire type %d)", fieldNum, wire)
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
				return fmt.Errorf("proto: wrong wireType = %d for field SubjectType", wireType)
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
			m.SubjectType = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Subject", wireType)
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
			m.Subject = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Component", wireType)
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
			m.Component = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Amount", wireType)
			}
			m.Amount = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Amount |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
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
func (m *MsgApplyReputationRewardResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgApplyReputationRewardResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgApplyReputationRewardResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RecordJson", wireType)
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
			m.RecordJson = string(dAtA[iNdEx:postIndex])
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
func (m *MsgRecomputeReputation) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgRecomputeReputation: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgRecomputeReputation: illegal tag %d (wire type %d)", fieldNum, wire)
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
				return fmt.Errorf("proto: wrong wireType = %d for field SubjectType", wireType)
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
			m.SubjectType = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Subject", wireType)
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
			m.Subject = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
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
func (m *MsgRecomputeReputationResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgRecomputeReputationResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgRecomputeReputationResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RecordJson", wireType)
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
			m.RecordJson = string(dAtA[iNdEx:postIndex])
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
