package keeper

import (
	"context"
	"errors"

	nativeaccount "github.com/sovereign-l1/l1/x/native-account/types"
)

var _ nativeaccount.MsgServer = msgServer{}

type msgServer struct {
	nativeaccount.UnimplementedMsgServer
	keeper	Keeper
}

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
		AddressUser:	result.Account.AddressUser,
		AddressRaw:	result.Account.AddressRaw,
		AccountNumber:	result.Account.AccountNumber,
		Sequence:	result.Account.Sequence,
	}, nil
}

func (m msgServer) UpdateAuthPolicy(ctx context.Context, msg *nativeaccount.MsgUpdateAuthPolicy) (*nativeaccount.MsgUpdateAuthPolicyResponse, error) {
	if msg == nil {
		return nil, errors.New("empty native account auth policy update request")
	}
	account, err := m.keeper.UpdateAuthPolicy(ctx, *msg)
	if err != nil {
		return nil, err
	}
	return &nativeaccount.MsgUpdateAuthPolicyResponse{AddressUser: account.AddressUser, Sequence: account.Sequence, Status: account.Status}, nil
}

func (m msgServer) RotateKey(ctx context.Context, msg *nativeaccount.MsgRotateKey) (*nativeaccount.MsgRotateKeyResponse, error) {
	if msg == nil {
		return nil, errors.New("empty native account key rotation request")
	}
	account, err := m.keeper.RotateKey(ctx, *msg)
	if err != nil {
		return nil, err
	}
	return &nativeaccount.MsgRotateKeyResponse{AddressUser: account.AddressUser, Sequence: account.Sequence, Status: account.Status}, nil
}

func (m msgServer) RecoverAccount(ctx context.Context, msg *nativeaccount.MsgRecoverAccount) (*nativeaccount.MsgRecoverAccountResponse, error) {
	if msg == nil {
		return nil, errors.New("empty native account recovery request")
	}
	account, err := m.keeper.RecoverAccount(ctx, *msg)
	if err != nil {
		return nil, err
	}
	return &nativeaccount.MsgRecoverAccountResponse{AddressUser: account.AddressUser, Sequence: account.Sequence, Status: account.Status}, nil
}

func (m msgServer) FreezeAccount(ctx context.Context, msg *nativeaccount.MsgFreezeAccount) (*nativeaccount.MsgFreezeAccountResponse, error) {
	if msg == nil {
		return nil, errors.New("empty native account freeze request")
	}
	account, err := m.keeper.FreezeAccount(ctx, *msg)
	if err != nil {
		return nil, err
	}
	return &nativeaccount.MsgFreezeAccountResponse{AddressUser: account.AddressUser, Sequence: account.Sequence, Status: account.Status}, nil
}

func (m msgServer) PayStorageDebt(ctx context.Context, msg *nativeaccount.MsgPayStorageDebt) (*nativeaccount.MsgPayStorageDebtResponse, error) {
	if msg == nil {
		return nil, errors.New("empty native account storage debt payment request")
	}
	account, err := m.keeper.PayStorageDebt(ctx, *msg)
	if err != nil {
		return nil, err
	}
	return &nativeaccount.MsgPayStorageDebtResponse{AddressUser: account.AddressUser, StorageRentDebt: account.StorageRentDebt, Status: account.Status}, nil
}

func (m msgServer) UnfreezeAccount(ctx context.Context, msg *nativeaccount.MsgUnfreezeAccount) (*nativeaccount.MsgUnfreezeAccountResponse, error) {
	if msg == nil {
		return nil, errors.New("empty native account unfreeze request")
	}
	account, err := m.keeper.UnfreezeAccount(ctx, *msg)
	if err != nil {
		return nil, err
	}
	return &nativeaccount.MsgUnfreezeAccountResponse{AddressUser: account.AddressUser, StorageRentDebt: account.StorageRentDebt, Status: account.Status}, nil
}

func (m msgServer) UpdateAccountMetadata(ctx context.Context, msg *nativeaccount.MsgUpdateAccountMetadata) (*nativeaccount.MsgUpdateAccountMetadataResponse, error) {
	if msg == nil {
		return nil, errors.New("empty native account metadata update request")
	}
	account, err := m.keeper.UpdateAccountMetadata(ctx, *msg)
	if err != nil {
		return nil, err
	}
	return &nativeaccount.MsgUpdateAccountMetadataResponse{AddressUser: account.AddressUser, Sequence: account.Sequence, Status: account.Status}, nil
}
