package types

import (
	context "context"
	fmt "fmt"
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

type QueryEmissionsParamsRequest struct {
}

func (m *QueryEmissionsParamsRequest) Reset()		{ *m = QueryEmissionsParamsRequest{} }
func (m *QueryEmissionsParamsRequest) String() string	{ return proto.CompactTextString(m) }
func (*QueryEmissionsParamsRequest) ProtoMessage()	{}
func (*QueryEmissionsParamsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_a9affa4fcc126b24, []int{0}
}
func (m *QueryEmissionsParamsRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryEmissionsParamsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryEmissionsParamsRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryEmissionsParamsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryEmissionsParamsRequest.Merge(m, src)
}
func (m *QueryEmissionsParamsRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryEmissionsParamsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryEmissionsParamsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryEmissionsParamsRequest proto.InternalMessageInfo

type QueryEmissionsParamsResponse struct {
	Params Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
}

func (m *QueryEmissionsParamsResponse) Reset()		{ *m = QueryEmissionsParamsResponse{} }
func (m *QueryEmissionsParamsResponse) String() string	{ return proto.CompactTextString(m) }
func (*QueryEmissionsParamsResponse) ProtoMessage()	{}
func (*QueryEmissionsParamsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_a9affa4fcc126b24, []int{1}
}
func (m *QueryEmissionsParamsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryEmissionsParamsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryEmissionsParamsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryEmissionsParamsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryEmissionsParamsResponse.Merge(m, src)
}
func (m *QueryEmissionsParamsResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryEmissionsParamsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryEmissionsParamsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryEmissionsParamsResponse proto.InternalMessageInfo

func (m *QueryEmissionsParamsResponse) GetParams() Params {
	if m != nil {
		return m.Params
	}
	return Params{}
}

type QueryCurrentInflationRequest struct {
}

func (m *QueryCurrentInflationRequest) Reset()		{ *m = QueryCurrentInflationRequest{} }
func (m *QueryCurrentInflationRequest) String() string	{ return proto.CompactTextString(m) }
func (*QueryCurrentInflationRequest) ProtoMessage()	{}
func (*QueryCurrentInflationRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_a9affa4fcc126b24, []int{2}
}
func (m *QueryCurrentInflationRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryCurrentInflationRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryCurrentInflationRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryCurrentInflationRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryCurrentInflationRequest.Merge(m, src)
}
func (m *QueryCurrentInflationRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryCurrentInflationRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryCurrentInflationRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryCurrentInflationRequest proto.InternalMessageInfo

type QueryCurrentInflationResponse struct {
	CurrentInflationBps uint32 `protobuf:"varint,1,opt,name=current_inflation_bps,json=currentInflationBps,proto3" json:"current_inflation_bps,omitempty"`
}

func (m *QueryCurrentInflationResponse) Reset()		{ *m = QueryCurrentInflationResponse{} }
func (m *QueryCurrentInflationResponse) String() string	{ return proto.CompactTextString(m) }
func (*QueryCurrentInflationResponse) ProtoMessage()	{}
func (*QueryCurrentInflationResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_a9affa4fcc126b24, []int{3}
}
func (m *QueryCurrentInflationResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryCurrentInflationResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryCurrentInflationResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryCurrentInflationResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryCurrentInflationResponse.Merge(m, src)
}
func (m *QueryCurrentInflationResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryCurrentInflationResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryCurrentInflationResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryCurrentInflationResponse proto.InternalMessageInfo

func (m *QueryCurrentInflationResponse) GetCurrentInflationBps() uint32 {
	if m != nil {
		return m.CurrentInflationBps
	}
	return 0
}

type QueryEmissionEpochRequest struct {
	Epoch uint64 `protobuf:"varint,1,opt,name=epoch,proto3" json:"epoch,omitempty"`
}

func (m *QueryEmissionEpochRequest) Reset()		{ *m = QueryEmissionEpochRequest{} }
func (m *QueryEmissionEpochRequest) String() string	{ return proto.CompactTextString(m) }
func (*QueryEmissionEpochRequest) ProtoMessage()	{}
func (*QueryEmissionEpochRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_a9affa4fcc126b24, []int{4}
}
func (m *QueryEmissionEpochRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryEmissionEpochRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryEmissionEpochRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryEmissionEpochRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryEmissionEpochRequest.Merge(m, src)
}
func (m *QueryEmissionEpochRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryEmissionEpochRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryEmissionEpochRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryEmissionEpochRequest proto.InternalMessageInfo

func (m *QueryEmissionEpochRequest) GetEpoch() uint64 {
	if m != nil {
		return m.Epoch
	}
	return 0
}

type QueryEmissionEpochResponse struct {
	EmissionEpoch EmissionEpoch `protobuf:"bytes,1,opt,name=emission_epoch,json=emissionEpoch,proto3" json:"emission_epoch"`
}

func (m *QueryEmissionEpochResponse) Reset()		{ *m = QueryEmissionEpochResponse{} }
func (m *QueryEmissionEpochResponse) String() string	{ return proto.CompactTextString(m) }
func (*QueryEmissionEpochResponse) ProtoMessage()	{}
func (*QueryEmissionEpochResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_a9affa4fcc126b24, []int{5}
}
func (m *QueryEmissionEpochResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryEmissionEpochResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryEmissionEpochResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryEmissionEpochResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryEmissionEpochResponse.Merge(m, src)
}
func (m *QueryEmissionEpochResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryEmissionEpochResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryEmissionEpochResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryEmissionEpochResponse proto.InternalMessageInfo

func (m *QueryEmissionEpochResponse) GetEmissionEpoch() EmissionEpoch {
	if m != nil {
		return m.EmissionEpoch
	}
	return EmissionEpoch{}
}

type QueryDistributionWeightsRequest struct {
}

func (m *QueryDistributionWeightsRequest) Reset()		{ *m = QueryDistributionWeightsRequest{} }
func (m *QueryDistributionWeightsRequest) String() string	{ return proto.CompactTextString(m) }
func (*QueryDistributionWeightsRequest) ProtoMessage()		{}
func (*QueryDistributionWeightsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_a9affa4fcc126b24, []int{6}
}
func (m *QueryDistributionWeightsRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryDistributionWeightsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryDistributionWeightsRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryDistributionWeightsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryDistributionWeightsRequest.Merge(m, src)
}
func (m *QueryDistributionWeightsRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryDistributionWeightsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryDistributionWeightsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryDistributionWeightsRequest proto.InternalMessageInfo

type QueryDistributionWeightsResponse struct {
	Weights DistributionWeights `protobuf:"bytes,1,opt,name=weights,proto3" json:"weights"`
}

func (m *QueryDistributionWeightsResponse) Reset()		{ *m = QueryDistributionWeightsResponse{} }
func (m *QueryDistributionWeightsResponse) String() string	{ return proto.CompactTextString(m) }
func (*QueryDistributionWeightsResponse) ProtoMessage()		{}
func (*QueryDistributionWeightsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_a9affa4fcc126b24, []int{7}
}
func (m *QueryDistributionWeightsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryDistributionWeightsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryDistributionWeightsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryDistributionWeightsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryDistributionWeightsResponse.Merge(m, src)
}
func (m *QueryDistributionWeightsResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryDistributionWeightsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryDistributionWeightsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryDistributionWeightsResponse proto.InternalMessageInfo

func (m *QueryDistributionWeightsResponse) GetWeights() DistributionWeights {
	if m != nil {
		return m.Weights
	}
	return DistributionWeights{}
}

func init() {
	proto.RegisterType((*QueryEmissionsParamsRequest)(nil), "l1.emissions.v1.QueryEmissionsParamsRequest")
	proto.RegisterType((*QueryEmissionsParamsResponse)(nil), "l1.emissions.v1.QueryEmissionsParamsResponse")
	proto.RegisterType((*QueryCurrentInflationRequest)(nil), "l1.emissions.v1.QueryCurrentInflationRequest")
	proto.RegisterType((*QueryCurrentInflationResponse)(nil), "l1.emissions.v1.QueryCurrentInflationResponse")
	proto.RegisterType((*QueryEmissionEpochRequest)(nil), "l1.emissions.v1.QueryEmissionEpochRequest")
	proto.RegisterType((*QueryEmissionEpochResponse)(nil), "l1.emissions.v1.QueryEmissionEpochResponse")
	proto.RegisterType((*QueryDistributionWeightsRequest)(nil), "l1.emissions.v1.QueryDistributionWeightsRequest")
	proto.RegisterType((*QueryDistributionWeightsResponse)(nil), "l1.emissions.v1.QueryDistributionWeightsResponse")
}

func init()	{ proto.RegisterFile("l1/emissions/v1/query.proto", fileDescriptor_a9affa4fcc126b24) }

var fileDescriptor_a9affa4fcc126b24 = []byte{

	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x94, 0x4f, 0x6f, 0xd3, 0x30,
	0x18, 0xc6, 0x6b, 0xd4, 0x0d, 0xe9, 0x9d, 0xca, 0x90, 0x37, 0x34, 0x96, 0xad, 0xe9, 0x16, 0x0d,
	0x0d, 0x8d, 0x35, 0x26, 0x45, 0x7c, 0x81, 0xb2, 0x21, 0x21, 0x2e, 0x50, 0x84, 0x90, 0xb8, 0x54,
	0x6d, 0x31, 0xa9, 0xa5, 0x34, 0xce, 0x62, 0xa7, 0x30, 0x21, 0x2e, 0x7c, 0x01, 0x40, 0x9c, 0xf9,
	0x00, 0x7c, 0x93, 0x1d, 0x27, 0x71, 0xe1, 0x84, 0x50, 0xcb, 0x97, 0xe0, 0x86, 0xe2, 0x38, 0xa3,
	0xcd, 0x9f, 0xa9, 0x9c, 0x1a, 0xfb, 0x7d, 0x9f, 0xf7, 0xf9, 0xd9, 0x7e, 0x54, 0xd8, 0xf2, 0x1c,
	0x42, 0x47, 0x4c, 0x08, 0xc6, 0x7d, 0x41, 0xc6, 0x0e, 0x39, 0x89, 0x68, 0x78, 0x6a, 0x07, 0x21,
	0x97, 0x1c, 0xaf, 0x7a, 0x8e, 0x7d, 0x51, 0xb4, 0xc7, 0x8e, 0xb1, 0xed, 0x72, 0xee, 0x7a, 0x94,
	0xf4, 0x02, 0x46, 0x7a, 0xbe, 0xcf, 0x65, 0x4f, 0xaa, 0x92, 0x6a, 0x37, 0xd6, 0x5d, 0xee, 0x72,
	0xf5, 0x49, 0xe2, 0x2f, 0xbd, 0x5b, 0xcf, 0x3a, 0xb8, 0xd4, 0xa7, 0x82, 0x69, 0x91, 0x55, 0x87,
	0xad, 0xa7, 0xb1, 0xe5, 0x71, 0xda, 0xf2, 0xa4, 0x17, 0xf6, 0x46, 0xa2, 0x43, 0x4f, 0x22, 0x2a,
	0xa4, 0xf5, 0x1c, 0xb6, 0x8b, 0xcb, 0x22, 0xe0, 0xbe, 0xa0, 0xf8, 0x3e, 0x2c, 0x07, 0x6a, 0xe7,
	0x26, 0xda, 0x41, 0xb7, 0x57, 0x5a, 0x1b, 0x76, 0x86, 0xd9, 0x4e, 0x04, 0xed, 0xea, 0xd9, 0xcf,
	0x46, 0xa5, 0xa3, 0x9b, 0x2d, 0x53, 0x8f, 0x7d, 0x10, 0x85, 0x21, 0xf5, 0xe5, 0x23, 0xff, 0xb5,
	0xa7, 0x8e, 0x92, 0xda, 0x3e, 0x83, 0x7a, 0x49, 0x5d, 0xfb, 0xb6, 0xe0, 0xc6, 0x20, 0xa9, 0x75,
	0x59, 0x5a, 0xec, 0xf6, 0x83, 0x04, 0xa3, 0xd6, 0x59, 0x1b, 0x64, 0x84, 0xed, 0x40, 0x58, 0x0e,
	0x6c, 0xce, 0x9d, 0xe5, 0x38, 0xe0, 0x83, 0xa1, 0x76, 0xc4, 0xeb, 0xb0, 0x44, 0xe3, 0xb5, 0x1a,
	0x50, 0xed, 0x24, 0x0b, 0x8b, 0x81, 0x51, 0x24, 0xd1, 0x10, 0x8f, 0xe1, 0x5a, 0x7a, 0xd4, 0xee,
	0x3f, 0xf1, 0x4a, 0xcb, 0xcc, 0x5d, 0xc2, 0x9c, 0x5e, 0xdf, 0x45, 0x8d, 0xce, 0x6e, 0x5a, 0xbb,
	0xd0, 0x50, 0x56, 0x47, 0x4c, 0xc8, 0x90, 0xf5, 0xa3, 0x98, 0xfa, 0x05, 0x65, 0xee, 0x50, 0x5e,
	0x3c, 0xc6, 0x10, 0x76, 0xca, 0x5b, 0x34, 0xd3, 0x11, 0x5c, 0x7d, 0x93, 0x6c, 0x69, 0x98, 0xbd,
	0x1c, 0x4c, 0x81, 0x5c, 0x23, 0xa5, 0xd2, 0xd6, 0x9f, 0x2a, 0x2c, 0x29, 0x2b, 0xfc, 0x11, 0xc1,
	0x6a, 0xe6, 0xf1, 0xf1, 0x61, 0x6e, 0xe4, 0x25, 0x11, 0x32, 0x9a, 0x0b, 0x76, 0x27, 0x07, 0xb0,
	0x1a, 0x1f, 0xbe, 0xff, 0xfe, 0x72, 0x65, 0x13, 0x6f, 0x90, 0x6c, 0x70, 0x93, 0xec, 0xe0, 0xaf,
	0x08, 0xae, 0x67, 0x73, 0x81, 0x4b, 0x4c, 0x4a, 0xf2, 0x65, 0xd8, 0x8b, 0xb6, 0x6b, 0xa8, 0x03,
	0x05, 0xb5, 0x87, 0xad, 0x1c, 0x54, 0x2e, 0x85, 0xf8, 0x33, 0x82, 0xda, 0xdc, 0x7b, 0xe3, 0x83,
	0xcb, 0x6f, 0x60, 0x36, 0x87, 0xc6, 0x9d, 0x85, 0x7a, 0x35, 0xd6, 0xbe, 0xc2, 0xda, 0xc5, 0x8d,
	0x1c, 0x96, 0x8a, 0xa3, 0x20, 0xef, 0xd4, 0xef, 0x7b, 0xfc, 0x0d, 0xc1, 0x5a, 0xc1, 0xb3, 0xe3,
	0xbb, 0xc5, 0x6e, 0xe5, 0x19, 0x34, 0x9c, 0xff, 0x50, 0x68, 0xca, 0xa6, 0xa2, 0xdc, 0xc7, 0xb7,
	0x72, 0x94, 0xaf, 0x66, 0x54, 0x5d, 0x9d, 0xbd, 0xf6, 0xc3, 0xb3, 0x89, 0x89, 0xce, 0x27, 0x26,
	0xfa, 0x35, 0x31, 0xd1, 0xa7, 0xa9, 0x59, 0x39, 0x9f, 0x9a, 0x95, 0x1f, 0x53, 0xb3, 0xf2, 0xf2,
	0xd0, 0x65, 0x72, 0x18, 0xf5, 0xed, 0x01, 0x1f, 0x11, 0xc1, 0xc7, 0x34, 0xa4, 0xcc, 0xf5, 0x9b,
	0x9e, 0x13, 0xcf, 0x7d, 0x3b, 0x33, 0x59, 0x9e, 0x06, 0x54, 0xf4, 0x97, 0xd5, 0x1f, 0xdc, 0xbd,
	0xbf, 0x01, 0x00, 0x00, 0xff, 0xff, 0x41, 0xe6, 0x92, 0xcf, 0x63, 0x05, 0x00, 0x00,
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
	EmissionsParams(ctx context.Context, in *QueryEmissionsParamsRequest, opts ...grpc.CallOption) (*QueryEmissionsParamsResponse, error)
	CurrentInflation(ctx context.Context, in *QueryCurrentInflationRequest, opts ...grpc.CallOption) (*QueryCurrentInflationResponse, error)
	EmissionEpoch(ctx context.Context, in *QueryEmissionEpochRequest, opts ...grpc.CallOption) (*QueryEmissionEpochResponse, error)
	DistributionWeights(ctx context.Context, in *QueryDistributionWeightsRequest, opts ...grpc.CallOption) (*QueryDistributionWeightsResponse, error)
}

type queryClient struct {
	cc grpc1.ClientConn
}

func NewQueryClient(cc grpc1.ClientConn) QueryClient {
	return &queryClient{cc}
}

func (c *queryClient) EmissionsParams(ctx context.Context, in *QueryEmissionsParamsRequest, opts ...grpc.CallOption) (*QueryEmissionsParamsResponse, error) {
	out := new(QueryEmissionsParamsResponse)
	err := c.cc.Invoke(ctx, "/l1.emissions.v1.Query/EmissionsParams", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) CurrentInflation(ctx context.Context, in *QueryCurrentInflationRequest, opts ...grpc.CallOption) (*QueryCurrentInflationResponse, error) {
	out := new(QueryCurrentInflationResponse)
	err := c.cc.Invoke(ctx, "/l1.emissions.v1.Query/CurrentInflation", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) EmissionEpoch(ctx context.Context, in *QueryEmissionEpochRequest, opts ...grpc.CallOption) (*QueryEmissionEpochResponse, error) {
	out := new(QueryEmissionEpochResponse)
	err := c.cc.Invoke(ctx, "/l1.emissions.v1.Query/EmissionEpoch", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) DistributionWeights(ctx context.Context, in *QueryDistributionWeightsRequest, opts ...grpc.CallOption) (*QueryDistributionWeightsResponse, error) {
	out := new(QueryDistributionWeightsResponse)
	err := c.cc.Invoke(ctx, "/l1.emissions.v1.Query/DistributionWeights", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QueryServer is the server API for Query service.
type QueryServer interface {
	EmissionsParams(context.Context, *QueryEmissionsParamsRequest) (*QueryEmissionsParamsResponse, error)
	CurrentInflation(context.Context, *QueryCurrentInflationRequest) (*QueryCurrentInflationResponse, error)
	EmissionEpoch(context.Context, *QueryEmissionEpochRequest) (*QueryEmissionEpochResponse, error)
	DistributionWeights(context.Context, *QueryDistributionWeightsRequest) (*QueryDistributionWeightsResponse, error)
}

// UnimplementedQueryServer can be embedded to have forward compatible implementations.
type UnimplementedQueryServer struct {
}

func (*UnimplementedQueryServer) EmissionsParams(ctx context.Context, req *QueryEmissionsParamsRequest) (*QueryEmissionsParamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method EmissionsParams not implemented")
}
func (*UnimplementedQueryServer) CurrentInflation(ctx context.Context, req *QueryCurrentInflationRequest) (*QueryCurrentInflationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CurrentInflation not implemented")
}
func (*UnimplementedQueryServer) EmissionEpoch(ctx context.Context, req *QueryEmissionEpochRequest) (*QueryEmissionEpochResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method EmissionEpoch not implemented")
}
func (*UnimplementedQueryServer) DistributionWeights(ctx context.Context, req *QueryDistributionWeightsRequest) (*QueryDistributionWeightsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DistributionWeights not implemented")
}

func RegisterQueryServer(s grpc1.Server, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc, srv)
}

func _Query_EmissionsParams_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryEmissionsParamsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).EmissionsParams(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.emissions.v1.Query/EmissionsParams",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).EmissionsParams(ctx, req.(*QueryEmissionsParamsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_CurrentInflation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryCurrentInflationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).CurrentInflation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.emissions.v1.Query/CurrentInflation",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).CurrentInflation(ctx, req.(*QueryCurrentInflationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_EmissionEpoch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryEmissionEpochRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).EmissionEpoch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.emissions.v1.Query/EmissionEpoch",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).EmissionEpoch(ctx, req.(*QueryEmissionEpochRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_DistributionWeights_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryDistributionWeightsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).DistributionWeights(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.emissions.v1.Query/DistributionWeights",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).DistributionWeights(ctx, req.(*QueryDistributionWeightsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var Query_serviceDesc = _Query_serviceDesc
var _Query_serviceDesc = grpc.ServiceDesc{
	ServiceName:	"l1.emissions.v1.Query",
	HandlerType:	(*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName:	"EmissionsParams",
			Handler:	_Query_EmissionsParams_Handler,
		},
		{
			MethodName:	"CurrentInflation",
			Handler:	_Query_CurrentInflation_Handler,
		},
		{
			MethodName:	"EmissionEpoch",
			Handler:	_Query_EmissionEpoch_Handler,
		},
		{
			MethodName:	"DistributionWeights",
			Handler:	_Query_DistributionWeights_Handler,
		},
	},
	Streams:	[]grpc.StreamDesc{},
	Metadata:	"l1/emissions/v1/query.proto",
}

func (m *QueryEmissionsParamsRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryEmissionsParamsRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryEmissionsParamsRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *QueryEmissionsParamsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryEmissionsParamsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryEmissionsParamsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
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

func (m *QueryCurrentInflationRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryCurrentInflationRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryCurrentInflationRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *QueryCurrentInflationResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryCurrentInflationResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryCurrentInflationResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.CurrentInflationBps != 0 {
		i = encodeVarintQuery(dAtA, i, uint64(m.CurrentInflationBps))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *QueryEmissionEpochRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryEmissionEpochRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryEmissionEpochRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
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

func (m *QueryEmissionEpochResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryEmissionEpochResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryEmissionEpochResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.EmissionEpoch.MarshalToSizedBuffer(dAtA[:i])
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

func (m *QueryDistributionWeightsRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryDistributionWeightsRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryDistributionWeightsRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *QueryDistributionWeightsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryDistributionWeightsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryDistributionWeightsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.Weights.MarshalToSizedBuffer(dAtA[:i])
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
func (m *QueryEmissionsParamsRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *QueryEmissionsParamsResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Params.Size()
	n += 1 + l + sovQuery(uint64(l))
	return n
}

func (m *QueryCurrentInflationRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *QueryCurrentInflationResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.CurrentInflationBps != 0 {
		n += 1 + sovQuery(uint64(m.CurrentInflationBps))
	}
	return n
}

func (m *QueryEmissionEpochRequest) Size() (n int) {
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

func (m *QueryEmissionEpochResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.EmissionEpoch.Size()
	n += 1 + l + sovQuery(uint64(l))
	return n
}

func (m *QueryDistributionWeightsRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *QueryDistributionWeightsResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Weights.Size()
	n += 1 + l + sovQuery(uint64(l))
	return n
}

func sovQuery(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozQuery(x uint64) (n int) {
	return sovQuery(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *QueryEmissionsParamsRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryEmissionsParamsRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryEmissionsParamsRequest: illegal tag %d (wire type %d)", fieldNum, wire)
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
func (m *QueryEmissionsParamsResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryEmissionsParamsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryEmissionsParamsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
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
func (m *QueryCurrentInflationRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryCurrentInflationRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryCurrentInflationRequest: illegal tag %d (wire type %d)", fieldNum, wire)
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
func (m *QueryCurrentInflationResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryCurrentInflationResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryCurrentInflationResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field CurrentInflationBps", wireType)
			}
			m.CurrentInflationBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.CurrentInflationBps |= uint32(b&0x7F) << shift
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
func (m *QueryEmissionEpochRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryEmissionEpochRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryEmissionEpochRequest: illegal tag %d (wire type %d)", fieldNum, wire)
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
func (m *QueryEmissionEpochResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryEmissionEpochResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryEmissionEpochResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field EmissionEpoch", wireType)
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
			if err := m.EmissionEpoch.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
func (m *QueryDistributionWeightsRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryDistributionWeightsRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryDistributionWeightsRequest: illegal tag %d (wire type %d)", fieldNum, wire)
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
func (m *QueryDistributionWeightsResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryDistributionWeightsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryDistributionWeightsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Weights", wireType)
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
			if err := m.Weights.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
