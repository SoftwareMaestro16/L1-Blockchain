package types

import errorsmod "cosmossdk.io/errors"

var (
	ErrInvalidParams	= errorsmod.Register(ModuleName, 2, "invalid params")
	ErrInvalidCommission	= errorsmod.Register(ModuleName, 3, "invalid commission")
	ErrUnauthorized		= errorsmod.Register(ModuleName, 4, "unauthorized")
	ErrRateLimited		= errorsmod.Register(ModuleName, 5, "commission rate limit exceeded")
	ErrNotFound		= errorsmod.Register(ModuleName, 6, "validator commission not found")
)
