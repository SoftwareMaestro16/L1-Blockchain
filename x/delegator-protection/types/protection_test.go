package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestProtocolFeeShareEntersFund(t *testing.T) {
	state := newProtectionState(t)
	next, allocation, err := ApplyProtocolFeeShare(state, sdkmath.NewInt(10_000))
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(250), allocation.AddedToFund)
	require.Equal(t, sdkmath.NewInt(9_750), allocation.Remainder)
	require.Equal(t, sdkmath.NewInt(250), QueryProtectionFund(next).Balance)
	require.NoError(t, CheckDelegatorProtectionInvariants(next))
}

func TestValidClaimApprovedAndPaid(t *testing.T) {
	state := fundedProtectionState(t, 20_000)
	next, claim, err := ApplySubmitDelegatorProtectionClaim(state, validClaimMsg("del-a", "val-a", 1_000, 800, 7, 100))
	require.NoError(t, err)
	require.Equal(t, ClaimStatusSubmitted, claim.Status)

	next, err = ApplyApproveDelegatorProtectionClaim(next, MsgApproveDelegatorProtectionClaim{
		Authority:	next.Params.Authority,
		ClaimID:	claim.ClaimID,
		ApprovedPayout:	sdkmath.NewInt(800),
		Height:		101,
	})
	require.NoError(t, err)
	approved := QueryProtectionClaims(next, QueryProtectionClaimsRequest{Delegator: "del-a", Status: ClaimStatusApproved})
	require.Len(t, approved, 1)

	next, payout, err := ApplyClaimDelegatorCompensation(next, MsgClaimDelegatorCompensation{
		Delegator:	"del-a",
		ClaimID:	claim.ClaimID,
		Epoch:		7,
		Height:		102,
	})
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(800), payout.Amount)
	require.Equal(t, fundedProtectionState(t, 20_000).Fund.Balance.Sub(sdkmath.NewInt(800)), next.Fund.Balance)
	paid := QueryProtectionClaims(next, QueryProtectionClaimsRequest{Delegator: "del-a", Status: ClaimStatusPaid})
	require.Len(t, paid, 1)
	require.True(t, paid[0].Paid)
	payouts := QueryDelegatorCompensation(next, QueryDelegatorCompensationRequest{Delegator: "del-a"})
	require.Len(t, payouts, 1)
	require.NoError(t, CheckDelegatorProtectionInvariants(next))
}

func TestDuplicateClaimRejected(t *testing.T) {
	state := fundedProtectionState(t, 20_000)
	msg := validClaimMsg("del-a", "val-a", 1_000, 800, 7, 100)
	next, _, err := ApplySubmitDelegatorProtectionClaim(state, msg)
	require.NoError(t, err)
	_, _, err = ApplySubmitDelegatorProtectionClaim(next, msg)
	require.ErrorContains(t, err, "duplicate")
}

func TestMaxPayoutPerEpochEnforced(t *testing.T) {
	state := fundedProtectionState(t, 20_000)
	next, claimA, err := ApplySubmitDelegatorProtectionClaim(state, validClaimMsg("del-a", "val-a", 9_000, 6_000, 8, 100))
	require.NoError(t, err)
	next, err = ApplyApproveDelegatorProtectionClaim(next, MsgApproveDelegatorProtectionClaim{Authority: next.Params.Authority, ClaimID: claimA.ClaimID, ApprovedPayout: sdkmath.NewInt(6_000), Height: 101})
	require.NoError(t, err)
	next, _, err = ApplyClaimDelegatorCompensation(next, MsgClaimDelegatorCompensation{Delegator: "del-a", ClaimID: claimA.ClaimID, Epoch: 8, Height: 102})
	require.NoError(t, err)

	next, claimB, err := ApplySubmitDelegatorProtectionClaim(next, validClaimMsg("del-b", "val-b", 9_000, 6_000, 8, 103))
	require.NoError(t, err)
	next, err = ApplyApproveDelegatorProtectionClaim(next, MsgApproveDelegatorProtectionClaim{Authority: next.Params.Authority, ClaimID: claimB.ClaimID, ApprovedPayout: sdkmath.NewInt(6_000), Height: 104})
	require.NoError(t, err)
	_, _, err = ApplyClaimDelegatorCompensation(next, MsgClaimDelegatorCompensation{Delegator: "del-b", ClaimID: claimB.ClaimID, Epoch: 8, Height: 105})
	require.ErrorContains(t, err, "max per epoch")
}

