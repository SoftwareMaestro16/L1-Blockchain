package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
	servicestypes "github.com/sovereign-l1/l1/x/services/types"
)

func TestServicesKeeperLifecycleFlowsAndIdentityHooks(t *testing.T) {
	k := NewKeeper()
	service := testServicesDescriptor("identity-resolver", coretypes.ServiceTypeOnChain, coretypes.ServiceLocationModule, "identity.aet")
	msg, err := coretypes.NewMsgRegisterService(service.Owner, service, coretypes.ComputeServiceOwnerAuthorizationHash(service.ServiceID, service.Owner, 10))
	require.NoError(t, err)
	require.NoError(t, k.RegisterService(msg))
	require.NoError(t, k.ValidateInvariants())
	assertServiceStoreIsIsolated(t, k)

	registered, err := k.Service(context.Background(), &servicestypes.QueryService{ServiceID: service.ServiceID, IncludeAnchor: true, IncludeProof: true})
	require.NoError(t, err)
	require.True(t, registered.Found)
	require.Equal(t, service.ServiceID, registered.Descriptor.ServiceID)
	require.NotEmpty(t, registered.Anchor.AnchorHash)
	require.NoError(t, registered.Proof.Validate())

	updated := registered.Descriptor
	updated.Discovery.ServiceName = "identity-main"
	updateMsg, err := coretypes.NewMsgUpdateService(updated.Owner, updated, 1)
	require.NoError(t, err)
	require.NoError(t, k.UpdateService(updateMsg))
	serviceAfterUpdate, _ := k.RegistryState().ServiceDescriptorByID(service.ServiceID)
	require.Equal(t, uint64(2), serviceAfterUpdate.Version)
	require.Equal(t, "identity-main", serviceAfterUpdate.Discovery.ServiceName)

	renewMsg, err := coretypes.NewMsgRenewService(service.Owner, service.ServiceID, 200, 2)
	require.NoError(t, err)
	require.NoError(t, k.RenewService(renewMsg))
	serviceAfterRenew, _ := k.RegistryState().ServiceDescriptorByID(service.ServiceID)
	require.Equal(t, uint64(3), serviceAfterRenew.Version)
	require.Equal(t, uint64(200), serviceAfterRenew.ExpiryHeight)

	bindMsg, err := coretypes.NewMsgBindServiceIdentity(service.Owner, service.ServiceID, "identity-v2.aet", 3)
	require.NoError(t, err)
	require.NoError(t, k.BindServiceIdentity(bindMsg))
	byIdentity, err := k.ServicesByIdentity(context.Background(), &servicestypes.QueryServicesByIdentity{IdentityName: "identity-v2.aet"})
	require.NoError(t, err)
	require.Equal(t, uint64(1), byIdentity.Total)

	unbindMsg, err := coretypes.NewMsgUnbindServiceIdentity(service.Owner, service.ServiceID, "identity-v2.aet", 4)
	require.NoError(t, err)
	require.NoError(t, k.UnbindServiceIdentity(unbindMsg))
	byIdentity, err = k.ServicesByIdentity(context.Background(), &servicestypes.QueryServicesByIdentity{IdentityName: "identity-v2.aet"})
	require.NoError(t, err)
	require.Zero(t, byIdentity.Total)

	newOwner := "4:0000000000000000000000000000000000000000000000000000000000000002"
	transferMsg, err := coretypes.NewMsgTransferService(service.Owner, service.ServiceID, newOwner, 5)
	require.NoError(t, err)
	require.NoError(t, k.TransferService(transferMsg))
	transferred, _ := k.RegistryState().ServiceDescriptorByID(service.ServiceID)
	require.Equal(t, newOwner, transferred.Owner)
	require.Equal(t, uint64(6), transferred.Version)

	disableMsg, err := coretypes.NewMsgDisableService(newOwner, service.ServiceID, "owner-request", 6)
	require.NoError(t, err)
	require.NoError(t, k.DisableService(disableMsg))
	disabled, _ := k.RegistryState().ServiceDescriptorByID(service.ServiceID)
	require.Equal(t, coretypes.ServiceStatusDisabled, disabled.Status)
	require.False(t, disabled.Enabled)
}

