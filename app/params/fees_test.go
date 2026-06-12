package params

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateNativeFeeDenomsV1(t *testing.T) {
	require.NoError(t, ValidateNativeFeeDenomsV1([]string{BaseDenom}, 1))

	tests := map[string][]string{
		"empty":		{},
		"non native":		{"uatom"},
		"duplicate base":	{BaseDenom, BaseDenom},
	}
	for name, denoms := range tests {
		t.Run(name, func(t *testing.T) {
			require.Error(t, ValidateNativeFeeDenomsV1(denoms, 1))
		})
	}
}
