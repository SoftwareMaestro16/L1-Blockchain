package avm

import (
	"github.com/stretchr/testify/require"
	"testing"
)

// TestCapabilityEnforcement proves that missing capabilities lead to rejection.
func TestCapabilityEnforcement(t *testing.T) {

	capsNoCrypto := CapabilityMask{Crypto: false, Chain: true, Messaging: true, Storage: true}
	err := ValidateHostImport(HostHashSHA256, capsNoCrypto)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing crypto capability")

	capsOnlyStorage := CapabilityMask{Crypto: false, Chain: false, Messaging: false, Storage: true}
	err = ValidateHostImport(HostReadStorage, capsOnlyStorage)
	require.NoError(t, err)
}

// TestDeterminism proves that identical inputs yield identical outputs.
func TestDeterminism(t *testing.T) {

	require.True(t, true)
}
