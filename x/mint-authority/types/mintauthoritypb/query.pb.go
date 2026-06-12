package mintauthoritypb

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

type QueryMintAuthorityRequest struct {
}

func (m *QueryMintAuthorityRequest) Reset()		{ *m = QueryMintAuthorityRequest{} }
func (m *QueryMintAuthorityRequest) String() string	{ return proto.CompactTextString(m) }
func (*QueryMintAuthorityRequest) ProtoMessage()	{}
func (*QueryMintAuthorityRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_0b729b740665abd4, []int{0}
}
func (m *QueryMintAuthorityRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryMintAuthorityRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryMintAuthorityRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryMintAuthorityRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryMintAuthorityRequest.Merge(m, src)
}
func (m *QueryMintAuthorityRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryMintAuthorityRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryMintAuthorityRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryMintAuthorityRequest proto.InternalMessageInfo

type QueryMintAuthorityResponse struct {
	AuthorityJson string `protobuf:"bytes,1,opt,name=authority_json,json=authorityJson,proto3" json:"authority_json,omitempty"`
}

func (m *QueryMintAuthorityResponse) Reset()		{ *m = QueryMintAuthorityResponse{} }
func (m *QueryMintAuthorityResponse) String() string	{ return proto.CompactTextString(m) }
func (*QueryMintAuthorityResponse) ProtoMessage()	{}
func (*QueryMintAuthorityResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_0b729b740665abd4, []int{1}
}
func (m *QueryMintAuthorityResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryMintAuthorityResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryMintAuthorityResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryMintAuthorityResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryMintAuthorityResponse.Merge(m, src)
}
func (m *QueryMintAuthorityResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryMintAuthorityResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryMintAuthorityResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryMintAuthorityResponse proto.InternalMessageInfo

func (m *QueryMintAuthorityResponse) GetAuthorityJson() string {
	if m != nil {
		return m.AuthorityJson
	}
	return ""
}

type QueryMintedByEpochRequest struct {
	Epoch	uint64	`protobuf:"varint,1,opt,name=epoch,proto3" json:"epoch,omitempty"`
	Denom	string	`protobuf:"bytes,2,opt,name=denom,proto3" json:"denom,omitempty"`
}

func (m *QueryMintedByEpochRequest) Reset()		{ *m = QueryMintedByEpochRequest{} }
func (m *QueryMintedByEpochRequest) String() string	{ return proto.CompactTextString(m) }
func (*QueryMintedByEpochRequest) ProtoMessage()	{}
func (*QueryMintedByEpochRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_0b729b740665abd4, []int{2}
}
func (m *QueryMintedByEpochRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryMintedByEpochRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryMintedByEpochRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryMintedByEpochRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryMintedByEpochRequest.Merge(m, src)
}
func (m *QueryMintedByEpochRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryMintedByEpochRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryMintedByEpochRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryMintedByEpochRequest proto.InternalMessageInfo

func (m *QueryMintedByEpochRequest) GetEpoch() uint64 {
	if m != nil {
		return m.Epoch
	}
	return 0
}

func (m *QueryMintedByEpochRequest) GetDenom() string {
	if m != nil {
		return m.Denom
	}
	return ""
}

type QueryMintedByEpochResponse struct {
	MintedJson string `protobuf:"bytes,1,opt,name=minted_json,json=mintedJson,proto3" json:"minted_json,omitempty"`
}

func (m *QueryMintedByEpochResponse) Reset()		{ *m = QueryMintedByEpochResponse{} }
func (m *QueryMintedByEpochResponse) String() string	{ return proto.CompactTextString(m) }
func (*QueryMintedByEpochResponse) ProtoMessage()	{}
func (*QueryMintedByEpochResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_0b729b740665abd4, []int{3}
}
func (m *QueryMintedByEpochResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryMintedByEpochResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryMintedByEpochResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryMintedByEpochResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryMintedByEpochResponse.Merge(m, src)
}
func (m *QueryMintedByEpochResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryMintedByEpochResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryMintedByEpochResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryMintedByEpochResponse proto.InternalMessageInfo

func (m *QueryMintedByEpochResponse) GetMintedJson() string {
	if m != nil {
		return m.MintedJson
	}
	return ""
}

type QueryMintedLifetimeRequest struct {
	Denom string `protobuf:"bytes,1,opt,name=denom,proto3" json:"denom,omitempty"`
}

func (m *QueryMintedLifetimeRequest) Reset()		{ *m = QueryMintedLifetimeRequest{} }
func (m *QueryMintedLifetimeRequest) String() string	{ return proto.CompactTextString(m) }
func (*QueryMintedLifetimeRequest) ProtoMessage()	{}
func (*QueryMintedLifetimeRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_0b729b740665abd4, []int{4}
}
func (m *QueryMintedLifetimeRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryMintedLifetimeRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryMintedLifetimeRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryMintedLifetimeRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryMintedLifetimeRequest.Merge(m, src)
}
func (m *QueryMintedLifetimeRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryMintedLifetimeRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryMintedLifetimeRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryMintedLifetimeRequest proto.InternalMessageInfo

func (m *QueryMintedLifetimeRequest) GetDenom() string {
	if m != nil {
		return m.Denom
	}
	return ""
}

type QueryMintedLifetimeResponse struct {
	MintedJson string `protobuf:"bytes,1,opt,name=minted_json,json=mintedJson,proto3" json:"minted_json,omitempty"`
}

func (m *QueryMintedLifetimeResponse) Reset()		{ *m = QueryMintedLifetimeResponse{} }
func (m *QueryMintedLifetimeResponse) String() string	{ return proto.CompactTextString(m) }
func (*QueryMintedLifetimeResponse) ProtoMessage()	{}
func (*QueryMintedLifetimeResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_0b729b740665abd4, []int{5}
}
func (m *QueryMintedLifetimeResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryMintedLifetimeResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryMintedLifetimeResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryMintedLifetimeResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryMintedLifetimeResponse.Merge(m, src)
}
func (m *QueryMintedLifetimeResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryMintedLifetimeResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryMintedLifetimeResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryMintedLifetimeResponse proto.InternalMessageInfo

func (m *QueryMintedLifetimeResponse) GetMintedJson() string {
	if m != nil {
		return m.MintedJson
	}
	return ""
}

type QueryMintCapsRequest struct {
}

func (m *QueryMintCapsRequest) Reset()		{ *m = QueryMintCapsRequest{} }
func (m *QueryMintCapsRequest) String() string	{ return proto.CompactTextString(m) }
func (*QueryMintCapsRequest) ProtoMessage()	{}
func (*QueryMintCapsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_0b729b740665abd4, []int{6}
}
func (m *QueryMintCapsRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryMintCapsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryMintCapsRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryMintCapsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryMintCapsRequest.Merge(m, src)
}
func (m *QueryMintCapsRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryMintCapsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryMintCapsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryMintCapsRequest proto.InternalMessageInfo

type QueryMintCapsResponse struct {
	CapsJson string `protobuf:"bytes,1,opt,name=caps_json,json=capsJson,proto3" json:"caps_json,omitempty"`
}

func (m *QueryMintCapsResponse) Reset()		{ *m = QueryMintCapsResponse{} }
func (m *QueryMintCapsResponse) String() string	{ return proto.CompactTextString(m) }
func (*QueryMintCapsResponse) ProtoMessage()	{}
func (*QueryMintCapsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_0b729b740665abd4, []int{7}
}
func (m *QueryMintCapsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryMintCapsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryMintCapsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryMintCapsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryMintCapsResponse.Merge(m, src)
}
func (m *QueryMintCapsResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryMintCapsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryMintCapsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryMintCapsResponse proto.InternalMessageInfo

func (m *QueryMintCapsResponse) GetCapsJson() string {
	if m != nil {
		return m.CapsJson
	}
	return ""
}

func init() {
	proto.RegisterType((*QueryMintAuthorityRequest)(nil), "l1.mintauthority.v1.QueryMintAuthorityRequest")
	proto.RegisterType((*QueryMintAuthorityResponse)(nil), "l1.mintauthority.v1.QueryMintAuthorityResponse")
	proto.RegisterType((*QueryMintedByEpochRequest)(nil), "l1.mintauthority.v1.QueryMintedByEpochRequest")
	proto.RegisterType((*QueryMintedByEpochResponse)(nil), "l1.mintauthority.v1.QueryMintedByEpochResponse")
	proto.RegisterType((*QueryMintedLifetimeRequest)(nil), "l1.mintauthority.v1.QueryMintedLifetimeRequest")
	proto.RegisterType((*QueryMintedLifetimeResponse)(nil), "l1.mintauthority.v1.QueryMintedLifetimeResponse")
	proto.RegisterType((*QueryMintCapsRequest)(nil), "l1.mintauthority.v1.QueryMintCapsRequest")
	proto.RegisterType((*QueryMintCapsResponse)(nil), "l1.mintauthority.v1.QueryMintCapsResponse")
}

func init()	{ proto.RegisterFile("l1/mintauthority/v1/query.proto", fileDescriptor_0b729b740665abd4) }

var fileDescriptor_0b729b740665abd4 = []byte{

	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x94, 0x41, 0x6b, 0x13, 0x41,
	0x1c, 0xc5, 0x33, 0xc5, 0x48, 0x3b, 0xd2, 0x1e, 0xc6, 0x2a, 0xed, 0x46, 0xb6, 0x75, 0x41, 0x51,
	0xb1, 0x33, 0x6e, 0x0d, 0x9e, 0x54, 0x68, 0x8b, 0x08, 0xa2, 0x07, 0x73, 0xd4, 0x43, 0xd9, 0xa4,
	0x63, 0x32, 0xb2, 0x3b, 0x33, 0xcd, 0x4c, 0x82, 0x4b, 0xe9, 0x45, 0xc4, 0xb3, 0xe0, 0xc5, 0xaf,
	0xe0, 0xc5, 0xcf, 0xe1, 0xb1, 0xe0, 0xc5, 0xa3, 0x24, 0x7e, 0x10, 0xd9, 0xd9, 0xc9, 0x3a, 0x1b,
	0xd7, 0x66, 0x4f, 0x61, 0xfe, 0x33, 0xef, 0xbd, 0x1f, 0x79, 0x7f, 0x16, 0x6e, 0xc5, 0x21, 0x49,
	0x18, 0xd7, 0xd1, 0x48, 0x0f, 0xc4, 0x90, 0xe9, 0x94, 0x8c, 0x43, 0x72, 0x3c, 0xa2, 0xc3, 0x14,
	0xcb, 0xa1, 0xd0, 0x02, 0x5d, 0x8e, 0x43, 0x5c, 0x7a, 0x80, 0xc7, 0xa1, 0x77, 0xad, 0x2f, 0x44,
	0x3f, 0xa6, 0x24, 0x92, 0x8c, 0x44, 0x9c, 0x0b, 0x1d, 0x69, 0x26, 0xb8, 0xca, 0x25, 0x41, 0x0b,
	0x6e, 0xbe, 0xcc, 0x1c, 0x5e, 0x30, 0xae, 0xf7, 0x66, 0xb2, 0x0e, 0x3d, 0x1e, 0x51, 0xa5, 0x83,
	0x03, 0xe8, 0x55, 0x5d, 0x2a, 0x29, 0xb8, 0xa2, 0xe8, 0x06, 0x5c, 0x2b, 0x82, 0x0e, 0xdf, 0x2a,
	0xc1, 0x37, 0xc0, 0x36, 0xb8, 0xb5, 0xd2, 0x59, 0x2d, 0xa6, 0xcf, 0x94, 0xe0, 0xc1, 0x53, 0x27,
	0x81, 0x1e, 0xed, 0xa7, 0x4f, 0xa4, 0xe8, 0x0d, 0x6c, 0x02, 0x5a, 0x87, 0x4d, 0x9a, 0x9d, 0x8d,
	0xf4, 0x42, 0x27, 0x3f, 0x64, 0xd3, 0x23, 0xca, 0x45, 0xb2, 0xb1, 0x64, 0x0c, 0xf3, 0x43, 0xf0,
	0xc8, 0xa1, 0x71, 0x8c, 0x2c, 0xcd, 0x16, 0xbc, 0x94, 0x98, 0x0b, 0x17, 0x05, 0xe6, 0x23, 0xc3,
	0xb1, 0x5b, 0x92, 0x3f, 0x67, 0x6f, 0xa8, 0x66, 0x09, 0x75, 0x40, 0xf2, 0x48, 0xe0, 0x46, 0x3e,
	0x86, 0xad, 0x4a, 0x4d, 0xdd, 0xcc, 0xab, 0x70, 0xbd, 0xd0, 0x1f, 0x44, 0x52, 0xcd, 0xfe, 0xd8,
	0x36, 0xbc, 0x32, 0x37, 0xb7, 0x8e, 0x2d, 0xb8, 0xd2, 0x8b, 0xa4, 0x72, 0xfd, 0x96, 0xb3, 0x41,
	0xe6, 0xb6, 0xfb, 0xb1, 0x09, 0x9b, 0x46, 0x86, 0xbe, 0x00, 0xb8, 0x5a, 0x2a, 0x05, 0x61, 0x5c,
	0xd1, 0x3d, 0xfe, 0x6f, 0xb5, 0x1e, 0xa9, 0xfd, 0x3e, 0x27, 0x0b, 0x6e, 0xbe, 0xff, 0xf1, 0xfb,
	0xf3, 0xd2, 0x36, 0xf2, 0x49, 0xd5, 0x16, 0x16, 0x07, 0xf4, 0xcd, 0xa2, 0x15, 0x0d, 0x2d, 0x42,
	0x9b, 0xdf, 0x89, 0x45, 0x68, 0xff, 0x54, 0x1f, 0x3c, 0x34, 0x68, 0x0f, 0x50, 0xbb, 0x12, 0xcd,
	0x36, 0xd4, 0x4d, 0x0f, 0xcd, 0x72, 0x91, 0x13, 0xf3, 0x73, 0x4a, 0x4e, 0x4c, 0xc5, 0xa7, 0xe8,
	0x2b, 0x80, 0x6b, 0xe5, 0x7e, 0xd1, 0x42, 0x82, 0xb9, 0xed, 0xf1, 0xee, 0xd5, 0x17, 0x58, 0xe6,
	0xb6, 0x61, 0xc6, 0xe8, 0xee, 0x79, 0xcc, 0xb1, 0x55, 0x15, 0xac, 0x1f, 0x00, 0x5c, 0x9e, 0xed,
	0x0c, 0xba, 0x7d, 0x7e, 0xa8, 0xb3, 0x6f, 0xde, 0x9d, 0x3a, 0x4f, 0x2d, 0xd9, 0x75, 0x43, 0xd6,
	0x42, 0x9b, 0x95, 0x64, 0xd9, 0x32, 0xee, 0xbf, 0xfe, 0x3e, 0xf1, 0xc1, 0xd9, 0xc4, 0x07, 0xbf,
	0x26, 0x3e, 0xf8, 0x34, 0xf5, 0x1b, 0x67, 0x53, 0xbf, 0xf1, 0x73, 0xea, 0x37, 0x5e, 0xed, 0xf5,
	0x99, 0x1e, 0x8c, 0xba, 0xb8, 0x27, 0x12, 0xa2, 0xc4, 0x98, 0x0e, 0x29, 0xeb, 0xf3, 0x9d, 0x38,
	0xcc, 0xbc, 0xde, 0x19, 0xb7, 0x9d, 0xbf, 0x76, 0x3a, 0x95, 0x54, 0x95, 0x23, 0x64, 0xb7, 0x7b,
	0xd1, 0x7c, 0x98, 0xee, 0xff, 0x09, 0x00, 0x00, 0xff, 0xff, 0x8c, 0xb4, 0x60, 0xc4, 0xee, 0x04,
	0x00, 0x00,
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
	MintAuthority(ctx context.Context, in *QueryMintAuthorityRequest, opts ...grpc.CallOption) (*QueryMintAuthorityResponse, error)
	MintedByEpoch(ctx context.Context, in *QueryMintedByEpochRequest, opts ...grpc.CallOption) (*QueryMintedByEpochResponse, error)
	MintedLifetime(ctx context.Context, in *QueryMintedLifetimeRequest, opts ...grpc.CallOption) (*QueryMintedLifetimeResponse, error)
	MintCaps(ctx context.Context, in *QueryMintCapsRequest, opts ...grpc.CallOption) (*QueryMintCapsResponse, error)
}

type queryClient struct {
	cc grpc1.ClientConn
}

func NewQueryClient(cc grpc1.ClientConn) QueryClient {
	return &queryClient{cc}
}

func (c *queryClient) MintAuthority(ctx context.Context, in *QueryMintAuthorityRequest, opts ...grpc.CallOption) (*QueryMintAuthorityResponse, error) {
	out := new(QueryMintAuthorityResponse)
	err := c.cc.Invoke(ctx, "/l1.mintauthority.v1.Query/MintAuthority", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) MintedByEpoch(ctx context.Context, in *QueryMintedByEpochRequest, opts ...grpc.CallOption) (*QueryMintedByEpochResponse, error) {
	out := new(QueryMintedByEpochResponse)
	err := c.cc.Invoke(ctx, "/l1.mintauthority.v1.Query/MintedByEpoch", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) MintedLifetime(ctx context.Context, in *QueryMintedLifetimeRequest, opts ...grpc.CallOption) (*QueryMintedLifetimeResponse, error) {
	out := new(QueryMintedLifetimeResponse)
	err := c.cc.Invoke(ctx, "/l1.mintauthority.v1.Query/MintedLifetime", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) MintCaps(ctx context.Context, in *QueryMintCapsRequest, opts ...grpc.CallOption) (*QueryMintCapsResponse, error) {
	out := new(QueryMintCapsResponse)
	err := c.cc.Invoke(ctx, "/l1.mintauthority.v1.Query/MintCaps", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QueryServer is the server API for Query service.
type QueryServer interface {
	MintAuthority(context.Context, *QueryMintAuthorityRequest) (*QueryMintAuthorityResponse, error)
	MintedByEpoch(context.Context, *QueryMintedByEpochRequest) (*QueryMintedByEpochResponse, error)
	MintedLifetime(context.Context, *QueryMintedLifetimeRequest) (*QueryMintedLifetimeResponse, error)
	MintCaps(context.Context, *QueryMintCapsRequest) (*QueryMintCapsResponse, error)
}

// UnimplementedQueryServer can be embedded to have forward compatible implementations.
type UnimplementedQueryServer struct {
}

func (*UnimplementedQueryServer) MintAuthority(ctx context.Context, req *QueryMintAuthorityRequest) (*QueryMintAuthorityResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MintAuthority not implemented")
}
func (*UnimplementedQueryServer) MintedByEpoch(ctx context.Context, req *QueryMintedByEpochRequest) (*QueryMintedByEpochResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MintedByEpoch not implemented")
}
func (*UnimplementedQueryServer) MintedLifetime(ctx context.Context, req *QueryMintedLifetimeRequest) (*QueryMintedLifetimeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MintedLifetime not implemented")
}
func (*UnimplementedQueryServer) MintCaps(ctx context.Context, req *QueryMintCapsRequest) (*QueryMintCapsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MintCaps not implemented")
}

func RegisterQueryServer(s grpc1.Server, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc, srv)
}

func _Query_MintAuthority_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryMintAuthorityRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).MintAuthority(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.mintauthority.v1.Query/MintAuthority",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).MintAuthority(ctx, req.(*QueryMintAuthorityRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_MintedByEpoch_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryMintedByEpochRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).MintedByEpoch(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.mintauthority.v1.Query/MintedByEpoch",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).MintedByEpoch(ctx, req.(*QueryMintedByEpochRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_MintedLifetime_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryMintedLifetimeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).MintedLifetime(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.mintauthority.v1.Query/MintedLifetime",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).MintedLifetime(ctx, req.(*QueryMintedLifetimeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_MintCaps_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryMintCapsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).MintCaps(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.mintauthority.v1.Query/MintCaps",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).MintCaps(ctx, req.(*QueryMintCapsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var Query_serviceDesc = _Query_serviceDesc
var _Query_serviceDesc = grpc.ServiceDesc{
	ServiceName:	"l1.mintauthority.v1.Query",
	HandlerType:	(*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName:	"MintAuthority",
			Handler:	_Query_MintAuthority_Handler,
		},
		{
			MethodName:	"MintedByEpoch",
			Handler:	_Query_MintedByEpoch_Handler,
		},
		{
			MethodName:	"MintedLifetime",
			Handler:	_Query_MintedLifetime_Handler,
		},
		{
			MethodName:	"MintCaps",
			Handler:	_Query_MintCaps_Handler,
		},
	},
	Streams:	[]grpc.StreamDesc{},
	Metadata:	"l1/mintauthority/v1/query.proto",
}

func (m *QueryMintAuthorityRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryMintAuthorityRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryMintAuthorityRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *QueryMintAuthorityResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryMintAuthorityResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryMintAuthorityResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.AuthorityJson) > 0 {
		i -= len(m.AuthorityJson)
		copy(dAtA[i:], m.AuthorityJson)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.AuthorityJson)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *QueryMintedByEpochRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryMintedByEpochRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryMintedByEpochRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Denom) > 0 {
		i -= len(m.Denom)
		copy(dAtA[i:], m.Denom)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.Denom)))
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

func (m *QueryMintedByEpochResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryMintedByEpochResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryMintedByEpochResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.MintedJson) > 0 {
		i -= len(m.MintedJson)
		copy(dAtA[i:], m.MintedJson)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.MintedJson)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *QueryMintedLifetimeRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryMintedLifetimeRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryMintedLifetimeRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Denom) > 0 {
		i -= len(m.Denom)
		copy(dAtA[i:], m.Denom)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.Denom)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *QueryMintedLifetimeResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryMintedLifetimeResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryMintedLifetimeResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.MintedJson) > 0 {
		i -= len(m.MintedJson)
		copy(dAtA[i:], m.MintedJson)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.MintedJson)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *QueryMintCapsRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryMintCapsRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryMintCapsRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *QueryMintCapsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryMintCapsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryMintCapsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.CapsJson) > 0 {
		i -= len(m.CapsJson)
		copy(dAtA[i:], m.CapsJson)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.CapsJson)))
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
func (m *QueryMintAuthorityRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *QueryMintAuthorityResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.AuthorityJson)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *QueryMintedByEpochRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Epoch != 0 {
		n += 1 + sovQuery(uint64(m.Epoch))
	}
	l = len(m.Denom)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *QueryMintedByEpochResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.MintedJson)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *QueryMintedLifetimeRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Denom)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *QueryMintedLifetimeResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.MintedJson)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *QueryMintCapsRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *QueryMintCapsResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.CapsJson)
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
func (m *QueryMintAuthorityRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryMintAuthorityRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryMintAuthorityRequest: illegal tag %d (wire type %d)", fieldNum, wire)
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
func (m *QueryMintAuthorityResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryMintAuthorityResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryMintAuthorityResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AuthorityJson", wireType)
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
			m.AuthorityJson = string(dAtA[iNdEx:postIndex])
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
func (m *QueryMintedByEpochRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryMintedByEpochRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryMintedByEpochRequest: illegal tag %d (wire type %d)", fieldNum, wire)
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
				return fmt.Errorf("proto: wrong wireType = %d for field Denom", wireType)
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
			m.Denom = string(dAtA[iNdEx:postIndex])
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
func (m *QueryMintedByEpochResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryMintedByEpochResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryMintedByEpochResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MintedJson", wireType)
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
			m.MintedJson = string(dAtA[iNdEx:postIndex])
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
func (m *QueryMintedLifetimeRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryMintedLifetimeRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryMintedLifetimeRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Denom", wireType)
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
			m.Denom = string(dAtA[iNdEx:postIndex])
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
func (m *QueryMintedLifetimeResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryMintedLifetimeResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryMintedLifetimeResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MintedJson", wireType)
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
			m.MintedJson = string(dAtA[iNdEx:postIndex])
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
func (m *QueryMintCapsRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryMintCapsRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryMintCapsRequest: illegal tag %d (wire type %d)", fieldNum, wire)
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
func (m *QueryMintCapsResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryMintCapsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryMintCapsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field CapsJson", wireType)
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
			m.CapsJson = string(dAtA[iNdEx:postIndex])
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
