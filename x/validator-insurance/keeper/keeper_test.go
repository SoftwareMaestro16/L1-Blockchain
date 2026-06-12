package keeper

import (
	"context"
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/x/internal/kvtest"
	"github.com/sovereign-l1/l1/x/internal/prototype"
	"github.com/sovereign-l1/l1/x/validator-insurance/types"
	validatorregistrytypes "github.com/sovereign-l1/l1/x/validator-registry/types"
)

func TestDefaultGenesisValidates(t *testing.T) {
	require.NoError(t, DefaultGenesis().Validate())
}

func TestFundInsurance(t *testing.T) {
	k := NewKeeper()
	insurance := fundInsurance(t, &k, rawInsuranceAddress("11"), 1_500)

	require.Equal(t, rawInsuranceAddress("11"), insurance.ValidatorAddress)
	require.Equal(t, uint64(1_500), insurance.Balance)
	require.Equal(t, validatorregistrytypes.StatusCandidate, insurance.ValidatorStatus)
}

func TestRejectValidatorActivationWithoutRequiredInsurance(t *testing.T) {
	k := NewKeeper()
	validator := rawInsuranceAddress("11")

	require.ErrorContains(t, k.ValidateValidatorActivation(validator), "minimum insurance")
	fundInsurance(t, &k, validator, 999)
	require.ErrorContains(t, k.ValidateValidatorActivation(validator), "minimum insurance")
	fundInsurance(t, &k, validator, 1)
	require.NoError(t, k.ValidateValidatorActivation(validator))

	active, err := k.MarkValidatorStatus(validator, validatorregistrytypes.StatusActive)
	require.NoError(t, err)
	require.Equal(t, validatorregistrytypes.StatusActive, active.ValidatorStatus)
}

func TestSlashDrainsInsuranceFirstAccordingToParams(t *testing.T) {
	k := NewKeeper()
	validator := rawInsuranceAddress("11")
	fundInsurance(t, &k, validator, 1_000)

	result, err := k.ApplyValidatorSlash(validator, 300, "double-sign")
	require.NoError(t, err)
	require.Equal(t, uint64(300), result.CoveredAmount)
	require.Equal(t, uint64(0), result.RemainingPenalty)
	stored, found := k.ValidatorInsurance(validator)
	require.True(t, found)
	require.Equal(t, uint64(700), stored.Balance)

	result, err = k.ApplyValidatorSlash(validator, 1_000, "double-sign")
	require.NoError(t, err)
	require.Equal(t, uint64(700), result.CoveredAmount)
	require.Equal(t, uint64(300), result.RemainingPenalty)
}

func TestClaimPayoutCappedAndCannotBePaidTwice(t *testing.T) {
	k := NewKeeper()
	validator := rawInsuranceAddress("11")
	fundInsurance(t, &k, validator, 500)

	claim, err := k.SubmitInsuranceClaim(types.MsgSubmitInsuranceClaim{
		Authority:		prototype.DefaultAuthority,
		ClaimID:		"claim-1",
		ValidatorAddress:	validator,
		Claimant:		rawInsuranceAddress("22"),
		Amount:			800,
		Reason:			"delegator loss",
		Height:			2,
	})
	require.NoError(t, err)
	require.Equal(t, types.ClaimStatusPending, claim.Status)

	paid, err := k.ResolveInsuranceClaim(types.MsgResolveInsuranceClaim{
		Authority:	prototype.DefaultAuthority,
		ClaimID:	"claim-1",
		Approved:	true,
		Height:		3,
	})
	require.NoError(t, err)
	require.Equal(t, types.ClaimStatusPaid, paid.Status)
	require.True(t, paid.Paid)
	require.Equal(t, uint64(500), paid.PayoutAmount)

	_, err = k.ResolveInsuranceClaim(types.MsgResolveInsuranceClaim{
		Authority:	prototype.DefaultAuthority,
		ClaimID:	"claim-1",
		Approved:	true,
		Height:		4,
	})
	require.ErrorContains(t, err, "already resolved")
}

func TestWithdrawalDelayEnforced(t *testing.T) {
	k := NewKeeper()
	validator := rawInsuranceAddress("11")
	fundInsurance(t, &k, validator, 2_000)

	_, err := k.WithdrawValidatorInsurance(types.MsgWithdrawValidatorInsurance{
		Authority:		prototype.DefaultAuthority,
		ValidatorAddress:	validator,
		Recipient:		rawInsuranceAddress("22"),
		Amount:			1_001,
		Height:			2,
		ValidatorStatus:	validatorregistrytypes.StatusActive,
	})
	require.ErrorContains(t, err, "minimum requirement")

	withdrawal, err := k.WithdrawValidatorInsurance(types.MsgWithdrawValidatorInsurance{
		Authority:		prototype.DefaultAuthority,
		ValidatorAddress:	validator,
		Recipient:		rawInsuranceAddress("22"),
		Amount:			500,
		Height:			3,
		ValidatorStatus:	validatorregistrytypes.StatusActive,
	})
	require.NoError(t, err)
	require.Equal(t, types.WithdrawalStatusPending, withdrawal.Status)
	require.Equal(t, uint64(3+k.InsuranceParams().WithdrawalLockBlocks), withdrawal.CompleteHeight)

	_, err = k.WithdrawValidatorInsurance(types.MsgWithdrawValidatorInsurance{
		Authority:		prototype.DefaultAuthority,
		ValidatorAddress:	validator,
		Recipient:		rawInsuranceAddress("22"),
		Height:			withdrawal.CompleteHeight - 1,
		ValidatorStatus:	validatorregistrytypes.StatusActive,
	})
	require.ErrorContains(t, err, "lock period")

	completed, err := k.WithdrawValidatorInsurance(types.MsgWithdrawValidatorInsurance{
		Authority:		prototype.DefaultAuthority,
		ValidatorAddress:	validator,
		Recipient:		rawInsuranceAddress("22"),
		Height:			withdrawal.CompleteHeight,
		ValidatorStatus:	validatorregistrytypes.StatusActive,
	})
	require.NoError(t, err)
	require.Equal(t, types.WithdrawalStatusCompleted, completed.Status)
}

