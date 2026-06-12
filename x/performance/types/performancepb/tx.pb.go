package performancepb

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

type MsgSubmitPerformanceReport struct {
	Authority	string	`protobuf:"bytes,1,opt,name=authority,proto3" json:"authority,omitempty"`
	ReportJson	string	`protobuf:"bytes,2,opt,name=report_json,json=reportJson,proto3" json:"report_json,omitempty"`
}

func (m *MsgSubmitPerformanceReport) Reset()		{ *m = MsgSubmitPerformanceReport{} }
func (m *MsgSubmitPerformanceReport) String() string	{ return proto.CompactTextString(m) }
func (*MsgSubmitPerformanceReport) ProtoMessage()	{}
func (*MsgSubmitPerformanceReport) Descriptor() ([]byte, []int) {
	return fileDescriptor_bd3810f3520acdd3, []int{0}
}
func (m *MsgSubmitPerformanceReport) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgSubmitPerformanceReport) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgSubmitPerformanceReport.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgSubmitPerformanceReport) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgSubmitPerformanceReport.Merge(m, src)
}
func (m *MsgSubmitPerformanceReport) XXX_Size() int {
	return m.Size()
}
func (m *MsgSubmitPerformanceReport) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgSubmitPerformanceReport.DiscardUnknown(m)
}

var xxx_messageInfo_MsgSubmitPerformanceReport proto.InternalMessageInfo

func (m *MsgSubmitPerformanceReport) GetAuthority() string {
	if m != nil {
		return m.Authority
	}
	return ""
}

func (m *MsgSubmitPerformanceReport) GetReportJson() string {
	if m != nil {
		return m.ReportJson
	}
	return ""
}

type MsgSubmitPerformanceReportResponse struct {
}

func (m *MsgSubmitPerformanceReportResponse) Reset()		{ *m = MsgSubmitPerformanceReportResponse{} }
func (m *MsgSubmitPerformanceReportResponse) String() string	{ return proto.CompactTextString(m) }
func (*MsgSubmitPerformanceReportResponse) ProtoMessage()	{}
func (*MsgSubmitPerformanceReportResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_bd3810f3520acdd3, []int{1}
}
func (m *MsgSubmitPerformanceReportResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgSubmitPerformanceReportResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgSubmitPerformanceReportResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgSubmitPerformanceReportResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgSubmitPerformanceReportResponse.Merge(m, src)
}
func (m *MsgSubmitPerformanceReportResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgSubmitPerformanceReportResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgSubmitPerformanceReportResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgSubmitPerformanceReportResponse proto.InternalMessageInfo

type MsgFinalizePerformanceEpoch struct {
	Authority	string	`protobuf:"bytes,1,opt,name=authority,proto3" json:"authority,omitempty"`
	Epoch		uint64	`protobuf:"varint,2,opt,name=epoch,proto3" json:"epoch,omitempty"`
}

func (m *MsgFinalizePerformanceEpoch) Reset()		{ *m = MsgFinalizePerformanceEpoch{} }
func (m *MsgFinalizePerformanceEpoch) String() string	{ return proto.CompactTextString(m) }
func (*MsgFinalizePerformanceEpoch) ProtoMessage()	{}
func (*MsgFinalizePerformanceEpoch) Descriptor() ([]byte, []int) {
	return fileDescriptor_bd3810f3520acdd3, []int{2}
}
func (m *MsgFinalizePerformanceEpoch) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgFinalizePerformanceEpoch) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgFinalizePerformanceEpoch.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgFinalizePerformanceEpoch) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgFinalizePerformanceEpoch.Merge(m, src)
}
func (m *MsgFinalizePerformanceEpoch) XXX_Size() int {
	return m.Size()
}
func (m *MsgFinalizePerformanceEpoch) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgFinalizePerformanceEpoch.DiscardUnknown(m)
}

var xxx_messageInfo_MsgFinalizePerformanceEpoch proto.InternalMessageInfo

func (m *MsgFinalizePerformanceEpoch) GetAuthority() string {
	if m != nil {
		return m.Authority
	}
	return ""
}

func (m *MsgFinalizePerformanceEpoch) GetEpoch() uint64 {
	if m != nil {
		return m.Epoch
	}
	return 0
}

