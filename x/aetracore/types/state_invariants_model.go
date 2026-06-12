package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

const StateInvariantModelVersion = uint64(1)

type StateInvariantID string

const (
	StateInvariantZoneCommitmentMatchesState	StateInvariantID	= "zone-commitment-matches-state"
	StateInvariantShardRootsIncludedInZoneRoot	StateInvariantID	= "shard-roots-included-in-zone-root"
	StateInvariantOutputMessagesInOutboxRoot	StateInvariantID	= "output-messages-in-outbox-root"
	StateInvariantConsumedMessagesHaveOneReceipt	StateInvariantID	= "consumed-messages-have-one-receipt"
	StateInvariantCrossZoneValueConserved		StateInvariantID	= "cross-zone-value-conserved"
	StateInvariantPaymentCollateralMatchesBankLock	StateInvariantID	= "payment-collateral-matches-bank-lock"
	StateInvariantIdentityDomainOwnershipBinding	StateInvariantID	= "identity-domain-ownership-binding"
	StateInvariantContractStorageProofZoneBinding	StateInvariantID	= "contract-storage-proof-zone-binding"
)

type StateInvariantDescriptor struct {
	InvariantID	StateInvariantID
	Invariant	string
	Enforcement	string
	DescriptorHash	string
}

type StateInvariantModel struct {
	Version		uint64
	Invariants	[]StateInvariantDescriptor
	Root		string
}

type StateInvariantEvidence struct {
	Height					uint64
	ZoneCommitmentMatchesExecutedState	bool
	ShardRootsIncludedInZoneRoot		bool
	OutputMessagesIncludedInSourceOutbox	bool
	ConsumedMessagesHaveOneReceipt		bool
	CrossZoneValueTransferConservesNaet	bool
	PaymentCollateralMatchesLockedBalance	bool
	IdentityDomainsHaveOwnershipBinding	bool
	ContractStorageProofsBindToZoneRoot	bool
	ZoneCommitmentRoot			string
	ShardRootsRoot				string
	MessageOutboxRoot			string
	MessageReceiptRoot			string
	ValueConservationRoot			string
	PaymentCollateralRoot			string
	IdentityOwnershipRoot			string
	ContractStorageProofRoot		string
	EvidenceHash				string
}

func StateInvariantDescriptors() []StateInvariantDescriptor {
	return []StateInvariantDescriptor{
		stateInvariant(StateInvariantZoneCommitmentMatchesState, "Every zone commitment must match executed zone state.", "Recompute zone state, inbox, outbox, receipt, event, shard, and summary roots before commit."),
		stateInvariant(StateInvariantShardRootsIncludedInZoneRoot, "Every shard root must be included in its zone root.", "Aggregate shard roots canonically into shard_roots_root."),
		stateInvariant(StateInvariantOutputMessagesInOutboxRoot, "Every output message must be included in source outbox root.", "Message emission writes outbox entries before global message root aggregation."),
		stateInvariant(StateInvariantConsumedMessagesHaveOneReceipt, "Every consumed message must have one receipt.", "Inbox delivery records exactly one success, failure, bounce, expired, rejected, or deferred receipt."),
		stateInvariant(StateInvariantCrossZoneValueConserved, "Every cross-zone value transfer must conserve naet.", "Source escrow, destination credit, refunds, fees, and bounces must sum to the original value and fee budget."),
		stateInvariant(StateInvariantPaymentCollateralMatchesBankLock, "Every active payment channel collateral record must match locked bank balance.", "Channel collateral roots reconcile with Financial Zone locked balances."),
		stateInvariant(StateInvariantIdentityDomainOwnershipBinding, "Every active identity domain must have valid ownership binding.", "Domain record verifies owner, NFT binding, delegation, expiry, and status."),
		stateInvariant(StateInvariantContractStorageProofZoneBinding, "Every contract storage proof must bind to contract zone root.", "Contract storage proofs verify through contract instance root, shard root, and Contract Zone root."),
	}
}

func BuildStateInvariantModel(invariants []StateInvariantDescriptor) (StateInvariantModel, error) {
	model := StateInvariantModel{
		Version:	StateInvariantModelVersion,
		Invariants:	normalizeStateInvariantDescriptors(invariants),
	}
	if err := model.ValidateFormat(); err != nil {
		return StateInvariantModel{}, err
	}
	model.Root = ComputeStateInvariantModelRoot(model.Invariants)
	return model, model.Validate()
}

