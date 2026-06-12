package types

import errorsmod "cosmossdk.io/errors"

var (
	ErrInvalidParams	= errorsmod.Register(ModuleName, 2, "invalid economics params")
	ErrInvalidState		= errorsmod.Register(ModuleName, 3, "invalid economics state")
	ErrUnauthorized		= errorsmod.Register(ModuleName, 4, "unauthorized economics operation")
	ErrNotFound		= errorsmod.Register(ModuleName, 5, "economics record not found")
)
