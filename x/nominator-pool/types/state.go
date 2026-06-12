package types

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
	appparams "github.com/sovereign-l1/l1/app/params"
	"github.com/sovereign-l1/l1/x/internal/prototype"
	validatorregistrytypes "github.com/sovereign-l1/l1/x/validator-registry/types"
)

const (
	PoolStatusActive	= "active"
	PoolStatusPaused	= "paused"
	PoolStatusFrozenLimited	= "frozen_limited"
	PoolStatusClosed	= "closed"

	WithdrawalStatusPending		= "pending"
	WithdrawalStatusCancelled	= "cancelled"
	WithdrawalStatusCompleted	= "completed"

	MaxPoolsV1				= uint32(10_000)
	MaxDelegatorsV1				= uint32(1_000_000)
	MaxPendingDepositsV1			= uint32(1_000_000)
	MaxPendingWithdrawalsV1			= uint32(1_000_000)
	MaxUnbondingEntriesV1			= uint32(1_000_000)
	MaxPoolIDBytesV1			= uint32(96)
	MaxBasisPoints				= uint32(10_000)
	IndexScale				= uint64(1_000_000_000)
	SecondsPerDay				= uint64(24 * 60 * 60)
	DefaultBaseDenom			= appparams.BaseDenom
	DefaultDisplayDenom			= appparams.DisplayDenom
	DefaultDisplayExponent			= appparams.DisplayDenomExponent
	DefaultAETBaseUnits			= uint64(appparams.BaseUnitsPerDisplay)
	DefaultMaxCommissionBps			= uint32(2_000)
	DefaultMaxValidatorCommissionBps	= uint32(2_000)
	DefaultMaxOperatorPerformanceBonusBps	= uint32(1_000)
	DefaultUnbondingDays			= uint64(18)
	DefaultUnbondingBlocks			= DefaultUnbondingDays * SecondsPerDay / appparams.StakingUnbondingBlockTimeSeconds
	DefaultValidatorChangeDelay		= uint64(100)
	DefaultMinValidatorStake		= uint64(1_000_000) * DefaultAETBaseUnits
	DefaultSoloValidatorMinSelfStake	= uint64(1_000_000) * DefaultAETBaseUnits
	DefaultPoolBackedMinSelfStake		= uint64(400_000) * DefaultAETBaseUnits
	DefaultPoolBackedMaxNominatorStake	= uint64(600_000) * DefaultAETBaseUnits
	DefaultMinPoolDeposit			= uint64(10) * DefaultAETBaseUnits
	DefaultMinTxFeeBaseUnits		= uint64(3_000_000)
	DefaultRewardEpochDurationBlocks	= SecondsPerDay / appparams.StakingUnbondingBlockTimeSeconds
	DefaultValidatorPowerCapBps		= uint32(300)
	DefaultMinPoolValidatorAllocationBps	= uint32(25)
	DefaultMaxPoolValidatorAllocationBps	= uint32(300)
)

type Params struct {
	Authority				string
	BaseDenom				string
	DisplayDenom				string
	DisplayExponent				uint32
	MaxPools				uint32
	MaxDelegators				uint32
	MaxPendingDeposits			uint32
	MaxPendingWithdrawals			uint32
	MaxUnbondingEntries			uint32
	MaxPoolIDBytes				uint32
	MinValidatorStake			uint64
	SoloValidatorMinSelfStake		uint64
	PoolBackedValidatorMinSelfStake		uint64
	PoolBackedValidatorMaxNominatorStake	uint64
	ValidatorSelfStakeMinRatioBps		uint32
	ValidatorNominatorStakeMaxRatioBps	uint32
	GovernanceMinValidatorCount		uint32
	TargetValidatorCount			uint32
	MaxValidatorCount			uint32
	GovernanceMaxValidatorCount		uint32
	MaxCommissionBps			uint32
	MaxValidatorCommissionBps		uint32
	MaxOperatorPerformanceBonusBps		uint32
	ValidatorCommissionFloorBps		uint32
	DefaultValidatorCommissionBps		uint32
	ValidatorCommissionCeilingBps		uint32
	ValidatorCommissionMaxDailyChangeBps	uint32
	ValidatorPowerCapBps			uint32
	ValidatorPowerCapSchedule		[]ValidatorPowerCapPhase
	DowntimeSlashBps			uint32
	DoubleSignSlashBps			uint32
	DoubleSignTombstone			bool
	OverflowRewardMultiplierMinBps		uint32
	OverflowRewardMultiplierMaxBps		uint32
	UnbondingBlocks				uint64
	RewardEpochDurationBlocks		uint64
	BaseRewardRateBps			uint32
	MaxRewardRateBps			uint32
	ValidatorChangeDelay			uint64
	MinPoolDeposit				uint64
	DirectUserValidatorDelegationEnabled	bool
	DirectUserDelegationEnabled		bool
	ReputationStakeWeightBps		uint32
	PoolReceiptDenomOrCodeID		string
	MaxPoolValidatorAllocationBps		uint32
	MinPoolValidatorAllocationBps		uint32
	AllocationRebalanceEpochs		uint64
	AllocationUptimeWeight			uint32
	AllocationCommissionWeight		uint32
	AllocationReputationWeight		uint32
	AllocationStakeEfficiencyWeight		uint32
	AllocationSlashingRiskWeight		uint32
	AllocationNetworkLoadWeight		uint32
	PoolProtocolFeeBps			uint32
	ValidatorOperatorBonusBps		uint32
	ValidatorInfrastructureCostModel	string
	BurnFeeShareBps				uint32
	RewardFeeShareBps			uint32
	TreasuryFeeShareBps			uint32
	InflationMinBps				uint32
	InitialInflationBps			uint32
	InflationMaxBps				uint32
	TargetBondedRatioBps			uint32
	AprTargetMinBps				uint32
	AprTargetMaxBps				uint32
	MinTxFeeBaseUnits			uint64
	FeeDenom				string
	StorageRentRatePerByteSecond		uint64
	SystemStorageReserveMinRunwayDays	uint64
	SystemStorageReserveWarningRunwayDays	uint64
	SystemStorageReserveCriticalRunwayDays	uint64
}

type ValidatorPowerCapPhase struct {
	MaxValidatorCount	uint32
	PowerCapBps		uint32
}

type ValidatorFundingMode string

const (
	ValidatorFundingSolo		ValidatorFundingMode	= "solo"
	ValidatorFundingPoolBacked	ValidatorFundingMode	= "pool_backed"
)

type ValidatorFunding struct {
	Mode		ValidatorFundingMode
	SelfStake	uint64
	NominatorStake	uint64
}

type ValidatorPolicyCandidate struct {
	ValidatorAddress	string
	ReputationScore		uint32
	UptimeBps		uint32
	UptimeWindow		uint64
	MissedBlocks		uint64
	CommissionBps		uint32
	StakeEfficiencyBps	uint32
	SlashingRiskBps		uint32
	NetworkLoadBps		uint32
	CurrentAllocationBps	uint32
	AllocationLimitBps	uint32
	OperationalHistoryBps	uint32
	Jailed			bool
	Slashed			bool
}

type AllocationWeight struct {
	ValidatorAddress	string
	Score			uint64
	WeightBps		uint32
}

type ValidatorScore struct {
	Eligible		bool
	OverallScoreBps		uint32
	SlashingRiskScoreBps	uint32
	UptimeScoreBps		uint32
	CommissionScoreBps	uint32
	ReputationScoreBps	uint32
	StakeEfficiencyScoreBps	uint32
	NetworkLoadScoreBps	uint32
}

type PoolStateMetadata struct {
	OwnerRaw		string
	PoolContractAddressRaw	string
	TouchedKeys		[]string
}

type StakingPoolDepositReceipt struct {
	PoolID			string
	OwnerAddress		string
	PoolContractAddressUser	string
	ReceiptToken		string
	Amount			uint64
	Shares			uint64
	Height			uint64
	InternalMetadata	PoolStateMetadata
}

type PoolUnbondReceipt struct {
	PoolID			string
	OwnerAddress		string
	RequestID		string
	Shares			uint64
	Amount			uint64
	RequestHeight		uint64
	CompleteHeight		uint64
	InternalMetadata	PoolStateMetadata
}

type PoolWithdrawalReceipt struct {
	PoolID			string
	OwnerAddress		string
	RequestID		string
	Amount			uint64
	Height			uint64
	InternalMetadata	PoolStateMetadata
}

type PoolTopUpReceipt struct {
	PoolID			string
	PayerAddress		string
	Amount			uint64
	StorageDebtPaid		uint64
	Height			uint64
	InternalMetadata	PoolStateMetadata
}

type PoolRewardClaimReceipt struct {
	PoolID			string
	OwnerAddress		string
	Amount			uint64
	Epoch			uint64
	Height			uint64
	InternalMetadata	PoolStateMetadata
}

type StakeReputationClaimReceipt struct {
	Account			string
	PoolID			string
	ReputationDelta		uint64
	ReputationScore		uint64
	Height			uint64
	InternalMetadata	PoolStateMetadata
}

type PoolRebalanceReceipt struct {
	PoolID			string
	Epoch			uint64
	Height			uint64
	Allocations		[]PoolValidatorAllocation
	InternalMetadata	PoolStateMetadata
}

type ValidatorRegistrationReceipt struct {
	Validator	string
	Status		string
	SelfStake	uint64
	PoolStake	uint64
	TouchedKeys	[]string
}

type State struct {
	Pools				[]NominatorPool
	Validators			[]Validator
	ValidatorPerformanceScores	[]ValidatorPerformanceScore
	ValidatorCommissions		[]ValidatorCommission
	ValidatorSlashingRisks		[]ValidatorSlashingRisk
	ValidatorAllocationLimits	[]ValidatorAllocationLimit
	LiquidStakingPools		[]LiquidStakingPool
	PoolShares			[]PoolShare
	PoolValidatorAllocations	[]PoolValidatorAllocation
	PoolUnbondingRequests		[]PoolUnbondingRequest
	PoolRewardIndexes		[]PoolRewardIndex
	RewardClaims			[]RewardClaim
	EpochStakingSnapshots		[]EpochStakingSnapshot
	ValidatorSetSnapshots		[]ValidatorSetSnapshot
	ValidatorSlashEvents		[]ValidatorSlashEvent
}

type NominatorPool struct {
	PoolID				string
	ContractAddressUser		string
	ContractAddressRaw		string
	OfficialLiquidStaking		bool
	PoolOperator			string
	ValidatorTarget			string
	PendingValidatorTarget		string
	ValidatorChangeHeight		uint64
	TotalShares			uint64
	TotalBondedStake		uint64
	Allocations			[]PoolAllocation
	PendingDeposits			[]PendingDeposit
	PendingWithdrawals		[]PendingWithdrawal
	DelegatorShares			[]DelegatorShare
	RewardIndex			uint64
	RewardRemainder			uint64
	SlashIndex			uint64
	PoolCommissionBps		uint32
	RewardEpoch			uint64
	ProtocolFeeAccrued		uint64
	ValidatorCommissionAccrued	uint64
	ValidatorOperatorIncome		[]ValidatorIncome
	ValidatorAllocations		[]ValidatorRewardAllocation
	Status				string
	UnbondingQueue			[]UnbondingEntry
}

type PendingDeposit struct {
	Delegator	string
	Amount		uint64
	Height		uint64
}

type PendingWithdrawal struct {
	WithdrawalID	string
	Delegator	string
	Shares		uint64
	Amount		uint64
	RequestHeight	uint64
	CompleteHeight	uint64
	Status		string
}

type DelegatorShare struct {
	Delegator		string	`protobuf:"bytes,1,opt,name=delegator,proto3" json:"delegator,omitempty"`
	Shares			uint64	`protobuf:"varint,2,opt,name=shares,proto3" json:"shares,omitempty"`
	RewardIndexCheckpoint	uint64	`protobuf:"varint,3,opt,name=reward_index_checkpoint,json=rewardIndexCheckpoint,proto3" json:"reward_index_checkpoint,omitempty"`
	PendingRewards		uint64	`protobuf:"varint,4,opt,name=pending_rewards,json=pendingRewards,proto3" json:"pending_rewards,omitempty"`
	SlashIndexCheckpoint	uint64	`protobuf:"varint,5,opt,name=slash_index_checkpoint,json=slashIndexCheckpoint,proto3" json:"slash_index_checkpoint,omitempty"`
}

