package types

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

type PaymentEnvelope struct {
	Asset			string
	Payer			string
	PayeeService		string
	Denom			string
	Amount			string
	MaxAmountOptional	string
	PricingUnit		coretypes.ServicePricingUnit
	SettlementMode		coretypes.ServicePaymentSettlementMode
	EscrowIDOptional	string
	StreamIDOptional	string
	MeterIDOptional		string
	ExpiryHeight		uint64
	EnvelopeHash		string
}

func NewPaymentEnvelope(envelope PaymentEnvelope) (PaymentEnvelope, error) {
	if envelope.EnvelopeHash != "" {
		return PaymentEnvelope{}, errors.New("services payment envelope hash must be empty before construction")
	}
	envelope = canonicalPaymentEnvelope(envelope)
	if err := envelope.ValidateFormat(); err != nil {
		return PaymentEnvelope{}, err
	}
	envelope.EnvelopeHash = ComputePaymentEnvelopeHash(envelope)
	return envelope, envelope.Validate()
}

func NewPaymentEnvelopeFromDescriptor(descriptor ServiceDescriptor, payer string) (PaymentEnvelope, error) {
	descriptor = coretypes.CanonicalServiceDescriptor(descriptor)
	if err := descriptor.Validate(); err != nil {
		return PaymentEnvelope{}, err
	}
	return NewPaymentEnvelope(PaymentEnvelope{
		Asset:			descriptor.Payment.Denom,
		Payer:			payer,
		PayeeService:		descriptor.ServiceID,
		Denom:			descriptor.Payment.Denom,
		Amount:			descriptor.Payment.Amount,
		MaxAmountOptional:	descriptor.Payment.MaxAmount,
		PricingUnit:		descriptor.Payment.PricingUnit,
		SettlementMode:		descriptor.Payment.SettlementMode,
		EscrowIDOptional:	descriptor.Payment.EscrowID,
		MeterIDOptional:	descriptor.Payment.MeterID,
		ExpiryHeight:		firstNonZeroHeight(descriptor.Payment.ExpiryHeight, descriptor.ExpiryHeight),
	})
}

func (envelope PaymentEnvelope) ValidateFormat() error {
	envelope = canonicalPaymentEnvelope(envelope)
	if err := validateInterfaceToken("services payment envelope asset", envelope.Asset); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("services payment envelope payer", envelope.Payer); err != nil {
		return err
	}
	if err := validateInterfaceToken("services payment envelope payee service", envelope.PayeeService); err != nil {
		return err
	}
	if err := validateInterfaceToken("services payment envelope denom", envelope.Denom); err != nil {
		return err
	}
	if envelope.Asset != envelope.Denom {
		return errors.New("services payment envelope asset must match denom")
	}
	if err := validatePositivePaymentAmount("services payment envelope amount", envelope.Amount); err != nil {
		return err
	}
	if envelope.MaxAmountOptional != "" {
		if err := validatePositivePaymentAmount("services payment envelope max amount", envelope.MaxAmountOptional); err != nil {
			return err
		}
		if comparePaymentAmounts(envelope.MaxAmountOptional, envelope.Amount) < 0 {
			return errors.New("services payment envelope max amount must cover amount")
		}
	}
	if !coretypes.IsServicePricingUnit(envelope.PricingUnit) {
		return fmt.Errorf("unknown services payment envelope pricing unit %q", envelope.PricingUnit)
	}
	if !coretypes.IsServicePaymentSettlementMode(envelope.SettlementMode) {
		return fmt.Errorf("unknown services payment envelope settlement mode %q", envelope.SettlementMode)
	}
	for _, item := range []struct {
		name	string
		value	string
	}{
		{name: "services payment envelope escrow id", value: envelope.EscrowIDOptional},
		{name: "services payment envelope stream id", value: envelope.StreamIDOptional},
		{name: "services payment envelope meter id", value: envelope.MeterIDOptional},
	} {
		if item.value == "" {
			continue
		}
		if err := validateInterfaceToken(item.name, item.value); err != nil {
			return err
		}
	}
	if envelope.ExpiryHeight == 0 {
		return errors.New("services payment envelope expiry height is required")
	}
	if envelope.EnvelopeHash != "" {
		if err := coretypes.ValidateHash("services payment envelope hash", envelope.EnvelopeHash); err != nil {
			return err
		}
	}
	return validatePaymentEnvelopeMode(envelope)
}