func TestServicesKeeperProviderReceiptDisputeExportImportAndQueries(t *testing.T) {
	k := NewKeeper()
	service := testServicesDescriptor("fog-compute", coretypes.ServiceTypeFogMarket, coretypes.ServiceLocationProviderPool, "")
	register, err := coretypes.NewMsgRegisterService(service.Owner, service, coretypes.ComputeServiceOwnerAuthorizationHash(service.ServiceID, service.Owner, 10))
	require.NoError(t, err)
	require.NoError(t, k.RegisterService(register))

	provider := testServicesProvider("provider.compute.low", service.Interface.InterfaceHash)
	registerProvider, err := coretypes.NewMsgRegisterProvider(service.Owner, service.ServiceID, provider)
	require.NoError(t, err)
	require.NoError(t, k.RegisterProvider(registerProvider))

	stake, err := coretypes.NewMsgStakeProviderCollateral(service.Owner, service.ServiceID, provider.ProviderID, coretypes.NativeFeePolicyID, "50", 10)
	require.NoError(t, err)
	require.NoError(t, k.StakeProviderCollateral(stake))
	unstake, err := coretypes.NewMsgUnstakeProviderCollateral(service.Owner, service.ServiceID, provider.ProviderID, coretypes.NativeFeePolicyID, "25", 11)
	require.NoError(t, err)
	require.NoError(t, k.UnstakeProviderCollateral(unstake))

	providers, err := k.ProvidersByService(context.Background(), &servicestypes.QueryProvidersByService{ServiceID: service.ServiceID})
	require.NoError(t, err)
	require.Equal(t, uint64(1), providers.Total)
	require.Equal(t, "175", providers.Providers[0].Provider.CollateralAmount)

	receipt := testServicesReceipt(service.ServiceID, "run")
	anchorReceipt, err := coretypes.NewMsgAnchorServiceReceipt(service.Owner, receipt, testServicesHash("receipt/anchor"))
	require.NoError(t, err)
	require.NoError(t, k.AnchorServiceReceipt(anchorReceipt))
	receiptQuery, err := k.ServiceReceipt(context.Background(), &servicestypes.QueryServiceReceipt{ServiceID: service.ServiceID, CallID: receipt.CallID})
	require.NoError(t, err)
	require.True(t, receiptQuery.Found)
	require.Equal(t, receipt.ReceiptHash, receiptQuery.Receipt.ReceiptHash)

	dispute, err := coretypes.NewMsgSubmitServiceDispute(service.Owner, service.ServiceID, receipt.CallID, provider.ProviderID, testServicesHash("dispute/evidence"), "bad-result", 12)
	require.NoError(t, err)
	require.NoError(t, k.SubmitServiceDispute(dispute))
	require.Len(t, k.ExportGenesis().Disputes, 1)

	exported := k.ExportGenesis()
	require.NoError(t, exported.Validate())
	imported := NewKeeper()
	require.NoError(t, imported.InitGenesis(exported))
	require.Equal(t, exported.Registry.StateRoot, imported.ExportGenesis().Registry.StateRoot)
	require.NoError(t, imported.ValidateInvariants())
}

