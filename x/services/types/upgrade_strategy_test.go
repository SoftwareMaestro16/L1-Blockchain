package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServiceUpgradeVersionRecordsAndSchemaCompatibility(t *testing.T) {
	compat := testUpgradeCompatibility(t, ServiceSchemaCompatibilityAdditive)
	require.Equal(t, []string{"field.memo", "field.timeout"}, compat.AddedFields)
	require.NoError(t, compat.Validate())

	record, err := NewServiceDescriptorVersionRecord(ServiceDescriptorVersionRecord{
		ObjectKind:			ServiceDescriptorObjectCanonicalDescriptor,
		ObjectID:			"svc.dex",
		Version:			2,
		DescriptorHash:			testInterfaceHash("upgrade/descriptor/v2"),
		InterfaceHash:			testInterfaceHash("upgrade/interface/v2"),
		SchemaCompatibilityHash:	compat.CompatibilityHash,
	})
	require.NoError(t, err)
	require.NoError(t, record.Validate())
	require.Equal(t, uint64(2), record.Version)
	require.NotEmpty(t, record.VersionHash)

	_, err = NewServiceSchemaCompatibilityMetadata(ServiceSchemaCompatibilityMetadata{
		SchemaID:	"svc.dex.schema",
		PreviousHash:	testInterfaceHash("upgrade/schema/v1"),
		NextHash:	testInterfaceHash("upgrade/schema/v2"),
		Mode:		ServiceSchemaCompatibilityBackward,
		RemovedFields:	[]string{"field.price"},
		RequiredFields:	[]string{"field.amount"},
	})
	require.ErrorContains(t, err, "cannot remove fields")
}

func TestServiceRegistryMigrationHandlersRequireGovernanceAndContiguousChain(t *testing.T) {
	handler, err := NewServiceRegistryMigrationHandler(ServiceRegistryMigrationHandler{
		FromRegistryVersion:		1,
		ToRegistryVersion:		2,
		HandlerName:			"registry_v1_to_v2",
		GovernanceProposalID:		ServiceUpgradeGovernanceProposalPrefix + "42",
		DescriptorMigrationHandler:	"migrate_descriptors_v2",
		InterfaceMigrationHandler:	"migrate_interfaces_v2",
		ProviderMigrationHandler:	"migrate_providers_v2",
	})
	require.NoError(t, err)
	require.NoError(t, handler.Validate())

	_, err = NewServiceRegistryMigrationHandler(ServiceRegistryMigrationHandler{
		FromRegistryVersion:		1,
		ToRegistryVersion:		2,
		HandlerName:			"registry_v1_to_v2",
		GovernanceProposalID:		"local/admin",
		DescriptorMigrationHandler:	"migrate_descriptors_v2",
		InterfaceMigrationHandler:	"migrate_interfaces_v2",
		ProviderMigrationHandler:	"migrate_providers_v2",
	})
	require.ErrorContains(t, err, "governance proposal")

	next, err := NewServiceRegistryMigrationHandler(ServiceRegistryMigrationHandler{
		FromRegistryVersion:		3,
		ToRegistryVersion:		4,
		HandlerName:			"registry_v3_to_v4",
		GovernanceProposalID:		ServiceUpgradeGovernanceProposalPrefix + "42",
		DescriptorMigrationHandler:	"migrate_descriptors_v4",
		InterfaceMigrationHandler:	"migrate_interfaces_v4",
		ProviderMigrationHandler:	"migrate_providers_v4",
	})
	require.NoError(t, err)
	require.ErrorContains(t, validateMigrationHandlerChain(1, 4, []ServiceRegistryMigrationHandler{handler, next}), "gap")
}

func TestServiceInterfaceDeprecationFlow(t *testing.T) {
	marker, err := NewServiceInterfaceDeprecationMarker(ServiceInterfaceDeprecationMarker{
		InterfaceHash:			testInterfaceHash("upgrade/interface/v1"),
		Version:			1,
		DeprecatedHeight:		100,
		RetirementHeight:		200,
		ReplacementInterfaceHash:	testInterfaceHash("upgrade/interface/v2"),
		Reason:				"replace_with_v2",
	})
	require.NoError(t, err)
	require.NoError(t, marker.Validate())

	status, err := InterfaceLifecycleStatusAt(marker, 99)
	require.NoError(t, err)
	require.Equal(t, ServiceInterfaceLifecycleActive, status)
	status, err = InterfaceLifecycleStatusAt(marker, 150)
	require.NoError(t, err)
	require.Equal(t, ServiceInterfaceLifecycleDeprecated, status)
	status, err = InterfaceLifecycleStatusAt(marker, 200)
	require.NoError(t, err)
	require.Equal(t, ServiceInterfaceLifecycleRetired, status)

	_, err = NewServiceInterfaceDeprecationMarker(ServiceInterfaceDeprecationMarker{
		InterfaceHash:		testInterfaceHash("upgrade/interface/bad"),
		Version:		1,
		DeprecatedHeight:	100,
		RetirementHeight:	100,
		Reason:			"bad_window",
	})
	require.ErrorContains(t, err, "after deprecation")
}

