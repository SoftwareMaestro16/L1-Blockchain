package types

import errorsmod "cosmossdk.io/errors"

var (
	ErrInvalidParams	= errorsmod.Register(ModuleName, 2, "invalid treasury params")
	ErrInvalidSpend		= errorsmod.Register(ModuleName, 3, "invalid treasury spend")
	ErrUnauthorized		= errorsmod.Register(ModuleName, 4, "unauthorized treasury operation")
	ErrInsufficientFunds	= errorsmod.Register(ModuleName, 5, "insufficient treasury funds")
	ErrSpendCapExceeded	= errorsmod.Register(ModuleName, 6, "treasury spend cap exceeded")
	ErrNotFound		= errorsmod.Register(ModuleName, 7, "treasury spend not found")
	ErrAccounting		= errorsmod.Register(ModuleName, 8, "treasury accounting invariant violation")
)
