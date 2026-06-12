package types

import (
	context "context"
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	_ "github.com/cosmos/cosmos-sdk/types/msgservice"
	_ "github.com/cosmos/cosmos-sdk/types/tx/amino"
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

type MsgSetBaseCommission struct {
	ValidatorAddress	string	`protobuf:"bytes,1,opt,name=validator_address,json=validatorAddress,proto3" json:"validator_address,omitempty"`
	BaseCommissionBps	uint32	`protobuf:"varint,2,opt,name=base_commission_bps,json=baseCommissionBps,proto3" json:"base_commission_bps,omitempty"`
	Height			uint64	`protobuf:"varint,3,opt,name=height,proto3" json:"height,omitempty"`
}

func (m *MsgSetBaseCommission) Reset()		{ *m = MsgSetBaseCommission{} }
func (m *MsgSetBaseCommission) String() string	{ return proto.CompactTextString(m) }
func (*MsgSetBaseCommission) ProtoMessage()	{}
func (*MsgSetBaseCommission) Descriptor() ([]byte, []int) {
	return fileDescriptor_74f0603bfc8a489d, []int{0}
}
func (m *MsgSetBaseCommission) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgSetBaseCommission) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgSetBaseCommission.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgSetBaseCommission) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgSetBaseCommission.Merge(m, src)
}
func (m *MsgSetBaseCommission) XXX_Size() int {
	return m.Size()
}
func (m *MsgSetBaseCommission) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgSetBaseCommission.DiscardUnknown(m)
}

var xxx_messageInfo_MsgSetBaseCommission proto.InternalMessageInfo

func (m *MsgSetBaseCommission) GetValidatorAddress() string {
	if m != nil {
		return m.ValidatorAddress
	}
	return ""
}

func (m *MsgSetBaseCommission) GetBaseCommissionBps() uint32 {
	if m != nil {
		return m.BaseCommissionBps
	}
	return 0
}

func (m *MsgSetBaseCommission) GetHeight() uint64 {
	if m != nil {
		return m.Height
	}
	return 0
}

type MsgSetBaseCommissionResponse struct {
	Commission ValidatorCommission `protobuf:"bytes,1,opt,name=commission,proto3" json:"commission"`
}

func (m *MsgSetBaseCommissionResponse) Reset()		{ *m = MsgSetBaseCommissionResponse{} }
func (m *MsgSetBaseCommissionResponse) String() string	{ return proto.CompactTextString(m) }
func (*MsgSetBaseCommissionResponse) ProtoMessage()	{}
func (*MsgSetBaseCommissionResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_74f0603bfc8a489d, []int{1}
}
func (m *MsgSetBaseCommissionResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgSetBaseCommissionResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgSetBaseCommissionResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgSetBaseCommissionResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgSetBaseCommissionResponse.Merge(m, src)
}
func (m *MsgSetBaseCommissionResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgSetBaseCommissionResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgSetBaseCommissionResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgSetBaseCommissionResponse proto.InternalMessageInfo

func (m *MsgSetBaseCommissionResponse) GetCommission() ValidatorCommission {
	if m != nil {
		return m.Commission
	}
	return ValidatorCommission{}
}

type MsgUpdateCommissionParams struct {
	Authority	string	`protobuf:"bytes,1,opt,name=authority,proto3" json:"authority,omitempty"`
	Params		Params	`protobuf:"bytes,2,opt,name=params,proto3" json:"params"`
}

func (m *MsgUpdateCommissionParams) Reset()		{ *m = MsgUpdateCommissionParams{} }
func (m *MsgUpdateCommissionParams) String() string	{ return proto.CompactTextString(m) }
func (*MsgUpdateCommissionParams) ProtoMessage()	{}
func (*MsgUpdateCommissionParams) Descriptor() ([]byte, []int) {
	return fileDescriptor_74f0603bfc8a489d, []int{2}
}
func (m *MsgUpdateCommissionParams) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgUpdateCommissionParams) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgUpdateCommissionParams.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgUpdateCommissionParams) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgUpdateCommissionParams.Merge(m, src)
}
func (m *MsgUpdateCommissionParams) XXX_Size() int {
	return m.Size()
}
func (m *MsgUpdateCommissionParams) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgUpdateCommissionParams.DiscardUnknown(m)
}

