package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServiceRegistryBuildsDeterministicRecordsAndProofLookup(t *testing.T) {
	services := []ServiceDescriptor{
		testService("identity-resolver", ZoneIDIdentity),
		testOffChainService("indexer-feed", ZoneIDApplication),
		testMixedService("hybrid-storage", ZoneIDApplication),
		testFogMarketService("fog-compute", ZoneIDApplication),
	}

	registry, err := NewServiceRegistry(ServiceRegistryHybrid, services, 10)
	require.NoError(t, err)
	require.Equal(t, ServiceRegistryHybrid, registry.Mode)
	require.Len(t, registry.Records, 4)
	require.NotEmpty(t, registry.RegistryRoot)
	require.NoError(t, registry.Validate())

	record, proof, found := registry.Lookup("identity-resolver")
	require.True(t, found)
	require.Equal(t, ServiceTypeOnChain, record.ServiceType)
	require.Equal(t, "identity", record.ModuleRoute)
	require.Equal(t, ServiceTrustConsensusExecuted, record.TrustModel)
	require.Equal(t, ServiceVerificationConsensusReceipt, record.VerificationModel)
	require.NoError(t, proof.ValidateForRegistry(registry))

	tampered := proof
	tampered.RecordHash = testHash("tampered/record")
	require.ErrorContains(t, tampered.ValidateForRegistry(registry), "does not match registry")
}

func TestServiceRegistryDiscoversPaymentTrustAndInterfaces(t *testing.T) {
	services := []ServiceDescriptor{
		testMixedService("hybrid-storage", ZoneIDApplication),
		testFogMarketService("fog-compute", ZoneIDApplication),
	}
	registry, err := NewServiceRegistry(ServiceRegistryOnChain, services, 12)
	require.NoError(t, err)

	payment, found := registry.PaymentDiscovery("hybrid-storage")
	require.True(t, found)
	require.Equal(t, ServicePaymentEscrow, payment.SettlementMode)
	require.Equal(t, NativeFeePolicyID, payment.Denom)
	require.Equal(t, "5", payment.Amount)
	require.Equal(t, ServicePricingPerByte, payment.PricingUnit)
	require.True(t, payment.EscrowRequired)
	require.Equal(t, "storage-escrow", payment.EscrowID)
	require.NotEmpty(t, payment.DiscoveryHash)

	trust, found := registry.TrustDiscovery("fog-compute")
	require.True(t, found)
	require.Equal(t, ServiceTrustEconomicallySecured, trust.TrustModel)
	require.Equal(t, ServiceVerificationEconomicCollateral, trust.VerificationModel)
	require.Equal(t, "100", trust.CollateralAmount)
	require.NotEmpty(t, trust.DiscoveryHash)

	require.Len(t, registry.InterfaceIndex, 2)
	for _, entry := range registry.InterfaceIndex {
		require.NotEmpty(t, entry.RecordHash)
		require.NoError(t, entry.Validate())
	}
}

func TestServiceRegistryLifecycleRenewUpdateAndExpire(t *testing.T) {
	service := testOffChainService("indexer-feed", ZoneIDApplication)
	registry, err := NewServiceRegistry(ServiceRegistryMesh, []ServiceDescriptor{service}, 20)
	require.NoError(t, err)

	registry, renewed, err := RenewServiceRegistryRecord(registry, "indexer-feed", 200, 30)
	require.NoError(t, err)
	require.Equal(t, uint64(200), renewed.ExpiryHeight)
	require.Equal(t, ServiceStatusActive, renewed.Status)

	updated := renewed
	updated.Reputation = 77
	updated.Version = 2
	updated.UpdatedHeight = 35
	updated.RecordHash = ComputeServiceRegistryRecordHash(updated)
	registry, err = UpdateServiceRegistryRecord(registry, updated, 35)
	require.NoError(t, err)
	record, found := registry.RecordByID("indexer-feed")
	require.True(t, found)
	require.Equal(t, uint64(77), record.Reputation)
	require.Equal(t, uint64(2), record.Version)

	registry, expired, err := ExpireServiceRegistryRecords(registry, 201)
	require.NoError(t, err)
	require.Equal(t, []string{"indexer-feed"}, expired)
	record, found = registry.RecordByID("indexer-feed")
	require.True(t, found)
	require.Equal(t, ServiceStatusDisabled, record.Status)
	require.NoError(t, registry.Validate())
}

func TestServiceRegistryRejectsDuplicateRegistrationAndBadMode(t *testing.T) {
	service := testService("identity-resolver", ZoneIDIdentity)
	record, err := NewServiceRegistryRecord(service)
	require.NoError(t, err)
	registry, err := NewServiceRegistry(ServiceRegistryOnChain, []ServiceDescriptor{service}, 10)
	require.NoError(t, err)

	_, err = RegisterServiceRegistryRecord(registry, record, 11)
	require.ErrorContains(t, err, "already exists")

	_, err = NewServiceRegistry(ServiceRegistryMode("bad"), []ServiceDescriptor{service}, 10)
	require.ErrorContains(t, err, "unknown")
}

func TestServiceRegistryRecordProjectionCoversExtendedFields(t *testing.T) {
	service := testFogMarketService("fog-compute", ZoneIDApplication)
	record, err := NewServiceRegistryRecord(service)
	require.NoError(t, err)

	require.Equal(t, "fog-compute", record.ServiceID)
	require.Equal(t, ServiceTypeFogMarket, record.ServiceType)
	require.Equal(t, "compute-pool", record.Endpoint)
	require.Equal(t, ExecutionModeAsync, record.ExecutionMode)
	require.Equal(t, "100", record.Stake)
	require.Equal(t, "100", record.Collateral)
	require.Equal(t, testHash("fog-compute/providers"), record.ProviderSet)
	require.Equal(t, ServiceVerificationEconomicCollateral, record.VerificationModel)
	require.Equal(t, ServiceTrustEconomicallySecured, record.TrustModel)
	require.Equal(t, ComputeServiceDescriptorHash(service), record.DescriptorHash)
	require.NoError(t, record.Validate())
}