type MsgFinalizePerformanceEpochResponse struct {
	EpochJson string `protobuf:"bytes,1,opt,name=epoch_json,json=epochJson,proto3" json:"epoch_json,omitempty"`
}

func (m *MsgFinalizePerformanceEpochResponse) Reset()		{ *m = MsgFinalizePerformanceEpochResponse{} }
func (m *MsgFinalizePerformanceEpochResponse) String() string	{ return proto.CompactTextString(m) }
func (*MsgFinalizePerformanceEpochResponse) ProtoMessage()	{}
func (*MsgFinalizePerformanceEpochResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_bd3810f3520acdd3, []int{3}
}
func (m *MsgFinalizePerformanceEpochResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgFinalizePerformanceEpochResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgFinalizePerformanceEpochResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgFinalizePerformanceEpochResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgFinalizePerformanceEpochResponse.Merge(m, src)
}
func (m *MsgFinalizePerformanceEpochResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgFinalizePerformanceEpochResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgFinalizePerformanceEpochResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgFinalizePerformanceEpochResponse proto.InternalMessageInfo

func (m *MsgFinalizePerformanceEpochResponse) GetEpochJson() string {
	if m != nil {
		return m.EpochJson
	}
	return ""
}

type MsgChallengePerformanceReport struct {
	Authority	string	`protobuf:"bytes,1,opt,name=authority,proto3" json:"authority,omitempty"`
	Epoch		uint64	`protobuf:"varint,2,opt,name=epoch,proto3" json:"epoch,omitempty"`
	ReportId	string	`protobuf:"bytes,3,opt,name=report_id,json=reportId,proto3" json:"report_id,omitempty"`
	Challenger	string	`protobuf:"bytes,4,opt,name=challenger,proto3" json:"challenger,omitempty"`
	Reason		string	`protobuf:"bytes,5,opt,name=reason,proto3" json:"reason,omitempty"`
	Accepted	bool	`protobuf:"varint,6,opt,name=accepted,proto3" json:"accepted,omitempty"`
}

func (m *MsgChallengePerformanceReport) Reset()		{ *m = MsgChallengePerformanceReport{} }
func (m *MsgChallengePerformanceReport) String() string	{ return proto.CompactTextString(m) }
func (*MsgChallengePerformanceReport) ProtoMessage()	{}
func (*MsgChallengePerformanceReport) Descriptor() ([]byte, []int) {
	return fileDescriptor_bd3810f3520acdd3, []int{4}
}
func (m *MsgChallengePerformanceReport) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgChallengePerformanceReport) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgChallengePerformanceReport.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgChallengePerformanceReport) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgChallengePerformanceReport.Merge(m, src)
}
func (m *MsgChallengePerformanceReport) XXX_Size() int {
	return m.Size()
}
func (m *MsgChallengePerformanceReport) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgChallengePerformanceReport.DiscardUnknown(m)
}

var xxx_messageInfo_MsgChallengePerformanceReport proto.InternalMessageInfo

func (m *MsgChallengePerformanceReport) GetAuthority() string {
	if m != nil {
		return m.Authority
	}
	return ""
}

func (m *MsgChallengePerformanceReport) GetEpoch() uint64 {
	if m != nil {
		return m.Epoch
	}
	return 0
}

func (m *MsgChallengePerformanceReport) GetReportId() string {
	if m != nil {
		return m.ReportId
	}
	return ""
}

func (m *MsgChallengePerformanceReport) GetChallenger() string {
	if m != nil {
		return m.Challenger
	}
	return ""
}

func (m *MsgChallengePerformanceReport) GetReason() string {
	if m != nil {
		return m.Reason
	}
	return ""
}

func (m *MsgChallengePerformanceReport) GetAccepted() bool {
	if m != nil {
		return m.Accepted
	}
	return false
}

type MsgChallengePerformanceReportResponse struct {
}

