package keeper

import (
	"context"
	"errors"

	"github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/x/internal/prototype"
	nativeaccount "github.com/sovereign-l1/l1/x/native-account/types"
)

var _ nativeaccount.QueryServer = queryServer{}

type queryServer struct{ keeper Keeper }

func NewQueryServerImpl(k Keeper) nativeaccount.QueryServer	{ return queryServer{keeper: k} }

func (q queryServer) Account(ctx context.Context, req *nativeaccount.QueryAccountRequest) (*nativeaccount.QueryAccountResponse, error) {
	if req == nil {
		return nil, errors.New("empty native account query request")
	}
	account, found, err := q.keeper.AccountByUser(ctx, req.Address)
	if err != nil {
		return nil, err
	}
	if found {
		resp := nativeaccount.NewQueryAccountResponse(account, true, false)
		return &resp, nil
	}
	view, err := virtualAccountView(req.Address)
	if err != nil {
		return nil, err
	}
	return &nativeaccount.QueryAccountResponse{
		Found:		false,
		Virtual:	true,
		AddressUser:	view.AddressUser,
		AddressRaw:	view.AddressRaw,
		Status:		view.Status,
	}, nil
}

func (q queryServer) AccountByRaw(ctx context.Context, req *nativeaccount.QueryAccountByRawRequest) (*nativeaccount.QueryAccountResponse, error) {
	if req == nil {
		return nil, errors.New("empty native account raw query request")
	}
	account, found, err := q.keeper.AccountByRaw(ctx, req.AddressRaw)
	if err != nil {
		return nil, err
	}
	resp := nativeaccount.NewQueryAccountResponse(account, found, false)
	return &resp, nil
}

func (q queryServer) VirtualAccount(ctx context.Context, req *nativeaccount.QueryVirtualAccountRequest) (*nativeaccount.QueryVirtualAccountResponse, error) {
	if req == nil {
		return nil, errors.New("empty native account virtual query request")
	}
	if account, found, err := q.keeper.AccountByUser(ctx, req.AddressUser); err != nil {
		return nil, err
	} else if found {
		resp := nativeaccount.NewQueryVirtualAccountResponse(persistentAccountView(account))
		return &resp, nil
	}
	view, err := virtualAccountView(req.AddressUser)
	if err != nil {
		return nil, err
	}
	resp := nativeaccount.NewQueryVirtualAccountResponse(view)
	return &resp, nil
}

func (q queryServer) Params(context.Context, *nativeaccount.QueryParamsRequest) (*nativeaccount.QueryParamsResponse, error) {
	resp := nativeaccount.NewQueryParamsResponse(prototype.DefaultParams())
	return &resp, nil
}

func (q queryServer) AccountStatus(ctx context.Context, req *nativeaccount.QueryAccountStatusRequest) (*nativeaccount.QueryAccountStatusResponse, error) {
	if req == nil {
		return nil, errors.New("empty native account status query request")
	}
	if account, found, err := q.keeper.AccountByUser(ctx, req.Address); err != nil {
		return nil, err
	} else if found {
		view := persistentAccountView(account)
		return &nativeaccount.QueryAccountStatusResponse{
			AddressUser:		view.AddressUser,
			AddressRaw:		view.AddressRaw,
			Status:			view.Status,
			Persistent:		view.Persistent,
			StorageRentActive:	view.StorageRentActive,
			StorageRentDebt:	account.StorageRentDebt,
		}, nil
	}
	view, err := virtualAccountView(req.Address)
	if err != nil {
		return nil, err
	}
	return &nativeaccount.QueryAccountStatusResponse{
		AddressUser:		view.AddressUser,
		AddressRaw:		view.AddressRaw,
		Status:			view.Status,
		Persistent:		view.Persistent,
		StorageRentActive:	view.StorageRentActive,
	}, nil
}

func virtualAccountView(userAddress string) (nativeaccount.VirtualAccountView, error) {
	pair, err := addressing.PairFromUserAddress(addressing.AddressRoleAccount, userAddress)
	if err != nil {
		return nativeaccount.VirtualAccountView{}, err
	}
	return nativeaccount.VirtualAccountView{
		AddressUser:		pair.User,
		AddressRaw:		pair.Raw,
		Status:			nativeaccount.VirtualAccountStatusInactive,
		Persistent:		false,
		StorageRentActive:	false,
	}, nil
}

func persistentAccountView(account nativeaccount.Account) nativeaccount.VirtualAccountView {
	view, err := nativeaccount.NormalizePersistentAccountView(nativeaccount.VirtualAccountView{
		AddressUser:	account.AddressUser,
		AddressRaw:	account.AddressRaw,
		Status:		account.Status,
	})
	if err != nil {
		return nativeaccount.VirtualAccountView{}
	}
	return view
}