func TestServicesMsgAndQueryServers(t *testing.T) {
	k := NewKeeper()
	msgServer := NewMsgServerImpl(&k)
	service := testServicesDescriptor("identity-resolver", coretypes.ServiceTypeOnChain, coretypes.ServiceLocationModule, "identity.aet")
	register, err := coretypes.NewMsgRegisterService(service.Owner, service, coretypes.ComputeServiceOwnerAuthorizationHash(service.ServiceID, service.Owner, 10))
	require.NoError(t, err)

	_, err = msgServer.RegisterService(context.Background(), &register)
	require.NoError(t, err)

	v2 := service.Interface
	v2.Version = 2
	v2.InterfaceName = "l1.services.v2.identity-resolver"
	v2.MetadataHash = testServicesHash("identity-resolver/interface-v2")
	v2.InterfaceHash = coretypes.ComputeServiceInterfaceHash(v2)
	v2Schema, err := servicestypes.NewInterfaceSchemaFormat(v2, "run")
	require.NoError(t, err)
	registerInterface, err := servicestypes.NewMsgRegisterInterface(service.Owner, v2, v2Schema)
	require.NoError(t, err)
	_, err = msgServer.RegisterInterface(context.Background(), &registerInterface)
	require.NoError(t, err)

	v3 := v2
	v3.Version = 3
	v3.InterfaceName = "l1.services.v3.identity-resolver"
	v3.MetadataHash = testServicesHash("identity-resolver/interface-v3")
	v3.InterfaceHash = coretypes.ComputeServiceInterfaceHash(v3)
	v3Schema, err := servicestypes.NewInterfaceSchemaFormat(v3, "run")
	require.NoError(t, err)
	updateInterface, err := servicestypes.NewMsgUpdateInterface(service.Owner, v2.InterfaceHash, v3, v3Schema, 2)
	require.NoError(t, err)
	_, err = msgServer.UpdateInterface(context.Background(), &updateInterface)
	require.NoError(t, err)

	query := servicestypes.QueryServer(k)
	serviceResponse, err := query.Service(context.Background(), &servicestypes.QueryService{ServiceID: service.ServiceID, IncludeProof: true})
	require.NoError(t, err)
	require.True(t, serviceResponse.Found)
	require.NoError(t, serviceResponse.Proof.Validate())

	interfaceResponse, err := query.ServiceInterface(context.Background(), &servicestypes.QueryServiceInterface{InterfaceHash: v3.InterfaceHash})
	require.NoError(t, err)
	require.True(t, interfaceResponse.Found)
	require.Equal(t, uint64(3), interfaceResponse.Interface.Version)
	interfaceProof, err := servicestypes.QueryInterfaceProofFromState(k.RegistryState(), servicestypes.QueryInterfaceProof{InterfaceHash: v3.InterfaceHash})
	require.NoError(t, err)
	require.True(t, interfaceProof.Found)
	require.NoError(t, interfaceProof.Proof.Validate())

	params, err := query.ServiceParams(context.Background(), &servicestypes.QueryServiceParams{})
	require.NoError(t, err)
	require.Equal(t, coretypes.DefaultAuthority, params.Params.Authority)
}

func TestServicesRegistryInvariantsRejectDescriptorInterfaceMismatch(t *testing.T) {
	k := NewKeeper()
	service := testServicesDescriptor("identity-resolver", coretypes.ServiceTypeOnChain, coretypes.ServiceLocationModule, "identity.aet")
	register, err := coretypes.NewMsgRegisterService(service.Owner, service, coretypes.ComputeServiceOwnerAuthorizationHash(service.ServiceID, service.Owner, 10))
	require.NoError(t, err)
	require.NoError(t, k.RegisterService(register))

	broken := k.ExportGenesis()
	broken.Registry.Descriptors[0].Interface.InterfaceHash = testServicesHash("wrong/interface")
	require.ErrorContains(t, servicestypes.ValidateRegistryInvariants(broken.Registry), "hash mismatch")
}