type UnbondingEntry struct {
	WithdrawalID	string
	Delegator	string
	Amount		uint64
	CompleteHeight	uint64
	Status		string
}

type PoolAllocation struct {
	ValidatorAddress	string
	Amount			uint64
	Height			uint64
}

type ValidatorRewardAllocation struct {
	Validator			string	`protobuf:"bytes,1,opt,name=validator,proto3" json:"validator,omitempty"`
	PoolAllocatedStake		uint64	`protobuf:"varint,2,opt,name=pool_allocated_stake,json=poolAllocatedStake,proto3" json:"pool_allocated_stake,omitempty"`
	ValidatorSelfStake		uint64	`protobuf:"varint,3,opt,name=validator_self_stake,json=validatorSelfStake,proto3" json:"validator_self_stake,omitempty"`
	PerformanceBps			uint32	`protobuf:"varint,4,opt,name=performance_bps,json=performanceBps,proto3" json:"performance_bps,omitempty"`
	CommissionBps			uint32	`protobuf:"varint,5,opt,name=commission_bps,json=commissionBps,proto3" json:"commission_bps,omitempty"`
	SlashingLoss			uint64	`protobuf:"varint,6,opt,name=slashing_loss,json=slashingLoss,proto3" json:"slashing_loss,omitempty"`
	Jailed				bool	`protobuf:"varint,7,opt,name=jailed,proto3" json:"jailed,omitempty"`
	InfrastructureCost		uint64	`protobuf:"varint,8,opt,name=infrastructure_cost,json=infrastructureCost,proto3" json:"infrastructure_cost,omitempty"`
	OperatorPerformanceBonusBps	uint32	`protobuf:"varint,9,opt,name=operator_performance_bonus_bps,json=operatorPerformanceBonusBps,proto3" json:"operator_performance_bonus_bps,omitempty"`
	GrossPoolRewards		uint64	`protobuf:"varint,10,opt,name=gross_pool_rewards,json=grossPoolRewards,proto3" json:"gross_pool_rewards,omitempty"`
	ValidatorCommission		uint64	`protobuf:"varint,11,opt,name=validator_commission,json=validatorCommission,proto3" json:"validator_commission,omitempty"`
	PoolProtocolFee			uint64	`protobuf:"varint,12,opt,name=pool_protocol_fee,json=poolProtocolFee,proto3" json:"pool_protocol_fee,omitempty"`
	NetPoolRewards			uint64	`protobuf:"varint,13,opt,name=net_pool_rewards,json=netPoolRewards,proto3" json:"net_pool_rewards,omitempty"`
	ValidatorSelfStakeRewards	uint64	`protobuf:"varint,14,opt,name=validator_self_stake_rewards,json=validatorSelfStakeRewards,proto3" json:"validator_self_stake_rewards,omitempty"`
	OperatorPerformanceBonus	uint64	`protobuf:"varint,15,opt,name=operator_performance_bonus,json=operatorPerformanceBonus,proto3" json:"operator_performance_bonus,omitempty"`
	ValidatorGrossIncome		uint64	`protobuf:"varint,16,opt,name=validator_gross_income,json=validatorGrossIncome,proto3" json:"validator_gross_income,omitempty"`
	ValidatorNetIncome		int64	`protobuf:"varint,17,opt,name=validator_net_income,json=validatorNetIncome,proto3" json:"validator_net_income,omitempty"`
	RewardIndexDelta		uint64	`protobuf:"varint,18,opt,name=reward_index_delta,json=rewardIndexDelta,proto3" json:"reward_index_delta,omitempty"`
	RewardIndexAfter		uint64	`protobuf:"varint,19,opt,name=reward_index_after,json=rewardIndexAfter,proto3" json:"reward_index_after,omitempty"`
}

type ValidatorIncome struct {
	Validator			string
	SelfStakeRewards		uint64
	CommissionIncome		uint64
	OperatorPerformanceBonus	uint64
	InfrastructureCost		uint64
	GrossIncome			uint64
	NetIncome			int64
}

type PoolRewardSummary struct {
	PoolID				string
	Epoch				uint64
	RewardRateBps			uint32
	EmissionsAllocated		uint64
	FeesAllocated			uint64
	RewardCap			uint64
	GrossPoolRewards		uint64
	ValidatorCommission		uint64
	PoolProtocolFee			uint64
	PoolUserRewards			uint64
	SlashingLosses			uint64
	ValidatorSelfStakeRewards	uint64
	OperatorPerformanceBonus	uint64
	ValidatorGrossIncome		uint64
	ValidatorNetIncome		int64
	RewardIndexBefore		uint64
	RewardIndexAfter		uint64
	RewardRemainder			uint64
	AllocationsTouched		uint64
}

type MsgCreateNominatorPool struct {
	Authority		string
	PoolID			string
	PoolOperator		string
	ValidatorTarget		string
	PoolCommissionBps	uint32
	Height			uint64
	ValidatorStatus		string
}

type MsgDepositToPool struct {
	Authority	string
	PoolID		string
	Delegator	string
	Amount		uint64
	Height		uint64
}

type MsgCreateOfficialLiquidStakingPool struct {
	Authority		string	`protobuf:"bytes,1,opt,name=authority,proto3" json:"authority,omitempty"`
	PoolID			string	`protobuf:"bytes,2,opt,name=pool_id,json=poolId,proto3" json:"pool_id,omitempty"`
	ContractAddressUser	string	`protobuf:"bytes,3,opt,name=contract_address_user,json=contractAddressUser,proto3" json:"contract_address_user,omitempty"`
	ContractAddressRaw	string	`protobuf:"bytes,4,opt,name=contract_address_raw,json=contractAddressRaw,proto3" json:"contract_address_raw,omitempty"`
	PoolOperator		string	`protobuf:"bytes,5,opt,name=pool_operator,json=poolOperator,proto3" json:"pool_operator,omitempty"`
	PoolCommissionBps	uint32	`protobuf:"varint,6,opt,name=pool_commission_bps,json=poolCommissionBps,proto3" json:"pool_commission_bps,omitempty"`
	Height			uint64	`protobuf:"varint,7,opt,name=height,proto3" json:"height,omitempty"`
}

type MsgDepositToOfficialLiquidStaking struct {
	Authority		string
	PoolID			string
	UserAddress		string
	Amount			uint64
	Height			uint64
	ValidatorAddress	string
}

type MsgDepositToStakingPool struct {
	PoolID			string	`protobuf:"bytes,1,opt,name=pool_id,json=poolid,proto3" json:"pool_id,omitempty"`
	WalletAddress		string	`protobuf:"bytes,2,opt,name=wallet_address,json=walletaddress,proto3" json:"wallet_address,omitempty"`
	Amount			uint64	`protobuf:"varint,3,opt,name=amount,proto3" json:"amount,omitempty"`
	Height			uint64	`protobuf:"varint,4,opt,name=height,proto3" json:"height,omitempty"`
	ReservedRouting		string	`protobuf:"bytes,5,opt,name=reserved_routing,json=reservedRouting,proto3" json:"reserved_routing,omitempty"`
	OfficialContract	string	`protobuf:"bytes,6,opt,name=official_contract,json=officialcontract,proto3" json:"official_contract,omitempty"`
}

type MsgDelegateToValidator struct {
	Authority		string	`protobuf:"bytes,1,opt,name=authority,proto3" json:"authority,omitempty"`
	UserAddress		string	`protobuf:"bytes,2,opt,name=user_address,json=userAddress,proto3" json:"user_address,omitempty"`
	ValidatorAddress	string	`protobuf:"bytes,3,opt,name=validator_address,json=validatorAddress,proto3" json:"validator_address,omitempty"`
	Amount			uint64	`protobuf:"varint,4,opt,name=amount,proto3" json:"amount,omitempty"`
	Height			uint64	`protobuf:"varint,5,opt,name=height,proto3" json:"height,omitempty"`
}

type MsgInjectPooledStake struct {
	CallerContractUser	string
	PoolID			string
	ValidatorAddress	string
	Amount			uint64
	Height			uint64
}

type MsgInjectPoolStake struct {
	CallerContractUser	string
	PoolID			string
	Allocations		[]PoolAllocation
	Height			uint64
}

type MsgWithdrawPoolStake struct {
	CallerContractUser	string	`protobuf:"bytes,1,opt,name=caller_contract_user,json=callerContractUser,proto3" json:"caller_contract_user,omitempty"`
	PoolID			string	`protobuf:"bytes,2,opt,name=pool_id,json=poolId,proto3" json:"pool_id,omitempty"`
	OwnerAddress		string	`protobuf:"bytes,3,opt,name=owner_address,json=ownerAddress,proto3" json:"owner_address,omitempty"`
	RequestID		string	`protobuf:"bytes,4,opt,name=request_id,json=requestId,proto3" json:"request_id,omitempty"`
	Height			uint64	`protobuf:"varint,5,opt,name=height,proto3" json:"height,omitempty"`
}

type MsgRebalancePoolAllocations struct {
	CallerContractUser	string
	PoolID			string
	Epoch			uint64
	Height			uint64
	Candidates		[]ValidatorPolicyCandidate
}

type MsgSetOfficialLiquidStakingContract struct {
	Authority		string
	PoolID			string
	ContractAddressUser	string
	ContractAddressRaw	string
	Height			uint64
}

type MsgUpdateParams struct {
	Authority	string
	Params		Params
	Height		uint64
}

type MsgApplyValidatorSlash struct {
	Authority		string
	ValidatorAddress	string
	Fault			string
	Epoch			uint64
	Height			uint64
}

type MsgRequestPoolWithdrawal struct {
	Authority	string
	PoolID		string
	WithdrawalID	string
	Delegator	string
	Shares		uint64
	Height		uint64
}

type MsgRequestPoolUnbond struct {
	PoolID		string	`protobuf:"bytes,1,opt,name=pool_id,json=poolId,proto3" json:"pool_id,omitempty"`
	OwnerAddress	string	`protobuf:"bytes,2,opt,name=owner_address,json=ownerAddress,proto3" json:"owner_address,omitempty"`
	RequestID	string	`protobuf:"bytes,3,opt,name=request_id,json=requestId,proto3" json:"request_id,omitempty"`
	Shares		uint64	`protobuf:"varint,4,opt,name=shares,proto3" json:"shares,omitempty"`
	Height		uint64	`protobuf:"varint,5,opt,name=height,proto3" json:"height,omitempty"`
}

type MsgTopUpPoolReserve struct {
	PoolID		string	`protobuf:"bytes,1,opt,name=pool_id,json=poolId,proto3" json:"pool_id,omitempty"`
	PayerAddress	string	`protobuf:"bytes,2,opt,name=payer_address,json=payerAddress,proto3" json:"payer_address,omitempty"`
	Amount		uint64	`protobuf:"varint,3,opt,name=amount,proto3" json:"amount,omitempty"`
	Height		uint64	`protobuf:"varint,4,opt,name=height,proto3" json:"height,omitempty"`
}

type MsgCancelPoolWithdrawal struct {
	Authority	string
	PoolID		string
	WithdrawalID	string
	Delegator	string
	Height		uint64
}

type MsgClaimPoolRewards struct {
	Authority	string	`protobuf:"bytes,1,opt,name=authority,proto3" json:"authority,omitempty"`
	PoolID		string	`protobuf:"bytes,2,opt,name=pool_id,json=poolId,proto3" json:"pool_id,omitempty"`
	Delegator	string	`protobuf:"bytes,3,opt,name=delegator,proto3" json:"delegator,omitempty"`
	OwnerAddress	string	`protobuf:"bytes,4,opt,name=owner_address,json=ownerAddress,proto3" json:"owner_address,omitempty"`
	Height		uint64	`protobuf:"varint,5,opt,name=height,proto3" json:"height,omitempty"`
}

type MsgClaimStakeReputation struct {
	PoolID		string	`protobuf:"bytes,1,opt,name=pool_id,json=poolId,proto3" json:"pool_id,omitempty"`
	OwnerAddress	string	`protobuf:"bytes,2,opt,name=owner_address,json=ownerAddress,proto3" json:"owner_address,omitempty"`
	Height		uint64	`protobuf:"varint,3,opt,name=height,proto3" json:"height,omitempty"`
}

