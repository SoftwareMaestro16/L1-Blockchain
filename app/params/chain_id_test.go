package params

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateAetraChainIDAcceptsLaunchProfiles(t *testing.T) {
	for _, chainID := range []string{
		"aetra-local-1",
		"aetra-testnet-1",
		"aetra-public-testnet-2026-1",
		"aetra-mainnet-1",
	} {
		require.NoError(t, ValidateAetraChainID(chainID), chainID)
	}

	for _, chainID := range []string{
		"aetra-local-1",
		"aetra-testnet-1",
		"aetra-preflight-1",
		"aetra-public-testnet-2026-1",
	} {
		require.NoError(t, ValidateAetraTestnetChainID(chainID), chainID)
	}
}

func TestValidateAetraChainIDRejectsMalformedIDs(t *testing.T) {
	for _, tc := range []struct {
		chainID	string
		want	string
	}{
		{"", "required"},
		{"cosmoshub-4", "aetra-"},
		{"Aetra-testnet-1", "lower-case"},
		{"aetra-testnet-1!", "lower-case"},
		{"aetra--testnet-1", "non-empty"},
		{"aetra-testnet-", "non-empty"},
		{"aetra-" + strings.Repeat("a", ChainIDMaxLength), "must not exceed"},
	} {
		t.Run(tc.chainID, func(t *testing.T) {
			require.ErrorContains(t, ValidateAetraChainID(tc.chainID), tc.want)
		})
	}
}

func TestValidateAetraTestnetChainIDRejectsMainnetID(t *testing.T) {
	require.ErrorContains(t, ValidateAetraTestnetChainID("aetra-mainnet-1"), "testnet chain-id")
}
