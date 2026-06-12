package types

import (
	"errors"
	"fmt"
	"sort"
)

type ZoneDescriptorSpec struct {
	ZoneID			ZoneID
	ZoneType		ZoneType
	ModuleName		string
	Enabled			bool
	StateMachineVersion	uint64
	MempoolPolicyID		string
	FeePolicyID		string
	ShardLayoutEpoch	uint64
	MaxShards		uint32
	MessageCapabilities	[]string
	ProofCapabilities	[]string
	UpgradeHeightOptional	uint64
}

type ZoneCommitmentSpec struct {
	Height			uint64
	ZoneID			ZoneID
	ZoneStateRoot		string
	ZoneMessageOutboxRoot	string
	ZoneMessageInboxRoot	string
	ZoneReceiptRoot		string
	ZoneEventRoot		string
	ShardRootsRoot		string
	ExecutionSummaryHash	string
}

type CoreExecutionPhaseSpec struct {
	Phase			KernelABCIPhase
	DeterministicWork	[]string
	RejectionChecks		[]string
	CommittedOutput		[]string
	PhaseSpecHash		string
}

type CoreExecutionPipelineSpec struct {
	Phases		[]CoreExecutionPhaseSpec
	PipelineHash	string
}

func NewZoneDescriptorSpec(descriptor ZoneDescriptor) (ZoneDescriptorSpec, error) {
	return NewZoneDescriptorSpecWithParams(descriptor, DefaultParams())
}

func NewZoneDescriptorSpecWithParams(descriptor ZoneDescriptor, params AetraCoreParams) (ZoneDescriptorSpec, error) {
	descriptor = CanonicalZoneDescriptor(descriptor)
	if err := descriptor.Validate(params); err != nil {
		return ZoneDescriptorSpec{}, err
	}
	return ZoneDescriptorSpec{
		ZoneID:			descriptor.ZoneID,
		ZoneType:		descriptor.ZoneType,
		ModuleName:		descriptor.ModuleName,
		Enabled:		descriptor.Enabled,
		StateMachineVersion:	descriptor.StateMachineVersion,
		MempoolPolicyID:	descriptor.MempoolPolicyID,
		FeePolicyID:		descriptor.FeePolicyID,
		ShardLayoutEpoch:	descriptor.ShardLayoutEpoch,
		MaxShards:		descriptor.MaxShards,
		MessageCapabilities:	append([]string(nil), descriptor.MessageCapabilities...),
		ProofCapabilities:	append([]string(nil), descriptor.ProofCapabilities...),
		UpgradeHeightOptional:	descriptor.UpgradeHeightOptional,
	}, nil
}

func NewZoneCommitmentSpec(commitment ZoneCommitment) (ZoneCommitmentSpec, error) {
	if err := commitment.ValidateHash(); err != nil {
		return ZoneCommitmentSpec{}, err
	}
	return ZoneCommitmentSpec{
		Height:			commitment.Height,
		ZoneID:			commitment.ZoneID,
		ZoneStateRoot:		commitment.StateRoot,
		ZoneMessageOutboxRoot:	commitment.OutboxRoot,
		ZoneMessageInboxRoot:	commitment.InboxRoot,
		ZoneReceiptRoot:	commitment.ReceiptsRoot,
		ZoneEventRoot:		commitment.EventsRoot,
		ShardRootsRoot:		commitment.ShardRootsRoot,
		ExecutionSummaryHash:	commitment.ExecutionSummaryHash,
	}, nil
}

