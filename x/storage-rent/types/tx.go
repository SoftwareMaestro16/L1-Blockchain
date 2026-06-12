package types

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"math/bits"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	gogoproto "github.com/cosmos/gogoproto/proto"
	grpc1 "github.com/cosmos/gogoproto/grpc"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	proto2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	MsgPayStorageRentTypeURL	= "/l1.storagerent.v1.MsgPayStorageRent"
	MsgUnfreezeContractTypeURL	= "/l1.storagerent.v1.MsgUnfreezeContract"
)

var ErrIntOverflowTx = fmt.Errorf("proto: integer overflow")
var ErrInvalidLengthTx = fmt.Errorf("proto: negative length found during unmarshaling")

func sovTx(x uint64) int {
	return (bits.Len64(x|1) + 6) / 7
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

type MsgPayStorageRent struct {
	Payer		string
	ContractAddress	string
	Amount		uint64
	Height		uint64
}

func (m *MsgPayStorageRent) Reset()		{ *m = MsgPayStorageRent{} }
func (m *MsgPayStorageRent) String() string	{ return gogoproto.CompactTextString(m) }
func (*MsgPayStorageRent) ProtoMessage()	{}

func (m *MsgPayStorageRent) ValidateBasic() error {
	if m.Payer == "" {
		return fmt.Errorf("payer is required")
	}
	if m.ContractAddress == "" {
		return fmt.Errorf("contract address is required")
	}
	if m.Amount == 0 {
		return fmt.Errorf("amount must be positive")
	}
	if m.Height == 0 {
		return fmt.Errorf("height must be positive")
	}
	return nil
}

func (m *MsgPayStorageRent) GetSigners() []sdk.AccAddress {
	payer, err := sdk.AccAddressFromBech32(m.Payer)
	if err != nil {
		return nil
	}
	return []sdk.AccAddress{payer}
}

func (m *MsgPayStorageRent) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}

func (m *MsgPayStorageRent) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return gogoproto.Marshal(m)
	}
	b = b[:cap(b)]
	n, err := m.MarshalToSizedBuffer(b)
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}

func (m *MsgPayStorageRent) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgPayStorageRent) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgPayStorageRent) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	if m.Height != 0 {
		i = encodeVarintTx(dAtA, i, m.Height)
		i--
		dAtA[i] = 0x20
	}
	if m.Amount != 0 {
		i = encodeVarintTx(dAtA, i, m.Amount)
		i--
		dAtA[i] = 0x18
	}
	if len(m.ContractAddress) > 0 {
		i -= len(m.ContractAddress)
		copy(dAtA[i:], m.ContractAddress)
		i = encodeVarintTx(dAtA, i, uint64(len(m.ContractAddress)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Payer) > 0 {
		i -= len(m.Payer)
		copy(dAtA[i:], m.Payer)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Payer)))
		i--
		dAtA[i] = 0x0a
	}
	if i < 0 {
		return 0, io.ErrShortBuffer
	}
	return len(dAtA) - i, nil
}

func (m *MsgPayStorageRent) Size() int {
	if m == nil {
		return 0
	}
	var n int
	var l int
	_ = l
	l = len(m.Payer)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.ContractAddress)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if m.Amount != 0 {
		n += 1 + sovTx(m.Amount)
	}
	if m.Height != 0 {
		n += 1 + sovTx(m.Height)
	}
	return n
}

func (m *MsgPayStorageRent) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgPayStorageRent: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgPayStorageRent: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Payer", wireType)
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
			m.Payer = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ContractAddress", wireType)
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
			m.ContractAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
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
				m.Amount |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
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
			if skippy < 0 {
				return ErrInvalidLengthTx
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthTx
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
				return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

type MsgPayStorageRentResponse struct{}

func (m *MsgPayStorageRentResponse) Reset()		{ *m = MsgPayStorageRentResponse{} }
func (m *MsgPayStorageRentResponse) String() string	{ return "MsgPayStorageRentResponse" }
func (*MsgPayStorageRentResponse) ProtoMessage()	{}

func (m *MsgPayStorageRentResponse) Marshal() (dAtA []byte, err error) {
	return nil, nil
}
func (m *MsgPayStorageRentResponse) MarshalTo(dAtA []byte) (int, error) {
	return 0, nil
}
func (m *MsgPayStorageRentResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	return len(dAtA), nil
}
func (m *MsgPayStorageRentResponse) Size() int {
	return 0
}
func (m *MsgPayStorageRentResponse) Unmarshal(dAtA []byte) error {
	return nil
}
func (m *MsgPayStorageRentResponse) XXX_Unmarshal(b []byte) error {
	return nil
}

type MsgUnfreezeContract struct {
	Payer		string
	ContractAddress	string
	Amount		uint64
	Height		uint64
}

func (m *MsgUnfreezeContract) Reset()		{ *m = MsgUnfreezeContract{} }
func (m *MsgUnfreezeContract) String() string	{ return gogoproto.CompactTextString(m) }
func (*MsgUnfreezeContract) ProtoMessage()	{}

func (m *MsgUnfreezeContract) ValidateBasic() error {
	if m.Payer == "" {
		return fmt.Errorf("payer is required")
	}
	if m.ContractAddress == "" {
		return fmt.Errorf("contract address is required")
	}
	if m.Amount == 0 {
		return fmt.Errorf("amount must be positive")
	}
	if m.Height == 0 {
		return fmt.Errorf("height must be positive")
	}
	return nil
}

func (m *MsgUnfreezeContract) GetSigners() []sdk.AccAddress {
	payer, err := sdk.AccAddressFromBech32(m.Payer)
	if err != nil {
		return nil
	}
	return []sdk.AccAddress{payer}
}

func (m *MsgUnfreezeContract) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}