func (m *MsgChallengePerformanceReportResponse) Reset()		{ *m = MsgChallengePerformanceReportResponse{} }
func (m *MsgChallengePerformanceReportResponse) String() string	{ return proto.CompactTextString(m) }
func (*MsgChallengePerformanceReportResponse) ProtoMessage()	{}
func (*MsgChallengePerformanceReportResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_bd3810f3520acdd3, []int{5}
}
func (m *MsgChallengePerformanceReportResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgChallengePerformanceReportResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgChallengePerformanceReportResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgChallengePerformanceReportResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgChallengePerformanceReportResponse.Merge(m, src)
}
func (m *MsgChallengePerformanceReportResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgChallengePerformanceReportResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgChallengePerformanceReportResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgChallengePerformanceReportResponse proto.InternalMessageInfo

func init() {
	proto.RegisterType((*MsgSubmitPerformanceReport)(nil), "l1.performance.v1.MsgSubmitPerformanceReport")
	proto.RegisterType((*MsgSubmitPerformanceReportResponse)(nil), "l1.performance.v1.MsgSubmitPerformanceReportResponse")
	proto.RegisterType((*MsgFinalizePerformanceEpoch)(nil), "l1.performance.v1.MsgFinalizePerformanceEpoch")
	proto.RegisterType((*MsgFinalizePerformanceEpochResponse)(nil), "l1.performance.v1.MsgFinalizePerformanceEpochResponse")
	proto.RegisterType((*MsgChallengePerformanceReport)(nil), "l1.performance.v1.MsgChallengePerformanceReport")
	proto.RegisterType((*MsgChallengePerformanceReportResponse)(nil), "l1.performance.v1.MsgChallengePerformanceReportResponse")
}

func init()	{ proto.RegisterFile("l1/performance/v1/tx.proto", fileDescriptor_bd3810f3520acdd3) }

var fileDescriptor_bd3810f3520acdd3 = []byte{

	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x54, 0xcf, 0x8b, 0xd3, 0x40,
	0x14, 0xee, 0xd8, 0x6d, 0x69, 0x9f, 0x20, 0x38, 0x88, 0x1b, 0x66, 0xdd, 0xb8, 0x44, 0xc5, 0x65,
	0x61, 0x13, 0xa3, 0x28, 0x22, 0x9e, 0xfc, 0x05, 0x0a, 0x05, 0x89, 0xe0, 0xc1, 0x8b, 0xa4, 0xe9,
	0x38, 0x8d, 0x26, 0x99, 0x61, 0x66, 0x5a, 0x76, 0xbd, 0x28, 0x7b, 0xf2, 0xe8, 0x3f, 0xe0, 0xff,
	0xb0, 0x7f, 0x86, 0xc7, 0x3d, 0x78, 0xf0, 0x28, 0xed, 0x61, 0xff, 0x0d, 0xe9, 0xa4, 0x69, 0x2a,
	0x6d, 0x2a, 0x39, 0xbe, 0xef, 0x7d, 0x5f, 0xde, 0xf7, 0x5e, 0x3e, 0x06, 0x48, 0xe2, 0x7b, 0x82,
	0xca, 0x0f, 0x5c, 0xa6, 0x61, 0x16, 0x51, 0x6f, 0xec, 0x7b, 0xfa, 0xc8, 0x15, 0x92, 0x6b, 0x8e,
	0x2f, 0x27, 0xbe, 0xbb, 0xd4, 0x73, 0xc7, 0x3e, 0xd9, 0x8e, 0xb8, 0x4a, 0xb9, 0xf2, 0x52, 0xc5,
	0x66, 0xd4, 0x54, 0xb1, 0x9c, 0xeb, 0x7c, 0x02, 0xd2, 0x53, 0xec, 0xcd, 0xa8, 0x9f, 0xc6, 0xfa,
	0x75, 0xa9, 0x09, 0xa8, 0xe0, 0x52, 0xe3, 0x6b, 0xd0, 0x0d, 0x47, 0x7a, 0xc8, 0x65, 0xac, 0x8f,
	0x2d, 0xb4, 0x87, 0xf6, 0xbb, 0x41, 0x09, 0xe0, 0xeb, 0x70, 0x51, 0x1a, 0xde, 0xfb, 0x8f, 0x8a,
	0x67, 0xd6, 0x05, 0xd3, 0x87, 0x1c, 0x7a, 0xa5, 0x78, 0xf6, 0xe8, 0xd2, 0xc9, 0xf9, 0xe9, 0x41,
	0x29, 0x70, 0x6e, 0x82, 0x53, 0x3d, 0x2c, 0xa0, 0x4a, 0xf0, 0x4c, 0x51, 0x27, 0x84, 0x9d, 0x9e,
	0x62, 0x2f, 0xe2, 0x2c, 0x4c, 0xe2, 0xcf, 0x74, 0x89, 0xf7, 0x5c, 0xf0, 0x68, 0xf8, 0x1f, 0x4f,
	0x57, 0xa0, 0x45, 0x67, 0x34, 0xe3, 0x66, 0x2b, 0xc8, 0x8b, 0x15, 0x23, 0xcf, 0xe0, 0xc6, 0x86,
	0x11, 0x85, 0x13, 0xbc, 0x0b, 0x60, 0xf4, 0xf9, 0x7e, 0xf3, 0x59, 0x06, 0x99, 0xad, 0xe7, 0xfc,
	0x42, 0xb0, 0xdb, 0x53, 0xec, 0xe9, 0x30, 0x4c, 0x12, 0x9a, 0x31, 0x5a, 0xf7, 0x7e, 0x6b, 0xbd,
	0xe2, 0x1d, 0xe8, 0xce, 0xaf, 0x1a, 0x0f, 0xac, 0xa6, 0xd1, 0x74, 0x72, 0xe0, 0xe5, 0x00, 0xdb,
	0x00, 0x51, 0x31, 0x4e, 0x5a, 0x5b, 0xf9, 0xc5, 0x4b, 0x04, 0x5f, 0x85, 0xb6, 0xa4, 0xe1, 0xcc,
	0x6d, 0xcb, 0xf4, 0xe6, 0x15, 0x26, 0xd0, 0x09, 0xa3, 0x88, 0x0a, 0x4d, 0x07, 0x56, 0x7b, 0x0f,
	0xed, 0x77, 0x82, 0x45, 0xbd, 0x72, 0x9c, 0xdb, 0x70, 0x6b, 0xe3, 0x56, 0xc5, 0x79, 0xee, 0xfe,
	0x68, 0x42, 0xb3, 0xa7, 0x18, 0xfe, 0x02, 0xdb, 0x55, 0x01, 0x3a, 0x74, 0x57, 0xb2, 0xe8, 0x56,
	0x47, 0x80, 0xdc, 0xaf, 0x45, 0x5f, 0xfc, 0xa7, 0x13, 0x04, 0x56, 0x65, 0x5e, 0xdc, 0xf5, 0xdf,
	0xac, 0xe2, 0x93, 0x07, 0xf5, 0xf8, 0x0b, 0x13, 0xdf, 0x10, 0x90, 0x0d, 0x51, 0xb8, 0xb3, 0xfe,
	0xb3, 0xd5, 0x0a, 0xf2, 0xb0, 0xae, 0xa2, 0xb0, 0x42, 0x5a, 0x5f, 0xcf, 0x4f, 0x0f, 0xd0, 0x93,
	0xb7, 0x3f, 0x27, 0x36, 0x3a, 0x9b, 0xd8, 0xe8, 0xcf, 0xc4, 0x46, 0xdf, 0xa7, 0x76, 0xe3, 0x6c,
	0x6a, 0x37, 0x7e, 0x4f, 0xed, 0xc6, 0xbb, 0xc7, 0x2c, 0xd6, 0xc3, 0x51, 0xdf, 0x8d, 0x78, 0xea,
	0x29, 0x3e, 0xa6, 0x92, 0xc6, 0x2c, 0x3b, 0x4c, 0x7c, 0x2f, 0xf1, 0xbd, 0xa3, 0x7f, 0xde, 0x15,
	0x7d, 0x2c, 0xa8, 0x5a, 0x46, 0x44, 0xbf, 0xdf, 0x36, 0x4f, 0xc7, 0xbd, 0xbf, 0x01, 0x00, 0x00,
	0xff, 0xff, 0xc8, 0x89, 0xc9, 0x5d, 0x84, 0x04, 0x00, 0x00,
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
	SubmitPerformanceReport(ctx context.Context, in *MsgSubmitPerformanceReport, opts ...grpc.CallOption) (*MsgSubmitPerformanceReportResponse, error)
	FinalizePerformanceEpoch(ctx context.Context, in *MsgFinalizePerformanceEpoch, opts ...grpc.CallOption) (*MsgFinalizePerformanceEpochResponse, error)
	ChallengePerformanceReport(ctx context.Context, in *MsgChallengePerformanceReport, opts ...grpc.CallOption) (*MsgChallengePerformanceReportResponse, error)
}

type msgClient struct {
	cc grpc1.ClientConn
}

func NewMsgClient(cc grpc1.ClientConn) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) SubmitPerformanceReport(ctx context.Context, in *MsgSubmitPerformanceReport, opts ...grpc.CallOption) (*MsgSubmitPerformanceReportResponse, error) {
	out := new(MsgSubmitPerformanceReportResponse)
	err := c.cc.Invoke(ctx, "/l1.performance.v1.Msg/SubmitPerformanceReport", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) FinalizePerformanceEpoch(ctx context.Context, in *MsgFinalizePerformanceEpoch, opts ...grpc.CallOption) (*MsgFinalizePerformanceEpochResponse, error) {
	out := new(MsgFinalizePerformanceEpochResponse)
	err := c.cc.Invoke(ctx, "/l1.performance.v1.Msg/FinalizePerformanceEpoch", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) ChallengePerformanceReport(ctx context.Context, in *MsgChallengePerformanceReport, opts ...grpc.CallOption) (*MsgChallengePerformanceReportResponse, error) {
	out := new(MsgChallengePerformanceReportResponse)
	err := c.cc.Invoke(ctx, "/l1.performance.v1.Msg/ChallengePerformanceReport", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MsgServer is the server API for Msg service.
type MsgServer interface {
	SubmitPerformanceReport(context.Context, *MsgSubmitPerformanceReport) (*MsgSubmitPerformanceReportResponse, error)
	FinalizePerformanceEpoch(context.Context, *MsgFinalizePerformanceEpoch) (*MsgFinalizePerformanceEpochResponse, error)
	ChallengePerformanceReport(context.Context, *MsgChallengePerformanceReport) (*MsgChallengePerformanceReportResponse, error)
}

// UnimplementedMsgServer can be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (*UnimplementedMsgServer) SubmitPerformanceReport(ctx context.Context, req *MsgSubmitPerformanceReport) (*MsgSubmitPerformanceReportResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitPerformanceReport not implemented")
}
func (*UnimplementedMsgServer) FinalizePerformanceEpoch(ctx context.Context, req *MsgFinalizePerformanceEpoch) (*MsgFinalizePerformanceEpochResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FinalizePerformanceEpoch not implemented")
}
func (*UnimplementedMsgServer) ChallengePerformanceReport(ctx context.Context, req *MsgChallengePerformanceReport) (*MsgChallengePerformanceReportResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ChallengePerformanceReport not implemented")
}

func RegisterMsgServer(s grpc1.Server, srv MsgServer) {
	s.RegisterService(&_Msg_serviceDesc, srv)
}

func _Msg_SubmitPerformanceReport_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgSubmitPerformanceReport)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).SubmitPerformanceReport(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.performance.v1.Msg/SubmitPerformanceReport",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).SubmitPerformanceReport(ctx, req.(*MsgSubmitPerformanceReport))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_FinalizePerformanceEpoch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgFinalizePerformanceEpoch)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).FinalizePerformanceEpoch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.performance.v1.Msg/FinalizePerformanceEpoch",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).FinalizePerformanceEpoch(ctx, req.(*MsgFinalizePerformanceEpoch))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_ChallengePerformanceReport_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgChallengePerformanceReport)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).ChallengePerformanceReport(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.performance.v1.Msg/ChallengePerformanceReport",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).ChallengePerformanceReport(ctx, req.(*MsgChallengePerformanceReport))
	}
	return interceptor(ctx, in, info, handler)
}

