package types

import errorsmod "cosmossdk.io/errors"

var (
	ErrInvalidParams	= errorsmod.Register(ModuleName, 2, "invalid concentration params")
	ErrInvalidConcentration	= errorsmod.Register(ModuleName, 3, "invalid stake concentration")
	ErrUnauthorized		= errorsmod.Register(ModuleName, 4, "unauthorized concentration operation")
	ErrNotFound		= errorsmod.Register(ModuleName, 5, "validator concentration not found")
)
