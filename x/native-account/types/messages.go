package types

import (
	"errors"
	"fmt"
	"strings"
)

const (
	AccountMessageKindExternal = "external"
	AccountMessageKindInternal = "internal"

	AuthModeSingleKey = "single_key"

	InternalMessageSourceModule   = "module"
	InternalMessageSourceContract = "contract"
	InternalMessageSourceSystem   = "system"
)

type ExternalMessage struct {
	AccountUser string
	Sequence    uint64
	Signers     []string
	Operation   string
}

type InternalMessage struct {
	AccountUser            string
	Source                 string
	Feature                string
	Operation              string
	WhitelistedWhileFrozen bool
}

type InternalMessagePolicy struct {
	Version        uint64
	EnabledFeature string
}

func ApplyExternalMessage(account Account, msg ExternalMessage) (Account, error) {
	if err := ValidateExternalMessage(account, msg); err != nil {
		return Account{}, err
	}
	next := cloneAccount(account)
	next.Sequence++
	return next, nil
}

func ValidateExternalMessage(account Account, msg ExternalMessage) error {
	if account.Status == AccountStatusInactive {
		return errors.New("inactive account cannot send external messages")
	}
	if account.Status != AccountStatusActive && account.Status != AccountStatusRecovered {
		return fmt.Errorf("%s account cannot send external messages", account.Status)
	}
	if msg.AccountUser != account.AddressUser {
		return errors.New("external message account address mismatch")
	}
	if msg.Sequence != account.Sequence {
		return fmt.Errorf("external message sequence %d does not match account sequence %d", msg.Sequence, account.Sequence)
	}
	return ValidateAuthPolicyForExternalMessage(account, msg)
}

func ValidateAuthPolicyForExternalMessage(account Account, msg ExternalMessage) error {
	if err := account.AuthPolicy.Validate(); err != nil {
		return err
	}
	switch account.AuthPolicy.Mode {
	case AuthModeSingleKey:
		for _, signer := range msg.Signers {
			signer = strings.TrimSpace(signer)
			for _, pubKey := range account.PubKeys {
				if signer == pubKey {
					return nil
				}
			}
		}
		return errors.New("external message missing authorized single-key signer")
	default:
		return fmt.Errorf("unsupported external auth policy mode %q", account.AuthPolicy.Mode)
	}
}

func ApplyInternalMessage(account Account, msg InternalMessage, policy InternalMessagePolicy) (Account, error) {
	if err := ValidateInternalMessage(account, msg, policy); err != nil {
		return Account{}, err
	}
	return cloneAccount(account), nil
}

func ValidateInternalMessage(account Account, msg InternalMessage, policy InternalMessagePolicy) error {
	if msg.AccountUser != account.AddressUser {
		return errors.New("internal message account address mismatch")
	}
	if account.Status == AccountStatusInactive || account.Status == AccountStatusClosed || account.Status == AccountStatusArchived {
		return fmt.Errorf("%s account cannot receive internal messages", account.Status)
	}
	if account.Status == AccountStatusFrozen && !msg.WhitelistedWhileFrozen {
		return errors.New("frozen account internal messages require explicit whitelist")
	}
	if policy.Version == 0 {
		return errors.New("internal message policy version must be positive")
	}
	if strings.TrimSpace(policy.EnabledFeature) == "" {
		return errors.New("internal message policy feature is required")
	}
	if strings.TrimSpace(msg.Feature) == "" || msg.Feature != policy.EnabledFeature {
		return errors.New("internal message feature is not enabled by policy")
	}
	if !accountHasFeature(account, msg.Feature) {
		return errors.New("internal message feature disabled on account")
	}
	switch msg.Source {
	case InternalMessageSourceModule, InternalMessageSourceContract, InternalMessageSourceSystem:
		return nil
	default:
		return fmt.Errorf("unsupported internal message source %q", msg.Source)
	}
}

func accountHasFeature(account Account, feature string) bool {
	for _, existing := range account.FeatureFlags {
		if existing == feature {
			return true
		}
	}
	return false
}
