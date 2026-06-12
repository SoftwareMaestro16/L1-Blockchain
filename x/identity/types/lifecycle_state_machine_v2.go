package types

import (
	"errors"
	"fmt"
)

type DomainLifecycleEventV2 string

const (
	DomainLifecycleEventCommitRegistration	DomainLifecycleEventV2	= "MsgCommitRegistration"
	DomainLifecycleEventRevealRegistration	DomainLifecycleEventV2	= "MsgRevealRegistration"
	DomainLifecycleEventRegisterDirect	DomainLifecycleEventV2	= "MsgRegisterDirect"
	DomainLifecycleEventRenewDomain		DomainLifecycleEventV2	= "MsgRenewDomain"
	DomainLifecycleEventStartAuction	DomainLifecycleEventV2	= "MsgStartAuction"
	DomainLifecycleEventFinalizeAuction	DomainLifecycleEventV2	= "MsgFinalizeAuction"
	DomainLifecycleEventRenewalWindowBegin	DomainLifecycleEventV2	= "renewal_window_begins"
	DomainLifecycleEventExpireDomain	DomainLifecycleEventV2	= "expiry_reached"
	DomainLifecycleEventGracePeriodBegin	DomainLifecycleEventV2	= "grace_period_begins"
	DomainLifecycleEventGracePeriodEnd	DomainLifecycleEventV2	= "grace_period_ends"
	DomainLifecycleEventReleaseDomain	DomainLifecycleEventV2	= "release_domain"
)

type DomainLifecycleTransitionContextV2 struct {
	Event				DomainLifecycleEventV2
	Height				uint64
	DepositPaid			bool
	RegistrationPaymentPaid		bool
	RenewalPaymentPaid		bool
	DirectRegistrationEnabled	bool
	NFTMintedOrBound		bool
	GracePeriodEnabled		bool
	GracePeriodEndHeight		uint64
	AuctionFinalized		bool
	DeterministicWinnerProof	bool
}

func ApplyDomainLifecycleTransitionV2(record DomainRecordV2, ctx DomainLifecycleTransitionContextV2) (DomainRecordV2, error) {
	if err := validateDomainLifecycleTransitionContextV2(ctx); err != nil {
		return DomainRecordV2{}, err
	}
	next := record
	switch record.Status {
	case DomainRecordV2Available:
		switch ctx.Event {
		case DomainLifecycleEventCommitRegistration:
			if !ctx.DepositPaid {
				return DomainRecordV2{}, errors.New("identity lifecycle available->committed requires deposit")
			}
			next.Status = DomainRecordV2Committed
		case DomainLifecycleEventRegisterDirect:
			if !ctx.DirectRegistrationEnabled {
				return DomainRecordV2{}, errors.New("identity lifecycle direct registration class is disabled")
			}
			if !ctx.RegistrationPaymentPaid {
				return DomainRecordV2{}, errors.New("identity lifecycle available->active requires payment")
			}
			if !ctx.NFTMintedOrBound {
				return DomainRecordV2{}, errors.New("identity lifecycle available->active requires nft mint or binding")
			}
			next.Status = DomainRecordV2Active
		case DomainLifecycleEventStartAuction:
			next.Status = DomainRecordV2Auction
		default:
			return DomainRecordV2{}, invalidDomainLifecycleTransitionV2(record.Status, ctx.Event)
		}
	case DomainRecordV2Committed:
		if ctx.Event != DomainLifecycleEventRevealRegistration && ctx.Event != DomainLifecycleEventRegisterDirect {
			return DomainRecordV2{}, invalidDomainLifecycleTransitionV2(record.Status, ctx.Event)
		}
		if !ctx.RegistrationPaymentPaid {
			return DomainRecordV2{}, errors.New("identity lifecycle committed->active requires payment")
		}
		if !ctx.NFTMintedOrBound {
			return DomainRecordV2{}, errors.New("identity lifecycle committed->active requires nft mint or binding")
		}
		next.Status = DomainRecordV2Active
	case DomainRecordV2Active:
		switch ctx.Event {
		case DomainLifecycleEventRenewalWindowBegin:
			if record.RenewalStartHeight == 0 || ctx.Height < record.RenewalStartHeight {
				return DomainRecordV2{}, errors.New("identity lifecycle renewal window has not begun")
			}
			next.Status = DomainRecordV2RenewalWindow
		case DomainLifecycleEventRenewDomain:
			if !ctx.RenewalPaymentPaid {
				return DomainRecordV2{}, errors.New("identity lifecycle active renewal requires payment")
			}
			next.Status = DomainRecordV2Active
		default:
			return DomainRecordV2{}, invalidDomainLifecycleTransitionV2(record.Status, ctx.Event)
		}
	case DomainRecordV2RenewalWindow:
		switch ctx.Event {
		case DomainLifecycleEventRenewDomain:
			if !ctx.RenewalPaymentPaid {
				return DomainRecordV2{}, errors.New("identity lifecycle renewal_window->active requires renewal payment")
			}
			next.Status = DomainRecordV2Active
		case DomainLifecycleEventExpireDomain:
			if record.ExpiryHeight == 0 || ctx.Height < record.ExpiryHeight {
				return DomainRecordV2{}, errors.New("identity lifecycle expiry has not been reached")
			}
			next.Status = DomainRecordV2Expired
		default:
			return DomainRecordV2{}, invalidDomainLifecycleTransitionV2(record.Status, ctx.Event)
		}
	case DomainRecordV2Expired:
		if ctx.Event != DomainLifecycleEventGracePeriodBegin {
			return DomainRecordV2{}, invalidDomainLifecycleTransitionV2(record.Status, ctx.Event)
		}
		if !ctx.GracePeriodEnabled {
			return DomainRecordV2{}, errors.New("identity lifecycle grace period is not enabled")
		}
		next.Status = DomainRecordV2GraceLocked
	case DomainRecordV2GraceLocked:
		switch ctx.Event {
		case DomainLifecycleEventGracePeriodEnd:
			if ctx.GracePeriodEndHeight == 0 || ctx.Height < ctx.GracePeriodEndHeight {
				return DomainRecordV2{}, errors.New("identity lifecycle grace period has not ended")
			}
			next.Status = DomainRecordV2Released
		case DomainLifecycleEventRenewDomain:
			if !ctx.RenewalPaymentPaid {
				return DomainRecordV2{}, errors.New("identity lifecycle grace recovery requires renewal payment")
			}
			next.Status = DomainRecordV2Active
		default:
			return DomainRecordV2{}, invalidDomainLifecycleTransitionV2(record.Status, ctx.Event)
		}
	case DomainRecordV2Released:
		switch ctx.Event {
		case DomainLifecycleEventReleaseDomain:
			next.Status = DomainRecordV2Available
			next.Owner = nil
			next.NFTItemID = ""
		case DomainLifecycleEventStartAuction:
			next.Status = DomainRecordV2Auction
		default:
			return DomainRecordV2{}, invalidDomainLifecycleTransitionV2(record.Status, ctx.Event)
		}
	case DomainRecordV2Auction:
		if ctx.Event != DomainLifecycleEventFinalizeAuction {
			return DomainRecordV2{}, invalidDomainLifecycleTransitionV2(record.Status, ctx.Event)
		}
		if !ctx.AuctionFinalized || !ctx.DeterministicWinnerProof {
			return DomainRecordV2{}, errors.New("identity lifecycle auction->active requires deterministic finalization")
		}
		if !ctx.NFTMintedOrBound {
			return DomainRecordV2{}, errors.New("identity lifecycle auction->active requires nft mint or binding")
		}
		next.Status = DomainRecordV2Active
	default:
		return DomainRecordV2{}, fmt.Errorf("unsupported identity lifecycle state %q", record.Status)
	}
	next.LifecycleEpoch = ctx.Height
	next.UpdatedAtHeight = ctx.Height
	next.Version++
	return next, nil
}

