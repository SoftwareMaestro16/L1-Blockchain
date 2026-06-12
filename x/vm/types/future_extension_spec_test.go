package types

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAVMFutureExtensionRegistryMatchesSection17(t *testing.T) {
	registry, err := DefaultAVMFutureExtensionRegistry()
	require.NoError(t, err)
	require.NoError(t, registry.Validate())
	require.Equal(t, "AVM future extensions", registry.RegistryName)
	require.Equal(t, ComputeAVMFutureExtensionRegistryHash(registry), registry.RegistryHash)
	require.Len(t, registry.Extensions, 9)

	byArea := map[AVMFutureExtensionArea]AVMFutureExtensionDescriptor{}
	for _, extension := range registry.Extensions {
		require.NoError(t, extension.Validate())
		require.Equal(t, AVMFutureExtensionStatusPlanned, extension.Status)
		require.NotEmpty(t, extension.Prerequisites)
		byArea[extension.Area] = extension
	}
	require.Equal(t, "Speculative execution layer", byArea[AVMFutureExtensionSpeculativeExecution].Name)
	require.Equal(t, "Parallel actor scheduling", byArea[AVMFutureExtensionParallelActorScheduling].Name)
	require.Equal(t, "Zero-knowledge execution attestation", byArea[AVMFutureExtensionZKExecutionAttestation].Name)
	require.Equal(t, "Distributed async scheduler", byArea[AVMFutureExtensionDistributedScheduler].Name)
	require.Equal(t, "Cross-chain message bridge layer", byArea[AVMFutureExtensionCrossChainBridge].Name)
	require.Equal(t, "Actor state rent", byArea[AVMFutureExtensionActorStateRent].Name)
	require.Equal(t, "Interface package registry", byArea[AVMFutureExtensionInterfacePackageRegistry].Name)
	require.Equal(t, "Formal VM verification test suite", byArea[AVMFutureExtensionFormalVMVerification].Name)
	require.Equal(t, "Deterministic replay debugger", byArea[AVMFutureExtensionReplayDebugger].Name)
}

func TestAVMFutureExtensionsRenderDocumentOrderMarkdown(t *testing.T) {
	registry, err := DefaultAVMFutureExtensionRegistry()
	require.NoError(t, err)

	markdown, err := RenderAVMFutureExtensionsMarkdown(registry)
	require.NoError(t, err)
	lines := strings.Split(markdown, "\n")
	require.Equal(t, "- Speculative execution layer.", lines[0])
	require.Equal(t, "- Parallel actor scheduling.", lines[1])
	require.Equal(t, "- Zero-knowledge execution attestation.", lines[2])
	require.Equal(t, "- Deterministic replay debugger.", lines[8])
	require.Len(t, lines, 9)
}

func TestAVMFutureExtensionRegistryRejectsMissingDuplicateAndHashMismatch(t *testing.T) {
	registry, err := DefaultAVMFutureExtensionRegistry()
	require.NoError(t, err)

	missing := registry
	missing.Extensions = missing.Extensions[:len(missing.Extensions)-1]
	missing.RegistryHash = ComputeAVMFutureExtensionRegistryHash(missing)
	require.ErrorContains(t, missing.Validate(), "every section 17 extension area")

	duplicate := registry
	duplicate.Extensions[0] = duplicate.Extensions[1]
	duplicate.RegistryHash = ComputeAVMFutureExtensionRegistryHash(duplicate)
	require.ErrorContains(t, duplicate.Validate(), "duplicate")

	mutated := registry
	mutated.Extensions[0].Description = "changed"
	require.ErrorContains(t, mutated.Validate(), "descriptor hash mismatch")
}

func TestAVMFutureExtensionConsensusAffectingRequiresGovernedUpgrade(t *testing.T) {
	_, err := NewAVMFutureExtensionDescriptor(AVMFutureExtensionDescriptor{
		Area:				AVMFutureExtensionSpeculativeExecution,
		Name:				"Speculative execution layer",
		Description:			"Pre-execute eligible deterministic workloads while preserving commit-time validation.",
		Prerequisites:			[]string{"deterministic_replay_tests"},
		ConsensusAffecting:		true,
		RequiresGovernedUpgrade:	false,
		Status:				AVMFutureExtensionStatusPlanned,
	})
	require.ErrorContains(t, err, "governed upgrade")

	descriptor, err := NewAVMFutureExtensionDescriptor(AVMFutureExtensionDescriptor{
		Area:				AVMFutureExtensionReplayDebugger,
		Name:				"Deterministic replay debugger",
		Description:			"Inspect deterministic execution traces without changing consensus outputs.",
		Prerequisites:			[]string{"execution_receipts", "replay_export_import"},
		ConsensusAffecting:		false,
		RequiresGovernedUpgrade:	false,
		Status:				AVMFutureExtensionStatusPlanned,
	})
	require.NoError(t, err)
	require.NoError(t, descriptor.Validate())
}

func TestAVMFutureExtensionRejectsInvalidMarkdownCharacters(t *testing.T) {
	_, err := NewAVMFutureExtensionDescriptor(AVMFutureExtensionDescriptor{
		Area:				AVMFutureExtensionInterfacePackageRegistry,
		Name:				"Interface | package registry",
		Description:			"Publish interface descriptors as versioned packages.",
		Prerequisites:			[]string{"interface_hash_verification"},
		ConsensusAffecting:		false,
		RequiresGovernedUpgrade:	false,
		Status:				AVMFutureExtensionStatusPlanned,
	})
	require.ErrorContains(t, err, "invalid character")
}
