package types

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

const (
	IdentityAdaptiveSyncSnapshotVersionV2	uint64	= 1

	IdentityAdaptiveSyncEventSnapshotV2	= "identity_adaptive_sync_snapshot"
	IdentityAdaptiveSyncEventRecoveredV2	= "identity_adaptive_sync_recovered"
	IdentityAdaptiveSyncEventCacheResyncV2	= "identity_adaptive_sync_cache_resync"
)

type IdentitySnapshotExpiryIndexEntryV2 struct {
	ExpiryHeight	uint64
	Name		string
	NameHash	string
	StoreKey	string
}

type IdentityAdaptiveSyncSnapshotV2 struct {
	SnapshotVersion	uint64
	Height		uint64
	State		IdentityState
	Delegations	[]DelegationRecordV2
	ExpiryIndex	[]IdentitySnapshotExpiryIndexEntryV2
	StateRoot	string
	DelegationRoot	string
	ExpiryIndexRoot	string
	SnapshotHash	string
}

type IdentityAdaptiveSyncRestoreResultV2 struct {
	State		IdentityState
	Delegations	[]DelegationRecordV2
	StateRoot	string
	ProofReady	bool
	Events		[]IdentityABCIEventV2
}

type IdentityCacheResyncPlanV2 struct {
	NameHash	string
	RecordVersion	uint64
	Height		uint64
	Events		[]IdentityABCIEventV2
	QueryNames	[]string
}

func BuildIdentityAdaptiveSyncSnapshotV2(state IdentityState, delegations []DelegationRecordV2, height uint64) (IdentityAdaptiveSyncSnapshotV2, error) {
	if height == 0 {
		return IdentityAdaptiveSyncSnapshotV2{}, errors.New("identity adaptive sync snapshot height is required")
	}
	exported, err := ImportIdentityState(state)
	if err != nil {
		return IdentityAdaptiveSyncSnapshotV2{}, err
	}
	canonicalDelegations, err := canonicalIdentityDelegationsV2(delegations)
	if err != nil {
		return IdentityAdaptiveSyncSnapshotV2{}, err
	}
	expiryIndex, err := BuildIdentitySnapshotExpiryIndexV2(exported)
	if err != nil {
		return IdentityAdaptiveSyncSnapshotV2{}, err
	}
	stateRoot, err := IdentityStateRoot(exported)
	if err != nil {
		return IdentityAdaptiveSyncSnapshotV2{}, err
	}
	snapshot := IdentityAdaptiveSyncSnapshotV2{
		SnapshotVersion:	IdentityAdaptiveSyncSnapshotVersionV2,
		Height:			height,
		State:			exported,
		Delegations:		canonicalDelegations,
		ExpiryIndex:		expiryIndex,
		StateRoot:		stateRoot,
		DelegationRoot:		ComputeIdentityZoneGrantRoot(canonicalDelegations),
		ExpiryIndexRoot:	ComputeIdentitySnapshotExpiryIndexRootV2(expiryIndex),
	}
	snapshot.SnapshotHash = ComputeIdentityAdaptiveSyncSnapshotHashV2(snapshot)
	return snapshot, nil
}

func ValidateIdentityAdaptiveSyncSnapshotV2(snapshot IdentityAdaptiveSyncSnapshotV2) error {
	if snapshot.SnapshotVersion != IdentityAdaptiveSyncSnapshotVersionV2 {
		return fmt.Errorf("unsupported identity adaptive sync snapshot version %d", snapshot.SnapshotVersion)
	}
	if snapshot.Height == 0 {
		return errors.New("identity adaptive sync snapshot height is required")
	}
	exported, err := ImportIdentityState(snapshot.State)
	if err != nil {
		return err
	}
	stateRoot, err := IdentityStateRoot(exported)
	if err != nil {
		return err
	}
	if snapshot.StateRoot != stateRoot {
		return errors.New("identity adaptive sync snapshot state root mismatch")
	}
	delegations, err := canonicalIdentityDelegationsV2(snapshot.Delegations)
	if err != nil {
		return err
	}
	if snapshot.DelegationRoot != ComputeIdentityZoneGrantRoot(delegations) {
		return errors.New("identity adaptive sync snapshot delegation root mismatch")
	}
	expectedExpiryIndex, err := BuildIdentitySnapshotExpiryIndexV2(exported)
	if err != nil {
		return err
	}
	expiryIndex := canonicalIdentitySnapshotExpiryIndexV2(snapshot.ExpiryIndex)
	if !reflect.DeepEqual(expectedExpiryIndex, expiryIndex) {
		return errors.New("identity adaptive sync snapshot expiry index mismatch")
	}
	if snapshot.ExpiryIndexRoot != ComputeIdentitySnapshotExpiryIndexRootV2(expiryIndex) {
		return errors.New("identity adaptive sync snapshot expiry index root mismatch")
	}
	if snapshot.SnapshotHash != ComputeIdentityAdaptiveSyncSnapshotHashV2(snapshot) {
		return errors.New("identity adaptive sync snapshot hash mismatch")
	}
	return nil
}

