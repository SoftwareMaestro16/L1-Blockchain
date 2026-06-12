package types

import errorsmod "cosmossdk.io/errors"

var (
	ErrInvalidParams	= errorsmod.Register(ModuleName, 2, "invalid params")
	ErrInvalidBurn		= errorsmod.Register(ModuleName, 3, "invalid burn")
	ErrUnauthorized		= errorsmod.Register(ModuleName, 4, "unauthorized burn")
)
