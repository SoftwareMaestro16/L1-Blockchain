package keeper

import (
	"context"
	"errors"

	nativeaccount "github.com/sovereign-l1/l1/x/native-account/types"
)

var _ nativeaccount.MsgServer = msgServer{}

type msgServer struct{ keeper Keeper }

func NewMsgServerImpl(k Keeper) nativeaccount.MsgServer {
	return msgServer{keeper: k}
}

func (m msgServer) ActivateAccount(ctx context.Context, msg *nativeaccount.MsgActivateAccount) (*nativeaccount.MsgActivateAccountResponse, error) {
	if msg == nil {
		return nil, errors.New("empty native account activation request")
	}
	result, err := m.keeper.ActivateAccount(ctx, *msg)
	if err != nil {
		return nil, err
	}
	return &nativeaccount.MsgActivateAccountResponse{
		AddressUser:   result.Account.AddressUser,
		AddressRaw:    result.Account.AddressRaw,
		AccountNumber: result.Account.AccountNumber,
		Sequence:      result.Account.Sequence,
	}, nil
}