func RestoreIdentityAdaptiveSyncSnapshotV2(snapshot IdentityAdaptiveSyncSnapshotV2, proofProbeNames []string) (IdentityAdaptiveSyncRestoreResultV2, error) {
	if err := ValidateIdentityAdaptiveSyncSnapshotV2(snapshot); err != nil {
		return IdentityAdaptiveSyncRestoreResultV2{}, err
	}
	restored, err := ImportIdentityState(snapshot.State)
	if err != nil {
		return IdentityAdaptiveSyncRestoreResultV2{}, err
	}
	if err := ValidateIdentityKeeperInvariantsV2(IdentityKeeperInvariantInputV2{State: restored, Height: snapshot.Height}); err != nil {
		return IdentityAdaptiveSyncRestoreResultV2{}, err
	}
	stateRoot, err := IdentityStateRoot(restored)
	if err != nil {
		return IdentityAdaptiveSyncRestoreResultV2{}, err
	}
	if err := verifyAdaptiveSyncProofReadinessV2(restored, snapshot.Height, proofProbeNames); err != nil {
		return IdentityAdaptiveSyncRestoreResultV2{}, err
	}
	events := []IdentityABCIEventV2{{
		Type:		IdentityAdaptiveSyncEventRecoveredV2,
		Height:		snapshot.Height,
		Message:	"identity adaptive sync snapshot restored",
		Attributes:	[]string{"snapshot=" + snapshot.SnapshotHash, "state_root=" + stateRoot},
	}}
	plan, err := BuildIdentityAdaptiveSyncCacheResyncPlanV2(snapshot, proofProbeNames)
	if err != nil {
		return IdentityAdaptiveSyncRestoreResultV2{}, err
	}
	events = append(events, plan.Events...)
	return IdentityAdaptiveSyncRestoreResultV2{
		State:		restored,
		Delegations:	cloneIdentityDelegationsV2(snapshot.Delegations),
		StateRoot:	stateRoot,
		ProofReady:	true,
		Events:		events,
	}, nil
}

func BuildIdentityAdaptiveSyncCacheResyncPlanV2(snapshot IdentityAdaptiveSyncSnapshotV2, names []string) (IdentityCacheResyncPlanV2, error) {
	if err := ValidateIdentityAdaptiveSyncSnapshotV2(snapshot); err != nil {
		return IdentityCacheResyncPlanV2{}, err
	}
	queryNames, err := canonicalAdaptiveSyncQueryNamesV2(snapshot.State, names)
	if err != nil {
		return IdentityCacheResyncPlanV2{}, err
	}
	plan := IdentityCacheResyncPlanV2{
		Height:		snapshot.Height,
		RecordVersion:	IdentityAdaptiveSyncSnapshotVersionV2,
		QueryNames:	queryNames,
	}
	for _, name := range queryNames {
		nameHash, err := DomainRecordV2NameHash(name)
		if err != nil {
			return IdentityCacheResyncPlanV2{}, err
		}
		if plan.NameHash == "" {
			plan.NameHash = nameHash
		}
		plan.Events = append(plan.Events, IdentityABCIEventV2{
			Type:		IdentityAdaptiveSyncEventCacheResyncV2,
			Height:		snapshot.Height,
			Name:		name,
			NameHash:	nameHash,
			Message:	"identity watcher and wallet cache must resync",
			Attributes: []string{
				"snapshot=" + snapshot.SnapshotHash,
				"state_root=" + snapshot.StateRoot,
				"record_version=1",
			},
		})
	}
	return plan, nil
}

func BuildIdentitySnapshotExpiryIndexV2(state IdentityState) ([]IdentitySnapshotExpiryIndexEntryV2, error) {
	exported := state.Export()
	if err := exported.Validate(); err != nil {
		return nil, err
	}
	entries := make([]IdentitySnapshotExpiryIndexEntryV2, 0, len(exported.Domains))
	for _, domain := range exported.Domains {
		nameHash, err := DomainRecordV2NameHash(domain.Name)
		if err != nil {
			return nil, err
		}
		storeKey, err := IdentityStoreV2SpecExpiryIndexKey(domain.ExpiryHeight, domain.Name)
		if err != nil {
			return nil, err
		}
		entries = append(entries, IdentitySnapshotExpiryIndexEntryV2{
			ExpiryHeight:	domain.ExpiryHeight,
			Name:		domain.Name,
			NameHash:	nameHash,
			StoreKey:	storeKey,
		})
	}
	return canonicalIdentitySnapshotExpiryIndexV2(entries), nil
}

