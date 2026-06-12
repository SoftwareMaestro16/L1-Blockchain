package reputationpb

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

type QueryValidatorReputationRequest struct {
	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
}

func (m *QueryValidatorReputationRequest) Reset()		{ *m = QueryValidatorReputationRequest{} }
func (m *QueryValidatorReputationRequest) String() string	{ return proto.CompactTextString(m) }
func (*QueryValidatorReputationRequest) ProtoMessage()		{}
func (*QueryValidatorReputationRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_20d8edb7a8ff18f4, []int{0}
}
func (m *QueryValidatorReputationRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryValidatorReputationRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryValidatorReputationRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryValidatorReputationRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryValidatorReputationRequest.Merge(m, src)
}
func (m *QueryValidatorReputationRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryValidatorReputationRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryValidatorReputationRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryValidatorReputationRequest proto.InternalMessageInfo

func (m *QueryValidatorReputationRequest) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

type QueryReporterReputationRequest struct {
	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
}

func (m *QueryReporterReputationRequest) Reset()		{ *m = QueryReporterReputationRequest{} }
func (m *QueryReporterReputationRequest) String() string	{ return proto.CompactTextString(m) }
func (*QueryReporterReputationRequest) ProtoMessage()		{}
func (*QueryReporterReputationRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_20d8edb7a8ff18f4, []int{1}
}
func (m *QueryReporterReputationRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryReporterReputationRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryReporterReputationRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryReporterReputationRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryReporterReputationRequest.Merge(m, src)
}
func (m *QueryReporterReputationRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryReporterReputationRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryReporterReputationRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryReporterReputationRequest proto.InternalMessageInfo

func (m *QueryReporterReputationRequest) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

type QueryValidatorReputationResponse struct {
	RecordJson string `protobuf:"bytes,1,opt,name=record_json,json=recordJson,proto3" json:"record_json,omitempty"`
}

func (m *QueryValidatorReputationResponse) Reset()		{ *m = QueryValidatorReputationResponse{} }
func (m *QueryValidatorReputationResponse) String() string	{ return proto.CompactTextString(m) }
func (*QueryValidatorReputationResponse) ProtoMessage()		{}
func (*QueryValidatorReputationResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_20d8edb7a8ff18f4, []int{2}
}
func (m *QueryValidatorReputationResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryValidatorReputationResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryValidatorReputationResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryValidatorReputationResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryValidatorReputationResponse.Merge(m, src)
}
func (m *QueryValidatorReputationResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryValidatorReputationResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryValidatorReputationResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryValidatorReputationResponse proto.InternalMessageInfo

func (m *QueryValidatorReputationResponse) GetRecordJson() string {
	if m != nil {
		return m.RecordJson
	}
	return ""
}

type QueryReporterReputationResponse struct {
	RecordJson string `protobuf:"bytes,1,opt,name=record_json,json=recordJson,proto3" json:"record_json,omitempty"`
}

func (m *QueryReporterReputationResponse) Reset()		{ *m = QueryReporterReputationResponse{} }
func (m *QueryReporterReputationResponse) String() string	{ return proto.CompactTextString(m) }
func (*QueryReporterReputationResponse) ProtoMessage()		{}
func (*QueryReporterReputationResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_20d8edb7a8ff18f4, []int{3}
}
func (m *QueryReporterReputationResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryReporterReputationResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryReporterReputationResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryReporterReputationResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryReporterReputationResponse.Merge(m, src)
}
func (m *QueryReporterReputationResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryReporterReputationResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryReporterReputationResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryReporterReputationResponse proto.InternalMessageInfo

func (m *QueryReporterReputationResponse) GetRecordJson() string {
	if m != nil {
		return m.RecordJson
	}
	return ""
}

type QueryReputationHistoryRequest struct {
	SubjectType	string	`protobuf:"bytes,1,opt,name=subject_type,json=subjectType,proto3" json:"subject_type,omitempty"`
	Address		string	`protobuf:"bytes,2,opt,name=address,proto3" json:"address,omitempty"`
	Limit		uint32	`protobuf:"varint,3,opt,name=limit,proto3" json:"limit,omitempty"`
}

func (m *QueryReputationHistoryRequest) Reset()		{ *m = QueryReputationHistoryRequest{} }
func (m *QueryReputationHistoryRequest) String() string	{ return proto.CompactTextString(m) }
func (*QueryReputationHistoryRequest) ProtoMessage()	{}
func (*QueryReputationHistoryRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_20d8edb7a8ff18f4, []int{4}
}
func (m *QueryReputationHistoryRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryReputationHistoryRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryReputationHistoryRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryReputationHistoryRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryReputationHistoryRequest.Merge(m, src)
}
func (m *QueryReputationHistoryRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryReputationHistoryRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryReputationHistoryRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryReputationHistoryRequest proto.InternalMessageInfo

func (m *QueryReputationHistoryRequest) GetSubjectType() string {
	if m != nil {
		return m.SubjectType
	}
	return ""
}

func (m *QueryReputationHistoryRequest) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func (m *QueryReputationHistoryRequest) GetLimit() uint32 {
	if m != nil {
		return m.Limit
	}
	return 0
}

type QueryReputationHistoryResponse struct {
	SnapshotsJson	string	`protobuf:"bytes,1,opt,name=snapshots_json,json=snapshotsJson,proto3" json:"snapshots_json,omitempty"`
	EventsJson	string	`protobuf:"bytes,2,opt,name=events_json,json=eventsJson,proto3" json:"events_json,omitempty"`
}

func (m *QueryReputationHistoryResponse) Reset()		{ *m = QueryReputationHistoryResponse{} }
func (m *QueryReputationHistoryResponse) String() string	{ return proto.CompactTextString(m) }
func (*QueryReputationHistoryResponse) ProtoMessage()		{}
func (*QueryReputationHistoryResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_20d8edb7a8ff18f4, []int{5}
}
func (m *QueryReputationHistoryResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryReputationHistoryResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryReputationHistoryResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryReputationHistoryResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryReputationHistoryResponse.Merge(m, src)
}
func (m *QueryReputationHistoryResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryReputationHistoryResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryReputationHistoryResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryReputationHistoryResponse proto.InternalMessageInfo

func (m *QueryReputationHistoryResponse) GetSnapshotsJson() string {
	if m != nil {
		return m.SnapshotsJson
	}
	return ""
}

func (m *QueryReputationHistoryResponse) GetEventsJson() string {
	if m != nil {
		return m.EventsJson
	}
	return ""
}

type QueryReputationParamsRequest struct {
}

func (m *QueryReputationParamsRequest) Reset()		{ *m = QueryReputationParamsRequest{} }
func (m *QueryReputationParamsRequest) String() string	{ return proto.CompactTextString(m) }
func (*QueryReputationParamsRequest) ProtoMessage()	{}
func (*QueryReputationParamsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_20d8edb7a8ff18f4, []int{6}
}
func (m *QueryReputationParamsRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryReputationParamsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryReputationParamsRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryReputationParamsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryReputationParamsRequest.Merge(m, src)
}
func (m *QueryReputationParamsRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryReputationParamsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryReputationParamsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryReputationParamsRequest proto.InternalMessageInfo

type QueryReputationParamsResponse struct {
	ParamsJson string `protobuf:"bytes,1,opt,name=params_json,json=paramsJson,proto3" json:"params_json,omitempty"`
}

func (m *QueryReputationParamsResponse) Reset()		{ *m = QueryReputationParamsResponse{} }
func (m *QueryReputationParamsResponse) String() string	{ return proto.CompactTextString(m) }
func (*QueryReputationParamsResponse) ProtoMessage()	{}
func (*QueryReputationParamsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_20d8edb7a8ff18f4, []int{7}
}
func (m *QueryReputationParamsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryReputationParamsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryReputationParamsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryReputationParamsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryReputationParamsResponse.Merge(m, src)
}
func (m *QueryReputationParamsResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryReputationParamsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryReputationParamsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryReputationParamsResponse proto.InternalMessageInfo

func (m *QueryReputationParamsResponse) GetParamsJson() string {
	if m != nil {
		return m.ParamsJson
	}
	return ""
}

func init() {
	proto.RegisterType((*QueryValidatorReputationRequest)(nil), "l1.reputation.v1.QueryValidatorReputationRequest")
	proto.RegisterType((*QueryReporterReputationRequest)(nil), "l1.reputation.v1.QueryReporterReputationRequest")
	proto.RegisterType((*QueryValidatorReputationResponse)(nil), "l1.reputation.v1.QueryValidatorReputationResponse")
	proto.RegisterType((*QueryReporterReputationResponse)(nil), "l1.reputation.v1.QueryReporterReputationResponse")
	proto.RegisterType((*QueryReputationHistoryRequest)(nil), "l1.reputation.v1.QueryReputationHistoryRequest")
	proto.RegisterType((*QueryReputationHistoryResponse)(nil), "l1.reputation.v1.QueryReputationHistoryResponse")
	proto.RegisterType((*QueryReputationParamsRequest)(nil), "l1.reputation.v1.QueryReputationParamsRequest")
	proto.RegisterType((*QueryReputationParamsResponse)(nil), "l1.reputation.v1.QueryReputationParamsResponse")
}

func init()	{ proto.RegisterFile("l1/reputation/v1/query.proto", fileDescriptor_20d8edb7a8ff18f4) }

var fileDescriptor_20d8edb7a8ff18f4 = []byte{

	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x94, 0x31, 0x6f, 0xd3, 0x40,
	0x14, 0xc7, 0x73, 0x45, 0x01, 0xf1, 0x4a, 0x51, 0x39, 0x18, 0x22, 0x2b, 0xb8, 0xc1, 0x52, 0x21,
	0x4b, 0x7d, 0x75, 0x60, 0x0a, 0x0b, 0x2a, 0x0b, 0x62, 0x82, 0xa8, 0x62, 0x60, 0xa9, 0x9c, 0xe4,
	0x94, 0xb8, 0x72, 0x7c, 0xd7, 0xbb, 0xb3, 0x45, 0x54, 0x75, 0xe1, 0x13, 0x20, 0x21, 0xbe, 0x03,
	0x03, 0x62, 0xe0, 0x53, 0x30, 0x56, 0x62, 0x61, 0x44, 0x09, 0x1f, 0xa4, 0xb2, 0xef, 0xdc, 0x38,
	0x71, 0xa2, 0x26, 0xa3, 0xdf, 0xbd, 0xff, 0x7b, 0xbf, 0xff, 0xf9, 0x6f, 0x43, 0x3d, 0xf4, 0x88,
	0xa0, 0x3c, 0x56, 0xbe, 0x0a, 0x58, 0x44, 0x12, 0x8f, 0x9c, 0xc5, 0x54, 0x8c, 0x5d, 0x2e, 0x98,
	0x62, 0x78, 0x37, 0xf4, 0xdc, 0xd9, 0xa9, 0x9b, 0x78, 0x56, 0x7d, 0xc0, 0xd8, 0x20, 0xa4, 0xc4,
	0xe7, 0x01, 0xf1, 0xa3, 0x88, 0xe9, 0x13, 0xa9, 0xfb, 0x9d, 0x97, 0xb0, 0xf7, 0x3e, 0x95, 0x7f,
	0xf0, 0xc3, 0xa0, 0xef, 0x2b, 0x26, 0x3a, 0xd7, 0xe2, 0x0e, 0x3d, 0x8b, 0xa9, 0x54, 0xb8, 0x06,
	0x77, 0xfc, 0x7e, 0x5f, 0x50, 0x29, 0x6b, 0xa8, 0x81, 0x9a, 0x77, 0x3b, 0xf9, 0xa3, 0xd3, 0x06,
	0x3b, 0x13, 0x77, 0x28, 0x67, 0x42, 0xd1, 0x8d, 0xb4, 0xaf, 0xa1, 0xb1, 0x7a, 0xb1, 0xe4, 0x2c,
	0x92, 0x14, 0xef, 0xc1, 0xb6, 0xa0, 0x3d, 0x26, 0xfa, 0x27, 0xa7, 0x92, 0x45, 0x66, 0x02, 0xe8,
	0xd2, 0x5b, 0xc9, 0x22, 0xe7, 0xc8, 0xd0, 0x2f, 0x03, 0x58, 0x77, 0x86, 0x80, 0xc7, 0xf9, 0x0c,
	0xa3, 0x7d, 0x13, 0x48, 0xc5, 0xd2, 0x82, 0xf6, 0xf0, 0x04, 0xee, 0xc9, 0xb8, 0x7b, 0x4a, 0x7b,
	0xea, 0x44, 0x8d, 0x39, 0x35, 0x23, 0xb6, 0x4d, 0xed, 0x78, 0xcc, 0x69, 0xd1, 0xe6, 0xd6, 0x9c,
	0x4d, 0xfc, 0x08, 0xaa, 0x61, 0x30, 0x0a, 0x54, 0xed, 0x56, 0x03, 0x35, 0x77, 0x3a, 0xfa, 0xc1,
	0x19, 0xce, 0x2e, 0x6e, 0x71, 0xa7, 0xc1, 0xde, 0x87, 0xfb, 0x32, 0xf2, 0xb9, 0x1c, 0x32, 0x25,
	0x8b, 0xe4, 0x3b, 0xd7, 0xd5, 0x14, 0x3e, 0x75, 0x47, 0x13, 0x1a, 0xe5, 0x3d, 0x7a, 0x39, 0xe8,
	0x52, 0xe6, 0xce, 0x86, 0xfa, 0xc2, 0xa6, 0x77, 0xbe, 0xf0, 0x47, 0xd2, 0x98, 0x73, 0x5e, 0x95,
	0xdc, 0xe7, 0xe7, 0xb3, 0xfb, 0xe3, 0x59, 0x65, 0xee, 0xfe, 0x74, 0x29, 0xdd, 0xd0, 0xfa, 0x59,
	0x85, 0x6a, 0x36, 0x02, 0xff, 0x40, 0xf0, 0x70, 0xc9, 0xeb, 0xc4, 0x9e, 0xbb, 0x18, 0x4a, 0xf7,
	0x86, 0xcc, 0x59, 0xad, 0x4d, 0x24, 0x9a, 0xd4, 0x71, 0x3f, 0xff, 0xf9, 0xff, 0x75, 0xab, 0x89,
	0x9f, 0x92, 0xd2, 0x17, 0x92, 0xe4, 0x32, 0x49, 0xce, 0xcd, 0x9b, 0xb9, 0xc0, 0xdf, 0x11, 0xe0,
	0x72, 0x70, 0xf0, 0xe1, 0x8a, 0xd5, 0x2b, 0x43, 0x6e, 0x79, 0x1b, 0x28, 0x0c, 0xeb, 0x41, 0xc6,
	0xfa, 0x0c, 0xef, 0x97, 0x59, 0x85, 0x51, 0x15, 0x51, 0x7f, 0x21, 0x78, 0x50, 0xca, 0x0a, 0x26,
	0xab, 0xf7, 0x2e, 0x4d, 0xb2, 0x75, 0xb8, 0xbe, 0xc0, 0x70, 0xb6, 0x33, 0xce, 0x17, 0xb8, 0x55,
	0xe6, 0x1c, 0xea, 0x56, 0x72, 0x5e, 0xfc, 0x38, 0x2e, 0x0a, 0xd0, 0xdf, 0x10, 0xec, 0x2e, 0xc6,
	0x0a, 0xbb, 0x37, 0x22, 0xcc, 0xe5, 0xd3, 0x22, 0x6b, 0xf7, 0x1b, 0xe2, 0x46, 0x46, 0x6c, 0xe1,
	0x5a, 0x99, 0x58, 0x87, 0xf6, 0xe8, 0xf8, 0xf7, 0xc4, 0x46, 0x97, 0x13, 0x1b, 0xfd, 0x9b, 0xd8,
	0xe8, 0xcb, 0xd4, 0xae, 0x5c, 0x4e, 0xed, 0xca, 0xdf, 0xa9, 0x5d, 0xf9, 0xd8, 0x1e, 0x04, 0x6a,
	0x18, 0x77, 0xdd, 0x1e, 0x1b, 0x11, 0xc9, 0x12, 0x2a, 0x68, 0x30, 0x88, 0x0e, 0x42, 0x2f, 0x1d,
	0xf5, 0xa9, 0x38, 0x2c, 0xb5, 0x2a, 0x0b, 0x05, 0xde, 0xed, 0xde, 0xce, 0xfe, 0xa7, 0xcf, 0xaf,
	0x02, 0x00, 0x00, 0xff, 0xff, 0x52, 0x22, 0x31, 0x72, 0x9f, 0x05, 0x00, 0x00,
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
	ValidatorReputation(ctx context.Context, in *QueryValidatorReputationRequest, opts ...grpc.CallOption) (*QueryValidatorReputationResponse, error)
	ReporterReputation(ctx context.Context, in *QueryReporterReputationRequest, opts ...grpc.CallOption) (*QueryReporterReputationResponse, error)
	ReputationHistory(ctx context.Context, in *QueryReputationHistoryRequest, opts ...grpc.CallOption) (*QueryReputationHistoryResponse, error)
	ReputationParams(ctx context.Context, in *QueryReputationParamsRequest, opts ...grpc.CallOption) (*QueryReputationParamsResponse, error)
}

type queryClient struct {
	cc grpc1.ClientConn
}

func NewQueryClient(cc grpc1.ClientConn) QueryClient {
	return &queryClient{cc}
}

func (c *queryClient) ValidatorReputation(ctx context.Context, in *QueryValidatorReputationRequest, opts ...grpc.CallOption) (*QueryValidatorReputationResponse, error) {
	out := new(QueryValidatorReputationResponse)
	err := c.cc.Invoke(ctx, "/l1.reputation.v1.Query/ValidatorReputation", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) ReporterReputation(ctx context.Context, in *QueryReporterReputationRequest, opts ...grpc.CallOption) (*QueryReporterReputationResponse, error) {
	out := new(QueryReporterReputationResponse)
	err := c.cc.Invoke(ctx, "/l1.reputation.v1.Query/ReporterReputation", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) ReputationHistory(ctx context.Context, in *QueryReputationHistoryRequest, opts ...grpc.CallOption) (*QueryReputationHistoryResponse, error) {
	out := new(QueryReputationHistoryResponse)
	err := c.cc.Invoke(ctx, "/l1.reputation.v1.Query/ReputationHistory", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) ReputationParams(ctx context.Context, in *QueryReputationParamsRequest, opts ...grpc.CallOption) (*QueryReputationParamsResponse, error) {
	out := new(QueryReputationParamsResponse)
	err := c.cc.Invoke(ctx, "/l1.reputation.v1.Query/ReputationParams", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QueryServer is the server API for Query service.
type QueryServer interface {
	ValidatorReputation(context.Context, *QueryValidatorReputationRequest) (*QueryValidatorReputationResponse, error)
	ReporterReputation(context.Context, *QueryReporterReputationRequest) (*QueryReporterReputationResponse, error)
	ReputationHistory(context.Context, *QueryReputationHistoryRequest) (*QueryReputationHistoryResponse, error)
	ReputationParams(context.Context, *QueryReputationParamsRequest) (*QueryReputationParamsResponse, error)
}

// UnimplementedQueryServer can be embedded to have forward compatible implementations.
type UnimplementedQueryServer struct {
}

func (*UnimplementedQueryServer) ValidatorReputation(ctx context.Context, req *QueryValidatorReputationRequest) (*QueryValidatorReputationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ValidatorReputation not implemented")
}
func (*UnimplementedQueryServer) ReporterReputation(ctx context.Context, req *QueryReporterReputationRequest) (*QueryReporterReputationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReporterReputation not implemented")
}
func (*UnimplementedQueryServer) ReputationHistory(ctx context.Context, req *QueryReputationHistoryRequest) (*QueryReputationHistoryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReputationHistory not implemented")
}
func (*UnimplementedQueryServer) ReputationParams(ctx context.Context, req *QueryReputationParamsRequest) (*QueryReputationParamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReputationParams not implemented")
}

func RegisterQueryServer(s grpc1.Server, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc, srv)
}

func _Query_ValidatorReputation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryValidatorReputationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).ValidatorReputation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.reputation.v1.Query/ValidatorReputation",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).ValidatorReputation(ctx, req.(*QueryValidatorReputationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_ReporterReputation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryReporterReputationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).ReporterReputation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.reputation.v1.Query/ReporterReputation",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).ReporterReputation(ctx, req.(*QueryReporterReputationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_ReputationHistory_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryReputationHistoryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).ReputationHistory(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.reputation.v1.Query/ReputationHistory",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).ReputationHistory(ctx, req.(*QueryReputationHistoryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_ReputationParams_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryReputationParamsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).ReputationParams(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.reputation.v1.Query/ReputationParams",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).ReputationParams(ctx, req.(*QueryReputationParamsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var Query_serviceDesc = _Query_serviceDesc
var _Query_serviceDesc = grpc.ServiceDesc{
	ServiceName:	"l1.reputation.v1.Query",
	HandlerType:	(*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName:	"ValidatorReputation",
			Handler:	_Query_ValidatorReputation_Handler,
		},
		{
			MethodName:	"ReporterReputation",
			Handler:	_Query_ReporterReputation_Handler,
		},
		{
			MethodName:	"ReputationHistory",
			Handler:	_Query_ReputationHistory_Handler,
		},
		{
			MethodName:	"ReputationParams",
			Handler:	_Query_ReputationParams_Handler,
		},
	},
	Streams:	[]grpc.StreamDesc{},
	Metadata:	"l1/reputation/v1/query.proto",
}

func (m *QueryValidatorReputationRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryValidatorReputationRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryValidatorReputationRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Address) > 0 {
		i -= len(m.Address)
		copy(dAtA[i:], m.Address)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.Address)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *QueryReporterReputationRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryReporterReputationRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryReporterReputationRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Address) > 0 {
		i -= len(m.Address)
		copy(dAtA[i:], m.Address)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.Address)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *QueryValidatorReputationResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryValidatorReputationResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryValidatorReputationResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.RecordJson) > 0 {
		i -= len(m.RecordJson)
		copy(dAtA[i:], m.RecordJson)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.RecordJson)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *QueryReporterReputationResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryReporterReputationResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryReporterReputationResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.RecordJson) > 0 {
		i -= len(m.RecordJson)
		copy(dAtA[i:], m.RecordJson)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.RecordJson)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *QueryReputationHistoryRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryReputationHistoryRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryReputationHistoryRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Limit != 0 {
		i = encodeVarintQuery(dAtA, i, uint64(m.Limit))
		i--
		dAtA[i] = 0x18
	}
	if len(m.Address) > 0 {
		i -= len(m.Address)
		copy(dAtA[i:], m.Address)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.Address)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.SubjectType) > 0 {
		i -= len(m.SubjectType)
		copy(dAtA[i:], m.SubjectType)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.SubjectType)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *QueryReputationHistoryResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryReputationHistoryResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryReputationHistoryResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.EventsJson) > 0 {
		i -= len(m.EventsJson)
		copy(dAtA[i:], m.EventsJson)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.EventsJson)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.SnapshotsJson) > 0 {
		i -= len(m.SnapshotsJson)
		copy(dAtA[i:], m.SnapshotsJson)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.SnapshotsJson)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *QueryReputationParamsRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryReputationParamsRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryReputationParamsRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *QueryReputationParamsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryReputationParamsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryReputationParamsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
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
func (m *QueryValidatorReputationRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Address)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *QueryReporterReputationRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Address)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *QueryValidatorReputationResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.RecordJson)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *QueryReporterReputationResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.RecordJson)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *QueryReputationHistoryRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.SubjectType)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	l = len(m.Address)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	if m.Limit != 0 {
		n += 1 + sovQuery(uint64(m.Limit))
	}
	return n
}

func (m *QueryReputationHistoryResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.SnapshotsJson)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	l = len(m.EventsJson)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *QueryReputationParamsRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *QueryReputationParamsResponse) Size() (n int) {
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
func (m *QueryValidatorReputationRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryValidatorReputationRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryValidatorReputationRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Address", wireType)
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
			m.Address = string(dAtA[iNdEx:postIndex])
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
func (m *QueryReporterReputationRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryReporterReputationRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryReporterReputationRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Address", wireType)
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
			m.Address = string(dAtA[iNdEx:postIndex])
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
func (m *QueryValidatorReputationResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryValidatorReputationResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryValidatorReputationResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RecordJson", wireType)
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
			m.RecordJson = string(dAtA[iNdEx:postIndex])
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
func (m *QueryReporterReputationResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryReporterReputationResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryReporterReputationResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RecordJson", wireType)
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
			m.RecordJson = string(dAtA[iNdEx:postIndex])
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
func (m *QueryReputationHistoryRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryReputationHistoryRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryReputationHistoryRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SubjectType", wireType)
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
			m.SubjectType = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Address", wireType)
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
			m.Address = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Limit", wireType)
			}
			m.Limit = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Limit |= uint32(b&0x7F) << shift
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
func (m *QueryReputationHistoryResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryReputationHistoryResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryReputationHistoryResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SnapshotsJson", wireType)
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
			m.SnapshotsJson = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field EventsJson", wireType)
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
			m.EventsJson = string(dAtA[iNdEx:postIndex])
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
func (m *QueryReputationParamsRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryReputationParamsRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryReputationParamsRequest: illegal tag %d (wire type %d)", fieldNum, wire)
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
func (m *QueryReputationParamsResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryReputationParamsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryReputationParamsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
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