var Msg_serviceDesc = _Msg_serviceDesc
var _Msg_serviceDesc = grpc.ServiceDesc{
	ServiceName:	"l1.performance.v1.Msg",
	HandlerType:	(*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName:	"SubmitPerformanceReport",
			Handler:	_Msg_SubmitPerformanceReport_Handler,
		},
		{
			MethodName:	"FinalizePerformanceEpoch",
			Handler:	_Msg_FinalizePerformanceEpoch_Handler,
		},
		{
			MethodName:	"ChallengePerformanceReport",
			Handler:	_Msg_ChallengePerformanceReport_Handler,
		},
	},
	Streams:	[]grpc.StreamDesc{},
	Metadata:	"l1/performance/v1/tx.proto",
}

func (m *MsgSubmitPerformanceReport) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgSubmitPerformanceReport) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgSubmitPerformanceReport) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.ReportJson) > 0 {
		i -= len(m.ReportJson)
		copy(dAtA[i:], m.ReportJson)
		i = encodeVarintTx(dAtA, i, uint64(len(m.ReportJson)))
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

func (m *MsgSubmitPerformanceReportResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgSubmitPerformanceReportResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgSubmitPerformanceReportResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *MsgFinalizePerformanceEpoch) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgFinalizePerformanceEpoch) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgFinalizePerformanceEpoch) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Epoch != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.Epoch))
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

func (m *MsgFinalizePerformanceEpochResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgFinalizePerformanceEpochResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgFinalizePerformanceEpochResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.EpochJson) > 0 {
		i -= len(m.EpochJson)
		copy(dAtA[i:], m.EpochJson)
		i = encodeVarintTx(dAtA, i, uint64(len(m.EpochJson)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MsgChallengePerformanceReport) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgChallengePerformanceReport) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgChallengePerformanceReport) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Accepted {
		i--
		if m.Accepted {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x30
	}
	if len(m.Reason) > 0 {
		i -= len(m.Reason)
		copy(dAtA[i:], m.Reason)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Reason)))
		i--
		dAtA[i] = 0x2a
	}
	if len(m.Challenger) > 0 {
		i -= len(m.Challenger)
		copy(dAtA[i:], m.Challenger)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Challenger)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.ReportId) > 0 {
		i -= len(m.ReportId)
		copy(dAtA[i:], m.ReportId)
		i = encodeVarintTx(dAtA, i, uint64(len(m.ReportId)))
		i--
		dAtA[i] = 0x1a
	}
	if m.Epoch != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.Epoch))
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

func (m *MsgChallengePerformanceReportResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgChallengePerformanceReportResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgChallengePerformanceReportResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
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
func (m *MsgSubmitPerformanceReport) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Authority)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.ReportJson)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	return n
}

func (m *MsgSubmitPerformanceReportResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *MsgFinalizePerformanceEpoch) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Authority)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if m.Epoch != 0 {
		n += 1 + sovTx(uint64(m.Epoch))
	}
	return n
}

func (m *MsgFinalizePerformanceEpochResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.EpochJson)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	return n
}

func (m *MsgChallengePerformanceReport) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Authority)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if m.Epoch != 0 {
		n += 1 + sovTx(uint64(m.Epoch))
	}
	l = len(m.ReportId)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.Challenger)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.Reason)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if m.Accepted {
		n += 2
	}
	return n
}

func (m *MsgChallengePerformanceReportResponse) Size() (n int) {
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
func (m *MsgSubmitPerformanceReport) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgSubmitPerformanceReport: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgSubmitPerformanceReport: illegal tag %d (wire type %d)", fieldNum, wire)
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
				return fmt.Errorf("proto: wrong wireType = %d for field ReportJson", wireType)
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
			m.ReportJson = string(dAtA[iNdEx:postIndex])
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
func (m *MsgSubmitPerformanceReportResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgSubmitPerformanceReportResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgSubmitPerformanceReportResponse: illegal tag %d (wire type %d)", fieldNum, wire)
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
func (m *MsgFinalizePerformanceEpoch) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgFinalizePerformanceEpoch: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgFinalizePerformanceEpoch: illegal tag %d (wire type %d)", fieldNum, wire)
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
func (m *MsgFinalizePerformanceEpochResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgFinalizePerformanceEpochResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgFinalizePerformanceEpochResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field EpochJson", wireType)
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
			m.EpochJson = string(dAtA[iNdEx:postIndex])
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
func (m *MsgChallengePerformanceReport) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgChallengePerformanceReport: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgChallengePerformanceReport: illegal tag %d (wire type %d)", fieldNum, wire)
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
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ReportId", wireType)
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
			m.ReportId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Challenger", wireType)
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
			m.Challenger = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
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
		case 6:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Accepted", wireType)
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
			m.Accepted = bool(v != 0)
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
func (m *MsgChallengePerformanceReportResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgChallengePerformanceReportResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgChallengePerformanceReportResponse: illegal tag %d (wire type %d)", fieldNum, wire)
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