type MsgSyncPoolRewards struct {
	Authority		string
	PoolID			string
	Epoch			uint64
	RewardRateBps		uint32
	EmissionsAllocated	uint64
	FeesAllocated		uint64
	Height			uint64
	Allocations		[]ValidatorRewardAllocation
}

type MsgClaimStakingRewards struct {
	Authority		string
	Delegator		string
	Validator		string
	Height			uint64
	InternalMigration	bool
}

type MsgUpdatePoolCommission struct {
	Authority		string
	PoolID			string
	PoolOperator		string
	PoolCommissionBps	uint32
	Height			uint64
}

type MsgChangePoolValidator struct {
	Authority	string
	PoolID		string
	PoolOperator	string
	ValidatorTarget	string
	ValidatorStatus	string
	Height		uint64
}

type MsgRegisterValidator struct {
	SignerAddress		string	`protobuf:"bytes,1,opt,name=signer_address,json=signerAddress,proto3" json:"signer_address,omitempty"`
	ValidatorAddress	string	`protobuf:"bytes,2,opt,name=validator_address,json=validatorAddress,proto3" json:"validator_address,omitempty"`
	SelfStake		uint64	`protobuf:"varint,3,opt,name=self_stake,json=selfStake,proto3" json:"self_stake,omitempty"`
	NominatorStake		uint64	`protobuf:"varint,4,opt,name=nominator_stake,json=nominatorStake,proto3" json:"nominator_stake,omitempty"`
	CommissionBps		uint32	`protobuf:"varint,5,opt,name=commission_bps,json=commissionBps,proto3" json:"commission_bps,omitempty"`
	Height			uint64	`protobuf:"varint,6,opt,name=height,proto3" json:"height,omitempty"`
}

type MsgUpdateValidator struct {
	SignerAddress		string	`protobuf:"bytes,1,opt,name=signer_address,json=signerAddress,proto3" json:"signer_address,omitempty"`
	ValidatorAddress	string	`protobuf:"bytes,2,opt,name=validator_address,json=validatorAddress,proto3" json:"validator_address,omitempty"`
	SelfStake		uint64	`protobuf:"varint,3,opt,name=self_stake,json=selfStake,proto3" json:"self_stake,omitempty"`
	NominatorStake		uint64	`protobuf:"varint,4,opt,name=nominator_stake,json=nominatorStake,proto3" json:"nominator_stake,omitempty"`
	PerformanceScore	uint32	`protobuf:"varint,5,opt,name=performance_score,json=performanceScore,proto3" json:"performance_score,omitempty"`
	CommissionBps		uint32	`protobuf:"varint,6,opt,name=commission_bps,json=commissionBps,proto3" json:"commission_bps,omitempty"`
	SlashingRiskBps		uint32	`protobuf:"varint,7,opt,name=slashing_risk_bps,json=slashingRiskBps,proto3" json:"slashing_risk_bps,omitempty"`
	AllocationLimitBps	uint32	`protobuf:"varint,8,opt,name=allocation_limit_bps,json=allocationLimitBps,proto3" json:"allocation_limit_bps,omitempty"`
	Status			string	`protobuf:"bytes,9,opt,name=status,proto3" json:"status,omitempty"`
	Height			uint64	`protobuf:"varint,10,opt,name=height,proto3" json:"height,omitempty"`
}

type MsgUpdateStakingParams struct {
	Authority	string	`protobuf:"bytes,1,opt,name=authority,proto3" json:"authority,omitempty"`
	Params		Params	`protobuf:"bytes,2,opt,name=params,proto3" json:"params"`
	Height		uint64	`protobuf:"varint,3,opt,name=height,proto3" json:"height,omitempty"`
}

type QueryPoolShareRequest struct {
	PoolID		string	`protobuf:"bytes,1,opt,name=pool_id,json=poolId,proto3" json:"pool_id,omitempty"`
	Delegator	string	`protobuf:"bytes,2,opt,name=delegator,proto3" json:"delegator,omitempty"`
}

type QueryNominatorPoolRequest struct {
	PoolID string `protobuf:"bytes,1,opt,name=pool_id,json=poolId,proto3" json:"pool_id,omitempty"`
}

type QueryNominatorPoolResponse struct {
	Pool NominatorPool `protobuf:"bytes,1,opt,name=pool,proto3" json:"pool"`
}

type QueryNominatorPoolsRequest struct {
	Offset	uint64	`protobuf:"varint,1,opt,name=offset,proto3" json:"offset,omitempty"`
	Limit	uint64	`protobuf:"varint,2,opt,name=limit,proto3" json:"limit,omitempty"`
}

type QueryNominatorPoolsResponse struct {
	Pools		[]NominatorPool	`protobuf:"bytes,1,rep,name=pools,proto3" json:"pools,omitempty"`
	NextOffset	uint64		`protobuf:"varint,2,opt,name=next_offset,json=nextOffset,proto3" json:"next_offset,omitempty"`
	Total		uint64		`protobuf:"varint,3,opt,name=total,proto3" json:"total,omitempty"`
}

type QueryPoolDelegatorRequest struct {
	PoolID		string	`protobuf:"bytes,1,opt,name=pool_id,json=poolId,proto3" json:"pool_id,omitempty"`
	Delegator	string	`protobuf:"bytes,2,opt,name=delegator,proto3" json:"delegator,omitempty"`
}

type QueryPoolDelegatorResponse struct {
	Delegator DelegatorShare `protobuf:"bytes,1,opt,name=delegator,proto3" json:"delegator"`
}

type QueryPoolRewardsRequest struct {
	PoolID		string	`protobuf:"bytes,1,opt,name=pool_id,json=poolId,proto3" json:"pool_id,omitempty"`
	Delegator	string	`protobuf:"bytes,2,opt,name=delegator,proto3" json:"delegator,omitempty"`
}

type QueryPoolRewardsResponse struct {
	RewardAmount uint64 `protobuf:"varint,1,opt,name=reward_amount,json=rewardAmount,proto3" json:"reward_amount,omitempty"`
}

type QueryPoolShareResponse struct {
	Share		DelegatorShare	`protobuf:"bytes,1,opt,name=share,proto3" json:"share"`
	PendingRewards	uint64		`protobuf:"varint,2,opt,name=pending_rewards,json=pendingRewards,proto3" json:"pending_rewards,omitempty"`
}

type QueryPoolAllocationsRequest struct {
	PoolID string `protobuf:"bytes,1,opt,name=pool_id,json=poolId,proto3" json:"pool_id,omitempty"`
}

type QueryPoolAllocationsResponse struct {
	Allocations []ValidatorRewardAllocation `protobuf:"bytes,1,rep,name=allocations,proto3" json:"allocations,omitempty"`
}

// Deprecated: QueryStakeReputationRequest is retained for proto compatibility.
// Wallet-facing stake reputation queries are removed. Use x/reputation instead.
type QueryStakeReputationRequest struct {
	Account string `protobuf:"bytes,1,opt,name=account,proto3" json:"account,omitempty"`
}

// Deprecated: QueryStakeReputationResponse is retained for proto compatibility.
type QueryStakeReputationResponse struct {
	Found bool `protobuf:"varint,2,opt,name=found,proto3" json:"found,omitempty"`
}

// Deprecated: QueryAccountReputationRequest is retained for proto compatibility.
type QueryAccountReputationRequest struct {
	Account string `protobuf:"bytes,1,opt,name=account,proto3" json:"account,omitempty"`
}

// Deprecated: QueryAccountReputationResponse is retained for proto compatibility.
type QueryAccountReputationResponse struct {
	Account string `protobuf:"bytes,1,opt,name=account,proto3" json:"account,omitempty"`
}

type QueryStakingRewardsRequest struct {
	Delegator		string	`protobuf:"bytes,1,opt,name=delegator,proto3" json:"delegator,omitempty"`
	Validator		string	`protobuf:"bytes,2,opt,name=validator,proto3" json:"validator,omitempty"`
	InternalMigration	bool	`protobuf:"varint,3,opt,name=internal_migration,json=internalMigration,proto3" json:"internal_migration,omitempty"`
}

type QueryStakingRewardsResponse struct {
	RewardAmount uint64 `protobuf:"varint,1,opt,name=reward_amount,json=rewardAmount,proto3" json:"reward_amount,omitempty"`
}

type QueryStakingProofRequest struct {
	Kind		string	`protobuf:"bytes,1,opt,name=kind,proto3" json:"kind,omitempty"`
	Height		uint64	`protobuf:"varint,2,opt,name=height,proto3" json:"height,omitempty"`
	PoolID		string	`protobuf:"bytes,3,opt,name=pool_id,json=poolId,proto3" json:"pool_id,omitempty"`
	Account		string	`protobuf:"bytes,4,opt,name=account,proto3" json:"account,omitempty"`
	Epoch		uint64	`protobuf:"varint,5,opt,name=epoch,proto3" json:"epoch,omitempty"`
	AppHash		string	`protobuf:"bytes,6,opt,name=app_hash,json=appHash,proto3" json:"app_hash,omitempty"`
	RootHash	string	`protobuf:"bytes,7,opt,name=root_hash,json=rootHash,proto3" json:"root_hash,omitempty"`
}

type QueryStakingProofResponse struct {
	MetadataJSON string `protobuf:"bytes,1,opt,name=metadata_json,json=metadataJson,proto3" json:"metadata_json,omitempty"`
}

type QueryPoolUnbondingQueueRequest struct {
	PoolID string `protobuf:"bytes,1,opt,name=pool_id,json=poolId,proto3" json:"pool_id,omitempty"`
}

type QueryPoolUnbondingQueueResponse struct {
	UnbondingQueue []UnbondingEntry `protobuf:"bytes,1,rep,name=unbonding_queue,json=unbondingQueue,proto3" json:"unbonding_queue,omitempty"`
}

func DefaultParams() Params {
	return Params{
		Authority:				prototype.DefaultAuthority,
		BaseDenom:				DefaultBaseDenom,
		DisplayDenom:				DefaultDisplayDenom,
		DisplayExponent:			DefaultDisplayExponent,
		MaxPools:				MaxPoolsV1,
		MaxDelegators:				MaxDelegatorsV1,
		MaxPendingDeposits:			MaxPendingDepositsV1,
		MaxPendingWithdrawals:			MaxPendingWithdrawalsV1,
		MaxUnbondingEntries:			MaxUnbondingEntriesV1,
		MaxPoolIDBytes:				MaxPoolIDBytesV1,
		MinValidatorStake:			DefaultMinValidatorStake,
		SoloValidatorMinSelfStake:		DefaultSoloValidatorMinSelfStake,
		PoolBackedValidatorMinSelfStake:	DefaultPoolBackedMinSelfStake,
		PoolBackedValidatorMaxNominatorStake:	DefaultPoolBackedMaxNominatorStake,
		ValidatorSelfStakeMinRatioBps:		4_000,
		ValidatorNominatorStakeMaxRatioBps:	6_000,
		GovernanceMinValidatorCount:		100,
		TargetValidatorCount:			128,
		MaxValidatorCount:			300,
		GovernanceMaxValidatorCount:		300,
		MaxCommissionBps:			DefaultMaxCommissionBps,
		MaxValidatorCommissionBps:		DefaultMaxValidatorCommissionBps,
		MaxOperatorPerformanceBonusBps:		DefaultMaxOperatorPerformanceBonusBps,
		ValidatorCommissionFloorBps:		500,
		DefaultValidatorCommissionBps:		1_000,
		ValidatorCommissionCeilingBps:		2_000,
		ValidatorCommissionMaxDailyChangeBps:	100,
		ValidatorPowerCapBps:			DefaultValidatorPowerCapBps,
		ValidatorPowerCapSchedule: []ValidatorPowerCapPhase{
			{MaxValidatorCount: 150, PowerCapBps: 300},
			{MaxValidatorCount: 250, PowerCapBps: 250},
			{MaxValidatorCount: 0, PowerCapBps: 200},
		},
		DowntimeSlashBps:			5,
		DoubleSignSlashBps:			500,
		DoubleSignTombstone:			true,
		OverflowRewardMultiplierMinBps:		0,
		OverflowRewardMultiplierMaxBps:		3_000,
		UnbondingBlocks:			DefaultUnbondingBlocks,
		RewardEpochDurationBlocks:		DefaultRewardEpochDurationBlocks,
		BaseRewardRateBps:			350,
		MaxRewardRateBps:			600,
		ValidatorChangeDelay:			DefaultValidatorChangeDelay,
		MinPoolDeposit:				DefaultMinPoolDeposit,
		DirectUserValidatorDelegationEnabled:	false,
		ReputationStakeWeightBps:		1_000,
		PoolReceiptDenomOrCodeID:		"aet-liquid-staking-share-v1",
		MaxPoolValidatorAllocationBps:		DefaultMaxPoolValidatorAllocationBps,
		MinPoolValidatorAllocationBps:		DefaultMinPoolValidatorAllocationBps,
		AllocationRebalanceEpochs:		1,
		AllocationUptimeWeight:			2_000,
		AllocationCommissionWeight:		1_500,
		AllocationReputationWeight:		2_000,
		AllocationStakeEfficiencyWeight:	1_500,
		AllocationSlashingRiskWeight:		1_500,
		AllocationNetworkLoadWeight:		1_500,
		PoolProtocolFeeBps:			100,
		ValidatorOperatorBonusBps:		100,
		ValidatorInfrastructureCostModel:	"declared_naet_per_epoch",
		BurnFeeShareBps:			5_000,
		RewardFeeShareBps:			3_500,
		TreasuryFeeShareBps:			1_500,
		InflationMinBps:			200,
		InitialInflationBps:			350,
		InflationMaxBps:			600,
		TargetBondedRatioBps:			6_000,
		AprTargetMinBps:			400,
		AprTargetMaxBps:			700,
		MinTxFeeBaseUnits:			DefaultMinTxFeeBaseUnits,
		FeeDenom:				DefaultBaseDenom,
		StorageRentRatePerByteSecond:		1,
		SystemStorageReserveMinRunwayDays:	365,
		SystemStorageReserveWarningRunwayDays:	180,
		SystemStorageReserveCriticalRunwayDays:	90,
	}
}

