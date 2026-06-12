package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	appparams "github.com/sovereign-l1/l1/app/params"
)

func TestIdentityResolverUpdateCostModelV2StorageDeltaChurnAndInlineFees(t *testing.T) {
	params := DefaultIdentityResolverUpdateCostParamsV2()
	before := costModelResolverRecord(t, "alice.aet", addr(1), addr(2), "")
	after := before
	after.PrimaryAddress = addr(3)
	inlineSchema := `{"type":"object","properties":{"amount":{"type":"string"}}}`
	schemaHash, err := InterfaceDescriptorHashV2(inlineSchema)
	require.NoError(t, err)
	after.InterfaceDescriptors = []InterfaceDescriptorV2{{
		InterfaceID:		"wallet",
		SchemaHash:		schemaHash,
		SchemaInlineOptional:	inlineSchema,
		Version:		"v1",
		RenderPolicy:		"wallet_confirm",
	}}
	after.RecordVersion = 2
	after.UpdatedAtHeight = 20

	delta, err := CalculateIdentityResolverStorageDeltaV2(before, after, params)
	require.NoError(t, err)
	require.Greater(t, delta.AfterBytes, delta.BeforeBytes)
	require.Equal(t, delta.AfterBytes-delta.BeforeBytes, delta.AddedBytes)
	require.Equal(t, EstimateIdentityInlineInterfaceBytesV2(after), delta.InlineBytes)
	require.Greater(t, delta.BillableBytes, delta.NetGrowthBytes)

	quote, err := QuoteIdentityResolverUpdateFeeV2(IdentityResolverUpdateCostRequestV2{
		Before:			before,
		After:			after,
		UpdatedFields:		[]string{"primary", "interface.wallet", "interface.wallet"},
		UpdatesInWindow:	params.PayloadParams.FreeUpdatesPerWindow + 4,
		ProofIndexWrites:	2,
	}, params)
	require.NoError(t, err)
	require.Equal(t, appparams.BaseDenom, quote.Denom)
	require.Equal(t, uint64(2), quote.FieldCount)
	require.True(t, quote.AddedStorageFee.IsPositive())
	require.True(t, quote.InlineInterfaceFee.IsPositive())
	require.True(t, quote.ChurnSurcharge.IsPositive())
	require.Equal(t, params.ProofIndexImpactFee.Mul(sdkmath.NewInt(2)), quote.ProofIndexImpactFee)
	require.True(t, quote.Total.GT(params.PayloadParams.UpdateBaseFee))

	shrunk := after
	shrunk.InterfaceDescriptors = nil
	shrunk.PrimaryAddress = before.PrimaryAddress
	shrunk.RecordVersion = 3
	shrunk.UpdatedAtHeight = 21
	shrinkQuote, err := QuoteIdentityResolverUpdateFeeV2(IdentityResolverUpdateCostRequestV2{
		Before:		after,
		After:		shrunk,
		UpdatedFields:	[]string{"interface.wallet"},
	}, params)
	require.NoError(t, err)
	require.True(t, shrinkQuote.StorageDelta.RemovedBytes > 0)
	require.True(t, shrinkQuote.RemovedByteCredit.IsPositive())
	require.False(t, shrinkQuote.Total.LT(params.PayloadParams.UpdateBaseFee))
}

func TestIdentityResolverBatchFeeAccountingV2NotBelowIndividualOrFloor(t *testing.T) {
	params := DefaultIdentityResolverUpdateCostParamsV2()
	params.BatchCostFloor = params.PayloadParams.UpdateBaseFee.MulRaw(2)
	before := costModelResolverRecord(t, "alice.aet", addr(1), addr(2), "")
	afterOne := before
	afterOne.PrimaryAddress = addr(3)
	afterOne.RecordVersion = 2
	afterOne.UpdatedAtHeight = 20
	afterTwo := before
	afterTwo.PrimaryAddress = addr(4)
	afterTwo.RecordVersion = 2
	afterTwo.UpdatedAtHeight = 20

	batch, err := QuoteIdentityResolverBatchUpdateFeesV2([]IdentityResolverUpdateCostRequestV2{
		{Before: before, After: afterOne, UpdatedFields: []string{"primary"}},
		{Before: before, After: afterTwo, UpdatedFields: []string{"primary"}},
	}, params)
	require.NoError(t, err)
	require.Len(t, batch.ItemQuotes, 2)
	require.True(t, batch.NotCheaperThanIndividual)
	require.False(t, batch.Total.LT(batch.IndividualTotal))
	require.False(t, batch.Total.LT(batch.BatchFloorTotal))
}

