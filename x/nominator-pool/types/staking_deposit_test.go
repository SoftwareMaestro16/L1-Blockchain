package types

import (
	"testing"
)

func TestValidateStakingPoolDeposit_MinDeposit(t *testing.T) {
	params := Params{
		MinPoolDeposit: 10_000_000_000,
	}
	msg := MsgDepositToStakingPool{
		Amount:	9_000_000_000,
		Height:	1,
	}
	err := ValidateStakingPoolDeposit(msg, params)
	if err == nil {
		t.Error("expected error for deposit < 10 AET, got nil")
	}
}
