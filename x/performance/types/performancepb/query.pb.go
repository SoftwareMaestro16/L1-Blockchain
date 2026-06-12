package performancepb

import (
	context "context"
	fmt "fmt"
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

type QueryValidatorPerformanceRequest struct {
	Epoch			uint64	`protobuf:"varint,1,opt,name=epoch,proto3" json:"epoch,omitempty"`
	ValidatorAddress	string	`protobuf:"bytes,2,opt,name=validator_address,json=validatorAddress,proto3" json:"validator_address,omitempty"`
}

func (m *QueryValidatorPerformanceRequest) Reset()		{ *m = QueryValidatorPerformanceRequest{} }
func (m *QueryValidatorPerformanceRequest) String() string	{ return proto.CompactTextString(m) }
func (*QueryValidatorPerformanceRequest) ProtoMessage()		{}
func (*QueryValidatorPerformanceRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_fb6a565486ae67f1, []int{0}
}
func (m *QueryValidatorPerformanceRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryValidatorPerformanceRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryValidatorPerformanceRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryValidatorPerformanceRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryValidatorPerformanceRequest.Merge(m, src)
}
func (m *QueryValidatorPerformanceRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryValidatorPerformanceRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryValidatorPerformanceRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryValidatorPerformanceRequest proto.InternalMessageInfo

func (m *QueryValidatorPerformanceRequest) GetEpoch() uint64 {
	if m != nil {
		return m.Epoch
	}
	return 0
}

func (m *QueryValidatorPerformanceRequest) GetValidatorAddress() string {
	if m != nil {
		return m.ValidatorAddress
	}
	return ""
}

type QueryValidatorPerformanceResponse struct {
	AggregateJson string `protobuf:"bytes,1,opt,name=aggregate_json,json=aggregateJson,proto3" json:"aggregate_json,omitempty"`
}

func (m *QueryValidatorPerformanceResponse) Reset()		{ *m = QueryValidatorPerformanceResponse{} }
func (m *QueryValidatorPerformanceResponse) String() string	{ return proto.CompactTextString(m) }
func (*QueryValidatorPerformanceResponse) ProtoMessage()	{}
func (*QueryValidatorPerformanceResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_fb6a565486ae67f1, []int{1}
}
func (m *QueryValidatorPerformanceResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryValidatorPerformanceResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryValidatorPerformanceResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryValidatorPerformanceResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryValidatorPerformanceResponse.Merge(m, src)
}
func (m *QueryValidatorPerformanceResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryValidatorPerformanceResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryValidatorPerformanceResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryValidatorPerformanceResponse proto.InternalMessageInfo

func (m *QueryValidatorPerformanceResponse) GetAggregateJson() string {
	if m != nil {
		return m.AggregateJson
	}
	return ""
}

type QueryPerformanceEpochRequest struct {
	Epoch uint64 `protobuf:"varint,1,opt,name=epoch,proto3" json:"epoch,omitempty"`
}

func (m *QueryPerformanceEpochRequest) Reset()		{ *m = QueryPerformanceEpochRequest{} }
func (m *QueryPerformanceEpochRequest) String() string	{ return proto.CompactTextString(m) }
func (*QueryPerformanceEpochRequest) ProtoMessage()	{}
func (*QueryPerformanceEpochRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_fb6a565486ae67f1, []int{2}
}
func (m *QueryPerformanceEpochRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryPerformanceEpochRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryPerformanceEpochRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryPerformanceEpochRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryPerformanceEpochRequest.Merge(m, src)
}
func (m *QueryPerformanceEpochRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryPerformanceEpochRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryPerformanceEpochRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryPerformanceEpochRequest proto.InternalMessageInfo

func (m *QueryPerformanceEpochRequest) GetEpoch() uint64 {
	if m != nil {
		return m.Epoch
	}
	return 0
}

type QueryPerformanceEpochResponse struct {
	EpochJson string `protobuf:"bytes,1,opt,name=epoch_json,json=epochJson,proto3" json:"epoch_json,omitempty"`
}

func (m *QueryPerformanceEpochResponse) Reset()		{ *m = QueryPerformanceEpochResponse{} }
func (m *QueryPerformanceEpochResponse) String() string	{ return proto.CompactTextString(m) }
func (*QueryPerformanceEpochResponse) ProtoMessage()	{}
func (*QueryPerformanceEpochResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_fb6a565486ae67f1, []int{3}
}
func (m *QueryPerformanceEpochResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryPerformanceEpochResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryPerformanceEpochResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryPerformanceEpochResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryPerformanceEpochResponse.Merge(m, src)
}
func (m *QueryPerformanceEpochResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryPerformanceEpochResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryPerformanceEpochResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryPerformanceEpochResponse proto.InternalMessageInfo

func (m *QueryPerformanceEpochResponse) GetEpochJson() string {
	if m != nil {
		return m.EpochJson
	}
	return ""
}

type QueryPerformanceReportsRequest struct {
	Epoch			uint64	`protobuf:"varint,1,opt,name=epoch,proto3" json:"epoch,omitempty"`
	ValidatorAddress	string	`protobuf:"bytes,2,opt,name=validator_address,json=validatorAddress,proto3" json:"validator_address,omitempty"`
}

func (m *QueryPerformanceReportsRequest) Reset()		{ *m = QueryPerformanceReportsRequest{} }
func (m *QueryPerformanceReportsRequest) String() string	{ return proto.CompactTextString(m) }
func (*QueryPerformanceReportsRequest) ProtoMessage()		{}
func (*QueryPerformanceReportsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_fb6a565486ae67f1, []int{4}
}
func (m *QueryPerformanceReportsRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryPerformanceReportsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryPerformanceReportsRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryPerformanceReportsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryPerformanceReportsRequest.Merge(m, src)
}
func (m *QueryPerformanceReportsRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryPerformanceReportsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryPerformanceReportsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryPerformanceReportsRequest proto.InternalMessageInfo

func (m *QueryPerformanceReportsRequest) GetEpoch() uint64 {
	if m != nil {
		return m.Epoch
	}
	return 0
}

func (m *QueryPerformanceReportsRequest) GetValidatorAddress() string {
	if m != nil {
		return m.ValidatorAddress
	}
	return ""
}

type QueryPerformanceReportsResponse struct {
	ReportsJson string `protobuf:"bytes,1,opt,name=reports_json,json=reportsJson,proto3" json:"reports_json,omitempty"`
}

func (m *QueryPerformanceReportsResponse) Reset()		{ *m = QueryPerformanceReportsResponse{} }
func (m *QueryPerformanceReportsResponse) String() string	{ return proto.CompactTextString(m) }
func (*QueryPerformanceReportsResponse) ProtoMessage()		{}
func (*QueryPerformanceReportsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_fb6a565486ae67f1, []int{5}
}
func (m *QueryPerformanceReportsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryPerformanceReportsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryPerformanceReportsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryPerformanceReportsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryPerformanceReportsResponse.Merge(m, src)
}
func (m *QueryPerformanceReportsResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryPerformanceReportsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryPerformanceReportsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryPerformanceReportsResponse proto.InternalMessageInfo

func (m *QueryPerformanceReportsResponse) GetReportsJson() string {
	if m != nil {
		return m.ReportsJson
	}
	return ""
}

type QueryPerformanceParamsRequest struct {
}

func (m *QueryPerformanceParamsRequest) Reset()		{ *m = QueryPerformanceParamsRequest{} }
func (m *QueryPerformanceParamsRequest) String() string	{ return proto.CompactTextString(m) }
func (*QueryPerformanceParamsRequest) ProtoMessage()	{}
func (*QueryPerformanceParamsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_fb6a565486ae67f1, []int{6}
}
func (m *QueryPerformanceParamsRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryPerformanceParamsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryPerformanceParamsRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryPerformanceParamsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryPerformanceParamsRequest.Merge(m, src)
}
func (m *QueryPerformanceParamsRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryPerformanceParamsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryPerformanceParamsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryPerformanceParamsRequest proto.InternalMessageInfo

type QueryPerformanceParamsResponse struct {
	ParamsJson string `protobuf:"bytes,1,opt,name=params_json,json=paramsJson,proto3" json:"params_json,omitempty"`
}

func (m *QueryPerformanceParamsResponse) Reset()		{ *m = QueryPerformanceParamsResponse{} }
func (m *QueryPerformanceParamsResponse) String() string	{ return proto.CompactTextString(m) }
func (*QueryPerformanceParamsResponse) ProtoMessage()		{}
func (*QueryPerformanceParamsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_fb6a565486ae67f1, []int{7}
}
func (m *QueryPerformanceParamsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryPerformanceParamsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryPerformanceParamsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryPerformanceParamsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryPerformanceParamsResponse.Merge(m, src)
}
func (m *QueryPerformanceParamsResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryPerformanceParamsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryPerformanceParamsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryPerformanceParamsResponse proto.InternalMessageInfo

func (m *QueryPerformanceParamsResponse) GetParamsJson() string {
	if m != nil {
		return m.ParamsJson
	}
	return ""
}

func init() {
	proto.RegisterType((*QueryValidatorPerformanceRequest)(nil), "l1.performance.v1.QueryValidatorPerformanceRequest")
	proto.RegisterType((*QueryValidatorPerformanceResponse)(nil), "l1.performance.v1.QueryValidatorPerformanceResponse")
	proto.RegisterType((*QueryPerformanceEpochRequest)(nil), "l1.performance.v1.QueryPerformanceEpochRequest")
	proto.RegisterType((*QueryPerformanceEpochResponse)(nil), "l1.performance.v1.QueryPerformanceEpochResponse")
	proto.RegisterType((*QueryPerformanceReportsRequest)(nil), "l1.performance.v1.QueryPerformanceReportsRequest")
	proto.RegisterType((*QueryPerformanceReportsResponse)(nil), "l1.performance.v1.QueryPerformanceReportsResponse")
	proto.RegisterType((*QueryPerformanceParamsRequest)(nil), "l1.performance.v1.QueryPerformanceParamsRequest")
	proto.RegisterType((*QueryPerformanceParamsResponse)(nil), "l1.performance.v1.QueryPerformanceParamsResponse")
}

func init()	{ proto.RegisterFile("l1/performance/v1/query.proto", fileDescriptor_fb6a565486ae67f1) }

var fileDescriptor_fb6a565486ae67f1 = []byte{

	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x94, 0x31, 0x6f, 0xd3, 0x40,
	0x1c, 0xc5, 0x73, 0x15, 0x41, 0xca, 0xbf, 0x80, 0x9a, 0x53, 0x87, 0x62, 0x1a, 0x27, 0x31, 0x20,
	0x15, 0x10, 0x3e, 0xdc, 0x76, 0x41, 0x20, 0xa4, 0x20, 0x58, 0x3a, 0x15, 0x0f, 0x1d, 0x58, 0xaa,
	0x4b, 0x72, 0xb8, 0x46, 0x8e, 0xef, 0x7a, 0x77, 0x89, 0xa8, 0xaa, 0x2e, 0x7c, 0x02, 0x24, 0x46,
	0x16, 0x36, 0x3e, 0x03, 0x3b, 0x03, 0x63, 0x25, 0x16, 0x46, 0x94, 0xf0, 0x41, 0x50, 0xce, 0x8e,
	0x71, 0xea, 0xa4, 0x24, 0x12, 0x63, 0xde, 0xdd, 0xfb, 0xbf, 0xdf, 0xdd, 0xbd, 0x18, 0x6a, 0x91,
	0x47, 0x04, 0x93, 0x6f, 0xb8, 0xec, 0xd1, 0xb8, 0xc3, 0xc8, 0xc0, 0x23, 0xc7, 0x7d, 0x26, 0x4f,
	0x5c, 0x21, 0xb9, 0xe6, 0xb8, 0x1a, 0x79, 0x6e, 0x6e, 0xd9, 0x1d, 0x78, 0xd6, 0x66, 0xc0, 0x79,
	0x10, 0x31, 0x42, 0x45, 0x48, 0x68, 0x1c, 0x73, 0x4d, 0x75, 0xc8, 0x63, 0x95, 0x18, 0x1c, 0x06,
	0x8d, 0x57, 0x63, 0xff, 0x01, 0x8d, 0xc2, 0x2e, 0xd5, 0x5c, 0xee, 0xff, 0x75, 0xfb, 0xec, 0xb8,
	0xcf, 0x94, 0xc6, 0xeb, 0x50, 0x66, 0x82, 0x77, 0x8e, 0x36, 0x50, 0x03, 0x6d, 0x5d, 0xf1, 0x93,
	0x1f, 0xf8, 0x01, 0x54, 0x07, 0x13, 0xd3, 0x21, 0xed, 0x76, 0x25, 0x53, 0x6a, 0x63, 0xa5, 0x81,
	0xb6, 0x2a, 0xfe, 0x5a, 0xb6, 0xd0, 0x4a, 0x74, 0x67, 0x0f, 0x9a, 0x97, 0xc4, 0x28, 0xc1, 0x63,
	0xc5, 0xf0, 0x5d, 0xb8, 0x41, 0x83, 0x40, 0xb2, 0x80, 0x6a, 0x76, 0xf8, 0x56, 0xf1, 0xd8, 0x04,
	0x56, 0xfc, 0xeb, 0x99, 0xba, 0xa7, 0x78, 0xec, 0xec, 0xc2, 0xa6, 0x99, 0x95, 0x1b, 0xf1, 0x72,
	0x4c, 0x74, 0x29, 0xae, 0xf3, 0x0c, 0x6a, 0x73, 0x5c, 0x69, 0x7a, 0x0d, 0xc0, 0xec, 0xcc, 0x27,
	0x57, 0x8c, 0x62, 0x52, 0x3b, 0x60, 0x5f, 0xf4, 0xfb, 0x4c, 0x70, 0xa9, 0xd5, 0x7f, 0xbc, 0xa6,
	0x17, 0x50, 0x9f, 0x1b, 0x92, 0x62, 0x36, 0xe1, 0x9a, 0x4c, 0xa4, 0x3c, 0xe8, 0x6a, 0xaa, 0x19,
	0xd4, 0x7a, 0xf1, 0xa8, 0xfb, 0x54, 0xd2, 0xde, 0x84, 0xd4, 0x69, 0x15, 0xcf, 0x32, 0xd9, 0x90,
	0xa6, 0xd4, 0x61, 0x55, 0x18, 0x25, 0x1f, 0x02, 0x89, 0x34, 0xce, 0xd8, 0xfe, 0x5a, 0x86, 0xb2,
	0x99, 0x81, 0xbf, 0x21, 0x58, 0x9f, 0xf5, 0xac, 0x78, 0xc7, 0x2d, 0x94, 0xd1, 0xfd, 0x57, 0xd7,
	0xac, 0xdd, 0xe5, 0x4c, 0x09, 0xae, 0xd3, 0x7a, 0xff, 0xe3, 0xf7, 0xc7, 0x95, 0x27, 0xf8, 0x31,
	0x29, 0xfe, 0x3d, 0xb2, 0x4b, 0x56, 0xe4, 0xb4, 0xf0, 0x12, 0x67, 0xe4, 0xd4, 0x3c, 0xd3, 0x19,
	0xfe, 0x8c, 0x60, 0xed, 0x62, 0x37, 0x30, 0x99, 0x47, 0x33, 0xa7, 0x7b, 0xd6, 0xa3, 0xc5, 0x0d,
	0x29, 0xfa, 0x3d, 0x83, 0x7e, 0x1b, 0x37, 0x67, 0xa0, 0x1b, 0x34, 0x95, 0x21, 0x7e, 0x41, 0x80,
	0x8b, 0xcd, 0xc0, 0xde, 0x02, 0x99, 0xd3, 0x55, 0xb5, 0xb6, 0x97, 0xb1, 0xa4, 0xa0, 0xf7, 0x0d,
	0xe8, 0x1d, 0xec, 0xcc, 0x00, 0x4d, 0xdb, 0x97, 0x91, 0x7e, 0x42, 0x50, 0x2d, 0x94, 0x0b, 0x2f,
	0x72, 0x39, 0x53, 0x45, 0xb5, 0xbc, 0x25, 0x1c, 0x29, 0x66, 0xd3, 0x60, 0xde, 0xc2, 0x37, 0x67,
	0x60, 0x26, 0xfd, 0x7d, 0x7e, 0xf0, 0x7d, 0x68, 0xa3, 0xf3, 0xa1, 0x8d, 0x7e, 0x0d, 0x6d, 0xf4,
	0x61, 0x64, 0x97, 0xce, 0x47, 0x76, 0xe9, 0xe7, 0xc8, 0x2e, 0xbd, 0x7e, 0x1a, 0x84, 0xfa, 0xa8,
	0xdf, 0x76, 0x3b, 0xbc, 0x47, 0x14, 0x1f, 0x30, 0xc9, 0xc2, 0x20, 0x7e, 0x18, 0x79, 0xe3, 0x59,
	0xef, 0xa6, 0xa6, 0xe9, 0x13, 0xc1, 0x54, 0x5e, 0x11, 0xed, 0xf6, 0x55, 0xf3, 0x49, 0xdd, 0xf9,
	0x13, 0x00, 0x00, 0xff, 0xff, 0xb0, 0xbf, 0x0e, 0x43, 0xa4, 0x05, 0x00, 0x00,
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
	ValidatorPerformance(ctx context.Context, in *QueryValidatorPerformanceRequest, opts ...grpc.CallOption) (*QueryValidatorPerformanceResponse, error)
	PerformanceEpoch(ctx context.Context, in *QueryPerformanceEpochRequest, opts ...grpc.CallOption) (*QueryPerformanceEpochResponse, error)
	PerformanceReports(ctx context.Context, in *QueryPerformanceReportsRequest, opts ...grpc.CallOption) (*QueryPerformanceReportsResponse, error)
	PerformanceParams(ctx context.Context, in *QueryPerformanceParamsRequest, opts ...grpc.CallOption) (*QueryPerformanceParamsResponse, error)
}

type queryClient struct {
	cc grpc1.ClientConn
}

func NewQueryClient(cc grpc1.ClientConn) QueryClient {
	return &queryClient{cc}
}

func (c *queryClient) ValidatorPerformance(ctx context.Context, in *QueryValidatorPerformanceRequest, opts ...grpc.CallOption) (*QueryValidatorPerformanceResponse, error) {
	out := new(QueryValidatorPerformanceResponse)
	err := c.cc.Invoke(ctx, "/l1.performance.v1.Query/ValidatorPerformance", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) PerformanceEpoch(ctx context.Context, in *QueryPerformanceEpochRequest, opts ...grpc.CallOption) (*QueryPerformanceEpochResponse, error) {
	out := new(QueryPerformanceEpochResponse)
	err := c.cc.Invoke(ctx, "/l1.performance.v1.Query/PerformanceEpoch", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) PerformanceReports(ctx context.Context, in *QueryPerformanceReportsRequest, opts ...grpc.CallOption) (*QueryPerformanceReportsResponse, error) {
	out := new(QueryPerformanceReportsResponse)
	err := c.cc.Invoke(ctx, "/l1.performance.v1.Query/PerformanceReports", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) PerformanceParams(ctx context.Context, in *QueryPerformanceParamsRequest, opts ...grpc.CallOption) (*QueryPerformanceParamsResponse, error) {
	out := new(QueryPerformanceParamsResponse)
	err := c.cc.Invoke(ctx, "/l1.performance.v1.Query/PerformanceParams", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QueryServer is the server API for Query service.
type QueryServer interface {
	ValidatorPerformance(context.Context, *QueryValidatorPerformanceRequest) (*QueryValidatorPerformanceResponse, error)
	PerformanceEpoch(context.Context, *QueryPerformanceEpochRequest) (*QueryPerformanceEpochResponse, error)
	PerformanceReports(context.Context, *QueryPerformanceReportsRequest) (*QueryPerformanceReportsResponse, error)
	PerformanceParams(context.Context, *QueryPerformanceParamsRequest) (*QueryPerformanceParamsResponse, error)
}

// UnimplementedQueryServer can be embedded to have forward compatible implementations.
type UnimplementedQueryServer struct {
}

func (*UnimplementedQueryServer) ValidatorPerformance(ctx context.Context, req *QueryValidatorPerformanceRequest) (*QueryValidatorPerformanceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ValidatorPerformance not implemented")
}
func (*UnimplementedQueryServer) PerformanceEpoch(ctx context.Context, req *QueryPerformanceEpochRequest) (*QueryPerformanceEpochResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PerformanceEpoch not implemented")
}
func (*UnimplementedQueryServer) PerformanceReports(ctx context.Context, req *QueryPerformanceReportsRequest) (*QueryPerformanceReportsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PerformanceReports not implemented")
}
func (*UnimplementedQueryServer) PerformanceParams(ctx context.Context, req *QueryPerformanceParamsRequest) (*QueryPerformanceParamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PerformanceParams not implemented")
}

func RegisterQueryServer(s grpc1.Server, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc, srv)
}

func _Query_ValidatorPerformance_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryValidatorPerformanceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).ValidatorPerformance(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.performance.v1.Query/ValidatorPerformance",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).ValidatorPerformance(ctx, req.(*QueryValidatorPerformanceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_PerformanceEpoch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryPerformanceEpochRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).PerformanceEpoch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.performance.v1.Query/PerformanceEpoch",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).PerformanceEpoch(ctx, req.(*QueryPerformanceEpochRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_PerformanceReports_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryPerformanceReportsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).PerformanceReports(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.performance.v1.Query/PerformanceReports",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).PerformanceReports(ctx, req.(*QueryPerformanceReportsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_PerformanceParams_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryPerformanceParamsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).PerformanceParams(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.performance.v1.Query/PerformanceParams",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).PerformanceParams(ctx, req.(*QueryPerformanceParamsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var Query_serviceDesc = _Query_serviceDesc
var _Query_serviceDesc = grpc.ServiceDesc{
	ServiceName:	"l1.performance.v1.Query",
	HandlerType:	(*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName:	"ValidatorPerformance",
			Handler:	_Query_ValidatorPerformance_Handler,
		},
		{
			MethodName:	"PerformanceEpoch",
			Handler:	_Query_PerformanceEpoch_Handler,
		},
		{
			MethodName:	"PerformanceReports",
			Handler:	_Query_PerformanceReports_Handler,
		},
		{
			MethodName:	"PerformanceParams",
			Handler:	_Query_PerformanceParams_Handler,
		},
	},
	Streams:	[]grpc.StreamDesc{},
	Metadata:	"l1/performance/v1/query.proto",
}

func (m *QueryValidatorPerformanceRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryValidatorPerformanceRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryValidatorPerformanceRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.ValidatorAddress) > 0 {
		i -= len(m.ValidatorAddress)
		copy(dAtA[i:], m.ValidatorAddress)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.ValidatorAddress)))
		i--
		dAtA[i] = 0x12
	}
	if m.Epoch != 0 {
		i = encodeVarintQuery(dAtA, i, uint64(m.Epoch))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *QueryValidatorPerformanceResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryValidatorPerformanceResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryValidatorPerformanceResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.AggregateJson) > 0 {
		i -= len(m.AggregateJson)
		copy(dAtA[i:], m.AggregateJson)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.AggregateJson)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *QueryPerformanceEpochRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryPerformanceEpochRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryPerformanceEpochRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Epoch != 0 {
		i = encodeVarintQuery(dAtA, i, uint64(m.Epoch))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *QueryPerformanceEpochResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryPerformanceEpochResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryPerformanceEpochResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.EpochJson) > 0 {
		i -= len(m.EpochJson)
		copy(dAtA[i:], m.EpochJson)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.EpochJson)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *QueryPerformanceReportsRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryPerformanceReportsRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryPerformanceReportsRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.ValidatorAddress) > 0 {
		i -= len(m.ValidatorAddress)
		copy(dAtA[i:], m.ValidatorAddress)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.ValidatorAddress)))
		i--
		dAtA[i] = 0x12
	}
	if m.Epoch != 0 {
		i = encodeVarintQuery(dAtA, i, uint64(m.Epoch))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *QueryPerformanceReportsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryPerformanceReportsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryPerformanceReportsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.ReportsJson) > 0 {
		i -= len(m.ReportsJson)
		copy(dAtA[i:], m.ReportsJson)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.ReportsJson)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *QueryPerformanceParamsRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryPerformanceParamsRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryPerformanceParamsRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *QueryPerformanceParamsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryPerformanceParamsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryPerformanceParamsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.ParamsJson) > 0 {
		i -= len(m.ParamsJson)
		copy(dAtA[i:], m.ParamsJson)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.ParamsJson)))
		i--
		dAtA[i] = 0xa
	}
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
func (m *QueryValidatorPerformanceRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Epoch != 0 {
		n += 1 + sovQuery(uint64(m.Epoch))
	}
	l = len(m.ValidatorAddress)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *QueryValidatorPerformanceResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.AggregateJson)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *QueryPerformanceEpochRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Epoch != 0 {
		n += 1 + sovQuery(uint64(m.Epoch))
	}
	return n
}

func (m *QueryPerformanceEpochResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.EpochJson)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *QueryPerformanceReportsRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Epoch != 0 {
		n += 1 + sovQuery(uint64(m.Epoch))
	}
	l = len(m.ValidatorAddress)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *QueryPerformanceReportsResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ReportsJson)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *QueryPerformanceParamsRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *QueryPerformanceParamsResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ParamsJson)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func sovQuery(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozQuery(x uint64) (n int) {
	return sovQuery(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *QueryValidatorPerformanceRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryValidatorPerformanceRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryValidatorPerformanceRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Epoch", wireType)
			}
			m.Epoch = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
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
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ValidatorAddress", wireType)
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
			m.ValidatorAddress = string(dAtA[iNdEx:postIndex])
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
func (m *QueryValidatorPerformanceResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryValidatorPerformanceResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryValidatorPerformanceResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AggregateJson", wireType)
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
			m.AggregateJson = string(dAtA[iNdEx:postIndex])
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
func (m *QueryPerformanceEpochRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryPerformanceEpochRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryPerformanceEpochRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Epoch", wireType)
			}
			m.Epoch = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
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
func (m *QueryPerformanceEpochResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryPerformanceEpochResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryPerformanceEpochResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field EpochJson", wireType)
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
			m.EpochJson = string(dAtA[iNdEx:postIndex])
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
func (m *QueryPerformanceReportsRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryPerformanceReportsRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryPerformanceReportsRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Epoch", wireType)
			}
			m.Epoch = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
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
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ValidatorAddress", wireType)
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
			m.ValidatorAddress = string(dAtA[iNdEx:postIndex])
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
func (m *QueryPerformanceReportsResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryPerformanceReportsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryPerformanceReportsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ReportsJson", wireType)
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
			m.ReportsJson = string(dAtA[iNdEx:postIndex])
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
func (m *QueryPerformanceParamsRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryPerformanceParamsRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryPerformanceParamsRequest: illegal tag %d (wire type %d)", fieldNum, wire)
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
func (m *QueryPerformanceParamsResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryPerformanceParamsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryPerformanceParamsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ParamsJson", wireType)
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
			m.ParamsJson = string(dAtA[iNdEx:postIndex])
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