func (p Params) Validate() error {
	if err := addressing.ValidateAuthorityAddress("nominator pool authority", p.Authority); err != nil {
		return err
	}
	if p.BaseDenom != appparams.BaseDenom || p.DisplayDenom != appparams.DisplayDenom || p.DisplayExponent != appparams.DisplayDenomExponent {
		return errors.New("nominator pool denom params must match native AET denomination")
	}
	if p.MaxPools == 0 || p.MaxPools > MaxPoolsV1 {
		return fmt.Errorf("nominator pool max pools must be between 1 and %d", MaxPoolsV1)
	}
	if p.MaxDelegators == 0 || p.MaxDelegators > MaxDelegatorsV1 {
		return fmt.Errorf("nominator pool max delegators must be between 1 and %d", MaxDelegatorsV1)
	}
	if p.MaxPendingDeposits == 0 || p.MaxPendingDeposits > MaxPendingDepositsV1 {
		return fmt.Errorf("nominator pool max pending deposits must be between 1 and %d", MaxPendingDepositsV1)
	}
	if p.MaxPendingWithdrawals == 0 || p.MaxPendingWithdrawals > MaxPendingWithdrawalsV1 {
		return fmt.Errorf("nominator pool max pending withdrawals must be between 1 and %d", MaxPendingWithdrawalsV1)
	}
	if p.MaxUnbondingEntries == 0 || p.MaxUnbondingEntries > MaxUnbondingEntriesV1 {
		return fmt.Errorf("nominator pool max unbonding entries must be between 1 and %d", MaxUnbondingEntriesV1)
	}
	if p.MaxPoolIDBytes == 0 || p.MaxPoolIDBytes > MaxPoolIDBytesV1 {
		return fmt.Errorf("nominator pool max pool id bytes must be between 1 and %d", MaxPoolIDBytesV1)
	}
	if err := p.ValidateValidatorFunding(ValidatorFunding{Mode: ValidatorFundingSolo, SelfStake: p.SoloValidatorMinSelfStake}); err != nil {
		return err
	}
	if err := p.ValidateValidatorFunding(ValidatorFunding{Mode: ValidatorFundingPoolBacked, SelfStake: p.PoolBackedValidatorMinSelfStake, NominatorStake: p.PoolBackedValidatorMaxNominatorStake}); err != nil {
		return err
	}
	if p.ValidatorSelfStakeMinRatioBps == 0 || p.ValidatorSelfStakeMinRatioBps > MaxBasisPoints ||
		p.ValidatorNominatorStakeMaxRatioBps == 0 || p.ValidatorNominatorStakeMaxRatioBps > MaxBasisPoints ||
		p.ValidatorSelfStakeMinRatioBps+p.ValidatorNominatorStakeMaxRatioBps != MaxBasisPoints {
		return errors.New("nominator pool validator self/nominator stake ratios are invalid")
	}
	if err := p.ValidateActiveValidatorCount(p.TargetValidatorCount, false); err != nil {
		return err
	}
	if p.MaxValidatorCount != p.GovernanceMaxValidatorCount {
		return errors.New("nominator pool max validator count must match governance max validator count")
	}
	if p.MaxCommissionBps > MaxBasisPoints {
		return fmt.Errorf("nominator pool max commission must be <= %d", MaxBasisPoints)
	}
	if p.MaxValidatorCommissionBps > MaxBasisPoints {
		return fmt.Errorf("nominator pool max validator commission must be <= %d", MaxBasisPoints)
	}
	if p.MaxOperatorPerformanceBonusBps > MaxBasisPoints {
		return fmt.Errorf("nominator pool max operator performance bonus must be <= %d", MaxBasisPoints)
	}
	if err := validateCommissionParams(p); err != nil {
		return err
	}
	if p.ValidatorPowerCapBps == 0 || p.ValidatorPowerCapBps > MaxBasisPoints {
		return errors.New("nominator pool validator power cap is invalid")
	}
	if err := validatePowerCapSchedule(p.ValidatorPowerCapSchedule); err != nil {
		return err
	}

	if !p.DoubleSignTombstone {
		return errors.New("double-sign slash must tombstone")
	}
	if p.OverflowRewardMultiplierMinBps > p.OverflowRewardMultiplierMaxBps || p.OverflowRewardMultiplierMaxBps > 3_000 {
		return errors.New("nominator pool overflow reward multiplier must stay within 0-3000 bps")
	}
	if err := appparams.ValidateStakingUnbondingBlocks(p.UnbondingBlocks); err != nil {
		return fmt.Errorf("nominator pool %w", err)
	}
	if p.RewardEpochDurationBlocks == 0 {
		return errors.New("nominator pool reward epoch duration must be positive")
	}
	if p.BaseRewardRateBps > p.MaxRewardRateBps || p.MaxRewardRateBps > MaxBasisPoints {
		return errors.New("nominator pool reward rate params are invalid")
	}
	if p.ValidatorChangeDelay == 0 {
		return errors.New("nominator pool validator change delay must be positive")
	}
	if p.MinPoolDeposit == 0 {
		return errors.New("nominator pool minimum pool deposit must be positive")
	}
	if strings.TrimSpace(p.PoolReceiptDenomOrCodeID) == "" {
		return errors.New("nominator pool receipt denom or code id is required")
	}
	if p.MinPoolValidatorAllocationBps == 0 || p.MinPoolValidatorAllocationBps > p.MaxPoolValidatorAllocationBps || p.MaxPoolValidatorAllocationBps > p.ValidatorPowerCapBps {
		return errors.New("nominator pool validator allocation bounds are invalid")
	}
	if p.AllocationRebalanceEpochs == 0 {
		return errors.New("nominator pool allocation rebalance epochs must be positive")
	}
	if err := validateAllocationWeights(p); err != nil {
		return err
	}
	if p.PoolProtocolFeeBps > MaxBasisPoints {
		return errors.New("nominator pool protocol fee exceeds basis points")
	}
	if p.ValidatorOperatorBonusBps > p.MaxOperatorPerformanceBonusBps {
		return errors.New("nominator pool validator operator bonus exceeds configured bound")
	}
	if strings.TrimSpace(p.ValidatorInfrastructureCostModel) == "" {
		return errors.New("nominator pool validator infrastructure cost model is required")
	}
	if p.BurnFeeShareBps+p.RewardFeeShareBps+p.TreasuryFeeShareBps != uint32(MaxBasisPoints) {
		return errors.New("nominator pool fee split must sum to 10000 bps")
	}
	if p.InflationMinBps > p.InitialInflationBps || p.InitialInflationBps > p.InflationMaxBps || p.InflationMaxBps > MaxBasisPoints {
		return errors.New("nominator pool inflation params are invalid")
	}
	if p.TargetBondedRatioBps == 0 || p.TargetBondedRatioBps > MaxBasisPoints {
		return errors.New("nominator pool target bonded ratio is invalid")
	}
	if p.AprTargetMinBps > p.AprTargetMaxBps || p.AprTargetMaxBps > MaxBasisPoints {
		return errors.New("nominator pool apr target params are invalid")
	}
	if p.MinTxFeeBaseUnits == 0 || p.FeeDenom != appparams.BaseDenom {
		return errors.New("nominator pool min tx fee denom or amount is invalid")
	}
	if p.StorageRentRatePerByteSecond == 0 {
		return errors.New("nominator pool storage rent rate must be positive")
	}
	if p.SystemStorageReserveMinRunwayDays < p.SystemStorageReserveWarningRunwayDays ||
		p.SystemStorageReserveWarningRunwayDays < p.SystemStorageReserveCriticalRunwayDays ||
		p.SystemStorageReserveCriticalRunwayDays == 0 {
		return errors.New("nominator pool system storage reserve runway thresholds are invalid")
	}
	return nil
}

func (p Params) Authorize(authority string) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("nominator pool update authority", authority); err != nil {
		return err
	}
	if authority != p.Authority {
		return errors.New("nominator pool update requires governance authority")
	}
	return nil
}

func (p Params) ValidateParamsUpdate(authority string, next Params) error {
	if err := p.Authorize(authority); err != nil {
		return err
	}
	next.Authority = p.Authority
	return next.Validate()
}

func (p Params) ValidateValidatorFunding(funding ValidatorFunding) error {
	totalStake, err := CheckedAddUint64(funding.SelfStake, funding.NominatorStake)
	if err != nil {
		return err
	}
	if totalStake < p.MinValidatorStake {
		return errors.New("validator below minimum validator stake")
	}
	if totalStake == 0 {
		return errors.New("validator stake must be positive")
	}
	switch funding.Mode {
	case ValidatorFundingSolo:
		if funding.NominatorStake != 0 {
			return errors.New("solo validator cannot use nominator stake")
		}
		if funding.SelfStake < p.SoloValidatorMinSelfStake {
			return errors.New("solo validator self-stake below configured minimum")
		}
	case ValidatorFundingPoolBacked:
		if funding.SelfStake < p.PoolBackedValidatorMinSelfStake {
			return errors.New("pool-backed validator self-stake below configured minimum")
		}
		if funding.NominatorStake > p.PoolBackedValidatorMaxNominatorStake {
			return errors.New("pool-backed validator nominator stake exceeds configured maximum")
		}
		minimumEntryNominatorStake := uint64(0)
		if p.MinValidatorStake > funding.SelfStake {
			minimumEntryNominatorStake = p.MinValidatorStake - funding.SelfStake
		}
		if minimumEntryNominatorStake > p.PoolBackedValidatorMaxNominatorStake {
			return errors.New("pool-backed validator exceeds nominator share for minimum entry")
		}
	default:
		return fmt.Errorf("unsupported validator funding mode %q", funding.Mode)
	}
	selfRatioBps := funding.SelfStake * uint64(MaxBasisPoints) / totalStake
	if selfRatioBps < uint64(p.ValidatorSelfStakeMinRatioBps) {
		return errors.New("validator self-stake ratio below configured minimum")
	}
	return nil
}

func (p Params) ValidateActiveValidatorCount(count uint32, testnetOverride bool) error {
	if count > p.MaxValidatorCount || count > p.GovernanceMaxValidatorCount {
		return errors.New("active validator count exceeds configured maximum")
	}
	if !testnetOverride && count < p.GovernanceMinValidatorCount {
		return errors.New("active validator count below governance minimum")
	}
	return nil
}

func (p Params) PowerCapBpsForValidatorCount(count uint32) (uint32, error) {
	if len(p.ValidatorPowerCapSchedule) == 0 {
		return 0, errors.New("validator power cap schedule is required")
	}
	for _, phase := range p.ValidatorPowerCapSchedule {
		if phase.MaxValidatorCount == 0 || count <= phase.MaxValidatorCount {
			return phase.PowerCapBps, nil
		}
	}
	return 0, errors.New("validator power cap schedule did not cover validator count")
}

