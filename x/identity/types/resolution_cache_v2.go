package types

import (
	"encoding/hex"
	"errors"
	"fmt"
)

type ResolutionCacheRecordV2 struct {
	NameHash           string
	ResolutionPathHash string
	ResolvedRecordHash string
	ValidUntilHeight   uint64
	SourceVersion      uint64
	ParentEpoch        uint64
	ChildEpoch         uint64
}

type ResolutionCacheUseContextV2 struct {
	Height        uint64
	SourceVersion uint64
	ParentEpoch   uint64
	ChildEpoch    uint64
	LightClient   bool
	ProofVerified bool
}

func NewResolutionCacheRecordV2(name string, resolutionPathHash string, resolvedRecordHash string, validUntilHeight uint64, sourceVersion uint64, parentEpoch uint64, childEpoch uint64) (ResolutionCacheRecordV2, error) {
	nameHash, err := DomainRecordV2NameHash(name)
	if err != nil {
		return ResolutionCacheRecordV2{}, err
	}
	record := ResolutionCacheRecordV2{
		NameHash:           nameHash,
		ResolutionPathHash: resolutionPathHash,
		ResolvedRecordHash: resolvedRecordHash,
		ValidUntilHeight:   validUntilHeight,
		SourceVersion:      sourceVersion,
		ParentEpoch:        parentEpoch,
		ChildEpoch:         childEpoch,
	}
	return record, ValidateResolutionCacheRecordV2(record)
}

func ValidateResolutionCacheRecordV2(record ResolutionCacheRecordV2) error {
	if err := validateHexHash("identity v2 cache name hash", record.NameHash); err != nil {
		return err
	}
	if err := validateHexHash("identity v2 cache resolution path hash", record.ResolutionPathHash); err != nil {
		return err
	}
	if err := validateHexHash("identity v2 cache resolved record hash", record.ResolvedRecordHash); err != nil {
		return err
	}
	if record.ValidUntilHeight == 0 {
		return errors.New("identity v2 cache valid_until_height is required")
	}
	if record.SourceVersion == 0 {
		return errors.New("identity v2 cache source_version is required")
	}
	return nil
}

func ValidateResolutionCacheRecordV2Use(record ResolutionCacheRecordV2, ctx ResolutionCacheUseContextV2) error {
	if err := ValidateResolutionCacheRecordV2(record); err != nil {
		return err
	}
	if ctx.Height == 0 {
		return errors.New("identity v2 cache use height is required")
	}
	if ctx.Height > record.ValidUntilHeight {
		return errors.New("identity v2 cache record is expired")
	}
	if ctx.SourceVersion != record.SourceVersion {
		return errors.New("identity v2 cache source version changed")
	}
	if ctx.ParentEpoch != record.ParentEpoch {
		return errors.New("identity v2 cache parent epoch changed")
	}
	if ctx.ChildEpoch != record.ChildEpoch {
		return errors.New("identity v2 cache child epoch changed")
	}
	if ctx.LightClient && !ctx.ProofVerified {
		return errors.New("identity v2 cache cannot bypass light-client proof verification")
	}
	return nil
}

func InvalidateResolutionCacheRecordV2ForDomainMutation(record ResolutionCacheRecordV2, sourceVersion uint64, parentEpoch uint64, childEpoch uint64) ResolutionCacheRecordV2 {
	next := record
	next.SourceVersion = sourceVersion
	next.ParentEpoch = parentEpoch
	next.ChildEpoch = childEpoch
	next.ValidUntilHeight = 0
	return next
}

func ComputeResolutionPathHashV2(candidates []string) (string, error) {
	if len(candidates) == 0 {
		return "", errors.New("identity v2 cache resolution path is required")
	}
	normalized := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		name, err := NormalizeAETDomain(candidate)
		if err != nil {
			return "", err
		}
		normalized = append(normalized, name)
	}
	return identityHash(append([]string{"identity-v2-resolution-path", fmt.Sprintf("%020d", len(normalized))}, normalized...)...), nil
}

func ComputeResolvedRecordHashV2(record UnifiedResolutionRecordV2) (string, error) {
	if err := ValidateUnifiedResolutionRecordV2(record); err != nil {
		return "", err
	}
	parts := []string{
		"identity-v2-resolved-record",
		record.NameHash,
		hex.EncodeToString(record.Owner),
		hex.EncodeToString(record.PrimaryAddress),
		fmt.Sprintf("%020d", record.RecordVersion),
		fmt.Sprintf("%020d", record.RecordTTL),
		fmt.Sprintf("%020d", record.UpdatedAtHeight),
		fmt.Sprintf("%020d", record.MaxPayloadBytes),
		fmt.Sprintf("%020d", record.SchemaVersion),
		record.RoutingMetadata.ZoneID,
		record.RoutingMetadata.ShardID,
		record.RoutingMetadata.VM,
		record.RoutingMetadata.Entrypoint,
	}
	for _, target := range record.ContractTargets {
		parts = append(parts,
			"contract",
			target.Key,
			hex.EncodeToString(target.Address),
			target.CodeID,
			target.TargetID,
			hex.EncodeToString(target.ContractAddress),
			target.Entrypoint,
			target.InterfaceHash,
			target.RequiredFundsPolicy,
			fmt.Sprintf("%020d", target.GasHint),
			fmt.Sprintf("%t", target.Enabled),
			fmt.Sprintf("%020d", target.UpdatedAtHeight),
		)
	}
	for _, endpoint := range record.ServiceEndpoints {
		parts = append(parts, "service", endpoint.Key, endpoint.Endpoint)
	}
	for _, descriptor := range record.InterfaceDescriptors {
		parts = append(parts, "interface", descriptor.InterfaceID, descriptor.Descriptor)
	}
	for _, hint := range record.ExecutionHints {
		parts = append(parts, "hint", hint.Key, hint.Value)
	}
	return identityHash(parts...), nil
}
