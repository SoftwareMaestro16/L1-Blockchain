package keeper

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/sovereign-l1/l1/x/nominator-pool/types"
)

var _ types.QueryServer = queryServer{}

type queryServer struct{ keeper *Keeper }

func NewQueryServerImpl(k *Keeper) types.QueryServer	{ return queryServer{keeper: k} }

func (q queryServer) NominatorPool(_ context.Context, req *types.QueryNominatorPoolRequest) (*types.QueryNominatorPoolResponse, error) {
	if req == nil {
		return nil, errors.New("empty nominator pool query")
	}
	pool, found := q.keeper.NominatorPool(req.PoolID)
	if !found {
		return nil, errors.New("nominator pool not found")
	}
	return &types.QueryNominatorPoolResponse{Pool: pool}, nil
}

func (q queryServer) NominatorPools(_ context.Context, req *types.QueryNominatorPoolsRequest) (*types.QueryNominatorPoolsResponse, error) {
	if req == nil {
		return nil, errors.New("empty nominator pools query")
	}
	all := q.keeper.NominatorPools()
	total := uint64(len(all))

	off := req.Offset
	tlimit := req.Limit
	if tlimit == 0 {

		tlimit = total
	}
	if off >= total {
		return &types.QueryNominatorPoolsResponse{Pools: []types.NominatorPool{}, NextOffset: 0, Total: total}, nil
	}
	end := off + tlimit
	if end > total {
		end = total
	}

	res := all[off:end]
	next := uint64(0)
	if end < total {
		next = end
	}
	return &types.QueryNominatorPoolsResponse{Pools: res, NextOffset: next, Total: total}, nil
}

func (q queryServer) PoolDelegator(_ context.Context, req *types.QueryPoolDelegatorRequest) (*types.QueryPoolDelegatorResponse, error) {
	if req == nil {
		return nil, errors.New("empty pool delegator query")
	}
	delegator, found := q.keeper.PoolDelegator(req.PoolID, req.Delegator)
	if !found {
		return nil, errors.New("pool delegator not found")
	}
	return &types.QueryPoolDelegatorResponse{Delegator: delegator}, nil
}

func (q queryServer) PoolRewards(_ context.Context, req *types.QueryPoolRewardsRequest) (*types.QueryPoolRewardsResponse, error) {
	if req == nil {
		return nil, errors.New("empty pool rewards query")
	}
	amount, found := q.keeper.PoolRewards(req.PoolID, req.Delegator)
	if !found {
		return nil, errors.New("pool rewards not found")
	}
	return &types.QueryPoolRewardsResponse{RewardAmount: amount}, nil
}

func (q queryServer) PoolShare(_ context.Context, req *types.QueryPoolShareRequest) (*types.QueryPoolShareResponse, error) {
	if req == nil {
		return nil, errors.New("empty pool share query")
	}
	res, found := q.keeper.PoolShare(*req)
	if !found {
		return nil, errors.New("pool share not found")
	}
	return &res, nil
}

func (q queryServer) StakingProof(_ context.Context, req *types.QueryStakingProofRequest) (*types.QueryStakingProofResponse, error) {
	if req == nil {
		return nil, errors.New("empty staking proof query")
	}
	metadata, err := q.keeper.StakingProof(types.StakingProofRequest{
		Kind:		types.StakingProofKind(req.Kind),
		Height:		req.Height,
		PoolID:		req.PoolID,
		Account:	req.Account,
		Epoch:		req.Epoch,
		AppHash:	req.AppHash,
		RootHash:	req.RootHash,
	})
	if err != nil {
		return nil, err
	}
	bz, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}
	return &types.QueryStakingProofResponse{MetadataJSON: string(bz)}, nil
}

func (q queryServer) PoolUnbondingQueue(_ context.Context, req *types.QueryPoolUnbondingQueueRequest) (*types.QueryPoolUnbondingQueueResponse, error) {
	if req == nil {
		return nil, errors.New("empty pool unbonding queue query")
	}
	return &types.QueryPoolUnbondingQueueResponse{UnbondingQueue: q.keeper.PoolUnbondingQueue(req.PoolID)}, nil
}

func (q queryServer) PoolAllocations(_ context.Context, req *types.QueryPoolAllocationsRequest) (*types.QueryPoolAllocationsResponse, error) {
	if req == nil {
		return nil, errors.New("empty pool allocations query")
	}
	res, found := q.keeper.PoolAllocations(*req)
	if !found {
		return nil, errors.New("nominator pool not found")
	}
	return &res, nil
}

func (q queryServer) StakeReputation(_ context.Context, _ *types.QueryStakeReputationRequest) (*types.QueryStakeReputationResponse, error) {
	return nil, errors.New("wallet-facing stake reputation query removed from nominator-pool; use x/reputation IdentityReputationQuery")
}

func (q queryServer) AccountReputation(_ context.Context, _ *types.QueryAccountReputationRequest) (*types.QueryAccountReputationResponse, error) {
	return nil, errors.New("wallet-facing account reputation query removed from nominator-pool; use x/reputation IdentityReputationQuery")
}

func (q queryServer) StakingRewards(_ context.Context, req *types.QueryStakingRewardsRequest) (*types.QueryStakingRewardsResponse, error) {
	if req == nil {
		return nil, errors.New("empty staking rewards query")
	}
	res, err := q.keeper.StakingRewards(*req)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
