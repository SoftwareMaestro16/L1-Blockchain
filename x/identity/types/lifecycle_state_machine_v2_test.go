package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDomainLifecycleStateMachineV2RegistrationRenewalGraceAndRelease(t *testing.T) {
	record := lifecycleRecordV2(t, DomainRecordV2Available)

	next, err := ApplyDomainLifecycleTransitionV2(record, DomainLifecycleTransitionContextV2{
		Event:		DomainLifecycleEventCommitRegistration,
		Height:		10,
		DepositPaid:	true,
	})
	require.NoError(t, err)
	require.Equal(t, DomainRecordV2Committed, next.Status)

	next, err = ApplyDomainLifecycleTransitionV2(next, DomainLifecycleTransitionContextV2{
		Event:				DomainLifecycleEventRevealRegistration,
		Height:				11,
		RegistrationPaymentPaid:	true,
		NFTMintedOrBound:		true,
	})
	require.NoError(t, err)
	require.Equal(t, DomainRecordV2Active, next.Status)

	next.RenewalStartHeight = 90
	next.ExpiryHeight = 100
	next, err = ApplyDomainLifecycleTransitionV2(next, DomainLifecycleTransitionContextV2{
		Event:	DomainLifecycleEventRenewalWindowBegin,
		Height:	90,
	})
	require.NoError(t, err)
	require.Equal(t, DomainRecordV2RenewalWindow, next.Status)

	next, err = ApplyDomainLifecycleTransitionV2(next, DomainLifecycleTransitionContextV2{
		Event:	DomainLifecycleEventExpireDomain,
		Height:	100,
	})
	require.NoError(t, err)
	require.Equal(t, DomainRecordV2Expired, next.Status)

	next, err = ApplyDomainLifecycleTransitionV2(next, DomainLifecycleTransitionContextV2{
		Event:			DomainLifecycleEventGracePeriodBegin,
		Height:			101,
		GracePeriodEnabled:	true,
	})
	require.NoError(t, err)
	require.Equal(t, DomainRecordV2GraceLocked, next.Status)

	next, err = ApplyDomainLifecycleTransitionV2(next, DomainLifecycleTransitionContextV2{
		Event:			DomainLifecycleEventGracePeriodEnd,
		Height:			120,
		GracePeriodEndHeight:	120,
	})
	require.NoError(t, err)
	require.Equal(t, DomainRecordV2Released, next.Status)

	next, err = ApplyDomainLifecycleTransitionV2(next, DomainLifecycleTransitionContextV2{
		Event:	DomainLifecycleEventReleaseDomain,
		Height:	121,
	})
	require.NoError(t, err)
	require.Equal(t, DomainRecordV2Available, next.Status)
	require.Empty(t, next.Owner)
	require.Empty(t, next.NFTItemID)
	require.Equal(t, uint64(121), next.LifecycleEpoch)
}

func TestDomainLifecycleStateMachineV2DirectRegistrationAndRenewalRecovery(t *testing.T) {
	record := lifecycleRecordV2(t, DomainRecordV2Available)
	_, err := ApplyDomainLifecycleTransitionV2(record, DomainLifecycleTransitionContextV2{
		Event:				DomainLifecycleEventRegisterDirect,
		Height:				10,
		RegistrationPaymentPaid:	true,
		NFTMintedOrBound:		true,
	})
	require.ErrorContains(t, err, "direct registration class is disabled")

	next, err := ApplyDomainLifecycleTransitionV2(record, DomainLifecycleTransitionContextV2{
		Event:				DomainLifecycleEventRegisterDirect,
		Height:				10,
		DirectRegistrationEnabled:	true,
		RegistrationPaymentPaid:	true,
		NFTMintedOrBound:		true,
	})
	require.NoError(t, err)
	require.Equal(t, DomainRecordV2Active, next.Status)

	next.Status = DomainRecordV2GraceLocked
	next, err = ApplyDomainLifecycleTransitionV2(next, DomainLifecycleTransitionContextV2{
		Event:			DomainLifecycleEventRenewDomain,
		Height:			111,
		RenewalPaymentPaid:	true,
	})
	require.NoError(t, err)
	require.Equal(t, DomainRecordV2Active, next.Status)
}

