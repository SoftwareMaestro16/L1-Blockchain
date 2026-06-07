package app

import (
	"github.com/sovereign-l1/l1/app/appconfig"
)

const appName = appconfig.AppName

const (
	AccountAddressPrefix   = appconfig.AccountAddressPrefix
	ValidatorAddressPrefix = appconfig.ValidatorAddressPrefix
	ConsensusAddressPrefix = appconfig.ConsensusAddressPrefix
	BondDenom              = appconfig.BondDenom
)

// DefaultNodeHome default home directories for the application daemon
var DefaultNodeHome = appconfig.ConfigureSDK(".aetra")
