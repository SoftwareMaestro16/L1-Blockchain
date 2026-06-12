package types

import errorsmod "cosmossdk.io/errors"

var (
	ErrInvalidParams	= errorsmod.Register(ModuleName, 2, "invalid staking policy params")
	ErrInvalidPolicy	= errorsmod.Register(ModuleName, 3, "invalid staking policy")
	ErrUnauthorized		= errorsmod.Register(ModuleName, 4, "unauthorized staking policy operation")
	ErrNotFound		= errorsmod.Register(ModuleName, 5, "staking policy record not found")
)
