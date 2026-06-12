package keeper

import (
	"context"
	"errors"

	"github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/x/system-registry/types"
)

var _ types.MsgServer = grpcMsgServer{}
var _ types.QueryServer = grpcQueryServer{}

type grpcMsgServer struct{ keeper *Keeper }
type grpcQueryServer struct{ keeper *Keeper }

func NewGRPCMsgServer(k *Keeper) types.MsgServer	{ return grpcMsgServer{keeper: k} }
func NewGRPCQueryServer(k *Keeper) types.QueryServer	{ return grpcQueryServer{keeper: k} }

func (s grpcMsgServer) RegisterSystemEntity(_ context.Context, msg *types.MsgRegisterSystemEntity) (*types.MsgRegisterSystemEntityResponse, error) {
	if msg == nil {
		return nil, errors.New("empty system entity registration")
	}
	entity, event, err := s.keeper.RegisterSystemEntity(*msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgRegisterSystemEntityResponse{Entity: entity, Event: event}, nil
}

func (s grpcMsgServer) UpdateSystemEntity(_ context.Context, msg *types.MsgUpdateSystemEntity) (*types.MsgUpdateSystemEntityResponse, error) {
	if msg == nil {
		return nil, errors.New("empty system entity update")
	}
	entity, event, err := s.keeper.UpdateSystemEntity(*msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgUpdateSystemEntityResponse{Entity: entity, Event: event}, nil
}

func (s grpcMsgServer) PauseSystemEntity(_ context.Context, msg *types.MsgPauseSystemEntity) (*types.MsgPauseSystemEntityResponse, error) {
	if msg == nil {
		return nil, errors.New("empty system entity pause")
	}
	entity, event, err := s.keeper.PauseSystemEntity(*msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgPauseSystemEntityResponse{Entity: entity, Event: event}, nil
}

func (s grpcMsgServer) ResumeSystemEntity(_ context.Context, msg *types.MsgResumeSystemEntity) (*types.MsgResumeSystemEntityResponse, error) {
	if msg == nil {
		return nil, errors.New("empty system entity resume")
	}
	entity, event, err := s.keeper.ResumeSystemEntity(*msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgResumeSystemEntityResponse{Entity: entity, Event: event}, nil
}

func (s grpcMsgServer) DeprecateSystemEntity(_ context.Context, msg *types.MsgDeprecateSystemEntity) (*types.MsgDeprecateSystemEntityResponse, error) {
	if msg == nil {
		return nil, errors.New("empty system entity deprecation")
	}
	entity, event, err := s.keeper.DeprecateSystemEntity(*msg)
	if err != nil {
		return nil, err
	}
	return &types.MsgDeprecateSystemEntityResponse{Entity: entity, Event: event}, nil
}

func (s grpcQueryServer) Params(context.Context, *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	return &types.QueryParamsResponse{Params: s.keeper.ExportGenesis().Params}, nil
}

func (s grpcQueryServer) SystemEntities(_ context.Context, req *types.QuerySystemEntitiesRequest) (*types.QuerySystemEntitiesResponse, error) {
	if req == nil {
		req = &types.QuerySystemEntitiesRequest{}
	}
	entities, err := s.keeper.SystemEntities()
	if err != nil {
		return nil, err
	}
	page, next := paginateEntities(entities, req.Offset, req.Limit)
	return &types.QuerySystemEntitiesResponse{Entities: page, Next: next, Total: uint64(len(entities))}, nil
}

func (s grpcQueryServer) SystemEntity(_ context.Context, req *types.QuerySystemEntityRequest) (*types.QuerySystemEntityResponse, error) {
	if req == nil {
		return nil, errors.New("empty system entity query")
	}
	entity, found, err := s.keeper.SystemEntity(req.ModuleName)
	if err != nil {
		return nil, err
	}
	return &types.QuerySystemEntityResponse{Entity: entity, Found: found}, nil
}

func (s grpcQueryServer) ReservedSystemAddresses(_ context.Context, req *types.QueryReservedSystemAddressesRequest) (*types.QueryReservedSystemAddressesResponse, error) {
	if req == nil {
		req = &types.QueryReservedSystemAddressesRequest{}
	}
	addresses := addressing.AllSystemAddresses()
	page, next := paginateAddresses(addresses, req.Offset, req.Limit)
	return &types.QueryReservedSystemAddressesResponse{Addresses: page, Next: next, Total: uint64(len(addresses))}, nil
}

func (s grpcQueryServer) SystemAddress(_ context.Context, req *types.QuerySystemAddressRequest) (*types.QuerySystemAddressResponse, error) {
	if req == nil {
		return nil, errors.New("empty system address query")
	}
	address, found := addressing.SystemAddressByText(req.Address)
	return &types.QuerySystemAddressResponse{Address: address, Found: found}, nil
}

func (s grpcQueryServer) DependencyGraph(context.Context, *types.QueryDependencyGraphRequest) (*types.QueryDependencyGraphResponse, error) {
	edges, err := s.keeper.DependencyGraph()
	if err != nil {
		return nil, err
	}
	return &types.QueryDependencyGraphResponse{Edges: edges}, nil
}

func paginateEntities(entities []types.SystemEntity, offset, limit uint64) ([]types.SystemEntity, uint64) {
	if limit == 0 || limit > 100 {
		limit = 100
	}
	if offset >= uint64(len(entities)) {
		return []types.SystemEntity{}, 0
	}
	end := offset + limit
	if end > uint64(len(entities)) {
		end = uint64(len(entities))
	}
	next := uint64(0)
	if end < uint64(len(entities)) {
		next = end
	}
	return entities[int(offset):int(end)], next
}

func paginateAddresses(addresses []addressing.SystemAddress, offset, limit uint64) ([]addressing.SystemAddress, uint64) {
	if limit == 0 || limit > 100 {
		limit = 100
	}
	if offset >= uint64(len(addresses)) {
		return []addressing.SystemAddress{}, 0
	}
	end := offset + limit
	if end > uint64(len(addresses)) {
		end = uint64(len(addresses))
	}
	next := uint64(0)
	if end < uint64(len(addresses)) {
		next = end
	}
	return addresses[int(offset):int(end)], next
}
