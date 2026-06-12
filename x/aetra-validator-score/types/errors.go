package types

import errorsmod "cosmossdk.io/errors"

var (
	ErrInvalidParams	= errorsmod.Register(ModuleName, 2, "invalid validator score params")
	ErrInvalidScore		= errorsmod.Register(ModuleName, 3, "invalid validator score state")
	ErrUnauthorized		= errorsmod.Register(ModuleName, 4, "unauthorized validator score operation")
	ErrNotFound		= errorsmod.Register(ModuleName, 5, "validator score record not found")
)
