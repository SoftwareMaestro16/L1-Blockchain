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

type QueryParamsRequest struct {
}

func (m *QueryParamsRequest) Reset()		{ *m = QueryParamsRequest{} }
func (m *QueryParamsRequest) String() string	{ return proto.CompactTextString(m) }
func (*QueryParamsRequest) ProtoMessage()	{}
func (*QueryParamsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_fa9f0ef4b4dc6535, []int{0}
}
func (m *QueryParamsRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryParamsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryParamsRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryParamsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryParamsRequest.Merge(m, src)
}
func (m *QueryParamsRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryParamsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryParamsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryParamsRequest proto.InternalMessageInfo

type QueryParamsResponse struct {
	Params Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
}

func (m *QueryParamsResponse) Reset()		{ *m = QueryParamsResponse{} }
func (m *QueryParamsResponse) String() string	{ return proto.CompactTextString(m) }
func (*QueryParamsResponse) ProtoMessage()	{}
func (*QueryParamsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_fa9f0ef4b4dc6535, []int{1}
}
func (m *QueryParamsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryParamsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryParamsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryParamsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryParamsResponse.Merge(m, src)
}
func (m *QueryParamsResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryParamsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryParamsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryParamsResponse proto.InternalMessageInfo

func (m *QueryParamsResponse) GetParams() Params {
	if m != nil {
		return m.Params
	}
	return Params{}
}

type QueryAccountingRequest struct {
}

func (m *QueryAccountingRequest) Reset()		{ *m = QueryAccountingRequest{} }
func (m *QueryAccountingRequest) String() string	{ return proto.CompactTextString(m) }
func (*QueryAccountingRequest) ProtoMessage()		{}
func (*QueryAccountingRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_fa9f0ef4b4dc6535, []int{2}
}
func (m *QueryAccountingRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryAccountingRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryAccountingRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryAccountingRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryAccountingRequest.Merge(m, src)
}
func (m *QueryAccountingRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryAccountingRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryAccountingRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryAccountingRequest proto.InternalMessageInfo

type QueryAccountingResponse struct {
	ProtocolFeeState ProtocolFeeState `protobuf:"bytes,1,opt,name=protocol_fee_state,json=protocolFeeState,proto3" json:"protocol_fee_state"`
}

func (m *QueryAccountingResponse) Reset()		{ *m = QueryAccountingResponse{} }
func (m *QueryAccountingResponse) String() string	{ return proto.CompactTextString(m) }
func (*QueryAccountingResponse) ProtoMessage()		{}
func (*QueryAccountingResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_fa9f0ef4b4dc6535, []int{3}
}
func (m *QueryAccountingResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryAccountingResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryAccountingResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryAccountingResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryAccountingResponse.Merge(m, src)
}
func (m *QueryAccountingResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryAccountingResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryAccountingResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryAccountingResponse proto.InternalMessageInfo

func (m *QueryAccountingResponse) GetProtocolFeeState() ProtocolFeeState {
	if m != nil {
		return m.ProtocolFeeState
	}
	return ProtocolFeeState{}
}

type QueryModuleBalancesRequest struct {
}

func (m *QueryModuleBalancesRequest) Reset()		{ *m = QueryModuleBalancesRequest{} }
func (m *QueryModuleBalancesRequest) String() string	{ return proto.CompactTextString(m) }
func (*QueryModuleBalancesRequest) ProtoMessage()	{}
func (*QueryModuleBalancesRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_fa9f0ef4b4dc6535, []int{4}
}
func (m *QueryModuleBalancesRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryModuleBalancesRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryModuleBalancesRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryModuleBalancesRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryModuleBalancesRequest.Merge(m, src)
}
func (m *QueryModuleBalancesRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryModuleBalancesRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryModuleBalancesRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryModuleBalancesRequest proto.InternalMessageInfo

type QueryModuleBalancesResponse struct {
	Balances []ModuleBalance `protobuf:"bytes,1,rep,name=balances,proto3" json:"balances"`
}

func (m *QueryModuleBalancesResponse) Reset()		{ *m = QueryModuleBalancesResponse{} }
func (m *QueryModuleBalancesResponse) String() string	{ return proto.CompactTextString(m) }
func (*QueryModuleBalancesResponse) ProtoMessage()	{}
func (*QueryModuleBalancesResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_fa9f0ef4b4dc6535, []int{5}
}
func (m *QueryModuleBalancesResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryModuleBalancesResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryModuleBalancesResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryModuleBalancesResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryModuleBalancesResponse.Merge(m, src)
}
func (m *QueryModuleBalancesResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryModuleBalancesResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryModuleBalancesResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryModuleBalancesResponse proto.InternalMessageInfo

func (m *QueryModuleBalancesResponse) GetBalances() []ModuleBalance {
	if m != nil {
		return m.Balances
	}
	return nil
}

type QueryNetworkLoadRequest struct {
}

func (m *QueryNetworkLoadRequest) Reset()		{ *m = QueryNetworkLoadRequest{} }
func (m *QueryNetworkLoadRequest) String() string	{ return proto.CompactTextString(m) }
func (*QueryNetworkLoadRequest) ProtoMessage()		{}
func (*QueryNetworkLoadRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_fa9f0ef4b4dc6535, []int{6}
}
func (m *QueryNetworkLoadRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryNetworkLoadRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryNetworkLoadRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryNetworkLoadRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryNetworkLoadRequest.Merge(m, src)
}
func (m *QueryNetworkLoadRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryNetworkLoadRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryNetworkLoadRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryNetworkLoadRequest proto.InternalMessageInfo

type QueryNetworkLoadResponse struct {
	BlockGasConsumed	uint64	`protobuf:"varint,1,opt,name=block_gas_consumed,json=blockGasConsumed,proto3" json:"block_gas_consumed,omitempty"`
	MaxBlockGas		uint64	`protobuf:"varint,2,opt,name=max_block_gas,json=maxBlockGas,proto3" json:"max_block_gas,omitempty"`
	UtilizationBps		uint32	`protobuf:"varint,3,opt,name=utilization_bps,json=utilizationBps,proto3" json:"utilization_bps,omitempty"`
	Congested		bool	`protobuf:"varint,4,opt,name=congested,proto3" json:"congested,omitempty"`
}

func (m *QueryNetworkLoadResponse) Reset()		{ *m = QueryNetworkLoadResponse{} }
func (m *QueryNetworkLoadResponse) String() string	{ return proto.CompactTextString(m) }
func (*QueryNetworkLoadResponse) ProtoMessage()		{}
func (*QueryNetworkLoadResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_fa9f0ef4b4dc6535, []int{7}
}
func (m *QueryNetworkLoadResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryNetworkLoadResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryNetworkLoadResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryNetworkLoadResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryNetworkLoadResponse.Merge(m, src)
}
func (m *QueryNetworkLoadResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryNetworkLoadResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryNetworkLoadResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryNetworkLoadResponse proto.InternalMessageInfo

func (m *QueryNetworkLoadResponse) GetBlockGasConsumed() uint64 {
	if m != nil {
		return m.BlockGasConsumed
	}
	return 0
}

func (m *QueryNetworkLoadResponse) GetMaxBlockGas() uint64 {
	if m != nil {
		return m.MaxBlockGas
	}
	return 0
}

func (m *QueryNetworkLoadResponse) GetUtilizationBps() uint32 {
	if m != nil {
		return m.UtilizationBps
	}
	return 0
}

func (m *QueryNetworkLoadResponse) GetCongested() bool {
	if m != nil {
		return m.Congested
	}
	return false
}

type QueryEstimateFeeRequest struct {
	GasLimit uint64 `protobuf:"varint,1,opt,name=gas_limit,json=gasLimit,proto3" json:"gas_limit,omitempty"`
}

func (m *QueryEstimateFeeRequest) Reset()		{ *m = QueryEstimateFeeRequest{} }
func (m *QueryEstimateFeeRequest) String() string	{ return proto.CompactTextString(m) }
func (*QueryEstimateFeeRequest) ProtoMessage()		{}
func (*QueryEstimateFeeRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_fa9f0ef4b4dc6535, []int{8}
}
func (m *QueryEstimateFeeRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryEstimateFeeRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryEstimateFeeRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryEstimateFeeRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryEstimateFeeRequest.Merge(m, src)
}
func (m *QueryEstimateFeeRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryEstimateFeeRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryEstimateFeeRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryEstimateFeeRequest proto.InternalMessageInfo

func (m *QueryEstimateFeeRequest) GetGasLimit() uint64 {
	if m != nil {
		return m.GasLimit
	}
	return 0
}

type QueryEstimateFeeResponse struct {
	RequiredFee	string	`protobuf:"bytes,1,opt,name=required_fee,json=requiredFee,proto3" json:"required_fee,omitempty"`
	BaseFee		string	`protobuf:"bytes,2,opt,name=base_fee,json=baseFee,proto3" json:"base_fee,omitempty"`
	MaxFee		string	`protobuf:"bytes,3,opt,name=max_fee,json=maxFee,proto3" json:"max_fee,omitempty"`
	UtilizationBps	uint32	`protobuf:"varint,4,opt,name=utilization_bps,json=utilizationBps,proto3" json:"utilization_bps,omitempty"`
	Congested	bool	`protobuf:"varint,5,opt,name=congested,proto3" json:"congested,omitempty"`
	AtHardCap	bool	`protobuf:"varint,6,opt,name=at_hard_cap,json=atHardCap,proto3" json:"at_hard_cap,omitempty"`
}

func (m *QueryEstimateFeeResponse) Reset()		{ *m = QueryEstimateFeeResponse{} }
func (m *QueryEstimateFeeResponse) String() string	{ return proto.CompactTextString(m) }
func (*QueryEstimateFeeResponse) ProtoMessage()		{}
func (*QueryEstimateFeeResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_fa9f0ef4b4dc6535, []int{9}
}
func (m *QueryEstimateFeeResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryEstimateFeeResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryEstimateFeeResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryEstimateFeeResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryEstimateFeeResponse.Merge(m, src)
}
func (m *QueryEstimateFeeResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryEstimateFeeResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryEstimateFeeResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryEstimateFeeResponse proto.InternalMessageInfo

func (m *QueryEstimateFeeResponse) GetRequiredFee() string {
	if m != nil {
		return m.RequiredFee
	}
	return ""
}

func (m *QueryEstimateFeeResponse) GetBaseFee() string {
	if m != nil {
		return m.BaseFee
	}
	return ""
}

func (m *QueryEstimateFeeResponse) GetMaxFee() string {
	if m != nil {
		return m.MaxFee
	}
	return ""
}

func (m *QueryEstimateFeeResponse) GetUtilizationBps() uint32 {
	if m != nil {
		return m.UtilizationBps
	}
	return 0
}

func (m *QueryEstimateFeeResponse) GetCongested() bool {
	if m != nil {
		return m.Congested
	}
	return false
}

func (m *QueryEstimateFeeResponse) GetAtHardCap() bool {
	if m != nil {
		return m.AtHardCap
	}
	return false
}

func init() {
	proto.RegisterType((*QueryParamsRequest)(nil), "l1.fees.v1.QueryParamsRequest")
	proto.RegisterType((*QueryParamsResponse)(nil), "l1.fees.v1.QueryParamsResponse")
	proto.RegisterType((*QueryAccountingRequest)(nil), "l1.fees.v1.QueryAccountingRequest")
	proto.RegisterType((*QueryAccountingResponse)(nil), "l1.fees.v1.QueryAccountingResponse")
	proto.RegisterType((*QueryModuleBalancesRequest)(nil), "l1.fees.v1.QueryModuleBalancesRequest")
	proto.RegisterType((*QueryModuleBalancesResponse)(nil), "l1.fees.v1.QueryModuleBalancesResponse")
	proto.RegisterType((*QueryNetworkLoadRequest)(nil), "l1.fees.v1.QueryNetworkLoadRequest")
	proto.RegisterType((*QueryNetworkLoadResponse)(nil), "l1.fees.v1.QueryNetworkLoadResponse")
	proto.RegisterType((*QueryEstimateFeeRequest)(nil), "l1.fees.v1.QueryEstimateFeeRequest")
	proto.RegisterType((*QueryEstimateFeeResponse)(nil), "l1.fees.v1.QueryEstimateFeeResponse")
}

func init()	{ proto.RegisterFile("l1/fees/v1/query.proto", fileDescriptor_fa9f0ef4b4dc6535) }

var fileDescriptor_fa9f0ef4b4dc6535 = []byte{

	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x54, 0x41, 0x4f, 0xdb, 0x4a,
	0x18, 0x8c, 0x21, 0x04, 0xd8, 0x3c, 0x78, 0x68, 0x1f, 0x02, 0x63, 0xf2, 0x4c, 0x6a, 0xaa, 0x92,
	0x43, 0x1b, 0x37, 0x54, 0xea, 0xa5, 0x27, 0x82, 0x0a, 0x3d, 0xd0, 0x8a, 0xa6, 0x37, 0x2e, 0xd6,
	0xc6, 0xf9, 0x30, 0x16, 0xb6, 0xd7, 0x78, 0xd7, 0x34, 0x70, 0xec, 0xa1, 0xe7, 0x4a, 0xfd, 0x29,
	0xfd, 0x13, 0x1c, 0x91, 0x7a, 0xe1, 0x54, 0x55, 0x50, 0xf5, 0x77, 0x54, 0xbb, 0xde, 0x10, 0x07,
	0xa7, 0x70, 0x8b, 0xbf, 0x99, 0xfd, 0x66, 0xc6, 0xde, 0x09, 0x5a, 0x0a, 0x5a, 0xf6, 0x21, 0x00,
	0xb3, 0x4f, 0x5b, 0xf6, 0x49, 0x0a, 0xc9, 0x59, 0x33, 0x4e, 0x28, 0xa7, 0x18, 0x05, 0xad, 0xa6,
	0x98, 0x37, 0x4f, 0x5b, 0x46, 0xcd, 0xa3, 0xd4, 0x0b, 0xc0, 0x26, 0xb1, 0x6f, 0x93, 0x28, 0xa2,
	0x9c, 0x70, 0x9f, 0x46, 0x2c, 0x63, 0x1a, 0x8b, 0x1e, 0xf5, 0xa8, 0xfc, 0x69, 0x8b, 0x5f, 0x6a,
	0xaa, 0xe7, 0xf6, 0x7a, 0x10, 0x01, 0xf3, 0x15, 0xdf, 0x5a, 0x44, 0xf8, 0xbd, 0x10, 0xda, 0x27,
	0x09, 0x09, 0x59, 0x07, 0x4e, 0x52, 0x60, 0xdc, 0xda, 0x45, 0xff, 0x8d, 0x4c, 0x59, 0x4c, 0x23,
	0x06, 0xf8, 0x39, 0xaa, 0xc4, 0x72, 0xa2, 0x6b, 0x75, 0xad, 0x51, 0xdd, 0xc4, 0xcd, 0xa1, 0xaf,
	0x66, 0xc6, 0x6d, 0x97, 0x2f, 0x7e, 0xac, 0x95, 0x3a, 0x8a, 0x67, 0xe9, 0x68, 0x49, 0x2e, 0xda,
	0x72, 0x5d, 0x9a, 0x46, 0xdc, 0x8f, 0xbc, 0x81, 0xc4, 0x31, 0x5a, 0x2e, 0x20, 0x4a, 0x66, 0x1f,
	0x61, 0x69, 0xce, 0xa5, 0x81, 0x73, 0x08, 0xe0, 0x30, 0x4e, 0x38, 0x28, 0xc9, 0xda, 0x88, 0xa4,
	0x62, 0xed, 0x00, 0x7c, 0x10, 0x1c, 0x25, 0xbe, 0x10, 0xdf, 0x99, 0x5b, 0x35, 0x64, 0x48, 0xb1,
	0xb7, 0xb4, 0x97, 0x06, 0xd0, 0x26, 0x01, 0x89, 0x5c, 0xb8, 0x4d, 0x7b, 0x80, 0x56, 0xc7, 0xa2,
	0xca, 0xce, 0x2b, 0x34, 0xd3, 0x55, 0x33, 0x5d, 0xab, 0x4f, 0x36, 0xaa, 0x9b, 0x2b, 0x79, 0x13,
	0x23, 0xa7, 0x94, 0x83, 0xdb, 0x03, 0xd6, 0x8a, 0x8a, 0xf9, 0x0e, 0xf8, 0x47, 0x9a, 0x1c, 0xef,
	0x51, 0xd2, 0x1b, 0xc8, 0x7e, 0xd3, 0x90, 0x5e, 0xc4, 0x94, 0xe8, 0x53, 0x84, 0xbb, 0x01, 0x75,
	0x8f, 0x1d, 0x8f, 0x30, 0xc7, 0xa5, 0x11, 0x4b, 0x43, 0xe8, 0xc9, 0x77, 0x50, 0xee, 0x2c, 0x48,
	0x64, 0x97, 0xb0, 0x6d, 0x35, 0xc7, 0x16, 0x9a, 0x0b, 0x49, 0xdf, 0xb9, 0x3d, 0xa1, 0x4f, 0x48,
	0x62, 0x35, 0x24, 0xfd, 0xb6, 0xe2, 0xe2, 0x0d, 0xf4, 0x6f, 0xca, 0xfd, 0xc0, 0x3f, 0x97, 0xf7,
	0xc5, 0xe9, 0xc6, 0x4c, 0x9f, 0xac, 0x6b, 0x8d, 0xb9, 0xce, 0x7c, 0x6e, 0xdc, 0x8e, 0x19, 0xae,
	0xa1, 0x59, 0x97, 0x46, 0x1e, 0x30, 0x0e, 0x3d, 0xbd, 0x5c, 0xd7, 0x1a, 0x33, 0x9d, 0xe1, 0xc0,
	0x7a, 0xa9, 0x02, 0xbd, 0x66, 0xdc, 0x0f, 0x09, 0x87, 0x1d, 0x00, 0x15, 0x08, 0xaf, 0xa2, 0x59,
	0xe1, 0x36, 0xf0, 0x43, 0x9f, 0x2b, 0xab, 0x33, 0x1e, 0x61, 0x7b, 0xe2, 0xd9, 0xba, 0x1a, 0xa4,
	0x1d, 0x39, 0xa8, 0xd2, 0x3e, 0x42, 0xff, 0x24, 0x70, 0x92, 0xfa, 0x09, 0xf4, 0xc4, 0x17, 0x97,
	0x87, 0x67, 0x3b, 0xd5, 0xc1, 0x6c, 0x07, 0x00, 0xaf, 0x88, 0xaf, 0xc0, 0x40, 0xc2, 0x13, 0x12,
	0x9e, 0x16, 0xcf, 0x02, 0x5a, 0x46, 0xd3, 0x22, 0xbd, 0x40, 0x26, 0x25, 0x52, 0x09, 0x49, 0x5f,
	0x00, 0x63, 0x22, 0x97, 0x1f, 0x8e, 0x3c, 0x75, 0x27, 0x32, 0x36, 0x51, 0x95, 0x70, 0xe7, 0x88,
	0x24, 0x3d, 0xc7, 0x25, 0xb1, 0x5e, 0xc9, 0x70, 0xc2, 0xdf, 0x90, 0xa4, 0xb7, 0x4d, 0xe2, 0xcd,
	0xdf, 0x65, 0x34, 0x25, 0xa3, 0x61, 0x40, 0x95, 0xac, 0x06, 0xd8, 0xcc, 0x5f, 0x91, 0x62, 0xc3,
	0x8c, 0xb5, 0xbf, 0xe2, 0xd9, 0x2b, 0xb1, 0x8c, 0x4f, 0xdf, 0x7f, 0x7d, 0x9d, 0x58, 0xc4, 0xd8,
	0xce, 0x75, 0x37, 0x6b, 0x15, 0x4e, 0x11, 0x1a, 0xd6, 0x06, 0x5b, 0x85, 0x55, 0x85, 0xb6, 0x19,
	0xeb, 0xf7, 0x72, 0x94, 0xa4, 0x29, 0x25, 0x75, 0xbc, 0x94, 0x97, 0x24, 0x43, 0xa1, 0xcf, 0x1a,
	0x9a, 0x1f, 0xed, 0x08, 0x7e, 0x52, 0xd8, 0x3b, 0xb6, 0x62, 0xc6, 0xc6, 0x83, 0x3c, 0xe5, 0x61,
	0x5d, 0x7a, 0xf8, 0x1f, 0xaf, 0xe6, 0x3d, 0x84, 0x92, 0xeb, 0x0c, 0x4a, 0x85, 0xcf, 0x51, 0x35,
	0xd7, 0x19, 0x5c, 0x0c, 0x57, 0x6c, 0x9b, 0xf1, 0xf8, 0x7e, 0x92, 0x92, 0xaf, 0x4b, 0x79, 0x03,
	0xeb, 0x79, 0xf9, 0x28, 0x23, 0x3a, 0x81, 0x10, 0x3b, 0x47, 0xd5, 0xdc, 0x0d, 0x1e, 0xa3, 0x5d,
	0x2c, 0xc6, 0x18, 0xed, 0x31, 0x25, 0x18, 0xaf, 0x0d, 0x8a, 0x28, 0x6e, 0x77, 0x7b, 0xeb, 0xe2,
	0xda, 0xd4, 0x2e, 0xaf, 0x4d, 0xed, 0xe7, 0xb5, 0xa9, 0x7d, 0xb9, 0x31, 0x4b, 0x97, 0x37, 0x66,
	0xe9, 0xea, 0xc6, 0x2c, 0x1d, 0x6c, 0x78, 0x3e, 0x3f, 0x4a, 0xbb, 0x4d, 0x97, 0x86, 0x36, 0xa3,
	0xa7, 0x90, 0x80, 0xef, 0x45, 0xcf, 0x82, 0x96, 0x58, 0xd5, 0xcf, 0x96, 0xf1, 0xb3, 0x18, 0x58,
	0xb7, 0x22, 0xff, 0x1b, 0x5f, 0xfc, 0x09, 0x00, 0x00, 0xff, 0xff, 0x73, 0xae, 0x41, 0xcc, 0x6a,
	0x06, 0x00, 0x00,
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
	Params(ctx context.Context, in *QueryParamsRequest, opts ...grpc.CallOption) (*QueryParamsResponse, error)
	Accounting(ctx context.Context, in *QueryAccountingRequest, opts ...grpc.CallOption) (*QueryAccountingResponse, error)
	ModuleBalances(ctx context.Context, in *QueryModuleBalancesRequest, opts ...grpc.CallOption) (*QueryModuleBalancesResponse, error)
	NetworkLoad(ctx context.Context, in *QueryNetworkLoadRequest, opts ...grpc.CallOption) (*QueryNetworkLoadResponse, error)
	EstimateFee(ctx context.Context, in *QueryEstimateFeeRequest, opts ...grpc.CallOption) (*QueryEstimateFeeResponse, error)
}

type queryClient struct {
	cc grpc1.ClientConn
}

func NewQueryClient(cc grpc1.ClientConn) QueryClient {
	return &queryClient{cc}
}

func (c *queryClient) Params(ctx context.Context, in *QueryParamsRequest, opts ...grpc.CallOption) (*QueryParamsResponse, error) {
	out := new(QueryParamsResponse)
	err := c.cc.Invoke(ctx, "/l1.fees.v1.Query/Params", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) Accounting(ctx context.Context, in *QueryAccountingRequest, opts ...grpc.CallOption) (*QueryAccountingResponse, error) {
	out := new(QueryAccountingResponse)
	err := c.cc.Invoke(ctx, "/l1.fees.v1.Query/Accounting", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) ModuleBalances(ctx context.Context, in *QueryModuleBalancesRequest, opts ...grpc.CallOption) (*QueryModuleBalancesResponse, error) {
	out := new(QueryModuleBalancesResponse)
	err := c.cc.Invoke(ctx, "/l1.fees.v1.Query/ModuleBalances", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) NetworkLoad(ctx context.Context, in *QueryNetworkLoadRequest, opts ...grpc.CallOption) (*QueryNetworkLoadResponse, error) {
	out := new(QueryNetworkLoadResponse)
	err := c.cc.Invoke(ctx, "/l1.fees.v1.Query/NetworkLoad", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) EstimateFee(ctx context.Context, in *QueryEstimateFeeRequest, opts ...grpc.CallOption) (*QueryEstimateFeeResponse, error) {
	out := new(QueryEstimateFeeResponse)
	err := c.cc.Invoke(ctx, "/l1.fees.v1.Query/EstimateFee", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QueryServer is the server API for Query service.
type QueryServer interface {
	Params(context.Context, *QueryParamsRequest) (*QueryParamsResponse, error)
	Accounting(context.Context, *QueryAccountingRequest) (*QueryAccountingResponse, error)
	ModuleBalances(context.Context, *QueryModuleBalancesRequest) (*QueryModuleBalancesResponse, error)
	NetworkLoad(context.Context, *QueryNetworkLoadRequest) (*QueryNetworkLoadResponse, error)
	EstimateFee(context.Context, *QueryEstimateFeeRequest) (*QueryEstimateFeeResponse, error)
}

// UnimplementedQueryServer can be embedded to have forward compatible implementations.
type UnimplementedQueryServer struct {
}

func (*UnimplementedQueryServer) Params(ctx context.Context, req *QueryParamsRequest) (*QueryParamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Params not implemented")
}
func (*UnimplementedQueryServer) Accounting(ctx context.Context, req *QueryAccountingRequest) (*QueryAccountingResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Accounting not implemented")
}
func (*UnimplementedQueryServer) ModuleBalances(ctx context.Context, req *QueryModuleBalancesRequest) (*QueryModuleBalancesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ModuleBalances not implemented")
}
func (*UnimplementedQueryServer) NetworkLoad(ctx context.Context, req *QueryNetworkLoadRequest) (*QueryNetworkLoadResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NetworkLoad not implemented")
}
func (*UnimplementedQueryServer) EstimateFee(ctx context.Context, req *QueryEstimateFeeRequest) (*QueryEstimateFeeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method EstimateFee not implemented")
}

func RegisterQueryServer(s grpc1.Server, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc, srv)
}

func _Query_Params_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryParamsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).Params(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.fees.v1.Query/Params",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Params(ctx, req.(*QueryParamsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_Accounting_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryAccountingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).Accounting(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.fees.v1.Query/Accounting",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Accounting(ctx, req.(*QueryAccountingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_ModuleBalances_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryModuleBalancesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).ModuleBalances(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.fees.v1.Query/ModuleBalances",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).ModuleBalances(ctx, req.(*QueryModuleBalancesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_NetworkLoad_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryNetworkLoadRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).NetworkLoad(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.fees.v1.Query/NetworkLoad",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).NetworkLoad(ctx, req.(*QueryNetworkLoadRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_EstimateFee_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryEstimateFeeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).EstimateFee(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.fees.v1.Query/EstimateFee",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).EstimateFee(ctx, req.(*QueryEstimateFeeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var Query_serviceDesc = _Query_serviceDesc
var _Query_serviceDesc = grpc.ServiceDesc{
	ServiceName:	"l1.fees.v1.Query",
	HandlerType:	(*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName:	"Params",
			Handler:	_Query_Params_Handler,
		},
		{
			MethodName:	"Accounting",
			Handler:	_Query_Accounting_Handler,
		},
		{
			MethodName:	"ModuleBalances",
			Handler:	_Query_ModuleBalances_Handler,
		},
		{
			MethodName:	"NetworkLoad",
			Handler:	_Query_NetworkLoad_Handler,
		},
		{
			MethodName:	"EstimateFee",
			Handler:	_Query_EstimateFee_Handler,
		},
	},
	Streams:	[]grpc.StreamDesc{},
	Metadata:	"l1/fees/v1/query.proto",
}

func (m *QueryParamsRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryParamsRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryParamsRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *QueryParamsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryParamsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryParamsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
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

func (m *QueryAccountingRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryAccountingRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryAccountingRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *QueryAccountingResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryAccountingResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryAccountingResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.ProtocolFeeState.MarshalToSizedBuffer(dAtA[:i])
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

func (m *QueryModuleBalancesRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryModuleBalancesRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryModuleBalancesRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *QueryModuleBalancesResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryModuleBalancesResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryModuleBalancesResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Balances) > 0 {
		for iNdEx := len(m.Balances) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Balances[iNdEx].MarshalToSizedBuffer(dAtA[:i])
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

func (m *QueryNetworkLoadRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryNetworkLoadRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryNetworkLoadRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *QueryNetworkLoadResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryNetworkLoadResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryNetworkLoadResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Congested {
		i--
		if m.Congested {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x20
	}
	if m.UtilizationBps != 0 {
		i = encodeVarintQuery(dAtA, i, uint64(m.UtilizationBps))
		i--
		dAtA[i] = 0x18
	}
	if m.MaxBlockGas != 0 {
		i = encodeVarintQuery(dAtA, i, uint64(m.MaxBlockGas))
		i--
		dAtA[i] = 0x10
	}
	if m.BlockGasConsumed != 0 {
		i = encodeVarintQuery(dAtA, i, uint64(m.BlockGasConsumed))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *QueryEstimateFeeRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryEstimateFeeRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryEstimateFeeRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.GasLimit != 0 {
		i = encodeVarintQuery(dAtA, i, uint64(m.GasLimit))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *QueryEstimateFeeResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryEstimateFeeResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryEstimateFeeResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.AtHardCap {
		i--
		if m.AtHardCap {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x30
	}
	if m.Congested {
		i--
		if m.Congested {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x28
	}
	if m.UtilizationBps != 0 {
		i = encodeVarintQuery(dAtA, i, uint64(m.UtilizationBps))
		i--
		dAtA[i] = 0x20
	}
	if len(m.MaxFee) > 0 {
		i -= len(m.MaxFee)
		copy(dAtA[i:], m.MaxFee)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.MaxFee)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.BaseFee) > 0 {
		i -= len(m.BaseFee)
		copy(dAtA[i:], m.BaseFee)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.BaseFee)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.RequiredFee) > 0 {
		i -= len(m.RequiredFee)
		copy(dAtA[i:], m.RequiredFee)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.RequiredFee)))
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
func (m *QueryParamsRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *QueryParamsResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Params.Size()
	n += 1 + l + sovQuery(uint64(l))
	return n
}

func (m *QueryAccountingRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *QueryAccountingResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.ProtocolFeeState.Size()
	n += 1 + l + sovQuery(uint64(l))
	return n
}

func (m *QueryModuleBalancesRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *QueryModuleBalancesResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Balances) > 0 {
		for _, e := range m.Balances {
			l = e.Size()
			n += 1 + l + sovQuery(uint64(l))
		}
	}
	return n
}

func (m *QueryNetworkLoadRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *QueryNetworkLoadResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.BlockGasConsumed != 0 {
		n += 1 + sovQuery(uint64(m.BlockGasConsumed))
	}
	if m.MaxBlockGas != 0 {
		n += 1 + sovQuery(uint64(m.MaxBlockGas))
	}
	if m.UtilizationBps != 0 {
		n += 1 + sovQuery(uint64(m.UtilizationBps))
	}
	if m.Congested {
		n += 2
	}
	return n
}

func (m *QueryEstimateFeeRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.GasLimit != 0 {
		n += 1 + sovQuery(uint64(m.GasLimit))
	}
	return n
}

func (m *QueryEstimateFeeResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.RequiredFee)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	l = len(m.BaseFee)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	l = len(m.MaxFee)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	if m.UtilizationBps != 0 {
		n += 1 + sovQuery(uint64(m.UtilizationBps))
	}
	if m.Congested {
		n += 2
	}
	if m.AtHardCap {
		n += 2
	}
	return n
}

func sovQuery(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozQuery(x uint64) (n int) {
	return sovQuery(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *QueryParamsRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryParamsRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryParamsRequest: illegal tag %d (wire type %d)", fieldNum, wire)
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
func (m *QueryParamsResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryParamsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryParamsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
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
func (m *QueryAccountingRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryAccountingRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryAccountingRequest: illegal tag %d (wire type %d)", fieldNum, wire)
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
func (m *QueryAccountingResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryAccountingResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryAccountingResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ProtocolFeeState", wireType)
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
			if err := m.ProtocolFeeState.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
func (m *QueryModuleBalancesRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryModuleBalancesRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryModuleBalancesRequest: illegal tag %d (wire type %d)", fieldNum, wire)
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
func (m *QueryModuleBalancesResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryModuleBalancesResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryModuleBalancesResponse: illegal tag %d (wire type %d)", fieldNum, wire)
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
			m.Balances = append(m.Balances, ModuleBalance{})
			if err := m.Balances[len(m.Balances)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
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
func (m *QueryNetworkLoadRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryNetworkLoadRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryNetworkLoadRequest: illegal tag %d (wire type %d)", fieldNum, wire)
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
func (m *QueryNetworkLoadResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryNetworkLoadResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryNetworkLoadResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field BlockGasConsumed", wireType)
			}
			m.BlockGasConsumed = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.BlockGasConsumed |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxBlockGas", wireType)
			}
			m.MaxBlockGas = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxBlockGas |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field UtilizationBps", wireType)
			}
			m.UtilizationBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.UtilizationBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Congested", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
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
			m.Congested = bool(v != 0)
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
func (m *QueryEstimateFeeRequest) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryEstimateFeeRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryEstimateFeeRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field GasLimit", wireType)
			}
			m.GasLimit = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.GasLimit |= uint64(b&0x7F) << shift
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
func (m *QueryEstimateFeeResponse) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: QueryEstimateFeeResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryEstimateFeeResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RequiredFee", wireType)
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
			m.RequiredFee = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BaseFee", wireType)
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
			m.BaseFee = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxFee", wireType)
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
			m.MaxFee = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field UtilizationBps", wireType)
			}
			m.UtilizationBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.UtilizationBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Congested", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
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
			m.Congested = bool(v != 0)
		case 6:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field AtHardCap", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
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
			m.AtHardCap = bool(v != 0)
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
