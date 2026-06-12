package types

import (
	"errors"
	"fmt"
	"sort"
)

type ExecutionZoneSurface string

const (
	ExecutionZoneSurfaceKeeperSet		ExecutionZoneSurface	= "module-keeper-set"
	ExecutionZoneSurfaceStatePrefix		ExecutionZoneSurface	= "state-prefix"
	ExecutionZoneSurfaceMempoolPolicy	ExecutionZoneSurface	= "mempool-policy"
	ExecutionZoneSurfaceShardLayout		ExecutionZoneSurface	= "shard-layout"
	ExecutionZoneSurfaceMessageQueues	ExecutionZoneSurface	= "message-inbox-outbox"
	ExecutionZoneSurfaceFeePolicy		ExecutionZoneSurface	= "fee-policy"
	ExecutionZoneSurfaceProofRoot		ExecutionZoneSurface	= "proof-root"
	ExecutionZoneSurfaceExecutionMetrics	ExecutionZoneSurface	= "execution-metrics"
)

type ExecutionZoneSurfaceRule struct {
	Surface		ExecutionZoneSurface
	Requirement	string
}

type ExecutionZoneModelSpec struct {
	ZoneID		ZoneID
	Kind		ZoneKind
	StatePrefix	string
	Surfaces	[]ExecutionZoneSurfaceRule
	BoundaryHash	string
}

func NewExecutionZoneModelSpec(zone Zone) (ExecutionZoneModelSpec, error) {
	if err := zone.Validate(); err != nil {
		return ExecutionZoneModelSpec{}, err
	}
	spec := ExecutionZoneModelSpec{
		ZoneID:		zone.ID,
		Kind:		zone.Kind,
		StatePrefix:	ZoneKVPrefix(zone.ID),
		Surfaces: []ExecutionZoneSurfaceRule{
			{Surface: ExecutionZoneSurfaceKeeperSet, Requirement: "zone-adapter-no-direct-cross-zone-mutation"},
			{Surface: ExecutionZoneSurfaceStatePrefix, Requirement: "prefix-isolated-exportable-state"},
			{Surface: ExecutionZoneSurfaceMempoolPolicy, Requirement: "local-admission-with-global-bounds"},
			{Surface: ExecutionZoneSurfaceShardLayout, Requirement: "committed-layout-epoch-routing"},
			{Surface: ExecutionZoneSurfaceMessageQueues, Requirement: "committed-inbox-outbox-effects"},
			{Surface: ExecutionZoneSurfaceFeePolicy, Requirement: "zone-local-accounting-global-settlement"},
			{Surface: ExecutionZoneSurfaceProofRoot, Requirement: "state-message-receipt-event-domain-roots"},
			{Surface: ExecutionZoneSurfaceExecutionMetrics, Requirement: "zone-execution-summary-inputs"},
		},
	}
	spec = canonicalExecutionZoneModelSpec(spec)
	spec.BoundaryHash = ComputeExecutionZoneModelHash(spec)
	return spec, spec.ValidateHash()
}

func (s ExecutionZoneModelSpec) ValidateHash() error {
	s = canonicalExecutionZoneModelSpec(s)
	if err := s.ValidateFormat(); err != nil {
		return err
	}
	expected := ComputeExecutionZoneModelHash(s)
	if s.BoundaryHash != expected {
		return fmt.Errorf("execution zone model hash mismatch: expected %s", expected)
	}
	return nil
}

func (s ExecutionZoneModelSpec) ValidateFormat() error {
	if err := ValidateZoneID(s.ZoneID); err != nil {
		return err
	}
	if !IsZoneKind(s.Kind) {
		return fmt.Errorf("unknown execution zone kind %q", s.Kind)
	}
	if err := validateZoneNamespace("execution zone state prefix", s.StatePrefix, ZoneKVPrefix(s.ZoneID)); err != nil {
		return err
	}
	if len(s.Surfaces) != 8 {
		return errors.New("execution zone model requires all eight zone-owned surfaces")
	}
	seen := make(map[ExecutionZoneSurface]struct{}, len(s.Surfaces))
	var previous ExecutionZoneSurface
	for i, surface := range s.Surfaces {
		if !IsExecutionZoneSurface(surface.Surface) {
			return fmt.Errorf("unknown execution zone surface %q", surface.Surface)
		}
		if _, found := seen[surface.Surface]; found {
			return fmt.Errorf("duplicate execution zone surface %s", surface.Surface)
		}
		seen[surface.Surface] = struct{}{}
		if err := validateRuntimeToken("execution zone surface requirement", surface.Requirement, MaxZoneNamespaceLength); err != nil {
			return err
		}
		if i > 0 && previous >= surface.Surface {
			return errors.New("execution zone surfaces must be sorted canonically")
		}
		previous = surface.Surface
	}
	if s.BoundaryHash != "" {
		return ValidateHash("execution zone model hash", s.BoundaryHash)
	}
	return nil
}

func IsExecutionZoneSurface(surface ExecutionZoneSurface) bool {
	switch surface {
	case ExecutionZoneSurfaceKeeperSet, ExecutionZoneSurfaceStatePrefix, ExecutionZoneSurfaceMempoolPolicy,
		ExecutionZoneSurfaceShardLayout, ExecutionZoneSurfaceMessageQueues, ExecutionZoneSurfaceFeePolicy,
		ExecutionZoneSurfaceProofRoot, ExecutionZoneSurfaceExecutionMetrics:
		return true
	default:
		return false
	}
}

func ComputeExecutionZoneModelHash(spec ExecutionZoneModelSpec) string {
	spec = canonicalExecutionZoneModelSpec(spec)
	parts := []string{"aetra-execution-zone-model-v1", string(spec.ZoneID), string(spec.Kind), spec.StatePrefix, fmt.Sprint(len(spec.Surfaces))}
	for _, surface := range spec.Surfaces {
		parts = append(parts, string(surface.Surface), surface.Requirement)
	}
	return hashRuntimeParts(parts...)
}

func canonicalExecutionZoneModelSpec(spec ExecutionZoneModelSpec) ExecutionZoneModelSpec {
	spec.Surfaces = append([]ExecutionZoneSurfaceRule(nil), spec.Surfaces...)
	sort.SliceStable(spec.Surfaces, func(i, j int) bool {
		return spec.Surfaces[i].Surface < spec.Surfaces[j].Surface
	})
	return spec
}