var xxx_messageInfo_MsgUpdateCommissionParams proto.InternalMessageInfo

func (m *MsgUpdateCommissionParams) GetAuthority() string {
	if m != nil {
		return m.Authority
	}
	return ""
}

func (m *MsgUpdateCommissionParams) GetParams() Params {
	if m != nil {
		return m.Params
	}
	return Params{}
}

type MsgUpdateCommissionParamsResponse struct {
}

func (m *MsgUpdateCommissionParamsResponse) Reset()		{ *m = MsgUpdateCommissionParamsResponse{} }
func (m *MsgUpdateCommissionParamsResponse) String() string	{ return proto.CompactTextString(m) }
func (*MsgUpdateCommissionParamsResponse) ProtoMessage()	{}
func (*MsgUpdateCommissionParamsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_74f0603bfc8a489d, []int{3}
}
func (m *MsgUpdateCommissionParamsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgUpdateCommissionParamsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgUpdateCommissionParamsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgUpdateCommissionParamsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgUpdateCommissionParamsResponse.Merge(m, src)
}
func (m *MsgUpdateCommissionParamsResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgUpdateCommissionParamsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgUpdateCommissionParamsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgUpdateCommissionParamsResponse proto.InternalMessageInfo

type MsgRecomputeEffectiveCommission struct {
	Authority		string	`protobuf:"bytes,1,opt,name=authority,proto3" json:"authority,omitempty"`
	ValidatorAddress	string	`protobuf:"bytes,2,opt,name=validator_address,json=validatorAddress,proto3" json:"validator_address,omitempty"`
	PerformanceScoreBps	uint32	`protobuf:"varint,3,opt,name=performance_score_bps,json=performanceScoreBps,proto3" json:"performance_score_bps,omitempty"`
	ReputationScoreBps	uint32	`protobuf:"varint,4,opt,name=reputation_score_bps,json=reputationScoreBps,proto3" json:"reputation_score_bps,omitempty"`
	Jailed			bool	`protobuf:"varint,5,opt,name=jailed,proto3" json:"jailed,omitempty"`
	Height			uint64	`protobuf:"varint,6,opt,name=height,proto3" json:"height,omitempty"`
}

func (m *MsgRecomputeEffectiveCommission) Reset()		{ *m = MsgRecomputeEffectiveCommission{} }
func (m *MsgRecomputeEffectiveCommission) String() string	{ return proto.CompactTextString(m) }
func (*MsgRecomputeEffectiveCommission) ProtoMessage()		{}
func (*MsgRecomputeEffectiveCommission) Descriptor() ([]byte, []int) {
	return fileDescriptor_74f0603bfc8a489d, []int{4}
}
func (m *MsgRecomputeEffectiveCommission) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgRecomputeEffectiveCommission) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgRecomputeEffectiveCommission.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgRecomputeEffectiveCommission) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgRecomputeEffectiveCommission.Merge(m, src)
}
func (m *MsgRecomputeEffectiveCommission) XXX_Size() int {
	return m.Size()
}
func (m *MsgRecomputeEffectiveCommission) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgRecomputeEffectiveCommission.DiscardUnknown(m)
}

var xxx_messageInfo_MsgRecomputeEffectiveCommission proto.InternalMessageInfo

func (m *MsgRecomputeEffectiveCommission) GetAuthority() string {
	if m != nil {
		return m.Authority
	}
	return ""
}

func (m *MsgRecomputeEffectiveCommission) GetValidatorAddress() string {
	if m != nil {
		return m.ValidatorAddress
	}
	return ""
}

func (m *MsgRecomputeEffectiveCommission) GetPerformanceScoreBps() uint32 {
	if m != nil {
		return m.PerformanceScoreBps
	}
	return 0
}

func (m *MsgRecomputeEffectiveCommission) GetReputationScoreBps() uint32 {
	if m != nil {
		return m.ReputationScoreBps
	}
	return 0
}

func (m *MsgRecomputeEffectiveCommission) GetJailed() bool {
	if m != nil {
		return m.Jailed
	}
	return false
}

func (m *MsgRecomputeEffectiveCommission) GetHeight() uint64 {
	if m != nil {
		return m.Height
	}
	return 0
}

