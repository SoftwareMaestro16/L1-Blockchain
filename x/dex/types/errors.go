package types

import errorsmod "cosmossdk.io/errors"

var (
	ErrInvalidPool       = errorsmod.Register(ModuleName, 2, "invalid pool")
	ErrPoolNotFound      = errorsmod.Register(ModuleName, 3, "pool not found")
	ErrInvalidLiquidity  = errorsmod.Register(ModuleName, 4, "invalid liquidity")
	ErrSlippage          = errorsmod.Register(ModuleName, 5, "slippage exceeded")
	ErrInvalidParams     = errorsmod.Register(ModuleName, 6, "invalid params")
	ErrUnauthorized      = errorsmod.Register(ModuleName, 7, "unauthorized")
	ErrOperationDisabled = errorsmod.Register(ModuleName, 8, "operation disabled")
)
