package delegatorprotectionpb

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

type QueryProtectionFundRequest struct {
}

func (m *QueryProtectionFundRequest) Reset()		{ *m = QueryProtectionFundRequest{} }
func (m *QueryProtectionFundRequest) String() string	{ return proto.CompactTextString(m) }
func (*QueryProtectionFundRequest) ProtoMessage()	{}
func (*QueryProtectionFundRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_9cede56dd8b408be, []int{0}
}
func (m *QueryProtectionFundRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryProtectionFundRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryProtectionFundRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryProtectionFundRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryProtectionFundRequest.Merge(m, src)
}
func (m *QueryProtectionFundRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryProtectionFundRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryProtectionFundRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryProtectionFundRequest proto.InternalMessageInfo

type QueryProtectionFundResponse struct {
	FundJson string `protobuf:"bytes,1,opt,name=fund_json,json=fundJson,proto3" json:"fund_json,omitempty"`
}

func (m *QueryProtectionFundResponse) Reset()		{ *m = QueryProtectionFundResponse{} }
func (m *QueryProtectionFundResponse) String() string	{ return proto.CompactTextString(m) }
func (*QueryProtectionFundResponse) ProtoMessage()	{}
func (*QueryProtectionFundResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_9cede56dd8b408be, []int{1}
}
func (m *QueryProtectionFundResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryProtectionFundResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryProtectionFundResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryProtectionFundResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryProtectionFundResponse.Merge(m, src)
}
func (m *QueryProtectionFundResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryProtectionFundResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryProtectionFundResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryProtectionFundResponse proto.InternalMessageInfo

func (m *QueryProtectionFundResponse) GetFundJson() string {
	if m != nil {
		return m.FundJson
	}
	return ""
}

type QueryProtectionClaimsRequest struct {
	Delegator	string	`protobuf:"bytes,1,opt,name=delegator,proto3" json:"delegator,omitempty"`
	Status		string	`protobuf:"bytes,2,opt,name=status,proto3" json:"status,omitempty"`
}

func (m *QueryProtectionClaimsRequest) Reset()		{ *m = QueryProtectionClaimsRequest{} }
func (m *QueryProtectionClaimsRequest) String() string	{ return proto.CompactTextString(m) }
func (*QueryProtectionClaimsRequest) ProtoMessage()	{}
func (*QueryProtectionClaimsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_9cede56dd8b408be, []int{2}
}
func (m *QueryProtectionClaimsRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryProtectionClaimsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryProtectionClaimsRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryProtectionClaimsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryProtectionClaimsRequest.Merge(m, src)
}
func (m *QueryProtectionClaimsRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryProtectionClaimsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryProtectionClaimsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryProtectionClaimsRequest proto.InternalMessageInfo

func (m *QueryProtectionClaimsRequest) GetDelegator() string {
	if m != nil {
		return m.Delegator
	}
	return ""
}

func (m *QueryProtectionClaimsRequest) GetStatus() string {
	if m != nil {
		return m.Status
	}
	return ""
}

type QueryProtectionClaimsResponse struct {
	ClaimsJson string `protobuf:"bytes,1,opt,name=claims_json,json=claimsJson,proto3" json:"claims_json,omitempty"`
}

func (m *QueryProtectionClaimsResponse) Reset()		{ *m = QueryProtectionClaimsResponse{} }
func (m *QueryProtectionClaimsResponse) String() string	{ return proto.CompactTextString(m) }
func (*QueryProtectionClaimsResponse) ProtoMessage()	{}
func (*QueryProtectionClaimsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_9cede56dd8b408be, []int{3}
}
func (m *QueryProtectionClaimsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryProtectionClaimsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryProtectionClaimsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryProtectionClaimsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryProtectionClaimsResponse.Merge(m, src)
}
func (m *QueryProtectionClaimsResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryProtectionClaimsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryProtectionClaimsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryProtectionClaimsResponse proto.InternalMessageInfo

func (m *QueryProtectionClaimsResponse) GetClaimsJson() string {
	if m != nil {
		return m.ClaimsJson
	}
	return ""
}

type QueryDelegatorCompensationRequest struct {
	Delegator string `protobuf:"bytes,1,opt,name=delegator,proto3" json:"delegator,omitempty"`
}

func (m *QueryDelegatorCompensationRequest) Reset()		{ *m = QueryDelegatorCompensationRequest{} }
func (m *QueryDelegatorCompensationRequest) String() string	{ return proto.CompactTextString(m) }
func (*QueryDelegatorCompensationRequest) ProtoMessage()	{}
func (*QueryDelegatorCompensationRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_9cede56dd8b408be, []int{4}
}
func (m *QueryDelegatorCompensationRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryDelegatorCompensationRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryDelegatorCompensationRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryDelegatorCompensationRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryDelegatorCompensationRequest.Merge(m, src)
}
func (m *QueryDelegatorCompensationRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryDelegatorCompensationRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryDelegatorCompensationRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryDelegatorCompensationRequest proto.InternalMessageInfo

func (m *QueryDelegatorCompensationRequest) GetDelegator() string {
	if m != nil {
		return m.Delegator
	}
	return ""
}

type QueryDelegatorCompensationResponse struct {
	PayoutsJson string `protobuf:"bytes,1,opt,name=payouts_json,json=payoutsJson,proto3" json:"payouts_json,omitempty"`
}

func (m *QueryDelegatorCompensationResponse) Reset()		{ *m = QueryDelegatorCompensationResponse{} }
func (m *QueryDelegatorCompensationResponse) String() string	{ return proto.CompactTextString(m) }
func (*QueryDelegatorCompensationResponse) ProtoMessage()	{}
func (*QueryDelegatorCompensationResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_9cede56dd8b408be, []int{5}
}
func (m *QueryDelegatorCompensationResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryDelegatorCompensationResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryDelegatorCompensationResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryDelegatorCompensationResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryDelegatorCompensationResponse.Merge(m, src)
}
func (m *QueryDelegatorCompensationResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryDelegatorCompensationResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryDelegatorCompensationResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryDelegatorCompensationResponse proto.InternalMessageInfo

func (m *QueryDelegatorCompensationResponse) GetPayoutsJson() string {
	if m != nil {
		return m.PayoutsJson
	}
	return ""
}

type QueryProtectionParamsRequest struct {
}

func (m *QueryProtectionParamsRequest) Reset()		{ *m = QueryProtectionParamsRequest{} }
func (m *QueryProtectionParamsRequest) String() string	{ return proto.CompactTextString(m) }
func (*QueryProtectionParamsRequest) ProtoMessage()	{}
func (*QueryProtectionParamsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_9cede56dd8b408be, []int{6}
}
func (m *QueryProtectionParamsRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryProtectionParamsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryProtectionParamsRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryProtectionParamsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryProtectionParamsRequest.Merge(m, src)
}
func (m *QueryProtectionParamsRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryProtectionParamsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryProtectionParamsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryProtectionParamsRequest proto.InternalMessageInfo

type QueryProtectionParamsResponse struct {
	ParamsJson string `protobuf:"bytes,1,opt,name=params_json,json=paramsJson,proto3" json:"params_json,omitempty"`
}

func (m *QueryProtectionParamsResponse) Reset()		{ *m = QueryProtectionParamsResponse{} }
func (m *QueryProtectionParamsResponse) String() string	{ return proto.CompactTextString(m) }
func (*QueryProtectionParamsResponse) ProtoMessage()	{}
func (*QueryProtectionParamsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_9cede56dd8b408be, []int{7}
}
func (m *QueryProtectionParamsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryProtectionParamsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryProtectionParamsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryProtectionParamsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryProtectionParamsResponse.Merge(m, src)
}
func (m *QueryProtectionParamsResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryProtectionParamsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryProtectionParamsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryProtectionParamsResponse proto.InternalMessageInfo

func (m *QueryProtectionParamsResponse) GetParamsJson() string {
	if m != nil {
		return m.ParamsJson
	}
	return ""
}

func init() {
	proto.RegisterType((*QueryProtectionFundRequest)(nil), "l1.delegatorprotection.v1.QueryProtectionFundRequest")
	proto.RegisterType((*QueryProtectionFundResponse)(nil), "l1.delegatorprotection.v1.QueryProtectionFundResponse")
	proto.RegisterType((*QueryProtectionClaimsRequest)(nil), "l1.delegatorprotection.v1.QueryProtectionClaimsRequest")
	proto.RegisterType((*QueryProtectionClaimsResponse)(nil), "l1.delegatorprotection.v1.QueryProtectionClaimsResponse")
	proto.RegisterType((*QueryDelegatorCompensationRequest)(nil), "l1.delegatorprotection.v1.QueryDelegatorCompensationRequest")
	proto.RegisterType((*QueryDelegatorCompensationResponse)(nil), "l1.delegatorprotection.v1.QueryDelegatorCompensationResponse")
	proto.RegisterType((*QueryProtectionParamsRequest)(nil), "l1.delegatorprotection.v1.QueryProtectionParamsRequest")
	proto.RegisterType((*QueryProtectionParamsResponse)(nil), "l1.delegatorprotection.v1.QueryProtectionParamsResponse")
}

func init() {
	proto.RegisterFile("l1/delegatorprotection/v1/query.proto", fileDescriptor_9cede56dd8b408be)
}

var fileDescriptor_9cede56dd8b408be = []byte{

	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x54, 0xcf, 0x8b, 0xd3, 0x40,
	0x14, 0xee, 0x2c, 0xec, 0x62, 0xdf, 0x8a, 0xc8, 0x80, 0xb2, 0x66, 0x6b, 0xd6, 0x46, 0x44, 0x3d,
	0x6c, 0x86, 0xb8, 0xac, 0x8a, 0x3f, 0x40, 0x5d, 0x51, 0x10, 0x84, 0x75, 0xf1, 0xe4, 0x45, 0xa6,
	0xed, 0x18, 0x23, 0xe9, 0xcc, 0x6c, 0x66, 0x52, 0x2c, 0xe2, 0xc5, 0xbf, 0x40, 0xf0, 0x9f, 0xf0,
	0xe8, 0x9f, 0xe1, 0x71, 0xc5, 0x8b, 0x47, 0x69, 0x3d, 0xfa, 0x47, 0x48, 0x26, 0xb3, 0x6d, 0x53,
	0xd3, 0xd8, 0x78, 0xcc, 0xf7, 0xe6, 0xfb, 0xde, 0xf7, 0xde, 0xfb, 0x08, 0x5c, 0x8a, 0x03, 0xd2,
	0x63, 0x31, 0x0b, 0xa9, 0x16, 0x89, 0x4c, 0x84, 0x66, 0x5d, 0x1d, 0x09, 0x4e, 0x06, 0x01, 0x39,
	0x4c, 0x59, 0x32, 0xf4, 0x33, 0x4c, 0xe0, 0x73, 0x71, 0xe0, 0x97, 0x3c, 0xf3, 0x07, 0x81, 0xd3,
	0x0a, 0x85, 0x08, 0x63, 0x46, 0xa8, 0x8c, 0x08, 0xe5, 0x5c, 0x68, 0x9a, 0x55, 0x54, 0x4e, 0xf4,
	0x5a, 0xe0, 0x3c, 0xcb, 0x74, 0xf6, 0x27, 0x9c, 0x47, 0x29, 0xef, 0x1d, 0xb0, 0xc3, 0x94, 0x29,
	0xed, 0xdd, 0x82, 0xcd, 0xd2, 0xaa, 0x92, 0x82, 0x2b, 0x86, 0x37, 0xa1, 0xf9, 0x2a, 0xe5, 0xbd,
	0x97, 0x6f, 0x94, 0xe0, 0x1b, 0xe8, 0x02, 0xba, 0xd2, 0x3c, 0x38, 0x91, 0x01, 0x4f, 0x94, 0xe0,
	0xde, 0x73, 0x68, 0xcd, 0x71, 0xf7, 0x62, 0x1a, 0xf5, 0x95, 0xd5, 0xc6, 0x2d, 0x68, 0x4e, 0x1c,
	0x5b, 0xf2, 0x14, 0xc0, 0x67, 0x61, 0x4d, 0x69, 0xaa, 0x53, 0xb5, 0xb1, 0x62, 0x4a, 0xf6, 0xcb,
	0xbb, 0x07, 0xe7, 0x17, 0xa8, 0x5a, 0x4f, 0x5b, 0xb0, 0xde, 0x35, 0xc8, 0xac, 0x2b, 0xc8, 0x21,
	0xe3, 0xeb, 0x3e, 0xb4, 0x8d, 0xc2, 0xc3, 0xe3, 0x5e, 0x7b, 0xa2, 0x2f, 0x19, 0x57, 0x66, 0x2d,
	0x4b, 0x99, 0xf3, 0x1e, 0x83, 0x57, 0x25, 0x61, 0x9d, 0xb4, 0xe1, 0xa4, 0xa4, 0x43, 0x91, 0xea,
	0x82, 0x95, 0x75, 0x8b, 0x19, 0x2f, 0xee, 0x5f, 0x3b, 0xda, 0xa7, 0x09, 0x9d, 0xec, 0xa8, 0x64,
	0xda, 0xe3, 0xfa, 0x74, 0x5a, 0x69, 0x90, 0xc2, 0xb4, 0x39, 0x94, 0x75, 0xb8, 0xf6, 0x7b, 0x15,
	0x56, 0x8d, 0x04, 0xfe, 0x8c, 0xe0, 0x54, 0xf1, 0x8e, 0x78, 0xd7, 0x5f, 0x18, 0x1b, 0x7f, 0x71,
	0x2a, 0x9c, 0xeb, 0x75, 0x69, 0xb9, 0x59, 0xef, 0xf2, 0x87, 0xef, 0xbf, 0x3e, 0xad, 0xb4, 0xf1,
	0x16, 0x59, 0x1c, 0xea, 0x2c, 0x3e, 0xf8, 0x0b, 0x82, 0xd3, 0xf3, 0x07, 0xc6, 0x37, 0x96, 0xef,
	0x5a, 0x08, 0x9a, 0x73, 0xb3, 0x3e, 0xd1, 0x1a, 0xbe, 0x6a, 0x0c, 0x5f, 0xc4, 0xed, 0x0a, 0xc3,
	0x79, 0xb2, 0xf0, 0x37, 0x04, 0x67, 0x4a, 0xe3, 0x80, 0xef, 0xfc, 0xab, 0x7d, 0x55, 0x10, 0x9d,
	0xbb, 0xff, 0xc9, 0xb6, 0x13, 0xdc, 0x36, 0x13, 0xec, 0xe2, 0x9d, 0xaa, 0x09, 0x66, 0x88, 0xe4,
	0xdd, 0xe4, 0xd1, 0xfb, 0xb9, 0x33, 0xe4, 0xc9, 0xab, 0x73, 0x86, 0x42, 0x96, 0xeb, 0x9c, 0xa1,
	0x18, 0xf2, 0xa5, 0xce, 0x90, 0x47, 0xfe, 0x41, 0xf8, 0x75, 0xe4, 0xa2, 0xa3, 0x91, 0x8b, 0x7e,
	0x8e, 0x5c, 0xf4, 0x71, 0xec, 0x36, 0x8e, 0xc6, 0x6e, 0xe3, 0xc7, 0xd8, 0x6d, 0xbc, 0x78, 0x1a,
	0x46, 0xfa, 0x75, 0xda, 0xf1, 0xbb, 0xa2, 0x4f, 0x94, 0x18, 0xb0, 0x84, 0x45, 0x21, 0xdf, 0x8e,
	0x83, 0x4c, 0xf3, 0xed, 0x54, 0x75, 0x7b, 0x46, 0x56, 0x0f, 0x25, 0x53, 0x65, 0x0d, 0x65, 0xa7,
	0xb3, 0x66, 0x7e, 0x9f, 0x3b, 0x7f, 0x02, 0x00, 0x00, 0xff, 0xff, 0xb5, 0x61, 0x38, 0x7b, 0xa0,
	0x05, 0x00, 0x00,
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
	ProtectionFund(ctx context.Context, in *QueryProtectionFundRequest, opts ...grpc.CallOption) (*QueryProtectionFundResponse, error)
	ProtectionClaims(ctx context.Context, in *QueryProtectionClaimsRequest, opts ...grpc.CallOption) (*QueryProtectionClaimsResponse, error)
	DelegatorCompensation(ctx context.Context, in *QueryDelegatorCompensationRequest, opts ...grpc.CallOption) (*QueryDelegatorCompensationResponse, error)
	ProtectionParams(ctx context.Context, in *QueryProtectionParamsRequest, opts ...grpc.CallOption) (*QueryProtectionParamsResponse, error)
}

type queryClient struct {
	cc grpc1.ClientConn
}

func NewQueryClient(cc grpc1.ClientConn) QueryClient {
	return &queryClient{cc}
}

func (c *queryClient) ProtectionFund(ctx context.Context, in *QueryProtectionFundRequest, opts ...grpc.CallOption) (*QueryProtectionFundResponse, error) {
	out := new(QueryProtectionFundResponse)
	err := c.cc.Invoke(ctx, "/l1.delegatorprotection.v1.Query/ProtectionFund", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) ProtectionClaims(ctx context.Context, in *QueryProtectionClaimsRequest, opts ...grpc.CallOption) (*QueryProtectionClaimsResponse, error) {
	out := new(QueryProtectionClaimsResponse)
	err := c.cc.Invoke(ctx, "/l1.delegatorprotection.v1.Query/ProtectionClaims", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) DelegatorCompensation(ctx context.Context, in *QueryDelegatorCompensationRequest, opts ...grpc.CallOption) (*QueryDelegatorCompensationResponse, error) {
	out := new(QueryDelegatorCompensationResponse)
	err := c.cc.Invoke(ctx, "/l1.delegatorprotection.v1.Query/DelegatorCompensation", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) ProtectionParams(ctx context.Context, in *QueryProtectionParamsRequest, opts ...grpc.CallOption) (*QueryProtectionParamsResponse, error) {
	out := new(QueryProtectionParamsResponse)
	err := c.cc.Invoke(ctx, "/l1.delegatorprotection.v1.Query/ProtectionParams", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QueryServer is the server API for Query service.
type QueryServer interface {
	ProtectionFund(context.Context, *QueryProtectionFundRequest) (*QueryProtectionFundResponse, error)
	ProtectionClaims(context.Context, *QueryProtectionClaimsRequest) (*QueryProtectionClaimsResponse, error)
	DelegatorCompensation(context.Context, *QueryDelegatorCompensationRequest) (*QueryDelegatorCompensationResponse, error)
	ProtectionParams(context.Context, *QueryProtectionParamsRequest) (*QueryProtectionParamsResponse, error)
}

// UnimplementedQueryServer can be embedded to have forward compatible implementations.
type UnimplementedQueryServer struct {
}

func (*UnimplementedQueryServer) ProtectionFund(ctx context.Context, req *QueryProtectionFundRequest) (*QueryProtectionFundResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ProtectionFund not implemented")
}
func (*UnimplementedQueryServer) ProtectionClaims(ctx context.Context, req *QueryProtectionClaimsRequest) (*QueryProtectionClaimsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ProtectionClaims not implemented")
}
func (*UnimplementedQueryServer) DelegatorCompensation(ctx context.Context, req *QueryDelegatorCompensationRequest) (*QueryDelegatorCompensationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DelegatorCompensation not implemented")
}
func (*UnimplementedQueryServer) ProtectionParams(ctx context.Context, req *QueryProtectionParamsRequest) (*QueryProtectionParamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ProtectionParams not implemented")
}

func RegisterQueryServer(s grpc1.Server, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc, srv)
}

func _Query_ProtectionFund_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryProtectionFundRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).ProtectionFund(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.delegatorprotection.v1.Query/ProtectionFund",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).ProtectionFund(ctx, req.(*QueryProtectionFundRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_ProtectionClaims_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryProtectionClaimsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).ProtectionClaims(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.delegatorprotection.v1.Query/ProtectionClaims",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).ProtectionClaims(ctx, req.(*QueryProtectionClaimsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_DelegatorCompensation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryDelegatorCompensationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).DelegatorCompensation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.delegatorprotection.v1.Query/DelegatorCompensation",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).DelegatorCompensation(ctx, req.(*QueryDelegatorCompensationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_ProtectionParams_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryProtectionParamsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).ProtectionParams(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.delegatorprotection.v1.Query/ProtectionParams",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).ProtectionParams(ctx, req.(*QueryProtectionParamsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var Query_serviceDesc = _Query_serviceDesc
var _Query_serviceDesc = grpc.ServiceDesc{
	ServiceName:	"l1.delegatorprotection.v1.Query",
	HandlerType:	(*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName:	"ProtectionFund",
			Handler:	_Query_ProtectionFund_Handler,
		},
		{
			MethodName:	"ProtectionClaims",
			Handler:	_Query_ProtectionClaims_Handler,
		},
		{
			MethodName:	"DelegatorCompensation",
			Handler:	_Query_DelegatorCompensation_Handler,
		},
		{
			MethodName:	"ProtectionParams",
			Handler:	_Query_ProtectionParams_Handler,
		},
	},
	Streams:	[]grpc.StreamDesc{},
	Metadata:	"l1/delegatorprotection/v1/query.proto",
}

func (m *QueryProtectionFundRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryProtectionFundRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryProtectionFundRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *QueryProtectionFundResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryProtectionFundResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryProtectionFundResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.FundJson) > 0 {
		i -= len(m.FundJson)
		copy(dAtA[i:], m.FundJson)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.FundJson)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *QueryProtectionClaimsRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryProtectionClaimsRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryProtectionClaimsRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Status) > 0 {
		i -= len(m.Status)
		copy(dAtA[i:], m.Status)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.Status)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Delegator) > 0 {
		i -= len(m.Delegator)
		copy(dAtA[i:], m.Delegator)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.Delegator)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *QueryProtectionClaimsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryProtectionClaimsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryProtectionClaimsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.ClaimsJson) > 0 {
		i -= len(m.ClaimsJson)
		copy(dAtA[i:], m.ClaimsJson)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.ClaimsJson)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *QueryDelegatorCompensationRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryDelegatorCompensationRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryDelegatorCompensationRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Delegator) > 0 {
		i -= len(m.Delegator)
		copy(dAtA[i:], m.Delegator)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.Delegator)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *QueryDelegatorCompensationResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryDelegatorCompensationResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryDelegatorCompensationResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.PayoutsJson) > 0 {
		i -= len(m.PayoutsJson)
		copy(dAtA[i:], m.PayoutsJson)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.PayoutsJson)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *QueryProtectionParamsRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryProtectionParamsRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryProtectionParamsRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *QueryProtectionParamsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryProtectionParamsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryProtectionParamsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
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
func (m *QueryProtectionFundRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *QueryProtectionFundResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.FundJson)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *QueryProtectionClaimsRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Delegator)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	l = len(m.Status)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *QueryProtectionClaimsResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ClaimsJson)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *QueryDelegatorCompensationRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Delegator)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *QueryDelegatorCompensationResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.PayoutsJson)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *QueryProtectionParamsRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *QueryProtectionParamsResponse) Size() (n int) {
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
func (m *QueryProtectionFundRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryProtectionFundRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryProtectionFundRequest: illegal tag %d (wire type %d)", fieldNum, wire)
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
func (m *QueryProtectionFundResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryProtectionFundResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryProtectionFundResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field FundJson", wireType)
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
			m.FundJson = string(dAtA[iNdEx:postIndex])
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
func (m *QueryProtectionClaimsRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryProtectionClaimsRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryProtectionClaimsRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Delegator", wireType)
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
			m.Delegator = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
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
func (m *QueryProtectionClaimsResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryProtectionClaimsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryProtectionClaimsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ClaimsJson", wireType)
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
			m.ClaimsJson = string(dAtA[iNdEx:postIndex])
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
func (m *QueryDelegatorCompensationRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryDelegatorCompensationRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryDelegatorCompensationRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Delegator", wireType)
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
			m.Delegator = string(dAtA[iNdEx:postIndex])
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
func (m *QueryDelegatorCompensationResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryDelegatorCompensationResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryDelegatorCompensationResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PayoutsJson", wireType)
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
			m.PayoutsJson = string(dAtA[iNdEx:postIndex])
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
func (m *QueryProtectionParamsRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryProtectionParamsRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryProtectionParamsRequest: illegal tag %d (wire type %d)", fieldNum, wire)
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
func (m *QueryProtectionParamsResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryProtectionParamsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryProtectionParamsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
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
