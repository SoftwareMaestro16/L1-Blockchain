package types

import (
	"fmt"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
)

const (
	AccountByUserPrefix		= "account/by_user/"
	AccountByRawPrefix		= "account/by_raw/"
	AccountByNumberPrefix		= "account/number/"
	AccountByReputationPrefix	= "account/reputation/"
	AccountStoragePrefix		= "account/storage/"
)

type AccountStoreKeyDescriptor struct {
	Name	string
	Prefix	string
}

func DefaultAccountStoreKeyDescriptors() []AccountStoreKeyDescriptor {
	return []AccountStoreKeyDescriptor{
		{Name: "account/by_user", Prefix: AccountByUserPrefix},
		{Name: "account/by_raw", Prefix: AccountByRawPrefix},
		{Name: "account/number", Prefix: AccountByNumberPrefix},
		{Name: "account/reputation", Prefix: AccountByReputationPrefix},
		{Name: "account/storage", Prefix: AccountStoragePrefix},
	}
}

func AccountByUserKey(userAddress string) (string, error) {
	pair, err := addressing.PairFromUserAddress(addressing.AddressRoleAccount, userAddress)
	if err != nil {
		return "", err
	}
	return AccountByUserPrefix + pair.User, nil
}

func AccountByRawKey(rawAddress string) (string, error) {
	pair, err := addressing.PairFromRawAddress(addressing.AddressRoleAccount, rawAddress)
	if err != nil {
		return "", err
	}
	return AccountByRawPrefix + pair.Raw, nil
}

func AccountByNumberKey(accountNumber uint64) string {
	return AccountByNumberPrefix + fmt.Sprintf("%020d", accountNumber)
}

func AccountByReputationKey(reputationID string) (string, error) {
	reputationID = strings.TrimSpace(reputationID)
	if err := validateReputationID(reputationID); err != nil {
		return "", err
	}
	if reputationID == "" {
		return "", fmt.Errorf("native account reputation id is required")
	}
	return AccountByReputationPrefix + reputationID, nil
}

func AccountStorageKey(userAddress string) (string, error) {
	pair, err := addressing.PairFromUserAddress(addressing.AddressRoleAccount, userAddress)
	if err != nil {
		return "", err
	}
	return AccountStoragePrefix + pair.User, nil
}
