package types

import errorsmod "cosmossdk.io/errors"

var (
	ErrInvalidParams	= errorsmod.Register(ModuleName, 2, "invalid params")
	ErrInvalidFee		= errorsmod.Register(ModuleName, 3, "invalid fee")
	ErrUnauthorized		= errorsmod.Register(ModuleName, 4, "unauthorized")
)