func DefaultStateInvariantModel() (StateInvariantModel, error) {
	return BuildStateInvariantModel(StateInvariantDescriptors())
}

func (m StateInvariantModel) Normalize() StateInvariantModel {
	if m.Version == 0 {
		m.Version = StateInvariantModelVersion
	}
	m.Invariants = normalizeStateInvariantDescriptors(m.Invariants)
	m.Root = strings.ToLower(strings.TrimSpace(m.Root))
	return m
}

func (m StateInvariantModel) ValidateFormat() error {
	m = m.Normalize()
	if m.Version != StateInvariantModelVersion {
		return fmt.Errorf("aetracore state invariant model version must be %d", StateInvariantModelVersion)
	}
	if len(m.Invariants) == 0 {
		return errors.New("aetracore state invariant model requires invariants")
	}
	seen := make(map[StateInvariantID]struct{}, len(m.Invariants))
	var previous StateInvariantID
	for i, invariant := range m.Invariants {
		if err := invariant.Validate(); err != nil {
			return err
		}
		if _, found := seen[invariant.InvariantID]; found {
			return fmt.Errorf("aetracore duplicate state invariant %s", invariant.InvariantID)
		}
		seen[invariant.InvariantID] = struct{}{}
		if i > 0 && previous >= invariant.InvariantID {
			return errors.New("aetracore state invariants must be sorted canonically")
		}
		previous = invariant.InvariantID
	}
	if m.Root != "" {
		if err := ValidateHash("aetracore state invariant model root", m.Root); err != nil {
			return err
		}
	}
	return nil
}

func (m StateInvariantModel) Validate() error {
	m = m.Normalize()
	if err := m.ValidateFormat(); err != nil {
		return err
	}
	if m.Root == "" {
		return errors.New("aetracore state invariant model root is required")
	}
	expected := ComputeStateInvariantModelRoot(m.Invariants)
	if m.Root != expected {
		return fmt.Errorf("aetracore state invariant model root mismatch: expected %s", expected)
	}
	return nil
}

func BuildStateInvariantDescriptor(invariant StateInvariantDescriptor) (StateInvariantDescriptor, error) {
	invariant = invariant.Normalize()
	if invariant.DescriptorHash != "" {
		return StateInvariantDescriptor{}, errors.New("aetracore state invariant descriptor hash must be empty before construction")
	}
	if err := invariant.ValidateFormat(); err != nil {
		return StateInvariantDescriptor{}, err
	}
	invariant.DescriptorHash = ComputeStateInvariantDescriptorHash(invariant)
	return invariant, invariant.Validate()
}

func (invariant StateInvariantDescriptor) Normalize() StateInvariantDescriptor {
	invariant.Invariant = normalizeModuleMapText(invariant.Invariant)
	invariant.Enforcement = normalizeModuleMapText(invariant.Enforcement)
	invariant.DescriptorHash = strings.ToLower(strings.TrimSpace(invariant.DescriptorHash))
	return invariant
}

func (invariant StateInvariantDescriptor) ValidateFormat() error {
	invariant = invariant.Normalize()
	if !IsStateInvariantID(invariant.InvariantID) {
		return fmt.Errorf("unknown aetracore state invariant %q", invariant.InvariantID)
	}
	if invariant.Invariant == "" {
		return errors.New("aetracore state invariant text is required")
	}
	if invariant.Enforcement == "" {
		return errors.New("aetracore state invariant enforcement is required")
	}
	if invariant.DescriptorHash != "" {
		if err := ValidateHash("aetracore state invariant descriptor hash", invariant.DescriptorHash); err != nil {
			return err
		}
	}
	return nil
}

func (invariant StateInvariantDescriptor) Validate() error {
	invariant = invariant.Normalize()
	if err := invariant.ValidateFormat(); err != nil {
		return err
	}
	if invariant.DescriptorHash == "" {
		return errors.New("aetracore state invariant descriptor hash is required")
	}
	expected := ComputeStateInvariantDescriptorHash(invariant)
	if invariant.DescriptorHash != expected {
		return fmt.Errorf("aetracore state invariant descriptor hash mismatch: expected %s", expected)
	}
	return nil
}

func BuildStateInvariantEvidence(evidence StateInvariantEvidence) (StateInvariantEvidence, error) {
	evidence = evidence.Normalize()
	if evidence.EvidenceHash != "" {
		return StateInvariantEvidence{}, errors.New("aetracore state invariant evidence hash must be empty before construction")
	}
	if err := evidence.ValidateFormat(); err != nil {
		return StateInvariantEvidence{}, err
	}
	evidence.EvidenceHash = ComputeStateInvariantEvidenceHash(evidence)
	return evidence, evidence.Validate()
}

