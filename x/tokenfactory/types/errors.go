package types

import errorsmod "cosmossdk.io/errors"

var (
	ErrInvalidDenom      = errorsmod.Register(ModuleName, 2, "invalid denom")
	ErrUnauthorized      = errorsmod.Register(ModuleName, 3, "unauthorized")
	ErrDenomExists       = errorsmod.Register(ModuleName, 4, "denom already exists")
	ErrDenomMissing      = errorsmod.Register(ModuleName, 5, "denom not found")
	ErrInvalidParams     = errorsmod.Register(ModuleName, 6, "invalid params")
	ErrOperationDisabled = errorsmod.Register(ModuleName, 7, "operation disabled")
)
