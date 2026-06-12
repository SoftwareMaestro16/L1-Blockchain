package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

func TestAVMUpgradeManifestStagesGovernedComponents(t *testing.T) {
	manifest := testAVMUpgradeManifest(t, 100)
	require.NoError(t, manifest.Validate())
	require.Equal(t, ComputeAVMUpgradeManifestHash(manifest), manifest.ManifestHash)
	require.Len(t, manifest.Components, len(AllAVMUpgradeComponentKinds()))
	require.Len(t, manifest.SchedulerRules, 2)
	require.Len(t, manifest.GasTables, 1)
	require.Len(t, manifest.Continuations, 1)
	require.Len(t, manifest.ContractCodes, 1)
}

func TestAVMUpgradeRejectsActivationBeforeStaging(t *testing.T) {
	component, err := NewAVMUpgradeComponent(AVMUpgradeComponent{
		Kind:			AVMUpgradeComponentGasModel,
		PreviousVersion:	"gas-v1",
		NextVersion:		"gas-v2",
		ActivationHeight:	99,
	})
	require.NoError(t, err)

	manifest := testAVMUpgradeManifest(t, 100)
	manifest.Components = []AVMUpgradeComponent{component}
	manifest.ManifestHash = ComputeAVMUpgradeManifestHash(manifest)
	require.ErrorContains(t, manifest.Validate(), "staged before activation")

	table := manifest.GasTables[0]
	table.ActivationHeight = 99
	table.TableHash = ComputeAVMGasTableActivationHash(table)
	manifest = testAVMUpgradeManifest(t, 100)
	manifest.GasTables = []AVMGasTableActivation{table}
	manifest.ManifestHash = ComputeAVMUpgradeManifestHash(manifest)
	require.ErrorContains(t, manifest.Validate(), "staged activation height")
}

func TestAVMUpgradeRejectsSchedulerRuleOverlapInOneBlock(t *testing.T) {
	oldRule, err := NewAVMSchedulerRuleVersion(AVMSchedulerRuleVersion{
		RuleSetID:		"async-scheduler",
		Version:		"scheduler-v1",
		EffectiveFromHeight:	100,
		EffectiveUntilHeight:	120,
	})
	require.NoError(t, err)
	overlappingRule, err := NewAVMSchedulerRuleVersion(AVMSchedulerRuleVersion{
		RuleSetID:		"async-scheduler",
		Version:		"scheduler-v2",
		EffectiveFromHeight:	120,
		EffectiveUntilHeight:	0,
	})
	require.NoError(t, err)

	manifest := testAVMUpgradeManifest(t, 100)
	manifest.SchedulerRules = []AVMSchedulerRuleVersion{oldRule, overlappingRule}
	manifest.ManifestHash = ComputeAVMUpgradeManifestHash(manifest)
	require.ErrorContains(t, manifest.Validate(), "must not overlap")
}

func TestAVMUpgradeVersionedMessagesUseCreationPolicy(t *testing.T) {
	preUpgrade := testAVMUpgradeMessage(t, "pre-upgrade", 1, 90)
	policy, err := NewAVMVersionedMessageExecutionPolicy(
		preUpgrade,
		100,
		110,
		"runtime-v1",
		"runtime-v2",
		"scheduler-v1",
		"scheduler-v2",
		"gas-v1",
		"gas-v2",
	)
	require.NoError(t, err)
	require.Equal(t, "runtime-v1", policy.RuntimeVersion)
	require.Equal(t, "scheduler-v1", policy.SchedulerVersion)
	require.Equal(t, "gas-v1", policy.GasPolicyVersion)

	postUpgrade := testAVMUpgradeMessage(t, "post-upgrade", 2, 100)
	policy, err = NewAVMVersionedMessageExecutionPolicy(
		postUpgrade,
		100,
		110,
		"runtime-v1",
		"runtime-v2",
		"scheduler-v1",
		"scheduler-v2",
		"gas-v1",
		"gas-v2",
	)
	require.NoError(t, err)
	require.Equal(t, "runtime-v2", policy.RuntimeVersion)
	require.Equal(t, "scheduler-v2", policy.SchedulerVersion)
	require.Equal(t, "gas-v2", policy.GasPolicyVersion)
}

func TestAVMUpgradeRequiresContinuationRuntimeAndContractVMVersions(t *testing.T) {
	continuation, err := NewAVMContinuationRuntimeVersion(AVMContinuationRuntimeVersion{
		ContinuationID:	"continuation-1",
		ActorID:	"actor-1",
		RuntimeVersion:	"runtime-v1",
		StoredHeight:	99,
	})
	require.NoError(t, err)
	require.NoError(t, continuation.Validate())

	contractCode, err := NewAVMContractCodeVMVersion(AVMContractCodeVMVersion{
		CodeID:		7,
		BackendKind:	AVMContractBackendWASMContract,
		CodeHash:	engineHash("code-7"),
		VMVersion:	"wasm-v1",
	})
	require.NoError(t, err)
	require.NoError(t, contractCode.Validate())

	badContinuation := continuation
	badContinuation.RuntimeVersion = ""
	badContinuation.VersionHash = ComputeAVMContinuationRuntimeVersionHash(badContinuation)
	require.ErrorContains(t, badContinuation.Validate(), "runtime version")

	badCode := contractCode
	badCode.VMVersion = ""
	badCode.VersionHash = ComputeAVMContractCodeVMVersionHash(badCode)
	require.ErrorContains(t, badCode.Validate(), "VM version")
}