func ComputeIdentitySnapshotExpiryIndexRootV2(entries []IdentitySnapshotExpiryIndexEntryV2) string {
	ordered := canonicalIdentitySnapshotExpiryIndexV2(entries)
	parts := []string{"identity-adaptive-sync-expiry-index-root-v2", fmt.Sprintf("%020d", len(ordered))}
	for _, entry := range ordered {
		parts = append(parts, fmt.Sprintf("%020d", entry.ExpiryHeight), entry.Name, entry.NameHash, entry.StoreKey)
	}
	return identityHash(parts...)
}

func ComputeIdentityAdaptiveSyncSnapshotHashV2(snapshot IdentityAdaptiveSyncSnapshotV2) string {
	return identityHash(
		"identity-adaptive-sync-snapshot-v2",
		fmt.Sprintf("%020d", snapshot.SnapshotVersion),
		fmt.Sprintf("%020d", snapshot.Height),
		snapshot.StateRoot,
		snapshot.DelegationRoot,
		snapshot.ExpiryIndexRoot,
	)
}

func verifyAdaptiveSyncProofReadinessV2(state IdentityState, height uint64, names []string) error {
	queryNames, err := canonicalAdaptiveSyncProofProbeNamesV2(state, names)
	if err != nil {
		return err
	}
	for _, name := range queryNames {
		proof, err := BuildIdentityResolutionProof(state, name, height)
		if err != nil {
			return err
		}
		if _, err := VerifyIdentityResolutionProof(proof, height); err != nil {
			return err
		}
	}
	return nil
}

func canonicalAdaptiveSyncProofProbeNamesV2(state IdentityState, names []string) ([]string, error) {
	if len(names) > 0 {
		return canonicalAdaptiveSyncQueryNamesV2(state, names)
	}
	queryNames := make([]string, 0, len(state.Resolvers))
	for _, resolver := range state.Export().Resolvers {
		queryNames = append(queryNames, resolver.Domain)
	}
	sort.Strings(queryNames)
	return queryNames, nil
}

func canonicalAdaptiveSyncQueryNamesV2(state IdentityState, names []string) ([]string, error) {
	if len(names) == 0 {
		queryNames := make([]string, 0, len(state.Domains))
		for _, domain := range state.Export().Domains {
			queryNames = append(queryNames, domain.Name)
		}
		return queryNames, nil
	}
	seen := make(map[string]struct{}, len(names))
	queryNames := make([]string, 0, len(names))
	for _, name := range names {
		normalized, err := NormalizeAETDomain(name)
		if err != nil {
			return nil, err
		}
		if _, found := seen[normalized]; found {
			continue
		}
		seen[normalized] = struct{}{}
		queryNames = append(queryNames, normalized)
	}
	sort.Strings(queryNames)
	return queryNames, nil
}

func canonicalIdentityDelegationsV2(delegations []DelegationRecordV2) ([]DelegationRecordV2, error) {
	out := cloneIdentityDelegationsV2(delegations)
	for i := range out {
		out[i].Permissions = sortStringSet(out[i].Permissions)
		out[i].RecordPrefixLimit = strings.TrimSpace(out[i].RecordPrefixLimit)
		if err := ValidateDelegationRecordV2(out[i]); err != nil {
			return nil, err
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].NameHash != out[j].NameHash {
			return out[i].NameHash < out[j].NameHash
		}
		if string(out[i].Delegate) != string(out[j].Delegate) {
			return string(out[i].Delegate) < string(out[j].Delegate)
		}
		if out[i].Scope != out[j].Scope {
			return out[i].Scope < out[j].Scope
		}
		return out[i].CreatedAtHeight < out[j].CreatedAtHeight
	})
	return out, nil
}

func cloneIdentityDelegationsV2(delegations []DelegationRecordV2) []DelegationRecordV2 {
	if len(delegations) == 0 {
		return nil
	}
	out := make([]DelegationRecordV2, len(delegations))
	for i, delegation := range delegations {
		out[i] = delegation
		out[i].Delegate = cloneSpecAddress(delegation.Delegate)
		out[i].Permissions = append([]string(nil), delegation.Permissions...)
	}
	return out
}

func canonicalIdentitySnapshotExpiryIndexV2(entries []IdentitySnapshotExpiryIndexEntryV2) []IdentitySnapshotExpiryIndexEntryV2 {
	out := append([]IdentitySnapshotExpiryIndexEntryV2(nil), entries...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].ExpiryHeight != out[j].ExpiryHeight {
			return out[i].ExpiryHeight < out[j].ExpiryHeight
		}
		if out[i].NameHash != out[j].NameHash {
			return out[i].NameHash < out[j].NameHash
		}
		return out[i].Name < out[j].Name
	})
	return out
}