func TestDomainLifecycleStateMachineV2AuctionAlternative(t *testing.T) {
	record := lifecycleRecordV2(t, DomainRecordV2Available)
	next, err := ApplyDomainLifecycleTransitionV2(record, DomainLifecycleTransitionContextV2{
		Event:	DomainLifecycleEventStartAuction,
		Height:	10,
	})
	require.NoError(t, err)
	require.Equal(t, DomainRecordV2Auction, next.Status)

	_, err = ApplyDomainLifecycleTransitionV2(next, DomainLifecycleTransitionContextV2{
		Event:			DomainLifecycleEventFinalizeAuction,
		Height:			20,
		AuctionFinalized:	true,
		NFTMintedOrBound:	true,
	})
	require.ErrorContains(t, err, "deterministic finalization")

	next, err = ApplyDomainLifecycleTransitionV2(next, DomainLifecycleTransitionContextV2{
		Event:				DomainLifecycleEventFinalizeAuction,
		Height:				20,
		AuctionFinalized:		true,
		DeterministicWinnerProof:	true,
		NFTMintedOrBound:		true,
	})
	require.NoError(t, err)
	require.Equal(t, DomainRecordV2Active, next.Status)
}

func TestDomainLifecycleStateMachineV2RejectsMissingPaymentsAndEarlyAutomaticTransitions(t *testing.T) {
	record := lifecycleRecordV2(t, DomainRecordV2Available)
	_, err := ApplyDomainLifecycleTransitionV2(record, DomainLifecycleTransitionContextV2{
		Event:	DomainLifecycleEventCommitRegistration,
		Height:	10,
	})
	require.ErrorContains(t, err, "requires deposit")

	record.Status = DomainRecordV2Committed
	_, err = ApplyDomainLifecycleTransitionV2(record, DomainLifecycleTransitionContextV2{
		Event:			DomainLifecycleEventRevealRegistration,
		Height:			11,
		NFTMintedOrBound:	true,
	})
	require.ErrorContains(t, err, "requires payment")

	record.Status = DomainRecordV2Active
	record.RenewalStartHeight = 90
	record.ExpiryHeight = 100
	_, err = ApplyDomainLifecycleTransitionV2(record, DomainLifecycleTransitionContextV2{
		Event:	DomainLifecycleEventRenewalWindowBegin,
		Height:	89,
	})
	require.ErrorContains(t, err, "has not begun")

	record.Status = DomainRecordV2RenewalWindow
	_, err = ApplyDomainLifecycleTransitionV2(record, DomainLifecycleTransitionContextV2{
		Event:	DomainLifecycleEventExpireDomain,
		Height:	99,
	})
	require.ErrorContains(t, err, "expiry has not been reached")
}

func TestAutomaticDomainLifecycleStatusV2(t *testing.T) {
	record := lifecycleRecordV2(t, DomainRecordV2Active)
	record.RenewalStartHeight = 90
	record.ExpiryHeight = 100
	require.Equal(t, DomainRecordV2RenewalWindow, AutomaticDomainLifecycleStatusV2(record, 90, 0))
	record.Status = DomainRecordV2RenewalWindow
	require.Equal(t, DomainRecordV2Expired, AutomaticDomainLifecycleStatusV2(record, 100, 0))
	record.Status = DomainRecordV2Expired
	require.Equal(t, DomainRecordV2GraceLocked, AutomaticDomainLifecycleStatusV2(record, 101, 0))
	record.Status = DomainRecordV2GraceLocked
	require.Equal(t, DomainRecordV2Released, AutomaticDomainLifecycleStatusV2(record, 120, 120))
}

func lifecycleRecordV2(t *testing.T, status DomainRecordV2Status) DomainRecordV2 {
	t.Helper()
	record, err := NewDomainRecordV2FromDomain(Domain{
		Name:			"alice.aet",
		Owner:			addr(1),
		NFTID:			"anft66:domain:alice.aet",
		RegisteredHeight:	1,
		ExpiryHeight:		100,
		UpdatedHeight:		1,
	}, status, 0, 1)
	require.NoError(t, err)
	return record
}
