package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/app/addressing"
	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

func TestPaymentEnvelopeFromDescriptor(t *testing.T) {
	descriptor := testInterfaceSystemDescriptor()
	descriptor.Payment.MaxAmount = "10"
	descriptor.Payment.ExpiryHeight = 80
	descriptor = coretypes.CanonicalServiceDescriptor(descriptor)

	payer := testPaymentPayer()
	envelope, err := NewPaymentEnvelopeFromDescriptor(descriptor, payer)
	require.NoError(t, err)
	require.Equal(t, descriptor.Payment.Denom, envelope.Asset)
	require.Equal(t, payer, envelope.Payer)
	require.Equal(t, descriptor.ServiceID, envelope.PayeeService)
	require.Equal(t, descriptor.Payment.Denom, envelope.Denom)
	require.Equal(t, descriptor.Payment.Amount, envelope.Amount)
	require.Equal(t, descriptor.Payment.MaxAmount, envelope.MaxAmountOptional)
	require.Equal(t, descriptor.Payment.PricingUnit, envelope.PricingUnit)
	require.Equal(t, descriptor.Payment.SettlementMode, envelope.SettlementMode)
	require.Equal(t, uint64(80), envelope.ExpiryHeight)
	require.NotEmpty(t, envelope.EnvelopeHash)
	require.NoError(t, envelope.Validate())
}

func TestPaymentEnvelopeSettlementModes(t *testing.T) {
	for _, tc := range []struct {
		name	string
		mode	coretypes.ServicePaymentSettlementMode
		escrow	string
		stream	string
		meter	string
	}{
		{name: "on_chain", mode: coretypes.ServicePaymentOnChain},
		{name: "prepaid", mode: coretypes.ServicePaymentPrepaid},
		{name: "streaming", mode: coretypes.ServicePaymentStreaming, stream: "stream-1"},
		{name: "metered", mode: coretypes.ServicePaymentMetered, meter: "meter-1"},
		{name: "escrow", mode: coretypes.ServicePaymentEscrow, escrow: "escrow-1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			envelope, err := NewPaymentEnvelope(PaymentEnvelope{
				Asset:			coretypes.NativeFeePolicyID,
				Payer:			testPaymentPayer(),
				PayeeService:		"portable-service",
				Denom:			coretypes.NativeFeePolicyID,
				Amount:			"5",
				MaxAmountOptional:	"10",
				PricingUnit:		coretypes.ServicePricingPerCall,
				SettlementMode:		tc.mode,
				EscrowIDOptional:	tc.escrow,
				StreamIDOptional:	tc.stream,
				MeterIDOptional:	tc.meter,
				ExpiryHeight:		100,
			})
			require.NoError(t, err)
			require.Equal(t, tc.mode, envelope.SettlementMode)
			require.NoError(t, envelope.Validate())
		})
	}
}

func TestPaymentEnvelopeRejectsMalformedOrIncompletePayments(t *testing.T) {
	base := PaymentEnvelope{
		Asset:		coretypes.NativeFeePolicyID,
		Payer:		testPaymentPayer(),
		PayeeService:	"portable-service",
		Denom:		coretypes.NativeFeePolicyID,
		Amount:		"5",
		PricingUnit:	coretypes.ServicePricingPerCall,
		SettlementMode:	coretypes.ServicePaymentOnChain,
		ExpiryHeight:	100,
	}

	zero := base
	zero.Amount = "0"
	_, err := NewPaymentEnvelope(zero)
	require.ErrorContains(t, err, "positive")

	underMax := base
	underMax.MaxAmountOptional = "4"
	_, err = NewPaymentEnvelope(underMax)
	require.ErrorContains(t, err, "cover amount")

	badPayer := base
	badPayer.Payer = "bad"
	_, err = NewPaymentEnvelope(badPayer)
	require.ErrorContains(t, err, "must use AE user-facing address format")

	streaming := base
	streaming.SettlementMode = coretypes.ServicePaymentStreaming
	_, err = NewPaymentEnvelope(streaming)
	require.ErrorContains(t, err, "stream id")

	metered := base
	metered.SettlementMode = coretypes.ServicePaymentMetered
	_, err = NewPaymentEnvelope(metered)
	require.ErrorContains(t, err, "meter id")

	escrow := base
	escrow.SettlementMode = coretypes.ServicePaymentEscrow
	_, err = NewPaymentEnvelope(escrow)
	require.ErrorContains(t, err, "escrow id")
}

func TestPaymentEnvelopeRejectsHashTampering(t *testing.T) {
	envelope, err := NewPaymentEnvelope(PaymentEnvelope{
		Asset:		coretypes.NativeFeePolicyID,
		Payer:		testPaymentPayer(),
		PayeeService:	"portable-service",
		Denom:		coretypes.NativeFeePolicyID,
		Amount:		"5",
		PricingUnit:	coretypes.ServicePricingPerCall,
		SettlementMode:	coretypes.ServicePaymentPrepaid,
		ExpiryHeight:	100,
	})
	require.NoError(t, err)

	envelope.Amount = "6"
	require.ErrorContains(t, envelope.Validate(), "hash mismatch")
}

func testPaymentPayer() string {
	return addressing.FormatAccAddress(sdk.AccAddress{
		0x61, 0x61, 0x61, 0x61, 0x61,
		0x61, 0x61, 0x61, 0x61, 0x61,
		0x61, 0x61, 0x61, 0x61, 0x61,
		0x61, 0x61, 0x61, 0x61, 0x61,
	})
}