func TestServiceProviderReregistrationRules(t *testing.T) {
	rule, err := NewServiceProviderReregistrationRule(ServiceProviderReregistrationRule{
		ProviderID:			"provider.storage",
		ServiceID:			"svc.storage",
		PreviousInterfaceHash:		testInterfaceHash("upgrade/provider/iface/v1"),
		NextInterfaceHash:		testInterfaceHash("upgrade/provider/iface/v2"),
		Mode:				ServiceProviderReregistrationFullReregistration,
		RequiresOwnerAuthorization:	true,
		RequiresCollateralRefresh:	true,
		RequiresCapabilityRefresh:	true,
		EarliestHeight:			250,
	})
	require.NoError(t, err)
	require.NoError(t, rule.Validate())
	require.NotEmpty(t, rule.RuleHash)

	_, err = NewServiceProviderReregistrationRule(ServiceProviderReregistrationRule{
		ProviderID:		"provider.storage",
		ServiceID:		"svc.storage",
		PreviousInterfaceHash:	testInterfaceHash("upgrade/provider/iface/v1"),
		NextInterfaceHash:	testInterfaceHash("upgrade/provider/iface/v2"),
		Mode:			ServiceProviderReregistrationUnchanged,
		EarliestHeight:		250,
	})
	require.ErrorContains(t, err, "requires re-registration mode")

	_, err = NewServiceProviderReregistrationRule(ServiceProviderReregistrationRule{
		ProviderID:			"provider.storage",
		ServiceID:			"svc.storage",
		PreviousInterfaceHash:		testInterfaceHash("upgrade/provider/iface/v1"),
		NextInterfaceHash:		testInterfaceHash("upgrade/provider/iface/v2"),
		Mode:				ServiceProviderReregistrationFullReregistration,
		RequiresOwnerAuthorization:	true,
		RequiresCollateralRefresh:	true,
		EarliestHeight:			250,
	})
	require.ErrorContains(t, err, "full re-registration")
}

func TestServiceRegistryUpgradeSimulationCompatiblePlan(t *testing.T) {
	plan := testUpgradePlan(t, false, false)
	result, err := SimulateServiceRegistryUpgrade(plan)
	require.NoError(t, err)
	require.NoError(t, result.Validate())
	require.True(t, result.GovernanceControlled)
	require.True(t, result.BackwardCompatible)
	require.Equal(t, uint32(2), result.DescriptorVersionsMigrated)
	require.Equal(t, uint32(2), result.BackwardCompatibleSchemas)
	require.Equal(t, uint32(0), result.BreakingSchemas)
	require.Equal(t, uint32(2), result.MigrationHandlersApplied)
	require.Equal(t, uint32(1), result.InterfaceDeprecationsApplied)
	require.Equal(t, uint32(0), result.ProvidersRequiringReregistration)
	require.NotEmpty(t, result.SimulationHash)
}

func TestServiceRegistryUpgradeSimulationFlagsBreakingAndProviderReregistration(t *testing.T) {
	plan := testUpgradePlan(t, true, true)
	result, err := SimulateServiceRegistryUpgrade(plan)
	require.NoError(t, err)
	require.False(t, result.BackwardCompatible)
	require.Equal(t, uint32(1), result.BreakingSchemas)
	require.Equal(t, uint32(1), result.ProvidersRequiringReregistration)
}

func TestServiceRegistryUpgradePlanRejectsMissingMigrationChain(t *testing.T) {
	plan := testUpgradePlan(t, false, false)
	plan.MigrationHandlers = plan.MigrationHandlers[:1]
	plan.PlanHash = ComputeServiceRegistryUpgradePlanHash(plan)
	require.ErrorContains(t, plan.Validate(), "chain ended")
}