func (m *MsgUnfreezeContract) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return gogoproto.Marshal(m)
	}
	b = b[:cap(b)]
	n, err := m.MarshalToSizedBuffer(b)
	if err != nil {
		return nil, err
	}
	return b[:n], nil
}

func (m *MsgUnfreezeContract) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgUnfreezeContract) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgUnfreezeContract) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	if m.Height != 0 {
		i = encodeVarintTx(dAtA, i, m.Height)
		i--
		dAtA[i] = 0x20
	}
	if m.Amount != 0 {
		i = encodeVarintTx(dAtA, i, m.Amount)
		i--
		dAtA[i] = 0x18
	}
	if len(m.ContractAddress) > 0 {
		i -= len(m.ContractAddress)
		copy(dAtA[i:], m.ContractAddress)
		i = encodeVarintTx(dAtA, i, uint64(len(m.ContractAddress)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Payer) > 0 {
		i -= len(m.Payer)
		copy(dAtA[i:], m.Payer)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Payer)))
		i--
		dAtA[i] = 0x0a
	}
	if i < 0 {
		return 0, io.ErrShortBuffer
	}
	return len(dAtA) - i, nil
}

func (m *MsgUnfreezeContract) Size() int {
	if m == nil {
		return 0
	}
	var n int
	var l int
	_ = l
	l = len(m.Payer)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.ContractAddress)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if m.Amount != 0 {
		n += 1 + sovTx(m.Amount)
	}
	if m.Height != 0 {
		n += 1 + sovTx(m.Height)
	}
	return n
}

func (m *MsgUnfreezeContract) Unmarshal(dAtA []byte) error {
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
			return fmt.Errorf("proto: MsgUnfreezeContract: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgUnfreezeContract: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Payer", wireType)
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
			m.Payer = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ContractAddress", wireType)
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
			m.ContractAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
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
				m.Amount |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
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
			if skippy < 0 {
				return ErrInvalidLengthTx
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthTx
			}
			iNdEx += skippy
		}
	}
	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}

type MsgUnfreezeContractResponse struct{}

func (m *MsgUnfreezeContractResponse) Reset()		{ *m = MsgUnfreezeContractResponse{} }
func (m *MsgUnfreezeContractResponse) String() string	{ return "MsgUnfreezeContractResponse" }
func (*MsgUnfreezeContractResponse) ProtoMessage()	{}

func (m *MsgUnfreezeContractResponse) Marshal() (dAtA []byte, err error) {
	return nil, nil
}
func (m *MsgUnfreezeContractResponse) MarshalTo(dAtA []byte) (int, error) {
	return 0, nil
}
func (m *MsgUnfreezeContractResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	return len(dAtA), nil
}
func (m *MsgUnfreezeContractResponse) Size() int {
	return 0
}
func (m *MsgUnfreezeContractResponse) Unmarshal(dAtA []byte) error {
	return nil
}
func (m *MsgUnfreezeContractResponse) XXX_Unmarshal(b []byte) error {
	return nil
}

func init() {
	gogoproto.RegisterType((*MsgPayStorageRent)(nil), "l1.storagerent.v1.MsgPayStorageRent")
	gogoproto.RegisterType((*MsgPayStorageRentResponse)(nil), "l1.storagerent.v1.MsgPayStorageRentResponse")
	gogoproto.RegisterType((*MsgUnfreezeContract)(nil), "l1.storagerent.v1.MsgUnfreezeContract")
	gogoproto.RegisterType((*MsgUnfreezeContractResponse)(nil), "l1.storagerent.v1.MsgUnfreezeContractResponse")
	gogoproto.RegisterFile("l1/storagerent/v1/tx.proto", buildStorageRentTxFileDescriptor())
}

