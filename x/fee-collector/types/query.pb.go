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

type QueryFeeCollectorRequest struct {
}

func (m *QueryFeeCollectorRequest) Reset()		{ *m = QueryFeeCollectorRequest{} }
func (m *QueryFeeCollectorRequest) String() string	{ return proto.CompactTextString(m) }
func (*QueryFeeCollectorRequest) ProtoMessage()		{}
func (*QueryFeeCollectorRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_67cb074878eea08a, []int{0}
}
func (m *QueryFeeCollectorRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryFeeCollectorRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryFeeCollectorRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryFeeCollectorRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryFeeCollectorRequest.Merge(m, src)
}
func (m *QueryFeeCollectorRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryFeeCollectorRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryFeeCollectorRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryFeeCollectorRequest proto.InternalMessageInfo

type QueryFeeCollectorResponse struct {
	ModuleName		string			`protobuf:"bytes,1,opt,name=module_name,json=moduleName,proto3" json:"module_name,omitempty"`
	ModuleAccount		string			`protobuf:"bytes,2,opt,name=module_account,json=moduleAccount,proto3" json:"module_account,omitempty"`
	Balances		FeeBalances		`protobuf:"bytes,3,opt,name=balances,proto3" json:"balances"`
	PendingDistribution	PendingDistribution	`protobuf:"bytes,4,opt,name=pending_distribution,json=pendingDistribution,proto3" json:"pending_distribution"`
}

func (m *QueryFeeCollectorResponse) Reset()		{ *m = QueryFeeCollectorResponse{} }
func (m *QueryFeeCollectorResponse) String() string	{ return proto.CompactTextString(m) }
func (*QueryFeeCollectorResponse) ProtoMessage()	{}
func (*QueryFeeCollectorResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_67cb074878eea08a, []int{1}
}
func (m *QueryFeeCollectorResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryFeeCollectorResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryFeeCollectorResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryFeeCollectorResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryFeeCollectorResponse.Merge(m, src)
}
func (m *QueryFeeCollectorResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryFeeCollectorResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryFeeCollectorResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryFeeCollectorResponse proto.InternalMessageInfo

func (m *QueryFeeCollectorResponse) GetModuleName() string {
	if m != nil {
		return m.ModuleName
	}
	return ""
}

func (m *QueryFeeCollectorResponse) GetModuleAccount() string {
	if m != nil {
		return m.ModuleAccount
	}
	return ""
}

func (m *QueryFeeCollectorResponse) GetBalances() FeeBalances {
	if m != nil {
		return m.Balances
	}
	return FeeBalances{}
}

func (m *QueryFeeCollectorResponse) GetPendingDistribution() PendingDistribution {
	if m != nil {
		return m.PendingDistribution
	}
	return PendingDistribution{}
}

type QueryFeeBalancesRequest struct {
}

func (m *QueryFeeBalancesRequest) Reset()		{ *m = QueryFeeBalancesRequest{} }
func (m *QueryFeeBalancesRequest) String() string	{ return proto.CompactTextString(m) }
func (*QueryFeeBalancesRequest) ProtoMessage()		{}
func (*QueryFeeBalancesRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_67cb074878eea08a, []int{2}
}
func (m *QueryFeeBalancesRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryFeeBalancesRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryFeeBalancesRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryFeeBalancesRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryFeeBalancesRequest.Merge(m, src)
}
func (m *QueryFeeBalancesRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryFeeBalancesRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryFeeBalancesRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryFeeBalancesRequest proto.InternalMessageInfo

type QueryFeeBalancesResponse struct {
	Balances FeeBalances `protobuf:"bytes,1,opt,name=balances,proto3" json:"balances"`
}

func (m *QueryFeeBalancesResponse) Reset()		{ *m = QueryFeeBalancesResponse{} }
func (m *QueryFeeBalancesResponse) String() string	{ return proto.CompactTextString(m) }
func (*QueryFeeBalancesResponse) ProtoMessage()		{}
func (*QueryFeeBalancesResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_67cb074878eea08a, []int{3}
}
func (m *QueryFeeBalancesResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryFeeBalancesResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryFeeBalancesResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryFeeBalancesResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryFeeBalancesResponse.Merge(m, src)
}
func (m *QueryFeeBalancesResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryFeeBalancesResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryFeeBalancesResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryFeeBalancesResponse proto.InternalMessageInfo

func (m *QueryFeeBalancesResponse) GetBalances() FeeBalances {
	if m != nil {
		return m.Balances
	}
	return FeeBalances{}
}

type QueryFeeDistributionRequest struct {
}

func (m *QueryFeeDistributionRequest) Reset()		{ *m = QueryFeeDistributionRequest{} }
func (m *QueryFeeDistributionRequest) String() string	{ return proto.CompactTextString(m) }
func (*QueryFeeDistributionRequest) ProtoMessage()	{}
func (*QueryFeeDistributionRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_67cb074878eea08a, []int{4}
}
func (m *QueryFeeDistributionRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryFeeDistributionRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryFeeDistributionRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryFeeDistributionRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryFeeDistributionRequest.Merge(m, src)
}
func (m *QueryFeeDistributionRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryFeeDistributionRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryFeeDistributionRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryFeeDistributionRequest proto.InternalMessageInfo

type QueryFeeDistributionResponse struct {
	Params			Params			`protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
	PendingDistribution	PendingDistribution	`protobuf:"bytes,2,opt,name=pending_distribution,json=pendingDistribution,proto3" json:"pending_distribution"`
}

func (m *QueryFeeDistributionResponse) Reset()		{ *m = QueryFeeDistributionResponse{} }
func (m *QueryFeeDistributionResponse) String() string	{ return proto.CompactTextString(m) }
func (*QueryFeeDistributionResponse) ProtoMessage()	{}
func (*QueryFeeDistributionResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_67cb074878eea08a, []int{5}
}
func (m *QueryFeeDistributionResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryFeeDistributionResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryFeeDistributionResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryFeeDistributionResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryFeeDistributionResponse.Merge(m, src)
}
func (m *QueryFeeDistributionResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryFeeDistributionResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryFeeDistributionResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryFeeDistributionResponse proto.InternalMessageInfo

func (m *QueryFeeDistributionResponse) GetParams() Params {
	if m != nil {
		return m.Params
	}
	return Params{}
}

func (m *QueryFeeDistributionResponse) GetPendingDistribution() PendingDistribution {
	if m != nil {
		return m.PendingDistribution
	}
	return PendingDistribution{}
}

type QueryFeeHistoryRequest struct {
	Epoch uint64 `protobuf:"varint,1,opt,name=epoch,proto3" json:"epoch,omitempty"`
}

func (m *QueryFeeHistoryRequest) Reset()		{ *m = QueryFeeHistoryRequest{} }
func (m *QueryFeeHistoryRequest) String() string	{ return proto.CompactTextString(m) }
func (*QueryFeeHistoryRequest) ProtoMessage()		{}
func (*QueryFeeHistoryRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_67cb074878eea08a, []int{6}
}
func (m *QueryFeeHistoryRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryFeeHistoryRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryFeeHistoryRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryFeeHistoryRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryFeeHistoryRequest.Merge(m, src)
}
func (m *QueryFeeHistoryRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryFeeHistoryRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryFeeHistoryRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryFeeHistoryRequest proto.InternalMessageInfo

func (m *QueryFeeHistoryRequest) GetEpoch() uint64 {
	if m != nil {
		return m.Epoch
	}
	return 0
}

type QueryFeeHistoryResponse struct {
	History []FeeHistoryEntry `protobuf:"bytes,1,rep,name=history,proto3" json:"history"`
}

func (m *QueryFeeHistoryResponse) Reset()		{ *m = QueryFeeHistoryResponse{} }
func (m *QueryFeeHistoryResponse) String() string	{ return proto.CompactTextString(m) }
func (*QueryFeeHistoryResponse) ProtoMessage()		{}
func (*QueryFeeHistoryResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_67cb074878eea08a, []int{7}
}
func (m *QueryFeeHistoryResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryFeeHistoryResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryFeeHistoryResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryFeeHistoryResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryFeeHistoryResponse.Merge(m, src)
}
func (m *QueryFeeHistoryResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryFeeHistoryResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryFeeHistoryResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryFeeHistoryResponse proto.InternalMessageInfo

func (m *QueryFeeHistoryResponse) GetHistory() []FeeHistoryEntry {
	if m != nil {
		return m.History
	}
	return nil
}

func init() {
	proto.RegisterType((*QueryFeeCollectorRequest)(nil), "l1.feecollector.v1.QueryFeeCollectorRequest")
	proto.RegisterType((*QueryFeeCollectorResponse)(nil), "l1.feecollector.v1.QueryFeeCollectorResponse")
	proto.RegisterType((*QueryFeeBalancesRequest)(nil), "l1.feecollector.v1.QueryFeeBalancesRequest")
	proto.RegisterType((*QueryFeeBalancesResponse)(nil), "l1.feecollector.v1.QueryFeeBalancesResponse")
	proto.RegisterType((*QueryFeeDistributionRequest)(nil), "l1.feecollector.v1.QueryFeeDistributionRequest")
	proto.RegisterType((*QueryFeeDistributionResponse)(nil), "l1.feecollector.v1.QueryFeeDistributionResponse")
	proto.RegisterType((*QueryFeeHistoryRequest)(nil), "l1.feecollector.v1.QueryFeeHistoryRequest")
	proto.RegisterType((*QueryFeeHistoryResponse)(nil), "l1.feecollector.v1.QueryFeeHistoryResponse")
}

func init()	{ proto.RegisterFile("l1/feecollector/v1/query.proto", fileDescriptor_67cb074878eea08a) }

var fileDescriptor_67cb074878eea08a = []byte{

	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x94, 0x4f, 0x6f, 0xd3, 0x4c,
	0x10, 0xc6, 0xb3, 0x69, 0xda, 0xf7, 0x65, 0xc3, 0x1f, 0x69, 0x89, 0xc0, 0x75, 0x13, 0x27, 0x72,
	0xa9, 0x88, 0x68, 0x63, 0x37, 0xe1, 0xc2, 0xb5, 0x29, 0x54, 0x88, 0x03, 0x82, 0x1c, 0x91, 0xa0,
	0x38, 0xce, 0xe0, 0x58, 0x72, 0x76, 0x5d, 0x7b, 0x1d, 0x91, 0x2b, 0x27, 0xc4, 0x01, 0x21, 0xb8,
	0xf3, 0x41, 0xf8, 0x04, 0x3d, 0x56, 0xe2, 0x82, 0x84, 0x84, 0x50, 0xc2, 0x07, 0x41, 0xb1, 0xd7,
	0x89, 0x69, 0x37, 0x51, 0x84, 0xb8, 0xd9, 0x3b, 0xcf, 0xcc, 0xfc, 0xe6, 0xf1, 0xac, 0xb1, 0xe6,
	0x35, 0xcd, 0x57, 0x00, 0x36, 0xf3, 0x3c, 0xb0, 0x39, 0x0b, 0xcc, 0x61, 0xd3, 0x3c, 0x89, 0x20,
	0x18, 0x19, 0x7e, 0xc0, 0x38, 0x23, 0xc4, 0x6b, 0x1a, 0xd9, 0xb8, 0x31, 0x6c, 0xaa, 0x65, 0x87,
	0x31, 0xc7, 0x03, 0xd3, 0xf2, 0x5d, 0xd3, 0xa2, 0x94, 0x71, 0x8b, 0xbb, 0x8c, 0x86, 0x49, 0x86,
	0x5a, 0x72, 0x98, 0xc3, 0xe2, 0x47, 0x73, 0xfa, 0x24, 0x4e, 0x6b, 0x92, 0x3e, 0x0e, 0x50, 0x08,
	0x5d, 0x91, 0xa7, 0xab, 0x58, 0x79, 0x3a, 0x6d, 0x7c, 0x04, 0x70, 0x98, 0xaa, 0x3a, 0x70, 0x12,
	0x41, 0xc8, 0xf5, 0xb7, 0x79, 0xbc, 0x29, 0x09, 0x86, 0x3e, 0xa3, 0x21, 0x90, 0x2a, 0x2e, 0x0e,
	0x58, 0x2f, 0xf2, 0xe0, 0x98, 0x5a, 0x03, 0x50, 0x50, 0x0d, 0xd5, 0x2f, 0x75, 0x70, 0x72, 0xf4,
	0xd8, 0x1a, 0x00, 0xd9, 0xc1, 0x57, 0x85, 0xc0, 0xb2, 0x6d, 0x16, 0x51, 0xae, 0xe4, 0x63, 0xcd,
	0x95, 0xe4, 0xf4, 0x20, 0x39, 0x24, 0x07, 0xf8, 0xff, 0xae, 0xe5, 0x59, 0xd4, 0x86, 0x50, 0x59,
	0xab, 0xa1, 0x7a, 0xb1, 0x55, 0x35, 0x2e, 0x8e, 0x6f, 0x1c, 0x01, 0xb4, 0x85, 0xac, 0x5d, 0x38,
	0xfd, 0x51, 0xcd, 0x75, 0x66, 0x69, 0xe4, 0x25, 0x2e, 0xf9, 0x40, 0x7b, 0x2e, 0x75, 0x8e, 0x7b,
	0x6e, 0xc8, 0x03, 0xb7, 0x1b, 0x4d, 0xbd, 0x51, 0x0a, 0x71, 0xb9, 0xdb, 0xb2, 0x72, 0x4f, 0x12,
	0xfd, 0xfd, 0x8c, 0x5c, 0x94, 0xbd, 0xee, 0x5f, 0x0c, 0xe9, 0x9b, 0xf8, 0x66, 0xea, 0x44, 0x4a,
	0x91, 0xba, 0xf4, 0x7c, 0xee, 0xe0, 0x3c, 0x24, 0x3c, 0xca, 0xce, 0x86, 0xfe, 0x6a, 0x36, 0xbd,
	0x82, 0xb7, 0xd2, 0xf2, 0x59, 0xa2, 0xb4, 0xfb, 0x17, 0x84, 0xcb, 0xf2, 0xb8, 0x40, 0xb8, 0x87,
	0x37, 0x7c, 0x2b, 0xb0, 0x06, 0x29, 0x80, 0x2a, 0x75, 0x23, 0x56, 0x88, 0xde, 0x42, 0xbf, 0xd0,
	0xd5, 0xfc, 0x3f, 0x73, 0xd5, 0xc0, 0x37, 0x52, 0xf6, 0x87, 0x6e, 0xc8, 0x59, 0x30, 0x12, 0x63,
	0x91, 0x12, 0x5e, 0x07, 0x9f, 0xd9, 0xfd, 0x18, 0xba, 0xd0, 0x49, 0x5e, 0xf4, 0x17, 0xf3, 0xaf,
	0x30, 0xd3, 0x8b, 0x31, 0x0f, 0xf1, 0x7f, 0xfd, 0xe4, 0x48, 0x41, 0xb5, 0xb5, 0x7a, 0xb1, 0xb5,
	0xbd, 0xc0, 0x68, 0x91, 0xf8, 0x80, 0xf2, 0x60, 0x24, 0xd8, 0xd2, 0xcc, 0xd6, 0xf7, 0x02, 0x5e,
	0x8f, 0x1b, 0x90, 0x8f, 0x08, 0x5f, 0xce, 0x6e, 0x3d, 0xd9, 0x93, 0x95, 0x5b, 0x74, 0x73, 0xd4,
	0xc6, 0x8a, 0xea, 0x04, 0x5e, 0xdf, 0x79, 0xf3, 0xf5, 0xd7, 0xa7, 0x7c, 0x95, 0x54, 0x4c, 0xc9,
	0x7d, 0x9d, 0xbd, 0x90, 0xf7, 0x08, 0x17, 0x33, 0xab, 0x42, 0x76, 0x97, 0x75, 0x39, 0xb7, 0xa6,
	0xea, 0xde, 0x6a, 0x62, 0x41, 0x74, 0x2b, 0x26, 0xd2, 0x48, 0x59, 0x46, 0x34, 0xbb, 0x77, 0x9f,
	0x11, 0xbe, 0x76, 0x6e, 0xef, 0x88, 0xb9, 0xac, 0x8f, 0x64, 0x83, 0xd5, 0xfd, 0xd5, 0x13, 0x04,
	0x5c, 0x3d, 0x86, 0xd3, 0x49, 0x4d, 0x06, 0x97, 0x5d, 0x55, 0xf2, 0x0e, 0x61, 0x3c, 0xff, 0xe6,
	0xe4, 0xce, 0xb2, 0x56, 0x7f, 0x6e, 0xa0, 0xba, 0xbb, 0x92, 0x56, 0x10, 0x6d, 0xc7, 0x44, 0x15,
	0xb2, 0x25, 0x23, 0x12, 0xdb, 0xd5, 0x7e, 0x74, 0x3a, 0xd6, 0xd0, 0xd9, 0x58, 0x43, 0x3f, 0xc7,
	0x1a, 0xfa, 0x30, 0xd1, 0x72, 0x67, 0x13, 0x2d, 0xf7, 0x6d, 0xa2, 0xe5, 0x9e, 0xed, 0x3b, 0x2e,
	0xef, 0x47, 0x5d, 0xc3, 0x66, 0x03, 0x33, 0x64, 0x43, 0x08, 0xc0, 0x75, 0x68, 0xc3, 0x6b, 0x4e,
	0xab, 0xbd, 0x9e, 0xd6, 0x6b, 0xcc, 0x0b, 0xf2, 0x91, 0x0f, 0x61, 0x77, 0x23, 0xfe, 0x7b, 0xdf,
	0xfd, 0x1d, 0x00, 0x00, 0xff, 0xff, 0xdf, 0xee, 0x2c, 0x07, 0x49, 0x06, 0x00, 0x00,
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
	FeeCollector(ctx context.Context, in *QueryFeeCollectorRequest, opts ...grpc.CallOption) (*QueryFeeCollectorResponse, error)
	FeeBalances(ctx context.Context, in *QueryFeeBalancesRequest, opts ...grpc.CallOption) (*QueryFeeBalancesResponse, error)
	FeeDistribution(ctx context.Context, in *QueryFeeDistributionRequest, opts ...grpc.CallOption) (*QueryFeeDistributionResponse, error)
	FeeHistory(ctx context.Context, in *QueryFeeHistoryRequest, opts ...grpc.CallOption) (*QueryFeeHistoryResponse, error)
}

type queryClient struct {
	cc grpc1.ClientConn
}

func NewQueryClient(cc grpc1.ClientConn) QueryClient {
	return &queryClient{cc}
}

func (c *queryClient) FeeCollector(ctx context.Context, in *QueryFeeCollectorRequest, opts ...grpc.CallOption) (*QueryFeeCollectorResponse, error) {
	out := new(QueryFeeCollectorResponse)
	err := c.cc.Invoke(ctx, "/l1.feecollector.v1.Query/FeeCollector", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) FeeBalances(ctx context.Context, in *QueryFeeBalancesRequest, opts ...grpc.CallOption) (*QueryFeeBalancesResponse, error) {
	out := new(QueryFeeBalancesResponse)
	err := c.cc.Invoke(ctx, "/l1.feecollector.v1.Query/FeeBalances", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) FeeDistribution(ctx context.Context, in *QueryFeeDistributionRequest, opts ...grpc.CallOption) (*QueryFeeDistributionResponse, error) {
	out := new(QueryFeeDistributionResponse)
	err := c.cc.Invoke(ctx, "/l1.feecollector.v1.Query/FeeDistribution", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) FeeHistory(ctx context.Context, in *QueryFeeHistoryRequest, opts ...grpc.CallOption) (*QueryFeeHistoryResponse, error) {
	out := new(QueryFeeHistoryResponse)
	err := c.cc.Invoke(ctx, "/l1.feecollector.v1.Query/FeeHistory", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QueryServer is the server API for Query service.
type QueryServer interface {
	FeeCollector(context.Context, *QueryFeeCollectorRequest) (*QueryFeeCollectorResponse, error)
	FeeBalances(context.Context, *QueryFeeBalancesRequest) (*QueryFeeBalancesResponse, error)
	FeeDistribution(context.Context, *QueryFeeDistributionRequest) (*QueryFeeDistributionResponse, error)
	FeeHistory(context.Context, *QueryFeeHistoryRequest) (*QueryFeeHistoryResponse, error)
}

// UnimplementedQueryServer can be embedded to have forward compatible implementations.
type UnimplementedQueryServer struct {
}

func (*UnimplementedQueryServer) FeeCollector(ctx context.Context, req *QueryFeeCollectorRequest) (*QueryFeeCollectorResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FeeCollector not implemented")
}
func (*UnimplementedQueryServer) FeeBalances(ctx context.Context, req *QueryFeeBalancesRequest) (*QueryFeeBalancesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FeeBalances not implemented")
}
func (*UnimplementedQueryServer) FeeDistribution(ctx context.Context, req *QueryFeeDistributionRequest) (*QueryFeeDistributionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FeeDistribution not implemented")
}
func (*UnimplementedQueryServer) FeeHistory(ctx context.Context, req *QueryFeeHistoryRequest) (*QueryFeeHistoryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FeeHistory not implemented")
}

func RegisterQueryServer(s grpc1.Server, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc, srv)
}

func _Query_FeeCollector_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryFeeCollectorRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).FeeCollector(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.feecollector.v1.Query/FeeCollector",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).FeeCollector(ctx, req.(*QueryFeeCollectorRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_FeeBalances_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryFeeBalancesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).FeeBalances(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.feecollector.v1.Query/FeeBalances",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).FeeBalances(ctx, req.(*QueryFeeBalancesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_FeeDistribution_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryFeeDistributionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).FeeDistribution(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.feecollector.v1.Query/FeeDistribution",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).FeeDistribution(ctx, req.(*QueryFeeDistributionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_FeeHistory_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryFeeHistoryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).FeeHistory(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.feecollector.v1.Query/FeeHistory",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).FeeHistory(ctx, req.(*QueryFeeHistoryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var Query_serviceDesc = _Query_serviceDesc
var _Query_serviceDesc = grpc.ServiceDesc{
	ServiceName:	"l1.feecollector.v1.Query",
	HandlerType:	(*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName:	"FeeCollector",
			Handler:	_Query_FeeCollector_Handler,
		},
		{
			MethodName:	"FeeBalances",
			Handler:	_Query_FeeBalances_Handler,
		},
		{
			MethodName:	"FeeDistribution",
			Handler:	_Query_FeeDistribution_Handler,
		},
		{
			MethodName:	"FeeHistory",
			Handler:	_Query_FeeHistory_Handler,
		},
	},
	Streams:	[]grpc.StreamDesc{},
	Metadata:	"l1/feecollector/v1/query.proto",
}

func (m *QueryFeeCollectorRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryFeeCollectorRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryFeeCollectorRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *QueryFeeCollectorResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryFeeCollectorResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryFeeCollectorResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.PendingDistribution.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintQuery(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x22
	{
		size, err := m.Balances.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintQuery(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x1a
	if len(m.ModuleAccount) > 0 {
		i -= len(m.ModuleAccount)
		copy(dAtA[i:], m.ModuleAccount)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.ModuleAccount)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.ModuleName) > 0 {
		i -= len(m.ModuleName)
		copy(dAtA[i:], m.ModuleName)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.ModuleName)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *QueryFeeBalancesRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryFeeBalancesRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryFeeBalancesRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *QueryFeeBalancesResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryFeeBalancesResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryFeeBalancesResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.Balances.MarshalToSizedBuffer(dAtA[:i])
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

func (m *QueryFeeDistributionRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryFeeDistributionRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryFeeDistributionRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *QueryFeeDistributionResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryFeeDistributionResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryFeeDistributionResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.PendingDistribution.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintQuery(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
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

func (m *QueryFeeHistoryRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryFeeHistoryRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryFeeHistoryRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
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

func (m *QueryFeeHistoryResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryFeeHistoryResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryFeeHistoryResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.History) > 0 {
		for iNdEx := len(m.History) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.History[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintQuery(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
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
func (m *QueryFeeCollectorRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *QueryFeeCollectorResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ModuleName)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	l = len(m.ModuleAccount)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	l = m.Balances.Size()
	n += 1 + l + sovQuery(uint64(l))
	l = m.PendingDistribution.Size()
	n += 1 + l + sovQuery(uint64(l))
	return n
}

func (m *QueryFeeBalancesRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *QueryFeeBalancesResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Balances.Size()
	n += 1 + l + sovQuery(uint64(l))
	return n
}

func (m *QueryFeeDistributionRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *QueryFeeDistributionResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Params.Size()
	n += 1 + l + sovQuery(uint64(l))
	l = m.PendingDistribution.Size()
	n += 1 + l + sovQuery(uint64(l))
	return n
}

func (m *QueryFeeHistoryRequest) Size() (n int) {
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

func (m *QueryFeeHistoryResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.History) > 0 {
		for _, e := range m.History {
			l = e.Size()
			n += 1 + l + sovQuery(uint64(l))
		}
	}
	return n
}

func sovQuery(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozQuery(x uint64) (n int) {
	return sovQuery(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *QueryFeeCollectorRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryFeeCollectorRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryFeeCollectorRequest: illegal tag %d (wire type %d)", fieldNum, wire)
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
func (m *QueryFeeCollectorResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryFeeCollectorResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryFeeCollectorResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ModuleName", wireType)
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
			m.ModuleName = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ModuleAccount", wireType)
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
			m.ModuleAccount = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Balances", wireType)
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
			if err := m.Balances.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PendingDistribution", wireType)
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
			if err := m.PendingDistribution.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
func (m *QueryFeeBalancesRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryFeeBalancesRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryFeeBalancesRequest: illegal tag %d (wire type %d)", fieldNum, wire)
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
func (m *QueryFeeBalancesResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryFeeBalancesResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryFeeBalancesResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Balances", wireType)
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
			if err := m.Balances.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
func (m *QueryFeeDistributionRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryFeeDistributionRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryFeeDistributionRequest: illegal tag %d (wire type %d)", fieldNum, wire)
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
func (m *QueryFeeDistributionResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryFeeDistributionResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryFeeDistributionResponse: illegal tag %d (wire type %d)", fieldNum, wire)
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
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PendingDistribution", wireType)
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
			if err := m.PendingDistribution.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
func (m *QueryFeeHistoryRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryFeeHistoryRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryFeeHistoryRequest: illegal tag %d (wire type %d)", fieldNum, wire)
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
func (m *QueryFeeHistoryResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryFeeHistoryResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryFeeHistoryResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field History", wireType)
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
			m.History = append(m.History, FeeHistoryEntry{})
			if err := m.History[len(m.History)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