type MsgRecomputeEffectiveCommissionResponse struct {
	Commission ValidatorCommission `protobuf:"bytes,1,opt,name=commission,proto3" json:"commission"`
}

func (m *MsgRecomputeEffectiveCommissionResponse) Reset() {
	*m = MsgRecomputeEffectiveCommissionResponse{}
}
func (m *MsgRecomputeEffectiveCommissionResponse) String() string	{ return proto.CompactTextString(m) }
func (*MsgRecomputeEffectiveCommissionResponse) ProtoMessage()		{}
func (*MsgRecomputeEffectiveCommissionResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_74f0603bfc8a489d, []int{5}
}
func (m *MsgRecomputeEffectiveCommissionResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgRecomputeEffectiveCommissionResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgRecomputeEffectiveCommissionResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgRecomputeEffectiveCommissionResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgRecomputeEffectiveCommissionResponse.Merge(m, src)
}
func (m *MsgRecomputeEffectiveCommissionResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgRecomputeEffectiveCommissionResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgRecomputeEffectiveCommissionResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgRecomputeEffectiveCommissionResponse proto.InternalMessageInfo

func (m *MsgRecomputeEffectiveCommissionResponse) GetCommission() ValidatorCommission {
	if m != nil {
		return m.Commission
	}
	return ValidatorCommission{}
}

func init() {
	proto.RegisterType((*MsgSetBaseCommission)(nil), "l1.dynamiccommission.v1.MsgSetBaseCommission")
	proto.RegisterType((*MsgSetBaseCommissionResponse)(nil), "l1.dynamiccommission.v1.MsgSetBaseCommissionResponse")
	proto.RegisterType((*MsgUpdateCommissionParams)(nil), "l1.dynamiccommission.v1.MsgUpdateCommissionParams")
	proto.RegisterType((*MsgUpdateCommissionParamsResponse)(nil), "l1.dynamiccommission.v1.MsgUpdateCommissionParamsResponse")
	proto.RegisterType((*MsgRecomputeEffectiveCommission)(nil), "l1.dynamiccommission.v1.MsgRecomputeEffectiveCommission")
	proto.RegisterType((*MsgRecomputeEffectiveCommissionResponse)(nil), "l1.dynamiccommission.v1.MsgRecomputeEffectiveCommissionResponse")
}

func init()	{ proto.RegisterFile("l1/dynamiccommission/v1/tx.proto", fileDescriptor_74f0603bfc8a489d) }

var fileDescriptor_74f0603bfc8a489d = []byte{

	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xb4, 0x55, 0x4f, 0x6b, 0x13, 0x41,
	0x1c, 0xcd, 0x34, 0x6d, 0xb0, 0x53, 0x04, 0xb3, 0x8d, 0x6d, 0x1a, 0x4a, 0x1a, 0x23, 0x62, 0x2c,
	0x66, 0xd7, 0x44, 0x2b, 0x52, 0x10, 0x6c, 0xb4, 0x78, 0x0a, 0xca, 0x16, 0x3d, 0x78, 0x09, 0x93,
	0xdd, 0xe9, 0x64, 0x64, 0x77, 0x67, 0x99, 0x99, 0x84, 0xe6, 0x20, 0x88, 0x27, 0xf1, 0xe4, 0xd9,
	0x2f, 0x61, 0x0f, 0x7e, 0x88, 0x1e, 0x8b, 0x20, 0x78, 0x12, 0x69, 0x0f, 0x3d, 0xf8, 0x21, 0x94,
	0xfd, 0x93, 0xec, 0x96, 0xec, 0xa6, 0xa4, 0xe0, 0x25, 0x64, 0xe6, 0xf7, 0x7b, 0x6f, 0xde, 0xef,
	0xcd, 0x9b, 0x04, 0x56, 0xac, 0x86, 0x66, 0x0e, 0x1d, 0x64, 0x53, 0xc3, 0x60, 0xb6, 0x4d, 0x85,
	0xa0, 0xcc, 0xd1, 0x06, 0x0d, 0x4d, 0x1e, 0xa8, 0x2e, 0x67, 0x92, 0x29, 0xab, 0x56, 0x43, 0x9d,
	0xe8, 0x50, 0x07, 0x8d, 0x52, 0x1e, 0xd9, 0xd4, 0x61, 0x9a, 0xff, 0x19, 0xf4, 0x96, 0x56, 0x0d,
	0x26, 0x6c, 0x26, 0x34, 0x5b, 0x10, 0x8f, 0xc3, 0x16, 0x24, 0x2c, 0xac, 0x05, 0x85, 0x8e, 0xbf,
	0xd2, 0x82, 0x45, 0x58, 0x2a, 0x10, 0x46, 0x58, 0xb0, 0xef, 0x7d, 0x0b, 0x77, 0x6f, 0xa5, 0xe9,
	0x22, 0xd8, 0xc1, 0x82, 0x86, 0xe0, 0xea, 0x1f, 0x00, 0x0b, 0x6d, 0x41, 0xf6, 0xb0, 0x6c, 0x21,
	0x81, 0x9f, 0x8e, 0x3b, 0x95, 0x5d, 0x98, 0x1f, 0x20, 0x8b, 0x9a, 0x48, 0x32, 0xde, 0x41, 0xa6,
	0xc9, 0xb1, 0x10, 0x45, 0x50, 0x01, 0xb5, 0xc5, 0x56, 0xf1, 0xfb, 0xb7, 0x7a, 0x21, 0x94, 0xb0,
	0x13, 0x54, 0xf6, 0x24, 0xa7, 0x0e, 0xd1, 0xaf, 0x8d, 0x21, 0xe1, 0xbe, 0xa2, 0xc2, 0xe5, 0x2e,
	0x12, 0xb8, 0x13, 0x69, 0xe8, 0x74, 0x5d, 0x51, 0x9c, 0xab, 0x80, 0xda, 0x55, 0x3d, 0xdf, 0x3d,
	0x77, 0x66, 0xcb, 0x15, 0xca, 0x0a, 0xcc, 0xf5, 0x30, 0x25, 0x3d, 0x59, 0xcc, 0x56, 0x40, 0x6d,
	0x5e, 0x0f, 0x57, 0xdb, 0xcf, 0x3e, 0x9c, 0x1d, 0x6e, 0x4e, 0x2a, 0xfa, 0x74, 0x76, 0xb8, 0x79,
	0x27, 0x71, 0xd2, 0xa4, 0xa1, 0xaa, 0x1c, 0xae, 0x27, 0xed, 0xeb, 0x58, 0xb8, 0xcc, 0x11, 0x58,
	0xd1, 0x21, 0x8c, 0x28, 0xfc, 0x69, 0x97, 0x9a, 0x77, 0xd5, 0x94, 0xfb, 0x53, 0x5f, 0x8f, 0xd4,
	0xc4, 0x46, 0x98, 0x3f, 0xfa, 0xb5, 0x91, 0xd1, 0x63, 0x2c, 0xd5, 0x1f, 0x00, 0xae, 0xb5, 0x05,
	0x79, 0xe5, 0x9a, 0x48, 0xc6, 0xce, 0x7c, 0x89, 0x38, 0xb2, 0x85, 0xf2, 0x10, 0x2e, 0xa2, 0xbe,
	0xec, 0x31, 0x4e, 0xe5, 0xf0, 0x42, 0x7b, 0xa3, 0x56, 0xe5, 0x31, 0xcc, 0xb9, 0x3e, 0x83, 0x6f,
	0xe5, 0x52, 0x73, 0x23, 0x55, 0x65, 0x70, 0x50, 0x28, 0x2c, 0x04, 0x6d, 0xef, 0x78, 0x76, 0x46,
	0x74, 0x9e, 0x8d, 0x6a, 0x9a, 0x8d, 0xc9, 0xca, 0xab, 0x37, 0xe1, 0x8d, 0xd4, 0xe2, 0xc8, 0xd0,
	0xea, 0xdf, 0x39, 0xb8, 0xd1, 0x16, 0x44, 0xc7, 0x06, 0xb3, 0xdd, 0xbe, 0xc4, 0xbb, 0xfb, 0xfb,
	0xd8, 0x90, 0x74, 0x10, 0x4f, 0xda, 0x65, 0x2d, 0x48, 0x4c, 0xe8, 0xdc, 0xcc, 0x09, 0x6d, 0xc2,
	0xeb, 0x2e, 0xe6, 0xfb, 0x8c, 0xdb, 0xc8, 0x31, 0x70, 0x47, 0x18, 0x8c, 0x63, 0x3f, 0xa3, 0x59,
	0x3f, 0xa3, 0xcb, 0xb1, 0xe2, 0x9e, 0x57, 0xf3, 0x52, 0x7a, 0x0f, 0x16, 0x38, 0x76, 0xfb, 0x12,
	0x49, 0x2f, 0xd0, 0x11, 0x64, 0xde, 0x87, 0x28, 0x51, 0x6d, 0x8c, 0x58, 0x81, 0xb9, 0xb7, 0x88,
	0x5a, 0xd8, 0x2c, 0x2e, 0x54, 0x40, 0xed, 0x8a, 0x1e, 0xae, 0x62, 0x79, 0xcf, 0x9d, 0xcb, 0xfb,
	0xf3, 0xc9, 0x0b, 0x7a, 0x90, 0x76, 0x41, 0xd3, 0xdc, 0xad, 0xbe, 0x83, 0xb7, 0x2f, 0x68, 0xf9,
	0x9f, 0xe9, 0x6f, 0x7e, 0xcd, 0xc2, 0x6c, 0x5b, 0x10, 0x65, 0x08, 0xf3, 0x93, 0xbf, 0x31, 0xf5,
	0x54, 0xf2, 0xa4, 0x57, 0x5a, 0xda, 0x9a, 0xa9, 0x7d, 0x3c, 0xd6, 0x47, 0x00, 0x57, 0x52, 0x5e,
	0x5f, 0x73, 0x1a, 0x63, 0x32, 0xa6, 0xb4, 0x3d, 0x3b, 0x66, 0x2c, 0xe5, 0x0b, 0x80, 0xeb, 0x53,
	0xdf, 0xc2, 0xa3, 0x69, 0xe4, 0xd3, 0x90, 0xa5, 0x27, 0x97, 0x45, 0x8e, 0xc4, 0x95, 0x16, 0xde,
	0x9f, 0x1d, 0x6e, 0x82, 0xd6, 0x8b, 0xa3, 0x93, 0x32, 0x38, 0x3e, 0x29, 0x83, 0xdf, 0x27, 0x65,
	0xf0, 0xf9, 0xb4, 0x9c, 0x39, 0x3e, 0x2d, 0x67, 0x7e, 0x9e, 0x96, 0x33, 0x6f, 0xb6, 0x08, 0x95,
	0xbd, 0x7e, 0x57, 0x35, 0x98, 0xad, 0x09, 0x36, 0xc0, 0x1c, 0x53, 0xe2, 0xd4, 0xad, 0x86, 0x66,
	0x35, 0xb4, 0x83, 0x51, 0x34, 0xeb, 0xb1, 0x6c, 0xca, 0xa1, 0x8b, 0x45, 0x37, 0xe7, 0xff, 0xd3,
	0xdc, 0xff, 0x17, 0x00, 0x00, 0xff, 0xff, 0x8c, 0xae, 0x80, 0x74, 0x2a, 0x07, 0x00, 0x00,
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
	SetBaseCommission(ctx context.Context, in *MsgSetBaseCommission, opts ...grpc.CallOption) (*MsgSetBaseCommissionResponse, error)
	UpdateCommissionParams(ctx context.Context, in *MsgUpdateCommissionParams, opts ...grpc.CallOption) (*MsgUpdateCommissionParamsResponse, error)
	RecomputeEffectiveCommission(ctx context.Context, in *MsgRecomputeEffectiveCommission, opts ...grpc.CallOption) (*MsgRecomputeEffectiveCommissionResponse, error)
}

type msgClient struct {
	cc grpc1.ClientConn
}

func NewMsgClient(cc grpc1.ClientConn) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) SetBaseCommission(ctx context.Context, in *MsgSetBaseCommission, opts ...grpc.CallOption) (*MsgSetBaseCommissionResponse, error) {
	out := new(MsgSetBaseCommissionResponse)
	err := c.cc.Invoke(ctx, "/l1.dynamiccommission.v1.Msg/SetBaseCommission", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) UpdateCommissionParams(ctx context.Context, in *MsgUpdateCommissionParams, opts ...grpc.CallOption) (*MsgUpdateCommissionParamsResponse, error) {
	out := new(MsgUpdateCommissionParamsResponse)
	err := c.cc.Invoke(ctx, "/l1.dynamiccommission.v1.Msg/UpdateCommissionParams", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) RecomputeEffectiveCommission(ctx context.Context, in *MsgRecomputeEffectiveCommission, opts ...grpc.CallOption) (*MsgRecomputeEffectiveCommissionResponse, error) {
	out := new(MsgRecomputeEffectiveCommissionResponse)
	err := c.cc.Invoke(ctx, "/l1.dynamiccommission.v1.Msg/RecomputeEffectiveCommission", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MsgServer is the server API for Msg service.
type MsgServer interface {
	SetBaseCommission(context.Context, *MsgSetBaseCommission) (*MsgSetBaseCommissionResponse, error)
	UpdateCommissionParams(context.Context, *MsgUpdateCommissionParams) (*MsgUpdateCommissionParamsResponse, error)
	RecomputeEffectiveCommission(context.Context, *MsgRecomputeEffectiveCommission) (*MsgRecomputeEffectiveCommissionResponse, error)
}

// UnimplementedMsgServer can be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (*UnimplementedMsgServer) SetBaseCommission(ctx context.Context, req *MsgSetBaseCommission) (*MsgSetBaseCommissionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetBaseCommission not implemented")
}
func (*UnimplementedMsgServer) UpdateCommissionParams(ctx context.Context, req *MsgUpdateCommissionParams) (*MsgUpdateCommissionParamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateCommissionParams not implemented")
}
func (*UnimplementedMsgServer) RecomputeEffectiveCommission(ctx context.Context, req *MsgRecomputeEffectiveCommission) (*MsgRecomputeEffectiveCommissionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RecomputeEffectiveCommission not implemented")
}

func RegisterMsgServer(s grpc1.Server, srv MsgServer) {
	s.RegisterService(&_Msg_serviceDesc, srv)
}

func _Msg_SetBaseCommission_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgSetBaseCommission)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).SetBaseCommission(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.dynamiccommission.v1.Msg/SetBaseCommission",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).SetBaseCommission(ctx, req.(*MsgSetBaseCommission))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_UpdateCommissionParams_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUpdateCommissionParams)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).UpdateCommissionParams(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.dynamiccommission.v1.Msg/UpdateCommissionParams",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).UpdateCommissionParams(ctx, req.(*MsgUpdateCommissionParams))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_RecomputeEffectiveCommission_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgRecomputeEffectiveCommission)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).RecomputeEffectiveCommission(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.dynamiccommission.v1.Msg/RecomputeEffectiveCommission",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).RecomputeEffectiveCommission(ctx, req.(*MsgRecomputeEffectiveCommission))
	}
	return interceptor(ctx, in, info, handler)
}