func testServicesDescriptor(serviceID string, serviceType coretypes.ServiceType, location coretypes.ServiceLocation, identityName string) coretypes.ServiceDescriptor {
	interfaceID := "l1.services.v1." + serviceID
	method := coretypes.ServiceMethodDescriptor{
		MethodID:		"run",
		Name:			"run",
		InputSchemaHash:	testServicesHash(serviceID + "/input"),
		OutputSchemaHash:	testServicesHash(serviceID + "/output"),
		ExecutionType:		coretypes.ServiceMethodAsync,
		RequiredPaymentModel:	"native",
		VerificationModel:	coretypes.ServiceVerificationSignedResult,
		TimeoutHeightDelta:	10,
		IdempotencyRequired:	true,
		FailureBehavior:	coretypes.ServiceFailureRetry,
	}
	if serviceType == coretypes.ServiceTypeOnChain {
		method.ExecutionType = coretypes.ServiceMethodSync
		method.GasModel = coretypes.DefaultGasPolicy
		method.VerificationModel = coretypes.ServiceVerificationConsensusReceipt
		method.FailureBehavior = coretypes.ServiceFailureRevert
	}
	iface := coretypes.ServiceInterfaceDescriptor{
		InterfaceID:	interfaceID,
		InterfaceName:	interfaceID,
		Version:	1,
		SchemaEncoding:	"json-schema-v1",
		Methods:	[]coretypes.ServiceMethodDescriptor{method},
		AuthModel:	"owner",
		PaymentModel:	"native",
		MetadataHash:	testServicesHash(serviceID + "/interface-metadata"),
		CreatedHeight:	1,
	}
	iface.InterfaceHash = coretypes.ComputeServiceInterfaceHash(iface)
	descriptor := coretypes.ServiceDescriptor{
		ServiceID:		serviceID,
		Owner:			coretypes.DefaultAuthority,
		ServiceType:		serviceType,
		ZoneID:			coretypes.ZoneIDApplication,
		InterfaceID:		interfaceID,
		EndpointKey:		serviceID + ".endpoint",
		Version:		1,
		AvailabilityHash:	testServicesHash(serviceID + "/availability"),
		Enabled:		true,
		Status:			coretypes.ServiceStatusActive,
		ExpiryHeight:		100,
		CreatedHeight:		1,
		UpdatedHeight:		1,
		Interface:		iface,
		Execution: coretypes.ServiceExecutionDescriptor{
			Location:		location,
			Target:			serviceID + ".target",
			Endpoint:		"https://" + serviceID + ".aetra.local/v1",
			ProviderPoolID:		serviceID + "-pool",
			Mode:			coretypes.ExecutionModeAsync,
			FailureBehavior:	coretypes.ServiceFailureRetry,
			ResultExpiry:		20,
		},
		Discovery: coretypes.ServiceDiscoveryDescriptor{
			ServiceName:		serviceID,
			IdentityName:		identityName,
			ProviderRoot:		testServicesHash(serviceID + "/providers"),
			MetadataHash:		testServicesHash(serviceID + "/metadata"),
			CacheExpiryHeight:	90,
			SignaturePolicy:	"owner-signature-v1",
		},
		Payment: coretypes.ServicePaymentDescriptor{
			SettlementMode:	coretypes.ServicePaymentOnChain,
			Denom:		coretypes.NativeFeePolicyID,
			Amount:		"1",
			PricingUnit:	coretypes.ServicePricingPerCall,
		},
		Storage:	coretypes.ServiceStorageDescriptor{Model: coretypes.ServiceStorageEphemeral},
		Verification: coretypes.ServiceVerificationDescriptor{
			TrustModel:			coretypes.ServiceTrustFullyTrusted,
			Model:				coretypes.ServiceVerificationSignedResult,
			RequestSigningRequired:		true,
			ResponseSigningRequired:	true,
		},
	}
	switch serviceType {
	case coretypes.ServiceTypeOnChain:
		descriptor.ZoneID = coretypes.ZoneIDIdentity
		descriptor.Execution.Location = coretypes.ServiceLocationModule
		descriptor.Execution.Endpoint = ""
		descriptor.Execution.ProviderPoolID = ""
		descriptor.Execution.ModuleRoute = "identity"
		descriptor.Execution.Mode = coretypes.ExecutionModeSync
		descriptor.Execution.Deterministic = true
		descriptor.Execution.ResultExpiry = 0
		descriptor.Execution.FailureBehavior = coretypes.ServiceFailureRevert
		descriptor.Storage = coretypes.ServiceStorageDescriptor{Model: coretypes.ServiceStorageOnChain, StateRootType: coretypes.StateProofRootType, ProofRequired: true}
		descriptor.Verification = coretypes.ServiceVerificationDescriptor{TrustModel: coretypes.ServiceTrustConsensusExecuted, Model: coretypes.ServiceVerificationConsensusReceipt}
	case coretypes.ServiceTypeFogMarket:
		descriptor.Execution.Location = coretypes.ServiceLocationProviderPool
		descriptor.Verification = coretypes.ServiceVerificationDescriptor{
			TrustModel:			coretypes.ServiceTrustEconomicallySecured,
			Model:				coretypes.ServiceVerificationEconomicCollateral,
			ProviderCollateralDenom:	coretypes.NativeFeePolicyID,
			ProviderCollateralAmount:	"100",
			FaultPolicy:			coretypes.ServiceFailureSlashProvider,
		}
	}
	return coretypes.CanonicalServiceDescriptor(descriptor)
}

