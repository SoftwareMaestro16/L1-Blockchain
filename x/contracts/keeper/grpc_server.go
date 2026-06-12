package keeper

import (
	"context"
	"errors"

	"github.com/sovereign-l1/l1/x/contracts/types"
)

var (
	_	types.GRPCMsgServer	= grpcMsgServer{}
	_	types.GRPCQueryServer	= grpcQueryServer{}
)

type grpcMsgServer struct {
	types.UnimplementedGRPCMsgServer
	keeper	*Keeper
}

type grpcQueryServer struct {
	types.UnimplementedGRPCQueryServer
	keeper	*Keeper
}

func NewGRPCMsgServer(k *Keeper) types.GRPCMsgServer {
	return grpcMsgServer{keeper: k}
}

func NewGRPCQueryServer(k *Keeper) types.GRPCQueryServer {
	return grpcQueryServer{keeper: k}
}

func (m grpcMsgServer) StoreCode(ctx context.Context, msg *types.MsgStoreCode) (*types.StoreCodeResponse, error) {
	if msg == nil {
		return nil, errors.New("empty contracts store code request")
	}
	res, err := m.keeper.StoreCodeState(ctx, *msg)
	return &res, err
}

func (m grpcMsgServer) DeployContract(ctx context.Context, msg *types.MsgDeployContract) (*types.InstantiateContractResponse, error) {
	if msg == nil {
		return nil, errors.New("empty contracts deploy request")
	}
	res, err := m.keeper.DeployContractState(ctx, *msg)
	return &res, err
}

func (m grpcMsgServer) ExecuteExternal(ctx context.Context, msg *types.MsgExecuteExternal) (*types.ExecuteContractResponse, error) {
	if msg == nil {
		return nil, errors.New("empty contracts external execution request")
	}
	res, err := m.keeper.ExecuteExternalState(ctx, *msg)
	return &res, err
}

func (m grpcMsgServer) ExecuteInternal(ctx context.Context, msg *types.MsgExecuteInternal) (*types.InternalMessage, error) {
	if msg == nil {
		return nil, errors.New("empty contracts internal execution request")
	}
	res, err := m.keeper.ExecuteInternal(*msg)
	if err != nil {
		return nil, err
	}
	if err := m.keeper.writeGenesis(ctx); err != nil {
		return nil, err
	}
	return &res, nil
}

func (m grpcMsgServer) SendInternalMessage(ctx context.Context, msg *types.MsgSendInternalMessage) (*types.InternalMessage, error) {
	if msg == nil {
		return nil, errors.New("empty contracts internal send request")
	}
	res, err := m.keeper.SendInternalMessage(*msg)
	if err != nil {
		return nil, err
	}
	if err := m.keeper.writeGenesis(ctx); err != nil {
		return nil, err
	}
	return &res, nil
}

func (m grpcMsgServer) UpdateContractParams(ctx context.Context, msg *types.MsgUpdateContractParams) (*types.MsgUpdateContractParamsResponse, error) {
	if msg == nil {
		return nil, errors.New("empty contracts params update request")
	}
	if err := m.keeper.UpdateContractParams(*msg); err != nil {
		return nil, err
	}
	if err := m.keeper.writeGenesis(ctx); err != nil {
		return nil, err
	}
	return &types.MsgUpdateContractParamsResponse{StateRoot: m.keeper.ExportGenesis().StateRoot}, nil
}

func (q grpcQueryServer) Params(context.Context, *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	return &types.QueryParamsResponse{Params: q.keeper.Params()}, nil
}

func (q grpcQueryServer) Code(_ context.Context, req *types.QueryCodeRequest) (*types.QueryCodeResponse, error) {
	if req == nil {
		return nil, errors.New("empty contracts code query")
	}
	code, found, err := q.keeper.Code(*req)
	return &types.QueryCodeResponse{Code: code, Found: found}, err
}

func (q grpcQueryServer) Codes(_ context.Context, req *types.QueryCodesRequest) (*types.QueryCodesResponse, error) {
	if req == nil {
		return nil, errors.New("empty contracts codes query")
	}
	codes, err := q.keeper.Codes(*req)
	return &types.QueryCodesResponse{Codes: codes}, err
}

func (q grpcQueryServer) Contract(_ context.Context, req *types.QueryContractRequest) (*types.QueryContractResponse, error) {
	if req == nil {
		return nil, errors.New("empty contracts contract query")
	}
	res, err := q.keeper.Contract(*req)
	return &res, err
}

func (q grpcQueryServer) Contracts(_ context.Context, req *types.QueryContractsRequest) (*types.QueryContractsResponse, error) {
	if req == nil {
		return nil, errors.New("empty contracts list query")
	}
	contracts, err := q.keeper.Contracts(*req)
	return &types.QueryContractsResponse{Contracts: contracts}, err
}

func (q grpcQueryServer) ContractStorage(_ context.Context, req *types.QueryContractStorageRequest) (*types.QueryContractStorageResponse, error) {
	if req == nil {
		return nil, errors.New("empty contracts storage query")
	}
	entries, err := q.keeper.ContractStorage(*req)
	return &types.QueryContractStorageResponse{Entries: entries}, err
}

func (q grpcQueryServer) ContractReceipts(_ context.Context, req *types.QueryContractReceiptsRequest) (*types.QueryContractReceiptsResponse, error) {
	if req == nil {
		return nil, errors.New("empty contracts receipts query")
	}
	receipts, err := q.keeper.ContractReceipts(*req)
	return &types.QueryContractReceiptsResponse{Receipts: receipts}, err
}

func (q grpcQueryServer) ContractQueue(_ context.Context, req *types.QueryContractQueueRequest) (*types.QueryContractQueueResponse, error) {
	if req == nil {
		return nil, errors.New("empty contracts queue query")
	}
	messages, err := q.keeper.ContractQueue(*req)
	return &types.QueryContractQueueResponse{Messages: messages}, err
}

func (q grpcQueryServer) ContractEvents(_ context.Context, req *types.QueryContractEventsRequest) (*types.QueryContractEventsResponse, error) {
	if req == nil {
		return nil, errors.New("empty contracts events query")
	}
	return &types.QueryContractEventsResponse{}, q.keeper.ContractEvents(*req)
}

func (q grpcQueryServer) ContractStateRoot(_ context.Context, req *types.QueryContractStateRootRequest) (*types.QueryContractStateRootResponse, error) {
	if req == nil {
		return nil, errors.New("empty contracts state root query")
	}
	root, err := q.keeper.ContractStateRoot(*req)
	return &types.QueryContractStateRootResponse{StateRoot: root}, err
}