var Msg_serviceDesc = _Msg_serviceDesc
var _Msg_serviceDesc = grpc.ServiceDesc{
	ServiceName:	"l1.dynamiccommission.v1.Msg",
	HandlerType:	(*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName:	"SetBaseCommission",
			Handler:	_Msg_SetBaseCommission_Handler,
		},
		{
			MethodName:	"UpdateCommissionParams",
			Handler:	_Msg_UpdateCommissionParams_Handler,
		},
		{
			MethodName:	"RecomputeEffectiveCommission",
			Handler:	_Msg_RecomputeEffectiveCommission_Handler,
		},
	},
	Streams:	[]grpc.StreamDesc{},
	Metadata:	"l1/dynamiccommission/v1/tx.proto",
}

func (m *MsgSetBaseCommission) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgSetBaseCommission) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgSetBaseCommission) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Height != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.Height))
		i--
		dAtA[i] = 0x18
	}
	if m.BaseCommissionBps != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.BaseCommissionBps))
		i--
		dAtA[i] = 0x10
	}
	if len(m.ValidatorAddress) > 0 {
		i -= len(m.ValidatorAddress)
		copy(dAtA[i:], m.ValidatorAddress)
		i = encodeVarintTx(dAtA, i, uint64(len(m.ValidatorAddress)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MsgSetBaseCommissionResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgSetBaseCommissionResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgSetBaseCommissionResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.Commission.MarshalToSizedBuffer(dAtA[:i])
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

func (m *MsgUpdateCommissionParams) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgUpdateCommissionParams) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgUpdateCommissionParams) MarshalToSizedBuffer(dAtA []byte) (int, error) {
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

func (m *MsgUpdateCommissionParamsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgUpdateCommissionParamsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgUpdateCommissionParamsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *MsgRecomputeEffectiveCommission) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgRecomputeEffectiveCommission) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgRecomputeEffectiveCommission) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Height != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.Height))
		i--
		dAtA[i] = 0x30
	}
	if m.Jailed {
		i--
		if m.Jailed {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x28
	}
	if m.ReputationScoreBps != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.ReputationScoreBps))
		i--
		dAtA[i] = 0x20
	}
	if m.PerformanceScoreBps != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.PerformanceScoreBps))
		i--
		dAtA[i] = 0x18
	}
	if len(m.ValidatorAddress) > 0 {
		i -= len(m.ValidatorAddress)
		copy(dAtA[i:], m.ValidatorAddress)
		i = encodeVarintTx(dAtA, i, uint64(len(m.ValidatorAddress)))
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

func (m *MsgRecomputeEffectiveCommissionResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgRecomputeEffectiveCommissionResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgRecomputeEffectiveCommissionResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.Commission.MarshalToSizedBuffer(dAtA[:i])
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
func (m *MsgSetBaseCommission) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ValidatorAddress)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if m.BaseCommissionBps != 0 {
		n += 1 + sovTx(uint64(m.BaseCommissionBps))
	}
	if m.Height != 0 {
		n += 1 + sovTx(uint64(m.Height))
	}
	return n
}

func (m *MsgSetBaseCommissionResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Commission.Size()
	n += 1 + l + sovTx(uint64(l))
	return n
}

func (m *MsgUpdateCommissionParams) Size() (n int) {
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

func (m *MsgUpdateCommissionParamsResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *MsgRecomputeEffectiveCommission) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Authority)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.ValidatorAddress)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if m.PerformanceScoreBps != 0 {
		n += 1 + sovTx(uint64(m.PerformanceScoreBps))
	}
	if m.ReputationScoreBps != 0 {
		n += 1 + sovTx(uint64(m.ReputationScoreBps))
	}
	if m.Jailed {
		n += 2
	}
	if m.Height != 0 {
		n += 1 + sovTx(uint64(m.Height))
	}
	return n
}

func (m *MsgRecomputeEffectiveCommissionResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Commission.Size()
	n += 1 + l + sovTx(uint64(l))
	return n
}

func sovTx(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozTx(x uint64) (n int) {
	return sovTx(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *MsgSetBaseCommission) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgSetBaseCommission: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgSetBaseCommission: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ValidatorAddress", wireType)
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
			m.ValidatorAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field BaseCommissionBps", wireType)
			}
			m.BaseCommissionBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.BaseCommissionBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
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
func (m *MsgSetBaseCommissionResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgSetBaseCommissionResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgSetBaseCommissionResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Commission", wireType)
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
			if err := m.Commission.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
func (m *MsgUpdateCommissionParams) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgUpdateCommissionParams: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgUpdateCommissionParams: illegal tag %d (wire type %d)", fieldNum, wire)
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
func (m *MsgUpdateCommissionParamsResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgUpdateCommissionParamsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgUpdateCommissionParamsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
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
func (m *MsgRecomputeEffectiveCommission) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgRecomputeEffectiveCommission: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgRecomputeEffectiveCommission: illegal tag %d (wire type %d)", fieldNum, wire)
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
				return fmt.Errorf("proto: wrong wireType = %d for field ValidatorAddress", wireType)
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
			m.ValidatorAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field PerformanceScoreBps", wireType)
			}
			m.PerformanceScoreBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.PerformanceScoreBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ReputationScoreBps", wireType)
			}
			m.ReputationScoreBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ReputationScoreBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Jailed", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.Jailed = bool(v != 0)
		case 6:
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
func (m *MsgRecomputeEffectiveCommissionResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgRecomputeEffectiveCommissionResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgRecomputeEffectiveCommissionResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Commission", wireType)
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
			if err := m.Commission.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