func TestInsufficientFundHandledCleanly(t *testing.T) {
	state := fundedProtectionState(t, 1_500)
	next, claim, err := ApplySubmitDelegatorProtectionClaim(state, validClaimMsg("del-a", "val-a", 1_000, 800, 9, 100))
	require.NoError(t, err)
	next, err = ApplyApproveDelegatorProtectionClaim(next, MsgApproveDelegatorProtectionClaim{
		Authority:	next.Params.Authority,
		ClaimID:	claim.ClaimID,
		ApprovedPayout:	sdkmath.NewInt(800),
		Height:		101,
	})
	require.NoError(t, err)
	_, _, err = ApplyClaimDelegatorCompensation(next, MsgClaimDelegatorCompensation{
		Delegator:	"del-a",
		ClaimID:	claim.ClaimID,
		Epoch:		9,
		Height:		102,
	})
	require.ErrorContains(t, err, "reserve floor")
	require.Equal(t, sdkmath.NewInt(1_500), next.Fund.Balance)
}

func TestExportImportPreservesClaimQueue(t *testing.T) {
	state := fundedProtectionState(t, 20_000)
	next, claimA, err := ApplySubmitDelegatorProtectionClaim(state, validClaimMsg("del-b", "val-b", 1_000, 500, 10, 100))
	require.NoError(t, err)
	next, _, err = ApplySubmitDelegatorProtectionClaim(next, validClaimMsg("del-a", "val-a", 2_000, 700, 10, 101))
	require.NoError(t, err)
	next, err = ApplyApproveDelegatorProtectionClaim(next, MsgApproveDelegatorProtectionClaim{
		Authority:	next.Params.Authority,
		ClaimID:	claimA.ClaimID,
		ApprovedPayout:	sdkmath.NewInt(500),
		Height:		102,
	})
	require.NoError(t, err)

	exported, err := ExportDelegatorProtectionState(next)
	require.NoError(t, err)
	imported, err := ImportDelegatorProtectionState(exported)
	require.NoError(t, err)
	require.Equal(t, exported, imported)
	require.Len(t, QueryProtectionClaims(imported, QueryProtectionClaimsRequest{}), 2)
	require.NoError(t, CheckDelegatorProtectionInvariants(imported))
}

func TestProtectionSecurityInvariantsRejectTamperingAndBadParams(t *testing.T) {
	state := newProtectionState(t)
	params := state.Params
	params.IncomingFeeShareBps = 9_500
	params.TreasuryDistributionBps = 1_000
	_, err := ApplyUpdateProtectionParams(state, MsgUpdateProtectionParams{Authority: state.Params.Authority, Params: params})
	require.ErrorContains(t, err, "incompatible")

	_, err = ApplyUpdateProtectionParams(state, MsgUpdateProtectionParams{Authority: "wrong", Params: state.Params})
	require.ErrorContains(t, err, "requires authority")

	tampered := fundedProtectionState(t, 20_000)
	tampered.Fund.Balance = sdkmath.NewInt(-1)
	tampered.Fund.FundHash = ComputeProtectionFundHash(tampered.Fund)
	require.ErrorContains(t, CheckDelegatorProtectionInvariants(tampered), "cannot go negative")

	paidTwice := fundedProtectionState(t, 30_000)
	paidTwice, claim, err := ApplySubmitDelegatorProtectionClaim(paidTwice, validClaimMsg("del-a", "val-a", 1_000, 500, 11, 100))
	require.NoError(t, err)
	paidTwice, err = ApplyApproveDelegatorProtectionClaim(paidTwice, MsgApproveDelegatorProtectionClaim{Authority: paidTwice.Params.Authority, ClaimID: claim.ClaimID, ApprovedPayout: sdkmath.NewInt(500), Height: 101})
	require.NoError(t, err)
	paidTwice, payout, err := ApplyClaimDelegatorCompensation(paidTwice, MsgClaimDelegatorCompensation{Delegator: "del-a", ClaimID: claim.ClaimID, Epoch: 11, Height: 102})
	require.NoError(t, err)
	paidTwice.Payouts = append(paidTwice.Payouts, payout)
	require.ErrorContains(t, CheckDelegatorProtectionInvariants(paidTwice), "paid twice")
}

func newProtectionState(t *testing.T) DelegatorProtectionState {
	t.Helper()
	state, err := NewDelegatorProtectionState(DefaultProtectionParams())
	require.NoError(t, err)
	return state
}

func fundedProtectionState(t *testing.T, balance int64) DelegatorProtectionState {
	t.Helper()
	state := newProtectionState(t)
	state.Fund.Balance = sdkmath.NewInt(balance)
	state.Fund.FundHash = ComputeProtectionFundHash(state.Fund)
	require.NoError(t, state.Validate())
	return state
}

func validClaimMsg(delegator string, validator string, loss int64, requested int64, epoch uint64, height uint64) MsgSubmitDelegatorProtectionClaim {
	return MsgSubmitDelegatorProtectionClaim{
		Delegator:		delegator,
		Validator:		validator,
		LossAmount:		sdkmath.NewInt(loss),
		RequestedPayout:	sdkmath.NewInt(requested),
		EligibilityHash:	protectionHashParts("eligibility", delegator, validator),
		Reason:			"delegator loss",
		Epoch:			epoch,
		Height:			height,
	}
}
