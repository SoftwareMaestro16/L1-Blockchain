package types

import errorsmod "cosmossdk.io/errors"

var (
	ErrInvalidParams	= errorsmod.Register(ModuleName, 2, "invalid params")
	ErrInvalidFee		= errorsmod.Register(ModuleName, 3, "invalid fee")
	ErrUnauthorized		= errorsmod.Register(ModuleName, 4, "unauthorized")
	ErrAccounting		= errorsmod.Register(ModuleName, 5, "fee accounting invariant failed")
	ErrDuplicateHistory	= errorsmod.Register(ModuleName, 6, "fee history already exists")
	ErrEmptyDistribution	= errorsmod.Register(ModuleName, 7, "empty fee distribution")
)
