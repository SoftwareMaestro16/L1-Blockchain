package addressing

import (
	"errors"
	"fmt"
	"strings"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

type AddressRole string

const (
	AddressRoleAccount	AddressRole	= "account"
	AddressRoleValidator	AddressRole	= "validator"
	AddressRoleConsensus	AddressRole	= "consensus"
)

type AddressPair struct {
	Role	AddressRole
	User	string
	Raw	string
}

func DeriveAccountAddress(pubKey cryptotypes.PubKey) (AddressPair, error) {
	return deriveAddressPair(AddressRoleAccount, pubKey)
}

func DeriveValidatorAddress(pubKey cryptotypes.PubKey) (AddressPair, error) {
	return deriveAddressPair(AddressRoleValidator, pubKey)
}

func DeriveConsensusAddress(pubKey cryptotypes.PubKey) (AddressPair, error) {
	return deriveAddressPair(AddressRoleConsensus, pubKey)
}

func PairFromUserAddress(role AddressRole, userAddress string) (AddressPair, error) {
	userAddress = strings.TrimSpace(userAddress)
	if !strings.HasPrefix(userAddress, UserFriendlyPrefix) {
		return AddressPair{}, fmt.Errorf("%s address must use AE user-facing address format", role)
	}
	bz, err := Parse(userAddress)
	if err != nil {
		return AddressPair{}, err
	}
	return addressPairFromBytes(role, bz)
}

func PairFromRawAddress(role AddressRole, rawAddress string) (AddressPair, error) {
	rawAddress = strings.TrimSpace(rawAddress)
	if !strings.HasPrefix(rawAddress, RawPrefix) {
		return AddressPair{}, fmt.Errorf("%s raw address must use 4: internal address format", role)
	}
	bz, err := Parse(rawAddress)
	if err != nil {
		return AddressPair{}, err
	}
	return addressPairFromBytes(role, bz)
}

func (p AddressPair) Validate() error {
	if !isAddressRole(p.Role) {
		return fmt.Errorf("unsupported address role %q", p.Role)
	}
	fromUser, err := PairFromUserAddress(p.Role, p.User)
	if err != nil {
		return err
	}
	fromRaw, err := PairFromRawAddress(p.Role, p.Raw)
	if err != nil {
		return err
	}
	if fromUser.Raw != fromRaw.Raw || fromUser.User != fromRaw.User {
		return fmt.Errorf("%s AE and raw addresses must represent the same account", p.Role)
	}
	return nil
}

func deriveAddressPair(role AddressRole, pubKey cryptotypes.PubKey) (AddressPair, error) {
	if pubKey == nil {
		return AddressPair{}, errors.New("public key is required")
	}
	return addressPairFromBytes(role, []byte(pubKey.Address()))
}

func addressPairFromBytes(role AddressRole, bz []byte) (AddressPair, error) {
	if !isAddressRole(role) {
		return AddressPair{}, fmt.Errorf("unsupported address role %q", role)
	}
	user, err := FormatUserFriendly(bz)
	if err != nil {
		return AddressPair{}, err
	}
	return AddressPair{
		Role:	role,
		User:	user,
		Raw:	Format(bz),
	}, nil
}

func isAddressRole(role AddressRole) bool {
	return role == AddressRoleAccount || role == AddressRoleValidator || role == AddressRoleConsensus
}