func (evidence StateInvariantEvidence) Normalize() StateInvariantEvidence {
	evidence.ZoneCommitmentRoot = strings.ToLower(strings.TrimSpace(evidence.ZoneCommitmentRoot))
	evidence.ShardRootsRoot = strings.ToLower(strings.TrimSpace(evidence.ShardRootsRoot))
	evidence.MessageOutboxRoot = strings.ToLower(strings.TrimSpace(evidence.MessageOutboxRoot))
	evidence.MessageReceiptRoot = strings.ToLower(strings.TrimSpace(evidence.MessageReceiptRoot))
	evidence.ValueConservationRoot = strings.ToLower(strings.TrimSpace(evidence.ValueConservationRoot))
	evidence.PaymentCollateralRoot = strings.ToLower(strings.TrimSpace(evidence.PaymentCollateralRoot))
	evidence.IdentityOwnershipRoot = strings.ToLower(strings.TrimSpace(evidence.IdentityOwnershipRoot))
	evidence.ContractStorageProofRoot = strings.ToLower(strings.TrimSpace(evidence.ContractStorageProofRoot))
	evidence.EvidenceHash = strings.ToLower(strings.TrimSpace(evidence.EvidenceHash))
	return evidence
}

func (evidence StateInvariantEvidence) ValidateFormat() error {
	evidence = evidence.Normalize()
	if evidence.Height == 0 {
		return errors.New("aetracore state invariant evidence height must be positive")
	}
	if !evidence.ZoneCommitmentMatchesExecutedState {
		return errors.New("aetracore state invariant failed: zone commitment must match executed state")
	}
	if !evidence.ShardRootsIncludedInZoneRoot {
		return errors.New("aetracore state invariant failed: shard roots must be included in zone root")
	}
	if !evidence.OutputMessagesIncludedInSourceOutbox {
		return errors.New("aetracore state invariant failed: output messages must be included in source outbox")
	}
	if !evidence.ConsumedMessagesHaveOneReceipt {
		return errors.New("aetracore state invariant failed: consumed messages must have exactly one receipt")
	}
	if !evidence.CrossZoneValueTransferConservesNaet {
		return errors.New("aetracore state invariant failed: cross-zone value transfer must conserve naet")
	}
	if !evidence.PaymentCollateralMatchesLockedBalance {
		return errors.New("aetracore state invariant failed: payment collateral must match locked bank balance")
	}
	if !evidence.IdentityDomainsHaveOwnershipBinding {
		return errors.New("aetracore state invariant failed: identity domain ownership binding is required")
	}
	if !evidence.ContractStorageProofsBindToZoneRoot {
		return errors.New("aetracore state invariant failed: contract storage proof must bind to contract zone root")
	}
	for _, item := range []struct {
		name	string
		value	string
	}{
		{"aetracore state invariant zone commitment root", evidence.ZoneCommitmentRoot},
		{"aetracore state invariant shard roots root", evidence.ShardRootsRoot},
		{"aetracore state invariant message outbox root", evidence.MessageOutboxRoot},
		{"aetracore state invariant message receipt root", evidence.MessageReceiptRoot},
		{"aetracore state invariant value conservation root", evidence.ValueConservationRoot},
		{"aetracore state invariant payment collateral root", evidence.PaymentCollateralRoot},
		{"aetracore state invariant identity ownership root", evidence.IdentityOwnershipRoot},
		{"aetracore state invariant contract storage proof root", evidence.ContractStorageProofRoot},
	} {
		if err := ValidateHash(item.name, item.value); err != nil {
			return err
		}
	}
	if evidence.EvidenceHash != "" {
		if err := ValidateHash("aetracore state invariant evidence hash", evidence.EvidenceHash); err != nil {
			return err
		}
	}
	return nil
}

func (evidence StateInvariantEvidence) Validate() error {
	evidence = evidence.Normalize()
	if err := evidence.ValidateFormat(); err != nil {
		return err
	}
	if evidence.EvidenceHash == "" {
		return errors.New("aetracore state invariant evidence hash is required")
	}
	expected := ComputeStateInvariantEvidenceHash(evidence)
	if evidence.EvidenceHash != expected {
		return fmt.Errorf("aetracore state invariant evidence hash mismatch: expected %s", expected)
	}
	return nil
}

