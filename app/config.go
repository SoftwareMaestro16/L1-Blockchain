package app

import (
	"github.com/sovereign-l1/l1/app/appconfig"
)

const appName = appconfig.AppName

const (
	SDKBech32AccountPrefix	= appconfig.SDKBech32AccountPrefix
	BondDenom		= appconfig.BondDenom
)

const (
	AccountAddressPrefix	= appconfig.AccountAddressPrefix
	ValidatorAddressPrefix	= appconfig.ValidatorAddressPrefix
	ConsensusAddressPrefix	= appconfig.ConsensusAddressPrefix
)

// DefaultNodeHome default home directories for the application daemon
var DefaultNodeHome = appconfig.ConfigureSDK(".aetra")
