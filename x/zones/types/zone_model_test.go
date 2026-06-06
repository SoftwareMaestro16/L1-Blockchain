package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExecutionZoneModelSpecCapturesIsolatedZoneSurfaces(t *testing.T) {
	zone := testZone(ZoneIDFinancial, ZoneKindFinancial, VMPolicyNativeModule, 1)
	spec, err := NewExecutionZoneModelSpec(zone)
	require.NoError(t, err)
	require.NoError(t, spec.ValidateHash())
	require.Equal(t, ZoneIDFinancial, spec.ZoneID)
	require.Equal(t, ZoneKindFinancial, spec.Kind)
	require.Equal(t, "zones/FINANCIAL_ZONE/", spec.StatePrefix)
	require.Len(t, spec.Surfaces, 8)
	require.Contains(t, spec.Surfaces, ExecutionZoneSurfaceRule{Surface: ExecutionZoneSurfaceMessageQueues, Requirement: "committed-inbox-outbox-effects"})
	require.Contains(t, spec.Surfaces, ExecutionZoneSurfaceRule{Surface: ExecutionZoneSurfaceExecutionMetrics, Requirement: "zone-execution-summary-inputs"})

	mutated := spec
	mutated.StatePrefix = ZoneKVPrefix(ZoneIDIdentity)
	mutated.BoundaryHash = ComputeExecutionZoneModelHash(mutated)
	require.ErrorContains(t, mutated.ValidateHash(), "state prefix")
}