func (envelope PaymentEnvelope) Validate() error {
	envelope = canonicalPaymentEnvelope(envelope)
	if err := envelope.ValidateFormat(); err != nil {
		return err
	}
	if envelope.EnvelopeHash == "" {
		return errors.New("services payment envelope hash is required")
	}
	if expected := ComputePaymentEnvelopeHash(envelope); envelope.EnvelopeHash != expected {
		return fmt.Errorf("services payment envelope hash mismatch: expected %s", expected)
	}
	return nil
}

func ComputePaymentEnvelopeHash(envelope PaymentEnvelope) string {
	envelope = canonicalPaymentEnvelope(envelope)
	return servicesHashParts(
		"aetra-services-payment-envelope-v1",
		envelope.Asset,
		envelope.Payer,
		envelope.PayeeService,
		envelope.Denom,
		envelope.Amount,
		envelope.MaxAmountOptional,
		string(envelope.PricingUnit),
		string(envelope.SettlementMode),
		envelope.EscrowIDOptional,
		envelope.StreamIDOptional,
		envelope.MeterIDOptional,
		fmt.Sprint(envelope.ExpiryHeight),
	)
}

func validatePaymentEnvelopeMode(envelope PaymentEnvelope) error {
	switch envelope.SettlementMode {
	case coretypes.ServicePaymentOnChain, coretypes.ServicePaymentPrepaid:
		return nil
	case coretypes.ServicePaymentStreaming:
		if envelope.StreamIDOptional == "" {
			return errors.New("services streaming payment envelope requires stream id")
		}
	case coretypes.ServicePaymentMetered:
		if envelope.MeterIDOptional == "" {
			return errors.New("services metered payment envelope requires meter id")
		}
	case coretypes.ServicePaymentEscrow:
		if envelope.EscrowIDOptional == "" {
			return errors.New("services escrow payment envelope requires escrow id")
		}
	default:
		return fmt.Errorf("unknown services payment envelope settlement mode %q", envelope.SettlementMode)
	}
	return nil
}

func canonicalPaymentEnvelope(envelope PaymentEnvelope) PaymentEnvelope {
	envelope.Asset = strings.TrimSpace(envelope.Asset)
	envelope.Payer = strings.TrimSpace(envelope.Payer)
	envelope.PayeeService = strings.TrimSpace(envelope.PayeeService)
	envelope.Denom = strings.TrimSpace(envelope.Denom)
	envelope.Amount = strings.TrimSpace(envelope.Amount)
	envelope.MaxAmountOptional = strings.TrimSpace(envelope.MaxAmountOptional)
	envelope.EscrowIDOptional = strings.TrimSpace(envelope.EscrowIDOptional)
	envelope.StreamIDOptional = strings.TrimSpace(envelope.StreamIDOptional)
	envelope.MeterIDOptional = strings.TrimSpace(envelope.MeterIDOptional)
	envelope.EnvelopeHash = strings.ToLower(strings.TrimSpace(envelope.EnvelopeHash))
	return envelope
}

func validatePositivePaymentAmount(fieldName, value string) error {
	if err := validateUnifiedAmount(fieldName, value); err != nil {
		return err
	}
	n, _ := new(big.Int).SetString(value, 10)
	if n.Sign() <= 0 {
		return fmt.Errorf("%s must be positive", fieldName)
	}
	return nil
}

func comparePaymentAmounts(left, right string) int {
	l, _ := new(big.Int).SetString(left, 10)
	r, _ := new(big.Int).SetString(right, 10)
	return l.Cmp(r)
}

func firstNonZeroHeight(values ...uint64) uint64 {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}
