package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFogProviderRegistryRegistersProviderWithCollateralPricingAndAvailability(t *testing.T) {
	registry := testFogRegistry(t)
	registry, provider, err := RegisterFogProvider(registry, testFogProvider("provider.compute.1", "3", 80))
	require.NoError(t, err)

	require.Equal(t, FogCategoryCompute, provider.Category)
	require.Equal(t, NativeFeePolicyID, provider.CollateralDenom)
	require.Equal(t, "150", provider.CollateralAmount)
	require.Equal(t, FogPricingPerComputeUnit, provider.Pricing.Unit)
	require.Equal(t, uint64(80), provider.ReputationScore)
	require.NotEmpty(t, provider.AvailabilityCommitment.CommitmentHash)
	require.NotEmpty(t, provider.ProviderHash)
	require.Len(t, registry.Providers, 1)
	require.NoError(t, registry.Validate())
}

func TestFogProviderRegistryRejectsInsufficientCollateral(t *testing.T) {
	registry := testFogRegistry(t)
	_, _, err := RegisterFogProvider(registry, testFogProvider("provider.compute.low", "3", 80, func(provider *FogProviderRecord) {
		provider.CollateralAmount = "25"
	}))
	require.ErrorContains(t, err, "must cover required amount")
}

func TestFogProviderSelectionIsDeterministicByPolicy(t *testing.T) {
	registry := testFogRegistry(t)
	var err error
	registry, _, err = RegisterFogProvider(registry, testFogProvider("provider.compute.1", "5", 90))
	require.NoError(t, err)
	registry, _, err = RegisterFogProvider(registry, testFogProvider("provider.compute.2", "2", 75))
	require.NoError(t, err)
	registry, _, err = RegisterFogProvider(registry, testFogProvider("provider.compute.3", "1", 40))
	require.NoError(t, err)

	selection, err := SelectFogProviders(registry, FogProviderSelectionPolicy{
		Category:		FogCategoryCompute,
		RequiredInterface:	"l1.fog.v1.Compute",
		Strategy:		FogSelectionLowestPrice,
		MinReputation:		70,
		MaxPriceAmount:		"5",
		Limit:			2,
		SelectionNonce:		"selection-1",
	}, 20)
	require.NoError(t, err)
	require.Equal(t, []string{"provider.compute.2", "provider.compute.1"}, selection.ProviderIDs)
	require.NotEmpty(t, selection.SelectionHash)
	require.NoError(t, selection.Validate())
}

func TestFogProviderReputationUpdateRules(t *testing.T) {
	registry := testFogRegistry(t)
	var err error
	registry, _, err = RegisterFogProvider(registry, testFogProvider("provider.compute.1", "3", 50))
	require.NoError(t, err)

	registry, provider, err := UpdateFogProviderReputation(registry, FogReputationEvent{
		ProviderID:	"provider.compute.1",
		Height:		25,
		Successes:	3,
		ScoreDelta:	15,
		Reason:		"receipts-accepted",
	})
	require.NoError(t, err)
	require.Equal(t, uint64(65), provider.ReputationScore)
	require.Len(t, registry.ReputationEvents, 1)
	require.NoError(t, registry.Validate())
}

func TestFogProviderDisputeAndSlashingForProvableFault(t *testing.T) {
	registry := testFogRegistry(t)
	var err error
	registry, _, err = RegisterFogProvider(registry, testFogProvider("provider.compute.1", "3", 80))
	require.NoError(t, err)

	registry, dispute, err := OpenFogProviderDispute(registry, FogProviderDispute{
		ProviderID:		"provider.compute.1",
		Challenger:		"challenger.fog.1",
		FaultClass:		MixedFaultHigh,
		EvidenceHash:		testHash("fog/fault/evidence"),
		OpenedHeight:		30,
		ResolveByHeight:	40,
	})
	require.NoError(t, err)
	require.Equal(t, FogDisputeOpen, dispute.Status)
	require.NotEmpty(t, dispute.DisputeHash)

	registry, slash, err := SlashFogProvider(registry, dispute.DisputeID, true, 35)
	require.NoError(t, err)
	require.Equal(t, "50", slash.PenaltyAmount)
	require.Equal(t, int64(-50), slash.ReputationDelta)
	require.Equal(t, "challenger.fog.1", slash.Recipient)
	provider, found := registry.ProviderByID("provider.compute.1")
	require.True(t, found)
	require.Equal(t, FogProviderSlashed, provider.Status)
	require.Equal(t, uint64(30), provider.ReputationScore)
	require.NoError(t, registry.Validate())
}

func TestFogProviderRenewalExtendsAvailability(t *testing.T) {
	registry := testFogRegistry(t)
	var err error
	registry, _, err = RegisterFogProvider(registry, testFogProvider("provider.compute.1", "3", 80))
	require.NoError(t, err)

	commitment := testFogAvailability("provider.compute.1", 40, 100, 2)
	registry, provider, err := RenewFogProvider(registry, "provider.compute.1", commitment, 120, 45)
	require.NoError(t, err)
	require.Equal(t, uint64(120), provider.ExpiryHeight)
	require.Equal(t, uint64(45), provider.UpdatedHeight)
	require.Equal(t, commitment.CommitmentHash, provider.AvailabilityCommitment.CommitmentHash)
	require.NoError(t, registry.Validate())
}

func testFogRegistry(t *testing.T) FogProviderRegistry {
	t.Helper()
	registry, err := NewFogProviderRegistry(testFogMarketService("fog-compute-v2", ZoneIDApplication))
	require.NoError(t, err)
	require.Equal(t, "compute-pool", registry.ProviderPoolID)
	require.Equal(t, "100", registry.MinCollateralAmount)
	require.NotEmpty(t, registry.RegistryHash)
	return registry
}

func testFogProvider(providerID string, price string, reputation uint64, mutate ...func(*FogProviderRecord)) FogProviderRecord {
	provider := FogProviderRecord{
		ProviderID:		providerID,
		IdentityKey:		providerID + ".identity",
		Category:		FogCategoryCompute,
		ReputationScore:	reputation,
		CollateralDenom:	NativeFeePolicyID,
		CollateralAmount:	"150",
		StakeAmount:		"150",
		Pricing: FogProviderPricing{
			Denom:		NativeFeePolicyID,
			Amount:		price,
			MaxAmount:	"10",
			Unit:		FogPricingPerComputeUnit,
			ModelHash:	testHash(providerID + "/pricing"),
		},
		AvailabilityCommitment:	testFogAvailability(providerID, 1, 60, 1),
		SupportedInterfaces:	[]string{"l1.fog.v1.Compute", "l1.fog.v1.Status"},
		Status:			FogProviderActive,
		RegisteredHeight:	1,
		UpdatedHeight:		1,
		ExpiryHeight:		80,
	}
	for _, apply := range mutate {
		apply(&provider)
	}
	provider = CanonicalFogProviderRecord(provider)
	provider.ProviderHash = ComputeFogProviderHash(provider)
	return provider
}

func testFogAvailability(providerID string, start uint64, end uint64, nonce uint64) FogAvailabilityCommitment {
	commitment := FogAvailabilityCommitment{
		EndpointHash:		testHash(providerID + "/endpoint"),
		WindowStart:		start,
		WindowEnd:		end,
		UptimeTargetBps:	9_900,
		RenewalNonce:		nonce,
		SignatureHash:		testHash(providerID + "/availability/signature"),
	}
	commitment.CommitmentHash = ComputeFogAvailabilityCommitmentHash(commitment)
	return commitment
}
