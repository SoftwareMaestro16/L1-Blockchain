package types

import (
	"fmt"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"

	"github.com/sovereign-l1/l1/app/addressing"
)

type MsgActivateAccount struct {
	AddressUser string
	PublicKey   cryptotypes.PubKey
}

func (m MsgActivateAccount) ValidateBasic() error {
	if err := addressing.ValidateUserAddress("activation address", m.AddressUser); err != nil {
		return err
	}
	pair, err := addressing.DeriveAccountAddress(m.PublicKey)
	if err != nil {
		return err
	}
	if pair.User != m.AddressUser {
		return fmt.Errorf("activation address must equal derived public key address")
	}
	return nil
}

func ActivationAddressPair(pubKey cryptotypes.PubKey) (addressing.AddressPair, error) {
	return addressing.DeriveAccountAddress(pubKey)
}