func (p Params) ValidateCommission(rateBps, previousRateBps, dailyChangeBps uint32) error {
	if rateBps < p.ValidatorCommissionFloorBps {
		return errors.New("validator commission below configured floor")
	}
	if rateBps > p.ValidatorCommissionCeilingBps {
		return errors.New("validator commission above configured ceiling")
	}
	if rateBps > previousRateBps {
		if rateBps-previousRateBps > p.ValidatorCommissionMaxDailyChangeBps || dailyChangeBps > p.ValidatorCommissionMaxDailyChangeBps {
			return errors.New("validator commission daily change exceeds configured maximum")
		}
	}
	return nil
}

func (p Params) AllocationWeights(candidates []ValidatorPolicyCandidate) ([]AllocationWeight, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	ordered := append([]ValidatorPolicyCandidate(nil), candidates...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].ValidatorAddress < ordered[j].ValidatorAddress })
	weights := make([]AllocationWeight, 0, len(ordered))
	totalScore := uint64(0)
	for _, candidate := range ordered {
		if err := ValidateUserFacingAEAddress("allocation validator address", candidate.ValidatorAddress); err != nil {
			return nil, err
		}
		score := p.AllocationScore(candidate)
		if candidate.Jailed || candidate.Slashed || candidate.CurrentAllocationBps >= p.MaxPoolValidatorAllocationBps {
			score = 0
		}
		weights = append(weights, AllocationWeight{ValidatorAddress: candidate.ValidatorAddress, Score: score})
		totalScore += score
	}
	remainder := uint32(MaxBasisPoints)
	lastPositive := -1
	for idx := range weights {
		if weights[idx].Score == 0 || totalScore == 0 {
			continue
		}
		weight := uint32(weights[idx].Score * uint64(MaxBasisPoints) / totalScore)
		weights[idx].WeightBps = weight
		if weight <= remainder {
			remainder -= weight
		} else {
			remainder = 0
		}
		lastPositive = idx
	}
	if lastPositive >= 0 {
		weights[lastPositive].WeightBps += remainder
	}
	return weights, nil
}

func (p Params) AllocationScore(candidate ValidatorPolicyCandidate) uint64 {
	if candidate.Jailed || candidate.Slashed {
		return 0
	}
	commissionQuality := uint32(0)
	if candidate.CommissionBps <= p.ValidatorCommissionCeilingBps {
		commissionQuality = p.ValidatorCommissionCeilingBps - candidate.CommissionBps
	}
	slashingSafety := uint32(MaxBasisPoints)
	if candidate.SlashingRiskBps <= MaxBasisPoints {
		slashingSafety = uint32(MaxBasisPoints) - candidate.SlashingRiskBps
	}
	networkHeadroom := uint32(MaxBasisPoints)
	if candidate.NetworkLoadBps <= MaxBasisPoints {
		networkHeadroom = uint32(MaxBasisPoints) - candidate.NetworkLoadBps
	}
	return uint64(candidate.UptimeBps)*uint64(p.AllocationUptimeWeight) +
		uint64(commissionQuality)*uint64(p.AllocationCommissionWeight) +
		uint64(candidate.ReputationScore)*uint64(p.AllocationReputationWeight) +
		uint64(candidate.StakeEfficiencyBps)*uint64(p.AllocationStakeEfficiencyWeight) +
		uint64(slashingSafety)*uint64(p.AllocationSlashingRiskWeight) +
		uint64(networkHeadroom)*uint64(p.AllocationNetworkLoadWeight)
}

func (p Params) ComputeValidatorScoreV1(candidate ValidatorPolicyCandidate) (ValidatorScore, error) {
	uptimeScoreBps := candidate.UptimeBps
	if candidate.UptimeWindow > 0 && candidate.MissedBlocks > 0 {
		if candidate.MissedBlocks > candidate.UptimeWindow {
			return ValidatorScore{}, errors.New("missed blocks exceed uptime window")
		}
		adjusted := uint64(candidate.UptimeBps) * uint64(candidate.UptimeWindow-candidate.MissedBlocks) / uint64(candidate.UptimeWindow)
		if adjusted > uint64(MaxBasisPoints) {
			adjusted = uint64(MaxBasisPoints)
		}
		uptimeScoreBps = uint32(adjusted)
	}

	score := ValidatorScore{
		Eligible:			true,
		UptimeScoreBps:			uptimeScoreBps,
		CommissionScoreBps:		0,
		ReputationScoreBps:		candidate.ReputationScore,
		StakeEfficiencyScoreBps:	candidate.StakeEfficiencyBps,
		NetworkLoadScoreBps:		MaxBasisPoints - candidate.NetworkLoadBps,
		SlashingRiskScoreBps:		MaxBasisPoints - candidate.SlashingRiskBps,
	}

	if candidate.MissedBlocks > candidate.UptimeWindow {
		return score, errors.New("missed blocks exceed uptime window")
	}

	if candidate.Jailed || candidate.Slashed {
		score.Eligible = false
	}

	if candidate.CommissionBps <= p.ValidatorCommissionCeilingBps {
		score.CommissionScoreBps = p.ValidatorCommissionCeilingBps - candidate.CommissionBps
	}

	if !score.Eligible {
		score.OverallScoreBps = 0
	} else {
		totalScore := uint64(score.UptimeScoreBps)*uint64(p.AllocationUptimeWeight) +
			uint64(score.CommissionScoreBps)*uint64(p.AllocationCommissionWeight) +
			uint64(score.ReputationScoreBps)*uint64(p.AllocationReputationWeight) +
			uint64(score.StakeEfficiencyScoreBps)*uint64(p.AllocationStakeEfficiencyWeight) +
			uint64(score.SlashingRiskScoreBps)*uint64(p.AllocationSlashingRiskWeight) +
			uint64(score.NetworkLoadScoreBps)*uint64(p.AllocationNetworkLoadWeight)

		if totalScore > 0 {
			score.OverallScoreBps = uint32(totalScore / (uint64(p.AllocationUptimeWeight + p.AllocationCommissionWeight + p.AllocationReputationWeight + p.AllocationStakeEfficiencyWeight + p.AllocationSlashingRiskWeight + p.AllocationNetworkLoadWeight)))
		}
	}

	return score, nil
}

func (s State) Validate(params Params) error {
	if err := params.Validate(); err != nil {
		return err
	}
	if uint32(len(s.Pools)) > params.MaxPools {
		return errors.New("nominator pool count limit exceeded")
	}
	ids := map[string]struct{}{}
	for _, pool := range s.Pools {
		if err := pool.Validate(params); err != nil {
			return err
		}
		if _, found := ids[pool.PoolID]; found {
			return fmt.Errorf("duplicate nominator pool id %s", pool.PoolID)
		}
		ids[pool.PoolID] = struct{}{}
	}
	validators := map[string]Validator{}
	for _, validator := range s.Validators {
		if err := validator.Validate(params); err != nil {
			return err
		}
		if _, found := validators[validator.Address]; found {
			return fmt.Errorf("duplicate staking validator %s", validator.Address)
		}
		validators[validator.Address] = validator
	}
	validatorScores := map[string]struct{}{}
	for _, score := range s.ValidatorPerformanceScores {
		if err := score.Validate(); err != nil {
			return err
		}
		key := fmt.Sprintf("%s/%020d", score.Validator, score.Epoch)
		if _, found := validatorScores[key]; found {
			return fmt.Errorf("duplicate validator performance score %s", key)
		}
		validatorScores[key] = struct{}{}
	}
	validatorCommissions := map[string]struct{}{}
	for _, commission := range s.ValidatorCommissions {
		if err := commission.Validate(); err != nil {
			return err
		}
		key := fmt.Sprintf("%s/%020d", commission.Validator, commission.Epoch)
		if _, found := validatorCommissions[key]; found {
			return fmt.Errorf("duplicate validator commission %s", key)
		}
		validatorCommissions[key] = struct{}{}
	}
	validatorRisks := map[string]struct{}{}
	for _, risk := range s.ValidatorSlashingRisks {
		if err := risk.Validate(); err != nil {
			return err
		}
		key := fmt.Sprintf("%s/%020d", risk.Validator, risk.Epoch)
		if _, found := validatorRisks[key]; found {
			return fmt.Errorf("duplicate validator slashing risk %s", key)
		}
		validatorRisks[key] = struct{}{}
	}
	validatorLimits := map[string]struct{}{}
	for _, limit := range s.ValidatorAllocationLimits {
		if err := limit.Validate(); err != nil {
			return err
		}
		key := fmt.Sprintf("%s/%020d", limit.Validator, limit.Epoch)
		if _, found := validatorLimits[key]; found {
			return fmt.Errorf("duplicate validator allocation limit %s", key)
		}
		validatorLimits[key] = struct{}{}
	}
	liquidPools := map[string]struct{}{}
	byContractUser := map[string]struct{}{}
	byContractRaw := map[string]struct{}{}
	for _, pool := range s.LiquidStakingPools {
		if err := pool.Validate(params); err != nil {
			return err
		}
		if _, found := liquidPools[pool.PoolID]; found {
			return fmt.Errorf("duplicate liquid staking pool %s", pool.PoolID)
		}
		if _, found := byContractUser[pool.ContractAddressUser]; found {
			return fmt.Errorf("duplicate liquid staking pool contract user address %s", pool.ContractAddressUser)
		}
		if _, found := byContractRaw[pool.ContractAddressRaw]; found {
			return fmt.Errorf("duplicate liquid staking pool contract raw address %s", pool.ContractAddressRaw)
		}
		liquidPools[pool.PoolID] = struct{}{}
		byContractUser[pool.ContractAddressUser] = struct{}{}
		byContractRaw[pool.ContractAddressRaw] = struct{}{}
	}
	poolShares := map[string]struct{}{}
	for _, share := range s.PoolShares {
		if err := share.Validate(params); err != nil {
			return err
		}
		key := share.PoolID + "/" + share.Owner
		if _, found := poolShares[key]; found {
			return fmt.Errorf("duplicate pool share %s", key)
		}
		poolShares[key] = struct{}{}
	}
	poolAllocations := map[string]struct{}{}
	for _, allocation := range s.PoolValidatorAllocations {
		validator, found := validators[allocation.Validator]
		if !found {
			return fmt.Errorf("pool allocation references unknown validator %s", allocation.Validator)
		}
		if err := allocation.Validate(params, validator); err != nil {
			return err
		}
		key := allocation.PoolID + "/" + allocation.Validator
		if _, found := poolAllocations[key]; found {
			return fmt.Errorf("duplicate pool allocation %s", key)
		}
		poolAllocations[key] = struct{}{}
	}
	unbondings := map[string]struct{}{}
	for _, unbonding := range s.PoolUnbondingRequests {
		if err := unbonding.Validate(params); err != nil {
			return err
		}
		key := unbonding.PoolID + "/" + unbonding.Owner + "/" + unbonding.RequestID
		if _, found := unbondings[key]; found {
			return fmt.Errorf("duplicate pool unbonding request %s", key)
		}
		unbondings[key] = struct{}{}
	}
	rewardIndexes := map[string]struct{}{}
	for _, index := range s.PoolRewardIndexes {
		if err := index.Validate(params); err != nil {
			return err
		}
		if _, found := rewardIndexes[index.PoolID]; found {
			return fmt.Errorf("duplicate pool reward index %s", index.PoolID)
		}
		rewardIndexes[index.PoolID] = struct{}{}
	}
	rewardClaims := map[string]struct{}{}
	for _, claim := range s.RewardClaims {
		if err := claim.Validate(params); err != nil {
			return err
		}
		key := fmt.Sprintf("%s/%s/%020d", claim.PoolID, claim.Owner, claim.Epoch)
		if _, found := rewardClaims[key]; found {
			return fmt.Errorf("duplicate reward claim %s", key)
		}
		rewardClaims[key] = struct{}{}
	}
	epochs := map[uint64]struct{}{}
	for _, snapshot := range s.EpochStakingSnapshots {
		if err := snapshot.Validate(); err != nil {
			return err
		}
		if _, found := epochs[snapshot.Epoch]; found {
			return fmt.Errorf("duplicate epoch staking snapshot %d", snapshot.Epoch)
		}
		epochs[snapshot.Epoch] = struct{}{}
	}
	validatorSetSnapshots := map[uint64]struct{}{}
	for _, snapshot := range s.ValidatorSetSnapshots {
		if err := snapshot.Validate(); err != nil {
			return err
		}
		if _, found := validatorSetSnapshots[snapshot.HeightOrEpoch]; found {
			return fmt.Errorf("duplicate validator set snapshot %d", snapshot.HeightOrEpoch)
		}
		validatorSetSnapshots[snapshot.HeightOrEpoch] = struct{}{}
	}
	return nil
}

