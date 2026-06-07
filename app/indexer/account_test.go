package indexer

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"

	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
)

func TestNewAccountMetadataFormatsAetraAddress(t *testing.T) {
	priv := secp256k1.GenPrivKey()
	account := authtypes.NewBaseAccount(priv.PubKey().Address().Bytes(), priv.PubKey(), 12, 34)

	metadata, err := NewAccountMetadata(account)
	require.NoError(t, err)
	require.Equal(t, aetraaddress.FormatAccAddress(account.GetAddress()), metadata.Address)
	require.Equal(t, uint64(12), metadata.AccountNumber)
	require.Equal(t, uint64(34), metadata.Sequence)
	require.Contains(t, metadata.PubKeyType, "secp256k1")
}

func TestNewAccountMetadataRejectsInvalidAccounts(t *testing.T) {
	_, err := NewAccountMetadata(nil)
	require.ErrorContains(t, err, "account must not be nil")

	_, err = NewAccountMetadata(authtypes.NewBaseAccount(make([]byte, 20), nil, 0, 0))
	require.ErrorContains(t, err, "account metadata address must not be zero address")
}