func DefaultCoreExecutionPipelineSpec() (CoreExecutionPipelineSpec, error) {
	phases := []CoreExecutionPhaseSpec{
		{
			Phase:	KernelPhasePrepareProposal,
			DeterministicWork: []string{
				"group-transactions-by-zone-and-shard",
				"apply-size-and-gas-bounds",
				"prefer-disjoint-shard-workloads",
				"include-pending-inbound-messages-by-priority",
			},
			RejectionChecks: []string{
				"malformed-transactions",
				"disabled-zones",
				"invalid-shard-targets",
				"proposal-limit-exceeded",
			},
			CommittedOutput:	[]string{"proposal-schedule", "local-transaction-batches", "inbound-message-batches"},
		},
		{
			Phase:	KernelPhaseProcessProposal,
			DeterministicWork: []string{
				"rebuild-schedule-from-proposal",
				"verify-committed-routing-state",
			},
			RejectionChecks: []string{
				"invalid-grouping",
				"disabled-zones",
				"missing-shard-layouts",
				"wrong-message-delivery-order",
				"malformed-execution-batches",
			},
			CommittedOutput:	[]string{"accepted-proposal-schedule", "deterministic-rejection"},
		},
		{
			Phase:	KernelPhaseFinalizeBlock,
			DeterministicWork: []string{
				"execute-zone-batches",
				"execute-inbound-messages",
				"collect-outboxes",
				"compute-shard-roots",
				"compute-zone-roots",
				"aggregate-global-roots",
			},
			RejectionChecks: []string{
				"non-deterministic-execution-output",
				"root-mismatch",
				"duplicate-receipts",
				"invalid-proof-roots",
			},
			CommittedOutput:	[]string{"zone-commitments", "global-message-root", "receipt-roots", "proof-roots", "global-state-root"},
		},
		{
			Phase:	KernelPhaseCommit,
			DeterministicWork: []string{
				"persist-final-app-hash",
				"persist-root-snapshots",
				"expose-next-height-delivery-queues",
			},
			RejectionChecks: []string{
				"missing-finality-metadata",
				"app-hash-commitment-mismatch",
			},
			CommittedOutput:	[]string{"committed-app-hash", "historical-proof-roots", "delivery-eligibility"},
		},
	}
	for i := range phases {
		phases[i] = canonicalCoreExecutionPhaseSpec(phases[i])
		phases[i].PhaseSpecHash = ComputeCoreExecutionPhaseSpecHash(phases[i])
		if err := phases[i].ValidateHash(); err != nil {
			return CoreExecutionPipelineSpec{}, err
		}
	}
	pipeline := CoreExecutionPipelineSpec{Phases: phases}
	pipeline.PipelineHash = ComputeCoreExecutionPipelineHash(pipeline)
	return pipeline, pipeline.ValidateHash()
}

func (s ZoneDescriptorSpec) Validate() error {
	return ZoneDescriptor{
		ZoneID:			s.ZoneID,
		ZoneType:		s.ZoneType,
		ModuleName:		s.ModuleName,
		Enabled:		s.Enabled,
		StateMachineVersion:	s.StateMachineVersion,
		MempoolPolicyID:	s.MempoolPolicyID,
		FeePolicyID:		s.FeePolicyID,
		ShardLayoutEpoch:	s.ShardLayoutEpoch,
		MaxShards:		s.MaxShards,
		MessageCapabilities:	append([]string(nil), s.MessageCapabilities...),
		ProofCapabilities:	append([]string(nil), s.ProofCapabilities...),
		UpgradeHeightOptional:	s.UpgradeHeightOptional,
	}.Validate(DefaultParams())
}

func (s ZoneCommitmentSpec) Validate() error {
	return ZoneCommitment{
		Height:			s.Height,
		ZoneID:			s.ZoneID,
		StateRoot:		s.ZoneStateRoot,
		InboxRoot:		s.ZoneMessageInboxRoot,
		OutboxRoot:		s.ZoneMessageOutboxRoot,
		ReceiptsRoot:		s.ZoneReceiptRoot,
		EventsRoot:		s.ZoneEventRoot,
		ShardRootsRoot:		s.ShardRootsRoot,
		ParamsHash:		EmptyRootHash,
		ExecutionSummaryHash:	s.ExecutionSummaryHash,
		CommitmentHash:		EmptyRootHash,
	}.ValidateFormat()
}

func (p CoreExecutionPhaseSpec) ValidateHash() error {
	p = canonicalCoreExecutionPhaseSpec(p)
	if err := p.ValidateFormat(); err != nil {
		return err
	}
	expected := ComputeCoreExecutionPhaseSpecHash(p)
	if p.PhaseSpecHash != expected {
		return fmt.Errorf("aetracore execution phase spec hash mismatch: expected %s", expected)
	}
	return nil
}