func (p NominatorPool) Validate(params Params) error {
	if err := validateID("nominator pool id", p.PoolID, params.MaxPoolIDBytes); err != nil {
		return err
	}
	if p.OfficialLiquidStaking || strings.TrimSpace(p.ContractAddressUser) != "" || strings.TrimSpace(p.ContractAddressRaw) != "" {
		if err := ValidateUserFacingAEAddress("official liquid staking contract address", p.ContractAddressUser); err != nil {
			return err
		}
		if err := ValidateRawAddress("official liquid staking contract raw address", p.ContractAddressRaw); err != nil {
			return err
		}
		if err := ValidateAddressPair("official liquid staking contract address pair", p.ContractAddressUser, p.ContractAddressRaw); err != nil {
			return err
		}
	}
	if err := addressing.ValidateAuthorityAddress("nominator pool operator", p.PoolOperator); err != nil {
		return err
	}
	if strings.TrimSpace(p.ValidatorTarget) != "" {
		if err := addressing.ValidateAuthorityAddress("nominator pool validator target", p.ValidatorTarget); err != nil {
			return err
		}
	}
	if strings.TrimSpace(p.PendingValidatorTarget) != "" {
		if err := addressing.ValidateAuthorityAddress("nominator pool pending validator target", p.PendingValidatorTarget); err != nil {
			return err
		}
		if p.ValidatorChangeHeight == 0 {
			return errors.New("nominator pool pending validator change requires activation height")
		}
	}
	if p.PoolCommissionBps > params.MaxCommissionBps {
		return errors.New("nominator pool commission exceeds configured bound")
	}
	if !isPoolStatus(p.Status) {
		return fmt.Errorf("unsupported nominator pool status %q", p.Status)
	}
	if uint32(len(p.DelegatorShares)) > params.MaxDelegators {
		return errors.New("nominator pool delegator limit exceeded")
	}
	if uint32(len(p.PendingDeposits)) > params.MaxPendingDeposits {
		return errors.New("nominator pool pending deposit limit exceeded")
	}
	if uint32(len(p.PendingWithdrawals)) > params.MaxPendingWithdrawals {
		return errors.New("nominator pool pending withdrawal limit exceeded")
	}
	if uint32(len(p.UnbondingQueue)) > params.MaxUnbondingEntries {
		return errors.New("nominator pool unbonding queue limit exceeded")
	}
	if p.TotalShares != sumShares(p.DelegatorShares) {
		return errors.New("nominator pool total shares do not match delegator shares")
	}
	if err := ValidateAllocations(p.Allocations, p.TotalBondedStake); err != nil {
		return err
	}
	delegators := map[string]struct{}{}
	for _, delegator := range p.DelegatorShares {
		if err := delegator.Validate(); err != nil {
			return err
		}
		if _, found := delegators[delegator.Delegator]; found {
			return fmt.Errorf("duplicate pool delegator %s", delegator.Delegator)
		}
		delegators[delegator.Delegator] = struct{}{}
	}
	withdrawals := map[string]struct{}{}
	for _, withdrawal := range p.PendingWithdrawals {
		if err := withdrawal.Validate(); err != nil {
			return err
		}
		if _, found := withdrawals[withdrawal.WithdrawalID]; found {
			return fmt.Errorf("duplicate pool withdrawal %s", withdrawal.WithdrawalID)
		}
		withdrawals[withdrawal.WithdrawalID] = struct{}{}
	}
	for _, deposit := range p.PendingDeposits {
		if err := deposit.Validate(); err != nil {
			return err
		}
	}
	for _, entry := range p.UnbondingQueue {
		if err := entry.Validate(); err != nil {
			return err
		}
	}
	incomeByValidator := map[string]struct{}{}
	for _, income := range p.ValidatorOperatorIncome {
		if err := income.Validate(); err != nil {
			return err
		}
		if _, found := incomeByValidator[income.Validator]; found {
			return fmt.Errorf("duplicate validator income %s", income.Validator)
		}
		incomeByValidator[income.Validator] = struct{}{}
	}
	for _, allocation := range p.ValidatorAllocations {
		if err := allocation.Validate(params); err != nil {
			return err
		}
	}
	return nil
}

func (a PoolAllocation) Validate() error {
	if err := ValidateUserFacingAEAddress("pool allocation validator address", a.ValidatorAddress); err != nil {
		return err
	}
	if a.Amount == 0 || a.Height == 0 {
		return errors.New("pool allocation amount and height must be positive")
	}
	return nil
}

func ValidateAllocations(allocations []PoolAllocation, totalBondedStake uint64) error {
	previous := ""
	total := uint64(0)
	for _, allocation := range allocations {
		if err := allocation.Validate(); err != nil {
			return err
		}
		if allocation.ValidatorAddress <= previous {
			return errors.New("pool allocations must be sorted by unique validator address")
		}
		previous = allocation.ValidatorAddress
		if allocation.Amount > totalBondedStake-total {
			return errors.New("pool allocations exceed bonded stake")
		}
		total += allocation.Amount
	}
	return nil
}

func (a ValidatorRewardAllocation) Validate(params Params) error {
	if err := addressing.ValidateAuthorityAddress("nominator pool reward validator", a.Validator); err != nil {
		return err
	}
	if a.PerformanceBps > MaxBasisPoints {
		return errors.New("nominator pool validator performance exceeds basis points")
	}
	if a.CommissionBps > params.MaxValidatorCommissionBps {
		return errors.New("nominator pool validator commission exceeds configured bound")
	}
	if a.OperatorPerformanceBonusBps > params.MaxOperatorPerformanceBonusBps {
		return errors.New("nominator pool operator performance bonus exceeds configured bound")
	}
	if a.Jailed && a.OperatorPerformanceBonus > 0 {
		return errors.New("jailed validator cannot receive positive operator bonus")
	}
	return nil
}

func (i ValidatorIncome) Validate() error {
	if err := addressing.ValidateAuthorityAddress("nominator pool validator income", i.Validator); err != nil {
		return err
	}
	expectedGross, err := CheckedAddUint64(i.SelfStakeRewards, i.CommissionIncome)
	if err != nil {
		return err
	}
	expectedGross, err = CheckedAddUint64(expectedGross, i.OperatorPerformanceBonus)
	if err != nil {
		return err
	}
	if i.GrossIncome != expectedGross {
		return errors.New("nominator pool validator gross income does not reconcile")
	}
	if i.NetIncome != SaturatingNetIncome(i.GrossIncome, i.InfrastructureCost) {
		return errors.New("nominator pool validator net income does not reconcile")
	}
	return nil
}

func (d PendingDeposit) Validate() error {
	if err := addressing.ValidateAuthorityAddress("nominator pool pending deposit delegator", d.Delegator); err != nil {
		return err
	}
	if d.Amount == 0 || d.Height == 0 {
		return errors.New("nominator pool pending deposit amount and height must be positive")
	}
	return nil
}

func (w PendingWithdrawal) Validate() error {
	if err := validateID("nominator pool withdrawal id", w.WithdrawalID, MaxPoolIDBytesV1); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("nominator pool withdrawal delegator", w.Delegator); err != nil {
		return err
	}
	if w.Shares == 0 || w.Amount == 0 || w.RequestHeight == 0 || w.CompleteHeight <= w.RequestHeight {
		return errors.New("nominator pool withdrawal amounts and heights are invalid")
	}
	if !isWithdrawalStatus(w.Status) {
		return fmt.Errorf("unsupported nominator pool withdrawal status %q", w.Status)
	}
	return nil
}

func (d DelegatorShare) Validate() error {
	if err := addressing.ValidateAuthorityAddress("nominator pool delegator", d.Delegator); err != nil {
		return err
	}
	if d.Shares == 0 {
		return errors.New("nominator pool delegator shares must be positive")
	}
	return nil
}

func (e UnbondingEntry) Validate() error {
	if err := validateID("nominator pool unbonding withdrawal id", e.WithdrawalID, MaxPoolIDBytesV1); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("nominator pool unbonding delegator", e.Delegator); err != nil {
		return err
	}
	if e.Amount == 0 || e.CompleteHeight == 0 {
		return errors.New("nominator pool unbonding amount and completion height must be positive")
	}
	if !isWithdrawalStatus(e.Status) {
		return fmt.Errorf("unsupported nominator pool unbonding status %q", e.Status)
	}
	return nil
}

func (s State) Normalize(params Params) State {
	s.Pools = SortPools(s.Pools)
	for idx := range s.Pools {
		s.Pools[idx].Allocations = SortAllocations(s.Pools[idx].Allocations)
		s.Pools[idx].PendingDeposits = SortDeposits(s.Pools[idx].PendingDeposits)
		s.Pools[idx].PendingWithdrawals = SortWithdrawals(s.Pools[idx].PendingWithdrawals)
		s.Pools[idx].DelegatorShares = SortDelegators(s.Pools[idx].DelegatorShares)
		s.Pools[idx].UnbondingQueue = SortUnbonding(s.Pools[idx].UnbondingQueue)
		s.Pools[idx].ValidatorOperatorIncome = SortValidatorIncome(s.Pools[idx].ValidatorOperatorIncome)
		s.Pools[idx].ValidatorAllocations = SortValidatorRewardAllocations(s.Pools[idx].ValidatorAllocations)
	}
	s.Validators = SortStateValidators(s.Validators)
	s.ValidatorPerformanceScores = SortValidatorPerformanceScores(s.ValidatorPerformanceScores)
	s.ValidatorCommissions = SortValidatorCommissions(s.ValidatorCommissions)
	s.ValidatorSlashingRisks = SortValidatorSlashingRisks(s.ValidatorSlashingRisks)
	s.ValidatorAllocationLimits = SortValidatorAllocationLimits(s.ValidatorAllocationLimits)
	s.LiquidStakingPools = SortLiquidStakingPools(s.LiquidStakingPools)
	s.PoolShares = SortPoolShares(s.PoolShares)
	s.PoolValidatorAllocations = SortPoolValidatorAllocations(s.PoolValidatorAllocations)
	s.PoolUnbondingRequests = SortPoolUnbondingRequests(s.PoolUnbondingRequests)
	s.PoolRewardIndexes = SortPoolRewardIndexes(s.PoolRewardIndexes)
	s.RewardClaims = SortRewardClaims(s.RewardClaims)
	s.EpochStakingSnapshots = SortEpochStakingSnapshots(s.EpochStakingSnapshots)
	s.ValidatorSetSnapshots = SortValidatorSetSnapshots(s.ValidatorSetSnapshots)
	return s
}

func SortAllocations(values []PoolAllocation) []PoolAllocation {
	out := append([]PoolAllocation(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].ValidatorAddress < out[j].ValidatorAddress })
	return out
}

func SortValidatorIncome(values []ValidatorIncome) []ValidatorIncome {
	out := append([]ValidatorIncome(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Validator < out[j].Validator })
	return out
}

func SortValidatorRewardAllocations(values []ValidatorRewardAllocation) []ValidatorRewardAllocation {
	out := append([]ValidatorRewardAllocation(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Validator < out[j].Validator })
	return out
}

func SortStateValidators(values []Validator) []Validator {
	out := append([]Validator(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Address < out[j].Address })
	return out
}

func SortValidatorPerformanceScores(values []ValidatorPerformanceScore) []ValidatorPerformanceScore {
	out := append([]ValidatorPerformanceScore(nil), values...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Validator != out[j].Validator {
			return out[i].Validator < out[j].Validator
		}
		return out[i].Epoch < out[j].Epoch
	})
	return out
}

func SortValidatorCommissions(values []ValidatorCommission) []ValidatorCommission {
	out := append([]ValidatorCommission(nil), values...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Validator != out[j].Validator {
			return out[i].Validator < out[j].Validator
		}
		return out[i].Epoch < out[j].Epoch
	})
	return out
}