func TestIdentitySubdomainCreationCostModelV2OwnerDelegatedAndDetached(t *testing.T) {
	state, parent := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	params := DefaultIdentitySubdomainCostParamsV2()

	ownerPolicy := SubdomainCreationPolicyV2{
		ParentName:	parent.Name,
		Label:		"node",
		Actor:		addr(1),
		ChildOwner:	addr(2),
		Height:		12,
		DelegationType:	SubdomainDelegationOwnerControlledV2,
	}
	ownerQuote, err := QuoteIdentitySubdomainCreationFeeV2(IdentitySubdomainCostRequestV2{
		State:			state,
		Policy:			ownerPolicy,
		ResolverPayloadBytes:	64,
		ZonePolicyComplexity:	1,
	}, params)
	require.NoError(t, err)
	require.Equal(t, "node.alice.aet", ownerQuote.ChildName)
	require.Equal(t, DomainLifecycleActive, ownerQuote.ParentStatus)
	require.Equal(t, addr(1), ownerQuote.ChargedAddress)
	require.True(t, ownerQuote.Total.IsPositive())
	require.True(t, ownerQuote.DetachedRegistrationFee.IsZero())

	nameHash, err := DomainRecordV2NameHash(parent.Name)
	require.NoError(t, err)
	delegation, err := NewDelegationRecordV2(parent.Name, addr(7), DelegationScopeSubdomainCreate, []string{DelegationPermissionCreateV2}, parent.ExpiryHeight, 1, "", 13)
	require.NoError(t, err)
	require.Equal(t, nameHash, delegation.NameHash)
	delegatePolicy := ownerPolicy
	delegatePolicy.Label = "api4"
	delegatePolicy.Actor = addr(7)
	delegatePolicy.ChildOwner = addr(8)
	delegatePolicy.Height = 14
	delegatePolicy.DelegationType = SubdomainDelegationDelegateControlledV2
	delegatePolicy.Delegation = &delegation
	delegateQuote, err := QuoteIdentitySubdomainCreationFeeV2(IdentitySubdomainCostRequestV2{
		State:		state,
		Policy:		delegatePolicy,
		BillingPolicy:	IdentityDelegatedBillingDelegateV2,
	}, params)
	require.NoError(t, err)
	require.Equal(t, IdentityDelegatedBillingDelegateV2, delegateQuote.BillingPolicy)
	require.Equal(t, addr(7), delegateQuote.ChargedAddress)

	detachedPolicy := ownerPolicy
	detachedPolicy.Label = "zone"
	detachedPolicy.ChildOwner = addr(9)
	detachedPolicy.Height = 15
	detachedPolicy.ChildExpiryHeight = parent.ExpiryHeight + params.PricingParams.BaseDurationBlocks
	detachedPolicy.DelegationType = SubdomainDelegationDetachedPaidV2
	detachedPolicy.DetachedPaid = true
	detachedPolicy.IndependentPayment = true
	detachedPolicy.ParentAuthorization = true
	detachedQuote, err := QuoteIdentitySubdomainCreationFeeV2(IdentitySubdomainCostRequestV2{
		State:			state,
		Policy:			detachedPolicy,
		ResolverPayloadBytes:	128,
	}, params)
	require.NoError(t, err)
	require.True(t, detachedQuote.Detached)
	require.True(t, detachedQuote.DetachedRegistrationFee.IsPositive())
	require.True(t, detachedQuote.DetachedRenewalFee.IsPositive())
	require.True(t, detachedQuote.ChildExpiryHeight > detachedQuote.ParentExpiryHeight)
}

func TestIdentitySubdomainCreationCostModelV2ExpiryConstraints(t *testing.T) {
	state, parent := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	params := DefaultIdentitySubdomainCostParamsV2()

	tooLong := SubdomainCreationPolicyV2{
		ParentName:		parent.Name,
		Label:			"node",
		Actor:			addr(1),
		ChildOwner:		addr(2),
		Height:			12,
		ChildExpiryHeight:	parent.ExpiryHeight + 1,
	}
	_, err := QuoteIdentitySubdomainCreationFeeV2(IdentitySubdomainCostRequestV2{State: state, Policy: tooLong}, params)
	require.ErrorContains(t, err, "cannot exceed parent expiry")

	expiredParent := state.Clone()
	for i := range expiredParent.Domains {
		if expiredParent.Domains[i].Name == parent.Name {
			expiredParent.Domains[i].ExpiryHeight = 20
		}
	}
	_, err = QuoteIdentitySubdomainCreationFeeV2(IdentitySubdomainCostRequestV2{
		State:	expiredParent,
		Policy: SubdomainCreationPolicyV2{
			ParentName:	parent.Name,
			Label:		"node",
			Actor:		addr(1),
			ChildOwner:	addr(2),
			Height:		20,
		},
	}, params)
	require.ErrorContains(t, err, "expired")
}

func costModelResolverRecord(t *testing.T, name string, owner sdk.AccAddress, primary sdk.AccAddress, inline string) UnifiedResolutionRecordV2 {
	t.Helper()
	nameHash, err := DomainRecordV2NameHash(name)
	require.NoError(t, err)
	record := UnifiedResolutionRecordV2{
		NameHash:		nameHash,
		Owner:			owner,
		PrimaryAddress:		primary,
		RecordVersion:		1,
		RecordTTL:		100,
		UpdatedAtHeight:	10,
		MaxPayloadBytes:	MaxUnifiedPayloadBytesV2,
		SchemaVersion:		UnifiedResolutionSchemaVersionV2,
	}
	if inline != "" {
		schemaHash, err := InterfaceDescriptorHashV2(inline)
		require.NoError(t, err)
		record.InterfaceDescriptors = []InterfaceDescriptorV2{{
			InterfaceID:		"wallet",
			SchemaHash:		schemaHash,
			SchemaInlineOptional:	inline,
			Version:		"v1",
			RenderPolicy:		"wallet_confirm",
		}}
	}
	require.NoError(t, ValidateUnifiedResolutionRecordV2(record))
	return record
}
