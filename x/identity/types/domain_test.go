package types

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateDomainNameAcceptsProductionAlphabet(t *testing.T) {
	for _, name := range []string{"alice", "alice-01", "alice_01", "alice.aet"} {
		require.NoError(t, ValidateDomainName(name), name)
	}
}

func TestValidateDomainNameRejectsUnsafeInput(t *testing.T) {
	cases := []string{
		"",
		" ",
		"Alice",
		"alice.eth",
		"alice beta",
		"alice.beta",
		"alice/aet",
		"alice\u200bbeta",
		"аlice",
		strings.Repeat("a", MaxDomainLabelBytes+1),
	}
	for _, name := range cases {
		require.Error(t, ValidateDomainName(name), name)
	}
}

func TestNormalizeDomainName(t *testing.T) {
	normalized, err := NormalizeDomainName("Alice.AET")
	require.NoError(t, err)
	require.Equal(t, "alice", normalized)
}

func TestValidateDomainRecord(t *testing.T) {
	owner := []byte{1, 2, 3}
	resolver := []byte{4, 5, 6}

	record := DomainRecord{
		Name:		"alice",
		TLD:		DomainTLD,
		Owner:		owner,
		Resolver:	resolver,
		ExpiryUnix:	100,
		NFTItemID:	"anft66:1",
		Status:		DomainStatusActive,
		CreatedAtUnix:	1,
		UpdatedAtUnix:	2,
	}
	require.NoError(t, ValidateDomainRecord(record))

	record.Owner = make([]byte, 20)
	require.ErrorContains(t, ValidateDomainRecord(record), "domain owner")
	record.Owner = owner

	record.Resolver = make([]byte, 20)
	require.ErrorContains(t, ValidateDomainRecord(record), "domain resolver")
	record.Resolver = resolver

	record.Status = "locked"
	require.ErrorContains(t, ValidateDomainRecord(record), "invalid domain status")
	record.Status = DomainStatusActive

	record.NFTItemID = ""
	require.ErrorContains(t, ValidateDomainRecord(record), "nft item id")
}