func testUpgradePlan(t *testing.T, includeBreaking bool, requireProviderReregistration bool) ServiceRegistryUpgradePlan {
	t.Helper()
	compat := []ServiceSchemaCompatibilityMetadata{
		testUpgradeCompatibility(t, ServiceSchemaCompatibilityAdditive),
		testUpgradeCompatibilityWithID(t, "svc.dex.output", ServiceSchemaCompatibilityBackward),
	}
	if includeBreaking {
		compat[1] = testUpgradeCompatibilityWithID(t, "svc.dex.output", ServiceSchemaCompatibilityBreaking)
	}
	records := make([]ServiceDescriptorVersionRecord, 0, 2)
	for idx, objectKind := range []ServiceDescriptorObjectKind{ServiceDescriptorObjectCanonicalDescriptor, ServiceDescriptorObjectInterfaceDescriptor} {
		record, err := NewServiceDescriptorVersionRecord(ServiceDescriptorVersionRecord{
			ObjectKind:			objectKind,
			ObjectID:			"svc.dex",
			Version:			uint64(idx + 2),
			DescriptorHash:			testInterfaceHash("upgrade/descriptor/record/" + string(objectKind)),
			InterfaceHash:			testInterfaceHash("upgrade/interface/record/" + string(objectKind)),
			SchemaCompatibilityHash:	compat[idx].CompatibilityHash,
		})
		require.NoError(t, err)
		records = append(records, record)
	}
	handlers := []ServiceRegistryMigrationHandler{
		testMigrationHandler(t, 1, 2),
		testMigrationHandler(t, 2, 3),
	}
	marker, err := NewServiceInterfaceDeprecationMarker(ServiceInterfaceDeprecationMarker{
		InterfaceHash:			testInterfaceHash("upgrade/plan/interface/v1"),
		Version:			1,
		DeprecatedHeight:		100,
		RetirementHeight:		300,
		ReplacementInterfaceHash:	testInterfaceHash("upgrade/plan/interface/v2"),
		Reason:				"replace_with_v2",
	})
	require.NoError(t, err)
	providerMode := ServiceProviderReregistrationUnchanged
	previousInterfaceHash := testInterfaceHash("upgrade/plan/provider/iface/v1")
	nextInterfaceHash := previousInterfaceHash
	requiresOwner := false
	requiresCollateral := false
	requiresCapability := false
	if requireProviderReregistration {
		providerMode = ServiceProviderReregistrationFullReregistration
		nextInterfaceHash = testInterfaceHash("upgrade/plan/provider/iface/v2")
		requiresOwner = true
		requiresCollateral = true
		requiresCapability = true
	}
	provider, err := NewServiceProviderReregistrationRule(ServiceProviderReregistrationRule{
		ProviderID:			"provider.dex",
		ServiceID:			"svc.dex",
		PreviousInterfaceHash:		previousInterfaceHash,
		NextInterfaceHash:		nextInterfaceHash,
		Mode:				providerMode,
		RequiresOwnerAuthorization:	requiresOwner,
		RequiresCollateralRefresh:	requiresCollateral,
		RequiresCapabilityRefresh:	requiresCapability,
		EarliestHeight:			300,
	})
	require.NoError(t, err)
	plan, err := NewServiceRegistryUpgradePlan(ServiceRegistryUpgradePlan{
		FromRegistryVersion:	1,
		ToRegistryVersion:	3,
		GovernanceProposalID:	ServiceUpgradeGovernanceProposalPrefix + "77",
		DescriptorVersions:	records,
		SchemaCompatibility:	compat,
		MigrationHandlers:	handlers,
		InterfaceDeprecations:	[]ServiceInterfaceDeprecationMarker{marker},
		ProviderReregistration:	[]ServiceProviderReregistrationRule{provider},
	})
	require.NoError(t, err)
	require.NoError(t, plan.Validate())
	return plan
}

func testUpgradeCompatibility(t *testing.T, mode ServiceSchemaCompatibilityMode) ServiceSchemaCompatibilityMetadata {
	t.Helper()
	return testUpgradeCompatibilityWithID(t, "svc.dex.input", mode)
}

func testUpgradeCompatibilityWithID(t *testing.T, schemaID string, mode ServiceSchemaCompatibilityMode) ServiceSchemaCompatibilityMetadata {
	t.Helper()
	metadata := ServiceSchemaCompatibilityMetadata{
		SchemaID:		schemaID,
		PreviousHash:		testInterfaceHash(schemaID + "/v1"),
		NextHash:		testInterfaceHash(schemaID + "/v2"),
		Mode:			mode,
		DeprecatedFields:	[]string{"field.legacy"},
		RequiredFields:		[]string{"field.amount"},
	}
	switch mode {
	case ServiceSchemaCompatibilityAdditive:
		metadata.AddedFields = []string{"field.timeout", "field.memo"}
	case ServiceSchemaCompatibilityBreaking:
		metadata.RemovedFields = []string{"field.legacy"}
	}
	out, err := NewServiceSchemaCompatibilityMetadata(metadata)
	require.NoError(t, err)
	return out
}

func testMigrationHandler(t *testing.T, from, to uint64) ServiceRegistryMigrationHandler {
	t.Helper()
	handler, err := NewServiceRegistryMigrationHandler(ServiceRegistryMigrationHandler{
		FromRegistryVersion:		from,
		ToRegistryVersion:		to,
		HandlerName:			"registry_migration",
		GovernanceProposalID:		ServiceUpgradeGovernanceProposalPrefix + "77",
		DescriptorMigrationHandler:	"migrate_descriptors",
		InterfaceMigrationHandler:	"migrate_interfaces",
		ProviderMigrationHandler:	"migrate_providers",
	})
	require.NoError(t, err)
	return handler
}