func SortValidatorSlashingRisks(values []ValidatorSlashingRisk) []ValidatorSlashingRisk {
	out := append([]ValidatorSlashingRisk(nil), values...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Validator != out[j].Validator {
			return out[i].Validator < out[j].Validator
		}
		return out[i].Epoch < out[j].Epoch
	})
	return out
}

func SortValidatorAllocationLimits(values []ValidatorAllocationLimit) []ValidatorAllocationLimit {
	out := append([]ValidatorAllocationLimit(nil), values...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Validator != out[j].Validator {
			return out[i].Validator < out[j].Validator
		}
		return out[i].Epoch < out[j].Epoch
	})
	return out
}

func SortLiquidStakingPools(values []LiquidStakingPool) []LiquidStakingPool {
	out := append([]LiquidStakingPool(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].PoolID < out[j].PoolID })
	return out
}

func SortPoolShares(values []PoolShare) []PoolShare {
	out := append([]PoolShare(nil), values...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].PoolID != out[j].PoolID {
			return out[i].PoolID < out[j].PoolID
		}
		return out[i].Owner < out[j].Owner
	})
	return out
}

func SortPoolValidatorAllocations(values []PoolValidatorAllocation) []PoolValidatorAllocation {
	out := append([]PoolValidatorAllocation(nil), values...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].PoolID != out[j].PoolID {
			return out[i].PoolID < out[j].PoolID
		}
		return out[i].Validator < out[j].Validator
	})
	return out
}

func SortPoolUnbondingRequests(values []PoolUnbondingRequest) []PoolUnbondingRequest {
	out := append([]PoolUnbondingRequest(nil), values...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].PoolID != out[j].PoolID {
			return out[i].PoolID < out[j].PoolID
		}
		if out[i].Owner != out[j].Owner {
			return out[i].Owner < out[j].Owner
		}
		return out[i].RequestID < out[j].RequestID
	})
	return out
}

func SortPoolRewardIndexes(values []PoolRewardIndex) []PoolRewardIndex {
	out := append([]PoolRewardIndex(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].PoolID < out[j].PoolID })
	return out
}

func SortRewardClaims(values []RewardClaim) []RewardClaim {
	out := append([]RewardClaim(nil), values...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].PoolID != out[j].PoolID {
			return out[i].PoolID < out[j].PoolID
		}
		if out[i].Owner != out[j].Owner {
			return out[i].Owner < out[j].Owner
		}
		return out[i].Epoch < out[j].Epoch
	})
	return out
}

func SortEpochStakingSnapshots(values []EpochStakingSnapshot) []EpochStakingSnapshot {
	out := append([]EpochStakingSnapshot(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Epoch < out[j].Epoch })
	return out
}

func SortValidatorSetSnapshots(values []ValidatorSetSnapshot) []ValidatorSetSnapshot {
	out := append([]ValidatorSetSnapshot(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].HeightOrEpoch < out[j].HeightOrEpoch })
	return out
}

func PaginatePoolSharesByOwner(shares []PoolShare, owner string, offset, limit uint32) []PoolShare {
	filtered := []PoolShare{}
	for _, share := range SortPoolShares(shares) {
		if share.Owner == owner {
			filtered = append(filtered, share)
		}
	}
	return paginate(filtered, offset, limit)
}

func PaginatePoolAllocationsByPool(allocations []PoolValidatorAllocation, poolID string, offset, limit uint32) []PoolValidatorAllocation {
	filtered := []PoolValidatorAllocation{}
	for _, allocation := range SortPoolValidatorAllocations(allocations) {
		if allocation.PoolID == poolID {
			filtered = append(filtered, allocation)
		}
	}
	return paginate(filtered, offset, limit)
}

func PaginateValidators(validators []Validator, offset, limit uint32) []Validator {
	return paginate(SortStateValidators(validators), offset, limit)
}

func paginate[T any](values []T, offset, limit uint32) []T {
	if limit == 0 || int(offset) >= len(values) {
		return []T{}
	}
	end := int(offset + limit)
	if end > len(values) {
		end = len(values)
	}
	start := int(offset)
	return append([]T(nil), values[start:end]...)
}

func SortPools(values []NominatorPool) []NominatorPool {
	out := append([]NominatorPool(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].PoolID < out[j].PoolID })
	return out
}

func SortDelegators(values []DelegatorShare) []DelegatorShare {
	out := append([]DelegatorShare(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Delegator < out[j].Delegator })
	return out
}

func SortWithdrawals(values []PendingWithdrawal) []PendingWithdrawal {
	out := append([]PendingWithdrawal(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].WithdrawalID < out[j].WithdrawalID })
	return out
}

func SortDeposits(values []PendingDeposit) []PendingDeposit {
	out := append([]PendingDeposit(nil), values...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Height != out[j].Height {
			return out[i].Height < out[j].Height
		}
		return out[i].Delegator < out[j].Delegator
	})
	return out
}

func SortUnbonding(values []UnbondingEntry) []UnbondingEntry {
	out := append([]UnbondingEntry(nil), values...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].CompleteHeight != out[j].CompleteHeight {
			return out[i].CompleteHeight < out[j].CompleteHeight
		}
		return out[i].WithdrawalID < out[j].WithdrawalID
	})
	return out
}

func ShareValue(pool NominatorPool, shares uint64) uint64 {
	if pool.TotalShares == 0 {
		return 0
	}
	return shares * pool.TotalBondedStake / pool.TotalShares
}

func SharesForDeposit(pool NominatorPool, amount uint64) uint64 {
	shares, err := SharesForDepositChecked(pool, amount)
	if err != nil {
		return 0
	}
	return shares
}

func SharesForDepositChecked(pool NominatorPool, amount uint64) (uint64, error) {
	if pool.TotalShares == 0 || pool.TotalBondedStake == 0 {
		return amount, nil
	}
	shares, err := MulDivUint64(amount, pool.TotalShares, pool.TotalBondedStake)
	if err != nil {
		return 0, err
	}
	if shares == 0 && amount > 0 {
		return 1, nil
	}
	return shares, nil
}

func ValidateOfficialLiquidStakingDeposit(msg MsgDepositToOfficialLiquidStaking, params Params) error {
	if err := params.Authorize(msg.Authority); err != nil {
		return err
	}
	if err := ValidateUserFacingAEAddress("official liquid staking depositor", msg.UserAddress); err != nil {
		return err
	}
	if strings.TrimSpace(msg.ValidatorAddress) != "" {
		return errors.New("official liquid staking deposit must not include a validator address")
	}
	if msg.Amount < params.MinPoolDeposit {
		return fmt.Errorf("official liquid staking deposit below configured minimum %d", params.MinPoolDeposit)
	}
	if msg.Height == 0 {
		return errors.New("official liquid staking deposit height must be positive")
	}
	return validateID("official liquid staking pool id", msg.PoolID, params.MaxPoolIDBytes)
}

func ValidateStakingPoolDeposit(msg MsgDepositToStakingPool, params Params) error {
	if err := ValidateUserFacingAEAddress("staking pool depositor", msg.WalletAddress); err != nil {
		return err
	}
	if strings.TrimSpace(msg.ReservedRouting) != "" {
		return errors.New("staking pool deposit must not include a routing field")
	}
	if msg.Amount < params.MinPoolDeposit {
		return fmt.Errorf("staking pool deposit below configured minimum %d", params.MinPoolDeposit)
	}
	if msg.Height == 0 {
		return errors.New("staking pool deposit height must be positive")
	}
	return validateID("staking pool id", msg.PoolID, params.MaxPoolIDBytes)
}

func ValidateDirectUserDelegation(msg MsgDelegateToValidator, params Params) error {
	if err := params.Authorize(msg.Authority); err != nil {
		return err
	}
	if !params.DirectUserDelegationEnabled && !params.DirectUserValidatorDelegationEnabled {
		return errors.New("direct user delegation to validators is disabled; use official liquid staking pool deposit")
	}
	if err := ValidateUserFacingAEAddress("direct delegation user address", msg.UserAddress); err != nil {
		return err
	}
	if err := ValidateUserFacingAEAddress("direct delegation validator address", msg.ValidatorAddress); err != nil {
		return err
	}
	if msg.Amount == 0 || msg.Height == 0 {
		return errors.New("direct delegation amount and height must be positive")
	}
	return nil
}

func ValidateUserFacingAEAddress(field, text string) error {
	text = strings.TrimSpace(text)
	if !strings.HasPrefix(text, addressing.UserFriendlyPrefix) {
		return fmt.Errorf("%s must use AE user-facing address format", field)
	}
	return addressing.ValidateUserAddress(field, text)
}

func ValidateRawAddress(field, text string) error {
	text = strings.TrimSpace(text)
	if !strings.HasPrefix(text, addressing.RawPrefix) {
		return fmt.Errorf("%s must use 4: raw address format", field)
	}
	_, err := addressing.Parse(text)
	if err != nil {
		return fmt.Errorf("invalid %s: %w", field, err)
	}
	return nil
}

func ValidateAddressPair(field, userAddress, rawAddress string) error {
	userBytes, err := addressing.Parse(userAddress)
	if err != nil {
		return fmt.Errorf("invalid %s user address: %w", field, err)
	}
	rawBytes, err := addressing.Parse(rawAddress)
	if err != nil {
		return fmt.Errorf("invalid %s raw address: %w", field, err)
	}
	userKey, err := addressing.AddressTextBytesKey(userAddress)
	if err != nil {
		return err
	}
	rawKey, err := addressing.AddressTextBytesKey(rawAddress)
	if err != nil {
		return err
	}
	if userKey != rawKey || string(userBytes) != string(rawBytes) {
		return fmt.Errorf("%s AE and raw addresses must represent the same account", field)
	}
	return nil
}

func RawAddressForUserAddress(userAddress string) (string, error) {
	if err := ValidateUserFacingAEAddress("user address", userAddress); err != nil {
		return "", err
	}
	bz, err := addressing.Parse(userAddress)
	if err != nil {
		return "", err
	}
	return addressing.Format(bz), nil
}

func RewardDelta(amount uint64, totalShares uint64) uint64 {
	if amount == 0 || totalShares == 0 {
		return 0
	}
	return amount * IndexScale / totalShares
}

func IndexedRewardAmount(delta uint64, totalShares uint64) uint64 {
	if delta == 0 || totalShares == 0 {
		return 0
	}
	return delta * totalShares / IndexScale
}

func AccruedReward(delegator DelegatorShare, rewardIndex uint64) uint64 {
	if rewardIndex <= delegator.RewardIndexCheckpoint {
		return delegator.PendingRewards
	}
	return delegator.PendingRewards + delegator.Shares*(rewardIndex-delegator.RewardIndexCheckpoint)/IndexScale
}