func testServicesProvider(providerID, interfaceHash string) coretypes.FogProviderRecord {
	provider := coretypes.FogProviderRecord{
		ProviderID:		providerID,
		IdentityKey:		providerID + ".identity",
		Category:		coretypes.FogCategoryCompute,
		ReputationScore:	80,
		CollateralDenom:	coretypes.NativeFeePolicyID,
		CollateralAmount:	"150",
		StakeAmount:		"150",
		Pricing: coretypes.FogProviderPricing{
			Denom:		coretypes.NativeFeePolicyID,
			Amount:		"3",
			MaxAmount:	"10",
			Unit:		coretypes.FogPricingPerComputeUnit,
			ModelHash:	testServicesHash(providerID + "/pricing"),
		},
		AvailabilityCommitment: coretypes.FogAvailabilityCommitment{
			EndpointHash:		testServicesHash(providerID + "/endpoint"),
			WindowStart:		1,
			WindowEnd:		60,
			UptimeTargetBps:	9_900,
			RenewalNonce:		1,
			SignatureHash:		testServicesHash(providerID + "/availability-signature"),
		},
		SupportedInterfaces:	[]string{interfaceHash},
		Status:			coretypes.FogProviderActive,
		RegisteredHeight:	1,
		UpdatedHeight:		1,
		ExpiryHeight:		80,
	}
	provider.AvailabilityCommitment.CommitmentHash = coretypes.ComputeFogAvailabilityCommitmentHash(provider.AvailabilityCommitment)
	provider = coretypes.CanonicalFogProviderRecord(provider)
	provider.ProviderHash = coretypes.ComputeFogProviderHash(provider)
	return provider
}

func testServicesReceipt(serviceID, methodID string) coretypes.ServiceCallReceipt {
	receipt := coretypes.ServiceCallReceipt{
		CallID:		testServicesHash(fmt.Sprintf("%s/%s/call", serviceID, methodID)),
		ServiceID:	serviceID,
		MethodID:	methodID,
		Caller:		coretypes.DefaultAuthority,
		Status:		coretypes.ServiceCallStatusExecuted,
		RequestHash:	testServicesHash(fmt.Sprintf("%s/%s/request", serviceID, methodID)),
		ResponseHash:	testServicesHash(fmt.Sprintf("%s/%s/response", serviceID, methodID)),
		PaymentStatus:	coretypes.ServicePaymentStatusSettled,
		GasUsed:	1_000,
		ProviderID:	"provider.compute.low",
		ExecutedHeight:	10,
		AnchoredHeight:	10,
	}
	receipt.ReceiptHash = coretypes.ComputeServiceCallReceiptHash(receipt)
	return receipt
}

func testServicesHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func assertServiceStoreIsIsolated(t *testing.T, k Keeper) {
	t.Helper()
	for _, key := range k.StoreKeys() {
		require.True(t, servicestypes.IsServiceStoreKey(key), key)
	}
}