func TestExportImportPreservesPendingClaims(t *testing.T) {
	source := NewKeeper()
	validator := rawInsuranceAddress("11")
	fundInsurance(t, &source, validator, 1_500)
	claim, err := source.SubmitInsuranceClaim(types.MsgSubmitInsuranceClaim{
		Authority:		prototype.DefaultAuthority,
		ClaimID:		"claim-1",
		ValidatorAddress:	validator,
		Claimant:		rawInsuranceAddress("22"),
		Amount:			700,
		Reason:			"pending review",
		Height:			2,
	})
	require.NoError(t, err)

	exported := source.ExportGenesis()
	require.NoError(t, exported.Validate())
	target := NewKeeper()
	require.NoError(t, target.InitGenesis(exported))
	require.Equal(t, exported, target.ExportGenesis())
	require.NoError(t, NewMigrator(&target).Migrate1to2())
	claims := target.InsuranceClaims(validator)
	require.Len(t, claims, 1)
	require.Equal(t, claim, claims[0])
}

func TestPersistentRuntimeMutationSurvivesRestartAndImport(t *testing.T) {
	ctx := context.Background()
	service := kvtest.NewStoreService()
	source := NewPersistentKeeper(service)
	require.NoError(t, source.InitGenesisState(ctx, DefaultGenesis()))

	validator := rawInsuranceAddress("33")
	fundInsurance(t, &source, validator, 1_500)
	claim, err := source.SubmitInsuranceClaim(types.MsgSubmitInsuranceClaim{
		Authority:		prototype.DefaultAuthority,
		ClaimID:		"persistent-claim",
		ValidatorAddress:	validator,
		Claimant:		rawInsuranceAddress("44"),
		Amount:			700,
		Reason:			"restart proof",
		Height:			2,
	})
	require.NoError(t, err)

	restarted := NewPersistentKeeper(service)
	exported, err := restarted.ExportGenesisState(ctx)
	require.NoError(t, err)
	require.Len(t, exported.State.Claims, 1)

	imported := NewPersistentKeeper(kvtest.NewStoreService())
	require.NoError(t, imported.InitGenesisState(ctx, exported))
	claims := imported.InsuranceClaims(validator)
	require.Len(t, claims, 1)
	require.Equal(t, claim, claims[0])
}

func TestFundingOverflowRejected(t *testing.T) {
	k := NewKeeper()
	validator := rawInsuranceAddress("11")
	fundInsurance(t, &k, validator, math.MaxUint64)

	_, err := k.FundValidatorInsurance(types.MsgFundValidatorInsurance{
		Authority:		prototype.DefaultAuthority,
		ValidatorAddress:	validator,
		Funder:			rawInsuranceAddress("22"),
		Amount:			1,
		Height:			2,
	})
	require.ErrorContains(t, err, "overflow")
}

func TestDuplicateClaimRejected(t *testing.T) {
	k := NewKeeper()
	validator := rawInsuranceAddress("11")
	fundInsurance(t, &k, validator, 1_000)
	msg := types.MsgSubmitInsuranceClaim{
		Authority:		prototype.DefaultAuthority,
		ClaimID:		"claim-1",
		ValidatorAddress:	validator,
		Claimant:		rawInsuranceAddress("22"),
		Amount:			100,
		Reason:			"duplicate test",
		Height:			2,
	}
	_, err := k.SubmitInsuranceClaim(msg)
	require.NoError(t, err)
	_, err = k.SubmitInsuranceClaim(msg)
	require.ErrorContains(t, err, "already exists")
}

func fundInsurance(t *testing.T, k *Keeper, validator string, amount uint64) types.ValidatorInsurance {
	t.Helper()
	insurance, err := k.FundValidatorInsurance(types.MsgFundValidatorInsurance{
		Authority:		prototype.DefaultAuthority,
		ValidatorAddress:	validator,
		Funder:			rawInsuranceAddress("22"),
		Amount:			amount,
		Height:			1,
	})
	require.NoError(t, err)
	return insurance
}

func rawInsuranceAddress(hexByte string) string {
	return "4:000000000000000000000000" + fmt.Sprintf("%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s", hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte)
}
