package keeper

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/x/nominator-pool/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct{ keeper *Keeper }

func NewMsgServerImpl(k *Keeper) types.MsgServer	{ return msgServer{keeper: k} }

func (m msgServer) CreateNominatorPool(ctx context.Context, msg *types.MsgCreateNominatorPool) (*types.MsgCreateNominatorPoolResponse, error) {
	if msg == nil {
		return nil, errors.New("empty nominator pool creation request")
	}
	msg.Height = defaultHeight(ctx, msg.Height)
	pool, err := m.keeper.CreateNominatorPool(*msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgCreateNominatorPoolResponse{Pool: pool}, nil
}

func (m msgServer) DepositToPool(ctx context.Context, msg *types.MsgDepositToPool) (*types.MsgDepositToPoolResponse, error) {
	if msg == nil {
		return nil, errors.New("empty nominator pool deposit request")
	}
	msg.Height = defaultHeight(ctx, msg.Height)
	delegator, err := m.keeper.DepositToPool(*msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgDepositToPoolResponse{Delegator: delegator}, nil
}

func (m msgServer) RequestPoolWithdrawal(ctx context.Context, msg *types.MsgRequestPoolWithdrawal) (*types.MsgRequestPoolWithdrawalResponse, error) {
	if msg == nil {
		return nil, errors.New("empty nominator pool withdrawal request")
	}
	msg.Height = defaultHeight(ctx, msg.Height)
	withdrawal, err := m.keeper.RequestPoolWithdrawal(*msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgRequestPoolWithdrawalResponse{Withdrawal: withdrawal}, nil
}

func (m msgServer) CancelPoolWithdrawal(ctx context.Context, msg *types.MsgCancelPoolWithdrawal) (*types.MsgCancelPoolWithdrawalResponse, error) {
	if msg == nil {
		return nil, errors.New("empty nominator pool withdrawal cancellation request")
	}
	msg.Height = defaultHeight(ctx, msg.Height)
	withdrawal, err := m.keeper.CancelPoolWithdrawal(*msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgCancelPoolWithdrawalResponse{Withdrawal: withdrawal}, nil
}

func (m msgServer) DepositToStakingPool(ctx context.Context, msg *types.MsgDepositToStakingPool) (*types.MsgDepositToStakingPoolResponse, error) {
	if msg == nil {
		return nil, errors.New("empty nominator pool deposit request")
	}
	m.bindRuntimeContext(ctx)
	msg.Height = defaultHeight(ctx, msg.Height)
	receipt, err := m.keeper.DepositToStakingPool(*msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgDepositToStakingPoolResponse{
		PoolID:		receipt.PoolID,
		OwnerAddress:	receipt.OwnerAddress,
		Amount:		receipt.Amount,
		Shares:		receipt.Shares,
		Height:		receipt.Height,
		ReceiptToken:	receipt.ReceiptToken,
	}, nil
}

func (m msgServer) RequestPoolUnbond(ctx context.Context, msg *types.MsgRequestPoolUnbond) (*types.MsgRequestPoolUnbondResponse, error) {
	if msg == nil {
		return nil, errors.New("empty nominator pool unbond request")
	}
	msg.Height = defaultHeight(ctx, msg.Height)
	receipt, err := m.keeper.RequestPoolUnbond(*msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgRequestPoolUnbondResponse{
		PoolID:		receipt.PoolID,
		OwnerAddress:	receipt.OwnerAddress,
		RequestID:	receipt.RequestID,
		Shares:		receipt.Shares,
		Amount:		receipt.Amount,
		CompleteHeight:	receipt.CompleteHeight,
	}, nil
}

func (m msgServer) WithdrawPoolStake(ctx context.Context, msg *types.MsgWithdrawPoolStake) (*types.MsgWithdrawPoolStakeResponse, error) {
	if msg == nil {
		return nil, errors.New("empty nominator pool withdrawal request")
	}
	msg.Height = defaultHeight(ctx, msg.Height)
	receipt, err := m.keeper.WithdrawPoolStake(*msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgWithdrawPoolStakeResponse{
		PoolID:		receipt.PoolID,
		OwnerAddress:	receipt.OwnerAddress,
		RequestID:	receipt.RequestID,
		Amount:		receipt.Amount,
		Height:		receipt.Height,
	}, nil
}

func (m msgServer) TopUpPoolReserve(ctx context.Context, msg *types.MsgTopUpPoolReserve) (*types.MsgTopUpPoolReserveResponse, error) {
	if msg == nil {
		return nil, errors.New("empty nominator pool top-up request")
	}
	msg.Height = defaultHeight(ctx, msg.Height)
	receipt, err := m.keeper.TopUpPoolReserve(*msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgTopUpPoolReserveResponse{
		PoolID:			receipt.PoolID,
		PayerAddress:		receipt.PayerAddress,
		Amount:			receipt.Amount,
		StorageDebtPaid:	receipt.StorageDebtPaid,
	}, nil
}

func (m msgServer) ClaimPoolRewards(ctx context.Context, msg *types.MsgClaimPoolRewards) (*types.MsgClaimPoolRewardsResponse, error) {
	if msg == nil {
		return nil, errors.New("empty nominator pool reward claim request")
	}
	msg.Height = defaultHeight(ctx, msg.Height)
	receipt, err := m.keeper.ClaimPoolRewardsWithReceipt(*msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgClaimPoolRewardsResponse{
		PoolID:		receipt.PoolID,
		OwnerAddress:	receipt.OwnerAddress,
		Amount:		receipt.Amount,
		Epoch:		receipt.Epoch,
	}, nil
}

func (m msgServer) SyncPoolRewards(ctx context.Context, msg *types.MsgSyncPoolRewards) (*types.MsgSyncPoolRewardsResponse, error) {
	if msg == nil {
		return nil, errors.New("empty nominator pool reward sync request")
	}
	msg.Height = defaultHeight(ctx, msg.Height)
	summary, err := m.keeper.SyncPoolRewards(*msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgSyncPoolRewardsResponse{
		PoolUserRewards:	summary.PoolUserRewards,
		ValidatorCommission:	summary.ValidatorCommission,
		PoolProtocolFee:	summary.PoolProtocolFee,
		RewardIndexAfter:	summary.RewardIndexAfter,
	}, nil
}

func (m msgServer) ClaimStakingRewards(ctx context.Context, msg *types.MsgClaimStakingRewards) (*types.MsgClaimStakingRewardsResponse, error) {
	if msg == nil {
		return nil, errors.New("empty staking reward claim request")
	}
	msg.Height = defaultHeight(ctx, msg.Height)
	amount, err := m.keeper.ClaimStakingRewards(*msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgClaimStakingRewardsResponse{RewardAmount: amount}, nil
}

func (m msgServer) ClaimStakeReputation(ctx context.Context, msg *types.MsgClaimStakeReputation) (*types.MsgClaimStakeReputationResponse, error) {
	if msg == nil {
		return nil, errors.New("empty stake reputation claim request")
	}
	msg.Height = defaultHeight(ctx, msg.Height)
	receipt, err := m.keeper.ClaimStakeReputation(*msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgClaimStakeReputationResponse{
		Account:		receipt.Account,
		PoolID:			receipt.PoolID,
		ReputationDelta:	receipt.ReputationDelta,
		ReputationScore:	receipt.ReputationScore,
	}, nil
}

func (m msgServer) DelegateToValidator(ctx context.Context, msg *types.MsgDelegateToValidator) (*types.MsgDelegateToValidatorResponse, error) {
	if msg == nil {
		return nil, errors.New("empty direct validator delegation request")
	}
	msg.Height = defaultHeight(ctx, msg.Height)
	if msg.Authority == "" {
		msg.Authority = m.keeper.genesis.Params.Authority
	}
	if err := m.keeper.DelegateUserToValidator(*msg); err != nil {
		return nil, err
	}
	return &types.MsgDelegateToValidatorResponse{}, nil
}

func (m msgServer) RegisterValidator(ctx context.Context, msg *types.MsgRegisterValidator) (*types.MsgRegisterValidatorResponse, error) {
	if msg == nil {
		return nil, errors.New("empty validator registration request")
	}
	msg.Height = defaultHeight(ctx, msg.Height)
	receipt, err := m.keeper.RegisterValidator(*msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgRegisterValidatorResponse{Validator: receipt.Validator, Status: receipt.Status, SelfStake: receipt.SelfStake, PoolStake: receipt.PoolStake}, nil
}

func (m msgServer) UpdateValidator(ctx context.Context, msg *types.MsgUpdateValidator) (*types.MsgUpdateValidatorResponse, error) {
	if msg == nil {
		return nil, errors.New("empty validator update request")
	}
	msg.Height = defaultHeight(ctx, msg.Height)
	receipt, err := m.keeper.UpdateValidator(*msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgUpdateValidatorResponse{Validator: receipt.Validator, Status: receipt.Status, SelfStake: receipt.SelfStake, PoolStake: receipt.PoolStake}, nil
}

func (m msgServer) UpdateStakingParams(ctx context.Context, msg *types.MsgUpdateStakingParams) (*types.MsgUpdateStakingParamsResponse, error) {
	if msg == nil {
		return nil, errors.New("empty staking params update request")
	}
	msg.Height = defaultHeight(ctx, msg.Height)
	if _, err := m.keeper.UpdateStakingParams(*msg); err != nil {
		return nil, err
	}
	return &types.MsgUpdateStakingParamsResponse{}, nil
}

func (m msgServer) CreateOfficialLiquidStakingPool(ctx context.Context, msg *types.MsgCreateOfficialLiquidStakingPool) (*types.MsgCreateOfficialLiquidStakingPoolResponse, error) {
	if msg == nil {
		return nil, errors.New("empty official liquid staking pool creation request")
	}
	m.bindRuntimeContext(ctx)
	msg.Height = defaultHeight(ctx, msg.Height)
	pool, err := m.keeper.CreateOfficialLiquidStakingPool(*msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgCreateOfficialLiquidStakingPoolResponse{
		PoolID:			pool.PoolID,
		ContractAddressUser:	pool.ContractAddressUser,
		ContractAddressRaw:	pool.ContractAddressRaw,
	}, nil
}

func (m msgServer) UpdatePoolCommission(ctx context.Context, msg *types.MsgUpdatePoolCommission) (*types.MsgUpdatePoolCommissionResponse, error) {
	if msg == nil {
		return nil, errors.New("empty nominator pool commission update request")
	}
	msg.Height = defaultHeight(ctx, msg.Height)
	pool, err := m.keeper.UpdatePoolCommission(*msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgUpdatePoolCommissionResponse{Pool: pool}, nil
}

func (m msgServer) ChangePoolValidator(ctx context.Context, msg *types.MsgChangePoolValidator) (*types.MsgChangePoolValidatorResponse, error) {
	if msg == nil {
		return nil, errors.New("empty nominator pool validator change request")
	}
	msg.Height = defaultHeight(ctx, msg.Height)
	pool, err := m.keeper.ChangePoolValidator(*msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgChangePoolValidatorResponse{Pool: pool}, nil
}

func defaultHeight(ctx context.Context, provided uint64) uint64 {
	if provided != 0 {
		return provided
	}
	height := sdk.UnwrapSDKContext(ctx).BlockHeight()
	if height <= 0 {
		return 1
	}
	return uint64(height)
}

func (m msgServer) bindRuntimeContext(ctx context.Context) {
	if ctx != nil {
		m.keeper.runtimeCtx = ctx
	}
}
