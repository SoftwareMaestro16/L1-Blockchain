package indexer

import (
	"fmt"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
)

type AccountMetadata struct {
	Address		string
	AccountNumber	uint64
	Sequence	uint64
	PubKeyType	string
}

type Account interface {
	GetAddress() sdk.AccAddress
	GetAccountNumber() uint64
	GetSequence() uint64
	GetPubKey() cryptotypes.PubKey
}

func NewAccountMetadata(account Account) (AccountMetadata, error) {
	if account == nil {
		return AccountMetadata{}, fmt.Errorf("account must not be nil")
	}
	address := account.GetAddress()
	if err := aetraaddress.RejectZeroAddress("account metadata address", address); err != nil {
		return AccountMetadata{}, err
	}
	pubKeyType := ""
	if pubKey := account.GetPubKey(); pubKey != nil {
		pubKeyType = fmt.Sprintf("%T", pubKey)
	}
	return AccountMetadata{
		Address:	aetraaddress.FormatAccAddress(address),
		AccountNumber:	account.GetAccountNumber(),
		Sequence:	account.GetSequence(),
		PubKeyType:	pubKeyType,
	}, nil
}