func TestAVMUpgradeGasTableAppliesAtActivationHeight(t *testing.T) {
	policy, err := DefaultAVMGasPolicy()
	require.NoError(t, err)
	policy.QueueInsertGas += 5
	policy.PolicyHash = ComputeAVMGasPolicyHash(policy)

	schedule, err := AVMGasScheduleFromPolicy(policy, true, 1_000_000)
	require.NoError(t, err)
	table, err := NewAVMGasTableActivation(AVMGasTableActivation{
		ActivationHeight:	200,
		PolicyVersion:		"gas-v2",
		Policy:			policy,
		Schedule:		schedule,
	})
	require.NoError(t, err)
	require.NoError(t, table.Validate())
	require.Equal(t, uint64(200), table.ActivationHeight)
	require.Equal(t, policy.PolicyHash, table.Policy.PolicyHash)
	require.Equal(t, schedule.ScheduleHash, table.Schedule.ScheduleHash)

	bad := table
	bad.ActivationHeight = 0
	bad.TableHash = ComputeAVMGasTableActivationHash(bad)
	require.ErrorContains(t, bad.Validate(), "activation height")
}

func testAVMUpgradeManifest(t *testing.T, stagedHeight uint64) AVMUpgradeManifest {
	t.Helper()
	components := make([]AVMUpgradeComponent, 0, len(AllAVMUpgradeComponentKinds()))
	for _, kind := range AllAVMUpgradeComponentKinds() {
		component, err := NewAVMUpgradeComponent(AVMUpgradeComponent{
			Kind:			kind,
			PreviousVersion:	fmt.Sprintf("%s-v1", kind),
			NextVersion:		fmt.Sprintf("%s-v2", kind),
			ActivationHeight:	stagedHeight + 10,
		})
		require.NoError(t, err)
		components = append(components, component)
	}
	oldRule, err := NewAVMSchedulerRuleVersion(AVMSchedulerRuleVersion{
		RuleSetID:		"async-scheduler",
		Version:		"scheduler-v1",
		EffectiveFromHeight:	stagedHeight,
		EffectiveUntilHeight:	stagedHeight + 9,
	})
	require.NoError(t, err)
	newRule, err := NewAVMSchedulerRuleVersion(AVMSchedulerRuleVersion{
		RuleSetID:		"async-scheduler",
		Version:		"scheduler-v2",
		EffectiveFromHeight:	stagedHeight + 10,
		EffectiveUntilHeight:	0,
	})
	require.NoError(t, err)
	policy, err := DefaultAVMGasPolicy()
	require.NoError(t, err)
	policy.QueueInsertGas += 7
	policy.PolicyHash = ComputeAVMGasPolicyHash(policy)
	schedule, err := AVMGasScheduleFromPolicy(policy, true, 1_000_000)
	require.NoError(t, err)
	gasTable, err := NewAVMGasTableActivation(AVMGasTableActivation{
		ActivationHeight:	stagedHeight + 10,
		PolicyVersion:		"gas-v2",
		Policy:			policy,
		Schedule:		schedule,
	})
	require.NoError(t, err)
	continuation, err := NewAVMContinuationRuntimeVersion(AVMContinuationRuntimeVersion{
		ContinuationID:	"continuation-1",
		ActorID:	"actor-1",
		RuntimeVersion:	"runtime-v1",
		StoredHeight:	stagedHeight,
	})
	require.NoError(t, err)
	contractCode, err := NewAVMContractCodeVMVersion(AVMContractCodeVMVersion{
		CodeID:		1,
		BackendKind:	AVMContractBackendActorContract,
		CodeHash:	engineHash("code-1"),
		VMVersion:	"avm-v1",
	})
	require.NoError(t, err)

	manifest, err := NewAVMUpgradeManifest(AVMUpgradeManifest{
		UpgradeID:		"upgrade-15",
		GovernanceProposalID:	"gov-15",
		StagedHeight:		stagedHeight,
		Components:		components,
		SchedulerRules:		[]AVMSchedulerRuleVersion{oldRule, newRule},
		GasTables:		[]AVMGasTableActivation{gasTable},
		Continuations:		[]AVMContinuationRuntimeVersion{continuation},
		ContractCodes:		[]AVMContractCodeVMVersion{contractCode},
	})
	require.NoError(t, err)
	return manifest
}

func testAVMUpgradeMessage(t *testing.T, sender string, nonce, createdHeight uint64) AVMAsyncMessage {
	t.Helper()
	msg := testAVMAsyncMessage(sender, zonestypes.ZoneIDApplication, "contract", zonestypes.ZoneIDContract, nonce, createdHeight)
	msg.PayloadType = "contract.call"
	msg.DelayHeight = 0
	msg.ExpiryHeight = createdHeight + 100
	msg.RetryPolicy = DefaultAVMRetryPolicy(msg.ExpiryHeight)
	built, err := NewAVMAsyncMessage(msg)
	require.NoError(t, err)
	return built
}