func (p CoreExecutionPhaseSpec) ValidateFormat() error {
	if p.Phase != KernelPhasePrepareProposal && p.Phase != KernelPhaseProcessProposal && p.Phase != KernelPhaseFinalizeBlock && p.Phase != KernelPhaseCommit {
		return fmt.Errorf("unknown aetracore execution phase %q", p.Phase)
	}
	if len(p.DeterministicWork) == 0 {
		return errors.New("aetracore execution phase requires deterministic work")
	}
	if len(p.RejectionChecks) == 0 {
		return errors.New("aetracore execution phase requires rejection checks")
	}
	if len(p.CommittedOutput) == 0 {
		return errors.New("aetracore execution phase requires committed output")
	}
	if err := validateCapabilitiesForField("aetracore execution phase work", p.DeterministicWork); err != nil {
		return err
	}
	if err := validateCapabilitiesForField("aetracore execution phase rejection", p.RejectionChecks); err != nil {
		return err
	}
	if err := validateCapabilitiesForField("aetracore execution phase output", p.CommittedOutput); err != nil {
		return err
	}
	if p.PhaseSpecHash != "" {
		return ValidateHash("aetracore execution phase spec hash", p.PhaseSpecHash)
	}
	return nil
}

func (p CoreExecutionPipelineSpec) ValidateHash() error {
	if err := p.ValidateFormat(); err != nil {
		return err
	}
	expected := ComputeCoreExecutionPipelineHash(p)
	if p.PipelineHash != expected {
		return fmt.Errorf("aetracore execution pipeline hash mismatch: expected %s", expected)
	}
	return nil
}

func (p CoreExecutionPipelineSpec) ValidateFormat() error {
	if len(p.Phases) != 4 {
		return errors.New("aetracore execution pipeline requires four ABCI phases")
	}
	expected := []KernelABCIPhase{KernelPhasePrepareProposal, KernelPhaseProcessProposal, KernelPhaseFinalizeBlock, KernelPhaseCommit}
	for i, phase := range p.Phases {
		phase = canonicalCoreExecutionPhaseSpec(phase)
		if phase.Phase != expected[i] {
			return fmt.Errorf("aetracore execution pipeline phase order mismatch at index %d", i)
		}
		if err := phase.ValidateHash(); err != nil {
			return err
		}
	}
	if p.PipelineHash != "" {
		return ValidateHash("aetracore execution pipeline hash", p.PipelineHash)
	}
	return nil
}

func ComputeCoreExecutionPhaseSpecHash(phase CoreExecutionPhaseSpec) string {
	phase = canonicalCoreExecutionPhaseSpec(phase)
	parts := []string{"aetra-aek-core-execution-phase-v1", string(phase.Phase)}
	parts = appendStringSliceParts(parts, "work", phase.DeterministicWork)
	parts = appendStringSliceParts(parts, "reject", phase.RejectionChecks)
	parts = appendStringSliceParts(parts, "output", phase.CommittedOutput)
	return hashParts(parts...)
}

func ComputeCoreExecutionPipelineHash(pipeline CoreExecutionPipelineSpec) string {
	parts := []string{"aetra-aek-core-execution-pipeline-v1", fmt.Sprint(len(pipeline.Phases))}
	for _, phase := range pipeline.Phases {
		phase = canonicalCoreExecutionPhaseSpec(phase)
		parts = append(parts, ComputeCoreExecutionPhaseSpecHash(phase))
	}
	return hashParts(parts...)
}

func canonicalCoreExecutionPhaseSpec(phase CoreExecutionPhaseSpec) CoreExecutionPhaseSpec {
	phase.DeterministicWork = append([]string(nil), phase.DeterministicWork...)
	phase.RejectionChecks = append([]string(nil), phase.RejectionChecks...)
	phase.CommittedOutput = append([]string(nil), phase.CommittedOutput...)
	sort.Strings(phase.DeterministicWork)
	sort.Strings(phase.RejectionChecks)
	sort.Strings(phase.CommittedOutput)
	return phase
}
