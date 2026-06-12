package types

import (
	fmt "fmt"
	types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
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

type GenesisState struct {
	Params			Params		`protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
	EpochHistory		[]EmissionEpoch	`protobuf:"bytes,2,rep,name=epoch_history,json=epochHistory,proto3" json:"epoch_history"`
	TotalMintedAccounting	types.Coin	`protobuf:"bytes,3,opt,name=total_minted_accounting,json=totalMintedAccounting,proto3" json:"total_minted_accounting"`
}

func (m *GenesisState) Reset()		{ *m = GenesisState{} }
func (m *GenesisState) String() string	{ return proto.CompactTextString(m) }
func (*GenesisState) ProtoMessage()	{}
func (*GenesisState) Descriptor() ([]byte, []int) {
	return fileDescriptor_900022dfbddab62a, []int{0}
}
func (m *GenesisState) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *GenesisState) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_GenesisState.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *GenesisState) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GenesisState.Merge(m, src)
}
func (m *GenesisState) XXX_Size() int {
	return m.Size()
}
func (m *GenesisState) XXX_DiscardUnknown() {
	xxx_messageInfo_GenesisState.DiscardUnknown(m)
}

var xxx_messageInfo_GenesisState proto.InternalMessageInfo

func (m *GenesisState) GetParams() Params {
	if m != nil {
		return m.Params
	}
	return Params{}
}

func (m *GenesisState) GetEpochHistory() []EmissionEpoch {
	if m != nil {
		return m.EpochHistory
	}
	return nil
}

func (m *GenesisState) GetTotalMintedAccounting() types.Coin {
	if m != nil {
		return m.TotalMintedAccounting
	}
	return types.Coin{}
}

type Params struct {
	BaseDenom			string			`protobuf:"bytes,1,opt,name=base_denom,json=baseDenom,proto3" json:"base_denom,omitempty"`
	CurrentInflationBps		uint32			`protobuf:"varint,2,opt,name=current_inflation_bps,json=currentInflationBps,proto3" json:"current_inflation_bps,omitempty"`
	TargetStakingRatioBps		uint32			`protobuf:"varint,3,opt,name=target_staking_ratio_bps,json=targetStakingRatioBps,proto3" json:"target_staking_ratio_bps,omitempty"`
	MinAnnualInflationBps		uint32			`protobuf:"varint,4,opt,name=min_annual_inflation_bps,json=minAnnualInflationBps,proto3" json:"min_annual_inflation_bps,omitempty"`
	MaxAnnualInflationBps		uint32			`protobuf:"varint,5,opt,name=max_annual_inflation_bps,json=maxAnnualInflationBps,proto3" json:"max_annual_inflation_bps,omitempty"`
	ConstitutionalMaxInflationBps	uint32			`protobuf:"varint,6,opt,name=constitutional_max_inflation_bps,json=constitutionalMaxInflationBps,proto3" json:"constitutional_max_inflation_bps,omitempty"`
	ResponsivenessBps		uint32			`protobuf:"varint,7,opt,name=responsiveness_bps,json=responsivenessBps,proto3" json:"responsiveness_bps,omitempty"`
	AnnualReferenceSupply		types.Coin		`protobuf:"bytes,8,opt,name=annual_reference_supply,json=annualReferenceSupply,proto3" json:"annual_reference_supply"`
	EpochsPerYear			uint64			`protobuf:"varint,9,opt,name=epochs_per_year,json=epochsPerYear,proto3" json:"epochs_per_year,omitempty"`
	DistributionWeights		DistributionWeights	`protobuf:"bytes,10,opt,name=distribution_weights,json=distributionWeights,proto3" json:"distribution_weights"`
}

func (m *Params) Reset()		{ *m = Params{} }
func (m *Params) String() string	{ return proto.CompactTextString(m) }
func (*Params) ProtoMessage()		{}
func (*Params) Descriptor() ([]byte, []int) {
	return fileDescriptor_900022dfbddab62a, []int{1}
}
func (m *Params) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Params) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Params.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Params) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Params.Merge(m, src)
}
func (m *Params) XXX_Size() int {
	return m.Size()
}
func (m *Params) XXX_DiscardUnknown() {
	xxx_messageInfo_Params.DiscardUnknown(m)
}

var xxx_messageInfo_Params proto.InternalMessageInfo

func (m *Params) GetBaseDenom() string {
	if m != nil {
		return m.BaseDenom
	}
	return ""
}

func (m *Params) GetCurrentInflationBps() uint32 {
	if m != nil {
		return m.CurrentInflationBps
	}
	return 0
}

func (m *Params) GetTargetStakingRatioBps() uint32 {
	if m != nil {
		return m.TargetStakingRatioBps
	}
	return 0
}

func (m *Params) GetMinAnnualInflationBps() uint32 {
	if m != nil {
		return m.MinAnnualInflationBps
	}
	return 0
}

func (m *Params) GetMaxAnnualInflationBps() uint32 {
	if m != nil {
		return m.MaxAnnualInflationBps
	}
	return 0
}

func (m *Params) GetConstitutionalMaxInflationBps() uint32 {
	if m != nil {
		return m.ConstitutionalMaxInflationBps
	}
	return 0
}

func (m *Params) GetResponsivenessBps() uint32 {
	if m != nil {
		return m.ResponsivenessBps
	}
	return 0
}

func (m *Params) GetAnnualReferenceSupply() types.Coin {
	if m != nil {
		return m.AnnualReferenceSupply
	}
	return types.Coin{}
}

func (m *Params) GetEpochsPerYear() uint64 {
	if m != nil {
		return m.EpochsPerYear
	}
	return 0
}

func (m *Params) GetDistributionWeights() DistributionWeights {
	if m != nil {
		return m.DistributionWeights
	}
	return DistributionWeights{}
}

type DistributionWeights struct {
	ValidatorRewardBps	uint32	`protobuf:"varint,1,opt,name=validator_reward_bps,json=validatorRewardBps,proto3" json:"validator_reward_bps,omitempty"`
	TreasuryBps		uint32	`protobuf:"varint,2,opt,name=treasury_bps,json=treasuryBps,proto3" json:"treasury_bps,omitempty"`
	ProtectionBps		uint32	`protobuf:"varint,3,opt,name=protection_bps,json=protectionBps,proto3" json:"protection_bps,omitempty"`
	BurnBps			uint32	`protobuf:"varint,4,opt,name=burn_bps,json=burnBps,proto3" json:"burn_bps,omitempty"`
	EcosystemBps		uint32	`protobuf:"varint,5,opt,name=ecosystem_bps,json=ecosystemBps,proto3" json:"ecosystem_bps,omitempty"`
}

func (m *DistributionWeights) Reset()		{ *m = DistributionWeights{} }
func (m *DistributionWeights) String() string	{ return proto.CompactTextString(m) }
func (*DistributionWeights) ProtoMessage()	{}
func (*DistributionWeights) Descriptor() ([]byte, []int) {
	return fileDescriptor_900022dfbddab62a, []int{2}
}
func (m *DistributionWeights) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *DistributionWeights) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_DistributionWeights.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *DistributionWeights) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DistributionWeights.Merge(m, src)
}
func (m *DistributionWeights) XXX_Size() int {
	return m.Size()
}
func (m *DistributionWeights) XXX_DiscardUnknown() {
	xxx_messageInfo_DistributionWeights.DiscardUnknown(m)
}

var xxx_messageInfo_DistributionWeights proto.InternalMessageInfo

func (m *DistributionWeights) GetValidatorRewardBps() uint32 {
	if m != nil {
		return m.ValidatorRewardBps
	}
	return 0
}

func (m *DistributionWeights) GetTreasuryBps() uint32 {
	if m != nil {
		return m.TreasuryBps
	}
	return 0
}

func (m *DistributionWeights) GetProtectionBps() uint32 {
	if m != nil {
		return m.ProtectionBps
	}
	return 0
}

func (m *DistributionWeights) GetBurnBps() uint32 {
	if m != nil {
		return m.BurnBps
	}
	return 0
}

func (m *DistributionWeights) GetEcosystemBps() uint32 {
	if m != nil {
		return m.EcosystemBps
	}
	return 0
}

type EmissionEpoch struct {
	Epoch			uint64		`protobuf:"varint,1,opt,name=epoch,proto3" json:"epoch,omitempty"`
	StakingRatioBps		uint32		`protobuf:"varint,2,opt,name=staking_ratio_bps,json=stakingRatioBps,proto3" json:"staking_ratio_bps,omitempty"`
	InflationBps		uint32		`protobuf:"varint,3,opt,name=inflation_bps,json=inflationBps,proto3" json:"inflation_bps,omitempty"`
	EmissionAmount		types.Coin	`protobuf:"bytes,4,opt,name=emission_amount,json=emissionAmount,proto3" json:"emission_amount"`
	ValidatorReward		types.Coin	`protobuf:"bytes,5,opt,name=validator_reward,json=validatorReward,proto3" json:"validator_reward"`
	Treasury		types.Coin	`protobuf:"bytes,6,opt,name=treasury,proto3" json:"treasury"`
	ProtectionFund		types.Coin	`protobuf:"bytes,7,opt,name=protection_fund,json=protectionFund,proto3" json:"protection_fund"`
	Burn			types.Coin	`protobuf:"bytes,8,opt,name=burn,proto3" json:"burn"`
	Ecosystem		types.Coin	`protobuf:"bytes,9,opt,name=ecosystem,proto3" json:"ecosystem"`
	RoundingRemainder	types.Coin	`protobuf:"bytes,10,opt,name=rounding_remainder,json=roundingRemainder,proto3" json:"rounding_remainder"`
	FinalizedHeight		int64		`protobuf:"varint,11,opt,name=finalized_height,json=finalizedHeight,proto3" json:"finalized_height,omitempty"`
}

func (m *EmissionEpoch) Reset()		{ *m = EmissionEpoch{} }
func (m *EmissionEpoch) String() string	{ return proto.CompactTextString(m) }
func (*EmissionEpoch) ProtoMessage()	{}
func (*EmissionEpoch) Descriptor() ([]byte, []int) {
	return fileDescriptor_900022dfbddab62a, []int{3}
}
func (m *EmissionEpoch) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *EmissionEpoch) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_EmissionEpoch.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *EmissionEpoch) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EmissionEpoch.Merge(m, src)
}
func (m *EmissionEpoch) XXX_Size() int {
	return m.Size()
}
func (m *EmissionEpoch) XXX_DiscardUnknown() {
	xxx_messageInfo_EmissionEpoch.DiscardUnknown(m)
}

var xxx_messageInfo_EmissionEpoch proto.InternalMessageInfo

func (m *EmissionEpoch) GetEpoch() uint64 {
	if m != nil {
		return m.Epoch
	}
	return 0
}

func (m *EmissionEpoch) GetStakingRatioBps() uint32 {
	if m != nil {
		return m.StakingRatioBps
	}
	return 0
}

func (m *EmissionEpoch) GetInflationBps() uint32 {
	if m != nil {
		return m.InflationBps
	}
	return 0
}

func (m *EmissionEpoch) GetEmissionAmount() types.Coin {
	if m != nil {
		return m.EmissionAmount
	}
	return types.Coin{}
}

func (m *EmissionEpoch) GetValidatorReward() types.Coin {
	if m != nil {
		return m.ValidatorReward
	}
	return types.Coin{}
}

func (m *EmissionEpoch) GetTreasury() types.Coin {
	if m != nil {
		return m.Treasury
	}
	return types.Coin{}
}

func (m *EmissionEpoch) GetProtectionFund() types.Coin {
	if m != nil {
		return m.ProtectionFund
	}
	return types.Coin{}
}

func (m *EmissionEpoch) GetBurn() types.Coin {
	if m != nil {
		return m.Burn
	}
	return types.Coin{}
}

func (m *EmissionEpoch) GetEcosystem() types.Coin {
	if m != nil {
		return m.Ecosystem
	}
	return types.Coin{}
}

func (m *EmissionEpoch) GetRoundingRemainder() types.Coin {
	if m != nil {
		return m.RoundingRemainder
	}
	return types.Coin{}
}

func (m *EmissionEpoch) GetFinalizedHeight() int64 {
	if m != nil {
		return m.FinalizedHeight
	}
	return 0
}

func init() {
	proto.RegisterType((*GenesisState)(nil), "l1.emissions.v1.GenesisState")
	proto.RegisterType((*Params)(nil), "l1.emissions.v1.Params")
	proto.RegisterType((*DistributionWeights)(nil), "l1.emissions.v1.DistributionWeights")
	proto.RegisterType((*EmissionEpoch)(nil), "l1.emissions.v1.EmissionEpoch")
}

func init()	{ proto.RegisterFile("l1/emissions/v1/genesis.proto", fileDescriptor_900022dfbddab62a) }

var fileDescriptor_900022dfbddab62a = []byte{

	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x95, 0xcd, 0x6e, 0xdb, 0x46,
	0x10, 0xc7, 0xcd, 0x58, 0x76, 0xec, 0xb1, 0x14, 0xc5, 0x6b, 0x1b, 0x66, 0x02, 0x58, 0x55, 0x9d,
	0xb6, 0x50, 0x8b, 0x86, 0xac, 0x1c, 0x14, 0x3d, 0x14, 0x3d, 0xd8, 0xcd, 0x87, 0x53, 0x20, 0x45,
	0x40, 0x1f, 0x8c, 0x16, 0x28, 0x88, 0x15, 0xb9, 0xa6, 0x16, 0x25, 0x77, 0x89, 0xdd, 0xa5, 0x22,
	0xf5, 0xd8, 0x27, 0xe8, 0x63, 0xe5, 0x98, 0xde, 0x7a, 0x32, 0x0a, 0xfb, 0xd4, 0xb7, 0x28, 0x76,
	0x48, 0xd1, 0x92, 0x6c, 0x20, 0xba, 0x71, 0xe7, 0x3f, 0xbf, 0xd9, 0x99, 0x9d, 0x19, 0x10, 0x0e,
	0xd2, 0xbe, 0xcf, 0x32, 0xae, 0x35, 0x97, 0x42, 0xfb, 0xa3, 0xbe, 0x9f, 0x30, 0xc1, 0x34, 0xd7,
	0x5e, 0xae, 0xa4, 0x91, 0xa4, 0x9d, 0xf6, 0xbd, 0x5a, 0xf6, 0x46, 0xfd, 0xc7, 0x9d, 0x48, 0xea,
	0x4c, 0x6a, 0x7f, 0x40, 0x35, 0xf3, 0x47, 0xfd, 0x01, 0x33, 0xb4, 0xef, 0x47, 0x92, 0x8b, 0x12,
	0x78, 0xbc, 0x9b, 0xc8, 0x44, 0xe2, 0xa7, 0x6f, 0xbf, 0x4a, 0xeb, 0xe1, 0x7f, 0x0e, 0x34, 0x5f,
	0x95, 0x81, 0xcf, 0x0c, 0x35, 0x8c, 0x7c, 0x0b, 0xeb, 0x39, 0x55, 0x34, 0xd3, 0xae, 0xd3, 0x75,
	0x7a, 0x5b, 0x47, 0xfb, 0xde, 0xc2, 0x45, 0xde, 0x5b, 0x94, 0x4f, 0x1a, 0xef, 0x2f, 0x3f, 0x59,
	0x09, 0x2a, 0x67, 0xf2, 0x1a, 0x5a, 0x2c, 0x97, 0xd1, 0x30, 0x1c, 0x72, 0x6d, 0xa4, 0x9a, 0xb8,
	0xf7, 0xba, 0xab, 0xbd, 0xad, 0xa3, 0xce, 0x2d, 0xfa, 0x45, 0x75, 0x78, 0x61, 0xbd, 0xab, 0x20,
	0x4d, 0x44, 0x4f, 0x4b, 0x92, 0x9c, 0xc3, 0xbe, 0x91, 0x86, 0xa6, 0x61, 0xc6, 0x85, 0x61, 0x71,
	0x48, 0xa3, 0x48, 0x16, 0xc2, 0x70, 0x91, 0xb8, 0xab, 0x98, 0xd2, 0x23, 0xaf, 0x2c, 0xd5, 0xb3,
	0xa5, 0x7a, 0x55, 0xa9, 0xde, 0x8f, 0x92, 0x8b, 0x2a, 0xde, 0x1e, 0xf2, 0x6f, 0x10, 0x3f, 0xae,
	0xe9, 0xc3, 0xcb, 0x06, 0xac, 0x97, 0xc9, 0x93, 0x03, 0x00, 0x0b, 0x87, 0x31, 0x13, 0x32, 0xc3,
	0x4a, 0x37, 0x83, 0x4d, 0x6b, 0x79, 0x6e, 0x0d, 0xe4, 0x08, 0xf6, 0xa2, 0x42, 0x29, 0x26, 0x4c,
	0xc8, 0xc5, 0x45, 0x4a, 0x0d, 0x97, 0x22, 0x1c, 0xe4, 0xda, 0xbd, 0xd7, 0x75, 0x7a, 0xad, 0x60,
	0xa7, 0x12, 0x5f, 0x4f, 0xb5, 0x93, 0x5c, 0x93, 0xef, 0xc0, 0x35, 0x54, 0x25, 0xcc, 0x84, 0xda,
	0xd0, 0xdf, 0xb9, 0x48, 0x42, 0x65, 0x35, 0xc4, 0x56, 0x11, 0xdb, 0x2b, 0xf5, 0xb3, 0x52, 0x0e,
	0xac, 0x5a, 0x81, 0x19, 0x17, 0x21, 0x15, 0xa2, 0xa0, 0xe9, 0xc2, 0x7d, 0x8d, 0x12, 0xcc, 0xb8,
	0x38, 0x46, 0x79, 0xf1, 0xc6, 0x8c, 0x8e, 0xef, 0x06, 0xd7, 0x2a, 0x90, 0x8e, 0xef, 0x00, 0x5f,
	0x41, 0x37, 0x92, 0x42, 0x1b, 0x6e, 0x0a, 0x6b, 0xb2, 0x4f, 0x4d, 0xc7, 0x0b, 0x01, 0xd6, 0x31,
	0xc0, 0xc1, 0xbc, 0xdf, 0x1b, 0x3a, 0x9e, 0x0b, 0xf4, 0x14, 0x88, 0x62, 0x3a, 0x97, 0x42, 0xf3,
	0x91, 0x1d, 0x22, 0x8d, 0xe8, 0x7d, 0x44, 0xb7, 0xe7, 0x15, 0xeb, 0x7e, 0x0e, 0xfb, 0x55, 0xb2,
	0x8a, 0x5d, 0x30, 0xc5, 0x44, 0xc4, 0x42, 0x5d, 0xe4, 0x79, 0x3a, 0x71, 0x37, 0x96, 0xec, 0x6c,
	0xc9, 0x07, 0x53, 0xfc, 0x0c, 0x69, 0xf2, 0x05, 0xb4, 0x71, 0x84, 0x74, 0x98, 0x33, 0x15, 0x4e,
	0x18, 0x55, 0xee, 0x66, 0xd7, 0xe9, 0x35, 0x82, 0x72, 0x28, 0xf5, 0x5b, 0xa6, 0x7e, 0x61, 0x54,
	0x91, 0xdf, 0x60, 0x37, 0xe6, 0xda, 0x28, 0x3e, 0xc0, 0x82, 0xc2, 0x77, 0x8c, 0x27, 0x43, 0xa3,
	0x5d, 0xc0, 0xdb, 0x3f, 0xbb, 0x35, 0xac, 0xcf, 0x67, 0x9c, 0xcf, 0x4b, 0xdf, 0x2a, 0x91, 0x9d,
	0xf8, 0xb6, 0x74, 0xf8, 0xb7, 0x03, 0x3b, 0x77, 0x20, 0xe4, 0x1b, 0xd8, 0x1d, 0xd1, 0x94, 0xc7,
	0xd4, 0x48, 0x15, 0x2a, 0xf6, 0x8e, 0xaa, 0x18, 0x1f, 0xca, 0xc1, 0x87, 0x22, 0xb5, 0x16, 0xa0,
	0x64, 0x5f, 0xea, 0x53, 0x68, 0x1a, 0xc5, 0xa8, 0x2e, 0xd4, 0x64, 0x66, 0xee, 0xb6, 0xa6, 0x36,
	0xeb, 0xf2, 0x39, 0x3c, 0xb0, 0x2b, 0xcc, 0xa2, 0xba, 0x65, 0xe5, 0x94, 0xb5, 0x6e, 0xac, 0xd6,
	0xed, 0x11, 0x6c, 0x0c, 0x0a, 0x35, 0x3b, 0x4d, 0xf7, 0xed, 0xd9, 0x4a, 0x4f, 0xa0, 0xc5, 0x22,
	0xa9, 0x27, 0xda, 0xb0, 0x6c, 0x66, 0x68, 0x9a, 0xb5, 0xf1, 0x24, 0xd7, 0x87, 0x7f, 0xae, 0x41,
	0x6b, 0x6e, 0x67, 0xc9, 0x2e, 0xac, 0xe1, 0xab, 0x62, 0xfa, 0x8d, 0xa0, 0x3c, 0x90, 0xaf, 0x60,
	0xfb, 0xf6, 0xdc, 0x97, 0x69, 0xb7, 0xf5, 0xc2, 0xc4, 0x3f, 0x81, 0xd6, 0xfc, 0xb0, 0x95, 0x99,
	0x37, 0xf9, 0xec, 0x6c, 0x9d, 0x42, 0x7b, 0xda, 0x8b, 0x90, 0x66, 0x76, 0x87, 0x31, 0xff, 0x25,
	0x86, 0xe4, 0xc1, 0x94, 0x3b, 0x46, 0x8c, 0xfc, 0x04, 0x0f, 0x17, 0x9f, 0x1f, 0x4b, 0x5d, 0x22,
	0x54, 0x7b, 0xa1, 0x37, 0xe4, 0x7b, 0xd8, 0x98, 0x36, 0x01, 0x57, 0x64, 0x89, 0x18, 0x35, 0x60,
	0x4b, 0x9a, 0x69, 0xd9, 0x45, 0x21, 0x62, 0xdc, 0x95, 0x65, 0x4a, 0xba, 0xe1, 0x5e, 0x16, 0x22,
	0x26, 0xcf, 0xa0, 0x61, 0xbb, 0xb8, 0xec, 0xda, 0xa0, 0x33, 0xf9, 0x01, 0x36, 0xeb, 0xd6, 0xe2,
	0x7e, 0x2c, 0x41, 0xde, 0x10, 0xe4, 0x67, 0x20, 0x4a, 0x16, 0x22, 0xc6, 0x16, 0xb3, 0x8c, 0x72,
	0x11, 0x33, 0x55, 0xad, 0xce, 0x47, 0xe3, 0x6c, 0x4f, 0xd1, 0x60, 0x4a, 0x92, 0x2f, 0xe1, 0xe1,
	0x05, 0x17, 0x34, 0xe5, 0x7f, 0xb0, 0x38, 0x1c, 0xe2, 0xaa, 0xb8, 0x5b, 0x5d, 0xa7, 0xb7, 0x1a,
	0xb4, 0x6b, 0xfb, 0x29, 0x9a, 0x4f, 0x5e, 0xbe, 0xbf, 0xea, 0x38, 0x1f, 0xae, 0x3a, 0xce, 0xbf,
	0x57, 0x1d, 0xe7, 0xaf, 0xeb, 0xce, 0xca, 0x87, 0xeb, 0xce, 0xca, 0x3f, 0xd7, 0x9d, 0x95, 0x5f,
	0xbf, 0x4e, 0xb8, 0x19, 0x16, 0x03, 0x2f, 0x92, 0x99, 0xaf, 0xe5, 0x88, 0x29, 0xc6, 0x13, 0xf1,
	0x34, 0xed, 0xfb, 0x69, 0xdf, 0x1f, 0xcf, 0xfc, 0x3f, 0xcd, 0x24, 0x67, 0x7a, 0xb0, 0x8e, 0x3f,
	0xbd, 0x67, 0xff, 0x07, 0x00, 0x00, 0xff, 0xff, 0x9a, 0xa9, 0xff, 0xb8, 0x5c, 0x07, 0x00, 0x00,
}

func (m *GenesisState) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GenesisState) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *GenesisState) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.TotalMintedAccounting.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x1a
	if len(m.EpochHistory) > 0 {
		for iNdEx := len(m.EpochHistory) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.EpochHistory[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
		}
	}
	{
		size, err := m.Params.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *Params) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Params) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Params) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.DistributionWeights.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x52
	if m.EpochsPerYear != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.EpochsPerYear))
		i--
		dAtA[i] = 0x48
	}
	{
		size, err := m.AnnualReferenceSupply.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x42
	if m.ResponsivenessBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.ResponsivenessBps))
		i--
		dAtA[i] = 0x38
	}
	if m.ConstitutionalMaxInflationBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.ConstitutionalMaxInflationBps))
		i--
		dAtA[i] = 0x30
	}
	if m.MaxAnnualInflationBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.MaxAnnualInflationBps))
		i--
		dAtA[i] = 0x28
	}
	if m.MinAnnualInflationBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.MinAnnualInflationBps))
		i--
		dAtA[i] = 0x20
	}
	if m.TargetStakingRatioBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.TargetStakingRatioBps))
		i--
		dAtA[i] = 0x18
	}
	if m.CurrentInflationBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.CurrentInflationBps))
		i--
		dAtA[i] = 0x10
	}
	if len(m.BaseDenom) > 0 {
		i -= len(m.BaseDenom)
		copy(dAtA[i:], m.BaseDenom)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.BaseDenom)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *DistributionWeights) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *DistributionWeights) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *DistributionWeights) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.EcosystemBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.EcosystemBps))
		i--
		dAtA[i] = 0x28
	}
	if m.BurnBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.BurnBps))
		i--
		dAtA[i] = 0x20
	}
	if m.ProtectionBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.ProtectionBps))
		i--
		dAtA[i] = 0x18
	}
	if m.TreasuryBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.TreasuryBps))
		i--
		dAtA[i] = 0x10
	}
	if m.ValidatorRewardBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.ValidatorRewardBps))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *EmissionEpoch) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *EmissionEpoch) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *EmissionEpoch) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.FinalizedHeight != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.FinalizedHeight))
		i--
		dAtA[i] = 0x58
	}
	{
		size, err := m.RoundingRemainder.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x52
	{
		size, err := m.Ecosystem.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x4a
	{
		size, err := m.Burn.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x42
	{
		size, err := m.ProtectionFund.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x3a
	{
		size, err := m.Treasury.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x32
	{
		size, err := m.ValidatorReward.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x2a
	{
		size, err := m.EmissionAmount.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x22
	if m.InflationBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.InflationBps))
		i--
		dAtA[i] = 0x18
	}
	if m.StakingRatioBps != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.StakingRatioBps))
		i--
		dAtA[i] = 0x10
	}
	if m.Epoch != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.Epoch))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintGenesis(dAtA []byte, offset int, v uint64) int {
	offset -= sovGenesis(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *GenesisState) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Params.Size()
	n += 1 + l + sovGenesis(uint64(l))
	if len(m.EpochHistory) > 0 {
		for _, e := range m.EpochHistory {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	l = m.TotalMintedAccounting.Size()
	n += 1 + l + sovGenesis(uint64(l))
	return n
}

func (m *Params) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.BaseDenom)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	if m.CurrentInflationBps != 0 {
		n += 1 + sovGenesis(uint64(m.CurrentInflationBps))
	}
	if m.TargetStakingRatioBps != 0 {
		n += 1 + sovGenesis(uint64(m.TargetStakingRatioBps))
	}
	if m.MinAnnualInflationBps != 0 {
		n += 1 + sovGenesis(uint64(m.MinAnnualInflationBps))
	}
	if m.MaxAnnualInflationBps != 0 {
		n += 1 + sovGenesis(uint64(m.MaxAnnualInflationBps))
	}
	if m.ConstitutionalMaxInflationBps != 0 {
		n += 1 + sovGenesis(uint64(m.ConstitutionalMaxInflationBps))
	}
	if m.ResponsivenessBps != 0 {
		n += 1 + sovGenesis(uint64(m.ResponsivenessBps))
	}
	l = m.AnnualReferenceSupply.Size()
	n += 1 + l + sovGenesis(uint64(l))
	if m.EpochsPerYear != 0 {
		n += 1 + sovGenesis(uint64(m.EpochsPerYear))
	}
	l = m.DistributionWeights.Size()
	n += 1 + l + sovGenesis(uint64(l))
	return n
}

func (m *DistributionWeights) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.ValidatorRewardBps != 0 {
		n += 1 + sovGenesis(uint64(m.ValidatorRewardBps))
	}
	if m.TreasuryBps != 0 {
		n += 1 + sovGenesis(uint64(m.TreasuryBps))
	}
	if m.ProtectionBps != 0 {
		n += 1 + sovGenesis(uint64(m.ProtectionBps))
	}
	if m.BurnBps != 0 {
		n += 1 + sovGenesis(uint64(m.BurnBps))
	}
	if m.EcosystemBps != 0 {
		n += 1 + sovGenesis(uint64(m.EcosystemBps))
	}
	return n
}

func (m *EmissionEpoch) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Epoch != 0 {
		n += 1 + sovGenesis(uint64(m.Epoch))
	}
	if m.StakingRatioBps != 0 {
		n += 1 + sovGenesis(uint64(m.StakingRatioBps))
	}
	if m.InflationBps != 0 {
		n += 1 + sovGenesis(uint64(m.InflationBps))
	}
	l = m.EmissionAmount.Size()
	n += 1 + l + sovGenesis(uint64(l))
	l = m.ValidatorReward.Size()
	n += 1 + l + sovGenesis(uint64(l))
	l = m.Treasury.Size()
	n += 1 + l + sovGenesis(uint64(l))
	l = m.ProtectionFund.Size()
	n += 1 + l + sovGenesis(uint64(l))
	l = m.Burn.Size()
	n += 1 + l + sovGenesis(uint64(l))
	l = m.Ecosystem.Size()
	n += 1 + l + sovGenesis(uint64(l))
	l = m.RoundingRemainder.Size()
	n += 1 + l + sovGenesis(uint64(l))
	if m.FinalizedHeight != 0 {
		n += 1 + sovGenesis(uint64(m.FinalizedHeight))
	}
	return n
}

func sovGenesis(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozGenesis(x uint64) (n int) {
	return sovGenesis(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *GenesisState) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
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
			return fmt.Errorf("proto: GenesisState: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GenesisState: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Params", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
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
				return fmt.Errorf("proto: wrong wireType = %d for field EpochHistory", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.EpochHistory = append(m.EpochHistory, EmissionEpoch{})
			if err := m.EpochHistory[len(m.EpochHistory)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TotalMintedAccounting", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.TotalMintedAccounting.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
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
func (m *Params) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
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
			return fmt.Errorf("proto: Params: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Params: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BaseDenom", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.BaseDenom = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field CurrentInflationBps", wireType)
			}
			m.CurrentInflationBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field TargetStakingRatioBps", wireType)
			}
			m.TargetStakingRatioBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.TargetStakingRatioBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MinAnnualInflationBps", wireType)
			}
			m.MinAnnualInflationBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MinAnnualInflationBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxAnnualInflationBps", wireType)
			}
			m.MaxAnnualInflationBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxAnnualInflationBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 6:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ConstitutionalMaxInflationBps", wireType)
			}
			m.ConstitutionalMaxInflationBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ConstitutionalMaxInflationBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 7:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ResponsivenessBps", wireType)
			}
			m.ResponsivenessBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ResponsivenessBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 8:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AnnualReferenceSupply", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.AnnualReferenceSupply.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 9:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field EpochsPerYear", wireType)
			}
			m.EpochsPerYear = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.EpochsPerYear |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 10:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DistributionWeights", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.DistributionWeights.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
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
func (m *DistributionWeights) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
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
			return fmt.Errorf("proto: DistributionWeights: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: DistributionWeights: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ValidatorRewardBps", wireType)
			}
			m.ValidatorRewardBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ValidatorRewardBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field TreasuryBps", wireType)
			}
			m.TreasuryBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.TreasuryBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ProtectionBps", wireType)
			}
			m.ProtectionBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ProtectionBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field BurnBps", wireType)
			}
			m.BurnBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.BurnBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field EcosystemBps", wireType)
			}
			m.EcosystemBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.EcosystemBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
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
func (m *EmissionEpoch) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
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
			return fmt.Errorf("proto: EmissionEpoch: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: EmissionEpoch: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Epoch", wireType)
			}
			m.Epoch = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field StakingRatioBps", wireType)
			}
			m.StakingRatioBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.StakingRatioBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field InflationBps", wireType)
			}
			m.InflationBps = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.InflationBps |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field EmissionAmount", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.EmissionAmount.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ValidatorReward", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.ValidatorReward.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Treasury", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Treasury.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ProtectionFund", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.ProtectionFund.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 8:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Burn", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Burn.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 9:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Ecosystem", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Ecosystem.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 10:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RoundingRemainder", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.RoundingRemainder.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 11:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field FinalizedHeight", wireType)
			}
			m.FinalizedHeight = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.FinalizedHeight |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
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
func skipGenesis(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowGenesis
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
					return 0, ErrIntOverflowGenesis
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
					return 0, ErrIntOverflowGenesis
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
				return 0, ErrInvalidLengthGenesis
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupGenesis
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthGenesis
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthGenesis		= fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowGenesis		= fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupGenesis	= fmt.Errorf("proto: unexpected end of group")
)
