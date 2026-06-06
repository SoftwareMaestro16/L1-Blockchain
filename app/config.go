package app

import (
	clienthelpers "cosmossdk.io/client/v2/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/sovereign-l1/l1/app/params"
)

const appName = appparams.ChainName

const (
	AccountAddressPrefix   = "ae"
	ValidatorAddressPrefix = "aevaloper"
	ConsensusAddressPrefix = "aevalcons"
	BondDenom              = appparams.BaseDenom
)

// DefaultNodeHome default home directories for the application daemon
var DefaultNodeHome string

func init() {
	var err error
	DefaultNodeHome, err = clienthelpers.GetNodeHomeDirectory(".aetra")
	if err != nil {
		panic(err)
	}
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount(AccountAddressPrefix, AccountAddressPrefix+"pub")
	cfg.SetBech32PrefixForValidator(ValidatorAddressPrefix, ValidatorAddressPrefix+"pub")
	cfg.SetBech32PrefixForConsensusNode(ConsensusAddressPrefix, ConsensusAddressPrefix+"pub")
	sdk.DefaultBondDenom = BondDenom
}
