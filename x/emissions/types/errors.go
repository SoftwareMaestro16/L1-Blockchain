package types

import errorsmod "cosmossdk.io/errors"

var (
	ErrInvalidParams	= errorsmod.Register(ModuleName, 2, "invalid emissions params")
	ErrInvalidEpoch		= errorsmod.Register(ModuleName, 3, "invalid emission epoch")
	ErrUnauthorized		= errorsmod.Register(ModuleName, 4, "unauthorized emissions operation")
	ErrDuplicateEpoch	= errorsmod.Register(ModuleName, 5, "emission epoch already finalized")
	ErrAccounting		= errorsmod.Register(ModuleName, 6, "emissions accounting invariant violation")
	ErrNotFound		= errorsmod.Register(ModuleName, 7, "emission epoch not found")
)