func IsStateInvariantID(id StateInvariantID) bool {
	switch id {
	case StateInvariantZoneCommitmentMatchesState,
		StateInvariantShardRootsIncludedInZoneRoot,
		StateInvariantOutputMessagesInOutboxRoot,
		StateInvariantConsumedMessagesHaveOneReceipt,
		StateInvariantCrossZoneValueConserved,
		StateInvariantPaymentCollateralMatchesBankLock,
		StateInvariantIdentityDomainOwnershipBinding,
		StateInvariantContractStorageProofZoneBinding:
		return true
	default:
		return false
	}
}

func ComputeStateInvariantDescriptorHash(invariant StateInvariantDescriptor) string {
	invariant = invariant.Normalize()
	return hashParts(
		"aetra-state-invariant-descriptor-v1",
		string(invariant.InvariantID),
		invariant.Invariant,
		invariant.Enforcement,
	)
}

func ComputeStateInvariantModelRoot(invariants []StateInvariantDescriptor) string {
	normalized := normalizeStateInvariantDescriptors(invariants)
	parts := []string{"aetra-state-invariant-model-v1", fmt.Sprintf("%020d", StateInvariantModelVersion)}
	for _, invariant := range normalized {
		parts = append(parts, string(invariant.InvariantID), invariant.DescriptorHash)
	}
	return hashParts(parts...)
}

func ComputeStateInvariantEvidenceHash(evidence StateInvariantEvidence) string {
	evidence = evidence.Normalize()
	return hashParts(
		"aetra-state-invariant-evidence-v1",
		fmt.Sprintf("%020d", evidence.Height),
		fmt.Sprintf("%t", evidence.ZoneCommitmentMatchesExecutedState),
		fmt.Sprintf("%t", evidence.ShardRootsIncludedInZoneRoot),
		fmt.Sprintf("%t", evidence.OutputMessagesIncludedInSourceOutbox),
		fmt.Sprintf("%t", evidence.ConsumedMessagesHaveOneReceipt),
		fmt.Sprintf("%t", evidence.CrossZoneValueTransferConservesNaet),
		fmt.Sprintf("%t", evidence.PaymentCollateralMatchesLockedBalance),
		fmt.Sprintf("%t", evidence.IdentityDomainsHaveOwnershipBinding),
		fmt.Sprintf("%t", evidence.ContractStorageProofsBindToZoneRoot),
		evidence.ZoneCommitmentRoot,
		evidence.ShardRootsRoot,
		evidence.MessageOutboxRoot,
		evidence.MessageReceiptRoot,
		evidence.ValueConservationRoot,
		evidence.PaymentCollateralRoot,
		evidence.IdentityOwnershipRoot,
		evidence.ContractStorageProofRoot,
	)
}

func ValidateStateInvariantModel() error {
	model, err := DefaultStateInvariantModel()
	if err != nil {
		return err
	}
	required := []StateInvariantID{
		StateInvariantZoneCommitmentMatchesState,
		StateInvariantShardRootsIncludedInZoneRoot,
		StateInvariantOutputMessagesInOutboxRoot,
		StateInvariantConsumedMessagesHaveOneReceipt,
		StateInvariantCrossZoneValueConserved,
		StateInvariantPaymentCollateralMatchesBankLock,
		StateInvariantIdentityDomainOwnershipBinding,
		StateInvariantContractStorageProofZoneBinding,
	}
	seen := make(map[StateInvariantID]struct{}, len(model.Invariants))
	for _, invariant := range model.Invariants {
		seen[invariant.InvariantID] = struct{}{}
	}
	for _, invariantID := range required {
		if _, found := seen[invariantID]; !found {
			return fmt.Errorf("aetracore state invariant model missing %s", invariantID)
		}
	}
	return nil
}

func stateInvariant(id StateInvariantID, invariant, enforcement string) StateInvariantDescriptor {
	desc, err := BuildStateInvariantDescriptor(StateInvariantDescriptor{
		InvariantID:	id,
		Invariant:	invariant,
		Enforcement:	enforcement,
	})
	if err != nil {
		panic(err)
	}
	return desc
}

func normalizeStateInvariantDescriptors(invariants []StateInvariantDescriptor) []StateInvariantDescriptor {
	out := make([]StateInvariantDescriptor, len(invariants))
	for i, invariant := range invariants {
		normalized := invariant.Normalize()
		if normalized.DescriptorHash == "" {
			normalized.DescriptorHash = ComputeStateInvariantDescriptorHash(normalized)
		}
		out[i] = normalized
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].InvariantID < out[j].InvariantID
	})
	return out
}