var (
	_	gogoproto.Message	= &MsgPayStorageRent{}
	_	gogoproto.Message	= &MsgPayStorageRentResponse{}
	_	gogoproto.Message	= &MsgUnfreezeContract{}
	_	gogoproto.Message	= &MsgUnfreezeContractResponse{}
)

type MsgServer interface {
	PayStorageRent(context.Context, *MsgPayStorageRent) (*MsgPayStorageRentResponse, error)
	UnfreezeContract(context.Context, *MsgUnfreezeContract) (*MsgUnfreezeContractResponse, error)
}

type UnimplementedMsgServer struct{}

func (UnimplementedMsgServer) PayStorageRent(ctx context.Context, req *MsgPayStorageRent) (*MsgPayStorageRentResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PayStorageRent not implemented")
}

func (UnimplementedMsgServer) UnfreezeContract(ctx context.Context, req *MsgUnfreezeContract) (*MsgUnfreezeContractResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UnfreezeContract not implemented")
}

func RegisterMsgServer(s grpc1.Server, srv MsgServer) {
	s.RegisterService(&Msg_serviceDesc, srv)
}

var Msg_serviceDesc = Msg_serviceDescVal
var Msg_serviceDescVal = grpc.ServiceDesc{
	ServiceName:	"l1.storagerent.v1.Msg",
	HandlerType:	(*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName:	"PayStorageRent",
			Handler:	_Msg_PayStorageRent_Handler,
		},
		{
			MethodName:	"UnfreezeContract",
			Handler:	_Msg_UnfreezeContract_Handler,
		},
	},
	Streams:	[]grpc.StreamDesc{},
	Metadata:	"l1/storagerent/v1/tx.proto",
}

func _Msg_PayStorageRent_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgPayStorageRent)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).PayStorageRent(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.storagerent.v1.Msg/PayStorageRent",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).PayStorageRent(ctx, req.(*MsgPayStorageRent))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_UnfreezeContract_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUnfreezeContract)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).UnfreezeContract(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:		srv,
		FullMethod:	"/l1.storagerent.v1.Msg/UnfreezeContract",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).UnfreezeContract(ctx, req.(*MsgUnfreezeContract))
	}
	return interceptor(ctx, in, info, handler)
}

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgPayStorageRent{},
		&MsgUnfreezeContract{},
	)
}

func buildStorageRentTxFileDescriptor() []byte {
	fd := &descriptorpb.FileDescriptorProto{
		Name:		descriptorString("l1/storagerent/v1/tx.proto"),
		Package:	descriptorString("l1.storagerent.v1"),
		Syntax:		descriptorString("proto3"),
		Options: &descriptorpb.FileOptions{
			GoPackage: descriptorString("github.com/sovereign-l1/l1/x/storage-rent/types"),
		},
		MessageType: []*descriptorpb.DescriptorProto{
			messageDescriptor("MsgPayStorageRent"),
			messageDescriptor("MsgPayStorageRentResponse"),
			messageDescriptor("MsgUnfreezeContract"),
			messageDescriptor("MsgUnfreezeContractResponse"),
		},
		Service: []*descriptorpb.ServiceDescriptorProto{
			{
				Name:	descriptorString("Msg"),
				Method: []*descriptorpb.MethodDescriptorProto{
					serviceMethodDescriptor("PayStorageRent", "MsgPayStorageRent", "MsgPayStorageRentResponse"),
					serviceMethodDescriptor("UnfreezeContract", "MsgUnfreezeContract", "MsgUnfreezeContractResponse"),
				},
			},
		},
	}
	raw, err := proto2.Marshal(fd)
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write(raw); err != nil {
		panic(err)
	}
	if err := zw.Close(); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func messageDescriptor(name string) *descriptorpb.DescriptorProto {
	return &descriptorpb.DescriptorProto{Name: descriptorString(name)}
}

func serviceMethodDescriptor(name, input, output string) *descriptorpb.MethodDescriptorProto {
	return &descriptorpb.MethodDescriptorProto{
		Name:		descriptorString(name),
		InputType:	descriptorString(".l1.storagerent.v1." + input),
		OutputType:	descriptorString(".l1.storagerent.v1." + output),
	}
}

func descriptorString(value string) *string {
	return &value
}