func SyncPoolRewards(params Params, pool NominatorPool, msg MsgSyncPoolRewards) (NominatorPool, PoolRewardSummary, error) {
	if err := params.Authorize(msg.Authority); err != nil {
		return NominatorPool{}, PoolRewardSummary{}, err
	}
	if msg.PoolID != pool.PoolID {
		return NominatorPool{}, PoolRewardSummary{}, errors.New("nominator pool reward sync pool mismatch")
	}
	if msg.Epoch == 0 || msg.Height == 0 {
		return NominatorPool{}, PoolRewardSummary{}, errors.New("nominator pool reward sync epoch and height must be positive")
	}
	if msg.Epoch <= pool.RewardEpoch {
		return NominatorPool{}, PoolRewardSummary{}, errors.New("nominator pool reward sync epoch must increase")
	}
	if msg.RewardRateBps > MaxBasisPoints {
		return NominatorPool{}, PoolRewardSummary{}, errors.New("nominator pool reward rate exceeds basis points")
	}
	if len(msg.Allocations) == 0 {
		return NominatorPool{}, PoolRewardSummary{}, errors.New("nominator pool reward sync requires validator allocations")
	}
	rewardCap, err := CheckedAddUint64(msg.EmissionsAllocated, msg.FeesAllocated)
	if err != nil {
		return NominatorPool{}, PoolRewardSummary{}, err
	}

	next := pool
	next.RewardEpoch = msg.Epoch
	next.ValidatorAllocations = nil
	summary := PoolRewardSummary{
		PoolID:			pool.PoolID,
		Epoch:			msg.Epoch,
		RewardRateBps:		msg.RewardRateBps,
		EmissionsAllocated:	msg.EmissionsAllocated,
		FeesAllocated:		msg.FeesAllocated,
		RewardCap:		rewardCap,
		RewardIndexBefore:	pool.RewardIndex,
	}
	income := map[string]ValidatorIncome{}
	totalRewardOut := uint64(0)
	index := pool.RewardIndex
	remainder := pool.RewardRemainder
	for _, allocation := range SortValidatorRewardAllocations(msg.Allocations) {
		if err := allocation.Validate(params); err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		effectivePerformanceBps := allocation.PerformanceBps
		operatorBonusBps := allocation.OperatorPerformanceBonusBps
		if allocation.Jailed {
			effectivePerformanceBps = 0
			operatorBonusBps = 0
		}
		if allocation.SlashingLoss > 0 {
			operatorBonusBps = 0
		}
		grossPoolRewards, err := RewardForStake(allocation.PoolAllocatedStake, msg.RewardRateBps, effectivePerformanceBps)
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		commission, err := MulDivUint64(grossPoolRewards, uint64(allocation.CommissionBps), uint64(MaxBasisPoints))
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		afterCommission := grossPoolRewards - commission
		poolFee, err := MulDivUint64(afterCommission, uint64(pool.PoolCommissionBps), uint64(MaxBasisPoints))
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		netPoolRewards := afterCommission - poolFee
		selfStakeRewards, err := RewardForStake(allocation.ValidatorSelfStake, msg.RewardRateBps, effectivePerformanceBps)
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		operatorBonus, err := MulDivUint64(grossPoolRewards, uint64(operatorBonusBps), uint64(MaxBasisPoints))
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		validatorGross, err := CheckedAddUint64(selfStakeRewards, commission)
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		validatorGross, err = CheckedAddUint64(validatorGross, operatorBonus)
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		totalRewardOut, err = CheckedAddUint64(totalRewardOut, grossPoolRewards)
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		totalRewardOut, err = CheckedAddUint64(totalRewardOut, selfStakeRewards)
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		totalRewardOut, err = CheckedAddUint64(totalRewardOut, operatorBonus)
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		if totalRewardOut > rewardCap {
			return NominatorPool{}, PoolRewardSummary{}, errors.New("nominator pool rewards exceed emissions and fee allocation cap")
		}

		delta := RewardDelta(netPoolRewards, pool.TotalShares)
		index, err = CheckedAddUint64(index, delta)
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		distributed := IndexedRewardAmount(delta, pool.TotalShares)
		if netPoolRewards >= distributed {
			remainder, err = CheckedAddUint64(remainder, netPoolRewards-distributed)
			if err != nil {
				return NominatorPool{}, PoolRewardSummary{}, err
			}
		}
		allocation.PerformanceBps = effectivePerformanceBps
		allocation.GrossPoolRewards = grossPoolRewards
		allocation.ValidatorCommission = commission
		allocation.PoolProtocolFee = poolFee
		allocation.NetPoolRewards = netPoolRewards
		allocation.ValidatorSelfStakeRewards = selfStakeRewards
		allocation.OperatorPerformanceBonusBps = operatorBonusBps
		allocation.OperatorPerformanceBonus = operatorBonus
		allocation.ValidatorGrossIncome = validatorGross
		allocation.ValidatorNetIncome = SaturatingNetIncome(validatorGross, allocation.InfrastructureCost)
		allocation.RewardIndexDelta = delta
		allocation.RewardIndexAfter = index
		next.ValidatorAllocations = append(next.ValidatorAllocations, allocation)

		entry := income[allocation.Validator]
		entry.Validator = allocation.Validator
		entry.SelfStakeRewards, err = CheckedAddUint64(entry.SelfStakeRewards, selfStakeRewards)
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		entry.CommissionIncome, err = CheckedAddUint64(entry.CommissionIncome, commission)
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		entry.OperatorPerformanceBonus, err = CheckedAddUint64(entry.OperatorPerformanceBonus, operatorBonus)
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		entry.InfrastructureCost, err = CheckedAddUint64(entry.InfrastructureCost, allocation.InfrastructureCost)
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		entry.GrossIncome, err = CheckedAddUint64(entry.GrossIncome, validatorGross)
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		entry.NetIncome = SaturatingNetIncome(entry.GrossIncome, entry.InfrastructureCost)
		income[allocation.Validator] = entry

		summary.GrossPoolRewards, err = CheckedAddUint64(summary.GrossPoolRewards, grossPoolRewards)
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		summary.ValidatorCommission, err = CheckedAddUint64(summary.ValidatorCommission, commission)
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		summary.PoolProtocolFee, err = CheckedAddUint64(summary.PoolProtocolFee, poolFee)
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		summary.PoolUserRewards, err = CheckedAddUint64(summary.PoolUserRewards, netPoolRewards)
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		summary.SlashingLosses, err = CheckedAddUint64(summary.SlashingLosses, allocation.SlashingLoss)
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		summary.ValidatorSelfStakeRewards, err = CheckedAddUint64(summary.ValidatorSelfStakeRewards, selfStakeRewards)
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		summary.OperatorPerformanceBonus, err = CheckedAddUint64(summary.OperatorPerformanceBonus, operatorBonus)
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		summary.ValidatorGrossIncome, err = CheckedAddUint64(summary.ValidatorGrossIncome, validatorGross)
		if err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		summary.AllocationsTouched++
	}
	next.ValidatorOperatorIncome = next.ValidatorOperatorIncome[:0]
	for _, value := range income {
		if err := value.Validate(); err != nil {
			return NominatorPool{}, PoolRewardSummary{}, err
		}
		next.ValidatorOperatorIncome = append(next.ValidatorOperatorIncome, value)
	}
	next.ValidatorOperatorIncome = SortValidatorIncome(next.ValidatorOperatorIncome)
	next.RewardIndex = index
	next.RewardRemainder = remainder
	next.ProtocolFeeAccrued, err = CheckedAddUint64(next.ProtocolFeeAccrued, summary.PoolProtocolFee)
	if err != nil {
		return NominatorPool{}, PoolRewardSummary{}, err
	}
	next.ValidatorCommissionAccrued, err = CheckedAddUint64(next.ValidatorCommissionAccrued, summary.ValidatorCommission)
	if err != nil {
		return NominatorPool{}, PoolRewardSummary{}, err
	}
	next.TotalBondedStake, err = CheckedAddUint64(next.TotalBondedStake, summary.PoolUserRewards)
	if err != nil {
		return NominatorPool{}, PoolRewardSummary{}, err
	}
	if summary.SlashingLosses > next.TotalBondedStake {
		summary.SlashingLosses = next.TotalBondedStake
	}
	next.TotalBondedStake -= summary.SlashingLosses
	next.SlashIndex, err = CheckedAddUint64(next.SlashIndex, RewardDelta(summary.SlashingLosses, pool.TotalShares))
	if err != nil {
		return NominatorPool{}, PoolRewardSummary{}, err
	}
	summary.RewardIndexAfter = next.RewardIndex
	summary.RewardRemainder = next.RewardRemainder
	summary.ValidatorNetIncome = SaturatingNetIncome(summary.ValidatorGrossIncome, totalInfrastructureCost(next.ValidatorOperatorIncome))
	return next, summary, nil
}

func RewardForStake(stake uint64, rewardRateBps uint32, performanceBps uint32) (uint64, error) {
	if rewardRateBps > MaxBasisPoints || performanceBps > MaxBasisPoints {
		return 0, errors.New("nominator pool reward inputs exceed basis points")
	}
	first, err := MulDivUint64(stake, uint64(rewardRateBps), uint64(MaxBasisPoints))
	if err != nil {
		return 0, err
	}
	return MulDivUint64(first, uint64(performanceBps), uint64(MaxBasisPoints))
}

func MulDivUint64(value, multiplier, denominator uint64) (uint64, error) {
	if denominator == 0 {
		return 0, errors.New("nominator pool division by zero")
	}
	product := new(big.Int).Mul(new(big.Int).SetUint64(value), new(big.Int).SetUint64(multiplier))
	product.Quo(product, new(big.Int).SetUint64(denominator))
	if !product.IsUint64() {
		return 0, errors.New("nominator pool uint64 accounting overflow")
	}
	return product.Uint64(), nil
}

func CheckedAddUint64(left, right uint64) (uint64, error) {
	if math.MaxUint64-left < right {
		return 0, errors.New("nominator pool uint64 accounting overflow")
	}
	return left + right, nil
}

func SaturatingNetIncome(gross uint64, cost uint64) int64 {
	if cost >= gross {
		delta := cost - gross
		if delta > uint64(math.MaxInt64) {
			return math.MinInt64
		}
		return -int64(delta)
	}
	delta := gross - cost
	if delta > uint64(math.MaxInt64) {
		return math.MaxInt64
	}
	return int64(delta)
}

func totalInfrastructureCost(values []ValidatorIncome) uint64 {
	total := uint64(0)
	for _, value := range values {
		next, err := CheckedAddUint64(total, value.InfrastructureCost)
		if err != nil {
			return math.MaxUint64
		}
		total = next
	}
	return total
}

func IsJailedValidatorStatus(status string) bool {
	return status == validatorregistrytypes.StatusJailed || status == validatorregistrytypes.StatusTombstoned
}

func validateCommissionParams(p Params) error {
	if p.ValidatorCommissionFloorBps > p.DefaultValidatorCommissionBps ||
		p.DefaultValidatorCommissionBps > p.ValidatorCommissionCeilingBps ||
		p.ValidatorCommissionCeilingBps > MaxBasisPoints {
		return errors.New("nominator pool validator commission floor/default/ceiling are invalid")
	}
	if p.ValidatorCommissionMaxDailyChangeBps == 0 || p.ValidatorCommissionMaxDailyChangeBps > p.ValidatorCommissionCeilingBps {
		return errors.New("nominator pool validator commission daily change is invalid")
	}
	if p.MaxValidatorCommissionBps < p.ValidatorCommissionCeilingBps || p.MaxCommissionBps < p.PoolProtocolFeeBps {
		return errors.New("nominator pool commission bounds are inconsistent")
	}
	return nil
}

func validatePowerCapSchedule(schedule []ValidatorPowerCapPhase) error {
	if len(schedule) == 0 {
		return errors.New("nominator pool validator power cap schedule is required")
	}
	previousMax := uint32(0)
	for idx, phase := range schedule {
		if phase.PowerCapBps == 0 || phase.PowerCapBps > MaxBasisPoints {
			return errors.New("nominator pool validator power cap phase is invalid")
		}
		if idx < len(schedule)-1 {
			if phase.MaxValidatorCount <= previousMax {
				return errors.New("nominator pool validator power cap schedule must be sorted")
			}
			previousMax = phase.MaxValidatorCount
			continue
		}
		if phase.MaxValidatorCount != 0 {
			return errors.New("nominator pool final validator power cap phase must be open-ended")
		}
	}
	return nil
}

func validateAllocationWeights(p Params) error {
	weights := []uint32{
		p.AllocationUptimeWeight,
		p.AllocationCommissionWeight,
		p.AllocationReputationWeight,
		p.AllocationStakeEfficiencyWeight,
		p.AllocationSlashingRiskWeight,
		p.AllocationNetworkLoadWeight,
	}
	total := uint32(0)
	for _, weight := range weights {
		if weight == 0 {
			return errors.New("nominator pool allocation weights must be positive")
		}
		total += weight
	}
	if total != MaxBasisPoints {
		return errors.New("nominator pool allocation weights must sum to 10000 bps")
	}
	return nil
}

func validateID(field, value string, maxBytes uint32) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("%s is required", field)
	}
	if uint32(len(value)) > maxBytes || strings.ContainsAny(value, " \t\r\n") {
		return fmt.Errorf("%s must be non-blank, whitespace-free, and within configured length", field)
	}
	return nil
}

func isPoolStatus(status string) bool {
	return status == PoolStatusActive || status == PoolStatusPaused || status == PoolStatusFrozenLimited || status == PoolStatusClosed
}

func isWithdrawalStatus(status string) bool {
	return status == WithdrawalStatusPending || status == WithdrawalStatusCancelled || status == WithdrawalStatusCompleted
}

func sumShares(values []DelegatorShare) uint64 {
	total := uint64(0)
	for _, value := range values {
		total += value.Shares
	}
	return total
}