func AutomaticDomainLifecycleStatusV2(record DomainRecordV2, height uint64, gracePeriodEndHeight uint64) DomainRecordV2Status {
	switch record.Status {
	case DomainRecordV2Active:
		if record.RenewalStartHeight != 0 && height >= record.RenewalStartHeight && height < record.ExpiryHeight {
			return DomainRecordV2RenewalWindow
		}
	case DomainRecordV2RenewalWindow:
		if record.ExpiryHeight != 0 && height >= record.ExpiryHeight {
			return DomainRecordV2Expired
		}
	case DomainRecordV2Expired:
		return DomainRecordV2GraceLocked
	case DomainRecordV2GraceLocked:
		if gracePeriodEndHeight != 0 && height >= gracePeriodEndHeight {
			return DomainRecordV2Released
		}
	}
	return record.Status
}

func validateDomainLifecycleTransitionContextV2(ctx DomainLifecycleTransitionContextV2) error {
	if ctx.Height == 0 {
		return errors.New("identity lifecycle transition height is required")
	}
	switch ctx.Event {
	case DomainLifecycleEventCommitRegistration,
		DomainLifecycleEventRevealRegistration,
		DomainLifecycleEventRegisterDirect,
		DomainLifecycleEventRenewDomain,
		DomainLifecycleEventStartAuction,
		DomainLifecycleEventFinalizeAuction,
		DomainLifecycleEventRenewalWindowBegin,
		DomainLifecycleEventExpireDomain,
		DomainLifecycleEventGracePeriodBegin,
		DomainLifecycleEventGracePeriodEnd,
		DomainLifecycleEventReleaseDomain:
		return nil
	default:
		return fmt.Errorf("invalid identity lifecycle event %q", ctx.Event)
	}
}

func invalidDomainLifecycleTransitionV2(status DomainRecordV2Status, event DomainLifecycleEventV2) error {
	return fmt.Errorf("invalid identity lifecycle transition %s -> %s", status, event)
}
