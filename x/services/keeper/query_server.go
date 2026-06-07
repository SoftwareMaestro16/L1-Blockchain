package keeper

import (
	"context"
	"errors"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
	servicestypes "github.com/sovereign-l1/l1/x/services/types"
)

var _ servicestypes.QueryServer = Keeper{}

func (k Keeper) Service(ctx context.Context, req *servicestypes.QueryService) (*servicestypes.QueryServiceResponse, error) {
	if req == nil {
		return nil, errors.New("services empty service query")
	}
	res, err := k.genesis.Registry.QueryService(*req)
	return &res, err
}

func (k Keeper) ServiceByName(ctx context.Context, req *servicestypes.QueryServiceByName) (*servicestypes.QueryServiceResponse, error) {
	if req == nil {
		return nil, errors.New("services empty service-by-name query")
	}
	res, err := k.genesis.Registry.QueryServiceByName(*req)
	return &res, err
}

func (k Keeper) ServicesByOwner(ctx context.Context, req *servicestypes.QueryServicesByOwner) (*servicestypes.QueryServicesResponse, error) {
	if req == nil {
		return nil, errors.New("services empty services-by-owner query")
	}
	res, err := k.genesis.Registry.QueryServicesByOwner(*req, k.genesis.Params)
	return &res, err
}

func (k Keeper) ServicesByIdentity(ctx context.Context, req *servicestypes.QueryServicesByIdentity) (*servicestypes.QueryServicesResponse, error) {
	if req == nil {
		return nil, errors.New("services empty services-by-identity query")
	}
	res, err := k.genesis.Registry.QueryServicesByIdentity(*req, k.genesis.Params)
	return &res, err
}

func (k Keeper) ProvidersByService(ctx context.Context, req *servicestypes.QueryProvidersByService) (*servicestypes.QueryProvidersResponse, error) {
	if req == nil {
		return nil, errors.New("services empty providers-by-service query")
	}
	res, err := k.genesis.Registry.QueryProvidersByService(*req, k.genesis.Params)
	return &res, err
}

func (k Keeper) ServiceInterface(ctx context.Context, req *servicestypes.QueryServiceInterface) (*servicestypes.QueryServiceInterfaceResponse, error) {
	if req == nil {
		return nil, errors.New("services empty interface query")
	}
	res, err := k.genesis.Registry.QueryServiceInterface(*req)
	return &res, err
}

func (k Keeper) ServicePaymentModel(ctx context.Context, req *servicestypes.QueryServicePaymentModel) (*servicestypes.QueryServicePaymentModelResponse, error) {
	if req == nil {
		return nil, errors.New("services empty payment model query")
	}
	res, err := k.genesis.Registry.QueryServicePaymentModel(*req)
	return &res, err
}

func (k Keeper) ServiceVerificationModel(ctx context.Context, req *servicestypes.QueryServiceVerificationModel) (*servicestypes.QueryServiceVerificationModelResponse, error) {
	if req == nil {
		return nil, errors.New("services empty verification model query")
	}
	res, err := k.genesis.Registry.QueryServiceVerificationModel(*req)
	return &res, err
}

func (k Keeper) ServiceReceipt(ctx context.Context, req *servicestypes.QueryServiceReceipt) (*servicestypes.QueryServiceReceiptResponse, error) {
	if req == nil {
		return nil, errors.New("services empty receipt query")
	}
	res, err := k.genesis.Registry.QueryServiceReceipt(*req)
	return &res, err
}

func (k Keeper) ServiceProof(ctx context.Context, req *servicestypes.QueryServiceProof) (*servicestypes.QueryServiceProofResponse, error) {
	if req == nil {
		return nil, errors.New("services empty proof query")
	}
	res, err := k.genesis.Registry.QueryServiceProof(*req)
	return &res, err
}

func (k Keeper) ServiceParams(ctx context.Context, req *servicestypes.QueryServiceParams) (*servicestypes.QueryServiceParamsResponse, error) {
	if req == nil {
		return nil, errors.New("services empty params query")
	}
	res, err := coretypes.QueryServiceParamsResponseFor(k.genesis.Params)
	return &res, err
}
